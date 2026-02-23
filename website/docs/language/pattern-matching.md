---
title: "Pattern Matching"
description: "Match expressions with or-patterns, range patterns, and guards in Haira."
---

# Pattern Matching

Haira's `match` expression supports powerful pattern matching.

## Basic Match

```haira
match status {
    "active" => io.println("Active"),
    "inactive" => io.println("Inactive"),
    _ => io.println("Unknown")
}
```

## Or-Patterns

Match multiple values with `|`:

```haira
match day {
    "Saturday" | "Sunday" => io.println("Weekend"),
    _ => io.println("Weekday")
}
```

## Range Patterns

Match numeric ranges:

```haira
match score {
    90..=100 => "A",
    80..90 => "B",
    70..80 => "C",
    60..70 => "D",
    _ => "F"
}
```

`..` is exclusive end, `..=` is inclusive.

## Guards

Add conditions to match arms:

```haira
match value {
    n if n > 0 => io.println("Positive"),
    n if n < 0 => io.println("Negative"),
    _ => io.println("Zero")
}
```

## Enum Matching

```haira
enum Shape {
    Circle(float)
    Rectangle(float, float)
    Point
}

match shape {
    Shape.Circle(r) => 3.14 * r * r,
    Shape.Rectangle(w, h) => w * h,
    Shape.Point => 0.0
}
```

The compiler warns if your match is not exhaustive over an enum type.

## Combining Patterns

```haira
match response_code {
    200 | 201 => "Success",
    301 | 302 => "Redirect",
    400 => "Bad Request",
    401 | 403 => "Unauthorized",
    404 => "Not Found",
    code if code >= 500 => "Server Error",
    _ => "Unknown"
}
```

## Match with Blocks

```haira
match command {
    "deploy" => {
        build()
        test()
        deploy()
        io.println("Deployed!")
    },
    "test" => {
        test()
        io.println("Tests passed!")
    },
    _ => io.println("Unknown command: ${command}")
}
```

## Codegen

The compiler generates a Go `switch` for simple matches and an `if/else if` chain when guards or range patterns are present.
