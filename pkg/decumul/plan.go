package decumul

import "github.com/bpineau/pofo/pkg/scenario"

// Cashflow is a real annual income (e.g. a pension) received from FromYear
// (0-based) to the horizon; it reduces the net amount sold from the
// portfolio. Pensions are modelled as cashflows, not as an asset.
type Cashflow struct {
	FromYear int
	Annual   float64
}

// BufferSleeve is a low-volatility cash or inflation-linked pocket sized at
// Years times annual spending (capped at the capital). It earns RealReturn
// and is drained first while the portfolio drawdown exceeds DrawThreshold;
// otherwise it is refilled from growth, by at most RefillCap of growth/year.
type BufferSleeve struct {
	Years         float64
	RealReturn    float64
	DrawThreshold float64 // default 0.10
	RefillCap     float64 // default 0.50
}

// FlexRule cuts the year's spending by Cut (e.g. 0.25) whenever the
// portfolio drawdown exceeds Threshold (e.g. 0.20). A zero rule is inactive.
type FlexRule struct {
	Threshold, Cut float64
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
}

// needAt is the net spending in a given year after active cashflows,
// floored at 0.
func (p Plan) needAt(year int) float64 {
	need := p.NeedAnnual
	for _, c := range p.Cashflows {
		if year >= c.FromYear {
			need -= c.Annual
		}
	}
	if need < 0 {
		return 0
	}
	return need
}
