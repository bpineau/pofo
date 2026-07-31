package permanent

import "math"

// Params holds the (reconstructed) tunables of the allocator. Darcet does not
// disclose his weight function; these encode his qualitative rules and are NOT
// fitted optima. See docs/darcet-permanent-portfolio-design.md, section on the
// epistemic ledger, before changing them.
type Params struct {
	// EquityMax caps the equity sleeve at full paradise. Above 1 it saturates
	// (weights are clamped to [0,1]); ~1.3 is a moderate default, ~1.6 the
	// return/drawdown sweet spot found post hoc.
	EquityMax float64
	// Damping is the exponent on the normalized distance to hell. 2 is Darcet's
	// quadratic (1/d^2) damping, the mechanism that keeps drawdowns shallow; 1
	// (linear) earns more raw return but deepens drawdowns badly.
	Damping float64
	// DefensivePoles are the (slope, realShort) attractors of the defensive
	// sleeve, in order bonds, cash, gold. DefensiveEps softens the 1/d^2 weight
	// near a pole so no single sleeve blows up.
	DefensivePoles [3][2]float64
	DefensiveEps   float64
}

// DefaultParams is the reconstruction validated in the design doc: quadratic
// damping, a moderate equity cap, and the three defensive poles.
func DefaultParams() Params {
	return Params{
		EquityMax:      1.3,
		Damping:        2,
		DefensivePoles: [3][2]float64{{2.0, 1.0}, {0.0, 2.5}, {0.5, -2.5}}, // bonds, cash, gold
		DefensiveEps:   0.25,
	}
}

// Allocation is a target weight on each of the four sleeves. The fields sum to 1.
type Allocation struct{ Equity, Bonds, Cash, Gold float64 }

// EquityWeight is the target equity fraction for the regime: the squared (or
// Damping-th power) normalized distance from the world point to the stagflation
// pole (0,1), scaled by EquityMax and clamped to [0,1]. Paradise (1,0) is the
// far corner, at the maximum distance sqrt(2).
func (r Regime) EquityWeight(pm Params) float64 {
	dHell := math.Hypot(r.GrowthBreadth-0, r.InflationBreadth-1)
	w := pm.EquityMax * math.Pow(dHell/math.Sqrt2, pm.Damping)
	return clamp01(w)
}

// defensiveSplit returns the bonds/cash/gold split of the defensive sleeve:
// inverse-square weights toward each pole in (slope, realShort) space,
// normalized to sum to 1.
func (r Regime) defensiveSplit(pm Params) [3]float64 {
	var w [3]float64
	var sum float64
	for i, p := range pm.DefensivePoles {
		dx, dy := r.Slope-p[0], r.RealShort-p[1]
		w[i] = 1.0 / (dx*dx + dy*dy + pm.DefensiveEps)
		sum += w[i]
	}
	for i := range w {
		w[i] /= sum
	}
	return w
}

// Allocate maps the regime to a four-sleeve target: the equity weight from the
// growth x inflation quadrant, the rest split across bonds/cash/gold by the
// monetary quadrant.
func (r Regime) Allocate(pm Params) Allocation {
	eq := r.EquityWeight(pm)
	d := r.defensiveSplit(pm)
	rest := 1 - eq
	return Allocation{Equity: eq, Bonds: rest * d[0], Cash: rest * d[1], Gold: rest * d[2]}
}

func clamp01(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}
