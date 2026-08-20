package main

import (
	"bytes"
	"log"
	"strconv"
	"strings"
	"sync"
	"testing"
)

// An informational line is printed once however often the fetch path repeats
// it, while a warning is printed every time: its repetition is the signal.
func TestDedupWriter(t *testing.T) {
	var buf bytes.Buffer
	lg := log.New(newDedupWriter(&buf), "", 0)

	const (
		info  = "IE00077IIPQ8: history extended via simdata starting 1969-12-31"
		other = "NTSG -> IE00077IIPQ8"
		warn  = "warning: FX rate USD->EUR unavailable before 1971-01-04, held constant earlier"
	)
	lg.Print(info)
	lg.Print(info)
	lg.Print(other)
	lg.Print(warn)
	lg.Print(warn)

	got := buf.String()
	for line, want := range map[string]int{info: 1, other: 1, warn: 2} {
		if n := strings.Count(got, line); n != want {
			t.Errorf("%q printed %d time(s), want %d\nlog:\n%s", line, n, want, got)
		}
	}
}

// The line is the unit: two lines that differ only in their identifier are two
// lines, and both are printed.
func TestDedupWriterKeepsDistinctLines(t *testing.T) {
	var buf bytes.Buffer
	lg := log.New(newDedupWriter(&buf), "", 0)
	lg.Print("resolved NTSG -> \"WisdomTree\"")
	lg.Print("resolved IWDA -> \"iShares\"")
	if n := strings.Count(buf.String(), "\n"); n != 2 {
		t.Errorf("%d lines, want 2:\n%s", n, buf.String())
	}
}

// The memory is bounded: past dedupLimit distinct lines the writer forgets
// everything rather than growing without end, and prints again.
func TestDedupWriterForgetsPastTheLimit(t *testing.T) {
	var buf bytes.Buffer
	d := newDedupWriter(&buf)
	first := []byte("first line\n")
	if _, err := d.Write(first); err != nil {
		t.Fatal(err)
	}
	for i := range dedupLimit {
		if _, err := d.Write([]byte("line " + strconv.Itoa(i) + "\n")); err != nil {
			t.Fatal(err)
		}
	}
	buf.Reset()
	if _, err := d.Write(first); err != nil {
		t.Fatal(err)
	}
	if buf.Len() == 0 {
		t.Error("the writer still remembers a line from before the limit")
	}
}

// Concurrent requests log through the same writer; the race detector and the
// count both have to be happy.
func TestDedupWriterConcurrent(t *testing.T) {
	var buf bytes.Buffer
	lg := log.New(newDedupWriter(&buf), "", 0)
	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			lg.Print("the same informational line")
		}()
	}
	wg.Wait()
	if n := strings.Count(buf.String(), "\n"); n != 1 {
		t.Errorf("%d lines, want 1", n)
	}
}
