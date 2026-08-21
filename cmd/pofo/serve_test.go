package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/bpineau/pofo/examples"
	"github.com/bpineau/pofo/pkg/portfolio"
	"github.com/bpineau/pofo/pkg/scenario"
)

// renderCall records one fake render: the options the server layered and the
// specs it was handed. Options recording lets tests assert global overrides
// (currency=native and friends) without running the real pipeline.
type renderCall struct {
	opt   *options
	specs []*portfolio.Spec
}

// testServer returns a server whose render function is a recording fake:
// no fetching, no network, no computation. Its buildPanel returns a minimal
// non-nil panel so the FIRE caches behave as in production (a nil panel is
// deliberately not cached); tests that need the degraded path stub their own.
func testServer(t *testing.T) (*server, *[]renderCall) {
	t.Helper()
	var calls []renderCall
	s := newServer(&options{currency: "EUR", benchmark: "^GSPC", rebalance: 90}, nil)
	s.render = func(ctx context.Context, o *options, specs []*portfolio.Spec) ([]byte, error) {
		calls = append(calls, renderCall{opt: o, specs: specs})
		return []byte("<html>fake report</html>"), nil
	}
	s.buildPanel = func(ctx context.Context, spec *portfolio.Spec) (*scenario.Panel, []string) {
		// A complete build: non-nil panel with one label per holding, so it
		// is cacheable (no fetching). See panelIncomplete.
		return &scenario.Panel{}, stubLabels(spec)
	}
	return s, &calls
}

// stubLabels returns one placeholder label per holding, standing in for a
// complete FIRE panel build (labels count == holdings, so panelIncomplete is
// false).
func stubLabels(spec *portfolio.Spec) []string {
	labels := make([]string, len(spec.Holdings))
	for i, h := range spec.Holdings {
		labels[i] = h.ID
	}
	return labels
}

