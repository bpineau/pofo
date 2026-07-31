// Command gen-cape builds the bundled Shiller CAPE series
// (pkg/datasets/cape/shiller-cape.csv) from the datahub mirror of Robert
// Shiller's S&P 500 dataset (the same source as the bundled SP500 history). It
// runs at data-generation time only (network); the pofo binary embeds the CSV.
//
// The datahub mirror stopped computing PE10 in 2023 (its earnings columns are
// blank ever since), so the recent months are completed from multpl.com's
// monthly Shiller PE table, which tracks Shiller's own updates. The deep
// history keeps the mirror's values; multpl only appends months after the
// mirror's last real PE10.
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
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	defaultURL = "https://raw.githubusercontent.com/datasets/s-and-p-500/main/data/data.csv"
	recentURL  = "https://www.multpl.com/shiller-pe/table/by-month"
)

func main() {
	url := flag.String("url", defaultURL, "datahub s-and-p-500 CSV URL")
	recent := flag.String("recent", recentURL, "multpl monthly Shiller-PE table URL for the months the mirror lacks (empty = skip)")
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
	if *recent != "" {
		added, err := appendRecent(&rows, *recent)
		if err != nil {
			log.Printf("recent months (multpl): %v; keeping the mirror's coverage", err)
		} else {
			log.Printf("appended %d recent months from multpl", added)
		}
	}
	last := rows[len(rows)-1]
	log.Printf("CAPE: %d months, %s..%s, latest %.2f", len(rows), rows[0].date, last.date, last.cape)
	logPercentile(rows, last.cape)
	if *dry {
		return
	}
	if err := write(*out, *url, rows); err != nil {
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

// appendRecent fills the months after the mirror's last PE10 from multpl's
// monthly table, appending in place. multpl's newest row is a mid-month
// current reading; it is kept as that month's value, matching how the FIRE
// explorer reads the last row as "today's valuation".
func appendRecent(rows *[]point, url string) (added int, err error) {
	raw, err := download(url)
	if err != nil {
		return 0, err
	}
	recent, err := parseMultpl(raw)
	if err != nil {
		return 0, err
	}
	lastDate := (*rows)[len(*rows)-1].date
	for _, p := range recent {
		if p.date > lastDate {
			*rows = append(*rows, p)
			added++
		}
	}
	return added, nil
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
		if len(f) <= pi || f[di] == "" {
			continue
		}
		v, err := strconv.ParseFloat(f[pi], 64)
		if err != nil || v <= 0 {
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

func write(path, url string, rows []point) error {
	var b strings.Builder
	fmt.Fprintf(&b, "# Shiller CAPE (PE10): real S&P 500 price / 10-year average real earnings.\n")
	fmt.Fprintf(&b, "# Source: datahub mirror of Robert Shiller's dataset (%s);\n", url)
	fmt.Fprintf(&b, "# months after the mirror's last computed PE10 come from multpl.com's monthly\n")
	fmt.Fprintf(&b, "# Shiller-PE table (day-of-month price, so ~1-3%% off the mirror's monthly-average\n")
	fmt.Fprintf(&b, "# convention: fine for a valuation gauge, do not use for return computations).\n")
	fmt.Fprintf(&b, "# 1/CAPE approximates the next decade's real return. Regenerate: make cape\n")
	fmt.Fprintf(&b, "date,cape\n")
	for _, p := range rows {
		fmt.Fprintf(&b, "%s,%.2f\n", p.date, p.cape)
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
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
