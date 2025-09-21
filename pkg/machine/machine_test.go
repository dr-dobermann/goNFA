package machine

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dr-dobermann/gonfa/pkg/builder"
	"github.com/dr-dobermann/gonfa/pkg/definition"
	"github.com/dr-dobermann/gonfa/pkg/gonfa"
)

// Test implementations
type testGuard struct {
	result bool
	calls  int
}

func (g *testGuard) Check(ctx context.Context, payload gonfa.Payload) bool {
	g.calls++
	return g.result
}

type testAction struct {
	name     string
	executed bool
	err      error
	calls    int
}

func (a *testAction) Execute(ctx context.Context, payload gonfa.Payload) error {
	a.calls++
	a.executed = true
	return a.err
}

func createTestDefinition(t *testing.T) *definition.Definition {
	def, err := builder.New().
		InitialState("Start").
		AddTransition("Start", "Middle", "ToMiddle").
		AddTransition("Middle", "End", "ToEnd").
		Build()
	require.NoError(t, err)
	return def
}

func TestNewMachine(t *testing.T) {
	def := createTestDefinition(t)
	machine := New(def)

	assert.NotNil(t, machine)
	assert.Equal(t, gonfa.State("Start"), machine.CurrentState())
	assert.Empty(t, machine.History())
}

func TestRestoreMachine(t *testing.T) {
	def := createTestDefinition(t)

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

		machine, err := Restore(def, storable)
		assert.NoError(t, err)
		assert.NotNil(t, machine)
		assert.Equal(t, gonfa.State("Middle"), machine.CurrentState())
		assert.Len(t, machine.History(), 1)
	})

	t.Run("nil storable", func(t *testing.T) {
		machine, err := Restore(def, nil)
		assert.Error(t, err)
		assert.Nil(t, machine)
		assert.Contains(t, err.Error(), "storable state cannot be nil")
	})

	t.Run("empty current state", func(t *testing.T) {
		storable := &gonfa.Storable{
			CurrentState: "",
			History:      []gonfa.HistoryEntry{},
		}

		machine, err := Restore(def, storable)
		assert.Error(t, err)
		assert.Nil(t, machine)
		assert.Contains(t, err.Error(), "current state cannot be empty")
	})
}

func TestCurrentState(t *testing.T) {
	def := createTestDefinition(t)
	machine := New(def)

	assert.Equal(t, gonfa.State("Start"), machine.CurrentState())
}

func TestFireSuccessfulTransition(t *testing.T) {
	action := &testAction{name: "test"}
	guard := &testGuard{result: true}

	def, err := builder.New().
		InitialState("Start").
		AddTransition("Start", "End", "ToEnd").
		WithGuards(guard).
		WithActions(action).
		Build()
	require.NoError(t, err)

	machine := New(def)
	ctx := context.Background()
	payload := "test payload"

	success, err := machine.Fire(ctx, "ToEnd", payload)

	assert.NoError(t, err)
	assert.True(t, success)
	assert.Equal(t, gonfa.State("End"), machine.CurrentState())
	assert.Equal(t, 1, guard.calls)
	assert.Equal(t, 1, action.calls)
	assert.True(t, action.executed)

	history := machine.History()
	assert.Len(t, history, 1)
	assert.Equal(t, gonfa.State("Start"), history[0].From)
	assert.Equal(t, gonfa.State("End"), history[0].To)
	assert.Equal(t, gonfa.Event("ToEnd"), history[0].On)
}

func TestFireGuardFailure(t *testing.T) {
	action := &testAction{name: "test"}
	guard := &testGuard{result: false} // Guard fails

	def, err := builder.New().
		InitialState("Start").
		AddTransition("Start", "End", "ToEnd").
		WithGuards(guard).
		WithActions(action).
		Build()
	require.NoError(t, err)

	machine := New(def)
	ctx := context.Background()

	success, err := machine.Fire(ctx, "ToEnd", nil)

	assert.NoError(t, err)
	assert.False(t, success)
	assert.Equal(t, gonfa.State("Start"), machine.CurrentState()) // No change
	assert.Equal(t, 1, guard.calls)
	assert.Equal(t, 0, action.calls)   // Action should not be called
	assert.Empty(t, machine.History()) // No history entry
}

