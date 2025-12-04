//! High-level Intermediate Representation for the Haira programming language.
//!
//! HIR is a desugared, type-annotated version of the AST.
//! It includes resolved types, lowered constructs, and AI-generated implementations.

use haira_ast::Span;
use haira_types::Type;
use la_arena::{Arena, Idx};
use smol_str::SmolStr;

/// A HIR module.
pub struct HirModule {
    /// All functions in the module.
    pub functions: Arena<HirFunction>,
    /// All types in the module.
    pub types: Arena<HirTypeDef>,
}

pub type FunctionId = Idx<HirFunction>;
pub type TypeId = Idx<HirTypeDef>;

/// A HIR function.
pub struct HirFunction {
    pub name: SmolStr,
    pub params: Vec<HirParam>,
    pub return_type: Type,
    pub body: HirBody,
    /// Whether this function was AI-generated.
    pub ai_generated: bool,
    /// Source span for error reporting.
    pub span: Span,
}

/// A function parameter.
pub struct HirParam {
    pub name: SmolStr,
    pub ty: Type,
    /// Source span for error reporting.
    pub span: Span,
}

/// Function body.
pub struct HirBody {
    pub exprs: Arena<HirExpr>,
    pub root: Option<Idx<HirExpr>>,
}

/// A HIR expression.
pub struct HirExpr {
    pub kind: HirExprKind,
    pub ty: Type,
    /// Source span for error reporting.
    pub span: Span,
}

/// HIR expression kinds.
pub enum HirExprKind {
    /// Integer literal.
    IntLit(i64),
    /// Float literal.
    FloatLit(f64),
    /// String literal.
    StringLit(SmolStr),
    /// Boolean literal.
    BoolLit(bool),
    /// Local variable reference.
    Local(SmolStr),
    /// Binary operation.
    Binary {
        op: BinaryOp,
        lhs: Idx<HirExpr>,
        rhs: Idx<HirExpr>,
    },
    /// Unary operation.
    Unary { op: UnaryOp, operand: Idx<HirExpr> },
    /// Function call.
    Call {
        func: FunctionId,
        args: Vec<Idx<HirExpr>>,
    },
    /// Method call.
    MethodCall {
        receiver: Idx<HirExpr>,
        method: SmolStr,
        args: Vec<Idx<HirExpr>>,
    },
    /// Field access.
    Field { base: Idx<HirExpr>, field: SmolStr },
    /// Index access.
    Index {
        base: Idx<HirExpr>,
        index: Idx<HirExpr>,
    },
    /// If expression.
    If {
        condition: Idx<HirExpr>,
        then_branch: Idx<HirExpr>,
        else_branch: Option<Idx<HirExpr>>,
    },
    /// Block expression.
    Block(Vec<Idx<HirExpr>>),
    /// Let binding.
    Let {
        name: SmolStr,
        ty: Type,
        value: Idx<HirExpr>,
    },
    /// Return.
    Return(Option<Idx<HirExpr>>),
    /// Struct instantiation.
    Struct {
        ty: TypeId,
        fields: Vec<(SmolStr, Idx<HirExpr>)>,
    },
    /// Lambda.
    Lambda {
        params: Vec<HirParam>,
        body: Idx<HirExpr>,
    },
    /// Error placeholder.
    Error,
}

/// Binary operators.
#[derive(Debug, Clone, Copy)]
pub enum BinaryOp {
    Add,
    Sub,
    Mul,
    Div,
    Mod,
    Eq,
    Ne,
    Lt,
    Le,
    Gt,
    Ge,
    And,
    Or,
    BitAnd,
    BitOr,
    BitXor,
    Shl,
    Shr,
}

/// Unary operators.
#[derive(Debug, Clone, Copy)]
pub enum UnaryOp {
    Neg,
    Not,
}

/// A type definition in HIR.
pub struct HirTypeDef {
    pub name: SmolStr,
    pub kind: HirTypeDefKind,
    /// Source span for error reporting.
    pub span: Span,
}

/// Type definition kinds.
pub enum HirTypeDefKind {
    /// Struct with fields.
    Struct { fields: Vec<(SmolStr, Type)> },
    /// Enum with variants.
    Enum { variants: Vec<HirEnumVariant> },
    /// Type alias.
    Alias(Type),
}

/// An enum variant.
pub struct HirEnumVariant {
    pub name: SmolStr,
    pub fields: Vec<Type>,
}

impl HirModule {
    pub fn new() -> Self {
        Self {
            functions: Arena::new(),
            types: Arena::new(),
        }
    }
}

impl Default for HirModule {
    fn default() -> Self {
        Self::new()
    }
}
