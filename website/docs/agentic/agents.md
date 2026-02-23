---
title: "Agents"
description: "Create AI agents with providers, tools, memory, and system prompts."
---

# Agents

An agent is an LLM entity with a provider, system prompt, optional tools, and memory configuration.

## Defining an Agent

```haira
agent Assistant {
    provider: openai
    system: "You are a helpful assistant. Be concise."
    tools: [get_weather, search_docs]
    memory: conversation(max_turns: 10)
    temperature: 0.7
}
```

## Agent Fields

| Field | Required | Description |
|-------|----------|-------------|
| `provider` | Yes | Which LLM provider to use |
| `system` | Yes | System prompt |
| `tools` | No | List of tools the agent can call |
| `memory` | No | Memory strategy |
| `temperature` | No | LLM temperature (0.0-2.0) |
| `handoffs` | No | List of agents to hand off to |

## Calling Agents

### `.ask()` — Simple request/response

```haira
reply, err = Assistant.ask("What's the weather in Tokyo?")
if err != nil {
    io.println("Error: ${err}")
}
io.println(reply)
```

### `.run()` — Full control

Returns an `AgentResult` with more details:

```haira
result: AgentResult, err = Assistant.run("Analyze this data")
if err != nil { /* handle */ }
// Access result fields
```

### `.stream()` — SSE streaming

Returns a stream for Server-Sent Events:

```haira
@post("/chat")
workflow Chat(msg: string) -> stream {
    return Assistant.stream(msg)
}
```

## Sessions

All agent methods accept a `session` named parameter for multi-turn conversations:

```haira
reply, err = Assistant.ask("Hello!", session: session_id)
// Later, same session...
reply2, err = Assistant.ask("What did I just say?", session: session_id)
```

The agent will remember the conversation history within the session.

## Temperature

Control creativity vs determinism:

```haira
agent Coder {
    provider: openai
    system: "Write clean, correct code."
    temperature: 0.1  // Low: deterministic
}

agent Writer {
    provider: openai
    system: "Write creative stories."
    temperature: 0.9  // High: creative
}
```

## Multiple Agents

Define as many agents as you need, potentially with different providers:

```haira
provider openai { api_key: env("OPENAI_API_KEY"), model: "gpt-4o" }
provider anthropic { api_key: env("ANTHROPIC_API_KEY"), model: "claude-sonnet-4-5-20250929" }

agent Researcher {
    provider: anthropic
    system: "Find and summarize information."
    tools: [search_docs]
    temperature: 0.3
}

agent Writer {
    provider: openai
    system: "Write engaging content based on research."
    temperature: 0.8
}

@post("/article")
workflow WriteArticle(topic: string) -> { article: string } {
    research, err = Researcher.ask("Research: ${topic}")
    if err != nil { return { article: "Research failed." } }

    article, err = Writer.ask("Write an article based on: ${research}")
    if err != nil { return { article: "Writing failed." } }

    return { article: article }
}
```

## Structured Output

Request structured responses from agents:

```haira
struct Analysis {
    sentiment: string
    score: float
    keywords: []string
}

result: Analysis, err = Analyzer.run("Analyze this customer review...")
```

The left-side type annotation tells the agent to return data matching that structure.

## Next Steps

- [Handoffs](/docs/agentic/handoffs) — route between agents
- [Memory](/docs/agentic/memory) — conversation and summary memory
- [Streaming](/docs/agentic/streaming) — real-time SSE responses
