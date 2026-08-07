package marketdata

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"
)

// A fund resolved through the FT path must arrive as clean as a Yahoo symbol.
// The real case: AXA Court Terme (FR0000288946), whose quotes run from 1984 in
// francs and continue in euros, leaving a 6.55957 cliff at the 1999 changeover.
// Before this guard the FIRST fetch of a newly resolved identifier returned the
// unmended series (every later one was mended, since it came back through the
// cached resolution), so a statistic computed the day an asset entered the
// catalog silently read a -85 % day and a decade of nonsense returns.
func TestResolvedFundIsCleanedOnFirstFetch(t *testing.T) {
	const (
		isin  = "FR0000288946"
		franc = 6.55957
	)
	// 40 points in francs, then 40 in euros: one denomination break, and both
	// sides long enough for the repair to trust them.
	start := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	var dates []string
	var closes []float64
	for i := 0; i < 80; i++ {
		dates = append(dates, start.AddDate(0, 0, i).Format("2006-01-02T15:04:05"))
		v := 6000 + float64(i)
		if i >= 40 {
			v /= franc
		}
		closes = append(closes, v)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	})
	mux.HandleFunc("/v1/finance/search", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"quotes":[]}`)
	})
	mux.HandleFunc("/data/searchapi/searchsecurities", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"data":{"security":[{"name":"AXA Court Terme AC","symbol":"%s:EUR","xid":"42","isPrimary":true}]}}`, isin)
	})
	mux.HandleFunc("/data/chartapi/series", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"Dates":[`)
		for i, d := range dates {
			if i > 0 {
				fmt.Fprint(w, ",")
			}
			fmt.Fprintf(w, `"%s"`, d)
		}
		fmt.Fprint(w, `],"Elements":[{"Currency":"EUR","ComponentSeries":[{"Type":"Close","Values":[`)
		for i, v := range closes {
			if i > 0 {
				fmt.Fprint(w, ",")
			}
			fmt.Fprintf(w, "%g", v)
		}
		fmt.Fprint(w, `]}]}]}`)
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	s, err := c.Fetch(context.Background(), isin, time.Time{})
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if len(s.Points) != len(closes) {
		t.Fatalf("%d points, want %d", len(s.Points), len(closes))
	}
	worst := 0.0
	for i := 1; i < len(s.Points); i++ {
		if r := s.Points[i].Close/s.Points[i-1].Close - 1; r < worst {
			worst = r
		}
	}
	if worst < -0.10 {
		t.Errorf("worst daily move %.1f %%: the franc/euro break was not mended on the first fetch", worst*100)
	}
	// The recent segment is authoritative, so the euro side keeps its level.
	if last := s.Points[len(s.Points)-1].Close; last < 900 || last > 940 {
		t.Errorf("last close %.2f, want the euro level (~920)", last)
	}
}
