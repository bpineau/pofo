package marketdata

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
)

// Band bounds the statistics an asset of a given class can plausibly show over
// its whole history. It is a data-quality instrument, not a forecast: the
// bounds sit outside every value the bundled catalog actually measures, so a
// series that leaves one is almost always a resolution or provider accident
// (a euro fund served through a pound line, a spliced share class, an isolated
// bad print) rather than an asset that surprised its class.
//
// Units are FRACTIONS per year (0.42 = 42 %/yr), magnitudes are positive
// (Move 0.23 means "a 23 % move in either direction"). Every bound scales with
// the record's notional leverage; see Scale.
//
// The table lives in ClassBand and is documented, next to the vocabularies it
// is keyed by, in pkg/datasets/assetmeta/README.md.
type Band struct {
	VolLo, VolHi   float64 // annualized volatility of daily returns
	CAGRLo, CAGRHi float64 // annualized growth over the whole history
	Move           float64 // largest plausible single-session move, magnitude
	Drawdown       float64 // deepest plausible peak-to-trough drawdown, magnitude
}

// classBands is the per-asset_class table, at leverage 1. It was calibrated on
// the measured statistics of the whole bundled catalog (238 records, 2026-08),
// each bound placed clear of the widest real value in its class so that a
// full-catalog run flags only genuine accidents. Deliberate consequences worth
// knowing:
//
//   - money-market VolHi accepts XEON's 1.0 %/yr, which its 2007 launch weeks
//     alone produce (the class targets ~0.7 %/yr and every other member sits
//     below); its Move of 3.5 % accepts March 2020's ultrashort credit
//     dislocation (-3.2 % for ERNA) and rejects the annual income drops of a
//     distributing money-market NAV.
//   - corporate-bond VolLo accepts the AAA CLO funds at 0.8 to 1.1 %/yr.
//   - equity Move accepts 1987-10-19 (-20.5 %) and Drawdown accepts a single
//     country index through its own crisis (Greece, -96 %).
//   - insurance-linked spans a very wide VolHi because the class covers both a
//     currency-hedged share class (~3 %/yr) and the same fund held unhedged by
//     a euro investor, where EURUSD supplies most of the variation (~8 %/yr);
//     its Move accepts a single event step (the reference index lost 8.6 % in
//     the month of Harvey, Irma and Maria) and its Drawdown accepts a fund that
//     lost half its capital to one bad season.
//   - "other" is a junk drawer, so its bounds are the widest of the table;
//     they still reject a quarter lost in one session.
var classBands = map[string]Band{
	"equity":                {VolLo: 0.06, VolHi: 0.42, CAGRLo: -0.15, CAGRHi: 0.40, Move: 0.23, Drawdown: 0.97},
	"government-bond":       {VolLo: 0.003, VolHi: 0.28, CAGRLo: -0.10, CAGRHi: 0.15, Move: 0.14, Drawdown: 0.80},
	"corporate-bond":        {VolLo: 0.005, VolHi: 0.14, CAGRLo: -0.08, CAGRHi: 0.12, Move: 0.10, Drawdown: 0.45},
	"aggregate-bond":        {VolLo: 0.01, VolHi: 0.12, CAGRLo: -0.06, CAGRHi: 0.10, Move: 0.08, Drawdown: 0.30},
	"inflation-linked-bond": {VolLo: 0.01, VolHi: 0.14, CAGRLo: -0.06, CAGRHi: 0.12, Move: 0.08, Drawdown: 0.30},
	"insurance-linked":      {VolLo: 0.005, VolHi: 0.16, CAGRLo: -0.06, CAGRHi: 0.15, Move: 0.10, Drawdown: 0.55},
	"money-market":          {VolLo: 0.0, VolHi: 0.03, CAGRLo: -0.02, CAGRHi: 0.07, Move: 0.035, Drawdown: 0.10},
	"gold":                  {VolLo: 0.08, VolHi: 0.30, CAGRLo: -0.10, CAGRHi: 0.25, Move: 0.15, Drawdown: 0.60},
	"broad-commodity":       {VolLo: 0.10, VolHi: 0.45, CAGRLo: -0.20, CAGRHi: 0.25, Move: 0.22, Drawdown: 0.85},
	"managed-futures":       {VolLo: 0.04, VolHi: 0.28, CAGRLo: -0.15, CAGRHi: 0.30, Move: 0.15, Drawdown: 0.45},
	"long-volatility":       {VolLo: 0.04, VolHi: 0.60, CAGRLo: -0.30, CAGRHi: 0.30, Move: 0.25, Drawdown: 0.60},
	"tail-risk":             {VolLo: 0.04, VolHi: 0.60, CAGRLo: -0.30, CAGRHi: 0.30, Move: 0.25, Drawdown: 0.60},
	"multi-asset":           {VolLo: 0.05, VolHi: 0.32, CAGRLo: -0.12, CAGRHi: 0.28, Move: 0.15, Drawdown: 0.45},
	"real-estate":           {VolLo: 0.12, VolHi: 0.38, CAGRLo: -0.15, CAGRHi: 0.25, Move: 0.25, Drawdown: 0.80},
	"other":                 {VolLo: 0.0, VolHi: 0.60, CAGRLo: -0.35, CAGRHi: 0.40, Move: 0.25, Drawdown: 0.60},
}

