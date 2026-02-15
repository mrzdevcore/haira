package codegen

import (
	"fmt"
	"strings"

	"github.com/haira-lang/haira/internal/ast"
)

// EmitProvider emits a provider declaration as a Go var.
func EmitProvider(em *GoEmitter, provider ast.ProviderDecl) {
	name := goVarName("provider", provider.Name.Node)
	em.OpenBlock(fmt.Sprintf("var %s = &haira.Provider", name))
	em.Line(fmt.Sprintf("Name: %q,", provider.Name.Node))
	for _, field := range provider.Fields {
		goKey := goFieldName(field.Key.Node)
		goVal := ExprToGo(field.Value)
		em.Line(fmt.Sprintf("%s: %s,", goKey, goVal))
	}
	em.CloseBlock()
	em.Blank()
}

// EmitAgent emits an agent declaration as a Go var + init function.
func EmitAgent(em *GoEmitter, agent ast.AgentDecl) {
	varName := goVarName("agent", agent.Name.Node)
	em.Line(fmt.Sprintf("var %s *haira.Agent", varName))
	em.Blank()

	initName := fmt.Sprintf("initAgent%s", agent.Name.Node)
	em.OpenBlock(fmt.Sprintf("func %s()", initName))

	// Pre-config setup
	for _, field := range agent.Fields {
		switch field.Key.Node {
		case "tools":
			em.Line("toolReg := haira.NewToolRegistry()")
			if list, ok := field.Value.Node.(ast.ListExpr); ok {
				for _, item := range list.Elems {
					if ident, ok := item.Node.(ast.IdentExpr); ok {
						defName := goVarName("toolDef", ident.Name)
						em.Line(fmt.Sprintf("toolReg.Register(%s)", defName))
					}
				}
			}
		case "handoffs":
			if list, ok := field.Value.Node.(ast.ListExpr); ok {
				var refs []string
				for _, item := range list.Elems {
					if ident, ok := item.Node.(ast.IdentExpr); ok {
						refs = append(refs, "agent"+SnakeToPascal(ident.Name))
					}
				}
				em.Line(fmt.Sprintf("handoffTargets := []*haira.Agent{%s}", strings.Join(refs, ", ")))
			}
		}
	}
	em.Blank()

	em.OpenBlock(fmt.Sprintf("%s = haira.NewAgent(haira.AgentConfig", varName))
	em.Line(fmt.Sprintf("Name: %q,", agent.Name.Node))
	for _, field := range agent.Fields {
		switch field.Key.Node {
		case "model":
			if ident, ok := field.Value.Node.(ast.IdentExpr); ok {
				providerVar := goVarName("provider", ident.Name)
				em.Line(fmt.Sprintf("Provider: %s,", providerVar))
			}
		case "system":
			em.Line(fmt.Sprintf("System: %s,", ExprToGo(field.Value)))
		case "tools":
			em.Line("Tools: toolReg,")
		case "handoffs":
			em.Line("Handoffs: handoffTargets,")
		case "temperature":
			em.Line(fmt.Sprintf("Temperature: %s,", ExprToGo(field.Value)))
		case "memory":
			if call, ok := field.Value.Node.(ast.CallExpr); ok {
				if ident, ok := call.Callee.Node.(ast.IdentExpr); ok {
					config := fmt.Sprintf("haira.MemoryConfig{Kind: %q", ident.Name)
					for _, arg := range call.Args {
						if arg.Name != nil && arg.Name.Node == "max_turns" {
							config += fmt.Sprintf(", MaxTurns: %s", ExprToGo(arg.Value))
						}
					}
					config += "}"
					em.Line(fmt.Sprintf("Memory: %s,", config))
				}
			}
		default:
			goKey := goFieldName(field.Key.Node)
			em.Line(fmt.Sprintf("%s: %s,", goKey, ExprToGo(field.Value)))
		}
	}
	em.Dedent()
	em.Line("})")
	em.CloseBlock()
	em.Blank()
}

