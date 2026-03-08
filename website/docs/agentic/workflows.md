---
title: "Workflows"
description: "Expose agents as HTTP endpoints with workflow triggers and lifecycle hooks."
---

# Workflows

Workflows are functions decorated with HTTP triggers. They define your API endpoints and orchestrate agents.

## Basic Workflow

```haira
@post("/api/chat")
workflow Chat(message: string) -> { reply: string } {
    reply, err = Assistant.ask(message)
    if err != nil {
        return { reply: "Something went wrong." }
    }
    return { reply: reply }
}
```

## HTTP Decorators

### `@post` — POST endpoint

```haira
@post("/api/analyze")
workflow Analyze(text: string) -> { result: string } {
    // ...
}
```

### `@webhook` — Webhook endpoint

```haira
@webhook("/hooks/github")
workflow GithubWebhook(payload: string) -> { status: string } {
    // Process webhook
    return { status: "ok" }
}
```

### `@webui` — Custom UI metadata

```haira
@webui(title: "Text Analyzer", description: "Analyze text with AI")
@post("/api/analyze")
workflow Analyze(text: string) -> { result: string } {
    // ...
}
```

## Input Parameters

Workflow parameters map to the JSON body of the HTTP request:

```haira
@post("/api/process")
workflow Process(
    text: string,
    language: string,
    max_length: int
) -> { result: string } {
    // Access parameters directly
    io.println("Processing ${language} text...")
    // ...
}
```

Request body:
```json
{
    "text": "Hello world",
    "language": "en",
    "max_length": 100
}
```

## File Uploads

Use the `file` type for file upload support:

```haira
@post("/api/summarize")
workflow Summarize(document: file, context: string) -> { summary: string } {
    content, err = io.read_file(document)
    if err != nil {
        return { summary: "Failed to read file." }
    }
    // Process content...
}
```

The auto-generated UI will render a file upload input.

## Streaming Workflows

Return `stream` for SSE responses:

```haira
@post("/chat")
workflow Chat(message: string, session_id: string) -> stream {
    return Assistant.stream(message, session: session_id)
}
```

See [Streaming](/docs/agentic/streaming) for details.

## Serving Workflows

Register workflows with `http.Server`:

```haira
fn main() {
    server = http.Server([Chat, Analyze, Summarize])
    io.println("Server running on :8080")
    server.listen(8080)
}
```

Every workflow gets its API endpoint (e.g., `POST /api/chat`). To view workflows in a web UI:

```bash
haira webui -c localhost:8080 -p 3000
```

## Multi-Step Workflows

Use `step` blocks for named stages:

```haira
@post("/api/analyze")
workflow Analyze(text: string) -> { result: string } {
    step "Validate input" {
        if len(text) == 0 {
            return { result: "Empty input." }
        }
    }

    step "Analyze text" {
        analysis, err = Analyst.ask("Analyze: ${text}")
        if err != nil {
            return { result: "Analysis failed." }
        }
    }

    return { result: analysis }
}
```

### Step Retry

`@retry` adds automatic retry with backoff:

```haira
@retry(max: 10, delay: 5000, backoff: "exponential")
step "Call external API" {
    result = http.get(url)
}
```

### Verification Loops

Use `verify { assert ... }` inside steps to define assertions that trigger retries automatically:

```haira
@retry(max: 3)
step "Generate analysis" {
    analysis, err = Analyst.ask("Analyze: ${data}")
    if err != nil { return { result: "Failed." } }

    verify {
        assert len(analysis) > 100
        assert strings.contains(analysis, "conclusion")
    }
}
```

When an assertion fails inside a `@retry` step, the step automatically retries. Without `@retry`, assertion failures return an error.

## Lifecycle Hooks

Workflows support lifecycle hooks for error handling and completion logic:

### `onerror` — Handle errors

```haira
@post("/api/process")
workflow Process(text: string) -> { result: string } {
    onerror err {
        io.println("Workflow failed: ${err}")
        return { result: "Error: could not process" }
    }

    result, err = Analyzer.ask(text)
    if err != nil { return { result: "AI error" } }
    return { result: result }
}
```

### `onsuccess` — Run on completion

```haira
@post("/api/analyze")
workflow Analyze(text: string) -> { summary: string } {
    onsuccess {
        io.println("Analysis completed successfully")
    }

    summary, err = Summarizer.ask(text)
    if err != nil { return { summary: "Failed" } }
    return { summary: summary }
}
```

### `oncancel` — Run on cancellation

```haira
@post("/api/long-task")
workflow LongTask(input: string) -> { result: string } {
    oncancel {
        io.println("Task was cancelled, cleaning up...")
    }

    // ... long-running work
    return { result: "done" }
}
```

## Additional HTTP Decorators

### `@get` — GET endpoint

```haira
@get("/api/status")
workflow Status() -> { status: string } {
    return { status: "ok" }
}
```

### `@put` — PUT endpoint

```haira
@put("/api/users")
workflow UpdateUser(id: int, name: string) -> { success: bool } {
    // update logic
    return { success: true }
}
```

### `@delete` — DELETE endpoint

```haira
@delete("/api/users")
workflow DeleteUser(id: int) -> { deleted: bool } {
    // delete logic
    return { deleted: true }
}
```

## Orchestrating Multiple Agents

```haira
@post("/api/research")
workflow Research(topic: string) -> { report: string } {
    // Gather data in parallel
    data = spawn {
        Researcher.ask("Find facts about: ${topic}")
        Researcher.ask("Find recent news about: ${topic}")
    }

    // Synthesize
    combined = string.join(data, "\n\n")
    report, err = Writer.ask("Write a report from: ${combined}")
    if err != nil {
        return { report: "Failed to generate report." }
    }
    return { report: report }
}
```
