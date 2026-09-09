package datasets

import (
	"encoding/json"
	"io/fs"
	"math"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestEmbeddedDatasetsPresent(t *testing.T) {
	sim, err := fs.ReadDir(Simdata(), ".")
	if err != nil || len(sim) < 8 {
		t.Fatalf("embedded simdata incomplete: %d files, %v", len(sim), err)
	}
}

func TestMacroPanelWellFormed(t *testing.T) {
	lines := strings.Split(strings.TrimSpace(string(MacroPanel())), "\n")
	var rows, header int
	isos := map[string]bool{}
	for _, l := range lines {
		if strings.HasPrefix(l, "#") {
			continue
		}
		if strings.HasPrefix(l, "iso,") {
			header++
			continue
		}
		f := strings.Split(l, ",")
		if len(f) != 7 {
			t.Fatalf("row has %d fields, want 7: %q", len(f), l)
		}
		isos[f[0]] = true
		rows++
	}
	if header != 1 {
		t.Fatalf("expected exactly one header row, got %d", header)
	}
	if rows < 10000 {
		t.Fatalf("macro panel looks truncated: %d rows", rows)
	}
	if len(isos) < 20 {
		t.Fatalf("macro panel covers only %d countries", len(isos))
	}
}

func TestAssetMetaIsValidJSON(t *testing.T) {
	var raw []map[string]any
	if err := json.Unmarshal(AssetMeta(), &raw); err != nil {
		t.Fatalf("AssetMeta is not a JSON array: %v", err)
	}
	if len(raw) < 100 {
		t.Fatalf("AssetMeta looks truncated: %d entries", len(raw))
	}
}

func TestCatalogParsesAndIsTyped(t *testing.T) {
	assets := Catalog()
	if len(assets) < 100 {
		t.Fatalf("Catalog looks truncated: %d assets", len(assets))
	}
	// Spot-check a well-known entry decodes into the full typed record.
	var iwda Asset
	for _, a := range assets {
		if a.ID == "IE00B4L5Y983" {
			iwda = a
			break
		}
	}
	if iwda.Name == "" || !iwda.UCITS || iwda.Fees == 0 || iwda.Geography["US"] == 0 || iwda.AssetClass != "equity" {
		t.Fatalf("IWDA did not decode into a full Asset: %+v", iwda)
	}
}

// TestCatalogEURetailConsistent enforces the invariants of the eu_retail
// flag: every UCITS fund is EU-retail buyable by definition, a US-ISIN
// instrument without a PRIIPs KID never is, and fee-free index series
// (source "index") are not tradable so the flag must be absent (false).
func TestCatalogEURetailConsistent(t *testing.T) {
	for _, a := range Catalog() {
		if a.UCITS && !a.EURetail {
			t.Errorf("%s: ucits implies eu_retail", a.ID)
		}
		if strings.HasPrefix(a.ISIN, "US") && a.EURetail {
			t.Errorf("%s: US-listed instrument flagged eu_retail", a.ID)
		}
		if a.Source == "index" && a.EURetail {
			t.Errorf("%s: index series cannot be eu_retail", a.ID)
		}
	}
}

// seriesRow is one parsed data line of a bundled price/level CSV.
type seriesRow struct {
	date  time.Time
	close float64
}

// parseSeries parses a bundled simdata/refdata CSV: "# " comment lines, an
// optional "date,close" header, then "YYYY-MM-DD,value" rows. It returns the
// rows and the value of the "# id:" stamp. Every parse error fails the test,
// naming the file: the bundled files are generated, so a malformed one is a
// generator bug, not a test fixture problem.
func parseSeries(t *testing.T, name string, body []byte) ([]seriesRow, string) {
	t.Helper()
	var rows []seriesRow
	var id string
	for n, line := range strings.Split(string(body), "\n") {
		line = strings.TrimSpace(line)
		switch {
		case line == "" || line == "date,close":
			continue
		case strings.HasPrefix(line, "#"):
			if key, val, found := strings.Cut(strings.TrimPrefix(line, "#"), ":"); found && strings.TrimSpace(key) == "id" {
				id = strings.TrimSpace(val)
			}
			continue
		}
		dateStr, closeStr, found := strings.Cut(line, ",")
		if !found {
			t.Fatalf("%s:%d: not a CSV row: %q", name, n+1, line)
		}
		d, err := time.ParseInLocation("2006-01-02", dateStr, time.UTC)
		if err != nil {
			t.Fatalf("%s:%d: invalid date %q", name, n+1, dateStr)
		}
		v, err := strconv.ParseFloat(closeStr, 64)
		if err != nil {
			t.Fatalf("%s:%d: invalid value %q", name, n+1, closeStr)
		}
		rows = append(rows, seriesRow{d, v})
	}
	return rows, id
}

// checkSeriesFS asserts the invariants every bundled series file must hold:
// a ".csv" name, an "# id:" stamp equal to that name (the guard against a
// generation writing over another asset's file), at least two points, dates
// normalized to 00:00 UTC and strictly increasing, and finite positive values
// (marketdata.ReadSimdataFS rejects anything else).
func checkSeriesFS(t *testing.T, fsys fs.FS, kind string, minFiles int) {
	t.Helper()
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		t.Fatalf("read %s: %v", kind, err)
	}
	if len(entries) < minFiles {
		t.Fatalf("embedded %s has %d files, want at least %d", kind, len(entries), minFiles)
	}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".csv") {
			t.Errorf("%s/%s: not a CSV file", kind, name)
			continue
		}
		body, err := fs.ReadFile(fsys, name)
		if err != nil {
			t.Errorf("%s/%s: %v", kind, name, err)
			continue
		}
		rows, id := parseSeries(t, kind+"/"+name, body)
		if base := strings.TrimSuffix(name, ".csv"); id != base {
			t.Errorf("%s/%s: id stamp %q does not match the file name", kind, name, id)
		}
		if len(rows) < 2 {
			t.Errorf("%s/%s: %d points, want at least 2", kind, name, len(rows))
			continue
		}
		for i, r := range rows {
			if r.date.Location() != time.UTC || !r.date.Equal(r.date.Truncate(24*time.Hour)) {
				t.Errorf("%s/%s: date %v is not normalized to 00:00 UTC", kind, name, r.date)
			}
			if math.IsNaN(r.close) || math.IsInf(r.close, 0) || r.close <= 0 {
				t.Errorf("%s/%s: value %v at %s is not a positive finite number", kind, name, r.close, r.date.Format("2006-01-02"))
			}
			if i > 0 && !rows[i-1].date.Before(r.date) {
				t.Errorf("%s/%s: dates not strictly increasing at %s", kind, name, r.date.Format("2006-01-02"))
			}
		}
	}
}

