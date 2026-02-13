package codegen

import (
	"fmt"
	"strings"

	"github.com/haira-lang/haira/internal/ast"
)

// ExprToGo converts a Haira expression to a Go expression string.
func ExprToGo(expr ast.Expr) string {
	switch e := expr.Node.(type) {
	case ast.LiteralExpr:
		return literalToGo(e.Lit)
	case ast.IdentExpr:
		return e.Name
	case ast.BinaryExpr:
		return binaryToGo(e)
	case ast.UnaryExpr:
		operand := ExprToGo(e.Operand)
		switch e.Op.Node {
		case ast.OpNeg:
			return "-" + operand
		case ast.OpNot:
			return "!" + operand
		}
		return operand
	case ast.CallExpr:
		if resolved, ok := ResolveStdlibCall(e); ok {
			return resolved
		}
		callee := ExprToGo(e.Callee)
		if ident, ok := e.Callee.Node.(ast.IdentExpr); ok {
			callee = SnakeToPascal(ident.Name)
		}
		args := make([]string, len(e.Args))
		for i, a := range e.Args {
			args[i] = ExprToGo(a.Value)
		}
		return fmt.Sprintf("%s(%s)", callee, strings.Join(args, ", "))
	case ast.MethodCallExpr:
		if resolved, ok := ResolveStdlibMethodCall(e); ok {
			return resolved
		}
		// Agent method calls: PascalCase receiver + ask/run/stream
		if ident, ok := e.Receiver.Node.(ast.IdentExpr); ok {
			if len(ident.Name) > 0 && ident.Name[0] >= 'A' && ident.Name[0] <= 'Z' {
				m := e.Method.Node
				if m == "ask" || m == "run" || m == "stream" {
					agentVar := "agent" + ident.Name
					method := Capitalize(m)
					var positional []string
					var sessionArg string
					hasSession := false
					for _, arg := range e.Args {
						if arg.Name != nil && arg.Name.Node == "session" {
							sessionArg = ExprToGo(arg.Value)
							hasSession = true
						} else {
							positional = append(positional, ExprToGo(arg.Value))
						}
					}
					if m == "ask" {
						if hasSession {
							positional = append(positional, sessionArg)
						} else {
							positional = append(positional, `""`)
						}
					} else if hasSession {
						positional = append(positional, sessionArg)
					}
					return fmt.Sprintf("%s.%s(%s)", agentVar, method, strings.Join(positional, ", "))
				}
			}
		}
		receiver := ExprToGo(e.Receiver)
		method := SnakeToPascal(e.Method.Node)
		args := make([]string, len(e.Args))
		for i, a := range e.Args {
			args[i] = ExprToGo(a.Value)
		}
		return fmt.Sprintf("%s.%s(%s)", receiver, method, strings.Join(args, ", "))
	case ast.FieldExpr:
		// Enum variant access: Direction.North → DirectionNorth
		if ident, ok := e.Object.Node.(ast.IdentExpr); ok {
			if len(ident.Name) > 0 && ident.Name[0] >= 'A' && ident.Name[0] <= 'Z' {
				return ident.Name + e.Field.Node
			}
		}
		return ExprToGo(e.Object) + "." + e.Field.Node
	case ast.IndexExpr:
		obj := ExprToGo(e.Object)
		index := ExprToGo(e.Index)
		return fmt.Sprintf("haira.Get(%s, %s)", obj, index)
	case ast.PipeExpr:
		left := ExprToGo(e.Left)
		right := ExprToGo(e.Right)
		if ident, ok := e.Right.Node.(ast.IdentExpr); ok {
			right = SnakeToPascal(ident.Name)
		}
		return fmt.Sprintf("%s(%s)", right, left)
	case ast.ListExpr:
		elems := make([]string, len(e.Elems))
		for i, el := range e.Elems {
			elems[i] = ExprToGo(el)
		}
		return fmt.Sprintf("[]any{%s}", strings.Join(elems, ", "))
	case ast.MapExpr:
		pairs := make([]string, len(e.Entries))
		for i, entry := range e.Entries {
			key := ExprToGo(entry.Key)
			if ident, ok := entry.Key.Node.(ast.IdentExpr); ok {
				key = fmt.Sprintf("%q", ident.Name)
			}
			pairs[i] = fmt.Sprintf("%s: %s", key, ExprToGo(entry.Value))
		}
		return fmt.Sprintf("map[string]any{%s}", strings.Join(pairs, ", "))
	case ast.ParenExpr:
		return "(" + ExprToGo(e.Inner) + ")"
	case ast.NoneExpr:
		return "nil"
	case ast.SomeExpr:
		return ExprToGo(e.Inner)
	case ast.LambdaExpr:
		return lambdaToGo(e)
	case ast.MatchExpr:
		return matchExprToGo(e)
	case ast.IfExpr:
		return ifExprToGo(e.If)
	case ast.BlockExpr:
		return "nil /* block-expr */"
	case ast.InstanceExpr:
		return instanceToGo(e)
	case ast.RangeExpr:
		return "nil /* range */"
	case ast.PropagateExpr:
		return ExprToGo(e.Inner)
	case ast.AsyncExpr:
		return "nil /* async */"
	case ast.SpawnExpr:
		return spawnToGo(e)
	case ast.SelectExpr:
		return selectToGo(e)
	default:
		return "nil"
	}
}

