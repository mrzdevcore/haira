# Variables & Types

## Variable Declaration

Haira uses type inference — no explicit type annotations needed for local variables:

```haira
name = "Haira"        // string
age = 1               // int
pi = 3.14             // float
active = true         // bool
items = [1, 2, 3]     // []int
data = {"key": "val"} // map[string]string
```

Variables are mutable by default. Reassignment uses the same `=` syntax:

```haira
count = 0
count = count + 1
```

## Primitive Types

| Type | Description | Example |
|------|-------------|---------|
| `int` | Integer | `42` |
| `float` | Floating point | `3.14` |
| `string` | UTF-8 string | `"hello"` |
| `bool` | Boolean | `true`, `false` |
| `nil` | Null value | `nil` |

## Compound Types

### Lists

```haira
numbers = [1, 2, 3, 4, 5]
names = ["Alice", "Bob"]

// Access by index
first = numbers[0]

// List methods
numbers = append(numbers, 6)
length = len(numbers)
```

### Maps

```haira
config = {
    "host": "localhost",
    "port": "8080"
}

// Access
host = config["host"]

// Add/update
config["debug"] = "true"
```

## Strings

Strings use double quotes with `${expr}` interpolation:

```haira
name = "World"
greeting = "Hello, ${name}!"

// Multi-line with triple quotes
description = """
This is a multi-line string.
It preserves line breaks.
"""
```

### String Operations

```haira
import "string"

text = "Hello, World!"
upper = string.to_upper(text)       // "HELLO, WORLD!"
lower = string.to_lower(text)       // "hello, world!"
contains = string.contains(text, "World")  // true
parts = string.split(text, ", ")    // ["Hello", "World!"]
```

## Type Aliases

Define type aliases with the `type` keyword:

```haira
type UserID = int
type Email = string
```

## Nil and Error Handling

Functions that can fail return a value and an error:

```haira
result, err = some_operation()
if err != nil {
    io.println("Error: ${err}")
}
```

See [Error Handling](/docs/language/error-handling) for the complete error handling guide.

## Compound Assignment

```haira
x = 10
x += 5   // x = 15
x -= 3   // x = 12
x *= 2   // x = 24
x /= 4   // x = 6
```

## Bitwise Operators

```haira
a = 0b1010
b = 0b1100

c = a & b    // AND: 0b1000
d = a | b    // OR:  0b1110
e = a ^ b    // XOR: 0b0110
f = ~a       // NOT
g = a << 2   // Left shift
h = a >> 1   // Right shift
```
