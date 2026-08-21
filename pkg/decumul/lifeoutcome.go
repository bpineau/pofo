package decumul

import "github.com/bpineau/pofo/pkg/metrics"

// LifeOutcome bundles the statistics only a drawn lifetime can produce: what
// share of households actually experienced running out of money, how long they
// lived broke, what they left behind, and what they lived on. Wealth figures
// are real euros, shares are fractions.
//
// It reads correctly on a fixed-horizon Ensemble too, as the degenerate case
// where every household reaches the horizon: RuinAlive is then the plain ruin
// probability, OutlivedPlan is 1 and the estate is the terminal wealth.
type LifeOutcome struct {
	// RuinAlive is the share of households that ran out of money WHILE ALIVE.
	// Under a Lifetime the kernel stops at death, so a withdrawal can only
	// fail while someone is there to feel it: this is counted, not weighted
	// after the fact. It is the figure Ensemble.LifeCurve approximates.
	RuinAlive float64
	// OutlivedPlan is the share still alive at the plan horizon, i.e. censored
	// rather than dead. A large value means Plan.Years is truncating the
	// longevity tail and the estate distribution is not what it seems.
	OutlivedPlan float64
	// MedianLifeYears is the median number of years lived, a sanity read on
	// the mortality law actually in force.
	MedianLifeYears float64
	// BrokeYearsMean and BrokeYearsP95 are the years lived AFTER running out,
	// over all households (0 for those that never did). The mean is the
	// expected length of the failure; the p95 is what the unlucky tail faces,
	// and it is the honest measure of how bad "ruin" is in this plan, since a
	// ruin two years before death and one at 70 are not the same event.
	BrokeYearsMean float64
	BrokeYearsP95  float64
	// The estate left at the household's end: the p5/p50/p90 of real wealth at
	// death (or at the horizon for the censored), and the share leaving
	// nothing at all.
	EstateP5, EstateP50, EstateP90 float64
	EstateZero                     float64
	// IncomeMean is the mean TOTAL real income per LIVED year: what the
	// portfolio delivered plus the pensions and annuity received. It is the
	// lifestyle side of the trade, against which the ruin and estate figures
	// are read, and it must be the total: annuitising moves income from the
	// portfolio to the insurer, so a portfolio-only reading would show a
	// collapse where the household felt none. Ensemble.SpendCV and SpendBands
	// carry the swing of the portfolio-funded part, which is where all the
	// variability is (a pension and a real annuity do not move).
	IncomeMean float64
}

// LifeOutcome computes the bundle.
func (e Ensemble) LifeOutcome() LifeOutcome {
	var o LifeOutcome
	n := len(e.Paths)
	if n == 0 {
		return o
	}
	estates := make([]float64, n)
	broke := make([]float64, n)
	lifeYears := make([]float64, n)
	var ruined, outlived, zero int
	var income float64
	var years int
	for i, p := range e.Paths {
		lived := p.end()
		estates[i] = p.Wealth[lived]
		lifeYears[i] = float64(lived)
		income += p.Received
		for _, s := range p.Spend[:p.spendYears()] {
			income += s
		}
		years += p.spendYears()
		if estates[i] <= 0 {
			zero++
		}
		if p.Outlived {
			outlived++
		}
		if p.Ruined && p.RuinYear >= 0 {
			ruined++
			if y := lived - p.RuinYear; y > 0 {
				broke[i] = float64(y)
			}
		}
	}
	fn := float64(n)
	o.RuinAlive = float64(ruined) / fn
	o.OutlivedPlan = float64(outlived) / fn
	o.EstateZero = float64(zero) / fn
	o.MedianLifeYears = metrics.Quantiles(lifeYears, 0.50)[0]
	q := metrics.Quantiles(estates, 0.05, 0.50, 0.90)
	o.EstateP5, o.EstateP50, o.EstateP90 = q[0], q[1], q[2]
	var sum float64
	for _, b := range broke {
		sum += b
	}
	o.BrokeYearsMean = sum / fn
	o.BrokeYearsP95 = metrics.Quantiles(broke, 0.95)[0]
	if years > 0 {
		o.IncomeMean = income / float64(years)
	}
	return o
}

// Estates returns each path's real wealth at the household's end, in path
// order: the raw material behind LifeOutcome's quantiles, for a histogram or a
// caller's own tail statistic.
func (e Ensemble) Estates() []float64 {
	out := make([]float64, len(e.Paths))
	for i, p := range e.Paths {
		out[i] = p.Wealth[p.end()]
	}
	return out
}

// LifeStates is the exact counterpart of LifeCurve: the same alive-broke-dead
// decomposition per year-end, counted from the lifespans each path was dealt
// instead of weighted afterwards by a survival curve.
//
// The two agree, to Monte-Carlo error, whenever mortality does not change the
// path itself (no annuity, no survivor adjustment, household-owned income),
// because a path's trajectory is then independent of the draw of death.
// They part company exactly where the weighting stops being able to model the
// question: an annuity whose income dies with its annuitant, a survivor
// spending less, a pension that only partly reverts.
//
// On a fixed-horizon Ensemble nobody ever dies, so Dead is 0 throughout and
// the curve degenerates to the plain cumulative-ruin split.
func (e Ensemble) LifeStates() []LifePoint {
	out := make([]LifePoint, e.Years+1)
	n := len(e.Paths)
	if n == 0 {
		return out
	}
	dead := make([]float64, e.Years+1)
	broke := make([]float64, e.Years+1)
	for _, p := range e.Paths {
		// A household with LifeYears years lives years 0..LifeYears-1, so at
		// year-end t it is gone exactly when t >= LifeYears.
		aliveTo := e.Years // last year-end it is still there
		if !p.Outlived && p.LifeYears > 0 && p.LifeYears <= e.Years {
			aliveTo = p.LifeYears - 1
			for t := p.LifeYears; t <= e.Years; t++ {
				dead[t]++
			}
		}
		if p.Ruined && p.RuinYear >= 0 {
			for t := max(p.RuinYear, 0); t <= aliveTo; t++ {
				broke[t]++
			}
		}
	}
	fn := float64(n)
	for t := range out {
		d, b := dead[t]/fn, broke[t]/fn
		out[t] = LifePoint{Dead: d, Broke: b, Funded: 1 - d - b}
	}
	return out
}
