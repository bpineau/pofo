package refgen

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// fastOECD shortens the politeness delays for the length of a test.
func fastOECD(t *testing.T) {
	t.Helper()
	spacing, backoff := oecdSpacing, oecdBackoff
	oecdSpacing, oecdBackoff = 0, time.Millisecond
	t.Cleanup(func() { oecdSpacing, oecdBackoff = spacing, backoff })
}

const finmarkCSV = "DATAFLOW,REF_AREA,FREQ,MEASURE,UNIT_MEASURE,ACTIVITY,ADJUSTMENT,TRANSFORMATION,TIME_HORIZ,METHODOLOGY,TIME_PERIOD,OBS_VALUE,OBS_STATUS,UNIT_MULT,DECIMALS,BASE_PER\n" +
	"OECD.SDD.STES:DSD_STES@DF_FINMARK(4.0),EA20,M,IRLT,PA,_Z,_Z,_Z,_Z,N,2026-03,3.385728,A,0,2,\n" +
	"OECD.SDD.STES:DSD_STES@DF_FINMARK(4.0),EA20,M,IRLT,PA,_Z,_Z,_Z,_Z,N,2026-01,3.220964,A,0,2,\n" +
	"OECD.SDD.STES:DSD_STES@DF_FINMARK(4.0),EA20,M,IRLT,PA,_Z,_Z,_Z,_Z,N,2026-02,,M,0,2,\n" +
	"OECD.SDD.STES:DSD_STES@DF_FINMARK(4.0),DEU,M,IRLT,PA,_Z,_Z,_Z,_Z,N,2026-01,2.8,A,0,2,\n" +
	"OECD.SDD.STES:DSD_STES@DF_FINMARK(4.0),DEU,M,SHARE,PA,_Z,_Z,_Z,_Z,N,2026-01,9,A,0,2,\n" +
	"OECD.SDD.STES:DSD_STES@DF_FINMARK(4.0),EA20,M,SHARE,IX,_Z,_Z,_Z,_Z,N,2026-01,151.2,A,0,2,2015\n"

func TestOECDBatchesKeysIntoOneDownload(t *testing.T) {
	fastOECD(t)
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		want := "/data/OECD.SDD.STES,DSD_STES@DF_FINMARK,/EA20+DEU+GBR.M.IRLT+SHARE.PA+IX._Z._Z._Z._Z.N"
		if r.URL.Path != want {
			t.Errorf("path %q, want %q", r.URL.Path, want)
		}
		w.Write(append(byteOrderMark, finmarkCSV...))
	}))
	defer srv.Close()
	got, err := OECD(srv.URL, FinMark,
		"EA20.M.IRLT.PA._Z._Z._Z._Z.N", "DEU.M.IRLT.PA._Z._Z._Z._Z.N",
		"EA20.M.SHARE.IX._Z._Z._Z._Z.N", "GBR.M.SHARE.IX._Z._Z._Z._Z.N")
	if err != nil {
		t.Fatal(err)
	}
	if n := calls.Load(); n != 1 {
		t.Errorf("%d downloads, want one", n)
	}
	ea := got["EA20.M.IRLT.PA._Z._Z._Z._Z.N"]
	if len(ea) != 2 || !ea[0].Date.Equal(day(2026, 1, 1)) || ea[1].Close != 3.385728 {
		t.Errorf("EA20 IRLT = %v, want 2026-01 and 2026-03 in order, the empty February skipped", ea)
	}
	if len(got) != 3 {
		t.Errorf("got %d series, want the three asked for that exist (not DEU SHARE PA, nobody asked)", len(got))
	}
	if _, ok := got["GBR.M.SHARE.IX._Z._Z._Z._Z.N"]; ok {
		t.Error("a series the table does not carry was invented")
	}
}

func TestOECDRetriesAThrottledDownload(t *testing.T) {
	fastOECD(t)
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) < 3 {
			http.Error(w, "slow down", http.StatusTooManyRequests)
			return
		}
		w.Write([]byte(finmarkCSV))
	}))
	defer srv.Close()
	got, err := OECD(srv.URL, FinMark, "DEU.M.IRLT.PA._Z._Z._Z._Z.N")
	if err != nil || len(got["DEU.M.IRLT.PA._Z._Z._Z._Z.N"]) != 1 {
		t.Fatalf("after two 429s: %v, %v", got, err)
	}
	if n := calls.Load(); n != 3 {
		t.Errorf("%d attempts, want 3", n)
	}
}

func TestOECDAbsenceAndRefusal(t *testing.T) {
	fastOECD(t)
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.URL.Path == "/data/OECD.SDD.TPS,DSD_PRICES@DF_PRICES_ALL,/ZZZ.M.N.CPI.IX._T.N._Z" {
			http.Error(w, "NoRecordsFound", http.StatusNotFound)
			return
		}
		http.Error(w, "bad query", http.StatusBadRequest)
	}))
	defer srv.Close()
	got, err := OECD(srv.URL, Prices, "ZZZ.M.N.CPI.IX._T.N._Z")
	if err != nil || len(got) != 0 {
		t.Errorf("a 404 is an empty answer, not a failure: %v, %v", got, err)
	}
	calls.Store(0)
	if _, err := OECD(srv.URL, Prices, "YYY.M.N.CPI.IX._T.N._Z"); err == nil {
		t.Error("a 400 was accepted")
	}
	if n := calls.Load(); n != 1 {
		t.Errorf("a 400 was tried %d times; a refusal is not retried", n)
	}
	if _, err := OECD(srv.URL, Prices, "A.M", "B.M.X"); err == nil {
		t.Error("keys of different lengths were accepted")
	}
}

func TestCompareSteps(t *testing.T) {
	old := pts(100, 101, 102.01, 103.0301)
	if o := CompareSteps(old, pts(200, 202, 204.02, 206.0602, 210)); o.Steps != 3 || o.Reproduced != 3 {
		t.Errorf("a rebased extension: %s", o)
	}
	moved := pts(100, 101, 103, 104.0301)
	o := CompareSteps(old, moved)
	if o.Steps != 3 || o.Reproduced != 1 || !o.First.Equal(day(2020, 3, 1)) || !o.Last.Equal(day(2020, 4, 1)) {
		t.Errorf("one revised level moves the two steps around it: %s", o)
	}
	if o.Share() >= 1 {
		t.Errorf("share %v", o.Share())
	}
	if (Overlap{}).Share() != 1 {
		t.Error("no common step reads as full agreement")
	}
}
