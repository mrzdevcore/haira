package checker

import (
	"testing"

	"github.com/haira-lang/haira/internal/ast"
	"github.com/haira-lang/haira/internal/parser"
)

func parse(t *testing.T, src string) *ast.SourceFile {
	t.Helper()
	file, errs := parser.Parse(src)
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	return file
}

// ---------------------------------------------------------------------------
// types.go tests
// ---------------------------------------------------------------------------

func TestTypeStringPrimitives(t *testing.T) {
	cases := []struct {
		ty   Type
		want string
	}{
		{IntType{}, "int"},
		{FloatType{}, "float"},
		{StringType{}, "string"},
		{BoolType{}, "bool"},
		{AnyType{}, "any"},
		{VoidType{}, "void"},
		{ErrorType{}, "error"},
	}
	for _, c := range cases {
		if got := c.ty.String(); got != c.want {
			t.Errorf("%T.String() = %q, want %q", c.ty, got, c.want)
		}
	}
}

func TestTypeStringCompound(t *testing.T) {
	list := ListType{Elem: IntType{}}
	if got := list.String(); got != "[int]" {
		t.Errorf("ListType.String() = %q, want %q", got, "[int]")
	}

	m := MapType{Key: StringType{}, Value: FloatType{}}
	if got := m.String(); got != "{string: float}" {
		t.Errorf("MapType.String() = %q, want %q", got, "{string: float}")
	}

	s := StructType{Name: "User", Fields: map[string]Type{"name": StringType{}}}
	if got := s.String(); got != "User" {
		t.Errorf("StructType.String() = %q, want %q", got, "User")
	}
}

func TestTypeEquals(t *testing.T) {
	// Identical types
	if !TypeEquals(IntType{}, IntType{}) {
		t.Error("int == int should be true")
	}
	if !TypeEquals(StringType{}, StringType{}) {
		t.Error("string == string should be true")
	}

	// Different types
	if TypeEquals(IntType{}, StringType{}) {
		t.Error("int == string should be false")
	}

	// AnyType matches everything
	if !TypeEquals(AnyType{}, IntType{}) {
		t.Error("any == int should be true")
	}
	if !TypeEquals(FloatType{}, AnyType{}) {
		t.Error("float == any should be true")
	}

	// Nil
	if TypeEquals(nil, IntType{}) {
		t.Error("nil == int should be false")
	}
	if TypeEquals(IntType{}, nil) {
		t.Error("int == nil should be false")
	}

	// List equality
	if !TypeEquals(ListType{Elem: IntType{}}, ListType{Elem: IntType{}}) {
		t.Error("[int] == [int] should be true")
	}
	if TypeEquals(ListType{Elem: IntType{}}, ListType{Elem: StringType{}}) {
		t.Error("[int] == [string] should be false")
	}

	// Map equality
	m1 := MapType{Key: StringType{}, Value: IntType{}}
	m2 := MapType{Key: StringType{}, Value: IntType{}}
	if !TypeEquals(m1, m2) {
		t.Error("{string: int} == {string: int} should be true")
	}

	// Struct equality (by name)
	s1 := StructType{Name: "User"}
	s2 := StructType{Name: "User"}
	s3 := StructType{Name: "Post"}
	if !TypeEquals(s1, s2) {
		t.Error("User == User should be true")
	}
	if TypeEquals(s1, s3) {
		t.Error("User == Post should be false")
	}

	// FuncType equality
	f1 := FuncType{Params: []Type{IntType{}}, Return: StringType{}}
	f2 := FuncType{Params: []Type{IntType{}}, Return: StringType{}}
	f3 := FuncType{Params: []Type{IntType{}, IntType{}}, Return: StringType{}}
	if !TypeEquals(f1, f2) {
		t.Error("fn(int)->string == fn(int)->string should be true")
	}
	if TypeEquals(f1, f3) {
		t.Error("fn(int)->string == fn(int,int)->string should be false")
	}
}

// ---------------------------------------------------------------------------
// env.go tests
// ---------------------------------------------------------------------------

func TestEnvVarLookup(t *testing.T) {
	env := NewEnv()
	env.DefineVar("x", IntType{})
	ty, ok := env.LookupVar("x")
	if !ok {
		t.Fatal("x should be defined")
	}
	if !TypeEquals(ty, IntType{}) {
		t.Errorf("x type = %s, want int", ty)
	}
}

