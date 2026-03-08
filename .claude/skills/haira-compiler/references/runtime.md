# Haira Runtime Library — File Reference

## Primitive (`primitive/haira/`) — Core Runtime

| File | Purpose |
|------|---------|
| `agent.go` | Agent execution loop, tool calling, memory management |
| `provider.go` | Provider management, API routing (OpenAI, Anthropic, Ollama) |
| `tool.go` | Tool registry, JSON schema generation |
| `workflow.go` | Workflow orchestration, step execution |
| `server.go` | HTTP server, routing, middleware |
| `memory.go` | Session management, conversation history |
| `store.go` | Store interface (pluggable backends) |
| `core.go` | Core functions (Str, Len, etc.) |
| `http.go` | HTTP primitives (Request, Response) |
| `http_client.go` | HTTP client (get, post, put, delete) |
| `io.go` | I/O functions (print, println, readln) |
| `strings.go` | String functions |
| `strings_ext.go` | Extended string functions |
| `array.go` | Array functions (map, filter, reduce, sort) |
| `maps.go` | Map functions (keys, values, merge) |
| `math.go` | Math functions (abs, sqrt, pow, trig) |
| `json.go` | JSON encode/decode |
| `regex.go` | Regex functions |
| `conv.go` | Type conversions |
| `time.go` | Time functions |
| `fs.go` | File system operations |
| `env.go` | Environment variables |
| `upload.go` | File upload handling |
| `ui_*.go` | Generative UI components |
| `observe.go` | Observability/telemetry, Exporter interface |
| `mcp_client.go` | MCP client (stdio + SSE transport) |
| `mcp_server.go` | MCP server mode |
| `helpers.go` | Shared utilities (ParseJSON, StrVal) |

## Stdlib — External Integrations (tree-shaken)

| Package | Files | Description |
|---------|-------|-------------|
| `stdlib/postgres/` | `postgres.go`, `postgres_ext.go`, `store_postgres.go` | PostgreSQL client + store backend |
| `stdlib/sqlite/` | `store_sqlite.go` | SQLite store backend (auto-included for workflows/servers) |
| `stdlib/excel/` | `excel.go`, `excel_tables.go` | Excel file reading/writing |
| `stdlib/vector/` | `vector.go`, `vector_local.go` | Vector embeddings (depends on postgres) |
| `stdlib/slack/` | `slack.go`, `slack_ext.go` | Slack API integration |
| `stdlib/github/` | `github.go` | GitHub API integration |
| `stdlib/gitlab/` | `gitlab.go` | GitLab API integration |
| `stdlib/langfuse/` | `langfuse.go` | Langfuse observability exporter |

## Dependency Graph

```
vector -> postgres
excel -> postgres
sqlite: auto-included when workflows/server detected
langfuse: explicit import + observe.use() registration
```

## Key Patterns

### Store Backend Registration
```go
// In init()
func init() {
    RegisterStoreBackend("postgres", func(dsn string) Store {
        return NewPostgresStore(dsn)
    })
}
```

### Exporter Interface
```go
type Exporter interface {
    OnGeneration(LLMGeneration)
}
// Registered via observe.use() from Haira code
```

### Import Rewriting
Stdlib files use `haira-go-runtime/haira` imports at dev time.
Rewritten to `haira-generated/haira` at compile time.

### All Runtime Files Use `package haira`
Merged into single package at bundle time.
