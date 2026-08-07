package firebook

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// Every recipe of the plate is a full allocation: the segments must sum to
// 100 % of the portfolio, or the bars lie about what is inside.
func TestSaisonsRowsSumTo100(t *testing.T) {
	for _, r := range saisonRows {
		if total := r.saisonTotal(); math.Abs(total-100) > 1e-9 {
			t.Errorf("%s: the weights sum to %g %%, not 100 %%", r.name, total)
		}
	}
}

// saisonPercents extracts, in order, the percentages of one sentence of the
// article ("25 % actions, 25 % obligations ...", French decimal comma).
var saisonPercents = regexp.MustCompile(`([0-9]+(?:,[0-9]+)?) %`)

// The plate draws no number of its own: every weight is copied from the
// article's own paragraph. This re-reads the embedded markdown, pulls the
// percentages out of each recipe's sentence and compares them one by one with
// the frozen row, in the article's own order. It fails when a weight or a
// whole recipe leaves the text, which is exactly when the plate would start
// contradicting the prose it illustrates.
func TestSaisonsWeightsAreTheArticles(t *testing.T) {
	raw, err := assets.ReadFile("assets/book/fr/portefeuilles-tous-temps.md")
	if err != nil {
		t.Fatal(err)
	}
	article := string(raw)

	cases := []struct {
		row      string
		sentence string
	}{
		{"Browne 4 × 25", "25 % actions, 25 % obligations d'État longues, 25 % or, 25 % cash"},
		{"All-Weather", "30 % actions, 40 % obligations longues, 15 % intermédiaires, 7,5 % or et 7,5 % matières premières"},
		{"Golden Butterfly", "20 % actions larges, 20 % small-cap value, 20 % obligations longues, 20 % courtes, 20 % or"},
		{"Dragon", "24 % actions, 18 % obligations longues, 19 % or, 18 % matières premières et trend, et 21 % de volatilité longue"},
	}
	for _, c := range cases {
		if !strings.Contains(article, c.sentence) {
			t.Errorf("%s: the article no longer states %q", c.row, c.sentence)
			continue
		}
		var want []float64
		for _, m := range saisonPercents.FindAllStringSubmatch(c.sentence, -1) {
			v, err := strconv.ParseFloat(strings.Replace(m[1], ",", ".", 1), 64)
			if err != nil {
				t.Fatal(err)
			}
			want = append(want, v)
		}
		row := saisonRowNamed(t, c.row)
		if len(row.segs) != len(want) {
			t.Errorf("%s: the plate draws %d sleeves, the article lists %d", c.row, len(row.segs), len(want))
			continue
		}
		for i, w := range want {
			if math.Abs(row.segs[i].weight-w) > 1e-9 {
				t.Errorf("%s: sleeve %d is %g %% on the plate, %g %% in the article",
					c.row, i+1, row.segs[i].weight, w)
			}
		}
	}
	// the reference line owes its two numbers to its own name
	ref := saisonRowNamed(t, "60/40")
	if len(ref.segs) != 2 || ref.segs[0].weight != 60 || ref.segs[1].weight != 40 {
		t.Errorf("the 60/40 line must be 60 %% + 40 %%, got %v", ref.segs)
	}
	if !strings.Contains(article, "60/40") {
		t.Error("the article no longer mentions the 60/40 reference")
	}
}

// The season of a sleeve is an editorial reading, and its only authority is
// the article's own Browne sentence. If that sentence goes, the plate's
// colours lose their ground and this test says so.
func TestSaisonsMappingFollowsBrowne(t *testing.T) {
	raw, err := assets.ReadFile("assets/book/fr/portefeuilles-tous-temps.md")
	if err != nil {
		t.Fatal(err)
	}
	const correspondence = "Les actions couvrent la prospérité, les obligations longues la déflation, " +
		"l'or l'inflation et les crises de confiance, le cash la récession"
	if !strings.Contains(string(raw), correspondence) {
		t.Error("the article no longer states the asset-to-season correspondence the plate applies")
	}
	// the four seasons, named on the plate exactly as the grid names them
	for i, want := range []string{"prospérité", "inflation", "déflation", "récession et liquidité"} {
		if saisonDefs[i].name != want {
			t.Errorf("season %d is %q, expected %q", i, saisonDefs[i].name, want)
		}
		if !strings.HasPrefix(saisonDefs[i].label, want+" :") {
			t.Errorf("the legend of season %d must open with its name, got %q", i, saisonDefs[i].label)
		}
	}
	// no sleeve may be counted twice: one segment, one season, by construction
	for _, r := range saisonRows {
		for _, s := range r.segs {
			if s.season < 0 || s.season >= len(saisonDefs) {
				t.Errorf("%s: a sleeve claims season %d, out of the grid", r.name, s.season)
			}
		}
	}
}

