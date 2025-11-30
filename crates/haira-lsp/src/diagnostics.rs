//! Diagnostics collection for Haira.

use haira_lexer::Lexer;
use haira_parser::parse;
use tower_lsp::lsp_types::*;

/// Collect diagnostics from source code.
pub fn collect_diagnostics(source: &str) -> Vec<Diagnostic> {
    let mut diagnostics = Vec::new();

    // Lex the source and collect errors
    let lexer = Lexer::new(source);
    for result in lexer {
        if let Err(error) = result {
            let range = span_to_range(source, error.span().start, error.span().end);
            diagnostics.push(Diagnostic {
                range,
                severity: Some(DiagnosticSeverity::ERROR),
                code: None,
                code_description: None,
                source: Some("haira".to_string()),
                message: format!("Lexer error: {}", error),
                related_information: None,
                tags: None,
                data: None,
            });
        }
    }

    // Parse the source and collect errors
    let result = parse(source);
    for error in result.errors {
        let span = error.span();
        let range = span_to_range(source, span.start, span.end);
        diagnostics.push(Diagnostic {
            range,
            severity: Some(DiagnosticSeverity::ERROR),
            code: None,
            code_description: None,
            source: Some("haira".to_string()),
            message: error.to_string(),
            related_information: None,
            tags: None,
            data: None,
        });
    }

    diagnostics
}

/// Convert byte offsets to an LSP range.
fn span_to_range(source: &str, start: usize, end: usize) -> Range {
    let start_pos = offset_to_position(source, start);
    let end_pos = offset_to_position(source, end);
    Range::new(start_pos, end_pos)
}

/// Convert a byte offset to an LSP position.
fn offset_to_position(source: &str, offset: usize) -> Position {
    let mut line = 0;
    let mut col = 0;
    for (i, c) in source.char_indices() {
        if i >= offset {
            break;
        }
        if c == '\n' {
            line += 1;
            col = 0;
        } else {
            col += 1;
        }
    }
    Position::new(line, col)
}
