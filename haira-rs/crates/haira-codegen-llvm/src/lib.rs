//! Haira LLVM codegen backend.
//!
//! Generates LLVM IR text from HIR, then invokes `clang` to compile to a native binary.
//! Uses alloca-based variable storage — LLVM's mem2reg pass optimizes it to SSA.

use std::collections::{HashMap, HashSet};
use std::fmt::Write;
use std::process::Command;

use haira_errors::{Diagnostic, Span};
use haira_ir::*;

// ===========================================================================
// Public API
// ===========================================================================

/// Result of LLVM code generation.
#[derive(Debug)]
pub enum CodegenResult {
    /// Compiled binary at the given path.
    Binary { path: String },
}

/// Code generation options.
#[derive(Debug, Clone)]
pub struct CodegenOptions {
    pub optimize: bool,
    pub output: String,
    pub emit_ir: bool,
}

/// Generate a native binary from a HIR module via LLVM IR + clang.
pub fn codegen(
    module: &HirModule,
    options: &CodegenOptions,
) -> Result<CodegenResult, Vec<Diagnostic>> {
    let mut cg = LlvmGen::new();
    cg.generate(module)?;
    let ir = cg.finish();

    let ir_path = format!("{}.ll", &options.output);
    std::fs::write(&ir_path, &ir).map_err(|e| {
        vec![Diagnostic::error(
            format!("cannot write LLVM IR: {e}"),
            Span::default(),
        )]
    })?;

    if options.emit_ir {
        print!("{ir}");
    }

    // Invoke clang to compile LLVM IR → native binary
    let opt = if options.optimize { "-O2" } else { "-O0" };
    let output = Command::new("clang")
        .args([opt, "-Wno-override-module", &ir_path, "-o", &options.output])
        .output()
        .map_err(|e| {
            vec![Diagnostic::error(
                format!("cannot invoke clang: {e}. Is clang installed?"),
                Span::default(),
            )]
        })?;

    if !options.emit_ir {
        let _ = std::fs::remove_file(&ir_path);
    }

    if !output.status.success() {
        let stderr = String::from_utf8_lossy(&output.stderr);
        return Err(vec![Diagnostic::error(
            format!("clang compilation failed:\n{stderr}"),
            Span::default(),
        )]);
    }

    Ok(CodegenResult::Binary {
        path: options.output.clone(),
    })
}

// ===========================================================================
// LLVM type tracking
// ===========================================================================

#[derive(Debug, Clone, PartialEq)]
enum LType {
    I64,
    F64,
    I1,
    Str,  // ptr (i8*)
    List, // ptr → heap: [i64 len][ptr elem0][ptr elem1]...
    Void,
}

impl LType {
    fn ir(&self) -> &'static str {
        match self {
            LType::I64 => "i64",
            LType::F64 => "double",
            LType::I1 => "i1",
            LType::Str | LType::List => "ptr",
            LType::Void => "void",
        }
    }

    fn from_hir(ty: &HirType) -> Self {
        match ty {
            HirType::Int => LType::I64,
            HirType::Float => LType::F64,
            HirType::Bool => LType::I1,
            HirType::Str => LType::Str,
            HirType::List(_) => LType::List,
            HirType::Void => LType::Void,
            HirType::Enum(_) => LType::I64, // enums as int discriminants
            _ => LType::Str, // complex types → pointer
        }
    }

    /// For function params: Void → I64 (handles unresolved enum types in HIR).
    fn from_hir_param(ty: &HirType) -> Self {
        match Self::from_hir(ty) {
            LType::Void => LType::I64,
            other => other,
        }
    }
}

// ===========================================================================
// Code generator
// ===========================================================================

struct LlvmGen {
    /// Global string constants buffer.
    globals: String,
    /// Function bodies buffer.
    functions: String,
    /// String constant counter.
    str_id: usize,
    /// Cache: string content → global label.
    str_cache: HashMap<String, (String, usize)>, // (label, byte_len including null)
    /// Temp register counter (reset per function).
    tmp: usize,
    /// Variable types for current function.
    var_types: HashMap<VarId, LType>,
    /// Function return types (for call resolution).
    func_ret_types: HashMap<String, LType>,
    /// Runtime helpers that need to be emitted.
    needed: HashSet<String>,
}

impl LlvmGen {
    fn new() -> Self {
        Self {
            globals: String::new(),
            functions: String::new(),
            str_id: 0,
            str_cache: HashMap::new(),
            tmp: 0,
            var_types: HashMap::new(),
            func_ret_types: HashMap::new(),
            needed: HashSet::new(),
        }
    }

    // -- Naming helpers --

    fn fresh_tmp(&mut self) -> String {
        let name = format!("%t{}", self.tmp);
        self.tmp += 1;
        name
    }

    fn var(id: VarId) -> String {
        format!("%v{id}")
    }

    fn label(id: BlockId) -> String {
        format!("bb{id}")
    }

    fn var_type(&self, id: VarId) -> LType {
        self.var_types.get(&id).cloned().unwrap_or(LType::I64)
    }

    /// Sanitize a name for use as an LLVM IR identifier.
    fn sanitize_name(name: &str) -> String {
        if name.contains(|c: char| !c.is_alphanumeric() && c != '_' && c != '.') {
            // Quote with double quotes for LLVM
            format!("\"{}\"", name.replace('"', "\\22"))
        } else {
            name.to_string()
        }
    }

    // -- String interning --

    fn intern_str(&mut self, s: &str) -> (String, usize) {
        if let Some(entry) = self.str_cache.get(s) {
            return entry.clone();
        }
        let id = self.str_id;
        self.str_id += 1;
        let lbl = format!(".str.{id}");
        let byte_len = s.len() + 1; // +1 for null terminator

        let escaped = escape_llvm_str(s);
        writeln!(
            self.globals,
            "@{lbl} = private unnamed_addr constant [{byte_len} x i8] c\"{escaped}\\00\""
        )
        .unwrap();

        self.str_cache
            .insert(s.to_string(), (lbl.clone(), byte_len));
        (lbl, byte_len)
    }

    /// Get a ptr to a global string constant. Emits a GEP into self.functions.
    fn str_ptr(&mut self, lbl: &str, byte_len: usize) -> String {
        let tmp = self.fresh_tmp();
        writeln!(
            self.functions,
            "  {tmp} = getelementptr [{byte_len} x i8], ptr @{lbl}, i64 0, i64 0"
        )
        .unwrap();
        tmp
    }

    // -- Load / Store helpers --

