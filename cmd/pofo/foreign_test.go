package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/portfolio"
	"github.com/bpineau/pofo/pkg/scenario"
)

// testGate returns a gate with a generous budget and a fixed clock, plus the
// budget so a test can inspect or exhaust it.
func testGate(perClient, global int) (*foreignGate, *foreignBudget) {
	b := newForeignBudget(perClient, global, time.Hour)
	return &foreignGate{budget: b, client: "203.0.113.7"}, b
}

// TestForeignSyntacticGate pins what a p= holding may be: a catalog
// identifier always, an identifier outside it only when the server runs a
// budget AND the identifier looks like an ISIN or a ticker.
func TestForeignSyntacticGate(t *testing.T) {
	cases := []struct {
		id       string
		withGate bool // accepted when the server accepts foreign identifiers
		why      string
	}{
		{"IWDA", true, "a catalog id needs no budget"},
		{"IWDASIM", true, "a catalog id with the SIM suffix"},
		{"DGRO", true, "a plain foreign ticker"},
		{"DGRO.US", true, "a foreign ticker with an exchange suffix"},
		{"US0378331005", true, "a valid ISIN outside the catalog"},
		{"US0378331004", false, "the same ISIN with a wrong check digit"},
		{"^GSPC", false, "a raw quote symbol"},
		{"AVERYLONGIDENTIFIER", false, "too long for any ticker"},
		{"DGRO/../ETC", false, "path characters"},
	}
	for _, c := range cases {
		local := marketdata.KnownLocal(c.id)
		// With no gate (the CLI, and a server with the feature off) only the
		// catalog is accepted: today's policy, unchanged.
		_, err := adhocSpec(c.id+":100", 1, nil)
		if (err == nil) != local {
			t.Errorf("no gate: adhocSpec(%s) error = %v, want accepted = %v", c.id, err, local)
		}
		if err != nil && !strings.Contains(err.Error(), "not in the local catalog") {
			t.Errorf("no gate: adhocSpec(%s) should keep the catalog message, got %v", c.id, err)
		}
		gate, _ := testGate(10, 60)
		_, err = adhocSpec(c.id+":100", 1, gate)
		if (err == nil) != c.withGate {
			t.Errorf("with a budget: adhocSpec(%s) error = %v, want accepted = %v (%s)",
				c.id, err, c.withGate, c.why)
		}
	}
}

func TestForeignBudgetPerClientWindow(t *testing.T) {
	gate, b := testGate(2, 10)
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	b.now = func() time.Time { return now }

	if _, err := adhocSpec("DGRO:50,SCHG:50", 1, gate); err != nil {
		t.Fatalf("two new identifiers fit a budget of two: %v", err)
	}
	_, err := adhocSpec("FNDX:100", 1, gate)
	if !errors.Is(err, errForeignBudget) {
		t.Fatalf("the third identifier must be refused, got %v", err)
	}
	// A catalog identifier is never charged, so it still works.
	if _, err := adhocSpec("IWDA:100", 1, gate); err != nil {
		t.Fatalf("a catalog portfolio must stay servable: %v", err)
	}
	// An identifier already charged is not cached (nothing fetched here), so
	// it is charged again and refused too: the budget counts requests, not
	// distinct symbols.
	if _, err := adhocSpec("DGRO:100", 1, gate); !errors.Is(err, errForeignBudget) {
		t.Fatalf("the window is spent, got %v", err)
	}
	// The window rolls: an hour later the allowance is back.
	now = now.Add(time.Hour + time.Minute)
	if _, err := adhocSpec("FNDX:100", 1, gate); err != nil {
		t.Fatalf("the rolling window must free the allowance: %v", err)
	}
}

func TestForeignBudgetGlobalCeiling(t *testing.T) {
	b := newForeignBudget(2, 3, time.Hour)
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	b.now = func() time.Time { return now }

	for i, client := range []string{"a", "a", "b"} {
		if err := b.charge(client, 1); err != nil {
			t.Fatalf("charge %d (%s) within both budgets: %v", i, client, err)
		}
	}
	// Client "c" has its full personal allowance, but the process does not.
	err := b.charge("c", 1)
	if !errors.Is(err, errForeignBudget) || !strings.Contains(err.Error(), "server-wide") {
		t.Fatalf("the global ceiling must bite and say so, got %v", err)
	}
	// An all-or-nothing charge leaves nothing behind: once the window rolls,
	// the refused client starts from a clean slate.
	now = now.Add(2 * time.Hour)
	if err := b.charge("c", 2); err != nil {
		t.Fatalf("after the window: %v", err)
	}
}

