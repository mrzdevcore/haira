//! Hover information for Haira.

use tower_lsp::lsp_types::*;

/// Get hover information at the given position.
pub fn get_hover(source: &str, position: Position) -> Option<Hover> {
    let offset = position_to_offset(source, position);
    let word = get_word_at_offset(source, offset)?;

    // Check for keywords
    let keyword_info = match word.as_str() {
        "if" => Some(("keyword", "Conditional expression\n\n```haira\nif condition {\n    // then branch\n} else {\n    // else branch\n}\n```")),
        "else" => Some(("keyword", "Alternative branch of an if expression")),
        "for" => Some(("keyword", "For loop\n\n```haira\nfor item in collection {\n    // loop body\n}\n```")),
        "while" => Some(("keyword", "While loop\n\n```haira\nwhile condition {\n    // loop body\n}\n```")),
        "in" => Some(("keyword", "Used in for loops to iterate over a collection")),
        "return" => Some(("keyword", "Return a value from a function")),
        "match" => Some(("keyword", "Pattern matching expression\n\n```haira\nmatch value {\n    pattern => result\n    _ => default\n}\n```")),
        "try" => Some(("keyword", "Error handling block\n\n```haira\ntry {\n    // code that might fail\n} catch e {\n    // handle error\n}\n```")),
        "catch" => Some(("keyword", "Error handler in a try block")),
        "break" => Some(("keyword", "Exit from a loop")),
        "continue" => Some(("keyword", "Skip to the next iteration of a loop")),
        "spawn" => Some(("keyword", "Spawn a concurrent task (fire-and-forget)\n\n```haira\nspawn {\n    // runs in background\n}\n```")),
        "async" => Some(("keyword", "Run statements concurrently and wait for all\n\n```haira\nasync {\n    task1()\n    task2()  // runs in parallel with task1\n}\n// continues after both complete\n```")),
        "true" => Some(("constant", "Boolean true value")),
        "false" => Some(("constant", "Boolean false value")),
        "none" => Some(("constant", "Represents absence of a value (Option type)")),
        "some" => Some(("function", "Wraps a value in an Option\n\n```haira\nsome(42)  // Option containing 42\n```")),
        "public" => Some(("keyword", "Makes a function or type publicly accessible")),
        "and" => Some(("operator", "Logical AND operator")),
        "or" => Some(("operator", "Logical OR operator")),
        "not" => Some(("operator", "Logical NOT operator")),
        _ => None,
    };

    if let Some((kind, doc)) = keyword_info {
        return Some(Hover {
            contents: HoverContents::Markup(MarkupContent {
                kind: MarkupKind::Markdown,
                value: format!("**{}** _{}_\n\n{}", word, kind, doc),
            }),
            range: None,
        });
    }

    // Check for built-in functions
    let builtin_info = match word.as_str() {
        "print" => Some("```haira\nprint(value: any)\n```\n\nPrint a value to standard output followed by a newline."),
        "println" => Some("```haira\nprintln()\n```\n\nPrint a newline to standard output."),
        "sleep" => Some("```haira\nsleep(ms: int)\n```\n\nSleep for the specified number of milliseconds."),
        "channel" => Some("```haira\nchannel(capacity: int = 1) -> Channel\n```\n\nCreate a new channel with the specified buffer capacity."),
        "channel_send" => Some("```haira\nchannel_send(ch: Channel, value: any)\n```\n\nSend a value to a channel. Blocks if the channel is full."),
        "channel_receive" => Some("```haira\nchannel_receive(ch: Channel) -> any\n```\n\nReceive a value from a channel. Blocks if the channel is empty."),
        "channel_close" => Some("```haira\nchannel_close(ch: Channel)\n```\n\nClose a channel, signaling no more values will be sent."),
        "spawn_fn" => Some("```haira\nspawn_fn(func: () -> any) -> ThreadHandle\n```\n\nSpawn a function in a new thread."),
        "err" => Some("```haira\nerr(value: any = 1)\n```\n\nSet an error value. Can be caught with try/catch or propagated with `?`."),
        _ => None,
    };

    if let Some(doc) = builtin_info {
        return Some(Hover {
            contents: HoverContents::Markup(MarkupContent {
                kind: MarkupKind::Markdown,
                value: format!("**{}** _built-in function_\n\n{}", word, doc),
            }),
            range: None,
        });
    }

    // Check for types
    let type_info = match word.as_str() {
        "int" => Some("64-bit signed integer type"),
        "float" => Some("64-bit floating point type"),
        "string" => Some("UTF-8 string type"),
        "bool" => Some("Boolean type (true or false)"),
        "List" => Some("Generic list/array type\n\n```haira\nList<int>  // list of integers\n```"),
        "Map" => Some("Generic map/dictionary type\n\n```haira\nMap<string, int>  // string to int mapping\n```"),
        "Option" => Some("Optional value type\n\n```haira\nOption<int>  // either some(value) or none\n```"),
        "Result" => Some("Result type for error handling\n\n```haira\nResult<int, string>  // either ok(value) or err(error)\n```"),
        _ => None,
    };

    if let Some(doc) = type_info {
        return Some(Hover {
            contents: HoverContents::Markup(MarkupContent {
                kind: MarkupKind::Markdown,
                value: format!("**{}** _type_\n\n{}", word, doc),
            }),
            range: None,
        });
    }

    None
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
