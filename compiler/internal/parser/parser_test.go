package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/haira-lang/haira/internal/ast"
)

// helper: find the examples directory relative to this test file.
func examplesDir() string {
	// Walk up from compiler/internal/parser to compiler, then to project root.
	cwd, _ := os.Getwd()
	root := filepath.Join(cwd, "..", "..", "..")
	dir := filepath.Join(root, "examples")
	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		return dir
	}
	return ""
}

func parseFile(t *testing.T, name string) *ast.SourceFile {
	t.Helper()
	dir := examplesDir()
	if dir == "" {
		t.Skip("examples directory not found")
	}
	data, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	sf, errs := Parse(string(data))
	if len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("parse error: %s at [%d,%d)", e.Message, e.Span.Start, e.Span.End)
		}
		t.Fatalf("%s had %d parse errors", name, len(errs))
	}
	return sf
}

// ---------------------------------------------------------------------------
// All examples parse without errors
// ---------------------------------------------------------------------------

func TestAllExamplesParse(t *testing.T) {
	dir := examplesDir()
	if dir == "" {
		t.Skip("examples directory not found")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read examples dir: %v", err)
	}
	count := 0
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".haira" {
			continue
		}
		count++
		t.Run(entry.Name(), func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			_, errs := Parse(string(data))
			if len(errs) > 0 {
				for _, e := range errs {
					t.Errorf("parse error: %s", e.Message)
				}
			}
		})
	}
	if count == 0 {
		t.Fatal("no .haira files found in examples directory")
	}
}

// ---------------------------------------------------------------------------
// Item counts for specific examples
// ---------------------------------------------------------------------------

func TestHelloItemCount(t *testing.T) {
	sf := parseFile(t, "01-hello.haira")
	// import "io" + fn main() = 2 items
	if len(sf.Items) != 2 {
		t.Errorf("01-hello: expected 2 items, got %d", len(sf.Items))
	}
}

func TestVariablesItemCount(t *testing.T) {
	sf := parseFile(t, "02-variables.haira")
	// import "io" + fn main() = 2 items
	if len(sf.Items) != 2 {
		t.Errorf("02-variables: expected 2 items, got %d", len(sf.Items))
	}
}

func TestFunctionsItemCount(t *testing.T) {
	sf := parseFile(t, "03-functions.haira")
	// import + add + multiply + greet + main = 5
	if len(sf.Items) != 5 {
		t.Errorf("03-functions: expected 5 items, got %d", len(sf.Items))
	}
}

func TestControlFlowItemCount(t *testing.T) {
	sf := parseFile(t, "04-control-flow.haira")
	// import + classify + fibonacci + main = 4
	if len(sf.Items) != 4 {
		t.Errorf("04-control-flow: expected 4 items, got %d", len(sf.Items))
	}
}

// ---------------------------------------------------------------------------
// Item types
// ---------------------------------------------------------------------------

func TestHelloItemTypes(t *testing.T) {
	sf := parseFile(t, "01-hello.haira")
	if _, ok := sf.Items[0].Node.(ast.ImportDecl); !ok {
		t.Error("item 0 should be ImportDecl")
	}
	if fn, ok := sf.Items[1].Node.(ast.FunctionDef); !ok {
		t.Error("item 1 should be FunctionDef")
	} else if fn.Name.Node != "main" {
		t.Errorf("expected function name 'main', got %q", fn.Name.Node)
	}
}

func TestStructsItemTypes(t *testing.T) {
	sf := parseFile(t, "08-structs.haira")
	// import, TypeDef(User), MethodDef(User.greet), FunctionDef(main)
	found := struct {
		imports  int
		typeDefs int
		methods  int
		funcs    int
	}{}
	for _, item := range sf.Items {
		switch item.Node.(type) {
		case ast.ImportDecl:
			found.imports++
		case ast.TypeDef:
			found.typeDefs++
		case ast.MethodDef:
			found.methods++
		case ast.FunctionDef:
			found.funcs++
		}
	}
	if found.imports != 1 {
		t.Errorf("expected 1 import, got %d", found.imports)
	}
	if found.typeDefs != 1 {
		t.Errorf("expected 1 type def, got %d", found.typeDefs)
	}
	if found.methods != 1 {
		t.Errorf("expected 1 method, got %d", found.methods)
	}
	if found.funcs != 1 {
		t.Errorf("expected 1 function, got %d", found.funcs)
	}
}

