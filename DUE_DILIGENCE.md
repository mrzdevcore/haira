# Haira Due Diligence — Issue Tracker

**Generated:** 2026-03-06
**Updated:** 2026-03-07
**Total findings:** 213 (40 critical, 41 high, 75 medium, 57 low)

Legend: `[ ]` = pending, `[x]` = done, `[-]` = skipped/wontfix

---

## 1. Compiler — Type Checker (CRITICAL — most impactful)

These are the highest-leverage fixes. The checker is essentially a pass-through for many constructs — everything resolves to `AnyType{}`, so the compiler relies on Go's type checker as a safety net.

- [x] **CHK-01** Option type (`T?`) not in type system — `OptionType` in AST but `resolveASTType()` has no case; silently becomes `AnyType{}` (`checker.go:1016-1052`)
- [x] **CHK-02** `UnionType` not handled in `resolveASTType()` — silently becomes `AnyType{}` (`checker.go:1016-1052`)
- [x] **CHK-03** `GenericType` not handled in `resolveASTType()` — silently becomes `AnyType{}` (`checker.go:1016-1052`)
- [x] **CHK-04** `TupleType` not handled in `resolveASTType()` — silently becomes `AnyType{}` (`checker.go:1016-1052`)
- [x] **CHK-05** `TypeAlias` not registered in `registerGlobals()` — aliases are undefined at check time (`checker.go:112-189`)
- [x] **CHK-06** Method calls always return `AnyType{}` — no method registry, no argument validation, no return type inference (`checker.go:756-761`)
- [x] **CHK-07** Instance creation not validated — no check for required/extra fields or field type match (`checker.go:828-836`)
- [x] **CHK-08** `BreakStmt`/`ContinueStmt` not checked — `break` at top-level passes checker, generates invalid Go (`checker.go:515-700`)
- [ ] **CHK-09** `SelectExpr` not type-checked — channel types, bindings, default block all skipped (`checker.go:707-908`)
- [x] **CHK-10** Bitwise operators (`&`, `|`, `^`, `<<`, `>>`) not type-checked — fall through to `AnyType{}` (`checker.go:926-952`)
- [x] **CHK-11** Unary `~` (BitNot) not validated — silently accepts any operand type (`checker.go:737-751`)
- [x] **CHK-12** Pattern types not validated — `LiteralPattern`, `ConstructorPattern`, `RangePattern` never checked against subject type (`checker.go:72-106`)
- [x] **CHK-13** Match binding variables always `AnyType{}` — bound vars from patterns never typed (`checker.go:651-667`)
- [x] **CHK-14** Return type mismatches are warnings, not errors — uses `addWarning()` instead of `addError()` (`checker.go:609-634`)
- [x] **CHK-15** Parameter/field default values never type-checked — `fn foo(x: int = "hello")` passes (`checker.go:129-160, 115-120`)
- [x] **CHK-16** Enum variant fields never validated — field types not resolved, counts not checked in patterns (`checker.go:122-127`)
- [ ] **CHK-17** `ErrorType{}` defined but never used — error returns are untyped (`types.go:20`)
- [ ] **CHK-18** `PropagateExpr` (`?`) incomplete — no validation inner expression returns error, `OrelseExpr` operands not typed (`checker.go:876-884`)
- [ ] **CHK-19** Selective imports never validated — `import { NonExistent } from "m"` passes (`resolver.go:228-311`)
- [x] **CHK-20** Duplicate declaration detection — checker now warns on duplicate type/function/tool/workflow definitions in `registerGlobals` (`checker.go`)
- [ ] **CHK-21** Complex assignment paths not validated — `FieldPath` and `IndexPath` field/key types unchecked (`checker.go:541-567`)
- [ ] **CHK-22** Streaming workflows not type-checked — `-> stream` return type has no checker support (`checker.go`)
- [ ] **CHK-23** Agent field values not schema-validated — `provider`, `tools`, `handoffs`, `memory` accept arbitrary expressions (`checker.go:230-317`)
- [ ] **CHK-24** Workflow trigger decorators not validated — no check that trigger name is valid or params correct (`checker.go:472-503`)
- [ ] **CHK-25** Workflow hook type mismatches — `@onsuccess` binds `AnyType{}` instead of workflow return type; `@onerror` uses `StringType{}` instead of `ErrorType{}` (`checker.go:488-500`)
- [ ] **CHK-26** Function type (`FuncType`) not utilized — function references always return `AnyType{}` (`types.go:53-56, checker.go:718`)
- [ ] **CHK-27** Named argument validation missing — named args parsed but never matched against function params (`ast.go:461-465`)
- [ ] **CHK-28** Rest parameters not type-checked — arity checking absent (`checker.go`)

