package firebook

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// The two bond tables carry no plate and no engine: they are specifications
// written in Markdown. What can still be checked mechanically is that they stay
// consistent with the prose they summarize. The cahier des charges of
// obligations-en-retrait states bands (a long dose of 10 to 20 % of the pocket,
// an indexed share of 25 to 50 %, a fonds euros capped at half); the article's
// worked example must land inside those bands, or the reader who compares the
// grid and the example finds the book contradicting itself.

func bondsArticle(t *testing.T, slug string) string {
	t.Helper()
	raw, err := assets.ReadFile("assets/book/fr/" + slug + ".md")
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

// bondsTable returns the rows of the first pipe table that follows probe, header
// row included, each row split into trimmed cells.
func bondsTable(t *testing.T, article, probe string) [][]string {
	t.Helper()
	_, tail, ok := strings.Cut(article, probe)
	if !ok {
		t.Fatalf("the article no longer says %q, so the table cannot be located", probe)
	}
	var rows [][]string
	started := false
	for _, line := range strings.Split(tail, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") {
			if started {
				break
			}
			continue
		}
		started = true
		body := strings.TrimSuffix(strings.TrimPrefix(line, "|"), "|")
		cells := strings.Split(body, "|")
		for i := range cells {
			cells[i] = strings.TrimSpace(cells[i])
		}
		rows = append(rows, cells)
	}
	if len(rows) < 3 {
		t.Fatalf("no table found after %q", probe)
	}
	// Drop the alignment row.
	return append(rows[:1], rows[2:]...)
}

