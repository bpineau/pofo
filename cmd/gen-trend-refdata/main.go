// Command gen-trend-refdata rebuilds pkg/datasets/refdata/TREND-TSMOM-USD.csv,
// the monthly reference path of time-series momentum that the managed-futures
// reconstructions are anchored on.
//
// It runs at data-generation time only (network); the pofo binary embeds the
// CSV and never fetches anything.
//
// WHY THIS EXISTS. The trend reconstructions used to invent their own monthly
// path: a seven-market book whose month-to-month returns correlated about 0.55
// with the funds it replicates, because seven markets cannot reproduce what a
// programme trading fifty does. Widening the basket with what the client can
// fetch does not fix it (the currency legs only start in 2003-2006, which is
// exactly the trend drought, and four of the six carry a negative standalone
// trend information ratio over that window). Anchoring the monthly path on a
// published record instead lifts the agreement with the real funds to 0.60-0.71
// while the engine keeps supplying the daily texture, which is all it was ever
// good for.
//
// THE SERIES. Monthly excess-of-cash returns of the canonical time-series
// momentum factor: a 12-month signal with a one-month holding period, applied
// across equity indices, fixed income, currencies and commodities, each market
// scaled by its own volatility. It is the construction of Moskowitz, Ooi and
// Pedersen, "Time Series Momentum", Journal of Financial Economics 104(2),
// 2012, kept current by its authors and published as a monthly spreadsheet at
// the URL in -src below. It is a gross academic factor: it carries no fee, no
// slippage and no capacity limit, and realizes an information ratio near 0.9
// where investable programmes deliver 0.25 to 0.4. Nothing anchors on it any
// more: every reconstruction now takes its months AND its level from a
// composite of real programmes already net of their managers' fees
// (cmd/gen-trendnet-refdata, cmd/gen-sgtrend-refdata). This series stays
// bundled as the SHAPE yardstick a gross factor is good for.
//
// The CSV holds a month-end level index (base 100) of those excess returns, so
// simgen can rescale it to any volatility target and add its own cash leg.
//
// Usage: gen-trend-refdata [-src URL] [-dir path] [-dry]
package main

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"
)

const defaultSrc = "https://www.aqr.com/-/media/AQR/Documents/Insights/Data-Sets/Time-Series-Momentum-Factors-Monthly.xlsx"

func main() {
	src := flag.String("src", defaultSrc, "monthly factor spreadsheet")
	dir := flag.String("dir", "pkg/datasets/refdata", "output directory")
	dry := flag.Bool("dry", false, "report without writing")
	flag.Parse()

	raw, err := download(*src)
	if err != nil {
		log.Fatalf("download: %v", err)
	}
	points, err := parse(raw)
	if err != nil {
		log.Fatalf("parse: %v", err)
	}
	if len(points) < 240 {
		log.Fatalf("only %d monthly returns: the source layout probably moved", len(points))
	}

	// Sanity before anything is written: a trend factor over four decades has
	// a positive but not absurd information ratio, and a double-digit
	// volatility. Anything outside these bands means the wrong column.
	vol, ir := stats(points)
	log.Printf("%d months, %s to %s, excess vol %.1f%%/yr, information ratio %.2f",
		len(points), points[0].date.Format("2006-01"), points[len(points)-1].date.Format("2006-01"), vol*100, ir)
	if vol < 0.05 || vol > 0.30 || ir < 0.3 || ir > 1.5 {
		log.Fatalf("out of band: vol %.1f%% ir %.2f", vol*100, ir)
	}

	var b bytes.Buffer
	fmt.Fprintf(&b, "# pofo simdata v1\n# id: TREND-TSMOM-USD\n")
	fmt.Fprintf(&b, "# name: time-series momentum, monthly excess-of-cash index (USD, month-end)\n")
	fmt.Fprintf(&b, "# source: published monthly returns of the Moskowitz-Ooi-Pedersen time-series-momentum factor "+
		"(12-month signal, 1-month holding, cross-asset, volatility-scaled per market), compounded into a base-100 "+
		"excess index. Gross of fees and capacity: the recipes rescale it to their volatility target and pin its "+
		"information ratio to the index each fund tracks. Regenerate with cmd/gen-trend-refdata.\n")
	fmt.Fprintf(&b, "date,close\n")
	level := 100.0
	fmt.Fprintf(&b, "%s,%.6f\n", points[0].date.AddDate(0, -1, 0).Format("2006-01-02"), level)
	for _, p := range points {
		level *= 1 + p.ret
		fmt.Fprintf(&b, "%s,%.6f\n", p.date.Format("2006-01-02"), level)
	}
	if *dry {
		log.Printf("dry run: %d bytes, final level %.1f", b.Len(), level)
		return
	}
	out := filepath.Join(*dir, "TREND-TSMOM-USD.csv")
	if err := os.WriteFile(out, b.Bytes(), 0o644); err != nil {
		log.Fatalf("write: %v", err)
	}
	log.Printf("wrote %s (%d points)", out, len(points)+1)
	log.Printf("rebuild (make simdata) to re-anchor the trend reconstructions")
}

