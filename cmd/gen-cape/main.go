// Command gen-cape builds the bundled Shiller CAPE series
// (pkg/datasets/cape/shiller-cape.csv) from the datahub mirror of Robert
// Shiller's S&P 500 dataset (the same source as the bundled SP500 history). It
// runs at data-generation time only (network); the pofo binary embeds the CSV.
//
// The datahub mirror stopped computing PE10 in 2023 (its earnings columns are
// blank ever since). Recent months come from multpl.com's monthly Shiller PE
// table, with the original Shiller workbook as a fallback. The mirror keeps
// precedence over both where they overlap.
//
// CAPE (cyclically-adjusted price/earnings, PE10) is the real price divided by
// the 10-year average of real earnings. Its inverse is a first-order estimate of
// the next decade's real return, so it anchors the FIRE explorer's central case
// to today's valuation. Source column: PE10 in the datahub s-and-p-500 dataset.
//
// Usage: gen-cape [-url URL] [-recent URL] [-o path] [-dry]
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	defaultURL = "https://raw.githubusercontent.com/datasets/s-and-p-500/main/data/data.csv"
	recentURL  = "https://www.multpl.com/shiller-pe/table/by-month"
	shillerURL = "https://shillerdata.com/"
)

func main() {
	url := flag.String("url", defaultURL, "datahub s-and-p-500 CSV URL")
	recent := flag.String("recent", recentURL, "multpl monthly Shiller-PE table URL (empty = skip both recent sources)")
	out := flag.String("o", "pkg/datasets/cape/shiller-cape.csv", "output CSV path")
	dry := flag.Bool("dry", false, "print coverage without writing")
	flag.Parse()

	raw, err := download(*url)
	if err != nil {
		log.Fatalf("download: %v", err)
	}
	rows, err := parse(raw)
	if err != nil {
		log.Fatalf("parse: %v", err)
	}
	var sources []string
	if *recent != "" {
		current := time.Now().UTC().Format("2006-01") + "-01"
		rows, sources, err = supplement(rows, current, func() ([]point, error) {
			raw, err := download(*recent)
			if err != nil {
				return nil, err
			}
			return parseMultpl(raw)
		}, fetchShiller)
		if err != nil {
			log.Fatal(err)
		}
		for _, source := range sources {
			log.Print(source)
		}
		if latest := rows[len(rows)-1].date; latest < current {
			log.Printf("latest available CAPE is %s; current month is %s", latest, current)
		}
	}
	last := rows[len(rows)-1]
	log.Printf("CAPE: %d months, %s..%s, latest %.2f", len(rows), rows[0].date, last.date, last.cape)
	logPercentile(rows, last.cape)
	if *dry {
		return
	}
	if err := write(*out, *url, rows, sources); err != nil {
		log.Fatalf("write: %v", err)
	}
	log.Printf("wrote %s", *out)
}

func download(url string) ([]byte, error) {
	c := &http.Client{Timeout: 60 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	// multpl serves an empty page to Go's default agent; a browser UA works.
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)")
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %s", resp.Status)
	}
	return io.ReadAll(resp.Body)
}

// supplement keeps multpl through its last usable month and consults Shiller
// when multpl fails or stops before the current month. A valid but lagging
// publication can still extend the shipped series.
func supplement(mirror []point, current string, multpl, shiller func() ([]point, error)) ([]point, []string, error) {
	rows := mirror
	var sources []string
	for _, s := range []struct {
		name  string
		fetch func() ([]point, error)
	}{
		{"multpl.com monthly Shiller PE (day-of-month price)", multpl},
		{"Shiller ie_data.xls, Data sheet, ordinary CAPE (https://shillerdata.com/)", shiller},
	} {
		last := rows[len(rows)-1].date
		if last >= current {
			break
		}
		tail, err := recentTail(s.fetch, mirror, last, current)
		if err != nil {
			log.Printf("%s: %v", s.name, err)
			continue
		}
		if len(tail) > 0 {
			rows = append(rows, tail...)
			sources = append(sources, fmt.Sprintf("%s..%s: %s", tail[0].date, tail[len(tail)-1].date, s.name))
		}
	}
	if len(rows) == len(mirror) && mirror[len(mirror)-1].date < current {
		return nil, nil, fmt.Errorf("neither recent source extended the mirror beyond %s", mirror[len(mirror)-1].date)
	}
	return rows, sources, nil
}

func recentTail(fetch func() ([]point, error), mirror []point, last, current string) ([]point, error) {
	data, err := fetch()
	if err != nil {
		return nil, err
	}
	if err := checkOverlap(mirror, data); err != nil {
		return nil, err
	}
	i := sort.Search(len(data), func(i int) bool { return data[i].date > last })
	tail := data[i:]
	month, err := time.Parse(time.DateOnly, last)
	if err != nil {
		return nil, err
	}
	for _, p := range tail {
		month = month.AddDate(0, 1, 0)
		if p.date != month.Format(time.DateOnly) || p.date > current {
			return nil, fmt.Errorf("unexpected month %s (expected %s, current %s)", p.date, month.Format(time.DateOnly), current)
		}
	}
	return tail, nil
}

