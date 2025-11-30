//! Code generation for the Haira programming language.
//!
//! This crate handles lowering MIR to native code via LLVM.
//! Currently a placeholder - full LLVM integration pending.

use haira_mir::MirFunction;

/// Code generation options.
#[derive(Default)]
pub struct CodegenOptions {
    /// Optimization level (0-3).
    pub opt_level: u8,
    /// Generate debug info.
    pub debug_info: bool,
    /// Target triple (e.g., "x86_64-unknown-linux-gnu").
    pub target: Option<String>,
}

/// Generated code output.
pub struct CodegenOutput {
    /// Object file bytes.
    pub object: Vec<u8>,
    /// Whether compilation was successful.
    pub success: bool,
}

/// Generate code from MIR.
pub fn codegen(_functions: &[MirFunction], _options: CodegenOptions) -> CodegenOutput {
    // TODO: Implement LLVM code generation
    CodegenOutput {
        object: Vec::new(),
        success: true,
    }
}

/// Compile to executable.
pub fn compile_executable(
    _functions: &[MirFunction],
    _output_path: &std::path::Path,
    _options: CodegenOptions,
) -> Result<(), CodegenError> {
    // TODO: Implement full compilation pipeline
    Ok(())
}

/// Code generation error.
#[derive(Debug, thiserror::Error)]
pub enum CodegenError {
    #[error("LLVM error: {0}")]
    LlvmError(String),
    #[error("Linker error: {0}")]
    LinkerError(String),
    #[error("IO error: {0}")]
    IoError(#[from] std::io::Error),
}
