package simgen

import (
	"math"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
)

// mkWave builds a deterministic daily series whose returns oscillate around
// drift with its own frequency and phase, so different components stay
// decorrelated enough for regressions and vol targeting to behave.
func mkWave(symbol string, n int, drift, amp, freq, phase float64) *marketdata.Series {
	s := &marketdata.Series{Symbol: symbol}
	v := 100.0
	for i := range n {
		s.Points = append(s.Points, marketdata.Point{Date: day(i), Close: v})
		v *= 1 + drift + amp*math.Sin(freq*float64(i)+phase)
	}
	return s
}

// mkCombo builds a series whose daily returns are an exact linear combination
// of the given series' returns: the offline stand-in for a fund a recipe
// splices real quotes from.
func mkCombo(symbol string, parts []*marketdata.Series, weights []float64) *marketdata.Series {
	s := &marketdata.Series{Symbol: symbol}
	v := 100.0
	n := len(parts[0].Points)
	for i := range n {
		s.Points = append(s.Points, marketdata.Point{Date: day(i), Close: v})
		if i+1 < n {
			r := 0.0
			for k, p := range parts {
				r += weights[k] * (p.Points[i+1].Close/p.Points[i].Close - 1)
			}
			v *= 1 + r
		}
	}
	return s
}

// from returns the tail of a series, for a component that only starts quoting
// partway through the window.
func from(s *marketdata.Series, i int) *marketdata.Series {
	return &marketdata.Series{Symbol: s.Symbol, Points: s.Points[i:]}
}