func serveGet(t *testing.T, h http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func TestServeRoutes(t *testing.T) {
	s, _ := testServer(t)
	h := s.handler(nil, nil)

	if rec := serveGet(t, h, "/examples/claude-dragonlite.txt"); rec.Code != 200 ||
		!strings.HasPrefix(rec.Header().Get("Content-Type"), "text/plain") ||
		!strings.Contains(rec.Body.String(), "dragon-lite") {
		t.Errorf("examples file: code=%d type=%q", rec.Code, rec.Header().Get("Content-Type"))
	}
	// The mux cleans "/examples/../secret.txt" to "/secret.txt" (which the
	// hub then 404s): a safe redirect, never a served file. Go's redirect
	// code for the cleaning has been 301, 307 and 308 across versions, so
	// accept the whole family alongside an outright 404.
	if rec := serveGet(t, h, "/examples/../secret.txt"); rec.Code != 404 &&
		rec.Code != 301 && rec.Code != 307 && rec.Code != 308 {
		t.Errorf("traversal: code=%d, want 404 (or the mux's cleaning redirect)", rec.Code)
	}
	if rec := serveGet(t, h, "/examples/nope.txt"); rec.Code != 404 {
		t.Errorf("unknown example file: code=%d", rec.Code)
	}
	if rec := serveGet(t, h, "/firebook/fr/"); rec.Code != 200 || !strings.Contains(rec.Body.String(), "book-sitenav") {
		t.Errorf("book: code=%d, navbar wanted", rec.Code)
	}
	// Both editions are mounted, and each is told where the other one sits:
	// the indexes cross-link, so the hreflang pair reaches the served page.
	if rec := serveGet(t, h, "/firebook/en/"); rec.Code != 200 || !strings.Contains(rec.Body.String(), "The Quiet FIRE") {
		t.Errorf("book-en: code=%d, The Quiet FIRE wanted", rec.Code)
	}
	// The hreflang hrefs are fully qualified, built from the request's own
	// origin (httptest's default host).
	for path, want := range map[string]string{
		"/firebook/fr/": `<link rel="alternate" hreflang="en" href="http://example.com/firebook/en/">`,
		"/firebook/en/": `<link rel="alternate" hreflang="fr" href="http://example.com/firebook/fr/">`,
	} {
		if rec := serveGet(t, h, path); !strings.Contains(rec.Body.String(), want) {
			t.Errorf("%s: misses the cross-edition link %q", path, want)
		}
	}
	// The old /book/ path permanently redirects to /firebook/.
	if rec := serveGet(t, h, "/book/fr/"); rec.Code != http.StatusMovedPermanently ||
		rec.Header().Get("Location") != "/firebook/fr/" {
		t.Errorf("book redirect: code=%d loc=%q, want 301 to /firebook/fr/", rec.Code, rec.Header().Get("Location"))
	}
	if rec := serveGet(t, h, "/theme.css"); !strings.HasPrefix(rec.Header().Get("Content-Type"), "text/css") {
		t.Error("theme.css content type")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/theme.css", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST theme.css: code=%d, want 405", rec.Code)
	}
	if rec := serveGet(t, h, "/no-such-page"); rec.Code != 404 {
		t.Errorf("unknown path: code=%d, want 404", rec.Code)
	}
	if rec := serveGet(t, h, "/firesimulator/"); rec.Code != 200 {
		t.Errorf("fire mount: code=%d", rec.Code)
	}
	// The old /fire/ path permanently redirects to /firesimulator/, preserving
	// the sub-path and query so existing bookmarks and shared links survive.
	if rec := serveGet(t, h, "/fire/p/IWDA:100/?x=1"); rec.Code != 301 ||
		rec.Header().Get("Location") != "/firesimulator/p/IWDA:100/?x=1" {
		t.Errorf("fire redirect: code=%d loc=%q, want 301 to /firesimulator/p/IWDA:100/?x=1", rec.Code, rec.Header().Get("Location"))
	}
	// The machine-readable root files: the sitemap covers both editions and
	// the app surfaces, robots.txt points at it, llms.txt indexes the book.
	if rec := serveGet(t, h, "/sitemap.xml"); rec.Code != 200 ||
		!strings.Contains(rec.Body.String(), "<loc>http://example.com/firebook/en/</loc>") ||
		!strings.Contains(rec.Body.String(), "<loc>http://example.com/visualizer</loc>") {
		t.Errorf("sitemap: code=%d, both editions and the app surfaces wanted", rec.Code)
	}
	if rec := serveGet(t, h, "/robots.txt"); rec.Code != 200 ||
		!strings.Contains(rec.Body.String(), "Sitemap: http://example.com/sitemap.xml") ||
		!strings.Contains(rec.Body.String(), "User-agent: ClaudeBot") {
		t.Errorf("robots.txt: code=%d", rec.Code)
	}
	if rec := serveGet(t, h, "/llms.txt"); rec.Code != 200 ||
		!strings.Contains(rec.Body.String(), "http://example.com/firebook/fr/fire-cest-quoi.md") ||
		!strings.Contains(rec.Body.String(), "http://example.com/firebook/en/feed.xml") {
		t.Errorf("llms.txt: code=%d", rec.Code)
	}
	// Each edition's Atom feed, on its own mount, with absolute URLs.
	for lang, title := range map[string]string{"fr": "Le FIRE tranquille", "en": "The Quiet FIRE"} {
		path := "/firebook/" + lang + "/feed.xml"
		rec := serveGet(t, h, path)
		if rec.Code != 200 || !strings.HasPrefix(rec.Header().Get("Content-Type"), "application/atom+xml") {
			t.Errorf("%s: code=%d type=%q", path, rec.Code, rec.Header().Get("Content-Type"))
		}
		if !strings.Contains(rec.Body.String(), "<title>"+title+"</title>") ||
			!strings.Contains(rec.Body.String(), `href="http://example.com`+path+`"`) {
			t.Errorf("%s: the feed does not name itself absolutely", path)
		}
	}
	// And the markdown mirror an agent is pointed at is really there.
	if rec := serveGet(t, h, "/firebook/fr/fire-cest-quoi.md"); rec.Code != 200 ||
		!strings.HasPrefix(rec.Header().Get("Content-Type"), "text/markdown") {
		t.Errorf("markdown mirror: code=%d type=%q", rec.Code, rec.Header().Get("Content-Type"))
	}
	// A per-example FIRE mount for an unknown name 404s before any panel
	// build (the known-name build path fetches quotes and is exercised by the
	// manual smoke test, not here).
	if rec := serveGet(t, h, "/firesimulator/e/nope/"); rec.Code != 404 {
		t.Errorf("unknown fire example: code=%d, want 404", rec.Code)
	}
}

