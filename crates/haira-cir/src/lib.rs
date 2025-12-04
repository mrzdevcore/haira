//! # Haira Canonical IR (CIR)
//!
//! The Canonical Intermediate Representation is the format that AI
//! outputs when interpreting developer intent. It's a JSON-based, deterministic
//! representation that can be reliably converted to Haira AST.
//!
//! ## Design Goals
//!
//! - **Deterministic**: Same input always produces the same output
//! - **Type-safe**: All operations have defined type semantics
//! - **Sandboxed**: Only allowed operations, no arbitrary code execution
//! - **Serializable**: JSON format for easy transmission and caching
//!
//! ## Example
//!
//! ```json
//! {
//!   "function": "summarize_user_activity",
//!   "params": [{"name": "user", "type": "User"}],
//!   "returns": "ActivitySummary",
//!   "body": [
//!     {"op": "get_field", "source": "user", "field": "activities", "result": "acts"},
//!     {"op": "count", "source": "acts", "result": "total"},
//!     {"op": "construct", "type": "ActivitySummary", "fields": {...}, "result": "return"}
//!   ]
//! }
//! ```

mod function;
mod operations;
mod types;
mod validation;

pub use function::*;
pub use operations::*;
pub use types::*;
pub use validation::*;

use serde::{Deserialize, Serialize};

/// Result of AI interpretation.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AIResponse {
    /// Whether interpretation was successful
    pub success: bool,
    /// The interpreted function (if successful)
    #[serde(skip_serializing_if = "Option::is_none")]
    pub interpretation: Option<CIRFunction>,
    /// Confidence score (0.0 - 1.0)
    pub confidence: f64,
    /// Alternative interpretations (if ambiguous)
    #[serde(default)]
    pub alternatives: Vec<CIRFunction>,
    /// Error message (if failed)
    #[serde(skip_serializing_if = "Option::is_none")]
    pub error: Option<String>,
}

/// Request for AI interpretation.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AIRequest {
    /// Type of request
    pub request_type: RequestType,
    /// Function name to interpret
    pub function_name: String,
    /// Context for interpretation
    pub context: InterpretationContext,
}

/// Type of AI request.
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum RequestType {
    /// Infer intent and generate implementation
    InferIntent,
    /// Generate type definition
    GenerateType,
    /// Suggest completion
    Suggest,
}

/// Context provided to AI for interpretation.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct InterpretationContext {
    /// Types currently in scope
    pub types_in_scope: Vec<TypeDefinition>,
    /// Information about the call site
    pub call_site: CallSiteInfo,
    /// Project-level schema information
    #[serde(default)]
    pub project_schema: ProjectSchema,
}

/// Information about where the function is being called.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CallSiteInfo {
    /// Source file
    pub file: String,
    /// Line number
    pub line: u32,
    /// Arguments at call site
    pub arguments: Vec<ArgumentInfo>,
    /// Expected return type (if known)
    #[serde(skip_serializing_if = "Option::is_none")]
    pub expected_return: Option<String>,
}

/// Information about an argument at call site.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ArgumentInfo {
    /// Argument name (if named)
    #[serde(skip_serializing_if = "Option::is_none")]
    pub name: Option<String>,
    /// Inferred type
    #[serde(rename = "type")]
    pub ty: String,
}

/// Project-level configuration.
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct ProjectSchema {
    /// Whether project has database
    #[serde(default)]
    pub has_database: bool,
    /// Whether project has HTTP
    #[serde(default)]
    pub has_http: bool,
    /// Database type (if any)
    #[serde(skip_serializing_if = "Option::is_none")]
    pub database_type: Option<String>,
}
