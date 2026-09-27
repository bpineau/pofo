package refgen

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// OECDBase is the root of the OECD's own SDMX REST API, the publisher the
// DBnomics mirror copies from. The generators read it directly since 2026-09:
// the mirror's OECD provider was last indexed 2026-06-16, which left every
// OECD-fed file stopped around 2026-05 while this API already served 2026-08.
const OECDBase = "https://sdmx.oecd.org/public/rest"

// Flow names one OECD dataflow the way the SDMX API addresses it: the agency
// that maintains it and its id.
type Flow struct{ Agency, ID string }

// The OECD dataflows the generators read.
var (
	// FinMark is the short-term statistics' financial-markets dataflow:
	// interest rates (IR3TIB, IRSTCI, IRLT) and share prices (SHARE).
	FinMark = Flow{"OECD.SDD.STES", "DSD_STES@DF_FINMARK"}
	// IndServ is the short-term statistics' production dataflow.
	IndServ = Flow{"OECD.SDD.STES", "DSD_STES@DF_INDSERV"}
	// Prices is the consumer-price dataflow under the COICOP 1999
	// classification, which a country stops in once it has moved to 2018.
	Prices = Flow{"OECD.SDD.TPS", "DSD_PRICES@DF_PRICES_ALL"}
	// PricesC2018 is the consumer-price dataflow under COICOP 2018.
	PricesC2018 = Flow{"OECD.SDD.TPS", "DSD_PRICES_COICOP2018@DF_PRICES_C2018_ALL"}
)

// The OECD API admits 60 data downloads an hour per client and answers the
// next one with HTTP 429 until the window frees up. OECD therefore batches:
// each call is ONE download whatever the number of series it asks for, so a
// generator reads a whole dataflow's worth of keys at once, and a full refresh
// of every generator costs a handful of downloads. The variables below are the
// politeness around that budget; tests shorten them.
var (
	// oecdSpacing is the pause kept between two downloads from one process.
	oecdSpacing = 2 * time.Second
	// oecdBackoff is the first wait after a 429 or a transient failure when the
	// server names none (Retry-After); it doubles on every further attempt.
	oecdBackoff = time.Minute
	// oecdAttempts bounds the tries of one download: a minute, then 2, 4, 8
	// and 16, which outlasts a full hourly window.
	oecdAttempts = 6
)

// oecdGate serializes the downloads of one process and spaces them.
var oecdGate struct {
	sync.Mutex
	last time.Time
}

// oecdClient bounds one download; a whole dataflow's worth of series for thirty
// countries over seventy years is a few tens of megabytes of CSV.
var oecdClient = &http.Client{Timeout: 5 * time.Minute}

// OECD reads the series named by keys from one dataflow under base (OECDBase
// outside tests) and returns those it found, keyed exactly as asked, each in
// date order. A key is an SDMX series key with every dimension spelled out
// ("EA20.M.IRLT.PA._Z._Z._Z._Z.N").
//
// All the keys are fetched in ONE download: the request carries, per
// dimension, the union of the values the keys use ("EA20+DEU.M.IRLT+SHARE..."),
// and the rows of the combinations nobody asked for are dropped on reading. A
// key the dataflow does not carry is simply absent from the result, as several
// countries publish no monthly production or prices at all; which absences are
// acceptable is the caller's decision.
//
// Observations are dated as the API labels them: a monthly period ("2026-08")
// on the first of its month, a daily one on its day. A row without a value is
// skipped. A 429 or a transient server failure is retried after the server's
// Retry-After, or after a doubling backoff when it names none.
func OECD(base string, flow Flow, keys ...string) (map[string][]marketdata.Point, error) {
	union, err := unionKey(keys)
	if err != nil {
		return nil, fmt.Errorf("OECD %s: %w", flow.ID, err)
	}
	url := fmt.Sprintf("%s/data/%s,%s,/%s", base, flow.Agency, flow.ID, union)
	body, err := oecdGet(url)
	if err != nil {
		return nil, fmt.Errorf("OECD %s: %w", flow.ID, err)
	}
	want := make(map[string]bool, len(keys))
	for _, k := range keys {
		want[k] = true
	}
	out, err := parseOECD(body, want)
	if err != nil {
		return nil, fmt.Errorf("OECD %s: %w", flow.ID, err)
	}
	return out, nil
}

// unionKey folds series keys into the one SDMX key that selects them all:
// dimension by dimension, the distinct values in first-seen order, joined by
// '+'.
func unionKey(keys []string) (string, error) {
	if len(keys) == 0 {
		return "", fmt.Errorf("no series key")
	}
	var dims [][]string
	for _, k := range keys {
		parts := strings.Split(k, ".")
		if dims == nil {
			dims = make([][]string, len(parts))
		}
		if len(parts) != len(dims) {
			return "", fmt.Errorf("key %q has %d dimensions, %q has %d", k, len(parts), keys[0], len(dims))
		}
		for i, v := range parts {
			if v == "" || strings.Contains(v, "+") {
				return "", fmt.Errorf("key %q: dimension %d must name exactly one value", k, i+1)
			}
			if !contains(dims[i], v) {
				dims[i] = append(dims[i], v)
			}
		}
	}
	parts := make([]string, len(dims))
	for i, vs := range dims {
		parts[i] = strings.Join(vs, "+")
	}
	return strings.Join(parts, "."), nil
}

