//! Run command — compile and execute a Haira file via Go transpilation.

use haira_codegen_go::{find_runtime_path, run_program};
use haira_parser::parse;
use std::fs;
use std::path::Path;

pub(crate) fn run(file: &Path) -> miette::Result<()> {
    let source =
        fs::read_to_string(file).map_err(|e| miette::miette!("Failed to read file: {}", e))?;

    let result = parse(&source);

    // Report parse errors
    if !result.errors.is_empty() {
        for err in &result.errors {
            eprintln!("Parse error: {}", err);
        }
        return Err(miette::miette!("{} parse error(s)", result.errors.len()));
    }

    // Find Go runtime path
    let runtime_path = find_runtime_path().ok_or_else(|| {
        miette::miette!(
            "Could not find go-runtime directory.\n\
             Expected at: <project_root>/go-runtime/\n\
             Make sure you're running from the Haira project directory."
        )
    })?;

    // Run via Go transpilation
    let status = run_program(&result.ast, &runtime_path).map_err(|e| miette::miette!("{}", e))?;

    if !status.success() {
        if let Some(code) = status.code() {
            return Err(miette::miette!("Program exited with code {}", code));
        }
    }

    Ok(())
}
