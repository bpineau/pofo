package simgen

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// Verdict grades one aspect of a reconstruction.
type Verdict string

// The three grades, plus the one given when nothing could be measured.
const (
	VerdictOK      Verdict = "ok"
	VerdictWarn    Verdict = "warn"
	VerdictBad     Verdict = "bad"
	VerdictUnknown Verdict = "n/a"
)

// Junction grades one link of a donor chain on the overlap the two sides
// share, which is the only window where that link can be judged at all.
type Junction struct {
	Span     string  // the years this link fills in the shipped file
	Pair     string  // "deeper vs nearer", the way the splice reads
	Months   int     // common calendar months
	Corr     float64 // monthly correlation over them
	GapYear  float64 // CAGR of the deeper minus the nearer, per year
	Measured bool    // false when the two cannot be compared (see Note)
	Note     string
}

// AuditResult is one recipe's engine graded against reality: the raw
// reconstruction (no real quotes spliced in) measured over the window where
// the asset's own quotes exist.
//
// Two verdicts are kept apart because the two failures are different beasts.
// Level asks whether the engine earns the asset's return, since a hot engine
// flatters every backtest downstream. Path asks whether it moves with the
// asset, judged on the monthly correlation and on the tracking error relative
// to the asset's own volatility, so a cash-like proxy is not condemned for a
// meaningless daily correlation.
type AuditResult struct {
	ID     string // the recipe's identifier
	Name   string // the recipe's display name
	Method string // the recipe's one-line construction summary
	Group  string // the family this recipe belongs to, for ordering

	Err      string   // non-empty when nothing could be measured
	Rejected []string // reference candidates refused, and why

	Reference  string    // the identifier that served as the truth
	Start, End time.Time // the measured window
	Years      float64
	Short      bool      // under two years: read the return gap as noise
	RealFrom   time.Time // date from which a SIM consumer gets real quotes

	DailyCorr, WeeklyCorr, MonthlyCorr float64
	Beta, TrackingErr                  float64
	CAGRSim, CAGRReal, Delta           float64 // Delta = engine - real, per year
	TotalDrift                         float64 // engine/real over the window, as a fraction
	VolSim, VolReal                    float64
	WorstSim, WorstReal                float64 // worst single-day return

	Level, Path Verdict
	Score       float64 // severity, worst first; presentation only

	Engine, Real *marketdata.Series   // clipped to the window, for charting
	Others       []*marketdata.Series // curated comparison curves, same window
	Chain        []Junction
	Caveat       string // a hand-written note the numbers alone would misread
}

// Measured reports whether the audit found an independent reference and
// produced statistics.
func (a AuditResult) Measured() bool { return a.Err == "" }

// AuditGroup is a family of recipes, in the order the report shows them.
type AuditGroup struct {
	Title   string
	Note    string
	Results []AuditResult
}

// auditGroups is the reading order of the report: the engines that need the
// most scrutiny first. A recipe absent from every list still shows up, under
// otherGroup, so a new one is never silently dropped.
var auditGroups = []struct {
	Title string
	Note  string
	IDs   []string
}{
	{
		Title: "Trend / managed futures",
		Note: "Donor chains behind a published trend index, the TSMOM engine supplying only the daily texture. " +
			"The engine is a proxy for the strategy, never the fund: read the drift panel for an overlay that runs systematically hot or cold.",
		IDs: []string{"DBMF", "LU2951555585", "DBMFE", "KMLM", "CTA", "LU1103257975",
			"LU1662501532", "IE000O1VI174", "RSST", "RSBT"},
	},
	{
		Title: "Capital-efficient / stacked",
		Note: "Leg-by-leg replications of a levered sleeve (90/60, 100/100, a gold overlay). " +
			"The level verdicts of this family read on windows too short to carry them: measured on whole months, " +
			"no fund's CAGR gap reaches one and a half standard errors from zero, and the two longest windows " +
			"(ZROZ's sixteen years, NTSX's eight) are the two smallest gaps. Measured 2026-08; see the design doc.",
		IDs: []string{"IE000KF370H3", "IE00077IIPQ8", "IE000OV4XWA3", "GDE", "RSSB", "ZROZ"},
	},
	{
		Title: "Equity",
		IDs: []string{"URTH", "IE00B4L5Y983", "WPEA", "MSCIWORLD", "SP500", "IE00BFMXXD54",
			"VTI", "VT", "IE00BK5BQT80", "IE00BKM4GZ66", "IE00BSPLC413", "IE000S67ID55",
			"IE0003R87OG3", "LU1832174962", "ERESMONDEM"},
	},
	{
		Title: "Bonds and cash",
		IDs: []string{"TLT", "IEF", "SHY", "IE00BSKRJZ44", "DTLA", "DTLETR", "DBXG", "MTH",
			"LU1645380442", "LU1459801780", "IE00BGCSB447", "IE000RHYOR04", "LU0290358497"},
	},
	{
		Title: "Commodities",
		IDs:   []string{"XAUUSD", "IE00B4ND3602", "IE00BDFL4P12"},
	},
}

