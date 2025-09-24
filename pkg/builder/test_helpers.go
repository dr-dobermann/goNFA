package builder

import (
	"context"

	"github.com/dr-dobermann/gonfa/pkg/gonfa"
)

// Test implementations
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

type testAction struct {
	name     string
	executed bool
}

func (a *testAction) Execute(
	ctx context.Context,
	state gonfa.MachineState,
	payload gonfa.Payload,
) error {
	a.executed = true
	return nil
}
