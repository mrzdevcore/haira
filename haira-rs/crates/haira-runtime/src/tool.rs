//! Tool definitions and registry.
//!
//! Tools are functions that agents can call. The registry holds tool metadata
//! (name, description, parameter schema) so the runtime can pass them as
//! tool specs to the LLM API. Tool bodies are executed by the interpreter,
//! not by the runtime directly.

use serde_json::json;

/// A tool definition with metadata for LLM function calling.
#[derive(Debug, Clone)]
pub struct ToolDef {
    pub name: String,
    pub description: String,
    /// JSON Schema for the tool's parameters (OpenAI function calling format).
    pub parameters_schema: serde_json::Value,
}

/// Registry of available tools.
#[derive(Debug, Clone, Default)]
pub struct ToolRegistry {
    tools: Vec<ToolDef>,
}

impl ToolRegistry {
    pub fn new() -> Self {
        Self { tools: Vec::new() }
    }

    pub fn register(&mut self, tool: ToolDef) {
        self.tools.push(tool);
    }

    pub fn get(&self, name: &str) -> Option<&ToolDef> {
        self.tools.iter().find(|t| t.name == name)
    }

    pub fn all(&self) -> &[ToolDef] {
        &self.tools
    }

    pub fn is_empty(&self) -> bool {
        self.tools.is_empty()
    }

    /// Generate OpenAI-compatible tool specifications for the chat API.
    pub fn to_tool_specs(&self) -> Vec<serde_json::Value> {
        self.tools
            .iter()
            .map(|t| {
                json!({
                    "type": "function",
                    "function": {
                        "name": t.name,
                        "description": t.description,
                        "parameters": t.parameters_schema,
                    }
                })
            })
            .collect()
    }
}

/// Build a JSON Schema object for tool parameters.
///
/// Each parameter is `(name, type_string)` where type_string is a Haira type
/// like "string", "int", "float", "bool".
pub fn build_tool_schema(params: &[(String, String)]) -> serde_json::Value {
    let mut properties = serde_json::Map::new();
    let mut required = Vec::new();

    for (name, ty) in params {
        let json_type = match ty.as_str() {
            "int" => "integer",
            "float" | "number" => "number",
            "bool" => "boolean",
            "string" | _ => "string",
        };
        properties.insert(
            name.clone(),
            json!({ "type": json_type }),
        );
        required.push(serde_json::Value::String(name.clone()));
    }

    json!({
        "type": "object",
        "properties": properties,
        "required": required,
    })
}
