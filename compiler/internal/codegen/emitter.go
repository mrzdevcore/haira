// Package codegen generates Go source code from Haira AST.
package codegen

import (
	"fmt"
	"strings"

	"github.com/haira-lang/haira/internal/ast"
	"github.com/haira-lang/haira/internal/checker"
)

// GoEmitter writes Go source code with proper indentation.
type GoEmitter struct {
	buf          strings.Builder
	indent       int
	declaredVars map[string]bool
	typeInfo     *checker.TypeInfo
	sourceFile   string // Haira source filename (e.g., "examples/08-structs.haira")
	sourceText   string // Haira source text (for offset→line conversion)
	inTryBlock   bool   // true when emitting inside a try body (? uses panic instead of return)
}

// LookupExprType returns the inferred checker type for an expression span, or nil.
func (e *GoEmitter) LookupExprType(span ast.Span) checker.Type {
	if e.typeInfo == nil {
		return nil
	}
	return e.typeInfo.ExprTypes[span]
}

// NewEmitter creates a new GoEmitter.
func NewEmitter() *GoEmitter {
	return &GoEmitter{
		declaredVars: make(map[string]bool),
	}
}

// DeclareVar marks a variable as declared. Returns true if this is a NEW declaration (use :=).
func (e *GoEmitter) DeclareVar(name string) bool {
	if e.declaredVars[name] {
		return false
	}
	e.declaredVars[name] = true
	return true
}

// ResetVars clears declared variables — call at the start of each function/workflow.
func (e *GoEmitter) ResetVars() {
	e.declaredVars = make(map[string]bool)
}

// Line writes a line with current indentation.
func (e *GoEmitter) Line(s string) {
	if s == "" {
		e.buf.WriteByte('\n')
		return
	}
	for i := 0; i < e.indent; i++ {
		e.buf.WriteByte('\t')
	}
	e.buf.WriteString(s)
	e.buf.WriteByte('\n')
}

// Linef writes a formatted line.
func (e *GoEmitter) Linef(format string, args ...any) {
	e.Line(fmt.Sprintf(format, args...))
}

// Indent increases indentation.
func (e *GoEmitter) Indent() {
	e.indent++
}

// Dedent decreases indentation.
func (e *GoEmitter) Dedent() {
	if e.indent > 0 {
		e.indent--
	}
}

// OpenBlock writes "prefix {" and indents.
func (e *GoEmitter) OpenBlock(prefix string) {
	e.Line(prefix + " {")
	e.Indent()
}

// CloseBlock dedents and writes "}".
func (e *GoEmitter) CloseBlock() {
	e.Dedent()
	e.Line("}")
}

// Blank writes an empty line.
func (e *GoEmitter) Blank() {
	e.buf.WriteByte('\n')
}

// String returns the generated source code.
func (e *GoEmitter) String() string {
	return e.buf.String()
}

// LineDirective emits a //line directive mapping back to the Haira source location.
func (e *GoEmitter) LineDirective(span ast.Span) {
	if e.sourceFile == "" || e.sourceText == "" {
		return
	}
	line, _ := offsetToLineCol(e.sourceText, span.Start)
	e.buf.WriteString(fmt.Sprintf("//line %s:%d\n", e.sourceFile, line))
}

// offsetToLineCol converts a byte offset to 1-based line and column numbers.
func offsetToLineCol(source string, offset int) (int, int) {
	if offset < 0 {
		offset = 0
	}
	if offset > len(source) {
		offset = len(source)
	}
	line := 1
	col := 1
	for i := 0; i < offset; i++ {
		if source[i] == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	return line, col
}
