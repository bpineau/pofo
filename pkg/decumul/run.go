package decumul

import "github.com/bpineau/pofo/pkg/scenario"

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
	peak := p.Capital
	spending := p.NeedAnnual         // dynamic spending level for the guardrails rule
	level := p.NeedAnnual            // ratcheted spending level (fixed/flex policy)
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
		if p.Guard.active() {
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
