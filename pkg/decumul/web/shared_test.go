package web

import (
	"encoding/json"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/bpineau/pofo/pkg/decumul"
	"github.com/bpineau/pofo/pkg/fire"
	"github.com/bpineau/pofo/pkg/scenario"
)

func TestSharedPlanKernelParity(t *testing.T) {
	for _, capital := range []float64{100000, 1000000} {
		for _, gain := range []float64{0, .5} {
			for _, buffer := range []float64{0, 3, 10} {
				for _, guard := range []bool{false, true} {
					pr := Params{Capital: capital, NeedAnnual: 40000, Years: 30, BufferYears: buffer, BufferReturn: .005, TaxRate: .328, GainFrac: gain, Inflation: .02, Guardrails: guard, PensionAnnual: 12000, PensionYear: 5, PensionDrift: -.005}
					policy := fire.FixedReal
					if guard {
						policy = fire.GuytonKlinger
					}
					shared, err := fire.BuildPlan(fire.PlanSpec{Capital: capital, AnnualSpending: 40000, Years: 30, BufferYears: buffer, BufferRealReturn: .005, TaxRate: .328, InitialGain: gain, Inflation: .02, Policy: policy, Pension: &fire.Pension{AnnualAmount: 12000, StartYear: 5, RealDrift: -.005}})
					if err != nil {
						t.Fatal(err)
					}
					web := pr.plan()
					src := scenario.ParametricSource{Mu: .04, Sigma: .15, Df: 5, Periods: 30}
					for seed := range uint64(4) {
						seq := src.Draw(rand.New(rand.NewPCG(seed, 7)))
						if !reflect.DeepEqual(shared.RunPath(seq, decumul.Lives{}), web.RunPath(seq, decumul.Lives{})) {
							t.Fatalf("kernel differs capital=%v gain=%v buffer=%v guard=%v", capital, gain, buffer, guard)
						}
					}
				}
			}
		}
	}
}

func TestSharedModelSourcesParity(t *testing.T) {
	panel := monthlyPanel(600)
	pr := Params{Mu: .06, Sigma: .15, Df: 7, Years: 30, Weights: []float64{1}}
	sources := modelSources(pr, &panel)
	ids := []string{fire.Historical, fire.Bootstrap, fire.Central, fire.Stress, fire.BroadSample, fire.LostDecade}
	if len(sources) != len(ids) {
		t.Fatalf("sources %d", len(sources))
	}
	for i, id := range ids {
		m, err := fire.PrepareModel(id, fire.MarketSpec{Mu: pr.Mu, Sigma: pr.Sigma, Nu: pr.Df, Panel: &panel, Weights: pr.Weights}, pr.Years)
		if err != nil {
			t.Fatal(err)
		}
		rng := func() *rand.Rand { return rand.New(rand.NewPCG(3, 9)) }
		if !reflect.DeepEqual(sources[i].source.Draw(rng()), m.Source.Draw(rng())) {
			t.Fatalf("%s draws differ", id)
		}
	}
}

func TestBrowserUsesSharedDefaults(t *testing.T) {
	w := httptest.NewRecorder()
	Handler(nil, nil).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	b, _ := json.Marshal(sharedDefaults())
	if !strings.Contains(w.Body.String(), "const FIRE_DEFAULTS = "+string(b)) {
		t.Fatal("shared defaults not injected")
	}
	defaults := fire.Defaults()
	for key, want := range map[string]float64{"bufferYears": defaults.BufferYears, "bufferReturn": defaults.BufferRealReturn, "taxRate": defaults.TaxRate, "gainFrac": defaults.InitialGain, "inflation": defaults.Inflation} {
		if sharedDefaults()[key] != want {
			t.Fatalf("%s default differs", key)
		}
	}
}

func TestUndefinedPolicyVolatilityOmitted(t *testing.T) {
	r := PolicyFrontier(Params{Capital: 1000000, NeedAnnual: 10000, PensionAnnual: 12000, Years: 20, Mu: .04, Sigma: .15, Df: 5, NPaths: 20}, nil)
	if strings.Contains(r.SVG, "NaN") || strings.Contains(r.SVG, "Inf") {
		t.Fatal("undefined chart coordinate")
	}
	if !strings.Contains(r.Note, "Undefined spending volatility") {
		t.Fatalf("missing omission note: %s", r.Note)
	}
}

func TestUnreachableCurvesHaveNoteAndFiniteCharts(t *testing.T) {
	r := Curves(Params{Capital: 600000, NeedAnnual: 60000, Years: 60, Mu: .01, Sigma: .20, Df: 3, TaxRate: .4, NPaths: 100, TargetRuin: .000001}, nil)
	if r.Note == "" {
		t.Fatal("expected limited capital searches")
	}
	if strings.Contains(r.HorizonSVG+r.CapitalSVG, "NaN") || strings.Contains(r.HorizonSVG+r.CapitalSVG, "Inf") {
		t.Fatal("undefined curve coordinate")
	}
}
