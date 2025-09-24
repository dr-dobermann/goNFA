package machine

import (
	"context"

	"github.com/dr-dobermann/gonfa/pkg/gonfa"
)

// Test implementations
type testStateExtender struct {
	data string
}

type testGuard struct {
	result bool
	calls  int
}

func (g *testGuard) Check(
	ctx context.Context,
	state gonfa.MachineState,
	payload gonfa.Payload,
) bool {
	g.calls++
	return g.result
}

type testAction struct {
	name     string
	executed bool
	err      error
	calls    int
}

func (a *testAction) Execute(
	ctx context.Context,
	state gonfa.MachineState,
	payload gonfa.Payload,
) error {
	a.calls++
	a.executed = true
	return a.err
}