// EmitTool emits a tool declaration as a Go handler function + ToolDef var.
func EmitTool(em *GoEmitter, tool ast.ToolDecl) {
	em.ResetVars()
	handlerName := goFuncName("tool", tool.Name.Node)
	defName := goVarName("toolDef", tool.Name.Node)

	// Handler function
	em.OpenBlock(fmt.Sprintf("func %s(args json.RawMessage) (any, error)", handlerName))
	em.Line("var params struct {")
	em.Indent()
	for _, param := range tool.Params {
		goName := SnakeToPascal(param.Name.Node)
		goType := "any"
		if param.Ty != nil {
			goType = HairaTypeToGo(param.Ty.Node)
		}
		jsonTag := param.Name.Node
		em.Line(fmt.Sprintf("%s %s `json:\"%s\"`", goName, goType, jsonTag))
	}
	em.Dedent()
	em.Line("}")
	em.Line("json.Unmarshal(args, &params)")

	// Apply defaults
	for _, param := range tool.Params {
		if param.Default != nil {
			goName := SnakeToPascal(param.Name.Node)
			defaultVal := ExprToGo(*param.Default)
			zeroCheck := fmt.Sprintf("params.%s == 0", goName)
			if param.Ty != nil {
				if named, ok := param.Ty.Node.(ast.NamedType); ok && named.Name == "string" {
					zeroCheck = fmt.Sprintf(`params.%s == ""`, goName)
				}
			}
			em.OpenBlock(fmt.Sprintf("if %s", zeroCheck))
			em.Line(fmt.Sprintf("params.%s = %s", goName, defaultVal))
			em.CloseBlock()
		}
	}

	// Body or stub
	if tool.Body != nil {
		for _, param := range tool.Params {
			goName := SnakeToPascal(param.Name.Node)
			em.Line(fmt.Sprintf("%s := params.%s", param.Name.Node, goName))
		}
		EmitToolBody(em, *tool.Body)
	} else {
		em.Line(fmt.Sprintf("return nil, fmt.Errorf(\"tool %s not yet implemented\")", tool.Name.Node))
	}
	em.CloseBlock()
	em.Blank()

	// ToolDef var
	schema := buildToolJSONSchema(tool.Params)
	em.OpenBlock(fmt.Sprintf("var %s = &haira.ToolDef", defName))
	em.Line(fmt.Sprintf("Name:        %q,", tool.Name.Node))
	if strings.Contains(tool.Description, "\n") {
		em.Line(fmt.Sprintf("Description: `%s`,", tool.Description))
	} else {
		em.Line(fmt.Sprintf("Description: %q,", tool.Description))
	}
	em.Line(fmt.Sprintf("Parameters:  json.RawMessage(`%s`),", schema))
	em.Line(fmt.Sprintf("Handler:     %s,", handlerName))
	em.CloseBlock()
	em.Blank()
}

