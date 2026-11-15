package simgen

import (
	"encoding/json"
	"math"
)

func jsonNumber(v float64) *float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return nil
	}
	return &v
}

// MarshalJSON preserves undefined validation ratios as null.
func (v Validation) MarshalJSON() ([]byte, error) {
	type plain Validation
	return json.Marshal(struct {
		plain
		Corr, WeeklyCorr, Beta *float64
	}{
		plain(v), jsonNumber(v.Corr), jsonNumber(v.WeeklyCorr), jsonNumber(v.Beta),
	})
}

// MarshalJSON preserves undefined audit correlations and beta as null.
func (a AuditResult) MarshalJSON() ([]byte, error) {
	type plain AuditResult
	return json.Marshal(struct {
		plain
		DailyCorr, WeeklyCorr, MonthlyCorr, Beta *float64
	}{
		plain(a), jsonNumber(a.DailyCorr), jsonNumber(a.WeeklyCorr), jsonNumber(a.MonthlyCorr), jsonNumber(a.Beta),
	})
}

// MarshalJSON preserves an undefined junction correlation as null.
func (j Junction) MarshalJSON() ([]byte, error) {
	type plain Junction
	return json.Marshal(struct {
		plain
		Corr *float64
	}{plain(j), jsonNumber(j.Corr)})
}
