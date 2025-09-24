package machine

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dr-dobermann/gonfa/pkg/gonfa"
)

func TestMarshal(t *testing.T) {
	def := createTestDefinition(t)
	machine, err := New(def, nil)
	require.NoError(t, err)

	// Test initial state
	storable, err := machine.Marshal()
	require.NoError(t, err)
	assert.Equal(t, gonfa.State("Start"), storable.CurrentState)
	assert.Empty(t, storable.History)

	// Test after transition
	success, err := machine.Fire(context.Background(), "ToMiddle", nil)
	require.NoError(t, err)
	assert.True(t, success)

	storable, err = machine.Marshal()
	require.NoError(t, err)
	assert.Equal(t, gonfa.State("Middle"), storable.CurrentState)
	assert.Len(t, storable.History, 1)
	assert.Equal(t, gonfa.State("Start"), storable.History[0].From)
	assert.Equal(t, gonfa.State("Middle"), storable.History[0].To)
	assert.Equal(t, gonfa.Event("ToMiddle"), storable.History[0].On)
}

func TestMarshalImmutable(t *testing.T) {
	def := createTestDefinition(t)
	machine, err := New(def, nil)
	require.NoError(t, err)

	// Get initial storable
	storable1, err := machine.Marshal()
	require.NoError(t, err)

	// Make transition
	success, err := machine.Fire(context.Background(), "ToMiddle", nil)
	require.NoError(t, err)
	assert.True(t, success)

	// Get new storable
	storable2, err := machine.Marshal()
	require.NoError(t, err)

	// Original storable should be unchanged
	assert.Equal(t, gonfa.State("Start"), storable1.CurrentState)
	assert.Empty(t, storable1.History)

	// New storable should reflect current state
	assert.Equal(t, gonfa.State("Middle"), storable2.CurrentState)
	assert.Len(t, storable2.History, 1)
}

func TestRestoreFromMarshal(t *testing.T) {
	def := createTestDefinition(t)
	extender := &testStateExtender{data: "test"}

	// Create machine and make transitions
	machine1, err := New(def, extender)
	require.NoError(t, err)

	success, err := machine1.Fire(context.Background(), "ToMiddle", nil)
	require.NoError(t, err)
	assert.True(t, success)

	success, err = machine1.Fire(context.Background(), "ToEnd", nil)
	require.NoError(t, err)
	assert.True(t, success)

	// Marshal state
	storable, err := machine1.Marshal()
	require.NoError(t, err)

	// Restore from marshaled state
	machine2, err := Restore(def, storable, extender)
	require.NoError(t, err)

	// Verify restored state
	assert.Equal(t, gonfa.State("End"), machine2.CurrentState())
	assert.Len(t, machine2.History(), 2)
	assert.Equal(t, extender, machine2.StateExtender())
}

func TestRestoreWithHistory(t *testing.T) {
	def := createTestDefinition(t)
	extender := &testStateExtender{data: "test"}

	// Create storable with history
	history := []gonfa.HistoryEntry{
		{
			From:      "Start",
			To:        "Middle",
			On:        "ToMiddle",
			Timestamp: time.Now().Add(-time.Hour),
		},
		{
			From:      "Middle",
			To:        "End",
			On:        "ToEnd",
			Timestamp: time.Now(),
		},
	}

	storable := &gonfa.Storable{
		CurrentState: "End",
		History:      history,
	}

	// Restore machine
	machine, err := Restore(def, storable, extender)
	require.NoError(t, err)

	// Verify restored state
	assert.Equal(t, gonfa.State("End"), machine.CurrentState())
	assert.Len(t, machine.History(), 2)
	assert.Equal(t, history[0], machine.History()[0])
	assert.Equal(t, history[1], machine.History()[1])
}
