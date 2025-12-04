//! Token definitions for Haira.

use logos::Logos;
use smol_str::SmolStr;

/// A token with its kind and span.
#[derive(Debug, Clone, PartialEq)]
pub struct Token {
    pub kind: TokenKind,
    pub span: std::ops::Range<usize>,
}

impl Token {
    pub fn new(kind: TokenKind, span: std::ops::Range<usize>) -> Self {
        Self { kind, span }
    }
}

/// Token kinds for Haira.
#[derive(Logos, Debug, Clone, PartialEq)]
#[logos(skip r"[ \t]+")] // Skip spaces and tabs
pub enum TokenKind {
    // ========================================================================
    // Keywords
    // ========================================================================
    #[token("if")]
    If,
    #[token("else")]
    Else,
    #[token("for")]
    For,
    #[token("while")]
    While,
    #[token("return")]
    Return,
    #[token("match")]
    Match,
    #[token("true")]
    True,
    #[token("false")]
    False,
    #[token("none")]
    None,
    #[token("some")]
    Some,
    #[token("and")]
    And,
    #[token("or")]
    Or,
    #[token("not")]
    Not,
    #[token("in")]
    In,
    #[token("async")]
    Async,
    #[token("spawn")]
    Spawn,
    #[token("select")]
    Select,
    #[token("try")]
    Try,
    #[token("catch")]
    Catch,
    #[token("public")]
    Public,
    #[token("err")]
    Err,
    #[token("ok")]
    Ok,
    #[token("break")]
    Break,
    #[token("continue")]
    Continue,
    #[token("from")]
    From,
    #[token("default")]
    Default,
    #[token("ai")]
    Ai,

    // ========================================================================
    // Operators
    // ========================================================================
    #[token("+")]
    Plus,
    #[token("-")]
    Minus,
    #[token("*")]
    Star,
    #[token("/")]
    Slash,
    #[token("%")]
    Percent,

    #[token("==")]
    EqEq,
    #[token("!=")]
    Ne,
    #[token("<")]
    Lt,
    #[token(">")]
    Gt,
    #[token("<=")]
    Le,
    #[token(">=")]
    Ge,

    #[token("=")]
    Eq,
    #[token("|")]
    Pipe,
    #[token("?")]
    Question,
    #[token("=>")]
    FatArrow,
    #[token("->")]
    Arrow,
    #[token("..=")]
    DotDotEq,
    #[token("..")]
    DotDot,
    #[token(".")]
    Dot,
    #[token(":")]
    Colon,
    #[token(",")]
    Comma,
    #[token("...")]
    Ellipsis,

    // ========================================================================
    // Delimiters
    // ========================================================================
    #[token("(")]
    LParen,
    #[token(")")]
    RParen,
    #[token("{")]
    LBrace,
    #[token("}")]
    RBrace,
    #[token("[")]
    LBracket,
    #[token("]")]
    RBracket,

    // ========================================================================
    // Literals
    // ========================================================================
    /// Integer literal
    #[regex(r"[0-9][0-9_]*", |lex| parse_int(lex.slice()))]
    #[regex(r"0x[0-9a-fA-F][0-9a-fA-F_]*", |lex| parse_hex(lex.slice()))]
    #[regex(r"0b[01][01_]*", |lex| parse_binary(lex.slice()))]
    #[regex(r"0o[0-7][0-7_]*", |lex| parse_octal(lex.slice()))]
    Int(i64),

    /// Float literal
    #[regex(r"[0-9][0-9_]*\.[0-9][0-9_]*", |lex| parse_float(lex.slice()))]
    Float(f64),

    /// String literal (simple strings without interpolation)
    #[regex(r#""([^"\\{]|\\.)*""#, |lex| parse_string(lex.slice()))]
    String(SmolStr),

    /// Interpolated string literal (contains `{...}` expressions)
    /// We match strings that contain `{` and parse them specially
    #[regex(r#""([^"\\]|\\.)*""#, |lex| parse_interpolated_string(lex.slice()), priority = 1)]
    InterpolatedString(SmolStr),

    /// Identifier
    #[regex(r"[a-zA-Z_][a-zA-Z0-9_]*", |lex| SmolStr::from(lex.slice()))]
    Ident(SmolStr),

    // ========================================================================
    // Whitespace and Comments
    // ========================================================================
    /// Newline (significant for statement separation)
    #[regex(r"\n|\r\n")]
    Newline,

