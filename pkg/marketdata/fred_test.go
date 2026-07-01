package marketdata

import (
	"math"
	"testing"
	"time"
)

func mon(y int, m time.Month, v float64) Point {
	return Point{Date: time.Date(y, m, 1, 0, 0, 0, 0, time.UTC), Close: v}
}

// extendMonthlyBack chains an older index in front of a base index, rescaled so
// the level is continuous across the splice.
func TestExtendMonthlyBack(t *testing.T) {
	base := []Point{mon(2000, 1, 100), mon(2000, 2, 101)}
	older := []Point{mon(1999, 11, 50), mon(1999, 12, 52), mon(2000, 1, 54), mon(2000, 2, 55)}

	got := extendMonthlyBack(base, older)

	if len(got) != 4 {
		t.Fatalf("len = %d, want 4 (2 prepended + 2 base)", len(got))
	}
	if !got[0].Date.Equal(mon(1999, 11, 0).Date) {
		t.Errorf("first date = %s, want 1999-11", got[0].Date.Format("2006-01"))
	}
	// The splice preserves the base levels exactly.
	if got[2].Close != 100 || got[3].Close != 101 {
		t.Errorf("base levels changed: %.3f, %.3f", got[2].Close, got[3].Close)
	}
	// Prepended points are rescaled by base/older at the anchor (100/54): the
	// 1999-12 point becomes 52×100/54.
	if want := 52 * 100.0 / 54.0; math.Abs(got[1].Close-want) > 1e-9 {
		t.Errorf("rescaled 1999-12 = %.4f, want %.4f", got[1].Close, want)
	}
}

// A base with no earlier proxy data is returned unchanged.
func TestExtendMonthlyBackNoEarlier(t *testing.T) {
	base := []Point{mon(2000, 1, 100)}
	older := []Point{mon(2001, 1, 100)} // all after the anchor
	if got := extendMonthlyBack(base, older); len(got) != 1 {
		t.Errorf("expected no extension, got %d points", len(got))
	}
}
