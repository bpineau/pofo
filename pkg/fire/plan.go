package fire

import (
	"fmt"
	"math"

	"github.com/bpineau/pofo/pkg/decumul"
)

const (
	// FixedReal keeps the requested real household spending level constant.
	FixedReal = "fixed_real"
	// GuytonKlinger adjusts spending by ten percent outside initial-rate bands.
	GuytonKlinger = "guyton_klinger"
	// TaxAssumptionLabel identifies the default planning rate's scope.
	TaxAssumptionLabel = "French taxable-account assumption (2026), verify or override"
	// TaxAssumptionYear dates the current default tax planning assumption.
	TaxAssumptionYear = 2026
	// BufferDrawThreshold is the default wealth drawdown fraction tapping cash.
	BufferDrawThreshold = decumul.DefaultBufferDrawThreshold
	// BufferRefillCap is the default annual refill fraction of growth wealth.
	BufferRefillCap = decumul.DefaultBufferRefillCap
	// GuardLowerMultiplier defines the lower band relative to initial spending/capital.
	GuardLowerMultiplier = 0.8
	// GuardUpperMultiplier defines the upper band relative to initial spending/capital.
	GuardUpperMultiplier = 1.2
	// GuardCut is the proportional spending reduction on an upper-band breach.
	GuardCut = 0.10
	// GuardRaise is the proportional spending increase on a lower-band breach.
	GuardRaise = 0.10
)

// Pension is one real annual household income, paying in StartYear (zero
// based) and compounding by RealDrift thereafter. All rates are fractions.
type Pension struct {
	AnnualAmount float64 // first annual payment, real money units
	StartYear    int     // first payment's zero-based plan year
	RealDrift    float64 // annual real growth fraction after the first payment
}

// PlanSpec builds the shared fixed-horizon retirement kernel. Money is real
// purchasing power at plan start; rates are fractions. Explicit zeros are
// preserved: start from Defaults when the shared assumptions are desired.
// Capital includes the cash buffer; AnnualSpending includes pension-funded needs.
type PlanSpec struct {
	Capital          float64  // initial wealth including the buffer, real money units
	AnnualSpending   float64  // annual household budget before pension, real money units
	Years            int      // fixed horizon in years
	Policy           string   // FixedReal or GuytonKlinger; empty means FixedReal
	BufferYears      float64  // buffer size in annual spending units
	BufferRealReturn float64  // annual real buffer return fraction
	TaxRate          float64  // fraction of realized nominal gains paid as tax
	InitialGain      float64  // fraction of initial growth wealth representing gains
	Inflation        float64  // annual inflation fraction eroding the nominal basis
	Pension          *Pension // nil means no outside income
}

// Defaults returns current planning assumptions, not legal rates or a frozen
// API contract. Capital, spending, horizon and pension are left for the caller.
// The tax assumption's scope is TaxAssumptionLabel. Values may evolve with data.
func Defaults() PlanSpec {
	return PlanSpec{Policy: FixedReal, BufferYears: 3, BufferRealReturn: .005, TaxRate: .328, InitialGain: .5, Inflation: .02}
}

// BuildPlan maps a specification onto the annual withdrawal kernel. Source is
// left unset for the caller. It rejects unknown policies and non-finite or
// negative money, horizon, buffer, tax and gain inputs; return and inflation
// rates must exceed -1, and tax must be below 1. Zero capital/horizon may be
// used to build an incomplete template but must be filled before simulation.
// The buffer is capped by the kernel at min(BufferYears*AnnualSpending,Capital).
// Guyton-Klinger bands are rebuilt against this candidate's capital. This does
// not establish monotonicity: adaptive capital bisection is unsupported.
func BuildPlan(s PlanSpec) (decumul.Plan, error) {
	for _, v := range []float64{s.Capital, s.AnnualSpending, s.BufferYears, s.TaxRate, s.InitialGain} {
		if !finite(v) || v < 0 {
			return decumul.Plan{}, fmt.Errorf("fire: plan amounts and rates must be finite and non-negative")
		}
	}
	if s.Years < 0 || s.TaxRate >= 1 || s.InitialGain > 1 || !finite(s.BufferRealReturn) || s.BufferRealReturn <= -1 || !finite(s.Inflation) || s.Inflation <= -1 {
		return decumul.Plan{}, fmt.Errorf("fire: invalid horizon or rate")
	}
	if s.Policy == "" {
		s.Policy = FixedReal
	}
	if s.Policy != FixedReal && s.Policy != GuytonKlinger {
		return decumul.Plan{}, fmt.Errorf("fire: unsupported policy %q", s.Policy)
	}
	p := decumul.Plan{Capital: s.Capital, NeedAnnual: s.AnnualSpending, Years: s.Years, Buffer: decumul.BufferSleeve{Years: s.BufferYears, RealReturn: s.BufferRealReturn}, Tax: decumul.CTOFlatTax{Rate: s.TaxRate}, Inflation: s.Inflation}
	if s.Pension != nil {
		c := s.Pension
		if !finite(c.AnnualAmount) || c.AnnualAmount < 0 || c.StartYear < 0 || !finite(c.RealDrift) || c.RealDrift <= -1 {
			return decumul.Plan{}, fmt.Errorf("fire: invalid pension")
		}
		if c.AnnualAmount > 0 {
			p.Cashflows = []decumul.Cashflow{{FromYear: c.StartYear, Annual: c.AnnualAmount, Drift: c.RealDrift}}
		}
	}
	if s.InitialGain > 0 {
		growth := math.Max(0, s.Capital-math.Min(s.BufferYears*s.AnnualSpending, s.Capital))
		p.Envelopes = []decumul.Envelope{{Name: "CTO", Amount: growth, GainFrac: s.InitialGain, Tax: decumul.CTOFlatTax{Rate: s.TaxRate}}}
	}
	if s.Policy == GuytonKlinger && s.Capital > 0 {
		rate := s.AnnualSpending / s.Capital
		p.Guard = decumul.Guardrails{Upper: rate * GuardUpperMultiplier, Lower: rate * GuardLowerMultiplier, Cut: GuardCut, Raise: GuardRaise}
	}
	return p, nil
}

func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
