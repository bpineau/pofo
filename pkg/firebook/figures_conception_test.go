package firebook

import (
	"strings"
	"testing"
)

// The design map is a diagram, not a computation: there is no engine to
// recompute it against, so the guard pins its STRUCTURE against the article it
// illustrates. Every node, every link and the whole shape of the map must stay
// readable in concevoir-un-portefeuille.md.

func conceptionArticle(t *testing.T) string {
	t.Helper()
	raw, err := assets.ReadFile("assets/book/fr/concevoir-un-portefeuille.md")
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

// Every label of both columns is one of the article's own lines, and the
// article carries the plate.
func TestRisquesBriquesNodesComeFromTheArticle(t *testing.T) {
	article := conceptionArticle(t)
	if !strings.Contains(article, "::: figure risques-briques") {
		t.Error("the article must carry the plate")
	}
	all := append(append(append([]cbNode{}, cbRisks...), cbBricks...), cbOffPortfolio...)
	seen := map[string]bool{}
	for _, n := range all {
		if seen[n.key] {
			t.Errorf("duplicate node key %q", n.key)
		}
		seen[n.key] = true
		if n.probe == "" || !strings.Contains(article, n.probe) {
			t.Errorf("node %q is drawn but the article does not say %q", n.label, n.probe)
		}
	}
	// The article's question 1 names five market risks, and the map draws those
	// five and no others (the sixth left-hand entry is the engine line of
	// question 5, which is a need and not a risk).
	var market, engine int
	for _, n := range cbRisks {
		switch n.kind {
		case cbDefended, cbAccepted:
			market++
		case cbEngine:
			engine++
		default:
			t.Errorf("risk %q has an unexpected kind", n.label)
		}
	}
	if market != 5 || engine != 1 {
		t.Errorf("left column holds %d market risks and %d engine lines, the article names 5 and 1", market, engine)
	}
	// Question 5 writes nine bricks; the map must draw exactly that list.
	if len(cbBricks) != 9 {
		t.Errorf("the map draws %d bricks, question 5 names 9", len(cbBricks))
	}
}

// Each link quotes the fragment of the article it is read from, which is also
// what places it on the three-rung intensity scale.
func TestRisquesBriquesLinksQuoteTheArticle(t *testing.T) {
	article := conceptionArticle(t)
	riskKeys, brickKeys := map[string]bool{}, map[string]bool{}
	for _, n := range cbRisks {
		riskKeys[n.key] = true
	}
	for _, n := range cbBricks {
		brickKeys[n.key] = true
	}
	for _, l := range cbLinks {
		if !riskKeys[l.risk] || !brickKeys[l.brick] {
			t.Errorf("link %s -> %s points at a node that is not drawn", l.risk, l.brick)
		}
		if !strings.Contains(article, l.quote) {
			t.Errorf("link %s -> %s claims the article says %q, it does not", l.risk, l.brick, l.quote)
		}
		if l.w < cbPartial || l.w > cbMain {
			t.Errorf("link %s -> %s sits outside the three-rung scale", l.risk, l.brick)
		}
	}
	// The scale must stay ordered and the legend must show every rung in use.
	if !(cbPartial.stroke() < cbSecondary.stroke() && cbSecondary.stroke() < cbMain.stroke()) {
		t.Error("the intensity scale is not ordered by width")
	}
	used := map[cbWeight]bool{}
	for _, l := range cbLinks {
		used[l.w] = true
	}
	for _, w := range []cbWeight{cbPartial, cbSecondary, cbMain} {
		if !used[w] {
			t.Errorf("rung %d is in the legend but no link uses it", w)
		}
	}
}

// The shape of the map is the chapter's point: no brick without an answer, one
// risk accepted out loud, and the out-of-portfolio risks kept out of it.
func TestRisquesBriquesShapeHoldsTheThesis(t *testing.T) {
	article := conceptionArticle(t)
	in, out := map[string]int{}, map[string]int{}
	for _, l := range cbLinks {
		in[l.brick]++
		out[l.risk]++
	}
	for _, n := range cbBricks {
		if in[n.key] == 0 {
			t.Errorf("brick %q has no incoming link: pas de réponse, pas de diversification", n.label)
		}
	}
	var accepted []cbNode
	for _, n := range cbRisks {
		switch n.kind {
		case cbAccepted:
			accepted = append(accepted, n)
			if out[n.key] != 0 {
				t.Errorf("risk %q is marked accepted but defends itself", n.label)
			}
		default:
			if out[n.key] == 0 {
				t.Errorf("risk %q has no outgoing link and is not marked accepted", n.label)
			}
		}
	}
	if len(accepted) != 1 {
		t.Fatalf("%d accepted risks drawn, the article accepts exactly one market risk", len(accepted))
	}
	// A risk drawn as accepted must be accepted IN WRITING in the article, which
	// is the chapter's own rule ("l'accepter par oubli est la façon dont meurent
	// les plans").
	for _, want := range []string{
		"n'a volontairement aucune ligne dans cette carte, le marché baissier prolongé",
		"celui qu'on est payé pour porter",
	} {
		if !strings.Contains(article, want) {
			t.Errorf("the article must state the acceptance: %q is missing", want)
		}
	}
	// The out-of-portfolio risks belong to the second paragraph of question 1
	// and must carry no link at all.
	_, tail, ok := strings.Cut(article, "les risques que le portefeuille ne traite")
	if !ok {
		t.Fatal("question 1 no longer names the risks the portfolio does not treat")
	}
	para, _, _ := strings.Cut(tail, "\n\n")
	for _, n := range cbOffPortfolio {
		if n.kind != cbOutside {
			t.Errorf("%q is drawn in the out-of-portfolio band but not marked as such", n.label)
		}
		if in[n.key]+out[n.key] != 0 {
			t.Errorf("%q is out of the portfolio and must carry no link", n.label)
		}
		if !strings.Contains(para, n.probe) {
			t.Errorf("%q is drawn outside the portfolio but question 1 does not list it there", n.label)
		}
	}
}

// What the data says is what the plate draws, and the plate obeys the book's
// rendering rules (no rgba, no fill opacity, no em-dash).
func TestRisquesBriquesRendersEveryLabel(t *testing.T) {
	svg := figRisquesBriques()
	for _, n := range append(append(append([]cbNode{}, cbRisks...), cbBricks...), cbOffPortfolio...) {
		if !strings.Contains(svg, n.label) {
			t.Errorf("label %q never reaches the plate", n.label)
		}
	}
	for _, bad := range []string{"rgba(", "opacity=", "—"} {
		if strings.Contains(svg, bad) {
			t.Errorf("the plate contains %q, which crengine or the house rules forbid", bad)
		}
	}
	// One stroke per link, plus the accepted risk's dead-end stub.
	if got := strings.Count(svg, "<path d=\"M "); got != len(cbLinks) {
		t.Errorf("%d link paths drawn for %d links", got, len(cbLinks))
	}
}
