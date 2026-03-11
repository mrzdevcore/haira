// Package parser implements a recursive descent parser with Pratt expression
// parsing for Haira source code.
package parser

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/haira-lang/haira/internal/ast"
	"github.com/haira-lang/haira/internal/lexer"
	"github.com/haira-lang/haira/internal/token"
)

// Precedence levels for Pratt expression parsing.
type Precedence int

const (
	PrecNone       Precedence = iota
	PrecOrelse                // orelse
	PrecPipe                  // |>
	PrecOr                    // or
	PrecAnd                   // and
	PrecBitOr                 // |
	PrecBitXor                // ^
	PrecBitAnd                // &
	PrecEquality              // == !=
	PrecComparison            // < > <= >= .. ..=
	PrecShift                 // << >>
	PrecTerm                  // + -
	PrecFactor                // * / %
	PrecUnary                 // - not ~
	PrecCall                  // () [] . ?
)

// precedenceOf returns the precedence of a token kind for infix parsing.
func precedenceOf(kind token.TokenKind) Precedence {
	switch kind {
	case token.Orelse:
		return PrecOrelse
	case token.PipeArrow:
		return PrecPipe
	case token.Or:
		return PrecOr
	case token.And:
		return PrecAnd
	case token.Pipe:
		return PrecBitOr
	case token.Caret:
		return PrecBitXor
	case token.Amp:
		return PrecBitAnd
	case token.EqEq, token.Ne:
		return PrecEquality
	case token.Lt, token.Gt, token.Le, token.Ge:
		return PrecComparison
	case token.Shl, token.Shr:
		return PrecShift
	case token.Plus, token.Minus:
		return PrecTerm
	case token.Star, token.Slash, token.Percent:
		return PrecFactor
	case token.LParen, token.LBracket, token.Dot, token.Question:
		return PrecCall
	case token.DotDot, token.DotDotEq:
		return PrecComparison
	default:
		return PrecNone
	}
}

// ParseError represents a single parse error with a message and source span.
type ParseError struct {
	Message string
	Span    ast.Span
}

// Parser holds the state for parsing a stream of tokens.
type Parser struct {
	tokens   []token.Token
	current  int
	previous token.Token
	errors   []ParseError
}

// Parse tokenizes source and parses it into a SourceFile.
func Parse(source string) (*ast.SourceFile, []ParseError) {
	l := lexer.New(source)
	tokens := l.AllTokens()

	p := &Parser{
		tokens:  tokens,
		current: 0,
		previous: token.Token{
			Kind:  token.EOF,
			Start: 0,
			End:   0,
		},
	}
	// Advance to the first significant token.
	p.skipTrivia()

	sf := p.parseSourceFile()
	return sf, p.errors
}

// ---------------------------------------------------------------------------
// Token helpers
// ---------------------------------------------------------------------------

// peek returns the current token without advancing.
func (p *Parser) peek() token.Token {
	if p.current < len(p.tokens) {
		return p.tokens[p.current]
	}
	end := 0
	if len(p.tokens) > 0 {
		end = p.tokens[len(p.tokens)-1].End
	}
	return token.Token{Kind: token.EOF, Start: end, End: end}
}

// advance moves past the current token (skipping newlines and comments) and
// stores the consumed token in p.previous.
func (p *Parser) advance() {
	if p.current < len(p.tokens) {
		p.previous = p.tokens[p.current]
		p.current++
	}
	p.skipTrivia()
}

// skipTrivia skips Newline, LineComment and BlockComment tokens.
func (p *Parser) skipTrivia() {
	for p.current < len(p.tokens) {
		k := p.tokens[p.current].Kind
		if k == token.Newline || k == token.LineComment || k == token.BlockComment {
			p.current++
		} else {
			break
		}
	}
}

// skipNewlines advances past any Newline tokens only (not comments, which are
// already handled by advance/skipTrivia).
func (p *Parser) skipNewlines() {
	for p.current < len(p.tokens) && p.tokens[p.current].Kind == token.Newline {
		p.current++
	}
}

// check returns true if the current token is of the given kind.
func (p *Parser) check(kind token.TokenKind) bool {
	return p.peek().Kind == kind
}

// atEnd returns true if the current token is EOF.
func (p *Parser) atEnd() bool {
	return p.peek().Kind == token.EOF
}

// consume checks that the current token matches kind, advances, and returns
// true. Otherwise it records an error and returns false.
func (p *Parser) consume(kind token.TokenKind, expected string) bool {
	if p.check(kind) {
		p.advance()
		return true
	}
	tok := p.peek()
	p.addError("expected "+expected+", found "+tok.Kind.String(), ast.Span{Start: tok.Start, End: tok.End})
	return false
}

// addError appends a ParseError.
func (p *Parser) addError(msg string, span ast.Span) {
	p.errors = append(p.errors, ParseError{Message: msg, Span: span})
}

// span returns a Span from start to the end of the previous token.
func (p *Parser) span(start int) ast.Span {
	return ast.Span{Start: start, End: p.previous.End}
}

// currentSpan returns the span of the current token.
func (p *Parser) currentSpan() ast.Span {
	tok := p.peek()
	return ast.Span{Start: tok.Start, End: tok.End}
}

// ---------------------------------------------------------------------------
// Top-level parsing
// ---------------------------------------------------------------------------

func (p *Parser) parseSourceFile() *ast.SourceFile {
	start := p.peek().Start
	var items []ast.Item

	for !p.atEnd() {
		p.skipNewlines()
		if p.atEnd() {
			break
		}
		item, ok := p.parseItem()
		if ok {
			items = append(items, item)
		} else {
			// Error recovery: skip token.
			p.advance()
		}
	}

	return &ast.SourceFile{
		Items: items,
		Span:  p.span(start),
	}
}

func (p *Parser) parseItem() (ast.Item, bool) {
	start := p.peek().Start

	// Check for `pub` modifier.
	isPublic := false
	if p.check(token.Pub) {
		p.advance()
		isPublic = true
	}

	switch p.peek().Kind {
	case token.Ident:
		return p.parseItemIdent(start, isPublic)

	case token.Import:
		p.advance()
		decl, ok := p.parseImportDecl()
		if !ok {
			return ast.Item{}, false
		}
		return ast.Item{Node: decl, Span: p.span(start)}, true

	case token.Export:
		p.advance()
		decl, ok := p.parseExportDecl()
		if !ok {
			return ast.Item{}, false
		}
		return ast.Item{Node: decl, Span: p.span(start)}, true

	case token.Provider:
		p.advance()
		decl, ok := p.parseProviderDecl()
		if !ok {
			return ast.Item{}, false
		}
		return ast.Item{Node: decl, Span: p.span(start)}, true

	case token.Tool:
		p.advance()
		decl, ok := p.parseToolDecl()
		if !ok {
			return ast.Item{}, false
		}
		return ast.Item{Node: decl, Span: p.span(start)}, true

	case token.Agent:
		p.advance()
		decl, ok := p.parseAgentDecl()
		if !ok {
			return ast.Item{}, false
		}
		return ast.Item{Node: decl, Span: p.span(start)}, true

	case token.At:
		// Collect all decorators before a workflow declaration
		var decorators []ast.Decorator
		for p.check(token.At) {
			dec, ok := p.parseDecorator()
			if !ok {
				return ast.Item{}, false
			}
			decorators = append(decorators, dec)
		}
		if p.check(token.Workflow) {
			p.advance()
			// Separate trigger from other decorators
			var trigger *ast.Decorator
			var extras []ast.Decorator
			for i := range decorators {
				switch decorators[i].Name.Node {
				case "webhook", "get", "post", "put", "delete",
					"cron", "event", "manual", "websocket":
					trigger = &decorators[i]
				default:
					extras = append(extras, decorators[i])
				}
			}
			wf, ok := p.parseWorkflowDecl(trigger)
			if !ok {
				return ast.Item{}, false
			}
			wf.Decorators = extras
			return ast.Item{Node: wf, Span: p.span(start)}, true
		}
		if p.check(token.Tool) {
			p.advance()
			tool, ok := p.parseToolDecl()
			if !ok {
				return ast.Item{}, false
			}
			tool.Decorators = decorators
			return ast.Item{Node: tool, Span: p.span(start)}, true
		}
		tok := p.peek()
		p.addError("expected workflow or tool after decorator", ast.Span{Start: tok.Start, End: tok.End})
		return ast.Item{}, false

	case token.Workflow:
		p.advance()
		wf, ok := p.parseWorkflowDecl(nil)
		if !ok {
			return ast.Item{}, false
		}
		return ast.Item{Node: wf, Span: p.span(start)}, true

	case token.Test:
		p.advance()
		td, ok := p.parseTestDecl()
		if !ok {
			return ast.Item{}, false
		}
		return ast.Item{Node: td, Span: p.span(start)}, true

	case token.Eval:
		p.advance()
		ed, ok := p.parseEvalDecl()
		if !ok {
			return ast.Item{}, false
		}
		return ast.Item{Node: ed, Span: p.span(start)}, true

	case token.Fn:
		p.advance()
		fd, ok := p.parseFnDecl(isPublic)
		if !ok {
			return ast.Item{}, false
		}
		return ast.Item{Node: fd, Span: p.span(start)}, true

	case token.Enum:
		p.advance()
		ed, ok := p.parseEnumDecl(isPublic)
		if !ok {
			return ast.Item{}, false
		}
		return ast.Item{Node: ed, Span: p.span(start)}, true

	case token.Struct:
		p.advance()
		name, ok := p.parseIdentifier()
		if !ok {
			return ast.Item{}, false
		}
		td, ok := p.parseTypeDefBody(isPublic, name)
		if !ok {
			return ast.Item{}, false
		}
		return ast.Item{Node: td, Span: p.span(start)}, true

	case token.Type:
		p.advance()
		name, ok := p.parseIdentifier()
		if !ok {
			return ast.Item{}, false
		}
		p.consume(token.Eq, "=")
		ty, ok := p.parseType()
		if !ok {
			return ast.Item{}, false
		}
		return ast.Item{
			Node: ast.TypeAlias{Name: name, Ty: ty},
			Span: p.span(start),
		}, true

	case token.If, token.For, token.While, token.Return, token.Match,
		token.Try, token.Defer, token.Break, token.Continue,
		token.Spawn, token.Async, token.Let, token.Const:
		stmt, ok := p.parseStatement()
		if !ok {
			return ast.Item{}, false
		}
		return ast.Item{Node: ast.ItemStatement{Stmt: stmt}, Span: stmt.Span}, true

	default:
		// Could be an expression statement starting with a literal, unary, etc.
		// Try to parse as a statement.
		if p.peek().Kind == token.Minus || p.peek().Kind == token.Not ||
			p.peek().Kind == token.LParen || p.peek().Kind == token.LBracket ||
			p.peek().Kind == token.LBrace ||
			p.peek().Kind == token.Int || p.peek().Kind == token.Float ||
			p.peek().Kind == token.String || p.peek().Kind == token.InterpolatedString ||
			p.peek().Kind == token.TripleQuoteString ||
			p.peek().Kind == token.True || p.peek().Kind == token.False ||
			p.peek().Kind == token.None || p.peek().Kind == token.Nil || p.peek().Kind == token.Some ||
			p.peek().Kind == token.Err || p.peek().Kind == token.Ok ||
			p.peek().Kind == token.Select {
			stmt, ok := p.parseStatement()
			if !ok {
				return ast.Item{}, false
			}
			return ast.Item{Node: ast.ItemStatement{Stmt: stmt}, Span: stmt.Span}, true
		}
		tok := p.peek()
		p.addError("expected item or statement", ast.Span{Start: tok.Start, End: tok.End})
		return ast.Item{}, false
	}
}

