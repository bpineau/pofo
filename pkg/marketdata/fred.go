package marketdata

import (
	"bytes"
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
	Timeout:   30 * time.Second,
	Transport: &http.Transport{TLSNextProto: map[string]func(string, *tls.Conn) http.RoundTripper{}},
}

// fetchFRED downloads a monthly (or daily) series from the FRED CSV endpoint,
// a free, key-less, reliable source. It is used for long macro histories the
// market-data providers do not carry: French and US consumer prices (decades of
// monthly data) and the USD/EUR reference rate. Rows with a missing value (".")
// are skipped; dates are FRED's YYYY-MM-DD.
func (c *Client) fetchFRED(id string) ([]Point, error) {
	u := fmt.Sprintf("%s/graph/fredgraph.csv?id=%s", c.FredBase, url.QueryEscape(id))
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.UserAgent)
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
		if verr != nil || v <= 0 {
			continue
		}
		pts = append(pts, Point{Date: t, Close: v})
	}
	if len(pts) < 2 {
		return nil, fmt.Errorf("fred %s: only %d usable points", id, len(pts))
	}
	return pts, nil
}
