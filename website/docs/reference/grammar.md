---
title: "Grammar"
description: "Formal grammar specification for the Haira programming language."
---

# Grammar Reference

This is a simplified reference of Haira's grammar. For the complete formal specification, see [Chapter 17 of the spec](https://github.com/mrzdevcore/haira/tree/main/spec).

## Program Structure

```
Program         = { Declaration } EOF
Declaration     = ImportDecl | ExportDecl | ProviderDecl | ToolDecl
                | AgentDecl | WorkflowDecl | FnDecl | StructDecl
                | EnumDecl | TypeDecl | MethodDecl | TestDecl
                | VarDecl | Statement
```

## Keywords

### Core Keywords

```
fn       return   if       else     for      while
in       match    struct   enum     type     pub
import   export   from     let      const    nil
true     false    break    continue defer    errdefer
try      catch    spawn    async    select   self
and      or       not      orelse   as
```

### Agentic Keywords

```
provider  tool  agent  workflow  step
onerror   onsuccess   oncancel
```

### Testing Keywords

```
test  assert
```

### Reserved (Future)

```
trait  impl  unsafe  where  chan
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
WorkflowDecl = { Decorator } "workflow" Identifier "(" Params ")" [ "->" Type ] "{"
               { LifecycleHook }
               Body
             "}"
Decorator     = "@" Identifier [ "(" Args ")" ]
LifecycleHook = "onerror" Identifier "{" Body "}"
              | "onsuccess" "{" Body "}"
              | "oncancel" "{" Body "}"
```

### Function

```
FnDecl = [ "pub" ] "fn" Identifier "(" Params ")" [ "->" Type ] "{" Body "}"
```

### Method

```
MethodDecl = Identifier "." Identifier "(" Params ")" [ "->" Type ] "{" Body "}"
```

### Struct

```
StructDecl = [ "pub" ] "struct" Identifier "{" StructFields "}"
StructFields = { Identifier [ ":" Type ] }
```

### Enum

```
EnumDecl = [ "pub" ] "enum" Identifier "{" EnumVariants "}"
EnumVariants = { Identifier [ "(" TypeList ")" ] }
```

### Type Alias

```
TypeDecl = "type" Identifier "=" Type
```

### Test

```
TestDecl = "test" StringLit "{" Body "}"
```

## Statements

```
Statement    = AssignStmt | LetStmt | IfStmt | ForStmt | WhileStmt
             | ReturnStmt | MatchStmt | TryStmt | DeferStmt | ErrDeferStmt
             | BreakStmt | ContinueStmt | StepStmt | AssertStmt | ExprStmt

StepStmt     = [ Decorator ] "step" StringLit "{" Body "}"
AssertStmt   = "assert" Expression [ "," Expression ]
DeferStmt    = "defer" ( Expression | "{" Body "}" )
ErrDeferStmt = "errdefer" ( Expression | "{" Body "}" )
TryStmt      = "try" "{" Body "}" "catch" Identifier "{" Body "}"
```

## Expressions

```
Expression = Assignment | Binary | Unary | Call | MethodCall | Index
           | Field | Literal | Identifier | Match | If | Spawn | Select
           | Pipe | Lambda | List | Map | Instance | Range | Block
           | Propagate | Orelse | Some | None | Paren

Pipe       = Expression "|>" Expression
Binary     = Expression Op Expression
Unary      = ("!" | "-" | "~" | "not") Expression
Call       = Expression "(" Args ")"
MethodCall = Expression "." Identifier "(" Args ")"
Index      = Expression "[" Expression "]"
Field      = Expression "." Identifier
Propagate  = Expression "?"
Orelse     = Expression "orelse" Expression
Lambda     = "fn" "(" Params ")" [ "->" Type ] "{" Body "}"
Spawn      = "spawn" "{" { Expression } "}"
Instance   = Identifier "{" { Identifier ":" Expression } "}"
Range      = Expression ( ".." | "..=" ) Expression
```

## Operators (Precedence)

From highest to lowest:

| Precedence | Operators | Associativity |
|-----------|-----------|---------------|
| 1 (highest) | `()` `[]` `.` `?.` (postfix) | Left |
| 2 | `!` `-` `~` `not` (unary) | Right |
| 3 | `*` `/` `%` | Left |
| 4 | `+` `-` | Left |
| 5 | `<<` `>>` | Left |
| 6 | `&` | Left |
| 7 | `^` | Left |
| 8 | `\|` | Left |
| 9 | `..` `..=` | Left |
| 10 | `\|>` | Left |
| 11 | `<` `>` `<=` `>=` | Left |
| 12 | `==` `!=` | Left |
| 13 | `and` `&&` | Left |
| 14 | `or` `\|\|` | Left |
| 15 (lowest) | `=` `+=` `-=` `*=` `/=` `&=` `\|=` `^=` `<<=` `>>=` | Right |

## Types

```
Type = "int" | "i8" | "i16" | "i32" | "i64"
     | "u8" | "u16" | "u32" | "u64"
     | "float" | "f32" | "f64"
     | "string" | "bool" | "nil" | "file"
     | "[" "]" Type                      // List: [int]
     | "[" Type ":" Type "]"             // Map: [string:int]
     | "(" TypeList ")"                  // Tuple: (int, string)
     | Type "?"                          // Option: int?
     | "stream"                          // SSE stream
     | "fn" "(" TypeList ")" "->" Type   // Function type
     | "{" FieldList "}"                 // Anonymous struct
     | Identifier                        // Named type
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
          | Identifier "." Identifier [ "(" PatternList ")" ]  // Constructor
          | Pattern "|" Pattern                                // Or-pattern
          | Expression ".." Expression                         // Range (exclusive)
          | Expression "..=" Expression                        // Range (inclusive)
```
