//! Cranelift-based compiler for Haira.

use cranelift::prelude::*;
use cranelift_module::{DataDescription, FuncId, Linkage, Module};
use cranelift_object::{ObjectBuilder, ObjectModule};
use haira_ast::{
    BinaryOp, Block, Expr, ExprKind, Item, ItemKind, Literal, MethodDef, SourceFile, Statement,
    StatementKind, TypeDef, UnaryOp,
};
use smol_str::SmolStr;
use std::collections::HashMap;
use std::path::Path;
use std::process::Command;

/// Information about a struct type.
#[derive(Debug, Clone)]
struct StructInfo {
    /// Field names in order.
    fields: Vec<SmolStr>,
    /// Size of each field in bytes (all i64 for now).
    field_offsets: Vec<usize>,
    /// Total size of the struct in bytes.
    size: usize,
}

/// Code generation options.
#[derive(Default, Clone)]
pub struct CodegenOptions {
    /// Optimization level (0-3).
    pub opt_level: u8,
    /// Generate debug info.
    pub debug_info: bool,
    /// Target triple (e.g., "x86_64-unknown-linux-gnu").
    pub target: Option<String>,
}

/// Code generation error.
#[derive(Debug, thiserror::Error)]
pub enum CodegenError {
    #[error("Cranelift error: {0}")]
    CraneliftError(String),
    #[error("Module error: {0}")]
    ModuleError(#[from] cranelift_module::ModuleError),
    #[error("IO error: {0}")]
    IoError(#[from] std::io::Error),
    #[error("Linker error: {0}")]
    LinkerError(String),
    #[error("Unsupported feature: {0}")]
    Unsupported(String),
    #[error("Undefined function: {0}")]
    UndefinedFunction(String),
    #[error("Undefined variable: {0}")]
    UndefinedVariable(String),
}

/// Haira compiler using Cranelift.
pub struct Compiler {
    /// The Cranelift module.
    module: ObjectModule,
    /// Builder context (reused).
    builder_ctx: FunctionBuilderContext,
    /// Cranelift context (reused).
    ctx: codegen::Context,
    /// Map of function names to their IDs.
    functions: HashMap<SmolStr, FuncId>,
    /// Map of string constants to their data IDs.
    strings: HashMap<SmolStr, cranelift_module::DataId>,
    /// Map of struct type names to their info.
    structs: HashMap<SmolStr, StructInfo>,
    /// Pointer type for the target.
    ptr_type: Type,
    /// Counter for generating unique spawn function names.
    spawn_counter: usize,
    /// Map of spawn block span start to their function names.
    spawn_functions: HashMap<u32, SmolStr>,
    /// Collected spawn blocks from AST (span start -> block).
    spawn_blocks: Vec<(u32, Block)>,
    /// Counter for generating unique async function names.
    async_counter: usize,
    /// Map of async block span start to their function names (one per statement in block).
    async_functions: HashMap<u32, Vec<SmolStr>>,
    /// Collected async blocks from AST (span start -> block).
    async_blocks: Vec<(u32, Block)>,
}

impl Compiler {
    /// Create a new compiler.
    pub fn new() -> Result<Self, CodegenError> {
        let mut flag_builder = settings::builder();
        flag_builder.set("opt_level", "speed").unwrap();
        flag_builder.set("is_pic", "true").unwrap();

        let isa_builder =
            cranelift_native::builder().map_err(|e| CodegenError::CraneliftError(e.to_string()))?;
        let isa = isa_builder
            .finish(settings::Flags::new(flag_builder))
            .map_err(|e| CodegenError::CraneliftError(e.to_string()))?;

        let ptr_type = isa.pointer_type();

        let builder = ObjectBuilder::new(
            isa,
            "haira_module",
            cranelift_module::default_libcall_names(),
        )
        .map_err(|e| CodegenError::CraneliftError(e.to_string()))?;

        let module = ObjectModule::new(builder);

        Ok(Self {
            module,
            builder_ctx: FunctionBuilderContext::new(),
            ctx: codegen::Context::new(),
            functions: HashMap::new(),
            strings: HashMap::new(),
            structs: HashMap::new(),
            ptr_type,
            spawn_counter: 0,
            spawn_functions: HashMap::new(),
            spawn_blocks: Vec::new(),
            async_counter: 0,
            async_functions: HashMap::new(),
            async_blocks: Vec::new(),
        })
    }

    /// Declare external runtime functions.
    fn declare_runtime_functions(&mut self) -> Result<(), CodegenError> {
        // haira_print(ptr, len)
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(self.ptr_type)); // string ptr
        sig.params.push(AbiParam::new(types::I64)); // string len
        let print_id = self
            .module
            .declare_function("haira_print", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("print"), print_id);

        // haira_print_int(i64)
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(types::I64));
        let print_int_id =
            self.module
                .declare_function("haira_print_int", Linkage::Import, &sig)?;
        self.functions
            .insert(SmolStr::from("print_int"), print_int_id);

        // haira_print_float(f64)
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(types::F64));
        let print_float_id =
            self.module
                .declare_function("haira_print_float", Linkage::Import, &sig)?;
        self.functions
            .insert(SmolStr::from("print_float"), print_float_id);

