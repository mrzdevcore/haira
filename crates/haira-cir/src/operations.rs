//! CIR Operations - the building blocks of AI-generated code.

use serde::{Deserialize, Serialize};
use std::collections::HashMap;

/// A CIR operation - the basic unit of generated code.
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "kind", rename_all = "snake_case")]
pub enum CIROperation {
    // ========================================================================
    // Data Access
    // ========================================================================
    /// Get a field from an object: `obj.field`
    GetField {
        source: String,
        field: String,
        result: String,
    },

    /// Get an index from a collection: `arr[i]`
    GetIndex {
        source: String,
        index: CIRValue,
        result: String,
    },

    /// Set a field on an object
    SetField {
        target: String,
        field: String,
        value: CIRValue,
    },

    // ========================================================================
    // Collection Operations
    // ========================================================================
    /// Map over a collection
    Map {
        source: String,
        /// Variable name for element
        element_var: String,
        /// Operations to apply to each element
        transform: Vec<CIROperation>,
        result: String,
    },

    /// Filter a collection
    Filter {
        source: String,
        element_var: String,
        /// Operations that produce a bool
        predicate: Vec<CIROperation>,
        result: String,
    },

    /// Reduce a collection
    Reduce {
        source: String,
        initial: CIRValue,
        accumulator_var: String,
        element_var: String,
        reducer: Vec<CIROperation>,
        result: String,
    },

    /// Group by a key
    GroupBy {
        source: String,
        element_var: String,
        /// Operations that produce the key
        key: Vec<CIROperation>,
        result: String,
    },

    /// Sort a collection
    Sort {
        source: String,
        element_var: String,
        /// Operations that produce the sort key
        key: Vec<CIROperation>,
        #[serde(default)]
        descending: bool,
        result: String,
    },

    /// Take first N elements
    Take {
        source: String,
        count: CIRValue,
        result: String,
    },

    /// Skip first N elements
    Skip {
        source: String,
        count: CIRValue,
        result: String,
    },

    /// Count elements
    Count { source: String, result: String },

    /// Find first matching element
    Find {
        source: String,
        element_var: String,
        predicate: Vec<CIROperation>,
        result: String,
    },

    /// Check if any element matches
    Any {
        source: String,
        element_var: String,
        predicate: Vec<CIROperation>,
        result: String,
    },

    /// Check if all elements match
    All {
        source: String,
        element_var: String,
        predicate: Vec<CIROperation>,
        result: String,
    },

    // ========================================================================
    // Aggregations
    // ========================================================================
    /// Sum of numbers
    Sum { source: String, result: String },

    /// Minimum value
    Min { source: String, result: String },

    /// Maximum value
    Max { source: String, result: String },

    /// Average value
    Avg { source: String, result: String },

    /// Maximum by a key
    MaxBy {
        source: String,
        element_var: String,
        key: Vec<CIROperation>,
        result: String,
    },

    /// Minimum by a key
    MinBy {
        source: String,
        element_var: String,
        key: Vec<CIROperation>,
        result: String,
    },

    // ========================================================================
    // Control Flow
    // ========================================================================
    /// Conditional
    If {
        /// Operations that produce the condition (bool)
        condition: Vec<CIROperation>,
        then_ops: Vec<CIROperation>,
        else_ops: Vec<CIROperation>,
        result: String,
    },

    /// Pattern match
    Match {
        subject: String,
        arms: Vec<MatchArm>,
        result: String,
    },

    /// Loop over items
    Loop {
        source: String,
        element_var: String,
        body: Vec<CIROperation>,
        result: String,
    },

    // ========================================================================
    // Construction
    // ========================================================================
    /// Construct a type instance
    Construct {
        #[serde(rename = "type")]
        ty: String,
        fields: HashMap<String, CIRValue>,
        result: String,
    },

    /// Create a list
    CreateList {
        elements: Vec<CIRValue>,
        result: String,
    },

    /// Create a map
    CreateMap {
        entries: Vec<(CIRValue, CIRValue)>,
        result: String,
    },

    // ========================================================================
    // Primitives
    // ========================================================================
    /// Binary operation
    BinaryOp {
        op: BinaryOperator,
        left: CIRValue,
        right: CIRValue,
        result: String,
    },

    /// Unary operation
    UnaryOp {
        op: UnaryOperator,
        operand: CIRValue,
        result: String,
    },

    /// Call a function
    Call {
        function: String,
        args: Vec<CIRValue>,
        result: String,
    },

    /// Literal value
    Literal { value: CIRValue, result: String },

