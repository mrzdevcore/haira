//! Haira HIR — High-level Intermediate Representation.
//!
//! The HIR is a lowered form of the AST that:
//! - Has explicit control flow (no implicit else-if chains, match desugared)
//! - Retains source spans for debugging
//! - Is easy to lower to both Cranelift IR and LLVM IR
//! - Retains agentic constructs as high-level operations (AgentAsk, ToolRegister, etc.)
//!
//! The HIR uses a **function-based** representation where every top-level item becomes a
//! [`HirFunction`] or [`HirDecl`]. Inside functions, the body is a sequence of [`HirInst`]
//! (instructions) organized into [`HirBlock`]s. Each block ends with a [`Terminator`] that
//! describes how control leaves the block.

use std::collections::HashMap;

use haira_ast::{self as ast, Span, SourceFile};
use haira_errors::Diagnostic;

// ===========================================================================
// ID types
// ===========================================================================

/// Index of a basic block within a function.
pub type BlockId = usize;

/// Index of a variable (SSA-like).
pub type VarId = usize;

/// Index of a function within the module.
pub type FuncId = usize;

// ===========================================================================
// HIR module
// ===========================================================================

/// A complete HIR module (one source file or merged program).
#[derive(Debug, Clone)]
pub struct HirModule {
    /// All functions (including tools, workflows, methods, lambdas).
    pub functions: Vec<HirFunction>,
    /// Top-level declarations (structs, enums, providers, agents, globals).
    pub declarations: Vec<HirDecl>,
    /// Index into `functions` of the entry point (main/init), if any.
    pub entry: Option<usize>,
}

// ===========================================================================
// Declarations
// ===========================================================================

/// A top-level declaration (struct, enum, provider, agent, etc.).
#[derive(Debug, Clone)]
pub enum HirDecl {
    Struct {
        name: String,
        fields: Vec<(String, HirType)>,
    },
    Enum {
        name: String,
        variants: Vec<(String, Vec<HirType>)>,
    },
    Provider {
        name: String,
        fields: Vec<(String, HirConst)>,
    },
    Agent {
        name: String,
        fields: Vec<(String, HirConst)>,
    },
    Global {
        name: String,
        ty: HirType,
        init: Option<HirConst>,
    },
}

// ===========================================================================
// Functions
// ===========================================================================

/// A function in the HIR (includes tools, workflows, methods, lambdas).
#[derive(Debug, Clone)]
pub struct HirFunction {
    pub name: String,
    pub params: Vec<HirParam>,
    pub return_ty: HirType,
    pub blocks: Vec<HirBlock>,
    pub is_workflow: bool,
    pub is_tool: bool,
    pub is_async: bool,
    pub span: Span,
}

/// A function parameter.
#[derive(Debug, Clone)]
pub struct HirParam {
    pub name: String,
    pub ty: HirType,
    pub default: Option<HirConst>,
    pub is_rest: bool,
}

// ===========================================================================
// Basic blocks
// ===========================================================================

/// A basic block -- a sequence of instructions ending with a terminator.
#[derive(Debug, Clone)]
pub struct HirBlock {
    pub id: BlockId,
    pub insts: Vec<HirInst>,
    pub terminator: Terminator,
}

// ===========================================================================
// Instructions
// ===========================================================================

/// A single HIR instruction.
#[derive(Debug, Clone)]
pub enum HirInst {
    /// Assign a value: `var = value`.
    Assign { dst: VarId, value: HirValue },
    /// Call a function/tool: `var = func(args)`.
    Call {
        dst: Option<VarId>,
        func: FuncRef,
        args: Vec<VarId>,
    },
    /// Method call: `var = receiver.method(args)`.
    MethodCall {
        dst: Option<VarId>,
        receiver: VarId,
        method: String,
        args: Vec<VarId>,
    },
    /// Field access: `var = obj.field`.
    GetField {
        dst: VarId,
        object: VarId,
        field: String,
    },
    /// Field set: `obj.field = value`.
    SetField {
        object: VarId,
        field: String,
        value: VarId,
    },
    /// Index access: `var = obj[index]`.
    GetIndex {
        dst: VarId,
        object: VarId,
        index: VarId,
    },
    /// Index set: `obj[index] = value`.
    SetIndex {
        object: VarId,
        index: VarId,
        value: VarId,
    },
    /// Construct a struct: `var = TypeName { fields }`.
    ConstructStruct {
        dst: VarId,
        type_name: String,
        fields: Vec<(String, VarId)>,
    },
    /// Construct a list: `var = [elems]`.
    ConstructList { dst: VarId, elems: Vec<VarId> },
    /// Construct a map: `var = { entries }`.
    ConstructMap {
        dst: VarId,
        entries: Vec<(VarId, VarId)>,
    },
    /// Binary operation: `var = left op right`.
    BinOp {
        dst: VarId,
        op: BinOp,
        left: VarId,
        right: VarId,
    },
    /// Unary operation: `var = op operand`.
    UnOp {
        dst: VarId,
        op: UnOp,
        operand: VarId,
    },
    /// Load a constant value into a variable.
    Const { dst: VarId, value: HirConst },
    /// Spawn a block as a concurrent task.
    Spawn { dst: VarId, block: BlockId },
    /// Channel receive.
    ChanRecv { dst: VarId, channel: VarId },
    /// Channel send.
    ChanSend { channel: VarId, value: VarId },

    // -- Agentic high-level ops --

    /// Register a tool with the runtime.
    ToolRegister {
        name: String,
        func: FuncRef,
        description: String,
    },
    /// Ask an agent: `var = agent.ask(prompt)`.
    AgentAsk {
        dst: VarId,
        agent: String,
        prompt: VarId,
        output_ty: Option<HirType>,
    },
    /// Run an agent: `var = agent.run(prompt)`.
    AgentRun {
        dst: VarId,
        agent: String,
        prompt: VarId,
    },
    /// Stream from an agent: `var = agent.stream(prompt)`.
    AgentStream {
        dst: VarId,
        agent: String,
        prompt: VarId,
    },
    /// Error propagation: unwrap or panic.
    Propagate { dst: VarId, inner: VarId },
    /// No-op (placeholder).
    Nop,
}

/// A simple value that can be assigned directly (used in Assign instructions).
#[derive(Debug, Clone)]
pub enum HirValue {
    /// Copy the value of another variable.
    Use(VarId),
    /// A constant value.
    Const(HirConst),
}

// ===========================================================================
// Terminators
// ===========================================================================

/// Block terminator -- how control leaves a basic block.
#[derive(Debug, Clone)]
pub enum Terminator {
    /// Unconditional jump to another block.
    Goto(BlockId),
    /// Conditional branch.
    Branch {
        cond: VarId,
        then_block: BlockId,
        else_block: BlockId,
    },
    /// Return from the function.
    Return(Option<VarId>),
    /// Match/switch dispatch with constant cases and a default block.
    Switch {
        scrutinee: VarId,
        cases: Vec<(HirConst, BlockId)>,
        default: BlockId,
    },
    /// Unreachable (after panic, etc.).
    Unreachable,
}

// ===========================================================================
// References
// ===========================================================================

/// A reference to a function (local, builtin, or external).
#[derive(Debug, Clone)]
pub enum FuncRef {
    /// Index into the module's function list.
    Local(FuncId),
    /// Built-in function name (e.g. "io.println", "len").
    Builtin(String),
    /// External / stdlib function name.
    External(String),
}

// ===========================================================================
// Constants
// ===========================================================================

/// Constant/literal values representable at compile time.
#[derive(Debug, Clone, PartialEq)]
pub enum HirConst {
    None,
    Bool(bool),
    Int(i64),
    Float(f64),
    Str(String),
    List(Vec<HirConst>),
    Map(Vec<(String, HirConst)>),
}

// ===========================================================================
// Operators
// ===========================================================================

/// Binary operators.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum BinOp {
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

/// Unary operators.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum UnOp {
    Neg,
    Not,
    BitNot,
}

// ===========================================================================
// Types
// ===========================================================================

/// HIR type representation -- mirrors the language's type system.
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum HirType {
    Void,
    Bool,
    Int,
    Float,
    Str,
    Any,
    List(Box<HirType>),
    Map(Box<HirType>, Box<HirType>),
    Struct(String),
    Enum(String),
    Func {
        params: Vec<HirType>,
        ret: Box<HirType>,
    },
    Option(Box<HirType>),
    Channel(Box<HirType>),
    Stream,
}

// ===========================================================================
// AST -> HIR lowering
// ===========================================================================

/// Lower an AST [`SourceFile`] into a [`HirModule`].
///
/// Returns the HIR module together with any diagnostics (warnings/errors)
/// produced during lowering. The lowering handles common constructs:
/// functions, let/assign, if/else, for/while, return, calls, binary/unary
/// ops, literals, field/index access, struct construction, list/map literals.
pub fn lower(file: &SourceFile) -> (HirModule, Vec<Diagnostic>) {
    let mut ctx = LowerCtx::new();
    ctx.lower_source_file(file);
    let module = ctx.module;
    let diags = ctx.diags;
    (module, diags)
}

// ===========================================================================
// Lowering context (internal)
// ===========================================================================

