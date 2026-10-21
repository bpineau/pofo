package decumul

import "math"

// Gompertz is a Gompertz mortality law parameterised by its modal age at
// death (Mode) and dispersion (Dispersion, in years): the hazard at age x is
// exp((x−Mode)/Dispersion)/Dispersion. Two parameters fit developed-country
// adult mortality well and avoid bundling a full life table.
type Gompertz struct {
	Mode, Dispersion float64 // modal age at death and dispersion, in years
}

// FrenchMortality is the bundled default law: a Gompertz law fitted by least
// squares on the remaining life expectancy at every age from 50 to 90 of
// INSEE's triennial France period table for 2023-2025, both sexes together
// ("Ensemble"; provisional results published 2026-01-13 with the "Bilan
// démographique 2025", https://www.insee.fr/fr/statistiques/8721189, file
// 6_FR_Table_mortalite_triennale.xlsx). It reproduces the table within half a
// year over that range: 35.0 years left at 50 (INSEE 34.7), 25.8 at 60
// (25.9), 21.5 at 65 (21.7), 17.5 at 70 (17.8), 10.5 at 80 (10.6).
//
// It is a PERIOD law, unisex, and carries no future mortality improvement,
// so it still runs shorter than the generational tables an insurer prices
// with; FrenchWomen and FrenchMen are the same fit on each sex's column. The
// fit, its residuals and the calibration it replaced (Gompertz{88, 10}, which
// ran 1.5 years short at 50 to 70) are recorded in
// docs/specs/stochastic-lifetime-kernel-design.md.
var FrenchMortality = Gompertz{Mode: 89.8, Dispersion: 9.3}

// FrenchWomen is the same calibration as FrenchMortality on the women's
// column of the INSEE table: 23.3 years left at 65 (INSEE 23.55).
var FrenchWomen = Gompertz{Mode: 91.9, Dispersion: 8.3}

// FrenchMen is the same calibration as FrenchMortality on the men's column of
// the INSEE table: 19.7 years left at 65 (INSEE 19.87).
var FrenchMen = Gompertz{Mode: 87.5, Dispersion: 9.9}

// LifeExpectancy is the remaining life expectancy at age, in years: the
// integral of Survival over the years ahead, the "complete expectation" a
// national life table publishes (INSEE's "espérance de vie à l'âge x"). The
// kernel draws whole years and pays at the start of each, so its expected
// count of yearly payments runs about half a year above this figure.
func (g Gompertz) LifeExpectancy(age float64) float64 {
	// Composite Simpson at a fortnight's step, until survival is negligible.
	const h, limit = 1.0 / 24, 150.0
	sum := 0.0
	for t := 0.0; t < limit; t += 2 * h {
		a, m, b := g.Survival(age, t), g.Survival(age, t+h), g.Survival(age, t+2*h)
		sum += h / 3 * (a + 4*m + b)
		if b < 1e-12 {
			break
		}
	}
	return sum
}

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
	Dead, Broke, Funded float64 // shares of the households, summing to 1
}

// LifeCurve combines the ensemble's ruin timing with a survival curve into
// the alive-broke-dead decomposition per year-end ("Rich, Broke or Dead"):
// point t uses surv(t) and the share of paths ruined in year t or earlier.
// Mortality and market outcomes are independent, so the alive share is simply
// split by the cumulative ruin probability.
//
// It applies mortality AFTER the fact, which is exact whenever mortality does
// not change the path itself, and it is the right call on a fixed-horizon
// ensemble. On an ensemble whose plan already carries a Lifetime the deaths
// have been drawn inside the paths, so weighting them again would count
// mortality twice: use Ensemble.LifeStates there.
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
