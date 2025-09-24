# SDR: Nondeterministic Finite Automation library (goNFA)

**Version:** 3.8
**Date:** September 20, 2025

## 1. Introduction

### 1.1. Library Purpose

`goNFA` is a universal, lightweight and idiomatic Go library for creating and managing non-deterministic finite automata (NFA).

### 1.2. Scope of Application

The main application is providing a reliable state management mechanism for complex systems, such as business process engines (BPM). The library is designed as a universal solution that can be used in any other projects requiring implementation of complex state logic, especially in long-running processes.

### 1.3. Terminology

* **Definition**: A static, immutable structure describing the state graph, transitions, and hooks.

* **Machine**: A dynamic object that "lives" on the Definition graph.

* **MachineState**: An interface for read-only access to the Machine instance state.

* **StateExtender**: A user-defined business object attached to the Machine instance.

* **Payload**: Arbitrary event-specific data passed to the machine during transitions.

* **Guard**: An object **bound to a specific transition** responsible for checking transition execution conditions. The Guard chain works on the middleware principle: a transition is allowed only if each Guard in the chain returns `true`.

* **Action**: An object that performs useful work. **Transition, entry, and exit actions are bound to specific transitions or states.** The Action chain executes sequentially within the transition transaction.

* **Hook**: An `Action` **bound to the entire machine** that is called after *any* transition attempt (successful or unsuccessful). The Hook chain works on the middleware principle, executing sequentially for logging, metrics, or other cross-cutting tasks.

* **Registry**: An object that maps string names to `Guard` and `Action` implementations.

## 2. High-Level Architecture

The library separates static **Definition** and its dynamic **Machine instances**. Definition describes the state graph and transitions. A Machine instance contains the current FSM state and carries a user-defined business object (`StateExtender`), providing full context for `Guard`s and `Action`s.

* **Definition** is created programmatically using a **Builder** or loaded from a file.

* **Machine** is created from Definition, works at runtime, and can be saved and restored.

* All operations on Machine instances are thread-safe.

## 3. API Design and Data Structures

### 3.1. Core Types and Interfaces

```go
import (
	"context"
	"io"
	"time"
)

// State represents a state in the state machine.
type State string

// Event represents an event that triggers a transition.
type Event string

// Payload is an interface for passing event-specific runtime data.
type Payload interface{}

// StateExtender is a placeholder for any user-defined business object.
type StateExtender interface{}

// MachineState provides a read-only view of the machine's state.
type MachineState interface {
	CurrentState() State
	History() []HistoryEntry
	IsInFinalState() bool
	// StateExtender returns the attached user-defined business object.
	StateExtender() StateExtender
}

// Guard is the interface for guard objects.
type Guard interface {
	Check(ctx context.Context, state MachineState, payload Payload) bool
}

// Action is the interface for action and hook objects.
type Action interface {
	Execute(ctx context.Context, state MachineState, payload Payload) error
}
```

### 3.2. Object Registry

```go
// Assuming this code is in package 'registry'

// Registry stores a mapping from string names to real objects.
type Registry struct { /* ... */ }

// New creates a new Registry.
func New() *Registry { /* ... */ }

// RegisterGuard registers a guard object under a unique name.
func (r *Registry) RegisterGuard(name string, guard Guard) error { /* ... */ }

// RegisterAction registers an action (or hook) object under a unique name.
func (r *Registry) RegisterAction(name string, action Action) error { /* ... */ }
```

### 3.3. Machine Definition

```go
// Assuming this code is in package 'definition'

// Transition describes one possible transition.
type Transition struct {
	From    State
	To      State
	On      Event
	Guards  []Guard  // A chain of guards.
	Actions []Action // A chain of transition actions.
}

// StateConfig describes actions associated with a specific state.
type StateConfig struct {
	OnEntry []Action // Actions to execute upon entering the state.
	OnExit  []Action // Actions to execute upon exiting the state.
}

// Hooks describes a set of hooks for the state machine.
type Hooks struct {
	OnSuccess []Action
	OnFailure []Action
}

// Definition is an immutable description of the state graph.
type Definition struct {
	InitialState  State
	FinalStates   map[State]bool // Set of final (accepting) states.
	States        map[State]StateConfig
	Hooks         Hooks
	// internal fields...
}

// Load loads a definition from an io.Reader using a registry and validates it.
func Load(r io.Reader, registry *Registry) (*Definition, error) { /* ... */ }
```

### 3.4. Programmatic Builder

