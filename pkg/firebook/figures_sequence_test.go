package firebook

import (
	"math"
	"strings"
	"testing"

	"github.com/bpineau/pofo/pkg/replay"
)

// millRun replays the plate's household (Bengen's fixed real withdrawal, 1 M EUR,
// 40 k EUR indexed, thirty years) from one start year.
func millRun(t *testing.T, start int) replay.Result {
	t.Helper()
	res, err := replay.Run(replay.Setup{
		Start: start, Capital: 1e6, Spend: 40000, Years: 30,
		Mu: 0.045, Sigma: 0.10, Df: 5, TargetRuin: 0.05, RaiseCap: 1.5,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Rules[0].NameFR != "Retrait fixe" {
		t.Fatalf("rule 0 is %q, the plate assumes the fixed rule", res.Rules[0].NameFR)
	}
	if res.Partial || res.Years != 30 {
		t.Fatalf("start %d: the record covers %d years (partial=%v), the plate draws thirty",
			start, res.Years, res.Partial)
	}
	return res
}

// millCheckWealth compares one frozen capital path (1000 then Rule.Wealth in
// k EUR) against the engine.
func millCheckWealth(t *testing.T, start int, want []float64, got []float64) {
	t.Helper()
	if want[0] != 1000 {
		t.Fatalf("start %d: the path must open on the 1 M EUR stake, not %.1f", start, want[0])
	}
	if len(want) > len(got)+1 {
		t.Fatalf("start %d: the plate freezes %d points, the engine has %d", start, len(want), len(got)+1)
	}
	for i, w := range want[1:] {
		if g := got[i] / 1000; math.Abs(g-w) > 0.1 {
			t.Errorf("start %d, year %d: engine %.1f k, plate %.1f k", start, i+1, g, w)
		}
	}
}

// The two vintages of the sequence plate carry frozen capital paths and four
// frozen returns; recompute all of them from pkg/replay and fail on any drift,
// the ruin year included.
func TestMillesimesFigureMatchesTheEngine(t *testing.T) {
	res66 := millRun(t, 1966)
	fixed66 := res66.Rules[0]
	if !fixed66.Ruined || fixed66.RuinYear != millRuinYear {
		t.Errorf("the 1966 vintage ruins at %d (ruined=%v), the plate says %d",
			fixed66.RuinYear, fixed66.Ruined, millRuinYear)
	}
	// The abscissa counts elapsed years, so the zero sits one further out than
	// the calendar difference: 1966 is the first year of retirement.
	if millRuinAt != millRuinYear-1966+1 {
		t.Errorf("the plate puts the zero %d years after the start, %d is the %dth year of retirement",
			millRuinAt, millRuinYear, millRuinYear-1966+1)
	}
	millCheckWealth(t, 1966, millWealth1966, fixed66.Wealth)
	// The frozen path stops on the zero, and the engine agrees the year before
	// still had something left, part paid.
	if n := len(millWealth1966); n != millRuinAt+1 || millWealth1966[n-1] != 0 {
		t.Errorf("the 1966 path has %d points ending on %.1f, expected %d ending on zero",
			n, millWealth1966[n-1], millRuinAt+1)
	}
	if got := fixed66.Wealth[millRuinAt-1]; got != 0 {
		t.Errorf("the engine still holds %.0f EUR at the end of %d", got, millRuinYear)
	}
	if got := fixed66.Wealth[millRuinAt-2]; got <= 0 {
		t.Errorf("the engine is already empty at the end of %d", millRuinYear-1)
	}
	if got := fixed66.Spend[millRuinAt-1]; got <= 0 || got >= 40000 {
		t.Errorf("%d pays %.0f EUR, the plate assumes a part-paid last year", millRuinYear, got)
	}

	res82 := millRun(t, 1982)
	fixed82 := res82.Rules[0]
	if fixed82.Ruined {
		t.Errorf("the 1982 vintage ruins at %d, the plate draws it whole", fixed82.RuinYear)
	}
	millCheckWealth(t, 1982, millWealth1982, fixed82.Wealth)
	if n := len(millWealth1982); n != 31 {
		t.Errorf("the 1982 path has %d points, the plate draws thirty years plus the stake", n)
	}
	if got := fixed82.Final / 1000; math.Abs(got-millWealth1982[30]) > 0.1 {
		t.Errorf("the 1982 vintage ends on %.1f k, the plate says %.1f k", got, millWealth1982[30])
	}

	// The four returns of the reading strip, in percent.
	for _, tc := range []struct {
		name string
		got  float64
		want float64
	}{
		{"1966, trente ans", res66.CAGR * 100, millCagr1966},
		{"1966, première décennie", res66.Decade * 100, millDecade1966},
		{"1982, trente ans", res82.CAGR * 100, millCagr1982},
		{"1982, première décennie", res82.Decade * 100, millDecade1982},
	} {
		if math.Abs(tc.got-tc.want) > 0.005 {
			t.Errorf("%s: engine %.3f %%, plate %.2f %%", tc.name, tc.got, tc.want)
		}
	}

	// The plate's punchline: the hostile vintage out-earned its own withdrawal
	// rate over the thirty years and still died. Without that, the figure would
	// only be saying that bad returns ruin a plan.
	if millCagr1966 <= 4.0 {
		t.Errorf("the 1966 vintage returned %.2f %%, the note claims more than the 4 %% withdrawn", millCagr1966)
	}

	// The tinted band's two claims: after its first year the 1966 household never
	// climbs back above its stake, and the 1982 one never falls back below it.
	for i, w := range fixed66.Wealth {
		if w > 1e6 {
			t.Errorf("the 1966 vintage is back above its stake in %d (%.0f EUR)", 1966+i, w)
			break
		}
	}
	for i, w := range fixed82.Wealth {
		if w < 1e6 {
			t.Errorf("the 1982 vintage falls back below its stake in %d (%.0f EUR)", 1982+i, w)
			break
		}
	}
}

// The plate writes its returns with a typographic minus and a French decimal
// comma; a hyphen or a dot would be a house-rule break.
func TestMillesimesPctFormatting(t *testing.T) {
	for _, tc := range []struct{ in float64 }{{millDecade1966}, {millDecade1982}, {millCagr1966}, {millCagr1982}} {
		s := millPct(tc.in)
		if strings.ContainsAny(s, ".-") {
			t.Errorf("millPct(%.2f) = %q, expected a comma and a U+2212 minus", tc.in, s)
		}
	}
	if got := millPct(millDecade1966); got != "−1,2 %/an" {
		t.Errorf("millPct(%.2f) = %q", millDecade1966, got)
	}
	if got := millPct(millDecade1982); got != "+11,4 %/an" {
		t.Errorf("millPct(%.2f) = %q", millDecade1982, got)
	}
}