// ClassBand returns the plausibility band of an asset_class of the catalog
// vocabulary, at leverage 1. known is false for a class the table does not
// cover, in which case no plausibility judgement should be made at all.
func ClassBand(class string) (b Band, known bool) {
	b, known = classBands[class]
	return b, known
}

// Scale returns the band widened by a notional leverage multiple (the catalog's
// leverage field: 1 for a plain fund, 3 for a daily 3x ETF, 1.5 for a 90/60
// stacked one). Everything scales linearly except the drawdown, which is capped
// just short of a total loss. A non-positive leverage reads as 1.
func (b Band) Scale(leverage float64) Band {
	if leverage <= 0 {
		leverage = 1
	}
	b.VolLo *= leverage
	b.VolHi *= leverage
	b.CAGRLo *= leverage
	b.CAGRHi *= leverage
	b.Move *= leverage
	b.Drawdown = math.Min(0.99, b.Drawdown*leverage)
	return b
}

// spikeLegSigmas is how many class-level daily standard deviations a leg of a
// one-session round trip must span before the cleaner may consider it a bad
// print. Four is far outside ordinary volatility yet well inside a crash: it
// keeps every real dislocation this catalog contains (2020-03-12 at -9.5 % on a
// class whose ceiling implies 10.6 %) out of the cleaner's reach.
const spikeLegSigmas = 4.0

// SpikeLeg is the smallest one-session move the round-trip cleaner will even
// consider a candidate bad print for this class: spikeLegSigmas times the daily
// standard deviation implied by the class volatility ceiling. It is a necessary
// condition, never a sufficient one (dropRoundTrips also demands a full
// reversal and 6 local sigmas), and it is deliberately far LOOSER than Move:
// Move answers "did this asset move more than its class ever can", which is a
// finding on its own, while SpikeLeg only stops the cleaner from touching
// ordinary volatility on the way to a reversal it can prove.
func (b Band) SpikeLeg() float64 {
	return spikeLegSigmas * b.VolHi / math.Sqrt(tradingDaysPerYear)
}

// tradingDaysPerYear is the annualization factor used across the toolkit.
const tradingDaysPerYear = 252

// widestBand answers for an identifier no catalog record claims. It is the
// loosest row of the table, so an uncatalogued series gets the most
// conservative treatment there is: the cleaner touches it only for moves no
// instrument makes honestly.
var widestBand = classBands["other"]

// bandBySymbol indexes the catalog by PROVIDER symbol (and FT xid), which
// Lookup deliberately does not: a quote symbol may be claimed by several
// records (the US-listed NTSX and the NTSX UCITS). The cleaning pass is the one
// caller that only ever knows the symbol, and it needs a band, not an identity,
// so a collision is resolved by keeping the WIDEST band of the claimants. That
// can only make the cleaner more timid, never bolder.
var bandBySymbol = sync.OnceValue(func() map[string]Band {
	m := make(map[string]Band, 2*len(catalog))
	add := func(key string, b Band) {
		if key == "" {
			return
		}
		key = strings.ToUpper(key)
		if old, seen := m[key]; seen {
			b = widest(old, b)
		}
		m[key] = b
	}
	for _, e := range catalog {
		b, known := ClassBand(e.AssetClass)
		if !known {
			continue
		}
		b = b.Scale(e.Leverage)
		add(e.Symbol, b)
		add(e.Xid, b)
	}
	return m
})

