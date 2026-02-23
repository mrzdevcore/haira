# Pipe Operator

The pipe operator `|>` chains function calls, passing the result of the left side as the first argument to the right side.

## Basic Usage

```haira
result = "hello world"
    |> string.to_upper
    |> string.trim
```

This is equivalent to:

```haira
result = string.trim(string.to_upper("hello world"))
```

## Data Processing Pipelines

```haira
import "string"

output = raw_input
    |> string.trim
    |> string.to_lower
    |> string.replace(" ", "-")
```

## With Custom Functions

```haira
fn double(x: int) -> int { return x * 2 }
fn add_one(x: int) -> int { return x + 1 }

result = 5 |> double |> add_one  // 11
```

## In Workflows

```haira
@post("/process")
workflow Process(text: string) -> { result: string } {
    output = text
        |> string.trim
        |> string.to_lower
        |> validate
        |> transform

    return { result: output }
}
```

## Important

The pipe operator is `|>`, not `|`. The single `|` is the bitwise OR operator.
