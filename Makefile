.PHONY: build test clean install install-local install-system uninstall run fmt check-examples build-examples run-examples tree-sitter-generate tree-sitter-wasm zed-sync-wasm vscode-build vscode-package ui ui-dev bundle-runtime spec help

COMPILER_DIR = compiler
BINARY = $(COMPILER_DIR)/haira
VERSION ?= dev
LDFLAGS = -ldflags "-X main.version=$(VERSION)"

# Runtime directories
PRIMITIVE_DIR = primitive
STDLIB_DIR = stdlib
UI_SDK_SRC = ui/sdk/src
UI_SDK_DIST = ui/sdk/dist
UI_APP_DIR = ui/application
BUNDLE = $(COMPILER_DIR)/internal/runtime/bundle.tar.gz

# Default target
all: build

# Build the UI SDK bundle (TypeScript → JS via Bun)
ui:
	@mkdir -p $(UI_SDK_DIST)
	cd ui/sdk && bun build src/index.ts --outfile dist/haira-ui.js --minify --target browser
	@echo "UI bundle built: $(UI_SDK_DIST)/haira-ui.js"

# Build UI in watch mode (development)
ui-dev:
	@mkdir -p $(UI_SDK_DIST)
	cd ui/sdk && bun build src/index.ts --outfile dist/haira-ui.js --target browser --watch

