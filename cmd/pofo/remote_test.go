package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

func TestRemoteOptInAndCommonForeignBudget(t *testing.T) {
	for _, tc := range []struct{ api, mcp bool }{{false, false}, {true, false}, {false, true}, {true, true}} {
		s := newServer(&options{serveAPI: tc.api, serveMCP: tc.mcp, foreignPerHour: 1, foreignGlobalPerHour: 1}, marketdata.NewClient(""))
		h := s.handler(fireBuild{})
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", "/api/v1/openapi.json", nil))
		if (w.Code == 200) != tc.api {
			t.Fatalf("api flag %+v: %d", tc, w.Code)
		}
		w = httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"ping"}`))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Accept", "application/json, text/event-stream")
		h.ServeHTTP(w, r)
		if (w.Code == 200) != tc.mcp {
			t.Fatalf("mcp flag %+v: %d", tc, w.Code)
		}
	}
}
func TestRemoteSharesViewGateAndBoundsQueue(t *testing.T) {
	s := newServer(&options{}, nil)
	for range cap(s.sem) {
		s.sem <- struct{}{}
	}
	for range cap(s.workQueue) {
		s.workQueue <- struct{}{}
	}
	if release, err := s.acquireWork(context.Background()); err == nil {
		release()
		t.Fatal("unbounded queue")
	}
	<-s.workQueue
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if _, err := s.acquireWork(ctx); err != context.DeadlineExceeded {
		t.Fatal(err)
	}
}

func TestFirePageBurstDoesNotUseViewQueue(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		s := newServer(&options{serveAPI: enabled}, marketdata.NewClient(""))
		// A busy visualizer must not reject any section of a FIRE page.
		for range cap(s.sem) {
			s.sem <- struct{}{}
		}
		for range cap(s.workQueue) {
			s.workQueue <- struct{}{}
		}
		h := s.handler(fireBuild{})
		results := make(chan int, 48)
		for range 48 {
			go func() {
				r := httptest.NewRequest("POST", "/firesimulator/api/models", strings.NewReader(`{"capital":800000,"needAnnual":25000,"years":5,"nPaths":16,"mu":0.04,"sigma":0.12,"df":5}`))
				w := httptest.NewRecorder()
				h.ServeHTTP(w, r)
				results <- w.Code
			}()
		}
		for range 48 {
			if code := <-results; code != 200 {
				t.Errorf("API=%v: status %d", enabled, code)
			}
		}
	}
}
func TestRemoteAccessLogRedacts(t *testing.T) {
	var log bytes.Buffer
	h := logAccess(&log, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }))
	r := httptest.NewRequest("POST", "/mcp?capital=123456SECRET", strings.NewReader("private-body"))
	r.Header.Set("Referer", "https://example.test/secret-referrer")
	r.Header.Set("User-Agent", "secret-agent")
	h.ServeHTTP(httptest.NewRecorder(), r)
	for _, word := range []string{"SECRET", "private-body", "secret-referrer", "secret-agent"} {
		if strings.Contains(log.String(), word) {
			t.Fatalf("log leaked %s", word)
		}
	}
}

func TestAPIAndWebUseOneForeignAllowance(t *testing.T) {
	s := newServer(&options{serveAPI: true, foreignPerHour: 1, foreignGlobalPerHour: 1}, marketdata.NewClient(""))
	request := httptest.NewRequest("GET", "/view", nil)
	if err := s.foreignGate(request).charge([]string{"ZZTEST"}); err != nil {
		t.Fatal(err)
	}
	h := s.handler(fireBuild{})
	r := httptest.NewRequest("POST", "/api/v1/assets/analyze", strings.NewReader(`{"ids":["ZZOTHER"]}`))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 429 || !strings.Contains(w.Body.String(), "rate_limited") {
		t.Fatalf("independent quota: %d %s", w.Code, w.Body)
	}
}

func TestRemotePreservesMarketDataDiagnostics(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		client := marketdata.NewClient("")
		logged := false
		client.Logf = func(string, ...any) { logged = true }
		newServer(&options{serveAPI: enabled}, client)
		client.Logf("downloading %s", "PRIVATE-REQUEST-ID")
		if !logged {
			t.Fatalf("enabled=%v source log retained=%v", enabled, logged)
		}
	}
}
