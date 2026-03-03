//! Abstract Syntax Tree types for the Haira programming language.
//!
//! This crate defines the core data structures produced by the parser and consumed
//! by every subsequent compiler pass (resolver, checker, codegen). It ports the Go
//! AST from `compiler/internal/ast/ast.go`, replacing Go interface-based sum types
//! with Rust enums.

// Re-export Span and Spanned from haira-errors so downstream crates can use them
// without depending on haira-errors directly.
pub use haira_errors::{Span, Spanned};

// ---------------------------------------------------------------------------
// Source file
// ---------------------------------------------------------------------------

/// A complete Haira source file.
#[derive(Debug, Clone)]
pub struct SourceFile {
    pub items: Vec<Item>,
    pub span: Span,
}

/// A top-level declaration or statement with span information.
pub type Item = Spanned<ItemKind>;

// ---------------------------------------------------------------------------
// ItemKind -- top-level declarations
// ---------------------------------------------------------------------------

/// Sum type for top-level items in a source file.
#[derive(Debug, Clone)]
pub enum ItemKind {
    /// `struct Name { fields }`
    TypeDef(TypeDef),
    /// `enum Name { variants }`
    EnumDef(EnumDef),
    /// `type Name = Type`
    TypeAlias(TypeAlias),
    /// `import ...`
    ImportDecl(ImportDecl),
    /// `export { names }`
    ExportDecl(ExportDecl),
    /// `provider Name { fields }`
    ProviderDecl(ProviderDecl),
    /// `tool Name(params) -> Type { body }`
    ToolDecl(ToolDecl),
    /// `agent Name { fields }`
    AgentDecl(AgentDecl),
    /// `workflow Name(params) -> Type { body }`
    WorkflowDecl(WorkflowDecl),
    /// `fn name(params) -> Type { body }` or `pub fn ...`
    FunctionDef(FunctionDef),
    /// `Type.method(params) -> Type { body }`
    MethodDef(MethodDef),
    /// A bare statement at the top level.
    ItemStatement(Statement),
    /// `test "name" { body }`
    TestDecl(TestDecl),
}

// ---------------------------------------------------------------------------
// Item kind structs
// ---------------------------------------------------------------------------

/// Struct (type) definition: `struct Name { fields }`.
#[derive(Debug, Clone)]
pub struct TypeDef {
    pub is_public: bool,
    pub name: Spanned<String>,
    pub fields: Vec<Field>,
}

/// A field in a struct definition.
#[derive(Debug, Clone)]
pub struct Field {
    pub name: Spanned<String>,
    pub ty: Option<Spanned<Type>>,
    pub default: Option<Box<Expr>>,
    pub span: Span,
}

/// Enum definition: `enum Name { variants }`.
#[derive(Debug, Clone)]
pub struct EnumDef {
    pub is_public: bool,
    pub name: Spanned<String>,
    pub variants: Vec<EnumVariant>,
}

/// A single variant of an enum, optionally with associated data.
#[derive(Debug, Clone)]
pub struct EnumVariant {
    pub name: Spanned<String>,
    pub fields: Vec<Param>,
    pub span: Span,
}

/// Type alias: `type Name = Type`.
#[derive(Debug, Clone)]
pub struct TypeAlias {
    pub name: Spanned<String>,
    pub ty: Spanned<Type>,
}

/// Import declaration with four forms:
/// - Basic: `import "io"`
/// - Aliased: `import fmt from "io"`
/// - Selective: `import { User, Post } from "models"`
/// - Glob: `import * from "math"`
#[derive(Debug, Clone)]
pub struct ImportDecl {
    /// The module path string, e.g. `"io"`, `"models/user"`.
    pub path: String,
    /// Alias for the import, e.g. `fmt` in `import fmt from "io"`.
    pub alias: Option<Spanned<String>>,
    /// Selective names, e.g. `{ User, Post }` in `import { User, Post } from "models"`.
    pub names: Vec<Spanned<String>>,
    /// Whether this is a glob import: `import * from "math"`.
    pub is_glob: bool,
}

/// Export declaration: `export { Name1, Name2 }`.
#[derive(Debug, Clone)]
pub struct ExportDecl {
    pub names: Vec<Spanned<String>>,
}

/// Provider declaration: `provider Name { fields }`.
#[derive(Debug, Clone)]
pub struct ProviderDecl {
    pub name: Spanned<String>,
    pub fields: Vec<ProviderField>,
}

/// A key-value field in a provider declaration.
#[derive(Debug, Clone)]
pub struct ProviderField {
    pub key: Spanned<String>,
    pub value: Expr,
}

