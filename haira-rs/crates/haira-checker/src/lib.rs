//! Haira type checker -- two-pass semantic analysis.
//!
//! This crate ports the Go type checker from `compiler/internal/checker/` (types.go,
//! env.go, checker.go) to Rust. It performs:
//!
//! - **Pass 1 (`register_globals`)**: Register all struct defs, enum defs, function
//!   signatures, tool/workflow signatures, provider/agent names, and top-level variables.
//! - **Pass 2 (`check_bodies`)**: Type-check function bodies, method bodies, tool bodies,
//!   workflow bodies, and test bodies.
//!
//! The public entry point is [`check`].

use std::collections::{HashMap, HashSet};
use std::fmt;

use haira_ast::{
    self as ast, AssignPath, Block, ElseBranch, ExprKind, ForPattern, ItemKind, Literal,
    MatchArmBody, SourceFile, Span, Spanned, StmtKind,
};
use haira_errors::{Diagnostic, Level};

// ===========================================================================
// Type enum (replaces Go Type interface)
// ===========================================================================

/// A checked type in the Haira type system.
#[derive(Debug, Clone)]
pub enum Type {
    Int,
    Float,
    String,
    Bool,
    Any,
    Void,
    Error,
    List {
        elem: Box<Type>,
    },
    Map {
        key: Box<Type>,
        value: Box<Type>,
    },
    Struct {
        name: std::string::String,
        fields: HashMap<std::string::String, Type>,
    },
    Enum {
        name: std::string::String,
        variants: Vec<std::string::String>,
    },
    Func {
        params: Vec<Type>,
        ret: Box<Type>,
    },
}

impl PartialEq for Type {
    fn eq(&self, other: &Self) -> bool {
        type_equals(self, other)
    }
}

impl Eq for Type {}

impl fmt::Display for Type {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Type::Int => write!(f, "int"),
            Type::Float => write!(f, "float"),
            Type::String => write!(f, "string"),
            Type::Bool => write!(f, "bool"),
            Type::Any => write!(f, "any"),
            Type::Void => write!(f, "void"),
            Type::Error => write!(f, "error"),
            Type::List { elem } => write!(f, "[{}]", elem),
            Type::Map { key, value } => write!(f, "{{{}: {}}}", key, value),
            Type::Struct { name, .. } => write!(f, "{}", name),
            Type::Enum { name, .. } => write!(f, "{}", name),
            Type::Func { .. } => write!(f, "fn"),
        }
    }
}

/// Checks structural equality of two types. `Any` matches everything.
pub fn type_equals(a: &Type, b: &Type) -> bool {
    // Any matches everything
    if matches!(a, Type::Any) || matches!(b, Type::Any) {
        return true;
    }
    match (a, b) {
        (Type::Int, Type::Int) => true,
        (Type::Float, Type::Float) => true,
        (Type::String, Type::String) => true,
        (Type::Bool, Type::Bool) => true,
        (Type::Void, Type::Void) => true,
        (Type::Error, Type::Error) => true,
        (Type::List { elem: a_elem }, Type::List { elem: b_elem }) => {
            type_equals(a_elem, b_elem)
        }
        (
            Type::Map {
                key: ak,
                value: av,
            },
            Type::Map {
                key: bk,
                value: bv,
            },
        ) => type_equals(ak, bk) && type_equals(av, bv),
        (Type::Struct { name: a_name, .. }, Type::Struct { name: b_name, .. }) => {
            a_name == b_name
        }
        (Type::Enum { name: a_name, .. }, Type::Enum { name: b_name, .. }) => a_name == b_name,
        (
            Type::Func {
                params: ap,
                ret: ar,
            },
            Type::Func {
                params: bp,
                ret: br,
            },
        ) => {
            if ap.len() != bp.len() {
                return false;
            }
            for (pa, pb) in ap.iter().zip(bp.iter()) {
                if !type_equals(pa, pb) {
                    return false;
                }
            }
            type_equals(ar, br)
        }
        _ => false,
    }
}

fn is_any(t: &Type) -> bool {
    matches!(t, Type::Any)
}

// ===========================================================================
// FuncSig helper (stored in Env)
// ===========================================================================

/// A function signature stored in the environment.
#[derive(Debug, Clone)]
pub struct FuncSig {
    pub params: Vec<Type>,
    pub ret: Type,
}

// ===========================================================================
// VarInfo
// ===========================================================================

/// Holds the type and mutability of a variable.
#[derive(Debug, Clone)]
pub struct VarInfo {
    pub ty: Type,
    pub is_const: bool,
}

// ===========================================================================
// Env (scoped symbol table)
// ===========================================================================

/// A scoped symbol table for type checking.
pub struct Env {
    parent: Option<Box<Env>>,
    vars: HashMap<std::string::String, VarInfo>,
    types: HashMap<std::string::String, Type>,
    funcs: HashMap<std::string::String, FuncSig>,
}

impl Env {
    /// Creates a root environment with stdlib pre-registered.
    pub fn new() -> Self {
        let mut env = Env {
            parent: None,
            vars: HashMap::new(),
            types: HashMap::new(),
            funcs: HashMap::new(),
        };
        env.register_stdlib();
        env
    }

    /// Creates a new child scope with this env as parent.
    pub fn child(self) -> Self {
        Env {
            parent: Some(Box::new(self)),
            vars: HashMap::new(),
            types: HashMap::new(),
            funcs: HashMap::new(),
        }
    }

    /// Defines a mutable variable in the current scope.
    pub fn define_var(&mut self, name: impl Into<std::string::String>, ty: Type) {
        self.vars.insert(
            name.into(),
            VarInfo {
                ty,
                is_const: false,
            },
        );
    }

    /// Defines an immutable (const) variable in the current scope.
    pub fn define_const(&mut self, name: impl Into<std::string::String>, ty: Type) {
        self.vars.insert(
            name.into(),
            VarInfo {
                ty,
                is_const: true,
            },
        );
    }

    /// Looks up a variable in the current scope and parents.
    pub fn lookup_var(&self, name: &str) -> Option<&Type> {
        if let Some(vi) = self.vars.get(name) {
            return Some(&vi.ty);
        }
        if let Some(ref parent) = self.parent {
            return parent.lookup_var(name);
        }
        None
    }

    /// Looks up full variable info (type + const flag).
    pub fn lookup_var_info(&self, name: &str) -> Option<&VarInfo> {
        if let Some(vi) = self.vars.get(name) {
            return Some(vi);
        }
        if let Some(ref parent) = self.parent {
            return parent.lookup_var_info(name);
        }
        None
    }

    /// Registers a type definition (struct/enum).
    pub fn define_type(&mut self, name: impl Into<std::string::String>, ty: Type) {
        self.types.insert(name.into(), ty);
    }

    /// Looks up a type by name.
    pub fn lookup_type(&self, name: &str) -> Option<&Type> {
        if let Some(ty) = self.types.get(name) {
            return Some(ty);
        }
        if let Some(ref parent) = self.parent {
            return parent.lookup_type(name);
        }
        None
    }

    /// Registers a function signature.
    pub fn define_func(&mut self, name: impl Into<std::string::String>, sig: FuncSig) {
        self.funcs.insert(name.into(), sig);
    }

    /// Looks up a function by name.
    pub fn lookup_func(&self, name: &str) -> Option<&FuncSig> {
        if let Some(sig) = self.funcs.get(name) {
            return Some(sig);
        }
        if let Some(ref parent) = self.parent {
            return parent.lookup_func(name);
        }
        None
    }

