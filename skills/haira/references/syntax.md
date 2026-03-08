# Haira Language — Complete Syntax Reference

## Lexical Structure

### Numeric Literals
```haira
42              // decimal int
1_000_000       // underscore separators
0xFF            // hexadecimal
0b1010          // binary
0o77            // octal
3.14            // float
1.5e10          // scientific notation
```

### String Literals
```haira
"hello"                    // basic string
"Hello, ${name}!"         // interpolation with ${expr}
r"no\escape\here"          // raw string
"""
    Multi-line string.
    Auto-dedented.
"""                        // triple-quoted (used in tool descriptions)
```

**Escape sequences:** `\n`, `\r`, `\t`, `\\`, `\"`, `\0`, `\x41`, `\u{1F600}`

### Comments
```haira
// single-line comment
/* block comment */
/* /* nested */ block comment */
/// doc comment
```

## Type System

### Primitive Types

| Type | Description |
|------|-------------|
| `int` | Platform integer |
| `i8`, `i16`, `i32`, `i64` | Signed integers |
| `u8`, `u16`, `u32`, `u64` | Unsigned integers |
| `float` | Platform float |
| `f32`, `f64` | Sized floats |
| `bool` | Boolean |
| `string` | UTF-8 string |
| `any` | Dynamic type |
| `void` | No value |

### Compound Types

```haira
[]int                       // array of int
[string:int]                // map of string to int
(int, string)               // tuple
int?                        // optional int
fn(int, int) -> int         // function type
chan<int>                    // channel
stream<string>              // lazy stream
```

### Type Aliases
```haira
type UserID = int
type Handler = fn(Request) -> Response
```

### Struct Definitions
```haira
struct User {
    name: string
    age: int
    email: string
}

// Instantiation
user = User{ name: "Alice", age: 30, email: "a@b.com" }

// Shorthand (field name = variable name)
name = "Bob"
email = "bob@b.com"
user = User{ name, age: 25, email }
```

### Enum Definitions
```haira
// Simple enum
enum Direction { North, South, East, West }

// With associated data
enum Shape {
    Circle(float),
    Rectangle(float, float),
    Triangle(float, float, float)
}

// Generic enum
enum Result<T> {
    Ok(T),
    Err(Error)
}
```

## Functions

### Declaration
```haira
fn name(param: Type, param: Type = default) -> ReturnType {
    // body
}
```

### Parameter Kinds
```haira
// Required
fn add(a: int, b: int) -> int { a + b }

// Default values
fn greet(name: string, greeting: string = "Hello") { /* ... */ }

// Named arguments at call site
greet(name: "Alice", greeting: "Hi")

// Variadic
fn sum(numbers: ...int) -> int { /* ... */ }
```

### Multiple Return Values
```haira
fn divide(a: int, b: int) -> (int, Error?) {
    if b == 0 { return 0, Error{message: "div by zero"} }
    return a / b, nil
}
quotient, err = divide(10, 3)
```

### Closures / Anonymous Functions
```haira
add = fn(a: int, b: int) -> int { a + b }
doubled = array.map([1,2,3], fn(n) { n * 2 })

fn make_counter() -> fn() -> int {
    count = 0
    return fn() -> int {
        count += 1
        return count
    }
}
```

### Methods
```haira
// Dot-attach syntax with implicit self
Type.method_name(params) -> ReturnType {
    self.field  // access receiver
    // self is protected (cannot reassign)
}

// Example
struct Rect { width: float, height: float }

Rect.area() -> float {
    return self.width * self.height
}

Rect.scale(factor: float) -> Rect {
    return Rect{ width: self.width * factor, height: self.height * factor }
}

rect = Rect{ width: 10.0, height: 5.0 }
rect.area()       // 50.0
rect.scale(2.0)   // Rect{width: 20.0, height: 10.0}
```

**Rules:**
- Methods inherit type visibility
- Methods on imported types are **forbidden**
- `self` is just a variable name outside methods

