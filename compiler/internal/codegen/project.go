package codegen

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/haira-lang/haira/internal/ast"
	"github.com/haira-lang/haira/internal/checker"
	"github.com/haira-lang/haira/internal/runtime"
)

// CompileToBinary generates Go source and runs go build.
// The embedded runtime is always used — no external runtime path needed.
// target is "native" (default) or "workers" (Cloudflare Workers WASM).
func CompileToBinary(file *ast.SourceFile, output, hairaFile, hairaSource, target string, typeInfo ...*checker.TypeInfo) error {
	activeTarget = target
	defer func() { activeTarget = "native" }()

	tmpDir := filepath.Join(os.TempDir(), fmt.Sprintf("haira-build-%d", os.Getpid()))
	usedPkgs := collectUsedStdlibPackages(file)

	// Workers target: replace sqlite with d1 (no filesystem in WASM, use Cloudflare D1)
	if target == "workers" {
		filtered := make([]string, 0, len(usedPkgs))
		hasSqlite := false
		for _, pkg := range usedPkgs {
			if pkg == "sqlite" {
				hasSqlite = true
			} else {
				filtered = append(filtered, pkg)
			}
		}
		if hasSqlite {
			filtered = append(filtered, "d1")
		}
		usedPkgs = filtered
	}

	if HasTests(file) && !hasMainFunction(file) {
		mainGo := GenerateMainGoForTest(file, hairaFile, hairaSource, typeInfo...)
		testGo := GenerateTestGo(file, hairaFile, hairaSource, typeInfo...)
		if err := writeTestProject(tmpDir, mainGo, testGo, usedPkgs); err != nil {
			return err
		}
	} else {
		mainGo := GenerateMainGo(file, hairaFile, hairaSource, typeInfo...)
		if err := writeProject(tmpDir, mainGo, usedPkgs); err != nil {
			return err
		}
	}

	// Workers target: add syumai/workers dependency
	if target == "workers" {
		if err := addWorkersGoModRequire(tmpDir); err != nil {
			return fmt.Errorf("add workers dependency: %w", err)
		}
	}

	if err := runGoModTidy(tmpDir); err != nil {
		return fmt.Errorf("go mod tidy failed: %w", err)
	}

	// Fork: workers target builds WASM + scaffolds deploy directory
	if target == "workers" {
		return buildWorkersProject(tmpDir, output, hairaFile)
	}

	if err := runGoBuild(tmpDir, output); err != nil {
		return fmt.Errorf("%s", cleanGoErrors(err.Error()))
	}

	os.RemoveAll(tmpDir)
	return nil
}

