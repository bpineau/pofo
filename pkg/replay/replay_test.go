package replay

import (
	"math"
	"testing"
)

// bookSetup is the household the book's replays use: the one every test below
// reasons about.
func bookSetup(start int) Setup {
	return Setup{Start: start, Capital: 600000, Spend: 24000, Years: 40,
		Mu: 0.045, Sigma: 0.10, Df: 5, TargetRuin: 0.05, RaiseCap: 1.5}
}

func TestRunShape(t *testing.T) {
	r, err := Run(bookSetup(1973))
	if err != nil {
		t.Fatal(err)
	}
	if r.Years != 40 || r.Partial {
		t.Errorf("1973 covers %d years (partial %v), want the full 40", r.Years, r.Partial)
	}
	if r.End != 2012 {
		t.Errorf("ends %d, want 2012", r.End)
	}
	if len(r.Rules) != 7 {
		t.Fatalf("%d rules, want 7", len(r.Rules))
	}
	if want := 12*r.Years + 1; len(r.Index) != want || len(r.Dates) != want {
		t.Errorf("index %d / dates %d points, want %d", len(r.Index), len(r.Dates), want)
	}
	for _, rule := range r.Rules {
		if len(rule.Spend) != r.Years || len(rule.Wealth) != r.Years {
			t.Errorf("%s: %d spends, %d wealths, want %d each", rule.Name, len(rule.Spend), len(rule.Wealth), r.Years)
		}
		if rule.Color == "" || rule.NameFR == "" || rule.Tag == "" {
			t.Errorf("%s: incomplete labelling (%q %q %q)", rule.Name, rule.Color, rule.NameFR, rule.Tag)
		}
	}
}

// A record shorter than the plan must report the covered prefix and say so,
// never pad the missing years with a zero return.
func TestRunPartialWindow(t *testing.T) {
	r, err := Run(bookSetup(2000))
	if err != nil {
		t.Fatal(err)
	}
	if !r.Partial {
		t.Error("a 40-year plan started in 2000 cannot be complete yet")
	}
	if r.Years >= 40 || r.End != r.Start+r.Years-1 {
		t.Errorf("%d years, ending %d", r.Years, r.End)
	}
	if len(r.Rules[0].Spend) != r.Years {
		t.Errorf("the fixed rule reports %d years, want %d", len(r.Rules[0].Spend), r.Years)
	}
}

// The horizon is the plan, not the data: shortening it must change what the
// amortisation rule pays, which is exactly why Setup.Years stays whole.
func TestHorizonDrivesTheAmortisationRule(t *testing.T) {
	long, err := Run(bookSetup(2000))
	if err != nil {
		t.Fatal(err)
	}
	short := bookSetup(2000)
	short.Years = 26
	brief, err := Run(short)
	if err != nil {
		t.Fatal(err)
	}
	abw := func(r Result) float64 {
		for _, rule := range r.Rules {
			if rule.Name == "ABW" {
				return rule.Spend[0]
			}
		}
		return 0
	}
	if abw(brief) <= abw(long) {
		t.Errorf("ABW first payment: 26-year plan %.0f, 40-year plan %.0f; a shorter horizon must pay MORE",
			abw(brief), abw(long))
	}
}

// The fixed rule is the control every other column is read against: it must
// deliver exactly the target until it cannot.
func TestFixedRuleIsTheControl(t *testing.T) {
	r, err := Run(bookSetup(1985))
	if err != nil {
		t.Fatal(err)
	}
	fixed := r.Rules[0]
	if fixed.Name != "Fixed" {
		t.Fatalf("first rule is %q", fixed.Name)
	}
	if fixed.CV > 1e-9 || fixed.LeanYears != 0 {
		t.Errorf("CV %.4f, %d lean years; the fixed rule never moves", fixed.CV, fixed.LeanYears)
	}
	for i, v := range fixed.Spend {
		if math.Abs(v-r.Setup.Spend) > 1 {
			t.Fatalf("year %d delivered %.0f, want %.0f", r.Start+i, v, r.Setup.Spend)
		}
	}
	// 1985 is the era where a rigid rule dies rich: the point of the section.
	if fixed.Final < 3e6 {
		t.Errorf("1985 fixed rule left %.0f, expected a multi-million unspent estate", fixed.Final)
	}
}