// TestAllRecipesBuildOffline runs every bundled recipe's Build against a
// synthetic offline universe (canned component series + the embedded refdata,
// no network), asserting each returns a plausible series. This exercises the
// full wiring: frames, composites, the TSMOM engine,
// FX conversion (fxOnDates/convertDaily), the longBack splices and the
// dailyShape blends over the real embedded refdata.
func TestAllRecipesBuildOffline(t *testing.T) {
	const n = 1600
	vfinx := mkWave("VFINX", n, 4e-4, 0.010, 1.0, 0.1)
	vtmgx := mkWave("VTMGX", n, 3e-4, 0.009, 1.3, 0.7)
	veiex := mkWave("VEIEX", n, 3e-4, 0.012, 0.7, 1.9)
	vfitx := mkWave("VFITX", n, 2e-4, 0.003, 1.7, 0.4)
	vustx := mkWave("VUSTX", n, 2e-4, 0.007, 0.9, 2.3)
	vfisx := mkWave("VFISX", n, 1e-4, 0.001, 2.1, 1.1)
	vipsx := mkWave("VIPSX", n, 2e-4, 0.004, 1.1, 2.9)
	tip := mkWave("TIP", n, 2e-4, 0.004, 1.15, 2.8)
	stip := mkWave("STIP", n, 1e-4, 0.002, 1.25, 2.5)
	gold := mkWave("GC=F", n, 3e-4, 0.011, 0.5, 0.8)
	crude := mkWave("CL=F", n, 2e-4, 0.020, 1.9, 2.2)
	bcom := mkWave("^BCOM", n, 2e-4, 0.012, 1.6, 0.9)
	dfsvx := mkWave("DFSVX", n, 4e-4, 0.013, 0.8, 1.4)
	disvx := mkWave("DISVX", n, 3e-4, 0.011, 1.5, 0.2)
	// The same manager's own small-value sleeves quote over the recent part of
	// the window only, so avantisBuild exercises the splice onto the
	// Dimensional pair behind them rather than a full-length fiction.
	avuv := from(mkWave("AVUV", n, 4e-4, 0.013, 0.85, 1.35), n/3)
	avdv := from(mkWave("AVDV", n, 3e-4, 0.011, 1.45, 0.25), n/3)
	ibci := mkWave("IBCI", n, 2e-4, 0.004, 1.2, 1.7)
	indepFr := mkWave("LU0131510165", n, 4e-4, 0.012, 0.9, 2.1)
	vix := mkWave("^VIX", n, 0, 0.030, 0.6, 0.5)
	eurusd := mkWave("EURUSD=X", n, 0, 0.005, 1.4, 2.6)
	gbpusd := mkWave("GBPUSD=X", n, 0, 0.005, 0.4, 1.2)
	ezu := mkWave("EZU", n, 3e-4, 0.011, 1.1, 0.9)
	eunh := mkWave("EUNH.DE", n, 2e-4, 0.003, 1.6, 0.5)

	// The real managed-futures NAVs the trend recipes chain behind their
	// funds (DonorChain): the fund calibrated on, then its donors.
	mf := func(name string, amp, freq, phase float64) *marketdata.Series {
		return mkWave(name, n, 2e-4, amp, freq, phase)
	}
	// The deepest donor of every trend chain deals WEEKLY and starts before
	// the others, which is what the chain's sparse-cadence projection exists
	// for: keep the offline stand-in shaped that way so the recipes exercise
	// it (densify) rather than a daily fiction.
	ahl := &marketdata.Series{Symbol: "IE0000360275"}
	for i, v := -350, 100.0; i < n; i, v = i+7, v*(1+14e-4+0.008*math.Sin(1.05*float64(i)+0.7)) {
		ahl.Points = append(ahl.Points, marketdata.Point{Date: day(i), Close: v})
	}

	f := fakeFetcher{
		"DBMF": mf("DBMF", 0.008, 1.05, 0.3), "ASFYX": mf("ASFYX", 0.009, 1.05, 0.35),
		"RYMFX": mf("RYMFX", 0.007, 1.05, 0.4), "AHLPX": mf("AHLPX", 0.010, 1.05, 0.45),
		"AQMIX": mf("AQMIX", 0.008, 1.05, 0.5), "KMLM": mf("KMLM", 0.010, 1.05, 0.55),
		"CTA": mf("CTA", 0.012, 1.05, 0.6), "LU1103257975": mf("LU1103257975", 0.007, 1.05, 0.65),
		"IE0000360275": ahl,
		"VFINX":        vfinx, "VTMGX": vtmgx, "VEIEX": veiex,
		"VFITX": vfitx, "VUSTX": vustx, "VFISX": vfisx, "VIPSX": vipsx,
		"TIP": tip, "STIP": stip,
		"GC=F": gold, "CL=F": crude, "^BCOM": bcom,
		"DFSVX": dfsvx, "DISVX": disvx, "AVUV": avuv, "AVDV": avdv, "IBCI": ibci,
		"LU0131510165": indepFr,
		"EZU":          ezu, "EUNH.DE": eunh,
		"^IRX":     mkLevels("^IRX", n, 3.0),
		"^EONIA":   mkLevels("^EONIA", n, 1.5),
		"^ESTR":    mkLevels("^ESTR", n/2, 1.4),
		"^VIX":     vix,
		"EURUSD=X": eurusd, "GBPUSD=X": gbpusd,
		// Real iShares Core MSCI World that wpeaBuild grafts over the mid-period:
		// a 60/40 US/international combo, the MSCI World stand-in.
		"IE00B4L5Y983": mkCombo("IE00B4L5Y983", []*marketdata.Series{vfinx, vtmgx}, []float64{0.6, 0.4}),
	}
	fetcher := WithRefData(datasets.Refdata(), f)

	for _, r := range All() {
		t.Run(r.ID, func(t *testing.T) {
			if r.ID == "" || r.Name == "" || r.Method == "" {
				t.Fatalf("recipe %q: incomplete metadata: %+v", r.ID, r)
			}
			s, err := r.Build(fetcher, time.Time{})
			if err != nil {
				t.Fatalf("Build: %v", err)
			}
			if s == nil || len(s.Points) < 300 {
				t.Fatalf("Build returned %d points, want a substantial series", len(s.Points))
			}
			prev := time.Time{}
			for _, p := range s.Points {
				if !p.Date.After(prev) {
					t.Fatalf("dates not strictly ascending at %s", p.Date)
				}
				if p.Close <= 0 || math.IsNaN(p.Close) || math.IsInf(p.Close, 0) {
					t.Fatalf("bad close %v at %s", p.Close, p.Date)
				}
				prev = p.Date
			}
		})
	}
}

