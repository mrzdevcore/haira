---
title: "Auth"
description: "API key resolution and authentication helpers."
---

# Auth

The `auth` stdlib package provides API key resolution for multiple backends — environment variables, credentials files, and Claude Code OAuth.

## Usage

```haira
import "auth"

fn main() {
    key, err = auth.resolve("OPENAI_API_KEY")
    if err != nil {
        io.println("No API key found: ${err}")
    }
    io.println("Key resolved: ${key}")
}
```

## Resolution Order

`auth.resolve(name)` checks these sources in order:

1. **Environment variable** — `os.Getenv(name)`
2. **Credentials file** — `~/.haira/credentials.json`
3. **Claude Code OAuth** — `~/.claude/credentials.json`

The first match wins. If no source has the key, an error is returned.

## Credentials File

Create `~/.haira/credentials.json`:

```json
{
    "OPENAI_API_KEY": "sk-...",
    "ANTHROPIC_API_KEY": "sk-ant-..."
}
```

## Functions

| Function | Description |
|----------|-------------|
| `auth.resolve(name)` | Resolve an API key by name |
| `auth.resolve_or(name, fallback)` | Resolve with fallback value |

## Next Steps

- [Providers](/docs/agentic/providers) — use resolved keys in providers
- [Agent Templates](/docs/stdlib/agents) — pre-built agents