        // haira_print_bool(i8)
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(types::I8));
        let print_bool_id =
            self.module
                .declare_function("haira_print_bool", Linkage::Import, &sig)?;
        self.functions
            .insert(SmolStr::from("print_bool"), print_bool_id);

        // haira_println() - print newline
        let sig = self.module.make_signature();
        let println_id = self
            .module
            .declare_function("haira_println", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("println"), println_id);

        // haira_alloc(size) -> ptr - allocate memory
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(types::I64)); // size
        sig.returns.push(AbiParam::new(self.ptr_type)); // pointer
        let alloc_id = self
            .module
            .declare_function("haira_alloc", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("alloc"), alloc_id);

        // haira_free(ptr) - free memory
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(self.ptr_type)); // pointer
        let free_id = self
            .module
            .declare_function("haira_free", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("free"), free_id);

        // haira_string_concat(a_ptr, a_len, b_ptr, b_len) -> HairaString*
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(self.ptr_type)); // a ptr
        sig.params.push(AbiParam::new(types::I64)); // a len
        sig.params.push(AbiParam::new(self.ptr_type)); // b ptr
        sig.params.push(AbiParam::new(types::I64)); // b len
        sig.returns.push(AbiParam::new(self.ptr_type)); // result HairaString*
        let concat_id =
            self.module
                .declare_function("haira_string_concat", Linkage::Import, &sig)?;
        self.functions
            .insert(SmolStr::from("string_concat"), concat_id);

        // haira_int_to_string(value) -> HairaString*
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(types::I64)); // value
        sig.returns.push(AbiParam::new(self.ptr_type)); // result HairaString*
        let int_to_string_id =
            self.module
                .declare_function("haira_int_to_string", Linkage::Import, &sig)?;
        self.functions
            .insert(SmolStr::from("int_to_string"), int_to_string_id);

        // haira_float_to_string(value) -> HairaString*
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(types::F64)); // value
        sig.returns.push(AbiParam::new(self.ptr_type)); // result HairaString*
        let float_to_string_id =
            self.module
                .declare_function("haira_float_to_string", Linkage::Import, &sig)?;
        self.functions
            .insert(SmolStr::from("float_to_string"), float_to_string_id);

        // haira_set_error(error)
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(types::I64)); // error value
        let set_error_id =
            self.module
                .declare_function("haira_set_error", Linkage::Import, &sig)?;
        self.functions
            .insert(SmolStr::from("set_error"), set_error_id);

        // haira_get_error() -> i64
        let mut sig = self.module.make_signature();
        sig.returns.push(AbiParam::new(types::I64));
        let get_error_id =
            self.module
                .declare_function("haira_get_error", Linkage::Import, &sig)?;
        self.functions
            .insert(SmolStr::from("get_error"), get_error_id);

        // haira_has_error() -> i64
        let mut sig = self.module.make_signature();
        sig.returns.push(AbiParam::new(types::I64));
        let has_error_id =
            self.module
                .declare_function("haira_has_error", Linkage::Import, &sig)?;
        self.functions
            .insert(SmolStr::from("has_error"), has_error_id);

        // haira_clear_error()
        let sig = self.module.make_signature();
        let clear_error_id =
            self.module
                .declare_function("haira_clear_error", Linkage::Import, &sig)?;
        self.functions
            .insert(SmolStr::from("clear_error"), clear_error_id);

        // haira_sleep(ms: i64)
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(types::I64));
        let sleep_id = self
            .module
            .declare_function("haira_sleep", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("sleep"), sleep_id);

        // haira_channel_new(capacity: i64) -> ptr
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(types::I64));
        sig.returns.push(AbiParam::new(self.ptr_type));
        let channel_new_id =
            self.module
                .declare_function("haira_channel_new", Linkage::Import, &sig)?;
        self.functions
            .insert(SmolStr::from("channel_new"), channel_new_id);

        // haira_channel_send(ch: ptr, value: i64)
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(self.ptr_type));
        sig.params.push(AbiParam::new(types::I64));
        let channel_send_id =
            self.module
                .declare_function("haira_channel_send", Linkage::Import, &sig)?;
        self.functions
            .insert(SmolStr::from("channel_send"), channel_send_id);

        // haira_channel_receive(ch: ptr) -> i64
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(self.ptr_type));
        sig.returns.push(AbiParam::new(types::I64));
        let channel_receive_id =
            self.module
                .declare_function("haira_channel_receive", Linkage::Import, &sig)?;
        self.functions
            .insert(SmolStr::from("channel_receive"), channel_receive_id);

        // haira_channel_close(ch: ptr)
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(self.ptr_type));
        let channel_close_id =
            self.module
                .declare_function("haira_channel_close", Linkage::Import, &sig)?;
        self.functions
            .insert(SmolStr::from("channel_close"), channel_close_id);

        // haira_spawn(func: ptr) -> i64
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(self.ptr_type)); // function pointer
        sig.returns.push(AbiParam::new(types::I64)); // thread handle
        let spawn_id = self
            .module
            .declare_function("haira_spawn", Linkage::Import, &sig)?;
        self.functions
            .insert(SmolStr::from("spawn_thread"), spawn_id);

        // haira_spawn_joinable(func: ptr) -> i64 (for async blocks)
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(self.ptr_type)); // function pointer
        sig.returns.push(AbiParam::new(types::I64)); // thread handle
        let spawn_joinable_id =
            self.module
                .declare_function("haira_spawn_joinable", Linkage::Import, &sig)?;
        self.functions
            .insert(SmolStr::from("spawn_joinable"), spawn_joinable_id);

        // haira_thread_join(handle: i64)
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(types::I64)); // thread handle
        let thread_join_id =
            self.module
                .declare_function("haira_thread_join", Linkage::Import, &sig)?;
        self.functions
            .insert(SmolStr::from("thread_join"), thread_join_id);

        // ====================================================================
        // Standard Library - String Functions
        // ====================================================================

        // haira_string_len(ptr, len) -> i64
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(self.ptr_type));
        sig.params.push(AbiParam::new(types::I64));
        sig.returns.push(AbiParam::new(types::I64));
        let id = self
            .module
            .declare_function("haira_string_len", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("len"), id);

        // haira_string_is_empty(ptr, len) -> i64
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(self.ptr_type));
        sig.params.push(AbiParam::new(types::I64));
        sig.returns.push(AbiParam::new(types::I64));
        let id = self
            .module
            .declare_function("haira_string_is_empty", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("is_empty"), id);

        // haira_string_upper(ptr, len) -> HairaString*
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(self.ptr_type));
        sig.params.push(AbiParam::new(types::I64));
        sig.returns.push(AbiParam::new(self.ptr_type));
        let id = self
            .module
            .declare_function("haira_string_upper", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("upper"), id);

        // haira_string_lower(ptr, len) -> HairaString*
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(self.ptr_type));
        sig.params.push(AbiParam::new(types::I64));
        sig.returns.push(AbiParam::new(self.ptr_type));
        let id = self
            .module
            .declare_function("haira_string_lower", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("lower"), id);

        // haira_string_trim(ptr, len) -> HairaString*
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(self.ptr_type));
        sig.params.push(AbiParam::new(types::I64));
        sig.returns.push(AbiParam::new(self.ptr_type));
        let id = self
            .module
            .declare_function("haira_string_trim", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("trim"), id);

        // haira_string_slice(ptr, len, start, end) -> HairaString*
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(self.ptr_type));
        sig.params.push(AbiParam::new(types::I64));
        sig.params.push(AbiParam::new(types::I64));
        sig.params.push(AbiParam::new(types::I64));
        sig.returns.push(AbiParam::new(self.ptr_type));
        let id = self
            .module
            .declare_function("haira_string_slice", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("slice"), id);

        // haira_string_contains(ptr, len, needle_ptr, needle_len) -> i64
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(self.ptr_type));
        sig.params.push(AbiParam::new(types::I64));
        sig.params.push(AbiParam::new(self.ptr_type));
        sig.params.push(AbiParam::new(types::I64));
        sig.returns.push(AbiParam::new(types::I64));
        let id = self
            .module
            .declare_function("haira_string_contains", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("contains"), id);

        // haira_string_starts_with(ptr, len, prefix_ptr, prefix_len) -> i64
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(self.ptr_type));
        sig.params.push(AbiParam::new(types::I64));
        sig.params.push(AbiParam::new(self.ptr_type));
        sig.params.push(AbiParam::new(types::I64));
        sig.returns.push(AbiParam::new(types::I64));
        let id = self
            .module
            .declare_function("haira_string_starts_with", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("starts_with"), id);

        // haira_string_ends_with(ptr, len, suffix_ptr, suffix_len) -> i64
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(self.ptr_type));
        sig.params.push(AbiParam::new(types::I64));
        sig.params.push(AbiParam::new(self.ptr_type));
        sig.params.push(AbiParam::new(types::I64));
        sig.returns.push(AbiParam::new(types::I64));
        let id = self
            .module
            .declare_function("haira_string_ends_with", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("ends_with"), id);

        // haira_string_index_of(ptr, len, needle_ptr, needle_len) -> i64
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(self.ptr_type));
        sig.params.push(AbiParam::new(types::I64));
        sig.params.push(AbiParam::new(self.ptr_type));
        sig.params.push(AbiParam::new(types::I64));
        sig.returns.push(AbiParam::new(types::I64));
        let id = self
            .module
            .declare_function("haira_string_index_of", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("index_of"), id);

        // haira_string_replace(ptr, len, old_ptr, old_len, new_ptr, new_len) -> HairaString*
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(self.ptr_type));
        sig.params.push(AbiParam::new(types::I64));
        sig.params.push(AbiParam::new(self.ptr_type));
        sig.params.push(AbiParam::new(types::I64));
        sig.params.push(AbiParam::new(self.ptr_type));
        sig.params.push(AbiParam::new(types::I64));
        sig.returns.push(AbiParam::new(self.ptr_type));
        let id = self
            .module
            .declare_function("haira_string_replace", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("replace"), id);

        // haira_string_repeat(ptr, len, n) -> HairaString*
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(self.ptr_type));
        sig.params.push(AbiParam::new(types::I64));
        sig.params.push(AbiParam::new(types::I64));
        sig.returns.push(AbiParam::new(self.ptr_type));
        let id = self
            .module
            .declare_function("haira_string_repeat", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("repeat"), id);

        // haira_string_reverse(ptr, len) -> HairaString*
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(self.ptr_type));
        sig.params.push(AbiParam::new(types::I64));
        sig.returns.push(AbiParam::new(self.ptr_type));
        let id = self
            .module
            .declare_function("haira_string_reverse", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("reverse"), id);

        // haira_string_char_at(ptr, len, index) -> i64
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(self.ptr_type));
        sig.params.push(AbiParam::new(types::I64));
        sig.params.push(AbiParam::new(types::I64));
        sig.returns.push(AbiParam::new(types::I64));
        let id = self
            .module
            .declare_function("haira_string_char_at", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("char_at"), id);

        // ====================================================================
        // Standard Library - Math Functions
        // ====================================================================

        // haira_abs(x) -> i64
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(types::I64));
        sig.returns.push(AbiParam::new(types::I64));
        let id = self
            .module
            .declare_function("haira_abs", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("abs"), id);

        // haira_min(a, b) -> i64
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(types::I64));
        sig.params.push(AbiParam::new(types::I64));
        sig.returns.push(AbiParam::new(types::I64));
        let id = self
            .module
            .declare_function("haira_min", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("min"), id);

        // haira_max(a, b) -> i64
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(types::I64));
        sig.params.push(AbiParam::new(types::I64));
        sig.returns.push(AbiParam::new(types::I64));
        let id = self
            .module
            .declare_function("haira_max", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("max"), id);

        // haira_clamp(x, min, max) -> i64
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(types::I64));
        sig.params.push(AbiParam::new(types::I64));
        sig.params.push(AbiParam::new(types::I64));
        sig.returns.push(AbiParam::new(types::I64));
        let id = self
            .module
            .declare_function("haira_clamp", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("clamp"), id);

        // haira_floor(x) -> f64
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(types::F64));
        sig.returns.push(AbiParam::new(types::F64));
        let id = self
            .module
            .declare_function("haira_floor", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("floor"), id);

        // haira_ceil(x) -> f64
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(types::F64));
        sig.returns.push(AbiParam::new(types::F64));
        let id = self
            .module
            .declare_function("haira_ceil", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("ceil"), id);

        // haira_round(x) -> f64
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(types::F64));
        sig.returns.push(AbiParam::new(types::F64));
        let id = self
            .module
            .declare_function("haira_round", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("round"), id);

        // haira_pow(base, exp) -> f64
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(types::F64));
        sig.params.push(AbiParam::new(types::F64));
        sig.returns.push(AbiParam::new(types::F64));
        let id = self
            .module
            .declare_function("haira_pow", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("pow"), id);

        // haira_sqrt(x) -> f64
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(types::F64));
        sig.returns.push(AbiParam::new(types::F64));
        let id = self
            .module
            .declare_function("haira_sqrt", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("sqrt"), id);

        // haira_log(x) -> f64
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(types::F64));
        sig.returns.push(AbiParam::new(types::F64));
        let id = self
            .module
            .declare_function("haira_log", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("log"), id);

        // haira_log10(x) -> f64
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(types::F64));
        sig.returns.push(AbiParam::new(types::F64));
        let id = self
            .module
            .declare_function("haira_log10", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("log10"), id);

        // haira_exp(x) -> f64
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(types::F64));
        sig.returns.push(AbiParam::new(types::F64));
        let id = self
            .module
            .declare_function("haira_exp", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("exp"), id);

        // haira_sin(x) -> f64
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(types::F64));
        sig.returns.push(AbiParam::new(types::F64));
        let id = self
            .module
            .declare_function("haira_sin", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("sin"), id);

        // haira_cos(x) -> f64
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(types::F64));
        sig.returns.push(AbiParam::new(types::F64));
        let id = self
            .module
            .declare_function("haira_cos", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("cos"), id);

        // haira_tan(x) -> f64
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(types::F64));
        sig.returns.push(AbiParam::new(types::F64));
        let id = self
            .module
            .declare_function("haira_tan", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("tan"), id);

        // haira_asin(x) -> f64
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(types::F64));
        sig.returns.push(AbiParam::new(types::F64));
        let id = self
            .module
            .declare_function("haira_asin", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("asin"), id);

        // haira_acos(x) -> f64
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(types::F64));
        sig.returns.push(AbiParam::new(types::F64));
        let id = self
            .module
            .declare_function("haira_acos", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("acos"), id);

        // haira_atan(x) -> f64
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(types::F64));
        sig.returns.push(AbiParam::new(types::F64));
        let id = self
            .module
            .declare_function("haira_atan", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("atan"), id);

        // haira_atan2(y, x) -> f64
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(types::F64));
        sig.params.push(AbiParam::new(types::F64));
        sig.returns.push(AbiParam::new(types::F64));
        let id = self
            .module
            .declare_function("haira_atan2", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("atan2"), id);

        // haira_random_int(max) -> i64
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(types::I64));
        sig.returns.push(AbiParam::new(types::I64));
        let id = self
            .module
            .declare_function("haira_random_int", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("random_int"), id);

        // haira_random_float() -> f64
        let mut sig = self.module.make_signature();
        sig.returns.push(AbiParam::new(types::F64));
        let id = self
            .module
            .declare_function("haira_random_float", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("random_float"), id);

        // haira_random_seed(seed)
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(types::I64));
        let id = self
            .module
            .declare_function("haira_random_seed", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("random_seed"), id);

        // ====================================================================
        // Standard Library - File I/O Functions
        // ====================================================================

        // haira_file_read(path_ptr, path_len) -> HairaString*
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(self.ptr_type));
        sig.params.push(AbiParam::new(types::I64));
        sig.returns.push(AbiParam::new(self.ptr_type));
        let id = self
            .module
            .declare_function("haira_file_read", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("file_read"), id);

        // haira_file_write(path_ptr, path_len, content_ptr, content_len) -> i64
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(self.ptr_type));
        sig.params.push(AbiParam::new(types::I64));
        sig.params.push(AbiParam::new(self.ptr_type));
        sig.params.push(AbiParam::new(types::I64));
        sig.returns.push(AbiParam::new(types::I64));
        let id = self
            .module
            .declare_function("haira_file_write", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("file_write"), id);

        // haira_file_append(path_ptr, path_len, content_ptr, content_len) -> i64
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(self.ptr_type));
        sig.params.push(AbiParam::new(types::I64));
        sig.params.push(AbiParam::new(self.ptr_type));
        sig.params.push(AbiParam::new(types::I64));
        sig.returns.push(AbiParam::new(types::I64));
        let id = self
            .module
            .declare_function("haira_file_append", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("file_append"), id);

        // haira_file_exists(path_ptr, path_len) -> i64
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(self.ptr_type));
        sig.params.push(AbiParam::new(types::I64));
        sig.returns.push(AbiParam::new(types::I64));
        let id = self
            .module
            .declare_function("haira_file_exists", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("file_exists"), id);

        // ====================================================================
        // Standard Library - Environment Functions
        // ====================================================================

        // haira_env_get(name_ptr, name_len) -> HairaString*
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(self.ptr_type));
        sig.params.push(AbiParam::new(types::I64));
        sig.returns.push(AbiParam::new(self.ptr_type));
        let id = self
            .module
            .declare_function("haira_env_get", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("env"), id);

        // haira_exit(code)
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(types::I64));
        let id = self
            .module
            .declare_function("haira_exit", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("exit"), id);

        // ====================================================================
        // Standard Library - Time Functions
        // ====================================================================

        // haira_time_now() -> i64
        let mut sig = self.module.make_signature();
        sig.returns.push(AbiParam::new(types::I64));
        let id = self
            .module
            .declare_function("haira_time_now", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("time_now"), id);

        // haira_time_monotonic() -> i64
        let mut sig = self.module.make_signature();
        sig.returns.push(AbiParam::new(types::I64));
        let id = self
            .module
            .declare_function("haira_time_monotonic", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("time_monotonic"), id);

        // ====================================================================
        // Standard Library - Testing Functions
        // ====================================================================

        // haira_test_start(name_ptr, name_len)
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(self.ptr_type));
        sig.params.push(AbiParam::new(types::I64));
        let id = self
            .module
            .declare_function("haira_test_start", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("test_start"), id);

        // haira_test_pass()
        let sig = self.module.make_signature();
        let id = self
            .module
            .declare_function("haira_test_pass", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("test_pass"), id);

        // haira_test_fail(msg_ptr, msg_len)
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(self.ptr_type));
        sig.params.push(AbiParam::new(types::I64));
        let id = self
            .module
            .declare_function("haira_test_fail", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("test_fail"), id);

        // haira_assert(condition) -> i64
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(types::I64));
        sig.returns.push(AbiParam::new(types::I64));
        let id = self
            .module
            .declare_function("haira_assert", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("assert"), id);

        // haira_assert_eq(expected, actual) -> i64
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(types::I64));
        sig.params.push(AbiParam::new(types::I64));
        sig.returns.push(AbiParam::new(types::I64));
        let id = self
            .module
            .declare_function("haira_assert_eq", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("assert_eq"), id);

        // haira_assert_ne(a, b) -> i64
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(types::I64));
        sig.params.push(AbiParam::new(types::I64));
        sig.returns.push(AbiParam::new(types::I64));
        let id = self
            .module
            .declare_function("haira_assert_ne", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("assert_ne"), id);

        // haira_assert_gt(a, b) -> i64
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(types::I64));
        sig.params.push(AbiParam::new(types::I64));
        sig.returns.push(AbiParam::new(types::I64));
        let id = self
            .module
            .declare_function("haira_assert_gt", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("assert_gt"), id);

        // haira_assert_ge(a, b) -> i64
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(types::I64));
        sig.params.push(AbiParam::new(types::I64));
        sig.returns.push(AbiParam::new(types::I64));
        let id = self
            .module
            .declare_function("haira_assert_ge", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("assert_ge"), id);

        // haira_assert_lt(a, b) -> i64
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(types::I64));
        sig.params.push(AbiParam::new(types::I64));
        sig.returns.push(AbiParam::new(types::I64));
        let id = self
            .module
            .declare_function("haira_assert_lt", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("assert_lt"), id);

        // haira_assert_le(a, b) -> i64
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(types::I64));
        sig.params.push(AbiParam::new(types::I64));
        sig.returns.push(AbiParam::new(types::I64));
        let id = self
            .module
            .declare_function("haira_assert_le", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("assert_le"), id);

        // haira_test_summary() -> i64
        let mut sig = self.module.make_signature();
        sig.returns.push(AbiParam::new(types::I64));
        let id = self
            .module
            .declare_function("haira_test_summary", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("test_summary"), id);

        // haira_test_section(name_ptr, name_len)
        let mut sig = self.module.make_signature();
        sig.params.push(AbiParam::new(self.ptr_type));
        sig.params.push(AbiParam::new(types::I64));
        let id = self
            .module
            .declare_function("haira_test_section", Linkage::Import, &sig)?;
        self.functions.insert(SmolStr::from("test_section"), id);

        Ok(())
    }

    /// Register a struct type definition.
    fn register_struct(&mut self, type_def: &TypeDef) {
        let mut fields = Vec::new();
        let mut field_offsets = Vec::new();
        let mut offset = 0;

        for field in &type_def.fields {
            fields.push(field.name.node.clone());
            field_offsets.push(offset);
            // All fields are i64 (8 bytes) for now
            offset += 8;
        }

        let info = StructInfo {
            fields,
            field_offsets,
            size: offset,
        };

        self.structs.insert(type_def.name.node.clone(), info);
    }

    /// Compile the AST.
    pub fn compile(&mut self, ast: &SourceFile) -> Result<(), CodegenError> {
        // Declare runtime functions
        self.declare_runtime_functions()?;

        // First pass: register all struct types
        for item in &ast.items {
            if let ItemKind::TypeDef(type_def) = &item.node {
                self.register_struct(type_def);
            }
        }

        // Collect all spawn blocks from the AST
        self.collect_spawn_blocks(ast);

        // Second pass: declare all user functions and methods
        for item in &ast.items {
            if let ItemKind::FunctionDef(func) = &item.node {
                let mut sig = self.module.make_signature();

                // Add parameters
                for _param in &func.params {
                    // For now, assume all params are i64 (or pointer for structs)
                    sig.params.push(AbiParam::new(types::I64));
                }

                // Return type (assume i64 for now)
                sig.returns.push(AbiParam::new(types::I64));

                let id =
                    self.module
                        .declare_function(func.name.node.as_str(), Linkage::Export, &sig)?;
                self.functions.insert(func.name.node.clone(), id);
            }

            if let ItemKind::MethodDef(method) = &item.node {
                let mut sig = self.module.make_signature();

                // First parameter is self (pointer to struct)
                sig.params.push(AbiParam::new(self.ptr_type));

                // Add other parameters
                for _param in &method.params {
                    sig.params.push(AbiParam::new(types::I64));
                }

                // Return type (assume i64 for now)
                sig.returns.push(AbiParam::new(types::I64));

                // Method name: TypeName_methodName
                let method_full_name = format!("{}_{}", method.type_name.node, method.name.node);
                let id = self
                    .module
                    .declare_function(&method_full_name, Linkage::Export, &sig)?;
                self.functions.insert(SmolStr::from(&method_full_name), id);
            }
        }

        // Declare spawn block functions (no params, returns i64)
        self.declare_spawn_functions()?;

        // Declare async block functions (no params, returns i64)
        self.declare_async_functions()?;

        // Third pass: compile function and method bodies
        for item in &ast.items {
            if let ItemKind::FunctionDef(func) = &item.node {
                self.compile_function(func)?;
            }
            if let ItemKind::MethodDef(method) = &item.node {
                self.compile_method(method)?;
            }
        }

        // Compile spawn block functions
        self.compile_spawn_functions()?;

        // Compile async block functions
        self.compile_async_functions()?;

        // Compile main function from top-level statements
        self.compile_main(ast)?;

        Ok(())
    }

    /// Collect all spawn blocks from the AST.
    fn collect_spawn_blocks(&mut self, ast: &SourceFile) {
        for item in &ast.items {
            self.collect_spawn_blocks_from_item(item);
        }
    }

    fn collect_spawn_blocks_from_item(&mut self, item: &Item) {
        match &item.node {
            ItemKind::FunctionDef(func) => {
                self.collect_spawn_blocks_from_block(&func.body);
            }
            ItemKind::MethodDef(method) => {
                self.collect_spawn_blocks_from_block(&method.body);
            }
            ItemKind::Statement(stmt) => {
                self.collect_spawn_blocks_from_stmt(stmt);
            }
            _ => {}
        }
    }

    fn collect_spawn_blocks_from_block(&mut self, block: &Block) {
        for stmt in &block.statements {
            self.collect_spawn_blocks_from_stmt(stmt);
        }
    }

    fn collect_spawn_blocks_from_stmt(&mut self, stmt: &Statement) {
        match &stmt.node {
            StatementKind::Expr(expr) => {
                self.collect_spawn_blocks_from_expr(expr);
            }
            StatementKind::Assignment(assign) => {
                self.collect_spawn_blocks_from_expr(&assign.value);
            }
            StatementKind::If(if_stmt) => {
                self.collect_spawn_blocks_from_expr(&if_stmt.condition);
                self.collect_spawn_blocks_from_block(&if_stmt.then_branch);
                if let Some(else_branch) = &if_stmt.else_branch {
                    match else_branch {
                        haira_ast::ElseBranch::Block(block) => {
                            self.collect_spawn_blocks_from_block(block);
                        }
                        haira_ast::ElseBranch::ElseIf(else_if) => {
                            let else_if_stmt = Statement {
                                node: StatementKind::If(else_if.node.clone()),
                                span: else_if.span.clone(),
                            };
                            self.collect_spawn_blocks_from_stmt(&else_if_stmt);
                        }
                    }
                }
            }
            StatementKind::While(while_stmt) => {
                self.collect_spawn_blocks_from_expr(&while_stmt.condition);
                self.collect_spawn_blocks_from_block(&while_stmt.body);
            }
            StatementKind::For(for_stmt) => {
                self.collect_spawn_blocks_from_expr(&for_stmt.iterator);
                self.collect_spawn_blocks_from_block(&for_stmt.body);
            }
            StatementKind::Return(ret) => {
                for val in &ret.values {
                    self.collect_spawn_blocks_from_expr(val);
                }
            }
            StatementKind::Try(try_stmt) => {
                self.collect_spawn_blocks_from_block(&try_stmt.body);
                self.collect_spawn_blocks_from_block(&try_stmt.catch_body);
            }
            StatementKind::Match(match_expr) => {
                self.collect_spawn_blocks_from_expr(&match_expr.subject);
                for arm in &match_expr.arms {
                    match &arm.body {
                        haira_ast::MatchArmBody::Expr(expr) => {
                            self.collect_spawn_blocks_from_expr(expr);
                        }
                        haira_ast::MatchArmBody::Block(block) => {
                            self.collect_spawn_blocks_from_block(block);
                        }
                    }
                }
            }
            _ => {}
        }
    }

    fn collect_spawn_blocks_from_expr(&mut self, expr: &Expr) {
        match &expr.node {
            ExprKind::Spawn(block) => {
                // Found a spawn block! Record it with its span start as key
                let span_start = expr.span.start;
                let func_name = SmolStr::from(format!("__spawn_block_{}", self.spawn_counter));
                self.spawn_counter += 1;
                self.spawn_functions.insert(span_start, func_name);
                self.spawn_blocks.push((span_start, block.clone()));
                // Also collect any nested spawn blocks within
                self.collect_spawn_blocks_from_block(block);
            }
            ExprKind::Binary(bin) => {
                self.collect_spawn_blocks_from_expr(&bin.left);
                self.collect_spawn_blocks_from_expr(&bin.right);
            }
            ExprKind::Unary(unary) => {
                self.collect_spawn_blocks_from_expr(&unary.operand);
            }
            ExprKind::Call(call) => {
                self.collect_spawn_blocks_from_expr(&call.callee);
                for arg in &call.args {
                    self.collect_spawn_blocks_from_expr(&arg.value);
                }
            }
            ExprKind::MethodCall(method_call) => {
                self.collect_spawn_blocks_from_expr(&method_call.receiver);
                for arg in &method_call.args {
                    self.collect_spawn_blocks_from_expr(&arg.value);
                }
            }
            ExprKind::If(if_stmt) => {
                self.collect_spawn_blocks_from_expr(&if_stmt.condition);
                self.collect_spawn_blocks_from_block(&if_stmt.then_branch);
                if let Some(else_branch) = &if_stmt.else_branch {
                    match else_branch {
                        haira_ast::ElseBranch::Block(block) => {
                            self.collect_spawn_blocks_from_block(block);
                        }
                        haira_ast::ElseBranch::ElseIf(_) => {}
                    }
                }
            }
            ExprKind::Block(block) => {
                self.collect_spawn_blocks_from_block(block);
            }
            ExprKind::Match(match_expr) => {
                self.collect_spawn_blocks_from_expr(&match_expr.subject);
                for arm in &match_expr.arms {
                    match &arm.body {
                        haira_ast::MatchArmBody::Expr(expr) => {
                            self.collect_spawn_blocks_from_expr(expr);
                        }
                        haira_ast::MatchArmBody::Block(block) => {
                            self.collect_spawn_blocks_from_block(block);
                        }
                    }
                }
            }
            ExprKind::Paren(inner) => {
                self.collect_spawn_blocks_from_expr(inner);
            }
            ExprKind::Propagate(inner) => {
                self.collect_spawn_blocks_from_expr(inner);
            }
            ExprKind::Some(inner) => {
                self.collect_spawn_blocks_from_expr(inner);
            }
            ExprKind::Lambda(lambda) => match &lambda.body {
                haira_ast::LambdaBody::Expr(expr) => {
                    self.collect_spawn_blocks_from_expr(expr);
                }
                haira_ast::LambdaBody::Block(block) => {
                    self.collect_spawn_blocks_from_block(block);
                }
            },
            ExprKind::Async(block) => {
                // Found an async block! Record it with its span start as key
                // Each statement in the block will become a separate function
                let span_start = expr.span.start;
                let mut func_names = Vec::new();
                for (i, _stmt) in block.statements.iter().enumerate() {
                    let func_name =
                        SmolStr::from(format!("__async_block_{}_{}", self.async_counter, i));
                    func_names.push(func_name);
                }
                self.async_counter += 1;
                self.async_functions.insert(span_start, func_names);
                self.async_blocks.push((span_start, block.clone()));
                // Also collect any nested spawn/async blocks within
                self.collect_spawn_blocks_from_block(block);
            }
            ExprKind::Pipe(pipe) => {
                self.collect_spawn_blocks_from_expr(&pipe.left);
                self.collect_spawn_blocks_from_expr(&pipe.right);
            }
            ExprKind::List(elements) => {
                for elem in elements {
                    self.collect_spawn_blocks_from_expr(elem);
                }
            }
            ExprKind::Index(index_expr) => {
                self.collect_spawn_blocks_from_expr(&index_expr.object);
                self.collect_spawn_blocks_from_expr(&index_expr.index);
            }
            ExprKind::Field(field_expr) => {
                self.collect_spawn_blocks_from_expr(&field_expr.object);
            }
            ExprKind::Instance(instance) => {
                for field in &instance.fields {
                    self.collect_spawn_blocks_from_expr(&field.value);
                }
            }
            ExprKind::Range(range) => {
                self.collect_spawn_blocks_from_expr(&range.start);
                self.collect_spawn_blocks_from_expr(&range.end);
            }
            ExprKind::Ai(_ai_block) => {
                // AI blocks are handled separately during pre-interpretation.
                // No nested spawn/async blocks to collect from the intent text.
            }
            _ => {}
        }
    }

    /// Declare spawn block functions.
    fn declare_spawn_functions(&mut self) -> Result<(), CodegenError> {
        for (_, func_name) in &self.spawn_functions {
            // Spawn functions take no parameters and return i64
            let mut sig = self.module.make_signature();
            sig.returns.push(AbiParam::new(types::I64));

            let id = self
                .module
                .declare_function(func_name.as_str(), Linkage::Local, &sig)?;
            self.functions.insert(func_name.clone(), id);
        }
        Ok(())
    }

    /// Compile spawn block functions.
    fn compile_spawn_functions(&mut self) -> Result<(), CodegenError> {
        // Take ownership of spawn_blocks to avoid borrow issues
        let spawn_blocks = std::mem::take(&mut self.spawn_blocks);

        for (span_start, block) in spawn_blocks {
            let func_name = self.spawn_functions.get(&span_start).unwrap().clone();
            self.compile_spawn_block_function(&func_name, &block)?;
        }

        Ok(())
    }

    /// Compile a single spawn block as a function.
    fn compile_spawn_block_function(
        &mut self,
        func_name: &SmolStr,
        block: &Block,
    ) -> Result<(), CodegenError> {
        let func_id = *self
            .functions
            .get(func_name)
            .ok_or_else(|| CodegenError::UndefinedFunction(func_name.to_string()))?;

        self.ctx.func.signature = self
            .module
            .declarations()
            .get_function_decl(func_id)
            .signature
            .clone();

        {
            let mut builder = FunctionBuilder::new(&mut self.ctx.func, &mut self.builder_ctx);

            let entry_block = builder.create_block();
            builder.switch_to_block(entry_block);
            builder.seal_block(entry_block);

            let mut scope = FunctionScope::new(self.ptr_type);

            let mut func_compiler = FunctionCompiler {
                module: &mut self.module,
                strings: &mut self.strings,
                functions: &self.functions,
                structs: &self.structs,
                ptr_type: self.ptr_type,
                spawn_functions: &self.spawn_functions,
                async_functions: &self.async_functions,
            };

            let result = func_compiler.compile_block(block, &mut scope, &mut builder)?;

            if !builder.is_unreachable() {
                let ret_val = result.unwrap_or_else(|| builder.ins().iconst(types::I64, 0));
                builder.ins().return_(&[ret_val]);
            }

            builder.finalize();
        }

        self.module
            .define_function(func_id, &mut self.ctx)
            .map_err(|e| CodegenError::ModuleError(e))?;

        self.ctx.clear();

        Ok(())
    }

    /// Declare async block functions.
    fn declare_async_functions(&mut self) -> Result<(), CodegenError> {
        for (_, func_names) in &self.async_functions {
            for func_name in func_names {
                // Async functions take no parameters and return i64
                let mut sig = self.module.make_signature();
                sig.returns.push(AbiParam::new(types::I64));

                let id = self
                    .module
                    .declare_function(func_name.as_str(), Linkage::Local, &sig)?;
                self.functions.insert(func_name.clone(), id);
            }
        }
        Ok(())
    }

    /// Compile async block functions.
    fn compile_async_functions(&mut self) -> Result<(), CodegenError> {
        // Take ownership of async_blocks to avoid borrow issues
        let async_blocks = std::mem::take(&mut self.async_blocks);

        for (span_start, block) in async_blocks {
            let func_names = self.async_functions.get(&span_start).unwrap().clone();
            // Compile each statement as a separate function
            for (i, stmt) in block.statements.iter().enumerate() {
                if i < func_names.len() {
                    self.compile_async_statement_function(&func_names[i], stmt)?;
                }
            }
        }

        Ok(())
    }

    /// Compile a single async statement as a function.
    fn compile_async_statement_function(
        &mut self,
        func_name: &SmolStr,
        stmt: &Statement,
    ) -> Result<(), CodegenError> {
        let func_id = *self
            .functions
            .get(func_name)
            .ok_or_else(|| CodegenError::UndefinedFunction(func_name.to_string()))?;

        self.ctx.func.signature = self
            .module
            .declarations()
            .get_function_decl(func_id)
            .signature
            .clone();

        {
            let mut builder = FunctionBuilder::new(&mut self.ctx.func, &mut self.builder_ctx);

            let entry_block = builder.create_block();
            builder.switch_to_block(entry_block);
            builder.seal_block(entry_block);

            let mut scope = FunctionScope::new(self.ptr_type);

            let mut func_compiler = FunctionCompiler {
                module: &mut self.module,
                strings: &mut self.strings,
                functions: &self.functions,
                structs: &self.structs,
                ptr_type: self.ptr_type,
                spawn_functions: &self.spawn_functions,
                async_functions: &self.async_functions,
            };

            let result = func_compiler.compile_statement(stmt, &mut scope, &mut builder)?;

            if !builder.is_unreachable() {
                let ret_val = result.unwrap_or_else(|| builder.ins().iconst(types::I64, 0));
                builder.ins().return_(&[ret_val]);
            }

            builder.finalize();
        }

        self.module
            .define_function(func_id, &mut self.ctx)
            .map_err(|e| CodegenError::ModuleError(e))?;

        self.ctx.clear();

        Ok(())
    }

    /// Compile a user-defined function.
    fn compile_function(&mut self, func: &haira_ast::FunctionDef) -> Result<(), CodegenError> {
        let func_id = *self
            .functions
            .get(&func.name.node)
            .ok_or_else(|| CodegenError::UndefinedFunction(func.name.node.to_string()))?;

        self.ctx.func.signature = self
            .module
            .declarations()
            .get_function_decl(func_id)
            .signature
            .clone();

        // Build function body
        {
            let mut builder = FunctionBuilder::new(&mut self.ctx.func, &mut self.builder_ctx);

            let entry_block = builder.create_block();
            builder.append_block_params_for_function_params(entry_block);
            builder.switch_to_block(entry_block);
            // Entry block has no predecessors, so seal immediately
            builder.seal_block(entry_block);

            // Create scope for variables
            let mut scope = FunctionScope::new(self.ptr_type);

            // Bind parameters to variables
            let params = builder.block_params(entry_block).to_vec();
            for (i, param) in func.params.iter().enumerate() {
                if i < params.len() {
                    // Create a Cranelift variable for each parameter
                    let var = scope.declare_var(&param.name.node, &mut builder);
                    builder.def_var(var, params[i]);
                }
            }

            // Create a function compiler that doesn't hold references to self
            let mut func_compiler = FunctionCompiler {
                module: &mut self.module,
                strings: &mut self.strings,
                functions: &self.functions,
                structs: &self.structs,
                ptr_type: self.ptr_type,
                spawn_functions: &self.spawn_functions,
                async_functions: &self.async_functions,
            };

            // Compile function body
            let result = func_compiler.compile_block(&func.body, &mut scope, &mut builder)?;

            // Only add a return if the current block is not already terminated
            // is_unreachable() returns true if we're after a terminator instruction
            if !builder.is_unreachable() {
                // Return the result or 0
                let ret_val = result.unwrap_or_else(|| builder.ins().iconst(types::I64, 0));
                builder.ins().return_(&[ret_val]);
            }

            builder.finalize();
        }

        self.module
            .define_function(func_id, &mut self.ctx)
            .map_err(|e| CodegenError::ModuleError(e))?;

        self.ctx.clear();

        Ok(())
    }

    /// Compile a method definition.
    fn compile_method(&mut self, method: &MethodDef) -> Result<(), CodegenError> {
        let method_full_name = format!("{}_{}", method.type_name.node, method.name.node);
        let func_id = *self
            .functions
            .get(&SmolStr::from(&method_full_name))
            .ok_or_else(|| CodegenError::UndefinedFunction(method_full_name.clone()))?;

        self.ctx.func.signature = self
            .module
            .declarations()
            .get_function_decl(func_id)
            .signature
            .clone();

        // Build method body
        {
            let mut builder = FunctionBuilder::new(&mut self.ctx.func, &mut self.builder_ctx);

            let entry_block = builder.create_block();
            builder.append_block_params_for_function_params(entry_block);
            builder.switch_to_block(entry_block);
            builder.seal_block(entry_block);

            let mut scope = FunctionScope::new(self.ptr_type);

            // Bind parameters to variables
            let params = builder.block_params(entry_block).to_vec();

            // First parameter is 'self' - the struct pointer
            if !params.is_empty() {
                let self_var = scope.declare_var(&SmolStr::from("self"), &mut builder);
                builder.def_var(self_var, params[0]);
            }

            // Bind other parameters
            for (i, param) in method.params.iter().enumerate() {
                if i + 1 < params.len() {
                    let var = scope.declare_var(&param.name.node, &mut builder);
                    builder.def_var(var, params[i + 1]);
                }
            }

            let mut func_compiler = FunctionCompiler {
                module: &mut self.module,
                strings: &mut self.strings,
                functions: &self.functions,
                structs: &self.structs,
                ptr_type: self.ptr_type,
                spawn_functions: &self.spawn_functions,
                async_functions: &self.async_functions,
            };

            let result = func_compiler.compile_block(&method.body, &mut scope, &mut builder)?;

            if !builder.is_unreachable() {
                let ret_val = result.unwrap_or_else(|| builder.ins().iconst(types::I64, 0));
                builder.ins().return_(&[ret_val]);
            }

            builder.finalize();
        }

        self.module
            .define_function(func_id, &mut self.ctx)
            .map_err(|e| CodegenError::ModuleError(e))?;

        self.ctx.clear();

        Ok(())
    }

    /// Compile the main function from top-level statements.
    fn compile_main(&mut self, ast: &SourceFile) -> Result<(), CodegenError> {
        // Create main function signature
        let mut sig = self.module.make_signature();
        sig.returns.push(AbiParam::new(types::I32)); // main returns i32

        let main_id = self
            .module
            .declare_function("main", Linkage::Export, &sig)?;

        self.ctx.func.signature = sig;

        {
            let mut builder = FunctionBuilder::new(&mut self.ctx.func, &mut self.builder_ctx);

            let entry_block = builder.create_block();
            builder.switch_to_block(entry_block);
            // Entry block has no predecessors, seal immediately
            builder.seal_block(entry_block);

            let mut scope = FunctionScope::new(self.ptr_type);

            // Create a function compiler
            let mut func_compiler = FunctionCompiler {
                module: &mut self.module,
                strings: &mut self.strings,
                functions: &self.functions,
                structs: &self.structs,
                ptr_type: self.ptr_type,
                spawn_functions: &self.spawn_functions,
                async_functions: &self.async_functions,
            };

            // Compile all top-level statements (not function defs)
            for item in &ast.items {
                if let ItemKind::Statement(stmt) = &item.node {
                    func_compiler.compile_statement(stmt, &mut scope, &mut builder)?;
                }
            }

            // Return 0
            let zero = builder.ins().iconst(types::I32, 0);
            builder.ins().return_(&[zero]);

            builder.finalize();
        }

        self.module
            .define_function(main_id, &mut self.ctx)
            .map_err(|e| CodegenError::ModuleError(e))?;

        self.ctx.clear();

        Ok(())
    }

    /// Finish compilation and return object bytes.
    pub fn finish(self) -> Vec<u8> {
        let product = self.module.finish();
        product.emit().unwrap()
    }
}

