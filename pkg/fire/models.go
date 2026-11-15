package fire

import (
	"fmt"
	"math"

	"github.com/bpineau/pofo/pkg/scenario"
)

const (
	// Central is the calibrated Student-t annual return model.
	Central = "central"
	// Bootstrap resamples the supplied portfolio's monthly real history.
	Bootstrap = "bootstrap"
	// Historical samples contiguous portfolio cohorts uniformly with replacement.
	Historical = "historical"
	// Stress preserves the central mean but clusters volatile bear sequences.
	Stress = "stress"
	// BroadSample resamples the external domestic 60/40 multi-country record.
	BroadSample = "broad_sample"
	// LostDecade lowers realized returns through prolonged bear regimes.
	LostDecade = "lost_decade"
	// MinPanelMonths is the minimum history accepted for portfolio-backed models.
	MinPanelMonths = 24
)

// MarketSpec supplies raw annual arithmetic mean, annual volatility and
// Student-t degrees of freedom. Mu and Sigma are fractions. Panel contains
// monthly real holding returns; nil selects parametric mode. Weights are
// fractions per panel row; nil uses Panel.Weights. CAPE replaces the central
// mean for Central, Stress and LostDecade, bypassing short-history blending.
type MarketSpec struct {
	Mu, Sigma, Nu float64         // annual real arithmetic mean/volatility fractions, Student-t degrees of freedom
	Panel         *scenario.Panel // monthly real history; nil for parametric assumptions
	Weights       []float64       // normalized fractions per holding; nil uses Panel.Weights
	CAPE          bool            // anchor synthetic means on the bundled CAPE yield
}

// PreparedModel is one annual source with its meaning and effective inputs.
// Source draws annual real returns, including compounded monthly history.
// HistoricalCohorts counts available start windows, not independent evidence.
type PreparedModel struct {
	ID, Name, Description, Basis, Caveat     string          // model identity, interpretation, input basis and short-history caution
	Source                                   scenario.Source // annual real-return source
	RawMu, RawSigma, RawNu                   float64         // supplied synthetic assumptions; annual fractions and degrees of freedom
	EffectiveMu, EffectiveSigma, EffectiveNu float64         // calibrated synthetic assumptions; unused for empirical models
	CAPEApplied                              bool            // whether the synthetic mean uses CAPE
	CAPE                                     CAPESnapshot    // valuation used when CAPEApplied
	HistoricalCohorts                        int             // available contiguous windows for Historical
	PeriodsPerYear                           int             // one annual withdrawal step for every prepared model
}

// CentralParameters applies the shared short-history blend or the requested
// CAPE anchor. A missing panel leaves raw values unchanged. It returns an
// error for an invalid requested CAPE observation; stale observations remain
// usable and must be reported with CurrentCAPE's Stale indication.
func CentralParameters(m MarketSpec, years int) (mu, sigma, nu float64, err error) {
	mu, sigma, nu = m.Mu, m.Sigma, m.Nu
	if m.CAPE {
		mu, err = CurrentCAPE().ArithmeticMean(sigma)
		return
	}
	if m.Panel == nil || years <= 0 {
		return
	}
	histYears := float64(m.Panel.Periods() / 12)
	s := (float64(years) - histYears) / float64(years)
	s = math.Max(0, math.Min(s, .5))
	blend := func(fit, prior float64) float64 { return (1-s)*fit + s*prior }
	return blend(mu, .045), blend(sigma, .13), blend(nu, 4), nil
}

