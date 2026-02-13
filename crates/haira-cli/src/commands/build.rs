//! Build command — compile a Haira file to a native binary via Go transpilation.

use haira_codegen_go::{compile_to_binary, find_runtime_path};
use haira_parser::parse;
use std::fs;
use std::path::Path;

pub(crate) fn run(file: &Path, output: Option<&Path>) -> miette::Result<()> {
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

    // Find Go runtime path
    let runtime_path = find_runtime_path().ok_or_else(|| {
        miette::miette!(
            "Could not find go-runtime directory.\n\
             Expected at: <project_root>/go-runtime/\n\
             Make sure you're running from the Haira project directory."
        )
    })?;

    // Determine output binary name
    let output_file = output.map(|p| p.to_path_buf()).unwrap_or_else(|| {
        let stem = file.file_stem().unwrap_or_default();
        let output_dir = Path::new(".output");
        if !output_dir.exists() {
            let _ = fs::create_dir_all(output_dir);
        }
        output_dir.join(stem)
    });

    // Compile via Go transpilation
    compile_to_binary(&result.ast, &output_file, &runtime_path)
        .map_err(|e| miette::miette!("{}", e))?;

    eprintln!("Built: {}", output_file.display());

    Ok(())
}
