package firebook

import (
	"math"
	"strings"
	"testing"
)

// Two reference tables of the portfolio part carry no drawing and no engine
// output, so the guards pin them against the prose they replace: the audit
// grid of primes-de-risque.md, and the instrument sheet of
// levier-et-marges.md. The one number both articles share with an engine-free
// but closed formula, the volatility drag of a daily-leveraged ETF, is
// recomputed here from the drag article's own volatility.

// markdownTable returns the rows of the first pipe table whose header starts
// with the given first cell. The header is row 0; the separator is dropped.
func markdownTable(t *testing.T, article, firstHeaderCell string) [][]string {
	t.Helper()
	var rows [][]string
	for _, line := range strings.Split(article, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") {
			if len(rows) > 0 {
				break
			}
			continue
		}
		cells := strings.Split(strings.Trim(line, "|"), "|")
		for i, c := range cells {
			cells[i] = strings.TrimSpace(c)
		}
		if len(rows) == 0 && cells[0] != firstHeaderCell {
			continue
		}
		if strings.HasPrefix(cells[0], "-") {
			continue // separator
		}
		rows = append(rows, cells)
	}
	if len(rows) == 0 {
		t.Fatalf("no table starting with header cell %q", firstHeaderCell)
	}
	return rows
}

// The audit grid asks the article's three questions, one per column, over the
// eight lines it claims to cover. It must stay level-free: the premium levels
// belong to the primes-echelle plate, the table carries the payer and the
// reason to last.
func TestPrimesAuditTable(t *testing.T) {
	article := bookArticle(t, "primes-de-risque")
	rows := markdownTable(t, article, "La ligne")

	header := rows[0]
	want := []string{"La ligne", "Quelle prime", "Qui la paie", "Pourquoi elle tient"}
	if len(header) != len(want) {
		t.Fatalf("the grid has %d columns, the article asks 3 questions per line plus the line itself", len(header))
	}
	for i, w := range want {
		if header[i] != w {
			t.Errorf("column %d is %q, want %q", i, header[i], w)
		}
	}

	body := rows[1:]
	if len(body) != 8 {
		t.Errorf("the grid holds %d lines, the lead-in announces eight", len(body))
	}
	for _, r := range body {
		for i, c := range r {
			if c == "" {
				t.Errorf("line %q leaves column %d empty; a line without an answer is the article's own red flag", r[0], i)
			}
			if strings.Contains(c, "%") {
				t.Errorf("line %q quotes a level in column %d; levels belong to the primes-echelle plate", r[0], i)
			}
		}
	}

	// The premium vocabulary of the table is the article's own.
	for _, prime := range []string{"prime actions", "prime de terme", "prime d'illiquidité", "comportementale"} {
		if !strings.Contains(strings.ToLower(article), prime) {
			t.Errorf("the table names %q but the prose never explains it", prime)
		}
	}
	// The plate stays, and the table must not duplicate its job.
	if !strings.Contains(article, "::: figure primes-echelle") {
		t.Error("the article must still carry the levels plate")
	}
}

// The instrument sheet sorts on one column, and the lead-in says so.
func TestLevierInstrumentsTable(t *testing.T) {
	article := bookArticle(t, "levier-et-marges")
	rows := markdownTable(t, article, "Instrument")

	header := rows[0]
	want := []string{"Instrument", "Appel de marge", "Coût", "Durée", "Verdict"}
	if len(header) != len(want) {
		t.Fatalf("the sheet has %d columns, want %d", len(header), len(want))
	}
	for i, w := range want {
		if header[i] != w {
			t.Errorf("column %d is %q, want %q", i, header[i], w)
		}
	}

	body := rows[1:]
	if len(body) != 7 {
		t.Errorf("the sheet holds %d instruments, the lead-in announces seven", len(body))
	}
	for _, r := range body {
		for i, c := range r {
			if c == "" {
				t.Errorf("instrument %q leaves column %d empty", r[0], i)
			}
		}
		// The sorting criterion is answered on every row, yes or no.
		if m := r[1]; !strings.HasPrefix(m, "Oui") && !strings.HasPrefix(m, "Aucun") {
			t.Errorf("instrument %q answers the margin-call column with %q, want a plain Oui or Aucun", r[0], m)
		}
	}

	// The two instruments the chapter forbids must read as forbidden, and the
	// two ponts it allows must not.
	verdict := map[string]string{}
	for _, r := range body {
		verdict[r[0]] = r[4]
	}
	for _, name := range []string{"Marge de courtage", "ETF à levier quotidien", "Avance d'assurance-vie", "Fonds 90/60"} {
		if _, ok := verdict[name]; !ok {
			t.Errorf("the sheet does not list %q", name)
		}
	}
	if !strings.Contains(verdict["Marge de courtage"], "Interdiction") {
		t.Errorf("the brokerage margin must stay the chapter's absolute ban, got %q", verdict["Marge de courtage"])
	}
	if !strings.HasPrefix(verdict["ETF à levier quotidien"], "Non") {
		t.Errorf("the daily-leveraged ETF must stay a no, got %q", verdict["ETF à levier quotidien"])
	}
}

// The one number the sheet's prose keeps is the drag of a daily x2: the drag
// article reads a 30 % volatility on that line, and 7 % arithmetic minus
// sigma^2/2 leaves the 2,5 % compounded both articles quote.
func TestLevierDragNumberMatchesTheDragArticle(t *testing.T) {
	drag := bookArticle(t, "rendements-arithmetiques-geometriques")
	if !strings.Contains(drag, "| Actions à levier ×2 | ~30 % | ~4,5 % | ~2,5 %") {
		t.Fatal("the drag article no longer reads 30 % volatility and 2,5 % compounded on the leveraged line")
	}
	const arithmetic, sigma = 0.07, 0.30
	geometric := arithmetic - sigma*sigma/2
	if math.Abs(geometric-0.025) > 1e-9 {
		t.Fatalf("7 %% arithmetic at 30 %% volatility leaves %.4f, not 2,5 %%", geometric)
	}
	levier := bookArticle(t, "levier-et-marges")
	if !strings.Contains(levier, "7 % de moyenne arithmétique ne laissent que 2,5 % composé") {
		t.Error("levier-et-marges must keep the drag figure it borrows from the drag article")
	}
}
