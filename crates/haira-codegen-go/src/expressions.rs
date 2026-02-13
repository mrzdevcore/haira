//! Expression codegen — ExprKind → Go expression strings.

use haira_ast::*;

use crate::stdlib::{resolve_stdlib_call, resolve_stdlib_method_call};

/// Convert a Haira expression to a Go expression string.
pub fn expr_to_go(expr: &Expr) -> String {
    match &expr.node {
        ExprKind::Literal(lit) => literal_to_go(lit),
        ExprKind::Identifier(name) => name.to_string(),
        ExprKind::Binary(bin) => {
            let left = expr_to_go(&bin.left);
            let right = expr_to_go(&bin.right);
            let op = binop_to_go(&bin.op.node);
            format!("{} {} {}", left, op, right)
        }
        ExprKind::Unary(un) => {
            let operand = expr_to_go(&un.operand);
            match un.op.node {
                UnaryOp::Neg => format!("-{}", operand),
                UnaryOp::Not => format!("!{}", operand),
            }
        }
        ExprKind::Call(call) => {
            // Check if this is a stdlib call (e.g. io.println)
            if let Some(resolved) = resolve_stdlib_call(call) {
                return resolved;
            }
            // Capitalize function name for Go export convention
            let callee = match &call.callee.node {
                ExprKind::Identifier(name) => capitalize(name),
                _ => expr_to_go(&call.callee),
            };
            let args = call
                .args
                .iter()
                .map(|a| expr_to_go(&a.value))
                .collect::<Vec<_>>()
                .join(", ");
            format!("{}({})", callee, args)
        }
        ExprKind::MethodCall(mc) => {
            // Check if this is a stdlib method call (e.g. io.println)
            if let Some(resolved) = resolve_stdlib_method_call(mc) {
                return resolved;
            }
            // Check if this is an agent method call (PascalCase receiver + ask/run/stream)
            if let ExprKind::Identifier(name) = &mc.receiver.node {
                if name.chars().next().map_or(false, |c| c.is_uppercase())
                    && matches!(mc.method.node.as_str(), "ask" | "run" | "stream")
                {
                    let agent_var = format!("agent{}", name);
                    let method = capitalize(&mc.method.node);
                    let mut positional = Vec::new();
                    let mut session_arg: Option<String> = None;
                    for arg in &mc.args {
                        if let Some(n) = &arg.name {
                            if n.node.as_str() == "session" {
                                session_arg = Some(expr_to_go(&arg.value));
                            } else {
                                positional.push(expr_to_go(&arg.value));
                            }
                        } else {
                            positional.push(expr_to_go(&arg.value));
                        }
                    }
                    // Agent.Ask always requires a session ID; default to ""
                    if mc.method.node.as_str() == "ask" {
                        positional.push(session_arg.unwrap_or_else(|| "\"\"".to_string()));
                    } else if let Some(s) = session_arg {
                        positional.push(s);
                    }
                    return format!("{}.{}({})", agent_var, method, positional.join(", "));
                }
            }
            let receiver = expr_to_go(&mc.receiver);
            let method = capitalize(&mc.method.node);
            let args = mc
                .args
                .iter()
                .map(|a| expr_to_go(&a.value))
                .collect::<Vec<_>>()
                .join(", ");
            format!("{}.{}({})", receiver, method, args)
        }
        ExprKind::Field(f) => {
            let obj = expr_to_go(&f.object);
            format!("{}.{}", obj, f.field.node)
        }
        ExprKind::Index(idx) => {
            let obj = expr_to_go(&idx.object);
            let index = expr_to_go(&idx.index);
            // Use haira.Get() for safe dynamic indexing on any values
            format!("haira.Get({}, {})", obj, index)
        }
        ExprKind::Pipe(pipe) => {
            // x | f | g  →  G(F(x))
            let left = expr_to_go(&pipe.left);
            let right = match &pipe.right.node {
                ExprKind::Identifier(name) => capitalize(name),
                _ => expr_to_go(&pipe.right),
            };
            format!("{}({})", right, left)
        }
        ExprKind::List(items) => {
            let elems = items
                .iter()
                .map(|e| expr_to_go(e))
                .collect::<Vec<_>>()
                .join(", ");
            format!("[]any{{{}}}", elems)
        }
        ExprKind::Map(entries) => {
            let pairs = entries
                .iter()
                .map(|(k, v)| {
                    let key_str = match &k.node {
                        ExprKind::Identifier(name) => format!("\"{}\"", name),
                        _ => expr_to_go(k),
                    };
                    format!("{}: {}", key_str, expr_to_go(v))
                })
                .collect::<Vec<_>>()
                .join(", ");
            format!("map[string]any{{{}}}", pairs)
        }
        ExprKind::Paren(inner) => {
            format!("({})", expr_to_go(inner))
        }
        ExprKind::None => "nil".to_string(),
        ExprKind::Some(inner) => expr_to_go(inner),
        // Stubs for constructs handled in later phases
        ExprKind::Lambda(_) => "nil /* lambda */".to_string(),
        ExprKind::Match(_) => "nil /* match */".to_string(),
        ExprKind::If(_) => "nil /* if-expr */".to_string(),
        ExprKind::Block(_) => "nil /* block-expr */".to_string(),
        ExprKind::Instance(_) => "nil /* instance */".to_string(),
        ExprKind::Range(_) => "nil /* range */".to_string(),
        ExprKind::Propagate(inner) => expr_to_go(inner),
        ExprKind::Async(_) => "nil /* async */".to_string(),
        ExprKind::Spawn(_) => "nil /* spawn */".to_string(),
        ExprKind::Select(_) => "nil /* select */".to_string(),
    }
}

fn literal_to_go(lit: &Literal) -> String {
    match lit {
        Literal::Int(n) => n.to_string(),
        Literal::Float(f) => {
            let s = f.to_string();
            // Ensure it has a decimal point for Go
            if s.contains('.') {
                s
            } else {
                format!("{}.0", s)
            }
        }
        Literal::String(s) => {
            if s.contains('\n') {
                // Multi-line strings use Go backtick syntax
                format!("`{}`", s)
            } else {
                format!("\"{}\"", s.replace('\\', "\\\\").replace('"', "\\\""))
            }
        }
        Literal::Bool(b) => b.to_string(),
        Literal::InterpolatedString(parts) => {
            // Convert to fmt.Sprintf
            let mut format_str = String::new();
            let mut args = Vec::new();
            for part in parts {
                match part {
                    StringPart::Literal(s) => format_str.push_str(s),
                    StringPart::Expr(e) => {
                        format_str.push_str("%v");
                        args.push(expr_to_go(e));
                    }
                }
            }
            if args.is_empty() {
                format!("\"{}\"", format_str)
            } else {
                format!("fmt.Sprintf(\"{}\", {})", format_str, args.join(", "))
            }
        }
    }
}

fn capitalize(s: &str) -> String {
    let mut chars = s.chars();
    match chars.next() {
        Some(c) => c.to_uppercase().to_string() + chars.as_str(),
        None => String::new(),
    }
}

fn binop_to_go(op: &BinaryOp) -> &'static str {
    match op {
        BinaryOp::Add => "+",
        BinaryOp::Sub => "-",
        BinaryOp::Mul => "*",
        BinaryOp::Div => "/",
        BinaryOp::Mod => "%",
        BinaryOp::Eq => "==",
        BinaryOp::Ne => "!=",
        BinaryOp::Lt => "<",
        BinaryOp::Gt => ">",
        BinaryOp::Le => "<=",
        BinaryOp::Ge => ">=",
        BinaryOp::And => "&&",
        BinaryOp::Or => "||",
    }
}
