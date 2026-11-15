package marketdata

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestCacheOnlyBlocksAllHTTP(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1) }))
	defer srv.Close()
	c := NewClient("")
	c.CookieBase = srv.URL
	c.FredBase = srv.URL
	ctx := WithCacheOnly(context.Background())
	if _, err := c.get(ctx, srv.URL); !errors.Is(err, ErrOffline) {
		t.Fatalf("get: %v", err)
	}
	if _, err := c.fetchYahooAuth(ctx); !errors.Is(err, ErrOffline) {
		t.Fatalf("auth: %v", err)
	}
	if _, err := c.fetchFRED(ctx, "TEST"); !errors.Is(err, ErrOffline) {
		t.Fatalf("FRED: %v", err)
	}
	if calls.Load() != 0 {
		t.Fatal("offline call used network")
	}
}

func TestControlsCountRedirectAndCancelWait(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/next", http.StatusFound)
		}
	}))
	defer srv.Close()
	c := NewClient("")
	c.Controls = NewControls(1, time.Hour, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	if _, err := c.get(ctx, srv.URL); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("error %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("requests %d", calls.Load())
	}
}

func TestCacheOnlyRejectsStaleMemoAndOfflineCache(t *testing.T) {
	c := NewClient(t.TempDir())
	from := testDays(1)[0]
	s := &Series{Symbol: "ZZZTEST", Currency: "USD", Points: []Point{{Date: from, Close: 100}, {Date: from.AddDate(0, 0, 1), Close: 101}}}
	c.saveCache(s, from)
	c.memoize("ZZZTEST|"+from.Format(time.DateOnly), s)
	c.MaxAge = -time.Second
	c.Offline = true
	if _, err := c.History(WithCacheOnly(context.Background()), "ZZZTEST", from); !errors.Is(err, ErrOffline) {
		t.Fatalf("stale response: %v", err)
	}
}

func TestCoalescingWaiterCancellation(t *testing.T) {
	var calls atomic.Int32
	entered, release := make(chan struct{}), make(chan struct{})
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/VOO", func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			close(entered)
		}
		select {
		case <-release:
		case <-r.Context().Done():
			return
		}
		fmt.Fprint(w, chartJSON("VOO", testDays(3), []float64{100, 101, 102}))
	})
	c, srv := newTestClient(t, "", mux)
	defer srv.Close()
	c.Controls = NewControls(0, 0, 0)
	opt := FetchOptions{From: testDays(1)[0], NoConvert: true, ExactOnly: true}
	ctx, cancel := context.WithCancel(context.Background())
	first := make(chan error, 1)
	go func() { _, err := c.FetchExtended(ctx, "VOO", opt); first <- err }()
	<-entered
	second := make(chan error, 1)
	go func() { _, err := c.FetchExtended(context.Background(), "VOO", opt); second <- err }()
	deadline := time.Now().Add(time.Second)
	for {
		c.Controls.mu.Lock()
		joined := false
		for _, f := range c.Controls.flights {
			if f.waiters == 2 {
				joined = true
			}
		}
		c.Controls.mu.Unlock()
		if joined {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("second waiter never joined")
		}
		time.Sleep(time.Millisecond)
	}
	cancel()
	if err := <-first; !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	close(release)
	if err := <-second; err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 1 {
		t.Fatalf("calls %d", calls.Load())
	}
}

func ExampleNewControls() {
	client := NewClient("")
	client.Controls = NewControls(60, time.Minute, 10)
	// Coalescing and outbound admission are opt-in for library clients.
	ctx := WithCacheOnly(context.Background())
	_, err := client.get(ctx, "https://example.invalid")
	fmt.Println(errors.Is(err, ErrOffline))
	// Output: true
}

func TestCacheOnlyCanUseFreshMemoryWithoutDisk(t *testing.T) {
	mux := http.NewServeMux()
	var calls atomic.Int32
	mux.HandleFunc("/v8/finance/chart/VOO", func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		fmt.Fprint(w, chartJSON("VOO", testDays(3), []float64{100, 101, 102}))
	})
	c, srv := newTestClient(t, "", mux)
	defer srv.Close()
	from := testDays(1)[0]
	if _, err := c.History(context.Background(), "VOO", from); err != nil {
		t.Fatal(err)
	}
	if _, err := c.History(WithCacheOnly(context.Background()), "VOO", from); err != nil {
		t.Fatalf("fresh memo refused: %v", err)
	}
	if calls.Load() != 1 {
		t.Fatal("network used for cached request")
	}
}
