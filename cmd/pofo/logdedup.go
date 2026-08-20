// Log hygiene for the long-lived servers (-serve, -fire).
package main

import (
	"io"
	"log"
	"strings"
	"sync"
)

// The fetch path narrates what it does: identifier resolutions, history
// extensions, cache decisions. Interactively (one portfolio, one report) that
// narration is the point, so the CLI modes keep it verbatim. A server replays
// the same pipeline on every request, and the same handful of sentences then
// scroll past forever, burying the lines that carry news.
//
// The narration reaches the logs through two doors: marketdata.Client.Logf,
// which cmd/pofo wires to log.Printf, and a few direct log.Printf calls in
// pkg/compare (the "resolved X -> name" line among them). Filtering at the
// Client.Logf seam would therefore catch only half of it, so the filter sits
// where both doors lead: the standard logger's destination. Nothing about the
// lines changes, and no package learns that a server is running.

// dedupLimit is how many distinct informational lines a dedupWriter remembers.
// Anonymous visitors can mint unbounded identifiers (the /view grammar accepts
// any catalog id), so the memory is bounded: on overflow the writer forgets
// everything and starts over, which at worst prints a line a second time.
const dedupLimit = 4096

// dedupWriter forwards log lines to w, printing each distinct informational
// line only once. A line whose text starts with "warning:" is always printed,
// however often it repeats: warnings report degraded data (a held-flat FX rate,
// a stale cache, a missing benchmark) and their repetition is itself the signal.
//
// The unit of deduplication is one Write, which the log package guarantees is
// one logged line. It also assumes the line carries no timestamp, which holds
// here: main sets log.SetFlags(0). Safe for concurrent use.
type dedupWriter struct {
	w    io.Writer
	mu   sync.Mutex
	seen map[string]bool
}

// newDedupWriter returns a dedupWriter in front of w.
func newDedupWriter(w io.Writer) *dedupWriter {
	return &dedupWriter{w: w, seen: make(map[string]bool)}
}

// Write prints p unless it is an informational line already printed. A
// suppressed line reports success: the caller wrote it, the filter chose not
// to repeat it.
func (d *dedupWriter) Write(p []byte) (int, error) {
	line := strings.TrimRight(string(p), "\n")
	if line == "" || strings.HasPrefix(line, "warning:") {
		return d.w.Write(p)
	}
	d.mu.Lock()
	if d.seen[line] {
		d.mu.Unlock()
		return len(p), nil
	}
	if len(d.seen) >= dedupLimit {
		d.seen = make(map[string]bool)
	}
	d.seen[line] = true
	d.mu.Unlock()
	return d.w.Write(p)
}

// dedupServerLog installs the filter in front of the standard logger's current
// destination. The server entrypoints call it once at startup; the CLI modes
// never do, so an interactive run keeps every line it always printed.
func dedupServerLog() {
	log.SetOutput(newDedupWriter(log.Writer()))
}
