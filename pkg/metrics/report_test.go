package metrics

import (
	"testing"
	"time"
)

func TestStandardWindows(t *testing.T) {
	to := time.Date(2026, 7, 4, 0, 0, 0, 0, time.UTC)
	ws := StandardWindows(to)
	byName := map[string]Window{}
	for _, w := range ws {
		byName[w.Name] = w
	}
	day := func(y, m, d int) time.Time { return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC) }
	cases := []struct {
		name     string
		from, to time.Time
	}{
		{"1d", day(2026, 7, 3), to},
		{"7d", day(2026, 6, 27), to},
		{"1m", day(2026, 6, 4), to},
		{"3m", day(2026, 4, 4), to},
		{"ytd", day(2025, 12, 31), to},
		{"1y", day(2025, 7, 4), to},
		{"prev-yr", day(2024, 12, 31), day(2025, 12, 31)},
	}
	if len(ws) != len(cases) {
		t.Fatalf("window count: %d", len(ws))
	}
	for _, c := range cases {
		w, ok := byName[c.name]
		if !ok || !w.From.Equal(c.from) || !w.To.Equal(c.to) {
			t.Errorf("%s: got %+v, want [%s, %s]", c.name, w, c.from, c.to)
		}
	}
}
