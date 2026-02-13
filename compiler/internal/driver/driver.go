// Package driver orchestrates the compilation pipeline.
package driver

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/haira-lang/haira/internal/ast"
	"github.com/haira-lang/haira/internal/codegen"
	"github.com/haira-lang/haira/internal/lexer"
	"github.com/haira-lang/haira/internal/parser"
)

// Compile reads a Haira source file and compiles it to a binary.
func Compile(file, output string) error {
	source, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Compiling: %s\n", file)

	ast, errs := parser.Parse(string(source))
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "Parse error: %s\n", e.Message)
		}
		return fmt.Errorf("%d parse error(s)", len(errs))
	}

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

	if err := codegen.CompileToBinary(ast, output, runtimePath); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Built: %s\n", output)
	return nil
}

// Run reads a Haira source file, compiles, and executes it.
func Run(file string) error {
	source, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	ast, errs := parser.Parse(string(source))
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "Parse error: %s\n", e.Message)
		}
		return fmt.Errorf("%d parse error(s)", len(errs))
	}

	runtimePath := codegen.FindRuntimePath()
	if runtimePath == "" {
		return fmt.Errorf("could not find go-runtime directory")
	}

	return codegen.RunProgram(ast, runtimePath)
}

// ParseFile parses a file and prints the AST.
func ParseFile(file string) error {
	source, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	fmt.Printf("Parsing: %s\n\n", file)

	ast, errs := parser.Parse(string(source))

	if len(errs) > 0 {
		fmt.Println("Errors:")
		for _, e := range errs {
			line, col := offsetToLineCol(string(source), e.Span.Start)
			fmt.Printf("  %s:%d:%d: %s\n", file, line, col, e.Message)
		}
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
	source, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	_, errs := parser.Parse(string(source))

	if len(errs) > 0 {
		for _, e := range errs {
			line, col := offsetToLineCol(string(source), e.Span.Start)
			fmt.Printf("%s:%d:%d: %s\n", file, line, col, e.Message)
		}
		return fmt.Errorf("%d error(s)", len(errs))
	}

	fmt.Printf("%s: OK\n", file)
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

func offsetToLineCol(source string, offset int) (int, int) {
	line, col := 1, 1
	for i, ch := range source {
		if i >= offset {
			break
		}
		if ch == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	return line, col
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
	case ast.ItemStatement:
		fmt.Printf("  Statement: %T\n", it.Stmt.Node)
	default:
		fmt.Printf("  %T\n", item.Node)
	}
}
