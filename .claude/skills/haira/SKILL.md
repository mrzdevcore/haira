---
name: haira
description: >
  Complete knowledge of the Haira programming language — syntax, semantics,
  standard library, and idiomatic patterns. Use this skill whenever the user
  asks about Haira language syntax, wants to write .haira code, needs help with
  agentic constructs (providers, tools, agents, workflows), or works on any
  .haira files. Triggers on: Haira, .haira, haira language, agentic workflow,
  provider/tool/agent/workflow declarations.
user-invocable: true
argument-hint: "[question or task]"
---

# Haira Language — Complete Agent Knowledge

**HAIRA** stands for **High-level Agentic Instruction & Runtime Architecture**.
Haira is a general-purpose agentic orchestration programming language.
Tagline: "Build agents and workflows, not boilerplate."

**Pipeline:** `.haira` source → Lexer → Parser → Checker → Go Codegen → `go build` → Native binary

## Quick Reference

### All Keywords

**Control flow:** `if`, `else`, `for`, `while`, `match`, `break`, `continue`, `return`
**Functions:** `fn`
**Type declarations:** `struct`, `enum`, `type`
**Agentic:** `agent`, `provider`, `tool`, `workflow`
**Modules:** `import`, `export`, `from`, `pub`
**Concurrency:** `spawn`, `chan`, `select`
**Error handling:** `try`, `catch`, `defer`, `errdefer`
**Operators as keywords:** `and`, `or`, `not`, `as`, `in`, `orelse`
**Literals:** `true`, `false`, `nil`
**Reserved (future):** `trait`, `impl`, `async`, `const`, `unsafe`, `where`, `step`

### All Types

**Integers:** `int`, `i8`, `i16`, `i32`, `i64`, `u8`, `u16`, `u32`, `u64`
**Floats:** `float`, `f32`, `f64`
**Other primitives:** `bool`, `string`
**Collections:** `[]T` (array), `[K:V]` (map), `(T, U)` (tuple)
**Special:** `T?` (option), `stream<T>`, `fn(T) -> U` (function type), `chan<T>` (channel)

### Operator Precedence (highest to lowest)

1. Postfix: `()` `[]` `.` `?.`
2. Unary: `!` `-` `~` `not`
3. Multiplicative: `*` `/` `%`
4. Additive: `+` `-`
5. Shift: `<<` `>>`
6. Bitwise AND: `&`
7. Bitwise XOR: `^`
8. Bitwise OR: `|`
9. Range: `..` `..=`
10. Pipe: `|>`
11. Comparison: `<` `>` `<=` `>=`
12. Equality: `==` `!=`
13. Logical AND: `and` `&&`
14. Logical OR: `or` `||`
15. Assignment: `=` `+=` `-=` `*=` `/=` `&=` `|=` `^=` `<<=` `>>=`

## Language Syntax

### Variables & Constants

```haira
x = 42                    // type inferred as int
name: string = "Haira"    // explicit type
a, b = 1, 2              // multiple assignment
const PI = 3.14159        // immutable
```

### Functions

```haira
fn add(a: int, b: int) -> int {
    return a + b
}

fn greet(name: string, greeting: string = "Hello") {
    io.println("${greeting}, ${name}!")
}

// Implicit return (last expression)
fn double(n: int) -> int { n * 2 }

// Multiple returns
fn divide(a: int, b: int) -> (int, Error?) {
    if b == 0 { return 0, Error{message: "division by zero"} }
    return a / b, nil
}

// Variadic
fn sum(numbers: ...int) -> int { /* ... */ }
```

### Structs & Methods

```haira
struct User {
    name: string
    age: int
    email: string
}

// Method with implicit self
User.greet() -> string {
    return "Hello, I'm ${self.name}"
}

user = User{ name: "Alice", age: 30, email: "alice@ex.com" }
user.greet()  // "Hello, I'm Alice"
```

### Enums & Match

```haira
enum Status { Pending, Active, Completed }
enum Result<T> { Ok(T), Err(Error) }

match status {
    Status.Pending => io.println("waiting")
    Status.Active | Status.Completed => io.println("in progress or done")
    _ => io.println("unknown")
}

// Guards
match n {
    x if x < 0 => "negative"
    0 => "zero"
    x if x > 0 => "positive"
}

// Range patterns
match score {
    90..=100 => "A"
    80..90 => "B"
    _ => "other"
}
```

### Control Flow

