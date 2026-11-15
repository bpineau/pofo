package marketdata

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type cacheOnlyKey struct{}

// WithCacheOnly forbids network access for this request and all its dependencies.
// Unlike Client.Offline, it preserves normal freshness/depth checks and refuses
// stale-cache fallbacks. Bundled snapshots remain usable. It does not mutate Client.
func WithCacheOnly(ctx context.Context) context.Context {
	return context.WithValue(ctx, cacheOnlyKey{}, true)
}
func cacheOnly(ctx context.Context) bool { b, _ := ctx.Value(cacheOnlyKey{}).(bool); return b }

// Controls shares outbound admission across Clients and coalesces simultaneous
// fetches within each Client (different caches/configurations remain separate).
// Configure it before use; the zero value coalesces fetches without rate limiting.
// Unknown-identifier failures are cached for five minutes (at most 1024).
// Temporary outages are not negative-cached; no credentials are retained.
// Each waiter keeps its own deadline. Shared work is cancelled when the last
// waiter leaves, with no independent shorter timeout; Client.HTTP still bounds
// individual provider requests.
type Controls struct {
	mu       sync.Mutex
	rate     float64
	burst    float64
	tokens   float64
	last     time.Time
	flights  map[string]*flight
	failures map[string]failedFetch
	attempts uint64
}
type failedFetch struct {
	err   error
	until time.Time
}

type flight struct {
	done    chan struct{}
	cancel  context.CancelFunc
	waiters int
	series  *Series
	err     error
}

// NewControls limits actual HTTP attempts to requests per interval, with an
// initial burst, and coalesces identical concurrent FetchExtended calls. A
// nonpositive requests, interval or burst disables rate limiting. Retries and
// redirects consume tokens too. Waiting honors each request's context.
func NewControls(requests int, interval time.Duration, burst int) *Controls {
	c := &Controls{flights: make(map[string]*flight)}
	if requests > 0 && interval > 0 && burst > 0 {
		c.rate = float64(requests) / interval.Seconds()
		c.burst = float64(burst)
		c.tokens = c.burst
		c.last = time.Now()
	}
	return c
}
func (c *Controls) wait(ctx context.Context) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		c.mu.Lock()
		if c.rate == 0 {
			c.mu.Unlock()
			return nil
		}
		now := time.Now()
		c.tokens = min(c.burst, c.tokens+now.Sub(c.last).Seconds()*c.rate)
		c.last = now
		if c.tokens >= 1 {
			c.tokens--
			c.mu.Unlock()
			return nil
		}
		delay := time.Duration((1 - c.tokens) / c.rate * float64(time.Second))
		c.mu.Unlock()
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

type controlledTransport struct {
	base     http.RoundTripper
	controls *Controls
}

func (t controlledTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if cacheOnly(r.Context()) {
		return nil, ErrOffline
	}
	if t.controls != nil {
		if err := t.controls.wait(r.Context()); err != nil {
			return nil, err
		}
	}
	if t.controls != nil {
		t.controls.mu.Lock()
		t.controls.attempts++
		t.controls.mu.Unlock()
	}
	return t.base.RoundTrip(r)
}
func (c *Client) requestHTTP(h *http.Client, r *http.Request) (*http.Response, error) {
	if c.Offline || cacheOnly(r.Context()) {
		return nil, ErrOffline
	}
	cp := *h
	base := cp.Transport
	if base == nil {
		base = http.DefaultTransport
	}
	cp.Transport = controlledTransport{base, c.Controls}
	return cp.Do(r)
}
func (c *Client) coalesced(ctx context.Context, id string, opt FetchOptions) (*Series, error) {
	// Custom filesystems can be mutable and have no stable identity.
	if c.Controls == nil || opt.Simdata != nil {
		return c.fetchExtended(ctx, id, opt)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	base, simulated := SplitSim(id)
	keyID := CanonicalID(base)
	if simulated {
		keyID += "SIM"
	}
	keyBytes, _ := json.Marshal(struct {
		ID      string
		Options FetchOptions
		Offline bool
	}{keyID, opt, cacheOnly(ctx)})
	key := fmt.Sprintf("%p:%s", c, keyBytes)
	ctrl := c.Controls
	ctrl.mu.Lock()
	if ctrl.flights == nil {
		ctrl.flights = make(map[string]*flight)
	}
	if failed, ok := ctrl.failures[key]; ok && time.Now().Before(failed.until) {
		ctrl.mu.Unlock()
		return nil, failed.err
	}
	f := ctrl.flights[key]
	if f == nil {
		work, cancel := context.WithCancel(context.WithoutCancel(ctx))
		f = &flight{done: make(chan struct{}), cancel: cancel}
		ctrl.flights[key] = f
		go func() {
			s, err := c.fetchExtended(work, id, opt)
			ctrl.mu.Lock()
			f.series, f.err = s, err
			if errors.Is(err, ErrUnknownIdentifier) {
				if ctrl.failures == nil {
					ctrl.failures = make(map[string]failedFetch)
				}
				if len(ctrl.failures) >= 1024 {
					now := time.Now()
					for k, v := range ctrl.failures {
						if now.After(v.until) {
							delete(ctrl.failures, k)
						}
					}
					if len(ctrl.failures) >= 1024 {
						for k := range ctrl.failures {
							delete(ctrl.failures, k)
							break
						}
					}
				}
				ctrl.failures[key] = failedFetch{err, time.Now().Add(5 * time.Minute)}
			}
			if ctrl.flights[key] == f {
				delete(ctrl.flights, key)
			}
			close(f.done)
			ctrl.mu.Unlock()
			cancel()
		}()
	}
	f.waiters++
	ctrl.mu.Unlock()
	select {
	case <-ctx.Done():
		ctrl.mu.Lock()
		f.waiters--
		if f.waiters == 0 {
			if ctrl.flights[key] == f {
				delete(ctrl.flights, key)
			}
			f.cancel()
		}
		ctrl.mu.Unlock()
		return nil, ctx.Err()
	case <-f.done:
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if f.series == nil {
			return nil, f.err
		}
		// A caller may safely change the returned slices without racing a waiter.
		s := *f.series
		s.Points = append([]Point(nil), s.Points...)
		s.Warnings = append([]string(nil), s.Warnings...)
		s.Dividends = append([]Dividend(nil), s.Dividends...)
		return &s, f.err
	}
}

// HTTPAttempts returns the count of actual outbound attempts admitted by these
// controls, including retries, redirects and failed transport calls. Cache-only
// refusals and cancellations while waiting do not count. It is concurrency-safe.
func (c *Controls) HTTPAttempts() uint64 {
	if c == nil {
		return 0
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.attempts
}
