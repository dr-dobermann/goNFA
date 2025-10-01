# Makefile for goNFA project
#
# goNFA is a universal, lightweight and idiomatic Go library for creating
# and managing non-deterministic finite automata (NFA). It provides reliable
# state management mechanisms for complex systems such as business process
# engines (BPM).
#
# Project: https://github.com/dr-dobermann/gonfa
# Author: dr-dobermann (rgabtiov@gmail.com)
# License: LGPL-2.1 (see LICENSE file in the project root)

.PHONY: all build test clean mocks lint fmt vet examples help tag

# Variables
BINARY_DIR := bin
EXAMPLES_DIR := examples
MOCKS_DIR := generated
GO_FILES := $(shell find . -name "*.go" -not -path "./$(MOCKS_DIR)/*" -not -path "./vendor/*")
# Version number
VERSION = $(shell cat .version)

# Default target
all: clean mocks test build examples

# Help target
help:
	@echo "Available targets:"
	@echo "  all       - Clean, generate mocks, test, build, and build examples"
	@echo "  build     - Build all binaries"
	@echo "  test      - Run tests with coverage"
	@echo "  mocks     - Generate mocks using mockery"
	@echo "  clean     - Clean build artifacts and mocks"
	@echo "  lint      - Run golangci-lint"
	@echo "  fmt       - Format code"
	@echo "  vet       - Run go vet"
	@echo "  examples  - Build example binaries"
	@echo "  install   - Install required tools"
	@echo "  tag       - Update project tag from .version"

# Install required tools
install:
	@echo "Installing required tools..."
	go install github.com/vektra/mockery/v3@v3.5.5
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Generate mocks
mocks:
	@echo "Generating mocks..."
	@mkdir -p $(MOCKS_DIR)
	mockery

# Format code
fmt:
	@echo "Formatting code..."
	gofmt -s -w $(GO_FILES)

# Vet code
vet:
	@echo "Running go vet..."
	go vet ./...

# Lint code
lint: fmt vet
	@echo "Running golangci-lint..."
	golangci-lint run

# Run tests
test: mocks
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./pkg/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report saved to coverage.html"

# Build main library (no main package, so just verify compilation)
build: mocks
	@echo "Building library..."
	go build ./pkg/...

# Build examples
examples: mocks
	@echo "Building examples..."
	@mkdir -p $(BINARY_DIR)
	@for example in $(shell find $(EXAMPLES_DIR) -name "*.go" -not -name "*_test.go"); do \
		name=$$(basename $$example .go); \
		echo "Building example: $$name"; \
		go build -o $(BINARY_DIR)/$$name $$example; \
	done

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf $(BINARY_DIR)
	rm -rf $(MOCKS_DIR)
	rm -f coverage.out coverage.html
	go clean ./...

# Run a specific example
run-example-%: examples
	@echo "Running example: $*"
	@if [ -f "$(BINARY_DIR)/$*" ]; then \
		./$(BINARY_DIR)/$*; \
	else \
		echo "Example $* not found. Available examples:"; \
		ls -1 $(BINARY_DIR)/; \
	fi

# Quick development cycle
dev: clean mocks test

# CI/CD target
ci: install lint test build examples

# Show coverage
coverage: test
	go tool cover -func=coverage.out | tail -1

# Set project tag from .version
tag: 
	@git tag -a ${VERSION} -m "version ${VERSION}"
	@git push origin --tags

