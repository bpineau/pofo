package decumul

import (
	"math/rand/v2"

	"github.com/bpineau/pofo/pkg/scenario"
)

// LedgerYear is one year of one simulated retirement, read the way a household
// would keep its own accounts: what the market did, what came in from outside,
// what the portfolio was asked for and delivered, what was sold to deliver it
// and the tax that cost, and where the wealth stood at both ends. Amounts are
// real, in the units of Plan.Capital; Return is a FRACTION.
//
// The identities a reader can check by hand: Spent is the net the portfolio
// delivered (PathResult.Spend), FromBuffer of it out of the cash buffer and the
// rest out of sales; Sold is the gross that left the growth sleeve (that net
// part of the spending, any buffer refill, any annuity premium) and Tax is
// what those sales paid on top of the net they delivered, so Sold - Tax is
// the net the sleeve gave up; Close is Open after the year's flows and returns
// (PathResult.Wealth[Year+1]).
type LedgerYear struct {
	Year   int     // plan year, 0-based
	Return float64 // the growth sleeve's real return that year, a FRACTION
	Open   float64 // total real wealth at the start of the year, buffer included
	Income float64 // cashflows and annuity received from outside the portfolio
	// Need is the net amount the year's spending rule asked of the
	// portfolio, after Income and after any cut.
	Need       float64
	Spent      float64 // net amount the portfolio delivered (PathResult.Spend)
	FromBuffer float64 // the part of Spent drawn from the cash buffer, untaxed
	Sold       float64 // gross sales out of the growth sleeve, tax included
	Tax        float64 // tax paid on those sales
	Buffer     float64 // the cash buffer at the end of the year, after its return
	Close      float64 // total real wealth at the end of the year
	// Cut reports a year delivered below the uncut standard of living (a flex
	// cut, a guardrail move, a lean wealth-based year); Broke a year at or
	// after the path's ruin, when the portfolio no longer funds the need.
	Cut, Broke bool
}

// RunLedger is RunPath with the accounts kept: the same path, bit for bit, and
// one LedgerYear per year the household lived. It is the reading of ONE
// simulated future, the drill-down beneath every percentile on a fan, and it
// always runs the annual kernel (a ledger is a yearly reading), so a Monthly
// plan's ledger is the one its annual twin would keep.
func (p Plan) RunLedger(returns scenario.Sequence, lives Lives) (PathResult, []LedgerYear) {
	led := make([]LedgerYear, 0, p.Years)
	res := p.runPathAnnual(returns, lives, nil, &led)
	return res, led
}

// DrawPath draws path i of the ensemble Draw(n, workers, seed) would draw, for
// any n above i, and nothing else: every path has its own RNG stream of the
// seed (see Simulate), so one future of a ten-thousand-path run can be
// replayed without the other 9 999. Lives is the zero value when the plan has
// no Lifetime.
func (p Plan) DrawPath(seed uint64, i int) (scenario.Sequence, Lives) {
	pcg := new(rand.PCG)
	seedPath(pcg, seed, i, 0)
	returns := scenario.Prepare(p.Source).Draw(rand.New(pcg))
	var lives Lives
	if p.Lifetime != nil {
		seedPath(pcg, seed, i, lifeStream)
		lives = p.Lifetime.sampler(p.Years).draw(rand.New(pcg))
	}
	return returns, lives
}
