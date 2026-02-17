package checker

import (
	"fmt"
	"strings"

	"github.com/haira-lang/haira/internal/ast"
	hairaerr "github.com/haira-lang/haira/internal/errors"
)

// TypeInfo holds inferred types keyed by AST span.
type TypeInfo struct {
	// ExprTypes maps expression spans to their inferred type.
	ExprTypes map[ast.Span]Type
	// VarTypes maps variable name spans to their inferred type.
	VarTypes map[ast.Span]Type
}

// Check performs type checking on a source file and returns type info + diagnostics.
func Check(file *ast.SourceFile) (*TypeInfo, []hairaerr.Diagnostic) {
	c := &checker{
		env: NewEnv(),
		info: &TypeInfo{
			ExprTypes: make(map[ast.Span]Type),
			VarTypes:  make(map[ast.Span]Type),
		},
		agents:    make(map[string]bool),
		providers: make(map[string]bool),
	}

	// Pass 1: register all global declarations
	c.registerGlobals(file)

	// Pass 2: check function bodies
	c.checkBodies(file)

	return c.info, c.diags
}

type checker struct {
	env       *Env
	info      *TypeInfo
	diags     []hairaerr.Diagnostic
	file      string
	inMethod  bool            // true when checking a method body (self is protected)
	returnTy  Type            // expected return type for current function/workflow/tool (nil = not set)
	agents    map[string]bool // registered agent names
	providers map[string]bool // registered provider names
}

func (c *checker) addError(msg string, span ast.Span) {
	c.diags = append(c.diags, hairaerr.Diagnostic{
		Level:   hairaerr.Error,
		Message: msg,
		Span:    span,
		File:    c.file,
	})
}

func (c *checker) addWarning(msg string, span ast.Span, hint string) {
	c.diags = append(c.diags, hairaerr.Diagnostic{
		Level:   hairaerr.Warning,
		Message: msg,
		Span:    span,
		File:    c.file,
		Hint:    hint,
	})
}

func (c *checker) checkEnumExhaustiveness(enumTy EnumType, arms []ast.MatchArm, span ast.Span) {
	hasWildcard := false
	covered := map[string]bool{}
	for _, arm := range arms {
		collectCoveredVariants(arm.Pattern.Node, enumTy.Name, covered, &hasWildcard)
	}
	if hasWildcard {
		return
	}
	var missing []string
	for _, v := range enumTy.Variants {
		fullName := enumTy.Name + v
		if !covered[fullName] {
			missing = append(missing, v)
		}
	}
	if len(missing) > 0 {
		hint := "add a wildcard '_' arm or handle all variants"
		msg := fmt.Sprintf("non-exhaustive match on %s: missing %s", enumTy.Name, strings.Join(missing, ", "))
		c.addWarning(msg, span, hint)
	}
}

func collectCoveredVariants(p ast.Pattern, enumName string, covered map[string]bool, hasWildcard *bool) {
	switch pat := p.(type) {
	case ast.WildcardPattern:
		*hasWildcard = true
	case ast.IdentPattern:
		covered[pat.Name] = true
	case ast.OrPattern:
		for _, sub := range pat.Patterns {
			collectCoveredVariants(sub.Node, enumName, covered, hasWildcard)
		}
	}
}

// ---------------------------------------------------------------------------
// Pass 1: Register global declarations
// ---------------------------------------------------------------------------

