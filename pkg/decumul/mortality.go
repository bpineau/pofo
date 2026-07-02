package decumul

import "math"

// Gompertz is a Gompertz mortality law parameterised by its modal age at
// death (Mode) and dispersion (Dispersion, in years): the hazard at age x is
// exp((x−Mode)/Dispersion)/Dispersion. Two parameters fit developed-country
// adult mortality well and avoid bundling a full life table.
type Gompertz struct {
	Mode, Dispersion float64
}

// FrenchMortality approximates the INSEE 2020s unisex period table for adults:
// modal age at death ≈ 88 with a ≈ 10-year dispersion (women ≈ 90/9, men ≈
// 85/10.5). Remaining life expectancy at 50 comes out in the mid-30s of years,
// matching the published tables closely enough for retirement planning.
var FrenchMortality = Gompertz{Mode: 88, Dispersion: 10}

// Survival is the probability that a person of the given age is still alive
// after years more years.
func (g Gompertz) Survival(age, years float64) float64 {
	if years <= 0 {
		return 1
	}
	// Integrated hazard from age to age+years for the Gompertz law.
	h := math.Exp((age-g.Mode)/g.Dispersion) * (math.Exp(years/g.Dispersion) - 1)
	return math.Exp(-h)
}

// CoupleSurvival is the probability that at least one member of a same-age
// couple is still alive after years more years (deaths assumed independent).
// For a household plan, ruin only matters while someone is alive to be ruined.
func (g Gompertz) CoupleSurvival(age, years float64) float64 {
	dead := 1 - g.Survival(age, years)
	return 1 - dead*dead
}

// LifePoint splits the households at one year-end into three exclusive
// states: Dead (nobody left alive), Broke (alive with the capital exhausted)
// and Funded (alive with capital remaining). The three sum to 1.
type LifePoint struct {
	Dead, Broke, Funded float64
}

// LifeCurve combines the ensemble's ruin timing with a survival curve into
// the alive-broke-dead decomposition per year-end ("Rich, Broke or Dead"):
// point t uses surv(t) and the share of paths ruined in year t or earlier.
// Mortality and market outcomes are independent, so the alive share is simply
// split by the cumulative ruin probability.
func (e Ensemble) LifeCurve(surv func(years float64) float64) []LifePoint {
	ruinedBy := e.cumulativeRuin()
	out := make([]LifePoint, e.Years+1)
	for t := range out {
		alive := surv(float64(t))
		broke := alive * ruinedBy[t]
		out[t] = LifePoint{Dead: 1 - alive, Broke: broke, Funded: alive - broke}
	}
	return out
}

// cumulativeRuin returns, for each year-end t (0..Years), the share of paths
// whose ruin happened in year t or earlier.
func (e Ensemble) cumulativeRuin() []float64 {
	out := make([]float64, e.Years+1)
	if len(e.Paths) == 0 {
		return out
	}
	for _, p := range e.Paths {
		if p.Ruined && p.RuinYear >= 0 {
			for t := min(p.RuinYear, e.Years); t <= e.Years; t++ {
				out[t]++
			}
		}
	}
	for t := range out {
		out[t] /= float64(len(e.Paths))
	}
	return out
}

// RuinYearHistogram returns the share of all paths ruined in each year
// (indices 0..Years-1): the "when do failures happen" distribution, which
// shows whether the plan dies of early sequence risk or of late longevity.
func (e Ensemble) RuinYearHistogram() []float64 {
	out := make([]float64, e.Years)
	if len(e.Paths) == 0 {
		return out
	}
	for _, p := range e.Paths {
		if p.Ruined && p.RuinYear >= 0 && p.RuinYear < e.Years {
			out[p.RuinYear]++
		}
	}
	for i := range out {
		out[i] /= float64(len(e.Paths))
	}
	return out
}
