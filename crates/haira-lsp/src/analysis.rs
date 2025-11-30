//! Code analysis for go-to-definition and find-references.

use haira_ast::{ItemKind, StatementKind};
use haira_parser::parse;
use tower_lsp::lsp_types::*;

/// Find the definition of the symbol at the given position.
pub fn find_definition(source: &str, position: Position, uri: Url) -> Option<Location> {
    let offset = position_to_offset(source, position);
    let word = get_word_at_offset(source, offset)?;

    // Parse the source
    let result = parse(source);

    // Search for definition
    for item in &result.ast.items {
        match &item.node {
            ItemKind::FunctionDef(func) => {
                if func.name.node.as_str() == word {
                    let range = span_to_range(
                        source,
                        func.name.span.start as usize,
                        func.name.span.end as usize,
                    );
                    return Some(Location { uri, range });
                }
            }
            ItemKind::TypeDef(type_def) => {
                if type_def.name.node.as_str() == word {
                    let range = span_to_range(
                        source,
                        type_def.name.span.start as usize,
                        type_def.name.span.end as usize,
                    );
                    return Some(Location { uri, range });
                }
                // Check fields
                for field in &type_def.fields {
                    if field.name.node.as_str() == word {
                        let range = span_to_range(
                            source,
                            field.name.span.start as usize,
                            field.name.span.end as usize,
                        );
                        return Some(Location { uri, range });
                    }
                }
            }
            ItemKind::MethodDef(method) => {
                if method.name.node.as_str() == word {
                    let range = span_to_range(
                        source,
                        method.name.span.start as usize,
                        method.name.span.end as usize,
                    );
                    return Some(Location { uri, range });
                }
            }
            ItemKind::Statement(stmt) => {
                if let StatementKind::Assignment(assign) = &stmt.node {
                    for target in &assign.targets {
                        if target.name.node.as_str() == word {
                            let range = span_to_range(
                                source,
                                target.name.span.start as usize,
                                target.name.span.end as usize,
                            );
                            return Some(Location { uri, range });
                        }
                    }
                }
            }
            _ => {}
        }
    }

    None
}

/// Find all references to the symbol at the given position.
pub fn find_references(source: &str, position: Position, uri: Url) -> Vec<Location> {
    let mut references = Vec::new();

    let offset = position_to_offset(source, position);
    let word = match get_word_at_offset(source, offset) {
        Some(w) => w,
        None => return references,
    };

    // Simple text-based search for now
    // A more sophisticated implementation would use the AST
    let mut line = 0;
    let mut col = 0;
    let mut in_word = false;
    let mut word_start = 0;
    let mut word_start_line = 0;
    let mut word_start_col = 0;

    for (i, c) in source.char_indices() {
        let is_word_char = c.is_alphanumeric() || c == '_';

        if is_word_char && !in_word {
            // Start of a word
            in_word = true;
            word_start = i;
            word_start_line = line;
            word_start_col = col;
        } else if !is_word_char && in_word {
            // End of a word
            in_word = false;
            let found_word = &source[word_start..i];
            if found_word == word {
                let range = Range::new(
                    Position::new(word_start_line, word_start_col),
                    Position::new(line, col),
                );
                references.push(Location {
                    uri: uri.clone(),
                    range,
                });
            }
        }

        if c == '\n' {
            line += 1;
            col = 0;
        } else {
            col += 1;
        }
    }

    // Handle word at end of file
    if in_word {
        let found_word = &source[word_start..];
        if found_word == word {
            let range = Range::new(
                Position::new(word_start_line, word_start_col),
                Position::new(line, col),
            );
            references.push(Location { uri, range });
        }
    }

    references
}

/// Convert an LSP position to a byte offset.
fn position_to_offset(source: &str, position: Position) -> usize {
    let mut current_line = 0;
    let mut line_start = 0;

    for (i, c) in source.char_indices() {
        if current_line == position.line {
            let col = (i - line_start) as u32;
            if col >= position.character {
                return i;
            }
        }
        if c == '\n' {
            if current_line == position.line {
                return i;
            }
            current_line += 1;
            line_start = i + 1;
        }
    }
    source.len()
}

/// Get the word at the given offset.
fn get_word_at_offset(source: &str, offset: usize) -> Option<String> {
    if offset >= source.len() {
        return None;
    }

    let bytes = source.as_bytes();

    // Find start of word
    let mut start = offset;
    while start > 0 {
        let c = bytes[start - 1] as char;
        if !c.is_alphanumeric() && c != '_' {
            break;
        }
        start -= 1;
    }

    // Find end of word
    let mut end = offset;
    while end < bytes.len() {
        let c = bytes[end] as char;
        if !c.is_alphanumeric() && c != '_' {
            break;
        }
        end += 1;
    }

    if start == end {
        return None;
    }

    Some(source[start..end].to_string())
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
