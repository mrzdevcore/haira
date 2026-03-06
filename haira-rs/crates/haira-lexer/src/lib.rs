//! Hand-written scanner for Haira source code.
//!
//! The lexer eagerly tokenizes the entire source on construction, storing all
//! tokens in a `Vec`. Iteration then walks the pre-built token list, skipping
//! trivia (and optionally newlines).

use haira_token::{keyword_lookup, Token, TokenKind};

// ---------------------------------------------------------------------------
// Lexer
// ---------------------------------------------------------------------------

/// Tokenizes Haira source code.
///
/// Created via [`Lexer::new`], which immediately scans the entire source.
/// Tokens are then consumed via [`Lexer::next`] or [`Lexer::next_with_newlines`].
pub struct Lexer {
    /// The original source text.
    source: String,
    /// Current byte position during scanning.
    pos: usize,
    /// All tokens produced by the scan (including trivia).
    tokens: Vec<Token>,
    /// Index into `tokens` for the iteration cursor.
    current: usize,
}

impl Lexer {
    /// Creates a new lexer and immediately tokenizes the entire `source`.
    pub fn new(source: &str) -> Self {
        let mut lexer = Self {
            source: source.to_owned(),
            pos: 0,
            tokens: Vec::new(),
            current: 0,
        };
        lexer.scan();
        lexer
    }

    /// Returns the next non-trivia, non-newline token, advancing the cursor.
    ///
    /// Once all tokens have been consumed, returns `EOF`.
    pub fn next(&mut self) -> Token {
        while self.current < self.tokens.len() {
            let tok = self.tokens[self.current].clone();
            self.current += 1;
            if tok.kind.is_trivia() || tok.kind == TokenKind::Newline {
                continue;
            }
            return tok;
        }
        let pos = self.source.len();
        Token::new(TokenKind::EOF, "", pos, pos)
    }

    /// Returns the next non-trivia token (including newlines), advancing the cursor.
    ///
    /// Once all tokens have been consumed, returns `EOF`.
    pub fn next_with_newlines(&mut self) -> Token {
        while self.current < self.tokens.len() {
            let tok = self.tokens[self.current].clone();
            self.current += 1;
            if tok.kind.is_trivia() {
                continue;
            }
            return tok;
        }
        let pos = self.source.len();
        Token::new(TokenKind::EOF, "", pos, pos)
    }

    /// Returns all tokens (including trivia and newlines) produced by the scan.
    pub fn all_tokens(&self) -> &[Token] {
        &self.tokens
    }

    // -----------------------------------------------------------------------
    // Main scanning loop
    // -----------------------------------------------------------------------

    /// Tokenizes the entire source into `self.tokens`.
    fn scan(&mut self) {
        let src = self.source.clone();
        let bytes = src.as_bytes();
        let len = bytes.len();

        while self.pos < len {
            self.skip_spaces_and_tabs(bytes, len);
            if self.pos >= len {
                break;
            }

            let start = self.pos;
            let ch = bytes[self.pos];

            match ch {
                b'\n' | b'\r' => self.scan_newline(bytes, len, start),
                b'/' if self.pos + 1 < len && bytes[self.pos + 1] == b'/' => {
                    self.scan_line_comment(bytes, len, start)
                }
                b'/' if self.pos + 1 < len && bytes[self.pos + 1] == b'*' => {
                    self.scan_block_comment(bytes, len, start)
                }
                b'"' => self.scan_string(bytes, len, start),
                b'0'..=b'9' => self.scan_number(bytes, len, start),
                _ if is_ident_start(ch) => self.scan_ident_or_keyword(bytes, len, start),
                _ => self.scan_operator_or_delimiter(bytes, len, start),
            }
        }

        // Append EOF sentinel.
        self.tokens.push(Token::new(TokenKind::EOF, "", self.pos, self.pos));
    }

    // -----------------------------------------------------------------------
    // Whitespace
    // -----------------------------------------------------------------------

    /// Skip spaces and tabs (but not newlines).
    #[inline]
    fn skip_spaces_and_tabs(&mut self, bytes: &[u8], len: usize) {
        while self.pos < len {
            let ch = bytes[self.pos];
            if ch == b' ' || ch == b'\t' {
                self.pos += 1;
            } else {
                break;
            }
        }
    }

    // -----------------------------------------------------------------------
    // Newlines
    // -----------------------------------------------------------------------

    fn scan_newline(&mut self, bytes: &[u8], len: usize, start: usize) {
        if bytes[self.pos] == b'\r' && self.pos + 1 < len && bytes[self.pos + 1] == b'\n' {
            self.pos += 2;
        } else {
            self.pos += 1;
        }
        self.tokens.push(Token::new(TokenKind::Newline, "", start, self.pos));
    }

    // -----------------------------------------------------------------------
    // Comments
    // -----------------------------------------------------------------------

    fn scan_line_comment(&mut self, bytes: &[u8], len: usize, start: usize) {
        self.pos += 2; // skip //
        while self.pos < len && bytes[self.pos] != b'\n' {
            self.pos += 1;
        }
        let value = &self.source[start..self.pos];
        self.tokens.push(Token::new(TokenKind::LineComment, value, start, self.pos));
    }

