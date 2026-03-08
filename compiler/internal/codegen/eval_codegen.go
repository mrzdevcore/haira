package codegen

import (
	"fmt"
	"strings"

	"github.com/haira-lang/haira/internal/ast"
	"github.com/haira-lang/haira/internal/checker"
)

// GenerateEvalGo generates a standalone Go program that runs eval blocks.
func GenerateEvalGo(file *ast.SourceFile, sourceFile, sourceText string, typeInfo ...*checker.TypeInfo) string {
	em := NewEmitter()
	em.sourceFile = sourceFile
	em.sourceText = sourceText
	if len(typeInfo) > 0 && typeInfo[0] != nil {
		em.typeInfo = typeInfo[0]
		activeTypeInfo = typeInfo[0]
	} else {
		activeTypeInfo = nil
	}

	activeSourceFile = file

	// Collect tool names
	knownToolNames = make(map[string]bool)
	for _, item := range file.Items {
		if t, ok := item.Node.(ast.ToolDecl); ok {
			knownToolNames[t.Name.Node] = true
		}
	}

	em.Line("package main")
	em.Blank()

	// Imports
	em.Line("import (")
	em.Line(`	"fmt"`)
	em.Line(`	"os"`)
	em.Line(`	haira "haira-generated/haira"`)

	// Check if any stdlib imports are needed
	for _, item := range file.Items {
		if imp, ok := item.Node.(ast.ImportDecl); ok {
			if goImport, ok := stdlibGoImport(imp.Path); ok {
				em.Line(fmt.Sprintf(`	%q`, "haira-generated/"+goImport))
			}
		}
	}

	em.Line(")")
	em.Blank()

	// Suppress unused import warnings
	em.Line("var _ = fmt.Sprintf")
	em.Line("var _ = haira.NewAgent")
	em.Blank()

	// Emit providers
	for _, item := range file.Items {
		if p, ok := item.Node.(ast.ProviderDecl); ok {
			EmitProvider(em, p)
		}
	}

	// Emit tools
	for _, item := range file.Items {
		if t, ok := item.Node.(ast.ToolDecl); ok {
			EmitTool(em, t)
		}
	}

	// Emit agents
	for _, item := range file.Items {
		if a, ok := item.Node.(ast.AgentDecl); ok {
			EmitAgent(em, a)
		}
	}

	// Emit functions (non-main)
	for _, item := range file.Items {
		if f, ok := item.Node.(ast.FunctionDef); ok {
			if f.Name.Node != "main" {
				emitFunction(em, f)
			}
		}
	}

	// Emit eval runner main
	em.OpenBlock("func main()")

	// Collect eval blocks
	var evals []ast.EvalDecl
	for _, item := range file.Items {
		if e, ok := item.Node.(ast.EvalDecl); ok {
			evals = append(evals, e)
		}
	}

	if len(evals) == 0 {
		em.Line(`fmt.Println("No eval blocks found.")`)
		em.CloseBlock()
		return em.String()
	}

	em.Line("allPassed := true")
	em.Blank()

	for _, ev := range evals {
		emitEvalBlock(em, ev)
	}

	em.Blank()
	em.OpenBlock("if !allPassed")
	em.Line("os.Exit(1)")
	em.CloseBlock()
	em.CloseBlock()

	return em.String()
}

// emitEvalBlock generates code for a single eval declaration.
func emitEvalBlock(em *GoEmitter, ev ast.EvalDecl) {
	evalName := ev.Name.Node

	// Extract fields
	var agentExpr, casesExpr, thresholdExpr string
	parallel := false

	for _, field := range ev.Fields {
		switch field.Key.Node {
		case "agent":
			agentExpr = ExprToGo(field.Value)
		case "cases":
			casesExpr = emitEvalCases(field.Value)
		case "threshold":
			thresholdExpr = ExprToGo(field.Value)
		case "parallel":
			if lit, ok := field.Value.Node.(ast.LiteralExpr); ok {
				if b, ok := lit.Lit.(ast.BoolLit); ok {
					parallel = b.Value
				}
			}
		}
	}

	if agentExpr == "" {
		em.Line(fmt.Sprintf(`fmt.Println("eval %q: missing agent field")`, evalName))
		return
	}
	if thresholdExpr == "" {
		thresholdExpr = "0.8" // default 80%
	}

	em.Line(fmt.Sprintf(`fmt.Printf("\nRunning eval: %s\n")`, evalName))
	em.OpenBlock("")

	if casesExpr != "" {
		em.Line(fmt.Sprintf("cases := %s", casesExpr))
	} else {
		em.Line("cases := []haira.EvalCase{}")
	}

	em.Line(fmt.Sprintf("summary := haira.RunEval(%q, %s, cases, %s, %v)",
		evalName, agentExpr, thresholdExpr, parallel))
	em.Line("haira.PrintEvalSummary(summary)")
	em.OpenBlock(fmt.Sprintf("if summary.Score < %s", thresholdExpr))
	em.Line("allPassed = false")
	em.CloseBlock()

	em.CloseBlock()
}

// emitEvalCases converts a list expression of eval case maps to Go code.
func emitEvalCases(expr ast.Expr) string {
	list, ok := expr.Node.(ast.ListExpr)
	if !ok {
		return "[]haira.EvalCase{}"
	}

	var cases []string
	for _, elem := range list.Elems {
		// Each element should be a map or instance like { input: "...", expected: "..." }
		if mapExpr, ok := elem.Node.(ast.MapExpr); ok {
			c := emitEvalCaseFromMap(mapExpr)
			cases = append(cases, c)
		} else if inst, ok := elem.Node.(ast.InstanceExpr); ok {
			c := emitEvalCaseFromInstance(inst)
			cases = append(cases, c)
		}
	}

	return fmt.Sprintf("[]haira.EvalCase{\n\t\t%s,\n\t}", strings.Join(cases, ",\n\t\t"))
}

func emitEvalCaseFromMap(m ast.MapExpr) string {
	var input, expected, rubric string
	for _, entry := range m.Entries {
		if key, ok := entry.Key.Node.(ast.LiteralExpr); ok {
			if s, ok := key.Lit.(ast.StringLit); ok {
				switch s.Value {
				case "input":
					input = ExprToGo(entry.Value)
				case "expected":
					expected = ExprToGo(entry.Value)
				case "rubric":
					rubric = ExprToGo(entry.Value)
				}
			}
		}
	}
	return fmt.Sprintf(`haira.EvalCase{Input: %s, Expected: %s, Rubric: %s}`, input, expected, rubric)
}

func emitEvalCaseFromInstance(inst ast.InstanceExpr) string {
	var input, expected, rubric string
	for _, field := range inst.Fields {
		if field.Name == nil {
			continue
		}
		switch field.Name.Node {
		case "input":
			input = ExprToGo(field.Value)
		case "expected":
			expected = ExprToGo(field.Value)
		case "rubric":
			rubric = ExprToGo(field.Value)
		}
	}
	if input == "" {
		input = `""`
	}
	if expected == "" {
		expected = `""`
	}
	if rubric == "" {
		rubric = `""`
	}
	return fmt.Sprintf(`haira.EvalCase{Input: %s, Expected: %s, Rubric: %s}`, input, expected, rubric)
}

// HasEvals returns true if the file contains any eval declarations.
func HasEvals(file *ast.SourceFile) bool {
	for _, item := range file.Items {
		if _, ok := item.Node.(ast.EvalDecl); ok {
			return true
		}
	}
	return false
}
