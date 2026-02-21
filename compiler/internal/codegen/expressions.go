package codegen

import (
	"fmt"
	"strings"

	"github.com/haira-lang/haira/internal/ast"
	"github.com/haira-lang/haira/internal/checker"
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
		case ast.OpBitNot:
			return "^" + operand // Go uses ^ for bitwise NOT
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
					if m == "ask" || m == "run" || m == "stream" {
						if hasSession {
							positional = append(positional, sessionArg)
						} else {
							positional = append(positional, `""`)
						}
					}
					return fmt.Sprintf("%s.%s(%s)", agentVar, method, strings.Join(positional, ", "))
				}
			}
		}
		receiver := ExprToGo(e.Receiver)
		method := SnakeToPascal(e.Method.Node)
		return fmt.Sprintf("%s.%s(%s)", receiver, method, argsWithNamedToGo(e.Args))
	case ast.FieldExpr:
		// Enum variant access: Direction.North → DirectionNorth
		if ident, ok := e.Object.Node.(ast.IdentExpr); ok {
			if len(ident.Name) > 0 && ident.Name[0] >= 'A' && ident.Name[0] <= 'Z' {
				return ident.Name + e.Field.Node
			}
		}
		// Map .message on error-typed variables to .Error()
		// Go's error interface has no .Message field — only the .Error() method.
		if e.Field.Node == "message" {
			if activeTypeInfo != nil {
				if objTy, ok := activeTypeInfo.ExprTypes[e.Object.Span]; ok {
					if _, isErr := objTy.(checker.ErrorType); isErr {
						return ExprToGo(e.Object) + ".Error()"
					}
				}
			}
			// Runtime fallback: use haira.ErrorMessage() for safe access
			// on any variable that might be an error at runtime.
			if ident, ok := e.Object.Node.(ast.IdentExpr); ok {
				if ident.Name == "err" || strings.HasSuffix(ident.Name, "_err") || strings.HasSuffix(ident.Name, "Error") || strings.HasSuffix(ident.Name, "_error") {
					return fmt.Sprintf("haira.ErrorMessage(%s)", ExprToGo(e.Object))
				}
			}
		}
		// If the object is a known struct, convert field name to PascalCase
		if activeTypeInfo != nil {
			if objTy, ok := activeTypeInfo.ExprTypes[e.Object.Span]; ok {
				if _, isSt := objTy.(checker.StructType); isSt {
					return ExprToGo(e.Object) + "." + SnakeToPascal(e.Field.Node)
				}
			}
		}
		// Default: convert field name to PascalCase for Go struct field access (e.g., resp.status_code → resp.StatusCode)
		return ExprToGo(e.Object) + "." + SnakeToPascal(e.Field.Node)
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
		// Always use []any for list literals. Go's type system does not
		// allow covariant slice types ([]string is not assignable to []any),
		// causing issues with append and assignment.
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
		// Always use map[string]any for map literals. Go's type system does
		// not allow covariant map types (map[string]string is not assignable
		// to map[string]any), and Haira is dynamically typed so any is correct.
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
		return blockExprToGo(e)
	case ast.InstanceExpr:
		return instanceToGo(e)
	case ast.RangeExpr:
		return rangeExprToGo(e)
	case ast.PropagateExpr:
		return propagateExprToGo(e)
	case ast.OrelseExpr:
		return orelseExprToGo(e)
	case ast.AsyncExpr:
		return asyncExprToGo(e)
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
	}
	return "+"
}

var propagateCounter int

var orelseCounter int

// orelseReturnType returns the concrete Go return type for known stdlib calls
// that return (T, error). Falls back to "any" for unknown calls.
func orelseReturnType(e ast.Expr) string {
	mc, ok := e.Node.(ast.MethodCallExpr)
	if !ok {
		return "any"
	}
	recv, ok := mc.Receiver.Node.(ast.IdentExpr)
	if !ok {
		return "any"
	}
	// Module-level stdlib calls
	key := recv.Name + "." + mc.Method.Node
	switch key {
	case "excel.open":
		return "*haira.Workbook"
	case "excel.read_sheets":
		return "*haira.ExcelTables"
	case "postgres.connect":
		return "*haira.DB"
	}
	// HTTP module calls all return *haira.Response
	if recv.Name == "http" {
		switch mc.Method.Node {
		case "get", "get_with_headers", "post", "post_with_headers",
			"put", "put_with_headers", "delete", "delete_with_headers":
			return "*haira.Response"
		}
	}
	// Instance method calls — return types based on method name
	switch mc.Method.Node {
	case "read_sheet", "query":
		return "[]map[string]any"
	case "get", "post", "put", "patch", "delete":
		return "*haira.Response"
	case "schema":
		return "*haira.PgSchema"
	case "tables":
		return "[]any"
	case "merge_request", "create_mr":
		return "*haira.GitlabMR"
	case "pull_request", "create_pr":
		return "*haira.GithubPR"
	}
	return "any"
}

