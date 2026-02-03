# Haira Programming Language

**Write clear code. Use AI when you need it.**

Haira is a programming language for the AI era - explicit imports, explicit AI usage, and local-first AI that runs on your machine.

## Quick Example

```haira
import "http"
import "json"

User { name, email, active }

// Regular function - you write the logic
get_active_users(users: [User]) -> [User] {
    users | filter(u => u.active) | sort_by(u => u.name)
}

// AI function - AI generates the implementation
ai summarize_users(users: [User]) -> Summary {
    Count total users, active users, and list unique email domains.
}

server = http.Server { port = 8080 }

server.route("GET", "/users") {
    users = load_users()
    active = get_active_users(users)
    json.encode(active)
}

server.route("GET", "/summary") {
    users = load_users()
    summary = summarize_users(users)
    json.encode(summary)
}

server.start()
```

## Key Features

- **Explicit imports** - Go-style dependency declarations
- **Explicit AI blocks** - AI assistance when you request it
- **Local-first AI** - llama.cpp, Ollama, or cloud backends
- **No null** - Option types prevent null pointer errors
- **Type inference** - Types exist but you rarely write them
- **Fast binaries** - Compiles to native code via Cranelift
- **Reproducible** - AI outputs are cached and locked

## How It Works

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  Your Code   │ --> │    Haira     │ --> │   Native     │
│              │     │   Compiler   │     │   Binary     │
└──────────────┘     └──────┬───────┘     └──────────────┘
                           │
                    ┌──────┴───────┐
                    │  Local AI    │
                    │  (llama.cpp  │
                    │  or Ollama)  │
                    └──────────────┘
```

1. You write explicit, clear code with `import` and `ai` keywords
2. The compiler resolves imports and processes AI blocks
3. Local AI (llama.cpp/Ollama) generates implementations for AI blocks
4. Generated code is cached for reproducibility
5. Everything compiles to a fast native binary

**No cloud dependency** - AI runs on your machine by default.

## Quick Start

```bash
# Build the compiler
cargo build

# Run a simple program
./target/debug/haira run examples/basics/hello.haira

# Build to native binary
./target/debug/haira build examples/basics/hello.haira -o hello
./hello
```

## AI Blocks

Use the `ai` keyword when you want AI to generate an implementation:

```haira
// AI generates the implementation based on your description
ai calculate_engagement(user: User) -> EngagementScore {
    Analyze user's activity over the last 30 days.
    Consider login frequency, feature usage, and interactions.
    Return a score from 0-100 with breakdown by category.
}

// Use it like a normal function
score = calculate_engagement(current_user)
print(score.total)
```

### Using with Ollama (Recommended)

```bash
# Install Ollama (https://ollama.ai)
ollama pull deepseek-coder-v2:16b

# Build with local AI
./target/debug/haira build examples/ai/ai_demo.haira --ollama

# Use a different model
./target/debug/haira build examples/ai/ai_demo.haira --ollama --ollama-model codellama:7b
```

### Recommended Models

| Model | Size | Use Case |
|-------|------|----------|
| `deepseek-coder-v2:16b` | 9GB | Best quality (default) |
| `deepseek-coder:6.7b` | 4GB | Good balance |
| `codellama:7b` | 4GB | Fast alternative |

## Examples

```bash
# Basic examples (no AI required)
./target/debug/haira run examples/basics/hello.haira
./target/debug/haira run examples/functions/pipe.haira
./target/debug/haira run examples/functions/fibonacci.haira

# AI examples (requires Ollama)
./target/debug/haira build --ollama examples/ai/ai_demo.haira -o ai_demo
./ai_demo
```

## Project Structure

```
haira/
├── spec/                    # Language specification (15 chapters)
├── examples/                # Example programs
│   ├── basics/
│   ├── functions/
│   ├── control-flow/
│   ├── data-types/
│   ├── advanced/
│   └── ai/
├── crates/                  # Compiler implementation (Rust)
│   ├── haira-lexer/
│   ├── haira-parser/
│   ├── haira-ai/
│   ├── haira-codegen/
│   └── ...
└── Cargo.toml
```

## Philosophy

| Approach | Haira | Magic Languages |
|----------|-------|-----------------|
| Dependencies | Explicit imports | Auto-resolved |
| AI usage | Explicit `ai` blocks | Implicit generation |
| AI backend | Local-first (your machine) | Cloud-dependent |
| Code generation | Visible, cached | Hidden, unpredictable |

## Status

**Early Development** - Core compiler working, AI integration functional.

### Working Features

- Lexer & Parser
- Type System (structs, functions, arrays, options)
- Control Flow (if/else, for, while, match)
- Functions (closures, methods, pipe operator)
- Native Codegen (Cranelift)
- AI Intent Blocks (with Ollama)

## Requirements

- Rust (for building the compiler)
- Ollama (for AI features) - optional for non-AI code

## License

Apache-2.0
