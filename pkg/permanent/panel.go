package permanent

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
)

// Panel is the parsed multi-country monthly macro panel: for each column
// (ip, cpi, shortrate, longrate, shareprice) a map from ISO country code to a
// month-keyed value series. Months are normalized to the first of the month at
// 00:00 UTC. Built from the embedded OECD MEI data (datasets.MacroPanel).
type Panel struct {
	// series[column][iso][month] = value. Columns: the five macro drivers.
	series map[string]map[string]map[time.Time]float64
	isos   []string // countries present, sorted
}

// panelColumns are the macro-driver columns, in the panel CSV's field order.
var panelColumns = []string{"ip", "cpi", "shortrate", "longrate", "shareprice"}

// LoadPanel parses the macro panel embedded in the binary.
func LoadPanel() (*Panel, error) { return ParsePanel(datasets.MacroPanel()) }

// ParsePanel parses a macro-panel CSV (iso,date,ip,cpi,shortrate,longrate,
// shareprice; date as YYYY-MM; comment lines start with '#'; empty cells mean
// the series does not cover that month).
func ParsePanel(csv []byte) (*Panel, error) {
	p := &Panel{series: map[string]map[string]map[time.Time]float64{}}
	for _, col := range panelColumns {
		p.series[col] = map[string]map[time.Time]float64{}
	}
	seen := map[string]bool{}
	for line := range strings.SplitSeq(string(csv), "\n") {
		if line == "" || line[0] == '#' || strings.HasPrefix(line, "iso,") {
			continue
		}
		f := strings.Split(line, ",")
		if len(f) != len(panelColumns)+2 {
			return nil, fmt.Errorf("macro panel: row has %d fields, want %d: %q", len(f), len(panelColumns)+2, line)
		}
		iso := f[0]
		t, err := time.Parse("2006-01", f[1])
		if err != nil {
			return nil, fmt.Errorf("macro panel: bad date %q: %w", f[1], err)
		}
		m := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
		for i, col := range panelColumns {
			cell := f[i+2]
			if cell == "" {
				continue
			}
			v, err := strconv.ParseFloat(cell, 64)
			if err != nil {
				return nil, fmt.Errorf("macro panel: bad %s %q: %w", col, cell, err)
			}
			if p.series[col][iso] == nil {
				p.series[col][iso] = map[time.Time]float64{}
			}
			p.series[col][iso][m] = v
		}
		if !seen[iso] {
			seen[iso] = true
			p.isos = append(p.isos, iso)
		}
	}
	sort.Strings(p.isos)
	if len(p.isos) == 0 {
		return nil, fmt.Errorf("macro panel: no data rows")
	}
	return p, nil
}

// Countries lists the ISO codes present in the panel, sorted.
func (p *Panel) Countries() []string { return append([]string(nil), p.isos...) }

// value returns column col for country iso at month m.
func (p *Panel) value(col, iso string, m time.Time) (float64, bool) {
	byISO := p.series[col]
	if byISO == nil {
		return 0, false
	}
	s := byISO[iso]
	if s == nil {
		return 0, false
	}
	v, ok := s[m]
	return v, ok
}

// yoy returns the year-on-year change of column col for iso at m (m vs m-12).
func (p *Panel) yoy(col, iso string, m time.Time) (float64, bool) {
	now, ok1 := p.value(col, iso, m)
	prev, ok2 := p.value(col, iso, m.AddDate(-1, 0, 0))
	if !ok1 || !ok2 || prev == 0 {
		return 0, false
	}
	return now/prev - 1, true
}
