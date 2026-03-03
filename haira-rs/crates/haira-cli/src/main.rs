//! Haira compiler CLI.
//!
//! Usage: haira-rs <command> [options] <file>

use std::env;
use std::process;

use haira_driver::{CompileMode, CompileOptions, CompileResult};

fn main() {
    let args: Vec<String> = env::args().collect();

    if args.len() < 2 {
        print_usage();
        process::exit(1);
    }

    let command = &args[1];

    match command.as_str() {
        "build" => run_command(CompileMode::Build, &args[2..]),
        "run" => run_command(CompileMode::Run, &args[2..]),
        "check" => run_command(CompileMode::Check, &args[2..]),
        "parse" => run_command(CompileMode::Parse, &args[2..]),
        "lex" => run_command(CompileMode::Lex, &args[2..]),
        "emit" => run_command(CompileMode::EmitHir, &args[2..]),
        "emit-llvm" => run_command(CompileMode::EmitLlvm, &args[2..]),
        "test" => run_command(CompileMode::Test, &args[2..]),
        "version" | "--version" | "-v" => {
            println!("haira-rs {}", env!("CARGO_PKG_VERSION"));
        }
        "help" | "--help" | "-h" => print_usage(),
        _ => {
            // If the first arg is a .haira file, treat as `run`
            if command.ends_with(".haira") {
                run_command(CompileMode::Run, &args[1..]);
            } else {
                eprintln!("unknown command: {command}");
                print_usage();
                process::exit(1);
            }
        }
    }
}

fn run_command(mode: CompileMode, args: &[String]) {
    if args.is_empty() {
        eprintln!("error: no input file specified");
        process::exit(1);
    }

    let input = &args[0];
    let optimize = args.contains(&"--release".to_string());

    let options = CompileOptions {
        input: input.clone(),
        mode,
        optimize,
    };

    match haira_driver::compile(&options) {
        Ok(result) => match result {
            CompileResult::Tokens(s) | CompileResult::Ast(s) | CompileResult::Hir(s) => {
                println!("{s}");
            }
            CompileResult::Checked => {
                println!("OK — no errors");
            }
            CompileResult::Built { output_path } => {
                println!("Built: {output_path}");
            }
            CompileResult::Executed { exit_code } => {
                process::exit(exit_code);
            }
            CompileResult::Tested {
                passed,
                failed,
                results,
            } => {
                println!("running {} tests\n", results.len());
                for (name, failure) in &results {
                    if let Some(msg) = failure {
                        println!("  FAIL  {name}");
                        println!("        {msg}");
                    } else {
                        println!("  PASS  {name}");
                    }
                }
                println!();
                if failed > 0 {
                    println!("{passed} passed, {failed} failed");
                    process::exit(1);
                } else {
                    println!("{passed} passed");
                }
            }
        },
        Err(diags) => {
            // Read source for pretty-printing
            if let Ok(source) = std::fs::read_to_string(input) {
                for d in &diags {
                    eprint!("{}", d.pretty_print(&source));
                }
            } else {
                for d in &diags {
                    eprintln!("{d}");
                }
            }
            process::exit(1);
        }
    }
}

fn print_usage() {
    println!("haira-rs compiler v{}", env!("CARGO_PKG_VERSION"));
    println!();
    println!("Usage: haira-rs <command> [options] <file>");
    println!();
    println!("Commands:");
    println!("  build <file>     Compile to native binary");
    println!("  run <file>       Compile and execute");
    println!("  check <file>     Type-check without compiling");
    println!("  parse <file>     Parse and dump AST");
    println!("  lex <file>       Tokenize and dump tokens");
    println!("  emit <file>      Emit HIR representation");
    println!("  test <file>      Discover and run tests");
    println!("  emit-llvm <file> Emit LLVM IR for debugging");
    println!("  version          Show version");
    println!("  help             Show this help");
    println!();
    println!("Options:");
    println!("  --release        Enable optimizations");
}
