package codegen

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/haira-lang/haira/internal/ast"
	"github.com/haira-lang/haira/internal/checker"
	"github.com/haira-lang/haira/internal/runtime"
)

// CompileToBinary generates Go source and runs go build.
// The embedded runtime is always used — no external runtime path needed.
func CompileToBinary(file *ast.SourceFile, output, hairaFile, hairaSource string, typeInfo ...*checker.TypeInfo) error {
	tmpDir := filepath.Join(os.TempDir(), fmt.Sprintf("haira-build-%d", os.Getpid()))

	if HasTests(file) && !hasMainFunction(file) {
		mainGo := GenerateMainGoForTest(file, hairaFile, hairaSource, typeInfo...)
		testGo := GenerateTestGo(file, hairaFile, hairaSource, typeInfo...)
		if err := writeTestProject(tmpDir, mainGo, testGo); err != nil {
			return err
		}
	} else {
		mainGo := GenerateMainGo(file, hairaFile, hairaSource, typeInfo...)
		if err := writeProject(tmpDir, mainGo); err != nil {
			return err
		}
	}

	if err := runGoModTidy(tmpDir); err != nil {
		return fmt.Errorf("go mod tidy failed: %w", err)
	}

	if err := runGoBuild(tmpDir, output); err != nil {
		fmt.Fprintf(os.Stderr, "Debug: generated Go project at %s\n", tmpDir)
		return fmt.Errorf("%s", cleanGoErrors(err.Error()))
	}

	os.RemoveAll(tmpDir)
	return nil
}

// RunProgram generates Go source and runs go run.
func RunProgram(file *ast.SourceFile, hairaFile, hairaSource string, typeInfo ...*checker.TypeInfo) error {
	tmpDir := filepath.Join(os.TempDir(), fmt.Sprintf("haira-run-%d", os.Getpid()))

	mainGo := GenerateMainGo(file, hairaFile, hairaSource, typeInfo...)
	if err := writeProject(tmpDir, mainGo); err != nil {
		return err
	}
	if err := runGoModTidy(tmpDir); err != nil {
		return fmt.Errorf("go mod tidy failed: %w", err)
	}

	cmd := exec.Command("go", "run", ".")
	cmd.Dir = tmpDir
	cmd.Stdout = os.Stdout
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()

	if stderr.Len() > 0 {
		if err != nil {
			fmt.Fprint(os.Stderr, cleanGoErrors(stderr.String()))
		} else {
			fmt.Fprint(os.Stderr, stderr.String())
		}
	}

	os.RemoveAll(tmpDir)
	return err
}

// RunTests generates Go source and runs go test.
func RunTests(file *ast.SourceFile, hairaFile, hairaSource string, testArgs []string, typeInfo ...*checker.TypeInfo) error {
	tmpDir := filepath.Join(os.TempDir(), fmt.Sprintf("haira-test-%d", os.Getpid()))

	mainGo := GenerateMainGoForTest(file, hairaFile, hairaSource, typeInfo...)
	testGo := GenerateTestGo(file, hairaFile, hairaSource, typeInfo...)

	if testGo == "" {
		return fmt.Errorf("no test blocks found in %s", hairaFile)
	}

	if err := writeTestProject(tmpDir, mainGo, testGo); err != nil {
		return err
	}
	if err := runGoModTidy(tmpDir); err != nil {
		return fmt.Errorf("go mod tidy failed: %w", err)
	}

	args := append([]string{"test", "-v", "."}, normalizeTestArgs(testArgs)...)
	cmd := exec.Command("go", args...)
	cmd.Dir = tmpDir
	cmd.Stdout = os.Stdout
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()

	if stderr.Len() > 0 {
		if err != nil {
			fmt.Fprint(os.Stderr, cleanGoErrors(stderr.String()))
		} else {
			fmt.Fprint(os.Stderr, stderr.String())
		}
	}

	os.RemoveAll(tmpDir)
	return err
}