// ---------------------------------------------------------------------------
// Agentic declarations
// ---------------------------------------------------------------------------

func TestAgenticExample(t *testing.T) {
	sf := parseFile(t, "07-agentic.haira")
	found := struct {
		imports   int
		providers int
		tools     int
		agents    int
		workflows int
		functions int
	}{}
	for _, item := range sf.Items {
		switch item.Node.(type) {
		case ast.ImportDecl:
			found.imports++
		case ast.ProviderDecl:
			found.providers++
		case ast.ToolDecl:
			found.tools++
		case ast.AgentDecl:
			found.agents++
		case ast.WorkflowDecl:
			found.workflows++
		case ast.FunctionDef:
			found.functions++
		}
	}
	if found.providers != 1 {
		t.Errorf("expected 1 provider, got %d", found.providers)
	}
	if found.tools != 1 {
		t.Errorf("expected 1 tool, got %d", found.tools)
	}
	if found.agents != 1 {
		t.Errorf("expected 1 agent, got %d", found.agents)
	}
	if found.workflows != 1 {
		t.Errorf("expected 1 workflow, got %d", found.workflows)
	}
	if found.functions != 1 {
		t.Errorf("expected 1 function (main), got %d", found.functions)
	}
}

func TestProviderDecl(t *testing.T) {
	src := `provider openai {
    api_key: env("OPENAI_KEY")
    model: "gpt-4"
}`
	sf, errs := Parse(src)
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	if len(sf.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(sf.Items))
	}
	p, ok := sf.Items[0].Node.(ast.ProviderDecl)
	if !ok {
		t.Fatal("expected ProviderDecl")
	}
	if p.Name.Node != "openai" {
		t.Errorf("expected name 'openai', got %q", p.Name.Node)
	}
	if len(p.Fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(p.Fields))
	}
}

func TestAgentDecl(t *testing.T) {
	src := `agent Bot {
    model: openai
    system: "You are helpful."
    tools: [search, calculate]
    temperature: 0.7
}`
	sf, errs := Parse(src)
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	a, ok := sf.Items[0].Node.(ast.AgentDecl)
	if !ok {
		t.Fatal("expected AgentDecl")
	}
	if a.Name.Node != "Bot" {
		t.Errorf("expected name 'Bot', got %q", a.Name.Node)
	}
	if len(a.Fields) != 4 {
		t.Errorf("expected 4 fields, got %d", len(a.Fields))
	}
}

func TestToolDecl(t *testing.T) {
	src := `tool search(query: string) -> string {
    """
    Search the web for information.
    """
    return "result for " + query
}`
	sf, errs := Parse(src)
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	tool, ok := sf.Items[0].Node.(ast.ToolDecl)
	if !ok {
		t.Fatal("expected ToolDecl")
	}
	if tool.Name.Node != "search" {
		t.Errorf("expected name 'search', got %q", tool.Name.Node)
	}
	if len(tool.Params) != 1 {
		t.Errorf("expected 1 param, got %d", len(tool.Params))
	}
	if tool.Description == "" {
		t.Error("expected non-empty description")
	}
	if tool.Body == nil {
		t.Error("expected non-nil body")
	}
}

func TestWorkflowDecl(t *testing.T) {
	src := `@webhook("/api/chat")
workflow Chat(msg: string) -> string {
    return "ok"
}`
	sf, errs := Parse(src)
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	wf, ok := sf.Items[0].Node.(ast.WorkflowDecl)
	if !ok {
		t.Fatal("expected WorkflowDecl")
	}
	if wf.Name.Node != "Chat" {
		t.Errorf("expected name 'Chat', got %q", wf.Name.Node)
	}
	if wf.Trigger == nil {
		t.Fatal("expected non-nil trigger")
	}
	if wf.Trigger.Name.Node != "webhook" {
		t.Errorf("expected trigger name 'webhook', got %q", wf.Trigger.Name.Node)
	}
}

// ---------------------------------------------------------------------------
// Expression parsing
// ---------------------------------------------------------------------------