/// Tool declaration: `tool Name(params) -> Type """description""" { body }`.
#[derive(Debug, Clone)]
pub struct ToolDecl {
    pub decorators: Vec<Decorator>,
    pub name: Spanned<String>,
    pub params: Vec<Param>,
    pub return_ty: Option<Spanned<Type>>,
    pub description: String,
    pub body: Option<Block>,
}

/// Agent declaration: `agent Name { fields }`.
#[derive(Debug, Clone)]
pub struct AgentDecl {
    pub name: Spanned<String>,
    pub fields: Vec<AgentField>,
}

/// A key-value field in an agent declaration.
#[derive(Debug, Clone)]
pub struct AgentField {
    pub key: Spanned<String>,
    pub value: Expr,
}

/// Workflow declaration with optional trigger, decorators, and lifecycle hooks.
#[derive(Debug, Clone)]
pub struct WorkflowDecl {
    pub name: Spanned<String>,
    pub trigger: Option<Decorator>,
    pub decorators: Vec<Decorator>,
    pub params: Vec<Param>,
    pub return_ty: Option<Spanned<Type>>,
    /// Optional triple-quoted description (for MCP).
    pub description: String,
    pub body: Block,
    pub hooks: Vec<LifecycleHook>,
}

/// A decorator: `@name(args)`.
#[derive(Debug, Clone)]
pub struct Decorator {
    pub name: Spanned<String>,
    pub args: Vec<Expr>,
}

/// Function definition: `fn name(params) -> Type { body }`.
#[derive(Debug, Clone)]
pub struct FunctionDef {
    pub is_public: bool,
    pub name: Spanned<String>,
    pub params: Vec<Param>,
    pub return_ty: Option<Spanned<Type>>,
    pub body: Block,
}

/// Method definition: `Type.method(params) -> Type { body }`.
#[derive(Debug, Clone)]
pub struct MethodDef {
    pub type_name: Spanned<String>,
    pub name: Spanned<String>,
    pub params: Vec<Param>,
    pub return_ty: Option<Spanned<Type>>,
    pub body: Block,
}

/// Test declaration: `test "name" { body }`.
#[derive(Debug, Clone)]
pub struct TestDecl {
    pub name: Spanned<String>,
    pub body: Block,
}

// ---------------------------------------------------------------------------
// Parameters
// ---------------------------------------------------------------------------

/// A function/tool/workflow parameter.
#[derive(Debug, Clone)]
pub struct Param {
    pub name: Spanned<String>,
    pub ty: Option<Spanned<Type>>,
    pub default: Option<Box<Expr>>,
    pub is_rest: bool,
    pub span: Span,
}

// ---------------------------------------------------------------------------
// Type expressions
// ---------------------------------------------------------------------------

/// Type expression sum type.
#[derive(Debug, Clone)]
pub enum Type {
    /// A simple named type, e.g. `int`, `string`, `User`.
    Named(String),
    /// A list type: `[Type]`.
    List(Box<Spanned<Type>>),
    /// A map type: `map[Key]Value`.
    Map {
        key: Box<Spanned<Type>>,
        value: Box<Spanned<Type>>,
    },
    /// An option type: `Type?`.
    Option(Box<Spanned<Type>>),
    /// A function type: `fn(Params) -> Ret`.
    Function {
        params: Vec<Spanned<Type>>,
        ret: Box<Spanned<Type>>,
    },
    /// A union type: `Type1 | Type2`.
    Union(Vec<Spanned<Type>>),
    /// A generic type: `Name<Args>`.
    Generic {
        name: String,
        args: Vec<Spanned<Type>>,
    },
}

// ---------------------------------------------------------------------------
// Statements
// ---------------------------------------------------------------------------

/// A statement with span information.
pub type Statement = Spanned<StmtKind>;

/// Sum type for statements.
#[derive(Debug, Clone)]
pub enum StmtKind {
    /// Assignment: `targets = value`.
    Assign(AssignStmt),
    /// Let binding: `let name = value` or `const name = value`.
    Let(LetStmt),
    /// If statement.
    If(IfStmt),
    /// For loop.
    For(ForStmt),
    /// While loop.
    While(WhileStmt),
    /// Return statement.
    Return(ReturnStmt),
    /// Try/catch.
    Try(TryStmt),
    /// Defer.
    Defer(DeferStmt),
    /// Error defer.
    ErrDefer(ErrDeferStmt),
    /// Match statement.
    Match(MatchStmt),
    /// Break.
    Break,
    /// Continue.
    Continue,
    /// Expression statement.
    Expr(ExprStmt),
    /// Workflow step.
    Step(StepStmt),
    /// Assert statement.
    Assert(AssertStmt),
}

