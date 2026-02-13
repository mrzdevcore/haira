//! Serve command - parse a .haira file and run it via the Bun runtime.

use haira_ast::ItemKind;
use haira_parser::parse;
use std::fs;
use std::path::Path;
use std::process::Command;

pub(crate) fn run(file: &Path) -> miette::Result<()> {
    let source =
        fs::read_to_string(file).map_err(|e| miette::miette!("Failed to read file: {}", e))?;

    eprintln!("[haira] Parsing {}...", file.display());

    let result = parse(&source);

    // Report errors
    if !result.errors.is_empty() {
        for err in &result.errors {
            eprintln!("[haira] Error: {}", err);
        }
        return Err(miette::miette!(
            "{} parse errors in {}",
            result.errors.len(),
            file.display()
        ));
    }

    // Count constructs
    let mut providers = 0;
    let mut tools = 0;
    let mut agents = 0;
    let mut workflows = 0;
    let mut functions = 0;
    let mut imports = 0;

    let mut agent_name = String::new();
    let mut workflow_name = String::new();

    for item in &result.ast.items {
        match &item.node {
            ItemKind::ImportDecl(_) => imports += 1,
            ItemKind::ProviderDecl(_) => providers += 1,
            ItemKind::ToolDecl(_) => tools += 1,
            ItemKind::AgentDecl(a) => {
                agents += 1;
                agent_name = a.name.node.to_string();
            }
            ItemKind::WorkflowDecl(w) => {
                workflows += 1;
                workflow_name = w.name.node.to_string();
            }
            ItemKind::FunctionDef(_) => functions += 1,
            _ => {}
        }
    }

    eprintln!(
        "[haira] Found: {} import(s), {} provider(s), {} tool(s), {} agent(s), {} workflow(s), {} function(s)",
        imports, providers, tools, agents, workflows, functions
    );

    if agents == 0 {
        return Err(miette::miette!(
            "No agent declaration found in {}",
            file.display()
        ));
    }

    eprintln!(
        "[haira] Starting {} via {} on :3000",
        agent_name, workflow_name
    );
    eprintln!("[haira] Delegating to Bun runtime...");
    eprintln!();

    // Resolve the poc directory relative to the .haira file
    let haira_dir = file
        .parent()
        .ok_or_else(|| miette::miette!("Cannot determine parent directory"))?;

    // Run the Bun runtime
    let status = Command::new("bun")
        .arg("run")
        .arg("src/index.ts")
        .current_dir(haira_dir)
        .env(
            "HAIRA_SOURCE_FILE",
            file.canonicalize().unwrap_or_else(|_| file.to_path_buf()),
        )
        .status()
        .map_err(|e| miette::miette!("Failed to start Bun runtime: {}. Is bun installed?", e))?;

    if !status.success() {
        return Err(miette::miette!("Bun runtime exited with error"));
    }

    Ok(())
}
