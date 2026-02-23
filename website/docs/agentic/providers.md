# Providers

A provider configures an LLM backend — API credentials, model, and optional parameters.

## Defining a Provider

```haira
provider openai {
    api_key: env("OPENAI_API_KEY")
    model: "gpt-4o"
}
```

## Supported Providers

### OpenAI

```haira
provider openai {
    api_key: env("OPENAI_API_KEY")
    model: "gpt-4o"
}
```

### Azure OpenAI

```haira
provider azure_openai {
    api_key: env("AZURE_OPENAI_API_KEY")
    endpoint: env("AZURE_OPENAI_ENDPOINT")
    model: env("AZURE_OPENAI_DEPLOYMENT_NAME")
    api_version: env("AZURE_OPENAI_API_VERSION")
}
```

### Anthropic

```haira
provider anthropic {
    api_key: env("ANTHROPIC_API_KEY")
    model: "claude-sonnet-4-5-20250929"
}
```

### Ollama

Run local models with Ollama. No API key needed:

```haira
provider local_ollama {
    model: "llama3"
    endpoint: "http://localhost:11434"
}
```

Or use the `host` shorthand (automatically resolves to `http://<host>/v1`):

```haira
provider local_ollama {
    model: "llama3"
    host: "localhost:11434"
}
```

### OpenAI-Compatible APIs

Any OpenAI-compatible endpoint works — Groq, Mistral, llama.cpp, vLLM, etc.:

```haira
provider groq {
    api_key: env("GROQ_API_KEY")
    model: "llama-3.1-70b-versatile"
    endpoint: "https://api.groq.com/openai/v1"
}

provider mistral {
    api_key: env("MISTRAL_API_KEY")
    model: "mistral-large-latest"
    endpoint: "https://api.mistral.ai/v1"
}
```

## Multiple Providers

You can define multiple providers and assign different ones to different agents:

```haira
provider fast_model {
    api_key: env("OPENAI_API_KEY")
    model: "gpt-4o-mini"
}

provider local {
    model: "llama3"
    endpoint: "http://localhost:11434"
}

provider smart_model {
    api_key: env("ANTHROPIC_API_KEY")
    model: "claude-sonnet-4-5-20250929"
}

agent QuickBot {
    provider: fast_model
    system: "Be brief."
}

agent LocalBot {
    provider: local
    system: "You run locally."
}

agent ThinkBot {
    provider: smart_model
    system: "Think deeply and thoroughly."
}
```

## Provider Fields

| Field | Required | Description |
|-------|----------|-------------|
| `api_key` | Cloud only | API key (typically from env) |
| `model` | Yes | Model identifier |
| `endpoint` | No | Full API endpoint URL |
| `host` | No | Local backend host (resolved to `http://<host>/v1`) |
| `api_version` | Azure only | API version string |
| `backend` | No | Informational label (e.g., `"ollama"`, `"llama.cpp"`) |
| `input_token_cost` | No | USD per 1M input tokens (for cost tracking) |
| `output_token_cost` | No | USD per 1M output tokens (for cost tracking) |

## Cost Tracking

Track LLM costs per provider:

```haira
provider azure_openai {
    api_key: env("AZURE_OPENAI_API_KEY")
    endpoint: env("AZURE_OPENAI_ENDPOINT")
    model: env("AZURE_OPENAI_DEPLOYMENT_NAME")
    api_version: env("AZURE_OPENAI_API_VERSION")
    input_token_cost: 0.15
    output_token_cost: 0.60
}
```

## Best Practices

- Always use `env()` for API keys — never hardcode secrets
- Name providers descriptively (`fast_model`, `local_ollama`, etc.)
- Switching models is a one-line change in the provider definition
- Use `host` for local backends, `endpoint` for full custom URLs
- Local providers (Ollama) don't need an `api_key`
