---
title: "Installation"
description: "Install Haira and set up your development environment."
---

# Installation

## Quick Install

The fastest way to install Haira on macOS or Linux:

```bash
curl -sSL https://haira.dev/install.sh | bash
```

This downloads the latest release, verifies the checksum, and installs the `haira` binary to `~/.local/bin`. The runtime is embedded in the compiler binary — no separate runtime installation needed.

## Prerequisites

- **Go 1.22+** — Haira compiles to Go, so you need Go installed
- **macOS or Linux** — Windows support is experimental

## Verify Installation

```bash
haira --version
```

## Manual Install

Download pre-built binaries from [GitHub Releases](https://github.com/mrzdevcore/haira/releases).

Available platforms:
- Linux x86_64 / ARM64
- macOS x86_64 / ARM64 (Apple Silicon)
- Windows x86_64

## Build from Source

```bash
git clone https://github.com/mrzdevcore/haira.git
cd haira/compiler
make build
```

The binary will be at `compiler/bin/haira`.

## Editor Support

### Zed

The Haira extension for Zed provides syntax highlighting and language support:

```
Extensions → Search "haira" → Install
```

### Other Editors

Tree-sitter grammar is available at [`tree-sitter-haira/`](https://github.com/mrzdevcore/haira/tree/main/tree-sitter-haira) for building editor plugins.

## Next Steps

Once installed, continue to [Hello World](/docs/getting-started/hello-world) to write your first Haira program.
