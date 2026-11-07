package metrics

import (
	"encoding/json"
	"math"
	"testing"
	"time"
)

func TestComputeWithConventions(t *testing.T) {
	ds := []time.Time{time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC), time.Date(2020, 3, 1, 0, 0, 0, 0, time.UTC)}
	vs := []float64{100, 110, 99}
	got, err := ComputeWith(ds, vs, ComputeOptions{PeriodsPerYear: 12, RiskFree: .06})
	if err != nil {
		t.Fatal(err)
	}
	want := -.06 / (math.Sqrt(.02) * math.Sqrt(12))
	if math.Abs(got.Sharpe-want) > 1e-12 {
		t.Fatalf("Sharpe %v want %v", got.Sharpe, want)
	}
	for _, opt := range []ComputeOptions{{PeriodsPerYear: -1}, {PeriodsPerYear: math.NaN()}, {RiskFree: math.Inf(1)}} {
		if _, err := ComputeWith(ds, vs, opt); err == nil {
			t.Fatalf("accepted %+v", opt)
		}
	}
	ds[1] = ds[0]
	if _, err := Compute(ds, vs); err == nil {
		t.Fatal("duplicate date accepted")
	}
}

func TestStatsMissingJSON(t *testing.T) {
	st := Stats{Sharpe: math.NaN(), Sortino: math.Inf(1)}
	b, err := json.Marshal(st)
	if err != nil {
		t.Fatal(err)
	}
	var values map[string]any
	if err := json.Unmarshal(b, &values); err != nil {
		t.Fatal(err)
	}
	if values["Sharpe"] != nil || values["Sortino"] != nil || values["CAGR"] != float64(0) {
		t.Fatalf("missing != zero: %s", b)
	}
	b, err = json.Marshal(Relative{UpCapture: math.NaN()})
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, &values); err != nil {
		t.Fatal(err)
	}
	if values["UpCapture"] != nil {
		t.Fatal(string(b))
	}
}

func TestRegressionUndefinedJSON(t *testing.T) {
	r, err := Regress([]float64{1, 1, 1, 1}, []float64{0, 1, 2, 3})
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got["R2"] != nil {
		t.Fatal(string(b))
	}
	estimate := got["Alpha"].(map[string]any)
	if estimate["T"] != nil || estimate["Value"] != float64(1) {
		t.Fatal(string(b))
	}
}
