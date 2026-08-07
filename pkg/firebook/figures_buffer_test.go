package firebook

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"testing"

	"github.com/bpineau/pofo/pkg/replay"
)

// underwaterEpisode is one crossing recomputed from the engine: the month the
// real all-time high was set, the month it was regained, and the worst real
// drawdown in between.
type underwaterEpisode struct {
	from, to string
	months   int
	depth    float64 // percent, negative
	open     bool    // still under water at the end of the record
}

// underwaterEpisodes walks the bundled real 60/40 and lists its crossings,
// exactly as the comment on bufferEpisodes describes the recipe.
func underwaterEpisodes(t *testing.T) []underwaterEpisode {
	t.Helper()
	s, err := replay.Reference()
	if err != nil {
		t.Fatal(err)
	}
	month := func(i int) string { return s.Dates[i].Format("2006-01") }

	var out []underwaterEpisode
	peak, peakAt, trough := s.Index[0], 0, s.Index[0]
	under := false
	for i := 1; i < len(s.Index); i++ {
		v := s.Index[i]
		if v >= peak {
			if under {
				out = append(out, underwaterEpisode{month(peakAt), month(i), i - peakAt, (trough/peak - 1) * 100, false})
				under = false
			}
			peak, peakAt, trough = v, i, v
			continue
		}
		if !under {
			under, trough = true, v
		}
		trough = math.Min(trough, v)
	}
	if under {
		last := len(s.Index) - 1
		out = append(out, underwaterEpisode{month(peakAt), month(last), last - peakAt, (trough/peak - 1) * 100, true})
	}
	return out
}

// The cash-buffer plate freezes the crossings of the bundled real 60/40 so the
// book's figures stay pure functions. This recomputes them and fails the moment
// the figure and the engine disagree.
func TestTraverseesMatelasMatchesTheEngine(t *testing.T) {
	var got []underwaterEpisode
	for _, e := range underwaterEpisodes(t) {
		if e.depth <= bufferDepthThreshold {
			got = append(got, e)
		}
	}
	sort.SliceStable(got, func(i, j int) bool { return got[i].months > got[j].months })

	if len(got) != len(bufferEpisodes) {
		t.Fatalf("engine finds %d crossings deeper than %.0f %%, the plate draws %d",
			len(got), -bufferDepthThreshold, len(bufferEpisodes))
	}
	for i, want := range bufferEpisodes {
		g := got[i]
		if g.open {
			t.Errorf("%s is still under water at the end of the record: mark the bar as open instead of closing it", g.from)
		}
		if g.from != want.from || g.to != want.to || g.months != want.months {
			t.Errorf("crossing %d: engine %s to %s (%d months), plate %s to %s (%d months)",
				i, g.from, g.to, g.months, want.from, want.to, want.months)
		}
		if math.Abs(g.depth-want.depth) > 0.1 {
			t.Errorf("crossing %s: engine bottoms at %.1f %%, plate says %.1f %%", g.from, g.depth, want.depth)
		}
	}
}

// The plate and the article agree on one arithmetic: how much of the underwater
// time a buffer of a given size actually pays for. Check it on the frozen list,
// which the test above ties to the engine.
func TestBufferReachArithmetic(t *testing.T) {
	total, longest := 0, 0
	for _, e := range bufferEpisodes {
		total += e.months
		if e.months > longest {
			longest = e.months
		}
	}
	if total != 393 || longest != 124 {
		t.Fatalf("%d underwater months, longest %d: expected 393 and 124", total, longest)
	}

	reach := func(buffer int) (covered, full int) {
		for _, e := range bufferEpisodes {
			if e.months <= buffer {
				covered += e.months
				full++
				continue
			}
			covered += buffer
		}
		return covered, full
	}
	for _, tc := range []struct {
		buffer, covered, full int
	}{
		{18, 160, 2}, // the article's lower bound covers 41 % of the months
		{24, 201, 4}, // two years: half the months under water, to the month
		{36, 257, 5}, // the upper bound: two thirds, and five crossings out of nine
	} {
		covered, full := reach(tc.buffer)
		if covered != tc.covered || full != tc.full {
			t.Errorf("a %d-month buffer covers %d months and %d whole crossings, expected %d and %d",
				tc.buffer, covered, full, tc.covered, tc.full)
		}
	}
	// "un matelas de deux ans couvre la moitié des mois passés sous l'eau"
	covered, _ := reach(24)
	if share := float64(covered) / float64(total); math.Abs(share-0.5) > 0.02 {
		t.Errorf("a two-year buffer covers %.0f %% of the underwater months, the article says half", share*100)
	}
	// The median crossing the article now quotes, and its worst case.
	months := make([]int, len(bufferEpisodes))
	for i, e := range bufferEpisodes {
		months[i] = e.months
	}
	sort.Ints(months)
	if median := months[len(months)/2]; median != 32 {
		t.Errorf("median crossing is %d months, the article says 32", median)
	}
	if longest != 124 {
		t.Errorf("longest crossing is %d months, the article says ten years", longest)
	}
}

// The closing note is computed from the frozen list, so it must land on the
// numbers the article's prose leans on.
func TestTraverseesMatelasPlateReads(t *testing.T) {
	svg := figTraverseesMatelas()
	for _, want := range []string{
		"5 épisodes sur 9 tiennent dans 36 mois.",
		"Les 4 autres laissent 136 mois à découvert,",
		"dont 88 pour la seule traversée des",
		">124<", ">16<", // the longest and the shortest bar, in months
		"stagflation des années 1970",
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("the plate does not carry %q", want)
		}
	}
	// Nothing may run past the viewBox.
	if !strings.Contains(svg, fmt.Sprintf(`viewBox="0 0 %d %d"`, 640, 404)) {
		t.Error("the plate's viewBox moved: recheck the rendered PNG before freezing it")
	}
}
