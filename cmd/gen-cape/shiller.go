package main

import (
	"fmt"
	"html"
	"math"
	"net/url"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/bpineau/pofo/cmd/internal/xls"
)

func fetchShiller() ([]point, error) {
	page, err := download(shillerURL)
	if err != nil {
		return nil, err
	}
	link, err := workbookLink(shillerURL, page)
	if err != nil {
		return nil, err
	}
	raw, err := download(link)
	if err != nil {
		return nil, err
	}
	cells, err := xls.NamedSheet(raw, "Data")
	if err != nil {
		return nil, err
	}
	return shillerPoints(cells)
}

var href = regexp.MustCompile(`(?i)\bhref\s*=\s*["']([^"']+)["']`)

// The publisher changes the download directory on updates. The same page
// links an unrelated housing workbook, so identify the stock file by name.
func workbookLink(pageURL string, page []byte) (string, error) {
	base, err := url.Parse(pageURL)
	if err != nil {
		return "", err
	}
	for _, match := range href.FindAllSubmatch(page, -1) {
		u, err := base.Parse(html.UnescapeString(string(match[1])))
		if err == nil && (u.Scheme == "https" || u.Scheme == "http") && strings.EqualFold(path.Base(u.Path), "ie_data.xls") {
			return u.String(), nil
		}
	}
	return "", fmt.Errorf("no ie_data.xls download link at %s", pageURL)
}

func shillerPoints(cells []xls.Cell) ([]point, error) {
	// "CAPE" also appears above "Yield" in the excess yield column. The
	// ordinary CAPE is labeled "P/E10 or" above "CAPE". The date shares
	// its header row (the preceding row labels a fractional date column).
	dateHeaders := map[int]int{}
	pe10 := map[[2]int]bool{}
	for _, c := range cells {
		switch strings.TrimSpace(c.Text) {
		case "Date":
			dateHeaders[c.Row] = c.Col
		case "P/E10 or":
			pe10[[2]int{c.Row + 1, c.Col}] = true
		}
	}
	header, dateCol, capeCol := -1, -1, -1
	for _, c := range cells {
		if d, ok := dateHeaders[c.Row]; ok && strings.TrimSpace(c.Text) == "CAPE" && pe10[[2]int{c.Row, c.Col}] {
			if header >= 0 {
				return nil, fmt.Errorf("ambiguous Date/CAPE header")
			}
			header, dateCol, capeCol = c.Row, d, c.Col
		}
	}
	if header < 0 {
		return nil, fmt.Errorf("missing Date and ordinary CAPE on the same Data sheet row")
	}
	dates := map[int]float64{}
	for _, c := range cells {
		if c.Row > header && c.Col == dateCol && !c.IsText {
			dates[c.Row] = c.Num
		}
	}
	var rows []point
	for _, c := range cells {
		if c.Row <= header || c.Col != capeCol || c.IsText {
			continue
		}
		if !(c.Num > 0) || math.IsInf(c.Num, 0) {
			return nil, fmt.Errorf("invalid CAPE %g in worksheet row %d", c.Num, c.Row+1)
		}
		// Shiller writes decimal YYYY.MM, so 2026.1 is October.
		d := dates[c.Row]
		year := int(d)
		fraction := (d - float64(year)) * 100
		month := int(math.Round(fraction))
		if year < 1000 || year > 9999 || month < 1 || month > 12 || math.Abs(fraction-float64(month)) > 1e-6 {
			return nil, fmt.Errorf("invalid YYYY.MM date %g in worksheet row %d", d, c.Row+1)
		}
		rows = append(rows, point{fmt.Sprintf("%04d-%02d-01", year, month), c.Num})
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("no ordinary CAPE observations in Data sheet")
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].date < rows[j].date })
	return rows, nil
}
