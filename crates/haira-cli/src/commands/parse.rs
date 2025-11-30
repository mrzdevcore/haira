//! Parse command - parse a file and show AST.

use haira_parser::parse;
use std::fs;
use std::path::Path;

pub fn run(file: &Path, json: bool) -> miette::Result<()> {
    let source = fs::read_to_string(file)
        .map_err(|e| miette::miette!("Failed to read file: {}", e))?;

    println!("Parsing: {}\n", file.display());

    let result = parse(&source);

    // Report errors
    if !result.errors.is_empty() {
        println!("Errors:");
        for err in &result.errors {
            let span = err.span();
            let (line, col) = offset_to_line_col(&source, span.start);
            println!("  {}:{}:{}: {}", file.display(), line, col, err);
        }
        println!();
    }

    // Show AST
    if json {
        // Would need serde feature on AST for this
        println!("JSON output not yet implemented");
    } else {
        println!("AST:");
        print_ast(&result.ast, &source);
    }

    println!(
        "\n{} items, {} errors",
        result.ast.items.len(),
        result.errors.len()
    );

    if !result.errors.is_empty() {
        Err(miette::miette!("{} parse errors", result.errors.len()))
    } else {
        Ok(())
    }
}

fn print_ast(ast: &haira_ast::SourceFile, source: &str) {
    for item in &ast.items {
        print_item(item, source, 0);
    }
}

fn print_item(item: &haira_ast::Item, source: &str, indent: usize) {
    let prefix = "  ".repeat(indent);

    match &item.node {
        haira_ast::ItemKind::TypeDef(def) => {
            println!(
                "{}TypeDef: {} ({} fields)",
                prefix,
                def.name.node,
                def.fields.len()
            );
            for field in &def.fields {
                println!("{}  - {}", prefix, field.name.node);
            }
        }
        haira_ast::ItemKind::FunctionDef(def) => {
            println!(
                "{}FunctionDef: {} ({} params)",
                prefix,
                def.name.node,
                def.params.len()
            );
            for param in &def.params {
                println!("{}  param: {}", prefix, param.name.node);
            }
            println!("{}  body: {} statements", prefix, def.body.statements.len());
        }
        haira_ast::ItemKind::MethodDef(def) => {
            println!(
                "{}MethodDef: {}.{} ({} params)",
                prefix, def.type_name.node, def.name.node, def.params.len()
            );
        }
        haira_ast::ItemKind::TypeAlias(alias) => {
            println!("{}TypeAlias: {}", prefix, alias.name.node);
        }
        haira_ast::ItemKind::Statement(stmt) => {
            print_statement_kind(stmt, source, indent);
        }
    }
}

fn print_statement_kind(stmt: &haira_ast::StatementKind, source: &str, indent: usize) {
    let prefix = "  ".repeat(indent);

    match stmt {
        haira_ast::StatementKind::Assignment(assign) => {
            let targets: Vec<_> = assign.targets.iter().map(|t| t.name.node.as_str()).collect();
            println!("{}Assignment: {} = ...", prefix, targets.join(", "));
        }
        haira_ast::StatementKind::If(_) => {
            println!("{}If statement", prefix);
        }
        haira_ast::StatementKind::For(for_stmt) => {
            let pattern = match &for_stmt.pattern {
                haira_ast::ForPattern::Single(name) => name.node.to_string(),
                haira_ast::ForPattern::Pair(a, b) => format!("{}, {}", a.node, b.node),
            };
            println!("{}For: {} in ...", prefix, pattern);
        }
        haira_ast::StatementKind::While(_) => {
            println!("{}While statement", prefix);
        }
        haira_ast::StatementKind::Match(_) => {
            println!("{}Match statement", prefix);
        }
        haira_ast::StatementKind::Return(ret) => {
            println!("{}Return ({} values)", prefix, ret.values.len());
        }
        haira_ast::StatementKind::Try(_) => {
            println!("{}Try-catch statement", prefix);
        }
        haira_ast::StatementKind::Break => {
            println!("{}Break", prefix);
        }
        haira_ast::StatementKind::Continue => {
            println!("{}Continue", prefix);
        }
        haira_ast::StatementKind::Expr(expr) => {
            print_expr_kind(&expr.node, source, indent);
        }
    }
}

fn print_expr_kind(expr: &haira_ast::ExprKind, _source: &str, indent: usize) {
    let prefix = "  ".repeat(indent);

    match expr {
        haira_ast::ExprKind::Literal(lit) => {
            let lit_str = match lit {
                haira_ast::Literal::Int(n) => format!("Int({})", n),
                haira_ast::Literal::Float(n) => format!("Float({})", n),
                haira_ast::Literal::String(s) => format!("String({:?})", s),
                haira_ast::Literal::Bool(b) => format!("Bool({})", b),
                haira_ast::Literal::InterpolatedString(_) => "InterpolatedString".to_string(),
            };
            println!("{}Literal: {}", prefix, lit_str);
        }
        haira_ast::ExprKind::Identifier(name) => {
            println!("{}Identifier: {}", prefix, name);
        }
        haira_ast::ExprKind::Call(call) => {
            println!("{}Call ({} args)", prefix, call.args.len());
        }
        haira_ast::ExprKind::Pipe(_) => {
            println!("{}Pipe expression", prefix);
        }
        haira_ast::ExprKind::Lambda(lambda) => {
            println!("{}Lambda ({} params)", prefix, lambda.params.len());
        }
        haira_ast::ExprKind::Instance(inst) => {
            println!(
                "{}Instance: {} ({} fields)",
                prefix,
                inst.type_name.node,
                inst.fields.len()
            );
        }
        _ => {
            println!("{}Expression", prefix);
        }
    }
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
