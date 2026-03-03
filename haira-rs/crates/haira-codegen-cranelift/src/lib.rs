//! Haira codegen backend — interpreter (with future Cranelift JIT/AOT).
//!
//! Currently implements a tree-walking interpreter that executes HIR directly.
//! The Cranelift JIT/AOT compilation paths are planned for a future phase.

use std::collections::HashMap;
use std::fmt;

use haira_errors::{Diagnostic, Span};
use haira_ir::*;

// ===========================================================================
// Public API
// ===========================================================================

/// Compilation target.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Target {
    /// Native binary via Cranelift AOT compilation (future).
    Native,
    /// In-memory execution via Cranelift JIT (future).
    Jit,
    /// Tree-walking interpreter (current default).
    Interpreter,
}

/// Result of code generation / execution.
#[derive(Debug)]
pub enum CodegenResult {
    /// Native object file bytes (for AOT compilation).
    ObjectFile(Vec<u8>),
    /// Direct execution result (for JIT or interpreter).
    Executed { exit_code: i32 },
}

/// Code generation options.
#[derive(Debug, Clone)]
pub struct CodegenOptions {
    pub target: Target,
    pub optimize: bool,
}

impl Default for CodegenOptions {
    fn default() -> Self {
        Self {
            target: Target::Interpreter,
            optimize: false,
        }
    }
}

/// Generate code / execute a HIR module.
pub fn codegen(
    module: &HirModule,
    options: &CodegenOptions,
) -> Result<CodegenResult, Vec<Diagnostic>> {
    match options.target {
        Target::Interpreter => {
            let exit_code = interpret(module)?;
            Ok(CodegenResult::Executed { exit_code })
        }
        Target::Jit | Target::Native => Err(vec![Diagnostic::error(
            format!(
                "{:?} target not yet implemented; use Interpreter",
                options.target
            ),
            Span::default(),
        )]),
    }
}

// ===========================================================================
// Interpreter value
// ===========================================================================

/// Runtime value for the interpreter.
#[derive(Clone, Debug)]
enum IValue {
    None,
    Bool(bool),
    Int(i64),
    Float(f64),
    Str(String),
    List(Vec<IValue>),
    Map(Vec<(String, IValue)>),
    Struct {
        type_name: String,
        fields: Vec<(String, IValue)>,
    },
    #[allow(dead_code)]
    Error(String),
}

impl IValue {
    fn is_truthy(&self) -> bool {
        match self {
            IValue::None => false,
            IValue::Bool(b) => *b,
            IValue::Int(n) => *n != 0,
            IValue::Float(f) => *f != 0.0,
            IValue::Str(s) => !s.is_empty(),
            IValue::List(v) => !v.is_empty(),
            IValue::Map(m) => !m.is_empty(),
            IValue::Struct { .. } => true,
            IValue::Error(_) => false,
        }
    }
}

impl fmt::Display for IValue {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            IValue::None => write!(f, "none"),
            IValue::Bool(b) => write!(f, "{b}"),
            IValue::Int(n) => write!(f, "{n}"),
            IValue::Float(v) => {
                if v.fract() == 0.0 && v.is_finite() {
                    write!(f, "{v:.1}")
                } else {
                    write!(f, "{v}")
                }
            }
            IValue::Str(s) => write!(f, "{s}"),
            IValue::List(elems) => {
                write!(f, "[")?;
                for (i, e) in elems.iter().enumerate() {
                    if i > 0 {
                        write!(f, ", ")?;
                    }
                    write!(f, "{e}")?;
                }
                write!(f, "]")
            }
            IValue::Map(entries) => {
                write!(f, "{{")?;
                for (i, (k, v)) in entries.iter().enumerate() {
                    if i > 0 {
                        write!(f, ", ")?;
                    }
                    write!(f, "{k}: {v}")?;
                }
                write!(f, "}}")
            }
            IValue::Struct { type_name, fields } => {
                write!(f, "{type_name}{{")?;
                for (i, (k, v)) in fields.iter().enumerate() {
                    if i > 0 {
                        write!(f, ", ")?;
                    }
                    write!(f, "{k}: {v}")?;
                }
                write!(f, "}}")
            }
            IValue::Error(msg) => write!(f, "error: {msg}"),
        }
    }
}

impl PartialEq for IValue {
    fn eq(&self, other: &Self) -> bool {
        match (self, other) {
            (IValue::None, IValue::None) => true,
            (IValue::Bool(a), IValue::Bool(b)) => a == b,
            (IValue::Int(a), IValue::Int(b)) => a == b,
            (IValue::Float(a), IValue::Float(b)) => a == b,
            (IValue::Int(a), IValue::Float(b)) => (*a as f64) == *b,
            (IValue::Float(a), IValue::Int(b)) => *a == (*b as f64),
            (IValue::Str(a), IValue::Str(b)) => a == b,
            (IValue::List(a), IValue::List(b)) => a == b,
            _ => false,
        }
    }
}

// ===========================================================================
// Interpreter
// ===========================================================================

/// Result of executing a single basic block.
enum BlockResult {
    /// Continue to the given block.
    Continue(BlockId),
    /// Return from the function with a value.
    Return(IValue),
}

/// The HIR interpreter.
struct Interpreter {
    /// Local variables for the current call frame.
    frames: Vec<Frame>,
    /// Captured stdout output (for testing).
    stdout: Vec<String>,
}

struct Frame {
    locals: HashMap<VarId, IValue>,
    #[allow(dead_code)]
    func_name: String,
}

impl Interpreter {
    fn new() -> Self {
        Self {
            frames: Vec::new(),
            stdout: Vec::new(),
        }
    }

