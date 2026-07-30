package metrics

import (
	"math"
	"testing"
	"time"
)

func day(i int) time.Time {
	return time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, i)
}

func days(n int) []time.Time {
	out := make([]time.Time, n)
	for i := range out {
		out[i] = day(i)
	}
	return out
}

func near(t *testing.T, name string, got, want, tol float64) {
	t.Helper()
	if math.Abs(got-want) > tol {
		t.Errorf("%s = %v, attendu %v (±%v)", name, got, want, tol)
	}
}

func TestComputeCAGRExactYear(t *testing.T) {
	// Exactly one year (365.25 days) and a doubling: CAGR must be 100 %.
	start := day(0)
	dates := []time.Time{start, start.Add(time.Duration(365.25 * 24 * float64(time.Hour)))}
	s, err := Compute(dates, []float64{100, 200})
	if err != nil {
		t.Fatal(err)
	}
	near(t, "CAGR", s.CAGR, 1.0, 1e-9)
}

func TestComputeRiskMetrics(t *testing.T) {
	// Returns: +10 %, then −10 %.
	s, err := Compute(days(3), []float64{100, 110, 99})
	if err != nil {
		t.Fatal(err)
	}
	near(t, "Volatilité", s.Volatility, math.Sqrt(0.02)*math.Sqrt(252), 1e-9)
	near(t, "Sharpe", s.Sharpe, 0, 1e-9)
	near(t, "Sortino", s.Sortino, 0, 1e-9)
	near(t, "MaxDrawdown", s.MaxDrawdown, -0.10, 1e-12)
	// Drawdowns en %: 0, 0, −10 → Ulcer = sqrt(100/3).
	near(t, "Ulcer", s.Ulcer, math.Sqrt(100.0/3), 1e-9)
	near(t, "CAGR", s.CAGR, math.Pow(0.99, 365.25/2)-1, 1e-9)
}

func TestComputeTTR(t *testing.T) {
	s, err := Compute(days(5), []float64{100, 120, 90, 95, 130})
	if err != nil {
		t.Fatal(err)
	}
	if s.TTRDays != 3 || s.TTROngoing {
		t.Errorf("TTR = %d j (en cours: %v), attendu 3 j récupérés", s.TTRDays, s.TTROngoing)
	}
	near(t, "MaxDrawdown", s.MaxDrawdown, -0.25, 1e-12)
}

func TestComputeTTROngoing(t *testing.T) {
	s, err := Compute(days(4), []float64{100, 120, 90, 95})
	if err != nil {
		t.Fatal(err)
	}
	if s.TTRDays != 2 || !s.TTROngoing {
		t.Errorf("TTR = %d j (en cours: %v), attendu 2 j en cours", s.TTRDays, s.TTROngoing)
	}
}

func TestComputeErrors(t *testing.T) {
	if _, err := Compute(days(1), []float64{100}); err == nil {
		t.Error("erreur attendue: série trop courte")
	}
	if _, err := Compute(days(2), []float64{100, -5}); err == nil {
		t.Error("erreur attendue: valeur négative")
	}
}

func TestBetaTwiceTheBenchmark(t *testing.T) {
	// The portfolio's daily returns are exactly twice the benchmark's, so
	// the regression slope must be 2.
	n := 40
	dates := days(n)
	bench := make([]float64, n)
	port := make([]float64, n)
	bench[0], port[0] = 100, 100
	for i := 1; i < n; i++ {
		r := 0.01 * float64(i%5-2) // -2 %, -1 %, 0, +1 %, +2 % en cycle
		bench[i] = bench[i-1] * (1 + r)
		port[i] = port[i-1] * (1 + 2*r)
	}
	beta, ok := Beta(dates, port, dates, bench)
	if !ok {
		t.Fatal("Beta devrait être calculable")
	}
	near(t, "Beta", beta, 2.0, 1e-9)
}

func TestBetaTooFewOverlaps(t *testing.T) {
	if _, ok := Beta(days(5), []float64{1, 2, 3, 4, 5}, days(5), []float64{1, 2, 3, 4, 5}); ok {
		t.Error("Beta ne devrait pas être calculé avec si peu de points")
	}
}
