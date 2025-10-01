package definition

import (
	"fmt"

	"github.com/dr-dobermann/gonfa/pkg/gonfa"
)

// stateSet represents a set of states for fast lookups
type stateSet map[gonfa.State]struct{}

// newStateSet creates a new state set from a slice of states
func newStateSet(states []gonfa.State) stateSet {
	set := make(stateSet, len(states))
	for _, state := range states {
		set[state] = struct{}{}
	}
	return set
}

// contains checks if state exists in the set
func (s stateSet) contains(state gonfa.State) bool {
	_, exists := s[state]
	return exists
}

// transitionGraph represents state transition graph
type transitionGraph map[gonfa.State]stateSet

// transitionKey uniquely identifies a transition
type transitionKey struct {
	from gonfa.State
	to   gonfa.State
	on   gonfa.Event
}

// newTransitionGraph builds transition graph from transitions slice
// and validates for duplicate transitions
func newTransitionGraph(transitions []Transition) (transitionGraph, error) {
	graph := make(transitionGraph)
	seen := make(map[transitionKey]struct{})

	for _, t := range transitions {
		key := transitionKey{from: t.From, to: t.To, on: t.On}

		// Check for exact duplicate transition (From, To, Event)
		if _, exists := seen[key]; exists {
			return nil, fmt.Errorf(
				"duplicate transition from '%s' to '%s' on event '%s'",
				t.From, t.To, t.On)
		}
		seen[key] = struct{}{}

		// Build graph for connectivity analysis
		if graph[t.From] == nil {
			graph[t.From] = make(stateSet)
		}
		graph[t.From][t.To] = struct{}{}
	}

	return graph, nil
}

// stateCounter tracks incoming and outgoing transition counts
type stateCounter struct {
	incoming int
	outgoing int
}

// checkStates performs optimized integrity check
func checkStates(
	initialState gonfa.State,
	states []gonfa.State,
	transitions []Transition,
	finalStates []gonfa.State,
) error {
	stateSet := newStateSet(states)
	finalSet := newStateSet(finalStates)

	if err := validateInitialState(initialState, stateSet); err != nil {
		return err
	}

	if err := validateFinalStates(finalSet, stateSet); err != nil {
		return err
	}

	graph, err := newTransitionGraph(transitions)
	if err != nil {
		return err
	}

	if err := validateTransitionStates(graph, stateSet); err != nil {
		return err
	}

	return analyzeGraphStructure(initialState, finalSet, stateSet, graph)
}

// validateInitialState checks if initial state exists
func validateInitialState(
	initialState gonfa.State,
	stateSet stateSet,
) error {
	if !stateSet.contains(initialState) {
		return fmt.Errorf(
			"initial state '%s' doesn't exist in states",
			initialState)
	}
	return nil
}

// validateFinalStates checks if all final states exist
func validateFinalStates(finalSet, stateSet stateSet) error {
	for state := range finalSet {
		if !stateSet.contains(state) {
			return fmt.Errorf(
				"final state '%s' doesn't exist in states",
				state)
		}
	}
	return nil
}

// validateTransitionStates checks if transition states exist
func validateTransitionStates(
	graph transitionGraph,
	stateSet stateSet,
) error {
	for fromState, toStates := range graph {
		if !stateSet.contains(fromState) {
			return fmt.Errorf(
				"state '%s' doesn't exist as transition source",
				fromState)
		}

		for toState := range toStates {
			if !stateSet.contains(toState) {
				return fmt.Errorf(
					"state '%s' doesn't exist as transition target",
					toState)
			}
		}
	}
	return nil
}

// analyzeGraphStructure performs graph connectivity and reachability checks
func analyzeGraphStructure(
	initialState gonfa.State,
	finalSet stateSet,
	stateSet stateSet,
	graph transitionGraph,
) error {
	counters := buildStateCounters(stateSet, graph)
	reachable := findReachableStates(initialState, graph)

	if err := validateInitialStateUsage(initialState, graph); err != nil {
		return err
	}

	if err := validateStateConnectivity(
		counters,
		initialState,
		finalSet,
	); err != nil {
		return err
	}

	return validateFinalStateReachability(finalSet, reachable)
}

// buildStateCounters creates transition counters for all states
func buildStateCounters(
	stateSet stateSet,
	graph transitionGraph,
) map[gonfa.State]*stateCounter {
	counters := make(map[gonfa.State]*stateCounter, len(stateSet))

	// Initialize counters for all states
	for state := range stateSet {
		counters[state] = &stateCounter{}
	}

	// Count transitions
	for fromState, toStates := range graph {
		if counter := counters[fromState]; counter != nil {
			counter.outgoing = len(toStates)
		}

		for toState := range toStates {
			if counter := counters[toState]; counter != nil {
				counter.incoming++
			}
		}
	}

	return counters
}

// findReachableStates performs BFS to find all reachable states
func findReachableStates(
	initialState gonfa.State,
	graph transitionGraph,
) stateSet {
	reachable := make(stateSet)
	queue := []gonfa.State{initialState}
	reachable[initialState] = struct{}{}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for nextState := range graph[current] {
			if !reachable.contains(nextState) {
				reachable[nextState] = struct{}{}
				queue = append(queue, nextState)
			}
		}
	}

	return reachable
}

// validateInitialStateUsage checks if initial state has transitions
func validateInitialStateUsage(
	initialState gonfa.State,
	graph transitionGraph,
) error {
	if _, exists := graph[initialState]; !exists {
		return fmt.Errorf(
			"no transitions start from initial state '%s'",
			initialState)
	}
	return nil
}

// validateStateConnectivity checks for hanging and dead-end states
func validateStateConnectivity(
	counters map[gonfa.State]*stateCounter,
	initialState gonfa.State,
	finalSet stateSet,
) error {
	for state, counter := range counters {
		if err := validateSingleState(
			state,
			counter,
			initialState,
			finalSet,
		); err != nil {
			return err
		}
	}
	return nil
}

// validateSingleState checks connectivity rules for a single state
func validateSingleState(
	state gonfa.State,
	counter *stateCounter,
	initialState gonfa.State,
	finalSet stateSet,
) error {
	isFinal := finalSet.contains(state)

	// Check for hanging states
	if counter.incoming == 0 && state != initialState {
		return fmt.Errorf(
			"state '%s' isn't an initial state but has no incoming transitions",
			state)
	}

	// Check for dead-end non-final states
	if counter.outgoing == 0 && !isFinal {
		return fmt.Errorf("state '%s' is a dead-end state", state)
	}

	// Check for final states with outgoing transitions
	if isFinal && counter.outgoing > 0 {
		return fmt.Errorf(
			"final state '%s' has outgoing transition(s)",
			state)
	}

	return nil
}

// validateFinalStateReachability checks if all final states are reachable
func validateFinalStateReachability(
	finalSet stateSet,
	reachable stateSet,
) error {
	for state := range finalSet {
		if !reachable.contains(state) {
			return fmt.Errorf(
				"final state '%s' is not reachable from initial state",
				state)
		}
	}
	return nil
}