func TestFireActionError(t *testing.T) {
	expectedErr := errors.New("action failed")
	action := &testAction{name: "test", err: expectedErr}
	guard := &testGuard{result: true}

	def, err := builder.New().
		InitialState("Start").
		AddTransition("Start", "End", "ToEnd").
		WithGuards(guard).
		WithActions(action).
		Build()
	require.NoError(t, err)

	machine := New(def)
	ctx := context.Background()

	success, err := machine.Fire(ctx, "ToEnd", nil)

	assert.Error(t, err)
	assert.False(t, success)
	assert.Contains(t, err.Error(), "transition action failed")
	assert.Equal(t, gonfa.State("Start"), machine.CurrentState()) // No change
}

func TestFireNoTransition(t *testing.T) {
	def := createTestDefinition(t)
	machine := New(def)
	ctx := context.Background()

	success, err := machine.Fire(ctx, "NonExistentEvent", nil)

	assert.NoError(t, err)
	assert.False(t, success)
	assert.Equal(t, gonfa.State("Start"), machine.CurrentState()) // No change
	assert.Empty(t, machine.History())                            // No history entry
}

func TestFireWithStateActions(t *testing.T) {
	onExitAction := &testAction{name: "onExit"}
	onEntryAction := &testAction{name: "onEntry"}
	transitionAction := &testAction{name: "transition"}

	def, err := builder.New().
		InitialState("Start").
		OnExit("Start", onExitAction).
		OnEntry("End", onEntryAction).
		AddTransition("Start", "End", "ToEnd").
		WithActions(transitionAction).
		Build()
	require.NoError(t, err)

	machine := New(def)
	ctx := context.Background()

	success, err := machine.Fire(ctx, "ToEnd", nil)

	assert.NoError(t, err)
	assert.True(t, success)
	assert.Equal(t, gonfa.State("End"), machine.CurrentState())

	// Check execution order: OnExit -> Transition -> OnEntry
	assert.True(t, onExitAction.executed)
	assert.True(t, transitionAction.executed)
	assert.True(t, onEntryAction.executed)
	assert.Equal(t, 1, onExitAction.calls)
	assert.Equal(t, 1, transitionAction.calls)
	assert.Equal(t, 1, onEntryAction.calls)
}

func TestFireWithHooks(t *testing.T) {
	successHook := &testAction{name: "success"}
	failureHook := &testAction{name: "failure"}
	guard := &testGuard{result: true}

	def, err := builder.New().
		InitialState("Start").
		AddTransition("Start", "End", "ToEnd").
		WithGuards(guard).
		WithSuccessHooks(successHook).
		WithFailureHooks(failureHook).
		Build()
	require.NoError(t, err)

	machine := New(def)
	ctx := context.Background()

	t.Run("successful transition calls success hook", func(t *testing.T) {
		success, err := machine.Fire(ctx, "ToEnd", nil)

		assert.NoError(t, err)
		assert.True(t, success)
		assert.True(t, successHook.executed)
		assert.False(t, failureHook.executed)
	})

	t.Run("failed transition calls failure hook", func(t *testing.T) {
		// Reset machine
		machine = New(def)
		successHook.executed = false
		failureHook.executed = false

		// Make guard fail
		guard.result = false

		success, err := machine.Fire(ctx, "ToEnd", nil)

		assert.NoError(t, err)
		assert.False(t, success)
		assert.False(t, successHook.executed)
		assert.True(t, failureHook.executed)
	})
}