    fn scan_block_comment(&mut self, bytes: &[u8], len: usize, start: usize) {
        self.pos += 2; // skip /*
        let mut depth: usize = 1;
        while self.pos < len && depth > 0 {
            if self.pos + 1 < len {
                if bytes[self.pos] == b'*' && bytes[self.pos + 1] == b'/' {
                    depth -= 1;
                    self.pos += 2;
                    continue;
                }
                if bytes[self.pos] == b'/' && bytes[self.pos + 1] == b'*' {
                    depth += 1;
                    self.pos += 2;
                    continue;
                }
            }
            self.pos += 1;
        }
        let value = &self.source[start..self.pos];
        self.tokens.push(Token::new(TokenKind::BlockComment, value, start, self.pos));
    }

    // -----------------------------------------------------------------------
    // Strings
    // -----------------------------------------------------------------------

    fn scan_string(&mut self, bytes: &[u8], len: usize, start: usize) {
        // Check for triple-quoted string: """..."""
        if self.pos + 2 < len && bytes[self.pos + 1] == b'"' && bytes[self.pos + 2] == b'"' {
            self.scan_triple_quote_string(len, start);
            return;
        }

        self.pos += 1; // skip opening "
        let mut has_interpolation = false;
        let mut buf = String::new();
        let mut interp_depth: usize = 0; // track brace depth inside ${...}

        while self.pos < len {
            let ch = bytes[self.pos];

            // Only treat " as string terminator when NOT inside ${...}
            if ch == b'"' && interp_depth == 0 {
                break;
            }

            if ch == b'\\' && interp_depth == 0 {
                self.pos += 1;
                if self.pos < len {
                    let esc = bytes[self.pos];
                    match esc {
                        b'n' => buf.push('\n'),
                        b't' => buf.push('\t'),
                        b'r' => buf.push('\r'),
                        b'\\' => buf.push('\\'),
                        b'"' => buf.push('"'),
                        b'{' => buf.push('{'),
                        b'}' => buf.push('}'),
                        _ => {
                            buf.push('\\');
                            buf.push(esc as char);
                        }
                    }
                    self.pos += 1;
                }
                continue;
            }

            if ch == b'$'
                && self.pos + 1 < len
                && bytes[self.pos + 1] == b'{'
                && interp_depth == 0
            {
                has_interpolation = true;
                interp_depth = 1;
                buf.push('$');
                self.pos += 1;
                buf.push(bytes[self.pos] as char);
                self.pos += 1;
                continue;
            }

            if interp_depth > 0 {
                if ch == b'{' {
                    interp_depth += 1;
                } else if ch == b'}' {
                    interp_depth -= 1;
                }
                // Inside interpolation, pass through quotes and everything else
                if ch == b'"' {
                    // Scan the nested string literal inside ${...}
                    buf.push(ch as char);
                    self.pos += 1;
                    while self.pos < len && bytes[self.pos] != b'"' {
                        if bytes[self.pos] == b'\\' && self.pos + 1 < len {
                            buf.push(bytes[self.pos] as char);
                            self.pos += 1;
                            buf.push(bytes[self.pos] as char);
                            self.pos += 1;
                            continue;
                        }
                        buf.push(bytes[self.pos] as char);
                        self.pos += 1;
                    }
                    if self.pos < len {
                        buf.push(bytes[self.pos] as char); // closing "
                        self.pos += 1;
                    }
                    continue;
                }
            }

            buf.push(ch as char);
            self.pos += 1;
        }

        if self.pos < len {
            self.pos += 1; // skip closing "
        }

        if has_interpolation {
            // Return the raw content (without quotes) for the parser to process
            let raw = &self.source[start + 1..self.pos - 1];
            self.tokens.push(Token::new(
                TokenKind::InterpolatedString,
                raw,
                start,
                self.pos,
            ));
        } else {
            self.tokens.push(Token::new(TokenKind::String, buf, start, self.pos));
        }
    }

    fn scan_triple_quote_string(&mut self, len: usize, start: usize) {
        self.pos += 3; // skip opening """
        let content_start = self.pos;

        // Find closing """
        if let Some(idx) = self.source[self.pos..].find("\"\"\"") {
            let content = &self.source[content_start..self.pos + idx];
            self.pos += idx + 3; // skip content + closing """

            // Dedent: strip common leading whitespace
            let content = dedent(content);
            self.tokens.push(Token::new(
                TokenKind::TripleQuoteString,
                content,
                start,
                self.pos,
            ));
        } else {
            // No closing """ -- consume everything
            self.pos = len;
            let value = &self.source[start..self.pos];
            self.tokens.push(Token::new(TokenKind::Error, value, start, self.pos));
        }
    }

    // -----------------------------------------------------------------------
    // Numbers
    // -----------------------------------------------------------------------