// otherGroup collects the recipes no list names, so adding a recipe never
// means losing it from the report.
const otherGroup = "Other"

// auditExtras adds hand-picked comparison series per recipe: donor share
// classes, US twins, the sibling whose history the recipe leans on. They are
// candidate references (in order, after the recipe's own identifiers) and
// they are drawn on the chart.
var auditExtras = map[string][]string{
	"LU1662501532": {"LU1103258197"}, // the B EUR sister class the recipe grafts
	"LU2951555585": {"DBMF"},         // US twin of the UCITS share class
	"DBMFE":        {"DBMF"},
	"IE000OV4XWA3": {"NTSX"},
	"IE00077IIPQ8": {"NTSX"},
	"WPEA":         {"IWDA.L"},
	"DTLETR":       {"DTLA"},
	"MSCIWORLD":    {"URTH", "IWDA.L"},
	"SP500":        {"VOO", "IE00BFMXXD54"},
	"ERESMONDEM":   {"WPEA", "XDWD.DE"},
	"XAUUSD":       {"GLD"},
}

// auditCaveats carry what the statistics alone would misread.
var auditCaveats = map[string]string{
	"DTLETR": "The positive gap is expected: DTLE is a distributing share class whose NAV is a price return, " +
		"missing its coupon (~3.1 %/yr measured), while the engine is a total return.",
	"LU1662501532": "The engine already contains the sister class and the real quotes, so the live window holds " +
		"nothing independent to validate; only the pre-2015 segment is a reconstruction.",
	"XAUUSD": "The engine is the real quote extended by the LBMA fixing: over the real period it is the real series.",
	"GDE": "The warn level is 0.45 standard errors from zero on whole months (+0.85 pt/yr over 51 of them, " +
		"3.86 % monthly tracking error), so it measures nothing yet. One named mechanism points the same way and " +
		"is deliberately not applied: the gold leg's price series carries no roll (Yahoo's GC=F compounds within " +
		"a hundredth of the LBMA fix), while the fund holds futures whose embedded financing ran ~1.4 %/yr above " +
		"the bill rate over 2009-2023, measured on the only public futures-based gold vehicle. That figure is a " +
		"vehicle's shortfall rather than a price list, and applying it would overshoot the gap. Measured 2026-08.",
	"RSSB": "The warn level is 1.4 standard errors from zero on whole months (+1.43 pt/yr over 30 of them), " +
		"which is the largest of this family and still not a measurement. Measured 2026-08.",
	"IE00BGCSB447": "The bad path is two months out of 96, and they are real. The engine is a compounded bill " +
		"rate, which has no credit spread and almost no daily variance; the fund is an investment-grade " +
		"ultrashort book. Monthly correlation is 0.29 over the whole window and 0.81 once March 2020 (-3.46 %) " +
		"and its April reversal (+3.54 %) are dropped, after which dropping further months changes nothing. " +
		"Both siblings moved the same day (ICSH -1.5 %, NEAR -6.2 % on 2020-03-19 against this fund's -3.2 %), " +
		"so nothing here is a bad print and no hygiene rule should remove it. The two months net out, the level " +
		"gap is -0.42 %/yr, and real quotes are grafted from 2018, so this governs only the pre-inception tail. " +
		"Measured 2026-08; read the level verdict, not the path.",
	"ERESMONDEM": "The path warning is a clock, not a defect: the fund strikes its NAV on the two ETFs' " +
		"official NAVs, i.e. after New York closes, while the engine's donor years are Xetra closes struck at " +
		"17:30 CET, so every US afternoon lands a day apart (daily correlation 0.62, weekly 0.95, monthly 0.98). " +
		"The level is the a-priori wrapper charge (0.35 % management + 0.06 % transactions, the FY2025 report) and " +
		"the +0.3 pt/yr residual on two and a half years is inside the timing noise; real NAVs are grafted from " +
		"2024-03, so all of this governs only the pre-inception tail. Measured 2026-08; do not retune.",
	"DBMF": "The warn verdicts are the known ceiling of a replication fund, not a defect: the fund against " +
		"its own target index reads 0.85 monthly, so no public donor can beat what the fund itself leaves " +
		"observable, and the negative gap is the manager's replication alpha over the index (net of the " +
		"constituents' performance fees, which the fee uplift deliberately does not claim back). Real quotes " +
		"are grafted from 2019, so all of this governs only the pre-inception tail. Measured 2026-08; do not retune.",
}

