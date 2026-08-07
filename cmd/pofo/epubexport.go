// The -export-epub mode: write the embedded FIRE book out as a standalone
// EPUB 3 file for offline reading.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/bpineau/pofo/pkg/firebook"
)

// runExportEpub builds "Le FIRE tranquille" as an EPUB 3 file and writes it to
// path. The Modified timestamp is the current time (the file is generated on
// demand, not served from a running process), so two exports differ only by
// that stamp. It prints the written path and its size.
func runExportEpub(path string) error {
	data, err := firebook.EPUB(time.Now())
	if err != nil {
		return fmt.Errorf("building EPUB: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	fmt.Fprintf(os.Stderr, "wrote %s (%d bytes)\n", path, len(data))
	return nil
}
