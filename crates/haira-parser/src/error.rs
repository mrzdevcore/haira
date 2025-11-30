//! Parser error definitions.

use haira_lexer::TokenKind;
use thiserror::Error;

/// A parser error.
#[derive(Debug, Clone, Error)]
pub enum ParseError {
    #[error("unexpected token: expected {expected}, found {found:?}")]
    UnexpectedToken {
        expected: String,
        found: TokenKind,
        span: std::ops::Range<usize>,
    },

    #[error("unexpected end of file")]
    UnexpectedEof {
        expected: String,
        span: std::ops::Range<usize>,
    },

    #[error("expected expression")]
    ExpectedExpr { span: std::ops::Range<usize> },

    #[error("expected statement")]
    ExpectedStatement { span: std::ops::Range<usize> },

    #[error("expected type")]
    ExpectedType { span: std::ops::Range<usize> },

    #[error("expected identifier")]
    ExpectedIdent { span: std::ops::Range<usize> },

    #[error("expected block")]
    ExpectedBlock { span: std::ops::Range<usize> },

    #[error("lexer error")]
    LexError { span: std::ops::Range<usize> },
}

impl ParseError {
    /// Get the span of this error.
    pub fn span(&self) -> std::ops::Range<usize> {
        match self {
            ParseError::UnexpectedToken { span, .. } => span.clone(),
            ParseError::UnexpectedEof { span, .. } => span.clone(),
            ParseError::ExpectedExpr { span } => span.clone(),
            ParseError::ExpectedStatement { span } => span.clone(),
            ParseError::ExpectedType { span } => span.clone(),
            ParseError::ExpectedIdent { span } => span.clone(),
            ParseError::ExpectedBlock { span } => span.clone(),
            ParseError::LexError { span } => span.clone(),
        }
    }
}
