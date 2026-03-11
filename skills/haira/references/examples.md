# Haira Language — Canonical Examples

## 1. Hello World
```haira
import "io"

fn main() {
    io.println("Hello, Haira!")
}
```

## 2. Variables & Types
```haira
import "io"

fn main() {
    name = "Haira"
    age = 1
    pi = 3.14159
    active = true
    count: int = 42
    const MAX = 100
    a, b = 1, 2
    io.println("${name} is ${age} year old")
}
```

## 3. Functions
```haira
import "io"

fn add(a: int, b: int) -> int {
    return a + b
}

fn greet(name: string, greeting: string = "Hello") -> string {
    return "${greeting}, ${name}!"
}

fn fibonacci(n: int) -> int {
    if n <= 1 { return n }
    return fibonacci(n - 1) + fibonacci(n - 2)
}

fn main() {
    io.println(add(2, 3))
    io.println(greet("World"))
    io.println(fibonacci(10))
}
```

## 4. Structs & Methods
```haira
import "io"

struct User {
    name: string
    age: int
    email: string
}

User.greet() -> string {
    return "Hi, I'm ${self.name} (${self.age})"
}

User.is_adult() -> bool {
    return self.age >= 18
}

fn main() {
    user = User{ name: "Alice", age: 30, email: "alice@example.com" }
    io.println(user.greet())
}
```

## 5. Enums & Pattern Matching
```haira
import "io"

enum Color { Red, Green, Blue }

enum Shape {
    Circle(float),
    Rectangle(float, float)
}

fn area(s: Shape) -> float {
    match s {
        Shape.Circle(r) => return 3.14159 * r * r
        Shape.Rectangle(w, h) => return w * h
    }
}

fn classify(n: int) -> string {
    match n {
        0 => return "zero"
        1 | 2 | 3 => return "small"
        4..10 => return "medium"
        n if n > 100 => return "huge"
        _ => return "other"
    }
}

fn main() {
    io.println(area(Shape.Circle(5.0)))
    io.println(classify(42))
}
```

## 6. Error Handling
```haira
import "io"
import "conv"

fn divide(a: int, b: int) -> (int, Error?) {
    if b == 0 {
        return 0, Error{message: "division by zero"}
    }
    return a / b, nil
}

fn main() {
    result, err = divide(10, 3)
    if err != nil {
        io.println("Error: ${err.message}")
    } else {
        io.println("Result: ${result}")
    }

    try {
        a = divide(10, 2)?
        b = divide(a, 0)?
    } catch err {
        io.println("Caught: " + err)
    }
}
```

## 7. Pipe Operator
```haira
import "io"
import "string"

fn double(n: int) -> int { return n * 2 }
fn add_one(n: int) -> int { return n + 1 }

fn main() {
    result = 5 |> double |> add_one
    io.println(result)  // 11

    processed = "  Hello, World  "
        |> string.trim
        |> string.to_upper
    io.println(processed)  // "HELLO, WORLD"
}
```

## 8. Collections
```haira
import "io"
import "array"
import "map"

fn main() {
    fruits = ["apple", "banana", "cherry"]
    for i, fruit in fruits {
        io.println("${i}: ${fruit}")
    }

    nums = [3, 1, 4, 1, 5, 9]
    sorted = array.sort(nums)
    doubled = array.map(nums, fn(n) { n * 2 })

    scores = { "alice": 95, "bob": 82 }
    for name, score in scores {
        io.println("${name}: ${score}")
    }
}
```

## 9. Basic Agent
```haira
import "io"

provider anthropic {
    api_key: env("ANTHROPIC_API_KEY")
    model: "claude-sonnet-4-20250514"
}

tool get_weather(city: string) -> string {
    """
    Get the current weather for a city.
    Returns temperature and conditions.
    """
    resp, err = http.get("https://wttr.in/${city}?format=j1")
    if err != nil { return "Failed to fetch weather." }
    data = resp.json()
    current = data["current_condition"][0]
    return "${city}: ${current["temp_C"]}C, ${current["weatherDesc"][0]["value"]}"
}

agent Assistant {
    provider: anthropic
    system: "You are a helpful assistant. Be concise."
    tools: [get_weather]
    temperature: 0.7
}

fn main() {
    answer, err = Assistant.ask("What's the weather in Paris?")
    if err != nil {
        io.println("Error: ${err.message}")
        return
    }
    io.println(answer)
}
```

## 10. Agent with Memory & Sessions
```haira
import "io"

provider anthropic {
    api_key: env("ANTHROPIC_API_KEY")
    model: "claude-sonnet-4-20250514"
}

agent Tutor {
    provider: anthropic
    system: "You are a patient math tutor. Explain step by step."
    memory: conversation(max_turns: 30)
    temperature: 0.5
}

fn main() {
    session = "student-001"
    reply1, _ = Tutor.ask("What is a derivative?", session: session)
    io.println(reply1)
    reply2, _ = Tutor.ask("Can you give me an example?", session: session)
    io.println(reply2)
}
```