    fn scan_number(&mut self, bytes: &[u8], len: usize, start: usize) {
        // Check for hex, binary, octal prefixes
        if bytes[self.pos] == b'0' && self.pos + 1 < len {
            let next = bytes[self.pos + 1];
            match next {
                b'x' | b'X' => {
                    self.pos += 2;
                    while self.pos < len
                        && (is_hex_digit(bytes[self.pos]) || bytes[self.pos] == b'_')
                    {
                        self.pos += 1;
                    }
                    let value = &self.source[start..self.pos];
                    self.tokens.push(Token::new(TokenKind::Int, value, start, self.pos));
                    return;
                }
                b'b' | b'B' => {
                    self.pos += 2;
                    while self.pos < len
                        && (bytes[self.pos] == b'0'
                            || bytes[self.pos] == b'1'
                            || bytes[self.pos] == b'_')
                    {
                        self.pos += 1;
                    }
                    let value = &self.source[start..self.pos];
                    self.tokens.push(Token::new(TokenKind::Int, value, start, self.pos));
                    return;
                }
                b'o' | b'O' => {
                    self.pos += 2;
                    while self.pos < len
                        && ((bytes[self.pos] >= b'0' && bytes[self.pos] <= b'7')
                            || bytes[self.pos] == b'_')
                    {
                        self.pos += 1;
                    }
                    let value = &self.source[start..self.pos];
                    self.tokens.push(Token::new(TokenKind::Int, value, start, self.pos));
                    return;
                }
                _ => {}
            }
        }

        // Decimal integer or float
        while self.pos < len && (is_digit(bytes[self.pos]) || bytes[self.pos] == b'_') {
            self.pos += 1;
        }

        // Check for float: must have digit after dot (not .. or ..=)
        if self.pos < len && bytes[self.pos] == b'.' {
            if self.pos + 1 < len && is_digit(bytes[self.pos + 1]) {
                self.pos += 1; // skip .
                while self.pos < len && (is_digit(bytes[self.pos]) || bytes[self.pos] == b'_') {
                    self.pos += 1;
                }
                let value = &self.source[start..self.pos];
                self.tokens.push(Token::new(TokenKind::Float, value, start, self.pos));
                return;
            }
        }

        let value = &self.source[start..self.pos];
        self.tokens.push(Token::new(TokenKind::Int, value, start, self.pos));
    }

    // -----------------------------------------------------------------------
    // Identifiers & keywords
    // -----------------------------------------------------------------------

    fn scan_ident_or_keyword(&mut self, bytes: &[u8], len: usize, start: usize) {
        while self.pos < len && is_ident_continue(bytes[self.pos]) {
            self.pos += 1;
        }
        let word = &self.source[start..self.pos];

        if let Some(kind) = keyword_lookup(word) {
            self.tokens.push(Token::new(kind, word, start, self.pos));
        } else {
            self.tokens.push(Token::new(TokenKind::Ident, word, start, self.pos));
        }
    }

    // -----------------------------------------------------------------------
    // Operators & delimiters
    // -----------------------------------------------------------------------

