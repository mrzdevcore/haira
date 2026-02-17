// Package driver orchestrates the compilation pipeline.
package driver

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/haira-lang/haira/internal/ast"
	"github.com/haira-lang/haira/internal/checker"
	"github.com/haira-lang/haira/internal/codegen"
	hairaerr "github.com/haira-lang/haira/internal/errors"
	"github.com/haira-lang/haira/internal/formatter"
	"github.com/haira-lang/haira/internal/lexer"
	"github.com/haira-lang/haira/internal/parser"
	"github.com/haira-lang/haira/internal/resolver"
)

// Compile reads a Haira source file and compiles it to a binary.
func Compile(file, output string) error {
	fmt.Fprintf(os.Stderr, "Compiling: %s\n", file)

	sf, src, err := resolveAndParse(file)
	if err != nil {
		return err
	}

	// Type check
	typeInfo, typeDiags := checker.Check(sf)
	if hairaerr.HasErrors(typeDiags) {
		return reportErrors(typeDiags, src)
	}
	reportWarnings(typeDiags, src)

	runtimePath := codegen.FindRuntimePath()
	if runtimePath == "" {
		return fmt.Errorf("could not find go-runtime directory.\nExpected at: <project_root>/go-runtime/\nMake sure you're running from the Haira project directory")
	}

	if output == "" {
		stem := filepath.Base(file)
		ext := filepath.Ext(stem)
		if ext != "" {
			stem = stem[:len(stem)-len(ext)]
		}
		outputDir := ".output"
		os.MkdirAll(outputDir, 0o755)
		output = filepath.Join(outputDir, stem)
	}

	if err := codegen.CompileToBinary(sf, output, runtimePath, file, src, typeInfo); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Built: %s\n", output)
	return nil
}

// Run reads a Haira source file, compiles, and executes it.
func Run(file string) error {
	sf, src, err := resolveAndParse(file)
	if err != nil {
		return err
	}

	// Type check
	typeInfo, typeDiags := checker.Check(sf)
	if hairaerr.HasErrors(typeDiags) {
		return reportErrors(typeDiags, src)
	}
	reportWarnings(typeDiags, src)

	runtimePath := codegen.FindRuntimePath()
	if runtimePath == "" {
		return fmt.Errorf("could not find go-runtime directory")
	}

	return codegen.RunProgram(sf, runtimePath, file, src, typeInfo)
}

// ParseFile parses a file and prints the AST.
func ParseFile(file string) error {
	source, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	fmt.Printf("Parsing: %s\n\n", file)

	src := string(source)
	ast, errs := parser.Parse(src)

	if len(errs) > 0 {
		fmt.Println("Errors:")
		diags := toDiagnostics(errs, file)
		fmt.Print(hairaerr.FormatAll(diags, src))
		fmt.Println()
	}

	fmt.Println("AST:")
	printAST(ast)

	fmt.Printf("\n%d items, %d errors\n", len(ast.Items), len(errs))

	if len(errs) > 0 {
		return fmt.Errorf("%d parse errors", len(errs))
	}
	return nil
}

// CheckFile parses a file and reports errors without generating code.
func CheckFile(file string) error {
	sf, src, err := resolveAndParse(file)
	if err != nil {
		return err
	}

	// Type check
	_, typeDiags := checker.Check(sf)
	if hairaerr.HasErrors(typeDiags) {
		return reportErrors(typeDiags, src)
	}
	reportWarnings(typeDiags, src)

	fmt.Printf("%s: OK\n", file)
	return nil
}

// EmitFile parses a file and prints the generated Go code.
func EmitFile(file string) error {
	sf, src, err := resolveAndParse(file)
	if err != nil {
		return err
	}

	// Type check
	typeInfo, typeDiags := checker.Check(sf)
	if hairaerr.HasErrors(typeDiags) {
		return reportErrors(typeDiags, src)
	}
	reportWarnings(typeDiags, src)

	fmt.Print(codegen.ShowGeneratedGo(sf, file, src, typeInfo))
	return nil
}

// LexFile tokenizes a file and prints the tokens.
func LexFile(file string) error {
	source, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	l := lexer.New(string(source))
	for _, tok := range l.AllTokens() {
		fmt.Println(tok)
	}
	return nil
}

// Test reads a Haira source file and runs its test blocks.
func Test(file string, testArgs []string) error {
	sf, src, err := resolveAndParse(file)
	if err != nil {
		return err
	}

	// Type check
	typeInfo, typeDiags := checker.Check(sf)
	if hairaerr.HasErrors(typeDiags) {
		return reportErrors(typeDiags, src)
	}
	reportWarnings(typeDiags, src)

	runtimePath := codegen.FindRuntimePath()
	if runtimePath == "" {
		return fmt.Errorf("could not find go-runtime directory")
	}

	return codegen.RunTests(sf, runtimePath, file, src, testArgs, typeInfo)
}

