package golden

import (
	"io/fs"
	"strings"
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
// purpose: marketdata's fetch-time cleaner asks the excursion to cancel to
// within 2 % and to clear a per-class floor, simgen's despike asks it to cancel
// to within a third of the smaller leg. Neither is wrong and neither is widened
// here. What was missing is a check on the ARTEFACT rather than on the inputs:
// nothing measured what actually shipped, so a print both passes let through
// reached every consumer in silence.
//
// ERESMONDEM's entry is gone since 2026-09-20 and the way it went is the model
// for the rest. Nothing was widened: the recipe now holds each donor to the
// MSCI World reference it tracks (simgen's trackIndex), which sees what a
// series read on its own cannot, and the fabricated print never reaches the
// file. Every remaining entry is a series with no such reference behind it.
//
// The rule is marketdata.FindSpikes, the one the data doctor reports, and the
// three passes share its mechanism (marketdata.SpikeRule): the two legs must
// point opposite ways, EACH exceed 2 %, EACH exceed six local standard
// deviations (the suspect pair excluded from that estimate), and the round
// trip must cancel to within a third of the smaller leg. 1987-10-19, October
// 2008 and the March 2020 sessions clear none of the last two clauses; the
// bundled files' own worst real days do not either.

// knownSpikes are the round trips the bundle currently ships, each measured and
// dated. An entry is a defect deliberately left in place, never a bar widened
// to make a test pass, and it pins the DAY: a second such print in the same
// file still fails. Emptying this map is the job; adding to it needs the same
// argument every data decision in this repository needs.
var knownSpikes = map[string]map[string]string{
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
	for _, dir := range []string{"refdata", "simdata"} {
		fsys := bundleDir(dir)
		names, err := fs.Glob(fsys, "*.csv")
		if err != nil {
			t.Fatalf("%s: %v", dir, err)
		}
		for _, name := range names {
			id := strings.TrimSuffix(name, ".csv")
			key := dir + "/" + id
			if rateLevelSeries[key] {
				continue
			}
			s, ok, err := marketdata.ReadSimdataFS(fsys, id)
			if err != nil || !ok {
				t.Errorf("%s: read: ok=%v err=%v", key, ok, err)
				continue
			}
			checked++
			for _, sp := range marketdata.FindSpikes(s) {
				day := sp.Date.Format(time.DateOnly)
				if _, known := knownSpikes[key][day]; known {
					continue
				}
				t.Errorf("%s %s: %+.2f %% then %+.2f %% cancels inside a %.2f %% neighbourhood: "+
					"no instrument makes that round trip. Find the print, or add a measured entry to knownSpikes.",
					key, day, sp.In*100, sp.Out*100, sp.Sigma*100)
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
		s, ok, err := marketdata.ReadSimdataFS(bundleDir(dir), id)
		if err != nil || !ok {
			t.Errorf("knownSpikes names %s, which is not bundled (ok=%v err=%v)", key, ok, err)
			continue
		}
		found := map[string]bool{}
		for _, sp := range marketdata.FindSpikes(s) {
			found[sp.Date.Format(time.DateOnly)] = true
		}
		for day := range days {
			if !found[day] {
				t.Errorf("knownSpikes[%s] still excuses %s, which the file no longer carries: drop the entry", key, day)
			}
		}
	}
}

// cutKey splits a "dir/ID" key.
func cutKey(key string) (dir, id string, ok bool) {
	return strings.Cut(key, "/")
}

// bundleDir is the embedded directory a key's prefix names.
func bundleDir(dir string) fs.FS {
	if dir == "refdata" {
		return datasets.Refdata()
	}
	return datasets.Simdata()
}
