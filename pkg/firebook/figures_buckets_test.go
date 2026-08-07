package firebook

import (
	"fmt"
	"math"
	"strings"
	"testing"
)

// The plate claims an arithmetic identity, so the identity is what the guard
// checks: the recipe stated in the article's prose (two years of cash, six of
// bonds, the rest in equities) must produce the three splits the article and
// the plate print, and every column must be a whole portfolio.
func TestBucketsSplitMatchesTheArticle(t *testing.T) {
	for _, c := range []struct {
		rate                float64
		cash, bonds, equity float64
		readout             string
	}{
		{2.5, 5, 15, 80, "5 / 15 / 80"},
		{bucketsRateRef, 7, 21, 72, "7 / 21 / 72"},
		{5, 10, 30, 60, "10 / 30 / 60"},
	} {
		cash, bonds, equity := bucketsSplit(c.rate)
		for _, got := range []struct {
			name     string
			got, exp float64
		}{
			{"cash", cash, c.cash},
			{"obligations", bonds, c.bonds},
			{"actions", equity, c.equity},
		} {
			if math.Abs(got.got-got.exp) > 1e-9 {
				t.Errorf("%.1f %%: %s = %.2f, expected %.2f", c.rate, got.name, got.got, got.exp)
			}
		}
		if sum := cash + bonds + equity; math.Abs(sum-100) > 1e-9 {
			t.Errorf("%.1f %%: the column sums to %.4f, not 100", c.rate, sum)
		}
		if got := bucketsReadout(c.rate); got != c.readout {
			t.Errorf("%.1f %%: readout %q, expected %q", c.rate, got, c.readout)
		}
	}
}

// The equity sleeve must be a residual that shrinks as the rate rises: it is
// the single reading the plate exists for.
func TestBucketsEquityIsAShrinkingResidual(t *testing.T) {
	prev := math.Inf(1)
	for r := bucketsRateMin; r <= bucketsRateMax+1e-9; r += 0.1 {
		_, _, equity := bucketsSplit(r)
		if equity >= prev {
			t.Fatalf("at %.1f %% the equity share (%.1f) did not fall below the previous one (%.1f)", r, equity, prev)
		}
		prev = equity
	}
}

// The article's key block prices the recipe at ~7/21/72, and the count the
// plate illustrates names the 3,5 % rate it belongs to. If either moves, the
// plate's reference mark is wrong and must move with it.
func TestBucketsArticleCarriesThePlate(t *testing.T) {
	raw, err := assets.ReadFile("assets/book/fr/strategie-buckets.md")
	if err != nil {
		t.Fatal(err)
	}
	art := string(raw)
	for _, want := range []string{
		"2 ans de cash + 6 ans d'obligations + le reste en actions",
		"~7/21/72",
		"taux de retrait de 3,5 %",
		"::: figure buckets-allocation",
	} {
		if !strings.Contains(art, want) {
			t.Errorf("the article no longer says %q", want)
		}
	}
}

// The rendered plate obeys the book's drawing rules and carries every number
// and label it claims to draw.
func TestBucketsPlateRenders(t *testing.T) {
	svg := figBucketsAllocation()
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
	for _, want := range []string{
		bucketsReadout(bucketsRateMin),
		bucketsReadout(bucketsRateRef),
		bucketsReadout(bucketsRateMax),
		">actions<", ">obligations<", ">cash<",
		fmt.Sprintf(">%s<", strings.Replace(fmt.Sprintf("%.1f", bucketsRateRef), ".", ",", 1)),
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("the plate does not draw %q", want)
		}
	}
}
