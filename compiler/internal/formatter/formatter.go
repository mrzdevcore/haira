// Package formatter implements canonical source formatting for Haira files.
// It walks the parsed AST to emit canonically formatted output while using
// the raw token stream to preserve and re-attach comments.
package formatter

import (
	"strings"

	"github.com/haira-lang/haira/internal/ast"
	"github.com/haira-lang/haira/internal/token"
)

// Formatter holds the state for formatting a single source file.
type Formatter struct {
	src        string
	tokens     []token.Token
	indent     int
	out        strings.Builder
	nextTokIdx int // cursor into tokens for comment scanning
}

// Format formats a Haira source file and returns the canonically formatted string.
func Format(src string, tokens []token.Token, file *ast.SourceFile) string {
	f := &Formatter{
		src:    src,
		tokens: tokens,
	}
	f.formatSourceFile(file)
	return f.out.String()
}

// --- Output helpers ---

func (f *Formatter) write(s string) {
	f.out.WriteString(s)
}

func (f *Formatter) writeln(s string) {
	f.out.WriteString(s)
	f.out.WriteByte('\n')
}

func (f *Formatter) newline() {
	f.out.WriteByte('\n')
}

// blank emits a blank line (two newlines total).
func (f *Formatter) blank() {
	f.out.WriteByte('\n')
}

func (f *Formatter) writeIndent() {
	for i := 0; i < f.indent; i++ {
		f.out.WriteString("    ")
	}
}

func (f *Formatter) incIndent() { f.indent++ }
func (f *Formatter) decIndent() { f.indent-- }

// --- Comment handling ---

// emitCommentsBefore scans the token stream for comments that appear before
// the given byte offset. It advances nextTokIdx past any consumed trivia.
func (f *Formatter) emitCommentsBefore(offset int) {
	for f.nextTokIdx < len(f.tokens) {
		tok := f.tokens[f.nextTokIdx]
		if tok.Start >= offset {
			break
		}
		if tok.Kind == token.LineComment {
			f.writeIndent()
			f.writeln(tok.Value)
			f.nextTokIdx++
		} else if tok.Kind == token.BlockComment {
			f.writeIndent()
			f.writeln(tok.Value)
			f.nextTokIdx++
		} else {
			f.nextTokIdx++
		}
	}
}

// emitTrailingComment checks if there's a line comment on the same line
// immediately after the given offset (before the next newline). If found,
// it emits " // comment" and advances the cursor.
func (f *Formatter) emitTrailingComment(offset int) {
	for f.nextTokIdx < len(f.tokens) {
		tok := f.tokens[f.nextTokIdx]
		if tok.Start < offset {
			f.nextTokIdx++
			continue
		}
		if tok.Kind == token.Newline {
			break
		}
		if tok.Kind == token.LineComment {
			f.write(" ")
			f.write(tok.Value)
			f.nextTokIdx++
			return
		}
		// Skip other non-trivia tokens that are part of the current line
		if tok.Kind != token.LineComment && tok.Kind != token.BlockComment && tok.Kind != token.Newline {
			break
		}
		f.nextTokIdx++
	}
}

// skipTokensPast advances the token cursor past the given byte offset.
func (f *Formatter) skipTokensPast(offset int) {
	for f.nextTokIdx < len(f.tokens) && f.tokens[f.nextTokIdx].Start < offset {
		f.nextTokIdx++
	}
}

// emitRemainingComments emits any comments left after the last AST item.
func (f *Formatter) emitRemainingComments() {
	for f.nextTokIdx < len(f.tokens) {
		tok := f.tokens[f.nextTokIdx]
		if tok.Kind == token.LineComment {
			f.writeIndent()
			f.writeln(tok.Value)
		} else if tok.Kind == token.BlockComment {
			f.writeIndent()
			f.writeln(tok.Value)
		}
		f.nextTokIdx++
	}
}
