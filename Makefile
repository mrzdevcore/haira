.PHONY: build test clean install install-local install-system uninstall run fmt check-examples build-examples run-examples tree-sitter-generate help

COMPILER_DIR = compiler
BINARY = $(COMPILER_DIR)/haira

# Default target
all: build

# Build the compiler
build:
	cd $(COMPILER_DIR) && go build -o haira .

# Run Go tests
test:
	cd $(COMPILER_DIR) && go test ./...

# Clean build artifacts
clean:
	rm -f $(BINARY)
	rm -rf .output
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

# Quick development cycle: format, vet, test
dev: fmt vet test

# CI pipeline: vet, test, build all examples
ci: vet test build-examples

# Help
help:
	@echo "Haira Makefile targets:"
	@echo ""
	@echo "  build            Build the compiler"
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
	@echo "  dev              Format, vet, test"
	@echo "  ci               Vet, test, build examples"