func (c *checker) registerGlobals(file *ast.SourceFile) {
	for _, item := range file.Items {
		switch it := item.Node.(type) {
		case ast.TypeDef:
			fields := make(map[string]Type)
			for _, f := range it.Fields {
				fields[f.Name.Node] = c.resolveTypeExpr(f.Ty)
			}
			c.env.DefineType(it.Name.Node, StructType{Name: it.Name.Node, Fields: fields})

		case ast.EnumDef:
			var variants []string
			for _, v := range it.Variants {
				variants = append(variants, v.Name.Node)
			}
			c.env.DefineType(it.Name.Node, EnumType{Name: it.Name.Node, Variants: variants})

		case ast.FunctionDef:
			params := make([]Type, len(it.Params))
			for i, p := range it.Params {
				params[i] = c.resolveTypeExpr(p.Ty)
			}
			ret := Type(VoidType{})
			if it.ReturnTy != nil {
				ret = c.resolveASTType(it.ReturnTy.Node)
			}
			c.env.DefineFunc(it.Name.Node, &FuncType{Params: params, Return: ret})

		case ast.ToolDecl:
			params := make([]Type, len(it.Params))
			for i, p := range it.Params {
				params[i] = c.resolveTypeExpr(p.Ty)
			}
			ret := Type(AnyType{})
			if it.ReturnTy != nil {
				ret = c.resolveASTType(it.ReturnTy.Node)
			}
			c.env.DefineFunc(it.Name.Node, &FuncType{Params: params, Return: ret})

		case ast.WorkflowDecl:
			params := make([]Type, len(it.Params))
			for i, p := range it.Params {
				params[i] = c.resolveTypeExpr(p.Ty)
			}
			ret := Type(AnyType{})
			if it.ReturnTy != nil {
				ret = c.resolveASTType(it.ReturnTy.Node)
			}
			c.env.DefineFunc(it.Name.Node, &FuncType{Params: params, Return: ret})

		case ast.ProviderDecl:
			c.providers[it.Name.Node] = true
			c.checkProviderFields(it)

		case ast.AgentDecl:
			c.agents[it.Name.Node] = true
			c.checkAgentFields(it)
		}
	}
}

// ---------------------------------------------------------------------------
// Agent/Provider field validation
// ---------------------------------------------------------------------------

var validProviderFields = map[string]bool{
	"api_key": true, "model": true, "endpoint": true, "api_version": true,
	"backend": true, "host": true, "temperature": true, "max_tokens": true,
	// MCP provider fields
	"transport": true, "command": true, "args": true, "env": true, "headers": true,
}

var validAgentFields = map[string]bool{
	"model": true, "system": true, "tools": true, "handoffs": true,
	"mcp": true, "temperature": true, "max_tokens": true, "max_steps": true,
	"memory": true, "output": true,
}

func (c *checker) checkProviderFields(provider ast.ProviderDecl) {
	for _, field := range provider.Fields {
		if !validProviderFields[field.Key.Node] {
			c.addWarning(
				fmt.Sprintf("unknown provider field %q", field.Key.Node),
				field.Key.Span,
				"valid fields: api_key, model, endpoint, api_version, backend, host, temperature, max_tokens",
			)
		}
	}
}

func (c *checker) checkAgentFields(agent ast.AgentDecl) {
	hasModel := false
	for _, field := range agent.Fields {
		if !validAgentFields[field.Key.Node] {
			c.addWarning(
				fmt.Sprintf("unknown agent field %q", field.Key.Node),
				field.Key.Span,
				"valid fields: model, system, tools, handoffs, mcp, temperature, max_tokens, max_steps, memory, output",
			)
		}
		if field.Key.Node == "model" {
			hasModel = true
			// Verify model references a known provider
			if ident, ok := field.Value.Node.(ast.IdentExpr); ok {
				if !c.providers[ident.Name] {
					c.addError(
						fmt.Sprintf("agent %q references unknown provider %q", agent.Name.Node, ident.Name),
						field.Value.Span,
					)
				}
			}
		}
		if field.Key.Node == "tools" {
			// Verify tool references are known functions
			if list, ok := field.Value.Node.(ast.ListExpr); ok {
				for _, item := range list.Elems {
					if ident, ok := item.Node.(ast.IdentExpr); ok {
						if _, ok := c.env.LookupFunc(ident.Name); !ok {
							c.addError(
								fmt.Sprintf("agent %q references unknown tool %q", agent.Name.Node, ident.Name),
								item.Span,
							)
						}
					}
				}
			}
		}
		if field.Key.Node == "handoffs" {
			// Verify handoff references are known agents
			if list, ok := field.Value.Node.(ast.ListExpr); ok {
				for _, item := range list.Elems {
					if ident, ok := item.Node.(ast.IdentExpr); ok {
						if !c.agents[ident.Name] {
							c.addWarning(
								fmt.Sprintf("agent %q handoff target %q not yet declared", agent.Name.Node, ident.Name),
								item.Span,
								"ensure the agent is declared in this file",
							)
						}
					}
				}
			}
		}
	}
	if !hasModel {
		c.addError(
			fmt.Sprintf("agent %q is missing required field 'model'", agent.Name.Node),
			agent.Name.Span,
		)
	}
}

