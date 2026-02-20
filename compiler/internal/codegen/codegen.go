package codegen

import (
	"fmt"
	"strings"

	"github.com/haira-lang/haira/internal/ast"
	"github.com/haira-lang/haira/internal/checker"
)

// RuntimeImportPath is the Go import path for the Haira runtime package.
// In full mode it points to the external go-runtime; in minimal mode it
// points to the local package within the generated module.
var RuntimeImportPath = "haira-go-runtime/haira"

// SetMinimalRuntime switches codegen to use the embedded minimal runtime.
func SetMinimalRuntime() {
	RuntimeImportPath = "haira-generated/haira"
}

// ResetRuntime switches codegen back to the full external runtime.
func ResetRuntime() {
	RuntimeImportPath = "haira-go-runtime/haira"
}

// fullRuntimeModules are import paths that require the full external go-runtime.
var fullRuntimeModules = map[string]bool{
	"postgres": true, "excel": true, "slack": true,
	"vector": true, "observe": true, "ui": true, "mcp": true,
}

// NeedsFullRuntime returns true if the program requires the full external go-runtime.
// Programs that only use minimal stdlib modules (io, http, string, math, etc.)
// can use the embedded runtime instead.
func NeedsFullRuntime(file *ast.SourceFile) bool {
	for _, item := range file.Items {
		switch it := item.Node.(type) {
		case ast.ProviderDecl, ast.ToolDecl, ast.AgentDecl, ast.WorkflowDecl:
			return true
		case ast.ImportDecl:
			if fullRuntimeModules[it.Path] {
				return true
			}
		}
	}
	return false
}

// GenerateMainGo generates the contents of main.go from the AST.
func GenerateMainGo(file *ast.SourceFile, sourceFile, sourceText string, typeInfo ...*checker.TypeInfo) string {
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
		imports = append(imports, fmt.Sprintf("%q", RuntimeImportPath))
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

	var mainFn *ast.FunctionDef
	var agents []ast.AgentDecl

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
			agents = append(agents, a)
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

	// Pass 8: non-main functions
	for _, item := range file.Items {
		if f, ok := item.Node.(ast.FunctionDef); ok {
			if f.Name.Node == "main" {
				mainFn = &f
			} else {
				em.LineDirective(item.Span)
				emitFunction(em, f)
			}
		}
	}

	// Finally: main function
	if mainFn != nil {
		emitMainFunction(em, *mainFn, agents, hasMCPProviders(file))
	}

	return em.String()
}

func emitTopLevelVar(em *GoEmitter, assign ast.AssignStmt) {
	value := ExprToGo(assign.Value)
	if len(assign.Targets) == 1 {
		name := assignTargetName(assign.Targets[0].Path)
		em.Line(fmt.Sprintf("var %s = %s", name, value))
	} else {
		names := make([]string, len(assign.Targets))
		for i, t := range assign.Targets {
			names[i] = assignTargetName(t.Path)
		}
		em.Line(fmt.Sprintf("var %s = %s", strings.Join(names, ", "), value))
	}
}

func emitTopLevelLet(em *GoEmitter, letStmt ast.LetStmt) {
	value := ExprToGo(letStmt.Value)
	em.Line(fmt.Sprintf("var %s = %s", letStmt.Name.Node, value))
}

func assignTargetName(path ast.AssignPath) string {
	switch p := path.(type) {
	case ast.IdentPath:
		return p.Name.Node
	case ast.FieldPath:
		return assignTargetName(p.Object) + "." + p.Field.Node
	case ast.IndexPath:
		return fmt.Sprintf("%s[%s]", assignTargetName(p.Object), ExprToGo(p.Index))
	}
	return "?"
}

func emitTypeDef(em *GoEmitter, td ast.TypeDef) {
	em.OpenBlock(fmt.Sprintf("type %s struct", td.Name.Node))
	for _, field := range td.Fields {
		goType := "any"
		if field.Ty != nil {
			goType = HairaTypeToGo(field.Ty.Node)
		}
		fieldName := Capitalize(field.Name.Node)
		em.Line(fmt.Sprintf("%s %s", fieldName, goType))
	}
	em.CloseBlock()
	em.Blank()
}

