//! Recursive descent parser with Pratt expression parsing for Haira source code.
//!
//! This is a 1:1 port of `compiler/internal/parser/parser.go`.

use haira_ast::*;
use haira_errors::{Diagnostic, Span};
use haira_lexer::Lexer;
use haira_token::{Token, TokenKind};

// ---------------------------------------------------------------------------
// Precedence
// ---------------------------------------------------------------------------

/// Precedence levels for Pratt expression parsing.
#[derive(Debug, Clone, Copy, PartialEq, Eq, PartialOrd, Ord)]
#[repr(u8)]
pub enum Precedence {
    None = 0,
    Orelse,     // orelse
    Pipe,       // |>
    Or,         // or
    And,        // and
    BitOr,      // |
    BitXor,     // ^
    BitAnd,     // &
    Equality,   // == !=
    Comparison, // < > <= >= .. ..=
    Shift,      // << >>
    Term,       // + -
    Factor,     // * / %
    Unary,      // - not ~
    Call,       // () [] . ?
}

/// Returns the precedence of a token kind for infix parsing.
fn precedence_of(kind: TokenKind) -> Precedence {
    match kind {
        TokenKind::Orelse => Precedence::Orelse,
        TokenKind::PipeArrow => Precedence::Pipe,
        TokenKind::Or => Precedence::Or,
        TokenKind::And => Precedence::And,
        TokenKind::Pipe => Precedence::BitOr,
        TokenKind::Caret => Precedence::BitXor,
        TokenKind::Amp => Precedence::BitAnd,
        TokenKind::EqEq | TokenKind::Ne => Precedence::Equality,
        TokenKind::Lt | TokenKind::Gt | TokenKind::Le | TokenKind::Ge => Precedence::Comparison,
        TokenKind::Shl | TokenKind::Shr => Precedence::Shift,
        TokenKind::Plus | TokenKind::Minus => Precedence::Term,
        TokenKind::Star | TokenKind::Slash | TokenKind::Percent => Precedence::Factor,
        TokenKind::LParen | TokenKind::LBracket | TokenKind::Dot | TokenKind::Question => {
            Precedence::Call
        }
        TokenKind::DotDot | TokenKind::DotDotEq => Precedence::Comparison,
        _ => Precedence::None,
    }
}

// ---------------------------------------------------------------------------
// ParseError
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Parser
// ---------------------------------------------------------------------------

/// Holds the state for parsing a stream of tokens.
pub struct Parser {
    tokens: Vec<Token>,
    current: usize,
    previous: Token,
    errors: Vec<Diagnostic>,
}

/// Tokenizes `source` and parses it into a [`SourceFile`] and a list of
/// [`Diagnostic`] errors.
pub fn parse(source: &str) -> (SourceFile, Vec<Diagnostic>) {
    let lexer = Lexer::new(source);
    let tokens = lexer.all_tokens().to_vec();

    let mut p = Parser {
        tokens,
        current: 0,
        previous: Token::new(TokenKind::EOF, "", 0, 0),
        errors: Vec::new(),
    };
    // Advance to the first significant token.
    p.skip_trivia();

    let sf = p.parse_source_file();
    let errors = p.errors;
    (sf, errors)
}

// ---------------------------------------------------------------------------
// Token helpers
// ---------------------------------------------------------------------------

impl Parser {
    /// Returns the current token without advancing.
    fn peek(&self) -> &Token {
        if self.current < self.tokens.len() {
            &self.tokens[self.current]
        } else {
            // Return a synthetic EOF. We keep a static-like approach by
            // returning a reference to the last token or constructing on the
            // fly — but since we need a reference, we'll just return the last
            // token (which should be EOF after lexing). The Go code constructs
            // a value; here we use the last token.
            self.tokens.last().unwrap_or(&self.previous)
        }
    }

    /// Clones the current token (used when we need an owned copy).
    fn peek_cloned(&self) -> Token {
        self.peek().clone()
    }

    /// Moves past the current token (skipping newlines and comments) and
    /// stores the consumed token in `self.previous`.
    fn advance(&mut self) {
        if self.current < self.tokens.len() {
            self.previous = self.tokens[self.current].clone();
            self.current += 1;
        }
        self.skip_trivia();
    }

    /// Skips Newline, LineComment and BlockComment tokens.
    fn skip_trivia(&mut self) {
        while self.current < self.tokens.len() {
            let k = self.tokens[self.current].kind;
            if k == TokenKind::Newline || k == TokenKind::LineComment || k == TokenKind::BlockComment
            {
                self.current += 1;
            } else {
                break;
            }
        }
    }

    /// Advances past any Newline tokens only (comments are handled by
    /// advance/skip_trivia).
    fn skip_newlines(&mut self) {
        while self.current < self.tokens.len()
            && self.tokens[self.current].kind == TokenKind::Newline
        {
            self.current += 1;
        }
    }

    /// Returns `true` if the current token is of the given kind.
    fn check(&self, kind: TokenKind) -> bool {
        self.peek().kind == kind
    }

    /// Returns `true` if the current token is EOF.
    fn at_end(&self) -> bool {
        self.peek().kind == TokenKind::EOF
    }

    /// Checks that the current token matches `kind`, advances, and returns
    /// `true`. Otherwise records an error and returns `false`.
    fn consume(&mut self, kind: TokenKind, expected: &str) -> bool {
        if self.check(kind) {
            self.advance();
            return true;
        }
        let tok = self.peek_cloned();
        self.add_error(
            format!("expected {}, found {}", expected, tok.kind),
            Span::new(tok.start, tok.end),
        );
        false
    }

    /// Appends a [`Diagnostic`] error.
    fn add_error(&mut self, msg: String, span: Span) {
        self.errors.push(Diagnostic::error(msg, span));
    }

    /// Returns a [`Span`] from `start` to the end of the previous token.
    fn span(&self, start: usize) -> Span {
        Span::new(start, self.previous.end)
    }

    /// Returns the span of the current token.
    fn current_span(&self) -> Span {
        let tok = self.peek();
        Span::new(tok.start, tok.end)
    }
}

// ---------------------------------------------------------------------------
// Top-level parsing
// ---------------------------------------------------------------------------

impl Parser {
    fn parse_source_file(&mut self) -> SourceFile {
        let start = self.peek().start;
        let mut items = Vec::new();

        while !self.at_end() {
            self.skip_newlines();
            if self.at_end() {
                break;
            }
            if let Some(item) = self.parse_item() {
                items.push(item);
            } else {
                // Error recovery: skip token.
                self.advance();
            }
        }

        SourceFile {
            items,
            span: self.span(start),
        }
    }

    fn parse_item(&mut self) -> Option<Item> {
        let start = self.peek().start;

        // Check for `pub` modifier.
        let is_public = if self.check(TokenKind::Pub) {
            self.advance();
            true
        } else {
            false
        };

        match self.peek().kind {
            TokenKind::Ident => self.parse_item_ident(start, is_public),

            TokenKind::Import => {
                self.advance();
                let decl = self.parse_import_decl()?;
                Some(Spanned::new(ItemKind::ImportDecl(decl), self.span(start)))
            }

            TokenKind::Export => {
                self.advance();
                let decl = self.parse_export_decl()?;
                Some(Spanned::new(ItemKind::ExportDecl(decl), self.span(start)))
            }

            TokenKind::Provider => {
                self.advance();
                let decl = self.parse_provider_decl()?;
                Some(Spanned::new(
                    ItemKind::ProviderDecl(decl),
                    self.span(start),
                ))
            }

            TokenKind::Tool => {
                self.advance();
                let decl = self.parse_tool_decl()?;
                Some(Spanned::new(ItemKind::ToolDecl(decl), self.span(start)))
            }

            TokenKind::Agent => {
                self.advance();
                let decl = self.parse_agent_decl()?;
                Some(Spanned::new(ItemKind::AgentDecl(decl), self.span(start)))
            }

            TokenKind::At => {
                // Collect all decorators before a workflow or tool declaration.
                let mut decorators = Vec::new();
                while self.check(TokenKind::At) {
                    let dec = self.parse_decorator()?;
                    decorators.push(dec);
                }
                if self.check(TokenKind::Workflow) {
                    self.advance();
                    // Separate trigger from other decorators.
                    let mut trigger = None;
                    let mut extras = Vec::new();
                    for d in decorators {
                        match d.name.node.as_str() {
                            "webhook" | "get" | "post" | "put" | "delete" => {
                                trigger = Some(d);
                            }
                            _ => {
                                extras.push(d);
                            }
                        }
                    }
                    let mut wf = self.parse_workflow_decl(trigger)?;
                    wf.decorators = extras;
                    return Some(Spanned::new(
                        ItemKind::WorkflowDecl(wf),
                        self.span(start),
                    ));
                }
                if self.check(TokenKind::Tool) {
                    self.advance();
                    let mut tool = self.parse_tool_decl()?;
                    tool.decorators = decorators;
                    return Some(Spanned::new(ItemKind::ToolDecl(tool), self.span(start)));
                }
                let tok = self.peek_cloned();
                self.add_error(
                    "expected workflow or tool after decorator".into(),
                    Span::new(tok.start, tok.end),
                );
                None
            }

            TokenKind::Workflow => {
                self.advance();
                let wf = self.parse_workflow_decl(None)?;
                Some(Spanned::new(
                    ItemKind::WorkflowDecl(wf),
                    self.span(start),
                ))
            }

            TokenKind::Test => {
                self.advance();
                let td = self.parse_test_decl()?;
                Some(Spanned::new(ItemKind::TestDecl(td), self.span(start)))
            }

            TokenKind::Fn => {
                self.advance();
                let fd = self.parse_fn_decl(is_public)?;
                Some(Spanned::new(ItemKind::FunctionDef(fd), self.span(start)))
            }

            TokenKind::Enum => {
                self.advance();
                let ed = self.parse_enum_decl(is_public)?;
                Some(Spanned::new(ItemKind::EnumDef(ed), self.span(start)))
            }

            TokenKind::Struct => {
                self.advance();
                let name = self.parse_identifier()?;
                let td = self.parse_type_def_body(is_public, name)?;
                Some(Spanned::new(ItemKind::TypeDef(td), self.span(start)))
            }

            TokenKind::Type => {
                self.advance();
                let name = self.parse_identifier()?;
                self.consume(TokenKind::Eq, "=");
                let ty = self.parse_type()?;
                Some(Spanned::new(
                    ItemKind::TypeAlias(TypeAlias { name, ty }),
                    self.span(start),
                ))
            }

            TokenKind::If
            | TokenKind::For
            | TokenKind::While
            | TokenKind::Return
            | TokenKind::Match
            | TokenKind::Try
            | TokenKind::Defer
            | TokenKind::Break
            | TokenKind::Continue
            | TokenKind::Spawn
            | TokenKind::Async
            | TokenKind::Let
            | TokenKind::Const => {
                let stmt = self.parse_statement()?;
                let sp = stmt.span;
                Some(Spanned::new(ItemKind::ItemStatement(stmt), sp))
            }

            _ => {
                // Could be an expression statement starting with a literal,
                // unary, etc. Try to parse as a statement.
                let k = self.peek().kind;
                if matches!(
                    k,
                    TokenKind::Minus
                        | TokenKind::Not
                        | TokenKind::LParen
                        | TokenKind::LBracket
                        | TokenKind::LBrace
                        | TokenKind::Int
                        | TokenKind::Float
                        | TokenKind::String
                        | TokenKind::InterpolatedString
                        | TokenKind::TripleQuoteString
                        | TokenKind::True
                        | TokenKind::False
                        | TokenKind::None
                        | TokenKind::Some
                        | TokenKind::Err
                        | TokenKind::Ok
                        | TokenKind::Select
                ) {
                    let stmt = self.parse_statement()?;
                    let sp = stmt.span;
                    return Some(Spanned::new(ItemKind::ItemStatement(stmt), sp));
                }
                let tok = self.peek_cloned();
                self.add_error(
                    "expected item or statement".into(),
                    Span::new(tok.start, tok.end),
                );
                None
            }
        }
    }