/// Internal mutable state threaded through the lowering pass.
struct LowerCtx {
    module: HirModule,
    /// Index of the function currently being built.
    current_func: usize,
    /// Index of the block currently being appended to.
    current_block: BlockId,
    /// Counter for fresh variable IDs.
    next_var: VarId,
    /// Name-to-VarId mapping for the current scope.
    vars: HashMap<String, VarId>,
    /// Accumulated diagnostics.
    diags: Vec<Diagnostic>,
    /// Map from function name to FuncId (for resolving calls).
    func_names: HashMap<String, FuncId>,
    /// Whether we are currently building blocks for a function (vs. top-level init).
    in_function: bool,
    /// Stack of deferred expressions for the current function (LIFO order).
    deferred: Vec<ast::Expr>,
}

impl LowerCtx {
    fn new() -> Self {
        Self {
            module: HirModule {
                functions: Vec::new(),
                declarations: Vec::new(),
                entry: None,
            },
            current_func: 0,
            current_block: 0,
            next_var: 0,
            vars: HashMap::new(),
            diags: Vec::new(),
            func_names: HashMap::new(),
            in_function: false,
            deferred: Vec::new(),
        }
    }

    // -----------------------------------------------------------------------
    // Helpers
    // -----------------------------------------------------------------------

    /// Allocate a new unique variable ID.
    fn fresh_var(&mut self) -> VarId {
        let id = self.next_var;
        self.next_var += 1;
        id
    }

    /// Create a new block in the current function and return its ID.
    fn new_block(&mut self) -> BlockId {
        let func = &mut self.module.functions[self.current_func];
        let id = func.blocks.len();
        func.blocks.push(HirBlock {
            id,
            insts: Vec::new(),
            terminator: Terminator::Unreachable, // placeholder
        });
        id
    }

    /// Append an instruction to the current block.
    fn emit(&mut self, inst: HirInst) {
        let func = &mut self.module.functions[self.current_func];
        func.blocks[self.current_block].insts.push(inst);
    }

    /// Set the terminator of the current block.
    fn set_terminator(&mut self, term: Terminator) {
        let func = &mut self.module.functions[self.current_func];
        func.blocks[self.current_block].terminator = term;
    }

    /// Switch to emitting instructions into the given block.
    fn switch_to_block(&mut self, id: BlockId) {
        self.current_block = id;
    }

    /// Start building a new function, returning its FuncId.
    fn begin_function(
        &mut self,
        name: String,
        params: Vec<HirParam>,
        return_ty: HirType,
        is_workflow: bool,
        is_tool: bool,
        is_async: bool,
        span: Span,
    ) -> FuncId {
        let id = self.module.functions.len();
        self.module.functions.push(HirFunction {
            name: name.clone(),
            params,
            return_ty,
            blocks: Vec::new(),
            is_workflow,
            is_tool,
            is_async,
            span,
        });
        self.func_names.insert(name, id);
        self.current_func = id;
        // Each function gets its own VarId namespace starting from 0.
        self.next_var = 0;
        // Reset deferred stack for new function scope.
        self.deferred.clear();
        // Create the entry block (block 0).
        let entry_block = self.new_block();
        self.switch_to_block(entry_block);
        id
    }

    // -----------------------------------------------------------------------
    // Type lowering
    // -----------------------------------------------------------------------

    fn lower_type(&mut self, ty: &ast::Type) -> HirType {
        match ty {
            ast::Type::Named(name) => match name.as_str() {
                "void" => HirType::Void,
                "bool" => HirType::Bool,
                "int" => HirType::Int,
                "float" => HirType::Float,
                "string" => HirType::Str,
                "any" => HirType::Any,
                "stream" => HirType::Stream,
                _ => HirType::Struct(name.clone()),
            },
            ast::Type::List(inner) => {
                HirType::List(Box::new(self.lower_type(&inner.node)))
            }
            ast::Type::Map { key, value } => HirType::Map(
                Box::new(self.lower_type(&key.node)),
                Box::new(self.lower_type(&value.node)),
            ),
            ast::Type::Option(inner) => {
                HirType::Option(Box::new(self.lower_type(&inner.node)))
            }
            ast::Type::Function { params, ret } => {
                let p = params.iter().map(|t| self.lower_type(&t.node)).collect();
                let r = self.lower_type(&ret.node);
                HirType::Func {
                    params: p,
                    ret: Box::new(r),
                }
            }
            ast::Type::Union(_) => {
                // Unions are not directly representable in HIR; lower to Any.
                HirType::Any
            }
            ast::Type::Generic { name, .. } => {
                // Generics are monomorphized or erased; lower to named for now.
                HirType::Struct(name.clone())
            }
        }
    }

    fn lower_optional_type(&mut self, ty: &Option<ast::Spanned<ast::Type>>) -> HirType {
        match ty {
            Some(t) => self.lower_type(&t.node),
            None => HirType::Void,
        }
    }

    // -----------------------------------------------------------------------
    // Param lowering
    // -----------------------------------------------------------------------

    fn lower_param(&mut self, p: &ast::Param) -> HirParam {
        let ty = self.lower_optional_type(&p.ty);
        let default = p.default.as_ref().and_then(|e| self.try_const_expr(e));
        HirParam {
            name: p.name.node.clone(),
            ty,
            default,
            is_rest: p.is_rest,
        }
    }

    /// Try to evaluate a simple expression as a compile-time constant.
    fn try_const_expr(&self, expr: &ast::Expr) -> Option<HirConst> {
        match &expr.node {
            ast::ExprKind::Literal(lit) => Some(self.lower_literal(lit)),
            ast::ExprKind::None => Some(HirConst::None),
            _ => None,
        }
    }

    /// Lower an interpolated string like `"Hello ${name}!"` into a chain of
    /// string concatenation operations: `"Hello " + to_string(name) + "!"`.
    fn lower_interpolated_string(&mut self, parts: &[ast::StringPart]) -> VarId {
        // Start with an empty string.
        let mut result = self.fresh_var();
        self.emit(HirInst::Const {
            dst: result,
            value: HirConst::Str(String::new()),
        });

        for part in parts {
            let part_var = match part {
                ast::StringPart::Literal(s) => {
                    let v = self.fresh_var();
                    self.emit(HirInst::Const {
                        dst: v,
                        value: HirConst::Str(s.clone()),
                    });
                    v
                }
                ast::StringPart::Expr(expr) => {
                    let v = self.lower_expr(expr);
                    // Wrap in to_string call so non-string values get converted.
                    let str_v = self.fresh_var();
                    self.emit(HirInst::Call {
                        dst: Some(str_v),
                        func: FuncRef::Builtin("to_string".to_string()),
                        args: vec![v],
                    });
                    str_v
                }
            };
            let new_result = self.fresh_var();
            self.emit(HirInst::BinOp {
                dst: new_result,
                op: BinOp::Add,
                left: result,
                right: part_var,
            });
            result = new_result;
        }

        result
    }

    fn lower_literal(&self, lit: &ast::Literal) -> HirConst {
        match lit {
            ast::Literal::Int(v) => HirConst::Int(*v),
            ast::Literal::Float(v) => HirConst::Float(*v),
            ast::Literal::String(s) => HirConst::Str(s.clone()),
            ast::Literal::Bool(b) => HirConst::Bool(*b),
            ast::Literal::InterpolatedString(_) => {
                // Cannot be a constant at compile time; return empty string placeholder.
                HirConst::Str(String::new())
            }
        }
    }

    // -----------------------------------------------------------------------
    // Source file
    // -----------------------------------------------------------------------

    fn lower_source_file(&mut self, file: &SourceFile) {
        // Collect top-level statements and find the main function definition.
        let mut top_level_stmts: Vec<&ast::Statement> = Vec::new();
        let mut main_def: Option<(&ast::FunctionDef, Span)> = None;

        for item in &file.items {
            match &item.node {
                ast::ItemKind::TypeDef(td) => self.lower_type_def(td),
                ast::ItemKind::EnumDef(ed) => self.lower_enum_def(ed),
                ast::ItemKind::TypeAlias(_) => {
                    // Type aliases are erased during lowering.
                }
                ast::ItemKind::ImportDecl(_) | ast::ItemKind::ExportDecl(_) => {
                    // Imports/exports are resolved earlier; nothing to emit.
                }
                ast::ItemKind::ProviderDecl(pd) => self.lower_provider_decl(pd),
                ast::ItemKind::AgentDecl(ad) => self.lower_agent_decl(ad),
                ast::ItemKind::ToolDecl(td) => self.lower_tool_decl(td, item.span),
                ast::ItemKind::WorkflowDecl(wd) => self.lower_workflow_decl(wd, item.span),
                ast::ItemKind::FunctionDef(fd) => {
                    // Defer main() — we'll inline it with top-level stmts.
                    if fd.name.node == "main" {
                        main_def = Some((fd, item.span));
                    } else {
                        self.lower_function_def(fd, item.span);
                    }
                }
                ast::ItemKind::MethodDef(md) => self.lower_method_def(md, item.span),
                ast::ItemKind::TestDecl(td) => self.lower_test_decl(td, item.span),
                ast::ItemKind::ItemStatement(stmt) => {
                    top_level_stmts.push(stmt);
                }
            }
        }

        // Build the entry function: top-level stmts + main() body in the same scope.
        let has_entry = !top_level_stmts.is_empty() || main_def.is_some();
        if has_entry {
            let saved_vars = std::mem::take(&mut self.vars);
            let func_id = self.begin_function(
                "__haira_main".to_string(),
                Vec::new(),
                HirType::Void,
                false,
                false,
                false,
                file.span,
            );
            self.in_function = true;

            // Lower top-level statements first (e.g. `d = 10`).
            for stmt in &top_level_stmts {
                self.lower_stmt(stmt);
            }

            // Then inline main()'s body so it shares the same variable scope.
            if let Some((fd, _span)) = main_def {
                self.lower_block_stmts(&fd.body);
            }

            // Terminate the last block with Return if not already terminated.
            self.ensure_terminated(Terminator::Return(None));

            self.module.entry = Some(func_id);
            self.in_function = false;
            self.vars = saved_vars;
        }
    }

