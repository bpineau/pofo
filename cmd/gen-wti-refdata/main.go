// Command gen-wti-refdata rebuilds pkg/datasets/refdata/WTI-ER-USD.csv, the
// daily EXCESS RETURN of a continuously rolled long position in WTI crude oil
// futures.
//
// It runs at data-generation time only (network); the pofo binary embeds the
// CSV and never fetches anything.
//
// WHY IT EXISTS. The bundled WTI series (WTI-USD, WTI-DAILY) are SPOT prices.
// A spot price is not investable: a futures holder earns the spot move plus a
// roll yield, positive when the curve is backwardated and negative when it is
// in contango, and over crude's history that term has been worth far more than
// the spot move itself. Measured on this series, a rolled long earned 9.5
// points a year MORE than spot from 1986 to 2000 and 12.8 points a year LESS
// from 2005 to 2016. Anything that prices an oil position off spot is therefore
// wrong by a double-digit annual rate, in a sign that flips by era.
//
// WHAT THE SERIES IS. An EXCESS RETURN index: the return of the futures
// position alone, with NO collateral interest. The funded (total return)
// variant is this series compounded with a cash rate; TBILL-3M is bundled for
// exactly that, and the validation below does the conversion. Keeping the
// shipped series unfunded follows the same convention as TREND-TSMOM-USD
// against TREND-NET-USD: the return of the strategy, not of the strategy plus
// a money-market fund.
//
// SOURCE. The U.S. Energy Information Administration's daily NYMEX WTI futures
// settlement prices for the first and second nearby contracts, series RCLC1 and
// RCLC2, taken from EIA's own key-free bulk archive
//
//	https://api.eia.gov/bulk/PET.zip
//
// EIA DISCONTINUED these two series after 2024-04-05 and still publishes the
// spot price (RWTC) daily, so the file ends there and cannot be extended from
// this source. That is the reliability bound: the series stops where its
// evidence stops. The published index families that would carry it further
// (S&P GSCI Crude Oil ER, Bloomberg Crude Oil Subindex ER) have no free
// historical download; see the notes in docs/wti-rolled-reference-design.md.
//
// METHOD. A long position in the first nearby contract, rolled into the second
// nearby over the fifth through ninth business day of each month, one fifth of
// the position per day. That is the S&P GSCI roll schedule, and for a
// single-commodity index the GSCI contract daily return reduces to the return
// of that basket. Two mechanical details decide whether it is right:
//
//   - EIA's "contract 1" is a SLOT, not a contract: it renumbers to the next
//     delivery month on the first trading day after the expiring contract's
//     last trading day. On that day the position, which is entirely in slot 2,
//     becomes entirely slot 1, so the day's return is P1(t)/P2(t-1) - 1 and
//     using P1(t)/P1(t-1) would book the whole calendar spread as a price move.
//     renumberDays computes those dates from the CME rule (trading terminates
//     three business days before the 25th calendar day of the month preceding
//     delivery, or before the last business day preceding the 25th when the
//     25th is not one) over the exchange calendar the data itself carries.
//     April 2020 settles the convention beyond doubt: EIA's slot 1 is -37.63 on
//     2020-04-20 and 10.01 on the 21st, the expiring May contract's last two
//     settlements, and only becomes the June contract on the 22nd.
//   - The roll is done at the PREVIOUS close: the weights that price day t are
//     the weights carried into it. The alternative (S&P's published formula
//     applies the same day's roll weights to both legs, which places the roll
//     one day earlier) was measured against the published index and is worse:
//     see checkAgainstGSCI.
//
// Because the roll completes on the ninth business day, around the 13th, the
// position is already in the next contract when the front month expires around
// the 21st. That is why this index, like the real S&P GSCI Crude Oil index,
// never touches the negative May-2020 settlement.
//
// VALIDATION. The reconstruction is checked, every run, against the published
// S&P GSCI Crude Oil Total Return Index year-end levels for 1986 to 2015, taken
// from a Barclays iPath ETN pricing supplement filed with the SEC (see
// gsciTRYearEnd). The check funds the reconstruction with the bundled TBILL-3M
// rate so the two are the same object, then compares calendar-year returns.
// 27 of the 29 years agree within five points and the median divergence is 1.9
// points; the two exceptions are 1990 (+16.5 points, the Gulf War, when the
// front-to-second spread reached double digits a month and any roll-timing
// difference is maximally levered) and 1994 (-6.2 points). The cumulative
// divergence is 1.2 %/yr over the 29 years, most of it 1990 alone.
//
// This is therefore a reconstruction that BEHAVES like the published index, not
// a copy of it, and it is not labelled as one. What it is good for is exactly
// what spot cannot do: price the roll.
//
// Usage: gen-wti-refdata [-src URL] [-dir path] [-dry]
package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
)

