// The report's composition blocks: look-through pies (geography, currency,
// equity sectors, asset type), segmented coverage bars and the duration /
// currency notes, all from pkg/suggest's composition splits.
package main

import (
	"fmt"
	"html/template"
	"math"
	"sort"
	"strings"

	"github.com/bpineau/pofo/pkg/chart"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/portfolio"
	"github.com/bpineau/pofo/pkg/report"
	"github.com/bpineau/pofo/pkg/suggest"
)

// neutralSliceColor fills the catch-all "Other" wedge of the composition
// pies; specialSliceColor fills the informative non-category wedges ("No
// country", "None (real assets)", …), a darker neutral so the two read as
// different kinds of remainder. Both stay visually distinct from the
// palette-colored slices.
const (
	neutralSliceColor = "#C6CEDA"
	specialSliceColor = "#9AA2B1"
)

// holdingsFor adapts a portfolio's assets to suggest holdings, resolving each
// identifier to its catalog metadata (aliases and SIM suffix tolerated) and
// keeping the base identifier for display.
func holdingsFor(assets []portfolio.Asset, meta map[string]suggest.Meta) []suggest.Holding {
	holdings := make([]suggest.Holding, len(assets))
	for i, a := range assets {
		base, _ := marketdata.SplitSim(a.ID)
		m, _, ok := metaFor(meta, a.ID)
		holdings[i] = suggest.Holding{ID: base, Weight: a.Weight, Meta: m, HasMeta: ok}
	}
	return holdings
}

// breakdownPies builds the look-through composition pies (geography, currency
// exposure, equity sectors, asset type) for a portfolio's detail section from
// the suggest composition splits. Returns the non-empty pie SVGs (nil when no
// metadata is available at all).
func breakdownPies(assets []portfolio.Asset, meta map[string]suggest.Meta) []template.HTML {
	if len(meta) == 0 {
		return nil
	}
	holdings := holdingsFor(assets, meta)

	geo := suggest.GeographySplit(holdings)
	foldInto(geo, "Other", suggest.BucketUnknown)

	cur := suggest.CurrencySplit(holdings)
	foldInto(cur, "Other", suggest.CurrencyOther, suggest.BucketUnknown)
	relabel(cur, suggest.CurrencyNone, "None (real assets)")
	relabel(cur, suggest.CurrencyDynamic, "Dynamic (futures)")

	sec, equity := suggest.EquitySectorSplit(holdings)
	secTitle := fmt.Sprintf("Equity sectors (%.0f%% of capital)", equity*100)

	cls := map[string]float64{}
	for class, w := range suggest.AssetClassSplit(holdings) {
		cls[prettyClass(class)] += w
	}

	svgs := []string{
		chart.Pie(chart.PieOptions{Title: "Geography"},
			breakdownSlices(geo, 8, suggest.BucketNoCountry)),
		chart.Pie(chart.PieOptions{Title: "Currency exposure"},
			breakdownSlices(cur, 8, "None (real assets)", "Dynamic (futures)")),
		chart.Pie(chart.PieOptions{Title: secTitle},
			breakdownSlices(sec, 9, suggest.BucketUnknown)),
		chart.Pie(chart.PieOptions{Title: "Asset type (look-through)"},
			breakdownSlices(cls, 8, prettyClass(suggest.BucketUnknown))),
	}
	var pies []template.HTML
	for _, s := range svgs {
		if s != "" {
			pies = append(pies, template.HTML(s))
		}
	}
	return pies
}

// foldInto merges the listed keys of a split into the dst key.
func foldInto(agg map[string]float64, dst string, keys ...string) {
	for _, k := range keys {
		if v, ok := agg[k]; ok && k != dst {
			agg[dst] += v
			delete(agg, k)
		}
	}
}

// relabel renames a split key, merging with any existing value.
func relabel(agg map[string]float64, from, to string) {
	if v, ok := agg[from]; ok {
		agg[to] += v
		delete(agg, from)
	}
}

