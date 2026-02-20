package codegen

import (
	"fmt"
	"strings"

	"github.com/haira-lang/haira/internal/ast"
	"github.com/haira-lang/haira/internal/checker"
)

// GenerateTestGo generates the contents of main_test.go from test declarations.
func GenerateTestGo(file *ast.SourceFile, sourceFile, sourceText string, typeInfo ...*checker.TypeInfo) string {
	em := NewEmitter()
	em.sourceFile = sourceFile
	em.sourceText = sourceText
	if len(typeInfo) > 0 && typeInfo[0] != nil {
		em.typeInfo = typeInfo[0]
	}

	em.Line("package main")
	em.Blank()

	// Collect test declarations
	var tests []ast.TestDecl
	for _, item := range file.Items {
		if td, ok := item.Node.(ast.TestDecl); ok {
			tests = append(tests, td)
		}
	}

	if len(tests) == 0 {
		return ""
	}

	// Check if agents exist (need TestMain for init)
	var agents []ast.AgentDecl
	for _, item := range file.Items {
		if a, ok := item.Node.(ast.AgentDecl); ok {
			agents = append(agents, a)
		}
	}

	needsOS := len(agents) > 0
	hasMCP := hasMCPProviders(file)

	// Imports
	imports := []string{`"testing"`}
	if needsOS {
		imports = append(imports, `"os"`)
	}
	if needsTestFmt(tests) {
		imports = append(imports, `"fmt"`)
	}
	if needsOS || hasMCP {
		imports = append(imports, `"haira-generated/haira"`)
	}

	if len(imports) == 1 {
		em.Line(fmt.Sprintf("import %s", imports[0]))
	} else {
		em.Line("import (")
		em.Indent()
		for _, imp := range imports {
			em.Line(imp)
		}
		em.Dedent()
		em.Line(")")
	}
	em.Blank()

	// Suppress unused import warnings
	if needsTestFmt(tests) {
		em.Line("var _ = fmt.Sprintf")
		em.Blank()
	}

	// TestMain for agent initialization
	if len(agents) > 0 || hasMCP {
		em.OpenBlock("func TestMain(m *testing.M)")
		if hasMCP {
			em.Line("defer haira.ShutdownMCP()")
		}
		sorted := topoSortAgents(agents)
		for _, name := range sorted {
			em.Line(fmt.Sprintf("initAgent%s()", name))
		}
		em.Line("os.Exit(m.Run())")
		em.CloseBlock()
		em.Blank()
	}

	// TestHaira wrapping all test blocks as sub-tests
	em.OpenBlock("func TestHaira(t *testing.T)")
	for _, td := range tests {
		em.ResetVars()
		em.OpenBlock(fmt.Sprintf("t.Run(%q, func(t *testing.T)", td.Name.Node))
		emitTestBody(em, td.Body)
		em.Dedent()
		em.Line("})")
	}
	em.CloseBlock()

	return em.String()
}

// GenerateMainGoForTest generates main.go with a stub main() for use with go test.
func GenerateMainGoForTest(file *ast.SourceFile, sourceFile, sourceText string, typeInfo ...*checker.TypeInfo) string {
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

	em.Line("package main")
	em.Blank()

	// Collect what we need to import
	needsFmt := needsFmtImport(file)
	needsJSON := needsJSONImport(file)
	needsHaira := needsHairaImport(file)
	needsSync := needsSyncImport(file)
	needsTime := needsTimeImport(file)

	var imports []string
	if needsFmt {
		imports = append(imports, `"fmt"`)
	}
	if needsJSON {
		imports = append(imports, `"encoding/json"`)
	}
	if needsSync {
		imports = append(imports, `"sync"`)
	}
	if needsTime {
		imports = append(imports, `"time"`)
	}
	if needsHaira {
		imports = append(imports, `"haira-generated/haira"`)
	}

	if len(imports) > 0 {
		if len(imports) == 1 {
			em.Line(fmt.Sprintf("import %s", imports[0]))
		} else {
			em.Line("import (")
			em.Indent()
			for _, imp := range imports {
				em.Line(imp)
			}
			em.Dedent()
			em.Line(")")
		}
		em.Blank()
	}

	// Suppress unused import warnings
	if needsFmt {
		em.Line("var _ = fmt.Sprintf")
	}
	if needsSync {
		em.Line("var _ sync.WaitGroup")
	}
	if needsFmt || needsSync {
		em.Blank()
	}

	// Pass 1: providers
	for _, item := range file.Items {
		if p, ok := item.Node.(ast.ProviderDecl); ok {
			em.LineDirective(item.Span)
			EmitProvider(em, p)
		}
	}

	// Pass 2: tools
	for _, item := range file.Items {
		if t, ok := item.Node.(ast.ToolDecl); ok {
			em.LineDirective(item.Span)
			EmitTool(em, t)
		}
	}

	// Pass 3: agents
	for _, item := range file.Items {
		if a, ok := item.Node.(ast.AgentDecl); ok {
			em.LineDirective(item.Span)
			EmitAgent(em, a)
		}
	}

	// Pass 4: workflows
	for _, item := range file.Items {
		if w, ok := item.Node.(ast.WorkflowDecl); ok {
			em.LineDirective(item.Span)
			EmitWorkflow(em, w)
		}
	}

	// Pass 5: top-level statements (variable declarations)
	for _, item := range file.Items {
		if is, ok := item.Node.(ast.ItemStatement); ok {
			if assign, ok := is.Stmt.Node.(ast.AssignStmt); ok {
				em.LineDirective(item.Span)
				emitTopLevelVar(em, assign)
			}
			if letStmt, ok := is.Stmt.Node.(ast.LetStmt); ok {
				em.LineDirective(item.Span)
				emitTopLevelLet(em, letStmt)
			}
		}
	}

	// Pass 6: type defs and enums
	for _, item := range file.Items {
		if td, ok := item.Node.(ast.TypeDef); ok {
			em.LineDirective(item.Span)
			emitTypeDef(em, td)
		}
		if ed, ok := item.Node.(ast.EnumDef); ok {
			em.LineDirective(item.Span)
			emitEnumDef(em, ed)
		}
	}

	// Pass 7: method definitions
	for _, item := range file.Items {
		if md, ok := item.Node.(ast.MethodDef); ok {
			em.LineDirective(item.Span)
			emitMethod(em, md)
		}
	}

	// Pass 8: non-main functions (skip main — stub replaces it)
	for _, item := range file.Items {
		if f, ok := item.Node.(ast.FunctionDef); ok {
			if f.Name.Node != "main" {
				em.LineDirective(item.Span)
				emitFunction(em, f)
			}
		}
	}

	// Stub main() — go test uses its own entry point
	em.OpenBlock("func main()")
	em.CloseBlock()

	return em.String()
}