const (
	defaultSrc = "https://api.eia.gov/bulk/PET.zip"
	outID      = "WTI-ER-USD"

	// front and next are EIA's series ids for the first and second nearby
	// NYMEX WTI contract settlement prices, daily.
	front = "PET.RCLC1.D"
	next  = "PET.RCLC2.D"

	// rollFirstBD and rollLastBD bracket the roll: the fifth through ninth
	// business day of the month, a fifth of the position each day. This is
	// the S&P GSCI schedule.
	rollFirstBD = 5
	rollLastBD  = 9

	// minPoints is the shortest history worth shipping. The pair quotes from
	// 1985-01-02 (RCLC1 alone reaches 1983-04, but a roll needs both), so
	// anything much under ten thousand days means a truncated download.
	minPoints = 9500
)

func main() {
	src := flag.String("src", defaultSrc, "EIA petroleum bulk archive")
	dir := flag.String("dir", "pkg/datasets/refdata", "output directory")
	dry := flag.Bool("dry", false, "report without writing")
	flag.Parse()

	raw, err := download(*src)
	if err != nil {
		log.Fatalf("download: %v", err)
	}
	log.Printf("downloaded %d MiB from %s", len(raw)>>20, *src)
	series, err := extract(raw, front, next)
	if err != nil {
		log.Fatalf("extract: %v", err)
	}
	for id, s := range series {
		log.Printf("%s: %d observations, %s to %s", id, len(s),
			minDate(s).Format("2006-01-02"), maxDate(s).Format("2006-01-02"))
	}

	idx, err := build(series[front], series[next])
	if err != nil {
		log.Fatalf("build: %v", err)
	}
	if err := check(idx); err != nil {
		log.Fatalf("sanity: %v", err)
	}
	if err := checkAgainstGSCI(idx); err != nil {
		log.Fatalf("published-index cross-check: %v", err)
	}

	var b bytes.Buffer
	fmt.Fprintf(&b, "# pofo simdata v1\n# id: %s\n", outID)
	fmt.Fprintf(&b, "# name: WTI crude oil, rolled futures EXCESS return (USD, daily, base 100)\n")
	fmt.Fprintf(&b, "# source: long first-nearby NYMEX WTI, rolled into the second nearby over the fifth "+
		"to ninth business day of each month (the S&P GSCI schedule), from EIA daily settlement prices "+
		"RCLC1/RCLC2 (%s). EXCESS return: the futures position alone, NO collateral interest; fund it "+
		"with TBILL-3M for a total return. EIA discontinued both contract series after 2024-04-05, which "+
		"is where this file ends. Regenerate with cmd/gen-wti-refdata.\n", *src)
	fmt.Fprintf(&b, "date,close\n")
	for _, p := range idx {
		fmt.Fprintf(&b, "%s,%.6f\n", p.date.Format("2006-01-02"), p.level)
	}
	if *dry {
		log.Printf("dry run: %d bytes, %d points, final level %.2f", b.Len(), len(idx), idx[len(idx)-1].level)
		return
	}
	out := filepath.Join(*dir, outID+".csv")
	if err := os.WriteFile(out, b.Bytes(), 0o644); err != nil {
		log.Fatalf("write: %v", err)
	}
	log.Printf("wrote %s (%d points)", out, len(idx))
}

// point is one day of the index.
type point struct {
	date  time.Time
	level float64
}

func download(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)")
	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http %d", resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 256<<20))
}

