---
title: "Concurrency"
description: "Parallel execution with spawn blocks and concurrent agent calls in Haira."
---

# Concurrency

## Spawn Blocks

The `spawn` block runs tasks in parallel and returns a list of results:

```haira
results = spawn {
    fetch_user(1)
    fetch_user(2)
    fetch_user(3)
}
// results is a list of all return values
```

All tasks inside a spawn block execute concurrently. The block waits for all tasks to complete before continuing.

## Parallel Agent Calls

A common pattern is fanning out to multiple agents:

```haira
results = spawn {
    Researcher.ask("Find info on topic A")
    Researcher.ask("Find info on topic B")
    Analyst.ask("Analyze current trends")
}
```

## Parallel Tool Execution

```haira
tool fetch_price(symbol: string) -> string {
    """Fetch stock price for a given symbol"""
    resp, err = http.get("https://api.stocks.com/${symbol}")
    if err != nil { return "error" }
    return resp.body
}

prices = spawn {
    fetch_price("AAPL")
    fetch_price("GOOG")
    fetch_price("MSFT")
}
```

## In Workflows

```haira
@post("/api/analyze")
workflow Analyze(topic: string) -> { summary: string } {
    // Gather data in parallel
    data = spawn {
        search_docs(topic)
        search_web(topic)
        get_recent_papers(topic)
    }

    // Synthesize results
    combined = string.join(data, "\n\n")
    summary, err = Analyst.ask("Summarize: ${combined}")
    if err != nil {
        return { summary: "Analysis failed." }
    }
    return { summary: summary }
}
```
