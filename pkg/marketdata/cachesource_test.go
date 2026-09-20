package marketdata

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"
)

// The Financial Times and Morningstar both serve European fund NAVs, and the
// package documentation says never to splice segments of both into one series:
// they stamp the same fund's NAVs on their own dates and levels. The shapes
// below are the real ones, trimmed - FT's chart POST answer and Morningstar's
// COMPACTJSON [[epoch_ms, value], …] - carrying two clearly different levels
// for one ISIN, so a series can be told apart by its first close alone.

const ftCacheISIN = "LU0171310443"

func ftSeriesJSON(days []time.Time, base float64) string {
	dates, vals := "", ""
	for i, d := range days {
		if i > 0 {
			dates, vals = dates+",", vals+","
		}
		dates += fmt.Sprintf("%q", d.Format("2006-01-02T15:04:05"))
		vals += fmt.Sprint(base + float64(i))
	}
	return fmt.Sprintf(`{"Dates":[%s],"Elements":[{"Currency":"EUR","ComponentSeries":[{"Type":"Close","Values":[%s]}]}]}`, dates, vals)
}

func msSeriesJSON(days []time.Time, base float64) string {
	out := ""
	for i, d := range days {
		if i > 0 {
			out += ","
		}
		out += fmt.Sprintf("[%d,%v]", d.UnixMilli(), base+float64(i))
	}
	return "[" + out + "]"
}

// TestFundSourcesDoNotShareACacheFile: an identifier served by the FT today
// and by Morningstar tomorrow must not read the other's file back. Keyed by
// the identifier alone, the Morningstar path answered with the FT history -
// silently, and through the stale-cache fallback even during an outage.
func TestFundSourcesDoNotShareACacheFile(t *testing.T) {
	days := testDays(80)
	mux := http.NewServeMux()
	mux.HandleFunc("/data/chartapi/series", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, ftSeriesJSON(days, 100))
	})
	mux.HandleFunc("/api/rest.svc/timeseries_price/"+morningstarToken, func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, msSeriesJSON(days, 500))
	})
	dir := t.TempDir()
	c, srv := newTestClient(t, dir, mux)
	defer srv.Close()
	ctx, from := context.Background(), days[0]

	ft, err := c.historyFT(ctx, ftCacheISIN, resolution{Source: "ft", Xid: "12345", Name: "FT line"}, from, false)
	if err != nil {
		t.Fatal(err)
	}
	if ft.First().Close != 100 {
		t.Fatalf("FT series starts at %v, want 100", ft.First().Close)
	}

	// A fresh client, so only the DISK cache can answer.
	c2, srv2 := newTestClient(t, dir, mux)
	defer srv2.Close()
	ms, err := c2.historyMS(ctx, ftCacheISIN, resolution{Source: "morningstar", Symbol: "F0GBR05B68", Name: "MS line"}, from, false)
	if err != nil {
		t.Fatal(err)
	}
	if ms.First().Close != 500 {
		t.Fatalf("the Morningstar path served %v, the FT file's own level: one cache file for two sources",
			ms.First().Close)
	}
	if ms.Source != "morningstar" {
		t.Fatalf("series source %q, want morningstar", ms.Source)
	}
}

// TestCachedFollowsTheSourceIdentity: Client.Cached is what the web app bills
// a visitor's fetch budget against, so it must look where the fetch path
// actually wrote.
func TestCachedFollowsTheSourceIdentity(t *testing.T) {
	days := testDays(80)
	mux := http.NewServeMux()
	mux.HandleFunc("/data/chartapi/series", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, ftSeriesJSON(days, 100))
	})
	dir := t.TempDir()
	c, srv := newTestClient(t, dir, mux)
	defer srv.Close()

	if c.Cached(ftCacheISIN) {
		t.Fatal("nothing is cached yet")
	}
	res := resolution{Source: "ft", Xid: "12345", Name: "FT line"}
	if _, err := c.historyFT(context.Background(), ftCacheISIN, res, resolveFrom(), false); err != nil {
		t.Fatal(err)
	}
	c.saveResolution(ftCacheISIN, res)
	if !c.Cached(ftCacheISIN) {
		t.Fatal("the FT history is on disk and Cached must say so")
	}
}
