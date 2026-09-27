package refgen

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// FREDBase is the public FRED site; its key-less graph endpoint serves every
// series as a two-column CSV.
const FREDBase = "https://fred.stlouisfed.org"

// client bounds every download; the largest source read through it is a few
// megabytes.
var client = &http.Client{Timeout: 120 * time.Second}

// fredClient speaks HTTP/1.1 only, and fredAgent is not a browser's: FRED's
// edge resets an HTTP/2 stream from Go's transport (INTERNAL_ERROR) and leaves
// a browser User-Agent over HTTP/1.1 hanging until the timeout, so a plain
// agent over HTTP/1.1 is the one combination that answers. pkg/marketdata's
// own FRED client carries the same measurement (2026-08).
var fredClient = &http.Client{
	Timeout:   120 * time.Second,
	Transport: &http.Transport{TLSNextProto: map[string]func(string, *tls.Conn) http.RoundTripper{}},
}

const (
	browserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)"
	fredAgent    = "pofo/1 (+https://github.com/bpineau/pofo)"
)

// Get downloads url with a browser User-Agent, which some academic hosts
// require, and returns the body (at most 64 MiB).
func Get(url string) ([]byte, error) {
	return get(client, url, browserAgent)
}

// Post sends body to url as contentType with the same browser User-Agent and
// returns the response body (at most 64 MiB): the chart APIs that take their
// query as a JSON document rather than in the URL.
func Post(url, contentType string, body []byte) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return do(client, req, browserAgent)
}

func get(c *http.Client, url, agent string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return do(c, req, agent)
}

func do(c *http.Client, req *http.Request, agent string) ([]byte, error) {
	url := req.URL.String()
	req.Header.Set("User-Agent", agent)
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: HTTP %d", url, resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 64<<20))
}

// FRED reads one series from FRED's key-less graph endpoint under base
// (FREDBase outside tests), skipping the "." rows it prints for a missing
// observation. Dates are kept as FRED labels them: a monthly series arrives on
// the first of each month.
func FRED(base, id string) ([]marketdata.Point, error) {
	body, err := get(fredClient, fmt.Sprintf("%s/graph/fredgraph.csv?id=%s", base, id), fredAgent)
	if err != nil {
		return nil, fmt.Errorf("FRED %s: %w", id, err)
	}
	var out []marketdata.Point
	for i, line := range strings.Split(string(body), "\n") {
		label, value, ok := strings.Cut(strings.TrimSpace(line), ",")
		if !ok || i == 0 || value == "." || value == "" {
			continue // the observation_date,<id> header, a blank line, a gap
		}
		t, err := time.Parse(time.DateOnly, label)
		if err != nil {
			return nil, fmt.Errorf("FRED %s: line %d: bad date %q", id, i+1, label)
		}
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, fmt.Errorf("FRED %s: line %d: bad value %q", id, i+1, value)
		}
		if n := len(out); n > 0 && !t.After(out[n-1].Date) {
			return nil, fmt.Errorf("FRED %s: %s does not follow %s", id, label, out[n-1].Date.Format(time.DateOnly))
		}
		out = append(out, marketdata.Point{Date: t, Close: v})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("FRED %s: no observation", id)
	}
	return out, nil
}

// sameValue is how close two values must be to count as the same observation:
// half a unit of the sixth decimal the bundle writes, so a value read back from
// a bundled file equals the one its source prints.
const sameValue = 5e-7

