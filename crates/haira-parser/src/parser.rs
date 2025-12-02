//! Recursive descent parser for Haira.

use haira_ast::*;
use haira_lexer::{Lexer, Token, TokenKind};
use smol_str::SmolStr;

use crate::error::ParseError;

/// Operator precedence levels for Pratt parsing.
#[derive(Debug, Clone, Copy, PartialEq, Eq, PartialOrd, Ord)]
enum Precedence {
    None,
    Pipe,       // |
    Or,         // or
    And,        // and
    Equality,   // == !=
    Comparison, // < > <= >=
    Term,       // + -
    Factor,     // * / %
    Unary,      // - not
    Call,       // () [] .
}

impl Precedence {
    fn of(kind: &TokenKind) -> Self {
        match kind {
            // Note: TokenKind::Eq (assignment) is NOT included here because
            // assignment is handled at the statement level, not as an expression.
            // Including it here would cause an infinite loop in parse_expr_precedence.
            TokenKind::Pipe => Precedence::Pipe,
            TokenKind::Or => Precedence::Or,
            TokenKind::And => Precedence::And,
            TokenKind::EqEq | TokenKind::Ne => Precedence::Equality,
            TokenKind::Lt | TokenKind::Gt | TokenKind::Le | TokenKind::Ge => Precedence::Comparison,
            TokenKind::Plus | TokenKind::Minus => Precedence::Term,
            TokenKind::Star | TokenKind::Slash | TokenKind::Percent => Precedence::Factor,
            TokenKind::LParen | TokenKind::LBracket | TokenKind::Dot | TokenKind::Question => {
                Precedence::Call
            }
            TokenKind::DotDot | TokenKind::DotDotEq => Precedence::Comparison,
            _ => Precedence::None,
        }
    }
}

/// Parser for Haira source code.
pub struct Parser<'source> {
    lexer: Lexer<'source>,
    current: Token,
    previous: Token,
    errors: Vec<ParseError>,
}

impl<'source> Parser<'source> {
    /// Create a new parser for the given source.
    pub fn new(source: &'source str) -> Self {
        let mut lexer = Lexer::new(source);

        // Get the first non-newline token
        let current = Self::next_significant_token(&mut lexer);

        Self {
            lexer,
            current,
            previous: Token::new(TokenKind::Eof, 0..0),
            errors: Vec::new(),
        }
    }

    /// Get the collected errors.
    pub fn into_errors(self) -> Vec<ParseError> {
        self.errors
    }

    fn next_significant_token(lexer: &mut Lexer) -> Token {
        loop {
            match lexer.next() {
                Some(Ok(token)) => {
                    // Skip whitespace, newlines, and comments
                    if !matches!(
                        token.kind,
                        TokenKind::Newline | TokenKind::LineComment | TokenKind::BlockComment
                    ) {
                        return token;
                    }
                }
                Some(Err(_)) => {
                    // Skip errors, they'll be reported elsewhere
                    continue;
                }
                None => {
                    return Token::new(TokenKind::Eof, 0..0);
                }
            }
        }
    }

    fn advance(&mut self) {
        self.previous = std::mem::replace(
            &mut self.current,
            Self::next_significant_token(&mut self.lexer),
        );
    }

    fn skip_newlines(&mut self) {
        while matches!(self.current.kind, TokenKind::Newline) {
            self.advance();
        }
    }

    fn check(&self, kind: &TokenKind) -> bool {
        std::mem::discriminant(&self.current.kind) == std::mem::discriminant(kind)
    }

    fn at_end(&self) -> bool {
        matches!(self.current.kind, TokenKind::Eof)
    }

    fn consume(&mut self, kind: TokenKind, expected: &str) -> bool {
        if self.check(&kind) {
            self.advance();
            true
        } else {
            self.error(ParseError::UnexpectedToken {
                expected: expected.to_string(),
                found: self.current.kind.clone(),
                span: self.current.span.clone(),
            });
            false
        }
    }

    fn error(&mut self, err: ParseError) {
        self.errors.push(err);
    }

    fn span(&self, start: usize) -> Span {
        Span::new(start as u32, self.previous.span.end as u32)
    }

    fn current_span(&self) -> Span {
        Span::new(self.current.span.start as u32, self.current.span.end as u32)
    }

    // ========================================================================
    // Top-level parsing
    // ========================================================================

    /// Parse a complete source file.
    pub fn parse_source_file(&mut self) -> SourceFile {
        let start = self.current.span.start;
        let mut items = Vec::new();

        while !self.at_end() {
            self.skip_newlines();
            if self.at_end() {
                break;
            }

            if let Some(item) = self.parse_item() {
                items.push(item);
            } else {
                // Error recovery: skip to next line
                self.advance();
            }
        }

        SourceFile {
            items,
            span: self.span(start),
        }
    }

    fn parse_item(&mut self) -> Option<Item> {
        let start = self.current.span.start;

        // Check for `public` modifier
        let is_public = if matches!(self.current.kind, TokenKind::Public) {
            self.advance();
            true
        } else {
            false
        };

        match &self.current.kind {
            // Type definition or function/method
            TokenKind::Ident(_) => {
                let name = self.parse_identifier()?;

                match &self.current.kind {
                    // Type definition: `User { ... }`
                    TokenKind::LBrace => {
                        let type_def = self.parse_type_def_body(is_public, name)?;
                        Some(Spanned::new(ItemKind::TypeDef(type_def), self.span(start)))
                    }
                    // Function definition: `foo(...) { ... }`
                    // or expression statement: `foo(...)`
                    TokenKind::LParen => {
                        // We need to look ahead to determine if this is a function definition or a call.
                        // Function definitions have a block after the params: `foo(x, y) { ... }`
                        // Function calls are just expressions: `foo(arg1, arg2)`
                        //
                        // The key difference: function definitions require a `{` after `)`,
                        // while function calls end with `)`.
                        //
                        // We'll parse the parens and then check what follows.
                        let expr = Spanned::new(ExprKind::Identifier(name.node.clone()), name.span);

                        // Try to parse it as a call expression
                        let call_expr = self.parse_infix(expr, Precedence::None)?;

                        // Check if there's a block following (which would indicate a function def)
                        if self.check(&TokenKind::LBrace) || self.check(&TokenKind::Arrow) {
                            // This is a function definition
                            // Extract parameters from the call expr
                            if let ExprKind::Call(call) = &call_expr.node {
                                // Convert arguments back to parameters
                                let params = self.args_to_params(&call.args)?;

                                let return_ty = if self.check(&TokenKind::Arrow) {
                                    self.advance();
                                    Some(self.parse_type()?)
                                } else {
                                    None
                                };

                                let body = self.parse_block()?;

                                Some(Spanned::new(
                                    ItemKind::FunctionDef(FunctionDef {
                                        is_public,
                                        name,
                                        params,
                                        return_ty,
                                        body,
                                    }),
                                    self.span(start),
                                ))
                            } else {
                                // Not a call expression, can't be a function def
                                self.error(ParseError::ExpectedStatement {
                                    span: call_expr.span.start as usize
                                        ..call_expr.span.end as usize,
                                });
                                None
                            }
                        } else {
                            // This is an expression statement (function call)
                            let stmt =
                                Spanned::new(StatementKind::Expr(call_expr), self.span(start));
                            Some(Spanned::new(ItemKind::Statement(stmt), self.span(start)))
                        }
                    }
                    // Method definition: `Type.method(...) { ... }`
                    TokenKind::Dot => {
                        self.advance(); // consume .
                        let method_name = self.parse_identifier()?;
                        let method = self.parse_method_def_body(name, method_name)?;
                        Some(Spanned::new(ItemKind::MethodDef(method), self.span(start)))
                    }
                    // Type alias: `UserId = int` OR assignment: `x = 10`
                    // Type aliases have uppercase first letter, assignments have lowercase
                    TokenKind::Eq => {
                        let first_char = name.node.chars().next().unwrap_or('a');
                        if first_char.is_uppercase() {
                            // Type alias
                            self.advance(); // consume =
                            let ty = self.parse_type()?;
                            Some(Spanned::new(
                                ItemKind::TypeAlias(TypeAlias { name, ty }),
                                self.span(start),
                            ))
                        } else {
                            // Variable assignment - parse as statement
                            let expr =
                                Spanned::new(ExprKind::Identifier(name.node.clone()), name.span);
                            let stmt = self.parse_statement_rest(expr)?;
                            Some(Spanned::new(ItemKind::Statement(stmt), self.span(start)))
                        }
                    }
                    // Otherwise it's a statement starting with an identifier
                    _ => {
                        // Put the name back as an expression
                        let expr = Spanned::new(ExprKind::Identifier(name.node.clone()), name.span);
                        let stmt = self.parse_statement_rest(expr)?;
                        Some(Spanned::new(ItemKind::Statement(stmt), self.span(start)))
                    }
                }
            }
            // AI-generated function definition: `ai func_name(params) -> Type { intent }`
            TokenKind::Ai => {
                self.advance();
                let ai_block = self.parse_ai_block()?;
                Some(Spanned::new(
                    ItemKind::AiFunctionDef(ai_block),
                    self.span(start),
                ))
            }
            // Keywords that start statements
            TokenKind::If
            | TokenKind::For
            | TokenKind::While
            | TokenKind::Return
            | TokenKind::Match
            | TokenKind::Try
            | TokenKind::Break
            | TokenKind::Continue
            | TokenKind::Spawn
            | TokenKind::Async => {
                let stmt = self.parse_statement()?;
                let span = stmt.span.clone();
                Some(Spanned::new(ItemKind::Statement(stmt), span))
            }
            _ => {
                self.error(ParseError::ExpectedStatement {
                    span: self.current.span.clone(),
                });
                None
            }
        }
    }

