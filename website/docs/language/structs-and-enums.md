# Structs & Enums

## Structs

Define structured data types with the `struct` keyword:

```haira
struct User {
    name: string
    email: string
    age: int
}
```

### Creating Instances

```haira
user = User{
    name: "Alice",
    email: "alice@example.com",
    age: 30
}
```

### Accessing Fields

```haira
io.println(user.name)   // "Alice"
user.age = 31            // Update field
```

### Methods

Attach methods to structs using dot-syntax:

```haira
User.display() -> string {
    return "${self.name} <${self.email}>"
}

io.println(user.display())  // "Alice <alice@example.com>"
```

Methods have an implicit `self` parameter referring to the instance.

### Visibility

Structs are private by default. Use `pub` to export:

```haira
pub struct Config {
    host: string
    port: int
}
```

Methods inherit the visibility of their type.

## Enums

Define enumerations:

```haira
enum Color {
    Red
    Green
    Blue
}

enum Status {
    Active
    Inactive
    Pending(string)  // Variant with data
}
```

### Using Enums

```haira
color = Color.Red
status = Status.Pending("awaiting review")
```

### Pattern Matching on Enums

```haira
match color {
    Color.Red => io.println("Red"),
    Color.Green => io.println("Green"),
    Color.Blue => io.println("Blue")
}
```

The compiler warns if a `match` on an enum is not exhaustive.

### Enum with Data

```haira
enum Result {
    Ok(string)
    Error(string)
}

match result {
    Result.Ok(value) => io.println("Got: ${value}"),
    Result.Error(msg) => io.println("Error: ${msg}")
}
```

## Type Aliases

Create aliases for existing types:

```haira
type UserID = int
type Headers = map[string]string
```

## Composition over Inheritance

Haira has no classes or inheritance. Use struct composition:

```haira
struct Address {
    street: string
    city: string
}

struct Employee {
    name: string
    address: Address
    salary: float
}

employee = Employee{
    name: "Bob",
    address: Address{street: "123 Main", city: "NYC"},
    salary: 75000.0
}
```
