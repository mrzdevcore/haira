//! # Haira Lexer
//!
//! Tokenizes Haira source code into a stream of tokens.
//!
//! The lexer uses the `logos` crate for fast, zero-copy tokenization.
//!
//! ## Example
//!
//! ```
//! use haira_lexer::{Lexer, TokenKind};
//!
//! let source = "x = 42";
//! let lexer = Lexer::new(source);
//!
//! for token in lexer {
//!     println!("{:?}", token);
//! }
//! ```

mod error;
mod lexer;
mod token;

pub use error::LexError;
pub use lexer::Lexer;
pub use token::{Token, TokenKind};

/// Tokenize source code into a vector of tokens.
pub fn tokenize(source: &str) -> (Vec<Token>, Vec<LexError>) {
    let lexer = Lexer::new(source);
    let mut tokens = Vec::new();
    let mut errors = Vec::new();

    for result in lexer {
        match result {
            Ok(token) => tokens.push(token),
            Err(err) => errors.push(err),
        }
    }

    (tokens, errors)
}
