<p align="center">
  <img src="assets/banner.svg" alt="Haira" width="600">
</p>

<p align="center">
  <strong>The programming language for AI agents and workflows.</strong>
</p>

<p align="center">
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-Apache%202.0-blue.svg" alt="License"></a>
  <img src="https://img.shields.io/badge/go-1.22+-00ADD8.svg" alt="Go 1.22+">
  <img src="https://img.shields.io/badge/examples-29-2962FF.svg" alt="29 examples">
</p>

<p align="center">
  <a href="https://haira.dev">Website</a> &middot;
  <a href="https://haira.dev/docs/getting-started/installation">Documentation</a> &middot;
  <a href="https://haira.dev/docs/examples">Examples</a> &middot;
  <a href="https://haira.dev/agentic-rendering-protocol">ARP Protocol</a> &middot;
  <a href="https://haira.dev/generative-ui">Generative UI</a>
</p>

---

> **Note:** Haira is under heavy development and not yet production-ready. APIs and syntax may change. Use at your own risk.

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

## Architecture

```
                              ┌─────────────────────────────────────┐
                              │           .haira source             │
                              └──────────────┬──────────────────────┘
                                             │
                    ┌────────────────────────────────────────────────┐
                    │                  COMPILER                      │
                    │                                                │
                    │   ┌───────┐   ┌────────┐   ┌──────────┐        │
                    │   │ Lexer │──▶│ Parser │──▶│ Checker  │        │
                    │   └───────┘   └────────┘   └─────┬────┘        │
                    │                                  │             │
                    │                            ┌─────▼──────┐      │
                    │                            │  Codegen   │      │
                    │                            │  (Go emit) │      │
                    │                            └─────┬──────┘      │
                    └──────────────────────────────────┼─────────────┘
                                                       │
                                                  go build
                                                       │
                                                       ▼
                    ┌─────────────────────────────────────────────────┐
                    │              NATIVE BINARY                      │
                    │                                                 │
                    │  ┌─────────────────────────────────────────┐    │
                    │  │            Haira Runtime                │    │
                    │  │                                         │    │
                    │  │  ┌──────────┐  ┌───────┐  ┌──────────┐  │    │
                    │  │  │ Provider │  │ Agent │  │ Workflow │  │    │
                    │  │  └──────────┘  └───┬───┘  └────┬─────┘  │    │
                    │  │                    │           │        │    │
                    │  │         ┌──────────┴───────────┘        │    │
                    │  │         │                               │    │
                    │  │  ┌──────▼──────┐     ┌──────────────┐   │    │
                    │  │  │ HTTP Server │     │  MCP Server  │   │    │
                    │  │  │  REST + SSE │     │ stdio / HTTP │   │    │
                    │  │  └──────┬──────┘     └──────────────┘   │    │
                    │  │         │                               │    │
                    │  │  ┌──────▼──────┐     ┌──────────────┐   │    │
                    │  │  │  ARP Bridge │     │  Observe /   │   │    │
                    │  │  │ (protocol)  │     │  Langfuse    │   │    │
                    │  │  └──┬──────┬───┘     └──────────────┘   │    │
                    │  │     │      │                            │    │
                    │  └─────┼──────┼────────────────────────────┘    │
                    │        │      │                                 │
                    │   ┌────▼──┐ ┌─▼────────┐   ┌───────────────┐    │
                    │   │  SSE  │ │WebSocket │   │  SQLite Store │    │
                    │   │(http) │ │(_arp/v1) │   │  (sessions)   │    │
                    │   └───┬───┘ └────┬─────┘   └───────────────┘    │
                    └───────┼──────────┼──────────────────────────────┘
                            │          │
                            ▼          ▼
                    ┌──────────────────────────────────┐
                    │          UI SDK (Lit)            │
                    │                                  │
                    │  ┌──────┐ ┌──────┐ ┌──────────┐  │
                    │  │ Chat │ │ Form │ │Generative│  │
                    │  │  UI  │ │  UI  │ │UI Comps  │  │
                    │  └──────┘ └──────┘ └──────────┘  │
                    │                                  │
                    │  tables, charts, status cards,   │
                    │  code blocks, diffs, key-value,  │
                    │  confirm, choices, forms,        │
                    │  product cards, progress views   │
                    └──────────────────────────────────┘
```

**Data flow:** `.haira` source is compiled through Lexer → Parser → Checker → Go Codegen, then `go build` produces a single native binary. At runtime, the binary embeds the full Haira runtime (agents, providers, tools, workflows, HTTP server, ARP protocol bridge, UI SDK) — zero external dependencies.

