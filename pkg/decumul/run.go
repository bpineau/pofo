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
//
// The last block is the household's lifetime. Without a Plan.Lifetime it still
// reads correctly: LifeYears is the full horizon, Outlived is true and Estate
// is the terminal wealth, since the fixed horizon is the case where the
// household is certain to reach the end. With one, the series are FROZEN after
// death (Wealth holds the estate, Spend holds 0) rather than zeroed, so an
// ordinary death is never mistaken for a ruin by a drawdown or a terminal
// statistic; the statistics that are genuinely per-lifetime are bounded by
// LifeYears instead.
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
	LifeYears int     // whole years the household lived, capped at the horizon
	Outlived  bool    // it was still alive at the horizon (censored, not dead)
	Estate    float64 // total real wealth at the household's end
	Annuity   float64 // cumulative real annuity income received over the path
	Premium   float64 // net premium actually converted into an annuity
	// Received is the cumulative real income from OUTSIDE the portfolio over
	// the lived years: cashflows after any reversion, plus the annuity (so it
	// includes Annuity). Spend records only what the portfolio delivered, since
	// income is netted off the budget before anything is sold; the household's
	// standard of living is the two together. A year's income above the year's
	// budget is still counted as received, the model having nowhere to reinvest
	// it.
	Received float64
}

// end is the index in Wealth of the household's last point. A PathResult built
// by hand (LifeYears unset) reads as the full series.
func (r PathResult) end() int {
	if r.LifeYears > 0 && r.LifeYears < len(r.Wealth) {
		return r.LifeYears
	}
	return len(r.Wealth) - 1
}