    /// Variable reference
    Var { name: String, result: String },

    // ========================================================================
    // I/O (Abstract)
    // ========================================================================
    /// Database query
    DbQuery {
        query_type: DbQueryType,
        table: String,
        #[serde(default)]
        filters: Vec<DbFilter>,
        #[serde(default)]
        order_by: Option<String>,
        #[serde(default)]
        limit: Option<u32>,
        result: String,
    },

    /// HTTP request
    HttpRequest {
        method: HttpMethod,
        url: CIRValue,
        #[serde(skip_serializing_if = "Option::is_none")]
        body: Option<CIRValue>,
        #[serde(default)]
        headers: HashMap<String, CIRValue>,
        result: String,
    },

    /// File read
    FileRead { path: CIRValue, result: String },

    /// File write
    FileWrite { path: CIRValue, content: CIRValue },

    // ========================================================================
    // String Operations
    // ========================================================================
    /// Format a string with interpolation
    Format {
        template: String,
        values: HashMap<String, CIRValue>,
        result: String,
    },

    /// Concatenate strings
    Concat {
        parts: Vec<CIRValue>,
        result: String,
    },

    // ========================================================================
    // Return
    // ========================================================================
    /// Return a value
    Return { value: CIRValue },
}

/// A value in CIR.
///
/// JSON representation uses untagged deserialization with the following priority:
/// - `null` -> None
/// - `true`/`false` -> Bool
/// - number (no decimal) -> Int
/// - number (with decimal) -> Float
/// - `{"ref": "name"}` -> Ref (variable reference)
/// - `"string"` -> String (literal string)
/// - object with "kind" -> Operation
///
/// For serialization, Ref uses `{"ref": "name"}` format.
#[derive(Debug, Clone)]
pub enum CIRValue {
    /// Reference to a variable
    Ref(String),
    /// Literal integer
    Int(i64),
    /// Literal float
    Float(f64),
    /// Literal string
    String(String),
    /// Literal boolean
    Bool(bool),
    /// None value
    None,
    /// Inline operation (for complex expressions)
    Operation(Box<CIROperation>),
}

impl serde::Serialize for CIRValue {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeMap;
        match self {
            CIRValue::Ref(name) => {
                let mut map = serializer.serialize_map(Some(1))?;
                map.serialize_entry("ref", name)?;
                map.end()
            }
            CIRValue::Int(n) => serializer.serialize_i64(*n),
            CIRValue::Float(f) => serializer.serialize_f64(*f),
            CIRValue::String(s) => serializer.serialize_str(s),
            CIRValue::Bool(b) => serializer.serialize_bool(*b),
            CIRValue::None => serializer.serialize_none(),
            CIRValue::Operation(op) => op.serialize(serializer),
        }
    }
}

impl<'de> serde::Deserialize<'de> for CIRValue {
    fn deserialize<D>(deserializer: D) -> Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        use serde::de::{self, MapAccess, Visitor};

        struct CIRValueVisitor;

        impl<'de> Visitor<'de> for CIRValueVisitor {
            type Value = CIRValue;

            fn expecting(&self, formatter: &mut std::fmt::Formatter) -> std::fmt::Result {
                formatter.write_str("a CIR value (null, bool, number, string, or object)")
            }

            fn visit_bool<E>(self, v: bool) -> Result<Self::Value, E> {
                Ok(CIRValue::Bool(v))
            }

            fn visit_i64<E>(self, v: i64) -> Result<Self::Value, E> {
                Ok(CIRValue::Int(v))
            }

            fn visit_u64<E>(self, v: u64) -> Result<Self::Value, E> {
                Ok(CIRValue::Int(v as i64))
            }

            fn visit_f64<E>(self, v: f64) -> Result<Self::Value, E> {
                Ok(CIRValue::Float(v))
            }

            fn visit_str<E>(self, v: &str) -> Result<Self::Value, E>
            where
                E: de::Error,
            {
                Ok(CIRValue::String(v.to_string()))
            }

            fn visit_string<E>(self, v: String) -> Result<Self::Value, E> {
                Ok(CIRValue::String(v))
            }

            fn visit_none<E>(self) -> Result<Self::Value, E> {
                Ok(CIRValue::None)
            }

            fn visit_unit<E>(self) -> Result<Self::Value, E> {
                Ok(CIRValue::None)
            }

