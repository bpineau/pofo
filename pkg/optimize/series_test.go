package optimize

import (
	"math"
	"testing"

	"github.com/bpineau/pofo/pkg/metrics"
)

// antiCorrelated builds two assets sharing a positive drift but whose dominant
// swing is anti-correlated (it cancels in a balanced blend), each with its own
// idiosyncratic wobble so neither is drawdown-free on its own. A blend cuts the
// downside and the drawdown below either single asset.
func antiCorrelated() (a, b []float64) {
	const n = 300
	a = make([]float64, n)
	b = make([]float64, n)
	for i := 0; i < n; i++ {
		swing := 0.012 * math.Sin(float64(i)*0.3)
		a[i] = 0.0008 + swing + 0.0020*math.Sin(float64(i)*1.7)
		b[i] = 0.0008 - swing + 0.0020*math.Cos(float64(i)*1.3)
	}
	return
}

// TestSolveMaxSortino: the max-Sortino weights beat the better single asset and
// form a genuine blend, since combining the anti-correlated pair cuts downside.
func TestSolveMaxSortino(t *testing.T) {
	a, b := antiCorrelated()
	best := math.Max(metrics.Sortino(a, 0), metrics.Sortino(b, 0))
	res, err := Solve([][]float64{a, b}, Spec{Objective: MaxSortino})
	if err != nil {
		t.Fatal(err)
	}
	if res.Sortino < best-1e-6 {
		t.Fatalf("optimized Sortino %.3f below best single %.3f", res.Sortino, best)
	}
	if res.Weights[0] < 0.1 || res.Weights[1] < 0.1 {
		t.Fatalf("expected a diversified blend, got %.2f / %.2f", res.Weights[0], res.Weights[1])
	}
}

// TestSolveReturnToDrawdown: same, for the return-to-max-drawdown objective.
func TestSolveReturnToDrawdown(t *testing.T) {
	a, b := antiCorrelated()
	ra, _ := metrics.ReturnToMaxDrawdown(a, 0)
	rb, _ := metrics.ReturnToMaxDrawdown(b, 0)
	best := math.Max(ra, rb)
	res, err := Solve([][]float64{a, b}, Spec{Objective: ReturnToDrawdown})
	if err != nil {
		t.Fatal(err)
	}
	if res.ReturnToMaxDD < best-1e-6 {
		t.Fatalf("optimized return-to-drawdown %.3f below best single %.3f", res.ReturnToMaxDD, best)
	}
	if res.Weights[0] < 0.1 || res.Weights[1] < 0.1 {
		t.Fatalf("expected a diversified blend, got %.2f / %.2f", res.Weights[0], res.Weights[1])
	}
}

// TestSolveMinUlcer: the min-Ulcer weights beat the lower-Ulcer single asset
// and form a genuine blend, since combining the anti-correlated pair shortens
// and shallows the underwater stretches.
func TestSolveMinUlcer(t *testing.T) {
	a, b := antiCorrelated()
	best := math.Min(metrics.Ulcer(a), metrics.Ulcer(b))
	res, err := Solve([][]float64{a, b}, Spec{Objective: MinUlcer})
	if err != nil {
		t.Fatal(err)
	}
	if res.Ulcer > best+1e-6 {
		t.Fatalf("optimized Ulcer %.3f above best single %.3f", res.Ulcer, best)
	}
	if res.Weights[0] < 0.1 || res.Weights[1] < 0.1 {
		t.Fatalf("expected a diversified blend, got %.2f / %.2f", res.Weights[0], res.Weights[1])
	}
}

// TestSolveMaxWorst5y: the worst-case five-year return of the blend is at least
// as good as the best single asset's, over a >5-year synthetic history.
func TestSolveMaxWorst5y(t *testing.T) {
	// ~6 years of anti-correlated daily returns.
	const n = 6 * 252
	a := make([]float64, n)
	b := make([]float64, n)
	for i := 0; i < n; i++ {
		swing := 0.010 * math.Sin(float64(i)*0.05)
		a[i] = 0.0004 + swing
		b[i] = 0.0004 - swing + 0.0002*math.Sin(float64(i)*0.9)
	}
	wa, _ := metrics.WorstRollingReturn(a, fiveYearWindow)
	wb, _ := metrics.WorstRollingReturn(b, fiveYearWindow)
	best := math.Max(wa, wb)
	res, err := Solve([][]float64{a, b}, Spec{Objective: MaxWorst5y})
	if err != nil {
		t.Fatal(err)
	}
	if res.Worst5y < best-1e-6 {
		t.Fatalf("optimized worst-5y %.4f below best single %.4f", res.Worst5y, best)
	}
}

func TestSolveMaxWorst5yShortHistory(t *testing.T) {
	a := make([]float64, 252) // only 1 year
	b := make([]float64, 252)
	if _, err := Solve([][]float64{a, b}, Spec{Objective: MaxWorst5y}); err == nil {
		t.Fatalf("expected an error for a history shorter than 5 years")
	}
}

func TestParseSpecSeries(t *testing.T) {
	for _, obj := range []Objective{MaxSortino, ReturnToDrawdown, MinUlcer, MaxWorst5y} {
		s, err := ParseSpec(string(obj) + ",max-weight:60")
		if err != nil {
			t.Fatalf("%s: %v", obj, err)
		}
		if s.Objective != obj || math.Abs(s.MaxWeight-0.6) > 1e-9 {
			t.Fatalf("%s: got %q cap %.2f", obj, s.Objective, s.MaxWeight)
		}
	}
}
