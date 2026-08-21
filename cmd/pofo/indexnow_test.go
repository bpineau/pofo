package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// The origin is the string every submitted URL is built on by concatenation,
// so a path, a query or a missing scheme is caught here rather than turned
// into a list of nonsense URLs.
func TestParseOrigin(t *testing.T) {
	cases := []struct {
		in         string
		normalized string
		host       string
		ok         bool
	}{
		{"https://example.org", "https://example.org", "example.org", true},
		{"https://example.org/", "https://example.org", "example.org", true},
		{"http://127.0.0.1:8787", "http://127.0.0.1:8787", "127.0.0.1:8787", true},
		{"example.org", "", "", false},
		{"ftp://example.org", "", "", false},
		{"https://example.org/firebook/fr/", "", "", false},
		{"https://example.org?x=1", "", "", false},
		{"https://", "", "", false},
		{"", "", "", false},
	}
	for _, c := range cases {
		got, host, err := parseOrigin(c.in)
		if (err == nil) != c.ok {
			t.Errorf("parseOrigin(%q) error = %v, want ok=%v", c.in, err, c.ok)
			continue
		}
		if c.ok && (got != c.normalized || host != c.host) {
			t.Errorf("parseOrigin(%q) = %q, %q, want %q, %q", c.in, got, host, c.normalized, c.host)
		}
	}
}

// -indexnow refuses to do anything without a key: the endpoint would reject an
// unsigned submission anyway, and the failure should not need a network round
// trip to be understood.
func TestRunIndexNowNeedsAKey(t *testing.T) {
	err := runIndexNow(context.Background(), "https://example.org", "")
	if err == nil || !strings.Contains(err.Error(), "-indexnow-key") {
		t.Errorf("error = %v, want it to name the missing key", err)
	}
	// A bad origin is caught first, still without touching the network.
	if err := runIndexNow(context.Background(), "example.org", "1a2b3c4d"); err == nil {
		t.Error("a schemeless origin was accepted")
	}
}

// What -indexnow submits is what the running server publishes: the two values
// come from one serveSite, so the list can never drift from the sitemap the
// mux serves.
func TestIndexNowSubmitsWhatTheServerServes(t *testing.T) {
	const origin = "http://example.com"
	pushed := serveSite().URLs(origin)

	srv, _ := testServer(t)
	h := srv.handler(nil, nil)
	rec := serveGet(t, h, "/sitemap.xml")
	sitemap := rec.Body.String()
	if len(pushed) == 0 {
		t.Fatal("nothing would be submitted")
	}
	for _, u := range pushed {
		if !strings.Contains(sitemap, "<loc>"+u+"</loc>") {
			t.Errorf("%q would be submitted but is not in the served sitemap", u)
		}
	}
	if got := strings.Count(sitemap, "<loc>"); got != len(pushed) {
		t.Errorf("the sitemap serves %d URLs, %d would be submitted", got, len(pushed))
	}
	// The app surfaces ride along with the book.
	for _, want := range []string{origin + "/", origin + "/visualizer", origin + fireBase + "/"} {
		found := false
		for _, u := range pushed {
			if u == want {
				found = true
			}
		}
		if !found {
			t.Errorf("%q would not be submitted", want)
		}
	}
}

// The ownership key file is served only when -indexnow-key names one, and it
// answers with the key itself: an engine reads it before trusting a
// submission signed with that key.
func TestServeIndexNowKeyFile(t *testing.T) {
	const key = "1a2b3c4d5e6f7890"

	s, _ := testServer(t)
	if rec := serveGet(t, s.handler(nil, nil), "/"+key+".txt"); rec.Code != http.StatusNotFound {
		t.Errorf("without -indexnow-key: status %d, want 404", rec.Code)
	}

	s, _ = testServer(t)
	s.opt.indexNowKey = key
	rec := serveGet(t, s.handler(nil, nil), "/"+key+".txt")
	if rec.Code != http.StatusOK {
		t.Fatalf("key file: status %d", rec.Code)
	}
	if rec.Body.String() != key {
		t.Errorf("key file body = %q, want the key", rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("key file Content-Type = %q, want text/plain", ct)
	}
}

// A guard on the wiring rather than on the protocol (pkg/seo tests that): the
// mode sends one POST carrying the host, the key and the URL list, and it is
// the only thing in the program that talks to a search engine.
func TestIndexNowPayloadShape(t *testing.T) {
	var body string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 1<<20)
		n, _ := r.Body.Read(buf)
		body = string(buf[:n])
	}))
	defer srv.Close()

	sub := indexNowSubmission("https://example.org", "example.org", "1a2b3c4d5e6f7890")
	if err := sub.Submit(context.Background(), srv.Client(), srv.URL); err != nil {
		t.Fatalf("Submit: %v", err)
	}
	for _, want := range []string{
		`"host":"example.org"`,
		`"key":"1a2b3c4d5e6f7890"`,
		`"keyLocation":"https://example.org/1a2b3c4d5e6f7890.txt"`,
		`https://example.org/firebook/fr/`,
		`https://example.org/visualizer`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("the submitted payload misses %q", want)
		}
	}
}

// An empty -indexnow-key defers to POFO_INDEXNOW_KEY, so a container image
// with a fixed command line can turn the feature on from its environment.
// An invalid env key must hit the same validation as the flag: reaching it
// proves the fallback is read before -serve would start listening.
func TestIndexNowKeyFromEnvironment(t *testing.T) {
	t.Setenv("POFO_INDEXNOW_KEY", "not a valid key")
	err := run(context.Background(), []string{"-serve"})
	if err == nil || !strings.Contains(err.Error(), "invalid -indexnow-key") {
		t.Fatalf("env key should reach the same validation as the flag; got %v", err)
	}
}
