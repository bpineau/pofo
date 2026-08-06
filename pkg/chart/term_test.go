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
		{Name: "Hausse", Dates: dates, Values: a},
		{Name: "Baisse", Dates: dates, Values: b},
	})
	if strings.Contains(out, "\x1b[") {
		t.Error("pas d'échappements ANSI attendus sans couleur")
	}
	if !strings.Contains(out, "•") || !strings.Contains(out, "+") {
		t.Error("chaque série doit avoir son marqueur distinct")
	}
	if !strings.Contains(out, "Hausse") || !strings.Contains(out, "Baisse") {
		t.Error("légende manquante")
	}
	if !strings.Contains(out, "2020") {
		t.Error("graduation temporelle manquante")
	}
	if lines := strings.Count(out, "\n"); lines < 14 || lines > 18 {
		t.Errorf("hauteur inattendue: %d lignes", lines)
	}
}

func TestTermColorUsesANSI(t *testing.T) {
	dates := []time.Time{time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)}
	out := Term(TermOptions{Color: true}, []Series{{Name: "X", Dates: dates, Values: []float64{100, 120}}})
	if !strings.Contains(out, "\x1b[38;5;") {
		t.Error("échappements ANSI attendus en couleur")
	}
}
