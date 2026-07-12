package decumul

import (
	"math"

	"github.com/bpineau/pofo/pkg/scenario"
)

// PathResult is the outcome of one simulated decumulation path. Wealth has
// Years+1 points: Wealth[0] is the starting capital and Wealth[k] is total
// real wealth (growth + buffer) at the end of year k. Ruined latches true
// the first year a withdrawal cannot be funded; RuinYear records that year
// (0-based) and stays -1 on a surviving path. Spend has one point per year:
// the net real spending actually delivered to the household that year, after
// any flex cut, guardrails move or under-delivery, so the series shows the
// lived standard of living (its dips and their duration), not the plan.
// FirstCut and CutYears account the spending cuts: FirstCut is the first year
// the household lived below its uncut standard (flex cut, guardrails cut or
// under-delivery), -1 when it never did, and CutYears counts such years.
type PathResult struct {
	Wealth    []float64
	Spend     []float64
	Ruined    bool
	RuinYear  int
	FirstCut  int
	CutYears  int
	TaxPaid   float64
	Withdrawn float64
	Ret10     float64 // annualized real market return of the first decade (sequence risk)
}

// ruinAt latches ruin at year k, keeping the first occurrence.
func (r *PathResult) ruinAt(k int) {
	if !r.Ruined {
		r.Ruined, r.RuinYear = true, k
	}
}

// cutAt accounts one year lived below the uncut spending standard.
func (r *PathResult) cutAt(k int) {
	if r.FirstCut < 0 {
		r.FirstCut = k
	}
	r.CutYears++
}

// newPathResult prepares a result with the wealth and spend series allocated
// and RuinYear at its -1 sentinel.
func newPathResult(capital float64, years int) PathResult {
	res := PathResult{
		Wealth:   make([]float64, years+1),
		Spend:    make([]float64, years),
		RuinYear: -1,
		FirstCut: -1,
	}
	res.Wealth[0] = capital
	return res
}

