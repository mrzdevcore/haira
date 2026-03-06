package formatter

import (
	"strings"
	"testing"

	"github.com/haira-lang/haira/internal/lexer"
	"github.com/haira-lang/haira/internal/parser"
)

func format(t *testing.T, src string) string {
	t.Helper()
	l := lexer.New(src)
	tokens := l.AllTokens()
	sf, errs := parser.Parse(src)
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	return Format(src, tokens, sf)
}

// ---------------------------------------------------------------------------
// Basic formatting
// ---------------------------------------------------------------------------

func TestFormatSimpleFunction(t *testing.T) {
	src := `fn main() {
    io.println("hello")
}`
	out := format(t, src)
	if !strings.Contains(out, "fn main()") {
		t.Error("formatted output should contain fn main()")
	}
	if !strings.Contains(out, "io.println") {
		t.Error("formatted output should contain io.println")
	}
}

func TestFormatImport(t *testing.T) {
	src := `import "io"

fn main() {
    io.println("hello")
}`
	out := format(t, src)
	if !strings.Contains(out, `import "io"`) {
		t.Errorf("formatted output should contain import, got:\n%s", out)
	}
}

func TestFormatStruct(t *testing.T) {
	src := `struct User {
    name: string
    age: int
}`
	out := format(t, src)
	if !strings.Contains(out, "struct User") {
		t.Error("formatted output should contain struct User")
	}
	if !strings.Contains(out, "name: string") {
		t.Error("formatted output should contain name: string")
	}
}

func TestFormatEnum(t *testing.T) {
	src := `enum Color {
    Red
    Green
    Blue
}`
	out := format(t, src)
	if !strings.Contains(out, "enum Color") {
		t.Error("formatted output should contain enum Color")
	}
}

func TestFormatProvider(t *testing.T) {
	src := `provider openai {
    api_key: env("KEY")
    model: "gpt-4"
}`
	out := format(t, src)
	if !strings.Contains(out, "provider openai") {
		t.Error("formatted output should contain provider openai")
	}
}

func TestFormatAgent(t *testing.T) {
	src := `agent Bot {
    provider: openai
    system: "You are helpful."
}`
	out := format(t, src)
	if !strings.Contains(out, "agent Bot") {
		t.Error("formatted output should contain agent Bot")
	}
}

func TestFormatNilKeyword(t *testing.T) {
	src := `fn main() {
    x = nil
}`
	out := format(t, src)
	if !strings.Contains(out, "nil") {
		t.Error("formatted output should contain nil")
	}
}

// ---------------------------------------------------------------------------
// Idempotency
// ---------------------------------------------------------------------------

func TestFormatIdempotent(t *testing.T) {
	src := `fn add(a: int, b: int) -> int {
    return a + b
}

fn main() {
    x = add(1, 2)
    io.println(x)
}`
	first := format(t, src)
	second := format(t, first)
	if first != second {
		t.Errorf("formatting is not idempotent:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

// ---------------------------------------------------------------------------
// Empty source
// ---------------------------------------------------------------------------

func TestFormatEmpty(t *testing.T) {
	out := format(t, "")
	if strings.TrimSpace(out) != "" {
		t.Errorf("formatting empty source should produce empty output, got %q", out)
	}
}
