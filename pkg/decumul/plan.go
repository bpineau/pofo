package decumul

import (
	"math"

	"github.com/bpineau/pofo/pkg/scenario"
)

// Cashflow is a real annual income (e.g. a pension or side income) received
// from FromYear (0-based) until ToYear (exclusive); it reduces the net amount
// sold from the portfolio. A zero ToYear runs to the horizon, so a lifelong
// pension leaves it unset while a temporary rental/activity income bounds it.
// Cashflows are modelled as income, not as an asset.
type Cashflow struct {
	FromYear int
	ToYear   int // exclusive end; 0 = to the horizon
	Annual   float64
}

// BufferSleeve is a low-volatility cash or inflation-linked pocket sized at
// Years times annual spending (capped at the capital). It earns RealReturn
// and is drained first while the portfolio drawdown exceeds DrawThreshold;
// otherwise it is refilled from growth, by at most RefillCap of growth/year.
//
// RealReturn also distinguishes the sleeve's nature: an inflation-linked sleeve
// holds its value in real terms (RealReturn ≈ 0), while pure cash bleeds to
// inflation (a negative real return), so the same field models both.
//
// RefillStopYear implements a melting / glidepath buffer: once that year is
// reached the buffer is no longer refilled (0 = always refill), so the sleeve
// can cover the early sequence-risk window and then be left to run down.
//
// DrawThreshold and RefillCap are pointers so a nil leaves the default while an
// explicit zero is honoured (always draw the buffer, resp. never refill), which
// a plain zero field could not express.
type BufferSleeve struct {
	Years          float64
	RealReturn     float64
	DrawThreshold  *float64 // nil = 0.10; 0 = always tap the buffer first
	RefillCap      *float64 // nil = 0.50; 0 = never refill
	RefillStopYear int      // stop refilling from this year (0 = never stop)
}

// refillsAt reports whether the buffer is still refilled in the given year,
// implementing the glidepath cutoff.
func (b BufferSleeve) refillsAt(year int) bool {
	return b.RefillStopYear == 0 || year < b.RefillStopYear
}

// drawThreshold resolves DrawThreshold, applying the 0.10 default when unset.
func (b BufferSleeve) drawThreshold() float64 {
	if b.DrawThreshold == nil {
		return 0.10
	}
	return *b.DrawThreshold
}

// refillCap resolves RefillCap, applying the 0.50 default when unset.
func (b BufferSleeve) refillCap() float64 {
	if b.RefillCap == nil {
		return 0.50
	}
	return *b.RefillCap
}

// FlexRule cuts the year's spending by Cut (e.g. 0.25) whenever the
// portfolio drawdown exceeds Threshold (e.g. 0.20) or, when WRThreshold is
// set, whenever the current withdrawal rate (this year's net need over the
// portfolio value) exceeds it. The two triggers OR together, matching the
// written-rules style "cut when drawdown > 20% or current rate > 3.6%"; a
// zero WRThreshold keeps the drawdown-only behaviour. A zero rule is inactive.
type FlexRule struct {
	Threshold, Cut float64
	WRThreshold    float64
}

// triggered reports whether the cut applies given the current drawdown and
// the year's net need against the portfolio value.
func (f FlexRule) triggered(dd, need, total float64) bool {
	if dd > f.Threshold {
		return true
	}
	return f.WRThreshold > 0 && total > 0 && need/total > f.WRThreshold
}

// Ratchet is a Kitces-style only-up spending rule: once total real wealth
// exceeds Trigger times the initial capital, the spending level rises by Step
// (real euros per year), at most once per Cooldown years and never past Cap.
// MaxWR optionally vetoes a raise while the current withdrawal rate is above
// it, so the level only ratchets when the rate is comfortable. Raises are
// permanent (the level never steps back down; the flex cut still applies on
// top in downturns). A zero Trigger or Step leaves the rule inactive.
type Ratchet struct {
	Trigger  float64 // raise when wealth > Trigger × initial capital (e.g. 1.2)
	Step     float64 // real € added to the annual spending level per raise
	Cap      float64 // ceiling on the ratcheted level (0 = none)
	Cooldown int     // minimum years between raises (0 = none)
	MaxWR    float64 // veto raises while spending/wealth exceeds this (0 = none)
}

// active reports whether the ratchet is configured.
func (r Ratchet) active() bool { return r.Trigger > 0 && r.Step > 0 }

// raise applies the rule for one year: given the current level, wealth,
// initial capital, year and last raise year, it returns the (possibly raised)
// level and the updated last raise year.
func (r Ratchet) raise(level, total, capital0 float64, year, lastRaise int) (float64, int) {
	switch {
	case !r.active() || total <= 0 || total < r.Trigger*capital0:
		return level, lastRaise
	case year-lastRaise < r.Cooldown:
		return level, lastRaise
	case r.MaxWR > 0 && level/total > r.MaxWR:
		return level, lastRaise
	case r.Cap > 0 && level >= r.Cap:
		return level, lastRaise
	}
	level += r.Step
	if r.Cap > 0 && level > r.Cap {
		level = r.Cap
	}
	return level, year
}

