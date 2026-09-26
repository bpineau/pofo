package marketdata

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// from2020 is a long-enough window for the resolution search to run in full.
func from2020() time.Time { return time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC) }

// A well-formed identifier no source carries must be reported as the
// visitor's own mistake, not as a failure of ours: every source answered, and
// every answer was "nothing here" (an empty mux 404s), so the failure carries
// the identifier and matches ErrUnknownIdentifier.
func TestFetchUnknownIdentifier(t *testing.T) {
	dir := t.TempDir()
	c, _ := newTestClient(t, dir, http.NewServeMux())

	_, err := c.Fetch(context.Background(), "ZQXWVT", from2020())
	if err == nil {
		t.Fatal("a ticker nothing quotes must fail")
	}
	if !errors.Is(err, ErrUnknownIdentifier) {
		t.Errorf("errors.Is(ErrUnknownIdentifier) = false for %v", err)
	}
	var unknown *UnknownIdentifierError
	if !errors.As(err, &unknown) {
		t.Fatalf("errors.As(*UnknownIdentifierError) = false for %v", err)
	}
	if unknown.ID != "ZQXWVT" {
		t.Errorf("identifier reported = %q, want ZQXWVT", unknown.ID)
	}
	// The per-source summary must survive, for the log.
	if got := err.Error(); !strings.Contains(got, "no usable source") {
		t.Errorf("the failure summary was lost: %q", got)
	}
	// A failed lookup must leave the quote cache alone, or the next visitor
	// inherits a poisoned entry for an identifier that does not exist.
	if entries, rerr := os.ReadDir(dir); rerr == nil && len(entries) > 0 {
		t.Errorf("a failed lookup wrote to the cache: %v", entries)
	}
	if c.Cached("ZQXWVT") {
		t.Error("a failed lookup must not report the identifier as cached")
	}
}

// The complementary case: the identifier may be perfectly real and every
// source simply down. That failure must NOT be blamed on the identifier, so a
// caller keeps answering "come back later" rather than "no such thing".
func TestFetchSourceOutageIsNotUnknown(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	})
	c, _ := newTestClient(t, t.TempDir(), mux)

	_, err := c.Fetch(context.Background(), "VOO", from2020())
	if err == nil {
		t.Fatal("every source down must fail")
	}
	if errors.Is(err, ErrUnknownIdentifier) {
		t.Errorf("an outage was read as an unknown identifier: %v", err)
	}
}

// Same for an ISIN: the kind of identifier does not change the verdict.
func TestFetchUnknownISIN(t *testing.T) {
	c, _ := newTestClient(t, t.TempDir(), http.NewServeMux())

	_, err := c.Fetch(context.Background(), "US0378331005", from2020())
	var unknown *UnknownIdentifierError
	if !errors.As(err, &unknown) || unknown.ID != "US0378331005" {
		t.Fatalf("an ISIN nothing quotes must name itself: %v", err)
	}
}

// markAbsentIfAll is the rule that keeps a combined failure honest: one
// unreachable source is enough to make the whole summary uninformative.
func TestMarkAbsentIfAll(t *testing.T) {
	nothing, down := markAbsent(errors.New("no data")), errors.New("HTTP 500")
	summary := errors.New("combined")
	if !absent(markAbsentIfAll(summary, nothing, nothing)) {
		t.Error("every source reporting an absence must combine into one")
	}
	if absent(markAbsentIfAll(summary, nothing, down)) {
		t.Error("one source down must void the combined absence")
	}
	if got := markAbsentIfAll(summary, nothing, nothing).Error(); got != "combined" {
		t.Errorf("marking must not change the message: %q", got)
	}
}
