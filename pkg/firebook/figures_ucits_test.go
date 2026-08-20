package firebook

import (
	"fmt"
	"math"
	"strings"
	"testing"
)

// A placement grid that does not add up to the whole portfolio would be a lie
// the eye cannot catch, since the areas only ever say "more" and "less". Both
// editions' tables are checked here, and so is the identity that makes the two
// comparable: the same seven blocks carry the same weights, and only the
// wrappers under them change.
func TestUcitsGridsSumToOneHundred(t *testing.T) {
	for name, l := range map[string]ucitsLayout{"FR": ucitsFR, "EN": ucitsEN} {
		if got := l.ucitsSum(); math.Abs(got-100) > 1e-9 {
			t.Errorf("%s grid sums to %.1f %%, not 100 %%", name, got)
		}
		if len(l.rows) != 7 {
			t.Errorf("%s grid has %d blocks, the article says seven", name, len(l.rows))
		}
		for _, r := range l.rows {
			held := 0
			for _, c := range r.cells {
				if c.weight > 0 {
					held++
				}
				if c.barred && c.weight > 0 {
					t.Errorf("%s grid, block %q: a barred cell also holds %.0f %%", name, r.brick, c.weight)
				}
			}
			if held == 0 {
				t.Errorf("%s grid, block %q sits nowhere", name, r.brick)
			}
			if r.role == "" {
				t.Errorf("%s grid, block %q serves no named role", name, r.brick)
			}
		}
	}
	// The blocks are the book's, so both editions weigh them the same way:
	// only the columns move.
	for i, fr := range ucitsFR.rows {
		en := ucitsEN.rows[i]
		if fr.brick != en.brick || fr.role != en.role {
			t.Errorf("row %d: the two editions no longer draw the same block (%q / %q)", i, fr.brick, en.brick)
		}
		sum := func(r ucitsRow) float64 {
			return r.cells[0].weight + r.cells[1].weight + r.cells[2].weight
		}
		if math.Abs(sum(fr)-sum(en)) > 1e-9 {
			t.Errorf("block %q weighs %.0f %% in French and %.0f %% in English",
				fr.brick, sum(fr), sum(en))
		}
	}
	// The French wrapper totals are the article's own envelope amounts over
	// its 1,6 M€: 300 k€, 130 k€ and 1 170 k€.
	for i, k := range []float64{300, 130, 1170} {
		if got := ucitsFR.ucitsTotal(i); math.Abs(got-k/1600*100) > 0.6 {
			t.Errorf("French column %d carries %.0f %%, but the article's %.0f k€ is %.1f %%",
				i, got, k, k/1600*100)
		}
	}
	// The English wrapper totals are the shares the US article states.
	for i, want := range []float64{50, 13, 37} {
		if got := ucitsEN.ucitsTotal(i); math.Abs(got-want) > 1e-9 {
			t.Errorf("English column %d carries %.0f %%, the article says %.0f %%", i, got, want)
		}
	}
}

// Every weight is quoted by an article, so the prose is the fixture. The two
// editions are checked separately because the English one is an adaptation
// with its own accounts and its own split.
func TestUcitsWeightsComeFromTheArticles(t *testing.T) {
	fr, err := assets.ReadFile("assets/book/fr/etf-ucits-europeens.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"Le PEA (300 k€) loge l'ETF Monde synthétique Acc → 19 %",
		"Le CTO (1 170 k€) loge le reste",
		"All-World physique IE Acc → 38 %",
		"SCV US et Europe → 8 %",
		"État euro 5-8 ans Acc → 11 %",
		"linkers euro 1-5 ans → 6 %",
		"ETC or physique → 5 %",
		"fonds trend UCITS à frais fixes → 5 %",
		"L'AV (130 k€) loge le fonds euros → 8 %",
		"environ 0,22 %/an tout compris",
		"::: figure ucits-implantation",
	} {
		if !strings.Contains(string(fr), want) {
			t.Errorf("the French article no longer says %q", want)
		}
	}
	en, err := assets.ReadFile("assets/book/en/building-it-with-us-etfs.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"holds the total world fund at 34%, small-cap value at 8%, and the T-bill fund at 8%",
		"The **Roth** (13%) holds nothing but that same world fund",
		"intermediate Treasuries 11%, short TIPS 6%, gold 5%, managed futures 5%, and the last 10% of the world fund",
		"about **0.12% a year** all in",
		"::: figure ucits-implantation",
	} {
		if !strings.Contains(string(en), want) {
			t.Errorf("the English article no longer says %q", want)
		}
	}
}

