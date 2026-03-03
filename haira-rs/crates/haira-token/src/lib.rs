//! Token kinds and the `Token` type for the Haira lexer.
//!
//! This crate defines every lexical token that the Haira language recognises,
//! together with helper predicates (`is_keyword`, `is_literal`, `is_trivia`)
//! and a fast keyword-lookup function.

use std::fmt;

// ---------------------------------------------------------------------------
// TokenKind
// ---------------------------------------------------------------------------

/// The type of a lexical token.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
#[repr(u8)]
pub enum TokenKind {
    // -- Keywords (0..=49) ---------------------------------------------------
    If,
    Else,
    For,
    While,
    Return,
    Match,
    True,
    False,
    None,
    Some,
    And,
    Or,
    Not,
    In,
    Async,
    Spawn,
    Select,
    Try,
    Catch,
    Pub,
    Export,
    Err,
    Ok,
    Break,
    Continue,
    From,
    Default,
    Provider,
    Tool,
    Agent,
    Workflow,
    Fn,
    Enum,
    Struct,
    Type,
    Trait,
    Impl,
    Defer,
    Import,
    Step,
    Orelse,
    Onerror,
    Onsuccess,
    Oncancel,
    Errdefer,
    Test,
    Assert,
    Let,
    Const,

    // -- Operators -----------------------------------------------------------
    Plus,       // +
    Minus,      // -
    Star,       // *
    Slash,      // /
    Percent,    // %
    EqEq,       // ==
    Ne,         // !=
    Lt,         // <
    Gt,         // >
    Le,         // <=
    Ge,         // >=
    Eq,         // =
    PlusEq,     // +=
    MinusEq,    // -=
    StarEq,     // *=
    SlashEq,    // /=
    PercentEq,  // %=
    Pipe,       // |
    PipeArrow,  // |>
    Amp,        // &
    AmpEq,      // &=
    Caret,      // ^
    CaretEq,    // ^=
    Tilde,      // ~
    Shl,        // <<
    ShlEq,      // <<=
    Shr,        // >>
    ShrEq,      // >>=
    PipeEq,     // |=
    Question,   // ?
    FatArrow,   // =>
    Arrow,      // ->
    DotDotEq,   // ..=
    DotDot,     // ..
    Dot,        // .
    Colon,      // :
    Comma,      // ,
    Ellipsis,   // ...
    At,         // @

    // -- Delimiters ----------------------------------------------------------
    LParen,   // (
    RParen,   // )
    LBrace,   // {
    RBrace,   // }
    LBracket, // [
    RBracket, // ]

    // -- Literals ------------------------------------------------------------
    Int,                // integer literal
    Float,              // float literal
    String,             // string literal
    InterpolatedString, // interpolated string (contains ${expr})
    TripleQuoteString,  // """..."""

    // -- Identifiers ---------------------------------------------------------
    Ident,

    // -- Whitespace / comments -----------------------------------------------
    Newline,      // \n
    LineComment,  // // ...
    BlockComment, // /* ... */

    // -- Special -------------------------------------------------------------
    EOF,
    Error,
}

impl TokenKind {
    /// Returns `true` if the token kind is a keyword (`If` through `Const`).
    #[inline]
    pub fn is_keyword(self) -> bool {
        (self as u8) >= (TokenKind::If as u8) && (self as u8) <= (TokenKind::Const as u8)
    }

    /// Returns `true` if the token kind is a literal value
    /// (`Int`, `Float`, `String`, `InterpolatedString`, `TripleQuoteString`).
    #[inline]
    pub fn is_literal(self) -> bool {
        (self as u8) >= (TokenKind::Int as u8) && (self as u8) <= (TokenKind::TripleQuoteString as u8)
    }

    /// Returns `true` if the token kind is a comment (trivia).
    #[inline]
    pub fn is_trivia(self) -> bool {
        matches!(self, TokenKind::LineComment | TokenKind::BlockComment)
    }
}

