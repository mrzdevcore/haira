//! High-level lexer interface.

use crate::error::LexError;
use crate::token::{Token, TokenKind};
use logos::Logos;

/// A lexer for Haira source code.
///
/// Wraps the logos-generated lexer with a nicer interface and error handling.
pub struct Lexer<'source> {
    inner: logos::Lexer<'source, TokenKind>,
    peeked: Option<Result<Token, LexError>>,
    /// Track if we've emitted EOF
    done: bool,
}

impl<'source> Lexer<'source> {
    /// Create a new lexer for the given source code.
    pub fn new(source: &'source str) -> Self {
        Self {
            inner: TokenKind::lexer(source),
            peeked: None,
            done: false,
        }
    }

    /// Peek at the next token without consuming it.
    pub fn peek(&mut self) -> Option<&Result<Token, LexError>> {
        if self.peeked.is_none() {
            self.peeked = self.next_inner();
        }
        self.peeked.as_ref()
    }

    /// Get the source text.
    pub fn source(&self) -> &'source str {
        self.inner.source()
    }

    /// Get the current byte position in the source.
    pub fn position(&self) -> usize {
        self.inner.span().start
    }

    fn next_inner(&mut self) -> Option<Result<Token, LexError>> {
        loop {
            match self.inner.next() {
                Some(Ok(kind)) => {
                    let span = self.inner.span();

                    // Skip trivia (comments)
                    if kind.is_trivia() {
                        continue;
                    }

                    return Some(Ok(Token::new(kind, span)));
                }
                Some(Err(())) => {
                    let span = self.inner.span();
                    return Some(Err(LexError::UnexpectedChar { span }));
                }
                None => {
                    if !self.done {
                        self.done = true;
                        let pos = self.inner.span().end;
                        return Some(Ok(Token::new(TokenKind::Eof, pos..pos)));
                    }
                    return None;
                }
            }
        }
    }
}

impl<'source> Iterator for Lexer<'source> {
    type Item = Result<Token, LexError>;

    fn next(&mut self) -> Option<Self::Item> {
        if let Some(peeked) = self.peeked.take() {
            return Some(peeked);
        }
        self.next_inner()
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use smol_str::SmolStr;

    #[test]
    fn test_simple_tokenization() {
        let source = "x = 42";
        let tokens: Vec<_> = Lexer::new(source).filter_map(|r| r.ok()).collect();

        assert_eq!(tokens.len(), 4); // x, =, 42, EOF
        assert_eq!(tokens[0].kind, TokenKind::Ident(SmolStr::from("x")));
        assert_eq!(tokens[1].kind, TokenKind::Eq);
        assert_eq!(tokens[2].kind, TokenKind::Int(42));
        assert_eq!(tokens[3].kind, TokenKind::Eof);
    }

    #[test]
    fn test_function_definition() {
        let source = r#"
            add(a, b) {
                a + b
            }
        "#;
        let tokens: Vec<_> = Lexer::new(source)
            .filter_map(|r| r.ok())
            .filter(|t| !matches!(t.kind, TokenKind::Newline))
            .collect();

        let kinds: Vec<_> = tokens.iter().map(|t| &t.kind).collect();
        assert!(matches!(kinds[0], TokenKind::Ident(_)));
        assert!(matches!(kinds[1], TokenKind::LParen));
        assert!(matches!(kinds[2], TokenKind::Ident(_)));
        assert!(matches!(kinds[3], TokenKind::Comma));
        assert!(matches!(kinds[4], TokenKind::Ident(_)));
        assert!(matches!(kinds[5], TokenKind::RParen));
        assert!(matches!(kinds[6], TokenKind::LBrace));
    }

    #[test]
    fn test_type_definition() {
        let source = "User { name, age, email }";
        let tokens: Vec<_> = Lexer::new(source).filter_map(|r| r.ok()).collect();

        assert_eq!(tokens[0].kind, TokenKind::Ident(SmolStr::from("User")));
        assert_eq!(tokens[1].kind, TokenKind::LBrace);
        assert_eq!(tokens[2].kind, TokenKind::Ident(SmolStr::from("name")));
        assert_eq!(tokens[3].kind, TokenKind::Comma);
    }

    #[test]
    fn test_pipe_expression() {
        let source = "users | filter_active | sort_by_name";
        let tokens: Vec<_> = Lexer::new(source).filter_map(|r| r.ok()).collect();

        assert_eq!(tokens[0].kind, TokenKind::Ident(SmolStr::from("users")));
        assert_eq!(tokens[1].kind, TokenKind::Pipe);
        assert_eq!(
            tokens[2].kind,
            TokenKind::Ident(SmolStr::from("filter_active"))
        );
        assert_eq!(tokens[3].kind, TokenKind::Pipe);
        assert_eq!(
            tokens[4].kind,
            TokenKind::Ident(SmolStr::from("sort_by_name"))
        );
    }

    #[test]
    fn test_lambda() {
        let source = "x => x * 2";
        let tokens: Vec<_> = Lexer::new(source).filter_map(|r| r.ok()).collect();

        assert_eq!(tokens[0].kind, TokenKind::Ident(SmolStr::from("x")));
        assert_eq!(tokens[1].kind, TokenKind::FatArrow);
        assert_eq!(tokens[2].kind, TokenKind::Ident(SmolStr::from("x")));
        assert_eq!(tokens[3].kind, TokenKind::Star);
        assert_eq!(tokens[4].kind, TokenKind::Int(2));
    }

    #[test]
    fn test_comments_skipped() {
        let source = r#"
            // This is a comment
            x = 42 // inline comment
            /* block
               comment */
            y = 10
        "#;
        let tokens: Vec<_> = Lexer::new(source)
            .filter_map(|r| r.ok())
            .filter(|t| !matches!(t.kind, TokenKind::Newline))
            .collect();

        let idents: Vec<_> = tokens
            .iter()
            .filter_map(|t| match &t.kind {
                TokenKind::Ident(s) => Some(s.as_str()),
                _ => None,
            })
            .collect();

        assert_eq!(idents, vec!["x", "y"]);
    }

    #[test]
    fn test_error_propagation_operator() {
        let source = "get_user(id)?";
        let tokens: Vec<_> = Lexer::new(source).filter_map(|r| r.ok()).collect();

        assert_eq!(tokens[0].kind, TokenKind::Ident(SmolStr::from("get_user")));
        assert_eq!(tokens[1].kind, TokenKind::LParen);
        assert_eq!(tokens[2].kind, TokenKind::Ident(SmolStr::from("id")));
        assert_eq!(tokens[3].kind, TokenKind::RParen);
        assert_eq!(tokens[4].kind, TokenKind::Question);
    }

    #[test]
    fn test_range_operators() {
        let source = "0..10 0..=10";
        let tokens: Vec<_> = Lexer::new(source).filter_map(|r| r.ok()).collect();

        assert_eq!(tokens[0].kind, TokenKind::Int(0));
        assert_eq!(tokens[1].kind, TokenKind::DotDot);
        assert_eq!(tokens[2].kind, TokenKind::Int(10));
        assert_eq!(tokens[3].kind, TokenKind::Int(0));
        assert_eq!(tokens[4].kind, TokenKind::DotDotEq);
        assert_eq!(tokens[5].kind, TokenKind::Int(10));
    }
}
