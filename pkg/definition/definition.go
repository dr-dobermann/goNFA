// Package definition provides structures and functions for creating
// immutable state machine definitions. A Definition describes the static
// structure of a state machine including states, transitions, and hooks.
//
// goNFA is a universal, lightweight and idiomatic Go library for creating
// and managing non-deterministic finite automata (NFA). It provides reliable
// state management mechanisms for complex systems such as business process
// engines (BPM).
//
// Project: https://github.com/dr-dobermann/gonfa
// Author: dr-dobermann (rgabtiov@gmail.com)
// License: LGPL-2.1 (see LICENSE file in the project root)
package definition

import (
	"fmt"

	"github.com/dr-dobermann/gonfa/pkg/gonfa"
)

// Transition describes one possible transition between states.
type Transition struct {
	From    gonfa.State    // Source state
	To      gonfa.State    // Target state
	On      gonfa.Event    // Triggering event
	Guards  []gonfa.Guard  // Chain of guards that must all pass
	Actions []gonfa.Action // Chain of actions to execute during transition
}

// StateConfig describes actions associated with a specific state.
type StateConfig struct {
	OnEntry []gonfa.Action // Actions to execute upon entering the state
	OnExit  []gonfa.Action // Actions to execute upon exiting the state
}

// Hooks describes a set of global hooks for the state machine.
type Hooks struct {
	OnSuccess []gonfa.Action // Called after successful transitions
	OnFailure []gonfa.Action // Called after failed transitions
}

// Definition is an immutable description of the state machine graph.
// It contains all states, transitions, and associated actions/guards.
type Definition struct {
	initialState gonfa.State
	states       map[gonfa.State]StateConfig
	transitions  []Transition
	hooks        Hooks
}

// New creates a new Definition with the given parameters.
func New(
	initialState gonfa.State,
	states map[gonfa.State]StateConfig,
	transitions []Transition,
	hooks Hooks,
) (*Definition, error) {
	if initialState == "" {
		return nil, fmt.Errorf("initial state cannot be empty")
	}

	// Validate that initial state exists in states or transitions
	stateExists := false
	if _, exists := states[initialState]; exists {
		stateExists = true
	}

	if !stateExists {
		for _, t := range transitions {
			if t.From == initialState || t.To == initialState {
				stateExists = true
				break
			}
		}
	}

	if !stateExists {
		return nil, fmt.Errorf(
			"initial state '%s' not found in states or transitions",
			initialState)
	}

	// Copy states map to ensure immutability
	statesCopy := make(map[gonfa.State]StateConfig, len(states))
	for k, v := range states {
		statesCopy[k] = v
	}

	// Copy transitions slice
	transitionsCopy := make([]Transition, len(transitions))
	copy(transitionsCopy, transitions)

	return &Definition{
		initialState: initialState,
		states:       statesCopy,
		transitions:  transitionsCopy,
		hooks:        hooks,
	}, nil
}

// InitialState returns the initial state of the machine.
func (d *Definition) InitialState() gonfa.State {
	return d.initialState
}

// States returns a copy of the states configuration.
func (d *Definition) States() map[gonfa.State]StateConfig {
	states := make(map[gonfa.State]StateConfig, len(d.states))
	for k, v := range d.states {
		states[k] = v
	}
	return states
}

// Transitions returns a copy of all transitions.
func (d *Definition) Transitions() []Transition {
	transitions := make([]Transition, len(d.transitions))
	copy(transitions, d.transitions)
	return transitions
}

// Hooks returns the global hooks configuration.
func (d *Definition) Hooks() Hooks {
	return d.hooks
}

// GetTransitions returns all transitions that can be triggered from the given
// state with the given event.
func (d *Definition) GetTransitions(
	from gonfa.State,
	event gonfa.Event,
) []Transition {
	var result []Transition
	for _, t := range d.transitions {
		if t.From == from && t.On == event {
			result = append(result, t)
		}
	}
	return result
}

// GetStateConfig returns the configuration for the given state.
// Returns an empty StateConfig if the state is not configured.
func (d *Definition) GetStateConfig(state gonfa.State) StateConfig {
	config, exists := d.states[state]
	if !exists {
		return StateConfig{}
	}
	return config
}
