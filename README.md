# Haira Programming Language

**Express intention, not mechanics.**

Haira is a programming language where you write what you want, and the compiler (powered by a fine-tuned AI model) figures out how to do it.

## Quick Example

```haira
// Define what you want - AI generates the implementation
ai get_answer() -> int {
    Return the answer to life, universe, and everything.
}

ai add(a: int, b: int) -> int {
    Return the sum of a and b.
}

ai is_positive(x: int) -> bool {
    Return true if x is greater than zero.
}

// Use them like normal functions
answer = get_answer()   // Returns 42
sum = add(10, 32)       // Returns 42
check = is_positive(5)  // Returns true

print(answer)
print(sum)
print(check)
```

**The `ai` block lets you describe your intent in plain English** - the compiler interprets it and generates working code that compiles to a native binary.

## How It Works

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  Your Code   │ --> │    Haira     │ --> │   Native     │
│  (Intent)    │     │   Compiler   │     │   Binary     │
└──────────────┘     └──────┬───────┘     └──────────────┘
                           │
                    ┌──────┴───────┐
                    │  Fine-tuned  │
                    │   AI Model   │
                    │  (Interprets │
                    │    Intent)   │
                    └──────────────┘
```

1. You write high-level, intent-based code
2. The compiler identifies undefined functions
3. A fine-tuned AI model interprets your intent from function names and context
4. Generated code is cached for reproducibility
5. Everything compiles to a fast native binary

## Key Features

- **No imports** - The compiler finds everything automatically
- **No null** - Option types prevent null pointer errors
- **No boilerplate** - AI generates CRUD, transformations, I/O
- **Type inference** - Types exist but you rarely write them
- **Fast binaries** - Compiles to native code via Cranelift
- **Reproducible** - AI outputs are cached and locked
- **Local-first** - Run AI models locally with Ollama, no API keys required

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
│   ├── haira-ai/           # AI intent engine
│   ├── haira-cir/          # Canonical IR
│   ├── haira-types/        # Type system
│   ├── haira-hir/          # High-level IR
│   ├── haira-mir/          # Mid-level IR
│   ├── haira-codegen/      # Cranelift code generation
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

🚧 **Early Development** - Core compiler working, AI integration functional.

### What's Working

- **Lexer & Parser**: Full tokenization and AST generation
- **Type System**: Structs, functions, primitives, arrays, options
- **Control Flow**: if/else, for loops, while loops, match expressions
- **Functions**: Definition, calls, closures, methods
- **Native Codegen**: Compiles to native binaries via Cranelift
- **AI Intent Blocks**: Explicit AI-powered function generation

## Quick Start

```bash
# Build the compiler
cargo build

# Run a simple program
./target/debug/haira run examples/hello.haira

# Build to native binary
./target/debug/haira build examples/hello.haira -o hello
./hello
```

## AI Intent Blocks

Haira supports explicit AI-powered function generation using the `ai` block syntax:

```haira
// AI generates the implementation based on your intent
ai get_answer() -> int {
    Return the answer to life, universe, and everything.
}

ai add(a: int, b: int) -> int {
    Return the sum of a and b.
}

ai factorial(n: int) -> int {
    Return the factorial of n.
}

// Use them like normal functions
answer = get_answer()  // Returns 42
sum = add(10, 32)      // Returns 42
f = factorial(5)       // Returns 120
```

### Using with Ollama (Recommended)

Run AI interpretation locally with Ollama - no API keys required:

```bash
# Install Ollama (https://ollama.ai)
# Then pull a coding model
ollama pull deepseek-coder-v2:16b

# Build with local AI
./target/debug/haira build examples/ai_minimal.haira --ollama

# Use a different model
./target/debug/haira build examples/ai_minimal.haira --ollama --ollama-model codellama:7b
```

Recommended models for code generation:
- `deepseek-coder-v2:16b` (default) - Best quality for complex logic
- `deepseek-coder:6.7b` - Good balance of speed and quality
- `codellama:7b` - Fast alternative
- `qwen2.5-coder:7b` - Strong reasoning capabilities

## Requirements

- Rust (for building the compiler)
- Ollama running locally (for AI features)

## License

Apache-2.0

---

*Haira: Because your code should say what you mean.*