/// Helper struct for compiling function bodies.
/// This is separate from Compiler to avoid borrow issues.
struct FunctionCompiler<'a> {
    module: &'a mut ObjectModule,
    strings: &'a mut HashMap<SmolStr, cranelift_module::DataId>,
    functions: &'a HashMap<SmolStr, FuncId>,
    structs: &'a HashMap<SmolStr, StructInfo>,
    ptr_type: Type,
    /// Map of spawn block span start to their function names.
    spawn_functions: &'a HashMap<u32, SmolStr>,
    /// Map of async block span start to their function names.
    async_functions: &'a HashMap<u32, Vec<SmolStr>>,
}

impl<'a> FunctionCompiler<'a> {
    /// Define a string constant and return its data ID.
    fn define_string(&mut self, s: &str) -> Result<cranelift_module::DataId, CodegenError> {
        let key = SmolStr::from(s);
        if let Some(&id) = self.strings.get(&key) {
            return Ok(id);
        }

        let name = format!(".str.{}", self.strings.len());
        let id = self
            .module
            .declare_data(&name, Linkage::Local, false, false)?;

        let mut desc = DataDescription::new();
        desc.define(s.as_bytes().to_vec().into_boxed_slice());

        self.module.define_data(id, &desc)?;
        self.strings.insert(key, id);

        Ok(id)
    }