// bulkRecord is the shape of one line of EIA's bulk archive: a JSON document
// per series, whose data is a list of [period, value] pairs, NEWEST FIRST, with
// nulls for days the series does not quote.
type bulkRecord struct {
	SeriesID string               `json:"series_id"`
	Name     string               `json:"name"`
	Units    string               `json:"units"`
	Data     [][2]json.RawMessage `json:"data"`
}

// extract pulls the wanted series out of the bulk archive. The archive holds a
// single 350 MB JSON-lines member, so it is scanned rather than parsed whole,
// and only lines that mention a wanted id are decoded.
func extract(archive []byte, ids ...string) (map[string]map[time.Time]float64, error) {
	zr, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return nil, err
	}
	if len(zr.File) != 1 {
		return nil, fmt.Errorf("archive holds %d members, want 1", len(zr.File))
	}
	f, err := zr.File[0].Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()

	want := make(map[string]bool, len(ids))
	for _, id := range ids {
		want[id] = true
	}
	out := make(map[string]map[time.Time]float64, len(ids))
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 1<<20), 64<<20)
	for sc.Scan() {
		line := sc.Bytes()
		hit := false
		for _, id := range ids {
			if bytes.Contains(line, []byte(`"`+id+`"`)) {
				hit = true
				break
			}
		}
		if !hit {
			continue
		}
		var rec bulkRecord
		if err := json.Unmarshal(line, &rec); err != nil {
			continue
		}
		if !want[rec.SeriesID] {
			continue
		}
		obs, err := observations(rec)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", rec.SeriesID, err)
		}
		log.Printf("%s: %q, units %q", rec.SeriesID, rec.Name, rec.Units)
		out[rec.SeriesID] = obs
		if len(out) == len(ids) {
			break
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	for _, id := range ids {
		if len(out[id]) == 0 {
			return nil, fmt.Errorf("%s not found in the archive", id)
		}
	}
	return out, nil
}

// observations turns a bulk record's [period, value] pairs into a date-keyed
// map. Periods are YYYYMMDD strings; a null value means the series did not
// quote that day and is dropped rather than carried forward.
func observations(rec bulkRecord) (map[time.Time]float64, error) {
	out := make(map[time.Time]float64, len(rec.Data))
	for _, pair := range rec.Data {
		var period string
		if err := json.Unmarshal(pair[0], &period); err != nil {
			return nil, err
		}
		if len(period) != 8 {
			return nil, fmt.Errorf("%q is not a daily period", period)
		}
		d, err := time.Parse("20060102", period)
		if err != nil {
			return nil, err
		}
		// A day the series does not quote arrives as null, and unmarshalling
		// null into a float64 succeeds while leaving it at zero, which would
		// enter the index as a free barrel of oil. Decode through a pointer so
		// the absence stays an absence.
		var v *float64
		if err := json.Unmarshal(pair[1], &v); err != nil || v == nil {
			continue
		}
		out[d.UTC()] = *v
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no observations")
	}
	return out, nil
}

// build compounds the daily excess return of the rolled position into a
// base-100 index over the days BOTH contracts quote.
func build(c1, c2 map[time.Time]float64) ([]point, error) {
	dates := commonDates(c1, c2)
	if len(dates) < 2 {
		return nil, fmt.Errorf("%d common quote days", len(dates))
	}
	renum := renumberDays(dates)
	bd := businessDayOfMonth(dates)

	out := make([]point, 0, len(dates))
	level := 100.0
	out = append(out, point{dates[0], level})
	for i := 1; i < len(dates); i++ {
		t, p := dates[i], dates[i-1]
		w := rollWeight(bd[p]) // weights carried into day t
		var r float64
		switch {
		case renum[t]:
			// The slot renumbered overnight: what was slot 2 yesterday is
			// slot 1 today, and the roll has long since put the whole
			// position there.
			if w != 1 {
				return nil, fmt.Errorf("%s renumbers while %.0f%% of the position is still in the front slot",
					t.Format("2006-01-02"), (1-w)*100)
			}
			r = c1[t]/c2[p] - 1
		default:
			now := w*c2[t] + (1-w)*c1[t]
			was := w*c2[p] + (1-w)*c1[p]
			if now <= 0 || was <= 0 {
				return nil, fmt.Errorf("%s: basket priced at %.2f (was %.2f)", t.Format("2006-01-02"), now, was)
			}
			r = now/was - 1
		}
		level *= 1 + r
		if level <= 0 || math.IsNaN(level) || math.IsInf(level, 0) {
			return nil, fmt.Errorf("%s: index level %v", t.Format("2006-01-02"), level)
		}
		out = append(out, point{t, level})
	}
	return out, nil
}

