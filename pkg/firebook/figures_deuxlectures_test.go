package firebook

import (
	"math"
	"strings"
	"testing"

	"github.com/bpineau/pofo/pkg/replay"
)

// The plate's two series are a snapshot of the bundled record, so this is what
// keeps them honest: the replay is run again and every year of both panels is
// recomputed. Gated with the other data-frozen checks, since a refreshed
// dataset is allowed to move a plate without breaking the build.
func TestDeuxLecturesMatchTheReplay(t *testing.T) {
	frozenAgainstData(t)
	res, err := replay.Run(replay.Setup{
		Start: deuxStart, Capital: deuxCapital * 1000, Spend: deuxCheque * 1000, Years: 30,
		Mu: 0.045, Sigma: 0.10, Df: 5, TargetRuin: 0.05, RaiseCap: 1.5,
	})
	if err != nil {
		t.Fatal(err)
	}
	rule := res.Rules[0] // the fixed inflation-indexed withdrawal
	if rule.NameFR != "Retrait fixe" {
		t.Fatalf("the plate reads rule %q, not the fixed withdrawal", rule.NameFR)
	}
	if len(deuxSpend) != res.Years {
		t.Fatalf("the plate draws %d years, the replay covers %d", len(deuxSpend), res.Years)
	}
	if !rule.Ruined || rule.RuinYear != deuxRuin {
		t.Errorf("the replay runs out in %d (ruined=%v), the plate says %d", rule.RuinYear, rule.Ruined, deuxRuin)
	}
	capital := deuxCapital * 1000
	for i, want := range deuxSpend {
		if got := rule.Spend[i] / 1000; math.Abs(got-want) > 0.05 {
			t.Errorf("%d: the replay pays %.1f k, the plate draws %.1f k", deuxStart+i, got, want)
		}
		if i < len(deuxRate) {
			rate := rule.Spend[i] / capital * 100
			if math.Abs(rate-deuxRate[i]) > 0.01 {
				t.Errorf("%d: the replay withdraws %.2f %% of what is left, the plate draws %.2f %%",
					deuxStart+i, rate, deuxRate[i])
			}
		}
		capital = rule.Wealth[i]
	}
}

// The shape of the story, which the plate exists to show and which no data
// refresh may quietly reverse: a cheque that never moves while the rate it
// represents climbs away from where it started.
func TestDeuxLecturesShape(t *testing.T) {
	if len(deuxRate) != len(deuxSpend)-1 {
		t.Fatalf("%d rates for %d payments: the last year has no capital to divide by",
			len(deuxRate), len(deuxSpend))
	}
	for i, v := range deuxSpend[:len(deuxSpend)-2] {
		if v != deuxCheque {
			t.Errorf("%d: the cheque is %.1f k, the rule pays a flat %.0f k until the end", deuxStart+i, v, deuxCheque)
		}
	}
	if deuxSpend[len(deuxSpend)-2] >= deuxCheque || deuxSpend[len(deuxSpend)-1] != 0 {
		t.Error("the last two years no longer show the cheque failing")
	}
	if deuxRate[0] != 4.0 {
		t.Errorf("the plan starts at %.2f %%, not at the 4 %% of the rule", deuxRate[0])
	}
	// The rate crosses the tinted level and never comes back under it.
	first := -1
	for i, v := range deuxRate {
		if v >= deuxTint && first < 0 {
			first = i
		}
		if first >= 0 && i > first && v < deuxTint && v > 0 {
			if i > first+2 {
				t.Errorf("%d: the rate fell back under %.0f %% long after crossing it", deuxStart+i, deuxTint)
			}
		}
	}
	if first < 0 || deuxStart+first != 1975 {
		t.Errorf("the rate first reaches %.0f %% in %d, the caption says 1975", deuxTint, deuxStart+first)
	}
	if got := deuxExit(); got != 1989 {
		t.Errorf("the curve leaves the panel in %d, the annotation says 1989", got)
	}
}

// The rendered plate obeys the book's drawing rules and carries both readings.
func TestDeuxLecturesPlateRenders(t *testing.T) {
	svg := FigureSVG("retrait-deux-lectures")
	if !strings.HasPrefix(svg, "<svg viewBox=") {
		t.Fatal("the plate must render an SVG")
	}
	// U+2014 and U+2013 are the em- and en-dashes, banned from the book down to
	// its figure labels; they stay escaped here so a repository-wide grep never
	// trips on them.
	for _, banned := range []string{"\u2014", "\u2013", "rgba(", "opacity", "rotate("} {
		if strings.Contains(svg, banned) {
			t.Errorf("the plate uses %q, which the book's figures never do", banned)
		}
	}
	for _, want := range []string{
		">Ce que voit le retraité<", ">Ce que voit le plan<",
		">4,0 %<", ">8,2 %<", // where the rate starts and where it turns
		">1966<", ">1975<", ">1994<",
		">40 k€<",
		">le dernier virement tombe en 1994<",
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("the plate does not draw %q", want)
		}
	}
	// One bar per paid year, the empty last one excepted.
	paid := 0
	for _, v := range deuxSpend {
		if v > 0 {
			paid++
		}
	}
	if n := strings.Count(svg, "<path"); n != paid {
		t.Errorf("%d cheques drawn, expected one per paid year (%d)", n, paid)
	}
	// One rate curve, and it stops at the ceiling rather than running off the
	// plate: the years above the scale are written out instead.
	if n := strings.Count(svg, "<polyline"); n != 1 {
		t.Errorf("%d curves drawn, the lower panel has exactly one", n)
	}
	if !strings.Contains(svg, ">le taux quitte l'échelle en 1989, passe 38 % en 1992,<") {
		t.Error("the plate does not say where the curve leaves the panel")
	}
}
