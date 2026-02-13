//! Stdlib mapping — Haira imports → Go runtime calls.

use haira_ast::{CallExpr, ExprKind, MethodCallExpr};

use crate::expressions::expr_to_go;

/// Try to resolve a call as a stdlib call.
/// Returns Some(go_code) if it's a known stdlib function, None otherwise.
pub fn resolve_stdlib_call(call: &CallExpr) -> Option<String> {
    // Check for qualified calls like io.println, env(...)
    if let ExprKind::Field(field) = &call.callee.node {
        if let ExprKind::Identifier(module) = &field.object.node {
            let method = field.field.node.as_str();
            let args = call
                .args
                .iter()
                .map(|a| expr_to_go(&a.value))
                .collect::<Vec<_>>()
                .join(", ");

            return match (module.as_str(), method) {
                // io module
                ("io", "println") => Some(format!("haira.Println({})", args)),
                ("io", "print") => Some(format!("haira.Print({})", args)),
                // http module
                ("http", "get") => Some(format!("haira.HttpGet({})", args)),
                ("http", "Server") => Some(format!("haira.NewServer({})", args)),
                // json module
                ("json", "marshal") => Some(format!("haira.JSONMarshal({})", args)),
                ("json", "unmarshal") => Some(format!("haira.JSONUnmarshal({})", args)),
                _ => None,
            };
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

        return match (module.as_str(), method) {
            // io module
            ("io", "println") => Some(format!("haira.Println({})", args)),
            ("io", "print") => Some(format!("haira.Print({})", args)),
            // http module
            ("http", "get") => Some(format!("haira.HttpGet({})", args)),
            ("http", "Server") => {
                if mc.args.len() == 1 {
                    if let ExprKind::List(items) = &mc.args[0].value.node {
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
                Some(format!("haira.NewServer({})", args))
            }
            // json module
            ("json", "marshal") => Some(format!("haira.JSONMarshal({})", args)),
            ("json", "unmarshal") => Some(format!("haira.JSONUnmarshal({})", args)),
            _ => None,
        };
    }
    None
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

/// Return the set of Haira imports that map to stdlib modules.
pub fn is_stdlib_import(path: &str) -> bool {
    matches!(path, "io" | "http" | "env" | "json")
}
