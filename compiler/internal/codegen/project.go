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

// RunTests generates Go source and runs go test.
func RunTests(file *ast.SourceFile, runtimePath, hairaFile, hairaSource string, testArgs []string, typeInfo ...*checker.TypeInfo) error {
	tmpDir := filepath.Join(os.TempDir(), fmt.Sprintf("haira-test-%d", os.Getpid()))

	mainGo := GenerateMainGoForTest(file, hairaFile, hairaSource, typeInfo...)
	testGo := GenerateTestGo(file, hairaFile, hairaSource, typeInfo...)

	if testGo == "" {
		return fmt.Errorf("no test blocks found in %s", hairaFile)
	}

	if err := writeTestProject(tmpDir, mainGo, testGo, runtimePath); err != nil {
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

func writeTestProject(dir, mainGo, testGo, runtimePath string) error {
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

	if err := os.WriteFile(filepath.Join(dir, "main_test.go"), []byte(testGo), 0o644); err != nil {
		return fmt.Errorf("write main_test.go: %w", err)
	}

	return nil
}

// normalizeTestArgs translates user-facing test flags to Go's format.
// Go's t.Run replaces spaces with underscores, so -run "multiply works"
// needs to become -run "TestHaira/multiply_works" for go test.
func normalizeTestArgs(args []string) []string {
	out := make([]string, len(args))
	for i, arg := range args {
		out[i] = arg
	}
	for i := 0; i < len(out); i++ {
		if (out[i] == "-run" || out[i] == "--run") && i+1 < len(out) {
			filter := out[i+1]
			// Replace spaces with underscores (Go's t.Run behavior)
			filter = strings.ReplaceAll(filter, " ", "_")
			// Prefix with TestHaira/ so users just write the test name
			if !strings.HasPrefix(filter, "TestHaira") {
				filter = "TestHaira/" + filter
			}
			out[i+1] = filter
			i++ // skip the value
		}
	}
	return out
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

// isValidRuntime checks that the runtime directory contains the required
// embedded assets (e.g. the built UI bundle) so go:embed won't fail.
func isValidRuntime(path string) bool {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return false
	}
	// server.go embeds ui/dist/haira-ui.js — must exist
	_, err = os.Stat(filepath.Join(path, "haira", "ui", "dist", "haira-ui.js"))
	return err == nil
}

// FindRuntimePath locates the go-runtime directory.
func FindRuntimePath() string {
	// 1. Explicit override via environment variable
	if envPath := os.Getenv("HAIRA_RUNTIME"); envPath != "" {
		if isValidRuntime(envPath) {
			return envPath
		}
	}

	// 2. User install location (~/.haira/runtime/)
	if home, err := os.UserHomeDir(); err == nil {
		runtime := filepath.Join(home, ".haira", "runtime")
		if isValidRuntime(runtime) {
			return runtime
		}
	}

	// 3. Relative to executable (installed locations)
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)

		// <exe_dir>/../lib/haira/runtime (system install: /usr/local/lib/haira/runtime)
		runtime := filepath.Join(exeDir, "..", "lib", "haira", "runtime")
		if isValidRuntime(runtime) {
			return runtime
		}

		// <exe_dir>/runtime (same-directory install)
		runtime = filepath.Join(exeDir, "runtime")
		if isValidRuntime(runtime) {
			return runtime
		}

		// Dev: <exe_dir>/../../go-runtime (running from compiler/)
		runtime = filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(exe))), "go-runtime")
		if isValidRuntime(runtime) {
			return runtime
		}
	}

	// 4. Current working directory (development)
	if cwd, err := os.Getwd(); err == nil {
		runtime := filepath.Join(cwd, "go-runtime")
		if isValidRuntime(runtime) {
			return runtime
		}
		// Parent directory (compiler is a subdirectory)
		runtime = filepath.Join(filepath.Dir(cwd), "go-runtime")
		if isValidRuntime(runtime) {
			return runtime
		}
	}

	return ""
}
