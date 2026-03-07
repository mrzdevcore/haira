# Haira — Claude Code Instructions

## Project Context

Primary languages: Go (main), TypeScript (frontend/tooling), Rust (occasional). When writing Go code, always verify API compatibility of external packages before using methods that may not exist in the version being used.

## Debugging

When fixing bugs, verify the root cause before applying a fix. Do not assume naming mismatches or surface-level issues — check the actual database schema, config, or runtime behavior first.

## UI/Frontend

For UI implementations, get the layout and positioning right on the first pass. Show clean, user-friendly labels instead of raw IDs. When implementing scrollable lists or virtual scroll, prefer div-based approaches over table row hacks.

## Testing

When writing tests, use page object patterns and existing test abstractions rather than raw selectors. Check existing test files for conventions before writing new tests.

## Writing & Communication

When the user asks for content drafting (emails, posts, replies), keep responses concise and incorporate ALL feedback in each iteration. Do not re-add details the user explicitly asked to remove.