// writeProject writes a Go project using the embedded runtime.
func writeProject(dir, mainGo string) error {
	hairaDir := filepath.Join(dir, "haira")
	if err := os.MkdirAll(hairaDir, 0o755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}

	// Write all Go runtime source files
	for name, data := range runtime.GoFiles() {
		if err := os.WriteFile(filepath.Join(hairaDir, name), data, 0o644); err != nil {
			return fmt.Errorf("write runtime %s: %w", name, err)
		}
	}

	// Write UI files (dist/, HTML templates)
	for relPath, data := range runtime.UIFiles() {
		fullPath := filepath.Join(hairaDir, relPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return fmt.Errorf("create ui dir: %w", err)
		}
		if err := os.WriteFile(fullPath, data, 0o644); err != nil {
			return fmt.Errorf("write ui %s: %w", relPath, err)
		}
	}

	// Write go.mod — rewrite the module to use the local haira/ directory
	goMod := rewriteGoMod(runtime.GoMod())
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), goMod, 0o644); err != nil {
		return fmt.Errorf("write go.mod: %w", err)
	}

	// Write go.sum
	goSum := runtime.GoSum()
	if goSum != nil {
		if err := os.WriteFile(filepath.Join(dir, "go.sum"), goSum, 0o644); err != nil {
			return fmt.Errorf("write go.sum: %w", err)
		}
	}

	// Write main.go
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(mainGo), 0o644); err != nil {
		return fmt.Errorf("write main.go: %w", err)
	}

	return nil
}

// writeTestProject writes a Go test project using the embedded runtime.
func writeTestProject(dir, mainGo, testGo string) error {
	if err := writeProject(dir, mainGo); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "main_test.go"), []byte(testGo), 0o644); err != nil {
		return fmt.Errorf("write main_test.go: %w", err)
	}
	return nil
}

// rewriteGoMod transforms the embedded go.mod to work in the temp build directory.
// It changes the module name and adds a replace directive for the runtime.
func rewriteGoMod(original []byte) []byte {
	if original == nil {
		// Fallback: generate a minimal go.mod
		return []byte("module haira-generated\n\ngo 1.22\n")
	}

	content := string(original)

	// Replace the module name
	content = strings.Replace(content, "module haira-go-runtime", "module haira-generated", 1)

	// The runtime Go files are in haira/ subdirectory of the build dir,
	// but since we're using the same module (haira-generated), Go resolves
	// the import "haira-generated/haira" to the haira/ subdirectory automatically.
	// No replace directive needed — it's all one module.

	return []byte(content)
}

// normalizeTestArgs translates user-facing test flags to Go's format.
func normalizeTestArgs(args []string) []string {
	out := make([]string, len(args))
	for i, arg := range args {
		out[i] = arg
	}
	for i := 0; i < len(out); i++ {
		if (out[i] == "-run" || out[i] == "--run") && i+1 < len(out) {
			filter := out[i+1]
			filter = strings.ReplaceAll(filter, " ", "_")
			if !strings.HasPrefix(filter, "TestHaira") {
				filter = "TestHaira/" + filter
			}
			out[i+1] = filter
			i++
		}
	}
	return out
}

// cleanGoErrors strips Go module/build noise from error output.
func cleanGoErrors(output string) string {
	var lines []string
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "# haira-generated") {
			continue
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

// ShowGeneratedGo generates Go source and prints it (for debugging).
func ShowGeneratedGo(file *ast.SourceFile, hairaFile, hairaSource string, typeInfo ...*checker.TypeInfo) string {
	return GenerateMainGo(file, hairaFile, hairaSource, typeInfo...)
}

func runGoModTidy(dir string) error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err, string(out))
	}
	return nil
}

func runGoBuild(dir, output string) error {
	outputAbs := output
	if !filepath.IsAbs(output) {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		outputAbs = filepath.Join(cwd, output)
	}

	cmd := exec.Command("go", "build", "-o", outputAbs, ".")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go build failed: %s", string(out))
	}
	return nil
}
