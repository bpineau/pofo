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

// List returns every bundled portfolio file, sorted by Name. A file
// without a leading comment line degrades to Title = Name.
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
	sort.Slice(infos, func(i, j int) bool { return infos[i].Name < infos[j].Name })
	return infos
}

// titleOf extracts the title line: the first line stripped of its comment
// marker, split on the " -- " separator (a few older files use an em-dash
// variant, matched via its escape so the character never appears in this
// source file). It returns ("", "") when the first line is not a genuine
// title, i.e. when the file has no leading comment line at all, or when
// that comment line is a "#meta ..." directive rather than prose; callers
// then fall back to Title = Name.
func titleOf(file string) (title, blurb string) {
	raw, err := FS.ReadFile(file)
	if err != nil {
		return "", ""
	}
	line, _, _ := strings.Cut(string(raw), "\n")
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
