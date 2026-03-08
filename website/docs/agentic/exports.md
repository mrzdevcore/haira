---
title: "Cross-Harness Export"
description: "Export Haira agents and tools to Claude Code format with --target claude-code."
---

# Cross-Harness Export

Haira agents can be exported to other agent harnesses. The `--target claude-code` flag generates Claude Code configuration files and an MCP server binary for custom tools.

## Usage

```bash
haira build main.haira --target claude-code
```

## Output Structure

```
.claude/
  agents/agent-name.md       # Agent system prompt + tools + config
  commands/workflow-name.md   # Slash commands for workflows
  mcp-servers.json            # Points to MCP binary
bin/
  haira-tools                 # MCP server binary with custom tools
```

## Agent Files

Each `agent` declaration generates a `.claude/agents/<name>.md` file containing:

- Agent name and model
- System prompt
- Tool references (via MCP)
- Handoff targets

```haira
agent CodeReviewer {
    provider: openai
    system: "Review code for bugs, security issues, and style."
    tools: [lint_code, check_security]
}
```

Generates `.claude/agents/code-reviewer.md`:

```markdown
# CodeReviewer

Model: gpt-4o

Review code for bugs, security issues, and style.

## Tools

- lint_code (via MCP: haira-tools)
- check_security (via MCP: haira-tools)
```

## Workflow Commands

Each `workflow` generates a `.claude/commands/<name>.md` slash command:

```haira
@post("/api/deploy")
workflow Deploy(service: string, env: string) -> { status: string } {
    step "Validate" { /* ... */ }
    step "Build" { /* ... */ }
    step "Deploy" { /* ... */ }
}
```

Generates `.claude/commands/deploy.md`:

```markdown
# /deploy

## Parameters

- `service` (string)
- `env` (string)

## Steps

1. **Validate**
2. **Build**
3. **Deploy**
```

## MCP Server

Custom tools are compiled into a standalone MCP server binary (`bin/haira-tools`) that communicates via stdio. The `.claude/mcp-servers.json` file points Claude Code to this binary:

```json
{
  "haira-tools": {
    "command": "/absolute/path/to/bin/haira-tools",
    "args": []
  }
}
```

## Complete Example

```haira
import "http"

provider openai {
    api_key: env("OPENAI_API_KEY")
    model: "gpt-4o"
}

tool analyze_code(code: string) -> string {
    """Analyze code for potential issues and improvements"""
    resp, err = Reviewer.ask("Review this code:\n${code}")
    if err != nil { return "Analysis failed." }
    return resp
}

agent Reviewer {
    provider: openai
    system: "You are an expert code reviewer."
    tools: [analyze_code]
}

@post("/api/review")
workflow Review(code: string) -> { feedback: string } {
    feedback, err = Reviewer.ask(code)
    if err != nil { return { feedback: "Error." } }
    return { feedback: feedback }
}
```

```bash
haira build review.haira --target claude-code
# Exported Claude Code config:
#   1 agent(s) → .claude/agents/
#   1 command(s) → .claude/commands/
#   1 tool(s) → bin/haira-tools (MCP)
```

## Next Steps

- [Tools](/docs/agentic/tools) — define tools for export
- [Agents](/docs/agentic/agents) — define agents with handoffs
- [Evaluation](/docs/agentic/evaluation) — test agents before exporting
