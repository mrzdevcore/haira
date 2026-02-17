package formatter

import (
	"fmt"
	"strings"

	"github.com/haira-lang/haira/internal/ast"
)

// formatExpr formats an expression.
func (f *Formatter) formatExpr(expr ast.Expr) {
	switch e := expr.Node.(type) {
	case ast.LiteralExpr:
		f.formatLiteral(e.Lit)
	case ast.IdentExpr:
		f.write(e.Name)
	case ast.BinaryExpr:
		f.formatBinary(e)
	case ast.UnaryExpr:
		f.formatUnary(e)
	case ast.CallExpr:
		f.formatCall(e)
	case ast.MethodCallExpr:
		f.formatMethodCall(e)
	case ast.FieldExpr:
		f.formatExpr(e.Object)
		f.write(".")
		f.write(e.Field.Node)
	case ast.IndexExpr:
		f.formatExpr(e.Object)
		f.write("[")
		f.formatExpr(e.Index)
		f.write("]")
	case ast.PipeExpr:
		f.formatPipe(e)
	case ast.LambdaExpr:
		f.formatLambda(e)
	case ast.MatchExpr:
		f.formatMatchExpr(e)
	case ast.IfExpr:
		f.formatIfExpr(e)
	case ast.BlockExpr:
		f.formatBlock(e.Body)
	case ast.ListExpr:
		f.formatList(e)
	case ast.MapExpr:
		f.formatMap(e)
	case ast.InstanceExpr:
		f.formatInstance(e)
	case ast.RangeExpr:
		f.formatExpr(e.Start)
		if e.Inclusive {
			f.write("..=")
		} else {
			f.write("..")
		}
		f.formatExpr(e.End)
	case ast.PropagateExpr:
		f.formatExpr(e.Inner)
		f.write("?")
	case ast.OrelseExpr:
		f.formatExpr(e.Left)
		f.write(" orelse ")
		f.formatExpr(e.Default)
	case ast.SomeExpr:
		f.write("some(")
		f.formatExpr(e.Inner)
		f.write(")")
	case ast.NoneExpr:
		f.write("none")
	case ast.AsyncExpr:
		f.write("async ")
		f.formatBlock(e.Body)
	case ast.SpawnExpr:
		f.write("spawn ")
		f.formatBlock(e.Body)
	case ast.SelectExpr:
		f.formatSelect(e)
	case ast.ParenExpr:
		f.write("(")
		f.formatExpr(e.Inner)
		f.write(")")
	}
}

// formatLiteral formats a literal value.
func (f *Formatter) formatLiteral(lit ast.Literal) {
	switch l := lit.(type) {
	case ast.IntLit:
		f.write(fmt.Sprintf("%d", l.Value))
	case ast.FloatLit:
		f.write(fmt.Sprintf("%g", l.Value))
	case ast.StringLit:
		f.write(`"`)
		f.write(escapeString(l.Value))
		f.write(`"`)
	case ast.BoolLit:
		if l.Value {
			f.write("true")
		} else {
			f.write("false")
		}
	case ast.InterpolatedStringLit:
		f.formatInterpolatedString(l)
	}
}