    fn scan_operator_or_delimiter(&mut self, bytes: &[u8], len: usize, start: usize) {
        let ch = bytes[self.pos];
        self.pos += 1;

        // Two-or-three character operators (check longest match first)
        if self.pos < len {
            let next = bytes[self.pos];
            match (ch, next) {
                // 3-char: ..=
                (b'.', b'.') if self.pos + 1 < len && bytes[self.pos + 1] == b'=' => {
                    self.pos += 2;
                    self.tokens.push(Token::new(TokenKind::DotDotEq, "", start, self.pos));
                    return;
                }
                // 3-char: ...
                (b'.', b'.') if self.pos + 1 < len && bytes[self.pos + 1] == b'.' => {
                    self.pos += 2;
                    self.tokens.push(Token::new(TokenKind::Ellipsis, "", start, self.pos));
                    return;
                }
                // 2-char: ..
                (b'.', b'.') => {
                    self.pos += 1;
                    self.tokens.push(Token::new(TokenKind::DotDot, "", start, self.pos));
                    return;
                }
                // 2-char: ==
                (b'=', b'=') => {
                    self.pos += 1;
                    self.tokens.push(Token::new(TokenKind::EqEq, "", start, self.pos));
                    return;
                }
                // 2-char: !=
                (b'!', b'=') => {
                    self.pos += 1;
                    self.tokens.push(Token::new(TokenKind::Ne, "", start, self.pos));
                    return;
                }
                // 3-char: <<= or 2-char: <<
                (b'<', b'<') => {
                    if self.pos + 1 < len && bytes[self.pos + 1] == b'=' {
                        self.pos += 2;
                        self.tokens.push(Token::new(TokenKind::ShlEq, "", start, self.pos));
                    } else {
                        self.pos += 1;
                        self.tokens.push(Token::new(TokenKind::Shl, "", start, self.pos));
                    }
                    return;
                }
                // 2-char: <=
                (b'<', b'=') => {
                    self.pos += 1;
                    self.tokens.push(Token::new(TokenKind::Le, "", start, self.pos));
                    return;
                }
                // 3-char: >>= or 2-char: >>
                (b'>', b'>') => {
                    if self.pos + 1 < len && bytes[self.pos + 1] == b'=' {
                        self.pos += 2;
                        self.tokens.push(Token::new(TokenKind::ShrEq, "", start, self.pos));
                    } else {
                        self.pos += 1;
                        self.tokens.push(Token::new(TokenKind::Shr, "", start, self.pos));
                    }
                    return;
                }
                // 2-char: >=
                (b'>', b'=') => {
                    self.pos += 1;
                    self.tokens.push(Token::new(TokenKind::Ge, "", start, self.pos));
                    return;
                }
                // 2-char: |>
                (b'|', b'>') => {
                    self.pos += 1;
                    self.tokens.push(Token::new(TokenKind::PipeArrow, "", start, self.pos));
                    return;
                }
                // 2-char: |=
                (b'|', b'=') => {
                    self.pos += 1;
                    self.tokens.push(Token::new(TokenKind::PipeEq, "", start, self.pos));
                    return;
                }
                // 2-char: &=
                (b'&', b'=') => {
                    self.pos += 1;
                    self.tokens.push(Token::new(TokenKind::AmpEq, "", start, self.pos));
                    return;
                }
                // 2-char: ^=
                (b'^', b'=') => {
                    self.pos += 1;
                    self.tokens.push(Token::new(TokenKind::CaretEq, "", start, self.pos));
                    return;
                }
                // 2-char: =>
                (b'=', b'>') => {
                    self.pos += 1;
                    self.tokens.push(Token::new(TokenKind::FatArrow, "", start, self.pos));
                    return;
                }
                // 2-char: ->
                (b'-', b'>') => {
                    self.pos += 1;
                    self.tokens.push(Token::new(TokenKind::Arrow, "", start, self.pos));
                    return;
                }
                // 2-char: +=
                (b'+', b'=') => {
                    self.pos += 1;
                    self.tokens.push(Token::new(TokenKind::PlusEq, "", start, self.pos));
                    return;
                }
                // 2-char: -=
                (b'-', b'=') => {
                    self.pos += 1;
                    self.tokens.push(Token::new(TokenKind::MinusEq, "", start, self.pos));
                    return;
                }
                // 2-char: *=
                (b'*', b'=') => {
                    self.pos += 1;
                    self.tokens.push(Token::new(TokenKind::StarEq, "", start, self.pos));
                    return;
                }
                // 2-char: /=
                (b'/', b'=') => {
                    self.pos += 1;
                    self.tokens.push(Token::new(TokenKind::SlashEq, "", start, self.pos));
                    return;
                }
                // 2-char: %=
                (b'%', b'=') => {
                    self.pos += 1;
                    self.tokens.push(Token::new(TokenKind::PercentEq, "", start, self.pos));
                    return;
                }
                _ => {}
            }
        }

        // Single character operators/delimiters
        let kind = match ch {
            b'+' => TokenKind::Plus,
            b'-' => TokenKind::Minus,
            b'*' => TokenKind::Star,
            b'/' => TokenKind::Slash,
            b'%' => TokenKind::Percent,
            b'<' => TokenKind::Lt,
            b'>' => TokenKind::Gt,
            b'=' => TokenKind::Eq,
            b'|' => TokenKind::Pipe,
            b'&' => TokenKind::Amp,
            b'^' => TokenKind::Caret,
            b'~' => TokenKind::Tilde,
            b'?' => TokenKind::Question,
            b'.' => TokenKind::Dot,
            b':' => TokenKind::Colon,
            b',' => TokenKind::Comma,
            b'@' => TokenKind::At,
            b'(' => TokenKind::LParen,
            b')' => TokenKind::RParen,
            b'{' => TokenKind::LBrace,
            b'}' => TokenKind::RBrace,
            b'[' => TokenKind::LBracket,
            b']' => TokenKind::RBracket,
            _ => TokenKind::Error,
        };

        let value = &self.source[start..self.pos];
        self.tokens.push(Token::new(kind, value, start, self.pos));
    }
}

// ---------------------------------------------------------------------------
// Helper functions
// ---------------------------------------------------------------------------

/// Returns `true` if the byte is an ASCII digit `0`..`9`.
#[inline]
fn is_digit(ch: u8) -> bool {
    ch >= b'0' && ch <= b'9'
}

/// Returns `true` if the byte is a hexadecimal digit.
#[inline]
fn is_hex_digit(ch: u8) -> bool {
    (ch >= b'0' && ch <= b'9') || (ch >= b'a' && ch <= b'f') || (ch >= b'A' && ch <= b'F')
}

/// Returns `true` if the byte can start an identifier (`a-zA-Z_`).
#[inline]
fn is_ident_start(ch: u8) -> bool {
    (ch >= b'a' && ch <= b'z') || (ch >= b'A' && ch <= b'Z') || ch == b'_'
}

/// Returns `true` if the byte can continue an identifier (`a-zA-Z0-9_`).
#[inline]
fn is_ident_continue(ch: u8) -> bool {
    is_ident_start(ch) || is_digit(ch)
}