func contains(vs []string, v string) bool {
	for _, x := range vs {
		if x == v {
			return true
		}
	}
	return false
}

// oecdGet downloads url as SDMX-CSV, spaced from the previous download and
// retried politely. A 404 is the API's "no series matches" and returns an empty
// body rather than an error.
func oecdGet(url string) ([]byte, error) {
	oecdGate.Lock()
	defer oecdGate.Unlock()
	wait := oecdBackoff
	var lastErr error
	for attempt := 1; attempt <= oecdAttempts; attempt++ {
		if pause := oecdSpacing - time.Since(oecdGate.last); pause > 0 {
			time.Sleep(pause)
		}
		body, retryAfter, err := oecdOnce(url)
		oecdGate.last = time.Now()
		if err == nil {
			return body, nil
		}
		lastErr = err
		if retryAfter < 0 || attempt == oecdAttempts {
			break // not worth retrying, or out of attempts
		}
		if retryAfter == 0 {
			retryAfter = wait
			wait *= 2
		}
		log.Printf("OECD: %v; retrying in %s (attempt %d of %d)", err, retryAfter, attempt+1, oecdAttempts)
		time.Sleep(retryAfter)
	}
	return nil, lastErr
}

// oecdOnce is one download. Its duration says what to do on failure: negative
// means do not retry, zero means retry after the backoff, positive is the
// server's own Retry-After.
func oecdOnce(url string) ([]byte, time.Duration, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, -1, err
	}
	req.Header.Set("Accept", "application/vnd.sdmx.data+csv; charset=utf-8")
	req.Header.Set("User-Agent", fredAgent)
	resp, err := oecdClient.Do(req)
	if err != nil {
		return nil, 0, err // a network failure is transient until proven otherwise
	}
	defer resp.Body.Close()
	switch {
	case resp.StatusCode == http.StatusOK:
		body, err := io.ReadAll(io.LimitReader(resp.Body, 256<<20))
		if err != nil {
			return nil, 0, err
		}
		return body, 0, nil
	case resp.StatusCode == http.StatusNotFound:
		return nil, 0, nil
	case resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500:
		return nil, retryAfter(resp.Header.Get("Retry-After")), fmt.Errorf("HTTP %d", resp.StatusCode)
	default:
		return nil, -1, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
}

// retryAfter reads a Retry-After header given in seconds; anything else (an
// HTTP date, nothing) leaves the choice to the backoff.
func retryAfter(h string) time.Duration {
	if s, err := strconv.Atoi(strings.TrimSpace(h)); err == nil && s > 0 {
		return time.Duration(s) * time.Second
	}
	return 0
}

// parseOECD reads an SDMX-CSV body: a DATAFLOW column, the dimensions in the
// order of the series key, TIME_PERIOD, OBS_VALUE, then attributes. Only the
// series in want are kept. A leading byte-order mark, which the API may send,
// is dropped.
func parseOECD(body []byte, want map[string]bool) (map[string][]marketdata.Point, error) {
	out := map[string][]marketdata.Point{}
	body = bytes.TrimPrefix(body, byteOrderMark)
	if len(body) == 0 {
		return out, nil
	}
	r := csv.NewReader(bytes.NewReader(body))
	r.FieldsPerRecord = -1
	head, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("header: %w", err)
	}
	period, value := -1, -1
	for i, h := range head {
		switch h {
		case "TIME_PERIOD":
			period = i
		case "OBS_VALUE":
			value = i
		}
	}
	if period < 2 || value < 0 || head[0] != "DATAFLOW" {
		return nil, fmt.Errorf("not an SDMX-CSV series table: header %q", head)
	}
	for line := 2; ; line++ {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", line, err)
		}
		if len(row) <= max(period, value) {
			return nil, fmt.Errorf("line %d: %d fields", line, len(row))
		}
		key := strings.Join(row[1:period], ".")
		if !want[key] {
			continue
		}
		v, err := strconv.ParseFloat(row[value], 64)
		if err != nil || math.IsNaN(v) {
			continue // no value for that period
		}
		t, err := parseSDMXPeriod(row[period])
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", line, err)
		}
		out[key] = append(out[key], marketdata.Point{Date: t, Close: v})
	}
	for key, s := range out {
		sort.Slice(s, func(i, j int) bool { return s[i].Date.Before(s[j].Date) })
		for i := 1; i < len(s); i++ {
			if s[i].Date.Equal(s[i-1].Date) {
				return nil, fmt.Errorf("%s: %s appears twice", key, s[i].Date.Format(time.DateOnly))
			}
		}
	}
	return out, nil
}

// byteOrderMark is the UTF-8 encoding of U+FEFF.
var byteOrderMark = []byte{0xEF, 0xBB, 0xBF}

// parseSDMXPeriod reads a monthly ("2026-08", dated the first of the month) or
// a daily ("2026-08-14") SDMX period.
func parseSDMXPeriod(p string) (time.Time, error) {
	switch len(p) {
	case len("2006-01"):
		return time.Parse("2006-01", p)
	case len(time.DateOnly):
		return time.Parse(time.DateOnly, p)
	}
	return time.Time{}, fmt.Errorf("period %q is neither monthly nor daily", p)
}
