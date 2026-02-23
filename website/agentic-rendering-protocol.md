---
layout: page
title: ARP — Agentic Rendering Protocol
description: A transport-agnostic protocol for communication between AI agents and rendering surfaces.
---

<div class="blog-page">
<div class="blog-hero">
  <div class="blog-hero-glow"></div>
  <div class="blog-hero-inner">
    <div class="blog-badge">Protocol Specification</div>
    <h1 class="blog-title">Agentic Rendering Protocol</h1>
    <p class="blog-subtitle">A transport-agnostic, bidirectional protocol for communication between AI agents and rendering surfaces. One agent, every screen.</p>
    <div class="blog-meta">
      <div class="blog-meta-item">v0.1 Draft</div>
      <div class="blog-meta-sep"></div>
      <div class="blog-meta-item">Open Standard</div>
      <div class="blog-meta-sep"></div>
      <div class="blog-meta-item">CC BY 4.0</div>
    </div>
  </div>
</div>

<div class="blog-body">

<div class="blog-section">

## The Problem

AI agents produce structured output — tables, forms, charts, status messages, code blocks — but the rendering of that output is tightly coupled to the transport and frontend framework. An agent built for a web chat cannot render on a desktop app. An agent behind an SSE stream cannot serve a mobile client expecting a different protocol.

Every new rendering surface requires custom integration code.

</div>

<div class="blog-section">

## The Solution

ARP defines a **protocol** — not a framework, not a library. Like how Wayland decoupled applications from display servers, ARP decouples agents from renderers.

The agent decides **what** to display. The renderer decides **how** to display it.

<div class="blog-card">
<div class="blog-card-header">Architecture</div>

```
Agent (backend)                              Renderer (web, CLI, mobile)
───────────────                              ──────────────────────────
Owns surfaces         ── render / delta ──►  Owns display
Owns UI state         ◄── input events ────  Owns input routing
```

</div>
</div>

<div class="blog-section">

## Design Principles

<div class="principle-grid">
  <div class="principle-item">
    <div class="principle-num">01</div>
    <div class="principle-text">
      <div class="principle-title">No middleman</div>
      <div class="principle-desc">Agent sends render commands directly. Renderer sends input directly. No intermediate framework.</div>
    </div>
  </div>
  <div class="principle-item">
    <div class="principle-num">02</div>
    <div class="principle-text">
      <div class="principle-title">Async &amp; non-blocking</div>
      <div class="principle-desc">All messages are asynchronous. Neither side blocks waiting for a response.</div>
    </div>
  </div>
  <div class="principle-item">
    <div class="principle-num">03</div>
    <div class="principle-text">
      <div class="principle-title">Every frame is perfect</div>
      <div class="principle-desc">Atomic commits. Render state accumulates in a pending buffer and applies atomically. No flickering.</div>
    </div>
  </div>
  <div class="principle-item">
    <div class="principle-num">04</div>
    <div class="principle-text">
      <div class="principle-title">Capability-driven</div>
      <div class="principle-desc">Renderers declare what they support. Agents adapt. CLI gets tables, web gets everything.</div>
    </div>
  </div>
  <div class="principle-item">
    <div class="principle-num">05</div>
    <div class="principle-text">
      <div class="principle-title">Transport-agnostic</div>
      <div class="principle-desc">Logical messages, not wire formats. WebSocket, SSE, gRPC, Unix sockets, stdio.</div>
    </div>
  </div>
  <div class="principle-item">
    <div class="principle-num">06</div>
    <div class="principle-text">
      <div class="principle-title">Typed components</div>
      <div class="principle-desc">Agents emit typed descriptors like <code>table</code> with <code>headers</code> and <code>rows</code> — not HTML.</div>
    </div>
  </div>
</div>

</div>

<div class="blog-section">

## Protocol Messages

Every ARP message is JSON with at minimum `{ v: 1, type: "<type>" }`.

<div class="blog-card">
<div class="blog-card-header">Server to Client</div>

| Type | Purpose |
|------|---------|
| `hello` | Capability handshake on connect |
| `delta` | Incremental text chunk |
| `tool_start` | Tool execution started |
| `tool_end` | Tool execution finished |
| `render` | Generative UI component |
| `patch` | Incremental component update |
| `error` | Error event |
| `commit` | Stream complete |

</div>

<div class="blog-card">
<div class="blog-card-header">Client to Server</div>

| Type | Input Type | Purpose |
|------|-----------|---------|
| `input` | `text` | User text message |
| `input` | `action` | Button click / UI action |
| `input` | `form_submit` | Form submission |

</div>
</div>

