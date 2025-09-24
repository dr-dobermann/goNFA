package machine

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dr-dobermann/gonfa/pkg/builder"
	"github.com/dr-dobermann/gonfa/pkg/gonfa"
)

func TestConcurrentAccess(t *testing.T) {
	def, err := builder.New().
		InitialState("Start").
		AddTransition("Start", "Middle", "ToMiddle").
		AddTransition("Middle", "End", "ToEnd").
		Build()
	require.NoError(t, err)

	machine, err := New(def, nil)
	require.NoError(t, err)

	const numGoroutines = 10
	const numOperations = 100

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*numOperations)

	// Start multiple goroutines that read and write to the machine
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				// Read operations
				state := machine.CurrentState()
				history := machine.History()
				_ = state
				_ = history

				// Write operations (only if in Start state)
				if state == "Start" {
					success, err := machine.Fire(context.Background(), "ToMiddle", nil)
					if err != nil {
						errors <- err
					}
					if success {
						// Try to transition to End
						success, err := machine.Fire(context.Background(), "ToEnd", nil)
						if err != nil {
							errors <- err
						}
						_ = success
					}
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		require.NoError(t, err)
	}
}

func TestConcurrentMarshal(t *testing.T) {
	def := createTestDefinition(t)
	machine, err := New(def, nil)
	require.NoError(t, err)

	const numGoroutines = 10
	const numOperations = 50

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*numOperations)

	// Start multiple goroutines that marshal the machine
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				storable, err := machine.Marshal()
				if err != nil {
					errors <- err
				}
				require.NotNil(t, storable)
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		require.NoError(t, err)
	}
}

func TestConcurrentFire(t *testing.T) {
	def, err := builder.New().
		InitialState("Start").
		AddTransition("Start", "Middle", "ToMiddle").
		AddTransition("Middle", "End", "ToEnd").
		Build()
	require.NoError(t, err)

	machine, err := New(def, nil)
	require.NoError(t, err)

	const numGoroutines = 5
	var wg sync.WaitGroup
	results := make(chan bool, numGoroutines)

	// Start multiple goroutines that try to fire the same event
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			success, err := machine.Fire(context.Background(), "ToMiddle", nil)
			require.NoError(t, err)
			results <- success
		}()
	}

	wg.Wait()
	close(results)

	// Only one transition should succeed
	successCount := 0
	for success := range results {
		if success {
			successCount++
		}
	}

	assert.Equal(t, 1, successCount)
	assert.Equal(t, gonfa.State("Middle"), machine.CurrentState())
}

func TestConcurrentStateExtender(t *testing.T) {
	def := createTestDefinition(t)
	extender := &testStateExtender{data: "test"}

	machine, err := New(def, extender)
	require.NoError(t, err)

	const numGoroutines = 10
	const numOperations = 100

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*numOperations)

	// Start multiple goroutines that access StateExtender
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				ext := machine.StateExtender()
				if ext == nil {
					errors <- assert.AnError
				}
				// Verify it's the same extender
				if ext != extender {
					errors <- assert.AnError
				}
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		require.NoError(t, err)
	}
}
