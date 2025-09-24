package machine

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dr-dobermann/gonfa/pkg/builder"
	"github.com/dr-dobermann/gonfa/pkg/gonfa"
)

func TestFireTransition(t *testing.T) {
	def := createTestDefinition(t)
	machine, err := New(def, nil)
	require.NoError(t, err)

	// Test successful transition
	success, err := machine.Fire(context.Background(), "ToMiddle", nil)
	require.NoError(t, err)
	assert.True(t, success)
	assert.Equal(t, gonfa.State("Middle"), machine.CurrentState())

	// Test second transition
	success, err = machine.Fire(context.Background(), "ToEnd", nil)
	require.NoError(t, err)
	assert.True(t, success)
	assert.Equal(t, gonfa.State("End"), machine.CurrentState())
}

func TestFireInvalidEvent(t *testing.T) {
	def := createTestDefinition(t)
	machine, err := New(def, nil)
	require.NoError(t, err)

	// Test invalid event
	success, err := machine.Fire(context.Background(), "InvalidEvent", nil)
	require.NoError(t, err)
	assert.False(t, success)
	assert.Equal(t, gonfa.State("Start"), machine.CurrentState())
}

func TestFireWithGuards(t *testing.T) {
	guard := &testGuard{result: true}
	action := &testAction{name: "testAction"}

	def, err := builder.New().
		InitialState("Start").
		AddTransition("Start", "End", "ToEnd").
		WithGuards(guard).
		WithActions(action).
		Build()
	require.NoError(t, err)

	machine, err := New(def, nil)
	require.NoError(t, err)

	// Test successful transition with guard
	success, err := machine.Fire(context.Background(), "ToEnd", nil)
	require.NoError(t, err)
	assert.True(t, success)
	assert.Equal(t, gonfa.State("End"), machine.CurrentState())
	assert.True(t, action.executed)
	assert.Equal(t, 1, guard.calls)
}

func TestFireWithFailingGuard(t *testing.T) {
	guard := &testGuard{result: false}

	def, err := builder.New().
		InitialState("Start").
		AddTransition("Start", "End", "ToEnd").
		WithGuards(guard).
		Build()
	require.NoError(t, err)

	machine, err := New(def, nil)
	require.NoError(t, err)

	// Test failed transition due to guard
	success, err := machine.Fire(context.Background(), "ToEnd", nil)
	require.NoError(t, err)
	assert.False(t, success)
	assert.Equal(t, gonfa.State("Start"), machine.CurrentState())
	assert.Equal(t, 1, guard.calls)
}

func TestFireWithFailingAction(t *testing.T) {
	action := &testAction{name: "failingAction", err: errors.New("action failed")}

	def, err := builder.New().
		InitialState("Start").
		AddTransition("Start", "End", "ToEnd").
		WithActions(action).
		Build()
	require.NoError(t, err)

	machine, err := New(def, nil)
	require.NoError(t, err)

	// Test failed transition due to action
	success, err := machine.Fire(context.Background(), "ToEnd", nil)
	assert.Error(t, err)
	assert.False(t, success)
	assert.Equal(t, gonfa.State("Start"), machine.CurrentState())
	assert.True(t, action.executed)
}

func TestFireWithStateActions(t *testing.T) {
	entryAction := &testAction{name: "entryAction"}
	exitAction := &testAction{name: "exitAction"}

	def, err := builder.New().
		InitialState("Start").
		OnEntry("End", entryAction).
		OnExit("Start", exitAction).
		AddTransition("Start", "End", "ToEnd").
		Build()
	require.NoError(t, err)

	machine, err := New(def, nil)
	require.NoError(t, err)

	// Test transition with state actions
	success, err := machine.Fire(context.Background(), "ToEnd", nil)
	require.NoError(t, err)
	assert.True(t, success)
	assert.Equal(t, gonfa.State("End"), machine.CurrentState())
	assert.True(t, entryAction.executed)
	assert.True(t, exitAction.executed)
}

func TestFireWithHooks(t *testing.T) {
	successHook := &testAction{name: "successHook"}
	failureHook := &testAction{name: "failureHook"}

	def, err := builder.New().
		InitialState("Start").
		WithSuccessHooks(successHook).
		WithFailureHooks(failureHook).
		AddTransition("Start", "End", "ToEnd").
		Build()
	require.NoError(t, err)

	machine, err := New(def, nil)
	require.NoError(t, err)

	// Test successful transition with hooks
	success, err := machine.Fire(context.Background(), "ToEnd", nil)
	require.NoError(t, err)
	assert.True(t, success)
	assert.True(t, successHook.executed)
	assert.False(t, failureHook.executed)

	// Reset hooks
	successHook.executed = false
	failureHook.executed = false

	// Test failed transition with hooks
	success, err = machine.Fire(context.Background(), "InvalidEvent", nil)
	require.NoError(t, err)
	assert.False(t, success)
	assert.False(t, successHook.executed)
	assert.True(t, failureHook.executed)
}