// The tally column is the plate's second claim: the four cousins cover the
// grid, the 60/40 does not, and the All-Weather leaves the liquidity season
// empty (the very concentration the article's 2022 paragraph criticises).
func TestSaisonsCoverageTally(t *testing.T) {
	want := map[string]int{
		"60/40": 2, "Browne 4 × 25": 4, "All-Weather": 3,
		"Golden Butterfly": 4, "Dragon": 4,
	}
	if len(want) != len(saisonRows) {
		t.Fatalf("the plate draws %d recipes, the test knows %d", len(saisonRows), len(want))
	}
	for _, r := range saisonRows {
		if got := r.saisonsCovered(); got != want[r.name] {
			t.Errorf("%s covers %d seasons, expected %d", r.name, got, want[r.name])
		}
	}
	held := map[int]bool{}
	for _, s := range saisonRowNamed(t, "60/40").segs {
		held[s.season] = true
	}
	if held[saisonInflation] || held[saisonLiquidite] {
		t.Error("the plate's headline is the 60/40's two holes, inflation and liquidity")
	}
}

// The Dragon's long-volatility pocket cannot be bought in UCITS form, which
// the article says at length. Exactly one segment of the whole plate carries
// the hatch, it is that one, and the plate says why.
func TestSaisonsHatchMarksTheUninvestableSleeve(t *testing.T) {
	var hatched []saisonSeg
	for _, r := range saisonRows {
		for _, s := range r.segs {
			if s.hatch {
				hatched = append(hatched, s)
			}
		}
	}
	if len(hatched) != 1 {
		t.Fatalf("the plate hatches %d sleeves, expected the Dragon's long volatility alone", len(hatched))
	}
	if hatched[0].weight != 21 || hatched[0].season != saisonLiquidite {
		t.Errorf("the hatched sleeve is %v, expected 21 %% in the liquidity season", hatched[0])
	}
	raw, err := assets.ReadFile("assets/book/fr/portefeuilles-tous-temps.md")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "n'existe pas en format UCITS accessible") {
		t.Error("the article no longer says the long-volatility pocket cannot be bought")
	}
	svg := figTousTempsSaisons()
	if !strings.Contains(svg, "hachures") {
		t.Error("the plate must name its hatch")
	}
	if !strings.Contains(svg, saisonHatchWash) {
		t.Error("the hatch is missing from the rendered plate")
	}
}

// The rendered plate must obey the book's own drawing rules and carry every
// number it claims to draw.
func TestSaisonsPlateRenders(t *testing.T) {
	svg := figTousTempsSaisons()
	if !strings.HasPrefix(svg, "<svg viewBox=") {
		t.Fatal("the plate must render an SVG")
	}
	// U+2014 is the em-dash, banned from the book down to its figure labels;
	// it stays escaped here so a repository-wide grep never trips on it.
	for _, banned := range []string{"rgba(", "opacity", "\u2014", "rotate("} {
		if strings.Contains(svg, banned) {
			t.Errorf("the plate uses %q, which the book's figures never do", banned)
		}
	}
	for _, r := range saisonRows {
		if !strings.Contains(svg, ">"+r.name+"<") {
			t.Errorf("the plate does not name %s", r.name)
		}
		for _, s := range r.segs {
			if !strings.Contains(svg, ">"+saisonWeight(s.weight)+"<") {
				t.Errorf("%s: the weight %g %% is not drawn", r.name, s.weight)
			}
		}
	}
	raw, err := assets.ReadFile("assets/book/fr/portefeuilles-tous-temps.md")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "::: figure tous-temps-saisons") {
		t.Error("the article must carry the plate")
	}
}

// saisonRowNamed returns the frozen recipe of that name.
func saisonRowNamed(t *testing.T, name string) saisonRow {
	t.Helper()
	for _, r := range saisonRows {
		if r.name == name {
			return r
		}
	}
	t.Fatalf("the plate no longer draws %q", name)
	return saisonRow{}
}
