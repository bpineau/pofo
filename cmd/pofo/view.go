// /view URL parsing: translating a shareable query string into portfolio
// specs plus overrides for the server's default options. Each parsed
// portfolio section also carries a Simulate link to its FIRE mount
// (/firesimulator/e/<name>/ for an example, /firesimulator/p/<spec>/ for a p= spec).
//
// Under -serve the live composer (composer.go) is a front end that edits
// exactly this grammar in-page, keeping the URL equal to the edited state.
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
	currency   *string // nil = no override; non-nil, "" = keep native currencies
	bench      *string
	noSim      *bool
	// fireHrefs maps each spec's (deduplicated) name to its FIRE simulator
	// link: /firesimulator/e/<name>/ for an embedded example,
	// /firesimulator/p/<escaped spec>/ for an ad-hoc p= portfolio. Only the
	// web report consumes it.
	fireHrefs map[string]string
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
	if vr.currency != nil {
		o.currency = *vr.currency
	}
	if vr.bench != nil {
		o.benchmark = *vr.bench
	}
	if vr.noSim != nil {
		o.noSim = *vr.noSim
	}
	o.fireHref = vr.fireHrefs
	o.composer = composerMount(vr)
	return &o
}

// parseViewQuery translates a /view query into a viewRequest. Portfolio
// parsing is delegated to pkg/portfolio by rebuilding the file text form,
// so the URL grammar can never drift from the file grammar. Each parsed
// portfolio also gets a FIRE simulator link recorded in fireHrefs, keyed by
// its final (deduplicated) name: /firesimulator/e/<name>/ for an embedded
// example, /firesimulator/p/<escaped spec>/ for an ad-hoc p= portfolio. The rendered /view
// report surfaces these as per-section "Simulate" links.
func parseViewQuery(q url.Values, base *options) (*viewRequest, error) {
	vr := &viewRequest{fireHrefs: map[string]string{}}
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
		vr.fireHrefs[spec.Name] = fireBase + "/e/" + name + "/"
	}
	for i, raw := range ps {
		spec, err := adhocSpec(raw, i+1)
		if err != nil {
			return nil, err
		}
		add(spec)
		vr.fireHrefs[spec.Name] = fireBase + "/p/" + url.PathEscape(raw) + "/"
	}
	if err := parseViewGlobals(q, vr, base); err != nil {
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
//
// Two parameters carry attacker-shaped identifiers that would otherwise reach
// an outbound fetch on behalf of an anonymous visitor, so both are gated:
//
//   - currency: absent leaves the server default; "native" (any case) or a
//     present-but-empty value keeps each series in its native currency;
//     otherwise a three-letter ISO code (validCurrency). Anything else is a
//     400, so no arbitrary bytes reach an FX fetch URL or mint an unbounded
//     cache file.
//   - bench: empty explicitly disables Beta; otherwise it must be a locally
//     resolvable identifier (marketdata.KnownLocal) or the exact server
//     default benchmark (a quote symbol like ^GSPC, which is not "local"), so
//     the shared quote cache can never be poisoned with an arbitrary symbol.
func parseViewGlobals(q url.Values, vr *viewRequest, base *options) error {
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
	if vs, ok := q["currency"]; ok && len(vs) > 0 {
		cur := strings.ToUpper(strings.TrimSpace(vs[0]))
		switch {
		case cur == "NATIVE":
			native := ""
			vr.currency = &native
		case validCurrency(cur):
			c := cur
			vr.currency = &c
		default:
			return fmt.Errorf("invalid currency %q (want a 3-letter ISO code or native)", cur)
		}
	}
	if vs, ok := q["bench"]; ok && len(vs) > 0 {
		v := vs[0]
		if v != "" && !marketdata.KnownLocal(v) && v != base.benchmark {
			return fmt.Errorf("benchmark %q is not in the local catalog (empty disables it)", v)
		}
		vr.bench = &v
	}
	switch v := q.Get("sim"); v {
	case "":
	case "on":
		// The web's global sim toggle is a POSITIVE control: turn the
		// backcast on for every portfolio on the page. The CLI opts in per
		// file with "#meta sim:on"; a composed p= card carries no such line,
		// so without this the toggle could only ever turn sim off (noSim),
		// never on, and a live portfolio would show no reconstructed history.
		// Build fetches each holding's SIM variant; holdings with no backcast
		// fall back to real quotes, so this is safe to apply uniformly.
		off := false
		vr.noSim = &off
		for _, s := range vr.specs {
			s.Sim = true
		}
	case "off":
		off := true
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
