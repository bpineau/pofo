// Command gen-snapshots refreshes the offline fallback snapshots embedded in
// pkg/marketdata/data/: the series pofo serves when its live sources are
// unreachable (^VIX, ^CPI-US, ^HICP-FR) and the four long daily dollar crosses
// that let a EUR, GBP, JPY or CHF investor backcast to 1971.
//
// WHY THIS EXISTS. Each of those seven files was hand-built once, from a source
// documented only in a comment header or a godoc, and then drifted: by 2026-08
// the snapshots trailed their sources by one to six months. Refreshing them by
// hand is exactly the kind of unreproducible step this repository avoids
// elsewhere, so the recipes now live here, one function per file.
//
// Sourcing, one line per file:
//
//   - vix.csv        CBOE's own daily VIX history (date + close), 1990->.
//   - cpi-us.csv     FRED CPIAUCNS, the BLS all-items CPI-U, not seasonally
//     adjusted, 1913->.
//   - gbpusd-long.csv, jpyusd-long.csv, chfusd-long.csv: FRED's daily H.10 noon
//     rates (DEXUSUK, DEXJPUS, DEXSZUS), 1971->, stored as USD per foreign
//     unit, so the two quoted the other way round are inverted.
//   - eurusd-long.csv  FRED DEXUSEU from 1999-01-04. The pre-euro head (ECU
//     anchors carrying the Frankfurt DM/USD shape) is settled history built
//     from two sources this command does not touch, so it is carried over from
//     the existing file unchanged.
//   - hicp-fr.csv    Eurostat prc_hicp_midx (geo=FR, all-items, 2015=100) from
//     1996-01, likewise over a carried-over OECD head chained at the 1996
//     overlap; the command re-checks that overlap and refuses to write if
//     Eurostat has rebased under it.
//
// Every file keeps its own comment header verbatim (curated prose that says
// where the data comes from); only the "# generated:" stamp is refreshed.
//
// A rebuild that loses rows, goes backwards in time, or moves a value already
// in the file by more than a rounding is reported and, unless -force, not
// written: the sources publish revisions (the CPI in particular), and a
// revision worth accepting is worth reading first.
//
// Usage: gen-snapshots [-dir path] [-only name,name] [-dry] [-force]
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

func main() {
	dir := flag.String("dir", "pkg/marketdata/data", "directory holding the embedded snapshots")
	only := flag.String("only", "", "comma-separated basenames to refresh (default: all)")
	dry := flag.Bool("dry", false, "build and report, but do not write anything")
	force := flag.Bool("force", false, "write even when a snapshot fails its consistency checks")
	flag.Parse()

	wanted := map[string]bool{}
	for name := range strings.SplitSeq(*only, ",") {
		if name = strings.TrimSpace(name); name != "" {
			wanted[name] = true
		}
	}

	stamp := time.Now().UTC().Format("2006-01-02")
	var failed int
	for _, s := range snapshots {
		if len(wanted) > 0 && !wanted[s.file] {
			continue
		}
		if err := refresh(s, *dir, stamp, *dry, *force); err != nil {
			log.Printf("%s: %v", s.file, err)
			failed++
		}
	}
	if failed > 0 {
		log.Fatalf("%d snapshot(s) not refreshed", failed)
	}
}

// snapshot is one embedded file: how to rebuild its rows, and how much of the
// existing file to carry over unchanged in front of them.
type snapshot struct {
	file  string                         // basename under -dir
	build func(old []row) ([]row, error) // old is the file as it stands, oldest first
}

// row is one "label,value" line: the value stays a string all the way through,
// so a source that already prints the precision we want (FRED) is copied
// verbatim rather than round-tripped through a float.
type row struct {
	label string
	value string
}

// snapshots lists every file this command owns, in the order it refreshes them.
var snapshots = []snapshot{
	{"vix.csv", buildVIX},
	{"cpi-us.csv", buildCPIUS},
	{"eurusd-long.csv", buildEURUSD},
	{"gbpusd-long.csv", buildGBPUSD},
	{"jpyusd-long.csv", buildJPYUSD},
	{"chfusd-long.csv", buildCHFUSD},
	{"hicp-fr.csv", buildHICPFR},
}

