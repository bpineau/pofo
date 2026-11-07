package marketdata

import (
	"fmt"
	"math"
	"slices"
	"time"
)

// AdjustmentOptions controls how a price or cash index is read on another
// series' dates. The zero value refuses extrapolation beyond its coverage.
type AdjustmentOptions struct {
	// AllowExtrapolation holds the first/last index level outside its coverage,
	// recording the affected interval in the result's Warnings.
	AllowExtrapolation bool
	// MaxAge limits the age of a forward-filled observation. Zero sets no age
	// limit within the covered window. A negative duration is invalid.
	MaxAge time.Duration
}

// Deflate returns s in purchasing power at its first date: value[t] times
// index[first]/index[t]. index is a positive price-level index, never an
// inflation rate. It need not share s's currency: the caller chooses the
// consumption basket separately from the investment's evaluation currency.
// Index observations are held forward on s's calendar. The zero options
// refuse missing coverage; permitted extrapolation is recorded in Warnings.
// Dates, cadence, currency and provenance survive; per-share Dividends do not
// belong to the deflated curve and are cleared. Neither input is modified.
func (s *Series) Deflate(index *Series, opt AdjustmentOptions) (*Series, error) {
	levels, warnings, err := adjustmentLevels(s, index, opt)
	if err != nil {
		return nil, err
	}
	out := s.clone()
	out.Dividends = nil
	out.Warnings = append(out.Warnings, warnings...)
	for i := range out.Points {
		out.Points[i].Close *= levels[0] / levels[i]
		if !finite(out.Points[i].Close) {
			return nil, fmt.Errorf("marketdata: deflation overflow on %s", out.Points[i].Date.Format(time.DateOnly))
		}
	}
	return out, nil
}

// Excess returns a base-one curve compounded from s's arithmetic period
// returns minus the cash index's period returns, on s's own calendar.
// cash must be a total-return index in s's currency, not an annual rate
// (CashIndex converts a rate). Unlike Deflate this subtracts returns; it
// does not divide growth factors. A loss of 100 percent or more in the
// resulting excess path is refused. Metadata survives, Dividends are cleared.
func (s *Series) Excess(cash *Series, opt AdjustmentOptions) (*Series, error) {
	if s != nil && cash != nil && s.Currency != "" && cash.Currency != "" && s.Currency != cash.Currency {
		return nil, fmt.Errorf("marketdata: excess currencies differ: %s and %s", s.Currency, cash.Currency)
	}
	levels, warnings, err := adjustmentLevels(s, cash, opt)
	if err != nil {
		return nil, err
	}
	out := s.clone()
	out.Dividends = nil
	out.Warnings = append(out.Warnings, warnings...)
	out.Points[0].Close = 1
	for i := 1; i < s.Len(); i++ {
		growth := 1 + s.Points[i].Close/s.Points[i-1].Close - levels[i]/levels[i-1]
		if !(growth > 0) || !finite(growth) {
			return nil, fmt.Errorf("marketdata: excess path exhausted on %s", s.Points[i].Date.Format(time.DateOnly))
		}
		out.Points[i].Close = out.Points[i-1].Close * growth
		if !finite(out.Points[i].Close) || out.Points[i].Close <= 0 {
			return nil, fmt.Errorf("marketdata: excess accumulation overflow or underflow")
		}
	}
	return out, nil
}

// adjustmentLevels validates both inputs and reads the index at the asset dates.
func adjustmentLevels(s, index *Series, opt AdjustmentOptions) ([]float64, []string, error) {
	if err := validLevels(s, false); err != nil {
		return nil, nil, err
	}
	for i, p := range s.Points {
		if p.Close <= 0 && !(p.Close == 0 && i == s.Len()-1 && i > 0) {
			return nil, nil, fmt.Errorf("marketdata: non-positive asset level on %s", p.Date.Format(time.DateOnly))
		}
	}
	if err := validLevels(index, true); err != nil {
		return nil, nil, err
	}
	if opt.MaxAge < 0 {
		return nil, nil, fmt.Errorf("marketdata: negative maximum observation age")
	}
	if len(s.Junctions) > 0 || len(index.Junctions) > 0 {
		for _, x := range []*Series{s, index} {
			if j, ok := x.junctionIn(s.First().Date, s.Last().Date); ok {
				return nil, nil, fmt.Errorf("marketdata: %s definition junction on %s", x.Symbol, j.Format(time.DateOnly))
			}
		}
	}
	warnings := slices.Clone(index.Warnings)
	before := s.First().Date.Before(index.First().Date)
	after := !knowsLevelAt(index.Last().Date, s.Last().Date)
	if before || after {
		if !opt.AllowExtrapolation {
			return nil, nil, fmt.Errorf("marketdata: %s covers %s to %s, not %s to %s", index.Symbol, index.First().Date.Format(time.DateOnly), index.Last().Date.Format(time.DateOnly), s.First().Date.Format(time.DateOnly), s.Last().Date.Format(time.DateOnly))
		}
		warnings = append(warnings, fmt.Sprintf("%s held constant outside %s to %s", index.Symbol, index.First().Date.Format(time.DateOnly), index.Last().Date.Format(time.DateOnly)))
	}
	levels := make([]float64, s.Len())
	for i, p := range s.Points {
		v, on, ok := index.At(p.Date)
		if !ok {
			v, on = index.First().Close, index.First().Date
		}
		if opt.MaxAge > 0 && p.Date.Sub(on) > opt.MaxAge {
			return nil, nil, fmt.Errorf("marketdata: %s observation on %s is too old for %s", index.Symbol, on.Format(time.DateOnly), p.Date.Format(time.DateOnly))
		}
		levels[i] = v
	}
	return levels, warnings, nil
}

