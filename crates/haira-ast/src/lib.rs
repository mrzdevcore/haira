//! # Haira AST
//!
//! Abstract Syntax Tree definitions for the Haira programming language.
//!
//! The AST represents the syntactic structure of Haira source code after parsing.
//! It preserves source locations for error reporting and is the input to
//! name resolution and type checking.

mod span;
mod ast;

pub use span::{Span, Spanned};
pub use ast::*;