// refresh rebuilds one snapshot, reports what moved, and writes it back.
func refresh(s snapshot, dir, stamp string, dry, force bool) error {
	path := filepath.Join(dir, s.file)
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	header, old := split(string(raw))

	fresh, err := s.build(old)
	if err != nil {
		return err
	}
	problems := check(old, fresh)
	log.Printf("%-16s %d rows -> %d, %s..%s, %s", s.file,
		len(old), len(fresh), fresh[0].label, fresh[len(fresh)-1].label, drift(old, fresh))
	for _, p := range problems {
		log.Printf("%-16s   %s", s.file, p)
	}
	switch {
	case dry:
		return nil
	case len(problems) > 0 && !force:
		return fmt.Errorf("%d consistency problem(s); re-run with -force to write anyway", len(problems))
	}
	return write(path, stamp, header, fresh)
}

// --- the seven builders ---

func buildVIX(old []row) ([]row, error) {
	body, err := get("https://cdn.cboe.com/api/global/us_indices/daily_prices/VIX_History.csv")
	if err != nil {
		return nil, err
	}
	var out []row
	for _, line := range strings.Split(body, "\n")[1:] { // DATE,OPEN,HIGH,LOW,CLOSE
		f := strings.Split(strings.TrimSpace(line), ",")
		if len(f) < 5 {
			continue
		}
		d, err := time.Parse("01/02/2006", f[0])
		if err != nil {
			continue
		}
		close, err := strconv.ParseFloat(f[4], 64)
		if err != nil || close <= 0 {
			continue
		}
		out = append(out, row{d.Format("2006-01-02"), strconv.FormatFloat(close, 'f', 2, 64)})
	}
	return sorted(out), nil
}

func buildCPIUS(old []row) ([]row, error) {
	pts, err := fred("CPIAUCNS")
	if err != nil {
		return nil, err
	}
	var out []row
	for _, p := range pts {
		out = append(out, row{p.label[:7], p.value}) // FRED dates a month on its 1st
	}
	return sorted(out), nil
}

// buildEURUSD refetches the euro leg only. Everything before the euro existed
// is a splice of ECU anchors and Frankfurt DM fixings that no longer moves, so
// it is carried over from the file rather than rebuilt from two more sources.
func buildEURUSD(old []row) ([]row, error) {
	pts, err := fred("DEXUSEU")
	if err != nil {
		return nil, err
	}
	var out []row
	for _, p := range pts {
		v, err := strconv.ParseFloat(p.value, 64)
		if err != nil {
			continue
		}
		out = append(out, row{p.label, strconv.FormatFloat(v, 'f', 6, 64)})
	}
	return sorted(append(before(old, "1999-01-01"), out...)), nil
}

func buildGBPUSD(old []row) ([]row, error) { return usdPer("DEXUSUK", false) }
func buildJPYUSD(old []row) ([]row, error) { return usdPer("DEXJPUS", true) }
func buildCHFUSD(old []row) ([]row, error) { return usdPer("DEXSZUS", true) }

// usdPer builds a daily USD-per-foreign-unit cross from a FRED H.10 series,
// inverting the ones FRED quotes as foreign units per dollar. A quoted rate is
// copied verbatim, an inverted one printed to eight significant digits: enough
// that re-inverting returns the published rate.
func usdPer(id string, invert bool) ([]row, error) {
	pts, err := fred(id)
	if err != nil {
		return nil, err
	}
	var out []row
	for _, p := range pts {
		if !invert {
			out = append(out, p)
			continue
		}
		v, err := strconv.ParseFloat(p.value, 64)
		if err != nil || v <= 0 {
			continue
		}
		out = append(out, row{p.label, strconv.FormatFloat(1/v, 'g', 8, 64)})
	}
	return sorted(out), nil
}