    /// Register all standard library functions.
    fn register_stdlib(&mut self) {
        let any = Type::Any;
        let str_ = Type::String;
        let int = Type::Int;
        let float = Type::Float;
        let bool_ = Type::Bool;
        let void = Type::Void;
        let str_list = Type::List {
            elem: Box::new(Type::String),
        };
        let any_list = Type::List {
            elem: Box::new(Type::Any),
        };

        // Helper macro to reduce boilerplate
        macro_rules! func {
            ($name:expr, [$($p:expr),*] => $ret:expr) => {
                self.funcs.insert($name.into(), FuncSig {
                    params: vec![$($p.clone()),*],
                    ret: $ret.clone(),
                });
            };
        }

        // io module functions
        func!("io.println", [any] => void);
        func!("io.print", [any] => void);
        func!("io.printf", [str_, any] => void);
        func!("io.readln", [] => str_);
        func!("io.eprintln", [any] => void);
        func!("io.eprintf", [str_, any] => void);

        // http module functions
        let resp_tuple = any.clone(); // (Response, error) -- simplified
        func!("http.get", [str_] => resp_tuple);
        func!("http.get_with_headers", [str_, any] => resp_tuple);
        func!("http.post", [str_, any] => resp_tuple);
        func!("http.post_with_headers", [str_, any, any] => resp_tuple);
        func!("http.put", [str_, any] => resp_tuple);
        func!("http.put_with_headers", [str_, any, any] => resp_tuple);
        func!("http.delete", [str_] => resp_tuple);
        func!("http.delete_with_headers", [str_, any] => resp_tuple);

        // json module functions
        func!("json.marshal", [any] => any);
        func!("json.unmarshal", [str_] => any);
        func!("json.marshal_pretty", [any] => any);
        func!("json.parse", [str_] => any);

        // string module functions
        func!("string.len", [str_] => int);
        func!("string.is_empty", [str_] => bool_);
        func!("string.trim", [str_] => str_);
        func!("string.trim_left", [str_] => str_);
        func!("string.trim_right", [str_] => str_);
        func!("string.to_upper", [str_] => str_);
        func!("string.to_lower", [str_] => str_);
        func!("string.contains", [str_, str_] => bool_);
        func!("string.starts_with", [str_, str_] => bool_);
        func!("string.ends_with", [str_, str_] => bool_);
        func!("string.index_of", [str_, str_] => int);
        func!("string.last_index_of", [str_, str_] => int);
        func!("string.split", [str_, str_] => str_list);
        func!("string.join", [any, str_] => str_);
        func!("string.substring", [str_, int, int] => str_);
        func!("string.char_at", [str_, int] => str_);
        func!("string.replace", [str_, str_, str_] => str_);
        func!("string.replace_all", [str_, str_, str_] => str_);
        func!("string.repeat", [str_, int] => str_);

        // regex module functions
        func!("regex.is_match", [str_, str_] => bool_);
        func!("regex.find", [str_, str_] => str_);
        func!("regex.find_all", [str_, str_] => any_list);
        func!("regex.replace", [str_, str_, str_] => str_);
        func!("regex.replace_all", [str_, str_, str_] => str_);
        func!("regex.captures", [str_, str_] => any_list);
        func!("regex.split", [str_, str_] => any_list);

        // math module functions
        func!("math.abs", [any] => float);
        func!("math.min", [any, any] => float);
        func!("math.max", [any, any] => float);
        func!("math.clamp", [any, any, any] => float);
        func!("math.floor", [any] => float);
        func!("math.ceil", [any] => float);
        func!("math.round", [any] => float);
        func!("math.trunc", [any] => float);
        func!("math.pow", [any, any] => float);
        func!("math.sqrt", [any] => float);
        func!("math.cbrt", [any] => float);
        func!("math.exp", [any] => float);
        func!("math.log", [any] => float);
        func!("math.log10", [any] => float);
        func!("math.log2", [any] => float);
        func!("math.sin", [any] => float);
        func!("math.cos", [any] => float);
        func!("math.tan", [any] => float);
        func!("math.asin", [any] => float);
        func!("math.acos", [any] => float);
        func!("math.atan", [any] => float);
        func!("math.atan2", [any, any] => float);
        func!("math.random", [] => float);
        func!("math.random_int", [any, any] => int);
        func!("math.pi", [] => float);
        func!("math.e", [] => float);

        // conv module functions
        func!("conv.int_to_string", [any] => str_);
        func!("conv.float_to_string", [any] => str_);
        func!("conv.bool_to_string", [any] => str_);
        func!("conv.string_to_int", [str_] => any);
        func!("conv.string_to_float", [str_] => any);
        func!("conv.string_to_bool", [str_] => any);
        func!("conv.int_to_float", [any] => float);
        func!("conv.float_to_int", [any] => int);
        func!("conv.int_to_hex", [any] => str_);
        func!("conv.int_to_binary", [any] => str_);
        func!("conv.int_to_octal", [any] => str_);
        func!("conv.hex_to_int", [str_] => any);

        // array module functions
        func!("array.len", [any] => int);
        func!("array.is_empty", [any] => bool_);
        func!("array.first", [any] => any);
        func!("array.last", [any] => any);
        func!("array.get", [any, int] => any);
        func!("array.push", [any, any] => any_list);
        func!("array.pop", [any] => any);
        func!("array.insert", [any, int, any] => any_list);
        func!("array.remove", [any, int] => any_list);
        func!("array.slice", [any, int, int] => any_list);
        func!("array.take", [any, int] => any_list);
        func!("array.drop", [any, int] => any_list);
        func!("array.contains", [any, any] => bool_);
        func!("array.index_of", [any, any] => int);
        func!("array.reverse", [any] => any_list);
        func!("array.concat", [any, any] => any_list);
        func!("array.flatten", [any] => any_list);
        func!("array.unique", [any] => any_list);
        func!("array.sort", [any] => any_list);
        func!("array.join", [any, str_] => str_);
        func!("array.map", [any, any] => any_list);
        func!("array.filter", [any, any] => any_list);
        func!("array.reduce", [any, any, any] => any);
        func!("array.find", [any, any] => any);
        func!("array.find_index", [any, any] => int);
        func!("array.sort_by", [any, any] => any_list);
        func!("array.every", [any, any] => bool_);
        func!("array.some", [any, any] => bool_);
        func!("array.for_each", [any, any] => void);
        func!("array.flat_map", [any, any] => any_list);

        // map module functions
        func!("map.len", [any] => int);
        func!("map.is_empty", [any] => bool_);
        func!("map.get", [any, str_] => any);
        func!("map.has", [any, str_] => bool_);
        func!("map.set", [any, str_, any] => any);
        func!("map.remove", [any, str_] => any);
        func!("map.keys", [any] => any_list);
        func!("map.values", [any] => any_list);
        func!("map.entries", [any] => any_list);
        func!("map.merge", [any, any] => any);
        func!("map.contains_value", [any, any] => bool_);

        // postgres module
        func!("postgres.connect", [str_] => any);

        // slack module
        func!("slack.send", [str_, str_, str_] => any);

        // excel module
        func!("excel.open", [str_] => any);

        // time module
        func!("time.sleep", [any] => void);
        func!("time.now", [] => str_);
        func!("time.slug", [] => str_);

        // fs module functions
        func!("fs.read_file", [str_] => any);
        func!("fs.write_file", [str_, str_] => any);
        func!("fs.append_file", [str_, str_] => any);
        func!("fs.exists", [str_] => bool_);
        func!("fs.remove", [str_] => any);
        func!("fs.remove_all", [str_] => any);
        func!("fs.rename", [str_, str_] => any);
        func!("fs.copy", [str_, str_] => any);
        func!("fs.mkdir", [str_] => any);
        func!("fs.mkdir_all", [str_] => any);
        func!("fs.read_dir", [str_] => any_list);
        func!("fs.stat", [str_] => any);

        // Built-in functions
        func!("env", [str_] => str_);
        func!("len", [any] => int);
        func!("keys", [any] => str_list);
        func!("join", [any, str_] => str_);
        func!("panic", [any] => any);
    }
}

impl Default for Env {
    fn default() -> Self {
        Self::new()
    }
}

// ===========================================================================
// TypeInfo -- output of the checker
// ===========================================================================

/// Holds inferred types keyed by AST span.
#[derive(Debug, Clone, Default)]
pub struct TypeInfo {
    /// Maps expression spans to their inferred type.
    pub expr_types: HashMap<Span, Type>,
    /// Maps variable name spans to their inferred type.
    pub var_types: HashMap<Span, Type>,
}

// ===========================================================================
// stdlib module set
// ===========================================================================

fn is_stdlib_module(name: &str) -> bool {
    matches!(
        name,
        "io" | "http"
            | "json"
            | "string"
            | "regex"
            | "math"
            | "conv"
            | "array"
            | "map"
            | "time"
            | "env"
            | "postgres"
            | "slack"
            | "excel"
            | "log"
            | "mcp"
            | "ui"
            | "vector"
            | "observe"
            | "fs"
            | "gitlab"
            | "github"
            | "langfuse"
            | "algolia"
            | "meilisearch"
            | "store"
    )
}

// ===========================================================================
// Valid field sets for agent/provider/ui validation
// ===========================================================================

fn is_valid_provider_field(name: &str) -> bool {
    matches!(
        name,
        "api_key"
            | "model"
            | "endpoint"
            | "api_version"
            | "backend"
            | "host"
            | "account_id"
            | "temperature"
            | "max_tokens"
            | "input_token_cost"
            | "output_token_cost"
            | "transport"
            | "command"
            | "args"
            | "env"
            | "headers"
    )
}

fn is_valid_agent_field(name: &str) -> bool {
    matches!(
        name,
        "provider"
            | "system"
            | "tools"
            | "handoffs"
            | "mcp"
            | "temperature"
            | "max_tokens"
            | "max_steps"
            | "memory"
            | "output"
            | "ui"
            | "timeout"
    )
}

