// Command gen-broadsample builds the bundled broad-sample world real-return
// panel (pkg/datasets/broadsample/country-real.csv) from the Jorda-Schularick-
// Taylor Macrohistory Database. It runs at data-generation time only (network);
// the pofo binary never fetches JST, it embeds the committed CSV.
//
// Source: O. Jorda, M. Schularick, A. M. Taylor, "Macrofinancial History and
// the New Business Cycle Facts", NBER Macroeconomics Annual 2016. Database R6,
// 18 advanced economies, 1870-2020, free to use with citation.
//
// Method: for each country-year, nominal total returns (eq_tr, bond_tr) and the
// bill rate are deflated by that country's CPI to real returns. The output is a
// long per-country table (iso,year,equity,bond,bill), NOT a pre-aggregated world
// index: the FIRE explorer pool-bootstraps runs from single markets, so national
// disasters (France/Portugal early-century, Japan post-1990) survive at full
// force. Pre-diversifying into a world index would erase exactly the sequence
// risk this dataset exists to capture.
//
// Usage:
//
//	gen-broadsample [-url URL] [-o path] [-dry]
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
	"sort"
	"strconv"
	"strings"
	"time"
)

const defaultURL = "https://www.macrohistory.net/app/download/9834512569/JSTdatasetR6.xlsx"

func main() {
	url := flag.String("url", defaultURL, "JST xlsx download URL")
	out := flag.String("o", "pkg/datasets/broadsample/country-real.csv", "output CSV path")
	dryRun := flag.Bool("dry", false, "print coverage and moments without writing")
	flag.Parse()

	raw, err := download(*url)
	if err != nil {
		log.Fatalf("download: %v", err)
	}
	rows, err := readSheet(raw)
	if err != nil {
		log.Fatalf("parse xlsx: %v", err)
	}
	recs, err := realReturns(rows)
	if err != nil {
		log.Fatalf("compute: %v", err)
	}
	logMoments(recs)
	if *dryRun {
		return
	}
	if err := writeCSV(*out, *url, recs); err != nil {
		log.Fatalf("write: %v", err)
	}
	log.Printf("wrote %s (%d rows)", *out, len(recs))
}

func download(url string) ([]byte, error) {
	c := &http.Client{Timeout: 120 * time.Second}
	resp, err := c.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %s", resp.Status)
	}
	return io.ReadAll(resp.Body)
}

// cell is one spreadsheet cell: Ref like "AO2", T the type ("s" for a shared
// string index), V the raw value.
type cell struct {
	Ref string `xml:"r,attr"`
	T   string `xml:"t,attr"`
	V   string `xml:"v"`
}

// readSheet returns the worksheet rows as maps of column letter -> resolved
// string value (shared strings dereferenced), the header row included.
func readSheet(raw []byte) ([]map[string]string, error) {
	zr, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		return nil, err
	}
	shared, err := readShared(zr)
	if err != nil {
		return nil, err
	}
	f, err := open(zr, "xl/worksheets/sheet1.xml")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var sheet struct {
		Rows []struct {
			Cells []cell `xml:"c"`
		} `xml:"sheetData>row"`
	}
	if err := xml.NewDecoder(f).Decode(&sheet); err != nil {
		return nil, err
	}
	out := make([]map[string]string, 0, len(sheet.Rows))
	for _, r := range sheet.Rows {
		m := make(map[string]string, len(r.Cells))
		for _, c := range r.Cells {
			v := c.V
			if c.T == "s" {
				i, err := strconv.Atoi(c.V)
				if err != nil || i < 0 || i >= len(shared) {
					continue
				}
				v = shared[i]
			}
			m[colLetters(c.Ref)] = v
		}
		out = append(out, m)
	}
	return out, nil
}

func readShared(zr *zip.Reader) ([]string, error) {
	f, err := open(zr, "xl/sharedStrings.xml")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var sst struct {
		SI []struct {
			T string `xml:"t"`
		} `xml:"si"`
	}
	if err := xml.NewDecoder(f).Decode(&sst); err != nil {
		return nil, err
	}
	out := make([]string, len(sst.SI))
	for i, si := range sst.SI {
		out[i] = si.T
	}
	return out, nil
}

func open(zr *zip.Reader, name string) (io.ReadCloser, error) {
	for _, f := range zr.File {
		if f.Name == name {
			return f.Open()
		}
	}
	return nil, fmt.Errorf("%s not found in xlsx", name)
}

// colLetters strips the trailing row digits from a cell ref ("AO2" -> "AO").
func colLetters(ref string) string {
	i := strings.IndexFunc(ref, func(r rune) bool { return r >= '0' && r <= '9' })
	if i < 0 {
		return ref
	}
	return ref[:i]
}