// parseItemIdent handles items that start with an identifier: type defs,
// function defs, method defs, type aliases, assignments, expression stmts.
func (p *Parser) parseItemIdent(start int, isPublic bool) (ast.Item, bool) {
	name, ok := p.parseIdentifier()
	if !ok {
		return ast.Item{}, false
	}

	switch p.peek().Kind {
	// Map/struct literal or block: name { ... }
	case token.LBrace:
		expr := ast.Expr{Node: ast.IdentExpr{Name: name.Node}, Span: name.Span}
		fullExpr, ok := p.parseExprRest(expr)
		if !ok {
			return ast.Item{}, false
		}
		stmt, ok := p.parseStatementRest(fullExpr)
		if !ok {
			return ast.Item{}, false
		}
		return ast.Item{Node: ast.ItemStatement{Stmt: stmt}, Span: p.span(start)}, true

	// Function def or call: name(...)
	case token.LParen:
		expr := ast.Expr{Node: ast.IdentExpr{Name: name.Node}, Span: name.Span}
		callExpr, ok := p.parseInfix(expr, PrecNone)
		if !ok {
			return ast.Item{}, false
		}
		// If followed by { or ->, it's a function definition.
		if p.check(token.LBrace) || p.check(token.Arrow) {
			call, isCall := callExpr.Node.(ast.CallExpr)
			if !isCall {
				tok := p.peek()
				p.addError("expected function definition", ast.Span{Start: tok.Start, End: tok.End})
				return ast.Item{}, false
			}
			params, ok := p.argsToParams(call.Args)
			if !ok {
				return ast.Item{}, false
			}
			var retTy *ast.Spanned[ast.Type]
			if p.check(token.Arrow) {
				p.advance()
				rt, ok := p.parseType()
				if !ok {
					return ast.Item{}, false
				}
				retTy = &rt
			}
			body, ok := p.parseBlock()
			if !ok {
				return ast.Item{}, false
			}
			fd := ast.FunctionDef{
				IsPublic: isPublic,
				Name:     name,
				Params:   params,
				ReturnTy: retTy,
				Body:     body,
			}
			return ast.Item{Node: fd, Span: p.span(start)}, true
		}
		// Otherwise expression statement (function call).
		stmt := ast.Statement{Node: ast.ExprStmt{Value: callExpr}, Span: p.span(start)}
		return ast.Item{Node: ast.ItemStatement{Stmt: stmt}, Span: p.span(start)}, true

	// Method definition or field access: Name.method or obj.field
	case token.Dot:
		firstChar := firstRune(name.Node)
		if unicode.IsUpper(firstChar) {
			// Method definition: Type.method(...)
			p.advance() // consume .
			methodName, ok := p.parseIdentifier()
			if !ok {
				return ast.Item{}, false
			}
			md, ok := p.parseMethodDefBody(name, methodName)
			if !ok {
				return ast.Item{}, false
			}
			return ast.Item{Node: md, Span: p.span(start)}, true
		}
		// Field access expression → statement
		expr := ast.Expr{Node: ast.IdentExpr{Name: name.Node}, Span: name.Span}
		fullExpr, ok := p.parseExprRest(expr)
		if !ok {
			return ast.Item{}, false
		}
		stmt, ok := p.parseStatementRest(fullExpr)
		if !ok {
			return ast.Item{}, false
		}
		return ast.Item{Node: ast.ItemStatement{Stmt: stmt}, Span: p.span(start)}, true

	// Variable assignment: name = ...
	case token.Eq:
		expr := ast.Expr{Node: ast.IdentExpr{Name: name.Node}, Span: name.Span}
		stmt, ok := p.parseStatementRest(expr)
		if !ok {
			return ast.Item{}, false
		}
		return ast.Item{Node: ast.ItemStatement{Stmt: stmt}, Span: p.span(start)}, true

	// Compound assignment: name += / -= / *= / /= / %=
	case token.PlusEq, token.MinusEq, token.StarEq, token.SlashEq, token.PercentEq:
		expr := ast.Expr{Node: ast.IdentExpr{Name: name.Node}, Span: name.Span}
		stmt, ok := p.parseStatementRest(expr)
		if !ok {
			return ast.Item{}, false
		}
		return ast.Item{Node: ast.ItemStatement{Stmt: stmt}, Span: p.span(start)}, true

	// Multi-assignment: a, b = ...
	case token.Comma:
		expr := ast.Expr{Node: ast.IdentExpr{Name: name.Node}, Span: name.Span}
		stmt, ok := p.parseStatementRest(expr)
		if !ok {
			return ast.Item{}, false
		}
		return ast.Item{Node: ast.ItemStatement{Stmt: stmt}, Span: p.span(start)}, true

	default:
		// Anything else: treat identifier as start of expression statement.
		// Must complete expression parsing (e.g. index, pipe) before checking for assignment.
		expr := ast.Expr{Node: ast.IdentExpr{Name: name.Node}, Span: name.Span}
		fullExpr, ok := p.parseExprRest(expr)
		if !ok {
			return ast.Item{}, false
		}
		stmt, ok := p.parseStatementRest(fullExpr)
		if !ok {
			return ast.Item{}, false
		}
		return ast.Item{Node: ast.ItemStatement{Stmt: stmt}, Span: p.span(start)}, true
	}
}

// ---------------------------------------------------------------------------
// Type definitions
// ---------------------------------------------------------------------------

func (p *Parser) parseTypeDefBody(isPublic bool, name ast.Spanned[string]) (ast.TypeDef, bool) {
	p.consume(token.LBrace, "{")
	p.skipNewlines()

	var fields []ast.Field
	for !p.check(token.RBrace) && !p.atEnd() {
		f, ok := p.parseField()
		if ok {
			fields = append(fields, f)
		} else {
			// Skip token to avoid infinite loop on parse failure
			p.advance()
		}
		if p.check(token.Comma) {
			p.advance()
		}
		p.skipNewlines()
	}

	p.consume(token.RBrace, "}")
	return ast.TypeDef{IsPublic: isPublic, Name: name, Fields: fields}, true
}

func (p *Parser) parseField() (ast.Field, bool) {
	start := p.peek().Start
	name, ok := p.parseIdentifier()
	if !ok {
		return ast.Field{}, false
	}

	var ty *ast.Spanned[ast.Type]
	if p.check(token.Colon) {
		p.advance()
		t, ok := p.parseType()
		if !ok {
			return ast.Field{}, false
		}
		ty = &t
	}

	var def *ast.Expr
	if p.check(token.Eq) {
		p.advance()
		e, ok := p.parseExpr()
		if !ok {
			return ast.Field{}, false
		}
		def = &e
	}

	return ast.Field{Name: name, Ty: ty, Default: def, Span: p.span(start)}, true
}

// ---------------------------------------------------------------------------
// Enum definitions
// ---------------------------------------------------------------------------

func (p *Parser) parseEnumDecl(isPublic bool) (ast.EnumDef, bool) {
	name, ok := p.parseIdentifier()
	if !ok {
		return ast.EnumDef{}, false
	}
	p.consume(token.LBrace, "{")
	p.skipNewlines()

	var variants []ast.EnumVariant
	for !p.check(token.RBrace) && !p.atEnd() {
		vstart := p.peek().Start
		vname, ok := p.parseIdentifier()
		if !ok {
			break
		}
		var fields []ast.Param
		if p.check(token.LParen) {
			f, ok := p.parseParams()
			if !ok {
				break
			}
			fields = f
		}
		variants = append(variants, ast.EnumVariant{
			Name:   vname,
			Fields: fields,
			Span:   p.span(vstart),
		})
		if p.check(token.Comma) {
			p.advance()
		}
		p.skipNewlines()
	}

	p.consume(token.RBrace, "}")
	return ast.EnumDef{IsPublic: isPublic, Name: name, Variants: variants}, true
}

// ---------------------------------------------------------------------------
// Function / method definitions
// ---------------------------------------------------------------------------

func (p *Parser) parseMethodDefBody(typeName, methodName ast.Spanned[string]) (ast.MethodDef, bool) {
	params, ok := p.parseParams()
	if !ok {
		return ast.MethodDef{}, false
	}

	var retTy *ast.Spanned[ast.Type]
	if p.check(token.Arrow) {
		p.advance()
		rt, ok := p.parseType()
		if !ok {
			return ast.MethodDef{}, false
		}
		retTy = &rt
	}

	body, ok := p.parseBlock()
	if !ok {
		return ast.MethodDef{}, false
	}

	return ast.MethodDef{
		TypeName: typeName,
		Name:     methodName,
		Params:   params,
		ReturnTy: retTy,
		Body:     body,
	}, true
}

func (p *Parser) parseParams() ([]ast.Param, bool) {
	p.consume(token.LParen, "(")

	var params []ast.Param
	sawRest := false
	for !p.check(token.RParen) && !p.atEnd() {
		param, ok := p.parseParam()
		if !ok {
			return nil, false
		}
		if sawRest {
			p.addError("rest parameter must be last", param.Span)
		}
		if param.IsRest {
			sawRest = true
		}
		params = append(params, param)
		if !p.check(token.RParen) {
			p.consume(token.Comma, ",")
		}
	}

	p.consume(token.RParen, ")")
	return params, true
}

func (p *Parser) parseParam() (ast.Param, bool) {
	start := p.peek().Start

	isRest := false
	if p.check(token.Ellipsis) {
		p.advance()
		isRest = true
	}

	name, ok := p.parseIdentifier()
	if !ok {
		return ast.Param{}, false
	}

	var ty *ast.Spanned[ast.Type]
	if p.check(token.Colon) {
		p.advance()
		t, ok := p.parseType()
		if !ok {
			return ast.Param{}, false
		}
		ty = &t
	}

	var def *ast.Expr
	if p.check(token.Eq) {
		p.advance()
		e, ok := p.parseExpr()
		if !ok {
			return ast.Param{}, false
		}
		def = &e
	}

	if !isRest && p.check(token.Ellipsis) {
		p.advance()
		isRest = true
	}

	return ast.Param{
		Name:    name,
		Ty:      ty,
		Default: def,
		IsRest:  isRest,
		Span:    p.span(start),
	}, true
}

// argsToParams converts call arguments back to function parameters when we
// discover that what looked like a call was actually a function definition.
func (p *Parser) argsToParams(args []ast.Argument) ([]ast.Param, bool) {
	var params []ast.Param
	for _, arg := range args {
		if ident, ok := arg.Value.Node.(ast.IdentExpr); ok {
			params = append(params, ast.Param{
				Name: ast.Spanned[string]{Node: ident.Name, Span: arg.Value.Span},
				Span: arg.Span,
			})
		} else if arg.Name != nil {
			params = append(params, ast.Param{
				Name:    *arg.Name,
				Default: &arg.Value,
				Span:    arg.Span,
			})
		} else {
			p.addError("expected identifier in parameter list", arg.Value.Span)
			return nil, false
		}
	}
	return params, true
}

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

