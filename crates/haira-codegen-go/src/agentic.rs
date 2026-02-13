//! Agentic codegen — provider/tool/agent/workflow → Go structs.
//! Implemented in Phase 4.

use haira_ast::*;

use crate::emitter::GoEmitter;
use crate::expressions::expr_to_go;

/// Emit a provider declaration as a Go var.
pub fn emit_provider(em: &mut GoEmitter, provider: &ProviderDecl) {
    let name = to_go_var_name("provider", &provider.name.node);
    em.open_block(&format!("var {} = &haira.Provider", name));
    em.line(&format!("Name: \"{}\",", provider.name.node));
    for (key, value) in &provider.fields {
        let go_key = to_go_field_name(&key.node);
        let go_val = expr_to_go(value);
        em.line(&format!("{}: {},", go_key, go_val));
    }
    em.close_block();
    em.blank();
}

/// Emit an agent declaration as a Go var + init function.
pub fn emit_agent(em: &mut GoEmitter, agent: &AgentDecl) {
    let var_name = to_go_var_name("agent", &agent.name.node);
    em.line(&format!("var {} *haira.Agent", var_name));
    em.blank();

    let init_name = format!("initAgent{}", agent.name.node);
    em.open_block(&format!("func {}()", init_name));

    // Extract fields
    for (key, value) in &agent.fields {
        match key.node.as_str() {
            "tools" => {
                em.line("toolReg := haira.NewToolRegistry()");
                // Expect a list of identifiers
                if let ExprKind::List(items) = &value.node {
                    for item in items {
                        if let ExprKind::Identifier(name) = &item.node {
                            let def_name = to_go_var_name("toolDef", name);
                            em.line(&format!("toolReg.Register({})", def_name));
                        }
                    }
                }
            }
            _ => {}
        }
    }
    em.blank();

    em.open_block(&format!("{} = haira.NewAgent(haira.AgentConfig", var_name));
    em.line(&format!("Name: \"{}\",", agent.name.node));
    for (key, value) in &agent.fields {
        match key.node.as_str() {
            "model" => {
                if let ExprKind::Identifier(name) = &value.node {
                    let provider_var = to_go_var_name("provider", name);
                    em.line(&format!("Provider: {},", provider_var));
                }
            }
            "system" => {
                let go_val = expr_to_go(value);
                em.line(&format!("System: {},", go_val));
            }
            "tools" => {
                em.line("Tools: toolReg,");
            }
            "temperature" => {
                let go_val = expr_to_go(value);
                em.line(&format!("Temperature: {},", go_val));
            }
            "memory" => {
                // memory: conversation(max_turns: 20)
                if let ExprKind::Call(call) = &value.node {
                    if let ExprKind::Identifier(kind) = &call.callee.node {
                        let mut config = format!("haira.MemoryConfig{{Kind: \"{}\"", kind);
                        for arg in &call.args {
                            if let Some(name) = &arg.name {
                                if name.node.as_str() == "max_turns" {
                                    let val = expr_to_go(&arg.value);
                                    config.push_str(&format!(", MaxTurns: {}", val));
                                }
                            }
                        }
                        config.push('}');
                        em.line(&format!("Memory: {},", config));
                    }
                }
            }
            _ => {
                let go_key = to_go_field_name(&key.node);
                let go_val = expr_to_go(value);
                em.line(&format!("{}: {},", go_key, go_val));
            }
        }
    }
    em.dedent();
    em.line("})");

    em.close_block();
    em.blank();
}

