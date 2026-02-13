//! Statement codegen — StatementKind → Go statements.

use haira_ast::*;

use crate::emitter::GoEmitter;
use crate::expressions::expr_to_go;

/// Emit a Haira statement as Go code.
pub fn emit_statement(em: &mut GoEmitter, stmt: &Statement) {
    match &stmt.node {
        StatementKind::Assignment(assign) => emit_assignment(em, assign),
        StatementKind::If(if_stmt) => emit_if(em, if_stmt),
        StatementKind::For(for_stmt) => emit_for(em, for_stmt),
        StatementKind::While(while_stmt) => emit_while(em, while_stmt),
        StatementKind::Return(ret) => emit_return(em, ret),
        StatementKind::Expr(expr) => {
            em.line(&expr_to_go(expr));
        }
        StatementKind::Break => em.line("break"),
        StatementKind::Continue => em.line("continue"),
        StatementKind::Match(match_expr) => emit_match(em, match_expr),
        StatementKind::Try(try_stmt) => emit_try(em, try_stmt),
    }
}

fn emit_assignment(em: &mut GoEmitter, assign: &Assignment) {
    // Special case: match expression as RHS → emit switch that assigns to variable
    if assign.targets.len() == 1 {
        if let ExprKind::Match(match_expr) = &assign.value.node {
            let target = assign_path_to_go(&assign.targets[0].path);
            if let AssignPath::Identifier(name) = &assign.targets[0].path {
                em.declare_var(&name.node);
            }
            em.line(&format!("var {} any", target));
            emit_match_assignment(em, &target, match_expr);
            return;
        }
    }

    let value = expr_to_go(&assign.value);

    if assign.targets.len() == 1 {
        let target = &assign.targets[0];
        let path = assign_path_to_go(&target.path);
        // Use := for first assignment of simple identifiers, = for reassignment or field/index
        let op = match &target.path {
            AssignPath::Identifier(name) => {
                if em.declare_var(&name.node) {
                    ":="
                } else {
                    "="
                }
            }
            _ => "=",
        };
        em.line(&format!("{} {} {}", path, op, value));
    } else {
        // Multi-target: check if any target is new
        let targets: Vec<String> = assign
            .targets
            .iter()
            .map(|t| assign_path_to_go(&t.path))
            .collect();
        let any_new = assign.targets.iter().any(|t| {
            if let AssignPath::Identifier(name) = &t.path {
                em.declare_var(&name.node)
            } else {
                false
            }
        });
        let op = if any_new { ":=" } else { "=" };
        em.line(&format!("{} {} {}", targets.join(", "), op, value));
    }
}

fn emit_match_assignment(em: &mut GoEmitter, target: &str, match_expr: &MatchExpr) {
    let subject = expr_to_go(&match_expr.subject);
    em.open_block(&format!("switch {}", subject));
    for arm in &match_expr.arms {
        emit_match_arm_header(em, &arm.pattern.node);
        em.indent();
        match &arm.body {
            MatchArmBody::Expr(expr) => em.line(&format!("{} = {}", target, expr_to_go(expr))),
            MatchArmBody::Block(block) => emit_block_body(em, block),
        }
        em.dedent();
    }
    em.close_block();
}

fn assign_path_to_go(path: &AssignPath) -> String {
    match path {
        AssignPath::Identifier(name) => name.node.to_string(),
        AssignPath::Field { object, field } => {
            format!("{}.{}", assign_path_to_go(object), field.node)
        }
        AssignPath::Index { object, index } => {
            format!("{}[{}]", assign_path_to_go(object), expr_to_go(index))
        }
    }
}

fn emit_if(em: &mut GoEmitter, if_stmt: &IfStatement) {
    let cond = expr_to_go(&if_stmt.condition);
    em.open_block(&format!("if {}", cond));
    emit_block_body(em, &if_stmt.then_branch);
    match &if_stmt.else_branch {
        Some(ElseBranch::Block(block)) => {
            em.dedent();
            em.line("} else {");
            em.indent();
            emit_block_body(em, block);
            em.close_block();
        }
        Some(ElseBranch::ElseIf(elif)) => {
            em.dedent();
            let cond = expr_to_go(&elif.node.condition);
            em.line(&format!("}} else if {} {{", cond));
            em.indent();
            emit_block_body(em, &elif.node.then_branch);
            // Recurse for further else-if/else chains
            if let Some(else_branch) = &elif.node.else_branch {
                match else_branch {
                    ElseBranch::Block(block) => {
                        em.dedent();
                        em.line("} else {");
                        em.indent();
                        emit_block_body(em, block);
                    }
                    ElseBranch::ElseIf(_) => {
                        // Close current and let recursion handle it
                        // For simplicity, flatten in Phase 2
                    }
                }
            }
            em.close_block();
        }
        None => {
            em.close_block();
        }
    }
}

