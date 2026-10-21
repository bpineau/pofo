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
// the worker striding or to the number or order of the RNG calls moves them.
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
		d.fs(s.Wealth).b(s.Ruined)
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
		{Name: "AV", Amount: 250_000, GainFrac: 0.3, Tax: AVTax{Rate: 0.247, Allowance: 9200}},
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
		"fixed":             "a99e488631f6bd8cbd193adf74a985b837c564a4cd1fcbfaaa4917bb4c854004",
		"ratchet":           "17fa08b47a92680085a14ffdb86509e9649cd9e9d424944a53091f5f2ef6ff44",
		"guardrails":        "b18b6485bcb987ec89d438c6716c20a03bed14f51e931898b7ad17cf2ce22abf",
		"riskguard":         "0506c8115cdff11650215ab7ba5f64b56f1ebbf55cd386cc251bd6ccd779addf",
		"percent":           "aa933ea317460f57bc84dd08c12c94e8eaf5b2e08a96ea6baf77bb941d2f9b49",
		"bounded":           "1f033fc29fb82257e34f4531788ed48ae2ed1f2eee6ef19fcadfb562f28a4cdf",
		"amortize":          "dfce07fe6aa20ca655172d5d236e594a26d047ec85652ebc0f11fc3f612fb65c",
		"envelopes":         "da5f6dd072c369f54e8ba71b8abbcb27c18d6755981b8baf382b891fd00676a8",
		"monthly":           "d8845288131956eef75525a79cc847320687c60dc98d129ebd084e9718e891c5",
		"monthly-riskguard": "d6d05fc1e7a892c7d7a2ad95af0b8eb48a842bca9ebc3ae0a5226fc0a928a3aa",
		"schedule":          "2583093ef4faa8cd92705090a2ce10ce7d6db87619e81cdae0fa9633306f2888",
		"lifetime-single":   "b7cf0ff71b94c4669657ed1ef7d8a321ad46b3d6305e1867dfd7b08dbf4806c1",
		"lifetime-couple":   "9d9f17a818fc10f6aff2be73a21f70bfc534920e779eddce48f801176b37fbcb",
		"lifetime-annuity":  "0ee6cc4151e6cee090653f8f7ec672d48186ccd90c886ef738116d23b421db73",
		"lifetime-amortize": "ea574821017cf11894ef001f53f753e5c70cba2f43b16c57a45aedab596f25c1",
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
		"parametric":           "e8e4815429213fea438112d8f9d71ee444861abba30d6e9e15a56237ffe357e4",
		"parametric-normal":    "bf81a5720bf85a522169083ecb2d2a8008b8faa1698e93e29f9e4bbc584abf6d",
		"markov-regime":        "e519a4b4d22595d8c08795cf700f39b712cd9817c0a1adc7d8078a751e223b21",
		"lost-decade":          "fd90d9fb79fad8285fd4df5cfdcf0357feda258c418bfb9408eaf7a699035fcb",
		"glidepath":            "09e06b62bde36fed8183b7439c0efbf2fd1821f954d259716c40e9d9d9b3ed40",
		"block-bootstrap":      "8ce450542509b66bd3804a12a0e33ce1419494b55a6cd84cc89e5c967e84f69e",
		"stationary-bootstrap": "72390f678551144c9b1186d74b183cf655460c50163d9aff49eff33a8eed3837",
		"historical-cohorts":   "8d3bbbcb9ab76c5a8c0046ff45d67200e32c3b65e21848ac094f424ae047ce7b",
		"pooled-bootstrap":     "77cb487d3d3179164a308a3bd3f0467c3452225c90f84358dd89f63bad652303",
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
