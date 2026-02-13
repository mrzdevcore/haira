//! Stdlib mapping — Haira imports → Go runtime calls.

use haira_ast::{CallExpr, ExprKind, MethodCallExpr};

use crate::expressions::expr_to_go;

/// Try to resolve a call as a stdlib call.
/// Returns Some(go_code) if it's a known stdlib function, None otherwise.
pub fn resolve_stdlib_call(call: &CallExpr) -> Option<String> {
    // Check for qualified calls like io.println, http.get, postgres.connect, etc.
    if let ExprKind::Field(field) = &call.callee.node {
        if let ExprKind::Identifier(module) = &field.object.node {
            let method = field.field.node.as_str();
            let args = call
                .args
                .iter()
                .map(|a| expr_to_go(&a.value))
                .collect::<Vec<_>>()
                .join(", ");

            return resolve_qualified(module.as_str(), method, &args, call);
        }
    }

    // Check for bare calls like env(...)
    if let ExprKind::Identifier(name) = &call.callee.node {
        let args = call
            .args
            .iter()
            .map(|a| expr_to_go(&a.value))
            .collect::<Vec<_>>()
            .join(", ");

        return match name.as_str() {
            "env" => Some(format!("haira.Env({})", args)),
            "len" => Some(format!("haira.Len({})", args)),
            "keys" => Some(format!("haira.Keys({})", args)),
            "join" => Some(format!("haira.Join({})", args)),
            _ => None,
        };
    }

    None
}

/// Try to resolve a method call as a stdlib call.
/// Handles `io.println(...)` parsed as MethodCall { receiver: Identifier("io"), method: "println" }.
pub fn resolve_stdlib_method_call(mc: &MethodCallExpr) -> Option<String> {
    if let ExprKind::Identifier(module) = &mc.receiver.node {
        let method = mc.method.node.as_str();
        let args = mc
            .args
            .iter()
            .map(|a| expr_to_go(&a.value))
            .collect::<Vec<_>>()
            .join(", ");

        return resolve_qualified(module.as_str(), method, &args, &to_call_expr_stub(mc));
    }
    None
}

/// Resolve a qualified module.method(args) call to Go code.
fn resolve_qualified(module: &str, method: &str, args: &str, call: &CallExpr) -> Option<String> {
    match (module, method) {
        // --- io module ---
        ("io", "println") => Some(format!("haira.Println({})", args)),
        ("io", "print") => Some(format!("haira.Print({})", args)),

        // --- http module ---
        ("http", "get") => Some(format!("haira.HttpGet({})", args)),
        ("http", "get_with_headers") => Some(format!("haira.HttpGetWithHeaders({})", args)),
        ("http", "post") => Some(format!("haira.HttpPost({})", args)),
        ("http", "post_with_headers") => Some(format!("haira.HttpPostWithHeaders({})", args)),
        ("http", "put") => Some(format!("haira.HttpPut({})", args)),
        ("http", "put_with_headers") => Some(format!("haira.HttpPutWithHeaders({})", args)),
        ("http", "delete") => Some(format!("haira.HttpDelete({})", args)),
        ("http", "delete_with_headers") => Some(format!("haira.HttpDeleteWithHeaders({})", args)),
        ("http", "Server") => resolve_server_call(call),

        // --- json module ---
        ("json", "marshal") => Some(format!("haira.JSONMarshal({})", args)),
        ("json", "unmarshal") => Some(format!("haira.JSONUnmarshal({})", args)),

        // --- postgres module ---
        ("postgres", "connect") => Some(format!("haira.PostgresConnect({})", args)),

        // --- slack module ---
        ("slack", "send") => Some(format!("haira.SlackSend({})", args)),

        // --- excel module ---
        ("excel", "open") => Some(format!("haira.ExcelOpen({})", args)),

        // --- time module ---
        ("time", "sleep") => Some(format!("haira.TimeSleep({})", args)),
        ("time", "now") => Some("haira.TimeNow()".to_string()),
        ("time", "slug") => Some("haira.TimeSlug()".to_string()),

        _ => None,
    }
}

/// Resolve http.Server([...]) call with workflow def references.
fn resolve_server_call(call: &CallExpr) -> Option<String> {
    if call.args.len() == 1 {
        if let ExprKind::List(items) = &call.args[0].value.node {
            let refs = items
                .iter()
                .map(|item| {
                    if let ExprKind::Identifier(name) = &item.node {
                        format!("workflowDef{}", snake_to_pascal(name))
                    } else {
                        expr_to_go(item)
                    }
                })
                .collect::<Vec<_>>()
                .join(", ");
            return Some(format!("haira.NewServer([]*haira.WorkflowDef{{{}}})", refs));
        }
    }
    let args = call
        .args
        .iter()
        .map(|a| expr_to_go(&a.value))
        .collect::<Vec<_>>()
        .join(", ");
    Some(format!("haira.NewServer({})", args))
}

/// Create a minimal CallExpr-like view from a MethodCallExpr for reuse.
fn to_call_expr_stub(mc: &MethodCallExpr) -> CallExpr {
    CallExpr {
        callee: Box::new((*mc.receiver).clone()),
        args: mc.args.clone(),
    }
}

/// Convert snake_case to PascalCase.
fn snake_to_pascal(name: &str) -> String {
    name.split('_')
        .map(|part| {
            let mut chars = part.chars();
            match chars.next() {
                Some(c) => c.to_uppercase().to_string() + chars.as_str(),
                None => String::new(),
            }
        })
        .collect::<String>()
}

/// Return whether a Haira import path maps to a stdlib module.
pub fn is_stdlib_import(path: &str) -> bool {
    matches!(
        path,
        "io" | "http" | "env" | "json" | "postgres" | "slack" | "excel" | "time"
    )
}