fn emit_for(em: &mut GoEmitter, for_stmt: &ForStatement) {
    let iter = expr_to_go(&for_stmt.iterator);
    match &for_stmt.pattern {
        ForPattern::Single(var) => {
            em.open_block(&format!("for _, {} := range {}", var.node, iter));
        }
        ForPattern::Pair(idx, val) => {
            em.open_block(&format!("for {}, {} := range {}", idx.node, val.node, iter));
        }
    }
    emit_block_body(em, &for_stmt.body);
    em.close_block();
}

fn emit_while(em: &mut GoEmitter, while_stmt: &WhileStatement) {
    let cond = expr_to_go(&while_stmt.condition);
    em.open_block(&format!("for {}", cond));
    emit_block_body(em, &while_stmt.body);
    em.close_block();
}

fn emit_return(em: &mut GoEmitter, ret: &ReturnStatement) {
    if ret.values.is_empty() {
        em.line("return");
    } else {
        let vals = ret
            .values
            .iter()
            .map(|v| expr_to_go(v))
            .collect::<Vec<_>>()
            .join(", ");
        em.line(&format!("return {}", vals));
    }
}

fn emit_match(em: &mut GoEmitter, match_expr: &MatchExpr) {
    let subject = expr_to_go(&match_expr.subject);
    em.open_block(&format!("switch {}", subject));
    for arm in &match_expr.arms {
        emit_match_arm_header(em, &arm.pattern.node);
        em.indent();
        match &arm.body {
            MatchArmBody::Expr(expr) => em.line(&expr_to_go(expr)),
            MatchArmBody::Block(block) => emit_block_body(em, block),
        }
        em.dedent();
    }
    em.close_block();
}

fn emit_match_arm_header(em: &mut GoEmitter, pattern: &Pattern) {
    if matches!(pattern, Pattern::Wildcard) {
        em.line("default:");
    } else {
        em.line(&format!("case {}:", pattern_to_go(pattern)));
    }
}

fn pattern_to_go(pattern: &Pattern) -> String {
    match pattern {
        Pattern::Wildcard => "default".to_string(),
        Pattern::Literal(lit) => match lit {
            Literal::Int(n) => n.to_string(),
            Literal::Float(f) => f.to_string(),
            Literal::String(s) => format!("\"{}\"", s),
            Literal::Bool(b) => b.to_string(),
            Literal::InterpolatedString(_) => "/* interp */".to_string(),
        },
        Pattern::Identifier(name) => name.to_string(),
        Pattern::Constructor { name, .. } => name.to_string(),
    }
}

fn emit_try(em: &mut GoEmitter, try_stmt: &TryStatement) {
    // Go doesn't have try/catch — emit as a comment + inline for now
    em.line("// try {");
    emit_block_body(em, &try_stmt.body);
    em.line(&format!("// }} catch {} {{", try_stmt.error_name.node));
    emit_block_body(em, &try_stmt.catch_body);
    em.line("// }");
}

/// Emit the body of a block (just the statements, no braces).
pub fn emit_block_body(em: &mut GoEmitter, block: &Block) {
    for stmt in &block.statements {
        emit_statement(em, stmt);
    }
}

/// Emit a tool body — like emit_block_body but wraps `return X` → `return X, nil`
/// since tool handlers return `(any, error)`.
pub fn emit_tool_body(em: &mut GoEmitter, block: &Block) {
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
                emit_tool_if(em, if_stmt);
            }
            _ => {
                emit_statement(em, stmt);
            }
        }
    }
}

/// Emit an if statement in tool context, wrapping returns in branches.
fn emit_tool_if(em: &mut GoEmitter, if_stmt: &IfStatement) {
    let cond = expr_to_go(&if_stmt.condition);
    em.open_block(&format!("if {}", cond));
    emit_tool_body(em, &if_stmt.then_branch);
    match &if_stmt.else_branch {
        Some(ElseBranch::Block(block)) => {
            em.dedent();
            em.line("} else {");
            em.indent();
            emit_tool_body(em, block);
            em.close_block();
        }
        Some(ElseBranch::ElseIf(elif)) => {
            em.dedent();
            let cond = expr_to_go(&elif.node.condition);
            em.line(&format!("}} else if {} {{", cond));
            em.indent();
            emit_tool_body(em, &elif.node.then_branch);
            em.close_block();
        }
        None => {
            em.close_block();
        }
    }
}
