package resolver

import (
	"os"
	"path/filepath"
	"testing"
)

func setupProject(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		path := filepath.Join(dir, name)
		os.MkdirAll(filepath.Dir(path), 0o755)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestResolveStdlibOnly(t *testing.T) {
	dir := setupProject(t, map[string]string{
		"main.haira": `import "io"
fn main() { io.println("hello") }`,
	})

	prog, diags := Resolve(filepath.Join(dir, "main.haira"))
	if len(diags) > 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if len(prog.Modules) != 0 {
		t.Errorf("expected 0 modules for stdlib-only import, got %d", len(prog.Modules))
	}
}

func TestResolveLocalImport(t *testing.T) {
	dir := setupProject(t, map[string]string{
		"main.haira": `import "models/user"

fn main() {
	u = User{name = "Alice"}
}`,
		"models/user.haira": `struct User {
	name: string
	age: int
}`,
	})

	prog, diags := Resolve(filepath.Join(dir, "main.haira"))
	if len(diags) > 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if len(prog.Modules) != 1 {
		t.Fatalf("expected 1 module, got %d", len(prog.Modules))
	}
	mod, ok := prog.Modules["models/user"]
	if !ok {
		t.Fatal("expected 'models/user' module")
	}
	if len(mod.File.Items) == 0 {
		t.Error("expected items in user module")
	}
}

func TestResolveTransitiveImport(t *testing.T) {
	dir := setupProject(t, map[string]string{
		"main.haira": `import "services/auth"

fn main() {
	auth.verify("token")
}`,
		"services/auth.haira": `import "models/user"

fn verify(token: string) -> bool {
	return true
}`,
		"models/user.haira": `struct User {
	name: string
}`,
	})

	prog, diags := Resolve(filepath.Join(dir, "main.haira"))
	if len(diags) > 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if len(prog.Modules) != 2 {
		t.Fatalf("expected 2 modules, got %d", len(prog.Modules))
	}
	if _, ok := prog.Modules["services/auth"]; !ok {
		t.Error("expected 'services/auth' module")
	}
	if _, ok := prog.Modules["models/user"]; !ok {
		t.Error("expected 'models/user' module")
	}
}

func TestCircularImportError(t *testing.T) {
	dir := setupProject(t, map[string]string{
		"main.haira": `import "a"
fn main() {}`,
		"a.haira": `import "b"
fn hello() {}`,
		"b.haira": `import "a"
fn world() {}`,
	})

	_, diags := Resolve(filepath.Join(dir, "main.haira"))
	found := false
	for _, d := range diags {
		if d.Level == 0 {
			for i := 0; i <= len(d.Message)-8; i++ {
				if d.Message[i:i+8] == "circular" {
					found = true
					break
				}
			}
		}
	}
	if !found {
		t.Error("expected circular import error")
	}
}

func TestMissingImportError(t *testing.T) {
	dir := setupProject(t, map[string]string{
		"main.haira": `import "nonexistent"
fn main() {}`,
	})

	_, diags := Resolve(filepath.Join(dir, "main.haira"))
	found := false
	for _, d := range diags {
		if d.Level == 0 {
			for i := 0; i <= len(d.Message)-9; i++ {
				if d.Message[i:i+9] == "not found" {
					found = true
					break
				}
			}
		}
	}
	if !found {
		t.Errorf("expected 'not found' error, got: %v", diags)
	}
}

func TestMergedItems(t *testing.T) {
	dir := setupProject(t, map[string]string{
		"main.haira": `import "models/user"

fn main() {
	u = User{name = "Alice"}
}`,
		"models/user.haira": `struct User {
	name: string
}`,
	})

	prog, diags := Resolve(filepath.Join(dir, "main.haira"))
	if len(diags) > 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	merged := prog.MergedItems()
	// Should have: User typedef from module + import + fn main from main
	if len(merged) < 2 {
		t.Errorf("expected at least 2 merged items, got %d", len(merged))
	}
}

func TestModuleName(t *testing.T) {
	cases := map[string]string{
		"io":            "io",
		"models/user":   "user",
		"services/auth": "auth",
		"a/b/c":         "c",
	}
	for input, want := range cases {
		if got := ModuleName(input); got != want {
			t.Errorf("ModuleName(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestDirectoryModule(t *testing.T) {
	dir := setupProject(t, map[string]string{
		"main.haira": `import "utils"
fn main() {}`,
		"utils/mod.haira": `fn helper() -> int { return 42 }`,
	})

	prog, diags := Resolve(filepath.Join(dir, "main.haira"))
	if len(diags) > 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if _, ok := prog.Modules["utils"]; !ok {
		t.Error("expected 'utils' module from utils/mod.haira")
	}
}

func TestParseErrorInModule(t *testing.T) {
	dir := setupProject(t, map[string]string{
		"main.haira": `import "broken"
fn main() {}`,
		"broken.haira": `fn {bad syntax`,
	})

	_, diags := Resolve(filepath.Join(dir, "main.haira"))
	if len(diags) == 0 {
		t.Error("expected parse errors from broken module")
	}
}
