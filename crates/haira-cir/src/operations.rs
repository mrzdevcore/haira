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
    Count {
        source: String,
        result: String,
    },

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
    Sum {
        source: String,
        result: String,
    },

    /// Minimum value
    Min {
        source: String,
        result: String,
    },

    /// Maximum value
    Max {
        source: String,
        result: String,
    },

    /// Average value
    Avg {
        source: String,
        result: String,
    },

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
    Literal {
        value: CIRValue,
        result: String,
    },

    /// Variable reference
    Var {
        name: String,
        result: String,
    },

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
    FileRead {
        path: CIRValue,
        result: String,
    },

    /// File write
    FileWrite {
        path: CIRValue,
        content: CIRValue,
    },

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
    Return {
        value: CIRValue,
    },
}

/// A value in CIR.
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(untagged)]
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
    Constructor {
        name: String,
        fields: Vec<String>,
    },
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
