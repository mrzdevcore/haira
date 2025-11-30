# 10. Standard Library

## 10.1 Overview

Haira's standard library provides essential functionality without requiring explicit imports. These functions are always available.

## 10.2 Collections

### List Operations

```haira
list = [1, 2, 3, 4, 5]

// Properties
list.first                  // Option: some(1)
list.last                   // Option: some(5)
list.count                  // 5
list.empty?                 // false

// Access
list[0]                     // Option: some(1)
list[10]                    // Option: none

// Queries
list.contains(3)            // true
list.index_of(3)            // Option: some(2)

// Transformations (return new list)
list | map(x => x * 2)                  // [2, 4, 6, 8, 10]
list | filter(x => x > 2)               // [3, 4, 5]
list | take(3)                          // [1, 2, 3]
list | skip(2)                          // [3, 4, 5]
list | reverse                          // [5, 4, 3, 2, 1]
list | sort                             // [1, 2, 3, 4, 5]
list | sort_by(x => -x)                 // [5, 4, 3, 2, 1]
list | unique                           // Remove duplicates
list | flatten                          // Flatten nested lists

// Aggregations
list | reduce(0, (a, b) => a + b)       // 15
list | sum                              // 15 (numbers only)
list | min                              // Option: some(1)
list | max                              // Option: some(5)
list | any(x => x > 3)                  // true
list | all(x => x > 0)                  // true

// Grouping
items | group_by(x => x.type)           // Map of lists

// Combining
list1 + list2                           // Concatenate
list | zip(other)                       // Pair elements
```

### Map Operations

```haira
map = { "a": 1, "b": 2, "c": 3 }

// Properties
map.keys                    // ["a", "b", "c"]
map.values                  // [1, 2, 3]
map.count                   // 3
map.empty?                  // false

// Access
map["a"]                    // Option: some(1)
map["z"]                    // Option: none
map.get("a")                // Option: some(1)
map.get("z", default: 0)    // 0

// Queries
map.has("a")                // true

// Transformations
map | filter((k, v) => v > 1)           // { "b": 2, "c": 3 }
map | map_values(v => v * 2)            // { "a": 2, "b": 4, "c": 6 }

// Combining
map1 | merge(map2)                      // Combine maps
```

## 10.3 Strings

```haira
s = "hello world"

// Properties
s.length                    // 11
s.empty?                    // false

// Case
s.upper                     // "HELLO WORLD"
s.lower                     // "hello world"
s.capitalize                // "Hello world"
s.title                     // "Hello World"

// Trimming
s.trim                      // Remove whitespace
s.trim_start                // Remove leading whitespace
s.trim_end                  // Remove trailing whitespace

// Searching
s.contains("world")         // true
s.starts_with("hello")      // true
s.ends_with("world")        // true
s.index_of("world")         // Option: some(6)

// Splitting/Joining
s.split(" ")                // ["hello", "world"]
words.join(", ")            // "hello, world"
s.lines                     // Split by newlines

// Replacing
s.replace("world", "haira") // "hello haira"
s.replace_all("l", "L")     // "heLLo worLd"

// Substrings
s.slice(0, 5)               // "hello"
s[0..5]                     // "hello"
s.chars                     // List of characters

// Formatting
"Hello, {name}!"            // String interpolation
"{x:05}"                    // Formatted (padded)
"{price:.2}"                // Formatted (decimals)
```

## 10.4 Numbers

```haira
// Integer operations
42.abs                      // 42
(-42).abs                   // 42
42.to_string                // "42"

// Float operations
3.14.floor                  // 3
3.14.ceil                   // 4
3.14.round                  // 3
3.14159.round(2)            // 3.14

// Parsing
parse_int("42")             // Option: some(42)
parse_float("3.14")         // Option: some(3.14)

// Math
math.min(a, b)              // Minimum
math.max(a, b)              // Maximum
math.pow(base, exp)         // Power
math.sqrt(x)                // Square root
math.sin(x)                 // Sine
math.cos(x)                 // Cosine
math.log(x)                 // Natural log
math.random()               // Random 0.0-1.0
math.random_int(min, max)   // Random integer in range
```

## 10.5 I/O Operations

### Console

```haira
print(value)                // Print to stdout
print("x = {x}")            // With interpolation

input()                     // Read line from stdin
input("Enter name: ")       // With prompt
```

### Files