// EmitWorkflow emits a workflow declaration.
func EmitWorkflow(em *GoEmitter, workflow ast.WorkflowDecl) {
	activeWorkflowName = workflow.Name.Node
	defer func() { activeWorkflowName = "" }()

	handlerName := goFuncName("workflow", workflow.Name.Node)
	defName := goVarName("workflowDef", workflow.Name.Node)
	isStream := isStreamWorkflow(workflow)

	if isStream {
		// Stream handler: returns (<-chan haira.StreamChunk, error)
		em.ResetVars()
		em.OpenBlock(fmt.Sprintf("func %s(params map[string]any) (<-chan haira.StreamChunk, error)", handlerName))
		for _, param := range workflow.Params {
			goType := "string"
			if param.Ty != nil {
				goType = HairaTypeToGo(param.Ty.Node)
			}
			em.Line(fmt.Sprintf("%s, _ := params[%q].(%s)", param.Name.Node, param.Name.Node, goType))
		}
		for _, param := range workflow.Params {
			em.Line(fmt.Sprintf("_ = %s", param.Name.Node))
		}
		emitStreamWorkflowBody(em, workflow.Body)
		em.CloseBlock()
		em.Blank()

		// Also emit a regular handler as fallback (non-streaming clients)
		fallbackName := handlerName + "Fallback"
		em.ResetVars()
		em.OpenBlock(fmt.Sprintf("func %s(params map[string]any) (any, error)", fallbackName))
		for _, param := range workflow.Params {
			goType := "string"
			if param.Ty != nil {
				goType = HairaTypeToGo(param.Ty.Node)
			}
			em.Line(fmt.Sprintf("%s, _ := params[%q].(%s)", param.Name.Node, param.Name.Node, goType))
		}
		for _, param := range workflow.Params {
			em.Line(fmt.Sprintf("_ = %s", param.Name.Node))
		}
		emitStreamFallbackBody(em, workflow.Body)
		em.CloseBlock()
		em.Blank()

		// WorkflowDef var with both handlers
		method, path := extractTriggerInfo(workflow.Trigger)
		em.OpenBlock(fmt.Sprintf("var %s = &haira.WorkflowDef", defName))
		em.Line(fmt.Sprintf("Name: %q,", workflow.Name.Node))
		em.Line(fmt.Sprintf("Method: %q,", method))
		em.Line(fmt.Sprintf("Path: %q,", path))
		emitWorkflowParams(em, workflow.Params)
		em.Line("IsStream: true,")
		emitWorkflowUIMetadata(em, workflow.Decorators)
		em.Line(fmt.Sprintf("Handler: %s,", fallbackName))
		em.Line(fmt.Sprintf("StreamHandler: %s,", handlerName))
		em.CloseBlock()
		em.Blank()
	} else {
		// Regular handler
		em.ResetVars()
		em.OpenBlock(fmt.Sprintf("func %s(params map[string]any) (any, error)", handlerName))
		for _, param := range workflow.Params {
			goType := "string"
			if param.Ty != nil {
				goType = HairaTypeToGo(param.Ty.Node)
			}
			em.Line(fmt.Sprintf("%s, _ := params[%q].(%s)", param.Name.Node, param.Name.Node, goType))
		}
		for _, param := range workflow.Params {
			em.Line(fmt.Sprintf("_ = %s", param.Name.Node))
		}
		if len(workflow.Hooks) > 0 {
			emitWorkflowBodyWithHooks(em, workflow.Body, workflow.Hooks)
		} else {
			emitWorkflowBody(em, workflow.Body)
		}
		em.CloseBlock()
		em.Blank()

		// WorkflowDef var
		method, path := extractTriggerInfo(workflow.Trigger)
		em.OpenBlock(fmt.Sprintf("var %s = &haira.WorkflowDef", defName))
		em.Line(fmt.Sprintf("Name: %q,", workflow.Name.Node))
		em.Line(fmt.Sprintf("Method: %q,", method))
		em.Line(fmt.Sprintf("Path: %q,", path))
		emitWorkflowParams(em, workflow.Params)
		emitWorkflowUIMetadata(em, workflow.Decorators)
		em.Line(fmt.Sprintf("Handler: %s,", handlerName))
		em.CloseBlock()
		em.Blank()
	}
}

func isStreamWorkflow(w ast.WorkflowDecl) bool {
	if w.ReturnTy != nil {
		if named, ok := w.ReturnTy.Node.(ast.NamedType); ok && named.Name == "stream" {
			return true
		}
	}
	return false
}

func emitStreamWorkflowBody(em *GoEmitter, block ast.Block) {
	// For stream workflows, return statements become channel returns
	for _, stmt := range block.Statements {
		switch s := stmt.Node.(type) {
		case ast.ReturnStmt:
			if len(s.Values) > 0 {
				em.Line(fmt.Sprintf("return %s, nil", ExprToGo(s.Values[0])))
			} else {
				em.Line("return nil, nil")
			}
		default:
			EmitStatement(em, stmt)
		}
	}
}

func emitStreamFallbackBody(em *GoEmitter, block ast.Block) {
	// Fallback: call .Ask instead of .Stream for non-SSE clients
	for _, stmt := range block.Statements {
		switch s := stmt.Node.(type) {
		case ast.ReturnStmt:
			if len(s.Values) > 0 {
				// Rewrite agent.stream() → agent.ask() for the fallback
				expr := s.Values[0]
				if mc, ok := expr.Node.(ast.MethodCallExpr); ok && mc.Method.Node == "stream" {
					// Replace .stream with .ask
					mc.Method = ast.Spanned[string]{Node: "ask"}
					rewritten := ast.Expr{Node: mc, Span: expr.Span}
					em.Line(fmt.Sprintf("reply, err := %s", ExprToGo(rewritten)))
					em.OpenBlock("if err != nil")
					em.Line(`return map[string]any{"error": err.Error()}, nil`)
					em.CloseBlock()
					em.Line(`return map[string]any{"reply": reply}, nil`)
					continue
				}
				em.Line(fmt.Sprintf("return %s, nil", ExprToGo(expr)))
			} else {
				em.Line("return nil, nil")
			}
		default:
			EmitStatement(em, stmt)
		}
	}
}

