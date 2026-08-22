package web

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// annuityParams is the shared scenario for the annuity tests: a couple
// retiring at 60 on a plan long enough for the longevity tail to matter.
func annuityParams() Params {
	pr := testParams()
	pr.Age, pr.Years = 60, 40
	pr.NPaths = 600
	return pr
}

// The control off is the zero value, and it stays off through the whole
// pricing layer.
func TestAnnuityOffByDefault(t *testing.T) {
	if a := annuityParams().annuity(); a != nil {
		t.Errorf("annuity() = %+v with no share, want nil", a)
	}
}

// The product the page buys, and the two clamps that keep it inside the plan:
// the page starts at retirement, so a purchase never lands before year 0, and
// never in a year the plan does not reach.
func TestAnnuityProduct(t *testing.T) {
	pr := annuityParams()
	pr.AnnuityShare, pr.AnnuityYear, pr.AnnuityLoad = 0.25, 10, 0.08
	a := pr.annuity()
	if a == nil {
		t.Fatal("annuity() = nil with a share set")
	}
	if a.Share != 0.25 || a.Year != 10 || a.Load != 0.08 {
		t.Errorf("annuity = %+v, want share 0.25 at year 10 with an 8%% load", a)
	}
	if !a.Joint || a.Rate != annuityRealRate || a.Law != annuitantLaw {
		t.Errorf("annuity = %+v, want a joint-life quote at the page's rate on the annuitant table", a)
	}
	if got := pr.annuityAge(a); got != 70 {
		t.Errorf("purchase age = %.0f, want 70 (retired at 60, bought in year 10)", got)
	}
	pr.AnnuityYear = -5
	if got := pr.annuity().Year; got != 0 {
		t.Errorf("year = %d for a purchase before retirement, want 0", got)
	}
	pr.AnnuityYear = pr.Years + 5
	if got := pr.annuity().Year; got != pr.Years-1 {
		t.Errorf("year = %d past the horizon, want the last plan year %d", got, pr.Years-1)
	}
	pr.AnnuityShare = 2
	if got := pr.annuity().Share; got != 1 {
		t.Errorf("share = %.2f, want it capped at the whole sleeve", got)
	}
}

// An unset margin means the honest default, not an actuarially fair annuity:
// a link written before the control existed carries no margin at all, and must
// not quietly buy a product nobody sells.
func TestAnnuityLoadSentinel(t *testing.T) {
	pr := annuityParams()
	if got := pr.annuityLoad(); got != defaultAnnuityLoad {
		t.Errorf("unset load = %.2f, want the %.2f default", got, defaultAnnuityLoad)
	}
	pr.AnnuityLoad = 0.03
	if got := pr.annuityLoad(); got != 0.03 {
		t.Errorf("load = %.2f, want the posted 0.03", got)
	}
	pr.AnnuityLoad = 5
	if got := pr.annuityLoad(); got != maxAnnuityLoad {
		t.Errorf("load = %.2f, want it bounded at %.2f", got, maxAnnuityLoad)
	}
}

// The fixed-horizon views are left untouched by the control, by construction:
// the plan they run is the same plan, share or no share. This is also the
// back-compatibility guarantee for every link written before the panel: with
// no annuity, nothing on the page moved at all.
func TestAnnuityLeavesTheFixedHorizonPlanAlone(t *testing.T) {
	pr := annuityParams()
	plain := pr.plan()
	pr.AnnuityShare, pr.AnnuityLoad = 0.5, 0.10
	annuitised := pr.plan()
	if annuitised.Capital != plain.Capital {
		t.Errorf("capital %.0f vs %.0f: the premium must not leave a fixed-horizon plan",
			annuitised.Capital, plain.Capital)
	}
	if !reflect.DeepEqual(annuitised.Cashflows, plain.Cashflows) {
		t.Errorf("cashflows %+v vs %+v: an annuity is not a cashflow paid to a certain horizon",
			annuitised.Cashflows, plain.Cashflows)
	}
	if annuitised.Annuity != nil {
		t.Error("plan() must leave the annuity to the mortality kernel")
	}
	// Year 0 is before the pension starts, so the deterministic income the
	// spending fan adds back is zero: no annuity floor sneaks in there either.
	if got := pr.cashflowAt(0); got != 0 {
		t.Errorf("cashflowAt(0) = %.0f, want 0: no annuity income in a fixed-horizon year", got)
	}
	// And the two fixed-horizon views that used to read the old control render
	// identically with it on.
	prPlain := annuityParams()
	if Income(pr, nil).SVG != Income(prPlain, nil).SVG {
		t.Error("the funding-mix stack moved: the annuity must not reach a fixed-horizon view")
	}
	if Spending(pr, nil).SVG != Spending(prPlain, nil).SVG {
		t.Error("the spending fan moved: the annuity must not reach a fixed-horizon view")
	}
}