func TestSimdataFilesWellFormed(t *testing.T) { checkSeriesFS(t, Simdata(), "simdata", 40) }

func TestRefdataFilesWellFormed(t *testing.T) { checkSeriesFS(t, Refdata(), "refdata", 20) }

// TestBroadSampleWellFormed checks the JST panel's shape and plausibility:
// "iso,year,equity,bond,bill" rows of REAL annual fractions, years strictly
// increasing within a country. The lower bound is -1 (a total loss: German
// bonds in 1923 print exactly that), the upper 3 (+300 % in one real year).
func TestBroadSampleWellFormed(t *testing.T) {
	years := map[string]int{}
	rows := 0
	for n, line := range strings.Split(strings.TrimSpace(string(BroadSample())), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		f := strings.Split(line, ",")
		if len(f) != 5 {
			t.Fatalf("line %d has %d fields, want 5: %q", n+1, len(f), line)
		}
		if rows == 0 && f[0] == "iso" {
			if got, want := line, "iso,year,equity,bond,bill"; got != want {
				t.Fatalf("header is %q, want %q", got, want)
			}
			continue
		}
		iso := f[0]
		if len(iso) != 3 || iso != strings.ToUpper(iso) {
			t.Fatalf("line %d: %q is not an ISO-3 country code", n+1, iso)
		}
		year, err := strconv.Atoi(f[1])
		if err != nil || year < 1850 || year > 2100 {
			t.Fatalf("line %d: implausible year %q", n+1, f[1])
		}
		if prev, seen := years[iso]; seen && year <= prev {
			t.Fatalf("line %d: %s years not increasing (%d after %d)", n+1, iso, year, prev)
		}
		years[iso] = year
		for i, col := range []string{"equity", "bond", "bill"} {
			cell := f[2+i]
			if cell == "" { // a gap in the source, kept as such
				continue
			}
			v, err := strconv.ParseFloat(cell, 64)
			if err != nil || math.IsNaN(v) || v < -1 || v > 3 {
				t.Fatalf("line %d: %s %s real return %q out of range", n+1, iso, col, cell)
			}
		}
		rows++
	}
	if len(years) < 16 {
		t.Fatalf("broad sample covers only %d countries", len(years))
	}
	if rows < 2000 {
		t.Fatalf("broad sample looks truncated: %d rows", rows)
	}
}

