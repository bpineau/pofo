// Command gen-tyield-refdata builds the bundled US Treasury references: the
// LONG par-yield history the zero-coupon STRIPS reconstruction is priced off
// (pkg/simgen.TreasuryZeroTR, behind the ZROZ recipe) and the two
// constant-maturity total-return series the Vanguard Treasury donors are
// extended with. It runs at data-generation time only (network); the pofo
// binary embeds the CSVs and never fetches the Federal Reserve.
//
// Why a yield series and not only total-return series. A coupon bond's duration
// shrinks as its yield rises while a zero's does not, so no fixed multiple of a
// coupon fund reproduces a 25+ year STRIPS fund across rate regimes: the
// multiple that fits the 2 to 4 % era is half the one the 12 % era needs. The
// honest input for a strip is the long yield itself, and a bundled yield lets
// the maturity and the fee stay parameters of the recipe that reads it rather
// than constants frozen into a CSV. The two total-return series are then built
// here as well, off the same H.15 points and by the same engine, so that one
// command owns the whole family and a refresh can never leave them describing
// different curves.
//
// Five files are written into pkg/datasets/refdata/:
//
//   - TREASURY-LONG-YIELD.csv  US long Treasury constant-maturity par yield,
//     annualized percent (1953-04 →). Business-daily
//     from 1962-01-02, monthly month-end before.
//   - TREASURY-LONG-USD.csv    total return of a 20-year par Treasury priced
//     off that long yield, month-end, 0.10 %/yr (1953-04 →). The deep proxy
//     behind VUSTX.
//   - TREASURY-INT-USD.csv     total return of a 5-year par Treasury priced off
//     the H.15 5-year point, month-end, 0.10 %/yr (1953-04 →). The deep proxy
//     behind VFITX, and the bond sleeve of the bundled US 60/40 (pkg/replay).
//   - TREASURY-LONG-DAILY.csv  the same 20-year par bond on the published
//     20-year point, daily, no fee (1962-01 → 1986-12, where that point is
//     discontinued). Its LEVELS are not authoritative: it is the intra-month
//     daily shape behind TREASURY-LONG-USD's anchors (simgen dailyShape).
//   - TREASURY-INT-DAILY.csv   the same for the 5-year point, daily
//     (1962-01 →). Besides shaping TREASURY-INT-USD it is the same-index
//     reference the Vanguard intermediate-Treasury donor is graded against
//     (simgen trackIndex), which is why it now runs to the present rather than
//     stopping where that donor's own quotes begin.
//
// The two daily files had no generator at all until 2026-09-20 and could not be
// refreshed; they are here now because the family must be refreshable as one,
// and because one of them carried the definitional break described next.
//
// # The 1973 definitional break
//
// H.15's 20-year point moved 6.04 → 6.78 between 1973-01-03 and 1973-01-04 and
// stayed there, while the 10-year point sat still (6.42 → 6.40) and the 5-year
// one too (6.26 → 6.23). It is not a market move: before that day the 20-year
// point ran on average 0.13 pt BELOW the 10-year over 1962-1972, and after it
// 0.13 pt above over 1973-1986, the sign of the term spread flipping overnight.
// The 0.74 pt step is three times the largest daily move the 20-year point
// makes anywhere else in its fifteen years of business-daily history before the
// spine takes over.
//
// The published levels are kept on BOTH sides, because both are what the Fed
// published and the pre-1973 level is independently right: over 1954-1972 a
// 20-year par bond on this series compounds at 2.49 %/yr against the published
// SBBI long-government record's 2.24, every calendar year within a few points.
// Shifting the head to close the step was tested and refused: it would move a
// nineteen-year stretch that already agrees with its reference in order to
// repair one day.
//
// What is refused instead is the RETURN across that one day. 1973-01-04 is
// declared a segment junction, exactly like the 1977-02-15 one, and travels in
// the file's own "# junctions:" header, so every consumer of the yield series
// skips the step rather than pricing it (marketdata.Series.Junctions,
// simgen.TreasuryTR). Without it the bundled 1973 read -6.78 % against SBBI's
// published -1.11 %, and ZROZ recorded its worst day ever, -17.0 %, on a day
// nothing happened.
//
// The yield series is assembled from the Federal Reserve's H.15 selected
// interest rates, read through the DBnomics mirror (free, key-less), in three
// segments:
//
//   - the spine, 1977-02-15 →: the 30-year constant maturity, business daily
//     (RIFLGFCY30_N.B). It is the longest point the Fed publishes and the
//     closest one to what a 25+ year STRIPS fund holds. The H.15 row is
//     CONTINUOUS across the 2002-02 to 2006-02 suspension of the 30-year
//     issue, when the published figure is the Treasury's long-term (25 years
//     and above) average rather than a 30-year bond; those four years are
//     checked against the 20-year point before anything is written, since a
//     substituted definition is exactly the kind of change that ships
//     unnoticed.
//   - 1962-01-02 to 1977-02-14: the 20-year constant maturity, business daily
//     (RIFLGFCY20_N.B), mapped onto the spine's curve point.
//   - 1953-04 to 1961-12: the 20-year constant maturity, monthly
//     (RIFLGFCY20_N.M), the same point at the only cadence that reaches that
//     far back, mapped the same way.
//
// The map is affine, y30 = a + b·y20, fitted by least squares on the ten
// thousand-odd days the two points overlap, then shifted so the mapped series
// MEETS the spine exactly at the junction (the same construction the euro-area long
// sleeve uses for its pre-2004 years). It is not a cosmetic adjustment: the
// long end of the curve is not parallel to the 20-year point, it was INVERTED
// through the high-rate years (the 30-year ran 0.15 below the 20-year when the
// 20-year was above 9 %, and 0.11 above it below 5 %), which a constant spread
// would get backwards in exactly the era this series exists to cover.
//
// The two total-return series are MONTH-END, one point per calendar month, and
// GAP-FREE from 1953-04.
//
// Month-end matters and cost this repository a correlation. Both files used to
// be built from FRED's MONTHLY series (GS20, GS5), which are the month's
// AVERAGE of the daily yields stamped on the first of the month: the smear puts
// every turning point half a month early and the label puts it another half
// month early again, and the monthly returns that come out correlate 0.70 with
// the real funds they stand behind (0.97 once the yields are sampled at each
// month's last quote). At the calendar-year frequency the two conventions are
// indistinguishable, which is why the flaw survived the yearly goldens for
// years. The pre-1962 head has no daily point to sample and stays a month
// AVERAGE, carried at its month-end label.
//
// Gap-free matters more. H.15 discontinued the 20-year constant maturity
// between 1987-01 and 1993-09, so the old long file simply had no row there and
// bridged seven years in a single step: a +60 % "month", no volatility for
// seven years, and a book plate that had to rebuild its own long leg to avoid
// it. The 30-year point publishes straight through that window, so pricing the
// long bond off the assembled long yield removes the hole by construction.
//
// Every check runs before anything is written (-check, on by default): era
// level bands, the 1977 junction's continuity, the 1973 definitional break
// still being where it is declared, the substituted 2002-2006 window, the
// absence of a frozen fetch, freshness, a cross-check of the 20-year par
// reconstruction the long yield implies against the daily 20-year one, and, on
// each total-return series, one point per month with no hole, plus the
// published Ibbotson SBBI calendar-year returns and the long-run CAGR the
// goldens assert. Nothing here is trusted on the strength of having downloaded
// cleanly.
//
// Usage: gen-tyield-refdata [-base URL] [-dir path] [-dry] [-check=false]
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
	"github.com/bpineau/pofo/pkg/simgen"
)