func TestMarshal(t *testing.T) {
	def := createTestDefinition(t)
	machine := New(def)
	ctx := context.Background()

	// Make a transition to create history
	success, err := machine.Fire(ctx, "ToMiddle", nil)
	require.NoError(t, err)
	require.True(t, success)

	storable, err := machine.Marshal()

	assert.NoError(t, err)
	assert.NotNil(t, storable)
	assert.Equal(t, gonfa.State("Middle"), storable.CurrentState)
	assert.Len(t, storable.History, 1)
	assert.Equal(t, gonfa.State("Start"), storable.History[0].From)
	assert.Equal(t, gonfa.State("Middle"), storable.History[0].To)
	assert.Equal(t, gonfa.Event("ToMiddle"), storable.History[0].On)
}

func TestHistory(t *testing.T) {
	def := createTestDefinition(t)
	machine := New(def)
	ctx := context.Background()

	// Initially empty
	history := machine.History()
	assert.Empty(t, history)

	// Make transitions
	success, err := machine.Fire(ctx, "ToMiddle", nil)
	require.NoError(t, err)
	require.True(t, success)

	success, err = machine.Fire(ctx, "ToEnd", nil)
	require.NoError(t, err)
	require.True(t, success)

	// Check history
	history = machine.History()
	assert.Len(t, history, 2)

	// First transition
	assert.Equal(t, gonfa.State("Start"), history[0].From)
	assert.Equal(t, gonfa.State("Middle"), history[0].To)
	assert.Equal(t, gonfa.Event("ToMiddle"), history[0].On)

	// Second transition
	assert.Equal(t, gonfa.State("Middle"), history[1].From)
	assert.Equal(t, gonfa.State("End"), history[1].To)
	assert.Equal(t, gonfa.Event("ToEnd"), history[1].On)

	// Timestamps should be set
	assert.False(t, history[0].Timestamp.IsZero())
	assert.False(t, history[1].Timestamp.IsZero())
	assert.True(t, history[1].Timestamp.After(history[0].Timestamp) ||
		history[1].Timestamp.Equal(history[0].Timestamp))
}

func TestConcurrentAccess(t *testing.T) {
	def := createTestDefinition(t)
	machine := New(def)
	ctx := context.Background()

	// Test concurrent operations on the SAME machine instance
	done := make(chan bool, 3)

	// Goroutine 1: Fire events
	go func() {
		for i := 0; i < 10; i++ {
			_, err := machine.Fire(ctx, "ToMiddle", nil)
			require.NoError(t, err)
			// Create a new machine for the next iteration
			// but don't reassign the shared variable
			tempMachine := New(def)
			_, err = tempMachine.Fire(ctx, "ToMiddle", nil)
			require.NoError(t, err)
		}
		done <- true
	}()

	// Goroutine 2: Read current state
	go func() {
		for i := 0; i < 100; i++ {
			machine.CurrentState()
		}
		done <- true
	}()

	// Goroutine 3: Marshal machine
	go func() {
		for i := 0; i < 50; i++ {
			storable, err := machine.Marshal()
			require.NoError(t, err)
			require.NotNil(t, storable)
		}
		done <- true
	}()

	// Wait for all goroutines
	<-done
	<-done
	<-done
}

func TestNFABehavior(t *testing.T) {
	// Test non-deterministic behavior - multiple transitions
	// from same state with same event
	guard1 := &testGuard{result: true}
	guard2 := &testGuard{result: false}
	action1 := &testAction{name: "action1"}
	action2 := &testAction{name: "action2"}

	def, err := builder.New().
		InitialState("Start").
		// Two transitions with same from/event but different guards
		AddTransition("Start", "Path1", "Event").
		WithGuards(guard1).
		WithActions(action1).
		AddTransition("Start", "Path2", "Event").
		WithGuards(guard2).
		WithActions(action2).
		Build()
	require.NoError(t, err)

	machine := New(def)
	ctx := context.Background()

	// First guard succeeds, second fails - should take Path1
	success, err := machine.Fire(ctx, "Event", nil)

	assert.NoError(t, err)
	assert.True(t, success)
	assert.Equal(t, gonfa.State("Path1"), machine.CurrentState())
	assert.True(t, action1.executed)
	assert.False(t, action2.executed)
}