    /// Single-line comment
    #[regex(r"//[^\n]*")]
    LineComment,

    /// Multi-line comment (handled specially)
    #[token("/*", |lex| skip_block_comment(lex))]
    BlockComment,

    /// End of file
    Eof,

    /// Error token
    Error,
}

impl TokenKind {
    /// Check if this token is a keyword.
    pub fn is_keyword(&self) -> bool {
        matches!(
            self,
            TokenKind::If
                | TokenKind::Else
                | TokenKind::For
                | TokenKind::While
                | TokenKind::Return
                | TokenKind::Match
                | TokenKind::True
                | TokenKind::False
                | TokenKind::None
                | TokenKind::Some
                | TokenKind::And
                | TokenKind::Or
                | TokenKind::Not
                | TokenKind::In
                | TokenKind::Async
                | TokenKind::Spawn
                | TokenKind::Select
                | TokenKind::Try
                | TokenKind::Catch
                | TokenKind::Public
                | TokenKind::Err
                | TokenKind::Ok
                | TokenKind::Break
                | TokenKind::Continue
                | TokenKind::From
                | TokenKind::Default
                | TokenKind::Ai
        )
    }

    /// Check if this token is a literal.
    pub fn is_literal(&self) -> bool {
        matches!(
            self,
            TokenKind::Int(_)
                | TokenKind::Float(_)
                | TokenKind::String(_)
                | TokenKind::InterpolatedString(_)
                | TokenKind::True
                | TokenKind::False
                | TokenKind::None
        )
    }

    /// Check if this token is trivia (comments, etc.)
    pub fn is_trivia(&self) -> bool {
        matches!(self, TokenKind::LineComment | TokenKind::BlockComment)
    }
}

// ============================================================================
// Helper functions for parsing
// ============================================================================

fn parse_int(s: &str) -> Option<i64> {
    let s = s.replace('_', "");
    s.parse().ok()
}

fn parse_hex(s: &str) -> Option<i64> {
    let s = s.strip_prefix("0x").unwrap_or(s).replace('_', "");
    i64::from_str_radix(&s, 16).ok()
}

fn parse_binary(s: &str) -> Option<i64> {
    let s = s.strip_prefix("0b").unwrap_or(s).replace('_', "");
    i64::from_str_radix(&s, 2).ok()
}

fn parse_octal(s: &str) -> Option<i64> {
    let s = s.strip_prefix("0o").unwrap_or(s).replace('_', "");
    i64::from_str_radix(&s, 8).ok()
}

fn parse_float(s: &str) -> Option<f64> {
    let s = s.replace('_', "");
    s.parse().ok()
}

fn parse_string(s: &str) -> Option<SmolStr> {
    // Remove quotes
    let s = s.strip_prefix('"')?.strip_suffix('"')?;

    // Handle escape sequences
    let mut result = String::with_capacity(s.len());
    let mut chars = s.chars().peekable();

    while let Some(c) = chars.next() {
        if c == '\\' {
            match chars.next() {
                Some('n') => result.push('\n'),
                Some('t') => result.push('\t'),
                Some('r') => result.push('\r'),
                Some('\\') => result.push('\\'),
                Some('"') => result.push('"'),
                Some('{') => result.push('{'),
                Some('}') => result.push('}'),
                Some(c) => {
                    result.push('\\');
                    result.push(c);
                }
                None => result.push('\\'),
            }
        } else {
            result.push(c);
        }
    }

    Some(SmolStr::from(result))
}

/// Parse an interpolated string, keeping the raw content for the parser to process.
/// Returns None if the string doesn't contain interpolation (handled by simple String).
fn parse_interpolated_string(s: &str) -> Option<SmolStr> {
    // Check if string contains unescaped `{`
    let inner = s.strip_prefix('"')?.strip_suffix('"')?;

    let mut has_interpolation = false;
    let mut chars = inner.chars().peekable();
    while let Some(c) = chars.next() {
        if c == '\\' {
            // Skip escaped character
            chars.next();
        } else if c == '{' {
            has_interpolation = true;
            break;
        }
    }

    if has_interpolation {
        // Return the raw content (without quotes) for the parser to process
        Some(SmolStr::from(inner))
    } else {
        // This shouldn't happen due to regex priority, but handle it
        None
    }
}

