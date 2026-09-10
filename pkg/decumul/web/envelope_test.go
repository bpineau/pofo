package web

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/bpineau/pofo/pkg/decumul"
)

// taxBook is a plan the tax controls can be read on: 1 M€ of capital and a
// 3-year buffer on 40 k€ of spending, so the growth sleeve is 880 k€.
func taxBook() Params {
	return Params{Capital: 1_000_000, NeedAnnual: 40_000, BufferYears: 3, Years: 30,
		Mu: 0.05, Sigma: 0.11, Df: 5, TaxRate: 0.328, NPaths: 200}
}

// A request naming no wrapper and no embedded gain keeps the historical
// single sleeve: nil Envelopes, a cost basis equal to the capital, and the
// exact plan every already-shared URL reproduced before the controls existed.
func TestNoTaxBookKeepsTheSingleSleeve(t *testing.T) {
	p := taxBook().plan()
	if p.Envelopes != nil {
		t.Fatalf("Envelopes = %+v, want nil", p.Envelopes)
	}
	if got, ok := p.Tax.(decumul.CTOFlatTax); !ok || got.Rate != 0.328 {
		t.Errorf("Tax = %+v, want CTOFlatTax at the slider rate", p.Tax)
	}
}

// The embedded gain alone builds one taxable pocket at the slider's rate: the
// same single rate, now applied to something.
func TestEmbeddedGainAloneBuildsOnePocket(t *testing.T) {
	pr := taxBook()
	pr.GainFrac = 0.5
	env := pr.plan().Envelopes
	if len(env) != 1 || env[0].Name != "CTO" {
		t.Fatalf("envelopes = %+v, want one CTO pocket", env)
	}
	if env[0].GainFrac != 0.5 || env[0].Amount != 880_000 {
		t.Errorf("pocket = %+v, want the whole growth sleeve at 50%% gain", env[0])
	}
	if got, ok := env[0].Tax.(decumul.CTOFlatTax); !ok || got.Rate != 0.328 {
		t.Errorf("pocket tax = %+v, want the slider rate", env[0].Tax)
	}
}

// Naming the two wrappers splits the sleeve into the French drain order, the
// taxable account first, and leaves it the remainder.
func TestEnvelopeAmountsSplitTheGrowthSleeve(t *testing.T) {
	pr := taxBook()
	pr.GainFrac = 0.5
	pr.PEACapital, pr.AVCapital = 200_000, 100_000
	env := pr.plan().Envelopes
	if len(env) != 3 {
		t.Fatalf("envelopes = %+v, want CTO, PEA and AV", env)
	}
	want := []struct {
		name   string
		amount float64
	}{{"CTO", 580_000}, {"PEA", 200_000}, {"AV", 100_000}}
	for i, w := range want {
		if env[i].Name != w.name || env[i].Amount != w.amount {
			t.Errorf("pocket %d = %+v, want %s at %.0f", i, env[i], w.name, w.amount)
		}
		if env[i].GainFrac != 0.5 {
			t.Errorf("pocket %d gain = %v, want the shared 0.5", i, env[i].GainFrac)
		}
	}
	if got, ok := env[1].Tax.(decumul.CTOFlatTax); !ok || got.Rate != 0.186 {
		t.Errorf("PEA tax = %+v, want the 18.6%% social levies", env[1].Tax)
	}
	if got, ok := env[2].Tax.(decumul.AVTax); !ok || got.Rate != 0.247 || got.Allowance != 9200 {
		t.Errorf("AV tax = %+v, want 24.7%% past the couple's 9 200 € allowance", env[2].Tax)
	}
}

// The embedded gain is what the rate bites on: a book made entirely of cost
// basis is taxed only on what it earns after retirement, so it pays a
// visibly lighter bill than the same plan half of whose capital is already
// gain. That gap is the reason the control exists.
func TestEmbeddedGainRaisesTheTaxBill(t *testing.T) {
	pr := taxBook()
	free := Compute(pr)
	pr.GainFrac = 0.5
	taxed := Compute(pr)
	if pct(t, free, "Effective tax rate") >= pct(t, taxed, "Effective tax rate") {
		t.Errorf("effective tax rate did not rise with the embedded gain: %s then %s",
			card(t, free, "Effective tax rate"), card(t, taxed, "Effective tax rate"))
	}
	if card(t, free, "Median cumulative tax") == card(t, taxed, "Median cumulative tax") {
		t.Errorf("median cumulative tax unchanged by the embedded gain: %s",
			card(t, taxed, "Median cumulative tax"))
	}
}

