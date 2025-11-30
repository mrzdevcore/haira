# 7. Error Handling

## 7.1 Philosophy

Haira uses value-based error handling inspired by Go, but with cleaner syntax. Errors are values, not exceptions. They must be explicitly handled.

## 7.2 Multiple Return Values

Functions that can fail return two values: result and error.

```haira
divide(a, b) {
    if b == 0 {
        return none, "division by zero"
    }
    a / b, ok
}

// Usage
result, err = divide(10, 2)
if err {
    print("Error: {err}")
    return
}
print("Result: {result}")
```

## 7.3 The `ok` Keyword

`ok` indicates success (no error):

```haira
process(data) {
    if data.empty? {
        return err("empty data")
    }
    // ... process ...
    result, ok
}
```

## 7.4 The `err()` Function

Create an error:

```haira
validate(user) {
    if user.name == "" {
        return err("name is required")
    }
    if user.age < 0 {
        return err("age must be positive")
    }
    ok
}
```

## 7.5 Error Propagation with `?`

The `?` operator propagates errors automatically:

```haira
// Verbose way
process(id) {
    user, error = get_user(id)
    if error {
        return none, error
    }

    posts, error = get_posts(user)
    if error {
        return none, error
    }

    // continue...
}

// With ? operator
process(id) {
    user = get_user(id)?
    posts = get_posts(user)?
    // continue...
}
```

If the expression after `?` returns an error, the function immediately returns that error.

## 7.6 Error Types

### String Errors

Simple error messages:

```haira
return err("something went wrong")
```

### Structured Errors

Define error types for more context:

```haira
NetworkError { code: int, message: string }
ValidationError { field: string, reason: string }
NotFoundError { resource: string, id }

// Return structured error
fetch_user(id) {
    if id <= 0 {
        return err(ValidationError { "id", "must be positive" })
    }
    user = database.find(id)
    if not user {
        return err(NotFoundError { "user", id })
    }
    user, ok
}
```

### Handling Specific Errors

```haira
result, error = fetch_user(id)

match error {
    none => process(result)
    ValidationError { field, reason } => {
        print("Invalid {field}: {reason}")
    }
    NotFoundError { resource, id } => {
        print("{resource} {id} not found")
    }
    _ => {
        print("Unknown error: {error}")
    }
}
```

## 7.7 Try-Catch Blocks

Group operations that might fail:

```haira
try {
    config = load_config()?
    db = connect_database(config)?
    users = fetch_users(db)?
    process(users)
} catch error {
    log_error(error)
    fallback()
}
```

### Catch with Pattern Matching

```haira
try {
    data = fetch_from_api()?
} catch error {
    match error {
        NetworkError { code: 404 } => {
            print("Not found")
            return default_data()
        }
        NetworkError { code } if code >= 500 => {
            print("Server error, retrying...")
            retry()
        }
        _ => panic(error)
    }
}
```

## 7.8 Ignoring Errors

Explicitly ignore with `_`:

```haira
result, _ = might_fail()
// Proceed without error handling
```

Use sparingly and intentionally.

## 7.9 Must Succeed (Panic)

When failure should crash:

```haira
// Panic if error
config = load_config() or panic("config required")

// With custom message
db = connect() or panic("database connection failed")
```

## 7.10 Default Values

Provide fallback on error:

```haira
// Use default if error
config = load_config() or default_config()

// With value
port = parse_int(env("PORT")) or 8080
```

## 7.11 Chaining Fallbacks

```haira
value = try_first()
    or try_second()
    or try_third()
    or default_value

// Each is tried in order until one succeeds
```

## 7.12 Best Practices

### Do

```haira
// Handle errors explicitly
user, err = get_user(id)
if err {
    log("Failed to get user: {err}")
    return none, err
}

// Use ? for propagation in chains
process(id) {
    user = get_user(id)?
    posts = get_posts(user)?
    format(user, posts)
}

// Provide context
validate(data) {
    if data.name == "" {
        return err("validation failed: name is required")
    }
}
```

### Don't

```haira
// Don't ignore silently
result, _ = dangerous_operation()

// Don't panic on recoverable errors
config = load_config() or panic("!")  // Bad if config is optional

// Don't use generic error messages
return err("error")  // Too vague
```