// TestServeLanding locks the front door: the two-tone mark, the four section
// cards, and the visualizer's canonical path (with the trailing-slash form
// redirected, and the old front-door content now living there).
func TestServeLanding(t *testing.T) {
	s, _ := testServer(t)
	h := s.handler(nil, nil)

	rec := serveGet(t, h, "/")
	if rec.Code != 200 {
		t.Fatalf("landing: code=%d", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{
		`<h1 class="land-mark">po<b>fo</b></h1>`,
		`href="/firebook/fr/"`, `href="/firebook/en/"`,
		`href="/visualizer"`, `href="/firesimulator/"`,
		`Fire Book (fr)`, `Fire Book (en)`, `Portfolio visualizer`, `Fire Simulator`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("landing missing %q", want)
		}
	}

	if rec := serveGet(t, h, "/visualizer/"); rec.Code != 301 ||
		rec.Header().Get("Location") != "/visualizer" {
		t.Errorf("visualizer slash: code=%d loc=%q, want 301 to /visualizer", rec.Code, rec.Header().Get("Location"))
	}
	viz := serveGet(t, h, "/visualizer")
	if viz.Code != 200 || !strings.Contains(viz.Body.String(), "Put portfolios side by side.") {
		t.Errorf("visualizer: code=%d, hub headline wanted", viz.Code)
	}
	// The visualizer's mark links back to the landing page.
	if !strings.Contains(viz.Body.String(), `<a class="hub-mark" href="/">po<b>fo</b></a>`) {
		t.Error("visualizer mark must link home")
	}
}

func TestServeComposerAssets(t *testing.T) {
	s, _ := testServer(t)
	h := s.handler(nil, nil)
	if rec := serveGet(t, h, "/composer.js"); rec.Code != 200 ||
		!strings.HasPrefix(rec.Header().Get("Content-Type"), "text/javascript") {
		t.Errorf("composer.js: code=%d type=%q", rec.Code, rec.Header().Get("Content-Type"))
	}
	if rec := serveGet(t, h, "/composer.css"); rec.Code != 200 ||
		!strings.HasPrefix(rec.Header().Get("Content-Type"), "text/css") {
		t.Errorf("composer.css: code=%d type=%q", rec.Code, rec.Header().Get("Content-Type"))
	}
	// GET-only, like the other embedded assets.
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/composer.js", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST composer.js: code=%d, want 405", rec.Code)
	}
}

// Static assets are content-fingerprinted so a deploy changes their URL and
// Cloudflare's edge cache cannot serve stale CSS/JS: the HTML links carry a
// ?v=<hash>, and a versioned request is cacheable-immutable (the URL is the
// version). A bare request stays uncached so it can never pin an old copy.
func TestServeAssetFingerprinting(t *testing.T) {
	s, _ := testServer(t)
	h := s.handler(nil, nil)
	hub := serveGet(t, h, "/visualizer").Body.String()
	for _, want := range []string{
		`href="/theme.css?v=`, `href="/fonts.css?v=`,
		`href="/composer.css?v=`, `src="/composer.js?v=`,
	} {
		if !strings.Contains(hub, want) {
			t.Errorf("hub missing fingerprinted asset %q", want)
		}
	}
	if got := serveGet(t, h, "/composer.js?v=abc123").Header().Get("Cache-Control"); !strings.Contains(got, "immutable") {
		t.Errorf("versioned asset Cache-Control = %q, want immutable", got)
	}
	if got := serveGet(t, h, "/composer.js").Header().Get("Cache-Control"); got != "" {
		t.Errorf("bare asset Cache-Control = %q, want none", got)
	}
}

