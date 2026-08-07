// The /view live composer's server side: the mount markup, the embedded
// front-end assets, and specToP, which translates a spec into the p= grammar
// for the Fork feature.
package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"sort"
	"strconv"
	"strings"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/portfolio"
)

// composerJS and composerCSS are the live composer's front end, served at
// /composer.js and /composer.css. This task ships the skeleton (render the
// panel from the data attributes, toggle collapsed/open); the editing logic
// lands with the next task.
//
//go:embed composer.js
var composerJS string

//go:embed composer.css
var composerCSS string

// composerCaps mirrors the /view guardrails for the front end, so the editor
// can warn before a link is rejected server-side. The server gates stay
// authoritative.
type composerCaps struct {
	Ports    int `json:"ports"`
	Holdings int `json:"holdings"`
	Bytes    int `json:"bytes"`
}

// composerMount builds the live composer panel for one /view request: the
// collapsed chip bar, the open stack of portfolio cards, the globals row and
// the run bar, plus the data attributes the front end hydrates from. It is
// injected under the site nav on every web report (report.Page.Composer) and
// is empty for the standalone CLI report.
//
// Each ex= card is read-only. If its spec has at least one locally resolvable
// holding it carries a Fork affordance and a data-fork-<i> payload (the
// specToP serialization the editor forks into an editable p= portfolio);
// a spec whose holdings all dropped is unforkable and renders as a plain
// "not composable" note with no Fork button. Each p= card is already in the
// grammar, so it renders as editable rows directly.
//
// The whole panel is rendered through html/template into a buffer: portfolio
// names and identifiers are user input, and every value (including the
// encoding/json data attributes) rides the template's contextual escaping,
// never string concatenation.
func composerMount(vr *viewRequest) template.HTML {
	caps := composerCaps{Ports: maxViewPortfolios, Holdings: maxViewHoldings, Bytes: maxViewSpecLen}
	capsJSON, err := json.Marshal(caps)
	if err != nil {
		return "" // constant shape; cannot fail
	}

	names := localNames()
	data := composerData{
		Caps:     string(capsJSON),
		Count:    len(vr.specs),
		MaxBytes: maxViewSpecLen,
		Globals:  composerGlobals(vr),
	}
	for i, spec := range vr.specs {
		card := composerCard{Index: i, Name: spec.Name}
		isEx := strings.HasPrefix(vr.fireHrefs[spec.Name], "/fire/e/")
		p, dropped := specToP(spec)
		if isEx {
			card.Kind, card.KindLabel, card.ReadOnly = "ex", "ex=", true
			card.Holdings = displayHoldings(spec, names)
			// Forkable only if at least one holding survives into the grammar:
			// a spec that is all name/meta (its holdings all dropped) yields a
			// non-empty p yet an empty leading holdings segment, and forking it
			// would produce a portfolio with no holdings.
			holdings, _, _ := strings.Cut(p, "!")
			if holdings != "" {
				fork := composerFork{Name: spec.Name, P: p, Dropped: dropped}
				if b, err := json.Marshal(fork); err == nil {
					card.Forkable, card.Fork = true, string(b)
				}
			} else {
				card.NotComposable = true
			}
		} else {
			card.Kind, card.KindLabel, card.Editable = "p", "p=", true
			card.Holdings = displayHoldings(spec, names)
		}
		data.Cards = append(data.Cards, card)
		if !isEx {
			data.Open = true // an editable portfolio is present: open the panel
		}
	}

	var buf bytes.Buffer
	if err := composerTmpl.Execute(&buf, data); err != nil {
		return ""
	}
	return template.HTML(buf.String())
}

// composerData is the composerTmpl model.
type composerData struct {
	Caps     string // encoding/json of composerCaps, for the data-caps attribute
	Count    int
	Open     bool // render the panel expanded (an editable p= portfolio is present)
	MaxBytes int
	Globals  composerGlobal
	Cards    []composerCard
}

// composerCard is one portfolio row of the stack.
type composerCard struct {
	Index         int
	Kind          string // "ex" or "p" (the CSS variant)
	KindLabel     string // "ex=" or "p="
	Name          string
	ReadOnly      bool // ex= cards: the name is not editable
	Editable      bool // p= cards: holdings render as editable rows
	Holdings      []composerHolding
	Forkable      bool
	Fork          string // encoding/json of composerFork, for the data-fork-<i> attribute
	NotComposable bool   // ex= whose holdings all dropped: no Fork affordance
}

