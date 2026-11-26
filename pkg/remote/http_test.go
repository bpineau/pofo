package remote

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bpineau/pofo/pkg/analyze"
	"github.com/bpineau/pofo/pkg/marketdata"
)

func callHTTP(h http.Handler, path, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest("POST", path, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Accept", "application/json, text/event-stream")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}
func TestHTTPMCPParity(t *testing.T) {
	h := Handler(New(nil, "test"), HTTPOptions{API: true, MCP: true})
	api := httptest.NewRecorder()
	h.ServeHTTP(api, httptest.NewRequest("GET", "/api/v1/assets/search?query=IWDA&limit=1", nil))
	mcp := callHTTP(h, "/mcp", `{"jsonrpc":"2.0","id":"exact-id","method":"tools/call","params":{"name":"search_assets","arguments":{"query":"IWDA","limit":1}}}`)
	if api.Code != 200 || mcp.Code != 200 {
		t.Fatalf("%d %s / %d %s", api.Code, api.Body, mcp.Code, mcp.Body)
	}
	var a, b map[string]any
	json.Unmarshal(api.Body.Bytes(), &a)
	var rpc struct {
		ID     string
		Result struct {
			Content           []struct{ Type, Text string }
			StructuredContent any
		}
	}
	if err := json.Unmarshal(mcp.Body.Bytes(), &rpc); err != nil {
		t.Fatal(err)
	}
	if rpc.ID != "exact-id" || len(rpc.Result.Content) != 1 || rpc.Result.StructuredContent != nil {
		t.Fatalf("%s", mcp.Body)
	}
	json.Unmarshal([]byte(rpc.Result.Content[0].Text), &b)
	aa, _ := json.Marshal(a["data"])
	bb, _ := json.Marshal(b["data"])
	if string(aa) != string(bb) {
		t.Fatal("different domain outputs")
	}
}
func TestProtocolAndTransport(t *testing.T) {
	h := Handler(New(nil, "test"), HTTPOptions{API: true, MCP: true})
	for _, tc := range []struct {
		body     string
		status   int
		contains string
	}{
		{`{"jsonrpc":"2.0","id":1234567890123456789,"method":"ping"}`, 200, `1234567890123456789`},
		{`{"jsonrpc":"2.0","method":"notifications/initialized"}`, 202, ""},
		{`{"jsonrpc":"2.0","method":"notifications/cancelled","params":{"requestId":1}}`, 202, ""},
		{`{"jsonrpc":"2.0","id":null,"method":"ping"}`, 200, `-32600`},
		{`{"jsonrpc":"2.0","id":1.0,"method":"ping"}`, 200, `"id":1.0`},
		{`{"jsonrpc":"2.0","id":1e3,"method":"ping"}`, 200, `"id":1e3`},
		{`{"jsonrpc":"2.0","id":1.2,"method":"ping"}`, 200, `-32600`},
		{`{"jsonrpc":"2.0","id":1,"method":"unknown"}`, 200, `-32601`},
		{`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"search_assets","arguments":{"limit":0}}}`, 200, `isError`},
		{`[]`, 200, `-32600`},
	} {
		w := callHTTP(h, "/mcp", tc.body)
		if w.Code != tc.status || !strings.Contains(w.Body.String(), tc.contains) {
			t.Errorf("%s: %d %s", tc.body, w.Code, w.Body)
		}
		if tc.status == 202 && w.Body.Len() != 0 {
			t.Fatal("notification body")
		}
	}
	r := httptest.NewRequest("POST", "/mcp", strings.NewReader(`{}`))
	r.Header.Set("MCP-Protocol-Version", "2099-01-01")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 400 {
		t.Fatal(w.Code)
	}
	w = callHTTP(h, "/mcp", `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2099-01-01","capabilities":{},"clientInfo":{"name":"test","version":"1"}}}`)
	if w.Code != 200 || !strings.Contains(w.Body.String(), "2025-11-25") {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
	r = httptest.NewRequest("POST", "/mcp", strings.NewReader(`{}`))
	r.Header.Set("Origin", "https://evil.example")
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 403 {
		t.Fatal(w.Code)
	}
}
func TestSchemasAndSizes(t *testing.T) {
	b, err := json.Marshal(Tools())
	if err != nil {
		t.Fatal(err)
	}
	if len(b) > 15<<10 {
		t.Fatalf("tool schemas %d bytes", len(b))
	}
	if strings.Contains(string(b), "outputSchema") || len(Instructions) > 2<<10 {
		t.Fatal("MCP duplication/instruction budget")
	}
	t.Logf("tools/list definitions: %d bytes; instructions: %d bytes", len(b), len(Instructions))
}

func TestTransportLimitsAndOrigins(t *testing.T) {
	h := Handler(New(nil, "test"), HTTPOptions{API: true, MCP: true, Origins: []string{"chrome-extension://example"}})
	r := httptest.NewRequest("OPTIONS", "/api/v1/fire/simulate", nil)
	r.Header.Set("Origin", "chrome-extension://example")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 204 || w.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Fatalf("preflight %d", w.Code)
	}
	r = httptest.NewRequest("POST", "/mcp", strings.NewReader(`{}`))
	r.Header.Set("Accept", "application/json;q=0.0,text/event-stream")
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 406 {
		t.Fatal("accepted zero quality")
	}
	w = callHTTP(h, "/mcp", `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":2,"version":true}}}`)
	if !strings.Contains(w.Body.String(), "-32602") {
		t.Fatal("invalid clientInfo accepted")
	}
	w = callHTTP(h, "/api/v1/assets/analyze", strings.Repeat("x", MaxRequestBytes+1))
	if w.Code != 413 {
		t.Fatal(w.Code)
	}
	for i := 0; i < 10; i++ {
		w = httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", "/api/v1/assets/search?limit=1", nil))
		if w.Code != 200 {
			t.Fatalf("request %d: %d", i, w.Code)
		}
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/api/v1/assets/search?limit=1", nil))
	if w.Code != 429 || w.Header().Get("Retry-After") == "" {
		t.Fatal("unbounded request rate")
	}
}

func TestResponseCapPreservesMCPDomainError(t *testing.T) {
	src := fixtureSource()
	src["AAA"].Warnings = []string{strings.Repeat("q", MaxResponseBytes)}
	h := Handler(New(src, "test"), HTTPOptions{API: true, MCP: true})
	w := callHTTP(h, "/api/v1/assets/analyze", `{"ids":["AAA"]}`)
	if w.Code != 422 || !strings.Contains(w.Body.String(), "response_too_large") {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
	w = callHTTP(h, "/mcp", `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"analyze_assets","arguments":{"ids":["AAA"]}}}`)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"isError":true`) || !strings.Contains(w.Body.String(), "response_too_large") {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
}

type waitingSource struct {
	analyze.SeriesSource
	started chan struct{}
	release chan struct{}
}

func (s waitingSource) FetchExtended(ctx context.Context, id string, opt marketdata.FetchOptions) (*marketdata.Series, error) {
	close(s.started)
	select {
	case <-s.release:
		return s.SeriesSource.FetchExtended(ctx, id, opt)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
func TestCancellationNotificationDoesNotOwnAnotherCall(t *testing.T) {
	src := waitingSource{fixtureSource(), make(chan struct{}), make(chan struct{})}
	h := Handler(New(src, "test"), HTTPOptions{MCP: true})
	done := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		done <- callHTTP(h, "/mcp", `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"analyze_assets","arguments":{"ids":["AAA"]}}}`)
	}()
	<-src.started
	w := callHTTP(h, "/mcp", `{"jsonrpc":"2.0","method":"notifications/cancelled","params":{"requestId":1}}`)
	if w.Code != 202 {
		t.Fatal(w.Code)
	}
	select {
	case response := <-done:
		t.Fatalf("another request was cancelled: %s", response.Body)
	default:
	}
	close(src.release)
	response := <-done
	if !strings.Contains(response.Body.String(), `"isError":false`) {
		t.Fatal(response.Body)
	}
}
