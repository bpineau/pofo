package firebook

import (
	"math"
	"strings"
	"testing"
)

// refUnroll re-derives the plate's unroll from scratch, written the other way
// round on purpose: real returns computed first, then compounded, instead of
// the render's nominal-and-prices form. Both must agree to the last decimal.
func refUnroll(inflation []float64) (price, rate []float64) {
	const nominal = (1+inflRealPlan)*(1+inflExpected) - 1
	value, level := inflCapital, 100.0
	for _, pi := range inflation {
		rr := (1+nominal)/(1+pi) - 1
		price = append(price, level)
		rate = append(rate, inflSpend/value*100)
		value = (value - inflSpend) * (1 + rr)
		level *= 1 + pi
	}
	return append(price, level), append(rate, inflSpend/value*100)
}

// The drift is SOLVED, not chosen: after thirty years it must land the price
// level exactly where the five-year episode lands it. That equality is the
// whole claim of the plate, so it is checked to a hair.
func TestInflationDriftLandsOnTheEpisode(t *testing.T) {
	d := inflDrift()
	if d <= inflExpected || d >= inflExpected+0.02 {
		t.Fatalf("drift %.4f is not a point-ish above the planned %.4f", d, inflExpected)
	}
	// The article calls it "un point de plus que prévu": check the solve agrees.
	if got := d - inflExpected; math.Abs(got-0.01) > 0.001 {
		t.Errorf("drift is %.4f above plan, the article says about one point", got)
	}
	episode, drift, planned := inflPaths()
	epPrice, _ := inflUnroll(episode)
	drPrice, _ := inflUnroll(drift)
	plPrice, _ := inflUnroll(planned)
	if len(epPrice) != inflYears+1 || len(drPrice) != inflYears+1 {
		t.Fatalf("unrolls of %d and %d readings, expected %d", len(epPrice), len(drPrice), inflYears+1)
	}
	if gap := math.Abs(epPrice[inflYears] - drPrice[inflYears]); gap > 1e-9 {
		t.Errorf("the two price paths end %.3g apart, they must meet", gap)
	}
	if epPrice[inflYears] <= plPrice[inflYears] {
		t.Errorf("both paths must overshoot the plan: %.2f against %.2f",
			epPrice[inflYears], plPrice[inflYears])
	}
	// Same price product and the same nominal return means the same thirty-year
	// real market return: only the ORDER of the real years differs.
	nominal := inflNominal()
	product := func(path []float64) float64 {
		out := 1.0
		for _, pi := range path {
			out *= (1 + nominal) / (1 + pi)
		}
		return out
	}
	if gap := math.Abs(product(episode) - product(drift)); gap > 1e-12 {
		t.Errorf("the two worlds earn different total real returns (%.3g apart)", gap)
	}
}

// The two price curves meet; the two warning lights must not. This is the
// figure's second half, and the reason it is worth a plate.
func TestInflationWithdrawalRatesDoNotConverge(t *testing.T) {
	episode, drift, planned := inflPaths()
	for _, tc := range []struct {
		name string
		path []float64
	}{{"épisode", episode}, {"dérive", drift}, {"plan", planned}} {
		gotPrice, gotRate := inflUnroll(tc.path)
		wantPrice, wantRate := refUnroll(tc.path)
		for i := range gotRate {
			if math.Abs(gotRate[i]-wantRate[i]) > 1e-9 || math.Abs(gotPrice[i]-wantPrice[i]) > 1e-9 {
				t.Fatalf("%s, year %d: unroll disagrees with the reference derivation", tc.name, i)
			}
		}
		if math.Abs(gotRate[0]-4.0) > 1e-9 {
			t.Errorf("%s starts at %.3f %%, the plan is a 4 %% withdrawal", tc.name, gotRate[0])
		}
	}
	_, epRate := inflUnroll(episode)
	_, drRate := inflUnroll(drift)
	if epRate[inflYears] < 2*drRate[inflYears] {
		t.Errorf("the two rates end at %.1f and %.1f %%: too close to carry the plate",
			epRate[inflYears], drRate[inflYears])
	}
	// The gap opens during the episode and never closes again.
	prev := 0.0
	for i := 0; i <= inflYears; i++ {
		gap := epRate[i] - drRate[i]
		if gap < prev-1e-9 {
			t.Errorf("year %d: the gap narrows (%.3f after %.3f)", i, gap, prev)
		}
		prev = gap
	}
	// The alert band is the one bengen-falaise draws: entered at year 19 with
	// eleven years still to fund, against the drift's very last reading.
	if got := inflAlertYear(epRate); got != 19 {
		t.Errorf("the episode enters the alert band at year %d, the plate says 19", got)
	}
	if got := inflAlertYear(drRate); got != inflYears-1 {
		t.Errorf("the drift enters the alert band at year %d, the plate leans on it being the last ones", got)
	}
}

// Every number the plate prints has to be the unroll's own, and the plate has
// to obey the book's rendering rules (no rgba, no opacity, no em-dash).
func TestInflationPlatePrintsTheUnroll(t *testing.T) {
	episode, drift, planned := inflPaths()
	epPrice, epRate := inflUnroll(episode)
	drPrice, drRate := inflUnroll(drift)
	plPrice, plRate := inflUnroll(planned)
	svg := figInflationEpisodeDerive()

	want := []string{
		frNum(inflDrift()*100, 2) + " %",              // "2,98 %", the solved drift
		"×" + frNum(epPrice[inflYears]/100, 2),        // the meeting point
		"×" + frNum(plPrice[inflYears]/100, 2),        // the plan's own price level
		frNum(epRate[inflYears], 1) + " %",            // the episode's last reading
		frNum(drRate[inflYears], 1) + " %",            // the drift's last reading
		frNum(plRate[inflYears], 1) + " %",            // the plan's last reading
		"à l'année 19, avec 11 ans encore à financer", // the alert crossing
		"6,1 %",                       // the nominal return both worlds earn
		"la zone d'alerte : 8 à 10 %", // the band bengen-falaise names
		"l'épisode : cinq ans à 8 %, puis retour à 2 %",                        // the article's own episode
		"600 k€, 24 k€ par an indexés sur les prix (4,0 % au départ), 30 ans.", // the plan
	}
	for _, s := range want {
		if !strings.Contains(svg, s) {
			t.Errorf("the plate does not print %q", s)
		}
	}
	if got := math.Round(inflNominal()*1000) / 10; math.Abs(got-6.1) > 1e-9 {
		t.Errorf("the nominal return rounds to %.1f %%, the plate prints 6,1 %%", got)
	}
	// One label carries both price paths, so they must round to one number.
	if frNum(drPrice[inflYears]/100, 2) != frNum(epPrice[inflYears]/100, 2) {
		t.Errorf("the two price ends print differently: %.6f against %.6f",
			drPrice[inflYears], epPrice[inflYears])
	}
	for _, banned := range []string{"rgba", "opacity", "—", "linearGradient"} {
		if strings.Contains(svg, banned) {
			t.Errorf("the plate uses %q, which the book bans", banned)
		}
	}
}
