# Haira Programming Language

**Express intention, not mechanics.**

Haira is a programming language where you write what you want, and the compiler (powered by AI) figures out how to do it.

## Quick Example

```haira
User { name, email, active }

server = Server { port = 8080 }

routes {
    get("/users") {
        users = get_active_users() | sort_by_name | take(10)
        json(users)
    }

    post("/users") {
        user = User { body().name, body().email, true }
        save_user(user)
        json(user)
    }
}

server.start()
```

**Notice**: `get_active_users()`, `sort_by_name`, `save_user()` are **never defined**. The compiler understands your intent and generates them.

## How It Works

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  Your Code   │ --> │    Haira     │ --> │   Native     │
│  (Intent)    │     │   Compiler   │     │   Binary     │
└──────────────┘     └──────┬───────┘     └──────────────┘
                           │
                    ┌──────┴───────┐
                    │   Claude AI   │
                    │  (Interprets  │
                    │    Intent)    │
                    └──────────────┘
```

1. You write high-level, intent-based code
2. The compiler identifies undefined functions
3. AI (Claude) interprets your intent from function names and context
4. Generated code is cached for reproducibility
5. Everything compiles to a fast native binary

## Key Features

- **No imports** - The compiler finds everything automatically
- **No null** - Option types prevent null pointer errors
- **No boilerplate** - AI generates CRUD, transformations, I/O
- **Type inference** - Types exist but you rarely write them
- **Fast binaries** - Compiles to native code via LLVM
- **Reproducible** - AI outputs are cached and locked

## Project Structure

```
haira/
├── spec/                    # Language specification
│   ├── 01-overview.md
│   ├── 02-lexical-structure.md
│   ├── 03-types.md
│   ├── 04-expressions.md
│   ├── 05-statements.md
│   ├── 06-functions.md
│   ├── 07-error-handling.md
│   ├── 08-concurrency.md
│   ├── 09-modules.md
│   ├── 10-standard-library.md
│   ├── 11-auto-generation.md
│   ├── 12-grammar.md
│   ├── 13-compiler-architecture.md
│   ├── 14-implementation-plan.md
│   └── 15-ai-integration.md
├── examples/                # Example programs
│   ├── hello.haira
│   ├── web-api.haira
│   ├── cli-tool.haira
│   ├── data-processing.haira
│   └── concurrency.haira
├── crates/                  # Compiler implementation (Rust)
│   ├── haira-lexer/        # Tokenization
│   ├── haira-parser/       # AST generation
│   ├── haira-ast/          # AST definitions
│   ├── haira-resolver/     # Name resolution
│   ├── haira-ai/           # AI intent engine (Claude)
│   ├── haira-cir/          # Canonical IR
│   ├── haira-types/        # Type system
│   ├── haira-hir/          # High-level IR
│   ├── haira-mir/          # Mid-level IR
│   ├── haira-codegen/      # LLVM code generation
│   ├── haira-driver/       # Compiler driver
│   └── haira-cli/          # CLI interface
└── Cargo.toml              # Rust workspace
```

## Philosophy

1. **Express intention, not mechanics**
2. **Natural-thinking, not natural language**
3. **Fast prototyping with production-grade output**
4. **Reproducibility as a core feature**
5. **Native-speed binaries from high-level logic**
6. **Compiler absorbs complexity, developer writes clarity**

## Status

🚧 **Early Development** - Specification complete, implementation starting.

## Requirements

- Rust (for building the compiler)
- LLVM 17+
- Anthropic API key (for AI features)

## License

MIT

---

*Haira: Because your code should say what you mean.*