// ---------------------------------------------------------------------------
// Pass 2: Check function bodies
// ---------------------------------------------------------------------------

func (c *checker) checkBodies(file *ast.SourceFile) {
	for _, item := range file.Items {
		switch it := item.Node.(type) {
		case ast.FunctionDef:
			c.checkFunction(it)
		case ast.MethodDef:
			c.checkMethod(it)
		case ast.ToolDecl:
			if it.Body != nil {
				c.checkToolBody(it)
			}
		case ast.WorkflowDecl:
			c.checkWorkflow(it)
		}
	}
}

func (c *checker) checkFunction(fn ast.FunctionDef) {
	env := c.env.Child()
	for _, p := range fn.Params {
		ty := c.resolveTypeExpr(p.Ty)
		env.DefineVar(p.Name.Node, ty)
	}
	saved := c.env
	savedReturn := c.returnTy
	c.env = env
	if fn.ReturnTy != nil {
		c.returnTy = c.resolveASTType(fn.ReturnTy.Node)
	} else {
		c.returnTy = nil
	}
	c.checkBlock(fn.Body)
	c.env = saved
	c.returnTy = savedReturn
}

func (c *checker) checkMethod(md ast.MethodDef) {
	env := c.env.Child()
	// Register self
	if ty, ok := c.env.LookupType(md.TypeName.Node); ok {
		env.DefineVar("self", ty)
	}
	for _, p := range md.Params {
		ty := c.resolveTypeExpr(p.Ty)
		env.DefineVar(p.Name.Node, ty)
	}
	saved := c.env
	savedInMethod := c.inMethod
	savedReturn := c.returnTy
	c.env = env
	c.inMethod = true
	if md.ReturnTy != nil {
		c.returnTy = c.resolveASTType(md.ReturnTy.Node)
	} else {
		c.returnTy = nil
	}
	c.checkBlock(md.Body)
	c.env = saved
	c.inMethod = savedInMethod
	c.returnTy = savedReturn
}

func (c *checker) checkToolBody(tool ast.ToolDecl) {
	env := c.env.Child()
	for _, p := range tool.Params {
		ty := c.resolveTypeExpr(p.Ty)
		env.DefineVar(p.Name.Node, ty)
	}
	saved := c.env
	savedReturn := c.returnTy
	c.env = env
	if tool.ReturnTy != nil {
		c.returnTy = c.resolveASTType(tool.ReturnTy.Node)
	} else {
		c.returnTy = nil
	}
	c.checkBlock(*tool.Body)
	c.env = saved
	c.returnTy = savedReturn
}

func (c *checker) checkWorkflow(wf ast.WorkflowDecl) {
	env := c.env.Child()
	for _, p := range wf.Params {
		ty := c.resolveTypeExpr(p.Ty)
		env.DefineVar(p.Name.Node, ty)
	}
	saved := c.env
	savedReturn := c.returnTy
	c.env = env
	if wf.ReturnTy != nil {
		c.returnTy = c.resolveASTType(wf.ReturnTy.Node)
	} else {
		c.returnTy = nil
	}
	c.checkBlock(wf.Body)
	// Check workflow-level lifecycle hooks
	for _, hook := range wf.Hooks {
		hookEnv := c.env.Child()
		if hook.Kind == ast.HookOnerror && hook.ErrName != "" {
			hookEnv.DefineVar(hook.ErrName, StringType{})
		}
		if hook.Kind == ast.HookOnsuccess && hook.ArgName != "" {
			hookEnv.DefineVar(hook.ArgName, AnyType{})
		}
		savedInner := c.env
		c.env = hookEnv
		c.checkBlock(hook.Body)
		c.env = savedInner
	}
	c.env = saved
	c.returnTy = savedReturn
}