    // -----------------------------------------------------------------------
    // Item lowering
    // -----------------------------------------------------------------------

    fn lower_type_def(&mut self, td: &ast::TypeDef) {
        let fields = td
            .fields
            .iter()
            .map(|f| {
                let ty = match &f.ty {
                    Some(t) => self.lower_type(&t.node),
                    None => HirType::Any,
                };
                (f.name.node.clone(), ty)
            })
            .collect();
        self.module.declarations.push(HirDecl::Struct {
            name: td.name.node.clone(),
            fields,
        });
    }

    fn lower_enum_def(&mut self, ed: &ast::EnumDef) {
        let variants = ed
            .variants
            .iter()
            .map(|v| {
                let fields = v
                    .fields
                    .iter()
                    .map(|p| self.lower_optional_type(&p.ty))
                    .collect();
                (v.name.node.clone(), fields)
            })
            .collect();
        self.module.declarations.push(HirDecl::Enum {
            name: ed.name.node.clone(),
            variants,
        });
    }

    fn lower_provider_decl(&mut self, pd: &ast::ProviderDecl) {
        let fields = pd
            .fields
            .iter()
            .map(|f| {
                let val = self
                    .try_const_expr(&f.value)
                    .unwrap_or(HirConst::None);
                (f.key.node.clone(), val)
            })
            .collect();
        self.module.declarations.push(HirDecl::Provider {
            name: pd.name.node.clone(),
            fields,
        });
    }

    fn lower_agent_decl(&mut self, ad: &ast::AgentDecl) {
        let fields = ad
            .fields
            .iter()
            .map(|f| {
                let val = self
                    .try_const_expr(&f.value)
                    .unwrap_or(HirConst::None);
                (f.key.node.clone(), val)
            })
            .collect();
        self.module.declarations.push(HirDecl::Agent {
            name: ad.name.node.clone(),
            fields,
        });
    }

    fn lower_tool_decl(&mut self, td: &ast::ToolDecl, item_span: Span) {
        let params: Vec<HirParam> = td.params.iter().map(|p| self.lower_param(p)).collect();
        let return_ty = self.lower_optional_type(&td.return_ty);

        let saved_vars = std::mem::take(&mut self.vars);
        let func_id = self.begin_function(
            td.name.node.clone(),
            params.clone(),
            return_ty,
            false,
            true,
            false,
            item_span,
        );
        self.in_function = true;

        // Bind parameters as variables.
        for p in &params {
            let vid = self.fresh_var();
            self.vars.insert(p.name.clone(), vid);
        }

        if let Some(body) = &td.body {
            self.lower_block_stmts(body);
        }

        self.ensure_terminated(Terminator::Return(None));
        self.in_function = false;
        self.vars = saved_vars;

        // Emit a ToolRegister instruction — this would go into the init function,
        // but for now we just record it. Callers can check is_tool on the function.
        let _ = func_id; // used by name lookup
    }

    fn lower_workflow_decl(&mut self, wd: &ast::WorkflowDecl, item_span: Span) {
        let params: Vec<HirParam> = wd.params.iter().map(|p| self.lower_param(p)).collect();
        let return_ty = self.lower_optional_type(&wd.return_ty);

        let saved_vars = std::mem::take(&mut self.vars);
        self.begin_function(
            wd.name.node.clone(),
            params.clone(),
            return_ty,
            true,
            false,
            false,
            item_span,
        );
        self.in_function = true;

        for p in &params {
            let vid = self.fresh_var();
            self.vars.insert(p.name.clone(), vid);
        }

        self.lower_block_stmts(&wd.body);
        self.ensure_terminated(Terminator::Return(None));

        self.in_function = false;
        self.vars = saved_vars;
    }

    fn lower_function_def(&mut self, fd: &ast::FunctionDef, item_span: Span) {
        let params: Vec<HirParam> = fd.params.iter().map(|p| self.lower_param(p)).collect();
        let return_ty = self.lower_optional_type(&fd.return_ty);

        let saved_vars = std::mem::take(&mut self.vars);
        self.begin_function(
            fd.name.node.clone(),
            params.clone(),
            return_ty,
            false,
            false,
            false,
            item_span,
        );
        self.in_function = true;

        for p in &params {
            let vid = self.fresh_var();
            self.vars.insert(p.name.clone(), vid);
        }

        self.lower_block_stmts(&fd.body);
        self.ensure_terminated(Terminator::Return(None));

        self.in_function = false;
        self.vars = saved_vars;
    }

    fn lower_method_def(&mut self, md: &ast::MethodDef, item_span: Span) {
        // Methods get a `self` parameter prepended.
        let self_param = HirParam {
            name: "self".to_string(),
            ty: HirType::Struct(md.type_name.node.clone()),
            default: None,
            is_rest: false,
        };
        let mut params = vec![self_param];
        params.extend(md.params.iter().map(|p| self.lower_param(p)));

        let return_ty = self.lower_optional_type(&md.return_ty);
        let func_name = format!("{}.{}", md.type_name.node, md.name.node);

        let saved_vars = std::mem::take(&mut self.vars);
        self.begin_function(
            func_name,
            params.clone(),
            return_ty,
            false,
            false,
            false,
            item_span,
        );
        self.in_function = true;

        for p in &params {
            let vid = self.fresh_var();
            self.vars.insert(p.name.clone(), vid);
        }

        self.lower_block_stmts(&md.body);
        self.ensure_terminated(Terminator::Return(None));

        self.in_function = false;
        self.vars = saved_vars;
    }

    fn lower_test_decl(&mut self, td: &ast::TestDecl, item_span: Span) {
        let func_name = format!("__test_{}", td.name.node);

        let saved_vars = std::mem::take(&mut self.vars);
        self.begin_function(
            func_name,
            Vec::new(),
            HirType::Void,
            false,
            false,
            false,
            item_span,
        );
        self.in_function = true;

        self.lower_block_stmts(&td.body);
        self.ensure_terminated(Terminator::Return(None));

        self.in_function = false;
        self.vars = saved_vars;
    }

    // -----------------------------------------------------------------------
    // Block / statement lowering
    // -----------------------------------------------------------------------

    fn lower_block_stmts(&mut self, block: &ast::Block) {
        for stmt in &block.statements {
            self.lower_stmt(stmt);
        }
    }

    fn lower_stmt(&mut self, stmt: &ast::Statement) {
        match &stmt.node {
            ast::StmtKind::Let(let_stmt) => self.lower_let_stmt(let_stmt),
            ast::StmtKind::Assign(assign_stmt) => self.lower_assign_stmt(assign_stmt),
            ast::StmtKind::If(if_stmt) => self.lower_if_stmt(if_stmt),
            ast::StmtKind::For(for_stmt) => self.lower_for_stmt(for_stmt),
            ast::StmtKind::While(while_stmt) => self.lower_while_stmt(while_stmt),
            ast::StmtKind::Return(ret_stmt) => self.lower_return_stmt(ret_stmt),
            ast::StmtKind::Match(match_stmt) => self.lower_match_stmt(match_stmt),
            ast::StmtKind::Expr(expr_stmt) => {
                // Expression statement: evaluate for side effects.
                self.lower_expr(&expr_stmt.value);
            }
            ast::StmtKind::Try(try_stmt) => self.lower_try_stmt(try_stmt),
            ast::StmtKind::Break => {
                // Break is lowered as a Goto to the loop exit block.
                // This requires loop context tracking; for now emit Nop.
                self.emit(HirInst::Nop);
            }
            ast::StmtKind::Continue => {
                // Continue is lowered as a Goto to the loop header.
                self.emit(HirInst::Nop);
            }
            ast::StmtKind::Defer(defer_stmt) => {
                // Defer: collect the expression, emit before returns (LIFO).
                self.deferred.push((*defer_stmt.value).clone());
            }
            ast::StmtKind::ErrDefer(err_defer_stmt) => {
                // ErrDefer: for now treat same as defer.
                self.deferred.push((*err_defer_stmt.value).clone());
            }
            ast::StmtKind::Step(step_stmt) => {
                // Workflow steps: lower their body inline.
                for s in &step_stmt.body {
                    self.lower_stmt(s);
                }
            }
            ast::StmtKind::Assert(assert_stmt) => {
                // Lower assert as: evaluate condition, branch to panic or continue.
                let cond = self.lower_expr(&assert_stmt.condition);
                let pass_block = self.new_block();
                let fail_block = self.new_block();
                self.set_terminator(Terminator::Branch {
                    cond,
                    then_block: pass_block,
                    else_block: fail_block,
                });
                // Fail block: unreachable (runtime panic).
                self.switch_to_block(fail_block);
                self.set_terminator(Terminator::Unreachable);
                // Continue in pass block.
                self.switch_to_block(pass_block);
            }
        }
    }

    fn lower_let_stmt(&mut self, let_stmt: &ast::LetStmt) {
        let val = self.lower_expr(&let_stmt.value);
        let dst = self.fresh_var();
        self.emit(HirInst::Assign {
            dst,
            value: HirValue::Use(val),
        });
        self.vars.insert(let_stmt.name.node.clone(), dst);
    }