// commonDates is the sorted list of days both contracts quote. A roll needs
// both legs priced, so a day either leg misses is skipped rather than filled.
func commonDates(c1, c2 map[time.Time]float64) []time.Time {
	out := make([]time.Time, 0, len(c1))
	for d := range c1 {
		if _, ok := c2[d]; ok {
			out = append(out, d)
		}
	}
	slices.SortFunc(out, func(a, b time.Time) int { return a.Compare(b) })
	return out
}

// businessDayOfMonth numbers each quote day within its calendar month, starting
// at 1. The exchange calendar is taken from the data itself: a day the contracts
// did not settle is a day the exchange was shut.
func businessDayOfMonth(dates []time.Time) map[time.Time]int {
	out := make(map[time.Time]int, len(dates))
	n, cur := 0, ""
	for _, d := range dates {
		if m := d.Format("2006-01"); m != cur {
			cur, n = m, 0
		}
		n++
		out[d] = n
	}
	return out
}

// rollWeight is the share of the position sitting in the SECOND nearby contract
// at the close of a day that is the nth business day of its month: nothing
// before the roll starts, a fifth more on each of its five days, all of it once
// the roll is done.
func rollWeight(n int) float64 {
	switch {
	case n < rollFirstBD:
		return 0
	case n >= rollLastBD:
		return 1
	default:
		return float64(n-rollFirstBD+1) / float64(rollLastBD-rollFirstBD+1)
	}
}

// renumberDays marks the days on which EIA's contract slots shift by one, that
// is the first trading day after the front contract stopped trading.
//
// CME terminates WTI trading three business days before the 25th calendar day
// of the month preceding delivery; when the 25th is not a business day, three
// business days before the last business day preceding it. Both are computed
// over the exchange calendar carried by dates, so exchange holidays need no
// separate table.
func renumberDays(dates []time.Time) map[time.Time]bool {
	last := dates[len(dates)-1]
	// reference[m] is the index of the last quote day of month m falling on or
	// before its 25th: the day CME counts three business days back from.
	reference := make(map[string]int, len(dates)/20)
	for i, d := range dates {
		if d.Day() <= 25 {
			reference[d.Format("2006-01")] = i
		}
	}
	out := make(map[time.Time]bool, len(reference))
	for m, i := range reference {
		// A month whose data stops before the 25th has no reference day: the
		// series was truncated there, not renumbered.
		cut, err := time.Parse("2006-01", m)
		if err != nil || !last.After(cut.AddDate(0, 0, 24)) {
			continue
		}
		if i-2 < 1 {
			continue
		}
		out[dates[i-2]] = true // the day after termination, which is dates[i-3]
	}
	return out
}

