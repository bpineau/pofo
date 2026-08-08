package firebook

import (
	"bytes"
	"encoding/csv"
	"math"
	"strconv"
	"testing"

	"github.com/bpineau/pofo/pkg/datasets"
)

// fraRealReturns reads the French annual real total returns out of the bundled
// Jorda-Schularick-Taylor panel: year -> column -> return, with the column
// absent when the panel has no record for that year.
func fraRealReturns(t *testing.T) map[int]map[string]float64 {
	t.Helper()
	frozenAgainstData(t)
	r := csv.NewReader(bytes.NewReader(datasets.BroadSample()))
	r.Comment = '#'
	recs, err := r.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) == 0 || recs[0][0] != "iso" {
		t.Fatalf("unexpected broad-sample header %v", recs[0])
	}
	cols := map[int]string{2: "equity", 3: "bond", 4: "bill"}
	out := map[int]map[string]float64{}
	for _, rec := range recs[1:] {
		if rec[0] != "FRA" {
			continue
		}
		year, err := strconv.Atoi(rec[1])
		if err != nil {
			t.Fatal(err)
		}
		row := map[string]float64{}
		for i, key := range cols {
			if rec[i] == "" {
				continue
			}
			v, err := strconv.ParseFloat(rec[i], 64)
			if err != nil {
				t.Fatalf("FRA %d %s: %v", year, key, err)
			}
			row[key] = v
		}
		out[year] = row
	}
	if len(out) == 0 {
		t.Fatal("no FRA rows in the broad-sample panel")
	}
	return out
}

// The victims plate freezes twelve multiples so the book's figures stay pure
// functions. This recomputes each one from pkg/datasets, with the plate's own
// convention (yearly compounding of the real returns, both endpoints included,
// missing years skipped), and fails the moment the two disagree.
func TestVictimesRegimesMatchTheJSTPanel(t *testing.T) {
	fra := fraRealReturns(t)
	for _, reg := range victimesRegimes {
		for _, want := range reg.Rows {
			mult, years := 1.0, 0
			for year := reg.From; year <= reg.To; year++ {
				row, ok := fra[year]
				if !ok {
					t.Fatalf("%s: the panel has no FRA row for %d", reg.Name, year)
				}
				r, ok := row[want.Key]
				if !ok {
					continue
				}
				mult *= 1 + r
				years++
			}
			if years != want.Years {
				t.Errorf("%s / %s: the panel compounds %d years, the plate says %d",
					reg.Name, want.Label, years, want.Years)
			}
			if math.Abs(mult-want.Mult) > 5e-5*want.Mult {
				t.Errorf("%s / %s: the panel leaves x %.6f, the plate says x %.6f",
					reg.Name, want.Label, mult, want.Mult)
			}
		}
	}
}

// The plate's takeaway line, and the article passage it illustrates, claim that
// equities finish ahead of nominal bonds in every one of the four regimes, and
// that only the disinflation preserved purchasing power across the board. Both
// are properties of the frozen numbers, so assert them rather than trust prose.
func TestVictimesRegimesHierarchyHolds(t *testing.T) {
	for _, reg := range victimesRegimes {
		eq, bond := reg.Rows[0], reg.Rows[1]
		if eq.Key != "equity" || bond.Key != "bond" {
			t.Fatalf("%s: rows are ordered %s/%s, the plate assumes equity/bond first",
				reg.Name, eq.Key, bond.Key)
		}
		if eq.Mult <= bond.Mult {
			t.Errorf("%s: equities x %.4f do not beat bonds x %.4f, the plate says they always do",
				reg.Name, eq.Mult, bond.Mult)
		}
	}
	last := victimesRegimes[len(victimesRegimes)-1]
	if last.From != 1985 {
		t.Fatalf("the last regime starts in %d, the plate calls it the disinflation", last.From)
	}
	for _, r := range last.Rows {
		if r.Mult <= 1 {
			t.Errorf("disinflation / %s: x %.4f, the plate says all three bricks gained", r.Label, r.Mult)
		}
	}
	// In the three inflation regimes the article claims nominal bonds finish
	// last of the three bricks, and that at least one brick always lost.
	for _, reg := range victimesRegimes[:len(victimesRegimes)-1] {
		lost := false
		for _, r := range reg.Rows {
			if r.Mult < 1 {
				lost = true
			}
			if r.Key != "bond" && r.Mult <= reg.Rows[1].Mult {
				t.Errorf("%s: %s x %.4f does not beat bonds x %.4f, the article says bonds finish last",
					reg.Name, r.Label, r.Mult, reg.Rows[1].Mult)
			}
		}
		if !lost {
			t.Errorf("%s: no brick lost purchasing power, unexpected for an inflation regime", reg.Name)
		}
	}
}

// The French number formatter the plate labels with.
func TestFrSignedNum(t *testing.T) {
	cases := []struct {
		v    float64
		dec  int
		want string
	}{
		{1306.63, 0, "+1 307 %"},
		{906.12, 0, "+906 %"},
		{-97.54, 1, "−97,5 %"},
		{-9.23, 1, "−9,2 %"},
		{6.59, 1, "+6,6 %"},
		{0, 1, "+0,0 %"},
		{-1234567, 0, "−1 234 567 %"},
	}
	for _, c := range cases {
		if got := frSignedNum(c.v, c.dec) + " %"; got != c.want {
			t.Errorf("frSignedNum(%v, %d) = %q, want %q", c.v, c.dec, got, c.want)
		}
	}
}