fn is_valid_ui_component(name: &str) -> bool {
    matches!(
        name,
        "StatusCard"
            | "Confirm"
            | "Choices"
            | "Table"
            | "CodeBlock"
            | "Diff"
            | "KeyValue"
            | "Progress"
            | "Chart"
    )
}

// ===========================================================================
// Public entry point
// ===========================================================================

/// Performs type checking on a source file and returns type info + diagnostics.
pub fn check(file: &SourceFile) -> (TypeInfo, Vec<Diagnostic>) {
    let mut checker = Checker {
        env: Env::new(),
        info: TypeInfo::default(),
        diags: Vec::new(),
        file: std::string::String::new(),
        in_method: false,
        in_test: false,
        in_try: false,
        return_ty: None,
        agents: HashSet::new(),
        providers: HashSet::new(),
    };

    // Pass 1: register all global declarations
    checker.register_globals(file);

    // Pass 2: check function bodies
    checker.check_bodies(file);

    (checker.info, checker.diags)
}

// ===========================================================================
// Checker struct
// ===========================================================================

struct Checker {
    env: Env,
    info: TypeInfo,
    diags: Vec<Diagnostic>,
    file: std::string::String,
    in_method: bool,
    in_test: bool,
    in_try: bool,
    return_ty: Option<Type>,
    agents: HashSet<std::string::String>,
    providers: HashSet<std::string::String>,
}

impl Checker {
    // -----------------------------------------------------------------------
    // Diagnostic helpers
    // -----------------------------------------------------------------------

    /// Push a new child scope onto the environment stack.
    fn push_scope(&mut self) {
        let parent = std::mem::replace(&mut self.env, Env::new());
        self.env = Env {
            parent: Some(Box::new(parent)),
            vars: HashMap::new(),
            types: HashMap::new(),
            funcs: HashMap::new(),
        };
    }

    /// Pop the current scope, restoring the parent.
    fn pop_scope(&mut self) {
        let current = std::mem::replace(&mut self.env, Env::new());
        self.env = match current.parent {
            Some(parent) => *parent,
            None => {
                // Should never happen -- push_scope always sets a parent.
                // Return an empty env as a fallback.
                Env::new()
            }
        };
    }

    fn add_error(&mut self, msg: impl Into<std::string::String>, span: Span) {
        self.diags.push(Diagnostic {
            level: Level::Error,
            message: msg.into(),
            span,
            file: self.file.clone(),
            hint: std::string::String::new(),
        });
    }

    fn add_warning(
        &mut self,
        msg: impl Into<std::string::String>,
        span: Span,
        hint: impl Into<std::string::String>,
    ) {
        self.diags.push(Diagnostic {
            level: Level::Warning,
            message: msg.into(),
            span,
            file: self.file.clone(),
            hint: hint.into(),
        });
    }

    // -----------------------------------------------------------------------
    // Enum exhaustiveness checking
    // -----------------------------------------------------------------------

    fn check_enum_exhaustiveness(
        &mut self,
        enum_ty: &Type,
        arms: &[ast::MatchArm],
        span: Span,
    ) {
        let (enum_name, variants) = match enum_ty {
            Type::Enum { name, variants } => (name.as_str(), variants.as_slice()),
            _ => return,
        };

        let mut has_wildcard = false;
        let mut covered = HashSet::new();
        for arm in arms {
            collect_covered_variants(&arm.pattern.node, enum_name, &mut covered, &mut has_wildcard);
        }
        if has_wildcard {
            return;
        }
        let mut missing = Vec::new();
        for v in variants {
            let full_name = format!("{}{}", enum_name, v);
            if !covered.contains(full_name.as_str()) {
                missing.push(v.as_str());
            }
        }
        if !missing.is_empty() {
            let hint = "add a wildcard '_' arm or handle all variants";
            let msg = format!(
                "non-exhaustive match on {}: missing {}",
                enum_name,
                missing.join(", ")
            );
            self.add_warning(msg, span, hint);
        }
    }

    // -----------------------------------------------------------------------
    // Pass 1: Register global declarations
    // -----------------------------------------------------------------------

    fn register_globals(&mut self, file: &SourceFile) {
        for item in &file.items {
            match &item.node {
                ItemKind::TypeDef(td) => {
                    let mut fields = HashMap::new();
                    for f in &td.fields {
                        fields.insert(
                            f.name.node.clone(),
                            self.resolve_type_expr(f.ty.as_ref()),
                        );
                    }
                    self.env.define_type(
                        td.name.node.clone(),
                        Type::Struct {
                            name: td.name.node.clone(),
                            fields,
                        },
                    );
                }
                ItemKind::EnumDef(ed) => {
                    let variants: Vec<std::string::String> =
                        ed.variants.iter().map(|v| v.name.node.clone()).collect();
                    self.env.define_type(
                        ed.name.node.clone(),
                        Type::Enum {
                            name: ed.name.node.clone(),
                            variants,
                        },
                    );
                }
                ItemKind::FunctionDef(fd) => {
                    let params: Vec<Type> = fd
                        .params
                        .iter()
                        .map(|p| self.resolve_type_expr(p.ty.as_ref()))
                        .collect();
                    let ret = match &fd.return_ty {
                        Some(rt) => self.resolve_ast_type(&rt.node),
                        None => Type::Void,
                    };
                    self.env
                        .define_func(fd.name.node.clone(), FuncSig { params, ret });
                }
                ItemKind::ToolDecl(td) => {
                    let params: Vec<Type> = td
                        .params
                        .iter()
                        .map(|p| self.resolve_type_expr(p.ty.as_ref()))
                        .collect();
                    let ret = match &td.return_ty {
                        Some(rt) => self.resolve_ast_type(&rt.node),
                        None => Type::Any,
                    };
                    self.env
                        .define_func(td.name.node.clone(), FuncSig { params, ret });
                }
                ItemKind::WorkflowDecl(wd) => {
                    let params: Vec<Type> = wd
                        .params
                        .iter()
                        .map(|p| self.resolve_type_expr(p.ty.as_ref()))
                        .collect();
                    let ret = match &wd.return_ty {
                        Some(rt) => self.resolve_ast_type(&rt.node),
                        None => Type::Any,
                    };
                    self.env
                        .define_func(wd.name.node.clone(), FuncSig { params, ret });
                }
                ItemKind::ItemStatement(stmt) => {
                    match &stmt.node {
                        StmtKind::Assign(assign) => {
                            let rhs_type = self.infer_expr(&assign.value);
                            for target in &assign.targets {
                                if let AssignPath::Ident(ident) = &target.path {
                                    self.env.define_var(ident.node.clone(), rhs_type.clone());
                                }
                            }
                        }
                        StmtKind::Let(let_stmt) => {
                            let rhs_type = self.infer_expr(&let_stmt.value);
                            if let_stmt.is_const {
                                self.env
                                    .define_const(let_stmt.name.node.clone(), rhs_type);
                            } else {
                                self.env.define_var(let_stmt.name.node.clone(), rhs_type);
                            }
                        }
                        _ => {}
                    }
                }
                ItemKind::ProviderDecl(pd) => {
                    self.providers.insert(pd.name.node.clone());
                    self.check_provider_fields(pd);
                }
                ItemKind::AgentDecl(ad) => {
                    self.agents.insert(ad.name.node.clone());
                    self.check_agent_fields(ad);
                }
                _ => {}
            }
        }
    }

    // -----------------------------------------------------------------------
    // Agent/Provider field validation
    // -----------------------------------------------------------------------

    fn check_provider_fields(&mut self, provider: &ast::ProviderDecl) {
        for field in &provider.fields {
            if !is_valid_provider_field(&field.key.node) {
                self.add_warning(
                    format!("unknown provider field {:?}", field.key.node),
                    field.key.span,
                    "valid fields: api_key, model, endpoint, api_version, backend, host, account_id, temperature, max_tokens, input_token_cost, output_token_cost",
                );
            }
        }
    }

