package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// captureOutput runs fn with both standard streams redirected to a pipe and
// returns what they wrote. The terminal modes print their whole answer with
// fmt.Print, so this is the only seam a test has on them. The standard logger
// keeps its own copy of the real stderr and is unaffected, by design: a test
// reads the mode's output, not the fetch narration.
func captureOutput(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	outSave, errSave := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = w, w
	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()
	fn()
	os.Stdout, os.Stderr = outSave, errSave
	_ = w.Close()
	out := <-done
	_ = r.Close()
	return out
}

// runArgs runs one command line with the output captured, so a failing
// expectation can show what the mode printed.
func runArgs(t *testing.T, argv ...string) (string, error) {
	t.Helper()
	var err error
	out := captureOutput(t, func() { err = run(context.Background(), argv) })
	return out, err
}

func TestFrameworkFor(t *testing.T) {
	for _, tc := range []struct{ name, want string }{
		{"", "regimes"}, {"regimes", "regimes"}, {"factors", "factors"},
	} {
		fw, err := frameworkFor(tc.name)
		if err != nil {
			t.Fatalf("-framework %q: %v", tc.name, err)
		}
		if !strings.Contains(strings.ToLower(fw.Name), tc.want) || len(fw.Categories) == 0 {
			t.Errorf("-framework %q resolved to %q with %d categories", tc.name, fw.Name, len(fw.Categories))
		}
	}
	if _, err := frameworkFor("quadrants"); err == nil {
		t.Error("an unknown -framework must fail the run, not fall back")
	}
}

// The explicit flag wins over $COLUMNS, both are bounded, and a value too
// narrow to draw anything falls back to the default.
func TestTermWidth(t *testing.T) {
	t.Setenv("COLUMNS", "120")
	for _, tc := range []struct {
		flag, want int
	}{
		{0, 120},   // $COLUMNS
		{80, 80},   // the flag wins
		{10, 120},  // too narrow: ignored, $COLUMNS again
		{900, 500}, // capped
	} {
		if got := termWidth(tc.flag); got != tc.want {
			t.Errorf("termWidth(%d) = %d, want %d", tc.flag, got, tc.want)
		}
	}
	t.Setenv("COLUMNS", "12")
	if got := termWidth(0); got != 100 {
		t.Errorf("a nonsense $COLUMNS gave %d, want the 100-column default", got)
	}
	t.Setenv("COLUMNS", "4000")
	if got := termWidth(0); got != 160 {
		t.Errorf("$COLUMNS is capped at 160, got %d", got)
	}
	t.Setenv("COLUMNS", "")
	if got := termWidth(0); got != 100 {
		t.Errorf("no width anywhere gave %d, want 100", got)
	}
}

func TestDefaultDataDir(t *testing.T) {
	dir := defaultDataDir()
	if dir == "" {
		t.Fatal("the cache directory must never be empty")
	}
	if filepath.Base(dir) != "pofo" && dir != "data" {
		t.Errorf("cache directory %q is neither <cache>/pofo nor the local fallback", dir)
	}
}

// isTerminal is what turns the ANSI colours on: under `go test` the streams
// are pipes or files, never a character device.
func TestIsTerminalOffAPipe(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = r.Close(); _ = w.Close() }()
	if isTerminal(w) {
		t.Error("a pipe reported itself as a terminal")
	}
}

// The command line's own guardrails, each rejected before anything is
// fetched: an empty run, a malformed window, an unknown classification, an
// exclusive mode combined with -serve, a key that is not one.
func TestRunRejectsBadCommandLines(t *testing.T) {
	file := writeSpec(t, "p.txt", "100 VTI\n")
	for _, tc := range []struct {
		name string
		argv []string
		want string
	}{
		{"nothing to do", nil, "no portfolio file"},
		{"bad -start", []string{"-start", "yesterday", file}, "invalid -start"},
		{"bad -end", []string{"-end", "nope", file}, "invalid -end"},
		{"-end before -start", []string{"-start", "2020-01-01", "-end", "2019-01-01", file}, "-end must be after -start"},
		{"unknown framework", []string{"-framework", "quadrants", file}, "unknown -framework"},
		{"-serve with -fire", []string{"-serve", "-fire"}, "-serve cannot be combined with -fire"},
		{"-serve with -cli", []string{"-serve", "-cli"}, "-serve cannot be combined with -cli"},
		{"bad IndexNow key", []string{"-indexnow-key", "short", file}, "invalid -indexnow-key"},
		{"unknown flag", []string{"-nosuchflag"}, "flag provided but not defined"},
		{"-assets with no identifier", []string{"-assets", " , "}, "no identifier"},
		{"unknown recipe", []string{"-gen-simdata", "NOSUCHRECIPE"}, `no recipe for "NOSUCHRECIPE"`},
		{"nonsense sweep step", []string{"-sweep", "-sweep-step", "0", file}, "-sweep-step must be"},
		{"sweep step too coarse", []string{"-sweep", "-sweep-step", "60", file}, "-sweep-step must be"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out, err := runArgs(t, tc.argv...)
			if err == nil {
				t.Fatalf("%v was accepted; output:\n%s", tc.argv, out)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("%v failed with %q, want something about %q", tc.argv, err, tc.want)
			}
		})
	}
}

// -h prints the usage and stops: asking for help is not an error.
func TestRunHelp(t *testing.T) {
	out, err := runArgs(t, "-h")
	if err != nil {
		t.Fatalf("-h returned %v", err)
	}
	if !strings.Contains(out, "Usage: pofo") || !strings.Contains(out, "-rebalance") {
		t.Errorf("-h printed neither the usage nor the flag list:\n%s", out)
	}
}

// The two offline modes that need no quote at all: the rate catalog and the
// book's translation ledger. Both are dispatched before any client is built.
func TestRunOfflineModes(t *testing.T) {
	out, err := runArgs(t, "-rates", "list")
	if err != nil {
		t.Fatalf("-rates list: %v", err)
	}
	if !strings.Contains(out, "^ESTR") || !strings.Contains(out, "^TNX") {
		t.Errorf("the rate catalog lists neither the policy family nor the yields:\n%s", out)
	}

	out, err = runArgs(t, "-book-drift")
	if err != nil {
		t.Fatalf("-book-drift: %v", err)
	}
	if strings.TrimSpace(out) == "" {
		t.Error("-book-drift printed nothing at all")
	}
}

// -coverage is the offline advisor: a portfolio in, its regime coverage and
// the catalog assets that would fill the gaps out, without a single quote.
func TestRunCoverageOffline(t *testing.T) {
	file := writeSpec(t, "p.txt", "100 NTSG\n")
	out, err := runArgs(t, "-data", t.TempDir(), "-coverage", file)
	if err != nil {
		t.Fatalf("-coverage: %v", err)
	}
	if !strings.Contains(out, "Coverage advisor for") || !strings.Contains(out, "Coverage (by weight)") {
		t.Errorf("-coverage printed no coverage report:\n%s", out)
	}
}