func TestEnvChildScope(t *testing.T) {
	parent := NewEnv()
	parent.DefineVar("x", IntType{})
	child := parent.Child()
	child.DefineVar("y", StringType{})

	// Child can see parent's vars
	if _, ok := child.LookupVar("x"); !ok {
		t.Error("child should see parent's x")
	}
	// Parent can't see child's vars
	if _, ok := parent.LookupVar("y"); ok {
		t.Error("parent should not see child's y")
	}
	// Child shadows parent
	child.DefineVar("x", FloatType{})
	ty, _ := child.LookupVar("x")
	if !TypeEquals(ty, FloatType{}) {
		t.Errorf("child x = %s, want float (shadowed)", ty)
	}
	// Parent unchanged
	ty, _ = parent.LookupVar("x")
	if !TypeEquals(ty, IntType{}) {
		t.Errorf("parent x = %s, want int (unchanged)", ty)
	}
}

func TestEnvStdlibRegistered(t *testing.T) {
	env := NewEnv()
	// Check a few stdlib functions
	fn, ok := env.LookupFunc("io.println")
	if !ok {
		t.Fatal("io.println should be registered")
	}
	if !TypeEquals(fn.Return, VoidType{}) {
		t.Errorf("io.println return = %s, want void", fn.Return)
	}

	fn, ok = env.LookupFunc("len")
	if !ok {
		t.Fatal("len should be registered")
	}
	if !TypeEquals(fn.Return, IntType{}) {
		t.Errorf("len return = %s, want int", fn.Return)
	}

	fn, ok = env.LookupFunc("env")
	if !ok {
		t.Fatal("env should be registered")
	}
	if !TypeEquals(fn.Return, StringType{}) {
		t.Errorf("env return = %s, want string", fn.Return)
	}
}

func TestEnvTypeLookup(t *testing.T) {
	env := NewEnv()
	env.DefineType("User", StructType{Name: "User", Fields: map[string]Type{"name": StringType{}}})
	ty, ok := env.LookupType("User")
	if !ok {
		t.Fatal("User should be defined")
	}
	st, ok := ty.(StructType)
	if !ok {
		t.Fatalf("User type = %T, want StructType", ty)
	}
	if st.Name != "User" {
		t.Errorf("User.Name = %q, want %q", st.Name, "User")
	}
}

// ---------------------------------------------------------------------------
// checker.go — literal inference
// ---------------------------------------------------------------------------

func TestInferLiterals(t *testing.T) {
	file := parse(t, `
fn main() {
	x = 42
	y = 3.14
	z = "hello"
	w = true
}
`)
	info, diags := Check(file)
	if len(diags) > 0 {
		t.Errorf("unexpected diagnostics: %v", diags)
	}

	// Check that expression types were recorded
	found := map[string]bool{}
	for _, ty := range info.ExprTypes {
		found[ty.String()] = true
	}
	if !found["int"] {
		t.Error("expected int literal to be inferred")
	}
	if !found["float"] {
		t.Error("expected float literal to be inferred")
	}
	if !found["string"] {
		t.Error("expected string literal to be inferred")
	}
	if !found["bool"] {
		t.Error("expected bool literal to be inferred")
	}
}

// ---------------------------------------------------------------------------
// checker.go — variable type flow
// ---------------------------------------------------------------------------

func TestInferVariableType(t *testing.T) {
	file := parse(t, `
fn main() {
	x = 42
	y = x
}
`)
	info, diags := Check(file)
	if len(diags) > 0 {
		t.Errorf("unexpected diagnostics: %v", diags)
	}

	// y should get its type from x which is int
	varTypes := map[string]bool{}
	for _, ty := range info.VarTypes {
		varTypes[ty.String()] = true
	}
	if !varTypes["int"] {
		t.Error("expected variable to be inferred as int")
	}
}

// ---------------------------------------------------------------------------
// checker.go — arithmetic inference
// ---------------------------------------------------------------------------