// check refuses to ship a series that is not what a rolled crude position looks
// like. The bands are wide on purpose: they catch a truncated download, a unit
// change or a broken roll, not a market that surprised us.
func check(idx []point) error {
	if len(idx) < minPoints {
		return fmt.Errorf("only %d daily points, want at least %d", len(idx), minPoints)
	}
	first, lastP := idx[0], idx[len(idx)-1]
	if first.date.After(time.Date(1985, 6, 1, 0, 0, 0, 0, time.UTC)) {
		return fmt.Errorf("history starts %s, want 1985", first.date.Format("2006-01-02"))
	}
	if lastP.date.Before(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)) {
		return fmt.Errorf("history ends %s, want 2024 or later", lastP.date.Format("2006-01-02"))
	}
	// One gap is expected and known: the exchange shut for a week after
	// 2001-09-11. Anything else means a hole in the download.
	for i := 1; i < len(idx); i++ {
		gap := idx[i].date.Sub(idx[i-1].date).Hours() / 24
		if gap > 6 && !(idx[i-1].date.Year() == 2001 && idx[i-1].date.Month() == time.September) {
			return fmt.Errorf("%s follows %s, %0.f days later",
				idx[i].date.Format("2006-01-02"), idx[i-1].date.Format("2006-01-02"), gap)
		}
	}
	rets := make([]float64, len(idx)-1)
	for i := 1; i < len(idx); i++ {
		rets[i-1] = idx[i].level/idx[i-1].level - 1
	}
	vol := stdev(rets) * math.Sqrt(252)
	years := lastP.date.Sub(first.date).Hours() / 24 / 365.25
	cagr := math.Pow(lastP.level/first.level, 1/years) - 1
	log.Printf("%d days, %s to %s, CAGR %+.2f%%/yr, vol %.1f%%/yr, worst day %.1f%%, best day %+.1f%%",
		len(idx), first.date.Format("2006-01-02"), lastP.date.Format("2006-01-02"),
		cagr*100, vol*100, slices.Min(rets)*100, slices.Max(rets)*100)
	if vol < 0.25 || vol > 0.60 {
		return fmt.Errorf("volatility %.1f%%/yr outside [25, 60]: this is not front-month crude", vol*100)
	}
	if cagr < -0.05 || cagr > 0.08 {
		return fmt.Errorf("CAGR %+.2f%%/yr outside [-5, 8]", cagr*100)
	}
	if worst := slices.Min(rets); worst < -0.60 {
		return fmt.Errorf("worst day %.1f%%: the roll should have carried the position out of the expiring contract", worst*100)
	}
	return nil
}

// gsciTRYearEnd is the published S&P GSCI Crude Oil TOTAL RETURN Index at each
// year end, base 100 on 1986-12-31, as filed with the SEC by Barclays Bank PLC
// in the pricing supplement of its iPath S&P GSCI Crude Oil ETN (form 424B2,
// 2016-11-17, accession 0001104659-16-157831, "Historical and Hypothetical
// Historical Performance of the Index"). The index sponsor began calculating it
// on 1991-05-01; the levels before that are the sponsor's own backtest.
//
// It is a TOTAL return: the filing defines each day's move as the contract
// daily return PLUS the Treasury bill return. checkAgainstGSCI funds the
// reconstruction the same way before comparing.
var gsciTRYearEnd = map[int]float64{
	1986: 100.00, 1987: 111.26, 1988: 128.86, 1989: 250.34, 1990: 366.73,
	1991: 305.95, 1992: 317.11, 1993: 205.34, 1994: 285.04, 1995: 380.84,
	1996: 793.33, 1997: 546.93, 1998: 286.55, 1999: 638.13, 2000: 961.62,
	2001: 716.96, 2002: 1134.98, 2003: 1446.80, 2004: 2094.82, 2005: 2538.80,
	2006: 2108.48, 2007: 3108.99, 2008: 1384.47, 2009: 1483.50, 2010: 1481.89,
	2011: 1462.50, 2012: 1293.96, 2013: 1371.77, 2014: 787.99, 2015: 430.74,
}

// Gates on the cross-check. They sit just outside what the reconstruction
// measured on 2026-08-21 (27 years within five points, median divergence 1.9
// points, cumulative drift 1.2 %/yr), so a refresh that quietly changes the
// data or the roll fails loudly while ordinary revisions do not.
const (
	minYearsWithin5pt = 25
	maxMedianDiff     = 0.030
	maxCumulativeDiff = 0.020
)