func TestArithmeticExpr(t *testing.T) {
	src := `fn main() { x = 1 + 2 * 3 }`
	sf, errs := Parse(src)
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	fn := sf.Items[0].Node.(ast.FunctionDef)
	// x = 1 + 2 * 3
	assign := fn.Body.Statements[0].Node.(ast.AssignStmt)
	// RHS should be BinaryExpr(+) with right being BinaryExpr(*)
	bin, ok := assign.Value.Node.(ast.BinaryExpr)
	if !ok {
		t.Fatal("expected BinaryExpr")
	}
	if bin.Op.Node != ast.OpAdd {
		t.Errorf("expected OpAdd, got %v", bin.Op.Node)
	}
	// Right should be multiply
	rightBin, ok := bin.Right.Node.(ast.BinaryExpr)
	if !ok {
		t.Fatal("right should be BinaryExpr for multiplication")
	}
	if rightBin.Op.Node != ast.OpMul {
		t.Errorf("expected OpMul, got %v", rightBin.Op.Node)
	}
}

func TestComparisonExpr(t *testing.T) {
	src := `fn main() { x = a > b }`
	sf, errs := Parse(src)
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	fn := sf.Items[0].Node.(ast.FunctionDef)
	assign := fn.Body.Statements[0].Node.(ast.AssignStmt)
	bin, ok := assign.Value.Node.(ast.BinaryExpr)
	if !ok {
		t.Fatal("expected BinaryExpr")
	}
	if bin.Op.Node != ast.OpGt {
		t.Errorf("expected OpGt, got %v", bin.Op.Node)
	}
}

func TestLogicalExpr(t *testing.T) {
	src := `fn main() { x = a and b or c }`
	sf, errs := Parse(src)
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	fn := sf.Items[0].Node.(ast.FunctionDef)
	assign := fn.Body.Statements[0].Node.(ast.AssignStmt)
	// Should be Or(And(a, b), c) since `and` has higher precedence
	bin, ok := assign.Value.Node.(ast.BinaryExpr)
	if !ok {
		t.Fatal("expected BinaryExpr")
	}
	if bin.Op.Node != ast.OpOr {
		t.Errorf("expected OpOr at top level, got %v", bin.Op.Node)
	}
	leftBin, ok := bin.Left.Node.(ast.BinaryExpr)
	if !ok {
		t.Fatal("left of 'or' should be BinaryExpr")
	}
	if leftBin.Op.Node != ast.OpAnd {
		t.Errorf("expected OpAnd, got %v", leftBin.Op.Node)
	}
}

func TestUnaryExpr(t *testing.T) {
	src := `fn main() { x = -42 }`
	sf, errs := Parse(src)
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	fn := sf.Items[0].Node.(ast.FunctionDef)
	assign := fn.Body.Statements[0].Node.(ast.AssignStmt)
	unary, ok := assign.Value.Node.(ast.UnaryExpr)
	if !ok {
		t.Fatal("expected UnaryExpr")
	}
	if unary.Op.Node != ast.OpNeg {
		t.Errorf("expected OpNeg, got %v", unary.Op.Node)
	}
}

// ---------------------------------------------------------------------------
// Statement parsing
// ---------------------------------------------------------------------------

func TestIfStatement(t *testing.T) {
	src := `fn main() {
    if x > 0 {
        io.println("positive")
    } else {
        io.println("non-positive")
    }
}`
	sf, errs := Parse(src)
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	fn := sf.Items[0].Node.(ast.FunctionDef)
	ifStmt, ok := fn.Body.Statements[0].Node.(ast.IfStmt)
	if !ok {
		t.Fatal("expected IfStmt")
	}
	if ifStmt.ElseBranch == nil {
		t.Error("expected else branch")
	}
}

func TestForStatement(t *testing.T) {
	src := `fn main() {
    for item in items {
        io.println(item)
    }
}`
	sf, errs := Parse(src)
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	fn := sf.Items[0].Node.(ast.FunctionDef)
	forStmt, ok := fn.Body.Statements[0].Node.(ast.ForStmt)
	if !ok {
		t.Fatal("expected ForStmt")
	}
	pat, ok := forStmt.Pattern.(ast.SinglePattern)
	if !ok {
		t.Fatal("expected SinglePattern")
	}
	if pat.Name.Node != "item" {
		t.Errorf("expected pattern name 'item', got %q", pat.Name.Node)
	}
}

