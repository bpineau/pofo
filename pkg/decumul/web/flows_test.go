package web

import (
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Every row is clamped into its bounds: amounts by kind, years to the
// horizon cap, the drift to +-5 %/yr, the label to 40 runes.
func TestBoundedFlowsClampsEveryRow(t *testing.T) {
	in := Params{Flows: []Flow{
		{Label: strings.Repeat("é", 90), Amount: 5e9, From: -3, To: 9999, Drift: 3},
		{Once: true, Amount: -5e12, From: 500},
		{Amount: math.NaN(), Drift: math.NaN()},
	}}
	got := in.bounded().Flows
	if f := got[0]; len([]rune(f.Label)) != maxFlowLabel || f.Amount != maxFlowYearly || f.From != 0 || f.To != maxYears || f.Drift != 0.05 {
		t.Errorf("yearly row not clamped: %+v", f)
	}
	if f := got[1]; f.Amount != -maxFlowOnce || f.From != maxYears {
		t.Errorf("one-off row not clamped: %+v", f)
	}
	if f := got[2]; f.Amount != 0 || f.Drift != 0 {
		t.Errorf("NaN row not zeroed: %+v", f)
	}
}

// A list longer than the page can post is refused, not truncated: dropping
// rows would simulate another household.
func TestTooManyFlowsAreRefused(t *testing.T) {
	rows := make([]string, maxFlows+1)
	for i := range rows {
		rows[i] = `{"amount":1000,"from":1}`
	}
	body := `{"capital":900000,"needAnnual":42000,"years":30,"mu":0.045,"sigma":0.13,"df":5,"nPaths":200,` +
		`"flows":[` + strings.Join(rows, ",") + `]}`
	rec := httptest.NewRecorder()
	Handler(nil, nil).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/sim", strings.NewReader(body)))
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "at most 12 dated flows") {
		t.Errorf("status %d: %s", rec.Code, rec.Body.String())
	}
}

// Yearly rows become drifting cashflows beside the two sliders, one-off rows
// become lumps, a zero row is nothing, and the views that sum the income read
// the yearly rows too, an expense counting negative.
func TestFlowsReachThePlan(t *testing.T) {
	pr := Params{Capital: 800_000, NeedAnnual: 30_000, Years: 30, PensionAnnual: 12_000, PensionYear: 10,
		Flows: []Flow{
			{Label: "studies", Amount: -8_000, From: 2, To: 6},
			{Label: "rent", Amount: 6_000, From: 0, To: 20, Drift: 0.01},
			{Label: "inheritance", Once: true, Amount: 250_000, From: 12},
			{Label: "nothing", Amount: 0, From: 3},
		}}
	p := pr.plan()
	if len(p.Cashflows) != 3 || p.Cashflows[1].Annual != -8_000 || p.Cashflows[2].Drift != 0.01 {
		t.Errorf("cashflows %+v", p.Cashflows)
	}
	if len(p.Lumps) != 1 || p.Lumps[0].Year != 12 || p.Lumps[0].Amount != 250_000 {
		t.Errorf("lumps %+v", p.Lumps)
	}
	if got, want := pr.cashflowAt(3), -8_000+6_000*math.Pow(1.01, 3); math.Abs(got-want) > 1e-9 {
		t.Errorf("cashflowAt(3) = %v, want %v", got, want)
	}
	if m := pr.lumpMarkers(pr.Years); len(m) != 1 || m[0].Value != 12 || m[0].Label != "+250k" {
		t.Errorf("markers %+v", m)
	}
	// The inheritance lowers ruin, the expense raises it: the list reaches the
	// kernel through the one plan construction every view shares.
	ruin := func(fs []Flow) float64 {
		q := pr
		q.Flows, q.Mu, q.Sigma, q.Df, q.NPaths = fs, 0.04, 0.12, 5, 2000
		p := q.plan()
		return p.RuinProb(q.NPaths, simWorkers, 7)
	}
	none, rich, poor := ruin(nil), ruin(pr.Flows[2:3]), ruin(pr.Flows[0:1])
	if !(rich < none && none < poor) {
		t.Errorf("ruin: with the inheritance %.3f, none %.3f, with the expense %.3f", rich, none, poor)
	}
}
