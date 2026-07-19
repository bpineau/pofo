package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func prefsRequest(cookie string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	if cookie != "" {
		r.AddCookie(&http.Cookie{Name: prefsCookie, Value: cookie})
	}
	return r
}

func TestReadPrefs(t *testing.T) {
	p := readPrefs(prefsRequest("currency=USD&rebalance=30&sim=off"))
	if p.currency == nil || *p.currency != "USD" {
		t.Errorf("currency = %v, want USD", p.currency)
	}
	if p.rebalance == nil || *p.rebalance != 30 {
		t.Errorf("rebalance = %v, want 30", p.rebalance)
	}
	if p.sim == nil || *p.sim != false {
		t.Errorf("sim = %v, want off", p.sim)
	}

	// No cookie: zero prefs.
	if p := readPrefs(prefsRequest("")); p.currency != nil || p.rebalance != nil || p.sim != nil {
		t.Errorf("no cookie: got %+v, want zero prefs", p)
	}

	// Invalid fields are dropped field-wise, valid ones kept.
	p = readPrefs(prefsRequest("currency=DROP%20TABLE&rebalance=-4&sim=on"))
	if p.currency != nil || p.rebalance != nil {
		t.Errorf("invalid fields kept: %+v", p)
	}
	if p.sim == nil || *p.sim != true {
		t.Error("valid sim dropped alongside invalid fields")
	}

	// An alien cookie value is ignored wholesale.
	if p := readPrefs(prefsRequest("%zz;bad")); p.currency != nil || p.rebalance != nil || p.sim != nil {
		t.Errorf("alien cookie: got %+v, want zero prefs", p)
	}
}

func TestValidCurrency(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"", true},    // keep native currencies
		{"EUR", true}, // canonical ISO code
		{"USD", true},
		{"eur", false},  // must be uppercase
		{"EU", false},   // too short
		{"EURO", false}, // too long
		{"E1R", false},  // digits are not letters
		{"E R", false},  // spaces are not letters
		{"DROP", false}, // an injection attempt
	}
	for _, c := range cases {
		if got := validCurrency(c.in); got != c.want {
			t.Errorf("validCurrency(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestPrefsMerge(t *testing.T) {
	base := readPrefs(prefsRequest("currency=USD&rebalance=30"))

	// A partial URL updates only its own fields.
	got, changed := base.merge(url.Values{"sim": {"off"}})
	if !changed {
		t.Fatal("merge with a new field must report changed")
	}
	if *got.currency != "USD" || *got.rebalance != 30 || *got.sim != false {
		t.Errorf("merge = %+v", got)
	}

	// Same values again: unchanged.
	if _, changed := got.merge(url.Values{"currency": {"USD"}, "sim": {"off"}}); changed {
		t.Error("identical merge must not report changed")
	}

	// Invalid values never merge.
	if _, changed := base.merge(url.Values{"currency": {"nope!"}, "rebalance": {"-1"}}); changed {
		t.Error("invalid values must not merge")
	}

	// Empty currency is a valid explicit value (keep native currencies).
	got, changed = base.merge(url.Values{"currency": {""}})
	if !changed || got.currency == nil || *got.currency != "" {
		t.Errorf("empty currency merge = %+v changed=%v", got, changed)
	}

	// The /view sentinel currency=native (any case) stores as the empty ISO
	// code, so the codec round-trips it as "keep native currencies".
	for _, v := range []string{"native", "NATIVE", "Native"} {
		got, changed = base.merge(url.Values{"currency": {v}})
		if !changed || got.currency == nil || *got.currency != "" {
			t.Errorf("currency=%s merge = %+v changed=%v, want stored empty", v, got, changed)
		}
	}
}

func TestPrefsCookieRoundTrip(t *testing.T) {
	in := readPrefs(prefsRequest("currency=GBP&rebalance=0&sim=on"))
	c := in.cookie()
	if c.Name != prefsCookie || c.Path != "/" || c.MaxAge <= 0 || !c.HttpOnly || c.SameSite != http.SameSiteLaxMode {
		t.Errorf("cookie attributes: %+v", c)
	}
	out := readPrefs(prefsRequest(c.Value))
	if *out.currency != "GBP" || *out.rebalance != 0 || *out.sim != true {
		t.Errorf("round trip = %+v", out)
	}
}
