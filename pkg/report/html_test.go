package report

import (
	"html/template"
	"strings"
	"testing"
)

func TestRenderFoldsPortfolioSections(t *testing.T) {
	var b strings.Builder
	err := Render(&b, &Page{
		Title:       "Test",
		GeneratedAt: "01/01/2026 at 00:00",
		Portfolios: []PortfolioSection{
			{Name: "P1", Subtitle: "2000-01-01 → 2026-01-01, rebalancing 30 d (#meta)", ChartSVG: "<svg></svg>"},
			{Name: "P2", ChartSVG: "<svg></svg>"},
		},
		PortfolioNames: []string{"P1", "P2"},
		CommonStart:    "2000-01-01",
		CommonEnd:      "2026-01-01",
	})
	if err != nil {
		t.Fatal(err)
	}
	html := b.String()
	if got := strings.Count(html, "<details class=\"pf\">"); got != 2 {
		t.Errorf("want 2 foldable sections, got %d", got)
	}
	if !strings.Contains(html, "<summary><span class=\"pf-name\">P1</span>") {
		t.Error("the portfolio name must be in the summary")
	}
	if !strings.Contains(html, "rebalancing 30 d (#meta)") {
		t.Error("the subtitle must appear")
	}
	if strings.Contains(html, "<details open") {
		t.Error("sections must be folded by default")
	}
	// The summary (statistics) precedes the detailed per-portfolio views.
	if strings.Index(html, "Statistics") > strings.Index(html, "<details class=\"pf\">") {
		t.Error("statistics must precede the detailed sections")
	}
}

func TestFireHref(t *testing.T) {
	page := &Page{
		Title:          "t",
		PortfolioNames: []string{"a"},
		Portfolios:     []PortfolioSection{{Name: "a", FireHref: "/fire/p/IWDA:100/"}},
	}
	var buf strings.Builder
	if err := Render(&buf, page); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), `class="pf-fire" href="/fire/p/IWDA:100/"`) {
		t.Error("FireHref link missing")
	}

	// Empty FireHref: no trace in the output (the CLI report is unchanged).
	page.Portfolios[0].FireHref = ""
	buf.Reset()
	if err := Render(&buf, page); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(buf.String(), "pf-fire") {
		t.Error("empty FireHref must leave no markup behind")
	}
}

func TestComposerSlot(t *testing.T) {
	page := &Page{Title: "t", PortfolioNames: []string{"a"},
		Portfolios: []PortfolioSection{{Name: "a"}}}
	page.Composer = template.HTML(`<div id="composer" data-x="1"></div>`)
	var buf strings.Builder
	if err := Render(&buf, page); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), `<div id="composer" data-x="1"></div>`) {
		t.Error("Composer markup missing")
	}
	// Empty Composer: no trace (CLI byte-identity).
	page.Composer = ""
	buf.Reset()
	if err := Render(&buf, page); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(buf.String(), "composer") {
		t.Error("empty Composer must leave no markup behind")
	}
}

func TestRenderSegmentedCoverage(t *testing.T) {
	var b strings.Builder
	err := Render(&b, &Page{
		Title: "Test",
		Portfolios: []PortfolioSection{{
			Name:          "P1",
			ChartSVG:      "<svg></svg>",
			CoverageLabel: "Macro-regime coverage (by weight)",
			Coverage: []CoverageBar{{
				Regime: "growth",
				Pct:    39,
				Segments: []CoverageSeg{
					{Width: 25.2, Color: "#0880A8", Tip: "NTSG 25%"},
					{Width: 9, Color: "#C2452B", Tip: "SMALL 9%"},
				},
				Detail: "NTSG 25 · SMALL 9",
			}},
		}},
		PortfolioNames: []string{"P1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	html := b.String()
	for _, want := range []string{
		`<span class="cov-seg" style="width:25.2%;background:#0880A8" data-tip="NTSG 25%"></span>`,
		`<span class="cov-seg" style="width:9%;background:#C2452B" data-tip="SMALL 9%"></span>`,
		`<div class="cov-detail">NTSG 25 · SMALL 9</div>`,
		`<script>`, `id="xtip"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("rendered coverage lacks %q", want)
		}
	}
}
