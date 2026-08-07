package marketdata

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// fredHTTP is a dedicated HTTP/1.1 client for FRED: the endpoint returns an
// HTTP/2 INTERNAL_ERROR stream reset to Go's default transport, so HTTP/2 is
// disabled here (a nil-but-present TLSNextProto map turns it off).
var fredHTTP = &http.Client{
	Timeout:   12 * time.Second,
	Transport: &http.Transport{TLSNextProto: map[string]func(string, *tls.Conn) http.RoundTripper{}},
}

// fredUserAgent is the one place the client does NOT send its browser
// User-Agent, and the exception is measured. FRED's edge answers a browser
// User-Agent over HTTP/1.1 by accepting the connection and then never sending
// headers: the request hangs until the timeout, on every attempt (2026-08).
// The same request with any non-browser agent answers in under 100 ms, and the
// browser agent over HTTP/2 fails the other way (the INTERNAL_ERROR above), so
// HTTP/1.1 plus a plain agent is the only combination that works. Left
// unfixed, this silently emptied every FRED-backed series: the effective
// federal funds rate fell back to the New York Fed's copy and lost its
// 1954-2000 history, and the CPI fallbacks were unreachable.
const fredUserAgent = "pofo/1 (+https://github.com/bpineau/pofo)"

// fetchFRED downloads a monthly (or daily) series from the FRED CSV endpoint,
// a free, key-less source. It is used for long macro histories the market-data
// providers do not carry, chiefly French consumer prices (monthly, 1955→) that
// extend ^HICP-FR. Best-effort with a short timeout, since it is a slow-changing
// series fetched at most once per cache period. Rows with a missing value (".")
// are skipped; dates are FRED's YYYY-MM-DD.
func (c *Client) fetchFRED(ctx context.Context, id string) ([]Point, error) {
	return c.fredSeries(ctx, id, false)
}

// fetchFREDRate is fetchFRED for a RATE series, where zero and negative values
// are readings of the world rather than bad prints: the euro area policy rate
// sat at -0.50 % for eight years, and the Fed target floor has been 0 twice.
// An index level (CPI) keeps the positive-only filter, which protects it from
// the provider's placeholder rows.
func (c *Client) fetchFREDRate(ctx context.Context, id string) ([]Point, error) {
	return c.fredSeries(ctx, id, true)
}

func (c *Client) fredSeries(ctx context.Context, id string, keepNonPositive bool) ([]Point, error) {
	u := fmt.Sprintf("%s/graph/fredgraph.csv?id=%s", c.FredBase, url.QueryEscape(id))
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", fredUserAgent)
	req.Header.Set("Accept", "text/csv,*/*")
	resp, err := fredHTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fred %s: HTTP %d", id, resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	if !bytes.HasPrefix(body, []byte("observation_date,")) {
		return nil, fmt.Errorf("fred %s: unexpected response", id)
	}
	rows, err := csv.NewReader(bytes.NewReader(body)).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("fred %s: unreadable CSV: %w", id, err)
	}
	pts := make([]Point, 0, len(rows))
	for _, r := range rows[1:] { // skip the header
		if len(r) < 2 || r[1] == "." || r[1] == "" {
			continue
		}
		t, terr := time.ParseInLocation("2006-01-02", r[0], time.UTC)
		if terr != nil {
			continue
		}
		v, verr := strconv.ParseFloat(r[1], 64)
		if verr != nil || (v <= 0 && !keepNonPositive) {
			continue
		}
		pts = append(pts, Point{Date: t, Close: v})
	}
	if len(pts) < 2 {
		return nil, fmt.Errorf("fred %s: only %d usable points", id, len(pts))
	}
	return pts, nil
}
