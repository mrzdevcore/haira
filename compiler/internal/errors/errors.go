// Package errors provides shared diagnostic types for the Haira compiler.
package errors

import (
	"fmt"
	"strings"

	"github.com/haira-lang/haira/internal/ast"
)

// Level represents the severity of a diagnostic.
type Level int

const (
	Error Level = iota
	Warning
	Info
)

func (l Level) String() string {
	switch l {
	case Error:
		return "error"
	case Warning:
		return "warning"
	case Info:
		return "info"
	}
	return "unknown"
}

// Diagnostic represents a compiler message with location information.
type Diagnostic struct {
	Level   Level
	Message string
	Span    ast.Span
	File    string
	Hint    string // optional suggestion
}

func (d Diagnostic) String() string {
	s := fmt.Sprintf("%s: %s", d.Level, d.Message)
	if d.Hint != "" {
		s += fmt.Sprintf(" (hint: %s)", d.Hint)
	}
	return s
}

// PrettyPrint renders a diagnostic with source context showing the exact location.
//
//	examples/03-functions.haira:5:12: error: type mismatch
//	  |
//	5 |     return "hello" + 42
//	  |            ^^^^^^^^^^^^^^ cannot add string and int
func PrettyPrint(d Diagnostic, source string) string {
	var b strings.Builder

	line, col := offsetToLineCol(source, d.Span.Start)
	endLine, endCol := offsetToLineCol(source, d.Span.End)

	// Header: file:line:col: level: message
	if d.File != "" {
		fmt.Fprintf(&b, "%s:%d:%d: %s: %s\n", d.File, line, col, d.Level, d.Message)
	} else {
		fmt.Fprintf(&b, "%d:%d: %s: %s\n", line, col, d.Level, d.Message)
	}

	// Source context
	lines := strings.Split(source, "\n")
	if line <= len(lines) {
		lineText := lines[line-1]
		lineNumStr := fmt.Sprintf("%d", line)
		padding := strings.Repeat(" ", len(lineNumStr))

		fmt.Fprintf(&b, "%s |\n", padding)
		fmt.Fprintf(&b, "%s | %s\n", lineNumStr, lineText)

		// Underline
		underStart := col - 1
		if underStart < 0 {
			underStart = 0
		}
		underLen := d.Span.End - d.Span.Start
		if endLine > line {
			// Multi-line span: underline to end of first line
			underLen = len(lineText) - underStart
		} else {
			underLen = endCol - col
		}
		if underLen <= 0 {
			underLen = 1
		}

		underline := strings.Repeat(" ", underStart) + strings.Repeat("^", underLen)
		label := d.Message
		if d.Hint != "" {
			label = d.Hint
		}
		fmt.Fprintf(&b, "%s | %s %s\n", padding, underline, label)
	}

	return b.String()
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

// HasErrors returns true if any diagnostic is at Error level.
func HasErrors(diags []Diagnostic) bool {
	for _, d := range diags {
		if d.Level == Error {
			return true
		}
	}
	return false
}

// FormatAll pretty-prints all diagnostics for a given source.
func FormatAll(diags []Diagnostic, source string) string {
	var b strings.Builder
	for i, d := range diags {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(PrettyPrint(d, source))
	}
	return b.String()
}
