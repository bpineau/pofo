package replay

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/scenario"
)

// The textbook portfolio of the withdrawal literature: a US 60/40 of the
// S&P 500 and intermediate (5-year) Treasuries, rebalanced every January and
// deflated by the US CPI-U. It is the yardstick Bengen, the Trinity study and
// every guardrails paper were calibrated on, so replaying the spending rules on
// it puts them on the ground they were designed for.
//
// Everything here is bundled: the two nominal total-return series come from
// pkg/datasets refdata and the deflator from the marketdata CPI snapshot, so
// a replay builds offline and identically on every machine.
const (
	refEquityID     = "SP500-USD"        // S&P 500 total return, month-end
	refBondID       = "TREASURY-INT-USD" // 5-year Treasury total return, monthly
	refEquityWeight = 0.60               // the "60" of the 60/40

	// Label names the reference portfolio for a caption or a chart title.
	Label = "US 60/40 (S&P 500 + 5-year Treasuries, rebalanced yearly)"
	// LabelFR is Label in French, for the book.
	LabelFR = "60/40 américain (S&P 500 + Treasuries 5 ans, rééquilibré chaque janvier)"
)

// Series is the reference portfolio in real terms: a month-end index
// (Index[0] = 100 at Dates[0]) plus the calendar-year real returns derived from
// it. Only complete calendar years reach Years/Annual, so a partial first or
// last year never enters a replay.
type Series struct {
	Dates  []time.Time // month-end, ascending
	Index  []float64   // real total-return index, base 100
	Years  []int       // complete calendar years, ascending
	Annual []float64   // real return of Years[i] (0.07 = +7 %)
}

var (
	refOnce sync.Once
	refData Series
	refErr  error
)

// Reference builds (once) the bundled real 60/40 series. The error is sticky:
// a broken bundle is a build problem, not a transient one.
func Reference() (Series, error) {
	refOnce.Do(func() { refData, refErr = buildReference() })
	return refData, refErr
}

// buildReference chains the two nominal sleeves into a yearly-rebalanced 60/40,
// deflates it by the US CPI and folds the monthly real returns into calendar
// years.
func buildReference() (Series, error) {
	eq, err := refMonthly(refEquityID)
	if err != nil {
		return Series{}, err
	}
	bd, err := refMonthly(refBondID)
	if err != nil {
		return Series{}, err
	}
	months := commonMonths(eq, bd)
	if len(months) < 24 {
		return Series{}, fmt.Errorf("reference 60/40: only %d common months", len(months))
	}

	// Nominal index: the two sleeves drift with their own returns during the
	// year and are reset to 60/40 at the January step, the standard annual
	// rebalancing convention of the literature.
	nominal := make([]marketdata.Point, 0, len(months))
	e, b := refEquityWeight*100, (1-refEquityWeight)*100
	nominal = append(nominal, marketdata.Point{Date: monthEnd(months[0]), Close: e + b})
	for i := 1; i < len(months); i++ {
		if monthOf(months[i]) == time.January {
			total := e + b
			e, b = refEquityWeight*total, (1-refEquityWeight)*total
		}
		e *= eq[months[i]] / eq[months[i-1]]
		b *= bd[months[i]] / bd[months[i-1]]
		nominal = append(nominal, marketdata.Point{Date: monthEnd(months[i]), Close: e + b})
	}

	cpi, err := referenceCPI()
	if err != nil {
		return Series{}, err
	}
	real := scenario.Deflate(nominal, cpi) // len(nominal)-1 monthly real returns

	out := Series{
		Dates: make([]time.Time, len(nominal)),
		Index: make([]float64, len(nominal)),
	}
	out.Dates[0], out.Index[0] = nominal[0].Date, 100
	for i, r := range real {
		out.Dates[i+1] = nominal[i+1].Date
		out.Index[i+1] = out.Index[i] * (1 + r)
	}
	out.Years, out.Annual = foldYears(out.Dates, out.Index)
	return out, nil
}