// TestCAPEWellFormed checks the Shiller series: monthly first-of-month dates,
// strictly increasing, with a PE10 inside the range history has ever printed
// (4.78 in 1920, 44.2 in 1999). The freshness floor catches a series that
// silently stopped being regenerated.
func TestCAPEWellFormed(t *testing.T) {
	rows, _ := parseSeries(t, "cape/shiller-cape.csv", replaceHeader(CAPE(), "date,cape"))
	if len(rows) < 1500 {
		t.Fatalf("CAPE looks truncated: %d months", len(rows))
	}
	var prev time.Time
	for _, r := range rows {
		if r.date.Day() != 1 {
			t.Fatalf("CAPE date %s is not a first of month", r.date.Format("2006-01-02"))
		}
		if !prev.IsZero() && !prev.Before(r.date) {
			t.Fatalf("CAPE dates not strictly increasing at %s", r.date.Format("2006-01-02"))
		}
		if r.close < 3 || r.close > 60 {
			t.Fatalf("implausible CAPE %v at %s", r.close, r.date.Format("2006-01-02"))
		}
		prev = r.date
	}
	if first := rows[0].date; first.Year() != 1881 {
		t.Fatalf("CAPE starts in %d, want 1881", first.Year())
	}
	if last := rows[len(rows)-1].date; last.Before(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("CAPE stops at %s: regenerate with \"make cape\"", last.Format("2006-01"))
	}
}

// replaceHeader rewrites a one-off CSV header into the "date,close" one
// parseSeries skips, so a two-column series with its own column name reuses
// the same parser.
func replaceHeader(body []byte, header string) []byte {
	return []byte(strings.Replace(string(body), header, "date,close", 1))
}

// TestMacroPanelValues checks the panel's cells, where TestMacroPanelWellFormed
// checks its shape: monthly "YYYY-MM" dates increasing within a country, index
// levels positive, and rates inside the band policy has ever visited.
func TestMacroPanelValues(t *testing.T) {
	months := map[string]string{}
	zeroed := map[string]int{}
	for n, line := range strings.Split(strings.TrimSpace(string(MacroPanel())), "\n") {
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "iso,") {
			continue
		}
		f := strings.Split(line, ",")
		iso, month := f[0], f[1]
		if _, err := time.Parse("2006-01", month); err != nil {
			t.Fatalf("line %d: %q is not a YYYY-MM month", n+1, month)
		}
		if prev, seen := months[iso]; seen && month <= prev {
			t.Fatalf("line %d: %s months not increasing (%s after %s)", n+1, iso, month, prev)
		}
		months[iso] = month
		for i, col := range []string{"ip", "cpi", "shortrate", "longrate", "shareprice"} {
			cell := f[2+i]
			if cell == "" { // the OECD publishes each series over its own span
				continue
			}
			v, err := strconv.ParseFloat(cell, 64)
			if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
				t.Fatalf("line %d: %s %s = %q is not a number", n+1, iso, col, cell)
			}
			isRate := strings.HasSuffix(col, "rate")
			// A crisis rate reaches three digits (Turkey printed 400 % in
			// February 2001) and a policy rate can be negative, so the band
			// is wide: it catches a unit slip (a fraction read as a percent),
			// not an unusual month.
			if isRate && (v < -10 || v > 500) {
				t.Fatalf("line %d: %s %s = %v is out of band", n+1, iso, col, v)
			}
			if !isRate && v <= 0 {
				// An index level cannot be zero: here it is a rebasing
				// artifact. Rebased on 2015 = 100, the hyperinflating CPI
				// and share prices of Brazil and Turkey round to 0.0000 in
				// the early decades. Consumers must not divide by such a
				// level (permanent.Panel.yoy drops both ends), so the
				// stretch is pinned rather than accepted anywhere.
				if col == "ip" || (iso != "BRA" && iso != "TUR") || month >= "1993" {
					t.Fatalf("line %d: %s %s = %v is not a positive index level", n+1, iso, col, v)
				}
				zeroed[iso+" "+col]++
			}
		}
	}
	if len(months) < 20 {
		t.Fatalf("macro panel covers only %d countries", len(months))
	}
	if len(zeroed) != 3 { // BRA cpi, BRA shareprice, TUR cpi
		t.Errorf("rounded-to-zero index levels changed: %v", zeroed)
	}
}
