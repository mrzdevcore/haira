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
)

// CompileToBinary generates Go source and runs go build.
func CompileToBinary(file *ast.SourceFile, output, runtimePath, hairaFile, hairaSource string, typeInfo ...*checker.TypeInfo) error {
	tmpDir := filepath.Join(os.TempDir(), fmt.Sprintf("haira-build-%d", os.Getpid()))

	mainGo := GenerateMainGo(file, hairaFile, hairaSource, typeInfo...)

	if err := writeProject(tmpDir, mainGo, runtimePath); err != nil {
		return err
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
func RunProgram(file *ast.SourceFile, runtimePath, hairaFile, hairaSource string, typeInfo ...*checker.TypeInfo) error {
	tmpDir := filepath.Join(os.TempDir(), fmt.Sprintf("haira-run-%d", os.Getpid()))

	mainGo := GenerateMainGo(file, hairaFile, hairaSource, typeInfo...)

	if err := writeProject(tmpDir, mainGo, runtimePath); err != nil {
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

// cleanGoErrors strips Go module/build noise from error output,
// since //line directives already point to the Haira source file.
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

func writeProject(dir, mainGo, runtimePath string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}

	absRuntime, err := filepath.Abs(runtimePath)
	if err != nil {
		return fmt.Errorf("resolve runtime path: %w", err)
	}

	goMod := fmt.Sprintf("module haira-generated\n\ngo 1.22\n\nrequire haira-go-runtime v0.0.0\n\nreplace haira-go-runtime => %s\n", absRuntime)
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0o644); err != nil {
		return fmt.Errorf("write go.mod: %w", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(mainGo), 0o644); err != nil {
		return fmt.Errorf("write main.go: %w", err)
	}

	return nil
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

// FindRuntimePath locates the go-runtime directory.
func FindRuntimePath() string {
	// Try relative to current exe
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(filepath.Dir(filepath.Dir(exe)))
		runtime := filepath.Join(dir, "go-runtime")
		if info, err := os.Stat(runtime); err == nil && info.IsDir() {
			return runtime
		}
	}

	// Try current working directory
	if cwd, err := os.Getwd(); err == nil {
		runtime := filepath.Join(cwd, "go-runtime")
		if info, err := os.Stat(runtime); err == nil && info.IsDir() {
			return runtime
		}
		// Try parent directory (compiler is a subdirectory)
		runtime = filepath.Join(filepath.Dir(cwd), "go-runtime")
		if info, err := os.Stat(runtime); err == nil && info.IsDir() {
			return runtime
		}
	}

	return ""
}
