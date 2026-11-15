package remote

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"slices"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
)

var catalog = sync.OnceValue(datasets.Catalog)
var extensions = sync.OnceValue(func() map[string]bool {
	out := map[string]bool{}
	for _, id := range marketdata.BundledIDs() {
		base, sim := marketdata.SplitSim(id)
		if sim {
			out[marketdata.CanonicalID(base)] = true
		}
	}
	return out
})

func identity(a datasets.Asset) Identity {
	i := Identity{ID: a.ID, Name: a.Name, ISIN: a.ISIN, AssetClass: a.AssetClass, QuoteCurrency: a.Currency, ExtendedHistoryAvailable: extensions()[a.ID]}
	if i.ISIN == "" && marketdata.IsISIN(a.ID) {
		i.ISIN = a.ID
	}
	if a.Fees > 0 {
		i.TERPct = ptr(a.Fees)
	}
	return i
}

type cursor struct {
	Offset   int    `json:"o"`
	Revision string `json:"r"`
	Filter   string `json:"f"`
}

func (s *Service) search(in SearchInput, r *Result) (SearchData, error) {
	out := SearchData{Items: []Identity{}}
	limit := def(in.Limit, 10, "limit", r)
	if utf8.RuneCountInString(in.Query) > 200 || limit < 1 || limit > 50 {
		return out, invalid("query/limit", "Query is limited to 200 characters; limit must be 1 through 50.")
	}
	classes, currencies := map[string]bool{}, map[string]bool{}
	for _, a := range catalog() {
		classes[a.AssetClass] = true
		currencies[a.Currency] = true
	}
	if in.AssetClass != "" && !classes[in.AssetClass] {
		return out, invalid("asset_class", "Unknown catalog asset class.")
	}
	if in.Currency != "" && !currencies[in.Currency] {
		return out, invalid("currency", "Unknown catalog quote currency.")
	}
	q := strings.ToLower(strings.TrimSpace(in.Query))
	h := sha256.Sum256([]byte(q + "\x00" + in.AssetClass + "\x00" + in.Currency))
	finger := hex.EncodeToString(h[:8])
	rev := s.Version + ":" + revision()
	offset := 0
	if in.Cursor != "" {
		b, e := base64.RawURLEncoding.DecodeString(in.Cursor)
		var c cursor
		if e != nil || json.Unmarshal(b, &c) != nil || c.Revision != rev || c.Filter != finger || c.Offset < 0 {
			return out, invalid("cursor", "Invalid or expired cursor; restart pagination.")
		}
		offset = c.Offset
	}
	resolved, known := marketdata.Lookup(in.Query)
	var exact, rest []datasets.Asset
	for _, a := range catalog() {
		if in.AssetClass != "" && a.AssetClass != in.AssetClass || in.Currency != "" && a.Currency != in.Currency {
			continue
		}
		match := known && a.ID == resolved.ID || strings.EqualFold(a.ID, q) || q != "" && strings.EqualFold(a.ISIN, q) || slices.ContainsFunc(a.Aliases, func(v string) bool { return strings.EqualFold(v, q) })
		if match {
			exact = append(exact, a)
		} else if q == "" || strings.Contains(strings.ToLower(a.Name), q) || strings.Contains(strings.ToLower(a.ID), q) {
			rest = append(rest, a)
		}
	}
	sortAssets := func(xs []datasets.Asset) {
		slices.SortFunc(xs, func(a, b datasets.Asset) int { return strings.Compare(a.ID, b.ID) })
	}
	sortAssets(exact)
	sortAssets(rest)
	all := append(exact, rest...)
	if offset > len(all) {
		return out, invalid("cursor", "Cursor is outside this catalog page.")
	}
	end := min(offset+limit, len(all))
	for _, a := range all[offset:end] {
		out.Items = append(out.Items, identity(a))
	}
	if end < len(all) {
		b, _ := json.Marshal(cursor{end, rev, finger})
		out.NextCursor = base64.RawURLEncoding.EncodeToString(b)
	}
	return out, nil
}
