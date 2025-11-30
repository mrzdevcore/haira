//! Run command - compile and execute a Haira file.

use haira_codegen::{compile_to_executable, CodegenOptions};
use haira_parser::parse;
use std::fs;
use std::path::Path;
use std::process::Command;

pub fn run(file: &Path) -> miette::Result<()> {
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

    // Create temporary output path
    let tmp_dir = std::env::temp_dir();
    let output_file = tmp_dir.join("haira_run_temp");

    // Compile to native binary
    let options = CodegenOptions::default();
    compile_to_executable(&result.ast, &output_file, options)
        .map_err(|e| miette::miette!("Compilation error: {}", e))?;

    // Execute the binary
    let status = Command::new(&output_file)
        .status()
        .map_err(|e| miette::miette!("Failed to execute: {}", e))?;

    // Clean up
    fs::remove_file(&output_file).ok();

    if !status.success() {
        if let Some(code) = status.code() {
            return Err(miette::miette!("Program exited with code {}", code));
        }
    }

    Ok(())
}