// FormatFile formats a Haira source file in-place.
func FormatFile(file string) error {
	source, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	src := string(source)

	l := lexer.New(src)
	tokens := l.AllTokens()

	sf, errs := parser.Parse(src)
	if len(errs) > 0 {
		diags := toDiagnostics(errs, file)
		fmt.Fprint(os.Stderr, hairaerr.FormatAll(diags, src))
		return fmt.Errorf("cannot format %s: %d parse error(s)", file, len(errs))
	}

	formatted := formatter.Format(src, tokens, sf)

	if formatted == src {
		return nil // already formatted
	}

	return os.WriteFile(file, []byte(formatted), 0o644)
}

// resolveAndParse resolves imports and parses all files into a merged SourceFile.
// Returns the merged AST, the main file source (for error reporting), and any error.
func resolveAndParse(file string) (*ast.SourceFile, string, error) {
	source, err := os.ReadFile(file)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read file: %w", err)
	}
	src := string(source)

	// Use resolver for multi-file support
	prog, diags := resolver.Resolve(file)
	if hairaerr.HasErrors(diags) {
		return nil, src, reportErrors(diags, src)
	}
	reportWarnings(diags, src)

	// If there are imported modules, merge their items into the main AST
	if len(prog.Modules) > 0 {
		merged := prog.MergedItems()
		prog.Main.Items = merged
	}

	return prog.Main, src, nil
}

// toDiagnostics converts parser errors to Diagnostic values.
func toDiagnostics(errs []parser.ParseError, file string) []hairaerr.Diagnostic {
	diags := make([]hairaerr.Diagnostic, len(errs))
	for i, e := range errs {
		diags[i] = hairaerr.Diagnostic{
			Level:   hairaerr.Error,
			Message: e.Message,
			Span:    e.Span,
			File:    file,
		}
	}
	return diags
}

// reportWarnings prints non-error diagnostics to stderr.
func reportWarnings(diags []hairaerr.Diagnostic, source string) {
	for _, d := range diags {
		if d.Level != hairaerr.Error {
			fmt.Fprint(os.Stderr, hairaerr.PrettyPrint(d, source))
		}
	}
}

// reportErrors pretty-prints diagnostics to stderr and returns an error summary.
func reportErrors(diags []hairaerr.Diagnostic, source string) error {
	fmt.Fprint(os.Stderr, hairaerr.FormatAll(diags, source))
	count := 0
	for _, d := range diags {
		if d.Level == hairaerr.Error {
			count++
		}
	}
	return fmt.Errorf("%d error(s)", count)
}

func printAST(file *ast.SourceFile) {
	for _, item := range file.Items {
		printItem(item)
	}
}

func printItem(item ast.Item) {
	switch it := item.Node.(type) {
	case ast.TypeDef:
		fmt.Printf("  TypeDef: %s (%d fields)\n", it.Name.Node, len(it.Fields))
	case ast.FunctionDef:
		fmt.Printf("  FunctionDef: %s (%d params)\n", it.Name.Node, len(it.Params))
	case ast.MethodDef:
		fmt.Printf("  MethodDef: %s.%s (%d params)\n", it.TypeName.Node, it.Name.Node, len(it.Params))
	case ast.EnumDef:
		fmt.Printf("  EnumDef: %s (%d variants)\n", it.Name.Node, len(it.Variants))
	case ast.TypeAlias:
		fmt.Printf("  TypeAlias: %s\n", it.Name.Node)
	case ast.ImportDecl:
		fmt.Printf("  ImportDecl: %q\n", it.Path)
	case ast.ExportDecl:
		names := make([]string, len(it.Names))
		for i, n := range it.Names {
			names[i] = n.Node
		}
		fmt.Printf("  ExportDecl: {%s}\n", strings.Join(names, ", "))
	case ast.ProviderDecl:
		fmt.Printf("  ProviderDecl: %s (%d fields)\n", it.Name.Node, len(it.Fields))
	case ast.ToolDecl:
		fmt.Printf("  ToolDecl: %s (%d params)\n", it.Name.Node, len(it.Params))
	case ast.AgentDecl:
		fmt.Printf("  AgentDecl: %s (%d fields)\n", it.Name.Node, len(it.Fields))
	case ast.WorkflowDecl:
		trigger := ""
		if it.Trigger != nil {
			trigger = "@" + it.Trigger.Name.Node + " "
		}
		fmt.Printf("  WorkflowDecl: %s%s (%d params)\n", trigger, it.Name.Node, len(it.Params))
	case ast.TestDecl:
		fmt.Printf("  TestDecl: %q\n", it.Name.Node)
	case ast.ItemStatement:
		fmt.Printf("  Statement: %T\n", it.Stmt.Node)
	default:
		fmt.Printf("  %T\n", item.Node)
	}
}
