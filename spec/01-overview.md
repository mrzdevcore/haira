# Haira Language Specification

## Version 0.1.0

---

## 1. Overview

Haira is a programming language built around one central belief:

**Software engineering should focus on expressing intention. The compiler should handle the complexity of execution.**

### 1.1 Design Goals

1. **Express intention, not mechanics** - Developers write what they want, not how to do it
2. **Natural-thinking, not natural language** - Syntax mirrors how engineers reason
3. **Fast prototyping with production-grade output** - Same code works for both
4. **Reproducibility as a core feature** - Deterministic compilation, no hidden states
5. **Native-speed binaries from high-level logic** - Compiles to fast executables
6. **Compiler absorbs complexity, developer writes clarity** - Simple surface, powerful internals

### 1.2 Target Domains

- General-purpose systems programming
- Web backends and APIs
- Data processing and ETL
- Command-line tools

### 1.3 Key Features

- **Full type inference** - Types exist but stay invisible
- **Auto-generated functions** - Compiler creates common operations from intent
- **No imports** - Compiler resolves dependencies automatically
- **No null** - Option types replace null/nil
- **Multiple return error handling** - Go-style but cleaner
- **First-class functions** - Closures and higher-order functions
- **Garbage collected** - Simple memory model
- **Native compilation** - LLVM backend for fast binaries

### 1.4 File Extension

Haira source files use the `.haira` extension.

### 1.5 Example

```haira
User { name, email, active }

server = Server { port = 8080 }

routes {
    get("/users") {
        users = get_active_users() | sort_by_name
        json(users)
    }

    post("/users") {
        user = User { body().name, body().email, true }
        save_user(user)
        json(user)
    }
}

server.start()
```

In this example:
- `User` is defined with three fields (types inferred)
- `get_active_users()`, `sort_by_name`, `save_user()` are auto-generated
- No imports, no boilerplate, just intent
