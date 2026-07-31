package firebook

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
)

var reRefsWikiLbl = regexp.MustCompile(`\[\[([^\]|]+)\|([^\]]+)\]\]`)

// Guards for the two numeric tables of the "References" part.
//
// The memo table of lexique.md is a book-wide consistency check disguised as a
// reference page: it may only restate numbers already written elsewhere. The
// guard therefore reads the table back out of the markdown and, for every row,
// checks that the cited chapter still states that number. A chapter that moves
// one of its anchors now breaks the build instead of silently contradicting the
// memo.
//
// The three-samples table of anarkulova-cederburg.md carries two rates that
// this repository computes, and the guard pins them on figures_pays.go, the
// plate the same article shows just above.

// tableRows returns the pipe-table rows that follow heading, header and
// separator excluded, each row split into trimmed cells.
func refsTableRows(t *testing.T, article, heading string) [][]string {
	t.Helper()
	i := strings.Index(article, heading)
	if i < 0 {
		t.Fatalf("heading %q is gone from the article", heading)
	}
	var rows [][]string
	seen := false
	for _, line := range strings.Split(article[i:], "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") {
			if seen {
				break
			}
			continue
		}
		seen = true
		body := strings.TrimSuffix(strings.TrimPrefix(line, "|"), "|")
		// a labelled wiki-link carries its own pipe, which is not a cell border
		body = reRefsWikiLbl.ReplaceAllString(body, "[[$1\x00$2]]")
		cells := strings.Split(body, "|")
		for j := range cells {
			cells[j] = strings.TrimSpace(strings.ReplaceAll(cells[j], "\x00", "|"))
		}
		rows = append(rows, cells)
	}
	if len(rows) < 3 {
		t.Fatalf("no table under %q", heading)
	}
	return rows[2:]
}

// memoAnchor is one line of the lexique's numeric memo: the repere as the table
// writes it, the glossary entry it belongs to, the short chapter name of the
// last column, the slug that chapter has, and the spellings that chapter is
// allowed to use for the same value (a chapter may write "0,5 à 1,5 point"
// where the memo writes "0,5-1,5 point").
type memoAnchor struct {
	repere   string
	terme    string
	chapitre string
	slug     string
	sources  []string
}

var memoAnchors = []memoAnchor{
	{"~4 % sur 30 ans", "SAFEMAX", "Trinity", "etude-trinity", []string{"~4,15 % pour 30 ans", "~4,15 % sur 30 ans"}},
	{"3,25-3,5 %", "SAFEMAX", "ERN", "serie-ern", []string{"3,25-3,5 %"}},
	{"~17 %", "Broad sample", "Échantillon mondial", "anarkulova-cederburg", []string{"environ 17 % des cas", "~17 % du temps"}},
	{"0,5-1,5 point", "Cascade", "Les deux moyennes", "rendements-arithmetiques-geometriques", []string{"0,5 à 1,5 point"}},
	{"50-80 % d'actions", "Allocation d'actifs", "Allocation", "allocation-actions-obligations", []string{"entre environ 50 et 80 % d'actions"}},
	{"4-6 points", "Prime de risque", "Primes de risque", "primes-de-risque", []string{"4 à 6 points", "4-6 points"}},
	{"1-2 points", "Prime de risque", "Primes de risque", "primes-de-risque", []string{"1 à 2 points"}},
	{"0,2-0,5 point/an", "Rebalancing premium", "Diversification", "pourquoi-la-diversification-marche", []string{"0,2 à 0,5 point par an"}},
	{"2-4 points", "Prime de variance", "Long volatility", "long-volatility", []string{"2 à 4 points", "2-4 points"}},
	{"2-4 % réels bruts", "Managed futures", "Trend", "managed-futures", []string{"2-4 % réels bruts"}},
	{"18-36 mois", "Buffer", "Matelas", "cash-buffer", []string{"18-36 mois"}},
	{"±20 %, coupes de 10 %", "Guardrails", "Guyton-Klinger", "guyton-klinger", []string{"corridor ±20 %, coupes 10 %"}},
	{"10-20 %", "Sourire", "Dépenses", "depenses-en-retraite", []string{"de 10-20 %"}},
	{"~6,5 % au-delà de ~24 k€", "CSM", "Taxe PUMa", "taxe-puma", []string{"~6,5 % des revenus du capital au-delà d'une franchise (0,5 PASS ≈ 24 k€)"}},
	{"~7,2 k€ puis ~9,6 k€", "Quadruplé", "Barista", "retour-au-travail", []string{"4 trimestres validés par an dès ~7 200 €"}},
}