    fn lower_assign_stmt(&mut self, assign_stmt: &ast::AssignStmt) {
        let val = self.lower_expr(&assign_stmt.value);

        for target in &assign_stmt.targets {
            self.lower_assign_target(&target.path, val);
        }
    }

    fn lower_assign_target(&mut self, path: &ast::AssignPath, value: VarId) {
        match path {
            ast::AssignPath::Ident(name) => {
                let dst = if let Some(&existing) = self.vars.get(&name.node) {
                    existing
                } else {
                    let v = self.fresh_var();
                    self.vars.insert(name.node.clone(), v);
                    v
                };
                self.emit(HirInst::Assign {
                    dst,
                    value: HirValue::Use(value),
                });
            }
            ast::AssignPath::Field { object, field } => {
                let obj = self.lower_assign_path_to_var(object);
                self.emit(HirInst::SetField {
                    object: obj,
                    field: field.node.clone(),
                    value,
                });
            }
            ast::AssignPath::Index { object, index } => {
                let obj = self.lower_assign_path_to_var(object);
                let idx = self.lower_expr(index);
                self.emit(HirInst::SetIndex {
                    object: obj,
                    index: idx,
                    value,
                });
            }
        }
    }

    fn lower_assign_path_to_var(&mut self, path: &ast::AssignPath) -> VarId {
        match path {
            ast::AssignPath::Ident(name) => {
                self.vars.get(&name.node).copied().unwrap_or_else(|| {
                    let v = self.fresh_var();
                    self.vars.insert(name.node.clone(), v);
                    v
                })
            }
            ast::AssignPath::Field { object, field } => {
                let obj = self.lower_assign_path_to_var(object);
                let dst = self.fresh_var();
                self.emit(HirInst::GetField {
                    dst,
                    object: obj,
                    field: field.node.clone(),
                });
                dst
            }
            ast::AssignPath::Index { object, index } => {
                let obj = self.lower_assign_path_to_var(object);
                let idx = self.lower_expr(index);
                let dst = self.fresh_var();
                self.emit(HirInst::GetIndex {
                    dst,
                    object: obj,
                    index: idx,
                });
                dst
            }
        }
    }

    fn lower_if_stmt(&mut self, if_stmt: &ast::IfStmt) {
        let cond = self.lower_expr(&if_stmt.condition);

        let then_block = self.new_block();
        let else_block = self.new_block();
        let merge_block = self.new_block();

        self.set_terminator(Terminator::Branch {
            cond,
            then_block,
            else_block,
        });

        // Then branch.
        self.switch_to_block(then_block);
        self.lower_block_stmts(&if_stmt.then_branch);
        self.ensure_terminated(Terminator::Goto(merge_block));

        // Else branch.
        self.switch_to_block(else_block);
        match &if_stmt.else_branch {
            ast::ElseBranch::None => {
                self.set_terminator(Terminator::Goto(merge_block));
            }
            ast::ElseBranch::Block(block) => {
                self.lower_block_stmts(block);
                self.ensure_terminated(Terminator::Goto(merge_block));
            }
            ast::ElseBranch::ElseIf(elif) => {
                self.lower_if_stmt(&elif.node);
                self.ensure_terminated(Terminator::Goto(merge_block));
            }
        }

        // Continue after merge.
        self.switch_to_block(merge_block);
    }

    fn lower_for_stmt(&mut self, for_stmt: &ast::ForStmt) {
        // Evaluate the collection.
        let collection_var = self.lower_expr(&for_stmt.iterator);

        // Initialize index = 0.
        let index_var = self.fresh_var();
        self.emit(HirInst::Const {
            dst: index_var,
            value: HirConst::Int(0),
        });

        // Get collection length.
        let len_var = self.fresh_var();
        self.emit(HirInst::Call {
            dst: Some(len_var),
            func: FuncRef::Builtin("len".to_string()),
            args: vec![collection_var],
        });

        // Create loop structure: header -> body -> latch -> exit.
        let header_block = self.new_block();
        let body_block = self.new_block();
        let latch_block = self.new_block();
        let exit_block = self.new_block();

        self.set_terminator(Terminator::Goto(header_block));

        // Header: check index < len.
        self.switch_to_block(header_block);
        let cond_var = self.fresh_var();
        self.emit(HirInst::BinOp {
            dst: cond_var,
            op: BinOp::Lt,
            left: index_var,
            right: len_var,
        });
        self.set_terminator(Terminator::Branch {
            cond: cond_var,
            then_block: body_block,
            else_block: exit_block,
        });

        // Body: bind loop variable = collection[index], then run body stmts.
        self.switch_to_block(body_block);
        match &for_stmt.pattern {
            ast::ForPattern::Single(name) => {
                let elem_var = self.fresh_var();
                self.emit(HirInst::GetIndex {
                    dst: elem_var,
                    object: collection_var,
                    index: index_var,
                });
                self.vars.insert(name.node.clone(), elem_var);
            }
            ast::ForPattern::Pair(k, v_name) => {
                // For pairs (k, v), k is the index, v is the element.
                self.vars.insert(k.node.clone(), index_var);
                let elem_var = self.fresh_var();
                self.emit(HirInst::GetIndex {
                    dst: elem_var,
                    object: collection_var,
                    index: index_var,
                });
                self.vars.insert(v_name.node.clone(), elem_var);
            }
        }
        self.lower_block_stmts(&for_stmt.body);
        self.ensure_terminated(Terminator::Goto(latch_block));

        // Latch: index = index + 1, goto header.
        self.switch_to_block(latch_block);
        let one_var = self.fresh_var();
        self.emit(HirInst::Const {
            dst: one_var,
            value: HirConst::Int(1),
        });
        self.emit(HirInst::BinOp {
            dst: index_var,
            op: BinOp::Add,
            left: index_var,
            right: one_var,
        });
        self.set_terminator(Terminator::Goto(header_block));

        // Exit.
        self.switch_to_block(exit_block);
    }

    fn lower_while_stmt(&mut self, while_stmt: &ast::WhileStmt) {
        let header_block = self.new_block();
        let body_block = self.new_block();
        let exit_block = self.new_block();

        self.set_terminator(Terminator::Goto(header_block));

        // Header: evaluate condition and branch.
        self.switch_to_block(header_block);
        let cond = self.lower_expr(&while_stmt.condition);
        self.set_terminator(Terminator::Branch {
            cond,
            then_block: body_block,
            else_block: exit_block,
        });

        // Body.
        self.switch_to_block(body_block);
        self.lower_block_stmts(&while_stmt.body);
        self.ensure_terminated(Terminator::Goto(header_block));

        // Exit.
        self.switch_to_block(exit_block);
    }

    fn lower_return_stmt(&mut self, ret: &ast::ReturnStmt) {
        // Emit deferred calls before explicit return.
        self.emit_deferred();
        if ret.values.is_empty() {
            self.set_terminator(Terminator::Return(None));
        } else {
            // Lower the first return value (multi-return not yet supported in HIR).
            let val = self.lower_expr(&ret.values[0]);
            self.set_terminator(Terminator::Return(Some(val)));
        }
    }

    fn lower_match_stmt(&mut self, match_stmt: &ast::MatchStmt) {
        let scrutinee = self.lower_expr(&match_stmt.subject);
        let merge_block = self.new_block();

        if match_stmt.arms.is_empty() {
            self.set_terminator(Terminator::Goto(merge_block));
            self.switch_to_block(merge_block);
            return;
        }

        // Try to lower as a Switch (all literal patterns, no guards).
        let all_const = match_stmt.arms.iter().all(|arm| {
            arm.guard.is_none() && self.pattern_is_const(&arm.pattern.node)
        });

        if all_const {
            let mut cases = Vec::new();
            let mut default_block = merge_block;

            for arm in &match_stmt.arms {
                let arm_block = self.new_block();

                if let Some(c) = self.pattern_to_const(&arm.pattern.node) {
                    cases.push((c, arm_block));
                } else {
                    // Wildcard / ident -> default.
                    default_block = arm_block;
                }

                // Lower arm body.
                let saved_block = self.current_block;
                self.switch_to_block(arm_block);
                self.lower_match_arm_body(&arm.body);
                self.ensure_terminated(Terminator::Goto(merge_block));
                self.current_block = saved_block;
            }

            self.set_terminator(Terminator::Switch {
                scrutinee,
                cases,
                default: default_block,
            });
        } else {
            // Lower as a chain of if/else branches.
            self.lower_match_as_branches(scrutinee, &match_stmt.arms, merge_block);
        }

        self.switch_to_block(merge_block);
    }

    fn pattern_is_const(&self, pattern: &ast::Pattern) -> bool {
        matches!(
            pattern,
            ast::Pattern::Literal(_) | ast::Pattern::Wildcard | ast::Pattern::Ident(_)
        )
    }

    fn pattern_to_const(&self, pattern: &ast::Pattern) -> Option<HirConst> {
        match pattern {
            ast::Pattern::Literal(lit) => Some(self.lower_literal(lit)),
            // Wildcard and Ident are "default" arms, not const.
            _ => None,
        }
    }

    fn lower_match_arm_body(&mut self, body: &ast::MatchArmBody) {
        match body {
            ast::MatchArmBody::Expr(expr) => {
                self.lower_expr(expr);
            }
            ast::MatchArmBody::Block(block) => {
                self.lower_block_stmts(block);
            }
        }
    }

