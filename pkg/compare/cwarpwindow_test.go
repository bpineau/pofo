package compare

import (
	"math"
	"strings"
	"testing"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// A benchmark that starts after the assets used to be aligned over the
// assets' window: forward-filled zeros, an infinite return out of the last of
// them, and a CWARP score that was no longer a number. The solver then fell
// back to its starting point and the report announced "+0.0 CWARP" over equal
// weights. The window must be the one the benchmark shares.
func TestOptimizeCWARPWindowIncludesTheBenchmark(t *testing.T) {
	base := optFixture(1500)
	full := optSeries("BENCH", 1500, func(i int) float64 { return 0.0004 + 0.008*math.Sin(float64(i)/7) })
	late := &marketdata.Series{Symbol: "BENCH", Name: "BENCH", Currency: "EUR",
		Source: "index", Points: full.Points[1000:]}

	got, note, err := optimizedPortfolio(base, optSpec(t, "cwarp"), late)
	if err != nil {
		t.Fatal(err)
	}
	// The first return of the fitted window is the move INTO the day after
	// the benchmark's first quote, so that is the date the note opens with.
	if want := late.First().Date.AddDate(0, 0, 1).Format("2006-01-02"); !strings.Contains(note, want) {
		t.Errorf("the fitting window must open with the benchmark (%s): %s", want, note)
	}
	// The solver must have had a finite score to work with, so it cannot have
	// stayed on the equal-weight point it starts from.
	if math.Abs(got.Assets[0].Weight-0.5) < 1e-6 {
		t.Errorf("weights unmoved (%v): the objective was undefined", got.Assets[0].Weight)
	}
	for i, a := range got.Assets {
		if math.IsNaN(a.Weight) || math.IsInf(a.Weight, 0) {
			t.Errorf("asset %d weight %v", i, a.Weight)
		}
	}

	// A benchmark that shares nothing with the assets is an error, not a
	// silently empty window.
	none := &marketdata.Series{Symbol: "BENCH", Points: optSeries("B", 40, func(int) float64 { return 0.001 }).Points}
	for i := range none.Points {
		none.Points[i].Date = none.Points[i].Date.AddDate(-10, 0, 0)
	}
	if _, _, err := optimizedPortfolio(base, optSpec(t, "cwarp"), none); err == nil {
		t.Error("expected an error when the benchmark shares no period with the assets")
	}
}
