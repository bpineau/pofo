package firebook

import (
	"bufio"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/bpineau/pofo/pkg/datasets"
)

// TestBriquesEsperanceArithmetic redoes, for every row of the plate, the
// addition the article states in prose. A brick that no longer sums to the
// printed total means the plate and the text have drifted apart.
func TestBriquesEsperanceArithmetic(t *testing.T) {
	// lo, hi are the expectation range the article prints for each class.
	want := map[string][2]float64{
		"Actions US":       {2.5, 4.0},
		"Actions hors US":  {4.0, 6.0},
		"Obligations euro": {1.2, 1.2},
		"Monétaire euro":   {0, 1.0},
		"Or":               {0, 1.0},
	}
	seen := map[string]bool{}
	for _, r := range briqueRows() {
		w, ok := want[r.name]
		if !ok {
			t.Fatalf("unexpected row %q", r.name)
		}
		seen[r.name] = true

		// bricks must be laid end to end from zero, up to the socle.
		at := 0.0
		for i, s := range r.bricks {
			if math.Abs(s.from-at) > 1e-9 {
				t.Errorf("%s: brick %d starts at %.2f, previous one ended at %.2f", r.name, i, s.from, at)
			}
			at = s.to
		}
		if r.socle > 0 && math.Abs(at-r.socle) > 1e-9 {
			t.Errorf("%s: bricks end at %.2f, socle is %.2f", r.name, at, r.socle)
		}

		// the expectation range: socle + the valuation term when there is one,
		// the dashed assumption box otherwise.
		lo, hi := at, at
		switch {
		case r.hasTerm:
			lo, hi = r.socle+r.termLo, r.socle+r.termHi
			// the drawn pieces must tile the term and touch the socle.
			segs := r.termSegs()
			if len(segs) == 0 {
				t.Errorf("%s: a valuation term that draws nothing", r.name)
			}
			prev := math.Min(lo, r.socle)
			for i, s := range segs {
				if math.Abs(s.from-prev) > 1e-9 {
					t.Errorf("%s: term piece %d starts at %.2f, previous one ended at %.2f", r.name, i, s.from, prev)
				}
				prev = s.to
			}
			if want := math.Max(hi, r.socle); math.Abs(prev-want) > 1e-9 {
				t.Errorf("%s: term pieces end at %.2f, want %.2f", r.name, prev, want)
			}
		case r.hollow[1] > r.hollow[0]:
			lo, hi = r.hollow[0], r.hollow[1]
		}
		if math.Abs(lo-w[0]) > 1e-9 || math.Abs(hi-w[1]) > 1e-9 {
			t.Errorf("%s: plate reaches [%.2f, %.2f], the article states [%.2f, %.2f]", r.name, lo, hi, w[0], w[1])
		}
		if want := math.Max(hi, r.socle); math.Abs(r.extent()-want) > 1e-9 {
			t.Errorf("%s: row extends to %.2f, want %.2f", r.name, r.extent(), want)
		}

		// the printed total must read back as that same range.
		plo, phi := briqueParseTotal(t, r.name, r.total)
		if math.Abs(plo-lo) > 1e-9 || math.Abs(phi-hi) > 1e-9 {
			t.Errorf("%s: total reads %q, the arithmetic gives [%.2f, %.2f]", r.name, r.total, lo, hi)
		}
	}
	for name := range want {
		if !seen[name] {
			t.Errorf("row %q missing from the plate", name)
		}
	}

	svg := figBriquesEsperance()
	for _, s := range []string{"2,5 à 4,0 %", "4,0 à 6,0 %", "1,2 %", "distribution", "croissance des bénéfices"} {
		if !strings.Contains(svg, s) {
			t.Errorf("plate does not print %q", s)
		}
	}
	if strings.Contains(svg, "rgba(") {
		t.Error("plate uses rgba, which crengine paints solid black")
	}
}

