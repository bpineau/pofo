package compare

import (
	"math"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/portfolio"
)

// sweepFixture builds a two-holding portfolio: a grower that compounds
// steadily and a flat line, over 800 daily points. Any weight moved from the
// grower to the flat line must cost return.
func sweepFixture(t *testing.T, weights ...float64) *portfolio.Portfolio {
	t.Helper()
	const n = 800
	d := time.Date(2000, 1, 3, 0, 0, 0, 0, time.UTC)
	grower := make([]marketdata.Point, n)
	flat := make([]marketdata.Point, n)
	for i := 0; i < n; i++ {
		grower[i] = marketdata.Point{Date: d, Close: 100 * math.Pow(1.0004, float64(i))}
		flat[i] = marketdata.Point{Date: d, Close: 100}
		d = d.AddDate(0, 0, 1)
	}
	p := &portfolio.Portfolio{Name: "fixture"}
	for i, w := range weights {
		pts, sym := grower, "GROW"
		if i == 1 {
			pts, sym = flat, "FLAT"
		}
		p.Assets = append(p.Assets, portfolio.Asset{
			ID: sym, Symbol: sym, Weight: w, Fees: -1,
			Series: &marketdata.Series{Symbol: sym, Currency: "EUR", Points: pts},
		})
	}
	return p
}

// TestSweepShape: one row per grid point plus the written weight, flagged
// once, and the written row reproduces the portfolio as written.
func TestSweepShape(t *testing.T) {
	p := sweepFixture(t, 0.62, 0.38)
	out, err := Sweep(p, SweepOptions{Weights: []float64{0, 0.5, 1}, RebalanceDays: 90})
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 2 {
		t.Fatalf("got %d holdings, want 2", len(out))
	}
	for _, h := range out {
		written := 0
		for _, pt := range h.Points {
			if pt.Written {
				written++
				if math.Abs(pt.Weight-h.Written) > 1e-9 {
					t.Fatalf("%s: flagged row at %.3f, written weight is %.3f", h.ID, pt.Weight, h.Written)
				}
			}
		}
		if written != 1 {
			t.Fatalf("%s: %d rows flagged as written, want exactly 1", h.ID, written)
		}
		if len(h.Points) != 4 {
			t.Fatalf("%s: %d rows, want 3 grid points + the written weight", h.ID, len(h.Points))
		}
		// The grid must stay sorted, written weight inserted in place.
		for i := 1; i < len(h.Points); i++ {
			if h.Points[i].Weight <= h.Points[i-1].Weight {
				t.Fatalf("%s: grid out of order at %d", h.ID, i)
			}
		}
	}
}

// TestSweepMonotoneOnFixture: moving weight into the flat line can only cost
// CAGR, and the sweep says so.
func TestSweepMonotoneOnFixture(t *testing.T) {
	p := sweepFixture(t, 0.5, 0.5)
	out, err := Sweep(p, SweepOptions{Weights: []float64{0, 0.25, 0.5, 0.75}, RebalanceDays: 0})
	if err != nil {
		t.Fatal(err)
	}
	flat := out[1]
	for i := 1; i < len(flat.Points); i++ {
		if flat.Points[i].Stats.CAGR >= flat.Points[i-1].Stats.CAGR {
			t.Fatalf("CAGR did not fall from %.1f %% to %.1f %% weight in the flat line",
				flat.Points[i-1].Weight*100, flat.Points[i].Weight*100)
		}
	}
}

// TestSweepKeepsTheWindow: every grid point is measured over the same period,
// including the ones where a holding sits at zero (it stays in the portfolio,
// so it keeps binding the common window and the rows stay comparable).
func TestSweepKeepsTheWindow(t *testing.T) {
	p := sweepFixture(t, 0.5, 0.5)
	// Cut the second holding's history short: it must bind every row's start.
	p.Assets[1].Series.Points = p.Assets[1].Series.Points[100:]
	out, err := Sweep(p, SweepOptions{Weights: []float64{0, 0.5}, RebalanceDays: 90})
	if err != nil {
		t.Fatal(err)
	}
	want := out[0].Points[0].Stats.Start
	for _, h := range out {
		for _, pt := range h.Points {
			if !pt.Stats.Start.Equal(want) {
				t.Fatalf("%s at %.0f %%: window starts %s, want %s",
					h.ID, pt.Weight*100, pt.Stats.Start, want)
			}
		}
	}
}