// RunProgram generates Go source and runs go run.
func RunProgram(file *ast.SourceFile, hairaFile, hairaSource string, typeInfo ...*checker.TypeInfo) error {
	tmpDir := filepath.Join(os.TempDir(), fmt.Sprintf("haira-run-%d", os.Getpid()))
	usedPkgs := collectUsedStdlibPackages(file)

	mainGo := GenerateMainGo(file, hairaFile, hairaSource, typeInfo...)
	if err := writeProject(tmpDir, mainGo, usedPkgs); err != nil {
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
	usedPkgs := collectUsedStdlibPackages(file)

	mainGo := GenerateMainGoForTest(file, hairaFile, hairaSource, typeInfo...)
	testGo := GenerateTestGo(file, hairaFile, hairaSource, typeInfo...)

	if testGo == "" {
		return fmt.Errorf("no test blocks found in %s", hairaFile)
	}

	if err := writeTestProject(tmpDir, mainGo, testGo, usedPkgs); err != nil {
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
// usedStdlibPkgs lists the stdlib package names (e.g., "postgres", "excel") to include.
func writeProject(dir, mainGo string, usedStdlibPkgs []string) error {
	// Write core haira/ package (always included)
	hairaDir := filepath.Join(dir, "haira")
	if err := os.MkdirAll(hairaDir, 0o755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}
	for name, data := range runtime.GoFilesForPackage("haira") {
		// Rewrite import paths from dev module to generated module (e.g., haira-go-runtime/arp → haira-generated/arp)
		content := strings.ReplaceAll(string(data), "haira-go-runtime/", "haira-generated/")
		if err := os.WriteFile(filepath.Join(hairaDir, name), []byte(content), 0o644); err != nil {
			return fmt.Errorf("write runtime %s: %w", name, err)
		}
	}

	// Write UI files (observe.html, loader.html, haira-ui.js)
	// Workers target: write minimal placeholders (real UI served via Cloudflare Static Assets)
	for relPath, data := range runtime.UIFiles() {
		fullPath := filepath.Join(hairaDir, relPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return fmt.Errorf("create ui dir: %w", err)
		}
		if activeTarget == "workers" {
			// Minimal placeholder so //go:embed compiles, but content stays tiny
			data = []byte("<!-- served via static assets -->")
		}
		if err := os.WriteFile(fullPath, data, 0o644); err != nil {
			return fmt.Errorf("write ui %s: %w", relPath, err)
		}
	}

	// Write arp/ package (always included — haira/arp.go depends on it)
	arpDir := filepath.Join(dir, "arp")
	if err := os.MkdirAll(arpDir, 0o755); err != nil {
		return fmt.Errorf("create arp dir: %w", err)
	}
	for name, data := range runtime.GoFilesForPackage("arp") {
		if err := os.WriteFile(filepath.Join(arpDir, name), data, 0o644); err != nil {
			return fmt.Errorf("write arp/%s: %w", name, err)
		}
	}

	// Write used stdlib packages (rewrite module imports to match generated module name)
	for _, pkg := range usedStdlibPkgs {
		pkgDir := filepath.Join(dir, pkg)
		if err := os.MkdirAll(pkgDir, 0o755); err != nil {
			return fmt.Errorf("create dir %s: %w", pkg, err)
		}
		for name, data := range runtime.GoFilesForPackage(pkg) {
			// Rewrite import paths from dev module to generated module
			content := strings.ReplaceAll(string(data), "haira-go-runtime/", "haira-generated/")
			if err := os.WriteFile(filepath.Join(pkgDir, name), []byte(content), 0o644); err != nil {
				return fmt.Errorf("write %s/%s: %w", pkg, name, err)
			}
		}
	}

	// Write go.mod — rewrite the module to use the local directory structure
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
func writeTestProject(dir, mainGo, testGo string, usedStdlibPkgs []string) error {
	if err := writeProject(dir, mainGo, usedStdlibPkgs); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "main_test.go"), []byte(testGo), 0o644); err != nil {
		return fmt.Errorf("write main_test.go: %w", err)
	}
	return nil
}

// collectUsedStdlibPackages determines which stdlib packages to include based on
// the Haira source file's imports, workflow usage, and transitive dependencies.
func collectUsedStdlibPackages(file *ast.SourceFile) []string {
	pkgs := map[string]bool{}

	// Scan imports for stdlib packages that map to separate Go packages
	for _, item := range file.Items {
		if imp, ok := item.Node.(ast.ImportDecl); ok {
			if goImport, ok := stdlibGoImport(imp.Path); ok {
				pkgs[goImport] = true
			}
		}
	}

	// If any workflow or http.Server is used, include sqlite (default store backend)
	if hasWorkflows(file) || hasServerCall(file) {
		pkgs["sqlite"] = true
	}

	// Resolve transitive dependencies
	addTransitiveDeps(pkgs)

	var result []string
	for pkg := range pkgs {
		result = append(result, pkg)
	}
	sort.Strings(result)
	return result
}

// stdlibGoImport maps a Haira import path to its Go package directory name.
// Returns the Go package name and true if it's a separate stdlib package,
// or empty string and false if it's a core module.
func stdlibGoImport(path string) (string, bool) {
	switch path {
	case "postgres":
		return "postgres", true
	case "excel":
		return "excel", true
	case "vector":
		return "vector", true
	case "slack":
		return "slack", true
	case "github":
		return "github", true
	case "gitlab":
		return "gitlab", true
	case "langfuse":
		return "langfuse", true
	case "algolia":
		return "algolia", true
	case "meilisearch":
		return "meilisearch", true
	case "auth":
		return "auth", true
	case "websearch":
		return "websearch", true
	case "healthcheck":
		return "healthcheck", true
	}
	return "", false
}

// addTransitiveDeps adds packages required by already-included packages.
func addTransitiveDeps(pkgs map[string]bool) {
	// vector requires postgres (for pgvector DB type and QuoteIdentifier)
	if pkgs["vector"] {
		pkgs["postgres"] = true
	}
	// excel references postgres types (PgSchema) for validation
	if pkgs["excel"] {
		pkgs["postgres"] = true
	}
}

// hasWorkflows returns true if the file contains any workflow declarations.
func hasWorkflows(file *ast.SourceFile) bool {
	for _, item := range file.Items {
		if _, ok := item.Node.(ast.WorkflowDecl); ok {
			return true
		}
	}
	return false
}

// hasServerCall returns true if the file contains http.Server() or mcp.Server() calls.
func hasServerCall(file *ast.SourceFile) bool {
	for _, item := range file.Items {
		if fn, ok := item.Node.(ast.FunctionDef); ok {
			if blockHasServerCall(fn.Body) {
				return true
			}
		}
	}
	return false
}

func blockHasServerCall(block ast.Block) bool {
	for _, stmt := range block.Statements {
		if exprStmt, ok := stmt.Node.(ast.ExprStmt); ok {
			if hasServerExpr(exprStmt.Value) {
				return true
			}
		}
		if assign, ok := stmt.Node.(ast.AssignStmt); ok {
			if hasServerExpr(assign.Value) {
				return true
			}
		}
		if letStmt, ok := stmt.Node.(ast.LetStmt); ok {
			if hasServerExpr(letStmt.Value) {
				return true
			}
		}
	}
	return false
}

func hasServerExpr(expr ast.Expr) bool {
	if call, ok := expr.Node.(ast.CallExpr); ok {
		if field, ok := call.Callee.Node.(ast.FieldExpr); ok {
			if ident, ok := field.Object.Node.(ast.IdentExpr); ok {
				if (ident.Name == "http" || ident.Name == "mcp") && field.Field.Node == "Server" {
					return true
				}
			}
		}
	}
	return false
}

// rewriteGoMod transforms the embedded go.mod to work in the temp build directory.
// It changes the module name — stdlib packages resolve automatically as subdirectories.
func rewriteGoMod(original []byte) []byte {
	if original == nil {
		// Fallback: generate a minimal go.mod
		return []byte("module haira-generated\n\ngo 1.22\n")
	}

	content := string(original)

	// Replace the module name
	content = strings.Replace(content, "module haira-go-runtime", "module haira-generated", 1)

	// The runtime Go files are in haira/ subdirectory of the build dir,
	// and stdlib packages in postgres/, excel/, etc. subdirectories.
	// Since we're using the same module (haira-generated), Go resolves
	// imports like "haira-generated/haira" and "haira-generated/postgres"
	// to their respective subdirectories automatically.

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

// --- Workers target (Cloudflare Workers WASM) ---

const workersVersion = "v0.26.3"

// addWorkersGoModRequire appends the syumai/workers dependency to go.mod.
func addWorkersGoModRequire(dir string) error {
	modPath := filepath.Join(dir, "go.mod")
	content, err := os.ReadFile(modPath)
	if err != nil {
		return err
	}
	addition := fmt.Sprintf("\nrequire github.com/syumai/workers %s\n", workersVersion)
	return os.WriteFile(modPath, append(content, []byte(addition)...), 0o644)
}

// buildWorkersProject compiles to WASM and generates a deploy-ready Cloudflare Workers directory.
func buildWorkersProject(tmpDir, output, hairaFile string) error {
	// Step 1: Generate workers JS assets (worker.mjs, wasm_exec.js)
	buildDir := filepath.Join(tmpDir, "build")
	if err := os.MkdirAll(buildDir, 0o755); err != nil {
		return err
	}
	assetsCmd := exec.Command("go", "run", "github.com/syumai/workers/cmd/workers-assets-gen@"+workersVersion,
		"-mode=go", "-o", buildDir)
	assetsCmd.Dir = tmpDir
	var assetsStderr bytes.Buffer
	assetsCmd.Stderr = &assetsStderr
	if err := assetsCmd.Run(); err != nil {
		return fmt.Errorf("workers-assets-gen failed: %s\n%s", err, assetsStderr.String())
	}

	// Step 2: Build WASM binary
	wasmOutput := filepath.Join(buildDir, "app.wasm")
	cmd := exec.Command("go", "build", "-ldflags", "-s -w", "-o", wasmOutput, ".")
	cmd.Dir = tmpDir
	cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("WASM build failed: %s", cleanGoErrors(string(out)))
	}

	// Step 3: Create output directory
	outputAbs := output
	if !filepath.IsAbs(output) {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		outputAbs = filepath.Join(cwd, output)
	}
	outBuildDir := filepath.Join(outputAbs, "build")
	if err := os.MkdirAll(outBuildDir, 0o755); err != nil {
		return err
	}

	// Step 4: Copy build artifacts
	entries, err := os.ReadDir(buildDir)
	if err != nil {
		return fmt.Errorf("read build dir: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(buildDir, entry.Name()))
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(outBuildDir, entry.Name()), data, 0o644); err != nil {
			return err
		}
	}

	// Step 5: Copy real UI files to public/ for Cloudflare Static Assets
	publicDir := filepath.Join(outputAbs, "public")
	if err := os.MkdirAll(publicDir, 0o755); err != nil {
		return err
	}
	for relPath, data := range runtime.UIFiles() {
		name := filepath.Base(relPath)
		if strings.HasPrefix(name, "._") {
			continue // skip macOS resource fork files
		}
		// Map: ui/loader.html → public/index.html, ui/haira-ui.js → public/haira-ui.js, etc.
		if name == "loader.html" {
			name = "index.html" // serve as SPA entry point
		}
		if err := os.WriteFile(filepath.Join(publicDir, name), data, 0o644); err != nil {
			return fmt.Errorf("write public/%s: %w", name, err)
		}
	}

	// Step 6: Generate wrangler.toml (with D1 database binding + static assets)
	projectName := deriveProjectName(hairaFile)
	dbName := projectName + "-db"
	wranglerContent := fmt.Sprintf(`name = %q
main = "./build/worker.mjs"
compatibility_date = "2024-09-23"

[assets]
directory = "./public"

[vars]
HAIRA_UI_URL = "/haira-ui.js"

[build]
command = ""

[[d1_databases]]
binding = "DB"
database_name = %q
database_id = "<run: npx wrangler d1 create %s>"
`, projectName, dbName, dbName)
	if err := os.WriteFile(filepath.Join(outputAbs, "wrangler.toml"), []byte(wranglerContent), 0o644); err != nil {
		return err
	}

	// Step 7: Generate package.json
	packageJSON := fmt.Sprintf("{\n  \"name\": %q,\n  \"version\": \"0.0.0\",\n  \"private\": true,\n  \"scripts\": {\n    \"deploy\": \"wrangler deploy\",\n    \"dev\": \"wrangler dev\"\n  },\n  \"devDependencies\": {\n    \"wrangler\": \"^3.109.2\"\n  }\n}\n", projectName)
	if err := os.WriteFile(filepath.Join(outputAbs, "package.json"), []byte(packageJSON), 0o644); err != nil {
		return err
	}

	os.RemoveAll(tmpDir)
	return nil
}

// deriveProjectName extracts a Cloudflare Workers project name from the Haira source path.
func deriveProjectName(hairaFile string) string {
	base := filepath.Base(hairaFile)
	ext := filepath.Ext(base)
	if ext != "" {
		base = base[:len(base)-len(ext)]
	}
	return strings.ReplaceAll(base, "_", "-")
}
