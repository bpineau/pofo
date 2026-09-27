package pofo

import (
	"errors"
	"fmt"
	"go/ast"
	"go/build"
	"go/doc"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// TestGodoc holds the library's documentation bar: every package under pkg/
// has a package comment, and every exported identifier a consumer can reach
// (function, type, method, constant, variable, struct field, interface
// method) has a doc comment. The comment of a function, a method or a type
// starts with its name, optionally after "A", "An" or "The", so go doc and
// pkg.go.dev read as sentences. A constant or variable in a documented group
// is covered by the group's comment, and a struct field or a grouped constant
// by a trailing line comment as well.
//
// It parses the sources with go/doc, standard library only, under the
// build constraints of the running platform, the same files go doc reads.
func TestGodoc(t *testing.T) {
	var problems []string
	err := filepath.WalkDir("pkg", func(dir string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		if name := d.Name(); name == "testdata" || strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") {
			return filepath.SkipDir
		}
		found, err := undocumented(dir)
		problems = append(problems, found...)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) > 0 {
		t.Errorf("%d exported identifiers break the godoc bar (a doc comment, starting with the name):\n\t%s",
			len(problems), strings.Join(problems, "\n\t"))
	}
}

// undocumented returns one line per documentation defect of the package in
// dir, nil when dir holds no Go package.
func undocumented(dir string) ([]string, error) {
	bp, err := build.ImportDir(dir, 0)
	if err != nil {
		var noGo *build.NoGoError
		if errors.As(err, &noGo) {
			return nil, nil
		}
		return nil, err
	}
	fset := token.NewFileSet()
	var files []*ast.File
	for _, name := range bp.GoFiles {
		f, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, parser.ParseComments)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	p, err := doc.NewFromFiles(fset, files, "github.com/bpineau/pofo/"+filepath.ToSlash(dir))
	if err != nil {
		return nil, err
	}

	var out []string
	report := func(pos token.Pos, format string, args ...any) {
		at := fset.Position(pos)
		out = append(out, fmt.Sprintf("%s:%d: %s", at.Filename, at.Line, fmt.Sprintf(format, args...)))
	}
	if strings.TrimSpace(p.Doc) == "" {
		out = append(out, dir+": no package comment")
	}
	named := func(pos token.Pos, kind, name, text string) {
		_, bare, _ := strings.Cut(name, ".") // a method's comment starts with the method's own name
		if bare == "" {
			bare = name
		}
		switch {
		case strings.TrimSpace(text) == "":
			report(pos, "%s %s has no doc comment", kind, name)
		case !startsWith(text, bare):
			report(pos, "%s %s: doc comment does not start with its name", kind, name)
		}
	}
	for _, v := range slices.Concat(p.Consts, p.Vars) {
		values(v, report)
	}
	for _, f := range p.Funcs {
		named(f.Decl.Pos(), "func", f.Name, f.Doc)
	}
	for _, typ := range p.Types {
		named(typ.Decl.Pos(), "type", typ.Name, typ.Doc)
		for _, spec := range typ.Decl.Specs {
			members(fset, spec.(*ast.TypeSpec), report)
		}
		for _, v := range slices.Concat(typ.Consts, typ.Vars) {
			values(v, report)
		}
		for _, f := range typ.Funcs {
			named(f.Decl.Pos(), "func", f.Name, f.Doc)
		}
		for _, m := range typ.Methods {
			named(m.Decl.Pos(), "method", typ.Name+"."+m.Name, m.Doc)
		}
	}
	return out, nil
}

// values reports the exported constants and variables of one declaration
// that neither the group's comment nor their own documents.
func values(v *doc.Value, report func(token.Pos, string, ...any)) {
	if strings.TrimSpace(v.Doc) != "" {
		return
	}
	for _, spec := range v.Decl.Specs {
		vs := spec.(*ast.ValueSpec)
		if vs.Doc != nil || vs.Comment != nil {
			continue
		}
		for _, n := range vs.Names {
			if n.IsExported() {
				report(n.Pos(), "%s %s has no doc comment", v.Decl.Tok, n.Name)
			}
		}
	}
}

// members reports the exported struct fields and interface methods of a type
// that carry neither a doc comment nor a line comment. An embedded field is
// documented by its own type, and a field on the line right after a
// documented one is covered when that comment names it ("FlowDates and
// FlowAmounts record...").
func members(fset *token.FileSet, ts *ast.TypeSpec, report func(token.Pos, string, ...any)) {
	var list *ast.FieldList
	switch typ := ts.Type.(type) {
	case *ast.StructType:
		list = typ.Fields
	case *ast.InterfaceType:
		list = typ.Methods
	default:
		return
	}
	line := func(p token.Pos) int { return fset.Position(p).Line }
	var run string // the comment of the run of adjacent fields f belongs to
	for i, f := range list.List {
		switch {
		case f.Doc != nil:
			run = f.Doc.Text()
		case i == 0 || line(f.Pos()) != line(list.List[i-1].End())+1:
			run = ""
		}
		if f.Doc != nil || f.Comment != nil {
			continue
		}
		for _, n := range f.Names {
			if n.IsExported() && !mentions(run, n.Name) {
				report(n.Pos(), "%s.%s has no doc comment", ts.Name.Name, n.Name)
			}
		}
	}
}

// mentions reports whether text names the identifier as a whole word.
func mentions(text, name string) bool {
	for rest := text; ; {
		i := strings.Index(rest, name)
		if i < 0 {
			return false
		}
		before, after := i == 0 || !isIdentChar(rest[i-1]), i+len(name) == len(rest) || !isIdentChar(rest[i+len(name)])
		if before && after {
			return true
		}
		rest = rest[i+len(name):]
	}
}

// startsWith reports whether a doc comment opens with name, the Go
// convention, allowing a leading article and the "Deprecated:" notice.
func startsWith(text, name string) bool {
	text = strings.TrimSpace(text)
	for _, article := range []string{"A ", "An ", "The "} {
		text = strings.TrimPrefix(text, article)
	}
	rest, ok := strings.CutPrefix(text, name)
	if !ok {
		return false
	}
	return rest == "" || !isIdentChar(rest[0])
}

// isIdentChar reports whether b continues an ASCII identifier, so "Fetch"
// does not pass for "FetchExtended".
func isIdentChar(b byte) bool {
	return b == '_' || '0' <= b && b <= '9' || 'a' <= b && b <= 'z' || 'A' <= b && b <= 'Z'
}
