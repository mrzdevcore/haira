//! Prompt engineering for Claude.

use haira_cir::InterpretationContext;

/// System prompt for intent interpretation.
pub const SYSTEM_PROMPT: &str = r#"You are a code generation assistant for the Haira programming language. Your task is to interpret function names and generate Canonical IR (CIR) implementations.

## Rules

1. Output ONLY valid JSON matching the CIR schema
2. Use only operations from the CIR specification
3. Generated code must be type-safe given the context
4. Prefer simple, readable implementations
5. If intent is ambiguous, indicate lower confidence
6. Never generate arbitrary code - only use CIR operations
7. **PREFER SIMPLE SOLUTIONS**: For simple requests like "return a number between X and Y", just use a literal return - don't generate complex loops
8. **LOOPS ARE FOR COLLECTIONS ONLY**: The `loop` operation iterates over an existing collection (source), not for C-style for/while loops. If you need iteration, use `map`, `filter`, or `find` on a collection.

## CIR Operations Available

### Data Access
- get_field: Get a field from an object
- get_index: Get an element by index
- set_field: Set a field on an object

### Collection Operations
- map: Transform each element
- filter: Keep elements matching predicate
- reduce: Aggregate to single value
- group_by: Group by key
- sort: Sort by key
- take: First N elements
- skip: Skip N elements
- count: Count elements
- find: Find first matching
- any: Check if any match
- all: Check if all match

### Aggregations
- sum, min, max, avg: Numeric aggregations
- max_by, min_by: By a key function

### Control Flow
- if: Conditional
- match: Pattern matching
- loop: Iterate

### Construction
- construct: Create type instance
- create_list: Create a list
- create_map: Create a map

### Primitives
- binary_op: +, -, *, /, %, ==, !=, <, >, <=, >=, and, or
- unary_op: -, not
- call: Call another function
- literal: Literal value
- var: Variable reference

### I/O (Abstract)
- db_query: Database operations (select, insert, update, delete)
- http_request: HTTP calls
- file_read: Read file
- file_write: Write file

### String Operations
- format: String interpolation with template (template is a literal string with {} placeholders, values is a map)
- concat: String concatenation (parts is an array of string values to join)

### Control
- return: Return a value

## Standard Library Functions (use with `call` operation)

These functions are built into Haira and can be called directly:

### String Functions
- `len(s)` -> int: Get string length
- `is_empty(s)` -> bool: Check if string is empty
- `upper(s)` -> string: Convert to uppercase
- `lower(s)` -> string: Convert to lowercase
- `trim(s)` -> string: Remove leading/trailing whitespace
- `slice(s, start, end)` -> string: Get substring (supports negative indices)
- `contains(s, needle)` -> bool: Check if contains substring
- `starts_with(s, prefix)` -> bool: Check prefix
- `ends_with(s, suffix)` -> bool: Check suffix
- `index_of(s, needle)` -> int: Find index (-1 if not found)
- `replace(s, old, new)` -> string: Replace all occurrences
- `repeat(s, n)` -> string: Repeat n times
- `reverse(s)` -> string: Reverse string

### Math Functions
- `abs(x)` -> int: Absolute value
- `min(a, b)` -> int: Minimum
- `max(a, b)` -> int: Maximum
- `clamp(x, min, max)` -> int: Clamp to range
- `floor(x)` -> float: Round down
- `ceil(x)` -> float: Round up
- `round(x)` -> float: Round to nearest
- `pow(base, exp)` -> float: Power
- `sqrt(x)` -> float: Square root
- `sin(x)`, `cos(x)`, `tan(x)` -> float: Trig functions
- `random_int(max)` -> int: Random int in [0, max)

### File Functions
- `file_read(path)` -> string: Read file contents
- `file_write(path, content)` -> int: Write to file (returns 0 on success)
- `file_exists(path)` -> bool: Check if file exists

### Environment Functions
- `env(name)` -> string: Get environment variable
- `exit(code)`: Exit program

### Time Functions
- `time_now()` -> int: Current time in milliseconds
- `sleep(ms)`: Sleep for milliseconds

