package marketdata

import (
	"math"
	"strings"
)

// cleanQuotes applies the price-hygiene passes to a freshly fetched daily
// series: strip provider placeholders and isolated bad prints, repair a single
// denomination break, drop one-session round trips no asset of the class could
// have made, and (for FX crosses) drop self-cancelling spikes. Rate symbols are
// exempt (a real low rate is not a bad print). BOTH the direct download path
// and the resolved-ISIN path (FT / Morningstar / Stooq) must run it, so a scale
// break or a redenomination never reaches the metrics whatever the source.
//
// The order is not free, and each pass hands the next a series it can reason
// about:
//
//  1. dropDropouts first. It removes points, and it recognizes its targets by
//     their sheer distance from BOTH neighbours (a quarter or less), which no
//     other pass would repair. Leaving a leading placeholder run in place would
//     also mislead mendScaleBreak about where the series really starts.
//  2. mendScaleBreak second, on a series whose stray prints are already gone,
//     so the one junction it insists on finding is the real one.
//  3. dropRoundTrips last, because its evidence is a LOCAL standard deviation:
//     it can only tell an impossible session from a violent one once the series
//     is a single instrument expressed in fixed units. Run before mending, a
//     hundredfold junction inside its window would swamp that sigma and silence
//     it exactly where the data is worst. It is also the only pass that needs
//     to know WHAT the asset is, which it asks bandFor.
//
// id is whatever the caller holds: a canonical id, an ISIN, or a bare provider
// symbol. All three reach a plausibility band, and an unrecognized one reaches
// the widest, which only makes the cleaning timid.
func cleanQuotes(id string, pts []Point) []Point {
	if !isRateSymbol(id) {
		pts = dropDropouts(pts)                // strip provider placeholders/bad prints
		pts = mendScaleBreak(pts)              // repair a single denomination break
		pts = dropRoundTrips(pts, bandFor(id)) // strip impossible one-session round trips
	}
	if strings.HasSuffix(id, "=X") {
		pts = dropFXSpikes(pts) // strip isolated self-cancelling FX prints
	}
	return pts
}

// dropoutRatio is the fraction below which a close, relative to the surrounding
// stable level, is treated as a data dropout (a provider placeholder or a bad
// print) rather than a real quote. 0.25 means a >75% one-day collapse that
// recovers 4x the next day, or a leading value below a quarter of the first
// real price: both are physically impossible for a genuine close-to-close move,
// so removing them never discards real data.
const dropoutRatio = 0.25

// isRateSymbol reports whether a symbol is a yield series quoted in annualized
// percent (Yahoo's ^IRX, ^TNX, …) rather than a price. Such series legitimately
// visit near-zero levels (e.g. ^IRX at ~0.003 % when the Fed cut to zero in
// March 2020), so the dropout filter must not touch them: a real low rate is
// not a bad print. Mirrors simgen.isRate. ^VIX, though an annualized percent
// level too, deliberately stays a "price" here: it never legitimately visits
// the extremes the repairs look for (near zero, 8x in a day), so the cleaning
// passes only protect it from provider bad prints.
func isRateSymbol(symbol string) bool {
	switch symbol {
	case "^IRX", "^FVX", "^TNX", "^TYX":
		return true
	}
	return false
}

// scaleBreakFactor is the minimum adjacent jump (>= this, or <= its reciprocal)
// treated as a denomination break rather than a real move. 8x in a single day of
// a split-adjusted close is not a market move or a residual split (Yahoo already
// back-adjusts splits): it is a provider splicing two segments of the same fund
// at different units (pence vs pounds, an old share class, …).
const scaleBreakFactor = 8.0

// minScaleSegment is the smallest segment (in points) on either side of a break
// that is trusted as a real scale. A shorter side is a stray tail or a leading
// placeholder (dropDropouts territory), not an authoritative denomination.
const minScaleSegment = 20

