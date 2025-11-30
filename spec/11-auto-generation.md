# 11. Auto-Generation

## 11.1 Overview

Haira's most distinctive feature is **intent-driven auto-generation**. When you call a function that doesn't exist, the compiler attempts to generate it based on:

1. Context (types in scope, schema definitions)
2. Naming conventions (semantic parsing of function names)
3. Project configuration (database, API schemas)
4. AI interpretation (for natural-language-like names)

## 11.2 How It Works

```
┌─────────────────┐
│   Haira Code    │  users = get_active_users()
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Name Parser    │  Parses: get + active + users
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Context Engine  │  Finds: User type with 'active' field
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Pattern Matcher │  Matches: get_[adjective]_[type]s pattern
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Code Generator  │  Generates: filter function for active=true
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Deterministic  │  Always produces identical output
│       IL        │
└─────────────────┘
```

## 11.3 Resolution Order

When the compiler sees `foo()`:

1. **Current scope** — Is `foo` a local variable?
2. **Current file** — Is `foo` defined in this file?
3. **Project files** — Is `foo` defined in another `.haira` file?
4. **Dependencies** — Is `foo` from an external package?
5. **Auto-generation** — Can `foo` be generated from patterns?
6. **Standard library** — Is `foo` a built-in function?
7. **Error** — Cannot resolve `foo`

## 11.4 Data Operation Patterns

Given a type definition:

```haira
User { id, name, email, age, active, created_at }
```

The compiler auto-generates:

### Retrieval

| Pattern | Generated Function |
|---------|-------------------|
| `get_users()` | Retrieve all users |
| `get_user_by_id(id)` | Retrieve by primary key |
| `get_user_by_email(email)` | Retrieve by unique field |
| `get_users_by_age(age)` | Retrieve by field value |
| `get_active_users()` | Filter by boolean field |
| `get_users_where(condition)` | Custom filter |
| `get_first_user()` | First record |
| `get_last_user()` | Last record |

### Creation & Modification

| Pattern | Generated Function |
|---------|-------------------|
| `create_user(...)` | Insert new record |
| `save_user(user)` | Upsert (insert or update) |
| `update_user(user)` | Update existing record |
| `update_user_email(user, email)` | Update specific field |
| `delete_user(user)` | Delete record |
| `delete_user_by_id(id)` | Delete by primary key |

### Counting

| Pattern | Generated Function |
|---------|-------------------|
| `count_users()` | Total count |
| `count_active_users()` | Count with filter |

## 11.5 Transformation Patterns

### Filtering

```haira
users | filter_active           // Filter where active == true
users | filter_by_age(30)       // Filter where age == 30
users | filter_where(u => u.age > 18)
```

### Sorting

```haira
users | sort_by_name            // Sort by name field
users | sort_by_age             // Sort by age field
users | sort_by_created_at      // Sort by created_at
users | sort_by(u => u.score)   // Custom sort
```

### Selection

```haira
users | take(10)                // First 10
users | skip(5)                 // Skip first 5
users | take_while(u => u.active)
users | skip_while(u => u.pending)
```

### Mapping

```haira
users | map_to_names            // Extract names
users | map_to_emails           // Extract emails
users | pluck("name")           // Same as above
```

## 11.6 I/O Patterns

### File Operations

```haira
read_file(path)                 // Read file contents
write_file(path, content)       // Write to file
append_to_file(path, content)   // Append to file
read_json_file(path)            // Read and parse JSON
write_json_file(path, data)     // Write as JSON
```

### HTTP Operations

```haira
fetch_from_url(url)             // HTTP GET
post_to_url(url, data)          // HTTP POST
fetch_user_from_api(id)         // Contextual HTTP call
```

### Serialization

```haira
to_json(data)                   // Serialize to JSON
to_csv(data)                    // Serialize to CSV
to_xml(data)                    // Serialize to XML
parse_json(text)                // Parse JSON
parse_csv(text)                 // Parse CSV
```

## 11.7 Action Patterns

### Communication

