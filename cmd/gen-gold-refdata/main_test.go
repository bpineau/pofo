package main

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

func TestParse(t *testing.T) {
	raw := `[{"is_cms_locked":0,"d":"1968-04-01","v":[37.7,15.68,null]},
		{"is_cms_locked":0,"d":"1968-04-02","v":[null,15.6,null]},
		{"is_cms_locked":0,"d":"1968-04-03","v":[37.6,15.63,null]}]`
	pts, err := parse([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if len(pts) != 2 || pts[1].Close != 37.6 || !pts[1].Date.Equal(time.Date(1968, 4, 3, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("got %v, want the two USD fixes with the USD-less day skipped", pts)
	}
	if _, err := parse([]byte(strings.Replace(raw, "1968-04-03", "1968-03-31", 1))); err == nil {
		t.Error("an out-of-order day was accepted")
	}
}

// history is a synthetic record shaped like the real one: a daily price that
// grows smoothly (log-linearly) through 38 in 1968, 40 at the end of 1971, 615
// in mid-1980 and 1770 in mid-2020, so no value repeats and no day jumps, and
// that ends on the day before now.
func history(now time.Time) []marketdata.Point {
	knots := []struct {
		at    time.Time
		level float64
	}{
		{time.Date(1968, 4, 1, 0, 0, 0, 0, time.UTC), 38},
		{time.Date(1971, 12, 31, 0, 0, 0, 0, time.UTC), 40},
		{time.Date(1980, 7, 1, 0, 0, 0, 0, time.UTC), 615},
		{time.Date(2020, 7, 1, 0, 0, 0, 0, time.UTC), 1770},
		{now, 2500},
	}
	var out []marketdata.Point
	k := 0
	for d := knots[0].at; d.Before(now); d = d.AddDate(0, 0, 1) {
		for d.After(knots[k+1].at) {
			k++
		}
		a, b := knots[k], knots[k+1]
		f := d.Sub(a.at).Hours() / b.at.Sub(a.at).Hours()
		out = append(out, marketdata.Point{Date: d, Close: a.level * math.Pow(b.level/a.level, f)})
	}
	return out
}

func TestCheck(t *testing.T) {
	now := time.Date(2026, 9, 27, 0, 0, 0, 0, time.UTC)
	pts := history(now)
	if err := check(pts, now); err != nil {
		t.Fatalf("a sound record was refused: %v", err)
	}
	if err := check(pts, now.AddDate(0, 1, 0)); err == nil {
		t.Error("a month-old last fix was accepted")
	}
	frozen := history(now)
	for i := len(frozen) - 10; i < len(frozen); i++ {
		frozen[i].Close = frozen[len(frozen)-11].Close
	}
	if err := check(frozen, now); err == nil {
		t.Error("a frozen tail was accepted")
	}
	pence := history(now)
	for i := range pence {
		pence[i].Close *= 100
	}
	if err := check(pence, now); err == nil {
		t.Error("a unit slip was accepted")
	}
}