    fn current_frame(&self) -> &Frame {
        self.frames.last().expect("no active call frame")
    }

    fn current_frame_mut(&mut self) -> &mut Frame {
        self.frames.last_mut().expect("no active call frame")
    }

    fn get_var(&self, id: VarId) -> IValue {
        self.current_frame()
            .locals
            .get(&id)
            .cloned()
            .unwrap_or(IValue::None)
    }

    fn set_var(&mut self, id: VarId, val: IValue) {
        self.current_frame_mut().locals.insert(id, val);
    }

    fn execute(
        &mut self,
        module: &HirModule,
    ) -> Result<i32, Vec<Diagnostic>> {
        let entry_idx = match module.entry {
            Some(idx) => idx,
            None => return Ok(0), // no entry point, nothing to execute
        };

        match self.execute_function(module, entry_idx, vec![]) {
            Ok(_) => Ok(0),
            Err(msg) => Err(vec![Diagnostic::error(
                format!("runtime error: {msg}"),
                Span::default(),
            )]),
        }
    }

    fn execute_function(
        &mut self,
        module: &HirModule,
        func_idx: usize,
        args: Vec<IValue>,
    ) -> Result<IValue, String> {
        let func = &module.functions[func_idx];

        // Push new frame
        let mut locals = HashMap::new();
        for (i, _param) in func.params.iter().enumerate() {
            let val = args.get(i).cloned().unwrap_or(IValue::None);
            // Param gets variable IDs starting at 0 for this function.
            // We use a simple convention: params are in the vars map by name
            // but for HIR they're assigned VarIds by the lowering pass.
            locals.insert(i, val);
        }

        self.frames.push(Frame {
            locals,
            func_name: func.name.clone(),
        });

        // Execute blocks starting from block 0
        if func.blocks.is_empty() {
            self.frames.pop();
            return Ok(IValue::None);
        }

        let mut block_id: BlockId = 0;
        let result = loop {
            if block_id >= func.blocks.len() {
                break Ok(IValue::None);
            }
            match self.execute_block(module, func, block_id)? {
                BlockResult::Continue(next) => block_id = next,
                BlockResult::Return(val) => break Ok(val),
            }
        };

        self.frames.pop();
        result
    }

    fn execute_block(
        &mut self,
        module: &HirModule,
        func: &HirFunction,
        block_id: BlockId,
    ) -> Result<BlockResult, String> {
        let block = &func.blocks[block_id];

        // Execute instructions
        for inst in &block.insts {
            self.execute_inst(module, inst)?;
        }

        // Execute terminator
        match &block.terminator {
            Terminator::Goto(target) => Ok(BlockResult::Continue(*target)),
            Terminator::Branch {
                cond,
                then_block,
                else_block,
            } => {
                let cond_val = self.get_var(*cond);
                if cond_val.is_truthy() {
                    Ok(BlockResult::Continue(*then_block))
                } else {
                    Ok(BlockResult::Continue(*else_block))
                }
            }
            Terminator::Return(var) => {
                let val = var.map(|v| self.get_var(v)).unwrap_or(IValue::None);
                Ok(BlockResult::Return(val))
            }
            Terminator::Switch {
                scrutinee,
                cases,
                default,
            } => {
                let val = self.get_var(*scrutinee);
                for (pattern, target) in cases {
                    if values_match(&val, pattern) {
                        return Ok(BlockResult::Continue(*target));
                    }
                }
                Ok(BlockResult::Continue(*default))
            }
            Terminator::Unreachable => Err("reached unreachable code".to_string()),
        }
    }