func emitWorkflowBody(em *GoEmitter, block ast.Block) {
	for _, stmt := range block.Statements {
		emitWorkflowStatement(em, stmt)
	}
}

func emitWorkflowStatement(em *GoEmitter, stmt ast.Statement) {
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
		emitWorkflowIf(em, s)
	case ast.MatchStmt:
		emitWorkflowMatch(em, s.Match)
	case ast.StepStmt:
		emitStep(em, s)
	default:
		EmitStatement(em, stmt)
	}
}

// emitWorkflowBodyWithHooks wraps the workflow body with lifecycle hooks.
func emitWorkflowBodyWithHooks(em *GoEmitter, block ast.Block, hooks []ast.LifecycleHook) {
	var onerrorHook *ast.LifecycleHook
	var onsuccessHook *ast.LifecycleHook
	for i := range hooks {
		switch hooks[i].Kind {
		case ast.HookOnerror:
			onerrorHook = &hooks[i]
		case ast.HookOnsuccess:
			onsuccessHook = &hooks[i]
		}
	}

	if onerrorHook != nil {
		// Wrap body in a closure to catch panics from ? operator.
		// Returns inside the closure become _wfResult assignments.
		em.Line("var _wfErr error")
		em.Line("var _wfResult any")
		em.OpenBlock("func()")
		em.OpenBlock("defer func()")
		em.OpenBlock("if r := recover(); r != nil")
		em.Line("_wfErr = fmt.Errorf(\"%v\", r)")
		em.CloseBlock()
		em.Dedent()
		em.Line("}()")
		inCapturedContext = true
		for _, stmt := range block.Statements {
			emitWorkflowStatementCaptured(em, stmt)
		}
		inCapturedContext = false
		em.Dedent()
		em.Line("}()")

		// onerror handler
		em.OpenBlock("if _wfErr != nil")
		em.Line(fmt.Sprintf("%s := fmt.Sprintf(\"%%v\", _wfErr)", onerrorHook.ErrName))
		em.Line(fmt.Sprintf("_ = %s", onerrorHook.ErrName))
		emitWorkflowBody(em, onerrorHook.Body)
		em.CloseBlock()

		// If no error and we have a result, return it
		em.OpenBlock("if _wfResult != nil")
		em.Line("return _wfResult, nil")
		em.CloseBlock()
	} else {
		// No onerror — emit body directly
		for _, stmt := range block.Statements {
			emitWorkflowStatement(em, stmt)
		}
	}

	// onsuccess hook (non-fatal)
	if onsuccessHook != nil {
		em.OpenBlock("if _wfErr == nil")
		em.OpenBlock("func()")
		em.OpenBlock("defer func()")
		em.Line("recover() // onsuccess errors are non-fatal")
		em.Dedent()
		em.Line("}()")
		if onsuccessHook.ArgName != "" {
			em.Line(fmt.Sprintf("%s := _wfResult", onsuccessHook.ArgName))
			em.Line(fmt.Sprintf("_ = %s", onsuccessHook.ArgName))
		}
		EmitBlockBody(em, onsuccessHook.Body)
		em.Dedent()
		em.Line("}()")
		em.CloseBlock()
	}

	// Fallback return — needed when onerror wraps body in IIFE
	if onerrorHook != nil {
		em.Line("return nil, nil")
	}
}

