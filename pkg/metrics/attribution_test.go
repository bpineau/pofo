package metrics_test

import (
	"errors"
	"math"
	"testing"

	"github.com/bpineau/pofo/pkg/metrics"
)

func TestAttributeSharesSumToOne(t *testing.T) {
	contrib := [][]float64{
		{0.01, -0.02, 0.03, 0.00, -0.01, 0.02},
		{0.00, 0.01, -0.01, 0.02, 0.00, -0.01},
		{-0.01, 0.00, 0.01, 0.01, 0.02, 0.00},
	}
	att, err := metrics.Attribute(contrib)
	if err != nil {
		t.Fatalf("Attribute: %v", err)
	}
	var risk, ret float64
	for i := range att.Risk {
		risk += att.Risk[i]
		ret += att.Return[i]
	}
	if math.Abs(risk-1) > 1e-12 {
		t.Errorf("risk shares sum to %v, want 1", risk)
	}
	if math.Abs(ret-1) > 1e-12 {
		t.Errorf("return shares sum to %v, want 1", ret)
	}
}

// A holding that never moves takes no risk and earns nothing, whatever its
// weight: the decomposition must not hand it a share.
func TestAttributeFlatHoldingTakesNoShare(t *testing.T) {
	contrib := [][]float64{
		{0.01, -0.02, 0.03, -0.01},
		{0, 0, 0, 0},
	}
	att, err := metrics.Attribute(contrib)
	if err != nil {
		t.Fatalf("Attribute: %v", err)
	}
	if att.Risk[1] != 0 {
		t.Errorf("flat holding risk share = %v, want 0", att.Risk[1])
	}
	if math.Abs(att.Risk[0]-1) > 1e-12 {
		t.Errorf("moving holding risk share = %v, want 1", att.Risk[0])
	}
}

// The point of the Euler decomposition: a hedge that moves against the book
// takes LESS risk share than its own volatility suggests, and may take a
// negative one. Here the second holding is the exact mirror of the first, so
// it cancels the portfolio's variance and its share must be negative.
func TestAttributeHedgeTakesNegativeRiskShare(t *testing.T) {
	contrib := [][]float64{
		{0.02, -0.03, 0.04, -0.01, 0.02},
		{-0.01, 0.015, -0.02, 0.005, -0.01},
	}
	att, err := metrics.Attribute(contrib)
	if err != nil {
		t.Fatalf("Attribute: %v", err)
	}
	if att.Risk[1] >= 0 {
		t.Errorf("mirror hedge risk share = %v, want negative", att.Risk[1])
	}
	if att.Risk[0] <= 1 {
		t.Errorf("driver risk share = %v, want above 1 (the hedge offsets it)", att.Risk[0])
	}
}

// Shares are scale-free: doubling every contribution leaves them unchanged.
func TestAttributeIsScaleFree(t *testing.T) {
	contrib := [][]float64{
		{0.01, -0.02, 0.03, 0.00, 0.01},
		{0.02, 0.01, -0.01, 0.02, -0.02},
	}
	scaled := make([][]float64, len(contrib))
	for i, c := range contrib {
		scaled[i] = make([]float64, len(c))
		for k, v := range c {
			scaled[i][k] = 2 * v
		}
	}
	a, err := metrics.Attribute(contrib)
	if err != nil {
		t.Fatalf("Attribute: %v", err)
	}
	b, err := metrics.Attribute(scaled)
	if err != nil {
		t.Fatalf("Attribute scaled: %v", err)
	}
	for i := range a.Risk {
		if math.Abs(a.Risk[i]-b.Risk[i]) > 1e-12 {
			t.Errorf("risk share %d: %v vs %v when scaled", i, a.Risk[i], b.Risk[i])
		}
	}
}

// An equally weighted book of independent, identically sized movers splits its
// risk evenly: the textbook case the decomposition must reproduce.
func TestAttributeIndependentEqualHoldingsSplitEvenly(t *testing.T) {
	// Orthogonal, same-amplitude series (a Hadamard-style pattern).
	contrib := [][]float64{
		{0.01, 0.01, -0.01, -0.01},
		{0.01, -0.01, 0.01, -0.01},
	}
	att, err := metrics.Attribute(contrib)
	if err != nil {
		t.Fatalf("Attribute: %v", err)
	}
	for i, r := range att.Risk {
		if math.Abs(r-0.5) > 1e-12 {
			t.Errorf("risk share %d = %v, want 0.5", i, r)
		}
	}
}

func TestAttributeRejectsUnusableInput(t *testing.T) {
	if _, err := metrics.Attribute(nil); !errors.Is(err, metrics.ErrNoContributions) {
		t.Errorf("nil: got %v, want ErrNoContributions", err)
	}
	if _, err := metrics.Attribute([][]float64{{0.01}}); !errors.Is(err, metrics.ErrNoContributions) {
		t.Errorf("single point: got %v, want ErrNoContributions", err)
	}
	// A motionless portfolio has no variance to share out.
	if _, err := metrics.Attribute([][]float64{{0, 0, 0}}); !errors.Is(err, metrics.ErrNoContributions) {
		t.Errorf("flat portfolio: got %v, want ErrNoContributions", err)
	}
	// The trap the report golden caught: contributions that are constant but
	// NOT zero leave floating-point cancellation noise instead of an exact
	// zero variance, and a naive test turns that noise into shares that look
	// authoritative (measured: -75 % and -150 % on a two-asset book).
	constant := make([][]float64, 2)
	for i := range constant {
		constant[i] = make([]float64, 30)
		for k := range constant[i] {
			constant[i][k] = 0.1 * float64(i+1)
		}
	}
	if _, err := metrics.Attribute(constant); !errors.Is(err, metrics.ErrNoContributions) {
		t.Errorf("constant contributions: got %v, want ErrNoContributions", err)
	}
	if _, err := metrics.Attribute([][]float64{{0.01, 0.02}, {0.01}}); err == nil {
		t.Error("unequal lengths: want an error")
	}
}