// checkAgainstGSCI compares the reconstruction with the published index. It
// funds the excess return with the bundled 3-month T-bill rate, because the
// published series is a total return and the two must be the same object before
// their calendar-year returns mean anything.
func checkAgainstGSCI(idx []point) error {
	tbill, err := tbillDaily()
	if err != nil {
		return err
	}
	// Fund the excess return day by day, accruing on every calendar day so a
	// weekend earns its interest, as the published index does.
	yearEnd := make(map[int]float64)
	tr := idx[0].level
	yearEnd[idx[0].date.Year()] = tr
	for i := 1; i < len(idx); i++ {
		t, p := idx[i].date, idx[i-1].date
		tr *= idx[i].level / idx[i-1].level
		for d := p.AddDate(0, 0, 1); !d.After(t); d = d.AddDate(0, 0, 1) {
			tr *= 1 + tbill(d)
		}
		yearEnd[t.Year()] = tr
	}

	years := make([]int, 0, len(gsciTRYearEnd))
	for y := range gsciTRYearEnd {
		if _, ok := gsciTRYearEnd[y-1]; ok {
			if _, ok := yearEnd[y]; ok {
				if _, ok := yearEnd[y-1]; ok {
					years = append(years, y)
				}
			}
		}
	}
	slices.Sort(years)
	if len(years) < 25 {
		return fmt.Errorf("only %d comparable calendar years", len(years))
	}
	diffs := make([]float64, 0, len(years))
	within := 0
	for _, y := range years {
		published := gsciTRYearEnd[y]/gsciTRYearEnd[y-1] - 1
		mine := yearEnd[y]/yearEnd[y-1] - 1
		d := (1+mine)/(1+published) - 1
		diffs = append(diffs, d)
		if math.Abs(d) <= 0.05 {
			within++
		}
		log.Printf("cross-check %d: published %+7.2f%%, rebuilt %+7.2f%%, divergence %+6.2f pt",
			y, published*100, mine*100, d*100)
	}
	abs := make([]float64, len(diffs))
	for i, d := range diffs {
		abs[i] = math.Abs(d)
	}
	slices.Sort(abs)
	median := abs[len(abs)/2]
	first, lastY := years[0]-1, years[len(years)-1]
	span := float64(lastY - first)
	cumulative := math.Pow((yearEnd[lastY]/yearEnd[first])/(gsciTRYearEnd[lastY]/gsciTRYearEnd[first]), 1/span) - 1
	log.Printf("cross-check: %d/%d years within 5 pt, median divergence %.2f pt, cumulative %+.2f%%/yr over %.0f years",
		within, len(years), median*100, cumulative*100, span)

	if within < minYearsWithin5pt {
		return fmt.Errorf("only %d of %d calendar years within 5 points of the published index", within, len(years))
	}
	if median > maxMedianDiff {
		return fmt.Errorf("median calendar-year divergence %.2f pt, want at most %.2f", median*100, maxMedianDiff*100)
	}
	if math.Abs(cumulative) > maxCumulativeDiff {
		return fmt.Errorf("cumulative divergence %+.2f%%/yr, want at most %.2f", cumulative*100, maxCumulativeDiff*100)
	}
	return nil
}

// tbillDaily reads the bundled 3-month T-bill rate and returns the per-calendar-
// day accrual on any date. The published rate is a DISCOUNT rate in percent on
// a 91-day bill, which is the convention the S&P GSCI's own Treasury bill
// return uses, hence the discount-to-price inversion before the daily root.
func tbillDaily() (func(time.Time) float64, error) {
	s, ok, err := marketdata.ReadSimdataFS(datasets.Refdata(), "TBILL-3M")
	if err != nil || !ok {
		return nil, fmt.Errorf("TBILL-3M: ok=%v err=%v", ok, err)
	}
	byMonth := make(map[string]float64, len(s.Points))
	var latest float64
	for _, p := range s.Points {
		byMonth[p.Date.Format("2006-01")] = p.Close
		latest = p.Close
	}
	return func(d time.Time) float64 {
		r, ok := byMonth[d.Format("2006-01")]
		if !ok {
			r = latest
		}
		return math.Pow(1/(1-91.0/360.0*r/100), 1.0/91.0) - 1
	}, nil
}

func stdev(xs []float64) float64 {
	if len(xs) < 2 {
		return 0
	}
	var m float64
	for _, x := range xs {
		m += x
	}
	m /= float64(len(xs))
	var s float64
	for _, x := range xs {
		s += (x - m) * (x - m)
	}
	return math.Sqrt(s / float64(len(xs)-1))
}

func minDate(m map[time.Time]float64) time.Time {
	var out time.Time
	for d := range m {
		if out.IsZero() || d.Before(out) {
			out = d
		}
	}
	return out
}

func maxDate(m map[time.Time]float64) time.Time {
	var out time.Time
	for d := range m {
		if d.After(out) {
			out = d
		}
	}
	return out
}