func TestWhileStatement(t *testing.T) {
	src := `fn main() {
    while x > 0 {
        x = x - 1
    }
}`
	sf, errs := Parse(src)
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	fn := sf.Items[0].Node.(ast.FunctionDef)
	_, ok := fn.Body.Statements[0].Node.(ast.WhileStmt)
	if !ok {
		t.Fatal("expected WhileStmt")
	}
}

func TestReturnMultipleValues(t *testing.T) {
	src := `fn divide(a: int, b: int) -> int {
    return a / b, nil
}`
	sf, errs := Parse(src)
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	fn := sf.Items[0].Node.(ast.FunctionDef)
	ret, ok := fn.Body.Statements[0].Node.(ast.ReturnStmt)
	if !ok {
		t.Fatal("expected ReturnStmt")
	}
	if len(ret.Values) != 2 {
		t.Errorf("expected 2 return values, got %d", len(ret.Values))
	}
}

func TestCompoundAssignment(t *testing.T) {
	src := `fn main() { x = 0
    x += 1 }`
	sf, errs := Parse(src)
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	fn := sf.Items[0].Node.(ast.FunctionDef)
	// Second statement should be an AssignStmt with desugared value
	assign, ok := fn.Body.Statements[1].Node.(ast.AssignStmt)
	if !ok {
		t.Fatal("expected AssignStmt")
	}
	// Desugared: x = x + 1
	bin, ok := assign.Value.Node.(ast.BinaryExpr)
	if !ok {
		t.Fatal("expected BinaryExpr in desugared compound assignment")
	}
	if bin.Op.Node != ast.OpAdd {
		t.Errorf("expected OpAdd, got %v", bin.Op.Node)
	}
}

func TestMultiAssignment(t *testing.T) {
	// Multi-assignment: a, b = some_call()  (RHS is a single expression)
	src := `fn main() { a, b = divide(10, 3) }`
	sf, errs := Parse(src)
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	fn := sf.Items[0].Node.(ast.FunctionDef)
	assign, ok := fn.Body.Statements[0].Node.(ast.AssignStmt)
	if !ok {
		t.Fatal("expected AssignStmt")
	}
	if len(assign.Targets) != 2 {
		t.Errorf("expected 2 targets, got %d", len(assign.Targets))
	}
}

// ---------------------------------------------------------------------------
// Match
// ---------------------------------------------------------------------------

func TestMatchStatement(t *testing.T) {
	src := `fn main() {
    match status {
        200 => "OK"
        404 => "Not Found"
        _ => "Unknown"
    }
}`
	sf, errs := Parse(src)
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	fn := sf.Items[0].Node.(ast.FunctionDef)
	matchStmt, ok := fn.Body.Statements[0].Node.(ast.MatchStmt)
	if !ok {
		t.Fatal("expected MatchStmt")
	}
	if len(matchStmt.Match.Arms) != 3 {
		t.Errorf("expected 3 match arms, got %d", len(matchStmt.Match.Arms))
	}
	// Last arm should be wildcard
	lastPat := matchStmt.Match.Arms[2].Pattern.Node
	if _, ok := lastPat.(ast.WildcardPattern); !ok {
		t.Error("last match arm should be WildcardPattern")
	}
}

// ---------------------------------------------------------------------------
// Enum
// ---------------------------------------------------------------------------

func TestEnumDecl(t *testing.T) {
	src := `enum Color {
    Red
    Green
    Blue
}`
	sf, errs := Parse(src)
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	ed, ok := sf.Items[0].Node.(ast.EnumDef)
	if !ok {
		t.Fatal("expected EnumDef")
	}
	if ed.Name.Node != "Color" {
		t.Errorf("expected name 'Color', got %q", ed.Name.Node)
	}
	if len(ed.Variants) != 3 {
		t.Errorf("expected 3 variants, got %d", len(ed.Variants))
	}
}

// ---------------------------------------------------------------------------
// Type def
// ---------------------------------------------------------------------------

func TestTypeDef(t *testing.T) {
	src := `struct User {
    name: string
    age: int
}`
	sf, errs := Parse(src)
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	td, ok := sf.Items[0].Node.(ast.TypeDef)
	if !ok {
		t.Fatal("expected TypeDef")
	}
	if td.Name.Node != "User" {
		t.Errorf("expected name 'User', got %q", td.Name.Node)
	}
	if len(td.Fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(td.Fields))
	}
}

// ---------------------------------------------------------------------------
// Interpolated string
// ---------------------------------------------------------------------------

