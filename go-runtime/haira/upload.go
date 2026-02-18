package haira

import (
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

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
// across requests. Path: /tmp/haira-uploads/sessions/{sessionID}/{filename}
func saveSessionFile(src multipart.File, originalName, sessionID string) (string, error) {
	if !safeSessionID.MatchString(sessionID) {
		return "", os.ErrInvalid
	}
	dir := filepath.Join(os.TempDir(), "haira-uploads", "sessions", sessionID)
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
	return dst, nil
}

// StartSessionCleanup starts a background goroutine that removes session
// directories older than maxAge, running every interval.
func StartSessionCleanup(interval, maxAge time.Duration) {
	go func() {
		for {
			time.Sleep(interval)
			sessionsDir := filepath.Join(os.TempDir(), "haira-uploads", "sessions")
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
