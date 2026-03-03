//! Span, Spanned, and diagnostic types for the Haira compiler.
//!
//! This crate provides source-location tracking ([`Span`], [`Spanned<T>`]),
//! diagnostic severity levels, and pretty-printing of compiler messages
//! with source context, line numbers, and caret underlines.

use std::fmt;

// ---------------------------------------------------------------------------
// Span
// ---------------------------------------------------------------------------

/// A half-open byte range `[start, end)` in the source code.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Default)]
pub struct Span {
    /// Byte offset start (inclusive).
    pub start: usize,
    /// Byte offset end (exclusive).
    pub end: usize,
}

impl Span {
    /// Creates a new span from start to end byte offsets.
    #[inline]
    pub fn new(start: usize, end: usize) -> Self {
        Self { start, end }
    }

    /// Returns a span covering both `self` and `other`.
    #[inline]
    pub fn merge(self, other: Span) -> Span {
        Span {
            start: self.start.min(other.start),
            end: self.end.max(other.end),
        }
    }

    /// Returns the length of this span in bytes.
    #[inline]
    pub fn len(&self) -> usize {
        self.end.saturating_sub(self.start)
    }

    /// Returns true if this span has zero length.
    #[inline]
    pub fn is_empty(&self) -> bool {
        self.start >= self.end
    }
}

// ---------------------------------------------------------------------------
// Spanned<T>
// ---------------------------------------------------------------------------

/// Wraps a value with source location information.
#[derive(Debug, Clone, PartialEq)]
pub struct Spanned<T> {
    /// The wrapped value.
    pub node: T,
    /// The source span covering this value.
    pub span: Span,
}

impl<T> Spanned<T> {
    /// Creates a new `Spanned` value.
    #[inline]
    pub fn new(node: T, span: Span) -> Self {
        Self { node, span }
    }

    /// Maps the inner value while preserving the span.
    pub fn map<U>(self, f: impl FnOnce(T) -> U) -> Spanned<U> {
        Spanned {
            node: f(self.node),
            span: self.span,
        }
    }

    /// Returns a reference to the inner node with the same span.
    pub fn as_ref(&self) -> Spanned<&T> {
        Spanned {
            node: &self.node,
            span: self.span,
        }
    }
}

impl<T: Copy> Copy for Spanned<T> {}

// ---------------------------------------------------------------------------
// Level
// ---------------------------------------------------------------------------

/// Severity of a diagnostic.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Level {
    Error,
    Warning,
    Info,
}

impl fmt::Display for Level {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Level::Error => f.write_str("error"),
            Level::Warning => f.write_str("warning"),
            Level::Info => f.write_str("info"),
        }
    }
}

// ---------------------------------------------------------------------------
// Diagnostic
// ---------------------------------------------------------------------------

/// A compiler message with location information.
#[derive(Debug, Clone)]
pub struct Diagnostic {
    /// Severity level.
    pub level: Level,
    /// Human-readable message.
    pub message: String,
    /// Source span this diagnostic refers to.
    pub span: Span,
    /// Source file path (may be empty).
    pub file: String,
    /// Optional suggestion / hint.
    pub hint: String,
}

impl Diagnostic {
    /// Creates a new diagnostic with the given level, message, and span.
    pub fn new(level: Level, message: impl Into<String>, span: Span) -> Self {
        Self {
            level,
            message: message.into(),
            span,
            file: String::new(),
            hint: String::new(),
        }
    }

    /// Creates a new error diagnostic.
    pub fn error(message: impl Into<String>, span: Span) -> Self {
        Self::new(Level::Error, message, span)
    }

    /// Creates a new warning diagnostic.
    pub fn warning(message: impl Into<String>, span: Span) -> Self {
        Self::new(Level::Warning, message, span)
    }

    /// Creates a new info diagnostic.
    pub fn info(message: impl Into<String>, span: Span) -> Self {
        Self::new(Level::Info, message, span)
    }

    /// Sets the file path for this diagnostic.
    pub fn with_file(mut self, file: impl Into<String>) -> Self {
        self.file = file.into();
        self
    }

    /// Sets the hint for this diagnostic.
    pub fn with_hint(mut self, hint: impl Into<String>) -> Self {
        self.hint = hint.into();
        self
    }

    /// Render this diagnostic with source context.
    ///
    /// Produces output of the form:
    ///
    /// ```text
    /// examples/03-functions.haira:5:12: error: type mismatch
    ///   |
    /// 5 |     return "hello" + 42
    ///   |            ^^^^^^^^^^^^^^ cannot add string and int
    /// ```
    pub fn pretty_print(&self, source: &str) -> String {
        pretty_print(self, source)
    }
}

impl fmt::Display for Diagnostic {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}: {}", self.level, self.message)?;
        if !self.hint.is_empty() {
            write!(f, " (hint: {})", self.hint)?;
        }
        Ok(())
    }
}