// spendYears is how many Spend entries the household actually lived through.
func (r PathResult) spendYears() int {
	if r.LifeYears > 0 && r.LifeYears < len(r.Spend) {
		return r.LifeYears
	}
	return len(r.Spend)
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
// and RuinYear at its -1 sentinel. buf, when it holds seriesLen(years) zeroed
// floats, backs the two series instead of a fresh allocation: the ensemble
// drivers hand out windows of one arena, so a thousand-path run allocates the
// series once rather than once per path (see SimulateOn).
func newPathResult(capital float64, years int, buf []float64) PathResult {
	// Wealth (years+1) and Spend (years) are the two per-path series, built
	// millions of times per page render; they share one backing slice, handed
	// out as two non-overlapping, capacity-capped windows, so a path costs at
	// most one allocation (none at all off an arena) without changing the
	// public fields.
	if len(buf) < seriesLen(years) {
		buf = make([]float64, seriesLen(years))
	}
	res := PathResult{
		Wealth:   buf[: years+1 : years+1],
		Spend:    buf[years+1 : seriesLen(years) : seriesLen(years)],
		RuinYear: -1,
		FirstCut: -1,
	}
	res.Wealth[0] = capital
	return res
}

// seriesLen is the number of floats one path's two series need.
func seriesLen(years int) int { return 2*years + 1 }

// RunPath simulates one path under the returns sequence (one return per
// year; missing years are treated as 0) and the household's drawn lifespans.
// The order each year is: buy the annuity if this is its year, compute the
// net need after income, apply the flex cut on deep drawdowns, withdraw via
// the bucket rule (buffer first while underwater, else growth + refill), then
// grow the sleeves. A year is ruin when it cannot deliver the full net need,
// i.e. when the gross required exceeds the available liquidity; only the net
// actually delivered is accounted, never the requested amount.
//
// The zero Lives means no draw: the path runs the plan's fixed horizon. That
// is what a plan without a Lifetime always does, and it is how a caller asks
// for the horizon kernel explicitly. Simulate draws the lifespans for you.
func (p Plan) RunPath(returns scenario.Sequence, lives Lives) PathResult {
	return p.runPathAnnual(returns, lives, nil)
}

// runPathAnnual is RunPath over an optional caller-owned arena window for the
// path's two series (nil = allocate them here).
func (p Plan) runPathAnnual(returns scenario.Sequence, lives Lives, buf []float64) PathResult {
	target := p.Buffer.Years * p.NeedAnnual
	buffer := target
	if buffer > p.Capital {
		buffer = p.Capital
	}
	// The single-sleeve case (no Envelopes) is by far the common one: give it a
	// stack array so a path's pockets cost no allocation at all.
	var pocketBuf [1]pocket
	pks := pocketOps(p.newPockets(pocketBuf[:0], p.Capital-buffer))

	drawTh := p.Buffer.drawThreshold()
	refillCap := p.Buffer.refillCap()

	lf := p.life(lives)
	end := lf.end()

	res := newPathResult(p.Capital, p.Years, buf)
	res.Ret10 = firstDecadeReturn(returns, min(10, p.Years), 1)
	peak := p.Capital
	spending := p.NeedAnnual         // dynamic spending level for the guardrails rule
	level := p.NeedAnnual            // ratcheted spending level (fixed/flex policy)
	bounded := p.NeedAnnual          // last delivered level for the bounded-percent rule
	lastRaise := -p.Ratchet.Cooldown // so a first raise is never cooldown-blocked
	ratchetActive := p.Ratchet.active()
	// The annuity is bought in one year and one year only: testing the year
	// here keeps a no-op call, on a Plan-sized receiver, out of every other.
	annuityYear := -1
	if p.Annuity != nil && p.Lifetime != nil {
		annuityYear = p.Annuity.Year
	}
	// newYear() only matters when a pocket carries per-year tax state (AVTax);
	// the common CTOFlatTax carries none, so decide once per path whether the
	// per-year call is needed rather than type-asserting every pocket every year.
	yearlyTax := false
	for i := range pks {
		if _, ok := pks[i].tax.(YearlyTax); ok {
			yearlyTax = true
			break
		}
	}

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

	for k := 0; k < end; k++ {
		if yearlyTax {
			pks.newYear()
		}
		if k == annuityYear {
			p.buyAnnuity(k, pks, &res, &lf)
		}
		res.Annuity += lf.annuityAt(k)
		// The year's income from outside the portfolio, read ONCE: every
		// spending rule below nets the same figure off its budget (netAfter,
		// needAtWith), rather than rescanning the cashflows two or three times
		// per year for the same answer.
		inc := p.income(k, lf)
		res.Received += inc
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
		// Every rule below sets the HOUSEHOLD budget for the year; pensions
		// and side income fund it first (netAfter) and the portfolio delivers
		// the remainder, exactly like the fixed rule. Without the netting,
		// the wealth-based rules would silently withdraw the pension's share
		// on top of it, making them incomparable in the model strip.
		if p.Amortize {
			// Amortization-based (ABW/TPAW): the actuarial payment exhausting
			// the AFTER-TAX liquidation value plus the present value of the
			// future cashflows over the remaining horizon (the gross wealth
			// is not net-deliverable, so amortizing it would manufacture a
			// fake final-years shortfall; ignoring a future pension would
			// understate today's sustainable budget). uncut stays the fixed
			// reference standard, so lean years count as lived cuts.
			wNet := pks.liquidationNet() + buffer
			budget := pmt(wNet+p.cashflowPV(k, p.AmortReturn, lf), p.AmortReturn, p.planYears()-k)
			need = math.Min(netAfter(budget, inc), wNet*(1-1e-9))
			uncut = p.needAtWith(k, lf, inc)
		} else if p.Bounded.active() {
			// Bounded percent-of-portfolio (Vanguard dynamic spending): target
			// a share of wealth, move at most Up/Down from last year's level.
			bounded = p.Bounded.clampStep(p.Bounded.Pct*total, bounded)
			need = netAfter(bounded, inc)
			uncut = p.needAtWith(k, lf, inc)
		} else if p.Percent > 0 {
			// Percentage-of-portfolio (VPW): spend a fixed share of current
			// wealth. uncut stays the fixed reference standard, so years where the
			// rule delivers less than that count as a lived cut.
			need = netAfter(p.Percent*total, inc)
			uncut = p.needAtWith(k, lf, inc)
		} else if p.RiskGuard.active() {
			// Risk-based guardrail: the same ±10 % moves as Guyton-Klinger,
			// but the band is the safe rate of the REMAINING horizon and the
			// rate is read on total wealth, pensions to come included.
			wealth := total + p.cashflowPV(k, p.RiskGuard.PVRate, lf)
			spending = p.RiskGuard.adjust(spending, wealth, k)
			need = netAfter(spending*p.schedAt(k)*lf.spendFactor(k), inc)
			uncut = p.needAtWith(k, lf, inc)
		} else if p.Guard.active() {
			spending = p.Guard.adjust(spending, total)
			need = netAfter(spending*p.schedAt(k)*lf.spendFactor(k), inc)
			uncut = p.needAtWith(k, lf, inc)
		} else {
			if ratchetActive {
				level, lastRaise = p.Ratchet.raise(level, total, p.Capital, k, lastRaise)
			}
			need = netAfter(level*p.schedAt(k)*lf.spendFactor(k), inc)
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
			if refill := target - buffer; refill > 0 && p.Buffer.refillsAt(k) {
				if avail := pks.total(); avail > 0 { // one total() for both the test and the cap
					if cap := avail * refillCap; refill > cap {
						refill = cap
					}
					if refill > p.NeedAnnual {
						refill = p.NeedAnnual
					}
					buffer += pks.sell(refill, &res.TaxPaid)
				}
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
	res.close(lf)
	return res
}

// close records the household's lifetime on a finished path and freezes the
// wealth series at the estate past its end, so a death is never read as a
// crash to zero. A path that ruined early already sits at zero, so the fill is
// a no-op there.
func (r *PathResult) close(l life) {
	end := l.end()
	r.LifeYears, r.Outlived = end, l.outlived()
	for k := end + 1; k < len(r.Wealth); k++ {
		r.Wealth[k] = r.Wealth[end]
	}
	r.Estate = r.Wealth[end]
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
