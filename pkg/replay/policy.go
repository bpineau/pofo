package replay

import (
	"math"
	"sync"

	"github.com/bpineau/pofo/pkg/decumul"
	"github.com/bpineau/pofo/pkg/scenario"
)

// Policy is one named withdrawal rule of the decumulation family: how it is
// labelled (Name in prose, Tag in a narrow table header, NameFR for the French
// book), what it does in plain language, the colour it wears wherever it is
// drawn, and the mutation it applies to a bare plan (one stripped of every
// adaptive rule).
//
// The list is shared so a rule looks and reads the same everywhere: the
// simulator's ruin-versus-lifestyle frontier and the book's historical replays
// are the same rules, in the same order and the same colours.
type Policy struct {
	Name   string
	NameFR string
	Tag    string
	Help   string
	Color  string
	Apply  func(*decumul.Plan)
}

// PolicyConfig is what the rules need from their caller to be configured: the
// plan's initial withdrawal rate (the anchor the rate-based rules calibrate
// their band or their percentage on) and the planning assumptions the two
// horizon-aware rules quote against.
//
// SafeWR is the risk-based guardrail's moving band centre, one rate per year of
// the horizon; SafeRateTable builds a generic one. RaiseCap bounds how far that
// rule may ratchet the standard of living up, in euros per year (0 defaults to
// the planned spend, i.e. protective only). PVRate discounts future cashflows
// into its sensor and AmortReturn is the real return the amortisation rule
// spreads wealth at.
type PolicyConfig struct {
	WR          float64
	SafeWR      []float64
	RaiseCap    float64
	PVRate      float64
	AmortReturn float64
}

// Policies returns the seven rules configured for one plan, from the most rigid
// to the most wealth-proportional.
func Policies(c PolicyConfig) []Policy {
	return []Policy{
		{"Fixed", "Retrait fixe", "Fixed",
			"Bengen's rule: withdraw the same real amount every year, whatever the market does. Maximum lifestyle stability, maximum ruin risk.",
			"#D2402F", func(p *decumul.Plan) {}},
		{"Flex -10%", "Flex −10 %", "Flex",
			"The fixed rule with one reflex: skip the inflation raise and cut spending by 10% in any year the portfolio is more than 20% below its peak.",
			"#C77E17", func(p *decumul.Plan) { p.Flex = decumul.FlexRule{Threshold: 0.20, Cut: 0.10} }},
		{"Guardrails", "Guardrails (GK)", "GK",
			"Guyton-Klinger: watch the current withdrawal rate against a band around the rate you started at; cut spending 10% above the upper rail, raise it 10% below the lower one.",
			"#0C8A47", func(p *decumul.Plan) {
				p.Guard = decumul.Guardrails{Upper: c.WR * 1.2, Lower: c.WR * 0.8, Cut: 0.10, Raise: 0.10}
			}},
		{"Risk guardrails", "Guardrails par risque", "Risk GR",
			"The same 10% moves, a better sensor (Kitces, Morningstar): the band tracks the rate still safe for the horizon you have LEFT, measured on total wealth, instead of a band around the rate you started at.",
			"#15803D", func(p *decumul.Plan) {
				p.RiskGuard = decumul.RiskGuardrails{SafeWR: c.SafeWR,
					Band: 0.20, Cut: 0.10, Raise: 0.10, Cap: c.RaiseCap, PVRate: c.PVRate}
			}},
		{"Bounded %", "% borné", "Bnd %",
			"Vanguard's dynamic spending: aim at a fixed percentage of current wealth, but never move more than +5% or -2.5% from last year's real spending. VPW's discipline, smoothed into small steps.",
			"#BE185D", func(p *decumul.Plan) {
				p.Bounded = decumul.BoundedPct{Pct: c.WR, Up: 0.05, Down: 0.025}
			}},
		{"ABW", "Amortissement (ABW)", "ABW",
			"Amortisation-based withdrawal (the ABW/TPAW family): every year, re-quote the payment that would exhaust what is left over the horizon that remains, like a mortgage run in reverse. Cannot run out before the horizon.",
			"#6D28D9", func(p *decumul.Plan) { p.Amortize, p.AmortReturn = true, c.AmortReturn }},
		{"VPW", "% du portefeuille (VPW)", "VPW",
			"Pure percentage-of-portfolio: spend the same share of what you have each year. It can never run out, but the income follows the market all the way down.",
			"#0B7285", func(p *decumul.Plan) { p.Percent = c.WR }},
	}
}