    fn load(&mut self, id: VarId) -> String {
        let ty = self.var_type(id);
        let tmp = self.fresh_tmp();
        writeln!(self.functions, "  {tmp} = load {}, ptr {}", ty.ir(), Self::var(id)).unwrap();
        tmp
    }

    fn store(&mut self, id: VarId, val: &str) {
        let ty = self.var_type(id);
        if ty == LType::Void {
            return; // can't store void
        }
        writeln!(
            self.functions,
            "  store {} {val}, ptr {}",
            ty.ir(),
            Self::var(id)
        )
        .unwrap();
    }

    // -----------------------------------------------------------------------
    // Module generation
    // -----------------------------------------------------------------------

    fn generate(&mut self, module: &HirModule) -> Result<(), Vec<Diagnostic>> {
        // Pre-register function return types
        for func in &module.functions {
            self.func_ret_types
                .insert(func.name.clone(), LType::from_hir(&func.return_ty));
        }

        // Emit each function
        for func in &module.functions {
            self.emit_function(func, module)?;
        }

        // Emit C main wrapper
        if let Some(entry_idx) = module.entry {
            let entry = Self::sanitize_name(&module.functions[entry_idx].name);
            writeln!(self.functions, "define i32 @main() {{").unwrap();
            writeln!(self.functions, "  call void @{entry}()").unwrap();
            writeln!(self.functions, "  ret i32 0").unwrap();
            writeln!(self.functions, "}}").unwrap();
        }

        Ok(())
    }

    fn emit_function(
        &mut self,
        func: &HirFunction,
        module: &HirModule,
    ) -> Result<(), Vec<Diagnostic>> {
        self.tmp = 0;
        self.var_types.clear();

        // Phase 1: Infer types for all VarIds
        self.infer_types(func, module);

        // Phase 2: Function header
        let ret_ty = LType::from_hir(&func.return_ty);
        let fn_name = Self::sanitize_name(&func.name);

        if func.params.is_empty() {
            writeln!(self.functions, "define {} @{fn_name}() {{", ret_ty.ir()).unwrap();
        } else {
            let params: Vec<String> = func
                .params
                .iter()
                .enumerate()
                .map(|(i, p)| format!("{} %arg{i}", LType::from_hir_param(&p.ty).ir()))
                .collect();
            writeln!(
                self.functions,
                "define {} @{fn_name}({}) {{",
                ret_ty.ir(),
                params.join(", ")
            )
            .unwrap();
        }

        // Phase 3: Entry block with allocas
        writeln!(self.functions, "entry:").unwrap();

        let mut all_vars: Vec<VarId> = self.var_types.keys().copied().collect();
        all_vars.sort();
        for vid in &all_vars {
            let ty = &self.var_types[vid];
            if *ty == LType::Void {
                continue; // can't alloca void
            }
            writeln!(
                self.functions,
                "  {} = alloca {}",
                Self::var(*vid),
                ty.ir()
            )
            .unwrap();
        }

        // Store params into their allocas
        for (i, param) in func.params.iter().enumerate() {
            let ty = LType::from_hir_param(&param.ty);
            writeln!(
                self.functions,
                "  store {} %arg{i}, ptr {}",
                ty.ir(),
                Self::var(i)
            )
            .unwrap();
        }

        if func.blocks.is_empty() {
            if ret_ty == LType::Void {
                writeln!(self.functions, "  ret void").unwrap();
            } else {
                writeln!(self.functions, "  ret {} 0", ret_ty.ir()).unwrap();
            }
            writeln!(self.functions, "}}").unwrap();
            writeln!(self.functions).unwrap();
            return Ok(());
        }

        writeln!(self.functions, "  br label %bb0").unwrap();

        // Phase 4: Emit blocks
        for block in &func.blocks {
            writeln!(self.functions, "{}:", Self::label(block.id)).unwrap();
            for inst in &block.insts {
                self.emit_inst(inst, module)?;
            }
            self.emit_terminator(&block.terminator, &ret_ty);
        }

        writeln!(self.functions, "}}").unwrap();
        writeln!(self.functions).unwrap();
        Ok(())
    }

    // -----------------------------------------------------------------------
    // Type inference
    // -----------------------------------------------------------------------

    fn infer_types(&mut self, func: &HirFunction, module: &HirModule) {
        for (i, param) in func.params.iter().enumerate() {
            self.var_types.insert(i, LType::from_hir_param(&param.ty));
        }
        for block in &func.blocks {
            for inst in &block.insts {
                self.infer_inst(inst, module);
            }
        }
    }

