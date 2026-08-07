// The /view live composer's server side: translating specs into the p=
// grammar for the Fork feature. The mount markup and embedded assets
// arrive with the composer front end.
package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/portfolio"
)

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
