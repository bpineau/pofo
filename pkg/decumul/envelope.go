package decumul

import "math"

// Envelope is one tax wrapper (a French CTO, PEA or assurance-vie) holding a
// slice of the growth sleeve. Amounts are relative: the growth part of the
// capital (after the buffer is carved out) is split across envelopes pro-rata,
// so a capital sweep or solve scales every pocket together while the web layer
// can pass the actual absolute amounts. GainFrac is the embedded unrealised
// gain fraction at the start (0 = cost basis equals value, 1 = pure gain);
// it drives how much of every sale is taxable from day one.
//
// Withdrawals drain envelopes in slice order (the first is sold first), the
// classic French sequencing being CTO first, then PEA, then assurance-vie.
// A nil Plan.Envelopes keeps the historical single sleeve taxed by Plan.Tax.
type Envelope struct {
	Name     string
	Amount   float64 // relative size of the pocket (pro-rata of the growth sleeve)
	GainFrac float64 // unrealised gain fraction at start, in [0, 1]
	Tax      Tax
}

// YearlyTax is a Tax whose liability carries per-year, per-path state, such as
// the assurance-vie annual allowance. The kernel calls NewPath once per
// simulated path (isolating state across paths) and NewYear at each year
// boundary.
type YearlyTax interface {
	Tax
	NewPath() YearlyTax
	NewYear()
}

// AVTax models the French assurance-vie taxation after 8 years: each year the
// first Allowance euros of realised gain are tax-free (4 600 € per person,
// 9 200 € for a couple), and the excess gain is taxed at Rate (7.5% + 17.2%
// social levies ≈ 24.7% below the 150 k€ premium threshold). The gain share of
// a sale follows the same cost-basis pro-rata as CTOFlatTax.
//
// AVTax is a stateless template: the kernel derives per-path state through
// NewPath, so a Plan can be simulated concurrently.
type AVTax struct {
	Rate      float64
	Allowance float64
}

// GrossUp implements Tax on the template with a zero used-allowance; the
// kernel normally calls it on the per-path state from NewPath.
func (t AVTax) GrossUp(net, growth, cost float64) (gross, newCost, taxPaid float64) {
	s := t.NewPath()
	return s.GrossUp(net, growth, cost)
}

// NewPath returns fresh per-path state.
func (t AVTax) NewPath() YearlyTax { return &avTaxState{AVTax: t} }

// NewYear on the template is a no-op (no state).
func (t AVTax) NewYear() {}

// avTaxState is AVTax plus the realised gain already counted against the
// current year's allowance.
type avTaxState struct {
	AVTax
	used float64
}

// NewPath resets the state (a fresh path starts with a full allowance).
func (t *avTaxState) NewPath() YearlyTax { return &avTaxState{AVTax: t.AVTax} }

// NewYear restores the annual allowance.
func (t *avTaxState) NewYear() { t.used = 0 }

// GrossUp implements Tax: the sale's realised gain first consumes what is left
// of the allowance, and only the excess is taxed at Rate. Like CTOFlatTax, the
// sale is capped at the available market value.
func (t *avTaxState) GrossUp(net, growth, cost float64) (gross, newCost, taxPaid float64) {
	if growth <= 0 {
		return net, cost, 0
	}
	gainFrac := 1 - cost/growth
	if gainFrac < 0 {
		gainFrac = 0
	}
	left := t.Allowance - t.used
	if left < 0 {
		left = 0
	}
	gross = net
	if gainFrac > 0 && net*gainFrac > left {
		// Solve gross = net + Rate*(gross*gainFrac - left) for gross.
		if d := 1 - t.Rate*gainFrac; d > 0 {
			gross = (net - t.Rate*left) / d
		}
	}
	if gross > growth {
		gross = growth // sale capped at the available market value
	}
	gain := gross * gainFrac
	if taxable := gain - left; taxable > 0 {
		taxPaid = t.Rate * taxable
	}
	t.used += gain
	newCost = cost * (1 - gross/growth)
	return gross, newCost, taxPaid
}

