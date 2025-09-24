package builder

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dr-dobermann/gonfa/pkg/definition"
	"github.com/dr-dobermann/gonfa/pkg/gonfa"
)

func TestWithHooks(t *testing.T) {
	builder := New()
	hooks := definition.Hooks{
		OnSuccess: []gonfa.Action{&testAction{name: "success1"}},
		OnFailure: []gonfa.Action{&testAction{name: "failure1"}},
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

func TestWithHooksMultipleCalls(t *testing.T) {
	builder := New()
	action1 := &testAction{name: "success1"}
	action2 := &testAction{name: "success2"}

	builder.WithSuccessHooks(action1)
	result := builder.WithSuccessHooks(action2)

	assert.Equal(t, builder, result) // Fluent interface
	assert.Len(t, builder.hooks.OnSuccess, 2)
	assert.Contains(t, builder.hooks.OnSuccess, action1)
	assert.Contains(t, builder.hooks.OnSuccess, action2)
}

func TestWithHooksCombined(t *testing.T) {
	builder := New()
	successAction := &testAction{name: "success"}
	failureAction := &testAction{name: "failure"}

	builder.WithSuccessHooks(successAction).
		WithFailureHooks(failureAction)

	assert.Len(t, builder.hooks.OnSuccess, 1)
	assert.Len(t, builder.hooks.OnFailure, 1)
	assert.Contains(t, builder.hooks.OnSuccess, successAction)
	assert.Contains(t, builder.hooks.OnFailure, failureAction)
}
