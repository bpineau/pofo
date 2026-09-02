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
