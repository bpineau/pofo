package chart

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

// day returns i days after 2020-01-01 UTC, the calendar every test here plots on.
func day(i int) time.Time {
	return time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, i)
}

// days returns n consecutive daily dates from 2020-01-01.
func days(n int) []time.Time {
	out := make([]time.Time, n)
	for i := range out {
		out[i] = day(i)
	}
	return out
}

// wellFormed checks that s is either omitted or a complete SVG document: the
// contract every primitive offers its callers, and the only thing a caller can
// rely on when the data it holds turns out to be degenerate.
func wellFormed(t *testing.T, name, s string) {
	t.Helper()
	if s == "" {
		return
	}
	if !strings.HasPrefix(s, "<svg ") || !strings.HasSuffix(s, "</svg>") {
		t.Errorf("%s: not a complete SVG document: %.80q ... %.20q", name, s, s[max(len(s)-20, 0):])
	}
	if strings.Contains(s, "NaN") || strings.Contains(s, "Inf") {
		t.Errorf("%s: a non-finite number reached the output", name)
	}
}

// The whole family must survive the inputs a caller cannot always avoid: no
// data at all, a single point, a flat series, non-finite values, negatives and
// a label far longer than the chart is wide. None may panic, and none may leak
// a NaN into a coordinate (an SVG with a NaN in a path silently draws nothing).
func TestDegenerateInputsNeverPanicNorLeakNaN(t *testing.T) {
	long := strings.Repeat("very long label ", 12)
	nan, inf := math.NaN(), math.Inf(1)
	opt := Options{Title: long, Width: 320, Height: 180, XLabel: long, YLabel: long}

	cases := []struct {
		name string
		draw func() string
	}{
		{"line/empty", func() string { return Line(opt, nil) }},
		{"line/no-points", func() string { return Line(opt, []Series{{Name: long}}) }},
		{"line/one-point", func() string {
			return Line(opt, []Series{{Name: long, Dates: days(1), Values: []float64{7}}})
		}},
		{"line/flat", func() string {
			return Line(opt, []Series{{Dates: days(3), Values: []float64{5, 5, 5}}})
		}},
		{"line/all-nan", func() string {
			return Line(opt, []Series{{Dates: days(3), Values: []float64{nan, inf, -inf}}})
		}},
		{"line/nan-hole", func() string {
			return Line(opt, []Series{{Dates: days(5), Values: []float64{1, nan, 3, nan, 5}}})
		}},
		{"line/negative", func() string {
			return Line(opt, []Series{{Dates: days(3), Values: []float64{-3, -8, -1}}})
		}},
		{"bars/empty", func() string { return Bars(opt, nil) }},
		{"bars/all-zero", func() string {
			return Bars(opt, []Bar{{Label: long, Value: 0}, {Label: "b", Value: 0}})
		}},
		{"bars/one", func() string { return Bars(opt, []Bar{{Label: long, Value: 1, Text: "1"}}) }},
		{"hbars/empty", func() string { return HBars(opt, nil) }},
		{"hbars/all-zero", func() string {
			return HBars(opt, []Bar{{Label: long}, {Label: "b"}})
		}},
		{"hbars/no-text", func() string {
			return HBars(opt, []Bar{{Label: long, Value: -2}, {Label: "b", Value: 3}})
		}},
		{"catbars/empty", func() string { return CategoryBars(opt, nil) }},
		{"catbars/negative", func() string {
			return CategoryBars(opt, []CatBar{{Label: long, Value: -0.5, Text: "-50%"}})
		}},
		{"catbars/default-size", func() string {
			return CategoryBars(Options{}, []CatBar{{Label: "a", Value: 0.4, Text: "40%"}})
		}},
		{"pie/empty", func() string { return Pie(PieOptions{Title: long}, nil) }},
		{"pie/all-non-positive", func() string {
			return Pie(PieOptions{Title: long}, []Slice{{Label: "a", Value: 0}, {Label: "b", Value: -1}})
		}},
		{"pie/lone-slice", func() string {
			return Pie(PieOptions{Title: long}, []Slice{{Label: long, Value: 3}})
		}},
		{"pie/sub-percent", func() string {
			return Pie(PieOptions{}, []Slice{{Label: "a", Value: 1000}, {Label: "b", Value: 3}})
		}},
		{"scatter/empty", func() string { return Scatter(opt, long, long, nil) }},
		{"scatter/origin", func() string {
			return Scatter(opt, long, long, []LabeledPoint{{X: 0, Y: 0, Label: long}})
		}},
		{"heatmap/empty", func() string { return Heatmap(opt, HeatmapData{}) }},
		{"heatmap/out-of-range", func() string {
			return Heatmap(opt, HeatmapData{Xs: []float64{1, 2}, Ys: []float64{1},
				Z: [][]float64{{-4, 9}}, XLabel: long, YLabel: long})
		}},
		{"gauge/empty", func() string { return Gauge(opt, "", long, "", "", 0) }},
		{"gauge/out-of-range", func() string { return Gauge(opt, "9", long, "l", "r", 4) }},
		{"spark/empty", func() string { return Sparkline(SparkOptions{}, nil) }},
		{"spark/flat", func() string { return Sparkline(SparkOptions{}, []float64{2, 2, 2}) }},
		{"stackedarea/empty", func() string { return StackedArea(opt, long, long, nil) }},
		{"stackedarea/one-step", func() string {
			return StackedArea(opt, long, long, []AreaSeries{{Values: []float64{1}}})
		}},
		{"multiline/empty", func() string { return MultiLine(opt, long, long, nil) }},
		{"multiline/one-x", func() string {
			return MultiLine(opt, long, long, []XYSeries{{Xs: []float64{2}, Ys: []float64{3}}})
		}},
		{"multiline/nameless-among-named", func() string {
			// Three names turn the end labels on; the one with no values must
			// be skipped rather than read past its end.
			return MultiLine(Options{Width: 480, Height: 240}, "x", "y", []XYSeries{
				{Name: "a", Xs: []float64{0, 1}, Ys: []float64{1, 2}},
				{Name: "b", Xs: []float64{0, 1}, Ys: []float64{2, 1}},
				{Name: "c"},
			})
		}},
		{"multiline/nan", func() string {
			return MultiLine(opt, long, long, []XYSeries{{Name: long, Xs: []float64{1, 2, 3}, Ys: []float64{nan, 2, inf}}})
		}},
		{"linedual/empty", func() string { return LineDual(opt, long, XYSeries{}, XYSeries{}) }},
		{"linedual/one-x", func() string {
			return LineDual(opt, long, XYSeries{Name: long, Xs: []float64{1}, Ys: []float64{1}},
				XYSeries{Name: "r", Xs: []float64{1}, Ys: []float64{nan}})
		}},
		{"linedual/short-ys", func() string {
			return LineDual(opt, long, XYSeries{Xs: []float64{1, 2, 3}, Ys: []float64{1}},
				XYSeries{Xs: []float64{1, 2, 3}, Ys: []float64{-1, -2, -3}})
		}},
		{"fan/empty", func() string { return Fan(opt, long, nil, nil) }},
		{"fan/samples-only", func() string {
			return Fan(opt, long, nil, [][]float64{{100, 90, 0}})
		}},
		{"fan/empty-bands", func() string {
			return Fan(opt, long, [][]float64{{}, {}}, [][]float64{{}, {50, 40}})
		}},
		{"divergingstack/empty", func() string { return DivergingStack(DivergingStackOptions{}, nil) }},
		{"divergingstack/all-zero", func() string {
			return DivergingStack(DivergingStackOptions{Title: long, Total: []float64{0, 0}},
				[]DivergingStackSeries{{Name: long, Values: []float64{0, 0}}})
		}},
		{"barmatrix/no-rows", func() string {
			return BarMatrix(BarMatrixOptions{}, []MatrixColumn{{Title: "a", Values: []float64{1}}})
		}},
		{"barmatrix/no-cols", func() string {
			return BarMatrix(BarMatrixOptions{RowLabels: []string{"a"}}, nil)
		}},
		{"barmatrix/all-zero", func() string {
			return BarMatrix(BarMatrixOptions{Title: long, RowLabels: []string{long, "b"}, Unit: "pts"},
				[]MatrixColumn{{Title: long, Values: []float64{0}}, {Title: "b", Values: []float64{0, 0}}})
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) { wellFormed(t, tc.name, tc.draw()) })
	}
}

