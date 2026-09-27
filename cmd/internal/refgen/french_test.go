package refgen

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"
)

// sorts mimics a portfolio-sort file: prose, the value-weighted table, then an
// equal-weighted one under the same column names, then a copyright line.
const sorts = `This file was created by using the 202608 CRSP database.
Missing data are indicated by -99.99 or -999.

  Average Value Weighted Returns -- Daily
,SMALL LoBM,ME1 BM2,SMALL HiBM,BIG LoBM
19630701,   0.10,  -0.09,  -0.76,   0.16
19630702,   0.14,  -0.19,   0.71,   0.52


  Average Equal Weighted Returns -- Daily
,SMALL LoBM,ME1 BM2,SMALL HiBM,BIG LoBM
19630701,   9.10,   9.09,   9.37,   9.16

Copyright 2026 Eugene F. Fama and Kenneth R. French
`

func TestFrenchTableReadsTheFirstTable(t *testing.T) {
	dates, vals, err := FrenchTable(sorts, "SMALL HiBM", "BIG LoBM")
	if err != nil {
		t.Fatal(err)
	}
	if len(dates) != 2 || dates[1].Format("2006-01-02") != "1963-07-02" {
		t.Fatalf("dates %v, want the two value-weighted days only", dates)
	}
	if vals[0][0] != -0.76 || vals[0][1] != 0.71 || vals[1][1] != 0.52 {
		t.Errorf("values %v, want the requested columns in percent, in the order asked", vals)
	}
}

func TestFrenchTableRefusesBadSources(t *testing.T) {
	if _, _, err := FrenchTable(strings.Replace(sorts, "0.71", "-99.99", 1), "SMALL HiBM"); err == nil {
		t.Error("a missing-value sentinel was accepted")
	}
	if _, _, err := FrenchTable(sorts, "Mkt-RF"); err == nil {
		t.Error("a file without the requested column was accepted")
	}
}

func TestFrenchCSV(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, _ := zw.Create("F.csv")
	w.Write([]byte(sorts))
	zw.Close()
	body, err := FrenchCSV(buf.Bytes())
	if err != nil || body != sorts {
		t.Errorf("FrenchCSV: err %v, body intact %v", err, body == sorts)
	}
	if _, err := FrenchCSV([]byte("not a zip")); err == nil {
		t.Error("a non-zip was accepted")
	}
}
