// Package lexer implements a hand-written scanner for Haira source code.
package lexer

import (
	"strings"
	"unicode"

	"github.com/haira-lang/haira/internal/token"
)

// Lexer tokenizes Haira source code.
type Lexer struct {
	source  string
	pos     int // current byte position
	tokens  []token.Token
	current int // index into tokens for iteration
}

// New creates a new lexer and immediately tokenizes the entire source.
func New(source string) *Lexer {
	l := &Lexer{source: source}
	l.scan()
	return l
}

// Next returns the next non-trivia, non-newline token, advancing past trivia.
func (l *Lexer) Next() token.Token {
	for l.current < len(l.tokens) {
		tok := l.tokens[l.current]
		l.current++
		if tok.Kind.IsTrivia() {
			continue
		}
		return tok
	}
	pos := len(l.source)
	return token.Token{Kind: token.EOF, Start: pos, End: pos}
}

// NextWithNewlines returns the next non-trivia token, including newlines.
func (l *Lexer) NextWithNewlines() token.Token {
	for l.current < len(l.tokens) {
		tok := l.tokens[l.current]
		l.current++
		if tok.Kind.IsTrivia() {
			continue
		}
		return tok
	}
	pos := len(l.source)
	return token.Token{Kind: token.EOF, Start: pos, End: pos}
}

// AllTokens returns all tokens (including trivia/newlines) for the lex command.
func (l *Lexer) AllTokens() []token.Token {
	return l.tokens
}

// scan tokenizes the entire source into l.tokens.
func (l *Lexer) scan() {
	for l.pos < len(l.source) {
		l.skipSpacesAndTabs()
		if l.pos >= len(l.source) {
			break
		}

		start := l.pos
		ch := l.source[l.pos]

		switch {
		case ch == '\n' || ch == '\r':
			l.scanNewline(start)
		case ch == '/' && l.pos+1 < len(l.source) && l.source[l.pos+1] == '/':
			l.scanLineComment(start)
		case ch == '/' && l.pos+1 < len(l.source) && l.source[l.pos+1] == '*':
			l.scanBlockComment(start)
		case ch == '"':
			l.scanString(start)
		case isDigit(ch):
			l.scanNumber(start)
		case isIdentStart(ch):
			l.scanIdentOrKeyword(start)
		default:
			l.scanOperatorOrDelimiter(start)
		}
	}

	// Append EOF
	l.tokens = append(l.tokens, token.Token{Kind: token.EOF, Start: l.pos, End: l.pos})
}

func (l *Lexer) skipSpacesAndTabs() {
	for l.pos < len(l.source) {
		ch := l.source[l.pos]
		if ch == ' ' || ch == '\t' {
			l.pos++
		} else {
			break
		}
	}
}

func (l *Lexer) scanNewline(start int) {
	if l.source[l.pos] == '\r' && l.pos+1 < len(l.source) && l.source[l.pos+1] == '\n' {
		l.pos += 2
	} else {
		l.pos++
	}
	l.tokens = append(l.tokens, token.Token{Kind: token.Newline, Start: start, End: l.pos})
}

func (l *Lexer) scanLineComment(start int) {
	l.pos += 2 // skip //
	for l.pos < len(l.source) && l.source[l.pos] != '\n' {
		l.pos++
	}
	l.tokens = append(l.tokens, token.Token{Kind: token.LineComment, Value: l.source[start:l.pos], Start: start, End: l.pos})
}

func (l *Lexer) scanBlockComment(start int) {
	l.pos += 2 // skip /*
	depth := 1
	for l.pos < len(l.source) && depth > 0 {
		if l.pos+1 < len(l.source) {
			if l.source[l.pos] == '*' && l.source[l.pos+1] == '/' {
				depth--
				l.pos += 2
				continue
			}
			if l.source[l.pos] == '/' && l.source[l.pos+1] == '*' {
				depth++
				l.pos += 2
				continue
			}
		}
		l.pos++
	}
	l.tokens = append(l.tokens, token.Token{Kind: token.BlockComment, Value: l.source[start:l.pos], Start: start, End: l.pos})
}