// emitWorkflowStatementCaptured emits a workflow statement inside an onerror IIFE,
// converting returns to _wfResult assignments so they don't try to return from the closure.
func emitWorkflowStatementCaptured(em *GoEmitter, stmt ast.Statement) {
	switch s := stmt.Node.(type) {
	case ast.ReturnStmt:
		if len(s.Values) == 0 {
			em.Line("return")
		} else {
			vals := make([]string, len(s.Values))
			for i, v := range s.Values {
				vals[i] = ExprToGo(v)
			}
			em.Line(fmt.Sprintf("_wfResult = %s", strings.Join(vals, ", ")))
			em.Line("return")
		}
	case ast.IfStmt:
		emitWorkflowIfCaptured(em, s)
	case ast.StepStmt:
		emitStep(em, s)
	default:
		EmitStatement(em, stmt)
	}
}

func emitWorkflowIfCaptured(em *GoEmitter, ifStmt ast.IfStmt) {
	cond := ExprToGo(ifStmt.Condition)
	em.OpenBlock(fmt.Sprintf("if %s", cond))
	for _, stmt := range ifStmt.ThenBranch.Statements {
		emitWorkflowStatementCaptured(em, stmt)
	}
	if ifStmt.ElseBranch != nil {
		switch eb := ifStmt.ElseBranch.(type) {
		case *ast.ElseBlock:
			em.Dedent()
			em.Line("} else {")
			em.Indent()
			for _, stmt := range eb.Body.Statements {
				emitWorkflowStatementCaptured(em, stmt)
			}
			em.CloseBlock()
			return
		case *ast.ElseIf:
			em.Dedent()
			cond := ExprToGo(eb.If.Node.Condition)
			em.Line(fmt.Sprintf("} else if %s {", cond))
			em.Indent()
			for _, stmt := range eb.If.Node.ThenBranch.Statements {
				emitWorkflowStatementCaptured(em, stmt)
			}
			em.CloseBlock()
			return
		}
	}
	em.CloseBlock()
}