func (p *Parser) parseType() (ast.Spanned[ast.Type], bool) {
	start := p.peek().Start

	var ty ast.Type

	switch p.peek().Kind {
	case token.Ident:
		name := p.peek().Value
		p.advance()
		// Check for generic args: Name<T>
		if p.check(token.Lt) {
			p.advance()
			var args []ast.Spanned[ast.Type]
			for !p.check(token.Gt) && !p.atEnd() {
				a, ok := p.parseType()
				if !ok {
					return ast.Spanned[ast.Type]{}, false
				}
				args = append(args, a)
				if !p.check(token.Gt) {
					p.consume(token.Comma, ",")
				}
			}
			p.consume(token.Gt, ">")
			ty = ast.GenericType{Name: name, Args: args}
		} else {
			ty = ast.NamedType{Name: name}
		}

		// Qualified type: ui.StatusCard → NamedType{Name: "ui.StatusCard"}
		if _, isNamed := ty.(ast.NamedType); isNamed && p.check(token.Dot) {
			// Peek ahead: if next after dot is an uppercase ident, consume as qualified type
			if p.current+1 < len(p.tokens) {
				next := p.tokens[p.current+1]
				if next.Kind == token.Ident && unicode.IsUpper(firstRune(next.Value)) {
					p.advance() // consume .
					qualName := p.peek().Value
					p.advance() // consume the qualified name
					ty = ast.NamedType{Name: name + "." + qualName}
				}
			}
		}

	case token.LBracket:
		p.advance()
		inner, ok := p.parseType()
		if !ok {
			return ast.Spanned[ast.Type]{}, false
		}
		p.consume(token.RBracket, "]")
		ty = ast.ListType{Elem: inner}

	case token.LBrace:
		p.advance()
		p.skipNewlines()
		key, ok := p.parseType()
		if !ok {
			return ast.Spanned[ast.Type]{}, false
		}
		p.consume(token.Colon, ":")
		value, ok := p.parseType()
		if !ok {
			return ast.Spanned[ast.Type]{}, false
		}
		// Skip additional fields (record types).
		for p.check(token.Comma) {
			p.advance()
			p.skipNewlines()
			if p.check(token.RBrace) {
				break
			}
			_, ok := p.parseType()
			if !ok {
				return ast.Spanned[ast.Type]{}, false
			}
			p.consume(token.Colon, ":")
			_, ok = p.parseType()
			if !ok {
				return ast.Spanned[ast.Type]{}, false
			}
		}
		p.skipNewlines()
		p.consume(token.RBrace, "}")
		ty = ast.MapType{Key: key, Value: value}

	case token.LParen:
		p.advance()
		var innerTypes []ast.Spanned[ast.Type]
		for !p.check(token.RParen) && !p.atEnd() {
			pt, ok := p.parseType()
			if !ok {
				return ast.Spanned[ast.Type]{}, false
			}
			innerTypes = append(innerTypes, pt)
			if !p.check(token.RParen) {
				p.consume(token.Comma, ",")
			}
		}
		p.consume(token.RParen, ")")
		if p.check(token.Arrow) {
			// Function type: (T1, T2) -> R
			p.advance()
			ret, ok := p.parseType()
			if !ok {
				return ast.Spanned[ast.Type]{}, false
			}
			ty = ast.FunctionType{Params: innerTypes, Ret: ret}
		} else {
			// Tuple type: (T1, T2)
			ty = ast.TupleType{Elems: innerTypes}
		}

	default:
		tok := p.peek()
		p.addError("expected type", ast.Span{Start: tok.Start, End: tok.End})
		return ast.Spanned[ast.Type]{}, false
	}

	// Check for option type: T?
	if p.check(token.Question) {
		p.advance()
		ty = ast.OptionType{Inner: ast.Spanned[ast.Type]{Node: ty, Span: p.span(start)}}
	}

	// Check for union: Type | Other
	if p.check(token.Pipe) {
		variants := []ast.Spanned[ast.Type]{{Node: ty, Span: p.span(start)}}
		for p.check(token.Pipe) {
			p.advance()
			v, ok := p.parseType()
			if !ok {
				return ast.Spanned[ast.Type]{}, false
			}
			variants = append(variants, v)
		}
		return ast.Spanned[ast.Type]{Node: ast.UnionType{Variants: variants}, Span: p.span(start)}, true
	}

	return ast.Spanned[ast.Type]{Node: ty, Span: p.span(start)}, true
}

// ---------------------------------------------------------------------------
// Statements
// ---------------------------------------------------------------------------

func (p *Parser) parseStatement() (ast.Statement, bool) {
	start := p.peek().Start

	switch p.peek().Kind {
	case token.If:
		p.advance()
		ifStmt, ok := p.parseIfStatement()
		if !ok {
			return ast.Statement{}, false
		}
		return ast.Statement{Node: ast.IfStmt{
			Condition:  ifStmt.Condition,
			ThenBranch: ifStmt.ThenBranch,
			ElseBranch: ifStmt.ElseBranch,
		}, Span: p.span(start)}, true

	case token.For:
		p.advance()
		fs, ok := p.parseForStatement()
		if !ok {
			return ast.Statement{}, false
		}
		return ast.Statement{Node: fs, Span: p.span(start)}, true

	case token.While:
		p.advance()
		ws, ok := p.parseWhileStatement()
		if !ok {
			return ast.Statement{}, false
		}
		return ast.Statement{Node: ws, Span: p.span(start)}, true

	case token.Return:
		p.advance()
		rs, ok := p.parseReturnStatement()
		if !ok {
			return ast.Statement{}, false
		}
		return ast.Statement{Node: rs, Span: p.span(start)}, true

	case token.Match:
		p.advance()
		me, ok := p.parseMatchExpr()
		if !ok {
			return ast.Statement{}, false
		}
		return ast.Statement{Node: ast.MatchStmt{Match: me}, Span: p.span(start)}, true

	case token.Try:
		p.advance()
		ts, ok := p.parseTryStatement()
		if !ok {
			return ast.Statement{}, false
		}
		return ast.Statement{Node: ts, Span: p.span(start)}, true

	case token.Defer:
		p.advance()
		expr, ok := p.parseExpr()
		if !ok {
			return ast.Statement{}, false
		}
		return ast.Statement{Node: ast.DeferStmt{Value: expr}, Span: p.span(start)}, true

	case token.Errdefer:
		p.advance()
		expr, ok := p.parseExpr()
		if !ok {
			return ast.Statement{}, false
		}
		return ast.Statement{Node: ast.ErrDeferStmt{Value: expr}, Span: p.span(start)}, true

	case token.Step:
		p.advance()
		ss, ok := p.parseStepStatement()
		if !ok {
			return ast.Statement{}, false
		}
		return ast.Statement{Node: ss, Span: p.span(start)}, true

	case token.Assert:
		p.advance()
		as, ok := p.parseAssertStmt()
		if !ok {
			return ast.Statement{}, false
		}
		return ast.Statement{Node: as, Span: p.span(start)}, true

	case token.Verify:
		p.advance()
		body, ok := p.parseBlock()
		if !ok {
			return ast.Statement{}, false
		}
		return ast.Statement{Node: ast.VerifyStmt{Body: body}, Span: p.span(start)}, true

	case token.Let, token.Const:
		isConst := p.peek().Kind == token.Const
		p.advance()
		ls, ok := p.parseLetStatement(isConst)
		if !ok {
			return ast.Statement{}, false
		}
		return ast.Statement{Node: ls, Span: p.span(start)}, true

	case token.Break:
		p.advance()
		return ast.Statement{Node: ast.BreakStmt{}, Span: p.span(start)}, true

	case token.Continue:
		p.advance()
		return ast.Statement{Node: ast.ContinueStmt{}, Span: p.span(start)}, true

	case token.Spawn, token.Async:
		expr, ok := p.parseExpr()
		if !ok {
			return ast.Statement{}, false
		}
		return p.parseStatementRest(expr)

	default:
		expr, ok := p.parseExpr()
		if !ok {
			return ast.Statement{}, false
		}
		return p.parseStatementRest(expr)
	}
}

// parseLetStatement parses: let name [: Type] = expr  /  const name [: Type] = expr
func (p *Parser) parseLetStatement(isConst bool) (ast.LetStmt, bool) {
	if !p.check(token.Ident) {
		kw := "let"
		if isConst {
			kw = "const"
		}
		tok := p.peek()
		p.addError("expected variable name after '"+kw+"'", ast.Span{Start: tok.Start, End: tok.End})
		return ast.LetStmt{}, false
	}
	nameTok := p.peek()
	name := ast.Spanned[string]{Node: nameTok.Value, Span: ast.Span{Start: nameTok.Start, End: nameTok.End}}
	p.advance()

	// Optional type annotation: let x: int = ...
	var typeAnn *ast.Spanned[ast.Type]
	if p.check(token.Colon) {
		p.advance()
		ty, ok := p.parseType()
		if !ok {
			return ast.LetStmt{}, false
		}
		typeAnn = &ty
	}

	if !p.consume(token.Eq, "'='") {
		return ast.LetStmt{}, false
	}

	value, ok := p.parseExpr()
	if !ok {
		return ast.LetStmt{}, false
	}

	return ast.LetStmt{
		Name:    name,
		TypeAnn: typeAnn,
		Value:   value,
		IsConst: isConst,
	}, true
}

// parseStatementRest takes an already-parsed expression and determines whether
// the overall statement is an assignment, compound assignment, multi-assignment,
// or expression statement.
func (p *Parser) parseStatementRest(firstExpr ast.Expr) (ast.Statement, bool) {
	start := firstExpr.Span.Start

	// Compound assignment: += -= *= /= %=
	compoundOp := ast.BinaryOp(-1)
	switch p.peek().Kind {
	case token.PlusEq:
		compoundOp = ast.OpAdd
	case token.MinusEq:
		compoundOp = ast.OpSub
	case token.StarEq:
		compoundOp = ast.OpMul
	case token.SlashEq:
		compoundOp = ast.OpDiv
	case token.PercentEq:
		compoundOp = ast.OpMod
	case token.AmpEq:
		compoundOp = ast.OpBitAnd
	case token.PipeEq:
		compoundOp = ast.OpBitOr
	case token.CaretEq:
		compoundOp = ast.OpBitXor
	case token.ShlEq:
		compoundOp = ast.OpShl
	case token.ShrEq:
		compoundOp = ast.OpShr
	}
	if compoundOp >= 0 {
		opSpan := p.currentSpan()
		p.advance()
		rhs, ok := p.parseExpr()
		if !ok {
			return ast.Statement{}, false
		}
		target, ok := p.exprToAssignTarget(firstExpr)
		if !ok {
			return ast.Statement{}, false
		}
		// Desugar: x += y  =>  x = x + y
		value := ast.Expr{
			Node: ast.BinaryExpr{
				Left:  firstExpr,
				Op:    ast.Spanned[ast.BinaryOp]{Node: compoundOp, Span: opSpan},
				Right: rhs,
			},
			Span: p.span(start),
		}
		return ast.Statement{
			Node: ast.AssignStmt{Targets: []ast.AssignTarget{target}, Value: value},
			Span: p.span(start),
		}, true
	}

	// Simple assignment: expr = value
	if p.check(token.Eq) {
		p.advance()
		value, ok := p.parseExpr()
		if !ok {
			return ast.Statement{}, false
		}
		target, ok := p.exprToAssignTarget(firstExpr)
		if !ok {
			return ast.Statement{}, false
		}
		return ast.Statement{
			Node: ast.AssignStmt{Targets: []ast.AssignTarget{target}, Value: value},
			Span: p.span(start),
		}, true
	}

	// Type-annotated assignment: result: Type = value
	// or multi-assignment with annotation: result: Type, err = value
	if p.check(token.Colon) {
		firstTarget, ok := p.exprToAnnotatedTarget(firstExpr)
		if !ok {
			return ast.Statement{}, false
		}
		targets := []ast.AssignTarget{firstTarget}

		for p.check(token.Comma) {
			p.advance()
			expr, ok := p.parseExpr()
			if !ok {
				return ast.Statement{}, false
			}
			tgt, ok := p.exprToAnnotatedTarget(expr)
			if !ok {
				return ast.Statement{}, false
			}
			targets = append(targets, tgt)
		}

		p.consume(token.Eq, "=")
		value, ok := p.parseExpr()
		if !ok {
			return ast.Statement{}, false
		}
		return ast.Statement{
			Node: ast.AssignStmt{Targets: targets, Value: value},
			Span: p.span(start),
		}, true
	}

	// Multi-assignment: a, b = ...
	if p.check(token.Comma) {
		targets := []ast.AssignTarget{}
		t, ok := p.exprToAssignTarget(firstExpr)
		if !ok {
			return ast.Statement{}, false
		}
		targets = append(targets, t)

		for p.check(token.Comma) {
			p.advance()
			expr, ok := p.parseExpr()
			if !ok {
				return ast.Statement{}, false
			}
			tgt, ok := p.exprToAssignTarget(expr)
			if !ok {
				return ast.Statement{}, false
			}
			targets = append(targets, tgt)
		}

		p.consume(token.Eq, "=")
		value, ok := p.parseExpr()
		if !ok {
			return ast.Statement{}, false
		}
		return ast.Statement{
			Node: ast.AssignStmt{Targets: targets, Value: value},
			Span: p.span(start),
		}, true
	}

	// Expression statement.
	return ast.Statement{Node: ast.ExprStmt{Value: firstExpr}, Span: p.span(start)}, true
}

