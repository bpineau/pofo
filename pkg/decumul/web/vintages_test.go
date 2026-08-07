package web

import (
	"strconv"
	"strings"
	"testing"
)

// The JST parse must keep the year index so a vintage can be located, and the
// canonical vintages must all exist in the bundled record.
func TestBroadSampleCountriesYears(t *testing.T) {
	byISO := map[string]countrySeries{}
	for _, c := range broadSampleCountries() {
		byISO[c.ISO] = c
	}
	for _, v := range vintageList {
		c, ok := byISO[v.iso]
		if !ok {
			t.Fatalf("vintage country %s missing from the bundled JST table", v.iso)
		}
		if v.year < c.FirstYear || v.year >= c.FirstYear+len(c.Returns) {
			t.Errorf("vintage %s %d outside the bundled record [%d..%d]",
				v.iso, v.year, c.FirstYear, c.FirstYear+len(c.Returns)-1)
		}
	}
	usa := byISO["USA"]
	if usa.FirstYear > 1900 || len(usa.Returns) < 100 {
		t.Errorf("USA record too short: first %d, %d years", usa.FirstYear, len(usa.Returns))
	}
}

// A plain 4%-rule plan replayed through the vintages must at least chart every
// vintage, and the 1929/1966 replays are known hard cases: at a 5% withdrawal
// with heavy taxes, USA 1966 historically fails. We only lock structure and
// the deterministic outcome text format here (the figures are data, not
// tuning).
func TestVintages(t *testing.T) {
	pr := Params{Capital: 1_000_000, NeedAnnual: 50000, Years: 40, TaxRate: 0.30}
	res := Vintages(pr, nil)
	if !strings.Contains(res.SVG, "<svg") {
		t.Fatalf("no SVG rendered")
	}
	if len(res.Cards) != len(vintageList) {
		t.Fatalf("want %d vintage cards, got %d", len(vintageList), len(res.Cards))
	}
	for _, c := range res.Cards {
		ok := strings.Contains(c.Value, "ruined in year") ||
			strings.Contains(c.Value, "survived all") ||
			strings.Contains(c.Value, "solvent when the record ends")
		if !ok {
			t.Errorf("card %q has unexpected verdict %q", c.Label, c.Value)
		}
		if c.Help == "" {
			t.Errorf("card %q missing its story hover", c.Label)
		}
	}
	// USA 2000 + 40y overruns the 2020 record end: the note must say so.
	if !strings.Contains(res.Note, "record ends") {
		t.Errorf("truncation note missing: %q", res.Note)
	}
	// Determinism: same params, same result.
	if again := Vintages(pr, nil); again.SVG != res.SVG {
		t.Errorf("vintage replay is not deterministic")
	}
}

// The decade decomposition must show ruin concentrated in the worst
// first-decade quintile: the whole point of the visual.
func TestDecade(t *testing.T) {
	pr := Params{Capital: 1_200_000, NeedAnnual: 55000, Years: 40,
		Mu: 0.05, Sigma: 0.12, Df: 5, TaxRate: 0.30, NPaths: 3000}
	res := Decade(pr, nil)
	if !strings.Contains(res.SVG, "<svg") {
		t.Fatalf("no SVG rendered")
	}
	if len(res.Cards) != 4 {
		t.Fatalf("want 4 cards, got %d", len(res.Cards))
	}
	var worst, best string
	for _, c := range res.Cards {
		switch c.Label {
		case "Ruin, worst first decade":
			worst = c.Value
		case "Ruin, best first decade":
			best = c.Value
		}
	}
	w, b := pctVal(t, worst), pctVal(t, best)
	if w <= b {
		t.Errorf("worst-decade ruin %v%% must exceed best-decade %v%%", w, b)
	}
	// The concentration is the point of the visual: a bad first decade must
	// carry at least twice the ruin of a good one (in practice far more).
	if b*2 > w {
		t.Errorf("sequence-risk concentration too weak: worst %v%% vs best %v%%", w, b)
	}
}

// pctVal parses "37%" into 37.
func pctVal(t *testing.T, s string) float64 {
	t.Helper()
	v, err := strconv.ParseFloat(strings.TrimSuffix(s, "%"), 64)
	if err != nil {
		t.Fatalf("bad pct %q: %v", s, err)
	}
	return v
}

// The income layers must stack pension and side income under the portfolio
// and count the gap years.
func TestIncome(t *testing.T) {
	pr := Params{Capital: 1_500_000, NeedAnnual: 60000, Years: 40,
		Mu: 0.05, Sigma: 0.11, Df: 5, TaxRate: 0.30, NPaths: 1000,
		PensionAnnual: 20000, PensionYear: 15, SideAnnual: 10000, SideUntilYear: 5}
	res := Income(pr, nil)
	if !strings.Contains(res.SVG, "Pension") || !strings.Contains(res.SVG, "Side income") ||
		!strings.Contains(res.SVG, "Portfolio withdrawals") {
		t.Errorf("income layers missing from the chart")
	}
	found := false
	for _, c := range res.Cards {
		if c.Label == "Years carried by the portfolio alone" && c.Value == "15 y" {
			found = true
		}
	}
	if !found {
		t.Errorf("gap-years card wrong: %+v", res.Cards)
	}
}

// roundShares100 must always sum to 100 for non-zero inputs and never distort
// a share by a full point.
func TestRoundShares100(t *testing.T) {
	cases := [][]float64{
		{0.005, 0.19, 0.805}, // the 1/19/81 rounding trap
		{1, 1, 1},
		{0.3334, 0.3333, 0.3333},
		{0, 0.5, 0.5},
	}
	for _, c := range cases {
		out := roundShares100(c...)
		sum := 0
		for _, v := range out {
			sum += v
		}
		if sum != 100 {
			t.Errorf("roundShares100(%v) = %v, sums to %d", c, out, sum)
		}
	}
	for _, v := range roundShares100(0, 0, 0) {
		if v != 0 {
			t.Errorf("all-zero shares must stay zero")
		}
	}
}

// fmtWealth switches units at the million.
func TestFmtWealth(t *testing.T) {
	cases := map[float64]string{
		999000:  "999 k€",
		1000000: "1.00 M€",
		5451800: "5.45 M€",
		0:       "0 k€",
	}
	for in, want := range cases {
		if got := fmtWealth(in); got != want {
			t.Errorf("fmtWealth(%v) = %q, want %q", in, got, want)
		}
	}
}