    fn infer_inst(&mut self, inst: &HirInst, module: &HirModule) {
        match inst {
            HirInst::Const { dst, value } => {
                let ty = match value {
                    HirConst::Int(_) => LType::I64,
                    HirConst::Float(_) => LType::F64,
                    HirConst::Str(_) => LType::Str,
                    HirConst::Bool(_) => LType::I1,
                    HirConst::List(_) => LType::List,
                    HirConst::None | HirConst::Map(_) | HirConst::EnvRef(_) => LType::Str,
                };
                self.var_types.insert(*dst, ty);
            }
            HirInst::Assign { dst, value } => {
                let ty = match value {
                    HirValue::Use(src) => self.var_type(*src),
                    HirValue::Const(c) => match c {
                        HirConst::Int(_) => LType::I64,
                        HirConst::Float(_) => LType::F64,
                        HirConst::Str(_) => LType::Str,
                        HirConst::Bool(_) => LType::I1,
                        _ => LType::Str,
                    },
                };
                self.var_types.insert(*dst, ty);
            }
            HirInst::BinOp {
                dst, op, left, right, ..
            } => {
                let lt = self.var_type(*left);
                let rt = self.var_type(*right);
                let ty = match op {
                    BinOp::Eq | BinOp::Ne | BinOp::Lt | BinOp::Gt | BinOp::Le | BinOp::Ge
                    | BinOp::And | BinOp::Or => LType::I1,
                    BinOp::Add if lt == LType::Str || rt == LType::Str => LType::Str,
                    BinOp::Add | BinOp::Sub | BinOp::Mul | BinOp::Div | BinOp::Mod => {
                        if lt == LType::F64 || rt == LType::F64 {
                            LType::F64
                        } else {
                            LType::I64
                        }
                    }
                    _ => LType::I64, // bitwise
                };
                self.var_types.insert(*dst, ty);
            }
            HirInst::UnOp { dst, op, operand } => {
                let ty = match op {
                    UnOp::Neg => self.var_type(*operand),
                    UnOp::Not => LType::I1,
                    UnOp::BitNot => LType::I64,
                };
                self.var_types.insert(*dst, ty);
            }
            HirInst::Call { dst, func: fref, .. } => {
                if let Some(d) = dst {
                    let ty = match fref {
                        FuncRef::Builtin(name) => builtin_ret_type(name),
                        FuncRef::Local(id) => module
                            .functions
                            .get(*id)
                            .map(|f| LType::from_hir(&f.return_ty))
                            .unwrap_or(LType::Str),
                        FuncRef::External(_) => LType::Str,
                    };
                    self.var_types.insert(*d, ty);
                }
            }
            HirInst::MethodCall { dst, .. } => {
                if let Some(d) = dst {
                    self.var_types.insert(*d, LType::Str);
                }
            }
            HirInst::GetField { dst, .. }
            | HirInst::ConstructStruct { dst, .. }
            | HirInst::ConstructMap { dst, .. } => {
                self.var_types.insert(*dst, LType::Str);
            }
            HirInst::ConstructList { dst, .. } => {
                self.var_types.insert(*dst, LType::List);
            }
            HirInst::GetIndex { dst, object, .. } => {
                // Indexing a list produces an element (ptr); indexing a string produces a string.
                let obj_ty = self.var_type(*object);
                match obj_ty {
                    LType::List => self.var_types.insert(*dst, LType::Str),
                    _ => self.var_types.insert(*dst, LType::Str),
                };
            }
            HirInst::AgentAsk { dst, .. }
            | HirInst::AgentRun { dst, .. }
            | HirInst::AgentStream { dst, .. }
            | HirInst::ConstructAgent { dst, .. } => {
                self.var_types.insert(*dst, LType::Str);
            }
            HirInst::Propagate { dst, inner } => {
                self.var_types.insert(*dst, self.var_type(*inner));
            }
            HirInst::Spawn { dst, .. } => {
                self.var_types.insert(*dst, LType::I64);
            }
            HirInst::ChanRecv { dst, .. } => {
                self.var_types.insert(*dst, LType::Str);
            }
            _ => {}
        }
    }

    // -----------------------------------------------------------------------
    // Instruction emission
    // -----------------------------------------------------------------------

    fn emit_inst(&mut self, inst: &HirInst, module: &HirModule) -> Result<(), Vec<Diagnostic>> {
        match inst {
            HirInst::Const { dst, value } => self.emit_const(*dst, value),
            HirInst::Assign { dst, value } => self.emit_assign(*dst, value),
            HirInst::BinOp {
                dst,
                op,
                left,
                right,
            } => self.emit_binop(*dst, *op, *left, *right),
            HirInst::UnOp { dst, op, operand } => self.emit_unop(*dst, *op, *operand),
            HirInst::Call {
                dst,
                func: fref,
                args,
            } => self.emit_call(*dst, fref, args, module)?,
            HirInst::MethodCall {
                dst,
                receiver,
                method,
                args,
            } => self.emit_method_call(*dst, *receiver, method, args)?,
            HirInst::Nop => {}
            HirInst::Propagate { dst, inner } => {
                let val = self.load(*inner);
                self.store(*dst, &val);
            }

            // Stubs for constructs not yet compiled to native
            HirInst::GetField { dst, field, .. } => {
                writeln!(self.functions, "  ; TODO: GetField .{field}").unwrap();
                self.store(*dst, "null");
            }
            HirInst::SetField { field, .. } => {
                writeln!(self.functions, "  ; TODO: SetField .{field}").unwrap();
            }
            HirInst::GetIndex { dst, object, index } => {
                self.emit_get_index(*dst, *object, *index);
            }
            HirInst::SetIndex { .. } => {
                writeln!(self.functions, "  ; TODO: SetIndex").unwrap();
            }
            HirInst::ConstructStruct { dst, type_name, .. } => {
                writeln!(self.functions, "  ; TODO: ConstructStruct {type_name}").unwrap();
                self.store(*dst, "null");
            }
            HirInst::ConstructList { dst, elems } => {
                self.emit_construct_list(*dst, elems);
            }
            HirInst::ConstructMap { dst, .. } => {
                writeln!(self.functions, "  ; TODO: ConstructMap").unwrap();
                self.store(*dst, "null");
            }
            HirInst::ToolRegister { name, .. } => {
                writeln!(self.functions, "  ; TODO: ToolRegister {name}").unwrap();
            }
            HirInst::AgentAsk { dst, agent, .. } => {
                writeln!(self.functions, "  ; TODO: AgentAsk {agent}").unwrap();
                let (lbl, len) = self.intern_str("[agent stub]");
                let ptr = self.str_ptr(&lbl, len);
                self.store(*dst, &ptr);
            }
            HirInst::AgentRun { dst, agent, .. } => {
                writeln!(self.functions, "  ; TODO: AgentRun {agent}").unwrap();
                let (lbl, len) = self.intern_str("[agent stub]");
                let ptr = self.str_ptr(&lbl, len);
                self.store(*dst, &ptr);
            }
            HirInst::AgentStream { dst, .. } => {
                writeln!(self.functions, "  ; TODO: AgentStream").unwrap();
                self.store(*dst, "null");
            }
            HirInst::Spawn { dst, .. } => {
                writeln!(self.functions, "  ; TODO: Spawn").unwrap();
                self.store(*dst, "0");
            }
            HirInst::ChanRecv { dst, .. } => {
                writeln!(self.functions, "  ; TODO: ChanRecv").unwrap();
                self.store(*dst, "null");
            }
            HirInst::ChanSend { .. } => {
                writeln!(self.functions, "  ; TODO: ChanSend").unwrap();
            }
            HirInst::ConstructAgent { dst, .. } => {
                writeln!(self.functions, "  ; TODO: ConstructAgent").unwrap();
                self.store(*dst, "null");
            }
        }
        Ok(())
    }

    fn emit_const(&mut self, dst: VarId, value: &HirConst) {
        match value {
            HirConst::Int(n) => self.store(dst, &n.to_string()),
            HirConst::Float(f) => {
                // Use exact hex representation for LLVM IR
                let bits = f.to_bits();
                self.store(dst, &format!("0x{bits:016X}"));
            }
            HirConst::Bool(b) => self.store(dst, if *b { "1" } else { "0" }),
            HirConst::Str(s) => {
                let (lbl, len) = self.intern_str(s);
                let ptr = self.str_ptr(&lbl, len);
                self.store(dst, &ptr);
            }
            HirConst::None => self.store(dst, "null"),
            HirConst::List(_) | HirConst::Map(_) => self.store(dst, "null"),
            HirConst::EnvRef(_) => {
                // EnvRef is resolved at runtime in the interpreter.
                // For LLVM native codegen, emit as null placeholder.
                self.store(dst, "null");
            }
        }
    }

