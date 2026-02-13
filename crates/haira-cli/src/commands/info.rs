//! Info command - show information about Haira installation.

pub(crate) fn run() -> miette::Result<()> {
    println!("Haira Programming Language");
    println!("===========================");
    println!();
    println!("Version: {}", env!("CARGO_PKG_VERSION"));
    println!();

    println!("Components:");
    println!("  haira-lexer      - Tokenization");
    println!("  haira-parser     - AST generation");
    println!("  haira-ast        - Abstract Syntax Tree definitions");
    println!("  haira-codegen-go - Go transpilation backend");
    println!();

    println!("Features:");
    println!("  - Provider, tool, agent, workflow primitives");
    println!("  - Agent orchestration (handoffs, parallel, routing)");
    println!("  - Go transpilation with native binary output");
    println!("  - Full type inference");
    println!("  - Pattern matching");
    println!("  - String interpolation");
    println!();

    println!("Documentation: https://haira-lang.org");
    println!("Source: https://github.com/haira-lang/haira");

    Ok(())
}