// widest returns the union of two bands: every bound relaxed to the laxer one.
func widest(a, b Band) Band {
	return Band{
		VolLo:    math.Min(a.VolLo, b.VolLo),
		VolHi:    math.Max(a.VolHi, b.VolHi),
		CAGRLo:   math.Min(a.CAGRLo, b.CAGRLo),
		CAGRHi:   math.Max(a.CAGRHi, b.CAGRHi),
		Move:     math.Max(a.Move, b.Move),
		Drawdown: math.Max(a.Drawdown, b.Drawdown),
	}
}

// bandFor is the band to judge an identifier by, whatever the caller knows it
// as: a canonical id, alias, ISIN or ticker (via Lookup), or a bare provider
// symbol (via bandBySymbol). It always answers, falling back to widestBand.
func bandFor(id string) Band {
	if a, ok := Lookup(id); ok {
		if b, known := ClassBand(a.AssetClass); known {
			return b.Scale(a.Leverage)
		}
	}
	if b, ok := bandBySymbol()[strings.ToUpper(strings.TrimSpace(id))]; ok {
		return b
	}
	return widestBand
}

// shape is the handful of statistics the plausibility checks judge, measured on
// the series as served (its own cadence included, so a weekly NAV annualizes by
// its own pace rather than by 252).
type shape struct {
	perYear  float64   // observations per year
	years    float64   // span of the series
	vol      float64   // annualized standard deviation of the returns
	cagr     float64   // annualized growth over the span
	move     float64   // largest single-session move, magnitude
	moveOn   time.Time // when it happened
	moveUp   bool      // its direction
	drawdown float64   // deepest peak-to-trough fall, magnitude
}

// measure computes the shape of a series, or ok = false when it is too short or
// too degenerate (non-positive prices, no span) to say anything about.
func measure(pts []Point) (sh shape, ok bool) {
	if len(pts) < 3 {
		return sh, false
	}
	sh.years = pts[len(pts)-1].Date.Sub(pts[0].Date).Hours() / 24 / 365.25
	if sh.years <= 0 || pts[0].Close <= 0 || pts[len(pts)-1].Close <= 0 {
		return sh, false
	}
	sh.perYear = float64(len(pts)-1) / sh.years
	sh.cagr = math.Pow(pts[len(pts)-1].Close/pts[0].Close, 1/sh.years) - 1

	rets := make([]float64, 0, len(pts)-1)
	peak := pts[0].Close
	for i := 1; i < len(pts); i++ {
		if pts[i-1].Close <= 0 || pts[i].Close <= 0 {
			return sh, false
		}
		r := pts[i].Close/pts[i-1].Close - 1
		rets = append(rets, r)
		if math.Abs(r) > sh.move {
			sh.move, sh.moveOn, sh.moveUp = math.Abs(r), pts[i].Date, r > 0
		}
		peak = math.Max(peak, pts[i].Close)
		sh.drawdown = math.Max(sh.drawdown, 1-pts[i].Close/peak)
	}
	var mean float64
	for _, r := range rets {
		mean += r
	}
	mean /= float64(len(rets))
	var ss float64
	for _, r := range rets {
		ss += (r - mean) * (r - mean)
	}
	sh.vol = math.Sqrt(ss/float64(len(rets)-1)) * math.Sqrt(sh.perYear)
	return sh, true
}