func TestServeCatalogJSON(t *testing.T) {
	s, _ := testServer(t)
	h := s.handler(nil, nil)

	rec := serveGet(t, h, "/catalog.json")
	if rec.Code != 200 ||
		!strings.HasPrefix(rec.Header().Get("Content-Type"), "application/json") ||
		rec.Header().Get("Cache-Control") != "public, max-age=3600" {
		t.Fatalf("catalog.json: code=%d type=%q cache=%q", rec.Code,
			rec.Header().Get("Content-Type"), rec.Header().Get("Cache-Control"))
	}
	var entries []struct {
		ID   string   `json:"id"`
		Name string   `json:"name"`
		Alt  []string `json:"alt"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &entries); err != nil {
		t.Fatalf("not JSON: %v", err)
	}
	if len(entries) < 50 {
		t.Fatalf("suspiciously small: %d entries", len(entries))
	}

	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, httptest.NewRequest(http.MethodPost, "/catalog.json", nil))
	if rec2.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST catalog.json: code=%d, want 405", rec2.Code)
	}
}

func TestServeFireComposed(t *testing.T) {
	s, _ := testServer(t)
	h := s.handler(nil, nil)

	// A valid composed spec mounts the simulator app shell.
	rec := serveGet(t, h, "/firesimulator/p/IWDA:60,IGLN:40!sim:on/")
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "<html") {
		t.Fatalf("composed fire: code=%d", rec.Code)
	}

	// The handler is cached under the raw spec key.
	if len(s.fireBySpec) != 1 {
		t.Errorf("cache size = %d, want 1", len(s.fireBySpec))
	}
	serveGet(t, h, "/firesimulator/p/IWDA:60,IGLN:40!sim:on/")
	if len(s.fireBySpec) != 1 {
		t.Errorf("cache size after rehit = %d, want 1", len(s.fireBySpec))
	}

	// Naked (no trailing slash): redirect to the directory form, no build.
	if rec := serveGet(t, h, "/firesimulator/p/IWDA:100"); rec.Code != 301 ||
		rec.Header().Get("Location") != "/firesimulator/p/IWDA:100/" {
		t.Errorf("naked composed: code=%d loc=%q", rec.Code, rec.Header().Get("Location"))
	}

	// Catalog gate and grammar errors: 400, never a panel build.
	if rec := serveGet(t, h, "/firesimulator/p/ZZZNOTANID:100/"); rec.Code != 400 ||
		!strings.Contains(rec.Body.String(), "not in the local catalog") {
		t.Errorf("catalog gate: code=%d", rec.Code)
	}
	if rec := serveGet(t, h, "/firesimulator/p/garbage/"); rec.Code != 400 {
		t.Errorf("malformed spec: code=%d", rec.Code)
	}

	// A percent-encoded slash cannot cross the segment boundary: the spec
	// grammar has no "/" so it must 400 as a malformed holding, not route.
	req := httptest.NewRequest(http.MethodGet, "/firesimulator/p/IWDA:100%2Fapi/", nil)
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req)
	if rec2.Code != 400 {
		t.Errorf("escaped slash: code=%d, want 400", rec2.Code)
	}

	// The cache is bounded: distinct specs evict, never grow past the cap.
	for i := 0; i <= fireSpecCacheMax+3; i++ {
		serveGet(t, h, fmt.Sprintf("/firesimulator/p/IWDA:%d,IGLN:%d/", 50+i, 50-i))
	}
	if len(s.fireBySpec) > fireSpecCacheMax {
		t.Errorf("cache size = %d, want <= %d", len(s.fireBySpec), fireSpecCacheMax)
	}
}

// Every -serve FIRE mount announces where it runs and how to load a
// portfolio: the provenance pill and the drawer's loader are both built from
// /api/meta, so the wiring is asserted there.
func TestServeFireMetaSourceAndPicker(t *testing.T) {
	s, _ := testServer(t)
	s.sourceLabel = "startup-portfolio"
	h := s.handler(nil, nil)

	meta := func(path string) map[string]any {
		t.Helper()
		rec := serveGet(t, h, path)
		if rec.Code != 200 {
			t.Fatalf("%s: code=%d", path, rec.Code)
		}
		var m map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
			t.Fatalf("%s: bad json: %v", path, err)
		}
		return m
	}

	for path, want := range map[string]string{
		"/firesimulator/api/meta":                          "startup-portfolio",
		"/firesimulator/e/claude-dragonlite/api/meta":      "claude-dragonlite",
		"/firesimulator/p/IWDA:60,IGLN:40!sim:on/api/meta": "custom portfolio",
	} {
		m := meta(path)
		if m["sourceLabel"] != want {
			t.Errorf("%s: sourceLabel = %v, want %q", path, m["sourceLabel"], want)
		}
		p, ok := m["picker"].(map[string]any)
		if !ok {
			t.Fatalf("%s: no picker in meta", path)
		}
		if p["base"] != fireBase || p["catalogURL"] != "/catalog.json" || p["viewURL"] != "/view" {
			t.Errorf("%s: picker mount = %v", path, p)
		}
		exs, _ := p["examples"].([]any)
		if len(exs) != len(examples.List()) {
			t.Errorf("%s: %d examples, want %d", path, len(exs), len(examples.List()))
		}
	}
}

func TestServeFireExampleNakedRedirect(t *testing.T) {
	s, _ := testServer(t)
	h := s.handler(nil, nil)
	if rec := serveGet(t, h, "/firesimulator/e/claude-dragonlite"); rec.Code != 301 ||
		rec.Header().Get("Location") != "/firesimulator/e/claude-dragonlite/" {
		t.Errorf("naked example: code=%d loc=%q", rec.Code, rec.Header().Get("Location"))
	}
	// Unknown names still 404, redirect or not.
	if rec := serveGet(t, h, "/firesimulator/e/nope"); rec.Code != 404 {
		t.Errorf("naked unknown example: code=%d, want 404", rec.Code)
	}
}

func TestServeView(t *testing.T) {
	s, calls := testServer(t)
	h := s.handler(nil, nil)

	rec := serveGet(t, h, "/view?ex=claude-dragonlite&rebalance=180")
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "fake report") {
		t.Fatalf("view: code=%d body=%q", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Error("view must be no-store")
	}
	if len(*calls) != 1 || len((*calls)[0].specs) != 1 || (*calls)[0].specs[0].Name != "claude-dragonlite" {
		t.Errorf("render calls = %+v", *calls)
	}

	if rec := serveGet(t, h, "/view"); rec.Code != 303 || rec.Header().Get("Location") != "/visualizer" {
		t.Errorf("empty view: code=%d loc=%q, want 303 to /visualizer", rec.Code, rec.Header().Get("Location"))
	}
	if rec := serveGet(t, h, "/view?p=ZZZNOTANID:100"); rec.Code != 400 ||
		!strings.Contains(rec.Body.String(), "not in the local catalog") {
		t.Errorf("catalog gate: code=%d", rec.Code)
	}
}

func TestServeViewCurrencyNative(t *testing.T) {
	s, calls := testServer(t)
	h := s.handler(nil, nil)

	// currency=native renders with an empty currency override (keep native
	// currencies), not the EUR server default.
	rec := serveGet(t, h, "/view?ex=claude-dragonlite&currency=native")
	if rec.Code != 200 {
		t.Fatalf("view: code=%d", rec.Code)
	}
	if len(*calls) != 1 || (*calls)[0].opt.currency != "" {
		t.Errorf("rendered currency = %q, want empty (native)", (*calls)[0].opt.currency)
	}

	// The cookie stores the native choice as the empty ISO code (the codec's
	// internal representation).
	var stored *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == prefsCookie {
			stored = c
		}
	}
	if stored == nil {
		t.Fatal("no pofo_prefs cookie set")
	}
	if stored.Value != "currency=" {
		t.Errorf("cookie value = %q, want exactly currency= (native as the empty code)", stored.Value)
	}

	// Feeding that cookie back to the hub re-selects the native option.
	req := httptest.NewRequest(http.MethodGet, "/visualizer", nil)
	req.AddCookie(stored)
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req)
	if !strings.Contains(rec2.Body.String(), `value="native" selected`) {
		t.Error("hub did not re-select the native currency option")
	}
	// The hub's Open links carry the native sentinel, never a bare currency=.
	if !strings.Contains(rec2.Body.String(), "currency=native") {
		t.Error("hub Open links must carry currency=native")
	}
}

func TestServeHubOutOfListPref(t *testing.T) {
	s, _ := testServer(t)
	h := s.handler(nil, nil)

	// A stored rebalance outside the hardcoded list must appear as its own
	// selected option, never silently rewritten by a lying select.
	req := httptest.NewRequest(http.MethodGet, "/visualizer", nil)
	req.AddCookie(&http.Cookie{Name: prefsCookie, Value: "rebalance=7"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if !strings.Contains(rec.Body.String(), `value="7" selected`) {
		t.Error("hub must surface an out-of-list rebalance as its own selected option")
	}
}

func TestFireForSpecPanelCaching(t *testing.T) {
	spec, err := adhocSpec("IWDA:60,IGLN:40", 1)
	if err != nil {
		t.Fatal(err)
	}

	// A degraded build (nil panel) is served but never cached: a transient
	// failure must not freeze a permanently degraded simulator.
	s, _ := testServer(t)
	s.buildPanel = func(ctx context.Context, spec *portfolio.Spec) (*scenario.Panel, []string) {
		return nil, nil
	}
	if h := s.fireForSpec(context.Background(), "k", spec); h == nil {
		t.Fatal("a degraded build must still serve this request")
	}
	if len(s.fireBySpec) != 0 {
		t.Errorf("nil-panel cache size = %d, want 0 (not cached)", len(s.fireBySpec))
	}

	// A partial build (non-nil panel but a holding dropped by a transient
	// fetch failure, so fewer labels than holdings) is served but not cached:
	// its composition is wrong and must not be frozen in.
	sp, _ := testServer(t)
	sp.buildPanel = func(ctx context.Context, spec *portfolio.Spec) (*scenario.Panel, []string) {
		return &scenario.Panel{}, []string{spec.Holdings[0].ID} // 1 label, spec has 2 holdings
	}
	if h := sp.fireForSpec(context.Background(), "k", spec); h == nil {
		t.Fatal("a partial build must still serve this request")
	}
	if len(sp.fireBySpec) != 0 {
		t.Errorf("partial-panel cache size = %d, want 0 (not cached)", len(sp.fireBySpec))
	}

	// A healthy build (non-nil panel, a label per holding) is cached, and two
	// sequential cold hits build exactly once.
	s2, _ := testServer(t)
	builds := 0
	s2.buildPanel = func(ctx context.Context, spec *portfolio.Spec) (*scenario.Panel, []string) {
		builds++
		return &scenario.Panel{}, stubLabels(spec)
	}
	s2.fireForSpec(context.Background(), "k", spec)
	s2.fireForSpec(context.Background(), "k", spec)
	if len(s2.fireBySpec) != 1 {
		t.Errorf("non-nil-panel cache size = %d, want 1", len(s2.fireBySpec))
	}
	if builds != 1 {
		t.Errorf("builds = %d, want 1 (second cold hit must reuse the cache)", builds)
	}
}

func TestServeHub(t *testing.T) {
	s, _ := testServer(t)
	h := s.handler(nil, nil)
	rec := serveGet(t, h, "/visualizer")
	if rec.Code != 200 {
		t.Fatalf("hub: code=%d", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{
		`action="/view"`, `method="get"`,
		`name="ex" value="dragon-decumulation-household"`,
		`href="/examples/dragon-decumulation-household.txt"`,
		`href="/view?ex=dragon-decumulation-household"`,
		`href="/firesimulator/e/dragon-decumulation-household/"`,
		`href="/firesimulator/"`, `href="/firebook/fr/"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("hub missing %q", want)
		}
	}
	// The warm book skin is applied (a book paper token, not the instrument
	// default).
	if !strings.Contains(body, "#faf6ef") {
		t.Error("hub missing the warm book skin")
	}
	// The examples catalog is now folded under the composer.
	if !strings.Contains(body, `<details class="hub-examples"`) {
		t.Error("hub examples must be folded into a <details>")
	}
}