## Why Haira?

| What you replace | With Haira |
|------------------|------------|
| Python + LangChain/LangGraph | `agent` + `tool` keywords |
| n8n / Make / Zapier | `workflow` with `@post`, `@get` triggers + auto UI |
| CrewAI / AutoGen | Multi-agent with `handoffs` and `spawn` |
| Custom chatbot backend | Agent `memory` + `-> stream` + built-in chat UI |
| YAML/JSON config files | `provider` keyword — config in code |
| MCP glue code | `mcp.Server()` / `provider { transport: "mcp" }` |
| Vercel AI SDK + React UI | Generative UI with `ui.*` components |

## Key Features

- **4 agentic keywords** — `provider`, `tool`, `agent`, `workflow`
- **Compiles to native binaries** — via Go codegen, single executable output
- **Generative UI** — agents render rich UI components (tables, charts, status cards, forms) via `ui.*` helpers
- **ARP (Agentic Rendering Protocol)** — transport-agnostic protocol for agent-to-renderer communication (WebSocket + SSE)
- **Auto UI** — every workflow gets a form UI at `/_ui/`, streaming workflows get a ChatGPT-style chat UI
- **RESTful triggers** — `@get`, `@post`, `@put`, `@delete` decorators
- **Streaming** — `-> stream` workflows served as SSE with WebSocket upgrade
- **Agent handoffs** — agents delegate to other agents automatically
- **Agent memory** — `conversation(max_turns: N)` per session
- **File uploads** — `file` type with multipart handling, auto file picker in UI
- **Workflow steps** — named steps with telemetry, `@retry`, lifecycle hooks (`onerror`, `onsuccess`)
- **Parallel execution** — `spawn { }` blocks for concurrent agent calls
- **Pipe operator** — `data |> transform |> output`
- **MCP support** — consume external tools (`provider { transport: "mcp" }`) and expose workflows as MCP tools (`mcp.Server()`)
- **Observability** — built-in `observe` module with Langfuse integration
- **10 stdlib packages** — postgres, sqlite, excel, vector, slack, github, gitlab, algolia, meilisearch, langfuse
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

## Generative UI

Agents can render rich UI components directly into the chat. Tools return `ui.*` helpers that display tables, charts, status cards, and more — no frontend code required:

```haira
tool query_data(sql: string) -> string {
    """Execute a SQL query and display results as a table"""
    rows, err = db.query(sql)
    if err != nil {
        return ui.status_card("error", "Query Failed", conv.to_string(err))
    }
    headers = keys(rows[0])
    table_rows = []
    for row in rows {
        cells = []
        for h in headers {
            cells = array.push(cells, conv.to_string(row[h]))
        }
        table_rows = array.push(table_rows, cells)
    }
    return ui.table("Results", headers, table_rows)
}

tool visualize(chart_type: string, title: string, labels: string, datasets: string) -> string {
    """Create a chart visualization"""
    return ui.chart(chart_type, title, json.parse(labels), json.parse(datasets))
}
```

Available UI components:

| Component | Helper | Description |
|-----------|--------|-------------|
| Status Card | `ui.status_card(status, title, message?)` | Success/error/warning/info indicator |
| Table | `ui.table(title, headers, rows)` | Searchable data table |
| Chart | `ui.chart(type, title, labels, datasets)` | Line, bar, pie, scatter, area charts |
| Key-Value | `ui.key_value(title, items)` | Labeled property list |
| Code Block | `ui.code_block(title, language, code)` | Syntax-highlighted code |
| Diff | `ui.diff(title, before, after)` | Before/after comparison |
| Progress | `ui.progress(title, steps)` | Multi-step progress tracker |
| Form | `ui.form(title, fields)` | Interactive form input |
| Confirm | `ui.confirm(title, message?)` | Yes/no confirmation dialog |
| Choices | `ui.choices(title, options)` | Option picker (buttons/list) |
| Product Cards | `ui.product_cards(title, cards)` | Product card grid with images |
| Group | `ui.group(child1, child2, ...)` | Compose multiple components |

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

Streaming workflows support two transports:
- **SSE** — clients requesting `Accept: text/event-stream` get SSE chunks
- **WebSocket** — clients connect to `/_arp/v1` for bidirectional ARP communication

Both transports deliver the same data. The built-in chat UI automatically upgrades to WebSocket when available, falling back to SSE.

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