```haira
send_email_to_user(user, message)
send_notification_to_user(user, text)
send_sms_to_user(user, message)
```

### Logging

```haira
log_user_action(user, action)
log_error(error)
```

## 11.8 Naming Convention Rules

The compiler parses function names semantically:

### Prefixes

| Prefix | Meaning |
|--------|---------|
| `get_` | Retrieve data |
| `create_` | Insert new |
| `save_` | Insert or update |
| `update_` | Modify existing |
| `delete_` | Remove |
| `count_` | Count records |
| `find_` | Search (may return none) |
| `fetch_` | Retrieve from external source |
| `send_` | Transmit/communicate |
| `read_` | Read from storage |
| `write_` | Write to storage |
| `parse_` | Parse/deserialize |
| `to_` | Convert/serialize |

### Suffixes

| Suffix | Meaning |
|--------|---------|
| `_by_[field]` | Query by specific field |
| `_where` | Custom filter condition |
| `_all` | All records |
| `_first` | First matching |
| `_last` | Last matching |

### Adjectives

Boolean field names become filters:

```haira
User { active, verified, deleted }

get_active_users()      // active == true
get_verified_users()    // verified == true
get_deleted_users()     // deleted == true
```

## 11.9 Context Requirements

Auto-generation requires sufficient context:

### Works

```haira
User { name, email, active }

// Compiler knows User exists with these fields
get_active_users()              // ✓ Can generate
get_user_by_email(email)        // ✓ Can generate
sort_by_name                    // ✓ Can generate
```

### Fails

```haira
get_stuff()                     // ✗ What is "stuff"?
get_user_by_foo(x)              // ✗ User has no "foo" field
sort_by_unknown                 // ✗ No "unknown" field in context
```

## 11.10 Ambiguity Handling

When intent is ambiguous, the compiler errors:

```haira
// Error: Ambiguous - multiple types match "item"
Item { name }
OrderItem { name }

get_item_by_name("x")   // Which Item?

// Solution: Be specific
get_order_item_by_name("x")
```

## 11.11 Storage Configuration

Auto-generated data operations use configured storage:

```haira
// haira.config
storage {
    default = "postgres"

    postgres {
        url = env("DATABASE_URL")
    }

    redis {
        url = env("REDIS_URL")
    }
}
```

```haira
// Uses postgres (default)
users = get_all_users()

// Explicit storage
sessions = get_all_sessions(storage: "redis")
```

## 11.12 Generated Code Visibility

Generated code is **hidden by default**. You never see it, but it's deterministic—same input always produces same output.

### Inspection (during development)

```bash
haira inspect get_active_users
```

Shows the generated implementation.

### No Ejection

Unlike some frameworks, Haira does not support "ejecting" generated code. This ensures:

- Consistency across projects
- No divergence from patterns
- Easier upgrades

If you need custom behavior, write an explicit function.

## 11.13 Custom Overrides

Explicit definitions always take precedence:

```haira
User { id, name, active }

// Auto-generation would create this, but you override:
get_active_users() {
    // Custom implementation
    users = get_all_users()
    users | filter(u => u.active and u.verified_recently)
}
```

## 11.14 AI Interpretation Layer

For complex natural-language-like names, an optional AI layer helps:

```haira
summarize_user_activity(user)
calculate_engagement_score(user)
generate_report_for_users(users)
```

The AI layer:
1. Interprets the intent
2. Maps to deterministic operations
3. Produces canonical IL

**Important**: AI is only used for interpretation, never for execution. The output is always deterministic and reproducible.

## 11.15 Best Practices

### Do

```haira
// Use clear, conventional names
get_user_by_id(id)
get_active_users()
save_user(user)
send_email_to_user(user, msg)

// Be specific when types are similar
get_blog_post_by_id(id)     // Not just get_post_by_id
```

### Don't

```haira
// Don't use vague names
get_data()                  // What data?
do_thing()                  // What thing?

// Don't fight the conventions
fetch_user_id_from_database_by_email(e)  // Too verbose
// Use: get_user_by_email(e).id
```
