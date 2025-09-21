# Changelog

All notable changes to the goNFA project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial implementation of goNFA library
- Core types and interfaces (State, Event, Payload, Guard, Action)
- Thread-safe Registry for mapping string names to Guard/Action objects
- Immutable Definition structure for state machine descriptions
- Fluent Builder API for programmatic state machine construction
- Runtime Machine implementation with thread-safe operations
- YAML file loading support for state machine definitions
- State persistence and restoration capabilities
- Comprehensive test suite with >90% coverage
- Complete documentation and examples
- Makefile with development commands
- Mock generation support with mockery

### Core Features
- **Non-deterministic Finite Automata**: Full NFA support with multiple transitions
- **Thread Safety**: All operations are concurrent-safe
- **State Persistence**: Serialize/deserialize machine state for long-running processes
- **YAML Configuration**: Load definitions from human-readable YAML files
- **Extensible Architecture**: Plugin system for custom guards and actions
- **Fluent API**: Intuitive builder pattern for programmatic construction

### Package Structure
- `pkg/gonfa` - Core types and main API
- `pkg/definition` - State machine definitions and YAML loading
- `pkg/builder` - Fluent API for building definitions
- `pkg/machine` - Runtime state machine implementation
- `pkg/registry` - Name-to-object mapping for YAML support

### Examples
- Document workflow example with complete state machine implementation
- YAML configuration examples
- Usage patterns and best practices

### Documentation
- Complete API documentation with GoDoc comments
- Technical specification document
- Package-level README files
- Comprehensive usage examples
- Development and contribution guidelines

### Development Tools
- Makefile with common development tasks
- Mock generation configuration
- Comprehensive test suite
- Linting and formatting tools
- CI/CD ready structure

## Project Information

- **Author**: dr-dobermann (rgabtiov@gmail.com)
- **License**: LGPL-2.1
- **Go Version**: 1.21+
- **Repository**: https://github.com/dr-dobermann/gonfa

---

*This changelog follows the [Keep a Changelog](https://keepachangelog.com/) format.*
