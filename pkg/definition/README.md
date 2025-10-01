# Package definition

The `definition` package provides structures and functions for creating immutable state machine definitions. A Definition describes the static structure of a state machine including states, transitions, and hooks.

## Overview

This package handles:
- Immutable state machine definitions
- YAML file loading with registry support
- Transition and state configuration
- Comprehensive definition validation and integrity checking

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

## Definition Validation

The package performs comprehensive integrity checking when creating definitions:

### Validation Rules

1. **State Existence**: All states referenced in transitions must exist in the states map
2. **Initial State**: Must exist in the states map and have outgoing transitions
3. **Final States**: Must exist in the states map and have no outgoing transitions
4. **Duplicate Transitions**: Exact duplicates (same From, To, Event) are forbidden
5. **Connectivity**: 
   - No hanging states (states with no incoming transitions except initial)
   - No dead-end states (non-final states with no outgoing transitions)
   - All final states must be reachable from the initial state

### Error Examples

```go
// Duplicate transition error
transitions := []Transition{
    {From: "A", To: "B", On: "event1"},
    {From: "A", To: "B", On: "event1"}, // Exact duplicate - ERROR
}

// Different events are allowed
transitions := []Transition{
    {From: "A", To: "B", On: "event1"},
    {From: "A", To: "B", On: "event2"}, // Different event - OK
}

// Hanging state error
states := []State{"Start", "Hanging", "End"}
transitions := []Transition{
    {From: "Start", To: "End", On: "finish"},
    // No transitions TO "Hanging" - ERROR
}
```

## API Reference

See [GoDoc](https://pkg.go.dev/github.com/dr-dobermann/gonfa/pkg/definition) for complete API documentation.