// RunPath simulates one path under the returns sequence (one return per
// year; missing years are treated as 0). The order each year is: compute the
// net need after cashflows, apply the flex cut on deep drawdowns, withdraw
// via the bucket rule (buffer first while underwater, else growth + refill),
// then grow the sleeves. A year is ruin when it cannot deliver the full net
// need, i.e. when the gross required exceeds the available liquidity; only the
// net actually delivered is accounted, never the requested amount.
func (p Plan) RunPath(returns scenario.Sequence) PathResult {
	target := p.Buffer.Years * p.NeedAnnual
	buffer := target
	if buffer > p.Capital {
		buffer = p.Capital
	}
	pks := pocketOps(p.newPockets(p.Capital - buffer))

	drawTh := p.Buffer.drawThreshold()
	refillCap := p.Buffer.refillCap()

	res := newPathResult(p.Capital, p.Years)
	res.Ret10 = firstDecadeReturn(returns, min(10, p.Years), 1)
	peak := p.Capital
	spending := p.NeedAnnual         // dynamic spending level for the guardrails rule
	level := p.NeedAnnual            // ratcheted spending level (fixed/flex policy)
	bounded := p.NeedAnnual          // last delivered level for the bounded-percent rule
	lastRaise := -p.Ratchet.Cooldown // so a first raise is never cooldown-blocked

	// drawBuffer takes up to want euros from the buffer (no tax), returning the
	// amount actually taken.
	drawBuffer := func(want float64) float64 {
		take := want
		if take > buffer {
			take = buffer
		}
		buffer -= take
		return take
	}

	for k := 0; k < p.Years; k++ {
		pks.newYear()
		growth := pks.total()
		total := growth + buffer
		if total <= 0 {
			res.ruinAt(k)
			// remaining years stay at 0.
			break
		}
		if total > peak {
			peak = total
		}
		dd := 1 - total/peak

		// uncut is the year's reference standard of living: what would be
		// spent with no flex cut and no guardrails move. Delivering less
		// counts the year as "cut" (cutAt), whatever the cause.
		var need, uncut float64
		if p.Amortize {
			// Amortization-based (ABW/TPAW): the actuarial payment exhausting
			// the AFTER-TAX liquidation value over the remaining horizon (the
			// gross wealth is not net-deliverable, so amortizing it would
			// manufacture a fake final-years shortfall). uncut stays the
			// fixed reference standard, so lean years count as lived cuts.
			wNet := pks.liquidationNet() + buffer
			need = math.Min(pmt(wNet, p.AmortReturn, p.Years-k), wNet*(1-1e-9))
			uncut = p.needAt(k)
		} else if p.Bounded.active() {
			// Bounded percent-of-portfolio (Vanguard dynamic spending): target
			// a share of wealth, move at most Up/Down from last year's level.
			bounded = p.Bounded.clampStep(p.Bounded.Pct*total, bounded)
			need = bounded
			uncut = p.needAt(k)
		} else if p.Percent > 0 {
			// Percentage-of-portfolio (VPW): spend a fixed share of current
			// wealth. uncut stays the fixed reference standard, so years where the
			// rule delivers less than that count as a lived cut.
			need = p.Percent * total
			uncut = p.needAt(k)
		} else if p.Guard.active() {
			spending = p.Guard.adjust(spending, total)
			need = p.netOf(spending*p.schedAt(k), k)
			uncut = p.needAt(k)
		} else {
			level, lastRaise = p.Ratchet.raise(level, total, p.Capital, k, lastRaise)
			need = p.netOf(level*p.schedAt(k), k)
			uncut = need
			if p.Flex.Cut > 0 && p.Flex.triggered(dd, need, total) {
				need *= 1 - p.Flex.Cut
			}
		}

		// Deliver the net need, each source falling back to the other: the
		// buffer first while underwater (it sells nothing, hence no tax),
		// otherwise growth first with a refill of the buffer from any surplus.
		var delivered float64
		if dd > drawTh && buffer > 0 {
			delivered = drawBuffer(need)
			delivered += pks.sell(need-delivered, &res.TaxPaid)
		} else {
			delivered = pks.sell(need, &res.TaxPaid)
			delivered += drawBuffer(need - delivered)
			if refill := target - buffer; refill > 0 && pks.total() > 0 && p.Buffer.refillsAt(k) {
				if cap := pks.total() * refillCap; refill > cap {
					refill = cap
				}
				if refill > p.NeedAnnual {
					refill = p.NeedAnnual
				}
				buffer += pks.sell(refill, &res.TaxPaid)
			}
		}
		res.Withdrawn += delivered
		res.Spend[k] = delivered
		if delivered < uncut-1e-6 {
			res.cutAt(k)
		}
		if delivered < need-1e-6 {
			res.ruinAt(k)
		}
		buffer = pks.settle(buffer) // a stub without a cap may oversell
		if pks.total()+buffer <= 0 {
			res.ruinAt(k)
		}

		pks.grow(ret(returns, k))
		buffer *= 1 + p.Buffer.RealReturn
		res.Wealth[k+1] = pks.total() + buffer
	}
	return res
}

// ret returns the k-th return, or 0 when the sequence is shorter.
func ret(s scenario.Sequence, k int) float64 {
	if k < len(s) {
		return s[k]
	}
	return 0
}

// firstDecadeReturn annualizes the compounded market return of the first n
// periods of a sequence, with perYear periods per year (1 for the annual
// kernel, 12 for the monthly one). It measures the sequence-of-returns luck a
// retirement is dealt in its decisive first decade, independent of the
// withdrawal policy. Returns 0 for an empty window.
func firstDecadeReturn(s scenario.Sequence, n, perYear int) float64 {
	if n > len(s) {
		n = len(s)
	}
	if n <= 0 {
		return 0
	}
	growth := 1.0
	for k := 0; k < n; k++ {
		growth *= 1 + s[k]
	}
	if growth <= 0 {
		return -1
	}
	return math.Pow(growth, float64(perYear)/float64(n)) - 1
}
