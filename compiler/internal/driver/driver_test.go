package driver

import (
	"os"
	"path/filepath"
	"testing"
)

// ---------------------------------------------------------------------------
// ParseFile / CheckFile / LexFile — file-based pipeline tests
// ---------------------------------------------------------------------------

func TestParseFileMissing(t *testing.T) {
	err := ParseFile("/nonexistent/file.haira")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestCheckFileMissing(t *testing.T) {
	err := CheckFile("/nonexistent/file.haira")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLexFileMissing(t *testing.T) {
	err := LexFile("/nonexistent/file.haira")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestParseFileValid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.haira")
	if err := os.WriteFile(path, []byte(`fn main() { io.println("hello") }`), 0o644); err != nil {
		t.Fatal(err)
	}
	err := ParseFile(path)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCheckFileValid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.haira")
	if err := os.WriteFile(path, []byte(`fn main() { io.println("hello") }`), 0o644); err != nil {
		t.Fatal(err)
	}
	err := CheckFile(path)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLexFileValid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.haira")
	if err := os.WriteFile(path, []byte(`fn main() { }`), 0o644); err != nil {
		t.Fatal(err)
	}
	err := LexFile(path)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestParseFileWithErrors(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.haira")
	if err := os.WriteFile(path, []byte(`fn main( {`), 0o644); err != nil {
		t.Fatal(err)
	}
	err := ParseFile(path)
	if err == nil {
		t.Error("expected error for invalid syntax")
	}
}

func TestFormatFileValid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "fmt.haira")
	src := `fn main() {
    io.println("hello")
}`
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	err := FormatFile(path)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestFormatFileMissing(t *testing.T) {
	err := FormatFile("/nonexistent/file.haira")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestFormatFileWithErrors(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.haira")
	if err := os.WriteFile(path, []byte(`fn main( {`), 0o644); err != nil {
		t.Fatal(err)
	}
	err := FormatFile(path)
	if err == nil {
		t.Error("expected error for invalid syntax")
	}
}

// ---------------------------------------------------------------------------
// toDiagnostics helper
// ---------------------------------------------------------------------------

func TestEmitFileMissing(t *testing.T) {
	err := EmitFile("/nonexistent/file.haira")
	if err == nil {
		t.Error("expected error for missing file")
	}
}