func (p *Parser) exprToAssignTarget(expr ast.Expr) (ast.AssignTarget, bool) {
	path, ok := p.exprToAssignPath(expr)
	if !ok {
		return ast.AssignTarget{}, false
	}
	return ast.AssignTarget{Path: path}, true
}

// exprToAnnotatedTarget converts an expression to an assignment target and
// checks for a trailing `: Type` annotation (e.g. `result: MyStruct`).
func (p *Parser) exprToAnnotatedTarget(expr ast.Expr) (ast.AssignTarget, bool) {
	path, ok := p.exprToAssignPath(expr)
	if !ok {
		return ast.AssignTarget{}, false
	}
	target := ast.AssignTarget{Path: path}

	// Check for type annotation: `name: Type`
	if p.check(token.Colon) {
		p.advance()
		ty, ok := p.parseType()
		if !ok {
			return ast.AssignTarget{}, false
		}
		target.Ty = &ty
	}
	return target, true
}

func (p *Parser) exprToAssignPath(expr ast.Expr) (ast.AssignPath, bool) {
	switch e := expr.Node.(type) {
	case ast.IdentExpr:
		return ast.IdentPath{Name: ast.Spanned[string]{Node: e.Name, Span: expr.Span}}, true
	case ast.FieldExpr:
		obj, ok := p.exprToAssignPath(e.Object)
		if !ok {
			return nil, false
		}
		return ast.FieldPath{Object: obj, Field: e.Field}, true
	case ast.IndexExpr:
		obj, ok := p.exprToAssignPath(e.Object)
		if !ok {
			return nil, false
		}
		return ast.IndexPath{Object: obj, Index: e.Index}, true
	default:
		p.addError("expected assignable expression", expr.Span)
		return nil, false
	}
}

// ---------------------------------------------------------------------------
// If / For / While / Return / Try
// ---------------------------------------------------------------------------

type parsedIf struct {
	Condition  ast.Expr
	ThenBranch ast.Block
	ElseBranch ast.ElseBranch
}

func (p *Parser) parseIfStatement() (*parsedIf, bool) {
	cond, ok := p.parseExpr()
	if !ok {
		return nil, false
	}
	thenBranch, ok := p.parseBlock()
	if !ok {
		return nil, false
	}

	var elseBranch ast.ElseBranch
	if p.check(token.Else) {
		p.advance()
		if p.check(token.If) {
			p.advance()
			elseStart := p.previous.Start
			inner, ok := p.parseIfStatement()
			if !ok {
				return nil, false
			}
			innerStmt := ast.Spanned[ast.IfStmt]{
				Node: ast.IfStmt{
					Condition:  inner.Condition,
					ThenBranch: inner.ThenBranch,
					ElseBranch: inner.ElseBranch,
				},
				Span: p.span(elseStart),
			}
			elseBranch = &ast.ElseIf{If: innerStmt}
		} else {
			block, ok := p.parseBlock()
			if !ok {
				return nil, false
			}
			elseBranch = &ast.ElseBlock{Body: block}
		}
	}

	return &parsedIf{
		Condition:  cond,
		ThenBranch: thenBranch,
		ElseBranch: elseBranch,
	}, true
}

func (p *Parser) parseForStatement() (ast.ForStmt, bool) {
	pat, ok := p.parseForPattern()
	if !ok {
		return ast.ForStmt{}, false
	}
	p.consume(token.In, "in")
	iter, ok := p.parseExpr()
	if !ok {
		return ast.ForStmt{}, false
	}
	body, ok := p.parseBlock()
	if !ok {
		return ast.ForStmt{}, false
	}
	return ast.ForStmt{Pattern: pat, Iterator: iter, Body: body}, true
}

func (p *Parser) parseForPattern() (ast.ForPattern, bool) {
	first, ok := p.parseIdentifier()
	if !ok {
		return nil, false
	}
	if p.check(token.Comma) {
		p.advance()
		second, ok := p.parseIdentifier()
		if !ok {
			return nil, false
		}
		return ast.PairPattern{First: first, Second: second}, true
	}
	return ast.SinglePattern{Name: first}, true
}

func (p *Parser) parseWhileStatement() (ast.WhileStmt, bool) {
	cond, ok := p.parseExpr()
	if !ok {
		return ast.WhileStmt{}, false
	}
	body, ok := p.parseBlock()
	if !ok {
		return ast.WhileStmt{}, false
	}
	return ast.WhileStmt{Condition: cond, Body: body}, true
}

func (p *Parser) parseReturnStatement() (ast.ReturnStmt, bool) {
	// No value if followed by newline, }, or EOF.
	if p.check(token.Newline) || p.check(token.RBrace) || p.atEnd() {
		return ast.ReturnStmt{}, true
	}
	// Also handle the case where the raw token stream has a newline that
	// skipTrivia already moved past — check if there's nothing meaningful.
	// Actually since advance skips trivia, if we're here the current token
	// is meaningful, so parse values.
	var values []ast.Expr
	first, ok := p.parseExpr()
	if !ok {
		return ast.ReturnStmt{}, false
	}
	values = append(values, first)
	for p.check(token.Comma) {
		p.advance()
		v, ok := p.parseExpr()
		if !ok {
			return ast.ReturnStmt{}, false
		}
		values = append(values, v)
	}
	return ast.ReturnStmt{Values: values}, true
}

func (p *Parser) parseStepStatement() (ast.StepStmt, bool) {
	return p.parseStepStatementWithDecorators(nil)
}

func (p *Parser) parseStepStatementWithDecorators(decorators []ast.Decorator) (ast.StepStmt, bool) {
	// Expect a string literal for the step name
	if p.peek().Kind != token.String && p.peek().Kind != token.InterpolatedString {
		tok := p.peek()
		p.addError("expected string literal for step name", ast.Span{Start: tok.Start, End: tok.End})
		return ast.StepStmt{}, false
	}
	name := p.peek().Value
	nameSpan := p.currentSpan()
	p.advance()

	// Parse block body
	p.consume(token.LBrace, "{")
	p.skipNewlines()

	// Parse lifecycle hooks at top of step body
	var hooks []ast.LifecycleHook
	for p.check(token.Onerror) || p.check(token.Onsuccess) || p.check(token.Oncancel) {
		hook, ok := p.parseLifecycleHook()
		if ok {
			hooks = append(hooks, hook)
		}
		p.skipNewlines()
	}

	var stmts []ast.Statement
	for !p.check(token.RBrace) && !p.atEnd() {
		stmt, ok := p.parseStatement()
		if ok {
			stmts = append(stmts, stmt)
		} else {
			p.advance()
		}
		p.skipNewlines()
	}
	p.consume(token.RBrace, "}")

	return ast.StepStmt{
		Name:       ast.Spanned[string]{Node: name, Span: nameSpan},
		Body:       stmts,
		Decorators: decorators,
		Hooks:      hooks,
	}, true
}

func (p *Parser) parseLifecycleHook() (ast.LifecycleHook, bool) {
	hook := ast.LifecycleHook{}

	switch p.peek().Kind {
	case token.Onerror:
		p.advance()
		hook.Kind = ast.HookOnerror
		// Parse optional error variable name (accept `err` keyword as variable name)
		if p.check(token.Ident) || p.check(token.Err) {
			hook.ErrName = p.peek().Value
			p.advance()
		} else {
			hook.ErrName = "err"
		}
	case token.Onsuccess:
		p.advance()
		hook.Kind = ast.HookOnsuccess
		// Parse optional result variable name
		if p.check(token.Ident) {
			hook.ArgName = p.peek().Value
			p.advance()
		}
	case token.Oncancel:
		p.advance()
		hook.Kind = ast.HookOncancel
	default:
		return ast.LifecycleHook{}, false
	}

	body, ok := p.parseBlock()
	if !ok {
		return ast.LifecycleHook{}, false
	}
	hook.Body = body

	return hook, true
}

func (p *Parser) parseTryStatement() (ast.TryStmt, bool) {
	body, ok := p.parseBlock()
	if !ok {
		return ast.TryStmt{}, false
	}
	p.consume(token.Catch, "catch")
	errName, ok := p.parseIdentifier()
	if !ok {
		return ast.TryStmt{}, false
	}
	catchBody, ok := p.parseBlock()
	if !ok {
		return ast.TryStmt{}, false
	}
	return ast.TryStmt{Body: body, ErrorName: errName, CatchBody: catchBody}, true
}

// ---------------------------------------------------------------------------
// Blocks
// ---------------------------------------------------------------------------

func (p *Parser) parseBlock() (ast.Block, bool) {
	start := p.peek().Start
	p.consume(token.LBrace, "{")
	p.skipNewlines()

	var stmts []ast.Statement
	for !p.check(token.RBrace) && !p.atEnd() {
		stmt, ok := p.parseStatement()
		if ok {
			stmts = append(stmts, stmt)
		} else {
			p.advance() // error recovery
		}
		p.skipNewlines()
	}

	p.consume(token.RBrace, "}")
	return ast.Block{Statements: stmts, Span: p.span(start)}, true
}

// ---------------------------------------------------------------------------
// Expressions — Pratt parser
// ---------------------------------------------------------------------------

func (p *Parser) parseExpr() (ast.Expr, bool) {
	return p.parseExprPrecedence(PrecNone)
}

func (p *Parser) parseExprRest(left ast.Expr) (ast.Expr, bool) {
	return p.parseExprRestPrecedence(left, PrecNone)
}

func (p *Parser) parseExprRestPrecedence(left ast.Expr, minPrec Precedence) (ast.Expr, bool) {
	cur := left
	for !p.atEnd() {
		prec := precedenceOf(p.peek().Kind)
		if prec <= minPrec {
			break
		}
		next, ok := p.parseInfix(cur, prec)
		if !ok {
			return cur, true // return what we have
		}
		cur = next
	}
	return cur, true
}

func (p *Parser) parseExprPrecedence(minPrec Precedence) (ast.Expr, bool) {
	left, ok := p.parsePrefix()
	if !ok {
		return ast.Expr{}, false
	}
	return p.parseExprRestPrecedence(left, minPrec)
}

// ---------------------------------------------------------------------------
// Prefix (primary) expressions
// ---------------------------------------------------------------------------