// euroLegacyRates are the irrevocable EUR fixed conversion rates of legacy
// currencies that fall BELOW scaleBreakFactor, so a fund NAV series spliced
// across the 1999/2001 changeover shows a denomination break the plain 8x
// magnitude test misses. The French franc (6.55957) and Finnish markka
// (5.94573) are the only euro-legacy rates in that 1..8 band; larger rates (the
// lira at 1936, peseta at 166, schilling at 13.76, …) already exceed
// scaleBreakFactor, and smaller ones (mark 1.96, guilder 2.20) sit too close to
// genuine intraday moves to mend automatically.
var euroLegacyRates = []float64{6.55957, 5.94573}

// isDenominationBreak reports whether an adjacent ratio r = new/old is a splice
// artefact rather than a market move: either beyond the 8x magnitude threshold,
// or matching (within 1.5%, absorbing a few days of real drift across the
// junction) a known euro-legacy conversion rate in either direction.
func isDenominationBreak(r float64) bool {
	if r >= scaleBreakFactor || r <= 1/scaleBreakFactor {
		return true
	}
	return nearRatio(r, euroLegacyRates)
}

// unitRatios are the magnitudes one instrument can legitimately change units
// by: powers of ten (pence vs pounds, cents vs units), the round ratios of a
// share split the provider forgot to back-adjust, and the euro-legacy
// conversion rates. Only ratios at or beyond scaleBreakFactor (plus the euro
// rates, which sit below it) can ever reach the test, so the small split
// ratios are deliberately absent.
var unitRatios = append([]float64{10, 20, 25, 50, 100, 200, 500, 1000, 10000}, euroLegacyRates...)

// nearRatio reports whether r matches one of the ratios, in either direction,
// within the 1.5 % that absorbs the real market move sitting across a junction.
func nearRatio(r float64, ratios []float64) bool {
	for _, want := range ratios {
		if math.Abs(r/want-1) <= 0.015 || math.Abs(r*want-1) <= 0.015 {
			return true
		}
	}
	return false
}

// mendScaleBreak repairs a single large, persistent denomination break: when a
// series has EXACTLY ONE adjacent jump beyond scaleBreakFactor with a
// substantial segment on both sides AND that jump is a plain change of units,
// the older segment is rescaled onto the newer one (the recent segment is the
// current NAV, hence authoritative). The real ITPS.L shape: pre-2009 history at
// ~100x the true NAV, a -99% cliff at the junction; after mending the whole
// series is continuous.
//
// The unit-ratio test is what keeps the repair honest. A break at an arbitrary
// magnitude is not one instrument re-expressed in new units but two DIFFERENT
// quote lines spliced together, and welding those hides a currency change
// inside a series that FX conversion then treats as uniform. The real IBGS.L
// case: its 2008 segment is the fund's EUR NAV in cents and its 2009+ segment
// its GBP line, so the junction carries the EUR/GBP rate on top of the 100x
// (104.0, not 100). Mending it used to bake a fictitious -21 % into 2008 as
// soon as the series was converted back to EUR; leaving it alone surfaces the
// cliff to the -verify-data doctor instead, which is the whole point.
//
// It is deliberately timid: no break, several breaks, an off-ratio junction or
// a too-short side are all left untouched for the -verify-data doctor to flag,
// because rescaling them automatically could corrupt genuine data. Run it AFTER
// dropDropouts, so leading placeholders and isolated bad prints are already
// gone.
//
// CL2.PA is the standing example of the off-ratio abstention, and it is a real
// split rather than a currency weld. Amundi's leveraged MSCI USA ETF divided
// its unit by 300 over the 2011/2012 new year and Yahoo never back-adjusted
// the older segment: 262.17 on 2011-12-30, 0.9094 on 2012-01-02, a junction of
// 1/288.3. The pre-split line survives on the holidays dropDropouts now clears
// (see there) and puts the two lines 296x to 303x apart, so 300 is certainly
// the number; but the junction itself is 3.9 % away from it, far outside the
// 1.5 % nearRatio allows, and the gap is unexplained. Mending at 288.3 would
// bake that 4 % into every pre-2012 return, so the series keeps its one honest
// finding instead.
func mendScaleBreak(pts []Point) []Point {
	n := len(pts)
	if n < 2*minScaleSegment {
		return pts
	}
	breakIdx, count := -1, 0
	for i := 1; i < n; i++ {
		if pts[i-1].Close <= 0 || pts[i].Close <= 0 {
			continue
		}
		if isDenominationBreak(pts[i].Close / pts[i-1].Close) {
			breakIdx, count = i, count+1
		}
	}
	if count != 1 || breakIdx < minScaleSegment || n-breakIdx < minScaleSegment {
		return pts
	}
	// Rescale the older segment [0, breakIdx) so its last point meets the newer
	// segment: the spurious junction move is absorbed (a normal day's real move
	// across it is negligible next to an 8x+ break).
	f := pts[breakIdx].Close / pts[breakIdx-1].Close
	if !nearRatio(f, unitRatios) {
		return pts
	}
	for i := 0; i < breakIdx; i++ {
		pts[i].Close *= f
	}
	return pts
}

