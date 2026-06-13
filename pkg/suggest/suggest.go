package suggest

import (
	"math"
	"sort"
)

// Holding is one position of the portfolio under analysis.
type Holding struct {
	ID      string
	Weight  float64 // fraction of the portfolio
	Meta    Meta    // catalog metadata; zero when unknown
	HasMeta bool
}

// Candidate is a catalog asset that could be added. PortReturns and Returns
// are the held portfolio's and the candidate's daily returns over their
// common (overlap) window, aligned to the same calendar and equal length.
type Candidate struct {
	Meta        Meta
	PortReturns []float64
	Returns     []float64
	Years       float64 // length of the overlap window, for display/filtering
	Simulated   bool    // the candidate's history includes simulated data
}

// Options tunes the analysis. The zero value is unusable; start from
// DefaultOptions.
type Options struct {
	Weights       []float64 // candidate weights to try (fractions)
	Windows       int       // walk-forward windows
	GapThreshold  float64   // regime coverage at/below this (fraction) is a gap
	RedundancyMin float64   // correlation at/above this flags redundancy
	MinWindowFrac float64   // require a Sharpe gain in at least this fraction of windows
	MaxSuggest    int       // most suggestions to return
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{
		Weights:       []float64{0.05, 0.10, 0.15, 0.20},
		Windows:       8,
		GapThreshold:  0.10,
		RedundancyMin: 0.95,
		MinWindowFrac: 0.5,
		MaxSuggest:    3,
	}
}

// Suggestion is one recommended asset to add.
type Suggestion struct {
	Meta          Meta
	Fills         Category // the gap category it primarily fills
	Weight        float64  // suggested weight (fraction)
	Corr          float64  // correlation to the held portfolio
	VolBefore     float64  // portfolio daily-return volatility before
	VolAfter      float64  // ... and after adding the candidate
	SharpeWins    int      // walk-forward windows where Sharpe improved
	DDWins        int      // ... where max-drawdown improved
	Windows       int      // total walk-forward windows evaluated
	MedSharpeGain float64  // median out-of-sample Sharpe gain across windows
	Years         float64
	Simulated     bool
}

// Result is the full analysis.
type Result struct {
	Framework    string
	Coverage     map[Category]float64
	Unclassified float64
	Gaps         []Category
	Redundancies []Group
	Suggestions  []Suggestion
}

// Analyze computes the framework coverage, redundancies and ranked
// suggestions. heldReturns[i] is holdings[i]'s daily-return series (aligned,
// equal length) over the held window; candidates carry their own
// overlap-aligned returns.
func Analyze(holdings []Holding, heldReturns [][]float64, candidates []Candidate, opts Options, fw Framework) Result {
	cov, uncl := Coverage(holdings, fw)
	gaps := Gaps(cov, fw, opts.GapThreshold)
	res := Result{
		Framework:    fw.Name,
		Coverage:     cov,
		Unclassified: uncl,
		Gaps:         gaps,
		Redundancies: Redundancies(holdings, heldReturns, opts.RedundancyMin),
		Suggestions:  RankCandidates(gaps, cov, candidates, opts, fw),
	}
	return res
}

