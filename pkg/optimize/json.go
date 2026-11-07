package optimize

import "encoding/json"

// jsonNumber preserves undefined floating-point values as JSON null.
func jsonNumber(x float64) *float64 {
	if !finite(x) {
		return nil
	}
	return &x
}

func jsonNumbers(values []float64) []*float64 {
	if values == nil {
		return nil
	}
	out := make([]*float64, len(values))
	for i, v := range values {
		out[i] = jsonNumber(v)
	}
	return out
}

// MarshalJSON preserves open weight bounds (represented by NaN in Go) as
// null. Every nonfinite numeric value is represented as null.
// This representation is intended for reporting, not Spec round trips:
// retain the original Spec or parse its text again before solving.
func (s Spec) MarshalJSON() ([]byte, error) {
	type plain Spec
	bounds := make(map[string][2]*float64, len(s.Bounds))
	if s.Bounds == nil {
		bounds = nil
	}
	for key, b := range s.Bounds {
		bounds[key] = [2]*float64{jsonNumber(b[0]), jsonNumber(b[1])}
	}
	type viewJSON struct {
		Asset, Versus      string
		Return, Confidence *float64
	}
	var views []viewJSON
	if s.Views != nil {
		views = make([]viewJSON, len(s.Views))
		for i, v := range s.Views {
			views[i] = viewJSON{v.Asset, v.Versus, jsonNumber(v.Return), jsonNumber(v.Confidence)}
		}
	}
	return json.Marshal(struct {
		plain
		Lower, Upper                                []*float64
		Bounds                                      map[string][2]*float64
		MaxWeight, MinWeight, PriorReturn, RiskFree *float64
		Prior, RiskBudgets                          []*float64
		Views                                       []viewJSON
		Limits                                      struct{ MaxVolatility, MinReturn, MaxDrawdown *float64 }
	}{plain: plain(s), Lower: jsonNumbers(s.Lower), Upper: jsonNumbers(s.Upper), Bounds: bounds,
		MaxWeight: jsonNumber(s.MaxWeight), MinWeight: jsonNumber(s.MinWeight), PriorReturn: jsonNumber(s.PriorReturn), RiskFree: jsonNumber(s.RiskFree),
		Prior: jsonNumbers(s.Prior), RiskBudgets: jsonNumbers(s.RiskBudgets), Views: views,
		Limits: struct{ MaxVolatility, MinReturn, MaxDrawdown *float64 }{jsonNumber(s.Limits.MaxVolatility), jsonNumber(s.Limits.MinReturn), jsonNumber(s.Limits.MaxDrawdown)},
	})
}

// MarshalJSON encodes undefined statistics as null, never a misleading zero.
func (r Result) MarshalJSON() ([]byte, error) {
	type plain Result
	return json.Marshal(struct {
		plain
		Weights                                                     []*float64
		Return, Volatility, Sharpe                                  *float64
		Sortino, ReturnToMaxDD, Ulcer, Worst5y, CWARP, CAGR, Lambda *float64
		Implied, Posterior                                          []*float64
	}{plain: plain(r), Weights: jsonNumbers(r.Weights),
		Return: jsonNumber(r.Return), Volatility: jsonNumber(r.Volatility), Sharpe: jsonNumber(r.Sharpe),
		Sortino: jsonNumber(r.Sortino), ReturnToMaxDD: jsonNumber(r.ReturnToMaxDD), Ulcer: jsonNumber(r.Ulcer), Worst5y: jsonNumber(r.Worst5y), CWARP: jsonNumber(r.CWARP), CAGR: jsonNumber(r.CAGR), Lambda: jsonNumber(r.Lambda),
		Implied: jsonNumbers(r.Implied), Posterior: jsonNumbers(r.Posterior)})
}
