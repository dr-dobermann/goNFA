package machine

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dr-dobermann/gonfa/pkg/builder"
	"github.com/dr-dobermann/gonfa/pkg/definition"
	"github.com/dr-dobermann/gonfa/pkg/gonfa"
)

func createTestDefinition(t *testing.T) *definition.Definition {
	def, err := builder.New().
		InitialState("Start").
		FinalStates("End").
		OnEntry("Middle", &testAction{name: "middleEntry"}).
		OnEntry("End", &testAction{name: "endEntry"}).
		AddTransition("Start", "Middle", "ToMiddle").
		AddTransition("Middle", "End", "ToEnd").
		Build()
	require.NoError(t, err)
	return def
}

func TestNewMachine(t *testing.T) {
	def := createTestDefinition(t)
	machine, err := New(def, nil)

	require.NoError(t, err)
	assert.NotNil(t, machine)
	assert.Equal(t, gonfa.State("Start"), machine.CurrentState())
	assert.Empty(t, machine.History())
}

func TestRestoreMachine(t *testing.T) {
	def := createTestDefinition(t)
	extender := &testStateExtender{data: "test"}

	t.Run("successful restoration", func(t *testing.T) {
		storable := &gonfa.Storable{
			CurrentState: "Middle",
			History: []gonfa.HistoryEntry{
				{
					From:      "Start",
					To:        "Middle",
					On:        "ToMiddle",
					Timestamp: time.Now(),
				},
			},
		}

		machine, err := Restore(def, storable, extender)
		assert.NoError(t, err)
		assert.NotNil(t, machine)
		assert.Equal(t, gonfa.State("Middle"), machine.CurrentState())
		assert.Len(t, machine.History(), 1)
	})

	t.Run("nil storable", func(t *testing.T) {
		machine, err := Restore(def, nil, extender)
		assert.Error(t, err)
		assert.Nil(t, machine)
		assert.Contains(t, err.Error(), "storable state cannot be nil")
	})

	t.Run("empty current state", func(t *testing.T) {
		storable := &gonfa.Storable{
			CurrentState: "",
			History:      []gonfa.HistoryEntry{},
		}

		machine, err := Restore(def, storable, extender)
		assert.Error(t, err)
		assert.Nil(t, machine)
		assert.Contains(t, err.Error(), "current state cannot be empty")
	})

	t.Run("invalid current state", func(t *testing.T) {
		storable := &gonfa.Storable{
			CurrentState: "InvalidState",
			History:      []gonfa.HistoryEntry{},
		}

		machine, err := Restore(def, storable, extender)
		assert.Error(t, err)
		assert.Nil(t, machine)
		assert.Contains(t, err.Error(), "not found in definition")
	})
}

func TestCurrentState(t *testing.T) {
	def := createTestDefinition(t)
	machine, err := New(def, nil)
	require.NoError(t, err)

	assert.Equal(t, gonfa.State("Start"), machine.CurrentState())
}

func TestHistory(t *testing.T) {
	def := createTestDefinition(t)
	machine, err := New(def, nil)
	require.NoError(t, err)

	// Initially empty
	history := machine.History()
	assert.Empty(t, history)

	// After a transition
	success, err := machine.Fire(context.Background(), "ToMiddle", nil)
	require.NoError(t, err)
	assert.True(t, success)

	history = machine.History()
	assert.Len(t, history, 1)
	assert.Equal(t, gonfa.State("Start"), history[0].From)
	assert.Equal(t, gonfa.State("Middle"), history[0].To)
	assert.Equal(t, gonfa.Event("ToMiddle"), history[0].On)
}

func TestStateExtender(t *testing.T) {
	def := createTestDefinition(t)
	extender := &testStateExtender{data: "test data"}

	machine, err := New(def, extender)
	require.NoError(t, err)

	assert.Equal(t, extender, machine.StateExtender())
}

func TestIsInFinalState(t *testing.T) {
	def, err := builder.New().
		InitialState("Start").
		FinalStates("End").
		AddTransition("Start", "End", "ToEnd").
		Build()
	require.NoError(t, err)

	machine, err := New(def, nil)
	require.NoError(t, err)

	// Initially not in final state
	assert.False(t, machine.IsInFinalState())

	// After transition to final state
	success, err := machine.Fire(context.Background(), "ToEnd", nil)
	require.NoError(t, err)
	assert.True(t, success)
	assert.True(t, machine.IsInFinalState())
}