// The rendered grids obey the book's drawing rules and carry every weight,
// wrapper and total they claim to draw. The English one is rendered through
// the edition's own path, since its table is not a translation of the French
// one but the US article's own example.
func TestUcitsPlatesRender(t *testing.T) {
	for name, svg := range map[string]string{
		"FR": FigureSVG("ucits-implantation"),
		"EN": FigureSVGEnglish("ucits-implantation"),
	} {
		if !strings.HasPrefix(svg, "<svg viewBox=") {
			t.Fatalf("%s: the plate must render an SVG", name)
		}
		// U+2014 and U+2013 are the em- and en-dashes, banned from the book
		// down to its figure labels; they stay escaped here so a
		// repository-wide grep never trips on them.
		for _, banned := range []string{"rgba(", "opacity", "\u2014", "\u2013", "rotate("} {
			if strings.Contains(svg, banned) {
				t.Errorf("%s: the plate uses %q, which the book's figures never do", name, banned)
			}
		}
	}
	fr := FigureSVG("ucits-implantation")
	for _, r := range ucitsFR.rows {
		if !strings.Contains(fr, ">"+r.brick+"<") || !strings.Contains(fr, ">"+r.role+"<") {
			t.Errorf("the French plate does not draw the block %q with its role", r.brick)
		}
	}
	for _, c := range ucitsFR.columns {
		if !strings.Contains(fr, ">"+c.name+"<") {
			t.Errorf("the French plate does not draw the wrapper %q", c.name)
		}
	}
	for i := range ucitsFR.columns {
		if want := fmt.Sprintf(">%.0f %%<", ucitsFR.ucitsTotal(i)); !strings.Contains(fr, want) {
			t.Errorf("the French plate does not print the column total %q", want)
		}
	}
	// One square per held cell, one quiet dot per barred cell.
	squares, dots := 0, 0
	for _, r := range ucitsFR.rows {
		for _, c := range r.cells {
			if c.weight > 0 {
				squares++
			}
			if c.barred {
				dots++
			}
		}
	}
	if n := strings.Count(fr, "<rect"); n != squares {
		t.Errorf("%d squares drawn, expected one per held cell (%d)", n, squares)
	}
	if n := strings.Count(fr, "<circle"); n != dots {
		t.Errorf("%d dots drawn, expected one per barred cell (%d)", n, dots)
	}
	// The English plate draws its own wrappers, and none of the French ones.
	en := FigureSVGEnglish("ucits-implantation")
	for _, want := range []string{">taxable<", ">Roth<", ">traditional 401(k) and IRA<", ">50%<", ">13%<", ">37%<"} {
		if !strings.Contains(en, want) {
			t.Errorf("the English plate does not draw %q", want)
		}
	}
	if strings.Contains(en, ">PEA<") || strings.Contains(en, ">73%<") {
		t.Error("the English plate still shows the French grid")
	}
	// A US account bars nothing, so that edition draws no dot at all.
	if n := strings.Count(en, "<circle"); n != 0 {
		t.Errorf("%d dots on the English plate, where no account is barred", n)
	}
	// The square is an AREA: doubling the weight must not double the side.
	if got := ucitsSide(8) / ucitsSide(2); math.Abs(got-2) > 1e-9 {
		t.Errorf("a 4x weight gives a %.2fx side, expected 2x", got)
	}
}
