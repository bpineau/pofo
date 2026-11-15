package fire

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func ExampleAcquire() {
	release, err := Acquire(context.Background())
	if err != nil {
		panic(err)
	}
	defer release()
	fmt.Println("FIRE computation admitted")
	// Output: FIRE computation admitted
}

func TestAcquireWaitsUntilCancelledOrReleased(t *testing.T) {
	var releases []func()
	for range cap(simulationSlots) {
		release, err := Acquire(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		releases = append(releases, release)
	}
	defer func() {
		for _, release := range releases {
			release()
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if _, err := Acquire(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("wait: %v", err)
	}
	releases[0]()
	releases = releases[1:]
	release, err := Acquire(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	release()
	if _, err := Acquire(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expired caller admitted: %v", err)
	}
}
