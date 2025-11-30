//! Code generation for the Haira programming language.
//!
//! This crate handles lowering AST to native code via Cranelift.

mod cir_to_ast;
mod compiler;
mod runtime;

pub use cir_to_ast::{cir_to_function_def, cir_types_to_ast, ConversionError};
pub use compiler::{compile_to_executable, CodegenError, CodegenOptions};