const (
	defaultBase = "https://api.db.nomics.world/v22"
	// The five bundled series this generator owns.
	outID         = "TREASURY-LONG-YIELD"
	outLong       = "TREASURY-LONG-USD"
	outInter      = "TREASURY-INT-USD"
	outLongDaily  = "TREASURY-LONG-DAILY"
	outInterDaily = "TREASURY-INT-DAILY"
	// The five H.15 series read, all of them "U.S. government securities /
	// Treasury constant maturities / Nominal".
	spine30D = "FED/H15/RIFLGFCY30_N.B"
	head20D  = "FED/H15/RIFLGFCY20_N.B"
	head20M  = "FED/H15/RIFLGFCY20_N.M"
	int5D    = "FED/H15/RIFLGFCY05_N.B"
	int5M    = "FED/H15/RIFLGFCY05_N.M"
	// neigh10D is read for the checks only: the 10-year point is the neighbour
	// that did NOT move across the 1973 definitional break, which is what makes
	// the break definitional.
	neigh10D = "FED/H15/RIFLGFCY10_N.B"
	// substFrom/substTo bound the window where H.15's 30-year row carries the
	// long-term (25 years and above) average instead of a 30-year bond.
	substFrom, substTo = "2002-02-19", "2006-02-08"

	// defBreak20 is the day the 20-year point changed definition (see the
	// package comment): 6.04 → 6.78 overnight with the 10-year point still. The
	// LEVELS on both sides are published and kept; the RETURN across the step is
	// refused, here and by every consumer, through the "# junctions:" header.
	defBreak20 = "1973-01-04"
	// defBreakMin is the smallest step that day may show before the declaration
	// stops describing the data. The measured step is 0.713 pt on the assembled
	// series (0.74 on the raw 20-year point, shrunk by the affine map's slope);
	// were a revision to close it, the junction would be forgiving a real move
	// and would have to go.
	defBreakMin = 0.30

	// longDailyEnd is where the daily 20-year shape stops: H.15 discontinued the
	// 20-year constant maturity at the end of 1986 and only reinstated it in
	// 1993-10. The shape is not carried across that hole, since what it exists
	// to supply (the intra-month texture of the years before the long funds
	// quote) ends there anyway.
	longDailyEnd = "1986-12-31"

	// The two reconstructions' parameters, both a priori. 20 years is the
	// maturity of the long-term government bond the published SBBI record
	// tracks and the neighbourhood of what the fund this series stands behind
	// (VUSTX) holds; 5 years is the curve point the intermediate one is read
	// at, and the fund it stands behind (VFITX) is a 5-10 year one, so the
	// reconstruction is deliberately the SHORTER of the two. Neither maturity
	// is fitted to a fund's realized volatility. The fee is the one the files
	// have always charged: a holder of a Treasury ladder pays something, and
	// 0.10 %/yr is half of what the Vanguard donors charge.
	longMaturity  = 20.0
	interMaturity = 5.0
	reconFee      = 0.001
)

