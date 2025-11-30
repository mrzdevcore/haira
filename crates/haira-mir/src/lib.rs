//! Mid-level Intermediate Representation for the Haira programming language.
//!
//! MIR is a control-flow graph representation used for:
//! - Borrow checking (future)
//! - Optimization passes
//! - Lowering to machine code

use haira_types::Type;
use smol_str::SmolStr;

/// A MIR function.
pub struct MirFunction {
    pub name: SmolStr,
    pub params: Vec<MirLocal>,
    pub return_type: Type,
    pub locals: Vec<MirLocal>,
    pub blocks: Vec<BasicBlock>,
}

/// A local variable.
#[derive(Clone)]
pub struct MirLocal {
    pub name: SmolStr,
    pub ty: Type,
}

/// A basic block.
pub struct BasicBlock {
    pub id: BlockId,
    pub statements: Vec<Statement>,
    pub terminator: Terminator,
}

/// Block identifier.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub struct BlockId(pub u32);

/// MIR statement.
pub enum Statement {
    /// Assignment: place = rvalue
    Assign { place: Place, rvalue: Rvalue },
    /// Storage live marker.
    StorageLive(LocalId),
    /// Storage dead marker.
    StorageDead(LocalId),
    /// No-op.
    Nop,
}

/// Local variable ID.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub struct LocalId(pub u32);

/// A place (lvalue).
#[derive(Clone)]
pub enum Place {
    Local(LocalId),
    Field { base: Box<Place>, field: SmolStr },
    Index { base: Box<Place>, index: Box<Operand> },
}

/// An rvalue.
pub enum Rvalue {
    Use(Operand),
    BinaryOp(BinOp, Operand, Operand),
    UnaryOp(UnOp, Operand),
    Aggregate { ty: Type, fields: Vec<Operand> },
    Ref(Place),
}

/// An operand.
#[derive(Clone)]
pub enum Operand {
    Copy(Box<Place>),
    Move(Box<Place>),
    Constant(Constant),
}

/// A constant value.
#[derive(Clone)]
pub enum Constant {
    Int(i64),
    Float(f64),
    Bool(bool),
    String(SmolStr),
    Unit,
}

/// Binary operations.
#[derive(Debug, Clone, Copy)]
pub enum BinOp {
    Add, Sub, Mul, Div, Rem,
    Eq, Ne, Lt, Le, Gt, Ge,
    BitAnd, BitOr, BitXor,
    Shl, Shr,
}

/// Unary operations.
#[derive(Debug, Clone, Copy)]
pub enum UnOp {
    Neg,
    Not,
}

/// Block terminator.
pub enum Terminator {
    /// Go to another block.
    Goto(BlockId),
    /// Conditional branch.
    If {
        condition: Operand,
        then_block: BlockId,
        else_block: BlockId,
    },
    /// Function call.
    Call {
        func: SmolStr,
        args: Vec<Operand>,
        destination: Place,
        target: BlockId,
    },
    /// Return from function.
    Return,
    /// Unreachable code.
    Unreachable,
}

impl MirFunction {
    pub fn new(name: SmolStr, return_type: Type) -> Self {
        Self {
            name,
            params: Vec::new(),
            return_type,
            locals: Vec::new(),
            blocks: Vec::new(),
        }
    }
}