impl fmt::Display for TokenKind {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        let s = match self {
            // Keywords
            TokenKind::If => "if",
            TokenKind::Else => "else",
            TokenKind::For => "for",
            TokenKind::While => "while",
            TokenKind::Return => "return",
            TokenKind::Match => "match",
            TokenKind::True => "true",
            TokenKind::False => "false",
            TokenKind::None => "none",
            TokenKind::Some => "some",
            TokenKind::And => "and",
            TokenKind::Or => "or",
            TokenKind::Not => "not",
            TokenKind::In => "in",
            TokenKind::Async => "async",
            TokenKind::Spawn => "spawn",
            TokenKind::Select => "select",
            TokenKind::Try => "try",
            TokenKind::Catch => "catch",
            TokenKind::Pub => "pub",
            TokenKind::Export => "export",
            TokenKind::Err => "err",
            TokenKind::Ok => "ok",
            TokenKind::Break => "break",
            TokenKind::Continue => "continue",
            TokenKind::From => "from",
            TokenKind::Default => "default",
            TokenKind::Provider => "provider",
            TokenKind::Tool => "tool",
            TokenKind::Agent => "agent",
            TokenKind::Workflow => "workflow",
            TokenKind::Fn => "fn",
            TokenKind::Enum => "enum",
            TokenKind::Struct => "struct",
            TokenKind::Type => "type",
            TokenKind::Trait => "trait",
            TokenKind::Impl => "impl",
            TokenKind::Defer => "defer",
            TokenKind::Import => "import",
            TokenKind::Step => "step",
            TokenKind::Orelse => "orelse",
            TokenKind::Onerror => "onerror",
            TokenKind::Onsuccess => "onsuccess",
            TokenKind::Oncancel => "oncancel",
            TokenKind::Errdefer => "errdefer",
            TokenKind::Test => "test",
            TokenKind::Assert => "assert",
            TokenKind::Let => "let",
            TokenKind::Const => "const",

            // Operators
            TokenKind::Plus => "+",
            TokenKind::Minus => "-",
            TokenKind::Star => "*",
            TokenKind::Slash => "/",
            TokenKind::Percent => "%",
            TokenKind::EqEq => "==",
            TokenKind::Ne => "!=",
            TokenKind::Lt => "<",
            TokenKind::Gt => ">",
            TokenKind::Le => "<=",
            TokenKind::Ge => ">=",
            TokenKind::Eq => "=",
            TokenKind::PlusEq => "+=",
            TokenKind::MinusEq => "-=",
            TokenKind::StarEq => "*=",
            TokenKind::SlashEq => "/=",
            TokenKind::PercentEq => "%=",
            TokenKind::Pipe => "|",
            TokenKind::PipeArrow => "|>",
            TokenKind::Amp => "&",
            TokenKind::AmpEq => "&=",
            TokenKind::Caret => "^",
            TokenKind::CaretEq => "^=",
            TokenKind::Tilde => "~",
            TokenKind::Shl => "<<",
            TokenKind::ShlEq => "<<=",
            TokenKind::Shr => ">>",
            TokenKind::ShrEq => ">>=",
            TokenKind::PipeEq => "|=",
            TokenKind::Question => "?",
            TokenKind::FatArrow => "=>",
            TokenKind::Arrow => "->",
            TokenKind::DotDotEq => "..=",
            TokenKind::DotDot => "..",
            TokenKind::Dot => ".",
            TokenKind::Colon => ":",
            TokenKind::Comma => ",",
            TokenKind::Ellipsis => "...",
            TokenKind::At => "@",

            // Delimiters
            TokenKind::LParen => "(",
            TokenKind::RParen => ")",
            TokenKind::LBrace => "{",
            TokenKind::RBrace => "}",
            TokenKind::LBracket => "[",
            TokenKind::RBracket => "]",

