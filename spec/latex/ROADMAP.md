# Haira Language Implementation Roadmap

This document maps specification chapters to implementation releases.

## Release Strategy

Each release implements one or more chapters from the specification. The spec is the "north star" - all features are fully documented before implementation begins.

## Release Plan

| Release | Chapters | Features | Status |
|---------|----------|----------|--------|
| **v0.1** | 1-2 | Lexer, basic parser, tokens | Planned |
| **v0.2** | 3 | Type system, primitives, structs, enums | Planned |
| **v0.3** | 4-5 | Expressions, statements, control flow | Planned |
| **v0.4** | 6-7 | Functions, error handling | Planned |
| **v0.5** | 8 | Concurrency (spawn, channels, select) | Planned |
| **v0.6** | 9-10 | Modules, imports, standard library | Planned |
| **v0.7** | 11 | Advanced features (generics, constraints) | Planned |
| **v0.8** | 12-13 | AI integration, CIR | Planned |
| **v1.0** | 14-15 | Complete grammar, optimizations | Planned |

## Chapter Details

### v0.1: Foundation
**Chapters:** 1 (Overview), 2 (Lexical Structure)

- Lexer implementation
- Token types (keywords, literals, operators)
- Basic parser skeleton
- Hello World compilation

**Deliverables:**
- `haira-lexer` crate
- `haira-parser` crate (skeleton)
- Basic CLI

### v0.2: Type System
**Chapter:** 3 (Types)

- Primitive types (int, float, bool, string)
- Composite types (arrays, maps, tuples)
- Option types
- Struct and enum definitions
- Type inference for locals

**Deliverables:**
- `haira-types` crate
- `haira-typeck` crate (basic)

### v0.3: Expressions & Statements
**Chapters:** 4 (Expressions), 5 (Statements)

- All expression types
- All statement types
- Control flow (if, for, while, match)
- Pattern matching
- Variable declarations and assignments

**Deliverables:**
- Complete parser
- AST definitions
- Basic interpreter (optional)

### v0.4: Functions
**Chapters:** 6 (Functions), 7 (Error Handling)

- Function declarations
- Parameters and return types
- Multiple return values
- Closures
- Methods
- Error type and handling patterns

**Deliverables:**
- Function resolution
- Type checking for functions
- Error handling implementation

### v0.5: Concurrency
**Chapter:** 8 (Concurrency)

- spawn keyword
- Channels (buffered/unbuffered)
- select statement
- Basic synchronization

**Deliverables:**
- Runtime with goroutine-like tasks
- Channel implementation
- Async scheduler

### v0.6: Modules
**Chapters:** 9 (Modules), 10 (Standard Library)

- Import resolution
- Module system
- Standard library modules:
  - io, string, array, map
  - math, conv, fs, os
  - time, json, http

**Deliverables:**
- `haira-resolver` crate
- `haira-stdlib` crate
- Package structure

### v0.7: Advanced Features
**Chapter:** 11 (Advanced Features)

- Generics
- Type constraints (traits)
- Operator overloading
- Attributes
- Macros (basic)

**Deliverables:**
- Generic type system
- Constraint resolution
- Macro expansion

### v0.8: AI Integration
**Chapters:** 12 (AI Integration), 13 (CIR Specification)

- AI block syntax
- CIR parser and validator
- Backend interfaces:
  - llama.cpp
  - Ollama
  - Cloud APIs
- Caching system

**Deliverables:**
- `haira-ai` crate
- `haira-cir` crate
- AI backend implementations

### v1.0: Production Ready
**Chapters:** 14 (Grammar), 15 (Compiler Architecture)

- Complete grammar verification
- Optimization passes
- Error message quality
- Documentation
- Performance tuning

**Deliverables:**
- Production-ready compiler
- Complete documentation
- Benchmark suite
- VS Code extension

## Build Commands

```bash
# Build the specification PDF
cd spec/latex
make

# Build with quick mode (development)
make quick

# View the PDF
make view

# Watch for changes
make watch
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
