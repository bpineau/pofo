package refgen

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

func day(y, m, d int) time.Time { return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC) }

func pts(vals ...float64) []marketdata.Point {
	out := make([]marketdata.Point, len(vals))
	for i, v := range vals {
		out[i] = marketdata.Point{Date: day(2020, 1+i, 1), Close: v}
	}
	return out
}

func TestFRED(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/graph/fredgraph.csv" || r.URL.Query().Get("id") != "TB3MS" {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte("observation_date,TB3MS\n1934-01-01,0.72\n1934-02-01,.\n1934-03-01,0.24\n"))
	}))
	defer srv.Close()
	got, err := FRED(srv.URL, "TB3MS")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || !got[1].Date.Equal(day(1934, 3, 1)) || got[1].Close != 0.24 {
		t.Errorf("got %v, want the two observations with the '.' gap skipped", got)
	}
	if _, err := FRED(srv.URL, "NOPE"); err == nil {
		t.Error("a 404 was accepted")
	}
}

func TestSameHistory(t *testing.T) {
	old := pts(1, 2, 3)
	if msg, err := SameHistory(old, pts(1, 2, 3, 4), 1); err != nil || !strings.Contains(msg, "3 bundled points reproduced, 1 new") {
		t.Errorf("an extension: %q, %v", msg, err)
	}
	if msg, err := SameHistory(old, pts(1, 2, 3.5, 4), 1); err != nil || !strings.Contains(msg, "revised at the tail") {
		t.Errorf("a revision of the last point, allowed: %q, %v", msg, err)
	}
	if _, err := SameHistory(old, pts(1, 2.5, 3, 4), 1); err == nil {
		t.Error("a revision of a firm point was accepted")
	}
	if _, err := SameHistory(old, pts(1, 2), 0); err == nil {
		t.Error("a history that lost its last point was accepted")
	}
}

func TestMonthlyCadence(t *testing.T) {
	if err := MonthlyCadence(pts(1, 2, 3)); err != nil {
		t.Errorf("a clean monthly series: %v", err)
	}
	hole := pts(1, 2, 3)
	hole[2].Date = day(2020, 4, 1)
	if err := MonthlyCadence(hole); err == nil {
		t.Error("a missing month was accepted")
	}
	mid := pts(1, 2)
	mid[1].Date = day(2020, 2, 15)
	if err := MonthlyCadence(mid); err == nil {
		t.Error("a mid-month label was accepted")
	}
}

func TestFlatRun(t *testing.T) {
	p := pts(0.38, 0.38, 0.38, 1, 2, 2, 3)
	if n, end := FlatRun(p, time.Time{}); n != 3 || !end.Equal(day(2020, 3, 1)) {
		t.Errorf("whole series: run %d ending %v, want 3 ending 2020-03-01", n, end)
	}
	if n, _ := FlatRun(p, day(2020, 4, 1)); n != 2 {
		t.Errorf("after the pegged months: run %d, want 2", n)
	}
}

func TestWriteReadsBack(t *testing.T) {
	dir := t.TempDir()
	h := Header{ID: "X-RATE", Name: "a rate", Source: "somewhere", Junctions: []time.Time{day(2020, 2, 1)},
		Ends: day(2020, 2, 1), EndsWhy: "stopped"}
	if err := Write(dir, h, pts(1.5, 2.5)); err != nil {
		t.Fatal(err)
	}
	s, ok, err := marketdata.ReadSimdata(dir, "X-RATE")
	if err != nil || !ok {
		t.Fatalf("read back: ok=%v err=%v", ok, err)
	}
	if s.Name != "a rate" || s.Len() != 2 || !s.Ends.Equal(day(2020, 2, 1)) || len(s.Junctions) != 1 {
		t.Errorf("read back %q, %d points, junctions %v, ends %v", s.Name, s.Len(), s.Junctions, s.Ends)
	}
}