// ---------------------------------------------------------------------------
// Statement checking
// ---------------------------------------------------------------------------

func (c *checker) checkBlock(block ast.Block) {
	for _, stmt := range block.Statements {
		c.checkStmt(stmt)
	}
}

func (c *checker) checkStmt(stmt ast.Statement) {
	switch s := stmt.Node.(type) {
	case ast.AssignStmt:
		rhsType := c.inferExpr(s.Value)
		for _, target := range s.Targets {
			if ident, ok := target.Path.(ast.IdentPath); ok {
				if c.inMethod && ident.Name.Node == "self" {
					c.addError("cannot reassign self in method", ident.Name.Span)
				}
				if target.Ty != nil {
					// Annotated type — check compatibility
					annotated := c.resolveASTType(target.Ty.Node)
					if !TypeEquals(annotated, rhsType) && !isAny(annotated) && !isAny(rhsType) {
						c.addError(
							"type mismatch: expected "+annotated.String()+", got "+rhsType.String(),
							s.Value.Span,
						)
					}
					c.env.DefineVar(ident.Name.Node, annotated)
				} else {
					c.env.DefineVar(ident.Name.Node, rhsType)
				}
				c.info.VarTypes[ident.Name.Span] = rhsType
			}
		}

	case ast.ExprStmt:
		c.inferExpr(s.Value)

	case ast.IfStmt:
		condType := c.inferExpr(s.Condition)
		if !isAny(condType) && !TypeEquals(condType, BoolType{}) {
			c.addWarning("condition should be bool, got "+condType.String(), s.Condition.Span, "")
		}
		c.checkBlock(s.ThenBranch)
		if s.ElseBranch != nil {
			switch eb := s.ElseBranch.(type) {
			case *ast.ElseBlock:
				c.checkBlock(eb.Body)
			case *ast.ElseIf:
				c.checkStmt(ast.Statement{
					Node: eb.If.Node,
					Span: eb.If.Span,
				})
			}
		}

	case ast.ForStmt:
		c.inferExpr(s.Iterator)
		env := c.env.Child()
		switch p := s.Pattern.(type) {
		case ast.SinglePattern:
			env.DefineVar(p.Name.Node, AnyType{})
		case ast.PairPattern:
			env.DefineVar(p.First.Node, AnyType{})
			env.DefineVar(p.Second.Node, AnyType{})
		}
		saved := c.env
		c.env = env
		c.checkBlock(s.Body)
		c.env = saved

	case ast.WhileStmt:
		c.inferExpr(s.Condition)
		c.checkBlock(s.Body)

	case ast.ReturnStmt:
		if len(s.Values) > 0 {
			retType := c.inferExpr(s.Values[0])
			for _, v := range s.Values[1:] {
				c.inferExpr(v)
			}
			// Check against declared return type
			if c.returnTy != nil && !isAny(c.returnTy) && !isAny(retType) {
				if !TypeEquals(c.returnTy, retType) {
					c.addWarning(
						"return type mismatch: expected "+c.returnTy.String()+", got "+retType.String(),
						s.Values[0].Span,
						"",
					)
				}
			}
		} else {
			// Empty return in a function with declared return type
			if c.returnTy != nil && !isAny(c.returnTy) && !TypeEquals(c.returnTy, VoidType{}) {
				c.addWarning(
					"empty return in function expecting "+c.returnTy.String(),
					stmt.Span,
					"",
				)
			}
		}

	case ast.TryStmt:
		c.checkBlock(s.Body)
		env := c.env.Child()
		env.DefineVar(s.ErrorName.Node, StringType{})
		saved := c.env
		c.env = env
		c.checkBlock(s.CatchBody)
		c.env = saved

	case ast.DeferStmt:
		c.inferExpr(s.Value)

	case ast.MatchStmt:
		subjectType := c.inferExpr(s.Match.Subject)
		for _, arm := range s.Match.Arms {
			if arm.Guard != nil {
				c.inferExpr(*arm.Guard)
			}
			switch body := arm.Body.(type) {
			case ast.MatchArmExpr:
				c.inferExpr(body.Value)
			case ast.MatchArmBlock:
				c.checkBlock(body.Value)
			}
		}
		// Exhaustiveness check for simple enums
		if enumTy, ok := subjectType.(EnumType); ok {
			c.checkEnumExhaustiveness(enumTy, s.Match.Arms, s.Match.Subject.Span)
		}

	case ast.StepStmt:
		// Step body shares the enclosing workflow scope (variables flow through)
		for _, stmt := range s.Body {
			c.checkStmt(stmt)
		}
		// Check lifecycle hooks
		for _, hook := range s.Hooks {
			env := c.env.Child()
			if hook.Kind == ast.HookOnerror && hook.ErrName != "" {
				env.DefineVar(hook.ErrName, StringType{})
			}
			saved := c.env
			c.env = env
			c.checkBlock(hook.Body)
			c.env = saved
		}

	case ast.ErrDeferStmt:
		c.inferExpr(s.Value)
	}
}

