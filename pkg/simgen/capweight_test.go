package simgen

import (
	"math"
	"testing"
	"time"
)

// frameOf builds a two-leg frame of n daily steps from constant daily returns.
func frameOf(a, b float64, n int) *Frame {
	fr := &Frame{Returns: map[string][]float64{"A": make([]float64, n), "B": make([]float64, n)}}
	for k := 0; k < n; k++ {
		fr.Dates = append(fr.Dates, time.Date(2000, 1, 3, 0, 0, 0, 0, time.UTC).AddDate(0, 0, k))
		if k > 0 {
			fr.Returns["A"][k] = a
			fr.Returns["B"][k] = b
		}
	}
	return fr
}

func TestCapWeightedIsBuyAndHold(t *testing.T) {
	const n = 400
	fr := frameOf(0.001, -0.0005, n) // A wins every day, B loses
	legs := []Leg{{ID: "A", Weight: 0.4}, {ID: "B", Weight: 0.6}}
	anchor := fr.Dates[200]
	values, err := CapWeighted(fr, legs, anchor, 0)
	if err != nil {
		t.Fatalf("CapWeighted: %v", err)
	}
	// Holding the anchor's shares: the level is the basket's own value, so the
	// whole path is 0.4×A + 0.6×B in the legs' own units, base 100 at step 0.
	la := func(k int) float64 { return math.Pow(1.001, float64(k-200)) }
	lb := func(k int) float64 { return math.Pow(0.9995, float64(k-200)) }
	base := 0.4*la(0) + 0.6*lb(0)
	for _, k := range []int{1, 100, 200, 399} {
		want := 100 * (0.4*la(k) + 0.6*lb(k)) / base
		if math.Abs(values[k]/want-1) > 1e-9 {
			t.Errorf("step %d: %.6f, want %.6f (the buy-and-hold basket)", k, values[k], want)
		}
	}
	// At the anchor the weights ARE the published split; the winner's weight
	// grows after it and shrinks before it, which is the whole point.
	at := capWeights(fr, legs, anchor, 200)
	if math.Abs(at["A"]-0.4) > 1e-12 {
		t.Errorf("anchor weight of A = %.6f, want the published 0.40", at["A"])
	}
	before, after := capWeights(fr, legs, anchor, 0)["A"], capWeights(fr, legs, anchor, 399)["A"]
	if !(before < 0.4 && 0.4 < after) {
		t.Errorf("A's weight reads %.4f before the anchor and %.4f after: it must drift up with its own returns", before, after)
	}
	if s := before + capWeights(fr, legs, anchor, 0)["B"]; math.Abs(s-1) > 1e-12 {
		t.Errorf("weights sum to %.6f, want 1", s)
	}
}

// A single leg has no weights to drift, so the blend must return that leg's own
// path: the check that no rebalancing effect creeps in.
func TestCapWeightedSingleLegIsTheLeg(t *testing.T) {
	fr := frameOf(0.0007, 0, 300)
	values, err := CapWeighted(fr, []Leg{{ID: "A", Weight: 1}}, fr.Dates[150], 0)
	if err != nil {
		t.Fatalf("CapWeighted: %v", err)
	}
	if want := 100 * math.Pow(1.0007, 299); math.Abs(values[299]/want-1) > 1e-9 {
		t.Errorf("last level %.4f, want the leg's own %.4f", values[299], want)
	}
}

func TestCapWeightedFeeAndRenormalization(t *testing.T) {
	fr := frameOf(0.0004, 0.0004, 300)
	legs := []Leg{{ID: "A", Weight: 40}, {ID: "B", Weight: 60}} // percentages, not fractions
	free, err := CapWeighted(fr, legs, fr.Dates[0], 0)
	if err != nil {
		t.Fatalf("CapWeighted: %v", err)
	}
	if want := 100 * math.Pow(1.0004, 299); math.Abs(free[299]/want-1) > 1e-9 {
		t.Errorf("unnormalized weights changed the level: %.4f, want %.4f", free[299], want)
	}
	charged, err := CapWeighted(fr, legs, fr.Dates[0], 0.0252) // 0.0001/day
	if err != nil {
		t.Fatalf("CapWeighted: %v", err)
	}
	if want := 100 * math.Pow(1.0003, 299); math.Abs(charged[299]/want-1) > 1e-9 {
		t.Errorf("after fee %.4f, want %.4f (0.0252/252 a day)", charged[299], want)
	}
}

func TestCapWeightedRefuses(t *testing.T) {
	fr := frameOf(0.001, 0.001, 100)
	anchor := fr.Dates[50]
	cases := map[string][]Leg{
		"missing component": {{ID: "A", Weight: 0.5}, {ID: "MISSING", Weight: 0.5}},
		"excess leg":        {{ID: "A", Weight: 0.5}, {ID: "B", Weight: 0.5, Excess: true}},
		"non-positive":      {{ID: "A", Weight: 0.5}, {ID: "B", Weight: 0}},
		"no leg":            {},
	}
	for name, legs := range cases {
		if _, err := CapWeighted(fr, legs, anchor, 0); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

func TestAnchorIndex(t *testing.T) {
	fr := frameOf(0, 0, 10)
	if got := anchorIndex(fr.Dates, fr.Dates[4].Add(12*time.Hour)); got != 4 {
		t.Errorf("mid-frame anchor: %d, want 4 (the last date on or before it)", got)
	}
	if got := anchorIndex(fr.Dates, fr.Dates[0].AddDate(-5, 0, 0)); got != 0 {
		t.Errorf("anchor before the frame: %d, want 0", got)
	}
	if got := anchorIndex(fr.Dates, fr.Dates[9].AddDate(5, 0, 0)); got != 9 {
		t.Errorf("anchor after the frame: %d, want the last step 9", got)
	}
}

// The published anchor split is what the recipes claim it is, and it is a
// proper distribution: the one thing a reader of worldMethod checks.
func TestWorldLegsMatchThePublishedSplit(t *testing.T) {
	want := map[string]float64{"VFINX": 0.404, "VTMGX": 0.461, "VEIEX": 0.135}
	var sum float64
	for _, l := range worldLegs {
		if w, ok := want[l.ID]; !ok || math.Abs(l.Weight-w) > 1e-12 {
			t.Errorf("leg %s at %v, want %v (the 2009-10-31 FTSE All-World split)", l.ID, l.Weight, w)
		}
		sum += l.Weight
	}
	if math.Abs(sum-1) > 1e-12 {
		t.Errorf("the split sums to %v, want 1", sum)
	}
	if worldAnchor.Year() != 2009 || worldAnchor.Month() != time.October {
		t.Errorf("anchor %s, want the published report's month end", worldAnchor.Format("2006-01-02"))
	}
}
