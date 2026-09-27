package pofo

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// fromLine is the first line of a documentation snippet that copies a
// runnable example: "// from analyze.Example_sixtyForty", optionally
// followed by a parenthesized note.
var fromLine = regexp.MustCompile(`^// from ([a-z]+)\.(Example\w*)`)

// usageDir holds the human guides, and unitsGuide the one whose units table
// repeats the root package documentation's.
const (
	usageDir   = "docs/usage"
	unitsGuide = "docs/usage/library/README.md"
)

// TestDocSnippets holds the documentation's promise that every Go snippet
// of README.md and of the guides under docs/usage is the body of the
// runnable example its first line names, so go test keeps it true: a snippet
// that drifts from its example, names one that does not exist, or names none
// at all fails here. The package comment's complete program is held to the
// root package Example the same way.
func TestDocSnippets(t *testing.T) {
	files, err := markdownFiles()
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	for _, file := range files {
		md, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		for _, block := range goBlocks(string(md)) {
			m := fromLine.FindStringSubmatch(block[0])
			if m == nil {
				t.Errorf("%s: a Go snippet names no example (want a first line \"// from pkg.ExampleX\"): %q", file, block[0])
				continue
			}
			checked++
			body, err := exampleBody(m[1], m[2])
			if err != nil {
				t.Errorf("%s: snippet %q: %v", file, block[0], err)
				continue
			}
			if got := strings.Join(block[1:], "\n"); got != body {
				t.Errorf("%s: snippet %q drifted from its example:\n%s", file, block[0], diff(body, got))
			}
		}
	}
	if checked < 25 {
		t.Errorf("only %d snippets name their example across %d files; the parser lost them", checked, len(files))
	}

	src, err := os.ReadFile("doc.go")
	if err != nil {
		t.Fatal(err)
	}
	program, ok := mainBody(string(src))
	if !ok {
		t.Fatal("doc.go: no complete program (func main) in the package comment")
	}
	body, err := exampleBody("pofo", "Example")
	if err != nil {
		t.Fatal(err)
	}
	if program != body {
		t.Errorf("doc.go's program drifted from the root Example:\n%s", diff(body, program))
	}
}

// TestUnitsTable holds the library guide's units table to the root package
// documentation's, which is authoritative: the same rows in the same order,
// and every name the guide quotes in a row (a backquoted identifier, field
// or directive) also named in doc.go's row. The guide may say it more
// briefly; it may not say something doc.go does not.
func TestUnitsTable(t *testing.T) {
	src, err := os.ReadFile("doc.go")
	if err != nil {
		t.Fatal(err)
	}
	labels, rows := docUnits(string(src))
	if len(labels) < 10 {
		t.Fatalf("doc.go: %d units rows parsed; the parser lost the table", len(labels))
	}
	md, err := os.ReadFile(unitsGuide)
	if err != nil {
		t.Fatal(err)
	}
	gotLabels, cells := guideUnits(string(md))
	if !slices.Equal(gotLabels, labels) {
		t.Fatalf("%s: units rows %q, doc.go has %q", unitsGuide, gotLabels, labels)
	}
	quoted := regexp.MustCompile("`([^`]+)`")
	for _, label := range labels {
		for _, m := range quoted.FindAllStringSubmatch(cells[label], -1) {
			if !strings.Contains(rows[label], m[1]) {
				t.Errorf("%s: units row %q quotes %q, which doc.go's row does not name", unitsGuide, label, m[1])
			}
		}
	}
}

// markdownFiles returns README.md and every Markdown file under usageDir.
func markdownFiles() ([]string, error) {
	files := []string{"README.md"}
	err := filepath.WalkDir(usageDir, func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(path, ".md") {
			files = append(files, path)
		}
		return err
	})
	return files, err
}

// goBlocks returns the lines of every ```go fenced block of a Markdown text.
func goBlocks(md string) [][]string {
	var blocks [][]string
	var cur []string
	in := false
	sc := bufio.NewScanner(strings.NewReader(md))
	for sc.Scan() {
		line := sc.Text()
		switch {
		case !in && line == "```go":
			in, cur = true, nil
		case in && line == "```":
			in = false
			if len(cur) > 0 {
				blocks = append(blocks, cur)
			}
		case in:
			cur = append(cur, line)
		}
	}
	return blocks
}

