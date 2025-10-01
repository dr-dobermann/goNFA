package builder

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dr-dobermann/gonfa/pkg/gonfa"
)

func TestBuildSuccess(t *testing.T) {
	action1 := &testAction{name: "action1"}
	action2 := &testAction{name: "action2"}
	guard1 := &testGuard{result: true}
	guard2 := &testGuard{result: false}

	def, err := New().
		InitialState("Start").
		FinalStates("End").
		OnEntry("Start", action1).
		OnExit("Start", action2).
		AddTransition("Start", "End", "ToEnd").
		WithGuards(guard1, guard2).
		WithActions(action1, action2).
		Build()

	require.NoError(t, err)
	assert.NotNil(t, def)
	assert.Equal(t, gonfa.State("Start"), def.InitialState())
	assert.True(t, def.IsFinalState("End"))
}

func TestBuildNoInitialState(t *testing.T) {
	_, err := New().
		AddTransition("Start", "End", "ToEnd").
		Build()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "initial state must be set")
}

func TestBuildNoTransitions(t *testing.T) {
	_, err := New().
		InitialState("Start").
		Build()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one transition must be defined")
}

func TestBuildWithFinalStates(t *testing.T) {
	def, err := New().
		InitialState("Start").
		FinalStates("End1", "End2").
		AddTransition("Start", "End1", "ToEnd1").
		AddTransition("Start", "End2", "ToEnd2").
		Build()

	require.NoError(t, err)
	assert.True(t, def.IsFinalState("End1"))
	assert.True(t, def.IsFinalState("End2"))
	assert.False(t, def.IsFinalState("Start"))
}

func TestBuildWithStateActions(t *testing.T) {
	entryAction := &testAction{name: "entry"}
	exitAction := &testAction{name: "exit"}

	def, err := New().
		InitialState("Start").
		FinalStates("End").
		OnEntry("End", entryAction).
		OnExit("Start", exitAction).
		AddTransition("Start", "End", "ToEnd").
		Build()

	require.NoError(t, err)

	startConfig := def.GetStateConfig("Start")
	endConfig := def.GetStateConfig("End")

	assert.Len(t, startConfig.OnExit, 1)
	assert.Len(t, endConfig.OnEntry, 1)
	assert.Contains(t, startConfig.OnExit, exitAction)
	assert.Contains(t, endConfig.OnEntry, entryAction)
}

func TestBuildWithHooks(t *testing.T) {
	successAction := &testAction{name: "success"}
	failureAction := &testAction{name: "failure"}

	def, err := New().
		InitialState("Start").
		FinalStates("End").
		WithSuccessHooks(successAction).
		WithFailureHooks(failureAction).
		AddTransition("Start", "End", "ToEnd").
		Build()

	require.NoError(t, err)

	hooks := def.Hooks()
	assert.Len(t, hooks.OnSuccess, 1)
	assert.Len(t, hooks.OnFailure, 1)
	assert.Contains(t, hooks.OnSuccess, successAction)
	assert.Contains(t, hooks.OnFailure, failureAction)
}
