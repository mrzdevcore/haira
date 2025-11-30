//! Code completion for Haira.

use tower_lsp::lsp_types::*;

/// Keywords in Haira.
const KEYWORDS: &[(&str, &str)] = &[
    ("if", "Conditional expression"),
    ("else", "Alternative branch"),
    ("for", "For loop"),
    ("while", "While loop"),
    ("in", "Iterator keyword"),
    ("return", "Return from function"),
    ("match", "Pattern matching"),
    ("try", "Error handling block"),
    ("catch", "Error handler"),
    ("break", "Break from loop"),
    ("continue", "Continue to next iteration"),
    ("spawn", "Spawn concurrent task"),
    ("async", "Async block"),
    ("true", "Boolean true"),
    ("false", "Boolean false"),
    ("none", "Option none value"),
    ("some", "Option some value"),
    ("public", "Public visibility modifier"),
    ("and", "Logical and"),
    ("or", "Logical or"),
    ("not", "Logical not"),
];

/// Built-in functions.
const BUILTINS: &[(&str, &str, &str)] = &[
    ("print", "print(value)", "Print a value to stdout"),
    ("println", "println()", "Print a newline"),
    ("sleep", "sleep(ms)", "Sleep for milliseconds"),
    ("channel", "channel(capacity)", "Create a new channel"),
    (
        "channel_send",
        "channel_send(ch, value)",
        "Send value to channel",
    ),
    (
        "channel_receive",
        "channel_receive(ch)",
        "Receive value from channel",
    ),
    ("channel_close", "channel_close(ch)", "Close a channel"),
    ("spawn_fn", "spawn_fn(func)", "Spawn function in new thread"),
    ("err", "err(value)", "Create an error"),
];

/// Get completions at the given position.
pub fn get_completions(source: &str, position: Position) -> Vec<CompletionItem> {
    let mut completions = Vec::new();

    // Get the text before the cursor to determine context
    let offset = position_to_offset(source, position);
    let prefix = get_word_prefix(source, offset);

    // Add keyword completions
    for (keyword, doc) in KEYWORDS {
        if keyword.starts_with(&prefix) || prefix.is_empty() {
            completions.push(CompletionItem {
                label: keyword.to_string(),
                kind: Some(CompletionItemKind::KEYWORD),
                detail: Some(doc.to_string()),
                insert_text: Some(keyword.to_string()),
                ..Default::default()
            });
        }
    }

    // Add built-in function completions
    for (name, signature, doc) in BUILTINS {
        if name.starts_with(&prefix) || prefix.is_empty() {
            completions.push(CompletionItem {
                label: name.to_string(),
                kind: Some(CompletionItemKind::FUNCTION),
                detail: Some(signature.to_string()),
                documentation: Some(Documentation::String(doc.to_string())),
                insert_text: Some(format!("{}($0)", name)),
                insert_text_format: Some(InsertTextFormat::SNIPPET),
                ..Default::default()
            });
        }
    }

    // Add type completions
    let types = &[
        ("int", "Integer type"),
        ("float", "Floating point type"),
        ("string", "String type"),
        ("bool", "Boolean type"),
        ("List", "List collection type"),
        ("Map", "Map collection type"),
        ("Option", "Optional value type"),
        ("Result", "Result type for error handling"),
    ];

    for (type_name, doc) in types {
        if type_name.starts_with(&prefix) || prefix.is_empty() {
            completions.push(CompletionItem {
                label: type_name.to_string(),
                kind: Some(CompletionItemKind::TYPE_PARAMETER),
                detail: Some(doc.to_string()),
                ..Default::default()
            });
        }
    }

    // Add snippet completions
    let snippets = &[
        (
            "fn",
            "fn ${1:name}(${2:params}) {\n\t$0\n}",
            "Function definition",
        ),
        ("if", "if ${1:condition} {\n\t$0\n}", "If statement"),
        (
            "ifelse",
            "if ${1:condition} {\n\t$2\n} else {\n\t$0\n}",
            "If-else statement",
        ),
        (
            "for",
            "for ${1:item} in ${2:iterator} {\n\t$0\n}",
            "For loop",
        ),
        ("while", "while ${1:condition} {\n\t$0\n}", "While loop"),
        (
            "match",
            "match ${1:value} {\n\t${2:pattern} => $0\n}",
            "Match expression",
        ),
        (
            "try",
            "try {\n\t$1\n} catch ${2:e} {\n\t$0\n}",
            "Try-catch block",
        ),
        ("spawn", "spawn {\n\t$0\n}", "Spawn block"),
        ("async", "async {\n\t$0\n}", "Async block"),
        (
            "type",
            "${1:TypeName} {\n\t${2:field}: ${3:type}\n}",
            "Type definition",
        ),
    ];

    for (trigger, snippet, doc) in snippets {
        if trigger.starts_with(&prefix) {
            completions.push(CompletionItem {
                label: trigger.to_string(),
                kind: Some(CompletionItemKind::SNIPPET),
                detail: Some(doc.to_string()),
                insert_text: Some(snippet.to_string()),
                insert_text_format: Some(InsertTextFormat::SNIPPET),
                ..Default::default()
            });
        }
    }

    completions
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

/// Get the word prefix at the given offset.
fn get_word_prefix(source: &str, offset: usize) -> String {
    let before = &source[..offset];
    let mut prefix = String::new();

    for c in before.chars().rev() {
        if c.is_alphanumeric() || c == '_' {
            prefix.insert(0, c);
        } else {
            break;
        }
    }

    prefix
}