### Generic Functions
```haira
fn identity<T>(value: T) -> T { return value }
fn first<T>(items: []T) -> T? { /* ... */ }
```

## Control Flow

### If/Else
```haira
if condition { /* ... */ }
if condition { /* ... */ } else { /* ... */ }
if condition { /* ... */ } else if other { /* ... */ } else { /* ... */ }

// As expression (all branches must have same type, else required)
max = if a > b { a } else { b }

// Condition binding
if value = get_optional(); value != nil { /* use value */ }
```

### For Loops
```haira
for i in 0..10 { /* 0 to 9 */ }
for i in 0..=10 { /* 0 to 10 */ }
for item in collection { /* iterate */ }
for i, item in collection { /* with index */ }
for key, value in my_map { /* iterate map */ }
for char in "hello" { /* iterate string */ }
```

### While Loops
```haira
while condition { /* body */ }
while true { if done { break } }
```

### Match
```haira
match value {
    1 => io.println("one")
    2 | 3 => io.println("two or three")    // or-patterns
    4..10 => io.println("four to nine")      // range patterns
    n if n > 100 => io.println("big")        // guard
    _ => io.println("other")                 // wildcard
}

// Destructuring
match user {
    User{name: "admin", active: true} => io.println("active admin")
    User{name: n} => io.println("user: " + n)
}

match result {
    Result.Ok(value) => io.println(value)
    Result.Err(err) => io.println(err.message)
}

match point {
    (0, 0) => "origin"
    (x, 0) => "x-axis"
    (0, y) => "y-axis"
    (x, y) => "at (${x}, ${y})"
}
```

**Exhaustiveness:** Compiler warns on non-exhaustive match over enum types.

### Break & Continue
```haira
break              // exit innermost loop
break label        // exit labeled loop
continue           // skip to next iteration
continue label     // skip in labeled loop

outer: for i in 0..10 {
    for j in 0..10 {
        if i * j > 25 { break outer }
    }
}
```

## Error Handling

### Error Type
```haira
struct Error {
    message: string
    code: int?
    source: Error?
}

err = Error{message: "not found", code: 404}
wrapped = Error{message: "load failed", source: err}
```

### Tuple Return Pattern
```haira
fn read_file(path: string) -> (string, Error?) { /* ... */ }

content, err = read_file("config.toml")
if err != nil {
    io.println("Error: " + err.message)
    return
}
```

### Error Propagation `?`
```haira
// Panics on error. Use with try/catch.
content = read_file(path)?
config = parse_config(content)?
```

### Try/Catch
```haira
try {
    config = load_config("app.toml")?
    db = connect(config.db_url)?
} catch err {
    io.println("Failed: " + err)
    os.exit(1)
}
```

### Orelse
```haira
count = parse_int(input) orelse 0
config = load_config("app.toml") orelse Config{port: 8080}
```

### Defer & Errdefer
```haira
file, err = fs.open(path)
defer fs.close(file)           // always runs on scope exit

db = connect()?
errdefer db.close()            // only runs if later ? panics
```

## Concurrency

### Spawn
```haira
spawn { background_task() }
spawn process(data)

// Parallel results
results = spawn {
    task1()
    task2()
    task3()
}
```

### Channels
```haira
ch = chan<int>()               // unbuffered
ch = chan<int>(10)             // buffered

ch <- 42                       // send
value = <-ch                   // receive
value, ok = <-ch               // receive with close check

for msg in ch { /* ... */ }    // iterate until closed
close(ch)
```

### Select
```haira
select {
    msg = <-ch1 => handle(msg)
    msg = <-ch2 => handle2(msg)
    <-timeout => io.println("timeout")
    default => io.println("no messages")
}
```

### Sync Primitives
```haira
wg = sync.WaitGroup()
wg.add(n)
wg.done()
wg.wait()

mu = sync.Mutex()
mu.lock()
mu.unlock()
```

## Modules & Visibility

