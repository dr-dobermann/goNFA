package builder

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dr-dobermann/gonfa/pkg/gonfa"
)

func TestNew(t *testing.T) {
	builder := New()

	assert.NotNil(t, builder)
	assert.NotNil(t, builder.states)
	assert.Empty(t, builder.transitions)
	assert.Empty(t, builder.initialState)
}

func TestInitialState(t *testing.T) {
	builder := New()
	state := gonfa.State("TestState")

	result := builder.InitialState(state)

	assert.Equal(t, builder, result) // Fluent interface
	assert.Equal(t, state, builder.initialState)
}

func TestFinalStates(t *testing.T) {
	builder := New()
	states := []gonfa.State{"Final1", "Final2", "Final3"}

	result := builder.FinalStates(states...)

	assert.Equal(t, builder, result) // Fluent interface
	assert.Equal(t, states, builder.finalStates)
}

func TestFinalStatesMultipleCalls(t *testing.T) {
	builder := New()
	states1 := []gonfa.State{"Final1", "Final2"}
	states2 := []gonfa.State{"Final3"}

	builder.FinalStates(states1...)
	result := builder.FinalStates(states2...)

	assert.Equal(t, builder, result) // Fluent interface
	expected := []gonfa.State{"Final1", "Final2", "Final3"}
	assert.Equal(t, expected, builder.finalStates)
}

func TestOnEntry(t *testing.T) {
	builder := New()
	state := gonfa.State("TestState")
	action1 := &testAction{name: "action1"}
	action2 := &testAction{name: "action2"}

	result := builder.OnEntry(state, action1, action2)

	assert.Equal(t, builder, result) // Fluent interface
	config := builder.states[state]
	assert.Len(t, config.OnEntry, 2)
	assert.Contains(t, config.OnEntry, action1)
	assert.Contains(t, config.OnEntry, action2)
}

func TestOnExit(t *testing.T) {
	builder := New()
	state := gonfa.State("TestState")
	action1 := &testAction{name: "action1"}
	action2 := &testAction{name: "action2"}

	result := builder.OnExit(state, action1, action2)

	assert.Equal(t, builder, result) // Fluent interface
	config := builder.states[state]
	assert.Len(t, config.OnExit, 2)
	assert.Contains(t, config.OnExit, action1)
	assert.Contains(t, config.OnExit, action2)
}

func TestOnEntryMultipleCalls(t *testing.T) {
	builder := New()
	state := gonfa.State("TestState")
	action1 := &testAction{name: "action1"}
	action2 := &testAction{name: "action2"}

	builder.OnEntry(state, action1)
	result := builder.OnEntry(state, action2)

	assert.Equal(t, builder, result) // Fluent interface
	config := builder.states[state]
	assert.Len(t, config.OnEntry, 2)
	assert.Contains(t, config.OnEntry, action1)
	assert.Contains(t, config.OnEntry, action2)
}

func TestOnExitMultipleCalls(t *testing.T) {
	builder := New()
	state := gonfa.State("TestState")
	action1 := &testAction{name: "action1"}
	action2 := &testAction{name: "action2"}

	builder.OnExit(state, action1)
	result := builder.OnExit(state, action2)

	assert.Equal(t, builder, result) // Fluent interface
	config := builder.states[state]
	assert.Len(t, config.OnExit, 2)
	assert.Contains(t, config.OnExit, action1)
	assert.Contains(t, config.OnExit, action2)
}
