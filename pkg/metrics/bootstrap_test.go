package metrics

import (
	"encoding/json"
	"math"
	"reflect"
	"strings"
	"testing"
)

func TestBlockBootstrapPointAndReproducibility(t *testing.T) {
	sample := []float64{.1, -.1, .1, -.1}
	opt := BootstrapOptions{PeriodsPerYear: 1, BlockLength: 2, Repetitions: 500, Confidence: .9, Seed: 17}
	a, err := BlockBootstrap(sample, opt)
	if err != nil {
		t.Fatal(err)
	}
	b, err := BlockBootstrap(sample, opt)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Fatal("same seed changed output")
	}
	if math.Abs(a.CAGR.Estimate-(math.Sqrt(.99)-1)) > 1e-14 || math.Abs(a.Volatility.Estimate-math.Sqrt(.04/3)) > 1e-14 || math.Abs(a.Sharpe.Estimate) > 1e-14 {
		t.Fatalf("point estimates: %+v", a)
	}
	if a.CAGR.Lower > a.CAGR.Estimate || a.CAGR.Upper < a.CAGR.Estimate || a.Volatility.Lower < 0 || a.Sharpe.Defined != 500 {
		t.Fatalf("intervals: %+v", a)
	}
	if !reflect.DeepEqual(sample, []float64{.1, -.1, .1, -.1}) {
		t.Fatal("mutated returns")
	}
	// A block spanning the whole circular sample is a rotation; all three
	// order-independent statistics must retain their point estimates.
	opt.BlockLength = len(sample)
	full, err := BlockBootstrap(sample, opt)
	if err != nil {
		t.Fatal(err)
	}
	for _, interval := range []Interval{full.CAGR, full.Volatility, full.Sharpe} {
		if math.Abs(interval.Lower-interval.Estimate) > 1e-14 || math.Abs(interval.Upper-interval.Estimate) > 1e-14 {
			t.Fatalf("full block changed sample statistic: %+v", interval)
		}
	}
}

func TestBlockBootstrapConstantAndJSON(t *testing.T) {
	sample := make([]float64, 25)
	for i := range sample {
		sample[i] = .01
	}
	got, err := BlockBootstrap(sample, BootstrapOptions{PeriodsPerYear: 12, BlockLength: 3, Repetitions: 30, Confidence: .95, Seed: 1})
	if err != nil {
		t.Fatal(err)
	}
	if got.Volatility.Estimate != 0 || got.Volatility.Lower != 0 || got.Volatility.Upper != 0 {
		t.Fatal("constant returns have nonzero volatility")
	}
	if !math.IsNaN(got.Sharpe.Estimate) || !math.IsNaN(got.Sharpe.Lower) || !math.IsNaN(got.Sharpe.Upper) || got.Sharpe.Defined != 0 {
		t.Fatalf("undefined Sharpe erased: %+v", got.Sharpe)
	}
	if math.Abs(got.CAGR.Estimate-(math.Pow(1.01, 12)-1)) > 1e-14 {
		t.Fatal("constant CAGR wrong")
	}
	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(encoded), `"Sharpe":{"Estimate":null,"Lower":null,"Upper":null,"Defined":0}`) {
		t.Fatalf("JSON undefined ratio erased: %s", encoded)
	}
}

func TestBlockBootstrapInvalid(t *testing.T) {
	good := BootstrapOptions{PeriodsPerYear: 12, BlockLength: 2, Repetitions: 20, Confidence: .9}
	for _, sample := range [][]float64{nil, {.1}, {.1, math.NaN()}, {.1, math.Inf(1)}, {.1, -1}} {
		if _, err := BlockBootstrap(sample, good); err == nil {
			t.Fatal("invalid sample accepted")
		}
	}
	for _, change := range []func(*BootstrapOptions){
		func(o *BootstrapOptions) { o.PeriodsPerYear = 0 }, func(o *BootstrapOptions) { o.PeriodsPerYear = math.NaN() },
		func(o *BootstrapOptions) { o.BlockLength = 0 }, func(o *BootstrapOptions) { o.BlockLength = 4 },
		func(o *BootstrapOptions) { o.Repetitions = 1 }, func(o *BootstrapOptions) { o.Repetitions = MaxBootstrapRepetitions + 1 },
		func(o *BootstrapOptions) { o.Confidence = 0 }, func(o *BootstrapOptions) { o.Confidence = 1 }, func(o *BootstrapOptions) { o.Confidence = math.NaN() },
		func(o *BootstrapOptions) { o.RiskFree = math.Inf(1) },
	} {
		opt := good
		change(&opt)
		if _, err := BlockBootstrap([]float64{.1, -.1, .2}, opt); err == nil {
			t.Fatalf("accepted options %+v", opt)
		}
	}
}

func TestBlockBootstrapTwoPointDistribution(t *testing.T) {
	// IID draws from {-10%,+10%} have equiprobable paths (--),(-+),(+-),(++).
	// The central 80% CAGR interval therefore spans -10% to +10%.
	// Only the middle two paths have a defined (zero) Sharpe ratio.
	got, err := BlockBootstrap([]float64{-.1, .1}, BootstrapOptions{PeriodsPerYear: 1, BlockLength: 1, Repetitions: 10000, Confidence: .8, Seed: 29})
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(got.CAGR.Lower+.1) > 1e-14 || math.Abs(got.CAGR.Upper-.1) > 1e-14 {
		t.Fatalf("two-point CAGR interval: %+v", got.CAGR)
	}
	if got.Sharpe.Defined < 4000 || got.Sharpe.Defined > 6000 || got.Sharpe.Lower != 0 || got.Sharpe.Upper != 0 {
		t.Fatalf("undefined replicate count: %+v", got.Sharpe)
	}
	if got.Volatility.Lower != 0 || math.Abs(got.Volatility.Upper-math.Sqrt(.02)) > 1e-14 {
		t.Fatalf("two-point volatility interval: %+v", got.Volatility)
	}
}
