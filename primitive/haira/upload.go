package haira

import (
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

// uploadsDir returns the base directory for file uploads.
// Uses HAIRA_UPLOADS_DIR if set, otherwise ./data/uploads (persistent, not /tmp).
func uploadsDir() string {
	if dir := os.Getenv("HAIRA_UPLOADS_DIR"); dir != "" {
		return dir
	}
	return filepath.Join(".", "data", "uploads")
}

// saveTempFile saves a multipart file upload to a temp file and returns the path.
func saveTempFile(src multipart.File, originalName string) (string, error) {
	dir := filepath.Join(os.TempDir(), "haira-uploads")
	os.MkdirAll(dir, 0755)
	ext := filepath.Ext(originalName)
	tmp, err := os.CreateTemp(dir, "upload-*"+ext)
	if err != nil {
		return "", err
	}
	defer tmp.Close()
	if _, err := io.Copy(tmp, src); err != nil {
		os.Remove(tmp.Name())
		return "", err
	}
	return tmp.Name(), nil
}

var safeSessionID = regexp.MustCompile(`^[a-zA-Z0-9\-]+$`)

// saveSessionFile saves an upload to a session-scoped directory that persists
// across requests. Path: {uploadsDir}/sessions/{sessionID}/{filename}
// Also registers the file in the store for metadata tracking.
func saveSessionFile(src multipart.File, originalName, sessionID string) (string, error) {
	if !safeSessionID.MatchString(sessionID) {
		return "", os.ErrInvalid
	}
	dir := filepath.Join(uploadsDir(), "sessions", sessionID)
	os.MkdirAll(dir, 0755)
	name := filepath.Base(originalName)
	dst := filepath.Join(dir, name)
	f, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := io.Copy(f, src); err != nil {
		os.Remove(dst)
		return "", err
	}

	// Auto-register in store for persistent metadata tracking
	registerUploadInStore(sessionID, name, dst)

	return dst, nil
}

// registerUploadInStore reads a file from disk and persists it in the store.
// Runs best-effort — errors are logged but don't fail the upload.
func registerUploadInStore(sessionID, name, filePath string) {
	if globalStore == nil {
		return
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[haira] Warning: could not read file for store registration: %v\n", err)
		return
	}
	contentType := mime.TypeByExtension(filepath.Ext(name))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	fileID, err := globalStore.SaveFile(sessionID, name, contentType, data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[haira] Warning: could not register file in store: %v\n", err)
		return
	}
	fmt.Fprintf(os.Stderr, "[haira] File registered in store: %s (id=%s, session=%s)\n", name, fileID, sessionID)
}

// RestoreSessionFile restores a file from the store to disk, returning the path.
// Used when a tool needs a file that was previously uploaded but may have been
// cleaned from the filesystem.
func RestoreSessionFile(fileID, sessionID string) (string, error) {
	if globalStore == nil {
		return "", fmt.Errorf("no store initialized")
	}
	stored, err := globalStore.GetFile(fileID)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve file: %w", err)
	}
	if stored == nil {
		return "", fmt.Errorf("file not found: %s", fileID)
	}

	dir := filepath.Join(uploadsDir(), "sessions", sessionID)
	os.MkdirAll(dir, 0755)
	dst := filepath.Join(dir, stored.Name)
	if err := os.WriteFile(dst, stored.Data, 0644); err != nil {
		return "", fmt.Errorf("failed to restore file to disk: %w", err)
	}
	return dst, nil
}

// StartSessionCleanup starts a background goroutine that removes session
// directories older than maxAge, running every interval.
// Note: files are also persisted in the store, so cleanup only affects disk cache.
func StartSessionCleanup(interval, maxAge time.Duration) {
	go func() {
		for {
			time.Sleep(interval)
			sessionsDir := filepath.Join(uploadsDir(), "sessions")
			entries, err := os.ReadDir(sessionsDir)
			if err != nil {
				continue
			}
			cutoff := time.Now().Add(-maxAge)
			for _, e := range entries {
				if !e.IsDir() {
					continue
				}
				info, err := e.Info()
				if err != nil {
					continue
				}
				if info.ModTime().Before(cutoff) {
					os.RemoveAll(filepath.Join(sessionsDir, e.Name()))
				}
			}
		}
	}()
}