// ---------------------------------------------------------------------------
// Free functions
// ---------------------------------------------------------------------------

/// Converts a byte offset to 1-based (line, column).
pub fn offset_to_line_col(source: &str, offset: usize) -> (usize, usize) {
    let offset = offset.min(source.len());
    let mut line: usize = 1;
    let mut col: usize = 1;
    for (i, ch) in source.bytes().enumerate() {
        if i >= offset {
            break;
        }
        if ch == b'\n' {
            line += 1;
            col = 1;
        } else {
            col += 1;
        }
    }
    (line, col)
}

/// Renders a diagnostic with source context showing the exact location.
///
/// Produces output of the form:
///
/// ```text
/// examples/03-functions.haira:5:12: error: type mismatch
///   |
/// 5 |     return "hello" + 42
///   |            ^^^^^^^^^^^^^^ cannot add string and int
/// ```
pub fn pretty_print(d: &Diagnostic, source: &str) -> String {
    let mut b = String::new();

    let (line, col) = offset_to_line_col(source, d.span.start);
    let (end_line, end_col) = offset_to_line_col(source, d.span.end);

    // Header: file:line:col: level: message
    if !d.file.is_empty() {
        use fmt::Write;
        write!(
            b,
            "{}:{}:{}: {}: {}\n",
            d.file, line, col, d.level, d.message
        )
        .unwrap();
    } else {
        use fmt::Write;
        write!(b, "{}:{}: {}: {}\n", line, col, d.level, d.message).unwrap();
    }

    // Source context
    let lines: Vec<&str> = source.split('\n').collect();
    if line <= lines.len() {
        let line_text = lines[line - 1];
        let line_num_str = line.to_string();
        let padding = " ".repeat(line_num_str.len());

        b.push_str(&padding);
        b.push_str(" |\n");

        b.push_str(&line_num_str);
        b.push_str(" | ");
        b.push_str(line_text);
        b.push('\n');

        // Underline
        let under_start = if col > 0 { col - 1 } else { 0 };
        let under_len = if end_line > line {
            // Multi-line span: underline to end of first line
            let len = line_text.len().saturating_sub(under_start);
            if len == 0 { 1 } else { len }
        } else {
            let len = end_col.saturating_sub(col);
            if len == 0 { 1 } else { len }
        };

        b.push_str(&padding);
        b.push_str(" | ");
        for _ in 0..under_start {
            b.push(' ');
        }
        for _ in 0..under_len {
            b.push('^');
        }
        b.push(' ');
        if !d.hint.is_empty() {
            b.push_str(&d.hint);
        } else {
            b.push_str(&d.message);
        }
        b.push('\n');
    }

    b
}

/// Returns `true` if any diagnostic in the slice is at [`Level::Error`].
pub fn has_errors(diags: &[Diagnostic]) -> bool {
    diags.iter().any(|d| d.level == Level::Error)
}

