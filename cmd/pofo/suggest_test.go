package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/portfolio"
	"github.com/bpineau/pofo/pkg/suggest"
)

// catalogMeta loads the embedded catalog metadata, the map every coverage
// path reads.
func catalogMeta(t *testing.T) map[string]suggest.Meta {
	t.Helper()
	meta, err := suggest.LoadMeta(bytes.NewReader(datasets.AssetMeta()))
	if err != nil {
		t.Fatal(err)
	}
	return meta
}

// An identifier reaches its catalog record however it was written: as an
// alias, as the canonical id, and with the SIM suffix the coverage modes must
// look through.
func TestMetaFor(t *testing.T) {
	meta := catalogMeta(t)
	m, canon, ok := metaFor(meta, "NTSG")
	if !ok {
		t.Fatalf("the alias NTSG resolved to nothing (canonical %q)", canon)
	}
	if m.AssetClass == "" {
		t.Errorf("NTSG resolved to a record with no asset class: %+v", m)
	}
	if _, simCanon, simOK := metaFor(meta, "NTSGSIM"); !simOK || simCanon != canon {
		t.Errorf("NTSGSIM resolved to (%q, %v), want the same record as NTSG (%q)", simCanon, simOK, canon)
	}
	if _, _, ok := metaFor(meta, "NOSUCHASSET"); ok {
		t.Error("an identifier absent from the catalog reported metadata")
	}
}

// Confidence orders the representatives a gap is filled with; anything the
// catalog does not say ranks last.
func TestConfRank(t *testing.T) {
	if !(confRank("high") > confRank("medium") && confRank("medium") > confRank("low")) {
		t.Errorf("confidence does not order: high=%d medium=%d low=%d",
			confRank("high"), confRank("medium"), confRank("low"))
	}
	if confRank("") != confRank("anything else") {
		t.Error("an unknown confidence must rank like an absent one")
	}
}

func TestIntersectsGap(t *testing.T) {
	fw := suggest.RegimeFramework()
	if len(fw.Categories) < 2 {
		t.Fatal("the regime framework lost its categories")
	}
	a, b := fw.Categories[0], fw.Categories[1]
	if !intersectsGap([]suggest.Category{b, a}, map[suggest.Category]bool{a: true}) {
		t.Error("a category in the gap set was not seen")
	}
	if intersectsGap([]suggest.Category{b}, map[suggest.Category]bool{a: true}) {
		t.Error("a category outside the gap set was counted")
	}
	if intersectsGap(nil, map[suggest.Category]bool{a: true}) {
		t.Error("an unclassified asset filled a gap")
	}
}

// The advisor answers from the embedded metadata alone: no quote, no network.
// A one-line portfolio necessarily leaves regimes uncovered, so the gap
// section and its per-class catalog picks must both appear.
func TestCoverageAdviceNamesTheGaps(t *testing.T) {
	spec := portfolio.Single("NTSG")
	opt := &options{fw: suggest.RegimeFramework()}
	out := captureOutput(t, func() { coverageAdvice(spec, opt, catalogMeta(t)) })
	if !strings.Contains(out, "Coverage advisor for NTSG") {
		t.Fatalf("no header:\n%s", out)
	}
	if !strings.Contains(out, "To fill the gaps") {
		t.Fatalf("a single holding covered every regime, which cannot be:\n%s", out)
	}
	if !strings.Contains(out, "gap") {
		t.Errorf("no category was marked as a gap:\n%s", out)
	}
}

// The factor framework is the other classification the same code serves, and
// an identifier absent from the catalog is reported as unclassified weight
// rather than silently dropped.
func TestCoverageAdviceFactorsAndUnclassified(t *testing.T) {
	spec, err := portfolio.Parse("mixed", strings.NewReader("50 NTSG\n50 NOSUCHASSET\n"))
	if err != nil {
		t.Fatal(err)
	}
	opt := &options{fw: suggest.FactorFramework()}
	out := captureOutput(t, func() { coverageAdvice(spec, opt, catalogMeta(t)) })
	if !strings.Contains(out, "unclassified") {
		t.Errorf("the unknown half of the portfolio was not reported:\n%s", out)
	}
	if !strings.Contains(out, suggest.FactorFramework().Name) {
		t.Errorf("the report does not name the factor framework:\n%s", out)
	}
}