// A wealth-proportional rule spends a share of what is left, so it cannot ruin,
// whatever the sequence. The book states this in prose; the test holds it.
func TestWealthProportionalRulesNeverRuin(t *testing.T) {
	for _, start := range []int{1973, 1985, 2000} {
		r, err := Run(bookSetup(start))
		if err != nil {
			t.Fatal(err)
		}
		for _, rule := range r.Rules {
			if (rule.Name == "VPW" || rule.Name == "ABW") && rule.Ruined {
				t.Errorf("%d: %s ruined in %d", start, rule.Name, rule.RuinYear)
			}
		}
	}
}

// The market statistics must match the era they claim to describe.
func TestMarketStats(t *testing.T) {
	for _, tc := range []struct {
		start            int
		cagrLo, cagrHi   float64
		decadeLo, decHi  float64
		worstAt          int
		maxDDLo, maxDDHi float64
	}{
		{1973, 0.040, 0.055, -0.03, 0.01, 1974, -0.45, -0.30},
		{1985, 0.058, 0.072, 0.06, 0.11, 2022, -0.35, -0.20},
	} {
		r, err := Run(bookSetup(tc.start))
		if err != nil {
			t.Fatal(err)
		}
		if r.CAGR < tc.cagrLo || r.CAGR > tc.cagrHi {
			t.Errorf("%d: CAGR %.2f%%, want %.1f-%.1f%%", tc.start, r.CAGR*100, tc.cagrLo*100, tc.cagrHi*100)
		}
		if r.Decade < tc.decadeLo || r.Decade > tc.decHi {
			t.Errorf("%d: first decade %.2f%%, want %.1f-%.1f%%", tc.start, r.Decade*100, tc.decadeLo*100, tc.decHi*100)
		}
		if r.WorstAt != tc.worstAt {
			t.Errorf("%d: worst year %d, want %d", tc.start, r.WorstAt, tc.worstAt)
		}
		if r.MaxDrawdown < tc.maxDDLo || r.MaxDrawdown > tc.maxDDHi {
			t.Errorf("%d: max drawdown %.1f%%, want %.0f to %.0f%%", tc.start, r.MaxDrawdown*100, tc.maxDDLo*100, tc.maxDDHi*100)
		}
	}
}

func TestRunRejectsNonsense(t *testing.T) {
	for _, s := range []Setup{
		{Start: 1973, Capital: 0, Spend: 24000, Years: 40},
		{Start: 1973, Capital: 600000, Spend: 0, Years: 40},
		{Start: 1973, Capital: 600000, Spend: 24000, Years: 0},
	} {
		if _, err := Run(s); err == nil {
			t.Errorf("Run(%+v) accepted a degenerate setup", s)
		}
	}
	if _, err := Run(bookSetup(1900)); err == nil {
		t.Error("1900 is before the record; want an error")
	}
}

// The safe-rate table is the risk guardrail's band centre: it must rise as the
// horizon shortens (a five-year horizon can afford far more than a forty-year
// one), which is the whole difference from the 2006 rule.
func TestSafeRateTableRisesAsTheHorizonShortens(t *testing.T) {
	tab := SafeRateTable(40, 0.045, 0.10, 5, 0.05)
	if len(tab) != 40 {
		t.Fatalf("%d entries, want 40", len(tab))
	}
	for i := 1; i < len(tab); i++ {
		if tab[i] < tab[i-1]-1e-9 {
			t.Fatalf("rate falls at year %d: %.4f then %.4f", i, tab[i-1], tab[i])
		}
	}
	if tab[0] < 0.02 || tab[0] > 0.06 {
		t.Errorf("40-year safe rate %.2f%%, want a plausible 2-6%%", tab[0]*100)
	}
	if tab[len(tab)-1] <= tab[0] {
		t.Error("the short end must be looser than the long end")
	}
	// Cached: the same call must return the same slice content.
	again := SafeRateTable(40, 0.045, 0.10, 5, 0.05)
	for i := range tab {
		if tab[i] != again[i] {
			t.Fatalf("table not stable at %d: %.6f vs %.6f", i, tab[i], again[i])
		}
	}
}

func TestGeometricReturn(t *testing.T) {
	if got := GeometricReturn(0.045, 0.10); math.Abs(got-0.04) > 1e-9 {
		t.Errorf("GeometricReturn(4.5%%, 10%%) = %.4f, want 0.04", got)
	}
	if got := GeometricReturn(0.01, 0.40); got != 0 {
		t.Errorf("a vol that swamps the mean must floor at 0, got %.4f", got)
	}
}