    fn emit_assign(&mut self, dst: VarId, value: &HirValue) {
        match value {
            HirValue::Use(src) => {
                let val = self.load(*src);
                self.store(dst, &val);
            }
            HirValue::Const(c) => self.emit_const(dst, c),
        }
    }

    fn emit_binop(&mut self, dst: VarId, op: BinOp, left: VarId, right: VarId) {
        let lt = self.var_type(left);
        let rt = self.var_type(right);
        let lval = self.load(left);
        let rval = self.load(right);

        // String concatenation
        if op == BinOp::Add && (lt == LType::Str || rt == LType::Str) {
            self.needed.insert("str_concat".into());
            let r = self.fresh_tmp();
            writeln!(
                self.functions,
                "  {r} = call ptr @haira_str_concat(ptr {lval}, ptr {rval})"
            )
            .unwrap();
            self.store(dst, &r);
            return;
        }

        // String comparison
        if matches!(op, BinOp::Eq | BinOp::Ne) && lt == LType::Str && rt == LType::Str {
            self.needed.insert("str_eq".into());
            let cmp = self.fresh_tmp();
            writeln!(
                self.functions,
                "  {cmp} = call i1 @haira_str_eq(ptr {lval}, ptr {rval})"
            )
            .unwrap();
            if op == BinOp::Ne {
                let neg = self.fresh_tmp();
                writeln!(self.functions, "  {neg} = xor i1 {cmp}, 1").unwrap();
                self.store(dst, &neg);
            } else {
                self.store(dst, &cmp);
            }
            return;
        }

        let is_float = lt == LType::F64 || rt == LType::F64;
        let ty = if is_float { "double" } else { "i64" };

        // Promote int → float if mixed
        let (lval, rval) = if is_float && lt != rt {
            if lt == LType::I64 {
                let c = self.fresh_tmp();
                writeln!(self.functions, "  {c} = sitofp i64 {lval} to double").unwrap();
                (c, rval)
            } else {
                let c = self.fresh_tmp();
                writeln!(self.functions, "  {c} = sitofp i64 {rval} to double").unwrap();
                (lval, c)
            }
        } else {
            (lval, rval)
        };

        let r = self.fresh_tmp();

        match op {
            BinOp::Add => {
                let i = if is_float { "fadd" } else { "add" };
                writeln!(self.functions, "  {r} = {i} {ty} {lval}, {rval}").unwrap();
            }
            BinOp::Sub => {
                let i = if is_float { "fsub" } else { "sub" };
                writeln!(self.functions, "  {r} = {i} {ty} {lval}, {rval}").unwrap();
            }
            BinOp::Mul => {
                let i = if is_float { "fmul" } else { "mul" };
                writeln!(self.functions, "  {r} = {i} {ty} {lval}, {rval}").unwrap();
            }
            BinOp::Div => {
                let i = if is_float { "fdiv" } else { "sdiv" };
                writeln!(self.functions, "  {r} = {i} {ty} {lval}, {rval}").unwrap();
            }
            BinOp::Mod => {
                let i = if is_float { "frem" } else { "srem" };
                writeln!(self.functions, "  {r} = {i} {ty} {lval}, {rval}").unwrap();
            }
            BinOp::Eq => {
                let i = if is_float { "fcmp oeq" } else { "icmp eq" };
                writeln!(self.functions, "  {r} = {i} {ty} {lval}, {rval}").unwrap();
            }
            BinOp::Ne => {
                let i = if is_float { "fcmp one" } else { "icmp ne" };
                writeln!(self.functions, "  {r} = {i} {ty} {lval}, {rval}").unwrap();
            }
            BinOp::Lt => {
                let i = if is_float { "fcmp olt" } else { "icmp slt" };
                writeln!(self.functions, "  {r} = {i} {ty} {lval}, {rval}").unwrap();
            }
            BinOp::Gt => {
                let i = if is_float { "fcmp ogt" } else { "icmp sgt" };
                writeln!(self.functions, "  {r} = {i} {ty} {lval}, {rval}").unwrap();
            }
            BinOp::Le => {
                let i = if is_float { "fcmp ole" } else { "icmp sle" };
                writeln!(self.functions, "  {r} = {i} {ty} {lval}, {rval}").unwrap();
            }
            BinOp::Ge => {
                let i = if is_float { "fcmp oge" } else { "icmp sge" };
                writeln!(self.functions, "  {r} = {i} {ty} {lval}, {rval}").unwrap();
            }
            BinOp::And => {
                // Handle int truthiness: convert to i1 if needed
                let lv = if lt == LType::I64 {
                    let c = self.fresh_tmp();
                    writeln!(self.functions, "  {c} = icmp ne i64 {lval}, 0").unwrap();
                    c
                } else {
                    lval
                };
                let rv = if rt == LType::I64 {
                    let c = self.fresh_tmp();
                    writeln!(self.functions, "  {c} = icmp ne i64 {rval}, 0").unwrap();
                    c
                } else {
                    rval
                };
                writeln!(self.functions, "  {r} = and i1 {lv}, {rv}").unwrap();
            }
            BinOp::Or => {
                let lv = if lt == LType::I64 {
                    let c = self.fresh_tmp();
                    writeln!(self.functions, "  {c} = icmp ne i64 {lval}, 0").unwrap();
                    c
                } else {
                    lval
                };
                let rv = if rt == LType::I64 {
                    let c = self.fresh_tmp();
                    writeln!(self.functions, "  {c} = icmp ne i64 {rval}, 0").unwrap();
                    c
                } else {
                    rval
                };
                writeln!(self.functions, "  {r} = or i1 {lv}, {rv}").unwrap();
            }
            BinOp::BitAnd => {
                writeln!(self.functions, "  {r} = and i64 {lval}, {rval}").unwrap()
            }
            BinOp::BitOr => {
                writeln!(self.functions, "  {r} = or i64 {lval}, {rval}").unwrap()
            }
            BinOp::BitXor => {
                writeln!(self.functions, "  {r} = xor i64 {lval}, {rval}").unwrap()
            }
            BinOp::Shl => {
                writeln!(self.functions, "  {r} = shl i64 {lval}, {rval}").unwrap()
            }
            BinOp::Shr => {
                writeln!(self.functions, "  {r} = ashr i64 {lval}, {rval}").unwrap()
            }
        }

        self.store(dst, &r);
    }

