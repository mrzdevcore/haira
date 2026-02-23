# Methods

Methods are functions attached to types using dot-syntax.

## Defining Methods

```haira
struct User {
    name: string
    email: string
    age: int
}

User.display() -> string {
    return "${self.name} <${self.email}>"
}

User.is_adult() -> bool {
    return self.age >= 18
}
```

## Calling Methods

```haira
user = User{name: "Alice", email: "alice@example.com", age: 30}

io.println(user.display())    // "Alice <alice@example.com>"
io.println(user.is_adult())   // true
```

## Implicit `self`

Methods have an implicit `self` parameter referring to the instance. You don't declare it — it's automatically available:

```haira
User.greet(greeting: string) -> string {
    return "${greeting}, ${self.name}!"
}

user.greet("Hello")  // "Hello, Alice!"
```

`self` is protected in methods — you cannot reassign it. Outside of methods, `self` is just a regular variable name.

## Visibility

Methods inherit the visibility of their type:

```haira
pub struct Config {
    host: string
    port: int
}

// This method is public because Config is public
Config.address() -> string {
    return "${self.host}:${self.port}"
}
```

## Restrictions

- You **cannot** add methods to imported types
- Methods are defined in the same file or module as their type
- No method overloading — each method name must be unique per type