// escapeString escapes special characters in a string literal.
func escapeString(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case '\t':
			b.WriteString(`\t`)
		case '\r':
			b.WriteString(`\r`)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// formatInterpolatedString reconstructs "...${expr}..." strings.
func (f *Formatter) formatInterpolatedString(lit ast.InterpolatedStringLit) {
	f.write(`"`)
	for _, part := range lit.Parts {
		switch p := part.(type) {
		case ast.LiteralPart:
			f.write(escapeString(p.Value))
		case ast.ExprPart:
			f.write("${")
			f.formatExpr(p.Value)
			f.write("}")
		}
	}
	f.write(`"`)
}

// binaryOpStr returns the string representation of a binary operator.
func binaryOpStr(op ast.BinaryOp) string {
	switch op {
	case ast.OpAdd:
		return "+"
	case ast.OpSub:
		return "-"
	case ast.OpMul:
		return "*"
	case ast.OpDiv:
		return "/"
	case ast.OpMod:
		return "%"
	case ast.OpEq:
		return "=="
	case ast.OpNe:
		return "!="
	case ast.OpLt:
		return "<"
	case ast.OpGt:
		return ">"
	case ast.OpLe:
		return "<="
	case ast.OpGe:
		return ">="
	case ast.OpAnd:
		return "and"
	case ast.OpOr:
		return "or"
	case ast.OpBitAnd:
		return "&"
	case ast.OpBitOr:
		return "|"
	case ast.OpBitXor:
		return "^"
	case ast.OpShl:
		return "<<"
	case ast.OpShr:
		return ">>"
	default:
		return "?"
	}
}

func (f *Formatter) formatBinary(e ast.BinaryExpr) {
	f.formatExpr(e.Left)
	f.write(" ")
	f.write(binaryOpStr(e.Op.Node))
	f.write(" ")
	f.formatExpr(e.Right)
}

func (f *Formatter) formatUnary(e ast.UnaryExpr) {
	switch e.Op.Node {
	case ast.OpNeg:
		f.write("-")
	case ast.OpNot:
		f.write("not ")
	case ast.OpBitNot:
		f.write("~")
	}
	f.formatExpr(e.Operand)
}

func (f *Formatter) formatCall(e ast.CallExpr) {
	f.formatExpr(e.Callee)
	f.write("(")
	f.formatArgs(e.Args)
	f.write(")")
}

func (f *Formatter) formatMethodCall(e ast.MethodCallExpr) {
	f.formatExpr(e.Receiver)
	f.write(".")
	f.write(e.Method.Node)
	f.write("(")
	f.formatArgs(e.Args)
	f.write(")")
}

func (f *Formatter) formatArgs(args []ast.Argument) {
	for i, arg := range args {
		if i > 0 {
			f.write(", ")
		}
		if arg.Name != nil {
			f.write(arg.Name.Node)
			f.write(": ")
		}
		f.formatExpr(arg.Value)
	}
}

func (f *Formatter) formatPipe(e ast.PipeExpr) {
	f.formatExpr(e.Left)
	f.write(" |> ")
	f.formatExpr(e.Right)
}

func (f *Formatter) formatLambda(e ast.LambdaExpr) {
	f.write("|")
	for i, p := range e.Params {
		if i > 0 {
			f.write(", ")
		}
		f.write(p.Name.Node)
		f.formatColonType(p.Ty)
	}
	f.write("| ")
	switch b := e.Body.(type) {
	case ast.LambdaExprBody:
		f.formatExpr(b.Value)
	case ast.LambdaBlockBody:
		f.formatBlock(b.Value)
	}
}

func (f *Formatter) formatList(e ast.ListExpr) {
	f.write("[")
	for i, elem := range e.Elems {
		if i > 0 {
			f.write(", ")
		}
		f.formatExpr(elem)
	}
	f.write("]")
}

func (f *Formatter) formatMap(e ast.MapExpr) {
	if len(e.Entries) == 0 {
		f.write("{}")
		return
	}
	f.write("{ ")
	for i, entry := range e.Entries {
		if i > 0 {
			f.write(", ")
		}
		f.formatExpr(entry.Key)
		f.write(": ")
		f.formatExpr(entry.Value)
	}
	f.write(" }")
}

func (f *Formatter) formatInstance(e ast.InstanceExpr) {
	f.write(e.TypeName.Node)
	if len(e.Fields) == 0 {
		f.write("{}")
		return
	}
	f.write("{ ")
	for i, field := range e.Fields {
		if i > 0 {
			f.write(", ")
		}
		if field.Name != nil {
			f.write(field.Name.Node)
			f.write(" = ")
		}
		f.formatExpr(field.Value)
	}
	f.write(" }")
}

func (f *Formatter) formatMatchExpr(e ast.MatchExpr) {
	f.write("match ")
	f.formatExpr(e.Subject)
	f.writeln(" {")
	f.incIndent()
	for _, arm := range e.Arms {
		f.writeIndent()
		f.formatPatternSpanned(arm.Pattern)
		if arm.Guard != nil {
			f.write(" if ")
			f.formatExpr(*arm.Guard)
		}
		f.write(" => ")
		f.formatMatchArmBody(arm.Body)
	}
	f.decIndent()
	f.writeIndent()
	f.write("}")
}

// formatMatchArmBody formats a match arm body. Single statements are emitted
// inline; multi-statement blocks use braces.
func (f *Formatter) formatMatchArmBody(body ast.MatchArmBody) {
	switch b := body.(type) {
	case ast.MatchArmExpr:
		f.formatExpr(b.Value)
		f.newline()
	case ast.MatchArmBlock:
		if len(b.Value.Statements) == 1 {
			// Single statement: emit inline without braces
			f.formatInlineStatement(b.Value.Statements[0])
			f.newline()
		} else {
			f.formatBlock(b.Value)
			f.newline()
		}
	}
}

// formatInlineStatement formats a statement without indentation or trailing newline.
func (f *Formatter) formatInlineStatement(stmt ast.Statement) {
	switch s := stmt.Node.(type) {
	case ast.ReturnStmt:
		f.formatReturnStmt(s)
	case ast.ExprStmt:
		f.formatExpr(s.Value)
	case ast.BreakStmt:
		f.write("break")
	case ast.ContinueStmt:
		f.write("continue")
	default:
		// For complex statements, fall back to block style
		f.formatExpr(ast.Expr{Node: ast.BlockExpr{Body: ast.Block{Statements: []ast.Statement{stmt}}}, Span: stmt.Span})
	}
}

func (f *Formatter) formatPatternSpanned(sp ast.Spanned[ast.Pattern]) {
	// For ident patterns, use the original source text to preserve
	// dotted syntax (e.g., Direction.North) which the parser normalizes.
	if _, ok := sp.Node.(ast.IdentPattern); ok {
		if sp.Span.Start < sp.Span.End && sp.Span.End <= len(f.src) {
			f.write(f.src[sp.Span.Start:sp.Span.End])
			return
		}
	}
	f.formatPattern(sp.Node)
}

func (f *Formatter) formatPattern(p ast.Pattern) {
	switch pat := p.(type) {
	case ast.WildcardPattern:
		f.write("_")
	case ast.LiteralPattern:
		f.formatLiteral(pat.Lit)
	case ast.IdentPattern:
		f.write(pat.Name)
	case ast.ConstructorPattern:
		f.write(pat.Name)
		if len(pat.Fields) > 0 {
			f.write("(")
			for i, field := range pat.Fields {
				if i > 0 {
					f.write(", ")
				}
				f.write(field.Node)
			}
			f.write(")")
		}
	case ast.OrPattern:
		for i, sub := range pat.Patterns {
			if i > 0 {
				f.write(" | ")
			}
			f.formatPattern(sub.Node)
		}
	case ast.RangePattern:
		f.formatLiteral(pat.Start)
		if pat.Inclusive {
			f.write("..=")
		} else {
			f.write("..")
		}
		f.formatLiteral(pat.End)
	}
}

func (f *Formatter) formatSelect(e ast.SelectExpr) {
	f.writeln("select {")
	f.incIndent()
	for _, arm := range e.Arms {
		f.writeIndent()
		f.write(arm.Binding.Node)
		f.write(" = ")
		f.formatExpr(arm.Channel)
		f.write(" => ")
		f.formatMatchArmBody(arm.Body)
	}
	if e.Default != nil {
		f.writeIndent()
		f.write("default => ")
		f.formatBlock(*e.Default)
		f.newline()
	}
	f.decIndent()
	f.writeIndent()
	f.write("}")
}

func (f *Formatter) formatIfExpr(e ast.IfExpr) {
	// Delegate to the if statement formatter
	f.formatIfStmt(e.If)
}
