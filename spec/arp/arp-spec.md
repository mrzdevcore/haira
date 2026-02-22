# Agentic Rendering Protocol (ARP)

**Version:** 0.1 Draft
**Date:** 2026-02-22
**Status:** Draft Specification
**License:** CC BY 4.0

---

## Abstract

The Agentic Rendering Protocol (ARP) is a transport-agnostic protocol for bidirectional communication between AI agents and rendering surfaces. It decouples agent logic from presentation, enabling a single agent to render interactive UI on web applications, desktop clients, mobile apps, CLI terminals, or any conforming renderer — without modification.

ARP draws architectural inspiration from the Wayland display protocol: agents are clients, renderers are compositors, communication is asynchronous and object-oriented, capabilities are negotiated at connection time, and agents are fully isolated from each other.

---

## Table of Contents

1. [Introduction](#1-introduction)
2. [Architecture](#2-architecture)
3. [Object Model](#3-object-model)
4. [Message Format](#4-message-format)
5. [Protocol Flow](#5-protocol-flow)
6. [Surface Lifecycle](#6-surface-lifecycle)
7. [Rendering](#7-rendering)
8. [Input](#8-input)
9. [Sessions and State](#9-sessions-and-state)
10. [Component Model](#10-component-model)
11. [Extension Lifecycle](#11-extension-lifecycle)
12. [Security Model](#12-security-model)
13. [Transport Bindings](#13-transport-bindings)
14. [Schema Definition Format](#14-schema-definition-format)
15. [Haira Integration](#15-haira-integration)
16. [Conformance Levels](#16-conformance-levels)
17. [Appendix A: Core Component Catalog](#appendix-a-core-component-catalog)
18. [Appendix B: Wire Examples](#appendix-b-wire-examples)

---

## 1. Introduction

### 1.1 Problem

AI agents produce structured output — tables, forms, charts, status messages, code blocks — but the rendering of that output is tightly coupled to the transport and frontend framework. An agent built for a web chat cannot render on a desktop app. An agent behind an SSE stream cannot serve a mobile client expecting a different protocol. Every new rendering surface requires custom integration code.

This is the same problem the display server world faced before Wayland: applications were coupled to X11, which was coupled to specific rendering paths, which made cross-platform display impossible without layers of compatibility shims.

### 1.2 Solution

ARP defines a **protocol** — not a framework, not a library, not a specification for a particular rendering technology. Like Wayland, it specifies:

- **How agents and renderers discover each other** (capability negotiation)
- **How agents describe what to display** (render commands with typed components)
- **How renderers send user input back** (structured input events with serial validation)
- **How both sides agree on component vocabulary** (extension lifecycle)
- **How agents are isolated from each other** (security model)

The protocol is transport-agnostic. It defines logical messages. Transport bindings (WebSocket, gRPC, Unix socket, stdio, SSE+HTTP) define how those messages are encoded and framed on the wire.

### 1.3 Design Principles

These principles are drawn from Wayland's successes and lessons from its adoption:

1. **Agent-renderer, no middleman.** Like Wayland eliminated the X server middleman, ARP has no intermediate framework between agent and renderer. The agent sends render commands directly; the renderer sends input directly.

2. **Asynchronous and non-blocking.** All messages are asynchronous. Neither side blocks waiting for a response. This prevents the latency problems that plagued X11's synchronous request/reply model.

3. **Every frame is perfect.** Borrowed directly from Wayland's atomic commit model. Render state accumulates in a pending buffer and is applied atomically on `COMMIT`. Users never see partial or flickering UI states.

4. **Capability-driven.** Renderers declare what they support at connection time. Agents adapt. A CLI renderer supports `table` and `text`; a web renderer supports everything. The protocol does not assume a rich renderer.

5. **Agent isolation.** Agents cannot see, modify, or intercept other agents' surfaces or input. This is the security model from Wayland, applied to the agentic context.

6. **Ship replacements with restrictions.** The key lesson from Wayland's painful adoption: if the protocol restricts a capability (e.g., no cross-agent surface access), the replacement mechanism (e.g., surface transfer via the renderer) MUST be defined in the same spec version. No gaps.

7. **Transport-agnostic.** The protocol defines logical messages. Transport bindings are separate specifications. This avoids X11's mistake of baking network transparency into the lowest level of the protocol.

8. **Typed components, not arbitrary markup.** Agents emit typed component descriptors (like `table` with `headers` and `rows`), not HTML or framework-specific templates. This enables renderers on any platform to interpret the intent.

### 1.4 Terminology

| Term | Definition |
|------|-----------|
| **Agent** | A program or AI model that produces output for rendering and consumes user input. Analogous to a Wayland client. |
| **Renderer** | A program that displays agent output on a surface (screen, terminal, speaker) and captures user input. Analogous to a Wayland compositor. |
| **Surface** | A logical rendering area owned by an agent. Analogous to a Wayland surface. |
| **Component** | A typed UI primitive (table, form, chart) with a defined props schema. |
| **Session** | A stateful conversation between a user and one or more agents, mediated by a renderer. |
| **Serial** | A monotonically increasing integer assigned to input events, used for validation and ordering. |
| **Object** | A protocol-level entity identified by a numeric ID and conforming to an interface definition. |
| **Interface** | A named set of requests and events that an object supports. |
| **Request** | A message from agent to renderer (analogous to Wayland requests). |
| **Event** | A message from renderer to agent (analogous to Wayland events). |

### 1.5 Notation

- `UPPER_CASE` denotes protocol message types (e.g., `RENDER`, `INPUT`, `COMMIT`).
- `snake_case` denotes field names (e.g., `surface_id`, `component_type`).
- `PascalCase` denotes interface names (e.g., `ArpDisplay`, `ArpSurface`).
- `?` after a type denotes an optional field (e.g., `string?`).
- `[]` denotes an array type (e.g., `Component[]`).

---

## 2. Architecture

### 2.1 Overview

```
                    ARP Protocol

  ┌─────────┐      Messages       ┌───────────┐
  │  Agent   │ ◄────────────────► │  Renderer  │
  │ (client) │   Requests ──►    │(compositor)│
  │          │   ◄── Events      │            │
  └─────────┘                    └───────────┘
       │                              │
  Owns surfaces              Owns display output
  Owns UI state              Owns input routing
  Decides what to show       Decides how to show it
  Produces components        Consumes components
```

### 2.2 Role Separation

**The agent decides WHAT to display.** It creates surfaces, emits render commands with typed components, and commits state changes. It does not know or control where on screen its surface appears, what visual theme is applied, or how the component is rendered (HTML table vs. ASCII table vs. voice readout).

**The renderer decides HOW to display it.** It arranges surfaces, applies themes, translates components into platform-native rendering, captures input events, and routes them to the appropriate agent. It does not modify the agent's state or inject content into the agent's surfaces.

This separation is identical to Wayland's client-compositor model and is the foundation of ARP's portability.

### 2.3 Multi-Agent Architecture

Multiple agents can connect to the same renderer. Each agent owns its own surfaces and receives only its own input events. The renderer arbitrates surface arrangement and input routing (the "focus" problem).

```
  Agent A ──┐
             ├──► Renderer ──► Display
  Agent B ──┘         │
                      ▼
                 Input routing
                 (focus management)
```

When agents need to hand off rendering responsibility (e.g., agent handoffs in Haira), the `SURFACE_TRANSFER` mechanism provides a renderer-mediated transfer. Agents cannot directly communicate with each other through the protocol.

### 2.4 Comparison with Wayland

| Wayland | ARP | Notes |
|---------|-----|-------|
| `wl_display` | `ArpDisplay` | Bootstrap object, implicit at connection |
| `wl_registry` | `ArpRegistry` | Capability advertisement |
| `wl_surface` | `ArpSurface` | Agent's rendering area |
| `wl_buffer` | Component tree | What gets displayed |
| `wl_seat` | `ArpInput` | Input abstraction |
| `wl_compositor` | `ArpCompositor` | Surface arrangement |
| `xdg_shell` | `ArpShell` | Surface roles and lifecycle |
| `commit` | `COMMIT` | Atomic state application |
| `damage` | `PATCH` | Incremental updates |
| `frame` callback | `FRAME_ACK` | Render pacing / backpressure |
| Protocol extensions | Component extensions | Three-phase lifecycle |
| `wayland-scanner` | `arp-scanner` | Schema → code generation |

---

## 3. Object Model

### 3.1 Objects

Every entity in ARP is an **object** with:

- A **numeric ID** (uint32): uniquely identifies the object within a connection
- An **interface**: defines what requests and events the object supports
- A **version**: the negotiated version of the interface

Objects are created through requests on existing objects (the "factory" pattern from Wayland). Object IDs are allocated by the requesting side:

- **Agent-allocated IDs**: `[1, 0x7FFFFFFF]`
- **Renderer-allocated IDs**: `[0x80000000, 0xFFFFFFFF]`
- **ID 0**: Reserved for `ArpDisplay` (implicit, like Wayland's `wl_display`)

### 3.2 Bootstrap Sequence

The bootstrap follows Wayland's elegant pattern:

1. Agent connects to renderer (transport-specific).
2. Object ID 0 is implicitly `ArpDisplay` — no negotiation needed.
3. Agent calls `ArpDisplay.get_registry()`, which creates an `ArpRegistry` object.
4. Renderer sends `ArpRegistry.global` events for each capability it supports.
5. Agent calls `ArpRegistry.bind()` to obtain typed objects for capabilities it needs.

This is identical to Wayland's `wl_display` → `wl_registry` → `global` → `bind` flow.

### 3.3 Core Interfaces

```
ArpDisplay (object 0, implicit)
├── get_registry() → ArpRegistry
├── sync() → ArpCallback
└── error event

ArpRegistry
├── global event (name, interface, version)
├── global_remove event (name)
└── bind(name, interface, version) → object

ArpCompositor
└── create_surface() → ArpSurface

ArpSurface
├── render(components[])
├── patch(ops[])
├── commit()
├── destroy()
├── set_title(title)
├── set_role(role)
├── frame_ack event (seq)
├── configure event (viewport)
└── closed event

ArpInput
├── input event (surface_id, source, serial, type, data)
├── focus event (surface_id)
├── unfocus event (surface_id)
└── ack(serial)

ArpSession
├── create(session_id, metadata)
├── state_update(state)
├── destroy()
└── resumed event (state)

ArpShell
├── get_toplevel(surface) → ArpToplevel
├── get_popup(surface, parent) → ArpPopup
└── get_overlay(surface) → ArpOverlay

ArpToplevel
├── set_title(title)
├── configure event (width, height, states[])
└── close event

ArpPopup
├── configure event (x, y, width, height)
└── popup_done event

ArpOverlay
├── set_anchor(surface_id, position)
├── configure event (bounds)
└── dismissed event
```

### 3.4 Object Lifecycle

Objects follow Wayland's cooperative destruction model:

1. One side sends a destroy request/event.
2. The other side acknowledges by ceasing to reference the object.
3. Both sides release the object ID for potential reuse.

Due to asynchronous message ordering, both sides MUST handle messages targeting recently-destroyed objects gracefully (by discarding them), until the destruction is fully synchronized.

### 3.5 Versioning

Each interface has a version number. When an agent binds to a global, it specifies the version it supports (which may be lower than what the renderer advertises). Both sides then use only the requests/events defined up to that version.

New requests and events are appended to the end of an interface definition, never inserted. This ensures backward compatibility — a version-2 agent talking to a version-1 renderer simply doesn't use version-2 features.

---

## 4. Message Format

### 4.1 Logical Message Envelope

ARP defines a logical message format independent of transport encoding. Every message contains:

```
{
  "v":       uint,      // Protocol version (currently 1)
  "id":      uint,      // Target object ID
  "op":      string,    // Operation name
  "seq":     uint,      // Monotonically increasing sequence number
  "payload": object     // Operation-specific data
}
```

| Field | Type | Description |
|-------|------|-------------|
| `v` | uint | Protocol version. Allows future incompatible changes. |
| `id` | uint | The object this message targets. |
| `op` | string | The operation (request or event name). |
| `seq` | uint | Sequence number. Sender-assigned, monotonically increasing. Used for ordering, acknowledgment, and serial validation. |
| `payload` | object | Operation-specific data. Schema defined by the interface + operation. |

### 4.2 Encoding

The logical message format is **encoding-agnostic**. Transport bindings choose the encoding:

| Encoding | Use Case | MIME Type |
|----------|----------|-----------|
| JSON | Web, debugging, interop | `application/json` |
| MessagePack | High-throughput native | `application/msgpack` |
| CBOR | Constrained environments | `application/cbor` |
| Protobuf | gRPC transport binding | `application/protobuf` |

All transport bindings MUST support JSON encoding. Other encodings are OPTIONAL.

### 4.3 Primitive Types

| Type | Description | JSON Encoding |
|------|-------------|---------------|
| `uint` | Unsigned 32-bit integer | number |
| `int` | Signed 32-bit integer | number |
| `float` | 64-bit IEEE 754 | number |
| `bool` | Boolean | true/false |
| `string` | UTF-8 text | string |
| `bytes` | Binary data | base64 string |
| `object_id` | Reference to a protocol object | number |
| `any` | Arbitrary JSON-compatible value | any JSON value |
| `T[]` | Array of T | array |
| `T?` | Optional (nullable) T | value or null |

### 4.4 Sequence Numbers

Every message carries a `seq` field:

- **Agent sequences** start at 1 and increment for each message the agent sends.
- **Renderer sequences** start at 1 and increment for each message the renderer sends.
- Sequences are per-connection, not per-object.
- Sequences MUST be monotonically increasing. A receiver that observes a non-increasing sequence SHOULD close the connection with an error.

Sequence numbers serve three purposes:

1. **Ordering**: Detect out-of-order delivery on transports that don't guarantee ordering.
2. **Acknowledgment**: `FRAME_ACK` references the `seq` of the `COMMIT` it acknowledges.
3. **Serial validation**: Input events carry their `seq` as a serial. Actions that require input context (popup creation, drag-and-drop) must reference a valid, recent serial.

---

## 5. Protocol Flow

### 5.1 Connection Lifecycle

```
Agent                                  Renderer
  │                                       │
  │◄──────── [transport connect] ────────►│
  │                                       │
  │── get_registry (id=0) ──────────────►│
  │                                       │
  │◄──── global("ArpCompositor", v1) ─────│
  │◄──── global("ArpInput", v1) ──────────│
  │◄──── global("ArpShell", v1) ──────────│
  │◄──── global("ArpSession", v1) ────────│
  │◄──── global("component:table", v1) ───│
  │◄──── global("component:form", v1) ────│
  │◄──── global("component:chart", v1) ───│
  │◄──── global("feature:streaming") ─────│
  │◄──── global("feature:multi-surface") ─│
  │                                       │
  │── bind("ArpCompositor", v1) ────────►│
  │── bind("ArpInput", v1) ─────────────►│
  │── bind("ArpShell", v1) ─────────────►│
  │── bind("ArpSession", v1) ───────────►│
  │                                       │
  │        [ready to render]              │
```

### 5.2 Component Capability Advertisement

Components are advertised as globals with the `component:` prefix. This tells the agent exactly which component types the renderer supports:

```
global("component:table", version=1)
global("component:form", version=1)
global("component:chart", version=2)
global("component:status-card", version=1)
global("component:code-block", version=1)
```

An agent that needs `chart` but the renderer only advertises `table` can either:

1. Fall back to a `table` representation of the chart data.
2. Use the `fallback` field in the render command (see Section 7).
3. Decline to use chart rendering for this connection.

Feature capabilities use the `feature:` prefix:

```
global("feature:streaming")        // Renderer supports streaming updates
global("feature:multi-surface")    // Renderer supports multiple surfaces
global("feature:theming")          // Renderer supports theme negotiation
global("feature:voice-input")      // Renderer supports voice input
global("feature:file-upload")      // Renderer supports file uploads
```

### 5.3 Synchronization

Like Wayland's `wl_display.sync`, the agent can request a synchronization point:

```
Agent → Renderer:
  { id: 0, op: "sync", seq: 100, payload: { callback_id: 42 } }

Renderer → Agent (after processing all prior messages):
  { id: 42, op: "done", seq: 55, payload: {} }
```

The `done` event on the callback object confirms that all agent messages up to `seq: 100` have been processed. This is the roundtrip mechanism for cases where the agent needs to ensure prior state is applied before continuing.

---

## 6. Surface Lifecycle

### 6.1 Creating a Surface

```
Agent → Renderer:
  {
    id: <compositor_id>,
    op: "create_surface",
    seq: 101,
    payload: {
      new_id: 10
    }
  }
```

This creates an `ArpSurface` object with ID 10. The surface has no role yet — it is inert until assigned a role via `ArpShell`.

### 6.2 Surface Roles

Like Wayland's xdg-shell, surfaces must be assigned a role before they can display content. ARP defines three roles:

| Role | Description | Wayland Analog |
|------|-------------|----------------|
| **Toplevel** | Primary application surface. Full rendering area. | `xdg_toplevel` |
| **Popup** | Contextual surface anchored to a parent (dropdown, tooltip). | `xdg_popup` |
| **Overlay** | Floating surface positioned relative to another surface (notification, toast). | N/A (ARP addition) |

```
Agent → Renderer:
  {
    id: <shell_id>,
    op: "get_toplevel",
    seq: 102,
    payload: {
      surface_id: 10,
      new_id: 11
    }
  }

Renderer → Agent:
  {
    id: 11,
    op: "configure",
    seq: 50,
    payload: {
      width: 800,
      height: 600,
      states: ["activated"]
    }
  }
```

The renderer sends a `configure` event telling the agent the available viewport size and current states. The agent MUST respond with a `COMMIT` to acknowledge the configuration (even if it has no content yet).

### 6.3 Surface States

| State | Description |
|-------|-------------|
| `activated` | Surface has focus and receives input |
| `suspended` | Surface is not visible (tab in background, minimized) |
| `resizing` | Renderer is resizing the surface |
| `fullscreen` | Surface occupies the full viewport |

Agents SHOULD respect `suspended` by reducing rendering activity (no streaming updates, no animations). This is analogous to Wayland's frame callback mechanism — a suspended surface receives no frame callbacks.

### 6.4 Destroying a Surface

```
Agent → Renderer:
  {
    id: 10,
    op: "destroy",
    seq: 150,
    payload: {}
  }
```

The renderer releases the surface and any associated role objects. Input events for the destroyed surface are discarded.

### 6.5 Surface Transfer

When one agent hands off to another (e.g., agent handoff in Haira), the surface can be transferred:

```
Agent A → Renderer:
  {
    id: 10,
    op: "transfer",
    seq: 200,
    payload: {
      to_session: "agent_b_session",
      retain_content: true
    }
  }
```

The renderer mediates the transfer:

1. Notifies Agent A with a `transferred` event (surface is no longer theirs).
2. Sends Agent B a `surface_acquired` event with the surface's current content (if `retain_content` is true) and a new object ID.
3. Agent B can now render to and commit on the transferred surface.

This prevents agents from needing direct communication channels while supporting seamless visual handoffs.

---

## 7. Rendering

### 7.1 The Render-Commit Cycle

Rendering follows Wayland's double-buffered commit model:

1. Agent sends one or more `render` or `patch` requests (pending state).
2. Agent sends `commit` (atomically applies pending state).
3. Renderer processes the commit and displays the result.
4. Renderer sends `frame_ack` (agent can prepare next frame).

```
Agent                                  Renderer
  │                                       │
  │── render(components) ──────────────►│  (pending)
  │── render(more components) ─────────►│  (pending)
  │── commit ──────────────────────────►│  [atomic apply]
  │                                       │  [display]
  │◄──────────── frame_ack ──────────────│
  │                                       │
  │── patch(updates) ──────────────────►│  (pending)
  │── commit ──────────────────────────►│  [atomic apply]
  │◄──────────── frame_ack ──────────────│
```

**Why this matters:** The user never sees a half-rendered table missing its last rows, or a form without its submit button. Every commit produces a complete, consistent view. This is Wayland's "every frame is perfect" guarantee applied to agentic UI.

### 7.2 Render Request

Sets the complete component tree for a surface:

```json
{
  "id": 10,
  "op": "render",
  "seq": 103,
  "payload": {
    "components": [
      {
        "id": "comp_1",
        "type": "table",
        "version": 1,
        "props": {
          "title": "Search Results",
          "headers": ["Name", "Price", "Rating"],
          "rows": [
            ["Widget A", "$29.99", "4.5"],
            ["Widget B", "$39.99", "4.8"]
          ]
        },
        "fallback": {
          "type": "text",
          "props": {
            "content": "Search Results:\n- Widget A: $29.99 (4.5)\n- Widget B: $39.99 (4.8)"
          }
        }
      }
    ]
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `components` | `Component[]` | Yes | The component tree to display |

Each component contains:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | Yes | Unique identifier within the surface, for targeting in patches and input events |
| `type` | `string` | Yes | Component type name (must match an advertised `component:*` global) |
| `version` | `uint` | No | Component version (defaults to 1) |
| `props` | `object` | Yes | Component-specific properties |
| `fallback` | `Component?` | No | Fallback component if the renderer doesn't support this type |

### 7.3 Patch Request

Incrementally updates the current component tree without replacing it entirely. This is ARP's equivalent of Wayland's damage tracking — only describe what changed.

```json
{
  "id": 10,
  "op": "patch",
  "seq": 110,
  "payload": {
    "ops": [
      {
        "op": "update",
        "target": "comp_1",
        "path": "props.rows",
        "value": [
          ["Widget A", "$29.99", "4.5"],
          ["Widget B", "$39.99", "4.8"],
          ["Widget C", "$19.99", "4.2"]
        ]
      },
      {
        "op": "insert",
        "after": "comp_1",
        "component": {
          "id": "comp_2",
          "type": "status-card",
          "props": {
            "status": "success",
            "title": "Search Complete",
            "message": "Found 3 results"
          }
        }
      },
      {
        "op": "remove",
        "target": "comp_old"
      }
    ]
  }
}
```

Patch operations:

| Operation | Fields | Description |
|-----------|--------|-------------|
| `update` | `target`, `path`, `value` | Update a property within an existing component |
| `insert` | `after` (or `null` for prepend), `component` | Insert a new component into the tree |
| `remove` | `target` | Remove a component from the tree |
| `replace` | `target`, `component` | Replace a component entirely |
| `reorder` | `target`, `after` | Move a component to a new position |

### 7.4 Commit Request

Atomically applies all pending render/patch operations:

```json
{
  "id": 10,
  "op": "commit",
  "seq": 111,
  "payload": {}
}
```

A `commit` with no preceding `render` or `patch` is valid and acts as a configuration acknowledgment (e.g., responding to a `configure` event).

### 7.5 Frame Acknowledgment

The renderer sends `frame_ack` after processing a commit:

```json
{
  "id": 10,
  "op": "frame_ack",
  "seq": 60,
  "payload": {
    "commit_seq": 111
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `commit_seq` | uint | The `seq` of the `commit` being acknowledged |

**Backpressure:** An agent SHOULD NOT send more than one `commit` without receiving a `frame_ack` for the previous one. This prevents overwhelming the renderer (analogous to Wayland's frame callback mechanism). An agent MAY pipeline up to 2 commits for low-latency scenarios, but MUST NOT exceed the renderer's advertised `max_pending_commits` (default: 2).

### 7.6 Streaming Rendering

For long-running operations (e.g., an agent generating a table row by row), ARP supports **streaming mode**. The agent sends multiple `render`/`patch` + `commit` cycles, and the renderer displays each committed state incrementally.

The renderer advertises streaming support via `global("feature:streaming")`. When streaming:

1. Agent sends initial `render` + `commit` (partial content).
2. Agent sends `patch` + `commit` (append more content).
3. Repeat until complete.
4. Agent sends a final `commit` with a `final: true` flag.

```json
{
  "id": 10,
  "op": "commit",
  "seq": 150,
  "payload": {
    "final": true
  }
}
```

The `final` flag tells the renderer this surface's content is complete — no more updates are expected until the next user interaction.

---

## 8. Input

### 8.1 Input Model

Input in ARP follows Wayland's model: the renderer captures user interaction, wraps it in structured events, and sends it to the agent that owns the focused surface.

```
User interacts with renderer
        │
        ▼
  Renderer captures event
        │
        ▼
  Renderer determines target surface (focus)
        │
        ▼
  Renderer sends INPUT event to owning agent
        │
        ▼
  Agent processes input and optionally re-renders
```

### 8.2 Input Event

```json
{
  "id": "<input_object_id>",
  "op": "input",
  "seq": 80,
  "payload": {
    "surface_id": 10,
    "source_component": "comp_5",
    "serial": 80,
    "type": "form_submit",
    "data": {
      "name": "Alice",
      "email": "alice@example.com"
    }
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `surface_id` | uint | The surface the input targets |
| `source_component` | string? | The component ID that generated the input (null for surface-level input) |
| `serial` | uint | The renderer's `seq` for this event — used for serial validation |
| `type` | string | Input type (see table below) |
| `data` | object | Type-specific input data |

### 8.3 Input Types

| Type | Data Schema | Source Components | Description |
|------|-------------|-------------------|-------------|
| `form_submit` | `{ [field]: value }` | `form` | User submitted a form |
| `action` | `{ action: string, payload?: any }` | `confirm`, `choices`, buttons | User triggered a named action |
| `selection` | `{ selected: string[], source: string }` | `choices`, `table` | User selected items |
| `text` | `{ text: string }` | chat input, text fields | Free text input |
| `file` | `{ files: FileRef[] }` | file inputs | File upload |
| `gesture` | `{ gesture: string, meta: object }` | any (touch/mobile) | Touch gesture |
| `voice` | `{ transcript: string, confidence: float, language: string }` | voice input | Speech-to-text result |
| `navigate` | `{ route: string, params?: object }` | renderer navigation | User navigated to a route |
| `focus` | `{}` | (surface-level) | Surface gained focus |
| `unfocus` | `{}` | (surface-level) | Surface lost focus |
| `resize` | `{ width: uint, height: uint }` | (surface-level) | Viewport resized |

### 8.4 Input Acknowledgment

Agents MUST acknowledge input events:

```json
{
  "id": "<input_object_id>",
  "op": "ack",
  "seq": 120,
  "payload": {
    "serial": 80
  }
}
```

Acknowledgment serves two purposes:

1. **Flow control**: The renderer knows the agent received and processed the input.
2. **Serial validation**: The `serial` confirms which input event is being acknowledged.

### 8.5 Serial Validation

Borrowed from Wayland's input serial model. When an agent performs an action that requires user authorization context, it MUST reference a valid, recent input serial:

```json
{
  "id": "<shell_id>",
  "op": "get_popup",
  "seq": 121,
  "payload": {
    "surface_id": 10,
    "parent_id": 10,
    "new_id": 20,
    "trigger_serial": 80
  }
}
```

The renderer validates that:

1. The serial exists and was sent to this agent.
2. The serial is recent (not older than a renderer-defined threshold).
3. The serial's surface matches the request context.

If validation fails, the renderer rejects the request with an error event. This prevents agents from performing privileged UI actions (popups, overlays, fullscreen) without legitimate user interaction.

### 8.6 Focus

The renderer manages focus — which surface receives input events:

```json
{
  "id": "<input_object_id>",
  "op": "focus",
  "seq": 81,
  "payload": {
    "surface_id": 10
  }
}
```

An agent's surface MUST have focus to receive input events (except `focus` and `unfocus` events themselves). The renderer decides focus policy (click-to-focus, follow-cursor, etc.) — agents cannot request focus, only receive notification of focus changes.

---

## 9. Sessions and State

### 9.1 Sessions

A session represents an ongoing interaction between a user and an agent. Sessions persist across connection interruptions (user refreshes the page, app goes to background).

```json
{
  "id": "<session_object_id>",
  "op": "create",
  "seq": 105,
  "payload": {
    "session_id": "sess_abc123",
    "metadata": {
      "agent": "search-assistant",
      "user_id": "user_42",
      "started_at": "2026-02-22T10:30:00Z"
    }
  }
}
```

### 9.2 State

UI state is **agent-owned**. The agent is the source of truth for what the UI displays. The renderer renders state but never mutates it. User input flows back as events — the agent decides how to update state in response.

```json
{
  "id": "<session_object_id>",
  "op": "state_update",
  "seq": 130,
  "payload": {
    "state": {
      "step": "review",
      "cart_items": 3,
      "user_name": "Alice"
    }
  }
}
```

State updates are informational — they tell the renderer about application state for potential use in status bars, breadcrumbs, or context indicators. They do NOT directly drive rendering (that's the component tree's job via `render`/`commit`).

### 9.3 Session Resumption

When a renderer reconnects to an existing session (e.g., page refresh):

```json
{
  "id": "<session_object_id>",
  "op": "resumed",
  "seq": 82,
  "payload": {
    "session_id": "sess_abc123",
    "last_commit_seq": 150,
    "state": {
      "step": "review",
      "cart_items": 3
    }
  }
}
```

The renderer tells the agent the last commit it displayed. The agent can then:

1. Re-render from scratch (simple, always correct).
2. Send only patches since `last_commit_seq` (optimized, if the agent maintains a commit log).

---

## 10. Component Model

### 10.1 What Is a Component

A component is a **typed UI primitive** with:

- A **name** (e.g., `table`, `form`, `chart`)
- A **version** (integer, incremented for backward-compatible additions)
- A **props schema** (the data shape the component accepts)
- A **category** (`display`, `input`, `layout`)
- A **fallback** (what to degrade to if unsupported)

Components are NOT HTML elements, React components, or platform-specific widgets. They are **abstract descriptions of UI intent**. The renderer maps them to its platform's native rendering.

### 10.2 Component Categories

| Category | Purpose | Examples |
|----------|---------|----------|
| **Display** | Show information to the user | `text`, `table`, `chart`, `status-card`, `code-block`, `diff`, `key-value`, `progress`, `image`, `markdown` |
| **Input** | Collect information from the user | `form`, `confirm`, `choices`, `text-input` |
| **Layout** | Arrange other components | `group`, `tabs`, `split` |

### 10.3 Component Props

Every component has a props schema defined in the ARP schema definition format (Section 14). Props are:

- **Strongly typed**: Each field has a defined type.
- **Versioned**: New optional fields can be added in later versions.
- **Validatable**: The renderer can validate props against the schema before rendering.

Example props schema for `table`:

```yaml
component: table
version: 1
category: display
fallback: text

props:
  title:
    type: string
    required: false
    description: "Table heading"
  headers:
    type: string[]
    required: true
    description: "Column header labels"
  rows:
    type: string[][]
    required: true
    description: "Row data (each row is an array of cell values)"
  highlight:
    type: uint[]
    required: false
    description: "Row indices to visually highlight"
  searchable:
    type: bool
    required: false
    default: false
    description: "Whether to show a search/filter input"

input_types:
  - type: selection
    description: "Row selection"
    data_schema:
      selected:
        type: uint[]
        description: "Selected row indices"
```

### 10.4 Fallback Chain

Every component declares a fallback — a simpler component to use when the renderer doesn't support the original. Fallback chains terminate at `text`, which every conforming renderer MUST support:

```
chart → table → text
product-cards → table → text
diff → code-block → text
code-block → text
form → text (render as instructions)
confirm → text (render as question)
```

The `fallback` field in a render command provides pre-computed fallback props:

```json
{
  "id": "comp_1",
  "type": "chart",
  "props": {
    "type": "bar",
    "title": "Sales by Region",
    "labels": ["North", "South", "East", "West"],
    "datasets": [{ "label": "Q1", "data": [100, 200, 150, 300] }]
  },
  "fallback": {
    "type": "table",
    "props": {
      "title": "Sales by Region",
      "headers": ["Region", "Q1 Sales"],
      "rows": [["North", "100"], ["South", "200"], ["East", "150"], ["West", "300"]]
    }
  }
}
```

If the renderer doesn't support `chart`, it uses the `fallback`. If it also doesn't support `table`, it follows the table's declared fallback chain to `text`.

### 10.5 Component Composition

The `group` layout component nests other components:

```json
{
  "id": "comp_group_1",
  "type": "group",
  "props": {
    "direction": "vertical",
    "gap": "medium",
    "children": [
      { "id": "child_1", "type": "key-value", "props": { ... } },
      { "id": "child_2", "type": "table", "props": { ... } },
      { "id": "child_3", "type": "status-card", "props": { ... } }
    ]
  }
}
```

Nesting depth is limited to **3 levels** to prevent pathological recursive rendering. Renderers MUST reject `group` components nested deeper than 3.

---

## 11. Extension Lifecycle

### 11.1 Motivation

The core component catalog covers common UI patterns. Domain-specific needs (Gantt charts, Kanban boards, floor plans, 3D viewers) require extensions. ARP's extension lifecycle is modeled directly on Wayland's `wayland-protocols` three-phase process.

### 11.2 Three Phases

#### Phase 1: Experimental

- **Prefix**: `x-{vendor}-{name}` (e.g., `x-haira-kanban`, `x-acme-floorplan`)
- **Location**: Published by the vendor, not part of the ARP spec
- **Stability**: Breaking changes allowed between versions
- **Requirement**: None — anyone can create an experimental component
- **Advertisement**: `global("component:x-haira-kanban", version=1)`

Experimental components are the innovation space. Vendors create them to solve domain-specific problems. The vendor prefix prevents name collisions.

#### Phase 2: Staging

- **Prefix**: `s-{name}` (e.g., `s-calendar`, `s-kanban`)
- **Location**: ARP staging registry (official, reviewed)
- **Stability**: No backward-incompatible changes, but can be superseded by a new version
- **Requirement**: At least 2 renderer implementations (e.g., web + CLI, or web + desktop)
- **Advertisement**: `global("component:s-calendar", version=1)`
- **Governance**: Requires review by ARP maintainers

Staging is the proving ground. Components here are stable enough for production use but haven't yet achieved universal adoption.

#### Phase 3: Core

- **Prefix**: None (e.g., `table`, `form`, `chart`)
- **Location**: ARP core spec (this document, Appendix A)
- **Stability**: Only additive, backward-compatible changes
- **Requirement**: Widely implemented, considered universal
- **Advertisement**: `global("component:table", version=1)`
- **Governance**: Part of the ARP spec itself

Core components are the baseline. All conforming renderers at the "Standard" level (see Section 16) MUST support all core components.

### 11.3 Promotion

```
Experimental (x-vendor-name)
    │
    │  2+ renderer implementations
    │  Governance review
    ▼
Staging (s-name)
    │
    │  Widespread adoption
    │  Spec inclusion vote
    ▼
Core (name)
```

Demotion is also possible: a staging component that proves flawed can be superseded by a new staging component under a different name. The old one is marked deprecated but remains supported for a defined sunset period.

### 11.4 Versioning Within Extensions

Like Wayland interfaces, component versions are additive:

- **Version 1**: `table` with `title`, `headers`, `rows`
- **Version 2**: Adds `searchable` (optional bool), `sortable` (optional bool)
- **Version 3**: Adds `pagination` (optional object)

New fields are always optional. A version-1 renderer ignores version-2 fields. A version-2 agent sending to a version-1 renderer simply doesn't get search/sort features — but the table still renders.

---

## 12. Security Model

### 12.1 Principles

ARP's security model is adapted from Wayland's:

1. **Full agent isolation**: Agent A cannot see, modify, or intercept Agent B's surfaces or input.
2. **No input snooping**: Input events route only to the agent owning the focused surface.
3. **No surface enumeration**: An agent cannot list or discover other agents' surfaces.
4. **Serial-validated actions**: Privileged operations (popups, overlays) require a recent, valid input serial.
5. **Renderer-mediated transfers**: Agents cannot directly communicate. Surface transfer goes through the renderer.

### 12.2 Threat Model

| Threat | Mitigation |
|--------|-----------|
| Malicious agent captures input from other agents | Input events are scoped to the owning agent's surfaces |
| Agent spawns unwanted popups/overlays | Serial validation requires genuine user interaction |
| Agent reads other agents' rendered content | No cross-agent surface access in the protocol |
| Agent impersonates another agent | Session IDs are renderer-assigned, not agent-chosen |
| Agent DoS via rapid rendering | `frame_ack` backpressure + `max_pending_commits` limit |
| Agent sends oversized component trees | Renderers enforce `max_components_per_commit` limit |

### 12.3 Renderer Limits

Renderers SHOULD enforce resource limits and advertise them via globals:

```
global("limit:max_surfaces", value=10)
global("limit:max_components_per_commit", value=100)
global("limit:max_pending_commits", value=2)
global("limit:max_patch_ops", value=50)
global("limit:max_component_depth", value=3)
```

Agents that exceed limits receive an error event. Repeated violations MAY result in connection termination.

---

## 13. Transport Bindings

### 13.1 Overview

ARP is transport-agnostic. Each transport binding is a separate document that specifies:

1. **Connection establishment**: How agent and renderer connect.
2. **Message framing**: How ARP messages are delimited on the wire.
3. **Encoding**: Which encodings are supported (JSON is mandatory).
4. **Lifecycle**: Connection teardown, reconnection, keepalive.

### 13.2 WebSocket Binding

**Target**: Web applications, browser-based renderers.

```
Connection:
  Agent connects via WebSocket to renderer URL
  ws://{host}:{port}/arp/v1

Framing:
  Each WebSocket message = one ARP message
  Text frames for JSON encoding
  Binary frames for MessagePack/CBOR encoding

Encoding:
  Default: JSON
  Negotiated via Sec-WebSocket-Protocol header:
    "arp.v1.json" (default)
    "arp.v1.msgpack"
    "arp.v1.cbor"

Keepalive:
  WebSocket ping/pong (standard)

Reconnection:
  Client reconnects and sends session resumption
  via ArpSession
```

### 13.3 gRPC Binding

**Target**: Native applications, microservices, high-performance scenarios.

```
Service definition:
  service ArpService {
    rpc Connect(stream ArpMessage) returns (stream ArpMessage);
  }

Encoding:
  Protobuf (ArpMessage wraps the logical envelope)

Connection:
  Standard gRPC channel establishment
  TLS recommended for production

Streaming:
  Bidirectional streaming RPC
  Each stream message = one ARP message
```

### 13.4 Unix Socket Binding

**Target**: Same-machine communication, highest performance.

```
Connection:
  Socket path: $XDG_RUNTIME_DIR/arp-{session_id}
  Or: specified via ARP_SOCKET environment variable

Framing:
  Length-prefixed messages:
    [4 bytes: message length (uint32 big-endian)]
    [N bytes: message payload]

Encoding:
  Default: MessagePack (for performance)
  JSON available via ARP_ENCODING=json

File descriptors:
  Passed via SCM_RIGHTS ancillary data
  (for file uploads, shared memory buffers)
```

### 13.5 stdio Binding

**Target**: CLI renderers, child process communication, pipes.

```
Connection:
  Agent writes to stdout, reads from stdin
  (or: renderer writes to stdout, reads from stdin)

Framing:
  Newline-delimited JSON (NDJSON)
  Each line = one ARP message as JSON

Encoding:
  JSON only (human-readable for debugging)

Lifecycle:
  EOF on stdin = connection closed
  Process exit = connection closed
```

### 13.6 SSE + HTTP POST Binding

**Target**: Serverless environments, constrained clients, backward compatibility with existing Haira SSE.

```
Renderer → Agent events:
  POST /arp/v1/input
  Body: ARP message as JSON
  Content-Type: application/json

Agent → Renderer events:
  GET /arp/v1/events
  Response: text/event-stream (SSE)
  Each SSE "data:" line = one ARP message as JSON

Session binding:
  Cookie or Authorization header links HTTP requests to session

Note: This binding does not support the full object model.
It operates in "simplified mode" (see Section 16,
Conformance Level: Minimal).
```

---

## 14. Schema Definition Format

### 14.1 Purpose

ARP protocols and components are defined in YAML schema files. The `arp-scanner` tool processes these files and generates language-specific code — analogous to Wayland's `wayland-scanner` processing XML protocol definitions.

### 14.2 Protocol Schema

```yaml
# arp-core.protocol.yaml
protocol: arp-core
version: 1

interfaces:
  ArpDisplay:
    version: 1
    requests:
      get_registry:
        args:
          new_id:
            type: new_id
            interface: ArpRegistry
      sync:
        args:
          callback_id:
            type: new_id
            interface: ArpCallback
    events:
      error:
        args:
          object_id:
            type: object_id
          code:
            type: uint
          message:
            type: string

  ArpRegistry:
    version: 1
    requests:
      bind:
        args:
          name:
            type: uint
          interface:
            type: string
          version:
            type: uint
          new_id:
            type: new_id
    events:
      global:
        args:
          name:
            type: uint
          interface:
            type: string
          version:
            type: uint
      global_remove:
        args:
          name:
            type: uint

  ArpSurface:
    version: 1
    requests:
      render:
        args:
          components:
            type: Component[]
      patch:
        args:
          ops:
            type: PatchOp[]
      commit:
        args:
          final:
            type: bool
            default: false
      destroy: {}
      set_title:
        args:
          title:
            type: string
      transfer:
        args:
          to_session:
            type: string
          retain_content:
            type: bool
    events:
      frame_ack:
        args:
          commit_seq:
            type: uint
      configure:
        args:
          width:
            type: uint
          height:
            type: uint
          states:
            type: string[]
      closed: {}
      transferred: {}
      surface_acquired:
        args:
          components:
            type: Component[]?

  ArpInput:
    version: 1
    requests:
      ack:
        args:
          serial:
            type: uint
    events:
      input:
        args:
          surface_id:
            type: object_id
          source_component:
            type: string?
          serial:
            type: uint
          type:
            type: string
          data:
            type: any
      focus:
        args:
          surface_id:
            type: object_id
      unfocus:
        args:
          surface_id:
            type: object_id

  ArpSession:
    version: 1
    requests:
      create:
        args:
          session_id:
            type: string
          metadata:
            type: any
      state_update:
        args:
          state:
            type: any
      destroy: {}
    events:
      resumed:
        args:
          session_id:
            type: string
          last_commit_seq:
            type: uint
          state:
            type: any

types:
  Component:
    fields:
      id:
        type: string
      type:
        type: string
      version:
        type: uint
        default: 1
      props:
        type: any
      fallback:
        type: Component?

  PatchOp:
    fields:
      op:
        type: string
        enum: [update, insert, remove, replace, reorder]
      target:
        type: string?
      path:
        type: string?
      value:
        type: any?
      after:
        type: string?
      component:
        type: Component?
```

### 14.3 Component Schema

```yaml
# arp-core.components.yaml
components:

  text:
    version: 1
    category: display
    fallback: null  # text is the terminal fallback
    description: "Plain text content"
    props:
      content:
        type: string
        required: true
        description: "Text content (may contain markdown)"
      format:
        type: string
        required: false
        enum: [plain, markdown]
        default: plain

  status-card:
    version: 1
    category: display
    fallback: text
    description: "Status indicator with title, message, and sections"
    props:
      status:
        type: string
        required: true
        enum: [success, error, warning, info]
      title:
        type: string
        required: true
      message:
        type: string
        required: false
      sections:
        type: Section[]
        required: false
    types:
      Section:
        fields:
          label:
            type: string
          content:
            type: string
          style:
            type: string
            enum: [default, code, error, success]
            default: default

  table:
    version: 1
    category: display
    fallback: text
    description: "Tabular data with headers and rows"
    props:
      title:
        type: string
        required: false
      headers:
        type: string[]
        required: true
      rows:
        type: string[][]
        required: true
      highlight:
        type: uint[]
        required: false
      searchable:
        type: bool
        required: false
        default: false
    input_types:
      - type: selection
        data_schema:
          selected:
            type: uint[]

  code-block:
    version: 1
    category: display
    fallback: text
    description: "Syntax-highlighted code"
    props:
      title:
        type: string
        required: false
      language:
        type: string
        required: true
      code:
        type: string
        required: true

  diff:
    version: 1
    category: display
    fallback: code-block
    description: "Before/after text comparison"
    props:
      title:
        type: string
        required: false
      before:
        type: string
        required: true
      after:
        type: string
        required: true
      before_label:
        type: string
        required: false
        default: "Before"
      after_label:
        type: string
        required: false
        default: "After"
      language:
        type: string
        required: false

  key-value:
    version: 1
    category: display
    fallback: text
    description: "Labeled property list"
    props:
      title:
        type: string
        required: false
      items:
        type: KVItem[]
        required: true
    types:
      KVItem:
        fields:
          key:
            type: string
          value:
            type: string
          style:
            type: string
            enum: [default, success, error, warning, muted]
            default: default

  progress:
    version: 1
    category: display
    fallback: text
    description: "Multi-step progress tracker"
    props:
      title:
        type: string
        required: false
      steps:
        type: Step[]
        required: true
    types:
      Step:
        fields:
          name:
            type: string
          status:
            type: string
            enum: [done, active, pending, failed]
          detail:
            type: string
            default: ""

  chart:
    version: 1
    category: display
    fallback: table
    description: "Data visualization (line, bar, pie, scatter, area)"
    props:
      type:
        type: string
        required: true
        enum: [line, bar, pie, scatter, area]
      title:
        type: string
        required: false
      labels:
        type: string[]
        required: true
      datasets:
        type: Dataset[]
        required: true
      height:
        type: uint
        required: false
    types:
      Dataset:
        fields:
          label:
            type: string
          data:
            type: float[]
          color:
            type: string?

  image:
    version: 1
    category: display
    fallback: text
    description: "Image display"
    props:
      src:
        type: string
        required: true
        description: "URL or base64 data URI"
      alt:
        type: string
        required: true
      width:
        type: uint
        required: false
      height:
        type: uint
        required: false

  markdown:
    version: 1
    category: display
    fallback: text
    description: "Rich markdown content"
    props:
      content:
        type: string
        required: true

  form:
    version: 1
    category: input
    fallback: text
    description: "Interactive form for collecting user input"
    props:
      title:
        type: string
        required: false
      fields:
        type: FormField[]
        required: true
      submit_label:
        type: string
        required: false
        default: "Submit"
      submit_action:
        type: string
        required: false
    types:
      FormField:
        fields:
          name:
            type: string
          label:
            type: string
          field_type:
            type: string
            enum: [text, textarea, select, checkbox, number, date, email, password, file]
          value:
            type: string
            default: ""
          options:
            type: string[]
            default: []
          required:
            type: bool
            default: false
          placeholder:
            type: string
            default: ""
    input_types:
      - type: form_submit
        data_schema:
          "[field.name]":
            type: string
            description: "One entry per form field"

  confirm:
    version: 1
    category: input
    fallback: text
    description: "Binary confirmation dialog"
    props:
      title:
        type: string
        required: true
      message:
        type: string
        required: false
      confirm_label:
        type: string
        required: false
        default: "Confirm"
      deny_label:
        type: string
        required: false
        default: "Cancel"
    input_types:
      - type: action
        data_schema:
          action:
            type: string
            enum: [confirm, deny]

  choices:
    version: 1
    category: input
    fallback: text
    description: "Option picker (buttons or list)"
    props:
      title:
        type: string
        required: false
      options:
        type: Choice[]
        required: true
      style:
        type: string
        required: false
        enum: [buttons, list]
        default: buttons
      multi_select:
        type: bool
        required: false
        default: false
    types:
      Choice:
        fields:
          label:
            type: string
          value:
            type: string
          description:
            type: string?
          icon:
            type: string?
    input_types:
      - type: selection
        data_schema:
          selected:
            type: string[]

  text-input:
    version: 1
    category: input
    fallback: text
    description: "Free text input field"
    props:
      placeholder:
        type: string
        required: false
      max_length:
        type: uint
        required: false
      multiline:
        type: bool
        required: false
        default: false
    input_types:
      - type: text
        data_schema:
          text:
            type: string

  group:
    version: 1
    category: layout
    fallback: null  # children rendered sequentially
    description: "Container for composing multiple components"
    props:
      direction:
        type: string
        required: false
        enum: [vertical, horizontal]
        default: vertical
      gap:
        type: string
        required: false
        enum: [none, small, medium, large]
        default: medium
      children:
        type: Component[]
        required: true

  tabs:
    version: 1
    category: layout
    fallback: group
    description: "Tabbed container showing one child at a time"
    props:
      tabs:
        type: TabDef[]
        required: true
      active_tab:
        type: uint
        required: false
        default: 0
    types:
      TabDef:
        fields:
          label:
            type: string
          component:
            type: Component
    input_types:
      - type: action
        data_schema:
          action:
            type: string
            enum: [tab_change]
          tab_index:
            type: uint

  split:
    version: 1
    category: layout
    fallback: group
    description: "Side-by-side layout with two panels"
    props:
      direction:
        type: string
        required: false
        enum: [horizontal, vertical]
        default: horizontal
      ratio:
        type: string
        required: false
        default: "1:1"
        description: "Panel size ratio (e.g., '1:2', '3:1')"
      left:
        type: Component
        required: true
      right:
        type: Component
        required: true
```

### 14.4 arp-scanner

The `arp-scanner` tool reads schema YAML files and generates language-specific code:

```
arp-scanner --lang=go --output=./gen arp-core.protocol.yaml arp-core.components.yaml
arp-scanner --lang=typescript --output=./gen arp-core.protocol.yaml arp-core.components.yaml
arp-scanner --lang=python --output=./gen arp-core.protocol.yaml arp-core.components.yaml
arp-scanner --lang=swift --output=./gen arp-core.protocol.yaml arp-core.components.yaml
arp-scanner --lang=kotlin --output=./gen arp-core.protocol.yaml arp-core.components.yaml
arp-scanner --lang=rust --output=./gen arp-core.protocol.yaml arp-core.components.yaml
```

Generated code includes:

| Language | Output |
|----------|--------|
| **Go** | Structs for all types/components, marshal/unmarshal, interface definitions, client/renderer stubs |
| **TypeScript** | Interfaces for all types/components, type guards, client/renderer classes, WebSocket transport |
| **Python** | Dataclasses for all types/components, async client/renderer, WebSocket transport |
| **Swift** | Codable structs, async/await client/renderer, URLSession transport |
| **Kotlin** | Data classes, coroutine-based client/renderer, OkHttp transport |
| **Rust** | Structs with serde derives, async client/renderer traits, tokio transport |

---

## 15. Haira Integration

### 15.1 ARP as Native Protocol

Haira implements ARP natively. The Haira compiler and runtime generate ARP messages instead of the current direct JSON-over-SSE format. From the Haira programmer's perspective, nothing changes — the same `@render`, `agent`, and `workflow` syntax works. The protocol change is internal.

### 15.2 How It Maps

| Haira Concept | ARP Mapping |
|---------------|-------------|
| `agent.stream()` in a `@webui` workflow | Creates `ArpSession`, `ArpSurface` (toplevel) |
| `@render("table")` tool returns data | `ArpSurface.render()` with `table` component + `ArpSurface.commit()` |
| Agent streams text deltas | `ArpSurface.patch()` appending to a `text` component |
| `@render("form")` tool shows a form | `ArpSurface.render()` with `form` component |
| User submits form | `ArpInput.input` event with type `form_submit` |
| Agent handoff | `ArpSurface.transfer()` to new agent's session |
| `@render("composite")` group | `ArpSurface.render()` with `group` component |
| `haira-action` DOM event | `ArpInput.input` event with type `action` |

### 15.3 Backward Compatibility

The existing SSE protocol (Chapter 16 of the Haira spec) becomes the **SSE + HTTP POST transport binding** for ARP. Existing clients that consume `tool_render` events continue to work — the Haira runtime translates ARP messages to SSE events for the SSE endpoint.

New clients can connect via WebSocket and speak full ARP. Both coexist on the same server.

### 15.4 Built-in Renderer

Haira's built-in web UI (the Lit web components SDK) becomes an ARP renderer. It:

1. Connects to the Haira server via WebSocket (ARP WebSocket binding).
2. Receives `ArpRegistry.global` events listing available components.
3. Maps component types to `haira-ui-*` custom elements (the existing `COMPONENT_MAP`).
4. Renders components via `setProps()` (unchanged).
5. Captures user interaction and sends `ArpInput.input` events.

The transition is internal. The built-in UI continues to work identically, but now speaks a standard protocol that other renderers can also implement.

### 15.5 CLI Renderer (Reference)

A reference CLI renderer demonstrates transport-agnostic design:

```
Connects via: stdio binding (NDJSON)
Supports: text, table (ASCII art), progress (spinner),
          status-card (colored text), form (interactive prompts),
          confirm (y/n prompt), choices (numbered list)
Does not support: chart, image, diff (falls back to text/code-block)
```

This renderer proves that ARP works beyond the browser.

---

## 16. Conformance Levels

### 16.1 Renderer Conformance

| Level | Requirements | Target |
|-------|-------------|--------|
| **Minimal** | `text` component only, `text` input type, simplified mode (no object model, direct messages) | Voice assistants, IoT displays, embedded |
| **Basic** | `text`, `status-card`, `table`, `form`, `confirm`, `choices`, `group` | CLI terminals, simple mobile apps |
| **Standard** | All core components (Appendix A), full object model, session support | Web apps, desktop apps |
| **Full** | Standard + streaming, multi-surface, overlays, file upload, theming | Rich web apps, IDE integrations |

### 16.2 Agent Conformance

| Level | Requirements |
|-------|-------------|
| **Minimal** | Implements ArpDisplay, ArpSurface (render + commit), ArpInput (ack) |
| **Standard** | Minimal + ArpSession, patch, fallback computation, capability-aware rendering |
| **Full** | Standard + multi-surface, streaming, surface transfer |

### 16.3 Minimal Mode

For constrained environments (SSE, serverless, simple REST), ARP defines a simplified **Minimal mode** that doesn't use the full object model:

```json
{
  "v": 1,
  "type": "render",
  "session_id": "sess_abc",
  "components": [
    { "id": "c1", "type": "table", "props": { ... } }
  ]
}

{
  "v": 1,
  "type": "input",
  "session_id": "sess_abc",
  "source_component": "c1",
  "input_type": "selection",
  "data": { "selected": [0, 2] }
}
```

Minimal mode uses flat messages with a `session_id` instead of object IDs. No `ArpRegistry`, no `bind`, no `commit`. It's syntactic sugar over the full protocol — a renderer can translate minimal-mode messages into full ARP internally.

This ensures ARP doesn't impose unreasonable complexity on simple integrations while maintaining protocol compatibility.

---

## Appendix A: Core Component Catalog

### Display Components

| Component | Version | Fallback | Description |
|-----------|---------|----------|-------------|
| `text` | 1 | (terminal) | Plain or markdown text |
| `status-card` | 1 | `text` | Status indicator with title, message, sections |
| `table` | 1 | `text` | Tabular data with headers and rows |
| `code-block` | 1 | `text` | Syntax-highlighted code |
| `diff` | 1 | `code-block` | Before/after text comparison |
| `key-value` | 1 | `text` | Labeled property list |
| `progress` | 1 | `text` | Multi-step progress tracker |
| `chart` | 1 | `table` | Data visualization (line, bar, pie, scatter, area) |
| `image` | 1 | `text` | Image display |
| `markdown` | 1 | `text` | Rich markdown content |

### Input Components

| Component | Version | Fallback | Description |
|-----------|---------|----------|-------------|
| `form` | 1 | `text` | Interactive form |
| `confirm` | 1 | `text` | Binary confirmation |
| `choices` | 1 | `text` | Option picker |
| `text-input` | 1 | `text` | Free text input |

### Layout Components

| Component | Version | Fallback | Description |
|-----------|---------|----------|-------------|
| `group` | 1 | (sequential) | Component container |
| `tabs` | 1 | `group` | Tabbed container |
| `split` | 1 | `group` | Side-by-side layout |

**Total core components: 17** (10 display + 4 input + 3 layout)

---

## Appendix B: Wire Examples

### B.1 Complete Session: Agent Renders a Table

```json
// 1. Agent bootstraps
→ {"v":1,"id":0,"op":"get_registry","seq":1,"payload":{"new_id":1}}

// 2. Renderer advertises capabilities
← {"v":1,"id":1,"op":"global","seq":1,"payload":{"name":1,"interface":"ArpCompositor","version":1}}
← {"v":1,"id":1,"op":"global","seq":2,"payload":{"name":2,"interface":"ArpInput","version":1}}
← {"v":1,"id":1,"op":"global","seq":3,"payload":{"name":3,"interface":"ArpShell","version":1}}
← {"v":1,"id":1,"op":"global","seq":4,"payload":{"name":4,"interface":"ArpSession","version":1}}
← {"v":1,"id":1,"op":"global","seq":5,"payload":{"name":100,"interface":"component:table","version":1}}
← {"v":1,"id":1,"op":"global","seq":6,"payload":{"name":101,"interface":"component:status-card","version":1}}
← {"v":1,"id":1,"op":"global","seq":7,"payload":{"name":102,"interface":"component:form","version":1}}

// 3. Agent binds to needed globals
→ {"v":1,"id":1,"op":"bind","seq":2,"payload":{"name":1,"interface":"ArpCompositor","version":1,"new_id":2}}
→ {"v":1,"id":1,"op":"bind","seq":3,"payload":{"name":2,"interface":"ArpInput","version":1,"new_id":3}}
→ {"v":1,"id":1,"op":"bind","seq":4,"payload":{"name":3,"interface":"ArpShell","version":1,"new_id":4}}
→ {"v":1,"id":1,"op":"bind","seq":5,"payload":{"name":4,"interface":"ArpSession","version":1,"new_id":5}}

// 4. Agent creates session
→ {"v":1,"id":5,"op":"create","seq":6,"payload":{"session_id":"sess_001","metadata":{"agent":"search-assistant"}}}

// 5. Agent creates surface
→ {"v":1,"id":2,"op":"create_surface","seq":7,"payload":{"new_id":10}}

// 6. Agent assigns toplevel role
→ {"v":1,"id":4,"op":"get_toplevel","seq":8,"payload":{"surface_id":10,"new_id":11}}

// 7. Renderer configures surface
← {"v":1,"id":11,"op":"configure","seq":8,"payload":{"width":800,"height":600,"states":["activated"]}}

// 8. Agent renders a table
→ {"v":1,"id":10,"op":"render","seq":9,"payload":{"components":[{"id":"t1","type":"table","props":{"title":"Search Results","headers":["Name","Price"],"rows":[["Widget A","$29.99"],["Widget B","$39.99"]]}}]}}
→ {"v":1,"id":10,"op":"commit","seq":10,"payload":{}}

// 9. Renderer acknowledges
← {"v":1,"id":10,"op":"frame_ack","seq":9,"payload":{"commit_seq":10}}

// 10. User selects a row
← {"v":1,"id":3,"op":"input","seq":10,"payload":{"surface_id":10,"source_component":"t1","serial":10,"type":"selection","data":{"selected":[0]}}}

// 11. Agent acknowledges input
→ {"v":1,"id":3,"op":"ack","seq":11,"payload":{"serial":10}}
```

### B.2 Minimal Mode Session

```json
// Renderer → Agent (capabilities, sent once on connect)
← {"v":1,"type":"hello","capabilities":{"components":["table","status-card","form","text"],"features":["streaming"]}}

// Agent → Renderer (render)
→ {"v":1,"type":"render","session_id":"sess_001","components":[{"id":"t1","type":"table","props":{"title":"Results","headers":["Name","Price"],"rows":[["Widget A","$29.99"]]}}]}

// Agent → Renderer (stream update)
→ {"v":1,"type":"patch","session_id":"sess_001","ops":[{"op":"update","target":"t1","path":"props.rows","value":[["Widget A","$29.99"],["Widget B","$39.99"]]}]}

// Agent → Renderer (final)
→ {"v":1,"type":"commit","session_id":"sess_001","final":true}

// Renderer → Agent (input)
← {"v":1,"type":"input","session_id":"sess_001","source_component":"t1","input_type":"selection","data":{"selected":[1]}}
```

### B.3 Fallback in Action

A renderer that doesn't support `chart`:

```json
// Agent sends chart with fallback
→ { "id":10, "op":"render", "seq":15, "payload": {
    "components": [{
      "id": "c1",
      "type": "chart",
      "props": {
        "type": "bar",
        "title": "Sales",
        "labels": ["Q1","Q2","Q3","Q4"],
        "datasets": [{"label":"Revenue","data":[100,200,150,300]}]
      },
      "fallback": {
        "type": "table",
        "props": {
          "title": "Sales",
          "headers": ["Quarter", "Revenue"],
          "rows": [["Q1","100"],["Q2","200"],["Q3","150"],["Q4","300"]]
        }
      }
    }]
  }}

// Renderer doesn't support "chart", uses fallback "table"
// Renders a table instead — user still sees the data
```

---

## Acknowledgments

ARP's design is directly informed by the Wayland display protocol, particularly its object model, capability negotiation via registry globals, double-buffered commit semantics, input serial validation, and three-phase extension lifecycle. The Wayland community's decade of experience building a display protocol that prioritizes security, correctness, and extensibility provided the architectural foundation for ARP.

---

*End of specification.*
