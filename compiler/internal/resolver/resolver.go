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
	"io": true, "http": true, "mcp": true, "env": true, "json": true,
	"postgres": true, "slack": true, "excel": true, "time": true,
	"string": true, "regex": true, "math": true, "conv": true,
	"array": true, "map": true, "log": true,
	"ui": true, "vector": true, "observe": true,
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
//
// Visibility rules:
//   - Only `pub` items are exported from modules.
//   - Selective imports (import { X, Y } from "m") only include named items.
//   - Glob imports (import * from "m") include all pub items.
//   - Basic imports (import "m") include all pub items (accessed as m.X).
func (p *Program) MergedItems() []ast.Item {
	// Build a map of import path → ImportDecl from main file
	importMap := make(map[string]ast.ImportDecl)
	for _, item := range p.Main.Items {
		if imp, ok := item.Node.(ast.ImportDecl); ok {
			importMap[imp.Path] = imp
		}
	}

	var items []ast.Item

	// Add imported module items first (types, functions — skip their imports)
	for _, mod := range p.Modules {
		imp := importMap[mod.Path]
		// Build a set of selectively imported names (if any)
		var selectiveNames map[string]bool
		if len(imp.Names) > 0 {
			selectiveNames = make(map[string]bool, len(imp.Names))
			for _, n := range imp.Names {
				selectiveNames[n.Node] = true
			}
		}

		for _, item := range mod.File.Items {
			switch item.Node.(type) {
			case ast.ImportDecl, ast.ExportDecl:
				continue // skip nested imports and exports
			}

			// Check pub visibility
			if !isItemPublic(item) {
				continue
			}

			// If selective import, only include named items
			if selectiveNames != nil {
				name := itemName(item)
				if name == "" || !selectiveNames[name] {
					continue
				}
			}

			items = append(items, item)
		}
	}

	// Add main file items
	items = append(items, p.Main.Items...)

	return items
}

// isItemPublic returns whether a top-level item is marked as pub (exported).
// Agentic declarations (provider, tool, agent, workflow) are always public.
func isItemPublic(item ast.Item) bool {
	switch it := item.Node.(type) {
	case ast.FunctionDef:
		return it.IsPublic
	case ast.TypeDef:
		return it.IsPublic
	case ast.EnumDef:
		return it.IsPublic
	case ast.MethodDef:
		return true // methods follow their type's visibility
	case ast.ProviderDecl, ast.ToolDecl, ast.AgentDecl, ast.WorkflowDecl:
		return true // agentic declarations are always public
	case ast.TypeAlias:
		return true // type aliases are always public for now
	case ast.ItemStatement:
		return true // top-level statements (vars) are public for now
	default:
		return false
	}
}

// itemName returns the name of a top-level item, or "" if unnamed.
func itemName(item ast.Item) string {
	switch it := item.Node.(type) {
	case ast.FunctionDef:
		return it.Name.Node
	case ast.TypeDef:
		return it.Name.Node
	case ast.EnumDef:
		return it.Name.Node
	case ast.MethodDef:
		return it.Name.Node
	case ast.TypeAlias:
		return it.Name.Node
	case ast.ProviderDecl:
		return it.Name.Node
	case ast.ToolDecl:
		return it.Name.Node
	case ast.AgentDecl:
		return it.Name.Node
	case ast.WorkflowDecl:
		return it.Name.Node
	default:
		return ""
	}
}

// ModuleName returns the short name for an import path (last segment).
// E.g., "models/user" → "user"
func ModuleName(importPath string) string {
	parts := strings.Split(importPath, "/")
	return parts[len(parts)-1]
}
