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
