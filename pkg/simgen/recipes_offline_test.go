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

// stale returns a copy of s whose first half prints the previous close on all
// but every nth day, the shape of a thinly-traded listing's feed. The level
// path is preserved (a refresh day carries the whole move since the last one),
// only the calendar is fiction, which is exactly the case movesOnly exists for.
func stale(s *marketdata.Series, every int) *marketdata.Series {
	out := *s
	out.Points = make([]marketdata.Point, len(s.Points))
	half, held := len(s.Points)/2, s.Points[0].Close
	for i, p := range s.Points {
		if i >= half || i%every == 0 {
			held = p.Close
		}
		out.Points[i] = marketdata.Point{Date: p.Date, Close: held}
	}
	return &out
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
	// CHSN's donor is the fund's own distributing class, whose feed goes STALE
	// over the first half of its life: keep the stand-in shaped that way so
	// chsnBuild exercises movesOnly and the reshaping, not a daily fiction.
	chsnTwin := stale(from(mkWave(chsnDonor, n, 2e-4, 0.004, 1.2, 1.75), n/3), 5)
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

	// The cat bond share classes deal WEEKLY (semi-monthly for Solidum), which
	// is why their recipes match volatility on MONTH-END returns: keep the
	// stand-ins at those cadences so catBondBuild exercises monthlyVolMatch
	// rather than a daily fiction.
	catBond := func(name string, every int, amp, freq, phase float64) *marketdata.Series {
		s := &marketdata.Series{Symbol: name}
		v := 100.0
		for i := 0; i < n; i += every {
			s.Points = append(s.Points, marketdata.Point{Date: day(i), Close: v})
			v *= 1 + 3e-4 + amp*math.Sin(freq*float64(i)+phase)
		}
		return s
	}

	f := fakeFetcher{
		"IE00B3Q8M574": catBond("IE00B3Q8M574", 7, 0.018, 0.35, 0.2),
		"LI0049587301": catBond("LI0049587301", 15, 0.013, 0.30, 1.1),
		"LI0115208543": catBond("LI0115208543", 7, 0.020, 0.40, 2.0),
		"DBMF":         mf("DBMF", 0.008, 1.05, 0.3), "ASFYX": mf("ASFYX", 0.009, 1.05, 0.35),
		"RYMFX": mf("RYMFX", 0.007, 1.05, 0.4), "AHLPX": mf("AHLPX", 0.010, 1.05, 0.45),
		"AQMIX": mf("AQMIX", 0.008, 1.05, 0.5), "KMLM": mf("KMLM", 0.010, 1.05, 0.55),
		"CTA": mf("CTA", 0.012, 1.05, 0.6), "LU1103257975": mf("LU1103257975", 0.007, 1.05, 0.65),
		"IE0000360275": ahl,
		// The two AQR EUR classes whose real NAVs the hedged recipes splice:
		// the legacy B EUR donor (whose first date anchors IAE1FT's fee
		// schedule) and the flat-fee RAEF class, both quoting only over the
		// recent part of the window so the splices onto the reconstruction
		// are exercised.
		"LU1103258197": from(mf("LU1103258197", 0.007, 1.05, 0.7), n/3),
		"LU1662501532": from(mf("LU1662501532", 0.007, 1.05, 0.75), n/2),
		// The Campbell class the all-styles composite is calibrated on: a
		// young class, quoting only over the last third of the window, like
		// the real one (2020-06 against a 2000-01 file start).
		campbellB: from(mf(campbellB, 0.006, 1.05, 0.8), 2*n/3),
		"VFINX":   vfinx, "VTMGX": vtmgx, "VEIEX": veiex,
		"VFITX": vfitx, "VUSTX": vustx, "VFISX": vfisx, "VIPSX": vipsx,
		"TIP": tip, "STIP": stip,
		"GC=F": gold, "CL=F": crude, "^BCOM": bcom,
		"DFSVX": dfsvx, "DISVX": disvx, "AVUV": avuv, "AVDV": avdv, "IBCI": ibci,
		chsnDonor:      chsnTwin,
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
		// The four Xtrackers MSCI World classes the ERESMONDEM legs chain, each
		// quoting over a different tail of the window.
		"XWD1.DE": from(mkCombo("XWD1.DE", []*marketdata.Series{vfinx, vtmgx}, []float64{0.6, 0.4}), 3*n/4),
		"DBXW.DE": from(mkCombo("DBXW.DE", []*marketdata.Series{vfinx, vtmgx}, []float64{0.6, 0.4}), n/2),
		"XDWL.DE": from(mkCombo("XDWL.DE", []*marketdata.Series{vfinx, vtmgx}, []float64{0.6, 0.4}), 2*n/3),
		"XDWD.DE": from(mkCombo("XDWD.DE", []*marketdata.Series{vfinx, vtmgx}, []float64{0.6, 0.4}), n/2),
		// The single stock behind the ERES_DATADOG FCPE.
		"DDOG": from(mkWave("DDOG", n, 6e-4, 0.030, 1.1, 0.4), n/2),
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
			// Every recipe but the ILS index builds at daily or weekly cadence.
			// That one is month-end throughout (no ILS series quotes daily
			// anywhere), so twenty years of it is 250 points, not 5000.
			want := 300
			if r.ID == "ILSFUND" {
				want = 200
			}
			if s == nil || len(s.Points) < want {
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

// CHSN splices its own distributing class, and that class's feed goes stale
// for years: the reconstruction must keep every level the class published and
// take the days in between from the proxy, never pin them flat. Both halves
// are checked here, on a donor whose first half prints once a week and whose
// second half prints daily, against a proxy that moves every day.
func TestCHSNReshapesTheStaleDonorAndPassesTheCleanOneThrough(t *testing.T) {
	const n, live = 900, 300
	donor := stale(from(mkWave(chsnDonor, n, 3e-4, 0.004, 1.2, 1.75), live), 5)
	f := fakeFetcher{
		"IBCI":    mkWave("IBCI", n, 2e-4, 0.004, 1.2, 1.7),
		chsnDonor: donor,
	}

	got, err := chsnBuild(WithRefData(datasets.Refdata(), f), time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	byDate := map[time.Time]float64{}
	for _, p := range got.Points {
		byDate[p.Date] = p.Close
	}
	// Over the stale era the file's day-to-day moves must come from the
	// proxy, not from the donor's frozen prints: correlate the output's
	// returns with each candidate texture and let the numbers say which one
	// it followed.
	proxy, err := composite("proxy", []Leg{
		{ID: "IBCI", Weight: chsnIBCIBeta},
		{ID: "EURCASH-EUR", Weight: 1 - chsnIBCIBeta},
	}, "", 0)(WithRefData(datasets.Refdata(), f), time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	half := live + (n-live)/2
	returns := func(s *marketdata.Series) map[time.Time]float64 {
		out := map[time.Time]float64{}
		for i := 1; i < len(s.Points); i++ {
			out[s.Points[i].Date] = s.Points[i].Close/s.Points[i-1].Close - 1
		}
		return out
	}
	against := func(other *marketdata.Series) float64 {
		mine, theirs := returns(got), returns(other)
		var a, b []float64
		for i := live + 2; i < half; i++ {
			ra, oka := mine[day(i)]
			rb, okb := theirs[day(i)]
			if oka && okb {
				a, b = append(a, ra), append(b, rb)
			}
		}
		if len(a) < (half-live)/4 {
			t.Fatalf("only %d common days in the stale era, the check would be vacuous", len(a))
		}
		return pearson(a, b)
	}
	if c := against(proxy); c < 0.95 {
		t.Errorf("stale era correlates %.2f with the proxy's texture, want it to follow the proxy", c)
	}
	if c := against(donor); c > 0.5 {
		t.Errorf("stale era correlates %.2f with the donor's frozen prints, want that texture dropped", c)
	}
	// It must still pass through the donor's own levels, up to the fee
	// uplift, which is all that separates the two share classes. The whole
	// synthetic window sits inside the schedule's first step.
	uplift := func(a, b time.Time) float64 {
		return math.Pow(1-feeAt(chsnDonorFeeSteps, b), b.Sub(a).Hours()/24/365.25)
	}
	for _, i := range []int{live + 5, live + 100, live + 400, n - 1} {
		d, base := day(i), day(live)
		want := donor.Points[i-live].Close / donor.Points[0].Close * uplift(base, d)
		if got := byDate[d] / byDate[base]; math.Abs(got/want-1) > 1e-9 {
			t.Errorf("day %d: level ratio %.9f, want the donor's %.9f", i, got, want)
		}
	}
	// The proxy era is spliced in front, so the file starts where IBCI does.
	if got.First().Date != day(0) {
		t.Errorf("file starts %s, want the proxy's own first day", got.First().Date)
	}
}