            // Literals
            TokenKind::Int => "Int",
            TokenKind::Float => "Float",
            TokenKind::String => "String",
            TokenKind::InterpolatedString => "InterpolatedString",
            TokenKind::TripleQuoteString => "TripleQuoteString",

            // Identifier
            TokenKind::Ident => "Ident",

            // Whitespace / comments
            TokenKind::Newline => "Newline",
            TokenKind::LineComment => "LineComment",
            TokenKind::BlockComment => "BlockComment",

            // Special
            TokenKind::EOF => "EOF",
            TokenKind::Error => "Error",
        };
        f.write_str(s)
    }
}

// ---------------------------------------------------------------------------
// Keyword lookup
// ---------------------------------------------------------------------------

/// Look up a keyword by its string representation.
///
/// Returns `Some(TokenKind)` if the string is a Haira keyword, `None` otherwise.
/// Uses a match statement for fast, allocation-free lookup.
pub fn keyword_lookup(s: &str) -> Option<TokenKind> {
    match s {
        "if" => Some(TokenKind::If),
        "else" => Some(TokenKind::Else),
        "for" => Some(TokenKind::For),
        "while" => Some(TokenKind::While),
        "return" => Some(TokenKind::Return),
        "match" => Some(TokenKind::Match),
        "true" => Some(TokenKind::True),
        "false" => Some(TokenKind::False),
        "none" => Some(TokenKind::None),
        "some" => Some(TokenKind::Some),
        "and" => Some(TokenKind::And),
        "or" => Some(TokenKind::Or),
        "not" => Some(TokenKind::Not),
        "in" => Some(TokenKind::In),
        "async" => Some(TokenKind::Async),
        "spawn" => Some(TokenKind::Spawn),
        "select" => Some(TokenKind::Select),
        "try" => Some(TokenKind::Try),
        "catch" => Some(TokenKind::Catch),
        "pub" => Some(TokenKind::Pub),
        "export" => Some(TokenKind::Export),
        "err" => Some(TokenKind::Err),
        "ok" => Some(TokenKind::Ok),
        "break" => Some(TokenKind::Break),
        "continue" => Some(TokenKind::Continue),
        "from" => Some(TokenKind::From),
        "default" => Some(TokenKind::Default),
        "provider" => Some(TokenKind::Provider),
        "tool" => Some(TokenKind::Tool),
        "agent" => Some(TokenKind::Agent),
        "workflow" => Some(TokenKind::Workflow),
        "fn" => Some(TokenKind::Fn),
        "enum" => Some(TokenKind::Enum),
        "struct" => Some(TokenKind::Struct),
        "type" => Some(TokenKind::Type),
        "trait" => Some(TokenKind::Trait),
        "impl" => Some(TokenKind::Impl),
        "defer" => Some(TokenKind::Defer),
        "import" => Some(TokenKind::Import),
        "step" => Some(TokenKind::Step),
        "orelse" => Some(TokenKind::Orelse),
        "onerror" => Some(TokenKind::Onerror),
        "onsuccess" => Some(TokenKind::Onsuccess),
        "oncancel" => Some(TokenKind::Oncancel),
        "errdefer" => Some(TokenKind::Errdefer),
        "test" => Some(TokenKind::Test),
        "assert" => Some(TokenKind::Assert),
        "let" => Some(TokenKind::Let),
        "const" => Some(TokenKind::Const),
        _ => None,
    }
}

// ---------------------------------------------------------------------------
// Token
// ---------------------------------------------------------------------------

/// A lexical token with its kind, literal value, and source position.
#[derive(Debug, Clone)]
pub struct Token {
    /// The kind of token.
    pub kind: TokenKind,
    /// Raw text of the token.
    pub value: std::string::String,
    /// Byte offset start (inclusive).
    pub start: usize,
    /// Byte offset end (exclusive).
    pub end: usize,
}

