# Control Flow

## If / Else

```haira
if age >= 18 {
    io.println("Adult")
} else if age >= 13 {
    io.println("Teenager")
} else {
    io.println("Child")
}
```

Conditions don't require parentheses. Braces are always required.

## For Loops

### Iterating over ranges

```haira
for i in 0..5 {
    io.println(i)  // 0, 1, 2, 3, 4
}

for i in 0..=5 {
    io.println(i)  // 0, 1, 2, 3, 4, 5 (inclusive)
}
```

### Iterating over lists

```haira
names = ["Alice", "Bob", "Charlie"]

for name in names {
    io.println(name)
}

for i, name in names {
    io.println("${i}: ${name}")
}
```

### Iterating over maps

```haira
config = {"host": "localhost", "port": "8080"}

for key, value in config {
    io.println("${key} = ${value}")
}
```

### While-style loops

```haira
count = 0
for count < 10 {
    count += 1
}
```

### Infinite loops

```haira
for {
    msg = receive_message()
    if msg == "quit" {
        break
    }
    process(msg)
}
```

## Match

Pattern matching with `match` — Haira's replacement for switch:

```haira
match status {
    "ok" => io.println("Success"),
    "error" => io.println("Failed"),
    _ => io.println("Unknown")
}
```

### Multi-line match arms

```haira
match command {
    "start" => {
        initialize()
        run()
    },
    "stop" => {
        cleanup()
        shutdown()
    },
    _ => io.println("Unknown command")
}
```

See [Pattern Matching](/docs/language/pattern-matching) for advanced patterns including or-patterns, range patterns, and guards.

## Break and Continue

```haira
for i in 0..100 {
    if i % 2 == 0 {
        continue  // Skip even numbers
    }
    if i > 10 {
        break  // Stop at 10
    }
    io.println(i)
}
```