    fn execute_inst(
        &mut self,
        module: &HirModule,
        inst: &HirInst,
    ) -> Result<(), String> {
        match inst {
            HirInst::Const { dst, value } => {
                let val = eval_const(value);
                self.set_var(*dst, val);
            }
            HirInst::Assign { dst, value } => {
                let val = match value {
                    HirValue::Use(var) => self.get_var(*var),
                    HirValue::Const(c) => eval_const(c),
                };
                self.set_var(*dst, val);
            }
            HirInst::BinOp {
                dst,
                op,
                left,
                right,
            } => {
                let l = self.get_var(*left);
                let r = self.get_var(*right);
                let result = eval_binop(op, &l, &r)?;
                self.set_var(*dst, result);
            }
            HirInst::UnOp { dst, op, operand } => {
                let v = self.get_var(*operand);
                let result = eval_unop(op, &v)?;
                self.set_var(*dst, result);
            }
            HirInst::Call { dst, func, args } => {
                let arg_vals: Vec<IValue> =
                    args.iter().map(|a| self.get_var(*a)).collect();
                let result = match func {
                    FuncRef::Builtin(name) => self.call_builtin(name, arg_vals)?,
                    FuncRef::Local(idx) => {
                        self.execute_function(module, *idx, arg_vals)?
                    }
                    FuncRef::External(name) => {
                        self.call_builtin(name, arg_vals)?
                    }
                };
                if let Some(d) = dst {
                    self.set_var(*d, result);
                }
            }
            HirInst::MethodCall {
                dst,
                receiver,
                method,
                args,
            } => {
                let recv = self.get_var(*receiver);
                let arg_vals: Vec<IValue> =
                    args.iter().map(|a| self.get_var(*a)).collect();
                let result = self.call_method(&recv, method, arg_vals)?;
                if let Some(d) = dst {
                    self.set_var(*d, result);
                }
            }
            HirInst::GetField { dst, object, field } => {
                let obj = self.get_var(*object);
                let val = get_field(&obj, field);
                self.set_var(*dst, val);
            }
            HirInst::SetField {
                object,
                field,
                value,
            } => {
                let mut obj = self.get_var(*object);
                let val = self.get_var(*value);
                set_field(&mut obj, field, val.clone());
                // Note: in a dynamic language, we'd need to update the variable
                // holding obj. For now this is a best-effort in-place mutation.
            }
            HirInst::GetIndex { dst, object, index } => {
                let obj = self.get_var(*object);
                let idx = self.get_var(*index);
                let val = get_index(&obj, &idx);
                self.set_var(*dst, val);
            }
            HirInst::SetIndex {
                object,
                index,
                value,
            } => {
                // Similar to SetField — best-effort
                let _obj = self.get_var(*object);
                let _idx = self.get_var(*index);
                let _val = self.get_var(*value);
            }
            HirInst::ConstructList { dst, elems } => {
                let items: Vec<IValue> =
                    elems.iter().map(|e| self.get_var(*e)).collect();
                self.set_var(*dst, IValue::List(items));
            }
            HirInst::ConstructMap { dst, entries } => {
                let items: Vec<(String, IValue)> = entries
                    .iter()
                    .map(|(k, v)| {
                        let key = match self.get_var(*k) {
                            IValue::Str(s) => s,
                            other => other.to_string(),
                        };
                        (key, self.get_var(*v))
                    })
                    .collect();
                self.set_var(*dst, IValue::Map(items));
            }
            HirInst::ConstructStruct {
                dst,
                type_name,
                fields,
            } => {
                let field_vals: Vec<(String, IValue)> = fields
                    .iter()
                    .map(|(name, var)| (name.clone(), self.get_var(*var)))
                    .collect();
                self.set_var(
                    *dst,
                    IValue::Struct {
                        type_name: type_name.clone(),
                        fields: field_vals,
                    },
                );
            }
            HirInst::Propagate { dst, inner } => {
                let val = self.get_var(*inner);
                match &val {
                    IValue::Error(msg) => {
                        return Err(format!("propagated error: {msg}"));
                    }
                    _ => self.set_var(*dst, val),
                }
            }
            HirInst::Spawn { dst, .. } => {
                // Simplified: spawn is a no-op in the interpreter
                self.set_var(*dst, IValue::List(vec![]));
            }
            HirInst::ChanRecv { dst, .. } => {
                self.set_var(*dst, IValue::None);
            }
            HirInst::ChanSend { .. } => {}
            HirInst::ToolRegister { .. } => {
                // Tool registration is handled at a higher level
            }
            HirInst::AgentAsk { dst, .. }
            | HirInst::AgentRun { dst, .. }
            | HirInst::AgentStream { dst, .. } => {
                // Agent ops require the full runtime; stub in interpreter
                self.set_var(*dst, IValue::Str("<agent result>".to_string()));
            }
            HirInst::Nop => {}
        }
        Ok(())
    }