/// Remove common leading whitespace from a triple-quoted string.
///
/// The input is first trimmed of leading/trailing whitespace, then the minimum
/// indentation of non-blank lines is subtracted from every line.
fn dedent(s: &str) -> String {
    let s = s.trim();
    let lines: Vec<&str> = s.split('\n').collect();
    if lines.len() <= 1 {
        return s.to_owned();
    }

    let mut min_indent: Option<usize> = None;
    for line in &lines {
        if line.trim().is_empty() {
            continue;
        }
        let indent = line.len() - line.trim_start_matches(|c: char| c == ' ' || c == '\t').len();
        match min_indent {
            None => min_indent = Some(indent),
            Some(current) if indent < current => min_indent = Some(indent),
            _ => {}
        }
    }

    let min_indent = match min_indent {
        Some(n) if n > 0 => n,
        _ => return s.to_owned(),
    };

    let mut result: Vec<&str> = Vec::with_capacity(lines.len());
    for line in &lines {
        if line.trim().is_empty() {
            result.push("");
        } else if line.len() >= min_indent {
            result.push(&line[min_indent..]);
        } else {
            result.push(line);
        }
    }
    result.join("\n")
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;

    // -- helpers --

    /// Lex source and return all non-trivia token kinds (excluding EOF).
    fn token_kinds(src: &str) -> Vec<TokenKind> {
        let l = Lexer::new(src);
        l.all_tokens()
            .iter()
            .filter(|t| t.kind != TokenKind::EOF && !t.kind.is_trivia())
            .map(|t| t.kind)
            .collect()
    }

    /// Lex source and return all non-trivia, non-newline token values (excluding EOF).
    fn token_values(src: &str) -> Vec<String> {
        let l = Lexer::new(src);
        l.all_tokens()
            .iter()
            .filter(|t| {
                t.kind != TokenKind::EOF
                    && !t.kind.is_trivia()
                    && t.kind != TokenKind::Newline
            })
            .map(|t| t.value.clone())
            .collect()
    }

    // -- Keywords --

    #[test]
    fn keywords() {
        let cases: &[(&str, TokenKind)] = &[
            ("if", TokenKind::If),
            ("else", TokenKind::Else),
            ("for", TokenKind::For),
            ("while", TokenKind::While),
            ("return", TokenKind::Return),
            ("match", TokenKind::Match),
            ("true", TokenKind::True),
            ("false", TokenKind::False),
            ("none", TokenKind::None),
            ("some", TokenKind::Some),
            ("and", TokenKind::And),
            ("or", TokenKind::Or),
            ("not", TokenKind::Not),
            ("in", TokenKind::In),
            ("async", TokenKind::Async),
            ("spawn", TokenKind::Spawn),
            ("select", TokenKind::Select),
            ("try", TokenKind::Try),
            ("catch", TokenKind::Catch),
            ("pub", TokenKind::Pub),
            ("export", TokenKind::Export),
            ("err", TokenKind::Err),
            ("ok", TokenKind::Ok),
            ("break", TokenKind::Break),
            ("continue", TokenKind::Continue),
            ("from", TokenKind::From),
            ("default", TokenKind::Default),
            ("provider", TokenKind::Provider),
            ("tool", TokenKind::Tool),
            ("agent", TokenKind::Agent),
            ("workflow", TokenKind::Workflow),
            ("fn", TokenKind::Fn),
            ("enum", TokenKind::Enum),
            ("struct", TokenKind::Struct),
            ("type", TokenKind::Type),
            ("trait", TokenKind::Trait),
            ("impl", TokenKind::Impl),
            ("defer", TokenKind::Defer),
            ("import", TokenKind::Import),
            ("step", TokenKind::Step),
            ("orelse", TokenKind::Orelse),
            ("onerror", TokenKind::Onerror),
            ("onsuccess", TokenKind::Onsuccess),
            ("oncancel", TokenKind::Oncancel),
            ("errdefer", TokenKind::Errdefer),
            ("test", TokenKind::Test),
            ("assert", TokenKind::Assert),
            ("let", TokenKind::Let),
            ("const", TokenKind::Const),
        ];
        for &(kw, expected) in cases {
            let kinds = token_kinds(kw);
            assert_eq!(
                kinds,
                vec![expected],
                "keyword {:?}: expected {:?}, got {:?}",
                kw,
                expected,
                kinds
            );
        }
    }

    // -- Identifiers --

    #[test]
    fn identifiers() {
        let tests = ["foo", "bar_baz", "_private", "camelCase", "PascalCase", "x1", "a_b_c"];
        for id in &tests {
            let kinds = token_kinds(id);
            assert_eq!(kinds, vec![TokenKind::Ident], "identifier {:?}", id);
        }
    }

    // -- Numbers --

    #[test]
    fn int_literals() {
        let cases = [("42", "42"), ("0", "0"), ("1_000_000", "1_000_000")];
        for (input, expected_value) in &cases {
            let mut l = Lexer::new(input);
            let tok = l.next();
            assert_eq!(tok.kind, TokenKind::Int, "int {:?}", input);
            assert_eq!(tok.value, *expected_value, "int value {:?}", input);
        }
    }

    #[test]
    fn hex_literals() {
        for input in &["0xFF", "0x1A2B", "0XAB"] {
            let mut l = Lexer::new(input);
            let tok = l.next();
            assert_eq!(tok.kind, TokenKind::Int, "hex {:?}", input);
        }
    }

    #[test]
    fn binary_literals() {
        for input in &["0b1010", "0B110"] {
            let mut l = Lexer::new(input);
            let tok = l.next();
            assert_eq!(tok.kind, TokenKind::Int, "binary {:?}", input);
        }
    }

    #[test]
    fn octal_literals() {
        for input in &["0o77", "0O12"] {
            let mut l = Lexer::new(input);
            let tok = l.next();
            assert_eq!(tok.kind, TokenKind::Int, "octal {:?}", input);
        }
    }

    #[test]
    fn float_literals() {
        let cases = [("3.14", "3.14"), ("0.5", "0.5"), ("1_000.5", "1_000.5")];
        for (input, expected_value) in &cases {
            let mut l = Lexer::new(input);
            let tok = l.next();
            assert_eq!(tok.kind, TokenKind::Float, "float {:?}", input);
            assert_eq!(tok.value, *expected_value, "float value {:?}", input);
        }
    }

    // -- Strings --

    #[test]
    fn simple_string() {
        let mut l = Lexer::new(r#""hello world""#);
        let tok = l.next();
        assert_eq!(tok.kind, TokenKind::String);
        assert_eq!(tok.value, "hello world");
    }

    #[test]
    fn string_escapes() {
        let mut l = Lexer::new(r#""line\nbreak\ttab""#);
        let tok = l.next();
        assert_eq!(tok.kind, TokenKind::String);
        assert_eq!(tok.value, "line\nbreak\ttab");
    }

    #[test]
    fn interpolated_string() {
        let mut l = Lexer::new(r#""hello ${name}""#);
        let tok = l.next();
        assert_eq!(tok.kind, TokenKind::InterpolatedString);
        assert_eq!(tok.value, "hello ${name}");
    }

    #[test]
    fn non_interpolated_dollar() {
        let mut l = Lexer::new(r#""costs $5""#);
        let tok = l.next();
        assert_eq!(tok.kind, TokenKind::String, "plain String for dollar without brace");
    }

    #[test]
    fn triple_quote_string() {
        let mut l = Lexer::new(r#""""hello world""""#);
        let tok = l.next();
        assert_eq!(tok.kind, TokenKind::TripleQuoteString);
        assert_eq!(tok.value, "hello world");
    }

    #[test]
    fn triple_quote_string_multiline() {
        let input = "\"\"\"\n    line one\n    line two\n\"\"\"";
        let mut l = Lexer::new(input);
        let tok = l.next();
        assert_eq!(tok.kind, TokenKind::TripleQuoteString);
        assert_eq!(tok.value, "line one\n    line two");
    }

    #[test]
    fn triple_quote_string_dedent() {
        let input = "\"\"\"\n  hello\n  world\n\"\"\"";
        let mut l = Lexer::new(input);
        let tok = l.next();
        assert_eq!(tok.kind, TokenKind::TripleQuoteString);
        // After TrimSpace on "\n  hello\n  world\n" -> "hello\n  world"
        // minIndent from "hello"=0, so "  world" keeps spaces
        assert_eq!(tok.value, "hello\n  world");
    }

    // -- Operators --

    #[test]
    fn single_char_operators() {
        let cases: &[(&str, TokenKind)] = &[
            ("+", TokenKind::Plus),
            ("-", TokenKind::Minus),
            ("*", TokenKind::Star),
            ("/", TokenKind::Slash),
            ("%", TokenKind::Percent),
            ("<", TokenKind::Lt),
            (">", TokenKind::Gt),
            ("=", TokenKind::Eq),
            ("|", TokenKind::Pipe),
            ("?", TokenKind::Question),
            (".", TokenKind::Dot),
            (":", TokenKind::Colon),
            (",", TokenKind::Comma),
            ("@", TokenKind::At),
        ];
        for &(op, expected) in cases {
            let mut l = Lexer::new(op);
            let tok = l.next();
            assert_eq!(tok.kind, expected, "operator {:?}", op);
        }
    }

    #[test]
    fn multi_char_operators() {
        let cases: &[(&str, TokenKind)] = &[
            ("==", TokenKind::EqEq),
            ("!=", TokenKind::Ne),
            ("<=", TokenKind::Le),
            (">=", TokenKind::Ge),
            ("=>", TokenKind::FatArrow),
            ("->", TokenKind::Arrow),
            ("+=", TokenKind::PlusEq),
            ("-=", TokenKind::MinusEq),
            ("*=", TokenKind::StarEq),
            ("/=", TokenKind::SlashEq),
            ("%=", TokenKind::PercentEq),
            ("..", TokenKind::DotDot),
            ("..=", TokenKind::DotDotEq),
            ("...", TokenKind::Ellipsis),
            ("|>", TokenKind::PipeArrow),
            ("|=", TokenKind::PipeEq),
            ("&=", TokenKind::AmpEq),
            ("^=", TokenKind::CaretEq),
            ("<<", TokenKind::Shl),
            ("<<=", TokenKind::ShlEq),
            (">>", TokenKind::Shr),
            (">>=", TokenKind::ShrEq),
        ];
        for &(op, expected) in cases {
            let mut l = Lexer::new(op);
            let tok = l.next();
            assert_eq!(tok.kind, expected, "operator {:?}", op);
        }
    }

    // -- Delimiters --

    #[test]
    fn delimiters() {
        let cases: &[(&str, TokenKind)] = &[
            ("(", TokenKind::LParen),
            (")", TokenKind::RParen),
            ("{", TokenKind::LBrace),
            ("}", TokenKind::RBrace),
            ("[", TokenKind::LBracket),
            ("]", TokenKind::RBracket),
        ];
        for &(ch, expected) in cases {
            let mut l = Lexer::new(ch);
            let tok = l.next();
            assert_eq!(tok.kind, expected, "delimiter {:?}", ch);
        }
    }

    // -- Comments --

    #[test]
    fn line_comment() {
        let mut l = Lexer::new("// this is a comment\nx");
        // Line comments are trivia -- next() skips them.
        // next() also skips Newline.
        let tok = l.next();
        assert_eq!(tok.kind, TokenKind::Ident);
        assert_eq!(tok.value, "x");
    }

    #[test]
    fn line_comment_with_newlines() {
        let mut l = Lexer::new("// this is a comment\nx");
        // next_with_newlines skips trivia but keeps Newline
        let tok = l.next_with_newlines();
        assert_eq!(tok.kind, TokenKind::Newline);
        let tok = l.next_with_newlines();
        assert_eq!(tok.kind, TokenKind::Ident);
        assert_eq!(tok.value, "x");
    }

    #[test]
    fn block_comment() {
        let mut l = Lexer::new("/* block comment */ x");
        let tok = l.next();
        assert_eq!(tok.kind, TokenKind::Ident);
        assert_eq!(tok.value, "x");
    }

    #[test]
    fn nested_block_comment() {
        let mut l = Lexer::new("/* outer /* inner */ outer */ x");
        let tok = l.next();
        assert_eq!(tok.kind, TokenKind::Ident);
        assert_eq!(tok.value, "x");
    }

    // -- Newlines --

    #[test]
    fn newlines() {
        let l = Lexer::new("a\nb");
        let all = l.all_tokens();
        // Should have: Ident("a"), Newline, Ident("b"), EOF
        assert!(all.len() >= 4, "expected at least 4 tokens, got {}", all.len());
        assert_eq!(all[0].kind, TokenKind::Ident);
        assert_eq!(all[0].value, "a");
        assert_eq!(all[1].kind, TokenKind::Newline);
        assert_eq!(all[2].kind, TokenKind::Ident);
        assert_eq!(all[2].value, "b");
    }

    // -- Token positions --

    #[test]
    fn token_positions() {
        let mut l = Lexer::new("x + 42");
        let tok = l.next();
        assert_eq!((tok.start, tok.end), (0, 1), "'x' position");
        let tok = l.next();
        assert_eq!((tok.start, tok.end), (2, 3), "'+' position");
        let tok = l.next();
        assert_eq!((tok.start, tok.end), (4, 6), "'42' position");
    }

    // -- Combined expression --

    #[test]
    fn expression_tokens() {
        let kinds = token_kinds("x + 42 * y");
        assert_eq!(
            kinds,
            vec![
                TokenKind::Ident,
                TokenKind::Plus,
                TokenKind::Int,
                TokenKind::Star,
                TokenKind::Ident,
            ]
        );
    }

    // -- Function signature --

    #[test]
    fn function_signature() {
        let kinds = token_kinds("fn add(a: int, b: int) -> int");
        assert_eq!(
            kinds,
            vec![
                TokenKind::Fn,
                TokenKind::Ident,
                TokenKind::LParen,
                TokenKind::Ident,
                TokenKind::Colon,
                TokenKind::Ident,
                TokenKind::Comma,
                TokenKind::Ident,
                TokenKind::Colon,
                TokenKind::Ident,
                TokenKind::RParen,
                TokenKind::Arrow,
                TokenKind::Ident,
            ]
        );
    }

    // -- Provider block --

    #[test]
    fn provider_tokens() {
        let src = "provider openai {\n    api_key: env(\"OPENAI_KEY\")\n    model: \"gpt-4\"\n}";
        let kinds = token_kinds(src);
        assert_eq!(kinds[0], TokenKind::Provider);
        assert_eq!(kinds[1], TokenKind::Ident);
        assert_eq!(kinds[2], TokenKind::LBrace);
    }

    // -- Agent block --

    #[test]
    fn agent_tokens() {
        let src = "agent Bot {\n    model: openai\n    tools: [search]\n}";
        let kinds = token_kinds(src);
        assert_eq!(kinds[0], TokenKind::Agent);
    }

    // -- Workflow with decorator --

    #[test]
    fn workflow_with_decorator() {
        let src = "@webhook(\"/api/chat\")\nworkflow Chat(msg: string) -> string {\n    return \"ok\"\n}";
        let kinds = token_kinds(src);
        assert_eq!(kinds[0], TokenKind::At);
        assert!(
            kinds.iter().any(|k| *k == TokenKind::Workflow),
            "expected Workflow keyword in token stream"
        );
    }

    // -- EOF --

    #[test]
    fn eof_empty() {
        let mut l = Lexer::new("");
        let tok = l.next();
        assert_eq!(tok.kind, TokenKind::EOF);
    }

    #[test]
    fn eof_after_tokens() {
        let mut l = Lexer::new("x");
        l.next(); // x
        let tok = l.next();
        assert_eq!(tok.kind, TokenKind::EOF);
    }

    // -- Token categories --

    #[test]
    fn token_kind_categories() {
        assert!(TokenKind::If.is_keyword());
        assert!(TokenKind::Import.is_keyword());
        assert!(!TokenKind::Ident.is_keyword());
        assert!(TokenKind::Int.is_literal());
        assert!(TokenKind::String.is_literal());
        assert!(!TokenKind::Plus.is_literal());
        assert!(TokenKind::LineComment.is_trivia());
        assert!(TokenKind::BlockComment.is_trivia());
        assert!(!TokenKind::Newline.is_trivia());
    }

    // -- next() skips newlines --

    #[test]
    fn next_skips_newlines() {
        let mut l = Lexer::new("a\n\nb");
        let tok = l.next();
        assert_eq!(tok.value, "a");
        let tok = l.next();
        assert_eq!(tok.kind, TokenKind::Ident);
        assert_eq!(tok.value, "b");
    }

    // -- Compound assignment --

    #[test]
    fn compound_assignment() {
        let kinds = token_kinds("x += 1");
        assert_eq!(
            kinds,
            vec![TokenKind::Ident, TokenKind::PlusEq, TokenKind::Int]
        );
    }

    // -- Range --

    #[test]
    fn range_tokens() {
        let kinds = token_kinds("0..10");
        assert_eq!(
            kinds,
            vec![TokenKind::Int, TokenKind::DotDot, TokenKind::Int]
        );
    }

    #[test]
    fn inclusive_range() {
        let kinds = token_kinds("0..=10");
        assert_eq!(
            kinds,
            vec![TokenKind::Int, TokenKind::DotDotEq, TokenKind::Int]
        );
    }

    // -- Bitwise operators --

    #[test]
    fn bitwise_operators() {
        let cases: &[(&str, TokenKind)] = &[
            ("&", TokenKind::Amp),
            ("|", TokenKind::Pipe),
            ("^", TokenKind::Caret),
            ("~", TokenKind::Tilde),
        ];
        for &(op, expected) in cases {
            let mut l = Lexer::new(op);
            let tok = l.next();
            assert_eq!(tok.kind, expected, "bitwise operator {:?}", op);
        }
    }

    // -- Dedent --

    #[test]
    fn dedent_no_indent() {
        assert_eq!(dedent("hello"), "hello");
    }

    #[test]
    fn dedent_single_line() {
        assert_eq!(dedent("  hello  "), "hello");
    }

    #[test]
    fn dedent_uniform_indent() {
        // trim() strips leading spaces from first line, making its indent 0.
        // min_indent becomes 0, so nothing is stripped from subsequent lines.
        assert_eq!(dedent("  line1\n  line2"), "line1\n  line2");
    }

    #[test]
    fn dedent_mixed_indent() {
        // After trim: "line1\n    line2". line1 indent=0, so min=0, no stripping.
        assert_eq!(dedent("  line1\n    line2"), "line1\n    line2");
    }

    #[test]
    fn dedent_with_blank_lines() {
        // After trim: "line1\n\n  line2". line1 indent=0, so min=0, no stripping.
        assert_eq!(dedent("  line1\n\n  line2"), "line1\n\n  line2");
    }

    #[test]
    fn dedent_preserves_when_first_line_not_indented() {
        // When content starts with newline (like triple-quote): "\n  a\n  b"
        // trim() -> "a\n  b", line "a" indent=0, so nothing stripped.
        assert_eq!(dedent("\n  a\n  b\n"), "a\n  b");
    }

    #[test]
    fn dedent_all_indented_after_trim() {
        // Input from triple-quote where all content lines are uniformly indented
        // and the surrounding newlines get trimmed.
        // "\n    line1\n    line2\n" -> trim -> "line1\n    line2"
        // line1 has indent 0, so min=0, result unchanged.
        assert_eq!(dedent("\n    line1\n    line2\n"), "line1\n    line2");
    }

    // -- Error token --

    #[test]
    fn unknown_char_produces_error() {
        let mut l = Lexer::new("#");
        let tok = l.next();
        assert_eq!(tok.kind, TokenKind::Error);
    }

    // -- Unterminated triple-quote string --

    #[test]
    fn unterminated_triple_quote() {
        let mut l = Lexer::new("\"\"\"no closing");
        let tok = l.next();
        assert_eq!(tok.kind, TokenKind::Error);
    }

    // -- CRLF newlines --

    #[test]
    fn crlf_newline() {
        let l = Lexer::new("a\r\nb");
        let all = l.all_tokens();
        assert_eq!(all[0].kind, TokenKind::Ident);
        assert_eq!(all[0].value, "a");
        assert_eq!(all[1].kind, TokenKind::Newline);
        assert_eq!(all[1].start, 1);
        assert_eq!(all[1].end, 3); // \r\n is 2 bytes
        assert_eq!(all[2].kind, TokenKind::Ident);
        assert_eq!(all[2].value, "b");
    }

    // -- Interpolation with nested braces --

    #[test]
    fn interpolated_string_nested_braces() {
        let mut l = Lexer::new(r#""${obj.map{x}}" "#);
        let tok = l.next();
        assert_eq!(tok.kind, TokenKind::InterpolatedString);
    }

    // -- Interpolation with nested string --

    #[test]
    fn interpolated_string_nested_string() {
        let mut l = Lexer::new(r#""hello ${"world"}""#);
        let tok = l.next();
        assert_eq!(tok.kind, TokenKind::InterpolatedString);
        assert_eq!(tok.value, r#"hello ${"world"}"#);
    }

    // -- AllTokens includes trivia --

    #[test]
    fn all_tokens_includes_trivia() {
        let l = Lexer::new("x // comment\ny");
        let all = l.all_tokens();
        let has_comment = all.iter().any(|t| t.kind == TokenKind::LineComment);
        assert!(has_comment, "all_tokens should include line comments");
    }

    // -- Token values for mixed source --

    #[test]
    fn token_values_mixed() {
        let vals = token_values("fn add(a: int) -> int");
        assert_eq!(vals[0], "fn");
        assert_eq!(vals[1], "add");
        assert_eq!(vals[2], "(");
    }
}