// RankCandidates keeps the candidates that fill a gap category and whose
// benefit is robust out-of-sample, ranked by the gap they fill (most
// under-covered first) then by median out-of-sample Sharpe gain.
func RankCandidates(gaps []Category, cov map[Category]float64, candidates []Candidate, opts Options, fw Framework) []Suggestion {
	gapSet := map[Category]bool{}
	for _, g := range gaps {
		gapSet[g] = true
	}
	var out []Suggestion
	for _, c := range candidates {
		fills := primaryGap(c.Meta, gapSet, cov, fw)
		if fills == "" {
			continue // helps no gap category
		}
		bestW, bestGain := 0.0, math.Inf(-1)
		var sWins, ddWins, total int
		for _, w := range opts.Weights {
			sw, dw, tot, gain := walkForward(c.PortReturns, c.Returns, w, opts.Windows)
			if tot == 0 || float64(sw)/float64(tot) < opts.MinWindowFrac {
				continue
			}
			if gain > bestGain {
				bestGain, bestW = gain, w
				sWins, ddWins, total = sw, dw, tot
			}
		}
		if total == 0 {
			continue // no robust weight
		}
		mixR := mix(c.PortReturns, c.Returns, bestW)
		out = append(out, Suggestion{
			Meta:          c.Meta,
			Fills:         fills,
			Weight:        bestW,
			Corr:          Correlation(c.PortReturns, c.Returns),
			VolBefore:     std(c.PortReturns),
			VolAfter:      std(mixR),
			SharpeWins:    sWins,
			DDWins:        ddWins,
			Windows:       total,
			MedSharpeGain: bestGain,
			Years:         c.Years,
			Simulated:     c.Simulated,
		})
	}
	sort.SliceStable(out, func(a, b int) bool {
		ca, cb := cov[out[a].Fills], cov[out[b].Fills]
		if ca != cb {
			return ca < cb // fill the most under-covered gap first
		}
		return out[a].MedSharpeGain > out[b].MedSharpeGain
	})
	// Keep the suggestions diverse: at most one per asset class (two golds
	// add nothing the second time — the whole point is diversification).
	seen := map[string]bool{}
	kept := out[:0]
	for _, s := range out {
		if seen[s.Meta.AssetClass] {
			continue
		}
		seen[s.Meta.AssetClass] = true
		kept = append(kept, s)
	}
	out = kept
	if len(out) > opts.MaxSuggest {
		out = out[:opts.MaxSuggest]
	}
	return out
}

// primaryGap returns the gap category with the lowest coverage that the
// asset helps in, or "" when it helps none.
func primaryGap(m Meta, gapSet map[Category]bool, cov map[Category]float64, fw Framework) Category {
	best := Category("")
	bestCov := math.Inf(1)
	for _, c := range fw.Classify(m) {
		if gapSet[c] && cov[c] < bestCov {
			best, bestCov = c, cov[c]
		}
	}
	return best
}

// walkForward splits the overlap into Windows contiguous slices and, in each,
// compares the augmented portfolio (existing rescaled to 1-w, candidate at w)
// to the baseline. It returns how many windows improved Sharpe and
// max-drawdown, the number evaluated, and the median Sharpe gain. Because
// adding an asset at a fixed weight fits nothing to the data, this measures
// the consistency of the benefit across periods, not an in-sample optimum.
func walkForward(portR, candR []float64, w float64, k int) (sharpeWins, ddWins, total int, medSharpeGain float64) {
	n := len(portR)
	if n == 0 || len(candR) != n || k < 1 {
		return 0, 0, 0, 0
	}
	size := n / k
	if size < 20 {
		size = n // too few points to split: one window
	}
	var gains []float64
	for start := 0; start < n; start += size {
		end := start + size
		if end > n {
			end = n
		}
		if end-start < 20 {
			break
		}
		base := portR[start:end]
		aug := mix(base, candR[start:end], w)
		bs, as := windowSharpe(base), windowSharpe(aug)
		bd, ad := windowMaxDD(base), windowMaxDD(aug)
		total++
		if as > bs {
			sharpeWins++
		}
		if ad >= bd { // less negative drawdown is better
			ddWins++
		}
		gains = append(gains, as-bs)
	}
	return sharpeWins, ddWins, total, median(gains)
}

// mix returns the augmented daily returns: (1-w)*port + w*cand.
func mix(portR, candR []float64, w float64) []float64 {
	out := make([]float64, len(portR))
	for i := range portR {
		out[i] = (1-w)*portR[i] + w*candR[i]
	}
	return out
}

// windowSharpe is the annualized Sharpe (risk-free 0) of a return slice.
func windowSharpe(r []float64) float64 {
	s := std(r)
	if s == 0 {
		return 0
	}
	return mean(r) / s * math.Sqrt(252)
}

// windowMaxDD is the deepest peak-to-trough loss implied by a return slice
// (a negative fraction, 0 when never underwater).
func windowMaxDD(r []float64) float64 {
	v, peak, mdd := 1.0, 1.0, 0.0
	for _, x := range r {
		v *= 1 + x
		if v > peak {
			peak = v
		}
		if dd := v/peak - 1; dd < mdd {
			mdd = dd
		}
	}
	return mdd
}

func median(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	s := append([]float64(nil), xs...)
	sort.Float64s(s)
	n := len(s)
	if n%2 == 1 {
		return s[n/2]
	}
	return (s[n/2-1] + s[n/2]) / 2
}
