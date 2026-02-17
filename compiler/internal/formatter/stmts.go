package formatter

import (
	"github.com/haira-lang/haira/internal/ast"
)

// formatBlock formats a { ... } block with indentation.
func (f *Formatter) formatBlock(block ast.Block) {
	f.writeln("{")
	f.incIndent()
	for _, stmt := range block.Statements {
		f.formatStatement(stmt)
	}
	f.decIndent()
	f.writeIndent()
	f.write("}")
}

// formatBlockStatements formats just the statements inside a block (no braces).
func (f *Formatter) formatBlockStatements(stmts []ast.Statement) {
	for _, stmt := range stmts {
		f.formatStatement(stmt)
	}
}

// formatStatement formats a single statement with indentation and newline.
func (f *Formatter) formatStatement(stmt ast.Statement) {
	f.emitCommentsBefore(stmt.Span.Start)
	switch s := stmt.Node.(type) {
	case ast.AssignStmt:
		f.formatAssign(s)
	case ast.IfStmt:
		f.writeIndent()
		f.formatIfStmt(s)
		f.newline()
	case ast.ForStmt:
		f.writeIndent()
		f.formatForStmt(s)
		f.newline()
	case ast.WhileStmt:
		f.writeIndent()
		f.formatWhileStmt(s)
		f.newline()
	case ast.ReturnStmt:
		f.writeIndent()
		f.formatReturnStmt(s)
		f.newline()
	case ast.TryStmt:
		f.writeIndent()
		f.formatTryStmt(s)
		f.newline()
	case ast.DeferStmt:
		f.writeIndent()
		f.write("defer ")
		f.formatExpr(s.Value)
		f.newline()
	case ast.ErrDeferStmt:
		f.writeIndent()
		f.write("errdefer ")
		f.formatExpr(s.Value)
		f.newline()
	case ast.MatchStmt:
		f.writeIndent()
		f.formatMatchExpr(s.Match)
		f.newline()
	case ast.BreakStmt:
		f.writeIndent()
		f.writeln("break")
	case ast.ContinueStmt:
		f.writeIndent()
		f.writeln("continue")
	case ast.ExprStmt:
		f.writeIndent()
		f.formatExpr(s.Value)
		f.newline()
	case ast.StepStmt:
		f.writeIndent()
		f.formatStepStmt(s)
		f.newline()
	}
}

func (f *Formatter) formatAssign(s ast.AssignStmt) {
	f.writeIndent()
	for i, target := range s.Targets {
		if i > 0 {
			f.write(", ")
		}
		f.formatAssignPath(target.Path)
		f.formatColonType(target.Ty)
	}
	f.write(" = ")
	f.formatExpr(s.Value)
	f.newline()
}

func (f *Formatter) formatAssignPath(path ast.AssignPath) {
	switch p := path.(type) {
	case ast.IdentPath:
		f.write(p.Name.Node)
	case ast.FieldPath:
		f.formatAssignPath(p.Object)
		f.write(".")
		f.write(p.Field.Node)
	case ast.IndexPath:
		f.formatAssignPath(p.Object)
		f.write("[")
		f.formatExpr(p.Index)
		f.write("]")
	}
}

func (f *Formatter) formatIfStmt(s ast.IfStmt) {
	f.write("if ")
	f.formatExpr(s.Condition)
	f.write(" ")
	f.formatBlock(s.ThenBranch)
	if s.ElseBranch != nil {
		switch eb := s.ElseBranch.(type) {
		case *ast.ElseBlock:
			f.write(" else ")
			f.formatBlock(eb.Body)
		case *ast.ElseIf:
			f.write(" else ")
			f.formatIfStmt(eb.If.Node)
		}
	}
}

func (f *Formatter) formatForStmt(s ast.ForStmt) {
	f.write("for ")
	switch p := s.Pattern.(type) {
	case ast.SinglePattern:
		f.write(p.Name.Node)
	case ast.PairPattern:
		f.write(p.First.Node)
		f.write(", ")
		f.write(p.Second.Node)
	}
	f.write(" in ")
	f.formatExpr(s.Iterator)
	f.write(" ")
	f.formatBlock(s.Body)
}

func (f *Formatter) formatWhileStmt(s ast.WhileStmt) {
	f.write("while ")
	f.formatExpr(s.Condition)
	f.write(" ")
	f.formatBlock(s.Body)
}

func (f *Formatter) formatReturnStmt(s ast.ReturnStmt) {
	if len(s.Values) == 0 {
		f.write("return")
		return
	}
	f.write("return ")
	for i, v := range s.Values {
		if i > 0 {
			f.write(", ")
		}
		f.formatExpr(v)
	}
}

func (f *Formatter) formatTryStmt(s ast.TryStmt) {
	f.write("try ")
	f.formatBlock(s.Body)
	f.write(" catch ")
	f.write(s.ErrorName.Node)
	f.write(" ")
	f.formatBlock(s.CatchBody)
}

func (f *Formatter) formatStepStmt(s ast.StepStmt) {
	// Step decorators
	for _, dec := range s.Decorators {
		f.writeIndent()
		f.formatDecorator(dec)
		f.newline()
	}
	f.write("step ")
	f.write(`"`)
	f.write(s.Name.Node)
	f.write(`"`)
	f.writeln(" {")
	f.incIndent()
	f.formatBlockStatements(s.Body)
	// Step lifecycle hooks
	for _, hook := range s.Hooks {
		f.formatLifecycleHook(hook)
	}
	f.decIndent()
	f.writeIndent()
	f.write("}")
}

func (f *Formatter) formatLifecycleHook(hook ast.LifecycleHook) {
	f.writeIndent()
	switch hook.Kind {
	case ast.HookOnerror:
		f.write("onerror")
		if hook.ErrName != "" {
			f.write(" ")
			f.write(hook.ErrName)
		}
	case ast.HookOnsuccess:
		f.write("onsuccess")
		if hook.ArgName != "" {
			f.write(" ")
			f.write(hook.ArgName)
		}
	case ast.HookOncancel:
		f.write("oncancel")
	}
	f.write(" ")
	f.formatBlock(hook.Body)
	f.newline()
}

func (f *Formatter) formatDecorator(dec ast.Decorator) {
	f.write("@")
	f.write(dec.Name.Node)
	if len(dec.Args) > 0 {
		f.write("(")
		for i, arg := range dec.Args {
			if i > 0 {
				f.write(", ")
			}
			// Detect single-entry map used for named decorator args
			// and emit as key: value instead of { key: value }
			if m, ok := arg.Node.(ast.MapExpr); ok && len(m.Entries) == 1 {
				f.formatExpr(m.Entries[0].Key)
				f.write(": ")
				f.formatExpr(m.Entries[0].Value)
			} else {
				f.formatExpr(arg)
			}
		}
		f.write(")")
	}
}