### Import Forms
```haira
import "io"                              // basic
import fmt from "io"                     // aliased
import { User, Post } from "models"      // selective
import * from "math"                     // glob
```

### Export (mod.haira re-exports)
```haira
export { Name1, Name2 }
```

### Visibility Rules
- **Default:** Private (file-local)
- **`pub`:** Makes declarations importable
- **Exception:** `provider`, `tool`, `agent`, `workflow` are always public

## Standard Library Functions

### io
```
io.print(s: string)
io.println(s: string)
io.printf(format: string, args: ...any)
io.readln() -> string
io.eprintln(s: string)
io.read_file(path) -> (string, Error?)
```

### string
```
string.len(s) -> int
string.is_empty(s) -> bool
string.trim(s) -> string
string.trim_left(s) -> string
string.trim_right(s) -> string
string.to_upper(s) -> string
string.to_lower(s) -> string
string.contains(s, sub) -> bool
string.starts_with(s, prefix) -> bool
string.ends_with(s, suffix) -> bool
string.index_of(s, sub) -> int
string.last_index_of(s, sub) -> int
string.split(s, sep) -> []string
string.join(parts, sep) -> string
string.substring(s, start, end) -> string
string.char_at(s, index) -> string
string.replace(s, old, new) -> string
string.replace_all(s, old, new) -> string
string.repeat(s, n) -> string
```

### array
```
array.len(arr) -> int
array.is_empty(arr) -> bool
array.first(arr) -> T?
array.last(arr) -> T?
array.get(arr, index) -> T?
array.push(arr, item) -> []T
array.pop(arr) -> ([]T, T)
array.insert(arr, index, item) -> []T
array.remove(arr, index) -> []T
array.slice(arr, start, end) -> []T
array.take(arr, n) -> []T
array.drop(arr, n) -> []T
array.contains(arr, item) -> bool
array.index_of(arr, item) -> int
array.find(arr, predicate) -> T?
array.map(arr, fn) -> []U
array.filter(arr, predicate) -> []T
array.reduce(arr, initial, fn) -> U
array.sort(arr) -> []T
array.sort_by(arr, comparator) -> []T
array.reverse(arr) -> []T
array.concat(arr1, arr2) -> []T
array.flatten(nested) -> []T
array.unique(arr) -> []T
```

### map
```
map.len(m) -> int
map.is_empty(m) -> bool
map.get(m, key) -> V?
map.has(m, key) -> bool
map.set(m, key, value) -> map
map.remove(m, key) -> map
map.keys(m) -> []K
map.values(m) -> []V
map.entries(m) -> [](K, V)
map.map_values(m, fn) -> map
map.filter(m, predicate) -> map
map.merge(m1, m2) -> map
```

### math
```
math.PI, math.E
math.abs(n) -> number
math.min(a, b) -> number
math.max(a, b) -> number
math.clamp(n, min, max) -> number
math.floor(n) -> float
math.ceil(n) -> float
math.round(n) -> float
math.pow(base, exp) -> float
math.sqrt(n) -> float
math.log(n) -> float
math.sin(x), math.cos(x), math.tan(x)
math.random() -> float
math.random_int(min, max) -> int
```

### conv
```
conv.int_to_string(n) -> string
conv.float_to_string(f) -> string
conv.bool_to_string(b) -> string
conv.string_to_int(s) -> (int, Error?)
conv.string_to_float(s) -> (float, Error?)
conv.string_to_bool(s) -> (bool, Error?)
conv.int_to_float(n) -> float
conv.float_to_int(f) -> int
conv.int_to_hex(n) -> string
conv.int_to_binary(n) -> string
conv.to_string(value) -> string
```

### json
```
json.marshal(value) -> (string, Error?)
json.unmarshal(data) -> (any, Error?)
json.encode(value) -> (string, Error?)
json.decode<T>(data) -> (T, Error?)
json.encode_pretty(value) -> (string, Error?)
```

