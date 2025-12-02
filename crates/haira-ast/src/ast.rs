//! AST node definitions for Haira.

use crate::{Span, Spanned};
use smol_str::SmolStr;

/// A complete Haira source file.
#[derive(Debug, Clone, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub struct SourceFile {
    /// Top-level declarations
    pub items: Vec<Item>,
    /// Full span of the file
    pub span: Span,
}

/// A top-level item in a source file.
pub type Item = Spanned<ItemKind>;

#[derive(Debug, Clone, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub enum ItemKind {
    /// Type definition: `User { name, age, email }`
    TypeDef(TypeDef),
    /// Function definition: `greet(name) { ... }`
    FunctionDef(FunctionDef),
    /// Method definition: `User.greet() { ... }`
    MethodDef(MethodDef),
    /// Type alias: `UserId = int`
    TypeAlias(TypeAlias),
    /// AI-generated function: `ai summarize(user: User) -> Summary { ... }`
    AiFunctionDef(AiBlock),
    /// A statement at module level
    Statement(Statement),
}

// ============================================================================
// Type Definitions
// ============================================================================

/// A type definition: `User { name, age, email }`
#[derive(Debug, Clone, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub struct TypeDef {
    /// Whether this type is public
    pub is_public: bool,
    /// Type name
    pub name: Spanned<SmolStr>,
    /// Fields
    pub fields: Vec<Field>,
}

/// A field in a type definition.
#[derive(Debug, Clone, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub struct Field {
    /// Field name
    pub name: Spanned<SmolStr>,
    /// Optional type annotation
    pub ty: Option<Spanned<Type>>,
    /// Optional default value
    pub default: Option<Expr>,
    /// Span of the entire field
    pub span: Span,
}

/// A type alias: `UserId = int`
#[derive(Debug, Clone, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub struct TypeAlias {
    /// Alias name
    pub name: Spanned<SmolStr>,
    /// Target type
    pub ty: Spanned<Type>,
}

// ============================================================================
// Type Expressions
// ============================================================================

/// A type expression.
#[derive(Debug, Clone, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub enum Type {
    /// Simple named type: `int`, `string`, `User`
    Named(SmolStr),
    /// List type: `[int]`, `[User]`
    List(Box<Spanned<Type>>),
    /// Map type: `{string: int}`
    Map {
        key: Box<Spanned<Type>>,
        value: Box<Spanned<Type>>,
    },
    /// Option type: `Option<User>`
    Option(Box<Spanned<Type>>),
    /// Function type: `(int, int) -> int`
    Function {
        params: Vec<Spanned<Type>>,
        ret: Box<Spanned<Type>>,
    },
    /// Union type: `Success | Failure`
    Union(Vec<Spanned<Type>>),
    /// Generic type: `Box<T>`
    Generic {
        name: SmolStr,
        args: Vec<Spanned<Type>>,
    },
}

// ============================================================================
// Functions
// ============================================================================

/// A function definition: `add(a, b) { a + b }`
#[derive(Debug, Clone, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub struct FunctionDef {
    /// Whether this function is public
    pub is_public: bool,
    /// Function name
    pub name: Spanned<SmolStr>,
    /// Parameters
    pub params: Vec<Param>,
    /// Optional return type annotation
    pub return_ty: Option<Spanned<Type>>,
    /// Function body
    pub body: Block,
}

/// A method definition: `User.greet() { ... }`
#[derive(Debug, Clone, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub struct MethodDef {
    /// Type this method is attached to
    pub type_name: Spanned<SmolStr>,
    /// Method name
    pub name: Spanned<SmolStr>,
    /// Parameters (excluding implicit self)
    pub params: Vec<Param>,
    /// Optional return type annotation
    pub return_ty: Option<Spanned<Type>>,
    /// Method body
    pub body: Block,
}

/// A function parameter.
#[derive(Debug, Clone, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub struct Param {
    /// Parameter name
    pub name: Spanned<SmolStr>,
    /// Optional type annotation
    pub ty: Option<Spanned<Type>>,
    /// Optional default value
    pub default: Option<Expr>,
    /// Whether this is a rest parameter (`args...`)
    pub is_rest: bool,
    /// Span of the entire parameter
    pub span: Span,
}