// The sources use different within-month prices, and the mirror's last
// observation can be provisional. Compare a full year's mean absolute
// relative difference to detect a different valuation series.
func checkOverlap(mirror, recent []point) error {
	const months = 12
	if len(mirror) < months {
		return fmt.Errorf("need %d mirror months to check overlap", months)
	}
	year := mirror[len(mirror)-months:]
	start := sort.Search(len(recent), func(i int) bool { return recent[i].date >= year[0].date })
	var gap float64
	for i, m := range year {
		if start+i >= len(recent) || recent[start+i].date != m.date {
			return fmt.Errorf("missing overlap at %s", m.date)
		}
		gap += math.Abs(recent[start+i].cape/m.cape - 1)
	}
	if gap/months > 0.02 {
		return fmt.Errorf("CAPE differs from the mirror by %.2f%% on average over %d months", 100*gap/months, months)
	}
	return nil
}

// multplRow matches one data row of multpl's table: a "Mon d, yyyy" date cell
// followed by a numeric value cell. The value cell starts with an &#x2002;
// space entity whose digits must not be read as the value, hence the explicit
// entity-or-whitespace skip.
var multplRow = regexp.MustCompile(`<td>([A-Z][a-z]{2}) \d{1,2}, (\d{4})</td>\s*<td>(?:\s|&#x[0-9a-fA-F]+;|&[a-z]+;)*([0-9.]+)`)

// parseMultpl extracts (month, CAPE) points from the multpl HTML table,
// normalised to the mirror's yyyy-mm-01 date format, ascending. The table
// lists at most one row per month, newest first.
func parseMultpl(raw []byte) ([]point, error) {
	months := map[string]string{
		"Jan": "01", "Feb": "02", "Mar": "03", "Apr": "04", "May": "05", "Jun": "06",
		"Jul": "07", "Aug": "08", "Sep": "09", "Oct": "10", "Nov": "11", "Dec": "12",
	}
	var out []point
	seen := map[string]bool{}
	for _, m := range multplRow.FindAllStringSubmatch(string(raw), -1) {
		mm, ok := months[m[1]]
		if !ok {
			continue
		}
		v, err := strconv.ParseFloat(m[3], 64)
		if err != nil || v <= 0 {
			continue
		}
		date := fmt.Sprintf("%s-%s-01", m[2], mm)
		if seen[date] {
			continue // newest-first table: the first row of a month is the freshest reading
		}
		seen[date] = true
		out = append(out, point{date: date, cape: v})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no rows matched in the multpl table")
	}
	sort.Slice(out, func(i, j int) bool { return out[i].date < out[j].date })
	return out, nil
}

type point struct {
	date string
	cape float64
}

// parse extracts Date and the PE10 (CAPE) column, keeping only rows with a real
// value (the first decade and the most recent unfilled months carry 0).
func parse(raw []byte) ([]point, error) {
	lines := strings.Split(string(raw), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("empty CSV")
	}
	head := strings.Split(strings.TrimSpace(lines[0]), ",")
	di, pi := indexOf(head, "Date"), indexOf(head, "PE10")
	if di < 0 || pi < 0 {
		return nil, fmt.Errorf("missing Date or PE10 column")
	}
	var out []point
	for _, ln := range lines[1:] {
		f := strings.Split(strings.TrimSpace(ln), ",")
		if len(f) <= max(pi, di) || f[di] == "" {
			continue
		}
		v, err := strconv.ParseFloat(f[pi], 64)
		if err != nil || !(v > 0) || math.IsInf(v, 0) {
			continue
		}
		out = append(out, point{date: f[di], cape: v})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no CAPE values found")
	}
	sort.Slice(out, func(i, j int) bool { return out[i].date < out[j].date })
	return out, nil
}

func write(path, url string, rows []point, sources []string) error {
	// The mirror has ended in 2023 since it stopped computing earnings. A
	// failed or stale supplement must never overwrite months already shipped.
	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	dates := make(map[string]bool, len(rows))
	for _, p := range rows {
		dates[p.date] = true
	}
	for line := range strings.Lines(string(existing)) {
		date, _, ok := strings.Cut(line, ",")
		if ok && len(date) == len(time.DateOnly) && date[4] == '-' && !dates[date] {
			return fmt.Errorf("refusing to lose CAPE for %s already in %s", date, path)
		}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# Shiller CAPE (PE10): real S&P 500 price / 10-year average real earnings.\n")
	fmt.Fprintf(&b, "# Source: datahub mirror of Robert Shiller's dataset (%s).\n", url)
	for _, source := range sources {
		fmt.Fprintf(&b, "# %s\n", source)
	}
	fmt.Fprintln(&b, "# Latest month is provisional.")
	fmt.Fprintf(&b, "# 1/CAPE approximates the next decade's real return. Regenerate: make cape\n")
	fmt.Fprintf(&b, "date,cape\n")
	for _, p := range rows {
		fmt.Fprintf(&b, "%s,%.2f\n", p.date, p.cape)
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".shiller-cape-*")
	if err != nil {
		return err
	}
	defer os.Remove(temp.Name())
	if err := temp.Chmod(0o644); err != nil {
		temp.Close()
		return err
	}
	if _, err := temp.WriteString(b.String()); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(temp.Name(), path)
}

func logPercentile(rows []point, v float64) {
	below := 0
	for _, p := range rows {
		if p.cape < v {
			below++
		}
	}
	pct := 100 * float64(below) / float64(len(rows))
	log.Printf("latest CAPE %.2f = %.0fth percentile; 1/CAPE = %.2f%% implied real", v, pct, 100/v)
}

func indexOf(ss []string, name string) int {
	for i, s := range ss {
		if strings.TrimSpace(s) == name {
			return i
		}
	}
	return -1
}
