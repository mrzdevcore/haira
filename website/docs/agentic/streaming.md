# Streaming

Haira supports Server-Sent Events (SSE) streaming for real-time LLM responses.

## Streaming Workflows

Use `-> stream` as the return type:

```haira
@post("/chat")
workflow Chat(message: string, session_id: string) -> stream {
    return Assistant.stream(message, session: session_id)
}
```

The compiler generates both:
- An SSE endpoint (for streaming clients)
- A JSON fallback endpoint (for non-streaming clients)

## Agent Streaming

The `.stream()` method on agents returns a stream:

```haira
agent Writer {
    provider: openai
    system: "You are a creative writer."
    memory: conversation(max_turns: 10)
    temperature: 0.9
}

@post("/write")
workflow Write(prompt: string, sid: string) -> stream {
    return Writer.stream(prompt, session: sid)
}
```

## Auto-generated Chat UI

Streaming workflows get a chat UI via `haira webui`:

```bash
# Start your server
haira run app.haira

# In another terminal, launch the webui
haira webui -c localhost:8080 -p 3000
```

The chat UI supports:
- Real-time token streaming
- Session management
- Message history display

## Client-side Consumption

### JavaScript/TypeScript

```javascript
const source = new EventSource('/chat');
source.onmessage = (event) => {
    const data = JSON.parse(event.data);
    console.log(data.content);
};
```

### curl

```bash
curl -N -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Tell me a story", "session_id": "abc"}'
```

## Complete Example

```haira
import "io"
import "http"

provider openai {
    api_key: env("OPENAI_API_KEY")
    model: "gpt-4o"
}

agent Writer {
    provider: openai
    system: "You are a creative writer. Write short, engaging responses."
    memory: conversation(max_turns: 10)
    temperature: 0.9
}

@post("/api/stream")
workflow Stream(message: string, session_id: string) -> stream {
    return Writer.stream(message, session: session_id)
}

fn main() {
    server = http.Server([Stream])
    io.println("Server running on :8080")
    server.listen(8080)
}
```
