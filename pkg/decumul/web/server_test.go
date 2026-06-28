package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPISim(t *testing.T) {
	body, _ := json.Marshal(Params{
		Capital: 1_500_000, NeedAnnual: 48000, BufferYears: 3,
		Mu: 0.035, Sigma: 0.12, Df: 6, Years: 35, NPaths: 3000, TaxRate: 0.30,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/sim", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	Handler(nil, nil).ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status = %d", rec.Code)
	}
	var res Result
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("bad json: %v", err)
	}
	if res.BufferSVG == "" || res.Cards["ruin"] == "" {
		t.Errorf("empty result: %+v", res.Cards)
	}
}

func TestServesIndex(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	Handler(nil, nil).ServeHTTP(rec, req)
	if rec.Code != 200 || !bytes.Contains(rec.Body.Bytes(), []byte("<html")) {
		t.Errorf("index not served: %d", rec.Code)
	}
}