// buildHICPFR refetches the Eurostat leg over the carried-over OECD head. The
// head was rescaled to Eurostat's level at the 1996 overlap when it was built,
// so a rebase under it would silently break the chain: the overlap month is
// re-read and compared before anything is written.
func buildHICPFR(old []row) ([]row, error) {
	const overlap = "1996-01"
	fresh, err := eurostatHICP("FR")
	if err != nil {
		return nil, err
	}
	if fresh[0].label > overlap {
		return nil, fmt.Errorf("eurostat starts at %s, after the %s chain point", fresh[0].label, overlap)
	}
	was, ok := value(old, overlap)
	if !ok {
		return nil, fmt.Errorf("the existing file has no %s to chain on", overlap)
	}
	is, _ := value(fresh, overlap)
	if gap := (is - was) / was; gap > 0.001 || gap < -0.001 {
		return nil, fmt.Errorf("eurostat now prints %.4f for %s against the file's %.4f (%.2f %%): the OECD head is chained to the old level, rebuild it before refreshing", is, overlap, was, 100*gap)
	}
	return sorted(append(before(old, overlap), fresh...)), nil
}

// --- sources ---

// fred reads a series from FRED's key-less graph endpoint, skipping the "."
// rows it uses for holidays. Its dates are already YYYY-MM-DD and its values
// already carry the published precision, so both are kept as printed.
func fred(id string) ([]row, error) {
	body, err := get("https://fred.stlouisfed.org/graph/fredgraph.csv?id=" + id)
	if err != nil {
		return nil, fmt.Errorf("FRED %s: %w", id, err)
	}
	var out []row
	for _, line := range strings.Split(body, "\n")[1:] { // observation_date,<id>
		label, value, ok := strings.Cut(strings.TrimSpace(line), ",")
		if !ok || value == "." || value == "" {
			continue
		}
		if _, err := time.Parse("2006-01-02", label); err != nil {
			continue
		}
		if v, err := strconv.ParseFloat(value, 64); err != nil || v <= 0 {
			continue
		}
		out = append(out, row{label, value})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("FRED %s: no usable observation", id)
	}
	return out, nil
}

// eurostatHICP reads the monthly all-items HICP (2015=100) of one geography
// from the Eurostat dissemination API, whose JSON-stat payload indexes values
// by position rather than by date.
func eurostatHICP(geo string) ([]row, error) {
	body, err := get("https://ec.europa.eu/eurostat/api/dissemination/statistics/1.0/data/prc_hicp_midx" +
		"?format=JSON&lang=EN&freq=M&unit=I15&coicop=CP00&geo=" + geo)
	if err != nil {
		return nil, fmt.Errorf("eurostat HICP %s: %w", geo, err)
	}
	var payload struct {
		Value     map[string]float64 `json:"value"`
		Dimension struct {
			Time struct {
				Category struct {
					Index map[string]int `json:"index"`
				} `json:"category"`
			} `json:"time"`
		} `json:"dimension"`
	}
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		return nil, fmt.Errorf("eurostat HICP %s: %w", geo, err)
	}
	var out []row
	for month, i := range payload.Dimension.Time.Category.Index {
		v, ok := payload.Value[strconv.Itoa(i)]
		if !ok || v <= 0 {
			continue // Eurostat leaves a gap for months it has not published
		}
		out = append(out, row{month, strconv.FormatFloat(v, 'f', 4, 64)})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("eurostat HICP %s: no usable month", geo)
	}
	return sorted(out), nil
}

// get fetches a URL over HTTP/1.1 (FRED's HTTP/2 resets Go's client).
func get(url string) (string, error) {
	client := &http.Client{
		Timeout:   2 * time.Minute,
		Transport: &http.Transport{ForceAttemptHTTP2: false},
	}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s: HTTP %d", url, resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	return string(body), err
}

// --- rows ---

