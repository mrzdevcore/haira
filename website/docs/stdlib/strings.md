---
title: "Strings"
description: "String manipulation functions — split, join, replace, trim, and more."
---

# Strings

## String Operations

```haira
import "string"

// Case
string.to_upper("hello")      // "HELLO"
string.to_lower("HELLO")      // "hello"

// Search
string.contains("hello world", "world")  // true
string.starts_with("hello", "he")        // true
string.ends_with("hello", "lo")          // true
string.index_of("hello", "ll")           // 2
string.last_index_of("hello", "l")       // 3

// Transform
string.trim("  hello  ")                     // "hello"
string.trim_left("  hello  ")                // "hello  "
string.trim_right("  hello  ")               // "  hello"
string.replace("foo bar foo", "foo", "baz")  // "baz bar foo" (first only)
string.replace_all("foo bar foo", "foo", "baz")  // "baz bar baz"

// Split & Join
string.split("a,b,c", ",")             // ["a", "b", "c"]
string.join(["a", "b", "c"], ", ")     // "a, b, c"

// Substring & Characters
string.substring("Hello World", 0, 5)  // "Hello"
string.char_at("Hello", 0)             // "H"

// Other
string.repeat("ha", 3)                 // "hahaha"
string.reverse("hello")                // "olleh"
string.capitalize("hello")             // "Hello"
string.title("hello world")            // "Hello World"
string.count("banana", "a")            // 3
string.is_empty("")                    // true
string.slugify("Hello World!")         // "hello-world"
```

## Extended String Operations

```haira
import "string"

// Padding
string.pad_left("42", 6, "0")          // "000042"
string.pad_right("hi", 6, ".")         // "hi...."

// Truncation
string.truncate("long text here", 8, "...")  // "long ..."

// Extraction
string.extract_between("key=\"value\"", "\"", "\"")  // "value"

// File paths
string.basename("/path/to/file.xlsx")   // "file.xlsx"
string.strip_ext("file.xlsx")           // "file"
string.ext("file.xlsx")                 // ".xlsx"

// Splitting
string.lines("a\nb\nc")                // ["a", "b", "c"]
string.words("hello  world")           // ["hello", "world"]
```

## String Interpolation

Haira strings use `${expr}` for interpolation:

```haira
name = "World"
greeting = "Hello, ${name}!"

// Any expression works
count = 42
msg = "Found ${count} items (${count * 2} total)"
```

## Multi-line Strings

Triple-quoted strings preserve line breaks:

```haira
prompt = """
You are a helpful assistant.
Please respond in a clear, concise manner.
Always be polite.
"""
```

## Raw Strings

For regex patterns or paths where you don't want interpolation, escape the `$`:

```haira
pattern = "\\$\\{[^}]+\\}"
```