    fn emit_unop(&mut self, dst: VarId, op: UnOp, operand: VarId) {
        let val = self.load(operand);
        let r = self.fresh_tmp();

        match op {
            UnOp::Neg => {
                if self.var_type(operand) == LType::F64 {
                    writeln!(self.functions, "  {r} = fneg double {val}").unwrap();
                } else {
                    writeln!(self.functions, "  {r} = sub i64 0, {val}").unwrap();
                }
            }
            UnOp::Not => writeln!(self.functions, "  {r} = xor i1 {val}, 1").unwrap(),
            UnOp::BitNot => writeln!(self.functions, "  {r} = xor i64 {val}, -1").unwrap(),
        }

        self.store(dst, &r);
    }

    fn emit_call(
        &mut self,
        dst: Option<VarId>,
        fref: &FuncRef,
        args: &[VarId],
        module: &HirModule,
    ) -> Result<(), Vec<Diagnostic>> {
        match fref {
            FuncRef::Builtin(name) => self.emit_builtin(dst, name, args),
            FuncRef::Local(id) => {
                let func = &module.functions[*id];
                let ret = LType::from_hir(&func.return_ty);
                let fn_name = Self::sanitize_name(&func.name);

                let arg_strs: Vec<String> = args
                    .iter()
                    .map(|&a| {
                        let val = self.load(a);
                        let ty = self.var_type(a);
                        format!("{} {val}", ty.ir())
                    })
                    .collect();

                if ret == LType::Void {
                    writeln!(
                        self.functions,
                        "  call void @{fn_name}({})",
                        arg_strs.join(", ")
                    )
                    .unwrap();
                } else {
                    let r = self.fresh_tmp();
                    writeln!(
                        self.functions,
                        "  {r} = call {} @{fn_name}({})",
                        ret.ir(),
                        arg_strs.join(", ")
                    )
                    .unwrap();
                    if let Some(d) = dst {
                        self.store(d, &r);
                    }
                }
            }
            FuncRef::External(name) => {
                writeln!(self.functions, "  ; external: {name} (stub)").unwrap();
                if let Some(d) = dst {
                    self.store(d, "null");
                }
            }
        }
        Ok(())
    }

    /// Convert a loaded value to a string ptr, emitting a to_string call if needed.
    fn coerce_to_str(&mut self, val: &str, ty: &LType) -> String {
        match ty {
            LType::Str => val.to_string(),
            LType::I64 => {
                self.needed.insert("i64_to_str".into());
                let r = self.fresh_tmp();
                writeln!(self.functions, "  {r} = call ptr @haira_i64_to_str(i64 {val})").unwrap();
                r
            }
            LType::F64 => {
                self.needed.insert("f64_to_str".into());
                let r = self.fresh_tmp();
                writeln!(
                    self.functions,
                    "  {r} = call ptr @haira_f64_to_str(double {val})"
                )
                .unwrap();
                r
            }
            LType::I1 => {
                self.needed.insert("bool_to_str".into());
                let r = self.fresh_tmp();
                writeln!(
                    self.functions,
                    "  {r} = call ptr @haira_bool_to_str(i1 {val})"
                )
                .unwrap();
                r
            }
            _ => val.to_string(),
        }
    }

    fn emit_builtin(&mut self, dst: Option<VarId>, name: &str, args: &[VarId]) {
        match name {
            "io.println" => {
                self.needed.insert("println".into());
                let arg_ty = self.var_type(args[0]);
                let val = self.load(args[0]);
                let val = self.coerce_to_str(&val, &arg_ty);
                writeln!(self.functions, "  call void @haira_println(ptr {val})").unwrap();
            }
            "io.print" => {
                self.needed.insert("print".into());
                let arg_ty = self.var_type(args[0]);
                let val = self.load(args[0]);
                let val = self.coerce_to_str(&val, &arg_ty);
                writeln!(self.functions, "  call void @haira_print(ptr {val})").unwrap();
            }
            "io.eprintln" => {
                self.needed.insert("eprintln".into());
                let arg_ty = self.var_type(args[0]);
                let val = self.load(args[0]);
                let val = self.coerce_to_str(&val, &arg_ty);
                writeln!(self.functions, "  call void @haira_eprintln(ptr {val})").unwrap();
            }
            "to_string" | "conv.int_to_string" | "conv.float_to_string"
            | "conv.bool_to_string" => {
                let arg_ty = self.var_type(args[0]);
                let val = self.load(args[0]);
                let result = match arg_ty {
                    LType::I64 => {
                        self.needed.insert("i64_to_str".into());
                        let r = self.fresh_tmp();
                        writeln!(
                            self.functions,
                            "  {r} = call ptr @haira_i64_to_str(i64 {val})"
                        )
                        .unwrap();
                        r
                    }
                    LType::F64 => {
                        self.needed.insert("f64_to_str".into());
                        let r = self.fresh_tmp();
                        writeln!(
                            self.functions,
                            "  {r} = call ptr @haira_f64_to_str(double {val})"
                        )
                        .unwrap();
                        r
                    }
                    LType::I1 => {
                        self.needed.insert("bool_to_str".into());
                        let r = self.fresh_tmp();
                        writeln!(
                            self.functions,
                            "  {r} = call ptr @haira_bool_to_str(i1 {val})"
                        )
                        .unwrap();
                        r
                    }
                    LType::Str => val,
                    _ => val,
                };
                if let Some(d) = dst {
                    self.store(d, &result);
                }
            }
            "conv.string_to_int" => {
                self.needed.insert("str_to_i64".into());
                let val = self.load(args[0]);
                let r = self.fresh_tmp();
                writeln!(
                    self.functions,
                    "  {r} = call i64 @haira_str_to_i64(ptr {val})"
                )
                .unwrap();
                if let Some(d) = dst {
                    self.store(d, &r);
                }
            }
            "conv.float_to_int" => {
                let val = self.load(args[0]);
                let r = self.fresh_tmp();
                writeln!(self.functions, "  {r} = fptosi double {val} to i64").unwrap();
                if let Some(d) = dst {
                    self.store(d, &r);
                }
            }
            "conv.int_to_float" => {
                let val = self.load(args[0]);
                let r = self.fresh_tmp();
                writeln!(self.functions, "  {r} = sitofp i64 {val} to double").unwrap();
                if let Some(d) = dst {
                    self.store(d, &r);
                }
            }
            "len" | "string.len" | "array.len" => {
                let arg_ty = self.var_type(args[0]);
                let val = self.load(args[0]);
                let r = self.fresh_tmp();
                if arg_ty == LType::List {
                    // List: length is stored as i64 at offset 0
                    writeln!(self.functions, "  {r} = load i64, ptr {val}").unwrap();
                } else {
                    // String: use strlen
                    writeln!(self.functions, "  {r} = call i64 @strlen(ptr {val})").unwrap();
                }
                if let Some(d) = dst {
                    self.store(d, &r);
                }
            }
            "string.contains" => {
                self.needed.insert("str_contains".into());
                let a = self.load(args[0]);
                let b = self.load(args[1]);
                let r = self.fresh_tmp();
                writeln!(
                    self.functions,
                    "  {r} = call i1 @haira_str_contains(ptr {a}, ptr {b})"
                )
                .unwrap();
                if let Some(d) = dst {
                    self.store(d, &r);
                }
            }
            "panic" => {
                self.needed.insert("panic".into());
                let val = self.load(args[0]);
                writeln!(self.functions, "  call void @haira_panic(ptr {val})").unwrap();
                writeln!(self.functions, "  unreachable").unwrap();
            }
            _ => {
                writeln!(self.functions, "  ; builtin: {name} (stub)").unwrap();
                if let Some(d) = dst {
                    let ty = self.var_type(d);
                    match ty {
                        LType::I64 => self.store(d, "0"),
                        LType::F64 => self.store(d, "0.0"),
                        LType::I1 => self.store(d, "0"),
                        _ => self.store(d, "null"),
                    }
                }
            }
        }
    }

