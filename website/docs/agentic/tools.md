---
title: "Tools"
description: "Define type-checked tools with LLM-visible descriptions for your agents."
---

# Tools

Tools are functions that agents can call. The LLM decides when and how to use them based on the docstring.

## Defining a Tool

```haira
tool get_weather(city: string) -> string {
    """Get the current weather for a given city"""
    resp, err = http.get("https://wttr.in/${city}?format=j1")
    if err != nil {
        return "Failed to fetch weather data."
    }
    data = resp.json()
    current = data["current_condition"][0]
    return "${city}: ${current["temp_C"]}°C, ${current["weatherDesc"][0]["value"]}"
}
```

## Docstrings Are Required

The triple-quote docstring (`"""..."""`) is **mandatory** and compiler-enforced. This description is sent to the LLM so it knows:
- What the tool does
- When to use it
- What the parameters mean

```haira
tool search_docs(query: string) -> string {
    """Search the documentation for relevant articles.
    Use this when the user asks about product features or troubleshooting."""
    // implementation
}
```

## Tool Parameters

Tools can have multiple typed parameters:

```haira
tool create_ticket(title: string, body: string, priority: string) -> string {
    """Create a support ticket. Priority can be: low, medium, high"""
    // Create the ticket
    resp, err = http.post("https://api.tickets.com/create", {
        "title": title,
        "body": body,
        "priority": priority
    })
    if err != nil { return "Failed to create ticket." }
    return "Ticket created: ${resp.json()["id"]}"
}
```

## Assigning Tools to Agents

```haira
agent SupportBot {
    provider: openai
    system: "Help users with their questions. Create tickets for complex issues."
    tools: [search_docs, create_ticket]
    memory: conversation(max_turns: 20)
}
```

The agent will automatically decide when to call each tool based on the conversation.

## Database Tools

```haira
import "postgres"

tool lookup_user(email: string) -> string {
    """Look up a user by their email address"""
    rows, err = postgres.query("SELECT name, plan FROM users WHERE email = $1", email)
    if err != nil { return "Database error." }
    if len(rows) == 0 { return "User not found." }
    return "User: ${rows[0]["name"]}, Plan: ${rows[0]["plan"]}"
}
```

## HTTP Tools

```haira
tool fetch_api(endpoint: string) -> string {
    """Fetch data from the internal API"""
    resp, err = http.get("https://internal-api.com/${endpoint}")
    if err != nil { return "API request failed." }
    return resp.body
}
```

## Lifecycle Hooks

Tools support `@before` and `@after` hooks for pre/post-processing logic:

### `@before` — Run before tool execution

```haira
tool create_ticket(title: string, body: string) -> string {
    """Create a support ticket"""

    @before {
        io.println("Creating ticket: ${title}")
        if len(title) == 0 {
            return "Title is required."
        }
    }

    resp, err = http.post("https://api.tickets.com/create", {
        "title": title,
        "body": body
    })
    if err != nil { return "Failed to create ticket." }
    return "Ticket created: ${resp.json()["id"]}"
}
```

### `@after` — Run after tool execution

```haira
tool query_db(sql: string) -> string {
    """Execute a database query"""

    @after {
        io.println("Query completed")
        observe.track("db_query", { "sql": sql })
    }

    rows, err = postgres.query(sql)
    if err != nil { return "Query failed." }
    return json.encode(rows)
}
```

Hooks are useful for:
- **Validation** — check inputs before running
- **Logging** — track tool usage
- **Metrics** — record performance data
- **Cleanup** — release resources after execution

## Best Practices

- Write clear, specific docstrings — the LLM relies on them
- Return strings — agents work with text
- Handle errors gracefully — return error messages instead of crashing
- Keep tools focused — one tool per action
- Use `@before` hooks for input validation
- Use `@after` hooks for observability and cleanup
