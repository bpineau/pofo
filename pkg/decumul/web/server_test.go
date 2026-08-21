package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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
	if res.ArbitrageSVG == "" || len(res.Cards) == 0 {
		t.Errorf("empty result: %+v", res.Cards)
	}
}

func TestAPISolve(t *testing.T) {
	body, _ := json.Marshal(Params{
		Capital: 1_500_000, NeedAnnual: 48000, BufferYears: 3,
		Mu: 0.035, Sigma: 0.12, Df: 6, Years: 35, NPaths: 2000, TaxRate: 0.30,
		TargetRuin: 0.05,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/solve", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	Handler(nil, nil).ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status = %d", rec.Code)
	}
	var res SolveResult
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("bad json: %v", err)
	}
	if res.RequiredCapital < solveLo || res.RequiredCapital > solveHi {
		t.Errorf("required capital %.0f outside the search bounds", res.RequiredCapital)
	}
	if res.BestBufferYears < 0 || res.BestBufferRuin < 0 || res.BestBufferRuin > 1 {
		t.Errorf("implausible buffer solve: %+v", res)
	}
}

func TestServesFontsCSS(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/fonts.css", nil)
	rec := httptest.NewRecorder()
	Handler(nil, nil).ServeHTTP(rec, req)
	if rec.Code != 200 || rec.Header().Get("Content-Type") != "text/css; charset=utf-8" ||
		!bytes.Contains(rec.Body.Bytes(), []byte("Instrument Sans")) {
		t.Errorf("fonts.css not served: %d %q", rec.Code, rec.Header().Get("Content-Type"))
	}
}

// metaOf reads /api/meta off a handler built with the given options.
func metaOf(t *testing.T, opts ...Option) map[string]any {
	t.Helper()
	rec := httptest.NewRecorder()
	Handler(nil, nil, opts...).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/meta", nil))
	if rec.Code != 200 {
		t.Fatalf("status = %d", rec.Code)
	}
	var meta map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &meta); err != nil {
		t.Fatalf("bad json: %v", err)
	}
	return meta
}

// The bare mount announces a generic market and no loader: the front end
// then shows the "generic market" pill and the static command-line hint.
func TestMetaWithoutPickerOrLabel(t *testing.T) {
	meta := metaOf(t)
	if got, ok := meta["sourceLabel"].(string); !ok || got != "" {
		t.Errorf("sourceLabel = %v, want an empty string", meta["sourceLabel"])
	}
	if _, ok := meta["picker"]; ok {
		t.Errorf("picker present without WithPicker: %v", meta["picker"])
	}
	if meta["hasPanel"] != false {
		t.Errorf("hasPanel = %v, want false", meta["hasPanel"])
	}
}

// A -serve mount carries both its provenance label and the loader payload
// the drawer's empty state is built from.
func TestMetaWithPickerAndLabel(t *testing.T) {
	meta := metaOf(t,
		WithSourceLabel("dragon-decumulation-household"),
		WithPicker(Picker{
			Base:       "/firesimulator",
			CatalogURL: "/catalog.json",
			ViewURL:    "/view",
			Examples:   []ExampleRef{{Name: "all-weather-dalio", Title: "All Weather", Blurb: "modernized"}},
		}))
	if meta["sourceLabel"] != "dragon-decumulation-household" {
		t.Errorf("sourceLabel = %v", meta["sourceLabel"])
	}
	p, ok := meta["picker"].(map[string]any)
	if !ok {
		t.Fatalf("picker = %v, want an object", meta["picker"])
	}
	if p["base"] != "/firesimulator" || p["catalogURL"] != "/catalog.json" || p["viewURL"] != "/view" {
		t.Errorf("picker mount = %v", p)
	}
	exs, ok := p["examples"].([]any)
	if !ok || len(exs) != 1 {
		t.Fatalf("examples = %v", p["examples"])
	}
	ex := exs[0].(map[string]any)
	if ex["name"] != "all-weather-dalio" || ex["title"] != "All Weather" || ex["blurb"] != "modernized" {
		t.Errorf("example = %v", ex)
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

// The offline mount serves both editions of the FIRE book, each pointing at
// the other: the -fire explorer is a complete, self-contained reading surface.
func TestFirebookMounts(t *testing.T) {
	// The hreflang hrefs are fully qualified, built from the request's own
	// origin (httptest's default host).
	for path, want := range map[string]string{
		"/firebook/fr/": `<link rel="alternate" hreflang="en" href="http://example.com/firebook/en/">`,
		"/firebook/en/": `<link rel="alternate" hreflang="fr" href="http://example.com/firebook/fr/">`,
	} {
		rec := httptest.NewRecorder()
		Handler(nil, nil).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusOK {
			t.Errorf("%s: status = %d", path, rec.Code)
		}
		if !strings.Contains(rec.Body.String(), want) {
			t.Errorf("%s: misses the cross-edition link %q", path, want)
		}
	}
}

// The standalone explorer publishes the same machine-readable root files as
// the -serve constellation: a crawler or an agent that lands on this mount
// finds the book the same way.
func TestSEORootFiles(t *testing.T) {
	for path, want := range map[string]string{
		"/robots.txt":  "Sitemap: http://example.com/sitemap.xml",
		"/sitemap.xml": "<loc>http://example.com/firebook/fr/</loc>",
		"/llms.txt":    "# pofo",
		// Each edition's Atom feed rides its own mount, so this server
		// publishes them too, absolute URLs and all.
		"/firebook/fr/feed.xml": `<link rel="self" type="application/atom+xml; charset=utf-8" ` +
			`href="http://example.com/firebook/fr/feed.xml"/>`,
		"/firebook/en/feed.xml": `<title>The Quiet FIRE</title>`,
	} {
		rec := httptest.NewRecorder()
		Handler(nil, nil).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusOK {
			t.Errorf("%s: status = %d", path, rec.Code)
		}
		if !strings.Contains(rec.Body.String(), want) {
			t.Errorf("%s: misses %q", path, want)
		}
	}
}

// The IndexNow ownership key file is off unless a mount is told it is a public
// host with a key of its own.
func TestIndexNowKeyOption(t *testing.T) {
	const key = "1a2b3c4d5e6f7890"
	rec := httptest.NewRecorder()
	Handler(nil, nil).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/"+key+".txt", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("without the option: status %d, want 404", rec.Code)
	}
	rec = httptest.NewRecorder()
	Handler(nil, nil, WithIndexNowKey(key)).ServeHTTP(rec,
		httptest.NewRequest(http.MethodGet, "/"+key+".txt", nil))
	if rec.Code != http.StatusOK || rec.Body.String() != key {
		t.Errorf("key file: status %d body %q, want 200 and the key", rec.Code, rec.Body.String())
	}
}