    /// Compile a block of statements.
    fn compile_block(
        &mut self,
        block: &Block,
        scope: &mut FunctionScope,
        builder: &mut FunctionBuilder,
    ) -> Result<Option<Value>, CodegenError> {
        let mut last_value = None;

        for stmt in &block.statements {
            last_value = self.compile_statement(stmt, scope, builder)?;
        }

        Ok(last_value)
    }

    /// Compile a statement.
    fn compile_statement(
        &mut self,
        stmt: &Statement,
        scope: &mut FunctionScope,
        builder: &mut FunctionBuilder,
    ) -> Result<Option<Value>, CodegenError> {
        match &stmt.node {
            StatementKind::Expr(expr) => {
                let val = self.compile_expr(expr, scope, builder)?;
                Ok(Some(val))
            }
            StatementKind::Assignment(assign) => {
                let value = self.compile_expr(&assign.value, scope, builder)?;
                for target in &assign.targets {
                    // Use Cranelift variables for proper SSA handling
                    let var = scope.get_or_declare_var(&target.name.node, builder);
                    builder.def_var(var, value);
                }
                Ok(Some(value))
            }
            StatementKind::Return(ret) => {
                if ret.values.is_empty() {
                    let zero = builder.ins().iconst(types::I64, 0);
                    builder.ins().return_(&[zero]);
                } else {
                    let val = self.compile_expr(&ret.values[0], scope, builder)?;
                    builder.ins().return_(&[val]);
                }
                // Create an unreachable block to switch to after return
                // This prevents adding more instructions to the terminated block
                let unreachable_block = builder.create_block();
                builder.switch_to_block(unreachable_block);
                builder.seal_block(unreachable_block);
                Ok(None)
            }
            StatementKind::If(if_stmt) => {
                let cond = self.compile_expr(&if_stmt.condition, scope, builder)?;

                let then_block = builder.create_block();
                let else_block = builder.create_block();
                let merge_block = builder.create_block();

                builder.ins().brif(cond, then_block, &[], else_block, &[]);

                // Then branch - seal since only predecessor is the branch source
                builder.switch_to_block(then_block);
                builder.seal_block(then_block);
                self.compile_block(&if_stmt.then_branch, scope, builder)?;
                builder.ins().jump(merge_block, &[]);

                // Else branch - seal since only predecessor is the branch source
                builder.switch_to_block(else_block);
                builder.seal_block(else_block);
                if let Some(else_branch) = &if_stmt.else_branch {
                    match else_branch {
                        haira_ast::ElseBranch::Block(block) => {
                            self.compile_block(block, scope, builder)?;
                        }
                        haira_ast::ElseBranch::ElseIf(else_if) => {
                            let else_if_stmt = Statement {
                                node: StatementKind::If(else_if.node.clone()),
                                span: else_if.span.clone(),
                            };
                            self.compile_statement(&else_if_stmt, scope, builder)?;
                        }
                    }
                }
                builder.ins().jump(merge_block, &[]);

                // Merge block - seal since both predecessors (then and else) have jumped
                builder.switch_to_block(merge_block);
                builder.seal_block(merge_block);

                Ok(None)
            }
            StatementKind::While(while_stmt) => {
                // For while loops, we need to collect all variables that might be modified
                // in the loop and pass them as block parameters to handle SSA properly.
                //
                // Strategy: Use a pre-header block to get initial values, then use block
                // parameters for the loop header to handle the phi nodes.

                // Collect all variables currently in scope - they may be modified in the loop
                let loop_vars: Vec<(SmolStr, Variable)> = scope
                    .variables
                    .iter()
                    .map(|(name, &var)| (name.clone(), var))
                    .collect();

                let header_block = builder.create_block();
                let body_block = builder.create_block();
                let exit_block = builder.create_block();

                // Add block parameters for all variables that might be used in loop
                for _ in &loop_vars {
                    builder.append_block_param(header_block, types::I64);
                }

                // Get current values of all loop variables and jump to header
                let initial_values: Vec<Value> = loop_vars
                    .iter()
                    .map(|(_, var)| builder.use_var(*var))
                    .collect();
                builder.ins().jump(header_block, &initial_values);

                // Header block
                builder.switch_to_block(header_block);

                // Update variables with block parameters (phi values)
                let header_params = builder.block_params(header_block).to_vec();
                for (i, (_, var)) in loop_vars.iter().enumerate() {
                    builder.def_var(*var, header_params[i]);
                }

                // Seal header after setting up block params
                builder.seal_block(header_block);

                // Compile condition
                let cond = self.compile_expr(&while_stmt.condition, scope, builder)?;
                builder.ins().brif(cond, body_block, &[], exit_block, &[]);

                // Body
                builder.switch_to_block(body_block);
                builder.seal_block(body_block);
                self.compile_block(&while_stmt.body, scope, builder)?;

                // Get current values after body and jump back to header
                let loop_values: Vec<Value> = loop_vars
                    .iter()
                    .map(|(_, var)| builder.use_var(*var))
                    .collect();
                builder.ins().jump(header_block, &loop_values);

                // Exit block
                builder.switch_to_block(exit_block);
                builder.seal_block(exit_block);

                Ok(None)
            }
            StatementKind::For(for_stmt) => {
                // For now, only support range iteration: for i in 0..10
                if let ExprKind::Range(range) = &for_stmt.iterator.node {
                    let start = self.compile_expr(&range.start, scope, builder)?;
                    let end = self.compile_expr(&range.end, scope, builder)?;

                    // Declare loop variable
                    let loop_var_name =
                        if let haira_ast::ForPattern::Single(name) = &for_stmt.pattern {
                            name.node.clone()
                        } else {
                            return Err(CodegenError::Unsupported(
                                "Only single variable for loops supported".to_string(),
                            ));
                        };
                    let loop_var = scope.declare_var(&loop_var_name, builder);
                    builder.def_var(loop_var, start);

                    let header_block = builder.create_block();
                    let body_block = builder.create_block();
                    let exit_block = builder.create_block();

                    // Jump to header (first predecessor)
                    builder.ins().jump(header_block, &[]);

                    // Header - DON'T seal yet, need back-edge from body
                    builder.switch_to_block(header_block);
                    let current = builder.use_var(loop_var);

                    let cmp = if range.inclusive {
                        // For inclusive range, use <= (less than OR equal)
                        let lt = builder.ins().icmp(IntCC::SignedLessThan, current, end);
                        let eq = builder.ins().icmp(IntCC::Equal, current, end);
                        builder.ins().bor(lt, eq)
                    } else {
                        builder.ins().icmp(IntCC::SignedLessThan, current, end)
                    };
                    builder.ins().brif(cmp, body_block, &[], exit_block, &[]);

                    // Body - seal since only predecessor is header
                    builder.switch_to_block(body_block);
                    builder.seal_block(body_block);
                    self.compile_block(&for_stmt.body, scope, builder)?;

                    // Increment
                    let current = builder.use_var(loop_var);
                    let one = builder.ins().iconst(types::I64, 1);
                    let next = builder.ins().iadd(current, one);
                    builder.def_var(loop_var, next);
                    // Back-edge to header (second predecessor)
                    builder.ins().jump(header_block, &[]);

                    // NOW seal header - both predecessors added
                    builder.seal_block(header_block);

                    // Exit block - seal since only predecessor is header
                    builder.switch_to_block(exit_block);
                    builder.seal_block(exit_block);
                } else {
                    return Err(CodegenError::Unsupported(
                        "Only range-based for loops are currently supported".to_string(),
                    ));
                }

                Ok(None)
            }
            StatementKind::Break => Ok(None),
            StatementKind::Continue => Ok(None),
            StatementKind::Match(match_expr) => {
                // Match as statement - compile as expression and discard result
                let _val = self.compile_match_expr(match_expr, scope, builder)?;
                Ok(None)
            }
            StatementKind::Try(try_stmt) => {
                // try { body } catch e { catch_body }
                // 1. Clear any existing error
                // 2. Execute try body
                // 3. Check for error
                // 4. If error, bind error to variable and execute catch body

                let clear_error_id = *self.functions.get(&SmolStr::from("clear_error")).unwrap();
                let clear_error_func = self
                    .module
                    .declare_func_in_func(clear_error_id, builder.func);

                let has_error_id = *self.functions.get(&SmolStr::from("has_error")).unwrap();
                let has_error_func = self.module.declare_func_in_func(has_error_id, builder.func);

                let get_error_id = *self.functions.get(&SmolStr::from("get_error")).unwrap();
                let get_error_func = self.module.declare_func_in_func(get_error_id, builder.func);

                // Clear error before try
                builder.ins().call(clear_error_func, &[]);

                // Compile try body
                self.compile_block(&try_stmt.body, scope, builder)?;

                // Check for error
                let call = builder.ins().call(has_error_func, &[]);
                let has_err = builder.inst_results(call)[0];

                let catch_block = builder.create_block();
                let continue_block = builder.create_block();

                builder
                    .ins()
                    .brif(has_err, catch_block, &[], continue_block, &[]);

                // Catch block
                builder.switch_to_block(catch_block);
                builder.seal_block(catch_block);

                // Get error value and bind to variable
                let call = builder.ins().call(get_error_func, &[]);
                let err_val = builder.inst_results(call)[0];

                let err_var = scope.get_or_declare_var(&try_stmt.error_name.node, builder);
                builder.def_var(err_var, err_val);

                // Compile catch body
                self.compile_block(&try_stmt.catch_body, scope, builder)?;
                builder.ins().jump(continue_block, &[]);

                // Continue block
                builder.switch_to_block(continue_block);
                builder.seal_block(continue_block);

                Ok(None)
            }
            _ => Err(CodegenError::Unsupported(format!(
                "Statement type not yet supported: {:?}",
                std::mem::discriminant(&stmt.node)
            ))),
        }
    }