    fn lower_match_as_branches(
        &mut self,
        scrutinee: VarId,
        arms: &[ast::MatchArm],
        merge_block: BlockId,
    ) {
        if arms.is_empty() {
            self.set_terminator(Terminator::Goto(merge_block));
            return;
        }

        let arm = &arms[0];
        let arm_block = self.new_block();

        if arms.len() == 1 {
            // Last arm: go directly to it (acts as default).
            self.set_terminator(Terminator::Goto(arm_block));
            self.switch_to_block(arm_block);
            self.lower_match_arm_body(&arm.body);
            self.ensure_terminated(Terminator::Goto(merge_block));
            return;
        }

        let next_block = self.new_block();

        // Create a comparison for this arm's pattern.
        let cond = match &arm.pattern.node {
            ast::Pattern::Literal(lit) => {
                let lit_var = self.fresh_var();
                self.emit(HirInst::Const {
                    dst: lit_var,
                    value: self.lower_literal(lit),
                });
                let cmp = self.fresh_var();
                self.emit(HirInst::BinOp {
                    dst: cmp,
                    op: BinOp::Eq,
                    left: scrutinee,
                    right: lit_var,
                });
                cmp
            }
            _ => {
                // Wildcard/Ident: always true.
                let t = self.fresh_var();
                self.emit(HirInst::Const {
                    dst: t,
                    value: HirConst::Bool(true),
                });
                t
            }
        };

        self.set_terminator(Terminator::Branch {
            cond,
            then_block: arm_block,
            else_block: next_block,
        });

        // Arm body.
        self.switch_to_block(arm_block);
        self.lower_match_arm_body(&arm.body);
        self.ensure_terminated(Terminator::Goto(merge_block));

        // Continue with remaining arms.
        self.switch_to_block(next_block);
        self.lower_match_as_branches(scrutinee, &arms[1..], merge_block);
    }

    fn lower_try_stmt(&mut self, try_stmt: &ast::TryStmt) {
        // Simplified lowering: try body then catch body in sequence.
        // Real implementation would use setjmp/longjmp or Go-style defer/recover.
        let try_block = self.new_block();
        let catch_block = self.new_block();
        let merge_block = self.new_block();

        self.set_terminator(Terminator::Goto(try_block));

        self.switch_to_block(try_block);
        self.lower_block_stmts(&try_stmt.body);
        self.ensure_terminated(Terminator::Goto(merge_block));

        // Catch block: bind the error variable.
        self.switch_to_block(catch_block);
        let err_var = self.fresh_var();
        self.vars
            .insert(try_stmt.error_name.node.clone(), err_var);
        self.lower_block_stmts(&try_stmt.catch_body);
        self.ensure_terminated(Terminator::Goto(merge_block));

        self.switch_to_block(merge_block);
    }

    // -----------------------------------------------------------------------
    // Expression lowering
    // -----------------------------------------------------------------------

    /// Lower an expression, producing a VarId that holds the result.
    fn lower_expr(&mut self, expr: &ast::Expr) -> VarId {
        match &expr.node {
            ast::ExprKind::Literal(lit) => {
                // Interpolated strings need special handling — they aren't constants.
                if let ast::Literal::InterpolatedString(parts) = lit {
                    return self.lower_interpolated_string(parts);
                }
                let dst = self.fresh_var();
                self.emit(HirInst::Const {
                    dst,
                    value: self.lower_literal(lit),
                });
                dst
            }
            ast::ExprKind::Ident(name) => {
                if let Some(&v) = self.vars.get(name) {
                    v
                } else {
                    // Unknown variable: allocate a fresh var and record it.
                    let dst = self.fresh_var();
                    self.vars.insert(name.clone(), dst);
                    dst
                }
            }
            ast::ExprKind::Binary(bin) => {
                let left = self.lower_expr(&bin.left);
                let right = self.lower_expr(&bin.right);
                let dst = self.fresh_var();
                let op = lower_binop(bin.op.node);
                self.emit(HirInst::BinOp {
                    dst,
                    op,
                    left,
                    right,
                });
                dst
            }
            ast::ExprKind::Unary(un) => {
                let operand = self.lower_expr(&un.operand);
                let dst = self.fresh_var();
                let op = lower_unop(un.op.node);
                self.emit(HirInst::UnOp { dst, op, operand });
                dst
            }
            ast::ExprKind::Call(call) => self.lower_call_expr(call),
            ast::ExprKind::MethodCall(mc) => {
                let args: Vec<VarId> = mc.args.iter().map(|a| self.lower_expr(&a.value)).collect();
                let dst = self.fresh_var();
                // If the receiver is a simple ident, treat as qualified call (e.g. io.println).
                if let ast::ExprKind::Ident(module_name) = &mc.receiver.node {
                    let qualified = format!("{}.{}", module_name, mc.method.node);
                    let func_ref = self.resolve_func_ref(&qualified);
                    self.emit(HirInst::Call {
                        dst: Some(dst),
                        func: func_ref,
                        args,
                    });
                } else {
                    let receiver = self.lower_expr(&mc.receiver);
                    self.emit(HirInst::MethodCall {
                        dst: Some(dst),
                        receiver,
                        method: mc.method.node.clone(),
                        args,
                    });
                }
                dst
            }
            ast::ExprKind::Field(field) => {
                let object = self.lower_expr(&field.object);
                let dst = self.fresh_var();
                self.emit(HirInst::GetField {
                    dst,
                    object,
                    field: field.field.node.clone(),
                });
                dst
            }
            ast::ExprKind::Index(idx) => {
                let object = self.lower_expr(&idx.object);
                let index = self.lower_expr(&idx.index);
                let dst = self.fresh_var();
                self.emit(HirInst::GetIndex {
                    dst,
                    object,
                    index,
                });
                dst
            }
            ast::ExprKind::Pipe(pipe) => {
                // `a |> f` desugars to `f(a)`.
                let left = self.lower_expr(&pipe.left);
                let dst = self.fresh_var();
                // The right side should be a callable; lower it and call with left as arg.
                match &pipe.right.node {
                    ast::ExprKind::Ident(name) => {
                        let func_ref = self.resolve_func_ref(name);
                        self.emit(HirInst::Call {
                            dst: Some(dst),
                            func: func_ref,
                            args: vec![left],
                        });
                    }
                    _ => {
                        // General case: lower right, call it.
                        let func_var = self.lower_expr(&pipe.right);
                        self.emit(HirInst::Call {
                            dst: Some(dst),
                            func: FuncRef::Builtin("__call_indirect".to_string()),
                            args: vec![func_var, left],
                        });
                    }
                }
                dst
            }
            ast::ExprKind::Lambda(lambda) => {
                // Create a new function for the lambda.
                let params: Vec<HirParam> =
                    lambda.params.iter().map(|p| self.lower_param(p)).collect();
                let saved_func = self.current_func;
                let saved_block = self.current_block;
                let saved_vars = self.vars.clone();

                let func_id = self.begin_function(
                    format!("__lambda_{}", self.module.functions.len()),
                    params.clone(),
                    HirType::Any, // inferred later
                    false,
                    false,
                    false,
                    expr.span,
                );

                for p in &params {
                    let vid = self.fresh_var();
                    self.vars.insert(p.name.clone(), vid);
                }

                match &lambda.body {
                    ast::LambdaBody::Expr(e) => {
                        let val = self.lower_expr(e);
                        self.set_terminator(Terminator::Return(Some(val)));
                    }
                    ast::LambdaBody::Block(block) => {
                        self.lower_block_stmts(block);
                        self.ensure_terminated(Terminator::Return(None));
                    }
                }

                // Restore context.
                self.current_func = saved_func;
                self.current_block = saved_block;
                self.vars = saved_vars;

                // Return a variable holding a reference to the lambda function.
                let dst = self.fresh_var();
                self.emit(HirInst::Const {
                    dst,
                    value: HirConst::Int(func_id as i64),
                });
                dst
            }
            ast::ExprKind::Match(match_expr) => {
                // Lower match expression by evaluating into a result variable.
                let scrutinee = self.lower_expr(&match_expr.subject);
                let result = self.fresh_var();
                let merge_block = self.new_block();

                // Use branch chain approach.
                self.lower_match_expr_arms(scrutinee, result, &match_expr.arms, merge_block);

                self.switch_to_block(merge_block);
                result
            }
            ast::ExprKind::If(if_expr) => {
                // If expression: both branches produce a value.
                let cond = self.lower_expr(&if_expr.if_stmt.condition);
                let result = self.fresh_var();
                let then_block = self.new_block();
                let else_block = self.new_block();
                let merge_block = self.new_block();

                self.set_terminator(Terminator::Branch {
                    cond,
                    then_block,
                    else_block,
                });

                // Then.
                self.switch_to_block(then_block);
                self.lower_block_stmts(&if_expr.if_stmt.then_branch);
                self.ensure_terminated(Terminator::Goto(merge_block));

                // Else.
                self.switch_to_block(else_block);
                match &if_expr.if_stmt.else_branch {
                    ast::ElseBranch::None => {}
                    ast::ElseBranch::Block(block) => {
                        self.lower_block_stmts(block);
                    }
                    ast::ElseBranch::ElseIf(elif) => {
                        self.lower_if_stmt(&elif.node);
                    }
                }
                self.ensure_terminated(Terminator::Goto(merge_block));

                self.switch_to_block(merge_block);
                result
            }
            ast::ExprKind::Block(block_expr) => {
                self.lower_block_stmts(&block_expr.body);
                // Block expression evaluates to the last expression if any.
                // For now, return a dummy var.
                self.fresh_var()
            }
            ast::ExprKind::List(list) => {
                let elems: Vec<VarId> =
                    list.elems.iter().map(|e| self.lower_expr(e)).collect();
                let dst = self.fresh_var();
                self.emit(HirInst::ConstructList { dst, elems });
                dst
            }
            ast::ExprKind::Map(map) => {
                let entries: Vec<(VarId, VarId)> = map
                    .entries
                    .iter()
                    .map(|e| (self.lower_expr(&e.key), self.lower_expr(&e.value)))
                    .collect();
                let dst = self.fresh_var();
                self.emit(HirInst::ConstructMap { dst, entries });
                dst
            }
            ast::ExprKind::Instance(inst) => {
                let fields: Vec<(String, VarId)> = inst
                    .fields
                    .iter()
                    .map(|f| {
                        let name = f
                            .name
                            .as_ref()
                            .map(|n| n.node.clone())
                            .unwrap_or_default();
                        let val = self.lower_expr(&f.value);
                        (name, val)
                    })
                    .collect();
                let dst = self.fresh_var();
                self.emit(HirInst::ConstructStruct {
                    dst,
                    type_name: inst.type_name.node.clone(),
                    fields,
                });
                dst
            }
            ast::ExprKind::Range(_range) => {
                // Ranges lower to a builtin call or special construct.
                let dst = self.fresh_var();
                self.emit(HirInst::Nop);
                dst
            }
            ast::ExprKind::Propagate(prop) => {
                let inner = self.lower_expr(&prop.inner);
                let dst = self.fresh_var();
                self.emit(HirInst::Propagate { dst, inner });
                dst
            }
            ast::ExprKind::Orelse(orelse) => {
                // `a ?? b` desugars to: if a is none then b else a.
                let left = self.lower_expr(&orelse.left);
                let right = self.lower_expr(&orelse.default);
                // Simplified: just return left (full impl needs branch).
                let dst = self.fresh_var();
                self.emit(HirInst::Assign {
                    dst,
                    value: HirValue::Use(left),
                });
                let _ = right;
                dst
            }
            ast::ExprKind::Some(some) => {
                self.lower_expr(&some.inner)
            }
            ast::ExprKind::None => {
                let dst = self.fresh_var();
                self.emit(HirInst::Const {
                    dst,
                    value: HirConst::None,
                });
                dst
            }
            ast::ExprKind::Async(async_expr) => {
                // Async blocks become spawned tasks.
                let body_block = self.new_block();
                let saved = self.current_block;

                self.switch_to_block(body_block);
                self.lower_block_stmts(&async_expr.body);
                self.ensure_terminated(Terminator::Return(None));

                self.switch_to_block(saved);
                let dst = self.fresh_var();
                self.emit(HirInst::Spawn {
                    dst,
                    block: body_block,
                });
                dst
            }
            ast::ExprKind::Spawn(spawn) => {
                let body_block = self.new_block();
                let saved = self.current_block;

                self.switch_to_block(body_block);
                self.lower_block_stmts(&spawn.body);
                self.ensure_terminated(Terminator::Return(None));

                self.switch_to_block(saved);
                let dst = self.fresh_var();
                self.emit(HirInst::Spawn {
                    dst,
                    block: body_block,
                });
                dst
            }
            ast::ExprKind::Select(_select) => {
                // Select is complex concurrency; placeholder for now.
                let dst = self.fresh_var();
                self.emit(HirInst::Nop);
                dst
            }
            ast::ExprKind::Paren(paren) => self.lower_expr(&paren.inner),
        }
    }