// composerHolding is one displayed holding (read-only readout or editable row).
type composerHolding struct {
	ID     string
	Name   string
	Weight string
}

// composerFork is the data-fork-<i> payload: the forkable serialization the
// editor turns into an editable p= portfolio, plus what could not ride the
// grammar.
type composerFork struct {
	Name    string   `json:"name"`
	P       string   `json:"p"`
	Dropped []string `json:"dropped,omitempty"`
}

// composerGlobal holds the current global settings shown in the globals row.
// Empty fields mean "server default", which the front end resolves.
type composerGlobal struct {
	Currency  string
	Rebalance string
	Sim       string
	Bench     string
	Start     string
	End       string
}

// composerGlobals reads the request's explicit global overrides. Absent
// overrides stay empty: the front end fills the effective defaults.
func composerGlobals(vr *viewRequest) composerGlobal {
	g := composerGlobal{}
	if vr.currency != nil {
		if *vr.currency == "" {
			g.Currency = "native"
		} else {
			g.Currency = *vr.currency
		}
	}
	if vr.rebalance != nil {
		g.Rebalance = strconv.Itoa(*vr.rebalance)
	}
	if vr.noSim != nil {
		if *vr.noSim {
			g.Sim = "off"
		} else {
			g.Sim = "on"
		}
	}
	if vr.bench != nil {
		g.Bench = *vr.bench
	}
	if !vr.start.IsZero() {
		g.Start = vr.start.Format("2006-01-02")
	}
	if !vr.end.IsZero() {
		g.End = vr.end.Format("2006-01-02")
	}
	return g
}

// displayHoldings renders a spec's holdings for the card, naming each id from
// the local catalog where known (the front end refines from /catalog.json).
func displayHoldings(spec *portfolio.Spec, names map[string]string) []composerHolding {
	out := make([]composerHolding, 0, len(spec.Holdings))
	for _, h := range spec.Holdings {
		out = append(out, composerHolding{
			ID:     h.ID,
			Name:   names[h.ID],
			Weight: strconv.FormatFloat(h.RawWeight, 'g', -1, 64),
		})
	}
	return out
}

// localNames maps every locally resolvable identifier (canonical and
// alternates) to its display name, for the composer's read-only readouts.
func localNames() map[string]string {
	m := map[string]string{}
	for _, a := range marketdata.LocalCatalog() {
		if a.Name == "" {
			continue
		}
		m[a.ID] = a.Name
		for _, alt := range a.Alt {
			m[alt] = a.Name
		}
	}
	return m
}