/// Emit a tool declaration as a Go handler function + ToolDef var.
pub fn emit_tool(em: &mut GoEmitter, tool: &ToolDecl) {
    let handler_name = to_go_func_name("tool", &tool.name.node);
    let def_name = to_go_var_name("toolDef", &tool.name.node);

    // Handler function
    em.open_block(&format!(
        "func {}(args json.RawMessage) (any, error)",
        handler_name
    ));
    // Parse params struct
    em.line("var params struct {");
    em.indent();
    for param in &tool.params {
        let go_name = snake_to_pascal(&param.name.node);
        let go_type = param
            .ty
            .as_ref()
            .map(|t| crate::types::haira_type_to_go(&t.node))
            .unwrap_or_else(|| "any".to_string());
        let json_tag = param.name.node.as_str();
        em.line(&format!("{} {} `json:\"{}\"`", go_name, go_type, json_tag));
    }
    em.dedent();
    em.line("}");
    em.line("json.Unmarshal(args, &params)");

    // Apply defaults
    for param in &tool.params {
        if let Some(default) = &param.default {
            let go_name = snake_to_pascal(&param.name.node);
            let default_val = expr_to_go(default);
            let zero_check = match param.ty.as_ref().map(|t| &t.node) {
                Some(haira_ast::Type::Named(n)) if n.as_str() == "int" => {
                    format!("params.{} == 0", go_name)
                }
                Some(haira_ast::Type::Named(n)) if n.as_str() == "string" => {
                    format!("params.{} == \"\"", go_name)
                }
                _ => format!("params.{} == 0", go_name),
            };
            em.open_block(&format!("if {}", zero_check));
            em.line(&format!("params.{} = {}", go_name, default_val));
            em.close_block();
        }
    }

    // Emit local aliases for params so tool body can use bare names
    for param in &tool.params {
        let go_name = snake_to_pascal(&param.name.node);
        em.line(&format!("{} := params.{}", param.name.node, go_name));
    }

    // Body or stub
    if let Some(body) = &tool.body {
        crate::statements::emit_tool_body(em, body);
    } else {
        em.line(&format!(
            "return nil, fmt.Errorf(\"tool {} not yet implemented\")",
            tool.name.node
        ));
    }
    em.close_block();
    em.blank();

    // ToolDef var
    // Build JSON schema for parameters
    let schema = build_tool_json_schema(&tool.params);
    em.open_block(&format!("var {} = &haira.ToolDef", def_name));
    em.line(&format!("Name:        \"{}\",", tool.name.node));
    if tool.description.contains('\n') {
        em.line(&format!("Description: `{}`,", tool.description));
    } else {
        let description = tool.description.replace('"', "\\\"");
        em.line(&format!("Description: \"{}\",", description));
    }
    em.line(&format!("Parameters:  json.RawMessage(`{}`),", schema));
    em.line(&format!("Handler:     {},", handler_name));
    em.close_block();
    em.blank();
}

/// Emit a workflow declaration.
pub fn emit_workflow(em: &mut GoEmitter, workflow: &WorkflowDecl) {
    let handler_name = to_go_func_name("workflow", &workflow.name.node);
    let def_name = to_go_var_name("workflowDef", &workflow.name.node);

    // Handler function
    em.open_block(&format!(
        "func {}(params map[string]any) (any, error)",
        handler_name
    ));
    // Extract params
    for param in &workflow.params {
        let go_type = param
            .ty
            .as_ref()
            .map(|t| crate::types::haira_type_to_go(&t.node))
            .unwrap_or_else(|| "string".to_string());
        em.line(&format!(
            "{}, _ := params[\"{}\"].({})",
            param.name.node, param.name.node, go_type
        ));
    }

    emit_workflow_body(em, &workflow.body);
    em.close_block();
    em.blank();

    // WorkflowDef var
    let (method, path) = extract_trigger_info(&workflow.trigger);
    em.open_block(&format!("var {} = &haira.WorkflowDef", def_name));
    em.line(&format!("Name: \"{}\",", workflow.name.node));
    em.line(&format!("Method: \"{}\",", method));
    em.line(&format!("Path: \"{}\",", path));
    em.line(&format!("Handler: {},", handler_name));
    em.close_block();
    em.blank();
}

