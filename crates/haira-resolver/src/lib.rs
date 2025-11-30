//! Name resolution for the Haira programming language.
//!
//! This crate handles:
//! - Resolving identifiers to their definitions
//! - Building scope trees
//! - Detecting undefined references
//! - Collecting unresolved function calls for AI interpretation

use haira_ast::SourceFile;
use rustc_hash::FxHashMap;
use smol_str::SmolStr;

/// Result of name resolution.
pub struct ResolvedModule {
    /// Map from identifier spans to their definitions.
    pub definitions: FxHashMap<usize, Definition>,
    /// Unresolved function calls that need AI interpretation.
    pub unresolved_calls: Vec<UnresolvedCall>,
    /// Resolution errors.
    pub errors: Vec<ResolutionError>,
}

/// A resolved definition.
#[derive(Debug, Clone)]
pub enum Definition {
    /// Local variable.
    Local {
        name: SmolStr,
        span: std::ops::Range<usize>,
    },
    /// Function parameter.
    Parameter { name: SmolStr, index: usize },
    /// Type definition.
    TypeDef { name: SmolStr },
    /// Function definition.
    Function { name: SmolStr },
    /// Built-in.
    Builtin { name: SmolStr },
}

/// An unresolved function call that needs AI interpretation.
#[derive(Debug, Clone)]
pub struct UnresolvedCall {
    /// Function name.
    pub name: SmolStr,
    /// Span in source.
    pub span: std::ops::Range<usize>,
    /// Argument count.
    pub arg_count: usize,
    /// Context type if this is a method call.
    pub receiver_type: Option<SmolStr>,
}

/// Resolution error.
#[derive(Debug, Clone)]
pub struct ResolutionError {
    pub message: String,
    pub span: std::ops::Range<usize>,
}

/// Resolve names in a source file.
pub fn resolve(_ast: &SourceFile) -> ResolvedModule {
    // TODO: Implement name resolution
    ResolvedModule {
        definitions: FxHashMap::default(),
        unresolved_calls: Vec::new(),
        errors: Vec::new(),
    }
}
