# Error Handling

Haira uses Go-style error handling with a `value, error` pattern.

## The Pattern

Functions that can fail return a value and an error:

```haira
result, err = some_operation()
if err != nil {
    io.println("Error: ${err}")
    return
}
// Use result safely
```

## Agent Calls

All agent methods follow this pattern:

```haira
reply, err = Assistant.ask("What's the weather?")
if err != nil {
    return { error: "AI request failed." }
}
```

## HTTP Requests

```haira
resp, err = http.get("https://api.example.com/data")
if err != nil {
    io.println("Request failed: ${err}")
    return
}
data = resp.json()
```

## Error Propagation with `?`

The `?` operator panics on error, caught by `try/catch`:

```haira
fn load_data() -> string {
    resp = http.get("https://api.example.com/data")?
    data = json.parse(resp.body)?
    return data["result"]
}

try {
    result = load_data()
    io.println(result)
} catch err {
    io.println("Failed: ${err}")
}
```

Under the hood, `?` uses Go's `panic`/`recover` mechanism, and `try/catch` compiles to `defer`/`recover`.

## Nil Checking

```haira
value = map["key"]
if value == nil {
    io.println("Key not found")
}
```

## Returning Errors

```haira
fn validate(input: string) -> string, error {
    if len(input) == 0 {
        return "", "input cannot be empty"
    }
    if len(input) > 100 {
        return "", "input too long"
    }
    return input, nil
}
```

## Defer for Cleanup

Use `defer` to ensure cleanup happens regardless of errors:

```haira
fn process() {
    conn = db.connect()
    defer db.close(conn)

    // If anything fails below, conn is still closed
    data = db.query(conn, "SELECT * FROM users")?
    return data
}
```