// emitTestBody emits the body of a test block, handling assert statements specially.
func emitTestBody(em *GoEmitter, block ast.Block) {
	for _, stmt := range block.Statements {
		em.LineDirective(stmt.Span)
		switch s := stmt.Node.(type) {
		case ast.AssertStmt:
			emitAssert(em, s)
		default:
			EmitStatement(em, stmt)
		}
	}
}

// emitAssert generates Go test assertions from a Haira assert statement.
func emitAssert(em *GoEmitter, s ast.AssertStmt) {
	// Smart equality: assert x == y → got/want diff
	if bin, ok := s.Condition.Node.(ast.BinaryExpr); ok && bin.Op.Node == ast.OpEq {
		lhs := ExprToGo(bin.Left)
		rhs := ExprToGo(bin.Right)
		em.OpenBlock(fmt.Sprintf("if (%s) != (%s)", lhs, rhs))
		if s.Message != nil {
			em.Line(fmt.Sprintf("t.Errorf(\"assertion failed (%%v): got %%v, want %%v\", %s, %s, %s)",
				ExprToGo(*s.Message), lhs, rhs))
		} else {
			em.Line(fmt.Sprintf("t.Errorf(\"assertion failed: got %%v, want %%v\", %s, %s)", lhs, rhs))
		}
		em.CloseBlock()
		return
	}

	// Smart not-equal: assert x != y
	if bin, ok := s.Condition.Node.(ast.BinaryExpr); ok && bin.Op.Node == ast.OpNe {
		lhs := ExprToGo(bin.Left)
		rhs := ExprToGo(bin.Right)
		em.OpenBlock(fmt.Sprintf("if (%s) == (%s)", lhs, rhs))
		if s.Message != nil {
			em.Line(fmt.Sprintf("t.Errorf(\"assertion failed (%%v): expected not equal, got %%v\", %s, %s)",
				ExprToGo(*s.Message), lhs))
		} else {
			em.Line(fmt.Sprintf("t.Errorf(\"assertion failed: expected not equal, got %%v\", %s)", lhs))
		}
		em.CloseBlock()
		return
	}

	// General case: assert <cond>
	cond := ExprToGo(s.Condition)
	em.OpenBlock(fmt.Sprintf("if !(%s)", cond))
	if s.Message != nil {
		em.Line(fmt.Sprintf("t.Fatalf(\"assertion failed: %%v\", %s)", ExprToGo(*s.Message)))
	} else {
		em.Line(fmt.Sprintf("t.Fatalf(\"assertion failed: %s\")", escapePercent(cond)))
	}
	em.CloseBlock()
}

// escapePercent escapes % signs in format strings.
func escapePercent(s string) string {
	return strings.ReplaceAll(s, "%", "%%")
}

// needsTestFmt checks if any test body contains interpolated strings.
func needsTestFmt(tests []ast.TestDecl) bool {
	for _, td := range tests {
		if blockHasInterpolatedString(td.Body) {
			return true
		}
	}
	return false
}

// HasTests returns true if the source file contains any test declarations.
func HasTests(file *ast.SourceFile) bool {
	for _, item := range file.Items {
		if _, ok := item.Node.(ast.TestDecl); ok {
			return true
		}
	}
	return false
}

// hasMainFunction returns true if the source file contains a fn main() definition.
func hasMainFunction(file *ast.SourceFile) bool {
	for _, item := range file.Items {
		if f, ok := item.Node.(ast.FunctionDef); ok && f.Name.Node == "main" {
			return true
		}
	}
	return false
}
