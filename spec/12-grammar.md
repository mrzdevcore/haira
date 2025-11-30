# 12. Formal Grammar

## 12.1 Notation

This grammar uses Extended Backus-Naur Form (EBNF):

```
=           Definition
|           Alternation
( )         Grouping
[ ]         Optional (0 or 1)
{ }         Repetition (0 or more)
" "         Terminal string
' '         Terminal string
..          Range
(* *)       Comment
```

## 12.2 Lexical Grammar

### Characters

```ebnf
letter          = "a".."z" | "A".."Z" ;
digit           = "0".."9" ;
hex_digit       = digit | "a".."f" | "A".."F" ;
binary_digit    = "0" | "1" ;
octal_digit     = "0".."7" ;
```

### Whitespace and Comments

```ebnf
whitespace      = " " | "\t" | "\r" | "\n" ;
line_comment    = "//" { any_char - "\n" } "\n" ;
block_comment   = "/*" { any_char } "*/" ;
```

### Identifiers

```ebnf
identifier      = ( letter | "_" ) { letter | digit | "_" } ;
private_ident   = "_" identifier ;
```

### Keywords

```ebnf
keyword         = "if" | "else" | "for" | "while" | "return"
                | "match" | "true" | "false" | "none" | "some"
                | "and" | "or" | "not" | "in"
                | "async" | "spawn" | "select"
                | "try" | "catch" | "public"
                | "err" | "ok" ;
```

### Literals

```ebnf
integer         = decimal_int | hex_int | binary_int | octal_int ;
decimal_int     = digit { digit | "_" } ;
hex_int         = "0x" hex_digit { hex_digit | "_" } ;
binary_int      = "0b" binary_digit { binary_digit | "_" } ;
octal_int       = "0o" octal_digit { octal_digit | "_" } ;

float           = digit { digit } "." digit { digit } ;

string          = '"' { string_char | escape | interpolation } '"' ;
string_char     = any_char - '"' - "\\" - "{" ;
escape          = "\\" ( "n" | "t" | "r" | "\\" | '"' | "{" ) ;
interpolation   = "{" expression "}" ;

boolean         = "true" | "false" ;
none_literal    = "none" ;
```

### Operators

```ebnf
operator        = arithmetic_op | comparison_op | logical_op | other_op ;

arithmetic_op   = "+" | "-" | "*" | "/" | "%" ;
comparison_op   = "==" | "!=" | "<" | ">" | "<=" | ">=" ;
logical_op      = "and" | "or" | "not" ;
other_op        = "=" | "|" | "?" | "=>" | ".." | "..=" | ":" | "." ;
```

### Delimiters

```ebnf
delimiter       = "(" | ")" | "{" | "}" | "[" | "]" | "," ;
```

## 12.3 Syntactic Grammar

### Program

```ebnf
program         = { top_level } ;

top_level       = type_definition
                | function_definition
                | method_definition
                | type_alias
                | statement ;
```

### Type Definitions

```ebnf
type_definition = [ "public" ] identifier "{" field_list "}" ;

field_list      = [ field { newline field } ] ;

field           = identifier [ ":" type ] [ "=" expression ] ;

type_alias      = identifier "=" type ;

union_type      = identifier "=" variant { "|" variant } ;
variant         = identifier [ "{" field_list "}" ] ;
```

### Types

```ebnf
type            = simple_type
                | list_type
                | map_type
                | function_type
                | union_ref
                | generic_type ;

simple_type     = identifier ;
list_type       = "[" type "]" ;
map_type        = "{" type ":" type "}" ;
function_type   = "(" [ type_list ] ")" "->" type ;
generic_type    = identifier "<" type_list ">" ;

type_list       = type { "," type } ;
```

### Functions

```ebnf
function_definition = [ "public" ] identifier "(" [ param_list ] ")"
                      [ "->" type ] block ;

method_definition   = identifier "." identifier "(" [ param_list ] ")"
                      [ "->" type ] block ;

param_list      = param { "," param } ;
param           = identifier [ ":" type ] [ "=" expression ] [ "..." ] ;

lambda          = "(" [ param_list ] ")" block
                | "(" [ param_list ] ")" "=>" expression
                | identifier "=>" expression ;
```

### Statements

