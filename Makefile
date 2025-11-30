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