// SameHistory checks that a fresh download reproduces the bundled file it is
// about to replace: every date old carries is still there, with the same value,
// except in old's last `revisable` points, where a publisher's revision of its
// latest prints is legitimate and is reported rather than refused. It returns a
// one-line account for the log.
//
// It is the check a refresh of an unrevised source (a market rate, a price fix)
// is judged by: anything else moving means the source changed its definition,
// its units or its rows, and every consumer of the old history would move with
// it unseen.
func SameHistory(old, fresh []marketdata.Point, revisable int) (string, error) {
	byDate := make(map[time.Time]float64, len(fresh))
	for _, p := range fresh {
		byDate[p.Date] = p.Close
	}
	firm := len(old) - revisable
	var revised []string
	for i, p := range old {
		v, ok := byDate[p.Date]
		switch {
		case !ok && i < firm:
			return "", fmt.Errorf("%s is no longer in the source", p.Date.Format(time.DateOnly))
		case !ok:
			revised = append(revised, p.Date.Format(time.DateOnly)+" withdrawn")
		case math.Abs(v-p.Close) <= sameValue:
		case i < firm:
			return "", fmt.Errorf("%s moved from %g to %g", p.Date.Format(time.DateOnly), p.Close, v)
		default:
			revised = append(revised, fmt.Sprintf("%s %g -> %g", p.Date.Format(time.DateOnly), p.Close, v))
		}
	}
	added := 0
	if n := len(old); n > 0 {
		for _, p := range fresh {
			if p.Date.After(old[n-1].Date) {
				added++
			}
		}
	}
	out := fmt.Sprintf("%d bundled points reproduced, %d new", len(old)-len(revised), added)
	if len(revised) > 0 {
		out += "; revised at the tail: " + strings.Join(revised, ", ")
	}
	return out, nil
}

// MonthlyCadence checks that pts is one observation per calendar month, each on
// the first of its month, with no month missing: the layout FRED gives its
// monthly averages, and the one their consumers accrue forward from.
func MonthlyCadence(pts []marketdata.Point) error {
	for i, p := range pts {
		if p.Date.Day() != 1 {
			return fmt.Errorf("%s is not the first of its month", p.Date.Format(time.DateOnly))
		}
		if i > 0 && !p.Date.Equal(pts[i-1].Date.AddDate(0, 1, 0)) {
			return fmt.Errorf("%s follows %s: a month is missing", p.Date.Format(time.DateOnly), pts[i-1].Date.Format(time.DateOnly))
		}
	}
	return nil
}

// FlatRun returns the longest run of consecutive identical values among the
// points dated on or after from, and the date that run ends. A market price or
// rate that stops moving for longer than its history ever did is a degraded
// feed; from lets a caller skip the eras when a value was administered (a
// pegged rate, a posted price) and legitimately stood still.
func FlatRun(pts []marketdata.Point, from time.Time) (n int, end time.Time) {
	run := 0
	for i, p := range pts {
		if p.Date.Before(from) {
			continue
		}
		if run > 0 && p.Close == pts[i-1].Close {
			run++
		} else {
			run = 1
		}
		if run > n {
			n, end = run, p.Date
		}
	}
	return n, end
}

// Header is what a bundled reference file says about itself.
type Header struct {
	ID        string      // the file is <ID>.csv
	Name      string      // "# name:", a display name
	Source    string      // "# source:", where the data comes from and what it is
	Junctions []time.Time // "# junctions:", see marketdata.Series.Junctions
	Ends      time.Time   // "# ends:", when the series stops by design; zero if live
	EndsWhy   string      // the reason written after Ends
}

// Write stores pts in dir as <ID>.csv, in the simdata format the bundle reads
// (marketdata.ReadSimdataFS), with a "# generated:" line of today's date.
func Write(dir string, h Header, pts []marketdata.Point) error {
	if h.ID == "" || len(pts) == 0 {
		return fmt.Errorf("refgen: nothing to write for %q", h.ID)
	}
	var b strings.Builder
	b.WriteString("# pofo simdata v1\n")
	fmt.Fprintf(&b, "# id: %s\n", h.ID)
	fmt.Fprintf(&b, "# name: %s\n", h.Name)
	fmt.Fprintf(&b, "# source: %s\n", h.Source)
	fmt.Fprintf(&b, "# generated: %s\n", time.Now().UTC().Format(time.DateOnly))
	if len(h.Junctions) > 0 {
		days := make([]string, len(h.Junctions))
		for i, j := range h.Junctions {
			days[i] = j.Format(time.DateOnly)
		}
		fmt.Fprintf(&b, "# junctions: %s\n", strings.Join(days, ","))
	}
	b.WriteString(marketdata.EndsHeader(h.Ends, h.EndsWhy))
	b.WriteString("date,close\n")
	for _, p := range pts {
		fmt.Fprintf(&b, "%s,%.6f\n", p.Date.Format(time.DateOnly), p.Close)
	}
	return os.WriteFile(filepath.Join(dir, h.ID+".csv"), []byte(b.String()), 0o644)
}
