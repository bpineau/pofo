package pofo

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// fromLine is the first line of a README snippet that copies a runnable
// example: "// from analyze.Example_sixtyForty", optionally followed by a
// parenthesized note.
var fromLine = regexp.MustCompile(`^// from ([a-z]+)\.(Example\w*)`)

// TestReadmeSnippets holds README.md's promise that every Go snippet of
// "Using it as a library" is the body of the runnable example its first
// line names, so go test keeps it true: a snippet that drifts from its
// example, or names one that does not exist, fails here. The package
// comment's complete program is held to the root package Example the same
// way.
func TestReadmeSnippets(t *testing.T) {
	readme, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	for _, block := range goBlocks(string(readme)) {
		m := fromLine.FindStringSubmatch(block[0])
		if m == nil {
			continue
		}
		checked++
		body, err := exampleBody(m[1], m[2])
		if err != nil {
			t.Errorf("README snippet %q: %v", block[0], err)
			continue
		}
		if got := strings.Join(block[1:], "\n"); got != body {
			t.Errorf("README snippet %q drifted from its example:\n%s", block[0], diff(body, got))
		}
	}
	if checked < 10 {
		t.Errorf("only %d README snippets name their example; the parser lost them", checked)
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
// indentation removed and its "// Output:" comment dropped, as a README
// snippet copies it. pkg is the package's directory name under pkg/, or
// "pofo" for the root package.
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
				if t := strings.TrimSpace(l); t == "// Output:" || t == "// Unordered output:" {
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
