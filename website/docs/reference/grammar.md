# Grammar Reference

This is a simplified reference of Haira's grammar. For the complete formal specification, see [Chapter 17 of the spec](https://github.com/mrzdevcore/haira/tree/main/spec).

## Program Structure

```
Program         = { Declaration } EOF
Declaration     = ImportDecl | ProviderDecl | ToolDecl | AgentDecl
                | WorkflowDecl | FnDecl | StructDecl | EnumDecl
                | TypeDecl | VarDecl | MethodDecl | ExportDecl
```

## Keywords

### Core Keywords

```
fn       return   if       else     for      in
match    struct   enum     type     pub      import
export   from     let      nil      true     false
break    continue defer    try      catch    spawn
self
```

### Agentic Keywords

```
provider  tool  agent  workflow
```

## Declarations

### Provider

```
ProviderDecl = "provider" Identifier "{" ProviderFields "}"
ProviderFields = { Identifier ":" Expression }
```

### Tool

```
ToolDecl = "tool" Identifier "(" Params ")" "->" Type "{" TripleQuoteString Body "}"
```

### Agent

```
AgentDecl = "agent" Identifier "{" AgentFields "}"
AgentFields = { Identifier ":" Expression }
```

### Workflow

```
WorkflowDecl = { Decorator } "workflow" Identifier "(" Params ")" [ "->" Type ] "{" Body "}"
Decorator = "@" Identifier [ "(" Args ")" ]
```

### Function

```
FnDecl = [ "pub" ] "fn" Identifier "(" Params ")" [ "->" Type ] "{" Body "}"
```

### Struct

```
StructDecl = [ "pub" ] "struct" Identifier "{" StructFields "}"
StructFields = { Identifier ":" Type }
```

### Enum

```
EnumDecl = [ "pub" ] "enum" Identifier "{" EnumVariants "}"
EnumVariants = { Identifier [ "(" TypeList ")" ] }
```

## Expressions

```
Expression = Assignment | Binary | Unary | Call | Index | Field
           | Literal | Identifier | Match | Spawn | Pipe

Pipe       = Expression "|>" Expression
Binary     = Expression Op Expression
Unary      = ("!" | "-" | "~") Expression
Call       = Expression "(" Args ")"
Index      = Expression "[" Expression "]"
Field      = Expression "." Identifier
```

## Operators (Precedence)

| Precedence | Operators | Associativity |
|-----------|-----------|---------------|
| 1 (lowest) | `\|>` | Left |
| 2 | `\|\|` | Left |
| 3 | `&&` | Left |
| 4 | `\|` | Left |
| 5 | `^` | Left |
| 6 | `&` | Left |
| 7 | `==` `!=` | Left |
| 8 | `<` `>` `<=` `>=` | Left |
| 9 | `<<` `>>` | Left |
| 10 | `+` `-` | Left |
| 11 | `*` `/` `%` | Left |
| 12 (highest) | `!` `-` `~` (unary) | Right |

## Types

```
Type = "int" | "float" | "string" | "bool" | "nil" | "file"
     | "[]" Type                    // List
     | "map[" Type "]" Type         // Map
     | "{" FieldList "}"            // Anonymous struct
     | "stream"                     // SSE stream
     | Identifier                   // Named type
```

## Import Forms

```
ImportDecl = "import" StringLit                           // Basic
           | "import" Identifier "from" StringLit         // Aliased
           | "import" "{" IdentList "}" "from" StringLit  // Selective
           | "import" "*" "from" StringLit                // Glob
```

## Match Expression

```
MatchExpr = "match" Expression "{" { MatchArm } "}"
MatchArm  = Pattern [ "if" Expression ] "=>" Expression ","
Pattern   = Literal | Identifier | "_"
          | Pattern "|" Pattern                           // Or-pattern
          | Expression ".." Expression                    // Range (exclusive)
          | Expression "..=" Expression                   // Range (inclusive)
```