func emitStep(em *GoEmitter, step ast.StepStmt) {
	wfName := activeWorkflowName
	stepName := step.Name.Node
	timerVar := fmt.Sprintf("_step%d", stepCounter)
	stepCounter++

	// Check for @retry decorator
	var retryMax int
	var retryBackoff string
	var retryDelay int
	hasRetry := false
	for _, dec := range step.Decorators {
		if dec.Name.Node == "retry" {
			hasRetry = true
			retryMax = 3 // default
			retryBackoff = "exponential"
			retryDelay = 1000
			for _, arg := range dec.Args {
				// Named args are encoded as single-entry MapExpr: {key: value}
				if mapExpr, ok := arg.Node.(ast.MapExpr); ok && len(mapExpr.Entries) == 1 {
					entry := mapExpr.Entries[0]
					if keyIdent, ok := entry.Key.Node.(ast.IdentExpr); ok {
						switch keyIdent.Name {
						case "max":
							if lit, ok := entry.Value.Node.(ast.LiteralExpr); ok {
								if intLit, ok := lit.Lit.(ast.IntLit); ok {
									retryMax = int(intLit.Value)
								}
							}
						case "delay":
							if lit, ok := entry.Value.Node.(ast.LiteralExpr); ok {
								if intLit, ok := lit.Lit.(ast.IntLit); ok {
									retryDelay = int(intLit.Value)
								}
							}
						case "backoff":
							if lit, ok := entry.Value.Node.(ast.LiteralExpr); ok {
								if strLit, ok := lit.Lit.(ast.StringLit); ok {
									retryBackoff = strLit.Value
								}
							}
						}
					}
				}
			}
		}
	}

	// Check for onerror hook
	var onerrorHook *ast.LifecycleHook
	var onsuccessHook *ast.LifecycleHook
	for i := range step.Hooks {
		switch step.Hooks[i].Kind {
		case ast.HookOnerror:
			onerrorHook = &step.Hooks[i]
		case ast.HookOnsuccess:
			onsuccessHook = &step.Hooks[i]
		}
	}

	em.Line(fmt.Sprintf("%s := haira.StepStart(%q, %q)", timerVar, wfName, stepName))

	// If retry, wrap in retry loop
	if hasRetry {
		retryVar := fmt.Sprintf("_retry%d", stepCounter)
		em.Line(fmt.Sprintf("var %s_err error", retryVar))
		em.OpenBlock(fmt.Sprintf("for %s := 0; %s < %d; %s++", retryVar, retryVar, retryMax, retryVar))
		em.Line(fmt.Sprintf("%s_err = nil", retryVar))
		// Emit body in a func to catch panics from ? operator
		em.OpenBlock("func()")
		em.OpenBlock("defer func()")
		em.OpenBlock("if r := recover(); r != nil")
		em.Line(fmt.Sprintf("%s_err = fmt.Errorf(\"%%v\", r)", retryVar))
		em.CloseBlock()
		em.Dedent()
		em.Line("}()")
		emitStepBodyStatements(em, step.Body, wfName, stepName, timerVar)
		em.Dedent()
		em.Line("}()")
		em.OpenBlock(fmt.Sprintf("if %s_err == nil", retryVar))
		em.Line("break")
		em.CloseBlock()
		// Backoff delay
		if retryBackoff == "exponential" {
			em.Line(fmt.Sprintf("haira.StepRetry(%q, %q, %s+1, %d << uint(%s))", wfName, stepName, retryVar, retryDelay, retryVar))
			em.Line(fmt.Sprintf("time.Sleep(time.Duration(%d<<uint(%s)) * time.Millisecond)", retryDelay, retryVar))
		} else {
			em.Line(fmt.Sprintf("haira.StepRetry(%q, %q, %s+1, %d)", wfName, stepName, retryVar, retryDelay))
			em.Line(fmt.Sprintf("time.Sleep(time.Duration(%d) * time.Millisecond)", retryDelay))
		}
		em.CloseBlock() // end retry loop

		// After retries exhausted, check error
		em.OpenBlock(fmt.Sprintf("if %s_err != nil", retryVar))
		if onerrorHook != nil {
			em.Line(fmt.Sprintf("%s := fmt.Sprintf(\"%%v\", %s_err)", onerrorHook.ErrName, retryVar))
			em.Line(fmt.Sprintf("_ = %s", onerrorHook.ErrName))
			em.Line(fmt.Sprintf("haira.StepEnd(%q, %q, %s, %s_err)", wfName, stepName, timerVar, retryVar))
			emitWorkflowBody(em, onerrorHook.Body)
		} else {
			em.Line(fmt.Sprintf("haira.StepEnd(%q, %q, %s, %s_err)", wfName, stepName, timerVar, retryVar))
			em.Line(fmt.Sprintf("return nil, %s_err", retryVar))
		}
		em.CloseBlock()
	} else if onerrorHook != nil {
		// No retry, but has onerror — wrap body in panic recovery
		errVar := fmt.Sprintf("_stepErr%d", stepCounter)
		em.Line(fmt.Sprintf("var %s error", errVar))
		em.OpenBlock("func()")
		em.OpenBlock("defer func()")
		em.OpenBlock("if r := recover(); r != nil")
		em.Line(fmt.Sprintf("%s = fmt.Errorf(\"%%v\", r)", errVar))
		em.CloseBlock()
		em.Dedent()
		em.Line("}()")
		emitStepBodyStatements(em, step.Body, wfName, stepName, timerVar)
		em.Dedent()
		em.Line("}()")
		em.OpenBlock(fmt.Sprintf("if %s != nil", errVar))
		em.Line(fmt.Sprintf("%s := fmt.Sprintf(\"%%v\", %s)", onerrorHook.ErrName, errVar))
		em.Line(fmt.Sprintf("_ = %s", onerrorHook.ErrName))
		em.Line(fmt.Sprintf("haira.StepEnd(%q, %q, %s, %s)", wfName, stepName, timerVar, errVar))
		emitWorkflowBody(em, onerrorHook.Body)
		em.CloseBlock()
	} else {
		// No retry, no onerror — emit body inline
		emitStepBodyStatements(em, step.Body, wfName, stepName, timerVar)
	}

	// Onsuccess hook
	if onsuccessHook != nil {
		em.Line("// onsuccess hook")
		em.OpenBlock("func()")
		em.OpenBlock("defer func()")
		em.Line("recover() // onsuccess errors are non-fatal")
		em.Dedent()
		em.Line("}()")
		EmitBlockBody(em, onsuccessHook.Body)
		em.Dedent()
		em.Line("}()")
	}

	em.Line(fmt.Sprintf("haira.StepEnd(%q, %q, %s, nil)", wfName, stepName, timerVar))
}

