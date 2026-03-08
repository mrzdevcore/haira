# Changelog

All notable changes to Haira are documented in this file.

## [v0.4.0] - 2026-03-08

### Agentic Framework

- **Eval framework** — New `eval` top-level declaration and `haira eval` CLI command for automated agent evaluation with test cases, scoring, and pass/fail thresholds
- **Cross-harness export** — `haira build --target claude-code` generates `.claude/agents/*.md`, `.claude/commands/*.md`, `mcp-servers.json`, and an MCP server binary for custom tools
- **Tool lifecycle hooks** — `@before { }` and `@after { }` blocks inside tool declarations for pre/post-processing logic
- **Verification loops** — `verify { assert ... }` blocks inside workflow steps that integrate with `@retry` for automatic assertion-driven retries
- **Agent delegation strategy** — `strategy: "parallel"` (fan-out) and `strategy: "sequential"` (chain) fields on agents with handoffs
- **Pre-built agent templates** — `stdlib/agents` package with 8 reusable templates: CodeReviewer, Planner, SecurityReviewer, Summarizer, DataAnalyst, CustomerSupport, TDDGuide, DocWriter
- **Dynamic agents** — `create_agent()` runtime support for programmatically creating agents from config maps
- **Auth module** — `stdlib/auth` for API key resolution (env var, credentials file, Claude Code OAuth)

### Compiler & Language

- **Stdlib expansion** — OS primitives, web search, healthcheck modules, and new UI components
- **Type alias codegen** — `type Name = Type` now correctly emits Go type aliases with typed list coercion in struct literals
- **Checker improvements** — Deeper type checking, spec alignment, enum exhaustiveness, and `websearch`/`healthcheck` module recognition
- **Return value enforcement** — Compiler enforces return values in functions
- **LSP improvements** — Better IDE support, formatting integration, and tree-sitter grammar fixes
- **Formatter** — `haira fmt` command for in-place source formatting
- **Tree-sitter** — Grammar fixes for editor syntax highlighting

### Runtime & Infrastructure

- **Built-in providers** — Default provider configurations
- **UI theming** — Theme support with embedded UI bundle
- **Workflow enhancements** — Improved workflow execution, form display, and table rendering
- **ARP/SSE fixes** — Proper `event:` headers for delta and done events in Server-Sent Events
- **Spawn capture** — `spawn { }` blocks correctly capture concurrent results

### Documentation & Tooling

- **Website** — Landing page, multilang docs, SEO, and analytics
- **Spec updates** — Language specification aligned with current compiler state
- **Due diligence** — Comprehensive code quality review (phases 1-5)
- **New examples** — 35 example programs covering all language features
- **POCs** — Coding agent, Cloudflare agent, data explorer, DevOps incident, and pipeline form proof-of-concepts

### Bug Fixes

- Fixed driver error handling and CLI decode errors
- Fixed XSS removal in UI rendering
- Fixed `fmt` import detection for tool declarations
- Fixed cropped text in UI components
- Fixed umami script type annotation

## [v0.3.0] - 2026-02-23

Initial public release with core language, agentic primitives, and Cloudflare Workers target.
