package definition

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dr-dobermann/gonfa/pkg/gonfa"
)

func TestStateSet(t *testing.T) {
	t.Run("newStateSet creates proper set", func(t *testing.T) {
		states := []gonfa.State{"A", "B", "C"}
		set := newStateSet(states)
		
		assert.Len(t, set, 3)
		assert.True(t, set.contains("A"))
		assert.True(t, set.contains("B"))
		assert.True(t, set.contains("C"))
		assert.False(t, set.contains("D"))
	})

	t.Run("empty state set", func(t *testing.T) {
		set := newStateSet([]gonfa.State{})
		assert.Len(t, set, 0)
		assert.False(t, set.contains("A"))
	})

	t.Run("duplicate states in slice", func(t *testing.T) {
		states := []gonfa.State{"A", "B", "A", "C"}
		set := newStateSet(states)
		
		// Set should deduplicate
		assert.Len(t, set, 3)
		assert.True(t, set.contains("A"))
		assert.True(t, set.contains("B"))
		assert.True(t, set.contains("C"))
	})
}

func TestTransitionGraph(t *testing.T) {
	t.Run("build graph from transitions", func(t *testing.T) {
		transitions := []Transition{
			{From: "A", To: "B", On: "event1"},
			{From: "A", To: "C", On: "event2"},
			{From: "B", To: "C", On: "event3"},
		}
		
		graph, err := newTransitionGraph(transitions)
		assert.NoError(t, err)
		
		assert.Len(t, graph, 2)
		assert.Len(t, graph["A"], 2)
		assert.Len(t, graph["B"], 1)
		assert.True(t, graph["A"].contains("B"))
		assert.True(t, graph["A"].contains("C"))
		assert.True(t, graph["B"].contains("C"))
	})

	t.Run("different events to same target are allowed", func(t *testing.T) {
		transitions := []Transition{
			{From: "A", To: "B", On: "event1"},
			{From: "A", To: "B", On: "event2"}, // Same states, different event - allowed
		}
		
		graph, err := newTransitionGraph(transitions)
		assert.NoError(t, err)
		
		// Should have one A->B in graph (connectivity), but both events are valid
		assert.Len(t, graph["A"], 1)
		assert.True(t, graph["A"].contains("B"))
	})

	t.Run("exact duplicate transitions cause error", func(t *testing.T) {
		transitions := []Transition{
			{From: "A", To: "B", On: "event1"},
			{From: "A", To: "B", On: "event1"}, // Exact duplicate - error
		}
		
		_, err := newTransitionGraph(transitions)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate transition from 'A' to 'B' on event 'event1'")
	})

	t.Run("empty transitions", func(t *testing.T) {
		graph, err := newTransitionGraph([]Transition{})
		assert.NoError(t, err)
		assert.Len(t, graph, 0)
	})
}

func TestCheckStatesOptimized(t *testing.T) {
	t.Run("valid simple state machine", func(t *testing.T) {
		initialState := gonfa.State("Start")
		states := []gonfa.State{"Start", "End"}
		finalStates := []gonfa.State{"End"}
		transitions := []Transition{
			{From: "Start", To: "End", On: "finish"},
		}

		err := checkStates(initialState, states, transitions, finalStates)
		assert.NoError(t, err)
	})

	t.Run("initial state not in states", func(t *testing.T) {
		initialState := gonfa.State("NonExistent")
		states := []gonfa.State{"Start", "End"}
		finalStates := []gonfa.State{"End"}
		transitions := []Transition{
			{From: "Start", To: "End", On: "finish"},
		}

		err := checkStates(initialState, states, transitions, finalStates)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "initial state 'NonExistent' doesn't exist in states")
	})

	t.Run("final state not in states", func(t *testing.T) {
		initialState := gonfa.State("Start")
		states := []gonfa.State{"Start", "Middle"}
		finalStates := []gonfa.State{"End"}
		transitions := []Transition{
			{From: "Start", To: "Middle", On: "move"},
		}

		err := checkStates(initialState, states, transitions, finalStates)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "final state 'End' doesn't exist in states")
	})

	t.Run("transition source state not in states", func(t *testing.T) {
		initialState := gonfa.State("Start")
		states := []gonfa.State{"Start", "End"}
		finalStates := []gonfa.State{"End"}
		transitions := []Transition{
			{From: "NonExistent", To: "End", On: "finish"},
		}

		err := checkStates(initialState, states, transitions, finalStates)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "state 'NonExistent' doesn't exist as transition source")
	})

	t.Run("transition target state not in states", func(t *testing.T) {
		initialState := gonfa.State("Start")
		states := []gonfa.State{"Start", "Middle"}
		finalStates := []gonfa.State{}
		transitions := []Transition{
			{From: "Start", To: "NonExistent", On: "move"},
		}

		err := checkStates(initialState, states, transitions, finalStates)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "state 'NonExistent' doesn't exist as transition target")
	})

	t.Run("duplicate transitions", func(t *testing.T) {
		initialState := gonfa.State("Start")
		states := []gonfa.State{"Start", "End"}
		finalStates := []gonfa.State{"End"}
		transitions := []Transition{
			{From: "Start", To: "End", On: "finish"},
			{From: "Start", To: "End", On: "finish"}, // Exact duplicate
		}

		err := checkStates(initialState, states, transitions, finalStates)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate transition from 'Start' to 'End' on event 'finish'")
	})
}