    fn check_agent_fields(&mut self, agent: &ast::AgentDecl) {
        let mut has_provider = false;
        for field in &agent.fields {
            if !is_valid_agent_field(&field.key.node) {
                self.add_warning(
                    format!("unknown agent field {:?}", field.key.node),
                    field.key.span,
                    "valid fields: provider, system, tools, handoffs, ui, mcp, temperature, max_tokens, max_steps, memory, output, timeout",
                );
            }
            if field.key.node == "provider" {
                has_provider = true;
                // Verify provider references a known provider declaration
                if let ExprKind::Ident(ref name) = field.value.node {
                    if !self.providers.contains(name.as_str()) {
                        self.add_error(
                            format!(
                                "agent {:?} references unknown provider {:?}",
                                agent.name.node, name
                            ),
                            field.value.span,
                        );
                    }
                }
            }
            if field.key.node == "tools" {
                // Verify tool references are known functions
                if let ExprKind::List(ref list) = field.value.node {
                    for elem in &list.elems {
                        if let ExprKind::Ident(ref name) = elem.node {
                            if self.env.lookup_func(name).is_none() {
                                self.add_error(
                                    format!(
                                        "agent {:?} references unknown tool {:?}",
                                        agent.name.node, name
                                    ),
                                    elem.span,
                                );
                            }
                        }
                    }
                }
            }
            if field.key.node == "ui" {
                // ui: ui (all built-in) or ui: [ui.Confirm, ui.Choices, ...]
                if let ExprKind::Ident(ref name) = field.value.node {
                    if name != "ui" {
                        self.add_error(
                            format!(
                                "agent {:?} ui field must be 'ui' (all components) or a list like [ui.Confirm, ...]",
                                agent.name.node
                            ),
                            field.value.span,
                        );
                    }
                } else if let ExprKind::List(ref list) = field.value.node {
                    for elem in &list.elems {
                        if let ExprKind::Field(ref fe) = elem.node {
                            if !is_valid_ui_component(&fe.field.node) {
                                self.add_warning(
                                    format!("unknown UI component {:?}", fe.field.node),
                                    fe.field.span,
                                    "built-in: StatusCard, Confirm, Choices, Table, CodeBlock, Diff, KeyValue, Progress, Chart",
                                );
                            }
                        }
                    }
                } else {
                    self.add_error(
                        format!(
                            "agent {:?} ui field must be 'ui' or a list like [ui.Confirm, ...]",
                            agent.name.node
                        ),
                        field.value.span,
                    );
                }
            }
            if field.key.node == "handoffs" {
                // Verify handoff references are known agents
                if let ExprKind::List(ref list) = field.value.node {
                    for elem in &list.elems {
                        if let ExprKind::Ident(ref name) = elem.node {
                            if !self.agents.contains(name.as_str()) {
                                self.add_error(
                                    format!(
                                        "agent {:?} handoff target {:?} is not declared",
                                        agent.name.node, name
                                    ),
                                    elem.span,
                                );
                            }
                        }
                    }
                }
            }
        }
        if !has_provider {
            self.add_error(
                format!(
                    "agent {:?} is missing required field 'provider'",
                    agent.name.node
                ),
                agent.name.span,
            );
        }
    }

    // -----------------------------------------------------------------------
    // Pass 2: Check function bodies
    // -----------------------------------------------------------------------

    fn check_bodies(&mut self, file: &SourceFile) {
        for item in &file.items {
            match &item.node {
                ItemKind::FunctionDef(fd) => self.check_function(fd),
                ItemKind::MethodDef(md) => self.check_method(md),
                ItemKind::ToolDecl(td) => {
                    if td.body.is_some() {
                        self.check_tool_body(td);
                    }
                }
                ItemKind::WorkflowDecl(wd) => self.check_workflow(wd),
                ItemKind::TestDecl(td) => self.check_test_body(td),
                _ => {}
            }
        }
    }

    fn check_function(&mut self, func: &ast::FunctionDef) {
        self.push_scope();

        for p in &func.params {
            let ty = self.resolve_type_expr(p.ty.as_ref());
            self.env.define_var(p.name.node.clone(), ty);
        }

        let saved_return = self.return_ty.take();
        self.return_ty = match &func.return_ty {
            Some(rt) => Some(self.resolve_ast_type(&rt.node)),
            None => None,
        };

        self.check_block(&func.body);

        self.pop_scope();
        self.return_ty = saved_return;
    }

    fn check_method(&mut self, md: &ast::MethodDef) {
        // Look up self type before pushing scope
        let self_ty = self.env.lookup_type(&md.type_name.node).cloned();

        self.push_scope();

        if let Some(ty) = self_ty {
            self.env.define_var("self", ty);
        }
        for p in &md.params {
            let ty = self.resolve_type_expr(p.ty.as_ref());
            self.env.define_var(p.name.node.clone(), ty);
        }

        let saved_in_method = self.in_method;
        let saved_return = self.return_ty.take();
        self.in_method = true;
        self.return_ty = match &md.return_ty {
            Some(rt) => Some(self.resolve_ast_type(&rt.node)),
            None => None,
        };

        self.check_block(&md.body);

        self.pop_scope();
        self.in_method = saved_in_method;
        self.return_ty = saved_return;
    }

    fn check_tool_body(&mut self, tool: &ast::ToolDecl) {
        let body = match &tool.body {
            Some(b) => b,
            None => return,
        };

        self.push_scope();

        for p in &tool.params {
            let ty = self.resolve_type_expr(p.ty.as_ref());
            self.env.define_var(p.name.node.clone(), ty);
        }

        let saved_return = self.return_ty.take();
        self.return_ty = match &tool.return_ty {
            Some(rt) => Some(self.resolve_ast_type(&rt.node)),
            None => None,
        };

        self.check_block(body);

        self.pop_scope();
        self.return_ty = saved_return;
    }

    fn check_test_body(&mut self, td: &ast::TestDecl) {
        self.push_scope();

        let saved_return = self.return_ty.take();
        let saved_in_test = self.in_test;
        self.return_ty = None;
        self.in_test = true;

        self.check_block(&td.body);

        self.pop_scope();
        self.return_ty = saved_return;
        self.in_test = saved_in_test;
    }

    fn check_workflow(&mut self, wf: &ast::WorkflowDecl) {
        self.push_scope();

        for p in &wf.params {
            let ty = self.resolve_type_expr(p.ty.as_ref());
            self.env.define_var(p.name.node.clone(), ty);
        }

        let saved_return = self.return_ty.take();
        self.return_ty = match &wf.return_ty {
            Some(rt) => Some(self.resolve_ast_type(&rt.node)),
            None => None,
        };

        self.check_block(&wf.body);

        // Check workflow-level lifecycle hooks
        for hook in &wf.hooks {
            self.push_scope();

            if hook.kind == ast::LifecycleHookKind::Onerror && !hook.err_name.is_empty() {
                self.env.define_var(hook.err_name.clone(), Type::String);
            }
            if hook.kind == ast::LifecycleHookKind::Onsuccess && !hook.arg_name.is_empty() {
                self.env.define_var(hook.arg_name.clone(), Type::Any);
            }

            self.check_block(&hook.body);

            self.pop_scope();
        }

        self.pop_scope();
        self.return_ty = saved_return;
    }

    // -----------------------------------------------------------------------
    // Statement checking
    // -----------------------------------------------------------------------

    fn check_block(&mut self, block: &Block) {
        for stmt in &block.statements {
            self.check_stmt(stmt);
        }
    }

