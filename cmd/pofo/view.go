// /view URL parsing: translating a shareable query string into portfolio
// specs plus overrides for the server's default options.
package main

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/bpineau/pofo/examples"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/portfolio"
)

// /view guardrails: the composer is meant for humans on a small server.
const (
	maxViewPortfolios = 6
	maxViewHoldings   = 20
	maxViewSpecLen    = 2000
)

// viewRequest is a parsed /view query: the portfolios to compare and the
// global overrides to layer on the server's default options.
type viewRequest struct {
	specs      []*portfolio.Spec
	start, end time.Time
	rebalance  *int
	currency   string
	bench      *string
	noSim      *bool
}

// serverOptions returns a copy of the server defaults with this request's
// overrides applied, ready for renderComparison.
func (vr *viewRequest) serverOptions(base *options) *options {
	o := *base
	o.cli, o.out, o.noOpen = false, "", true
	o.web = true // warm skin + site nav in the rendered report
	if !vr.start.IsZero() {
		o.start = vr.start
	}
	if !vr.end.IsZero() {
		o.end = vr.end
	}
	if vr.rebalance != nil {
		o.rebalance = *vr.rebalance
	}
	if vr.currency != "" {
		o.currency = vr.currency
	}
	if vr.bench != nil {
		o.benchmark = *vr.bench
	}
	if vr.noSim != nil {
		o.noSim = *vr.noSim
	}
	return &o
}

// parseViewQuery translates a /view query into a viewRequest. Portfolio
// parsing is delegated to pkg/portfolio by rebuilding the file text form,
// so the URL grammar can never drift from the file grammar.
func parseViewQuery(q url.Values) (*viewRequest, error) {
	vr := &viewRequest{}
	exs, ps := q["ex"], q["p"]
	if len(exs)+len(ps) > maxViewPortfolios {
		return nil, fmt.Errorf("at most %d portfolios per page", maxViewPortfolios)
	}
	known := knownExamples()
	nameCount := map[string]int{}
	add := func(spec *portfolio.Spec) {
		nameCount[spec.Name]++
		if n := nameCount[spec.Name]; n > 1 {
			spec.Name = fmt.Sprintf("%s (%d)", spec.Name, n)
		}
		vr.specs = append(vr.specs, spec)
	}
	for _, name := range exs {
		if _, ok := known[name]; !ok {
			return nil, fmt.Errorf("unknown example %q", name)
		}
		raw, err := examples.FS.ReadFile(name + ".txt")
		if err != nil {
			return nil, fmt.Errorf("unknown example %q", name)
		}
		spec, err := portfolio.Parse(name, strings.NewReader(string(raw)))
		if err != nil {
			return nil, fmt.Errorf("example %s: %w", name, err)
		}
		add(spec)
	}
	for i, raw := range ps {
		spec, err := adhocSpec(raw, i+1)
		if err != nil {
			return nil, err
		}
		add(spec)
	}
	if err := parseViewGlobals(q, vr); err != nil {
		return nil, err
	}
	return vr, nil
}

// adhocSpec parses one p= value: "ID:WEIGHT,ID:WEIGHT[!meta:value]...".
// The '!' meta delimiter keeps a shareable link hand-typable (a raw ';'
// is invalid in a Go query string). It rebuilds the portfolio file text
// and feeds portfolio.Parse; only locally-resolvable identifiers are
// accepted (no network on behalf of anonymous visitors). Control
// characters (notably a URL-decoded newline) are rejected up front:
// since the rebuilt text is line-based, a smuggled newline would inject
// an extra holding line that bypasses both the catalog gate and the
// holdings-count limit below.
func adhocSpec(raw string, n int) (*portfolio.Spec, error) {
	if len(raw) > maxViewSpecLen {
		return nil, fmt.Errorf("p parameter too long (%d bytes, max %d)", len(raw), maxViewSpecLen)
	}
	if strings.IndexFunc(raw, func(r rune) bool { return r < 0x20 || r == 0x7f }) >= 0 {
		return nil, fmt.Errorf("control characters in p parameter")
	}
	segments := strings.Split(raw, "!")
	name := fmt.Sprintf("adhoc-%d", n)
	var text strings.Builder
	for _, meta := range segments[1:] {
		meta = strings.TrimSpace(meta)
		if meta == "" {
			continue
		}
		if v, ok := strings.CutPrefix(meta, "name:"); ok {
			if v = strings.TrimSpace(v); v != "" {
				name = v
			}
			continue
		}
		fmt.Fprintf(&text, "#meta %s\n", meta)
	}
	pairs := strings.Split(segments[0], ",")
	if len(pairs) > maxViewHoldings {
		return nil, fmt.Errorf("at most %d holdings per portfolio", maxViewHoldings)
	}
	for _, pair := range pairs {
		id, weight, ok := strings.Cut(strings.TrimSpace(pair), ":")
		if !ok || id == "" || weight == "" {
			return nil, fmt.Errorf("malformed holding %q, want ID:WEIGHT (decimal point, no comma)", pair)
		}
		if !marketdata.KnownLocal(id) {
			return nil, fmt.Errorf("identifier not in the local catalog: %s", id)
		}
		fmt.Fprintf(&text, "%s %s\n", weight, id)
	}
	spec, err := portfolio.Parse(name, strings.NewReader(text.String()))
	if err != nil {
		return nil, fmt.Errorf("portfolio %s: %w", name, err)
	}
	return spec, nil
}

// parseViewGlobals fills the option overrides from the query's global
// parameters, mirroring the CLI flags they correspond to.
func parseViewGlobals(q url.Values, vr *viewRequest) error {
	if v := q.Get("start"); v != "" {
		t, err := time.ParseInLocation("2006-01-02", v, time.UTC)
		if err != nil {
			return fmt.Errorf("invalid start date %q (want YYYY-MM-DD)", v)
		}
		vr.start = t
	}
	if v := q.Get("end"); v != "" {
		t, err := time.ParseInLocation("2006-01-02", v, time.UTC)
		if err != nil {
			return fmt.Errorf("invalid end date %q (want YYYY-MM-DD)", v)
		}
		if !vr.start.IsZero() && !t.After(vr.start) {
			return fmt.Errorf("end must be after start")
		}
		vr.end = t
	}
	if v := q.Get("rebalance"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			return fmt.Errorf("invalid rebalance %q (want a day count, 0 = never)", v)
		}
		vr.rebalance = &n
	}
	vr.currency = strings.ToUpper(strings.TrimSpace(q.Get("currency")))
	if vs, ok := q["bench"]; ok && len(vs) > 0 {
		v := vs[0]
		vr.bench = &v
	}
	switch v := q.Get("sim"); v {
	case "":
	case "on", "off":
		off := v == "off"
		vr.noSim = &off
	default:
		return fmt.Errorf("invalid sim value %q (on or off)", v)
	}
	return nil
}

// knownExamples indexes the embedded portfolio files by name.
func knownExamples() map[string]examples.Info {
	byName := map[string]examples.Info{}
	for _, in := range examples.List() {
		byName[in.Name] = in
	}
	return byName
}