func main() {
	base := flag.String("base", defaultBase, "DBnomics API base URL")
	dir := flag.String("dir", "pkg/datasets/refdata", "output refdata directory")
	dry := flag.Bool("dry", false, "print coverage and checks without writing")
	check := flag.Bool("check", true, "run the sanity checks before writing")
	flag.Parse()

	spine := fetch(*base, spine30D)
	daily20 := fetch(*base, head20D)
	monthly20 := fetch(*base, head20M)
	daily5 := fetch(*base, int5D)
	monthly5 := fetch(*base, int5M)
	daily10 := fetch(*base, neigh10D)

	a, b, resid, n := fitAffine(daily20, spine)
	log.Printf("30y on 20y: y30 = %+.4f %+.4f*y20  (n=%d days, residual sd %.3f pt)", a, b, n, resid)

	junctions := []time.Time{mustDate(defBreak20)}
	long := assemble(spine, daily20, monthly20, a, b)
	report(outID, "%/yr", long)

	// The two MONTH-END total-return series: one step per calendar month, each
	// on the month's LAST quoted yield, through the shared par-bond engine. The
	// long one also steps on the junction's own boundary days, so that the real
	// moves of the month it falls in are kept and only the splice is skipped;
	// the extra steps are dropped again by monthEndIndex.
	longTR := monthEndIndex(outLong, simgen.TreasuryTR(outLong, monthEndSample(long, junctions), longMaturity, reconFee))
	interTR := simgen.TreasuryTR(outInter, monthEndSample(joinCadences(daily5, monthly5), nil), interMaturity, reconFee)
	report(outLong, "index", longTR.Points)
	report(outInter, "index", interTR.Points)

	// The two DAILY shape series, on the published points themselves rather
	// than on the assembled long yield: their levels are never authoritative,
	// their day-to-day moves are, and those come from the curve point the
	// monthly sibling names.
	longDaily := simgen.TreasuryTR(outLongDaily, withJunctions(points(trimTo(daily20, mustDate(longDailyEnd))), junctions), longMaturity, 0)
	interDaily := simgen.TreasuryTR(outInterDaily, withJunctions(points(daily5), nil), interMaturity, 0)
	report(outLongDaily, "index", longDaily.Points)
	report(outInterDaily, "index", interDaily.Points)

	if *check {
		runChecks(long, spine, daily20, junctions)
		checkJunctionNeighbour(daily10, junctions[0], 0.10)
		checkReconstruction(outLong, longTR, sbbiLong, 7.6, 1.0, 9, 14)
		checkReconstruction(outInter, interTR, sbbiInter, 6.4, 0.8, 3, 7)
		checkShape(outLongDaily, longDaily, longTR, date(1962, 1), date(1987, 1))
		checkShape(outInterDaily, interDaily, interTR, date(1962, 1), date(1993, 1))
		if failed > 0 {
			log.Fatalf("%d sanity check(s) failed, nothing written", failed)
		}
	}
	if *dry {
		return
	}
	write(*dir, outID, "US long Treasury constant-maturity par yield (annualized percent, daily from 1962, monthly before)",
		fmt.Sprintf("Federal Reserve H.15 selected interest rates, Treasury constant maturities, nominal, via DBnomics: the 30-year business-daily point (%s, 1977-02-15->, its row carrying the long-term 25-years-and-above average over the %s..%s suspension of the 30-year issue), extended back by the 20-year business-daily point (%s, 1962-01-02->) and the 20-year monthly point (%s, 1953-04->), both mapped onto the 30-year curve point by the affine fit y30 = %+.4f %+.4f*y20 (least squares over the %d overlapping days, residual sd %.3f pt) and shifted to meet the spine at the junction. RATE levels (annualized percent), NOT a price: read as a yield by simgen.TreasuryZeroTR (the 25+ STRIPS reconstruction behind ZROZ) and by TreasuryTR. DEFINITIONAL BREAK at %s, declared in the junctions header: the 20-year point steps %+.2f pt that day while the 10-year point does not move, and the term spread to the 10-year flips sign (-0.13 pt on average over 1962-1972, +0.13 over 1973-1986). Both published levels are kept; the RETURN across that one step is not a rate move and every consumer skips it.",
			spine30D, substFrom, substTo, head20D, head20M, a, b, n, resid, defBreak20, breakSize(long, junctions[0])), junctions, long)
	write(*dir, outLong, fmt.Sprintf("US long-term Treasury total return (%.0f-year par bond on the long constant-maturity yield, month-end)", longMaturity),
		fmt.Sprintf("month-end samples of the bundled %s run through simgen.TreasuryTR (%.0f-year par bond, %.2f%%/yr). GAP-FREE from 1953-04: H.15 suspended the 20-year constant maturity between 1987-01 and 1993-09 and the 30-year point publishes straight through, so the long yield this is priced off has no hole and neither does this series. The BOND is 20-year throughout (the maturity of the published long-term government bond record, and the neighbourhood of what VUSTX holds); the CURVE POINT it is discounted at is the long one, i.e. the 30-year par yield from 1977-02 and the 20-year point mapped onto it before, which differs from the published 20-year point by %+.4f %+.4f*y20 (about +0.07 pt at a 4%%/yr yield, -0.22 pt at 12%%) and whose moves are %.3f as large. Month-end, one point per calendar month; the pre-1962 head is a month-AVERAGE yield carried at its month-end label, the only cadence H.15 published then. January 1973 is compounded in two pieces around the %s definitional break of the 20-year point, so it carries the real moves on each side of it and not the break itself. Proxy behind VUSTX.",
			outID, longMaturity, reconFee*100, a, b-1, b, defBreak20), nil, longTR.Points)
	write(*dir, outInter, fmt.Sprintf("US intermediate-term Treasury total return (%.0f-year constant maturity, month-end)", interMaturity),
		fmt.Sprintf("Federal Reserve H.15 selected interest rates, Treasury constant maturities, nominal, 5-year, via DBnomics: the business-daily point (%s, 1962-01-02->) sampled at each month's LAST quote, extended back by the monthly point (%s, 1953-04->, a month-AVERAGE yield carried at its month-end label, the only cadence H.15 published then), run through simgen.TreasuryTR (%.0f-year par bond, %.2f%%/yr). One curve point throughout: no map, no splice and no hole. Proxy behind VFITX, and the bond sleeve of the bundled US 60/40 (pkg/replay).",
			int5D, int5M, interMaturity, reconFee*100), nil, interTR.Points)
	write(*dir, outLongDaily, fmt.Sprintf("US long-term Treasury total return (%.0f-year CMT, daily), DAILY SHAPE series", longMaturity),
		fmt.Sprintf("Federal Reserve H.15 20-year constant maturity, nominal, business-daily (%s), via DBnomics, run through simgen.TreasuryTR (%.0f-year par bond, no fee), 1962-01-02 to %s where H.15 discontinues that point. The %s definitional break of the 20-year point is declared a junction and its step is skipped. LEVELS ARE NOT AUTHORITATIVE: used only as the intra-month daily shape behind the %s monthly anchors (simgen dailyShape/anchorShape).",
			head20D, longMaturity, longDailyEnd, defBreak20, outLong), nil, longDaily.Points)
	write(*dir, outInterDaily, fmt.Sprintf("US intermediate-term Treasury total return (%.0f-year CMT, daily), DAILY SHAPE series", interMaturity),
		fmt.Sprintf("Federal Reserve H.15 5-year constant maturity, nominal, business-daily (%s), via DBnomics, run through simgen.TreasuryTR (%.0f-year par bond, no fee), 1962-01-02 onwards. LEVELS ARE NOT AUTHORITATIVE: used as the intra-month daily shape behind the %s monthly anchors (simgen dailyShape/anchorShape) and as the same-index reference the Vanguard intermediate-Treasury donor is graded against (simgen trackIndex), which is why it runs to the present rather than stopping at that donor's inception.",
			int5D, interMaturity, outInter), nil, interDaily.Points)
}

