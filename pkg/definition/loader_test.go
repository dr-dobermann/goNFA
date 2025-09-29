package definition

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dr-dobermann/gonfa/pkg/gonfa"
	"github.com/dr-dobermann/gonfa/pkg/registry"
)

func createTestRegistry() *registry.Registry {
	reg := registry.New()
	reg.RegisterAction("action1", &testAction{name: "action1"})
	reg.RegisterAction("action2", &testAction{name: "action2"})
	reg.RegisterGuard("guard1", &testGuard{result: true})
	reg.RegisterGuard("guard2", &testGuard{result: false})
	return reg
}

func TestLoadDefinition(t *testing.T) {
	t.Run("successful load", func(t *testing.T) {
		yamlData := `
initialState: Start
finalStates:
  - End1
  - End2
hooks:
  onSuccess:
    - action1
  onFailure:
    - action2
states:
  Start:
    onEntry:
      - action1
    onExit:
      - action2
  Middle:
    onEntry:
      - action1
  End1:
    onEntry:
      - action1
  End2: {}
transitions:
  - from: Start
    to: Middle
    on: Event1
    guards:
      - guard1
    actions:
      - action1
  - from: Middle
    to: End1
    on: Event2
    guards:
      - guard2
    actions:
      - action2
  - from: Middle
    to: End2
    on: Event3
`

		reg := createTestRegistry()
		reader := strings.NewReader(yamlData)

		def, err := LoadDefinition(reader, reg)
		require.NoError(t, err)
		assert.NotNil(t, def)
		assert.Equal(t, gonfa.State("Start"), def.InitialState())
		assert.True(t, def.IsFinalState("End1"))
		assert.True(t, def.IsFinalState("End2"))
		assert.False(t, def.IsFinalState("Start"))
	})

	t.Run("missing initialState", func(t *testing.T) {
		yamlData := `
transitions:
  - from: Start
    to: End
    on: Event1
`

		reg := createTestRegistry()
		reader := strings.NewReader(yamlData)

		_, err := LoadDefinition(reader, reg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "initialState is required")
	})

	t.Run("no transitions", func(t *testing.T) {
		yamlData := `
initialState: Start
`

		reg := createTestRegistry()
		reader := strings.NewReader(yamlData)

		_, err := LoadDefinition(reader, reg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least one transition is required")
	})

	t.Run("invalid YAML", func(t *testing.T) {
		yamlData := `invalid: yaml: content: [`

		reg := createTestRegistry()
		reader := strings.NewReader(yamlData)

		_, err := LoadDefinition(reader, reg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse YAML")
	})

	t.Run("missing action in registry", func(t *testing.T) {
		yamlData := `
initialState: Start
transitions:
  - from: Start
    to: End
    on: Event1
    actions:
      - nonExistentAction
`

		reg := createTestRegistry()
		reader := strings.NewReader(yamlData)

		_, err := LoadDefinition(reader, reg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "action 'nonExistentAction' not found in registry")
	})

	t.Run("missing guard in registry", func(t *testing.T) {
		yamlData := `
initialState: Start
transitions:
  - from: Start
    to: End
    on: Event1
    guards:
      - nonExistentGuard
`

		reg := createTestRegistry()
		reader := strings.NewReader(yamlData)

		_, err := LoadDefinition(reader, reg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "guard 'nonExistentGuard' not found in registry")
	})

	t.Run("missing hook action in registry", func(t *testing.T) {
		yamlData := `
initialState: Start
hooks:
  onSuccess:
    - nonExistentAction
transitions:
  - from: Start
    to: End
    on: Event1
`

		reg := createTestRegistry()
		reader := strings.NewReader(yamlData)

		_, err := LoadDefinition(reader, reg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "success hook action 'nonExistentAction' not found in registry")
	})

	t.Run("missing failure hook action in registry", func(t *testing.T) {
		yamlData := `
initialState: Start
hooks:
  onFailure:
    - nonExistentAction
transitions:
  - from: Start
    to: End
    on: Event1
`

		reg := createTestRegistry()
		reader := strings.NewReader(yamlData)

		_, err := LoadDefinition(reader, reg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failure hook action 'nonExistentAction' not found in registry")
	})
}

func TestLoadDefinitionWithStates(t *testing.T) {
	yamlData := `
initialState: Start
finalStates:
  - End
states:
  Start:
    onEntry:
      - action1
    onExit:
      - action2
  End:
    onEntry:
      - action1
transitions:
  - from: Start
    to: End
    on: Event1
`

	reg := createTestRegistry()
	reader := strings.NewReader(yamlData)

	def, err := LoadDefinition(reader, reg)
	require.NoError(t, err)

	// Check states configuration
	states := def.States()
	assert.Len(t, states, 2)

	startConfig := def.GetStateConfig("Start")
	assert.Len(t, startConfig.OnEntry, 1)
	assert.Len(t, startConfig.OnExit, 1)

	endConfig := def.GetStateConfig("End")
	assert.Len(t, endConfig.OnEntry, 1)
	assert.Empty(t, endConfig.OnExit)
}

func TestLoadDefinitionWithHooks(t *testing.T) {
	yamlData := `
initialState: Start
finalStates:
  - End
states:
  Start: {}
  End: {}
hooks:
  onSuccess:
    - action1
  onFailure:
    - action2
transitions:
  - from: Start
    to: End
    on: Event1
`

	reg := createTestRegistry()
	reader := strings.NewReader(yamlData)

	def, err := LoadDefinition(reader, reg)
	require.NoError(t, err)

	hooks := def.Hooks()
	assert.Len(t, hooks.OnSuccess, 1)
	assert.Len(t, hooks.OnFailure, 1)
}

func TestLoadDefinitionWithTransitions(t *testing.T) {
	yamlData := `
initialState: Start
finalStates:
  - End
states:
  Start: {}
  Middle: {}
  End: {}
transitions:
  - from: Start
    to: Middle
    on: Event1
    guards:
      - guard1
    actions:
      - action1
  - from: Middle
    to: End
    on: Event2
    guards:
      - guard2
    actions:
      - action2
`

	reg := createTestRegistry()
	reader := strings.NewReader(yamlData)

	def, err := LoadDefinition(reader, reg)
	require.NoError(t, err)

	transitions := def.Transitions()
	assert.Len(t, transitions, 2)

	// Check first transition
	assert.Equal(t, "Start", string(transitions[0].From))
	assert.Equal(t, "Middle", string(transitions[0].To))
	assert.Equal(t, "Event1", string(transitions[0].On))
	assert.Len(t, transitions[0].Guards, 1)
	assert.Len(t, transitions[0].Actions, 1)

	// Check second transition
	assert.Equal(t, "Middle", string(transitions[1].From))
	assert.Equal(t, "End", string(transitions[1].To))
	assert.Equal(t, "Event2", string(transitions[1].On))
	assert.Len(t, transitions[1].Guards, 1)
	assert.Len(t, transitions[1].Actions, 1)
}
