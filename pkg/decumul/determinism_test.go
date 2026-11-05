package decumul

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"hash"
	"math"
	"math/rand/v2"
	"testing"

	"github.com/bpineau/pofo/pkg/scenario"
)

// Determinism is a feature of this engine, not a happy accident: a shared
// /view or -fire URL must reproduce byte for byte, and the solvers rely on
// common random numbers to stay free of Monte-Carlo noise. The digests below
// therefore pin the EXACT bits of everything the package produces (per-path
// series included) for a representative matrix of plans and of every
// scenario.Source kind.
//
// They are a guard against silent drift, so they carry no tolerance: any
// change to the arithmetic, to the order of the floating-point summations, to
// the per-path seeding or to the number or order of the RNG calls moves them.
// The worker count must NOT move them (TestResultsIndependentOfWorkers).
// An optimisation that moves a digest is not an optimisation, it is a
// different model; revert it. A DELIBERATE model change updates the literals
// here in the same commit that justifies it.

// digest accumulates the bits of a result into a hash.
type digest struct{ h hash.Hash }

func newDigest() *digest { return &digest{h: sha256.New()} }

// f mixes one float64, by its bits, so a NaN or a signed zero still counts.
func (d *digest) f(v float64) *digest {
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], math.Float64bits(v))
	d.h.Write(b[:])
	return d
}

// i mixes one integer.
func (d *digest) i(v int) *digest {
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], uint64(v))
	d.h.Write(b[:])
	return d
}

// b mixes one boolean.
func (d *digest) b(v bool) *digest {
	if v {
		return d.i(1)
	}
	return d.i(0)
}

// fs mixes a float slice, length included.
func (d *digest) fs(vs []float64) *digest {
	d.i(len(vs))
	for _, v := range vs {
		d.f(v)
	}
	return d
}

// sum renders the accumulated digest.
func (d *digest) sum() string { return hex.EncodeToString(d.h.Sum(nil)) }

// path mixes every field of one simulated path, series included.
func (d *digest) path(p PathResult) *digest {
	return d.fs(p.Wealth).fs(p.Spend).b(p.Ruined).i(p.RuinYear).i(p.FirstCut).i(p.CutYears).
		f(p.TaxPaid).f(p.Withdrawn).f(p.Ret10).i(p.LifeYears).b(p.Outlived).
		f(p.Estate).f(p.Annuity).f(p.Premium).f(p.Received)
}

// ensemble mixes a whole ensemble and every aggregate built on it.
func (d *digest) ensemble(e Ensemble) *digest {
	d.i(e.Years).i(len(e.Paths))
	for _, p := range e.Paths {
		d.path(p)
	}
	o := e.Outcome()
	d.f(o.RuinProb).f(o.TerminalP5).f(o.TerminalP50).f(o.MedianYearsUnderwater).
		f(o.Worst10yCAGR).f(o.Worst10yP5).f(o.CDaR).f(o.MedianCumTax).f(o.EffectiveTaxRate)
	lo := e.LifeOutcome()
	d.f(lo.RuinAlive).f(lo.OutlivedPlan).f(lo.MedianLifeYears).f(lo.BrokeYearsMean).
		f(lo.BrokeYearsP95).f(lo.EstateP5).f(lo.EstateP50).f(lo.EstateP90).
		f(lo.EstateZero).f(lo.IncomeMean)
	d.fs(e.Estates())
	fan := e.Fan([]float64{0.05, 0.25, 0.50, 0.75, 0.95}, 5)
	d.i(fan.Years).fs(fan.Pcts)
	for _, band := range fan.Bands {
		d.fs(band)
	}
	for _, s := range fan.Samples {
		d.fs(s.Wealth).b(s.Ruined).i(s.Index)
	}
	for _, pt := range e.LifeStates() {
		d.f(pt.Dead).f(pt.Broke).f(pt.Funded)
	}
	for _, pt := range e.LifeCurve(Life{Age: 52}.Survival) {
		d.f(pt.Dead).f(pt.Broke).f(pt.Funded)
	}
	d.f(e.RuinProb()).f(e.SpendCV())
	for _, band := range e.SpendBands([]float64{0.10, 0.50, 0.90}) {
		d.fs(band)
	}
	return d
}

