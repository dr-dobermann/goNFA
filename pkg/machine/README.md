# Package machine

The `machine` package provides the runtime implementation of state machines. A Machine represents a dynamic instance that "lives" on a Definition graph and maintains current state and transition history.

## Overview

This package implements the core runtime functionality:
- Thread-safe machine operations
- Event firing and state transitions  
- State persistence and restoration
- Transition history tracking
- NFA (Non-deterministic Finite Automata) behavior

## Types

### Machine

```go
type Machine struct { /* ... */ }
```

The Machine represents a runtime instance of a state machine. All operations are thread-safe.

## Functions

### NewMachine

```go
func NewMachine(def *definition.Definition) *Machine
```

Creates a new Machine instance from a Definition. The machine starts in the initial state specified by the definition.

**Example:**
```go
definition, err := builder.New().
    InitialState("Start").
    FinalStates("End").
    AddTransition("Start", "End", "Finish").
    Build()
if err != nil {
    log.Fatal(err)
}

machine := machine.New(definition, nil)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Initial state: %s\n", machine.CurrentState()) // Output: Start
```

### RestoreMachine

```go
func RestoreMachine(def *definition.Definition, state *gonfa.Storable) (*Machine, error)
```

Restores a Machine instance from a previously serialized state. This enables persistence of long-running processes.

**Example:**
```go
// Restore from JSON data
var storable gonfa.Storable
err := json.Unmarshal(jsonData, &storable)
if err != nil {
    return err
}

machine, err := machine.RestoreMachine(definition, &storable)
if err != nil {
    return err
}

fmt.Printf("Restored state: %s\n", machine.CurrentState())
```

## Methods

### CurrentState

```go
func (m *Machine) CurrentState() gonfa.State
```

Returns the current state of the machine. This method is thread-safe.

### Fire

```go
func (m *Machine) Fire(ctx context.Context, event gonfa.Event, payload gonfa.Payload) (bool, error)
```

Triggers a transition based on an event. This is the main method for advancing the state machine.

**Execution Order:**
1. Find matching transitions for current state and event
2. For each transition (NFA behavior):
   - Check all Guards (must all pass)
   - Execute OnExit actions for current state
   - Execute transition Actions
   - Change state
   - Execute OnEntry actions for new state
   - Call success/failure Hooks
3. Return true if any transition succeeded, false otherwise

**Parameters:**
- `ctx`: Context for cancellation and request-scoped values
- `event`: Event that should trigger the transition
- `payload`: Data to pass to guards and actions

**Returns:**
- `bool`: true if transition succeeded, false otherwise
- `error`: Any error that occurred during transition

**Example:**
```go
ctx := context.Background()
payload := &DocumentPayload{
    ID: "DOC-001",
    Author: "John Doe",
}

success, err := machine.Fire(ctx, "Submit", payload)
if err != nil {
    log.Printf("Transition error: %v", err)
    return
}

if success {
    fmt.Printf("Transition successful, new state: %s\n", machine.CurrentState())
} else {
    fmt.Println("Transition failed (guards rejected or no matching transition)")
}
```

### Marshal

```go
func (m *Machine) Marshal() (*gonfa.Storable, error)
```

Creates a serializable representation of the machine's current state and history.

**Example:**
```go
storable, err := machine.Marshal()
if err != nil {
    return err
}

// Serialize to JSON for persistence
jsonData, err := json.Marshal(storable)
if err != nil {
    return err
}

// Save to database, file, etc.
err = saveToDatabase(jsonData)
```

### History

```go
func (m *Machine) History() []gonfa.HistoryEntry
```

Returns a copy of the machine's transition history. Useful for auditing and debugging.

**Example:**
```go
history := machine.History()
fmt.Printf("Machine has made %d transitions:\n", len(history))

for i, entry := range history {
    fmt.Printf("%d. %s -> %s (event: %s) at %s\n",
        i+1, entry.From, entry.To, entry.On,
        entry.Timestamp.Format("2006-01-02 15:04:05"))
}
```

## NFA Behavior

The machine supports non-deterministic finite automata behavior:

- Multiple transitions can exist from the same state with the same event
- Transitions are tried in the order they were defined
- The first transition where all guards pass will be executed
- If no transition succeeds, the machine remains in the current state

**Example:**
```go
// Multiple transitions with different guards
definition, err := builder.New().
    InitialState("Start").
    FinalStates("Path1", "Path2").
    AddTransition("Start", "Path1", "Event").
    WithGuards(&Guard1{}).
    AddTransition("Start", "Path2", "Event").
    WithGuards(&Guard2{}).
    Build()

// When "Event" is fired, the machine will try Path1 first,
// then Path2 if Guard1 fails
```

## Thread Safety

All Machine operations are thread-safe:

```go
var wg sync.WaitGroup
machine, err := machine.New(definition, nil)
if err != nil {
    log.Fatal(err)
}

// Multiple goroutines can safely fire events
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        ctx := context.Background()
        success, err := machine.Fire(ctx, "Process", fmt.Sprintf("payload-%d", id))
        if err != nil {
            log.Printf("Error in goroutine %d: %v", id, err)
        }
    }(i)
}

wg.Wait()
```

## Error Handling

The machine handles various error scenarios:

- **Guard Failures**: Not considered errors, just prevent transitions
- **Action Errors**: Stop transition and return error
- **Invalid Events**: No matching transition, return false but no error
- **Context Cancellation**: Respected in all operations

**Example with Error Handling:**
```go
success, err := machine.Fire(ctx, "Submit", payload)
switch {
case err != nil:
    // Action failed or other error occurred
    log.Printf("Transition failed with error: %v", err)
    
case !success:
    // Guards rejected or no matching transition
    log.Printf("Transition rejected, staying in state: %s", machine.CurrentState())
    
default:
    // Transition successful
    log.Printf("Transitioned to state: %s", machine.CurrentState())
}
```

## Performance Considerations

- Machine operations have minimal overhead
- History is kept in memory (consider cleanup for long-running processes)
- Guards and actions should be efficient as they're called frequently
- Use context timeouts for potentially long-running actions

## Usage Patterns

### Long-Running Processes

```go
// Periodically save machine state
ticker := time.NewTicker(5 * time.Minute)
defer ticker.Stop()

for {
    select {
    case <-ticker.C:
        storable, err := machine.Marshal()
        if err != nil {
            log.Printf("Failed to marshal machine: %v", err)
            continue
        }
        
        if err := saveMachineState(storable); err != nil {
            log.Printf("Failed to save machine state: %v", err)
        }
        
    case event := <-eventChannel:
        success, err := machine.Fire(ctx, event.Type, event.Payload)
        if err != nil {
            log.Printf("Event processing failed: %v", err)
        }
    }
}
```

### Event-Driven Architecture

```go
type EventProcessor struct {
    machine *machine.Machine
}

func (p *EventProcessor) ProcessEvent(ctx context.Context, event Event) error {
    success, err := p.machine.Fire(ctx, gonfa.Event(event.Type), event.Payload)
    if err != nil {
        return fmt.Errorf("failed to process event %s: %w", event.Type, err)
    }
    
    if !success {
        log.Printf("Event %s was not processed (no valid transition)", event.Type)
    }
    
    return nil
}
```

## Related Packages

- [`pkg/gonfa`](../gonfa/README.md) - Core types and interfaces used by Machine
- [`pkg/definition`](../definition/README.md) - Definition structure that Machine operates on
- [`pkg/builder`](../builder/README.md) - Creates definitions for Machine instances
