package marketdata

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"
)

// This file is the plain-text face of a Series. Two layouts share one line
// grammar:
//
//   - the SINGLE layout, one series per file, "date,value" rows: every file
//     bundled under pkg/datasets (simdata and refdata) is written this way,
//     and ReadCSV and ReadSimdataFS read it;
//   - the LONG layout, several series in one file, "id,date,value" rows:
//     WriteCSV writes it and ReadLongCSV reads it back.
//
// In both, blank lines are skipped and a line starting with "#" is a
// comment. A comment of the form "<key>: <value>" (single) or
// "<id> <key>: <value>" (long) is a METADATA header when the key is one of
// seriesKeys, and ignored otherwise; any other comment is ignored too. The
// first non-comment line may be a column header (see columnHeader).

// The metadata keys a series carries through a CSV file.
const (
	keyName            = "name"             // Series.Name
	keyCurrency        = "currency"         // Series.Currency
	keyJunctions       = "junctions"        // Series.Junctions, comma-separated ISO dates
	keySimulatedBefore = "simulated-before" // Series.SimulatedBefore, one ISO date
)

// seriesKeys lists the metadata keys in the order WriteCSV writes them.
var seriesKeys = []string{keyName, keyCurrency, keySimulatedBefore, keyJunctions}

// ReadCSV reads one series from a "date,value" CSV: the layout of every file
// bundled under pkg/datasets, of a file written by hand, or of an old version
// of a bundled file piped out of "git show". id becomes Series.Symbol; when it
// is empty, the file's "# id:" header supplies it.
//
// Blank lines and "#" comments are skipped, except for the metadata headers
// "# name:", "# currency:", "# simulated-before:" (one ISO date) and
// "# junctions:" (comma-separated ISO dates), which set the Series field of
// the same name. Junctions are the one header a consumer must honour to be
// correct rather than merely informed: see Series.Junctions. The first other
// line may be a column header ("date,close", "date,value": a date field that
// is not a date and a value that is not a number); every other line is an ISO
// date (2006-01-02)
// and a finite value, zero and negative included, since a rate or a spread is
// as welcome as a price. Rows may come in any order and are returned sorted;
// a date given twice is an error, as is a file without a single row.
//
// Errors name the offending line.
func ReadCSV(r io.Reader, id string) (*Series, error) {
	s, err := readSingle(r, id, finite)
	if err != nil {
		if id != "" {
			return nil, fmt.Errorf("marketdata: ReadCSV %s: %w", id, err)
		}
		return nil, fmt.Errorf("marketdata: ReadCSV: %w", err)
	}
	return s, nil
}

// ReadLongCSV reads several series from one "id,date,value" CSV, the layout
// WriteCSV writes, and returns them in the order their identifiers first
// appear, each sorted by date. Blank lines and "#" comments are skipped,
// except for the per-series metadata headers WriteCSV writes,
// "# <id> <key>: <value>" with the keys ReadCSV knows (name, currency,
// simulated-before, junctions). The first other line may be the column
// header, recognized as ReadCSV recognizes one.
//
// A date given twice for one identifier is an error, as is metadata for an
// identifier that has no row. Errors name the offending line.
func ReadLongCSV(r io.Reader) ([]*Series, error) {
	var order []*seriesBuilder
	byID := map[string]*seriesBuilder{}
	get := func(id string) *seriesBuilder {
		b, ok := byID[id]
		if !ok {
			b = &seriesBuilder{s: &Series{Symbol: id}}
			byID[id] = b
			order = append(order, b)
		}
		return b
	}
	header := true
	err := eachLine(r, func(n int, line string) error {
		if text, ok := strings.CutPrefix(line, "#"); ok {
			if id, key, val, ok := longMetadata(strings.TrimSpace(text)); ok {
				return get(id).metadata(n, key, val)
			}
			return nil
		}
		fields := strings.Split(line, ",")
		if len(fields) != 3 {
			return fmt.Errorf("line %d: %d fields, want id,date,value", n, len(fields))
		}
		if header {
			header = false
			if columnHeader(fields[1], fields[2]) {
				return nil
			}
		}
		id := strings.TrimSpace(fields[0])
		if id == "" {
			return fmt.Errorf("line %d: empty id", n)
		}
		return get(id).row(n, fields[1], fields[2], finite)
	})
	if err != nil {
		return nil, fmt.Errorf("marketdata: ReadLongCSV: %w", err)
	}
	out := make([]*Series, len(order))
	for i, b := range order {
		s, err := b.series()
		if err != nil {
			return nil, fmt.Errorf("marketdata: ReadLongCSV %s: %w", b.s.Symbol, err)
		}
		out[i] = s
	}
	return out, nil
}

