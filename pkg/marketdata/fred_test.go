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

// orientUSDEUR converts FRED dollars-per-euro into the requested orientation.
func TestOrientUSDEUR(t *testing.T) {
	usdPerEur := []Point{mon(2000, 1, 1.25), mon(2000, 2, 2.0)}

	eur := orientUSDEUR(usdPerEur, true) // euros per dollar = 1/rate
	if math.Abs(eur[0].Close-0.8) > 1e-9 || math.Abs(eur[1].Close-0.5) > 1e-9 {
		t.Errorf("eur-per-usd = %.3f, %.3f, want 0.8, 0.5", eur[0].Close, eur[1].Close)
	}
	usd := orientUSDEUR(usdPerEur, false) // unchanged dollars per euro
	if usd[0].Close != 1.25 || usd[1].Close != 2.0 {
		t.Errorf("usd-per-eur = %.3f, %.3f, want 1.25, 2.0", usd[0].Close, usd[1].Close)
	}
}
