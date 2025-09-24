package builder

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dr-dobermann/gonfa/pkg/gonfa"
)

func TestAddTransition(t *testing.T) {
	builder := New()
	from := gonfa.State("From")
	to := gonfa.State("To")
	on := gonfa.Event("Event")

	result := builder.AddTransition(from, to, on)

	assert.Equal(t, builder, result) // Fluent interface
	assert.Len(t, builder.transitions, 1)
	transition := builder.transitions[0]
	assert.Equal(t, from, transition.From)
	assert.Equal(t, to, transition.To)
	assert.Equal(t, on, transition.On)
	assert.NotNil(t, builder.lastTransition)
	assert.Equal(t, &builder.transitions[0], builder.lastTransition)
}

func TestAddMultipleTransitions(t *testing.T) {
	builder := New()

	builder.AddTransition("From1", "To1", "Event1")
	builder.AddTransition("From2", "To2", "Event2")

	assert.Len(t, builder.transitions, 2)
	assert.Equal(t, gonfa.State("From1"), builder.transitions[0].From)
	assert.Equal(t, gonfa.State("To1"), builder.transitions[0].To)
	assert.Equal(t, gonfa.Event("Event1"), builder.transitions[0].On)
	assert.Equal(t, gonfa.State("From2"), builder.transitions[1].From)
	assert.Equal(t, gonfa.State("To2"), builder.transitions[1].To)
	assert.Equal(t, gonfa.Event("Event2"), builder.transitions[1].On)
}

func TestWithGuards(t *testing.T) {
	builder := New()
	guard1 := &testGuard{result: true}
	guard2 := &testGuard{result: false}

	// Add transition first
	builder.AddTransition("From", "To", "Event")
	result := builder.WithGuards(guard1, guard2)

	assert.Equal(t, builder, result) // Fluent interface
	assert.Len(t, builder.lastTransition.Guards, 2)
	assert.Contains(t, builder.lastTransition.Guards, guard1)
	assert.Contains(t, builder.lastTransition.Guards, guard2)
}

func TestWithGuardsNoTransition(t *testing.T) {
	builder := New()
	guard := &testGuard{result: true}

	// Call WithGuards without adding transition first
	result := builder.WithGuards(guard)

	assert.Equal(t, builder, result) // Fluent interface
	// Should not panic, but guards won't be added
}

func TestWithActions(t *testing.T) {
	builder := New()
	action1 := &testAction{name: "action1"}
	action2 := &testAction{name: "action2"}

	// Add transition first
	builder.AddTransition("From", "To", "Event")
	result := builder.WithActions(action1, action2)

	assert.Equal(t, builder, result) // Fluent interface
	assert.Len(t, builder.lastTransition.Actions, 2)
	assert.Contains(t, builder.lastTransition.Actions, action1)
	assert.Contains(t, builder.lastTransition.Actions, action2)
}

func TestWithActionsNoTransition(t *testing.T) {
	builder := New()
	action := &testAction{name: "action"}

	// Call WithActions without adding transition first
	result := builder.WithActions(action)

	assert.Equal(t, builder, result) // Fluent interface
	// Should not panic, but actions won't be added
}

func TestWithGuardsAndActions(t *testing.T) {
	builder := New()
	guard := &testGuard{result: true}
	action := &testAction{name: "action"}

	builder.AddTransition("From", "To", "Event").
		WithGuards(guard).
		WithActions(action)

	transition := builder.transitions[0]
	assert.Len(t, transition.Guards, 1)
	assert.Len(t, transition.Actions, 1)
	assert.Contains(t, transition.Guards, guard)
	assert.Contains(t, transition.Actions, action)
}