```go
// Assuming this code is in a package like 'builder'

// Builder provides a fluent interface for creating a Definition.
type Builder struct { /* ... */ }

// New creates a new Builder.
func New() *Builder { /* ... */ }

// InitialState sets the initial state for the state machine.
func (b *Builder) InitialState(s State) *Builder { /* ... */ }

// FinalStates sets the final (accepting) states for the state machine.
// Can be called multiple times to add more states.
func (b *Builder) FinalStates(states ...State) *Builder { /* ... */ }

// OnEntry defines actions to be executed upon EVERY entry into the specified state.
func (b *Builder) OnEntry(s State, actions ...Action) *Builder { /* ... */ }

// OnExit defines actions to be executed upon EVERY exit from the specified state.
func (b *Builder) OnExit(s State, actions ...Action) *Builder { /* ... */ }

// AddTransition adds a new transition.
func (b *Builder) AddTransition(from State, to State, on Event) *Builder { /* ... */ }

// WithGuards adds guards to the LAST added transition.
func (b *Builder) WithGuards(guards ...Guard) *Builder { /* ... */ }

// WithActions adds actions to the LAST added transition.
func (b *Builder) WithActions(actions ...Action) *Builder { /* ... */ }

// WithHooks sets global hooks for the state machine.
func (b *Builder) WithHooks(hooks Hooks) *Builder { /* ... */ }

// Build finalizes the building process, performs full validation, and returns an immutable Definition.
func (b *Builder) Build() (*Definition, error) { /* ... */ }
```

### 3.5. Machine Instance

```go
// Assuming this code is in package 'machine' or 'gonfa'

// HistoryEntry is for recording transition history.
type HistoryEntry struct {
	From      State
	To        State
	On        Event
	Timestamp time.Time
}

// Storable represents a serializable state of a Machine instance.
type Storable struct {
	CurrentState State          `json:"currentState"`
	History      []HistoryEntry `json:"history"`
}

// Machine represents an instance of a state machine.
// It automatically satisfies the MachineState interface.
type Machine struct {
    // ...
    history       []HistoryEntry
    stateExtender StateExtender
}

// New creates a new instance of the state machine from a definition,
// attaching a user-defined business object as its state extender.
func New(def *Definition, extender StateExtender) *Machine { /* ... */ }

// Restore restores an instance of the state machine from a storable state,
// attaching a user-defined business object as its state extender.
func Restore(def *Definition, state *Storable, extender StateExtender) (*Machine, error) { /* ... */ }

// Fire triggers a transition based on an event. See section 3.7 for execution order and error handling.
func (m *Machine) Fire(ctx context.Context, event Event, payload Payload) (bool, error) { /* ... */ }

// CurrentState returns the current state.
func (m *Machine) CurrentState() State { /* ... */ }

// History returns the transition history.
func (m *Machine) History() []HistoryEntry { return m.history }

// IsInFinalState checks if the machine is currently in a final (accepting) state.
func (m *Machine) IsInFinalState() bool { /* ... */ }

// StateExtender returns the attached user-defined business object.
func (m *Machine) StateExtender() StateExtender { return m.stateExtender }

// Marshal creates a serializable representation of the instance's state.
func (m *Machine) Marshal() (*Storable, error) { /* ... */ }
```

### 3.6. Definition Validation Rules

The `builder.Build()` and `definition.Load()` functions must perform full validation of correctness and integrity of the definition before creating it. A definition is considered valid if all the following conditions are met:

1. **Initial state is defined:** `initialState` cannot be empty.

2. **All states must be explicitly declared:** Any state name used in `initialState`, `finalStates`, `transitions` (in `from` and `to` fields) must be declared either through `OnEntry`/`OnExit` calls in the builder, or present in the `states` section in the YAML file. This rule guarantees the absence of transitions to "non-existent" states.

3. **Unreachable states:** Optionally, the validator may issue a warning if there are states in the definition that cannot be reached from the initial state (except the initial state itself).

### 3.7. Error Handling and Transactionality

A transition from one state to another is an **atomic operation**. The `Fire` method guarantees that the machine state remains consistent.

**`Fire` execution order:**

1. Find all transitions matching the current state and event.

2. For each found transition, execute `Guard` checks. The first transition for which all `Guard`s return `true` is selected. If no transition is suitable, the operation completes without changing state.

3. **Begin transaction:**
   a. Execute all `Action`s from `OnExit` for the current state.
   b. Execute all `Action`s of the transition itself.
   c. Execute all `Action`s from `OnEntry` for the **target** state.

4. **Complete transaction:**

   * **On success (all `Action`s returned `nil`):**

     * The machine state atomically changes to the target state.

     * A new entry is added to history.

     * `OnSuccess` hooks are called.

     * The method returns `(true, nil)`.

   * **On error (any `Action` returned an error):**

     * The operation is immediately aborted.

     * **The machine state does NOT change.**

     * `OnFailure` hooks are called.

     * The method returns `(false, err)`, where `err` is the original error from the `Action`.

This approach guarantees that the machine will not end up in an inconsistent state and gives the application full control over how to respond to business logic execution errors.

## 4. Recommended Usage Pattern

With the new architecture, the usage pattern becomes significantly cleaner. `Payload` is used for passing event-specific data, while the main business context is available through `MachineState`.