func orelseExprToGo(e ast.OrelseExpr) string {
	orelseCounter++
	id := orelseCounter
	left := ExprToGo(e.Left)
	def := ExprToGo(e.Default)
	retType := orelseReturnType(e.Left)
	// When the return type is a concrete slice type, cast empty list defaults
	if retType != "any" && def == "[]any{}" {
		def = retType + "{}"
	}
	return fmt.Sprintf("func() %s { _r%d, _e%d := %s; if _e%d != nil { return %s }; return _r%d }()",
		retType, id, id, left, id, def, id)
}

func propagateExprToGo(e ast.PropagateExpr) string {
	propagateCounter++
	id := propagateCounter
	inner := ExprToGo(e.Inner)
	return fmt.Sprintf("func() any { _r%d, _e%d := %s; if _e%d != nil { panic(_e%d) }; return _r%d }()",
		id, id, inner, id, id, id)
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
				// Escape % as %% so fmt.Sprintf treats them as literal percent signs
				fmtStr.WriteString(strings.ReplaceAll(p.Value, "%", "%%"))
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
	if needsIfChain(m) {
		return matchExprToGoIfChain(m)
	}
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
		emitMatchExprArmBody(em, arm.Body)
	}
	em.Line("\t}")
	em.Line("\treturn nil")
	em.Line("}()")
	return strings.TrimSpace(em.String())
}

func matchExprToGoIfChain(m ast.MatchExpr) string {
	subject := ExprToGo(m.Subject)
	em := NewEmitter()
	em.Line("func() any {")
	em.Line(fmt.Sprintf("\t_match := %s", subject))
	for i, arm := range m.Arms {
		cond := patternToCondition("_match", arm.Pattern.Node)
		if arm.Guard != nil {
			cond = fmt.Sprintf("%s && %s", cond, ExprToGo(*arm.Guard))
		}
		if _, ok := arm.Pattern.Node.(ast.WildcardPattern); ok && arm.Guard == nil {
			if i > 0 {
				em.Line("\t} else {")
			} else {
				em.Line("\t{")
			}
		} else {
			if i > 0 {
				em.Line(fmt.Sprintf("\t} else if %s {", cond))
			} else {
				em.Line(fmt.Sprintf("\tif %s {", cond))
			}
		}
		emitMatchExprArmBody(em, arm.Body)
	}
	em.Line("\t}")
	em.Line("\treturn nil")
	em.Line("}()")
	return strings.TrimSpace(em.String())
}

func emitMatchExprArmBody(em *GoEmitter, body ast.MatchArmBody) {
	switch b := body.(type) {
	case ast.MatchArmExpr:
		em.Line(fmt.Sprintf("\t\treturn %s", ExprToGo(b.Value)))
	case ast.MatchArmBlock:
		stmts := b.Value.Statements
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
	typeName := inst.TypeName.Node
	// Map qualified stdlib types: ui.StatusCard → haira.UIStatusCard
	isRuntime := false
	if goType, ok := qualifiedTypeToGo(typeName); ok {
		typeName = goType
		isRuntime = true
	}
	fields := make([]string, len(inst.Fields))
	for i, f := range inst.Fields {
		val := ExprToGo(f.Value)
		if f.Name != nil {
			// Runtime types use SnakeToPascal (e.g. confirm_label → ConfirmLabel)
			// User types use Capitalize (e.g. name → Name)
			key := Capitalize(f.Name.Node)
			if isRuntime {
				key = SnakeToPascal(f.Name.Node)
			}
			fields[i] = key + ": " + val
		} else {
			fields[i] = val
		}
	}
	return fmt.Sprintf("%s{%s}", typeName, strings.Join(fields, ", "))
}

func spawnToGo(spawn ast.SpawnExpr) string {
	em := NewEmitter()
	count := len(spawn.Body.Statements)
	em.Line("func() []any {")
	em.Line(fmt.Sprintf("\tresults := make([]any, %d)", count))
	em.Line(fmt.Sprintf("\terrs := make([]error, %d)", count))
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
		em.Line("\tgo func(idx int) {")
		em.Line("\t\tdefer wg.Done()")
		em.Line("\t\tdefer func() {")
		em.Line("\t\t\tif r := recover(); r != nil {")
		em.Line("\t\t\t\terrs[idx] = fmt.Errorf(\"spawn task %d panicked: %v\", idx, r)")
		em.Line("\t\t\t}")
		em.Line("\t\t}()")
		em.Line(fmt.Sprintf("\t\tresults[idx] = %s", exprStr))
		em.Line(fmt.Sprintf("\t}(%d)", i))
	}
	em.Line("\twg.Wait()")
	em.Line("\tfor _, e := range errs {")
	em.Line("\t\tif e != nil { panic(e) }")
	em.Line("\t}")
	em.Line("\treturn results")
	em.Line("}()")
	return strings.TrimSpace(em.String())
}

