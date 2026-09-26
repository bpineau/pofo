package simgen

import (
	"fmt"
	"os"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// longBack maps a short-history component to a longer real proxy whose
// (rescaled) history is spliced in before the component's own inception, so a
// reconstruction covers each period with the most reliable real series
// available for it. Every proxy is a bundled refdata series (go:embed via
// datasets.Refdata), so the splices work offline and reproducibly; Yahoo's
// long MSCI index symbols (^990300-USD-STRD etc.) return nothing to the client
// and are deliberately not used. All proxies are total-return / spot levels in
// USD, homogeneous with the component they extend:
//
//   - VTMGX (Vanguard Developed Markets, 1999) → developed-ex-US equity TR
//     (refdata DEVEXUS-USD: MSCI World ex USA net TR via Curvo, monthly
//     month-END levels ~1969, carried at daily granularity from 1990-07 by the
//     Ken French daily shape, see dailyShape and alignMonthEnd).
//   - VEIEX (Vanguard Emerging Markets, 1994) → emerging-market equity TR
//     (refdata EM-USD: MSCI Emerging Markets net TR via Curvo, monthly
//     month-END levels, ~1988; no daily shape, the leg stays monthly there).
//   - GC=F (COMEX gold futures, 2000) → daily London/LBMA PM gold fix
//     (refdata XAUUSD-LBMA, ~1968).
//   - CL=F (NYMEX WTI futures, 2000) → monthly WTI spot (refdata WTI-USD, ~1946).
//   - VFITX (Intermediate-Term Treasury, 1991) and VUSTX (Long-Term, 1986) →
//     constant-maturity Treasury total-return reconstructions (refdata
//     TREASURY-INT-USD / TREASURY-LONG-USD: a 5-year and a 20-year par bond on
//     the Fed's H.15 constant-maturity yields, month-END levels from 1953-04,
//     see cmd/gen-tyield-refdata), carried at daily granularity from 1962 by the
//     TREASURY-*-DAILY daily-yield shapes (see dailyShape and alignMonthEnd).
//   - VFINX (Vanguard 500, 1976) → S&P 500 total return (refdata SP500-USD:
//     month-end levels from the ^SP500TR index 1988->, ^GSPC + Shiller dividend
//     1928-1988, Shiller 1871-1928; see cmd/gen-sp500-refdata), the index VFINX
//     tracks, carried at daily granularity from 1927-12 by the ^GSPC price shape.
//   - ^IRX (13-week T-bill rate) → the 3-month T-bill rate (refdata TBILL-3M:
//     FRED TB3MS, ~1934). A rate, not a price: rescaled by a ≈1 factor at the
//     splice (^IRX ≈ TB3MS there), then read as an isRate series.
//   - GBPUSD=X (Yahoo, ~2003) → the real daily FRED noon GBP/USD rate
//     (refdata GBPUSD-DAILY: DEXUSUK, 1971→), for the GBP-quoted recipes.
//   - DFSVX (DFA US Small Cap Value, 1993) → US small-cap value TR (refdata
//     USSCV-USD: Ken French value-weighted SMALL HiBM daily, cumulated, from
//     1963-07), the size×value factor behind ZPRV/USSC. Real daily total-return
//     levels, so no daily shape is needed, but a GROSS one: see longBackFee.
//   - VTSMX (Vanguard Total Stock Market Investor, 1992-04) → US total market TR
//     (refdata USMKT-USD: the CRSP value-weighted market factor, Mkt-RF + RF,
//     cumulated daily from 1926-07; see cmd/gen-usmkt-refdata), the whole market
//     behind VTI. Daily and total-return already, and GROSS: see longBackFee.
//   - EUNH.DE (iShares Core Euro Govt Bond, 2009) → euro-area government bond TR
//     (refdata EUROGOV-EUR: the ECB daily 10y yield-curve point through
//     TreasuryTR, sampled month-END from 2004-09, and the OECD euro-area 10y
//     yield the same way before it, ~1970), carried at daily granularity from
//     2004 by that same curve as a shape (EUROGOV-DAILY). The bond leg of the
//     eurozone NTSZ recipe.
var longBack = map[string]string{
	"VTMGX":    "DEVEXUS-USD",
	"VEIEX":    "EM-USD",
	"GC=F":     "XAUUSD-LBMA",
	"CL=F":     "WTI-USD",
	"VFITX":    "TREASURY-INT-USD",
	"VUSTX":    "TREASURY-LONG-USD",
	"VFINX":    "SP500-USD",
	"VTSMX":    "USMKT-USD",
	"^IRX":     "TBILL-3M",
	"GBPUSD=X": "GBPUSD-DAILY",
	"DFSVX":    "USSCV-USD",
	"EUNH.DE":  "EUROGOV-EUR",
}

// longBackFee is what a proxy owes before it may stand in for a fund: a
// constant fraction per year taken off its returns, applied to the proxy alone
// and never to the component's own quotes. Most proxies owe nothing, because
// they are already what an investor received (another fund's NAV, a total-return
// index a tracker exists for). One is an ACADEMIC FACTOR, and a factor pays no
// fee, no commission and no spread, in the one corner of the equity market where
// those are largest.
//
// USSCV-USD, the Ken French small-value portfolio, is that one. Its overstatement
// is measurable rather than assumed, because it and the fund it extends overlap
// for thirty-three years, and it is small: over 1993-03 to 2026-05 (399 months)
// the factor compounds at 12.51 %/yr against DFSVX's 11.49 %, a gap of
// +1.02 pts/yr with a standard error of 0.69, at a monthly correlation of 0.980
// and a volatility ratio of 1.002. The two are the same trade, minus a wrapper.
// Against the target fund's own quotes (ZPRV, 2015-2026) the factor runs
// +0.38 pts/yr hot, and against DFA's Targeted Value fund -0.07, so the
// correction is bounded by the universe it is measured on. Adopted: 1.0 %/yr,
// the full-overlap figure rounded, since neither the standard error nor the era
// justifies a second decimal.
//
// Two caveats belong with the number, and both point the same way. The gap
// decays across the overlap, +1.89 pts/yr over the first half and +0.16 over the
// second (per decade: 1990s +2.62, 2000s +1.30, 2010s +0.32, 2020s +0.02), which
// is what falling commissions and spreads look like. The repo's own stability
// criterion does not accept that as established, the swing of 1.73 sitting inside
// the 1.39 standard error of the difference, so the constant stays the
// full-overlap one rather than a fitted half. But it means the deep segment, all
// of which predates the overlap and most of which predates the end of fixed
// commissions in 1975, is corrected by a floor and not by its own cost: if the
// trend is real, the 1963-1993 tail is still a touch rich. The other direction
// is covered too: the price lists alone (DFSVX charged 0.31 to 0.53 %/yr over
// the overlap, ZPRV charges 0.30 %) put a hard lower bound near 0.4 %/yr on the
// gap, so a zero haircut is known to be wrong whatever the sampling error says.
//
// Grossness is a level error, not a cadence error, and the file does not have a
// cadence problem: the ratio of daily-annualized to monthly-annualized
// volatility runs 0.54 to 0.66 through the tail, and DFSVX's own NAVs show the
// same 0.68 in the one decade they share (both rise above 1 in the 2000s). Thin
// trading in small caps is the era, not the reconstruction, so nothing here is
// projected onto another calendar the way a weekly-dealing donor is.
// It is exported because the same series stands behind figures outside this
// package (the FIRE book's small-value plates read USSCV-USD directly), and a
// measured constant must have exactly one home: charge this one, never derive
// a second.
const USSCVGrossCost = 0.010

// usmktGrossCost is the same correction for USMKT-USD, the CRSP total-market
// factor, and it is small for the reason USSCV's is large: the whole market
// cap-weighted is the cheapest portfolio there is to hold, where small value
// is the dearest. Measured the same way, on the 411 months the factor shares
// with the fund it extends (VTSMX, 1992-04 to 2026-07), it compounds at
// 11.02 %/yr against the fund's 10.73, a gap of +0.29 pts/yr at a monthly
// correlation of 0.9992. Adopted: 0.3 %/yr, the full-overlap figure rounded.
//
// About half of that is the donor's own price list (VTSMX charged 0.20 %/yr in
// the 1990s and charges 0.14 now; the target ETF charges 0.03), the rest being
// commissions, spreads and the CRSP tail of micro caps no fund replicates
// share for share. The overlap does not identify the split and nothing here
// pretends otherwise: what a proxy owes is what a holder of the fund did not
// receive. Per decade the gap runs +0.80 (1992-1999), -0.12 (2000s), +0.30
// (2010s), +0.56 (2020s), a swing no stability criterion would accept as a
// trend, so the constant stays the full-overlap one rather than a fitted era.
// Unexported, unlike USSCVGrossCost: nothing outside this package reads the
// total-market factor directly.
const usmktGrossCost = 0.003

var longBackFee = map[string]float64{
	"USSCV-USD": USSCVGrossCost,
	"USMKT-USD": usmktGrossCost,
}

// dailyShape maps a monthly longBack proxy to a daily series of the same
// market whose LEVELS are not authoritative (a close but not identical
// universe, gross of withholding, or a price index without income) but
// whose day-to-day moves are real. The proxy's monthly anchors keep setting
// the levels and the shape supplies the daily granularity in between (see
// shapedSeries), so reconstructions stop feeding month-sized moves to
// daily-frequency statistics. A shape may stop before the anchors' end:
// real fund quotes take over from their inception anyway, and shapedSeries
// keeps the later anchors at their own cadence.
var dailyShape = map[string]string{
	"DEVEXUS-USD":       "DEVEXUS-DAILY",       // Ken French developed-ex-US market TR, daily 1990-07→
	"SP500-USD":         "^GSPC",               // S&P 500 daily price index (Yahoo, 1927-12→)
	"TREASURY-INT-USD":  "TREASURY-INT-DAILY",  // FRED DGS5 daily 5y CMT through TreasuryTR, 1962→1992
	"TREASURY-LONG-USD": "TREASURY-LONG-DAILY", // FRED DGS20 daily 20y CMT through TreasuryTR, 1962→1986
	"WTI-USD":           "WTI-DAILY",           // FRED DCOILWTICO daily WTI spot, 1986→2000
	"EUROGOV-EUR":       "EUROGOV-DAILY",       // ECB daily euro-area 10y yield through TreasuryTR, 2004→ (also the anchors' own source there)
	"EUROGOV-LONG-EUR":  "EUROGOV-LONG-DAILY",  // ECB daily euro-area 25y yield through TreasuryTR, 2004→ (idem)
}

// tracked names the donors whose provider line can be graded against a bundled
// series tracking the SAME instrument, with the reference and the tolerance
// measured for that pair (see trackIndex). The grading happens before any
// splice, so what is judged is the donor's own quotes.
//
// VFITX is here because its adjusted close carries an unhandled year-end
// distribution. On 1993-12-31 the fund's raw NAV falls 3.60 % and the provider
// reports 0.051 of dividend, 0.46 % of NAV, so the adjusted series steps down
// 3.16 % and never recovers: a permanent level error in a donor that stands
// behind IEF, NTSX, RSSB, RSBT and NTSG. The same day the longer-duration
// sibling VUSTX ROSE 0.52 % and the shorter VFISX fell 0.19 %, and the 5-year
// constant maturity moved 0.03 pt. The missing amount is a capital-gains
// distribution the provider does not publish, so it cannot be added back from
// the source; what can be done is refuse the session and take the
// reconstruction's return over the same two dates.
//
// The tolerance is 1.5 %, measured over the fund's whole life (8 713 sessions
// against a 5-year par bond on the H.15 5-year point, the reference allowed to
// lead or lag one session): the defect stands at 2.88 % and the largest session
// with nothing behind it at 0.84 %, so 1.5 % sits near the middle of an empty
// band, a factor of 1.8 from either side. It is two orders of magnitude tighter
// than the equity-ETF tolerance of the FCPE recipe, which is why trackIndex
// takes the tolerance as an argument rather than owning one.
//
// VFINX carries the same signature on nine Decembers of 1980-1986 (1980-12-30,
// 1981-12-29, 1982-12-28, 1983-12-28, 1984-12-28, 1985-12-27, 1986-12-09, plus
// the 1981-04-20/21 round trip), the raw NAV falling far more than the
// distribution the provider reports: -7.22 % on 1986-12-09 against the S&P
// 500's -0.75, about 30 % of cumulative level lost, i.e. every recipe whose
// equity leg it carried over 1980-1987 ran some 4 points a year cold there (the
// S&P 500 tracker IE00BFMXXD54 read 11.45 %/yr against the index's 15.88). It
// is graded against the S&P 500 PRICE index ^GSPC, the index the fund tracks
// (a day's dividend is two orders of magnitude under the tolerance). Measured
// over 11 773 sessions with the stale-print allowance of trackIndex, the
// largest disagreement with no defect behind it is 0.93 % (the repeated close
// of 1987-11-27) and the smallest defect 1.37 % (1982-12-28); 1.15 % sits in
// the middle of that empty band. A first measurement, against the CRSP
// total-market factor rather than the S&P 500, found no band at all (the
// 1987-10-19 crash reads 3.77 % of honest excess there, the size premium's own
// move): the reference has to be the index the fund holds.
//
// One neighbour was measured and is deliberately NOT here, because its
// separating band closes:
//
//   - VUSTX carries a contaminated patch of its own, 1992-12-11 (-5.69 %
//     against the long reconstruction's -0.18) to 1992-12-31 (+6.76 % against
//     -0.08, on a day its raw NAV moved -0.10 % and the provider credited a
//     0.632 distribution). Both clear 5.7 %, but 1987-10-22 is a REAL +7.76 %
//     against the long curve point's +2.64, i.e. 4.86 % of honest excess, and
//     no tolerance separates 4.86 from 5.69. Worse, the H.15 20-year point is
//     suspended over 1987-01..1993-09, so the daily reference does not even
//     cover the days in question.
var tracked = map[string]struct {
	ref string
	tol float64
}{
	"VFITX": {"TREASURY-INT-DAILY", 0.015},
	"VFINX": {sp500ShapeID, 0.0115},
}

// extendingFetcher wraps a Fetcher so that a configured component is spliced
// with a longer proxy (marketdata.ExtendBack) at fetch time. A missing or empty
// proxy is skipped silently, leaving the component unchanged, so the wrapper is
// safe to apply unconditionally.
type extendingFetcher struct {
	inner Fetcher
	back  map[string]string
}

// Fetch fetches id and, when a longer proxy is configured for it, prepends the
// proxy's rescaled history before the component's first quote. A donor listed
// in tracked is first held to the reference that grades it.
func (e extendingFetcher) Fetch(id string, from time.Time) (*marketdata.Series, error) {
	s, err := e.inner.Fetch(id, from)
	if err != nil {
		return nil, err
	}
	if t, ok := tracked[id]; ok && s != nil {
		if ref, rerr := e.inner.Fetch(t.ref, from); rerr == nil && ref != nil {
			repaired, rejects := trackIndex(s, ref, t.tol)
			for _, r := range rejects {
				fmt.Fprintf(os.Stderr, "extend: %s does not track %s on %s\n", id, t.ref, r)
			}
			s = repaired
		} else {
			fmt.Fprintf(os.Stderr, "extend: %s: reference %s unavailable (%v), the donor is used ungraded\n", id, t.ref, rerr)
		}
	}
	pid, ok := e.back[id]
	if !ok || s == nil {
		return s, err
	}
	// Diagnostics to stderr (gen-simdata only): a silent skip hides why a
	// backcast did not lengthen, so report the proxy fetch and splice outcome.
	// alignMonthEnd before shaping, exactly as shapedIndex does: a month-END
	// proxy dated on a weekend would otherwise pin into the next month and slide
	// the whole spliced era (DEVEXUS-USD rebuilt 2022 at -4.3 % against the
	// index's -14.3 % that way).
	p, perr := e.inner.Fetch(pid, from)
	if sid, shaped := dailyShape[pid]; shaped && perr == nil && p != nil {
		if sh, serr := e.inner.Fetch(sid, from); serr == nil {
			p = shapedSeries(alignMonthEnd(pid, p, sh), sh)
		}
	}
	if fee, owed := longBackFee[pid]; owed && p != nil {
		p = afterAnnualFee(p.Name+" (fee-aligned)", p, fee)
	}
	switch {
	case perr != nil:
		fmt.Fprintf(os.Stderr, "extend: %s: proxy %s fetch failed: %v\n", id, pid, perr)
	case p == nil || len(p.Points) < 2:
		fmt.Fprintf(os.Stderr, "extend: %s: proxy %s returned no usable history\n", id, pid)
	case marketdata.ExtendBack(s, p):
		fmt.Fprintf(os.Stderr, "extend: %s extended with %s back to %s\n", id, pid, s.Points[0].Date.Format("2006-01-02"))
	default:
		fmt.Fprintf(os.Stderr, "extend: %s: proxy %s (from %s) added no earlier data than %s\n",
			id, pid, p.Points[0].Date.Format("2006-01-02"), s.Points[0].Date.Format("2006-01-02"))
	}
	return s, nil
}

// extend wraps a fetcher with the standard long-history proxy splices. The
// shared composite and tsmom builders route their component fetches through it,
// so every reconstruction using an extendable component reaches further back.
func extend(f Fetcher) Fetcher { return extendingFetcher{inner: f, back: longBack} }