// dropDropouts removes obvious bad prints from a sorted daily series:
//
//   - an isolated interior point away from BOTH neighbours by at least the
//     fourfold dropoutRatio implies, in either direction. Down is the classic
//     dropout: a one-day collapse that immediately recovers. Up is the same
//     artefact seen from the other side, and the real CL2.PA case: on the
//     French half-days and holidays the current quote line does not print
//     (24/12, 31/12, 01/05), Yahoo fills the session from the fund's
//     PRE-SPLIT line instead, so 2014-05-01 closes at 522.81 between 1.7312
//     and 1.7440, a ~300x round trip in two sessions. Seven such pairs stand
//     in that series between 2013 and 2015.
//   - a leading run of placeholder points before the price first jumps up (>=4x)
//     to its stable level, e.g. Yahoo emitting a tiny value at an ETF inception
//     (the real IB01 case: closes of 5 before the true ~99 NAV, which otherwise
//     shows as a +1884% daily move).
//
// The order inside the function is the one that survives CL2.PA. The isolated
// prints go first, because the leading-run rule reads the first fourfold jump
// up as the end of the placeholders, and a lone 300x spike further in is
// exactly such a jump: judged before the spike is gone, it would condemn every
// real quote before it. The leading run is then bounded by minScaleSegment for
// the same reason it bounds a scale: a prefix that long is a segment of real
// history, whatever its level, and rescaling or dropping it is mendScaleBreak's
// business and no one else's.
//
// It deliberately keeps everything else: moderate spikes (they may be real
// distributions or genuine volatility), permanent declines (a low tail that
// never recovers), permanent steps (a split or a redenomination moves the level
// and LEAVES it there, so the both-neighbours test never sees one), and gradual
// growth from a low base (which never quadruples in a single day). The
// -verify-data doctor still surfaces the moderate moves for human review; this
// only strips the unambiguous artefacts.
func dropDropouts(pts []Point) []Point {
	n := len(pts)
	if n < 3 {
		return pts
	}
	bad := make([]bool, n)

	// Isolated interior outliers: a point far from both its neighbours, either
	// way. No instrument quadruples in a session and hands it all back the next,
	// nor the reverse; a real move of that size stays where it went.
	for i := 1; i+1 < n; i++ {
		cur, prev, next := pts[i].Close, pts[i-1].Close, pts[i+1].Close
		low := cur < dropoutRatio*prev && cur < dropoutRatio*next
		high := cur*dropoutRatio > prev && cur*dropoutRatio > next
		if low || high {
			bad[i] = true
		}
	}

	// Leading placeholder run, read over the points the pass above left
	// standing: find the first adjacent jump up of at least 1/dropoutRatio
	// (>=4x). If every point before it sits below dropoutRatio of that first
	// stable price, and there are few enough of them to be noise rather than
	// history, the whole prefix is placeholder noise.
	keep := make([]int, 0, n)
	for i := range pts {
		if !bad[i] {
			keep = append(keep, i)
		}
	}
	for k := 0; k+1 < len(keep); k++ {
		lo, hi := pts[keep[k]].Close, pts[keep[k+1]].Close
		if hi < lo/dropoutRatio {
			continue // no >=4x jump here yet
		}
		allLow := k+1 < minScaleSegment
		for j := 0; allLow && j <= k; j++ {
			if pts[keep[j]].Close > dropoutRatio*hi {
				allLow = false
			}
		}
		if allLow {
			for j := 0; j <= k; j++ {
				bad[keep[j]] = true
			}
		}
		break // only the first jump can bound a leading run
	}

	anyBad := false
	for _, b := range bad {
		if b {
			anyBad = true
			break
		}
	}
	if !anyBad {
		return pts
	}
	out := make([]Point, 0, n)
	for i, p := range pts {
		if !bad[i] {
			out = append(out, p)
		}
	}
	return out
}

