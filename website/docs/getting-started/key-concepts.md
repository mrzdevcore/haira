# Key Concepts

Haira is built around **four agentic keywords** that are part of the language itself — not a library.

## The Four Primitives

### Provider

A provider configures an LLM backend. It specifies API credentials, the model, and optional parameters.

```haira
provider openai {
    api_key: env("OPENAI_API_KEY")
    model: "gpt-4o"
}

provider anthropic {
    api_key: env("ANTHROPIC_API_KEY")
    model: "claude-sonnet-4-5-20250929"
}
```

Supported providers: OpenAI, Azure OpenAI, Anthropic.

### Tool

A tool is a function that agents can call. Tools must have a `"""docstring"""` — this is sent to the LLM so it knows when and how to use the tool.

```haira
tool get_weather(city: string) -> string {
    """Get the current weather for a given city"""
    resp, err = http.get("https://wttr.in/${city}?format=j1")
    if err != nil {
        return "Failed to fetch weather data."
    }
    data = resp.json()
    return "${city}: ${data["current_condition"][0]["temp_C"]}°C"
}
```

### Agent

An agent is an LLM entity with a provider, system prompt, optional tools, and memory.

```haira
agent Assistant {
    provider: openai
    system: "You are a helpful assistant."
    tools: [get_weather]
    memory: conversation(max_turns: 10)
    temperature: 0.7
}
```

Agents expose three methods:
- `agent.ask(msg)` — send a message, get a response
- `agent.run(msg)` — like ask, but returns `AgentResult` for manual control
- `agent.stream(msg)` — stream the response as SSE

### Workflow

A workflow is a function decorated with an HTTP trigger. It defines your API endpoints.

```haira
@post("/api/chat")
workflow Chat(message: string, session_id: string) -> { reply: string } {
    reply, err = Assistant.ask(message, session: session_id)
    if err != nil {
        return { reply: "Something went wrong." }
    }
    return { reply: reply }
}
```

Workflows are served by `http.Server`:

```haira
fn main() {
    http.Server([Chat]).listen(8080)
}
```

To view your workflows in a web UI, run:

```bash
haira webui -c localhost:8080 -p 3000
```

## How They Fit Together

```
Provider  →  configures which LLM to use
   ↓
Tool      →  functions the agent can call
   ↓
Agent     →  LLM entity with tools + memory
   ↓
Workflow  →  HTTP endpoint that orchestrates agents
   ↓
Server    →  serves everything as a native binary
```

## Core Language Features

Beyond the agentic primitives, Haira has a full set of language features:

| Feature | Syntax |
|---------|--------|
| Variables | `x = 42` (type inferred) |
| Functions | `fn add(a: int, b: int) -> int { return a + b }` |
| Strings | `"Hello, ${name}!"` with interpolation |
| Control flow | `if`, `for`, `match` |
| Structs | `struct User { name: string, age: int }` |
| Enums | `enum Color { Red, Green, Blue }` |
| Lists | `[1, 2, 3]` |
| Maps | `{"key": "value"}` |
| Error handling | `result, err = call()` |
| Pattern matching | `match x { 1..5 => "low", _ => "high" }` |
| Pipe operator | `data |> transform |> format` |
| Parallel execution | `spawn { task1(), task2() }` |
| Methods | `Type.method()` with implicit `self` |

## Visibility

Everything is **private by default**. Use `pub` to export:

```haira
pub fn helper() -> string {
    return "I'm public"
}

fn internal() -> string {
    return "I'm private"
}
```

Agentic declarations (`provider`, `tool`, `agent`, `workflow`) are always public.

## Next Steps

- [Variables & Types](/docs/language/variables-and-types) — learn the type system
- [Providers](/docs/agentic/providers) — configure LLM backends in detail
- [Agents](/docs/agentic/agents) — build agents with tools and memory
- [Workflows](/docs/agentic/workflows) — create HTTP endpoints
