package portfolio_test

import (
	"fmt"
	"strings"

	"portfodor/pkg/portfolio"
)

// Parse reads a portfolio description: "<weight %> <identifier>
// [free text]", everything after a # being a comment.
func ExampleParse() {
	spec, err := portfolio.Parse("my-portfolio", strings.NewReader(`
# Comment lines and blank lines are ignored.
60   VTI    US stocks       # free text accepted
25,5 IE00B4L5Y983           # decimal comma accepted
14.5 GLD
`))
	if err != nil {
		panic(err)
	}
	for _, h := range spec.Holdings {
		fmt.Printf("%5.1f %% %s\n", h.Weight*100, h.ID)
	}
	// Output:
	//  60.0 % VTI
	//  25.5 % IE00B4L5Y983
	//  14.5 % GLD
}
