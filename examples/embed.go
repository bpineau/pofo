package examples

import (
	"embed"
	"sort"
	"strings"
)

// FS holds the bundled portfolio files (one "<name>.txt" per portfolio).
//
//go:embed *.txt
var FS embed.FS

// Info describes one bundled portfolio file.
type Info struct {
	Name  string // file base name without .txt (URL-safe by construction)
	Title string // first comment line, without the trailing blurb
	Blurb string // the part of the title line after the "--" separator
}

// List returns every bundled portfolio file, sorted by displayed Title
// (case-insensitive). A file without a leading comment line degrades to
// Title = Name, so the fallback still sorts predictably.
func List() []Info {
	entries, err := FS.ReadDir(".")
	if err != nil {
		return nil // cannot happen with a valid embed; keep the API total
	}
	infos := make([]Info, 0, len(entries))
	for _, e := range entries {
		name := strings.TrimSuffix(e.Name(), ".txt")
		title, blurb := titleOf(e.Name())
		if title == "" {
			title = name
		}
		infos = append(infos, Info{Name: name, Title: title, Blurb: blurb})
	}
	sort.Slice(infos, func(i, j int) bool {
		li, lj := strings.ToLower(infos[i].Title), strings.ToLower(infos[j].Title)
		if li != lj {
			return li < lj
		}
		return infos[i].Name < infos[j].Name // stable tie-break on the file id
	})
	return infos
}

// titleOf extracts the title line of a bundled file: it reads the first line
// and parses it with titleFromLine. It returns ("", "") when the file cannot
// be read or when its first line is not a genuine title; callers then fall
// back to Title = Name.
func titleOf(file string) (title, blurb string) {
	raw, err := FS.ReadFile(file)
	if err != nil {
		return "", ""
	}
	line, _, _ := strings.Cut(string(raw), "\n")
	return titleFromLine(line)
}

// titleFromLine parses a file's first raw line into a title and blurb: the
// line stripped of its comment marker, split on the " -- " separator (a few
// older files use an em-dash variant, matched via its escape so the
// character never appears in this source file). It returns ("", "") when the
// line is not a genuine title, i.e. when it is not a leading comment line at
// all, or when it is a "#meta ..." directive rather than prose.
func titleFromLine(line string) (title, blurb string) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "#") {
		return "", ""
	}
	line = strings.TrimSpace(strings.TrimPrefix(line, "#"))
	if line == "meta" || strings.HasPrefix(line, "meta ") {
		return "", ""
	}
	for _, sep := range []string{" -- ", " \u2014 "} {
		if t, b, found := strings.Cut(line, sep); found {
			return strings.TrimSpace(t), strings.TrimSuffix(strings.TrimSpace(b), ".")
		}
	}
	return strings.TrimSuffix(line, "."), ""
}