func emitStepBodyStatements(em *GoEmitter, stmts []ast.Statement, wfName, stepName, timerVar string) {
	for _, stmt := range stmts {
		switch s := stmt.Node.(type) {
		case ast.ReturnStmt:
			em.Line(fmt.Sprintf("haira.StepEnd(%q, %q, %s, nil)", wfName, stepName, timerVar))
			if inCapturedContext {
				if len(s.Values) == 0 {
					em.Line("return")
				} else {
					vals := make([]string, len(s.Values))
					for i, v := range s.Values {
						vals[i] = ExprToGo(v)
					}
					em.Line(fmt.Sprintf("_wfResult = %s", strings.Join(vals, ", ")))
					em.Line("return")
				}
			} else {
				if len(s.Values) == 0 {
					em.Line("return nil, nil")
				} else {
					vals := make([]string, len(s.Values))
					for i, v := range s.Values {
						vals[i] = ExprToGo(v)
					}
					em.Line(fmt.Sprintf("return %s, nil", strings.Join(vals, ", ")))
				}
			}
		case ast.IfStmt:
			if inCapturedContext {
				emitWorkflowIfCaptured(em, s)
			} else {
				emitWorkflowIf(em, s)
			}
		default:
			EmitStatement(em, stmt)
		}
	}
}

var stepCounter int

// inCapturedContext is true when emitting inside a workflow-level onerror IIFE.
// Returns become _wfResult assignments instead of actual returns.
var inCapturedContext bool

func emitWorkflowIf(em *GoEmitter, ifStmt ast.IfStmt) {
	cond := ExprToGo(ifStmt.Condition)
	em.OpenBlock(fmt.Sprintf("if %s", cond))
	emitWorkflowBody(em, ifStmt.ThenBranch)
	if ifStmt.ElseBranch != nil {
		switch eb := ifStmt.ElseBranch.(type) {
		case *ast.ElseBlock:
			em.Dedent()
			em.Line("} else {")
			em.Indent()
			emitWorkflowBody(em, eb.Body)
			em.CloseBlock()
			return
		case *ast.ElseIf:
			em.Dedent()
			cond := ExprToGo(eb.If.Node.Condition)
			em.Line(fmt.Sprintf("} else if %s {", cond))
			em.Indent()
			emitWorkflowBody(em, eb.If.Node.ThenBranch)
			em.CloseBlock()
			return
		}
	}
	em.CloseBlock()
}

func emitWorkflowMatch(em *GoEmitter, matchExpr ast.MatchExpr) {
	subject := ExprToGo(matchExpr.Subject)
	em.OpenBlock(fmt.Sprintf("switch %s", subject))
	for _, arm := range matchExpr.Arms {
		emitMatchArmHeader(em, arm.Pattern.Node)
		em.Indent()
		switch body := arm.Body.(type) {
		case ast.MatchArmExpr:
			em.Line(ExprToGo(body.Value))
		case ast.MatchArmBlock:
			emitWorkflowBody(em, body.Value)
		}
		em.Dedent()
	}
	em.CloseBlock()
}

func extractTriggerInfo(trigger *ast.Decorator) (string, string) {
	if trigger == nil {
		return "POST", "/"
	}
	method := "POST"
	switch trigger.Name.Node {
	case "webhook", "post":
		method = "POST"
	case "get":
		method = "GET"
	case "put":
		method = "PUT"
	case "delete":
		method = "DELETE"
	}
	path := "/"
	if len(trigger.Args) > 0 {
		if lit, ok := trigger.Args[0].Node.(ast.LiteralExpr); ok {
			if s, ok := lit.Lit.(ast.StringLit); ok {
				path = s.Value
			}
		}
	}
	return method, path
}

