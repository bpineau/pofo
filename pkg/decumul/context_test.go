package decumul

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/bpineau/pofo/pkg/scenario"
)

func TestContextSimulationParity(t *testing.T) {
	for _, monthly := range []bool{false, true} {
		periods := 20
		if monthly {
			periods *= 12
		}
		p := Plan{Capital: 1e6, NeedAnnual: 40000, Years: 20, Monthly: monthly, Source: scenario.ParametricSource{Mu: 0.04, Sigma: 0.15, Df: 5, Periods: periods}}
		want := p.Simulate(80, 1, 7)
		got, err := p.SimulateContext(context.Background(), 80, 3, 7)
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatalf("monthly=%v context parity: %v", monthly, err)
		}
		d, err := p.DrawContext(context.Background(), 80, 3, 7)
		if err != nil || !reflect.DeepEqual(d, p.Draw(80, 1, 7)) {
			t.Fatalf("draw parity: %v", err)
		}
		ruin, err := p.RuinProbOnContext(context.Background(), d, 3)
		if err != nil || ruin != want.RuinProb() {
			t.Fatalf("ruin parity: %v %v", ruin, err)
		}
	}
}

func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	p := Plan{Capital: 100, NeedAnnual: 1, Years: 1, Source: scenario.ParametricSource{Periods: 1}}
	if _, err := p.DrawContext(ctx, 100, 2, 7); !errors.Is(err, context.Canceled) {
		t.Fatalf("draw: %v", err)
	}
	if _, err := p.SimulateContext(ctx, 100, 2, 7); !errors.Is(err, context.Canceled) {
		t.Fatalf("simulate: %v", err)
	}
	d := p.Draw(100, 1, 7)
	if _, err := p.SimulateOnContext(ctx, d, 2); !errors.Is(err, context.Canceled) {
		t.Fatalf("simulate on: %v", err)
	}
	if _, err := p.RuinProbOnContext(ctx, d, 2); !errors.Is(err, context.Canceled) {
		t.Fatalf("ruin: %v", err)
	}
}

type cancelTax struct {
	cancel context.CancelFunc
	calls  int
}

func (t *cancelTax) GrossUp(net, growth, cost float64) (float64, float64, float64) {
	t.calls++
	t.cancel()
	return (CTOFlatTax{Rate: .3}).GrossUp(net, growth, cost)
}

func TestCancellationDuringPath(t *testing.T) {
	for _, monthly := range []bool{false, true} {
		ctx, cancel := context.WithCancel(context.Background())
		tax := &cancelTax{cancel: cancel}
		p := Plan{Capital: 1e9, NeedAnnual: 10000, Years: 100, Monthly: monthly, Tax: tax}
		d := Draws{Returns: []scenario.Sequence{make(scenario.Sequence, 1200), make(scenario.Sequence, 1200)}}
		got, err := p.SimulateOnContext(ctx, d, 1)
		if !errors.Is(err, context.Canceled) || len(got.Paths) != 0 || tax.calls != 1 {
			t.Fatalf("monthly=%v canceled path returned %d paths, %d tax calls, %v", monthly, len(got.Paths), tax.calls, err)
		}
		cancel()
	}
}
