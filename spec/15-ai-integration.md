# 15. AI Integration Architecture

## 15.1 Overview

Haira's compiler uses AI (Claude) as a core component during compilation to interpret developer intent and generate code. This is not optional—it's fundamental to Haira's "express intention, not mechanics" philosophy.

```
┌─────────────────────────────────────────────────────────────────┐
│                     HAIRA COMPILER                              │
│                                                                 │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐        │
│  │   Lexer     │ -> │   Parser    │ -> │  Resolver   │        │
│  └─────────────┘    └─────────────┘    └─────────────┘        │
│                                              │                  │
│                                              ▼                  │
│                     ┌─────────────────────────────────┐        │
│                     │      AI INFERENCE ENGINE        │        │
│                     │         (Claude API)            │        │
│                     │                                 │        │
│                     │  • Intent interpretation        │        │
│                     │  • Code generation              │        │
│                     │  • Context understanding        │        │
│                     │  • Disambiguation               │        │
│                     └─────────────────────────────────┘        │
│                                              │                  │
│                                              ▼                  │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐        │
│  │  Type Check │ <- │  Generated  │ <- │  Canonical  │        │
│  │             │    │    Code     │    │     IR      │        │
│  └─────────────┘    └─────────────┘    └─────────────┘        │
│         │                                                       │
│         ▼                                                       │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐        │
│  │    HIR      │ -> │    MIR      │ -> │   LLVM      │        │
│  └─────────────┘    └─────────────┘    └─────────────┘        │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## 15.2 When AI is Invoked

The AI engine is called during compilation when:

### 1. Unresolved Function Calls
```haira
// Pattern-based resolution fails
// AI interprets "summarize_user_activity"
summary = summarize_user_activity(user)
```

### 2. Complex Transformations
```haira
// AI understands intent and generates pipeline
report = users | generate_engagement_report
```

### 3. Natural-Language-Like Names
```haira
// AI interprets these as concrete operations
active_premium_users = get_users_who_are_active_and_premium()
```

### 4. Ambiguous Contexts
```haira
// AI uses context to disambiguate
data = process_appropriately(input)  // AI infers from input type
```

## 15.3 AI Request Format

When the compiler needs AI assistance, it sends a structured request:

```json
{
  "request_type": "infer_intent",
  "function_name": "summarize_user_activity",
  "context": {
    "types_in_scope": [
      {
        "name": "User",
        "fields": [
          {"name": "id", "type": "int"},
          {"name": "name", "type": "string"},
          {"name": "email", "type": "string"},
          {"name": "activity_log", "type": "[Activity]"}
        ]
      },
      {
        "name": "Activity",
        "fields": [
          {"name": "type", "type": "string"},
          {"name": "timestamp", "type": "datetime"},
          {"name": "details", "type": "string"}
        ]
      }
    ],
    "call_site": {
      "file": "main.haira",
      "line": 42,
      "arguments": [{"name": "user", "type": "User"}],
      "expected_return": "unknown"
    },
    "project_schema": {
      "has_database": true,
      "has_http": false
    }
  }
}
```

## 15.4 AI Response Format

The AI returns a canonical intermediate representation:

```json
{
  "success": true,
  "interpretation": {
    "function_name": "summarize_user_activity",
    "description": "Generates a summary of a user's recent activity",
    "parameters": [
      {"name": "user", "type": "User"}
    ],
    "return_type": "ActivitySummary",
    "generated_types": [
      {
        "name": "ActivitySummary",
        "fields": [
          {"name": "total_activities", "type": "int"},
          {"name": "most_common_type", "type": "string"},
          {"name": "last_active", "type": "datetime"},
          {"name": "summary_text", "type": "string"}
        ]
      }
    ],
    "implementation": {
      "ir_version": "1.0",
      "operations": [
        {
          "op": "get_field",
          "source": "user",
          "field": "activity_log",
          "result": "activities"
        },
        {
          "op": "count",
          "source": "activities",
          "result": "total"
        },
        {
          "op": "group_by",
          "source": "activities",
          "key": "type",
          "result": "grouped"
        },
        {
          "op": "max_by_count",
          "source": "grouped",
          "result": "most_common"
        },
        {
          "op": "max_by",
          "source": "activities",
          "key": "timestamp",
          "result": "last"
        },
        {
          "op": "construct",
          "type": "ActivitySummary",
          "fields": {
            "total_activities": "total",
            "most_common_type": "most_common.key",
            "last_active": "last.timestamp",
            "summary_text": {
              "op": "format",
              "template": "{total} activities, mostly {most_common}, last active {last_active}"
            }
          },
          "result": "return"
        }
      ]
    }
  },
  "confidence": 0.95,
  "alternatives": []
}
```

## 15.5 Canonical IR (CIR)

AI outputs are converted to a **Canonical Intermediate Representation** that is:

- **Deterministic** - Same input always produces same output
- **Verifiable** - Can be type-checked and validated
- **Cacheable** - Results are cached by hash
- **Portable** - Language-agnostic representation

### CIR Operations

```
CIR_Operation =
    // Data access
    | GetField { source, field, result }
    | GetIndex { source, index, result }
    | SetField { target, field, value }

    // Collections
    | Map { source, transform, result }
    | Filter { source, predicate, result }
    | Reduce { source, initial, reducer, result }
    | GroupBy { source, key, result }
    | Sort { source, key, descending, result }
    | Take { source, count, result }
    | Count { source, result }

    // Aggregations
    | Sum { source, result }
    | Min { source, result }
    | Max { source, result }
    | Avg { source, result }

    // Control flow
    | If { condition, then_ops, else_ops, result }
    | Match { subject, arms, result }
    | Loop { items, body_ops, result }

    // Construction
    | Construct { type, fields, result }
    | CreateList { elements, result }
    | CreateMap { entries, result }

    // Primitives
    | BinaryOp { op, left, right, result }
    | UnaryOp { op, operand, result }
    | Call { function, args, result }
    | Literal { value, result }

    // I/O (abstract)
    | DbQuery { query_type, params, result }
    | HttpRequest { method, url, body, result }
    | FileRead { path, result }
    | FileWrite { path, content }