            fn visit_map<M>(self, mut map: M) -> Result<Self::Value, M::Error>
            where
                M: MapAccess<'de>,
            {
                // Check if this is a {"ref": "name"} or {"literal": value} object
                let mut ref_name: Option<String> = None;
                let mut literal_value: Option<serde_json::Value> = None;
                let mut is_operation = false;
                let mut collected: std::collections::HashMap<String, serde_json::Value> =
                    std::collections::HashMap::new();

                while let Some(key) = map.next_key::<String>()? {
                    let value: serde_json::Value = map.next_value()?;
                    if key == "ref" {
                        if let serde_json::Value::String(s) = &value {
                            ref_name = Some(s.clone());
                        }
                    }
                    if key == "literal" || key == "value" {
                        literal_value = Some(value.clone());
                    }
                    if key == "kind" || key == "op" {
                        is_operation = true;
                    }
                    collected.insert(key, value);
                }

                // If it's just {"ref": "name"}, return a Ref
                if let Some(name) = ref_name {
                    if collected.len() == 1 {
                        return Ok(CIRValue::Ref(name));
                    }
                }

                // If it's {"literal": value} or {"value": value}, extract the literal
                if let Some(lit_val) = literal_value {
                    if collected.len() == 1 {
                        return match lit_val {
                            serde_json::Value::String(s) => Ok(CIRValue::String(s)),
                            serde_json::Value::Number(n) => {
                                if let Some(i) = n.as_i64() {
                                    Ok(CIRValue::Int(i))
                                } else if let Some(f) = n.as_f64() {
                                    Ok(CIRValue::Float(f))
                                } else {
                                    Err(de::Error::custom("invalid number"))
                                }
                            }
                            serde_json::Value::Bool(b) => Ok(CIRValue::Bool(b)),
                            serde_json::Value::Null => Ok(CIRValue::None),
                            _ => Err(de::Error::custom("unsupported literal type")),
                        };
                    }
                }

                // Otherwise try to parse as an operation
                if is_operation {
                    let json_obj = serde_json::Value::Object(collected.into_iter().collect());
                    match serde_json::from_value::<CIROperation>(json_obj) {
                        Ok(op) => return Ok(CIRValue::Operation(Box::new(op))),
                        Err(e) => {
                            return Err(de::Error::custom(format!("invalid operation: {}", e)))
                        }
                    }
                }

                Err(de::Error::custom(
                    "expected a CIR value object with 'ref', 'literal', or 'kind' key",
                ))
            }
        }

        deserializer.deserialize_any(CIRValueVisitor)
    }
}

impl CIRValue {
    pub fn var(name: impl Into<String>) -> Self {
        CIRValue::Ref(name.into())
    }

    pub fn string(s: impl Into<String>) -> Self {
        CIRValue::String(s.into())
    }

    pub fn int(n: i64) -> Self {
        CIRValue::Int(n)
    }

    pub fn bool(b: bool) -> Self {
        CIRValue::Bool(b)
    }
}

/// Binary operators.
#[derive(Debug, Clone, Copy, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum BinaryOperator {
    Add,
    Sub,
    Mul,
    Div,
    Mod,
    Eq,
    Ne,
    Lt,
    Gt,
    Le,
    Ge,
    And,
    Or,
}

/// Unary operators.
#[derive(Debug, Clone, Copy, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum UnaryOperator {
    Neg,
    Not,
}

/// Match arm in pattern matching.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MatchArm {
    /// Pattern to match
    pub pattern: CIRPattern,
    /// Operations to execute if matched
    pub body: Vec<CIROperation>,
}

/// Pattern for matching.
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "kind", rename_all = "snake_case")]
pub enum CIRPattern {
    /// Wildcard: matches anything
    Wildcard,
    /// Literal value
    Literal { value: CIRValue },
    /// Variable binding
    Binding { name: String },
    /// Constructor pattern
    Constructor { name: String, fields: Vec<String> },
}

/// Database query types.
#[derive(Debug, Clone, Copy, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum DbQueryType {
    Select,
    Insert,
    Update,
    Delete,
    Count,
}

/// Database filter.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DbFilter {
    pub field: String,
    pub op: FilterOp,
    pub value: CIRValue,
}

/// Filter operators.
#[derive(Debug, Clone, Copy, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum FilterOp {
    Eq,
    Ne,
    Lt,
    Gt,
    Le,
    Ge,
    Like,
    In,
}

/// HTTP methods.
#[derive(Debug, Clone, Copy, Serialize, Deserialize)]
#[serde(rename_all = "UPPERCASE")]
pub enum HttpMethod {
    Get,
    Post,
    Put,
    Patch,
    Delete,
}
