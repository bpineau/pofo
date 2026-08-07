package permanent

import "time"

// SignalConfig tunes how the world regime is read from the panel. The zero
// value is not useful; start from DefaultSignalConfig.
type SignalConfig struct {
	// BreadthCountries are the economies whose growth/inflation acceleration is
	// counted for breadth. Empty means every country in the panel.
	BreadthCountries []string
	// RateCountries are the economies averaged for the monetary slope and real
	// short rate (a smaller, reliable set with both long and short rates).
	RateCountries []string
	// AccelMonths is the lookback for "accelerating": the year-on-year rate now
	// exceeding the year-on-year rate this many months ago.
	AccelMonths int
	// SmoothMonths averages the raw world point over a trailing window, so the
	// signal moves in slow waves rather than month to month.
	SmoothMonths int
	// MinBreadth and MinRate are the minimum reporting countries required for a
	// month to yield a regime.
	MinBreadth, MinRate int
}

// DefaultSignalConfig is the reconstruction used in the design doc: a broad
// OECD breadth set, a G8 rate set, 3-month acceleration and 3-month smoothing.
// See docs/darcet-permanent-portfolio-design.md; these are not fitted optima.
func DefaultSignalConfig() SignalConfig {
	return SignalConfig{
		RateCountries: []string{"USA", "JPN", "DEU", "FRA", "GBR", "ITA", "CAN", "AUS"},
		AccelMonths:   3,
		SmoothMonths:  3,
		MinBreadth:    8,
		MinRate:       3,
	}
}

// Regime is the world macro state at a month, in Darcet's two quadrants.
type Regime struct {
	Date time.Time
	// GrowthBreadth and InflationBreadth are the share of countries whose
	// industrial-production / CPI year-on-year is accelerating, in [0,1].
	// Paradise is (1,0); stagflation "hell" is (0,1).
	GrowthBreadth, InflationBreadth float64
	// Slope is the mean long-minus-short rate; RealShort the mean short rate
	// minus year-on-year inflation. Both in percentage points.
	Slope, RealShort float64
}

// Quadrant is the coarse growth x inflation reading of a Regime: the four
// classic macro seasons behind All-Weather-style thinking.
type Quadrant int

const (
	GrowthQuadrant    Quadrant = iota // growth accelerating, inflation not
	InflationQuadrant                 // both accelerating (reflation / overheating)
	DeflationQuadrant                 // both decelerating (recession / disinflation)
	CrisisQuadrant                    // inflation without growth (stagflation)
)

// String returns the quadrant's lower-case label.
func (q Quadrant) String() string {
	switch q {
	case InflationQuadrant:
		return "inflation"
	case DeflationQuadrant:
		return "deflation"
	case CrisisQuadrant:
		return "crisis"
	default:
		return "growth"
	}
}

// Quadrant reduces the regime's continuous breadths to their quadrant, both
// thresholded at one half. This coarse view exists for labeling and
// reporting (regime timelines, per-quadrant aggregations); the allocator
// itself works on the continuous breadths (Allocate) and never uses this
// discretization.
func (r Regime) Quadrant() Quadrant {
	switch {
	case r.GrowthBreadth >= 0.5 && r.InflationBreadth >= 0.5:
		return InflationQuadrant
	case r.GrowthBreadth < 0.5 && r.InflationBreadth >= 0.5:
		return CrisisQuadrant
	case r.GrowthBreadth < 0.5 && r.InflationBreadth < 0.5:
		return DeflationQuadrant
	default:
		return GrowthQuadrant
	}
}

// accelerating reports whether column col for iso is accelerating at m: its
// year-on-year rate now above its year-on-year rate AccelMonths ago.
func (p *Panel) accelerating(col, iso string, m time.Time, cfg SignalConfig) (bool, bool) {
	now, ok1 := p.yoy(col, iso, m)
	then, ok2 := p.yoy(col, iso, m.AddDate(0, -cfg.AccelMonths, 0))
	if !ok1 || !ok2 {
		return false, false
	}
	return now > then, true
}

// rawPoint computes the un-smoothed world point at m, or ok=false if too few
// countries report.
func (p *Panel) rawPoint(m time.Time, cfg SignalConfig) (Regime, bool) {
	breadth := cfg.BreadthCountries
	if len(breadth) == 0 {
		breadth = p.isos
	}
	var gAcc, gTot, iAcc, iTot int
	for _, iso := range breadth {
		if a, ok := p.accelerating("ip", iso, m, cfg); ok {
			gTot++
			if a {
				gAcc++
			}
		}
		if a, ok := p.accelerating("cpi", iso, m, cfg); ok {
			iTot++
			if a {
				iAcc++
			}
		}
	}
	if gTot < cfg.MinBreadth || iTot < cfg.MinBreadth {
		return Regime{}, false
	}
	var slope, realShort float64
	var rc int
	for _, iso := range cfg.RateCountries {
		long, okl := p.value("longrate", iso, m)
		short, oks := p.value("shortrate", iso, m)
		infl, oki := p.yoy("cpi", iso, m)
		if okl && oks && oki {
			slope += long - short
			realShort += short - infl*100
			rc++
		}
	}
	if rc < cfg.MinRate {
		return Regime{}, false
	}
	return Regime{
		Date:             m,
		GrowthBreadth:    float64(gAcc) / float64(gTot),
		InflationBreadth: float64(iAcc) / float64(iTot),
		Slope:            slope / float64(rc),
		RealShort:        realShort / float64(rc),
	}, true
}

// RegimeAt returns the smoothed world regime at month m (m normalized to the
// first of its month), or ok=false if the panel does not cover it.
func (p *Panel) RegimeAt(m time.Time, cfg SignalConfig) (Regime, bool) {
	m = time.Date(m.Year(), m.Month(), 1, 0, 0, 0, 0, time.UTC)
	n := max(cfg.SmoothMonths, 1)
	var acc Regime
	var cnt int
	for k := range n {
		if r, ok := p.rawPoint(m.AddDate(0, -k, 0), cfg); ok {
			acc.GrowthBreadth += r.GrowthBreadth
			acc.InflationBreadth += r.InflationBreadth
			acc.Slope += r.Slope
			acc.RealShort += r.RealShort
			cnt++
		}
	}
	if cnt == 0 {
		return Regime{}, false
	}
	f := float64(cnt)
	return Regime{
		Date:             m,
		GrowthBreadth:    acc.GrowthBreadth / f,
		InflationBreadth: acc.InflationBreadth / f,
		Slope:            acc.Slope / f,
		RealShort:        acc.RealShort / f,
	}, true
}

// Regimes returns the smoothed regime for every month in [from, to] the panel
// covers, ascending. from and to are inclusive and normalized to month starts.
func (p *Panel) Regimes(from, to time.Time, cfg SignalConfig) []Regime {
	from = time.Date(from.Year(), from.Month(), 1, 0, 0, 0, 0, time.UTC)
	to = time.Date(to.Year(), to.Month(), 1, 0, 0, 0, 0, time.UTC)
	var out []Regime
	for m := from; !m.After(to); m = m.AddDate(0, 1, 0) {
		if r, ok := p.RegimeAt(m, cfg); ok {
			out = append(out, r)
		}
	}
	return out
}
