package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bpineau/pofo/pkg/scenario"
)

// simEndpoints are every POST endpoint the page fires per render, each with
// one field of its own payload the front end reads, so a silently renamed or
// emptied response is caught here rather than in the browser.
var simEndpoints = map[string]string{
	"/api/sim":            "arbitrageSvg",
	"/api/models":         "models",
	"/api/paths":          "fans",
	"/api/market":         "fans",
	"/api/sensitivity":    "sensitivitySvg",
	"/api/frontier":       "frontierSvg",
	"/api/policyfrontier": "policyFrontierSvg",
	"/api/solvemenu":      "options",
	"/api/solve":          "requiredCapital",
	"/api/spending":       "spendingSvg",
	"/api/decade":         "decadeSvg",
	"/api/vintages":       "vintagesSvg",
	"/api/income":         "incomeSvg",
	"/api/lifecycle":      "lifeSvg",
	"/api/curves":         "horizonSvg",
}

// tinyParams is a plan small enough to simulate a dozen times per test and
// rich enough to reach every block of every endpoint: a real horizon, a
// pension, a side income and a cash buffer.
func tinyParams() Params {
	return Params{
		Capital: 800_000, NeedAnnual: 32_000, BufferYears: 2,
		Mu: 0.04, Sigma: 0.12, Df: 5, Years: 8, NPaths: 60, TaxRate: 0.30,
		PensionAnnual: 9_000, PensionYear: 5, SideAnnual: 6_000, SideUntilYear: 3,
	}
}

// post fires one endpoint and returns the decoded JSON object.
func post(t *testing.T, h http.Handler, path string, pr Params) map[string]any {
	t.Helper()
	body, err := json.Marshal(pr)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body)))
	if rec.Code != http.StatusOK {
		t.Fatalf("%s: status %d: %s", path, rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("%s: content type %q", path, ct)
	}
	var out map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("%s: bad json: %v", path, err)
	}
	return out
}

// Every analysis endpoint answers the same shape on the same params: a JSON
// object carrying its own payload field. The page fires all of them at once,
// so one broken route is a hole in the render, not a missing page.
func TestEveryEndpointAnswers(t *testing.T) {
	h := Handler(nil, nil)
	for path, field := range simEndpoints {
		t.Run(strings.TrimPrefix(path, "/api/"), func(t *testing.T) {
			out := post(t, h, path, tinyParams())
			if _, ok := out[field]; !ok {
				keys := make([]string, 0, len(out))
				for k := range out {
					keys = append(keys, k)
				}
				t.Errorf("no %q in the response; keys = %v", field, keys)
			}
			// Note is a legend on some endpoints and a caveat on others,
			// so only the caveats are asserted against: a well-formed plan
			// must never be turned away for want of inputs or history.
			if note, _ := out["note"].(string); strings.HasPrefix(note, "set a") ||
				strings.HasPrefix(note, "Not enough") || strings.HasPrefix(note, "no ") {
				t.Errorf("well-formed plan refused: %q", note)
			}
		})
	}
}

// A panel switches on the data-driven columns and the monthly kernel; every
// endpoint must survive that second wiring too, including the two panel-only
// models reached through the monthly source.
func TestEveryEndpointAnswersWithAPanel(t *testing.T) {
	panel := monthlyPanel(600) // 50 years of monthly history
	h := Handler(&panel, []string{"equity"})
	for _, central := range []string{"", "boot", "hist"} {
		pr := tinyParams()
		pr.Monthly, pr.Central, pr.Weights = true, central, []float64{1}
		for path, field := range simEndpoints {
			t.Run(strings.TrimPrefix(path, "/api/")+"-"+central, func(t *testing.T) {
				if _, ok := post(t, h, path, pr)[field]; !ok {
					t.Errorf("%s: no %q in the response", path, field)
				}
			})
		}
	}
}

// The error shapes of the shared POST boilerplate: only POST is answered, a
// body that is not a Params is refused before any slot is taken, and neither
// path ever renders a partial JSON result.
func TestEndpointErrorShapes(t *testing.T) {
	h := Handler(nil, nil)
	for path := range simEndpoints {
		t.Run(strings.TrimPrefix(path, "/api/"), func(t *testing.T) {
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
			if rec.Code != http.StatusMethodNotAllowed ||
				!strings.Contains(rec.Body.String(), "POST only") {
				t.Errorf("GET: status %d body %q, want 405", rec.Code, rec.Body.String())
			}
			rec = httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, path,
				strings.NewReader(`{"capital":`)))
			if rec.Code != http.StatusBadRequest {
				t.Errorf("truncated body: status %d, want 400", rec.Code)
			}
			rec = httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, path,
				strings.NewReader(`{"capital":"a lot"}`)))
			if rec.Code != http.StatusBadRequest {
				t.Errorf("mistyped field: status %d, want 400", rec.Code)
			}
		})
	}
}

