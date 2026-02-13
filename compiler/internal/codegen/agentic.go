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
	handlerName := goFuncName("workflow", workflow.Name.Node)
	defName := goVarName("workflowDef", workflow.Name.Node)

	// Handler function
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
	emitWorkflowBody(em, workflow.Body)
	em.CloseBlock()
	em.Blank()

	// WorkflowDef var
	method, path := extractTriggerInfo(workflow.Trigger)
	em.OpenBlock(fmt.Sprintf("var %s = &haira.WorkflowDef", defName))
	em.Line(fmt.Sprintf("Name: %q,", workflow.Name.Node))
	em.Line(fmt.Sprintf("Method: %q,", method))
	em.Line(fmt.Sprintf("Path: %q,", path))
	em.Line(fmt.Sprintf("Handler: %s,", handlerName))
	em.CloseBlock()
	em.Blank()
}

func emitWorkflowBody(em *GoEmitter, block ast.Block) {
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
			emitWorkflowIf(em, s)
		case ast.MatchStmt:
			emitWorkflowMatch(em, s.Match)
		default:
			EmitStatement(em, stmt)
		}
	}
}

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
	case "webhook":
		method = "POST"
	case "get":
		method = "GET"
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