// exampleBody returns the source of an example function's body, one tab of
// indentation removed and its "// Output:" comment dropped, as a snippet
// copies it. pkg is the package's directory name under pkg/, or "pofo" for
// the root package.
func exampleBody(pkg, name string) (string, error) {
	dir := filepath.Join("pkg", pkg)
	if pkg == "pofo" {
		dir = "."
	}
	files, err := filepath.Glob(filepath.Join(dir, "*_test.go"))
	if err != nil {
		return "", err
	}
	fset := token.NewFileSet()
	for _, file := range files {
		src, err := os.ReadFile(file)
		if err != nil {
			return "", err
		}
		f, err := parser.ParseFile(fset, file, src, parser.ParseComments)
		if err != nil {
			return "", err
		}
		for _, d := range f.Decls {
			fn, ok := d.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || fn.Name.Name != name {
				continue
			}
			tf := fset.File(fn.Pos())
			text := string(src[tf.Offset(fn.Body.Lbrace)+1 : tf.Offset(fn.Body.Rbrace)])
			var lines []string
			for _, l := range strings.Split(strings.Trim(text, "\n"), "\n") {
				if t := strings.TrimSpace(l); t == "// Output:" || t == "// Unordered output:" || strings.HasPrefix(t, "// Output: ") {
					break
				}
				lines = append(lines, strings.TrimPrefix(l, "\t"))
			}
			return strings.TrimRight(strings.Join(lines, "\n"), "\n"), nil
		}
	}
	return "", fmt.Errorf("no func %s in %s/*_test.go", name, dir)
}

// mainBody returns the body of the "func main" of the program a package
// comment shows (tab-indented comment lines), in the layout exampleBody
// returns.
func mainBody(src string) (string, bool) {
	var lines []string
	in := false
	for _, l := range strings.Split(src, "\n") {
		code, isCode := strings.CutPrefix(l, "//\t")
		switch {
		case !in && isCode && code == "func main() {":
			in = true
		case in && isCode && code == "}":
			return strings.Join(lines, "\n"), true
		case in && l == "//":
			lines = append(lines, "")
		case in && isCode:
			lines = append(lines, strings.TrimPrefix(code, "\t"))
		case in:
			return "", false
		}
	}
	return "", false
}

// docUnits parses the units table of the root package comment: the
// preformatted block after "# Units and conventions", a label column then
// a text column, a row running on over lines that open with spaces. It
// returns the labels in order, lower-cased, and each row's text on one line.
func docUnits(src string) ([]string, map[string]string) {
	var labels []string
	rows := map[string]string{}
	width, in := 0, false
	for _, l := range strings.Split(src, "\n") {
		if strings.TrimSpace(l) == "// # Units and conventions" {
			in = true
			continue
		}
		code, isCode := strings.CutPrefix(l, "//\t")
		if !in || !isCode {
			if in && len(labels) > 0 {
				break
			}
			continue
		}
		if width == 0 {
			gap := strings.Index(code, "  ")
			width = gap + len(code[gap:]) - len(strings.TrimLeft(code[gap:], " "))
		}
		if len(code) < width {
			continue
		}
		label, text := strings.TrimSpace(code[:width]), strings.TrimSpace(code[width:])
		if code[0] != ' ' {
			labels = append(labels, strings.ToLower(label))
		} else if last := labels[len(labels)-1]; label != "" {
			labels[len(labels)-1] = last + " " + strings.ToLower(label)
			rows[labels[len(labels)-1]] = rows[last]
			delete(rows, last)
		}
		key := labels[len(labels)-1]
		rows[key] = strings.TrimSpace(rows[key] + " " + text)
	}
	return labels, rows
}

// guideUnits parses the Markdown table under a guide's "## Units" heading:
// the labels of its first column in order, lower-cased, and the rest of each
// row.
func guideUnits(md string) ([]string, map[string]string) {
	var labels []string
	cells := map[string]string{}
	in, header := false, 0
	for _, l := range strings.Split(md, "\n") {
		switch {
		case l == "## Units":
			in = true
		case in && strings.HasPrefix(l, "## "):
			return labels, cells
		case in && strings.HasPrefix(l, "|"):
			if header++; header <= 2 {
				continue // the header row and its separator
			}
			parts := strings.Split(strings.Trim(l, "|"), "|")
			label := strings.ToLower(strings.TrimSpace(parts[0]))
			labels = append(labels, label)
			cells[label] = strings.Join(parts[1:], "|")
		}
	}
	return labels, cells
}

// diff shows the first line where want and got part, with some context.
func diff(want, got string) string {
	w, g := strings.Split(want, "\n"), strings.Split(got, "\n")
	for i := range max(len(w), len(g)) {
		var wl, gl string
		if i < len(w) {
			wl = w[i]
		}
		if i < len(g) {
			gl = g[i]
		}
		if wl != gl {
			return fmt.Sprintf("line %d:\n  example: %q\n  copy:    %q", i+1, wl, gl)
		}
	}
	return "(identical)"
}