// The lifecycle view is where the purchase happens: it flags itself, and adds
// the before-and-after readout to its cards.
func TestLifecycleAnnuityReadout(t *testing.T) {
	pr := annuityParams()
	plain := Lifecycle(pr, nil)
	if plain.Annuitised {
		t.Error("Annuitised set with no annuity")
	}
	for _, c := range plain.Cards {
		if strings.HasPrefix(c.Label, "Annuity") {
			t.Errorf("annuity card %q on a plan that bought none", c.Label)
		}
	}

	pr.AnnuityShare = 0.5
	with := Lifecycle(pr, nil)
	if !with.Annuitised {
		t.Error("Annuitised not set on an annuitised plan")
	}
	if len(with.Cards) != len(plain.Cards)+3 {
		t.Fatalf("cards = %d, want the %d base cards plus the three-card readout",
			len(with.Cards), len(plain.Cards))
	}
	readout := with.Cards[len(plain.Cards):]
	for _, c := range readout {
		if !strings.HasPrefix(c.Label, "Annuity") {
			t.Errorf("card %q is not part of the annuity readout", c.Label)
		}
		if c.Help == "" {
			t.Errorf("card %q has no hover explanation", c.Label)
		}
	}
	// The first two cards are before-and-after pairs, the third the ratio that
	// explains them; the front end sizes them by exactly these markers.
	if !strings.Contains(readout[0].Value, "→") || !strings.Contains(readout[1].Value, "→") {
		t.Errorf("readout values %q / %q, want before-and-after pairs", readout[0].Value, readout[1].Value)
	}
	if !strings.Contains(readout[2].Value, " vs ") {
		t.Errorf("readout value %q, want the pays-versus-withdraws ratio", readout[2].Value)
	}
	// The purchase really reached the kernel: the premium leaves the portfolio
	// and pays its tax, so what is left at the household's end must fall.
	before, after := estatePair(t, readout[1].Value)
	if after >= before {
		t.Errorf("median estate %s: an annuity is paid for out of the bequest", readout[1].Value)
	}
	// And the mortality-free card says it does not carry the annuity.
	if !strings.Contains(with.Cards[0].Help, "no annuity") {
		t.Errorf("the ignoring-mortality card must say it excludes the annuity: %q", with.Cards[0].Help)
	}
}

// estatePair parses a "543 k€ → 212 k€" card value into its two figures. The
// unit is carried once when both figures share it ("2.63 → 1.69 M€"), so the
// left one inherits it.
func estatePair(t *testing.T, v string) (before, after float64) {
	t.Helper()
	parts := strings.Split(v, "→")
	if len(parts) != 2 {
		t.Fatalf("value %q is not a pair", v)
	}
	left := strings.TrimSpace(parts[0])
	if !strings.HasSuffix(left, "€") {
		left += " " + strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(parts[1]), "0123456789."))
	}
	return wealthValue(t, left), wealthValue(t, parts[1])
}

// wealthValue reads one fmtWealth figure back into euros.
func wealthValue(t *testing.T, s string) float64 {
	t.Helper()
	s = strings.TrimSpace(s)
	mult := 1000.0
	if strings.HasSuffix(s, "M€") {
		mult = 1e6
	}
	s = strings.TrimSuffix(strings.TrimSuffix(s, "M€"), "k€")
	var v float64
	if _, err := fmt.Sscanf(strings.TrimSpace(s), "%f", &v); err != nil {
		t.Fatalf("cannot read %q: %v", s, err)
	}
	return v * mult
}
