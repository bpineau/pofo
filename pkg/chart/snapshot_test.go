package chart

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// snapshotCharts renders every SVG chart with fixed inputs. The snapshot test
// pins the exact output so the theme-token consolidation (and any later reskin)
// is a deliberate, reviewed change: byte-identical unless a token value moves.
func snapshotCharts() map[string]string {
	d0 := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	dates := []time.Time{d0, d0.AddDate(0, 0, 30), d0.AddDate(0, 0, 60), d0.AddDate(0, 0, 90)}
	opt := Options{Title: "Snap", Width: 640, Height: 360}
	return map[string]string{
		"line": Line(opt, []Series{
			{Name: "a", Dates: dates, Values: []float64{100, 108, 103, 120}},
			{Name: "b", Dates: dates, Values: []float64{100, 96, 101, 99}},
		}),
		"bars": Bars(opt, []Bar{{Label: "x", Value: 3, Text: "3"}, {Label: "y", Value: 7, Text: "7"}}),
		"fan": Fan(opt, "Years", [][]float64{
			{100, 90, 80, 70}, {100, 105, 108, 112}, {100, 130, 150, 180},
		}, [][]float64{{100, 95, 40, 0}, {100, 120, 140, 200}}),
		"stackedarea": StackedArea(opt, "t", "%", []AreaSeries{
			{Name: "f", Values: []float64{60, 50, 40}, Color: "#12B76A"},
			{Name: "b", Values: []float64{10, 20, 25}, Color: "#D92D20"},
		}),
		"multiline": MultiLine(opt, "wr", "ruin", []XYSeries{
			{Name: "m", Xs: []float64{2, 3, 4}, Ys: []float64{1, 8, 30}, Color: PaletteColor(0)},
		}, Marker{Axis: 'x', Value: 3, Label: "plan"}),
		"linedual": LineDual(opt, "x",
			XYSeries{Name: "l", Xs: []float64{0, 1, 2}, Ys: []float64{1, 2, 3}, Color: PaletteColor(0)},
			XYSeries{Name: "r", Xs: []float64{0, 1, 2}, Ys: []float64{9, 6, 3}, Color: PaletteColor(1)}),
		"hbars": HBars(opt, []Bar{{Label: "up", Value: 2, Text: "+2"}, {Label: "dn", Value: -3, Text: "-3"}}),
		"heatmap": Heatmap(opt, HeatmapData{
			Xs: []float64{1, 2}, Ys: []float64{1, 2}, Z: [][]float64{{0.1, 0.5}, {0.4, 0.9}}, XLabel: "x", YLabel: "y"}),
		"pie":      Pie(PieOptions{Title: "P"}, []Slice{{Label: "a", Value: 60}, {Label: "b", Value: 40}}),
		"spark":    Sparkline(SparkOptions{Width: 72, Height: 20}, []float64{1, 3, 2, 5, 4}),
		"gauge":    Gauge(opt, "30.8", "CAPE", "cheap", "rich", 0.9),
		"scatter":  Scatter(opt, "vol", "ruin", []LabeledPoint{{X: 2, Y: 14, Label: "Fixed", Color: themeBad}, {X: 18, Y: 1, Label: "VPW", Color: themeAccent}}),
		"category": CategoryBars(Options{Width: 460}, []CatBar{{Label: "Early", Value: 0.5, Text: "50%", Color: themeBad}}),
	}
}

func TestChartSnapshots(t *testing.T) {
	update := os.Getenv("UPDATE_SNAPSHOTS") != ""
	for name, svg := range snapshotCharts() {
		path := filepath.Join("testdata", "snap-"+name+".svg")
		want, err := os.ReadFile(path)
		if update || os.IsNotExist(err) {
			if werr := os.WriteFile(path, []byte(svg), 0o644); werr != nil {
				t.Fatalf("write %s: %v", path, werr)
			}
			continue
		}
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if string(want) != svg {
			t.Errorf("%s snapshot changed; if intentional (a reskin), rerun with UPDATE_SNAPSHOTS=1", name)
		}
	}
}