// WriteCSV writes the series in the long "id,date,value" layout, one after the
// other in the order given: first each series' metadata as comments
// ("# <id> name: …", then currency, simulated-before and junctions, each only
// when set), then the column header, then one row per point. It is what
// "pofo -dump" prints, and the form to hand a series to another tool: any CSV
// reader that skips "#" lines (pandas' comment="#", R's comment.char) reads
// the rows.
//
// Values are written in the shortest form that parses back to the same
// float64 (strconv 'g', -1), so ReadLongCSV returns exactly the symbols,
// names and currencies (trimmed of surrounding blanks), simulation frontiers,
// junctions and points it was given. Nothing else a Series carries is
// written: not its source, its dividends nor a nowcast frontier (strip a
// nowcast tail with WithoutEstimates first, as every storing consumer does).
//
// Every series must have a symbol, unique in the list, free of commas, double
// quotes, line breaks, surrounding blanks and a leading "#", a name and a
// currency free of line breaks, and at least one point, with strictly
// ascending dates and finite values. The whole list is checked before
// anything is written, so a refused list writes nothing.
func WriteCSV(w io.Writer, list ...*Series) error {
	seen := map[string]bool{}
	for _, s := range list {
		if err := writable(s); err != nil {
			return fmt.Errorf("marketdata: WriteCSV: %w", err)
		}
		if seen[s.Symbol] {
			return fmt.Errorf("marketdata: WriteCSV: %s given twice", s.Symbol)
		}
		seen[s.Symbol] = true
	}
	bw := bufio.NewWriter(w)
	for _, s := range list {
		for _, key := range seriesKeys {
			if val := metadataValue(s, key); val != "" {
				fmt.Fprintf(bw, "# %s %s: %s\n", s.Symbol, key, val)
			}
		}
	}
	bw.WriteString("id,date,value\n")
	for _, s := range list {
		for _, p := range s.Points {
			bw.WriteString(s.Symbol)
			bw.WriteByte(',')
			bw.WriteString(p.Date.Format(time.DateOnly))
			bw.WriteByte(',')
			bw.WriteString(strconv.FormatFloat(p.Close, 'g', -1, 64))
			bw.WriteByte('\n')
		}
	}
	return bw.Flush()
}

// writable reports why WriteCSV cannot write s so that it reads back whole.
func writable(s *Series) error {
	switch {
	case s == nil:
		return errors.New("nil series")
	case s.Symbol == "":
		return errors.New("a series has no symbol")
	case strings.ContainsAny(s.Symbol, ",\"\r\n"):
		return fmt.Errorf("symbol %q holds a comma, a quote or a line break", s.Symbol)
	case s.Symbol != strings.TrimSpace(s.Symbol) || strings.HasPrefix(s.Symbol, "#"):
		return fmt.Errorf("symbol %q would not read back: surrounding blanks or a leading #", s.Symbol)
	case strings.ContainsAny(s.Name+s.Currency, "\r\n"):
		return fmt.Errorf("%s: the name or the currency holds a line break", s.Symbol)
	case len(s.Points) == 0:
		return fmt.Errorf("%s has no point", s.Symbol)
	}
	for i, p := range s.Points {
		if !finite(p.Close) {
			return fmt.Errorf("%s: the value on %s is not finite", s.Symbol, p.Date.Format(time.DateOnly))
		}
		if i > 0 && !p.Date.After(s.Points[i-1].Date) {
			return fmt.Errorf("%s: %s does not follow %s", s.Symbol,
				p.Date.Format(time.DateOnly), s.Points[i-1].Date.Format(time.DateOnly))
		}
	}
	return nil
}

// metadataValue is the header value WriteCSV writes for key, empty when the
// field is unset.
func metadataValue(s *Series, key string) string {
	switch key {
	case keyName:
		return strings.TrimSpace(s.Name)
	case keyCurrency:
		return strings.TrimSpace(s.Currency)
	case keySimulatedBefore:
		if !s.SimulatedBefore.IsZero() {
			return s.SimulatedBefore.Format(time.DateOnly)
		}
	case keyJunctions:
		days := make([]string, len(s.Junctions))
		for i, j := range s.Junctions {
			days[i] = j.Format(time.DateOnly)
		}
		return strings.Join(days, ",")
	}
	return ""
}