---

## 2. Compiler — Parser

- [x] **PAR-01** Left-side type annotation not parsed — `AssignTarget.Ty` field exists but `exprToAssignTarget()` never populates it; `result: Type, err = agent.run(req)` doesn't work (`parser.go:1186, ast.go:217-220`)
- [x] **PAR-02** `OptionType` syntax not recognized — no production rule for `T?` in `parseType()` (`parser.go:773-906`)
- [ ] **PAR-03** Pipe operator right-associative — uses `prec` instead of `prec+1` for left-associative ops; `a |> b |> c` parses as `a |> (b |> c)` (`parser.go:1786-1900`)
- [ ] **PAR-04** Range operators share comparison precedence — `1 < 5 .. 10` is ambiguous (`parser.go:65-66`)
- [ ] **PAR-05** Decorator struct lacks `Span` field — can't provide accurate diagnostics (`ast.go:100-103`)
- [x] **PAR-06** Rest parameters can appear anywhere — added enforcement that `...` params must be last (`parser.go:679-696`)
- [ ] **PAR-07** Error recovery skips only one token — cascading errors on complex constructs (`parser.go:208-210`)
- [x] **PAR-08** Potential infinite loop in field parsing — added token advance on parse failure to prevent stall (`parser.go:557-567`)

---

## 3. Compiler — Lexer

- [ ] **LEX-01** Standalone `!` not handled — becomes `token.Error`; language uses `not` but `!` in `!=` works (`lexer.go:368-523`)
- [ ] **LEX-02** `NextWithNewlines()` identical to `Next()` — documented to preserve newlines but skips them (`lexer.go:26-52`)
- [ ] **LEX-03** Empty numeric prefixes accepted — `0x`, `0b`, `0o` with no digits produce valid `token.Int` (`lexer.go:306-353`)
- [ ] **LEX-04** Unterminated block comments not reported — `/* unclosed` at EOF produces `token.BlockComment` instead of `token.Error` (`lexer.go:120-139`)
- [ ] **LEX-05** No scientific notation — `1e10`, `1.5e-2` lex as int+ident (`lexer.go:306-353`)
- [x] **LEX-06** Dead code `IsKeywordIdent()` removed — was never called (`lexer.go`)
- [ ] **LEX-07** 9 keywords missing from lexer tests (`lexer_test.go:42-57`)

---

## 4. Compiler — Codegen

- [x] **GEN-01** Backtick injection — JSON schemas/descriptions embedded in Go backtick strings; backtick in content breaks generated Go (`agentic.go:240, 344, 348`)
- [x] **GEN-02** `patternToCondition` now returns `"false /* unknown pattern */"` instead of `"true"` — prevents silent match-all (`statements.go:488`)
- [ ] **GEN-03** Error propagation recovery can panic — `fmt.Sprintf("%v", r)` on `recover()` value (`statements.go:503`)
- [ ] **GEN-04** Spawn blocks lose multi-error context — `panic(e)` on first error, others lost (`expressions.go:536-546`)
- [x] **GEN-05** `json.Unmarshal` error now checked in tool handlers — returns `"invalid tool arguments"` error (`agentic.go:309-311`)
- [x] **GEN-06** Tool default values now handle bool (`!params.X`) and float (`params.X == 0.0`) in addition to int/string (`agentic.go:313-328`)
- [x] **GEN-07** Workflow parameter type assertions already safe — uses comma-ok pattern `%s, _ := params[%q].(%s)` which never panics (`agentic.go:442`)
- [ ] **GEN-08** Match if-chain indentation bugs — inconsistent `OpenBlock`/manual `else if` handling (`statements.go:193-203`)
- [ ] **GEN-09** Missing type conversion cases — `stream`, `error`, user-defined struct types not in `HairaTypeToGo()` (`types.go:115-170`)
- [ ] **GEN-10** Global state not thread-safe — `activeTypeInfo`, `activeWorkflowName`, `inCapturedContext` etc. (`types.go, agentic.go, expressions.go`)
- [ ] **GEN-11** 11 global mutable variables track codegen state — `activeTypeInfo`, `activeSourceFile`, `activeWorkflowName`, `activeStepName`, `knownToolNames`, `activeTarget`, `propagateCounter`, `orelseCounter`, `stepCounter`, `inCapturedContext`, `inRetryContext`, `currentRetryVar` — should be refactored into a context struct (`types.go, agentic.go, expressions.go`)
- [ ] **GEN-12** `resolveQualified()` is 590+ lines with one giant switch — should be split into per-module resolvers (`stdlib.go:66`)
- [ ] **GEN-13** Hardcoded Cloudflare Workers version `v0.26.3` — should track latest or be configurable (`project.go:460`)
- [ ] **GEN-14** Wrangler.toml generation emits placeholder `<run: npx wrangler d1 create ...>` — incomplete output requires manual editing (`project.go:552-569`)