```

## 15.6 Determinism Guarantees

Even though AI is non-deterministic, Haira guarantees reproducible builds:

### 1. Caching
```
┌─────────────────────────────────────────┐
│              AI CACHE                    │
│                                          │
│  Key: hash(function_name + context)      │
│  Value: CIR + metadata                   │
│                                          │
│  Cache is:                               │
│  - Project-local (.haira-cache/)         │
│  - Version-controlled (optional)         │
│  - Shareable across team                 │
└─────────────────────────────────────────┘
```

### 2. Lock File
```haira
// haira.lock
[generated]
summarize_user_activity = "cir:sha256:abc123..."
get_engagement_score = "cir:sha256:def456..."

[version]
ai_model = "claude-3.5-sonnet"
cir_version = "1.0"
```

### 3. Regeneration
```bash
# Use cached interpretations (default)
haira build

# Force regeneration (updates cache)
haira build --refresh-ai

# Verify cache matches AI output
haira build --verify-ai
```

## 15.7 Confidence and Fallbacks

### Confidence Levels

```
HIGH (0.9-1.0):
  - AI is confident
  - Compiles without warning

MEDIUM (0.7-0.9):
  - AI has interpretation but alternatives exist
  - Compiles with info message
  - Suggests adding explicit implementation

LOW (0.5-0.7):
  - AI unsure
  - Compiles with warning
  - Strongly suggests explicit implementation

FAILED (<0.5):
  - AI cannot interpret
  - Compilation error
  - Must provide explicit implementation
```

### Example Output
```
info: AI interpreted 'summarize_user_activity' (confidence: 0.95)
  --> src/main.haira:42:15
   |
42 |     summary = summarize_user_activity(user)
   |               ^^^^^^^^^^^^^^^^^^^^^^^^^^
   |
   = AI generated: aggregates activity_log, returns ActivitySummary
   = to see generated code: haira inspect summarize_user_activity

warning: AI interpretation has medium confidence for 'calculate_engagement'
  --> src/analytics.haira:15:5
   |
15 |     score = calculate_engagement(user)
   |             ^^^^^^^^^^^^^^^^^^^^^^^^^
   |
   = confidence: 0.75
   = consider adding explicit implementation for production use
```

## 15.8 AI Engine Implementation

### Architecture

```rust
pub struct AIEngine {
    client: AnthropicClient,
    cache: AICache,
    config: AIConfig,
}

impl AIEngine {
    pub fn infer_intent(&self, request: InferRequest) -> Result<CIR> {
        // 1. Check cache
        let cache_key = request.cache_key();
        if let Some(cached) = self.cache.get(&cache_key) {
            return Ok(cached);
        }

        // 2. Build prompt
        let prompt = self.build_prompt(&request);

        // 3. Call Claude
        let response = self.client.complete(&prompt)?;

        // 4. Parse response to CIR
        let cir = self.parse_response(&response)?;

        // 5. Validate CIR
        self.validate_cir(&cir, &request.context)?;

        // 6. Cache result
        self.cache.set(&cache_key, &cir);

        Ok(cir)
    }
}
```

### Prompt Engineering

The prompt is carefully structured:

```
System: You are a code generation assistant for the Haira programming
language. Your task is to interpret function names and generate
Canonical IR (CIR) implementations.

Rules:
1. Output must be valid CIR JSON
2. Use only operations from the CIR specification
3. Generated code must be type-safe given the context
4. Prefer simple, readable implementations
5. If intent is ambiguous, ask for clarification via low confidence

Context:
{context_json}

Task: Interpret the function '{function_name}' and generate CIR.

Output format:
{output_schema}
```

## 15.9 Offline Mode

For environments without internet:

```haira
// haira.config
[ai]
mode = "offline"        // Use cache only, fail on miss
# mode = "online"       // Call AI, update cache
# mode = "hybrid"       // Try cache, fall back to AI
```

```bash
# Pre-generate all AI interpretations
haira build --cache-all