## MCP (Model Context Protocol)

Haira has built-in MCP support in both directions — consume external tools and expose workflows as tools.

### MCP Client — Use External Tools

Connect to any MCP server. The agent discovers and uses its tools automatically:

```haira
import "http"

provider filesystem {
    transport: "mcp"
    command: "npx"
    args: ["-y", "@modelcontextprotocol/server-filesystem", "/tmp"]
}

agent Assistant {
    model: openai
    system: "You are a helpful assistant with file system access."
    mcp: [filesystem]
}
```

SSE transport works too — connect to remote MCP servers over HTTP:

```haira
provider remote_tools {
    transport: "mcp"
    endpoint: "http://tools-server:9000/sse"
}
```

### MCP Server — Expose Workflows as Tools

Any workflow can be exposed as an MCP tool for external agents (Claude Code, Cursor, other Haira agents):

```haira
import "mcp"

workflow Summarize(text: string) -> { summary: string } {
    """Summarize the given text into key points."""
    summary, err = Summarizer.ask(text)
    if err != nil { return { summary: "Error." } }
    return { summary: summary }
}

fn main() {
    mcp_server = mcp.Server([Summarize])
    mcp_server.listen(9000)  // SSE on http://localhost:9000/sse
}
```

Both transports are supported:
- `mcp_server.serve()` — stdio (for subprocess integration)
- `mcp_server.listen(9000)` — SSE over HTTP (for remote agents)

### Distributed Agent Network

Combine MCP client + server + handoffs for cross-machine agent orchestration:

```
Server A (Summarizer)  <--MCP-->  Server B (Translator)
       ^                                ^
       +-------- MCP ---- Server C (Orchestrator)
```

## Benchmarks

Measured on Apple Silicon (M-series). Competitor numbers from published benchmarks and framework documentation.

### Compiler Performance

| Phase | 29 examples | Per file |
|-------|-------------|----------|
| Lex | 85ms | ~2.9ms |
| Parse | 80ms | ~2.8ms |
| Codegen (emit Go) | 86ms | ~3.0ms |
| Full build (agentic) | 440ms | -- |

### Runtime Performance

| Metric | Haira | Python + LangGraph | Python + CrewAI | Node.js + Vercel AI SDK |
|--------|-------|--------------------|-----------------|-----------------------|
| **Startup time** | **18ms** | ~1000ms | ~700ms | ~200ms |
| **Memory (idle)** | **11 MB** | ~200 MB | ~150 MB | ~100 MB |
| **Binary / deploy size** | **11 MB** | ~500 MB+ (Docker) | ~400 MB+ (Docker) | ~300 MB+ (Docker) |
| **HTTP req/sec** | **~19,000** | ~1,000-3,000 | ~1,000-3,000 | ~6,000-8,000 |
| **Dependencies** | **0** | 50-200 packages | 50-150 packages | 100-300 packages |

### Usability: Lines of Code

| Task | Haira | LangGraph | CrewAI | Vercel AI SDK |
|------|-------|-----------|--------|---------------|
| Agent + tool + HTTP server | **47 lines, 1 file** | ~130 lines, 3-5 files | ~100 lines, 2-4 files | ~90 lines, 3-4 files |
| Multi-agent handoffs | **48 lines** | ~200+ lines | ~120 lines | ~150+ lines |
| MCP client integration | **39 lines** | N/A | N/A | ~80 lines |
| MCP server (expose as tool) | **32 lines** | N/A | N/A | N/A |

### Feature Comparison

| Capability | Haira | LangGraph | CrewAI | AutoGen | Vercel AI SDK |
|------------|-------|-----------|--------|---------|---------------|
| Custom tools | First-class keyword | Decorator | Decorator/class | Function | Zod schema |
| Generative UI | Built-in (`ui.*`) | No | No | No | React components |
| Multi-agent | Handoffs (built-in) | Graph edges | Role delegation | Conversations | Manual |
| MCP client | Built-in | Via plugin | No | No | Plugin |
| MCP server | Built-in | No | No | No | No |
| HTTP server | Built-in | Manual (Flask) | No | No | Via Next.js |
| SSE streaming | `-> stream` keyword | Manual | No | No | Built-in |
| WebSocket (ARP) | Built-in | No | No | No | No |
| Memory/sessions | Language keyword | Checkpointer | Config | Config | Manual |
| Type safety | Compile-time | Runtime | Runtime | Runtime | TypeScript |
| Parallel execution | `spawn { }` | `Send()` API | Task config | Group chat | `Promise.all` |
| Auto UI | Built-in | No | No | No | No |
| Observability | Built-in + Langfuse | Via callbacks | Via callbacks | Via callbacks | Via callbacks |
| Deploy | Single binary | Docker + venv | Docker + venv | Docker + node_modules | Docker + node_modules |

