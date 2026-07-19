package examples_test

import (
	"fmt"

	"github.com/bpineau/pofo/examples"
)

func ExampleList() {
	for _, in := range examples.List() {
		if in.Name == "claude-dragonlite" {
			fmt.Println(in.Title)
		}
	}
	// Output: Claude dragon-lite
}