type point struct {
	date time.Time
	ret  float64
}

func download(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http %d", resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 32<<20))
}

// parse reads the first worksheet of the workbook: one row per month, the
// date as an Excel serial in the first column and the all-asset factor's
// excess return in the second.
func parse(raw []byte) ([]point, error) {
	zr, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		return nil, err
	}
	var sheet []byte
	for _, f := range zr.File {
		if f.Name == "xl/worksheets/sheet1.xml" {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			sheet, err = io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return nil, err
			}
		}
	}
	if sheet == nil {
		return nil, fmt.Errorf("no first worksheet in the workbook")
	}

	type cell struct {
		Ref   string `xml:"r,attr"`
		Type  string `xml:"t,attr"`
		Value string `xml:"v"`
	}
	type row struct {
		Cells []cell `xml:"c"`
	}
	var doc struct {
		Rows []row `xml:"sheetData>row"`
	}
	if err := xml.Unmarshal(sheet, &doc); err != nil {
		return nil, err
	}

	var out []point
	for _, r := range doc.Rows {
		var serial, ret string
		for _, c := range r.Cells {
			if c.Type == "s" { // a shared string: a header row, not data
				continue
			}
			switch column(c.Ref) {
			case "A":
				serial = c.Value
			case "B":
				ret = c.Value
			}
		}
		if serial == "" || ret == "" {
			continue
		}
		days, err := strconv.ParseFloat(serial, 64)
		if err != nil {
			continue
		}
		v, err := strconv.ParseFloat(ret, 64)
		if err != nil {
			continue
		}
		if days < 20000 || math.Abs(v) > 0.6 { // not a date, or not a monthly return
			continue
		}
		out = append(out, point{excelDate(days), v})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].date.Before(out[j].date) })
	return out, nil
}

func column(ref string) string {
	for i, c := range ref {
		if c >= '0' && c <= '9' {
			return ref[:i]
		}
	}
	return ref
}

// excelDate converts a spreadsheet serial day to a UTC date (the 1900 system,
// with its deliberate leap-year bug already absorbed by the 1899-12-30 epoch).
func excelDate(days float64) time.Time {
	return time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC).AddDate(0, 0, int(days))
}

func stats(points []point) (vol, ir float64) {
	var m float64
	for _, p := range points {
		m += p.ret
	}
	m /= float64(len(points))
	var s float64
	for _, p := range points {
		s += (p.ret - m) * (p.ret - m)
	}
	vol = math.Sqrt(s/float64(len(points)-1)) * math.Sqrt(12)
	g := 1.0
	for _, p := range points {
		g *= 1 + p.ret
	}
	cagr := math.Pow(g, 12/float64(len(points))) - 1
	if vol > 0 {
		ir = cagr / vol
	}
	return vol, ir
}