func TestInferArithmetic(t *testing.T) {
	file := parse(t, `
fn main() {
	a = 1 + 2
	b = 3.0 * 4.0
	c = "hello" + " world"
}
`)
	info, diags := Check(file)
	if len(diags) > 0 {
		t.Errorf("unexpected diagnostics: %v", diags)
	}

	types := map[string]bool{}
	for _, ty := range info.ExprTypes {
		types[ty.String()] = true
	}
	if !types["int"] {
		t.Error("expected int from 1 + 2")
	}
	if !types["float"] {
		t.Error("expected float from 3.0 * 4.0")
	}
	if !types["string"] {
		t.Error("expected string from string + string")
	}
}

func TestInferComparison(t *testing.T) {
	file := parse(t, `
fn main() {
	x = 1 < 2
	y = 3 == 3
	z = true and false
}
`)
	info, diags := Check(file)
	if len(diags) > 0 {
		t.Errorf("unexpected diagnostics: %v", diags)
	}

	// All results should be bool
	boolCount := 0
	for _, ty := range info.ExprTypes {
		if TypeEquals(ty, BoolType{}) {
			boolCount++
		}
	}
	if boolCount == 0 {
		t.Error("expected at least one bool expression type")
	}
}

// ---------------------------------------------------------------------------
// checker.go — function call inference
// ---------------------------------------------------------------------------

func TestInferFunctionReturnType(t *testing.T) {
	file := parse(t, `
fn add(a: int, b: int) -> int {
	return a + b
}

fn main() {
	result = add(1, 2)
}
`)
	info, diags := Check(file)
	if len(diags) > 0 {
		t.Errorf("unexpected diagnostics: %v", diags)
	}

	// 'result' should be int
	foundInt := false
	for _, ty := range info.VarTypes {
		if TypeEquals(ty, IntType{}) {
			foundInt = true
		}
	}
	if !foundInt {
		t.Error("expected result to be inferred as int from add() return")
	}
}

func TestInferStdlibCall(t *testing.T) {
	file := parse(t, `
import "io"

fn main() {
	x = len([1, 2, 3])
}
`)
	info, diags := Check(file)
	if len(diags) > 0 {
		t.Errorf("unexpected diagnostics: %v", diags)
	}

	foundInt := false
	for _, ty := range info.VarTypes {
		if TypeEquals(ty, IntType{}) {
			foundInt = true
		}
	}
	if !foundInt {
		t.Error("expected len() return to be inferred as int")
	}
}

// ---------------------------------------------------------------------------
// checker.go — list and map inference
// ---------------------------------------------------------------------------

func TestInferHomogeneousList(t *testing.T) {
	file := parse(t, `
fn main() {
	nums = [1, 2, 3]
}
`)
	info, diags := Check(file)
	if len(diags) > 0 {
		t.Errorf("unexpected diagnostics: %v", diags)
	}

	foundIntList := false
	for _, ty := range info.ExprTypes {
		if lt, ok := ty.(ListType); ok {
			if TypeEquals(lt.Elem, IntType{}) {
				foundIntList = true
			}
		}
	}
	if !foundIntList {
		t.Error("expected [1,2,3] to be inferred as [int]")
	}
}

func TestInferHeterogeneousList(t *testing.T) {
	file := parse(t, `
fn main() {
	mixed = [1, "two", 3]
}
`)
	info, diags := Check(file)
	if len(diags) > 0 {
		t.Errorf("unexpected diagnostics: %v", diags)
	}

	foundAnyList := false
	for _, ty := range info.ExprTypes {
		if lt, ok := ty.(ListType); ok {
			if isAny(lt.Elem) {
				foundAnyList = true
			}
		}
	}
	if !foundAnyList {
		t.Error("expected [1, 'two', 3] to be inferred as [any]")
	}
}

func TestInferMap(t *testing.T) {
	file := parse(t, `
fn main() {
	m = {"a": 1, "b": 2}
}
`)
	info, diags := Check(file)
	if len(diags) > 0 {
		t.Errorf("unexpected diagnostics: %v", diags)
	}

	foundMap := false
	for _, ty := range info.ExprTypes {
		if mt, ok := ty.(MapType); ok {
			if TypeEquals(mt.Key, StringType{}) && TypeEquals(mt.Value, IntType{}) {
				foundMap = true
			}
		}
	}
	if !foundMap {
		t.Error("expected map literal to be inferred as {string: int}")
	}
}

