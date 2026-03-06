# Haira Language Implementation Roadmap

This document maps specification chapters to implementation releases.

## Release Strategy

Each release implements one or more chapters from the specification. The spec is the "north star" — all features are fully documented before implementation begins.

## Compiler

The Haira compiler is written in Go with zero external dependencies. The pipeline is:

```
.haira source → Lexer → Parser → Resolver → Checker → Go Codegen → go build → Native binary
```

## Release Plan

| Release | Chapters | Features | Status |
|---------|----------|----------|--------|
| **v0.1** | 1-2 | Lexer, parser, tokens, basic expressions | Done |
| **v0.2** | 3 | Type system, primitives, structs, enums, type aliases | Done |
| **v0.3** | 4-5 | Expressions, statements, control flow, pattern matching | Done |
| **v0.4** | 6-7 | Functions, methods, error handling (try/catch, defer, ?) | Done |
| **v0.5** | 8 | Concurrency (spawn blocks, channels, select) | Done |
| **v0.6** | 9-10 | Modules (4 import forms), resolver, standard library | Done |
| **v0.7** | 11-14 | Providers, tools, agents, workflows | Done |
| **v0.8** | 15-16 | Chatbot patterns, generative UI, ARP protocol | Done |
| **v0.9** | 17-18 | Grammar formalization, compiler architecture docs | In progress |
| **v1.0** | All | Production hardening, optimizations, complete docs | Planned |

## Chapter Details

### v0.1: Foundation (Done)
**Chapters:** 1 (Overview), 2 (Lexical Structure)

- Hand-written lexer (`compiler/internal/lexer/`)
- Token types — 61 keywords + 24 operators
- Recursive descent parser with Pratt expression parsing
- Basic CLI (build, run, parse, lex)

### v0.2: Type System (Done)
**Chapter:** 3 (Types)

- Primitive types (int, i8-i64, u8-u64, float, f32, f64, bool, string)
- Composite types (arrays, maps, tuples)
- Option types (`T?`)
- Struct and enum definitions
- Type aliases (`type Name = Type`)
- Type inference for locals

### v0.3: Expressions & Statements (Done)
**Chapters:** 4 (Expressions), 5 (Statements)

- All expression types including pipe (`|>`), range (`..`, `..=`), spawn
- All statement types including step, assert
- Control flow (if, for, while, match)
- Pattern matching with or-patterns, range patterns, guards
- Bitwise operators (`&`, `|`, `^`, `~`, `<<`, `>>`)

### v0.4: Functions (Done)
**Chapters:** 6 (Functions), 7 (Error Handling)

- Function declarations with default parameters
- Multiple return values
- Closures / lambdas
- Methods (dot-attach with implicit `self`)
- Error propagation (`?` operator, panic-based)
- Try/catch, defer, errdefer

### v0.5: Concurrency (Done)
**Chapter:** 8 (Concurrency)

- `spawn { }` blocks for parallel execution
- Channel types (`chan<T>`)
- `select` statement
- Async blocks (AST support)

### v0.6: Modules (Done)
**Chapters:** 9 (Modules), 10 (Standard Library)

- Import resolution (4 forms: basic, aliased, selective, glob)
- Export declarations
- `pub` visibility modifier
- Core stdlib: io, string, array, map, math, conv, fs, os, time, json, http, regex, env
- Extended stdlib: postgres, sqlite, d1, excel, vector, slack, github, gitlab, algolia, meilisearch, langfuse

### v0.7: Agentic Primitives (Done)
**Chapters:** 11 (Providers), 12 (Tools), 13 (Agents), 14 (Workflows)

- Provider declarations (OpenAI, Azure, Ollama, MCP)
- Tool declarations with mandatory `"""..."""` descriptions
- Agent declarations with memory, handoffs, structured output
- Workflow declarations with HTTP triggers (`@get`, `@post`, `@put`, `@delete`, `@webhook`)
- Lifecycle hooks (`onerror`, `onsuccess`, `oncancel`)
- Step blocks with `@retry`
- MCP client and server support

### v0.8: UI & Patterns (Done)
**Chapters:** 15 (Chatbot Patterns), 16 (Generative UI)

- Streaming workflows (`-> stream`, SSE)
- ARP (Agentic Rendering Protocol) — WebSocket + SSE
- Generative UI components (tables, charts, forms, status cards, etc.)
- Auto UI for workflows (`/_ui/`)
- `@webui` decorator
- Observability with Langfuse integration

### v0.9: Formalization (In Progress)
**Chapters:** 17 (Grammar), 18 (Compiler Architecture)

- Complete EBNF grammar
- 8-pass codegen emission order documentation
- LSP implementation (hover, diagnostics, go-to-definition, formatting)
- Tree-sitter grammar for editor support
- Test framework (`test` blocks, `assert`)

### v1.0: Production Ready (Planned)

- Comprehensive test suite
- Error message quality improvements
- Performance optimization
- Complete documentation
- Cloudflare Workers target (`--target workers`)
- Orchestration daemon (deploy, ps, logs, etc.)

## Build Commands

```bash
# Build the specification PDF
cd spec/latex
make

# Build with quick mode (development)
make quick

# View the PDF
make view
```

## Specification Location

- **LaTeX source:** `spec/latex/`
- **PDF output:** `spec/latex/build/haira-spec.pdf`
- **Chapters:** `spec/latex/chapters/`

## Contributing

1. Spec changes require updating the LaTeX source
2. Implementation must match the specification
3. Each release requires all chapter features to be complete
4. Tests must cover specification examples