// determinismPanel is a portfolio-shaped panel of monthly real returns, the
// input the data-driven sources resample. It is generated, not fetched, so the
// digests never depend on the bundled datasets.
func determinismPanel() scenario.Panel {
	rows := make([][]float64, 4)
	for a := range rows {
		rows[a] = make([]float64, 480)
		for t := range rows[a] {
			rows[a][t] = 0.003 + 0.03*math.Sin(float64(t*(a+2))/7.0)
		}
	}
	return scenario.Panel{Returns: rows, Weights: []float64{0.4, 0.3, 0.2, 0.1}}
}

// determinismPlan is the shared skeleton: a buffered, taxed, pensioned plan of
// the shape the FIRE page simulates.
func determinismPlan(years int) Plan {
	return Plan{
		Capital: 1_000_000, NeedAnnual: 40_000, Years: years,
		Tax:       CTOFlatTax{Rate: 0.30},
		Buffer:    BufferSleeve{Years: 2, RealReturn: -0.01},
		Flex:      FlexRule{Threshold: 0.20, Cut: 0.10},
		Cashflows: []Cashflow{{FromYear: 12, Annual: 14_000}},
		Source:    scenario.ParametricSource{Mu: 0.045, Sigma: 0.13, Df: 5, Periods: years},
	}
}

// determinismPlans is the plan matrix: every spending rule, the envelope
// split, the monthly kernel, and the lifetime kernel with and without an
// annuity and a partner.
func determinismPlans() []struct {
	name string
	plan Plan
} {
	const years = 30
	fixed := determinismPlan(years)

	flex := fixed
	flex.Ratchet = Ratchet{Trigger: 1.2, Step: 2000, Cap: 60000, Cooldown: 3, MaxWR: 0.045}

	guard := fixed
	guard.Guard = Guardrails{Upper: 0.055, Lower: 0.035, Cut: 0.10, Raise: 0.10, Floor: 30000}

	safe := make([]float64, years)
	for i := range safe {
		safe[i] = 0.035 + 0.002*float64(i)
	}
	risk := fixed
	risk.RiskGuard = RiskGuardrails{SafeWR: safe, Band: 0.20, Cut: 0.10, Raise: 0.10,
		Floor: 30000, Cap: 55000, PVRate: 0.03}

	pct := fixed
	pct.Percent = 0.045

	bounded := fixed
	bounded.Bounded = BoundedPct{Pct: 0.045, Up: 0.05, Down: 0.025}

	amort := fixed
	amort.Amortize, amort.AmortReturn = true, 0.03

	envelopes := fixed
	envelopes.Envelopes = []Envelope{
		{Name: "CTO", Amount: 500_000, GainFrac: 0.4, Tax: CTOFlatTax{Rate: 0.314}},
		{Name: "PEA", Amount: 250_000, GainFrac: 0.5, Tax: CTOFlatTax{Rate: 0.186}},
		{Name: "AV", Amount: 250_000, GainFrac: 0.3, Tax: AVTax{Rate: 0.075, Levies: 0.172, Allowance: 9200}},
	}

	monthly := fixed
	monthly.Monthly = true
	monthly.Source = scenario.ParametricSource{Mu: 0.045 / 12, Sigma: 0.13 / math.Sqrt(12), Df: 5, Periods: years * 12}

	monthlyRisk := monthly
	monthlyRisk.RiskGuard = risk.RiskGuard

	sched := fixed
	sched.SpendSchedule = make([]float64, years)
	for i := range sched.SpendSchedule {
		sched.SpendSchedule[i] = 1 + 0.005*float64(i)
	}

	// The lifetime plans run a horizon past any plausible age, as the design
	// requires, and plan over the shorter PlanHorizon.
	const longYears = 55
	single := determinismPlan(longYears)
	single.Years, single.PlanHorizon = longYears, 40
	single.Source = scenario.ParametricSource{Mu: 0.045, Sigma: 0.13, Df: 5, Periods: longYears}
	single.Lifetime = &Lifetime{Self: Life{Age: 52}}

	partner := single
	partner.Lifetime = &Lifetime{Self: Life{Age: 52}, Partner: &Life{Age: 50}, SurvivorSpend: 0.7}
	partner.Cashflows = []Cashflow{{FromYear: 12, Annual: 14_000, Owner: Self, Reversion: 0.54}}

	annuity := partner
	annuity.Annuity = &Annuity{Share: 0.25, Year: 15, Load: 0.10}

	lifeAmort := single
	lifeAmort.Amortize, lifeAmort.AmortReturn = true, 0.03

	return []struct {
		name string
		plan Plan
	}{
		{"fixed", fixed},
		{"ratchet", flex},
		{"guardrails", guard},
		{"riskguard", risk},
		{"percent", pct},
		{"bounded", bounded},
		{"amortize", amort},
		{"envelopes", envelopes},
		{"monthly", monthly},
		{"monthly-riskguard", monthlyRisk},
		{"schedule", sched},
		{"lifetime-single", single},
		{"lifetime-couple", partner},
		{"lifetime-annuity", annuity},
		{"lifetime-amortize", lifeAmort},
	}
}