// /api/fit re-estimates the parametric model from live slider weights. It
// needs a portfolio to fit, refuses a body it cannot read, and answers an
// EMPTY object rather than zeros when the fit is degenerate: mu=0/sigma=0
// would seed the page with a certain-doom market.
func TestAPIFit(t *testing.T) {
	rec := httptest.NewRecorder()
	Handler(nil, nil).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/fit",
		strings.NewReader(`{"weights":[1]}`)))
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "no portfolio") {
		t.Errorf("without a panel: status %d body %q", rec.Code, rec.Body.String())
	}

	panel := monthlyPanel(240)
	h := Handler(&panel, []string{"equity"})
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/fit", strings.NewReader(`{`)))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("truncated body: status %d, want 400", rec.Code)
	}

	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/fit",
		strings.NewReader(`{"weights":[1]}`)))
	var fit map[string]float64
	if err := json.Unmarshal(rec.Body.Bytes(), &fit); err != nil {
		t.Fatalf("bad json: %v", err)
	}
	if fit["sigma"] <= 0 {
		t.Errorf("fit = %v, want a positive sigma", fit)
	}

	// A one-month panel cannot be fitted; the response is empty, not zeroed.
	short := scenario.Panel{Returns: [][]float64{{0.01}}, Weights: []float64{1}}
	rec = httptest.NewRecorder()
	Handler(&short, []string{"equity"}).ServeHTTP(rec,
		httptest.NewRequest(http.MethodPost, "/api/fit", strings.NewReader(`{"weights":[1]}`)))
	if got := strings.TrimSpace(rec.Body.String()); got != "{}" {
		t.Errorf("degenerate fit answered %q, want an empty object", got)
	}
}

// With a portfolio, /api/meta carries the allocation and the fitted sliders;
// with one too short to fit, it carries the allocation and no figures, so the
// UI keeps its own defaults instead of a doom market.
func TestMetaWithPanel(t *testing.T) {
	// json.Unmarshal merges into a non-empty map, so each mount's meta is
	// read into a fresh one: what is ABSENT is half of what is asserted here.
	metaWith := func(p *scenario.Panel) map[string]any {
		t.Helper()
		rec := httptest.NewRecorder()
		Handler(p, []string{"equity"}).ServeHTTP(rec,
			httptest.NewRequest(http.MethodGet, "/api/meta", nil))
		out := map[string]any{}
		if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
			t.Fatalf("bad json: %v", err)
		}
		return out
	}
	if meta := metaWith(nil); meta["hasPanel"] != false || meta["weights"] != nil {
		t.Errorf("no panel: hasPanel = %v, weights = %v", meta["hasPanel"], meta["weights"])
	}

	panel := monthlyPanel(240)
	meta := metaWith(&panel)
	if meta["hasPanel"] != true || meta["panelMonths"] != float64(240) {
		t.Errorf("panel meta = %v / %v", meta["hasPanel"], meta["panelMonths"])
	}
	for _, k := range []string{"mu", "sigma", "df", "weights"} {
		if _, ok := meta[k]; !ok {
			t.Errorf("no %q in the panel meta", k)
		}
	}

	// A one-month panel cannot be fitted: the allocation still ships (the
	// sliders need it) but the market figures do not, so the UI keeps its own
	// defaults rather than a mu=0/sigma=0 certain-doom market.
	short := metaWith(&scenario.Panel{Returns: [][]float64{{0.01}}, Weights: []float64{1}})
	if _, ok := short["weights"]; !ok {
		t.Errorf("the allocation must ship even when the fit does not")
	}
	for _, k := range []string{"mu", "sigma", "df"} {
		if v, ok := short[k]; ok {
			t.Errorf("%q shipped from a degenerate fit: %v", k, v)
		}
	}
}

// The page's own asset URLs carry a content fingerprint, and a fingerprinted
// request is cacheable for a year: the URL changes whenever the bytes do. The
// unversioned URL must NOT be, or an edge cache pins a stale app.js forever.
func TestVersionedAssetsAreImmutable(t *testing.T) {
	h := Handler(nil, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if !strings.Contains(rec.Body.String(), `src="app.js?v=`) ||
		!strings.Contains(rec.Body.String(), `href="theme.css?v=`) {
		t.Errorf("the index page does not fingerprint its assets")
	}
	const immutable = "public, max-age=31536000, immutable"
	for _, path := range []string{"/app.js", "/app.css", "/theme.css", "/fonts.css"} {
		rec = httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path+"?v=abc123", nil))
		if rec.Code != http.StatusOK || rec.Header().Get("Cache-Control") != immutable {
			t.Errorf("%s?v=: status %d cache %q", path, rec.Code, rec.Header().Get("Cache-Control"))
		}
		rec = httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusOK || rec.Header().Get("Cache-Control") == immutable {
			t.Errorf("%s unversioned: status %d cache %q", path, rec.Code, rec.Header().Get("Cache-Control"))
		}
	}
	// The favicon is the same bytes at both paths, cached for a day.
	for _, path := range []string{"/favicon.svg", "/favicon.ico"} {
		rec = httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusOK || rec.Header().Get("Content-Type") != "image/svg+xml" {
			t.Errorf("%s: status %d type %q", path, rec.Code, rec.Header().Get("Content-Type"))
		}
	}
	// An asset that does not exist is a 404, never the index page.
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/nope.js", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("missing asset: status %d, want 404", rec.Code)
	}
}

// WithNav splices the sibling links into the page's top bar, escaped; without
// it the placeholder vanishes rather than leaving an empty nav element.
func TestTopnavRendered(t *testing.T) {
	rec := httptest.NewRecorder()
	Handler(nil, nil, WithNav([]NavLink{
		{Label: "Portfolios", Href: "/visualizer"},
		{Label: `Book & "notes"`, Href: "/firebook/fr/?a=1&b=2"},
	})).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	body := rec.Body.String()
	for _, want := range []string{
		`<nav class="topnav">`,
		`<a href="/visualizer">Portfolios</a>`,
		`<a href="/firebook/fr/?a=1&amp;b=2">Book &amp; &#34;notes&#34;</a>`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("the top bar misses %q", want)
		}
	}
	rec = httptest.NewRecorder()
	Handler(nil, nil).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if strings.Contains(rec.Body.String(), "topnav") ||
		strings.Contains(rec.Body.String(), "<!--topnav-->") {
		t.Errorf("a mount with no siblings must render no nav and no placeholder")
	}
}