    /// Compile an expression.
    fn compile_expr(
        &mut self,
        expr: &Expr,
        scope: &mut FunctionScope,
        builder: &mut FunctionBuilder,
    ) -> Result<Value, CodegenError> {
        match &expr.node {
            ExprKind::Literal(lit) => self.compile_literal(lit, scope, builder),
            ExprKind::Identifier(name) => {
                // Use Cranelift variable
                if let Some(var) = scope.get_var(name) {
                    Ok(builder.use_var(var))
                } else {
                    Err(CodegenError::UndefinedVariable(name.to_string()))
                }
            }
            ExprKind::Binary(bin) => {
                let left = self.compile_expr(&bin.left, scope, builder)?;
                let right = self.compile_expr(&bin.right, scope, builder)?;
                self.compile_binary_op(&bin.op.node, left, right, builder)
            }
            ExprKind::Unary(unary) => {
                let operand = self.compile_expr(&unary.operand, scope, builder)?;
                self.compile_unary_op(&unary.op.node, operand, builder)
            }
            ExprKind::Call(call) => self.compile_call(call, scope, builder),
            ExprKind::MethodCall(method_call) => {
                // Method call: obj.method(args)
                // Compile receiver (the object)
                let receiver = self.compile_expr(&method_call.receiver, scope, builder)?;

                // Try to find the method - we need to search through all types
                // For now, we'll try each struct type to find a matching method
                let method_name = &method_call.method.node;

                for (type_name, _) in self.structs.iter() {
                    let full_method_name = format!("{}_{}", type_name, method_name);
                    if let Some(&func_id) = self.functions.get(&SmolStr::from(&full_method_name)) {
                        let local_callee = self.module.declare_func_in_func(func_id, builder.func);

                        // First argument is self (the receiver), then other args
                        let mut args = vec![receiver];
                        for arg in &method_call.args {
                            args.push(self.compile_expr(&arg.value, scope, builder)?);
                        }

                        let call_inst = builder.ins().call(local_callee, &args);
                        let results = builder.inst_results(call_inst);

                        return if results.is_empty() {
                            Ok(builder.ins().iconst(types::I64, 0))
                        } else {
                            Ok(results[0])
                        };
                    }
                }

                Err(CodegenError::UndefinedFunction(format!(
                    "Method {} not found",
                    method_name
                )))
            }
            ExprKind::Paren(inner) => self.compile_expr(inner, scope, builder),
            ExprKind::If(if_stmt) => {
                // If as expression
                let cond = self.compile_expr(&if_stmt.condition, scope, builder)?;

                let then_block = builder.create_block();
                let else_block = builder.create_block();
                let merge_block = builder.create_block();
                builder.append_block_param(merge_block, types::I64);

                builder.ins().brif(cond, then_block, &[], else_block, &[]);

                // Then - seal since only predecessor is branch source
                builder.switch_to_block(then_block);
                builder.seal_block(then_block);
                let then_val = self
                    .compile_block(&if_stmt.then_branch, scope, builder)?
                    .unwrap_or_else(|| builder.ins().iconst(types::I64, 0));
                builder.ins().jump(merge_block, &[then_val]);

                // Else - seal since only predecessor is branch source
                builder.switch_to_block(else_block);
                builder.seal_block(else_block);
                let else_val = if let Some(else_branch) = &if_stmt.else_branch {
                    match else_branch {
                        haira_ast::ElseBranch::Block(block) => self
                            .compile_block(block, scope, builder)?
                            .unwrap_or_else(|| builder.ins().iconst(types::I64, 0)),
                        haira_ast::ElseBranch::ElseIf(_) => builder.ins().iconst(types::I64, 0),
                    }
                } else {
                    builder.ins().iconst(types::I64, 0)
                };
                builder.ins().jump(merge_block, &[else_val]);

                // Merge - seal since both predecessors have jumped
                builder.switch_to_block(merge_block);
                builder.seal_block(merge_block);

                Ok(builder.block_params(merge_block)[0])
            }
            ExprKind::Block(block) => {
                let val = self.compile_block(block, scope, builder)?;
                Ok(val.unwrap_or_else(|| builder.ins().iconst(types::I64, 0)))
            }
            ExprKind::Match(match_expr) => self.compile_match_expr(match_expr, scope, builder),
            ExprKind::Propagate(inner) => {
                // Error propagation: expr?
                // 1. Evaluate the expression
                // 2. Check if there's an error
                // 3. If error, return early from function
                // 4. Otherwise, return the value

                let val = self.compile_expr(inner, scope, builder)?;

                let has_error_id = *self.functions.get(&SmolStr::from("has_error")).unwrap();
                let has_error_func = self.module.declare_func_in_func(has_error_id, builder.func);

                let call = builder.ins().call(has_error_func, &[]);
                let has_err = builder.inst_results(call)[0];

                let error_block = builder.create_block();
                let continue_block = builder.create_block();

                builder
                    .ins()
                    .brif(has_err, error_block, &[], continue_block, &[]);

                // Error block - return early with error value (0 for now)
                builder.switch_to_block(error_block);
                builder.seal_block(error_block);
                let zero = builder.ins().iconst(types::I64, 0);
                builder.ins().return_(&[zero]);

                // Continue block - return the value
                builder.switch_to_block(continue_block);
                builder.seal_block(continue_block);

                Ok(val)
            }
            ExprKind::Instance(instance) => {
                // Struct instantiation: User { name: "Alice", age: 30 }
                let type_name = &instance.type_name.node;
                let struct_info = self
                    .structs
                    .get(type_name)
                    .ok_or_else(|| {
                        CodegenError::Unsupported(format!("Unknown type: {}", type_name))
                    })?
                    .clone();

                // Allocate memory for the struct
                let size = builder.ins().iconst(types::I64, struct_info.size as i64);
                let alloc_id = *self.functions.get(&SmolStr::from("alloc")).unwrap();
                let alloc_func = self.module.declare_func_in_func(alloc_id, builder.func);
                let call = builder.ins().call(alloc_func, &[size]);
                let ptr = builder.inst_results(call)[0];

                // Store each field value
                for inst_field in &instance.fields {
                    let field_name = inst_field
                        .name
                        .as_ref()
                        .map(|n| n.node.clone())
                        .unwrap_or_else(|| SmolStr::from(""));

                    // Find field offset
                    let field_idx = struct_info
                        .fields
                        .iter()
                        .position(|f| f == &field_name)
                        .ok_or_else(|| {
                            CodegenError::Unsupported(format!(
                                "Unknown field: {} in type {}",
                                field_name, type_name
                            ))
                        })?;

                    let offset = struct_info.field_offsets[field_idx];
                    let value = self.compile_expr(&inst_field.value, scope, builder)?;

                    // Store value at ptr + offset
                    let offset_val = builder.ins().iconst(types::I64, offset as i64);
                    let field_ptr = builder.ins().iadd(ptr, offset_val);
                    builder.ins().store(MemFlags::new(), value, field_ptr, 0);
                }

                Ok(ptr)
            }
            ExprKind::Field(field_expr) => {
                // Field access: obj.field
                let obj_ptr = self.compile_expr(&field_expr.object, scope, builder)?;
                let field_name = &field_expr.field.node;

                // We need to determine the type of the object to find the field offset
                // For now, we'll try to infer it from the scope or use a simple approach
                // This is a simplified version - a full implementation would need type inference

                // Try to find the struct type by checking all known structs
                for (_, struct_info) in self.structs.iter() {
                    if let Some(field_idx) = struct_info.fields.iter().position(|f| f == field_name)
                    {
                        let offset = struct_info.field_offsets[field_idx];
                        let offset_val = builder.ins().iconst(types::I64, offset as i64);
                        let field_ptr = builder.ins().iadd(obj_ptr, offset_val);
                        let value = builder
                            .ins()
                            .load(types::I64, MemFlags::new(), field_ptr, 0);
                        return Ok(value);
                    }
                }

                Err(CodegenError::Unsupported(format!(
                    "Unknown field: {}",
                    field_name
                )))
            }
            ExprKind::List(elements) => {
                // List literal: [1, 2, 3]
                // Allocate memory for the list: 8 bytes for length + 8 bytes per element
                let num_elements = elements.len();
                let total_size = 8 + (num_elements * 8); // length + elements
                let size_val = builder.ins().iconst(types::I64, total_size as i64);

                let alloc_id = *self.functions.get(&SmolStr::from("alloc")).unwrap();
                let alloc_func = self.module.declare_func_in_func(alloc_id, builder.func);
                let call = builder.ins().call(alloc_func, &[size_val]);
                let ptr = builder.inst_results(call)[0];

                // Store length at offset 0
                let len_val = builder.ins().iconst(types::I64, num_elements as i64);
                builder.ins().store(MemFlags::new(), len_val, ptr, 0);

                // Store each element at offset 8 + (index * 8)
                for (i, elem) in elements.iter().enumerate() {
                    let value = self.compile_expr(elem, scope, builder)?;
                    let offset = 8 + (i * 8);
                    let offset_val = builder.ins().iconst(types::I64, offset as i64);
                    let elem_ptr = builder.ins().iadd(ptr, offset_val);
                    builder.ins().store(MemFlags::new(), value, elem_ptr, 0);
                }

                Ok(ptr)
            }
            ExprKind::Index(index_expr) => {
                // Index access: arr[i]
                let arr_ptr = self.compile_expr(&index_expr.object, scope, builder)?;
                let index = self.compile_expr(&index_expr.index, scope, builder)?;

                // Element is at offset 8 + (index * 8)
                let eight = builder.ins().iconst(types::I64, 8);
                let offset = builder.ins().imul(index, eight);
                let base_offset = builder.ins().iadd(offset, eight);
                let elem_ptr = builder.ins().iadd(arr_ptr, base_offset);

                let value = builder.ins().load(types::I64, MemFlags::new(), elem_ptr, 0);
                Ok(value)
            }
            ExprKind::Lambda(_lambda) => {
                // Lambda expression: (x) { x * 2 } or x => x * 2
                // Full lambda/closure support requires more complex compilation
                // (creating functions during expression evaluation, handling closures)
                // For now, return error - use regular functions instead
                Err(CodegenError::Unsupported(
                    "Standalone lambdas not yet supported. Use regular functions instead."
                        .to_string(),
                ))
            }
            ExprKind::Async(_block) => {
                // Async blocks run operations concurrently and wait for all to complete
                // Look up the pre-compiled functions for each statement
                let span_start = expr.span.start;
                let func_names = self.async_functions.get(&span_start).ok_or_else(|| {
                    CodegenError::Unsupported(format!(
                        "Async block not found (span {}). This is a compiler bug.",
                        span_start
                    ))
                })?;

                // Get runtime functions
                let spawn_joinable_id = *self
                    .functions
                    .get(&SmolStr::from("spawn_joinable"))
                    .unwrap();
                let spawn_joinable_func = self
                    .module
                    .declare_func_in_func(spawn_joinable_id, builder.func);
                let thread_join_id = *self.functions.get(&SmolStr::from("thread_join")).unwrap();
                let thread_join_func = self
                    .module
                    .declare_func_in_func(thread_join_id, builder.func);

                // Spawn all statements as joinable threads
                let mut thread_handles = Vec::new();
                for func_name in func_names {
                    let func_id = *self
                        .functions
                        .get(func_name)
                        .ok_or_else(|| CodegenError::UndefinedFunction(func_name.to_string()))?;

                    // Get function address
                    let local_target = self.module.declare_func_in_func(func_id, builder.func);
                    let func_ptr = builder.ins().func_addr(self.ptr_type, local_target);

                    // Call haira_spawn_joinable with function pointer
                    let call_inst = builder.ins().call(spawn_joinable_func, &[func_ptr]);
                    let thread_handle = builder.inst_results(call_inst)[0];
                    thread_handles.push(thread_handle);
                }

                // Join all threads (wait for completion)
                for thread_handle in thread_handles {
                    builder.ins().call(thread_join_func, &[thread_handle]);
                }

                // Return 0 (async blocks don't produce a value currently)
                Ok(builder.ins().iconst(types::I64, 0))
            }
            ExprKind::Spawn(_block) => {
                // Spawn blocks create a new thread to run the block
                // Look up the pre-compiled function for this spawn block using its span
                let span_start = expr.span.start;
                let func_name = self.spawn_functions.get(&span_start).ok_or_else(|| {
                    CodegenError::Unsupported(format!(
                        "Spawn block not found (span {}). This is a compiler bug.",
                        span_start
                    ))
                })?;

                // Get the function ID
                let func_id = *self
                    .functions
                    .get(func_name)
                    .ok_or_else(|| CodegenError::UndefinedFunction(func_name.to_string()))?;

                // Get function address
                let local_target = self.module.declare_func_in_func(func_id, builder.func);
                let func_ptr = builder.ins().func_addr(self.ptr_type, local_target);

                // Call haira_spawn with function pointer
                let spawn_id = *self.functions.get(&SmolStr::from("spawn_thread")).unwrap();
                let spawn_func = self.module.declare_func_in_func(spawn_id, builder.func);
                let call_inst = builder.ins().call(spawn_func, &[func_ptr]);
                Ok(builder.inst_results(call_inst)[0])
            }
            ExprKind::Select(_select) => {
                // Select waits on multiple channels
                // Requires channel implementation and runtime support
                Err(CodegenError::Unsupported(
                    "Select expressions not yet supported.".to_string(),
                ))
            }
            ExprKind::Pipe(pipe) => {
                // Pipe expression: x | f or x | f(y, z)
                // Transform to: f(x) or f(x, y, z)
                // The left side becomes the first argument of the right side
                let left_val = self.compile_expr(&pipe.left, scope, builder)?;

                // The right side should be a call expression
                match &pipe.right.node {
                    ExprKind::Call(call) => {
                        // Get the function name
                        let func_name = match &call.callee.node {
                            ExprKind::Identifier(name) => name.clone(),
                            _ => {
                                return Err(CodegenError::Unsupported(
                                    "Pipe right side must be a simple function call".to_string(),
                                ))
                            }
                        };

                        // Look up function
                        let func_id = *self.functions.get(&func_name).ok_or_else(|| {
                            CodegenError::UndefinedFunction(func_name.to_string())
                        })?;

                        let local_callee = self.module.declare_func_in_func(func_id, builder.func);

                        // First argument is the piped value, then the rest
                        let mut args = vec![left_val];
                        for arg in &call.args {
                            args.push(self.compile_expr(&arg.value, scope, builder)?);
                        }

                        let call_inst = builder.ins().call(local_callee, &args);
                        let results = builder.inst_results(call_inst);

                        if results.is_empty() {
                            Ok(builder.ins().iconst(types::I64, 0))
                        } else {
                            Ok(results[0])
                        }
                    }
                    ExprKind::Identifier(func_name) => {
                        // Just a function name without parens: x | f means f(x)
                        let func_id = *self.functions.get(func_name).ok_or_else(|| {
                            CodegenError::UndefinedFunction(func_name.to_string())
                        })?;

                        let local_callee = self.module.declare_func_in_func(func_id, builder.func);
                        let call_inst = builder.ins().call(local_callee, &[left_val]);
                        let results = builder.inst_results(call_inst);

                        if results.is_empty() {
                            Ok(builder.ins().iconst(types::I64, 0))
                        } else {
                            Ok(results[0])
                        }
                    }
                    _ => Err(CodegenError::Unsupported(
                        "Pipe right side must be a function call or identifier".to_string(),
                    )),
                }
            }
            ExprKind::None => {
                // None is represented as 0 (null pointer / sentinel value)
                // We use a tagged representation: high bit = 0 means None, = 1 means Some
                // For simplicity, just use 0 for None
                Ok(builder.ins().iconst(types::I64, 0))
            }
            ExprKind::Some(inner) => {
                // Some(value) - compile the inner value
                // For proper tagged union we'd allocate and tag, but for simplicity
                // we encode Some as: (value << 1) | 1 to distinguish from None(0)
                // This works for small integers; for pointers we'd need heap allocation
                let val = self.compile_expr(inner, scope, builder)?;
                // Tag the value: shift left by 1 and set low bit to 1
                let one = builder.ins().iconst(types::I64, 1);
                let shifted = builder.ins().ishl(val, one);
                let tagged = builder.ins().bor(shifted, one);
                Ok(tagged)
            }
            ExprKind::Ai(ai_block) => {
                // AI blocks require pre-interpretation before compilation.
                // The AI engine must interpret the intent and generate CIR,
                // which is then compiled to native code.
                //
                // For now, we return an error indicating that AI blocks need
                // to be pre-processed. In a full implementation:
                // 1. A pre-compilation pass would interpret all AI blocks
                // 2. The generated CIR would be stored alongside the AST
                // 3. This code would compile the pre-generated CIR
                //
                // See `haira-ai` crate's `AIEngine::interpret_intent()` for
                // the AI interpretation logic.
                let name = ai_block
                    .name
                    .as_ref()
                    .map(|n| n.node.as_str())
                    .unwrap_or("<anonymous>");
                Err(CodegenError::Unsupported(format!(
                    "AI block '{}' requires pre-interpretation. \
                     Run `haira build --interpret-ai` to generate code from AI intents.",
                    name
                )))
            }
            _ => Err(CodegenError::Unsupported(format!(
                "Expression type not yet supported: {:?}",
                std::mem::discriminant(&expr.node)
            ))),
        }
    }