```haira
// If/else (also an expression)
max = if a > b { a } else { b }

// For loops
for i in 0..10 { /* 0-9 */ }
for i in 0..=10 { /* 0-10 */ }
for item in items { /* iterate */ }
for i, item in items { /* with index */ }
for key, value in my_map { /* iterate map */ }

// While
while condition { /* body */ }

// Break/continue with labels
outer: for i in 0..10 {
    for j in 0..10 {
        if condition { break outer }
    }
}
```

### String Interpolation

```haira
name = "World"
greeting = "Hello, ${name}!"
complex = "Result: ${compute(x + y)}"
```

### Pipe Operator

```haira
result = "  hello, world  "
    |> string.trim
    |> string.split(", ")
    |> array.map(string.to_upper)
    |> string.join(" - ")
// "HELLO - WORLD"
```

### Error Handling

```haira
// Tuple pattern
result, err = operation()
if err != nil { return err }

// ? operator (panics on error, use with try/catch)
content = read_file(path)?

// Try/catch
try {
    config = load_config("app.toml")?
    db = connect(config.db_url)?
} catch err {
    io.println("Failed: " + err)
}

// Orelse (default on error)
count = parse_int(input) orelse 0

// Defer / errdefer
defer fs.close(file)
errdefer db.close()  // only runs on panic
```

### Closures / Lambdas

```haira
// Lambda syntax uses => (NOT fn keyword)
add = (a: int, b: int) => a + b
double = (n) => n * 2

// Block body with parens
counter = (n) { return n + 1 }

// Single param without parens
inc = x => x + 1

// Multi-line block
process = (data) {
    result = transform(data)
    return result
}

// Captures by reference
fn make_counter() -> fn() -> int {
    count = 0
    return () {
        count += 1
        return count
    }
}
```

**IMPORTANT:** There is NO `fn(x) { ... }` syntax for inline lambdas. Use `(x) => expr` or `(x) { block }` instead.

### Using Higher-Order Functions

```haira
// With named functions (preferred for readability):
fn double(n: int) -> int { return n * 2 }
result = 5 |> double  // pipe passes value as first arg

// With lambdas as call arguments:
// Lambdas work as arguments to stdlib calls
filtered = array.filter(nums, (n) => n > 10)

// For complex operations, use for loops instead of map with lambdas:
results = []
for item in items {
    results = array.push(results, process(item))
}
```

### Concurrency

```haira
// Spawn
spawn { do_work() }

// Spawn block (parallel, returns list)
results = spawn {
    agent1.run(req1)
    agent2.run(req2)
    agent3.run(req3)
}

// Channels
ch = chan<int>(10)          // buffered
ch <- 42                   // send
value = <-ch               // receive
for msg in ch { /* ... */ } // iterate until closed
close(ch)

// Select
select {
    msg = <-ch1 => handle(msg)
    msg = <-ch2 => handle2(msg)
    default => io.println("no messages")
}
```

### Imports & Modules

```haira
import "io"                          // basic
import fmt from "io"                 // aliased
import { User, Post } from "models"  // selective
import * from "math"                 // glob
export { Name1, Name2 }             // re-export from mod.haira
```

**Visibility:** Private by default. `pub` to export. Agentic declarations always public.

## Agentic Constructs

### Provider

```haira
provider anthropic {
    api_key: env("ANTHROPIC_API_KEY")
    model: "claude-sonnet-4-20250514"
}

provider openai {
    api_key: env("OPENAI_API_KEY")
    model: "gpt-4o"
}

provider local {
    backend: "ollama"
    host: "localhost:11434"
    model: "llama3:8b"
}

// MCP Provider (stdio)
provider filesystem {
    transport: "mcp"
    command: "npx"
    args: ["-y", "@modelcontextprotocol/server-filesystem", "/tmp"]
}

// MCP Provider (SSE)
provider remote_tools {
    transport: "sse"
    endpoint: "http://localhost:9000/sse"
}
```

**Fields:** `model` (required), `api_key`, `backend`, `host`, `temperature`, `max_tokens`, `transport`, `command`, `args`, `endpoint`, `env`, `headers`, `input_token_cost`, `output_token_cost`, `api_version`

### Tool

```haira
tool search(query: string, max_results: int = 5) -> [SearchResult] {
    """
    Search the web for information.
    Returns up to max_results results.
    """
    resp, err = http.get("https://api.search.com?q=${query}")
    if err != nil { return [], err }
    return json.decode(resp.body, [SearchResult])
}
```

