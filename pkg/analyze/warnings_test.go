package analyze

import (
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

func day(y int, m time.Month, d int) time.Time { return time.Date(y, m, d, 0, 0, 0, 0, time.UTC) }

// Each data problem a series carries inside the studied window is a warning,
// and one outside it is not.
func TestSeriesWarnings(t *testing.T) {
	s := &marketdata.Series{
		Symbol: "X",
		Source: "stooq",
		Points: []marketdata.Point{
			{Date: day(2020, 1, 1), Close: 100},
			{Date: day(2020, 1, 2), Close: 101},
			{Date: day(2020, 1, 3), Close: 102},
		},
		SimulatedBefore: day(2020, 1, 2),
		EstimatedFrom:   day(2020, 1, 3),
		EstimateProxy:   "URTH",
		Junctions:       []time.Time{day(2019, 12, 1), day(2020, 1, 2)},
	}
	got := strings.Join(seriesWarnings(s), "\n")
	for _, want := range []string{
		"simulated before 2020-01-02:",
		"estimated from 2020-01-03 via URTH",
		"Stooq closes are not dividend-adjusted",
		"definition change on 2020-01-02",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("warnings lack %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "2019-12-01") {
		t.Errorf("a junction before the window is reported:\n%s", got)
	}

	clean := &marketdata.Series{
		Source:          "ft",
		Name:            "Some UCITS ETF (Acc)",
		Points:          s.Points,
		SimulatedBefore: day(2019, 1, 1), // reconstruction entirely before the window
		EstimatedFrom:   day(2021, 1, 1), // nowcast entirely after it
	}
	if w := seriesWarnings(clean); len(w) != 0 {
		t.Errorf("clean series warned: %q", w)
	}
	if w := seriesWarnings(&marketdata.Series{}); w != nil {
		t.Errorf("empty series warned: %q", w)
	}
}
