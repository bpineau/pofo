package metrics

import (
	"encoding/json"
	"math"
	"strconv"
)

// jsonNumber keeps missing statistical estimates distinct from measured zero.
type jsonNumber float64

func (v jsonNumber) MarshalJSON() ([]byte, error) {
	if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
		return []byte("null"), nil
	}
	return []byte(strconv.FormatFloat(float64(v), 'g', -1, 64)), nil
}

// MarshalJSON writes undefined or non-finite estimates as null. The Go API
// retains IEEE NaN for undefined ratios; zero remains a measured zero.
func (s Stats) MarshalJSON() ([]byte, error) {
	type plain Stats
	return json.Marshal(struct {
		plain
		Years, PeriodsPerYear, CAGR, Volatility, Sharpe, Sortino, Ulcer, MaxDrawdown, Beta, CWARP, Skew, Kurtosis jsonNumber
	}{plain(s), jsonNumber(s.Years), jsonNumber(s.PeriodsPerYear), jsonNumber(s.CAGR), jsonNumber(s.Volatility), jsonNumber(s.Sharpe), jsonNumber(s.Sortino), jsonNumber(s.Ulcer), jsonNumber(s.MaxDrawdown), jsonNumber(s.Beta), jsonNumber(s.CWARP), jsonNumber(s.Skew), jsonNumber(s.Kurtosis)})
}

// MarshalJSON writes an undefined relative statistic as null, never zero.
func (r Relative) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct{ Beta, Alpha, InfoRatio, UpCapture, DownCapture jsonNumber }{jsonNumber(r.Beta), jsonNumber(r.Alpha), jsonNumber(r.InfoRatio), jsonNumber(r.UpCapture), jsonNumber(r.DownCapture)})
}

// MarshalJSON represents undefined or infinite coefficient statistics as null.
func (e Estimate) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct{ Value, SE, T jsonNumber }{jsonNumber(e.Value), jsonNumber(e.SE), jsonNumber(e.T)})
}

// MarshalJSON preserves undefined fit quality as null, including the R2 of
// a constant response. Coefficient estimates use the same missing-value rule.
func (r Regression) MarshalJSON() ([]byte, error) {
	type plain Regression
	return json.Marshal(struct {
		plain
		R2, ResidualSD jsonNumber
	}{plain(r), jsonNumber(r.R2), jsonNumber(r.ResidualSD)})
}
