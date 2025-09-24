package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	registry := New()

	assert.NotNil(t, registry)
	assert.NotNil(t, registry.guards)
	assert.NotNil(t, registry.actions)
	assert.Empty(t, registry.ListGuards())
	assert.Empty(t, registry.ListActions())
}

func TestRegisterGuard(t *testing.T) {
	registry := New()
	guard := &testGuard{result: true}

	t.Run("successful registration", func(t *testing.T) {
		err := registry.RegisterGuard("testGuard", guard)
		assert.NoError(t, err)

		retrievedGuard, exists := registry.GetGuard("testGuard")
		assert.True(t, exists)
		assert.Equal(t, guard, retrievedGuard)
	})

	t.Run("duplicate registration", func(t *testing.T) {
		guard2 := &testGuard{result: false}
		err := registry.RegisterGuard("testGuard", guard2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already registered")
	})

	t.Run("empty name", func(t *testing.T) {
		err := registry.RegisterGuard("", guard)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name cannot be empty")
	})
}

func TestRegisterAction(t *testing.T) {
	registry := New()
	action := &testAction{executed: false}

	t.Run("successful registration", func(t *testing.T) {
		err := registry.RegisterAction("testAction", action)
		assert.NoError(t, err)

		retrievedAction, exists := registry.GetAction("testAction")
		assert.True(t, exists)
		assert.Equal(t, action, retrievedAction)
	})

	t.Run("duplicate registration", func(t *testing.T) {
		action2 := &testAction{executed: true}
		err := registry.RegisterAction("testAction", action2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already registered")
	})

	t.Run("empty name", func(t *testing.T) {
		err := registry.RegisterAction("", action)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name cannot be empty")
	})
}

func TestGetGuard(t *testing.T) {
	registry := New()
	guard := &testGuard{result: true}

	// Test non-existent guard
	_, exists := registry.GetGuard("nonExistent")
	assert.False(t, exists)

	// Register and test
	err := registry.RegisterGuard("testGuard", guard)
	require.NoError(t, err)

	retrievedGuard, exists := registry.GetGuard("testGuard")
	assert.True(t, exists)
	assert.Equal(t, guard, retrievedGuard)
}

func TestGetAction(t *testing.T) {
	registry := New()
	action := &testAction{executed: false}

	// Test non-existent action
	_, exists := registry.GetAction("nonExistent")
	assert.False(t, exists)

	// Register and test
	err := registry.RegisterAction("testAction", action)
	require.NoError(t, err)

	retrievedAction, exists := registry.GetAction("testAction")
	assert.True(t, exists)
	assert.Equal(t, action, retrievedAction)
}