// extractDecoratorStringArg extracts a named string argument from a decorator.
func extractDecoratorStringArg(dec ast.Decorator, key string) string {
	for _, arg := range dec.Args {
		if mapExpr, ok := arg.Node.(ast.MapExpr); ok && len(mapExpr.Entries) == 1 {
			entry := mapExpr.Entries[0]
			if keyIdent, ok := entry.Key.Node.(ast.IdentExpr); ok && keyIdent.Name == key {
				if lit, ok := entry.Value.Node.(ast.LiteralExpr); ok {
					if strLit, ok := lit.Lit.(ast.StringLit); ok {
						return strLit.Value
					}
				}
			}
		}
	}
	return ""
}

// extractDecoratorBoolArg extracts a named bool argument from a decorator.
// Returns nil if not found, pointer to bool if found.
func extractDecoratorBoolArg(dec ast.Decorator, key string) *bool {
	for _, arg := range dec.Args {
		if mapExpr, ok := arg.Node.(ast.MapExpr); ok && len(mapExpr.Entries) == 1 {
			entry := mapExpr.Entries[0]
			if keyIdent, ok := entry.Key.Node.(ast.IdentExpr); ok && keyIdent.Name == key {
				if lit, ok := entry.Value.Node.(ast.LiteralExpr); ok {
					if boolLit, ok := lit.Lit.(ast.BoolLit); ok {
						val := boolLit.Value
						return &val
					}
				}
			}
		}
	}
	return nil
}

// emitWorkflowUIMetadata emits @webui and @chatui decorator fields into a WorkflowDef.
func emitWorkflowUIMetadata(em *GoEmitter, decorators []ast.Decorator) {
	for _, dec := range decorators {
		switch dec.Name.Node {
		case "webui":
			if title := extractDecoratorStringArg(dec, "title"); title != "" {
				em.Line(fmt.Sprintf("UITitle: %q,", title))
			}
			if desc := extractDecoratorStringArg(dec, "description"); desc != "" {
				em.Line(fmt.Sprintf("UIDescription: %q,", desc))
			}
		case "chatui":
			if enabled := extractDecoratorBoolArg(dec, "enabled"); enabled != nil {
				if !*enabled {
					em.Line("ChatEnabled: func() *bool { v := false; return &v }(),")
				}
			}
		}
	}
}

// hairaTypeToUIType converts a Haira AST type to a UI type string for WorkflowParam.
func hairaTypeToUIType(ty ast.Type) string {
	if named, ok := ty.(ast.NamedType); ok {
		switch named.Name {
		case "int":
			return "int"
		case "float":
			return "float"
		case "bool":
			return "bool"
		case "file":
			return "file"
		default:
			return "string"
		}
	}
	return "string"
}

// emitWorkflowParams emits the Params slice for a WorkflowDef.
func emitWorkflowParams(em *GoEmitter, params []ast.Param) {
	em.Line("Params: []haira.WorkflowParam{")
	em.Indent()
	for _, param := range params {
		paramType := "string"
		if param.Ty != nil {
			paramType = hairaTypeToUIType(param.Ty.Node)
		}
		em.Line(fmt.Sprintf("{Name: %q, Type: %q},", param.Name.Node, paramType))
	}
	em.Dedent()
	em.Line("},")
}

func buildToolJSONSchema(params []ast.Param) string {
	var props []string
	var required []string
	for _, p := range params {
		jsonType := "string"
		if p.Ty != nil {
			if named, ok := p.Ty.Node.(ast.NamedType); ok {
				switch named.Name {
				case "int", "float":
					jsonType = "number"
				case "bool":
					jsonType = "boolean"
				case "string":
					jsonType = "string"
				default:
					jsonType = "object"
				}
			} else if _, ok := p.Ty.Node.(ast.ListType); ok {
				jsonType = "array"
			}
		}
		props = append(props, fmt.Sprintf(`"%s":{"type":"%s"}`, p.Name.Node, jsonType))
		if p.Default == nil {
			required = append(required, fmt.Sprintf(`"%s"`, p.Name.Node))
		}
	}
	return fmt.Sprintf(`{"type":"object","properties":{%s},"required":[%s]}`,
		strings.Join(props, ","), strings.Join(required, ","))
}

// Naming helpers

func goVarName(prefix, name string) string {
	return prefix + SnakeToPascal(name)
}

func goFuncName(prefix, name string) string {
	return prefix + SnakeToPascal(name)
}

func goFieldName(name string) string {
	return SnakeToPascal(name)
}
