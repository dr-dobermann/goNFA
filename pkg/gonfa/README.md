# Package gonfa

The `gonfa` package provides the core types and interfaces for the goNFA state machine library.

## Overview

This package contains the fundamental building blocks used across all other packages in the goNFA library:
- Basic types for states, events, and payloads
- Core interfaces for guards and actions
- Serialization types for state persistence

## Types

### Basic Types

#### State
```go
type State string
```
Represents a state in the state machine. States are string-based identifiers that should be unique within a definition.

**Example:**
```go
draftState := gonfa.State("Draft")
reviewState := gonfa.State("InReview")
```

#### Event
```go
type Event string
```
Represents an event that triggers transitions between states.

**Example:**
```go
submitEvent := gonfa.Event("Submit")
approveEvent := gonfa.Event("Approve")
```

#### Payload
```go
type Payload interface{}
```
Interface for passing runtime data during transitions. Can be any type.

**Example:**
```go
type DocumentPayload struct {
    ID     string
    Author string
    Title  string
}

payload := &DocumentPayload{
    ID:     "DOC-001",
    Author: "John Doe",
    Title:  "Project Proposal",
}
```

### Core Interfaces

#### Guard
```go
type Guard interface {
    Check(ctx context.Context, payload Payload) bool
}
```
Guards control whether transitions can occur. They evaluate conditions and return true if the transition should be allowed.

**Implementation Example:**
```go
type ManagerGuard struct {
    requiredRole string
}

func (g *ManagerGuard) Check(ctx context.Context, payload gonfa.Payload) bool {
    // Extract user role from context or payload
    userRole := getUserRole(ctx)
    return userRole == g.requiredRole
}
```

#### Action
```go
type Action interface {
    Execute(ctx context.Context, payload Payload) error
}
```
Actions perform business logic during transitions, state entry/exit, or as hooks. They can modify external systems, send notifications, etc.

**Implementation Example:**
```go
type NotificationAction struct {
    emailService EmailService
}

func (a *NotificationAction) Execute(ctx context.Context, payload gonfa.Payload) error {
    doc, ok := payload.(*DocumentPayload)
    if !ok {
        return fmt.Errorf("expected DocumentPayload, got %T", payload)
    }
    
    return a.emailService.SendNotification(doc.Author, "Document submitted")
}
```

### Serialization Types

#### HistoryEntry
```go
type HistoryEntry struct {
    From      State     `json:"from"`
    To        State     `json:"to"`
    On        Event     `json:"on"`
    Timestamp time.Time `json:"timestamp"`
}
```
Records a single transition in the machine's history for audit and debugging purposes.

#### Storable
```go
type Storable struct {
    CurrentState State          `json:"currentState"`
    History      []HistoryEntry `json:"history"`
}
```
Represents the serializable state of a Machine instance. This structure can be marshaled to JSON for persistence and later restored.

**Usage Example:**
```go
// Serialize machine state
storable, err := machine.Marshal()
if err != nil {
    return err
}

jsonData, err := json.Marshal(storable)
if err != nil {
    return err
}

// Save to database or file
err = saveToDatabase(jsonData)

// Later, restore machine state
var restoredStorable gonfa.Storable
err = json.Unmarshal(jsonData, &restoredStorable)
if err != nil {
    return err
}

restoredMachine, err := machine.RestoreMachine(definition, &restoredStorable)
```

## Usage Patterns

### Implementing Guards

Guards should be stateless and thread-safe:

```go
type RoleBasedGuard struct {
    allowedRoles []string
}

func (g *RoleBasedGuard) Check(ctx context.Context, payload gonfa.Payload) bool {
    userRole := ctx.Value("userRole").(string)
    for _, role := range g.allowedRoles {
        if role == userRole {
            return true
        }
    }
    return false
}
```

### Implementing Actions

Actions can have side effects but should handle errors gracefully:

```go
type DatabaseUpdateAction struct {
    db Database
}

func (a *DatabaseUpdateAction) Execute(ctx context.Context, payload gonfa.Payload) error {
    doc := payload.(*Document)
    
    if err := a.db.UpdateStatus(doc.ID, "in_review"); err != nil {
        return fmt.Errorf("failed to update database: %w", err)
    }
    
    return nil
}
```

### Context Usage

Always use the provided context for cancellation, timeouts, and carrying request-scoped values:

```go
func (a *HTTPCallAction) Execute(ctx context.Context, payload gonfa.Payload) error {
    req, err := http.NewRequestWithContext(ctx, "POST", a.url, nil)
    if err != nil {
        return err
    }
    
    resp, err := a.client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    return nil
}
```

## Thread Safety

All types in this package are designed to be thread-safe when used as intended:
- Basic types (State, Event) are immutable
- Interfaces should be implemented in a thread-safe manner
- Serialization types are value types and safe for concurrent access

## Best Practices

1. **State Names**: Use descriptive, consistent naming for states
2. **Event Names**: Use action-oriented names for events (Submit, Approve, Reject)
3. **Guards**: Keep guard logic simple and focused on a single condition
4. **Actions**: Make actions idempotent when possible
5. **Error Handling**: Always handle and wrap errors appropriately
6. **Context**: Respect context cancellation in long-running operations

## Related Packages

- [`pkg/definition`](../definition/README.md) - Uses these types to define state machines
- [`pkg/builder`](../builder/README.md) - Fluent API for creating definitions
- [`pkg/machine`](../machine/README.md) - Runtime implementation using these interfaces
- [`pkg/registry`](../registry/README.md) - Maps string names to Guard/Action implementations
