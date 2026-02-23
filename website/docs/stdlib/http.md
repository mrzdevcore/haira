---
title: "HTTP & Server"
description: "HTTP client, server, and REST endpoint handling in Haira."
---

# HTTP & Server

## HTTP Client

### GET Request

```haira
import "http"

resp, err = http.get("https://api.example.com/data")
if err != nil {
    io.println("Error: ${err}")
}

body = resp.body
status = resp.status
data = resp.json()  // Parse JSON response
```

### POST Request

```haira
resp, err = http.post("https://api.example.com/create", {
    "name": "Alice",
    "email": "alice@example.com"
})
```

### With Headers

```haira
resp, err = http.get("https://api.example.com/data", {
    "Authorization": "Bearer ${token}",
    "Content-Type": "application/json"
})
```

## HTTP Server

### Creating a Server

```haira
fn main() {
    server = http.Server([Chat, Analyze, Upload])
    io.println("Server running on :8080")
    server.listen(8080)
}
```

### Server Features

The server automatically provides:
- REST endpoints for each workflow
- SSE endpoints for streaming workflows
- Auto-generated UI via `haira webui`
- CORS handling
- Request parsing (JSON body, file uploads)

### Workflow Routing

Workflows are mapped to routes by their decorators:

```haira
@post("/api/chat")
workflow Chat(msg: string) -> { reply: string } { /* ... */ }

@post("/api/analyze")
workflow Analyze(text: string) -> { result: string } { /* ... */ }

@webhook("/hooks/notify")
workflow Notify(payload: string) -> { status: string } { /* ... */ }
```
