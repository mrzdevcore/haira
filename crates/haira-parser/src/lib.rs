//! # Haira Parser
//!
//! Parses Haira source code into an Abstract Syntax Tree.
//!
//! Uses recursive descent with Pratt parsing for expressions.
//!
//! ## Example
//!
//! ```
//! use haira_parser::parse;
//!
//! let source = r#"
//!     User { name, age }
//!
//!     greet(user) {
//!         "Hello, " + user.name
//!     }
//! "#;
//!
//! let result = parse(source);
//! assert!(result.errors.is_empty());
//! ```

mod parser;
mod error;

pub use parser::Parser;
pub use error::ParseError;

use haira_ast::SourceFile;

/// Result of parsing.
pub struct ParseResult {
    /// The parsed AST (may be partial if errors occurred)
    pub ast: SourceFile,
    /// Any errors encountered during parsing
    pub errors: Vec<ParseError>,
}

/// Parse source code into an AST.
pub fn parse(source: &str) -> ParseResult {
    let mut parser = Parser::new(source);
    let ast = parser.parse_source_file();
    ParseResult {
        ast,
        errors: parser.into_errors(),
    }
}
