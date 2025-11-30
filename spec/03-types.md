# 3. Type System

## 3.1 Philosophy

Types exist but stay invisible. The compiler infers everything. Annotations are optional, used only when you want to constrain.

## 3.2 Primitive Types

| Type | Description | Examples |
|------|-------------|----------|
| `int` | Integer numbers | `42`, `-7`, `1_000` |
| `float` | Floating-point numbers | `3.14`, `-0.5` |
| `string` | Text | `"hello"`, `"world"` |
| `bool` | Boolean | `true`, `false` |
| `none` | Absence of value | `none` |

## 3.3 Type Inference

Types are inferred from context:

```haira
name = "Alice"        // string
age = 30              // int
price = 19.99         // float
active = true         // bool
nothing = none        // none
```

## 3.4 Type Annotations

Optional annotations when needed:

```haira
// Without annotation (inferred)
x = 42

// With annotation (explicit)
x: int = 42

// Function parameter types
process(data: User) {
    // ...
}

// Return type
calculate(x, y) -> int {
    x + y
}
```

## 3.5 Collection Types

### Lists

```haira
numbers = [1, 2, 3, 4, 5]           // [int]
names = ["alice", "bob"]            // [string]
empty: [int] = []                   // Empty list with type
```

List operations:
```haira
list.first              // First element (Option)
list.last               // Last element (Option)
list.count              // Length
list.empty?             // Is empty?
list.contains(x)        // Contains element?
list[0]                 // Index access (Option)
```

### Maps

```haira
ages = { "alice": 30, "bob": 25 }   // {string: int}
config = { "debug": true }          // {string: bool}
empty: {string: int} = {}           // Empty map with type
```

Map operations:
```haira
map.keys                // List of keys
map.values              // List of values
map.has(key)            // Key exists?
map.get(key)            // Get value (Option)
map[key]                // Index access (Option)
```

## 3.6 User-Defined Types

### Basic Structure

```haira
User { name, age, email }
```

Fields are inferred from usage. Explicit types when needed:

```haira
User {
    name: string
    age: int
    email: string
    active = true           // Default value
}
```

### Instantiation

```haira
// Positional
user = User { "Alice", 30, "alice@mail.com" }

// Named
user = User { name = "Alice", age = 30, email = "alice@mail.com" }

// Mixed (positional first, then named)
user = User { "Alice", 30, email = "alice@mail.com" }
```

### Nested Types

```haira
Address { street, city, country }

User {
    name
    address: Address
}

user = User {
    name = "Alice"
    address = Address { "123 Main", "NYC", "USA" }
}
```

## 3.7 Option Type

Replaces null/nil. A value is either present (`some`) or absent (`none`).

```haira
// Function that might not return a value
find_user(id) {
    if id == 0 {
        none
    } else {
        some(user)
    }
}

// Usage - simple check
user = find_user(5)
if user {
    print(user.name)
}

// Usage - explicit unwrap
user = find_user(5)
if user.some {
    print(user.value.name)
}
```

## 3.8 Union Types

A value that can be one of several types:

```haira
Result = Success { value } | Failure { error }

// Usage
fetch_data(url) -> Result {
    if error_occurred {
        Failure { "network error" }
    } else {
        Success { data }
    }
}

// Handling with match
result = fetch_data(url)
match result {
    Success { value } => process(value)
    Failure { error } => log(error)
}
```

## 3.9 Type Aliases

Create named aliases for types:

```haira
UserId = int
Email = string
UserList = [User]
StringMap = {string: string}

// Usage
get_user(id: UserId) -> User { ... }
```

## 3.10 Structural Typing

Types are compatible based on structure, not name:

```haira
// Interface-like behavior
Printable { to_string() }

// Any type with to_string() is Printable
User { name, age }

User.to_string() {
    "{name}, age {age}"
}

// User is now compatible with Printable
print_it(p: Printable) {
    print(p.to_string())
}

print_it(user)  // Works - User has to_string()
```

## 3.11 Generic Types

Type parameters for reusable structures:

```haira
// Generic container
Box { value: T }

// Usage (type inferred)
int_box = Box { 42 }
str_box = Box { "hello" }

// Generic function
first(items: [T]) -> T {
    items[0]
}
```

## 3.12 Type Constraints

Constrain generic types:

```haira
// T must have a compare method
sort(items: [T]) where T: Comparable {
    // ...
}

// Multiple constraints
process(x: T) where T: Printable + Serializable {
    // ...
}
```