// Audit measures one recipe's engine against reality. f must serve the
// components the recipe needs (wrap it with WithRefData for the bundled
// reference series); a reference candidate that f answers with a pofo
// reconstruction rather than a quote is rejected rather than believed, since
// a series the engine already contains cannot arbitrate it.
func Audit(f Fetcher, r Recipe) AuditResult {
	a := AuditResult{ID: r.ID, Name: r.Name, Method: r.Method, Group: groupOf(r.ID), Caveat: auditCaveats[r.ID]}

	engine, err := r.Build(f, ComponentsFrom)
	if err != nil {
		a.Err = "engine failed: " + err.Error()
		a.Level, a.Path, a.Score = VerdictBad, VerdictBad, math.MaxFloat64
		return a
	}

	real, v, ok := reference(f, r, engine, &a)
	if !ok {
		a.Level, a.Path, a.Score = VerdictUnknown, VerdictUnknown, -1
		return a
	}

	a.Start, a.End = v.Start, v.End
	a.Years = v.End.Sub(v.Start).Hours() / 24 / 365.25
	a.Short = a.Years < 2
	a.DailyCorr, a.WeeklyCorr, a.Beta, a.TrackingErr = v.Corr, v.WeeklyCorr, v.Beta, v.TrackingErr
	a.MonthlyCorr = monthlyCorr(engine, real, v.Start, v.End)
	a.CAGRSim, a.CAGRReal = v.CAGRSim, v.CAGRReal
	a.Delta = v.CAGRSim - v.CAGRReal
	a.TotalDrift = math.Pow(1+a.Delta, a.Years) - 1
	a.VolSim, a.WorstSim = windowVol(engine, v.Start, v.End), worstDay(engine, v.Start, v.End)
	a.VolReal, a.WorstReal = windowVol(real, v.Start, v.End), worstDay(real, v.Start, v.End)
	a.Level, a.Path, a.Score = grade(a)

	// What a SIM consumer actually gets: real quotes wherever they exist, the
	// reconstruction only in front of them, whether or not the recipe also
	// splices them into the shipped file.
	if own, err := f.Fetch(r.ID, ComponentsFrom); err == nil && isQuote(own) {
		a.RealFrom = own.First().Date
	}

	a.Engine = clip(engine, v.Start, v.End)
	a.Real = clip(real, v.Start, v.End)
	seen := map[string]bool{upper(a.Reference): true}
	for _, id := range append(comparisonIDs(r), auditExtras[r.ID]...) {
		if seen[upper(id)] {
			continue
		}
		seen[upper(id)] = true
		s, err := f.Fetch(id, ComponentsFrom)
		if err != nil || s == nil || len(s.Points) < 30 {
			continue
		}
		// A curve in another currency would read as a divergence it is not.
		if real.Currency != "" && s.Currency != "" && s.Currency != real.Currency {
			continue
		}
		if c := clip(s, v.Start, v.End); len(c.Points) >= 30 {
			a.Others = append(a.Others, c)
		}
	}
	a.Chain = chainOf(f, r)
	return a
}

