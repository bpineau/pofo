package marketdata

import (
	"math"
	"testing"
	"time"
)

func TestDeflateAndExcess(t *testing.T) {
	ds := []time.Time{date(2020, 1, 1), date(2021, 1, 1), date(2022, 1, 1)}
	s, _ := NewSeries("ASSET", ds, []float64{100, 110, 121})
	s.Currency = "EUR"
	cpi, _ := NewSeries("CPI", ds, []float64{100, 105, 110.25})
	real, err := s.Deflate(cpi, AdjustmentOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(real.Last().Close-109.75056689342404) > 1e-10 {
		t.Fatalf("real end %v", real.Last().Close)
	}
	excess, err := s.Excess(cpi, AdjustmentOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(excess.Last().Close-1.1025) > 1e-12 {
		t.Fatalf("arithmetic excess end %v", excess.Last().Close)
	}
	late := Trim(cpi, ds[1], time.Time{})
	if _, err := s.Deflate(late, AdjustmentOptions{}); err == nil {
		t.Fatal("uncovered CPI accepted")
	}
	approx, err := s.Deflate(late, AdjustmentOptions{AllowExtrapolation: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(approx.Warnings) == 0 {
		t.Fatal("silent CPI extrapolation")
	}
	cpi.Points[1].Close = 0
	if _, err := s.Deflate(cpi, AdjustmentOptions{}); err == nil {
		t.Fatal("zero CPI accepted")
	}
}

func TestCashIndexDayCount(t *testing.T) {
	ds := []time.Time{date(2020, 1, 1), date(2020, 1, 11), date(2020, 1, 21)}
	rate, _ := NewSeries("RATE", []time.Time{ds[0], ds[1]}, []float64{3.6, 7.2})
	out, err := CashIndex(rate, ds, Actual360)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(out.Last().Close-math.Pow(1.0001, 10)*math.Pow(1.0002, 10)) > 1e-12 {
		t.Fatalf("cash end %v", out.Last().Close)
	}
	if _, err := CashIndex(rate, ds, DayCount("")); err == nil {
		t.Fatal("implicit day count accepted")
	}
	if _, err := CashIndex(rate, []time.Time{ds[0].AddDate(0, 0, -1), ds[1]}, Actual360); err == nil {
		t.Fatal("rate missing at start")
	}
	rate.Points[0].Close = math.NaN()
	if _, err := CashIndex(rate, ds, Actual360); err == nil {
		t.Fatal("NaN rate accepted")
	}
}

func TestAdjustmentBoundaries(t *testing.T) {
	ds := []time.Time{date(2024, 1, 1), date(2024, 2, 1), date(2024, 3, 1)}
	s, _ := NewSeries("A", ds, []float64{1, 1.1, 1.2})
	s.Currency = "EUR"
	index, _ := NewSeries("INDEX", []time.Time{ds[0], ds[2]}, []float64{1, 1.03})
	if _, err := s.Deflate(index, AdjustmentOptions{MaxAge: 20 * 24 * time.Hour}); err == nil {
		t.Fatal("stale observation accepted")
	}
	index.Currency = "USD"
	if _, err := s.Excess(index, AdjustmentOptions{}); err == nil {
		t.Fatal("cash currency mismatch")
	}
	index.Currency = "EUR"
	index.Junctions = []time.Time{ds[1]}
	if _, err := s.Deflate(index, AdjustmentOptions{}); err == nil {
		t.Fatal("definition junction accepted")
	}
	index.Junctions = nil
	s.Points[2].Close = 0
	real, err := s.Deflate(index, AdjustmentOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if real.Last().Close != 0 {
		t.Fatal("terminal ruin lost")
	}
	if _, err := s.Excess(index, AdjustmentOptions{}); err == nil {
		t.Fatal("excess loss greater than capital accepted")
	}
	if _, err := (*Series)(nil).Deflate(index, AdjustmentOptions{}); err == nil {
		t.Fatal("nil asset accepted")
	}
	if _, err := s.Deflate(nil, AdjustmentOptions{}); err == nil {
		t.Fatal("nil index accepted")
	}
}

func TestCashIndexSamplingInvariant(t *testing.T) {
	first := date(2024, 1, 1)
	rate, _ := NewSeries("RATE", []time.Time{first, first.AddDate(0, 0, 3), first.AddDate(0, 0, 10)}, []float64{2, -1, 1})
	ds := make([]time.Time, 11)
	for i := range ds {
		ds[i] = first.AddDate(0, 0, i)
	}
	daily, err := CashIndex(rate, ds, Actual365)
	if err != nil {
		t.Fatal(err)
	}
	sparse, err := CashIndex(rate, []time.Time{ds[0], ds[10]}, Actual365)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(daily.Last().Close-sparse.Last().Close) > 1e-14 {
		t.Fatal("cash depends on output calendar")
	}
}
