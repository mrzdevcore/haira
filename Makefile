.PHONY: build release test clean check fmt lint doc install run

# Default target
all: build

# Build debug version
build:
	cargo build

# Build release version
release:
	cargo build --release

# Run tests
test:
	cargo test

# Clean build artifacts
clean:
	cargo clean
	@# Remove compiled Haira output directory
	rm -rf .output
	@# Remove object files
	rm -f *.o
	@echo "Cleaned all build artifacts"

# Check code without building
check:
	cargo check

# Format code
fmt:
	cargo fmt

# Check formatting
fmt-check:
	cargo fmt --check

# Run clippy linter
lint:
	cargo clippy -- -D warnings

# Generate documentation
doc:
	cargo doc --no-deps --open

# Install haira binary
install:
	cargo install --path crates/haira-cli

# Uninstall haira binary
uninstall:
	cargo uninstall haira

# Run haira with arguments (use: make run ARGS="parse examples/hello.haira")
run:
	cargo run -- $(ARGS)

# Quick development cycle: format, check, test
dev: fmt check test

# CI pipeline: format check, lint, test
ci: fmt-check lint test

# Parse an example file
parse-hello:
	cargo run -- parse examples/hello.haira

# Lex an example file
lex-hello:
	cargo run -- lex examples/hello.haira

# Show haira info
info:
	cargo run -- info

# Test AI interpretation
interpret:
	cargo run -- interpret get_active_users

# Build all examples (non-AI)
build-examples: build
	@echo "Building all examples..."
	@failed=0; \
	for f in $$(find examples -name "*.haira" ! -path "examples/ai/*" ! -path "examples/testing/*"); do \
		echo "Building $$f..."; \
		./target/debug/haira build "$$f" 2>&1 || { echo "FAILED: $$f"; failed=$$((failed + 1)); }; \
	done; \
	echo ""; \
	if [ $$failed -gt 0 ]; then \
		echo "$$failed example(s) failed to build"; \
		exit 1; \
	else \
		echo "All examples built successfully!"; \
	fi

# Build AI examples (requires --ollama flag)
build-ai-examples: build
	@echo "Building AI examples with Ollama..."
	@failed=0; \
	for f in examples/ai/*.haira; do \
		echo "Building $$f..."; \
		./target/debug/haira build --ollama "$$f" 2>&1 || { echo "FAILED: $$f"; failed=$$((failed + 1)); }; \
	done; \
	echo ""; \
	if [ $$failed -gt 0 ]; then \
		echo "$$failed AI example(s) failed to build"; \
		exit 1; \
	else \
		echo "All AI examples built successfully!"; \
	fi

# Run test examples
test-examples: build
	@echo "Running test examples..."
	@failed=0; \
	for f in examples/testing/*.haira; do \
		echo "Testing $$f..."; \
		./target/debug/haira build "$$f" 2>&1 || { echo "FAILED: $$f"; failed=$$((failed + 1)); }; \
	done; \
	echo ""; \
	if [ $$failed -gt 0 ]; then \
		echo "$$failed test(s) failed"; \
		exit 1; \
	else \
		echo "All tests passed!"; \
	fi

# Build and run all non-AI examples
run-examples: build-examples
	@echo "Running all built examples..."
	@for f in $$(find examples -name "*.haira" ! -path "examples/ai/*" ! -path "examples/testing/*"); do \
		name=$$(basename "$$f" .haira); \
		if [ -f ".output/$$name" ]; then \
			echo "Running $$name..."; \
			".output/$$name" 2>&1 || true; \
			echo ""; \
		fi; \
	done

# Build all examples (including AI with Ollama)
build-all-examples: build-examples build-ai-examples test-examples
	@echo "All examples built!"

# Help
help:
	@echo "Haira Makefile targets:"
	@echo "  build        - Build debug version"
	@echo "  release      - Build release version"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  check        - Check code without building"
	@echo "  fmt          - Format code"
	@echo "  fmt-check    - Check formatting"
	@echo "  lint         - Run clippy linter"
	@echo "  doc          - Generate and open documentation"
	@echo "  install      - Install haira binary"
	@echo "  uninstall    - Uninstall haira binary"
	@echo "  run          - Run haira (use ARGS=\"...\")"
	@echo "  dev          - Format, check, test"
	@echo "  ci           - CI pipeline"
	@echo "  parse-hello  - Parse examples/hello.haira"
	@echo "  lex-hello    - Lex examples/hello.haira"
	@echo "  info         - Show haira info"
	@echo "  interpret    - Test AI interpretation"
	@echo ""
	@echo "Examples:"
	@echo "  build-examples     - Build all non-AI examples"
	@echo "  build-ai-examples  - Build AI examples (requires Ollama)"
	@echo "  test-examples      - Run test examples"
	@echo "  run-examples       - Build and run all non-AI examples"
	@echo "  build-all-examples - Build all examples (including AI)"
