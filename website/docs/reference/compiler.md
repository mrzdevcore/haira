---
title: "Compiler Architecture"
description: "Haira compiler pipeline — lexer, parser, checker, Go codegen, and binary output."
---

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
  Resolver   →  Resolved imports + merged AST
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

### Resolver

Import resolver and module merger. Located in `compiler/internal/resolver/`.

Key responsibilities:
- Resolves `import` declarations to source files
- Handles all 4 import forms (basic, aliased, selective, glob)
- Merges multi-file programs into a single AST

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

### Formatter

Code formatter. Located in `compiler/internal/formatter/`.

Key responsibilities:
- Normalizes whitespace, indentation, and line breaks
- Preserves comments
- Invoked via `haira fmt`

## CLI Commands

### Core

```bash
haira build file.haira           # Compile to binary
haira build file.haira -o out    # Compile with custom output path
haira build file.haira --target workers     # Compile for Cloudflare Workers
haira build file.haira --target claude-code # Export to Claude Code format
haira run file.haira             # Compile and run
haira test file.haira            # Run test blocks
haira eval file.haira            # Run eval blocks (agent evaluation)
haira fmt file.haira             # Format source file
haira check file.haira           # Type check only
haira parse file.haira           # Show AST
haira lex file.haira             # Show tokens
haira emit file.haira            # Show generated Go
haira init                       # Create package.haira manifest
```

### IDE & UI

```bash
haira lsp                        # Start language server (stdio)
haira webui --connect host:port  # Serve web UI for a running server
haira console host:port          # Interactive terminal for running servers
```

### Orchestration

```bash
haira serve --port 8900          # Start orchestration daemon
haira deploy file.haira          # Deploy to orchestrator
haira ps                         # List deployments
haira stop <name>                # Stop a deployment
haira restart <name>             # Restart a deployment
haira logs <name> --follow       # Stream deployment logs
haira undeploy <name>            # Remove a deployment
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
