package metrics

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
)

func TestIdenticalBenchmarkInformationRatioUndefined(t *testing.T) {
	d := days(50)
	v := make([]float64, len(d))
	for i := range v {
		v[i] = 100 + float64(i%7)
	}
	r, ok := VsBenchmark(d, v, d, v)
	if !ok {
		t.Fatal("valid comparison rejected")
	}
	if !math.IsNaN(r.InfoRatio) {
		t.Errorf("identical benchmark information ratio = %v, want NaN", r.InfoRatio)
	}
	b, err := json.Marshal(r)
	if err != nil || !strings.Contains(string(b), `"InfoRatio":null`) {
		t.Fatalf("JSON = %s, %v", b, err)
	}
}

func TestConstantReferenceTrackingUndefined(t *testing.T) {
	a, b := []float64{.02, -.01, .03, .01}, []float64{0, 0, 0, 0}
	r, err := Track(a, b, 12)
	if err != nil {
		t.Fatal(err)
	}
	for name, x := range map[string]float64{"Corr": r.Corr, "Beta": r.Beta, "Alpha": r.Alpha, "VolRatio": r.VolRatio} {
		if !math.IsNaN(x) {
			t.Errorf("%s = %v, want NaN", name, x)
		}
	}
	if !(r.TrackingError > 0) || r.VolB != 0 {
		t.Fatalf("defined measurements lost: %+v", r)
	}
	data, err := json.Marshal(r)
	if err != nil || !strings.Contains(string(data), `"Corr":null`) {
		t.Errorf("JSON = %s, %v", data, err)
	}
}

func TestNonzeroConstantMoments(t *testing.T) {
	a := []float64{.1, .1, .1}
	if Mean(a) != a[0] || Volatility(a, 12) != 0 || !math.IsNaN(Corr(a, a)) || !math.IsNaN(Skewness(a)) {
		t.Fatal("roundoff created variation in a constant sample")
	}
}

func TestZeroTotalContributionHasUndefinedReturnShares(t *testing.T) {
	a, err := Attribute([][]float64{{.125, -.125}, {0, 0}})
	if err != nil {
		t.Fatal(err)
	}
	if a.Risk[0] != 1 || a.Risk[1] != 0 {
		t.Fatalf("lost measurable risk: %+v", a)
	}
	for _, v := range a.Return {
		if !math.IsNaN(v) {
			t.Fatalf("zero denominator became a measured return share: %+v", a)
		}
	}
	b, err := json.Marshal(a)
	if err != nil || !strings.Contains(string(b), `"Return":[null,null]`) {
		t.Fatalf("JSON=%s err=%v", b, err)
	}
}