    fn lower_call_expr(&mut self, call: &ast::CallExpr) -> VarId {
        let args: Vec<VarId> = call
            .args
            .iter()
            .map(|a| self.lower_expr(&a.value))
            .collect();
        let dst = self.fresh_var();

        // Determine the function reference from the callee expression.
        match &call.callee.node {
            ast::ExprKind::Ident(name) => {
                let func_ref = self.resolve_func_ref(name);
                self.emit(HirInst::Call {
                    dst: Some(dst),
                    func: func_ref,
                    args,
                });
            }
            ast::ExprKind::Field(field) => {
                // `module.func(args)` or `obj.method(args)`.
                // If the object is a simple ident, treat as qualified name.
                if let ast::ExprKind::Ident(module_name) = &field.object.node {
                    let qualified = format!("{}.{}", module_name, field.field.node);
                    let func_ref = self.resolve_func_ref(&qualified);
                    self.emit(HirInst::Call {
                        dst: Some(dst),
                        func: func_ref,
                        args,
                    });
                } else {
                    // General method call.
                    let receiver = self.lower_expr(&field.object);
                    self.emit(HirInst::MethodCall {
                        dst: Some(dst),
                        receiver,
                        method: field.field.node.clone(),
                        args,
                    });
                }
            }
            _ => {
                // Indirect call.
                let callee = self.lower_expr(&call.callee);
                let mut all_args = vec![callee];
                all_args.extend(args);
                self.emit(HirInst::Call {
                    dst: Some(dst),
                    func: FuncRef::Builtin("__call_indirect".to_string()),
                    args: all_args,
                });
            }
        }

        dst
    }

    fn lower_match_expr_arms(
        &mut self,
        scrutinee: VarId,
        result: VarId,
        arms: &[ast::MatchArm],
        merge_block: BlockId,
    ) {
        if arms.is_empty() {
            self.set_terminator(Terminator::Goto(merge_block));
            return;
        }

        let arm = &arms[0];
        let arm_block = self.new_block();

        if arms.len() == 1 {
            self.set_terminator(Terminator::Goto(arm_block));
            self.switch_to_block(arm_block);
            let val = self.lower_match_arm_body_as_expr(&arm.body);
            self.emit(HirInst::Assign {
                dst: result,
                value: HirValue::Use(val),
            });
            self.ensure_terminated(Terminator::Goto(merge_block));
            return;
        }

        let next_block = self.new_block();

        let cond = match &arm.pattern.node {
            ast::Pattern::Literal(lit) => {
                let lit_var = self.fresh_var();
                self.emit(HirInst::Const {
                    dst: lit_var,
                    value: self.lower_literal(lit),
                });
                let cmp = self.fresh_var();
                self.emit(HirInst::BinOp {
                    dst: cmp,
                    op: BinOp::Eq,
                    left: scrutinee,
                    right: lit_var,
                });
                cmp
            }
            _ => {
                let t = self.fresh_var();
                self.emit(HirInst::Const {
                    dst: t,
                    value: HirConst::Bool(true),
                });
                t
            }
        };

        self.set_terminator(Terminator::Branch {
            cond,
            then_block: arm_block,
            else_block: next_block,
        });

        self.switch_to_block(arm_block);
        let val = self.lower_match_arm_body_as_expr(&arm.body);
        self.emit(HirInst::Assign {
            dst: result,
            value: HirValue::Use(val),
        });
        self.ensure_terminated(Terminator::Goto(merge_block));

        self.switch_to_block(next_block);
        self.lower_match_expr_arms(scrutinee, result, &arms[1..], merge_block);
    }

    fn lower_match_arm_body_as_expr(&mut self, body: &ast::MatchArmBody) -> VarId {
        match body {
            ast::MatchArmBody::Expr(expr) => self.lower_expr(expr),
            ast::MatchArmBody::Block(block) => {
                self.lower_block_stmts(block);
                self.fresh_var() // placeholder
            }
        }
    }

    // -----------------------------------------------------------------------
    // Utilities
    // -----------------------------------------------------------------------

    /// Resolve a function name to a FuncRef. Checks local functions first,
    /// then falls back to Builtin.
    fn resolve_func_ref(&self, name: &str) -> FuncRef {
        if let Some(&id) = self.func_names.get(name) {
            FuncRef::Local(id)
        } else {
            FuncRef::Builtin(name.to_string())
        }
    }

    /// If the current block does not yet have a terminator (still Unreachable
    /// from initial creation), set it to the given terminator.
    fn ensure_terminated(&mut self, term: Terminator) {
        let func = &self.module.functions[self.current_func];
        if matches!(func.blocks[self.current_block].terminator, Terminator::Unreachable) {
            // Emit deferred calls before implicit returns.
            if matches!(term, Terminator::Return(_)) {
                self.emit_deferred();
            }
            self.set_terminator(term);
        }
    }

    /// Emit all deferred expressions in LIFO order (last defer runs first).
    fn emit_deferred(&mut self) {
        // Clone to avoid borrow conflict — deferred exprs are AST nodes.
        let deferred = self.deferred.clone();
        for expr in deferred.iter().rev() {
            self.lower_expr(expr);
        }
    }
}

// ===========================================================================
// Operator mapping
// ===========================================================================

fn lower_binop(op: ast::BinaryOp) -> BinOp {
    match op {
        ast::BinaryOp::Add => BinOp::Add,
        ast::BinaryOp::Sub => BinOp::Sub,
        ast::BinaryOp::Mul => BinOp::Mul,
        ast::BinaryOp::Div => BinOp::Div,
        ast::BinaryOp::Mod => BinOp::Mod,
        ast::BinaryOp::Eq => BinOp::Eq,
        ast::BinaryOp::Ne => BinOp::Ne,
        ast::BinaryOp::Lt => BinOp::Lt,
        ast::BinaryOp::Gt => BinOp::Gt,
        ast::BinaryOp::Le => BinOp::Le,
        ast::BinaryOp::Ge => BinOp::Ge,
        ast::BinaryOp::And => BinOp::And,
        ast::BinaryOp::Or => BinOp::Or,
        ast::BinaryOp::BitAnd => BinOp::BitAnd,
        ast::BinaryOp::BitOr => BinOp::BitOr,
        ast::BinaryOp::BitXor => BinOp::BitXor,
        ast::BinaryOp::Shl => BinOp::Shl,
        ast::BinaryOp::Shr => BinOp::Shr,
    }
}

