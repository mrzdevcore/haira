package parser

import (
	"testing"
)

// ---------------------------------------------------------------------------
// Parser negative tests — invalid syntax should produce parse errors
// ---------------------------------------------------------------------------

func TestMissingClosingBrace(t *testing.T) {
	src := `fn main() {`
	_, errs := Parse(src)
	if len(errs) == 0 {
		t.Error("expected parse errors for missing closing brace")
	}
}

func TestMissingClosingParen(t *testing.T) {
	src := `fn main( { }`
	_, errs := Parse(src)
	if len(errs) == 0 {
		t.Error("expected parse errors for missing closing paren")
	}
}

func TestMissingFunctionBody(t *testing.T) {
	src := `fn main()`
	_, errs := Parse(src)
	if len(errs) == 0 {
		t.Error("expected parse errors for missing function body")
	}
}

func TestInvalidTopLevel(t *testing.T) {
	src := `+ + +`
	_, errs := Parse(src)
	if len(errs) == 0 {
		t.Error("expected parse errors for invalid top-level tokens")
	}
}

func TestUnclosedStringLiteral(t *testing.T) {
	src := `fn main() { x = "hello }`
	_, errs := Parse(src)
	if len(errs) == 0 {
		t.Error("expected parse errors for unclosed string literal")
	}
}

func TestMissingArrowInFunction(t *testing.T) {
	src := `fn add(a: int, b: int) int { return a + b }`
	_, errs := Parse(src)
	if len(errs) == 0 {
		t.Error("expected parse errors for missing -> in return type")
	}
}

func TestEmptyMatchArms(t *testing.T) {
	src := `fn main() { match x { } }`
	// Empty match should parse without error (valid syntax, empty arms)
	_, errs := Parse(src)
	if len(errs) > 0 {
		t.Errorf("empty match should parse cleanly, got %d errors", len(errs))
	}
}

func TestMissingMatchArrow(t *testing.T) {
	src := `fn main() {
    match x {
        1 "one"
    }
}`
	_, errs := Parse(src)
	if len(errs) == 0 {
		t.Error("expected parse errors for missing => in match arm")
	}
}

func TestToolMissingDescription(t *testing.T) {
	src := `tool search(query: string) -> string {
    return "result"
}`
	// Tool without triple-quote description — should still parse but is a semantic error
	_, errs := Parse(src)
	// Parser may or may not error here; the checker enforces this
	_ = errs
}

func TestStructMissingFieldType(t *testing.T) {
	// Fields without type annotation are valid (type inference).
	// But a field with a colon and no type IS an error.
	src := `struct User { name: }`
	_, errs := Parse(src)
	if len(errs) == 0 {
		t.Error("expected parse errors for struct field with colon but no type")
	}
}

func TestDuplicateElse(t *testing.T) {
	src := `fn main() {
    if true { x = 1 } else { x = 2 } else { x = 3 }
}`
	_, errs := Parse(src)
	if len(errs) == 0 {
		t.Error("expected parse errors for duplicate else clause")
	}
}

func TestInvalidOperator(t *testing.T) {
	src := `fn main() { x = 1 ++ 2 }`
	_, errs := Parse(src)
	// ++ is not a valid operator in Haira
	if len(errs) == 0 {
		t.Error("expected parse errors for ++ operator")
	}
}

func TestMissingForIn(t *testing.T) {
	src := `fn main() { for x { io.println(x) } }`
	// Missing 'in' keyword — should be "for x in ..."
	_, errs := Parse(src)
	if len(errs) == 0 {
		t.Error("expected parse errors for for-loop missing 'in'")
	}
}

func TestWorkflowMissingBody(t *testing.T) {
	src := `@webhook("/test")
workflow Test(x: string) -> string`
	_, errs := Parse(src)
	if len(errs) == 0 {
		t.Error("expected parse errors for workflow missing body")
	}
}

func TestAgentMissingBraces(t *testing.T) {
	src := `agent Bot
    provider: openai`
	_, errs := Parse(src)
	if len(errs) == 0 {
		t.Error("expected parse errors for agent missing braces")
	}
}
