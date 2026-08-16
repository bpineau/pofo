// The -book-drift mode: report what the FIRE book's translations owe their
// French source.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/bpineau/pofo/pkg/firebook"
)

// runBookDrift prints the book's synchronization report, one line per French
// article that is not settled: first the translations whose French original
// moved since they were made ("stale"), then the articles no translation
// covers yet ("untranslated"), each naming the English slug it is owed under,
// then the articles the French edition keeps for itself ("fr-only"). That
// output IS the translation worklist. It always succeeds: drift is a state of
// the book, not an error.
func runBookDrift() error {
	items := firebook.Drift()
	if len(items) == 0 {
		fmt.Println("book: every translation is current")
		return nil
	}
	counts := map[string]int{}
	for _, it := range items {
		counts[it.Reason]++
		line := fmt.Sprintf("%-13s %-38s", it.Reason, it.FRSlug)
		if it.ENSlug != "" {
			line += " -> " + it.ENSlug
		}
		fmt.Println(strings.TrimRight(line, " "))
	}
	fmt.Fprintf(os.Stderr, "\n%d stale, %d untranslated, %d fr-only\n",
		counts["stale"], counts["untranslated"], counts["fr-only"])
	return nil
}
