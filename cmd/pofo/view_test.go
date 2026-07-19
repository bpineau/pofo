package main

import (
	"net/url"
	"strings"
	"testing"
	"time"
)

func mustQuery(t *testing.T, raw string) url.Values {
	t.Helper()
	q, err := url.ParseQuery(raw)
	if err != nil {
		t.Fatal(err)
	}
	return q
}

func TestParseViewQueryExamples(t *testing.T) {
	vr, err := parseViewQuery(mustQuery(t, "ex=claude-dragonlite&ex=efficient-core-9060"))
	if err != nil {
		t.Fatal(err)
	}
	if len(vr.specs) != 2 {
		t.Fatalf("specs = %d, want 2", len(vr.specs))
	}
	if vr.specs[0].Name != "claude-dragonlite" {
		t.Errorf("name = %q", vr.specs[0].Name)
	}
	if len(vr.specs[0].Holdings) == 0 {
		t.Error("no holdings parsed from the embedded file")
	}
}

func TestParseViewQueryAdhoc(t *testing.T) {
	vr, err := parseViewQuery(mustQuery(t,
		"p=NTSG:60,IGLN:20,IBCI:20!rebalance:30!sim:on!name:essai"))
	if err != nil {
		t.Fatal(err)
	}
	s := vr.specs[0]
	if s.Name != "essai" {
		t.Errorf("name = %q, want essai", s.Name)
	}
	if len(s.Holdings) != 3 {
		t.Fatalf("holdings = %d, want 3", len(s.Holdings))
	}
	if s.Holdings[0].ID != "NTSG" || s.Holdings[0].RawWeight != 60 {
		t.Errorf("holding[0] = %+v", s.Holdings[0])
	}
	if s.RebalanceDays != 30 {
		t.Errorf("RebalanceDays = %d, want 30", s.RebalanceDays)
	}
	if !s.Sim {
		t.Error("sim:on not applied")
	}
}

func TestParseViewQueryGlobals(t *testing.T) {
	vr, err := parseViewQuery(mustQuery(t,
		"ex=claude-dragonlite&start=2015-07-18&end=2026-06-30&rebalance=180&currency=USD&bench=&sim=off"))
	if err != nil {
		t.Fatal(err)
	}
	if vr.start != time.Date(2015, 7, 18, 0, 0, 0, 0, time.UTC) {
		t.Errorf("start = %v", vr.start)
	}
	if vr.rebalance == nil || *vr.rebalance != 180 {
		t.Error("rebalance override missing")
	}
	if vr.currency != "USD" {
		t.Errorf("currency = %q", vr.currency)
	}
	if vr.bench == nil || *vr.bench != "" {
		t.Error("bench= (empty) must disable Beta, not keep the default")
	}
	if vr.noSim == nil || !*vr.noSim {
		t.Error("sim=off must set noSim")
	}
	opt := &options{currency: "EUR", benchmark: "^GSPC", rebalance: 90}
	o := vr.serverOptions(opt)
	if o.currency != "USD" || o.benchmark != "" || o.rebalance != 180 || !o.noSim {
		t.Errorf("serverOptions = %+v", o)
	}
	if opt.currency != "EUR" {
		t.Error("serverOptions must not mutate the base options")
	}
}

func TestParseViewQueryErrors(t *testing.T) {
	cases := []struct{ name, raw, wantErr string }{
		{"unknown example", "ex=nope", "unknown example"},
		{"traversal", "ex=../../etc/passwd", "unknown example"},
		{"unknown id", "p=ZZZNOTANID:100", "not in the local catalog"},
		{"quote symbol", "p=^GSPC:100", "not in the local catalog"},
		{"malformed pair", "p=NTSG", "ID:WEIGHT"},
		{"too many portfolios", "ex=claude-dragonlite&ex=claude-dragonlite&ex=claude-dragonlite&ex=claude-dragonlite&ex=claude-dragonlite&ex=claude-dragonlite&ex=claude-dragonlite", "at most 6"},
		{"too many holdings", "p=" + strings.Repeat("NTSG:1,", 20) + "NTSG:1", "at most 20"},
		{"bad start", "ex=claude-dragonlite&start=notadate", "start"},
		{"end before start", "ex=claude-dragonlite&start=2020-01-01&end=2019-01-01", "after"},
		{"bad rebalance", "ex=claude-dragonlite&rebalance=x", "rebalance"},
		{"bad sim", "ex=claude-dragonlite&sim=maybe", "sim"},
		{"newline injection", "p=NTSG:100!rebalance:30%0A50%20BADIDXYZ", "control character"},
		{"carriage return", "p=NTSG:100%0D:1", "control character"},
	}
	for _, c := range cases {
		_, err := parseViewQuery(mustQuery(t, c.raw))
		if err == nil || !strings.Contains(err.Error(), c.wantErr) {
			t.Errorf("%s: err = %v, want containing %q", c.name, err, c.wantErr)
		}
	}
}

func TestViewFireHrefs(t *testing.T) {
	vr, err := parseViewQuery(url.Values{
		"ex": {"claude-dragonlite"},
		"p":  {"IWDA:60,IGLN:40!sim:on"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := vr.fireHrefs["claude-dragonlite"]; got != "/fire/e/claude-dragonlite/" {
		t.Errorf("example href = %q", got)
	}
	adhoc := vr.specs[1].Name
	want := "/fire/p/" + url.PathEscape("IWDA:60,IGLN:40!sim:on") + "/"
	if got := vr.fireHrefs[adhoc]; got != want {
		t.Errorf("adhoc href = %q, want %q", got, want)
	}
	// The hrefs ride into the render options only for the web app.
	o := vr.serverOptions(&options{})
	if o.fireHref[adhoc] != want {
		t.Errorf("serverOptions fireHref = %q", o.fireHref[adhoc])
	}
}

func TestParseViewQueryEmpty(t *testing.T) {
	vr, err := parseViewQuery(mustQuery(t, ""))
	if err != nil {
		t.Fatal(err)
	}
	if len(vr.specs) != 0 {
		t.Error("empty query must yield zero specs (the handler redirects)")
	}
}

func TestParseViewQueryDuplicateNames(t *testing.T) {
	vr, err := parseViewQuery(mustQuery(t, "ex=claude-dragonlite&ex=claude-dragonlite"))
	if err != nil {
		t.Fatal(err)
	}
	if vr.specs[0].Name == vr.specs[1].Name {
		t.Error("duplicate names must be disambiguated like the CLI does")
	}
}