    // ========================================================================
    // Type definitions
    // ========================================================================

    fn parse_type_def_body(&mut self, is_public: bool, name: Spanned<SmolStr>) -> Option<TypeDef> {
        self.consume(TokenKind::LBrace, "{");
        self.skip_newlines();

        let mut fields = Vec::new();

        while !self.check(&TokenKind::RBrace) && !self.at_end() {
            if let Some(field) = self.parse_field() {
                fields.push(field);
            }

            // Expect comma or newline between fields
            if self.check(&TokenKind::Comma) {
                self.advance();
            }
            self.skip_newlines();
        }

        self.consume(TokenKind::RBrace, "}");

        Some(TypeDef {
            is_public,
            name,
            fields,
        })
    }

    fn parse_field(&mut self) -> Option<Field> {
        let start = self.current.span.start;
        let name = self.parse_identifier()?;

        let ty = if self.check(&TokenKind::Colon) {
            self.advance();
            Some(self.parse_type()?)
        } else {
            None
        };

        let default = if self.check(&TokenKind::Eq) {
            self.advance();
            Some(self.parse_expr()?)
        } else {
            None
        };

        Some(Field {
            name,
            ty,
            default,
            span: self.span(start),
        })
    }

    // ========================================================================
    // Function definitions
    // ========================================================================

    fn parse_method_def_body(
        &mut self,
        type_name: Spanned<SmolStr>,
        name: Spanned<SmolStr>,
    ) -> Option<MethodDef> {
        let params = self.parse_params()?;

        let return_ty = if self.check(&TokenKind::Arrow) {
            self.advance();
            Some(self.parse_type()?)
        } else {
            None
        };

        let body = self.parse_block()?;

        Some(MethodDef {
            type_name,
            name,
            params,
            return_ty,
            body,
        })
    }

    fn parse_params(&mut self) -> Option<Vec<Param>> {
        self.consume(TokenKind::LParen, "(");

        let mut params = Vec::new();

        while !self.check(&TokenKind::RParen) && !self.at_end() {
            if let Some(param) = self.parse_param() {
                params.push(param);
            }

            if !self.check(&TokenKind::RParen) {
                self.consume(TokenKind::Comma, ",");
            }
        }

        self.consume(TokenKind::RParen, ")");

        Some(params)
    }

    /// Convert call arguments back to function parameters.
    /// Used when we realize a "call" was actually a function definition.
    fn args_to_params(&mut self, args: &[Argument]) -> Option<Vec<Param>> {
        let mut params = Vec::new();

        for arg in args {
            // Each argument should be a simple identifier (parameter name)
            // or a named argument with default value
            match &arg.value.node {
                ExprKind::Identifier(name) => {
                    params.push(Param {
                        name: Spanned::new(name.clone(), arg.value.span.clone()),
                        ty: None,
                        default: None,
                        is_rest: false,
                        span: arg.span.clone(),
                    });
                }
                _ => {
                    // If arg has a name, it's `name = default_value`
                    if let Some(param_name) = &arg.name {
                        params.push(Param {
                            name: param_name.clone(),
                            ty: None,
                            default: Some(arg.value.clone()),
                            is_rest: false,
                            span: arg.span.clone(),
                        });
                    } else {
                        self.error(ParseError::ExpectedIdent {
                            span: arg.value.span.start as usize..arg.value.span.end as usize,
                        });
                        return None;
                    }
                }
            }
        }

        Some(params)
    }

    fn parse_param(&mut self) -> Option<Param> {
        let start = self.current.span.start;
        let name = self.parse_identifier()?;

        let ty = if self.check(&TokenKind::Colon) {
            self.advance();
            Some(self.parse_type()?)
        } else {
            None
        };

        let default = if self.check(&TokenKind::Eq) {
            self.advance();
            Some(self.parse_expr()?)
        } else {
            None
        };

        let is_rest = if self.check(&TokenKind::Ellipsis) {
            self.advance();
            true
        } else {
            false
        };

        Some(Param {
            name,
            ty,
            default,
            is_rest,
            span: self.span(start),
        })
    }

    // ========================================================================
    // Types
    // ========================================================================

    fn parse_type(&mut self) -> Option<Spanned<Type>> {
        let start = self.current.span.start;

        let ty = match &self.current.kind {
            // Simple named type
            TokenKind::Ident(name) => {
                let name = name.clone();
                self.advance();

                // Check for generic args: `Box<T>`
                if self.check(&TokenKind::Lt) {
                    self.advance();
                    let mut args = Vec::new();
                    while !self.check(&TokenKind::Gt) && !self.at_end() {
                        args.push(self.parse_type()?);
                        if !self.check(&TokenKind::Gt) {
                            self.consume(TokenKind::Comma, ",");
                        }
                    }
                    self.consume(TokenKind::Gt, ">");
                    Type::Generic { name, args }
                } else {
                    Type::Named(name)
                }
            }
            // List type: `[int]`
            TokenKind::LBracket => {
                self.advance();
                let inner = self.parse_type()?;
                self.consume(TokenKind::RBracket, "]");
                Type::List(Box::new(inner))
            }
            // Map type: `{string: int}`
            TokenKind::LBrace => {
                self.advance();
                let key = self.parse_type()?;
                self.consume(TokenKind::Colon, ":");
                let value = self.parse_type()?;
                self.consume(TokenKind::RBrace, "}");
                Type::Map {
                    key: Box::new(key),
                    value: Box::new(value),
                }
            }
            // Function type: `(int, int) -> int`
            TokenKind::LParen => {
                self.advance();
                let mut params = Vec::new();
                while !self.check(&TokenKind::RParen) && !self.at_end() {
                    params.push(self.parse_type()?);
                    if !self.check(&TokenKind::RParen) {
                        self.consume(TokenKind::Comma, ",");
                    }
                }
                self.consume(TokenKind::RParen, ")");
                self.consume(TokenKind::Arrow, "->");
                let ret = self.parse_type()?;
                Type::Function {
                    params,
                    ret: Box::new(ret),
                }
            }
            _ => {
                self.error(ParseError::ExpectedType {
                    span: self.current.span.clone(),
                });
                return None;
            }
        };

        // Check for union: `Type | Other`
        if self.check(&TokenKind::Pipe) {
            let mut variants = vec![Spanned::new(ty, self.span(start))];
            while self.check(&TokenKind::Pipe) {
                self.advance();
                variants.push(self.parse_type()?);
            }
            return Some(Spanned::new(Type::Union(variants), self.span(start)));
        }

        Some(Spanned::new(ty, self.span(start)))
    }