func TestInterpolatedStringParts(t *testing.T) {
	src := `fn main() { x = "hello ${name}!" }`
	sf, errs := Parse(src)
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	fn := sf.Items[0].Node.(ast.FunctionDef)
	assign := fn.Body.Statements[0].Node.(ast.AssignStmt)
	lit, ok := assign.Value.Node.(ast.LiteralExpr)
	if !ok {
		t.Fatal("expected LiteralExpr")
	}
	interp, ok := lit.Lit.(ast.InterpolatedStringLit)
	if !ok {
		t.Fatal("expected InterpolatedStringLit")
	}
	// Should have 3 parts: "hello ", ${name}, "!"
	if len(interp.Parts) != 3 {
		t.Fatalf("expected 3 parts, got %d", len(interp.Parts))
	}
	if lp, ok := interp.Parts[0].(ast.LiteralPart); !ok || lp.Value != "hello " {
		t.Errorf("part 0: expected LiteralPart 'hello ', got %v", interp.Parts[0])
	}
	if _, ok := interp.Parts[1].(ast.ExprPart); !ok {
		t.Error("part 1: expected ExprPart")
	}
	if lp, ok := interp.Parts[2].(ast.LiteralPart); !ok || lp.Value != "!" {
		t.Errorf("part 2: expected LiteralPart '!', got %v", interp.Parts[2])
	}
}

// ---------------------------------------------------------------------------
// Lambda
// ---------------------------------------------------------------------------

func TestLambdaExpr(t *testing.T) {
	src := `fn main() { f = x => x + 1 }`
	sf, errs := Parse(src)
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	fn := sf.Items[0].Node.(ast.FunctionDef)
	assign := fn.Body.Statements[0].Node.(ast.AssignStmt)
	lambda, ok := assign.Value.Node.(ast.LambdaExpr)
	if !ok {
		t.Fatal("expected LambdaExpr")
	}
	if len(lambda.Params) != 1 {
		t.Errorf("expected 1 param, got %d", len(lambda.Params))
	}
	if lambda.Params[0].Name.Node != "x" {
		t.Errorf("expected param name 'x', got %q", lambda.Params[0].Name.Node)
	}
}

// ---------------------------------------------------------------------------
// List and map literals
// ---------------------------------------------------------------------------

func TestListLiteral(t *testing.T) {
	src := `fn main() { xs = [1, 2, 3] }`
	sf, errs := Parse(src)
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	fn := sf.Items[0].Node.(ast.FunctionDef)
	assign := fn.Body.Statements[0].Node.(ast.AssignStmt)
	list, ok := assign.Value.Node.(ast.ListExpr)
	if !ok {
		t.Fatal("expected ListExpr")
	}
	if len(list.Elems) != 3 {
		t.Errorf("expected 3 elements, got %d", len(list.Elems))
	}
}

func TestMapLiteral(t *testing.T) {
	src := `fn main() { m = {"a": 1, "b": 2} }`
	sf, errs := Parse(src)
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	fn := sf.Items[0].Node.(ast.FunctionDef)
	assign := fn.Body.Statements[0].Node.(ast.AssignStmt)
	mapExpr, ok := assign.Value.Node.(ast.MapExpr)
	if !ok {
		t.Fatal("expected MapExpr")
	}
	if len(mapExpr.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(mapExpr.Entries))
	}
}

// ---------------------------------------------------------------------------
// Try/catch
// ---------------------------------------------------------------------------

func TestTryCatch(t *testing.T) {
	src := `fn main() {
    try {
        io.println("try")
    } catch e {
        io.println(e)
    }
}`
	sf, errs := Parse(src)
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	fn := sf.Items[0].Node.(ast.FunctionDef)
	tryStmt, ok := fn.Body.Statements[0].Node.(ast.TryStmt)
	if !ok {
		t.Fatal("expected TryStmt")
	}
	if tryStmt.ErrorName.Node != "e" {
		t.Errorf("expected error name 'e', got %q", tryStmt.ErrorName.Node)
	}
}

// ---------------------------------------------------------------------------
// Defer
// ---------------------------------------------------------------------------

func TestDefer(t *testing.T) {
	src := `fn main() { defer io.println("done") }`
	sf, errs := Parse(src)
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	fn := sf.Items[0].Node.(ast.FunctionDef)
	_, ok := fn.Body.Statements[0].Node.(ast.DeferStmt)
	if !ok {
		t.Fatal("expected DeferStmt")
	}
}