func (p *Parser) parsePrefix() (ast.Expr, bool) {
	start := p.peek().Start

	switch p.peek().Kind {
	// Integer literal
	case token.Int:
		raw := p.peek().Value
		p.advance()
		val := parseInt(raw)
		return ast.Expr{
			Node: ast.LiteralExpr{Lit: ast.IntLit{Value: val}},
			Span: p.span(start),
		}, true

	// Float literal
	case token.Float:
		raw := p.peek().Value
		p.advance()
		val := parseFloat(raw)
		return ast.Expr{
			Node: ast.LiteralExpr{Lit: ast.FloatLit{Value: val}},
			Span: p.span(start),
		}, true

	// String literal
	case token.String:
		s := p.peek().Value
		p.advance()
		return ast.Expr{
			Node: ast.LiteralExpr{Lit: ast.StringLit{Value: s}},
			Span: p.span(start),
		}, true

	// Interpolated string
	case token.InterpolatedString:
		raw := p.peek().Value
		p.advance()
		parts := p.parseInterpolatedStringParts(raw)
		return ast.Expr{
			Node: ast.LiteralExpr{Lit: ast.InterpolatedStringLit{Parts: parts}},
			Span: p.span(start),
		}, true

	// Triple-quoted string — may or may not contain interpolation
	case token.TripleQuoteString:
		raw := p.peek().Value
		p.advance()
		if strings.Contains(raw, "${") {
			parts := p.parseInterpolatedStringParts(raw)
			return ast.Expr{
				Node: ast.LiteralExpr{Lit: ast.InterpolatedStringLit{Parts: parts, TripleQuoted: true}},
				Span: p.span(start),
			}, true
		}
		return ast.Expr{
			Node: ast.LiteralExpr{Lit: ast.StringLit{Value: raw, TripleQuoted: true}},
			Span: p.span(start),
		}, true

	// Boolean true
	case token.True:
		p.advance()
		return ast.Expr{
			Node: ast.LiteralExpr{Lit: ast.BoolLit{Value: true}},
			Span: p.span(start),
		}, true

	// Boolean false
	case token.False:
		p.advance()
		return ast.Expr{
			Node: ast.LiteralExpr{Lit: ast.BoolLit{Value: false}},
			Span: p.span(start),
		}, true

	// none / nil
	case token.None, token.Nil:
		p.advance()
		return ast.Expr{Node: ast.NoneExpr{}, Span: p.span(start)}, true

	// some(expr)
	case token.Some:
		p.advance()
		p.consume(token.LParen, "(")
		inner, ok := p.parseExpr()
		if !ok {
			return ast.Expr{}, false
		}
		p.consume(token.RParen, ")")
		return ast.Expr{Node: ast.SomeExpr{Inner: inner}, Span: p.span(start)}, true

	// Identifier (possibly lambda: x => ...)
	case token.Ident:
		name := p.peek().Value
		p.advance()

		// Arrow lambda: x => expr
		if p.check(token.FatArrow) {
			p.advance()
			body, ok := p.parseExpr()
			if !ok {
				return ast.Expr{}, false
			}
			param := ast.Param{
				Name: ast.Spanned[string]{Node: name, Span: p.span(start)},
				Span: p.span(start),
			}
			return ast.Expr{
				Node: ast.LambdaExpr{
					Params: []ast.Param{param},
					Body:   ast.LambdaExprBody{Value: body},
				},
				Span: p.span(start),
			}, true
		}

		// Instance creation: User { name = "Alice" }
		firstChar := firstRune(name)
		if unicode.IsUpper(firstChar) && p.check(token.LBrace) {
			return p.parseInstance(name, start)
		}

		return ast.Expr{Node: ast.IdentExpr{Name: name}, Span: p.span(start)}, true

	// Unary minus
	case token.Minus:
		p.advance()
		operand, ok := p.parseExprPrecedence(PrecUnary)
		if !ok {
			return ast.Expr{}, false
		}
		return ast.Expr{
			Node: ast.UnaryExpr{
				Op:      ast.Spanned[ast.UnaryOp]{Node: ast.OpNeg, Span: p.span(start)},
				Operand: operand,
			},
			Span: p.span(start),
		}, true

	// Bitwise NOT (~)
	case token.Tilde:
		p.advance()
		operand, ok := p.parseExprPrecedence(PrecUnary)
		if !ok {
			return ast.Expr{}, false
		}
		return ast.Expr{
			Node: ast.UnaryExpr{
				Op:      ast.Spanned[ast.UnaryOp]{Node: ast.OpBitNot, Span: p.span(start)},
				Operand: operand,
			},
			Span: p.span(start),
		}, true

	// Unary not
	case token.Not:
		p.advance()
		operand, ok := p.parseExprPrecedence(PrecUnary)
		if !ok {
			return ast.Expr{}, false
		}
		return ast.Expr{
			Node: ast.UnaryExpr{
				Op:      ast.Spanned[ast.UnaryOp]{Node: ast.OpNot, Span: p.span(start)},
				Operand: operand,
			},
			Span: p.span(start),
		}, true

	// Parenthesized expression or lambda
	case token.LParen:
		return p.parseParenOrLambda(start)

	// List literal
	case token.LBracket:
		return p.parseList(start)

	// Map or block expression
	case token.LBrace:
		return p.parseMapOrBlock(start)

	// If expression
	case token.If:
		p.advance()
		ifRes, ok := p.parseIfStatement()
		if !ok {
			return ast.Expr{}, false
		}
		return ast.Expr{
			Node: ast.IfExpr{If: ast.IfStmt{
				Condition:  ifRes.Condition,
				ThenBranch: ifRes.ThenBranch,
				ElseBranch: ifRes.ElseBranch,
			}},
			Span: p.span(start),
		}, true

	// Match expression
	case token.Match:
		p.advance()
		me, ok := p.parseMatchExpr()
		if !ok {
			return ast.Expr{}, false
		}
		return ast.Expr{Node: me, Span: p.span(start)}, true

	// Async block
	case token.Async:
		p.advance()
		block, ok := p.parseBlock()
		if !ok {
			return ast.Expr{}, false
		}
		return ast.Expr{Node: ast.AsyncExpr{Body: block}, Span: p.span(start)}, true

	// Spawn block
	case token.Spawn:
		p.advance()
		block, ok := p.parseBlock()
		if !ok {
			return ast.Expr{}, false
		}
		return ast.Expr{Node: ast.SpawnExpr{Body: block}, Span: p.span(start)}, true

	// Select expression
	case token.Select:
		p.advance()
		sel, ok := p.parseSelectExpr()
		if !ok {
			return ast.Expr{}, false
		}
		return ast.Expr{Node: sel, Span: p.span(start)}, true

	// Contextual keywords usable as identifiers in expressions
	case token.Err:
		p.advance()
		return ast.Expr{Node: ast.IdentExpr{Name: "err"}, Span: p.span(start)}, true

	case token.Ok:
		p.advance()
		return ast.Expr{Node: ast.IdentExpr{Name: "ok"}, Span: p.span(start)}, true

	default:
		tok := p.peek()
		p.addError("expected expression", ast.Span{Start: tok.Start, End: tok.End})
		return ast.Expr{}, false
	}
}

// ---------------------------------------------------------------------------
// Infix expressions
// ---------------------------------------------------------------------------

func (p *Parser) parseInfix(left ast.Expr, prec Precedence) (ast.Expr, bool) {
	start := left.Span.Start
	opSpan := p.currentSpan()

	switch p.peek().Kind {
	// Binary operators
	case token.Plus, token.Minus, token.Star, token.Slash, token.Percent,
		token.EqEq, token.Ne, token.Lt, token.Gt, token.Le, token.Ge,
		token.And, token.Or,
		token.Amp, token.Caret, token.Shl, token.Shr:
		op, ok := p.parseBinaryOp()
		if !ok {
			return left, true
		}
		right, ok := p.parseExprPrecedence(prec)
		if !ok {
			return ast.Expr{}, false
		}
		return ast.Expr{
			Node: ast.BinaryExpr{
				Left:  left,
				Op:    ast.Spanned[ast.BinaryOp]{Node: op, Span: opSpan},
				Right: right,
			},
			Span: p.span(start),
		}, true

	// Bitwise OR (|) — handled as binary op
	case token.Pipe:
		p.advance()
		right, ok := p.parseExprPrecedence(prec)
		if !ok {
			return ast.Expr{}, false
		}
		return ast.Expr{
			Node: ast.BinaryExpr{
				Left:  left,
				Op:    ast.Spanned[ast.BinaryOp]{Node: ast.OpBitOr, Span: opSpan},
				Right: right,
			},
			Span: p.span(start),
		}, true

	// Pipe (|>)
	case token.PipeArrow:
		p.advance()
		right, ok := p.parseExprPrecedence(prec)
		if !ok {
			return ast.Expr{}, false
		}
		return ast.Expr{
			Node: ast.PipeExpr{Left: left, Right: right},
			Span: p.span(start),
		}, true

	// Range ..
	case token.DotDot:
		p.advance()
		end, ok := p.parseExprPrecedence(prec)
		if !ok {
			return ast.Expr{}, false
		}
		return ast.Expr{
			Node: ast.RangeExpr{Start: left, End: end, Inclusive: false},
			Span: p.span(start),
		}, true

	// Range ..=
	case token.DotDotEq:
		p.advance()
		end, ok := p.parseExprPrecedence(prec)
		if !ok {
			return ast.Expr{}, false
		}
		return ast.Expr{
			Node: ast.RangeExpr{Start: left, End: end, Inclusive: true},
			Span: p.span(start),
		}, true

	// Call: expr(args)
	case token.LParen:
		args, ok := p.parseCallArgs()
		if !ok {
			return ast.Expr{}, false
		}
		return ast.Expr{
			Node: ast.CallExpr{Callee: left, Args: args},
			Span: p.span(start),
		}, true

	// Index: expr[index]
	case token.LBracket:
		p.advance()
		idx, ok := p.parseExpr()
		if !ok {
			return ast.Expr{}, false
		}
		p.consume(token.RBracket, "]")
		return ast.Expr{
			Node: ast.IndexExpr{Object: left, Index: idx},
			Span: p.span(start),
		}, true

	// Field access or method call: expr.field or expr.method(args)
	case token.Dot:
		p.advance()
		field, ok := p.parseIdentifier()
		if !ok {
			return ast.Expr{}, false
		}
		if p.check(token.LParen) {
			args, ok := p.parseCallArgs()
			if !ok {
				return ast.Expr{}, false
			}
			return ast.Expr{
				Node: ast.MethodCallExpr{Receiver: left, Method: field, Args: args},
				Span: p.span(start),
			}, true
		}
		// Qualified instance: ui.StatusCard { ... }
		if unicode.IsUpper(firstRune(field.Node)) && p.check(token.LBrace) {
			if ident, ok := left.Node.(ast.IdentExpr); ok {
				qualName := ident.Name + "." + field.Node
				return p.parseInstance(qualName, start)
			}
		}

		return ast.Expr{
			Node: ast.FieldExpr{Object: left, Field: field},
			Span: p.span(start),
		}, true

	// Error propagation: expr?
	case token.Question:
		p.advance()
		return ast.Expr{
			Node: ast.PropagateExpr{Inner: left},
			Span: p.span(start),
		}, true

	// Orelse: expr orelse default
	case token.Orelse:
		p.advance()
		def, ok := p.parseExprPrecedence(prec)
		if !ok {
			return ast.Expr{}, false
		}
		return ast.Expr{
			Node: ast.OrelseExpr{Left: left, Default: def},
			Span: p.span(start),
		}, true

	default:
		return left, true
	}
}

func (p *Parser) parseBinaryOp() (ast.BinaryOp, bool) {
	var op ast.BinaryOp
	switch p.peek().Kind {
	case token.Plus:
		op = ast.OpAdd
	case token.Minus:
		op = ast.OpSub
	case token.Star:
		op = ast.OpMul
	case token.Slash:
		op = ast.OpDiv
	case token.Percent:
		op = ast.OpMod
	case token.EqEq:
		op = ast.OpEq
	case token.Ne:
		op = ast.OpNe
	case token.Lt:
		op = ast.OpLt
	case token.Gt:
		op = ast.OpGt
	case token.Le:
		op = ast.OpLe
	case token.Ge:
		op = ast.OpGe
	case token.And:
		op = ast.OpAnd
	case token.Or:
		op = ast.OpOr
	case token.Amp:
		op = ast.OpBitAnd
	case token.Caret:
		op = ast.OpBitXor
	case token.Shl:
		op = ast.OpShl
	case token.Shr:
		op = ast.OpShr
	default:
		return 0, false
	}
	p.advance()
	return op, true
}

// ---------------------------------------------------------------------------
// Call arguments
// ---------------------------------------------------------------------------

func (p *Parser) parseCallArgs() ([]ast.Argument, bool) {
	p.consume(token.LParen, "(")

	var args []ast.Argument
	for !p.check(token.RParen) && !p.atEnd() {
		argStart := p.peek().Start

		// Check for named argument: name = value or name: value
		if p.peek().Kind == token.Ident {
			ident, ok := p.parseIdentifier()
			if !ok {
				return nil, false
			}

			if p.check(token.Eq) || p.check(token.Colon) {
				p.advance()
				val, ok := p.parseExpr()
				if !ok {
					return nil, false
				}
				args = append(args, ast.Argument{
					Name:  &ident,
					Value: val,
					Span:  p.span(argStart),
				})
				if !p.check(token.RParen) {
					p.consume(token.Comma, ",")
				}
				continue
			}

			// Not a named arg — build an expression from the identifier and
			// continue parsing the rest of the expression.
			identExpr := ast.Expr{Node: ast.IdentExpr{Name: ident.Node}, Span: ident.Span}
			val, ok := p.parseExprRest(identExpr)
			if !ok {
				return nil, false
			}
			args = append(args, ast.Argument{
				Value: val,
				Span:  p.span(argStart),
			})
			if !p.check(token.RParen) {
				p.consume(token.Comma, ",")
			}
			continue
		}

		// Positional argument.
		val, ok := p.parseExpr()
		if !ok {
			return nil, false
		}
		args = append(args, ast.Argument{
			Value: val,
			Span:  p.span(argStart),
		})
		if !p.check(token.RParen) {
			p.consume(token.Comma, ",")
		}
	}

	p.consume(token.RParen, ")")
	return args, true
}