// card reads one summary figure by label, pct its percentage value.
func card(t *testing.T, r Result, label string) string {
	t.Helper()
	for _, c := range r.Cards {
		if c.Label == label {
			return c.Value
		}
	}
	t.Fatalf("no %q card", label)
	return ""
}

func pct(t *testing.T, r Result, label string) float64 {
	t.Helper()
	v, err := strconv.ParseFloat(strings.TrimSuffix(card(t, r, label), "%"), 64)
	if err != nil {
		t.Fatalf("card %q: %v", label, err)
	}
	return v
}

// A shared URL from before these controls existed carries none of the three
// fields, and must still produce exactly the run its sender saw: the zero
// values are the legacy single sleeve, byte for byte.
func TestLegacyRequestUnchangedByTheNewFields(t *testing.T) {
	const legacy = `{"capital":1000000,"needAnnual":40000,"bufferYears":3,"years":30,` +
		`"mu":0.05,"sigma":0.11,"df":5,"taxRate":0.328,"nPaths":400}`
	explicit := strings.TrimSuffix(legacy, `}`) + `,"gainFrac":0,"peaCapital":0,"avCapital":0}`
	post := func(body string) []byte {
		req := httptest.NewRequest(http.MethodPost, "/api/sim", strings.NewReader(body))
		rec := httptest.NewRecorder()
		Handler(nil, nil).ServeHTTP(rec, req)
		if rec.Code != 200 {
			t.Fatalf("status = %d: %s", rec.Code, rec.Body.String())
		}
		return rec.Body.Bytes()
	}
	if a, b := post(legacy), post(explicit); !bytes.Equal(a, b) {
		t.Errorf("an old shared URL no longer reproduces its run:\n%s\n%s", a, b)
	}
}

// The page's own control definitions: the two figures in the open, the two
// wrapper amounts folded away, all four inside the Taxes group, with the
// defaults the design doc recommends (a 50% embedded gain, no wrapper named).
func TestTaxControlsShipWithTheirDefaults(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/app.js", nil)
	rec := httptest.NewRecorder()
	Handler(nil, nil).ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("app.js not served: %d", rec.Code)
	}
	js := rec.Body.String()
	group := js[strings.Index(js, `{title: "Taxes"`):]
	group = group[:strings.Index(group, `{title: "Simulation"`)]
	for _, want := range []string{
		`r("taxRate", "Tax on gains", 0, 0.40, 0.01, 0.328, "pct",`,
		`r("gainFrac", "Embedded gain in the capital", 0, 1, 0.05, 0.5, "pct",`,
		`...fold("envelopes",`,
		`r("peaCapital", "PEA capital", 0, 3000000, 10000, 0, "eur",`,
		`r("avCapital", "Assurance-vie capital", 0, 3000000, 10000, 0, "eur",`,
	} {
		if !strings.Contains(group, want) {
			t.Errorf("the Taxes group is missing %s", want)
		}
	}
	// The two amounts belong to the disclosure, not to the open rail.
	if strings.Index(group, `...fold("envelopes",`) > strings.Index(group, `r("peaCapital"`) {
		t.Error("peaCapital sits outside the envelopes fold")
	}
	// Every control carries a hover, and the rate's now names the gain share.
	if !strings.Contains(group, "charged on the GAIN SHARE of every sale") {
		t.Error("the tax rate's help does not say what the rate applies to")
	}
	// The tax book is a standing fact about the household, so it rides in
	// the personal-defaults cookie beside the situation.
	for _, k := range []string{"gainFrac", "peaCapital", "avCapital"} {
		if !strings.Contains(js[strings.Index(js, "const SAVEKEYS"):strings.Index(js, "const PREF_COOKIE")], `"`+k+`"`) {
			t.Errorf("%s is not saved as a personal default", k)
		}
	}
}
