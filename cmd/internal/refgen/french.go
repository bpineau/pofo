package refgen

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// FrenchCSV returns the one CSV a Ken French Data Library zip holds.
func FrenchCSV(zipped []byte) (string, error) {
	zr, err := zip.NewReader(bytes.NewReader(zipped), int64(len(zipped)))
	if err != nil {
		return "", err
	}
	if len(zr.File) != 1 {
		return "", fmt.Errorf("zip holds %d files, want 1", len(zr.File))
	}
	f, err := zr.File[0].Open()
	if err != nil {
		return "", err
	}
	defer f.Close()
	body, err := io.ReadAll(io.LimitReader(f, 64<<20))
	return string(body), err
}

// FrenchTable reads the FIRST daily table of a Ken French Data Library CSV:
// the first header line naming every requested column, then one row per
// trading day keyed YYYYMMDD, up to the blank line that closes the table. The
// files that carry several tables (the portfolio sorts: value-weighted returns
// first, then equal-weighted ones, firm counts and sizes under the SAME column
// names) are therefore read for their value-weighted returns. Columns are found
// by NAME, since the library has added columns before, and the values come
// back as the library prints them, in PERCENT, one slice per requested column
// in the order asked.
//
// The library's missing-value sentinels (-99.99, -999) are refused rather than
// read as a crash, and so is a file where no header names every column.
func FrenchTable(body string, columns ...string) (dates []time.Time, values [][]float64, err error) {
	var at []int
	values = make([][]float64, len(columns))
	for n, line := range strings.Split(body, "\n") {
		fields := strings.Split(strings.TrimRight(line, "\r"), ",")
		if at == nil {
			at = columnsAt(fields, columns)
			continue
		}
		key := strings.TrimSpace(fields[0])
		t, err := time.Parse("20060102", key)
		if len(key) != 8 || err != nil {
			if len(dates) > 0 {
				break // the table is over
			}
			continue
		}
		for i, c := range at {
			if c >= len(fields) {
				return nil, nil, fmt.Errorf("line %d: no %s column", n+1, columns[i])
			}
			v, err := strconv.ParseFloat(strings.TrimSpace(fields[c]), 64)
			if err != nil {
				return nil, nil, fmt.Errorf("line %d: bad %s value %q", n+1, columns[i], fields[c])
			}
			if v <= -99 {
				return nil, nil, fmt.Errorf("%s: missing %s value in the source", t.Format(time.DateOnly), columns[i])
			}
			values[i] = append(values[i], v)
		}
		dates = append(dates, t)
	}
	if at == nil {
		return nil, nil, fmt.Errorf("no header naming %s", strings.Join(columns, ", "))
	}
	if len(dates) == 0 {
		return nil, nil, fmt.Errorf("no daily row under the %s header", strings.Join(columns, ", "))
	}
	return dates, values, nil
}

// columnsAt returns the position of each wanted column in a header line, or
// nil when the line does not name them all.
func columnsAt(fields, wanted []string) []int {
	at := make([]int, len(wanted))
	for i, w := range wanted {
		at[i] = -1
		for j, f := range fields {
			if strings.TrimSpace(f) == w {
				at[i] = j
				break
			}
		}
		if at[i] < 0 {
			return nil
		}
	}
	return at
}