    fn check_stmt(&mut self, stmt: &ast::Statement) {
        match &stmt.node {
            StmtKind::Let(s) => {
                let rhs_type = self.infer_expr(&s.value);
                if let Some(ref ann) = s.type_ann {
                    let annotated = self.resolve_ast_type(&ann.node);
                    if !type_equals(&annotated, &rhs_type)
                        && !is_any(&annotated)
                        && !is_any(&rhs_type)
                    {
                        self.add_error(
                            format!(
                                "type mismatch: expected {}, got {}",
                                annotated, rhs_type
                            ),
                            s.value.span,
                        );
                    }
                    if s.is_const {
                        self.env.define_const(s.name.node.clone(), annotated);
                    } else {
                        self.env.define_var(s.name.node.clone(), annotated);
                    }
                } else if s.is_const {
                    self.env
                        .define_const(s.name.node.clone(), rhs_type.clone());
                } else {
                    self.env.define_var(s.name.node.clone(), rhs_type.clone());
                }
                self.info.var_types.insert(s.name.span, rhs_type);
            }

            StmtKind::Assign(s) => {
                let rhs_type = self.infer_expr(&s.value);
                for target in &s.targets {
                    if let AssignPath::Ident(ref ident) = target.path {
                        if self.in_method && ident.node == "self" {
                            self.add_error("cannot reassign self in method", ident.span);
                        }
                        // Check const reassignment
                        if let Some(vi) = self.env.lookup_var_info(&ident.node) {
                            if vi.is_const {
                                self.add_error(
                                    format!("cannot reassign const '{}'", ident.node),
                                    ident.span,
                                );
                            }
                        }
                        if let Some(ref ty_ann) = target.ty {
                            let annotated = self.resolve_ast_type(&ty_ann.node);
                            if !type_equals(&annotated, &rhs_type)
                                && !is_any(&annotated)
                                && !is_any(&rhs_type)
                            {
                                self.add_error(
                                    format!(
                                        "type mismatch: expected {}, got {}",
                                        annotated, rhs_type
                                    ),
                                    s.value.span,
                                );
                            }
                            self.env.define_var(ident.node.clone(), annotated);
                        } else {
                            self.env.define_var(ident.node.clone(), rhs_type.clone());
                        }
                        self.info.var_types.insert(ident.span, rhs_type.clone());
                    }
                }
            }

            StmtKind::Expr(s) => {
                self.infer_expr(&s.value);
            }

            StmtKind::If(s) => {
                let cond_type = self.infer_expr(&s.condition);
                if !is_any(&cond_type) && !type_equals(&cond_type, &Type::Bool) {
                    self.add_warning(
                        format!("condition should be bool, got {}", cond_type),
                        s.condition.span,
                        "",
                    );
                }
                self.check_block(&s.then_branch);
                match &s.else_branch {
                    ElseBranch::None => {}
                    ElseBranch::Block(block) => {
                        self.check_block(block);
                    }
                    ElseBranch::ElseIf(else_if) => {
                        // Construct a statement wrapping the inner IfStmt
                        let inner_stmt = Spanned {
                            node: StmtKind::If(else_if.node.clone()),
                            span: else_if.span,
                        };
                        self.check_stmt(&inner_stmt);
                    }
                }
            }

            StmtKind::For(s) => {
                self.infer_expr(&s.iterator);

                self.push_scope();

                match &s.pattern {
                    ForPattern::Single(name) => {
                        self.env.define_var(name.node.clone(), Type::Any);
                    }
                    ForPattern::Pair(first, second) => {
                        self.env.define_var(first.node.clone(), Type::Any);
                        self.env.define_var(second.node.clone(), Type::Any);
                    }
                }

                self.check_block(&s.body);

                self.pop_scope();
            }

            StmtKind::While(s) => {
                self.infer_expr(&s.condition);
                self.check_block(&s.body);
            }

            StmtKind::Return(s) => {
                if !s.values.is_empty() {
                    let ret_type = self.infer_expr(&s.values[0]);
                    for v in &s.values[1..] {
                        self.infer_expr(v);
                    }
                    // Check against declared return type
                    if let Some(ref expected) = self.return_ty {
                        if !is_any(expected) && !is_any(&ret_type) && !type_equals(expected, &ret_type)
                        {
                            self.add_warning(
                                format!(
                                    "return type mismatch: expected {}, got {}",
                                    expected, ret_type
                                ),
                                s.values[0].span,
                                "",
                            );
                        }
                    }
                } else {
                    // Empty return in a function with declared return type
                    if let Some(ref expected) = self.return_ty {
                        if !is_any(expected) && !type_equals(expected, &Type::Void) {
                            self.add_warning(
                                format!("empty return in function expecting {}", expected),
                                stmt.span,
                                "",
                            );
                        }
                    }
                }
            }

            StmtKind::Try(s) => {
                let saved_try = self.in_try;
                self.in_try = true;
                self.check_block(&s.body);
                self.in_try = saved_try;

                self.push_scope();
                self.env
                    .define_var(s.error_name.node.clone(), Type::String);
                self.check_block(&s.catch_body);
                self.pop_scope();
            }

            StmtKind::Defer(s) => {
                self.infer_expr(&s.value);
            }

            StmtKind::Match(s) => {
                let subject_type = self.infer_expr(&s.subject);
                for arm in &s.arms {
                    if let Some(ref guard) = arm.guard {
                        self.infer_expr(guard);
                    }
                    match &arm.body {
                        MatchArmBody::Expr(expr) => {
                            self.infer_expr(&expr);
                        }
                        MatchArmBody::Block(block) => {
                            self.check_block(block);
                        }
                    }
                }
                // Exhaustiveness check for simple enums
                if matches!(subject_type, Type::Enum { .. }) {
                    let subj_span = s.subject.span;
                    self.check_enum_exhaustiveness(&subject_type, &s.arms, subj_span);
                }
            }

            StmtKind::Step(s) => {
                // Step body shares the enclosing workflow scope
                for inner_stmt in &s.body {
                    self.check_stmt(inner_stmt);
                }
                // Check lifecycle hooks
                for hook in &s.hooks {
                    self.push_scope();

                    if hook.kind == ast::LifecycleHookKind::Onerror && !hook.err_name.is_empty() {
                        self.env.define_var(hook.err_name.clone(), Type::String);
                    }

                    self.check_block(&hook.body);

                    self.pop_scope();
                }
            }

            StmtKind::ErrDefer(s) => {
                self.infer_expr(&s.value);
            }

            StmtKind::Assert(s) => {
                if !self.in_test {
                    self.add_error("assert can only be used inside test blocks", stmt.span);
                }
                let cond_type = self.infer_expr(&s.condition);
                if !is_any(&cond_type) && !type_equals(&cond_type, &Type::Bool) {
                    self.add_warning(
                        format!("assert condition should be bool, got {}", cond_type),
                        s.condition.span,
                        "",
                    );
                }
                if let Some(ref msg) = s.message {
                    self.infer_expr(msg);
                }
            }

            StmtKind::Break | StmtKind::Continue => {}
        }
    }

    // -----------------------------------------------------------------------
    // Expression type inference
    // -----------------------------------------------------------------------

    fn infer_expr(&mut self, expr: &ast::Expr) -> Type {
        let ty = match &expr.node {
            ExprKind::Literal(lit) => self.infer_literal(lit),

            ExprKind::Ident(name) => {
                if let Some(t) = self.env.lookup_var(name) {
                    t.clone()
                } else if self.env.lookup_func(name).is_some() {
                    Type::Any // function reference (e.g. passed as callback)
                } else if self.env.lookup_type(name).is_some() {
                    Type::Any // type reference (e.g. enum variant prefix)
                } else if self.agents.contains(name.as_str())
                    || self.providers.contains(name.as_str())
                {
                    Type::Any // agent or provider reference
                } else if is_stdlib_module(name) {
                    Type::Any // stdlib module qualifier
                } else if name == "nil" || name == "true" || name == "false" {
                    Type::Any // language builtins
                } else {
                    self.add_warning(
                        format!("undefined variable '{}'", name),
                        expr.span,
                        "",
                    );
                    Type::Any
                }
            }

            ExprKind::Binary(bin) => {
                let left = self.infer_expr(&bin.left);
                let right = self.infer_expr(&bin.right);
                self.infer_binary_op(&bin.op.node, &left, &right, expr.span)
            }

            ExprKind::Unary(u) => {
                let operand = self.infer_expr(&u.operand);
                match u.op.node {
                    ast::UnaryOp::Neg => {
                        if type_equals(&operand, &Type::Int)
                            || type_equals(&operand, &Type::Float)
                            || is_any(&operand)
                        {
                            operand
                        } else {
                            self.add_error(format!("cannot negate {}", operand), expr.span);
                            Type::Any
                        }
                    }
                    ast::UnaryOp::Not => Type::Bool,
                    _ => operand,
                }
            }

            ExprKind::Call(call) => self.infer_call(call, expr.span),

            ExprKind::MethodCall(mc) => {
                self.infer_expr(&mc.receiver);
                for arg in &mc.args {
                    self.infer_expr(&arg.value);
                }
                Type::Any
            }

            ExprKind::Field(fe) => {
                let obj_type = self.infer_expr(&fe.object);
                if let Type::Struct {
                    ref name,
                    ref fields,
                } = obj_type
                {
                    if let Some(ft) = fields.get(&fe.field.node) {
                        ft.clone()
                    } else {
                        self.add_error(
                            format!("unknown field '{}' on type {}", fe.field.node, name),
                            fe.field.span,
                        );
                        Type::Any
                    }
                } else {
                    Type::Any
                }
            }

            ExprKind::Index(ie) => {
                let obj_type = self.infer_expr(&ie.object);
                self.infer_expr(&ie.index);
                match obj_type {
                    Type::List { ref elem } => (**elem).clone(),
                    Type::Map { ref value, .. } => (**value).clone(),
                    _ => Type::Any,
                }
            }

            ExprKind::List(list) => {
                if list.elems.is_empty() {
                    Type::List {
                        elem: Box::new(Type::Any),
                    }
                } else {
                    let elem_type = self.infer_expr(&list.elems[0]);
                    let mut all_same = true;
                    for el in &list.elems[1..] {
                        let et = self.infer_expr(el);
                        if !type_equals(&elem_type, &et) {
                            all_same = false;
                        }
                    }
                    if all_same && !is_any(&elem_type) {
                        Type::List {
                            elem: Box::new(elem_type),
                        }
                    } else {
                        Type::List {
                            elem: Box::new(Type::Any),
                        }
                    }
                }
            }

            ExprKind::Map(map) => {
                if map.entries.is_empty() {
                    Type::Map {
                        key: Box::new(Type::String),
                        value: Box::new(Type::Any),
                    }
                } else {
                    let val_type = self.infer_expr(&map.entries[0].value);
                    self.infer_map_key(&map.entries[0].key);
                    let mut all_same = true;
                    for entry in &map.entries[1..] {
                        self.infer_map_key(&entry.key);
                        let vt = self.infer_expr(&entry.value);
                        if !type_equals(&val_type, &vt) {
                            all_same = false;
                        }
                    }
                    if all_same && !is_any(&val_type) {
                        Type::Map {
                            key: Box::new(Type::String),
                            value: Box::new(val_type),
                        }
                    } else {
                        Type::Map {
                            key: Box::new(Type::String),
                            value: Box::new(Type::Any),
                        }
                    }
                }
            }

            ExprKind::Instance(ie) => {
                let ty = if let Some(t) = self.env.lookup_type(&ie.type_name.node) {
                    t.clone()
                } else {
                    Type::Any
                };
                for f in &ie.fields {
                    self.infer_expr(&f.value);
                }
                ty
            }

            ExprKind::Pipe(pe) => {
                self.infer_expr(&pe.left);
                self.infer_expr(&pe.right)
            }

            ExprKind::Lambda(_) => Type::Any,

            ExprKind::Paren(pe) => self.infer_expr(&pe.inner),

            ExprKind::None => Type::Any,

            ExprKind::Some(se) => self.infer_expr(&se.inner),

            ExprKind::If(ie) => {
                self.infer_expr(&ie.if_stmt.condition);
                self.check_block(&ie.if_stmt.then_branch);
                Type::Any
            }

            ExprKind::Match(me) => {
                self.infer_expr(&me.subject);
                for arm in &me.arms {
                    match &arm.body {
                        MatchArmBody::Expr(expr) => {
                            self.infer_expr(expr);
                        }
                        MatchArmBody::Block(block) => {
                            self.check_block(block);
                        }
                    }
                }
                Type::Any
            }

            ExprKind::Range(re) => {
                self.infer_expr(&re.start);
                self.infer_expr(&re.end);
                Type::List {
                    elem: Box::new(Type::Int),
                }
            }

            ExprKind::Propagate(pe) => {
                let ty = self.infer_expr(&pe.inner);
                if !self.in_try {
                    self.add_warning(
                        "'?' operator used outside a try block -- will panic on error",
                        expr.span,
                        "wrap in try { ... } catch err { ... } to handle the error, or use 'result, err = call()' pattern",
                    );
                }
                ty
            }

            ExprKind::Orelse(oe) => {
                self.infer_expr(&oe.left);
                self.infer_expr(&oe.default)
            }

            ExprKind::Spawn(se) => {
                self.check_block(&se.body);
                Type::List {
                    elem: Box::new(Type::Any),
                }
            }

            ExprKind::Async(ae) => {
                self.check_block(&ae.body);
                Type::Any
            }

            ExprKind::Block(be) => {
                self.check_block(&be.body);
                Type::Any
            }

            ExprKind::Select(_) => Type::Any,
        };

        self.info.expr_types.insert(expr.span, ty.clone());
        ty
    }