---

## 5. Compiler — Spec vs Implementation

- [ ] **SPEC-01** Channels (`chan`) — spec chapter 8 describes full semantics but no token, AST, parser, or codegen exists
- [x] **SPEC-02** `nil` vs `none`/`some` — spec and examples use `nil`, token system defines `None`/`Some`; incoherent
- [ ] **SPEC-03** Logical right shift (`>>>`) — in spec chapter 2 but not tokenized or parsed
- [x] **SPEC-04** `@websocket`, `@cron`, `@event`, `@manual` triggers — in spec but codegen only handles HTTP triggers
- [x] **SPEC-05** 16 keywords implemented but undocumented in spec — `step`, `onerror`, `onsuccess`, `oncancel`, `errdefer`, `test`, `assert`, `let`, `const`, `orelse`, `none`, `some`, `async`, `err`, `ok`, `default`
- [ ] **SPEC-06** Sized/unsigned integer types (`i8`, `i16`, `u8`, `u32`, etc.) — in spec chapter 3 but not in type system
- [ ] **SPEC-07** Error struct — spec defines `Error { message, code?, source? }` but implementation uses Go error interface

---

## 6. Runtime (primitive/haira/) — 35 files

- [x] **RT-01** `Response.Json()` ignores unmarshal errors — now panics on error (caught by try/catch) (`http.go:18-25`)
- [x] **RT-02** Tool param unmarshal error ignored — `json.Unmarshal(td.Parameters, &params)` (`agent.go:712`)
- [x] **RT-03** Summary memory not implemented — improved fallback warning with max_turns info, defaults to 50 turns (`agent.go:162`)
- [ ] **RT-04** Session cleanup goroutine never stops — no shutdown mechanism (`upload.go:55-79`)
- [x] **RT-05** Observe slices grow unbounded — capped at 10,000 entries with FIFO eviction (`observe.go`)
- [x] **RT-06** Sessions never evicted from in-memory map — evicts oldest when >1000 sessions (`memory.go`)
- [ ] **RT-07** `TimeTick()` channel never closed — receiver blocks forever (`time.go:104-114`)
- [ ] **RT-08** SSE response body leak on error path (`mcp_client.go:191-195`)
- [x] **RT-09** Panic recovery now includes stack traces via `debug.Stack()` (`agent.go`)
- [x] **RT-10** `StringTruncate` off-by-one fixed — handles edge cases for suffix >= maxLen (`strings_ext.go`)
- [x] **RT-11** Regex errors now panic (caught by Haira's try/catch) via `mustCompile()` helper (`regex.go`)
- [x] **RT-12** `ParseJSON` now panics on error (caught by try/catch) instead of returning empty map (`helpers.go`)
- [ ] **RT-13** Server JSON marshal errors ignored (`server.go:224`)
- [ ] **RT-14** HTTP response `io.ReadAll` error ignored (`http_client.go:124`)

### Runtime — Hardcoded Values (NEW)

- [ ] **RT-15** `knownBackends` map hardcodes 10 provider URLs — should be extensible or configurable (`agent.go:62-73`)
- [ ] **RT-16** Ollama default host hardcoded to `localhost:11434` — no env var override (`agent.go:112-116`)
- [ ] **RT-17** Default `maxSteps = 10` and timeout `120s` hardcoded — should be documented constants (`agent.go:169-170, 268, 543`)
- [ ] **RT-18** Scope denial message hardcoded: `"I can only help with topics within my scope."` (`agent.go:555`)
- [ ] **RT-19** `recentWindowSize = 16` hardcoded — affects message compaction behavior (`memory.go:112`)
- [ ] **RT-20** Memory `MaxCommands: 20` and `MaxFacts: 10` hardcoded (`memory.go:73, 82`)
- [ ] **RT-21** Multipart form limit `32 << 20` (32MB) hardcoded (`server.go:67`)
- [ ] **RT-22** Static file size limit `500KB` hardcoded (`server.go:602`)
- [ ] **RT-23** Hidden file/dir patterns hardcoded: `.`, `node_modules`, `__pycache__`, `.output` (`server.go:543`)
- [ ] **RT-24** HTTP client default timeout `30s` hardcoded (`http_client.go:28`)
- [ ] **RT-25** Upload temp dir hardcoded to `os.TempDir()/haira-uploads` — not configurable (`upload.go:14, 37`)
- [ ] **RT-26** Session cleanup intervals hardcoded: 1h check, 24h max age (`upload.go:55, 64`)
- [ ] **RT-27** SQLite store filename hardcoded to `.haira.db` (`store.go:102`)
- [ ] **RT-28** MCP protocol version hardcoded to `"2024-11-05"` (`mcp_client.go:332, mcp_server.go:59`)
- [ ] **RT-29** MCP client info hardcoded: `name: "haira", version: "1.0.0"` (`mcp_client.go:336`)
- [ ] **RT-30** `maxHandoffDepth = 5` hardcoded — not configurable per agent (`agent.go:59`)

### Runtime — Duplicate Code (NEW)

- [ ] **RT-31** HTTP request building duplicated between `http.go:87-125` and `http_client.go:84-144` — consolidate into shared utility
- [x] **RT-32** LLM generation recording duplicated 3x — extracted `recordGeneration()` helper method (`agent.go`)
- [-] **RT-33** String truncation pattern duplicated — different lengths per context, acceptable duplication
- [x] **RT-34** File extension → language detection extracted to `DetectLanguage()` (`strings_ext.go`, `server.go`)
- [ ] **RT-35** Time format strings inconsistent: `"2006-01-02T15:04:05"` vs `"2006-01-02T15:04:05Z07:00"` across `time.go`, `fs.go`

---

## 7. Stdlib — 16 packages

- [x] **STD-01** SQL via string concatenation in `ValidateWithDb()` — should use parameterized queries (`excel/excel_tables.go:662-672`)
- [x] **STD-02** Race condition — `localVectorDB` lazy singleton without `sync.Once` (`vector/vector_local.go:19-27`)
- [x] **STD-03** 12+ ignored `json.Unmarshal` errors across all 3 store backends (`store_postgres.go:306-309`, `store_sqlite.go:324-330`, `store_d1.go:325-330`)
- [ ] **STD-04** GitHub API hard-codes `"main"` as default branch (`github/github.go:83, 179, 334`)
- [ ] **STD-05** No connection pool settings exposed for Postgres (`postgres/postgres.go`)
- [ ] **STD-06** `Query()` loads all rows into memory — OOM on large results (`postgres/postgres.go`)
- [ ] **STD-07** Slack has no retry logic while GitHub/GitLab do (`slack/slack.go`)
- [ ] **STD-08** Excel validation doesn't distinguish NULL from empty string (`excel/excel_tables.go:464-470`)

### Stdlib — Code Quality (NEW)

- [x] **STD-09** UI helper functions extracted to `primitive/haira/ui_helpers.go` — shared by `algolia/` and `meilisearch/`
- [ ] **STD-10** `PostgresGenerateUpsert()` lives in `postgres/postgres_ext.go` but depends on `excelize` (Excel library) — circular coupling with `excel/` package. Move to `excel/` package.
- [ ] **STD-11** Store backends (`store_postgres.go`, `store_sqlite.go`, `store_d1.go`) share 80%+ code — consider abstract query builder or accept pragmatic duplication
- [ ] **STD-12** HTTP timeouts hardcoded: GitHub/GitLab `30000ms`, Algolia/Meilisearch `10000ms` — not configurable via constructor (`github.go:66, gitlab.go:60, algolia.go:29, meilisearch.go:28`)
- [ ] **STD-13** GitHub/GitLab polling loops hardcoded to 120 iterations with no exponential backoff (`github.go:199, gitlab.go:199, 223`)
- [ ] **STD-14** Langfuse flush interval `5s` and buffer size `50` hardcoded — no config parameters (`langfuse.go:78, 159`)
- [ ] **STD-15** Langfuse has no error recovery — dropped events on failed flush are lost silently (`langfuse.go`)
- [ ] **STD-16** `postgres.Transaction()` wrapper bug — wraps raw `*sql.DB` but doesn't apply to transaction (`postgres_ext.go:404`)
- [ ] **STD-17** Excel `SkipFirstRow()` heuristic checks for type keywords — could be fooled by actual data matching type names (`excel_tables.go:781-821`)
- [ ] **STD-18** Postgres schema fetched on every call — no caching (`postgres.go`)
- [ ] **STD-19** Vector filter parameter only works for pgvector, not chromem-go local backend — API inconsistency (`vector.go`)
- [ ] **STD-20** Langfuse exporter called with empty strings `("", "", "")` in all POCs — should fail-safe or read from env vars

---

## 8. LSP

- [x] **LSP-01** Race condition — `asts` and `typeInfos` updated in separate lock scopes (`handler.go:677-692`)
- [x] **LSP-02** Write errors ignored — `Writer.Write()` return values discarded (`server.go:223-224`)
- [ ] **LSP-03** No JSON-RPC error response for parse failures — client hangs (`server.go:105-108`)
- [x] **LSP-04** UTF-8 offset bug — `positionToOffset()` treats characters as bytes (`handler.go:742-753`)
- [x] **LSP-05** Missing: `textDocument/formatting` (formatter exists but not wired)
- [ ] **LSP-06** Missing: `textDocument/references`
- [ ] **LSP-07** Missing: `workspace/symbol`
- [ ] **LSP-08** Missing: incremental document sync (always full reload)
- [ ] **LSP-09** Missing: `textDocument/signatureHelp`
- [x] **LSP-10** Markdown escaping missing in hover — not applicable (type strings are inside code fences)

---

## 9. Driver / Formatter / Errors / CLI

### Driver
- [x] **DRV-01** `os.MkdirAll()` error ignored (`driver.go:44`)
- [x] **DRV-02** No build artifact cleanup on error — stale binaries remain (`driver.go:52`)
- [ ] **DRV-03** No inter-phase consistency validation (`driver.go`)

### Formatter (0 tests, 1,297 LOC)
- [x] **FMT-01** Comment cursor state not reset — false positive (Format() creates fresh struct each call)
- [ ] **FMT-02** Blank line rules only consider import vs non-import transitions (`items.go:21-23`)
- [ ] **FMT-03** Idempotency not guaranteed — format twice may differ

### Errors/Diagnostics (0 tests, 146 LOC)
- [ ] **ERR-01** Multi-line span underlines only first line (`errors.go:84-92`)
- [ ] **ERR-02** Hint overwrites message when both present (`errors.go:95-98`)
- [ ] **ERR-03** No error codes or categories

### CLI
- [x] **CLI-01** 4 unhandled JSON decode errors in orchestrator commands (`main.go:319, 346, 389, 503`)
- [ ] **CLI-02** All errors exit with code 1 — no distinction between error types (`main.go:86-88`)
- [ ] **CLI-03** Flag parser silently uses default on missing value (`main.go:249-258`)

---

## 10. UI SDK

- [x] **UI-01** XSS vector — replaced inline `onclick` with event delegation (`haira-message.ts`, `haira-markdown.ts`)
- [ ] **UI-02** `innerHTML` with template strings for icons (`haira-message.ts:469-473`)
- [ ] **UI-03** `{{META}}` placeholder requires backend escaping, no CSP (`loader.html:10`)
- [ ] **UI-04** Missing ARIA labels on buttons, avatars, toggles (`haira-form.ts`, `haira-table.ts`, `haira-message.ts`)
- [ ] **UI-05** No keyboard trap management in modals
- [ ] **UI-06** 8 highlight.js languages bundled statically — no lazy loading (`haira-message.ts`)
- [ ] **UI-07** SSE client only catches JSON errors, not network failures (`sse-client.ts`)

---

## 11. Tree-sitter Grammar & Editor Extensions

- [x] **TS-01** `select_block` referenced but rule not defined — false positive (rule exists at line 248)
- [x] **TS-02** `triple_quote_string` referenced but not defined — false positive (rule exists at line 463)
- [x] **TS-03** Pipe operator `|>` not in grammar rules — false positive (rule exists at line 354)
- [x] **TS-04** Error propagation `?` not in grammar rules — added `try_expression` rule
- [ ] **EXT-01** Zed extension uses stale commit hash for grammar (`extension.toml:11`)
- [ ] **EXT-02** No LSP path configured in Zed extension (`extension.toml`)
- [ ] **EXT-03** VS Code extension may lack LSP client implementation (`package.json`)

---

## 12. Test Coverage

### Missing test files (10 untested packages)
- [x] **TST-01** Add driver tests — entire pipeline untested (0/313 LOC)
- [x] **TST-02** Add formatter tests — 0/1,297 LOC
- [ ] **TST-03** Add errors/diagnostics tests — 0/146 LOC
- [ ] **TST-04** Add token utility tests — 0/428 LOC

### Missing test categories
- [x] **TST-05** Add parser negative tests — currently zero tests for invalid syntax
- [x] **TST-06** Add codegen compilation verification — generated Go verified via go/parser
- [ ] **TST-07** Add fuzz tests for lexer/parser
- [ ] **TST-08** Add benchmark tests for lexer/parser/checker on large files
- [ ] **TST-09** Add race detection to CI — `go test -race ./...`

### Build system
- [x] **TST-10** Add `make cover` target with coverage reporting
- [x] **TST-11** Add `make race` target with `-race` flag
- [x] **TST-12** Add `make bench` target for benchmarks

---

## 13. Examples Coverage Gaps

- [x] **EX-01** Add example for type aliases — `examples/30-type-aliases.haira`
- [x] **EX-02** Add example for spawn blocks — `examples/31-spawn.haira`
- [ ] **EX-03** Add example for select blocks
- [x] **EX-04** Add example for bitwise operators — `examples/32-bitwise.haira`
- [ ] **EX-05** Add example for named sessions (`session: "id"`)
- [ ] **EX-06** Add example for `@cron`/`@event`/`@manual` triggers
- [x] **EX-07** Add example for lifecycle hooks — `examples/33-lifecycle-hooks.haira`
- [ ] **EX-08** Add example for dynamic agent creation and parallel agent execution

---

## 14. Runtime — Dynamic Agent Creation (NEW)

Haira should support creating and launching agents at runtime, enabling an agent to spawn other agents dynamically for parallel task execution. This is critical for the "agentic orchestration" value prop.

### Current State
- `NewAgent()` in `agent.go` creates agents from `AgentConfig` — purely static, called at init time
- `spawn {}` block exists for parallel execution but only runs static expressions
- `handoffs` allow agent-to-agent delegation but are statically declared
- No runtime API for creating agents from within tool functions or agent logic

### Required Capabilities

- [x] **DYN-01** `create_agent()` runtime API — `CreateAgent(config, provider, tools)` creates agent from config map (`primitive/haira/agent.go`). Uses bare function call (not `agent.create()` — `agent` is a keyword).
- [x] **DYN-02** `spawn_agents()` — `SpawnAgents(tasks)` runs multiple agent calls in parallel, returns results (`primitive/haira/agent.go`). Uses bare function call.
- [ ] **DYN-03** Dynamic tool registry — allow agents to register/unregister tools at runtime
- [ ] **DYN-04** Agent pool pattern — pre-create N worker agents, dispatch tasks round-robin
- [ ] **DYN-05** Parent-child agent relationship — track which agent spawned which, propagate cancellation
- [ ] **DYN-06** Shared context/memory between dynamically spawned agents — e.g., shared conversation or facts
- [x] **DYN-07** Codegen support — `create_agent()` and `spawn_agents()` mapped as bare functions in `stdlib.go`. Also available via import alias: `import rt from "agent"` → `rt.create()`.
- [x] **DYN-08** Added `examples/34-dynamic-agents.haira` demonstrating runtime agent creation and parallel execution

### Design Sketch

```haira
// Static agent definition (existing)
agent Researcher {
    provider: Claude
    tools: [web_search]
    system: "You are a researcher"
}

// Dynamic agent creation (new)
fn research_parallel(topics: []string) -> []string {
    results = spawn {
        for topic in topics {
            // Create agent at runtime with dynamic config
            a = agent.create({
                name: "researcher-${topic}",
                provider: Claude,
                tools: [web_search],
                system: "Research this topic: ${topic}"
            })
            a.ask("Find the latest information")
        }
    }
    return results
}

// Or using a coordinator agent with a tool
tool spawn_researcher(topic: string) -> string {
    """Spawn a sub-agent to research a specific topic"""
    a = agent.create({
        name: "sub-researcher",
        provider: Claude,
        tools: [web_search],
        system: "You research: ${topic}"
    })
    return a.ask("Find key findings")
}
```

---

## 15. Stdlib Expansion — Extracted from POCs (NEW)

Analysis of 6 POCs identified reusable patterns that should become stdlib packages. Organized as `stdlib/tools/` (reusable tool implementations) and `stdlib/agents/` (reusable agent patterns).

### stdlib/tools/ — Reusable Tool Implementations

- [x] **STOOL-01** `stdlib/websearch/` — DuckDuckGo search API + web fetch with truncation. Created `stdlib/websearch/websearch.go`. Usage: `websearch.search(query)`, `websearch.fetch(url, max_len)`.
- [ ] **STOOL-02** `stdlib/tools/file_ops/` — Safe file read/write/edit with line numbers, find-and-replace, directory traversal with filtering. Extracted from `poc/coding-agent/` tools `read_file`, `write_file`, `edit_file`, `find_files`, `list_directory`, `search_files`.
- [x] **STOOL-03** `os.safe_exec(cmd, timeout)` — Safe command execution with blocklist (prevents `rm -rf`, `sudo`, `chmod 777`, `dd`, etc.). Added `OsSafeExec()` to `primitive/haira/os.go`.
- [ ] **STOOL-04** `stdlib/tools/test_runner/` — Auto-detect and run test frameworks (Go, npm, cargo, Python, make). Extracted from `poc/coding-agent/` `run_tests` (lines 424-476).
- [x] **STOOL-05** `stdlib/healthcheck/` — HTTP health check with timing + parallel batch checks. Created `stdlib/healthcheck/healthcheck.go`. Usage: `healthcheck.check(name, url)`, `healthcheck.check_all(services)`.
- [ ] **STOOL-06** `stdlib/tools/db_schema/` — Database schema introspection, read-only query enforcement, dry-run validator. Extracted from `poc/data-explorer/` (lines 114-154) and `poc/maltimize/` (lines 80-137).
- [ ] **STOOL-07** `stdlib/tools/diff/` — Before/after diff rendering with syntax highlighting and approval workflow. Extracted from `poc/coding-agent/` `show_diff` and `propose_edit`.
- [ ] **STOOL-08** `stdlib/tools/chart/` — Data visualization (bar, line, pie, area, scatter) via UI components. Extracted from `poc/data-explorer/` `visualize()` (lines 255-275).
- [ ] **STOOL-09** `stdlib/tools/runbook/` — Executable runbook pattern (restart, scale, rollback, toggle). Extracted from `poc/devops-incident/` (lines 263-305).
- [x] **STOOL-10** Notification extensions — `slack.send_alert(webhook, message, severity)` with severity icons (new `stdlib/slack/slack_alert.go`). GitHub `gh.create_incident(title, body, severity)` with P0/P1/P2 labels (added to `stdlib/github/github.go`).

### stdlib/agents/ — Reusable Agent Patterns

- [ ] **SAGENT-01** `stdlib/agents/coder/` — Coding assistant agent with file ops, search, command execution, approval workflow. Extracted from `poc/coding-agent/main.haira` (636 lines → reusable ~200 lines).
- [ ] **SAGENT-02** `stdlib/agents/data_explorer/` — Natural language data querying across Postgres, Algolia, Meilisearch with chart generation. Extracted from `poc/data-explorer/main.haira`.
- [ ] **SAGENT-03** `stdlib/agents/incident_commander/` — Multi-agent incident response: triage → diagnose → remediate → notify. Extracted from `poc/devops-incident/main.haira` handoff pattern.
- [ ] **SAGENT-04** `stdlib/agents/pipeline/` — Step-based pipeline execution with progress tracking, validation, and error handling. Extracted from `poc/pipeline-form/main.haira`.

### stdlib/ — Missing Infrastructure

- [ ] **SINF-01** `stdlib/config/` — Configuration management: `config.get(key, default)`, `config.get_int()`, `config.require()`. All POCs hardcode ports, endpoints, limits.
- [ ] **SINF-02** `stdlib/retry/` — Retry with exponential backoff, jitter, max attempts. Currently missing; GitHub/GitLab have ad-hoc retry but Slack has none.
- [ ] **SINF-03** `stdlib/approval/` — Generic approval workflow pattern (propose → render → wait → apply/reject). Used in coding-agent and maltimize but not reusable.

---

## 16. POC Cleanup (NEW)

Issues found in proof-of-concept code that should be fixed or documented.

- [x] **POC-01** `poc/coding-agent/` contains compiled `hcode` binary (16MB) — added to `.gitignore`
- [ ] **POC-02** `agent` binary (16MB) at repo root — should be in `.gitignore`
- [x] **POC-03** All POCs now use `env("PORT", int) orelse <default>` — no more hardcoded ports. Port conflicts resolved (cloudflare-agent→9006, pipeline-form→9007)
- [x] **POC-04** All POCs now use `langfuse.exporter(env("LANGFUSE_PUBLIC_KEY"), env("LANGFUSE_SECRET_KEY"), env("LANGFUSE_HOST"))` — reads from env vars
- [ ] **POC-05** `poc/maltimize/` has 47 hardcoded table→sheet mappings and conflict column configs — should be external JSON config
- [ ] **POC-06** `poc/maltimize/` has French-only UI/system strings — no i18n
- [ ] **POC-07** `poc/maltimize/` hardcodes SQL output paths to specific project structure (`apps/maite-api/src/infrastructure/...`)
- [x] **POC-08** `poc/coding-agent/.gitignore` now covers `hcode`, `.haira.db*` files
- [ ] **POC-09** No POC has test files (`*_test.haira`) — add basic smoke tests
- [x] **POC-10** SQLite database files (`.haira.db`, `.haira.db-shm`, `.haira.db-wal`) added to `poc/coding-agent/.gitignore`

---

## Recommended Order of Attack

### Phase 1: Core Language Correctness (highest impact) ✅ DONE
1. ~~CHK-01 through CHK-05 (type system foundations)~~
2. ~~PAR-01, PAR-02 (option type + left-side annotation parsing)~~
3. ~~CHK-08 (break/continue validation)~~
4. ~~CHK-14 (return type errors not warnings)~~
5. ~~GEN-01 (backtick injection)~~

### Phase 2: Safety & Robustness ✅ DONE
6. ~~LSP-01 (race condition)~~
7. ~~STD-02 (vector race condition)~~
8. ~~STD-01 (SQL injection)~~
9. ~~STD-03 (ignored unmarshal errors)~~
10. ~~RT-01 through RT-03 (runtime error handling)~~

### Phase 3: Checker Depth ✅ DONE
11. ~~CHK-06, CHK-07 (method calls, instance validation)~~
12. ~~CHK-10, CHK-11 (bitwise operator checking)~~
13. ~~CHK-12, CHK-13 (pattern matching validation)~~
14. ~~CHK-15, CHK-16 (default values, enum variants)~~

### Phase 4: Spec Alignment ✅ PARTIALLY DONE
15. ~~SPEC-02 (nil vs none)~~
16. ~~SPEC-04 (trigger codegen)~~
17. ~~SPEC-05 (document undocumented keywords)~~

### Phase 5: Test Infrastructure ✅ PARTIALLY DONE
18. ~~TST-05 (parser negative tests)~~
19. ~~TST-06 (codegen compilation verification)~~
20. ~~TST-01, TST-02 (driver, formatter tests)~~
21. ~~TST-10, TST-11, TST-12 (coverage, race, bench targets)~~

### Phase 6: Codebase Sanitization (NEW — do next)
22. RT-31 through RT-35 (duplicate code consolidation)
23. STD-09, STD-10 (stdlib duplicates and coupling)
24. GEN-11, GEN-12 (codegen global state refactor)
25. LEX-06 (dead code removal)
26. ~~POC-01, POC-08, POC-10 (gitignore cleanup)~~ ✅
27. ~~POC-03, POC-04 (hardcoded ports and empty langfuse calls)~~ ✅

### Phase 7: Runtime Hardening
28. RT-04 through RT-14 (memory leaks, error handling, resource cleanup)
29. RT-15 through RT-30 (hardcoded values → configurable constants)
30. STD-04 through STD-08, STD-12 through STD-20 (stdlib fixes)

### Phase 8: Dynamic Agent Creation (NEW — high value)
31. DYN-01, DYN-02 (runtime agent creation and parallel spawn)
32. DYN-03 (dynamic tool registry)
33. DYN-05, DYN-06 (parent-child agents, shared context)
34. DYN-07, DYN-08 (codegen support and examples)

### Phase 9: Stdlib Expansion (from POCs) — IN PROGRESS
35. ~~STOOL-01 (websearch), STOOL-03 (safe exec), STOOL-05 (healthcheck), STOOL-10 (alerts)~~ ✅
36. STOOL-02 (file ops), STOOL-06 (db schema — used in multiple POCs)
37. SINF-01 through SINF-03 (config, retry, approval infrastructure)
38. SAGENT-01 (coder agent — strongest demo piece)
39. Remaining STOOL and SAGENT items

### Phase 10: Remaining Checker & Codegen
40. CHK-09, CHK-17 through CHK-28 (remaining checker gaps)
41. GEN-02 through GEN-10, GEN-13, GEN-14 (remaining codegen issues)
42. PAR-03 through PAR-08 (remaining parser issues)

### Phase 11: LSP & Tooling
43. LSP-03 through LSP-09 (LSP improvements)
44. FMT-02, FMT-03 (formatter fixes)
45. EXT-01 through EXT-03 (editor extensions)

### Phase 12: Polish
46. UI-02 through UI-07
47. EX-03 through EX-08
48. CLI-02, CLI-03
49. ERR-01 through ERR-03
50. LEX-01 through LEX-05, LEX-07

---

## Progress Summary

| Section | Total | Done | Pending | % |
|---------|-------|------|---------|---|
| 1. Checker | 28 | 17 | 11 | 61% |
| 2. Parser | 8 | 4 | 4 | 50% |
| 3. Lexer | 7 | 1 | 6 | 14% |
| 4. Codegen | 14 | 5 | 9 | 36% |
| 5. Spec | 7 | 3 | 4 | 43% |
| 6. Runtime | 35 | 11 | 24 | 31% |
| 7. Stdlib | 20 | 4 | 16 | 20% |
| 8. LSP | 10 | 5 | 5 | 50% |
| 9. Driver/Fmt/Err/CLI | 12 | 5 | 7 | 42% |
| 10. UI SDK | 7 | 1 | 6 | 14% |
| 11. Tree-sitter/Ext | 7 | 4 | 3 | 57% |
| 12. Tests | 12 | 8 | 4 | 67% |
| 13. Examples | 8 | 4 | 4 | 50% |
| 14. Dynamic Agents | 8 | 4 | 4 | 50% |
| 15. Stdlib Expansion | 17 | 4 | 13 | 24% |
| 16. POC Cleanup | 10 | 5 | 5 | 50% |
| **TOTAL** | **213** | **85** | **128** | **40%** |