// A chart drawn with a minimal style, a custom background and an intraday
// span takes the other side of every branch the default look takes.
func TestLineStyleBranches(t *testing.T) {
	base := day(0)
	intraday := []time.Time{base, base.Add(2 * time.Hour), base.Add(6 * time.Hour)}
	svg := Line(Options{Title: "T", Style: StyleMinimal()},
		[]Series{{Name: "a", Dates: intraday, Values: []float64{1, math.NaN(), 3}}})
	if strings.Contains(svg, "<rect") {
		t.Error(`Background "none" must draw no background rect`)
	}
	if !strings.Contains(svg, "fill-opacity=\"0.07\"") {
		t.Error("Fill must draw the area polygon")
	}
	// CornerDates on a sub-day span labels the clock, not the calendar.
	if !regexp.MustCompile(`>\d\d:\d\d<`).MatchString(svg) {
		t.Errorf("an intraday span must carry clock labels:\n%s", svg)
	}
	wellFormed(t, "minimal", svg)

	// An all-NaN series under Fill draws neither polygon nor path.
	blank := Line(Options{Style: StyleMinimal()},
		[]Series{{Dates: days(3), Values: []float64{math.NaN(), math.NaN(), math.NaN()}}})
	if strings.Contains(blank, "<polygon") || strings.Contains(blank, "<path") {
		t.Error("a series with no finite point must draw no ink")
	}

	// An explicit background colour is honoured.
	if svg := Line(Options{Style: Style{Background: "#123456"}},
		[]Series{{Dates: days(2), Values: []float64{1, 2}}}); !strings.Contains(svg, `fill="#123456"`) {
		t.Error("a custom background must be painted")
	}
}

