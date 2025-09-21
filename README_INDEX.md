# goNFA Documentation Index

This document provides links to all documentation available in the goNFA project.

## Project Overview

- **[Main README](README.md)** - Project overview, quick start, and usage examples
- **[Changelog](CHANGELOG.md)** - Version history and release notes
- **[License](LICENSE)** - LGPL 2.1 license terms

## Technical Documentation

- **[Technical Specification](doc/SDR_Nondetermenistic_Finite_Automation_Go_lib.en.md)** - Complete technical requirements and architecture specification

## Package Documentation

Each package contains detailed documentation about its specific functionality:

### Core Packages

- **[pkg/gonfa](pkg/gonfa/README.md)** - Core types, interfaces, and main API
  - Basic types: State, Event, Payload
  - Core interfaces: Guard, Action
  - Serialization types: Storable, HistoryEntry

- **[pkg/definition](pkg/definition/README.md)** - State machine definitions and YAML loading
  - Definition creation and management
  - YAML file loading with registry support
  - Transition and state configuration

- **[pkg/builder](pkg/builder/README.md)** - Fluent API for programmatic definition creation
  - Builder pattern implementation
  - Fluent interface for state machine construction
  - Validation and error handling

- **[pkg/machine](pkg/machine/README.md)** - Runtime state machine implementation
  - Thread-safe machine operations
  - Event firing and state transitions
  - State persistence and restoration

- **[pkg/registry](pkg/registry/README.md)** - Name-to-object mapping for YAML support
  - Guard and Action registration
  - Thread-safe registry operations
  - Name resolution for YAML loading

## Examples and Usage

- **[examples/](examples/)** - Working code examples
  - **[document_workflow.go](examples/document_workflow.go)** - Complete document approval workflow example
  - **[document_workflow.yaml](examples/document_workflow.yaml)** - YAML configuration example

## Development Documentation

- **[Makefile](Makefile)** - Build system and development commands
- **[.mockery.yaml](.mockery.yaml)** - Mock generation configuration
- **[go.mod](go.mod)** - Go module dependencies

## API Reference

- **[GoDoc](https://pkg.go.dev/github.com/dr-dobermann/gonfa)** - Complete API documentation (online)

## Quick Navigation

| Component | Purpose | Documentation |
|-----------|---------|---------------|
| Core API | Main library interface | [pkg/gonfa](pkg/gonfa/README.md) |
| Definitions | State machine structure | [pkg/definition](pkg/definition/README.md) |
| Builder | Programmatic creation | [pkg/builder](pkg/builder/README.md) |
| Machine | Runtime execution | [pkg/machine](pkg/machine/README.md) |
| Registry | YAML support | [pkg/registry](pkg/registry/README.md) |
| Examples | Usage patterns | [examples/](examples/) |

## Getting Started

1. Read the [Main README](README.md) for project overview
2. Check the [Technical Specification](doc/SDR_Nondetermenistic_Finite_Automation_Go_lib.en.md) for detailed requirements
3. Explore [examples/](examples/) for practical usage patterns
4. Refer to package-specific documentation for detailed API information

---

*For questions or contributions, see the main [README](README.md) file.*