// A fund donor's NAV already carries its own manager's charge, so the target's
// load is due in full only over the era a fee-free proxy stands in for it, and
// the charge steps down as each donor's own quotes begin (feeGap). NTSX's legs
// are the case in point: the two Vanguard funds together cost more than the
// fund they stand in for, so the schedule ends at zero.
func TestFeeGapStepsDownAsEachDonorStartsCarryingItsOwnCharge(t *testing.T) {
	const n = 900
	f := fakeFetcher{
		"VFINX": from(mkLevels("VFINX", n, 100), 300),
		"VFITX": from(mkLevels("VFITX", n, 100), 600),
	}
	steps, err := feeGap(f, time.Time{}, 0.0020, []pricedLeg{
		{ID: "VFINX", Weight: 0.90, Load: vfinxTER},
		{ID: "VFITX", Weight: 0.60, Excess: true, Load: vfitxTER},
		{ID: "^IRX", Weight: 0.10},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		when time.Time
		want float64
		why  string
	}{
		{day(100), 0.0020, "index and CMT reconstruction, nothing deducted yet"},
		{day(400), 0.0020 - 0.90*vfinxTER, "the equity fund carries its own 0.126 %/yr"},
		{day(700), 0, "both funds together cost 0.246 %/yr, more than the target"},
	} {
		if got := feeAt(steps, c.when); math.Abs(got-c.want) > 1e-12 {
			t.Errorf("charge on %s is %.5f, want %.5f (%s)", c.when.Format("2006-01-02"), got, c.want, c.why)
		}
	}
}

// afterFeeSteps compounds each step on the calendar, so a year under a given
// charge costs exactly that charge whatever the quote count.
func TestAfterFeeStepsCompoundsEachEraOnTheCalendar(t *testing.T) {
	flat := mkLevels("X", 900, 100)
	steps := []feeStep{{Annual: 0.02}, {From: day(365), Annual: 0.01}}
	got := afterFeeSteps("X (fee-aligned)", flat, steps)
	at := func(i int) float64 { return got.Points[i].Close }
	for _, c := range []struct {
		from, to int
		want     float64
	}{
		{0, 364, 1 - 0.02},
		{365, 729, 1 - 0.01},
	} {
		years := day(c.to).Sub(day(c.from)).Hours() / 24 / 365.25
		if ratio := at(c.to) / at(c.from); math.Abs(ratio-math.Pow(c.want, years)) > 1e-9 {
			t.Errorf("days %d-%d lost %.6f, want %.6f", c.from, c.to, 1-ratio, 1-math.Pow(c.want, years))
		}
	}
}

// The Avantis fund takes its geography from its own factsheet (70 % US) and
// its donor from its own manager: over the era AVUV and AVDV quote, they alone
// drive the blend, at the adopted weights and net of the difference in ongoing
// charges (avantisRecipe). Before them the Dimensional pair takes over, which
// here is flat, so every move in the output belongs to the same-manager pair.
func TestAvantisBlendsSameManagerSleevesAtTheFundsGeography(t *testing.T) {
	const n, live = 900, 600
	flat := func(sym string, start int) *marketdata.Series {
		return atSeries(sym, start, n-start, 100)
	}
	f := fakeFetcher{
		"DFSVX": flat("DFSVX", 0), "DISVX": flat("DISVX", 0),
		"AVUV":     from(mkWave("AVUV", n, 1e-3, 0, 1, 0), live),
		"AVDV":     flat("AVDV", live),
		"EURUSD=X": flat("EURUSD=X", 0),
	}

	got, err := avantisBuild(f, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	byDate := map[time.Time]float64{}
	for _, p := range got.Points {
		byDate[p.Date] = p.Close
	}
	// A day inside the Dimensional era: both legs flat, so is the blend.
	if a, b := byDate[day(300)], byDate[day(301)]; a == 0 || b/a-1 > 1e-9 {
		t.Errorf("Dimensional era moves by %.2e/day, want flat legs to stay flat", b/a-1)
	}
	// A day inside the same-manager era: 0.70 of AVUV's move, less the
	// 0.107 %/yr the pair's cheaper wrapper owes the fund's.
	want := avwsUS*1e-3 - (avwsTER-(avwsUS*avuvTER+avwsIntl*avdvTER))/252
	a, b := byDate[day(live+100)], byDate[day(live+101)]
	if a == 0 || math.Abs(b/a-1-want) > 1e-9 {
		t.Errorf("same-manager era moves %.6f/day, want %.6f (0.70×AVUV, fee-aligned)", b/a-1, want)
	}
}
