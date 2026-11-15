package remote

import (
	"context"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
)

func TestWarningsDeduplicatedWithDistinctPathsPreserved(t *testing.T) {
	r := &Result{}
	for range 3 {
		warnings(r, []string{"FX held constant"}, "market")
	}
	warnings(r, []string{"FX held constant"}, "benchmark")
	if len(r.Warnings) != 2 {
		t.Fatalf("warnings: %+v", r.Warnings)
	}
}

func TestAssetClassErrorAndSchemaListChoices(t *testing.T) {
	_, err := New(nil, "test").Execute(context.Background(), "search_assets", []byte(`{"asset_class":"bond"}`))
	if err == nil || !strings.Contains(err.Error(), "government-bond") {
		t.Fatalf("error: %v", err)
	}
	props := Tools()[0].InputSchema["properties"].(map[string]any)
	choices, _ := props["asset_class"].(map[string]any)["enum"].([]string)
	if !slices.Contains(choices, "government-bond") {
		t.Fatal(choices)
	}
}

func TestFIRETransportsDoNotUseAssetQueue(t *testing.T) {
	h := Handler(New(nil, "test"), HTTPOptions{API: true, MCP: true, Acquire: func(context.Context) (func(), error) {
		return nil, &Error{Code: "server_busy", Message: "asset queue full", Retryable: true}
	}})
	input := `{"plan":{"capital":800000,"annual_spending":30000,"horizon_years":2},` + parametric + `,"paths":100}`
	for _, w := range []*httptest.ResponseRecorder{
		callHTTP(h, "/api/v1/fire/simulate", input),
		callHTTP(h, "/mcp", `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"simulate_fire","arguments":`+input+`}}`),
	} {
		if w.Code != 200 || strings.Contains(w.Body.String(), "server_busy") {
			t.Fatalf("FIRE: %d %s", w.Code, w.Body)
		}
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/assets/search", nil))
	if w.Code != 503 {
		t.Fatalf("asset: %d", w.Code)
	}
}