// ---------------------------------------------------------------------------
// Parenthesized expr / lambda
// ---------------------------------------------------------------------------

func (p *Parser) parseParenOrLambda(start int) (ast.Expr, bool) {
	p.advance() // consume (

	// Empty parens: () => expr  or  () { block }  or  unit
	if p.check(token.RParen) {
		p.advance()

		if p.check(token.FatArrow) {
			p.advance()
			body, ok := p.parseExpr()
			if !ok {
				return ast.Expr{}, false
			}
			return ast.Expr{
				Node: ast.LambdaExpr{
					Body: ast.LambdaExprBody{Value: body},
				},
				Span: p.span(start),
			}, true
		}

		if p.check(token.LBrace) {
			block, ok := p.parseBlock()
			if !ok {
				return ast.Expr{}, false
			}
			return ast.Expr{
				Node: ast.LambdaExpr{
					Body: ast.LambdaBlockBody{Value: block},
				},
				Span: p.span(start),
			}, true
		}

		// Empty parens as unit — return empty list for now.
		return ast.Expr{
			Node: ast.ListExpr{},
			Span: p.span(start),
		}, true
	}

	// Parse first expression inside parens.
	first, ok := p.parseExpr()
	if !ok {
		return ast.Expr{}, false
	}

	// Comma or colon after first expression => lambda parameter list.
	if p.check(token.Comma) || p.check(token.Colon) {
		params := []ast.Param{}
		fp, ok := p.exprToParam(first)
		if !ok {
			return ast.Expr{}, false
		}
		params = append(params, fp)

		for p.check(token.Comma) {
			p.advance()
			expr, ok := p.parseExpr()
			if !ok {
				return ast.Expr{}, false
			}
			param, ok := p.exprToParam(expr)
			if !ok {
				return ast.Expr{}, false
			}
			params = append(params, param)
		}

		p.consume(token.RParen, ")")

		if p.check(token.FatArrow) {
			p.advance()
			body, ok := p.parseExpr()
			if !ok {
				return ast.Expr{}, false
			}
			return ast.Expr{
				Node: ast.LambdaExpr{Params: params, Body: ast.LambdaExprBody{Value: body}},
				Span: p.span(start),
			}, true
		}

		block, ok := p.parseBlock()
		if !ok {
			return ast.Expr{}, false
		}
		return ast.Expr{
			Node: ast.LambdaExpr{Params: params, Body: ast.LambdaBlockBody{Value: block}},
			Span: p.span(start),
		}, true
	}

	p.consume(token.RParen, ")")

	// Single-param lambda: (x) => expr  or  (x) { block }
	if p.check(token.FatArrow) {
		p.advance()
		param, ok := p.exprToParam(first)
		if !ok {
			return ast.Expr{}, false
		}
		body, ok := p.parseExpr()
		if !ok {
			return ast.Expr{}, false
		}
		return ast.Expr{
			Node: ast.LambdaExpr{Params: []ast.Param{param}, Body: ast.LambdaExprBody{Value: body}},
			Span: p.span(start),
		}, true
	}

	if p.check(token.LBrace) {
		param, ok := p.exprToParam(first)
		if !ok {
			return ast.Expr{}, false
		}
		block, ok := p.parseBlock()
		if !ok {
			return ast.Expr{}, false
		}
		return ast.Expr{
			Node: ast.LambdaExpr{Params: []ast.Param{param}, Body: ast.LambdaBlockBody{Value: block}},
			Span: p.span(start),
		}, true
	}

	// Just a parenthesized expression.
	return ast.Expr{
		Node: ast.ParenExpr{Inner: first},
		Span: p.span(start),
	}, true
}

func (p *Parser) exprToParam(expr ast.Expr) (ast.Param, bool) {
	if ident, ok := expr.Node.(ast.IdentExpr); ok {
		return ast.Param{
			Name: ast.Spanned[string]{Node: ident.Name, Span: expr.Span},
			Span: expr.Span,
		}, true
	}
	p.addError("expected identifier for parameter", expr.Span)
	return ast.Param{}, false
}

// ---------------------------------------------------------------------------
// List literal
// ---------------------------------------------------------------------------

func (p *Parser) parseList(start int) (ast.Expr, bool) {
	p.advance() // consume [

	var elems []ast.Expr
	for !p.check(token.RBracket) && !p.atEnd() {
		e, ok := p.parseExpr()
		if !ok {
			return ast.Expr{}, false
		}
		elems = append(elems, e)
		if !p.check(token.RBracket) {
			p.consume(token.Comma, ",")
		}
	}

	p.consume(token.RBracket, "]")
	return ast.Expr{Node: ast.ListExpr{Elems: elems}, Span: p.span(start)}, true
}

// ---------------------------------------------------------------------------
// Map or block expression
// ---------------------------------------------------------------------------

func (p *Parser) parseMapOrBlock(start int) (ast.Expr, bool) {
	p.advance() // consume {
	p.skipNewlines()

	// Empty braces → empty map
	if p.check(token.RBrace) {
		p.advance()
		return ast.Expr{Node: ast.MapExpr{}, Span: p.span(start)}, true
	}

	// Parse first expression and see if it's followed by a colon (map) or not (block).
	firstExpr, ok := p.parseExpr()
	if !ok {
		return ast.Expr{}, false
	}

	if p.check(token.Colon) {
		// Map literal.
		p.advance()
		firstValue, ok := p.parseExpr()
		if !ok {
			return ast.Expr{}, false
		}
		entries := []ast.MapEntry{{Key: firstExpr, Value: firstValue}}

		for p.check(token.Comma) {
			p.advance()
			p.skipNewlines()
			if p.check(token.RBrace) {
				break
			}
			k, ok := p.parseExpr()
			if !ok {
				return ast.Expr{}, false
			}
			p.consume(token.Colon, ":")
			v, ok := p.parseExpr()
			if !ok {
				return ast.Expr{}, false
			}
			entries = append(entries, ast.MapEntry{Key: k, Value: v})
		}

		p.skipNewlines()
		p.consume(token.RBrace, "}")
		return ast.Expr{Node: ast.MapExpr{Entries: entries}, Span: p.span(start)}, true
	}

	// Block expression.
	firstStmt := ast.Statement{Node: ast.ExprStmt{Value: firstExpr}, Span: p.span(start)}
	stmts := []ast.Statement{firstStmt}

	p.skipNewlines()
	for !p.check(token.RBrace) && !p.atEnd() {
		stmt, ok := p.parseStatement()
		if ok {
			stmts = append(stmts, stmt)
		} else {
			p.advance()
		}
		p.skipNewlines()
	}

	p.consume(token.RBrace, "}")
	return ast.Expr{
		Node: ast.BlockExpr{Body: ast.Block{Statements: stmts, Span: p.span(start)}},
		Span: p.span(start),
	}, true
}

// ---------------------------------------------------------------------------
// Instance creation: User { name = "Alice" }
// ---------------------------------------------------------------------------

func (p *Parser) parseInstance(typeName string, start int) (ast.Expr, bool) {
	p.advance() // consume {
	p.skipNewlines()

	var fields []ast.InstanceField
	for !p.check(token.RBrace) && !p.atEnd() {
		fieldStart := p.peek().Start

		// Check for named field: name = value
		if p.peek().Kind == token.Ident {
			ident, ok := p.parseIdentifier()
			if !ok {
				return ast.Expr{}, false
			}

			if p.check(token.Eq) {
				p.advance()
				val, ok := p.parseExpr()
				if !ok {
					return ast.Expr{}, false
				}
				fields = append(fields, ast.InstanceField{
					Name:  &ident,
					Value: val,
					Span:  p.span(fieldStart),
				})
			} else {
				// Positional/shorthand: just the identifier as value.
				fields = append(fields, ast.InstanceField{
					Value: ast.Expr{Node: ast.IdentExpr{Name: ident.Node}, Span: ident.Span},
					Span:  p.span(fieldStart),
				})
			}
		} else {
			val, ok := p.parseExpr()
			if !ok {
				return ast.Expr{}, false
			}
			fields = append(fields, ast.InstanceField{
				Value: val,
				Span:  p.span(fieldStart),
			})
		}

		if !p.check(token.RBrace) {
			if p.check(token.Comma) {
				p.advance()
			}
			p.skipNewlines()
		}
	}

	p.consume(token.RBrace, "}")
	nameSpanned := ast.Spanned[string]{Node: typeName, Span: p.span(start)}
	return ast.Expr{
		Node: ast.InstanceExpr{TypeName: nameSpanned, Fields: fields},
		Span: p.span(start),
	}, true
}

// ---------------------------------------------------------------------------
// Match expression
// ---------------------------------------------------------------------------

func (p *Parser) parseMatchExpr() (ast.MatchExpr, bool) {
	subject, ok := p.parseExpr()
	if !ok {
		return ast.MatchExpr{}, false
	}
	p.consume(token.LBrace, "{")
	p.skipNewlines()

	var arms []ast.MatchArm
	for !p.check(token.RBrace) && !p.atEnd() {
		arm, ok := p.parseMatchArm()
		if ok {
			arms = append(arms, arm)
		} else {
			p.advance() // error recovery
		}
		p.skipNewlines()
	}

	p.consume(token.RBrace, "}")
	return ast.MatchExpr{Subject: subject, Arms: arms}, true
}

func (p *Parser) parseMatchArm() (ast.MatchArm, bool) {
	start := p.peek().Start
	pat, ok := p.parsePattern()
	if !ok {
		return ast.MatchArm{}, false
	}

	// Optional guard: if condition
	var guard *ast.Expr
	if p.check(token.If) {
		p.advance()
		g, ok := p.parseExpr()
		if !ok {
			return ast.MatchArm{}, false
		}
		guard = &g
	}

	p.consume(token.FatArrow, "=>")

	var body ast.MatchArmBody
	if p.check(token.LBrace) {
		block, ok := p.parseBlock()
		if !ok {
			return ast.MatchArm{}, false
		}
		body = ast.MatchArmBlock{Value: block}
	} else if p.peek().Kind == token.Return || p.peek().Kind == token.Break || p.peek().Kind == token.Continue {
		stmt, ok := p.parseStatement()
		if !ok {
			return ast.MatchArm{}, false
		}
		body = ast.MatchArmBlock{Value: ast.Block{Statements: []ast.Statement{stmt}, Span: stmt.Span}}
	} else {
		expr, ok := p.parseExpr()
		if !ok {
			return ast.MatchArm{}, false
		}
		body = ast.MatchArmExpr{Value: expr}
	}

	return ast.MatchArm{
		Pattern: pat,
		Guard:   guard,
		Body:    body,
		Span:    p.span(start),
	}, true
}

func (p *Parser) parsePattern() (ast.Spanned[ast.Pattern], bool) {
	start := p.peek().Start
	first, ok := p.parseSinglePattern()
	if !ok {
		return ast.Spanned[ast.Pattern]{}, false
	}

	// Or-pattern: A | B | C
	if p.check(token.Pipe) {
		patterns := []ast.Spanned[ast.Pattern]{first}
		for p.check(token.Pipe) {
			p.advance()
			next, ok := p.parseSinglePattern()
			if !ok {
				return ast.Spanned[ast.Pattern]{}, false
			}
			patterns = append(patterns, next)
		}
		return ast.Spanned[ast.Pattern]{
			Node: ast.OrPattern{Patterns: patterns},
			Span: p.span(start),
		}, true
	}

	return first, true
}

