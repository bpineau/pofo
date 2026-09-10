package main

import (
	"bufio"
	"context"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// monthly is a month-end level series keyed "YYYY-MM", which sorts
// chronologically as a string.
type monthly map[string]float64

func (m monthly) keys() []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

func (m monthly) last() string {
	ks := m.keys()
	if len(ks) == 0 {
		return ""
	}
	return ks[len(ks)-1]
}

// completeMonthEnds reduces a daily series to one level per COMPLETE calendar
// month, the last close in the month. The series' own last month is dropped
// whatever the day of the month: a month-end level struck mid-month is not a
// month-end level, and a reference series may not carry one.
func completeMonthEnds(s *marketdata.Series) (monthly, string) {
	type dated struct {
		date  time.Time
		close float64
	}
	seen := make(map[string]dated, len(s.Points)/20)
	for _, p := range s.Points {
		k := month(p.Date)
		if cur, ok := seen[k]; !ok || p.Date.After(cur.date) {
			seen[k] = dated{p.Date, p.Close}
		}
	}
	out := make(monthly, len(seen))
	for k, v := range seen {
		out[k] = v.close
	}
	partial := out.last()
	delete(out, partial)
	return out, partial
}

// result is one candidate proxy measured against the anchors.
type result struct {
	proxy   string // catalog id
	symbol  string // quote symbol, for the provenance line
	ter     float64
	full    stat    // the whole overlap, reported but not gated
	gated   stat    // the graded window, the last gradeWindow months of it
	monthly monthly // the proxy's complete month-end levels
}

// stat is a proxy graded over one window of overlapping monthly returns.
type stat struct {
	from, to string
	months   int
	td       float64 // annualized tracking difference, %/yr, proxy minus anchor
	corr     float64 // Pearson on the monthly returns
	rmse     float64 // root mean square monthly return difference, % per month
}

func (s stat) String() string {
	return fmt.Sprintf("%s..%s (%3d mo) TD %+.3f %%/yr corr %.4f rmse %.3f %%/mo", s.from, s.to, s.months, s.td, s.corr, s.rmse)
}

func (r result) String() string {
	return fmt.Sprintf("+TER %.2f %%/yr  graded %s  |  whole overlap %s", r.ter, r.gated, r.full)
}

// provenance is the one line written into the file's tail-source header.
func (r result) provenance(today time.Time) string {
	name := r.proxy
	if r.symbol != "" && r.symbol != r.proxy {
		name = fmt.Sprintf("%s (%s)", r.proxy, r.symbol)
	}
	return fmt.Sprintf("tail from %s +TER %.2f %%/yr, TD on overlap %s..%s = %+.3f %%/yr, corr %.4f, rmse %.3f %%/mo, validated %s",
		name, r.ter, r.gated.from, r.gated.to, r.gated.td, r.gated.corr, r.gated.rmse, today.Format("2006-01-02"))
}

// measure fetches a candidate proxy and grades it against the anchors. The
// result is filled in even when a guard refuses the candidate, so the caller
// can log the numbers behind the refusal.
func measure(ctx context.Context, client *marketdata.Client, t target, anchors []marketdata.Point, id string) (result, error) {
	meta, ok := marketdata.Lookup(id)
	if !ok {
		return result{proxy: id}, fmt.Errorf("not in the bundled catalog (its ongoing charge must be vetted there)")
	}
	if meta.Fees <= 0 {
		return result{proxy: id}, fmt.Errorf("no ongoing charge pinned in the catalog")
	}
	r := result{proxy: id, symbol: meta.Symbol, ter: meta.Fees}
	s, err := client.Fetch(ctx, id, anchors[0].Date)
	if err != nil {
		return r, fmt.Errorf("fetch: %w", err)
	}
	if s.Currency != t.currency {
		return r, fmt.Errorf("quotes in %s, want %s", s.Currency, t.currency)
	}
	levels, partial := completeMonthEnds(s)
	if partial != "" {
		client.Logf("%s: %s stops mid-month, dropping its incomplete %s", t.id, id, partial)
	}
	r.monthly = levels
	r.full, r.gated, err = compare(anchors, levels, meta.Fees)
	return r, err
}

// compare grades a proxy's month-end levels against the anchors, with the
// proxy's annual ongoing charge added back (the reference is an index, gross
// of any fund fee, while the proxy's adjusted closes are net of it). It is the
// validation spec bar the currency check, which needs the fetched series.
//
// Two windows come back: the whole overlap, for the record, and the last
// gradeWindow months of it, which the bands are applied to. The gated window
// exists because these ETFs' launch years carry visibly bad vendor prints
// (a flat month against a -9.6 % index, and worse), and because a proxy used
// only at the front edge of a reference series should be graded on the regime
// the tail is actually drawn from, not on its first illiquid year.
func compare(anchors []marketdata.Point, proxy monthly, ter float64) (full, gated stat, err error) {
	anchor := make(monthly, len(anchors))
	for _, p := range anchors {
		anchor[month(p.Date)] = p.Close
	}
	if proxy.last() <= anchor.last() {
		return full, gated, fmt.Errorf("stops at %s, the anchors already reach %s", orDash(proxy.last()), anchor.last())
	}
	gross := math.Pow(1+ter/100, 1.0/12) // one month of the charge, added back

	var months []string
	var ra, rp []float64
	keys := proxy.keys()
	for i := 1; i < len(keys); i++ {
		prev, cur := keys[i-1], keys[i]
		a0, ok0 := anchor[prev]
		a1, ok1 := anchor[cur]
		p0, p1 := proxy[prev], proxy[cur]
		if !ok0 || !ok1 || a0 <= 0 || p0 <= 0 {
			continue
		}
		months = append(months, cur)
		ra = append(ra, a1/a0-1)
		rp = append(rp, (p1/p0)*gross-1)
	}
	if len(ra) < minOverlapMonths {
		return full, gated, fmt.Errorf("only %d overlapping monthly returns, want %d", len(ra), minOverlapMonths)
	}
	full = grade(months, ra, rp)
	if n := len(ra); n > gradeWindow {
		gated = grade(months[n-gradeWindow:], ra[n-gradeWindow:], rp[n-gradeWindow:])
	} else {
		gated = full
	}
	switch {
	case math.Abs(gated.td) > maxTD:
		err = fmt.Errorf("tracking difference %+.3f %%/yr over %s..%s exceeds the %.2f point band", gated.td, gated.from, gated.to, maxTD)
	case gated.corr < minCorr:
		err = fmt.Errorf("monthly correlation %.4f over %s..%s below %.3f", gated.corr, gated.from, gated.to, minCorr)
	}
	return full, gated, err
}

// grade is the tracking difference and correlation of one window of paired
// monthly returns (anchor, proxy).
func grade(months []string, ra, rp []float64) stat {
	ga, gp := 1.0, 1.0
	for i := range ra {
		ga *= 1 + ra[i]
		gp *= 1 + rp[i]
	}
	n := float64(len(ra))
	var se float64
	for i := range ra {
		se += (rp[i] - ra[i]) * (rp[i] - ra[i]) / n
	}
	return stat{
		from:   months[0],
		to:     months[len(months)-1],
		months: len(months),
		td:     (math.Pow(gp/ga, 12/n) - 1) * 100,
		corr:   pearson(ra, rp),
		rmse:   math.Sqrt(se) * 100,
	}
}

func pearson(x, y []float64) float64 {
	n := float64(len(x))
	var mx, my float64
	for i := range x {
		mx += x[i] / n
		my += y[i] / n
	}
	var sxy, sxx, syy float64
	for i := range x {
		dx, dy := x[i]-mx, y[i]-my
		sxy += dx * dy
		sxx += dx * dx
		syy += dy * dy
	}
	if sxx <= 0 || syy <= 0 {
		return 0
	}
	return sxy / math.Sqrt(sxx*syy)
}

// buildTail chains the proxy's monthly returns onto the anchors' last level,
// one point per complete month after the anchors, dated on the last CALENDAR
// day of its month (the convention every point of these files follows, so that
// simgen's alignMonthEnd can pin it to the shape's own month-end).
//
// The proxy's ongoing charge is added back, month by month and with the same
// factor compare grades it with: the reference is an index and is gross of any
// fund fee, so a tail chained on the fund's net return would bleed the charge
// into the index.
func buildTail(anchors []marketdata.Point, proxy monthly, ter, maxRet float64) ([]marketdata.Point, error) {
	last := anchors[len(anchors)-1]
	prevKey, prevLevel := month(last.Date), last.Close
	if _, ok := proxy[prevKey]; !ok {
		return nil, fmt.Errorf("proxy has no level for the anchors' last month %s", prevKey)
	}
	gross := math.Pow(1+ter/100, 1.0/12)
	var tail []marketdata.Point
	for _, k := range proxy.keys() {
		if k <= prevKey {
			continue
		}
		r := (proxy[k]/proxy[prevKey])*gross - 1
		if math.Abs(r) > maxRet {
			return nil, fmt.Errorf("appended return %s = %+.2f%% exceeds the %.0f%% sanity bound", k, r*100, maxRet*100)
		}
		prevLevel *= 1 + r
		tail = append(tail, marketdata.Point{Date: calendarMonthEnd(k), Close: prevLevel})
		prevKey = k
	}
	if len(tail) == 0 {
		return nil, fmt.Errorf("nothing to append after %s", prevKey)
	}
	return tail, nil
}

// calendarMonthEnd is the last calendar day of the month "YYYY-MM".
func calendarMonthEnd(k string) time.Time {
	t, _ := time.Parse("2006-01", k)
	return t.AddDate(0, 1, -1).UTC()
}

// refdata is one reference CSV: the headers this generator owns plus the
// points. anchors holds the export's points on read (the tail a previous run
// appended is dropped) and export plus fresh tail on write.
type refdata struct {
	id, name, source, tailSource, tailFrom, generated string
	anchors                                           []marketdata.Point
}

// readRefdata loads a reference CSV and returns its export era only: every
// point strictly before the month named by "# tail-from:". Without that header
// the whole file is the export and the boundary is minted by the caller.
func readRefdata(path string) (refdata, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return refdata{}, err
	}
	var out refdata
	var pts []marketdata.Point
	sc := bufio.NewScanner(strings.NewReader(string(b)))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || line == "date,close" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			key, val, found := strings.Cut(strings.TrimPrefix(line, "#"), ":")
			if !found {
				continue
			}
			if strings.TrimSpace(key) == "tail-from" {
				out.tailFrom = strings.TrimSpace(val)
			}
			continue
		}
		dateStr, closeStr, found := strings.Cut(line, ",")
		if !found {
			return refdata{}, fmt.Errorf("%s: invalid line %q", path, line)
		}
		d, err := time.ParseInLocation("2006-01-02", dateStr, time.UTC)
		if err != nil {
			return refdata{}, fmt.Errorf("%s: invalid date %q", path, dateStr)
		}
		c, err := strconv.ParseFloat(closeStr, 64)
		if err != nil || c <= 0 {
			return refdata{}, fmt.Errorf("%s: invalid close %q", path, closeStr)
		}
		pts = append(pts, marketdata.Point{Date: d, Close: c})
	}
	if err := sc.Err(); err != nil {
		return refdata{}, err
	}
	sort.Slice(pts, func(i, j int) bool { return pts[i].Date.Before(pts[j].Date) })
	for _, p := range pts {
		if out.tailFrom != "" && month(p.Date) >= out.tailFrom {
			break
		}
		out.anchors = append(out.anchors, p)
	}
	if len(out.anchors) == 0 {
		return refdata{}, fmt.Errorf("%s: no export points before tail-from %q", path, out.tailFrom)
	}
	return out, nil
}

// writeRefdata renders a reference CSV in the bundle's format.
func writeRefdata(path string, r refdata) error {
	var b strings.Builder
	b.WriteString("# pofo simdata v1\n")
	fmt.Fprintf(&b, "# id: %s\n", r.id)
	fmt.Fprintf(&b, "# name: %s\n", r.name)
	fmt.Fprintf(&b, "# source: %s\n", r.source)
	fmt.Fprintf(&b, "# tail-source: %s\n", r.tailSource)
	fmt.Fprintf(&b, "# tail-from: %s\n", r.tailFrom)
	fmt.Fprintf(&b, "# generated: %s\n", r.generated)
	b.WriteString("date,close\n")
	for _, p := range r.anchors {
		fmt.Fprintf(&b, "%s,%.6f\n", p.Date.Format("2006-01-02"), p.Close)
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