// record is one country-year of real returns.
type record struct {
	ISO                string
	Year               int
	Equity, Bond, Bill float64
	HasBond, HasBill   bool
}

// realReturns deflates each country-year's nominal returns by its own CPI. Only
// rows with an equity return (and the previous year's CPI) are kept.
func realReturns(rows []map[string]string) ([]record, error) {
	if len(rows) < 2 {
		return nil, fmt.Errorf("no data rows")
	}
	col := map[string]string{}
	for letters, name := range rows[0] {
		col[name] = letters
	}
	for _, n := range []string{"year", "iso", "cpi", "eq_tr", "bond_tr", "bill_rate"} {
		if col[n] == "" {
			return nil, fmt.Errorf("missing JST column %q", n)
		}
	}
	num := func(r map[string]string, name string) (float64, bool) {
		v, ok := r[col[name]]
		if !ok || v == "" {
			return 0, false
		}
		f, err := strconv.ParseFloat(v, 64)
		return f, err == nil
	}

	// Index CPI by (iso, year) for the previous-year lookup.
	type raw struct {
		iso                             string
		cpi, eq, bond, bill             float64
		hasCPI, hasEq, hasBond, hasBill bool
	}
	byYear := map[string]map[int]raw{}
	for _, r := range rows[1:] {
		y, ok := num(r, "year")
		if !ok {
			continue
		}
		iso := r[col["iso"]]
		if iso == "" {
			continue
		}
		var e raw
		e.iso = iso
		e.cpi, e.hasCPI = num(r, "cpi")
		e.eq, e.hasEq = num(r, "eq_tr")
		e.bond, e.hasBond = num(r, "bond_tr")
		e.bill, e.hasBill = num(r, "bill_rate")
		if byYear[iso] == nil {
			byYear[iso] = map[int]raw{}
		}
		byYear[iso][int(y)] = e
	}

	var out []record
	for iso, ser := range byYear {
		for y, r := range ser {
			p, okp := ser[y-1]
			if !r.hasEq || !r.hasCPI || !okp || !p.hasCPI || p.cpi == 0 {
				continue
			}
			infl := r.cpi/p.cpi - 1
			real := func(nom float64) float64 { return (1+nom)/(1+infl) - 1 }
			rec := record{ISO: iso, Year: y, Equity: real(r.eq)}
			if r.hasBond {
				rec.Bond, rec.HasBond = real(r.bond), true
			}
			if r.hasBill {
				rec.Bill, rec.HasBill = real(r.bill), true
			}
			out = append(out, rec)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].ISO != out[j].ISO {
			return out[i].ISO < out[j].ISO
		}
		return out[i].Year < out[j].Year
	})
	return out, nil
}

func writeCSV(path, url string, recs []record) error {
	var b strings.Builder
	fmt.Fprintf(&b, "# Per-country real annual total returns (fractions), 18 advanced economies.\n")
	fmt.Fprintf(&b, "# Source: Jorda-Schularick-Taylor Macrohistory Database R6\n")
	fmt.Fprintf(&b, "# (%s), each series deflated by its own CPI. Cite JST when reusing.\n", url)
	fmt.Fprintf(&b, "# The FIRE explorer pool-bootstraps single-market runs from this table, so\n")
	fmt.Fprintf(&b, "# national disasters survive; it is deliberately NOT a diversified world index.\n")
	fmt.Fprintf(&b, "# Regenerate: make broadsample\n")
	fmt.Fprintf(&b, "iso,year,equity,bond,bill\n")
	for _, r := range recs {
		fmt.Fprintf(&b, "%s,%d,%s,%s,%s\n", r.ISO, r.Year,
			strconv.FormatFloat(r.Equity, 'f', 6, 64), f(r.Bond, r.HasBond), f(r.Bill, r.HasBill))
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func f(v float64, ok bool) string {
	if !ok {
		return ""
	}
	return strconv.FormatFloat(v, 'f', 6, 64)
}

// logMoments reports the pooled (all country-years) equity moments, the honest
// summary of a random developed-market retiree's draw.
func logMoments(recs []record) {
	var sum, logp float64
	byISO := map[string]int{}
	for _, r := range recs {
		sum += r.Equity
		logp += math.Log(1 + r.Equity)
		byISO[r.ISO]++
	}
	n := float64(len(recs))
	mean := sum / n
	geo := math.Exp(logp/n) - 1
	var vsum float64
	for _, r := range recs {
		vsum += (r.Equity - mean) * (r.Equity - mean)
	}
	log.Printf("pooled equity: %d rows, %d countries, arith=%.4f geo=%.4f vol=%.4f",
		len(recs), len(byISO), mean, geo, math.Sqrt(vsum/n))
}