// Round-trip cleaning constants. A candidate must satisfy ALL of them, which is
// what makes the pass safe to run on every series pofo fetches.
const (
	// roundTripZ: each leg must exceed this many standard deviations of the
	// LOCAL return distribution. Six is the same bar simgen's despike uses on
	// index shapes, and for the same reason: a crash raises the local sigma
	// with it, so the days that feel extreme to a reader (2020-03-12 at -9.5 %
	// against a ~4.3 % local sigma, barely two sigmas) never come close.
	roundTripZ = 6.0
	// roundTripWindow: how many returns on each side define "local". Five weeks
	// of trading either way is long enough for a stable sigma and short enough
	// to follow a regime change.
	roundTripWindow = 25
	// roundTripNet: how completely the excursion must undo itself. Two percent
	// over the two sessions is the whole tolerance, so a move that merely
	// bounced does not qualify: 1987-10-19 fell 20.5 % and rebounded 5.3 %, and
	// the 2001-09-24 print in the MSCI World shape gains 13.1 % then gives back
	// 7.1 %, leaving 5.0 % standing. Both stay, deliberately, for the same
	// reason: what does not fully reverse may be history.
	roundTripNet = 0.02
)

// dropRoundTrips removes an interior print that leaves the level and comes back
// in the next session, when three independent tests all say no asset could have
// done it: the two legs point opposite ways and cancel to within roundTripNet;
// each exceeds roundTripZ local standard deviations, the suspect pair excluded
// from that estimate; and each exceeds band.SpikeLeg, the class-level floor
// below which a round trip is just volatility. The dropped point's neighbours
// then meet directly, which carries the true two-session move.
//
// Four such prints in the bundled catalog, each confirmed against a sibling
// listing of the same fund:
//
//	IE00B5M4WH52  2015-07-01  +21.9 % / -17.6 %   EM local govt bond
//	IE00B2NPKV68  2017-06-27  -10.3 % / +11.0 %   EM hard-currency bond
//	IE00B8GF1M35  2012-10-29  -22.6 % / +29.7 %   global real estate
//	BTAL          2015-04-29  -32.8 % / +48.8 %   market-neutral anti-beta
//
// And the deliberate abstentions, all of them real market history: 1987-10-19
// and its rebound, the March 2020 sessions (2020-03-12/13 and 2020-03-16),
// October 2008, ERNA's -3.2 % ultrashort credit dislocation on 2020-03-19, and
// the 2001-09-24 pair in the MSCI World shape. The last one is instructive: it
// looks exactly like a bad print and IS one, but its legs leave 5 % standing,
// so this pass declines it and simgen's shape-only despike, which may be laxer
// because it judges an index rather than an investor's NAV, takes it instead.
//
// Rate symbols never reach this pass (see cleanQuotes): a policy rate crossing
// zero and back produces ratios that mean nothing.
func dropRoundTrips(pts []Point, band Band) []Point {
	n := len(pts)
	if n < 2*roundTripWindow {
		return pts
	}
	leg := band.SpikeLeg()
	ret := func(i int) float64 { return pts[i].Close/pts[i-1].Close - 1 }
	out := make([]Point, 0, n)
	out = append(out, pts[0])
	for i := 1; i+1 < n; i++ {
		if pts[i-1].Close <= 0 || pts[i].Close <= 0 || pts[i+1].Close <= 0 {
			out = append(out, pts[i])
			continue
		}
		r1, r2 := ret(i), ret(i+1)
		if r1*r2 >= 0 || math.Abs(r1) < leg || math.Abs(r2) < leg {
			out = append(out, pts[i])
			continue
		}
		if math.Abs((1+r1)*(1+r2)-1) > roundTripNet {
			out = append(out, pts[i])
			continue
		}
		sigma, ok := localSigma(pts, i)
		if !ok || math.Abs(r1) <= roundTripZ*sigma || math.Abs(r2) <= roundTripZ*sigma {
			out = append(out, pts[i])
			continue
		}
		// A print no asset made: drop it and let its neighbours meet.
	}
	out = append(out, pts[n-1])
	return out
}

