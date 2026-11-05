package main

import (
	"context"
	"strings"
	"testing"
)

// Without a portfolio the FIRE explorer runs on its parametric models alone:
// no fetch, no panel, no labels.
func TestFirePanelWithoutAPortfolio(t *testing.T) {
	b := firePanel(context.Background(), &options{}, nil, nil)
	if b.panel != nil || b.labels != nil || b.notes != nil {
		t.Errorf("a portfolio-less run built a panel (%v), labels (%v) or notes (%q)", b.panel, b.labels, b.notes)
	}
}

// The FIRE panel is a fully invested mix of the holdings: what it cannot
// carry of the written portfolio is stated, never dropped in silence.
func TestSpecNotesNameWhatThePanelCannotCarry(t *testing.T) {
	spec, err := adhocSpec("SP500:90,DTLA:60!leverage:on", 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	notes := specNotes(spec)
	if len(notes) != 1 || !strings.Contains(notes[0], "150 % invested") {
		t.Errorf("leveraged 90/60: notes %q, want the 150 %% read as 100 %%", notes)
	}
	plain, err := adhocSpec("SP500:60,DTLA:40", 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	if notes := specNotes(plain); len(notes) != 0 {
		t.Errorf("a plain 60/40: notes %q, want none", notes)
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
