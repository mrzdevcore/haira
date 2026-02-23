---
title: "Memory & Sessions"
description: "Conversation memory, session persistence, and summary memory for agents."
---

# Memory & Sessions

Agents in Haira have built-in memory management for multi-turn conversations.

## Memory Strategies

### Conversation Memory

Stores the last N turns of conversation:

```haira
agent Assistant {
    provider: openai
    system: "You are a helpful assistant."
    memory: conversation(max_turns: 10)
}
```

This keeps the 10 most recent user/assistant exchanges in context.

### Summary Memory

Summarizes older messages to stay within a token budget:

```haira
agent LongChat {
    provider: openai
    system: "You are a long-running assistant."
    memory: summary(max_tokens: 2000)
}
```

When the conversation exceeds the token limit, older messages are summarized.

### No Memory

Stateless — each call is independent:

```haira
agent Classifier {
    provider: openai
    system: "Classify the given text."
    memory: none
}
```

## Sessions

Sessions identify a conversation. Pass `session` as a named parameter:

```haira
// First message in a session
reply, err = Assistant.ask("Hi, I'm Alice", session: "user-123")

// Second message — agent remembers the first
reply, err = Assistant.ask("What's my name?", session: "user-123")
// Agent replies: "Your name is Alice"
```

### Session IDs in Workflows

Typically the client sends a session ID:

```haira
@post("/api/chat")
workflow Chat(message: string, session_id: string) -> { reply: string } {
    reply, err = Assistant.ask(message, session: session_id)
    if err != nil {
        return { reply: "Error." }
    }
    return { reply: reply }
}
```

### Independent Sessions

Different sessions are completely independent:

```haira
// Session A
Assistant.ask("I like cats", session: "session-a")

// Session B — knows nothing about cats
Assistant.ask("What do I like?", session: "session-b")
// Agent doesn't know
```

## Choosing a Strategy

| Strategy | Use Case | Trade-off |
|----------|----------|-----------|
| `conversation(max_turns: N)` | Chatbots, support agents | Simple, predictable context window |
| `summary(max_tokens: N)` | Long-running conversations | Preserves context, uses more LLM calls |
| `none` | Stateless tasks (classification, extraction) | No overhead, no context |

## Storage

Session data is stored automatically. When workflows are detected, the compiler includes SQLite for persistence. For production, you can use Postgres:

```haira
import "postgres"

// Sessions persist across restarts
```
