package fire_test

import (
	"context"
	"fmt"

	"github.com/bpineau/pofo/pkg/fire"
)

func ExampleBuildPlan() {
	spec := fire.Defaults()
	spec.Capital, spec.AnnualSpending, spec.Years = 1000000, 30000, 30
	plan, err := fire.BuildPlan(spec)
	if err != nil {
		panic(err)
	}
	model, err := fire.PrepareModel(fire.Central, fire.MarketSpec{Mu: .04, Sigma: .12, Nu: 5}, spec.Years)
	if err != nil {
		panic(err)
	}
	plan.Source = model.Source
	ensemble, err := plan.SimulateContext(context.Background(), 100, 2, 7)
	if err != nil {
		panic(err)
	}
	fmt.Println(model.Basis, len(ensemble.Paths), ensemble.Years)
	// Output: parametric_assumptions 100 30
}