func emitEnumDef(em *GoEmitter, ed ast.EnumDef) {
	name := ed.Name.Node
	hasData := false
	for _, v := range ed.Variants {
		if len(v.Fields) > 0 {
			hasData = true
			break
		}
	}

	if hasData {
		// Data-carrying enum → Go interface + variant structs
		em.OpenBlock(fmt.Sprintf("type %s interface", name))
		em.Line(fmt.Sprintf("is%s()", name))
		em.CloseBlock()
		em.Blank()

		for _, variant := range ed.Variants {
			vname := variant.Name.Node
			if len(variant.Fields) == 0 {
				em.OpenBlock(fmt.Sprintf("type %s%s struct", name, vname))
				em.CloseBlock()
			} else {
				em.OpenBlock(fmt.Sprintf("type %s%s struct", name, vname))
				for _, field := range variant.Fields {
					goType := "any"
					if field.Ty != nil {
						goType = HairaTypeToGo(field.Ty.Node)
					}
					em.Line(fmt.Sprintf("%s %s", Capitalize(field.Name.Node), goType))
				}
				em.CloseBlock()
			}
			em.Line(fmt.Sprintf("func (%s%s) is%s() {}", name, vname, name))
			em.Blank()
		}
	} else {
		// Simple enum → const + iota
		em.Line(fmt.Sprintf("type %s int", name))
		em.Blank()
		em.Line("const (")
		em.Indent()
		for i, variant := range ed.Variants {
			if i == 0 {
				em.Line(fmt.Sprintf("%s%s %s = iota", name, variant.Name.Node, name))
			} else {
				em.Line(fmt.Sprintf("%s%s", name, variant.Name.Node))
			}
		}
		em.Dedent()
		em.Line(")")
		em.Blank()
	}
}

func emitMethod(em *GoEmitter, method ast.MethodDef) {
	em.ResetVars()
	typeName := method.TypeName.Node
	methodName := SnakeToPascal(method.Name.Node)
	params := make([]string, len(method.Params))
	for i, p := range method.Params {
		ty := "any"
		if p.Ty != nil {
			ty = HairaTypeToGo(p.Ty.Node)
		}
		params[i] = p.Name.Node + " " + ty
	}
	ret := ""
	if method.ReturnTy != nil {
		ret = " " + HairaTypeToGo(method.ReturnTy.Node)
	}
	em.OpenBlock(fmt.Sprintf("func (self *%s) %s(%s)%s", typeName, methodName, strings.Join(params, ", "), ret))
	EmitBlockBody(em, method.Body)
	em.CloseBlock()
	em.Blank()
}

func emitFunction(em *GoEmitter, fn ast.FunctionDef) {
	em.ResetVars()
	name := SnakeToPascal(fn.Name.Node)
	params := make([]string, len(fn.Params))
	for i, p := range fn.Params {
		ty := "any"
		if p.Ty != nil {
			ty = HairaTypeToGo(p.Ty.Node)
		}
		params[i] = p.Name.Node + " " + ty
	}
	ret := ""
	if fn.ReturnTy != nil {
		ret = " " + HairaTypeToGo(fn.ReturnTy.Node)
	}
	em.OpenBlock(fmt.Sprintf("func %s(%s)%s", name, strings.Join(params, ", "), ret))
	EmitBlockBody(em, fn.Body)
	em.CloseBlock()
	em.Blank()
}

func emitMainFunction(em *GoEmitter, fn ast.FunctionDef, agents []ast.AgentDecl, hasMCP bool) {
	em.ResetVars()
	em.OpenBlock("func main()")
	if hasMCP {
		em.Line("defer haira.ShutdownMCP()")
	}
	sorted := topoSortAgents(agents)
	for _, name := range sorted {
		em.Line(fmt.Sprintf("initAgent%s()", name))
	}
	if len(sorted) > 0 || hasMCP {
		em.Blank()
	}
	EmitBlockBody(em, fn.Body)
	em.CloseBlock()
}

// topoSortAgents returns agent names ordered so that handoff targets come first.
func topoSortAgents(agents []ast.AgentDecl) []string {
	// Build dependency map: agent -> agents it hands off to
	deps := map[string][]string{}
	nameSet := map[string]bool{}
	for _, a := range agents {
		name := a.Name.Node
		nameSet[name] = true
		for _, field := range a.Fields {
			if field.Key.Node == "handoffs" {
				if list, ok := field.Value.Node.(ast.ListExpr); ok {
					for _, item := range list.Elems {
						if ident, ok := item.Node.(ast.IdentExpr); ok {
							deps[name] = append(deps[name], ident.Name)
						}
					}
				}
			}
		}
	}

	// Topological sort via DFS
	visited := map[string]bool{}
	var result []string
	var visit func(string)
	visit = func(name string) {
		if visited[name] {
			return
		}
		visited[name] = true
		for _, dep := range deps[name] {
			if nameSet[dep] {
				visit(dep)
			}
		}
		result = append(result, name)
	}
	// Visit in declaration order for stable output
	for _, a := range agents {
		visit(a.Name.Node)
	}
	return result
}

