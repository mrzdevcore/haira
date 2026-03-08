---
name: haira-compiler
description: >
  Deep knowledge of the Haira compiler internals — lexer, parser, checker,
  codegen pipeline, runtime library, and AST design. Use this skill when
  contributing to the Haira compiler, adding new language features, fixing
  compiler bugs, working on the Go codegen, modifying the runtime, or
  understanding compiler architecture. Triggers on: compiler internals,
  lexer, parser, checker, codegen, AST, token, runtime, primitive, stdlib.
user-invocable: true
argument-hint: "[compiler task or question]"
---

# Haira Compiler — Development Guide

## Pipeline (immutable order)

```
Source (.haira) -> Lexer -> Parser -> Resolver -> Checker -> Codegen -> go build -> Binary
```

**No AI phase** in the compiler. All LLM interaction happens at **runtime** through agent method calls.

## Directory Structure

```
compiler/
  main.go                          # CLI entry: build, run, parse, check, lex, emit, lsp
  go.mod                           # Go 1.22, zero external dependencies
  internal/
    token/token.go                 # TokenKind constants, Token struct
    lexer/lexer.go                 # Hand-written scanner
    ast/ast.go                     # AST nodes (Go interfaces as sum types), Span, Spanned[T]
    parser/parser.go               # Recursive descent + Pratt expression parsing
    resolver/resolver.go           # Multi-file import resolution
    checker/checker.go             # Type checking + semantic analysis (2 passes)
    codegen/
      codegen.go                   # 8-pass emission orchestrator
      emitter.go                   # Low-level Go source emitter (buffer + indent)
      expressions.go               # Expression-to-Go conversion
      statements.go                # Statement emission
      agentic.go                   # Provider/tool/agent/workflow emission
      stdlib.go                    # Standard library function mapping
      types.go                     # Haira-type-to-Go-type conversion
      project.go                   # Temp project setup + go build invocation
    driver/driver.go               # Pipeline orchestrator
    errors/errors.go               # Diagnostic system (Diagnostic, PrettyPrint)
    formatter/                     # Source code formatter
    lsp/                           # Language server protocol
    manifest/                      # package.haira handling
    runtime/
      runtime.go                   # go:embed bundle.tar.gz
      embedded/haira/*.go.src      # Runtime source files (renamed to avoid compilation)
```

## Package Dependency Rules

| Package | Can Import |
|---------|------------|
| `token` | (none) |
| `ast` | (none) |
| `lexer` | token |
| `parser` | token, ast |
| `resolver` | ast |
| `checker` | ast, token, errors |
| `codegen` | ast, token |
| `errors` | ast |
| `driver` | all above |
| `lsp` | driver, ast, checker, errors |

## Lexer

Hand-written scanner. Converts UTF-8 source to token stream.

**Key token kinds:**
- Agentic: `Provider`, `Tool`, `Agent`, `Workflow`
- Core: `Fn`, `Struct`, `Enum`, `Type`, `Import`, `Export`, `From`, `Pub`, `If`, `Else`, `For`, `While`, `Match`, `Return`, `Break`, `Continue`, `True`, `False`, `Nil`, `And`, `Or`, `Not`, `In`, `Spawn`, `Select`, `Try`, `Catch`, `Defer`, `Step`, `Test`, `Assert`, `Const`
- Operators: `Plus`, `Minus`, `Star`, `Slash`, `Percent`, `Eq`, `EqEq`, `Ne`, `Lt`, `Gt`, `Le`, `Ge`, `PipeArrow`, `Amp`, `Caret`, `Tilde`, `Shl`, `Shr`, `Pipe`, `Arrow`, `FatArrow`, `DotDot`, `DotDotEq`, `Dot`, `Question`, `At`, `Ellipsis`
- Literals: `Int`, `Float`, `String`, `InterpolatedString`, `TripleQuoteString`

**Special handling:**
- `${expr}` interpolation: nested brace depth tracking
- `"""..."""` triple-quoted strings: auto-dedenting
- `/* /* nested */ */` block comments
- Numeric underscores: `1_000_000`
- Hex/binary/octal: `0xFF`, `0b1010`, `0o77`

## Parser

Recursive descent for declarations/statements + Pratt (precedence-climbing) for expressions.

### AST Node Categories