// briqueParseTotal reads back what the totals column prints ("2,5 à 4,0 %",
// "1,2 %", "0 à 1 %") so the test compares numbers and not spellings.
func briqueParseTotal(t *testing.T, row, s string) (lo, hi float64) {
	t.Helper()
	num := func(p string) float64 {
		v, err := strconv.ParseFloat(strings.Replace(strings.TrimSpace(p), ",", ".", 1), 64)
		if err != nil {
			t.Fatalf("%s: total %q is not a number: %v", row, s, err)
		}
		return v
	}
	body := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(s), "%"))
	if a, bb, ok := strings.Cut(body, " à "); ok {
		return num(a), num(bb)
	}
	v := num(body)
	return v, v
}

// TestEscalierMorningstarCAPE rebuilds the plate's conditioning band from the
// bundled Shiller series: each edition of "The State of Retirement Income" is
// arrested on 30 September, so the band carries the CAPE of that September.
func TestEscalierMorningstarCAPE(t *testing.T) {
	frozenAgainstData(t)
	cape := map[string]float64{}
	sc := bufio.NewScanner(strings.NewReader(string(datasets.CAPE())))
	for sc.Scan() {
		l := strings.TrimSpace(sc.Text())
		if l == "" || strings.HasPrefix(l, "#") || strings.HasPrefix(l, "date") {
			continue
		}
		f := strings.Split(l, ",")
		if len(f) < 2 {
			continue
		}
		if v, err := strconv.ParseFloat(f[1], 64); err == nil && v > 0 {
			cape[f[0]] = v
		}
	}
	if len(cape) < 1500 {
		t.Fatalf("bundled CAPE holds only %d months", len(cape))
	}
	if len(morningstarCAPE) != len(morningstarSWR) || len(morningstarWhy) != len(morningstarSWR) ||
		len(morningstarYears) != len(morningstarSWR) {
		t.Fatal("the plate's vintage arrays have different lengths")
	}
	for i, y := range morningstarYears {
		key := y + "-09-01"
		got, ok := cape[key]
		if !ok {
			t.Fatalf("bundled CAPE has no %s", key)
		}
		if math.Abs(got-morningstarCAPE[i]) > 0.005 {
			t.Errorf("%s: plate says CAPE %.2f, the bundled Shiller series says %.2f", key, morningstarCAPE[i], got)
		}
	}
}

// TestEscalierMorningstarSteps guards the published vintages and the plate that
// draws them. The rates are Morningstar's own base cases (3,3 / 3,8 / 4,0 in
// the 2025 edition's recap, 3,7 for 2024 and 3,9 for 2025 in its key takeaways),
// and the article must state the same series.
func TestEscalierMorningstarSteps(t *testing.T) {
	want := []float64{3.3, 3.8, 4.0, 3.7, 3.9}
	if len(morningstarSWR) != len(want) {
		t.Fatalf("plate draws %d vintages, %d published", len(morningstarSWR), len(want))
	}
	for i, v := range want {
		if math.Abs(morningstarSWR[i]-v) > 1e-9 {
			t.Errorf("vintage %s: plate says %.1f, Morningstar published %.1f", morningstarYears[i], morningstarSWR[i], v)
		}
	}
	svg := figEscalierMorningstar()
	for _, s := range []string{"3,3 %", "3,8 %", "4,0 %", "3,7 %", "3,9 %", "37,6", "28,2", "38,6", "2021", "2025"} {
		if !strings.Contains(svg, s) {
			t.Errorf("plate does not print %q", s)
		}
	}
	if strings.Contains(svg, "rgba(") {
		t.Error("plate uses rgba, which crengine paints solid black")
	}
	// the tint must stay a solid hex, monotone in the CAPE
	if capeTint(28) == capeTint(39) {
		t.Error("the CAPE band does not shade with the CAPE")
	}
	for _, c := range morningstarCAPE {
		if got := capeTint(c); len(got) != 7 || got[0] != '#' {
			t.Errorf("capeTint(%.2f) = %q, want a solid hex", c, got)
		}
	}
}
