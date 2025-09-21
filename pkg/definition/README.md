# Package definition

The `definition` package provides structures and functions for creating immutable state machine definitions. A Definition describes the static structure of a state machine including states, transitions, and hooks.

## Overview

This package handles:
- Immutable state machine definitions
- YAML file loading with registry support
- Transition and state configuration
- Definition validation

## Types

### Definition
Immutable description of the state machine graph containing states, transitions, and hooks.

### Transition
Describes a single transition between states with associated guards and actions.

### StateConfig
Configuration for state-specific entry and exit actions.

### Hooks
Global success and failure hooks for the state machine.

## Usage

### YAML Loading

```go
registry := registry.New()
// ... register guards and actions ...

file, err := os.Open("definition.yaml")
if err != nil {
    return err
}
defer file.Close()

definition, err := definition.LoadDefinition(file, registry)
if err != nil {
    return err
}

machine := machine.NewMachine(definition)
```

### YAML Format

```yaml
initialState: Draft

hooks:
  onSuccess: [logSuccess]
  onFailure: [logFailure]

states:
  InReview:
    onEntry: [assignReviewer]
    onExit: [cleanupTask]

transitions:
  - from: Draft
    to: InReview
    on: Submit
    guards: [hasPermission]
    actions: [notifyAuthor]
```

## API Reference

See [GoDoc](https://pkg.go.dev/github.com/dr-dobermann/gonfa/pkg/definition) for complete API documentation.
