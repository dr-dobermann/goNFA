# Package builder

The `builder` package provides a fluent interface for programmatically creating state machine definitions. It allows step-by-step construction of complex state machines with a readable, chainable API.

## Overview

The Builder pattern implementation enables:
- Fluent, chainable API for definition creation
- Validation during construction
- Type-safe state machine building
- Clear separation between definition and runtime

## Usage

### Basic Usage

```go
definition, err := builder.New().
    InitialState("Draft").
    FinalStates("Approved").
    AddTransition("Draft", "InReview", "Submit").
    AddTransition("InReview", "Approved", "Approve").
    Build()
```

### Complete Example

```go
guard := &ManagerGuard{}
action := &NotifyAction{}

definition, err := builder.New().
    InitialState("Draft").
    FinalStates("Approved").
    OnEntry("InReview", &AssignReviewerAction{}).
    OnExit("InReview", &CleanupAction{}).
    AddTransition("Draft", "InReview", "Submit").
    WithActions(action).
    AddTransition("InReview", "Approved", "Approve").
    WithGuards(guard).
    WithActions(&ApprovalAction{}).
    WithSuccessHooks(&LogSuccessAction{}).
    WithFailureHooks(&LogFailureAction{}).
    Build()
```

## API Reference

See [GoDoc](https://pkg.go.dev/github.com/dr-dobermann/gonfa/pkg/builder) for complete API documentation.