// hasMCPProviders checks if any provider in the source file uses transport: "mcp".
func hasMCPProviders(file *ast.SourceFile) bool {
	for _, item := range file.Items {
		if p, ok := item.Node.(ast.ProviderDecl); ok {
			if isMCPProvider(p) {
				return true
			}
		}
	}
	return false
}

// Import detection helpers

func needsFmtImport(file *ast.SourceFile) bool {
	for _, item := range file.Items {
		switch it := item.Node.(type) {
		case ast.ToolDecl:
			if it.Body == nil {
				return true
			}
			if it.Body != nil && (blockHasInterpolatedString(*it.Body) || blockHasTry(*it.Body)) {
				return true
			}
		case ast.FunctionDef:
			if blockHasInterpolatedString(it.Body) || blockHasTry(it.Body) {
				return true
			}
		case ast.MethodDef:
			if blockHasInterpolatedString(it.Body) || blockHasTry(it.Body) {
				return true
			}
		case ast.WorkflowDecl:
			if blockHasInterpolatedString(it.Body) || blockHasTry(it.Body) {
				return true
			}
			// Lifecycle hooks, @retry, and step-early-return use fmt.Errorf/fmt.Sprintf
			if len(it.Hooks) > 0 || blockHasRetryDecorator(it.Body) || blockHasStepHooks(it.Body) || blockHasStepWithReturn(it.Body) {
				return true
			}
		}
	}
	return false
}

func blockHasStepHooks(block ast.Block) bool {
	for _, stmt := range block.Statements {
		if s, ok := stmt.Node.(ast.StepStmt); ok {
			if len(s.Hooks) > 0 {
				return true
			}
		}
	}
	return false
}

func blockHasStepWithReturn(block ast.Block) bool {
	for _, stmt := range block.Statements {
		if s, ok := stmt.Node.(ast.StepStmt); ok {
			if statementsHaveReturn(s.Body) {
				return true
			}
		}
	}
	return false
}

func statementsHaveReturn(stmts []ast.Statement) bool {
	for _, stmt := range stmts {
		switch s := stmt.Node.(type) {
		case ast.ReturnStmt:
			return true
		case ast.IfStmt:
			if statementsHaveReturn(s.ThenBranch.Statements) {
				return true
			}
			if s.ElseBranch != nil {
				switch eb := s.ElseBranch.(type) {
				case *ast.ElseBlock:
					if statementsHaveReturn(eb.Body.Statements) {
						return true
					}
				case *ast.ElseIf:
					if statementsHaveReturn([]ast.Statement{{Node: eb.If.Node, Span: eb.If.Span}}) {
						return true
					}
				}
			}
		case ast.ForStmt:
			if statementsHaveReturn(s.Body.Statements) {
				return true
			}
		}
	}
	return false
}

func needsJSONImport(file *ast.SourceFile) bool {
	for _, item := range file.Items {
		if _, ok := item.Node.(ast.ToolDecl); ok {
			return true
		}
	}
	return false
}

func needsHairaImport(file *ast.SourceFile) bool {
	for _, item := range file.Items {
		switch item.Node.(type) {
		case ast.ImportDecl, ast.ProviderDecl, ast.ToolDecl, ast.AgentDecl, ast.WorkflowDecl:
			return true
		}
	}
	// Also need haira if any statement uses stdlib functions
	return hasStdlibCalls(file)
}

func needsTimeImport(file *ast.SourceFile) bool {
	for _, item := range file.Items {
		if w, ok := item.Node.(ast.WorkflowDecl); ok {
			if blockHasRetryDecorator(w.Body) {
				return true
			}
		}
	}
	return false
}

func blockHasRetryDecorator(block ast.Block) bool {
	for _, stmt := range block.Statements {
		if s, ok := stmt.Node.(ast.StepStmt); ok {
			for _, dec := range s.Decorators {
				if dec.Name.Node == "retry" {
					return true
				}
			}
		}
	}
	return false
}

func needsSyncImport(file *ast.SourceFile) bool {
	for _, item := range file.Items {
		switch it := item.Node.(type) {
		case ast.FunctionDef:
			if blockHasSpawn(it.Body) {
				return true
			}
		case ast.WorkflowDecl:
			if blockHasSpawn(it.Body) {
				return true
			}
		}
	}
	return false
}

func blockHasTry(block ast.Block) bool {
	for _, stmt := range block.Statements {
		if _, ok := stmt.Node.(ast.TryStmt); ok {
			return true
		}
	}
	return false
}

