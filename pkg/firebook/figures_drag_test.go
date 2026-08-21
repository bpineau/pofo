package firebook

import (
	"fmt"
	"math"
	"strings"
	"testing"
)

// The plate is the article's table, redrawn: every point it poses must
// reproduce the row it comes from, drag and geometric result alike.
func TestDragRecomputesTheArticlesTable(t *testing.T) {
	art := bookArticle(t, "rendements-arithmetiques-geometriques")
	for _, row := range []struct {
		line            string
		sigma           float64
		drag, geometric float64
	}{
		{"| Fonds monétaire | ~1 % | ~0,005 % | 7,0 % |", 1, 0.005, 7.0},
		{"| Portefeuille 60/40 | ~10 % | ~0,5 % | 6,5 % |", 10, 0.5, 6.5},
		{"| Actions mondiales | ~15 % | ~1,1 % | 5,9 % |", 15, 1.1, 5.9},
		{"| Empilement 90/60 (type NTSG) | ~15 % | ~1,1 % | 5,9 % |", 15, 1.1, 5.9},
		{"| Actions émergentes | ~22 % | ~2,4 % | 4,6 % |", 22, 2.4, 4.6},
		{"| Actions à levier ×2 | ~30 % | ~4,5 % | ~2,5 % (avant frais de levier !) |", 30, 4.5, 2.5},
	} {
		if !strings.Contains(art, row.line) {
			t.Errorf("the article's table no longer carries %q", row.line)
		}
		if got := dragDrag(row.sigma); math.Abs(got-row.drag) > 0.055 {
			t.Errorf("σ = %.0f %%: drag %.3f points, the table says %.3f", row.sigma, got, row.drag)
		}
		if got := dragGeometric(row.sigma); math.Abs(got-row.geometric) > 0.055 {
			t.Errorf("σ = %.0f %%: geometric %.2f %%, the table says %.1f %%", row.sigma, got, row.geometric)
		}
	}
	// The table's premise, which the plate holds constant.
	for _, want := range []string{
		"La dernière colonne part chaque fois de 7 % arithmétique, pour isoler le seul effet de la volatilité.",
		"il double l'arithmétique mais quadruple la variance",
		"::: figure drag-volatilite",
	} {
		if !strings.Contains(art, want) {
			t.Errorf("the article no longer says %q", want)
		}
	}
	// The plate poses exactly the table's assets, at the table's volatilities.
	posed := 0
	for _, a := range dragAssets {
		posed += len(a.names)
	}
	if posed != 6 {
		t.Errorf("the plate poses %d assets, the table has six rows", posed)
	}
	// Global equities and the 90/60 stack share one point, which is the
	// comparison the article's conclusion rests on.
	shared := false
	for _, a := range dragAssets {
		if len(a.names) == 2 && a.sigma == 15 {
			shared = true
		}
	}
	if !shared {
		t.Error("the plate no longer puts the two 15 % assets on one point")
	}
}

// The quadratic is the whole argument: doubling the volatility must quadruple
// the drag, and the two arrows must read what the prose reads.
func TestDragIsQuadraticAndTheArrowsAgree(t *testing.T) {
	if got := dragDrag(20) / dragDrag(10); math.Abs(got-4) > 1e-9 {
		t.Errorf("doubling σ multiplies the drag by %.2f, not four", got)
	}
	if got := dragDrag(30) / dragDrag(10); math.Abs(got-9) > 1e-9 {
		t.Errorf("tripling σ multiplies the drag by %.2f, not nine", got)
	}
	if dragGeometric(0) != dragArith {
		t.Errorf("with no volatility the plan compounds at %.2f %%, not its own mean", dragGeometric(0))
	}
	// The two moves the plate draws under the axis, in points of compounded
	// return: diversifying pays, daily leverage costs, and it costs more.
	div := dragMove(dragDiversifyFrom, dragDiversifyTo)
	lev := dragMove(dragLeverFrom, dragLeverTo)
	if frNum(div, 1) != "1,9" {
		t.Errorf("diversifying from 22 to 10 pays %s points, the arrow says 1,9", frNum(div, 1))
	}
	if frNum(-lev, 1) != "3,4" {
		t.Errorf("levering from 15 to 30 costs %s points, the arrow says 3,4", frNum(-lev, 1))
	}
	if div <= 0 || lev >= 0 {
		t.Errorf("the arrows point the wrong way: %.2f and %.2f", div, lev)
	}
	if -lev <= div {
		t.Error("the leverage arrow no longer costs more than the diversification arrow pays")
	}
}

// The rendered plate obeys the book's drawing rules and carries every reading
// it claims to draw.
func TestDragPlateRenders(t *testing.T) {
	svg := FigureSVG("drag-volatilite")
	if !strings.HasPrefix(svg, "<svg viewBox=") {
		t.Fatal("the plate must render an SVG")
	}
	// U+2014 and U+2013 are the em- and en-dashes, banned from the book down to
	// its figure labels; they stay escaped here so a repository-wide grep never
	// trips on them.
	for _, banned := range []string{"\u2014", "\u2013", "rgba(", "opacity", "rotate("} {
		if strings.Contains(svg, banned) {
			t.Errorf("the plate uses %q, which the book's figures never do", banned)
		}
	}
	for _, a := range dragAssets {
		for _, n := range a.names {
			if !strings.Contains(svg, ">"+n+"<") {
				t.Errorf("the plate does not name %q", n)
			}
		}
		want := fmt.Sprintf(">%s %%<", frNum(dragGeometric(a.sigma), 1))
		if !strings.Contains(svg, want) {
			t.Errorf("the plate does not draw the reading %q", want)
		}
	}
	for _, want := range []string{
		">diversifier : moins de σ, +1,9 point de géométrique<",
		">levier quotidien : deux fois le σ, −3,4 points<",
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("the plate does not draw %q", want)
		}
	}
	// One dot per posed point, and the two arrowheads: no other shape.
	if n := strings.Count(svg, "<circle"); n != len(dragAssets) {
		t.Errorf("%d dots drawn, expected one per posed point (%d)", n, len(dragAssets))
	}
	if n := strings.Count(svg, "<path"); n != 2 {
		t.Errorf("%d filled paths drawn, expected the two arrowheads", n)
	}
	if n := strings.Count(svg, "<polyline"); n != 1 {
		t.Errorf("%d curves drawn, the plate has exactly one", n)
	}
}
