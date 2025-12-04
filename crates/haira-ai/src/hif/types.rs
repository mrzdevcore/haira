//! HIF type definitions.

use serde::{Deserialize, Serialize};
use std::collections::BTreeMap;

/// A complete HIF file containing structs and intents.
#[derive(Debug, Clone, Default)]
pub struct HIFFile {
    /// Format version.
    pub version: u32,
    /// Struct type definitions (sorted by name).
    pub structs: BTreeMap<String, HIFStruct>,
    /// Intent function definitions (sorted by name).
    pub intents: BTreeMap<String, HIFIntent>,
}

/// A struct type definition with AI-inferred field types.
#[derive(Debug, Clone)]
pub struct HIFStruct {
    /// Struct name.
    pub name: String,
    /// Context hash for cache invalidation.
    pub hash: String,
    /// Fields with inferred types (ordered).
    pub fields: Vec<HIFField>,
}

/// A struct field with its inferred type.
#[derive(Debug, Clone)]
pub struct HIFField {
    /// Field name.
    pub name: String,
    /// Inferred type.
    pub ty: HIFType,
}

/// An intent function definition.
#[derive(Debug, Clone)]
pub struct HIFIntent {
    /// Function name.
    pub name: String,
    /// Context hash for cache invalidation.
    pub hash: String,
    /// Parameters with types.
    pub params: Vec<HIFParam>,
    /// Return type.
    pub returns: HIFType,
    /// Function body operations.
    pub body: Vec<HIFOperation>,
}

/// A function parameter.
#[derive(Debug, Clone)]
pub struct HIFParam {
    /// Parameter name.
    pub name: String,
    /// Parameter type.
    pub ty: HIFType,
}

/// HIF type representation.
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub enum HIFType {
    /// Integer type.
    Int,
    /// Floating point type.
    Float,
    /// String type.
    String,
    /// Boolean type.
    Bool,
    /// Void/unit type.
    Void,
    /// DateTime type.
    DateTime,
    /// Array type.
    Array(Box<HIFType>),
    /// Optional type.
    Optional(Box<HIFType>),
    /// Map type.
    Map(Box<HIFType>, Box<HIFType>),
    /// Custom struct type.
    Struct(String),
    /// Unknown type (to be inferred).
    Unknown,
}

impl HIFType {
    /// Parse a type string into HIFType.
    pub fn parse(s: &str) -> Self {
        let s = s.trim();

        // Check for optional type
        if let Some(inner) = s.strip_suffix('?') {
            return HIFType::Optional(Box::new(HIFType::parse(inner)));
        }

        // Check for array type
        if s.starts_with('[') && s.ends_with(']') {
            let inner = &s[1..s.len() - 1];
            return HIFType::Array(Box::new(HIFType::parse(inner)));
        }

        // Check for map type
        if s.starts_with('{') && s.ends_with('}') {
            let inner = &s[1..s.len() - 1];
            if let Some(colon_pos) = inner.find(':') {
                let key = inner[..colon_pos].trim();
                let value = inner[colon_pos + 1..].trim();
                return HIFType::Map(
                    Box::new(HIFType::parse(key)),
                    Box::new(HIFType::parse(value)),
                );
            }
        }

        // Primitive types
        match s {
            "int" => HIFType::Int,
            "float" => HIFType::Float,
            "string" => HIFType::String,
            "bool" => HIFType::Bool,
            "void" => HIFType::Void,
            "datetime" => HIFType::DateTime,
            _ => HIFType::Struct(s.to_string()),
        }
    }

    /// Convert to string representation.
    pub fn to_hif_string(&self) -> String {
        match self {
            HIFType::Int => "int".to_string(),
            HIFType::Float => "float".to_string(),
            HIFType::String => "string".to_string(),
            HIFType::Bool => "bool".to_string(),
            HIFType::Void => "void".to_string(),
            HIFType::DateTime => "datetime".to_string(),
            HIFType::Array(inner) => format!("[{}]", inner.to_hif_string()),
            HIFType::Optional(inner) => format!("{}?", inner.to_hif_string()),
            HIFType::Map(key, value) => {
                format!("{{{}: {}}}", key.to_hif_string(), value.to_hif_string())
            }
            HIFType::Struct(name) => name.clone(),
            HIFType::Unknown => "unknown".to_string(),
        }
    }
}

impl std::fmt::Display for HIFType {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.to_hif_string())
    }
}

/// An operation in the function body.
#[derive(Debug, Clone)]
pub struct HIFOperation {
    /// Operation kind.
    pub kind: HIFOpKind,
    /// Result variable name (if any).
    pub result: Option<String>,
    /// Result type (if any).
    pub result_type: Option<HIFType>,
}

