package codegen

import (
	"fmt"
	"strings"

	"github.com/haira-lang/haira/internal/ast"
)

// EmitStatement emits a Haira statement as Go code.
func EmitStatement(em *GoEmitter, stmt ast.Statement) {
	switch s := stmt.Node.(type) {
	case ast.AssignStmt:
		emitAssignment(em, s)
	case ast.IfStmt:
		emitIf(em, s)
	case ast.ForStmt:
		emitFor(em, s)
	case ast.WhileStmt:
		emitWhile(em, s)
	case ast.ReturnStmt:
		emitReturn(em, s)
	case ast.ExprStmt:
		em.Line(ExprToGo(s.Value))
	case ast.DeferStmt:
		em.Line("defer " + ExprToGo(s.Value))
	case ast.BreakStmt:
		em.Line("break")
	case ast.ContinueStmt:
		em.Line("continue")
	case ast.MatchStmt:
		emitMatch(em, s.Match)
	case ast.TryStmt:
		emitTry(em, s)
	case ast.StepStmt:
		emitStep(em, s)
	}
}

func emitAssignment(em *GoEmitter, assign ast.AssignStmt) {
	// Special case: match expression as RHS → emit switch that assigns to variable
	if len(assign.Targets) == 1 {
		if matchExpr, ok := assign.Value.Node.(ast.MatchExpr); ok {
			target := assignPathToGo(assign.Targets[0].Path)
			if ident, ok := assign.Targets[0].Path.(ast.IdentPath); ok {
				em.DeclareVar(ident.Name.Node)
			}
			em.Line(fmt.Sprintf("var %s any", target))
			emitMatchAssignment(em, target, matchExpr)
			return
		}
	}

	value := ExprToGo(assign.Value)

	if len(assign.Targets) == 1 {
		target := assign.Targets[0]
		// Index assignment: table_data[key] = val → haira.Set(obj, key, val)
		if idx, ok := target.Path.(ast.IndexPath); ok {
			obj := assignPathToGo(idx.Object)
			index := ExprToGo(idx.Index)
			em.Line(fmt.Sprintf("haira.Set(%s, %s, %s)", obj, index, value))
			return
		}
		path := assignPathToGo(target.Path)
		op := "="
		if ident, ok := target.Path.(ast.IdentPath); ok {
			if em.DeclareVar(ident.Name.Node) {
				op = ":="
			}
		}
		em.Line(fmt.Sprintf("%s %s %s", path, op, value))
	} else {
		targets := make([]string, len(assign.Targets))
		anyNew := false
		for i, t := range assign.Targets {
			targets[i] = assignPathToGo(t.Path)
			if ident, ok := t.Path.(ast.IdentPath); ok {
				if em.DeclareVar(ident.Name.Node) {
					anyNew = true
				}
			}
		}
		op := "="
		if anyNew {
			op = ":="
		}
		em.Line(fmt.Sprintf("%s %s %s", strings.Join(targets, ", "), op, value))
	}
}

func emitMatchAssignment(em *GoEmitter, target string, matchExpr ast.MatchExpr) {
	subject := ExprToGo(matchExpr.Subject)
	em.OpenBlock(fmt.Sprintf("switch %s", subject))
	for _, arm := range matchExpr.Arms {
		emitMatchArmHeader(em, arm.Pattern.Node)
		em.Indent()
		switch body := arm.Body.(type) {
		case ast.MatchArmExpr:
			em.Line(fmt.Sprintf("%s = %s", target, ExprToGo(body.Value)))
		case ast.MatchArmBlock:
			EmitBlockBody(em, body.Value)
		}
		em.Dedent()
	}
	em.CloseBlock()
}

func assignPathToGo(path ast.AssignPath) string {
	switch p := path.(type) {
	case ast.IdentPath:
		return p.Name.Node
	case ast.FieldPath:
		return assignPathToGo(p.Object) + "." + p.Field.Node
	case ast.IndexPath:
		return fmt.Sprintf("%s[%s]", assignPathToGo(p.Object), ExprToGo(p.Index))
	}
	return "?"
}

