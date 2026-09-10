package web

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestBoundedClampsEverySizeField(t *testing.T) {
	pr := Params{NPaths: 1 << 30, Years: 1 << 30, PensionYear: -5, SideUntilYear: 9999, BufferStopYear: 9999, Age: 500}.bounded()
	if pr.NPaths != maxPaths || pr.Years != maxYears {
		t.Errorf("paths/years not clamped: %d %d", pr.NPaths, pr.Years)
	}
	if pr.PensionYear != 0 || pr.SideUntilYear != maxYears || pr.BufferStopYear != maxYears || pr.Age != 110 {
		t.Errorf("year-like fields not clamped: %+v", pr)
	}
	// In-range values pass through untouched.
	in := Params{NPaths: 3000, Years: 35, PensionYear: 12, Age: 52}
	if got := in.bounded(); got.NPaths != in.NPaths || got.Years != in.Years || got.PensionYear != in.PensionYear || got.Age != in.Age {
		t.Errorf("in-range params altered: %+v", got)
	}
}

// The tax book has bounds of its own: a gain fraction is a fraction, and
// neither envelope can hold more than the invested capital (1 M€ less the
// 120 k€ of cash buffer here).
func TestBoundedClampsTheTaxBook(t *testing.T) {
	in := Params{Capital: 1_000_000, NeedAnnual: 40_000, BufferYears: 3,
		GainFrac: 3, PEACapital: 5_000_000, AVCapital: -1}
	pr := in.bounded()
	if pr.GainFrac != 1 || pr.PEACapital != 880_000 || pr.AVCapital != 0 {
		t.Errorf("tax book not clamped: %+v", pr)
	}
	// A coherent book passes through untouched.
	ok := Params{Capital: 1_000_000, NeedAnnual: 40_000, BufferYears: 3,
		GainFrac: 0.5, PEACapital: 200_000, AVCapital: 100_000}
	if got := ok.bounded(); got.GainFrac != ok.GainFrac || got.PEACapital != ok.PEACapital || got.AVCapital != ok.AVCapital {
		t.Errorf("coherent tax book altered: %+v", got)
	}
	if err := ok.validate(); err != nil {
		t.Errorf("coherent tax book refused: %v", err)
	}
}

// Two pockets that add up to more than the sleeve they are carved from are a
// contradiction, not a plan: clamping would silently pro-rate them and drop
// the taxable pocket, so the request is refused and told why.
func TestValidateRejectsAnOverfullTaxBook(t *testing.T) {
	pr := Params{Capital: 1_000_000, NeedAnnual: 40_000, BufferYears: 3,
		PEACapital: 600_000, AVCapital: 400_000}
	err := pr.validate()
	if err == nil {
		t.Fatal("an over-full envelope book was accepted")
	}
	for _, want := range []string{"PEA 600000", "assurance-vie 400000", "880000", "cash buffer"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("message %q does not name %s", err, want)
		}
	}
	body := `{"capital":1000000,"needAnnual":40000,"bufferYears":3,"years":30,` +
		`"mu":0.05,"sigma":0.11,"df":5,"nPaths":200,"peaCapital":600000,"avCapital":400000}`
	req := httptest.NewRequest(http.MethodPost, "/api/sim", strings.NewReader(body))
	rec := httptest.NewRecorder()
	Handler(nil, nil).ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "more than the invested capital") {
		t.Errorf("body = %q", rec.Body.String())
	}
}

// A request asking for a billion paths must be answered from the clamped
// count, in the time a normal request takes, never allocated as asked.
func TestAPISimClampsHugePathCount(t *testing.T) {
	body := []byte(`{"capital":1000000,"needAnnual":40000,"years":40,"mu":0.04,"sigma":0.12,"df":5,"nPaths":1000000000}`)
	req := httptest.NewRequest(http.MethodPost, "/api/sim", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	start := time.Now()
	Handler(nil, nil).ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body.String())
	}
	if d := time.Since(start); d > 30*time.Second {
		t.Errorf("clamped request took %s", d)
	}
}

func TestAPIRejectsOversizedBody(t *testing.T) {
	body := `{"capital":1000000,"pad":"` + strings.Repeat("x", maxBodyBytes) + `"}`
	for _, path := range []string{"/api/sim", "/api/fit"} {
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
		rec := httptest.NewRecorder()
		Handler(nil, nil).ServeHTTP(rec, req)
		if rec.Code == 200 {
			t.Errorf("%s accepted a %d-byte body", path, len(body))
		}
	}
}

// A request whose client has already gone away is refused at the gate, not
// computed: the gate's slots are all held, so acquire can only return on
// the context.
func TestSimGateRefusesAbandonedRequest(t *testing.T) {
	g := newSimGate(1)
	if !g.acquire(httptest.NewRequest(http.MethodPost, "/api/sim", nil)) {
		t.Fatal("first acquire should succeed")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequest(http.MethodPost, "/api/sim", nil).WithContext(ctx)
	if g.acquire(req) {
		t.Error("acquire succeeded on a canceled request with no free slot")
	}
	g.release()
	if !g.acquire(httptest.NewRequest(http.MethodPost, "/api/sim", nil)) {
		t.Error("slot not released")
	}
}

func TestAPISimGateReturns503WhenAbandoned(t *testing.T) {
	h := Handler(nil, nil, withSimGate(newSimGate(0)))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	body, _ := json.Marshal(Params{Capital: 1e6, NeedAnnual: 4e4, Years: 30, NPaths: 100})
	req := httptest.NewRequest(http.MethodPost, "/api/sim", bytes.NewReader(body)).WithContext(ctx)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", rec.Code)
	}
}

// A horizon of zero is not a short retirement but an empty one, and several
// views index the last plan year unconditionally: a POST of "{}" used to
// index -1 inside /api/income and take the whole process down (the panic
// happens in a simulation worker, where net/http cannot recover it). The
// horizon is therefore floored at one year, not at zero.
func TestBoundedFloorsTheHorizon(t *testing.T) {
	if got := (Params{}).bounded().Years; got != 1 {
		t.Errorf("empty request: years = %d, want the 1-year floor", got)
	}
	if got := (Params{Years: -30}).bounded().Years; got != 1 {
		t.Errorf("negative years = %d, want the 1-year floor", got)
	}
	h := Handler(nil, nil)
	for _, path := range []string{"/api/income", "/api/sensitivity", "/api/lifecycle", "/api/sim"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{}`)))
		if rec.Code != http.StatusOK {
			t.Errorf("%s on an empty body: status %d: %s", path, rec.Code, rec.Body.String())
		}
	}
}

// The sensitivity chart's horizon lever shortens the plan by five years, so a
// plan shorter than that would hand the kernel a negative year count: an
// invalid plan, not a shorter one, and another crash of the whole process.
func TestSensitivityHorizonNudgeStaysValid(t *testing.T) {
	for _, years := range []int{1, 2, 4, 5, 6} {
		pr := Params{Capital: 1e6, NeedAnnual: 4e4, Years: years,
			Mu: 0.04, Sigma: 0.12, Df: 5, NPaths: 50}
		if svg := Sensitivity(pr, nil).SVG; !strings.HasPrefix(svg, "<svg") {
			t.Errorf("%d-year horizon: no chart", years)
		}
	}
}