```haira
// Reading
content = read_file(path)                   // String
bytes = read_file_bytes(path)               // Bytes
lines = read_lines(path)                    // List of strings

// Writing
write_file(path, content)                   // Write string
write_file_bytes(path, bytes)               // Write bytes
append_file(path, content)                  // Append

// File operations
exists = file_exists(path)                  // bool
delete_file(path)                           // Delete
copy_file(src, dst)                         // Copy
move_file(src, dst)                         // Move
file_size(path)                             // Size in bytes
file_modified(path)                         // Modification time

// Directories
files = list_files(dir)                     // List files
files = list_files(dir, pattern: "*.haira") // With pattern
create_dir(path)                            // Create directory
delete_dir(path)                            // Delete directory
```

## 10.6 JSON

```haira
// Parsing
data = json.parse(text)                     // Parse JSON string
data = json.parse_file(path)                // Parse JSON file

// Serialization
text = json.to_string(data)                 // Serialize to string
json.to_file(path, data)                    // Serialize to file

// Pretty printing
text = json.to_string(data, pretty: true)   // Formatted output
```

## 10.7 HTTP

### Client

```haira
// Simple requests
response = http.get(url)
response = http.post(url, body)
response = http.put(url, body)
response = http.delete(url)

// With options
response = http.get(url, {
    headers: { "Authorization": "Bearer {token}" }
    timeout: 5000
})

// Response
response.status             // 200
response.body               // Response body
response.headers            // Response headers
response.ok?                // status 200-299
```

### Server

```haira
server = http.Server { port = 8080 }

server.routes {
    get("/") {
        "Hello, World!"
    }

    get("/users/:id") {
        user = get_user_by_id(params.id)
        json(user)
    }

    post("/users") {
        data = body()
        user = create_user(data)
        json(user, status: 201)
    }
}

server.start()
```

### Request Helpers

```haira
// Inside route handlers
params.id                   // URL parameters
query.page                  // Query string
body()                      // Parse request body
header("Content-Type")      // Get header
```

### Response Helpers

```haira
json(data)                  // JSON response
json(data, status: 201)     // With status
text("hello")               // Plain text
html("<h1>Hi</h1>")        // HTML
redirect("/other")          // Redirect
error(404, "Not found")     // Error response
```

## 10.8 Time and Dates

```haira
// Current time
now = time.now()                            // Current datetime
today = time.today()                        // Current date

// Creating
date = time.date(2024, 1, 15)              // Date
datetime = time.datetime(2024, 1, 15, 10, 30, 0)

// Components
now.year                    // 2024
now.month                   // 1
now.day                     // 15
now.hour                    // 10
now.minute                  // 30
now.second                  // 0
now.weekday                 // "monday"

// Arithmetic
tomorrow = today + days(1)
next_week = today + weeks(1)
next_month = today + months(1)
later = now + hours(2) + minutes(30)

// Comparison
date1 < date2
date1 == date2

// Formatting
now.format("YYYY-MM-DD")                    // "2024-01-15"
now.format("HH:mm:ss")                      // "10:30:00"

// Parsing
time.parse("2024-01-15", "YYYY-MM-DD")

// Durations
duration = time.since(start)
elapsed = duration.seconds
elapsed = duration.milliseconds
```

## 10.9 Environment

```haira
// Environment variables
value = env("API_KEY")                      // Option
value = env("PORT") or "8080"               // With default

// Command line arguments
args = os.args                              // List of arguments

// System info
os.platform                 // "linux", "macos", "windows"
os.arch                     // "x64", "arm64"
os.cpus                     // Number of CPUs
os.memory                   // Total memory

// Process
os.exit(0)                  // Exit with code
os.pid                      // Process ID
```

## 10.10 Logging

```haira
log.debug("Debug message")
log.info("Info message")
log.warn("Warning message")
log.error("Error message")

// With values
log.info("User {user.id} logged in")

// Configure
log.level = "info"          // debug, info, warn, error
log.format = "json"         // text, json
```

## 10.11 Crypto

```haira
// Hashing
hash = crypto.sha256(data)
hash = crypto.sha512(data)
hash = crypto.md5(data)

// HMAC
sig = crypto.hmac_sha256(data, key)

// Random
bytes = crypto.random_bytes(32)
token = crypto.random_string(32)
uuid = crypto.uuid()

// Encoding
encoded = crypto.base64_encode(data)
decoded = crypto.base64_decode(encoded)
```

## 10.12 Regular Expressions

```haira
// Match
"hello".matches("[a-z]+")                   // true

// Find
"hello123".find("[0-9]+")                   // Option: some("123")
"a1b2c3".find_all("[0-9]")                  // ["1", "2", "3"]

// Replace
"hello123".replace_regex("[0-9]+", "X")     // "helloX"

// Split
"a,b;c".split_regex("[,;]")                 // ["a", "b", "c"]

// Capture groups
match = "John Doe".capture("(\\w+) (\\w+)")
match[1]                    // "John"
match[2]                    // "Doe"
```
