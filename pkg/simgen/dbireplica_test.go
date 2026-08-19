package simgen

import (
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// dbiFakeDays is long enough for the regression's warm-up and the minimum
// history dbiProjection insists on.
const dbiFakeDays = 400

// dbiFake builds a self-contained world for the replication engine: every leg
// quotes a random walk, cash pays nothing, and the reference is whatever the
// caller makes of the legs' own excess returns.
type dbiFake struct {
	f      fakeFetcher
	dates  []time.Time
	excess [][]float64 // per leg, as the engine reads them
}

func newDBiFake(t *testing.T) *dbiFake {
	t.Helper()
	rng := rand.New(rand.NewSource(7))
	f := fakeFetcher{}
	walk := func(id string) *marketdata.Series {
		s := &marketdata.Series{Symbol: id}
		v := 100.0
		for i := range dbiFakeDays {
			s.Points = append(s.Points, marketdata.Point{Date: day(i), Close: v})
			v *= 1 + rng.NormFloat64()*0.01
		}
		return s
	}
	flat := func(id string, level float64) *marketdata.Series {
		s := &marketdata.Series{Symbol: id}
		for i := range dbiFakeDays {
			s.Points = append(s.Points, marketdata.Point{Date: day(i), Close: level})
		}
		return s
	}
	for _, leg := range dbiLegs {
		if leg.Near != "" {
			f[leg.Near] = walk(leg.Near)
		}
		f[leg.Deep] = walk(leg.Deep)
		if leg.Maturity > 0 {
			f[leg.Deep] = flat(leg.Deep, 3) // a yield series, priced as a par bond
		}
	}
	f["^IRX"] = flat("^IRX", 0)           // no cash, so an excess return is a return
	f["^EONIA"] = flat("^EONIA", 0)       // and no euro carry either
	f["JPCASH-JPY"] = flat("JPCASH", 100) // nor yen carry

	d := &dbiFake{f: f}
	for i := range dbiFakeDays {
		d.dates = append(d.dates, day(i))
	}
	cash := make([]float64, len(d.dates))
	ef := extend(f)
	for _, leg := range dbiLegs {
		s, cut, err := dbiLegSeries(ef, leg, day(0))
		if err != nil {
			t.Fatalf("leg %s: %v", leg.Name, err)
		}
		var carry *marketdata.Series
		if leg.Kind == dbiFX {
			carry = f["JPCASH-JPY"]
			if leg.Carry == "EUR" {
				carry, err = eurOvernightDeep(ef, day(0))
				if err != nil {
					t.Fatal(err)
				}
			}
		}
		d.excess = append(d.excess, dbiExcess(leg, s, carry, cut, d.dates, cash))
	}
	return d
}

// reference installs a composite whose daily return is drift plus the given
// combination of the legs' excess returns.
func (d *dbiFake) reference(betas []float64, drift float64) {
	s := &marketdata.Series{Symbol: allStylesIndex, Name: "fake composite"}
	v := 100.0
	s.Points = append(s.Points, marketdata.Point{Date: d.dates[0], Close: v})
	for i := 1; i < len(d.dates); i++ {
		r := drift
		for j, b := range betas {
			r += b * d.excess[j][i]
		}
		v *= 1 + r
		s.Points = append(s.Points, marketdata.Point{Date: d.dates[i], Close: v})
	}
	d.f[allStylesIndex] = s
}

func dbiReturns(s *marketdata.Series) map[time.Time]float64 {
	out := make(map[time.Time]float64, len(s.Points))
	for i := 1; i < len(s.Points); i++ {
		out[s.Points[i].Date] = s.Points[i].Close/s.Points[i-1].Close - 1
	}
	return out
}

// A composite that IS a fixed portfolio of the ten contracts is reproduced by
// the projection exactly, which is the only end-to-end statement about the
// regression worth making: the betas it recovers are the ones that generated
// the data.
func TestDBiProjectionRecoversKnownBetas(t *testing.T) {
	d := newDBiFake(t)
	betas := []float64{0.4, -0.3, 0.2, -1.5, 0.8, -0.25, 0.5, -0.15, 0.6, -0.45}
	d.reference(betas, 0)

	proj, err := DBiProjection(d.f, day(0))
	if err != nil {
		t.Fatal(err)
	}
	want := dbiReturns(d.f[allStylesIndex])
	n := 0
	for date, g := range dbiReturns(proj) {
		w, ok := want[date]
		if !ok {
			t.Fatalf("projection quotes %s, the composite does not", date.Format("2006-01-02"))
		}
		if math.Abs(g-w) > 1e-9 {
			t.Fatalf("%s: projection %.10f, composite %.10f", date.Format("2006-01-02"), g, w)
		}
		n++
	}
	if n < 300 {
		t.Fatalf("only %d days compared", n)
	}
}

// The intercept is fitted and thrown away: a composite that beats its own ten
// contracts by a constant every day hands the projection none of it. That is
// what makes the projection a replicator's return rather than a manager's.
func TestDBiProjectionDropsTheIntercept(t *testing.T) {
	d := newDBiFake(t)
	betas := []float64{0.4, -0.3, 0.2, -1.5, 0.8, -0.25, 0.5, -0.15, 0.6, -0.45}
	const drift = 0.0002 // per day
	d.reference(betas, drift)

	proj, err := DBiProjection(d.f, day(0))
	if err != nil {
		t.Fatal(err)
	}
	want := dbiReturns(d.f[allStylesIndex])
	for date, g := range dbiReturns(proj) {
		if math.Abs(g-(want[date]-drift)) > 1e-9 {
			t.Fatalf("%s: projection %.10f, composite less its drift %.10f",
				date.Format("2006-01-02"), g, want[date]-drift)
		}
	}
}

// With no exposure at all the projection is pure collateral: it earns the bill
// rate and nothing else. Reading a futures book as if it were unfunded loses
// exactly that, which was worth six points a year in the 1990s.
func TestDBiProjectionEarnsCashOnItsCollateral(t *testing.T) {
	d := newDBiFake(t)
	d.f["^IRX"] = &marketdata.Series{Symbol: "^IRX"}
	for i := range dbiFakeDays {
		d.f["^IRX"].Points = append(d.f["^IRX"].Points, marketdata.Point{Date: day(i), Close: 4})
	}
	d.reference(make([]float64, len(dbiLegs)), 0) // the composite is cash-like: no exposure

	proj, err := DBiProjection(d.f, day(0))
	if err != nil {
		t.Fatal(err)
	}
	years := proj.Points[len(proj.Points)-1].Date.Sub(proj.First().Date).Hours() / 24 / 365.25
	grew := math.Pow(proj.Points[len(proj.Points)-1].Close/proj.First().Close, 1/years) - 1
	if math.Abs(grew-0.04) > 0.002 {
		t.Fatalf("collateral earned %.4f %%/yr, want the 4 %% bill rate", 100*grew)
	}
}

// A proxy's own ongoing charge is given back, because the futures contract it
// stands in for levies none, and cash is netted off a funded leg.
func TestDBiExcessRestoresTheProxyCharge(t *testing.T) {
	const daily = 0.001
	s := mkSeries("PROXY", 30, daily)
	dates := make([]time.Time, 30)
	cash := make([]float64, 30)
	for i := range dates {
		dates[i] = day(i)
		cash[i] = 0.0001
	}
	leg := dbiLeg{Name: "test", Near: "PROXY", NearFee: 0.00365, Kind: dbiFunded}
	got := dbiExcess(leg, s, nil, day(0), dates, cash)
	want := daily - 0.0001 + 0.00365/365.25
	if math.Abs(got[5]-want) > 1e-12 {
		t.Fatalf("funded excess %.10f, want %.10f", got[5], want)
	}

	leg.Kind = dbiPrice
	if got = dbiExcess(leg, s, nil, day(0), dates, cash); math.Abs(got[5]-daily) > 1e-12 {
		t.Fatalf("a price return is already an excess return: %.10f, want %.10f", got[5], daily)
	}
}

// A currency future pays the spot move plus the interest differential its
// forward carries; leaving the differential out ships an unfunded currency.
func TestDBiExcessAddsTheCurrencyCarry(t *testing.T) {
	const spot = 0.002
	s := mkSeries("EURUSD=X", 30, spot)
	carry := mkSeries("EUR cash", 30, 0.0003)
	dates := make([]time.Time, 30)
	cash := make([]float64, 30)
	for i := range dates {
		dates[i] = day(i)
		cash[i] = 0.0001
	}
	leg := dbiLeg{Name: "EUR", Deep: "EURUSD=X", Kind: dbiFX}
	got := dbiExcess(leg, s, carry, time.Time{}, dates, cash)
	want := spot + 0.0003 - 0.0001
	if math.Abs(got[5]-want) > 1e-9 {
		t.Fatalf("currency excess %.10f, want %.10f", got[5], want)
	}
}

// The blend averages RETURNS day by day, so a series blended with itself is
// itself and the weights are what they say they are.
func TestBlendReturnsAveragesDailyReturns(t *testing.T) {
	a := mkSeries("A", 40, 0.002)
	b := mkSeries("B", 40, -0.001)
	same := blendReturns("same", a, a, 0.5)
	for i := 1; i < len(same.Points); i++ {
		r := same.Points[i].Close/same.Points[i-1].Close - 1
		if math.Abs(r-0.002) > 1e-12 {
			t.Fatalf("a series blended with itself moved %.10f, want 0.002", r)
		}
	}
	half := blendReturns("half", a, b, 0.5)
	for i := 1; i < len(half.Points); i++ {
		r := half.Points[i].Close/half.Points[i-1].Close - 1
		if math.Abs(r-0.0005) > 1e-12 {
			t.Fatalf("blended return %.10f, want 0.0005", r)
		}
	}
	quarter := blendReturns("quarter", a, b, 0.25)
	r := quarter.Points[1].Close/quarter.Points[0].Close - 1
	if math.Abs(r-(0.75*0.002+0.25*-0.001)) > 1e-12 {
		t.Fatalf("weighted blend %.10f", r)
	}
}

// A singular fit must not trade: the engine keeps the positions it had rather
// than solving a broken system.
func TestSolveSymmetricRefusesASingularSystem(t *testing.T) {
	if _, ok := solveSymmetric([][]float64{{1, 2}, {2, 4}}, []float64{1, 2}); ok {
		t.Fatal("a singular system was solved")
	}
	x, ok := solveSymmetric([][]float64{{2, 0}, {0, 4}}, []float64{2, 8})
	if !ok || math.Abs(x[0]-1) > 1e-12 || math.Abs(x[1]-2) > 1e-12 {
		t.Fatalf("solve gave %v (ok=%v)", x, ok)
	}
}