    /// Handles items that start with an identifier: type defs, function defs,
    /// method defs, type aliases, assignments, expression stmts.
    fn parse_item_ident(&mut self, start: usize, is_public: bool) -> Option<Item> {
        let name = self.parse_identifier()?;

        match self.peek().kind {
            // Map/struct literal or block: name { ... }
            TokenKind::LBrace => {
                let expr = Spanned::new(ExprKind::Ident(name.node.clone()), name.span);
                let full_expr = self.parse_expr_rest(expr)?;
                let stmt = self.parse_statement_rest(full_expr)?;
                Some(Spanned::new(
                    ItemKind::ItemStatement(stmt),
                    self.span(start),
                ))
            }

            // Function def or call: name(...)
            TokenKind::LParen => {
                let expr = Spanned::new(ExprKind::Ident(name.node.clone()), name.span);
                let call_expr = self.parse_infix(expr, Precedence::None)?;
                // If followed by { or ->, it's a function definition.
                if self.check(TokenKind::LBrace) || self.check(TokenKind::Arrow) {
                    if let ExprKind::Call(ref call) = call_expr.node {
                        let params = self.args_to_params(&call.args)?;
                        let return_ty = if self.check(TokenKind::Arrow) {
                            self.advance();
                            Some(self.parse_type()?)
                        } else {
                            Option::None
                        };
                        let body = self.parse_block()?;
                        let fd = FunctionDef {
                            is_public,
                            name,
                            params,
                            return_ty,
                            body,
                        };
                        return Some(Spanned::new(
                            ItemKind::FunctionDef(fd),
                            self.span(start),
                        ));
                    }
                    let tok = self.peek_cloned();
                    self.add_error(
                        "expected function definition".into(),
                        Span::new(tok.start, tok.end),
                    );
                    return None;
                }
                // Otherwise expression statement (function call).
                let stmt = Spanned::new(
                    StmtKind::Expr(ExprStmt {
                        value: Box::new(call_expr),
                    }),
                    self.span(start),
                );
                Some(Spanned::new(
                    ItemKind::ItemStatement(stmt),
                    self.span(start),
                ))
            }

            // Method definition or field access: Name.method or obj.field
            TokenKind::Dot => {
                let first_char = first_rune(&name.node);
                if first_char.is_uppercase() {
                    // Method definition: Type.method(...)
                    self.advance(); // consume .
                    let method_name = self.parse_identifier()?;
                    let md = self.parse_method_def_body(name, method_name)?;
                    return Some(Spanned::new(ItemKind::MethodDef(md), self.span(start)));
                }
                // Field access expression -> statement
                let expr = Spanned::new(ExprKind::Ident(name.node.clone()), name.span);
                let full_expr = self.parse_expr_rest(expr)?;
                let stmt = self.parse_statement_rest(full_expr)?;
                Some(Spanned::new(
                    ItemKind::ItemStatement(stmt),
                    self.span(start),
                ))
            }

            // Variable assignment: name = ...
            TokenKind::Eq => {
                let expr = Spanned::new(ExprKind::Ident(name.node.clone()), name.span);
                let stmt = self.parse_statement_rest(expr)?;
                Some(Spanned::new(
                    ItemKind::ItemStatement(stmt),
                    self.span(start),
                ))
            }

            // Compound assignment: name += / -= / *= / /= / %=
            TokenKind::PlusEq
            | TokenKind::MinusEq
            | TokenKind::StarEq
            | TokenKind::SlashEq
            | TokenKind::PercentEq => {
                let expr = Spanned::new(ExprKind::Ident(name.node.clone()), name.span);
                let stmt = self.parse_statement_rest(expr)?;
                Some(Spanned::new(
                    ItemKind::ItemStatement(stmt),
                    self.span(start),
                ))
            }

            // Multi-assignment: a, b = ...
            TokenKind::Comma => {
                let expr = Spanned::new(ExprKind::Ident(name.node.clone()), name.span);
                let stmt = self.parse_statement_rest(expr)?;
                Some(Spanned::new(
                    ItemKind::ItemStatement(stmt),
                    self.span(start),
                ))
            }

            _ => {
                // Anything else: treat identifier as start of expression statement.
                let expr = Spanned::new(ExprKind::Ident(name.node.clone()), name.span);
                let full_expr = self.parse_expr_rest(expr)?;
                let stmt = self.parse_statement_rest(full_expr)?;
                Some(Spanned::new(
                    ItemKind::ItemStatement(stmt),
                    self.span(start),
                ))
            }
        }
    }
}

// ---------------------------------------------------------------------------
// Type definitions
// ---------------------------------------------------------------------------

