// Identifiers outside the bundled catalog, for the web app: the syntactic
// gate a visitor's identifier must pass, and the budgets that bound how many
// of them this server fetches from the upstream sources.
//
// The policy in one line: a p= identifier is either locally known
// (marketdata.KnownLocal, free and unlimited) or well-formed enough to be
// worth a lookup (marketdata.PlausibleID: a checksummed ISIN or a plausible
// exchange ticker) AND within the requesting client's hourly allowance and
// the server's own. Only identifiers that would really cost an upstream
// request are charged: one already in the quote cache is free.
package main

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// errForeignBudget marks a rejection caused by an exhausted fetch budget
// rather than by a malformed request: handlers answer it with 429 instead of
// 400, so a client (or a crawler) can tell "come back later" from "never".
var errForeignBudget = errors.New("new identifiers: hourly budget spent, try later or pick a catalog id")

// maxForeignClients bounds the per-client table: it is keyed by an
// unauthenticated client address, so it must never grow with traffic. Ten
// thousand entries is a few hundred kilobytes and far more clients than a
// personal server sees in an hour; past that, the least recently charged
// entry is evicted (its allowance resets, which is the mild failure mode:
// the global budget still holds the line).
const maxForeignClients = 10000

// foreignBudget rations the identifiers a server resolves upstream on behalf
// of anonymous visitors: perClient charges per client address and global
// charges per process, both over the same rolling window. The zero value is
// unusable; see newForeignBudget. A nil *foreignBudget means the feature is
// off (catalog-only), and every method tolerates it.
type foreignBudget struct {
	perClient int
	global    int
	window    time.Duration
	now       func() time.Time // injectable clock, for the tests

	mu      sync.Mutex
	clients map[string]*foreignSpend
	all     foreignSpend
}

// foreignSpend is one rolling window of charges: the timestamps of the
// identifiers charged to a client (or to the whole process), oldest first.
type foreignSpend struct{ at []time.Time }

// charged returns how much of the window is spent, dropping what has aged out
// (the slice is append-ordered, so the expired part is always a prefix).
func (s *foreignSpend) charged(cutoff time.Time) int {
	i := 0
	for i < len(s.at) && s.at[i].Before(cutoff) {
		i++
	}
	if i > 0 {
		s.at = append(s.at[:0], s.at[i:]...)
	}
	return len(s.at)
}

// last is the most recent charge, the zero time for an empty window.
func (s *foreignSpend) last() time.Time {
	if len(s.at) == 0 {
		return time.Time{}
	}
	return s.at[len(s.at)-1]
}

// newForeignBudget returns a budget of perClient identifiers per client and
// global identifiers per process, per window. A perClient of zero returns nil:
// the feature is off and the catalog-only policy applies.
func newForeignBudget(perClient, global int, window time.Duration) *foreignBudget {
	if perClient <= 0 {
		return nil
	}
	if global < perClient {
		global = perClient // a global ceiling below one client's share is a typo
	}
	return &foreignBudget{
		perClient: perClient,
		global:    global,
		window:    window,
		now:       time.Now,
		clients:   map[string]*foreignSpend{},
	}
}

// charge accounts n new identifiers to client, all or nothing: it returns
// errForeignBudget (wrapped, naming which budget gave way) and charges
// nothing when either the client's window or the server's cannot take them.
func (b *foreignBudget) charge(client string, n int) error {
	if b == nil || n <= 0 {
		return nil
	}
	now := b.now()
	cutoff := now.Add(-b.window)

	b.mu.Lock()
	defer b.mu.Unlock()
	if b.all.charged(cutoff)+n > b.global {
		return fmt.Errorf("%w (server-wide)", errForeignBudget)
	}
	spend := b.clients[client]
	if spend == nil {
		b.evictLocked(cutoff)
		spend = &foreignSpend{}
		b.clients[client] = spend
	}
	if spend.charged(cutoff)+n > b.perClient {
		return errForeignBudget
	}
	for range n {
		spend.at = append(spend.at, now)
		b.all.at = append(b.all.at, now)
	}
	return nil
}

// evictLocked makes room for one new client entry, and only bothers once the
// table is full: it then drops every window that has fully aged out and, if
// that freed nothing, the least recently charged entry. Called with b.mu held.
func (b *foreignBudget) evictLocked(cutoff time.Time) {
	if len(b.clients) < maxForeignClients {
		return
	}
	for key, spend := range b.clients {
		if spend.charged(cutoff) == 0 {
			delete(b.clients, key)
		}
	}
	if len(b.clients) < maxForeignClients {
		return
	}
	oldestKey, oldest := "", time.Time{}
	for key, spend := range b.clients {
		if at := spend.last(); oldest.IsZero() || at.Before(oldest) {
			oldestKey, oldest = key, at
		}
	}
	delete(b.clients, oldestKey)
}

// foreignGate is one request's authority over the identifiers it carries that
// the bundled catalog does not know. A nil gate is the catalog-only policy:
// nothing outside marketdata.KnownLocal is accepted, which is what every
// non-web caller and a server started with -serve-foreign-per-hour=0 gets.
type foreignGate struct {
	budget *foreignBudget
	client string               // the requesting address (clientIP)
	cached func(id string) bool // already quotable offline, hence free; nil = nothing is
}

// accepts reports whether a foreign identifier is a candidate at all: the
// feature is on and the identifier has the shape of an ISIN or a ticker.
// Judging the shape before anything else keeps junk out of the fetch path and
// out of the budget.
func (g *foreignGate) accepts(id string) bool {
	return g != nil && g.budget != nil && marketdata.PlausibleID(id)
}

// charge accounts the foreign identifiers of one portfolio, deduplicated and
// minus those already in the quote cache (which cost no upstream request). It
// returns an errForeignBudget error when the allowance is spent, and nothing
// is charged in that case.
func (g *foreignGate) charge(ids []string) error {
	if g == nil || g.budget == nil || len(ids) == 0 {
		return nil
	}
	seen := map[string]bool{}
	n := 0
	for _, id := range ids {
		key := marketdata.CanonicalID(id)
		if seen[key] {
			continue
		}
		seen[key] = true
		if g.cached != nil && g.cached(key) {
			continue
		}
		n++
	}
	return g.budget.charge(g.client, n)
}