## 11. Agent Handoffs
```haira
import "io"

provider anthropic {
    api_key: env("ANTHROPIC_API_KEY")
    model: "claude-sonnet-4-20250514"
}

agent BillingAgent {
    provider: anthropic
    system: "Handle billing questions."
    memory: conversation(max_turns: 20)
}

agent TechAgent {
    provider: anthropic
    system: "Handle technical support."
    memory: conversation(max_turns: 20)
}

agent FrontDesk {
    provider: anthropic
    system: """
        Greet users. Route billing to BillingAgent.
        Route tech issues to TechAgent.
    """
    handoffs: [BillingAgent, TechAgent]
    memory: conversation(max_turns: 10)
}

// .ask() auto-follows handoffs
// .run() returns AgentResult for manual control
```

## 12. Structured Output
```haira
import "io"

provider openai {
    api_key: env("OPENAI_API_KEY")
    model: "gpt-4o"
}

struct Analysis {
    sentiment: string
    confidence: float
    topics: [string]
    summary: string
}

agent Analyzer {
    provider: openai
    system: "Analyze text and return structured data."
    output: Analysis
}

fn main() {
    result, err = Analyzer.run("I love programming in Haira!")
    if err != nil { io.println("Error: ${err}"); return }
    io.println(result)
}
```

## 13. Webhook Workflow
```haira
import "io"
import "http"

provider anthropic {
    api_key: env("ANTHROPIC_API_KEY")
    model: "claude-sonnet-4-20250514"
}

agent Assistant {
    provider: anthropic
    system: "You are a helpful assistant."
    memory: conversation(max_turns: 50)
}

@webhook("/api/chat")
workflow Chat(message: string, session_id: string) -> { reply: string } {
    reply, err = Assistant.ask(message, session: session_id)
    if err != nil { return { reply: "Something went wrong." } }
    return { reply: reply }
}

fn main() {
    server = http.Server([Chat])
    io.println("Listening on :8080")
    server.listen(8080)
}
```

## 14. Streaming Workflow
```haira
import "io"
import "http"

provider anthropic {
    api_key: env("ANTHROPIC_API_KEY")
    model: "claude-sonnet-4-20250514"
}

agent Writer {
    provider: anthropic
    system: "You are a creative writer."
    memory: conversation(max_turns: 20)
}

@webhook("/api/stream")
workflow StreamChat(message: string, session_id: string) -> stream {
    return Writer.stream(message, session: session_id)
}

fn main() {
    server = http.Server([StreamChat])
    server.listen(8080)
}
```

## 15. Parallel Execution
```haira
import "io"

results = spawn {
    Reviewer.run(ReviewRequest{ article: draft, focus: "accuracy" })
    Reviewer.run(ReviewRequest{ article: draft, focus: "clarity" })
    Reviewer.run(ReviewRequest{ article: draft, focus: "engagement" })
}

for review in reviews {
    io.println(review)
}
```

## 16. MCP Integration
```haira
// MCP provider (stdio transport)
provider filesystem {
    transport: "mcp"
    command: "npx"
    args: ["-y", "@modelcontextprotocol/server-filesystem", "/tmp"]
}

agent Assistant {
    provider: anthropic
    system: "Helpful assistant with file system access."
    tools: [greet]
    mcp: [filesystem]
    memory: conversation(max_turns: 10)
}
```

## 17. MCP Server
```haira
import "io"
import "mcp"

workflow Summarize(text: string) -> { summary: string } {
    """Summarize the given text into key points."""
    summary, err = Summarizer.ask(text)
    if err != nil { return { summary: "Error." } }
    return { summary: summary }
}

fn main() {
    mcp_server = mcp.Server([Summarize])
    mcp_server.listen(9000)
}
```

## 18. RAG Pattern
```haira
import "io"
import "vector"
import "postgres"

provider openai_embed {
    api_key: env("OPENAI_API_KEY")
    model: "text-embedding-3-small"
}

db, err = postgres.connect(env("DATABASE_URL"))
knowledge = vector.collection(db, "documents", dimensions: 1536)

tool search_docs(query: string) -> string {
    """Search the knowledge base for relevant documents."""
    results = vector.search(knowledge, {
        query: vector.embed(openai_embed, query),
        limit: 5
    })
    return vector.format(results)
}

agent RAGAssistant {
    provider: openai
    system: "Answer using the knowledge base. Always cite sources."
    tools: [search_docs]
    memory: conversation(max_turns: 20)
}
```