func TestValidateInitialState(t *testing.T) {
	t.Run("valid initial state", func(t *testing.T) {
		stateSet := newStateSet([]gonfa.State{"Start", "End"})
		err := validateInitialState("Start", stateSet)
		assert.NoError(t, err)
	})

	t.Run("invalid initial state", func(t *testing.T) {
		stateSet := newStateSet([]gonfa.State{"Start", "End"})
		err := validateInitialState("NonExistent", stateSet)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "initial state 'NonExistent' doesn't exist in states")
	})
}

func TestValidateFinalStates(t *testing.T) {
	t.Run("valid final states", func(t *testing.T) {
		stateSet := newStateSet([]gonfa.State{"Start", "End1", "End2"})
		finalSet := newStateSet([]gonfa.State{"End1", "End2"})
		err := validateFinalStates(finalSet, stateSet)
		assert.NoError(t, err)
	})

	t.Run("invalid final state", func(t *testing.T) {
		stateSet := newStateSet([]gonfa.State{"Start", "End1"})
		finalSet := newStateSet([]gonfa.State{"End1", "NonExistent"})
		err := validateFinalStates(finalSet, stateSet)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "final state 'NonExistent' doesn't exist in states")
	})

	t.Run("empty final states", func(t *testing.T) {
		stateSet := newStateSet([]gonfa.State{"Start", "End"})
		finalSet := newStateSet([]gonfa.State{})
		err := validateFinalStates(finalSet, stateSet)
		assert.NoError(t, err)
	})
}

func TestBuildStateCounters(t *testing.T) {
	t.Run("simple graph counters", func(t *testing.T) {
		stateSet := newStateSet([]gonfa.State{"A", "B", "C"})
		graph := transitionGraph{
			"A": newStateSet([]gonfa.State{"B", "C"}),
			"B": newStateSet([]gonfa.State{"C"}),
		}

		counters := buildStateCounters(stateSet, graph)

		assert.Equal(t, 0, counters["A"].incoming)
		assert.Equal(t, 2, counters["A"].outgoing)
		
		assert.Equal(t, 1, counters["B"].incoming)
		assert.Equal(t, 1, counters["B"].outgoing)
		
		assert.Equal(t, 2, counters["C"].incoming)
		assert.Equal(t, 0, counters["C"].outgoing)
	})

	t.Run("isolated state", func(t *testing.T) {
		stateSet := newStateSet([]gonfa.State{"A", "B"})
		graph := transitionGraph{
			"A": newStateSet([]gonfa.State{"B"}),
		}

		counters := buildStateCounters(stateSet, graph)

		assert.Equal(t, 0, counters["A"].incoming)
		assert.Equal(t, 1, counters["A"].outgoing)
		
		assert.Equal(t, 1, counters["B"].incoming)
		assert.Equal(t, 0, counters["B"].outgoing)
	})
}