/// Operation kinds.
#[derive(Debug, Clone)]
pub enum HIFOpKind {
    // Literals
    Literal(HIFValue),

    // Variable reference
    Var(String),

    // Return
    Return(String),

    // Binary operations
    Add(String, String),
    Sub(String, String),
    Mul(String, String),
    Div(String, String),
    Mod(String, String),
    Eq(String, String),
    Ne(String, String),
    Lt(String, String),
    Gt(String, String),
    Le(String, String),
    Ge(String, String),
    And(String, String),
    Or(String, String),

    // Unary operations
    Neg(String),
    Not(String),

    // Field access
    GetField(String, String),         // source, field
    SetField(String, String, String), // target, field, value

    // Index access
    GetIndex(String, String), // source, index

    // Collection operations
    Map {
        source: String,
        element_var: String,
        body: Vec<HIFOperation>,
    },
    Filter {
        source: String,
        element_var: String,
        body: Vec<HIFOperation>,
    },
    Reduce {
        source: String,
        initial: String,
        accumulator_var: String,
        element_var: String,
        body: Vec<HIFOperation>,
    },
    Sum(String),
    Min(String),
    Max(String),
    Avg(String),
    Count(String),
    Take(String, String), // source, count
    Skip(String, String), // source, count
    Find {
        source: String,
        element_var: String,
        body: Vec<HIFOperation>,
    },
    Any {
        source: String,
        element_var: String,
        body: Vec<HIFOperation>,
    },
    All {
        source: String,
        element_var: String,
        body: Vec<HIFOperation>,
    },

    // Control flow
    If {
        condition: Vec<HIFOperation>,
        then_ops: Vec<HIFOperation>,
        else_ops: Vec<HIFOperation>,
    },
    Loop {
        source: String,
        element_var: String,
        body: Vec<HIFOperation>,
    },

    // Construction
    Construct {
        ty: String,
        fields: Vec<(String, String)>, // field name, value ref
    },
    CreateList(Vec<String>),

    // Function call
    Call {
        function: String,
        args: Vec<String>,
    },

    // String operations
    Concat(Vec<String>),
    Format {
        template: String,
        values: Vec<(String, String)>, // key, value ref
    },
}

/// A literal value.
#[derive(Debug, Clone)]
pub enum HIFValue {
    Int(i64),
    Float(f64),
    String(String),
    Bool(bool),
    None,
}

impl HIFValue {
    /// Convert to HIF string representation.
    pub fn to_hif_string(&self) -> String {
        match self {
            HIFValue::Int(n) => n.to_string(),
            HIFValue::Float(f) => {
                let s = f.to_string();
                if s.contains('.') {
                    s
                } else {
                    format!("{}.0", s)
                }
            }
            HIFValue::String(s) => format!("\"{}\"", s.replace('\"', "\\\"")),
            HIFValue::Bool(b) => b.to_string(),
            HIFValue::None => "none".to_string(),
        }
    }

    /// Infer type from value.
    pub fn infer_type(&self) -> HIFType {
        match self {
            HIFValue::Int(_) => HIFType::Int,
            HIFValue::Float(_) => HIFType::Float,
            HIFValue::String(_) => HIFType::String,
            HIFValue::Bool(_) => HIFType::Bool,
            HIFValue::None => HIFType::Void,
        }
    }
}

impl HIFFile {
    /// Create a new empty HIF file.
    pub fn new() -> Self {
        Self {
            version: crate::hif::HIF_VERSION,
            structs: BTreeMap::new(),
            intents: BTreeMap::new(),
        }
    }

    /// Add a struct definition.
    pub fn add_struct(&mut self, s: HIFStruct) {
        self.structs.insert(s.name.clone(), s);
    }

    /// Add an intent definition.
    pub fn add_intent(&mut self, i: HIFIntent) {
        self.intents.insert(i.name.clone(), i);
    }

    /// Get a struct by name.
    pub fn get_struct(&self, name: &str) -> Option<&HIFStruct> {
        self.structs.get(name)
    }

    /// Get an intent by name.
    pub fn get_intent(&self, name: &str) -> Option<&HIFIntent> {
        self.intents.get(name)
    }

    /// Check if a struct exists.
    pub fn has_struct(&self, name: &str) -> bool {
        self.structs.contains_key(name)
    }

    /// Check if an intent exists.
    pub fn has_intent(&self, name: &str) -> bool {
        self.intents.contains_key(name)
    }
}