// mustDate parses one of this file's own ISO date constants.
func mustDate(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		log.Fatalf("bad date constant %q: %v", s, err)
	}
	return t
}

// points turns raw observations into dated closes.
func points(obs []obs) []marketdata.Point {
	out := make([]marketdata.Point, len(obs))
	for i, p := range obs {
		out[i] = marketdata.Point{Date: p.date, Close: p.val}
	}
	return out
}

// trimTo drops the observations after the given day.
func trimTo(obs []obs, last time.Time) []obs {
	out := obs[:0:0]
	for _, p := range obs {
		if !p.date.After(last) {
			out = append(out, p)
		}
	}
	return out
}

// withJunctions wraps a yield path as the series the reconstruction engines
// read, carrying the definition junctions they must not price through.
func withJunctions(pts []marketdata.Point, junctions []time.Time) *marketdata.Series {
	return &marketdata.Series{Name: "yields", Points: pts, Junctions: junctions}
}

// monthEndSample keeps one point per calendar month, the month's LAST
// observation, so a daily yield history becomes the monthly path a
// month-end-dated reconstruction steps along. The trailing month is kept even
// when it is incomplete: it is the latest the source knows, and it is what the
// other monthly references bundled here do.
//
// Each junction also contributes its own two boundary observations (the last
// quote before it and the quote of the day itself), whatever the month they
// fall in. Without them the one monthly step containing a junction would have
// to be skipped whole, which would throw away a real month of return to avoid
// one fabricated day; with them the month is compounded in two pieces and only
// the splice between them is refused. The extra points are dropped again by
// monthEndIndex, which samples the resulting INDEX rather than the yields.
func monthEndSample(pts []marketdata.Point, junctions []time.Time) *marketdata.Series {
	keep := make(map[time.Time]bool, 2*len(junctions))
	for _, j := range junctions {
		for i, p := range pts {
			if p.Date.Equal(j) {
				keep[p.Date] = true
				if i > 0 {
					keep[pts[i-1].Date] = true
				}
			}
		}
	}
	s := &marketdata.Series{Name: "month-end yields", Junctions: junctions}
	for _, p := range pts {
		n := len(s.Points)
		if n > 0 && s.Points[n-1].Date.Format("2006-01") == p.Date.Format("2006-01") && !keep[s.Points[n-1].Date] {
			s.Points[n-1] = p
			continue
		}
		s.Points = append(s.Points, p)
	}
	return s
}

// monthEndIndex keeps one point per calendar month of a compounded index, the
// month's last. Sampling the index rather than the yields is exact: the level
// already carries every step taken inside the month.
func monthEndIndex(name string, s *marketdata.Series) *marketdata.Series {
	out := &marketdata.Series{Name: name, Source: s.Source}
	for _, p := range s.Points {
		if n := len(out.Points); n > 0 && out.Points[n-1].Date.Format("2006-01") == p.Date.Format("2006-01") {
			out.Points[n-1] = p
			continue
		}
		out.Points = append(out.Points, p)
	}
	return out
}

// breakSize is the yield step the declared junction actually shows in the
// assembled series, in percentage points.
func breakSize(pts []marketdata.Point, at time.Time) float64 {
	for i := 1; i < len(pts); i++ {
		if pts[i].Date.Equal(at) {
			return pts[i].Close - pts[i-1].Close
		}
	}
	return math.NaN()
}

// joinCadences puts a monthly head in front of a daily spine of the SAME curve
// point: the monthly observations that predate the daily ones, then the daily
// ones. No map and no rescaling, because the two are the same series published
// at two cadences (the monthly figure is the month's average of the daily one).
func joinCadences(daily, monthly []obs) []marketdata.Point {
	first := daily[0].date
	out := make([]marketdata.Point, 0, len(daily)+len(monthly))
	for _, p := range monthly {
		if p.date.Before(first) {
			out = append(out, marketdata.Point{Date: p.date, Close: p.val})
		}
	}
	for _, p := range daily {
		out = append(out, marketdata.Point{Date: p.date, Close: p.val})
	}
	return out
}

// obs is one dated observation.
type obs struct {
	date time.Time
	val  float64
}