// The memo restates the book, line by line, and never invents a value.
func TestLexiqueMemoRestatesTheBook(t *testing.T) {
	lexique := bookArticle(t, "lexique")
	rows := refsTableRows(t, lexique, "## Le mémo chiffré")
	if len(rows) != len(memoAnchors) {
		t.Fatalf("the memo has %d rows, the guard knows %d", len(rows), len(memoAnchors))
	}
	for i, a := range memoAnchors {
		row := rows[i]
		if len(row) != 4 {
			t.Fatalf("row %d has %d cells, want 4", i, len(row))
		}
		if row[0] != a.repere {
			t.Errorf("row %d starts with %q, want %q", i, row[0], a.repere)
		}
		if row[2] != a.terme {
			t.Errorf("row %d names the term %q, want %q", i, row[2], a.terme)
		}
		if row[3] != a.chapitre {
			t.Errorf("row %d names the chapter %q, want %q", i, row[3], a.chapitre)
		}
		// the term is a real entry of this very glossary
		if !strings.Contains(lexique, "\n**"+a.terme) {
			t.Errorf("row %d names %q, which is not an entry of the lexique", i, a.terme)
		}
		// and the cited chapter still states the number
		chapter := bookArticle(t, a.slug)
		ok := false
		for _, s := range a.sources {
			if strings.Contains(chapter, s) {
				ok = true
				break
			}
		}
		if !ok {
			t.Errorf("%s no longer states %q (memo row %q); the memo is stale or the chapter moved",
				a.slug, a.sources, a.repere)
		}
	}
}

// The three-samples table agrees with the plate the same article carries: the
// equal-weight world basket and the American bar of figSafemaxPays.
func TestAnarkulovaSamplesTableAgreesWithThePlate(t *testing.T) {
	article := bookArticle(t, "anarkulova-cederburg")
	rows := refsTableRows(t, article, "Trois échantillons circulent")
	if len(rows) != 3 {
		t.Fatalf("the table has %d rows, want 3", len(rows))
	}

	world := fmt.Sprintf("%s %%", frNum(safemaxWorld, 1))
	if got := rows[1][4]; !strings.Contains(got, world) {
		t.Errorf("the broad-sample row says %q, the plate's world basket is %s", got, world)
	}
	var usa float64
	for _, c := range safemaxCountries {
		if c.ISO == "USA" {
			usa = c.Rate
		}
	}
	if usa == 0 {
		t.Fatal("the plate lost its American bar")
	}
	// the paragraph under the table opposes the panel's United States to the
	// 4,15 % of the Bengen tradition, and both numbers must stay true
	if s := frNum(usa, 2) + " %"; !strings.Contains(article, s) {
		t.Errorf("the article no longer states the panel's American rate %s", s)
	}
	if !strings.Contains(rows[2][4], "4,15 %") || !strings.Contains(article, "4,15 %") {
		t.Error("the American-history row must carry the 4,15 % SAFEMAX")
	}
	// the paper's two published figures, as the article states them
	for _, s := range []string{"~17 % d'échec à 4 %", "2,26 % à 5 % d'échec"} {
		if !strings.Contains(rows[0][4], s) {
			t.Errorf("the paper's row lost %q", s)
		}
	}
	if !strings.Contains(article, "2,26 %") || !strings.Contains(article, "environ 17 % des cas") {
		t.Error("the prose no longer backs the paper's row")
	}
}