    fn call_builtin(
        &mut self,
        name: &str,
        args: Vec<IValue>,
    ) -> Result<IValue, String> {
        match name {
            "io.println" => {
                let text = args
                    .iter()
                    .map(|a| a.to_string())
                    .collect::<Vec<_>>()
                    .join(" ");
                println!("{text}");
                self.stdout.push(text);
                Ok(IValue::None)
            }
            "io.print" => {
                let text = args
                    .iter()
                    .map(|a| a.to_string())
                    .collect::<Vec<_>>()
                    .join(" ");
                print!("{text}");
                self.stdout.push(text);
                Ok(IValue::None)
            }
            "io.eprintln" => {
                let text = args
                    .iter()
                    .map(|a| a.to_string())
                    .collect::<Vec<_>>()
                    .join(" ");
                eprintln!("{text}");
                Ok(IValue::None)
            }
            "len" => {
                let val = args.first().cloned().unwrap_or(IValue::None);
                match &val {
                    IValue::Str(s) => Ok(IValue::Int(s.len() as i64)),
                    IValue::List(v) => Ok(IValue::Int(v.len() as i64)),
                    IValue::Map(m) => Ok(IValue::Int(m.len() as i64)),
                    _ => Ok(IValue::Int(0)),
                }
            }
            "keys" => {
                let val = args.first().cloned().unwrap_or(IValue::None);
                match val {
                    IValue::Map(m) => Ok(IValue::List(
                        m.into_iter()
                            .map(|(k, _)| IValue::Str(k))
                            .collect(),
                    )),
                    _ => Ok(IValue::List(vec![])),
                }
            }
            "join" => {
                let list = args.first().cloned().unwrap_or(IValue::List(vec![]));
                let sep = match args.get(1) {
                    Some(IValue::Str(s)) => s.clone(),
                    _ => String::new(),
                };
                match list {
                    IValue::List(items) => {
                        let joined = items
                            .iter()
                            .map(|v| v.to_string())
                            .collect::<Vec<_>>()
                            .join(&sep);
                        Ok(IValue::Str(joined))
                    }
                    _ => Ok(IValue::Str(String::new())),
                }
            }
            "type_of" => {
                let val = args.first().cloned().unwrap_or(IValue::None);
                let name = match &val {
                    IValue::None => "none",
                    IValue::Bool(_) => "bool",
                    IValue::Int(_) => "int",
                    IValue::Float(_) => "float",
                    IValue::Str(_) => "string",
                    IValue::List(_) => "list",
                    IValue::Map(_) => "map",
                    IValue::Struct { .. } => "struct",
                    IValue::Error(_) => "error",
                };
                Ok(IValue::Str(name.to_string()))
            }
            "to_string" | "conv.int_to_string" | "conv.float_to_string"
            | "conv.bool_to_string" => {
                let val = args.first().cloned().unwrap_or(IValue::None);
                Ok(IValue::Str(val.to_string()))
            }
            "to_int" | "conv.string_to_int" | "conv.float_to_int" => {
                let val = args.first().cloned().unwrap_or(IValue::None);
                match &val {
                    IValue::Int(_) => Ok(val),
                    IValue::Float(f) => Ok(IValue::Int(*f as i64)),
                    IValue::Str(s) => s
                        .parse::<i64>()
                        .map(IValue::Int)
                        .map_err(|_| format!("cannot convert '{s}' to int")),
                    IValue::Bool(b) => Ok(IValue::Int(if *b { 1 } else { 0 })),
                    _ => Err("cannot convert to int".to_string()),
                }
            }
            "to_float" | "conv.string_to_float" | "conv.int_to_float" => {
                let val = args.first().cloned().unwrap_or(IValue::None);
                match &val {
                    IValue::Float(_) => Ok(val),
                    IValue::Int(n) => Ok(IValue::Float(*n as f64)),
                    IValue::Str(s) => s
                        .parse::<f64>()
                        .map(IValue::Float)
                        .map_err(|_| format!("cannot convert '{s}' to float")),
                    IValue::Bool(b) => {
                        Ok(IValue::Float(if *b { 1.0 } else { 0.0 }))
                    }
                    _ => Err("cannot convert to float".to_string()),
                }
            }
            "panic" => {
                let msg = args
                    .first()
                    .map(|a| a.to_string())
                    .unwrap_or_else(|| "panic".to_string());
                Err(format!("panic: {msg}"))
            }
            "env" => {
                let key = match args.first() {
                    Some(IValue::Str(s)) => s.clone(),
                    _ => return Ok(IValue::Str(String::new())),
                };
                let val =
                    std::env::var(&key).unwrap_or_default();
                Ok(IValue::Str(val))
            }
            "string.len" => {
                match args.first() {
                    Some(IValue::Str(s)) => Ok(IValue::Int(s.len() as i64)),
                    _ => Ok(IValue::Int(0)),
                }
            }
            "string.contains" => {
                match (args.first(), args.get(1)) {
                    (Some(IValue::Str(s)), Some(IValue::Str(sub))) => {
                        Ok(IValue::Bool(s.contains(sub.as_str())))
                    }
                    _ => Ok(IValue::Bool(false)),
                }
            }
            "string.split" => {
                match (args.first(), args.get(1)) {
                    (Some(IValue::Str(s)), Some(IValue::Str(sep))) => {
                        let parts: Vec<IValue> = s
                            .split(sep.as_str())
                            .map(|p| IValue::Str(p.to_string()))
                            .collect();
                        Ok(IValue::List(parts))
                    }
                    _ => Ok(IValue::List(vec![])),
                }
            }
            "string.trim" => {
                match args.first() {
                    Some(IValue::Str(s)) => {
                        Ok(IValue::Str(s.trim().to_string()))
                    }
                    _ => Ok(IValue::Str(String::new())),
                }
            }
            "string.to_upper" => {
                match args.first() {
                    Some(IValue::Str(s)) => {
                        Ok(IValue::Str(s.to_uppercase()))
                    }
                    _ => Ok(IValue::Str(String::new())),
                }
            }
            "string.to_lower" => {
                match args.first() {
                    Some(IValue::Str(s)) => {
                        Ok(IValue::Str(s.to_lowercase()))
                    }
                    _ => Ok(IValue::Str(String::new())),
                }
            }
            "array.len" => {
                match args.first() {
                    Some(IValue::List(v)) => Ok(IValue::Int(v.len() as i64)),
                    _ => Ok(IValue::Int(0)),
                }
            }
            "array.push" => {
                match args.first() {
                    Some(IValue::List(v)) => {
                        let mut new = v.clone();
                        if let Some(item) = args.get(1) {
                            new.push(item.clone());
                        }
                        Ok(IValue::List(new))
                    }
                    _ => Ok(IValue::List(vec![])),
                }
            }
            "map.get" => {
                match (args.first(), args.get(1)) {
                    (Some(IValue::Map(m)), Some(IValue::Str(key))) => {
                        let val = m
                            .iter()
                            .find(|(k, _)| k == key)
                            .map(|(_, v)| v.clone())
                            .unwrap_or(IValue::None);
                        Ok(val)
                    }
                    _ => Ok(IValue::None),
                }
            }
            "map.has" => {
                match (args.first(), args.get(1)) {
                    (Some(IValue::Map(m)), Some(IValue::Str(key))) => {
                        let found = m.iter().any(|(k, _)| k == key);
                        Ok(IValue::Bool(found))
                    }
                    _ => Ok(IValue::Bool(false)),
                }
            }
            "map.keys" => {
                match args.first() {
                    Some(IValue::Map(m)) => Ok(IValue::List(
                        m.iter()
                            .map(|(k, _)| IValue::Str(k.clone()))
                            .collect(),
                    )),
                    _ => Ok(IValue::List(vec![])),
                }
            }
            "json.marshal" | "json.marshal_pretty" => {
                let val = args.first().cloned().unwrap_or(IValue::None);
                Ok(IValue::Str(ivalue_to_json(&val)))
            }
            "json.unmarshal" | "json.parse" => {
                // Simplified: return the raw string as-is for now
                match args.first() {
                    Some(IValue::Str(s)) => Ok(IValue::Str(s.clone())),
                    _ => Ok(IValue::None),
                }
            }
            "time.now" => {
                Ok(IValue::Str("2026-01-01T00:00:00Z".to_string()))
            }
            "time.sleep" => {
                // No-op in interpreter for now
                Ok(IValue::None)
            }
            "math.abs" => {
                match args.first() {
                    Some(IValue::Int(n)) => Ok(IValue::Int(n.abs())),
                    Some(IValue::Float(f)) => Ok(IValue::Float(f.abs())),
                    _ => Ok(IValue::Float(0.0)),
                }
            }
            "math.min" => {
                match (args.first(), args.get(1)) {
                    (Some(IValue::Int(a)), Some(IValue::Int(b))) => {
                        Ok(IValue::Int(*a.min(b)))
                    }
                    (Some(IValue::Float(a)), Some(IValue::Float(b))) => {
                        Ok(IValue::Float(a.min(*b)))
                    }
                    _ => Ok(IValue::Float(0.0)),
                }
            }
            "math.max" => {
                match (args.first(), args.get(1)) {
                    (Some(IValue::Int(a)), Some(IValue::Int(b))) => {
                        Ok(IValue::Int(*a.max(b)))
                    }
                    (Some(IValue::Float(a)), Some(IValue::Float(b))) => {
                        Ok(IValue::Float(a.max(*b)))
                    }
                    _ => Ok(IValue::Float(0.0)),
                }
            }
            _ => {
                // Unknown builtin — return None instead of erroring,
                // since many stdlib functions are not yet implemented.
                Ok(IValue::None)
            }
        }
    }