// ============================================================================
// Statements
// ============================================================================

/// A statement.
pub type Statement = Spanned<StatementKind>;

#[derive(Debug, Clone, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub enum StatementKind {
    /// Variable assignment: `x = 42`
    Assignment(Assignment),
    /// If statement: `if cond { ... }`
    If(IfStatement),
    /// For loop: `for x in items { ... }`
    For(ForStatement),
    /// While loop: `while cond { ... }`
    While(WhileStatement),
    /// Match statement: `match x { ... }`
    Match(MatchExpr),
    /// Return statement: `return x`
    Return(ReturnStatement),
    /// Try-catch: `try { ... } catch e { ... }`
    Try(TryStatement),
    /// Break statement
    Break,
    /// Continue statement
    Continue,
    /// Expression statement
    Expr(Expr),
}

/// An assignment: `x = 42` or `x, y = get_pair()`
#[derive(Debug, Clone, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub struct Assignment {
    /// Assignment target(s)
    pub targets: Vec<AssignTarget>,
    /// Value being assigned
    pub value: Expr,
}

/// An assignment target.
#[derive(Debug, Clone, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub struct AssignTarget {
    /// The path being assigned to (variable, field, or index)
    pub path: AssignPath,
    /// Optional type annotation (only valid for simple identifiers)
    pub ty: Option<Spanned<Type>>,
}

/// A path that can be assigned to.
#[derive(Debug, Clone, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub enum AssignPath {
    /// Simple variable: `x`
    Identifier(Spanned<SmolStr>),
    /// Field access: `obj.field`
    Field {
        object: Box<AssignPath>,
        field: Spanned<SmolStr>,
    },
    /// Index access: `arr[index]`
    Index {
        object: Box<AssignPath>,
        index: Box<Expr>,
    },
}

/// An if statement.
#[derive(Debug, Clone, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub struct IfStatement {
    /// Condition
    pub condition: Expr,
    /// Then branch
    pub then_branch: Block,
    /// Optional else branch (can be another if for else-if chains)
    pub else_branch: Option<ElseBranch>,
}

/// An else branch - either a block or another if.
#[derive(Debug, Clone, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub enum ElseBranch {
    Block(Block),
    ElseIf(Box<Spanned<IfStatement>>),
}

/// A for loop: `for x in items { ... }`
#[derive(Debug, Clone, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub struct ForStatement {
    /// Loop variable(s): `x` or `i, x` for indexed
    pub pattern: ForPattern,
    /// Iterator expression
    pub iterator: Expr,
    /// Loop body
    pub body: Block,
}

/// A for loop pattern.
#[derive(Debug, Clone, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub enum ForPattern {
    /// Single variable: `for x in items`
    Single(Spanned<SmolStr>),
    /// Two variables: `for i, x in items` or `for k, v in map`
    Pair(Spanned<SmolStr>, Spanned<SmolStr>),
}

/// A while loop: `while cond { ... }`
#[derive(Debug, Clone, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub struct WhileStatement {
    /// Condition
    pub condition: Expr,
    /// Loop body
    pub body: Block,
}

/// A return statement: `return x` or `return x, y`
#[derive(Debug, Clone, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub struct ReturnStatement {
    /// Values to return (empty for bare `return`)
    pub values: Vec<Expr>,
}

/// A try-catch statement.
#[derive(Debug, Clone, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub struct TryStatement {
    /// Try body
    pub body: Block,
    /// Error variable name in catch
    pub error_name: Spanned<SmolStr>,
    /// Catch body
    pub catch_body: Block,
}

// ============================================================================
// Expressions
// ============================================================================

/// An expression.
pub type Expr = Spanned<ExprKind>;