/// Assignment statement: `targets = value`.
#[derive(Debug, Clone)]
pub struct AssignStmt {
    pub targets: Vec<AssignTarget>,
    pub value: Box<Expr>,
}

/// An assignment target with optional type annotation.
#[derive(Debug, Clone)]
pub struct AssignTarget {
    pub path: AssignPath,
    pub ty: Option<Spanned<Type>>,
}

/// The left-hand side of an assignment.
#[derive(Debug, Clone)]
pub enum AssignPath {
    /// A simple identifier: `x`.
    Ident(Spanned<String>),
    /// Field access: `object.field`.
    Field {
        object: Box<AssignPath>,
        field: Spanned<String>,
    },
    /// Index access: `object[index]`.
    Index {
        object: Box<AssignPath>,
        index: Box<Expr>,
    },
}

/// Let/const binding: `let name: Type = value`.
#[derive(Debug, Clone)]
pub struct LetStmt {
    pub name: Spanned<String>,
    /// Optional type annotation.
    pub type_ann: Option<Spanned<Type>>,
    pub value: Box<Expr>,
    /// True for `const` declarations.
    pub is_const: bool,
}

/// If statement.
#[derive(Debug, Clone)]
pub struct IfStmt {
    pub condition: Box<Expr>,
    pub then_branch: Block,
    pub else_branch: ElseBranch,
}

/// The else part of an if statement.
#[derive(Debug, Clone)]
pub enum ElseBranch {
    /// No else.
    None,
    /// `else { block }`.
    Block(Block),
    /// `else if ...`.
    ElseIf(Box<Spanned<IfStmt>>),
}

/// For loop: `for pattern in iterator { body }`.
#[derive(Debug, Clone)]
pub struct ForStmt {
    pub pattern: ForPattern,
    pub iterator: Box<Expr>,
    pub body: Block,
}

/// Destructuring pattern in a for loop.
#[derive(Debug, Clone)]
pub enum ForPattern {
    /// Single binding: `for x in ...`.
    Single(Spanned<String>),
    /// Pair binding: `for k, v in ...`.
    Pair(Spanned<String>, Spanned<String>),
}

/// While loop: `while condition { body }`.
#[derive(Debug, Clone)]
pub struct WhileStmt {
    pub condition: Box<Expr>,
    pub body: Block,
}

/// Return statement: `return values`.
#[derive(Debug, Clone)]
pub struct ReturnStmt {
    pub values: Vec<Expr>,
}

/// Try/catch statement.
#[derive(Debug, Clone)]
pub struct TryStmt {
    pub body: Block,
    pub error_name: Spanned<String>,
    pub catch_body: Block,
}

/// Defer statement: `defer expr`.
#[derive(Debug, Clone)]
pub struct DeferStmt {
    pub value: Box<Expr>,
}

/// Error defer statement: `errdefer expr`.
#[derive(Debug, Clone)]
pub struct ErrDeferStmt {
    pub value: Box<Expr>,
}

/// Match statement (wraps the match expression for statement context).
#[derive(Debug, Clone)]
pub struct MatchStmt {
    pub subject: Box<Expr>,
    pub arms: Vec<MatchArm>,
}

/// Expression statement.
#[derive(Debug, Clone)]
pub struct ExprStmt {
    pub value: Box<Expr>,
}

/// Workflow step: `step "name" { body }`.
#[derive(Debug, Clone)]
pub struct StepStmt {
    pub name: Spanned<String>,
    pub body: Vec<Statement>,
    pub decorators: Vec<Decorator>,
    pub hooks: Vec<LifecycleHook>,
}

/// Lifecycle hook for workflows and steps.
#[derive(Debug, Clone)]
pub struct LifecycleHook {
    pub kind: LifecycleHookKind,
    /// Only for `onerror`: the error variable name.
    pub err_name: String,
    /// Only for `onsuccess` at workflow level: the result variable name.
    pub arg_name: String,
    pub body: Block,
}

/// Kind of lifecycle hook.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum LifecycleHookKind {
    Onerror,
    Onsuccess,
    Oncancel,
}

/// Assert statement: `assert condition, "message"`.
#[derive(Debug, Clone)]
pub struct AssertStmt {
    pub condition: Box<Expr>,
    /// Optional custom message.
    pub message: Option<Box<Expr>>,
}

// ---------------------------------------------------------------------------
// Expressions
// ---------------------------------------------------------------------------

/// An expression with span information.
pub type Expr = Spanned<ExprKind>;