## 19. Workflow with Steps
```haira
@webui(title: "File Summarizer", description: "Upload a text file to summarize")
@post("/api/summarize")
workflow Summarize(document: file, context: string) -> { summary: string } {
    onerror err {
        return { summary: "Error: ${err}" }
    }

    step "Read file" {
        content, read_err = io.read_file(document)
        if read_err != nil { return { summary: "Failed to read." } }
    }

    step "Summarize" {
        reply, err = Summarizer.ask(content)
        if err != nil { return { summary: "AI error." } }
    }

    return { summary: reply }
}
```

## 20. Observability
```haira
import "io"
import "http"
import "observe"

provider openai {
    api_key: env("OPENAI_API_KEY")
    model: "gpt-4o-mini"
    input_token_cost: 0.15
    output_token_cost: 0.60
}

@get("/api/stats")
workflow Stats() -> { total_cost: float, summary: map } {
    return {
        total_cost: observe.cost(),
        summary: observe.usage()
    }
}

fn main() {
    server = http.Server([Chat, Stats])
    observe.start(server)
    observe.langfuse("pk", "sk", "https://langfuse.example.com")
    server.listen(8080)
}
```

## 21. Testing
```haira
fn add(a: int, b: int) -> int {
    return a + b
}

fn is_even(n: int) -> bool {
    return n % 2 == 0
}

test "add returns correct sum" {
    assert add(1, 2) == 3, "1+2 should equal 3"
    assert add(0, 0) == 0
    assert add(-1, 1) == 0
}

test "is_even checks parity" {
    assert is_even(4)
    assert not is_even(3)
    assert is_even(0)
}
```

## 22. Excel Spreadsheet Processing (Correct API)
```haira
import "io"
import "excel"
import "string"

// Low-level workbook API
tool read_xlsx(filepath: string, sheet_name: string = "") -> string {
    """
    Read and extract structured content from an .xlsx spreadsheet file.
    If sheet_name is empty, reads the first available sheet.
    """
    wb, err = excel.open(filepath)
    if err != nil { return "Error opening file: ${err}" }
    defer wb.close()

    // Get sheet names
    names = wb.sheet_names()
    if array.is_empty(names) { return "No sheets found" }

    // Read the specified or first sheet
    target = if sheet_name != "" { sheet_name } else { names[0] }
    rows, err = wb.read_sheet(target)
    if err != nil { return "Error reading sheet: ${err}" }

    // Format rows (each row is a map of {column: value})
    output = ""
    for i, row in rows {
        output = output + json.encode(row) + "\n"
    }
    return output
}

// High-level tables API (recommended)
tool analyze_spreadsheet(filepath: string) -> string {
    """
    Analyze an .xlsx spreadsheet using the high-level tables API.
    Returns a summary of all sheets with headers and row counts.
    """
    tables, err = excel.read_sheets(filepath)
    if err != nil { return "Error: ${err}" }

    output = "Sheets: " + string.join(tables.names(), ", ") + "\n"
    output = output + "Total sheets: " + conv.int_to_string(tables.len()) + "\n\n"

    for name in tables.names() {
        headers = tables.sheet_headers(name)
        rows = tables.sheet(name)
        output = output + "Sheet '${name}': ${array.len(headers)} columns, ${array.len(rows)} rows\n"
        output = output + "Headers: " + string.join(headers, ", ") + "\n\n"
    }

    return output
}
```

**IMPORTANT:** The excel API uses method calls on the workbook/tables object:
- `wb.close()` NOT `excel.close(wb)`
- `wb.sheet_names()` NOT `excel.get_sheet_names(wb)`
- `wb.read_sheet(name)` NOT `excel.get_sheet(wb, name)` or `excel.get_rows(sheet)`
- There is NO `excel.get_cell()`, `excel.get_sheet_by_index()`, or `excel.get_rows()`

