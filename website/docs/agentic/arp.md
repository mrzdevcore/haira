# ARP — Agent Rendering Protocol

ARP is a transport-agnostic, bidirectional protocol for communication between agent backends and frontend renderers. It solves the "generative UI" problem — how does a streaming agent tell a frontend "render a table/chart/form here" alongside plain text?

Every Haira server speaks ARP natively. No configuration needed.

## Overview

```
┌──────────────────────┐         ARP Messages         ┌──────────────────────┐
│   Haira Agent        │ ─────────────────────────────▸│   Frontend           │
│   (Go backend)       │◂──────────────────────────────│   (React/Vue/Web)    │
│                      │     WebSocket or SSE          │                      │
│  tool returns        │                               │  ArpRenderer maps    │
│  ui.table(...)  ─────┼──▸ { type: "render",          │  "table" → <Table>   │
│                      │      component: "table",      │                      │
│                      │      props: {...} }           │                      │
└──────────────────────┘                               └──────────────────────┘
```

## Protocol (v1)

ARP v1 operates in **Minimal Mode** — a flat, session-addressed message format. Every message is JSON with at minimum `{ v: 1, type: "<type>" }`.

### Server → Client Messages

| Type | Purpose | Key Fields |
|------|---------|------------|
| `hello` | Capability handshake on connect | `capabilities.components[]`, `capabilities.features[]` |
| `delta` | Incremental text chunk | `payload.delta` |
| `tool_start` | Tool execution started | `payload.tool`, `payload.args` |
| `tool_end` | Tool execution finished | `payload.tool`, `payload.ok` |
| `render` | Generative UI component | `components[].type`, `components[].props`, `tool_name` |
| `patch` | Incremental component update | `ops[]` with `op`, `target`, `path`, `value` |
| `error` | Error event | `payload.error` |
| `commit` | Stream complete | `final: true` |

### Client → Server Messages

| Type | Input Type | Data Fields |
|------|-----------|-------------|
| `input` | `text` | `{ text }` |
| `input` | `action` | `{ action, payload }` |
| `input` | `form_submit` | `{ action, fields }` |

### Example Message Flow

```
Server → Client:  { v: 1, type: "hello", capabilities: { components: [...], features: ["streaming", "input"] } }
Client → Server:  { v: 1, type: "input", input_type: "text", data: { text: "Show me the sales data" } }
Server → Client:  { v: 1, type: "delta", payload: { delta: "Here are " } }
Server → Client:  { v: 1, type: "delta", payload: { delta: "the results:" } }
Server → Client:  { v: 1, type: "tool_start", payload: { tool: "query_database", args: "..." } }
Server → Client:  { v: 1, type: "render", components: [{ id: "c_1", type: "table", props: {...} }], tool_name: "query_database" }
Server → Client:  { v: 1, type: "tool_end", payload: { tool: "query_database", ok: true } }
Server → Client:  { v: 1, type: "commit", final: true }
```

## Transports

ARP supports two transports, both available from the same Haira server:

### WebSocket (`/_arp/v1`)

The primary transport for chat workflows. Persistent bidirectional connection.

```
1. Client connects to ws://localhost:8080/_arp/v1
2. Server sends "hello" message with capabilities
3. Client sends "input" messages (text, actions, form submits)
4. Server streams back delta/render/tool_start/tool_end/commit messages
```

Features:
- Auto-reconnect with exponential backoff (1s → 16s cap, max 5 attempts)
- Session-addressed messages for multi-session support
- Full bidirectional communication (text input, button clicks, form submissions)

### SSE (Server-Sent Events)

Used as fallback when WebSocket is unavailable, and **required** for file uploads (multipart/form-data). POST to any streaming workflow endpoint with `Accept: text/event-stream`.

```
POST /chat HTTP/1.1
Accept: text/event-stream
Content-Type: application/json

{"message": "Show me the sales data", "session_id": "abc"}
```

SSE event types:

| SSE Event | ARP Type | Format |
|-----------|----------|--------|
| (default) | `delta` | `data: {"delta":"..."}` |
| `tool_start` | `tool_start` | `event: tool_start\ndata: {"tool":"...","args":"..."}` |
| `tool_render` | `render` | `event: tool_render\ndata: {"tool":"...","component":"...","props":{}}` |
| `tool_end` | `tool_end` | `event: tool_end\ndata: {"tool":"...","ok":true}` |
| `error` | `error` | `event: error\ndata: {"error":"..."}` |
| (default) | `commit` | `data: [DONE]` |

