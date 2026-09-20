package golden

import (
	"io/fs"
	"math"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
)

// A ONE-SESSION ROUND TRIP is a level that leaves and comes straight back: a
// print no instrument made, sitting between two neighbours that agree with each
// other. It is the cheapest defect to see and the dearest to leave in, because
// a reconstruction does not merely carry it, it multiplies it. ERESMONDEM's
// 2008-02-04 is the case that prompted this guard: one fabricated Xetra close
// in a donor ETF, spliced at a 0.75 weight, gives the shipped world-equity file
// a -24.1 % day followed by a +31.1 % one. The whole file's annualized daily
// volatility reads 17.09 % against 16.25 % without that single point, and 2008,
// the year a reader most wants to look up, reads 50.0 % against 31.5 %.
//
// Two hygiene passes already hunt this shape and both declined these prints on
// purpose: marketdata's dropRoundTrips asks the excursion to cancel to within
// 2 % and to clear a per-class floor, simgen's despike asks it to cancel to
// within a third of the smaller leg. Neither is wrong and neither is widened
// here. What was missing is a check on the ARTEFACT rather than on the inputs:
// nothing measured what actually shipped, so a print both passes let through
// reached every consumer in silence.
//
// The rule is deliberately narrow, so that a real market day can never trip it:
// the two legs must point opposite ways, EACH exceed spikeFloor, EACH exceed
// spikeZ local standard deviations (the suspect pair excluded from that
// estimate), and the round trip must cancel to within a third of the smaller
// leg. 1987-10-19, October 2008 and the March 2020 sessions clear none of the
// last two clauses; the bundled files' own worst real days do not either.
const (
	// spikeZ: how many local standard deviations each leg must span. Six is
	// the bar both existing passes use, for the same reason.
	spikeZ = 6.0
	// spikeWindow: returns on each side defining "local".
	spikeWindow = 25
	// spikeFloor: the smallest leg worth calling fabricated. Below it the
	// candidates are cash-like files whose local sigma is a rounding error, so
	// the z test alone flags arithmetic rather than data.
	spikeFloor = 0.02
)

// knownSpikes are the round trips the bundle currently ships, each measured and
// dated. An entry is a defect deliberately left in place, never a bar widened
// to make a test pass, and it pins the DAY: a second such print in the same
// file still fails. Emptying this map is the job; adding to it needs the same
// argument every data decision in this repository needs.
var knownSpikes = map[string]map[string]string{
	"simdata/ERESMONDEM": {
		"2008-02-04": "Yahoo quotes the donor DBXW.DE at 23.04 between 33.93 and 32.75 " +
			"(-32.1 % then +42.2 %); at the leg's 0.75 weight the file reads -24.1 % then +31.1 %. " +
			"The excursion leaves 3.5 % standing, past dropRoundTrips' 2 % bar. Measured 2026-09-20.",
	},
	"simdata/IE00BKM4GZ66": {
		"2001-07-16": "Yahoo quotes the donor VEIEX at 4.5857 between 4.3443 and 4.3175 " +
			"(+5.6 % then -5.9 %). Both legs clear six local sigmas and the round trip cancels, " +
			"but VEIEX reaches no plausibility band of its own, so dropRoundTrips' class floor " +
			"(the widest row, ~15 % a leg) declines it. Measured 2026-09-20.",
	},
	"simdata/IE00B3Q8M574": {
		"2025-03-09": "a Sunday-dated provider NAV of the fund itself (17.261 between 16.6055 " +
			"and 16.6571, +4.0 % then -3.5 % on a line whose local sigma is 0.28 %). It sits in " +
			"the REAL quote era, which a recipe never rewrites. Measured 2026-09-20.",
	},
}

// rateLevelSeries are the bundled files quoted as annualized percent LEVELS
// rather than as prices. A ratio of two rates is not a return, so the whole
// test means nothing on them.
var rateLevelSeries = map[string]bool{
	"refdata/TBILL-3M":            true,
	"refdata/TREASURY-LONG-YIELD": true,
}

