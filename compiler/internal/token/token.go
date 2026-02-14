// Package token defines token kinds and the Token type for the Haira lexer.
package token

import "fmt"

// TokenKind represents the type of a lexical token.
type TokenKind int

const (
	// Keywords
	If TokenKind = iota
	Else
	For
	While
	Return
	Match
	True
	False
	None
	Some
	And
	Or
	Not
	In
	Async
	Spawn
	Select
	Try
	Catch
	Public
	Err
	Ok
	Break
	Continue
	From
	Default
	Provider
	Tool
	Agent
	Workflow
	Fn
	Enum
	Defer
	Import
	Step

	// Operators
	Plus      // +
	Minus     // -
	Star      // *
	Slash     // /
	Percent   // %
	EqEq      // ==
	Ne        // !=
	Lt        // <
	Gt        // >
	Le        // <=
	Ge        // >=
	Eq        // =
	PlusEq    // +=
	MinusEq   // -=
	StarEq    // *=
	SlashEq   // /=
	PercentEq // %=
	Pipe      // |
	Question  // ?
	FatArrow  // =>
	Arrow     // ->
	DotDotEq  // ..=
	DotDot    // ..
	Dot       // .
	Colon     // :
	Comma     // ,
	Ellipsis  // ...
	At        // @

	// Delimiters
	LParen   // (
	RParen   // )
	LBrace   // {
	RBrace   // }
	LBracket // [
	RBracket // ]

	// Literals
	Int                // integer literal
	Float              // float literal
	String             // string literal
	InterpolatedString // interpolated string (contains {expr})
	TripleQuoteString  // """..."""

	// Identifiers
	Ident // identifier

	// Whitespace / comments
	Newline      // \n
	LineComment  // // ...
	BlockComment // /* ... */

	// Special
	EOF   // end of file
	Error // lexer error
)

// Keywords maps keyword strings to their TokenKind.
var Keywords = map[string]TokenKind{
	"if":       If,
	"else":     Else,
	"for":      For,
	"while":    While,
	"return":   Return,
	"match":    Match,
	"true":     True,
	"false":    False,
	"none":     None,
	"some":     Some,
	"and":      And,
	"or":       Or,
	"not":      Not,
	"in":       In,
	"async":    Async,
	"spawn":    Spawn,
	"select":   Select,
	"try":      Try,
	"catch":    Catch,
	"public":   Public,
	"err":      Err,
	"ok":       Ok,
	"break":    Break,
	"continue": Continue,
	"from":     From,
	"default":  Default,
	"provider": Provider,
	"tool":     Tool,
	"agent":    Agent,
	"workflow": Workflow,
	"fn":       Fn,
	"enum":     Enum,
	"defer":    Defer,
	"import":   Import,
	"step":     Step,
}

// Token represents a lexical token with its kind, literal value, and position.
type Token struct {
	Kind  TokenKind
	Value string // raw text of the token
	Start int    // byte offset start (inclusive)
	End   int    // byte offset end (exclusive)
}

func (t Token) String() string {
	if t.Value != "" {
		return fmt.Sprintf("%s(%q)", t.Kind, t.Value)
	}
	return t.Kind.String()
}

func (k TokenKind) String() string {
	switch k {
	case If:
		return "if"
	case Else:
		return "else"
	case For:
		return "for"
	case While:
		return "while"
	case Return:
		return "return"
	case Match:
		return "match"
	case True:
		return "true"
	case False:
		return "false"
	case None:
		return "none"
	case Some:
		return "some"
	case And:
		return "and"
	case Or:
		return "or"
	case Not:
		return "not"
	case In:
		return "in"
	case Async:
		return "async"
	case Spawn:
		return "spawn"
	case Select:
		return "select"
	case Try:
		return "try"
	case Catch:
		return "catch"
	case Public:
		return "public"
	case Err:
		return "err"
	case Ok:
		return "ok"
	case Break:
		return "break"
	case Continue:
		return "continue"
	case From:
		return "from"
	case Default:
		return "default"
	case Provider:
		return "provider"
	case Tool:
		return "tool"
	case Agent:
		return "agent"
	case Workflow:
		return "workflow"
	case Fn:
		return "fn"
	case Enum:
		return "enum"
	case Defer:
		return "defer"
	case Import:
		return "import"
	case Step:
		return "step"
	case Plus:
		return "+"
	case Minus:
		return "-"
	case Star:
		return "*"
	case Slash:
		return "/"
	case Percent:
		return "%"
	case EqEq:
		return "=="
	case Ne:
		return "!="
	case Lt:
		return "<"
	case Gt:
		return ">"
	case Le:
		return "<="
	case Ge:
		return ">="
	case Eq:
		return "="
	case PlusEq:
		return "+="
	case MinusEq:
		return "-="
	case StarEq:
		return "*="
	case SlashEq:
		return "/="
	case PercentEq:
		return "%="
	case Pipe:
		return "|"
	case Question:
		return "?"
	case FatArrow:
		return "=>"
	case Arrow:
		return "->"
	case DotDotEq:
		return "..="
	case DotDot:
		return ".."
	case Dot:
		return "."
	case Colon:
		return ":"
	case Comma:
		return ","
	case Ellipsis:
		return "..."
	case At:
		return "@"
	case LParen:
		return "("
	case RParen:
		return ")"
	case LBrace:
		return "{"
	case RBrace:
		return "}"
	case LBracket:
		return "["
	case RBracket:
		return "]"
	case Int:
		return "Int"
	case Float:
		return "Float"
	case String:
		return "String"
	case InterpolatedString:
		return "InterpolatedString"
	case TripleQuoteString:
		return "TripleQuoteString"
	case Ident:
		return "Ident"
	case Newline:
		return "Newline"
	case LineComment:
		return "LineComment"
	case BlockComment:
		return "BlockComment"
	case EOF:
		return "EOF"
	case Error:
		return "Error"
	default:
		return fmt.Sprintf("TokenKind(%d)", int(k))
	}
}

// IsKeyword returns true if the token kind is a keyword.
func (k TokenKind) IsKeyword() bool {
	return k >= If && k <= Step
}

// IsLiteral returns true if the token kind is a literal value.
func (k TokenKind) IsLiteral() bool {
	return k >= Int && k <= TripleQuoteString
}

// IsTrivia returns true if the token kind is a comment.
func (k TokenKind) IsTrivia() bool {
	return k == LineComment || k == BlockComment
}
