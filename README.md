# Haira

**The programming language for AI agents.**

Haira is a compiled language designed from the ground up for building agentic applications. Providers, tools, agents, and workflows are part of the language itself — not frameworks bolted on top. Write your agent logic, compile it to a native binary, and ship it.

## Quick Example

```haira
import "io"
import "http"

provider openai {
    api_key: env("OPENAI_API_KEY")
    model: "gpt-4o"
}

tool get_weather(city: string) -> string {
    """Get the current weather for a given city"""
    resp, err = http.get("https://wttr.in/${city}?format=j1")
    if err != nil { return "Failed to fetch weather data." }
    data = resp.json()
    current = data["current_condition"][0]
    return "${city}: ${current["temp_C"]}C"
}

agent Assistant {
    model: openai
    system: "You are a helpful assistant. Be concise."
    tools: [get_weather]
    memory: conversation(max_turns: 10)
    temperature: 0.7
}

@webhook("/api/chat")
workflow Chat(message: string, session_id: string) -> { reply: string } {
    reply, err = Assistant.ask(message, session: session_id)
    if err != nil { return { reply: "Something went wrong." } }
    return { reply: reply }
}

fn main() {
    server = http.Server([Chat])
    io.println("Server running on :8080")
    server.listen(8080)
}
```

## Why Haira?

| What you replace | With Haira |
|------------------|------------|
| Python + LangChain/LangGraph | `agent` + `tool` keywords |
| n8n / Make / Zapier | `workflow` with `@webhook` triggers |
| CrewAI / AutoGen | Multi-agent with `handoffs` and `spawn` |
| Custom chatbot backend | Agent `memory` + `-> stream` workflows |
| YAML/JSON config files | `provider` keyword — config in code |

## Key Features

- **4 new keywords** — `provider`, `tool`, `agent`, `workflow` — that's it
- **Compiles to native binaries** — via Go codegen, single executable output
- **Streaming** — `-> stream` workflows served as SSE
- **Agent handoffs** — agents delegate to other agents automatically
- **Built-in triggers** — `@webhook` to expose workflows as HTTP endpoints
- **Agent memory** — `conversation(max_turns: N)` and `summary(max_tokens: N)`
- **Parallel execution** — `spawn { }` blocks for concurrent agent calls
- **Go-style simplicity** — familiar syntax, explicit error handling

## The Four Primitives

### Provider — LLM backend configuration

```haira
provider openai {
    api_key: env("OPENAI_API_KEY")
    model: "gpt-4o"
}
```

### Tool — function with LLM-visible description

```haira
tool search_kb(query: string) -> string {
    """Search the knowledge base for relevant articles"""
    resp, err = http.get("https://api.example.com/search?q=${query}")
    if err != nil { return "Search failed." }
    return resp.body
}
```

### Agent — LLM entity with model, prompt, and tools

```haira
agent SupportBot {
    model: openai
    system: "You are a helpful customer support agent."
    tools: [search_kb]
    memory: conversation(max_turns: 20)
    temperature: 0.3
}
```

Three ways to call an agent:

```haira
// Text in, text out
reply, err = SupportBot.ask("How do I reset my password?")

// Full result with metadata
result, err = SupportBot.run("Help with billing")

// Streaming (token by token, served as SSE)
return SupportBot.stream(message, session: session_id)
```

### Workflow — function with a trigger

```haira
@webhook("/api/support")
workflow Support(message: string, session_id: string) -> { reply: string } {
    reply, err = SupportBot.ask(message, session: session_id)
    if err != nil { return { reply: "Something went wrong." } }
    return { reply: reply }
}
```

## Agent Handoffs

Agents can delegate to specialized agents automatically:

```haira
agent FrontDesk {
    model: openai
    system: "Greet users. Hand off billing questions to BillingAgent."
    handoffs: [BillingAgent, TechAgent]
    memory: conversation(max_turns: 10)
}

agent BillingAgent {
    model: openai
    system: "You handle billing and payment questions."
}

agent TechAgent {
    model: openai
    system: "You handle technical support questions."
}
```

## Streaming

```haira
@webhook("/api/stream")
workflow Stream(message: string, session_id: string) -> stream {
    return Assistant.stream(message, session: session_id)
}
```

Clients requesting `Accept: text/event-stream` get SSE chunks. Others get a JSON response.

## Multi-Agent with Parallel Execution

```haira
@webhook("/api/analyze")
workflow Analyze(topic: string) -> { results: [string] } {
    results = spawn {
        Researcher.ask("Find facts about ${topic}")
        Critic.ask("Find counterarguments about ${topic}")
        Summarizer.ask("Write a summary about ${topic}")
    }
    return { results: results }
}
```

## Getting Started

### Requirements

- Go 1.22+

### Build

```bash
make build
```

### Run

```bash
# Compile and run
./compiler/haira run examples/01-hello.haira

# Build a native binary
./compiler/haira build examples/07-agentic.haira -o myapp

# Show generated Go code
./compiler/haira emit examples/07-agentic.haira

# Parse and show AST
./compiler/haira parse examples/01-hello.haira

# Type-check only
./compiler/haira check examples/01-hello.haira
```

### Install

```bash
make install-local    # installs to ~/.local/bin/haira
```

## Project Structure

```
haira/
├── compiler/                # Compiler (Go)
│   ├── main.go              # CLI entry point
│   └── internal/
│       ├── token/            # Token types
│       ├── lexer/            # Hand-written scanner
│       ├── ast/              # AST node types
│       ├── parser/           # Recursive descent + Pratt parsing
│       ├── codegen/          # Go code generation
│       └── driver/           # Pipeline orchestrator
├── go-runtime/              # Runtime library (Go)
│   └── haira/
│       ├── agent.go          # Agent execution, streaming, handoffs
│       ├── provider.go       # LLM provider config
│       ├── tool.go           # Tool registry
│       ├── workflow.go       # Workflow definitions
│       ├── server.go         # HTTP server with SSE support
│       ├── memory.go         # Session memory
│       ├── http.go           # HTTP client helpers
│       └── io.go             # Print helpers
├── examples/                # 19 example programs
│   ├── 01-hello.haira       # Hello world
│   ├── 07-agentic.haira     # Agent with tools
│   ├── 14-multi-agent.haira  # Multiple agents
│   ├── 15-handoffs.haira    # Agent handoffs
│   ├── 19-streaming.haira   # SSE streaming
│   └── ...
├── spec/                    # Language specification (17 chapters, LaTeX)
└── Makefile
```

## Examples

All 19 examples compile and run:

```bash
make build-examples    # compile all
make run-examples      # run non-agentic examples
```

| Example | Description |
|---------|-------------|
| 01-hello | Hello world |
| 02-variables | Variable declarations |
| 03-functions | Functions, closures |
| 04-control-flow | If/else, for, while |
| 05-match | Pattern matching |
| 06-lists | List operations |
| 07-agentic | Agent with tools and webhook |
| 08-structs | Struct types |
| 09-string-interpolation | `${expr}` interpolation |
| 10-maps | Map operations |
| 11-pipes | Pipe operator |
| 12-methods | Methods on types |
| 13-error-handling | Try/catch, error propagation |
| 14-multi-agent | Multiple agents and providers |
| 15-handoffs | Agent-to-agent handoffs |
| 16-enums | Enum types |
| 17-compound-assign | `+=`, `-=`, etc. |
| 18-defer | Defer statements |
| 19-streaming | SSE streaming workflow |

## License

Apache-2.0