func TestFindReachableStates(t *testing.T) {
	t.Run("linear graph", func(t *testing.T) {
		graph := transitionGraph{
			"A": newStateSet([]gonfa.State{"B"}),
			"B": newStateSet([]gonfa.State{"C"}),
		}

		reachable := findReachableStates("A", graph)

		assert.True(t, reachable.contains("A"))
		assert.True(t, reachable.contains("B"))
		assert.True(t, reachable.contains("C"))
		assert.Len(t, reachable, 3)
	})

	t.Run("branching graph", func(t *testing.T) {
		graph := transitionGraph{
			"A": newStateSet([]gonfa.State{"B", "C"}),
			"B": newStateSet([]gonfa.State{"D"}),
			"C": newStateSet([]gonfa.State{"D"}),
		}

		reachable := findReachableStates("A", graph)

		assert.True(t, reachable.contains("A"))
		assert.True(t, reachable.contains("B"))
		assert.True(t, reachable.contains("C"))
		assert.True(t, reachable.contains("D"))
		assert.Len(t, reachable, 4)
	})

	t.Run("cyclic graph", func(t *testing.T) {
		graph := transitionGraph{
			"A": newStateSet([]gonfa.State{"B"}),
			"B": newStateSet([]gonfa.State{"C"}),
			"C": newStateSet([]gonfa.State{"A"}), // Cycle back
		}

		reachable := findReachableStates("A", graph)

		assert.True(t, reachable.contains("A"))
		assert.True(t, reachable.contains("B"))
		assert.True(t, reachable.contains("C"))
		assert.Len(t, reachable, 3)
	})

	t.Run("single state", func(t *testing.T) {
		graph := transitionGraph{
			"A": newStateSet([]gonfa.State{}),
		}

		reachable := findReachableStates("A", graph)

		assert.True(t, reachable.contains("A"))
		assert.Len(t, reachable, 1)
	})
}

func TestValidateInitialStateUsage(t *testing.T) {
	t.Run("initial state has transitions", func(t *testing.T) {
		graph := transitionGraph{
			"Start": newStateSet([]gonfa.State{"End"}),
		}

		err := validateInitialStateUsage("Start", graph)
		assert.NoError(t, err)
	})

	t.Run("initial state has no transitions", func(t *testing.T) {
		graph := transitionGraph{
			"Other": newStateSet([]gonfa.State{"End"}),
		}

		err := validateInitialStateUsage("Start", graph)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no transitions start from initial state 'Start'")
	})
}

func TestValidateSingleState(t *testing.T) {
	t.Run("valid normal state", func(t *testing.T) {
		counter := &stateCounter{incoming: 1, outgoing: 1}
		finalSet := newStateSet([]gonfa.State{"End"})

		err := validateSingleState("Middle", counter, "Start", finalSet)
		assert.NoError(t, err)
	})

	t.Run("valid initial state", func(t *testing.T) {
		counter := &stateCounter{incoming: 0, outgoing: 1}
		finalSet := newStateSet([]gonfa.State{"End"})

		err := validateSingleState("Start", counter, "Start", finalSet)
		assert.NoError(t, err)
	})

	t.Run("valid final state", func(t *testing.T) {
		counter := &stateCounter{incoming: 1, outgoing: 0}
		finalSet := newStateSet([]gonfa.State{"End"})

		err := validateSingleState("End", counter, "Start", finalSet)
		assert.NoError(t, err)
	})

	t.Run("hanging state", func(t *testing.T) {
		counter := &stateCounter{incoming: 0, outgoing: 1}
		finalSet := newStateSet([]gonfa.State{"End"})

		err := validateSingleState("Hanging", counter, "Start", finalSet)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "state 'Hanging' isn't an initial state but has no incoming transitions")
	})

	t.Run("dead-end non-final state", func(t *testing.T) {
		counter := &stateCounter{incoming: 1, outgoing: 0}
		finalSet := newStateSet([]gonfa.State{"End"})

		err := validateSingleState("DeadEnd", counter, "Start", finalSet)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "state 'DeadEnd' is a dead-end state")
	})

	t.Run("final state with outgoing transitions", func(t *testing.T) {
		counter := &stateCounter{incoming: 1, outgoing: 1}
		finalSet := newStateSet([]gonfa.State{"BadFinal"})

		err := validateSingleState("BadFinal", counter, "Start", finalSet)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "final state 'BadFinal' has outgoing transition(s)")
	})
}