// TestDeterminismPlans pins the bits of a full run of every spending rule,
// every envelope arrangement and the lifetime kernel.
func TestDeterminismPlans(t *testing.T) {
	want := map[string]string{
		"fixed":             "69256edbb1056c04b5fa6b84dc8051dd5ee5429016fb430779d79d66aefd61f7",
		"ratchet":           "acd7e52953faa17a47a7e34debaa8bc60d019beebeb81f5fc5427355fa004106",
		"guardrails":        "0f7defc3c6aeaa57265af3a2f734935a2c8dda8eb55b36b03c5d6c87e9d33073",
		"riskguard":         "1a6cf017b600f29abfca2946405be1d448ee960f1ec361b1a9daeaea390dd141",
		"percent":           "de0b4831a3dcfccf2318d105b6fda16101d526d2c8fcb01be6aec2f8ada946b7",
		"bounded":           "34b5e30f6628f7c984346bd57e9734d2efd7c32f664906cb72322fd68b068c44",
		"amortize":          "91034a9c6b6e64f32e199818231339319d097e23d0cdf6c2015a4db36e3cc44a",
		"envelopes":         "720e17528e8d54cb86969c9e1a02205392efb2490d54bdfffdb6a8b545d11af7",
		"monthly":           "50a8fdaabb8a7d01de431ea8fe95836878dcedf5f804d299511647acf48a5489",
		"monthly-riskguard": "d9ab5a63fdf0cb5f4db11171c1dfb7399dc086fba1fade62cfa772bf0c6db24d",
		"schedule":          "69926640b9433dccf0303e63ee520c3e919c2b7fbbcfc2c77c4ddc8dcc90a6da",
		"lifetime-single":   "d7de2389d542ddb2662ebf1ae8e83394d0f4d936b152e0ad4c639aea63a84a26",
		"lifetime-couple":   "8caac299bff265dec174ec09c7b40a5de7a2420fa606f4b42800da5aca7052f7",
		"lifetime-annuity":  "1dbada2b29714d5429e7ec8f07d97aa8c19a5a04b9160fbcf4a02584c1c87045",
		"lifetime-amortize": "36bda999b1b35a153524e292ab0f88f1006aa03be87dc1c879c6eba06247ae08",
	}
	for _, c := range determinismPlans() {
		d := newDigest()
		d.ensemble(c.plan.Simulate(500, 4, 7))
		// A second run at a different worker count: the kernel is deterministic,
		// so only the draws are tied to the striding and the digest must match
		// the single-worker one below through the same Draws.
		draws := c.plan.Draw(500, 4, 7)
		d.ensemble(c.plan.SimulateOn(draws, 1))
		d.f(c.plan.Solve(0.05, WithdrawalAxis(0, 150_000), 400, 4, 7))
		d.f(c.plan.CapitalForRuin(0.05, 0.5e6, 4e6, 400, 4, 7))
		pts, err := c.plan.Sweep1D(BufferYears, []float64{0, 1, 2, 3, 5}, 400, 4, 7)
		if err != nil {
			t.Fatalf("%s: sweep: %v", c.name, err)
		}
		for _, pt := range pts {
			d.f(pt.Value).f(pt.RuinProb).f(pt.TerminalP50)
		}
		if got := d.sum(); got != want[c.name] {
			t.Errorf("%s digest = %q, want %q", c.name, got, want[c.name])
		}
	}
}