    fn call_method(
        &mut self,
        receiver: &IValue,
        method: &str,
        args: Vec<IValue>,
    ) -> Result<IValue, String> {
        match (receiver, method) {
            (IValue::Str(s), "len") => Ok(IValue::Int(s.len() as i64)),
            (IValue::Str(s), "contains") => {
                match args.first() {
                    Some(IValue::Str(sub)) => {
                        Ok(IValue::Bool(s.contains(sub.as_str())))
                    }
                    _ => Ok(IValue::Bool(false)),
                }
            }
            (IValue::Str(s), "split") => {
                match args.first() {
                    Some(IValue::Str(sep)) => {
                        let parts: Vec<IValue> = s
                            .split(sep.as_str())
                            .map(|p| IValue::Str(p.to_string()))
                            .collect();
                        Ok(IValue::List(parts))
                    }
                    _ => Ok(IValue::List(vec![])),
                }
            }
            (IValue::Str(s), "trim") => {
                Ok(IValue::Str(s.trim().to_string()))
            }
            (IValue::List(v), "len") => Ok(IValue::Int(v.len() as i64)),
            (IValue::List(v), "push") => {
                let mut new = v.clone();
                if let Some(item) = args.first() {
                    new.push(item.clone());
                }
                Ok(IValue::List(new))
            }
            (IValue::Map(m), "keys") => Ok(IValue::List(
                m.iter()
                    .map(|(k, _)| IValue::Str(k.clone()))
                    .collect(),
            )),
            _ => Ok(IValue::None),
        }
    }
}

// ===========================================================================
// Helper functions
// ===========================================================================

fn eval_const(c: &HirConst) -> IValue {
    match c {
        HirConst::None => IValue::None,
        HirConst::Bool(b) => IValue::Bool(*b),
        HirConst::Int(n) => IValue::Int(*n),
        HirConst::Float(f) => IValue::Float(*f),
        HirConst::Str(s) => IValue::Str(s.clone()),
        HirConst::List(items) => {
            IValue::List(items.iter().map(eval_const).collect())
        }
        HirConst::Map(entries) => IValue::Map(
            entries
                .iter()
                .map(|(k, v)| (k.clone(), eval_const(v)))
                .collect(),
        ),
    }
}

