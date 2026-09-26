package compare

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/suggest"
)

// The macro-regime reading behind the report's regime strip and per-regime
// matrix. It is DESCRIPTIVE: it says which growth x inflation season the world
// was in during a past month, dated by the month the numbers refer to (not by
// when they were published), and never drives an allocation.
//
// The world state is read as BREADTH over the embedded OECD macro panel
// (datasets.MacroPanel): the share of countries whose industrial production,
// respectively consumer prices, is accelerating (year-on-year rate above its
// value three months earlier), averaged over a trailing three months so the
// strip moves in waves rather than month to month. Each breadth is then
// thresholded at one half into the four seasons of suggest's vocabulary.

const (
	regimeAccelMonths  = 3 // "accelerating": yoy now above yoy this many months ago
	regimeSmoothMonths = 3 // trailing average of the raw breadths
	regimeMinCountries = 8 // fewer reporting countries: no reading that month
)

// macroPanel holds the two panel columns the regime reads, keyed by country
// then month (first of the month, 00:00 UTC).
type macroPanel struct {
	ip, cpi map[string]map[time.Time]float64
}

// parseMacroPanel parses the macro-panel CSV (iso,date,ip,cpi,shortrate,
// longrate,shareprice; date as YYYY-MM; '#' comments; an empty cell means the
// series does not cover that month). Only ip and cpi are kept.
func parseMacroPanel(csv []byte) (*macroPanel, error) {
	p := &macroPanel{ip: map[string]map[time.Time]float64{}, cpi: map[string]map[time.Time]float64{}}
	for line := range strings.SplitSeq(string(csv), "\n") {
		if line == "" || line[0] == '#' || strings.HasPrefix(line, "iso,") {
			continue
		}
		f := strings.Split(line, ",")
		if len(f) != 7 {
			return nil, fmt.Errorf("macro panel: row has %d fields, want 7: %q", len(f), line)
		}
		t, err := time.Parse("2006-01", f[1])
		if err != nil {
			return nil, fmt.Errorf("macro panel: bad date %q: %w", f[1], err)
		}
		for _, c := range []struct {
			cell string
			into map[string]map[time.Time]float64
		}{{f[2], p.ip}, {f[3], p.cpi}} {
			if c.cell == "" {
				continue
			}
			v, err := strconv.ParseFloat(c.cell, 64)
			if err != nil {
				return nil, fmt.Errorf("macro panel: bad value %q: %w", c.cell, err)
			}
			if c.into[f[0]] == nil {
				c.into[f[0]] = map[time.Time]float64{}
			}
			c.into[f[0]][t] = v
		}
	}
	if len(p.ip) == 0 && len(p.cpi) == 0 {
		return nil, fmt.Errorf("macro panel: no data rows")
	}
	return p, nil
}

// breadth returns the share of countries of col whose year-on-year rate is
// accelerating at m, or ok=false when fewer than regimeMinCountries report.
func breadth(col map[string]map[time.Time]float64, m time.Time) (float64, bool) {
	yoy := func(s map[time.Time]float64, m time.Time) (float64, bool) {
		now, ok1 := s[m]
		prev, ok2 := s[m.AddDate(-1, 0, 0)]
		if !ok1 || !ok2 || prev == 0 {
			return 0, false
		}
		return now/prev - 1, true
	}
	var acc, tot int
	for _, s := range col {
		now, ok1 := yoy(s, m)
		then, ok2 := yoy(s, m.AddDate(0, -regimeAccelMonths, 0))
		if !ok1 || !ok2 {
			continue
		}
		tot++
		if now > then {
			acc++
		}
	}
	if tot < regimeMinCountries {
		return 0, false
	}
	return float64(acc) / float64(tot), true
}

// regimeAt returns the world season at month m (first of the month), or
// ok=false when the panel does not cover it.
func (p *macroPanel) regimeAt(m time.Time) (suggest.Category, bool) {
	var g, i float64
	var n int
	for k := range regimeSmoothMonths {
		mk := m.AddDate(0, -k, 0)
		gb, ok1 := breadth(p.ip, mk)
		ib, ok2 := breadth(p.cpi, mk)
		if ok1 && ok2 {
			g, i, n = g+gb, i+ib, n+1
		}
	}
	if n == 0 {
		return unclassified, false
	}
	growth, inflation := g/float64(n) >= 0.5, i/float64(n) >= 0.5
	switch {
	case growth && inflation:
		return suggest.Inflation, true
	case inflation:
		return suggest.Crisis, true
	case growth:
		return suggest.Growth, true
	default:
		return suggest.Deflation, true
	}
}

// loadMacroPanel parses the panel embedded in the binary.
func loadMacroPanel() (*macroPanel, error) { return parseMacroPanel(datasets.MacroPanel()) }
