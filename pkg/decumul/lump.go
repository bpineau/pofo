package decumul

// Lump is a one-off real capital event in a plan year, the counterpart of a
// Cashflow for money that arrives or leaves ONCE: an inheritance or the sale
// of a house comes in, a purchase or a gift to the children goes out.
//
// It is applied at the start of its year, before that year's budget is set.
// A positive Amount is invested into the first pocket of the growth sleeve
// (the taxable account, where a windfall lands) at a cost basis equal to
// itself, so it carries no gain on arrival. A negative one is sold out of the
// growth sleeve, taxed like any sale and grossed up, then out of the cash
// buffer; a lump the portfolio cannot pay in full is that year's ruin. A lump
// due after the money has run out never arrives in the simulation, so it
// cannot undo a ruin already counted.
//
// A large inflow must be a Lump, not a one-year Cashflow: income above the
// year's need is received but never reinvested (the kernel has nowhere to put
// it), so a 300 000 inheritance entered as income would mostly vanish. The
// spending rules that read total wealth (ABW, the risk-based guardrail) count
// the lumps still to come in their forecast, like the cashflows.
type Lump struct {
	Year   int     // plan year it happens, 0-based, at the start of the year
	Amount float64 // real amount in the units of Plan.Capital: positive in, negative out
}

// lumpAt is the net of the plan's lumps in a year.
func (p *Plan) lumpAt(year int) float64 {
	sum := 0.0
	for _, l := range p.Lumps {
		if l.Year == year {
			sum += l.Amount
		}
	}
	return sum
}

// applyLumps books the year's lumps into the pockets and the buffer, latching
// ruin on a payment the portfolio cannot make, and returns the new buffer.
func (p *Plan) applyLumps(k int, pks pocketOps, buffer float64, res *PathResult) float64 {
	for _, l := range p.Lumps {
		switch {
		case l.Year != k || l.Amount == 0:
		case l.Amount > 0:
			pks.deposit(l.Amount)
		default:
			want := -l.Amount
			got := pks.sell(want, &res.TaxPaid)
			take := min(want-got, buffer)
			buffer -= take
			if got+take < want-1e-6 {
				res.ruinAt(k)
			}
		}
	}
	return pks.settle(buffer)
}