func TestValidateFinalStateReachability(t *testing.T) {
	t.Run("all final states reachable", func(t *testing.T) {
		finalSet := newStateSet([]gonfa.State{"End1", "End2"})
		reachable := newStateSet([]gonfa.State{"Start", "Middle", "End1", "End2"})

		err := validateFinalStateReachability(finalSet, reachable)
		assert.NoError(t, err)
	})

	t.Run("unreachable final state", func(t *testing.T) {
		finalSet := newStateSet([]gonfa.State{"End1", "End2"})
		reachable := newStateSet([]gonfa.State{"Start", "Middle", "End1"})

		err := validateFinalStateReachability(finalSet, reachable)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "final state 'End2' is not reachable from initial state")
	})

	t.Run("empty final states", func(t *testing.T) {
		finalSet := newStateSet([]gonfa.State{})
		reachable := newStateSet([]gonfa.State{"Start", "End"})

		err := validateFinalStateReachability(finalSet, reachable)
		assert.NoError(t, err)
	})
}

// Integration tests for complex scenarios
func TestCheckStatesIntegration(t *testing.T) {
	t.Run("document workflow", func(t *testing.T) {
		initialState := gonfa.State("Draft")
		states := []gonfa.State{"Draft", "InReview", "Approved", "Rejected"}
		finalStates := []gonfa.State{"Approved", "Rejected"}
		transitions := []Transition{
			{From: "Draft", To: "InReview", On: "Submit"},
			{From: "InReview", To: "Approved", On: "Approve"},
			{From: "InReview", To: "Rejected", On: "Reject"},
		}

		err := checkStates(initialState, states, transitions, finalStates)
		assert.NoError(t, err)
	})

	t.Run("branching workflow", func(t *testing.T) {
		initialState := gonfa.State("Start")
		states := []gonfa.State{"Start", "PathA", "PathB", "End"}
		finalStates := []gonfa.State{"End"}
		transitions := []Transition{
			{From: "Start", To: "PathA", On: "ChooseA"},
			{From: "Start", To: "PathB", On: "ChooseB"},
			{From: "PathA", To: "End", On: "Finish"},
			{From: "PathB", To: "End", On: "Finish"},
		}

		err := checkStates(initialState, states, transitions, finalStates)
		assert.NoError(t, err)
	})

	t.Run("cyclic workflow with final state", func(t *testing.T) {
		initialState := gonfa.State("Start")
		states := []gonfa.State{"Start", "Loop", "End"}
		finalStates := []gonfa.State{"End"}
		transitions := []Transition{
			{From: "Start", To: "Loop", On: "Begin"},
			{From: "Loop", To: "Loop", On: "Continue"},
			{From: "Loop", To: "End", On: "Finish"},
		}

		err := checkStates(initialState, states, transitions, finalStates)
		assert.NoError(t, err)
	})

	t.Run("invalid: unreachable final state", func(t *testing.T) {
		initialState := gonfa.State("Start")
		states := []gonfa.State{"Start", "Middle", "Reachable", "Unreachable"}
		finalStates := []gonfa.State{"Reachable", "Unreachable"}
		transitions := []Transition{
			{From: "Start", To: "Middle", On: "Move"},
			{From: "Middle", To: "Reachable", On: "Reach"},
			// No path to Unreachable
		}

		err := checkStates(initialState, states, transitions, finalStates)
		assert.Error(t, err)
		// The error can be either about hanging state or unreachable final state
		// Both are valid detection points for this invalid configuration
		assert.True(t, 
			err.Error() == "state 'Unreachable' isn't an initial state but has no incoming transitions" ||
			err.Error() == "final state 'Unreachable' is not reachable from initial state",
			"Expected hanging state or unreachable final state error, got: %s", err.Error())
	})

	t.Run("initial state as final state", func(t *testing.T) {
		initialState := gonfa.State("SingleState")
		states := []gonfa.State{"SingleState"}
		finalStates := []gonfa.State{"SingleState"}
		transitions := []Transition{}

		err := checkStates(initialState, states, transitions, finalStates)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no transitions start from initial state")
	})
}