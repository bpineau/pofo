package report

import (
	"strings"
	"testing"
)

func TestRenderTextMarksBestCells(t *testing.T) {
	var b strings.Builder
	err := RenderText(&b, &Page{
		Title:          "Portefeuilles : A, B",
		CommonStart:    "2010-01-01",
		CommonEnd:      "2026-01-01",
		PortfolioNames: []string{"A", "B"},
		StatRows: []StatRow{
			{Label: "CAGR", Cells: []StatCell{{Text: "10.0 %", Best: true}, {Text: "8.0 %"}}},
			{Label: "Volatilité", Cells: []StatCell{{Text: "15.0 %"}, {Text: "12.0 %", Best: true}}},
		},
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	out := b.String()
	if !strings.Contains(out, "*10.0 %") || !strings.Contains(out, "*12.0 %") {
		t.Errorf("les meilleures cellules doivent être étoilées:\n%s", out)
	}
	if strings.Contains(out, "\x1b[") {
		t.Error("pas d'ANSI sans couleur")
	}
	if !strings.Contains(out, "Période commune : 2010-01-01 → 2026-01-01") {
		t.Error("période commune manquante")
	}
}
