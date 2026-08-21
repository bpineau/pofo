package firebook

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
)

// defensesDay parses one of the plate's ISO bounds.
func defensesDay(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		t.Fatalf("bound %q: %v", s, err)
	}
	return d
}

// defensesAsOf returns the last quote of a series at or before d, which is how
// an episode bound lands on a calendar date the market did not trade.
func defensesAsOf(t *testing.T, s *marketdata.Series, d time.Time) float64 {
	t.Helper()
	v := 0.0
	for _, p := range s.Points {
		if p.Date.After(d) {
			break
		}
		v = p.Close
	}
	if v == 0 {
		t.Fatalf("no quote at or before %s", d.Format("2006-01-02"))
	}
	return v
}

// defensesCPI reads the US CPI the deflator needs, from the snapshot
// pkg/marketdata embeds: monthly anchors dated the first of the month, served
// without any network.
func defensesCPI(t *testing.T) map[string]float64 {
	t.Helper()
	s, err := marketdata.NewClient("").Fetch(t.Context(), "^CPI-US", time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]float64{}
	for _, p := range s.Points {
		if p.Date.Day() == 1 {
			out[p.Date.Format("2006-01")] = p.Close
		}
	}
	return out
}

// defensesRates reads the 3-month bill rate, annualized percent per month.
func defensesRates(t *testing.T) map[string]float64 {
	t.Helper()
	s, ok, err := marketdata.ReadSimdataFS(datasets.Refdata(), "TBILL-3M")
	if err != nil || !ok {
		t.Fatalf("read TBILL-3M: ok=%v err=%v", ok, err)
	}
	out := map[string]float64{}
	for _, p := range s.Points {
		out[p.Date.Format("2006-01")] = p.Close
	}
	return out
}

// defensesCashGrowth compounds the bill rate over an episode. The bill series
// is a RATE, not an index: each month of the window earns its own annualized
// rate over the days the window actually spends in it, which is what makes the
// covid crash's five weeks comparable with two calendar years.
func defensesCashGrowth(t *testing.T, rates map[string]float64, from, to time.Time) float64 {
	t.Helper()
	grow := 1.0
	for cur := from; cur.Before(to); {
		next := time.Date(cur.Year(), cur.Month()+1, 1, 0, 0, 0, 0, time.UTC)
		if next.After(to) {
			next = to
		}
		r, ok := rates[cur.Format("2006-01")]
		if !ok {
			t.Fatalf("no bill rate for %s", cur.Format("2006-01"))
		}
		grow *= 1 + r/100*next.Sub(cur).Hours()/24/365
		cur = next
	}
	return grow
}

// defensesMeasure rebuilds the plate's whole grid from the bundled datasets:
// for every candidate that has a series reaching the episode, its total return
// over the episode's dates, deflated by the US CPI of the months the two bounds
// fall in, in percent.
func defensesMeasure(t *testing.T) [6][4]float64 {
	t.Helper()
	frozenAgainstData(t)
	cpi := defensesCPI(t)
	rates := defensesRates(t)

	var out [6][4]float64
	for i, row := range defensesRows {
		var s *marketdata.Series
		if row.id != "TBILL-3M" {
			fsys := datasets.Simdata()
			if row.ref {
				fsys = datasets.Refdata()
			}
			got, ok, err := marketdata.ReadSimdataFS(fsys, row.id)
			if err != nil || !ok {
				t.Fatalf("read %s: ok=%v err=%v", row.id, ok, err)
			}
			s = got
		}
		for j, ep := range defensesEpisodes {
			if !defensesHas(i, j) {
				continue
			}
			from, to := defensesDay(t, ep.from), defensesDay(t, ep.to)
			inflation := cpi[to.Format("2006-01")] / cpi[from.Format("2006-01")]
			if inflation == 0 {
				t.Fatalf("no CPI for %s..%s", ep.from, ep.to)
			}
			grow := 0.0
			if s == nil {
				grow = defensesCashGrowth(t, rates, from, to)
			} else {
				grow = defensesAsOf(t, s, to) / defensesAsOf(t, s, from)
			}
			out[i][j] = (grow/inflation - 1) * 100
		}
	}
	return out
}

