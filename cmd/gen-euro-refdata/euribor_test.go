package main

import (
	"math"
	"testing"
	"time"
)

func TestEuriborAccruesTheLastPublishedCompleteMonth(t *testing.T) {
	const header = "OBS_VALUE,KEY,TIME_PERIOD\n"
	rates, err := parseEuribor([]byte(header + "-1," + euriborKey + ",1994-01\n3," + euriborKey + ",1994-02\n"))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		asOf   time.Time
		months int
	}{
		{date(1994, 3).AddDate(0, 0, -1), 1},
		{date(1994, 3), 2},
	} {
		points := atAccrualEnd(accrue(rates, tc.asOf, simpleAct360))
		if len(points) != tc.months+1 {
			t.Fatalf("as of %v: got %d levels", tc.asOf, len(points))
		}
		last := points[len(points)-1]
		want := 100 * (1 - 0.01*31/360)
		if tc.months == 2 {
			want *= 1 + 0.03*28/360
		}
		if !last.Date.Equal(date(1994, tc.months+1).AddDate(0, 0, -1)) || math.Abs(last.Close-want) > 1e-12 {
			t.Errorf("as of %v: wrong month-end level %+v, expected %.8f", tc.asOf, last, want)
		}
	}
	for _, row := range []string{"1,other,1994-01", "NaN," + euriborKey + ",1994-01", "3," + euriborKey + ",1994-02"} {
		if _, err := parseEuribor([]byte(header + row)); err == nil {
			t.Errorf("accepted %q", row)
		}
	}
}