func emitIf(em *GoEmitter, ifStmt ast.IfStmt) {
	cond := ExprToGo(ifStmt.Condition)
	em.OpenBlock(fmt.Sprintf("if %s", cond))
	EmitBlockBody(em, ifStmt.ThenBranch)
	if ifStmt.ElseBranch != nil {
		switch eb := ifStmt.ElseBranch.(type) {
		case *ast.ElseBlock:
			em.Dedent()
			em.Line("} else {")
			em.Indent()
			EmitBlockBody(em, eb.Body)
			em.CloseBlock()
			return
		case *ast.ElseIf:
			em.Dedent()
			cond := ExprToGo(eb.If.Node.Condition)
			em.Line(fmt.Sprintf("} else if %s {", cond))
			em.Indent()
			EmitBlockBody(em, eb.If.Node.ThenBranch)
			if eb.If.Node.ElseBranch != nil {
				switch inner := eb.If.Node.ElseBranch.(type) {
				case *ast.ElseBlock:
					em.Dedent()
					em.Line("} else {")
					em.Indent()
					EmitBlockBody(em, inner.Body)
				}
			}
			em.CloseBlock()
			return
		}
	}
	em.CloseBlock()
}

func emitFor(em *GoEmitter, forStmt ast.ForStmt) {
	// Range expression: for i in 0..10
	if rangeExpr, ok := forStmt.Iterator.Node.(ast.RangeExpr); ok {
		start := ExprToGo(rangeExpr.Start)
		end := ExprToGo(rangeExpr.End)
		op := "<"
		if rangeExpr.Inclusive {
			op = "<="
		}
		varName := forPatternVar(forStmt.Pattern)
		em.OpenBlock(fmt.Sprintf("for %s := %s; %s %s %s; %s++", varName, start, varName, op, end, varName))
	} else {
		iter := ExprToGo(forStmt.Iterator)
		_, isPairPattern := forStmt.Pattern.(ast.PairPattern)
		if isPairPattern {
			// PairPattern (for key, value in map) — use haira.ToMap() for safe map iteration
			switch forStmt.Iterator.Node.(type) {
			case ast.MapExpr:
				// Already a map literal, use directly
			default:
				if isTypedNonAny(forStmt.Iterator.Span) {
					// Type checker knows this is a typed map, use directly
				} else {
					iter = fmt.Sprintf("haira.ToMap(%s)", iter)
				}
			}
			p := forStmt.Pattern.(ast.PairPattern)
			em.OpenBlock(fmt.Sprintf("for %s, %s := range %s", p.First.Node, p.Second.Node, iter))
		} else {
			// SinglePattern (for item in list) — use haira.ToSlice() for safe slice iteration
			switch forStmt.Iterator.Node.(type) {
			case ast.ListExpr, ast.MapExpr:
				// Already typed, use directly
			default:
				if isTypedNonAny(forStmt.Iterator.Span) {
					// Type checker knows this is a typed slice, use directly
				} else {
					iter = fmt.Sprintf("haira.ToSlice(%s)", iter)
				}
			}
			p := forStmt.Pattern.(ast.SinglePattern)
			em.OpenBlock(fmt.Sprintf("for _, %s := range %s", p.Name.Node, iter))
		}
	}
	EmitBlockBody(em, forStmt.Body)
	em.CloseBlock()
}

func forPatternVar(pattern ast.ForPattern) string {
	switch p := pattern.(type) {
	case ast.SinglePattern:
		return p.Name.Node
	case ast.PairPattern:
		return p.Second.Node
	}
	return "i"
}

func emitWhile(em *GoEmitter, whileStmt ast.WhileStmt) {
	cond := ExprToGo(whileStmt.Condition)
	em.OpenBlock(fmt.Sprintf("for %s", cond))
	EmitBlockBody(em, whileStmt.Body)
	em.CloseBlock()
}