/// Sum type for expressions.
#[derive(Debug, Clone)]
pub enum ExprKind {
    /// A literal value.
    Literal(Literal),
    /// An identifier reference.
    Ident(String),
    /// Binary operation: `left op right`.
    Binary(BinaryExpr),
    /// Unary operation: `op operand`.
    Unary(UnaryExpr),
    /// Function call: `callee(args)`.
    Call(CallExpr),
    /// Method call: `receiver.method(args)`.
    MethodCall(MethodCallExpr),
    /// Field access: `object.field`.
    Field(FieldExpr),
    /// Index access: `object[index]`.
    Index(IndexExpr),
    /// Pipe: `left |> right`.
    Pipe(PipeExpr),
    /// Lambda: `(params) => body`.
    Lambda(LambdaExpr),
    /// Match expression.
    Match(MatchExpr),
    /// If expression (if used in expression context).
    If(IfExpr),
    /// Block expression.
    Block(BlockExpr),
    /// List literal: `[elems]`.
    List(ListExpr),
    /// Map literal: `{ entries }`.
    Map(MapExpr),
    /// Struct instantiation: `TypeName { fields }`.
    Instance(InstanceExpr),
    /// Range: `start..end` or `start..=end`.
    Range(RangeExpr),
    /// Error propagation: `expr?`.
    Propagate(PropagateExpr),
    /// Or-else: `expr ?? default`.
    Orelse(OrelseExpr),
    /// Some wrapping: `some(expr)`.
    Some(SomeExpr),
    /// None literal.
    None,
    /// Async block: `async { body }`.
    Async(AsyncExpr),
    /// Spawn block: `spawn { body }`.
    Spawn(SpawnExpr),
    /// Select expression.
    Select(SelectExpr),
    /// Parenthesized expression.
    Paren(ParenExpr),
}

// ---------------------------------------------------------------------------
// Literal
// ---------------------------------------------------------------------------

/// Literal value.
#[derive(Debug, Clone)]
pub enum Literal {
    Int(i64),
    Float(f64),
    String(String),
    Bool(bool),
    InterpolatedString(Vec<StringPart>),
}

/// Part of an interpolated string.
#[derive(Debug, Clone)]
pub enum StringPart {
    /// A literal text fragment.
    Literal(String),
    /// An interpolated expression: `${expr}`.
    Expr(Box<Expr>),
}

// ---------------------------------------------------------------------------
// Expression structs
// ---------------------------------------------------------------------------

/// Binary expression: `left op right`.
#[derive(Debug, Clone)]
pub struct BinaryExpr {
    pub left: Box<Expr>,
    pub op: Spanned<BinaryOp>,
    pub right: Box<Expr>,
}

/// Binary operator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum BinaryOp {
    Add,
    Sub,
    Mul,
    Div,
    Mod,
    Eq,
    Ne,
    Lt,
    Gt,
    Le,
    Ge,
    And,
    Or,
    BitAnd,
    BitOr,
    BitXor,
    Shl,
    Shr,
}

/// Unary expression: `op operand`.
#[derive(Debug, Clone)]
pub struct UnaryExpr {
    pub op: Spanned<UnaryOp>,
    pub operand: Box<Expr>,
}

/// Unary operator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum UnaryOp {
    Neg,
    Not,
    BitNot,
}

/// Function call: `callee(args)`.
#[derive(Debug, Clone)]
pub struct CallExpr {
    pub callee: Box<Expr>,
    pub args: Vec<Argument>,
}

/// A call argument, optionally named.
#[derive(Debug, Clone)]
pub struct Argument {
    /// Named argument label, if present.
    pub name: Option<Spanned<String>>,
    pub value: Expr,
    pub span: Span,
}

/// Method call: `receiver.method(args)`.
#[derive(Debug, Clone)]
pub struct MethodCallExpr {
    pub receiver: Box<Expr>,
    pub method: Spanned<String>,
    pub args: Vec<Argument>,
}

/// Field access: `object.field`.
#[derive(Debug, Clone)]
pub struct FieldExpr {
    pub object: Box<Expr>,
    pub field: Spanned<String>,
}

/// Index access: `object[index]`.
#[derive(Debug, Clone)]
pub struct IndexExpr {
    pub object: Box<Expr>,
    pub index: Box<Expr>,
}

/// Pipe expression: `left |> right`.
#[derive(Debug, Clone)]
pub struct PipeExpr {
    pub left: Box<Expr>,
    pub right: Box<Expr>,
}

/// Lambda expression: `(params) => body`.
#[derive(Debug, Clone)]
pub struct LambdaExpr {
    pub params: Vec<Param>,
    pub body: LambdaBody,
}

