package haira

import (
	"encoding/json"
	"fmt"
	"mime"
	"os"
	"path/filepath"
)

// StorageTools returns built-in tool definitions for session file storage.
// These tools allow agents to persist and retrieve artifacts across turns.
func StorageTools(sessionID string) []*ToolDef {
	return []*ToolDef{
		{
			Name:        "store_artifact",
			Description: "Save a generated artifact (SQL, text, config, etc.) to persistent session storage. Use this to save important outputs so they survive across conversation turns and can be re-downloaded.",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"name": {"type": "string", "description": "File name for the artifact (e.g. 'migration.sql', 'report.txt')"},
					"content": {"type": "string", "description": "The text content to store"},
					"content_type": {"type": "string", "description": "MIME type (default: text/plain)", "default": "text/plain"}
				},
				"required": ["name", "content"]
			}`),
			Handler: func(args json.RawMessage) (any, error) {
				var p struct {
					Name        string `json:"name"`
					Content     string `json:"content"`
					ContentType string `json:"content_type"`
				}
				if err := json.Unmarshal(args, &p); err != nil {
					return nil, fmt.Errorf("invalid arguments: %w", err)
				}
				if p.ContentType == "" {
					p.ContentType = "text/plain"
				}
				id, err := StoreSessionFile(sessionID, p.Name, p.ContentType, []byte(p.Content))
				if err != nil {
					return nil, fmt.Errorf("failed to store artifact: %w", err)
				}
				return map[string]any{
					"id":   id,
					"name": p.Name,
					"size": len(p.Content),
				}, nil
			},
		},
		{
			Name:        "list_session_files",
			Description: "List all files stored in this session (uploaded files and saved artifacts). Returns file IDs, names, sizes, and timestamps. Use to check what files are available before trying to access them.",
			Parameters:  json.RawMessage(`{"type": "object", "properties": {}}`),
			Handler: func(args json.RawMessage) (any, error) {
				files, err := ListSessionFiles(sessionID)
				if err != nil {
					return nil, fmt.Errorf("failed to list files: %w", err)
				}
				return files, nil
			},
		},
		{
			Name:        "get_artifact",
			Description: "Retrieve the content of a previously stored artifact by ID. Use after list_session_files to get the actual content.",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"id": {"type": "string", "description": "The file ID returned by store_artifact or list_session_files"}
				},
				"required": ["id"]
			}`),
			Handler: func(args json.RawMessage) (any, error) {
				var p struct {
					ID string `json:"id"`
				}
				if err := json.Unmarshal(args, &p); err != nil {
					return nil, fmt.Errorf("invalid arguments: %w", err)
				}
				file, err := GetSessionFile(p.ID)
				if err != nil {
					return nil, fmt.Errorf("failed to get file: %w", err)
				}
				if file == nil {
					return nil, fmt.Errorf("file not found: %s", p.ID)
				}
				return map[string]any{
					"id":           file.ID,
					"name":         file.Name,
					"content_type": file.ContentType,
					"size":         file.Size,
					"content":      string(file.Data),
				}, nil
			},
		},
		{
			Name:        "restore_file",
			Description: "Restore a previously uploaded file from storage to disk. Use when a tool needs a file_path but the original file may have been cleaned up. Returns the restored file path.",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"id": {"type": "string", "description": "The file ID from list_session_files"}
				},
				"required": ["id"]
			}`),
			Handler: func(args json.RawMessage) (any, error) {
				var p struct {
					ID string `json:"id"`
				}
				if err := json.Unmarshal(args, &p); err != nil {
					return nil, fmt.Errorf("invalid arguments: %w", err)
				}
				path, err := RestoreSessionFile(p.ID, sessionID)
				if err != nil {
					return nil, fmt.Errorf("failed to restore file: %w", err)
				}
				return map[string]any{
					"file_path": path,
				}, nil
			},
		},
	}
}

// RegisterStorageTools registers built-in storage tools on an agent's tool registry.
// Called during agent initialization when the agent has session-scoped storage enabled.
func RegisterStorageTools(registry *ToolRegistry, sessionID string) {
	for _, tool := range StorageTools(sessionID) {
		registry.Register(tool)
	}
}

// StoreArtifact is a convenience function for storing a text artifact from Go code.
func StoreArtifact(sessionID, name, content, contentType string) (string, error) {
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(name))
		if contentType == "" {
			contentType = "text/plain"
		}
	}
	return StoreSessionFile(sessionID, name, contentType, []byte(content))
}

// GetArtifactContent retrieves the text content of a stored artifact.
func GetArtifactContent(fileID string) (string, error) {
	file, err := GetSessionFile(fileID)
	if err != nil {
		return "", err
	}
	if file == nil {
		return "", fmt.Errorf("file not found: %s", fileID)
	}
	return string(file.Data), nil
}

// EnsureSessionFile checks if a file exists on disk and restores it from the store if not.
// Returns the file path (original or restored).
func EnsureSessionFile(filePath, sessionID string) (string, error) {
	// Check if file exists on disk
	if _, err := os.Stat(filePath); err == nil {
		return filePath, nil // still on disk
	}

	// File is gone — try to find it in the store by name
	name := filepath.Base(filePath)
	files, err := ListSessionFiles(sessionID)
	if err != nil {
		return "", fmt.Errorf("cannot list session files: %w", err)
	}
	for _, f := range files {
		if f.Name == name {
			restored, err := RestoreSessionFile(f.ID, sessionID)
			if err != nil {
				return "", fmt.Errorf("cannot restore file %q: %w", name, err)
			}
			return restored, nil
		}
	}
	return "", fmt.Errorf("file %q not found on disk or in store", name)
}