func (l *Lexer) scanString(start int) {
	// Check for triple-quoted string: """..."""
	if l.pos+2 < len(l.source) && l.source[l.pos:l.pos+3] == `"""` {
		l.scanTripleQuoteString(start)
		return
	}

	l.pos++ // skip opening "
	var hasInterpolation bool
	var buf strings.Builder
	interpDepth := 0 // track brace depth inside ${...}

	for l.pos < len(l.source) {
		ch := l.source[l.pos]

		// Only treat " as string terminator when NOT inside ${...}
		if ch == '"' && interpDepth == 0 {
			break
		}

		if ch == '\\' && interpDepth == 0 {
			l.pos++
			if l.pos < len(l.source) {
				esc := l.source[l.pos]
				switch esc {
				case 'n':
					buf.WriteByte('\n')
				case 't':
					buf.WriteByte('\t')
				case 'r':
					buf.WriteByte('\r')
				case '\\':
					buf.WriteByte('\\')
				case '"':
					buf.WriteByte('"')
				case '{':
					buf.WriteByte('{')
				case '}':
					buf.WriteByte('}')
				default:
					buf.WriteByte('\\')
					buf.WriteByte(esc)
				}
				l.pos++
			}
			continue
		}

		if ch == '$' && l.pos+1 < len(l.source) && l.source[l.pos+1] == '{' && interpDepth == 0 {
			hasInterpolation = true
			interpDepth = 1
			buf.WriteByte(ch)
			l.pos++
			buf.WriteByte(l.source[l.pos])
			l.pos++
			continue
		}

		if interpDepth > 0 {
			if ch == '{' {
				interpDepth++
			} else if ch == '}' {
				interpDepth--
			}
			// Inside interpolation, pass through quotes and everything else
			if ch == '"' {
				// Scan the nested string literal inside ${...}
				buf.WriteByte(ch)
				l.pos++
				for l.pos < len(l.source) && l.source[l.pos] != '"' {
					if l.source[l.pos] == '\\' && l.pos+1 < len(l.source) {
						buf.WriteByte(l.source[l.pos])
						l.pos++
						buf.WriteByte(l.source[l.pos])
						l.pos++
						continue
					}
					buf.WriteByte(l.source[l.pos])
					l.pos++
				}
				if l.pos < len(l.source) {
					buf.WriteByte(l.source[l.pos]) // closing "
					l.pos++
				}
				continue
			}
		}

		buf.WriteByte(ch)
		l.pos++
	}

	if l.pos < len(l.source) {
		l.pos++ // skip closing "
	}

	if hasInterpolation {
		// Return the raw content (without quotes) for the parser to process
		raw := l.source[start+1 : l.pos-1]
		l.tokens = append(l.tokens, token.Token{Kind: token.InterpolatedString, Value: raw, Start: start, End: l.pos})
	} else {
		l.tokens = append(l.tokens, token.Token{Kind: token.String, Value: buf.String(), Start: start, End: l.pos})
	}
}

func (l *Lexer) scanTripleQuoteString(start int) {
	l.pos += 3 // skip opening """
	contentStart := l.pos

	// Find closing """
	idx := strings.Index(l.source[l.pos:], `"""`)
	if idx >= 0 {
		content := l.source[contentStart : l.pos+idx]
		l.pos += idx + 3 // skip content + closing """

		// Dedent: strip common leading whitespace
		content = dedent(content)

		// Check for interpolation: if content contains ${, emit as InterpolatedString
		if strings.Contains(content, "${") {
			l.tokens = append(l.tokens, token.Token{Kind: token.InterpolatedString, Value: content, Start: start, End: l.pos})
		} else {
			l.tokens = append(l.tokens, token.Token{Kind: token.TripleQuoteString, Value: content, Start: start, End: l.pos})
		}
	} else {
		// No closing """ — consume everything
		l.pos = len(l.source)
		l.tokens = append(l.tokens, token.Token{Kind: token.Error, Value: l.source[start:l.pos], Start: start, End: l.pos})
	}
}

func dedent(s string) string {
	s = strings.TrimSpace(s)
	lines := strings.Split(s, "\n")
	if len(lines) <= 1 {
		return s
	}

	minIndent := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " \t"))
		if minIndent < 0 || indent < minIndent {
			minIndent = indent
		}
	}
	if minIndent <= 0 {
		return s
	}

	var result []string
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			result = append(result, "")
		} else if len(line) >= minIndent {
			result = append(result, line[minIndent:])
		} else {
			result = append(result, line)
		}
	}
	return strings.Join(result, "\n")
}

