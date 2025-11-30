# 2. Lexical Structure

## 2.1 Source Encoding

Haira source files are encoded in UTF-8.

## 2.2 Comments

```haira
// Single line comment

/*
   Multi-line
   comment
*/
```

## 2.3 Identifiers

Identifiers start with a letter or underscore, followed by letters, digits, or underscores.

```
identifier = letter ( letter | digit | "_" )*
letter     = "a".."z" | "A".."Z" | "_"
digit      = "0".."9"
```

Identifiers prefixed with `_` are file-private:

```haira
_internal_helper()    // Only visible in this file
public_function()     // Visible project-wide
```

## 2.4 Keywords

Reserved keywords:

```
if        else      for       while     return
match     true      false     none      some
and       or        not       in        async
spawn     select    try       catch     public
err       ok
```

## 2.5 Operators

### Arithmetic
```
+    Addition
-    Subtraction / Negation
*    Multiplication
/    Division
%    Modulo
```

### Comparison
```
==   Equal
!=   Not equal
<    Less than
>    Greater than
<=   Less than or equal
>=   Greater than or equal
```

### Logical
```
and   Logical AND
or    Logical OR
not   Logical NOT
```

### Assignment
```
=    Assignment
```

### Other
```
|    Pipe operator
?    Error propagation
=>   Arrow (match arms, lambdas)
..   Range
:    Type annotation / map key-value
,    Separator
.    Member access
```

## 2.6 Delimiters

```
{    }    Blocks, type definitions, maps
[    ]    Lists, index access
(    )    Function calls, grouping, parameters
```

## 2.7 Literals

### Numbers

```haira
42          // Integer
3.14        // Float
1_000_000   // With separators (ignored)
0xFF        // Hexadecimal
0b1010      // Binary
0o755       // Octal
```

### Strings

```haira
"hello"                    // Basic string
"line1\nline2"            // Escape sequences
"Hello, {name}!"          // String interpolation
```

Escape sequences:
```
\n    Newline
\t    Tab
\\    Backslash
\"    Double quote
\{    Literal brace (in interpolation)
```

### Booleans

```haira
true
false
```

### None

```haira
none    // Absence of value
```

## 2.8 Whitespace

Whitespace (spaces, tabs, newlines) separates tokens but is not significant for block structure. Haira uses braces `{}` for blocks.

## 2.9 Semicolons

Semicolons are **not used**. Statements are separated by newlines.

```haira
x = 1
y = 2
z = x + y
```