    // ========================================================================
    // Statements
    // ========================================================================

    fn parse_statement(&mut self) -> Option<Statement> {
        let start = self.current.span.start;

        let kind = match &self.current.kind {
            TokenKind::If => {
                self.advance();
                StatementKind::If(self.parse_if_statement()?)
            }
            TokenKind::For => {
                self.advance();
                StatementKind::For(self.parse_for_statement()?)
            }
            TokenKind::While => {
                self.advance();
                StatementKind::While(self.parse_while_statement()?)
            }
            TokenKind::Return => {
                self.advance();
                StatementKind::Return(self.parse_return_statement()?)
            }
            TokenKind::Match => {
                self.advance();
                StatementKind::Match(self.parse_match_expr()?)
            }
            TokenKind::Try => {
                self.advance();
                StatementKind::Try(self.parse_try_statement()?)
            }
            TokenKind::Break => {
                self.advance();
                StatementKind::Break
            }
            TokenKind::Continue => {
                self.advance();
                StatementKind::Continue
            }
            TokenKind::Spawn | TokenKind::Async => {
                // Spawn and Async are parsed as expressions
                let expr = self.parse_expr()?;
                return self.parse_statement_rest(expr);
            }
            _ => {
                let expr = self.parse_expr()?;
                return self.parse_statement_rest(expr);
            }
        };

        Some(Spanned::new(kind, self.span(start)))
    }

    fn parse_statement_rest(&mut self, first_expr: Expr) -> Option<Statement> {
        let start = first_expr.span.start as usize;

        // Check for assignment
        if self.check(&TokenKind::Eq) {
            self.advance();
            let value = self.parse_expr()?;

            // Convert expression to assignment target
            let target = self.expr_to_assign_target(first_expr)?;
            let targets = vec![target];

            return Some(Spanned::new(
                StatementKind::Assignment(Assignment { targets, value }),
                self.span(start),
            ));
        }

        // Check for multi-assignment: `a, b = ...`
        if self.check(&TokenKind::Comma) {
            let mut targets = vec![self.expr_to_assign_target(first_expr)?];

            while self.check(&TokenKind::Comma) {
                self.advance();
                let expr = self.parse_expr()?;
                targets.push(self.expr_to_assign_target(expr)?);
            }

            self.consume(TokenKind::Eq, "=");
            let value = self.parse_expr()?;

            return Some(Spanned::new(
                StatementKind::Assignment(Assignment { targets, value }),
                self.span(start),
            ));
        }

        // Otherwise it's an expression statement
        Some(Spanned::new(
            StatementKind::Expr(first_expr),
            self.span(start),
        ))
    }

    fn expr_to_assign_target(&mut self, expr: Expr) -> Option<AssignTarget> {
        match expr.node {
            ExprKind::Identifier(name) => Some(AssignTarget {
                name: Spanned::new(name, expr.span),
                ty: None,
            }),
            _ => {
                self.error(ParseError::ExpectedIdent {
                    span: expr.span.start as usize..expr.span.end as usize,
                });
                None
            }
        }
    }

    fn parse_if_statement(&mut self) -> Option<IfStatement> {
        let condition = self.parse_expr()?;
        let then_branch = self.parse_block()?;

        let else_branch = if self.check(&TokenKind::Else) {
            self.advance();
            if self.check(&TokenKind::If) {
                self.advance();
                let start = self.previous.span.start;
                let if_stmt = self.parse_if_statement()?;
                Some(ElseBranch::ElseIf(Box::new(Spanned::new(
                    if_stmt,
                    self.span(start),
                ))))
            } else {
                Some(ElseBranch::Block(self.parse_block()?))
            }
        } else {
            None
        };

        Some(IfStatement {
            condition,
            then_branch,
            else_branch,
        })
    }

    fn parse_for_statement(&mut self) -> Option<ForStatement> {
        let pattern = self.parse_for_pattern()?;
        self.consume(TokenKind::In, "in");
        let iterator = self.parse_expr()?;
        let body = self.parse_block()?;

        Some(ForStatement {
            pattern,
            iterator,
            body,
        })
    }

    fn parse_for_pattern(&mut self) -> Option<ForPattern> {
        let first = self.parse_identifier()?;

        if self.check(&TokenKind::Comma) {
            self.advance();
            let second = self.parse_identifier()?;
            Some(ForPattern::Pair(first, second))
        } else {
            Some(ForPattern::Single(first))
        }
    }

    fn parse_while_statement(&mut self) -> Option<WhileStatement> {
        let condition = self.parse_expr()?;
        let body = self.parse_block()?;

        Some(WhileStatement { condition, body })
    }

    fn parse_return_statement(&mut self) -> Option<ReturnStatement> {
        // Check if there are values to return
        if self.check(&TokenKind::Newline) || self.check(&TokenKind::RBrace) || self.at_end() {
            return Some(ReturnStatement { values: Vec::new() });
        }

        let mut values = vec![self.parse_expr()?];

        while self.check(&TokenKind::Comma) {
            self.advance();
            values.push(self.parse_expr()?);
        }

        Some(ReturnStatement { values })
    }

    fn parse_try_statement(&mut self) -> Option<TryStatement> {
        let body = self.parse_block()?;
        self.consume(TokenKind::Catch, "catch");
        let error_name = self.parse_identifier()?;
        let catch_body = self.parse_block()?;

        Some(TryStatement {
            body,
            error_name,
            catch_body,
        })
    }

    // ========================================================================
    // Blocks
    // ========================================================================

    fn parse_block(&mut self) -> Option<Block> {
        let start = self.current.span.start;

        self.consume(TokenKind::LBrace, "{");
        self.skip_newlines();

        let mut statements = Vec::new();

        while !self.check(&TokenKind::RBrace) && !self.at_end() {
            if let Some(stmt) = self.parse_statement() {
                statements.push(stmt);
            } else {
                // Error recovery
                self.advance();
            }
            self.skip_newlines();
        }

        self.consume(TokenKind::RBrace, "}");

        Some(Block {
            statements,
            span: self.span(start),
        })
    }

    // ========================================================================
    // Expressions (Pratt parser)
    // ========================================================================

    fn parse_expr(&mut self) -> Option<Expr> {
        self.parse_expr_precedence(Precedence::None)
    }

    fn parse_expr_precedence(&mut self, min_prec: Precedence) -> Option<Expr> {
        let mut left = self.parse_prefix()?;

        while !self.at_end() {
            let prec = Precedence::of(&self.current.kind);
            if prec <= min_prec {
                break;
            }

            left = self.parse_infix(left, prec)?;
        }

        Some(left)
    }

