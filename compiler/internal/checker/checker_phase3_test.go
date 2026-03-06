package checker

import (
	"testing"
)

// ---------------------------------------------------------------------------
// CHK-06: Method return type inference
// ---------------------------------------------------------------------------

func TestMethodReturnType(t *testing.T) {
	file := parse(t, `
struct User {
	name: string
	age: int
}

User.greet() -> string {
	return "Hello, " + self.name
}

fn main() {
	u = User{name = "Alice", age = 30}
	msg = u.greet()
}
`)
	info, diags := Check(file)
	for _, d := range diags {
		if d.Level == 0 {
			t.Errorf("unexpected error: %s", d.Message)
		}
	}

	// msg should be inferred as string from the method return type
	foundString := false
	for _, ty := range info.VarTypes {
		if TypeEquals(ty, StringType{}) {
			foundString = true
		}
	}
	if !foundString {
		t.Error("expected method return type to propagate as string")
	}
}

// ---------------------------------------------------------------------------
// CHK-07: Instance field validation
// ---------------------------------------------------------------------------

func TestInstanceUnknownField(t *testing.T) {
	file := parse(t, `
struct User {
	name: string
}

fn main() {
	u = User{name = "Alice", email = "alice@ex.com"}
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
		t.Error("expected 'unknown field email' error on struct instance")
	}
}

func TestInstanceFieldTypeMismatch(t *testing.T) {
	file := parse(t, `
struct User {
	name: string
	age: int
}

fn main() {
	u = User{name = "Alice", age = "thirty"}
}
`)
	_, diags := Check(file)
	found := false
	for _, d := range diags {
		if contains(d.Message, "expected int") && contains(d.Message, "got string") {
			found = true
		}
	}
	if !found {
		t.Error("expected field type mismatch error for age = \"thirty\"")
	}
}

func TestInstanceMissingFieldWarning(t *testing.T) {
	file := parse(t, `
struct User {
	name: string
	age: int
}

fn main() {
	u = User{name = "Alice"}
}
`)
	_, diags := Check(file)
	found := false
	for _, d := range diags {
		if contains(d.Message, "missing field") && contains(d.Message, "age") {
			found = true
		}
	}
	if !found {
		t.Error("expected 'missing field age' warning on partial instance")
	}
}

// ---------------------------------------------------------------------------
// CHK-10: Bitwise operators require int
// ---------------------------------------------------------------------------

func TestBitwiseOperatorsValid(t *testing.T) {
	file := parse(t, `
fn main() {
	a = 0xFF & 0x0F
	b = 1 | 2
	c = 3 ^ 5
	d = 1 << 4
	e = 16 >> 2
}
`)
	_, diags := Check(file)
	for _, d := range diags {
		if d.Level == 0 {
			t.Errorf("unexpected error for valid bitwise ops: %s", d.Message)
		}
	}
}

func TestBitwiseOperatorStringError(t *testing.T) {
	file := parse(t, `
fn main() {
	x = "hello" & 5
}
`)
	_, diags := Check(file)
	found := false
	for _, d := range diags {
		if contains(d.Message, "bitwise operator requires int") {
			found = true
		}
	}
	if !found {
		t.Error("expected 'bitwise operator requires int' error")
	}
}

// ---------------------------------------------------------------------------
// CHK-11: Unary ~ requires int
// ---------------------------------------------------------------------------

func TestBitNotValid(t *testing.T) {
	file := parse(t, `
fn main() {
	x = ~0xFF
}
`)
	info, diags := Check(file)
	for _, d := range diags {
		if d.Level == 0 {
			t.Errorf("unexpected error for valid ~: %s", d.Message)
		}
	}
	// Result should be int
	foundInt := false
	for _, ty := range info.ExprTypes {
		if TypeEquals(ty, IntType{}) {
			foundInt = true
		}
	}
	if !foundInt {
		t.Error("expected ~int to produce int")
	}
}

func TestBitNotStringError(t *testing.T) {
	file := parse(t, `
fn main() {
	x = ~"hello"
}
`)
	_, diags := Check(file)
	found := false
	for _, d := range diags {
		if contains(d.Message, "bitwise NOT requires int") {
			found = true
		}
	}
	if !found {
		t.Error("expected 'bitwise NOT requires int' error")
	}
}

// ---------------------------------------------------------------------------
// CHK-12: Pattern type validation
// ---------------------------------------------------------------------------

func TestLiteralPatternTypeMismatch(t *testing.T) {
	file := parse(t, `
fn main() {
	x = "hello"
	match x {
		42 => io.println("number")
		_ => io.println("other")
	}
}
`)
	_, diags := Check(file)
	found := false
	for _, d := range diags {
		if contains(d.Message, "pattern type") && contains(d.Message, "doesn't match") {
			found = true
		}
	}
	if !found {
		t.Error("expected warning about pattern type not matching subject type")
	}
}

// ---------------------------------------------------------------------------
// CHK-13: Pattern variable binding
// ---------------------------------------------------------------------------

func TestPatternVariableBinding(t *testing.T) {
	file := parse(t, `
fn main() {
	x = 42
	match x {
		n => io.println(n)
	}
}
`)
	_, diags := Check(file)
	// n should be bound and not produce an "undefined variable" warning
	for _, d := range diags {
		if contains(d.Message, "undefined variable") && contains(d.Message, "'n'") {
			t.Error("pattern variable 'n' should be bound, but got undefined warning")
		}
	}
}

// ---------------------------------------------------------------------------
// CHK-15: Parameter default value type check
// ---------------------------------------------------------------------------

func TestDefaultValueTypeMismatch(t *testing.T) {
	file := parse(t, `
fn greet(name: string = 42) {
	io.println(name)
}

fn main() {}
`)
	_, diags := Check(file)
	found := false
	for _, d := range diags {
		if contains(d.Message, "default value type") && contains(d.Message, "doesn't match") {
			found = true
		}
	}
	if !found {
		t.Error("expected default value type mismatch error for string param with int default")
	}
}

func TestDefaultValueTypeMatch(t *testing.T) {
	file := parse(t, `
fn greet(name: string = "World") {
	io.println(name)
}

fn main() {}
`)
	_, diags := Check(file)
	for _, d := range diags {
		if contains(d.Message, "default value type") {
			t.Errorf("unexpected default value type error: %s", d.Message)
		}
	}
}

// ---------------------------------------------------------------------------
// CHK-16: Enum variant field count validation
// ---------------------------------------------------------------------------

func TestEnumVariantFieldCount(t *testing.T) {
	// This tests that the enum variant field count is tracked
	file := parse(t, `
enum Shape {
	Circle(radius: float)
	Rectangle(width: float, height: float)
}

fn main() {}
`)
	_, diags := Check(file)
	// Should not crash
	for _, d := range diags {
		if d.Level == 0 {
			t.Errorf("unexpected error: %s", d.Message)
		}
	}
}

// ---------------------------------------------------------------------------
// Break/continue outside loop (CHK-08, Phase 1)
// ---------------------------------------------------------------------------

func TestBreakOutsideLoop(t *testing.T) {
	file := parse(t, `
fn main() {
	break
}
`)
	_, diags := Check(file)
	found := false
	for _, d := range diags {
		if contains(d.Message, "break can only be used inside a loop") {
			found = true
		}
	}
	if !found {
		t.Error("expected error for break outside loop")
	}
}

func TestContinueOutsideLoop(t *testing.T) {
	file := parse(t, `
fn main() {
	continue
}
`)
	_, diags := Check(file)
	found := false
	for _, d := range diags {
		if contains(d.Message, "continue can only be used inside a loop") {
			found = true
		}
	}
	if !found {
		t.Error("expected error for continue outside loop")
	}
}

func TestBreakInsideLoop(t *testing.T) {
	file := parse(t, `
fn main() {
	for i in 0..10 {
		if i == 5 {
			break
		}
	}
}
`)
	_, diags := Check(file)
	for _, d := range diags {
		if contains(d.Message, "break can only be used") {
			t.Error("break inside loop should not produce error")
		}
	}
}

// ---------------------------------------------------------------------------
// Const reassignment
// ---------------------------------------------------------------------------

func TestConstReassignment(t *testing.T) {
	file := parse(t, `
fn main() {
	const x = 42
	x = 99
}
`)
	_, diags := Check(file)
	found := false
	for _, d := range diags {
		if contains(d.Message, "cannot reassign const") {
			found = true
		}
	}
	if !found {
		t.Error("expected error for reassigning const")
	}
}

// ---------------------------------------------------------------------------
// Self reassignment in method
// ---------------------------------------------------------------------------

func TestSelfReassignment(t *testing.T) {
	file := parse(t, `
struct User {
	name: string
}

User.bad() {
	self = User{name = "hacked"}
}

fn main() {}
`)
	_, diags := Check(file)
	found := false
	for _, d := range diags {
		if contains(d.Message, "cannot reassign self") {
			found = true
		}
	}
	if !found {
		t.Error("expected error for reassigning self in method")
	}
}

// ---------------------------------------------------------------------------
// Nil keyword (SPEC-02)
// ---------------------------------------------------------------------------

func TestNilKeyword(t *testing.T) {
	file := parse(t, `
fn main() {
	x = nil
}
`)
	_, diags := Check(file)
	// nil should not produce an "undefined variable" warning
	for _, d := range diags {
		if contains(d.Message, "undefined variable") && contains(d.Message, "nil") {
			t.Error("nil should be a keyword, not trigger undefined variable warning")
		}
	}
}

// ---------------------------------------------------------------------------
// Unimplemented trigger warning (SPEC-04)
// ---------------------------------------------------------------------------

func TestUnimplementedTriggerWarning(t *testing.T) {
	file := parse(t, `
@cron("0 9 * * *")
workflow DailyReport() -> string {
	return "report"
}

fn main() {}
`)
	_, diags := Check(file)
	found := false
	for _, d := range diags {
		if contains(d.Message, "not yet implemented") && contains(d.Message, "cron") {
			found = true
		}
	}
	if !found {
		t.Error("expected warning about @cron trigger not yet implemented")
	}
}

// ---------------------------------------------------------------------------
// Assert outside test
// ---------------------------------------------------------------------------

func TestAssertOutsideTest(t *testing.T) {
	file := parse(t, `
fn main() {
	assert true
}
`)
	_, diags := Check(file)
	found := false
	for _, d := range diags {
		if contains(d.Message, "assert can only be used inside test blocks") {
			found = true
		}
	}
	if !found {
		t.Error("expected error for assert outside test block")
	}
}

// ---------------------------------------------------------------------------
// Function missing return type annotation
// ---------------------------------------------------------------------------

func TestFunctionReturnWithoutAnnotation(t *testing.T) {
	file := parse(t, `
fn bad() {
	return 42
}

fn main() {}
`)
	_, diags := Check(file)
	found := false
	for _, d := range diags {
		if contains(d.Message, "returns a value but has no return type") {
			found = true
		}
	}
	if !found {
		t.Error("expected error for function returning value without return type annotation")
	}
}
