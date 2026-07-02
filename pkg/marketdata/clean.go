package marketdata

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

// mendScaleBreak repairs a single large, persistent denomination break: when a
// series has EXACTLY ONE adjacent jump beyond scaleBreakFactor with a
// substantial segment on both sides, the older segment is rescaled onto the
// newer one (the recent segment is the current NAV, hence authoritative). The
// real IBGS.L / ITPS.L case: their pre-2009 history sits at ~100x the true NAV,
// a -99% cliff at the junction; after mending the whole series is continuous.
//
// It is deliberately timid: no break, several breaks (a spliced share class like
// CL2.PA), a reversed round-trip (GRE), or a too-short side are all left
// untouched for the -verify-data doctor to flag, because rescaling them
// automatically could corrupt genuine data. Run it AFTER dropDropouts, so
// leading placeholders and isolated dropouts are already gone.
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
		if r := pts[i].Close / pts[i-1].Close; r >= scaleBreakFactor || r <= 1/scaleBreakFactor {
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
	for i := 0; i < breakIdx; i++ {
		pts[i].Close *= f
	}
	return pts
}

// dropDropouts removes obvious bad prints from a sorted daily series:
//
//   - a leading run of placeholder points before the price first jumps up (>=4x)
//     to its stable level, e.g. Yahoo emitting a tiny value at an ETF inception
//     (the real IB01 case: closes of 5 before the true ~99 NAV, which otherwise
//     shows as a +1884% daily move);
//   - an isolated interior point below dropoutRatio of BOTH neighbours (a one-day
//     collapse that immediately recovers).
//
// It deliberately keeps everything else: moderate spikes (they may be real
// distributions or genuine volatility), permanent declines (a low tail that
// never recovers), and gradual growth from a low base (which never quadruples in
// a single day). The -verify-data doctor still surfaces the moderate moves for
// human review; this only strips the unambiguous artefacts.
func dropDropouts(pts []Point) []Point {
	n := len(pts)
	if n < 3 {
		return pts
	}
	bad := make([]bool, n)

	// Leading placeholder run: find the first adjacent jump up of at least
	// 1/dropoutRatio (>=4x). If every point before it sits below dropoutRatio of
	// that first stable price, the whole prefix is placeholder noise.
	for k := 0; k+1 < n; k++ {
		if pts[k+1].Close < pts[k].Close/dropoutRatio {
			continue // no >=4x jump here yet
		}
		allLow := true
		for i := 0; i <= k; i++ {
			if pts[i].Close > dropoutRatio*pts[k+1].Close {
				allLow = false
				break
			}
		}
		if allLow {
			for i := 0; i <= k; i++ {
				bad[i] = true
			}
		}
		break // only the first jump can bound a leading run
	}

	// Isolated interior dropouts: a point far below both its neighbours.
	for i := 1; i+1 < n; i++ {
		if bad[i-1] || bad[i+1] {
			continue
		}
		if pts[i].Close < dropoutRatio*pts[i-1].Close && pts[i].Close < dropoutRatio*pts[i+1].Close {
			bad[i] = true
		}
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