```ebnf
statement       = assignment
                | if_statement
                | for_statement
                | while_statement
                | match_statement
                | return_statement
                | try_statement
                | expression_statement
                | block ;

assignment      = pattern "=" expression ;

pattern         = identifier [ ":" type ]
                | identifier { "," identifier } ;

if_statement    = "if" expression block [ "else" ( block | if_statement ) ] ;

for_statement   = "for" for_pattern "in" expression block ;
for_pattern     = identifier [ "," identifier ] ;

while_statement = "while" expression block ;

match_statement = "match" expression "{" { match_arm } "}" ;
match_arm       = match_pattern [ "if" expression ] "=>" ( expression | block ) ;
match_pattern   = identifier [ "{" [ pattern_fields ] "}" ]
                | literal
                | "_" ;
pattern_fields  = identifier { "," identifier } ;

return_statement = "return" [ expression { "," expression } ] ;

try_statement   = "try" block "catch" identifier block ;

expression_statement = expression ;

block           = "{" { statement } "}" ;
```

### Expressions

```ebnf
expression      = assignment_expr ;

assignment_expr = pipe_expr [ "=" expression ] ;

pipe_expr       = or_expr { "|" or_expr } ;

or_expr         = and_expr { "or" and_expr } ;

and_expr        = equality_expr { "and" equality_expr } ;

equality_expr   = comparison_expr { ( "==" | "!=" ) comparison_expr } ;

comparison_expr = additive_expr { ( "<" | ">" | "<=" | ">=" ) additive_expr } ;

additive_expr   = multiplicative_expr { ( "+" | "-" ) multiplicative_expr } ;

multiplicative_expr = unary_expr { ( "*" | "/" | "%" ) unary_expr } ;

unary_expr      = ( "not" | "-" ) unary_expr
                | postfix_expr ;

postfix_expr    = primary_expr { postfix_op } ;
postfix_op      = "(" [ arg_list ] ")"       (* function call *)
                | "[" expression "]"          (* index *)
                | "." identifier              (* member access *)
                | "?" ;                       (* error propagation *)

primary_expr    = identifier
                | literal
                | list_expr
                | map_expr
                | instance_expr
                | lambda
                | match_expr
                | if_expr
                | async_expr
                | spawn_expr
                | select_expr
                | "(" expression ")"
                | "some" "(" expression ")" ;

arg_list        = arg { "," arg } ;
arg             = [ identifier "=" ] expression ;
```

### Collection Expressions

```ebnf
list_expr       = "[" [ expression { "," expression } ] "]" ;

map_expr        = "{" [ map_entry { "," map_entry } ] "}" ;
map_entry       = expression ":" expression ;

instance_expr   = identifier "{" [ instance_field { "," instance_field } ] "}" ;
instance_field  = [ identifier "=" ] expression ;
```

### Control Flow Expressions

```ebnf
if_expr         = "if" expression block [ "else" ( block | if_expr ) ] ;

match_expr      = "match" expression "{" { match_arm } "}" ;

range_expr      = expression ".." [ "=" ] expression ;
```

### Concurrency Expressions

```ebnf
async_expr      = "async" block ;

spawn_expr      = "spawn" block ;

select_expr     = "select" "{" { select_arm } [ default_arm ] "}" ;
select_arm      = identifier "from" expression "=>" ( expression | block ) ;
default_arm     = "default" "=>" ( expression | block ) ;
```

### Literals

```ebnf
literal         = integer
                | float
                | string
                | boolean
                | none_literal ;
```

## 12.4 Operator Precedence

From highest to lowest:

| Level | Operators | Associativity |
|-------|-----------|---------------|
| 1 | `()` `[]` `.` `?` | Left |
| 2 | `-` `not` (unary) | Right |
| 3 | `*` `/` `%` | Left |
| 4 | `+` `-` | Left |
| 5 | `<` `>` `<=` `>=` | Left |
| 6 | `==` `!=` | Left |
| 7 | `and` | Left |
| 8 | `or` | Left |
| 9 | `\|` (pipe) | Left |
| 10 | `=` | Right |
| 11 | `=>` | Right |

## 12.5 Automatic Semicolon Insertion

Haira does not use semicolons. Statements are terminated by newlines, with automatic continuation when:

- Line ends with an operator (`+`, `-`, `|`, etc.)
- Line ends with an opening delimiter (`{`, `[`, `(`)
- Line ends with a comma
- Next line starts with `.` (method chaining)

```haira
// Single statement (automatic continuation)
result = users
    | filter(active)
    | map(name)

// Two statements
x = 1
y = 2
```

## 12.6 Reserved for Future Use

```ebnf
reserved        = "class" | "interface" | "extends" | "implements"
                | "import" | "export" | "from" | "as"
                | "const" | "let" | "var"
                | "switch" | "case" | "break" | "continue"
                | "throw" | "finally"
                | "new" | "this" | "self" | "super"
                | "null" | "nil" | "void"
                | "await" | "yield" ;
```

These words are reserved and cannot be used as identifiers.
