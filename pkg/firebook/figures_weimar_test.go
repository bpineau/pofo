package firebook

import (
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/bpineau/pofo/pkg/datasets"
)

// weimarPanel reads the German real annual returns out of the bundled JST
// panel: year -> {equity, bond}.
func weimarPanel(t *testing.T) map[int][2]float64 {
	t.Helper()
	frozenAgainstData(t)
	out := map[int][2]float64{}
	for _, line := range strings.Split(string(datasets.BroadSample()), "\n") {
		if strings.HasPrefix(line, "#") || !strings.HasPrefix(line, "DEU,") {
			continue
		}
		f := strings.Split(strings.TrimSpace(line), ",")
		if len(f) < 4 {
			continue
		}
		year, err := strconv.Atoi(f[1])
		if err != nil {
			t.Fatalf("bad year in %q: %v", line, err)
		}
		if year < weimarFrom || year > weimarTo {
			continue // the panel leaves later German years blank, war by war
		}
		eq, err := strconv.ParseFloat(f[2], 64)
		if err != nil {
			t.Fatalf("bad equity return in %q: %v", line, err)
		}
		bd, err := strconv.ParseFloat(f[3], 64)
		if err != nil {
			t.Fatalf("bad bond return in %q: %v", line, err)
		}
		out[year] = [2]float64{eq, bd}
	}
	if len(out) == 0 {
		t.Fatal("no DEU rows in the bundled panel")
	}
	return out
}

// The plate freezes two cumulative paths so the book's figures stay pure
// functions. This recompounds them from the panel and fails on any drift.
func TestWeimarPathsMatchTheJSTPanel(t *testing.T) {
	panel := weimarPanel(t)
	for _, p := range weimarPaths {
		if got, want := len(p.Vals), weimarTo-weimarFrom+1; got != want {
			t.Fatalf("%s: %d values for %d years", p.Label, got, want)
		}
		if p.Vals[0] != 100 {
			t.Errorf("%s: base is %.2f, the plate is indexed to 100", p.Label, p.Vals[0])
		}
	}
	eq, bd := 100.0, 100.0
	for year := weimarFrom + 1; year <= weimarTo; year++ {
		r, ok := panel[year]
		if !ok {
			t.Fatalf("the panel has no German row for %d", year)
		}
		eq *= 1 + r[0]
		if bd > 0 {
			bd *= 1 + r[1]
		}
		if bd < 0 {
			bd = 0 // a total wipe-out is coded −1 exactly; nothing compounds out of nothing
		}
		i := weimarYear(year)
		if got := weimarPaths[0].Vals[i]; math.Abs(got-eq) > 0.02 {
			t.Errorf("%d equities: panel gives %.2f, plate says %.2f", year, eq, got)
		}
		if got := weimarPaths[1].Vals[i]; math.Abs(got-bd) > 0.02 {
			t.Errorf("%d bonds: panel gives %.2f, plate says %.2f", year, bd, got)
		}
	}
}

// The plate and the article both hang on three facts: the bond is wiped out in
// 1923, the equity leg bottoms in 1925 having lost more than 80 %, and it is
// visibly alive the year after. Pin them.
func TestWeimarClaimsHold(t *testing.T) {
	panel := weimarPanel(t)
	if r := panel[1923][1]; r != -1 {
		t.Errorf("the panel's 1923 German bond return is %.4f, the plate reads it as a total wipe-out", r)
	}
	bonds := weimarPaths[1].Vals
	for year := 1923; year <= weimarTo; year++ {
		if v := bonds[weimarYear(year)]; v != 0 {
			t.Errorf("%d: the bond leg is %.4f, it must stay at zero once wiped out", year, v)
		}
	}
	eq := weimarPaths[0].Vals
	low := eq[weimarYear(1925)]
	for year := weimarMurkTo + 1; year <= weimarTo; year++ {
		if v := eq[weimarYear(year)]; v < low {
			t.Errorf("%d: equities at %.2f, below the 1925 low of %.2f the plate marks", year, v, low)
		}
	}
	if low >= 20 || low <= 14 {
		t.Errorf("the 1925 equity low is %.2f, the plate annotates it as 17", low)
	}
	if end := eq[weimarYear(weimarTo)]; end <= low*2 {
		t.Errorf("equities end at %.2f against a %.2f low, the plate shows a visible rebound", end, low)
	}
}

// The plate must survive crengine (no rgba, no opacity) and the house rules
// (no em-dash, no rotated text), and must not silently draw the bond's zero as
// a plottable value.
func TestWeimarPlateIsClean(t *testing.T) {
	svg := FigureSVG("weimar-reel")
	if svg == "" {
		t.Fatal("weimar-reel is not registered")
	}
	for _, bad := range []string{"rgba", "opacity", "—", "rotate("} {
		if strings.Contains(svg, bad) {
			t.Errorf("the plate contains %q", bad)
		}
	}
	for _, want := range []string{"1922-1924", "mesure non fiable", "atteint zéro"} {
		if !strings.Contains(svg, want) {
			t.Errorf("the plate no longer says %q", want)
		}
	}
}