    fn parse_prefix(&mut self) -> Option<Expr> {
        let start = self.current.span.start;

        match &self.current.kind {
            // Literals
            TokenKind::Int(n) => {
                let n = *n;
                self.advance();
                Some(Spanned::new(
                    ExprKind::Literal(Literal::Int(n)),
                    self.span(start),
                ))
            }
            TokenKind::Float(n) => {
                let n = *n;
                self.advance();
                Some(Spanned::new(
                    ExprKind::Literal(Literal::Float(n)),
                    self.span(start),
                ))
            }
            TokenKind::String(s) => {
                let s = s.clone();
                self.advance();
                Some(Spanned::new(
                    ExprKind::Literal(Literal::String(s)),
                    self.span(start),
                ))
            }
            TokenKind::InterpolatedString(s) => {
                let s = s.clone();
                self.advance();
                let parts = self.parse_interpolated_string_parts(&s)?;
                Some(Spanned::new(
                    ExprKind::Literal(Literal::InterpolatedString(parts)),
                    self.span(start),
                ))
            }
            TokenKind::True => {
                self.advance();
                Some(Spanned::new(
                    ExprKind::Literal(Literal::Bool(true)),
                    self.span(start),
                ))
            }
            TokenKind::False => {
                self.advance();
                Some(Spanned::new(
                    ExprKind::Literal(Literal::Bool(false)),
                    self.span(start),
                ))
            }
            TokenKind::None => {
                self.advance();
                Some(Spanned::new(ExprKind::None, self.span(start)))
            }
            TokenKind::Some => {
                self.advance();
                self.consume(TokenKind::LParen, "(");
                let inner = self.parse_expr()?;
                self.consume(TokenKind::RParen, ")");
                Some(Spanned::new(
                    ExprKind::Some(Box::new(inner)),
                    self.span(start),
                ))
            }

            // Identifier (might be lambda: `x => ...`)
            TokenKind::Ident(name) => {
                let name = name.clone();
                self.advance();

                // Check for arrow lambda: `x => expr`
                if self.check(&TokenKind::FatArrow) {
                    self.advance();
                    let body = self.parse_expr()?;
                    return Some(Spanned::new(
                        ExprKind::Lambda(LambdaExpr {
                            params: vec![Param {
                                name: Spanned::new(name, self.span(start)),
                                ty: None,
                                default: None,
                                is_rest: false,
                                span: self.span(start),
                            }],
                            body: LambdaBody::Expr(Box::new(body)),
                        }),
                        self.span(start),
                    ));
                }

                // Check for type instantiation: `User { ... }`
                // Only treat as instance if the name starts with uppercase (type name convention)
                let first_char = name.chars().next().unwrap_or('a');
                if first_char.is_uppercase() && self.check(&TokenKind::LBrace) {
                    return self.parse_instance(name, start);
                }

                Some(Spanned::new(ExprKind::Identifier(name), self.span(start)))
            }

            // Unary operators
            TokenKind::Minus => {
                self.advance();
                let operand = self.parse_expr_precedence(Precedence::Unary)?;
                Some(Spanned::new(
                    ExprKind::Unary(UnaryExpr {
                        op: Spanned::new(UnaryOp::Neg, self.span(start)),
                        operand: Box::new(operand),
                    }),
                    self.span(start),
                ))
            }
            TokenKind::Not => {
                self.advance();
                let operand = self.parse_expr_precedence(Precedence::Unary)?;
                Some(Spanned::new(
                    ExprKind::Unary(UnaryExpr {
                        op: Spanned::new(UnaryOp::Not, self.span(start)),
                        operand: Box::new(operand),
                    }),
                    self.span(start),
                ))
            }

            // Grouping or lambda: `(...)` or `(x, y) { ... }` or `(x, y) => ...`
            TokenKind::LParen => self.parse_paren_or_lambda(start),

            // List: `[1, 2, 3]`
            TokenKind::LBracket => self.parse_list(start),

            // Map: `{ "a": 1, "b": 2 }`
            TokenKind::LBrace => self.parse_map_or_block(start),

            // If expression
            TokenKind::If => {
                self.advance();
                let if_stmt = self.parse_if_statement()?;
                Some(Spanned::new(
                    ExprKind::If(Box::new(if_stmt)),
                    self.span(start),
                ))
            }

            // Match expression
            TokenKind::Match => {
                self.advance();
                let match_expr = self.parse_match_expr()?;
                Some(Spanned::new(ExprKind::Match(match_expr), self.span(start)))
            }

            // Async block
            TokenKind::Async => {
                self.advance();
                let block = self.parse_block()?;
                Some(Spanned::new(ExprKind::Async(block), self.span(start)))
            }

            // Spawn block
            TokenKind::Spawn => {
                self.advance();
                let block = self.parse_block()?;
                Some(Spanned::new(ExprKind::Spawn(block), self.span(start)))
            }

            // Select expression
            TokenKind::Select => {
                self.advance();
                let select = self.parse_select_expr()?;
                Some(Spanned::new(ExprKind::Select(select), self.span(start)))
            }

            // AI block expression: `ai(params) -> Type { intent }` or `ai func_name(params) { intent }`
            TokenKind::Ai => {
                self.advance();
                let ai_block = self.parse_ai_block()?;
                Some(Spanned::new(ExprKind::Ai(ai_block), self.span(start)))
            }

            // err(...) expression - treat as a call
            TokenKind::Err => {
                self.advance();
                // Create identifier "err"
                let callee =
                    Spanned::new(ExprKind::Identifier(SmolStr::from("err")), self.span(start));

                // Parse arguments if present
                if self.check(&TokenKind::LParen) {
                    self.advance();
                    let mut args = Vec::new();

                    if !self.check(&TokenKind::RParen) {
                        loop {
                            let arg_start = self.current.span.start;
                            let value = self.parse_expr()?;
                            args.push(Argument {
                                name: None,
                                value,
                                span: self.span(arg_start),
                            });
                            if !self.check(&TokenKind::Comma) {
                                break;
                            }
                            self.advance();
                        }
                    }
                    self.consume(TokenKind::RParen, ")");

                    Some(Spanned::new(
                        ExprKind::Call(CallExpr {
                            callee: Box::new(callee),
                            args,
                        }),
                        self.span(start),
                    ))
                } else {
                    // Just `err` without parens - treat as call with no args
                    Some(Spanned::new(
                        ExprKind::Call(CallExpr {
                            callee: Box::new(callee),
                            args: vec![],
                        }),
                        self.span(start),
                    ))
                }
            }

            _ => {
                self.error(ParseError::ExpectedExpr {
                    span: self.current.span.clone(),
                });
                None
            }
        }
    }

    fn parse_infix(&mut self, left: Expr, prec: Precedence) -> Option<Expr> {
        let start = left.span.start as usize;
        let op_span = self.current_span();

        match &self.current.kind {
            // Binary operators
            TokenKind::Plus
            | TokenKind::Minus
            | TokenKind::Star
            | TokenKind::Slash
            | TokenKind::Percent
            | TokenKind::EqEq
            | TokenKind::Ne
            | TokenKind::Lt
            | TokenKind::Gt
            | TokenKind::Le
            | TokenKind::Ge
            | TokenKind::And
            | TokenKind::Or => {
                let op = self.parse_binary_op()?;
                let right = self.parse_expr_precedence(prec)?;
                Some(Spanned::new(
                    ExprKind::Binary(BinaryExpr {
                        left: Box::new(left),
                        op: Spanned::new(op, op_span),
                        right: Box::new(right),
                    }),
                    self.span(start),
                ))
            }

            // Pipe
            TokenKind::Pipe => {
                self.advance();
                let right = self.parse_expr_precedence(prec)?;
                Some(Spanned::new(
                    ExprKind::Pipe(PipeExpr {
                        left: Box::new(left),
                        right: Box::new(right),
                    }),
                    self.span(start),
                ))
            }

            // Range
            TokenKind::DotDot => {
                self.advance();
                let end = self.parse_expr_precedence(prec)?;
                Some(Spanned::new(
                    ExprKind::Range(RangeExpr {
                        start: Box::new(left),
                        end: Box::new(end),
                        inclusive: false,
                    }),
                    self.span(start),
                ))
            }
            TokenKind::DotDotEq => {
                self.advance();
                let end = self.parse_expr_precedence(prec)?;
                Some(Spanned::new(
                    ExprKind::Range(RangeExpr {
                        start: Box::new(left),
                        end: Box::new(end),
                        inclusive: true,
                    }),
                    self.span(start),
                ))
            }

            // Call
            TokenKind::LParen => {
                let args = self.parse_call_args()?;
                Some(Spanned::new(
                    ExprKind::Call(CallExpr {
                        callee: Box::new(left),
                        args,
                    }),
                    self.span(start),
                ))
            }

            // Index
            TokenKind::LBracket => {
                self.advance();
                let index = self.parse_expr()?;
                self.consume(TokenKind::RBracket, "]");
                Some(Spanned::new(
                    ExprKind::Index(IndexExpr {
                        object: Box::new(left),
                        index: Box::new(index),
                    }),
                    self.span(start),
                ))
            }

            // Field access or method call
            TokenKind::Dot => {
                self.advance();
                let field = self.parse_identifier()?;

                // Check for method call
                if self.check(&TokenKind::LParen) {
                    let args = self.parse_call_args()?;
                    Some(Spanned::new(
                        ExprKind::MethodCall(MethodCallExpr {
                            receiver: Box::new(left),
                            method: field,
                            args,
                        }),
                        self.span(start),
                    ))
                } else {
                    Some(Spanned::new(
                        ExprKind::Field(FieldExpr {
                            object: Box::new(left),
                            field,
                        }),
                        self.span(start),
                    ))
                }
            }

            // Error propagation
            TokenKind::Question => {
                self.advance();
                Some(Spanned::new(
                    ExprKind::Propagate(Box::new(left)),
                    self.span(start),
                ))
            }

            _ => Some(left),
        }
    }

