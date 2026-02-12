# Haira Programming Language

**Build agents and workflows, not boilerplate.**

Haira is a statically-typed, compiled programming language where workflows, agents, and tools are first-class citizens — not library imports. It replaces the entire agentic stack: Python + LangChain, n8n/Make/Zapier, CrewAI/AutoGen — all in one language with native binaries.

## Quick Example

```haira
import "io"

provider anthropic {
    api_key: env("ANTHROPIC_API_KEY")
    model: "claude-sonnet-4-20250514"
}

tool search(query: string) -> [Result] {
    """Search the web for information"""
    resp, err = http.get("https://api.search.com?q={query}")
    if err != nil { return [], err }
    return json.decode(resp.body, [Result])
}

agent Researcher {
    model: anthropic
    system: "You are a thorough researcher. Always cite sources."
    tools: [search]
    temperature: 0.2
}

@webhook("/research")
workflow Research(topic: string) -> ResearchResult {
    result: ResearchResult, err = Researcher.run(ResearchRequest{ topic: topic })
    if err != nil { return ResearchResult{}, err }
    return result
}

fn main() {
    server = haira.serve([Research])
    io.println("Running on :8080")
    server.listen(8080)
}
```

## Why Haira?

| What you replace | With Haira |
|------------------|------------|
| Python + LangChain/LangGraph | `agent` + `tool` keywords |
| n8n / Make / Zapier | `workflow` with `@webhook`, `@cron`, `@event` triggers |
| CrewAI / AutoGen | Multi-agent composition with `spawn` blocks |
| Custom chatbot backend | `@websocket` + agent `memory` + `stream<T>` |
| YAML/JSON config files | `provider` keyword — type-checked config in code |

## Key Features

- **4 new keywords** — `provider`, `tool`, `agent`, `workflow` — that's it
- **Type-safe orchestration** — agent inputs, outputs, and tool schemas verified at compile time
- **Native binaries** — compiles to standalone executables via Cranelift
- **Streaming** — `stream<T>` type for token-by-token agent output
- **Built-in triggers** — `@webhook`, `@websocket`, `@cron`, `@event`, `@manual`
- **Agent memory** — `conversation(max_turns: N)` and `summary(max_tokens: N)`
- **Parallel execution** — `spawn { }` blocks for concurrent agent calls
- **No null** — option types (`T?`) prevent null pointer errors
- **Go-style simplicity** — familiar syntax, explicit error handling

## The Four Primitives

### Provider — LLM backend configuration

```haira
provider anthropic {
    api_key: env("ANTHROPIC_API_KEY")
    model: "claude-sonnet-4-20250514"
}
```

### Tool — typed function with LLM-visible description

```haira
tool search_kb(query: string) -> [Article] {
    """Search the knowledge base for relevant articles"""
    results, err = pg.query("SELECT * FROM articles WHERE content @@ to_tsquery($1)", [query])
    if err != nil { return [], err }
    return results
}
```

### Agent — LLM entity with model, prompt, and tools

```haira
agent SupportBot {
    model: anthropic
    system: "You are a helpful customer support agent."
    tools: [search_kb, lookup_order, create_ticket]
    memory: conversation(max_turns: 100)
    temperature: 0.3
}
```

Three ways to use an agent:

```haira
// Text in, text out
answer, err = SupportBot.ask("How do I reset my password?")

// Structured in, structured out
result: TicketResult, err = SupportBot.run(TicketRequest{ issue: "billing" })

// Streaming (token by token)
for chunk in SupportBot.stream("Explain our pricing") {
    io.print(chunk)
}
```

### Workflow — function with a trigger

```haira
@webhook("/support")
workflow CustomerSupport(message: string, session_id: string) -> { reply: string } {
    reply, err = SupportBot.ask(message, session: session_id)
    return { reply: reply or "Something went wrong." }
}
```

## Streaming Chatbot (7 lines)

```haira
agent Assistant {
    model: anthropic
    system: "You are a helpful assistant."
    memory: conversation(max_turns: 50)
}

@websocket("/chat")
workflow Chat(message: string, session_id: string) -> stream<string> {
    return Assistant.stream(message, session: session_id)
}
```

## Multi-Agent Workflow

```haira
@manual
workflow ArticlePipeline(topic: string) -> ArticleResult {
    // Step 1: Research
    research: ResearchResult, err = Researcher.run(ResearchRequest{ topic: topic })
    if err != nil { return ArticleResult{}, err }

    // Step 2: Write draft
    draft: Article, err = Writer.run(WriteRequest{
        content: research.summary,
        sources: research.sources
    })
    if err != nil { return ArticleResult{}, err }

    // Step 3: Parallel reviews
    reviews = spawn {
        Reviewer.run(ReviewRequest{ article: draft, focus: "accuracy" })
        Reviewer.run(ReviewRequest{ article: draft, focus: "clarity" })
        Reviewer.run(ReviewRequest{ article: draft, focus: "engagement" })
    }

    return ArticleResult{ article: draft, reviews: reviews }
}
```

## Running

```bash
# Compile and run
haira run main.haira

# Start server (for webhook/websocket/cron workflows)
haira serve main.haira --port 8080

# Build native binary
haira build main.haira -o myapp

# Type-check only
haira check main.haira

# CLI invocation of a @manual workflow
haira run ArticlePipeline --input '{"topic": "AI agents"}'
```

## Project Structure

```
haira/
├── spec/                    # Language specification (17 chapters)
├── examples/                # Example programs
│   ├── basics/
│   ├── functions/
│   ├── control-flow/
│   ├── data-types/
│   ├── advanced/
│   └── agentic/
├── crates/                  # Compiler implementation (Rust)
│   ├── haira-lexer/         # Lexical analysis
│   ├── haira-parser/        # Parsing
│   ├── haira-ast/           # AST definitions
│   ├── haira-resolver/      # Name resolution
│   ├── haira-types/         # Type system
│   ├── haira-hir/           # High-level IR
│   ├── haira-codegen/       # Code generation (Cranelift)
│   ├── haira-runtime/       # Runtime (agents, workflows, providers)
│   ├── haira-driver/        # Compilation driver
│   ├── haira-cli/           # CLI
│   └── haira-lsp/           # Language server
└── Cargo.toml
```

## Status

**Early Development** — Language specification complete. Compiler core working (lexer, parser, type system, codegen). Agentic runtime in progress.

### Working

- Lexer & Parser
- Type system (structs, enums, generics, options)
- Control flow (if/else, for, while, match)
- Functions (closures, methods, pipe operator)
- Concurrency (spawn, chan, select)
- Native codegen (Cranelift)

### In Progress

- Provider, tool, agent, workflow parsing
- Agentic runtime (LLM API calls, tool execution, memory)
- Workflow triggers (webhook, websocket, cron, event)
- Stream type implementation

## Requirements

- Rust (for building the compiler)

## License

Apache-2.0
