// Visitor preferences: the pofo_prefs cookie remembers the composer's
// non-sensitive defaults (base currency, rebalance cadence, sim on/off).
// The cookie only ever pre-fills the hub form and its row links; /view
// rendering never reads it, so a shared /view URL reproduces exactly for
// every visitor (the M1 URL-as-state invariant).
package main

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	prefsCookie = "pofo_prefs"
	prefsMaxAge = 365 * 24 * 60 * 60 // one year, in seconds
)

// prefs carries the remembered defaults. A nil field means "not stored";
// consumers then fall back to the server defaults.
type prefs struct {
	currency  *string // ISO code, or "" = keep native currencies
	rebalance *int    // days, 0 = never
	sim       *bool   // true = SIM-extended history
}

// readPrefs decodes and validates the cookie. Values are attacker-shaped
// input like any header: invalid fields are dropped field-wise and an
// unparseable cookie yields the zero prefs, never an error.
func readPrefs(r *http.Request) prefs {
	var p prefs
	c, err := r.Cookie(prefsCookie)
	if err != nil {
		return p
	}
	v, err := url.ParseQuery(c.Value)
	if err != nil {
		return p
	}
	if vs, ok := v["currency"]; ok && len(vs) > 0 && validCurrency(vs[0]) {
		p.currency = &vs[0]
	}
	if vs, ok := v["rebalance"]; ok && len(vs) > 0 {
		if n, err := strconv.Atoi(vs[0]); err == nil && n >= 0 {
			p.rebalance = &n
		}
	}
	if vs, ok := v["sim"]; ok && len(vs) > 0 && (vs[0] == "on" || vs[0] == "off") {
		on := vs[0] == "on"
		p.sim = &on
	}
	return p
}

// validCurrency accepts the empty string (keep native currencies) or a
// three-letter uppercase ISO code.
func validCurrency(s string) bool {
	if s == "" {
		return true
	}
	if len(s) != 3 {
		return false
	}
	for _, r := range s {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}

// merge layers the request's explicit, valid global parameters over p and
// reports whether anything changed, so /view can refresh the cookie only
// when a URL actually carries new preferences.
func (p prefs) merge(q url.Values) (prefs, bool) {
	changed := false
	if vs, ok := q["currency"]; ok && len(vs) > 0 {
		cur := strings.ToUpper(strings.TrimSpace(vs[0]))
		if validCurrency(cur) && (p.currency == nil || *p.currency != cur) {
			p.currency, changed = &cur, true
		}
	}
	if vs, ok := q["rebalance"]; ok && len(vs) > 0 {
		if n, err := strconv.Atoi(vs[0]); err == nil && n >= 0 && (p.rebalance == nil || *p.rebalance != n) {
			p.rebalance, changed = &n, true
		}
	}
	if v := q.Get("sim"); v == "on" || v == "off" {
		on := v == "on"
		if p.sim == nil || *p.sim != on {
			p.sim, changed = &on, true
		}
	}
	return p, changed
}

// cookie encodes p for Set-Cookie. HttpOnly: nothing in the front end reads
// it; the hub renders the stored values server-side.
func (p prefs) cookie() *http.Cookie {
	v := url.Values{}
	if p.currency != nil {
		v.Set("currency", *p.currency)
	}
	if p.rebalance != nil {
		v.Set("rebalance", strconv.Itoa(*p.rebalance))
	}
	if p.sim != nil {
		s := "off"
		if *p.sim {
			s = "on"
		}
		v.Set("sim", s)
	}
	return &http.Cookie{
		Name:     prefsCookie,
		Value:    v.Encode(),
		Path:     "/",
		MaxAge:   prefsMaxAge,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
	}
}
