//! Haira → Go type mapping.

use haira_ast::Type;

/// Convert a Haira type to its Go representation.
pub fn haira_type_to_go(ty: &Type) -> String {
    match ty {
        Type::Named(name) => match name.as_str() {
            "int" => "int".to_string(),
            "float" => "float64".to_string(),
            "string" => "string".to_string(),
            "bool" => "bool".to_string(),
            "any" => "any".to_string(),
            "map" => "map[string]any".to_string(),
            other => other.to_string(), // User-defined types
        },
        Type::List(inner) => format!("[]{}", haira_type_to_go(&inner.node)),
        Type::Map { key, value } => {
            format!(
                "map[{}]{}",
                haira_type_to_go(&key.node),
                haira_type_to_go(&value.node)
            )
        }
        Type::Option(inner) => {
            // Go doesn't have Option — use pointer for value types, nil-able for reference types
            format!("*{}", haira_type_to_go(&inner.node))
        }
        Type::Function { params, ret } => {
            let params_str = params
                .iter()
                .map(|p| haira_type_to_go(&p.node))
                .collect::<Vec<_>>()
                .join(", ");
            format!("func({}) {}", params_str, haira_type_to_go(&ret.node))
        }
        Type::Union(_) => "any".to_string(), // Go doesn't have union types
        Type::Generic { name, args } => {
            let args_str = args
                .iter()
                .map(|a| haira_type_to_go(&a.node))
                .collect::<Vec<_>>()
                .join(", ");
            format!("{}[{}]", name, args_str)
        }
    }
}