// AuditAll audits every recipe in rs and lays them out in reading order,
// worst first inside each family. progress, when non-nil, is called with each
// identifier before it is measured: the whole pass fetches a lot and is slow.
func AuditAll(f Fetcher, rs []Recipe, progress func(id string)) []AuditGroup {
	byGroup := map[string][]AuditResult{}
	for _, r := range rs {
		if progress != nil {
			progress(r.ID)
		}
		a := Audit(f, r)
		byGroup[a.Group] = append(byGroup[a.Group], a)
	}
	var out []AuditGroup
	for _, g := range append(auditGroups, struct {
		Title string
		Note  string
		IDs   []string
	}{Title: otherGroup}) {
		rs := byGroup[g.Title]
		if len(rs) == 0 {
			continue
		}
		sort.SliceStable(rs, func(i, j int) bool { return rs[i].Score > rs[j].Score })
		out = append(out, AuditGroup{Title: g.Title, Note: g.Note, Results: rs})
	}
	return out
}

// groupOf places a recipe in its family.
func groupOf(id string) string {
	for _, g := range auditGroups {
		for _, want := range g.IDs {
			if strings.EqualFold(id, want) {
				return g.Title
			}
		}
	}
	return otherGroup
}

// reference picks the series the engine is judged against: the asset's own
// quotes first, then whatever the recipe validates or splices, then the
// curated siblings. A candidate the engine already contains would read as a
// perfect fit and prove nothing, so it is recorded as rejected instead.
func reference(f Fetcher, r Recipe, engine *marketdata.Series, a *AuditResult) (*marketdata.Series, Validation, bool) {
	seen := map[string]bool{}
	for _, id := range append(append([]string{r.ID}, comparisonIDs(r)...), auditExtras[r.ID]...) {
		if seen[upper(id)] {
			continue
		}
		seen[upper(id)] = true
		s, err := f.Fetch(id, ComponentsFrom)
		if err != nil || s == nil || len(s.Points) < 60 {
			continue
		}
		if !isQuote(s) {
			a.Rejected = append(a.Rejected, id+" (a pofo reconstruction)")
			continue
		}
		v, err := Validate(engine, s)
		if err != nil {
			continue
		}
		if v.TrackingErr < 0.005 && v.Corr > 0.99 {
			a.Rejected = append(a.Rejected, id+" (already inside the engine)")
			continue
		}
		a.Reference = id
		return s, v, true
	}
	a.Err = "no independent reference"
	if len(a.Rejected) > 0 {
		a.Err += ": " + strings.Join(a.Rejected, ", ")
	} else {
		a.Err += ": no real quote could be fetched"
	}
	return nil, Validation{}, false
}

// comparisonIDs returns the identifiers the recipe itself declares.
func comparisonIDs(r Recipe) []string {
	var ids []string
	if r.ValidateAgainst != "" {
		ids = append(ids, r.ValidateAgainst)
	}
	if r.SpliceReal != "" {
		ids = append(ids, r.SpliceReal)
	}
	return ids
}

