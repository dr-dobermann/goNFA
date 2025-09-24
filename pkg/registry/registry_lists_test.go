package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListGuards(t *testing.T) {
	registry := New()

	// Initially empty
	guards := registry.ListGuards()
	assert.Empty(t, guards)

	// Register guards
	guard1 := &testGuard{result: true}
	guard2 := &testGuard{result: false}

	err := registry.RegisterGuard("guard1", guard1)
	require.NoError(t, err)

	err = registry.RegisterGuard("guard2", guard2)
	require.NoError(t, err)

	// Test listing
	guards = registry.ListGuards()
	assert.Len(t, guards, 2)
	assert.Contains(t, guards, "guard1")
	assert.Contains(t, guards, "guard2")
}

func TestListActions(t *testing.T) {
	registry := New()

	// Initially empty
	actions := registry.ListActions()
	assert.Empty(t, actions)

	// Register actions
	action1 := &testAction{executed: false}
	action2 := &testAction{executed: true}

	err := registry.RegisterAction("action1", action1)
	require.NoError(t, err)

	err = registry.RegisterAction("action2", action2)
	require.NoError(t, err)

	// Test listing
	actions = registry.ListActions()
	assert.Len(t, actions, 2)
	assert.Contains(t, actions, "action1")
	assert.Contains(t, actions, "action2")
}

func TestMultipleRegistrations(t *testing.T) {
	registry := New()

	// Register multiple guards and actions
	guard1 := &testGuard{result: true}
	guard2 := &testGuard{result: false}
	action1 := &testAction{executed: false}
	action2 := &testAction{executed: true}

	err := registry.RegisterGuard("guard1", guard1)
	require.NoError(t, err)

	err = registry.RegisterGuard("guard2", guard2)
	require.NoError(t, err)

	err = registry.RegisterAction("action1", action1)
	require.NoError(t, err)

	err = registry.RegisterAction("action2", action2)
	require.NoError(t, err)

	// Test all registrations
	guards := registry.ListGuards()
	actions := registry.ListActions()

	assert.Len(t, guards, 2)
	assert.Len(t, actions, 2)

	// Test individual retrievals
	retrievedGuard1, exists := registry.GetGuard("guard1")
	assert.True(t, exists)
	assert.Equal(t, guard1, retrievedGuard1)

	retrievedGuard2, exists := registry.GetGuard("guard2")
	assert.True(t, exists)
	assert.Equal(t, guard2, retrievedGuard2)

	retrievedAction1, exists := registry.GetAction("action1")
	assert.True(t, exists)
	assert.Equal(t, action1, retrievedAction1)

	retrievedAction2, exists := registry.GetAction("action2")
	assert.True(t, exists)
	assert.Equal(t, action2, retrievedAction2)
}