func TestForeignBudgetCacheHitsAndDuplicatesAreFree(t *testing.T) {
	gate, b := testGate(1, 10)
	b.now = time.Now
	gate.cached = func(id string) bool { return id == "DGRO" }

	// DGRO is already quotable offline, so only SCHG costs an upstream call.
	if _, err := adhocSpec("DGRO:50,SCHG:50", 1, gate); err != nil {
		t.Fatalf("a cached identifier must not be charged: %v", err)
	}
	if _, err := adhocSpec("DGRO:100", 1, gate); err != nil {
		t.Fatalf("a cached identifier stays free once the budget is spent: %v", err)
	}
	if _, err := adhocSpec("FNDX:100", 1, gate); !errors.Is(err, errForeignBudget) {
		t.Fatalf("the one new identifier was already spent, got %v", err)
	}

	// Within one portfolio the same identifier is fetched once, so it is
	// charged once, however often it is written.
	gate2, b2 := testGate(1, 10)
	b2.now = time.Now
	if _, err := adhocSpec("DFAC:50,dfac:50", 1, gate2); err != nil {
		t.Fatalf("a repeated identifier must be charged once: %v", err)
	}
}

func TestForeignBudgetClientTableBounded(t *testing.T) {
	b := newForeignBudget(1, maxForeignClients*2, time.Hour)
	base := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	tick := 0
	b.now = func() time.Time { tick++; return base.Add(time.Duration(tick) * time.Millisecond) }

	for i := range maxForeignClients {
		if err := b.charge(fmt.Sprintf("client-%d", i), 1); err != nil {
			t.Fatalf("charge %d: %v", i, err)
		}
	}
	if len(b.clients) != maxForeignClients {
		t.Fatalf("table holds %d entries, want %d", len(b.clients), maxForeignClients)
	}
	if err := b.charge("newcomer", 1); err != nil {
		t.Fatalf("a newcomer must still be served: %v", err)
	}
	if len(b.clients) > maxForeignClients {
		t.Errorf("the table grew to %d entries past its cap", len(b.clients))
	}
	if _, ok := b.clients["client-0"]; ok {
		t.Error("the least recently charged entry should have been evicted")
	}
	if _, ok := b.clients["newcomer"]; !ok {
		t.Error("the newcomer's window was not recorded")
	}
}

// foreignServer is a testServer whose /view accepts identifiers outside the
// catalog, with a cache-less client (so nothing ever reads as already cached)
// and the fake renderer, so no test touches the network.
func foreignServer(t *testing.T, perClient, global int) *server {
	t.Helper()
	s := newServer(&options{
		currency: "EUR", benchmark: "^GSPC", rebalance: 90,
		foreignPerHour: perClient, foreignGlobalPerHour: global,
	}, marketdata.NewClient(""))
	s.render = func(ctx context.Context, o *options, specs []*portfolio.Spec) ([]byte, error) {
		return []byte("<html>fake report</html>"), nil
	}
	s.buildPanel = func(ctx context.Context, spec *portfolio.Spec) (*scenario.Panel, []string) {
		return &scenario.Panel{}, stubLabels(spec)
	}
	return s
}

// serveGetFrom is serveGet with an explicit client address, so the per-client
// budget can be exercised.
func serveGetFrom(t *testing.T, h http.Handler, path, addr string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, path, nil)
	r.RemoteAddr = addr + ":54321"
	h.ServeHTTP(rec, r)
	return rec
}

func TestServeForeignIdentifiers(t *testing.T) {
	s := foreignServer(t, 2, 3)
	h := s.handler(nil, nil)

	if rec := serveGetFrom(t, h, "/view?p=DGRO:100", "198.51.100.1"); rec.Code != 200 {
		t.Fatalf("a well-formed foreign ticker should render: code=%d body=%q", rec.Code, rec.Body.String())
	}
	if rec := serveGetFrom(t, h, "/view?p=%5EGSPC:100", "198.51.100.1"); rec.Code != 400 ||
		!strings.Contains(rec.Body.String(), "not in the local catalog") {
		t.Errorf("a malformed identifier is a 400: code=%d body=%q", rec.Code, rec.Body.String())
	}
	// Second new identifier for this client: still inside its budget of two.
	if rec := serveGetFrom(t, h, "/view?p=SCHG:100", "198.51.100.1"); rec.Code != 200 {
		t.Fatalf("the second identifier fits the budget: code=%d", rec.Code)
	}
	rec := serveGetFrom(t, h, "/view?p=FNDX:100", "198.51.100.1")
	if rec.Code != http.StatusTooManyRequests || !strings.Contains(rec.Body.String(), "budget spent") {
		t.Fatalf("a spent budget is a 429 with an explanation: code=%d body=%q", rec.Code, rec.Body.String())
	}
	// The catalog is never rationed: the same client keeps composing.
	if rec := serveGetFrom(t, h, "/view?p=IWDA:100", "198.51.100.1"); rec.Code != 200 {
		t.Errorf("a catalog portfolio must stay servable: code=%d", rec.Code)
	}
	// Another client has its own allowance, but the process ceiling (3) is
	// one short of it.
	if rec := serveGetFrom(t, h, "/view?p=FNDX:100", "203.0.113.9"); rec.Code != 200 {
		t.Errorf("a fresh client's first identifier: code=%d", rec.Code)
	}
	rec = serveGetFrom(t, h, "/view?p=DFAC:100", "203.0.113.9")
	if rec.Code != http.StatusTooManyRequests || !strings.Contains(rec.Body.String(), "server-wide") {
		t.Errorf("the global ceiling must answer 429: code=%d body=%q", rec.Code, rec.Body.String())
	}
	// The FIRE mount of a composed portfolio shares the same gate.
	if rec := serveGetFrom(t, h, "/firesimulator/p/ZZZZ:100/", "192.0.2.5"); rec.Code != http.StatusTooManyRequests {
		t.Errorf("the composed FIRE mount must charge the same budget: code=%d", rec.Code)
	}
}

