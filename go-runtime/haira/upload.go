package haira

import (
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
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