**Top-level declarations:**
- `TypeDef` — struct with typed fields
- `EnumDef` — enum with variants
- `TypeAlias` — `type Name = Type`
- `FunctionDef` — regular functions
- `MethodDef` — `Type.method()` with implicit `self`
- `ImportDecl` — 4 import forms
- `ExportDecl` — `export { ... }`
- `ProviderDecl` — provider blocks
- `ToolDecl` — tool functions with description
- `AgentDecl` — agent configuration blocks
- `WorkflowDecl` — workflow with optional decorator trigger
- `TestDecl` — `test "name" { ... }`

**Expressions (27 types):**
Literals, identifiers, binary/unary ops, calls, method calls, field access, indexing, pipe, lambda, match, if-expressions, lists, maps, struct instances, ranges, error propagation (`?`), `orelse`, `some`/`none`, spawn, select, async.

**Statements (14 types):**
Assignment, let/const, if/else, for-in, while, return, try/catch, defer/errdefer, match, break/continue, step, assert, expr statements.

**Patterns (6 types):**
Wildcard (`_`), literal, identifier, constructor, or-pattern (`A | B`), range (`1..5`).

### Operator Precedence in Parser (14 levels, lowest to highest)

1. `orelse`
2. `|>` (pipe)
3. `or` / `||` (logical OR)
4. `and` / `&&` (logical AND)
5. `|` (bitwise OR)
6. `^` (bitwise XOR)
7. `&` (bitwise AND)
8. `==`, `!=`
9. `<`, `>`, `<=`, `>=`, `..`, `..=`
10. `<<`, `>>`
11. `+`, `-`
12. `*`, `/`, `%`
13. `-`, `not`, `~` (unary)
14. `()`, `[]`, `.`, `?` (postfix)

### AST Design

Every node carries a `Span`:
```go
type Expr struct {
    Node ExprNode
    Span Span     // Required for error reporting
}

type Span struct {
    Start int  // byte offset
    End   int
}
```

Go interfaces serve as sum types:
```go
type ExprNode interface { exprKind() }
type StmtNode interface { stmtKind() }
```

## Resolver

Multi-file module resolution:

1. Read main source, extract imports
2. Resolve each import: stdlib -> project-local -> external
3. Recursively resolve transitive imports
4. Detect circular dependencies (compile error)
5. Filter exports: only `pub` items and agentic declarations visible
6. Produce `Program` with merged items in dependency order

**Resolution order:**
1. Transitive module items (deepest first)
2. Direct import items (source order)
3. Main file items

## Checker

Two-pass type checking:

### Pass 1 — Registration
- Register struct types with field types
- Register enum types with variants
- Register function signatures
- Register provider, tool, agent, workflow declarations
- Register global variables (inferred from initializers)

### Pass 2 — Validation
- Type inference for expressions
- Variable definition tracking
- Const immutability enforcement
- Agent field validation (model references provider, tools exist, known fields)
- Enum exhaustiveness warnings in match
- Return type checking
- Method `self` binding

### Agentic Validation
- **Providers:** Warn on unknown fields. Valid: `model`, `api_key`, `backend`, `host`, `temperature`, `max_tokens`, `transport`, `command`, `args`, `url`, `endpoint`, `env`, `headers`, `input_token_cost`, `output_token_cost`, `api_version`
- **Agents:** Require `provider` field (or `model`), validate tool references exist, check handoff targets exist, warn on unknown fields
- **Tools:** Enforce mandatory triple-quoted description
- **Workflows:** Validate decorator triggers, type-check body

### Type System
```
Primitive: int, float, string, bool, any, void, error
Compound: [T] (list), {K: V} (map)
Named: struct, enum
Function: (T, U) -> V
```

Generic types are parse-accepted but erased to `any` in v1.

## Code Generation

### 8-Pass Emission Order (DO NOT reorder)

1. **Providers** -> `var Claude = &haira.Provider{...}`
2. **Tools** -> tool functions + `haira.ToolDef` vars
3. **Agents** -> agent vars + init functions (topologically sorted)
4. **Workflows** -> HTTP handler functions + `haira.WorkflowDef` vars
5. **Top-level variables** -> global `var` declarations
6. **Types and enums** -> struct definitions, enum iota constants
7. **Methods** -> receiver methods
8. **Functions** -> regular functions + `main()`

### Agent Initialization (Topological Sort)

