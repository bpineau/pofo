package report

import (
	"errors"
	"html/template"
	"strings"
	"testing"
	"unicode/utf8"
)

// errWriter fails on its n-th write, so a renderer's error path can be
// exercised without a filesystem or a socket.
type errWriter struct {
	fail int
	n    int
}

var errFull = errors.New("no space left")

func (w *errWriter) Write(p []byte) (int, error) {
	w.n++
	if w.n >= w.fail {
		return 0, errFull
	}
	return len(p), nil
}

// bodyOf returns the document's rendered body, without the stylesheet and the
// interaction script: those name every optional class unconditionally, so only
// the body says whether a block was actually drawn.
func bodyOf(t *testing.T, html string) string {
	t.Helper()
	_, body, ok := strings.Cut(html, "<body>")
	if !ok {
		t.Fatal("no <body> in the rendered document")
	}
	body, _, ok = strings.Cut(body, "<script>")
	if !ok {
		t.Fatal("no <script> tail in the rendered document")
	}
	return body
}

// Every optional block of the document has two states, and the CLI report
// depends on the empty one leaving nothing behind. Render them both and check
// the markup appears only when its data does.
func TestRenderOptionalBlocks(t *testing.T) {
	full := &Page{
		Title:           "Portfolios",
		GeneratedAt:     "01/01/2026 at 00:00",
		RebalanceDays:   90,
		CompareSVG:      template.HTML(`<svg id="cmp"></svg>`),
		OverviewHeading: "Growth of 100",
		UnderwaterSVG:   template.HTML(`<svg id="uw"></svg>`),
		CommonStart:     "2000-01-01",
		CommonEnd:       "2026-01-01",
		PortfolioNames:  []string{"A"},
		StatRows: []StatRow{
			{Label: "CAGR", Hint: "compound annual growth rate", Cells: []StatCell{{Text: "7.0 %", Best: true}}},
			{Label: "Volatility", Cells: []StatCell{{Text: "12.0 %"}}},
		},
		Footnotes: []string{"real returns are CPI-deflated", "fees are informational"},
		SkinCSS:   template.CSS(`body{--accent:#8A5A2B}`),
		SiteNav:   template.HTML(`<nav id="sitenav"></nav>`),
		Composer:  template.HTML(`<div id="composer"></div>`),
		Portfolios: []PortfolioSection{{
			Name:              "A",
			Subtitle:          "rebalancing 90 d",
			ChartSVG:          template.HTML(`<svg id="pf"></svg>`),
			ContribSVG:        template.HTML(`<svg id="c12"></svg>`),
			ContribMonthlySVG: template.HTML(`<svg id="cm"></svg>`),
			Breakdowns:        []template.HTML{`<svg id="pie1"></svg>`, `<svg id="pie2"></svg>`},
			CoverageLabel:     "Factor coverage",
			Coverage: []CoverageBar{
				{Regime: "growth", Pct: 120, Segments: []CoverageSeg{{Width: 60, Color: "#0880A8"}}},
				{Regime: "inflation", Pct: 4, Gap: true, Detail: "IGLN 4"},
			},
			RiskBudget: []RiskRow{
				{Label: "equity", Capital: "60.0 %", Risk: "80.0 %", Return: "70.0 %", RiskWidth: 80, CapitalMark: 60},
				{Label: "trend", Capital: "10.0 %", Risk: "-5.0 %", Return: "2.0 %", RiskWidth: 0, CapitalMark: 10, Negative: true},
			},
			RegimeSVG: template.HTML(`<svg id="regime"></svg>`),
			Assets: []AssetRow{{Weight: "60 %", ID: "IWDA", Symbol: "IWDA.AS", Name: "iShares Core MSCI World",
				Class: "equity", UCITS: "yes", Fees: "0.20 %", Currency: "EUR", History: "2009-", CWARP: "-", Note: "core"}},
			Notes:    []string{"weights come from max-sharpe"},
			Warnings: []string{"DTLE looks like a distributing share class: its income is missing from every statistic"},
			FireHref: "/fire/p/IWDA:100/",
		}},
	}

	var b strings.Builder
	if err := Render(&b, full); err != nil {
		t.Fatal(err)
	}
	html := bodyOf(t, b.String())
	for _, want := range []string{
		`<svg id="cmp">`, `Growth of 100`, `<svg id="uw">`,
		`<nav id="sitenav">`, `<div id="composer">`,
		`title="compound annual growth rate"`, `class="n best">7.0 %`,
		`<li>real returns are CPI-deflated</li>`, `<li>fees are informational</li>`,
		`class="tbtn`, `<div class="tpane"><svg id="c12"></svg></div>`,
		`<div class="tpane" hidden><svg id="cm"></svg></div>`,
		`<div class="pies"><svg id="pie1"></svg><svg id="pie2"></svg></div>`,
		`Factor coverage`, `>120 %</span>`, `class="cov-val gap">4 % (gap)`,
		`<div class="cov-detail">IGLN 4</div>`,
		`class="rb-fill" style="width:80%"`, `class="rb-fill neg" style="width:0%"`,
		`class="rb-mark" style="left:60%"`,
		`<svg id="regime">`, `iShares Core MSCI World`,
		`<p class="note">weights come from max-sharpe</p>`,
		`<p class="warn">⚠ DTLE looks like a distributing share class`,
		`class="pf-fire"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("the full report lacks %q", want)
		}
	}
	// Two head-only options: the served skin, and the fire-link stylesheet
	// gated on a section actually carrying a link.
	for _, want := range []string{"--accent:#8A5A2B", ".pf-fire:hover"} {
		if !strings.Contains(b.String(), want) {
			t.Errorf("the document head lacks %q", want)
		}
	}

	// The same page stripped of every optional field: none of the markup the
	// options carry may survive, since the CLI report is that stripped page.
	bare := &Page{Title: "Portfolios", PortfolioNames: []string{"A"},
		Portfolios: []PortfolioSection{{Name: "A"}}}
	b.Reset()
	if err := Render(&b, bare); err != nil {
		t.Fatal(err)
	}
	doc := b.String()
	html = bodyOf(t, doc)
	for _, absent := range []string{
		"sitenav", "composer", "cov-", "rb-fill", "class=\"pies\"", "tbtn", "tpane",
		"<li>", "<p class=\"note\"", "<p class=\"warn\"", "pf-fire", "pf-sub",
	} {
		if strings.Contains(html, absent) {
			t.Errorf("the bare report leaks %q", absent)
		}
	}
	// The document itself is always complete.
	if !strings.HasPrefix(doc, "<!DOCTYPE html>") || !strings.HasSuffix(strings.TrimSpace(doc), "</html>") {
		t.Error("the bare report must still be a complete document")
	}
}

// A single contribution chart needs no tab strip: the toggle exists only to
// choose between the trailing-12m and the raw monthly window.
func TestContributionPanesOnlyWithBothWindows(t *testing.T) {
	page := &Page{Title: "t", PortfolioNames: []string{"a"},
		Portfolios: []PortfolioSection{{Name: "a", ContribSVG: template.HTML(`<svg id="c12"></svg>`)}}}
	var b strings.Builder
	if err := Render(&b, page); err != nil {
		t.Fatal(err)
	}
	html := bodyOf(t, b.String())
	if !strings.Contains(html, `<svg id="c12"></svg>`) {
		t.Error("the lone contribution chart must be drawn")
	}
	if strings.Contains(html, "tbtn") || strings.Contains(html, "tpane") {
		t.Error("one window needs no tab strip")
	}
}

// A section carrying a coverage chart but no risk budget (and the reverse)
// must still open the block that holds them.
func TestCoverageAndRiskBudgetAreIndependent(t *testing.T) {
	for _, tc := range []struct {
		name    string
		section PortfolioSection
		want    string
		absent  string
	}{
		{"coverage only", PortfolioSection{Name: "a", CoverageLabel: "Macro",
			Coverage: []CoverageBar{{Regime: "growth", Pct: 40}}}, "cov-row", "rb-head"},
		{"risk budget only", PortfolioSection{Name: "a",
			RiskBudget: []RiskRow{{Label: "equity", Risk: "80 %"}}}, "rb-head", "cov-row"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var b strings.Builder
			if err := Render(&b, &Page{Title: "t", PortfolioNames: []string{"a"},
				Portfolios: []PortfolioSection{tc.section}}); err != nil {
				t.Fatal(err)
			}
			body := bodyOf(t, b.String())
			if !strings.Contains(body, tc.want) {
				t.Errorf("missing %q", tc.want)
			}
			if strings.Contains(body, tc.absent) {
				t.Errorf("leaked %q", tc.absent)
			}
		})
	}
}

// A page with nothing in it at all still renders: the empty-portfolio case a
// caller hits when every fetch failed.
func TestRenderEmptyPage(t *testing.T) {
	var b strings.Builder
	if err := Render(&b, &Page{}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(b.String(), "<title></title>") {
		t.Error("an empty page must still produce a document head")
	}
	if strings.Contains(b.String(), `<details class="pf">`) {
		t.Error("no portfolios, no sections")
	}
}

// Render must report a broken sink rather than swallow it.
func TestRenderPropagatesWriteErrors(t *testing.T) {
	if err := Render(&errWriter{fail: 1}, &Page{Title: "t"}); !errors.Is(err, errFull) {
		t.Errorf("Render error = %v, want %v", err, errFull)
	}
}

// The terminal table: colored best cells, and the header written before the
// sink breaks.
func TestRenderTextColorAndWriteErrors(t *testing.T) {
	page := &Page{
		Title:          "Portfolios: A, B",
		CommonStart:    "2010-01-01",
		CommonEnd:      "2026-01-01",
		PortfolioNames: []string{"A", "B"},
		StatRows: []StatRow{
			{Label: "CAGR", Cells: []StatCell{{Text: "10.0 %", Best: true}, {Text: "8.0 %"}}},
		},
	}
	var b strings.Builder
	if err := RenderText(&b, page, true); err != nil {
		t.Fatal(err)
	}
	out := b.String()
	if !strings.Contains(out, "\x1b[32;1m10.0 %\x1b[0m") {
		t.Errorf("the best cell must be green:\n%q", out)
	}
	if strings.Contains(out, "*10.0 %") {
		t.Error("with colors the star is redundant and must not appear")
	}
	// The colored cell keeps its column: the ANSI escape rides inside the pad.
	if !strings.Contains(out, "  \x1b[32;1m") {
		t.Errorf("the padding must precede the escape, not follow it:\n%q", out)
	}
	if err := RenderText(&errWriter{fail: 1}, page, false); !errors.Is(err, errFull) {
		t.Errorf("RenderText error = %v, want %v", err, errFull)
	}
}

// Column widths are counted in runes, so a table whose labels and values
// carry accents, arrows or a thousands separator still lines up. Rune width
// is the right proxy only as long as the report avoids double-width glyphs;
// this test is what says so.
func TestRenderTextAlignsOnNonASCII(t *testing.T) {
	var b strings.Builder
	if err := RenderText(&b, &Page{
		Title:          "Portefeuilles",
		CommonStart:    "2010-01-01",
		CommonEnd:      "2026-01-01",
		PortfolioNames: []string{"Cible", "Actuel (héritée)"},
		StatRows: []StatRow{
			{Label: "Volatilité", Cells: []StatCell{{Text: "12.0 %"}, {Text: "15.0 %", Best: true}}},
			{Label: "Perte maximale (drawdown)", Cells: []StatCell{{Text: "−34.1 %", Best: true}, {Text: "−41.0 %"}}},
			{Label: "CAGR", Cells: []StatCell{{Text: "7.0 %"}, {Text: "6.0 %"}}},
		},
	}, false); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimRight(b.String(), "\n"), "\n")
	// Title, common period, blank, header, rule, then one line per row.
	if len(lines) != 8 {
		t.Fatalf("got %d lines, want 8:\n%s", len(lines), b.String())
	}
	width := utf8.RuneCountInString(lines[4]) // the rule spans the whole table
	for i, l := range lines[3:] {
		if got := utf8.RuneCountInString(l); got != width {
			t.Errorf("line %d is %d runes wide, want %d:\n%q", i+3, got, width, l)
		}
	}
	// The widest label sets the first column, and the values right-align.
	if !strings.HasPrefix(lines[3], strings.Repeat(" ", 0)+"Metric") {
		t.Errorf("header must start with the metric column: %q", lines[3])
	}
	if !strings.HasSuffix(lines[5], "*15.0 %") {
		t.Errorf("the best cell must sit flush right: %q", lines[5])
	}
}

// A row with fewer cells than there are portfolios (a statistic one portfolio
// cannot produce) must not misalign the ones that follow.
func TestRenderTextShortRow(t *testing.T) {
	var b strings.Builder
	if err := RenderText(&b, &Page{
		PortfolioNames: []string{"A", "B", "C"},
		StatRows: []StatRow{
			{Label: "CAGR", Cells: []StatCell{{Text: "7.0 %"}, {Text: "6.0 %"}, {Text: "5.0 %"}}},
			{Label: "CWARP", Cells: []StatCell{{Text: "1.0"}}},
		},
	}, false); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimRight(b.String(), "\n"), "\n")
	// An empty title still occupies its line, so: title, period, blank,
	// header, rule, then the two rows.
	if n := len(lines); n != 7 {
		t.Fatalf("got %d lines, want 7:\n%s", n, b.String())
	}
	if !strings.HasPrefix(lines[6], "CWARP") || !strings.HasSuffix(lines[6], "1.0") {
		t.Errorf("the short row must render what it has: %q", lines[6])
	}
}
