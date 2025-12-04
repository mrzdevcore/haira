//! Compiler driver for the Haira programming language.
//!
//! This crate orchestrates the full compilation pipeline:
//! 1. Lexing
//! 2. Parsing
//! 3. Name resolution
//! 4. AI interpretation (for unresolved functions)
//! 5. Type checking
//! 6. HIR lowering
//! 7. MIR lowering
//! 8. Code generation

use haira_ai::{AIConfig, AIEngine};
use haira_codegen::CodegenOptions;
use std::path::Path;

/// Compiler configuration.
#[derive(Default)]
pub struct CompilerConfig {
    /// AI configuration.
    pub ai: AIConfig,
    /// Code generation options.
    pub codegen: CodegenOptions,
    /// Enable verbose output.
    pub verbose: bool,
}

/// Compilation result.
pub struct CompilationResult {
    /// Whether compilation succeeded.
    pub success: bool,
    /// Errors encountered.
    pub errors: Vec<CompilationError>,
    /// Warnings encountered.
    pub warnings: Vec<CompilationWarning>,
}

/// A compilation error.
#[derive(Debug)]
pub struct CompilationError {
    pub message: String,
    pub file: Option<String>,
    pub span: Option<std::ops::Range<usize>>,
}

/// A compilation warning.
#[derive(Debug)]
pub struct CompilationWarning {
    pub message: String,
    pub file: Option<String>,
    pub span: Option<std::ops::Range<usize>>,
}

/// Compile a single file.
pub async fn compile_file(
    path: &Path,
    output: &Path,
    config: CompilerConfig,
) -> miette::Result<CompilationResult> {
    let source =
        std::fs::read_to_string(path).map_err(|e| miette::miette!("Failed to read file: {}", e))?;

    compile_source(&source, Some(path), output, config).await
}

/// Compile source code.
pub async fn compile_source(
    source: &str,
    source_path: Option<&Path>,
    _output: &Path,
    config: CompilerConfig,
) -> miette::Result<CompilationResult> {
    let mut errors = Vec::new();
    let mut warnings = Vec::new();

    // Phase 1: Lexing + Parsing
    if config.verbose {
        tracing::info!("Parsing...");
    }

    let parse_result = haira_parser::parse(source);

    for err in &parse_result.errors {
        errors.push(CompilationError {
            message: err.to_string(),
            file: source_path.map(|p| p.display().to_string()),
            span: Some(err.span()),
        });
    }

    if !parse_result.errors.is_empty() {
        return Ok(CompilationResult {
            success: false,
            errors,
            warnings,
        });
    }

    // Phase 2: Name resolution
    if config.verbose {
        tracing::info!("Resolving names...");
    }

    let resolved = haira_resolver::resolve(&parse_result.ast);

    for err in &resolved.errors {
        errors.push(CompilationError {
            message: err.message.clone(),
            file: source_path.map(|p| p.display().to_string()),
            span: Some(err.span.clone()),
        });
    }

    // Phase 3: AI interpretation for unresolved calls
    if !resolved.unresolved_calls.is_empty() {
        if config.verbose {
            tracing::info!(
                "Interpreting {} unresolved function(s)...",
                resolved.unresolved_calls.len()
            );
        }

        let _engine = AIEngine::new(config.ai);

        // TODO: Interpret unresolved calls and generate implementations
        for call in &resolved.unresolved_calls {
            warnings.push(CompilationWarning {
                message: format!(
                    "Unresolved function '{}' - AI interpretation pending",
                    call.name
                ),
                file: source_path.map(|p| p.display().to_string()),
                span: Some(call.span.clone()),
            });
        }
    }

    // Phase 4-8: Type checking, lowering, codegen (TODO)
    if config.verbose {
        tracing::info!("Compilation pipeline incomplete - remaining phases pending");
    }

    Ok(CompilationResult {
        success: errors.is_empty(),
        errors,
        warnings,
    })
}

/// Check a source file without generating code.
pub fn check_file(path: &Path) -> miette::Result<CompilationResult> {
    let source =
        std::fs::read_to_string(path).map_err(|e| miette::miette!("Failed to read file: {}", e))?;

    check_source(&source, Some(path))
}

/// Check source code without generating code.
pub fn check_source(source: &str, source_path: Option<&Path>) -> miette::Result<CompilationResult> {
    let mut errors = Vec::new();
    let warnings = Vec::new();

    // Parse
    let parse_result = haira_parser::parse(source);

    for err in &parse_result.errors {
        errors.push(CompilationError {
            message: err.to_string(),
            file: source_path.map(|p| p.display().to_string()),
            span: Some(err.span()),
        });
    }

    // Resolve names
    let resolved = haira_resolver::resolve(&parse_result.ast);

    for err in &resolved.errors {
        errors.push(CompilationError {
            message: err.message.clone(),
            file: source_path.map(|p| p.display().to_string()),
            span: Some(err.span.clone()),
        });
    }

    Ok(CompilationResult {
        success: errors.is_empty(),
        errors,
        warnings,
    })
}