## 23. Document Summariser (Real-World POC)
```haira
import "io"
import "http"
import "string"
import "log"
import "excel"

provider anthropic {
    backend: "anthropic"
    model: "claude-sonnet-4-20250514"
}

tool read_docx(filepath: string) -> string {
    """
    Read and extract text content from a .docx file.
    Returns the plain text content of the document.
    """
    content, err = fs.read_file(filepath)
    if err != nil { return "" }
    return content
}

tool read_xlsx(filepath: string, sheet_name: string = "") -> string {
    """
    Read structured content from an .xlsx spreadsheet file.
    If sheet_name is empty, reads the first sheet.
    """
    wb, err = excel.open(filepath)
    if err != nil { return "Error opening file: ${err}" }
    defer wb.close()

    names = wb.sheet_names()
    if array.is_empty(names) { return "No sheets found" }

    target = if sheet_name != "" { sheet_name } else { names[0] }
    rows, err = wb.read_sheet(target)
    if err != nil { return "Error reading sheet: ${err}" }

    output = ""
    for row in rows {
        line = json.encode(row)
        output = output + line + "\n"
    }
    return output
}

tool list_sheets(filepath: string) -> string {
    """
    List all sheet names in an .xlsx spreadsheet file.
    """
    wb, err = excel.open(filepath)
    if err != nil { return "Error: ${err}" }
    defer wb.close()

    names = wb.sheet_names()
    return string.join(names, "\n")
}

agent Summariser {
    provider: anthropic
    system: """
        You are an expert document summariser. When given file content, produce a clear,
        concise summary that captures the key points, structure, and important details.

        For spreadsheets (.xlsx):
        - Use list_sheets to discover all sheets in the workbook.
        - Use read_xlsx to read each relevant sheet.

        For documents (.docx), summarise the main sections and conclusions.

        Always structure your summary with:
        - A one-line overview
        - Key points as bullet points
        - Any notable data or findings
    """
    tools: [read_docx, read_xlsx, list_sheets]
    memory: conversation(max_turns: 20)
    temperature: 0.3
    max_steps: 10
}

@webui(title: "Document Summariser", description: "Upload a .docx or .xlsx file to get an AI-powered summary")
@post("/api/summarise")
workflow Summarise(document: file, message: string = "Please summarise this document.") -> { reply: string } {
    onerror err {
        return { reply: "Something went wrong: ${err}" }
    }

    step "Validate file" {
        log.info("Validating file: ${document}")
        is_docx = string.ends_with(document, ".docx")
        is_xlsx = string.ends_with(document, ".xlsx")

        if not is_docx and not is_xlsx {
            return { reply: "Please upload a .docx or .xlsx file." }
        }
    }

    step "Summarise" {
        prompt = "${message}\n\nFile: ${document}\nPlease read and summarise the uploaded file."
        reply, err = Summariser.ask(prompt, session: "summarise")
        if err != nil {
            return { reply: "Failed to summarise: ${err}" }
        }
    }

    return { reply: reply }
}

@webui(title: "Document Summariser", description: "Chat about your uploaded documents")
@post("/api/chat")
workflow Chat(message: string, session_id: string = "default") -> { reply: string } {
    onerror err {
        return { reply: "Something went wrong: ${err}" }
    }

    reply, err = Summariser.ask(message, session: session_id)
    if err != nil { return { reply: "Error: ${err}" } }
    return { reply: reply }
}

@webhook("/api/chat/stream")
workflow ChatStream(message: string, session_id: string = "default") -> stream {
    return Summariser.stream(message, session: session_id)
}

fn main() {
    server = http.Server([Summarise, Chat, ChatStream])
    io.println("Document Summariser running on :8080")
    server.listen(8080)
}
```

## 24. Dynamic Agent Creation
```haira
import "io"

provider claude {
    api_key: env("ANTHROPIC_API_KEY")
    model: "claude-sonnet-4-20250514"
}

fn create_topic_agent(topic: string) -> any {
    return create_agent({
        "name": "researcher-${topic}",
        "system": "You are an expert on ${topic}. Give concise, factual answers.",
        "memory_kind": "none"
    }, claude, nil)
}

fn main() {
    topics = ["machine learning", "robotics", "cryptography"]
    results = spawn {
        for topic in topics {
            researcher = create_topic_agent(topic)
            researcher.ask("What are the top 3 recent developments?")
        }
    }

    for i, result in results {
        io.println("--- ${topics[i]} ---")
        io.println(result)
    }
}
```

## Pattern Summary

| Pattern | Key Syntax |
|---------|-----------|
| Hello World | `fn main() { io.println("Hello") }` |
| Error handling | `result, err = fn()` + `if err != nil` |
| Structured agent | `result: Type, err = Agent.run(input)` |
| Streaming | `-> stream` + `Agent.stream()` |
| Webhook | `@webhook("/path") workflow Name(...)` |
| Steps | `step "name" { ... }` inside workflow |
| Lifecycle hooks | `onerror err { }` / `onsuccess { }` in workflow |
| File uploads | `workflow W(document: file)` + `@webui(...)` |
| Parallel | `spawn { call1(); call2() }` |
| Handoffs | `handoffs: [Agent1, Agent2]` |
| MCP client | `mcp: [provider]` in agent |
| MCP server | `mcp.Server([workflows])` |
| RAG | `vector.embed()` + `vector.search()` |
| Excel (low-level) | `excel.open()` + `wb.read_sheet()` + `wb.close()` |
| Excel (high-level) | `excel.read_sheets()` + `tables.sheet()` |
| Dynamic agents | `create_agent(config, provider, memory)` |
| Testing | `test "name" { assert condition }` |