// VerifyAsset is the data doctor's full pass on a catalogued identifier: the
// series hygiene Verify judges from the quotes alone, plus the two families of
// finding that need the catalog record next to them.
//
//   - PLAUSIBILITY. Volatility, CAGR, largest single-session move and deepest
//     drawdown against the asset class's Band, scaled by the record's leverage.
//     This is the check that names, in one line, the accidents a green fetch
//     hides: an aggregate-bond fund at 20 %/yr volatility is being served
//     through a foreign-currency line, a one-to-three year government fund with
//     a 27 % drawdown is two quote lines welded together.
//   - IDENTITY. The served currency against the record's currency field (which
//     names the quote line, see the assetmeta README), the served share-class
//     name against its distribution field, and the first quote against since,
//     the share class's official launch date.
//
// Series whose numbers are not prices are exempt from plausibility: rate levels
// (^IRX, ^ESTR, …), index-sourced reconstructions, and the continuous-futures
// and FX symbols behind spot quotes. So are identifiers the catalog does not
// know, which have no class to be judged against.
//
// Findings are heuristics on real, sometimes wild, market data, and several
// true ones are permanent (a fund whose provider serves its predecessor's
// history really does start before its inception). Read them, do not chase
// them to zero.
func VerifyAsset(id string, s *Series, now time.Time) []Issue {
	if s == nil || len(s.Points) < 3 {
		return Verify(s, now)
	}
	base, _ := SplitSim(id) // the SIM suffix names a view, not another asset
	a, ok := Lookup(base)
	if !ok {
		return Verify(s, now)
	}
	// One move check, the sharpest available: when the class is known, its own
	// Move replaces Verify's blanket quarter, so a 3x fund stops warning four
	// times over March 2020 and a money market starts warning at 3.5 %.
	band, judged := assetBand(a, s)
	moveLimit := 0.0
	if judged {
		moveLimit = band.Move
	}
	issues := verify(s, now, moveLimit)
	issues = append(issues, identityIssues(a, s)...)
	if judged {
		issues = append(issues, plausibilityIssues(a, s, band)...)
	}
	return issues
}

// assetBand returns the band a record is to be judged by, and whether it should
// be judged at all: a class the table covers, on a series whose numbers are
// prices.
func assetBand(a datasets.Asset, s *Series) (Band, bool) {
	b, known := ClassBand(a.AssetClass)
	if !known || !pricedInstrument(a, s) {
		return Band{}, false
	}
	return b.Scale(a.Leverage), true
}

// pricedInstrument reports whether a series' numbers are prices whose ratios
// mean something, the premise of every plausibility bound. Rate levels, index
// reconstructions and the futures/FX symbols standing in for spot are not
// judged: the first two are not instruments an investor holds, and the third
// carries roll seams and, in April 2020, a negative price.
func pricedInstrument(a datasets.Asset, s *Series) bool {
	if a.Source == "index" {
		return false
	}
	for _, sym := range []string{a.Symbol, s.Symbol} {
		if isRateSymbol(sym) || isPolicyRate(sym) ||
			strings.HasSuffix(sym, "=F") || strings.HasSuffix(sym, "=X") {
			return false
		}
	}
	return true
}

// plausibilityIssues judges a series' whole shape against its asset class's
// band, already scaled to the record's leverage. The single-observation move is
// not judged here: verify owns it, so that one series never produces two
// verdicts on the same session.
func plausibilityIssues(a datasets.Asset, s *Series, band Band) []Issue {
	sh, ok := measure(s.Points)
	if !ok {
		return nil
	}
	lev := ""
	if a.Leverage > 0 && a.Leverage != 1 {
		lev = fmt.Sprintf(" at leverage %g", a.Leverage)
	}
	var issues []Issue
	warn := func(d time.Time, format string, args ...any) {
		issues = append(issues, Issue{Severity: "warn", Date: d, Message: fmt.Sprintf(format, args...)})
	}
	if sh.vol < band.VolLo || sh.vol > band.VolHi {
		warn(time.Time{}, "volatility %.1f %%/yr is outside the %s band [%.1f, %.1f]%s, wrong quote line?",
			sh.vol*100, a.AssetClass, band.VolLo*100, band.VolHi*100, lev)
	}
	// A CAGR needs a span to mean anything: three years of a young share class
	// says more about its first months than about its class.
	if sh.years >= 3 && (sh.cagr < band.CAGRLo || sh.cagr > band.CAGRHi) {
		warn(time.Time{}, "CAGR %+.1f %%/yr over %.1f yr is outside the %s band [%+.1f, %+.1f]%s",
			sh.cagr*100, sh.years, a.AssetClass, band.CAGRLo*100, band.CAGRHi*100, lev)
	}
	if sh.drawdown > band.Drawdown {
		warn(time.Time{}, "max drawdown -%.1f %% beyond the %.1f %% a %s%s can fall",
			sh.drawdown*100, band.Drawdown*100, a.AssetClass, lev)
	}
	return issues
}