func (l *Lexer) scanNumber(start int) {
	// Check for hex, binary, octal prefixes
	if l.source[l.pos] == '0' && l.pos+1 < len(l.source) {
		next := l.source[l.pos+1]
		switch next {
		case 'x', 'X':
			l.pos += 2
			for l.pos < len(l.source) && (isHexDigit(l.source[l.pos]) || l.source[l.pos] == '_') {
				l.pos++
			}
			l.tokens = append(l.tokens, token.Token{Kind: token.Int, Value: l.source[start:l.pos], Start: start, End: l.pos})
			return
		case 'b', 'B':
			l.pos += 2
			for l.pos < len(l.source) && (l.source[l.pos] == '0' || l.source[l.pos] == '1' || l.source[l.pos] == '_') {
				l.pos++
			}
			l.tokens = append(l.tokens, token.Token{Kind: token.Int, Value: l.source[start:l.pos], Start: start, End: l.pos})
			return
		case 'o', 'O':
			l.pos += 2
			for l.pos < len(l.source) && ((l.source[l.pos] >= '0' && l.source[l.pos] <= '7') || l.source[l.pos] == '_') {
				l.pos++
			}
			l.tokens = append(l.tokens, token.Token{Kind: token.Int, Value: l.source[start:l.pos], Start: start, End: l.pos})
			return
		}
	}

	// Decimal integer or float
	for l.pos < len(l.source) && (isDigit(l.source[l.pos]) || l.source[l.pos] == '_') {
		l.pos++
	}

	// Check for float: must have digit after dot (not .. or ..=)
	if l.pos < len(l.source) && l.source[l.pos] == '.' {
		if l.pos+1 < len(l.source) && isDigit(l.source[l.pos+1]) {
			l.pos++ // skip .
			for l.pos < len(l.source) && (isDigit(l.source[l.pos]) || l.source[l.pos] == '_') {
				l.pos++
			}
			l.tokens = append(l.tokens, token.Token{Kind: token.Float, Value: l.source[start:l.pos], Start: start, End: l.pos})
			return
		}
	}

	l.tokens = append(l.tokens, token.Token{Kind: token.Int, Value: l.source[start:l.pos], Start: start, End: l.pos})
}

func (l *Lexer) scanIdentOrKeyword(start int) {
	for l.pos < len(l.source) && isIdentContinue(l.source[l.pos]) {
		l.pos++
	}
	word := l.source[start:l.pos]

	if kind, ok := token.Keywords[word]; ok {
		l.tokens = append(l.tokens, token.Token{Kind: kind, Value: word, Start: start, End: l.pos})
	} else {
		l.tokens = append(l.tokens, token.Token{Kind: token.Ident, Value: word, Start: start, End: l.pos})
	}
}