#[derive(Debug, Clone, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub enum ExprKind {
    /// Literal value
    Literal(Literal),
    /// Identifier: `foo`
    Identifier(SmolStr),
    /// Binary operation: `a + b`
    Binary(BinaryExpr),
    /// Unary operation: `-x`, `not x`
    Unary(UnaryExpr),
    /// Function call: `foo(x, y)`
    Call(CallExpr),
    /// Method call: `obj.method(x)`
    MethodCall(MethodCallExpr),
    /// Field access: `obj.field`
    Field(FieldExpr),
    /// Index access: `arr[0]`
    Index(IndexExpr),
    /// Pipe expression: `x | f | g`
    Pipe(PipeExpr),
    /// Lambda: `(x) { x * 2 }` or `x => x * 2`
    Lambda(LambdaExpr),
    /// Match expression
    Match(MatchExpr),
    /// If expression (when used as expression)
    If(Box<IfStatement>),
    /// Block expression: `{ ... }`
    Block(Block),
    /// List literal: `[1, 2, 3]`
    List(Vec<Expr>),
    /// Map literal: `{ "a": 1, "b": 2 }`
    Map(Vec<(Expr, Expr)>),
    /// Type instantiation: `User { name = "Alice", age = 30 }`
    Instance(InstanceExpr),
    /// Range: `0..10` or `0..=10`
    Range(RangeExpr),
    /// Error propagation: `expr?`
    Propagate(Box<Expr>),
    /// Some constructor: `some(x)`
    Some(Box<Expr>),
    /// None literal
    None,
    /// Async block: `async { ... }`
    Async(Block),
    /// Spawn block: `spawn { ... }`
    Spawn(Block),
    /// Select expression
    Select(SelectExpr),
    /// Parenthesized expression
    Paren(Box<Expr>),
    /// AI intent block: `ai func_name(params) -> Type { intent }` or anonymous `ai(params) { intent }`
    Ai(AiBlock),
}

/// A literal value.
#[derive(Debug, Clone, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub enum Literal {
    /// Integer: `42`
    Int(i64),
    /// Float: `3.14`
    Float(f64),
    /// String: `"hello"`
    String(SmolStr),
    /// Interpolated string parts
    InterpolatedString(Vec<StringPart>),
    /// Boolean: `true`, `false`
    Bool(bool),
}

/// A part of an interpolated string.
#[derive(Debug, Clone, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub enum StringPart {
    /// Literal text
    Literal(SmolStr),
    /// Interpolated expression: `{expr}`
    Expr(Expr),
}

/// A binary expression: `a + b`
#[derive(Debug, Clone, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub struct BinaryExpr {
    pub left: Box<Expr>,
    pub op: Spanned<BinaryOp>,
    pub right: Box<Expr>,
}

/// Binary operators.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub enum BinaryOp {
    // Arithmetic
    Add, // +
    Sub, // -
    Mul, // *
    Div, // /
    Mod, // %
    // Comparison
    Eq, // ==
    Ne, // !=
    Lt, // <
    Gt, // >
    Le, // <=
    Ge, // >=
    // Logical
    And, // and
    Or,  // or
}

/// A unary expression: `-x`, `not x`
#[derive(Debug, Clone, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub struct UnaryExpr {
    pub op: Spanned<UnaryOp>,
    pub operand: Box<Expr>,
}

/// Unary operators.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub enum UnaryOp {
    Neg, // -
    Not, // not
}

/// A function call: `foo(x, y)`
#[derive(Debug, Clone, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub struct CallExpr {
    /// Function being called
    pub callee: Box<Expr>,
    /// Arguments
    pub args: Vec<Argument>,
}

/// A method call: `obj.method(x)`
#[derive(Debug, Clone, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub struct MethodCallExpr {
    /// Object
    pub receiver: Box<Expr>,
    /// Method name
    pub method: Spanned<SmolStr>,
    /// Arguments
    pub args: Vec<Argument>,
}

/// A function argument.
#[derive(Debug, Clone, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub struct Argument {
    /// Optional name for named arguments
    pub name: Option<Spanned<SmolStr>>,
    /// Argument value
    pub value: Expr,
    /// Span of the entire argument
    pub span: Span,
}

/// A field access: `obj.field`
#[derive(Debug, Clone, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub struct FieldExpr {
    pub object: Box<Expr>,
    pub field: Spanned<SmolStr>,
}

/// An index access: `arr[0]`
#[derive(Debug, Clone, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub struct IndexExpr {
    pub object: Box<Expr>,
    pub index: Box<Expr>,
}

/// A pipe expression: `x | f | g`
#[derive(Debug, Clone, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub struct PipeExpr {
    pub left: Box<Expr>,
    pub right: Box<Expr>,
}