impl Parser {
    fn parse_type_def_body(
        &mut self,
        is_public: bool,
        name: Spanned<String>,
    ) -> Option<TypeDef> {
        self.consume(TokenKind::LBrace, "{");
        self.skip_newlines();

        let mut fields = Vec::new();
        while !self.check(TokenKind::RBrace) && !self.at_end() {
            if let Some(f) = self.parse_field() {
                fields.push(f);
            }
            if self.check(TokenKind::Comma) {
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
        let start = self.peek().start;
        let name = self.parse_identifier()?;

        let ty = if self.check(TokenKind::Colon) {
            self.advance();
            Some(self.parse_type()?)
        } else {
            Option::None
        };

        let default = if self.check(TokenKind::Eq) {
            self.advance();
            let e = self.parse_expr()?;
            Some(Box::new(e))
        } else {
            Option::None
        };

        Some(Field {
            name,
            ty,
            default,
            span: self.span(start),
        })
    }
}

// ---------------------------------------------------------------------------
// Enum definitions
// ---------------------------------------------------------------------------

impl Parser {
    fn parse_enum_decl(&mut self, is_public: bool) -> Option<EnumDef> {
        let name = self.parse_identifier()?;
        self.consume(TokenKind::LBrace, "{");
        self.skip_newlines();

        let mut variants = Vec::new();
        while !self.check(TokenKind::RBrace) && !self.at_end() {
            let vstart = self.peek().start;
            let vname = self.parse_identifier();
            if vname.is_none() {
                break;
            }
            let vname = vname.unwrap();
            let fields = if self.check(TokenKind::LParen) {
                match self.parse_params() {
                    Some(f) => f,
                    Option::None => break,
                }
            } else {
                Vec::new()
            };
            variants.push(EnumVariant {
                name: vname,
                fields,
                span: self.span(vstart),
            });
            if self.check(TokenKind::Comma) {
                self.advance();
            }
            self.skip_newlines();
        }

        self.consume(TokenKind::RBrace, "}");
        Some(EnumDef {
            is_public,
            name,
            variants,
        })
    }
}

// ---------------------------------------------------------------------------
// Function / method definitions
// ---------------------------------------------------------------------------

impl Parser {
    fn parse_method_def_body(
        &mut self,
        type_name: Spanned<String>,
        method_name: Spanned<String>,
    ) -> Option<MethodDef> {
        let params = self.parse_params()?;

        let return_ty = if self.check(TokenKind::Arrow) {
            self.advance();
            Some(self.parse_type()?)
        } else {
            Option::None
        };

        let body = self.parse_block()?;

        Some(MethodDef {
            type_name,
            name: method_name,
            params,
            return_ty,
            body,
        })
    }

    fn parse_params(&mut self) -> Option<Vec<Param>> {
        self.consume(TokenKind::LParen, "(");

        let mut params = Vec::new();
        while !self.check(TokenKind::RParen) && !self.at_end() {
            let param = self.parse_param()?;
            params.push(param);
            if !self.check(TokenKind::RParen) {
                self.consume(TokenKind::Comma, ",");
            }
        }

        self.consume(TokenKind::RParen, ")");
        Some(params)
    }

    fn parse_param(&mut self) -> Option<Param> {
        let start = self.peek().start;

        let mut is_rest = false;
        if self.check(TokenKind::Ellipsis) {
            self.advance();
            is_rest = true;
        }

        let name = self.parse_identifier()?;

        let ty = if self.check(TokenKind::Colon) {
            self.advance();
            Some(self.parse_type()?)
        } else {
            Option::None
        };

        let default = if self.check(TokenKind::Eq) {
            self.advance();
            let e = self.parse_expr()?;
            Some(Box::new(e))
        } else {
            Option::None
        };

        if !is_rest && self.check(TokenKind::Ellipsis) {
            self.advance();
            is_rest = true;
        }

        Some(Param {
            name,
            ty,
            default,
            is_rest,
            span: self.span(start),
        })
    }

    /// Converts call arguments back to function parameters when we discover
    /// that what looked like a call was actually a function definition.
    fn args_to_params(&mut self, args: &[Argument]) -> Option<Vec<Param>> {
        let mut params = Vec::new();
        for arg in args {
            if let ExprKind::Ident(ref ident_name) = arg.value.node {
                params.push(Param {
                    name: Spanned::new(ident_name.clone(), arg.value.span),
                    ty: Option::None,
                    default: Option::None,
                    is_rest: false,
                    span: arg.span,
                });
            } else if arg.name.is_some() {
                params.push(Param {
                    name: arg.name.clone().unwrap(),
                    ty: Option::None,
                    default: Some(Box::new(arg.value.clone())),
                    is_rest: false,
                    span: arg.span,
                });
            } else {
                self.add_error(
                    "expected identifier in parameter list".into(),
                    arg.value.span,
                );
                return None;
            }
        }
        Some(params)
    }
}

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

impl Parser {
    fn parse_type(&mut self) -> Option<Spanned<Type>> {
        let start = self.peek().start;

        let ty = match self.peek().kind {
            TokenKind::Ident => {
                let name = self.peek().value.clone();
                self.advance();
                // Check for generic args: Name<T>
                if self.check(TokenKind::Lt) {
                    self.advance();
                    let mut args = Vec::new();
                    while !self.check(TokenKind::Gt) && !self.at_end() {
                        let a = self.parse_type()?;
                        args.push(a);
                        if !self.check(TokenKind::Gt) {
                            self.consume(TokenKind::Comma, ",");
                        }
                    }
                    self.consume(TokenKind::Gt, ">");
                    Type::Generic { name, args }
                } else {
                    let mut ty = Type::Named(name.clone());

                    // Qualified type: ui.StatusCard -> NamedType{Name: "ui.StatusCard"}
                    if matches!(ty, Type::Named(_)) && self.check(TokenKind::Dot) {
                        if self.current + 1 < self.tokens.len() {
                            let next = &self.tokens[self.current + 1];
                            if next.kind == TokenKind::Ident
                                && first_rune(&next.value).is_uppercase()
                            {
                                self.advance(); // consume .
                                let qual_name = self.peek().value.clone();
                                self.advance(); // consume the qualified name
                                ty = Type::Named(format!("{}.{}", name, qual_name));
                            }
                        }
                    }

                    ty
                }
            }

            TokenKind::LBracket => {
                self.advance();
                let inner = self.parse_type()?;
                self.consume(TokenKind::RBracket, "]");
                Type::List(Box::new(inner))
            }

            TokenKind::LBrace => {
                self.advance();
                self.skip_newlines();
                let key = self.parse_type()?;
                self.consume(TokenKind::Colon, ":");
                let value = self.parse_type()?;
                // Skip additional fields (record types).
                while self.check(TokenKind::Comma) {
                    self.advance();
                    self.skip_newlines();
                    if self.check(TokenKind::RBrace) {
                        break;
                    }
                    let _ = self.parse_type()?;
                    self.consume(TokenKind::Colon, ":");
                    let _ = self.parse_type()?;
                }
                self.skip_newlines();
                self.consume(TokenKind::RBrace, "}");
                Type::Map {
                    key: Box::new(key),
                    value: Box::new(value),
                }
            }

            TokenKind::LParen => {
                self.advance();
                let mut param_types = Vec::new();
                while !self.check(TokenKind::RParen) && !self.at_end() {
                    let pt = self.parse_type()?;
                    param_types.push(pt);
                    if !self.check(TokenKind::RParen) {
                        self.consume(TokenKind::Comma, ",");
                    }
                }
                self.consume(TokenKind::RParen, ")");
                self.consume(TokenKind::Arrow, "->");
                let ret = self.parse_type()?;
                Type::Function {
                    params: param_types,
                    ret: Box::new(ret),
                }
            }

            _ => {
                let tok = self.peek_cloned();
                self.add_error("expected type".into(), Span::new(tok.start, tok.end));
                return None;
            }
        };

        // Check for union: Type | Other
        if self.check(TokenKind::Pipe) {
            let mut variants = vec![Spanned::new(ty, self.span(start))];
            while self.check(TokenKind::Pipe) {
                self.advance();
                let v = self.parse_type()?;
                variants.push(v);
            }
            return Some(Spanned::new(Type::Union(variants), self.span(start)));
        }

        Some(Spanned::new(ty, self.span(start)))
    }
}

// ---------------------------------------------------------------------------
// Statements
// ---------------------------------------------------------------------------

impl Parser {
    fn parse_statement(&mut self) -> Option<Statement> {
        let start = self.peek().start;

        match self.peek().kind {
            TokenKind::If => {
                self.advance();
                let if_res = self.parse_if_statement()?;
                Some(Spanned::new(
                    StmtKind::If(IfStmt {
                        condition: if_res.condition,
                        then_branch: if_res.then_branch,
                        else_branch: if_res.else_branch,
                    }),
                    self.span(start),
                ))
            }

            TokenKind::For => {
                self.advance();
                let fs = self.parse_for_statement()?;
                Some(Spanned::new(StmtKind::For(fs), self.span(start)))
            }

            TokenKind::While => {
                self.advance();
                let ws = self.parse_while_statement()?;
                Some(Spanned::new(StmtKind::While(ws), self.span(start)))
            }

            TokenKind::Return => {
                self.advance();
                let rs = self.parse_return_statement()?;
                Some(Spanned::new(StmtKind::Return(rs), self.span(start)))
            }

            TokenKind::Match => {
                self.advance();
                let me = self.parse_match_expr()?;
                Some(Spanned::new(
                    StmtKind::Match(MatchStmt {
                        subject: me.subject,
                        arms: me.arms,
                    }),
                    self.span(start),
                ))
            }

            TokenKind::Try => {
                self.advance();
                let ts = self.parse_try_statement()?;
                Some(Spanned::new(StmtKind::Try(ts), self.span(start)))
            }

            TokenKind::Defer => {
                self.advance();
                let expr = self.parse_expr()?;
                Some(Spanned::new(
                    StmtKind::Defer(DeferStmt {
                        value: Box::new(expr),
                    }),
                    self.span(start),
                ))
            }

            TokenKind::Errdefer => {
                self.advance();
                let expr = self.parse_expr()?;
                Some(Spanned::new(
                    StmtKind::ErrDefer(ErrDeferStmt {
                        value: Box::new(expr),
                    }),
                    self.span(start),
                ))
            }

            TokenKind::Step => {
                self.advance();
                let ss = self.parse_step_statement()?;
                Some(Spanned::new(StmtKind::Step(ss), self.span(start)))
            }

            TokenKind::Assert => {
                self.advance();
                let a = self.parse_assert_stmt()?;
                Some(Spanned::new(StmtKind::Assert(a), self.span(start)))
            }

            TokenKind::Let | TokenKind::Const => {
                let is_const = self.peek().kind == TokenKind::Const;
                self.advance();
                let ls = self.parse_let_statement(is_const)?;
                Some(Spanned::new(StmtKind::Let(ls), self.span(start)))
            }

            TokenKind::Break => {
                self.advance();
                Some(Spanned::new(StmtKind::Break, self.span(start)))
            }

            TokenKind::Continue => {
                self.advance();
                Some(Spanned::new(StmtKind::Continue, self.span(start)))
            }

            TokenKind::Spawn | TokenKind::Async => {
                let expr = self.parse_expr()?;
                self.parse_statement_rest(expr)
            }

            _ => {
                let expr = self.parse_expr()?;
                self.parse_statement_rest(expr)
            }
        }
    }

    /// Parses: `let name [: Type] = expr` / `const name [: Type] = expr`
    fn parse_let_statement(&mut self, is_const: bool) -> Option<LetStmt> {
        if !self.check(TokenKind::Ident) {
            let kw = if is_const { "const" } else { "let" };
            let tok = self.peek_cloned();
            self.add_error(
                format!("expected variable name after '{}'", kw),
                Span::new(tok.start, tok.end),
            );
            return None;
        }
        let name_tok = self.peek_cloned();
        let name = Spanned::new(name_tok.value.clone(), Span::new(name_tok.start, name_tok.end));
        self.advance();

        // Optional type annotation: let x: int = ...
        let type_ann = if self.check(TokenKind::Colon) {
            self.advance();
            Some(self.parse_type()?)
        } else {
            Option::None
        };

        if !self.consume(TokenKind::Eq, "'='") {
            return None;
        }

        let value = self.parse_expr()?;

        Some(LetStmt {
            name,
            type_ann,
            value: Box::new(value),
            is_const,
        })
    }

    /// Takes an already-parsed expression and determines whether the overall
    /// statement is an assignment, compound assignment, multi-assignment, or
    /// expression statement.
    fn parse_statement_rest(&mut self, first_expr: Expr) -> Option<Statement> {
        let start = first_expr.span.start;

        // Compound assignment: += -= *= /= %= &= |= ^= <<= >>=
        let compound_op = match self.peek().kind {
            TokenKind::PlusEq => Some(BinaryOp::Add),
            TokenKind::MinusEq => Some(BinaryOp::Sub),
            TokenKind::StarEq => Some(BinaryOp::Mul),
            TokenKind::SlashEq => Some(BinaryOp::Div),
            TokenKind::PercentEq => Some(BinaryOp::Mod),
            TokenKind::AmpEq => Some(BinaryOp::BitAnd),
            TokenKind::PipeEq => Some(BinaryOp::BitOr),
            TokenKind::CaretEq => Some(BinaryOp::BitXor),
            TokenKind::ShlEq => Some(BinaryOp::Shl),
            TokenKind::ShrEq => Some(BinaryOp::Shr),
            _ => Option::None,
        };
        if let Some(op) = compound_op {
            let op_span = self.current_span();
            self.advance();
            let rhs = self.parse_expr()?;
            let target = self.expr_to_assign_target(&first_expr)?;
            // Desugar: x += y  =>  x = x + y
            let value = Spanned::new(
                ExprKind::Binary(BinaryExpr {
                    left: Box::new(first_expr),
                    op: Spanned::new(op, op_span),
                    right: Box::new(rhs),
                }),
                self.span(start),
            );
            return Some(Spanned::new(
                StmtKind::Assign(AssignStmt {
                    targets: vec![target],
                    value: Box::new(value),
                }),
                self.span(start),
            ));
        }

        // Simple assignment: expr = value
        if self.check(TokenKind::Eq) {
            self.advance();
            let value = self.parse_expr()?;
            let target = self.expr_to_assign_target(&first_expr)?;
            return Some(Spanned::new(
                StmtKind::Assign(AssignStmt {
                    targets: vec![target],
                    value: Box::new(value),
                }),
                self.span(start),
            ));
        }

        // Multi-assignment: a, b = ...
        if self.check(TokenKind::Comma) {
            let mut targets = Vec::new();
            let t = self.expr_to_assign_target(&first_expr)?;
            targets.push(t);

            while self.check(TokenKind::Comma) {
                self.advance();
                let expr = self.parse_expr()?;
                let tgt = self.expr_to_assign_target(&expr)?;
                targets.push(tgt);
            }

            self.consume(TokenKind::Eq, "=");
            let value = self.parse_expr()?;
            return Some(Spanned::new(
                StmtKind::Assign(AssignStmt {
                    targets,
                    value: Box::new(value),
                }),
                self.span(start),
            ));
        }

        // Expression statement.
        Some(Spanned::new(
            StmtKind::Expr(ExprStmt {
                value: Box::new(first_expr),
            }),
            self.span(start),
        ))
    }

    fn expr_to_assign_target(&mut self, expr: &Expr) -> Option<AssignTarget> {
        let path = self.expr_to_assign_path(expr)?;
        Some(AssignTarget {
            path,
            ty: Option::None,
        })
    }

    fn expr_to_assign_path(&mut self, expr: &Expr) -> Option<AssignPath> {
        match &expr.node {
            ExprKind::Ident(name) => Some(AssignPath::Ident(Spanned::new(
                name.clone(),
                expr.span,
            ))),
            ExprKind::Field(fe) => {
                let obj = self.expr_to_assign_path(&fe.object)?;
                Some(AssignPath::Field {
                    object: Box::new(obj),
                    field: fe.field.clone(),
                })
            }
            ExprKind::Index(ie) => {
                let obj = self.expr_to_assign_path(&ie.object)?;
                Some(AssignPath::Index {
                    object: Box::new(obj),
                    index: Box::new((*ie.index).clone()),
                })
            }
            _ => {
                self.add_error("expected assignable expression".into(), expr.span);
                None
            }
        }
    }
}

// ---------------------------------------------------------------------------
// If / For / While / Return / Try
// ---------------------------------------------------------------------------

/// Internal struct for parse_if_statement results.
struct ParsedIf {
    condition: Box<Expr>,
    then_branch: Block,
    else_branch: ElseBranch,
}

impl Parser {
    fn parse_if_statement(&mut self) -> Option<ParsedIf> {
        let cond = self.parse_expr()?;
        let then_branch = self.parse_block()?;

        let else_branch = if self.check(TokenKind::Else) {
            self.advance();
            if self.check(TokenKind::If) {
                self.advance();
                let else_start = self.previous.start;
                let inner = self.parse_if_statement()?;
                let inner_stmt = Spanned::new(
                    IfStmt {
                        condition: inner.condition,
                        then_branch: inner.then_branch,
                        else_branch: inner.else_branch,
                    },
                    self.span(else_start),
                );
                ElseBranch::ElseIf(Box::new(inner_stmt))
            } else {
                let block = self.parse_block()?;
                ElseBranch::Block(block)
            }
        } else {
            ElseBranch::None
        };

        Some(ParsedIf {
            condition: Box::new(cond),
            then_branch,
            else_branch,
        })
    }

    fn parse_for_statement(&mut self) -> Option<ForStmt> {
        let pat = self.parse_for_pattern()?;
        self.consume(TokenKind::In, "in");
        let iter = self.parse_expr()?;
        let body = self.parse_block()?;
        Some(ForStmt {
            pattern: pat,
            iterator: Box::new(iter),
            body,
        })
    }

    fn parse_for_pattern(&mut self) -> Option<ForPattern> {
        let first = self.parse_identifier()?;
        if self.check(TokenKind::Comma) {
            self.advance();
            let second = self.parse_identifier()?;
            Some(ForPattern::Pair(first, second))
        } else {
            Some(ForPattern::Single(first))
        }
    }

    fn parse_while_statement(&mut self) -> Option<WhileStmt> {
        let cond = self.parse_expr()?;
        let body = self.parse_block()?;
        Some(WhileStmt {
            condition: Box::new(cond),
            body,
        })
    }

    fn parse_return_statement(&mut self) -> Option<ReturnStmt> {
        // No value if followed by newline, }, or EOF.
        if self.check(TokenKind::Newline) || self.check(TokenKind::RBrace) || self.at_end() {
            return Some(ReturnStmt { values: Vec::new() });
        }

        let mut values = Vec::new();
        let first = self.parse_expr()?;
        values.push(first);
        while self.check(TokenKind::Comma) {
            self.advance();
            let v = self.parse_expr()?;
            values.push(v);
        }
        Some(ReturnStmt { values })
    }

    fn parse_step_statement(&mut self) -> Option<StepStmt> {
        self.parse_step_statement_with_decorators(Vec::new())
    }

    fn parse_step_statement_with_decorators(
        &mut self,
        decorators: Vec<Decorator>,
    ) -> Option<StepStmt> {
        // Expect a string literal for the step name
        if self.peek().kind != TokenKind::String
            && self.peek().kind != TokenKind::InterpolatedString
        {
            let tok = self.peek_cloned();
            self.add_error(
                "expected string literal for step name".into(),
                Span::new(tok.start, tok.end),
            );
            return None;
        }
        let name = self.peek().value.clone();
        let name_span = self.current_span();
        self.advance();

        // Parse block body
        self.consume(TokenKind::LBrace, "{");
        self.skip_newlines();

        // Parse lifecycle hooks at top of step body
        let mut hooks = Vec::new();
        while self.check(TokenKind::Onerror)
            || self.check(TokenKind::Onsuccess)
            || self.check(TokenKind::Oncancel)
        {
            if let Some(hook) = self.parse_lifecycle_hook() {
                hooks.push(hook);
            }
            self.skip_newlines();
        }

        let mut stmts = Vec::new();
        while !self.check(TokenKind::RBrace) && !self.at_end() {
            if let Some(stmt) = self.parse_statement() {
                stmts.push(stmt);
            } else {
                self.advance();
            }
            self.skip_newlines();
        }
        self.consume(TokenKind::RBrace, "}");

        Some(StepStmt {
            name: Spanned::new(name, name_span),
            body: stmts,
            decorators,
            hooks,
        })
    }

    fn parse_lifecycle_hook(&mut self) -> Option<LifecycleHook> {
        let kind;
        let mut err_name = String::new();
        let mut arg_name = String::new();

        match self.peek().kind {
            TokenKind::Onerror => {
                self.advance();
                kind = LifecycleHookKind::Onerror;
                // Parse optional error variable name (accept `err` keyword as variable name)
                if self.check(TokenKind::Ident) || self.check(TokenKind::Err) {
                    err_name = self.peek().value.clone();
                    self.advance();
                } else {
                    err_name = "err".into();
                }
            }
            TokenKind::Onsuccess => {
                self.advance();
                kind = LifecycleHookKind::Onsuccess;
                // Parse optional result variable name
                if self.check(TokenKind::Ident) {
                    arg_name = self.peek().value.clone();
                    self.advance();
                }
            }
            TokenKind::Oncancel => {
                self.advance();
                kind = LifecycleHookKind::Oncancel;
            }
            _ => return None,
        }

        let body = self.parse_block()?;

        Some(LifecycleHook {
            kind,
            err_name,
            arg_name,
            body,
        })
    }

    fn parse_try_statement(&mut self) -> Option<TryStmt> {
        let body = self.parse_block()?;
        self.consume(TokenKind::Catch, "catch");
        let error_name = self.parse_identifier()?;
        let catch_body = self.parse_block()?;
        Some(TryStmt {
            body,
            error_name,
            catch_body,
        })
    }
}

// ---------------------------------------------------------------------------
// Blocks
// ---------------------------------------------------------------------------

impl Parser {
    fn parse_block(&mut self) -> Option<Block> {
        let start = self.peek().start;
        self.consume(TokenKind::LBrace, "{");
        self.skip_newlines();

        let mut stmts = Vec::new();
        while !self.check(TokenKind::RBrace) && !self.at_end() {
            if let Some(stmt) = self.parse_statement() {
                stmts.push(stmt);
            } else {
                self.advance(); // error recovery
            }
            self.skip_newlines();
        }

        self.consume(TokenKind::RBrace, "}");
        Some(Block {
            statements: stmts,
            span: self.span(start),
        })
    }
}

// ---------------------------------------------------------------------------
// Expressions -- Pratt parser
// ---------------------------------------------------------------------------

impl Parser {
    fn parse_expr(&mut self) -> Option<Expr> {
        self.parse_expr_precedence(Precedence::None)
    }

    fn parse_expr_rest(&mut self, left: Expr) -> Option<Expr> {
        self.parse_expr_rest_precedence(left, Precedence::None)
    }

    fn parse_expr_rest_precedence(
        &mut self,
        left: Expr,
        min_prec: Precedence,
    ) -> Option<Expr> {
        let mut cur = left;
        while !self.at_end() {
            let prec = precedence_of(self.peek().kind);
            if prec <= min_prec {
                break;
            }
            match self.parse_infix(cur.clone(), prec) {
                Some(next) => cur = next,
                Option::None => return Some(cur), // return what we have
            }
        }
        Some(cur)
    }

    fn parse_expr_precedence(&mut self, min_prec: Precedence) -> Option<Expr> {
        let left = self.parse_prefix()?;
        self.parse_expr_rest_precedence(left, min_prec)
    }
}

// ---------------------------------------------------------------------------
// Prefix (primary) expressions
// ---------------------------------------------------------------------------

impl Parser {
    fn parse_prefix(&mut self) -> Option<Expr> {
        let start = self.peek().start;

        match self.peek().kind {
            // Integer literal
            TokenKind::Int => {
                let raw = self.peek().value.clone();
                self.advance();
                let val = parse_int(&raw);
                Some(Spanned::new(
                    ExprKind::Literal(Literal::Int(val)),
                    self.span(start),
                ))
            }

            // Float literal
            TokenKind::Float => {
                let raw = self.peek().value.clone();
                self.advance();
                let val = parse_float(&raw);
                Some(Spanned::new(
                    ExprKind::Literal(Literal::Float(val)),
                    self.span(start),
                ))
            }

            // String literal
            TokenKind::String => {
                let s = self.peek().value.clone();
                self.advance();
                Some(Spanned::new(
                    ExprKind::Literal(Literal::String(s)),
                    self.span(start),
                ))
            }

            // Interpolated string
            TokenKind::InterpolatedString => {
                let raw = self.peek().value.clone();
                self.advance();
                let parts = parse_interpolated_string_parts(&raw);
                Some(Spanned::new(
                    ExprKind::Literal(Literal::InterpolatedString(parts)),
                    self.span(start),
                ))
            }

            // Triple-quoted string (treated as plain string)
            TokenKind::TripleQuoteString => {
                let s = self.peek().value.clone();
                self.advance();
                Some(Spanned::new(
                    ExprKind::Literal(Literal::String(s)),
                    self.span(start),
                ))
            }

            // Boolean true
            TokenKind::True => {
                self.advance();
                Some(Spanned::new(
                    ExprKind::Literal(Literal::Bool(true)),
                    self.span(start),
                ))
            }

            // Boolean false
            TokenKind::False => {
                self.advance();
                Some(Spanned::new(
                    ExprKind::Literal(Literal::Bool(false)),
                    self.span(start),
                ))
            }

            // none
            TokenKind::None => {
                self.advance();
                Some(Spanned::new(ExprKind::None, self.span(start)))
            }

            // some(expr)
            TokenKind::Some => {
                self.advance();
                self.consume(TokenKind::LParen, "(");
                let inner = self.parse_expr()?;
                self.consume(TokenKind::RParen, ")");
                Some(Spanned::new(
                    ExprKind::Some(SomeExpr {
                        inner: Box::new(inner),
                    }),
                    self.span(start),
                ))
            }

            // Identifier (possibly lambda: x => ...)
            TokenKind::Ident => {
                let name = self.peek().value.clone();
                self.advance();

                // Arrow lambda: x => expr
                if self.check(TokenKind::FatArrow) {
                    self.advance();
                    let body = self.parse_expr()?;
                    let param = Param {
                        name: Spanned::new(name, self.span(start)),
                        ty: Option::None,
                        default: Option::None,
                        is_rest: false,
                        span: self.span(start),
                    };
                    return Some(Spanned::new(
                        ExprKind::Lambda(LambdaExpr {
                            params: vec![param],
                            body: LambdaBody::Expr(Box::new(body)),
                        }),
                        self.span(start),
                    ));
                }

                // Instance creation: User { name = "Alice" }
                let first_char = first_rune(&name);
                if first_char.is_uppercase() && self.check(TokenKind::LBrace) {
                    return self.parse_instance(&name, start);
                }

                Some(Spanned::new(
                    ExprKind::Ident(name),
                    self.span(start),
                ))
            }

            // Unary minus
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

            // Bitwise NOT (~)
            TokenKind::Tilde => {
                self.advance();
                let operand = self.parse_expr_precedence(Precedence::Unary)?;
                Some(Spanned::new(
                    ExprKind::Unary(UnaryExpr {
                        op: Spanned::new(UnaryOp::BitNot, self.span(start)),
                        operand: Box::new(operand),
                    }),
                    self.span(start),
                ))
            }

            // Unary not
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

            // Parenthesized expression or lambda
            TokenKind::LParen => self.parse_paren_or_lambda(start),

            // List literal
            TokenKind::LBracket => self.parse_list(start),

            // Map or block expression
            TokenKind::LBrace => self.parse_map_or_block(start),

            // If expression
            TokenKind::If => {
                self.advance();
                let if_res = self.parse_if_statement()?;
                Some(Spanned::new(
                    ExprKind::If(IfExpr {
                        if_stmt: IfStmt {
                            condition: if_res.condition,
                            then_branch: if_res.then_branch,
                            else_branch: if_res.else_branch,
                        },
                    }),
                    self.span(start),
                ))
            }

            // Match expression
            TokenKind::Match => {
                self.advance();
                let me = self.parse_match_expr()?;
                Some(Spanned::new(
                    ExprKind::Match(me),
                    self.span(start),
                ))
            }

            // Async block
            TokenKind::Async => {
                self.advance();
                let block = self.parse_block()?;
                Some(Spanned::new(
                    ExprKind::Async(AsyncExpr { body: block }),
                    self.span(start),
                ))
            }

            // Spawn block
            TokenKind::Spawn => {
                self.advance();
                let block = self.parse_block()?;
                Some(Spanned::new(
                    ExprKind::Spawn(SpawnExpr { body: block }),
                    self.span(start),
                ))
            }

            // Agent expression: `agent { provider: openai, system: "..." }`
            TokenKind::Agent => {
                self.advance();
                // If followed by `{`, parse as agent expression (dynamic creation).
                // Otherwise, treat as identifier.
                if self.check(TokenKind::LBrace) {
                    self.advance(); // consume `{`
                    self.skip_newlines();
                    let mut fields = Vec::new();
                    while !self.check(TokenKind::RBrace) && !self.at_end() {
                        let key = match self.parse_identifier() {
                            Some(k) => k,
                            std::option::Option::None => break,
                        };
                        self.consume(TokenKind::Colon, ":");
                        let val = match self.parse_expr() {
                            Some(v) => v,
                            std::option::Option::None => break,
                        };
                        fields.push(AgentField { key, value: val });
                        if self.check(TokenKind::Comma) {
                            self.advance();
                        }
                        self.skip_newlines();
                    }
                    self.consume(TokenKind::RBrace, "}");
                    Some(Spanned::new(
                        ExprKind::Agent(AgentExpr { fields }),
                        self.span(start),
                    ))
                } else {
                    // Bare `agent` used as identifier.
                    Some(Spanned::new(
                        ExprKind::Ident("agent".into()),
                        self.span(start),
                    ))
                }
            }

            // Select expression
            TokenKind::Select => {
                self.advance();
                let sel = self.parse_select_expr()?;
                Some(Spanned::new(
                    ExprKind::Select(sel),
                    self.span(start),
                ))
            }

            // Contextual keywords usable as identifiers in expressions
            TokenKind::Err => {
                self.advance();
                Some(Spanned::new(
                    ExprKind::Ident("err".into()),
                    self.span(start),
                ))
            }

            TokenKind::Ok => {
                self.advance();
                Some(Spanned::new(
                    ExprKind::Ident("ok".into()),
                    self.span(start),
                ))
            }

            _ => {
                let tok = self.peek_cloned();
                self.add_error(
                    "expected expression".into(),
                    Span::new(tok.start, tok.end),
                );
                None
            }
        }
    }
}

// ---------------------------------------------------------------------------
// Infix expressions
// ---------------------------------------------------------------------------

impl Parser {
    fn parse_infix(&mut self, left: Expr, prec: Precedence) -> Option<Expr> {
        let start = left.span.start;
        let op_span = self.current_span();

        match self.peek().kind {
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
            | TokenKind::Or
            | TokenKind::Amp
            | TokenKind::Caret
            | TokenKind::Shl
            | TokenKind::Shr => {
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

            // Bitwise OR (|) -- handled as binary op
            TokenKind::Pipe => {
                self.advance();
                let right = self.parse_expr_precedence(prec)?;
                Some(Spanned::new(
                    ExprKind::Binary(BinaryExpr {
                        left: Box::new(left),
                        op: Spanned::new(BinaryOp::BitOr, op_span),
                        right: Box::new(right),
                    }),
                    self.span(start),
                ))
            }

            // Pipe (|>)
            TokenKind::PipeArrow => {
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

            // Range ..
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

            // Range ..=
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

            // Call: expr(args)
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

            // Index: expr[index]
            TokenKind::LBracket => {
                self.advance();
                let idx = self.parse_expr()?;
                self.consume(TokenKind::RBracket, "]");
                Some(Spanned::new(
                    ExprKind::Index(IndexExpr {
                        object: Box::new(left),
                        index: Box::new(idx),
                    }),
                    self.span(start),
                ))
            }

            // Field access or method call: expr.field or expr.method(args)
            TokenKind::Dot => {
                self.advance();
                let field = self.parse_identifier()?;
                if self.check(TokenKind::LParen) {
                    let args = self.parse_call_args()?;
                    return Some(Spanned::new(
                        ExprKind::MethodCall(MethodCallExpr {
                            receiver: Box::new(left),
                            method: field,
                            args,
                        }),
                        self.span(start),
                    ));
                }
                // Qualified instance: ui.StatusCard { ... }
                if first_rune(&field.node).is_uppercase() && self.check(TokenKind::LBrace) {
                    if let ExprKind::Ident(ref ident_name) = left.node {
                        let qual_name = format!("{}.{}", ident_name, field.node);
                        return self.parse_instance(&qual_name, start);
                    }
                }

                Some(Spanned::new(
                    ExprKind::Field(FieldExpr {
                        object: Box::new(left),
                        field,
                    }),
                    self.span(start),
                ))
            }

            // Error propagation: expr?
            TokenKind::Question => {
                self.advance();
                Some(Spanned::new(
                    ExprKind::Propagate(PropagateExpr {
                        inner: Box::new(left),
                    }),
                    self.span(start),
                ))
            }

            // Orelse: expr orelse default
            TokenKind::Orelse => {
                self.advance();
                let def = self.parse_expr_precedence(prec)?;
                Some(Spanned::new(
                    ExprKind::Orelse(OrelseExpr {
                        left: Box::new(left),
                        default: Box::new(def),
                    }),
                    self.span(start),
                ))
            }

            _ => Some(left),
        }
    }

    fn parse_binary_op(&mut self) -> Option<BinaryOp> {
        let op = match self.peek().kind {
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
            TokenKind::Amp => BinaryOp::BitAnd,
            TokenKind::Caret => BinaryOp::BitXor,
            TokenKind::Shl => BinaryOp::Shl,
            TokenKind::Shr => BinaryOp::Shr,
            _ => return None,
        };
        self.advance();
        Some(op)
    }
}

// ---------------------------------------------------------------------------
// Call arguments
// ---------------------------------------------------------------------------

impl Parser {
    fn parse_call_args(&mut self) -> Option<Vec<Argument>> {
        self.consume(TokenKind::LParen, "(");

        let mut args = Vec::new();
        while !self.check(TokenKind::RParen) && !self.at_end() {
            let arg_start = self.peek().start;

            // Check for named argument: name = value or name: value
            if self.peek().kind == TokenKind::Ident {
                let ident = self.parse_identifier()?;

                if self.check(TokenKind::Eq) || self.check(TokenKind::Colon) {
                    self.advance();
                    let val = self.parse_expr()?;
                    args.push(Argument {
                        name: Some(ident),
                        value: val,
                        span: self.span(arg_start),
                    });
                    if !self.check(TokenKind::RParen) {
                        self.consume(TokenKind::Comma, ",");
                    }
                    continue;
                }

                // Not a named arg -- build an expression from the identifier
                // and continue parsing the rest of the expression.
                let ident_expr = Spanned::new(ExprKind::Ident(ident.node.clone()), ident.span);
                let val = self.parse_expr_rest(ident_expr)?;
                args.push(Argument {
                    name: Option::None,
                    value: val,
                    span: self.span(arg_start),
                });
                if !self.check(TokenKind::RParen) {
                    self.consume(TokenKind::Comma, ",");
                }
                continue;
            }

            // Positional argument.
            let val = self.parse_expr()?;
            args.push(Argument {
                name: Option::None,
                value: val,
                span: self.span(arg_start),
            });
            if !self.check(TokenKind::RParen) {
                self.consume(TokenKind::Comma, ",");
            }
        }

        self.consume(TokenKind::RParen, ")");
        Some(args)
    }
}

// ---------------------------------------------------------------------------
// Parenthesized expr / lambda
// ---------------------------------------------------------------------------

impl Parser {
    fn parse_paren_or_lambda(&mut self, start: usize) -> Option<Expr> {
        self.advance(); // consume (

        // Empty parens: () => expr  or  () { block }  or  unit
        if self.check(TokenKind::RParen) {
            self.advance();

            if self.check(TokenKind::FatArrow) {
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

            if self.check(TokenKind::LBrace) {
                let block = self.parse_block()?;
                return Some(Spanned::new(
                    ExprKind::Lambda(LambdaExpr {
                        params: Vec::new(),
                        body: LambdaBody::Block(block),
                    }),
                    self.span(start),
                ));
            }

            // Empty parens as unit -- return empty list for now.
            return Some(Spanned::new(
                ExprKind::List(ListExpr { elems: Vec::new() }),
                self.span(start),
            ));
        }

        // Parse first expression inside parens.
        let first = self.parse_expr()?;

        // Comma or colon after first expression => lambda parameter list.
        if self.check(TokenKind::Comma) || self.check(TokenKind::Colon) {
            let mut params = Vec::new();
            let fp = self.expr_to_param(&first)?;
            params.push(fp);

            while self.check(TokenKind::Comma) {
                self.advance();
                let expr = self.parse_expr()?;
                let param = self.expr_to_param(&expr)?;
                params.push(param);
            }

            self.consume(TokenKind::RParen, ")");

            if self.check(TokenKind::FatArrow) {
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

            let block = self.parse_block()?;
            return Some(Spanned::new(
                ExprKind::Lambda(LambdaExpr {
                    params,
                    body: LambdaBody::Block(block),
                }),
                self.span(start),
            ));
        }

        self.consume(TokenKind::RParen, ")");

        // Single-param lambda: (x) => expr  or  (x) { block }
        if self.check(TokenKind::FatArrow) {
            self.advance();
            let param = self.expr_to_param(&first)?;
            let body = self.parse_expr()?;
            return Some(Spanned::new(
                ExprKind::Lambda(LambdaExpr {
                    params: vec![param],
                    body: LambdaBody::Expr(Box::new(body)),
                }),
                self.span(start),
            ));
        }

        if self.check(TokenKind::LBrace) {
            let param = self.expr_to_param(&first)?;
            let block = self.parse_block()?;
            return Some(Spanned::new(
                ExprKind::Lambda(LambdaExpr {
                    params: vec![param],
                    body: LambdaBody::Block(block),
                }),
                self.span(start),
            ));
        }

        // Just a parenthesized expression.
        Some(Spanned::new(
            ExprKind::Paren(ParenExpr {
                inner: Box::new(first),
            }),
            self.span(start),
        ))
    }

    fn expr_to_param(&mut self, expr: &Expr) -> Option<Param> {
        if let ExprKind::Ident(ref name) = expr.node {
            return Some(Param {
                name: Spanned::new(name.clone(), expr.span),
                ty: Option::None,
                default: Option::None,
                is_rest: false,
                span: expr.span,
            });
        }
        self.add_error("expected identifier for parameter".into(), expr.span);
        None
    }
}

// ---------------------------------------------------------------------------
// List literal
// ---------------------------------------------------------------------------

impl Parser {
    fn parse_list(&mut self, start: usize) -> Option<Expr> {
        self.advance(); // consume [

        let mut elems = Vec::new();
        while !self.check(TokenKind::RBracket) && !self.at_end() {
            let e = self.parse_expr()?;
            elems.push(e);
            if !self.check(TokenKind::RBracket) {
                self.consume(TokenKind::Comma, ",");
            }
        }

        self.consume(TokenKind::RBracket, "]");
        Some(Spanned::new(
            ExprKind::List(ListExpr { elems }),
            self.span(start),
        ))
    }
}

// ---------------------------------------------------------------------------
// Map or block expression
// ---------------------------------------------------------------------------

impl Parser {
    fn parse_map_or_block(&mut self, start: usize) -> Option<Expr> {
        self.advance(); // consume {
        self.skip_newlines();

        // Empty braces -> empty map
        if self.check(TokenKind::RBrace) {
            self.advance();
            return Some(Spanned::new(
                ExprKind::Map(MapExpr {
                    entries: Vec::new(),
                }),
                self.span(start),
            ));
        }

        // Parse first expression and see if it's followed by a colon (map) or not (block).
        let first_expr = self.parse_expr()?;

        if self.check(TokenKind::Colon) {
            // Map literal.
            self.advance();
            let first_value = self.parse_expr()?;
            let mut entries = vec![MapEntry {
                key: first_expr,
                value: first_value,
            }];

            while self.check(TokenKind::Comma) {
                self.advance();
                self.skip_newlines();
                if self.check(TokenKind::RBrace) {
                    break;
                }
                let k = self.parse_expr()?;
                self.consume(TokenKind::Colon, ":");
                let v = self.parse_expr()?;
                entries.push(MapEntry { key: k, value: v });
            }

            self.skip_newlines();
            self.consume(TokenKind::RBrace, "}");
            return Some(Spanned::new(
                ExprKind::Map(MapExpr { entries }),
                self.span(start),
            ));
        }

        // Block expression.
        let first_stmt = Spanned::new(
            StmtKind::Expr(ExprStmt {
                value: Box::new(first_expr),
            }),
            self.span(start),
        );
        let mut stmts = vec![first_stmt];

        self.skip_newlines();
        while !self.check(TokenKind::RBrace) && !self.at_end() {
            if let Some(stmt) = self.parse_statement() {
                stmts.push(stmt);
            } else {
                self.advance();
            }
            self.skip_newlines();
        }

        self.consume(TokenKind::RBrace, "}");
        Some(Spanned::new(
            ExprKind::Block(BlockExpr {
                body: Block {
                    statements: stmts,
                    span: self.span(start),
                },
            }),
            self.span(start),
        ))
    }
}

// ---------------------------------------------------------------------------
// Instance creation: User { name = "Alice" }
// ---------------------------------------------------------------------------

impl Parser {
    fn parse_instance(&mut self, type_name: &str, start: usize) -> Option<Expr> {
        self.advance(); // consume {
        self.skip_newlines();

        let mut fields = Vec::new();
        while !self.check(TokenKind::RBrace) && !self.at_end() {
            let field_start = self.peek().start;

            // Check for named field: name = value
            if self.peek().kind == TokenKind::Ident {
                let ident = self.parse_identifier()?;

                if self.check(TokenKind::Eq) {
                    self.advance();
                    let val = self.parse_expr()?;
                    fields.push(InstanceField {
                        name: Some(ident),
                        value: val,
                        span: self.span(field_start),
                    });
                } else {
                    // Positional/shorthand: just the identifier as value.
                    fields.push(InstanceField {
                        name: Option::None,
                        value: Spanned::new(
                            ExprKind::Ident(ident.node.clone()),
                            ident.span,
                        ),
                        span: self.span(field_start),
                    });
                }
            } else {
                let val = self.parse_expr()?;
                fields.push(InstanceField {
                    name: Option::None,
                    value: val,
                    span: self.span(field_start),
                });
            }

            if !self.check(TokenKind::RBrace) {
                if self.check(TokenKind::Comma) {
                    self.advance();
                }
                self.skip_newlines();
            }
        }

        self.consume(TokenKind::RBrace, "}");
        let name_spanned = Spanned::new(type_name.to_string(), self.span(start));
        Some(Spanned::new(
            ExprKind::Instance(InstanceExpr {
                type_name: name_spanned,
                fields,
            }),
            self.span(start),
        ))
    }
}

// ---------------------------------------------------------------------------
// Match expression
// ---------------------------------------------------------------------------

impl Parser {
    fn parse_match_expr(&mut self) -> Option<MatchExpr> {
        let subject = self.parse_expr()?;
        self.consume(TokenKind::LBrace, "{");
        self.skip_newlines();

        let mut arms = Vec::new();
        while !self.check(TokenKind::RBrace) && !self.at_end() {
            if let Some(arm) = self.parse_match_arm() {
                arms.push(arm);
            } else {
                self.advance(); // error recovery
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
        let start = self.peek().start;
        let pat = self.parse_pattern()?;

        // Optional guard: if condition
        let guard = if self.check(TokenKind::If) {
            self.advance();
            let g = self.parse_expr()?;
            Some(g)
        } else {
            Option::None
        };

        self.consume(TokenKind::FatArrow, "=>");

        let body = if self.check(TokenKind::LBrace) {
            let block = self.parse_block()?;
            MatchArmBody::Block(block)
        } else if matches!(
            self.peek().kind,
            TokenKind::Return | TokenKind::Break | TokenKind::Continue
        ) {
            let stmt = self.parse_statement()?;
            MatchArmBody::Block(Block {
                statements: vec![stmt.clone()],
                span: stmt.span,
            })
        } else {
            let expr = self.parse_expr()?;
            MatchArmBody::Expr(expr)
        };

        Some(MatchArm {
            pattern: pat,
            guard,
            body,
            span: self.span(start),
        })
    }

    fn parse_pattern(&mut self) -> Option<Spanned<Pattern>> {
        let start = self.peek().start;
        let first = self.parse_single_pattern()?;

        // Or-pattern: A | B | C
        if self.check(TokenKind::Pipe) {
            let mut patterns = vec![first];
            while self.check(TokenKind::Pipe) {
                self.advance();
                let next = self.parse_single_pattern()?;
                patterns.push(next);
            }
            return Some(Spanned::new(
                Pattern::Or(patterns),
                self.span(start),
            ));
        }

        Some(first)
    }

    fn parse_single_pattern(&mut self) -> Option<Spanned<Pattern>> {
        let start = self.peek().start;

        match self.peek().kind {
            TokenKind::Ident => {
                let name = self.peek().value.clone();
                self.advance();

                if name == "_" {
                    return Some(Spanned::new(Pattern::Wildcard, self.span(start)));
                }

                // Dotted pattern: Direction.North -> "DirectionNorth"
                if self.check(TokenKind::Dot) {
                    self.advance();
                    let field = self.parse_identifier()?;
                    let combined = format!("{}{}", name, field.node);
                    return Some(Spanned::new(
                        Pattern::Ident(combined),
                        self.span(start),
                    ));
                }

                // Constructor pattern: Some { value }
                if self.check(TokenKind::LBrace) {
                    self.advance();
                    let mut fields = Vec::new();
                    while !self.check(TokenKind::RBrace) && !self.at_end() {
                        let f = self.parse_identifier()?;
                        fields.push(f);
                        if !self.check(TokenKind::RBrace) {
                            self.consume(TokenKind::Comma, ",");
                        }
                    }
                    self.consume(TokenKind::RBrace, "}");
                    return Some(Spanned::new(
                        Pattern::Constructor { name, fields },
                        self.span(start),
                    ));
                }

                Some(Spanned::new(
                    Pattern::Ident(name),
                    self.span(start),
                ))
            }

            TokenKind::Int => {
                let raw = self.peek().value.clone();
                self.advance();
                let val = parse_int(&raw);

                // Range pattern: 1..5 or 1..=5
                if self.check(TokenKind::DotDot) || self.check(TokenKind::DotDotEq) {
                    let inclusive = self.check(TokenKind::DotDotEq);
                    self.advance();
                    if !self.check(TokenKind::Int) {
                        self.add_error(
                            "expected integer in range pattern".into(),
                            self.current_span(),
                        );
                        return None;
                    }
                    let end_raw = self.peek().value.clone();
                    self.advance();
                    let end_val = parse_int(&end_raw);
                    return Some(Spanned::new(
                        Pattern::Range {
                            start: Literal::Int(val),
                            end: Literal::Int(end_val),
                            inclusive,
                        },
                        self.span(start),
                    ));
                }

                Some(Spanned::new(
                    Pattern::Literal(Literal::Int(val)),
                    self.span(start),
                ))
            }

            TokenKind::String => {
                let s = self.peek().value.clone();
                self.advance();
                Some(Spanned::new(
                    Pattern::Literal(Literal::String(s)),
                    self.span(start),
                ))
            }

            TokenKind::True => {
                self.advance();
                Some(Spanned::new(
                    Pattern::Literal(Literal::Bool(true)),
                    self.span(start),
                ))
            }

            TokenKind::False => {
                self.advance();
                Some(Spanned::new(
                    Pattern::Literal(Literal::Bool(false)),
                    self.span(start),
                ))
            }

            _ => {
                let tok = self.peek_cloned();
                self.add_error("expected pattern".into(), Span::new(tok.start, tok.end));
                None
            }
        }
    }
}

// ---------------------------------------------------------------------------
// Select expression
// ---------------------------------------------------------------------------

impl Parser {
    fn parse_select_expr(&mut self) -> Option<SelectExpr> {
        self.consume(TokenKind::LBrace, "{");
        self.skip_newlines();

        let mut arms = Vec::new();
        let mut default = Option::None;

        while !self.check(TokenKind::RBrace) && !self.at_end() {
            if self.check(TokenKind::Default) {
                self.advance();
                self.consume(TokenKind::FatArrow, "=>");
                let block = self.parse_block()?;
                default = Some(block);
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
        let start = self.peek().start;
        let binding = self.parse_identifier()?;
        self.consume(TokenKind::From, "from");
        let channel = self.parse_expr()?;
        self.consume(TokenKind::FatArrow, "=>");

        let body = if self.check(TokenKind::LBrace) {
            let block = self.parse_block()?;
            MatchArmBody::Block(block)
        } else {
            let expr = self.parse_expr()?;
            MatchArmBody::Expr(expr)
        };

        Some(SelectArm {
            binding,
            channel,
            body,
            span: self.span(start),
        })
    }
}

// ---------------------------------------------------------------------------
// Agentic declarations
// ---------------------------------------------------------------------------

impl Parser {
    fn parse_import_decl(&mut self) -> Option<ImportDecl> {
        match self.peek().kind {
            TokenKind::String => {
                // Basic: import "io"
                let path = self.peek().value.clone();
                self.advance();
                Some(ImportDecl {
                    path,
                    alias: Option::None,
                    names: Vec::new(),
                    is_glob: false,
                })
            }

            TokenKind::LBrace => {
                // Selective: import { User, Post } from "models"
                self.advance(); // consume {
                self.skip_newlines();
                let mut names = Vec::new();
                while !self.check(TokenKind::RBrace) && !self.at_end() {
                    let name = self.parse_identifier()?;
                    names.push(name);
                    if self.check(TokenKind::Comma) {
                        self.advance();
                    }
                    self.skip_newlines();
                }
                self.consume(TokenKind::RBrace, "}");
                if !self.check(TokenKind::From) {
                    let tok = self.peek_cloned();
                    self.add_error(
                        "expected 'from' after import names".into(),
                        Span::new(tok.start, tok.end),
                    );
                    return None;
                }
                self.advance(); // consume from
                if self.peek().kind != TokenKind::String {
                    let tok = self.peek_cloned();
                    self.add_error(
                        "expected string for import path".into(),
                        Span::new(tok.start, tok.end),
                    );
                    return None;
                }
                let path = self.peek().value.clone();
                self.advance();
                Some(ImportDecl {
                    path,
                    alias: Option::None,
                    names,
                    is_glob: false,
                })
            }

            TokenKind::Star => {
                // Glob: import * from "math"
                self.advance(); // consume *
                if !self.check(TokenKind::From) {
                    let tok = self.peek_cloned();
                    self.add_error(
                        "expected 'from' after '*'".into(),
                        Span::new(tok.start, tok.end),
                    );
                    return None;
                }
                self.advance(); // consume from
                if self.peek().kind != TokenKind::String {
                    let tok = self.peek_cloned();
                    self.add_error(
                        "expected string for import path".into(),
                        Span::new(tok.start, tok.end),
                    );
                    return None;
                }
                let path = self.peek().value.clone();
                self.advance();
                Some(ImportDecl {
                    path,
                    alias: Option::None,
                    names: Vec::new(),
                    is_glob: true,
                })
            }

            TokenKind::Ident => {
                // Aliased: import fmt from "io"
                let alias = self.parse_identifier()?;
                if !self.check(TokenKind::From) {
                    let tok = self.peek_cloned();
                    self.add_error(
                        "expected 'from' after alias name".into(),
                        Span::new(tok.start, tok.end),
                    );
                    return None;
                }
                self.advance(); // consume from
                if self.peek().kind != TokenKind::String {
                    let tok = self.peek_cloned();
                    self.add_error(
                        "expected string for import path".into(),
                        Span::new(tok.start, tok.end),
                    );
                    return None;
                }
                let path = self.peek().value.clone();
                self.advance();
                Some(ImportDecl {
                    path,
                    alias: Some(alias),
                    names: Vec::new(),
                    is_glob: false,
                })
            }

            _ => {
                let tok = self.peek_cloned();
                self.add_error(
                    "expected string, identifier, '{', or '*' for import".into(),
                    Span::new(tok.start, tok.end),
                );
                None
            }
        }
    }

    fn parse_export_decl(&mut self) -> Option<ExportDecl> {
        // export { User, Post }
        if !self.check(TokenKind::LBrace) {
            let tok = self.peek_cloned();
            self.add_error(
                "expected '{' after export".into(),
                Span::new(tok.start, tok.end),
            );
            return None;
        }
        self.advance(); // consume {
        self.skip_newlines();
        let mut names = Vec::new();
        while !self.check(TokenKind::RBrace) && !self.at_end() {
            let name = self.parse_identifier()?;
            names.push(name);
            if self.check(TokenKind::Comma) {
                self.advance();
            }
            self.skip_newlines();
        }
        self.consume(TokenKind::RBrace, "}");
        Some(ExportDecl { names })
    }

    fn parse_provider_decl(&mut self) -> Option<ProviderDecl> {
        let name = self.parse_identifier()?;
        self.consume(TokenKind::LBrace, "{");
        self.skip_newlines();

        let mut fields = Vec::new();
        while !self.check(TokenKind::RBrace) && !self.at_end() {
            let key = match self.parse_identifier() {
                Some(k) => k,
                Option::None => break,
            };
            self.consume(TokenKind::Colon, ":");
            let val = match self.parse_expr() {
                Some(v) => v,
                Option::None => break,
            };
            fields.push(ProviderField { key, value: val });
            if self.check(TokenKind::Comma) {
                self.advance();
            }
            self.skip_newlines();
        }

        self.consume(TokenKind::RBrace, "}");
        Some(ProviderDecl { name, fields })
    }

    fn parse_tool_decl(&mut self) -> Option<ToolDecl> {
        let name = self.parse_identifier()?;
        let params = self.parse_params()?;

        let return_ty = if self.check(TokenKind::Arrow) {
            self.advance();
            Some(self.parse_type()?)
        } else {
            Option::None
        };

        self.consume(TokenKind::LBrace, "{");
        self.skip_newlines();

        // Expect triple-quoted description.
        let description = if self.peek().kind == TokenKind::TripleQuoteString {
            let d = self.peek().value.clone();
            self.advance();
            d
        } else {
            let tok = self.peek_cloned();
            self.add_error(
                "expected triple-quoted description string".into(),
                Span::new(tok.start, tok.end),
            );
            return None;
        };

        self.skip_newlines();

        // Optional body statements.
        let body = if !self.check(TokenKind::RBrace) {
            let block_start = self.peek().start;
            let mut stmts = Vec::new();
            while !self.check(TokenKind::RBrace) && !self.at_end() {
                if let Some(stmt) = self.parse_statement() {
                    stmts.push(stmt);
                } else {
                    self.advance();
                }
                self.skip_newlines();
            }
            Some(Block {
                statements: stmts,
                span: self.span(block_start),
            })
        } else {
            Option::None
        };

        self.consume(TokenKind::RBrace, "}");
        Some(ToolDecl {
            decorators: Vec::new(),
            name,
            params,
            return_ty,
            description,
            body,
        })
    }

    fn parse_agent_decl(&mut self) -> Option<AgentDecl> {
        let name = self.parse_identifier()?;
        self.consume(TokenKind::LBrace, "{");
        self.skip_newlines();

        let mut fields = Vec::new();
        while !self.check(TokenKind::RBrace) && !self.at_end() {
            let key = match self.parse_identifier() {
                Some(k) => k,
                Option::None => break,
            };
            self.consume(TokenKind::Colon, ":");
            let val = match self.parse_expr() {
                Some(v) => v,
                Option::None => break,
            };
            fields.push(AgentField { key, value: val });
            if self.check(TokenKind::Comma) {
                self.advance();
            }
            self.skip_newlines();
        }

        self.consume(TokenKind::RBrace, "}");
        Some(AgentDecl { name, fields })
    }

    fn parse_workflow_decl(&mut self, trigger: Option<Decorator>) -> Option<WorkflowDecl> {
        let name = self.parse_identifier()?;
        let params = self.parse_params()?;

        let return_ty = if self.check(TokenKind::Arrow) {
            self.advance();
            Some(self.parse_type()?)
        } else {
            Option::None
        };

        self.consume(TokenKind::LBrace, "{");
        self.skip_newlines();

        // Parse lifecycle hooks at top of workflow body
        let mut hooks = Vec::new();
        while self.check(TokenKind::Onerror)
            || self.check(TokenKind::Onsuccess)
            || self.check(TokenKind::Oncancel)
        {
            if let Some(hook) = self.parse_lifecycle_hook() {
                hooks.push(hook);
            }
            self.skip_newlines();
        }

        // Parse optional triple-quoted description (for MCP tool exposure)
        let description = if self.peek().kind == TokenKind::TripleQuoteString {
            let d = self.peek().value.clone();
            self.advance();
            self.skip_newlines();
            d
        } else {
            String::new()
        };

        // Parse body statements
        let mut stmts = Vec::new();
        while !self.check(TokenKind::RBrace) && !self.at_end() {
            // Check for @decorator before step
            if self.check(TokenKind::At) {
                let deco_start = self.peek().start;
                let mut decorators = Vec::new();
                while self.check(TokenKind::At) {
                    if let Some(dec) = self.parse_decorator() {
                        decorators.push(dec);
                    }
                    self.skip_newlines();
                }
                // After decorators, expect step
                if self.check(TokenKind::Step) {
                    self.advance();
                    if let Some(ss) = self.parse_step_statement_with_decorators(decorators) {
                        stmts.push(Spanned::new(StmtKind::Step(ss), self.span(deco_start)));
                    }
                } else {
                    // Not a step -- treat decorators as error
                    let tok = self.peek_cloned();
                    self.add_error(
                        "decorator can only precede step inside workflow".into(),
                        Span::new(tok.start, tok.end),
                    );
                    self.advance();
                }
            } else {
                if let Some(stmt) = self.parse_statement() {
                    stmts.push(stmt);
                } else {
                    self.advance();
                }
            }
            self.skip_newlines();
        }
        self.consume(TokenKind::RBrace, "}");

        Some(WorkflowDecl {
            name,
            trigger,
            decorators: Vec::new(),
            params,
            return_ty,
            description,
            body: Block {
                statements: stmts,
                span: Span::default(),
            },
            hooks,
        })
    }

    fn parse_decorator(&mut self) -> Option<Decorator> {
        self.consume(TokenKind::At, "@");
        let name = self.parse_identifier()?;

        let mut args = Vec::new();
        if self.check(TokenKind::LParen) {
            self.advance();
            while !self.check(TokenKind::RParen) && !self.at_end() {
                let arg_start = self.peek().start;
                let e = self.parse_expr()?;
                // Handle named args: key: value -> wrap as MapEntry-style MapExpr
                if self.check(TokenKind::Colon) {
                    self.advance();
                    let val = self.parse_expr()?;
                    // Encode as a single-entry map: {key: value}
                    let entry = MapExpr {
                        entries: vec![MapEntry {
                            key: e,
                            value: val,
                        }],
                    };
                    args.push(Spanned::new(
                        ExprKind::Map(entry),
                        self.span(arg_start),
                    ));
                } else {
                    args.push(e);
                }
                if !self.check(TokenKind::RParen) {
                    self.consume(TokenKind::Comma, ",");
                }
            }
            self.consume(TokenKind::RParen, ")");
        }

        self.skip_newlines();
        Some(Decorator { name, args })
    }

    fn parse_fn_decl(&mut self, is_public: bool) -> Option<FunctionDef> {
        let name = self.parse_identifier()?;
        let params = self.parse_params()?;

        let return_ty = if self.check(TokenKind::Arrow) {
            self.advance();
            Some(self.parse_type()?)
        } else {
            Option::None
        };

        let body = self.parse_block()?;

        Some(FunctionDef {
            is_public,
            name,
            params,
            return_ty,
            body,
        })
    }

    fn parse_test_decl(&mut self) -> Option<TestDecl> {
        // Expect a string literal name: test "name" { ... }
        if self.peek().kind != TokenKind::String {
            let tok = self.peek_cloned();
            self.add_error(
                "expected test name string".into(),
                Span::new(tok.start, tok.end),
            );
            return None;
        }
        let name_tok = self.peek_cloned();
        let name = Spanned::new(
            name_tok.value.clone(),
            Span::new(name_tok.start, name_tok.end),
        );
        self.advance();

        let body = self.parse_block()?;
        Some(TestDecl { name, body })
    }

    fn parse_assert_stmt(&mut self) -> Option<AssertStmt> {
        let cond = self.parse_expr()?;

        let message = if self.check(TokenKind::Comma) {
            self.advance();
            let m = self.parse_expr()?;
            Some(Box::new(m))
        } else {
            Option::None
        };

        Some(AssertStmt {
            condition: Box::new(cond),
            message,
        })
    }
}

// ---------------------------------------------------------------------------
// Identifier helper
// ---------------------------------------------------------------------------

impl Parser {
    fn parse_identifier(&mut self) -> Option<Spanned<String>> {
        let tok = self.peek_cloned();
        match tok.kind {
            TokenKind::Ident => {
                let name = tok.value.clone();
                let span = self.current_span();
                self.advance();
                Some(Spanned::new(name, span))
            }
            // Contextual keywords usable as identifiers.
            TokenKind::Err => {
                let span = self.current_span();
                self.advance();
                Some(Spanned::new("err".into(), span))
            }
            TokenKind::Ok => {
                let span = self.current_span();
                self.advance();
                Some(Spanned::new("ok".into(), span))
            }
            TokenKind::Tool => {
                let span = self.current_span();
                self.advance();
                Some(Spanned::new("tool".into(), span))
            }
            TokenKind::Agent => {
                let span = self.current_span();
                self.advance();
                Some(Spanned::new("agent".into(), span))
            }
            TokenKind::Provider => {
                let span = self.current_span();
                self.advance();
                Some(Spanned::new("provider".into(), span))
            }
            _ => {
                self.add_error(
                    format!("expected identifier, found {}", tok.kind),
                    Span::new(tok.start, tok.end),
                );
                None
            }
        }
    }
}

// ---------------------------------------------------------------------------
// Interpolated string parts
// ---------------------------------------------------------------------------

/// Parse the interior of an interpolated string, splitting it into literal
/// fragments and `${expr}` parts.
fn parse_interpolated_string_parts(raw: &str) -> Vec<StringPart> {
    let mut parts = Vec::new();
    let mut literal = String::new();
    let bytes = raw.as_bytes();
    let len = bytes.len();
    let mut i = 0;

    while i < len {
        let ch = bytes[i];

        if ch == b'\\' && i + 1 < len {
            let next = bytes[i + 1];
            match next {
                b'n' => literal.push('\n'),
                b't' => literal.push('\t'),
                b'r' => literal.push('\r'),
                b'\\' => literal.push('\\'),
                b'"' => literal.push('"'),
                b'{' => literal.push('{'),
                b'}' => literal.push('}'),
                _ => {
                    literal.push('\\');
                    literal.push(next as char);
                }
            }
            i += 2;
            continue;
        }

        if ch == b'$' && i + 1 < len && bytes[i + 1] == b'{' {
            // Flush literal.
            if !literal.is_empty() {
                parts.push(StringPart::Literal(literal.clone()));
                literal.clear();
            }

            // Extract expression text between ${ and }.
            i += 2; // skip ${
            let mut depth: usize = 1;
            let expr_start = i;
            while i < len && depth > 0 {
                if bytes[i] == b'{' {
                    depth += 1;
                } else if bytes[i] == b'}' {
                    depth -= 1;
                    if depth == 0 {
                        break;
                    }
                }
                i += 1;
            }
            let expr_text = &raw[expr_start..i];
            if i < len {
                i += 1; // skip closing }
            }

            // Parse the expression via a sub-parser.
            if !expr_text.is_empty() {
                let (sf, _) = parse(expr_text);
                let mut found = false;
                if let Some(first_item) = sf.items.first() {
                    if let ItemKind::ItemStatement(ref stmt) = first_item.node {
                        if let StmtKind::Expr(ref es) = stmt.node {
                            parts.push(StringPart::Expr(Box::new((*es.value).clone())));
                            found = true;
                        }
                    }
                }
                if !found {
                    // Fallback: treat as literal.
                    parts.push(StringPart::Literal(format!("{{{}}}", expr_text)));
                }
            }
            continue;
        }

        literal.push(ch as char);
        i += 1;
    }

    if !literal.is_empty() {
        parts.push(StringPart::Literal(literal));
    }

    parts
}

// ---------------------------------------------------------------------------
// Utility functions
// ---------------------------------------------------------------------------

/// Returns the first character of a string, or 'a' if empty.
fn first_rune(s: &str) -> char {
    s.chars().next().unwrap_or('a')
}

/// Parse an integer literal, handling hex, binary, octal prefixes and underscores.
fn parse_int(raw: &str) -> i64 {
    let clean = raw.replace('_', "");

    if clean.len() > 2 {
        match &clean[..2] {
            "0x" | "0X" => return i64::from_str_radix(&clean[2..], 16).unwrap_or(0),
            "0b" | "0B" => return i64::from_str_radix(&clean[2..], 2).unwrap_or(0),
            "0o" | "0O" => return i64::from_str_radix(&clean[2..], 8).unwrap_or(0),
            _ => {}
        }
    }

    clean.parse::<i64>().unwrap_or(0)
}

/// Parse a float literal, stripping underscores.
fn parse_float(raw: &str) -> f64 {
    let clean = raw.replace('_', "");
    clean.parse::<f64>().unwrap_or(0.0)
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;

    fn parse_ok(source: &str) -> SourceFile {
        let (sf, errors) = parse(source);
        assert!(
            errors.is_empty(),
            "unexpected errors: {:?}\nsource: {}",
            errors,
            source
        );
        sf
    }

    // 1. Empty source
    #[test]
    fn empty_source() {
        let sf = parse_ok("");
        assert!(sf.items.is_empty());
    }

    // 2. Let/const bindings
    #[test]
    fn let_binding() {
        let sf = parse_ok("let x = 42");
        assert_eq!(sf.items.len(), 1);
        if let ItemKind::ItemStatement(ref stmt) = sf.items[0].node {
            if let StmtKind::Let(ref ls) = stmt.node {
                assert_eq!(ls.name.node, "x");
                assert!(!ls.is_const);
                return;
            }
        }
        panic!("expected let binding");
    }

    #[test]
    fn const_binding() {
        let sf = parse_ok("const PI = 3.14");
        assert_eq!(sf.items.len(), 1);
        if let ItemKind::ItemStatement(ref stmt) = sf.items[0].node {
            if let StmtKind::Let(ref ls) = stmt.node {
                assert_eq!(ls.name.node, "PI");
                assert!(ls.is_const);
                return;
            }
        }
        panic!("expected const binding");
    }

    // 3. Function definitions
    #[test]
    fn function_def() {
        let sf = parse_ok("fn greet(name: string) -> string { return name }");
        assert_eq!(sf.items.len(), 1);
        if let ItemKind::FunctionDef(ref fd) = sf.items[0].node {
            assert_eq!(fd.name.node, "greet");
            assert_eq!(fd.params.len(), 1);
            assert!(fd.return_ty.is_some());
            return;
        }
        panic!("expected function def");
    }

    #[test]
    fn pub_function() {
        let sf = parse_ok("pub fn add(a: int, b: int) -> int { return a + b }");
        if let ItemKind::FunctionDef(ref fd) = sf.items[0].node {
            assert!(fd.is_public);
            assert_eq!(fd.name.node, "add");
            return;
        }
        panic!("expected pub function def");
    }

    // 4. If/else statements
    #[test]
    fn if_else() {
        let sf = parse_ok("if x > 0 { return x } else { return 0 }");
        assert_eq!(sf.items.len(), 1);
        if let ItemKind::ItemStatement(ref stmt) = sf.items[0].node {
            if let StmtKind::If(ref is) = stmt.node {
                assert!(matches!(is.else_branch, ElseBranch::Block(_)));
                return;
            }
        }
        panic!("expected if/else statement");
    }

    #[test]
    fn if_else_if() {
        let sf = parse_ok(
            "if x > 0 { return 1 } else if x == 0 { return 0 } else { return -1 }",
        );
        assert_eq!(sf.items.len(), 1);
        if let ItemKind::ItemStatement(ref stmt) = sf.items[0].node {
            if let StmtKind::If(ref is) = stmt.node {
                assert!(matches!(is.else_branch, ElseBranch::ElseIf(_)));
                return;
            }
        }
        panic!("expected if/else if statement");
    }

    // 5. For loops
    #[test]
    fn for_loop() {
        let sf = parse_ok("for x in items { print(x) }");
        assert_eq!(sf.items.len(), 1);
        if let ItemKind::ItemStatement(ref stmt) = sf.items[0].node {
            if let StmtKind::For(ref fs) = stmt.node {
                assert!(matches!(fs.pattern, ForPattern::Single(_)));
                return;
            }
        }
        panic!("expected for loop");
    }

    #[test]
    fn for_loop_pair() {
        let sf = parse_ok("for k, v in map { print(k) }");
        if let ItemKind::ItemStatement(ref stmt) = sf.items[0].node {
            if let StmtKind::For(ref fs) = stmt.node {
                assert!(matches!(fs.pattern, ForPattern::Pair(_, _)));
                return;
            }
        }
        panic!("expected pair for loop");
    }

    // 6. Match expressions
    #[test]
    fn match_expr() {
        let sf = parse_ok("match x { 1 => true\n 2 => false\n _ => false }");
        assert_eq!(sf.items.len(), 1);
        if let ItemKind::ItemStatement(ref stmt) = sf.items[0].node {
            if let StmtKind::Match(ref ms) = stmt.node {
                assert_eq!(ms.arms.len(), 3);
                return;
            }
        }
        panic!("expected match statement");
    }

    // 7. Binary expressions with correct precedence
    #[test]
    fn binary_precedence() {
        let sf = parse_ok("let x = 1 + 2 * 3");
        if let ItemKind::ItemStatement(ref stmt) = sf.items[0].node {
            if let StmtKind::Let(ref ls) = stmt.node {
                // Should be Add(1, Mul(2, 3))
                if let ExprKind::Binary(ref bin) = ls.value.node {
                    assert!(matches!(bin.op.node, BinaryOp::Add));
                    if let ExprKind::Binary(ref rhs) = bin.right.node {
                        assert!(matches!(rhs.op.node, BinaryOp::Mul));
                        return;
                    }
                }
            }
        }
        panic!("expected correct precedence: 1 + (2 * 3)");
    }

    // 8. Struct definitions
    #[test]
    fn struct_def() {
        let sf = parse_ok("struct User { name: string\n age: int }");
        assert_eq!(sf.items.len(), 1);
        if let ItemKind::TypeDef(ref td) = sf.items[0].node {
            assert_eq!(td.name.node, "User");
            assert_eq!(td.fields.len(), 2);
            return;
        }
        panic!("expected struct def");
    }

    // 9. Import declarations (all 4 forms)
    #[test]
    fn import_basic() {
        let sf = parse_ok(r#"import "io""#);
        if let ItemKind::ImportDecl(ref id) = sf.items[0].node {
            assert_eq!(id.path, "io");
            assert!(id.alias.is_none());
            assert!(id.names.is_empty());
            assert!(!id.is_glob);
            return;
        }
        panic!("expected basic import");
    }

    #[test]
    fn import_aliased() {
        let sf = parse_ok(r#"import fmt from "io""#);
        if let ItemKind::ImportDecl(ref id) = sf.items[0].node {
            assert_eq!(id.path, "io");
            assert_eq!(id.alias.as_ref().unwrap().node, "fmt");
            return;
        }
        panic!("expected aliased import");
    }

    #[test]
    fn import_selective() {
        let sf = parse_ok(r#"import { User, Post } from "models""#);
        if let ItemKind::ImportDecl(ref id) = sf.items[0].node {
            assert_eq!(id.path, "models");
            assert_eq!(id.names.len(), 2);
            assert_eq!(id.names[0].node, "User");
            assert_eq!(id.names[1].node, "Post");
            return;
        }
        panic!("expected selective import");
    }

    #[test]
    fn import_glob() {
        let sf = parse_ok(r#"import * from "math""#);
        if let ItemKind::ImportDecl(ref id) = sf.items[0].node {
            assert_eq!(id.path, "math");
            assert!(id.is_glob);
            return;
        }
        panic!("expected glob import");
    }

    // 10. Agent declarations
    #[test]
    fn agent_decl() {
        let sf = parse_ok(
            r#"agent Assistant {
                model: "gpt-4"
                prompt: "You are helpful"
            }"#,
        );
        assert_eq!(sf.items.len(), 1);
        if let ItemKind::AgentDecl(ref ad) = sf.items[0].node {
            assert_eq!(ad.name.node, "Assistant");
            assert_eq!(ad.fields.len(), 2);
            return;
        }
        panic!("expected agent decl");
    }

    // 11. Workflow declarations
    #[test]
    fn workflow_decl() {
        let sf = parse_ok(
            r#"workflow process(input: string) -> string {
                return input
            }"#,
        );
        assert_eq!(sf.items.len(), 1);
        if let ItemKind::WorkflowDecl(ref wf) = sf.items[0].node {
            assert_eq!(wf.name.node, "process");
            assert_eq!(wf.params.len(), 1);
            assert!(wf.return_ty.is_some());
            return;
        }
        panic!("expected workflow decl");
    }

    // 12. Error recovery (parser continues after errors)
    #[test]
    fn error_recovery() {
        let (sf, errors) = parse("let x = 42\n@@@ bad syntax\nlet y = 10");
        // Parser should recover and parse at least some items
        assert!(!errors.is_empty(), "expected errors from bad syntax");
        // Should have parsed at least the first let binding
        assert!(
            !sf.items.is_empty(),
            "parser should recover and produce items"
        );
    }

    // -- Additional coverage tests --

    #[test]
    fn enum_def() {
        let sf = parse_ok("enum Color { Red, Green, Blue }");
        if let ItemKind::EnumDef(ref ed) = sf.items[0].node {
            assert_eq!(ed.name.node, "Color");
            assert_eq!(ed.variants.len(), 3);
            return;
        }
        panic!("expected enum def");
    }

    #[test]
    fn type_alias() {
        let sf = parse_ok("type ID = int");
        if let ItemKind::TypeAlias(ref ta) = sf.items[0].node {
            assert_eq!(ta.name.node, "ID");
            return;
        }
        panic!("expected type alias");
    }

    #[test]
    fn while_loop() {
        let sf = parse_ok("while x > 0 { x = x - 1 }");
        if let ItemKind::ItemStatement(ref stmt) = sf.items[0].node {
            assert!(matches!(stmt.node, StmtKind::While(_)));
            return;
        }
        panic!("expected while loop");
    }

    #[test]
    fn try_catch() {
        let sf = parse_ok("try { risky() } catch e { print(e) }");
        if let ItemKind::ItemStatement(ref stmt) = sf.items[0].node {
            assert!(matches!(stmt.node, StmtKind::Try(_)));
            return;
        }
        panic!("expected try/catch");
    }

    #[test]
    fn provider_decl() {
        let sf = parse_ok(r#"provider OpenAI { model: "gpt-4" }"#);
        if let ItemKind::ProviderDecl(ref pd) = sf.items[0].node {
            assert_eq!(pd.name.node, "OpenAI");
            return;
        }
        panic!("expected provider decl");
    }

    #[test]
    fn method_def() {
        let sf = parse_ok("User.greet() -> string { return self.name }");
        if let ItemKind::MethodDef(ref md) = sf.items[0].node {
            assert_eq!(md.type_name.node, "User");
            assert_eq!(md.name.node, "greet");
            return;
        }
        panic!("expected method def");
    }

    #[test]
    fn compound_assignment() {
        let sf = parse_ok("x += 1");
        if let ItemKind::ItemStatement(ref stmt) = sf.items[0].node {
            if let StmtKind::Assign(ref a) = stmt.node {
                if let ExprKind::Binary(ref b) = a.value.node {
                    assert!(matches!(b.op.node, BinaryOp::Add));
                    return;
                }
            }
        }
        panic!("expected compound assignment");
    }

    #[test]
    fn spawn_expr() {
        let sf = parse_ok("spawn { task1() }");
        if let ItemKind::ItemStatement(ref stmt) = sf.items[0].node {
            if let StmtKind::Expr(ref es) = stmt.node {
                assert!(matches!(es.value.node, ExprKind::Spawn(_)));
                return;
            }
        }
        panic!("expected spawn expression");
    }

    #[test]
    fn pipe_operator() {
        let sf = parse_ok("let x = data |> transform |> output");
        if let ItemKind::ItemStatement(ref stmt) = sf.items[0].node {
            if let StmtKind::Let(ref ls) = stmt.node {
                if let ExprKind::Pipe(_) = ls.value.node {
                    return;
                }
            }
        }
        panic!("expected pipe expression");
    }

    #[test]
    fn lambda_expr() {
        let sf = parse_ok("let f = (x) => x + 1");
        if let ItemKind::ItemStatement(ref stmt) = sf.items[0].node {
            if let StmtKind::Let(ref ls) = stmt.node {
                assert!(matches!(ls.value.node, ExprKind::Lambda(_)));
                return;
            }
        }
        panic!("expected lambda expression");
    }

    #[test]
    fn export_decl() {
        let sf = parse_ok("export { Foo, Bar }");
        if let ItemKind::ExportDecl(ref ed) = sf.items[0].node {
            assert_eq!(ed.names.len(), 2);
            return;
        }
        panic!("expected export decl");
    }

    #[test]
    fn test_decl() {
        let sf = parse_ok(r#"test "basic math" { assert 1 + 1 == 2 }"#);
        if let ItemKind::TestDecl(ref td) = sf.items[0].node {
            assert_eq!(td.name.node, "basic math");
            return;
        }
        panic!("expected test decl");
    }

    #[test]
    fn unary_ops() {
        let sf = parse_ok("let x = -42");
        if let ItemKind::ItemStatement(ref stmt) = sf.items[0].node {
            if let StmtKind::Let(ref ls) = stmt.node {
                if let ExprKind::Unary(ref u) = ls.value.node {
                    assert!(matches!(u.op.node, UnaryOp::Neg));
                    return;
                }
            }
        }
        panic!("expected unary negation");
    }

    #[test]
    fn list_literal() {
        let sf = parse_ok("let x = [1, 2, 3]");
        if let ItemKind::ItemStatement(ref stmt) = sf.items[0].node {
            if let StmtKind::Let(ref ls) = stmt.node {
                assert!(matches!(ls.value.node, ExprKind::List(_)));
                return;
            }
        }
        panic!("expected list literal");
    }

    #[test]
    fn break_continue() {
        let sf = parse_ok("for x in items { break }");
        if let ItemKind::ItemStatement(ref stmt) = sf.items[0].node {
            if let StmtKind::For(ref fs) = stmt.node {
                assert_eq!(fs.body.statements.len(), 1);
                assert!(matches!(fs.body.statements[0].node, StmtKind::Break));
                return;
            }
        }
        panic!("expected break in for loop");
    }

    #[test]
    fn parse_int_literals() {
        assert_eq!(parse_int("42"), 42);
        assert_eq!(parse_int("1_000"), 1000);
        assert_eq!(parse_int("0xFF"), 255);
        assert_eq!(parse_int("0b1010"), 10);
        assert_eq!(parse_int("0o17"), 15);
    }

    #[test]
    fn parse_float_literals() {
        assert!((parse_float("3.14") - 3.14).abs() < f64::EPSILON);
        assert!((parse_float("1_000.5") - 1000.5).abs() < f64::EPSILON);
    }
}