// ---------------------------------------------------------------------------
// Expression type inference
// ---------------------------------------------------------------------------

func (c *checker) inferExpr(expr ast.Expr) Type {
	var ty Type

	switch e := expr.Node.(type) {
	case ast.LiteralExpr:
		ty = c.inferLiteral(e.Lit)

	case ast.IdentExpr:
		if t, ok := c.env.LookupVar(e.Name); ok {
			ty = t
		} else if _, ok := c.env.LookupFunc(e.Name); ok {
			ty = AnyType{} // function reference (e.g. passed as callback)
		} else if _, ok := c.env.LookupType(e.Name); ok {
			ty = AnyType{} // type reference (e.g. enum variant prefix)
		} else if c.agents[e.Name] || c.providers[e.Name] {
			ty = AnyType{} // agent or provider reference
		} else if isStdlibModule(e.Name) {
			ty = AnyType{} // stdlib module qualifier (e.g. io.println)
		} else {
			c.addWarning("undefined variable '"+e.Name+"'", expr.Span, "")
			ty = AnyType{}
		}

	case ast.BinaryExpr:
		left := c.inferExpr(e.Left)
		right := c.inferExpr(e.Right)
		ty = c.inferBinaryOp(e.Op.Node, left, right, expr.Span)

	case ast.UnaryExpr:
		operand := c.inferExpr(e.Operand)
		switch e.Op.Node {
		case ast.OpNeg:
			if TypeEquals(operand, IntType{}) || TypeEquals(operand, FloatType{}) || isAny(operand) {
				ty = operand
			} else {
				c.addError("cannot negate "+operand.String(), expr.Span)
				ty = AnyType{}
			}
		case ast.OpNot:
			ty = BoolType{}
		default:
			ty = operand
		}

	case ast.CallExpr:
		ty = c.inferCall(e, expr.Span)

	case ast.MethodCallExpr:
		c.inferExpr(e.Receiver)
		for _, arg := range e.Args {
			c.inferExpr(arg.Value)
		}
		ty = AnyType{}

	case ast.FieldExpr:
		objType := c.inferExpr(e.Object)
		if st, ok := objType.(StructType); ok {
			if ft, ok := st.Fields[e.Field.Node]; ok {
				ty = ft
			} else {
				c.addError("unknown field '"+e.Field.Node+"' on type "+st.Name, e.Field.Span)
				ty = AnyType{}
			}
		} else {
			ty = AnyType{}
		}

	case ast.IndexExpr:
		objType := c.inferExpr(e.Object)
		c.inferExpr(e.Index)
		switch ot := objType.(type) {
		case ListType:
			ty = ot.Elem
		case MapType:
			ty = ot.Value
		default:
			ty = AnyType{}
		}

	case ast.ListExpr:
		if len(e.Elems) == 0 {
			ty = ListType{Elem: AnyType{}}
		} else {
			elemType := c.inferExpr(e.Elems[0])
			allSame := true
			for _, el := range e.Elems[1:] {
				et := c.inferExpr(el)
				if !TypeEquals(elemType, et) {
					allSame = false
				}
			}
			if allSame && !isAny(elemType) {
				ty = ListType{Elem: elemType}
			} else {
				ty = ListType{Elem: AnyType{}}
			}
		}

	case ast.MapExpr:
		if len(e.Entries) == 0 {
			ty = MapType{Key: StringType{}, Value: AnyType{}}
		} else {
			valType := c.inferExpr(e.Entries[0].Value)
			c.inferExpr(e.Entries[0].Key)
			allSame := true
			for _, entry := range e.Entries[1:] {
				c.inferExpr(entry.Key)
				vt := c.inferExpr(entry.Value)
				if !TypeEquals(valType, vt) {
					allSame = false
				}
			}
			if allSame && !isAny(valType) {
				ty = MapType{Key: StringType{}, Value: valType}
			} else {
				ty = MapType{Key: StringType{}, Value: AnyType{}}
			}
		}

	case ast.InstanceExpr:
		if t, ok := c.env.LookupType(e.TypeName.Node); ok {
			ty = t
		} else {
			ty = AnyType{}
		}
		for _, f := range e.Fields {
			c.inferExpr(f.Value)
		}

	case ast.PipeExpr:
		c.inferExpr(e.Left)
		ty = c.inferExpr(e.Right)

	case ast.LambdaExpr:
		ty = AnyType{} // lambdas are untyped for now

	case ast.ParenExpr:
		ty = c.inferExpr(e.Inner)

	case ast.NoneExpr:
		ty = AnyType{}

	case ast.SomeExpr:
		ty = c.inferExpr(e.Inner)

	case ast.IfExpr:
		c.inferExpr(e.If.Condition)
		c.checkBlock(e.If.ThenBranch)
		ty = AnyType{}

	case ast.MatchExpr:
		c.inferExpr(e.Subject)
		for _, arm := range e.Arms {
			switch body := arm.Body.(type) {
			case ast.MatchArmExpr:
				c.inferExpr(body.Value)
			case ast.MatchArmBlock:
				c.checkBlock(body.Value)
			}
		}
		ty = AnyType{}

	case ast.RangeExpr:
		c.inferExpr(e.Start)
		c.inferExpr(e.End)
		ty = ListType{Elem: IntType{}}

	case ast.PropagateExpr:
		ty = c.inferExpr(e.Inner)

	case ast.OrelseExpr:
		c.inferExpr(e.Left)
		ty = c.inferExpr(e.Default)

	case ast.SpawnExpr:
		c.checkBlock(e.Body)
		ty = ListType{Elem: AnyType{}}

	case ast.AsyncExpr:
		c.checkBlock(e.Body)
		ty = AnyType{}

	case ast.BlockExpr:
		c.checkBlock(e.Body)
		ty = AnyType{}

	default:
		ty = AnyType{}
	}

	c.info.ExprTypes[expr.Span] = ty
	return ty
}

