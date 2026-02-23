---
title: "Agent Handoffs"
description: "Route between specialized agents automatically with the handoffs field."
---

# Agent Handoffs

Agent handoffs let you route conversations between specialized agents. A front desk agent can automatically forward to billing, tech support, or other specialists.

## Defining Handoffs

Use the `handoffs` field on an agent:

```haira
agent BillingAgent {
    provider: openai
    system: "You handle billing and payment questions."
    memory: conversation(max_turns: 20)
}

agent TechAgent {
    provider: openai
    system: "You handle technical support questions."
    memory: conversation(max_turns: 20)
}

agent FrontDesk {
    provider: openai
    system: """You are a front desk agent. Greet users and route them:
    - Billing/payment questions → BillingAgent
    - Technical issues → TechAgent
    - General questions → answer yourself"""
    handoffs: [BillingAgent, TechAgent]
    memory: conversation(max_turns: 10)
    temperature: 0.3
}
```

## `.ask()` — Automatic Handoffs

When using `.ask()`, handoffs are followed automatically. The front desk agent will route to the appropriate specialist, and you get the final response:

```haira
@post("/support")
workflow Support(message: string, session_id: string) -> { reply: string } {
    reply, err = FrontDesk.ask(message, session: session_id)
    if err != nil {
        return { reply: "Something went wrong." }
    }
    return { reply: reply }
}
```

If the user asks "I can't log in", FrontDesk automatically hands off to TechAgent.

## `.run()` — Manual Control

With `.run()`, you get an `AgentResult` and can inspect the handoff:

```haira
result: AgentResult, err = FrontDesk.run(message, session: session_id)
// Check which agent handled the request
// Access intermediate routing decisions
```

## How It Works

1. The FrontDesk agent receives the user message
2. Based on the conversation, it decides to hand off (or not)
3. The handoff target agent receives the conversation context
4. The target agent generates the response
5. With `.ask()`, the final response is returned automatically

## Initialization Order

The compiler topologically sorts agent initialization — handoff targets are initialized first. So `BillingAgent` and `TechAgent` are ready before `FrontDesk`.

## Multi-level Handoffs

Agents can hand off to agents that also have handoffs, creating a routing tree:

```haira
agent L1Support {
    provider: openai
    system: "Basic support. Escalate complex issues."
    handoffs: [L2Support, BillingAgent]
}

agent L2Support {
    provider: anthropic
    system: "Advanced technical support."
    handoffs: [Engineering]
}

agent Engineering {
    provider: anthropic
    system: "Engineering-level debugging."
    tools: [query_logs, check_status]
}
```

## Complete Example

```haira
import "io"
import "http"

provider openai {
    api_key: env("OPENAI_API_KEY")
    model: "gpt-4o"
}

agent BillingAgent {
    provider: openai
    system: "Handle billing inquiries. Be helpful and clear about pricing."
    memory: conversation(max_turns: 20)
}

agent TechAgent {
    provider: openai
    system: "Handle technical issues. Ask for error messages and steps to reproduce."
    memory: conversation(max_turns: 20)
}

agent FrontDesk {
    provider: openai
    system: "Greet users. Route billing to BillingAgent, tech to TechAgent."
    handoffs: [BillingAgent, TechAgent]
    memory: conversation(max_turns: 10)
    temperature: 0.3
}

@post("/api/chat")
workflow Chat(message: string, session_id: string) -> { reply: string } {
    reply, err = FrontDesk.ask(message, session: session_id)
    if err != nil {
        return { reply: "Something went wrong." }
    }
    return { reply: reply }
}

fn main() {
    http.Server([Chat]).listen(8080)
}
```