// Axis helpers on the inputs that make their scale collapse.
func TestAxisHelpersOnDegenerateSpans(t *testing.T) {
	if got := niceStep(0, 6); got != 1 {
		t.Errorf("niceStep(0) = %g, want 1 (a collapsed span still needs a step)", got)
	}
	if got := niceStep(-4, 6); got != 1 {
		t.Errorf("niceStep(negative) = %g, want 1", got)
	}
	// A span whose 1/2/5 mantissa lands exactly on the decade boundary.
	if got := niceStep(6e-17, 6); got <= 0 || math.IsInf(got, 0) {
		t.Errorf("niceStep on a tiny span = %g, want a finite positive step", got)
	}
	// Large magnitudes compact rather than print every digit.
	for _, tc := range []struct {
		v, step float64
		want    string
	}{
		{15e6, 1e6, "15M"},
		{500e3, 1e5, "500k"},
		{0.004, 1, "0"},
		{1.25, 0.5, "1.2"},
		{1.25, 0.05, "1.25"},
	} {
		if got := fmtTick(tc.v, tc.step); got != tc.want {
			t.Errorf("fmtTick(%g, %g) = %q, want %q", tc.v, tc.step, got, tc.want)
		}
	}
}

// timeTicks must land on round years whatever the span, and never return a
// single lonely label.
func TestTimeTicksSpans(t *testing.T) {
	mk := func(y int, m time.Month, d int) time.Time {
		return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	}
	for _, tc := range []struct {
		name     string
		from, to time.Time
		step     int
	}{
		{"mid-year start, 5-year step", mk(1903, time.June, 15), mk(1998, time.March, 2), 5},
		{"twelve decades", mk(1900, time.January, 1), mk(2020, time.January, 1), 20},
		{"a decade", mk(2000, time.January, 1), mk(2012, time.January, 1), 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ticks := timeTicks(tc.from, tc.to)
			if len(ticks) < 2 {
				t.Fatalf("got %d ticks, want at least 2", len(ticks))
			}
			if len(ticks) > 11 {
				t.Errorf("got %d ticks, want at most 11", len(ticks))
			}
			for _, tk := range ticks {
				if y := tk.t.Year(); y%tc.step != 0 {
					t.Errorf("tick %s is not on a %d-year boundary", tk.label, tc.step)
				}
				if tk.t.Before(tc.from) || tk.t.After(tc.to) {
					t.Errorf("tick %s falls outside [%s, %s]", tk.label, tc.from, tc.to)
				}
			}
		})
	}
	// A span too short for two year boundaries falls back to month labels.
	ticks := timeTicks(mk(2020, time.March, 1), mk(2020, time.November, 1))
	if len(ticks) != 6 {
		t.Fatalf("got %d month ticks, want 6", len(ticks))
	}
	if !regexp.MustCompile(`^\d{4}-\d{2}$`).MatchString(ticks[0].label) {
		t.Errorf("short spans want month labels, got %q", ticks[0].label)
	}
}

