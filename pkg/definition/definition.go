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
	"io"

	"gopkg.in/yaml.v3"

	"github.com/dr-dobermann/gonfa/pkg/gonfa"
	"github.com/dr-dobermann/gonfa/pkg/registry"
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

// NewDefinition creates a new Definition with the given parameters.
func NewDefinition(
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

// NewMachine creates a new Machine instance from this Definition.
// This method will be implemented by importing the machine package.
func (d *Definition) NewMachine() interface{} {
	// This is a placeholder - actual implementation will be in machine package
	panic("NewMachine should be called from machine package")
}

// RestoreMachine restores a Machine instance from a Storable state.
// This method will be implemented by importing the machine package.
func (d *Definition) RestoreMachine(state *gonfa.Storable) (interface{}, error) {
	// This is a placeholder - actual implementation will be in machine package
	panic("RestoreMachine should be called from machine package")
}

// yamlDefinition represents the YAML structure for loading definitions
type yamlDefinition struct {
	InitialState string                     `yaml:"initialState"`
	Hooks        yamlHooks                  `yaml:"hooks,omitempty"`
	States       map[string]yamlStateConfig `yaml:"states,omitempty"`
	Transitions  []yamlTransition           `yaml:"transitions"`
}

type yamlHooks struct {
	OnSuccess []string `yaml:"onSuccess,omitempty"`
	OnFailure []string `yaml:"onFailure,omitempty"`
}

type yamlStateConfig struct {
	OnEntry []string `yaml:"onEntry,omitempty"`
	OnExit  []string `yaml:"onExit,omitempty"`
}

type yamlTransition struct {
	From    string   `yaml:"from"`
	To      string   `yaml:"to"`
	On      string   `yaml:"on"`
	Guards  []string `yaml:"guards,omitempty"`
	Actions []string `yaml:"actions,omitempty"`
}

// LoadDefinition loads a definition from an io.Reader using a registry.
// The format is expected to be YAML as described in the specification.
func LoadDefinition(
	r io.Reader,
	registry *registry.Registry,
) (*Definition, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML data: %w", err)
	}

	var yamlDef yamlDefinition
	if err := yaml.Unmarshal(data, &yamlDef); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate required fields
	if yamlDef.InitialState == "" {
		return nil, fmt.Errorf("initialState is required")
	}
	if len(yamlDef.Transitions) == 0 {
		return nil, fmt.Errorf("at least one transition is required")
	}

	// Convert YAML structure to internal types
	states := make(map[gonfa.State]StateConfig)
	for stateName, stateConfig := range yamlDef.States {
		config := StateConfig{}

		// Convert OnEntry actions
		for _, actionName := range stateConfig.OnEntry {
			action, exists := registry.GetAction(actionName)
			if !exists {
				return nil, fmt.Errorf(
					"action '%s' not found in registry", actionName)
			}
			config.OnEntry = append(config.OnEntry, action)
		}

		// Convert OnExit actions
		for _, actionName := range stateConfig.OnExit {
			action, exists := registry.GetAction(actionName)
			if !exists {
				return nil, fmt.Errorf(
					"action '%s' not found in registry", actionName)
			}
			config.OnExit = append(config.OnExit, action)
		}

		states[gonfa.State(stateName)] = config
	}

	// Convert transitions
	var transitions []Transition
	for _, yamlTrans := range yamlDef.Transitions {
		transition := Transition{
			From: gonfa.State(yamlTrans.From),
			To:   gonfa.State(yamlTrans.To),
			On:   gonfa.Event(yamlTrans.On),
		}

		// Convert guards
		for _, guardName := range yamlTrans.Guards {
			guard, exists := registry.GetGuard(guardName)
			if !exists {
				return nil, fmt.Errorf(
					"guard '%s' not found in registry", guardName)
			}
			transition.Guards = append(transition.Guards, guard)
		}

		// Convert actions
		for _, actionName := range yamlTrans.Actions {
			action, exists := registry.GetAction(actionName)
			if !exists {
				return nil, fmt.Errorf(
					"action '%s' not found in registry", actionName)
			}
			transition.Actions = append(transition.Actions, action)
		}

		transitions = append(transitions, transition)
	}

	// Convert hooks
	hooks := Hooks{}
	for _, actionName := range yamlDef.Hooks.OnSuccess {
		action, exists := registry.GetAction(actionName)
		if !exists {
			return nil, fmt.Errorf(
				"success hook action '%s' not found in registry", actionName)
		}
		hooks.OnSuccess = append(hooks.OnSuccess, action)
	}

	for _, actionName := range yamlDef.Hooks.OnFailure {
		action, exists := registry.GetAction(actionName)
		if !exists {
			return nil, fmt.Errorf(
				"failure hook action '%s' not found in registry", actionName)
		}
		hooks.OnFailure = append(hooks.OnFailure, action)
	}

	// Create and return the definition
	return NewDefinition(
		gonfa.State(yamlDef.InitialState),
		states,
		transitions,
		hooks,
	)
}