    fn infer_literal(&self, lit: &Literal) -> Type {
        match lit {
            Literal::Int(_) => Type::Int,
            Literal::Float(_) => Type::Float,
            Literal::String(_) => Type::String,
            Literal::Bool(_) => Type::Bool,
            Literal::InterpolatedString(_) => Type::String,
        }
    }

    fn infer_binary_op(
        &self,
        op: &ast::BinaryOp,
        left: &Type,
        right: &Type,
        _span: Span,
    ) -> Type {
        use ast::BinaryOp::*;
        match op {
            Add | Sub | Mul | Div | Mod => {
                if is_any(left) || is_any(right) {
                    return Type::Any;
                }
                if type_equals(left, &Type::String) && matches!(op, Add) {
                    return Type::String;
                }
                if type_equals(left, right) {
                    return left.clone();
                }
                // int + float -> float
                if (type_equals(left, &Type::Int) && type_equals(right, &Type::Float))
                    || (type_equals(left, &Type::Float) && type_equals(right, &Type::Int))
                {
                    return Type::Float;
                }
                Type::Any
            }
            Eq | Ne | Lt | Gt | Le | Ge => Type::Bool,
            And | Or => Type::Bool,
            BitAnd | BitOr | BitXor | Shl | Shr => {
                // Bitwise ops -- return Int if both sides are int, else Any
                if type_equals(left, &Type::Int) && type_equals(right, &Type::Int) {
                    Type::Int
                } else {
                    Type::Any
                }
            }
        }
    }

    fn infer_call(&mut self, call: &ast::CallExpr, span: Span) -> Type {
        // Check argument types
        for arg in &call.args {
            self.infer_expr(&arg.value);
        }

        // Qualified call: module.function
        if let ExprKind::Field(ref field) = call.callee.node {
            if let ExprKind::Ident(ref module_name) = field.object.node {
                let key = format!("{}.{}", module_name, field.field.node);
                if let Some(sig) = self.env.lookup_func(&key) {
                    return sig.ret.clone();
                }
                // Don't warn for agent/provider method calls or stdlib modules
                if !self.agents.contains(module_name.as_str())
                    && !self.providers.contains(module_name.as_str())
                    && !is_stdlib_module(module_name)
                {
                    let is_var = self.env.lookup_var(module_name).is_some();
                    let is_type = self.env.lookup_type(module_name).is_some();
                    if !is_var && !is_type {
                        self.add_warning(
                            format!("undefined function '{}'", key),
                            span,
                            "",
                        );
                    }
                }
            }
            return Type::Any;
        }

        // Bare function call
        if let ExprKind::Ident(ref name) = call.callee.node {
            if let Some(sig) = self.env.lookup_func(name) {
                let ret = sig.ret.clone();
                // env("KEY", float) -> FloatType, env("KEY", int) -> IntType
                if name == "env" && call.args.len() >= 2 {
                    if let ExprKind::Ident(ref hint) = call.args[1].value.node {
                        match hint.as_str() {
                            "float" => return Type::Float,
                            "int" => return Type::Int,
                            _ => {}
                        }
                    }
                }
                return ret;
            }
            // Don't warn for known types used as constructors
            if self.env.lookup_type(name).is_none() {
                self.add_warning(format!("undefined function '{}'", name), span, "");
            }
            return Type::Any;
        }

        Type::Any
    }

    // -----------------------------------------------------------------------
    // Type resolution helpers
    // -----------------------------------------------------------------------

    fn resolve_type_expr(&self, ty: Option<&Spanned<ast::Type>>) -> Type {
        match ty {
            None => Type::Any,
            Some(spanned) => self.resolve_ast_type(&spanned.node),
        }
    }

    fn resolve_ast_type(&self, ty: &ast::Type) -> Type {
        match ty {
            ast::Type::Named(name) => match name.as_str() {
                "int" => Type::Int,
                "float" => Type::Float,
                "string" => Type::String,
                "bool" => Type::Bool,
                "any" => Type::Any,
                "void" => Type::Void,
                "file" => Type::String, // file is semantically a string (temp path)
                _ => {
                    if let Some(user_ty) = self.env.lookup_type(name) {
                        user_ty.clone()
                    } else {
                        Type::Any
                    }
                }
            },
            ast::Type::List(elem) => Type::List {
                elem: Box::new(self.resolve_ast_type(&elem.node)),
            },
            ast::Type::Map { key, value } => Type::Map {
                key: Box::new(self.resolve_ast_type(&key.node)),
                value: Box::new(self.resolve_ast_type(&value.node)),
            },
            ast::Type::Function { params, ret } => {
                let p: Vec<Type> = params.iter().map(|p| self.resolve_ast_type(&p.node)).collect();
                Type::Func {
                    params: p,
                    ret: Box::new(self.resolve_ast_type(&ret.node)),
                }
            }
            ast::Type::Option(inner) => {
                // Option types resolve to the inner type for simplicity in the checker
                self.resolve_ast_type(&inner.node)
            }
            ast::Type::Union(_) => Type::Any,
            ast::Type::Generic { .. } => Type::Any,
        }
    }

    /// Infers the type of a map key without triggering undefined variable
    /// warnings for bare identifiers (which are implicitly string keys in Haira).
    fn infer_map_key(&mut self, key: &ast::Expr) -> Type {
        if let ExprKind::Ident(_) = &key.node {
            // Bare identifier used as map key -- treat as string, no warning
            self.info.expr_types.insert(key.span, Type::String);
            return Type::String;
        }
        self.infer_expr(key)
    }
}

// ===========================================================================
// Enum exhaustiveness helper (free function)
// ===========================================================================

fn collect_covered_variants(
    p: &ast::Pattern,
    enum_name: &str,
    covered: &mut HashSet<std::string::String>,
    has_wildcard: &mut bool,
) {
    match p {
        ast::Pattern::Wildcard => {
            *has_wildcard = true;
        }
        ast::Pattern::Ident(name) => {
            covered.insert(name.clone());
        }
        ast::Pattern::Or(patterns) => {
            for sub in patterns {
                collect_covered_variants(&sub.node, enum_name, covered, has_wildcard);
            }
        }
        _ => {}
    }
}