<div class="blog-section">

## 14 Built-in Components

Every ARP-conformant renderer must support at least `text`. Components declare fallback chains — `chart` falls back to `table`, which falls back to `text`.

<div class="component-pills">
  <div class="pill">text</div>
  <div class="pill">markdown</div>
  <div class="pill">status-card</div>
  <div class="pill">table</div>
  <div class="pill">code-block</div>
  <div class="pill">diff</div>
  <div class="pill">key-value</div>
  <div class="pill">progress</div>
  <div class="pill">chart</div>
  <div class="pill">form</div>
  <div class="pill">confirm</div>
  <div class="pill">choices</div>
  <div class="pill">product-cards</div>
  <div class="pill">image</div>
</div>

</div>

<div class="blog-section">

## Transports

<div class="transport-grid">
  <div class="transport-card transport-active">
    <div class="transport-label">Available</div>
    <div class="transport-name">WebSocket</div>
    <div class="transport-path"><code>/_arp/v1</code></div>
    <div class="transport-desc">Primary transport. Persistent bidirectional connection with auto-reconnect.</div>
  </div>
  <div class="transport-card transport-active">
    <div class="transport-label">Available</div>
    <div class="transport-name">SSE</div>
    <div class="transport-path">Server-Sent Events</div>
    <div class="transport-desc">Fallback transport. Required for file uploads via multipart/form-data.</div>
  </div>
  <div class="transport-card">
    <div class="transport-label">Planned</div>
    <div class="transport-name">gRPC</div>
    <div class="transport-path">High-performance</div>
    <div class="transport-desc">For native desktop and mobile applications.</div>
  </div>
  <div class="transport-card">
    <div class="transport-label">Planned</div>
    <div class="transport-name">stdio</div>
    <div class="transport-path">NDJSON framing</div>
    <div class="transport-desc">For CLI renderers and pipe-based integrations.</div>
  </div>
</div>

</div>

<div class="blog-section">

## Client SDKs

<div class="blog-card">
<div class="blog-card-header">@haira/arp &mdash; Core (zero dependencies)</div>

```typescript
import { ArpClient } from '@haira/arp'

const client = new ArpClient('ws://localhost:8080/_arp/v1', {
  onDelta: (text) => appendToChat(text),
  onRender: (event) => renderComponent(event.component, event.props),
  onDone: () => markStreamComplete(),
})

client.connect()
client.sendText('Show me the sales data')
```

</div>

<div class="blog-card">
<div class="blog-card-header">@haira/arp-react &mdash; Drop-in Chat UI</div>

```tsx
import { ArpChat } from '@haira/arp-react'

function App() {
  return (
    <ArpChat
      url="ws://localhost:8080/_arp/v1"
      theme="dark"
      title="Data Explorer"
    />
  )
}
```

</div>

Also available: **`@haira/arp-vue`** for Vue 3 and **`github.com/haira-lang/arp-go`** for Go backends.

</div>

<div class="blog-section">

## Haira Integration

Every Haira server speaks ARP natively. No configuration needed.

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

agent DataExplorer {
    provider: OpenAI
    tools: [query_database]
    ui: ui
}

