# 4. Expressions

## 4.1 Overview

Everything in Haira that produces a value is an expression. Expressions can be combined and composed.

## 4.2 Literals

```haira
42                      // Integer
3.14                    // Float
"hello"                 // String
true                    // Boolean
false                   // Boolean
none                    // None
[1, 2, 3]              // List
{ "a": 1, "b": 2 }     // Map
```

## 4.3 Identifiers

```haira
x                       // Variable reference
user                    // Variable reference
User                    // Type reference
```

## 4.4 Arithmetic Expressions

```haira
a + b                   // Addition
a - b                   // Subtraction
a * b                   // Multiplication
a / b                   // Division
a % b                   // Modulo
-a                      // Negation
```

Operator precedence (highest to lowest):
1. `-` (unary negation)
2. `*`, `/`, `%`
3. `+`, `-`

```haira
2 + 3 * 4               // 14 (not 20)
(2 + 3) * 4             // 20
```

## 4.5 Comparison Expressions

```haira
a == b                  // Equal
a != b                  // Not equal
a < b                   // Less than
a > b                   // Greater than
a <= b                  // Less than or equal
a >= b                  // Greater than or equal
```

All comparison operators return `bool`.

## 4.6 Logical Expressions

```haira
a and b                 // Logical AND
a or b                  // Logical OR
not a                   // Logical NOT
```

Short-circuit evaluation:
- `a and b`: `b` not evaluated if `a` is false
- `a or b`: `b` not evaluated if `a` is true

## 4.7 String Interpolation

```haira
name = "Alice"
age = 30

greeting = "Hello, {name}!"                    // "Hello, Alice!"
info = "{name} is {age} years old"            // "Alice is 30 years old"
calc = "Sum: {1 + 2}"                          // "Sum: 3"
```

## 4.8 Member Access

```haira
user.name               // Access field
user.address.city       // Nested access
list.count              // Property
```

## 4.9 Index Access

```haira
list[0]                 // First element (returns Option)
map["key"]              // Map lookup (returns Option)
matrix[i][j]            // Nested access
```

## 4.10 Function Calls

```haira
print("hello")                      // Simple call
add(1, 2)                           // Multiple arguments
user.greet()                        // Method call
process()                           // No arguments
create_user(name = "Alice", age = 30)  // Named arguments
```

## 4.11 Pipe Expressions

Chain operations left-to-right:

```haira
// Without pipes
result = take(sort(filter(users, active), by_name), 10)

// With pipes
result = users
    | filter(active)
    | sort(by_name)
    | take(10)
```

The left side becomes the first argument of the right side:

```haira
x | f           // f(x)
x | f(y)        // f(x, y)
x | f(y, z)     // f(x, y, z)
```

## 4.12 Lambda Expressions

Anonymous functions:

```haira
// Full form
add = (a, b) { a + b }

// Single expression with arrow
double = x => x * 2

// Multiple parameters with arrow
add = (a, b) => a + b

// In-line usage
numbers | map(x => x * 2)
users | filter(u => u.active)
```

## 4.13 Match Expressions

Pattern matching:

```haira
result = match value {
    0 => "zero"
    1 => "one"
    n => "other: {n}"
}

// With types
match result {
    Success { value } => process(value)
    Failure { error } => handle(error)
}

// With guards
match age {
    n if n < 0 => "invalid"
    n if n < 18 => "minor"
    n if n < 65 => "adult"
    _ => "senior"
}
```

## 4.14 Conditional Expressions

If as expression:

```haira
// Statement form
if condition {
    do_something()
}

// Expression form (returns value)
status = if active { "on" } else { "off" }

// Chained
category = if age < 13 {
    "child"
} else if age < 20 {
    "teen"
} else {
    "adult"
}
```

## 4.15 Range Expressions

```haira
0..10                   // 0 to 9 (exclusive end)
0..=10                  // 0 to 10 (inclusive end)
start..end              // Variable range

// Usage
for i in 0..10 {
    print(i)
}

list[0..5]              // Slice (first 5 elements)
```

## 4.16 Error Propagation

The `?` operator propagates errors:

```haira
// Without ?
process(id) {
    user, err = get_user(id)
    if err {
        return none, err
    }
    posts, err = get_posts(user)
    if err {
        return none, err
    }
    // continue...
}

// With ?
process(id) {
    user = get_user(id)?
    posts = get_posts(user)?
    // continue...
}
```

If the expression returns an error, the function immediately returns that error.

## 4.17 Block Expressions

Blocks are expressions; they return their last value:

```haira
result = {
    x = calculate()
    y = transform(x)
    x + y               // This is returned
}
```

## 4.18 Construction Expressions

Create instances:

```haira
// Type instantiation
user = User { "Alice", 30, "alice@mail.com" }
user = User { name = "Alice", age = 30 }

// List construction
numbers = [1, 2, 3, 4, 5]

// Map construction
ages = { "alice": 30, "bob": 25 }
```