// ---------------------------------------------------------------------------
// checker.go — struct inference
// ---------------------------------------------------------------------------

func TestInferStructInstance(t *testing.T) {
	file := parse(t, `
struct User {
	name: string
	age: int
}

fn main() {
	u = User{name = "Alice", age = 30}
}
`)
	info, diags := Check(file)
	if len(diags) > 0 {
		t.Errorf("unexpected diagnostics: %v", diags)
	}

	foundUser := false
	for _, ty := range info.VarTypes {
		if st, ok := ty.(StructType); ok && st.Name == "User" {
			foundUser = true
		}
	}
	if !foundUser {
		t.Error("expected u to be inferred as User")
	}
}

func TestInferFieldAccess(t *testing.T) {
	file := parse(t, `
struct User {
	name: string
	age: int
}

fn main() {
	u = User{name = "Alice", age = 30}
	n = u.name
}
`)
	info, diags := Check(file)
	if len(diags) > 0 {
		t.Errorf("unexpected diagnostics: %v", diags)
	}

	// n should be string (from User.name)
	foundString := false
	for _, ty := range info.VarTypes {
		if TypeEquals(ty, StringType{}) {
			foundString = true
		}
	}
	if !foundString {
		t.Error("expected u.name to be inferred as string")
	}
}

// ---------------------------------------------------------------------------
// checker.go — type annotation mismatch
// ---------------------------------------------------------------------------

func TestNegateStringError(t *testing.T) {
	file := parse(t, `
fn main() {
	x = -"hello"
}
`)
	_, diags := Check(file)
	found := false
	for _, d := range diags {
		if contains(d.Message, "cannot negate") {
			found = true
		}
	}
	if !found {
		t.Error("expected 'cannot negate' error for -\"hello\"")
	}
}

// ---------------------------------------------------------------------------
// checker.go — unknown field error
// ---------------------------------------------------------------------------

func TestUnknownFieldError(t *testing.T) {
	file := parse(t, `
struct User {
	name: string
}

fn main() {
	u = User{name = "Alice"}
	x = u.email
}
`)
	_, diags := Check(file)
	found := false
	for _, d := range diags {
		if contains(d.Message, "unknown field") && contains(d.Message, "email") {
			found = true
		}
	}
	if !found {
		t.Error("expected 'unknown field email' error")
	}
}

// ---------------------------------------------------------------------------
// checker.go — enum registration
// ---------------------------------------------------------------------------

func TestEnumRegistered(t *testing.T) {
	file := parse(t, `
enum Color {
	Red
	Green
	Blue
}

fn main() {
	c = Color.Red
}
`)
	_, diags := Check(file)
	// Should not crash — enum should be registered
	_ = diags
}

// ---------------------------------------------------------------------------
// checker.go — unary operators
// ---------------------------------------------------------------------------

func TestInferUnaryNeg(t *testing.T) {
	file := parse(t, `
fn main() {
	x = -42
}
`)
	info, diags := Check(file)
	if len(diags) > 0 {
		t.Errorf("unexpected diagnostics: %v", diags)
	}

	foundInt := false
	for _, ty := range info.ExprTypes {
		if TypeEquals(ty, IntType{}) {
			foundInt = true
		}
	}
	if !foundInt {
		t.Error("expected -42 to be inferred as int")
	}
}

func TestInferUnaryNot(t *testing.T) {
	file := parse(t, `
fn main() {
	x = not true
}
`)
	info, diags := Check(file)
	if len(diags) > 0 {
		t.Errorf("unexpected diagnostics: %v", diags)
	}

	foundBool := false
	for _, ty := range info.ExprTypes {
		if TypeEquals(ty, BoolType{}) {
			foundBool = true
		}
	}
	if !foundBool {
		t.Error("expected !true to be inferred as bool")
	}
}

// ---------------------------------------------------------------------------
// checker.go — index expression inference
// ---------------------------------------------------------------------------

func TestInferListIndex(t *testing.T) {
	file := parse(t, `
fn main() {
	nums = [10, 20, 30]
	x = nums[0]
}
`)
	info, diags := Check(file)
	if len(diags) > 0 {
		t.Errorf("unexpected diagnostics: %v", diags)
	}

	// x should be int (elem type of [int])
	foundInt := false
	for _, ty := range info.VarTypes {
		if TypeEquals(ty, IntType{}) {
			foundInt = true
		}
	}
	if !foundInt {
		t.Error("expected nums[0] to be inferred as int")
	}
}