func binaryToGo(bin ast.BinaryExpr) string {
	// Detect slice concatenation: x + [...] → append(x, ...items)
	if bin.Op.Node == ast.OpAdd {
		if _, ok := bin.Right.Node.(ast.ListExpr); ok {
			return fmt.Sprintf("append(%s, %s...)", ExprToGo(bin.Left), ExprToGo(bin.Right))
		}
	}
	// Detect string concatenation involving dynamic values
	if bin.Op.Node == ast.OpAdd && hasAnyTypedOperand(bin.Left, bin.Right) {
		return fmt.Sprintf("haira.Concat(%s, %s)", ExprToGo(bin.Left), ExprToGo(bin.Right))
	}
	left := ExprToGo(bin.Left)
	right := ExprToGo(bin.Right)
	op := binopToGo(bin.Op.Node)
	return fmt.Sprintf("%s %s %s", left, op, right)
}

func binopToGo(op ast.BinaryOp) string {
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
		return "&&"
	case ast.OpOr:
		return "||"
	}
	return "+"
}

func literalToGo(lit ast.Literal) string {
	switch l := lit.(type) {
	case ast.IntLit:
		return fmt.Sprintf("%d", l.Value)
	case ast.FloatLit:
		s := fmt.Sprintf("%g", l.Value)
		if !strings.Contains(s, ".") {
			s += ".0"
		}
		return s
	case ast.StringLit:
		if strings.Contains(l.Value, "\n") {
			return "`" + l.Value + "`"
		}
		return fmt.Sprintf("%q", l.Value)
	case ast.BoolLit:
		if l.Value {
			return "true"
		}
		return "false"
	case ast.InterpolatedStringLit:
		var fmtStr strings.Builder
		var args []string
		for _, part := range l.Parts {
			switch p := part.(type) {
			case ast.LiteralPart:
				fmtStr.WriteString(p.Value)
			case ast.ExprPart:
				fmtStr.WriteString("%v")
				args = append(args, ExprToGo(p.Value))
			}
		}
		if len(args) == 0 {
			return fmt.Sprintf("%q", fmtStr.String())
		}
		return fmt.Sprintf("fmt.Sprintf(%q, %s)", fmtStr.String(), strings.Join(args, ", "))
	}
	return "nil"
}

func lambdaToGo(lambda ast.LambdaExpr) string {
	params := make([]string, len(lambda.Params))
	for i, p := range lambda.Params {
		ty := "any"
		if p.Ty != nil {
			ty = HairaTypeToGo(p.Ty.Node)
		}
		params[i] = p.Name.Node + " " + ty
	}
	paramStr := strings.Join(params, ", ")

	switch body := lambda.Body.(type) {
	case ast.LambdaExprBody:
		return fmt.Sprintf("func(%s) any { return %s }", paramStr, ExprToGo(body.Value))
	case ast.LambdaBlockBody:
		em := NewEmitter()
		EmitBlockBody(em, body.Value)
		return fmt.Sprintf("func(%s) any {\n%s}", paramStr, em.String())
	}
	return "nil"
}

func matchExprToGo(m ast.MatchExpr) string {
	subject := ExprToGo(m.Subject)
	em := NewEmitter()
	em.Line("func() any {")
	em.Line(fmt.Sprintf("\tswitch %s {", subject))
	for _, arm := range m.Arms {
		if _, ok := arm.Pattern.Node.(ast.WildcardPattern); ok {
			em.Line("\tdefault:")
		} else {
			em.Line(fmt.Sprintf("\tcase %s:", patternToGoExpr(arm.Pattern.Node)))
		}
		switch body := arm.Body.(type) {
		case ast.MatchArmExpr:
			em.Line(fmt.Sprintf("\t\treturn %s", ExprToGo(body.Value)))
		case ast.MatchArmBlock:
			stmts := body.Value.Statements
			for i, stmt := range stmts {
				if i == len(stmts)-1 {
					if es, ok := stmt.Node.(ast.ExprStmt); ok {
						em.Line(fmt.Sprintf("\t\treturn %s", ExprToGo(es.Value)))
						continue
					}
				}
				inner := NewEmitter()
				EmitStatement(inner, stmt)
				em.Line("\t\t" + strings.TrimSpace(inner.String()))
			}
		}
	}
	em.Line("\t}")
	em.Line("\treturn nil")
	em.Line("}()")
	return strings.TrimSpace(em.String())
}

