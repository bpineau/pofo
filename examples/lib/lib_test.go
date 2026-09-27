package lib_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	// The packages the programs use: a change in any of them invalidates
	// the cached result of this test, which runs the programs as binaries.
	_ "github.com/bpineau/pofo/pkg/analyze"
	_ "github.com/bpineau/pofo/pkg/decumul"
	_ "github.com/bpineau/pofo/pkg/marketdata"
	_ "github.com/bpineau/pofo/pkg/metrics"
	_ "github.com/bpineau/pofo/pkg/portfolio"
	_ "github.com/bpineau/pofo/pkg/replay"
	_ "github.com/bpineau/pofo/pkg/scenario"
)

// TestProgramsRun builds every example program and runs it on the bundled
// data, offline: it must exit cleanly and print something. Their output is
// data-dependent (a data refresh moves every figure), so it is not compared.
func TestProgramsRun(t *testing.T) {
	if testing.Short() {
		t.Skip("builds and runs the example programs")
	}
	args := map[string][]string{
		"oldnew": {"-rev", "HEAD"}, // the shallowest history a checkout has
	}
	dirs, err := filepath.Glob("*/main.go")
	if err != nil || len(dirs) == 0 {
		t.Fatalf("no example program found: %v", err)
	}
	bin := t.TempDir()
	build := exec.Command("go", "build", "-o", bin, "./...")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
	for _, src := range dirs {
		name := filepath.Dir(src)
		t.Run(name, func(t *testing.T) {
			if _, err := os.ReadFile(src); err != nil { // read, so a change to it reruns the test
				t.Fatal(err)
			}
			if name == "oldnew" {
				if err := exec.Command("git", "rev-parse", "HEAD").Run(); err != nil {
					t.Skip("needs a git checkout")
				}
			}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()
			var stdout, stderr bytes.Buffer
			cmd := exec.CommandContext(ctx, filepath.Join(bin, name), args[name]...)
			cmd.Stdout, cmd.Stderr = &stdout, &stderr
			if err := cmd.Run(); err != nil {
				t.Fatalf("%v\n%s", err, stderr.String())
			}
			if stdout.Len() == 0 {
				t.Fatal("printed nothing")
			}
		})
	}
}