impl Token {
    /// Create a new token.
    pub fn new(kind: TokenKind, value: impl Into<std::string::String>, start: usize, end: usize) -> Self {
        Self {
            kind,
            value: value.into(),
            start,
            end,
        }
    }
}

impl fmt::Display for Token {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        if !self.value.is_empty() {
            write!(f, "{}({:?})", self.kind, self.value)
        } else {
            write!(f, "{}", self.kind)
        }
    }
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn keyword_lookup_returns_correct_kinds() {
        assert_eq!(keyword_lookup("if"), Some(TokenKind::If));
        assert_eq!(keyword_lookup("workflow"), Some(TokenKind::Workflow));
        assert_eq!(keyword_lookup("const"), Some(TokenKind::Const));
        assert_eq!(keyword_lookup("notakeyword"), None);
        assert_eq!(keyword_lookup(""), None);
    }

    #[test]
    fn is_keyword_range() {
        assert!(TokenKind::If.is_keyword());
        assert!(TokenKind::Const.is_keyword());
        assert!(TokenKind::Workflow.is_keyword());
        assert!(!TokenKind::Plus.is_keyword());
        assert!(!TokenKind::Ident.is_keyword());
        assert!(!TokenKind::EOF.is_keyword());
    }

    #[test]
    fn is_literal_range() {
        assert!(TokenKind::Int.is_literal());
        assert!(TokenKind::Float.is_literal());
        assert!(TokenKind::String.is_literal());
        assert!(TokenKind::InterpolatedString.is_literal());
        assert!(TokenKind::TripleQuoteString.is_literal());
        assert!(!TokenKind::Ident.is_literal());
        assert!(!TokenKind::Plus.is_literal());
    }

    #[test]
    fn is_trivia_for_comments() {
        assert!(TokenKind::LineComment.is_trivia());
        assert!(TokenKind::BlockComment.is_trivia());
        assert!(!TokenKind::Newline.is_trivia());
        assert!(!TokenKind::Ident.is_trivia());
    }

    #[test]
    fn token_display_with_value() {
        let tok = Token::new(TokenKind::Ident, "foo", 0, 3);
        assert_eq!(tok.to_string(), "Ident(\"foo\")");
    }

    #[test]
    fn token_display_without_value() {
        let tok = Token::new(TokenKind::EOF, "", 10, 10);
        assert_eq!(tok.to_string(), "EOF");
    }

    #[test]
    fn token_kind_display_operators() {
        assert_eq!(TokenKind::PipeArrow.to_string(), "|>");
        assert_eq!(TokenKind::FatArrow.to_string(), "=>");
        assert_eq!(TokenKind::Arrow.to_string(), "->");
        assert_eq!(TokenKind::DotDotEq.to_string(), "..=");
        assert_eq!(TokenKind::Ellipsis.to_string(), "...");
        assert_eq!(TokenKind::ShlEq.to_string(), "<<=");
        assert_eq!(TokenKind::ShrEq.to_string(), ">>=");
    }

    #[test]
    fn all_keywords_round_trip() {
        let keywords = [
            "if", "else", "for", "while", "return", "match", "true", "false",
            "none", "some", "and", "or", "not", "in", "async", "spawn",
            "select", "try", "catch", "pub", "export", "err", "ok", "break",
            "continue", "from", "default", "provider", "tool", "agent",
            "workflow", "fn", "enum", "struct", "type", "trait", "impl",
            "defer", "import", "step", "orelse", "onerror", "onsuccess",
            "oncancel", "errdefer", "test", "assert", "let", "const",
        ];
        for kw in &keywords {
            let kind = keyword_lookup(kw).unwrap_or_else(|| panic!("keyword_lookup failed for {:?}", kw));
            assert!(kind.is_keyword(), "{:?} should be a keyword", kw);
            assert_eq!(kind.to_string(), *kw, "Display round-trip failed for {:?}", kw);
        }
    }
}