@webhook("/chat")
workflow Chat(message: string, session_id: string) -> stream {
    return DataExplorer.stream(message, session: session_id)
}
```

</div>

<div class="blog-section">

## Extension Lifecycle

New components follow a three-phase lifecycle inspired by Wayland:

<div class="lifecycle-steps">
  <div class="lifecycle-step">
    <div class="lifecycle-phase">1. Experimental</div>
    <div class="lifecycle-prefix"><code>x-vendor-name</code></div>
    <div class="lifecycle-desc">Vendor-created. Breaking changes allowed.</div>
  </div>
  <div class="lifecycle-step">
    <div class="lifecycle-phase">2. Staging</div>
    <div class="lifecycle-prefix"><code>s-name</code></div>
    <div class="lifecycle-desc">Requires 2+ renderer implementations. Governance review.</div>
  </div>
  <div class="lifecycle-step">
    <div class="lifecycle-phase">3. Core</div>
    <div class="lifecycle-prefix">No prefix</div>
    <div class="lifecycle-desc">Part of the ARP spec. Only additive changes.</div>
  </div>
</div>

</div>

<div class="blog-section">

## Conformance Levels

| Level | Required Components | Target |
|-------|-------------------|--------|
| Minimal | `text` + text input | Voice assistants, IoT |
| Basic | text, table, form, confirm, choices | CLI terminals |
| Standard | All core components, full object model | Web / desktop |
| Full | Standard + streaming, multi-surface, file upload | Rich web apps |

</div>

<div class="blog-cta">
  <a class="blog-cta-btn" href="/docs/agentic/arp">Read the Full Reference</a>
</div>

</div>
</div>

<style>
.blog-page {
  --gold: #E8A317;
  --gold-light: #F0BD4F;
  --gold-glow: #FDE68A;
}

/* ── Hero ── */
.blog-hero {
  position: relative;
  text-align: center;
  padding: 6rem 2rem 5rem;
  overflow: hidden;
}
.blog-hero-glow {
  position: absolute;
  top: -150px;
  left: 50%;
  transform: translateX(-50%);
  width: 800px;
  height: 500px;
  background: radial-gradient(ellipse, rgba(232, 163, 23, 0.1) 0%, transparent 70%);
  pointer-events: none;
  z-index: 0;
}
.blog-hero-inner {
  position: relative;
  z-index: 1;
  max-width: 960px;
  margin: 0 auto;
}
.blog-badge {
  display: inline-block;
  padding: 0.375rem 1.125rem;
  border-radius: 999px;
  font-size: 0.8125rem;
  font-weight: 600;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  background: rgba(232, 163, 23, 0.1);
  color: var(--gold);
  border: 1px solid rgba(232, 163, 23, 0.18);
  margin-bottom: 2rem;
}
.blog-title {
  font-size: clamp(2.5rem, 5.5vw, 4rem);
  font-weight: 800;
  letter-spacing: -0.04em;
  line-height: 1.2;
  background: linear-gradient(135deg, var(--gold) 0%, var(--gold-light) 50%, var(--gold-glow) 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  margin: 0 0 1.5rem;
  padding: 0 1rem 0.1em;
}
.blog-subtitle {
  font-size: 1.25rem;
  color: var(--vp-c-text-2);
  max-width: 600px;
  margin: 0 auto 2rem;
  line-height: 1.7;
}
.blog-meta {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.875rem;
  font-size: 0.875rem;
  color: var(--vp-c-text-3);
}
.blog-meta-sep {
  width: 4px;
  height: 4px;
  border-radius: 50%;
  background: var(--vp-c-text-3);
  opacity: 0.4;
}

/* ── Body ── */
.blog-body {
  max-width: 960px;
  margin: 0 auto;
  padding: 0 2rem 6rem;
}
.blog-section {
  padding: 3rem 0;
  border-bottom: 1px solid var(--vp-c-divider);
}
.blog-section:last-child {
  border-bottom: none;
}
.blog-body h2 {
  font-size: 1.875rem;
  font-weight: 700;
  letter-spacing: -0.025em;
  margin: 0 0 1.25rem;
  color: var(--vp-c-text-1);
}
.blog-body h3 {
  font-size: 1.125rem;
  font-weight: 600;
  margin: 1.75rem 0 0.75rem;
  color: var(--vp-c-text-1);
}
.blog-body p {
  margin: 1rem 0;
  color: var(--vp-c-text-2);
  line-height: 1.8;
  font-size: 1.0625rem;
}
.blog-body a {
  color: var(--gold);
  text-decoration: none;
  font-weight: 500;
}
.blog-body a:hover {
  color: var(--gold-light);
}
.blog-body table {
  width: 100%;
  border-collapse: collapse;
  margin: 1.25rem 0;
  font-size: 0.9375rem;
}
.blog-body th {
  text-align: left;
  padding: 0.75rem 1.25rem;
  border-bottom: 2px solid var(--vp-c-divider);
  font-weight: 600;
  color: var(--vp-c-text-1);
  font-size: 0.8125rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}
.blog-body td {
  padding: 0.625rem 1.25rem;
  border-bottom: 1px solid var(--vp-c-divider);
  color: var(--vp-c-text-2);
}

/* ── Cards ── */
.blog-card {
  border: 1px solid var(--vp-c-divider);
  border-radius: 14px;
  overflow: hidden;
  margin: 1.5rem 0;
  background: var(--vp-c-bg-soft);
}
.blog-card-header {
  padding: 0.875rem 1.5rem;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--vp-c-text-2);
  border-bottom: 1px solid var(--vp-c-divider);
  background: var(--vp-c-bg-alt);
  letter-spacing: 0.01em;
}
.blog-card table { margin: 0; }
.blog-card th:first-child,
.blog-card td:first-child { padding-left: 1.5rem; }
.blog-card div[class*="language-"] {
  margin: 0 !important;
  border-radius: 0 !important;
}

/* ── Principle Grid ── */
.principle-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1px;
  background: var(--vp-c-divider);
  border: 1px solid var(--vp-c-divider);
  border-radius: 14px;
  overflow: hidden;
  margin: 1.5rem 0;
}
.principle-item {
  display: flex;
  gap: 1.125rem;
  padding: 1.5rem;
  background: var(--vp-c-bg-soft);
}
.principle-num {
  font-size: 0.8125rem;
  font-weight: 700;
  color: var(--gold);
  font-family: 'JetBrains Mono', monospace;
  line-height: 1.6;
  opacity: 0.6;
}
.principle-title {
  font-size: 1rem;
  font-weight: 600;
  color: var(--vp-c-text-1);
  margin-bottom: 0.375rem;
}
.principle-desc {
  font-size: 0.9375rem;
  color: var(--vp-c-text-3);
  line-height: 1.6;
}

/* ── Component Pills ── */
.component-pills {
  display: flex;
  flex-wrap: wrap;
  gap: 0.625rem;
  margin: 1.5rem 0;
}
.component-pills .pill {
  padding: 0.5rem 1.125rem;
  border-radius: 999px;
  font-size: 0.9375rem;
  font-weight: 500;
  font-family: 'JetBrains Mono', monospace;
  background: var(--vp-c-bg-soft);
  border: 1px solid var(--vp-c-divider);
  color: var(--vp-c-text-2);
  transition: border-color 0.15s, color 0.15s;
}
.component-pills .pill:hover {
  border-color: var(--gold);
  color: var(--gold);
}

/* ── Transport Grid ── */
.transport-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
  margin: 1.5rem 0;
}
.transport-card {
  padding: 1.5rem;
  border: 1px solid var(--vp-c-divider);
  border-radius: 12px;
  background: var(--vp-c-bg-soft);
}
.transport-card.transport-active {
  border-color: rgba(232, 163, 23, 0.3);
}
.transport-label {
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--vp-c-text-3);
  margin-bottom: 0.5rem;
}
.transport-active .transport-label { color: var(--gold); }
.transport-name {
  font-size: 1.125rem;
  font-weight: 700;
  color: var(--vp-c-text-1);
  margin-bottom: 0.25rem;
}
.transport-path {
  font-size: 0.8125rem;
  color: var(--vp-c-text-3);
  margin-bottom: 0.625rem;
}
.transport-desc {
  font-size: 0.9375rem;
  color: var(--vp-c-text-3);
  line-height: 1.6;
}

/* ── Lifecycle ── */
.lifecycle-steps {
  display: flex;
  gap: 1rem;
  margin: 1.5rem 0;
}
.lifecycle-step {
  flex: 1;
  padding: 1.5rem;
  border: 1px solid var(--vp-c-divider);
  border-radius: 12px;
  background: var(--vp-c-bg-soft);
}
.lifecycle-phase {
  font-size: 1rem;
  font-weight: 700;
  color: var(--vp-c-text-1);
  margin-bottom: 0.375rem;
}
.lifecycle-prefix {
  font-size: 0.8125rem;
  color: var(--gold);
  margin-bottom: 0.625rem;
}
.lifecycle-desc {
  font-size: 0.9375rem;
  color: var(--vp-c-text-3);
  line-height: 1.6;
}

/* ── CTA ── */
.blog-cta {
  text-align: center;
  padding: 4rem 0 2rem;
}
.blog-cta-btn {
  display: inline-block;
  padding: 0.875rem 2.5rem;
  border-radius: 12px;
  font-weight: 700;
  font-size: 1rem;
  color: #1a1a2e !important;
  background: linear-gradient(135deg, var(--gold) 0%, var(--gold-light) 100%);
  text-decoration: none !important;
  transition: opacity 0.15s, transform 0.15s;
}
.blog-cta-btn:hover {
  opacity: 0.9;
  transform: translateY(-1px);
  color: #1a1a2e !important;
}

/* ── Overflow ── */
.blog-page {
  overflow-x: hidden;
}
.blog-body div[class*="language-"] {
  overflow-x: auto;
}
.blog-body table {
  display: block;
  overflow-x: auto;
}
.blog-card {
  overflow-x: auto;
}

/* ── Responsive ── */
@media (max-width: 768px) {
  .blog-hero { padding: 4rem 1.5rem 3rem; }
  .blog-title { font-size: 2rem; }
  .blog-subtitle { font-size: 1.0625rem; }
  .blog-body { padding: 0 1.25rem 4rem; }
  .blog-body p { font-size: 1rem; }
  .principle-grid { grid-template-columns: 1fr; }
  .transport-grid { grid-template-columns: 1fr; }
  .lifecycle-steps { flex-direction: column; }
}
</style>