/// Body of a lambda -- either a single expression or a block.
#[derive(Debug, Clone)]
pub enum LambdaBody {
    Expr(Box<Expr>),
    Block(Block),
}

/// Match expression.
#[derive(Debug, Clone)]
pub struct MatchExpr {
    pub subject: Box<Expr>,
    pub arms: Vec<MatchArm>,
}

/// A single arm of a match expression.
#[derive(Debug, Clone)]
pub struct MatchArm {
    pub pattern: Spanned<Pattern>,
    pub guard: Option<Expr>,
    pub body: MatchArmBody,
    pub span: Span,
}

/// Body of a match arm -- either a single expression or a block.
#[derive(Debug, Clone)]
pub enum MatchArmBody {
    Expr(Expr),
    Block(Block),
}

// ---------------------------------------------------------------------------
// Patterns
// ---------------------------------------------------------------------------

/// Pattern for match arms.
#[derive(Debug, Clone)]
pub enum Pattern {
    /// `_` wildcard.
    Wildcard,
    /// A literal value pattern.
    Literal(Literal),
    /// An identifier binding.
    Ident(String),
    /// Constructor pattern: `Name(fields)`.
    Constructor {
        name: String,
        fields: Vec<Spanned<String>>,
    },
    /// Or-pattern: `A | B`.
    Or(Vec<Spanned<Pattern>>),
    /// Range pattern: `1..5` or `1..=5`.
    Range {
        start: Literal,
        end: Literal,
        inclusive: bool,
    },
}

// ---------------------------------------------------------------------------
// More expression structs
// ---------------------------------------------------------------------------

/// If expression (when used in expression position).
#[derive(Debug, Clone)]
pub struct IfExpr {
    pub if_stmt: IfStmt,
}

/// Block expression: `{ statements }`.
#[derive(Debug, Clone)]
pub struct BlockExpr {
    pub body: Block,
}

/// List literal: `[elems]`.
#[derive(Debug, Clone)]
pub struct ListExpr {
    pub elems: Vec<Expr>,
}

/// Map literal: `{ entries }`.
#[derive(Debug, Clone)]
pub struct MapExpr {
    pub entries: Vec<MapEntry>,
}

/// A key-value entry in a map literal.
#[derive(Debug, Clone)]
pub struct MapEntry {
    pub key: Expr,
    pub value: Expr,
}

/// Struct instantiation: `TypeName { fields }`.
#[derive(Debug, Clone)]
pub struct InstanceExpr {
    pub type_name: Spanned<String>,
    pub fields: Vec<InstanceField>,
}

/// A field in a struct instantiation.
#[derive(Debug, Clone)]
pub struct InstanceField {
    /// Field name (None for positional / shorthand).
    pub name: Option<Spanned<String>>,
    pub value: Expr,
    pub span: Span,
}

/// Range expression: `start..end` or `start..=end`.
#[derive(Debug, Clone)]
pub struct RangeExpr {
    pub start: Box<Expr>,
    pub end: Box<Expr>,
    pub inclusive: bool,
}

/// Error propagation: `expr?`.
#[derive(Debug, Clone)]
pub struct PropagateExpr {
    pub inner: Box<Expr>,
}

/// Or-else expression: `expr ?? default`.
#[derive(Debug, Clone)]
pub struct OrelseExpr {
    pub left: Box<Expr>,
    pub default: Box<Expr>,
}

/// Some wrapping: `some(expr)`.
#[derive(Debug, Clone)]
pub struct SomeExpr {
    pub inner: Box<Expr>,
}

/// Async block: `async { body }`.
#[derive(Debug, Clone)]
pub struct AsyncExpr {
    pub body: Block,
}

/// Spawn block: `spawn { body }`.
#[derive(Debug, Clone)]
pub struct SpawnExpr {
    pub body: Block,
}

/// Select expression with arms and optional default.
#[derive(Debug, Clone)]
pub struct SelectExpr {
    pub arms: Vec<SelectArm>,
    pub default: Option<Block>,
}

/// A single arm of a select expression.
#[derive(Debug, Clone)]
pub struct SelectArm {
    pub binding: Spanned<String>,
    pub channel: Expr,
    pub body: MatchArmBody,
    pub span: Span,
}

/// Parenthesized expression: `(inner)`.
#[derive(Debug, Clone)]
pub struct ParenExpr {
    pub inner: Box<Expr>,
}

// ---------------------------------------------------------------------------
// Blocks
// ---------------------------------------------------------------------------

/// A block of statements: `{ statements }`.
#[derive(Debug, Clone)]
pub struct Block {
    pub statements: Vec<Statement>,
    pub span: Span,
}