```go
// 1. Your business object in goBPM
type Instance struct {
    ID string
    Document DocumentData
    State *machine.Storable // FSM state lives inside your object
}

// 2. Guard implementation
type ManagerGuard struct {}

func (g *ManagerGuard) Check(ctx context.Context, state machine.MachineState, payload machine.Payload) bool {
    // Get the main business object through MachineState
    instance, ok := state.StateExtender().(*Instance)
    if !ok { return false /* log error */ }
    
    // Get event data from Payload (if any)
    // approvalParams, _ := payload.(ApprovalParams)

    user := user.FromContext(ctx)
    return user.IsManager && user.Department == instance.Document.Department
}

// 3. Usage in service
func (s *Service) ApproveDocument(ctx context.Context, instanceID string, params ApprovalParams) error {
    // Load your business object
    instance, err := s.repo.GetInstance(instanceID)
    if err != nil { return err }

    // "Revive" the machine, ATTACHING your business object to it
    m, err := machine.Restore(s.definition, instance.State, instance)
    if err != nil { return err }
    
    // Call Fire, passing only event data in payload
    changed, err := m.Fire(ctx, "Approve", params)
    if err != nil { return err }

    if changed {
        storable, _ := m.Marshal()
        instance.State = storable
        return s.repo.SaveInstance(instance)
    }
    
    return nil
}
```

## 5. Non-Functional Requirements

* **Performance**: Minimal overhead.

* **Reliability**: Full test coverage (>90%).

* **Documentation**: Comprehensive godoc comments.

* **Dependencies**: Minimal number of external dependencies.

## 6. Usage Example

```go
// Assuming 'definition' and 'machine' are separate packages

// ... (registering guard/action objects in the Registry) ...

// Loading the definition from a file
file, _ := os.Open("definition.yaml")
def, err := definition.Load(file, registry)
// ...

// Your business object
type MyProcess struct {
    ID string
    Data string
    FSMState *machine.Storable
}
process := &MyProcess{ID: "p1", Data: "initial data"}

// Creating a new machine instance, attaching your business object
m := machine.New(def, process)

// Firing an event with event-specific payload
ctx := context.Background()
eventParams := map[string]string{"user": "Ruslan"}
m.Fire(ctx, "Submit", eventParams)

// Check if the machine has reached an accepting state
if m.IsInFinalState() {
    // ... logic for a completed process
}

// 1. Saving the state
storable, err := m.Marshal()
if err != nil { /* ... */ }
process.FSMState = storable

// ... (save your entire 'process' object to DB)

// ... (later, in another process)

// 2. Restoring the state
// Load your 'process' object from DB first
loadedProcess, err := repo.Get("p1")
if err != nil { /* ... */ }

// Restore the machine instance, re-attaching your business object
restoredMachine, err := machine.Restore(def, loadedProcess.FSMState, loadedProcess)
if err != nil { /* ... */ }
```

## Appendix A: Definition File Structure Example

```yaml
# The initial state of the machine
initialState: Draft

# The final (accepting) states of the machine
finalStates:
  - Approved
  - Archived

# Global hooks
hooks:
  onSuccess:
    - logSuccess
  onFailure:
    - logFailure

# Description of state-specific actions.
# All states must be listed here, even if they have no actions.
states:
  Draft: {} # Explicitly defined state
  InReview:
    onEntry:
      - assignReviewer
    onExit:
      - cleanupTask
  Approved:
    onEntry:
      - archiveDocument
  Archived: {}
  Rejected: {}

# List of all possible transitions
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

  - from: Rejected
    to: InReview
    on: Rework
```

## Appendix B: JSON Schema for Definition File

**Note:** This schema validates only the *structure* of the file. Semantic validation (e.g., checking that all used states are declared in the `states` section) is performed by the library code during loading (see section 3.6).

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "https://example.com/gonfa-definition.schema.json",
  "title": "goNFA Definition",
  "description": "Schema for a goNFA state machine definition file.",
  "type": "object",
  "properties": {
    "initialState": {
      "description": "The initial state of the machine.",
      "type": "string"
    },
    "finalStates": {
      "description": "A list of final (accepting) states.",
      "type": "array",
      "items": { "type": "string" },
      "uniqueItems": true
    },
    "hooks": {
      "description": "Global hooks for all transitions.",
      "type": "object",
      "properties": {
        "onSuccess": {
          "type": "array",
          "items": { "type": "string" },
          "uniqueItems": true
        },
        "onFailure": {
          "type": "array",
          "items": { "type": "string" },
          "uniqueItems": true
        }
      },
      "additionalProperties": false
    },
    "states": {
      "description": "State-specific entry and exit actions. All states used in the machine must be defined here.",
      "type": "object",
      "patternProperties": {
        "^.+$": {
          "type": "object",
          "properties": {
            "onEntry": {
              "type": "array",
              "items": { "type": "string" },
              "uniqueItems": true
            },
            "onExit": {
              "type": "array",
              "items": { "type": "string" },
              "uniqueItems": true
            }
          },
          "additionalProperties": false
        }
      }
    },
    "transitions": {
      "description": "The list of all possible transitions.",
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "from": { "type": "string" },
          "to": { "type": "string" },
          "on": { "type": "string" },
          "guards": {
            "type": "array",
            "items": { "type": "string" },
            "uniqueItems": true
          },
          "actions": {
            "type": "array",
            "items": { "type": "string" },
            "uniqueItems": true
          }
        },
        "required": ["from", "to", "on"],
        "additionalProperties": false
      }
    }
  },
  "required": ["initialState", "states", "transitions"],
  "additionalProperties": false
}
```
