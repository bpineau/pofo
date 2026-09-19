package compare

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/bpineau/pofo/pkg/chart"
	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/portfolio"
	"github.com/bpineau/pofo/pkg/suggest"
)

// breakdownSlices must be deterministic even when several categories carry
// the exact same weight: agg is a map, so ties are broken by label
// (ascending) rather than by map-iteration order, keeping the rendered
// report byte-stable run to run.
func TestBreakdownSlicesTiesDeterministic(t *testing.T) {
	agg := map[string]float64{"Delta": 25, "Alpha": 25, "Charlie": 25, "Bravo": 25}

	first := breakdownSlices(agg, 10)
	for i := range 50 {
		if got := breakdownSlices(agg, 10); !reflect.DeepEqual(got, first) {
			t.Fatalf("run %d differs from the first: %+v vs %+v", i, got, first)
		}
	}

	// All weights tie, so the order is purely the label tiebreak (ascending).
	want := []string{"Alpha", "Bravo", "Charlie", "Delta"}
	got := make([]string, len(first))
	for i, s := range first {
		got[i] = s.Label
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("tie order = %v, want %v (label ascending)", got, want)
	}
}

// Distinct weights are ordered by weight descending, unchanged by the
// tiebreak, so existing reports keep their exact slice order.
func TestBreakdownSlicesWeightOrder(t *testing.T) {
	agg := map[string]float64{"low": 10, "high": 60, "mid": 30}
	got := breakdownSlices(agg, 10)
	want := []string{"high", "mid", "low"}
	labels := make([]string, len(got))
	for i, s := range got {
		labels[i] = s.Label
	}
	if !reflect.DeepEqual(labels, want) {
		t.Errorf("weight order = %v, want %v", labels, want)
	}
}

// A holding wears ONE color inside its section: the coverage bars and the
// realized-contribution charts are read against each other ("who was supposed
// to cover this regime" vs "who actually delivered"), so the segment colors
// must be the ones the timeline and the regime matrix give the same holdings.
// The two blocks used to disagree from three holdings on, the bars taking the
// palette's first n slots and the charts the n slots PaletteFor keeps apart,
// which silently repainted holding 2 of a five-line portfolio.
func TestCoverageBarColorsMatchTheContributionCharts(t *testing.T) {
	meta, err := suggest.LoadMeta(bytes.NewReader(datasets.AssetMeta()))
	if err != nil {
		t.Fatal(err)
	}
	// Five catalog lines in five different asset classes, so every holding
	// contributes to at least one regime bar and none is dropped.
	ids := []string{"IWDA", "IGLN", "DBMF", "IB01", "TLT"}
	assets := make([]portfolio.Asset, len(ids))
	for i, id := range ids {
		assets[i] = portfolio.Asset{ID: id, Symbol: id, Weight: 1 / float64(len(ids))}
	}
	want := chart.PaletteFor(len(assets)) // the assignment contributionCharts uses
	bars := coverageBars(assets, meta, suggest.RegimeFramework())
	if len(bars) == 0 {
		t.Fatal("no coverage bars: the fixture lost its catalog metadata")
	}
	seen := map[string]string{}
	for _, b := range bars {
		for _, seg := range b.Segments {
			id, _, _ := strings.Cut(seg.Tip, " ")
			if prev, ok := seen[id]; ok && prev != seg.Color {
				t.Errorf("%s colored %s here and %s there", id, prev, seg.Color)
			}
			seen[id] = seg.Color
		}
	}
	if len(seen) == 0 {
		t.Fatal("no coverage segments to color")
	}
	for i, id := range ids {
		if c, ok := seen[id]; ok && c != want[i] {
			t.Errorf("%s (holding %d) colored %s in the coverage bars, %s in the contribution charts", id, i, c, want[i])
		}
	}
}