Agents with handoffs must be initialized **after** their handoff targets:
```
// If FrontDesk.handoffs = [BillingAgent, TechAgent]
// Init order: BillingAgent -> TechAgent -> FrontDesk
```

### Tool Schema Generation

Compiler generates JSON schema from Haira types:
| Haira Type | JSON Schema |
|-----------|-------------|
| `int` | `"integer"` |
| `float` | `"number"` |
| `bool` | `"boolean"` |
| `string` | `"string"` |
| `[]T` | `"array"` with `items` |
| struct | `"object"` with `properties` |

### Workflow Codegen

- Workflows with triggers generate HTTP handler functions
- Streaming workflows (`-> stream`) generate both:
  - `StreamHandler` — SSE endpoint
  - `Handler` — JSON fallback endpoint
- `@webui` generates form UI registration

### Import Detection (Multi-Pass)

Codegen scans generated Go code to determine imports:
- `fmt` — string interpolation, print calls
- `encoding/json` — JSON marshal/unmarshal
- `sync` — spawn blocks (WaitGroup)
- `haira` — runtime library (always needed for agentic code)

### Go Compilation Process

1. Create temp Go project directory with `go.mod`
2. Extract embedded runtime bundle (`bundle.tar.gz`) -> `haira/` package
3. Write generated `main.go`
4. Run `go build -o <output>` to produce binary
5. Move binary to output path, clean up temp directory

## Runtime Library

### Two Layers

**Primitive** (`primitive/haira/`): Core runtime — always included
**Stdlib** (`stdlib/`): External integrations — tree-shaken at compile time

See [references/runtime.md](references/runtime.md) for detailed runtime file listing.

### Store Backend Pattern
```go
func init() {
    RegisterStoreBackend("postgres", func(dsn string) Store { /* ... */ })
}
```

### Bundling Process
```
1. make bundle-runtime -> copies primitive/ + stdlib/ to staging
2. UI SDK built (bun build) and included
3. All archived as bundle.tar.gz
4. Embedded via go:embed in compiler binary
5. At compile time: extract -> write main.go -> go build
```

## Diagnostics

```go
errors.Diagnostic{
    Level:   errors.Error,    // Error, Warning, Info
    Message: "type mismatch",
    Span:    node.Span,
    File:    filename,
    Hint:    "expected int, got string",
}
```

## CLI Commands

```bash
haira build [file] [-o output]   # Compile to native binary
haira run [file]                  # Compile and execute
haira check [file]                # Type-check only
haira emit [file]                 # Show generated Go code
haira test [file]                 # Run test blocks
haira fmt [file]                  # Format source
haira parse [file]                # Show AST
haira lex [file]                  # Show tokens
haira init                        # Create package.haira
haira lsp                         # Start language server (stdio)
haira version                     # Show version
```

## Common Development Tasks

### Adding a New AST Node

1. `internal/token/token.go` — Add TokenKind (if new keyword)
2. `internal/lexer/` — Add keyword recognition
3. `internal/ast/ast.go` — Define node struct with Span, add marker method
4. `internal/parser/` — Implement parsing rule
5. `internal/checker/` — Add type checking
6. `internal/codegen/` — Add Go code generation
7. `make dev` — Verify all passes

### Adding a New Standard Library Module

1. `primitive/haira/<module>.go` — Runtime implementation
2. `internal/codegen/stdlib.go` — Map qualified names in `resolveQualified()`
3. `internal/codegen/stdlib.go` — Register in `IsStdlibImport()`
4. `internal/codegen/expressions.go` — Add return type in `orelseReturnType()` if needed
5. Add example in `examples/`
6. `make dev && make build-examples`

### Adding an Agentic Feature

1. Define AST node with Span
2. Add token (if new keyword)
3. Update lexer and parser
4. Add checker validation
5. Add codegen in `internal/codegen/agentic.go`
6. Add runtime support in `primitive/haira/`
7. `make ci`

## Forbidden Actions

- Reorder compiler pipeline phases
- Change file extension from `.haira`
- Add semicolons to Haira syntax
- Add the `ai` keyword back
- Add external dependencies to the compiler
- Break the 8-pass codegen emission order

## Build Commands

```bash
make build            # Build compiler
make test             # Run tests
make dev              # fmt + vet + test
make ci               # vet + test + build-examples
make build-examples   # Compile all .haira examples
make install          # Install to $GOPATH/bin
```
