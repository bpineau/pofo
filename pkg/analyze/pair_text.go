package analyze

import (
	"fmt"
	"io"
	"math"
	"text/tabwriter"
	"time"
)

// WriteText renders the study as plain text, aligned columns and no markup,
// for a terminal or a program to read: the two series, the level and
// identity figures, the per-calendar table, the dated divergences, the
// calendar years side by side and the warnings. Returns are in percent with
// two decimals, yearly gaps in points. It returns the first write error.
func (st *PairStudy) WriteText(w io.Writer) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	p := func(format string, args ...any) { fmt.Fprintf(tw, format, args...) }
	// last ends a row with an optional cell, leaving no trailing blanks.
	last := func(cell string) string {
		if cell == "" {
			return "\n"
		}
		return "\t" + cell + "\n"
	}

	for _, s := range []struct {
		tag  string
		side PairSide
	}{{"A", st.A}, {"B", st.B}} {
		p("%s\t%s\t%s to %s\t%s%s", s.tag, s.side.ID, ymd(s.side.First), ymd(s.side.Last),
			cadenceLabel(s.side.PeriodsPerYear), last(s.side.Currency))
	}
	p("\n")
	p("window\t%s to %s (%.1f years)\n", ymd(st.Start), ymd(st.End), st.SpanYears)
	p("dates\t%d shared, %d only in A, %d only in B\n", st.Shared, st.OnlyA, st.OnlyB)
	p("identity\t%s\n", st.identityLine())
	p("CAGR\tA %s, B %s, gap %s/yr%s\n", pct(st.A.CAGR), pct(st.B.CAGR), points(st.CAGRGap), gapSE(st.GapSE))
	p("level\tA ends %s from B, both rebased at the start\n", pct(st.LevelGap))
	p("volatility\tA %s, B %s, each at its own cadence\n", size(st.A.Volatility), size(st.B.Volatility))
	p("drawdown\tA %s, B %s\n", pct(st.A.MaxDrawdown), pct(st.B.MaxDrawdown))

	blocks := []struct {
		name string
		r    *PairReturns
	}{{"daily", st.Daily}, {"monthly", st.Monthly}}
	if st.Daily != nil || st.Monthly != nil {
		p("\nreturns\tperiods\tper year\tcorr\tvol A\tvol B\tvol ratio\ttracking\tbeta\talpha/yr\n")
		for _, b := range blocks {
			if r := b.r; r != nil {
				p("%s\t%d\t%.0f\t%.3f\t%s\t%s\t%.3f\t%s\t%.3f\t%s\n", b.name, r.Periods, r.PeriodsPerYear, r.Corr,
					size(r.VolA), size(r.VolB), r.VolRatio, size(r.TrackingError), r.Beta, pct(r.Alpha))
			}
		}
	}
	for _, b := range blocks {
		if b.r == nil || len(b.r.Divergences) == 0 {
			continue
		}
		p("\nlargest %s divergences\tA\tB\tA-B\texcess\n", b.name)
		for _, d := range b.r.Divergences {
			p("%s to %s\t%s\t%s\t%s\t%s\n", ymd(d.Start), ymd(d.End), pct(d.A), pct(d.B), points(d.Gap), size(d.Excess))
		}
	}
	if len(st.Years) > 0 {
		p("\nyear\tA\tB\tA-B\n")
		for _, y := range st.Years {
			note := ""
			if y.Partial {
				note = "partial"
			}
			p("%d\t%s\t%s\t%s%s", y.Year, pct(y.A), pct(y.B), points(y.Diff), last(note))
		}
	}
	if len(st.Warnings) > 0 {
		p("\nwarnings\n")
		for _, w := range st.Warnings {
			p("  %s\n", w)
		}
	}
	return tw.Flush()
}

// identityLine says whether the two agree date by date.
func (st *PairStudy) identityLine() string {
	switch {
	case st.Shared < 2:
		return fmt.Sprintf("A/B %.6g at the start; too few shared dates to follow the ratio", st.LevelRatio)
	case !st.FirstDivergence.IsZero():
		return fmt.Sprintf("A/B %.6g at the start, first moves on %s", st.LevelRatio, ymd(st.FirstDivergence))
	case math.Abs(st.LevelRatio-1) > DivergenceTolerance:
		return fmt.Sprintf("A/B %.6g at the start and on every shared date: one series up to scale", st.LevelRatio)
	case st.OnlyA == 0 && st.OnlyB == 0:
		return "identical on every date"
	}
	return "equal on every shared date"
}

// ymd formats a date as YYYY-MM-DD.
func ymd(t time.Time) string { return t.Format(time.DateOnly) }

// pct formats a fraction as a signed percent.
func pct(x float64) string { return fmt.Sprintf("%+.2f %%", x*100) }

// size formats a non-negative fraction (a volatility, a distance) as a
// percent.
func size(x float64) string { return fmt.Sprintf("%.2f %%", x*100) }

// points formats a return difference in percentage points.
func points(x float64) string { return fmt.Sprintf("%+.2f pt", x*100) }

// gapSE formats the standard error beside a gap, when there is one.
func gapSE(se float64) string {
	if se == 0 {
		return ""
	}
	return fmt.Sprintf(" (standard error %.2f)", se*100)
}

// cadenceLabel is a measured cadence with its count.
func cadenceLabel(ppy float64) string {
	return fmt.Sprintf("%s (%.0f/yr)", cadenceName(ppy), ppy)
}
