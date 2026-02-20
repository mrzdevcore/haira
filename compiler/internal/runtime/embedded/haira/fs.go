package haira

import (
	"io"
	"os"
	"path/filepath"
)

func FsReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func FsWriteFile(path string, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

func FsAppendFile(path string, content string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(content)
	return err
}

func FsExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func FsRemove(path string) error     { return os.Remove(path) }
func FsRemoveAll(path string) error  { return os.RemoveAll(path) }
func FsRename(old, new string) error { return os.Rename(old, new) }
func FsMkdir(path string) error      { return os.Mkdir(path, 0755) }
func FsMkdirAll(path string) error   { return os.MkdirAll(path, 0755) }

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
