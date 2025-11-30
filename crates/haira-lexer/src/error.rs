//! Lexer error definitions.

use thiserror::Error;

/// A lexer error.
#[derive(Debug, Clone, Error, PartialEq, Eq)]
pub enum LexError {
    #[error("unexpected character")]
    UnexpectedChar { span: std::ops::Range<usize> },

    #[error("unterminated string literal")]
    UnterminatedString { span: std::ops::Range<usize> },

    #[error("unterminated block comment")]
    UnterminatedComment { span: std::ops::Range<usize> },

    #[error("invalid number literal")]
    InvalidNumber { span: std::ops::Range<usize> },

    #[error("invalid escape sequence")]
    InvalidEscape { span: std::ops::Range<usize> },
}

impl LexError {
    /// Get the span of this error.
    pub fn span(&self) -> std::ops::Range<usize> {
        match self {
            LexError::UnexpectedChar { span } => span.clone(),
            LexError::UnterminatedString { span } => span.clone(),
            LexError::UnterminatedComment { span } => span.clone(),
            LexError::InvalidNumber { span } => span.clone(),
            LexError::InvalidEscape { span } => span.clone(),
        }
    }
}
