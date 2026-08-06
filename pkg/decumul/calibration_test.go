package decumul

import (
	"testing"

	"github.com/bpineau/pofo/pkg/scenario"
)

// Calibration anchors. These pin the engine to the published safe-withdrawal
// literature so the headline figures are trustworthy, not free-floating. The
// model is i.i.d. (no cross-year mean reversion), so it reads a touch stricter
// than the mean-reverting historical record, especially at long horizons; the
// tolerances below bracket the anchors rather than reproducing them exactly
// (an exact match awaits the deferred broad-sample historical panel).

const calPaths = 20000

// Trinity / Bengen anchor: a US-historical-ish i.i.d. real-return model
// (~5% real geometric: arithmetic 5.5%, sigma 10%), a fixed 4% real rule over
// 30 years with no tax, should succeed about 95% of the time (ruin ~5%).
func TestCalibrationTrinityAnchor(t *testing.T) {
	p := Plan{
		Capital: 1_000_000, NeedAnnual: 40000, Years: 30, Tax: CTOFlatTax{Rate: 0},
		Source: scenario.ParametricSource{Mu: 0.055, Sigma: 0.10, Df: 8, Periods: 30},
	}
	ruin := p.Simulate(calPaths, 8, 7).RuinProb()
	if ruin < 0.02 || ruin > 0.10 {
		t.Errorf("Trinity anchor: fixed 4%%/30y ruin = %.1f%%, want ~5%% (2-10%%)", ruin*100)
	}
}

// Broad-sample (Anarkulova-class) anchor: the conservative prior (a
// mean-preserving regime at ~3.5% real geometric, fatter tails, clustered
// drawdowns), a fixed 4% real rule over 30 years, must fail far more often than
// the Trinity anchor, and its 5%-ruin safe withdrawal must sit well below 4%,
// near the broad-sample ~2.26% (we are a little stricter under i.i.d.).
func TestCalibrationBroadSampleConservativeAnchor(t *testing.T) {
	p := Plan{
		Capital: 1_000_000, NeedAnnual: 40000, Years: 30, Tax: CTOFlatTax{Rate: 0},
		Source: scenario.NewMarkovRegime(0.045, 0.13, 4, 30),
	}
	ruin := p.Simulate(calPaths, 8, 7).RuinProb()
	if ruin < 0.20 {
		t.Errorf("broad-sample anchor: fixed 4%%/30y ruin = %.1f%%, want materially higher (>=20%%)", ruin*100)
	}

	safe := p.Solve(0.05, WithdrawalAxis(0, 100000), calPaths, 8, 7)
	wr := safe / 1_000_000
	if wr < 0.012 || wr > 0.030 {
		t.Errorf("broad-sample 5%%-ruin SWR = %.2f%%, want near ~2.26%% (1.2-3.0%%)", wr*100)
	}
}

// The optimistic and conservative anchors must bracket: at the same fixed
// 4%/30y rule, the broad-sample model fails far more than the Trinity model.
// This is the calibrated spread that the multi-model view exposes.
func TestCalibrationAnchorsBracket(t *testing.T) {
	trinity := Plan{
		Capital: 1_000_000, NeedAnnual: 40000, Years: 30, Tax: CTOFlatTax{Rate: 0},
		Source: scenario.ParametricSource{Mu: 0.055, Sigma: 0.10, Df: 8, Periods: 30},
	}.Simulate(calPaths, 8, 7).RuinProb()
	broad := Plan{
		Capital: 1_000_000, NeedAnnual: 40000, Years: 30, Tax: CTOFlatTax{Rate: 0},
		Source: scenario.NewMarkovRegime(0.045, 0.13, 4, 30),
	}.Simulate(calPaths, 8, 7).RuinProb()

	if broad < trinity+0.15 {
		t.Errorf("anchors do not bracket: Trinity %.1f%% vs broad %.1f%% (want broad >= Trinity+15pp)", trinity*100, broad*100)
	}
}