// validLevels checks the public Series representation at a transformation boundary.
func validLevels(s *Series, positiveOnly bool) error {
	if s.Len() == 0 {
		return fmt.Errorf("marketdata: empty series")
	}
	for i, p := range s.Points {
		if !finite(p.Close) || positiveOnly && p.Close <= 0 {
			return fmt.Errorf("marketdata: %s invalid level on %s", s.Symbol, p.Date.Format(time.DateOnly))
		}
		if p.Date != dayUTC(p.Date) || i > 0 && !p.Date.After(s.Points[i-1].Date) {
			return fmt.Errorf("marketdata: %s dates must be increasing UTC civil dates", s.Symbol)
		}
	}
	return nil
}

// DayCount is the denominator used to accrue an annual quoted rate over
// actual calendar days. It does not convert a discount yield to a deposit rate.
type DayCount string

// Explicit day-count bases for CashIndex; no implicit default is selected.
const (
	Actual360   DayCount = "ACT/360"
	Actual365   DayCount = "ACT/365F"
	Actual36525 DayCount = "ACT/365.25"
)

// CashIndex compounds an annual nominal rate, quoted in PERCENT, into a
// base-one total-return cash index on dates. Each calendar day earns the
// rate in force at its start divided by 100 and the explicit day-count
// denominator. Daily compounding makes the answer independent of the output
// sampling calendar and includes all rate changes between output dates.
// dates are normalized to midnight UTC civil dates as in NewSeries, must be
// increasing after normalization, and must be covered at their start by rates.
// Negative rates are allowed as long as a day's capital stays positive.
// The last observed rate remains in force; a tail beyond its observation
// is recorded in Warnings. Currency is copied from rates; when absent the
// caller must supply the denomination before mixing with another currency.
func CashIndex(rates *Series, dates []time.Time, basis DayCount) (*Series, error) {
	if err := validLevels(rates, false); err != nil {
		return nil, err
	}
	denom := map[DayCount]float64{Actual360: 360, Actual365: 365, Actual36525: 365.25}[basis]
	if denom == 0 {
		return nil, fmt.Errorf("marketdata: choose ACT/360, ACT/365F or ACT/365.25")
	}
	values := make([]float64, len(dates))
	for i := range values {
		values[i] = 1
	}
	out, err := NewSeries(rates.Symbol+"-CASH", dates, values)
	if err != nil {
		return nil, err
	}
	if out.Len() == 0 {
		return nil, fmt.Errorf("marketdata: cash index needs at least one date")
	}
	if out.First().Date.Before(rates.First().Date) {
		return nil, fmt.Errorf("marketdata: no cash rate at the index start")
	}
	for _, p := range rates.Points {
		if 1+p.Close/100/denom <= 0 {
			return nil, fmt.Errorf("marketdata: cash rate exhausts a day's capital")
		}
	}
	if j, ok := rates.junctionIn(out.First().Date, out.Last().Date); ok {
		return nil, fmt.Errorf("marketdata: rate definition changes on %s", j.Format(time.DateOnly))
	}
	out.Currency = rates.Currency
	out.Source = rates.Source
	out.Warnings = slices.Clone(rates.Warnings)
	ratePos := 0
	for i := 1; i < out.Len(); i++ {
		growth := 1.0
		for d := out.Points[i-1].Date; d.Before(out.Points[i].Date); d = d.AddDate(0, 0, 1) {
			for ratePos+1 < rates.Len() && !rates.Points[ratePos+1].Date.After(d) {
				ratePos++
			}
			growth *= 1 + rates.Points[ratePos].Close/100/denom
		}
		out.Points[i].Close = out.Points[i-1].Close * growth
		if math.IsInf(out.Points[i].Close, 0) || out.Points[i].Close <= 0 {
			return nil, fmt.Errorf("marketdata: cash accumulation overflow or underflow")
		}
	}
	if out.Last().Date.After(rates.Last().Date) {
		out.Warnings = append(out.Warnings, fmt.Sprintf("last %s rate held after %s", rates.Symbol, rates.Last().Date.Format(time.DateOnly)))
	}
	return out, nil
}