// ---------------------------------------------------------------------------
// Instance creation
// ---------------------------------------------------------------------------

func TestInstanceCreation(t *testing.T) {
	src := `fn main() { u = User{ name = "Alice", age = 30 } }`
	sf, errs := Parse(src)
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	fn := sf.Items[0].Node.(ast.FunctionDef)
	assign := fn.Body.Statements[0].Node.(ast.AssignStmt)
	inst, ok := assign.Value.Node.(ast.InstanceExpr)
	if !ok {
		t.Fatal("expected InstanceExpr")
	}
	if inst.TypeName.Node != "User" {
		t.Errorf("expected type 'User', got %q", inst.TypeName.Node)
	}
	if len(inst.Fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(inst.Fields))
	}
}

// ---------------------------------------------------------------------------
// Function with params and return type
// ---------------------------------------------------------------------------

func TestFunctionDefParams(t *testing.T) {
	src := `fn add(a: int, b: int) -> int { return a + b }`
	sf, errs := Parse(src)
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	fn := sf.Items[0].Node.(ast.FunctionDef)
	if fn.Name.Node != "add" {
		t.Errorf("expected name 'add', got %q", fn.Name.Node)
	}
	if len(fn.Params) != 2 {
		t.Errorf("expected 2 params, got %d", len(fn.Params))
	}
	if fn.ReturnTy == nil {
		t.Error("expected return type")
	}
}

// ---------------------------------------------------------------------------
// Method definition
// ---------------------------------------------------------------------------

func TestMethodDef(t *testing.T) {
	src := `User.greet() -> string { return "hello" }`
	sf, errs := Parse(src)
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	md, ok := sf.Items[0].Node.(ast.MethodDef)
	if !ok {
		t.Fatal("expected MethodDef")
	}
	if md.TypeName.Node != "User" {
		t.Errorf("expected type name 'User', got %q", md.TypeName.Node)
	}
	if md.Name.Node != "greet" {
		t.Errorf("expected method name 'greet', got %q", md.Name.Node)
	}
}

// ---------------------------------------------------------------------------
// Import
// ---------------------------------------------------------------------------

func TestImportDecl(t *testing.T) {
	src := `import "io"`
	sf, errs := Parse(src)
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	imp, ok := sf.Items[0].Node.(ast.ImportDecl)
	if !ok {
		t.Fatal("expected ImportDecl")
	}
	if imp.Path != "io" {
		t.Errorf("expected path 'io', got %q", imp.Path)
	}
}

// ---------------------------------------------------------------------------
// Pipe expression
// ---------------------------------------------------------------------------

func TestPipeExpr(t *testing.T) {
	src := `fn main() { x = data | transform | output }`
	sf, errs := Parse(src)
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	fn := sf.Items[0].Node.(ast.FunctionDef)
	assign := fn.Body.Statements[0].Node.(ast.AssignStmt)
	// Should be PipeExpr(PipeExpr(data, transform), output)
	pipe, ok := assign.Value.Node.(ast.PipeExpr)
	if !ok {
		t.Fatal("expected PipeExpr")
	}
	innerPipe, ok := pipe.Left.Node.(ast.PipeExpr)
	if !ok {
		t.Fatal("left of outer pipe should be PipeExpr")
	}
	if ident, ok := innerPipe.Left.Node.(ast.IdentExpr); !ok || ident.Name != "data" {
		t.Error("innermost left should be 'data'")
	}
}

// ---------------------------------------------------------------------------
// Empty source
// ---------------------------------------------------------------------------

func TestEmptySource(t *testing.T) {
	sf, errs := Parse("")
	if len(errs) > 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}
	if len(sf.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(sf.Items))
	}
}

// ---------------------------------------------------------------------------
// Span coverage
// ---------------------------------------------------------------------------

func TestSpanCoverage(t *testing.T) {
	src := `fn main() { x = 42 }`
	sf, errs := Parse(src)
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	if sf.Span.Start != 0 {
		t.Errorf("source file span start should be 0, got %d", sf.Span.Start)
	}
	if sf.Span.End <= 0 {
		t.Errorf("source file span end should be > 0, got %d", sf.Span.End)
	}
}
