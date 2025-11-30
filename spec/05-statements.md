# 5. Statements

## 5.1 Overview

Statements perform actions and control program flow. Unlike expressions, statements don't produce values.

## 5.2 Assignment

```haira
// Simple assignment (type inferred)
x = 42
name = "Alice"
active = true

// With type annotation
x: int = 42
name: string = "Alice"

// Destructuring assignment
a, b = get_pair()
user, err = fetch_user(id)
first, rest... = list
```

## 5.3 If Statement

```haira
// Simple if
if condition {
    do_something()
}

// If-else
if condition {
    do_this()
} else {
    do_that()
}

// If-else chain
if x < 0 {
    print("negative")
} else if x == 0 {
    print("zero")
} else {
    print("positive")
}
```

## 5.4 For Loop

### Iteration

```haira
// Over list
for item in items {
    process(item)
}

// Over map (key-value pairs)
for key, value in map {
    print("{key}: {value}")
}

// Over range
for i in 0..10 {
    print(i)
}

// With index
for i, item in items {
    print("{i}: {item}")
}
```

### Loop Control

```haira
for item in items {
    if item.skip {
        continue        // Skip to next iteration
    }
    if item.done {
        break           // Exit loop
    }
    process(item)
}
```

## 5.5 While Loop

```haira
while condition {
    do_something()
}

// Infinite loop
while true {
    msg = receive()
    if msg == "quit" {
        break
    }
    process(msg)
}
```

## 5.6 Match Statement

```haira
match value {
    0 => print("zero")
    1 => print("one")
    _ => print("other")
}

// With blocks
match command {
    "start" => {
        init()
        run()
    }
    "stop" => {
        cleanup()
        shutdown()
    }
    _ => print("unknown")
}

// Type matching
match result {
    Success { value } => {
        process(value)
        log("success")
    }
    Failure { error } => {
        log_error(error)
        retry()
    }
}
```

## 5.7 Return Statement

```haira
// Single return value
calculate(x) {
    return x * 2
}

// Implicit return (last expression)
calculate(x) {
    x * 2
}

// Multiple return values
divide(a, b) {
    if b == 0 {
        return none, "division by zero"
    }
    return a / b, ok
}

// Early return
process(user) {
    if not user.active {
        return
    }
    // continue processing...
}
```

## 5.8 Try-Catch Statement

```haira
try {
    data = fetch_data()?
    result = process(data)?
    save(result)?
} catch error {
    log("Error: {error}")
    fallback()
}

// With specific error types
try {
    connect()?
} catch error {
    match error {
        NetworkError { code } => retry(code)
        AuthError { msg } => prompt_login()
        _ => panic(error)
    }
}
```

## 5.9 Block Statement

Group statements:

```haira
{
    x = 1
    y = 2
    print(x + y)
}
```

## 5.10 Expression Statement

Any expression can be a statement:

```haira
print("hello")          // Function call
user.save()             // Method call
x + y                   // Expression (value discarded)
```
