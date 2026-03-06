# Haira Due Diligence — Issue Tracker

**Generated:** 2026-03-06
**Total findings:** 142 (40 critical, 41 high, 45 medium, 16 low)

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
- [ ] **CHK-20** No duplicate declaration detection — `DefineVar`/`DefineType`/`DefineFunc` silently overwrite (`env.go:38-88`)
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
- [ ] **PAR-06** Rest parameters can appear anywhere — no enforcement that `...` params are last (`parser.go:697-743`)
- [ ] **PAR-07** Error recovery skips only one token — cascading errors on complex constructs (`parser.go:208-210`)
- [ ] **PAR-08** Potential infinite loop in field parsing — if `parseField()` fails and no comma, loop doesn't advance (`parser.go:557-566`)

---

## 3. Compiler — Lexer

- [ ] **LEX-01** Standalone `!` not handled — becomes `token.Error`; language uses `not` but `!` in `!=` works (`lexer.go:368-523`)
- [ ] **LEX-02** `NextWithNewlines()` identical to `Next()` — documented to preserve newlines but skips them (`lexer.go:26-52`)
- [ ] **LEX-03** Empty numeric prefixes accepted — `0x`, `0b`, `0o` with no digits produce valid `token.Int` (`lexer.go:306-353`)
- [ ] **LEX-04** Unterminated block comments not reported — `/* unclosed` at EOF produces `token.BlockComment` instead of `token.Error` (`lexer.go:120-139`)
- [ ] **LEX-05** No scientific notation — `1e10`, `1.5e-2` lex as int+ident (`lexer.go:306-353`)
- [ ] **LEX-06** Dead code `IsKeywordIdent()` — never called, contradicts ASCII-only identifier scanning (`lexer.go:541-544`)
- [ ] **LEX-07** 9 keywords missing from lexer tests (`lexer_test.go:42-57`)

---

## 4. Compiler — Codegen

- [x] **GEN-01** Backtick injection — JSON schemas/descriptions embedded in Go backtick strings; backtick in content breaks generated Go (`agentic.go:240, 344, 348`)
- [ ] **GEN-02** `patternToCondition` returns `"true"` for unknown patterns — silently matches everything (`statements.go:488`)
- [ ] **GEN-03** Error propagation recovery can panic — `fmt.Sprintf("%v", r)` on `recover()` value (`statements.go:503`)
- [ ] **GEN-04** Spawn blocks lose multi-error context — `panic(e)` on first error, others lost (`expressions.go:536-546`)
- [ ] **GEN-05** `json.Unmarshal` error ignored in tool handlers — malformed calls proceed with zero params (`agentic.go:307`)
- [ ] **GEN-06** Tool default values only handle int/string — bool, float, array defaults broken (`agentic.go:310-324`)
- [ ] **GEN-07** Workflow parameter type assertions panic — `params[key].(type)` on mismatch (`agentic.go:439-448`)
- [ ] **GEN-08** Match if-chain indentation bugs — inconsistent `OpenBlock`/manual `else if` handling (`statements.go:193-203`)
- [ ] **GEN-09** Missing type conversion cases — `stream`, `error`, user-defined struct types not in `HairaTypeToGo()` (`types.go:115-170`)
- [ ] **GEN-10** Global state not thread-safe — `activeTypeInfo`, `activeWorkflowName`, `inCapturedContext` etc. (`types.go, agentic.go, expressions.go`)

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
- [ ] **RT-05** Observe slices grow unbounded — `generations`/`toolExecs` accumulate forever (`observe.go:100-101`)
- [ ] **RT-06** Sessions never evicted from in-memory map (`memory.go:65-77`)
- [ ] **RT-07** `TimeTick()` channel never closed — receiver blocks forever (`time.go:104-114`)
- [ ] **RT-08** SSE response body leak on error path (`mcp_client.go:191-195`)
- [ ] **RT-09** Panic recovery loses stack traces — `fmt.Errorf("%v", r)` (`agent.go:667-671`)
- [ ] **RT-10** `StringTruncate` off-by-one — suffix can exceed maxLen (`strings_ext.go:74-82`)
- [ ] **RT-11** Regex errors silently return empty string (`regex.go:15-21`)
- [ ] **RT-12** `ParseJSON` returns empty map on error (`helpers.go:12`)
- [ ] **RT-13** Server JSON marshal errors ignored (`server.go:224`)
- [ ] **RT-14** HTTP response `io.ReadAll` error ignored (`http_client.go:124`)

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

---

## Recommended Order of Attack

### Phase 1: Core Language Correctness (highest impact)
1. CHK-01 through CHK-05 (type system foundations)
2. PAR-01, PAR-02 (option type + left-side annotation parsing)
3. CHK-08 (break/continue validation)
4. CHK-14 (return type errors not warnings)
5. GEN-01 (backtick injection)

### Phase 2: Safety & Robustness
6. LSP-01 (race condition)
7. STD-02 (vector race condition)
8. STD-01 (SQL injection)
9. STD-03 (ignored unmarshal errors)
10. RT-01 through RT-03 (runtime error handling)

### Phase 3: Checker Depth
11. CHK-06, CHK-07 (method calls, instance validation)
12. CHK-10, CHK-11 (bitwise operator checking)
13. CHK-12, CHK-13 (pattern matching validation)
14. CHK-15, CHK-16 (default values, enum variants)

### Phase 4: Spec Alignment
15. SPEC-02 (nil vs none)
16. SPEC-04 (trigger codegen)
17. SPEC-05 (document undocumented keywords)

### Phase 5: Test Infrastructure
18. TST-05 (parser negative tests)
19. TST-06 (codegen compilation verification)
20. TST-01, TST-02 (driver, formatter tests)
21. TST-10, TST-11 (coverage, race targets)

### Phase 6: LSP & Tooling
22. LSP-02 through LSP-09 (LSP improvements)
23. TS-01 through TS-04 (tree-sitter fixes)
24. FMT-01 through FMT-03 (formatter fixes)

### Phase 7: Polish
25. UI-01 through UI-07
26. EX-01 through EX-07
27. CLI-01 through CLI-03
28. DRV-01 through DRV-03
