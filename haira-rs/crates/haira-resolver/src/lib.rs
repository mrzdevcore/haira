//! Multi-file module resolution for the Haira programming language.
//!
//! This is a 1:1 port of `compiler/internal/resolver/resolver.go`.
//!
//! Import rules:
//!   - Stdlib imports ("io", "http", "json", etc.) are handled by codegen -- no file resolution needed.
//!   - Project-local imports ("models/user") resolve to `<project_root>/models/user.haira`.
//!   - The project root is the directory containing the main file.
//!   - Imported symbols are available as `module.Symbol` (e.g., `user.User`, `auth.verify`).
//!   - Circular imports are detected and reported as errors.

use std::collections::{HashMap, HashSet};
use std::fs;
use std::path::PathBuf;

use haira_ast::{Item, ItemKind, SourceFile, Span};
use haira_errors::{Diagnostic, Level};

// ---------------------------------------------------------------------------
// Public types
// ---------------------------------------------------------------------------

/// A resolved multi-file Haira program.
#[derive(Debug, Clone)]
pub struct Program {
    /// The main source file's AST.
    pub main: SourceFile,
    /// Import path -> parsed module.
    pub modules: HashMap<String, Module>,
}

/// A single imported file.
#[derive(Debug, Clone)]
pub struct Module {
    /// The import path (e.g., "models/user").
    pub path: String,
    /// The absolute file path on disk.
    pub file_path: String,
    /// The parsed AST.
    pub file: SourceFile,
}

// ---------------------------------------------------------------------------
// Stdlib modules
// ---------------------------------------------------------------------------