## CRITICAL CONSTRAINTS

1. **NO method chaining**: Never use syntax like `->method` or `obj.method()`. Each operation must be a separate CIR instruction.
2. **NO invented operators**: Only use the operators listed above. Do NOT invent operators like `->substr`, `->format`, `->split`, etc.
3. **Variable references vs string literals**:
   - Variable references use object format: `{"ref": "varname"}`
   - String literals are just strings: `"hello"`
   - Numbers are just numbers: `42`, `3.14`
   - Booleans are just booleans: `true`, `false`
4. **String manipulation**: For string operations like substring, replace, split, use `call` with a plain string function name and args array:
   ```json
   {"kind": "call", "function": "substring", "args": [{"ref": "str"}, 0, 3], "result": "part"}
   {"kind": "call", "function": "replace_all", "args": [{"ref": "str"}, "-", ""], "result": "cleaned"}
   ```
   - The `function` field is a plain string with the function name (NOT a variable reference)
   - The `args` array contains values: use `{"ref": "name"}` for variables, plain strings for literals
5. **Binary operators**: The `op` field in binary_op must be one of: add, sub, mul, div, mod, eq, ne, lt, gt, le, ge, and, or
6. **Each operation produces a result variable** that subsequent operations can reference via `{"ref": "result_name"}`

## Output Format

```json
{
  "success": true,
  "interpretation": {
    "name": "function_name",
    "description": "What the function does",
    "params": [{"name": "param", "type": "Type"}],
    "returns": "ReturnType",
    "new_types": [],
    "body": [
      {"kind": "call", "function": "helper_fn", "args": ["param"], "result": "temp"},
      {"kind": "return", "value": "temp"}
    ]
  },
  "confidence": 0.95,
  "alternatives": []
}
```

## Example: Simple Return

For a function that returns a specific value (like "return an odd number between 200 and 300"), use a simple literal:

```json
{
  "success": true,
  "interpretation": {
    "name": "get_number",
    "params": [],
    "returns": "int",
    "body": [
      {"kind": "literal", "value": 201, "result": "_lit"},
      {"kind": "return", "value": {"ref": "_lit"}}
    ]
  },
  "confidence": 0.95
}
```

## Example: Factorial (Recursive Computation)

For computing factorial, use recursion. IMPORTANT rules:
1. Every function MUST end with a return statement at the TOP LEVEL of the body
2. For recursive calls, the `function` field must be the FUNCTION NAME (like "factorial"), NOT a variable name
3. The `args` array contains the VALUES to pass, using `{"ref": "varname"}` for variables

```json
{
  "success": true,
  "interpretation": {
    "name": "factorial",
    "params": [{"name": "n", "type": "int"}],
    "returns": "int",
    "body": [
      {"kind": "if", "condition": [
        {"kind": "binary_op", "op": "le", "left": {"ref": "n"}, "right": 1, "result": "_cond"}
      ], "then_ops": [
        {"kind": "literal", "value": 1, "result": "_one"}
      ], "else_ops": [
        {"kind": "binary_op", "op": "sub", "left": {"ref": "n"}, "right": 1, "result": "_n1"},
        {"kind": "call", "function": "factorial", "args": [{"ref": "_n1"}], "result": "_rec"},
        {"kind": "binary_op", "op": "mul", "left": {"ref": "n"}, "right": {"ref": "_rec"}, "result": "_result"}
      ], "result": "_if"},
      {"kind": "return", "value": {"ref": "_if"}}
    ]
  },
  "confidence": 0.95
}
```

CRITICAL: In the call operation above, `"function": "factorial"` is the function name to call recursively. The argument `{"ref": "_n1"}` passes the value of the variable `_n1`. DO NOT confuse these - `function` is always a function name string, `args` contains values to pass.

## Example: String Formatting

For a function that formats a phone number, use call operations with proper variable references:

```json
{
  "success": true,
  "interpretation": {
    "name": "format_phone",
    "params": [{"name": "number", "type": "string"}],
    "returns": "string",
    "body": [
      {"kind": "call", "function": "replace_all", "args": [{"ref": "number"}, "[^0-9]", ""], "result": "digits"},
      {"kind": "call", "function": "slice", "args": [{"ref": "digits"}, -10], "result": "last10"},
      {"kind": "call", "function": "substring", "args": [{"ref": "last10"}, 0, 3], "result": "area"},
      {"kind": "call", "function": "substring", "args": [{"ref": "last10"}, 3, 6], "result": "exchange"},
      {"kind": "call", "function": "substring", "args": [{"ref": "last10"}, 6, 10], "result": "line"},
      {"kind": "format", "template": "({}) {}-{}", "args": [{"ref": "area"}, {"ref": "exchange"}, {"ref": "line"}], "result": "formatted"},
      {"kind": "return", "value": {"ref": "formatted"}}
    ]
  },
  "confidence": 0.95
}
```

Note: Variable references use `{"ref": "name"}` format, string literals are plain strings like `"[^0-9]"`.

If you cannot interpret the function, return:
```json
{
  "success": false,
  "confidence": 0.0,
  "error": "Reason why interpretation failed"
}
```
"#;

/// Build the user prompt for a specific request.
pub fn build_user_prompt(function_name: &str, context: &InterpretationContext) -> String {
    let context_json = serde_json::to_string_pretty(context).unwrap_or_default();

    format!(
        r#"Interpret the function `{function_name}` and generate a CIR implementation.

## Context

```json
{context_json}
```

## Instructions

1. Analyze the function name to understand intent
2. Use the types in scope to determine parameters and return type
3. Generate appropriate CIR operations
4. Return valid JSON following the output format

Generate the CIR now:"#
    )
}

/// Build the user prompt for an explicit AI intent block.
///
/// This is used when the user explicitly specifies what they want via
/// the `ai` block syntax with natural language description.
pub fn build_intent_prompt(
    function_name: Option<&str>,
    intent: &str,
    params: &[(String, String)], // (name, type) pairs
    return_type: Option<&str>,
    context: &InterpretationContext,
) -> String {
    let context_json = serde_json::to_string_pretty(context).unwrap_or_default();

    let name = function_name.unwrap_or("anonymous_ai_function");

    let params_desc = if params.is_empty() {
        "No parameters".to_string()
    } else {
        params
            .iter()
            .map(|(n, t)| format!("  - {}: {}", n, t))
            .collect::<Vec<_>>()
            .join("\n")
    };

    let return_desc = return_type.unwrap_or("infer from intent");

    format!(
        r#"Generate a CIR implementation for an explicitly defined AI function.

## Function Definition

**Name**: `{name}`

**Parameters**:
{params_desc}

**Return Type**: {return_desc}

**Developer's Intent** (this is what the function should do):
```
{intent}
```

## Context (types in scope)

```json
{context_json}
```

## Instructions

1. Follow the developer's intent EXACTLY as described above
2. Use the provided parameters with their specified types
3. If a return type is specified, ensure the function returns that type
4. If return type should be inferred, determine the most appropriate type from the intent
5. Generate appropriate CIR operations to implement the described behavior
6. Be precise - the developer has explicitly stated what they want

Generate the CIR now:"#
    )
}

