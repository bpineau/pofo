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
		"fixed":             "c492941f124a478e9488c164305242b88b814af6137c3e07c24e83466dd3e94f",
		"ratchet":           "f03dd8b5e9e32174634ac65c4e0efc65a2984c7f2c5b1ecd13b104c83eb09c42",
		"guardrails":        "8721bd8d89408fb872fe49c0156b962fdde6614d81a1f9e201b0b77f58651271",
		"riskguard":         "ae59b85955431ac5e96d4c27a9c666130c50c6d47609fd0679c2162a4637b2e7",
		"percent":           "0fb11c7724fe5178e10f5dd83dc85bdfd4da4777aa121d6c7fbcdd6c671cd990",
		"bounded":           "b9656d0ad795ac47778e3636f0ccdb0d186db9f59923f65956cddc9fd313151a",
		"amortize":          "d9b268b1ea73b66c8aa05b64532f41e7e0430eacea4e3ef95a91e5b5edd844b3",
		"envelopes":         "1cfb23892a9d05efdbd66881debdc01a5105b9f2d43225ae9dab73860c512971",
		"monthly":           "358c4597414baea7b17a0134434ac7144c1b74c50633c15ef1923085abd282cb",
		"monthly-riskguard": "d029445ca13b08166088f48e58b3cc63ef727fcbe376ee3c1f1b52b76672e62a",
		"schedule":          "c836df55470ed70418294ef8a94e33cbfbf7c1933d98f6fb5cde3bf439272784",
		"lifetime-single":   "ec4ea7251ab386528b311913f47c6791ab14105c3b57268a550b8fb03eda4471",
		"lifetime-couple":   "2164eb86f2cc43d116f717b2319758cc20d18b5a7d7b507d600fb1978b0cccf4",
		"lifetime-annuity":  "e30b3c223d1c5ccd096301061c97929e500988da478d3d50e3ef592082332f81",
		"lifetime-amortize": "baf815526ca45a480c8a25a87bc5f98e4833e37a2b8839c0e73d2aed1bb49928",
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
		"parametric":           "1a2076379a8fbbad3b197ef13af88e4e1b063db309bff622d634a749ecdbc338",
		"parametric-normal":    "31b6895eec06259ee6f49391a7d23b69ceca5f56fe2771c13013894e7f5c5692",
		"markov-regime":        "061e49bfc3975dbf57b7d3728af60bcbad6d389777f5e42d6d15a997e129d3be",
		"lost-decade":          "c6d8d4b7d1bd0c669f9aa8c042af37410e6a975e4e426bab46d6ef354ee473f2",
		"glidepath":            "2e742df1bf94cfd176a39c22e60060c3023c1c177fc5abc0e006bfcf8bb4c86f",
		"block-bootstrap":      "2f87a7b074a036ee6d1d356bdc16f9e9e38a89051f7416367e7d13d07f814080",
		"stationary-bootstrap": "c06555f720afd948745b4bb209c67ea8c18af7a4f399860b276c203330e2c091",
		"historical-cohorts":   "9ad64c69483dde4dfbfe38f73b020b0e6cb5f739ac902ef3f79f53f0617eb25b",
		"pooled-bootstrap":     "6e373d5984971de3e6fd789442bd009d2e736880bc8d4ddc6be92400444a98c8",
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