// -coverage without a portfolio is a user error, not an empty report.
func TestRunCoverageNeedsAPortfolio(t *testing.T) {
	if err := runCoverage(nil, &options{fw: suggest.RegimeFramework()}); err == nil {
		t.Error("-coverage accepted an empty run")
	}
}

// Every portfolio of the run gets its own advice.
func TestRunCoverageEachPortfolio(t *testing.T) {
	specs := []*portfolio.Spec{portfolio.Single("NTSG"), portfolio.Single("IWDA")}
	opt := &options{fw: suggest.RegimeFramework()}
	out := captureOutput(t, func() {
		if err := runCoverage(specs, opt); err != nil {
			t.Errorf("runCoverage: %v", err)
		}
	})
	for _, name := range []string{"NTSG", "IWDA"} {
		if !strings.Contains(out, "Coverage advisor for "+name) {
			t.Errorf("%s got no advice:\n%s", name, out)
		}
	}
}

// -suggest fetches, so only its guardrail is exercised here.
func TestRunSuggestNeedsAPortfolio(t *testing.T) {
	if err := runSuggest(t.Context(), nil, nil, &options{fw: suggest.RegimeFramework()}); err == nil {
		t.Error("-suggest accepted an empty run")
	}
}

// The suggestion report is the -suggest mode's whole output: every number the
// analysis produced has to reach the terminal, redundancies included.
func TestRenderSuggest(t *testing.T) {
	fw := suggest.RegimeFramework()
	gap := fw.Categories[0]
	res := suggest.Result{
		Framework:    fw.Name,
		Coverage:     map[suggest.Category]float64{gap: 0.05, fw.Categories[1]: 0.95},
		Unclassified: 0.1,
		Gaps:         []suggest.Category{gap},
		Redundancies: []suggest.Group{{IDs: []string{"VOO", "CSPX"}, Weight: 0.4, MinCorr: 0.99}},
		Suggestions: []suggest.Suggestion{{
			Meta:  suggest.Meta{ID: "IGLN", AssetClass: "commodity", Notes: "physical gold"},
			Fills: gap, Weight: 0.1, Corr: 0.05, VolBefore: 0.14, VolAfter: 0.13,
			SharpeWins: 7, DDWins: 6, Windows: 8, MedSharpeGain: 0.12, Simulated: true,
		}},
	}
	start, end := day(2000, 1, 3), day(2020, 12, 31)
	out := captureOutput(t, func() { renderSuggest("mine", start, end, res, fw) })
	for _, want := range []string{
		"Suggestions for mine", "2000-01-03", "2020-12-31",
		"Redundancies", "VOO + CSPX", "correlation ≥ 0.99",
		"1. IGLN (commodity)", "suggested weight 10 %", "7/8", "6/8",
		"physical gold", "[history partly simulated]",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("the report does not mention %q:\n%s", want, out)
		}
	}
}

// A portfolio that covers every category is told so, and one whose gaps
// nothing filled out of sample is told that instead: neither prints an empty
// suggestion list.
func TestRenderSuggestNothingToSay(t *testing.T) {
	fw := suggest.RegimeFramework()
	start, end := day(2000, 1, 3), day(2020, 12, 31)

	out := captureOutput(t, func() {
		renderSuggest("covered", start, end, suggest.Result{Framework: fw.Name}, fw)
	})
	if !strings.Contains(out, "is covered, no gap to fill") {
		t.Errorf("a gapless portfolio got no verdict:\n%s", out)
	}

	out = captureOutput(t, func() {
		renderSuggest("gappy", start, end, suggest.Result{
			Framework: fw.Name, Gaps: []suggest.Category{fw.Categories[0]},
		}, fw)
	})
	if !strings.Contains(out, "No gap-filling asset showed a robust out-of-sample benefit") {
		t.Errorf("an unfillable gap got no verdict:\n%s", out)
	}
}
