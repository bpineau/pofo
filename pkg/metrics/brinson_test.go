package metrics

import (
	"math"
	"testing"
)

func TestBrinsonHandCalculated(t *testing.T) {
	got, err := Brinson([]float64{.6, .4}, []float64{.5, .5}, []float64{.12, .02}, []float64{.1, .04})
	if err != nil {
		t.Fatal(err)
	}
	// Rp=.08, Rb=.07; allocation=.003+.003, selection=.01-.01,
	// interaction=.002+.002. The active return is .01.
	for _, tc := range []struct{ got, want float64 }{{got.PortfolioReturn, .08}, {got.BenchmarkReturn, .07}, {got.ActiveReturn, .01}, {got.Allocation, .006}, {got.Selection, 0}, {got.Interaction, .004}} {
		if math.Abs(tc.got-tc.want) > 1e-15 {
			t.Fatalf("got %.16g, want %.16g", tc.got, tc.want)
		}
	}
	total := 0.0
	for _, s := range got.Sectors {
		total += s.Allocation + s.Selection + s.Interaction
	}
	if math.Abs(total-got.ActiveReturn) > 1e-15 {
		t.Fatal("sector effects do not reconcile")
	}
}

func TestBrinsonZeroWeightSectors(t *testing.T) {
	for _, weights := range []struct{ wp, wb []float64 }{{[]float64{1, 0}, []float64{0, 1}}, {[]float64{0, 1}, []float64{1, 0}}} {
		got, err := Brinson(weights.wp, weights.wb, []float64{.2, .1}, []float64{.03, .07})
		if err != nil {
			t.Fatal(err)
		}
		if math.Abs(got.Allocation+got.Selection+got.Interaction-got.ActiveReturn) > 1e-15 {
			t.Fatalf("zero weight did not reconcile: %+v", got)
		}
	}
}

func TestBrinsonInvalid(t *testing.T) {
	for _, bad := range [][]float64{nil, {1}, {-.1, 1.1}, {.4, .4}, {math.NaN(), .5}, {math.Inf(1), 0}} {
		if _, err := Brinson(bad, []float64{.5, .5}, []float64{.1, .2}, []float64{.1, .2}); err == nil {
			t.Fatalf("accepted weights %v", bad)
		}
	}
	if _, err := Brinson([]float64{.5, .5}, []float64{.5, .5}, []float64{math.NaN(), .1}, []float64{.1, .2}); err == nil {
		t.Fatal("accepted NaN sector return")
	}
}