func (l *Lexer) scanOperatorOrDelimiter(start int) {
	ch := l.source[l.pos]
	l.pos++

	// Two-or-three character operators (check longest match first)
	if l.pos < len(l.source) {
		next := l.source[l.pos]
		switch {
		case ch == '.' && next == '.' && l.pos+1 < len(l.source) && l.source[l.pos+1] == '=':
			l.pos += 2
			l.tokens = append(l.tokens, token.Token{Kind: token.DotDotEq, Start: start, End: l.pos})
			return
		case ch == '.' && next == '.' && l.pos+1 < len(l.source) && l.source[l.pos+1] == '.':
			l.pos += 2
			l.tokens = append(l.tokens, token.Token{Kind: token.Ellipsis, Start: start, End: l.pos})
			return
		case ch == '.' && next == '.':
			l.pos++
			l.tokens = append(l.tokens, token.Token{Kind: token.DotDot, Start: start, End: l.pos})
			return
		case ch == '=' && next == '=':
			l.pos++
			l.tokens = append(l.tokens, token.Token{Kind: token.EqEq, Start: start, End: l.pos})
			return
		case ch == '!' && next == '=':
			l.pos++
			l.tokens = append(l.tokens, token.Token{Kind: token.Ne, Start: start, End: l.pos})
			return
		case ch == '<' && next == '<':
			if l.pos+1 < len(l.source) && l.source[l.pos+1] == '=' {
				l.pos += 2
				l.tokens = append(l.tokens, token.Token{Kind: token.ShlEq, Start: start, End: l.pos})
			} else {
				l.pos++
				l.tokens = append(l.tokens, token.Token{Kind: token.Shl, Start: start, End: l.pos})
			}
			return
		case ch == '<' && next == '=':
			l.pos++
			l.tokens = append(l.tokens, token.Token{Kind: token.Le, Start: start, End: l.pos})
			return
		case ch == '>' && next == '>':
			if l.pos+1 < len(l.source) && l.source[l.pos+1] == '=' {
				l.pos += 2
				l.tokens = append(l.tokens, token.Token{Kind: token.ShrEq, Start: start, End: l.pos})
			} else {
				l.pos++
				l.tokens = append(l.tokens, token.Token{Kind: token.Shr, Start: start, End: l.pos})
			}
			return
		case ch == '>' && next == '=':
			l.pos++
			l.tokens = append(l.tokens, token.Token{Kind: token.Ge, Start: start, End: l.pos})
			return
		case ch == '|' && next == '>':
			l.pos++
			l.tokens = append(l.tokens, token.Token{Kind: token.PipeArrow, Start: start, End: l.pos})
			return
		case ch == '|' && next == '=':
			l.pos++
			l.tokens = append(l.tokens, token.Token{Kind: token.PipeEq, Start: start, End: l.pos})
			return
		case ch == '&' && next == '=':
			l.pos++
			l.tokens = append(l.tokens, token.Token{Kind: token.AmpEq, Start: start, End: l.pos})
			return
		case ch == '^' && next == '=':
			l.pos++
			l.tokens = append(l.tokens, token.Token{Kind: token.CaretEq, Start: start, End: l.pos})
			return
		case ch == '=' && next == '>':
			l.pos++
			l.tokens = append(l.tokens, token.Token{Kind: token.FatArrow, Start: start, End: l.pos})
			return
		case ch == '-' && next == '>':
			l.pos++
			l.tokens = append(l.tokens, token.Token{Kind: token.Arrow, Start: start, End: l.pos})
			return
		case ch == '+' && next == '=':
			l.pos++
			l.tokens = append(l.tokens, token.Token{Kind: token.PlusEq, Start: start, End: l.pos})
			return
		case ch == '-' && next == '=':
			l.pos++
			l.tokens = append(l.tokens, token.Token{Kind: token.MinusEq, Start: start, End: l.pos})
			return
		case ch == '*' && next == '=':
			l.pos++
			l.tokens = append(l.tokens, token.Token{Kind: token.StarEq, Start: start, End: l.pos})
			return
		case ch == '/' && next == '=':
			l.pos++
			l.tokens = append(l.tokens, token.Token{Kind: token.SlashEq, Start: start, End: l.pos})
			return
		case ch == '%' && next == '=':
			l.pos++
			l.tokens = append(l.tokens, token.Token{Kind: token.PercentEq, Start: start, End: l.pos})
			return
		}
	}

	// Single character operators/delimiters
	var kind token.TokenKind
	switch ch {
	case '+':
		kind = token.Plus
	case '-':
		kind = token.Minus
	case '*':
		kind = token.Star
	case '/':
		kind = token.Slash
	case '%':
		kind = token.Percent
	case '<':
		kind = token.Lt
	case '>':
		kind = token.Gt
	case '=':
		kind = token.Eq
	case '|':
		kind = token.Pipe
	case '&':
		kind = token.Amp
	case '^':
		kind = token.Caret
	case '~':
		kind = token.Tilde
	case '?':
		kind = token.Question
	case '.':
		kind = token.Dot
	case ':':
		kind = token.Colon
	case ',':
		kind = token.Comma
	case '@':
		kind = token.At
	case '(':
		kind = token.LParen
	case ')':
		kind = token.RParen
	case '{':
		kind = token.LBrace
	case '}':
		kind = token.RBrace
	case '[':
		kind = token.LBracket
	case ']':
		kind = token.RBracket
	default:
		kind = token.Error
	}

	l.tokens = append(l.tokens, token.Token{Kind: kind, Value: string(ch), Start: start, End: l.pos})
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isHexDigit(ch byte) bool {
	return (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

func isIdentStart(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func isIdentContinue(ch byte) bool {
	return isIdentStart(ch) || isDigit(ch)
}

// IsKeywordIdent checks if a rune could start an identifier (for unicode support).
func IsKeywordIdent(r rune) bool {
	return unicode.IsLetter(r) || r == '_'
}