fn eval_binop(op: &BinOp, left: &IValue, right: &IValue) -> Result<IValue, String> {
    match op {
        BinOp::Add => match (left, right) {
            (IValue::Int(a), IValue::Int(b)) => Ok(IValue::Int(a + b)),
            (IValue::Float(a), IValue::Float(b)) => Ok(IValue::Float(a + b)),
            (IValue::Int(a), IValue::Float(b)) => Ok(IValue::Float(*a as f64 + b)),
            (IValue::Float(a), IValue::Int(b)) => Ok(IValue::Float(a + *b as f64)),
            (IValue::Str(a), IValue::Str(b)) => {
                Ok(IValue::Str(format!("{a}{b}")))
            }
            (IValue::Str(a), b) => Ok(IValue::Str(format!("{a}{b}"))),
            (a, IValue::Str(b)) => Ok(IValue::Str(format!("{a}{b}"))),
            (IValue::List(a), IValue::List(b)) => {
                let mut v = a.clone();
                v.extend(b.iter().cloned());
                Ok(IValue::List(v))
            }
            _ => Err(format!("cannot add {} and {}", left, right)),
        },
        BinOp::Sub => match (left, right) {
            (IValue::Int(a), IValue::Int(b)) => Ok(IValue::Int(a - b)),
            (IValue::Float(a), IValue::Float(b)) => Ok(IValue::Float(a - b)),
            (IValue::Int(a), IValue::Float(b)) => Ok(IValue::Float(*a as f64 - b)),
            (IValue::Float(a), IValue::Int(b)) => Ok(IValue::Float(a - *b as f64)),
            _ => Err(format!("cannot subtract {} from {}", right, left)),
        },
        BinOp::Mul => match (left, right) {
            (IValue::Int(a), IValue::Int(b)) => Ok(IValue::Int(a * b)),
            (IValue::Float(a), IValue::Float(b)) => Ok(IValue::Float(a * b)),
            (IValue::Int(a), IValue::Float(b)) => Ok(IValue::Float(*a as f64 * b)),
            (IValue::Float(a), IValue::Int(b)) => Ok(IValue::Float(a * *b as f64)),
            _ => Err(format!("cannot multiply {} and {}", left, right)),
        },
        BinOp::Div => match (left, right) {
            (IValue::Int(_), IValue::Int(0)) => Err("division by zero".to_string()),
            (IValue::Int(a), IValue::Int(b)) => Ok(IValue::Int(a / b)),
            (IValue::Float(a), IValue::Float(b)) => {
                if *b == 0.0 {
                    Err("division by zero".to_string())
                } else {
                    Ok(IValue::Float(a / b))
                }
            }
            (IValue::Int(a), IValue::Float(b)) => {
                if *b == 0.0 {
                    Err("division by zero".to_string())
                } else {
                    Ok(IValue::Float(*a as f64 / b))
                }
            }
            (IValue::Float(a), IValue::Int(b)) => {
                if *b == 0 {
                    Err("division by zero".to_string())
                } else {
                    Ok(IValue::Float(a / *b as f64))
                }
            }
            _ => Err(format!("cannot divide {} by {}", left, right)),
        },
        BinOp::Mod => match (left, right) {
            (IValue::Int(_), IValue::Int(0)) => Err("modulo by zero".to_string()),
            (IValue::Int(a), IValue::Int(b)) => Ok(IValue::Int(a % b)),
            _ => Err(format!("cannot modulo {} by {}", left, right)),
        },
        BinOp::Eq => Ok(IValue::Bool(left == right)),
        BinOp::Ne => Ok(IValue::Bool(left != right)),
        BinOp::Lt => match (left, right) {
            (IValue::Int(a), IValue::Int(b)) => Ok(IValue::Bool(a < b)),
            (IValue::Float(a), IValue::Float(b)) => Ok(IValue::Bool(a < b)),
            (IValue::Int(a), IValue::Float(b)) => Ok(IValue::Bool((*a as f64) < *b)),
            (IValue::Float(a), IValue::Int(b)) => Ok(IValue::Bool(*a < *b as f64)),
            (IValue::Str(a), IValue::Str(b)) => Ok(IValue::Bool(a < b)),
            _ => Ok(IValue::Bool(false)),
        },
        BinOp::Gt => match (left, right) {
            (IValue::Int(a), IValue::Int(b)) => Ok(IValue::Bool(a > b)),
            (IValue::Float(a), IValue::Float(b)) => Ok(IValue::Bool(a > b)),
            (IValue::Int(a), IValue::Float(b)) => Ok(IValue::Bool((*a as f64) > *b)),
            (IValue::Float(a), IValue::Int(b)) => Ok(IValue::Bool(*a > *b as f64)),
            (IValue::Str(a), IValue::Str(b)) => Ok(IValue::Bool(a > b)),
            _ => Ok(IValue::Bool(false)),
        },
        BinOp::Le => match (left, right) {
            (IValue::Int(a), IValue::Int(b)) => Ok(IValue::Bool(a <= b)),
            (IValue::Float(a), IValue::Float(b)) => Ok(IValue::Bool(a <= b)),
            (IValue::Int(a), IValue::Float(b)) => Ok(IValue::Bool((*a as f64) <= *b)),
            (IValue::Float(a), IValue::Int(b)) => Ok(IValue::Bool(*a <= *b as f64)),
            _ => Ok(IValue::Bool(false)),
        },
        BinOp::Ge => match (left, right) {
            (IValue::Int(a), IValue::Int(b)) => Ok(IValue::Bool(a >= b)),
            (IValue::Float(a), IValue::Float(b)) => Ok(IValue::Bool(a >= b)),
            (IValue::Int(a), IValue::Float(b)) => Ok(IValue::Bool((*a as f64) >= *b)),
            (IValue::Float(a), IValue::Int(b)) => Ok(IValue::Bool(*a >= *b as f64)),
            _ => Ok(IValue::Bool(false)),
        },
        BinOp::And => Ok(IValue::Bool(left.is_truthy() && right.is_truthy())),
        BinOp::Or => Ok(IValue::Bool(left.is_truthy() || right.is_truthy())),
        BinOp::BitAnd => match (left, right) {
            (IValue::Int(a), IValue::Int(b)) => Ok(IValue::Int(a & b)),
            _ => Err("bitwise AND requires integers".to_string()),
        },
        BinOp::BitOr => match (left, right) {
            (IValue::Int(a), IValue::Int(b)) => Ok(IValue::Int(a | b)),
            _ => Err("bitwise OR requires integers".to_string()),
        },
        BinOp::BitXor => match (left, right) {
            (IValue::Int(a), IValue::Int(b)) => Ok(IValue::Int(a ^ b)),
            _ => Err("bitwise XOR requires integers".to_string()),
        },
        BinOp::Shl => match (left, right) {
            (IValue::Int(a), IValue::Int(b)) => {
                if *b < 0 || *b > 63 {
                    Err("shift amount out of range".to_string())
                } else {
                    Ok(IValue::Int(a << b))
                }
            }
            _ => Err("shift requires integers".to_string()),
        },
        BinOp::Shr => match (left, right) {
            (IValue::Int(a), IValue::Int(b)) => {
                if *b < 0 || *b > 63 {
                    Err("shift amount out of range".to_string())
                } else {
                    Ok(IValue::Int(a >> b))
                }
            }
            _ => Err("shift requires integers".to_string()),
        },
    }
}

fn eval_unop(op: &UnOp, operand: &IValue) -> Result<IValue, String> {
    match op {
        UnOp::Neg => match operand {
            IValue::Int(n) => Ok(IValue::Int(-n)),
            IValue::Float(f) => Ok(IValue::Float(-f)),
            _ => Err(format!("cannot negate {operand}")),
        },
        UnOp::Not => Ok(IValue::Bool(!operand.is_truthy())),
        UnOp::BitNot => match operand {
            IValue::Int(n) => Ok(IValue::Int(!n)),
            _ => Err("bitwise NOT requires integer".to_string()),
        },
    }
}

