// Package registry provides a mapping from string names to Guard and Action
// objects. This allows for decoupling of declarative definitions (from files)
// and actual implementation code.
//
// goNFA is a universal, lightweight and idiomatic Go library for creating
// and managing non-deterministic finite automata (NFA). It provides reliable
// state management mechanisms for complex systems such as business process
// engines (BPM).
//
// Project: https://github.com/dr-dobermann/gonfa
// Author: dr-dobermann (rgabtiov@gmail.com)
// License: LGPL-2.1 (see LICENSE file in the project root)
package registry

import (
	"fmt"
	"sync"

	"github.com/dr-dobermann/gonfa/pkg/gonfa"
)

// Registry stores a mapping from string names to real objects.
// It provides thread-safe registration and retrieval of Guard and Action
// implementations.
type Registry struct {
	mu      sync.RWMutex
	guards  map[string]gonfa.Guard
	actions map[string]gonfa.Action
}

// New creates a new Registry instance.
func New() *Registry {
	return &Registry{
		guards:  make(map[string]gonfa.Guard),
		actions: make(map[string]gonfa.Action),
	}
}

// RegisterGuard registers a guard object under a unique name.
// Returns an error if the name is already registered.
func (r *Registry) RegisterGuard(name string, guard gonfa.Guard) error {
	if name == "" {
		return fmt.Errorf("guard name cannot be empty")
	}
	if guard == nil {
		return fmt.Errorf("guard cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.guards[name]; exists {
		return fmt.Errorf("guard with name '%s' is already registered", name)
	}

	r.guards[name] = guard
	return nil
}

// RegisterAction registers an action (or hook) object under a unique name.
// Returns an error if the name is already registered.
func (r *Registry) RegisterAction(
	name string,
	action gonfa.Action,
) error {
	if name == "" {
		return fmt.Errorf("action name cannot be empty")
	}
	if action == nil {
		return fmt.Errorf("action cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.actions[name]; exists {
		return fmt.Errorf("action with name '%s' is already registered",
			name)
	}

	r.actions[name] = action
	return nil
}

// GetGuard retrieves a guard by name.
// Returns the guard and true if found, nil and false otherwise.
func (r *Registry) GetGuard(name string) (gonfa.Guard, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	guard, exists := r.guards[name]
	return guard, exists
}

// GetAction retrieves an action by name.
// Returns the action and true if found, nil and false otherwise.
func (r *Registry) GetAction(name string) (gonfa.Action, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	action, exists := r.actions[name]
	return action, exists
}

// ListGuards returns a slice of all registered guard names.
func (r *Registry) ListGuards() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.guards))
	for name := range r.guards {
		names = append(names, name)
	}
	return names
}

// ListActions returns a slice of all registered action names.
func (r *Registry) ListActions() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.actions))
	for name := range r.actions {
		names = append(names, name)
	}
	return names
}