# Bundle the runtime into a tar.gz for embedding in the compiler
# Merges primitive/ + stdlib/ into a single haira/ package
bundle-runtime: ui
	@echo "Bundling runtime..."
	@rm -rf .bundle-tmp
	@mkdir -p .bundle-tmp/haira/ui/dist
	@cp $(PRIMITIVE_DIR)/haira/*.go .bundle-tmp/haira/
	@cp $(STDLIB_DIR)/haira/*.go .bundle-tmp/haira/
	@cp $(UI_SDK_DIST)/haira-ui.js .bundle-tmp/haira/ui/dist/
	@cp $(UI_APP_DIR)/*.html .bundle-tmp/haira/ui/
	@cp $(PRIMITIVE_DIR)/go.mod .bundle-tmp/go.mod
	@cp $(PRIMITIVE_DIR)/go.sum .bundle-tmp/go.sum
	@tar czf $(BUNDLE) -C .bundle-tmp haira go.mod go.sum
	@rm -rf .bundle-tmp
	@echo "Runtime bundle: $(BUNDLE) ($$(du -h $(BUNDLE) | cut -f1))"

# Build the compiler (depends on runtime bundle)
build: bundle-runtime
	cd $(COMPILER_DIR) && go build $(LDFLAGS) -o haira .

# Run Go tests
test:
	cd $(COMPILER_DIR) && go test ./...

# Clean build artifacts
clean:
	rm -f $(BINARY)
	rm -f $(BUNDLE)
	rm -rf .output
	rm -rf .bundle-tmp
	rm -rf $(UI_SDK_DIST)
	@echo "Cleaned all build artifacts"

# Format code
fmt:
	cd $(COMPILER_DIR) && go fmt ./...

# Vet code
vet:
	cd $(COMPILER_DIR) && go vet ./...

# Install haira binary to $GOPATH/bin
install: build
	cp $(BINARY) $(shell go env GOPATH)/bin/haira
	@echo "Installed haira to $$(go env GOPATH)/bin/haira"

# Install haira binary to ~/.local/bin
install-local: build
	@mkdir -p ~/.local/bin
	@cp $(BINARY) ~/.local/bin/haira
	@echo "Installed haira to ~/.local/bin/haira"
	@echo ""
	@echo "Make sure ~/.local/bin is in your PATH:"
	@echo '  export PATH="$$PATH:$$HOME/.local/bin"'

# Install haira system-wide (requires sudo)
install-system: build
	@sudo cp $(BINARY) /usr/local/bin/haira
	@echo "Installed haira to /usr/local/bin/haira"

# Uninstall haira binary
uninstall:
	rm -f $(shell go env GOPATH)/bin/haira ~/.local/bin/haira
	@sudo rm -f /usr/local/bin/haira 2>/dev/null || true

# Run haira with arguments (use: make run ARGS="build examples/01-hello.haira")
run: build
	./$(BINARY) $(ARGS)

# Parse an example file
parse:
	./$(BINARY) parse $(FILE)

# Lex an example file
lex:
	./$(BINARY) lex $(FILE)

# Build all examples
build-examples: build
	@echo "Building all examples..."
	@failed=0; \
	for f in examples/*.haira; do \
		echo "  Building $$f..."; \
		./$(BINARY) build "$$f" 2>&1 || { echo "  FAILED: $$f"; failed=$$((failed + 1)); }; \
	done; \
	echo ""; \
	if [ $$failed -gt 0 ]; then \
		echo "$$failed example(s) failed to build"; \
		exit 1; \
	else \
		echo "All examples built successfully!"; \
	fi

# Run all non-agentic examples (07, 14, 15 need API keys)
run-examples: build
	@echo "Running non-agentic examples..."
	@failed=0; \
	for f in examples/*.haira; do \
		name=$$(basename "$$f" .haira); \
		case "$$name" in \
			07-*|14-*|15-*) echo "  Skipping $$f (needs API keys)"; continue ;; \
		esac; \
		echo "  Running $$f..."; \
		./$(BINARY) run "$$f" 2>&1 || { echo "  FAILED: $$f"; failed=$$((failed + 1)); }; \
		echo ""; \
	done; \
	if [ $$failed -gt 0 ]; then \
		echo "$$failed example(s) failed to run"; \
		exit 1; \
	else \
		echo "All examples ran successfully!"; \
	fi

# Regenerate tree-sitter grammar
tree-sitter-generate:
	cd tree-sitter-haira && tree-sitter generate

# Build tree-sitter WASM
tree-sitter-wasm: tree-sitter-generate
	cd tree-sitter-haira && tree-sitter build --wasm

# Sync tree-sitter WASM to Zed extension
zed-sync-wasm: tree-sitter-wasm
	cp tree-sitter-haira/tree-sitter-haira.wasm editors/zed-haira/grammars/haira.wasm
	@echo "Synced tree-sitter WASM → editors/zed-haira/grammars/haira.wasm"

# Build VS Code extension
vscode-build:
	cd editors/vscode-haira && bun install && bun run compile
	@echo "VS Code extension built: editors/vscode-haira/out/"

# Package VS Code extension (.vsix)
vscode-package: vscode-build
	cd editors/vscode-haira && bunx @vscode/vsce package
	@echo "VS Code extension packaged"

# Build the language specification PDF
spec:
	$(MAKE) -C spec/latex

# Quick development cycle: format, vet, test
dev: fmt vet test

# CI pipeline: vet, test, build all examples
ci: vet test build-examples

# Help
help:
	@echo "Haira Makefile targets:"
	@echo ""
	@echo "  build            Build the compiler (primitive + stdlib + UI embedded)"
	@echo "  test             Run Go tests"
	@echo "  clean            Clean build artifacts"
	@echo "  fmt              Format code"
	@echo "  vet              Vet code"
	@echo "  install          Install to GOPATH/bin"
	@echo "  install-local    Install to ~/.local/bin"
	@echo "  install-system   Install to /usr/local/bin (sudo)"
	@echo "  uninstall        Remove installed binary"
	@echo "  run              Run haira (use ARGS=\"...\")"
	@echo "  parse            Parse a file (use FILE=\"...\")"
	@echo "  lex              Lex a file (use FILE=\"...\")"
	@echo "  build-examples   Build all examples"
	@echo "  run-examples     Run non-agentic examples"
	@echo "  tree-sitter-generate  Regenerate tree-sitter grammar"
	@echo "  tree-sitter-wasm      Build tree-sitter WASM binary"
	@echo "  zed-sync-wasm         Sync WASM to Zed extension"
	@echo "  vscode-build          Build VS Code extension"
	@echo "  vscode-package        Package VS Code extension (.vsix)"
	@echo "  ui               Build UI SDK bundle (TypeScript → JS)"
	@echo "  ui-dev           Build UI SDK in watch mode"
	@echo "  bundle-runtime   Bundle primitive + stdlib into tar.gz for embedding"
	@echo "  spec             Build the language specification PDF"
	@echo "  dev              Format, vet, test"
	@echo "  ci               Vet, test, build examples"
