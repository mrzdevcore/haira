//! CIR Function definitions.

use serde::{Deserialize, Serialize};
use crate::{CIROperation, CIRType, TypeDefinition};

/// A complete function definition in CIR.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CIRFunction {
    /// Function name
    pub name: String,

    /// Description of what the function does
    #[serde(skip_serializing_if = "Option::is_none")]
    pub description: Option<String>,

    /// Parameters
    pub params: Vec<CIRParam>,

    /// Return type
    pub returns: CIRType,

    /// New types to generate (if the function needs custom types)
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub new_types: Vec<TypeDefinition>,

    /// Function body as sequence of operations
    pub body: Vec<CIROperation>,
}

/// A function parameter in CIR.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CIRParam {
    /// Parameter name
    pub name: String,

    /// Parameter type
    #[serde(rename = "type")]
    pub ty: CIRType,

    /// Default value (if optional)
    #[serde(skip_serializing_if = "Option::is_none")]
    pub default: Option<String>,
}

impl CIRFunction {
    /// Create a new function.
    pub fn new(name: impl Into<String>) -> Self {
        Self {
            name: name.into(),
            description: None,
            params: Vec::new(),
            returns: CIRType::simple("none"),
            new_types: Vec::new(),
            body: Vec::new(),
        }
    }

    /// Set description.
    pub fn with_description(mut self, desc: impl Into<String>) -> Self {
        self.description = Some(desc.into());
        self
    }

    /// Add a parameter.
    pub fn with_param(mut self, name: impl Into<String>, ty: impl Into<CIRType>) -> Self {
        self.params.push(CIRParam {
            name: name.into(),
            ty: ty.into(),
            default: None,
        });
        self
    }

    /// Set return type.
    pub fn returning(mut self, ty: impl Into<CIRType>) -> Self {
        self.returns = ty.into();
        self
    }

    /// Add an operation to the body.
    pub fn with_op(mut self, op: CIROperation) -> Self {
        self.body.push(op);
        self
    }

    /// Add a new type definition.
    pub fn with_type(mut self, ty: TypeDefinition) -> Self {
        self.new_types.push(ty);
        self
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::{CIRValue, FieldDefinition};

    #[test]
    fn test_function_serialization() {
        let func = CIRFunction::new("get_active_users")
            .with_description("Get all active users")
            .returning(CIRType::list(CIRType::simple("User")))
            .with_op(CIROperation::DbQuery {
                query_type: crate::DbQueryType::Select,
                table: "users".to_string(),
                filters: vec![crate::DbFilter {
                    field: "active".to_string(),
                    op: crate::FilterOp::Eq,
                    value: CIRValue::Bool(true),
                }],
                order_by: None,
                limit: None,
                result: "users".to_string(),
            })
            .with_op(CIROperation::Return {
                value: CIRValue::var("users"),
            });

        let json = serde_json::to_string_pretty(&func).unwrap();
        println!("{}", json);

        // Verify it round-trips
        let parsed: CIRFunction = serde_json::from_str(&json).unwrap();
        assert_eq!(parsed.name, "get_active_users");
    }

    #[test]
    fn test_complex_function() {
        let func = CIRFunction::new("summarize_user_activity")
            .with_description("Generate a summary of user activity")
            .with_param("user", "User")
            .returning("ActivitySummary")
            .with_type(TypeDefinition {
                name: "ActivitySummary".to_string(),
                fields: vec![
                    FieldDefinition {
                        name: "total".to_string(),
                        ty: "int".to_string(),
                        optional: false,
                        default: None,
                    },
                    FieldDefinition {
                        name: "most_common_type".to_string(),
                        ty: "string".to_string(),
                        optional: false,
                        default: None,
                    },
                ],
            })
            .with_op(CIROperation::GetField {
                source: "user".to_string(),
                field: "activity_log".to_string(),
                result: "activities".to_string(),
            })
            .with_op(CIROperation::Count {
                source: "activities".to_string(),
                result: "total".to_string(),
            })
            .with_op(CIROperation::Construct {
                ty: "ActivitySummary".to_string(),
                fields: [
                    ("total".to_string(), CIRValue::var("total")),
                    ("most_common_type".to_string(), CIRValue::string("login")),
                ]
                .into_iter()
                .collect(),
                result: "summary".to_string(),
            })
            .with_op(CIROperation::Return {
                value: CIRValue::var("summary"),
            });

        let json = serde_json::to_string_pretty(&func).unwrap();
        println!("{}", json);

        // Verify it parses back
        let parsed: CIRFunction = serde_json::from_str(&json).unwrap();
        assert_eq!(parsed.params.len(), 1);
        assert_eq!(parsed.new_types.len(), 1);
        assert_eq!(parsed.body.len(), 4);
    }
}
