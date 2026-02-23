# Modules & Imports

## Import Forms

Haira supports four import forms:

### Basic Import

```haira
import "io"
import "http"

io.println("Hello")
```

### Aliased Import

```haira
import fmt from "io"

fmt.println("Hello")
```

### Selective Import

```haira
import { User, Config } from "models"

user = User{name: "Alice"}
```

### Glob Import

```haira
import * from "math"

result = sqrt(16.0)  // No prefix needed
```

## Standard Library Modules

| Module | Description |
|--------|-------------|
| `"io"` | Console I/O (`println`, `print`) |
| `"http"` | HTTP client and server |
| `"json"` | JSON encode/decode |
| `"string"` | String manipulation |
| `"math"` | Math functions |
| `"time"` | Time and date |
| `"env"` | Environment variables |
| `"fs"` | File system operations |
| `"regex"` | Regular expressions |
| `"conv"` | Type conversions |
| `"postgres"` | PostgreSQL client |
| `"excel"` | Excel file operations |
| `"slack"` | Slack integration |

## Module Files

Multi-file projects use a `mod.haira` file to re-export:

```haira
// models/mod.haira
export { User, Config }
```

```haira
// models/user.haira
pub struct User {
    name: string
    email: string
}
```

```haira
// main.haira
import { User } from "models"
```

## Visibility Rules

- Everything is **private by default**
- Use `pub` to make items visible to importers
- Agentic declarations (`provider`, `tool`, `agent`, `workflow`) are always public
- Methods inherit the visibility of their type
- You cannot add methods to imported types

## Environment Variables

The built-in `env()` function reads environment variables:

```haira
api_key = env("API_KEY")
port = env("PORT")
```

This is commonly used in provider configurations.
