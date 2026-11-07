package analyze

import (
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
	"slices"
)

// AdjustedStudy describes a transformed curve on the nominal study's calendar.
// Stats use a zero risk-free rate: Real measures purchasing power and Excess
// already subtracts the chosen cash return. Units are those of metrics.Stats.
type AdjustedStudy struct {
	Series    *marketdata.Series     // transformed curve, including its warnings
	Stats     metrics.Stats          // measured on Series
	Years     []metrics.PeriodReturn // calendar years
	Months    []metrics.PeriodReturn // calendar months
	Drawdowns []metrics.Episode      // chronological drawdown episodes
}

func adjusted(s *marketdata.Series) (*AdjustedStudy, error) {
	st, err := s.Stats()
	if err != nil {
		return nil, err
	}
	a := &AdjustedStudy{Series: s, Stats: st}
	a.Years, a.Months, a.Drawdowns = calendar(s.Dates(), s.Values())
	return a, nil
}

func adjustStudy(a *AssetStudy, opt Options) error {
	if opt.RiskFree != 0 {
		st, err := metrics.ComputeWith(a.Series.Dates(), a.Series.Values(), metrics.ComputeOptions{PeriodsPerYear: a.Series.Cadence, RiskFree: opt.RiskFree})
		if err != nil {
			return err
		}
		a.Stats.Sharpe, a.Stats.Sortino = st.Sharpe, st.Sortino
		if a.Relative != nil {
			a.Relative.Alpha += (a.Relative.Beta - 1) * opt.RiskFree
		}
	}
	if opt.Inflation != nil {
		s, err := a.Series.Deflate(opt.Inflation, opt.Adjustment)
		if err != nil {
			return err
		}
		a.Real, err = adjusted(s)
		if err != nil {
			return err
		}
		a.Warnings = append(a.Warnings, s.Warnings...)
	}
	if opt.RiskFreeIndex != nil {
		s, err := a.Series.Excess(opt.RiskFreeIndex, opt.Adjustment)
		if err != nil {
			return err
		}
		a.Excess, err = adjusted(s)
		if err != nil {
			return err
		}
		a.Stats.Sharpe, a.Stats.Sortino = a.Excess.Stats.Sharpe, a.Excess.Stats.Sortino
		a.Warnings = append(a.Warnings, s.Warnings...)
	}
	unique := make([]string, 0, len(a.Warnings))
	for _, w := range a.Warnings {
		if !slices.Contains(unique, w) {
			unique = append(unique, w)
		}
	}
	a.Warnings = unique
	return nil
}
