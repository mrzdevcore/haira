---
title: "Agent Templates"
description: "Pre-built agent templates for common use cases."
---

# Agent Templates

The `agents` stdlib package provides pre-built agent configuration templates for common use cases. Use them with `create_agent()` to quickly spin up specialized agents.

## Usage

```haira
import "agents"

provider openai {
    api_key: env("OPENAI_API_KEY")
    model: "gpt-4o"
}

fn main() {
    config = agents.code_reviewer()
    reviewer = create_agent(config, openai, [lint_code])
    result, err = reviewer.ask("Review this function...")
}
```

## Available Templates

| Template | Function | Description |
|----------|----------|-------------|
| Code Reviewer | `agents.code_reviewer()` | Reviews code for bugs, security, and style |
| Planner | `agents.planner()` | Breaks tasks into actionable plans |
| Security Reviewer | `agents.security_reviewer()` | Audits code for security vulnerabilities |
| Summarizer | `agents.summarizer()` | Summarizes text into key points |
| Data Analyst | `agents.data_analyst()` | Analyzes data and generates insights |
| Customer Support | `agents.customer_support()` | Handles customer support queries |
| TDD Guide | `agents.tdd_guide()` | Guides test-driven development |
| Doc Writer | `agents.doc_writer()` | Writes documentation from code |

## Template Configuration

Each template returns a configuration map with:

```haira
{
    "name": "CodeReviewer",
    "system": "You are an expert code reviewer...",
    "temperature": 0.3
}
```

You can override any field after getting the config:

```haira
config = agents.code_reviewer()
config["temperature"] = 0.1  // More deterministic
reviewer = create_agent(config, anthropic, [lint_code, check_tests])
```

## Dynamic Agents

The `create_agent()` built-in function creates agents at runtime from configuration maps:

```haira
fn create_custom_agent(role: string) {
    config = {
        "name": role,
        "system": "You are a ${role}. Be helpful and concise.",
        "temperature": 0.5
    }
    return create_agent(config, openai, [])
}
```

## Next Steps

- [Agents](/docs/agentic/agents) — static agent declarations
- [Tools](/docs/agentic/tools) — tools for agent templates
- [Evaluation](/docs/agentic/evaluation) — test your agents