func (p *Parser) parseSinglePattern() (ast.Spanned[ast.Pattern], bool) {
	start := p.peek().Start

	switch p.peek().Kind {
	case token.Ident:
		name := p.peek().Value
		p.advance()

		if name == "_" {
			return ast.Spanned[ast.Pattern]{Node: ast.WildcardPattern{}, Span: p.span(start)}, true
		}

		// Dotted pattern: Direction.North → "DirectionNorth"
		if p.check(token.Dot) {
			p.advance()
			field, ok := p.parseIdentifier()
			if !ok {
				return ast.Spanned[ast.Pattern]{}, false
			}
			combined := name + field.Node
			return ast.Spanned[ast.Pattern]{Node: ast.IdentPattern{Name: combined}, Span: p.span(start)}, true
		}

		// Constructor pattern: Some { value }
		if p.check(token.LBrace) {
			p.advance()
			var fields []ast.Spanned[string]
			for !p.check(token.RBrace) && !p.atEnd() {
				f, ok := p.parseIdentifier()
				if !ok {
					return ast.Spanned[ast.Pattern]{}, false
				}
				fields = append(fields, f)
				if !p.check(token.RBrace) {
					p.consume(token.Comma, ",")
				}
			}
			p.consume(token.RBrace, "}")
			return ast.Spanned[ast.Pattern]{
				Node: ast.ConstructorPattern{Name: name, Fields: fields},
				Span: p.span(start),
			}, true
		}

		return ast.Spanned[ast.Pattern]{Node: ast.IdentPattern{Name: name}, Span: p.span(start)}, true

	case token.Int:
		raw := p.peek().Value
		p.advance()
		val := parseInt(raw)

		// Range pattern: 1..5 or 1..=5
		if p.check(token.DotDot) || p.check(token.DotDotEq) {
			inclusive := p.check(token.DotDotEq)
			p.advance()
			if !p.check(token.Int) {
				p.addError("expected integer in range pattern", p.currentSpan())
				return ast.Spanned[ast.Pattern]{}, false
			}
			endRaw := p.peek().Value
			p.advance()
			endVal := parseInt(endRaw)
			return ast.Spanned[ast.Pattern]{
				Node: ast.RangePattern{
					Start:     ast.IntLit{Value: val},
					End:       ast.IntLit{Value: endVal},
					Inclusive: inclusive,
				},
				Span: p.span(start),
			}, true
		}

		return ast.Spanned[ast.Pattern]{
			Node: ast.LiteralPattern{Lit: ast.IntLit{Value: val}},
			Span: p.span(start),
		}, true

	case token.String:
		s := p.peek().Value
		p.advance()
		return ast.Spanned[ast.Pattern]{
			Node: ast.LiteralPattern{Lit: ast.StringLit{Value: s}},
			Span: p.span(start),
		}, true

	case token.True:
		p.advance()
		return ast.Spanned[ast.Pattern]{
			Node: ast.LiteralPattern{Lit: ast.BoolLit{Value: true}},
			Span: p.span(start),
		}, true

	case token.False:
		p.advance()
		return ast.Spanned[ast.Pattern]{
			Node: ast.LiteralPattern{Lit: ast.BoolLit{Value: false}},
			Span: p.span(start),
		}, true

	default:
		tok := p.peek()
		p.addError("expected pattern", ast.Span{Start: tok.Start, End: tok.End})
		return ast.Spanned[ast.Pattern]{}, false
	}
}

// ---------------------------------------------------------------------------
// Select expression
// ---------------------------------------------------------------------------

func (p *Parser) parseSelectExpr() (ast.SelectExpr, bool) {
	p.consume(token.LBrace, "{")
	p.skipNewlines()

	var arms []ast.SelectArm
	var defaultBlock *ast.Block

	for !p.check(token.RBrace) && !p.atEnd() {
		if p.check(token.Default) {
			p.advance()
			p.consume(token.FatArrow, "=>")
			block, ok := p.parseBlock()
			if !ok {
				return ast.SelectExpr{}, false
			}
			defaultBlock = &block
		} else {
			arm, ok := p.parseSelectArm()
			if ok {
				arms = append(arms, arm)
			}
		}
		p.skipNewlines()
	}

	p.consume(token.RBrace, "}")
	return ast.SelectExpr{Arms: arms, Default: defaultBlock}, true
}

func (p *Parser) parseSelectArm() (ast.SelectArm, bool) {
	start := p.peek().Start
	binding, ok := p.parseIdentifier()
	if !ok {
		return ast.SelectArm{}, false
	}
	p.consume(token.From, "from")
	channel, ok := p.parseExpr()
	if !ok {
		return ast.SelectArm{}, false
	}
	p.consume(token.FatArrow, "=>")

	var body ast.MatchArmBody
	if p.check(token.LBrace) {
		block, ok := p.parseBlock()
		if !ok {
			return ast.SelectArm{}, false
		}
		body = ast.MatchArmBlock{Value: block}
	} else {
		expr, ok := p.parseExpr()
		if !ok {
			return ast.SelectArm{}, false
		}
		body = ast.MatchArmExpr{Value: expr}
	}

	return ast.SelectArm{
		Binding: binding,
		Channel: channel,
		Body:    body,
		Span:    p.span(start),
	}, true
}

// ---------------------------------------------------------------------------
// Agentic declarations
// ---------------------------------------------------------------------------

func (p *Parser) parseImportDecl() (ast.ImportDecl, bool) {
	switch p.peek().Kind {
	case token.String:
		// Basic: import "io"
		path := p.peek().Value
		p.advance()
		return ast.ImportDecl{Path: path}, true

	case token.LBrace:
		// Selective: import { User, Post } from "models"
		p.advance() // consume {
		p.skipNewlines()
		var names []ast.Spanned[string]
		for !p.check(token.RBrace) && !p.atEnd() {
			name, ok := p.parseIdentifier()
			if !ok {
				return ast.ImportDecl{}, false
			}
			names = append(names, name)
			if p.check(token.Comma) {
				p.advance()
			}
			p.skipNewlines()
		}
		p.consume(token.RBrace, "}")
		if !p.check(token.From) {
			tok := p.peek()
			p.addError("expected 'from' after import names", ast.Span{Start: tok.Start, End: tok.End})
			return ast.ImportDecl{}, false
		}
		p.advance() // consume from
		if p.peek().Kind != token.String {
			tok := p.peek()
			p.addError("expected string for import path", ast.Span{Start: tok.Start, End: tok.End})
			return ast.ImportDecl{}, false
		}
		path := p.peek().Value
		p.advance()
		return ast.ImportDecl{Path: path, Names: names}, true

	case token.Star:
		// Glob: import * from "math"
		p.advance() // consume *
		if !p.check(token.From) {
			tok := p.peek()
			p.addError("expected 'from' after '*'", ast.Span{Start: tok.Start, End: tok.End})
			return ast.ImportDecl{}, false
		}
		p.advance() // consume from
		if p.peek().Kind != token.String {
			tok := p.peek()
			p.addError("expected string for import path", ast.Span{Start: tok.Start, End: tok.End})
			return ast.ImportDecl{}, false
		}
		path := p.peek().Value
		p.advance()
		return ast.ImportDecl{Path: path, IsGlob: true}, true

	case token.Ident:
		// Aliased: import fmt from "io"
		alias, ok := p.parseIdentifier()
		if !ok {
			return ast.ImportDecl{}, false
		}
		if !p.check(token.From) {
			tok := p.peek()
			p.addError("expected 'from' after alias name", ast.Span{Start: tok.Start, End: tok.End})
			return ast.ImportDecl{}, false
		}
		p.advance() // consume from
		if p.peek().Kind != token.String {
			tok := p.peek()
			p.addError("expected string for import path", ast.Span{Start: tok.Start, End: tok.End})
			return ast.ImportDecl{}, false
		}
		path := p.peek().Value
		p.advance()
		return ast.ImportDecl{Path: path, Alias: &alias}, true

	default:
		tok := p.peek()
		p.addError("expected string, identifier, '{', or '*' for import", ast.Span{Start: tok.Start, End: tok.End})
		return ast.ImportDecl{}, false
	}
}

func (p *Parser) parseExportDecl() (ast.ExportDecl, bool) {
	// export { User, Post }
	if !p.check(token.LBrace) {
		tok := p.peek()
		p.addError("expected '{' after export", ast.Span{Start: tok.Start, End: tok.End})
		return ast.ExportDecl{}, false
	}
	p.advance() // consume {
	p.skipNewlines()
	var names []ast.Spanned[string]
	for !p.check(token.RBrace) && !p.atEnd() {
		name, ok := p.parseIdentifier()
		if !ok {
			return ast.ExportDecl{}, false
		}
		names = append(names, name)
		if p.check(token.Comma) {
			p.advance()
		}
		p.skipNewlines()
	}
	p.consume(token.RBrace, "}")
	return ast.ExportDecl{Names: names}, true
}

func (p *Parser) parseProviderDecl() (ast.ProviderDecl, bool) {
	name, ok := p.parseIdentifier()
	if !ok {
		return ast.ProviderDecl{}, false
	}
	p.consume(token.LBrace, "{")
	p.skipNewlines()

	var fields []ast.ProviderField
	for !p.check(token.RBrace) && !p.atEnd() {
		key, ok := p.parseIdentifier()
		if !ok {
			break
		}
		p.consume(token.Colon, ":")
		val, ok := p.parseExpr()
		if !ok {
			break
		}
		fields = append(fields, ast.ProviderField{Key: key, Value: val})
		if p.check(token.Comma) {
			p.advance()
		}
		p.skipNewlines()
	}

	p.consume(token.RBrace, "}")
	return ast.ProviderDecl{Name: name, Fields: fields}, true
}

func (p *Parser) parseToolDecl() (ast.ToolDecl, bool) {
	name, ok := p.parseIdentifier()
	if !ok {
		return ast.ToolDecl{}, false
	}
	params, ok := p.parseParams()
	if !ok {
		return ast.ToolDecl{}, false
	}

	var retTy *ast.Spanned[ast.Type]
	if p.check(token.Arrow) {
		p.advance()
		rt, ok := p.parseType()
		if !ok {
			return ast.ToolDecl{}, false
		}
		retTy = &rt
	}

	p.consume(token.LBrace, "{")
	p.skipNewlines()

	// Expect triple-quoted description.
	var description string
	if p.peek().Kind == token.TripleQuoteString {
		description = p.peek().Value
		p.advance()
	} else {
		tok := p.peek()
		p.addError("expected triple-quoted description string", ast.Span{Start: tok.Start, End: tok.End})
		return ast.ToolDecl{}, false
	}

	p.skipNewlines()

	// Parse optional @before/@after hooks before the main body.
	var beforeHook, afterHook *ast.Block
	for p.check(token.At) {
		p.advance() // consume @
		tok := p.peek()
		if tok.Value == "before" || tok.Value == "after" {
			hookName := tok.Value
			p.advance() // consume "before"/"after"
			hookBlock, ok := p.parseBlock()
			if ok {
				if hookName == "before" {
					beforeHook = &hookBlock
				} else {
					afterHook = &hookBlock
				}
			}
			p.skipNewlines()
		} else {
			// Not a hook decorator — put @ back by not consuming further
			break
		}
	}

	// Optional body statements.
	var body *ast.Block
	if !p.check(token.RBrace) {
		blockStart := p.peek().Start
		var stmts []ast.Statement
		for !p.check(token.RBrace) && !p.atEnd() {
			stmt, ok := p.parseStatement()
			if ok {
				stmts = append(stmts, stmt)
			} else {
				p.advance()
			}
			p.skipNewlines()
		}
		b := ast.Block{Statements: stmts, Span: p.span(blockStart)}
		body = &b
	}

	p.consume(token.RBrace, "}")
	return ast.ToolDecl{
		Name:        name,
		Params:      params,
		ReturnTy:    retTy,
		Description: description,
		Body:        body,
		BeforeHook:  beforeHook,
		AfterHook:   afterHook,
	}, true
}

