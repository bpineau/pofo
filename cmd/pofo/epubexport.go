// The -export-epub mode: write one edition of the embedded FIRE book out as a
// standalone EPUB 3 file for offline reading. -book-lang picks the edition:
// "fr" (the default) writes "Le FIRE tranquille", "en" writes "The Quiet
// FIRE"; the file name is always the one the caller asked for.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/bpineau/pofo/pkg/firebook"
)

// runExportEpub builds the lang edition of the book as an EPUB 3 file and
// writes it to path. The Modified timestamp is the current time (the file is
// generated on demand, not served from a running process), so two exports
// differ only by that stamp. It prints the written path and its size.
func runExportEpub(path, lang string) error {
	ed, err := bookEdition(lang)
	if err != nil {
		return err
	}
	data, err := ed.EPUB(time.Now())
	if err != nil {
		return fmt.Errorf("building EPUB: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	fmt.Fprintf(os.Stderr, "wrote %s (%d bytes)\n", path, len(data))
	return nil
}

// bookEdition resolves a -book-lang value to an edition of the FIRE book.
func bookEdition(lang string) (*firebook.Edition, error) {
	switch lang {
	case "", "fr":
		return firebook.French, nil
	case "en":
		return firebook.English, nil
	}
	return nil, fmt.Errorf("unknown -book-lang %q: use fr or en", lang)
}
