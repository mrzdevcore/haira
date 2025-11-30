//! Lex command - tokenize a file.

use haira_lexer::{Lexer, TokenKind};
use std::fs;
use std::path::Path;

pub fn run(file: &Path) -> miette::Result<()> {
    let source = fs::read_to_string(file)
        .map_err(|e| miette::miette!("Failed to read file: {}", e))?;

    println!("Tokenizing: {}\n", file.display());

    let lexer = Lexer::new(&source);
    let mut token_count = 0;
    let mut error_count = 0;

    for result in lexer {
        match result {
            Ok(token) => {
                let text = &source[token.span.clone()];
                let text_display = if text.len() > 40 {
                    format!("{}...", &text[..40])
                } else {
                    text.to_string()
                };

                println!(
                    "{:4}..{:4}  {:20}  {:?}",
                    token.span.start,
                    token.span.end,
                    format!("{:?}", token.kind).chars().take(20).collect::<String>(),
                    text_display.replace('\n', "\\n")
                );
                token_count += 1;
            }
            Err(err) => {
                println!("ERROR at {:?}: {}", err.span(), err);
                error_count += 1;
            }
        }
    }

    println!("\n{} tokens, {} errors", token_count, error_count);

    if error_count > 0 {
        Err(miette::miette!("{} lexer errors", error_count))
    } else {
        Ok(())
    }
}
