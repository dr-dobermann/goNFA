package registry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dr-dobermann/gonfa/pkg/gonfa"
)

// Test implementations
type testGuard struct {
	result bool
}

func (g *testGuard) Check(ctx context.Context, payload gonfa.Payload) bool {
	return g.result
}

type testAction struct {
	executed bool
	err      error
}

func (a *testAction) Execute(ctx context.Context, payload gonfa.Payload) error {
	a.executed = true
	return a.err
}

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

	t.Run("empty name error", func(t *testing.T) {
		err := registry.RegisterGuard("", guard)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "guard name cannot be empty")
	})

	t.Run("nil guard error", func(t *testing.T) {
		err := registry.RegisterGuard("nilGuard", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "guard cannot be nil")
	})

	t.Run("duplicate name error", func(t *testing.T) {
		err := registry.RegisterGuard("testGuard", guard)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already registered")
	})
}

func TestRegisterAction(t *testing.T) {
	registry := New()
	action := &testAction{}

	t.Run("successful registration", func(t *testing.T) {
		err := registry.RegisterAction("testAction", action)
		assert.NoError(t, err)

		retrievedAction, exists := registry.GetAction("testAction")
		assert.True(t, exists)
		assert.Equal(t, action, retrievedAction)
	})

	t.Run("empty name error", func(t *testing.T) {
		err := registry.RegisterAction("", action)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "action name cannot be empty")
	})

	t.Run("nil action error", func(t *testing.T) {
		err := registry.RegisterAction("nilAction", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "action cannot be nil")
	})

	t.Run("duplicate name error", func(t *testing.T) {
		err := registry.RegisterAction("testAction", action)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already registered")
	})
}

func TestGetGuard(t *testing.T) {
	registry := New()
	guard := &testGuard{result: true}

	// Test non-existent guard
	retrievedGuard, exists := registry.GetGuard("nonexistent")
	assert.False(t, exists)
	assert.Nil(t, retrievedGuard)

	// Register and test existing guard
	err := registry.RegisterGuard("testGuard", guard)
	require.NoError(t, err)

	retrievedGuard, exists = registry.GetGuard("testGuard")
	assert.True(t, exists)
	assert.Equal(t, guard, retrievedGuard)
}

func TestGetAction(t *testing.T) {
	registry := New()
	action := &testAction{}

	// Test non-existent action
	retrievedAction, exists := registry.GetAction("nonexistent")
	assert.False(t, exists)
	assert.Nil(t, retrievedAction)

	// Register and test existing action
	err := registry.RegisterAction("testAction", action)
	require.NoError(t, err)

	retrievedAction, exists = registry.GetAction("testAction")
	assert.True(t, exists)
	assert.Equal(t, action, retrievedAction)
}

func TestListGuards(t *testing.T) {
	registry := New()

	// Empty registry
	guards := registry.ListGuards()
	assert.Empty(t, guards)

	// Add guards
	guard1 := &testGuard{result: true}
	guard2 := &testGuard{result: false}

	err := registry.RegisterGuard("guard1", guard1)
	require.NoError(t, err)
	err = registry.RegisterGuard("guard2", guard2)
	require.NoError(t, err)

	guards = registry.ListGuards()
	assert.Len(t, guards, 2)
	assert.Contains(t, guards, "guard1")
	assert.Contains(t, guards, "guard2")
}

func TestListActions(t *testing.T) {
	registry := New()

	// Empty registry
	actions := registry.ListActions()
	assert.Empty(t, actions)

	// Add actions
	action1 := &testAction{}
	action2 := &testAction{}

	err := registry.RegisterAction("action1", action1)
	require.NoError(t, err)
	err = registry.RegisterAction("action2", action2)
	require.NoError(t, err)

	actions = registry.ListActions()
	assert.Len(t, actions, 2)
	assert.Contains(t, actions, "action1")
	assert.Contains(t, actions, "action2")
}

func TestConcurrentAccess(t *testing.T) {
	registry := New()

	// Test concurrent registration and retrieval
	done := make(chan bool, 2)

	// Goroutine 1: Register guards
	go func() {
		for i := 0; i < 100; i++ {
			guard := &testGuard{result: true}
			err := registry.RegisterGuard(
				"guard_"+string(rune('0'+i%10)),
				guard,
			)
			if err != nil && !assert.Contains(t, err.Error(),
				"already registered") {
				t.Errorf("Unexpected error: %v", err)
			}
		}
		done <- true
	}()

	// Goroutine 2: Read guards
	go func() {
		for i := 0; i < 100; i++ {
			registry.GetGuard("guard_" + string(rune('0'+i%10)))
			registry.ListGuards()
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done
}
