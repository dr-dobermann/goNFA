package definition

import (
	"fmt"
	"slices"

	"github.com/dr-dobermann/gonfa/pkg/gonfa"
)

// checkStates checks final state presence in states and if they
// are reachable.
func checkStates(
	initialState gonfa.State,
	states []gonfa.State,
	transitions []Transition,
	finalStates []gonfa.State,
) error {
	// Check if initial state is in states
	if !slices.Contains(states, initialState) {
		return fmt.Errorf("initial state '%s' doesn't exist in states", initialState)
	}

	// Check if all final states are present in states
	for _, fs := range finalStates {
		if !slices.Contains(states, fs) {
			return fmt.Errorf("final state '%s' isn't found in states", fs)
		}
	}

	// Build state graph for states and transitions analysis
	tss := map[gonfa.State][]gonfa.State{initialState: {}}
	for _, t := range transitions {
		if !slices.Contains(states, t.From) {
			return fmt.Errorf("state '%s' doesn't exist as transition source",
				t.From)
		}

		if !slices.Contains(states, t.To) {
			return fmt.Errorf("state '%s' doesn't exist as transition target", t.To)
		}

		tt, found := tss[t.From]
		if !found {
			tss[t.From] = []gonfa.State{t.To}

			continue
		}

		if slices.Contains(tt, t.To) {
			return fmt.Errorf("duplicate transition from '%s' to '%s'",
				t.From, t.To)
		}

		tss[t.From] = append(tt, t.To)

	}

	return analyzeGraph(initialState, finalStates, states, tss)
}

// transCounter holds number of input and output transitions from the state.
type transCounter struct {
	// traversed indicates whether this state was tracked in previous passes
	traversed bool
	in        int
	out       int
}

// analyzeGraph checks reachability of final states, "hanging" and "deadend" states
func analyzeGraph(
	initialState gonfa.State,
	finalStates []gonfa.State,
	states []gonfa.State,
	tss map[gonfa.State][]gonfa.State,
) error {
	// Check if initial state is really used as initial state
	if _, found := tss[initialState]; !found {
		return fmt.Errorf("no transitions start from initial state '%s'", initialState)
	}

	stc := make(map[gonfa.State]transCounter, len(states))

	// Trace state transitions from initial state
	us := []gonfa.State{initialState}
	stc[initialState] = transCounter{}
	for len(us) > 0 {
		s := us[0]

		// Add unchecked states from traverse
		us = append(us, traverse(s, tss, stc)...)

		// Remove s as checked
		us = us[1:]
	}

	// To check if the final state is reachable,
	// all the final states are marked as unreachable and put to urfs.
	// If state has no outgoing transitions and it is presented in final states list
	// the state is removed from urfs.
	// If urfs isn't empty after check then the definition is broken since it has
	// unreachable final states.
	urfs := make([]gonfa.State, len(finalStates))
	n := copy(urfs, finalStates)
	if n != len(finalStates) {
		panic("copying finalStates to unreachable final states list failed")
	}

	for s, sc := range stc {
		// Check for hanging states -- the states, which have no incoming transitions but
		// aren't initialState.
		if sc.in == 0 && s != initialState {
			return fmt.Errorf(
				"state '%s' isn't an initial state but has no incoming transitions", s)
		}

		isFinal := slices.Contains(finalStates, s)

		// Check for deadend states -- the states, which have no outgoing transitions but
		// aren't one of the final states.
		if isFinal && sc.out == 0 {
			return fmt.Errorf("state '%s' is a deadend state", s)
		}

		// Check for invalid final states (has outgoing transitions)
		if isFinal && sc.out > 0 {
			return fmt.Errorf("final state '%s' has outgoing transition(s)", s)
		}

		// Delete final state from the unreachable final states list if it has incoming
		// transition.
		if isFinal && sc.in > 0 {
			urfs = slices.DeleteFunc(
				urfs,
				func(cs gonfa.State) bool {
					return cs == s
				})
		}
	}

	if len(urfs) > 0 {
		return fmt.Errorf("definition has unreachable final state(s): %v", urfs)
	}

	return nil
}

// traverse traces single transition path from state s to the end state and
// returns all possible state's forks on the way.
// traverse checks the repeatable passes of the states and updates input and
// output transitions of the state.
func traverse(
	s gonfa.State,
	tss map[gonfa.State][]gonfa.State,
	stc map[gonfa.State]transCounter,
) []gonfa.State {
	ss := []gonfa.State{}

	si, ok := stc[s]
	if !ok {
		panic(fmt.Sprintf("failed to get stateCounter for state '%s'", s))
	}

	si.in += 1

	for _, ts := range tss[s] {
		si.out += 1

		// Find target state in state counters map
		tsc, found := stc[ts]
		if !found {
			tsc = transCounter{}
			stc[ts] = tsc
		}

		if !tsc.traversed {
			ss = append(ss, ts)
		}
	}

	si.traversed = true

	return ss
}
