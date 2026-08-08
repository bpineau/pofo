package simgen

import (
	"math"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// growing returns a level series over the given day indices compounding at a
// constant daily rate.
func growing(symbol string, days []int, daily float64) *marketdata.Series {
	s := &marketdata.Series{Symbol: symbol}
	v := 100.0
	for i, d := range days {
		if i > 0 {
			v *= 1 + daily
		}
		s.Points = append(s.Points, marketdata.Point{Date: day(d), Close: v})
	}
	return s
}

// The blend is a weighted average of the sleeves that quote, RENORMALIZED: with
// only one sleeve open the index must earn that sleeve's own return in full, not
// its share of it, since the fund always carries a full bond notional.
func TestBlendExcessRenormalizesOverTheSleevesThatQuote(t *testing.T) {
	all := []int{0, 1, 2, 3, 4}
	usd := growing("USD", all, 0.001)
	// The foreign sleeve only opens on day 2, and then loses what the US one
	// gains, so the blend must be pure USD before and blended after.
	foreign := growing("FOR", []int{2, 3, 4}, -0.001)

	sleeves := []bondSleeve{{name: "USD", weight: 0.80}, {name: "FOR", weight: 0.20}}
	got := blendExcess("blend", usd.Points, sleeves, []*marketdata.Series{usd, foreign})

	if len(got.Points) != len(all) {
		t.Fatalf("%d points, want %d", len(got.Points), len(all))
	}
	// Days 1 and 2: the foreign sleeve has no return yet (day 2 is its first
	// quote, so it has no previous level either), the blend is all USD.
	near(t, "day 1", got.Points[1].Close/got.Points[0].Close-1, 0.001, 1e-12)
	near(t, "day 2", got.Points[2].Close/got.Points[1].Close-1, 0.001, 1e-12)
	// Day 3 onwards: 0.8x(+0.1%) + 0.2x(-0.1%), the weights summing to 1.
	near(t, "day 3", got.Points[3].Close/got.Points[2].Close-1, 0.8*0.001-0.2*0.001, 1e-12)
}

// A sleeve on a coarser calendar than the blend's is read by forward fill: it
// contributes nothing between two of its own quotes and its whole step at the
// next one, which is what a monthly series honestly says.
func TestBlendExcessForwardFillsACoarserSleeve(t *testing.T) {
	all := []int{0, 1, 2, 3}
	usd := growing("USD", all, 0)
	monthly := &marketdata.Series{Symbol: "GBP", Points: []marketdata.Point{
		{Date: day(0), Close: 100}, {Date: day(3), Close: 110},
	}}
	sleeves := []bondSleeve{{name: "USD", weight: 0.5}, {name: "GBP", weight: 0.5}}
	got := blendExcess("blend", usd.Points, sleeves, []*marketdata.Series{usd, monthly})

	near(t, "day 1", got.Points[1].Close/got.Points[0].Close-1, 0, 1e-12)
	near(t, "day 3", got.Points[3].Close/got.Points[2].Close-1, 0.5*0.10, 1e-12)
	near(t, "whole span", got.Last().Close/100-1, 0.5*0.10, 1e-12)
}

// A local sleeve earns its bond LESS its own currency's financing, and nothing
// else: no FX and no carry, because the fund rolls the currency exposure away
// with forwards whose points are exactly that interest differential.
func TestLocalExcessNetsTheLocalFinancing(t *testing.T) {
	days := []int{0, 1, 2, 3, 4}
	bond := growing("BUND", days, 0.0020)
	cash := growing("CASH", days, 0.0005)
	got, err := localExcess("german", bond, cash, day(0))
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i < len(got.Points); i++ {
		near(t, "excess step", got.Points[i].Close/got.Points[i-1].Close-1, 0.0015, 1e-12)
	}
}

// The basket weights are the fund's audited country notionals and must stay a
// full notional: anything else silently levers or de-levers the bond sleeve.
func TestBondBasketWeightsSumToOne(t *testing.T) {
	sum := usdBondShare + deuBondShare + jpnBondShare + gbrBondShare
	if math.Abs(sum-1) > 1e-12 {
		t.Errorf("basket weights sum to %v, want 1", sum)
	}
}

// The overlay must survive a missing foreign reference (renormalizing over what
// is left) and must NOT survive a missing US one, which is four fifths of it.
func TestGlobalBondOverlayToleratesAMissingForeignSleeve(t *testing.T) {
	days := []int{0, 1, 2, 3, 4, 5}
	f := fakeFetcher{
		"VFITX":       growing("VFITX", days, 0.0010),
		"VUSTX":       growing("VUSTX", days, 0.0010),
		"^IRX":        levelsOn("^IRX", days, 2.52), // 0.01 %/day
		"BUND-EUR":    growing("BUND-EUR", days, 0.0020),
		"EURCASH-EUR": growing("EURCASH-EUR", days, 0.0005),
		"JGB-JPY":     growing("JGB-JPY", days, 0.0020),
		"JPCASH-JPY":  growing("JPCASH-JPY", days, 0.0005),
		"GILT-GBP":    growing("GILT-GBP", days, 0.0020),
		"GBCASH-GBP":  growing("GBCASH-GBP", days, 0.0005),
	}
	if _, err := globalBondOverlay(f, day(0)); err != nil {
		t.Fatalf("full basket: %v", err)
	}
	delete(f, "GILT-GBP") // the British reference goes missing
	if _, err := globalBondOverlay(f, day(0)); err != nil {
		t.Errorf("a missing 3 %% sleeve must not sink the overlay: %v", err)
	}
	delete(f, "VFITX") // the US donor goes missing
	if _, err := globalBondOverlay(f, day(0)); err == nil {
		t.Error("a missing US sleeve must fail the overlay, it is four fifths of it")
	}
}

// cashDaily interpolates a monthly accrual onto business days and lands exactly
// on each published level, so the financing it feeds an excess leg is a smooth
// daily rate rather than one jump a month.
func TestCashDailyInterpolatesToBusinessDays(t *testing.T) {
	m := &marketdata.Series{Points: []marketdata.Point{
		{Date: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), Close: 100},
		{Date: time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC), Close: 101},
		{Date: time.Date(2020, 3, 1, 0, 0, 0, 0, time.UTC), Close: 102},
	}}
	got := cashDaily("cash", m)
	if len(got.Points) < 40 {
		t.Fatalf("%d points, want two months of business days", len(got.Points))
	}
	if v := got.Last().Close; v != 102 {
		t.Errorf("last level %v, want the last published 102", v)
	}
	// Every interpolated point is a business day; the final one is the last
	// published anchor, appended as-is whatever day of the week it fell on.
	for _, p := range got.Points[:len(got.Points)-1] {
		if wd := p.Date.Weekday(); wd == time.Saturday || wd == time.Sunday {
			t.Fatalf("%s is a weekend", p.Date.Format("2006-01-02"))
		}
	}
	for i := 1; i < len(got.Points); i++ {
		if got.Points[i].Close < got.Points[i-1].Close {
			t.Fatalf("a money-market accrual must not fall: %v", got.Points[i-1:i+1])
		}
	}
}