func (c *checker) inferLiteral(lit ast.Literal) Type {
	switch lit.(type) {
	case ast.IntLit:
		return IntType{}
	case ast.FloatLit:
		return FloatType{}
	case ast.StringLit:
		return StringType{}
	case ast.BoolLit:
		return BoolType{}
	case ast.InterpolatedStringLit:
		return StringType{}
	}
	return AnyType{}
}

func (c *checker) inferBinaryOp(op ast.BinaryOp, left, right Type, span ast.Span) Type {
	switch op {
	case ast.OpAdd, ast.OpSub, ast.OpMul, ast.OpDiv, ast.OpMod:
		if isAny(left) || isAny(right) {
			return AnyType{}
		}
		if TypeEquals(left, StringType{}) && op == ast.OpAdd {
			return StringType{}
		}
		if TypeEquals(left, right) {
			return left
		}
		// int + float → float
		if (TypeEquals(left, IntType{}) && TypeEquals(right, FloatType{})) ||
			(TypeEquals(left, FloatType{}) && TypeEquals(right, IntType{})) {
			return FloatType{}
		}
		return AnyType{}

	case ast.OpEq, ast.OpNe, ast.OpLt, ast.OpGt, ast.OpLe, ast.OpGe:
		return BoolType{}

	case ast.OpAnd, ast.OpOr:
		return BoolType{}
	}
	return AnyType{}
}