// bondsPct reads "<n> % en <what>" out of the worked example.
func bondsPct(t *testing.T, article, what string) float64 {
	t.Helper()
	re := regexp.MustCompile(`(\d+(?:,\d+)?) % en ` + regexp.QuoteMeta(what))
	m := re.FindStringSubmatch(article)
	if m == nil {
		t.Fatalf("the worked example no longer sizes %q", what)
	}
	v, err := strconv.ParseFloat(strings.Replace(m[1], ",", ".", 1), 64)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestPocheObligataireExampleObeysItsOwnBands(t *testing.T) {
	article := bondsArticle(t, "obligations-en-retrait")

	fondsEuros := bondsPct(t, article, "fonds euros")
	coeur := bondsPct(t, article, "obligations d'État euro 5-8 ans")
	long := bondsPct(t, article, "État euro 15 ans et plus")
	linkers := bondsPct(t, article, "linkers euro courts")
	or := regexp.MustCompile(`plus (\d+) % d'or`).FindStringSubmatch(article)
	if or == nil {
		t.Fatal("the worked example no longer sizes its gold line")
	}
	gold, _ := strconv.ParseFloat(or[1], 64)

	// The example announces a defensive pocket, and its lines must fill it.
	if !strings.Contains(article, "65 % actions, 35 % défensif") {
		t.Fatal("the worked example no longer states its 65/35 split")
	}
	if got := fondsEuros + coeur + long + linkers + gold; got != 35 {
		t.Errorf("the defensive lines sum to %g %%, the example announces 35 %%", got)
	}

	// Gold is a defence, not a bond: the grid's shares are read on the bond
	// pocket alone, which is what its lead-in says.
	pocket := fondsEuros + coeur + long + linkers
	share := func(v float64) float64 { return 100 * v / pocket }

	if s := share(fondsEuros); s > 50 {
		t.Errorf("fonds euros at %.1f %% of the pocket, the article caps it at half", s)
	}
	if s := share(long); s < 10 || s > 20 {
		t.Errorf("long sleeve at %.1f %% of the pocket, the grid writes 10 to 20 %%", s)
	}
	if s := share(linkers); s < 25 || s > 50 {
		t.Errorf("indexed share at %.1f %%, the grid and [[obligations-indexees]] write 25 to 50 %%", s)
	}
	// "Le solde" only reads as the core if the core is in fact the largest line.
	if coeur <= fondsEuros || coeur <= long || coeur <= linkers {
		t.Errorf("the core is %g %% and is not the largest line, the grid calls it le solde", coeur)
	}
}

func TestPocheObligataireGridIsWellFormed(t *testing.T) {
	article := bondsArticle(t, "obligations-en-retrait")
	rows := bondsTable(t, article, "## Le cahier des charges, en une grille")

	want := []string{"Brique", "Service", "Duration", "Ce qui la casse", "Part"}
	if len(rows[0]) != len(want) {
		t.Fatalf("the grid has %d columns, %d were designed", len(rows[0]), len(want))
	}
	for i, h := range want {
		if rows[0][i] != h {
			t.Errorf("column %d is %q, want %q", i, rows[0][i], h)
		}
	}
	if got := len(rows) - 1; got != 6 {
		t.Errorf("the grid holds %d bricks, the four decisions write 6", got)
	}
	// The duration column is what makes this table specific to this article,
	// so no cell of it may be left blank.
	for _, r := range rows[1:] {
		if len(r) != len(want) {
			t.Errorf("row %q has %d cells", r[0], len(r))
			continue
		}
		if r[2] == "" {
			t.Errorf("brick %q carries no duration", r[0])
		}
	}
	// Every brick must be a brick the prose actually discusses.
	for _, probe := range []string{"fonds euros", "5 à 8", "15 à 20", "linkers", "investment grade", "high yield"} {
		if !strings.Contains(strings.ToLower(article), probe) {
			t.Errorf("the grid draws on %q, the prose no longer says it", probe)
		}
	}
}

func TestEchelleToolboxShowsTheHoleAsAnEmptyRow(t *testing.T) {
	article := bondsArticle(t, "echelle-obligataire")
	rows := bondsTable(t, article, "## La pratique française")

	want := []string{"Véhicule", "Horizon utile", "Indexé", "Frais", "Accès"}
	if len(rows[0]) != len(want) {
		t.Fatalf("the toolbox has %d columns, %d were designed", len(rows[0]), len(want))
	}
	for i, h := range want {
		if rows[0][i] != h {
			t.Errorf("column %d is %q, want %q", i, rows[0][i], h)
		}
	}
	if got := len(rows) - 1; got != 7 {
		t.Errorf("the toolbox holds %d vehicles, 6 plus the missing one were designed", got)
	}

	var holes int
	for _, r := range rows[1:] {
		if len(r) != len(want) {
			t.Fatalf("row %q has %d cells", r[0], len(r))
		}
		blank := r[1] == "" || r[3] == ""
		if !blank {
			continue
		}
		holes++
		if r[1] != "" || r[3] != "" {
			t.Errorf("row %q is half empty, the hole must read as blank horizon AND blank fees", r[0])
		}
		if r[2] != "Oui" {
			t.Errorf("the missing vehicle %q is the indexed one, its column must read Oui", r[0])
		}
		if r[4] != "N'existe pas" {
			t.Errorf("row %q must say plainly that it does not exist", r[0])
		}
	}
	if holes != 1 {
		t.Errorf("%d empty rows, the offer has exactly one hole to show", holes)
	}
	// The lead-in must send the reader to that empty row, otherwise the blank
	// cells read as an editing accident.
	if !strings.Contains(article, "la ligne vide dit le reste") {
		t.Error("the lead-in no longer points at the empty row")
	}
}

// Both tables are read on a six-inch screen: long header cells wreck the
// column widths, and an em-dash is banned book-wide.
func TestBondTablesStayNarrow(t *testing.T) {
	for _, tc := range []struct{ slug, probe string }{
		{"obligations-en-retrait", "## Le cahier des charges, en une grille"},
		{"echelle-obligataire", "## La pratique française"},
	} {
		article := bondsArticle(t, tc.slug)
		rows := bondsTable(t, article, tc.probe)
		for _, h := range rows[0] {
			if len([]rune(h)) > 16 {
				t.Errorf("%s: header %q is %d characters, keep it short", tc.slug, h, len([]rune(h)))
			}
		}
		for _, r := range rows {
			for _, c := range r {
				if len([]rune(c)) > 36 {
					t.Errorf("%s: cell %q is too wide for a narrow screen", tc.slug, c)
				}
				if strings.Contains(c, "—") {
					t.Errorf("%s: cell %q contains an em-dash", tc.slug, c)
				}
			}
		}
	}
}