    fn emit_method_call(
        &mut self,
        dst: Option<VarId>,
        receiver: VarId,
        method: &str,
        args: &[VarId],
    ) -> Result<(), Vec<Diagnostic>> {
        let recv_ty = self.var_type(receiver);

        match (&recv_ty, method) {
            (LType::Str, "len") => {
                let val = self.load(receiver);
                let r = self.fresh_tmp();
                writeln!(self.functions, "  {r} = call i64 @strlen(ptr {val})").unwrap();
                if let Some(d) = dst {
                    self.store(d, &r);
                }
            }
            (LType::List, "len") => {
                let val = self.load(receiver);
                let r = self.fresh_tmp();
                writeln!(self.functions, "  {r} = load i64, ptr {val}").unwrap();
                if let Some(d) = dst {
                    self.store(d, &r);
                }
            }
            (LType::Str, "contains") => {
                self.needed.insert("str_contains".into());
                let a = self.load(receiver);
                let b = self.load(args[0]);
                let r = self.fresh_tmp();
                writeln!(
                    self.functions,
                    "  {r} = call i1 @haira_str_contains(ptr {a}, ptr {b})"
                )
                .unwrap();
                if let Some(d) = dst {
                    self.store(d, &r);
                }
            }
            _ => {
                writeln!(self.functions, "  ; method: {method} (stub)").unwrap();
                if let Some(d) = dst {
                    self.store(d, "null");
                }
            }
        }
        Ok(())
    }

    // -----------------------------------------------------------------------
    // List operations
    // -----------------------------------------------------------------------

    /// Emit ConstructList: malloc [i64 len][ptr e0][ptr e1]..., store elements.
    fn emit_construct_list(&mut self, dst: VarId, elems: &[VarId]) {
        let n = elems.len();
        // Total size: 8 (length) + N * 8 (element pointers)
        let total = 8 + n * 8;
        let raw = self.fresh_tmp();
        writeln!(self.functions, "  {raw} = call ptr @malloc(i64 {total})").unwrap();

        // Store length at offset 0
        writeln!(self.functions, "  store i64 {n}, ptr {raw}").unwrap();

        // Store each element at offset 8 + i*8
        for (i, elem_var) in elems.iter().enumerate() {
            let val = self.load(*elem_var);
            let elem_ty = self.var_type(*elem_var);
            let offset = 8 + i * 8;
            let ep = self.fresh_tmp();
            writeln!(self.functions, "  {ep} = getelementptr i8, ptr {raw}, i64 {offset}").unwrap();
            writeln!(self.functions, "  store {} {val}, ptr {ep}", elem_ty.ir()).unwrap();
        }

        // Store the list pointer
        self.store(dst, &raw);
    }

    /// Emit GetIndex: load element at list[index].
    fn emit_get_index(&mut self, dst: VarId, object: VarId, index: VarId) {
        let list_ptr = self.load(object);
        let idx = self.load(index);

        // Element pointer = list_ptr + 8 + idx * 8
        let offset_base = self.fresh_tmp();
        writeln!(self.functions, "  {offset_base} = mul i64 {idx}, 8").unwrap();
        let offset = self.fresh_tmp();
        writeln!(self.functions, "  {offset} = add i64 {offset_base}, 8").unwrap();
        let ep = self.fresh_tmp();
        writeln!(self.functions, "  {ep} = getelementptr i8, ptr {list_ptr}, i64 {offset}").unwrap();

        // Load the element
        let dst_ty = self.var_type(dst);
        let val = self.fresh_tmp();
        writeln!(self.functions, "  {val} = load {}, ptr {ep}", dst_ty.ir()).unwrap();
        self.store(dst, &val);
    }

    // -----------------------------------------------------------------------
    // Terminators
    // -----------------------------------------------------------------------