// chainOf grades every junction of a recipe's declared donor chain: the asset
// against its nearest donor, that donor against the next, down to the deepest.
// The card's own statistics grade the first junction only, because only there
// does the asset itself exist to be the judge; this is where the rest of the
// file answers for itself.
func chainOf(f Fetcher, r Recipe) []Junction {
	if len(r.Donors) == 0 {
		return nil
	}
	ids := append([]string{r.ID}, r.Donors...)
	series := make([]*marketdata.Series, len(ids))
	for i, id := range ids {
		if s, err := f.Fetch(id, ComponentsFrom); err == nil && s != nil && len(s.Points) > 60 {
			series[i] = s
		}
	}
	var out []Junction
	for i := 0; i+1 < len(ids); i++ {
		near, deep := series[i], series[i+1]
		j := Junction{Pair: ids[i+1] + " vs " + ids[i]}
		switch {
		case near == nil || deep == nil:
			j.Note = "series unavailable"
		case near.Currency != "" && deep.Currency != "" && near.Currency != deep.Currency:
			j.Note = "different currencies (the recipe converts or hedges this junction)"
		default:
			j.Span = deep.First().Date.Format("2006") + " to " + near.First().Date.Format("2006")
			from, to := deep.First().Date, near.Last().Date
			xa, xb := pairMonthly(near, deep, from, to)
			j.Months = len(xa)
			if len(xa) >= 12 {
				j.Measured = true
				j.Corr = pearson(xa, xb)
				j.GapYear = annualizeMonthly(xb) - annualizeMonthly(xa)
			} else {
				j.Note = "overlap too short"
			}
		}
		out = append(out, j)
	}
	return out
}

// grade turns the measurements into the two verdicts and a severity score.
func grade(a AuditResult) (level, path Verdict, score float64) {
	gap := math.Abs(a.Delta) * 100
	switch {
	case gap > 2.5:
		level = VerdictBad
	case gap > 1.0:
		level = VerdictWarn
	default:
		level = VerdictOK
	}
	corr := a.MonthlyCorr
	if corr == 0 || math.IsNaN(corr) {
		corr = a.WeeklyCorr
	}
	rel := 1.0
	if a.VolReal > 0 {
		rel = a.TrackingErr / a.VolReal
	}
	switch {
	case corr < 0.75 && rel > 0.6:
		path = VerdictBad
	case corr < 0.90 || rel > 0.4:
		path = VerdictWarn
	default:
		path = VerdictOK
	}
	score = gap + 4*rel
	if a.Short {
		score *= 0.6 // a one-year gap is mostly noise: do not head the list with it
	}
	return level, path, score
}

// isQuote reports whether a series is market data rather than something pofo
// reconstructed (a bundled simdata or reference file, an index rebuild).
func isQuote(s *marketdata.Series) bool {
	return s != nil && len(s.Points) > 0 && s.Source != "index" && s.Source != "simdata"
}

func upper(s string) string { return strings.ToUpper(s) }

// clip returns the series restricted to [from, to], dropping non-positive
// closes; the result is a copy and never nil.
func clip(s *marketdata.Series, from, to time.Time) *marketdata.Series {
	out := &marketdata.Series{Symbol: s.Symbol, Name: s.Name, Currency: s.Currency, Source: s.Source}
	for _, p := range s.Points {
		if p.Date.Before(from) || p.Date.After(to) || p.Close <= 0 {
			continue
		}
		out.Points = append(out.Points, p)
	}
	return out
}

// Drift is the cumulative ratio of the engine to the reference on their common
// dates, rebased at 100: a flat line means the engine tracks, a slope means a
// systematic return gap. It returns nil when the two share no date.
func (a AuditResult) Drift() ([]time.Time, []float64) {
	if a.Engine == nil || a.Real == nil {
		return nil, nil
	}
	byDate := make(map[time.Time]float64, len(a.Real.Points))
	for _, p := range a.Real.Points {
		byDate[p.Date] = p.Close
	}
	var dates []time.Time
	var vals []float64
	for _, p := range a.Engine.Points {
		r, ok := byDate[p.Date]
		if !ok || r <= 0 {
			continue
		}
		dates = append(dates, p.Date)
		vals = append(vals, p.Close/r)
	}
	return dates, Rebase(vals)
}

// Rebase scales a value series to 100 at its first point.
func Rebase(v []float64) []float64 {
	out := make([]float64, len(v))
	if len(v) == 0 || v[0] == 0 {
		return out
	}
	for i, x := range v {
		out[i] = 100 * x / v[0]
	}
	return out
}

