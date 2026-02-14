// Package resolver handles multi-file module resolution for Haira.
//
// Import rules:
//   - Stdlib imports ("io", "http", "json", etc.) are handled by codegen — no file resolution needed.
//   - Project-local imports ("models/user") resolve to <project_root>/models/user.haira.
//   - The project root is the directory containing the main file.
//   - Imported symbols are available as module.Symbol (e.g., user.User, auth.verify).
//   - Circular imports are detected and reported as errors.
package resolver

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/haira-lang/haira/internal/ast"
	hairaerr "github.com/haira-lang/haira/internal/errors"
	"github.com/haira-lang/haira/internal/parser"
)

// Program represents a resolved multi-file Haira program.
type Program struct {
	Main    *ast.SourceFile
	Modules map[string]*Module // import path → parsed module
}

// Module represents a single imported file.
type Module struct {
	Path     string          // import path (e.g., "models/user")
	FilePath string          // absolute file path
	File     *ast.SourceFile // parsed AST
}

// stdlibModules are import paths that map to runtime support, not files.
var stdlibModules = map[string]bool{
	"io": true, "http": true, "env": true, "json": true,
	"postgres": true, "slack": true, "excel": true, "time": true,
}

// Resolve parses the main file and all its transitive project-local imports.
// Returns a Program with the main file and all resolved modules, plus any diagnostics.
func Resolve(mainFile string) (*Program, []hairaerr.Diagnostic) {
	absMain, err := filepath.Abs(mainFile)
	if err != nil {
		return nil, []hairaerr.Diagnostic{{
			Level:   hairaerr.Error,
			Message: fmt.Sprintf("cannot resolve path: %s", err),
			File:    mainFile,
		}}
	}

	projectRoot := filepath.Dir(absMain)

	r := &resolver{
		projectRoot: projectRoot,
		parsed:      make(map[string]*Module),
		inProgress:  make(map[string]bool),
	}

	// Parse main file
	mainSource, err := os.ReadFile(absMain)
	if err != nil {
		return nil, []hairaerr.Diagnostic{{
			Level:   hairaerr.Error,
			Message: fmt.Sprintf("cannot read file: %s", err),
			File:    mainFile,
		}}
	}

	mainSrc := string(mainSource)
	mainAST, parseErrs := parser.Parse(mainSrc)
	if len(parseErrs) > 0 {
		diags := make([]hairaerr.Diagnostic, len(parseErrs))
		for i, e := range parseErrs {
			diags[i] = hairaerr.Diagnostic{
				Level:   hairaerr.Error,
				Message: e.Message,
				Span:    e.Span,
				File:    mainFile,
			}
		}
		return nil, diags
	}

	// Resolve imports from main file
	r.resolveImports(mainAST, absMain)

	prog := &Program{
		Main:    mainAST,
		Modules: r.parsed,
	}

	return prog, r.diags
}

type resolver struct {
	projectRoot string
	parsed      map[string]*Module
	inProgress  map[string]bool
	diags       []hairaerr.Diagnostic
}

func (r *resolver) addError(msg, file string, span ast.Span) {
	r.diags = append(r.diags, hairaerr.Diagnostic{
		Level:   hairaerr.Error,
		Message: msg,
		Span:    span,
		File:    file,
	})
}

func (r *resolver) resolveImports(file *ast.SourceFile, filePath string) {
	for _, item := range file.Items {
		imp, ok := item.Node.(ast.ImportDecl)
		if !ok {
			continue
		}

		path := imp.Path

		// Skip stdlib modules
		if stdlibModules[path] {
			continue
		}

		// Cycle detection (check before "already resolved" since in-progress modules are also in parsed)
		if r.inProgress[path] {
			r.addError(
				fmt.Sprintf("circular import: %q", path),
				filePath,
				item.Span,
			)
			continue
		}

		// Already fully resolved?
		if _, done := r.parsed[path]; done {
			continue
		}

		// Resolve file path
		resolved := r.resolveFilePath(path)
		if resolved == "" {
			r.addError(
				fmt.Sprintf("cannot resolve import %q: file not found", path),
				filePath,
				item.Span,
			)
			continue
		}

		// Parse the imported file
		r.inProgress[path] = true
		mod := r.parseModule(path, resolved)

		if mod != nil {
			r.parsed[path] = mod
			// Recursively resolve imports in the module (while still in-progress for cycle detection)
			r.resolveImports(mod.File, resolved)
		}
		delete(r.inProgress, path)
	}
}

func (r *resolver) resolveFilePath(importPath string) string {
	// Try <root>/<path>.haira
	candidate := filepath.Join(r.projectRoot, importPath+".haira")
	if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
		return candidate
	}

	// Try <root>/<path>/mod.haira (directory module)
	candidate = filepath.Join(r.projectRoot, importPath, "mod.haira")
	if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
		return candidate
	}

	return ""
}

func (r *resolver) parseModule(importPath, filePath string) *Module {
	source, err := os.ReadFile(filePath)
	if err != nil {
		r.diags = append(r.diags, hairaerr.Diagnostic{
			Level:   hairaerr.Error,
			Message: fmt.Sprintf("cannot read %q: %s", filePath, err),
			File:    filePath,
		})
		return nil
	}

	src := string(source)
	fileAST, parseErrs := parser.Parse(src)
	if len(parseErrs) > 0 {
		for _, e := range parseErrs {
			r.diags = append(r.diags, hairaerr.Diagnostic{
				Level:   hairaerr.Error,
				Message: e.Message,
				Span:    e.Span,
				File:    filePath,
			})
		}
		return nil
	}

	return &Module{
		Path:     importPath,
		FilePath: filePath,
		File:     fileAST,
	}
}

// MergedItems returns all items from the main file and imported modules,
// suitable for passing to codegen. Imported types and functions are included
// so they appear in the generated Go output.
func (p *Program) MergedItems() []ast.Item {
	var items []ast.Item

	// Add imported module items first (types, functions — skip their imports)
	for _, mod := range p.Modules {
		for _, item := range mod.File.Items {
			switch item.Node.(type) {
			case ast.ImportDecl:
				continue // skip nested imports
			default:
				items = append(items, item)
			}
		}
	}

	// Add main file items
	items = append(items, p.Main.Items...)

	return items
}

// ModuleName returns the short name for an import path (last segment).
// E.g., "models/user" → "user"
func ModuleName(importPath string) string {
	parts := strings.Split(importPath, "/")
	return parts[len(parts)-1]
}
