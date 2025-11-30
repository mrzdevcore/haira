//! Type definitions for CIR.

use serde::{Deserialize, Serialize};

/// A type definition in the context.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TypeDefinition {
    /// Type name
    pub name: String,
    /// Fields of the type
    pub fields: Vec<FieldDefinition>,
}

/// A field in a type definition.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FieldDefinition {
    /// Field name
    pub name: String,
    /// Field type
    #[serde(rename = "type")]
    pub ty: String,
    /// Whether field is optional
    #[serde(default)]
    pub optional: bool,
    /// Default value (as string representation)
    #[serde(skip_serializing_if = "Option::is_none")]
    pub default: Option<String>,
}

/// A CIR type reference.
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(untagged)]
pub enum CIRType {
    /// Simple named type: "int", "string", "User"
    Simple(String),
    /// Complex type with structure
    Complex(CIRTypeKind),
}

/// Complex type kinds.
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "kind", rename_all = "snake_case")]
pub enum CIRTypeKind {
    /// List type
    List { element: Box<CIRType> },
    /// Map type
    Map {
        key: Box<CIRType>,
        value: Box<CIRType>,
    },
    /// Option type
    Option { inner: Box<CIRType> },
    /// Function type
    Function {
        params: Vec<CIRType>,
        returns: Box<CIRType>,
    },
    /// Union type
    Union { variants: Vec<CIRType> },
}

impl CIRType {
    /// Create a simple type.
    pub fn simple(name: impl Into<String>) -> Self {
        CIRType::Simple(name.into())
    }

    /// Create a list type.
    pub fn list(element: CIRType) -> Self {
        CIRType::Complex(CIRTypeKind::List {
            element: Box::new(element),
        })
    }

    /// Create a map type.
    pub fn map(key: CIRType, value: CIRType) -> Self {
        CIRType::Complex(CIRTypeKind::Map {
            key: Box::new(key),
            value: Box::new(value),
        })
    }

    /// Create an option type.
    pub fn option(inner: CIRType) -> Self {
        CIRType::Complex(CIRTypeKind::Option {
            inner: Box::new(inner),
        })
    }

    /// Check if this is a primitive type.
    pub fn is_primitive(&self) -> bool {
        match self {
            CIRType::Simple(name) => {
                matches!(name.as_str(), "int" | "float" | "string" | "bool" | "none")
            }
            CIRType::Complex(_) => false,
        }
    }
}

impl From<&str> for CIRType {
    fn from(s: &str) -> Self {
        CIRType::Simple(s.to_string())
    }
}

impl From<String> for CIRType {
    fn from(s: String) -> Self {
        CIRType::Simple(s)
    }
}