// monthlyCorr is the correlation of calendar-month returns over the window,
// the honest yardstick for a reconstruction meant to be held for years: the
// daily and weekly figures are dominated by intra-month texture, which no
// reconstruction of a fifty-market programme can match day by day.
func monthlyCorr(a, b *marketdata.Series, from, to time.Time) float64 {
	xa, xb := pairMonthly(a, b, from, to)
	if len(xa) < 12 {
		return math.NaN()
	}
	return pearson(xa, xb)
}

// pairMonthly returns the two series' calendar-month returns over the months
// they both quote.
func pairMonthly(a, b *marketdata.Series, from, to time.Time) (xa, xb []float64) {
	ma, mb := monthlyReturns(a, from, to), monthlyReturns(b, from, to)
	keys := make([]string, 0, len(ma))
	for k := range ma {
		if _, ok := mb[k]; ok {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	for _, k := range keys {
		xa, xb = append(xa, ma[k]), append(xb, mb[k])
	}
	return xa, xb
}

// monthlyReturns maps "2006-01" to that month's return, measured from the last
// quote of the previous month.
func monthlyReturns(s *marketdata.Series, from, to time.Time) map[string]float64 {
	last := map[string]float64{}
	var keys []string
	for _, p := range s.Points {
		if p.Date.Before(from) || p.Date.After(to) || p.Close <= 0 {
			continue
		}
		k := p.Date.Format("2006-01")
		if _, ok := last[k]; !ok {
			keys = append(keys, k)
		}
		last[k] = p.Close
	}
	out := make(map[string]float64, len(keys))
	for i := 1; i < len(keys); i++ {
		if last[keys[i-1]] > 0 {
			out[keys[i]] = last[keys[i]]/last[keys[i-1]] - 1
		}
	}
	return out
}

// annualizeMonthly compounds monthly returns into a per-year rate.
func annualizeMonthly(r []float64) float64 {
	if len(r) == 0 {
		return 0
	}
	p := 1.0
	for _, x := range r {
		p *= 1 + x
	}
	return math.Pow(p, 12/float64(len(r))) - 1
}

func pearson(a, b []float64) float64 {
	var ma, mb float64
	for i := range a {
		ma, mb = ma+a[i], mb+b[i]
	}
	ma, mb = ma/float64(len(a)), mb/float64(len(b))
	var sab, sa, sb float64
	for i := range a {
		da, db := a[i]-ma, b[i]-mb
		sab, sa, sb = sab+da*db, sa+da*da, sb+db*db
	}
	if sa <= 0 || sb <= 0 {
		return math.NaN()
	}
	return sab / math.Sqrt(sa*sb)
}

// windowVol is the annualized volatility of a series over the window, the
// scale against which a tracking error means something.
func windowVol(s *marketdata.Series, from, to time.Time) float64 {
	c := clip(s, from, to)
	if len(c.Points) < 30 {
		return 0
	}
	r := make([]float64, 0, len(c.Points)-1)
	for i := 1; i < len(c.Points); i++ {
		r = append(r, c.Points[i].Close/c.Points[i-1].Close-1)
	}
	var m float64
	for _, x := range r {
		m += x
	}
	m /= float64(len(r))
	var sum float64
	for _, x := range r {
		sum += (x - m) * (x - m)
	}
	return math.Sqrt(sum/float64(len(r))) * math.Sqrt(252)
}

// worstDay is the worst single-day return over the window: a replication that
// levers into a volatility spike shows it here long before the CAGR does.
func worstDay(s *marketdata.Series, from, to time.Time) float64 {
	c := clip(s, from, to)
	worst := 0.0
	for i := 1; i < len(c.Points); i++ {
		if r := c.Points[i].Close/c.Points[i-1].Close - 1; r < worst {
			worst = r
		}
	}
	return worst
}

// String renders one audit as a single log line.
func (a AuditResult) String() string {
	if !a.Measured() {
		return fmt.Sprintf("%-14s %s", a.ID, a.Err)
	}
	return fmt.Sprintf("%-14s level=%-4s path=%-4s monthly=%.2f gap=%+.2f%%/yr over %.1fy vs %s",
		a.ID, a.Level, a.Path, a.MonthlyCorr, a.Delta*100, a.Years, a.Reference)
}
