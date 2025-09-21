package builder

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dr-dobermann/gonfa/pkg/definition"
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
	name     string
	executed bool
}

func (a *testAction) Execute(ctx context.Context, payload gonfa.Payload) error {
	a.executed = true
	return nil
}

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

func TestAddTransition(t *testing.T) {
	builder := New()
	from := gonfa.State("From")
	to := gonfa.State("To")
	event := gonfa.Event("Event")

	result := builder.AddTransition(from, to, event)

	assert.Equal(t, builder, result) // Fluent interface
	require.Len(t, builder.transitions, 1)

	transition := builder.transitions[0]
	assert.Equal(t, from, transition.From)
	assert.Equal(t, to, transition.To)
	assert.Equal(t, event, transition.On)
	assert.NotNil(t, builder.lastTransition)
	assert.Equal(t, &builder.transitions[0], builder.lastTransition)
}

func TestWithGuards(t *testing.T) {
	builder := New()
	guard1 := &testGuard{result: true}
	guard2 := &testGuard{result: false}

	t.Run("with previous transition", func(t *testing.T) {
		result := builder.
			AddTransition("From", "To", "Event").
			WithGuards(guard1, guard2)

		assert.Equal(t, builder, result) // Fluent interface
		require.Len(t, builder.transitions, 1)

		transition := builder.transitions[0]
		assert.Len(t, transition.Guards, 2)
		assert.Contains(t, transition.Guards, guard1)
		assert.Contains(t, transition.Guards, guard2)
	})

	t.Run("without previous transition", func(t *testing.T) {
		newBuilder := New()
		result := newBuilder.WithGuards(guard1)

		assert.Equal(t, newBuilder, result) // Should not panic
		assert.Empty(t, newBuilder.transitions)
	})
}

func TestWithActions(t *testing.T) {
	builder := New()
	action1 := &testAction{name: "action1"}
	action2 := &testAction{name: "action2"}

	t.Run("with previous transition", func(t *testing.T) {
		result := builder.
			AddTransition("From", "To", "Event").
			WithActions(action1, action2)

		assert.Equal(t, builder, result) // Fluent interface
		require.Len(t, builder.transitions, 1)

		transition := builder.transitions[0]
		assert.Len(t, transition.Actions, 2)
		assert.Contains(t, transition.Actions, action1)
		assert.Contains(t, transition.Actions, action2)
	})

	t.Run("without previous transition", func(t *testing.T) {
		newBuilder := New()
		result := newBuilder.WithActions(action1)

		assert.Equal(t, newBuilder, result) // Should not panic
		assert.Empty(t, newBuilder.transitions)
	})
}

func TestWithHooks(t *testing.T) {
	builder := New()
	successAction := &testAction{name: "success"}
	failureAction := &testAction{name: "failure"}

	hooks := definition.Hooks{
		OnSuccess: []gonfa.Action{successAction},
		OnFailure: []gonfa.Action{failureAction},
	}

	result := builder.WithHooks(hooks)

	assert.Equal(t, builder, result) // Fluent interface
	assert.Equal(t, hooks, builder.hooks)
}

func TestWithSuccessHooks(t *testing.T) {
	builder := New()
	action1 := &testAction{name: "success1"}
	action2 := &testAction{name: "success2"}

	result := builder.WithSuccessHooks(action1, action2)

	assert.Equal(t, builder, result) // Fluent interface
	assert.Len(t, builder.hooks.OnSuccess, 2)
	assert.Contains(t, builder.hooks.OnSuccess, action1)
	assert.Contains(t, builder.hooks.OnSuccess, action2)
}

func TestWithFailureHooks(t *testing.T) {
	builder := New()
	action1 := &testAction{name: "failure1"}
	action2 := &testAction{name: "failure2"}

	result := builder.WithFailureHooks(action1, action2)

	assert.Equal(t, builder, result) // Fluent interface
	assert.Len(t, builder.hooks.OnFailure, 2)
	assert.Contains(t, builder.hooks.OnFailure, action1)
	assert.Contains(t, builder.hooks.OnFailure, action2)
}

func TestBuild(t *testing.T) {
	t.Run("successful build", func(t *testing.T) {
		action := &testAction{name: "test"}
		guard := &testGuard{result: true}

		definition, err := New().
			InitialState("Start").
			OnEntry("Start", action).
			AddTransition("Start", "End", "Event").
			WithGuards(guard).
			WithActions(action).
			WithSuccessHooks(action).
			Build()

		assert.NoError(t, err)
		assert.NotNil(t, definition)
		assert.Equal(t, gonfa.State("Start"), definition.InitialState())

		transitions := definition.Transitions()
		assert.Len(t, transitions, 1)
		assert.Equal(t, gonfa.State("Start"), transitions[0].From)
		assert.Equal(t, gonfa.State("End"), transitions[0].To)
		assert.Equal(t, gonfa.Event("Event"), transitions[0].On)
	})

	t.Run("missing initial state", func(t *testing.T) {
		definition, err := New().
			AddTransition("Start", "End", "Event").
			Build()

		assert.Error(t, err)
		assert.Nil(t, definition)
		assert.Contains(t, err.Error(), "initial state must be set")
	})

	t.Run("no transitions", func(t *testing.T) {
		definition, err := New().
			InitialState("Start").
			Build()

		assert.Error(t, err)
		assert.Nil(t, definition)
		assert.Contains(t, err.Error(), "at least one transition must be defined")
	})
}

func TestFluentInterface(t *testing.T) {
	// Test complete fluent chain
	action := &testAction{name: "test"}
	guard := &testGuard{result: true}

	definition, err := New().
		InitialState("Draft").
		OnEntry("Draft", action).
		OnExit("Draft", action).
		AddTransition("Draft", "InReview", "Submit").
		WithGuards(guard).
		WithActions(action).
		AddTransition("InReview", "Approved", "Approve").
		WithGuards(guard).
		WithSuccessHooks(action).
		WithFailureHooks(action).
		Build()

	assert.NoError(t, err)
	assert.NotNil(t, definition)

	// Verify the definition
	assert.Equal(t, gonfa.State("Draft"), definition.InitialState())

	transitions := definition.Transitions()
	assert.Len(t, transitions, 2)

	// Check first transition
	assert.Equal(t, gonfa.State("Draft"), transitions[0].From)
	assert.Equal(t, gonfa.State("InReview"), transitions[0].To)
	assert.Equal(t, gonfa.Event("Submit"), transitions[0].On)
	assert.Len(t, transitions[0].Guards, 1)
	assert.Len(t, transitions[0].Actions, 1)

	// Check second transition
	assert.Equal(t, gonfa.State("InReview"), transitions[1].From)
	assert.Equal(t, gonfa.State("Approved"), transitions[1].To)
	assert.Equal(t, gonfa.Event("Approve"), transitions[1].On)
	assert.Len(t, transitions[1].Guards, 1)

	// Check hooks
	hooks := definition.Hooks()
	assert.Len(t, hooks.OnSuccess, 1)
	assert.Len(t, hooks.OnFailure, 1)
}