// pocket is the per-path state of one envelope: its market value, cost basis
// and (possibly per-path stateful) tax.
type pocket struct {
	value, cost float64
	tax         Tax
}

// newPockets carves the growth sleeve into per-path tax pockets. Without
// envelopes it is the historical single sleeve on Plan.Tax with a cost basis
// equal to the invested amount; with envelopes each pocket takes its pro-rata
// share of growth, its GainFrac-implied cost basis, and a per-path clone of
// any stateful tax.
func (p Plan) newPockets(growth float64) []pocket {
	defaulted := func(t Tax) Tax {
		if t == nil {
			return CTOFlatTax{}
		}
		if yt, ok := t.(YearlyTax); ok {
			return yt.NewPath()
		}
		return t
	}
	if len(p.Envelopes) == 0 {
		return []pocket{{value: growth, cost: growth, tax: defaulted(p.Tax)}}
	}
	sum := 0.0
	for _, e := range p.Envelopes {
		sum += e.Amount
	}
	out := make([]pocket, len(p.Envelopes))
	for i, e := range p.Envelopes {
		v := growth
		if sum > 0 {
			v = growth * e.Amount / sum
		}
		g := min(max(e.GainFrac, 0), 1)
		out[i] = pocket{value: v, cost: v * (1 - g), tax: defaulted(e.Tax)}
	}
	return out
}

// pocketOps bundles the operations both kernels need over a pocket set.
type pocketOps []pocket

// total is the growth sleeve's market value.
func (ps pocketOps) total() float64 {
	s := 0.0
	for _, pk := range ps {
		s += pk.value
	}
	return s
}

// sell delivers up to want net euros by draining pockets in order, updating
// values and cost bases and accruing tax into taxPaid. It returns the net
// actually delivered (below want when every pocket is exhausted).
func (ps pocketOps) sell(want float64, taxPaid *float64) float64 {
	delivered := 0.0
	for i := range ps {
		if want-delivered <= 1e-9 {
			break
		}
		pk := &ps[i]
		if pk.value <= 0 {
			continue
		}
		gross, nc, paid := pk.tax.GrossUp(want-delivered, pk.value, pk.cost)
		pk.value -= gross
		pk.cost = nc
		*taxPaid += paid
		delivered += gross - paid
	}
	return delivered
}

// grow applies one periodic return to every pocket.
func (ps pocketOps) grow(r float64) {
	for i := range ps {
		ps[i].value *= 1 + r
	}
}

// settle folds any negative pocket (a stub tax without a cap may oversell)
// into the buffer, returning the adjusted buffer.
func (ps pocketOps) settle(buffer float64) float64 {
	for i := range ps {
		if ps[i].value < 0 {
			buffer += ps[i].value
			ps[i].value = 0
		}
	}
	return buffer
}

// liquidationNet previews the net proceeds of selling every pocket today,
// without touching the real state: the amortization rule amortizes THIS
// value, not the gross market value, because the gross is not deliverable
// (the final actuarial payment would need a tax gross-up beyond what
// exists). Stateful taxes are previewed on a fresh per-path clone; the
// annual kernel prices needs before any sale of the year, so the fresh
// allowance matches the real state.
func (ps pocketOps) liquidationNet() float64 {
	cp := make(pocketOps, len(ps))
	for i, pk := range ps {
		t := pk.tax
		if yt, ok := t.(YearlyTax); ok {
			t = yt.NewPath()
		}
		cp[i] = pocket{value: pk.value, cost: pk.cost, tax: t}
	}
	var tax float64
	return cp.sell(math.MaxFloat64/4, &tax)
}

// newYear signals a year boundary to every stateful tax.
func (ps pocketOps) newYear() {
	for _, pk := range ps {
		if yt, ok := pk.tax.(YearlyTax); ok {
			yt.NewYear()
		}
	}
}