// TestServeHubComposer locks the composer-first front door: the empty composer
// mount with its blank-boot and asset links, and the embedded presets (valid
// JSON whose p= round-trips through adhocSpec, unforkable examples excluded).
func TestServeHubComposer(t *testing.T) {
	s, _ := testServer(t)
	h := s.handler(nil, nil)
	body := serveGet(t, h, "/visualizer").Body.String()

	for _, want := range []string{
		`id="composer"`, `data-boot="blank"`, `data-globals=`,
		// Asset URLs are content-fingerprinted (…?v=<hash>); match the prefix.
		`href="/composer.css?v=`, `src="/composer.js?v=`, `data-preset-count=`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("hub composer missing %q", want)
		}
	}

	// Every embedded preset is valid JSON whose p= the server would accept.
	re := regexp.MustCompile(`data-preset-\d+="([^"]*)"`)
	matches := re.FindAllStringSubmatch(body, -1)
	if len(matches) < 40 {
		t.Fatalf("only %d presets embedded, want >= 40", len(matches))
	}
	for _, m := range matches {
		var p composerPreset
		if err := json.Unmarshal([]byte(html.UnescapeString(m[1])), &p); err != nil {
			t.Fatalf("preset JSON %q: %v", m[1], err)
		}
		if holdings, _, _ := strings.Cut(p.P, "!"); holdings == "" {
			t.Errorf("preset %q is unforkable and must have been excluded", p.Name)
		}
		if _, err := adhocSpec(p.P, 1); err != nil {
			t.Errorf("preset %q p=%q does not round-trip through adhocSpec: %v", p.Name, p.P, err)
		}
	}
}