// A Fan carries its reference markers, and ignores the horizontal ones (it
// has no y domain to hang them on). Five bands get their percentile names.
func TestFanMarkersAndBandNames(t *testing.T) {
	bands := [][]float64{
		{100, 60, 20}, {100, 80, 60}, {100, 100, 100}, {100, 130, 160}, {100, 180, 260},
	}
	svg := Fan(Options{Width: 640, Height: 360}, "year", bands, [][]float64{{100, 50, 0}},
		Marker{Axis: 'x', Value: 1, Label: "pension"},
		Marker{Axis: 'y', Value: 100, Label: "ignored"})
	if !strings.Contains(svg, ">pension<") {
		t.Error("a vertical marker must be drawn")
	}
	if strings.Contains(svg, "ignored") {
		t.Error("a fan must ignore horizontal markers")
	}
	hm := parseHover(t, svg)
	var names []string
	for _, s := range hm.Series {
		names = append(names, s.Name)
	}
	// Bands are published high to low, so the canonical names read top-down.
	want := []string{"p95", "p75", "median", "p25", "p5"}
	for _, w := range want {
		found := false
		for _, n := range names {
			if n == w {
				found = true
			}
		}
		if !found {
			t.Errorf("hover payload lacks the %q band; got %v", w, names)
		}
	}
	// An uncanonical band count falls back to positional names.
	hm = parseHover(t, Fan(Options{}, "year", [][]float64{{1, 2}, {3, 4}}, nil))
	found := false
	for _, s := range hm.Series {
		if s.Name == "band 1" || s.Name == "band 2" {
			found = true
		}
	}
	if !found {
		t.Errorf("a 2-band fan wants positional names, got %v", hm.Series)
	}
	// Three bands are the other canonical set.
	if got := fanBandNames(3); got[1] != "median" {
		t.Errorf("fanBandNames(3) = %v, want the median in the middle", got)
	}
}

// The DivergingStack's optional layers: a total line, a regime strip with its
// legend, an empty band, and series of unequal length (the shorter ones read
// as zero rather than shifting the stack).
func TestDivergingStackOptionalLayers(t *testing.T) {
	svg := DivergingStack(DivergingStackOptions{
		Title:       "contributions",
		XLabels:     []string{"2020", "", "2022", ""},
		XTips:       []string{"2020-01", "2021-01", "2022-01", "2023-01"},
		XLabel:      "year",
		YLabel:      "pts",
		Total:       []float64{2, -1, 3, 0},
		TotalName:   "net",
		Strip:       []StripBand{{From: 0, To: 1, Label: "growth", Color: ColorGood}, {From: 3, To: 2, Label: "dropped"}},
		StripName:   "regime",
		StripLegend: []Slice{{Label: "growth", Color: ColorGood}, {Label: "slump", Color: ColorBad}},
	}, []DivergingStackSeries{
		{Name: "equity", Values: []float64{3, -2, 4, 1}},
		{Name: "gold", Values: []float64{-1}}, // shorter: zero past its end
	})
	for _, want := range []string{">net<", ">regime<", ">growth<", ">slump<", ">2022<"} {
		if !strings.Contains(svg, want) {
			t.Errorf("rendered stack lacks %q", want)
		}
	}
	if strings.Contains(svg, "dropped") {
		t.Error("a band whose To precedes its From must be skipped")
	}
	hm := parseHover(t, svg)
	if hm.Kind != "stack" {
		t.Errorf("hover kind = %q, want %q", hm.Kind, "stack")
	}
	if len(hm.Rows) != 4 {
		t.Errorf("got %d hover headers, want one per x position", len(hm.Rows))
	}
	// The drawing pads a short series with zeros; the payload publishes only
	// the values it was given, which is what the front end indexes against.
	for _, s := range hm.Series {
		want := map[string]int{"equity": 4, "gold": 1, "net": 4}[s.Name]
		if len(s.Ys) != want {
			t.Errorf("series %q publishes %d values, want %d", s.Name, len(s.Ys), want)
		}
	}
	wellFormed(t, "divergingstack", svg)
}