fn skip_block_comment(lex: &mut logos::Lexer<TokenKind>) -> logos::Skip {
    let remainder = lex.remainder();
    let mut depth = 1;
    let mut chars = remainder.char_indices();

    while let Some((i, c)) = chars.next() {
        match c {
            '*' => {
                if let Some((_, '/')) = chars.clone().next() {
                    chars.next();
                    depth -= 1;
                    if depth == 0 {
                        lex.bump(i + 2);
                        return logos::Skip;
                    }
                }
            }
            '/' => {
                if let Some((_, '*')) = chars.clone().next() {
                    chars.next();
                    depth += 1;
                }
            }
            _ => {}
        }
    }

    // Unclosed comment - bump to end
    lex.bump(remainder.len());
    logos::Skip
}

#[cfg(test)]
mod tests {
    use super::*;
    use logos::Logos;

    #[test]
    fn test_keywords() {
        let mut lex = TokenKind::lexer("if else for while return");
        assert_eq!(lex.next(), Some(Ok(TokenKind::If)));
        assert_eq!(lex.next(), Some(Ok(TokenKind::Else)));
        assert_eq!(lex.next(), Some(Ok(TokenKind::For)));
        assert_eq!(lex.next(), Some(Ok(TokenKind::While)));
        assert_eq!(lex.next(), Some(Ok(TokenKind::Return)));
    }

    #[test]
    fn test_integers() {
        let mut lex = TokenKind::lexer("42 1_000_000 0xFF 0b1010 0o755");
        assert_eq!(lex.next(), Some(Ok(TokenKind::Int(42))));
        assert_eq!(lex.next(), Some(Ok(TokenKind::Int(1_000_000))));
        assert_eq!(lex.next(), Some(Ok(TokenKind::Int(255))));
        assert_eq!(lex.next(), Some(Ok(TokenKind::Int(10))));
        assert_eq!(lex.next(), Some(Ok(TokenKind::Int(493))));
    }

    #[test]
    fn test_floats() {
        let mut lex = TokenKind::lexer("3.15 1_000.5");
        assert_eq!(lex.next(), Some(Ok(TokenKind::Float(3.15))));
        assert_eq!(lex.next(), Some(Ok(TokenKind::Float(1000.5))));
    }

    #[test]
    fn test_strings() {
        let mut lex = TokenKind::lexer(r#""hello" "world\n""#);
        assert_eq!(
            lex.next(),
            Some(Ok(TokenKind::String(SmolStr::from("hello"))))
        );
        assert_eq!(
            lex.next(),
            Some(Ok(TokenKind::String(SmolStr::from("world\n"))))
        );
    }

    #[test]
    fn test_identifiers() {
        let mut lex = TokenKind::lexer("foo bar_baz _private");
        assert_eq!(lex.next(), Some(Ok(TokenKind::Ident(SmolStr::from("foo")))));
        assert_eq!(
            lex.next(),
            Some(Ok(TokenKind::Ident(SmolStr::from("bar_baz"))))
        );
        assert_eq!(
            lex.next(),
            Some(Ok(TokenKind::Ident(SmolStr::from("_private"))))
        );
    }

    #[test]
    fn test_operators() {
        let mut lex = TokenKind::lexer("+ - * / == != <= >= = | ? => -> .. ..=");
        assert_eq!(lex.next(), Some(Ok(TokenKind::Plus)));
        assert_eq!(lex.next(), Some(Ok(TokenKind::Minus)));
        assert_eq!(lex.next(), Some(Ok(TokenKind::Star)));
        assert_eq!(lex.next(), Some(Ok(TokenKind::Slash)));
        assert_eq!(lex.next(), Some(Ok(TokenKind::EqEq)));
        assert_eq!(lex.next(), Some(Ok(TokenKind::Ne)));
        assert_eq!(lex.next(), Some(Ok(TokenKind::Le)));
        assert_eq!(lex.next(), Some(Ok(TokenKind::Ge)));
        assert_eq!(lex.next(), Some(Ok(TokenKind::Eq)));
        assert_eq!(lex.next(), Some(Ok(TokenKind::Pipe)));
        assert_eq!(lex.next(), Some(Ok(TokenKind::Question)));
        assert_eq!(lex.next(), Some(Ok(TokenKind::FatArrow)));
        assert_eq!(lex.next(), Some(Ok(TokenKind::Arrow)));
        assert_eq!(lex.next(), Some(Ok(TokenKind::DotDot)));
        assert_eq!(lex.next(), Some(Ok(TokenKind::DotDotEq)));
    }
}
