//! Document symbols for Haira.

use haira_ast::{AssignPath, ItemKind, StatementKind};
use haira_parser::parse;
use tower_lsp::lsp_types::*;

/// Get document symbols from source code.
pub fn get_document_symbols(source: &str) -> Vec<SymbolInformation> {
    let mut symbols = Vec::new();

    // Parse the source
    let result = parse(source);

    // Extract symbols from AST
    for item in &result.ast.items {
        match &item.node {
            ItemKind::FunctionDef(func) => {
                let range = span_to_range(
                    source,
                    func.name.span.start as usize,
                    func.name.span.end as usize,
                );
                #[allow(deprecated)]
                symbols.push(SymbolInformation {
                    name: func.name.node.to_string(),
                    kind: SymbolKind::FUNCTION,
                    tags: None,
                    deprecated: None,
                    location: Location {
                        uri: Url::parse("file:///").unwrap(), // Will be fixed by caller
                        range,
                    },
                    container_name: None,
                });
            }
            ItemKind::TypeDef(type_def) => {
                let range = span_to_range(
                    source,
                    type_def.name.span.start as usize,
                    type_def.name.span.end as usize,
                );
                #[allow(deprecated)]
                symbols.push(SymbolInformation {
                    name: type_def.name.node.to_string(),
                    kind: SymbolKind::STRUCT,
                    tags: None,
                    deprecated: None,
                    location: Location {
                        uri: Url::parse("file:///").unwrap(),
                        range,
                    },
                    container_name: None,
                });

                // Add fields as children
                for field in &type_def.fields {
                    let field_range = span_to_range(
                        source,
                        field.name.span.start as usize,
                        field.name.span.end as usize,
                    );
                    #[allow(deprecated)]
                    symbols.push(SymbolInformation {
                        name: field.name.node.to_string(),
                        kind: SymbolKind::FIELD,
                        tags: None,
                        deprecated: None,
                        location: Location {
                            uri: Url::parse("file:///").unwrap(),
                            range: field_range,
                        },
                        container_name: Some(type_def.name.node.to_string()),
                    });
                }
            }
            ItemKind::MethodDef(method) => {
                let range = span_to_range(
                    source,
                    method.name.span.start as usize,
                    method.name.span.end as usize,
                );
                #[allow(deprecated)]
                symbols.push(SymbolInformation {
                    name: format!("{}.{}", method.type_name.node, method.name.node),
                    kind: SymbolKind::METHOD,
                    tags: None,
                    deprecated: None,
                    location: Location {
                        uri: Url::parse("file:///").unwrap(),
                        range,
                    },
                    container_name: Some(method.type_name.node.to_string()),
                });
            }
            ItemKind::TypeAlias(alias) => {
                let range = span_to_range(
                    source,
                    alias.name.span.start as usize,
                    alias.name.span.end as usize,
                );
                #[allow(deprecated)]
                symbols.push(SymbolInformation {
                    name: alias.name.node.to_string(),
                    kind: SymbolKind::TYPE_PARAMETER,
                    tags: None,
                    deprecated: None,
                    location: Location {
                        uri: Url::parse("file:///").unwrap(),
                        range,
                    },
                    container_name: None,
                });
            }
            ItemKind::AiFunctionDef(ai_block) => {
                // Only add named AI functions
                if let Some(name) = &ai_block.name {
                    let range =
                        span_to_range(source, name.span.start as usize, name.span.end as usize);
                    #[allow(deprecated)]
                    symbols.push(SymbolInformation {
                        name: name.node.to_string(),
                        kind: SymbolKind::FUNCTION,
                        tags: None,
                        deprecated: None,
                        location: Location {
                            uri: Url::parse("file:///").unwrap(),
                            range,
                        },
                        container_name: None,
                    });
                }
            }
            ItemKind::Statement(stmt) => {
                // Check for top-level assignments (global variables)
                if let StatementKind::Assignment(assign) = &stmt.node {
                    for target in &assign.targets {
                        // Only extract simple identifier assignments as symbols
                        if let AssignPath::Identifier(name) = &target.path {
                            let range = span_to_range(
                                source,
                                name.span.start as usize,
                                name.span.end as usize,
                            );
                            #[allow(deprecated)]
                            symbols.push(SymbolInformation {
                                name: name.node.to_string(),
                                kind: SymbolKind::VARIABLE,
                                tags: None,
                                deprecated: None,
                                location: Location {
                                    uri: Url::parse("file:///").unwrap(),
                                    range,
                                },
                                container_name: None,
                            });
                        }
                    }
                }
            }
        }
    }

    symbols
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