    fn parse_binary_op(&mut self) -> Option<BinaryOp> {
        let op = match &self.current.kind {
            TokenKind::Plus => BinaryOp::Add,
            TokenKind::Minus => BinaryOp::Sub,
            TokenKind::Star => BinaryOp::Mul,
            TokenKind::Slash => BinaryOp::Div,
            TokenKind::Percent => BinaryOp::Mod,
            TokenKind::EqEq => BinaryOp::Eq,
            TokenKind::Ne => BinaryOp::Ne,
            TokenKind::Lt => BinaryOp::Lt,
            TokenKind::Gt => BinaryOp::Gt,
            TokenKind::Le => BinaryOp::Le,
            TokenKind::Ge => BinaryOp::Ge,
            TokenKind::And => BinaryOp::And,
            TokenKind::Or => BinaryOp::Or,
            _ => return None,
        };
        self.advance();
        Some(op)
    }

    fn parse_call_args(&mut self) -> Option<Vec<Argument>> {
        self.consume(TokenKind::LParen, "(");

        let mut args = Vec::new();

        while !self.check(&TokenKind::RParen) && !self.at_end() {
            let start = self.current.span.start;

            // Check for named argument
            let name = if matches!(self.current.kind, TokenKind::Ident(_)) {
                let ident = self.parse_identifier()?;

                if self.check(&TokenKind::Eq) {
                    self.advance();
                    Some(ident)
                } else {
                    // Not a named argument, put it back
                    // We need to re-parse this as an expression
                    let value = Spanned::new(ExprKind::Identifier(ident.node), ident.span);
                    let value = self.parse_infix(value, Precedence::None)?;
                    args.push(Argument {
                        name: None,
                        value,
                        span: self.span(start),
                    });

                    if !self.check(&TokenKind::RParen) {
                        self.consume(TokenKind::Comma, ",");
                    }
                    continue;
                }
            } else {
                None
            };

            let value = self.parse_expr()?;
            args.push(Argument {
                name,
                value,
                span: self.span(start),
            });

            if !self.check(&TokenKind::RParen) {
                self.consume(TokenKind::Comma, ",");
            }
        }

        self.consume(TokenKind::RParen, ")");
        Some(args)
    }

    fn parse_paren_or_lambda(&mut self, start: usize) -> Option<Expr> {
        self.advance(); // consume (

        // Empty parens followed by block or arrow is lambda with no params
        if self.check(&TokenKind::RParen) {
            self.advance();

            if self.check(&TokenKind::FatArrow) {
                self.advance();
                let body = self.parse_expr()?;
                return Some(Spanned::new(
                    ExprKind::Lambda(LambdaExpr {
                        params: Vec::new(),
                        body: LambdaBody::Expr(Box::new(body)),
                    }),
                    self.span(start),
                ));
            }

            if self.check(&TokenKind::LBrace) {
                let body = self.parse_block()?;
                return Some(Spanned::new(
                    ExprKind::Lambda(LambdaExpr {
                        params: Vec::new(),
                        body: LambdaBody::Block(body),
                    }),
                    self.span(start),
                ));
            }

            // Empty tuple/unit - treat as empty list for now
            return Some(Spanned::new(ExprKind::List(Vec::new()), self.span(start)));
        }

        // Parse first expression
        let first = self.parse_expr()?;

        // Check if this looks like a parameter list (has comma or type annotation)
        if self.check(&TokenKind::Comma) || self.check(&TokenKind::Colon) {
            // This is a lambda parameter list
            let mut params = vec![self.expr_to_param(first)?];

            while self.check(&TokenKind::Comma) {
                self.advance();
                let expr = self.parse_expr()?;
                params.push(self.expr_to_param(expr)?);
            }

            self.consume(TokenKind::RParen, ")");

            // Must be followed by => or {
            if self.check(&TokenKind::FatArrow) {
                self.advance();
                let body = self.parse_expr()?;
                return Some(Spanned::new(
                    ExprKind::Lambda(LambdaExpr {
                        params,
                        body: LambdaBody::Expr(Box::new(body)),
                    }),
                    self.span(start),
                ));
            }

            let body = self.parse_block()?;
            return Some(Spanned::new(
                ExprKind::Lambda(LambdaExpr {
                    params,
                    body: LambdaBody::Block(body),
                }),
                self.span(start),
            ));
        }

        self.consume(TokenKind::RParen, ")");

        // Check if followed by => or { (single param lambda)
        if self.check(&TokenKind::FatArrow) {
            self.advance();
            let params = vec![self.expr_to_param(first)?];
            let body = self.parse_expr()?;
            return Some(Spanned::new(
                ExprKind::Lambda(LambdaExpr {
                    params,
                    body: LambdaBody::Expr(Box::new(body)),
                }),
                self.span(start),
            ));
        }

        if self.check(&TokenKind::LBrace) {
            let params = vec![self.expr_to_param(first)?];
            let body = self.parse_block()?;
            return Some(Spanned::new(
                ExprKind::Lambda(LambdaExpr {
                    params,
                    body: LambdaBody::Block(body),
                }),
                self.span(start),
            ));
        }

        // Just a parenthesized expression
        Some(Spanned::new(
            ExprKind::Paren(Box::new(first)),
            self.span(start),
        ))
    }

    fn expr_to_param(&mut self, expr: Expr) -> Option<Param> {
        match expr.node {
            ExprKind::Identifier(name) => Some(Param {
                name: Spanned::new(name, expr.span),
                ty: None,
                default: None,
                is_rest: false,
                span: expr.span,
            }),
            _ => {
                self.error(ParseError::ExpectedIdent {
                    span: expr.span.start as usize..expr.span.end as usize,
                });
                None
            }
        }
    }

    fn parse_list(&mut self, start: usize) -> Option<Expr> {
        self.advance(); // consume [

        let mut elements = Vec::new();

        while !self.check(&TokenKind::RBracket) && !self.at_end() {
            elements.push(self.parse_expr()?);

            if !self.check(&TokenKind::RBracket) {
                self.consume(TokenKind::Comma, ",");
            }
        }

        self.consume(TokenKind::RBracket, "]");

        Some(Spanned::new(ExprKind::List(elements), self.span(start)))
    }

    fn parse_map_or_block(&mut self, start: usize) -> Option<Expr> {
        self.advance(); // consume {
        self.skip_newlines();

        // Empty braces
        if self.check(&TokenKind::RBrace) {
            self.advance();
            return Some(Spanned::new(ExprKind::Map(Vec::new()), self.span(start)));
        }

        // Check if first item looks like a map entry (has colon)
        let first_expr = self.parse_expr()?;

        if self.check(&TokenKind::Colon) {
            // It's a map
            self.advance();
            let first_value = self.parse_expr()?;
            let mut entries = vec![(first_expr, first_value)];

            while self.check(&TokenKind::Comma) {
                self.advance();
                self.skip_newlines();
                if self.check(&TokenKind::RBrace) {
                    break;
                }
                let key = self.parse_expr()?;
                self.consume(TokenKind::Colon, ":");
                let value = self.parse_expr()?;
                entries.push((key, value));
            }

            self.skip_newlines();
            self.consume(TokenKind::RBrace, "}");

            return Some(Spanned::new(ExprKind::Map(entries), self.span(start)));
        }

        // It's a block - convert first_expr to a statement
        let first_stmt = Spanned::new(StatementKind::Expr(first_expr), self.span(start));
        let mut statements = vec![first_stmt];

        self.skip_newlines();

        while !self.check(&TokenKind::RBrace) && !self.at_end() {
            if let Some(stmt) = self.parse_statement() {
                statements.push(stmt);
            }
            self.skip_newlines();
        }

        self.consume(TokenKind::RBrace, "}");

        Some(Spanned::new(
            ExprKind::Block(Block {
                statements,
                span: self.span(start),
            }),
            self.span(start),
        ))
    }