// ===========================================================================
// Tests
// ===========================================================================

#[cfg(test)]
mod tests {
    use super::{check, type_equals, Env, FuncSig, Type};
    use haira_ast::{
        AgentDecl, AgentField, AssignPath, AssignStmt, AssignTarget, BinaryExpr, BinaryOp, Block,
        Expr, ExprKind, ExprStmt, FunctionDef, Item, ItemKind, LetStmt, Literal, MethodDef,
        ProviderDecl, ProviderField, ReturnStmt, SourceFile, Span, Spanned, Statement, StmtKind,
        TypeDef,
    };
    use haira_errors::Level;
    use std::collections::HashMap;

    // -- Type display --------------------------------------------------------

    #[test]
    fn type_display() {
        assert_eq!(Type::Int.to_string(), "int");
        assert_eq!(Type::Float.to_string(), "float");
        assert_eq!(Type::String.to_string(), "string");
        assert_eq!(Type::Bool.to_string(), "bool");
        assert_eq!(Type::Any.to_string(), "any");
        assert_eq!(Type::Void.to_string(), "void");
        assert_eq!(Type::Error.to_string(), "error");
        assert_eq!(
            Type::List {
                elem: Box::new(Type::Int)
            }
            .to_string(),
            "[int]"
        );
        assert_eq!(
            Type::Map {
                key: Box::new(Type::String),
                value: Box::new(Type::Int)
            }
            .to_string(),
            "{string: int}"
        );
        assert_eq!(
            Type::Struct {
                name: "User".into(),
                fields: HashMap::new()
            }
            .to_string(),
            "User"
        );
        assert_eq!(
            Type::Enum {
                name: "Color".into(),
                variants: vec![]
            }
            .to_string(),
            "Color"
        );
        assert_eq!(
            Type::Func {
                params: vec![],
                ret: Box::new(Type::Void)
            }
            .to_string(),
            "fn"
        );
    }

    // -- Type equality -------------------------------------------------------

    #[test]
    fn type_equality() {
        assert!(type_equals(&Type::Int, &Type::Int));
        assert!(type_equals(&Type::Float, &Type::Float));
        assert!(type_equals(&Type::String, &Type::String));
        assert!(type_equals(&Type::Bool, &Type::Bool));
        assert!(type_equals(&Type::Void, &Type::Void));
        assert!(type_equals(&Type::Error, &Type::Error));

        // Any matches everything
        assert!(type_equals(&Type::Any, &Type::Int));
        assert!(type_equals(&Type::Int, &Type::Any));
        assert!(type_equals(&Type::Any, &Type::Any));

        // Different primitives don't match
        assert!(!type_equals(&Type::Int, &Type::Float));
        assert!(!type_equals(&Type::String, &Type::Bool));

        // List equality
        assert!(type_equals(
            &Type::List {
                elem: Box::new(Type::Int)
            },
            &Type::List {
                elem: Box::new(Type::Int)
            }
        ));
        assert!(!type_equals(
            &Type::List {
                elem: Box::new(Type::Int)
            },
            &Type::List {
                elem: Box::new(Type::String)
            }
        ));

        // Map equality
        assert!(type_equals(
            &Type::Map {
                key: Box::new(Type::String),
                value: Box::new(Type::Int)
            },
            &Type::Map {
                key: Box::new(Type::String),
                value: Box::new(Type::Int)
            }
        ));

        // Struct equality (by name)
        assert!(type_equals(
            &Type::Struct {
                name: "User".into(),
                fields: HashMap::new()
            },
            &Type::Struct {
                name: "User".into(),
                fields: HashMap::new()
            }
        ));
        assert!(!type_equals(
            &Type::Struct {
                name: "User".into(),
                fields: HashMap::new()
            },
            &Type::Struct {
                name: "Post".into(),
                fields: HashMap::new()
            }
        ));

        // Enum equality (by name)
        assert!(type_equals(
            &Type::Enum {
                name: "Color".into(),
                variants: vec![]
            },
            &Type::Enum {
                name: "Color".into(),
                variants: vec![]
            }
        ));

        // Func equality
        assert!(type_equals(
            &Type::Func {
                params: vec![Type::Int, Type::String],
                ret: Box::new(Type::Bool)
            },
            &Type::Func {
                params: vec![Type::Int, Type::String],
                ret: Box::new(Type::Bool)
            }
        ));
        assert!(!type_equals(
            &Type::Func {
                params: vec![Type::Int],
                ret: Box::new(Type::Void)
            },
            &Type::Func {
                params: vec![Type::Int, Type::String],
                ret: Box::new(Type::Void)
            }
        ));
    }

    // -- Env scoping ---------------------------------------------------------

    #[test]
    fn env_scoping_parent_lookup() {
        let mut root = Env::new();
        root.define_var("x", Type::Int);
        let child = root.child();
        // Child can see parent's variable
        assert!(child.lookup_var("x").is_some());
        assert!(type_equals(child.lookup_var("x").unwrap(), &Type::Int));

        // Child's own variable shadows parent
        let mut child = child;
        child.define_var("x", Type::String);
        assert!(type_equals(child.lookup_var("x").unwrap(), &Type::String));

        // Parent is unaffected (through the child's parent pointer)
        let parent = child.parent.as_ref().unwrap();
        assert!(type_equals(parent.lookup_var("x").unwrap(), &Type::Int));
    }

    #[test]
    fn env_type_and_func_lookup() {
        let mut env = Env::new();
        env.define_type(
            "User",
            Type::Struct {
                name: "User".into(),
                fields: HashMap::new(),
            },
        );
        env.define_func(
            "greet",
            FuncSig {
                params: vec![Type::String],
                ret: Type::Void,
            },
        );

        assert!(env.lookup_type("User").is_some());
        assert!(env.lookup_func("greet").is_some());
        assert!(env.lookup_type("Unknown").is_none());
        assert!(env.lookup_func("unknown").is_none());
    }

    // -- Helper to build a minimal source file for checker tests -----------

    fn span(start: usize, end: usize) -> Span {
        Span::new(start, end)
    }

    fn spanned<T>(node: T, s: Span) -> Spanned<T> {
        Spanned::new(node, s)
    }

    fn make_source(items: Vec<Item>) -> SourceFile {
        SourceFile {
            items,
            span: span(0, 100),
        }
    }

    fn make_int_literal(s: Span) -> Expr {
        spanned(ExprKind::Literal(Literal::Int(42)), s)
    }

    fn make_string_literal(val: &str, s: Span) -> Expr {
        spanned(ExprKind::Literal(Literal::String(val.into())), s)
    }

    fn make_ident_expr(name: &str, s: Span) -> Expr {
        spanned(ExprKind::Ident(name.into()), s)
    }

    fn make_block(stmts: Vec<Statement>) -> Block {
        Block {
            statements: stmts,
            span: span(0, 100),
        }
    }

    // -- Const reassignment detection ----------------------------------------

    #[test]
    fn const_reassignment_detection() {
        // const x = 42
        // x = 10   <-- should produce error
        let let_stmt = spanned(
            StmtKind::Let(LetStmt {
                name: spanned("x".into(), span(6, 7)),
                type_ann: None,
                value: Box::new(make_int_literal(span(10, 12))),
                is_const: true,
            }),
            span(0, 12),
        );

        let assign_stmt = spanned(
            StmtKind::Assign(AssignStmt {
                targets: vec![AssignTarget {
                    path: AssignPath::Ident(spanned("x".into(), span(13, 14))),
                    ty: None,
                }],
                value: Box::new(make_int_literal(span(17, 19))),
            }),
            span(13, 19),
        );

        let func = FunctionDef {
            is_public: false,
            name: spanned("test_fn".into(), span(0, 7)),
            params: vec![],
            return_ty: None,
            body: make_block(vec![let_stmt, assign_stmt]),
        };

        let src = make_source(vec![spanned(ItemKind::FunctionDef(func), span(0, 100))]);
        let (_, diags) = check(&src);

        let has_const_error = diags
            .iter()
            .any(|d| d.level == Level::Error && d.message.contains("cannot reassign const"));
        assert!(
            has_const_error,
            "expected const reassignment error, got: {:?}",
            diags
        );
    }

    // -- Self reassignment in methods ----------------------------------------

