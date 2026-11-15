package fire

import (
	"math"
	"math/rand/v2"
	"reflect"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/scenario"
)

func monthlyFixture(n int) *scenario.Panel {
	r := make([]float64, n)
	for i := range r {
		r[i] = .005 + .02*math.Sin(float64(i))
	}
	return &scenario.Panel{Returns: [][]float64{r}, Weights: []float64{1}}
}

func TestPrepareModelSources(t *testing.T) {
	m := MarketSpec{Mu: .04, Sigma: .15, Nu: 5, Panel: monthlyFixture(360)}
	for _, id := range []string{Central, Bootstrap, Historical, Stress, BroadSample, LostDecade} {
		r, err := PrepareModel(id, m, 20)
		if err != nil {
			t.Fatal(err)
		}
		if r.Source.Len() != 20 || r.Name == "" || r.Description == "" || r.Basis == "" || r.PeriodsPerYear != 1 {
			t.Fatalf("incomplete %s: %+v", id, r)
		}
		if id == Historical && r.HistoricalCohorts != 121 {
			t.Fatalf("cohorts %d", r.HistoricalCohorts)
		}
	}
	for _, id := range []string{Bootstrap, Historical} {
		if _, err := PrepareModel(id, MarketSpec{}, 20); err == nil {
			t.Fatalf("%s silently lacks history", id)
		}
	}
	if _, err := PrepareModel(Historical, m, 31); err == nil {
		t.Fatal("unavailable historical horizon accepted")
	}
	if _, err := PrepareModel(Central, MarketSpec{Panel: monthlyFixture(23)}, 20); err == nil {
		t.Fatal("short history accepted")
	}
	if _, err := PrepareModel("unknown", m, 20); err == nil {
		t.Fatal("unknown model accepted")
	}
}

func TestCAPEBypassesBlendAndLeavesEmpiricalDraws(t *testing.T) {
	m := MarketSpec{Mu: .08, Sigma: .10, Nu: 8, Panel: monthlyFixture(240), CAPE: true}
	r, err := PrepareModel(Central, m, 40)
	if err != nil {
		t.Fatal(err)
	}
	want, err := CurrentCAPE().ArithmeticMean(m.Sigma)
	if err != nil {
		t.Fatal(err)
	}
	if r.EffectiveMu != want || r.EffectiveSigma != m.Sigma || r.EffectiveNu != m.Nu || !r.CAPEApplied {
		t.Fatalf("CAPE branch blended: %+v", r)
	}
	if r.Caveat == "" {
		t.Fatal("missing short-window caution")
	}
	m.Panel = monthlyFixture(600)
	for _, id := range []string{Bootstrap, Historical, BroadSample} {
		on, err := PrepareModel(id, m, 20)
		if err != nil {
			t.Fatal(err)
		}
		offm := m
		offm.CAPE = false
		off, err := PrepareModel(id, offm, 20)
		if err != nil {
			t.Fatal(err)
		}
		if on.CAPEApplied || !reflect.DeepEqual(on.Source.Draw(rand.New(rand.NewPCG(2, 3))), off.Source.Draw(rand.New(rand.NewPCG(2, 3)))) {
			t.Fatalf("CAPE altered %s", id)
		}
	}
	if _, err := (CAPESnapshot{}).ArithmeticMean(.1); err == nil {
		t.Fatal("missing CAPE substituted")
	}
	now := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	if !CAPEIsStale("malformed", now) || !CAPEIsStale("2020-01-01", now) || CAPEIsStale("2029-12-01", now) {
		t.Fatal("CAPE staleness")
	}
}

func TestCAPEParsingRejectsNonfinite(t *testing.T) {
	got := parseCape([]byte("date,cape\n2020-01-01,NaN\n2020-02-01,+Inf\n2020-03-01,20\n"))
	if len(got) != 1 || got[0].Value != 20 {
		t.Fatalf("malformed CAPE observations retained: %+v", got)
	}
}
