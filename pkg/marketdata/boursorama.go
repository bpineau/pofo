package marketdata

import (
	"fmt"
	"html"
	"net/url"
	"regexp"
	"strings"
)

// boursoFundRe captures the Morningstar id and the displayed name of the
// first fund of a Boursorama search result page.
var boursoFundRe = regexp.MustCompile(
	`(?s)href="/bourse/(?:opcvm|trackers)/cours/(0P[0-9A-Za-z]+)/".*?search__item-title">\s*(.*?)\s*</`)

// morningstarIDRe is the fallback when the page layout changes: id only.
var morningstarIDRe = regexp.MustCompile(`/bourse/(?:opcvm|trackers)/cours/(0P[0-9A-Za-z]+)/`)

// boursoramaMorningstarID finds the Morningstar identifier (and name) of a
// fund by querying the Boursorama search with an ISIN. It is used as a last
// resort for ISINs that neither Yahoo nor the FT know.
func (c *Client) boursoramaMorningstarID(isin string) (id, name string, err error) {
	u := fmt.Sprintf("%s/recherche/ajax?query=%s", c.BoursoramaBase, url.QueryEscape(isin))
	body, err := c.do("GET", u, "", nil, map[string]string{"X-Requested-With": "XMLHttpRequest"})
	if err != nil {
		return "", "", err
	}
	page := string(body)
	if m := boursoFundRe.FindStringSubmatch(page); m != nil {
		return m[1], strings.TrimSpace(html.UnescapeString(m[2])), nil
	}
	if m := morningstarIDRe.FindStringSubmatch(page); m != nil {
		return m[1], "", nil
	}
	return "", "", fmt.Errorf("no Morningstar identifier for %s", isin)
}