    fn parse_instance(&mut self, type_name: SmolStr, start: usize) -> Option<Expr> {
        self.advance(); // consume {
        self.skip_newlines();

        let mut fields = Vec::new();

        while !self.check(&TokenKind::RBrace) && !self.at_end() {
            let field_start = self.current.span.start;

            // Check for named field: `name = value`
            let name = if matches!(self.current.kind, TokenKind::Ident(_)) {
                let ident = self.parse_identifier()?;

                if self.check(&TokenKind::Eq) {
                    self.advance();
                    Some(ident)
                } else {
                    // Positional field - the identifier itself is the value
                    let value = Spanned::new(ExprKind::Identifier(ident.node), ident.span);
                    fields.push(InstanceField {
                        name: None,
                        value,
                        span: self.span(field_start),
                    });

                    if !self.check(&TokenKind::RBrace) {
                        if self.check(&TokenKind::Comma) {
                            self.advance();
                        }
                        self.skip_newlines();
                    }
                    continue;
                }
            } else {
                None
            };

            let value = self.parse_expr()?;
            fields.push(InstanceField {
                name,
                value,
                span: self.span(field_start),
            });

            if !self.check(&TokenKind::RBrace) {
                if self.check(&TokenKind::Comma) {
                    self.advance();
                }
                self.skip_newlines();
            }
        }

        self.consume(TokenKind::RBrace, "}");

        Some(Spanned::new(
            ExprKind::Instance(InstanceExpr {
                type_name: Spanned::new(type_name, self.span(start)),
                fields,
            }),
            self.span(start),
        ))
    }

    fn parse_match_expr(&mut self) -> Option<MatchExpr> {
        let subject = self.parse_expr()?;
        self.consume(TokenKind::LBrace, "{");
        self.skip_newlines();

        let mut arms = Vec::new();

        while !self.check(&TokenKind::RBrace) && !self.at_end() {
            if let Some(arm) = self.parse_match_arm() {
                arms.push(arm);
            }
            self.skip_newlines();
        }

        self.consume(TokenKind::RBrace, "}");

        Some(MatchExpr {
            subject: Box::new(subject),
            arms,
        })
    }

    fn parse_match_arm(&mut self) -> Option<MatchArm> {
        let start = self.current.span.start;
        let pattern = self.parse_pattern()?;

        // Optional guard
        let guard = if self.check(&TokenKind::If) {
            self.advance();
            Some(self.parse_expr()?)
        } else {
            None
        };

        self.consume(TokenKind::FatArrow, "=>");

        let body = if self.check(&TokenKind::LBrace) {
            MatchArmBody::Block(self.parse_block()?)
        } else {
            MatchArmBody::Expr(self.parse_expr()?)
        };

        Some(MatchArm {
            pattern,
            guard,
            body,
            span: self.span(start),
        })
    }

    fn parse_pattern(&mut self) -> Option<Spanned<Pattern>> {
        let start = self.current.span.start;

        let pattern = match &self.current.kind {
            // Wildcard
            TokenKind::Ident(s) if s == "_" => {
                self.advance();
                Pattern::Wildcard
            }
            // Identifier or constructor
            TokenKind::Ident(name) => {
                let name = name.clone();
                self.advance();

                // Check for constructor pattern: `Some { value }`
                if self.check(&TokenKind::LBrace) {
                    self.advance();
                    let mut fields = Vec::new();

                    while !self.check(&TokenKind::RBrace) && !self.at_end() {
                        fields.push(self.parse_identifier()?);
                        if !self.check(&TokenKind::RBrace) {
                            self.consume(TokenKind::Comma, ",");
                        }
                    }

                    self.consume(TokenKind::RBrace, "}");
                    Pattern::Constructor { name, fields }
                } else {
                    Pattern::Identifier(name)
                }
            }
            // Literal patterns
            TokenKind::Int(n) => {
                let n = *n;
                self.advance();
                Pattern::Literal(Literal::Int(n))
            }
            TokenKind::String(s) => {
                let s = s.clone();
                self.advance();
                Pattern::Literal(Literal::String(s))
            }
            TokenKind::True => {
                self.advance();
                Pattern::Literal(Literal::Bool(true))
            }
            TokenKind::False => {
                self.advance();
                Pattern::Literal(Literal::Bool(false))
            }
            _ => {
                self.error(ParseError::ExpectedExpr {
                    span: self.current.span.clone(),
                });
                return None;
            }
        };

        Some(Spanned::new(pattern, self.span(start)))
    }

    fn parse_select_expr(&mut self) -> Option<SelectExpr> {
        self.consume(TokenKind::LBrace, "{");
        self.skip_newlines();

        let mut arms = Vec::new();
        let mut default = None;

        while !self.check(&TokenKind::RBrace) && !self.at_end() {
            if self.check(&TokenKind::Default) {
                self.advance();
                self.consume(TokenKind::FatArrow, "=>");
                default = Some(self.parse_block()?);
            } else {
                if let Some(arm) = self.parse_select_arm() {
                    arms.push(arm);
                }
            }
            self.skip_newlines();
        }

        self.consume(TokenKind::RBrace, "}");

        Some(SelectExpr { arms, default })
    }

    fn parse_select_arm(&mut self) -> Option<SelectArm> {
        let start = self.current.span.start;
        let binding = self.parse_identifier()?;
        self.consume(TokenKind::From, "from");
        let channel = self.parse_expr()?;
        self.consume(TokenKind::FatArrow, "=>");

        let body = if self.check(&TokenKind::LBrace) {
            MatchArmBody::Block(self.parse_block()?)
        } else {
            MatchArmBody::Expr(self.parse_expr()?)
        };

        Some(SelectArm {
            binding,
            channel,
            body,
            span: self.span(start),
        })
    }

    // ========================================================================
    // AI Block
    // ========================================================================

    /// Parse an AI intent block.
    ///
    /// Syntax variants:
    /// - Named: `ai func_name(params) -> ReturnType { intent text }`
    /// - Anonymous: `ai(params) -> ReturnType { intent text }`
    ///
    /// The block body contains natural language intent description.
    fn parse_ai_block(&mut self) -> Option<AiBlock> {
        let start = self.previous.span.start; // 'ai' was already consumed

        // Check if this is named or anonymous
        // Named: `ai func_name(...)` - identifier followed by `(`
        // Anonymous: `ai(...)` - directly `(`
        let name = if matches!(self.current.kind, TokenKind::Ident(_)) {
            Some(self.parse_identifier()?)
        } else {
            None
        };

        // Parse parameters
        let params = self.parse_params()?;

        // Optional return type
        let return_ty = if self.check(&TokenKind::Arrow) {
            self.advance();
            Some(self.parse_type()?)
        } else {
            None
        };

        // Parse the intent body - everything inside { } is natural language
        let intent = self.parse_intent_body()?;

        Some(AiBlock {
            name,
            params,
            return_ty,
            intent,
            span: self.span(start),
        })
    }

