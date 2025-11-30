//! Build command - compile a Haira file to a native binary.

use haira_codegen::{compile_to_executable, CodegenOptions};
use haira_parser::parse;
use std::fs;
use std::path::Path;

pub fn run(file: &Path, output: Option<&Path>) -> miette::Result<()> {
    let source =
        fs::read_to_string(file).map_err(|e| miette::miette!("Failed to read file: {}", e))?;

    eprintln!("Compiling: {}", file.display());

    let result = parse(&source);

    // Report parse errors
    if !result.errors.is_empty() {
        for err in &result.errors {
            eprintln!("Parse error: {}", err);
        }
        return Err(miette::miette!("{} parse error(s)", result.errors.len()));
    }

    // Determine output binary name
    let output_file = output.map(|p| p.to_path_buf()).unwrap_or_else(|| {
        let stem = file.file_stem().unwrap_or_default();
        Path::new(stem).to_path_buf()
    });

    // Compile to native binary
    let options = CodegenOptions::default();
    compile_to_executable(&result.ast, &output_file, options)
        .map_err(|e| miette::miette!("Compilation error: {}", e))?;

    eprintln!("Built: {}", output_file.display());

    Ok(())
}