fn values_match(val: &IValue, pattern: &HirConst) -> bool {
    match (val, pattern) {
        (IValue::None, HirConst::None) => true,
        (IValue::Bool(a), HirConst::Bool(b)) => a == b,
        (IValue::Int(a), HirConst::Int(b)) => a == b,
        (IValue::Float(a), HirConst::Float(b)) => a == b,
        (IValue::Str(a), HirConst::Str(b)) => a == b,
        _ => false,
    }
}

fn get_field(obj: &IValue, field: &str) -> IValue {
    match obj {
        IValue::Struct { fields, .. } => fields
            .iter()
            .find(|(k, _)| k == field)
            .map(|(_, v)| v.clone())
            .unwrap_or(IValue::None),
        IValue::Map(m) => m
            .iter()
            .find(|(k, _)| k == field)
            .map(|(_, v)| v.clone())
            .unwrap_or(IValue::None),
        _ => IValue::None,
    }
}

fn set_field(obj: &mut IValue, field: &str, val: IValue) {
    match obj {
        IValue::Struct { fields, .. } => {
            if let Some(entry) = fields.iter_mut().find(|(k, _)| k == field) {
                entry.1 = val;
            } else {
                fields.push((field.to_string(), val));
            }
        }
        IValue::Map(m) => {
            if let Some(entry) = m.iter_mut().find(|(k, _)| k == field) {
                entry.1 = val;
            } else {
                m.push((field.to_string(), val));
            }
        }
        _ => {}
    }
}

fn get_index(obj: &IValue, idx: &IValue) -> IValue {
    match (obj, idx) {
        (IValue::List(v), IValue::Int(i)) => {
            let i = *i as usize;
            v.get(i).cloned().unwrap_or(IValue::None)
        }
        (IValue::Map(m), IValue::Str(key)) => m
            .iter()
            .find(|(k, _)| k == key)
            .map(|(_, v)| v.clone())
            .unwrap_or(IValue::None),
        (IValue::Str(s), IValue::Int(i)) => {
            let i = *i as usize;
            s.chars()
                .nth(i)
                .map(|c| IValue::Str(c.to_string()))
                .unwrap_or(IValue::None)
        }
        _ => IValue::None,
    }
}

fn ivalue_to_json(val: &IValue) -> String {
    match val {
        IValue::None => "null".to_string(),
        IValue::Bool(b) => b.to_string(),
        IValue::Int(n) => n.to_string(),
        IValue::Float(f) => f.to_string(),
        IValue::Str(s) => format!("\"{}\"", s.replace('\\', "\\\\").replace('"', "\\\"")),
        IValue::List(items) => {
            let inner: Vec<String> = items.iter().map(ivalue_to_json).collect();
            format!("[{}]", inner.join(","))
        }
        IValue::Map(entries) => {
            let inner: Vec<String> = entries
                .iter()
                .map(|(k, v)| format!("\"{}\":{}", k, ivalue_to_json(v)))
                .collect();
            format!("{{{}}}", inner.join(","))
        }
        IValue::Struct { fields, .. } => {
            let inner: Vec<String> = fields
                .iter()
                .map(|(k, v)| format!("\"{}\":{}", k, ivalue_to_json(v)))
                .collect();
            format!("{{{}}}", inner.join(","))
        }
        IValue::Error(msg) => format!("\"error: {}\"", msg),
    }
}

/// Interpret a HIR module directly (tree-walking interpreter).
fn interpret(module: &HirModule) -> Result<i32, Vec<Diagnostic>> {
    let mut interp = Interpreter::new();
    interp.execute(module)
}

/// Result of running a single test.
#[derive(Debug)]
pub struct TestResult {
    /// Human-readable test name (without `__test_` prefix).
    pub name: String,
    /// `None` if passed, `Some(message)` if failed.
    pub failure: Option<String>,
}

/// Discover and run all `__test_*` functions in the module via the interpreter.
///
/// Returns a list of test results (one per test function found).
pub fn run_tests(module: &HirModule) -> Vec<TestResult> {
    let test_funcs: Vec<(usize, String)> = module
        .functions
        .iter()
        .enumerate()
        .filter(|(_, f)| f.name.starts_with("__test_"))
        .map(|(i, f)| {
            let display_name = f.name.strip_prefix("__test_").unwrap_or(&f.name).to_string();
            (i, display_name)
        })
        .collect();

    let mut results = Vec::new();
    for (idx, name) in test_funcs {
        let mut interp = Interpreter::new();
        match interp.execute_function(module, idx, vec![]) {
            Ok(_) => results.push(TestResult {
                name,
                failure: None,
            }),
            Err(msg) => results.push(TestResult {
                name,
                failure: Some(msg),
            }),
        }
    }
    results
}

// ===========================================================================
// Tests
// ===========================================================================

#[cfg(test)]
mod tests {
    use super::*;

    fn make_empty_module() -> HirModule {
        HirModule {
            functions: vec![],
            declarations: vec![],
            entry: None,
        }
    }

    fn make_simple_module(insts: Vec<HirInst>, terminator: Terminator) -> HirModule {
        HirModule {
            functions: vec![HirFunction {
                name: "__main".to_string(),
                params: vec![],
                return_ty: HirType::Void,
                blocks: vec![HirBlock {
                    id: 0,
                    insts,
                    terminator,
                }],
                is_workflow: false,
                is_tool: false,
                is_async: false,
                span: Span::default(),
            }],
            declarations: vec![],
            entry: Some(0),
        }
    }

