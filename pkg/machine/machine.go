// Package machine provides the runtime implementation of state machines.
// A Machine represents a dynamic instance that "lives" on a Definition graph
// and maintains current state and transition history.
//
// goNFA is a universal, lightweight and idiomatic Go library for creating
// and managing non-deterministic finite automata (NFA). It provides reliable
// state management mechanisms for complex systems such as business process
// engines (BPM).
//
// Project: https://github.com/dr-dobermann/gonfa
// Author: dr-dobermann (rgabtiov@gmail.com)
// License: LGPL-2.1 (see LICENSE file in the project root)
package machine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dr-dobermann/gonfa/pkg/definition"
	"github.com/dr-dobermann/gonfa/pkg/gonfa"
)

// Machine represents an instance of a state machine.
// All operations on Machine are thread-safe.
// Machine automatically satisfies the MachineState interface.
type Machine struct {
	mu            sync.RWMutex
	definition    *definition.Definition
	currentState  gonfa.State
	history       []gonfa.HistoryEntry
	stateExtender gonfa.StateExtender
}

// New creates a new Machine instance from a Definition,
// attaching a user-defined business object as its state extender.
func New(def *definition.Definition, extender gonfa.StateExtender) (*Machine, error) {
	if def == nil {
		return nil, fmt.Errorf("definition cannot be nil")
	}

	return &Machine{
		definition:    def,
		currentState:  def.InitialState(),
		history:       make([]gonfa.HistoryEntry, 0),
		stateExtender: extender,
	}, nil
}

// Restore restores a Machine instance from a Storable state,
// attaching a user-defined business object as its state extender.
func Restore(
	def *definition.Definition,
	state *gonfa.Storable,
	extender gonfa.StateExtender,
) (*Machine, error) {
	if def == nil {
		return nil, fmt.Errorf("definition cannot be nil")
	}

	if state == nil {
		return nil, fmt.Errorf("storable state cannot be nil")
	}

	if state.CurrentState == "" {
		return nil, fmt.Errorf("current state cannot be empty")
	}

	if _, exists := def.States()[state.CurrentState]; !exists {
		return nil,
			fmt.Errorf("current state '%s' not found in definition",
				state.CurrentState)
	}

	return &Machine{
		definition:    def,
		currentState:  state.CurrentState,
		history:       append([]gonfa.HistoryEntry{}, state.History...),
		stateExtender: extender,
	}, nil
}

// CurrentState returns the current state of the machine.
func (m *Machine) CurrentState() gonfa.State {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentState
}

// Fire triggers a transition based on an event with the provided payload.
// The method is thread-safe and follows this execution order:
// 1. Find matching transitions
// 2. Check all Guards
// 3. Execute OnExit actions for current state
// 4. Execute transition Actions
// 5. Change state
// 6. Execute OnEntry actions for new state
// 7. Call appropriate Hooks (OnSuccess/OnFailure)
func (m *Machine) Fire(
	ctx context.Context,
	event gonfa.Event,
	payload gonfa.Payload,
) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find possible transitions
	transitions := m.definition.GetTransitions(m.currentState, event)

	// For NFA, try each transition until one succeeds
	for _, transition := range transitions {
		success, err := m.attemptTransition(ctx, transition, payload)
		if err != nil {
			// Call failure hooks and return error
			if hookErr := m.callHooks(ctx, payload, false); hookErr != nil {
				return false, fmt.Errorf("transition failed: %v, hook error: %v",
					err, hookErr)
			}

			return false, err
		}

		if success {
			// Transition succeeded, call success hooks
			return true, m.callHooks(ctx, payload, true)
		}
	}

	// No transition succeeded, call failure hooks
	return false, m.callHooks(ctx, payload, false)
}

// attemptTransition attempts to execute a single transition.
// Returns true if successful, false if guards failed, error on action failure.
func (m *Machine) attemptTransition(
	ctx context.Context,
	transition definition.Transition,
	payload gonfa.Payload,
) (bool, error) {
	// 1. Check all guards
	for _, guard := range transition.Guards {
		if !guard.Check(ctx, m, payload) {
			return false, nil // Guard failed, try next transition
		}
	}

	// 2. Execute OnExit actions for current state
	currentConfig := m.definition.GetStateConfig(m.currentState)
	for _, action := range currentConfig.OnExit {
		if err := action.Execute(ctx, m, payload); err != nil {
			return false, fmt.Errorf("OnExit action failed: %w", err)
		}
	}

	// 3. Execute transition actions
	for _, action := range transition.Actions {
		if err := action.Execute(ctx, m, payload); err != nil {
			return false, fmt.Errorf("transition action failed: %w", err)
		}
	}

	// 4. Change state and record history
	oldState := m.currentState
	m.currentState = transition.To

	historyEntry := gonfa.HistoryEntry{
		From:      oldState,
		To:        transition.To,
		On:        transition.On,
		Timestamp: time.Now(),
	}
	m.history = append(m.history, historyEntry)

	// 5. Execute OnEntry actions for new state
	newConfig := m.definition.GetStateConfig(m.currentState)
	for _, action := range newConfig.OnEntry {
		if err := action.Execute(ctx, m, payload); err != nil {
			// Transition already happened, but OnEntry failed
			return false, fmt.Errorf("OnEntry action failed: %w", err)
		}
	}

	return true, nil
}

// callHooks executes the appropriate global hooks.
func (m *Machine) callHooks(
	ctx context.Context,
	payload gonfa.Payload,
	success bool,
) error {
	hooks := m.definition.Hooks()
	var actionsToRun []gonfa.Action

	if success {
		actionsToRun = hooks.OnSuccess
	} else {
		actionsToRun = hooks.OnFailure
	}

	for _, action := range actionsToRun {
		if err := action.Execute(ctx, m, payload); err != nil {
			return fmt.Errorf("hook execution failed: %w", err)
		}
	}

	return nil
}

// Marshal creates a serializable representation of the instance's state.
func (m *Machine) Marshal() (*gonfa.Storable, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create a deep copy of history to ensure immutability
	historyCopy := make([]gonfa.HistoryEntry, len(m.history))
	copy(historyCopy, m.history)

	return &gonfa.Storable{
		CurrentState: m.currentState,
		History:      historyCopy,
	}, nil
}

// History returns a copy of the machine's transition history.
func (m *Machine) History() []gonfa.HistoryEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	historyCopy := make([]gonfa.HistoryEntry, len(m.history))
	copy(historyCopy, m.history)
	return historyCopy
}

// IsInFinalState checks if the machine is currently in a final (accepting) state.
func (m *Machine) IsInFinalState() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.definition.IsFinalState(m.currentState)
}

// StateExtender returns the attached user-defined business object.
// Note: This method is safe to call from within actions/guards as it doesn't
// acquire additional locks (the machine is already locked during Fire).
func (m *Machine) StateExtender() gonfa.StateExtender {
	return m.stateExtender
}
