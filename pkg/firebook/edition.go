package firebook

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bpineau/pofo/pkg/bookmd"
)

// UIStrings collects every user-visible chrome string of one edition: the
// furniture the package generates around the articles, plus the file-size
// formatter the EPUB download link shows. Article text itself always lives in
// the markdown sources, never here.
type UIStrings struct {
	IndexLink          string           // link back to the table of contents ("Sommaire")
	SameCategory       string           // article footer heading ("Dans la même partie")
	LinkCopied         string           // the heading-anchor toast ("lien copié")
	SectionAnchorLabel string           // aria-label of a heading's "§" anchor
	PartAnchorLabel    string           // aria-label of a part's "§" anchor on the index
	EPUBLink           string           // label of the EPUB download link ("Version epub")
	EPUBUnavailable    string           // 500 body of the epub route
	CatalogUnavailable string           // 500 body of the opds route
	NotFound           string           // body shown when an article file cannot be read
	EditionNote        string           // EPUB title-page sentence; exactly one %s, the HomePath
	HumanSize          func(int) string // byte count for humans ("1,4 Mo", "1.4 MB")
}

// Edition is one language of the book: its manifest, its chrome strings, its
// EPUB identity and its figure renderer. Everything the package renders goes
// through an Edition, and the package-level API (Handler, EPUB, Titles,
// ToHTML) is a thin wrapper over French.
//
// The two editions are two distinct publications: they never share an
// EPUBIdentifier (e-readers use it to recognize a book across refreshes) and
// they never share a slug space (English articles carry English slugs, paired
// to their French source by Article.Source; see Drift).
type Edition struct {
	Lang     string // "fr", "en": html lang, EPUB dc:language, JSON-LD inLanguage
	OGLocale string // "fr_FR", "en_US"

	SiteName        string // the book's title
	SiteDescription string // SEO meta description of the index page
	SiteLede        string // hero sentence, shared by the index page and the EPUB title page

	HomePath string // where the online edition lives ("/firebook/fr/")
	AssetDir string // article root inside the embedded FS ("assets/book/fr")

	EPUBFileName   string // download URL (relative) and Content-Disposition name
	EPUBIdentifier string // fixed urn:uuid, one per edition, never changes

	Categories []Category                // the table of contents, in reading order
	UI         UIStrings                 // the chrome strings
	Callouts   map[string]bookmd.Callout // ::: display labels; nil keeps bookmd's French ones
	Figure     func(id string) string    // "::: figure <id>" payload
}

// French is the source edition of the book. Every other edition is a
// translation of it, and all writing happens here first.
var French = &Edition{
	Lang:     "fr",
	OGLocale: "fr_FR",
	SiteName: "Le FIRE tranquille",
	SiteDescription: "Vivre de son capital sans le survivre : la science du retrait, " +
		"les stratégies et portefeuilles qui résistent, l'inflation, la fiscalité française et le facteur humain.",
	SiteLede: "Vivre de son capital sans le survivre : la science du retrait, " +
		"les modèles et leurs pièges, les stratégies, les portefeuilles qui résistent, les buffers, l'inflation.",
	HomePath:       "/firebook/fr/",
	AssetDir:       "assets/book/fr",
	EPUBFileName:   "le-fire-tranquille.epub",
	EPUBIdentifier: "urn:uuid:6f8a2c1e-4d3b-4a7e-9c21-8f5b0e6d4a92",
	Categories:     Categories,
	UI: UIStrings{
		IndexLink:          "Sommaire",
		SameCategory:       "Dans la même partie",
		LinkCopied:         "lien copié",
		SectionAnchorLabel: "Lien direct vers cette section",
		PartAnchorLabel:    "Lien direct vers cette partie",
		EPUBLink:           "Version epub",
		EPUBUnavailable:    "EPUB indisponible",
		CatalogUnavailable: "Catalogue indisponible",
		NotFound:           "Article introuvable.",
		EditionNote:        "La version en ligne, tenue à jour, est publiée par pofo à %s.",
		HumanSize:          humanSizeFR,
	},
	Callouts: nil, // bookmd's built-in labels are the French ones
	Figure:   FigureSVG,
}

// humanSizeFR formats a byte count as a compact, French-locale file size
// ("312 Ko", "1,4 Mo"), the way the EPUB download link presents it.
func humanSizeFR(n int) string {
	switch {
	case n >= 1<<20:
		return strings.Replace(fmt.Sprintf("%.1f Mo", float64(n)/(1<<20)), ".", ",", 1)
	case n >= 1<<10:
		return fmt.Sprintf("%d Ko", (n+512)/(1<<10))
	default:
		return fmt.Sprintf("%d o", n)
	}
}

// The package-level API is the French edition. It predates the per-language
// Edition value and stays the shape other repositories import.

// Handler serves the French edition; see (*Edition).Handler.
func Handler(opts ...Option) http.Handler { return French.Handler(opts...) }

// EPUB renders the French edition; see (*Edition).EPUB.
func EPUB(modified time.Time) ([]byte, error) { return French.EPUB(modified) }

// Titles maps the French edition's written slugs to their display titles;
// see (*Edition).Titles.
func Titles() map[string]string { return French.Titles() }

// ToHTML renders one French article body; see (*Edition).ToHTML.
func ToHTML(src string, titles map[string]string) string { return French.ToHTML(src, titles) }
