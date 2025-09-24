package definition

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dr-dobermann/gonfa/pkg/gonfa"
)

func TestNewDefinition(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		initialState := gonfa.State("Start")
		finalStates := []gonfa.State{"End1", "End2"}
		states := map[gonfa.State]StateConfig{
			"Start": {OnEntry: []gonfa.Action{&testAction{name: "startEntry"}}},
			"End1":  {OnExit: []gonfa.Action{&testAction{name: "end1Exit"}}},
		}
		transitions := []Transition{
			{From: "Start", To: "End1", On: "Event1"},
		}
		hooks := Hooks{
			OnSuccess: []gonfa.Action{&testAction{name: "success"}},
		}

		def, err := New(initialState, finalStates, states, transitions, hooks)
		require.NoError(t, err)
		assert.NotNil(t, def)
		assert.Equal(t, initialState, def.InitialState())
		assert.True(t, def.IsFinalState("End1"))
		assert.True(t, def.IsFinalState("End2"))
		assert.False(t, def.IsFinalState("Start"))
	})

	t.Run("empty initial state", func(t *testing.T) {
		_, err := New("", nil, nil, nil, Hooks{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "initial state cannot be empty")
	})

	t.Run("initial state not found in states or transitions", func(t *testing.T) {
		_, err := New("NonExistent", nil, nil, nil, Hooks{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "initial state 'NonExistent' not found")
	})

	t.Run("initial state found in transitions", func(t *testing.T) {
		transitions := []Transition{
			{From: "Start", To: "End", On: "Event1"},
		}

		def, err := New("Start", nil, nil, transitions, Hooks{})
		require.NoError(t, err)
		assert.Equal(t, gonfa.State("Start"), def.InitialState())
	})

	t.Run("initial state found in states", func(t *testing.T) {
		states := map[gonfa.State]StateConfig{
			"Start": {},
		}

		def, err := New("Start", nil, states, nil, Hooks{})
		require.NoError(t, err)
		assert.Equal(t, gonfa.State("Start"), def.InitialState())
	})
}

func TestDefinitionGetters(t *testing.T) {
	initialState := gonfa.State("Start")
	finalStates := []gonfa.State{"End"}
	states := map[gonfa.State]StateConfig{
		"Start": {OnEntry: []gonfa.Action{&testAction{name: "startEntry"}}},
		"End":   {OnExit: []gonfa.Action{&testAction{name: "endExit"}}},
	}
	transitions := []Transition{
		{From: "Start", To: "End", On: "Event1"},
	}
	hooks := Hooks{
		OnSuccess: []gonfa.Action{&testAction{name: "success"}},
		OnFailure: []gonfa.Action{&testAction{name: "failure"}},
	}

	def, err := New(initialState, finalStates, states, transitions, hooks)
	require.NoError(t, err)

	t.Run("InitialState", func(t *testing.T) {
		assert.Equal(t, initialState, def.InitialState())
	})

	t.Run("FinalStates", func(t *testing.T) {
		finalStatesMap := def.FinalStates()
		assert.True(t, finalStatesMap["End"])
		assert.False(t, finalStatesMap["Start"])
	})

	t.Run("IsFinalState", func(t *testing.T) {
		assert.True(t, def.IsFinalState("End"))
		assert.False(t, def.IsFinalState("Start"))
		assert.False(t, def.IsFinalState("NonExistent"))
	})

	t.Run("States", func(t *testing.T) {
		statesMap := def.States()
		assert.Len(t, statesMap, 2)
		assert.Contains(t, statesMap, gonfa.State("Start"))
		assert.Contains(t, statesMap, gonfa.State("End"))
	})

	t.Run("Transitions", func(t *testing.T) {
		transitionsList := def.Transitions()
		assert.Len(t, transitionsList, 1)
		assert.Equal(t, "Start", string(transitionsList[0].From))
		assert.Equal(t, "End", string(transitionsList[0].To))
		assert.Equal(t, "Event1", string(transitionsList[0].On))
	})

	t.Run("Hooks", func(t *testing.T) {
		hooksResult := def.Hooks()
		assert.Len(t, hooksResult.OnSuccess, 1)
		assert.Len(t, hooksResult.OnFailure, 1)
	})
}

func TestGetTransitions(t *testing.T) {
	transitions := []Transition{
		{From: "Start", To: "Middle", On: "Event1"},
		{From: "Start", To: "End", On: "Event1"},
		{From: "Middle", To: "End", On: "Event2"},
	}

	def, err := New("Start", nil, nil, transitions, Hooks{})
	require.NoError(t, err)

	t.Run("multiple transitions for same state and event", func(t *testing.T) {
		result := def.GetTransitions("Start", "Event1")
		assert.Len(t, result, 2)
		assert.Equal(t, "Middle", string(result[0].To))
		assert.Equal(t, "End", string(result[1].To))
	})

	t.Run("single transition", func(t *testing.T) {
		result := def.GetTransitions("Middle", "Event2")
		assert.Len(t, result, 1)
		assert.Equal(t, "End", string(result[0].To))
	})

	t.Run("no transitions found", func(t *testing.T) {
		result := def.GetTransitions("Start", "NonExistent")
		assert.Empty(t, result)
	})
}

func TestGetStateConfig(t *testing.T) {
	states := map[gonfa.State]StateConfig{
		"Start": {
			OnEntry: []gonfa.Action{&testAction{name: "startEntry"}},
			OnExit:  []gonfa.Action{&testAction{name: "startExit"}},
		},
	}

	def, err := New("Start", nil, states, nil, Hooks{})
	require.NoError(t, err)

	t.Run("existing state", func(t *testing.T) {
		config := def.GetStateConfig("Start")
		assert.Len(t, config.OnEntry, 1)
		assert.Len(t, config.OnExit, 1)
	})

	t.Run("non-existent state", func(t *testing.T) {
		config := def.GetStateConfig("NonExistent")
		assert.Empty(t, config.OnEntry)
		assert.Empty(t, config.OnExit)
	})
}
