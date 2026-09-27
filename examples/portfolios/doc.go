// Package portfolios embeds the bundled portfolio files so the pofo web
// server (the -serve mode) ships them inside a self-contained binary:
// the hub lists them, /view renders them, /examples/<name>.txt serves the
// raw commented text (the public URL keeps its historical path). The CLI
// itself keeps reading the files from disk
// ("pofo examples/portfolios/foo.txt"); nothing changes for that path.
//
// README.md in this directory presents every file; the Go programs over the
// library live next door, in examples/code.
package portfolios