// PrepareModel resolves one of the six model IDs without fallback. It rejects
// unknown IDs, nonpositive horizons, unavailable historical cohorts, a panel
// shorter than MinPanelMonths, invalid weights and invalid synthetic inputs.
// Historical Draw samples cohorts with replacement; it does not enumerate all
// available windows. BroadSample ignores portfolio and parametric inputs.
func PrepareModel(model string, m MarketSpec, years int) (PreparedModel, error) {
	r := PreparedModel{ID: model, RawMu: m.Mu, RawSigma: m.Sigma, RawNu: m.Nu, PeriodsPerYear: 1}
	if years <= 0 {
		return r, fmt.Errorf("fire: horizon must be positive")
	}
	if m.Panel != nil {
		if err := validatePanel(*m.Panel, m.Weights); err != nil {
			return r, err
		}
		if m.Panel.Periods() < MinPanelMonths {
			return r, fmt.Errorf("fire: portfolio requires at least %d consecutive real-return months", MinPanelMonths)
		}
		if m.Panel.Periods() < years*12 {
			r.Caveat = fmt.Sprintf("History %d months is shorter than the %d-year horizon; resampling this limited window cannot establish full-horizon reliability.", m.Panel.Periods(), years)
		}
	}
	switch model {
	case Bootstrap, Historical:
		if m.Panel == nil {
			return r, fmt.Errorf("fire: %s requires portfolio history", model)
		}
		r.Basis = "portfolio_history"
		if model == Bootstrap {
			r.Name = "Block bootstrap"
			r.Description = "Resamples multi-year blocks of this portfolio's own monthly real returns at the supplied weights, preserving clustered bear markets and cross-asset correlation. Data-driven: parametric mu/sigma/df inputs have no effect here. Manufactures many full-length retirements from a short history, but stays anchored to that one favourable window."
			r.Source = scenario.Compounded{Inner: scenario.StationaryBootstrap{Panel: *m.Panel, Weights: m.Weights, MeanBlock: 24, Periods: years * 12}, Group: 12}
		} else {
			c := scenario.HistoricalCohorts{Panel: *m.Panel, Weights: m.Weights, Periods: years * 12}
			if c.Count() == 0 {
				return r, fmt.Errorf("fire: historical model requires a complete horizon")
			}
			r.Name = "Historical windows"
			r.Description = "Samples this portfolio's actual contiguous historical return windows uniformly with replacement. Honest but limited: a short history holds few independent retirements, so it tends to look optimistic for long horizons."
			r.HistoricalCohorts = c.Count()
			r.Source = scenario.Compounded{Inner: c, Group: 12}
		}
	case BroadSample:
		r.Name = "Broad-sample"
		r.Basis = "external_reference"
		r.Description = "The empirical century: a 60/40 domestic stock/bond portfolio (the broad-sample SWR literature's baseline, Anarkulova et al.) block-bootstrapped from the actual 1870-2020 record of 16 developed markets (Jorda-Schularick-Taylor), resampling runs of single countries. Fully data-driven and fully external: it ignores both the supplied portfolio and the parametric inputs, and it carries the real disasters (1929-32, the inflationary 1970s, Japan post-1990) that cause ruin, on bonds as well as stocks."
		r.Source = BroadSampleSource(years)
	case Central, Stress, LostDecade:
		if !finite(m.Mu) || !finite(m.Sigma) || m.Sigma < 0 || !finite(m.Nu) || m.Nu < 0 {
			return r, fmt.Errorf("fire: invalid parametric assumptions")
		}
		var err error
		r.EffectiveMu, r.EffectiveSigma, r.EffectiveNu, err = CentralParameters(m, years)
		if err != nil {
			return r, err
		}
		r.Basis = "parametric_assumptions"
		if m.Panel != nil {
			r.Basis = "portfolio_calibration"
		}
		r.CAPEApplied = m.CAPE
		if m.CAPE {
			r.CAPE = CurrentCAPE()
		}
		mu, sigma, nu := r.EffectiveMu, r.EffectiveSigma, r.EffectiveNu
		switch model {
		case Central:
			r.Name = "Student-t"
			r.Description = "The central case to plan on: synthetic i.i.d. annual real returns generated from parametric inputs (mu, sigma, tail df), which in portfolio mode are seeded from the holdings' fitted statistics, never their actual sequence of years. When the history is shorter than the horizon the values are blended toward the broad-sample prior (a short window cannot show long-horizon tail and sequence risk)."
			r.Source = scenario.ParametricSource{Mu: mu, Sigma: sigma, Df: nu, Periods: years}
		case Stress:
			r.Name = "Sequence stress"
			r.Description = "Sequence-of-returns stress: clustered, persistent bull/bear regimes at the SAME long-run mean as Student-t, so a run of bad years can land early in retirement. The expected return is unchanged, but bear years are also 1.5× more volatile, so this column stresses mostly the ordering, not the ordering alone. Read it as the downside if the sequence is unlucky."
			r.Source = scenario.NewMarkovRegime(mu, sigma, nu, years)
		case LostDecade:
			r.Name = "Lost decade"
			r.Description = "Japan-style tail: a very sticky, deep bear averaging a whole decade, layered on your mean. Unlike Sequence stress this lowers the realised return too, modelling a prolonged real drawdown (Japanese equities 1990-2010). The grimmest planning model, the scenario where a retirement begins inside a lost decade."
			r.Source = scenario.NewLostDecadeRegime(mu, sigma, nu, years)
		}
	default:
		return r, fmt.Errorf("fire: unknown model %q", model)
	}
	return r, nil
}

func validatePanel(p scenario.Panel, weights []float64) error {
	n := p.Periods()
	if n == 0 {
		return fmt.Errorf("fire: empty monthly panel")
	}
	for _, row := range p.Returns {
		if len(row) != n {
			return fmt.Errorf("fire: inconsistent monthly panel")
		}
		for _, r := range row {
			if !finite(r) || r < -1 {
				return fmt.Errorf("fire: invalid real return")
			}
		}
	}
	if weights == nil {
		weights = p.Weights
	}
	if len(weights) != len(p.Returns) {
		return fmt.Errorf("fire: weights must match panel rows")
	}
	sum := 0.
	for _, w := range weights {
		if !finite(w) || w < 0 {
			return fmt.Errorf("fire: invalid weight")
		}
		sum += w
	}
	if !finite(sum) || math.Abs(sum-1) > 1e-8 {
		return fmt.Errorf("fire: weights must sum to one")
	}
	return nil
}
