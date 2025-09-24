// Package gonfa provides the main API for the goNFA state machine library.
//
// goNFA is a universal, lightweight and idiomatic Go library for creating
// and managing non-deterministic finite automata (NFA). It provides reliable
// state management mechanisms for complex systems such as business process
// engines (BPM).
//
// This package contains fundamental types and interfaces used across all
// other packages in the goNFA state machine library.
//
// Project: https://github.com/dr-dobermann/gonfa
// Author: dr-dobermann (rgabtiov@gmail.com)
// License: LGPL-2.1 (see LICENSE file in the project root)
package gonfa

import (
	"context"
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
	// CurrentState returns the current state of the machine.
	CurrentState() State
	// History returns the transition history.
	History() []HistoryEntry
	// IsInFinalState checks if the machine is currently in a final (accepting) state.
	IsInFinalState() bool
	// StateExtender returns the attached user-defined business object.
	StateExtender() StateExtender
}

// Guard is the interface for guard objects.
// Guards are used to control whether a transition can occur.
type Guard interface {
	// Check evaluates whether the transition should be allowed.
	// Returns true if the transition is permitted, false otherwise.
	Check(ctx context.Context, state MachineState, payload Payload) bool
}

// Action is the interface for action and hook objects.
// Actions are executed during transitions, state entry/exit, or as hooks.
type Action interface {
	// Execute performs the action with the given context, state, and payload.
	// Returns an error if the action fails.
	Execute(ctx context.Context, state MachineState, payload Payload) error
}

// HistoryEntry records a single transition in the machine's history.
type HistoryEntry struct {
	From      State     `json:"from"`
	To        State     `json:"to"`
	On        Event     `json:"on"`
	Timestamp time.Time `json:"timestamp"`
}

// Storable represents a serializable state of a Machine instance.
// This structure can be marshaled to JSON for persistence.
type Storable struct {
	CurrentState State          `json:"currentState"`
	History      []HistoryEntry `json:"history"`
}