    fn emit_terminator(&mut self, term: &Terminator, ret_ty: &LType) {
        match term {
            Terminator::Return(None) => {
                if *ret_ty == LType::Void {
                    writeln!(self.functions, "  ret void").unwrap();
                } else {
                    let def = match ret_ty {
                        LType::I64 => "0",
                        LType::F64 => "0.0",
                        LType::I1 => "0",
                        LType::Str | LType::List => "null",
                        LType::Void => unreachable!(),
                    };
                    writeln!(self.functions, "  ret {} {def}", ret_ty.ir()).unwrap();
                }
            }
            Terminator::Return(Some(vid)) => {
                let val = self.load(*vid);
                writeln!(self.functions, "  ret {} {val}", ret_ty.ir()).unwrap();
            }
            Terminator::Goto(block) => {
                writeln!(self.functions, "  br label %{}", Self::label(*block)).unwrap();
            }
            Terminator::Branch {
                cond,
                then_block,
                else_block,
            } => {
                let c = self.load(*cond);
                // Convert to i1 if needed
                let c = if self.var_type(*cond) == LType::I64 {
                    let conv = self.fresh_tmp();
                    writeln!(self.functions, "  {conv} = icmp ne i64 {c}, 0").unwrap();
                    conv
                } else {
                    c
                };
                writeln!(
                    self.functions,
                    "  br i1 {c}, label %{}, label %{}",
                    Self::label(*then_block),
                    Self::label(*else_block)
                )
                .unwrap();
            }
            Terminator::Switch {
                scrutinee,
                cases,
                default,
            } => {
                let val = self.load(*scrutinee);
                let ty = self.var_type(*scrutinee);
                let cases_str: Vec<String> = cases
                    .iter()
                    .map(|(c, bid)| {
                        let cv = match c {
                            HirConst::Int(n) => n.to_string(),
                            HirConst::Bool(b) => {
                                if *b {
                                    "1".into()
                                } else {
                                    "0".into()
                                }
                            }
                            _ => "0".into(),
                        };
                        format!("    {} {cv}, label %{}", ty.ir(), Self::label(*bid))
                    })
                    .collect();
                writeln!(
                    self.functions,
                    "  switch {} {val}, label %{} [\n{}\n  ]",
                    ty.ir(),
                    Self::label(*default),
                    cases_str.join("\n")
                )
                .unwrap();
            }
            Terminator::Unreachable => {
                writeln!(self.functions, "  unreachable").unwrap();
            }
        }
    }

    // -----------------------------------------------------------------------
    // Runtime helpers emission
    // -----------------------------------------------------------------------

    fn emit_runtime(&self) -> String {
        let mut rt = String::new();

        // External C library declarations
        writeln!(rt, "; C library declarations").unwrap();
        writeln!(rt, "declare i32 @puts(ptr)").unwrap();
        writeln!(rt, "declare i32 @printf(ptr, ...)").unwrap();
        writeln!(rt, "declare i32 @dprintf(i32, ptr, ...)").unwrap();
        writeln!(rt, "declare ptr @malloc(i64)").unwrap();
        writeln!(rt, "declare void @free(ptr)").unwrap();
        writeln!(rt, "declare i64 @strlen(ptr)").unwrap();
        writeln!(rt, "declare ptr @strcpy(ptr, ptr)").unwrap();
        writeln!(rt, "declare ptr @strcat(ptr, ptr)").unwrap();
        writeln!(rt, "declare i32 @snprintf(ptr, i64, ptr, ...)").unwrap();
        writeln!(rt, "declare i32 @strcmp(ptr, ptr)").unwrap();
        writeln!(rt, "declare ptr @strstr(ptr, ptr)").unwrap();
        writeln!(rt, "declare void @exit(i32)").unwrap();
        writeln!(rt, "declare i64 @strtol(ptr, ptr, i32)").unwrap();
        writeln!(rt).unwrap();

        // Format strings
        writeln!(rt, "; Format strings").unwrap();
        writeln!(
            rt,
            r#"@.fmt.int = private unnamed_addr constant [4 x i8] c"%ld\00""#
        )
        .unwrap();
        writeln!(
            rt,
            r#"@.fmt.float = private unnamed_addr constant [3 x i8] c"%g\00""#
        )
        .unwrap();
        writeln!(
            rt,
            r#"@.fmt.str = private unnamed_addr constant [3 x i8] c"%s\00""#
        )
        .unwrap();
        writeln!(
            rt,
            r#"@.fmt.str_nl = private unnamed_addr constant [4 x i8] c"%s\0A\00""#
        )
        .unwrap();
        writeln!(
            rt,
            r#"@.str.true = private unnamed_addr constant [5 x i8] c"true\00""#
        )
        .unwrap();
        writeln!(
            rt,
            r#"@.str.false = private unnamed_addr constant [6 x i8] c"false\00""#
        )
        .unwrap();
        writeln!(rt).unwrap();

        // Emit only needed helpers
        writeln!(rt, "; Runtime helpers").unwrap();

        if self.needed.contains("println") {
            writeln!(rt, "define void @haira_println(ptr %s) {{").unwrap();
            writeln!(rt, "  %r = call i32 @puts(ptr %s)").unwrap();
            writeln!(rt, "  ret void").unwrap();
            writeln!(rt, "}}").unwrap();
            writeln!(rt).unwrap();
        }

        if self.needed.contains("print") {
            writeln!(rt, "define void @haira_print(ptr %s) {{").unwrap();
            writeln!(rt, "  %fmt = getelementptr [3 x i8], ptr @.fmt.str, i64 0, i64 0").unwrap();
            writeln!(rt, "  call i32 (ptr, ...) @printf(ptr %fmt, ptr %s)").unwrap();
            writeln!(rt, "  ret void").unwrap();
            writeln!(rt, "}}").unwrap();
            writeln!(rt).unwrap();
        }

        if self.needed.contains("eprintln") {
            writeln!(rt, "define void @haira_eprintln(ptr %s) {{").unwrap();
            writeln!(
                rt,
                "  %fmt = getelementptr [4 x i8], ptr @.fmt.str_nl, i64 0, i64 0"
            )
            .unwrap();
            writeln!(
                rt,
                "  call i32 (i32, ptr, ...) @dprintf(i32 2, ptr %fmt, ptr %s)"
            )
            .unwrap();
            writeln!(rt, "  ret void").unwrap();
            writeln!(rt, "}}").unwrap();
            writeln!(rt).unwrap();
        }

        if self.needed.contains("str_concat") {
            writeln!(rt, "define ptr @haira_str_concat(ptr %a, ptr %b) {{").unwrap();
            writeln!(rt, "  %la = call i64 @strlen(ptr %a)").unwrap();
            writeln!(rt, "  %lb = call i64 @strlen(ptr %b)").unwrap();
            writeln!(rt, "  %total = add i64 %la, %lb").unwrap();
            writeln!(rt, "  %total1 = add i64 %total, 1").unwrap();
            writeln!(rt, "  %buf = call ptr @malloc(i64 %total1)").unwrap();
            writeln!(rt, "  call ptr @strcpy(ptr %buf, ptr %a)").unwrap();
            writeln!(rt, "  call ptr @strcat(ptr %buf, ptr %b)").unwrap();
            writeln!(rt, "  ret ptr %buf").unwrap();
            writeln!(rt, "}}").unwrap();
            writeln!(rt).unwrap();
        }

        if self.needed.contains("i64_to_str") {
            writeln!(rt, "define ptr @haira_i64_to_str(i64 %n) {{").unwrap();
            writeln!(rt, "  %buf = call ptr @malloc(i64 32)").unwrap();
            writeln!(
                rt,
                "  %fmt = getelementptr [4 x i8], ptr @.fmt.int, i64 0, i64 0"
            )
            .unwrap();
            writeln!(
                rt,
                "  call i32 (ptr, i64, ptr, ...) @snprintf(ptr %buf, i64 32, ptr %fmt, i64 %n)"
            )
            .unwrap();
            writeln!(rt, "  ret ptr %buf").unwrap();
            writeln!(rt, "}}").unwrap();
            writeln!(rt).unwrap();
        }

        if self.needed.contains("f64_to_str") {
            writeln!(rt, "define ptr @haira_f64_to_str(double %n) {{").unwrap();
            writeln!(rt, "  %buf = call ptr @malloc(i64 64)").unwrap();
            writeln!(
                rt,
                "  %fmt = getelementptr [3 x i8], ptr @.fmt.float, i64 0, i64 0"
            )
            .unwrap();
            writeln!(
                rt,
                "  call i32 (ptr, i64, ptr, ...) @snprintf(ptr %buf, i64 64, ptr %fmt, double %n)"
            )
            .unwrap();
            writeln!(rt, "  ret ptr %buf").unwrap();
            writeln!(rt, "}}").unwrap();
            writeln!(rt).unwrap();
        }

        if self.needed.contains("bool_to_str") {
            writeln!(rt, "define ptr @haira_bool_to_str(i1 %b) {{").unwrap();
            writeln!(rt, "  br i1 %b, label %is_true, label %is_false").unwrap();
            writeln!(rt, "is_true:").unwrap();
            writeln!(
                rt,
                "  %t = getelementptr [5 x i8], ptr @.str.true, i64 0, i64 0"
            )
            .unwrap();
            writeln!(rt, "  ret ptr %t").unwrap();
            writeln!(rt, "is_false:").unwrap();
            writeln!(
                rt,
                "  %f = getelementptr [6 x i8], ptr @.str.false, i64 0, i64 0"
            )
            .unwrap();
            writeln!(rt, "  ret ptr %f").unwrap();
            writeln!(rt, "}}").unwrap();
            writeln!(rt).unwrap();
        }

        if self.needed.contains("str_eq") {
            writeln!(rt, "define i1 @haira_str_eq(ptr %a, ptr %b) {{").unwrap();
            writeln!(rt, "  %r = call i32 @strcmp(ptr %a, ptr %b)").unwrap();
            writeln!(rt, "  %eq = icmp eq i32 %r, 0").unwrap();
            writeln!(rt, "  ret i1 %eq").unwrap();
            writeln!(rt, "}}").unwrap();
            writeln!(rt).unwrap();
        }

        if self.needed.contains("str_contains") {
            writeln!(
                rt,
                "define i1 @haira_str_contains(ptr %haystack, ptr %needle) {{"
            )
            .unwrap();
            writeln!(
                rt,
                "  %r = call ptr @strstr(ptr %haystack, ptr %needle)"
            )
            .unwrap();
            writeln!(rt, "  %found = icmp ne ptr %r, null").unwrap();
            writeln!(rt, "  ret i1 %found").unwrap();
            writeln!(rt, "}}").unwrap();
            writeln!(rt).unwrap();
        }

        if self.needed.contains("str_to_i64") {
            writeln!(rt, "define i64 @haira_str_to_i64(ptr %s) {{").unwrap();
            writeln!(
                rt,
                "  %r = call i64 @strtol(ptr %s, ptr null, i32 10)"
            )
            .unwrap();
            writeln!(rt, "  ret i64 %r").unwrap();
            writeln!(rt, "}}").unwrap();
            writeln!(rt).unwrap();
        }

        if self.needed.contains("panic") {
            writeln!(rt, "define void @haira_panic(ptr %msg) {{").unwrap();
            writeln!(
                rt,
                "  %fmt = getelementptr [4 x i8], ptr @.fmt.str_nl, i64 0, i64 0"
            )
            .unwrap();
            writeln!(
                rt,
                "  call i32 (i32, ptr, ...) @dprintf(i32 2, ptr %fmt, ptr %msg)"
            )
            .unwrap();
            writeln!(rt, "  call void @exit(i32 1)").unwrap();
            writeln!(rt, "  unreachable").unwrap();
            writeln!(rt, "}}").unwrap();
            writeln!(rt).unwrap();
        }

        rt
    }