// split cuts a snapshot file into its comment header and its rows.
func split(text string) (header []string, rows []row) {
	for line := range strings.SplitSeq(text, "\n") {
		switch {
		case strings.HasPrefix(line, "#"):
			header = append(header, line)
		case strings.TrimSpace(line) == "":
		default:
			if label, value, ok := strings.Cut(strings.TrimSpace(line), ","); ok {
				rows = append(rows, row{label, value})
			}
		}
	}
	return header, rows
}

// sorted orders rows by label (every label is a fixed-width date, so string
// order is date order) and drops any duplicate, keeping the last one seen.
func sorted(rows []row) []row {
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].label < rows[j].label })
	out := rows[:0]
	for i, r := range rows {
		if i+1 < len(rows) && rows[i+1].label == r.label {
			continue
		}
		out = append(out, r)
	}
	return out
}

// before returns the rows strictly older than a label: the settled head a
// builder carries over instead of rebuilding.
func before(rows []row, label string) []row {
	var out []row
	for _, r := range rows {
		if r.label < label {
			out = append(out, r)
		}
	}
	return out
}

// value reads one row's value as a number.
func value(rows []row, label string) (float64, bool) {
	for _, r := range rows {
		if r.label == label {
			v, err := strconv.ParseFloat(r.value, 64)
			return v, err == nil
		}
	}
	return 0, false
}

// check reports everything that would make the rebuild a regression rather
// than a refresh: rows lost, history that no longer reaches as far back, or a
// value moved by more than a rounding of the printed precision. A revision is
// not by itself an error (FRED revises the CPI), but it must be read.
func check(old, fresh []row) []string {
	var out []string
	if len(fresh) < len(old) {
		out = append(out, fmt.Sprintf("lost %d rows", len(old)-len(fresh)))
	}
	if len(old) > 0 && len(fresh) > 0 && fresh[0].label > old[0].label {
		out = append(out, fmt.Sprintf("history now starts at %s instead of %s", fresh[0].label, old[0].label))
	}
	if len(old) > 0 && len(fresh) > 0 && fresh[len(fresh)-1].label < old[len(old)-1].label {
		out = append(out, fmt.Sprintf("history now ends at %s instead of %s", fresh[len(fresh)-1].label, old[len(old)-1].label))
	}
	index := map[string]string{}
	for _, r := range fresh {
		index[r.label] = r.value
	}
	var missing int
	for _, r := range old {
		if _, ok := index[r.label]; !ok {
			missing++
		}
	}
	if missing > 0 {
		out = append(out, fmt.Sprintf("%d dates present before are gone", missing))
	}
	return out
}

// drift summarises how far the rebuild moved the values the file already had:
// the count of revised dates and the worst relative move, which is what a
// reader needs to decide whether a revision is routine or a broken source.
func drift(old, fresh []row) string {
	index := map[string]float64{}
	for _, r := range fresh {
		if v, err := strconv.ParseFloat(r.value, 64); err == nil {
			index[r.label] = v
		}
	}
	var revised int
	var worst float64
	var where string
	for _, r := range old {
		was, err := strconv.ParseFloat(r.value, 64)
		is, ok := index[r.label]
		if err != nil || !ok || was == 0 {
			continue
		}
		gap := (is - was) / was
		if gap < 0 {
			gap = -gap
		}
		if gap < 1e-9 {
			continue
		}
		revised++
		if gap > worst {
			worst, where = gap, r.label
		}
	}
	if revised == 0 {
		return "no value revised"
	}
	return fmt.Sprintf("%d values revised, worst %.3f %% on %s", revised, 100*worst, where)
}

// write puts the header back in front of the rows, refreshing the generation
// stamp (adding one to the files that never carried it).
func write(path, stamp string, header []string, rows []row) error {
	var b strings.Builder
	stamped := false
	for _, line := range header {
		if strings.HasPrefix(line, "# generated:") {
			line, stamped = "# generated: "+stamp, true
		}
		b.WriteString(line + "\n")
	}
	if !stamped {
		b.WriteString("# generated: " + stamp + "\n")
	}
	for _, r := range rows {
		b.WriteString(r.label + "," + r.value + "\n")
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