## Built-in Components

ARP ships with 14 built-in component types. Every Haira server advertises these in the `hello` message:

| Component | Description | Key Props |
|-----------|-------------|-----------|
| `text` | Plain text | `text` |
| `markdown` | Rendered markdown | `content` |
| `status-card` | Status indicator card | `status`, `title`, `message` |
| `table` | Data table with optional tabs | `title`, `headers`, `rows`, `tabs` |
| `code-block` | Syntax-highlighted code | `title`, `language`, `code`, `tabs` |
| `diff` | Side-by-side diff view | `title`, `before`, `after`, `language` |
| `key-value` | Key-value pairs display | `title`, `items[].key`, `items[].value` |
| `progress` | Multi-step progress tracker | `title`, `steps[].name`, `steps[].status` |
| `chart` | Data visualization | `type`, `title`, `labels`, `datasets` |
| `form` | Interactive form | `title`, `fields[]`, `submit_label` |
| `confirm` | Yes/no confirmation dialog | `title`, `confirm_label`, `deny_label` |
| `choices` | Option picker | `title`, `options[]` |
| `product-cards` | Image card grid | `title`, `cards[].name`, `cards[].price` |
| `image` | Image display | `src`, `alt` |

## Capability Discovery

Query `GET /_api/arp` to discover what the server supports:

```json
{
  "v": 1,
  "type": "hello",
  "capabilities": {
    "components": [
      "text", "status-card", "table", "code-block", "diff",
      "key-value", "progress", "chart", "form", "confirm",
      "choices", "product-cards", "markdown", "image"
    ],
    "features": ["streaming", "input"]
  }
}
```

## Using ARP in Haira

### Rendering UI Components from Tools

Tools return UI components using the `ui` module. The component is streamed to the frontend via ARP while the LLM receives a compact text summary:

```haira
import "ui"

tool query_database(query: string) -> any {
    """Executes a SQL query and displays results."""

    rows, err = postgres.query(db, query)
    if err != nil {
        return ui.status_card("error", "Query Failed", conv.to_string(err))
    }

    return ui.table("Query Results", headers, rows)
}
```

### Available UI Functions

```haira
import "ui"

// Status cards
ui.status_card("success", "Title", "Optional message")
ui.status_card("error", "Title", "Error details")
ui.status_card("warning", "Title")
ui.status_card("info", "Title")

// Tables
ui.table("Title", ["Col A", "Col B"], [["row1a", "row1b"], ["row2a", "row2b"]])

// Key-value displays
ui.key_value("User Details", {"Name": "Alice", "Role": "Admin"})

// Charts
ui.chart("bar", "Revenue", ["Q1", "Q2", "Q3"], [dataset])

// Confirmation dialogs
ui.confirm("Delete this?", "Yes, delete", "Cancel")

// Product card grids
ui.product_cards("Search Results", cards)

// Compose multiple components
ui.group(
    ui.status_card("success", "Query Complete", "42 rows"),
    ui.table("Results", headers, rows)
)
```

### Enabling UI on Agents

Add `ui: ui` to your agent declaration to register all built-in UI tools:

```haira
import "ui"

agent DataExplorer {
    provider: OpenAI
    model: "gpt-4o"
    tools: [query_database, search_data]
    ui: ui
    memory: conversation(max_turns: 30)
}
```

The agent can then use UI components in any of its tools. The frontend receives `render` events and displays the appropriate component.

## Client SDKs

ARP client libraries let you build custom frontends that connect to any Haira server (or any ARP-compatible backend).

### `@haira/arp` — Core (Zero Dependencies)

```bash
npm install @haira/arp
```

WebSocket client:

```typescript
import { ArpClient } from '@haira/arp'

const client = new ArpClient('ws://localhost:8080/_arp/v1', {
  onConnect: (caps) => console.log('Components:', caps.components),
  onDelta: (text) => appendToChat(text),
  onRender: (event) => renderComponent(event.component, event.props),
  onToolStart: (tool) => showToolRunning(tool),
  onToolEnd: (tool, ok) => showToolDone(tool, ok),
  onDone: () => markStreamComplete(),
  onError: (err) => showError(err),
})

client.connect()
client.sendText('Show me the sales data')
```

SSE client (for file uploads or WebSocket fallback):

```typescript
import { streamSSE } from '@haira/arp'

streamSSE('http://localhost:8080/chat', {
  body: formData, // supports FormData for file uploads
  onDelta: (text) => appendToChat(text),
  onToolRender: (event) => renderComponent(event),
  onDone: () => markComplete(),
})
```

