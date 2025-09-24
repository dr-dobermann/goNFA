package definition

import (
	"context"

	"github.com/dr-dobermann/gonfa/pkg/gonfa"
)

// testGuard is a simple implementation of gonfa.Guard for testing.
type testGuard struct {
	result bool
}

func (g *testGuard) Check(
	ctx context.Context,
	state gonfa.MachineState,
	payload gonfa.Payload,
) bool {
	return g.result
}

// testAction is a simple implementation of gonfa.Action for testing.
type testAction struct {
	name     string
	executed bool
	err      error
}

func (a *testAction) Execute(
	ctx context.Context,
	state gonfa.MachineState,
	payload gonfa.Payload,
) error {
	a.executed = true
	return a.err
}