// A BarMatrix's summary row, per-row colour overrides and short columns.
func TestBarMatrixSummaryAndRowColors(t *testing.T) {
	svg := BarMatrix(BarMatrixOptions{
		Title:        "per regime",
		RowLabels:    []string{"equity", "gold", "trend"},
		RowColors:    []string{"#123456", ""}, // second falls back to the palette
		Unit:         "pts/yr",
		Summary:      []float64{4.5, -1.5},
		SummaryLabel: "net",
	}, []MatrixColumn{
		{Title: "growth", Subtitle: "56 months", Color: ColorGood, Values: []float64{3, 1.5}}, // short: the third row reads zero
		{Title: "slump", Values: []float64{-2, 0.5, 0}},
	})
	for _, want := range []string{`fill="#123456"`, ">net<", ">growth<", ">56 months<"} {
		if !strings.Contains(svg, want) {
			t.Errorf("rendered matrix lacks %q", want)
		}
	}
	wellFormed(t, "barmatrix", svg)

	// No summary label: the row is still drawn, under the default name.
	svg = BarMatrix(BarMatrixOptions{RowLabels: []string{"a"}, Summary: []float64{1}},
		[]MatrixColumn{{Title: "c", Values: []float64{2}}})
	if !strings.Contains(svg, ">total<") {
		t.Error("an unnamed summary row must default to \"total\"")
	}
	// A summary shorter than the column count stops at its last value.
	wellFormed(t, "barmatrix/short-summary", BarMatrix(
		BarMatrixOptions{RowLabels: []string{"a"}, Summary: []float64{1}},
		[]MatrixColumn{{Title: "c", Values: []float64{2}}, {Title: "d", Values: []float64{3}}}))
}

// Scatter labels never escape the chart, however many points collide.
func TestScatterLabelsStayInsideTheFrame(t *testing.T) {
	const h = 200
	pts := make([]LabeledPoint, 0, 12)
	for i := range 12 {
		pts = append(pts, LabeledPoint{X: 1, Y: 0.02, Label: "policy " + strconv.Itoa(i)})
	}
	svg := Scatter(Options{Width: 480, Height: h}, "x", "y", pts)
	re := regexp.MustCompile(`<text x="[0-9.-]+" y="([0-9.-]+)" font-size="12"`)
	ms := re.FindAllStringSubmatch(svg, -1)
	if len(ms) != len(pts) {
		t.Fatalf("got %d labels, want %d", len(ms), len(pts))
	}
	for _, m := range ms {
		y, _ := strconv.ParseFloat(m[1], 64)
		if y > float64(h)-16 {
			t.Errorf("label at y=%.1f escapes the %dpx frame", y, h)
		}
	}
}

// A stacked area picks its own colours and skips unnamed layers in the legend.
func TestStackedAreaDefaultsAndLegend(t *testing.T) {
	svg := StackedArea(Options{Width: 480, Height: 240}, "year", "%", []AreaSeries{
		{Name: "alive", Values: []float64{100, 90, 80}},
		{Values: []float64{0, 10, 20}}, // unnamed: no legend entry, palette colour
	})
	if !strings.Contains(svg, PaletteColor(1)) {
		t.Errorf("an uncoloured layer must take its palette slot:\n%s", svg)
	}
	if got := strings.Count(svg, `<rect x=`); got == 0 {
		t.Error("the legend swatch of the named layer is missing")
	}
	if strings.Count(svg, "alive") == 0 {
		t.Error("the named layer must appear in the legend")
	}
	wellFormed(t, "stackedarea", svg)
}