// readSingle is the one parser of the single layout, shared by ReadCSV and
// ReadSimdataFS; valid is the rule every value must pass (finite for an
// arbitrary file, positive for a bundled price file).
func readSingle(r io.Reader, id string, valid func(float64) bool) (*Series, error) {
	b := &seriesBuilder{s: &Series{Symbol: id}}
	header := true
	err := eachLine(r, func(n int, line string) error {
		if text, ok := strings.CutPrefix(line, "#"); ok {
			key, val, found := strings.Cut(strings.TrimSpace(text), ":")
			if !found {
				return nil
			}
			key, val = strings.TrimSpace(key), strings.TrimSpace(val)
			if key == "id" && id == "" {
				b.s.Symbol = val
				return nil
			}
			return b.metadata(n, key, val)
		}
		date, value, found := strings.Cut(line, ",")
		if header {
			header = false
			if columnHeader(date, value) {
				return nil
			}
		}
		if !found || strings.Contains(value, ",") {
			return fmt.Errorf("line %d: %q is not a date,value row", n, line)
		}
		return b.row(n, date, value, valid)
	})
	if err != nil {
		return nil, err
	}
	return b.series()
}

// seriesBuilder accumulates one series as its lines arrive, remembering the
// line each point came from so a duplicate can be named.
type seriesBuilder struct {
	s    *Series
	rows []csvRow
}

type csvRow struct {
	p    Point
	line int
}

// metadata applies one header; an unknown key is ignored, as any comment.
func (b *seriesBuilder) metadata(n int, key, val string) error {
	switch key {
	case keyName:
		b.s.Name = val
	case keyCurrency:
		b.s.Currency = val
	case keySimulatedBefore:
		t, err := parseDate(val)
		if err != nil {
			return fmt.Errorf("line %d: invalid %s date %q", n, key, val)
		}
		b.s.SimulatedBefore = t
	case keyJunctions:
		for d := range strings.SplitSeq(val, ",") {
			t, err := parseDate(d)
			if err != nil {
				return fmt.Errorf("line %d: invalid junction date %q", n, strings.TrimSpace(d))
			}
			b.s.Junctions = append(b.s.Junctions, t)
		}
	}
	return nil
}

// row adds one data row.
func (b *seriesBuilder) row(n int, date, value string, valid func(float64) bool) error {
	t, err := parseDate(date)
	if err != nil {
		return fmt.Errorf("line %d: invalid date %q", n, strings.TrimSpace(date))
	}
	v, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil || !valid(v) {
		return fmt.Errorf("line %d: invalid value %q", n, strings.TrimSpace(value))
	}
	b.rows = append(b.rows, csvRow{Point{Date: t, Close: v}, n})
	return nil
}

// series sorts the rows by date and returns the finished series, refusing an
// empty one and a date given twice.
func (b *seriesBuilder) series() (*Series, error) {
	if len(b.rows) == 0 {
		return nil, errors.New("no data")
	}
	slices.SortStableFunc(b.rows, func(x, y csvRow) int { return x.p.Date.Compare(y.p.Date) })
	b.s.Points = make([]Point, len(b.rows))
	for i, r := range b.rows {
		if i > 0 && r.p.Date.Equal(b.rows[i-1].p.Date) {
			return nil, fmt.Errorf("lines %d and %d: %s given twice",
				b.rows[i-1].line, r.line, r.p.Date.Format(time.DateOnly))
		}
		b.s.Points[i] = r.p
	}
	return b.s, nil
}

// longMetadata splits a long-layout comment "<id> <key>: <value>" on the
// leftmost known key, so an identifier may hold spaces and a name colons.
func longMetadata(text string) (id, key, val string, ok bool) {
	at := -1
	for _, k := range seriesKeys {
		if i := strings.Index(text, " "+k+":"); i > 0 && (at < 0 || i < at) {
			at, key = i, k
		}
	}
	if at < 0 {
		return "", "", "", false
	}
	return strings.TrimSpace(text[:at]), key, strings.TrimSpace(text[at+len(key)+2:]), true
}

// eachLine calls fn with every non-blank line of r, trimmed, and its 1-based
// number, stopping at the first error.
func eachLine(r io.Reader, fn func(n int, line string) error) error {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64<<10), 1<<20) // a header comment may run long
	for n := 1; sc.Scan(); n++ {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		if err := fn(n, line); err != nil {
			return err
		}
	}
	return sc.Err()
}

func parseDate(s string) (time.Time, error) {
	return time.ParseInLocation(time.DateOnly, strings.TrimSpace(s), time.UTC)
}

// columnHeader reports whether the first non-comment line is a column header
// rather than a data row: its date field is not a date AND its value field is
// not a number, so a first row with a typo in its date is still reported as
// the error it is instead of being skipped as a header.
func columnHeader(date, value string) bool {
	_, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	return !isDate(date) && err != nil
}

func isDate(s string) bool { _, err := parseDate(s); return err == nil }

func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }

func positive(v float64) bool { return finite(v) && v > 0 }