### fs
```
fs.read_file(path) -> (string, Error?)
fs.write_file(path, content) -> Error?
fs.append_file(path, content) -> Error?
fs.exists(path) -> bool
fs.remove(path) -> Error?
fs.rename(old, new) -> Error?
fs.copy(src, dst) -> Error?
fs.mkdir(path) -> Error?
fs.mkdir_all(path) -> Error?
fs.read_dir(path) -> ([]DirEntry, Error?)
fs.stat(path) -> (FileInfo, Error?)
```

### time
```
time.now() -> Time
time.format(t, layout) -> string
time.parse(str, layout) -> (Time, Error?)
time.sleep(seconds: float)
time.since(start) -> Duration
time.after(ms) -> chan<bool>
time.tick(ms) -> chan<bool>
```

### os
```
os.getenv(key) -> string?
os.setenv(key, value)
os.environ() -> map[string]string
os.args -> []string
os.exit(code)
os.exec(cmd, args) -> (string, Error?)
```

### http
```
http.get(url) -> (Response, Error?)
http.get_with_headers(url, headers) -> (Response, Error?)
http.post(url, headers, body) -> (Response, Error?)
http.put(url, headers, body) -> (Response, Error?)
http.delete(url) -> (Response, Error?)
http.encode_uri(s) -> string
http.Server(workflows) -> Server
server.listen(port)

// Response methods
resp.json() -> any
resp.json_array() -> []any
resp.body -> string
resp.status -> int
resp.headers -> map[string]string
```

### regex
```
regex.is_match(pattern, text) -> bool
regex.find(pattern, text) -> string?
regex.find_all(pattern, text) -> []string
regex.replace(pattern, text, replacement) -> string
regex.replace_all(pattern, text, replacement) -> string
regex.captures(pattern, text) -> []string
```

### log
```
log.info(message: string)
log.warn(message: string)
log.error(message: string)
```

### vector (embeddings)
```
vector.embed(provider, text) -> []float
vector.embed_batch(provider, texts) -> [][]float
vector.collection(db, name, dimensions: int) -> Collection
vector.insert(collection, doc)
vector.search(collection, query) -> []Result
vector.format(results) -> string
```

### observe (telemetry)
```
observe.cost() -> float
observe.agent_cost(name) -> float
observe.usage() -> map
observe.start(server)
observe.langfuse(public_key, secret_key, host)
```

### mcp
```
mcp.Server(workflows) -> MCPServer
mcp_server.listen(port)    // SSE mode
mcp_server.serve()         // stdio mode
```

## Testing

```haira
test "test name" {
    assert condition
    assert condition, "failure message"
    assert value == expected
    assert value != other
    assert not bad_condition
}
```

Run with: `haira test file.haira`

## Complete EBNF Grammar (Key Productions)

```ebnf
program       = { import_stmt } { top_level } ;
top_level     = fn_decl | struct_decl | enum_decl | type_alias
              | provider_decl | tool_decl | agent_decl | workflow_decl
              | const_decl | impl_block | test_decl ;

import_stmt   = "import" ( string_lit
              | identifier "from" string_lit
              | "{" ident_list "}" "from" string_lit
              | "*" "from" string_lit ) ;

fn_decl       = "fn" identifier [ "<" type_params ">" ]
                "(" [ params ] ")" [ "->" type ] block ;

struct_decl   = "struct" identifier [ "<" type_params ">" ]
                "{" { field_decl } "}" ;

enum_decl     = "enum" identifier [ "<" type_params ">" ]
                "{" enum_variant { "," enum_variant } "}" ;

provider_decl = "provider" identifier "{" { identifier ":" expression } "}" ;

tool_decl     = "tool" identifier "(" [ params ] ")" [ "->" type ]
                "{" triple_string { statement } "}" ;

agent_decl    = "agent" identifier "{" { identifier ":" agent_value } "}" ;

workflow_decl = { decorator } "workflow" identifier
                "(" [ params ] ")" [ "->" type ] block ;

decorator     = "@" identifier [ "(" [ arguments ] ")" ] ;
```