/// Build prompt for a simple pattern (optimization - may not need AI).
pub fn build_simple_pattern_prompt(
    pattern: &str,
    type_name: &str,
    field_name: Option<&str>,
) -> Option<haira_cir::CIRFunction> {
    use haira_cir::*;

    match pattern {
        "get_all" => Some(
            CIRFunction::new(format!("get_{}s", type_name.to_lowercase()))
                .with_description(format!("Get all {type_name} records"))
                .returning(CIRType::list(CIRType::simple(type_name)))
                .with_op(CIROperation::DbQuery {
                    query_type: DbQueryType::Select,
                    table: type_name.to_lowercase() + "s",
                    filters: vec![],
                    order_by: None,
                    limit: None,
                    result: "result".to_string(),
                })
                .with_op(CIROperation::Return {
                    value: CIRValue::var("result"),
                }),
        ),
        "get_by_id" => Some(
            CIRFunction::new(format!("get_{}_by_id", type_name.to_lowercase()))
                .with_description(format!("Get a {type_name} by ID"))
                .with_param("id", "int")
                .returning(CIRType::option(CIRType::simple(type_name)))
                .with_op(CIROperation::DbQuery {
                    query_type: DbQueryType::Select,
                    table: type_name.to_lowercase() + "s",
                    filters: vec![DbFilter {
                        field: "id".to_string(),
                        op: FilterOp::Eq,
                        value: CIRValue::var("id"),
                    }],
                    order_by: None,
                    limit: Some(1),
                    result: "result".to_string(),
                })
                .with_op(CIROperation::Return {
                    value: CIRValue::var("result"),
                }),
        ),
        "get_by_field" if field_name.is_some() => {
            let field = field_name.unwrap();
            Some(
                CIRFunction::new(format!("get_{}_by_{}", type_name.to_lowercase(), field))
                    .with_description(format!("Get a {type_name} by {field}"))
                    .with_param(field, "string") // Assume string, could be smarter
                    .returning(CIRType::option(CIRType::simple(type_name)))
                    .with_op(CIROperation::DbQuery {
                        query_type: DbQueryType::Select,
                        table: type_name.to_lowercase() + "s",
                        filters: vec![DbFilter {
                            field: field.to_string(),
                            op: FilterOp::Eq,
                            value: CIRValue::var(field),
                        }],
                        order_by: None,
                        limit: Some(1),
                        result: "result".to_string(),
                    })
                    .with_op(CIROperation::Return {
                        value: CIRValue::var("result"),
                    }),
            )
        }
        "get_filtered" if field_name.is_some() => {
            let field = field_name.unwrap();
            // Assumes field is a boolean
            Some(
                CIRFunction::new(format!("get_{}_{}", field, type_name.to_lowercase() + "s"))
                    .with_description(format!("Get all {type_name} records where {field} is true"))
                    .returning(CIRType::list(CIRType::simple(type_name)))
                    .with_op(CIROperation::DbQuery {
                        query_type: DbQueryType::Select,
                        table: type_name.to_lowercase() + "s",
                        filters: vec![DbFilter {
                            field: field.to_string(),
                            op: FilterOp::Eq,
                            value: CIRValue::Bool(true),
                        }],
                        order_by: None,
                        limit: None,
                        result: "result".to_string(),
                    })
                    .with_op(CIROperation::Return {
                        value: CIRValue::var("result"),
                    }),
            )
        }
        "save" => {
            let param_name = type_name.to_lowercase();
            Some(
                CIRFunction::new(format!("save_{}", param_name))
                    .with_description(format!("Save a {type_name}"))
                    .with_param(&param_name, type_name)
                    .returning("none")
                    .with_op(CIROperation::DbQuery {
                        query_type: DbQueryType::Insert,
                        table: type_name.to_lowercase() + "s",
                        filters: vec![],
                        order_by: None,
                        limit: None,
                        result: "_".to_string(),
                    })
                    .with_op(CIROperation::Return {
                        value: CIRValue::None,
                    }),
            )
        }
        "delete" => {
            let param_name = type_name.to_lowercase();
            Some(
                CIRFunction::new(format!("delete_{}", param_name))
                    .with_description(format!("Delete a {type_name}"))
                    .with_param(&param_name, type_name)
                    .returning("none")
                    .with_op(CIROperation::DbQuery {
                        query_type: DbQueryType::Delete,
                        table: type_name.to_lowercase() + "s",
                        filters: vec![],
                        order_by: None,
                        limit: None,
                        result: "_".to_string(),
                    })
                    .with_op(CIROperation::Return {
                        value: CIRValue::None,
                    }),
            )
        }
        "count" => Some(
            CIRFunction::new(format!("count_{}s", type_name.to_lowercase()))
                .with_description(format!("Count all {type_name} records"))
                .returning("int")
                .with_op(CIROperation::DbQuery {
                    query_type: DbQueryType::Count,
                    table: type_name.to_lowercase() + "s",
                    filters: vec![],
                    order_by: None,
                    limit: None,
                    result: "result".to_string(),
                })
                .with_op(CIROperation::Return {
                    value: CIRValue::var("result"),
                }),
        ),
        _ => None,
    }
}

