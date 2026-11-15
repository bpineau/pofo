package fire

import (
	"context"
	"runtime"
)

var simulationSlots = make(chan struct{}, max(2, runtime.GOMAXPROCS(0)/2))

// Acquire waits for the process-wide FIRE computation capacity, shared by web,
// API and MCP handlers. It permits max(2, GOMAXPROCS/2) concurrent computations.
// A page submits many sections at once: waiting is cancellable, never rejected
// because a fixed waiting room is full. On success call the returned release
// function exactly once. On cancellation it returns the context's error.
func Acquire(ctx context.Context) (release func(), err error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	select {
	case simulationSlots <- struct{}{}:
		return func() { <-simulationSlots }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
