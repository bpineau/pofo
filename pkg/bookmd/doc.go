// Package bookmd renders the small Markdown dialect shared by pofo's embedded
// books (firebook's "Le FIRE tranquille") and locador's embedded
// documentation, so articles are portable between them and the renderer can
// be reused without embedding any book's assets.
//
// The dialect: headings (# is demoted to h2, the page shell owns the h1),
// pipe tables, - and 1. lists (with GitHub task items), > blockquotes,
// --- rules, ::: callout blocks, ::: figure blocks, and inline bold / italic
// / code / [label](url) links / [[slug]] and [[slug|label]] wiki-links.
//
// ToHTML renders one article body. Options tunes three host-specific points
// without forking the renderer:
//
//   - Titles maps the slugs of WRITTEN articles to their display titles and
//     drives [[slug]] wiki-links: a known slug becomes a link, an unknown one
//     degrades to its label as plain text so readers never hit a dead link.
//   - Href maps a wiki-link slug to its target URL; nil keeps the bare slug
//     (the web default). An EPUB export, for example, sets it to slug+".xhtml".
//   - Figure supplies the payload of a "::: figure <id>" block (in firebook an
//     inline SVG generated in Go); nil drops the figure block entirely.
//
// The built-in callout types are in Callouts (encart, cle, astuce, attention,
// exemple, science, terrain, admin); an unknown ::: type degrades to encart.
package bookmd