    /// Compile a literal.
    fn compile_literal(
        &mut self,
        lit: &Literal,
        scope: &mut FunctionScope,
        builder: &mut FunctionBuilder,
    ) -> Result<Value, CodegenError> {
        match lit {
            Literal::Int(n) => Ok(builder.ins().iconst(types::I64, *n)),
            Literal::Float(n) => Ok(builder.ins().f64const(*n)),
            Literal::Bool(b) => Ok(builder.ins().iconst(types::I8, if *b { 1 } else { 0 })),
            Literal::String(s) => {
                // Store string data and return pointer
                let data_id = self.define_string(s)?;
                let local_id = self.module.declare_data_in_func(data_id, builder.func);
                let ptr = builder.ins().symbol_value(self.ptr_type, local_id);
                Ok(ptr)
            }
            Literal::InterpolatedString(parts) => {
                self.compile_interpolated_string(parts, scope, builder)
            }
        }
    }

    /// Compile an interpolated string by concatenating all parts.
    /// Returns a pointer to a HairaString struct (data, len, cap).
    fn compile_interpolated_string(
        &mut self,
        parts: &[haira_ast::StringPart],
        scope: &mut FunctionScope,
        builder: &mut FunctionBuilder,
    ) -> Result<Value, CodegenError> {
        if parts.is_empty() {
            // Empty string - return empty HairaString
            let data_id = self.define_string("")?;
            let local_id = self.module.declare_data_in_func(data_id, builder.func);
            let ptr = builder.ins().symbol_value(self.ptr_type, local_id);
            return Ok(ptr);
        }

        // Convert each part to a (ptr, len) pair
        let mut string_parts: Vec<(Value, Value)> = Vec::new();

        for part in parts {
            match part {
                haira_ast::StringPart::Literal(s) => {
                    let data_id = self.define_string(s)?;
                    let local_id = self.module.declare_data_in_func(data_id, builder.func);
                    let ptr = builder.ins().symbol_value(self.ptr_type, local_id);
                    let len = builder.ins().iconst(types::I64, s.len() as i64);
                    string_parts.push((ptr, len));
                }
                haira_ast::StringPart::Expr(expr) => {
                    // Compile the expression and convert to string
                    let value = self.compile_expr(expr, scope, builder)?;

                    // Detect expression type and convert to string
                    // For now, assume integers (most common case)
                    // TODO: Add type inference for proper handling
                    let int_to_string_id =
                        *self.functions.get(&SmolStr::from("int_to_string")).unwrap();
                    let int_to_string_func = self
                        .module
                        .declare_func_in_func(int_to_string_id, builder.func);
                    let call = builder.ins().call(int_to_string_func, &[value]);
                    let haira_string_ptr = builder.inst_results(call)[0];

                    // HairaString struct: { data: *char, len: i64, cap: i64 }
                    // Load data pointer (offset 0) and len (offset 8)
                    let data_ptr =
                        builder
                            .ins()
                            .load(self.ptr_type, MemFlags::new(), haira_string_ptr, 0);
                    let len = builder
                        .ins()
                        .load(types::I64, MemFlags::new(), haira_string_ptr, 8);

                    string_parts.push((data_ptr, len));
                }
            }
        }

        // Now concatenate all parts
        if string_parts.len() == 1 {
            // Single part - just return it as-is
            // But we need to wrap it in a HairaString for consistency
            let (ptr, len) = string_parts[0];

            // Allocate a HairaString struct (24 bytes: ptr, len, cap)
            let alloc_id = *self.functions.get(&SmolStr::from("alloc")).unwrap();
            let alloc_func = self.module.declare_func_in_func(alloc_id, builder.func);
            let size = builder.ins().iconst(types::I64, 24);
            let call = builder.ins().call(alloc_func, &[size]);
            let result_ptr = builder.inst_results(call)[0];

            // Store data, len, cap
            builder.ins().store(MemFlags::new(), ptr, result_ptr, 0);
            builder.ins().store(MemFlags::new(), len, result_ptr, 8);
            builder.ins().store(MemFlags::new(), len, result_ptr, 16); // cap = len

            return Ok(result_ptr);
        }

        // Multiple parts - concatenate them pairwise
        let concat_id = *self.functions.get(&SmolStr::from("string_concat")).unwrap();
        let concat_func = self.module.declare_func_in_func(concat_id, builder.func);

        let (mut result_ptr, mut result_len) = string_parts[0];

        for (ptr, len) in string_parts.into_iter().skip(1) {
            // Concatenate current result with next part
            let call = builder
                .ins()
                .call(concat_func, &[result_ptr, result_len, ptr, len]);
            let new_haira_string = builder.inst_results(call)[0];

            // Load new data pointer and length
            result_ptr = builder
                .ins()
                .load(self.ptr_type, MemFlags::new(), new_haira_string, 0);
            result_len = builder
                .ins()
                .load(types::I64, MemFlags::new(), new_haira_string, 8);
        }

        // Allocate final HairaString struct
        let alloc_id = *self.functions.get(&SmolStr::from("alloc")).unwrap();
        let alloc_func = self.module.declare_func_in_func(alloc_id, builder.func);
        let size = builder.ins().iconst(types::I64, 24);
        let call = builder.ins().call(alloc_func, &[size]);
        let final_ptr = builder.inst_results(call)[0];

        builder
            .ins()
            .store(MemFlags::new(), result_ptr, final_ptr, 0);
        builder
            .ins()
            .store(MemFlags::new(), result_len, final_ptr, 8);
        builder
            .ins()
            .store(MemFlags::new(), result_len, final_ptr, 16);

        Ok(final_ptr)
    }

