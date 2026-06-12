package report

import (
	"strings"
	"testing"
)

func TestRenderFoldsPortfolioSections(t *testing.T) {
	var b strings.Builder
	err := Render(&b, &Page{
		Title:       "Test",
		GeneratedAt: "01/01/2026 à 00:00",
		Portfolios: []PortfolioSection{
			{Name: "P1", Subtitle: "2000-01-01 → 2026-01-01 — rebalancement 30 j (#meta)", ChartSVG: "<svg></svg>"},
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
		t.Errorf("2 sections repliables attendues, trouvé %d", got)
	}
	if !strings.Contains(html, "<summary><span class=\"pf-name\">P1</span>") {
		t.Error("le nom du portefeuille doit être dans le summary")
	}
	if !strings.Contains(html, "rebalancement 30 j (#meta)") {
		t.Error("le sous-titre doit apparaître")
	}
	if strings.Contains(html, "<details open") {
		t.Error("les sections doivent être repliées par défaut")
	}
	// La synthèse (statistiques) précède les vues détaillées par portefeuille.
	if strings.Index(html, "Statistiques") > strings.Index(html, "<details class=\"pf\">") {
		t.Error("les statistiques doivent précéder les sections détaillées")
	}
}