// TestSweepNeedsTwoHoldings: a single line has nothing to renormalize onto.
func TestSweepNeedsTwoHoldings(t *testing.T) {
	if _, err := Sweep(sweepFixture(t, 1), SweepOptions{}); err == nil {
		t.Fatal("a one-line portfolio must be rejected")
	}
}

// TestSweepChargesFinancing: a levered book borrows, and the borrowed money
// costs the cash rate plus the spread. Sweeping must carry that financing
// into every grid point, otherwise the whole table describes a portfolio that
// leveraged itself for free (the trap cmd/pofo's -sweep fell into by building
// its portfolio without BuildOptions.Cash).
func TestSweepChargesFinancing(t *testing.T) {
	sweepCAGR := func(cash *marketdata.Series) float64 {
		t.Helper()
		p := sweepFixture(t, 0.9, 0.6)
		p.Leverage, p.BorrowSpread, p.Cash = true, 1.0, cash
		out, err := Sweep(p, SweepOptions{Weights: []float64{0.6}, RebalanceDays: 90})
		if err != nil {
			t.Fatal(err)
		}
		return out[0].Points[0].Stats.CAGR
	}
	// A flat 5 %/yr financing rate on the same calendar as the holdings.
	src := sweepFixture(t, 1).Assets[0].Series
	rate := &marketdata.Series{Symbol: "^IRX", Points: make([]marketdata.Point, len(src.Points))}
	for i, pt := range src.Points {
		rate.Points[i] = marketdata.Point{Date: pt.Date, Close: 5}
	}
	free, financed := sweepCAGR(nil), sweepCAGR(rate)
	if !(financed < free-0.005) {
		t.Fatalf("financed CAGR %.4f vs free %.4f: the cash rate did not reach the sweep", financed, free)
	}
}

// TestReweightedProportions: the other holdings keep their relative
// proportions and the weights still sum to one.
func TestReweightedProportions(t *testing.T) {
	p := &portfolio.Portfolio{Assets: []portfolio.Asset{
		{ID: "A", Weight: 0.5}, {ID: "B", Weight: 0.3}, {ID: "C", Weight: 0.2},
	}}
	got, ok := reweighted(p, 0, 0.1)
	if !ok {
		t.Fatal("reweight refused a feasible move")
	}
	sum := 0.0
	for _, a := range got.Assets {
		sum += a.Weight
	}
	if math.Abs(sum-1) > 1e-12 {
		t.Fatalf("weights sum to %.6f", sum)
	}
	// B:C was 3:2 and must stay 3:2.
	if ratio := got.Assets[1].Weight / got.Assets[2].Weight; math.Abs(ratio-1.5) > 1e-12 {
		t.Fatalf("B:C ratio = %.4f, want 1.5", ratio)
	}
	if math.Abs(got.Assets[0].Weight-0.1) > 1e-12 {
		t.Fatalf("swept weight = %.4f, want 0.1", got.Assets[0].Weight)
	}
	if p.Assets[0].Weight != 0.5 {
		t.Fatal("reweight mutated the source portfolio")
	}
}

// TestReweightedKeepsLeverage: a levered book's total weight is part of its
// meaning, so sweeping one line must not renormalize it back to 100 %.
func TestReweightedKeepsLeverage(t *testing.T) {
	p := &portfolio.Portfolio{Leverage: true, Assets: []portfolio.Asset{
		{ID: "A", Weight: 0.9}, {ID: "B", Weight: 0.6},
	}}
	got, ok := reweighted(p, 0, 0.5)
	if !ok {
		t.Fatal("reweight refused a feasible move")
	}
	sum := got.Assets[0].Weight + got.Assets[1].Weight
	if math.Abs(sum-1.5) > 1e-12 {
		t.Fatalf("total weight = %.4f, want the written 1.5", sum)
	}
	if _, ok := reweighted(p, 0, 1.6); ok {
		t.Fatal("a weight above the portfolio's total must be refused")
	}
}