// TestBundledSeriesHaveNoFabricatedRoundTrips walks every embedded refdata and
// simdata file and refuses a one-session round trip no instrument could have
// made, unless it is a measured entry of knownSpikes.
func TestBundledSeriesHaveNoFabricatedRoundTrips(t *testing.T) {
	checked := 0
	for _, dir := range []struct {
		name string
		fsys fs.FS
	}{{"refdata", datasets.Refdata()}, {"simdata", datasets.Simdata()}} {
		names, err := fs.Glob(dir.fsys, "*.csv")
		if err != nil {
			t.Fatalf("%s: %v", dir.name, err)
		}
		for _, name := range names {
			id := name[:len(name)-len(".csv")]
			key := dir.name + "/" + id
			if rateLevelSeries[key] {
				continue
			}
			s, ok, err := marketdata.ReadSimdataFS(dir.fsys, id)
			if err != nil || !ok {
				t.Errorf("%s: read: ok=%v err=%v", key, ok, err)
				continue
			}
			checked++
			for _, sp := range roundTrips(s.Points) {
				day := sp.date.Format("2006-01-02")
				if _, known := knownSpikes[key][day]; known {
					continue
				}
				t.Errorf("%s %s: %+.2f %% then %+.2f %% cancels inside a %.2f %% neighbourhood: "+
					"no instrument makes that round trip. Find the print, or add a measured entry to knownSpikes.",
					key, day, sp.first*100, sp.second*100, sp.sigma*100)
			}
		}
	}
	if checked < 80 {
		t.Fatalf("only %d bundled series checked: the embed is not wired", checked)
	}
}

// TestKnownSpikesAreStillThere keeps the exception list honest: an entry that
// no longer matches anything is a note about a file that has moved on, and it
// would silently stop guarding the day it names.
func TestKnownSpikesAreStillThere(t *testing.T) {
	for key, days := range knownSpikes {
		dir, id, _ := cutKey(key)
		var fsys fs.FS = datasets.Simdata()
		if dir == "refdata" {
			fsys = datasets.Refdata()
		}
		s, ok, err := marketdata.ReadSimdataFS(fsys, id)
		if err != nil || !ok {
			t.Errorf("knownSpikes names %s, which is not bundled (ok=%v err=%v)", key, ok, err)
			continue
		}
		found := map[string]bool{}
		for _, sp := range roundTrips(s.Points) {
			found[sp.date.Format("2006-01-02")] = true
		}
		for day := range days {
			if !found[day] {
				t.Errorf("knownSpikes[%s] still excuses %s, which the file no longer carries: drop the entry", key, day)
			}
		}
	}
}

func cutKey(key string) (dir, id string, ok bool) {
	for i := range key {
		if key[i] == '/' {
			return key[:i], key[i+1:], true
		}
	}
	return "", key, false
}

// spike is one fabricated round trip: the two legs and the local sigma they
// were judged against.
type spike struct {
	date          time.Time
	first, second float64
	sigma         float64
}

// roundTrips reports the points whose arrival and departure returns are
// opposite, each beyond spikeFloor and spikeZ local sigmas, and which cancel to
// within a third of the smaller leg.
func roundTrips(pts []marketdata.Point) []spike {
	n := len(pts)
	if n < 2*spikeWindow {
		return nil
	}
	ret := func(i int) float64 { return pts[i].Close/pts[i-1].Close - 1 }
	var out []spike
	for i := 1; i+1 < n; i++ {
		r1, r2 := ret(i), ret(i+1)
		if r1*r2 >= 0 || math.Abs(r1) < spikeFloor || math.Abs(r2) < spikeFloor {
			continue
		}
		if math.Abs((1+r1)*(1+r2)-1) >= math.Min(math.Abs(r1), math.Abs(r2))/3 {
			continue
		}
		sigma, ok := localSigma(pts, i)
		if !ok || math.Abs(r1) <= spikeZ*sigma || math.Abs(r2) <= spikeZ*sigma {
			continue
		}
		out = append(out, spike{date: pts[i].Date, first: r1, second: r2, sigma: sigma})
	}
	return out
}

// localSigma is the standard deviation of the returns around index i with the
// suspect pair (i and i+1) left out, so a fabricated print cannot inflate the
// yardstick it is measured against.
func localSigma(pts []marketdata.Point, i int) (float64, bool) {
	var sum, sumsq float64
	n := 0
	for j := max(1, i-spikeWindow); j <= min(len(pts)-1, i+spikeWindow); j++ {
		if j == i || j == i+1 || pts[j-1].Close <= 0 {
			continue
		}
		r := pts[j].Close/pts[j-1].Close - 1
		sum += r
		sumsq += r * r
		n++
	}
	if n < 10 {
		return 0, false
	}
	mean := sum / float64(n)
	v := sumsq/float64(n) - mean*mean
	if v <= 0 {
		return 0, false
	}
	return math.Sqrt(v), true
}