- Triple-quoted description is **mandatory** (compiler-enforced)
- Compiler auto-generates JSON schema from type signature
- Tools are always public

### Agent

```haira
agent Researcher {
    provider: anthropic
    system: "You are a thorough researcher. Always cite sources."
    tools: [search, read_url]
    temperature: 0.2
    max_steps: 15
    memory: conversation(max_turns: 50)
    timeout: 120
}

// With handoffs
agent FrontDesk {
    provider: anthropic
    system: """
        Route billing questions to BillingAgent.
        Route tech issues to TechAgent.
    """
    handoffs: [BillingAgent, TechAgent]
    memory: conversation(max_turns: 10)
}

// With MCP tools
agent Assistant {
    provider: anthropic
    system: "Helpful assistant with file access."
    tools: [greet]
    mcp: [filesystem]
    memory: conversation(max_turns: 10)
}

// With structured output
agent Analyzer {
    provider: openai
    system: "Analyze text and return structured data."
    output: Analysis
}
```

**Fields:** `provider` (required), `system` (required), `tools`, `temperature`, `max_tokens`, `max_steps`, `memory`, `handoffs`, `timeout`, `mcp`, `output`, `ui`

**Memory types:** `conversation(max_turns: N)`, `summary(max_tokens: N)`, `none` (default)

### Agent Methods

```haira
// .ask() — text in, text out
answer, err = Agent.ask("question")
answer, err = Agent.ask("question", session: "user-123")

// .run() — structured in, structured out (left-side type annotation)
result: OutputType, err = Agent.run(InputStruct{ field: value })
result: AgentResult, err = Agent.run(msg, session: sid)

// .stream() — text in, streaming out
for chunk in Agent.stream("prompt", session: "user-123") {
    io.print(chunk)
}
```

### Workflow

```haira
// HTTP webhook
@webhook("/api/chat")
workflow Chat(message: string, session_id: string) -> { reply: string } {
    reply, err = Assistant.ask(message, session: session_id)
    if err != nil { return { reply: "Error" } }
    return { reply: reply }
}

// Streaming
@webhook("/api/stream")
workflow Stream(message: string, session_id: string) -> stream {
    return Assistant.stream(message, session: session_id)
}

// With steps and error handlers
@webui(title: "Summarizer", description: "Upload and summarize files")
@post("/api/summarize")
workflow Summarize(document: file, context: string) -> { summary: string } {
    onerror err {
        return { summary: "Error: ${err}" }
    }

    step "Read file" {
        content, read_err = io.read_file(document)
        if read_err != nil { return { summary: "Failed" } }
    }

    step "Summarize" {
        reply, err = Summarizer.ask(content)
        if err != nil { return { summary: "AI error" } }
    }

    return { summary: reply }
}
```

**Triggers:** `@webhook("/path")`, `@webhook("/path", method: "GET")`, `@websocket("/path")`, `@cron("0 9 * * *")`, `@event("order.created")`, `@manual`, `@webui(...)`, `@get("/path")`, `@post("/path")`

**Sub-workflows:** Call other workflows as functions (no decorator needed).

### Server

```haira
fn main() {
    server = http.Server([Chat, Stream, Stats])
    io.println("Running on :8080")
    server.listen(8080)
}
```

### MCP Server

```haira
fn main() {
    mcp_server = mcp.Server([Summarize, Translate])
    mcp_server.listen(9000)       // SSE mode
    // or: mcp_server.serve()     // stdio mode
}
```

## Standard Library

See [references/syntax.md](references/syntax.md) for complete stdlib function signatures.

**Core modules:** `io`, `string`, `array`, `map`, `math`, `json`, `conv`, `fs`, `os`, `time`, `regex`, `http`, `log`, `ui`

**Agentic modules:** `vector`, `mcp`, `observe`

**External modules (stdlib, tree-shaken):** `excel`, `postgres`, `sqlite`, `slack`, `github`, `gitlab`, `langfuse`, `websearch`, `healthcheck`, `algolia`, `meilisearch`

### CRITICAL: Import Rules

Every stdlib module MUST be imported before use:
```haira
import "conv"    // required to use conv.int_to_string(), conv.to_string(), etc.
import "math"    // required to use math.min(), math.max(), etc.
import "array"   // required to use array.map(), array.reduce(), etc.
```

### CRITICAL: Functions That Do NOT Exist

