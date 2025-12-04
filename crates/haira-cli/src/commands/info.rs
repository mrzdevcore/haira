//! Info command - show information about Haira installation.

pub(crate) fn run() -> miette::Result<()> {
    println!("Haira Programming Language");
    println!("===========================");
    println!();
    println!("Version: {}", env!("CARGO_PKG_VERSION"));
    println!();

    println!("Components:");
    println!("  haira-lexer    - Tokenization");
    println!("  haira-parser   - AST generation");
    println!("  haira-ast      - Abstract Syntax Tree definitions");
    println!("  haira-cir      - Canonical Intermediate Representation");
    println!("  haira-ai       - AI-powered intent interpretation");
    println!();

    println!("AI Integration:");
    println!("  Local AI: llama.cpp (recommended)");
    println!("  Ollama:   Optional fallback");
    println!("  Note: Use --local-ai or --ollama flags during build");
    println!();

    println!("Features:");
    println!("  - Intent-driven development");
    println!("  - Automatic function generation");
    println!("  - Full type inference");
    println!("  - Pattern-based auto-implementation");
    println!("  - AI-assisted code synthesis");
    println!();

    println!("Documentation: https://haira-lang.org");
    println!("Source: https://github.com/haira-lang/haira");

    Ok(())
}