// determinismSources is one instance of every Source kind, at the shapes the
// FIRE page uses them.
func determinismSources() []struct {
	name string
	src  scenario.Source
} {
	panel := determinismPanel()
	return []struct {
		name string
		src  scenario.Source
	}{
		{"parametric", scenario.ParametricSource{Mu: 0.045, Sigma: 0.13, Df: 5, Periods: 30}},
		{"parametric-normal", scenario.ParametricSource{Mu: 0.045, Sigma: 0.13, Df: 2, Periods: 30}},
		{"markov-regime", scenario.NewMarkovRegime(0.045, 0.13, 5, 30)},
		{"lost-decade", scenario.NewLostDecadeRegime(0.045, 0.13, 5, 30)},
		{"glidepath", scenario.Glidepath{EquityMu: 0.05, EquitySigma: 0.17, BondMu: 0.01,
			BondSigma: 0.06, Df: 5, Corr: 0.2, StartEquity: 0.4, EndEquity: 0.8, Periods: 30}},
		{"block-bootstrap", scenario.Compounded{
			Inner: scenario.BlockBootstrap{Panel: panel, BlockLen: 24, Periods: 360}, Group: 12}},
		{"stationary-bootstrap", scenario.Compounded{
			Inner: scenario.StationaryBootstrap{Panel: panel, MeanBlock: 24, Periods: 360}, Group: 12}},
		{"historical-cohorts", scenario.Compounded{
			Inner: scenario.HistoricalCohorts{Panel: panel, Periods: 360}, Group: 12}},
		{"pooled-bootstrap", scenario.PooledBootstrap{
			Series: panel.Returns, MeanBlock: 24, Periods: 30}},
	}
}

// TestDeterminismSources pins the bits every Source kind draws, raw and
// through scenario.Prepare (which must be byte-identical), and the bits of a
// full plan run on each of them.
func TestDeterminismSources(t *testing.T) {
	want := map[string]string{
		"parametric":           "314f2e3f4bfc353c49d813f9a2889902cd1b9f45ae044694613c10b1745a1f59",
		"parametric-normal":    "dfed3be7ae653138095df1de51acba990194c978ed397f02c868daf036a424d9",
		"markov-regime":        "36a4898194d40dab5798e2644d7932c829ea815e995f193fbc6f86b95e5eb399",
		"lost-decade":          "44cdc2d631cc8b9e4231f9fc153c8975da5ff55d3af63e78da8474d5c8715765",
		"glidepath":            "a9a38ddbca1a80b30aa56d9f607d481d3213e3d757f19adae14f01845f4db369",
		"block-bootstrap":      "77652b7162cd09d1acde72aa6fa86e86a309f40f0086b2bb6aa420cc3464b13b",
		"stationary-bootstrap": "3668128bc14b7ae2189086302e0a9dfc89e72b5c5589dc7e7135c4c8e9ee8694",
		"historical-cohorts":   "efba101dbadc7611bb1093fab34c7911f88f19b29291a66ae80234399926fef6",
		"pooled-bootstrap":     "42ad8a25aac7d2337b82f152403991919e937e429958347ced5279acd5b7c416",
	}
	for _, c := range determinismSources() {
		d := newDigest()
		rng := rand.New(rand.NewPCG(7, 1))
		for range 50 {
			d.fs(c.src.Draw(rng))
		}
		prepared := scenario.Prepare(c.src)
		rng = rand.New(rand.NewPCG(7, 1))
		for range 50 {
			d.fs(prepared.Draw(rng))
		}
		p := determinismPlan(c.src.Len())
		p.Source = c.src
		d.ensemble(p.Simulate(400, 4, 11))
		if got := d.sum(); got != want[c.name] {
			t.Errorf("%s digest = %q, want %q", c.name, got, want[c.name])
		}
	}
}

// TestResultsIndependentOfWorkers: the worker count is a scheduling choice and
// must never reach a result. Every path draws from its OWN stream, derived
// from the seed and its index, so a FIRE URL rendered on a two-core server and
// on a ten-core laptop gives the same bits: the draws, every per-path series
// and every aggregate, for every plan and source kind of the digests above.
func TestResultsIndependentOfWorkers(t *testing.T) {
	type run struct {
		name string
		plan Plan
	}
	var runs []run
	for _, c := range determinismPlans() {
		runs = append(runs, run{c.name, c.plan})
	}
	for _, c := range determinismSources() {
		p := determinismPlan(c.src.Len())
		p.Source = c.src
		runs = append(runs, run{"source " + c.name, p})
	}
	for _, r := range runs {
		var ref string
		for _, workers := range []int{1, 3, 8} {
			d := newDigest()
			d.ensemble(r.plan.Simulate(300, workers, 7))
			d.f(r.plan.Solve(0.05, WithdrawalAxis(0, 150_000), 300, workers, 7))
			got := d.sum()
			if workers == 1 {
				ref = got
				continue
			}
			if got != ref {
				t.Errorf("%s: %d workers give %s, 1 worker %s", r.name, workers, got, ref)
			}
		}
	}
}