func ifExprToGo(ifStmt ast.IfStmt) string {
	cond := ExprToGo(ifStmt.Condition)
	thenVal := blockLastExpr(ifStmt.ThenBranch)
	elseVal := "nil"
	if ifStmt.ElseBranch != nil {
		switch eb := ifStmt.ElseBranch.(type) {
		case *ast.ElseBlock:
			elseVal = blockLastExpr(eb.Body)
		case *ast.ElseIf:
			inner := ast.Expr{Node: ast.IfExpr{If: eb.If.Node}, Span: eb.If.Span}
			elseVal = ExprToGo(inner)
		}
	}
	return fmt.Sprintf("func() any { if %s { return %s } else { return %s } }()", cond, thenVal, elseVal)
}

func blockLastExpr(block ast.Block) string {
	if len(block.Statements) == 0 {
		return "nil"
	}
	last := block.Statements[len(block.Statements)-1]
	switch s := last.Node.(type) {
	case ast.ExprStmt:
		return ExprToGo(s.Value)
	case ast.ReturnStmt:
		if len(s.Values) > 0 {
			return ExprToGo(s.Values[0])
		}
	}
	return "nil"
}

func instanceToGo(inst ast.InstanceExpr) string {
	fields := make([]string, len(inst.Fields))
	for i, f := range inst.Fields {
		val := ExprToGo(f.Value)
		if f.Name != nil {
			fields[i] = Capitalize(f.Name.Node) + ": " + val
		} else {
			fields[i] = val
		}
	}
	return fmt.Sprintf("%s{%s}", inst.TypeName.Node, strings.Join(fields, ", "))
}

func spawnToGo(spawn ast.SpawnExpr) string {
	em := NewEmitter()
	count := len(spawn.Body.Statements)
	em.Line("func() []any {")
	em.Line(fmt.Sprintf("\tresults := make([]any, %d)", count))
	em.Line("\tvar wg sync.WaitGroup")
	em.Line(fmt.Sprintf("\twg.Add(%d)", count))
	for i, stmt := range spawn.Body.Statements {
		exprStr := "nil"
		switch s := stmt.Node.(type) {
		case ast.ExprStmt:
			exprStr = ExprToGo(s.Value)
		case ast.AssignStmt:
			exprStr = ExprToGo(s.Value)
		}
		em.Line("\tgo func() {")
		em.Line("\t\tdefer wg.Done()")
		em.Line(fmt.Sprintf("\t\tresults[%d] = %s", i, exprStr))
		em.Line("\t}()")
	}
	em.Line("\twg.Wait()")
	em.Line("\treturn results")
	em.Line("}()")
	return strings.TrimSpace(em.String())
}

func selectToGo(sel ast.SelectExpr) string {
	em := NewEmitter()
	em.OpenBlock("select")
	for _, arm := range sel.Arms {
		ch := ExprToGo(arm.Channel)
		em.Line(fmt.Sprintf("case %s := <-%s:", arm.Binding.Node, ch))
		em.Indent()
		switch body := arm.Body.(type) {
		case ast.MatchArmExpr:
			em.Line(ExprToGo(body.Value))
		case ast.MatchArmBlock:
			EmitBlockBody(em, body.Value)
		}
		em.Dedent()
	}
	if sel.Default != nil {
		em.Line("default:")
		em.Indent()
		EmitBlockBody(em, *sel.Default)
		em.Dedent()
	}
	em.CloseBlock()
	return strings.TrimSpace(em.String())
}

func patternToGoExpr(pattern ast.Pattern) string {
	switch p := pattern.(type) {
	case ast.WildcardPattern:
		return "default"
	case ast.LiteralPattern:
		return literalToGo(p.Lit)
	case ast.IdentPattern:
		return p.Name
	case ast.ConstructorPattern:
		return p.Name
	}
	return "nil"
}

// hasAnyTypedOperand checks if either operand might produce an `any` value.
func hasAnyTypedOperand(left, right ast.Expr) bool {
	hasStr := involvesString(left) || involvesString(right)
	hasDyn := isDynamicValue(left) || isDynamicValue(right)
	return (hasStr || hasDyn) && !bothNumeric(left, right)
}

func isDynamicValue(expr ast.Expr) bool {
	switch e := expr.Node.(type) {
	case ast.IndexExpr:
		return true
	case ast.MethodCallExpr:
		return true
	case ast.BinaryExpr:
		if e.Op.Node == ast.OpAdd {
			return hasAnyTypedOperand(e.Left, e.Right)
		}
	}
	return false
}

func bothNumeric(left, right ast.Expr) bool {
	return isNumericLiteral(left) && isNumericLiteral(right)
}

func isNumericLiteral(expr ast.Expr) bool {
	if lit, ok := expr.Node.(ast.LiteralExpr); ok {
		switch lit.Lit.(type) {
		case ast.IntLit, ast.FloatLit:
			return true
		}
	}
	return false
}

func involvesString(expr ast.Expr) bool {
	if lit, ok := expr.Node.(ast.LiteralExpr); ok {
		switch lit.Lit.(type) {
		case ast.StringLit, ast.InterpolatedStringLit:
			return true
		}
	}
	return false
}