fn lower_unop(op: ast::UnaryOp) -> UnOp {
    match op {
        ast::UnaryOp::Neg => UnOp::Neg,
        ast::UnaryOp::Not => UnOp::Not,
        ast::UnaryOp::BitNot => UnOp::BitNot,
    }
}

// ===========================================================================
// Tests
// ===========================================================================

#[cfg(test)]
mod tests {
    use super::*;
    use haira_ast::*;
    use haira_errors::Span;

    /// Helper: create a Span at position 0.
    fn sp() -> Span {
        Span::new(0, 0)
    }

    /// Helper: wrap a value in a Spanned with a dummy span.
    fn spanned<T>(node: T) -> Spanned<T> {
        Spanned::new(node, sp())
    }

    /// Helper: create an empty SourceFile.
    fn empty_source() -> SourceFile {
        SourceFile {
            items: Vec::new(),
            span: sp(),
        }
    }

    // -----------------------------------------------------------------------
    // Test 1: Lower an empty source file
    // -----------------------------------------------------------------------

    #[test]
    fn lower_empty_source_file() {
        let src = empty_source();
        let (module, diags) = lower(&src);

        assert!(diags.is_empty(), "no diagnostics for empty file");
        assert!(module.functions.is_empty(), "no functions for empty file");
        assert!(module.declarations.is_empty(), "no declarations for empty file");
        assert!(module.entry.is_none(), "no entry point for empty file");
    }

    // -----------------------------------------------------------------------
    // Test 2: Lower a function with a return statement
    // -----------------------------------------------------------------------

    #[test]
    fn lower_function_with_return() {
        let src = SourceFile {
            items: vec![spanned(ItemKind::FunctionDef(FunctionDef {
                is_public: false,
                name: spanned("add".to_string()),
                params: vec![
                    Param {
                        name: spanned("a".to_string()),
                        ty: Some(spanned(Type::Named("int".to_string()))),
                        default: None,
                        is_rest: false,
                        span: sp(),
                    },
                    Param {
                        name: spanned("b".to_string()),
                        ty: Some(spanned(Type::Named("int".to_string()))),
                        default: None,
                        is_rest: false,
                        span: sp(),
                    },
                ],
                return_ty: Some(spanned(Type::Named("int".to_string()))),
                body: Block {
                    statements: vec![spanned(StmtKind::Return(ReturnStmt {
                        values: vec![spanned(ExprKind::Binary(BinaryExpr {
                            left: Box::new(spanned(ExprKind::Ident("a".to_string()))),
                            op: spanned(BinaryOp::Add),
                            right: Box::new(spanned(ExprKind::Ident("b".to_string()))),
                        }))],
                    }))],
                    span: sp(),
                },
            }))],
            span: sp(),
        };

        let (module, diags) = lower(&src);
        assert!(diags.is_empty());
        assert_eq!(module.functions.len(), 1);

        let func = &module.functions[0];
        assert_eq!(func.name, "add");
        assert_eq!(func.params.len(), 2);
        assert_eq!(func.return_ty, HirType::Int);
        assert!(!func.blocks.is_empty());

        // The entry block should have a BinOp instruction and Return terminator.
        let entry = &func.blocks[0];
        let has_binop = entry.insts.iter().any(|inst| matches!(inst, HirInst::BinOp { op: BinOp::Add, .. }));
        assert!(has_binop, "entry block should contain Add binop");
        assert!(matches!(entry.terminator, Terminator::Return(Some(_))));
    }

    // -----------------------------------------------------------------------
    // Test 3: Lower an if/else into branch blocks
    // -----------------------------------------------------------------------

    #[test]
    fn lower_if_else_into_branches() {
        let src = SourceFile {
            items: vec![spanned(ItemKind::FunctionDef(FunctionDef {
                is_public: false,
                name: spanned("check".to_string()),
                params: vec![Param {
                    name: spanned("x".to_string()),
                    ty: Some(spanned(Type::Named("bool".to_string()))),
                    default: None,
                    is_rest: false,
                    span: sp(),
                }],
                return_ty: Some(spanned(Type::Named("int".to_string()))),
                body: Block {
                    statements: vec![spanned(StmtKind::If(IfStmt {
                        condition: Box::new(spanned(ExprKind::Ident("x".to_string()))),
                        then_branch: Block {
                            statements: vec![spanned(StmtKind::Return(ReturnStmt {
                                values: vec![spanned(ExprKind::Literal(Literal::Int(1)))],
                            }))],
                            span: sp(),
                        },
                        else_branch: ElseBranch::Block(Block {
                            statements: vec![spanned(StmtKind::Return(ReturnStmt {
                                values: vec![spanned(ExprKind::Literal(Literal::Int(0)))],
                            }))],
                            span: sp(),
                        }),
                    }))],
                    span: sp(),
                },
            }))],
            span: sp(),
        };

        let (module, diags) = lower(&src);
        assert!(diags.is_empty());
        assert_eq!(module.functions.len(), 1);

        let func = &module.functions[0];
        // Should have at least 4 blocks: entry, then, else, merge.
        assert!(
            func.blocks.len() >= 4,
            "if/else should produce at least 4 blocks, got {}",
            func.blocks.len()
        );

        // Entry block should have a Branch terminator.
        let entry = &func.blocks[0];
        assert!(
            matches!(entry.terminator, Terminator::Branch { .. }),
            "entry block should end with Branch terminator"
        );
    }

    // -----------------------------------------------------------------------
    // Test 4: Lower a let binding
    // -----------------------------------------------------------------------

    #[test]
    fn lower_let_binding() {
        let src = SourceFile {
            items: vec![spanned(ItemKind::FunctionDef(FunctionDef {
                is_public: false,
                name: spanned("f".to_string()),
                params: Vec::new(),
                return_ty: None,
                body: Block {
                    statements: vec![spanned(StmtKind::Let(LetStmt {
                        name: spanned("x".to_string()),
                        type_ann: Some(spanned(Type::Named("int".to_string()))),
                        value: Box::new(spanned(ExprKind::Literal(Literal::Int(42)))),
                        is_const: false,
                    }))],
                    span: sp(),
                },
            }))],
            span: sp(),
        };

        let (module, diags) = lower(&src);
        assert!(diags.is_empty());
        assert_eq!(module.functions.len(), 1);

        let func = &module.functions[0];
        let entry = &func.blocks[0];

        // Should have a Const instruction (loading 42) followed by an Assign.
        let has_const_42 = entry.insts.iter().any(|inst| {
            matches!(inst, HirInst::Const { value: HirConst::Int(42), .. })
        });
        assert!(has_const_42, "should contain Const(42) instruction");

        let has_assign = entry.insts.iter().any(|inst| matches!(inst, HirInst::Assign { .. }));
        assert!(has_assign, "should contain Assign instruction");
    }

    // -----------------------------------------------------------------------
    // Test 5: Lower a binary expression
    // -----------------------------------------------------------------------

    #[test]
    fn lower_binary_expression() {
        let src = SourceFile {
            items: vec![spanned(ItemKind::FunctionDef(FunctionDef {
                is_public: false,
                name: spanned("calc".to_string()),
                params: Vec::new(),
                return_ty: None,
                body: Block {
                    statements: vec![spanned(StmtKind::Let(LetStmt {
                        name: spanned("result".to_string()),
                        type_ann: None,
                        value: Box::new(spanned(ExprKind::Binary(BinaryExpr {
                            left: Box::new(spanned(ExprKind::Literal(Literal::Int(10)))),
                            op: spanned(BinaryOp::Mul),
                            right: Box::new(spanned(ExprKind::Literal(Literal::Int(5)))),
                        }))),
                        is_const: false,
                    }))],
                    span: sp(),
                },
            }))],
            span: sp(),
        };

        let (module, diags) = lower(&src);
        assert!(diags.is_empty());

        let func = &module.functions[0];
        let entry = &func.blocks[0];

        // Should produce: Const(10), Const(5), BinOp(Mul), Assign.
        let has_mul = entry.insts.iter().any(|inst| {
            matches!(inst, HirInst::BinOp { op: BinOp::Mul, .. })
        });
        assert!(has_mul, "should contain Mul binop instruction");

        // At least 4 instructions: two consts, one binop, one assign.
        assert!(
            entry.insts.len() >= 4,
            "expected at least 4 instructions, got {}",
            entry.insts.len()
        );
    }

    // -----------------------------------------------------------------------
    // Test 6: Lower a function call
    // -----------------------------------------------------------------------

