package chart

import (
	"strings"
	"testing"
	"time"
)

func TestTermRendersSeriesWithoutANSI(t *testing.T) {
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	var dates []time.Time
	var a, b []float64
	for i := range 500 {
		dates = append(dates, start.AddDate(0, 0, i))
		a = append(a, 100+float64(i)*0.1)
		b = append(b, 100-float64(i)*0.02)
	}
	out := Term(TermOptions{Title: "Test", Width: 80, Height: 12, Color: false}, []Series{
		{Name: "Rising", Dates: dates, Values: a},
		{Name: "Falling", Dates: dates, Values: b},
	})
	if strings.Contains(out, "\x1b[") {
		t.Error("no ANSI escapes expected without color")
	}
	if !strings.Contains(out, "•") || !strings.Contains(out, "+") {
		t.Error("each series must have its own distinct marker")
	}
	if !strings.Contains(out, "Rising") || !strings.Contains(out, "Falling") {
		t.Error("missing legend")
	}
	if !strings.Contains(out, "2020") {
		t.Error("missing time tick")
	}
	if lines := strings.Count(out, "\n"); lines < 14 || lines > 18 {
		t.Errorf("unexpected height: %d lines", lines)
	}
}

func TestTermColorUsesANSI(t *testing.T) {
	dates := []time.Time{time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)}
	out := Term(TermOptions{Color: true}, []Series{{Name: "X", Dates: dates, Values: []float64{100, 120}}})
	if !strings.Contains(out, "\x1b[38;5;") {
		t.Error("ANSI escapes expected with color")
	}
}

func TestTermBraille(t *testing.T) {
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	var dates []time.Time
	var v []float64
	for i := range 200 {
		dates = append(dates, start.AddDate(0, 0, i))
		v = append(v, 100+float64(i))
	}
	series := []Series{{Name: "up", Dates: dates, Values: v}}
	out := Term(TermOptions{Width: 60, Height: 8, Braille: true}, series)
	brailleRunes := 0
	for _, r := range out {
		if r >= 0x2801 && r <= 0x28FF {
			brailleRunes++
		}
	}
	if brailleRunes == 0 {
		t.Fatalf("braille mode should emit braille runes:\n%s", out)
	}
	if !strings.Contains(out, "┤") || !strings.Contains(out, "└") {
		t.Error("braille mode should keep the gutter and axis frame")
	}
	if !strings.Contains(out, "2020") {
		t.Error("braille mode should keep the time labels")
	}
}

func TestTermBrailleMultiSeries(t *testing.T) {
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	var dates []time.Time
	var a, b []float64
	for i := range 100 {
		dates = append(dates, start.AddDate(0, 0, i))
		a = append(a, 100+float64(i))
		b = append(b, 200-float64(i))
	}
	series := []Series{
		{Name: "alpha", Dates: dates, Values: a},
		{Name: "beta", Dates: dates, Values: b},
	}
	out := Term(TermOptions{Width: 60, Height: 8, Braille: true, Color: true}, series)
	if !strings.Contains(out, "alpha") || !strings.Contains(out, "beta") {
		t.Error("braille multi-series should keep the legend")
	}
	if !strings.Contains(out, "\x1b[38;5;") {
		t.Error("Color should colorize braille cells")
	}
}
