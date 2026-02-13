//! GoCodegen — walks the AST and orchestrates Go code emission.

use haira_ast::*;

use crate::agentic;
use crate::emitter::GoEmitter;
use crate::goproject::GoProject;
use crate::statements::emit_block_body;
use crate::types::haira_type_to_go;

use std::path::PathBuf;

/// Generate a GoProject from a parsed Haira AST.
pub fn generate_go_project(ast: &SourceFile, runtime_path: PathBuf, out_dir: PathBuf) -> GoProject {
    let main_go = generate_main_go(ast);
    GoProject {
        dir: out_dir,
        main_go,
        runtime_path,
    }
}

/// Generate the contents of main.go from the AST.
fn generate_main_go(ast: &SourceFile) -> String {
    let mut em = GoEmitter::new();

    em.line("package main");
    em.blank();

    // Collect what we need to import
    let needs_fmt = needs_fmt_import(ast);
    let needs_json = needs_json_import(ast);
    let needs_haira = needs_haira_import(ast);

    // Emit imports
    let mut imports = Vec::new();
    if needs_fmt {
        imports.push("\"fmt\"");
    }
    if needs_json {
        imports.push("\"encoding/json\"");
    }
    if needs_haira {
        imports.push("\"haira-go-runtime/haira\"");
    }

    if !imports.is_empty() {
        if imports.len() == 1 {
            em.line(&format!("import {}", imports[0]));
        } else {
            em.line("import (");
            em.indent();
            for imp in &imports {
                em.line(imp);
            }
            em.dedent();
            em.line(")");
        }
        em.blank();
    }

    // Emit top-level items
    // Order: providers → tools → agents → workflows → type defs → functions → main
    let mut main_fn: Option<&FunctionDef> = None;
    let mut agent_names: Vec<String> = Vec::new();

    // First pass: providers
    for item in &ast.items {
        if let ItemKind::ProviderDecl(provider) = &item.node {
            agentic::emit_provider(&mut em, provider);
        }
    }

    // Second pass: tools
    for item in &ast.items {
        if let ItemKind::ToolDecl(tool) = &item.node {
            agentic::emit_tool(&mut em, tool);
        }
    }

    // Third pass: agents
    for item in &ast.items {
        if let ItemKind::AgentDecl(agent) = &item.node {
            agent_names.push(agent.name.node.to_string());
            agentic::emit_agent(&mut em, agent);
        }
    }

    // Fourth pass: workflows
    for item in &ast.items {
        if let ItemKind::WorkflowDecl(workflow) = &item.node {
            agentic::emit_workflow(&mut em, workflow);
        }
    }

    // Fifth pass: type defs
    for item in &ast.items {
        if let ItemKind::TypeDef(type_def) = &item.node {
            emit_type_def(&mut em, type_def);
        }
    }

    // Sixth pass: non-main functions
    for item in &ast.items {
        if let ItemKind::FunctionDef(func) = &item.node {
            if func.name.node.as_str() == "main" {
                main_fn = Some(func);
            } else {
                emit_function(&mut em, func);
            }
        }
    }

    // Finally: main function
    if let Some(func) = main_fn {
        emit_main_function(&mut em, func, &agent_names);
    }

    // Suppress unused import warnings
    if needs_fmt {
        // fmt is only imported when actually used (interpolated strings)
    }

    em.finish()
}

fn emit_type_def(em: &mut GoEmitter, type_def: &TypeDef) {
    em.open_block(&format!("type {} struct", type_def.name.node));
    for field in &type_def.fields {
        let go_type = field
            .ty
            .as_ref()
            .map(|t| haira_type_to_go(&t.node))
            .unwrap_or_else(|| "any".to_string());
        let field_name = capitalize(&field.name.node);
        em.line(&format!("{} {}", field_name, go_type));
    }
    em.close_block();
    em.blank();
}

fn emit_function(em: &mut GoEmitter, func: &FunctionDef) {
    let name = capitalize(&func.name.node);
    let params = func
        .params
        .iter()
        .map(|p| {
            let ty =
                p.ty.as_ref()
                    .map(|t| haira_type_to_go(&t.node))
                    .unwrap_or_else(|| "any".to_string());
            format!("{} {}", p.name.node, ty)
        })
        .collect::<Vec<_>>()
        .join(", ");

    let ret = func
        .return_ty
        .as_ref()
        .map(|t| format!(" {}", haira_type_to_go(&t.node)))
        .unwrap_or_default();

    em.open_block(&format!("func {}({}){}", name, params, ret));
    emit_block_body(em, &func.body);
    em.close_block();
    em.blank();
}

