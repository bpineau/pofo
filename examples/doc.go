// Package examples embeds the bundled portfolio files so the pofo web
// server (the -serve mode) ships them inside a self-contained binary:
// the hub lists them, /view renders them, /examples/<name>.txt serves the
// raw commented text. The CLI itself keeps reading the files from disk
// ("pofo examples/foo.txt"); nothing changes for that path.
package examples