func blockHasSpawn(block ast.Block) bool {
	for _, stmt := range block.Statements {
		switch s := stmt.Node.(type) {
		case ast.ExprStmt:
			if _, ok := s.Value.Node.(ast.SpawnExpr); ok {
				return true
			}
		case ast.AssignStmt:
			if _, ok := s.Value.Node.(ast.SpawnExpr); ok {
				return true
			}
		case ast.IfStmt:
			if blockHasSpawn(s.ThenBranch) {
				return true
			}
		case ast.ForStmt:
			if blockHasSpawn(s.Body) {
				return true
			}
		case ast.WhileStmt:
			if blockHasSpawn(s.Body) {
				return true
			}
		}
	}
	return false
}

func blockHasInterpolatedString(block ast.Block) bool {
	for _, stmt := range block.Statements {
		if stmtHasInterpolatedString(stmt) {
			return true
		}
	}
	return false
}

func stmtHasInterpolatedString(stmt ast.Statement) bool {
	switch s := stmt.Node.(type) {
	case ast.ReturnStmt:
		for _, v := range s.Values {
			if exprHasInterpolatedString(v) {
				return true
			}
		}
	case ast.ExprStmt:
		return exprHasInterpolatedString(s.Value)
	case ast.AssignStmt:
		return exprHasInterpolatedString(s.Value)
	case ast.IfStmt:
		return blockHasInterpolatedString(s.ThenBranch)
	case ast.ForStmt:
		return blockHasInterpolatedString(s.Body)
	case ast.WhileStmt:
		return blockHasInterpolatedString(s.Body)
	}
	return false
}

func exprHasInterpolatedString(expr ast.Expr) bool {
	switch e := expr.Node.(type) {
	case ast.LiteralExpr:
		if _, ok := e.Lit.(ast.InterpolatedStringLit); ok {
			return true
		}
	case ast.CallExpr:
		for _, a := range e.Args {
			if exprHasInterpolatedString(a.Value) {
				return true
			}
		}
		return exprHasInterpolatedString(e.Callee)
	case ast.BinaryExpr:
		return exprHasInterpolatedString(e.Left) || exprHasInterpolatedString(e.Right)
	case ast.UnaryExpr:
		return exprHasInterpolatedString(e.Operand)
	case ast.ParenExpr:
		return exprHasInterpolatedString(e.Inner)
	case ast.MethodCallExpr:
		for _, a := range e.Args {
			if exprHasInterpolatedString(a.Value) {
				return true
			}
		}
		return exprHasInterpolatedString(e.Receiver)
	case ast.PipeExpr:
		return exprHasInterpolatedString(e.Left) || exprHasInterpolatedString(e.Right)
	}
	return false
}

func hasStdlibCalls(file *ast.SourceFile) bool {
	for _, item := range file.Items {
		switch it := item.Node.(type) {
		case ast.FunctionDef:
			if blockHasStdlibCalls(it.Body) {
				return true
			}
		case ast.ItemStatement:
			if stmtHasStdlibCalls(it.Stmt) {
				return true
			}
		}
	}
	return false
}

func blockHasStdlibCalls(block ast.Block) bool {
	for _, stmt := range block.Statements {
		if stmtHasStdlibCalls(stmt) {
			return true
		}
	}
	return false
}

func stmtHasStdlibCalls(stmt ast.Statement) bool {
	switch s := stmt.Node.(type) {
	case ast.ExprStmt:
		return exprHasStdlibCall(s.Value)
	case ast.AssignStmt:
		return exprHasStdlibCall(s.Value)
	case ast.ReturnStmt:
		for _, v := range s.Values {
			if exprHasStdlibCall(v) {
				return true
			}
		}
	case ast.IfStmt:
		return blockHasStdlibCalls(s.ThenBranch)
	case ast.ForStmt:
		return exprHasStdlibCall(s.Iterator) || blockHasStdlibCalls(s.Body)
	}
	return false
}

func exprHasStdlibCall(expr ast.Expr) bool {
	switch e := expr.Node.(type) {
	case ast.CallExpr:
		if _, ok := ResolveStdlibCall(e); ok {
			return true
		}
	case ast.MethodCallExpr:
		if _, ok := ResolveStdlibMethodCall(e); ok {
			return true
		}
	case ast.PipeExpr:
		return exprHasStdlibCall(e.Left) || exprHasStdlibCall(e.Right)
	case ast.BinaryExpr:
		return exprHasStdlibCall(e.Left) || exprHasStdlibCall(e.Right)
	}
	return false
}

// Capitalize capitalizes the first letter.
func Capitalize(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// SnakeToPascal converts snake_case to PascalCase.
func SnakeToPascal(name string) string {
	parts := strings.Split(name, "_")
	result := make([]string, len(parts))
	for i, p := range parts {
		result[i] = Capitalize(p)
	}
	return strings.Join(result, "")
}
