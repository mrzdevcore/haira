//! Check command - check files for errors without full compilation.

use haira_parser::parse;
use std::fs;
use std::path::Path;

pub fn run(files: &[std::path::PathBuf]) -> miette::Result<()> {
    if files.is_empty() {
        return Err(miette::miette!("No files specified"));
    }

    let mut total_errors = 0;
    let mut total_warnings = 0;

    for file in files {
        let (errors, warnings) = check_file(file)?;
        total_errors += errors;
        total_warnings += warnings;
    }

    println!();
    if total_errors > 0 {
        println!(
            "Check complete: {} error(s), {} warning(s)",
            total_errors, total_warnings
        );
        Err(miette::miette!("{} errors found", total_errors))
    } else if total_warnings > 0 {
        println!("Check complete: {} warning(s)", total_warnings);
        Ok(())
    } else {
        println!("Check complete: no issues found");
        Ok(())
    }
}

fn check_file(file: &Path) -> miette::Result<(usize, usize)> {
    let source = fs::read_to_string(file)
        .map_err(|e| miette::miette!("Failed to read {}: {}", file.display(), e))?;

    println!("Checking: {}", file.display());

    let result = parse(&source);

    let mut errors = 0;
    let warnings = 0;

    // Report parse errors
    for err in &result.errors {
        let span = err.span();
        let (line, col) = offset_to_line_col(&source, span.start as usize);
        println!("  error[E0001]: {}", err);
        println!("   --> {}:{}:{}", file.display(), line, col);
        println!("    |");
        print_source_line(&source, line);
        println!("    |");
        errors += 1;
    }

    // Basic semantic checks could be added here
    // For now, we just do parsing validation

    if errors == 0 {
        println!("  ok");
    }

    Ok((errors, warnings))
}

fn offset_to_line_col(source: &str, offset: usize) -> (usize, usize) {
    let mut line = 1;
    let mut col = 1;

    for (i, c) in source.char_indices() {
        if i >= offset {
            break;
        }
        if c == '\n' {
            line += 1;
            col = 1;
        } else {
            col += 1;
        }
    }

    (line, col)
}

fn print_source_line(source: &str, line_num: usize) {
    if let Some(line) = source.lines().nth(line_num - 1) {
        println!("{:4} | {}", line_num, line);
    }
}
