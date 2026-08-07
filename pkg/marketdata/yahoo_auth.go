package marketdata

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// yahooAuth is the cookie+crumb pair Yahoo's quote API requires: Yahoo hands
// the cookie on any page hit and derives the crumb from it. Both expire
// together, so they are cached and renewed as one unit.
type yahooAuth struct {
	cookie string // Cookie header value, e.g. "A3=d=…"
	crumb  string
}

// yahooAuthPair returns the cached cookie+crumb, performing the bootstrap
// dance on first use. Safe for concurrent callers.
func (c *Client) yahooAuthPair(ctx context.Context) (yahooAuth, error) {
	c.authMu.Lock()
	defer c.authMu.Unlock()
	if c.auth != nil {
		return *c.auth, nil
	}
	a, err := c.fetchYahooAuth(ctx)
	if err != nil {
		return yahooAuth{}, err
	}
	c.auth = &a
	return a, nil
}

// invalidateYahooAuth drops the cached pair after an HTTP 401/403: the next
// yahooAuthPair call fetches a fresh one.
func (c *Client) invalidateYahooAuth() {
	c.authMu.Lock()
	c.auth = nil
	c.authMu.Unlock()
}

// fetchYahooAuth hits CookieBase to collect the consent cookie - on the raw
// redirect response, which must not be followed or the cookie is lost - then
// trades it for the crumb at /v1/test/getcrumb.
func (c *Client) fetchYahooAuth(ctx context.Context) (yahooAuth, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.CookieBase, nil)
	if err != nil {
		return yahooAuth{}, err
	}
	req.Header.Set("User-Agent", c.UserAgent)
	hc := *c.HTTP // shallow copy: same transport, no redirect following
	hc.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := hc.Do(req)
	if err != nil {
		return yahooAuth{}, err
	}
	defer resp.Body.Close()
	parts := make([]string, 0, 2)
	for _, ck := range resp.Cookies() {
		parts = append(parts, ck.Name+"="+ck.Value)
	}
	if len(parts) == 0 {
		return yahooAuth{}, fmt.Errorf("yahoo auth: no cookie from %s", c.CookieBase)
	}
	cookie := strings.Join(parts, "; ")
	crumb, err := c.do(ctx, http.MethodGet, c.ChartBase+"/v1/test/getcrumb", "", nil,
		map[string]string{"Cookie": cookie})
	if err != nil {
		return yahooAuth{}, fmt.Errorf("yahoo auth: crumb: %w", err)
	}
	if len(crumb) == 0 || len(crumb) > 64 {
		return yahooAuth{}, fmt.Errorf("yahoo auth: implausible crumb %q", crumb)
	}
	return yahooAuth{cookie: cookie, crumb: string(crumb)}, nil
}
