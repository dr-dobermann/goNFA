# goNFA - Nondeterministic Finite Automaton Go Library

[![Go Version](https://img.shields.io/github/go-mod/go-version/dr-dobermann/gonfa)](https://golang.org/)
[![GitHub release](https://img.shields.io/github/v/release/dr-dobermann/gonfa)](https://github.com/dr-dobermann/gonfa/releases)
[![License](https://img.shields.io/badge/License-LGPL%202.1-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/dr-dobermann/gonfa)](https://goreportcard.com/report/github.com/dr-dobermann/gonfa)
[![codecov](https://codecov.io/gh/dr-dobermann/gonfa/branch/master/graph/badge.svg)](https://codecov.io/gh/dr-dobermann/gonfa)
[![CI/CD Pipeline](https://github.com/dr-dobermann/gonfa/workflows/CI/CD%20Pipeline/badge.svg)](https://github.com/dr-dobermann/gonfa/actions)
[![GoDoc](https://godoc.org/github.com/dr-dobermann/gonfa?status.svg)](https://godoc.org/github.com/dr-dobermann/gonfa)
[![GitHub issues](https://img.shields.io/github/issues/dr-dobermann/gonfa)](https://github.com/dr-dobermann/gonfa/issues)
[![GitHub stars](https://img.shields.io/github/stars/dr-dobermann/gonfa)](https://github.com/dr-dobermann/gonfa/stargazers)

A universal, lightweight and idiomatic Go library for creating and managing non-deterministic finite automata (NFA). goNFA provides reliable state management mechanisms for complex systems such as business process engines (BPM), workflow systems, and any application requiring sophisticated state machine logic.

## Features

- **Non-deterministic Finite Automata Support**: Full NFA implementation with multiple transitions from the same state with the same event
- **Thread-Safe Operations**: All machine operations are concurrent-safe
- **Fluent Builder API**: Intuitive programmatic state machine construction
- **YAML Configuration**: Load state machine definitions from YAML files
- **State Persistence**: Serialize and restore machine state for long-running processes
- **Extensible Actions & Guards**: Plugin-based system for custom business logic
- **Comprehensive Testing**: >90% test coverage with extensive unit and integration tests
- **Zero External Dependencies**: Core library has no external dependencies (except for YAML support)

## Quick Start

### Installation

```bash
go get github.com/dr-dobermann/gonfa
```

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/dr-dobermann/gonfa/pkg/builder"
    "github.com/dr-dobermann/gonfa/pkg/gonfa"
    "github.com/dr-dobermann/gonfa/pkg/machine"
)

// Simple guard implementation
type ManagerGuard struct{}

func (g *ManagerGuard) Check(ctx context.Context, payload gonfa.Payload) bool {
    // Your business logic here
    return true
}

// Simple action implementation
type NotifyAction struct{}

func (a *NotifyAction) Execute(ctx context.Context, payload gonfa.Payload) error {
    fmt.Println("Notification sent!")
    return nil
}

func main() {
    // Build state machine definition
    definition, err := builder.New().
        InitialState("Draft").
        AddTransition("Draft", "InReview", "Submit").
        WithActions(&NotifyAction{}).
        AddTransition("InReview", "Approved", "Approve").
        WithGuards(&ManagerGuard{}).
        Build()
    if err != nil {
        log.Fatal(err)
    }

    // Create machine instance
    sm := machine.New(definition)
    
    // Fire events
    ctx := context.Background()
    success, err := sm.Fire(ctx, "Submit", "document payload")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Transition successful: %v\n", success)
    fmt.Printf("Current state: %s\n", sm.CurrentState())
}
```

### YAML Configuration

Create a state machine from YAML configuration:

```yaml
# workflow.yaml
initialState: Draft

hooks:
  onSuccess:
    - logSuccess
  onFailure:
    - logFailure

states:
  InReview:
    onEntry:
      - assignReviewer
    onExit:
      - cleanupTask

transitions:
  - from: Draft
    to: InReview
    on: Submit
    actions:
      - notifyAuthor
  - from: InReview
    to: Approved
    on: Approve
    guards: 
      - isManager
```

```go
// Load from YAML
file, err := os.Open("workflow.yaml")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

registry := registry.New()
// Register your guards and actions...
registry.RegisterGuard("isManager", &ManagerGuard{})
registry.RegisterAction("notifyAuthor", &NotifyAction{})

definition, err := definition.Load(file, registry)
if err != nil {
    log.Fatal(err)
}

machine := machine.New(definition)
```

## Architecture

goNFA separates static **Definitions** from dynamic **Machine instances**:

- **Definition**: Immutable description of the state graph, transitions, and associated actions
- **Machine**: Runtime instance that "lives" on the Definition graph with current state
- **Registry**: Maps string names to Guard/Action implementations for YAML loading
- **Builder**: Fluent API for programmatic Definition creation

## Package Structure

- [`pkg/gonfa`](pkg/gonfa/README.md) - Core types and interfaces
- [`pkg/definition`](pkg/definition/README.md) - State machine definitions and YAML loading
- [`pkg/builder`](pkg/builder/README.md) - Fluent API for building definitions
- [`pkg/machine`](pkg/machine/README.md) - Runtime state machine implementation
- [`pkg/registry`](pkg/registry/README.md) - Name-to-object mapping for YAML support
- [`examples/`](examples/) - Usage examples and sample configurations

## Documentation

- [Technical Specification](doc/SDR_Nondetermenistic_Finite_Automation_Go_lib.en.md) - Detailed technical requirements
- [API Documentation](https://pkg.go.dev/github.com/dr-dobermann/gonfa) - GoDoc reference
- [Examples](examples/) - Working code examples
- [Changelog](CHANGELOG.md) - Version history and changes

## Building and Testing

### Prerequisites

- Go 1.21 or later
- Make (optional, for convenience commands)

### Development Commands

```bash
# Install development tools
make install

# Run tests with coverage
make test

# Build library
make build

# Build examples
make examples

# Run specific example
make run-example-document_workflow

# Generate mocks (requires mockery)
make mocks

# Lint code
make lint

# Development cycle (clean, mocks, test)
make dev
```

### Manual Commands

```bash
# Run tests
go test ./pkg/...

# Build examples
go build -o bin/document_workflow examples/document_workflow.go

# Run example
./bin/document_workflow
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass (`make test`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

### Code Standards

- Follow Go conventions and idioms
- Maintain test coverage >90%
- Add comprehensive documentation for public APIs
- Use meaningful commit messages
- Keep line length ≤80 characters where possible

## License

This project is licensed under the GNU Lesser General Public License v2.1 - see the [LICENSE](LICENSE) file for details.

## Author

**dr-dobermann** (rgabtiov@gmail.com)

## Links

- [GitHub Repository](https://github.com/dr-dobermann/gonfa)
- [Issue Tracker](https://github.com/dr-dobermann/gonfa/issues)
- [Discussions](https://github.com/dr-dobermann/gonfa/discussions)
- [Русская версия README](README.ru.md)

---

*goNFA - Bringing the power of non-deterministic finite automata to Go applications.*