# Export cache for offline use
haira cache export > ai-cache.json

# Import cache
haira cache import < ai-cache.json
```

## 15.10 Security Considerations

### Code Review
- All AI-generated code can be inspected
- Lock file tracks all AI generations
- CI can enforce cache-only builds

### Sandboxing
- AI only generates CIR, not executable code
- CIR is validated before use
- No arbitrary code execution

### Privacy
- Only type information sent to AI
- No actual data values
- Project config can exclude sensitive schemas

## 15.11 Cost Management

```haira
// haira.config
[ai]
# Limit AI calls per build
max_calls_per_build = 100

# Cache TTL (reuse old interpretations)
cache_ttl = "30d"

# Batch similar requests
batch_requests = true

# Use cheaper model for simple patterns
model_routing = {
    simple_patterns = "claude-3-haiku"
    complex_intent = "claude-3.5-sonnet"
}
```

## 15.12 Development Workflow

### Local Development
```bash
# AI generates and caches
haira build

# See what AI generated
haira inspect get_active_users

# Override with explicit implementation
# (add function to your code, takes precedence)
```

### CI/CD Pipeline
```yaml
# .github/workflows/build.yml
- name: Build
  run: |
    # Use cached AI only (no API calls)
    haira build --offline

    # Or verify cache is current
    haira build --verify-ai
```

### Team Collaboration
```bash
# Commit AI cache with code
git add .haira-cache/
git add haira.lock

# Team members get same AI interpretations
git pull
haira build  # Uses cached interpretations
```

## 15.13 Explicit AI Intent Blocks

While AI can implicitly interpret unresolved function calls, Haira also supports **explicit AI intent blocks** for cases where developers want precise control over AI-generated code.

### Syntax

```haira
ai function_name(params) -> ReturnType {
    Natural language description of what the function should do.
    Can span multiple lines.
}
```

### Named AI Functions

Define a reusable AI-generated function at module level:

```haira
// Top-level definition
ai summarize_activity(user: User) -> ActivitySummary {
    Summarize the user's activity over the last 30 days.
    Group by activity type and find the most common.
    Return total count, most frequent type, and last active timestamp.
}

// Usage - calls the AI-generated function
main() {
    user = get_current_user()
    summary = summarize_activity(user)
    print(summary)
}
```

### Anonymous AI Blocks

Use AI inline without naming the function:

```haira
// Anonymous AI block assigned to variable
process_data(data: Data) -> Result {
    analysis = ai(d: Data) -> Stats {
        Calculate mean, median, mode, and standard deviation.
        Handle empty data gracefully.
    }
    
    // Call the anonymous function
    analysis(data)
}
```

### Type Inference

Return types can be omitted when inferrable:

```haira
// AI will infer the return type
ai find_active_users(users: [User]) {
    Filter users who have been active in the last 7 days.
    Sort by most recent activity.
}
```

### Caching Behavior

Explicit AI blocks are cached using:
- Hash of the intent text
- Function signature (name, parameters, return type)
- Types in scope

```
Cache key: sha256(intent + signature + context_types)
```

This ensures:
1. Same intent text always produces same cached result
2. Changes to intent text trigger regeneration
3. Changes to referenced types trigger regeneration

### Intent vs Implicit Resolution

| Feature | Implicit (unresolved call) | Explicit (ai block) |
|---------|---------------------------|---------------------|
| Syntax | `result = summarize(data)` | `ai summarize(data: Data) { ... }` |
| Intent | Inferred from name | Explicitly specified |
| Control | Less precise | Full control |
| Readability | Concise | Self-documenting |
| Best for | Simple, obvious functions | Complex business logic |

### Best Practices

1. **Use explicit blocks for complex logic**
   ```haira
   // Good: Complex logic is explicit
   ai generate_quarterly_report(sales: [Sale], period: Quarter) -> Report {
       Aggregate sales by region and product category.
       Calculate year-over-year growth percentages.
       Identify top performers and underperformers.
       Include trend analysis and recommendations.
   }
   ```

2. **Use implicit for simple patterns**
   ```haira
   // Good: Simple patterns are implicit
   active_users = get_active_users()
   sorted = sort_by_name(users)
   ```

3. **Be specific in intent descriptions**
   ```haira
   // Good: Specific
   ai calculate_churn_risk(user: User) -> ChurnRisk {
       Analyze user's activity patterns over last 90 days.
       Compare against historical churn indicators.
       Weight factors: login frequency (30%), feature usage (40%), support tickets (30%).
       Return risk score 0-100 with contributing factors.
   }
   
   // Bad: Vague
   ai calculate_churn_risk(user: User) {
       Calculate if user might leave.
   }
   ```

## 15.14 Future: Local AI

For fully offline operation, Haira will support local models:

```haira
// haira.config
[ai]
provider = "local"
model_path = "~/.haira/models/haira-intent-7b"

# Or use Ollama
provider = "ollama"
model = "haira-intent"
```

This is planned for future versions once smaller, specialized models are available.