func (c *checker) inferCall(call ast.CallExpr, span ast.Span) Type {
	// Check argument types
	for _, arg := range call.Args {
		c.inferExpr(arg.Value)
	}

	// Qualified call: module.function
	if field, ok := call.Callee.Node.(ast.FieldExpr); ok {
		if ident, ok := field.Object.Node.(ast.IdentExpr); ok {
			key := ident.Name + "." + field.Field.Node
			if fn, ok := c.env.LookupFunc(key); ok {
				return fn.Return
			}
			// Don't warn for agent/provider method calls or stdlib modules
			if !c.agents[ident.Name] && !c.providers[ident.Name] && !isStdlibModule(ident.Name) {
				if _, isVar := c.env.LookupVar(ident.Name); !isVar {
					if _, isType := c.env.LookupType(ident.Name); !isType {
						c.addWarning("undefined function '"+key+"'", span, "")
					}
				}
			}
		}
		return AnyType{}
	}

	// Bare function call
	if ident, ok := call.Callee.Node.(ast.IdentExpr); ok {
		if fn, ok := c.env.LookupFunc(ident.Name); ok {
			return fn.Return
		}
		// Don't warn for known types used as constructors
		if _, ok := c.env.LookupType(ident.Name); !ok {
			c.addWarning("undefined function '"+ident.Name+"'", span, "")
		}
		return AnyType{}
	}

	return AnyType{}
}

// ---------------------------------------------------------------------------
// Type resolution helpers
// ---------------------------------------------------------------------------

func (c *checker) resolveTypeExpr(ty *ast.Spanned[ast.Type]) Type {
	if ty == nil {
		return AnyType{}
	}
	return c.resolveASTType(ty.Node)
}

func (c *checker) resolveASTType(ty ast.Type) Type {
	switch t := ty.(type) {
	case ast.NamedType:
		switch t.Name {
		case "int":
			return IntType{}
		case "float":
			return FloatType{}
		case "string":
			return StringType{}
		case "bool":
			return BoolType{}
		case "any":
			return AnyType{}
		case "void":
			return VoidType{}
		case "file":
			return StringType{} // file is semantically a string (temp path)
		default:
			if userTy, ok := c.env.LookupType(t.Name); ok {
				return userTy
			}
			return AnyType{}
		}
	case ast.ListType:
		return ListType{Elem: c.resolveASTType(t.Elem.Node)}
	case ast.MapType:
		return MapType{Key: c.resolveASTType(t.Key.Node), Value: c.resolveASTType(t.Value.Node)}
	case ast.FunctionType:
		params := make([]Type, len(t.Params))
		for i, p := range t.Params {
			params[i] = c.resolveASTType(p.Node)
		}
		return FuncType{Params: params, Return: c.resolveASTType(t.Ret.Node)}
	}
	return AnyType{}
}

func isAny(t Type) bool {
	_, ok := t.(AnyType)
	return ok
}

var stdlibModules = map[string]bool{
	"io": true, "http": true, "json": true, "string": true, "regex": true,
	"math": true, "conv": true, "array": true, "map": true, "time": true,
	"env": true, "postgres": true, "slack": true, "excel": true, "log": true,
	"mcp": true, "ui": true, "vector": true, "observe": true, "fs": true,
}

func isStdlibModule(name string) bool {
	return stdlibModules[name]
}