fn emit_main_function(em: &mut GoEmitter, func: &FunctionDef, agent_names: &[String]) {
    em.open_block("func main()");

    // Call agent init functions
    for agent_name in agent_names {
        em.line(&format!("initAgent{}()", agent_name));
    }
    if !agent_names.is_empty() {
        em.blank();
    }

    emit_block_body(em, &func.body);
    em.close_block();
}

/// Check if we need to import "fmt" (for interpolated strings or stub tools).
fn needs_fmt_import(ast: &SourceFile) -> bool {
    ast.items.iter().any(|item| match &item.node {
        // Stub tools (no body) need fmt for fmt.Errorf; tools with body need it for interpolation
        ItemKind::ToolDecl(tool) => {
            tool.body.is_none()
                || tool
                    .body
                    .as_ref()
                    .map_or(false, |b| block_has_interpolated_string(b))
        }
        ItemKind::FunctionDef(func) => block_has_interpolated_string(&func.body),
        _ => false,
    })
}

fn block_has_interpolated_string(block: &Block) -> bool {
    block
        .statements
        .iter()
        .any(|stmt| stmt_has_interpolated_string(&stmt.node))
}

fn stmt_has_interpolated_string(stmt: &StatementKind) -> bool {
    match stmt {
        StatementKind::Return(ret) => ret.values.iter().any(|e| expr_has_interpolated_string(e)),
        StatementKind::Expr(e) | StatementKind::Assignment(Assignment { value: e, .. }) => {
            expr_has_interpolated_string(e)
        }
        StatementKind::If(if_stmt) => {
            block_has_interpolated_string(&if_stmt.then_branch)
                || match &if_stmt.else_branch {
                    Some(ElseBranch::Block(b)) => block_has_interpolated_string(b),
                    Some(ElseBranch::ElseIf(ei)) => {
                        stmt_has_interpolated_string(&StatementKind::If(ei.node.clone()))
                    }
                    None => false,
                }
        }
        StatementKind::For(for_stmt) => block_has_interpolated_string(&for_stmt.body),
        StatementKind::While(while_stmt) => block_has_interpolated_string(&while_stmt.body),
        StatementKind::Match(match_expr) => match_expr.arms.iter().any(|arm| match &arm.body {
            MatchArmBody::Block(b) => block_has_interpolated_string(b),
            MatchArmBody::Expr(e) => expr_has_interpolated_string(e),
        }),
        _ => false,
    }
}

fn expr_has_interpolated_string(expr: &Expr) -> bool {
    match &expr.node {
        ExprKind::Literal(Literal::InterpolatedString(_)) => true,
        ExprKind::Call(call) => {
            call.args
                .iter()
                .any(|a| expr_has_interpolated_string(&a.value))
                || expr_has_interpolated_string(&call.callee)
        }
        ExprKind::Binary(bin) => {
            expr_has_interpolated_string(&bin.left) || expr_has_interpolated_string(&bin.right)
        }
        ExprKind::Unary(un) => expr_has_interpolated_string(&un.operand),
        ExprKind::Paren(inner) => expr_has_interpolated_string(inner),
        ExprKind::MethodCall(mc) => {
            mc.args
                .iter()
                .any(|a| expr_has_interpolated_string(&a.value))
                || expr_has_interpolated_string(&mc.receiver)
        }
        ExprKind::Pipe(pipe) => {
            expr_has_interpolated_string(&pipe.left) || expr_has_interpolated_string(&pipe.right)
        }
        _ => false,
    }
}

/// Check if we need "encoding/json" (for tools/agents).
fn needs_json_import(ast: &SourceFile) -> bool {
    ast.items
        .iter()
        .any(|item| matches!(&item.node, ItemKind::ToolDecl(_) | ItemKind::AgentDecl(_)))
}

/// Check if we need the haira runtime import.
fn needs_haira_import(ast: &SourceFile) -> bool {
    ast.items.iter().any(|item| {
        matches!(
            &item.node,
            ItemKind::ImportDecl(_)
                | ItemKind::ProviderDecl(_)
                | ItemKind::ToolDecl(_)
                | ItemKind::AgentDecl(_)
                | ItemKind::WorkflowDecl(_)
        )
    })
}

fn capitalize(s: &str) -> String {
    let mut chars = s.chars();
    match chars.next() {
        Some(c) => c.to_uppercase().to_string() + chars.as_str(),
        None => String::new(),
    }
}