// ---------------------------------------------------------------------------
// checker.go — if condition warning
// ---------------------------------------------------------------------------

func TestIfConditionNonBoolWarning(t *testing.T) {
	file := parse(t, `
fn main() {
	if 42 {
		x = 1
	}
}
`)
	_, diags := Check(file)
	found := false
	for _, d := range diags {
		if contains(d.Message, "condition should be bool") {
			found = true
		}
	}
	if !found {
		t.Error("expected 'condition should be bool' warning")
	}
}

// ---------------------------------------------------------------------------
// checker.go — range expression
// ---------------------------------------------------------------------------

func TestInferRange(t *testing.T) {
	file := parse(t, `
fn main() {
	r = 0..10
}
`)
	info, diags := Check(file)
	if len(diags) > 0 {
		t.Errorf("unexpected diagnostics: %v", diags)
	}

	foundIntList := false
	for _, ty := range info.ExprTypes {
		if lt, ok := ty.(ListType); ok {
			if TypeEquals(lt.Elem, IntType{}) {
				foundIntList = true
			}
		}
	}
	if !foundIntList {
		t.Error("expected 0..10 to be inferred as [int]")
	}
}

// ---------------------------------------------------------------------------
// checker.go — all examples parse and check without crashing
// ---------------------------------------------------------------------------