// The plate's twenty-three cells are frozen literals, so the book's figures
// stay pure functions with no data dependency at render time. This rebuilds
// every one of them from the bundled series and fails the moment the plate and
// the data disagree, which is also what happens when those series are
// regenerated.
func TestDefensesGridMatchesTheData(t *testing.T) {
	got := defensesMeasure(t)
	for i, row := range defensesRows {
		for j, ep := range defensesEpisodes {
			if !defensesHas(i, j) {
				continue
			}
			// half a point, the resolution the plate prints its cells at
			if math.Abs(got[i][j]-defensesGrid[i][j]) > 0.5 {
				t.Errorf("%s in %s: the data says %.1f %%, the plate draws %.1f %%",
					row.name, ep.label, got[i][j], defensesGrid[i][j])
			}
		}
	}
}

// The legend's euro correction, recomputed from the bundled exchange rate.
func TestDefensesGoldInEurosMatchesTheData(t *testing.T) {
	frozenAgainstData(t)
	fx, err := marketdata.NewClient("").Fetch(t.Context(), "EURUSD=X", time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	from := defensesAsOf(t, fx, defensesDay(t, defensesEpisodes[3].from))
	to := defensesAsOf(t, fx, defensesDay(t, defensesEpisodes[3].to))
	got := (from/to - 1) * 100
	if math.Abs(got-defensesGoldEUR) > 0.05 {
		t.Errorf("the dollar's 2022 rise gave a euro holder %.2f points, the plate says %.2f",
			got, defensesGoldEUR)
	}
}

// The framed cells must keep saying what the plate exists to say: several
// different defenders share the four episodes (three of them, long bonds
// serving both growth crashes exactly as the article's review claims), and none
// of them wins everywhere. That alternation is a property of the data, not of
// the drawing.
func TestDefensesWinnersAreSeveralAndNeverOne(t *testing.T) {
	seen := map[int]int{}
	for j, ep := range defensesEpisodes {
		best := defensesBest(j)
		if best <= 0 {
			t.Fatalf("%s has no framed defender", ep.label)
		}
		seen[best]++
	}
	if len(seen) < 3 {
		t.Errorf("only %d different defenders win the four episodes: the plate's argument is gone", len(seen))
	}
	for row, n := range seen {
		if n == len(defensesEpisodes) {
			t.Errorf("%q wins every episode: no asset defends against everything", defensesRows[row].name)
		}
	}
	// The two growth crashes are the pair the article gives long bonds; the
	// inflationary regime and the modern rate shock must go to someone else.
	if defensesBest(1) != defensesBest(2) {
		t.Errorf("2008 and the covid crash are no longer won by the same defender (%q, %q)",
			defensesRows[defensesBest(1)].name, defensesRows[defensesBest(2)].name)
	}
	if defensesBest(0) == defensesBest(1) || defensesBest(3) == defensesBest(1) {
		t.Error("the inflationary episodes are now won by the deflation defender")
	}
}

// Every candidate the plate frames as a winner must also collapse somewhere
// else in its own row: a defender that never fails would contradict the
// article's central claim.
func TestDefensesWinnersAlsoFailSomewhere(t *testing.T) {
	for j := range defensesEpisodes {
		best := defensesBest(j)
		worst := math.Inf(1)
		for k := range defensesEpisodes {
			if defensesHas(best, k) && defensesGrid[best][k] < worst {
				worst = defensesGrid[best][k]
			}
		}
		if worst >= 0 {
			t.Errorf("%q never loses in the grid (worst cell %.1f %%): no asset defends against everything",
				defensesRows[best].name, worst)
		}
	}
}

// The plate's readings against the article that carries it: the verdicts its
// review states for the episodes the grid measures.
func TestDefensesAgreesWithItsArticle(t *testing.T) {
	art := bookArticle(t, "actifs-defensifs")
	for _, want := range []string{
		"::: figure defenses-bulletin",
		"aucun** actif ne défend contre tous",
		"2008, +25 à +30 % ; 2020, +20 %",
		"2022, −30 à −40 %",
		"2022, +27 %",
		"comme le krach trop rapide de 2020",
		"2022, honorable en euros",
	} {
		if !strings.Contains(art, want) {
			t.Errorf("the article no longer says %q", want)
		}
	}
	// The article's verdicts, read on the grid: long bonds are the deflation
	// insurance (2008 and the covid crash), trend the long-regime one (2022,
	// and a failure in the V), gold the confidence one (the inflationary
	// seventies), cash the thing inflation eats.
	long, trend, gold, cash := 3, 5, 4, 1
	if defensesGrid[long][1] <= 0 || defensesGrid[long][2] <= 0 {
		t.Errorf("long bonds no longer rise in 2008 (%.1f %%) and in the covid crash (%.1f %%)",
			defensesGrid[long][1], defensesGrid[long][2])
	}
	if defensesGrid[long][3] >= -20 {
		t.Errorf("long bonds lose %.1f %% in 2022, the article says 30 to 40", defensesGrid[long][3])
	}
	if defensesGrid[trend][3] <= 0 || defensesGrid[trend][2] >= 0 {
		t.Errorf("trend no longer wins 2022 (%.1f %%) while failing the V of 2020 (%.1f %%)",
			defensesGrid[trend][3], defensesGrid[trend][2])
	}
	if defensesGrid[gold][0] <= 0 {
		t.Errorf("gold no longer defends the seventies (%.1f %%)", defensesGrid[gold][0])
	}
	if defensesGrid[cash][0] >= 0 || defensesGrid[cash][3] >= 0 {
		t.Errorf("cash no longer loses in real terms to inflation (%.1f %% and %.1f %%)",
			defensesGrid[cash][0], defensesGrid[cash][3])
	}
}

// The rendered plate obeys the book's drawing rules and carries every reading
// it claims to draw.
func TestDefensesPlateRenders(t *testing.T) {
	svg := FigureSVG("defenses-bulletin")
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
	for _, row := range defensesRows {
		if !strings.Contains(svg, ">"+row.name+"<") {
			t.Errorf("the plate does not name %q", row.name)
		}
	}
	for _, ep := range defensesEpisodes {
		if !strings.Contains(svg, ">"+ep.label+"<") {
			t.Errorf("the plate does not head the column %q", ep.label)
		}
		if !strings.Contains(svg, ep.dates) {
			t.Errorf("the plate does not date the column %q", ep.label)
		}
	}
	for i := range defensesRows {
		for j := range defensesEpisodes {
			if !defensesHas(i, j) {
				continue
			}
			want := ">" + defensesCell(defensesGrid[i][j]) + "<"
			if !strings.Contains(svg, want) {
				t.Errorf("the plate does not print the cell %q", want)
			}
		}
	}
	// The one cell without a series says so, and says it once.
	if n := strings.Count(svg, ">pas de série<"); n != 1 {
		t.Errorf("%d cells hatched, expected the single one the data misses", n)
	}
	// One frame per column, and no cell of the equities row framed.
	if n := strings.Count(svg, `class="win"`); n != len(defensesEpisodes) {
		t.Errorf("%d framed cells, expected one per episode", n)
	}
	for _, want := range []string{
		"déflaté par l'IPC américain",
		"Le BTOP50 commence fin 1986 (case vide, jamais extrapolée)",
		"L'or perd 6 % en dollars réels en 2022, mais la hausse du dollar rend 6,2 points",
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("the plate does not carry the note %q", want)
		}
	}
}

// defensesCellFormats keeps the cell format honest: a signed percent with the
// book's typographic minus and no decimal noise.
func TestDefensesCellFormat(t *testing.T) {
	for _, c := range []struct {
		v    float64
		want string
	}{
		{-37.4, figMinus + "37 %"},
		{0.4, "0 %"},
		{21.6, "+22 %"},
	} {
		if got := defensesCell(c.v); got != c.want {
			t.Errorf("defensesCell(%.1f) = %q, want %q", c.v, got, c.want)
		}
	}
}
