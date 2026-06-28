package report

import (
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
