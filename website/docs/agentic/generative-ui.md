---
title: "Generative UI"
description: "Agents render rich UI components — tables, charts, forms, status cards — inline."
---

# Generative UI

Every Haira workflow automatically gets a web UI — no frontend code needed.

## Auto-generated Forms

When you define a workflow, Haira can generate a web UI for it via the `haira webui` command:

```haira
@post("/api/analyze")
workflow Analyze(text: string, mode: string) -> { result: string } {
    result, err = Analyst.ask("${mode}: ${text}")
    if err != nil { return { result: "Error." } }
    return { result: result }
}

fn main() {
    http.Server([Analyze]).listen(8080)
}
```

Run the webui:

```bash
haira webui -c localhost:8080 -p 3000
```

Open `http://localhost:3000` and you'll see a form with:
- A text input for `text`
- A text input for `mode`
- A "Run" button
- A response display area

## Custom UI Metadata

Use `@webui` to customize the generated UI:

```haira
@webui(title: "File Summarizer", description: "Upload a text file and get an AI summary")
@post("/summarize")
workflow Summarize(document: file, context: string, format: string) -> { summary: string } {
    content, err = io.read_file(document)
    if err != nil { return { summary: "Failed to read file." } }

    prompt = "Summarize in ${format}: ${content}"
    reply, err = Summarizer.ask(prompt)
    if err != nil { return { summary: "AI error." } }
    return { summary: reply }
}
```

The `@webui` decorator sets:
- `title` — displayed at the top of the form
- `description` — subtitle text

## File Uploads

Parameters with type `file` render as file upload inputs:

```haira
workflow Process(document: file, image: file) -> { result: string } {
    // Both parameters get file upload inputs in the UI
}
```

## Streaming Chat UI

Streaming workflows (`-> stream`) get a full chat interface instead of a form:

```haira
@post("/chat")
workflow Chat(message: string, session_id: string) -> stream {
    return Assistant.stream(message, session: session_id)
}
```

The chat UI includes:
- Message input with send button
- Real-time token streaming display
- Conversation history
- Session management

## How It Works

The UI is built with Lit web components (TypeScript), bundled as `bundle.tar.gz`, and embedded directly in the compiled binary. No external dependencies or CDN calls at runtime.

Under the hood, the UI communicates with the server via the [ARP protocol](/docs/agentic/arp) (Agent Rendering Protocol) — a bidirectional streaming protocol that handles text deltas, tool events, and rich UI component rendering.

The generated UI reads workflow metadata:
- Parameter names and types from the workflow signature
- Title and description from `@webui`
- Streaming capability from the `-> stream` return type

## Multiple Workflows

All workflows share a single UI hub:

```haira
fn main() {
    http.Server([Chat, Analyze, Summarize]).listen(8080)
}
```

```bash
haira webui -c localhost:8080 -p 3000
```

The webui discovers all workflows and presents them in a unified interface.