    /// Parse the intent body - collects raw text until closing brace.
    /// The content is natural language, not code.
    fn parse_intent_body(&mut self) -> Option<SmolStr> {
        self.consume(TokenKind::LBrace, "{");

        // We need to collect all text until the matching closing brace
        // Since the lexer tokenizes everything, we'll collect token text
        let mut intent_parts = Vec::new();
        let mut brace_depth = 1;

        while !self.at_end() && brace_depth > 0 {
            match &self.current.kind {
                TokenKind::LBrace => {
                    brace_depth += 1;
                    intent_parts.push("{".to_string());
                    self.advance();
                }
                TokenKind::RBrace => {
                    brace_depth -= 1;
                    if brace_depth > 0 {
                        intent_parts.push("}".to_string());
                        self.advance();
                    }
                    // Don't advance on final } - let consume handle it
                }
                TokenKind::Newline => {
                    intent_parts.push("\n".to_string());
                    // Manually advance past newlines
                    self.previous = std::mem::replace(
                        &mut self.current,
                        self.lexer
                            .next()
                            .and_then(|r| r.ok())
                            .unwrap_or(Token::new(TokenKind::Eof, 0..0)),
                    );
                }
                TokenKind::Ident(s) => {
                    intent_parts.push(s.to_string());
                    intent_parts.push(" ".to_string());
                    self.advance();
                }
                TokenKind::String(s) => {
                    intent_parts.push(format!("\"{}\"", s));
                    intent_parts.push(" ".to_string());
                    self.advance();
                }
                TokenKind::Int(n) => {
                    intent_parts.push(n.to_string());
                    intent_parts.push(" ".to_string());
                    self.advance();
                }
                TokenKind::Float(n) => {
                    intent_parts.push(n.to_string());
                    intent_parts.push(" ".to_string());
                    self.advance();
                }
                // Convert tokens back to text for the intent
                TokenKind::Plus => {
                    intent_parts.push("+".to_string());
                    self.advance();
                }
                TokenKind::Minus => {
                    intent_parts.push("-".to_string());
                    self.advance();
                }
                TokenKind::Star => {
                    intent_parts.push("*".to_string());
                    self.advance();
                }
                TokenKind::Slash => {
                    intent_parts.push("/".to_string());
                    self.advance();
                }
                TokenKind::Dot => {
                    intent_parts.push(".".to_string());
                    self.advance();
                }
                TokenKind::Comma => {
                    intent_parts.push(",".to_string());
                    self.advance();
                }
                TokenKind::Colon => {
                    intent_parts.push(":".to_string());
                    self.advance();
                }
                TokenKind::LParen => {
                    intent_parts.push("(".to_string());
                    self.advance();
                }
                TokenKind::RParen => {
                    intent_parts.push(")".to_string());
                    self.advance();
                }
                TokenKind::LBracket => {
                    intent_parts.push("[".to_string());
                    self.advance();
                }
                TokenKind::RBracket => {
                    intent_parts.push("]".to_string());
                    self.advance();
                }
                TokenKind::Arrow => {
                    intent_parts.push("->".to_string());
                    self.advance();
                }
                TokenKind::FatArrow => {
                    intent_parts.push("=>".to_string());
                    self.advance();
                }
                TokenKind::And => {
                    intent_parts.push("and".to_string());
                    intent_parts.push(" ".to_string());
                    self.advance();
                }
                TokenKind::Or => {
                    intent_parts.push("or".to_string());
                    intent_parts.push(" ".to_string());
                    self.advance();
                }
                TokenKind::Not => {
                    intent_parts.push("not".to_string());
                    intent_parts.push(" ".to_string());
                    self.advance();
                }
                TokenKind::In => {
                    intent_parts.push("in".to_string());
                    intent_parts.push(" ".to_string());
                    self.advance();
                }
                TokenKind::For => {
                    intent_parts.push("for".to_string());
                    intent_parts.push(" ".to_string());
                    self.advance();
                }
                TokenKind::If => {
                    intent_parts.push("if".to_string());
                    intent_parts.push(" ".to_string());
                    self.advance();
                }
                TokenKind::Else => {
                    intent_parts.push("else".to_string());
                    intent_parts.push(" ".to_string());
                    self.advance();
                }
                TokenKind::While => {
                    intent_parts.push("while".to_string());
                    intent_parts.push(" ".to_string());
                    self.advance();
                }
                TokenKind::Return => {
                    intent_parts.push("return".to_string());
                    intent_parts.push(" ".to_string());
                    self.advance();
                }
                TokenKind::True => {
                    intent_parts.push("true".to_string());
                    intent_parts.push(" ".to_string());
                    self.advance();
                }
                TokenKind::False => {
                    intent_parts.push("false".to_string());
                    intent_parts.push(" ".to_string());
                    self.advance();
                }
                TokenKind::None => {
                    intent_parts.push("none".to_string());
                    intent_parts.push(" ".to_string());
                    self.advance();
                }
                TokenKind::Some => {
                    intent_parts.push("some".to_string());
                    intent_parts.push(" ".to_string());
                    self.advance();
                }

                TokenKind::Match => {
                    intent_parts.push("match".to_string());
                    intent_parts.push(" ".to_string());
                    self.advance();
                }
                TokenKind::Try => {
                    intent_parts.push("try".to_string());
                    intent_parts.push(" ".to_string());
                    self.advance();
                }
                TokenKind::Catch => {
                    intent_parts.push("catch".to_string());
                    intent_parts.push(" ".to_string());
                    self.advance();
                }
                TokenKind::Break => {
                    intent_parts.push("break".to_string());
                    intent_parts.push(" ".to_string());
                    self.advance();
                }
                TokenKind::Continue => {
                    intent_parts.push("continue".to_string());
                    intent_parts.push(" ".to_string());
                    self.advance();
                }
                TokenKind::Spawn => {
                    intent_parts.push("spawn".to_string());
                    intent_parts.push(" ".to_string());
                    self.advance();
                }
                TokenKind::Async => {
                    intent_parts.push("async".to_string());
                    intent_parts.push(" ".to_string());
                    self.advance();
                }
                TokenKind::Select => {
                    intent_parts.push("select".to_string());
                    intent_parts.push(" ".to_string());
                    self.advance();
                }
                TokenKind::Ai => {
                    intent_parts.push("ai".to_string());
                    intent_parts.push(" ".to_string());
                    self.advance();
                }
                TokenKind::Eq => {
                    intent_parts.push("=".to_string());
                    self.advance();
                }
                TokenKind::EqEq => {
                    intent_parts.push("==".to_string());
                    self.advance();
                }
                TokenKind::Ne => {
                    intent_parts.push("!=".to_string());
                    self.advance();
                }
                TokenKind::Lt => {
                    intent_parts.push("<".to_string());
                    self.advance();
                }
                TokenKind::Le => {
                    intent_parts.push("<=".to_string());
                    self.advance();
                }
                TokenKind::Gt => {
                    intent_parts.push(">".to_string());
                    self.advance();
                }
                TokenKind::Ge => {
                    intent_parts.push(">=".to_string());
                    self.advance();
                }
                TokenKind::Pipe => {
                    intent_parts.push("|".to_string());
                    self.advance();
                }
                TokenKind::Question => {
                    intent_parts.push("?".to_string());
                    self.advance();
                }
                TokenKind::Percent => {
                    intent_parts.push("%".to_string());
                    self.advance();
                }
                TokenKind::DotDot => {
                    intent_parts.push("..".to_string());
                    self.advance();
                }
                TokenKind::DotDotEq => {
                    intent_parts.push("..=".to_string());
                    self.advance();
                }
                _ => {
                    // For any unhandled token, skip with a space
                    // This shouldn't happen often if we've covered all tokens
                    self.advance();
                }
            }
        }

        self.consume(TokenKind::RBrace, "}");

        // Clean up the intent text
        let intent = intent_parts
            .join("")
            .lines()
            .map(|line| line.trim())
            .filter(|line| !line.is_empty())
            .collect::<Vec<_>>()
            .join(" ");

        Some(SmolStr::from(intent.trim()))
    }

    // ========================================================================
    // Helpers
    // ========================================================================

    fn parse_identifier(&mut self) -> Option<Spanned<SmolStr>> {
        match &self.current.kind {
            TokenKind::Ident(name) => {
                let name = name.clone();
                let span = self.current_span();
                self.advance();
                Some(Spanned::new(name, span))
            }
            _ => {
                self.error(ParseError::ExpectedIdent {
                    span: self.current.span.clone(),
                });
                None
            }
        }
    }

