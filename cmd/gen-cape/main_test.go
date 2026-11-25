package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWritePreservesRecentCoverage(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cape.csv")
	const previous = "# provisional observation\ndate,cape\n2026-11-01,39.37\n"
	if err := os.WriteFile(path, []byte(previous), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := write(path, "mirror", []point{{date: "2023-09-01", cape: 29}}); err == nil {
		t.Fatal("accepted a truncated download")
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != previous {
		t.Fatalf("previous observations changed: %q, %v", data, err)
	}
	if err := write(path, "mirror", []point{{date: "2026-11-01", cape: 39.38}}); err != nil {
		t.Fatalf("revision of the latest month: %v", err)
	}
}
