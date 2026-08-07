package firebook

import (
	"math"
	"strings"
	"testing"
)

// TestPanierContributionsArithmetic redoes the weighted sum the article states
// in prose: six weights, six rates, and a total the household compares with the
// published index. The plate must not print a number the arithmetic denies.
func TestPanierContributionsArithmetic(t *testing.T) {
	// weight, rate and contribution, straight from the article's example.
	want := map[string][3]float64{
		"Services, aide, assurances": {22, 3.0, 0.66},
		"Santé et mutuelle":          {14, 4.5, 0.63},
		"Loisirs et voyages":         {24, 2.0, 0.48},
		"Alimentation":               {18, 2.0, 0.36},
		"Énergie et transport":       {12, 2.5, 0.30},
		"Divers":                     {10, 2.0, 0.20},
	}
	if len(panierPostes) != len(want) {
		t.Fatalf("plate draws %d budget lines, the article states %d", len(panierPostes), len(want))
	}
	weights, total, lead := 0.0, 0.0, 0.0
	prev := math.Inf(1)
	for _, p := range panierPostes {
		w, ok := want[p.name]
		if !ok {
			t.Fatalf("unexpected budget line %q", p.name)
		}
		if p.weight != w[0] || p.rate != w[1] {
			t.Errorf("%s: plate says %.0f %% at %.1f %%, the article says %.0f %% at %.1f %%",
				p.name, p.weight, p.rate, w[0], w[1])
		}
		if got := p.contribution(); math.Abs(got-w[2]) > 1e-9 {
			t.Errorf("%s: contribution %.4f, %.0f %% × %.1f %% = %.4f", p.name, got, p.weight, p.rate, w[2])
		}
		if p.contribution() > prev+1e-9 {
			t.Errorf("%s breaks the plate's order: contributions must decrease", p.name)
		}
		prev = p.contribution()
		weights += p.weight
		total += p.contribution()
		if p.lead {
			lead = p.contribution()
		}
	}
	if math.Abs(weights-100) > 1e-9 {
		t.Errorf("the weights add up to %.1f %%, not to a whole budget", weights)
	}
	if math.Abs(total-2.63) > 1e-9 {
		t.Errorf("the six contributions add up to %.4f, the plate prints 2,63", total)
	}
	// the article rounds that total to ~2,6 % and the gap to the index to +0,5.
	if got := math.Round(total*10) / 10; got != 2.6 {
		t.Errorf("total rounds to %.1f %%, the article says 2,6 %%", got)
	}
	gap := total - panierIPCH
	if got := math.Round(gap*10) / 10; got != 0.5 {
		t.Errorf("gap to the index rounds to %.1f point, the article says +0,5", got)
	}

	// The plate's claim about health: a quarter of the household's inflation
	// for 14 % of the budget, and most of the gap to the index.
	if share := lead / total; share < 0.23 || share > 0.26 {
		t.Errorf("health makes %.1f %% of the total, the plate says a quarter", share*100)
	}
	// health's excess = its contribution minus what the index would have given
	// that same weight (14 % × 2,1 %).
	excess := lead - 0.14*panierIPCH
	if math.Abs(excess-0.336) > 1e-9 {
		t.Errorf("health's excess over the index is %.4f, expected 0,336", excess)
	}
	if excess < gap/2 {
		t.Errorf("health carries %.3f of the %.3f point gap, the plate says most of it", excess, gap)
	}

	svg := figPanierContributions()
	for _, s := range []string{"0,66", "0,63", "0,48", "0,36", "0,30", "0,20", "2,63 %",
		"l'IPCH : 2,1 %", "14 % × 4,5 %", "+0,5 point au-dessus de l'indice"} {
		if !strings.Contains(svg, s) {
			t.Errorf("plate does not print %q", s)
		}
	}
	if strings.Contains(svg, "rgba(") {
		t.Error("plate uses rgba, which crengine paints solid black")
	}
}

// TestEcartComposeAnnees checks the only conversion the second plate makes: a
// price gap that compounds year after year, turned into years of spending.
func TestEcartComposeAnnees(t *testing.T) {
	// closed form: sum_{t=1..n} ((1+d)^t − 1) = (1+d)((1+d)^n − 1)/d − n
	for _, d := range panierDrifts {
		for _, n := range []int{1, 5, 10, 20, 25, 30} {
			p := math.Pow(1+d, float64(n))
			closed := (1+d)*(p-1)/d - float64(n)
			if got := panierManque(d, n); math.Abs(got-closed) > 1e-9 {
				t.Errorf("drift %.3f over %d years: plate sums %.6f, closed form gives %.6f", d, n, got, closed)
			}
		}
	}
	// the three anchor readings the plate marks.
	want := map[float64][3]float64{
		0.003: {0.166494, 0.642133, 1.436318},
		0.005: {0.279167, 1.084011, 2.441417},
	}
	for _, d := range panierDrifts {
		for i, n := range []int{10, 20, 30} {
			if got := panierManque(d, n); math.Abs(got-want[d][i]) > 5e-5 {
				t.Errorf("drift %.3f at %d years: %.5f year of spending, expected %.5f", d, n, got, want[d][i])
			}
		}
	}

	// what the plate writes in words. Year 10, in months of spending.
	lo, hi := panierManque(0.003, 10)*12, panierManque(0.005, 10)*12
	if lo < 1.5 || lo > 2.5 || hi < 2.5 || hi > 3.5 {
		t.Errorf("year 10 costs %.1f to %.1f months, the plate says two to three", lo, hi)
	}
	// The same gap read in index points, the reading the plate refuses: the
	// price level has only diverged by 9 to 16 % after thirty years.
	for i, d := range panierDrifts {
		pct := (math.Pow(1+d, 30) - 1) * 100
		wantPct := []float64{9.40, 16.14}[i]
		if math.Abs(pct-wantPct) > 0.05 {
			t.Errorf("drift %.3f moves the price level by %.2f %%, expected %.2f %%", d, pct, wantPct)
		}
	}
	// and the article's own shortcut, "une année et demie" for +0,3 point.
	if math.Abs(panierManque(0.003, 30)-1.5) > 0.1 {
		t.Errorf("+0,3 point over 30 years costs %.2f year, the article says a year and a half", panierManque(0.003, 30))
	}

	svg := figEcartCompose()
	for _, s := range []string{"1,4 année", "2,4 années", "0,6 année", "1,1 année",
		"+0,3 point/an", "+0,5 point/an", "2 années", "années de retrait", "l'indice officiel"} {
		if !strings.Contains(svg, s) {
			t.Errorf("plate does not print %q", s)
		}
	}
	if strings.Contains(svg, "rgba(") {
		t.Error("plate uses rgba, which crengine paints solid black")
	}
}

// TestPanierArticleStillStatesTheNumbers ties both plates to the prose they
// illustrate: if the article's example is retuned, the plates must follow.
func TestPanierArticleStillStatesTheNumbers(t *testing.T) {
	raw, err := assets.ReadFile("assets/book/fr/suivre-inflation.md")
	if err != nil {
		t.Fatal(err)
	}
	src := string(raw)
	for _, s := range []string{
		"22 % à +3 %", "14 % à +4,5 %", "24 % à +2 %", "18 % à +2 %", "12 % à +2,5 %", "10 % à +2 %",
		"2,6 %", "2,1 %", // the household's inflation and the index it is compared with
		"+0,3 point", "30 ans", // the composed-gap thesis
		"::: figure panier-contributions", "::: figure ecart-compose",
	} {
		if !strings.Contains(src, s) {
			t.Errorf("the article no longer states %q", s)
		}
	}
}