// Guardrails is a Guyton-Klinger-style withdrawal rule: real spending starts at
// NeedAnnual and is re-checked against the current withdrawal rate
// (spending / portfolio). Above Upper it is cut by Cut, below Lower it is raised
// by Raise, keeping spending inside a band as the portfolio moves. It is a
// richer alternative to FlexRule (a single drawdown-triggered cut); when active
// it replaces FlexRule. A zero rule is inactive.
//
// Floor, when set, is the yearly real spending level cuts never go below (the
// household's incompressible standard, in euros). Without it, repeated cuts
// compound geometrically in a persistent bear: the rule then trades nearly all
// ruin risk for an unbounded lifestyle risk. A floor re-creates some ruin (the
// floor itself can prove unsustainable) but bounds the descent.
//
// The annual kernel evaluates the rule yearly (Guyton-Klinger's own cadence);
// the monthly kernel evaluates it every month at the pace-preserving step
// (see stepped) so a persistent breach produces the same annual intensity
// without the anniversary-date lottery of a once-a-year check.
type Guardrails struct {
	Upper, Lower float64 // withdrawal-rate guardrails, e.g. 0.06 and 0.03
	Cut, Raise   float64 // proportional spending adjustments, e.g. 0.10 each
	Floor        float64 // spending level cuts stop at (euros/yr); 0 = none
}

// active reports whether the guardrails band is set.
func (g Guardrails) active() bool { return g.Upper > 0 && g.Lower > 0 }

// adjust moves spending toward the band given the current withdrawal rate.
func (g Guardrails) adjust(spending, portfolio float64) float64 {
	if portfolio <= 0 {
		return spending
	}
	switch wr := spending / portfolio; {
	case wr > g.Upper:
		spending *= 1 - g.Cut
		if g.Floor > 0 && spending < g.Floor {
			spending = g.Floor
		}
	case wr < g.Lower:
		spending *= 1 + g.Raise
	}
	return spending
}

// stepped returns the rule with its moves rescaled to n evaluations per year,
// so a breach persisting a full year compounds to the same annual adjustment:
// (1-cut_n)^n == 1-Cut and (1+raise_n)^n == 1+Raise. The band and the floor
// are levels, not rates, and stay unchanged.
func (g Guardrails) stepped(n int) Guardrails {
	if n <= 1 {
		return g
	}
	inv := 1 / float64(n)
	g.Cut = 1 - math.Pow(1-g.Cut, inv)
	g.Raise = math.Pow(1+g.Raise, inv) - 1
	return g
}

// Tax grosses up a net withdrawal taken by selling part of a growth sleeve
// whose market value is growth and whose cost basis is cost. It returns the
// gross amount to sell, the new cost basis after the sale, and the tax paid.
type Tax interface {
	GrossUp(net, growth, cost float64) (gross, newCost, taxPaid float64)
}

// CTOFlatTax is the French taxable-account flat tax: only the realised gain
// fraction of a sale is taxed at Rate, so the effective rate starts low and
// drifts toward Rate as unrealised gains compound.
type CTOFlatTax struct{ Rate float64 }

// GrossUp implements Tax.
func (t CTOFlatTax) GrossUp(net, growth, cost float64) (gross, newCost, taxPaid float64) {
	if growth <= 0 {
		return net, cost, 0
	}
	gainFrac := 1 - cost/growth
	if gainFrac < 0 {
		gainFrac = 0
	}
	eff := t.Rate * gainFrac
	gross = net
	if eff < 1 {
		gross = net / (1 - eff)
	}
	if gross > growth {
		gross = growth // sale capped at the available market value
	}
	newCost = cost * (1 - gross/growth)
	// Tax is the effective rate on the gross actually sold; deriving it from
	// the requested net (gross - net) would misstate it, even turning negative,
	// when the sale was capped and delivers less than net.
	return gross, newCost, gross * eff
}

