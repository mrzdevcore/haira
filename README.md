<p align="center">
  <img src="assets/banner.svg" alt="Haira" width="600">
</p>

<p align="center">
  <strong>The programming language for AI agents and workflows.</strong>
</p>

<p align="center">
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-Apache%202.0-blue.svg" alt="License"></a>
  <img src="https://img.shields.io/badge/go-1.22+-00ADD8.svg" alt="Go 1.22+">
  <img src="https://img.shields.io/badge/examples-21-2962FF.svg" alt="21 examples">
</p>

---

Haira is a compiled language designed from the ground up for building agentic applications. Providers, tools, agents, and workflows are part of the language itself — not frameworks bolted on top. Write your agent logic, compile it to a native binary, and ship it.

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

@post("/api/chat")
workflow Chat(message: string, session_id: string) -> { reply: string } {
    reply, err = Assistant.ask(message, session: session_id)
    if err != nil { return { reply: "Something went wrong." } }
    return { reply: reply }
}

fn main() {
    server = http.Server([Chat])
    io.println("Server running on :8080")
    io.println("UI: http://localhost:8080/_ui/")
    server.listen(8080)
}
```

## Why Haira?

| What you replace | With Haira |
|------------------|------------|
| Python + LangChain/LangGraph | `agent` + `tool` keywords |
| n8n / Make / Zapier | `workflow` with `@post`, `@get` triggers + auto UI |
| CrewAI / AutoGen | Multi-agent with `handoffs` and `spawn` |
| Custom chatbot backend | Agent `memory` + `-> stream` + built-in chat UI |
| YAML/JSON config files | `provider` keyword — config in code |

## Key Features

- **4 agentic keywords** — `provider`, `tool`, `agent`, `workflow`
- **Compiles to native binaries** — via Go codegen, single executable output
- **Auto UI** — every workflow gets a form UI at `/_ui/`, streaming workflows get a ChatGPT-style chat UI
- **RESTful triggers** — `@get`, `@post`, `@put`, `@delete` decorators
- **Streaming** — `-> stream` workflows served as SSE
- **Agent handoffs** — agents delegate to other agents automatically
- **Agent memory** — `conversation(max_turns: N)` per session
- **File uploads** — `file` type with multipart handling, auto file picker in UI
- **Workflow steps** — named steps with telemetry, `@retry`, lifecycle hooks (`onerror`, `onsuccess`)
- **Parallel execution** — `spawn { }` blocks for concurrent agent calls
- **Pipe operator** — `data |> transform |> output`
- **Go-style simplicity** — familiar syntax, explicit error handling

## The Four Primitives

### Provider — LLM backend configuration

```haira
provider openai {
    api_key: env("OPENAI_API_KEY")
    model: "gpt-4o"
}

// Azure OpenAI
provider azure {
    api_key: env("AZURE_OPENAI_API_KEY")
    endpoint: env("AZURE_OPENAI_ENDPOINT")
    model: env("AZURE_OPENAI_DEPLOYMENT_NAME")
    api_version: "2025-01-01-preview"
}

// Local models via Ollama
provider local {
    endpoint: "http://localhost:11434/v1"
    model: "llama3"
}
```

Any OpenAI-compatible API works — set `endpoint` and `model`.

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
reply, err = SupportBot.ask("How do I reset my password?")
result, err = SupportBot.run("Help with billing")
return SupportBot.stream(message, session: session_id)
```

### Workflow — function with a trigger

```haira
@post("/api/support")
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
@post("/api/stream")
workflow Stream(message: string, session_id: string) -> stream {
    return Assistant.stream(message, session: session_id)
}
```

Clients requesting `Accept: text/event-stream` get SSE chunks. Others get a JSON response. Streaming workflows automatically get a ChatGPT-style chat UI at `/_ui/`.

## Workflow Steps & Lifecycle Hooks

```haira
@webui(title: "File Summarizer", description: "Upload a text file and get an AI summary")
@post("/api/summarize")
workflow Summarize(document: file, context: string) -> { summary: string } {
    onerror err {
        io.eprintln("Workflow failed: ${err}")
        return { summary: "Error: ${err}" }
    }

    step "Read file" {
        content, read_err = io.read_file(document)
        if read_err != nil { return { summary: "Failed to read file." } }
    }

    step "Summarize" {
        reply, err = Summarizer.ask(content)
        if err != nil { return { summary: "AI error." } }
    }

    return { summary: reply }
}
```

Steps provide named telemetry. `@retry` adds automatic retry with backoff:

```haira
@retry(max: 10, delay: 5000, backoff: "exponential")
step "Call external API" {
    result = http.get(url)
}
```

## Auto UI

Every workflow automatically gets a web UI — zero configuration:

- **`/_ui/`** — index page listing all workflows
- **`/_ui/<path>`** — form UI for regular workflows, chat UI for streaming workflows
- **`@webui(title: "...", description: "...")`** — optional UI customization
- **`file` params** — automatically render as file pickers with multipart upload
- **`HAIRA_DISABLE_UI=true`** — disable all auto-UIs for production

## Multi-Agent with Parallel Execution

```haira
@post("/api/analyze")
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
│       ├── checker/          # Type checker + semantic analysis
│       ├── codegen/          # Go code generation
│       ├── lsp/              # Language server protocol
│       └── driver/           # Pipeline orchestrator
├── go-runtime/              # Runtime library (Go)
│   └── haira/
│       ├── agent.go          # Agent execution, streaming, handoffs
│       ├── provider.go       # LLM provider config
│       ├── tool.go           # Tool registry
│       ├── workflow.go       # Workflow definitions
│       ├── server.go         # HTTP server with SSE + auto UI routing
│       ├── memory.go         # Session memory store
│       ├── upload.go         # File upload handling
│       ├── ui_form.go        # Auto form UI
│       ├── ui_chat.go        # Auto chat UI
│       └── ui/               # Embedded HTML templates
├── examples/                # 21 example programs
├── poc/                     # Real-world proof of concept
├── spec/                    # Language specification (17 chapters, LaTeX)
├── editors/                 # Editor extensions (Zed)
├── tree-sitter-haira/       # Tree-sitter grammar
└── Makefile
```

## Examples

All 21 examples compile and run:

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
| 20-stdlib | Standard library showcase |
| 21-file-upload | File upload with AI summarization |

## License

Apache-2.0