    #[test]
    fn self_reassignment_in_method() {
        // Type.method() { self = 42 }
        let assign_stmt = spanned(
            StmtKind::Assign(AssignStmt {
                targets: vec![AssignTarget {
                    path: AssignPath::Ident(spanned("self".into(), span(20, 24))),
                    ty: None,
                }],
                value: Box::new(make_int_literal(span(27, 29))),
            }),
            span(20, 29),
        );

        // First register a type so self can resolve
        let type_def = spanned(
            ItemKind::TypeDef(TypeDef {
                is_public: false,
                name: spanned("MyType".into(), span(0, 6)),
                fields: vec![],
            }),
            span(0, 10),
        );

        let method = spanned(
            ItemKind::MethodDef(MethodDef {
                type_name: spanned("MyType".into(), span(10, 16)),
                name: spanned("do_stuff".into(), span(17, 25)),
                params: vec![],
                return_ty: None,
                body: make_block(vec![assign_stmt]),
            }),
            span(10, 50),
        );

        let src = make_source(vec![type_def, method]);
        let (_, diags) = check(&src);

        let has_self_error = diags
            .iter()
            .any(|d| d.level == Level::Error && d.message.contains("cannot reassign self"));
        assert!(
            has_self_error,
            "expected self reassignment error, got: {:?}",
            diags
        );
    }

    // -- Undefined variable warning ------------------------------------------

    #[test]
    fn undefined_variable_warning() {
        // fn test_fn() { x }  where x is not defined
        let expr_stmt = spanned(
            StmtKind::Expr(ExprStmt {
                value: Box::new(make_ident_expr("undefined_var", span(15, 28))),
            }),
            span(15, 28),
        );

        let func = FunctionDef {
            is_public: false,
            name: spanned("test_fn".into(), span(0, 7)),
            params: vec![],
            return_ty: None,
            body: make_block(vec![expr_stmt]),
        };

        let src = make_source(vec![spanned(ItemKind::FunctionDef(func), span(0, 100))]);
        let (_, diags) = check(&src);

        let has_undef = diags
            .iter()
            .any(|d| d.level == Level::Warning && d.message.contains("undefined variable"));
        assert!(
            has_undef,
            "expected undefined variable warning, got: {:?}",
            diags
        );
    }

    // -- Return type mismatch ------------------------------------------------

    #[test]
    fn return_type_mismatch() {
        // fn test_fn() -> int { return "hello" }
        let ret_stmt = spanned(
            StmtKind::Return(ReturnStmt {
                values: vec![make_string_literal("hello", span(30, 37))],
            }),
            span(24, 37),
        );

        let func = FunctionDef {
            is_public: false,
            name: spanned("test_fn".into(), span(0, 7)),
            params: vec![],
            return_ty: Some(spanned(
                haira_ast::Type::Named("int".into()),
                span(15, 18),
            )),
            body: make_block(vec![ret_stmt]),
        };

        let src = make_source(vec![spanned(ItemKind::FunctionDef(func), span(0, 100))]);
        let (_, diags) = check(&src);

        let has_mismatch = diags
            .iter()
            .any(|d| d.message.contains("return type mismatch"));
        assert!(
            has_mismatch,
            "expected return type mismatch, got: {:?}",
            diags
        );
    }

    // -- Provider field validation -------------------------------------------

    #[test]
    fn provider_field_validation() {
        let provider = spanned(
            ItemKind::ProviderDecl(ProviderDecl {
                name: spanned("MyProvider".into(), span(0, 10)),
                fields: vec![
                    ProviderField {
                        key: spanned("model".into(), span(12, 17)),
                        value: make_string_literal("gpt-4", span(19, 24)),
                    },
                    ProviderField {
                        key: spanned("bogus_field".into(), span(26, 37)),
                        value: make_string_literal("value", span(39, 44)),
                    },
                ],
            }),
            span(0, 50),
        );

        let src = make_source(vec![provider]);
        let (_, diags) = check(&src);

        let has_unknown = diags
            .iter()
            .any(|d| d.level == Level::Warning && d.message.contains("unknown provider field"));
        assert!(
            has_unknown,
            "expected unknown provider field warning, got: {:?}",
            diags
        );
    }

    // -- Agent field validation (missing provider) ---------------------------

    #[test]
    fn agent_missing_provider() {
        let agent = spanned(
            ItemKind::AgentDecl(AgentDecl {
                name: spanned("MyAgent".into(), span(0, 7)),
                fields: vec![AgentField {
                    key: spanned("system".into(), span(10, 16)),
                    value: make_string_literal("You are helpful.", span(18, 34)),
                }],
            }),
            span(0, 40),
        );

        let src = make_source(vec![agent]);
        let (_, diags) = check(&src);

        let has_missing = diags
            .iter()
            .any(|d| d.level == Level::Error && d.message.contains("missing required field 'provider'"));
        assert!(
            has_missing,
            "expected missing provider error, got: {:?}",
            diags
        );
    }

    // -- Simple expression type inference ------------------------------------

    #[test]
    fn simple_expression_inference() {
        // fn test_fn() { let a = 1 + 2 }
        let add_expr = spanned(
            ExprKind::Binary(BinaryExpr {
                left: Box::new(spanned(ExprKind::Literal(Literal::Int(1)), span(20, 21))),
                op: spanned(BinaryOp::Add, span(22, 23)),
                right: Box::new(spanned(ExprKind::Literal(Literal::Int(2)), span(24, 25))),
            }),
            span(20, 25),
        );

        let let_stmt = spanned(
            StmtKind::Let(LetStmt {
                name: spanned("a".into(), span(14, 15)),
                type_ann: None,
                value: Box::new(add_expr),
                is_const: false,
            }),
            span(10, 25),
        );

        let func = FunctionDef {
            is_public: false,
            name: spanned("test_fn".into(), span(0, 7)),
            params: vec![],
            return_ty: None,
            body: make_block(vec![let_stmt]),
        };

        let src = make_source(vec![spanned(ItemKind::FunctionDef(func), span(0, 100))]);
        let (info, diags) = check(&src);

        // int + int = int
        let expr_ty = info.expr_types.get(&span(20, 25));
        assert!(expr_ty.is_some(), "expression type should be recorded");
        assert!(
            type_equals(expr_ty.unwrap(), &Type::Int),
            "int + int should be int, got: {:?}",
            expr_ty
        );

        // No errors expected
        let errors: Vec<_> = diags.iter().filter(|d| d.level == Level::Error).collect();
        assert!(errors.is_empty(), "unexpected errors: {:?}", errors);
    }

    #[test]
    fn string_concat_inference() {
        // "a" + "b" should be string
        let add_expr = spanned(
            ExprKind::Binary(BinaryExpr {
                left: Box::new(make_string_literal("a", span(0, 3))),
                op: spanned(BinaryOp::Add, span(4, 5)),
                right: Box::new(make_string_literal("b", span(6, 9))),
            }),
            span(0, 9),
        );

        let let_stmt = spanned(
            StmtKind::Let(LetStmt {
                name: spanned("s".into(), span(14, 15)),
                type_ann: None,
                value: Box::new(add_expr),
                is_const: false,
            }),
            span(10, 20),
        );

        let func = FunctionDef {
            is_public: false,
            name: spanned("test_fn".into(), span(20, 27)),
            params: vec![],
            return_ty: None,
            body: make_block(vec![let_stmt]),
        };

        let src = make_source(vec![spanned(ItemKind::FunctionDef(func), span(0, 100))]);
        let (info, _) = check(&src);

        let expr_ty = info.expr_types.get(&span(0, 9));
        assert!(expr_ty.is_some());
        assert!(
            type_equals(expr_ty.unwrap(), &Type::String),
            "string + string should be string, got: {:?}",
            expr_ty
        );
    }

    #[test]
    fn int_float_promotion() {
        // 1 + 2.0 should be float
        let add_expr = spanned(
            ExprKind::Binary(BinaryExpr {
                left: Box::new(spanned(ExprKind::Literal(Literal::Int(1)), span(0, 1))),
                op: spanned(BinaryOp::Add, span(2, 3)),
                right: Box::new(spanned(ExprKind::Literal(Literal::Float(2.0)), span(4, 7))),
            }),
            span(0, 7),
        );

        let let_stmt = spanned(
            StmtKind::Let(LetStmt {
                name: spanned("x".into(), span(10, 11)),
                type_ann: None,
                value: Box::new(add_expr),
                is_const: false,
            }),
            span(8, 12),
        );

        let func = FunctionDef {
            is_public: false,
            name: spanned("test_fn".into(), span(20, 27)),
            params: vec![],
            return_ty: None,
            body: make_block(vec![let_stmt]),
        };

        let src = make_source(vec![spanned(ItemKind::FunctionDef(func), span(0, 100))]);
        let (info, _) = check(&src);

        let expr_ty = info.expr_types.get(&span(0, 7));
        assert!(expr_ty.is_some());
        assert!(
            type_equals(expr_ty.unwrap(), &Type::Float),
            "int + float should be float, got: {:?}",
            expr_ty
        );
    }
}