// Plan is a full decumulation scenario.
type Plan struct {
	Capital    float64
	NeedAnnual float64
	Cashflows  []Cashflow
	Years      int
	Buffer     BufferSleeve
	Flex       FlexRule
	Tax        Tax
	Source     scenario.Source
	Guard      Guardrails // optional Guyton-Klinger spending rule (replaces Flex when active)
	Ratchet    Ratchet    // optional only-up spending rule (ignored while Guard is active)
	// Percent, when > 0, switches to a percentage-of-portfolio (VPW-style)
	// withdrawal: each year the household spends Percent of current total real
	// wealth instead of a fixed real need. It cannot ruin (spending is always a
	// fraction of what remains) but the standard of living swings with the
	// market, so it trades ruin risk for spending volatility, the other end of
	// the decumulation frontier. It overrides Flex, Guard and Ratchet, and runs
	// annually (a rebalancing rule), never monthly.
	Percent float64
	// Amortize, when true, switches to an amortization-based withdrawal (the
	// ABW / TPAW family): each year the household spends the actuarial
	// payment that would exhaust CURRENT wealth exactly over the REMAINING
	// horizon at the expected real return AmortReturn, like a mortgage run in
	// reverse and re-quoted yearly. It cannot ruin before the horizon
	// (spending is always the sustainable share of what remains) and it
	// cannot die rich by accident (the payment rises as the horizon
	// shortens); the cost is that income tracks the market, though more
	// smoothly than VPW because the shortening horizon lifts the payout rate
	// over time. Overrides every other spending rule; annual.
	Amortize    bool
	AmortReturn float64 // expected real return the amortization assumes
	// Bounded, when active, is the bounded percent-of-portfolio rule
	// (Vanguard's "dynamic spending"): each year target Pct of current
	// wealth, but never move more than Up above or Down below last year's
	// real spending. The bounds smooth VPW's swings into small yearly steps;
	// unlike VPW it CAN ruin (in a deep bear the floor-bounded spending can
	// outrun the portfolio). Overrides Flex, Guard and Ratchet; annual.
	Bounded BoundedPct
	// Envelopes optionally splits the growth sleeve across tax wrappers
	// (CTO/PEA/AV), drained in slice order; nil keeps the single sleeve
	// taxed by Tax. See Envelope.
	Envelopes []Envelope
	// SpendSchedule optionally scales the base spending year by year (real
	// multipliers): a slow health-cost drift ({1, 1.005, 1.010, …}) or a
	// retirement smile (falling then rising). Years beyond the slice keep a
	// factor of 1; nil means constant real spending.
	SpendSchedule []float64
	// Monthly steps the kernel monthly (RunPathMonthly) instead of annually;
	// the Source must then yield monthly returns (Years*12 per path).
	Monthly bool
}

// runPath dispatches to the monthly or the annual kernel; Simulate and the
// sweeps go through it so a monthly plan is simulated end to end, while the
// annual RunPath stays the validated reference (and its golden tests). The
// wealth-based rules (VPW, ABW, bounded %) are yearly rebalancing decisions
// and always run on the annual kernel.
func (p Plan) runPath(seq scenario.Sequence) PathResult {
	if p.Monthly && p.Percent <= 0 && !p.Amortize && !p.Bounded.active() {
		return p.RunPathMonthly(seq)
	}
	return p.RunPath(seq)
}

// BoundedPct parametrises the bounded percent-of-portfolio rule: target Pct
// of current wealth, with the yearly change in real spending clamped to
// [+Up, -Down] (Vanguard's classic bounds are +5% / -2.5%).
type BoundedPct struct {
	Pct, Up, Down float64
}

// active reports whether the bounded rule is set.
func (b BoundedPct) active() bool { return b.Pct > 0 }

// clampStep bounds this year's target spending by last year's level.
func (b BoundedPct) clampStep(target, prev float64) float64 {
	if hi := prev * (1 + b.Up); target > hi {
		return hi
	}
	if lo := prev * (1 - b.Down); target < lo {
		return lo
	}
	return target
}

// pmt is the begin-of-year payment that exhausts wealth exactly over n
// periods at rate r (the kernel withdraws first, then applies the return, so
// the annuity is due, not ordinary). n <= 0 spends everything.
func pmt(wealth, r float64, n int) float64 {
	if n <= 0 {
		return wealth
	}
	if r <= 0 {
		return wealth / float64(n)
	}
	f := math.Pow(1+r, float64(n))
	return wealth * r * f / (f - 1) / (1 + r)
}

// needAt is the scheduled net spending in a given year: the base need scaled
// by the spend schedule, minus active cashflows, floored at 0.
func (p Plan) needAt(year int) float64 { return p.netOf(p.NeedAnnual*p.schedAt(year), year) }

// schedAt is the spending multiplier for a year: SpendSchedule[year] when
// present, 1 otherwise.
func (p Plan) schedAt(year int) float64 {
	if year < len(p.SpendSchedule) {
		return p.SpendSchedule[year]
	}
	return 1
}

// cashflowPV is the present value, discounted at r to year from, of every
// cashflow between from (inclusive) and the horizon. The amortization rule
// adds it to the portfolio's liquidation value so a future pension raises
// today's sustainable budget, the TPAW treatment of retirement income.
func (p Plan) cashflowPV(from int, r float64) float64 {
	pv := 0.0
	for j := from; j < p.Years; j++ {
		cf := 0.0
		for _, c := range p.Cashflows {
			if j >= c.FromYear && (c.ToYear == 0 || j < c.ToYear) {
				cf += c.Annual
			}
		}
		if cf > 0 {
			pv += cf / math.Pow(1+r, float64(j-from))
		}
	}
	return pv
}

// netOf reduces a gross annual spend by the cashflows active in the year,
// floored at 0. It lets the guardrails rule feed a dynamic spending level
// through the same cashflow netting as the fixed NeedAnnual.
func (p Plan) netOf(spend float64, year int) float64 {
	for _, c := range p.Cashflows {
		if year >= c.FromYear && (c.ToYear == 0 || year < c.ToYear) {
			spend -= c.Annual
		}
	}
	if spend < 0 {
		return 0
	}
	return spend
}