// localSigma is the standard deviation of the returns around index i, with the
// suspect pair (the returns into and out of i) excluded so a bad print cannot
// inflate the very yardstick meant to convict it. ok is false when too few
// returns surround i for the estimate to mean anything.
func localSigma(pts []Point, i int) (float64, bool) {
	var sum, sumsq float64
	count := 0
	for j := max(1, i-roundTripWindow); j <= min(len(pts)-1, i+roundTripWindow); j++ {
		if j == i || j == i+1 || pts[j-1].Close <= 0 || pts[j].Close <= 0 {
			continue
		}
		r := pts[j].Close/pts[j-1].Close - 1
		sum += r
		sumsq += r * r
		count++
	}
	if count < 10 {
		return 0, false
	}
	mean := sum / float64(count)
	variance := sumsq/float64(count) - mean*mean
	if variance <= 0 {
		return 0, false
	}
	return math.Sqrt(variance), true
}

// fxSpikeRatio is the minimum adjacent move treated as a candidate bad print
// in a currency cross. Major crosses close within a fraction of a percent of
// the prior day; even their historic extremes (the 2015 CHF depeg, -17%, or
// the 2016 GBP flash crash, ~-2% close-to-close) either far exceed it without
// reverting or stay well below it, so an 8% spike that immediately cancels is
// a provider artefact, not a market move.
const fxSpikeRatio = 0.08

// dropFXSpikes removes isolated bad prints from a currency cross: an interior
// point whose move in exceeds fxSpikeRatio, whose move out goes the opposite
// way, and whose two moves nearly cancel (the two-day net is at most a quarter
// of the spike). Real currency shocks fail the cancellation test: a
// devaluation does not un-happen the next day. The real case: Yahoo printing
// EURUSD=X at 1.4918 on 2008-12-08 between 1.2717 and 1.2926 (+17%/−13%),
// which leaked a ±15% one-day artefact into every EUR-converted series.
// Only applied to "=X" symbols; equities legitimately whipsaw harder.
func dropFXSpikes(pts []Point) []Point {
	n := len(pts)
	if n < 3 {
		return pts
	}
	out := make([]Point, 0, n)
	out = append(out, pts[0])
	for i := 1; i+1 < n; i++ {
		prev, cur, next := pts[i-1].Close, pts[i].Close, pts[i+1].Close
		if prev <= 0 || cur <= 0 || next <= 0 {
			out = append(out, pts[i])
			continue
		}
		a, b := cur/prev, next/cur
		spike := math.Max(math.Abs(a-1), math.Abs(b-1))
		if spike >= fxSpikeRatio && (a-1)*(b-1) < 0 && math.Abs(a*b-1) <= spike/4 {
			continue // an isolated spike that cancels: a bad print
		}
		out = append(out, pts[i])
	}
	out = append(out, pts[n-1])
	return out
}