    /// Compile a match expression.
    fn compile_match_expr(
        &mut self,
        match_expr: &haira_ast::MatchExpr,
        scope: &mut FunctionScope,
        builder: &mut FunctionBuilder,
    ) -> Result<Value, CodegenError> {
        // Compile the subject expression
        let subject_val = self.compile_expr(&match_expr.subject, scope, builder)?;

        // Create merge block for all arms to jump to with result
        let merge_block = builder.create_block();
        builder.append_block_param(merge_block, types::I64);

        // Create blocks for each arm body
        let mut arm_blocks: Vec<cranelift::prelude::Block> = Vec::new();
        for _ in &match_expr.arms {
            arm_blocks.push(builder.create_block());
        }

        // Default block (unreachable in exhaustive match, but needed)
        let default_block = builder.create_block();

        // Generate pattern matching logic as a chain of if-then-else
        // We stay in the current block and branch to arm blocks or continue checking
        let mut exhaustive = false;

        for (i, arm) in match_expr.arms.iter().enumerate() {
            let arm_block = arm_blocks[i];

            // Compile pattern check
            match &arm.pattern.node {
                haira_ast::Pattern::Wildcard => {
                    // Wildcard always matches - jump directly to arm
                    builder.ins().jump(arm_block, &[]);
                    // No more patterns will be checked after wildcard
                    exhaustive = true;
                    break;
                }
                haira_ast::Pattern::Literal(lit) => {
                    // Compare subject with literal value
                    let lit_val = self.compile_literal(lit, scope, builder)?;
                    let cmp = builder.ins().icmp(IntCC::Equal, subject_val, lit_val);

                    // Create a block for continuing to check next pattern
                    let next_check = builder.create_block();
                    builder.ins().brif(cmp, arm_block, &[], next_check, &[]);

                    // Continue in next_check block
                    builder.switch_to_block(next_check);
                    builder.seal_block(next_check);
                }
                haira_ast::Pattern::Identifier(name) => {
                    // Identifier pattern - binds the value to a variable
                    // Always matches, but first bind the variable
                    let var = scope.get_or_declare_var(name, builder);
                    builder.def_var(var, subject_val);
                    builder.ins().jump(arm_block, &[]);
                    // No more patterns will be checked after identifier (catch-all)
                    exhaustive = true;
                    break;
                }
                haira_ast::Pattern::Constructor { name, fields } => {
                    // Constructor pattern - for Option types like Some { value }
                    let next_check = builder.create_block();

                    if name == "Some" {
                        // Some is represented as (value << 1) | 1
                        // Check if low bit is 1 (is Some)
                        let one = builder.ins().iconst(types::I64, 1);
                        let zero = builder.ins().iconst(types::I64, 0);
                        let low_bit = builder.ins().band(subject_val, one);
                        let is_some = builder.ins().icmp(IntCC::NotEqual, low_bit, zero);

                        // If matches, bind fields in the arm block
                        // We need a separate block for binding
                        let bind_block = builder.create_block();
                        builder
                            .ins()
                            .brif(is_some, bind_block, &[], next_check, &[]);

                        builder.switch_to_block(bind_block);
                        builder.seal_block(bind_block);
                        if !fields.is_empty() {
                            let field_name = &fields[0].node;
                            let var = scope.get_or_declare_var(field_name, builder);
                            // Extract the value: (subject >> 1)
                            let one = builder.ins().iconst(types::I64, 1);
                            let extracted_val = builder.ins().ushr(subject_val, one);
                            builder.def_var(var, extracted_val);
                        }
                        builder.ins().jump(arm_block, &[]);
                    } else if name == "None" || name == "none" {
                        // None is represented as 0
                        let zero = builder.ins().iconst(types::I64, 0);
                        let is_none = builder.ins().icmp(IntCC::Equal, subject_val, zero);
                        builder.ins().brif(is_none, arm_block, &[], next_check, &[]);
                    } else {
                        // Other constructors - for now treat as always match
                        builder.ins().jump(arm_block, &[]);
                        exhaustive = true;
                        break;
                    }

                    builder.switch_to_block(next_check);
                    builder.seal_block(next_check);
                }
            }
        }

        // Jump to default block if match is not exhaustive
        if !exhaustive {
            builder.ins().jump(default_block, &[]);
        }

        // Default block - return 0 (should be unreachable in exhaustive match)
        builder.switch_to_block(default_block);
        builder.seal_block(default_block);
        let default_val = builder.ins().iconst(types::I64, 0);
        builder.ins().jump(merge_block, &[default_val]);

        // Compile arm bodies
        for (i, arm) in match_expr.arms.iter().enumerate() {
            let arm_block = arm_blocks[i];
            builder.switch_to_block(arm_block);
            builder.seal_block(arm_block);

            // Check guard if present
            if let Some(guard) = &arm.guard {
                let guard_val = self.compile_expr(guard, scope, builder)?;
                let guard_true_block = builder.create_block();
                let guard_false_block = default_block;
                builder
                    .ins()
                    .brif(guard_val, guard_true_block, &[], guard_false_block, &[]);
                builder.switch_to_block(guard_true_block);
                builder.seal_block(guard_true_block);
            }

            // Compile arm body
            let arm_val = match &arm.body {
                haira_ast::MatchArmBody::Expr(expr) => self.compile_expr(expr, scope, builder)?,
                haira_ast::MatchArmBody::Block(block) => self
                    .compile_block(block, scope, builder)?
                    .unwrap_or_else(|| builder.ins().iconst(types::I64, 0)),
            };

            builder.ins().jump(merge_block, &[arm_val]);
        }

        // Switch to merge block
        builder.switch_to_block(merge_block);
        builder.seal_block(merge_block);

        Ok(builder.block_params(merge_block)[0])
    }

    /// Compile a binary operation.
    fn compile_binary_op(
        &mut self,
        op: &BinaryOp,
        left: Value,
        right: Value,
        builder: &mut FunctionBuilder,
    ) -> Result<Value, CodegenError> {
        let result = match op {
            BinaryOp::Add => builder.ins().iadd(left, right),
            BinaryOp::Sub => builder.ins().isub(left, right),
            BinaryOp::Mul => builder.ins().imul(left, right),
            BinaryOp::Div => builder.ins().sdiv(left, right),
            BinaryOp::Mod => builder.ins().srem(left, right),
            BinaryOp::Eq => {
                let cmp = builder.ins().icmp(IntCC::Equal, left, right);
                builder.ins().uextend(types::I64, cmp)
            }
            BinaryOp::Ne => {
                let cmp = builder.ins().icmp(IntCC::NotEqual, left, right);
                builder.ins().uextend(types::I64, cmp)
            }
            BinaryOp::Lt => {
                let cmp = builder.ins().icmp(IntCC::SignedLessThan, left, right);
                builder.ins().uextend(types::I64, cmp)
            }
            BinaryOp::Le => {
                let lt = builder.ins().icmp(IntCC::SignedLessThan, left, right);
                let eq = builder.ins().icmp(IntCC::Equal, left, right);
                let cmp = builder.ins().bor(lt, eq);
                builder.ins().uextend(types::I64, cmp)
            }
            BinaryOp::Gt => {
                let cmp = builder.ins().icmp(IntCC::SignedGreaterThan, left, right);
                builder.ins().uextend(types::I64, cmp)
            }
            BinaryOp::Ge => {
                let gt = builder.ins().icmp(IntCC::SignedGreaterThan, left, right);
                let eq = builder.ins().icmp(IntCC::Equal, left, right);
                let cmp = builder.ins().bor(gt, eq);
                builder.ins().uextend(types::I64, cmp)
            }
            BinaryOp::And => builder.ins().band(left, right),
            BinaryOp::Or => builder.ins().bor(left, right),
        };
        Ok(result)
    }

    /// Compile a unary operation.
    fn compile_unary_op(
        &mut self,
        op: &UnaryOp,
        operand: Value,
        builder: &mut FunctionBuilder,
    ) -> Result<Value, CodegenError> {
        let result = match op {
            UnaryOp::Neg => builder.ins().ineg(operand),
            UnaryOp::Not => {
                let one = builder.ins().iconst(types::I64, 1);
                builder.ins().bxor(operand, one)
            }
        };
        Ok(result)
    }

