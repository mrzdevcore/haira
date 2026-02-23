---
title: "Functions"
description: "Function declarations, parameters, return types, and closures in Haira."
---

# Functions

## Defining Functions

Functions use the `fn` keyword:

```haira
fn add(a: int, b: int) -> int {
    return a + b
}

fn greet(name: string) -> string {
    return "Hello, ${name}!"
}
```

## Calling Functions

```haira
result = add(2, 3)        // 5
message = greet("World")   // "Hello, World!"
```

## No Return Type

Functions without a return type implicitly return `nil`:

```haira
fn log_message(msg: string) {
    io.println("[LOG] ${msg}")
}
```

## Multiple Return Values

Functions can return multiple values, commonly used for error handling:

```haira
fn divide(a: float, b: float) -> float, error {
    if b == 0.0 {
        return 0.0, "division by zero"
    }
    return a / b, nil
}

result, err = divide(10.0, 3.0)
if err != nil {
    io.println("Error: ${err}")
}
```

## Visibility

Functions are private by default. Use `pub` to export:

```haira
pub fn public_helper() -> string {
    return "accessible from other modules"
}

fn internal_helper() -> string {
    return "only accessible in this file"
}
```

## Closures and Anonymous Functions

Haira supports anonymous functions:

```haira
double = fn(x: int) -> int {
    return x * 2
}

result = double(5)  // 10
```

## Defer

Use `defer` to schedule cleanup at the end of a function:

```haira
fn process_file(path: string) {
    file = fs.open(path)
    defer fs.close(file)

    // Work with file...
    // fs.close(file) runs when this function returns
}
```

## Error Propagation

The `?` operator propagates errors by panicking, caught by `try/catch`:

```haira
fn load_config() -> Config {
    data = fs.read_file("config.json")?
    config = json.parse(data)?
    return config
}

try {
    config = load_config()
} catch err {
    io.println("Failed: ${err}")
}
```