/// Emit a workflow body, wrapping `return X` → `return X, nil`
/// since workflow handlers return `(any, error)`.
fn emit_workflow_body(em: &mut GoEmitter, block: &Block) {
    for stmt in &block.statements {
        match &stmt.node {
            StatementKind::Return(ret) => {
                if ret.values.is_empty() {
                    em.line("return nil, nil");
                } else {
                    let vals = ret
                        .values
                        .iter()
                        .map(|v| expr_to_go(v))
                        .collect::<Vec<_>>()
                        .join(", ");
                    em.line(&format!("return {}, nil", vals));
                }
            }
            StatementKind::If(if_stmt) => {
                // Need to recurse into if/else to wrap returns there too
                emit_workflow_if(em, if_stmt);
            }
            _ => {
                crate::statements::emit_statement(em, stmt);
            }
        }
    }
}

/// Emit an if statement in workflow context, wrapping returns in branches.
fn emit_workflow_if(em: &mut GoEmitter, if_stmt: &IfStatement) {
    let cond = expr_to_go(&if_stmt.condition);
    em.open_block(&format!("if {}", cond));
    emit_workflow_body(em, &if_stmt.then_branch);
    match &if_stmt.else_branch {
        Some(ElseBranch::Block(block)) => {
            em.dedent();
            em.line("} else {");
            em.indent();
            emit_workflow_body(em, block);
            em.close_block();
        }
        Some(ElseBranch::ElseIf(elif)) => {
            em.dedent();
            let cond = expr_to_go(&elif.node.condition);
            em.line(&format!("}} else if {} {{", cond));
            em.indent();
            emit_workflow_body(em, &elif.node.then_branch);
            em.close_block();
        }
        None => {
            em.close_block();
        }
    }
}

fn extract_trigger_info(trigger: &Option<Decorator>) -> (String, String) {
    if let Some(dec) = trigger {
        let method = match dec.name.node.as_str() {
            "webhook" => "POST",
            "get" => "GET",
            _ => "POST",
        };
        let path = dec
            .args
            .first()
            .map(|a| {
                if let ExprKind::Literal(Literal::String(s)) = &a.node {
                    s.to_string()
                } else {
                    "/".to_string()
                }
            })
            .unwrap_or_else(|| "/".to_string());
        (method.to_string(), path)
    } else {
        ("POST".to_string(), "/".to_string())
    }
}

fn build_tool_json_schema(params: &[Param]) -> String {
    let mut props = Vec::new();
    let mut required = Vec::new();

    for param in params {
        let json_type = match param.ty.as_ref().map(|t| &t.node) {
            Some(haira_ast::Type::Named(n)) => match n.as_str() {
                "string" => "string",
                "int" | "float" => "number",
                "bool" => "boolean",
                _ => "object",
            },
            Some(haira_ast::Type::List(_)) => "array",
            _ => "string",
        };
        props.push(format!(
            "\"{}\":{{\"type\":\"{}\"}}",
            param.name.node, json_type
        ));
        if param.default.is_none() {
            required.push(format!("\"{}\"", param.name.node));
        }
    }

    format!(
        "{{\"type\":\"object\",\"properties\":{{{}}},\"required\":[{}]}}",
        props.join(","),
        required.join(",")
    )
}

// Helper functions for Go naming conventions

fn to_go_var_name(prefix: &str, name: &str) -> String {
    format!("{}{}", prefix, snake_to_pascal(name))
}

fn to_go_func_name(prefix: &str, name: &str) -> String {
    format!("{}{}", prefix, snake_to_pascal(name))
}

/// Convert snake_case or plain name to PascalCase.
/// "search_recipes" → "SearchRecipes", "Chat" → "Chat"
fn snake_to_pascal(name: &str) -> String {
    name.split('_')
        .map(|part| capitalize(part))
        .collect::<String>()
}

fn to_go_field_name(name: &str) -> String {
    // Convert snake_case to PascalCase
    name.split('_')
        .map(|part| capitalize(part))
        .collect::<String>()
}

fn capitalize(s: &str) -> String {
    let mut chars = s.chars();
    match chars.next() {
        Some(c) => c.to_uppercase().to_string() + chars.as_str(),
        None => String::new(),
    }
}