    /// Parse an interpolated string into parts.
    /// Input is the raw string content (without quotes) containing `{expr}` sequences.
    fn parse_interpolated_string_parts(&mut self, raw: &str) -> Option<Vec<StringPart>> {
        let mut parts = Vec::new();
        let mut current_literal = String::new();
        let mut chars = raw.chars().peekable();

        while let Some(c) = chars.next() {
            if c == '\\' {
                // Handle escape sequences
                match chars.next() {
                    Some('n') => current_literal.push('\n'),
                    Some('t') => current_literal.push('\t'),
                    Some('r') => current_literal.push('\r'),
                    Some('\\') => current_literal.push('\\'),
                    Some('"') => current_literal.push('"'),
                    Some('{') => current_literal.push('{'),
                    Some('}') => current_literal.push('}'),
                    Some(other) => {
                        current_literal.push('\\');
                        current_literal.push(other);
                    }
                    None => current_literal.push('\\'),
                }
            } else if c == '{' {
                // Start of interpolation - save current literal if not empty
                if !current_literal.is_empty() {
                    parts.push(StringPart::Literal(SmolStr::from(&current_literal)));
                    current_literal.clear();
                }

                // Extract the expression inside braces
                let mut expr_str = String::new();
                let mut brace_depth = 1;

                while let Some(ec) = chars.next() {
                    if ec == '{' {
                        brace_depth += 1;
                        expr_str.push(ec);
                    } else if ec == '}' {
                        brace_depth -= 1;
                        if brace_depth == 0 {
                            break;
                        }
                        expr_str.push(ec);
                    } else {
                        expr_str.push(ec);
                    }
                }

                // Parse the expression
                if !expr_str.is_empty() {
                    let mut expr_parser = Parser::new(&expr_str);
                    if let Some(expr) = expr_parser.parse_expr() {
                        parts.push(StringPart::Expr(expr));
                    } else {
                        // If parsing fails, treat it as literal
                        self.error(ParseError::ExpectedExpr {
                            span: self.current.span.clone(),
                        });
                        return None;
                    }
                }
            } else {
                current_literal.push(c);
            }
        }

        // Add any remaining literal
        if !current_literal.is_empty() {
            parts.push(StringPart::Literal(SmolStr::from(&current_literal)));
        }

        Some(parts)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn parse(source: &str) -> SourceFile {
        let mut parser = Parser::new(source);
        parser.parse_source_file()
    }

    #[test]
    fn test_type_definition() {
        let ast = parse("User { name, age, email }");
        assert_eq!(ast.items.len(), 1);
        match &ast.items[0].node {
            ItemKind::TypeDef(def) => {
                assert_eq!(def.name.node.as_str(), "User");
                assert_eq!(def.fields.len(), 3);
            }
            _ => panic!("expected type def"),
        }
    }

    #[test]
    fn test_function_definition() {
        let ast = parse("add(a, b) { a + b }");
        assert_eq!(ast.items.len(), 1);
        match &ast.items[0].node {
            ItemKind::FunctionDef(def) => {
                assert_eq!(def.name.node.as_str(), "add");
                assert_eq!(def.params.len(), 2);
            }
            _ => panic!("expected function def"),
        }
    }

    #[test]
    fn test_assignment() {
        let ast = parse("x = 42");
        assert_eq!(ast.items.len(), 1);
        match &ast.items[0].node {
            ItemKind::Statement(stmt) => match &stmt.node {
                StatementKind::Assignment(assign) => {
                    assert_eq!(assign.targets.len(), 1);
                    assert_eq!(assign.targets[0].name.node.as_str(), "x");
                }
                _ => panic!("expected assignment"),
            },
            _ => panic!("expected statement"),
        }
    }

    #[test]
    fn test_pipe_expression() {
        let ast = parse("x = users | filter_active | sort_by_name");
        assert_eq!(ast.items.len(), 1);
        match &ast.items[0].node {
            ItemKind::Statement(stmt) => match &stmt.node {
                StatementKind::Assignment(assign) => match &assign.value.node {
                    ExprKind::Pipe(_) => {}
                    _ => panic!("expected pipe"),
                },
                _ => panic!("expected assignment"),
            },
            _ => panic!("expected statement"),
        }
    }

    #[test]
    fn test_lambda_arrow() {
        let ast = parse("f = x => x * 2");
        assert_eq!(ast.items.len(), 1);
        match &ast.items[0].node {
            ItemKind::Statement(stmt) => match &stmt.node {
                StatementKind::Assignment(assign) => match &assign.value.node {
                    ExprKind::Lambda(lambda) => {
                        assert_eq!(lambda.params.len(), 1);
                        match &lambda.body {
                            LambdaBody::Expr(_) => {}
                            _ => panic!("expected expr body"),
                        }
                    }
                    _ => panic!("expected lambda"),
                },
                _ => panic!("expected assignment"),
            },
            _ => panic!("expected statement"),
        }
    }

    #[test]
    fn test_lambda_block() {
        let ast = parse("f = (a, b) { a + b }");
        assert_eq!(ast.items.len(), 1);
        match &ast.items[0].node {
            ItemKind::Statement(stmt) => match &stmt.node {
                StatementKind::Assignment(assign) => match &assign.value.node {
                    ExprKind::Lambda(lambda) => {
                        assert_eq!(lambda.params.len(), 2);
                        match &lambda.body {
                            LambdaBody::Block(_) => {}
                            _ => panic!("expected block body"),
                        }
                    }
                    _ => panic!("expected lambda"),
                },
                _ => panic!("expected assignment"),
            },
            _ => panic!("expected statement"),
        }
    }

    #[test]
    fn test_instance_creation() {
        let ast = parse(r#"user = User { name = "Alice", age = 30 }"#);
        assert_eq!(ast.items.len(), 1);
        match &ast.items[0].node {
            ItemKind::Statement(stmt) => match &stmt.node {
                StatementKind::Assignment(assign) => match &assign.value.node {
                    ExprKind::Instance(inst) => {
                        assert_eq!(inst.type_name.node.as_str(), "User");
                        assert_eq!(inst.fields.len(), 2);
                    }
                    _ => panic!("expected instance"),
                },
                _ => panic!("expected assignment"),
            },
            _ => panic!("expected statement"),
        }
    }

    #[test]
    fn test_for_loop() {
        let ast = parse("for item in items { print(item) }");
        assert_eq!(ast.items.len(), 1);
        match &ast.items[0].node {
            ItemKind::Statement(stmt) => match &stmt.node {
                StatementKind::For(for_stmt) => match &for_stmt.pattern {
                    ForPattern::Single(name) => {
                        assert_eq!(name.node.as_str(), "item");
                    }
                    _ => panic!("expected single pattern"),
                },
                _ => panic!("expected for"),
            },
            _ => panic!("expected statement"),
        }
    }

    #[test]
    fn test_error_propagation() {
        let ast = parse("result = get_user(id)?");
        assert_eq!(ast.items.len(), 1);
        match &ast.items[0].node {
            ItemKind::Statement(stmt) => match &stmt.node {
                StatementKind::Assignment(assign) => match &assign.value.node {
                    ExprKind::Propagate(_) => {}
                    _ => panic!("expected propagate"),
                },
                _ => panic!("expected assignment"),
            },
            _ => panic!("expected statement"),
        }
    }

    #[test]
    fn test_match_expression() {
        let ast = parse(
            r#"
            match x {
                0 => "zero"
                n => "other"
            }
        "#,
        );
        assert_eq!(ast.items.len(), 1);
    }

    #[test]
    fn test_ai_block_named() {
        let ast = parse(
            r#"
            ai summarize_activity(user: User) -> ActivitySummary {
                Summarize the user activity over the last 30 days.
                Group by activity type and find most common.
            }
        "#,
        );
        assert_eq!(ast.items.len(), 1);
        match &ast.items[0].node {
            ItemKind::AiFunctionDef(ai_block) => {
                assert!(ai_block.name.is_some());
                assert_eq!(
                    ai_block.name.as_ref().unwrap().node.as_str(),
                    "summarize_activity"
                );
                assert_eq!(ai_block.params.len(), 1);
                assert!(ai_block.return_ty.is_some());
                assert!(!ai_block.intent.is_empty());
            }
            _ => panic!("expected ai function def"),
        }
    }

    #[test]
    fn test_ai_block_anonymous() {
        let ast = parse(
            r#"
            result = ai(data: Data) -> Summary {
                Analyze and summarize this data
            }
        "#,
        );
        assert_eq!(ast.items.len(), 1);
        match &ast.items[0].node {
            ItemKind::Statement(stmt) => match &stmt.node {
                StatementKind::Assignment(assign) => match &assign.value.node {
                    ExprKind::Ai(ai_block) => {
                        assert!(ai_block.name.is_none());
                        assert_eq!(ai_block.params.len(), 1);
                    }
                    _ => panic!("expected ai block"),
                },
                _ => panic!("expected assignment"),
            },
            _ => panic!("expected statement"),
        }
    }
}