// The terminal renderer on the inputs a CLI actually hands it.
func TestTermDegenerateInputs(t *testing.T) {
	const notEnough = "(not enough data to plot)\n"
	if got := Term(TermOptions{}, nil); got != notEnough {
		t.Errorf("no series: got %q, want %q", got, notEnough)
	}
	if got := Term(TermOptions{}, []Series{{Dates: days(1), Values: []float64{1}}}); got != notEnough {
		t.Errorf("one point: got %q, want %q", got, notEnough)
	}
	if got := Term(TermOptions{}, []Series{{Dates: days(3), Values: []float64{2, 2, 2}}}); got != notEnough {
		t.Errorf("flat series: got %q, want %q", got, notEnough)
	}
	// A series that starts after the window's left edge leaves that stretch
	// blank rather than extrapolating backwards.
	early := Series{Name: "early", Dates: days(400), Values: make([]float64, 400)}
	late := Series{Name: "late", Dates: days(400)[200:], Values: make([]float64, 200)}
	for i := range early.Values {
		early.Values[i] = float64(i)
		if i%37 == 0 {
			early.Values[i] = math.NaN() // holes must not break the plot
		}
	}
	for i := range late.Values {
		late.Values[i] = 400 - float64(i)
	}
	out := Term(TermOptions{Title: "two", Width: 60, Height: 8, Color: true}, []Series{early, late})
	if !strings.Contains(out, "two\n") || !strings.Contains(out, "early") || !strings.Contains(out, "late") {
		t.Errorf("title and legend missing:\n%s", out)
	}
	if !strings.Contains(out, "\x1b[38;5;") {
		t.Error("Color must emit ANSI sequences")
	}
	// Braille packs 2x4 dots per cell; without colors each series gets a marker.
	plain := Term(TermOptions{Width: 60, Height: 6, Braille: true}, []Series{early, late})
	if !regexp.MustCompile(`[\x{2800}-\x{28ff}]`).MatchString(plain) {
		t.Errorf("braille mode must draw braille cells:\n%s", plain)
	}
	// A span over three years labels years, a shorter one months.
	if !regexp.MustCompile(`\b202[0-9]\b`).MatchString(Term(TermOptions{Width: 60, Height: 6},
		[]Series{{Dates: days(1500), Values: ramp(1500)}})) {
		t.Error("a multi-year span wants year labels")
	}
	if !regexp.MustCompile(`\b\d{4}-\d{2}\b`).MatchString(Term(TermOptions{Width: 60, Height: 6},
		[]Series{{Dates: days(200), Values: ramp(200)}})) {
		t.Error("a sub-three-year span wants month labels")
	}
}

// ramp is a strictly increasing series of n values.
func ramp(n int) []float64 {
	out := make([]float64, n)
	for i := range out {
		out[i] = float64(i)
	}
	return out
}

// valueAt is the terminal renderer's sampler: before the first quote there is
// nothing to report, and a non-finite quote is a hole, not a value.
func TestValueAtBeforeFirstQuote(t *testing.T) {
	s := Series{Dates: days(3), Values: []float64{1, math.NaN(), 3}}
	if _, ok := valueAt(s, day(0).Unix()-1); ok {
		t.Error("a time before the first quote has no value")
	}
	if v, ok := valueAt(s, day(0).Unix()); !ok || v != 1 {
		t.Errorf("got (%g, %v), want (1, true)", v, ok)
	}
	if _, ok := valueAt(s, day(1).Unix()); ok {
		t.Error("a NaN quote must read as a hole")
	}
}

// SetDark switches every subsequent render to the terminal-dark theme, and
// leaves the light output untouched when off.
func TestSetDarkAppliesTheDarkChrome(t *testing.T) {
	t.Cleanup(func() { SetDark(false) })
	series := []Series{{Dates: days(3), Values: []float64{1, 2, 3}}}
	light := Line(Options{}, series)
	if !strings.Contains(light, themeSurface) {
		t.Fatal("the light render must carry the light surface")
	}
	SetDark(true)
	dark := Line(Options{}, series)
	if strings.Contains(dark, themeSurface) {
		t.Error("SetDark(true) must translate the surface colour")
	}
	if dark != Darken(light) {
		t.Error("SetDark must be exactly Darken applied at render time")
	}
	SetDark(false)
	if Line(Options{}, series) != light {
		t.Error("SetDark(false) must restore the light output byte for byte")
	}
}

// The colour helpers behind the palette search, on the inputs the palette
// itself never produces: lowercase hex, and a linear RGB component below zero
// (which a CVD simulation matrix can produce out of an in-gamut colour).
func TestColorHelpers(t *testing.T) {
	lo, up := [3]float64{}, [3]float64{}
	lo[0], lo[1], lo[2] = hexToLinear("#0a1bff")
	up[0], up[1], up[2] = hexToLinear("#0A1BFF")
	if lo != up {
		t.Errorf("hex parsing must be case-insensitive: %v vs %v", lo, up)
	}
	if r, _, _ := hexToLinear("#zzzzzz"); r != 0 {
		t.Errorf("a non-hex digit must contribute nothing, got %g", r)
	}
	neg := linearToOklab(-0.2, 0.1, 0.1)
	if math.IsNaN(neg[0]) {
		t.Error("a negative component must not produce a NaN lightness")
	}
}