// A p= identifier that passes the syntactic gate but that no source quotes
// (ZQXWVT is a perfectly well-formed ticker) is the request's own fault: a
// 404 naming it, since no retry will ever help. An upstream outage on the
// same URL keeps the opaque 500, so a broken source is never dressed up as a
// visitor's typo.
func TestServeUnknownIdentifierStatus(t *testing.T) {
	s := foreignServer(t, 5, 10)
	h := s.handler(nil, nil)

	s.render = func(ctx context.Context, o *options, specs []*portfolio.Spec) ([]byte, error) {
		return nil, fmt.Errorf("portfolio %s, asset %q: %w", "p1", "ZQXWVT",
			&marketdata.UnknownIdentifierError{
				ID:       "ZQXWVT",
				Failures: errors.New("ticker ZQXWVT: no usable source (yahoo: HTTP 404)"),
			})
	}
	rec := serveGetFrom(t, h, "/view?p=ZQXWVT:100", "198.51.100.4")
	if rec.Code != http.StatusNotFound || !strings.Contains(rec.Body.String(), "no source quotes ZQXWVT") {
		t.Errorf("an identifier nothing quotes is a 404 naming it: code=%d body=%q", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "no usable source") {
		t.Error("the per-source detail belongs in the log, not in the page")
	}

	s.render = func(ctx context.Context, o *options, specs []*portfolio.Spec) ([]byte, error) {
		return nil, fmt.Errorf("portfolio %s, asset %q: %w", "p1", "IWDA",
			errors.New("downloading IWDA failed (yahoo: HTTP 500; stooq: HTTP 500)"))
	}
	rec = serveGetFrom(t, h, "/view?p=IWDA:100", "198.51.100.4")
	if rec.Code != http.StatusInternalServerError ||
		!strings.Contains(rec.Body.String(), "the comparison failed") {
		t.Errorf("an upstream outage stays a 500: code=%d body=%q", rec.Code, rec.Body.String())
	}
}

// With the feature off (the default), the whole surface behaves exactly as
// before: a foreign identifier is a 400, whatever its shape.
func TestServeForeignOffByDefault(t *testing.T) {
	s, _ := testServer(t)
	if s.foreign != nil {
		t.Fatal("the default options must leave the foreign-identifier budget off")
	}
	h := s.handler(nil, nil)
	rec := serveGetFrom(t, h, "/view?p=DGRO:100", "198.51.100.1")
	if rec.Code != 400 || !strings.Contains(rec.Body.String(), "not in the local catalog") {
		t.Errorf("catalog-only server: code=%d body=%q", rec.Code, rec.Body.String())
	}
}

// The composer only advertises live identifiers when the server accepts them:
// the front end reds an unknown id otherwise.
func TestComposerForeignCaps(t *testing.T) {
	vr, err := parseViewQuery(mustQuery(t, "p=IWDA:100"), viewBase(), nil)
	if err != nil {
		t.Fatal(err)
	}
	off := string(composerMount(vr, 0))
	if strings.Contains(off, `&#34;foreign&#34;:true`) || strings.Contains(off, "per hour outside the catalog") {
		t.Error("a catalog-only server must not advertise live identifiers")
	}
	on := string(composerMount(vr, 10))
	if !strings.Contains(on, `&#34;foreign&#34;:true`) {
		t.Error("the caps must tell the front end that foreign identifiers are accepted")
	}
	if !strings.Contains(on, "10 new ones per hour outside the catalog") {
		t.Error("the panel must state the allowance")
	}
}