// refMonthly reads one bundled reference series and keys its levels by month.
// The two files carry different date conventions (month-end closes for the
// index, first-of-month observations for the CMT-derived bond track), so the
// month key, not the day, is what aligns them.
func refMonthly(id string) (map[int]float64, error) {
	s, ok, err := marketdata.ReadSimdataFS(datasets.Refdata(), id)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("bundled reference series %s not found", id)
	}
	out := make(map[int]float64, len(s.Points))
	for _, p := range s.Points {
		if p.Close > 0 {
			out[monthKey(p.Date)] = p.Close // last observation of the month wins
		}
	}
	return out, nil
}

// monthKey maps a date to a dense month index (year*12 + month), so consecutive
// calendar months differ by exactly one; monthOf and monthEnd invert it.
func monthKey(t time.Time) int { return t.Year()*12 + int(t.Month()) - 1 }

func monthOf(key int) time.Month { return time.Month(key%12 + 1) }

func monthEnd(key int) time.Time {
	return time.Date(key/12, monthOf(key), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, -1)
}

// commonMonths lists the months both sleeves quote, ascending and contiguous
// from the first common month (a hole ends the run: chaining across a gap
// would invent a return).
func commonMonths(a, b map[int]float64) []int {
	first, ok := 0, false
	for k := range a {
		if _, in := b[k]; in && (!ok || k < first) {
			first, ok = k, true
		}
	}
	if !ok {
		return nil
	}
	var out []int
	for k := first; ; k++ {
		if _, in := a[k]; !in {
			break
		}
		if _, in := b[k]; !in {
			break
		}
		out = append(out, k)
	}
	return out
}

// foldYears compounds the monthly real index into calendar-year returns,
// keeping only years covered end to end (a December-to-December step for each
// of their twelve months).
func foldYears(dates []time.Time, index []float64) ([]int, []float64) {
	var years []int
	var rets []float64
	for i := 1; i < len(index); i++ {
		y := dates[i].Year()
		if dates[i].Month() != time.December {
			continue
		}
		// December closes the year: find the previous December, twelve months back.
		j := i - 12
		if j < 0 || dates[j].Year() != y-1 || dates[j].Month() != time.December {
			continue
		}
		years = append(years, y)
		rets = append(rets, index[i]/index[j]-1)
	}
	return years, rets
}

// referenceCPI returns the bundled US CPI-U levels. marketdata serves ^CPI-US
// offline-first from its embedded snapshot, so this never reaches the network;
// the cache-less client makes that explicit.
func referenceCPI() ([]marketdata.Point, error) {
	s, err := marketdata.NewClient("").Fetch(context.Background(), "^CPI-US", time.Time{})
	if err != nil {
		return nil, err
	}
	return s.Points, nil
}

// Window returns the slice of the real index starting at the December that
// opens startYear (so the first plotted step is startYear's own market), capped
// at years calendar years, together with those years' real returns. It reports
// ok=false when the record does not cover that start.
func (r Series) Window(startYear, years int) (dates []time.Time, index []float64, seq scenario.Sequence, ok bool) {
	base := -1
	for i, d := range r.Dates {
		if d.Year() == startYear-1 && d.Month() == time.December {
			base = i
			break
		}
	}
	if base < 0 {
		return nil, nil, nil, false
	}
	for i, y := range r.Years {
		if y >= startYear && len(seq) < years {
			seq = append(seq, r.Annual[i])
		}
	}
	if len(seq) == 0 {
		return nil, nil, nil, false
	}
	// Stop the index at the December closing the last year returned, so the
	// curve and the sequence describe exactly the same span.
	end := base + 12*len(seq)
	dates = r.Dates[base : end+1]
	index = make([]float64, len(dates))
	for i, v := range r.Index[base : end+1] {
		index[i] = v / r.Index[base] * 100
	}
	return dates, index, seq, true
}
