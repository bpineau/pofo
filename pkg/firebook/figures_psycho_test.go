package firebook

import (
	"math"
	"strings"
	"testing"

	"github.com/bpineau/pofo/pkg/replay"
)

// mourirRicheReplay recomputes the plate's estates: the book's reference
// household (1 M EUR, Bengen's fixed real withdrawal, thirty years) started
// every year the bundled record covers whole, for one yearly spending.
func mourirRicheReplay(t *testing.T, spend float64) (first int, out []float64) {
	t.Helper()
	first = 0
	for start := 1950; start <= 2025; start++ {
		res, err := replay.Run(replay.Setup{
			Start: start, Capital: 1e6, Spend: spend, Years: 30,
			Mu: 0.045, Sigma: 0.10, Df: 5, TargetRuin: 0.05, RaiseCap: 1.5,
		})
		if err != nil || res.Partial || res.Years != 30 {
			continue // the record does not cover that vintage whole
		}
		if res.Rules[0].NameFR != "Retrait fixe" {
			t.Fatalf("rule 0 is %q, the plate assumes the fixed rule", res.Rules[0].NameFR)
		}
		if first == 0 {
			first = start
		}
		out = append(out, res.Rules[0].Final/1e6)
	}
	return first, out
}

// Every frozen multiple, both medians and all four counts come from pkg/replay;
// recompute them and fail on any drift. The sample must stay complete-window
// only: padding it with truncated vintages would flatter the distribution.
func TestMourirRicheFigureMatchesTheEngine(t *testing.T) {
	for _, tc := range []struct {
		name  string
		spend float64
		want  []float64
	}{
		{"40 000 EUR", 40000, mourirRicheSpent},
		{"30 000 EUR", 30000, mourirRicheThrift},
	} {
		first, got := mourirRicheReplay(t, tc.spend)
		if first != mourirRicheFirst {
			t.Errorf("%s: the record's first whole vintage is %d, the plate says %d",
				tc.name, first, mourirRicheFirst)
		}
		if len(got) != len(tc.want) {
			t.Fatalf("%s: the engine has %d whole vintages, the plate freezes %d",
				tc.name, len(got), len(tc.want))
		}
		for i, w := range tc.want {
			if math.Abs(got[i]-w) > 0.001 {
				t.Errorf("%s, vintage %d: engine %.4fx, plate %.3fx",
					tc.name, mourirRicheFirst+i, got[i], w)
			}
		}
	}

	// The four counts and the two medians the plate and the article state.
	for _, tc := range []struct {
		name            string
		vs              []float64
		median          float64
		above, atLeast3 int
	}{
		{"40 000 EUR", mourirRicheSpent, 1.9, 25, 12},
		{"30 000 EUR", mourirRicheThrift, 2.6, 41, 19},
	} {
		if got := mourirRicheMedian(tc.vs); math.Abs(got-tc.median) > 0.05 {
			t.Errorf("%s: median %.3fx, the plate rounds it to %.1fx", tc.name, got, tc.median)
		}
		if got := mourirRicheAboveStake(tc.vs); got != tc.above {
			t.Errorf("%s: %d vintages above the stake, the plate says %d", tc.name, got, tc.above)
		}
		if got := mourirRicheAtLeast(tc.vs, 3); got != tc.atLeast3 {
			t.Errorf("%s: %d vintages at 3x or more, the plate says %d", tc.name, got, tc.atLeast3)
		}
	}

	// The sample is odd-sized, so the median is a vintage the record produced.
	if len(mourirRicheSpent)%2 != 1 {
		t.Error("the plate reads its median off the middle rank, which needs an odd sample")
	}
	// Exactly one vintage ran out of money, and it is the one the plate names.
	zeros := 0
	for i, v := range mourirRicheSpent {
		if v == 0 {
			zeros++
			if mourirRicheFirst+i != 1966 {
				t.Errorf("the vintage that ends at zero is %d, the plate names 1966", mourirRicheFirst+i)
			}
		}
	}
	if zeros != 1 {
		t.Errorf("%d vintages end at zero, the plate draws one cross", zeros)
	}
	// The thrifty household never ran out, which is why its cloud has no cross.
	for i, v := range mourirRicheThrift {
		if v <= 0 {
			t.Errorf("the 30 000 EUR vintage %d ends at %.3fx, the plate draws no ruin there",
				mourirRicheFirst+i, v)
		}
	}
	// The closing line: the quarter given up escaped ruin in exactly one
	// vintage, and merely grew the estate in every other one.
	if zeros != 1 {
		t.Errorf("the closing line claims one ruin avoided, the record has %d ruins", zeros)
	}
	for i := range mourirRicheSpent {
		if mourirRicheThrift[i] <= mourirRicheSpent[i] {
			t.Errorf("vintage %d: spending less left %.3fx against %.3fx, the plate claims a bigger estate every time",
				mourirRicheFirst+i, mourirRicheThrift[i], mourirRicheSpent[i])
		}
	}
	if mourirRicheMedian(mourirRicheThrift) <= mourirRicheMedian(mourirRicheSpent) {
		t.Error("the thrifty median is not above the spending one")
	}
}

// The plate must render with French decimal commas and no em-dash, and must
// carry the counts it claims.
func TestMourirRicheFigureRenders(t *testing.T) {
	s := figMourirRiche()
	if strings.Contains(s, "—") { // the em-dash, banned book-wide
		t.Error("the plate contains an em-dash")
	}
	if strings.Contains(s, "rgba(") || strings.Contains(s, "opacity") {
		t.Error("the plate uses rgba or opacity, which crengine paints solid black")
	}
	for _, want := range []string{"25 sur 43", "12 sur 43", "41 sur 43", "19 sur 43",
		"2,6 fois la mise", "1,9 fois la mise", "1966"} {
		if !strings.Contains(s, want) {
			t.Errorf("the plate does not state %q", want)
		}
	}
	if fn, ok := figures["mourir-riche"]; !ok || fn == nil {
		t.Error("the plate is not registered as mourir-riche")
	}
}