func TestCheckAllExamples(t *testing.T) {
	examples := []string{
		`fn main() { io.println("hello") }`,
		`fn main() { x = 42 }`,
		`fn add(a: int, b: int) -> int { return a + b }
fn main() { io.println(add(1, 2)) }`,
		`fn main() {
	x = 10
	if x > 5 {
		io.println("big")
	} else {
		io.println("small")
	}
}`,
		`fn main() {
	x = 3
	match x {
		1 => io.println("one")
		2 => io.println("two")
		_ => io.println("other")
	}
}`,
		`fn main() {
	nums = [1, 2, 3]
	for n in nums {
		io.println(n)
	}
}`,
	}

	for i, src := range examples {
		file := parse(t, src)
		_, diags := Check(file)
		// Just check it doesn't panic or produce errors
		for _, d := range diags {
			if d.Level == 0 { // Error level
				t.Errorf("example %d: unexpected error: %s", i, d.Message)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// checker.go — tool and workflow registration
// ---------------------------------------------------------------------------

func TestToolRegistration(t *testing.T) {
	file := parse(t, `
tool get_weather(city: string) -> string {
	"""Get the current weather for a city"""
	return "sunny"
}

fn main() {
	w = get_weather("NYC")
}
`)
	info, diags := Check(file)
	if len(diags) > 0 {
		t.Errorf("unexpected diagnostics: %v", diags)
	}

	foundString := false
	for _, ty := range info.VarTypes {
		if TypeEquals(ty, StringType{}) {
			foundString = true
		}
	}
	if !foundString {
		t.Error("expected get_weather() return to be inferred as string")
	}
}

func TestWorkflowRegistration(t *testing.T) {
	file := parse(t, `
workflow greet(name: string) -> string {
	return "Hello, " + name
}

fn main() {
	msg = greet("Alice")
}
`)
	info, diags := Check(file)
	if len(diags) > 0 {
		t.Errorf("unexpected diagnostics: %v", diags)
	}

	foundString := false
	for _, ty := range info.VarTypes {
		if TypeEquals(ty, StringType{}) {
			foundString = true
		}
	}
	if !foundString {
		t.Error("expected workflow return to be inferred as string")
	}
}

// ---------------------------------------------------------------------------
// checker.go — mixed int/float arithmetic
// ---------------------------------------------------------------------------

func TestMixedIntFloatArithmetic(t *testing.T) {
	file := parse(t, `
fn main() {
	x = 1 + 2.5
}
`)
	info, diags := Check(file)
	if len(diags) > 0 {
		t.Errorf("unexpected diagnostics: %v", diags)
	}

	foundFloat := false
	for _, ty := range info.VarTypes {
		if TypeEquals(ty, FloatType{}) {
			foundFloat = true
		}
	}
	if !foundFloat {
		t.Error("expected int + float to produce float")
	}
}

// ---------------------------------------------------------------------------
// Phase 4A: Agent/provider field validation
// ---------------------------------------------------------------------------

func TestAgentMissingModel(t *testing.T) {
	file := parse(t, `
provider openai {
	api_key: env("OPENAI_API_KEY")
	model: "gpt-4o-mini"
}

agent Writer {
	system: "You are a writer."
}

fn main() {}
`)
	_, diags := Check(file)
	found := false
	for _, d := range diags {
		if contains(d.Message, "missing required field") && contains(d.Message, "provider") {
			found = true
		}
	}
	if !found {
		t.Error("expected 'missing required field provider' error for agent without provider")
	}
}

func TestAgentUnknownField(t *testing.T) {
	file := parse(t, `
provider openai {
	api_key: env("OPENAI_API_KEY")
	model: "gpt-4o-mini"
}

agent Writer {
	provider: openai
	system: "You are a writer."
	foo: "bar"
}

fn main() {}
`)
	_, diags := Check(file)
	found := false
	for _, d := range diags {
		if contains(d.Message, "unknown agent field") && contains(d.Message, "foo") {
			found = true
		}
	}
	if !found {
		t.Error("expected 'unknown agent field foo' warning")
	}
}

func TestProviderUnknownField(t *testing.T) {
	file := parse(t, `
provider openai {
	api_key: env("OPENAI_API_KEY")
	model: "gpt-4o-mini"
	baz: "qux"
}

fn main() {}
`)
	_, diags := Check(file)
	found := false
	for _, d := range diags {
		if contains(d.Message, "unknown provider field") && contains(d.Message, "baz") {
			found = true
		}
	}
	if !found {
		t.Error("expected 'unknown provider field baz' warning")
	}
}

func TestAgentUnknownProvider(t *testing.T) {
	file := parse(t, `
agent Writer {
	provider: nonexistent
	system: "You are a writer."
}

fn main() {}
`)
	_, diags := Check(file)
	found := false
	for _, d := range diags {
		if contains(d.Message, "unknown provider") && contains(d.Message, "nonexistent") {
			found = true
		}
	}
	if !found {
		t.Error("expected 'unknown provider' error for agent referencing nonexistent provider")
	}
}

func TestAgentUnknownTool(t *testing.T) {
	file := parse(t, `
provider openai {
	api_key: env("OPENAI_API_KEY")
	model: "gpt-4o-mini"
}

agent Assistant {
	provider: openai
	tools: [nonexistent_tool]
}

fn main() {}
`)
	_, diags := Check(file)
	found := false
	for _, d := range diags {
		if contains(d.Message, "unknown tool") && contains(d.Message, "nonexistent_tool") {
			found = true
		}
	}
	if !found {
		t.Error("expected 'unknown tool' error for agent referencing nonexistent tool")
	}
}

func TestAgentValidDecl(t *testing.T) {
	file := parse(t, `
provider openai {
	api_key: env("OPENAI_API_KEY")
	model: "gpt-4o-mini"
}

tool get_time() -> string {
	"""Get the time"""
	return "now"
}

agent Assistant {
	provider: openai
	system: "You are helpful."
	tools: [get_time]
	temperature: 0.7
	max_tokens: 1000
}

fn main() {}
`)
	_, diags := Check(file)
	for _, d := range diags {
		if d.Level == 0 { // Error level
			t.Errorf("unexpected error in valid agent decl: %s", d.Message)
		}
	}
}

// ---------------------------------------------------------------------------
// Phase 4B: Undefined variable/function diagnostics
// ---------------------------------------------------------------------------

func TestUndefinedVariable(t *testing.T) {
	file := parse(t, `
fn main() {
	x = y + 1
}
`)
	_, diags := Check(file)
	found := false
	for _, d := range diags {
		if contains(d.Message, "undefined variable") && contains(d.Message, "'y'") {
			found = true
		}
	}
	if !found {
		t.Error("expected 'undefined variable y' warning")
	}
}

func TestUndefinedFunction(t *testing.T) {
	file := parse(t, `
fn main() {
	x = nonexistent(42)
}
`)
	_, diags := Check(file)
	found := false
	for _, d := range diags {
		if contains(d.Message, "undefined function") && contains(d.Message, "nonexistent") {
			found = true
		}
	}
	if !found {
		t.Error("expected 'undefined function nonexistent' warning")
	}
}

func TestDefinedVariableNoWarning(t *testing.T) {
	file := parse(t, `
fn main() {
	x = 42
	y = x + 1
}
`)
	_, diags := Check(file)
	for _, d := range diags {
		if contains(d.Message, "undefined") {
			t.Errorf("unexpected undefined warning: %s", d.Message)
		}
	}
}

func TestStdlibCallNoWarning(t *testing.T) {
	file := parse(t, `
import "io"
import "string"

fn main() {
	x = string.len("hello")
	io.println(x)
}
`)
	_, diags := Check(file)
	for _, d := range diags {
		if contains(d.Message, "undefined") {
			t.Errorf("unexpected undefined warning for stdlib: %s", d.Message)
		}
	}
}

// ---------------------------------------------------------------------------
// Phase 4C: Return type checking
// ---------------------------------------------------------------------------

func TestReturnTypeMismatch(t *testing.T) {
	file := parse(t, `
fn add(a: int, b: int) -> int {
	return "hello"
}

fn main() {}
`)
	_, diags := Check(file)
	found := false
	for _, d := range diags {
		if contains(d.Message, "return type mismatch") {
			found = true
		}
	}
	if !found {
		t.Error("expected 'return type mismatch' warning for returning string from int function")
	}
}

func TestReturnTypeMatch(t *testing.T) {
	file := parse(t, `
fn add(a: int, b: int) -> int {
	return a + b
}

fn main() {}
`)
	_, diags := Check(file)
	for _, d := range diags {
		if contains(d.Message, "return type mismatch") {
			t.Errorf("unexpected return type mismatch: %s", d.Message)
		}
	}
}

func TestToolReturnTypeCheck(t *testing.T) {
	file := parse(t, `
tool get_name() -> string {
	"""Get the name"""
	return 42
}

fn main() {}
`)
	_, diags := Check(file)
	found := false
	for _, d := range diags {
		if contains(d.Message, "return type mismatch") {
			found = true
		}
	}
	if !found {
		t.Error("expected 'return type mismatch' warning for returning int from string tool")
	}
}

func TestPropagateOutsideTryWarning(t *testing.T) {
	file := parse(t, `
fn risky() -> string {
	return "ok"
}

fn main() {
	x = risky()?
}
`)
	_, diags := Check(file)
	found := false
	for _, d := range diags {
		if contains(d.Message, "outside a try block") {
			found = true
		}
	}
	if !found {
		t.Error("expected warning about '?' used outside a try block")
	}
}

func TestPropagateInsideTryNoWarning(t *testing.T) {
	file := parse(t, `
fn risky() -> string {
	return "ok"
}

fn main() {
	try {
		x = risky()?
	} catch err {
		io.println(err)
	}
}
`)
	_, diags := Check(file)
	for _, d := range diags {
		if contains(d.Message, "outside a try block") {
			t.Error("should NOT warn about '?' inside a try block")
		}
	}
}

func TestAgentTimeoutField(t *testing.T) {
	file := parse(t, `
provider openai {
	api_key: env("OPENAI_API_KEY")
	model: "gpt-4o-mini"
}

agent Writer {
	provider: openai
	system: "You are a writer."
	timeout: 30
}

fn main() {}
`)
	_, diags := Check(file)
	for _, d := range diags {
		if contains(d.Message, "unknown agent field") && contains(d.Message, "timeout") {
			t.Error("timeout should be a valid agent field, but got unknown field warning")
		}
	}
}

func TestAgentHandoffUnknownTarget(t *testing.T) {
	file := parse(t, `
provider openai {
	api_key: env("OPENAI_API_KEY")
	model: "gpt-4o-mini"
}

agent Router {
	provider: openai
	system: "You route."
	handoffs: [NonExistent]
}

fn main() {}
`)
	_, diags := Check(file)
	found := false
	for _, d := range diags {
		if contains(d.Message, "handoff target") && contains(d.Message, "NonExistent") {
			found = true
		}
	}
	if !found {
		t.Error("expected error about unknown handoff target 'NonExistent'")
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