/// A lambda expression.
#[derive(Debug, Clone, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub struct LambdaExpr {
    /// Parameters
    pub params: Vec<Param>,
    /// Body - either a single expression or a block
    pub body: LambdaBody,
}

/// Lambda body - expression or block.
#[derive(Debug, Clone, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub enum LambdaBody {
    /// Arrow expression: `x => x * 2`
    Expr(Box<Expr>),
    /// Block: `(x) { x * 2 }`
    Block(Block),
}

/// A match expression.
#[derive(Debug, Clone, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub struct MatchExpr {
    /// Value being matched
    pub subject: Box<Expr>,
    /// Match arms
    pub arms: Vec<MatchArm>,
}

/// A match arm.
#[derive(Debug, Clone, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub struct MatchArm {
    /// Pattern to match
    pub pattern: Spanned<Pattern>,
    /// Optional guard: `if condition`
    pub guard: Option<Expr>,
    /// Body
    pub body: MatchArmBody,
    /// Span of the entire arm
    pub span: Span,
}

/// Match arm body - expression or block.
#[derive(Debug, Clone, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub enum MatchArmBody {
    Expr(Expr),
    Block(Block),
}

/// A pattern in a match arm.
#[derive(Debug, Clone, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub enum Pattern {
    /// Wildcard: `_`
    Wildcard,
    /// Literal: `42`, `"hello"`
    Literal(Literal),
    /// Identifier binding: `x`
    Identifier(SmolStr),
    /// Constructor pattern: `Some { value }` or `User { name, age }`
    Constructor {
        name: SmolStr,
        fields: Vec<Spanned<SmolStr>>,
    },
}

/// Type instantiation: `User { name = "Alice" }`
#[derive(Debug, Clone, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub struct InstanceExpr {
    /// Type name
    pub type_name: Spanned<SmolStr>,
    /// Fields (can be positional or named)
    pub fields: Vec<InstanceField>,
}

/// An instance field.
#[derive(Debug, Clone, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub struct InstanceField {
    /// Optional field name (None for positional)
    pub name: Option<Spanned<SmolStr>>,
    /// Field value
    pub value: Expr,
    /// Span
    pub span: Span,
}

/// A range expression: `0..10` or `0..=10`
#[derive(Debug, Clone, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub struct RangeExpr {
    /// Start value
    pub start: Box<Expr>,
    /// End value
    pub end: Box<Expr>,
    /// Whether end is inclusive (`..=`)
    pub inclusive: bool,
}

/// A select expression for channel operations.
#[derive(Debug, Clone, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub struct SelectExpr {
    /// Select arms
    pub arms: Vec<SelectArm>,
    /// Optional default arm
    pub default: Option<Block>,
}

/// A select arm.
#[derive(Debug, Clone, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub struct SelectArm {
    /// Variable to bind received value
    pub binding: Spanned<SmolStr>,
    /// Channel expression
    pub channel: Expr,
    /// Body
    pub body: MatchArmBody,
    /// Span
    pub span: Span,
}

/// An AI intent block for explicit AI-generated functions.
///
/// Syntax:
/// ```haira
/// // Named function
/// ai summarize_activity(user: User) -> ActivitySummary {
///     Summarize the user activity over the last 30 days.
///     Group by activity type and find most common.
/// }
///
/// // Anonymous/inline
/// result = ai(data: Data) -> Summary {
///     Analyze and summarize this data
/// }(my_data)
/// ```
#[derive(Debug, Clone, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub struct AiBlock {
    /// Optional function name (None for anonymous)
    pub name: Option<Spanned<SmolStr>>,
    /// Parameters
    pub params: Vec<Param>,
    /// Optional return type annotation
    pub return_ty: Option<Spanned<Type>>,
    /// Natural language intent description (the block body)
    pub intent: SmolStr,
    /// Span of the entire block
    pub span: Span,
}

// ============================================================================
// Blocks
// ============================================================================

/// A block of statements.
#[derive(Debug, Clone, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub struct Block {
    /// Statements in the block
    pub statements: Vec<Statement>,
    /// Span of the entire block including braces
    pub span: Span,
}

impl Block {
    /// Create an empty block.
    pub fn empty(span: Span) -> Self {
        Self {
            statements: Vec::new(),
            span,
        }
    }
}