    #[test]
    fn lower_function_call() {
        let src = SourceFile {
            items: vec![spanned(ItemKind::FunctionDef(FunctionDef {
                is_public: false,
                name: spanned("main".to_string()),
                params: Vec::new(),
                return_ty: None,
                body: Block {
                    statements: vec![spanned(StmtKind::Expr(ExprStmt {
                        value: Box::new(spanned(ExprKind::Call(CallExpr {
                            callee: Box::new(spanned(ExprKind::Ident(
                                "println".to_string(),
                            ))),
                            args: vec![Argument {
                                name: None,
                                value: spanned(ExprKind::Literal(Literal::String(
                                    "hello".to_string(),
                                ))),
                                span: sp(),
                            }],
                        }))),
                    }))],
                    span: sp(),
                },
            }))],
            span: sp(),
        };

        let (module, diags) = lower(&src);
        assert!(diags.is_empty());
        assert_eq!(module.functions.len(), 1);

        let func = &module.functions[0];
        let entry = &func.blocks[0];

        // Should have a Const (string "hello") and a Call instruction.
        let has_call = entry.insts.iter().any(|inst| {
            matches!(inst, HirInst::Call { func: FuncRef::Builtin(name), .. } if name == "println")
        });
        assert!(has_call, "should contain Call to println");

        let has_str_const = entry.insts.iter().any(|inst| {
            matches!(inst, HirInst::Const { value: HirConst::Str(s), .. } if s == "hello")
        });
        assert!(has_str_const, "should contain Const(\"hello\") instruction");
    }

    // -----------------------------------------------------------------------
    // Additional tests
    // -----------------------------------------------------------------------

    #[test]
    fn lower_struct_declaration() {
        let src = SourceFile {
            items: vec![spanned(ItemKind::TypeDef(TypeDef {
                is_public: true,
                name: spanned("User".to_string()),
                fields: vec![
                    Field {
                        name: spanned("name".to_string()),
                        ty: Some(spanned(Type::Named("string".to_string()))),
                        default: None,
                        span: sp(),
                    },
                    Field {
                        name: spanned("age".to_string()),
                        ty: Some(spanned(Type::Named("int".to_string()))),
                        default: None,
                        span: sp(),
                    },
                ],
            }))],
            span: sp(),
        };

        let (module, diags) = lower(&src);
        assert!(diags.is_empty());
        assert_eq!(module.declarations.len(), 1);

        match &module.declarations[0] {
            HirDecl::Struct { name, fields } => {
                assert_eq!(name, "User");
                assert_eq!(fields.len(), 2);
                assert_eq!(fields[0], ("name".to_string(), HirType::Str));
                assert_eq!(fields[1], ("age".to_string(), HirType::Int));
            }
            _ => panic!("expected Struct declaration"),
        }
    }

    #[test]
    fn lower_while_loop() {
        let src = SourceFile {
            items: vec![spanned(ItemKind::FunctionDef(FunctionDef {
                is_public: false,
                name: spanned("loop_fn".to_string()),
                params: Vec::new(),
                return_ty: None,
                body: Block {
                    statements: vec![spanned(StmtKind::While(WhileStmt {
                        condition: Box::new(spanned(ExprKind::Literal(Literal::Bool(true)))),
                        body: Block {
                            statements: vec![spanned(StmtKind::Expr(ExprStmt {
                                value: Box::new(spanned(ExprKind::Literal(Literal::Int(1)))),
                            }))],
                            span: sp(),
                        },
                    }))],
                    span: sp(),
                },
            }))],
            span: sp(),
        };

        let (module, diags) = lower(&src);
        assert!(diags.is_empty());

        let func = &module.functions[0];
        // While loop should create: entry -> header -> body -> exit.
        assert!(
            func.blocks.len() >= 4,
            "while loop should produce at least 4 blocks, got {}",
            func.blocks.len()
        );

        // Entry block should Goto the header.
        assert!(matches!(func.blocks[0].terminator, Terminator::Goto(_)));
    }

    #[test]
    fn lower_top_level_statements_create_entry() {
        let src = SourceFile {
            items: vec![spanned(ItemKind::ItemStatement(spanned(
                StmtKind::Expr(ExprStmt {
                    value: Box::new(spanned(ExprKind::Literal(Literal::Int(42)))),
                }),
            )))],
            span: sp(),
        };

        let (module, diags) = lower(&src);
        assert!(diags.is_empty());
        assert!(module.entry.is_some(), "should have entry function");
        assert_eq!(
            module.functions[module.entry.unwrap()].name,
            "__haira_main"
        );
    }

    #[test]
    fn lower_enum_declaration() {
        let src = SourceFile {
            items: vec![spanned(ItemKind::EnumDef(EnumDef {
                is_public: false,
                name: spanned("Color".to_string()),
                variants: vec![
                    EnumVariant {
                        name: spanned("Red".to_string()),
                        fields: Vec::new(),
                        span: sp(),
                    },
                    EnumVariant {
                        name: spanned("Green".to_string()),
                        fields: Vec::new(),
                        span: sp(),
                    },
                    EnumVariant {
                        name: spanned("Blue".to_string()),
                        fields: Vec::new(),
                        span: sp(),
                    },
                ],
            }))],
            span: sp(),
        };

        let (module, diags) = lower(&src);
        assert!(diags.is_empty());
        assert_eq!(module.declarations.len(), 1);

        match &module.declarations[0] {
            HirDecl::Enum { name, variants } => {
                assert_eq!(name, "Color");
                assert_eq!(variants.len(), 3);
                assert_eq!(variants[0].0, "Red");
                assert_eq!(variants[1].0, "Green");
                assert_eq!(variants[2].0, "Blue");
            }
            _ => panic!("expected Enum declaration"),
        }
    }

    #[test]
    fn lower_provider_declaration() {
        let src = SourceFile {
            items: vec![spanned(ItemKind::ProviderDecl(ProviderDecl {
                name: spanned("openai".to_string()),
                fields: vec![ProviderField {
                    key: spanned("model".to_string()),
                    value: spanned(ExprKind::Literal(Literal::String(
                        "gpt-4".to_string(),
                    ))),
                }],
            }))],
            span: sp(),
        };

        let (module, diags) = lower(&src);
        assert!(diags.is_empty());

        match &module.declarations[0] {
            HirDecl::Provider { name, fields } => {
                assert_eq!(name, "openai");
                assert_eq!(fields[0], ("model".to_string(), HirConst::Str("gpt-4".to_string())));
            }
            _ => panic!("expected Provider declaration"),
        }
    }

    #[test]
    fn lower_method_def_prepends_self() {
        let src = SourceFile {
            items: vec![spanned(ItemKind::MethodDef(MethodDef {
                type_name: spanned("User".to_string()),
                name: spanned("greet".to_string()),
                params: Vec::new(),
                return_ty: Some(spanned(Type::Named("string".to_string()))),
                body: Block {
                    statements: vec![spanned(StmtKind::Return(ReturnStmt {
                        values: vec![spanned(ExprKind::Literal(Literal::String(
                            "hi".to_string(),
                        )))],
                    }))],
                    span: sp(),
                },
            }))],
            span: sp(),
        };

        let (module, diags) = lower(&src);
        assert!(diags.is_empty());

        let func = &module.functions[0];
        assert_eq!(func.name, "User.greet");
        assert_eq!(func.params.len(), 1, "should have self param");
        assert_eq!(func.params[0].name, "self");
        assert_eq!(func.params[0].ty, HirType::Struct("User".to_string()));
    }

    #[test]
    fn lower_list_literal() {
        let src = SourceFile {
            items: vec![spanned(ItemKind::FunctionDef(FunctionDef {
                is_public: false,
                name: spanned("f".to_string()),
                params: Vec::new(),
                return_ty: None,
                body: Block {
                    statements: vec![spanned(StmtKind::Let(LetStmt {
                        name: spanned("xs".to_string()),
                        type_ann: None,
                        value: Box::new(spanned(ExprKind::List(ListExpr {
                            elems: vec![
                                spanned(ExprKind::Literal(Literal::Int(1))),
                                spanned(ExprKind::Literal(Literal::Int(2))),
                                spanned(ExprKind::Literal(Literal::Int(3))),
                            ],
                        }))),
                        is_const: false,
                    }))],
                    span: sp(),
                },
            }))],
            span: sp(),
        };

        let (module, diags) = lower(&src);
        assert!(diags.is_empty());

        let func = &module.functions[0];
        let entry = &func.blocks[0];

        let has_construct_list = entry.insts.iter().any(|inst| {
            matches!(inst, HirInst::ConstructList { elems, .. } if elems.len() == 3)
        });
        assert!(has_construct_list, "should contain ConstructList with 3 elements");
    }

    #[test]
    fn lower_type_maps_correctly() {
        let mut ctx = LowerCtx::new();

        assert_eq!(ctx.lower_type(&Type::Named("void".to_string())), HirType::Void);
        assert_eq!(ctx.lower_type(&Type::Named("bool".to_string())), HirType::Bool);
        assert_eq!(ctx.lower_type(&Type::Named("int".to_string())), HirType::Int);
        assert_eq!(ctx.lower_type(&Type::Named("float".to_string())), HirType::Float);
        assert_eq!(ctx.lower_type(&Type::Named("string".to_string())), HirType::Str);
        assert_eq!(ctx.lower_type(&Type::Named("any".to_string())), HirType::Any);
        assert_eq!(ctx.lower_type(&Type::Named("stream".to_string())), HirType::Stream);
        assert_eq!(
            ctx.lower_type(&Type::Named("MyStruct".to_string())),
            HirType::Struct("MyStruct".to_string())
        );
        assert_eq!(
            ctx.lower_type(&Type::List(Box::new(spanned(Type::Named("int".to_string()))))),
            HirType::List(Box::new(HirType::Int))
        );
        assert_eq!(
            ctx.lower_type(&Type::Option(Box::new(spanned(Type::Named("string".to_string()))))),
            HirType::Option(Box::new(HirType::Str))
        );
    }
}