    /// Compile a function call.
    fn compile_call(
        &mut self,
        call: &haira_ast::CallExpr,
        scope: &mut FunctionScope,
        builder: &mut FunctionBuilder,
    ) -> Result<Value, CodegenError> {
        // Get function name
        let func_name = match &call.callee.node {
            ExprKind::Identifier(name) => name.clone(),
            _ => {
                return Err(CodegenError::Unsupported(
                    "Only direct function calls are supported".to_string(),
                ))
            }
        };

        // Handle print specially - detect argument types
        if func_name.as_str() == "print" {
            return self.compile_print_call(call, scope, builder);
        }

        // Handle err() - set error and return error value
        if func_name.as_str() == "err" {
            let set_error_id = *self.functions.get(&SmolStr::from("set_error")).unwrap();
            let set_error_func = self.module.declare_func_in_func(set_error_id, builder.func);

            // Get error value from argument (default to 1 if no arg)
            let err_val = if call.args.is_empty() {
                builder.ins().iconst(types::I64, 1)
            } else {
                self.compile_expr(&call.args[0].value, scope, builder)?
            };

            builder.ins().call(set_error_func, &[err_val]);
            return Ok(err_val);
        }

        // Handle channel() - create a new channel
        if func_name.as_str() == "channel" {
            let channel_new_id = *self.functions.get(&SmolStr::from("channel_new")).unwrap();
            let channel_new_func = self
                .module
                .declare_func_in_func(channel_new_id, builder.func);

            // Get capacity from argument (default to 1 if no arg)
            let capacity = if call.args.is_empty() {
                builder.ins().iconst(types::I64, 1)
            } else {
                self.compile_expr(&call.args[0].value, scope, builder)?
            };

            let call_inst = builder.ins().call(channel_new_func, &[capacity]);
            return Ok(builder.inst_results(call_inst)[0]);
        }

        // Handle spawn_fn(func_name) - spawn a function in a new thread
        if func_name.as_str() == "spawn_fn" {
            if call.args.is_empty() {
                return Err(CodegenError::Unsupported(
                    "spawn_fn requires a function name argument".to_string(),
                ));
            }

            // Get the function name from the argument (should be an identifier)
            let target_func_name = match &call.args[0].value.node {
                ExprKind::Identifier(name) => name.clone(),
                _ => {
                    return Err(CodegenError::Unsupported(
                        "spawn_fn argument must be a function name".to_string(),
                    ));
                }
            };

            // Look up the target function
            let target_func_id = *self
                .functions
                .get(&target_func_name)
                .ok_or_else(|| CodegenError::UndefinedFunction(target_func_name.to_string()))?;

            // Get function address
            let local_target = self
                .module
                .declare_func_in_func(target_func_id, builder.func);
            let func_ptr = builder.ins().func_addr(self.ptr_type, local_target);

            // Call haira_spawn with function pointer
            let spawn_id = *self.functions.get(&SmolStr::from("spawn_thread")).unwrap();
            let spawn_func = self.module.declare_func_in_func(spawn_id, builder.func);
            let call_inst = builder.ins().call(spawn_func, &[func_ptr]);
            return Ok(builder.inst_results(call_inst)[0]);
        }

        // Look up function
        let func_id = *self
            .functions
            .get(&func_name)
            .ok_or_else(|| CodegenError::UndefinedFunction(func_name.to_string()))?;

        let local_callee = self.module.declare_func_in_func(func_id, builder.func);

        // Compile arguments
        let mut args = Vec::new();
        for arg in &call.args {
            args.push(self.compile_expr(&arg.value, scope, builder)?);
        }

        let call_inst = builder.ins().call(local_callee, &args);
        let results = builder.inst_results(call_inst);

        if results.is_empty() {
            Ok(builder.ins().iconst(types::I64, 0))
        } else {
            Ok(results[0])
        }
    }

    /// Compile a print call with type detection.
    fn compile_print_call(
        &mut self,
        call: &haira_ast::CallExpr,
        scope: &mut FunctionScope,
        builder: &mut FunctionBuilder,
    ) -> Result<Value, CodegenError> {
        if call.args.is_empty() {
            // Just print newline
            let println_id = *self.functions.get(&SmolStr::from("println")).unwrap();
            let local_callee = self.module.declare_func_in_func(println_id, builder.func);
            builder.ins().call(local_callee, &[]);
            return Ok(builder.ins().iconst(types::I64, 0));
        }

        let arg = &call.args[0].value;

        // Detect type from expression
        match &arg.node {
            ExprKind::Literal(Literal::String(s)) => {
                // String literal - call haira_print with pointer and length
                let print_id = *self.functions.get(&SmolStr::from("print")).unwrap();
                let local_callee = self.module.declare_func_in_func(print_id, builder.func);

                let data_id = self.define_string(s)?;
                let local_id = self.module.declare_data_in_func(data_id, builder.func);
                let ptr = builder.ins().symbol_value(self.ptr_type, local_id);
                let len = builder.ins().iconst(types::I64, s.len() as i64);

                builder.ins().call(local_callee, &[ptr, len]);

                // Print newline
                let println_id = *self.functions.get(&SmolStr::from("println")).unwrap();
                let local_callee = self.module.declare_func_in_func(println_id, builder.func);
                builder.ins().call(local_callee, &[]);
            }
            ExprKind::Literal(Literal::Int(_)) => {
                let val = self.compile_expr(arg, scope, builder)?;
                let print_int_id = *self.functions.get(&SmolStr::from("print_int")).unwrap();
                let local_callee = self.module.declare_func_in_func(print_int_id, builder.func);
                builder.ins().call(local_callee, &[val]);

                let println_id = *self.functions.get(&SmolStr::from("println")).unwrap();
                let local_callee = self.module.declare_func_in_func(println_id, builder.func);
                builder.ins().call(local_callee, &[]);
            }
            ExprKind::Literal(Literal::Float(_)) => {
                let val = self.compile_expr(arg, scope, builder)?;
                let print_float_id = *self.functions.get(&SmolStr::from("print_float")).unwrap();
                let local_callee = self
                    .module
                    .declare_func_in_func(print_float_id, builder.func);
                builder.ins().call(local_callee, &[val]);

                let println_id = *self.functions.get(&SmolStr::from("println")).unwrap();
                let local_callee = self.module.declare_func_in_func(println_id, builder.func);
                builder.ins().call(local_callee, &[]);
            }
            ExprKind::Literal(Literal::Bool(_)) => {
                let val = self.compile_expr(arg, scope, builder)?;
                let print_bool_id = *self.functions.get(&SmolStr::from("print_bool")).unwrap();
                let local_callee = self
                    .module
                    .declare_func_in_func(print_bool_id, builder.func);
                builder.ins().call(local_callee, &[val]);

                let println_id = *self.functions.get(&SmolStr::from("println")).unwrap();
                let local_callee = self.module.declare_func_in_func(println_id, builder.func);
                builder.ins().call(local_callee, &[]);
            }
            ExprKind::Literal(Literal::InterpolatedString(_)) => {
                // Interpolated string returns a HairaString* (ptr to struct with data, len, cap)
                let haira_string_ptr = self.compile_expr(arg, scope, builder)?;

                // Load data pointer (offset 0) and len (offset 8)
                let data_ptr =
                    builder
                        .ins()
                        .load(self.ptr_type, MemFlags::new(), haira_string_ptr, 0);
                let len = builder
                    .ins()
                    .load(types::I64, MemFlags::new(), haira_string_ptr, 8);

                // Call haira_print with data and length
                let print_id = *self.functions.get(&SmolStr::from("print")).unwrap();
                let local_callee = self.module.declare_func_in_func(print_id, builder.func);
                builder.ins().call(local_callee, &[data_ptr, len]);

                // Print newline
                let println_id = *self.functions.get(&SmolStr::from("println")).unwrap();
                let local_callee = self.module.declare_func_in_func(println_id, builder.func);
                builder.ins().call(local_callee, &[]);
            }
            _ => {
                // Assume integer for other expressions
                let val = self.compile_expr(arg, scope, builder)?;
                let print_int_id = *self.functions.get(&SmolStr::from("print_int")).unwrap();
                let local_callee = self.module.declare_func_in_func(print_int_id, builder.func);
                builder.ins().call(local_callee, &[val]);

                let println_id = *self.functions.get(&SmolStr::from("println")).unwrap();
                let local_callee = self.module.declare_func_in_func(println_id, builder.func);
                builder.ins().call(local_callee, &[]);
            }
        }

        Ok(builder.ins().iconst(types::I64, 0))
    }
}

/// Scope for variables within a function.
/// Uses Cranelift Variables for proper SSA handling.
struct FunctionScope {
    /// Map of variable names to Cranelift Variables.
    variables: HashMap<SmolStr, Variable>,
    /// Counter for generating unique variable indices.
    next_var: usize,
    #[allow(dead_code)]
    ptr_type: Type,
}

impl FunctionScope {
    fn new(ptr_type: Type) -> Self {
        Self {
            variables: HashMap::new(),
            next_var: 0,
            ptr_type,
        }
    }

    /// Declare a new Cranelift variable.
    fn declare_var(&mut self, name: &SmolStr, builder: &mut FunctionBuilder) -> Variable {
        let var = Variable::new(self.next_var);
        self.next_var += 1;
        builder.declare_var(var, types::I64);
        self.variables.insert(name.clone(), var);
        var
    }

    /// Get an existing variable or declare a new one.
    fn get_or_declare_var(&mut self, name: &SmolStr, builder: &mut FunctionBuilder) -> Variable {
        if let Some(&var) = self.variables.get(name) {
            var
        } else {
            self.declare_var(name, builder)
        }
    }

    /// Get an existing variable.
    fn get_var(&self, name: &SmolStr) -> Option<Variable> {
        self.variables.get(name).copied()
    }
}

/// Compile AST to executable.
pub fn compile_to_executable(
    ast: &SourceFile,
    output_path: &Path,
    _options: CodegenOptions,
) -> Result<(), CodegenError> {
    let mut compiler = Compiler::new()?;
    compiler.compile(ast)?;

    let object_bytes = compiler.finish();

    // Write object file
    let obj_path = output_path.with_extension("o");
    std::fs::write(&obj_path, &object_bytes)?;

    // Link with runtime
    link_executable(&obj_path, output_path)?;

    // Clean up object file
    std::fs::remove_file(&obj_path).ok();

    Ok(())
}

/// Link object file with runtime to create executable.
fn link_executable(obj_path: &Path, output_path: &Path) -> Result<(), CodegenError> {
    // Find the haira-runtime staticlib
    let runtime_path = find_runtime_library()?;

    // Determine platform-specific linker flags
    #[cfg(target_os = "macos")]
    let platform_libs = vec!["-framework", "Security", "-framework", "CoreFoundation"];

    #[cfg(target_os = "linux")]
    let platform_libs = vec!["-ldl", "-lm"];

    #[cfg(target_os = "windows")]
    let platform_libs = vec!["-lws2_32", "-luserenv"];

    // Use cc to link with pthread for concurrency support
    let mut cmd = Command::new("cc");
    cmd.arg(obj_path)
        .arg(&runtime_path)
        .arg("-o")
        .arg(output_path)
        .arg("-lpthread");

    // Add platform-specific libraries
    for lib in &platform_libs {
        cmd.arg(lib);
    }

    let status = cmd.status()?;

    if !status.success() {
        return Err(CodegenError::LinkerError("Linker failed".to_string()));
    }

    Ok(())
}

/// Find the haira-runtime static library.
fn find_runtime_library() -> Result<std::path::PathBuf, CodegenError> {
    // Try to find the runtime library in common locations

    // 1. Check if HAIRA_RUNTIME_LIB env var is set
    if let Ok(path) = std::env::var("HAIRA_RUNTIME_LIB") {
        let path = std::path::PathBuf::from(path);
        if path.exists() {
            return Ok(path);
        }
    }

    // 2. Check relative to the executable (for installed binaries)
    if let Ok(exe_path) = std::env::current_exe() {
        if let Some(exe_dir) = exe_path.parent() {
            // Check ../lib/libhaira_runtime.a
            let lib_path = exe_dir.join("../lib/libhaira_runtime.a");
            if lib_path.exists() {
                return Ok(lib_path);
            }

            // Check in same directory
            let lib_path = exe_dir.join("libhaira_runtime.a");
            if lib_path.exists() {
                return Ok(lib_path);
            }
        }
    }

    // 3. Check in target directory (for development)
    let target_dirs = [
        "target/release/libhaira_runtime.a",
        "target/debug/libhaira_runtime.a",
        "../target/release/libhaira_runtime.a",
        "../target/debug/libhaira_runtime.a",
        "../../target/release/libhaira_runtime.a",
        "../../target/debug/libhaira_runtime.a",
    ];

    for dir in &target_dirs {
        let path = std::path::PathBuf::from(dir);
        if path.exists() {
            return Ok(path);
        }
    }

    // 4. Check CARGO_MANIFEST_DIR for development builds
    if let Ok(manifest_dir) = std::env::var("CARGO_MANIFEST_DIR") {
        let workspace_root = std::path::Path::new(&manifest_dir)
            .parent()
            .and_then(|p| p.parent());

        if let Some(root) = workspace_root {
            for profile in &["release", "debug"] {
                let lib_path = root.join("target").join(profile).join("libhaira_runtime.a");
                if lib_path.exists() {
                    return Ok(lib_path);
                }
            }
        }
    }

    Err(CodegenError::LinkerError(
        "Could not find haira-runtime library. \
         Build with `cargo build -p haira-runtime --release` or \
         set HAIRA_RUNTIME_LIB environment variable."
            .to_string(),
    ))
}