Session management:

```typescript
import { createSessionAPI } from '@haira/arp'

const api = createSessionAPI('http://localhost:8080')
const sessions = await api.listSessions()
const session = await api.getSession('session-id')
await api.deleteSession('session-id')
```

### `@haira/arp-react` — React Hooks & Components

```bash
npm install @haira/arp-react
```

**Drop-in chat UI** — zero-config full chat interface:

```tsx
import { ArpChat } from '@haira/arp-react'

function App() {
  return (
    <ArpChat
      url="ws://localhost:8080/_arp/v1"
      theme="dark"
      accentColor="#E8A317"
      title="Data Explorer"
      suggestions={['Show database schema', 'Query recent orders']}
    />
  )
}
```

**Custom UI with hooks** — full control over rendering:

```tsx
import { useArpChat } from '@haira/arp-react'
import { DEFAULT_COMPONENTS } from '@haira/arp-react/ui'

function Chat() {
  const {
    messages,
    isStreaming,
    isConnected,
    toolCards,
    sendMessage,
    sessions,
    startNewChat,
  } = useArpChat({
    url: 'ws://localhost:8080/_arp/v1',
    components: DEFAULT_COMPONENTS,
  })

  return (
    <div>
      {messages.map(msg => <Message key={msg.id} {...msg} />)}
      <input onSubmit={(text) => sendMessage(text)} />
    </div>
  )
}
```

Built-in UI components (import from `@haira/arp-react/ui`):
`StatusCard`, `Table`, `CodeBlock`, `Diff`, `KeyValue`, `Progress`, `Chart`, `Form`, `Confirm`, `Choices`, `ProductCards`

### `@haira/arp-vue` — Vue Composables & Components

```bash
npm install @haira/arp-vue
```

**Drop-in chat UI:**

```vue
<script setup>
import { ArpChat } from '@haira/arp-vue'
</script>

<template>
  <ArpChat
    url="ws://localhost:8080/_arp/v1"
    theme="dark"
    accentColor="#E8A317"
    title="Data Explorer"
    :suggestions="['Show database schema', 'Query recent orders']"
  />
</template>
```

**Custom UI with composables:**

```vue
<script setup>
import { useArpChat } from '@haira/arp-vue'

const { messages, isStreaming, sendMessage, startNewChat } = useArpChat({
  url: 'ws://localhost:8080/_arp/v1',
})
</script>
```

Built-in UI components available at `@haira/arp-vue/ui`.

### Go SDK

The Go SDK is available as a standalone package for building ARP-compatible backends outside Haira:

```bash
go get github.com/haira-lang/arp-go
```

```go
import "github.com/haira-lang/arp-go"

// Create a server with default capabilities
server := arp.NewServer(arp.ServerConfig{
    InputHandler: func(sessionID, inputType, text string, raw json.RawMessage) (<-chan arp.StreamChunk, error) {
        chunks := make(chan arp.StreamChunk)
        go func() {
            defer close(chunks)
            chunks <- arp.StreamChunk{Delta: "Hello from ARP!"}
            chunks <- arp.StreamChunk{Done: true}
        }()
        return chunks, nil
    },
})

server.ListenAndServe(":8080")
```

## API Routes

Every Haira server automatically exposes these ARP-related routes:

| Route | Method | Description |
|-------|--------|-------------|
| `/_arp/v1` | WebSocket | ARP WebSocket endpoint |
| `/_api/arp` | GET | Capability discovery |
| `/_api/workflows` | GET | List available workflows |
| `/_api/chats` | GET | List chat sessions |
| `/_api/chats/{id}` | GET | Get session with message history |
| `/_api/chats/{id}` | DELETE | Delete a session |
| `/_api/runs` | GET | Run history |
| `/_api/runs/{id}` | GET | Run detail |
| `/_api/runs/stream/{id}` | GET (SSE) | Reconnect to in-progress run |

## Patch Operations

For incremental component updates without re-rendering the entire component:

| Operation | Description |
|-----------|-------------|
| `update` | Update a property at `path` on `target` component |
| `insert` | Insert a new component `after` a target |
| `remove` | Remove the `target` component |
| `replace` | Replace `target` with a new `component` |
| `reorder` | Move `target` to `after` another component |

```json
{
  "v": 1,
  "type": "patch",
  "ops": [
    { "op": "update", "target": "c_1", "path": "props.rows", "value": [[...]] }
  ]
}
```
