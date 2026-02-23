# Compiler Architecture

Haira compiles to Go, which then compiles to a native binary.

## Pipeline

```
.haira source
    ↓
  Lexer      →  Token stream
    ↓
  Parser     →  AST (Abstract Syntax Tree)
    ↓
  Checker    →  Type-checked AST + diagnostics
    ↓
  Codegen    →  Go source code
    ↓
  go build   →  Native binary
```

## Stages

### Lexer

Hand-written scanner that produces a stream of tokens. Located in `compiler/internal/lexer/`.

Key responsibilities:
- Tokenizes keywords, identifiers, literals, operators
- Handles string interpolation (`${expr}`)
- Tracks source positions for error reporting

### Parser

Recursive descent parser with Pratt expression parsing. Located in `compiler/internal/parser/`.

Key responsibilities:
- Builds the AST from the token stream
- Handles operator precedence via Pratt parsing
- Produces `Spanned[T]` nodes with source locations

### Checker

Type checker and semantic analyzer. Located in `compiler/internal/checker/`.

Key responsibilities:
- Type inference for local variables
- Type checking for function calls, operations
- Semantic validation (e.g., tool docstrings required)
- Enum exhaustiveness checking in match expressions

### Codegen

Go code generator. Located in `compiler/internal/codegen/`.

Key responsibilities:
- Emits valid Go source code from the AST
- 8-pass emission order: providers → tools → agents → workflows → vars → types/enums → methods → functions → main
- Topological sorting of agent init calls (handoff targets first)
- Import detection (multi-pass for fmt, encoding/json, sync, haira runtime)
- Streaming workflows generate both SSE and JSON fallback handlers
- Stdlib import rewriting (dev-time `haira-go-runtime/haira` → compile-time `haira-generated/haira`)

## CLI Commands

```bash
haira build file.haira     # Compile to binary
haira run file.haira       # Compile and run
haira parse file.haira     # Show AST
haira check file.haira     # Type check only
haira lex file.haira       # Show tokens
haira emit file.haira      # Show generated Go
haira lsp                  # Start language server
```

## Tree-shaking

Only imported stdlib packages are included in the compiled binary:
- If you don't `import "postgres"`, the Postgres driver is not compiled in
- Transitive dependencies are resolved automatically (`vector` pulls in `postgres`)
- SQLite is auto-included when workflows or server are detected

## Runtime

The Haira runtime (`primitive/haira/`) provides:
- Agent, Provider, Tool execution
- Memory management (conversation, summary)
- HTTP server with routing
- Store interface (SQLite default, Postgres optional)
- UI serving (embedded web components)
- MCP client/server support
- Observability hooks

The runtime is bundled as `bundle.tar.gz` and embedded in the compiler binary via `go:embed`.
