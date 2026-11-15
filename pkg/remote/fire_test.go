package remote

import (
	"context"
	"encoding/json"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/fire"
	"github.com/bpineau/pofo/pkg/marketdata"
)

const parametric = `"market":{"mode":"parametric","real_mean_return_pct":4,"annual_volatility_pct":15,"degrees_of_freedom":5}`

func TestSimulationAssumptionsReadable(t *testing.T) {
	r, err := New(nil, "test").Execute(context.Background(), "simulate_fire", []byte(`{"plan":{"capital":800000,"annual_spending":30000,"horizon_years":2},`+parametric+`,"paths":100,"models":["central","stress","broad_sample","lost_decade"]}`))
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range r.Meta.Defaulted {
		if d.Path == "plan.tax_rate_pct" && d.Value != 32.8 {
			t.Errorf("tax default: %#v", d.Value)
		}
	}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "32.800000000000004") || strings.Contains(string(b), "sliders") {
		t.Errorf("unreadable projection: %s", b)
	}
}

func TestSimulationDirectParityAndDefaults(t *testing.T) {
	s := New(nil, "test")
	r, err := s.Execute(context.Background(), "simulate_fire", []byte(`{"plan":{"capital":800000,"annual_spending":30000,"horizon_years":20,"tax_rate_pct":0,"buffer_years":0},`+parametric+`,"paths":100,"sections":["annual_bands","outcome_metrics","safe_spending"]}`))
	if err != nil {
		t.Fatal(err)
	}
	a := r.Meta.Assumptions.(FireAssumptions)
	if *a.Plan.TaxRatePct != 0 || *a.Plan.BufferYears != 0 {
		t.Fatal("explicit zero replaced")
	}
	for _, d := range r.Meta.Defaulted {
		if d.Path == "plan.tax_rate_pct" || d.Path == "plan.buffer_years" {
			t.Fatal("explicit zero defaulted")
		}
	}
	p := fire.Defaults()
	p.Capital = 800000
	p.AnnualSpending = 30000
	p.Years = 20
	p.TaxRate = 0
	p.BufferYears = 0
	plan, err := fire.BuildPlan(p)
	if err != nil {
		t.Fatal(err)
	}
	model, err := fire.PrepareModel("central", fire.MarketSpec{Mu: .04, Sigma: .15, Nu: 5}, 20)
	if err != nil {
		t.Fatal(err)
	}
	plan.Source = model.Source
	e := plan.Simulate(100, 1, 7)
	o := e.Outcome()
	d := r.Data.(SimulationData).Models[0]
	if float64(d.RuinPct) != o.RuinProb*100 || float64(d.TerminalRealWealthP50) != o.TerminalP50 {
		t.Fatal("simulation differs from direct library")
	}
	if len(d.AnnualBands) != 21 || d.SafeSpending == nil || d.SafeSpending.SolvedPolicy != "fixed_real" {
		t.Fatalf("missing sections: %+v", d)
	}
	if _, err = json.Marshal(r); err != nil {
		t.Fatal(err)
	}
}
func TestSolverReplayAndStatuses(t *testing.T) {
	s := New(nil, "test")
	for _, axis := range []string{"capital", "annual_spending"} {
		amount := `"capital":800000`
		if axis == "capital" {
			amount = `"annual_spending":30000`
		}
		r, err := s.Execute(context.Background(), "solve_fire", []byte(`{"plan":{`+amount+`,"horizon_years":20},`+parametric+`,"solve_for":"`+axis+`","paths":100,"target_ruin_pct":0}`))
		if err != nil {
			t.Fatal(err)
		}
		d := r.Data.(SolveData)
		if d.TargetRuinPct != 0 || d.EffectivePlan == nil {
			t.Fatalf("%+v", d)
		}
		b, _ := json.Marshal(SimulationInput{Plan: &d.EffectivePlan.PlanInput, Market: &MarketInput{Mode: "parametric", RealMeanReturnPct: ptr(4.), AnnualVolatilityPct: ptr(15.), DegreesOfFreedom: ptr(5.)}, Paths: ptr(100)})
		replay, err := s.Execute(context.Background(), "simulate_fire", b)
		if err != nil {
			t.Fatal(err)
		}
		if replay.Data.(SimulationData).Models[0].RuinPct != d.RuinPct {
			t.Fatal("solution fails direct replay")
		}
		if d.Status == "solved" && d.BracketWidth > d.Tolerance {
			t.Fatal("precision")
		}
	}
	for _, tc := range []struct {
		target float64
		lo, hi float64
		status string
	}{{100, 0, 100000, "boundary_satisfies"}, {0, 1e9, 1e10, "target_unreachable_in_bounds"}} {
		in := SolveInput{Plan: &PlanInput{Capital: ptr(1000.), HorizonYears: ptr(20)}, Market: &MarketInput{Mode: "parametric", RealMeanReturnPct: ptr(0.), AnnualVolatilityPct: ptr(0.), DegreesOfFreedom: ptr(0.)}, SolveFor: "annual_spending", TargetRuinPct: &tc.target, Bounds: &Bounds{&tc.lo, &tc.hi}, Paths: ptr(100)}
		b, _ := json.Marshal(in)
		r, err := s.Execute(context.Background(), "solve_fire", b)
		if err != nil {
			t.Fatal(err)
		}
		d := r.Data.(SolveData)
		if d.Status != tc.status {
			t.Fatalf("%+v", d)
		}
		if tc.status == "target_unreachable_in_bounds" && !math.IsNaN(float64(d.Solution)) {
			t.Fatal("fabricated solution")
		}
	}
}
func TestFIREValidationBudgetAndCancellation(t *testing.T) {
	s := New(nil, "test")
	for _, body := range []string{
		`{"plan":{"capital":800000,"annual_spending":30000,"horizon_years":20},` + parametric + `,"models":["historical"]}`,
		`{"plan":{"capital":800000,"annual_spending":30000,"horizon_years":20},` + parametric + `,"safe_spending":{}}`,
		`{"plan":{"capital":800000,"annual_spending":30000,"horizon_years":100},` + parametric + `,"paths":10000,"models":["central","stress","lost_decade","broad_sample"],"sections":["safe_spending"]}`,
	} {
		if _, err := s.Execute(context.Background(), "simulate_fire", []byte(body)); err == nil {
			t.Errorf("accepted %s", body)
		}
	}
	_, err := s.Execute(context.Background(), "solve_fire", []byte(`{"plan":{"annual_spending":30000,"horizon_years":20,"policy":"guyton_klinger"},`+parametric+`,"solve_for":"capital"}`))
	if err == nil || publicError(err).Code != "unsupported_configuration" {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = s.Execute(ctx, "simulate_fire", []byte(`{}`)); err == nil || publicError(err).Code != "cancelled" {
		t.Fatal(err)
	}
}
func TestAllPortfolioModelsAndShortDecade(t *testing.T) {
	src := fixtureSource()
	inflation := &marketdata.Series{Symbol: "^HICP-FR", Source: "fixture"}
	for k := 0; k < 84; k++ {
		inflation.Points = append(inflation.Points, marketdata.Point{Date: time.Date(2018, time.Month(1+k), 1, 0, 0, 0, 0, time.UTC), Close: 100 + float64(k)/10})
	}
	src["^HICP-FR"] = inflation
	s := New(src, "test")
	r, err := s.Execute(context.Background(), "simulate_fire", []byte(`{"plan":{"capital":800000,"annual_spending":30000,"horizon_years":2},"market":{"mode":"portfolio","holdings":[{"id":"AAA","weight_pct":100}]},"models":["central","bootstrap","historical","stress","broad_sample","lost_decade"],"paths":100,"sections":["outcome_metrics"]}`))
	if err != nil {
		t.Fatal(err)
	}
	models := r.Data.(SimulationData).Models
	if len(models) != 6 {
		t.Fatal(len(models))
	}
	for _, m := range models {
		if !math.IsNaN(float64(m.OutcomeMetrics.WorstDecadeP5Pct)) {
			t.Fatal("missing decade became zero")
		}
		if m.Model.ID == "historical" && (m.Model.Sampling != "sampled_cohorts" || m.Model.AvailableCohorts < 1) {
			t.Fatal("incorrect historical sampling")
		}
		if m.Model.ID == "broad_sample" && m.Model.Basis != "external_reference" {
			t.Fatal("misidentified broad sample")
		}
	}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("six detailed models: %d bytes", len(b))
}

func BenchmarkLargeSimulation(b *testing.B) {
	service := New(nil, "benchmark")
	body := []byte(`{"plan":{"capital":800000,"annual_spending":30000,"horizon_years":100},` + parametric + `,"paths":10000,"models":["central","stress","broad_sample","lost_decade"],"sections":["annual_bands","outcome_metrics"]}`)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := service.Execute(context.Background(), "simulate_fire", body); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLargeSolve(b *testing.B) {
	service := New(nil, "benchmark")
	body := []byte(`{"plan":{"capital":800000,"horizon_years":100},` + parametric + `,"solve_for":"annual_spending","paths":10000,"bounds":{"min":0,"max":10000000000}}`)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := service.Execute(context.Background(), "solve_fire", body); err != nil {
			b.Fatal(err)
		}
	}
}

func TestRunningSimulationCancellation(t *testing.T) {
	service := New(nil, "test")
	revision() // Initialize the dataset fingerprint before timing active cancellation.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, err := service.Execute(ctx, "simulate_fire", []byte(`{"plan":{"capital":800000,"annual_spending":30000,"horizon_years":100},`+parametric+`,"paths":10000,"models":["central","stress","broad_sample","lost_decade"]}`))
		done <- err
	}()
	time.Sleep(time.Millisecond)
	start := time.Now()
	cancel()
	select {
	case err := <-done:
		if err == nil || publicError(err).Code != "cancelled" {
			t.Fatalf("cancelled computation returned %v", err)
		}
		t.Logf("cancellation latency: %s", time.Since(start))
	case <-time.After(time.Second):
		t.Fatal("cancelled computation kept running")
	}
}
