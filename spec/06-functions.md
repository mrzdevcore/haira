# 6. Functions

## 6.1 Function Definition

```haira
// Basic function
greet(name) {
    print("Hello, {name}!")
}

// With return value (implicit)
add(a, b) {
    a + b
}

// With explicit return
add(a, b) {
    return a + b
}

// With type annotations
add(a: int, b: int) -> int {
    a + b
}
```

## 6.2 Parameters

### Required Parameters

```haira
greet(name) {
    print("Hello, {name}")
}
```

### Default Values

```haira
greet(name, greeting = "Hello") {
    print("{greeting}, {name}")
}

greet("Alice")              // "Hello, Alice"
greet("Alice", "Hi")        // "Hi, Alice"
```

### Named Arguments

```haira
create_user(name, age, active = true) {
    User { name, age, active }
}

// Call with named arguments
user = create_user(name = "Alice", age = 30)
user = create_user(age = 30, name = "Alice")  // Order doesn't matter
```

### Rest Parameters

```haira
sum(numbers...) {
    total = 0
    for n in numbers {
        total = total + n
    }
    total
}

sum(1, 2, 3, 4, 5)          // 15
```

## 6.3 Return Values

### Single Return

```haira
double(x) {
    x * 2
}

result = double(5)          // 10
```

### Multiple Returns

```haira
divide(a, b) {
    if b == 0 {
        return none, "division by zero"
    }
    a / b, ok
}

result, err = divide(10, 2)
if err {
    print("Error: {err}")
} else {
    print("Result: {result}")
}
```

### No Return (Unit)

```haira
log(message) {
    print("[LOG] {message}")
    // No return value
}
```

## 6.4 Methods

Functions attached to types:

```haira
User { name, age }

// Method definition
User.greet() {
    "Hello, I'm {name}"
}

User.is_adult() {
    age >= 18
}

User.with_age(new_age) {
    User { name, new_age }
}

// Usage
user = User { "Alice", 30 }
print(user.greet())         // "Hello, I'm Alice"
print(user.is_adult())      // true
older = user.with_age(31)
```

### Self Reference

Inside methods, fields are accessed directly by name:

```haira
User { name, age }

User.info() {
    // 'name' and 'age' refer to this instance's fields
    "{name} ({age})"
}
```

## 6.5 First-Class Functions

Functions are values:

```haira
// Assign to variable
add = (a, b) { a + b }

// Pass as argument
apply(x, f) {
    f(x)
}

double = (n) { n * 2 }
result = apply(5, double)   // 10

// Return from function
make_multiplier(factor) {
    (x) { x * factor }
}

triple = make_multiplier(3)
triple(4)                   // 12
```

## 6.6 Lambdas / Closures

### Full Form

```haira
add = (a, b) {
    a + b
}
```

### Arrow Form (single expression)

```haira
add = (a, b) => a + b
double = x => x * 2
```

### Closures

Lambdas capture variables from enclosing scope:

```haira
make_counter() {
    count = 0
    () {
        count = count + 1
        count
    }
}

counter = make_counter()
counter()                   // 1
counter()                   // 2
counter()                   // 3
```

## 6.7 Higher-Order Functions

Functions that take or return functions:

```haira
// Takes a function
map_items(items, f) {
    result = []
    for item in items {
        result = result + [f(item)]
    }
    result
}

doubled = map_items([1, 2, 3], x => x * 2)  // [2, 4, 6]

// Returns a function
compose(f, g) {
    (x) { f(g(x)) }
}

add1 = x => x + 1
mul2 = x => x * 2
combined = compose(add1, mul2)
combined(3)                 // 7 (3 * 2 + 1)
```

## 6.8 Function Types

```haira
// Type annotation for function parameters
apply(x: int, f: (int) -> int) -> int {
    f(x)
}

// Multiple parameters
combine(a, b, f: (int, int) -> int) {
    f(a, b)
}
```

## 6.9 Recursive Functions

```haira
factorial(n) {
    if n <= 1 {
        1
    } else {
        n * factorial(n - 1)
    }
}

fibonacci(n) {
    match n {
        0 => 0
        1 => 1
        _ => fibonacci(n - 1) + fibonacci(n - 2)
    }
}
```

## 6.10 Auto-Generated Functions

When you define a type, the compiler auto-generates common functions:

```haira
User { id, name, email, active }

// Auto-generated (you don't write these):
get_users()                     // Get all
get_user_by_id(id)             // Get by primary key
get_user_by_email(email)       // Get by field
get_active_users()             // Filter by boolean
save_user(user)                // Persist
update_user(user)              // Update
delete_user(user)              // Delete
```

See [Chapter 11: Auto-Generation](11-auto-generation.md) for details.
