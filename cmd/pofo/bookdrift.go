// The -book-drift mode: report what the FIRE book's translations owe their
// French source.
package main

import (
	"fmt"
	"os"

	"github.com/bpineau/pofo/pkg/firebook"
)

// runBookDrift prints the book's synchronization report: first the translated
// articles whose French original moved since they were made ("stale"), then
// the French articles no translation covers yet ("untranslated"). That output
// IS the translation worklist. It always succeeds: drift is a state of the
// book, not an error.
func runBookDrift() error {
	items := firebook.Drift()
	if len(items) == 0 {
		fmt.Println("book: every translation is current")
		return nil
	}
	stale, untranslated := 0, 0
	for _, it := range items {
		switch it.Reason {
		case "stale":
			stale++
			fmt.Printf("stale         %-38s <- %s\n", it.ENSlug, it.FRSlug)
		default:
			untranslated++
			fmt.Printf("untranslated  %s\n", it.FRSlug)
		}
	}
	fmt.Fprintf(os.Stderr, "\n%d stale, %d untranslated\n", stale, untranslated)
	return nil
}
