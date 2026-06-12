package portfolio_test

import (
	"fmt"
	"strings"

	"portfodor/portfolio"
)

// Parse lit une description de portefeuille: « <poids %> <identifiant>
// [texte libre] », tout ce qui suit un # étant un commentaire.
func ExampleParse() {
	spec, err := portfolio.Parse("mon-portefeuille", strings.NewReader(`
# Lignes de commentaire et lignes vides ignorées.
60   VTI    actions US      # texte libre accepté
25,5 IE00B4L5Y983           # virgule décimale acceptée
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