// composerTmpl renders the composer panel. The panel markup, class names and
// structure come from the chosen "stacked cards" mock; the benchmark control
// is a text input (the /view grammar accepts any locally known id), not the
// mock's select.
var composerTmpl = template.Must(template.New("composer").Parse(`<link rel="stylesheet" href="/composer.css">
<details class="cmp" id="composer" data-caps="{{.Caps}}"{{if .Open}} open{{end}}>
<summary class="cmp-bar">
<span class="eyebrow">Composer</span>
<div class="cmp-sum"><span class="chip"><b>{{.Count}}</b> portfolios</span>{{with .Globals.Currency}}<span class="chip">{{.}}</span>{{end}}{{with .Globals.Rebalance}}<span class="chip">rebalance <b>{{.}}d</b></span>{{end}}{{with .Globals.Sim}}<span class="chip">sim {{.}}</span>{{end}}</div>
<span class="grow"></span>
<span class="btn btn-ghost cmp-toggle"></span>
</summary>
<div class="cmp-panel">
<div class="stack">
{{range .Cards}}<div class="pcard"{{if .Forkable}} data-fork-{{.Index}}="{{.Fork}}"{{end}}>
<div class="pcard-head"><span class="kind {{.Kind}}">{{.KindLabel}}</span><input class="pname" value="{{.Name}}"{{if .ReadOnly}} readonly{{end}}><span class="grow"></span>{{if .Forkable}}<button class="fork" type="button">Fork to edit</button>{{end}}</div>
<div class="pcard-body">
{{if .Editable}}{{range .Holdings}}<div class="hrow"><div class="idbox"><input class="field" value="{{.ID}}"></div><span class="rn">{{.Name}}</span><input class="field wt" value="{{.Weight}}"><button class="rm" type="button">&times;</button></div>
{{end}}<button class="add" type="button">+ add holding</button>{{else}}<div class="exlist">{{range .Holdings}}<span><b>{{.ID}}</b> {{.Weight}}</span>{{end}}{{if .NotComposable}}<span class="g">carries options the composer cannot express &middot; read-only</span>{{else}}<span class="g">read-only example &middot; fork to edit</span>{{end}}</div>{{end}}
</div>
</div>
{{end}}</div>
<div class="globals">
<span class="gt">Globals</span>
<div class="gfield"><label>currency</label><input class="field" name="currency" value="{{.Globals.Currency}}" placeholder="EUR"></div>
<div class="gfield"><label>rebalance</label><input class="field" name="rebalance" value="{{.Globals.Rebalance}}" placeholder="90"></div>
<div class="gfield"><label>sim</label><select class="field" name="sim"><option value="on"{{if eq .Globals.Sim "on"}} selected{{end}}>on</option><option value="off"{{if eq .Globals.Sim "off"}} selected{{end}}>off</option></select></div>
<div class="gfield"><label>benchmark</label><input class="field" name="bench" value="{{.Globals.Bench}}"></div>
<div class="gfield"><label>start</label><input class="field" name="start" value="{{.Globals.Start}}" style="width:120px"></div>
<div class="gfield"><label>end</label><input class="field" name="end" value="{{.Globals.End}}" style="width:110px"></div>
</div>
<div class="cmp-foot">
<div class="budget"><span>link</span><span class="meter"><i></i></span><span class="cmp-bytes">&ndash; / {{.MaxBytes}} B</span></div>
<span class="hint cmp-hint"></span>
<span class="grow"></span>
<button class="btn btn-run" type="button">Run comparison</button>
</div>
</div>
</details>
<script defer src="/composer.js"></script>
`))

// specToP serializes a spec into the /view p= value grammar
// ("ID:WEIGHT,..." plus "!name:" and "!key:value" directives), the exact
// inverse of adhocSpec for what the grammar can express.
//
// The spec's typed fields (Sim, RebalanceDays, ...) are all mirrored verbatim
// in Spec.Meta by portfolio.Parse, so emitting from Meta alone reproduces
// every recognized directive. Content that cannot ride the grammar is left
// behind: holdings that do not resolve locally (an anonymous visitor could
// never render them), the metas that shape the whole comparison rather than
// one portfolio's simulation (optimize, currencies), and any name or meta
// whose text would break the "!" segment grammar (a "!" separator or a
// control character). dropped lists everything left behind, holdings first,
// then the dropped page-level keys sorted.
func specToP(spec *portfolio.Spec) (p string, dropped []string) {
	var pairs []string
	for _, h := range spec.Holdings {
		if !marketdata.KnownLocal(h.ID) {
			dropped = append(dropped, h.ID)
			continue
		}
		pairs = append(pairs, fmt.Sprintf("%s:%g", h.ID, h.RawWeight))
	}

	var b strings.Builder
	b.WriteString(strings.Join(pairs, ","))

	var droppedMeta []string
	// The name rides the same "!"-delimited grammar as the metas, so a name
	// carrying a "!" or a control character cannot be expressed: drop it.
	if spec.Name != "" {
		if breaksSegment(spec.Name) {
			droppedMeta = append(droppedMeta, "name")
		} else {
			fmt.Fprintf(&b, "!name:%s", spec.Name)
		}
	}

	keys := make([]string, 0, len(spec.Meta))
	for k := range spec.Meta {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := spec.Meta[k]
		if k == "optimize" || k == "currencies" || breaksSegment(k) || breaksSegment(v) {
			droppedMeta = append(droppedMeta, k)
			continue
		}
		fmt.Fprintf(&b, "!%s:%s", k, v)
	}
	sort.Strings(droppedMeta)
	return b.String(), append(dropped, droppedMeta...)
}

// breaksSegment reports whether s cannot ride a "!"-delimited p= segment: it
// carries the "!" separator itself, or a control character (a URL-decoded
// newline would inject an extra holding line in adhocSpec's rebuilt text).
func breaksSegment(s string) bool {
	return strings.ContainsRune(s, '!') ||
		strings.IndexFunc(s, func(r rune) bool { return r < 0x20 || r == 0x7f }) >= 0
}
