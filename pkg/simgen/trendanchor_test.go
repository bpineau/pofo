package simgen

import (
	"math"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// anchorFixture builds a monthly reference and a daily calendar to anchor onto:
// 60 month-end reference points with an alternating return, and every calendar
// day of the same span carrying a constant cash accrual.
func anchorFixture(refReturn func(i int) float64) (ref *marketdata.Series, dates []time.Time, values, cash []float64) {
	ref = &marketdata.Series{Symbol: "REF"}
	level := 100.0
	for i := range 61 {
		d := time.Date(2000, time.Month(1+i), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, -1)
		if i > 0 {
			level *= 1 + refReturn(i)
		}
		ref.Points = append(ref.Points, marketdata.Point{Date: d, Close: level})
	}
	const dailyCash = 0.0002 // ~7.6 %/yr, big enough that a second one would show
	for d := ref.First().Date; !d.After(ref.Last().Date); d = d.AddDate(0, 0, 1) {
		dates = append(dates, d)
		cash = append(cash, dailyCash)
		values = append(values, 100) // a flat engine: the anchor supplies everything
	}
	return ref, dates, values, cash
}

// monthReturns folds an anchored output back into the month-end returns it
// realized, which is what the anchor is supposed to control.
func monthReturns(dates []time.Time, out []float64) map[string]float64 {
	last := map[string]int{}
	var months []string
	for i, d := range dates {
		k := d.Format("2006-01")
		if _, seen := last[k]; !seen {
			months = append(months, k)
		}
		last[k] = i
	}
	got := map[string]float64{}
	for i := 1; i < len(months); i++ {
		a, b := last[months[i-1]], last[months[i]]
		got[months[i]] = out[b]/out[a] - 1
	}
	return got
}

// A funded reference already earns cash on its collateral. Rescaled at 1x and
// funded again, a funded anchor must reproduce the reference month for month:
// the cash leg it carries is stripped and put back, never counted twice. This
// is the failure the net managed-futures reference was shipped with, and it was
// worth six points a year in the 1990s.
func TestAnchorTrendFundedReferenceKeepsOneCashLeg(t *testing.T) {
	refReturn := func(i int) float64 {
		if i%2 == 0 {
			return 0.03
		}
		return -0.01
	}
	ref, dates, values, cash := anchorFixture(refReturn)
	f := fakeFetcher{"REF": ref}
	targetVol := monthlyVol(ref) // scale = 1: the output must BE the reference

	funded, err := AnchorTrend(f, TrendAnchor{ID: "REF", Funded: true}, dates, values, cash, targetVol, true)
	if err != nil {
		t.Fatal(err)
	}
	// The first anchored month only fixes the span's start, so the anchor's
	// control begins with the one after it.
	got := monthReturns(dates, funded)
	for i := 2; i < len(ref.Points); i++ {
		k := ref.Points[i].Date.Format("2006-01")
		want := ref.Points[i].Close/ref.Points[i-1].Close - 1
		if g, ok := got[k]; ok && math.Abs(g-want) > 1e-9 {
			t.Fatalf("%s: anchored return %.6f, want the reference's %.6f", k, g, want)
		}
	}

	// Read as an excess index instead, the same reference hands the output a
	// second cash leg: every month comes out richer by that month's cash.
	excess, err := AnchorTrend(f, TrendAnchor{ID: "REF"}, dates, values, cash, targetVol, true)
	if err != nil {
		t.Fatal(err)
	}
	a, b := funded[len(funded)-1]/funded[0], excess[len(excess)-1]/excess[0]
	years := dates[len(dates)-1].Sub(dates[0]).Hours() / 24 / 365.25
	gap := (math.Pow(b, 1/years) - math.Pow(a, 1/years)) * 100
	if gap < 7 || gap > 9 {
		t.Errorf("the double cash leg is worth %.2f points a year, expected the ~7.6 of the fixture's rate", gap)
	}
}

// The volatility target rescales the reference's EXCESS, never its cash: a book
// run twice as hot borrows to trade, it does not earn twice the T-bill.
func TestAnchorTrendScalesTheExcessOnly(t *testing.T) {
	ref, dates, values, cash := anchorFixture(func(i int) float64 {
		if i%2 == 0 {
			return 0.03
		}
		return -0.01
	})
	f := fakeFetcher{"REF": ref}
	anchor := TrendAnchor{ID: "REF", Funded: true}

	out, err := AnchorTrend(f, anchor, dates, values, cash, 2*monthlyVol(ref), true)
	if err != nil {
		t.Fatal(err)
	}
	got := monthReturns(dates, out)
	for i := 2; i < len(ref.Points); i++ {
		k := ref.Points[i].Date.Format("2006-01")
		g, ok := got[k]
		if !ok {
			continue
		}
		// The month's own cash factor, then twice the reference's excess.
		c := 1.0
		for j, d := range dates {
			if d.After(ref.Points[i-1].Date) && !d.After(ref.Points[i].Date) {
				c *= 1 + cash[j]
			}
		}
		refRet := ref.Points[i].Close/ref.Points[i-1].Close - 1
		want := (1+2*((1+refRet)/c-1))*c - 1
		if math.Abs(g-want) > 1e-9 {
			t.Fatalf("%s: doubled return %.6f, want %.6f", k, g, want)
		}
	}
}

// An overlay wants the excess itself: an unfunded reference, an unfunded
// output, and the cash slice ignored end to end.
func TestAnchorTrendExcessOverlay(t *testing.T) {
	ref, dates, values, cash := anchorFixture(func(i int) float64 {
		if i%2 == 0 {
			return 0.03
		}
		return -0.01
	})
	f := fakeFetcher{"REF": ref}
	out, err := AnchorTrend(f, TrendAnchor{ID: "REF"}, dates, values, cash, monthlyVol(ref), false)
	if err != nil {
		t.Fatal(err)
	}
	got := monthReturns(dates, out)
	for i := 2; i < len(ref.Points); i++ {
		k := ref.Points[i].Date.Format("2006-01")
		want := ref.Points[i].Close/ref.Points[i-1].Close - 1
		if g, ok := got[k]; ok && math.Abs(g-want) > 1e-9 {
			t.Errorf("%s: overlay returned %.6f, want the reference's own %.6f, no cash added", k, g, want)
		}
	}
}

func TestAnchorTrendRejectsMismatchedInputs(t *testing.T) {
	ref, dates, values, cash := anchorFixture(func(i int) float64 { return 0.01 })
	f := fakeFetcher{"REF": ref}
	if _, err := AnchorTrend(f, TrendAnchor{ID: "REF"}, dates, values[:10], cash, 0.1, true); err == nil {
		t.Error("accepted a values slice shorter than dates")
	}
	if _, err := AnchorTrend(f, TrendAnchor{ID: "MISSING"}, dates, values, cash, 0.1, true); err == nil {
		t.Error("accepted an unknown anchor")
	}
}
