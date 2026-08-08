package main

import (
	"os"
	"path/filepath"
	"testing"
)

// writeSpec drops a portfolio file in a temporary directory and returns its
// path.
func writeSpec(t *testing.T, name, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// -assets becomes one 100 %-invested portfolio per identifier, in order.
func TestBuildSpecsFromAssets(t *testing.T) {
	specs, err := buildSpecs(nil, " AVWS , ZPRV ,, VOO ", false)
	if err != nil {
		t.Fatal(err)
	}
	if len(specs) != 3 {
		t.Fatalf("%d specs for three identifiers", len(specs))
	}
	for i, want := range []string{"AVWS", "ZPRV", "VOO"} {
		if specs[i].Name != want {
			t.Errorf("spec %d is %q, want %q", i, specs[i].Name, want)
		}
		if n := len(specs[i].Holdings); n != 1 {
			t.Errorf("%s holds %d assets, want one", want, n)
		}
		if specs[i].Sim {
			t.Errorf("%s asks for the backcast without -simulate", want)
		}
	}
}

// -simulate is the whole point of the flag: every spec of the run comes out
// asking for the backcast, files and -assets alike, and the identifiers stay
// exactly as the user typed them (portfolio.Build appends the suffix at fetch
// time, so the report still names what was asked for).
func TestBuildSpecsSimulateAll(t *testing.T) {
	file := writeSpec(t, "p.txt", "60 VTI\n40 IE00BSKRJZ44\n")
	specs, err := buildSpecs([]string{file}, "AVWS,ZPRVSIM", true)
	if err != nil {
		t.Fatal(err)
	}
	if len(specs) != 3 {
		t.Fatalf("%d specs for one file and two identifiers", len(specs))
	}
	for _, s := range specs {
		if !s.Sim {
			t.Errorf("%s does not ask for the backcast under -simulate", s.Name)
		}
	}
	if got := specs[0].Holdings[0].ID; got != "VTI" {
		t.Errorf("the file's holding became %q; -simulate must not rewrite identifiers", got)
	}
	if got := specs[2].Name; got != "ZPRVSIM" {
		t.Errorf("an already-suffixed identifier became %q", got)
	}
}

// Without -simulate a file's own "#meta sim:on" still stands: the flag adds
// the backcast, it never takes one away.
func TestBuildSpecsKeepsTheFileDirective(t *testing.T) {
	file := writeSpec(t, "p.txt", "#meta sim:on\n100 VTI\n")
	specs, err := buildSpecs([]string{file}, "", false)
	if err != nil {
		t.Fatal(err)
	}
	if !specs[0].Sim {
		t.Error("-simulate absent cleared the file's own sim:on directive")
	}
}

// Two files of the same name are told apart, so a report never shows one
// curve twice under one label.
func TestBuildSpecsDisambiguatesNames(t *testing.T) {
	a := writeSpec(t, "p.txt", "100 VTI\n")
	b := writeSpec(t, "p.txt", "100 VOO\n")
	specs, err := buildSpecs([]string{a, b}, "", false)
	if err != nil {
		t.Fatal(err)
	}
	if specs[0].Name == specs[1].Name {
		t.Fatalf("both files kept the name %q", specs[0].Name)
	}
	if want := specs[0].Name + " (2)"; specs[1].Name != want {
		t.Errorf("the second file is %q, want %q", specs[1].Name, want)
	}
}

// A file that does not parse fails the run rather than being skipped.
func TestBuildSpecsReportsAFileError(t *testing.T) {
	if _, err := buildSpecs([]string{filepath.Join(t.TempDir(), "absent.txt")}, "", false); err == nil {
		t.Error("a missing portfolio file must fail the run")
	}
}
