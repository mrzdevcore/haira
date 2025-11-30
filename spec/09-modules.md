# 9. Modules and Packages

## 9.1 Philosophy

Haira has **no import statements**. The compiler automatically resolves all references by scanning the project and its dependencies. Files are just organization.

## 9.2 Project Structure

```
myproject/
├── haira.config            # Project configuration
├── schema/
│   ├── user.haira
│   └── post.haira
├── core/
│   ├── auth.haira
│   └── utils.haira
├── api/
│   ├── routes.haira
│   └── handlers.haira
├── main.haira
└── tests/
    └── user_test.haira
```

## 9.3 Project Configuration

`haira.config` defines project metadata and dependencies:

```haira
name = "myproject"
version = "1.0.0"
author = "Your Name"

dependencies {
    http = "haira/http"
    json = "haira/json"
    db = "haira/postgres"
    aws = "github.com/haira-libs/aws"
}

build {
    output = "bin/myproject"
    target = "native"           // native, wasm, etc.
}

test {
    directory = "tests"
}
```

## 9.4 Automatic Resolution

### Local Files

The compiler scans all `.haira` files and resolves references:

```haira
// schema/user.haira
User { id, name, email, active }


// core/auth.haira
authenticate(token) {
    // User is automatically found from schema/user.haira
    user = get_user_by_token(token)
    user
}


// main.haira
// Both User and authenticate are available
user = authenticate(token)
print(user.name)
```

### Resolution Order

When the compiler sees an identifier:

1. **Current scope** - Local variables, parameters
2. **Current file** - Functions/types in this file
3. **Project files** - All other `.haira` files
4. **Dependencies** - External packages
5. **Auto-generation** - Can it be generated from schema?
6. **Standard library** - Built-in functions

## 9.5 Namespacing

When names conflict, use file/directory prefix:

```haira
// schema/user.haira
User { id, name, email }

// external/user.haira
User { user_id, username }

// main.haira
internal_user = schema.User { 1, "Alice", "a@mail.com" }
external_user = external.User { 1, "alice" }
```

### Explicit Namespace

```haira
// Can always be explicit
auth_result = core.auth.authenticate(token)
```

## 9.6 Visibility

### Default (Project-Wide)

By default, everything is visible within the project:

```haira
// schema/user.haira
User { name, email }          // Visible everywhere in project

create_user(name, email) {    // Visible everywhere in project
    User { name, email }
}
```

### File-Private

Prefix with `_` to restrict to current file:

```haira
// core/auth.haira

// Public
authenticate(token) {
    _validate_token(token)
    // ...
}

// Private to this file
_validate_token(token) {
    // Implementation detail
}

_secret_key = "internal"      // Private variable
```

### Public API (for Libraries)

Use `public` to expose from a package:

```haira
// In a library package

public User { name, email }           // Exported

public create_user(name, email) {     // Exported
    User { name, email }
}

validate(user) {                      // Not exported (internal)
    // ...
}
```

## 9.7 External Dependencies

### Declaring Dependencies

In `haira.config`:

```haira
dependencies {
    // Standard library packages
    http = "haira/http"
    json = "haira/json"
    db = "haira/postgres"

    // Third-party packages
    aws = "github.com/haira-libs/aws"
    stripe = "github.com/haira-libs/stripe@2.0.0"
}
```

### Using Dependencies

Just use them—no imports:

```haira
// The compiler finds these from dependencies
server = http.Server { port = 8080 }
data = json.parse(text)
bucket = aws.s3_bucket("my-bucket")
```

### Version Pinning

```haira
dependencies {
    // Latest
    http = "haira/http"

    // Specific version
    json = "haira/json@1.2.3"

    // Version range
    db = "haira/postgres@^2.0.0"

    // Git reference
    utils = "github.com/user/utils#main"
}
```

## 9.8 Standard Library

Built-in packages that don't need declaration:

| Package | Description |
|---------|-------------|
| `io` | File I/O operations |
| `net` | Networking primitives |
| `json` | JSON parsing/serialization |
| `http` | HTTP client/server |
| `time` | Date/time operations |
| `math` | Mathematical functions |
| `crypto` | Cryptographic functions |
| `os` | Operating system interface |

Usage (automatic):

```haira
// These just work
content = io.read_file(path)
data = json.parse(text)
now = time.now()
hash = crypto.sha256(data)
```

## 9.9 Creating Packages

### Package Structure

```
my-package/
├── haira.config
├── lib.haira           # Main entry (optional)
├── types.haira
├── utils.haira
└── README.md
```

### Package Config

```haira
// haira.config
name = "my-package"
version = "1.0.0"
description = "A useful package"
license = "MIT"
repository = "github.com/user/my-package"

public_modules = ["lib", "types"]   // What to export
```

### Publishing

```bash
haira publish
```

## 9.10 Conditional Compilation

Target-specific code:

```haira
// Platform-specific
#[target(windows)]
path_separator = "\\"

#[target(unix)]
path_separator = "/"

// Feature flags
#[feature(experimental)]
new_algorithm(data) {
    // ...
}
```

## 9.11 Best Practices

### Do

```haira
// Organize by domain
schema/
    user.haira
    post.haira
core/
    auth.haira
api/
    routes.haira

// Use meaningful file names
// user.haira, not u.haira

// Keep files focused
// One type per file for schemas
```

### Don't

```haira
// Don't create circular dependencies
// a.haira uses b.haira uses a.haira

// Don't put everything in one file
// Split by responsibility

// Don't use generic names
// utils.haira with 1000 lines
```