    // -----------------------------------------------------------------------
    // Final assembly
    // -----------------------------------------------------------------------

    fn finish(&self) -> String {
        let mut out = String::new();

        writeln!(out, "; Generated by haira-rs LLVM codegen").unwrap();
        writeln!(out).unwrap();

        if !self.globals.is_empty() {
            writeln!(out, "; String constants").unwrap();
            out.push_str(&self.globals);
            writeln!(out).unwrap();
        }

        out.push_str(&self.emit_runtime());
        writeln!(out).unwrap();

        writeln!(out, "; User functions").unwrap();
        out.push_str(&self.functions);

        out
    }
}

// ===========================================================================
// Helpers
// ===========================================================================

fn escape_llvm_str(s: &str) -> String {
    let mut out = String::new();
    for byte in s.bytes() {
        match byte {
            b'\n' => out.push_str("\\0A"),
            b'\r' => out.push_str("\\0D"),
            b'\t' => out.push_str("\\09"),
            b'\\' => out.push_str("\\5C"),
            b'"' => out.push_str("\\22"),
            0x20..=0x7E => out.push(byte as char),
            _ => write!(out, "\\{byte:02X}").unwrap(),
        }
    }
    out
}

fn builtin_ret_type(name: &str) -> LType {
    match name {
        "io.println" | "io.print" | "io.eprintln" => LType::Void,
        "to_string" | "conv.int_to_string" | "conv.float_to_string"
        | "conv.bool_to_string" | "string.trim" | "string.to_upper"
        | "string.to_lower" | "string.replace" | "json.marshal" => LType::Str,
        "len" | "string.len" | "array.len" | "conv.string_to_int"
        | "conv.float_to_int" | "math.abs" => LType::I64,
        "string.contains" | "map.has" => LType::I1,
        "conv.string_to_float" | "conv.int_to_float" | "math.sqrt" => LType::F64,
        _ => LType::Str,
    }
}