// breakdownSlices turns an aggregation map into pie slices: largest first,
// wedges below 3 % and the literal "Other" key merged into a trailing neutral
// "Other" slice, capped at maxSlices colored entries. The special labels
// (informative non-categories like "No country") are pinned after it in the
// given order, in a darker neutral. A pie carrying no colored slice at all
// (no real composition) returns nil so it is omitted.
func breakdownSlices(agg map[string]float64, maxSlices int, special ...string) []chart.Slice {
	specialSet := map[string]bool{}
	for _, s := range special {
		specialSet[s] = true
	}
	type kv struct {
		k string
		v float64
	}
	items := make([]kv, 0, len(agg))
	total, other := 0.0, 0.0
	for k, v := range agg {
		total += v
		if k == "Other" {
			other += v
			continue
		}
		if !specialSet[k] {
			items = append(items, kv{k, v})
		}
	}
	if total <= 0 {
		return nil
	}
	sort.Slice(items, func(i, j int) bool { return items[i].v > items[j].v })
	slices := make([]chart.Slice, 0, maxSlices)
	for _, it := range items {
		if it.v/total < 0.03 || len(slices) >= maxSlices-1 {
			other += it.v
			continue
		}
		slices = append(slices, chart.Slice{Label: it.k, Value: it.v})
	}
	if len(slices) == 0 {
		return nil // only remainders: nothing to show
	}
	if other > 0 {
		slices = append(slices, chart.Slice{Label: "Other", Value: other, Color: neutralSliceColor})
	}
	for _, s := range special {
		if v := agg[s]; v > 0 {
			slices = append(slices, chart.Slice{Label: s, Value: v, Color: specialSliceColor})
		}
	}
	return slices
}

// prettyClass turns a catalog asset_class slug ("aggregate-bond") into a
// display label ("Aggregate bond").
func prettyClass(s string) string {
	s = strings.ReplaceAll(s, "-", " ")
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// coverageBars computes a portfolio's macro-regime coverage for the report,
// each bar split into per-holding segments (stable color per holding across
// the rows) with a compact contributor line beneath. It returns nil when no
// asset carries metadata (nothing meaningful to show).
func coverageBars(assets []portfolio.Asset, meta map[string]suggest.Meta, fw suggest.Framework) []report.CoverageBar {
	holdings := holdingsFor(assets, meta)
	anyMeta := false
	for _, h := range holdings {
		anyMeta = anyMeta || h.HasMeta
	}
	if !anyMeta {
		return nil
	}
	cov, _ := suggest.Coverage(holdings, fw)
	contrib := suggest.Contributors(holdings, fw)
	gapSet := map[suggest.Category]bool{}
	for _, g := range suggest.Gaps(cov, fw, suggest.DefaultOptions().GapThreshold) {
		gapSet[g] = true
	}
	bars := make([]report.CoverageBar, 0, len(fw.Categories))
	for _, rg := range fw.Categories {
		// The track represents max(coverage, 100 %): segments stay
		// proportional even when notional coverage exceeds the portfolio.
		scale := math.Max(cov[rg], 1)
		var segs []report.CoverageSeg
		var parts []string
		for _, c := range contrib[rg] {
			segs = append(segs, report.CoverageSeg{
				Width: math.Round(c.Weight/scale*1000) / 10,
				Color: chart.PaletteColor(c.Index),
				Tip:   fmt.Sprintf("%s %.0f%%", c.ID, c.Weight*100),
			})
			parts = append(parts, fmt.Sprintf("%s %.0f", c.ID, c.Weight*100))
		}
		bars = append(bars, report.CoverageBar{
			Regime:   string(rg),
			Pct:      int(cov[rg]*100 + 0.5),
			Gap:      gapSet[rg],
			Segments: segs,
			Detail:   strings.Join(parts, " · "),
		})
	}
	return bars
}

// compositionNotes renders the look-through duration and currency summary
// lines shown under a portfolio's composition (empty without metadata).
func compositionNotes(assets []portfolio.Asset, meta map[string]suggest.Meta, base string) []string {
	if len(meta) == 0 {
		return nil
	}
	holdings := holdingsFor(assets, meta)
	var notes []string

	led := suggest.DurationSplit(holdings)
	switch {
	case led.Nominal > 0:
		line := fmt.Sprintf("Rate duration (look-through): %.1f y nominal per unit of capital (≈ %.0f pts of 7y-bond equivalent)",
			led.Nominal, led.Nominal/7*100)
		if led.Real > 0 {
			line += fmt.Sprintf(", plus %.1f y real-rate from inflation-linked bonds", led.Real)
		}
		if led.Missing > 0.02 {
			line += fmt.Sprintf("; no duration figure for %.0f%% of the bond notional", led.Missing*100)
		}
		notes = append(notes, line+".")
	case led.Real > 0:
		notes = append(notes, fmt.Sprintf("Rate duration (look-through): %.1f y real-rate from inflation-linked bonds.", led.Real))
	}

	p := suggest.CurrencyProfile(suggest.CurrencySplit(holdings), base)
	if p.Base+p.Foreign+p.NonFiat > 0 {
		line := fmt.Sprintf("Currency (look-through): %.0f%% %s-native or hedged · %.0f%% unhedged foreign", p.Base*100, base, p.Foreign*100)
		if p.Top != "" {
			line += fmt.Sprintf(" (mostly %s, %.0f%%)", p.Top, p.TopShare*100)
		}
		if p.NonFiat > 0 {
			line += fmt.Sprintf(" · %.0f%% non-fiat or futures-driven", p.NonFiat*100)
		}
		notes = append(notes, line+".")
	}
	return notes
}
