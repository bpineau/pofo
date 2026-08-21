package firebook

import (
	"math"
	"strings"
	"testing"
)

// The theorem the plate exists to draw: the two accountings of the same
// position land on the SAME number at the horizon, and not approximately. The
// face is fixed at the purchase and the market curve is that face discounted
// over the maturity left, so at the horizon there is nothing left to discount:
// the equality is an identity of the arithmetic, and the test holds it to a
// billionth rather than to a tolerance.
func TestTenirConvergesExactlyAtTheHorizon(t *testing.T) {
	book, market := tenirBook(tenirHorizon), tenirMarket(tenirHorizon)
	if math.Abs(book-market) > 1e-9 {
		t.Errorf("the two accountings end at %.12f and %.12f, a gap of %.3g",
			book, market, book-market)
	}
	if math.Abs(book-tenirFace()) > 1e-9 || math.Abs(market-tenirFace()) > 1e-9 {
		t.Errorf("the horizon wealth is %.12f / %.12f, not the face %.12f",
			book, market, tenirFace())
	}
	if math.Abs(tenirGap(tenirHorizon)) > 1e-9 {
		t.Errorf("the plate would draw a gap of %.3g at the horizon", tenirGap(tenirHorizon))
	}
	// And the shock does not move the horizon wealth at all: that is what
	// "tenir à échéance verrouille le rendement" means, and it is why the
	// opportunity cost is the whole of the loss.
	if math.Abs(tenirFace()-tenirCapital*math.Pow(1+tenirYield0, tenirHorizon)) > 1e-12 {
		t.Error("the face no longer depends on the purchase yield alone")
	}
}

// The shape of the two curves, which is the rest of the argument: identical
// before the shock, split by it, never crossing after it, and both climbing
// back to the same point.
func TestTenirShape(t *testing.T) {
	// Before the shock the market and the book are the same number.
	for _, m := range []int{0, 3, 6, 11} {
		y := float64(m) / 12
		if math.Abs(tenirGap(y)) > 1e-9 {
			t.Errorf("month %d: the two accountings already differ by %.3g", m, tenirGap(y))
		}
	}
	// The shock prints a fall on the market side and nothing at all on the
	// book side.
	if tenirDrop() >= 0 {
		t.Errorf("the shock raises the printed price by %.2f %%", tenirDrop()*100)
	}
	if tenirBook(tenirShock) <= tenirBook(tenirShock-1.0/12) {
		t.Error("the book value fell at the shock, which amortized cost never does")
	}
	// After the shock the printed price is always below the book, and both
	// climb, month after month.
	for m := int(tenirShock * 12); m < int(tenirHorizon*12); m++ {
		y, next := float64(m)/12, float64(m+1)/12
		if m > int(tenirShock*12) && tenirGap(y) <= 0 {
			t.Errorf("month %d: the printed price is no longer below the book value", m)
		}
		if tenirMarket(next) <= tenirMarket(y) {
			t.Errorf("month %d: the printed price stopped climbing", m)
		}
		if tenirBook(next) <= tenirBook(y) {
			t.Errorf("month %d: the book value stopped climbing", m)
		}
		if tenirGap(next) > tenirGap(y)+1e-12 {
			t.Errorf("month %d: the gap widened after the shock", m)
		}
	}
}

// The readings the plate prints, pinned: the widest gap and the month it falls
// on, the printed fall, the face and the month the price is back to where it
// was. All of them are closed-form consequences of the four constants above.
func TestTenirReadings(t *testing.T) {
	worstM, worst := 0, 0.0
	for m := 0; m <= int(tenirHorizon*12); m++ {
		if g := tenirGap(float64(m) / 12); g > worst {
			worstM, worst = m, g
		}
	}
	if worstM != 12 {
		t.Errorf("the gap is widest at month %d, the plate cotes it at the shock (month 12)", worstM)
	}
	if math.Abs(worst-11.2177) > 1e-3 {
		t.Errorf("the widest gap is %.4f points, the plate draws 11,2", worst)
	}
	if math.Abs(tenirDrop()*100+10.9977) > 1e-3 {
		t.Errorf("the printed fall is %.4f %%, the plate says 11,0", tenirDrop()*100)
	}
	if math.Abs(tenirFace()-114.868566) > 1e-5 {
		t.Errorf("the face is %.6f, the plate says 114,87", tenirFace())
	}
	if got := tenirRecoveryMonth(); got != 48 {
		t.Errorf("the printed price is back to its pre-shock level at month %d, the plate says 48", got)
	}
	// The fall is close to the duration rule the article states (remaining
	// maturity times the shift), and below it, which is convexity.
	rule := (tenirHorizon - tenirShock) * (tenirYield1 - tenirYield0) * 100
	if -tenirDrop()*100 >= rule {
		t.Errorf("the fall (%.2f %%) is not smaller than the duration rule (%.2f %%)",
			-tenirDrop()*100, rule)
	}
	if rule+tenirDrop()*100 > 1.5 {
		t.Errorf("the fall (%.2f %%) is %.2f points away from the duration rule (%.2f %%)",
			-tenirDrop()*100, rule+tenirDrop()*100, rule)
	}
}

// The plate against the article that carries it: its numbers are the article's
// own, and so is its vocabulary.
func TestTenirAgreesWithItsArticle(t *testing.T) {
	art := bookArticle(t, "obligations-en-retrait")
	for _, want := range []string{
		"::: figure tenir-echeance",
		"Tenir à échéance verrouille le rendement nominal promis.",
		"Vous touchez 2 % quand le marché en sert 4, au lieu d'une baisse de prix affichée.",
		"La richesse finale est identique, seule la comptabilité mentale change.",
		"sous forme de coût d'opportunité",
		"environ 7 pour l'aggregate standard",
		"la duration de la poche doit rester inférieure ou égale à l'horizon",
	} {
		if !strings.Contains(art, want) {
			t.Errorf("the article no longer says %q", want)
		}
	}
	// The plate's constants ARE the article's example: a 2 % coupon against a
	// 4 % market, on the seven-year duration it calls the standard aggregate.
	if tenirYield0 != 0.02 || tenirYield1 != 0.04 {
		t.Errorf("the plate shocks %.2f %% to %.2f %%, the article says 2 to 4",
			tenirYield0*100, tenirYield1*100)
	}
	if tenirHorizon != 7 {
		t.Errorf("the plate runs to %.0f years, the article's aggregate duration is 7", tenirHorizon)
	}
}

// The rendered plate obeys the book's drawing rules and carries every reading
// it claims to draw.
func TestTenirPlateRenders(t *testing.T) {
	svg := FigureSVG("tenir-echeance")
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
	for _, want := range []string{
		">coût d'opportunité invisible<",
		">baisse de prix affichée<",
		">le titre détenu à échéance<",
		">le fonds, valorisé au marché<",
		">+2 points de taux, au mois 12<",
		">11,2 points<",
		"même richesse finale : 114,87 pour 100 investis",
		"le prix tombe de 11,0 % au choc",
		"48 mois au prix affiché pour retrouver son niveau d'avant",
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("the plate does not carry %q", want)
		}
	}
	// The two curves, and the two halves of the band between them.
	if n := strings.Count(svg, "<polyline"); n != 2 {
		t.Errorf("%d curves drawn, the plate has exactly two", n)
	}
	if n := strings.Count(svg, "<path"); n != 2 {
		t.Errorf("%d filled bands drawn, expected the two halves of the gap", n)
	}
	if n := strings.Count(svg, "<circle"); n != 1 {
		t.Errorf("%d dots drawn, expected the convergence one alone", n)
	}
}
