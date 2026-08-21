package decumul

import (
	"math"

	"github.com/bpineau/pofo/pkg/scenario"
)

// RunPathMonthly simulates one path stepping monthly: it withdraws NeedAnnual/12
// each month, re-evaluates the drawdown, flex cut and bucket rule monthly, and
// applies one monthly real return per step. returns are monthly (Years*12
// values; missing months are treated as 0).
//
// Durations that are naturally in years stay in years: the buffer is sized at
// Buffer.Years × NeedAnnual, the horizon is Years, and cashflows switch on at
// their FromYear. The buffer's annual real return is applied as its 12th root
// each month. Wealth is reported at annual granularity (Years+1 points, one per
// year-end) so the Outcome statistics read the same as for the annual kernel.
//
// Mortality stays a yearly decision here too: the household's drawn lifespans
// are in whole years, so the kernel simply stops at the end of the last year
// it lives. A monthly death date would change one year's spending by at most
// half a year's need, far below the Monte-Carlo error of any statistic the
// package reports.
//
// This is a distinct kernel from RunPath (the validated annual reference); it
// has its own validation tests.
func (p Plan) RunPathMonthly(returns scenario.Sequence, lives Lives) PathResult {
	target := p.Buffer.Years * p.NeedAnnual
	buffer := target
	if buffer > p.Capital {
		buffer = p.Capital
	}
	pks := pocketOps(p.newPockets(p.Capital - buffer))

	drawTh := p.Buffer.drawThreshold()
	refillCap := p.Buffer.refillCap()
	bufferStep := math.Pow(1+p.Buffer.RealReturn, 1.0/12) - 1
	monthlyNeedCap := p.NeedAnnual / 12

	lf := p.life(lives)
	end := lf.end()

	res := newPathResult(p.Capital, p.Years)
	res.Ret10 = firstDecadeReturn(returns, min(120, p.Years*12), 12)
	peak := p.Capital
	spending := p.NeedAnnual         // dynamic spending level for the guardrails rule
	level := p.NeedAnnual            // ratcheted spending level (fixed/flex policy)
	lastRaise := -p.Ratchet.Cooldown // so a first raise is never cooldown-blocked

	drawBuffer := func(want float64) float64 {
		take := want
		if take > buffer {
			take = buffer
		}
		buffer -= take
		return take
	}

	// Guardrails run at the monthly cadence in this kernel, with the moves
	// rescaled so a persistent breach compounds to the same annual intensity
	// (Guardrails.stepped): the rule reacts to the market as it moves instead
	// of gambling on the state of the world at one anniversary date.
	guardM := p.Guard.stepped(12)
	riskM := p.RiskGuard.stepped(12)
	adaptive := p.Guard.active() || p.RiskGuard.active()

	ruined := false
	for k := 0; k < end && !ruined; k++ {
		// The ratchet, the annuity purchase and stateful taxes stay yearly
		// decisions, taken at the start of each year against current wealth.
		pks.newYear()
		p.buyAnnuity(k, pks, &res, &lf)
		res.Annuity += lf.annuityAt(k)
		res.Received += p.income(k, lf)
		if !adaptive {
			level, lastRaise = p.Ratchet.raise(level, pks.total()+buffer, p.Capital, k, lastRaise)
		}
		// The year's uncut reference standard, for the cut accounting: the
		// initial level under guardrails, the ratcheted level otherwise.
		uncut := p.needAt(k, lf)
		if !adaptive {
			uncut = p.netOf(level*p.schedAt(k)*lf.spendFactor(k), k, lf)
		}
		for m := range 12 {
			total := pks.total() + buffer
			if total <= 0 {
				res.ruinAt(k)
				ruined = true
				break
			}
			if total > peak {
				peak = total
			}
			dd := 1 - total/peak

			var need float64
			if p.RiskGuard.active() {
				spending = riskM.adjust(spending, total+p.cashflowPV(k, p.RiskGuard.PVRate, lf), k)
				need = p.netOf(spending*p.schedAt(k)*lf.spendFactor(k), k, lf) / 12
			} else if p.Guard.active() {
				spending = guardM.adjust(spending, total)
				need = p.netOf(spending*p.schedAt(k)*lf.spendFactor(k), k, lf) / 12
			} else {
				yearNeed := p.netOf(level*p.schedAt(k)*lf.spendFactor(k), k, lf)
				need = yearNeed / 12
				if p.Flex.Cut > 0 && p.Flex.triggered(dd, yearNeed, total) {
					need *= 1 - p.Flex.Cut
				}
			}

			// Deliver the month's net need, each source falling back to the
			// other (buffer first while underwater, else growth + a refill).
			var delivered float64
			if dd > drawTh && buffer > 0 {
				delivered = drawBuffer(need)
				delivered += pks.sell(need-delivered, &res.TaxPaid)
			} else {
				delivered = pks.sell(need, &res.TaxPaid)
				delivered += drawBuffer(need - delivered)
				if refill := target - buffer; refill > 0 && pks.total() > 0 && p.Buffer.refillsAt(k) {
					if cap := pks.total() * refillCap / 12; refill > cap {
						refill = cap
					}
					if refill > monthlyNeedCap {
						refill = monthlyNeedCap
					}
					buffer += pks.sell(refill, &res.TaxPaid)
				}
			}
			res.Withdrawn += delivered
			res.Spend[k] += delivered
			if delivered < need-1e-6 {
				res.ruinAt(k)
			}
			buffer = pks.settle(buffer)
			if pks.total()+buffer <= 0 {
				res.ruinAt(k)
			}

			pks.grow(ret(returns, k*12+m))
			buffer *= 1 + bufferStep
		}
		if res.Spend[k] < uncut-1e-6 {
			res.cutAt(k)
		}
		res.Wealth[k+1] = pks.total() + buffer
	}
	res.close(lf)
	return res
}
