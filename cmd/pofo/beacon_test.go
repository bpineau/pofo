package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bpineau/pofo/pkg/webui"
)

// cfTestToken stands in for a Cloudflare Web Analytics site token.
const cfTestToken = "0123456789abcdef0123456789abcdef"

// beaconServer is testServer with an analytics token set, so s.handler wraps
// the constellation mux the way runServe does.
func beaconServer(t *testing.T) http.Handler {
	t.Helper()
	s, _ := testServer(t)
	s.opt.cfBeaconToken = cfTestToken
	return s.handler(nil, nil)
}

// htmlFamilies names one route per HTML-producing family of the constellation:
// each is rendered by a different package, and every one of them must carry the
// tag. The comparison report is the fake render of testServer, which is enough
// to prove the /view route goes through the wrapper.
var htmlFamilies = []string{
	"/",                                 // landing.go
	"/visualizer",                       // hub.go
	"/view?ex=claude-dragonlite",        // pkg/report, through pkg/compare
	"/firebook/fr/",                     // pkg/firebook, French edition
	"/firebook/en/",                     // pkg/firebook, English edition
	"/firebook/fr/actifs-defensifs",     // one article, not just an index
	"/firebook/en/anarkulova-cederburg", // and its English counterpart
	"/firesimulator/",                   // pkg/decumul/web, mounted
	"/view?p=ZZZNOTANID:100",            // the shared error page (serve.go)
}

// notHTML names the machine-readable routes: markdown mirrors, feeds, sitemap,
// JSON, the raw portfolio files. None of them is a page, so none may be
// rewritten.
var notHTML = []string{
	"/firebook/fr/actifs-defensifs.md",
	"/firebook/en/feed.xml",
	"/sitemap.xml",
	"/robots.txt",
	"/llms.txt",
	"/catalog.json",
	"/healthz",
	"/theme.css",
	"/favicon.svg",
	"/examples/claude-dragonlite.txt",
}

func TestServeBeaconOnEveryHTMLFamily(t *testing.T) {
	h := beaconServer(t)
	want := webui.BeaconSnippet(cfTestToken)
	for _, path := range htmlFamilies {
		rec := serveGet(t, h, path)
		body := rec.Body.String()
		if !strings.Contains(body, want) {
			t.Errorf("%s (code=%d): no analytics tag", path, rec.Code)
			continue
		}
		if n := strings.Count(body, webui.BeaconScriptURL); n != 1 {
			t.Errorf("%s: %d tags, want exactly 1", path, n)
		}
	}
}

func TestServeBeaconSkipsNonPages(t *testing.T) {
	h := beaconServer(t)
	for _, path := range notHTML {
		rec := serveGet(t, h, path)
		if rec.Code != http.StatusOK {
			t.Errorf("%s: code=%d", path, rec.Code)
		}
		if strings.Contains(rec.Body.String(), "cloudflareinsights") {
			t.Errorf("%s: a non-page must carry no analytics markup", path)
		}
	}
}

// Without a token the feature leaves no trace anywhere, and the pages are the
// bytes they were: each family is compared against the same route served by a
// server that never heard of the beacon.
func TestServeWithoutBeaconTokenIsByteIdentical(t *testing.T) {
	s, _ := testServer(t)
	plain := s.handler(nil, nil)
	s2, _ := testServer(t)
	s2.opt.cfBeaconToken = ""
	same := s2.handler(nil, nil)
	for _, path := range append(append([]string{}, htmlFamilies...), notHTML...) {
		a, b := serveGet(t, plain, path), serveGet(t, same, path)
		if a.Body.String() != b.Body.String() {
			t.Errorf("%s: an empty token changed the bytes", path)
		}
		if strings.Contains(a.Body.String(), "cloudflareinsights") {
			t.Errorf("%s: analytics markup with no token", path)
		}
	}
}

// The book's markdown mirror computes a strong ETag over the bytes it sends and
// answers If-None-Match with a 304. The wrapper must not touch that route, so
// the conditional request keeps working with the beacon on.
func TestServeBeaconLeavesMarkdownETagIntact(t *testing.T) {
	h := beaconServer(t)
	const path = "/firebook/fr/actifs-defensifs.md"
	etag := serveGet(t, h, path).Header().Get("ETag")
	if etag == "" {
		t.Fatal("the markdown mirror lost its ETag")
	}
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("If-None-Match", etag)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotModified {
		t.Errorf("If-None-Match: code=%d, want 304", rec.Code)
	}
}

// An empty -cf-beacon-token defers to POFO_CF_BEACON_TOKEN, so a container
// image with a fixed command line can turn the beacon on from its environment.
// -serve is not started here: reaching the flag-validation error that follows
// proves the fallback ran first.
func TestBeaconTokenFromEnvironment(t *testing.T) {
	t.Setenv("POFO_CF_BEACON_TOKEN", cfTestToken)
	t.Setenv("POFO_INDEXNOW_KEY", "not a valid key")
	if err := run(context.Background(), []string{"-serve"}); err == nil ||
		!strings.Contains(err.Error(), "invalid -indexnow-key") {
		t.Fatalf("unexpected error: %v", err)
	}
}
