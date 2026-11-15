package analyze

import (
	"encoding/json"
	"math"
	"time"

	"github.com/bpineau/pofo/pkg/metrics"
)

func jsonNumber(v float64) *float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return nil
	}
	return &v
}

// MarshalJSON preserves undefined correlations as null, in asset order.
func (p PortfolioStudy) MarshalJSON() ([]byte, error) {
	type plain PortfolioStudy
	var correlation [][]*float64
	if p.Correlation != nil {
		correlation = make([][]*float64, len(p.Correlation))
		for i, row := range p.Correlation {
			correlation[i] = make([]*float64, len(row))
			for j, v := range row {
				correlation[i][j] = jsonNumber(v)
			}
		}
	}
	return json.Marshal(struct {
		plain
		Correlation [][]*float64
	}{plain(p), correlation})
}

// MarshalJSON keeps the flattened tracking fields, dates and divergences.
// An embedded Tracking's marshaler alone would discard the latter fields.
func (p PairReturns) MarshalJSON() ([]byte, error) {
	type tracking metrics.Tracking
	return json.Marshal(struct {
		tracking
		Start, End                                                         time.Time
		Divergences                                                        []Divergence
		Corr, VolA, VolB, VolRatio, TrackingError, Difference, Beta, Alpha *float64
	}{tracking(p.Tracking), p.Start, p.End, p.Divergences,
		jsonNumber(p.Corr), jsonNumber(p.VolA), jsonNumber(p.VolB), jsonNumber(p.VolRatio),
		jsonNumber(p.TrackingError), jsonNumber(p.Difference), jsonNumber(p.Beta), jsonNumber(p.Alpha)})
}