func (p *Parser) parseTestDecl() (ast.TestDecl, bool) {
	// Expect a string literal name: test "name" { ... }
	if p.peek().Kind != token.String {
		tok := p.peek()
		p.addError("expected test name string", ast.Span{Start: tok.Start, End: tok.End})
		return ast.TestDecl{}, false
	}
	nameTok := p.peek()
	name := ast.Spanned[string]{Node: nameTok.Value, Span: ast.Span{Start: nameTok.Start, End: nameTok.End}}
	p.advance()

	body, ok := p.parseBlock()
	if !ok {
		return ast.TestDecl{}, false
	}
	return ast.TestDecl{Name: name, Body: body}, true
}

func (p *Parser) parseAssertStmt() (ast.AssertStmt, bool) {
	cond, ok := p.parseExpr()
	if !ok {
		return ast.AssertStmt{}, false
	}

	var msg *ast.Expr
	if p.check(token.Comma) {
		p.advance()
		m, ok := p.parseExpr()
		if !ok {
			return ast.AssertStmt{}, false
		}
		msg = &m
	}
	return ast.AssertStmt{Condition: cond, Message: msg}, true
}

func (p *Parser) parseEvalDecl() (ast.EvalDecl, bool) {
	// Expect a string literal name: eval "name" { ... }
	if p.peek().Kind != token.String {
		tok := p.peek()
		p.addError("expected eval name string", ast.Span{Start: tok.Start, End: tok.End})
		return ast.EvalDecl{}, false
	}
	nameTok := p.peek()
	name := ast.Spanned[string]{Node: nameTok.Value, Span: ast.Span{Start: nameTok.Start, End: nameTok.End}}
	p.advance()

	p.consume(token.LBrace, "{")
	p.skipNewlines()

	var fields []ast.EvalField
	for !p.check(token.RBrace) && !p.atEnd() {
		key, ok := p.parseIdentifier()
		if !ok {
			p.advance()
			continue
		}
		p.consume(token.Colon, ":")
		val, ok := p.parseExpr()
		if !ok {
			p.advance()
			continue
		}
		fields = append(fields, ast.EvalField{Key: key, Value: val})
		if p.check(token.Comma) {
			p.advance()
		}
		p.skipNewlines()
	}
	p.consume(token.RBrace, "}")

	return ast.EvalDecl{Name: name, Fields: fields}, true
}

func (p *Parser) parseAgentDecl() (ast.AgentDecl, bool) {
	name, ok := p.parseIdentifier()
	if !ok {
		return ast.AgentDecl{}, false
	}
	p.consume(token.LBrace, "{")
	p.skipNewlines()

	var fields []ast.AgentField
	for !p.check(token.RBrace) && !p.atEnd() {
		key, ok := p.parseIdentifier()
		if !ok {
			break
		}
		p.consume(token.Colon, ":")
		val, ok := p.parseExpr()
		if !ok {
			break
		}
		fields = append(fields, ast.AgentField{Key: key, Value: val})
		if p.check(token.Comma) {
			p.advance()
		}
		p.skipNewlines()
	}

	p.consume(token.RBrace, "}")
	return ast.AgentDecl{Name: name, Fields: fields}, true
}

func (p *Parser) parseWorkflowDecl(trigger *ast.Decorator) (ast.WorkflowDecl, bool) {
	name, ok := p.parseIdentifier()
	if !ok {
		return ast.WorkflowDecl{}, false
	}
	params, ok := p.parseParams()
	if !ok {
		return ast.WorkflowDecl{}, false
	}

	var retTy *ast.Spanned[ast.Type]
	if p.check(token.Arrow) {
		p.advance()
		rt, ok := p.parseType()
		if !ok {
			return ast.WorkflowDecl{}, false
		}
		retTy = &rt
	}

	p.consume(token.LBrace, "{")
	p.skipNewlines()

	// Parse lifecycle hooks at top of workflow body
	var hooks []ast.LifecycleHook
	for p.check(token.Onerror) || p.check(token.Onsuccess) || p.check(token.Oncancel) {
		hook, ok := p.parseLifecycleHook()
		if ok {
			hooks = append(hooks, hook)
		}
		p.skipNewlines()
	}

	// Parse optional triple-quoted description (for MCP tool exposure)
	var description string
	if p.peek().Kind == token.TripleQuoteString {
		description = p.peek().Value
		p.advance()
		p.skipNewlines()
	}

	// Parse body statements
	var stmts []ast.Statement
	for !p.check(token.RBrace) && !p.atEnd() {
		// Check for @decorator before step
		if p.check(token.At) {
			decoStart := p.current
			var decorators []ast.Decorator
			for p.check(token.At) {
				dec, ok := p.parseDecorator()
				if ok {
					decorators = append(decorators, dec)
				}
				p.skipNewlines()
			}
			// After decorators, expect step
			if p.check(token.Step) {
				p.advance()
				ss, ok := p.parseStepStatementWithDecorators(decorators)
				if ok {
					stmts = append(stmts, ast.Statement{Node: ss, Span: p.span(decoStart)})
				}
			} else {
				// Not a step — treat decorators as error, re-parse as statement
				tok := p.peek()
				p.addError("decorator can only precede step inside workflow", ast.Span{Start: tok.Start, End: tok.End})
				p.advance()
			}
		} else {
			stmt, ok := p.parseStatement()
			if ok {
				stmts = append(stmts, stmt)
			} else {
				p.advance()
			}
		}
		p.skipNewlines()
	}
	p.consume(token.RBrace, "}")

	return ast.WorkflowDecl{
		Name:        name,
		Trigger:     trigger,
		Params:      params,
		ReturnTy:    retTy,
		Description: description,
		Body:        ast.Block{Statements: stmts},
		Hooks:       hooks,
	}, true
}

func (p *Parser) parseDecorator() (ast.Decorator, bool) {
	p.consume(token.At, "@")
	name, ok := p.parseIdentifier()
	if !ok {
		return ast.Decorator{}, false
	}

	var args []ast.Expr
	if p.check(token.LParen) {
		p.advance()
		for !p.check(token.RParen) && !p.atEnd() {
			argStart := p.current
			e, ok := p.parseExpr()
			if !ok {
				return ast.Decorator{}, false
			}
			// Handle named args: key: value → wrap as MapEntry-style MapExpr
			if p.check(token.Colon) {
				p.advance()
				val, ok := p.parseExpr()
				if !ok {
					return ast.Decorator{}, false
				}
				// Encode as a single-entry map: {key: value}
				entry := ast.MapExpr{
					Entries: []ast.MapEntry{{Key: e, Value: val}},
				}
				args = append(args, ast.Expr{Node: entry, Span: p.span(argStart)})
			} else {
				args = append(args, e)
			}
			if !p.check(token.RParen) {
				p.consume(token.Comma, ",")
			}
		}
		p.consume(token.RParen, ")")
	}

	p.skipNewlines()
	return ast.Decorator{Name: name, Args: args}, true
}

func (p *Parser) parseFnDecl(isPublic bool) (ast.FunctionDef, bool) {
	name, ok := p.parseIdentifier()
	if !ok {
		return ast.FunctionDef{}, false
	}
	params, ok := p.parseParams()
	if !ok {
		return ast.FunctionDef{}, false
	}

	var retTy *ast.Spanned[ast.Type]
	if p.check(token.Arrow) {
		p.advance()
		rt, ok := p.parseType()
		if !ok {
			return ast.FunctionDef{}, false
		}
		retTy = &rt
	}

	body, ok := p.parseBlock()
	if !ok {
		return ast.FunctionDef{}, false
	}

	return ast.FunctionDef{
		IsPublic: isPublic,
		Name:     name,
		Params:   params,
		ReturnTy: retTy,
		Body:     body,
	}, true
}

// ---------------------------------------------------------------------------
// Identifier helper
// ---------------------------------------------------------------------------

func (p *Parser) parseIdentifier() (ast.Spanned[string], bool) {
	tok := p.peek()
	switch tok.Kind {
	case token.Ident:
		name := tok.Value
		span := p.currentSpan()
		p.advance()
		return ast.Spanned[string]{Node: name, Span: span}, true
	// Contextual keywords usable as identifiers.
	case token.Err:
		span := p.currentSpan()
		p.advance()
		return ast.Spanned[string]{Node: "err", Span: span}, true
	case token.Ok:
		span := p.currentSpan()
		p.advance()
		return ast.Spanned[string]{Node: "ok", Span: span}, true
	case token.Tool:
		span := p.currentSpan()
		p.advance()
		return ast.Spanned[string]{Node: "tool", Span: span}, true
	case token.Agent:
		span := p.currentSpan()
		p.advance()
		return ast.Spanned[string]{Node: "agent", Span: span}, true
	case token.Provider:
		span := p.currentSpan()
		p.advance()
		return ast.Spanned[string]{Node: "provider", Span: span}, true
	default:
		p.addError("expected identifier, found "+tok.Kind.String(), ast.Span{Start: tok.Start, End: tok.End})
		return ast.Spanned[string]{}, false
	}
}

// ---------------------------------------------------------------------------
// Interpolated string parts
// ---------------------------------------------------------------------------

func (p *Parser) parseInterpolatedStringParts(raw string) []ast.StringPart {
	var parts []ast.StringPart
	var literal strings.Builder
	i := 0

	for i < len(raw) {
		ch := raw[i]

		if ch == '\\' && i+1 < len(raw) {
			next := raw[i+1]
			switch next {
			case 'n':
				literal.WriteByte('\n')
			case 't':
				literal.WriteByte('\t')
			case 'r':
				literal.WriteByte('\r')
			case '\\':
				literal.WriteByte('\\')
			case '"':
				literal.WriteByte('"')
			case '{':
				literal.WriteByte('{')
			case '}':
				literal.WriteByte('}')
			default:
				literal.WriteByte('\\')
				literal.WriteByte(next)
			}
			i += 2
			continue
		}

		if ch == '$' && i+1 < len(raw) && raw[i+1] == '{' {
			// Flush literal.
			if literal.Len() > 0 {
				parts = append(parts, ast.LiteralPart{Value: literal.String()})
				literal.Reset()
			}

			// Extract expression text between ${ and }.
			i += 2 // skip ${
			depth := 1
			exprStart := i
			for i < len(raw) && depth > 0 {
				if raw[i] == '{' {
					depth++
				} else if raw[i] == '}' {
					depth--
					if depth == 0 {
						break
					}
				}
				i++
			}
			exprText := raw[exprStart:i]
			if i < len(raw) {
				i++ // skip closing }
			}

			// Parse the expression via a sub-parser.
			if len(exprText) > 0 {
				sf, _ := Parse(exprText)
				if sf != nil && len(sf.Items) > 0 {
					if is, ok := sf.Items[0].Node.(ast.ItemStatement); ok {
						if es, ok := is.Stmt.Node.(ast.ExprStmt); ok {
							parts = append(parts, ast.ExprPart{Value: es.Value})
							continue
						}
					}
				}
				// Fallback: treat as literal.
				parts = append(parts, ast.LiteralPart{Value: "{" + exprText + "}"})
			}
			continue
		}

		literal.WriteByte(ch)
		i++
	}

	if literal.Len() > 0 {
		parts = append(parts, ast.LiteralPart{Value: literal.String()})
	}

	return parts
}

// ---------------------------------------------------------------------------
// Utility functions
// ---------------------------------------------------------------------------

func firstRune(s string) rune {
	for _, r := range s {
		return r
	}
	return 'a'
}

func parseInt(raw string) int64 {
	// Remove underscores.
	clean := strings.ReplaceAll(raw, "_", "")

	// Handle hex, binary, octal prefixes.
	if len(clean) > 2 {
		switch clean[:2] {
		case "0x", "0X":
			v, _ := strconv.ParseInt(clean[2:], 16, 64)
			return v
		case "0b", "0B":
			v, _ := strconv.ParseInt(clean[2:], 2, 64)
			return v
		case "0o", "0O":
			v, _ := strconv.ParseInt(clean[2:], 8, 64)
			return v
		}
	}

	v, _ := strconv.ParseInt(clean, 10, 64)
	return v
}

func parseFloat(raw string) float64 {
	clean := strings.ReplaceAll(raw, "_", "")
	v, _ := strconv.ParseFloat(clean, 64)
	return v
}
