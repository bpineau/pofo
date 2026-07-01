package marketdata

// dropoutRatio is the fraction below which a close, relative to the surrounding
// stable level, is treated as a data dropout (a provider placeholder or a bad
// print) rather than a real quote. 0.25 means a >75% one-day collapse that
// recovers 4x the next day, or a leading value below a quarter of the first
// real price: both are physically impossible for a genuine close-to-close move,
// so removing them never discards real data.
const dropoutRatio = 0.25

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