    #[test]
    fn default_options() {
        let opts = CodegenOptions::default();
        assert_eq!(opts.target, Target::Interpreter);
        assert!(!opts.optimize);
    }

    #[test]
    fn empty_module_exits_zero() {
        let module = make_empty_module();
        let result =
            codegen(&module, &CodegenOptions::default()).unwrap();
        match result {
            CodegenResult::Executed { exit_code } => assert_eq!(exit_code, 0),
            _ => panic!("expected Executed"),
        }
    }

    #[test]
    fn simple_const_and_return() {
        let module = make_simple_module(
            vec![HirInst::Const {
                dst: 0,
                value: HirConst::Int(42),
            }],
            Terminator::Return(Some(0)),
        );
        let result =
            codegen(&module, &CodegenOptions::default()).unwrap();
        match result {
            CodegenResult::Executed { exit_code } => assert_eq!(exit_code, 0),
            _ => panic!("expected Executed"),
        }
    }

    #[test]
    fn binary_operation() {
        let module = make_simple_module(
            vec![
                HirInst::Const {
                    dst: 0,
                    value: HirConst::Int(10),
                },
                HirInst::Const {
                    dst: 1,
                    value: HirConst::Int(5),
                },
                HirInst::BinOp {
                    dst: 2,
                    op: BinOp::Add,
                    left: 0,
                    right: 1,
                },
            ],
            Terminator::Return(Some(2)),
        );
        let mut interp = Interpreter::new();
        let result = interp.execute_function(&module, 0, vec![]).unwrap();
        assert_eq!(result, IValue::Int(15));
    }

    #[test]
    fn branch_execution() {
        let module = HirModule {
            functions: vec![HirFunction {
                name: "__main".to_string(),
                params: vec![],
                return_ty: HirType::Int,
                blocks: vec![
                    // Block 0: condition
                    HirBlock {
                        id: 0,
                        insts: vec![HirInst::Const {
                            dst: 0,
                            value: HirConst::Bool(true),
                        }],
                        terminator: Terminator::Branch {
                            cond: 0,
                            then_block: 1,
                            else_block: 2,
                        },
                    },
                    // Block 1: then
                    HirBlock {
                        id: 1,
                        insts: vec![HirInst::Const {
                            dst: 1,
                            value: HirConst::Int(1),
                        }],
                        terminator: Terminator::Return(Some(1)),
                    },
                    // Block 2: else
                    HirBlock {
                        id: 2,
                        insts: vec![HirInst::Const {
                            dst: 1,
                            value: HirConst::Int(0),
                        }],
                        terminator: Terminator::Return(Some(1)),
                    },
                ],
                is_workflow: false,
                is_tool: false,
                is_async: false,
                span: Span::default(),
            }],
            declarations: vec![],
            entry: Some(0),
        };
        let mut interp = Interpreter::new();
        let result = interp.execute_function(&module, 0, vec![]).unwrap();
        assert_eq!(result, IValue::Int(1));
    }

    #[test]
    fn builtin_call() {
        let module = make_simple_module(
            vec![
                HirInst::Const {
                    dst: 0,
                    value: HirConst::Str("hello".to_string()),
                },
                HirInst::Call {
                    dst: Some(1),
                    func: FuncRef::Builtin("len".to_string()),
                    args: vec![0],
                },
            ],
            Terminator::Return(Some(1)),
        );
        let mut interp = Interpreter::new();
        let result = interp.execute_function(&module, 0, vec![]).unwrap();
        assert_eq!(result, IValue::Int(5));
    }

    #[test]
    fn native_target_not_implemented() {
        let module = make_empty_module();
        let opts = CodegenOptions {
            target: Target::Native,
            optimize: false,
        };
        let err = codegen(&module, &opts).unwrap_err();
        assert!(!err.is_empty());
        assert!(err[0].message.contains("not yet implemented"));
    }

    #[test]
    fn unary_negation() {
        let result = eval_unop(&UnOp::Neg, &IValue::Int(5)).unwrap();
        assert_eq!(result, IValue::Int(-5));
    }

    #[test]
    fn division_by_zero() {
        let result = eval_binop(&BinOp::Div, &IValue::Int(10), &IValue::Int(0));
        assert!(result.is_err());
        assert!(result.unwrap_err().contains("division by zero"));
    }

    #[test]
    fn string_concatenation() {
        let result = eval_binop(
            &BinOp::Add,
            &IValue::Str("hello ".to_string()),
            &IValue::Str("world".to_string()),
        )
        .unwrap();
        assert_eq!(result, IValue::Str("hello world".to_string()));
    }

    #[test]
    fn ivalue_display() {
        assert_eq!(IValue::None.to_string(), "none");
        assert_eq!(IValue::Bool(true).to_string(), "true");
        assert_eq!(IValue::Int(42).to_string(), "42");
        assert_eq!(IValue::Str("hello".to_string()).to_string(), "hello");
        assert_eq!(
            IValue::List(vec![IValue::Int(1), IValue::Int(2)]).to_string(),
            "[1, 2]"
        );
    }

    #[test]
    fn construct_list() {
        let module = make_simple_module(
            vec![
                HirInst::Const {
                    dst: 0,
                    value: HirConst::Int(1),
                },
                HirInst::Const {
                    dst: 1,
                    value: HirConst::Int(2),
                },
                HirInst::ConstructList {
                    dst: 2,
                    elems: vec![0, 1],
                },
            ],
            Terminator::Return(Some(2)),
        );
        let mut interp = Interpreter::new();
        let result = interp.execute_function(&module, 0, vec![]).unwrap();
        assert_eq!(
            result,
            IValue::List(vec![IValue::Int(1), IValue::Int(2)])
        );
    }
}
