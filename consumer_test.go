package pofo

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// TestLibraryConsumer runs the documented workflow from a separate module.
// It catches accidental dependence on repository-only imports, test helpers,
// disk caches or the network. Output values may move with bundled data; the
// test checks the public contract and round trip, not a duplicated golden.
func TestLibraryConsumer(t *testing.T) {
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	source, err := os.ReadFile("examples/code/workflow.go")
	if err != nil {
		t.Fatal(err)
	}
	source = []byte(strings.TrimPrefix(string(source), "//go:build ignore\n"))
	mod := "module example.org/pofo-consumer\n\ngo 1.26\n\nrequire github.com/bpineau/pofo v0.0.0\nreplace github.com/bpineau/pofo => " + strconv.Quote(root) + "\n"
	for name, data := range map[string][]byte{"go.mod": []byte(mod), "main.go": source} {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	cmd := exec.Command("go", "run", ".", "-out", "study.csv")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off", "GOPROXY=off", "GOSUMDB=off", "GOTOOLCHAIN=local")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("consumer: %v\n%s", err, output)
	}
	if !strings.Contains(string(output), "held-out volatility") {
		t.Fatalf("incomplete workflow: %s", output)
	}
	f, err := os.Open(filepath.Join(dir, "study.csv"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	series, err := marketdata.ReadSeriesCSV(f, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(series) != 3 {
		t.Fatalf("%d exported curves, want 3", len(series))
	}
	for _, s := range series {
		if s.Currency != "EUR" {
			t.Fatalf("%s lost currency", s.Symbol)
		}
		if _, err := s.Stats(); err != nil {
			t.Fatalf("%s: %v", s.Symbol, err)
		}
	}
}