func emitReturn(em *GoEmitter, ret ast.ReturnStmt) {
	if len(ret.Values) == 0 {
		em.Line("return")
		return
	}
	vals := make([]string, len(ret.Values))
	for i, v := range ret.Values {
		vals[i] = ExprToGo(v)
	}
	em.Line("return " + strings.Join(vals, ", "))
}

func emitMatch(em *GoEmitter, matchExpr ast.MatchExpr) {
	subject := ExprToGo(matchExpr.Subject)
	em.OpenBlock(fmt.Sprintf("switch %s", subject))
	for _, arm := range matchExpr.Arms {
		emitMatchArmHeader(em, arm.Pattern.Node)
		em.Indent()
		switch body := arm.Body.(type) {
		case ast.MatchArmExpr:
			em.Line(ExprToGo(body.Value))
		case ast.MatchArmBlock:
			EmitBlockBody(em, body.Value)
		}
		em.Dedent()
	}
	em.CloseBlock()
}

func emitMatchArmHeader(em *GoEmitter, pattern ast.Pattern) {
	if _, ok := pattern.(ast.WildcardPattern); ok {
		em.Line("default:")
	} else {
		em.Line(fmt.Sprintf("case %s:", patternToGo(pattern)))
	}
}

func patternToGo(pattern ast.Pattern) string {
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

func emitTry(em *GoEmitter, tryStmt ast.TryStmt) {
	errName := tryStmt.ErrorName.Node
	em.OpenBlock("func()")
	em.OpenBlock("defer func()")
	em.OpenBlock("if r := recover(); r != nil")
	em.Line(fmt.Sprintf("%s := fmt.Sprintf(\"%%v\", r)", errName))
	em.Line(fmt.Sprintf("_ = %s", errName))
	EmitBlockBody(em, tryStmt.CatchBody)
	em.CloseBlock()
	em.CloseBlock()
	em.Line("()")
	EmitBlockBody(em, tryStmt.Body)
	em.CloseBlock()
	em.Line("()")
}

// EmitBlockBody emits the body of a block (just the statements, no braces).
func EmitBlockBody(em *GoEmitter, block ast.Block) {
	for _, stmt := range block.Statements {
		EmitStatement(em, stmt)
	}
}

// EmitToolBody emits tool body — wraps `return X` → `return X, nil`.
func EmitToolBody(em *GoEmitter, block ast.Block) {
	for _, stmt := range block.Statements {
		switch s := stmt.Node.(type) {
		case ast.ReturnStmt:
			if len(s.Values) == 0 {
				em.Line("return nil, nil")
			} else {
				vals := make([]string, len(s.Values))
				for i, v := range s.Values {
					vals[i] = ExprToGo(v)
				}
				em.Line(fmt.Sprintf("return %s, nil", strings.Join(vals, ", ")))
			}
		case ast.IfStmt:
			emitToolIf(em, s)
		default:
			EmitStatement(em, stmt)
		}
	}
}

func emitToolIf(em *GoEmitter, ifStmt ast.IfStmt) {
	cond := ExprToGo(ifStmt.Condition)
	em.OpenBlock(fmt.Sprintf("if %s", cond))
	EmitToolBody(em, ifStmt.ThenBranch)
	if ifStmt.ElseBranch != nil {
		switch eb := ifStmt.ElseBranch.(type) {
		case *ast.ElseBlock:
			em.Dedent()
			em.Line("} else {")
			em.Indent()
			EmitToolBody(em, eb.Body)
			em.CloseBlock()
			return
		case *ast.ElseIf:
			em.Dedent()
			cond := ExprToGo(eb.If.Node.Condition)
			em.Line(fmt.Sprintf("} else if %s {", cond))
			em.Indent()
			EmitToolBody(em, eb.If.Node.ThenBranch)
			em.CloseBlock()
			return
		}
	}
	em.CloseBlock()
}
