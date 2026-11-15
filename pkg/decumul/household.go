package decumul

import (
	"fmt"
	"math"
)

// HouseholdSpendBands returns per-year quantiles of household real spending,
// adding this plan's deterministic cashflows to Ensemble.SpendBands. Incomes
// such as pensions are included, dated expenses subtracted. It reuses the
// ensemble without drawing or simulating again. Annuities and mortality are
// unsupported: a non-nil Lifetime or a mismatched horizon/path length is an
// error, since mortality-dependent income cannot be recovered from the ensemble.
// Each band has Years entries (year zero through Years-1), none at the terminal
// wealth point. Empty ensembles return empty bands, as SpendBands does.
func (p Plan) HouseholdSpendBands(e Ensemble, pcts []float64) ([][]float64, error) {
	income, err := p.householdIncome(e)
	if err != nil {
		return nil, err
	}
	bands := e.SpendBands(pcts)
	for _, band := range bands {
		for y := range band {
			band[y] += income[y]
		}
	}
	return bands, nil
}

// HouseholdSpendCV returns the population coefficient of variation of real
// household spending over every fixed-horizon path-year, including deterministic
// pension/income as in HouseholdSpendBands. Empty samples and zero mean return
// NaN. The same horizon/mortality restrictions as HouseholdSpendBands apply.
func (p Plan) HouseholdSpendCV(e Ensemble) (float64, error) {
	income, err := p.householdIncome(e)
	if err != nil {
		return math.NaN(), err
	}
	n := len(e.Paths) * p.Years
	if n == 0 {
		return math.NaN(), nil
	}
	sum := 0.
	for _, path := range e.Paths {
		for y, s := range path.Spend {
			sum += s + income[y]
		}
	}
	mean := sum / float64(n)
	if mean == 0 {
		return math.NaN(), nil
	}
	variance := 0.
	for _, path := range e.Paths {
		for y, s := range path.Spend {
			delta := s + income[y] - mean
			variance += delta * delta
		}
	}
	return math.Sqrt(variance/float64(n)) / mean, nil
}

func (p Plan) householdIncome(e Ensemble) ([]float64, error) {
	if p.Lifetime != nil || p.Years < 0 || e.Years != p.Years {
		return nil, fmt.Errorf("decumul: household spending requires a matching fixed-horizon plan")
	}
	for _, path := range e.Paths {
		if len(path.Spend) != p.Years || path.spendYears() != p.Years {
			return nil, fmt.Errorf("decumul: household spending requires complete fixed-horizon paths")
		}
	}
	income := make([]float64, p.Years)
	for y := range income {
		for _, flow := range p.Cashflows {
			income[y] += flow.AmountAt(y)
		}
	}
	return income, nil
}
