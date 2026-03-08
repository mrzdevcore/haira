---
title: "Evaluation"
description: "Automated agent evaluation with test cases, scoring, and pass/fail thresholds."
---

# Evaluation

The `eval` framework lets you define automated test suites for your agents. Each eval block specifies test cases with expected outputs and a pass/fail threshold.

## Defining an Eval

```haira
eval "Customer Support Quality" {
    agent: SupportBot
    cases: [
        {
            input: "How do I reset my password?",
            expected: "Go to Settings > Security > Reset Password"
        },
        {
            input: "What are your pricing plans?",
            expected: "starter, professional, enterprise"
        },
        {
            input: "I need a refund",
            expected: "refund policy"
        }
    ]
    threshold: 0.8
}
```

## Eval Fields

| Field | Required | Description |
|-------|----------|-------------|
| `agent` | Yes | The agent to evaluate |
| `cases` | Yes | List of test cases with `input` and `expected` |
| `threshold` | No | Minimum pass rate (0.0–1.0, default 0.7) |

## Test Cases

Each case has:
- `input` — the prompt sent to the agent
- `expected` — substring(s) that should appear in the agent's response

The eval runner checks if the agent's response contains the expected text (case-insensitive substring match).

## Running Evals

```bash
haira eval main.haira
```

Output:

```
=== Eval: Customer Support Quality ===
  [PASS] "How do I reset my password?" (score: 1.00)
  [PASS] "What are your pricing plans?" (score: 1.00)
  [FAIL] "I need a refund" (score: 0.00)
         Expected to contain: "refund policy"

Results: 2/3 passed (66.7%) — Threshold: 80.0%
FAIL
```

## Multiple Eval Blocks

A single file can contain multiple eval declarations:

```haira
eval "Accuracy" {
    agent: Assistant
    cases: [
        { input: "What is 2+2?", expected: "4" },
        { input: "Capital of France?", expected: "Paris" }
    ]
    threshold: 1.0
}

eval "Tone" {
    agent: Assistant
    cases: [
        { input: "Hello", expected: "hello" },
        { input: "Thanks!", expected: "welcome" }
    ]
    threshold: 0.5
}
```

## CI Integration

Add eval to your CI pipeline:

```yaml
- name: Run agent evaluations
  run: haira eval tests/eval.haira
  env:
    OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
```

The `haira eval` command exits with code 1 if any eval block fails its threshold, making it CI-friendly.

## Next Steps

- [Agents](/docs/agentic/agents) — define agents to evaluate
- [Cross-Harness Export](/docs/agentic/exports) — export agents to Claude Code
