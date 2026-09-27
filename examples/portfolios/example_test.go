package portfolios_test

import (
	"fmt"

	"github.com/bpineau/pofo/examples/portfolios"
)

func ExampleList() {
	for _, in := range portfolios.List() {
		if in.Name == "golden-butterfly" {
			fmt.Println(in.Title)
		}
	}
	// Output: Golden Butterfly
}
