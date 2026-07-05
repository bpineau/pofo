package decumul

import (
	"fmt"
	"math"
	"math/rand/v2"

	"github.com/bpineau/pofo/pkg/scenario"
)

// GeoMean estimates a return model's effective real geometric mean per period by
// drawing nPaths and averaging log(1+r) across every drawn return. It is the
// single most useful sanity number for a decumulation model: the arithmetic mean
// and the one-period volatility can look reasonable while the geometric mean, the
// rate wealth actually compounds at, is quietly negative (arithmetic-vs-geometric
// confusion, volatility drag, an over-pessimistic prior). A plan built on an
// implausible geometric mean is impossible by construction, not by bad luck.
func GeoMean(src scenario.Source, nPaths int, rng *rand.Rand) float64 {
	if nPaths <= 0 || src.Len() == 0 {
		return 0
	}
	var sumLog float64
	var n int
	for range nPaths {
		for _, r := range src.Draw(rng) {
			// Floor 1+r: an extreme tail draw clamped to a total loss (1+r<=0)
			// would send log to -inf and poison the estimate. Such a draw is a
			// measure-zero model artifact, not the typical compounding rate the
			// guard cares about.
			sumLog += math.Log(math.Max(1+r, 1e-9))
			n++
		}
	}
	if n == 0 {
		return 0
	}
	return math.Exp(sumLog/float64(n)) - 1
}

// PlausibleGeoMean is the band a real-equity-ish decumulation model's effective
// geometric mean should sit in. Below it the plan is doomed by construction (the
// recurring calibration failure); above it the model is implausibly rosy versus
// a century of developed-market evidence (DMS world real ~5% geometric).
const (
	PlausibleGeoLo = 0.00
	PlausibleGeoHi = 0.08
)

// Plausibility returns human-readable warnings when a model's headline numbers
// leave their anchor bands: the effective real geometric mean outside
// [PlausibleGeoLo, PlausibleGeoHi], or a 30-year safe withdrawal rate outside
// [1.5%, 5%] (Anarkulova's broad-sample ~2.3% to the US-backtest ~4%+). Pass
// safeWR30 <= 0 to skip the withdrawal-rate check. An empty slice means the
// calibration looks sane. It exists to catch regressions before they reach a
// user as a spuriously hopeless (or hopeless-to-believe) headline.
func Plausibility(geoMean, safeWR30 float64) []string {
	var w []string
	if geoMean < PlausibleGeoLo || geoMean > PlausibleGeoHi {
		w = append(w, fmt.Sprintf("effective real geometric mean %.1f%%/yr is outside the plausible %.0f-%.0f%% band; check arithmetic-vs-geometric and volatility drag",
			geoMean*100, PlausibleGeoLo*100, PlausibleGeoHi*100))
	}
	if safeWR30 > 0 && (safeWR30 < 0.015 || safeWR30 > 0.05) {
		w = append(w, fmt.Sprintf("30-year safe withdrawal rate %.1f%% is outside the literature's 1.5-5%% range",
			safeWR30*100))
	}
	return w
}
