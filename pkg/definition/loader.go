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

// yamlDefinition represents the YAML structure for loading definitions
type yamlDefinition struct {
	InitialState string                     `yaml:"initialState"`
	FinalStates  []string                   `yaml:"finalStates,omitempty"`
	Hooks        yamlHooks                  `yaml:"hooks,omitempty"`
	States       map[string]yamlStateConfig `yaml:"states,omitempty"`
	Transitions  []yamlTransition           `yaml:"transitions"`
}

// yamlHooks represents hooks configuration in YAML format
type yamlHooks struct {
	OnSuccess []string `yaml:"onSuccess,omitempty"`
	OnFailure []string `yaml:"onFailure,omitempty"`
}

// yamlStateConfig represents state configuration in YAML format
type yamlStateConfig struct {
	OnEntry []string `yaml:"onEntry,omitempty"`
	OnExit  []string `yaml:"onExit,omitempty"`
}

// yamlTransition represents a transition configuration in YAML format
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

	// Convert final states
	var finalStates []gonfa.State
	for _, stateName := range yamlDef.FinalStates {
		finalStates = append(finalStates, gonfa.State(stateName))
	}

	// Create and return the definition
	return New(
		gonfa.State(yamlDef.InitialState),
		finalStates,
		states,
		transitions,
		hooks,
	)
}
