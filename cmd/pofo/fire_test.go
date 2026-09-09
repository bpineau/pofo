package main

import (
	"context"
	"strings"
	"testing"
)

// Without a portfolio the FIRE explorer runs on its parametric models alone:
// no fetch, no panel, no labels.
func TestFirePanelWithoutAPortfolio(t *testing.T) {
	panel, labels := firePanel(context.Background(), &options{}, nil, nil)
	if panel != nil || labels != nil {
		t.Errorf("a portfolio-less run built a panel (%v) and labels (%v)", panel, labels)
	}
}

// -fire serves until the context is canceled, then returns cleanly, like
// -serve. An already-canceled context exercises the shutdown path without a
// request, and no portfolio means nothing is fetched.
func TestRunFireShutsDownWithTheContext(t *testing.T) {
	quietLog(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	out := captureOutput(t, func() {
		if err := runFire(ctx, &options{noOpen: true}, nil, nil); err != nil {
			t.Errorf("runFire: %v", err)
		}
	})
	if !strings.Contains(out, "FIRE explorer on http://127.0.0.1:") {
		t.Errorf("-fire did not announce its address:\n%s", out)
	}
}
