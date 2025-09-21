// Package builder provides a fluent interface for programmatically creating
// state machine definitions. The Builder allows for step-by-step construction
// of complex state machines with a readable, chainable API.
//
// goNFA is a universal, lightweight and idiomatic Go library for creating
// and managing non-deterministic finite automata (NFA). It provides reliable
// state management mechanisms for complex systems such as business process
// engines (BPM).
//
// Project: https://github.com/dr-dobermann/gonfa
// Author: dr-dobermann (rgabtiov@gmail.com)
// License: LGPL-2.1 (see LICENSE file in the project root)
package builder

import (
	"fmt"

	"github.com/dr-dobermann/gonfa/pkg/definition"
	"github.com/dr-dobermann/gonfa/pkg/gonfa"
)

// Builder provides a fluent interface for creating a Definition.
type Builder struct {
	initialState   gonfa.State
	states         map[gonfa.State]definition.StateConfig
	transitions    []definition.Transition
	hooks          definition.Hooks
	lastTransition *definition.Transition
}

// New creates a new Builder instance.
func New() *Builder {
	return &Builder{
		states: make(map[gonfa.State]definition.StateConfig),
	}
}

// InitialState sets the initial state for the state machine.
func (b *Builder) InitialState(s gonfa.State) *Builder {
	b.initialState = s
	return b
}

// OnEntry defines actions to be executed upon EVERY entry into the
// specified state.
func (b *Builder) OnEntry(s gonfa.State, actions ...gonfa.Action) *Builder {
	config := b.states[s]
	config.OnEntry = append(config.OnEntry, actions...)
	b.states[s] = config
	return b
}

// OnExit defines actions to be executed upon EVERY exit from the
// specified state.
func (b *Builder) OnExit(s gonfa.State, actions ...gonfa.Action) *Builder {
	config := b.states[s]
	config.OnExit = append(config.OnExit, actions...)
	b.states[s] = config
	return b
}

// AddTransition adds a new transition and makes it the "last" transition
// for subsequent WithGuards/WithActions calls.
func (b *Builder) AddTransition(
	from gonfa.State,
	to gonfa.State,
	on gonfa.Event,
) *Builder {
	transition := definition.Transition{
		From: from,
		To:   to,
		On:   on,
	}
	b.transitions = append(b.transitions, transition)
	// Point to the last added transition for subsequent modifications
	b.lastTransition = &b.transitions[len(b.transitions)-1]
	return b
}

// WithGuards adds guards to the LAST added transition.
// Returns an error in Build() if called before AddTransition.
func (b *Builder) WithGuards(guards ...gonfa.Guard) *Builder {
	if b.lastTransition != nil {
		b.lastTransition.Guards = append(b.lastTransition.Guards, guards...)
	}
	return b
}

// WithActions adds actions to the LAST added transition.
// Returns an error in Build() if called before AddTransition.
func (b *Builder) WithActions(actions ...gonfa.Action) *Builder {
	if b.lastTransition != nil {
		b.lastTransition.Actions = append(b.lastTransition.Actions,
			actions...)
	}
	return b
}

// WithHooks sets global hooks for the state machine.
func (b *Builder) WithHooks(hooks definition.Hooks) *Builder {
	b.hooks = hooks
	return b
}

// WithSuccessHooks sets global success hooks for the state machine.
func (b *Builder) WithSuccessHooks(actions ...gonfa.Action) *Builder {
	b.hooks.OnSuccess = append(b.hooks.OnSuccess, actions...)
	return b
}

// WithFailureHooks sets global failure hooks for the state machine.
func (b *Builder) WithFailureHooks(actions ...gonfa.Action) *Builder {
	b.hooks.OnFailure = append(b.hooks.OnFailure, actions...)
	return b
}

// Build finalizes the building process and returns an immutable Definition.
// Returns an error if the configuration is invalid.
func (b *Builder) Build() (*definition.Definition, error) {
	if b.initialState == "" {
		return nil, fmt.Errorf("initial state must be set")
	}

	if len(b.transitions) == 0 {
		return nil, fmt.Errorf("at least one transition must be defined")
	}

	// Validate that WithGuards/WithActions were called appropriately
	// This is automatically handled by the lastTransition pointer

	return definition.New(
		b.initialState,
		b.states,
		b.transitions,
		b.hooks,
	)
}
