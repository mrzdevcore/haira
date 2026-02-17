package haira

import (
	"io"
	"os"
	"path/filepath"
)

// FsReadFile reads the contents of a file and returns it as a string.
func FsReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FsWriteFile writes content to a file, creating it if necessary.
func FsWriteFile(path string, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

// FsAppendFile appends content to a file, creating it if necessary.
func FsAppendFile(path string, content string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(content)
	return err
}

// FsExists returns true if the path exists.
func FsExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// FsRemove deletes a file.
func FsRemove(path string) error {
	return os.Remove(path)
}

// FsRemoveAll removes a path and all its children.
func FsRemoveAll(path string) error {
	return os.RemoveAll(path)
}

// FsRename renames (moves) a file or directory.
func FsRename(oldPath string, newPath string) error {
	return os.Rename(oldPath, newPath)
}

// FsCopy copies a file from src to dst.
func FsCopy(src string, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

// FsMkdir creates a directory.
func FsMkdir(path string) error {
	return os.Mkdir(path, 0755)
}

// FsMkdirAll creates a directory and all parent directories.
func FsMkdirAll(path string) error {
	return os.MkdirAll(path, 0755)
}

// FsReadDir lists directory contents as a slice of maps.
func FsReadDir(path string) ([]map[string]any, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var result []map[string]any
	for _, entry := range entries {
		info, _ := entry.Info()
		size := int64(0)
		modified := ""
		if info != nil {
			size = info.Size()
			modified = info.ModTime().Format("2006-01-02T15:04:05Z07:00")
		}
		result = append(result, map[string]any{
			"name":     entry.Name(),
			"is_dir":   entry.IsDir(),
			"size":     size,
			"modified": modified,
		})
	}
	return result, nil
}

// FsStat returns file metadata as a map.
func FsStat(path string) (map[string]any, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	absPath, _ := filepath.Abs(path)
	return map[string]any{
		"name":     info.Name(),
		"size":     info.Size(),
		"is_dir":   info.IsDir(),
		"modified": info.ModTime().Format("2006-01-02T15:04:05Z07:00"),
		"path":     absPath,
	}, nil
}
