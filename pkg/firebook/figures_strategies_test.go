package firebook

import (
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/bpineau/pofo/pkg/datasets"
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

// The smoothing plate replays the real 1973 weather; keep the frozen returns in
// step with the engine, and the three paths in step with the article's example.
func TestSmoothingFigureMatchesTheEngine(t *testing.T) {
	res, err := replay.Run(replay.Setup{
		Start: 1973, Capital: 600000, Spend: 24000, Years: 40,
		Mu: 0.045, Sigma: 0.10, Df: 5, TargetRuin: 0.05, RaiseCap: 1.5,
	})
	if err != nil {
		t.Fatal(err)
	}
	for i, want := range pourcentageReturns {
		if got := res.Market[i]; math.Abs(got-want) > 0.0001 {
			t.Errorf("%d: engine %.4f, plate %.4f", 1973+i, got, want)
		}
	}
	raw, yale, corridor, end := smoothingPaths()
	for _, tc := range []struct {
		name string
		got  float64
		want float64
	}{
		{"brut an 1", raw[0], 56.0}, {"brut an 2", raw[1], 45.6}, {"brut an 3", raw[2], 33.8},
		{"yale an 2", yale[1], 52.9}, {"yale an 3", yale[2], 47.1},
		{"corridor an 3", corridor[2], 53.2},
	} {
		if math.Abs(tc.got-tc.want) > 0.05 {
			t.Errorf("%s: %.1f, the article says %.1f", tc.name, tc.got, tc.want)
		}
	}
	// The corridor's gentleness is paid on the capital: that is the plate's note.
	if end[2] >= end[0] {
		t.Errorf("corridor ends with %.0f k against the raw rule's %.0f k, the note claims less", end[2], end[0])
	}
}

// The CAPE plate freezes 146 Januaries; read the dataset back and fail on drift.
func TestCapeJanuariesMatchTheDataset(t *testing.T) {
	byYear := map[int]float64{}
	for _, l := range strings.Split(string(datasets.CAPE()), "\n") {
		f := strings.Split(l, ",")
		if len(f) < 2 || !strings.HasSuffix(f[0], "-01-01") {
			continue
		}
		y, err := strconv.Atoi(f[0][:4])
		if err != nil {
			continue
		}
		if v, err := strconv.ParseFloat(f[1], 64); err == nil {
			byYear[y] = v
		}
	}
	for i, want := range capeJanuaries {
		year := capeFirstYear + i
		got, ok := byYear[year]
		if !ok {
			t.Errorf("january %d is in the plate but not in the dataset", year)
			continue
		}
		if math.Abs(got-want) > 0.005 {
			t.Errorf("january %d: dataset %.2f, plate %.2f", year, got, want)
		}
	}
	if n := len(byYear); n != len(capeJanuaries) {
		t.Errorf("the dataset has %d januaries, the plate freezes %d", n, len(capeJanuaries))
	}
	// The two readings the caption names.
	if got := capeRate(capeJanuaries[1921-capeFirstYear]); math.Abs(got-11.5) > 0.05 {
		t.Errorf("1921 would have quoted %.1f %%, the caption says 11,5 %%", got)
	}
	if got := capeRate(capeJanuaries[2000-capeFirstYear]); math.Abs(got-2.9) > 0.05 {
		t.Errorf("2000 would have quoted %.1f %%, the caption says 2,9 %%", got)
	}
}
