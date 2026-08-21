package seo

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// mockEndpoint stands in for api.indexnow.org: it records what was POSTed and
// answers with the status the test asks for. Nothing in this package's tests
// ever reaches the real endpoint.
func mockEndpoint(t *testing.T, status int, answer string) (*httptest.Server, *[]indexNowBody) {
	t.Helper()
	var got []indexNowBody
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("endpoint got %s, want POST", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}
		body, _ := io.ReadAll(r.Body)
		var b indexNowBody
		if err := json.Unmarshal(body, &b); err != nil {
			t.Errorf("endpoint got malformed JSON: %v", err)
		}
		got = append(got, b)
		w.WriteHeader(status)
		_, _ = w.Write([]byte(answer))
	}))
	t.Cleanup(srv.Close)
	return srv, &got
}

func sample() IndexNow {
	return IndexNow{
		Host:        "example.org",
		Key:         "1a2b3c4d5e6f7890",
		KeyLocation: "https://example.org/1a2b3c4d5e6f7890.txt",
		URLs:        []string{"https://example.org/", "https://example.org/book/one"},
	}
}

func TestIndexNowSubmit(t *testing.T) {
	srv, got := mockEndpoint(t, http.StatusOK, "")
	n := sample()
	if err := n.Submit(context.Background(), srv.Client(), srv.URL); err != nil {
		t.Fatalf("Submit: %v", err)
	}
	if len(*got) != 1 {
		t.Fatalf("endpoint saw %d requests, want 1", len(*got))
	}
	b := (*got)[0]
	if b.Host != n.Host || b.Key != n.Key || b.KeyLocation != n.KeyLocation {
		t.Errorf("payload identity = %+v", b)
	}
	if len(b.URLList) != len(n.URLs) || b.URLList[0] != n.URLs[0] {
		t.Errorf("payload urlList = %v", b.URLList)
	}
}

// 202 means "queued behind the key-file check", which is a normal answer for a
// key an engine has not read yet, not a failure.
func TestIndexNowAcceptsAccepted(t *testing.T) {
	srv, _ := mockEndpoint(t, http.StatusAccepted, "")
	if err := sample().Submit(context.Background(), srv.Client(), srv.URL); err != nil {
		t.Errorf("202 was reported as an error: %v", err)
	}
}

// A refusal is reported with the endpoint's own words: 403 says the key could
// not be verified, and that is what a deploy needs to read.
func TestIndexNowReportsRefusal(t *testing.T) {
	srv, _ := mockEndpoint(t, http.StatusForbidden, "key not valid")
	err := sample().Submit(context.Background(), srv.Client(), srv.URL)
	if err == nil {
		t.Fatal("a 403 was reported as success")
	}
	if !strings.Contains(err.Error(), "key not valid") || !strings.Contains(err.Error(), "403") {
		t.Errorf("error does not carry the endpoint's answer: %v", err)
	}
}

// The protocol caps a request at IndexNowBatch URLs, so a longer list is split
// and every URL still gets submitted exactly once.
func TestIndexNowBatches(t *testing.T) {
	n := sample()
	n.URLs = make([]string, IndexNowBatch+7)
	for i := range n.URLs {
		n.URLs[i] = "https://example.org/p/" + strings.Repeat("x", i%3) + string(rune('a'+i%26))
	}
	bodies := n.Bodies()
	if len(bodies) != 2 {
		t.Fatalf("%d URLs made %d requests, want 2", len(n.URLs), len(bodies))
	}
	srv, got := mockEndpoint(t, http.StatusOK, "")
	if err := n.Submit(context.Background(), srv.Client(), srv.URL); err != nil {
		t.Fatalf("Submit: %v", err)
	}
	total := 0
	for _, b := range *got {
		if len(b.URLList) > IndexNowBatch {
			t.Errorf("a batch carried %d URLs, over the %d cap", len(b.URLList), IndexNowBatch)
		}
		total += len(b.URLList)
	}
	if total != len(n.URLs) {
		t.Errorf("submitted %d URLs, want %d", total, len(n.URLs))
	}
}

// Everything that would be refused, or that would mount a route other than a
// file, is caught before a single byte is sent.
func TestIndexNowValidate(t *testing.T) {
	cases := []struct {
		name  string
		mutex func(*IndexNow)
		ok    bool
	}{
		{"as built", func(*IndexNow) {}, true},
		{"short key", func(n *IndexNow) { n.Key = "abc" }, false},
		{"key with a slash", func(n *IndexNow) { n.Key = "abc/../etc" }, false},
		{"key with a dot", func(n *IndexNow) { n.Key = "abcdefg.h" }, false},
		{"no host", func(n *IndexNow) { n.Host = "" }, false},
		{"no URLs", func(n *IndexNow) { n.URLs = nil }, false},
		{"a URL on another host", func(n *IndexNow) {
			n.URLs = append(n.URLs, "https://elsewhere.example/")
		}, false},
	}
	for _, c := range cases {
		n := sample()
		c.mutex(&n)
		err := n.Validate()
		if (err == nil) != c.ok {
			t.Errorf("%s: Validate() = %v, want ok=%v", c.name, err, c.ok)
		}
		if c.ok {
			continue
		}
		// A submission that does not validate is never sent.
		srv, got := mockEndpoint(t, http.StatusOK, "")
		if err := n.Submit(context.Background(), srv.Client(), srv.URL); err == nil {
			t.Errorf("%s: Submit succeeded on an invalid submission", c.name)
		}
		if len(*got) != 0 {
			t.Errorf("%s: an invalid submission reached the endpoint", c.name)
		}
	}
}

func TestValidIndexNowKey(t *testing.T) {
	for key, want := range map[string]bool{
		"1a2b3c4d":               true,
		"a-b-c-d-e-f":            true,
		"short7c":                false,
		"has space aa":           false,
		"has.dot.aa":             false,
		"has/slash/aa":           false,
		"héhéhéhé":               false,
		"":                       false,
		strings.Repeat("a", 128): true,
		strings.Repeat("a", 129): false,
	} {
		if got := ValidIndexNowKey(key); got != want {
			t.Errorf("ValidIndexNowKey(%q) = %v, want %v", key, got, want)
		}
	}
}
