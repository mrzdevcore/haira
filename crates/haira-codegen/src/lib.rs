//! Code generation for the Haira programming language.
//!
//! This crate handles lowering AST to native code via Cranelift.

mod compiler;
mod runtime;

pub use compiler::{compile_to_executable, CodegenError, CodegenOptions};