func TestServeHubPrefs(t *testing.T) {
	s, _ := testServer(t)
	h := s.handler(nil, nil)

	// Without a cookie: server defaults selected, bare Open links.
	body := serveGet(t, h, "/visualizer").Body.String()
	for _, want := range []string{
		`name="currency"`, `name="rebalance"`, `name="sim"`,
		`value="EUR" selected`, `value="90" selected`, `value="on" selected`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("hub defaults missing %q", want)
		}
	}
	if !strings.Contains(body, `href="/view?ex=claude-dragonlite"`) {
		t.Error("bare Open link expected without a cookie")
	}

	// With a cookie: controls pre-selected and Open links carry the prefs.
	req := httptest.NewRequest(http.MethodGet, "/visualizer", nil)
	req.AddCookie(&http.Cookie{Name: prefsCookie, Value: "currency=USD&rebalance=30&sim=off"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	body = rec.Body.String()
	for _, want := range []string{
		`value="USD" selected`, `value="30" selected`, `value="off" selected`,
		`href="/view?ex=claude-dragonlite&amp;currency=USD&amp;rebalance=30&amp;sim=off"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("hub with cookie missing %q", want)
		}
	}
}

func TestServeViewSetsPrefsCookie(t *testing.T) {
	s, _ := testServer(t)
	h := s.handler(nil, nil)

	// Explicit globals set the cookie.
	rec := serveGet(t, h, "/view?ex=claude-dragonlite&currency=USD&rebalance=30")
	cookies := rec.Result().Cookies()
	var got *http.Cookie
	for _, c := range cookies {
		if c.Name == prefsCookie {
			got = c
		}
	}
	if got == nil {
		t.Fatal("no pofo_prefs cookie set")
	}
	if !strings.Contains(got.Value, "currency=USD") || !strings.Contains(got.Value, "rebalance=30") {
		t.Errorf("cookie value = %q", got.Value)
	}

	// A partial URL merges with the stored cookie instead of erasing it.
	req := httptest.NewRequest(http.MethodGet, "/view?ex=claude-dragonlite&sim=off", nil)
	req.AddCookie(&http.Cookie{Name: prefsCookie, Value: got.Value})
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req)
	var merged *http.Cookie
	for _, c := range rec2.Result().Cookies() {
		if c.Name == prefsCookie {
			merged = c
		}
	}
	if merged == nil {
		t.Fatal("no merged cookie set")
	}
	for _, want := range []string{"currency=USD", "rebalance=30", "sim=off"} {
		if !strings.Contains(merged.Value, want) {
			t.Errorf("merged cookie %q missing %q", merged.Value, want)
		}
	}

	// No explicit globals: no Set-Cookie at all.
	rec3 := serveGet(t, h, "/view?ex=claude-dragonlite")
	if len(rec3.Result().Cookies()) != 0 {
		t.Error("bare /view must not set a cookie")
	}
}

// The nested FIRE mounts are Embedded: they serve the app, never a second
// copy of the book or of the site files. A crawler was seen wandering into
// /firesimulator/e/<name>/firebook/..., which used to answer 200.
func TestFireMountsDoNotRepublishTheSite(t *testing.T) {
	s, _ := testServer(t)
	h := s.handler(nil, nil)
	name := ""
	for n := range s.examples {
		name = n
		break
	}
	if name == "" {
		t.Fatal("no known examples")
	}
	for _, path := range []string{
		"/firesimulator/firebook/fr/",
		"/firesimulator/sitemap.xml",
		"/firesimulator/e/" + name + "/firebook/fr/echelle-obligataire",
		"/firesimulator/e/" + name + "/sitemap.xml",
	} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusNotFound {
			t.Errorf("%s: status %d, want 404", path, rec.Code)
		}
	}
	for _, path := range []string{"/firesimulator/", "/firesimulator/e/" + name + "/", "/firebook/fr/"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusOK {
			t.Errorf("%s: status %d, want 200", path, rec.Code)
		}
	}
}

// The liveness endpoint answers plain text, says nothing about the market data,
// and refuses anything but a read.
func TestServeHealthz(t *testing.T) {
	s, _ := testServer(t)
	h := s.handler(nil, nil)
	rec := serveGet(t, h, healthPath)
	if rec.Code != http.StatusOK ||
		!strings.HasPrefix(rec.Header().Get("Content-Type"), "text/plain") ||
		strings.TrimSpace(rec.Body.String()) != "ok" {
		t.Errorf("healthz: code=%d type=%q body=%q", rec.Code,
			rec.Header().Get("Content-Type"), rec.Body.String())
	}
	post := httptest.NewRecorder()
	h.ServeHTTP(post, httptest.NewRequest(http.MethodPost, healthPath, nil))
	if post.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST healthz: code=%d, want 405", post.Code)
	}
	// It is a probe, not a page: nothing points at it.
	for _, path := range []string{"/sitemap.xml", "/robots.txt", "/llms.txt"} {
		if body := serveGet(t, h, path).Body.String(); strings.Contains(body, healthPath) {
			t.Errorf("%s advertises %s", path, healthPath)
		}
	}
}

// The container's HEALTHCHECK polls every 30 seconds; that must leave no trace
// in the access log, while ordinary traffic still logs one line per request.
func TestLogAccessSkipsHealthz(t *testing.T) {
	s, _ := testServer(t)
	var buf bytes.Buffer
	h := logAccess(&buf, s.handler(nil, nil))

	for range 3 {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, healthPath, nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("healthz through the access log: code=%d", rec.Code)
		}
	}
	if buf.Len() != 0 {
		t.Errorf("the probe logged %q", buf.String())
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	got := buf.String()
	if n := strings.Count(got, "\n"); n != 1 || !strings.Contains(got, `"GET / HTTP/1.1" 200`) {
		t.Errorf("landing page: %d access-log line(s): %q", n, got)
	}
}