// sinceDrift is how far the first quote may sit from the recorded inception
// before the doctor says so, in either direction. Provider depth and share
// class launches rarely coincide to the day, and a year of slack absorbs every
// honest mismatch; beyond it, one of the two is telling a different story:
// earlier means the provider is serving a predecessor's history under this
// class's name, later means the depth the record promises does not exist.
const sinceDrift = 400 * 24 * time.Hour

// identityIssues cross-checks the served series against what the record claims
// it is. Nothing here is ever repaired: the catalog is curated by hand, so the
// doctor's job is to name the disagreement, not to pick a winner.
func identityIssues(a datasets.Asset, s *Series) []Issue {
	var issues []Issue
	warn := func(d time.Time, format string, args ...any) {
		issues = append(issues, Issue{Severity: "warn", Date: d, Message: fmt.Sprintf(format, args...)})
	}
	// The record's currency names the quote line. A disagreement means the
	// pinned symbol is another listing of the fund (one more FX layer) or
	// another instrument, and it aims the re-resolution tie-break at the wrong
	// one. GBp and GBP are not distinguished: providers spell the pence line
	// both ways, and a hundredfold is the scale-break pass's business.
	if a.Currency != "" && s.Currency != "" && !strings.EqualFold(a.Currency, s.Currency) {
		warn(time.Time{}, "catalog says the line quotes in %s, the source serves %s", a.Currency, s.Currency)
	}
	// The served share-class name is the only place a pinned symbol admits it
	// serves the sibling class. Only an outright contradiction is reported.
	if s.Name != "" {
		switch {
		case a.Distribution == "accumulating" && LooksDistributing(s.Name) && !looksAccumulating(s.Name):
			warn(time.Time{}, "catalog says accumulating, the served name reads distributing: %q", s.Name)
		case a.Distribution == "distributing" && looksAccumulating(s.Name) && !LooksDistributing(s.Name):
			warn(time.Time{}, "catalog says distributing, the served name reads accumulating: %q", s.Name)
		}
	}
	// since is the share class's own launch date, so the first real quote
	// should be within a year of it. A SIM-extended series starts before any
	// quote by construction, and an index reconstruction predates its vehicle
	// on purpose; neither is asked the question.
	if a.Since != "" && a.Source != "index" && s.SimulatedBefore.IsZero() {
		if since, err := time.Parse("2006-01-02", a.Since); err == nil {
			first := s.First().Date
			if drift := first.Sub(since); drift > sinceDrift {
				warn(first, "first quote is %.0f days after the %s inception: missing provider depth?",
					drift.Hours()/24, a.Since)
			} else if -drift > sinceDrift {
				warn(first, "first quote is %.0f days before the %s inception: predecessor history?",
					-drift.Hours()/24, a.Since)
			}
		}
	}
	return issues
}

// accumulationMarkers are the share-class suffixes an issuer appends to say a
// class reinvests its income. "cap" is deliberately absent, and so is any bare
// letter: a substring test on either turns "Small Cap Value", "Alternative
// Access" and "Preferred and Capital Securities" into accumulating classes,
// which is how a heuristic earns its bad name.
var accumulationMarkers = map[string]bool{
	"acc": true, "accumulating": true, "accumulation": true, "accumulative": true,
	"c": true, "1c": true, "2c": true, "thesaurierend": true,
	"capitalisation": true, "capitalising": true, "capitalizing": true,
}

// looksAccumulating reports whether a fund share-class name advertises
// accumulation. It mirrors LooksDistributing and stays deliberately narrow: a
// marker counts only as a whole word, so the bare "C" of "Europe Small A (C)"
// is read while the "Cap" of "Small Cap Value" is not.
func looksAccumulating(name string) bool {
	for _, word := range strings.FieldsFunc(strings.ToLower(name), func(r rune) bool {
		return !('a' <= r && r <= 'z' || '0' <= r && r <= '9')
	}) {
		if accumulationMarkers[word] {
			return true
		}
	}
	return false
}
