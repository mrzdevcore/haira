//! Haira driver — compilation pipeline orchestrator.
//!
//! Ties together all compiler passes: lex → parse → check → lower → codegen.

use haira_errors::{Diagnostic, Span};
use haira_token::TokenKind;

// ===========================================================================
// Public types
// ===========================================================================

/// Pipeline options.
#[derive(Debug, Clone)]
pub struct CompileOptions {
    /// The main source file path.
    pub input: String,
    /// Output mode.
    pub mode: CompileMode,
    /// Whether to optimise (--release).
    pub optimize: bool,
}

/// What the driver should produce.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum CompileMode {
    /// Tokenize and dump tokens.
    Lex,
    /// Parse and dump AST.
    Parse,
    /// Resolve + type-check, report diagnostics.
    Check,
    /// Compile to native binary (via LLVM).
    Build,
    /// Compile and run immediately (interpreter).
    Run,
    /// Emit HIR for debugging.
    EmitHir,
    /// Emit LLVM IR for debugging.
    EmitLlvm,
    /// Discover and run tests.
    Test,
}

/// Pipeline result.
#[derive(Debug)]
pub enum CompileResult {
    /// Token dump (for Lex mode).
    Tokens(String),
    /// AST dump (for Parse mode).
    Ast(String),
    /// Check succeeded (warnings may have been printed).
    Checked,
    /// Build produced a binary at the given path.
    Built { output_path: String },
    /// Run completed with exit code.
    Executed { exit_code: i32 },
    /// HIR dump.
    Hir(String),
    /// Test run results.
    Tested {
        passed: usize,
        failed: usize,
        results: Vec<(String, Option<String>)>,
    },
}

// ===========================================================================
// Compile entry point
// ===========================================================================

/// Run the compilation pipeline.
pub fn compile(options: &CompileOptions) -> Result<CompileResult, Vec<Diagnostic>> {
    // Step 1: Read source file
    let source = std::fs::read_to_string(&options.input).map_err(|e| {
        vec![Diagnostic::error(
            format!("cannot read {}: {}", options.input, e),
            Span::default(),
        )]
    })?;

    // Step 2: Lex
    if options.mode == CompileMode::Lex {
        let lexer = haira_lexer::Lexer::new(&source);
        let dump = lexer
            .all_tokens()
            .iter()
            .filter(|t| t.kind != TokenKind::Newline && !t.kind.is_trivia())
            .map(|t| format!("{t}"))
            .collect::<Vec<_>>()
            .join("\n");
        return Ok(CompileResult::Tokens(dump));
    }

    // Step 3: Parse
    let (ast, parse_diags) = haira_parser::parse(&source);

    if haira_errors::has_errors(&parse_diags) {
        // Attach file path to parse diagnostics
        let diags = parse_diags
            .into_iter()
            .map(|d| {
                if d.file.is_empty() {
                    d.with_file(&options.input)
                } else {
                    d
                }
            })
            .collect();
        return Err(diags);
    }

    if options.mode == CompileMode::Parse {
        return Ok(CompileResult::Ast(format!("{ast:#?}")));
    }

    // Step 4: Check
    let (_type_info, check_diags) = haira_checker::check(&ast);
    let mut all_diags: Vec<Diagnostic> = parse_diags;
    all_diags.extend(check_diags);

    if haira_errors::has_errors(&all_diags) {
        let diags = all_diags
            .into_iter()
            .map(|d| {
                if d.file.is_empty() {
                    d.with_file(&options.input)
                } else {
                    d
                }
            })
            .collect();
        return Err(diags);
    }

    if options.mode == CompileMode::Check {
        // Print warnings if any
        if !all_diags.is_empty() {
            let output = haira_errors::format_all(&all_diags, &source);
            eprint!("{output}");
        }
        return Ok(CompileResult::Checked);
    }

    // Step 5: Lower to HIR
    let (hir_module, lower_diags) = haira_ir::lower(&ast);
    all_diags.extend(lower_diags);

    if haira_errors::has_errors(&all_diags) {
        return Err(all_diags);
    }

    if options.mode == CompileMode::EmitHir {
        return Ok(CompileResult::Hir(format!("{hir_module:#?}")));
    }

    // Step 6: Test execution
    if options.mode == CompileMode::Test {
        let test_results = haira_codegen_cranelift::run_tests(&hir_module);
        let mut passed = 0;
        let mut failed = 0;
        let mut results = Vec::new();
        for tr in test_results {
            if tr.failure.is_none() {
                passed += 1;
            } else {
                failed += 1;
            }
            results.push((tr.name, tr.failure));
        }
        return Ok(CompileResult::Tested {
            passed,
            failed,
            results,
        });
    }

    // Step 7: Codegen / Execute
    match options.mode {
        CompileMode::Run => {
            let codegen_options = haira_codegen_cranelift::CodegenOptions {
                target: haira_codegen_cranelift::Target::Interpreter,
                optimize: options.optimize,
            };
            let result = haira_codegen_cranelift::codegen(&hir_module, &codegen_options)
                .map_err(|mut diags| {
                    all_diags.append(&mut diags);
                    all_diags.clone()
                })?;
            match result {
                haira_codegen_cranelift::CodegenResult::Executed { exit_code } => {
                    Ok(CompileResult::Executed { exit_code })
                }
                haira_codegen_cranelift::CodegenResult::ObjectFile(_) => {
                    Ok(CompileResult::Executed { exit_code: 0 })
                }
            }
        }
        CompileMode::Build | CompileMode::EmitLlvm => {
            let output_path = options.input.replace(".haira", "");
            let llvm_options = haira_codegen_llvm::CodegenOptions {
                optimize: options.optimize,
                output: output_path.clone(),
                emit_ir: options.mode == CompileMode::EmitLlvm,
            };
            haira_codegen_llvm::codegen(&hir_module, &llvm_options).map_err(|mut diags| {
                all_diags.append(&mut diags);
                all_diags.clone()
            })?;
            Ok(CompileResult::Built { output_path })
        }
        _ => unreachable!(),
    }
}