// safeRateAnchors is how many horizons SafeRateTable actually solves; the table
// interpolates between them. The safe rate is a smooth, monotone function of the
// horizon left, so five anchors reproduce it to a few basis points at a fraction
// of the cost of solving every year.
const safeRateAnchors = 5

// safeRatePaths is the path count of those solves. The table is a planning
// input, not a headline: a few basis points of Monte-Carlo noise on the band's
// centre is invisible next to the +-20 % band itself.
const safeRatePaths = 600

var safeRateCache sync.Map // key -> []float64

// SafeRateTable is the risk-based guardrail's band centre for each year of a
// horizon: the withdrawal rate, as a share of total wealth, that still meets the
// ruin target over the years remaining, solved under an i.i.d. Student-t with
// the given real mean, volatility and tail.
//
// It is scale-invariant by construction (solved on a pension-free plan of unit
// capital), which is what lets one table apply at any wealth level; the
// household's future cashflows enter on the other side of the comparison, as
// discounted wealth. Results are cached per assumption set.
func SafeRateTable(years int, mu, sigma, df, targetRuin float64) []float64 {
	if years <= 0 {
		return nil
	}
	if targetRuin <= 0 {
		targetRuin = 0.05
	}
	type key struct {
		years                 int
		mu, sigma, df, target float64
	}
	k := key{years, mu, sigma, df, targetRuin}
	if v, ok := safeRateCache.Load(k); ok {
		return v.([]float64)
	}

	// Solve the anchors, longest horizon first.
	horizons := make([]int, 0, safeRateAnchors)
	rates := make([]float64, 0, safeRateAnchors)
	for i := range safeRateAnchors {
		n := years - years*i/safeRateAnchors
		if i == safeRateAnchors-1 {
			n = max(1, years/(2*safeRateAnchors)) // the short end, where the rate runs away
		}
		n = max(n, 1)
		if len(horizons) > 0 && n == horizons[len(horizons)-1] {
			continue
		}
		horizons = append(horizons, n)
		rates = append(rates, safeRateAt(n, mu, sigma, df, targetRuin))
	}

	out := make([]float64, years)
	for i := range out {
		out[i] = interpRate(horizons, rates, years-i)
	}
	safeRateCache.Store(k, out)
	return out
}

// safeRateAt solves the safe withdrawal rate for one remaining horizon, on the
// unit-capital, pension-free, tax-free, fixed-rule plan the table is defined on.
func safeRateAt(years int, mu, sigma, df, target float64) float64 {
	const unit = 1e6
	p := decumul.Plan{
		Capital:    unit,
		NeedAnnual: unit * 0.04,
		Years:      years,
		Source:     scenario.ParametricSource{Mu: mu, Sigma: sigma, Df: df, Periods: years},
	}
	// A one-year horizon has no risk to speak of: the whole pot is spendable,
	// and the solver's upper bound would clip an answer of 100 %.
	if years <= 1 {
		return 1
	}
	safe := p.Solve(target, decumul.WithdrawalAxis(0, unit*0.60), safeRatePaths, 0, 11)
	return safe / unit
}

// interpRate reads the table between two solved anchors, linearly in the
// horizon. horizons is descending; a horizon outside the solved range clamps to
// the nearest anchor.
func interpRate(horizons []int, rates []float64, n int) float64 {
	if len(horizons) == 0 {
		return 0
	}
	if n >= horizons[0] {
		return rates[0]
	}
	for i := 1; i < len(horizons); i++ {
		if n >= horizons[i] {
			lo, hi := horizons[i], horizons[i-1]
			if hi == lo {
				return rates[i]
			}
			t := float64(n-lo) / float64(hi-lo)
			return rates[i] + t*(rates[i-1]-rates[i])
		}
	}
	return rates[len(rates)-1]
}

// GeometricReturn is the honest compounding rate implied by an arithmetic mean
// and a volatility (mu - sigma^2/2), floored at zero: what the amortisation rule
// should spread wealth at, and what the risk sensor should discount with.
func GeometricReturn(mu, sigma float64) float64 { return math.Max(0, mu-sigma*sigma/2) }