/// Pretty-prints all diagnostics for a given source, separated by blank lines.
pub fn format_all(diags: &[Diagnostic], source: &str) -> String {
    let mut b = String::new();
    for (i, d) in diags.iter().enumerate() {
        if i > 0 {
            b.push('\n');
        }
        b.push_str(&pretty_print(d, source));
    }
    b
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;

    // -- Span tests ----------------------------------------------------------

    #[test]
    fn span_merge() {
        let a = Span::new(5, 10);
        let b = Span::new(3, 12);
        let merged = a.merge(b);
        assert_eq!(merged, Span::new(3, 12));

        let c = Span::new(7, 8);
        let merged2 = a.merge(c);
        assert_eq!(merged2, Span::new(5, 10));
    }

    #[test]
    fn span_default() {
        let s = Span::default();
        assert_eq!(s.start, 0);
        assert_eq!(s.end, 0);
    }

    #[test]
    fn span_len_and_empty() {
        assert_eq!(Span::new(3, 7).len(), 4);
        assert!(!Span::new(3, 7).is_empty());
        assert_eq!(Span::new(5, 5).len(), 0);
        assert!(Span::new(5, 5).is_empty());
    }

    // -- Spanned tests -------------------------------------------------------

    #[test]
    fn spanned_creation() {
        let s = Spanned::new(42, Span::new(0, 2));
        assert_eq!(s.node, 42);
        assert_eq!(s.span, Span::new(0, 2));
    }

    #[test]
    fn spanned_map() {
        let s = Spanned::new(10, Span::new(0, 1));
        let s2 = s.map(|n| n * 2);
        assert_eq!(s2.node, 20);
        assert_eq!(s2.span, Span::new(0, 1));
    }

    #[test]
    fn spanned_as_ref() {
        let s = Spanned::new(String::from("hello"), Span::new(0, 5));
        let r = s.as_ref();
        assert_eq!(r.node, &"hello".to_string());
        assert_eq!(r.span, Span::new(0, 5));
    }

    // -- Level tests ---------------------------------------------------------

    #[test]
    fn level_display() {
        assert_eq!(Level::Error.to_string(), "error");
        assert_eq!(Level::Warning.to_string(), "warning");
        assert_eq!(Level::Info.to_string(), "info");
    }

    // -- Diagnostic tests ----------------------------------------------------

    #[test]
    fn diagnostic_display() {
        let d = Diagnostic::new(Level::Error, "type mismatch", Span::new(0, 5));
        assert_eq!(d.to_string(), "error: type mismatch");

        let d2 = Diagnostic::error("type mismatch", Span::new(0, 5)).with_hint("expected int");
        assert_eq!(d2.to_string(), "error: type mismatch (hint: expected int)");
    }

    #[test]
    fn diagnostic_constructors() {
        let e = Diagnostic::error("bad", Span::new(0, 1));
        assert_eq!(e.level, Level::Error);

        let w = Diagnostic::warning("unused", Span::new(0, 1));
        assert_eq!(w.level, Level::Warning);

        let i = Diagnostic::info("note", Span::new(0, 1));
        assert_eq!(i.level, Level::Info);

        let n = Diagnostic::new(Level::Warning, "custom", Span::new(2, 4));
        assert_eq!(n.level, Level::Warning);
        assert_eq!(n.message, "custom");
    }

    // -- offset_to_line_col tests --------------------------------------------

    #[test]
    fn offset_to_line_col_basic() {
        let src = "hello\nworld\n";
        assert_eq!(offset_to_line_col(src, 0), (1, 1));
        assert_eq!(offset_to_line_col(src, 5), (1, 6));
        assert_eq!(offset_to_line_col(src, 6), (2, 1));
        assert_eq!(offset_to_line_col(src, 11), (2, 6));
        // Past end clamps
        assert_eq!(offset_to_line_col(src, 100), (3, 1));
    }

    #[test]
    fn offset_to_line_col_empty() {
        assert_eq!(offset_to_line_col("", 0), (1, 1));
        assert_eq!(offset_to_line_col("", 5), (1, 1));
    }

    // -- pretty_print tests --------------------------------------------------

    #[test]
    fn pretty_print_with_file() {
        let source = "let x = 10\nlet y = \"hello\" + 42\n";
        let d = Diagnostic::error("type mismatch", Span::new(20, 34))
            .with_file("test.haira")
            .with_hint("cannot add string and int");
        let output = d.pretty_print(source);
        assert!(output.contains("test.haira:2:"));
        assert!(output.contains("error: type mismatch"));
        assert!(output.contains("cannot add string and int"));
        assert!(output.contains("^"));
    }

    #[test]
    fn pretty_print_without_file() {
        let source = "let x = 10";
        let d = Diagnostic::warning("unused variable", Span::new(4, 5));
        let output = d.pretty_print(source);
        assert!(output.starts_with("1:5: warning: unused variable\n"));
    }

    #[test]
    fn pretty_print_zero_length_span() {
        let source = "hello";
        let d = Diagnostic::error("unexpected", Span::new(3, 3));
        let output = d.pretty_print(source);
        // Should produce at least one caret even for zero-length span
        assert!(output.contains("^"));
    }

    #[test]
    fn pretty_print_method_matches_free_function() {
        let source = "let x = 1";
        let d = Diagnostic::error("bad", Span::new(0, 3));
        assert_eq!(d.pretty_print(source), pretty_print(&d, source));
    }

    // -- has_errors / format_all tests ---------------------------------------

    #[test]
    fn has_errors_check() {
        let diags = vec![
            Diagnostic::warning("unused", Span::new(0, 1)),
            Diagnostic::info("note", Span::new(0, 1)),
        ];
        assert!(!has_errors(&diags));

        let diags_with_error = vec![
            Diagnostic::warning("unused", Span::new(0, 1)),
            Diagnostic::error("bad", Span::new(0, 1)),
        ];
        assert!(has_errors(&diags_with_error));
    }

    #[test]
    fn has_errors_empty() {
        assert!(!has_errors(&[]));
    }

    #[test]
    fn format_all_multiple() {
        let source = "aaa\nbbb\n";
        let diags = vec![
            Diagnostic::error("first", Span::new(0, 3)),
            Diagnostic::warning("second", Span::new(4, 7)),
        ];
        let output = format_all(&diags, source);
        assert!(output.contains("first"));
        assert!(output.contains("second"));
        // Two diagnostics separated by a blank line
        assert!(output.contains("\n\n"));
    }

    #[test]
    fn format_all_empty() {
        assert_eq!(format_all(&[], "source"), "");
    }
}