// ===========================================================================
// Tests
// ===========================================================================

#[cfg(test)]
mod tests {
    use super::*;
    use std::io::Write;

    fn write_temp_file(content: &str) -> String {
        use std::sync::atomic::{AtomicU64, Ordering};
        static COUNTER: AtomicU64 = AtomicU64::new(0);
        let dir = std::env::temp_dir();
        let id = COUNTER.fetch_add(1, Ordering::Relaxed);
        let path = dir.join(format!("haira_test_{}_{}.haira", std::process::id(), id));
        let mut f = std::fs::File::create(&path).unwrap();
        f.write_all(content.as_bytes()).unwrap();
        path.to_string_lossy().to_string()
    }

    fn cleanup(path: &str) {
        let _ = std::fs::remove_file(path);
    }

    #[test]
    fn lex_mode_produces_tokens() {
        let path = write_temp_file("let x = 42\n");
        let result = compile(&CompileOptions {
            input: path.clone(),
            mode: CompileMode::Lex,
            optimize: false,
        })
        .unwrap();
        cleanup(&path);

        match result {
            CompileResult::Tokens(s) => {
                assert!(s.contains("let"));
                assert!(s.contains("Ident"));
                assert!(s.contains("42"));
            }
            other => panic!("expected Tokens, got {other:?}"),
        }
    }

    #[test]
    fn parse_mode_produces_ast() {
        let path = write_temp_file("let x = 42\n");
        let result = compile(&CompileOptions {
            input: path.clone(),
            mode: CompileMode::Parse,
            optimize: false,
        })
        .unwrap();
        cleanup(&path);

        match result {
            CompileResult::Ast(s) => {
                assert!(s.contains("SourceFile"));
            }
            other => panic!("expected Ast, got {other:?}"),
        }
    }

    #[test]
    fn check_mode_succeeds() {
        let path = write_temp_file("let x = 42\n");
        let result = compile(&CompileOptions {
            input: path.clone(),
            mode: CompileMode::Check,
            optimize: false,
        })
        .unwrap();
        cleanup(&path);

        assert!(matches!(result, CompileResult::Checked));
    }

    #[test]
    fn emit_hir_mode() {
        let path = write_temp_file("let x = 42\n");
        let result = compile(&CompileOptions {
            input: path.clone(),
            mode: CompileMode::EmitHir,
            optimize: false,
        })
        .unwrap();
        cleanup(&path);

        match result {
            CompileResult::Hir(s) => {
                assert!(s.contains("HirModule"));
            }
            other => panic!("expected Hir, got {other:?}"),
        }
    }

    #[test]
    fn run_mode_simple() {
        let path = write_temp_file("let x = 42\n");
        let result = compile(&CompileOptions {
            input: path.clone(),
            mode: CompileMode::Run,
            optimize: false,
        })
        .unwrap();
        cleanup(&path);

        match result {
            CompileResult::Executed { exit_code } => {
                assert_eq!(exit_code, 0);
            }
            other => panic!("expected Executed, got {other:?}"),
        }
    }

    #[test]
    fn missing_file_error() {
        let result = compile(&CompileOptions {
            input: "/nonexistent/file.haira".to_string(),
            mode: CompileMode::Run,
            optimize: false,
        });
        assert!(result.is_err());
        let diags = result.unwrap_err();
        assert!(diags[0].message.contains("cannot read"));
    }
}
