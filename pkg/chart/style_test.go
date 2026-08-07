package chart

import (
	"strings"
	"testing"
	"time"
)

func styleFixture() []Series {
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	dates := make([]time.Time, 40)
	values := make([]float64, 40)
	for i := range dates {
		dates[i] = start.AddDate(0, 0, i*30)
		values[i] = 100 + float64(i*i%37)
	}
	return []Series{{Name: "fixture", Dates: dates, Values: values}}
}

// The zero-value Style must not change Line's output at all: existing
// callers (reports, the fire UI) rely on the default chrome.
func TestLineZeroStyleUnchanged(t *testing.T) {
	svg := Line(Options{Width: 640, Height: 300}, styleFixture())
	for _, want := range []string{
		`fill="#FFFFFF"`,               // background rect
		`stroke="#E9EDF3"`,             // grid lines
		`stroke="#C6CEDA"`,             // axes
		`font-family="-apple-system, `, // default font
		`stroke-width="1.8"`,           // default stroke
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("zero-style Line output lost %q", want)
		}
	}
	if strings.Contains(svg, "fill-opacity") {
		t.Error("zero-style Line should not draw an area fill")
	}
}

func TestLineMinimalStyle(t *testing.T) {
	opt := Options{Width: 640, Height: 300, Style: StyleMinimal()}
	svg := Line(opt, styleFixture())
	if strings.Contains(svg, "<rect") {
		t.Error("minimal style should not draw a background rect")
	}
	if !strings.Contains(svg, "ui-monospace") {
		t.Error("minimal style should use the monospace font")
	}
	if !strings.Contains(svg, "fill-opacity") {
		t.Error("minimal style should fill under the first series")
	}
	if strings.Contains(svg, `stroke="#E9EDF3"`) || strings.Contains(svg, `stroke="#C6CEDA"`) {
		t.Error("minimal style should draw neither grid nor axes")
	}
	if !strings.Contains(svg, "2020-01-01") {
		t.Error("minimal style should label the first date at the corner")
	}
}

// A legend appears for multi-series charts unless hidden.
func TestLineHideLegend(t *testing.T) {
	series := append(styleFixture(), styleFixture()...)
	series[1].Name = "second"
	if svg := Line(Options{}, series); !strings.Contains(svg, ">second</text>") {
		t.Error("default multi-series chart should show a legend")
	}
	opt := Options{Style: Style{HideLegend: true}}
	if svg := Line(opt, series); strings.Contains(svg, ">second</text>") {
		t.Error("HideLegend should drop the legend row")
	}
}

func TestCompact(t *testing.T) {
	cases := []struct {
		in   float64
		want string
	}{
		{1234567, "1.23M"},
		{473900, "473.9k"},
		{1500, "1.5k"},
		{12.345, "12.35"},
		{-2500000, "-2.5M"},
		{0, "0"},
	}
	for _, c := range cases {
		if got := Compact(c.in); got != c.want {
			t.Errorf("Compact(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}

// Fill on an empty chart must not panic.
func TestLineFillEmpty(t *testing.T) {
	if got := Line(Options{Style: Style{Fill: true}}, nil); got == "" {
		t.Fatal("Line should still render an empty frame")
	}
}
