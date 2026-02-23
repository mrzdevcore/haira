# Hello World

Let's write your first Haira program.

## A Simple Program

Create a file called `hello.haira`:

```haira
import "io"

fn main() {
    io.println("Hello, World!")
}
```

## Build & Run

```bash
# Compile to binary
haira build hello.haira

# Run the binary
./hello
```

Or use `haira run` to compile and run in one step:

```bash
haira run hello.haira
```

Output:
```
Hello, World!
```

## Variables and String Interpolation

```haira
import "io"

fn main() {
    name = "Haira"
    version = 0.1
    io.println("Welcome to ${name} v${version}!")
}
```

Haira uses `${expr}` for string interpolation — any expression can go inside the braces.

## Your First Agent

Here's a minimal AI agent with an HTTP endpoint:

```haira
import "io"
import "http"

provider openai {
    api_key: env("OPENAI_API_KEY")
    model: "gpt-4o"
}

agent Assistant {
    provider: openai
    system: "You are a helpful assistant. Be concise."
}

@post("/chat")
workflow Chat(message: string) -> { reply: string } {
    reply, err = Assistant.ask(message)
    if err != nil {
        return { reply: "Something went wrong." }
    }
    return { reply: reply }
}

fn main() {
    server = http.Server([Chat])
    io.println("Running on :8080")
    server.listen(8080)
}
```

```bash
# Set your API key
export OPENAI_API_KEY="sk-..."

# Build and run
haira run agent.haira
```

Your agent is now serving at `http://localhost:8080/chat`. To view the UI, run:

```bash
haira webui -c localhost:8080 -p 3000
```

Then open `http://localhost:3000` in your browser.

## What Just Happened?

In ~20 lines you defined:

1. **A provider** — configures which LLM backend to use
2. **An agent** — an LLM entity with a system prompt
3. **A workflow** — an HTTP endpoint that calls the agent
4. **A server** — serves everything as a native binary

No frameworks. No boilerplate. Just four keywords.

## Next Steps

- [Key Concepts](/docs/getting-started/key-concepts) — understand the four primitives
- [Providers](/docs/agentic/providers) — configure LLM backends
- [Agents](/docs/agentic/agents) — build intelligent agents
- [Examples](/docs/examples) — see more real-world patterns