/// Known standard library module names that are handled by codegen and don't
/// need file resolution.
fn is_stdlib_module(name: &str) -> bool {
    matches!(
        name,
        "io" | "http"
            | "mcp"
            | "env"
            | "json"
            | "postgres"
            | "slack"
            | "excel"
            | "time"
            | "string"
            | "regex"
            | "math"
            | "conv"
            | "array"
            | "map"
            | "log"
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

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

/// Parses the main file and all its transitive project-local imports.
///
/// Returns a [`Program`] with the main file and all resolved modules, plus any
/// diagnostics encountered during resolution.
pub fn resolve(main_file: &str) -> (Program, Vec<Diagnostic>) {
    let abs_main = match fs::canonicalize(main_file) {
        Ok(p) => p,
        Err(e) => {
            // If canonicalize fails (e.g. file doesn't exist), try to form an absolute path
            // manually so we can still report a useful error.
            let path = PathBuf::from(main_file);
            let abs = if path.is_absolute() {
                path
            } else {
                std::env::current_dir().unwrap_or_default().join(&path)
            };

            return (
                Program {
                    main: SourceFile {
                        items: Vec::new(),
                        span: Span::default(),
                    },
                    modules: HashMap::new(),
                },
                vec![Diagnostic {
                    level: Level::Error,
                    message: format!("cannot resolve path: {e}"),
                    span: Span::default(),
                    file: abs.display().to_string(),
                    hint: String::new(),
                }],
            );
        }
    };

    let project_root = abs_main
        .parent()
        .map(|p| p.to_path_buf())
        .unwrap_or_default();

    let mut r = Resolver {
        project_root,
        parsed: HashMap::new(),
        in_progress: HashSet::new(),
        diags: Vec::new(),
    };

    // Read main file
    let main_source = match fs::read_to_string(&abs_main) {
        Ok(s) => s,
        Err(e) => {
            return (
                Program {
                    main: SourceFile {
                        items: Vec::new(),
                        span: Span::default(),
                    },
                    modules: HashMap::new(),
                },
                vec![Diagnostic {
                    level: Level::Error,
                    message: format!("cannot read file: {e}"),
                    span: Span::default(),
                    file: main_file.to_string(),
                    hint: String::new(),
                }],
            );
        }
    };

    // Parse main file
    let (main_ast, parse_errors) = haira_parser::parse(&main_source);
    if !parse_errors.is_empty() {
        let diags = parse_errors
            .into_iter()
            .map(|e| Diagnostic {
                level: Level::Error,
                message: e.message,
                span: e.span,
                file: main_file.to_string(),
                hint: e.hint,
            })
            .collect();
        return (
            Program {
                main: SourceFile {
                    items: Vec::new(),
                    span: Span::default(),
                },
                modules: HashMap::new(),
            },
            diags,
        );
    }

    // Resolve imports from main file
    let abs_main_str = abs_main.display().to_string();
    r.resolve_imports(&main_ast, &abs_main_str);

    let diags = r.diags;
    let program = Program {
        main: main_ast,
        modules: r.parsed,
    };

    (program, diags)
}

/// Returns the short name for an import path (last segment).
///
/// E.g., `"models/user"` -> `"user"`, `"io"` -> `"io"`.
pub fn module_name(import_path: &str) -> &str {
    import_path
        .rsplit('/')
        .next()
        .unwrap_or(import_path)
}

// ---------------------------------------------------------------------------
// Program methods
// ---------------------------------------------------------------------------

impl Program {
    /// Returns all items from the main file and imported modules, suitable for
    /// passing to codegen. Imported types and functions are included so they
    /// appear in the generated output.
    ///
    /// Visibility rules:
    ///   - Only `pub` items are exported from modules.
    ///   - Selective imports (`import { X, Y } from "m"`) only include named items.
    ///   - Glob imports (`import * from "m"`) include all pub items.
    ///   - Basic imports (`import "m"`) include all pub items (accessed as `m.X`).
    ///
    /// Ordering: transitive dependencies first, then direct imports in source
    /// order, then main file items.
    pub fn merged_items(&self) -> Vec<Item> {
        // Build an ordered list of import paths from the main file to ensure
        // deterministic merge order (following source order of import decls).
        let mut import_order: Vec<String> = Vec::new();
        let mut import_map: HashMap<String, &haira_ast::ImportDecl> = HashMap::new();

        for item in &self.main.items {
            if let ItemKind::ImportDecl(ref imp) = item.node {
                import_map.insert(imp.path.clone(), imp);
                if self.modules.contains_key(&imp.path) {
                    import_order.push(imp.path.clone());
                }
            }
        }

        let mut items: Vec<Item> = Vec::new();

        // Track which modules are directly imported by main.
        let direct_imports: HashSet<&str> = import_order.iter().map(|s| s.as_str()).collect();

        // Add transitive modules first (not directly imported by main).
        // These are dependencies of dependencies -- e.g. tools and providers
        // imported by an agents file. They must come before the direct imports
        // so that providers and tools are registered before agents that
        // reference them.
        for module in self.modules.values() {
            if direct_imports.contains(module.path.as_str()) {
                continue;
            }
            for item in &module.file.items {
                match &item.node {
                    ItemKind::ImportDecl(_) | ItemKind::ExportDecl(_) => continue,
                    _ => {}
                }
                if !is_item_public(item) {
                    continue;
                }
                items.push(item.clone());
            }
        }

        // Add directly imported module items in source order.
        for path in &import_order {
            let module = &self.modules[path];
            let imp = import_map[&module.path];

            // Build a set of selectively imported names (if any).
            let selective_names: Option<HashSet<&str>> = if !imp.names.is_empty() {
                Some(imp.names.iter().map(|n| n.node.as_str()).collect())
            } else {
                None
            };

            for item in &module.file.items {
                match &item.node {
                    ItemKind::ImportDecl(_) | ItemKind::ExportDecl(_) => continue,
                    _ => {}
                }

                // Check pub visibility.
                if !is_item_public(item) {
                    continue;
                }

                // If selective import, only include named items.
                if let Some(ref names) = selective_names {
                    let name = item_name(item);
                    match name {
                        Some(n) if names.contains(n) => {}
                        _ => continue,
                    }
                }

                items.push(item.clone());
            }
        }

        // Add main file items.
        items.extend(self.main.items.iter().cloned());

        items
    }
}

// ---------------------------------------------------------------------------
// Internal resolver
// ---------------------------------------------------------------------------

struct Resolver {
    project_root: PathBuf,
    parsed: HashMap<String, Module>,
    in_progress: HashSet<String>,
    diags: Vec<Diagnostic>,
}

impl Resolver {
    fn add_error(&mut self, msg: impl Into<String>, file: &str, span: Span) {
        self.diags.push(Diagnostic {
            level: Level::Error,
            message: msg.into(),
            span,
            file: file.to_string(),
            hint: String::new(),
        });
    }

    fn resolve_imports(&mut self, file: &SourceFile, file_path: &str) {
        // Collect imports first to avoid borrow issues with &self.
        let imports: Vec<(String, Span)> = file
            .items
            .iter()
            .filter_map(|item| {
                if let ItemKind::ImportDecl(ref imp) = item.node {
                    Some((imp.path.clone(), item.span))
                } else {
                    None
                }
            })
            .collect();

        for (path, span) in imports {
            // Skip stdlib modules.
            if is_stdlib_module(&path) {
                continue;
            }

            // Cycle detection (check before "already resolved" since in-progress
            // modules are also in parsed).
            if self.in_progress.contains(&path) {
                self.add_error(
                    format!("circular import: \"{path}\""),
                    file_path,
                    span,
                );
                continue;
            }

            // Already fully resolved?
            if self.parsed.contains_key(&path) {
                continue;
            }

            // Resolve file path.
            let resolved = self.resolve_file_path(&path);
            let resolved = match resolved {
                Some(r) => r,
                None => {
                    self.add_error(
                        format!("cannot resolve import \"{path}\": file not found"),
                        file_path,
                        span,
                    );
                    continue;
                }
            };

            // Parse the imported file.
            self.in_progress.insert(path.clone());
            let module = self.parse_module(&path, &resolved);

            if let Some(module) = module {
                self.parsed.insert(path.clone(), module);
                // Recursively resolve imports in the module (while still
                // in-progress for cycle detection).
                // We need to clone the file to avoid borrow issues.
                let module_file = self.parsed[&path].file.clone();
                let resolved_clone = resolved.clone();
                self.resolve_imports(&module_file, &resolved_clone);
            }
            self.in_progress.remove(&path);
        }
    }

    fn resolve_file_path(&self, import_path: &str) -> Option<String> {
        // Try <root>/<path>.haira
        let candidate = self.project_root.join(format!("{import_path}.haira"));
        if candidate.is_file() {
            return Some(candidate.display().to_string());
        }

        // Try <root>/<path>/mod.haira (directory module)
        let candidate = self.project_root.join(import_path).join("mod.haira");
        if candidate.is_file() {
            return Some(candidate.display().to_string());
        }

        None
    }

    fn parse_module(&mut self, import_path: &str, file_path: &str) -> Option<Module> {
        let source = match fs::read_to_string(file_path) {
            Ok(s) => s,
            Err(e) => {
                self.diags.push(Diagnostic {
                    level: Level::Error,
                    message: format!("cannot read \"{file_path}\": {e}"),
                    span: Span::default(),
                    file: file_path.to_string(),
                    hint: String::new(),
                });
                return None;
            }
        };

        let (file_ast, parse_errors) = haira_parser::parse(&source);
        if !parse_errors.is_empty() {
            for e in parse_errors {
                self.diags.push(Diagnostic {
                    level: Level::Error,
                    message: e.message,
                    span: e.span,
                    file: file_path.to_string(),
                    hint: e.hint,
                });
            }
            return None;
        }

        Some(Module {
            path: import_path.to_string(),
            file_path: file_path.to_string(),
            file: file_ast,
        })
    }
}

// ---------------------------------------------------------------------------
// Visibility & naming helpers
// ---------------------------------------------------------------------------

/// Returns whether a top-level item is marked as pub (exported).
///
/// Agentic declarations (provider, tool, agent, workflow) are always public.
pub fn is_item_public(item: &Item) -> bool {
    match &item.node {
        ItemKind::FunctionDef(f) => f.is_public,
        ItemKind::TypeDef(t) => t.is_public,
        ItemKind::EnumDef(e) => e.is_public,
        ItemKind::MethodDef(_) => true,       // methods follow their type's visibility
        ItemKind::ProviderDecl(_) => true,     // agentic declarations are always public
        ItemKind::ToolDecl(_) => true,
        ItemKind::AgentDecl(_) => true,
        ItemKind::WorkflowDecl(_) => true,
        ItemKind::TestDecl(_) => false,        // tests are never exported
        ItemKind::TypeAlias(_) => true,        // type aliases are always public for now
        ItemKind::ItemStatement(_) => true,    // top-level statements (vars) are public for now
        ItemKind::ImportDecl(_) => false,
        ItemKind::ExportDecl(_) => false,
    }
}

/// Extracts the name from a top-level item, or `None` if unnamed.
pub fn item_name(item: &Item) -> Option<&str> {
    match &item.node {
        ItemKind::FunctionDef(f) => Some(&f.name.node),
        ItemKind::TypeDef(t) => Some(&t.name.node),
        ItemKind::EnumDef(e) => Some(&e.name.node),
        ItemKind::MethodDef(m) => Some(&m.name.node),
        ItemKind::TypeAlias(a) => Some(&a.name.node),
        ItemKind::ProviderDecl(p) => Some(&p.name.node),
        ItemKind::ToolDecl(t) => Some(&t.name.node),
        ItemKind::AgentDecl(a) => Some(&a.name.node),
        ItemKind::WorkflowDecl(w) => Some(&w.name.node),
        ItemKind::TestDecl(t) => Some(&t.name.node),
        _ => None,
    }
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;
    use haira_ast::{
        AgentDecl, EnumDef, FunctionDef, ItemKind, MethodDef, ProviderDecl, ToolDecl, TypeAlias,
        TypeDef, WorkflowDecl, TestDecl, Block,
    };
    use haira_errors::{Span, Spanned};

    // -- module_name tests ---------------------------------------------------

    #[test]
    fn module_name_simple() {
        assert_eq!(module_name("io"), "io");
    }

    #[test]
    fn module_name_nested() {
        assert_eq!(module_name("models/user"), "user");
    }

    #[test]
    fn module_name_deeply_nested() {
        assert_eq!(module_name("a/b/c/d"), "d");
    }

    #[test]
    fn module_name_empty_string() {
        assert_eq!(module_name(""), "");
    }

    // -- is_item_public tests ------------------------------------------------

    fn make_item(kind: ItemKind) -> Item {
        Spanned::new(kind, Span::default())
    }

    #[test]
    fn is_item_public_function_private() {
        let item = make_item(ItemKind::FunctionDef(FunctionDef {
            is_public: false,
            name: Spanned::new("foo".to_string(), Span::default()),
            params: vec![],
            return_ty: None,
            body: Block {
                statements: vec![],
                span: Span::default(),
            },
        }));
        assert!(!is_item_public(&item));
    }

    #[test]
    fn is_item_public_function_pub() {
        let item = make_item(ItemKind::FunctionDef(FunctionDef {
            is_public: true,
            name: Spanned::new("foo".to_string(), Span::default()),
            params: vec![],
            return_ty: None,
            body: Block {
                statements: vec![],
                span: Span::default(),
            },
        }));
        assert!(is_item_public(&item));
    }

    #[test]
    fn is_item_public_typedef_private() {
        let item = make_item(ItemKind::TypeDef(TypeDef {
            is_public: false,
            name: Spanned::new("Foo".to_string(), Span::default()),
            fields: vec![],
        }));
        assert!(!is_item_public(&item));
    }

    #[test]
    fn is_item_public_typedef_pub() {
        let item = make_item(ItemKind::TypeDef(TypeDef {
            is_public: true,
            name: Spanned::new("Foo".to_string(), Span::default()),
            fields: vec![],
        }));
        assert!(is_item_public(&item));
    }

    #[test]
    fn is_item_public_enum_private() {
        let item = make_item(ItemKind::EnumDef(EnumDef {
            is_public: false,
            name: Spanned::new("Color".to_string(), Span::default()),
            variants: vec![],
        }));
        assert!(!is_item_public(&item));
    }

    #[test]
    fn is_item_public_enum_pub() {
        let item = make_item(ItemKind::EnumDef(EnumDef {
            is_public: true,
            name: Spanned::new("Color".to_string(), Span::default()),
            variants: vec![],
        }));
        assert!(is_item_public(&item));
    }

    #[test]
    fn is_item_public_method_always() {
        let item = make_item(ItemKind::MethodDef(MethodDef {
            type_name: Spanned::new("Foo".to_string(), Span::default()),
            name: Spanned::new("bar".to_string(), Span::default()),
            params: vec![],
            return_ty: None,
            body: Block {
                statements: vec![],
                span: Span::default(),
            },
        }));
        assert!(is_item_public(&item));
    }

    #[test]
    fn is_item_public_provider_always() {
        let item = make_item(ItemKind::ProviderDecl(ProviderDecl {
            name: Spanned::new("OpenAI".to_string(), Span::default()),
            fields: vec![],
        }));
        assert!(is_item_public(&item));
    }

    #[test]
    fn is_item_public_tool_always() {
        let item = make_item(ItemKind::ToolDecl(ToolDecl {
            decorators: vec![],
            name: Spanned::new("search".to_string(), Span::default()),
            params: vec![],
            return_ty: None,
            description: "A tool".to_string(),
            body: None,
        }));
        assert!(is_item_public(&item));
    }

    #[test]
    fn is_item_public_agent_always() {
        let item = make_item(ItemKind::AgentDecl(AgentDecl {
            name: Spanned::new("MyAgent".to_string(), Span::default()),
            fields: vec![],
        }));
        assert!(is_item_public(&item));
    }

    #[test]
    fn is_item_public_workflow_always() {
        let item = make_item(ItemKind::WorkflowDecl(WorkflowDecl {
            name: Spanned::new("deploy".to_string(), Span::default()),
            trigger: None,
            decorators: vec![],
            params: vec![],
            return_ty: None,
            description: String::new(),
            body: Block {
                statements: vec![],
                span: Span::default(),
            },
            hooks: vec![],
        }));
        assert!(is_item_public(&item));
    }

    #[test]
    fn is_item_public_test_never() {
        let item = make_item(ItemKind::TestDecl(TestDecl {
            name: Spanned::new("my test".to_string(), Span::default()),
            body: Block {
                statements: vec![],
                span: Span::default(),
            },
        }));
        assert!(!is_item_public(&item));
    }

    #[test]
    fn is_item_public_type_alias_always() {
        let item = make_item(ItemKind::TypeAlias(TypeAlias {
            name: Spanned::new("Alias".to_string(), Span::default()),
            ty: Spanned::new(haira_ast::Type::Named("int".to_string()), Span::default()),
        }));
        assert!(is_item_public(&item));
    }

    // -- item_name tests -----------------------------------------------------

    #[test]
    fn item_name_function() {
        let item = make_item(ItemKind::FunctionDef(FunctionDef {
            is_public: false,
            name: Spanned::new("greet".to_string(), Span::default()),
            params: vec![],
            return_ty: None,
            body: Block {
                statements: vec![],
                span: Span::default(),
            },
        }));
        assert_eq!(item_name(&item), Some("greet"));
    }

    #[test]
    fn item_name_typedef() {
        let item = make_item(ItemKind::TypeDef(TypeDef {
            is_public: false,
            name: Spanned::new("User".to_string(), Span::default()),
            fields: vec![],
        }));
        assert_eq!(item_name(&item), Some("User"));
    }

    #[test]
    fn item_name_enum() {
        let item = make_item(ItemKind::EnumDef(EnumDef {
            is_public: false,
            name: Spanned::new("Color".to_string(), Span::default()),
            variants: vec![],
        }));
        assert_eq!(item_name(&item), Some("Color"));
    }

    #[test]
    fn item_name_import_returns_none() {
        let item = make_item(ItemKind::ImportDecl(haira_ast::ImportDecl {
            path: "io".to_string(),
            alias: None,
            names: vec![],
            is_glob: false,
        }));
        assert_eq!(item_name(&item), None);
    }

    // -- stdlib module detection tests ---------------------------------------

    #[test]
    fn stdlib_modules_detected() {
        let known = [
            "io", "http", "mcp", "env", "json", "postgres", "slack", "excel",
            "time", "string", "regex", "math", "conv", "array", "map", "log",
            "ui", "vector", "observe", "fs", "gitlab", "github", "langfuse",
            "algolia", "meilisearch", "store",
        ];
        for name in &known {
            assert!(
                is_stdlib_module(name),
                "\"{name}\" should be a stdlib module"
            );
        }
    }

    #[test]
    fn non_stdlib_modules_not_detected() {
        assert!(!is_stdlib_module("models"));
        assert!(!is_stdlib_module("models/user"));
        assert!(!is_stdlib_module("utils"));
        assert!(!is_stdlib_module(""));
    }

    // -- resolve with no imports (just main file) ----------------------------

    #[test]
    fn resolve_no_imports() {
        // Create a temporary file with simple Haira code (no imports).
        let dir = std::env::temp_dir().join("haira_resolver_test");
        let _ = fs::create_dir_all(&dir);
        let main_path = dir.join("main.haira");
        fs::write(&main_path, "fn main() {\n}\n").unwrap();

        let (program, diags) = resolve(main_path.to_str().unwrap());

        // Should succeed with no diagnostics.
        assert!(
            diags.is_empty(),
            "expected no diagnostics, got: {diags:?}"
        );
        // The main AST should have items.
        assert!(!program.main.items.is_empty());
        // No modules resolved (no imports).
        assert!(program.modules.is_empty());

        // Cleanup.
        let _ = fs::remove_file(&main_path);
        let _ = fs::remove_dir(&dir);
    }

    #[test]
    fn resolve_nonexistent_file() {
        let (_, diags) = resolve("/tmp/haira_resolver_test_nonexistent_abc123.haira");
        assert!(!diags.is_empty());
        assert!(diags[0].message.contains("cannot resolve path"));
    }

    #[test]
    fn resolve_stdlib_import_no_file_lookup() {
        // An import of a stdlib module should not trigger file resolution.
        let dir = std::env::temp_dir().join("haira_resolver_test_stdlib");
        let _ = fs::create_dir_all(&dir);
        let main_path = dir.join("main.haira");
        fs::write(&main_path, "import \"io\"\n\nfn main() {\n}\n").unwrap();

        let (program, diags) = resolve(main_path.to_str().unwrap());

        assert!(
            diags.is_empty(),
            "expected no diagnostics for stdlib import, got: {diags:?}"
        );
        // "io" should NOT appear as a resolved module.
        assert!(!program.modules.contains_key("io"));

        let _ = fs::remove_file(&main_path);
        let _ = fs::remove_dir(&dir);
    }
}