## Getting Started

For full documentation, visit **[haira.dev/docs](https://haira.dev/docs/getting-started/installation)**.

### Quick Install

```bash
curl -fsSL https://haira.dev/install.sh | sh
```

### Build from Source

Requires Go 1.22+.

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

### Install Locally

```bash
make install-local    # installs to ~/.local/bin/haira
```

## Project Structure

```
haira/
├── compiler/                   # Compiler (Go)
│   ├── main.go                 # CLI: build, run, parse, check, lex, emit, lsp
│   └── internal/
│       ├── token/              # Token types
│       ├── lexer/              # Hand-written scanner
│       ├── ast/                # AST node types
│       ├── parser/             # Recursive descent + Pratt parsing
│       ├── checker/            # Type checker + semantic analysis
│       ├── resolver/           # Name resolution
│       ├── codegen/            # Go code generation
│       ├── errors/             # Diagnostic system
│       ├── lsp/                # Language server protocol
│       ├── driver/             # Pipeline orchestrator
│       └── runtime/            # Embedded UI bundle (bundle.tar.gz)
├── primitive/haira/            # Core runtime (Go)
│   ├── agent.go                # Agent execution, streaming, handoffs
│   ├── provider.go             # LLM provider config
│   ├── tool.go                 # Tool registry
│   ├── workflow.go             # Workflow definitions
│   ├── server.go               # HTTP server, SSE, auto UI routing
│   ├── arp.go                  # ARP protocol types + bridge
│   ├── arp_ws.go               # ARP WebSocket transport
│   ├── mcp_client.go           # MCP client (stdio + SSE)
│   ├── mcp_server.go           # MCP server (stdio + SSE)
│   ├── memory.go               # Session memory store
│   ├── store.go                # Session persistence interface
│   ├── observe.go              # Observability / telemetry
│   ├── upload.go               # File upload handling
│   └── ui_*.go                 # Generative UI components + tools
├── stdlib/                     # Standard library (tree-shaken)
│   ├── postgres/               # PostgreSQL client
│   ├── sqlite/                 # SQLite store backend
│   ├── excel/                  # Excel file generation
│   ├── vector/                 # Vector embeddings + search
│   ├── slack/                  # Slack integration
│   ├── github/                 # GitHub API client
│   ├── gitlab/                 # GitLab API client
│   ├── algolia/                # Algolia search client
│   ├── meilisearch/            # Meilisearch client
│   └── langfuse/               # Langfuse observability exporter
├── ui/sdk/                     # UI SDK (TypeScript, Lit web components)
│   └── src/
│       ├── core/               # Types, styles, ARP client
│       ├── components/         # Chat, form, generative UI components
│       ├── pages/              # App shell pages
│       └── services/           # SSE client
├── spec/                       # Language specification
│   ├── latex/                  # 18-chapter spec (LaTeX)
│   └── arp/                    # ARP protocol spec + component catalog
├── examples/                   # 29 example programs
├── poc/                        # Real-world proof of concepts
│   ├── data-explorer/          # AI-powered data querying + visualization
│   ├── devops-incident/        # DevOps incident management
├── editors/zed-haira/          # Zed editor extension
├── tree-sitter-haira/          # Tree-sitter grammar
└── Makefile
```

## Examples

All 29 examples compile and run:

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
| 22-pipeline-ui | Workflow steps with pipeline UI |
| 23-mcp | MCP client — agent with external tools |
| 24-mcp-server | MCP server — expose workflows as tools |
| 25-embeddings | Vector embeddings + similarity search |
| 26-rag | Retrieval-augmented generation |
| 27-structured-output | Typed agent output with structs |
| 28-observe | Observability with Langfuse |
| 29-testing | Testing workflows |

## Learn More

- **[Website](https://haira.dev)** — landing page and overview
- **[Documentation](https://haira.dev/docs/getting-started/installation)** — installation, language guide, agentic features, stdlib reference
- **[ARP Protocol](https://haira.dev/agentic-rendering-protocol)** — transport-agnostic protocol for agent-to-renderer communication
- **[Generative UI](https://haira.dev/generative-ui)** — agents that render rich, interactive components

## License

Apache-2.0
