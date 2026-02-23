---
title: "Examples"
description: "29 example programs demonstrating Haira language features and agentic patterns."
---

# Examples

Complete examples from the [examples directory](https://github.com/mrzdevcore/haira/tree/main/examples).

## Hello World

```haira
import "io"

fn main() {
    io.println("Hello, World!")
}
```

## Simple Chatbot

A streaming chatbot with memory:

```haira
import "io"
import "http"

provider openai {
    api_key: env("OPENAI_API_KEY")
    model: "gpt-4o"
}

agent Writer {
    provider: openai
    system: "You are a creative writer."
    memory: conversation(max_turns: 10)
    temperature: 0.9
}

@post("/chat")
workflow Chat(msg: string, sid: string) -> stream {
    return Writer.stream(msg, session: sid)
}

fn main() {
    http.Server([Chat]).listen(8080)
}
```

## Weather Agent

Agent with a tool that fetches live weather data:

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
    temp_c = current["temp_C"]
    desc = current["weatherDesc"][0]["value"]
    return "${city}: ${temp_c}°C, ${desc}"
}

agent Assistant {
    provider: openai
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
    io.println("Server running on :9001")
    server.listen(9001)
}
```

## Multi-Agent Support Desk

Front desk routing to specialized agents:

```haira
import "io"
import "http"

provider openai {
    api_key: env("OPENAI_API_KEY")
    model: "gpt-4o"
}

agent BillingAgent {
    provider: openai
    system: "Handle billing and payment questions."
    memory: conversation(max_turns: 20)
}

agent TechAgent {
    provider: openai
    system: "Handle technical support questions."
    memory: conversation(max_turns: 20)
}

agent FrontDesk {
    provider: openai
    system: "Greet users. Route billing to BillingAgent, tech to TechAgent."
    handoffs: [BillingAgent, TechAgent]
    memory: conversation(max_turns: 10)
    temperature: 0.3
}

@post("/support")
workflow Support(msg: string, sid: string) -> { reply: string } {
    reply, err = FrontDesk.ask(msg, session: sid)
    if err != nil { return { reply: "Something went wrong." } }
    return { reply: reply }
}

fn main() {
    http.Server([Support]).listen(8080)
}
```

## File Upload & Summarize

```haira
import "io"
import "http"

provider openai {
    api_key: env("OPENAI_API_KEY")
    model: "gpt-4o"
}

agent Summarizer {
    provider: openai
    system: "Summarize the given text clearly and concisely."
    temperature: 0.3
}

@webui(title: "File Summarizer", description: "Upload a text file and get an AI summary")
@post("/summarize")
workflow Summarize(document: file, context: string) -> { summary: string } {
    content, err = io.read_file(document)
    if err != nil { return { summary: "Failed to read file." } }

    prompt = "Summarize this: ${context}\n\n${content}"
    reply, err = Summarizer.ask(prompt)
    if err != nil { return { summary: "AI error." } }
    return { summary: reply }
}

fn main() {
    http.Server([Summarize]).listen(8080)
}
```

## Pipe Operator

```haira
import "io"
import "string"

fn double(x: int) -> int { return x * 2 }
fn add_ten(x: int) -> int { return x + 10 }

fn main() {
    result = 5 |> double |> add_ten
    io.println("Result: ${result}")  // 20

    text = "  Hello World  "
        |> string.trim
        |> string.to_lower
    io.println(text)  // "hello world"
}
```

## Pattern Matching

```haira
import "io"

enum Shape {
    Circle(float)
    Rectangle(float, float)
    Point
}

fn area(s: Shape) -> float {
    match s {
        Shape.Circle(r) => 3.14159 * r * r,
        Shape.Rectangle(w, h) => w * h,
        Shape.Point => 0.0
    }
}

fn main() {
    shapes = [
        Shape.Circle(5.0),
        Shape.Rectangle(3.0, 4.0),
        Shape.Point
    ]

    for shape in shapes {
        io.println("Area: ${area(shape)}")
    }
}
```

## More Examples

See the full [examples directory on GitHub](https://github.com/mrzdevcore/haira/tree/main/examples) with 29 examples covering:

| # | Example | Description |
|---|---------|-------------|
| 01 | Hello World | Basic program |
| 02 | Variables | Type inference and assignment |
| 03 | Functions | Function definitions |
| 04 | Control Flow | If/else, for loops |
| 05 | Match | Pattern matching |
| 06 | Lists | List operations |
| 07 | Agentic | Providers, tools, agents |
| 08 | Structs | Struct definitions and use |
| 09 | String Interpolation | `${expr}` syntax |
| 10 | Maps | Map operations |
| 11 | Pipes | Pipe operator |
| 12 | Methods | Type methods with self |
| 13 | Error Handling | Error patterns |
| 14 | Multi-Agent | Multiple agents |
| 15 | Handoffs | Agent routing |
| 16 | Enums | Enum definitions |
| 17 | Compound Assignment | `+=`, `-=`, etc. |
| 18 | Defer | Cleanup with defer |
| 19 | Streaming | SSE streaming |
| 20 | Stdlib | Standard library usage |
| 21 | File Upload | File handling |
| 22 | Pipeline UI | Step-based workflows |
| 23 | MCP Client | MCP protocol |
| 24 | MCP Server | MCP server |
| 25 | Embeddings | Vector embeddings |
| 26 | RAG | Retrieval augmented generation |
| 27 | Structured Output | Typed agent responses |
| 28 | Observe | LLM observability |
| 29 | Testing | Test framework |