// fetch downloads one DBnomics series and returns its non-null observations in
// date order. Monthly ("YYYY-MM") and daily ("YYYY-MM-DD") periods are both
// accepted; a monthly period is anchored on the first of the month.
func fetch(base, path string) []obs {
	url := fmt.Sprintf("%s/series/%s?observations=1", base, path)
	cl := &http.Client{Timeout: 120 * time.Second}
	resp, err := cl.Get(url)
	if err != nil {
		log.Fatalf("%s: %v", path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("%s: HTTP %d", path, resp.StatusCode)
	}
	var body struct {
		Series struct {
			Docs []struct {
				Period []string `json:"period"`
				Value  []any    `json:"value"`
			} `json:"docs"`
		} `json:"series"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		log.Fatalf("%s: decode: %v", path, err)
	}
	if len(body.Series.Docs) == 0 {
		log.Fatalf("%s: no series returned", path)
	}
	doc := body.Series.Docs[0]
	out := make([]obs, 0, len(doc.Period))
	for i, per := range doc.Period {
		v, ok := doc.Value[i].(float64)
		if !ok {
			continue // DBnomics encodes gaps as the JSON string "NA" or null
		}
		t, err := parsePeriod(per)
		if err != nil {
			log.Fatalf("%s: bad period %q: %v", path, per, err)
		}
		out = append(out, obs{date: t, val: v})
	}
	if len(out) < 2 {
		log.Fatalf("%s: only %d usable observations", path, len(out))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].date.Before(out[j].date) })
	return out
}

func parsePeriod(p string) (time.Time, error) {
	if len(p) == 7 { // YYYY-MM
		return time.Parse("2006-01", p)
	}
	return time.Parse("2006-01-02", p)
}

// fitAffine fits y30 = a + b*y20 by least squares over the days both points
// quote, and returns the coefficients, the residual standard deviation (in
// percentage points) and the number of days fitted. The whole overlap is used
// rather than a window: the fit has to hold at 14 % yields and at 2 %, and
// those eras are decades apart.
func fitAffine(y20, y30 []obs) (a, b, resid float64, n int) {
	by30 := make(map[time.Time]float64, len(y30))
	for _, p := range y30 {
		by30[p.date] = p.val
	}
	var xs, ys []float64
	for _, p := range y20 {
		if v, ok := by30[p.date]; ok {
			xs = append(xs, p.val)
			ys = append(ys, v)
		}
	}
	if len(xs) < 500 {
		log.Fatalf("only %d overlapping days between the 20-year and 30-year points, too few to fit", len(xs))
	}
	mx, my := metrics.Mean(xs), metrics.Mean(ys)
	var sxy, sxx float64
	for i := range xs {
		sxy += (xs[i] - mx) * (ys[i] - my)
		sxx += (xs[i] - mx) * (xs[i] - mx)
	}
	b = sxy / sxx
	a = my - b*mx
	var ss float64
	for i := range xs {
		e := ys[i] - (a + b*xs[i])
		ss += e * e
	}
	return a, b, math.Sqrt(ss / float64(len(xs))), len(xs)
}

// assemble joins the three segments into one yield path: the 30-year spine
// where it quotes, and before it the 20-year point (daily, then monthly deeper
// still) mapped through the affine fit and shifted by the residual measured AT
// the junction, so the series is continuous there. The two 20-year segments are
// the same curve point at two cadences (the monthly figure is the month's
// average of the daily one), so they are concatenated without a second shift.
func assemble(spine, daily20, monthly20 []obs, a, b float64) []marketdata.Point {
	junction := spine[0].date

	// The 20-year path up to the junction: monthly before the daily point
	// starts, daily after.
	firstDaily := daily20[0].date
	var head []obs
	for _, p := range monthly20 {
		if p.date.Before(firstDaily) {
			head = append(head, p)
		}
	}
	for _, p := range daily20 {
		if p.date.Before(junction) {
			head = append(head, p)
		}
	}
	if len(head) < 2 {
		log.Fatalf("the 20-year head has %d observations before %s", len(head), junction.Format("2006-01-02"))
	}

	// The shift that makes the mapped head meet the spine at the junction: the
	// 20-year point of that same day, mapped, against the spine's first value.
	var atJunction float64
	for _, p := range daily20 {
		if p.date.Equal(junction) {
			atJunction = p.val
		}
	}
	if atJunction == 0 {
		log.Fatalf("the 20-year point does not quote on the junction day %s", junction.Format("2006-01-02"))
	}
	shift := spine[0].val - (a + b*atJunction)
	log.Printf("junction %s: 20y %.2f mapped to %.2f, spine %.2f, shift %+.3f pt",
		junction.Format("2006-01-02"), atJunction, a+b*atJunction, spine[0].val, shift)

	out := make([]marketdata.Point, 0, len(head)+len(spine))
	for _, p := range head {
		out = append(out, marketdata.Point{Date: p.date, Close: a + b*p.val + shift})
	}
	for _, p := range spine {
		out = append(out, marketdata.Point{Date: p.date, Close: p.val})
	}
	return out
}

// report logs one series' coverage and range; unit names what the closes are
// (a rate in percent a year, or the levels of a total-return index).
func report(id, unit string, pts []marketdata.Point) {
	if len(pts) == 0 {
		log.Fatalf("%s: empty", id)
	}
	lo, hi := math.Inf(1), math.Inf(-1)
	for _, p := range pts {
		lo, hi = min(lo, p.Close), max(hi, p.Close)
	}
	log.Printf("%-20s %6d points  %s..%s  range %.2f..%.2f %s", id, len(pts),
		pts[0].Date.Format("2006-01-02"), pts[len(pts)-1].Date.Format("2006-01-02"), lo, hi, unit)
}

func write(dir, id, name, source string, junctions []time.Time, pts []marketdata.Point) {
	var b strings.Builder
	b.WriteString("# pofo simdata v1\n")
	fmt.Fprintf(&b, "# id: %s\n", id)
	fmt.Fprintf(&b, "# name: %s\n", name)
	fmt.Fprintf(&b, "# source: %s\n", source)
	fmt.Fprintf(&b, "# generated: %s\n", time.Now().UTC().Format("2006-01-02"))
	for i, j := range junctions {
		if i == 0 {
			b.WriteString("# junctions: ")
		} else {
			b.WriteString(",")
		}
		b.WriteString(j.Format("2006-01-02"))
	}
	if len(junctions) > 0 {
		b.WriteString("\n")
	}
	b.WriteString("date,close\n")
	for _, p := range pts {
		fmt.Fprintf(&b, "%s,%.6f\n", p.Date.Format("2006-01-02"), p.Close)
	}
	path := filepath.Join(dir, id+".csv")
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		log.Fatalf("write %s: %v", path, err)
	}
	log.Printf("wrote %s", path)
}

// runChecks measures the assembled series where an outside answer is known and
// stops the generator when one of them is not met.
//
// The six checks, and why each number is the one to expect:
//
//   - Era levels. The long Treasury yield is one of the best documented numbers
//     in finance: it averaged 3 to 5 % through the 1950s and 1960s, went into
//     double digits from 1980 to 1985, came back under 3 % in the 2010s and
//     rose again after 2021. Any segment landing outside those bands means a
//     misread column, a wrong curve point or a broken map, and the 1980-1985
//     band is the one the strip reconstruction lives or dies on.
//   - The junction. The mapped head has to meet the spine, so the yield change
//     across that one day must be no larger than a normal day's move. A jump
//     there would print as a double-digit one-day return once a 27-year strip
//     is priced off it.
//   - No frozen fetch. A yield that repeats for more than 25 business days in a
//     row is not a market, it is a degraded download: the euro-area generator
//     exists in its current form because two such failures shipped unnoticed.
//     The pre-1962 monthly segment is exempt (a month-average yield did
//     legitimately repeat).
//   - The substituted 2002-2006 window. Over those four years the 30-year row
//     is a 25-years-and-above average, not a 30-year bond. Graded against the
//     20-year point, whose spread to the real 30-year stays inside ±0.35 over
//     the whole record: if the substituted figure were another tenor or another
//     definition, the spread would leave that range.
//   - Freshness. The Fed publishes H.15 daily; a last observation more than 45
//     days old means the source moved, not that the market closed.
//   - The declared 1973 definitional break. A junction that no longer shows a
//     step is a junction forgiving a real market move, so the step is measured
//     and must still be there; and the 10-year point, which did not move that
//     day, must still not move, since it is what makes the step definitional
//     rather than a bond selloff nobody else saw.
func runChecks(long []marketdata.Point, spine, daily20 []obs, junctions []time.Time) {
	for _, e := range []struct {
		lo, hi   int
		min, max float64
	}{
		{1953, 1959, 2.5, 4.5},
		{1960, 1969, 3.5, 5.5},
		{1970, 1979, 6.0, 9.5},
		{1980, 1985, 10.0, 14.0},
		{1990, 1999, 5.5, 8.5},
		{2010, 2019, 2.5, 4.5},
		{2020, 2026, 2.5, 5.0},
	} {
		var sum float64
		var n int
		for _, p := range long {
			if y := p.Date.Year(); y >= e.lo && y <= e.hi {
				sum += p.Close
				n++
			}
		}
		if n == 0 {
			fail("no observation at all over %d-%d", e.lo, e.hi)
			continue
		}
		mean := sum / float64(n)
		log.Printf("check %d-%d: mean long yield %.2f %%/yr over %d observations", e.lo, e.hi, mean, n)
		if mean < e.min || mean > e.max {
			fail("the %d-%d long yield averages %.2f %%/yr, outside the %.1f..%.1f band", e.lo, e.hi, mean, e.min, e.max)
		}
	}

	for i := 1; i < len(long); i++ {
		if long[i].Date.Equal(spine[0].date) {
			jump := long[i].Close - long[i-1].Close
			log.Printf("check junction %s: yield change %+.3f pt", long[i].Date.Format("2006-01-02"), jump)
			if math.Abs(jump) > 0.25 {
				fail("the mapped head does not meet the spine at %s (%+.3f pt)", long[i].Date.Format("2006-01-02"), jump)
			}
		}
	}

	run, worst := 1, 1
	var worstAt time.Time
	firstDaily := daily20[0].date
	for i := 1; i < len(long); i++ {
		if long[i].Date.Before(firstDaily) {
			continue
		}
		if long[i].Close == long[i-1].Close {
			run++
			if run > worst {
				worst, worstAt = run, long[i].Date
			}
			continue
		}
		run = 1
	}
	log.Printf("check flat runs: longest unchanged stretch %d observations (ending %s)", worst, worstAt.Format("2006-01-02"))
	if worst > 25 {
		fail("the yield repeats for %d observations in a row, which is a degraded fetch, not a market", worst)
	}

	from, _ := time.Parse("2006-01-02", substFrom)
	to, _ := time.Parse("2006-01-02", substTo)
	by20 := make(map[time.Time]float64, len(daily20))
	for _, p := range daily20 {
		by20[p.date] = p.val
	}
	var spread []float64
	for _, p := range spine {
		if p.date.Before(from) || p.date.After(to) {
			continue
		}
		if v, ok := by20[p.date]; ok {
			spread = append(spread, p.val-v)
		}
	}
	if len(spread) < 500 {
		fail("only %d days of the substituted %s..%s window can be graded against the 20-year point", len(spread), substFrom, substTo)
	} else {
		m := metrics.Mean(spread)
		log.Printf("check substituted %s..%s: mean spread to the 20-year point %+.3f pt over %d days", substFrom, substTo, m, len(spread))
		if math.Abs(m) > 0.35 {
			fail("the substituted long-term average sits %+.3f pt from the 20-year point, outside the spread's historical range", m)
		}
	}

	last := long[len(long)-1].Date
	age := time.Since(last).Hours() / 24
	log.Printf("check freshness: last observation %s (%.0f days old)", last.Format("2006-01-02"), age)
	if age > 45 {
		fail("the last observation is %.0f days old, the source has moved", age)
	}

	for _, j := range junctions {
		step := breakSize(long, j)
		log.Printf("check definitional break %s: the assembled yield steps %+.3f pt", j.Format("2006-01-02"), step)
		if math.IsNaN(step) {
			fail("the declared junction %s is not a day this series quotes", j.Format("2006-01-02"))
		} else if math.Abs(step) < defBreakMin {
			fail("the declared junction %s shows a step of only %+.3f pt: it no longer describes the data", j.Format("2006-01-02"), step)
		}
	}
}

// checkJunctionNeighbour grades the 1973 declaration against the curve point
// next to it. What makes that day definitional rather than a selloff is that
// the 10-year constant maturity, quoted by the same release on the same day, did
// not move: if it ever does, the junction is hiding a real market move and has
// to go.
func checkJunctionNeighbour(neighbour []obs, at time.Time, maxMove float64) {
	var prev, cur float64
	for i, p := range neighbour {
		if p.date.Equal(at) && i > 0 {
			prev, cur = neighbour[i-1].val, p.val
		}
	}
	if prev == 0 {
		fail("the 10-year point does not quote on the declared junction day %s", at.Format("2006-01-02"))
		return
	}
	log.Printf("check junction neighbour %s: the 10-year point moves %+.3f pt", at.Format("2006-01-02"), cur-prev)
	if math.Abs(cur-prev) > maxMove {
		fail("the 10-year point moves %+.3f pt on %s, more than the %.2f pt a definitional break in the 20-year point allows",
			cur-prev, at.Format("2006-01-02"), maxMove)
	}
}

// checkShape grades a daily shape series against the month-end sibling it
// textures: their monthly returns must be the same series seen twice. The
// LEVELS may differ (the shape prices the published curve point, the monthly
// one the assembled long yield) and the shape carries no fee, so only the shape
// of the path is graded.
func checkShape(id string, shape, monthly *marketdata.Series, from, to time.Time) {
	corr := monthlyCorr(shape, monthly, from, to)
	log.Printf("check %s vs %s over %s..%s: monthly corr %.3f", id, monthly.Name,
		from.Format("2006-01"), to.Format("2006-01"), corr)
	if !(corr > 0.97) {
		fail("%s does not track %s (monthly corr %.3f)", id, monthly.Name, corr)
	}
	days := 0
	for _, p := range shape.Points {
		if p.Date.Year() == 1975 {
			days++
		}
	}
	log.Printf("check %s: %d observations in 1975", id, days)
	if days < 230 {
		fail("%s carries %d points in 1975, which is not daily density", id, days)
	}
}

// failed counts the checks that did not pass, across every series this command
// writes; nothing is written while it is non-zero.
var failed int

func fail(format string, args ...any) {
	failed++
	log.Printf("CHECK FAILED: "+format, args...)
}

// sbbiLong and sbbiInter are published calendar-year total returns of
// long-term (~20-year) and intermediate-term (5-year) US GOVERNMENT BONDS, from
// the Ibbotson SBBI record, with the tolerance each reconstruction is held to.
// They are the years pkg/datasets/golden asserts, checked here so a refresh
// that breaks them stops before it writes rather than after. A
// constant-maturity par bond is not the SBBI portfolio, hence the couple of
// points of slack; it is tight enough to catch a unit, day-count, cadence or
// repricing regression.
//
// 1973 is the year the 20-year point changed definition (defBreak20) and it is
// here for that reason: priced straight through the break, the long series read
// -6.78 % where the published record says -1.11 %. With the break declared it
// reads +0.91 %, a residual of +2.0 points, which is the ordinary size of this
// engine's disagreement with the SBBI portfolio (1969 misses by 1.3, 1982 by
// 3.6, 1994 by 0.8, 1995 by 0.5: a constant-maturity par bond is not the
// portfolio SBBI tracked). Its tolerance is therefore 2.5, the same as 1995's,
// and it still leaves a factor of nearly three against the 5.7-point miss the
// defect produced. It is the one calendar year in the whole record the break is
// visible in, so it is the one that guards the junction from the outside.
//
// The figures are read off the year-by-year total-return table of Ibbotson and
// Sinquefield, "Stocks, Bonds, Bills, and Inflation: Historical Returns
// (1926-1987)", CFA Institute Research Foundation, at
// rpc.cfainstitute.org/sites/default/files/-/media/documents/book/rf-publication/1989/rf-v1989-n3-sbbi-historical-returns.pdf
// (read 2026-09-20): long-term government bonds -5.08 % in 1969, -1.11 % in
// 1973, and intermediate-term -0.74 % and +4.60 % in the same two years, the
// same table also giving the S&P's -14.66 % of 1973 as a cross-check that the
// columns are the ones they are taken for. 1982, 1994 and 1995 postdate that
// edition and carry the figures the goldens have always asserted.
var (
	sbbiLong  = []yearRef{{1969, -5.1, 2.0}, {1973, -1.1, 2.5}, {1982, 40.4, 4.0}, {1994, -7.8, 2.0}, {1995, 31.7, 2.5}}
	sbbiInter = []yearRef{{1969, -0.7, 1.5}, {1973, 4.6, 1.5}, {1982, 29.1, 2.5}, {1994, -5.1, 1.5}, {1995, 16.8, 1.5}}
)

// yearRef is one published calendar-year return (percent) and the tolerance the
// reconstruction is graded at.
type yearRef struct {
	year     int
	ret, tol float64
}

// checkReconstruction grades one total-return series before it is written.
//
// The four things measured, and why each is the one to measure:
//
//   - One point per calendar month, ascending, with no hole. This is the check
//     the old files did not have and it is the reason this command now owns
//     them: the long one bridged 1987-01..1993-09 in a single step for years,
//     which reads as one +60 % month and seven years of no volatility to any
//     direct consumer. A step longer than 40 days is a missing month.
//   - The published SBBI calendar years. An external number, on both a
//     disinflation year and two bear years, which is what grades the engine.
//   - The long-run CAGR over 1972-2021, the window the bundled goldens assert,
//     against the same SBBI-era figure.
//   - Annualized monthly volatility over that window, which is what a cadence
//     or dating regression moves first.
func checkReconstruction(id string, tr *marketdata.Series, refs []yearRef, cagr, ctol, volLo, volHi float64) {
	pts := tr.Points
	if len(pts) < 600 {
		fail("%s: only %d monthly points", id, len(pts))
		return
	}
	worst, at := 0, time.Time{}
	for i := 1; i < len(pts); i++ {
		if !pts[i].Date.After(pts[i-1].Date) {
			fail("%s: %s does not follow %s", id, pts[i].Date.Format("2006-01-02"), pts[i-1].Date.Format("2006-01-02"))
		}
		if d := int(pts[i].Date.Sub(pts[i-1].Date).Hours() / 24); d > worst {
			worst, at = d, pts[i].Date
		}
	}
	log.Printf("check %s: %d points %s..%s, longest step %d days (ending %s)", id, len(pts),
		pts[0].Date.Format("2006-01-02"), pts[len(pts)-1].Date.Format("2006-01-02"), worst, at.Format("2006-01-02"))
	if worst > 40 {
		fail("%s: a %d-day step ending %s, i.e. a missing month", id, worst, at.Format("2006-01-02"))
	}

	for _, r := range refs {
		got := calendarYear(tr, r.year)
		log.Printf("check %s %d: %+.2f %%, published %+.2f %%", id, r.year, got, r.ret)
		if math.Abs(got-r.ret) > r.tol {
			fail("%s: %d returns %+.2f %%, %.2f from the published %+.2f %%", id, r.year, got, math.Abs(got-r.ret), r.ret)
		}
	}

	rets := monthlyReturns(monthly(tr, date(1971, 12), date(2022, 1)))
	if len(rets) < 500 {
		fail("%s: only %d months over 1972-2021", id, len(rets))
		return
	}
	prod := 1.0
	for _, r := range rets {
		prod *= 1 + r
	}
	got := (math.Pow(prod, 12/float64(len(rets))) - 1) * 100
	vol := stddev(rets) * math.Sqrt(12) * 100
	log.Printf("check %s 1972-2021: CAGR %.2f %%/yr (published %.1f), monthly volatility %.2f %%/yr", id, got, cagr, vol)
	if math.Abs(got-cagr) > ctol {
		fail("%s: 1972-2021 compounds at %.2f %%/yr, %.2f from the published %.1f", id, got, math.Abs(got-cagr), cagr)
	}
	if vol < volLo || vol > volHi {
		fail("%s: 1972-2021 volatility %.2f %%/yr, outside the %.0f..%.0f band", id, vol, volLo, volHi)
	}
}

// calendarYear is the December-to-December total return (percent) of year y.
func calendarYear(s *marketdata.Series, y int) float64 {
	dec := func(yr int) float64 {
		cut := time.Date(yr, 12, 31, 23, 0, 0, 0, time.UTC)
		var v float64
		for _, p := range s.Points {
			if p.Date.After(cut) {
				break
			}
			v = p.Close
		}
		return v
	}
	a, b := dec(y-1), dec(y)
	if a == 0 {
		return math.NaN()
	}
	return (b/a - 1) * 100
}

func stddev(xs []float64) float64 {
	m := metrics.Mean(xs)
	var v float64
	for _, x := range xs {
		v += (x - m) * (x - m)
	}
	return math.Sqrt(v / float64(len(xs)-1))
}

func date(y, m int) time.Time { return time.Date(y, time.Month(m), 1, 0, 0, 0, 0, time.UTC) }

// monthEnd is one calendar month's closing level, keyed "YYYY-MM", in order.
type monthEnd struct {
	key   string
	close float64
}

// monthly returns the last level of each calendar month in [from, to).
func monthly(s *marketdata.Series, from, to time.Time) []monthEnd {
	var out []monthEnd
	for _, p := range s.Points {
		if p.Date.Before(from) || !p.Date.Before(to) {
			continue
		}
		k := p.Date.Format("2006-01")
		if n := len(out); n > 0 && out[n-1].key == k {
			out[n-1].close = p.Close
			continue
		}
		out = append(out, monthEnd{key: k, close: p.Close})
	}
	return out
}

func monthlyReturns(m []monthEnd) []float64 {
	r := make([]float64, 0, len(m))
	for i := 1; i < len(m); i++ {
		r = append(r, m[i].close/m[i-1].close-1)
	}
	return r
}

// monthlyCorr correlates two series' monthly returns over the window, matching
// them by calendar month rather than by position, so two series that skip a
// different month are never compared off by one.
func monthlyCorr(x, y *marketdata.Series, from, to time.Time) float64 {
	byKey := make(map[string]float64)
	for _, m := range monthly(y, from, to) {
		byKey[m.key] = m.close
	}
	var mx, my []monthEnd
	for _, m := range monthly(x, from, to) {
		if c, ok := byKey[m.key]; ok {
			mx = append(mx, m)
			my = append(my, monthEnd{key: m.key, close: c})
		}
	}
	rx, ry := monthlyReturns(mx), monthlyReturns(my)
	if len(rx) < 12 {
		return math.NaN()
	}
	return pearson(rx, ry)
}

func pearson(a, b []float64) float64 {
	ma, mb := metrics.Mean(a), metrics.Mean(b)
	var cov, va, vb float64
	for i := range a {
		da, db := a[i]-ma, b[i]-mb
		cov += da * db
		va += da * da
		vb += db * db
	}
	if va <= 0 || vb <= 0 {
		return math.NaN()
	}
	return cov / math.Sqrt(va*vb)
}