/// Parse function name to extract pattern and type.
pub fn parse_function_name(name: &str) -> Option<(String, String, Option<String>)> {
    // Common patterns:
    // get_users -> (get_all, User, None)
    // get_user_by_id -> (get_by_id, User, None)
    // get_user_by_email -> (get_by_field, User, Some("email"))
    // get_active_users -> (get_filtered, User, Some("active"))
    // save_user -> (save, User, None)
    // delete_user -> (delete, User, None)
    // count_users -> (count, User, None)

    let parts: Vec<&str> = name.split('_').collect();
    if parts.len() < 2 {
        return None;
    }

    let prefix = parts[0];

    match prefix {
        "get" => {
            if parts.len() == 2 {
                // get_users -> get_all User
                let type_name = singular(parts[1]);
                return Some(("get_all".to_string(), capitalize(&type_name), None));
            }
            if parts.len() >= 4 && parts[2] == "by" {
                // get_user_by_field
                let type_name = parts[1];
                let field = parts[3..].join("_");
                if field == "id" {
                    return Some(("get_by_id".to_string(), capitalize(type_name), None));
                }
                return Some((
                    "get_by_field".to_string(),
                    capitalize(type_name),
                    Some(field),
                ));
            }
            if parts.len() == 3 {
                // get_active_users -> get_filtered User active
                let adjective = parts[1];
                let type_name = singular(parts[2]);
                return Some((
                    "get_filtered".to_string(),
                    capitalize(&type_name),
                    Some(adjective.to_string()),
                ));
            }
        }
        "save" => {
            if parts.len() == 2 {
                return Some(("save".to_string(), capitalize(parts[1]), None));
            }
        }
        "delete" => {
            if parts.len() == 2 {
                return Some(("delete".to_string(), capitalize(parts[1]), None));
            }
        }
        "count" => {
            if parts.len() == 2 {
                let type_name = singular(parts[1]);
                return Some(("count".to_string(), capitalize(&type_name), None));
            }
        }
        _ => {}
    }

    None
}

/// Convert plural to singular (simple version).
fn singular(s: &str) -> String {
    if let Some(stripped) = s.strip_suffix("ies") {
        format!("{}y", stripped)
    } else if let Some(stripped) = s.strip_suffix("es") {
        stripped.to_string()
    } else if let Some(stripped) = s.strip_suffix('s') {
        stripped.to_string()
    } else {
        s.to_string()
    }
}

/// Capitalize first letter.
fn capitalize(s: &str) -> String {
    let mut chars = s.chars();
    match chars.next() {
        None => String::new(),
        Some(c) => c.to_uppercase().chain(chars).collect(),
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parse_get_all() {
        let result = parse_function_name("get_users");
        assert_eq!(
            result,
            Some(("get_all".to_string(), "User".to_string(), None))
        );
    }

    #[test]
    fn test_parse_get_by_id() {
        let result = parse_function_name("get_user_by_id");
        assert_eq!(
            result,
            Some(("get_by_id".to_string(), "User".to_string(), None))
        );
    }

    #[test]
    fn test_parse_get_by_field() {
        let result = parse_function_name("get_user_by_email");
        assert_eq!(
            result,
            Some((
                "get_by_field".to_string(),
                "User".to_string(),
                Some("email".to_string())
            ))
        );
    }

    #[test]
    fn test_parse_get_filtered() {
        let result = parse_function_name("get_active_users");
        assert_eq!(
            result,
            Some((
                "get_filtered".to_string(),
                "User".to_string(),
                Some("active".to_string())
            ))
        );
    }

    #[test]
    fn test_parse_save() {
        let result = parse_function_name("save_user");
        assert_eq!(result, Some(("save".to_string(), "User".to_string(), None)));
    }

    #[test]
    fn test_singular() {
        assert_eq!(singular("users"), "user");
        assert_eq!(singular("companies"), "company");
        assert_eq!(singular("boxes"), "box");
    }
}