Do NOT use any of these — they will cause compile errors:
- `math.average()` — does NOT exist. Use `math.avg(arr)` instead
- `string.length()` — does NOT exist. Use `string.len(s)` or `len(s)`
- `conv.parse_int()` / `conv.parse_float()` — do NOT exist. Use `conv.string_to_int(s)` and `conv.string_to_float(s)`
- `excel.close(wb)` — does NOT exist. Use `wb.close()` (method call)
- `excel.get_sheet()`, `excel.get_rows()`, `excel.get_cell()` — do NOT exist

### CRITICAL: Function Signatures — Watch Out

- `math.min(a, b)` / `math.max(a, b)` — take exactly **2 scalar args** (not an array). For array min/max use `array.min(arr)` / `array.max(arr)`
- `math.sum(arr)` / `math.avg(arr)` — take an **array** argument
- `array.sum(arr)` / `array.avg(arr)` — also take an **array** (equivalent to math.sum/avg)

### Type Handling with `any` Values

Excel rows, map values, and some array operations return `any` typed values. The runtime functions accept `any` parameters (they do internal type assertions), so you can pass them directly:

```haira
// Excel rows are [map] where values are `any` — this works directly:
rows, err = wb.read_sheet("Sheet1")
for row in rows {
    name = row["Name"]           // any — can be used in string interpolation
    value = row["Amount"]        // any — can be passed to math/conv functions
    io.println("${name}: ${conv.to_string(value)}")
}

// math functions accept `any` — no casting needed:
math.min(row["A"], row["B"])     // works with any numeric values
conv.float_to_int(row["Price"])  // works with any

// Aggregate functions on arrays:
amounts = []
for row in rows {
    amounts = array.push(amounts, row["Amount"])
}
total = math.sum(amounts)        // or array.sum(amounts)
avg = math.avg(amounts)          // or array.avg(amounts)
smallest = array.min(amounts)
largest = array.max(amounts)
```

### Excel Module (stdlib)

```haira
import "excel"

// Open a workbook (low-level)
wb, err = excel.open("data.xlsx")
if err != nil { io.println("Error: ${err}") }
defer wb.close()

names = wb.sheet_names()              // -> [string]
rows, err = wb.read_sheet("Sheet1")   // -> [map], each row is {col: value}

// High-level API (recommended)
tables, err = excel.read_sheets("data.xlsx")
sheet_names = tables.names()          // -> [string]
rows = tables.sheet(name)             // -> [map]
headers = tables.sheet_headers(name)  // -> [string]
summary = tables.summary()            // -> string
count = tables.len()                  // -> int
```

**IMPORTANT:** There is NO `excel.get_sheet()`, `excel.get_sheet_by_index()`, `excel.get_cell()`, `excel.get_rows()`, or `excel.close(wb)`. Use method calls on the workbook/tables object.

### Log Module (codegen-injected)

```haira
import "log"

// Inside workflow steps — logs appear in step UI timeline
step "Process data" {
    log.info("Starting processing...")
    log.warn("Missing optional field")
    log.error("Critical failure")
}

// Outside steps — printed to stderr
log.info("Server starting")
```

**Note:** `log` has no runtime file — it is injected at codegen time. Inside steps it calls `StepLog()`, outside steps it calls `LogPrint()`.

### Websearch & Healthcheck (stdlib)

```haira
import "websearch"
import "healthcheck"

// Web search (DuckDuckGo, no API key needed)
result = websearch.search("query")

// Health check
check = healthcheck.check("service-name", "https://api.example.com/health")
// Returns: {name, status, response_time}

// Batch health checks
services = [
    { "name": "api", "url": "https://api.example.com" },
    { "name": "db", "url": "https://db.example.com" }
]
results = healthcheck.check_all(services)
```

## Settled Design Decisions (DO NOT change)

- No `ai` keyword — all LLM via `agent` at runtime
- No semicolons
- No class/inheritance/traits in v1
- `pub` for visibility (private by default)
- `${expr}` for string interpolation
- `|>` for pipe (not `|`)
- `?` panics on error (caught by try/catch)
- Workflows are function-style
- Triggers use `@decorator` syntax
- Tool descriptions use `"""..."""` (mandatory)
- Agent output uses left-side annotation: `result: Type, err = agent.run(req)`

## Build Commands

```bash
make build            # Build compiler
make test             # Run tests
make dev              # fmt + vet + test
make ci               # vet + test + build-examples
make build-examples   # Compile all .haira examples
make install          # Install to $GOPATH/bin
```