// blockExprToGo generates an IIFE that evaluates block statements and returns the last expression.
func blockExprToGo(block ast.BlockExpr) string {
	em := NewEmitter()
	em.Line("func() any {")
	stmts := block.Body.Statements
	for i, stmt := range stmts {
		if i == len(stmts)-1 {
			// Last statement: if it's an expression, return it
			if es, ok := stmt.Node.(ast.ExprStmt); ok {
				em.Line(fmt.Sprintf("\treturn %s", ExprToGo(es.Value)))
				continue
			}
		}
		inner := NewEmitter()
		EmitStatement(inner, stmt)
		em.Line("\t" + strings.TrimSpace(inner.String()))
	}
	if len(stmts) == 0 {
		em.Line("\treturn nil")
	}
	em.Line("}()")
	return strings.TrimSpace(em.String())
}

// rangeExprToGo generates a Go slice for a range expression (e.g., 0..5 → []any{0, 1, 2, 3, 4}).
// When used as an iterator in for-loops, the for-loop codegen handles ranges directly.
// This handles the case where a range is used as a standalone expression value.
func rangeExprToGo(r ast.RangeExpr) string {
	start := ExprToGo(r.Start)
	end := ExprToGo(r.End)
	return fmt.Sprintf("haira.Range(%s, %s, %v)", start, end, r.Inclusive)
}

// asyncExprToGo generates a goroutine that returns a channel with the result.
func asyncExprToGo(a ast.AsyncExpr) string {
	em := NewEmitter()
	em.Line("func() chan any {")
	em.Line("\tch := make(chan any, 1)")
	em.Line("\tgo func() {")
	stmts := a.Body.Statements
	for i, stmt := range stmts {
		if i == len(stmts)-1 {
			if es, ok := stmt.Node.(ast.ExprStmt); ok {
				em.Line(fmt.Sprintf("\t\tch <- %s", ExprToGo(es.Value)))
				continue
			}
		}
		inner := NewEmitter()
		EmitStatement(inner, stmt)
		em.Line("\t\t" + strings.TrimSpace(inner.String()))
	}
	if len(stmts) == 0 {
		em.Line("\t\tch <- nil")
	}
	em.Line("\t}()")
	em.Line("\treturn ch")
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
	case ast.OrPattern:
		parts := make([]string, len(p.Patterns))
		for i, sub := range p.Patterns {
			parts[i] = patternToGoExpr(sub.Node)
		}
		return strings.Join(parts, ", ")
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

// nodeKind returns a string tag identifying the broad category of an AST expression node.
// Used to detect heterogeneous map literal values (e.g., bool + string).
func nodeKind(node ast.ExprKind) string {
	switch n := node.(type) {
	case ast.LiteralExpr:
		switch n.Lit.(type) {
		case ast.BoolLit:
			return "bool"
		case ast.IntLit:
			return "int"
		case ast.FloatLit:
			return "float"
		case ast.StringLit, ast.InterpolatedStringLit:
			return "string"
		default:
			return "literal"
		}
	case ast.BinaryExpr:
		return "binary"
	case ast.CallExpr:
		return "call"
	case ast.MethodCallExpr:
		return "method"
	case ast.IdentExpr:
		return "ident"
	case ast.IndexExpr:
		return "index"
	case ast.MapExpr:
		return "map"
	case ast.ListExpr:
		return "list"
	default:
		return fmt.Sprintf("%T", node)
	}
}

// argsWithNamedToGo converts a list of arguments to Go, collecting named args into a map.
// If there are no named args, returns all args positionally as usual.
// If there are named args, positional args come first, then a map[string]any{...} for named args.
func argsWithNamedToGo(args []ast.Argument) string {
	hasNamed := false
	for _, a := range args {
		if a.Name != nil {
			hasNamed = true
			break
		}
	}
	if !hasNamed {
		parts := make([]string, len(args))
		for i, a := range args {
			parts[i] = ExprToGo(a.Value)
		}
		return strings.Join(parts, ", ")
	}

	var positional []string
	var named []string
	for _, a := range args {
		if a.Name != nil {
			named = append(named, fmt.Sprintf("%q: %s", a.Name.Node, ExprToGo(a.Value)))
		} else {
			positional = append(positional, ExprToGo(a.Value))
		}
	}
	parts := positional
	parts = append(parts, fmt.Sprintf("map[string]any{%s}", strings.Join(named, ", ")))
	return strings.Join(parts, ", ")
}
