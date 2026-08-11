package compare

import (
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/optimize"
)

// dailyDates returns n daily dates from 2000-01-03.
func dailyDates(n int) []time.Time {
	out := make([]time.Time, n)
	d := time.Date(2000, 1, 3, 0, 0, 0, 0, time.UTC)
	for i := range out {
		out[i] = d
		d = d.AddDate(0, 0, 1)
	}
	return out
}

// TestTrainSpanSelectsTheWindow: the fitting slice covers exactly the dates
// the window names, both ends included.
func TestTrainSpanSelectsTheWindow(t *testing.T) {
	dates := dailyDates(3000) // 2000-01-03 → 2008-03
	spec, err := optimize.ParseSpec("max-sharpe,train:2002..2005")
	if err != nil {
		t.Fatal(err)
	}
	got, err := trainSpan(dates, spec.Train)
	if err != nil {
		t.Fatal(err)
	}
	if y := dates[got.from].Year(); y != 2002 {
		t.Fatalf("slice starts in %d, want 2002", y)
	}
	if y := dates[got.to-1].Year(); y != 2005 {
		t.Fatalf("slice ends in %d, want 2005", y)
	}
	if dates[got.from-1].Year() != 2001 || dates[got.to].Year() != 2006 {
		t.Fatal("the slice leaks outside the window")
	}
}

// TestTrainSpanRefusesShortWindows: a window too short to fit anything is an
// error, not a silently overfitted answer.
func TestTrainSpanRefusesShortWindows(t *testing.T) {
	dates := dailyDates(3000)
	spec, err := optimize.ParseSpec("max-sharpe,train:2002..2002")
	if err != nil {
		t.Fatal(err)
	}
	_, err = trainSpan(dates, spec.Train)
	if err == nil || !strings.Contains(err.Error(), "two years") {
		t.Fatalf("a one-year window must be refused, got %v", err)
	}

	spec, err = optimize.ParseSpec("max-sharpe,train:1980..1985")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := trainSpan(dates, spec.Train); err == nil {
		t.Fatal("a window with no quotes in it must be refused")
	}
}

// TestLongestOutside: the holdout is the longer of the two untouched
// stretches, and it is empty when the fit covers everything.
func TestLongestOutside(t *testing.T) {
	if got := longestOutside(1000, span{0, 800}); got != (span{800, 1000}) {
		t.Fatalf("suffix holdout = %v", got)
	}
	if got := longestOutside(1000, span{300, 1000}); got != (span{0, 300}) {
		t.Fatalf("prefix holdout = %v", got)
	}
	// Fit in the middle: the longer side wins, no gluing across the hole.
	if got := longestOutside(1000, span{100, 400}); got != (span{400, 1000}) {
		t.Fatalf("middle fit holdout = %v", got)
	}
	if got := longestOutside(1000, span{0, 1000}); got.len() != 0 {
		t.Fatalf("a full-window fit leaves no holdout, got %v", got)
	}
}

// TestSpanSlicing: a span cuts every asset's returns to the same range.
func TestSpanSlicing(t *testing.T) {
	returns := [][]float64{{1, 2, 3, 4, 5}, {6, 7, 8, 9, 10}}
	got := span{1, 4}.of(returns)
	if len(got) != 2 || len(got[0]) != 3 || got[0][0] != 2 || got[1][2] != 9 {
		t.Fatalf("sliced returns = %v", got)
	}
	if (span{1, 4}).cut(nil) != nil {
		t.Fatal("a nil series must stay nil (CWARP without a benchmark)")
	}
}

// TestObjectiveLabel: the note names the limits a solve ran under, so the
// reader knows what was asked, not only what came back.
func TestObjectiveLabel(t *testing.T) {
	spec, err := optimize.ParseSpec("max-return,max-vol:9.5,max-drawdown:20")
	if err != nil {
		t.Fatal(err)
	}
	got := objectiveLabel(spec)
	if !strings.Contains(got, "vol ≤ 9.5 %") || !strings.Contains(got, "drawdown ≤ 20 %") {
		t.Fatalf("label = %q", got)
	}
	plain, err := optimize.ParseSpec("max-sharpe")
	if err != nil {
		t.Fatal(err)
	}
	if got := objectiveLabel(plain); got != "max-sharpe" {
		t.Fatalf("unconstrained label = %q", got)
	}
}
