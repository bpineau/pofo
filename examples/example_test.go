package examples_test

import (
	"fmt"

	"github.com/bpineau/pofo/examples"
)

func ExampleList() {
	for _, in := range examples.List() {
		if in.Name == "golden-butterfly" {
			fmt.Println(in.Title)
		}
	}
	// Output: Golden Butterfly
}
