# Standard Library

Haira includes a rich standard library. Only imported packages are included in compiled binaries (tree-shaking).

## Core Modules

| Module | Import | Description |
|--------|--------|-------------|
| IO | `import "io"` | Console I/O (`println`, `print`) |
| HTTP | `import "http"` | HTTP client and server |
| JSON | `import "json"` | JSON encode/decode |
| String | `import "string"` | String manipulation |
| Math | `import "math"` | Math functions |
| Time | `import "time"` | Time and date |
| Env | `import "env"` | Environment variables |
| FS | `import "fs"` | File system operations |
| Regex | `import "regex"` | Regular expressions |
| Conv | `import "conv"` | Type conversions |
| Array | `import "array"` | Array/list utilities |
| Maps | `import "maps"` | Map utilities |

## Extended Modules

These are separate packages, tree-shaken at compile time:

| Module | Import | Description |
|--------|--------|-------------|
| Postgres | `import "postgres"` | PostgreSQL client |
| Excel | `import "excel"` | Excel file operations |
| Vector | `import "vector"` | Vector embeddings |
| Slack | `import "slack"` | Slack integration |
| GitHub | `import "github"` | GitHub API |
| GitLab | `import "gitlab"` | GitLab API |
| SQLite | (auto-included) | Session storage for workflows |
| Langfuse | `import "langfuse"` | LLM observability |

## Built-in Functions

Available without import:

| Function | Description |
|----------|-------------|
| `env(name)` | Read environment variable |
| `len(x)` | Length of string, list, or map |
| `append(list, item)` | Append to list |
| `print(args...)` | Print without newline |
| `println(args...)` | Print with newline |

## Automatic Inclusions

- **SQLite** is automatically included when workflows or a server are detected
- **Langfuse** requires explicit `import "langfuse"` and registration via `observe.use(langfuse.exporter(...))`
- Transitive dependencies are resolved automatically (e.g., `vector` pulls in `postgres`)
