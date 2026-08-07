// Command gen-cape builds the bundled Shiller CAPE series
// (pkg/datasets/cape/shiller-cape.csv) from the datahub mirror of Robert
// Shiller's S&P 500 dataset (the same source as the bundled SP500 history). It
// runs at data-generation time only (network); the pofo binary embeds the CSV.
//
// CAPE (cyclically-adjusted price/earnings, PE10) is the real price divided by
// the 10-year average of real earnings. Its inverse is a first-order estimate of
// the next decade's real return, so it anchors the FIRE explorer's central case
// to today's valuation. Source column: PE10 in the datahub s-and-p-500 dataset.
//
// Usage: gen-cape [-url URL] [-o path] [-dry]
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const defaultURL = "https://raw.githubusercontent.com/datasets/s-and-p-500/main/data/data.csv"

func main() {
	url := flag.String("url", defaultURL, "datahub s-and-p-500 CSV URL")
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
	resp, err := c.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %s", resp.Status)
	}
	return io.ReadAll(resp.Body)
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
	fmt.Fprintf(&b, "# Source: datahub mirror of Robert Shiller's dataset (%s).\n", url)
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
