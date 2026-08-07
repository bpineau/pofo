package firebook

import (
	"math"
	"testing"

	"github.com/bpineau/pofo/pkg/replay"
)

// The Bengen plate carries frozen numbers so the book's figures stay pure
// functions. This recomputes them from pkg/replay and the bundled record, and
// fails the moment the figure and the engine disagree.
func TestBengenFigureMatchesTheEngine(t *testing.T) {
	res, err := replay.Run(replay.Setup{
		Start: bengenStart, Capital: 600000, Spend: 24000, Years: 40,
		Mu: 0.045, Sigma: 0.10, Df: 5, TargetRuin: 0.05, RaiseCap: 1.5,
	})
	if err != nil {
		t.Fatal(err)
	}
	fixed := res.Rules[0]
	if fixed.NameFR != "Retrait fixe" {
		t.Fatalf("rule 0 is %q, the plate assumes the fixed rule", fixed.NameFR)
	}
	if !fixed.Ruined || fixed.RuinYear != bengenRuin {
		t.Errorf("engine ruins at %d (ruined=%v), plate says %d", fixed.RuinYear, fixed.Ruined, bengenRuin)
	}
	if len(bengenRates) != len(bengenWealth) {
		t.Fatalf("%d rates against %d wealth points", len(bengenRates), len(bengenWealth))
	}

	// The reading of a January is the year's spend over what was standing at the
	// end of the year before, which the plate freezes as a percent.
	open := 600000.0
	for i, wantRate := range bengenRates {
		if got := fixed.Spend[i] / open * 100; math.Abs(got-wantRate) > 0.02 {
			t.Errorf("january %d: engine reads %.2f %%, plate says %.2f %%", bengenStart+i, got, wantRate)
		}
		if got := open / 1000; math.Abs(got-bengenWealth[i]) > 0.05 {
			t.Errorf("january %d: engine holds %.1f k, plate says %.1f k", bengenStart+i, got, bengenWealth[i])
		}
		open = fixed.Wealth[i]
	}

	// The article and the plate both hang on 1975 being the first reading in the
	// alert zone, and on the nineteen years that follow it.
	for i, r := range bengenRates {
		if r >= 8 {
			if bengenStart+i != 1975 {
				t.Errorf("first reading above 8 %% is %d, the plate names 1975", bengenStart+i)
			}
			break
		}
	}
}

// The CAPE plate states a rule and three of its readings; check the arithmetic
// the article quotes, since figure and prose must agree.
func TestCapeRateMatchesTheArticle(t *testing.T) {
	for _, tc := range []struct {
		cape, want float64
	}{{30, 3.417}, {21, 4.131}, {38, 3.066}} {
		if got := capeRate(tc.cape); math.Abs(got-tc.want) > 0.002 {
			t.Errorf("CAPE %.0f: rate %.3f %%, expected %.3f %%", tc.cape, got, tc.want)
		}
	}
	// The krach costs the rule 15 % of income where a fixed percentage costs 30 %.
	base := 1.50 * capeRate(30)
	cape := 1.05 * capeRate(21)
	if got := cape/base - 1; math.Abs(got+0.154) > 0.005 {
		t.Errorf("krach income move %.3f, expected -0.154", got)
	}
}

// The mortality plate's Gompertz law is calibrated on the orders of magnitude
// the article quotes in words; keep the two in step.
func TestMortalityCreditsMatchTheArticle(t *testing.T) {
	for _, tc := range []struct {
		age, lo, hi float64
	}{
		{65, 0.5, 1.0}, // "faibles à 65 ans"
		{80, 3.0, 5.0}, // "3 à 5 %/an après 80 ans"
		{85, 3.0, 6.0},
	} {
		got := gompertzCredit(tc.age)
		if got < tc.lo || got > tc.hi {
			t.Errorf("credit at %.0f is %.2f %%, outside the article's %.1f-%.1f %%", tc.age, got, tc.lo, tc.hi)
		}
	}
	if gompertzCredit(70) >= gompertzCredit(75) {
		t.Error("the credit must grow with age")
	}
}
