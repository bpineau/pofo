package firebook

import (
	"math"
	"strings"
	"testing"

	"github.com/bpineau/pofo/pkg/replay"
)

// The replay plates carry frozen numbers so the book's figures stay pure
// functions. This test is what keeps them honest: it recomputes every series
// from pkg/replay and the bundled record, and fails the moment a figure and the
// engine disagree. A refreshed dataset therefore breaks the build here rather
// than quietly leaving a wrong chart in the book.
func TestReplayFiguresMatchTheEngine(t *testing.T) {
	for _, tc := range []struct {
		start  int
		index  []float64
		income [][]float64
	}{
		{1973, replayIndex1973, replayIncome1973},
		{1985, replayIndex1985, replayIncome1985},
		{2000, replayIndex2000, replayIncome2000},
	} {
		res, err := replay.Run(replay.Setup{
			Start: tc.start, Capital: 600000, Spend: 24000, Years: 40,
			Mu: 0.045, Sigma: 0.10, Df: 5, TargetRuin: 0.05, RaiseCap: 1.5,
		})
		if err != nil {
			t.Fatalf("%d: %v", tc.start, err)
		}

		if len(tc.index) != res.Years+1 {
			t.Errorf("%d: plate has %d index points, engine covers %d years", tc.start, len(tc.index), res.Years)
			continue
		}
		for k, want := range tc.index {
			if got := res.Index[12*k]; math.Abs(got-want) > 0.05 {
				t.Errorf("%d: index year %d = %.1f, plate says %.1f", tc.start, tc.start-1+k, got, want)
			}
		}

		if len(tc.income) != len(res.Rules) {
			t.Fatalf("%d: plate has %d rules, engine has %d", tc.start, len(tc.income), len(res.Rules))
		}
		for i, rule := range res.Rules {
			if rule.NameFR != replayRuleNames[i] {
				t.Errorf("panel %d is labelled %q, the engine calls it %q", i, replayRuleNames[i], rule.NameFR)
			}
			if len(tc.income[i]) != res.Years {
				t.Errorf("%d/%s: plate has %d years, engine has %d", tc.start, rule.Name, len(tc.income[i]), res.Years)
				continue
			}
			for k, want := range tc.income[i] {
				if got := rule.Spend[k] / 1000; math.Abs(got-want) > 0.05 {
					t.Errorf("%d/%s year %d: engine %.1f, plate %.1f", tc.start, rule.Name, tc.start+k, got, want)
				}
			}
		}
	}
}

// The article's summary tables are transposed on purpose (rules as rows), so
// they fit a book page and an e-reader without horizontal scrolling. This test
// reads the numbers back out of the markdown and checks them against the
// engine, the same guard the figures get.
func TestReplayTablesMatchTheEngine(t *testing.T) {
	raw, err := assets.ReadFile("assets/book/fr/sept-facons-de-vivre.md")
	if err != nil {
		t.Fatal(err)
	}
	md := string(raw)
	for _, start := range []int{1973, 1985, 2000} {
		res, err := replay.Run(replay.Setup{
			Start: start, Capital: 600000, Spend: 24000, Years: 40,
			Mu: 0.045, Sigma: 0.10, Df: 5, TargetRuin: 0.05, RaiseCap: 1.5,
		})
		if err != nil {
			t.Fatal(err)
		}
		for _, rule := range res.Rules {
			row := tableRow(md, start, rule.NameFR)
			if row == nil {
				t.Errorf("%d: no table row for %q", start, rule.NameFR)
				continue
			}
			for _, chk := range []struct {
				label string
				col   int
				want  float64
				tol   float64
			}{
				{"moyenne", 1, rule.Mean / 1000, 0.05},
				{"écart", 2, rule.CV * 100, 0.5},
				{"pire", 3, rule.Low / 1000, 0.05},
				{"années sous le plan", 4, float64(rule.LeanYears), 0.01},
				{"capital final", 5, rule.Final / 1000, 1},
			} {
				if math.Abs(row[chk.col]-chk.want) > chk.tol {
					t.Errorf("%d/%s %s: table %.1f, engine %.1f", start, rule.NameFR, chk.label, row[chk.col], chk.want)
				}
			}
		}
	}
}

// tableRow finds the row for one rule inside the section of one start year and
// returns its numeric cells (index 0 is the name column, left at 0). Each era
// opens on a "## Janvier <year>" heading, which is what bounds the search: the
// same rule name appears in all three tables.
func tableRow(md string, start int, nameFR string) []float64 {
	head := "## Janvier " + itoa(start)
	i := strings.Index(md, head)
	if i < 0 {
		return nil
	}
	section := md[i+len(head):]
	if j := strings.Index(section, "\n## "); j >= 0 {
		section = section[:j]
	}
	for _, line := range strings.Split(section, "\n") {
		if !strings.HasPrefix(line, "| "+nameFR+" ") {
			continue
		}
		cells := strings.Split(strings.Trim(line, "| "), "|")
		out := make([]float64, len(cells))
		for i := 1; i < len(cells); i++ {
			out[i] = parseFRNumber(cells[i])
		}
		return out
	}
	return nil
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	var b []byte
	for v > 0 {
		b = append([]byte{byte('0' + v%10)}, b...)
		v /= 10
	}
	return string(b)
}

// parseFRNumber reads a French-formatted cell ("46,6", "1 656", "26 %").
func parseFRNumber(s string) float64 {
	s = strings.TrimSpace(s)
	s = strings.NewReplacer(",", ".", " ", "", " ", "", "%", "", "**", "").Replace(s)
	var v float64
	neg := strings.HasPrefix(s, "-") || strings.HasPrefix(s, "−")
	s = strings.TrimLeft(s, "-−")
	intPart, frac, hasFrac := strings.Cut(s, ".")
	for _, c := range intPart {
		if c < '0' || c > '9' {
			return math.NaN()
		}
		v = v*10 + float64(c-'0')
	}
	if hasFrac {
		scale := 0.1
		for _, c := range frac {
			if c < '0' || c > '9' {
				return math.NaN()
			}
			v += float64(c-'0') * scale
			scale /= 10
		}
	}
	if neg {
		v = -v
	}
	return v
}
