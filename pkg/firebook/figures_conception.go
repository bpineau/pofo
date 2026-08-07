package firebook

import (
	"fmt"
	"strings"
)

// The design map of "concevoir-un-portefeuille". The article's thesis is that
// an allocation is the SUM of the answers, so the plate is a bipartite diagram
// and not a chart: the risks the article names at question 1 on the left, the
// bricks it chooses at question 5 on the right, one link per answer it writes
// down. Nothing is measured here. Every node is one of the article's own lines
// and every link carries the fragment of the article it is read from, so
// figures_conception_test.go can pin the whole structure against the embedded
// markdown and fail the day the text and the map diverge.

// cbWeight is the ordered editorial scale of defence intensity, read off the
// article's wording and nothing else: the answer the article leads with for a
// risk is the main defence, another answer it names with a definite article is
// secondary, and an answer it hedges ("un peu de duration", "un tilt value",
// "une devise refuge") is partial. Three widths, no continuous scale, because
// there is no number to draw.
type cbWeight int

const (
	cbPartial cbWeight = iota
	cbSecondary
	cbMain
)

// stroke is the link width each rung of the scale is drawn with.
func (w cbWeight) stroke() float64 {
	switch w {
	case cbMain:
		return 3.4
	case cbSecondary:
		return 2.1
	default:
		return 1.2
	}
}

// cbKind separates the four natures of left-hand entry the article makes.
type cbKind int

const (
	cbDefended cbKind = iota // a market risk the article answers with bricks
	cbAccepted               // a market risk it accepts: no outgoing link
	cbEngine                 // the engine line of the map, a need and not a risk
	cbOutside                // a risk the portfolio does not treat at all
)

// cbNode is one label of either column.
type cbNode struct {
	key   string
	label string
	note  string  // second, muted line
	probe string  // verbatim fragment of the article the node is read from
	kind  cbKind  // left column only
	y     float64 // row centre
}

// cbLink is one answer of the article's map, with the fragment it comes from.
type cbLink struct {
	risk, brick string
	w           cbWeight
	quote       string
}

// Geometry: two columns of horizontal labels, links in the band between them.
const (
	cbRiskLabelX  = 214.0
	cbRiskDotX    = 224.0
	cbLinkX0      = 232.0
	cbLinkX1      = 412.0
	cbBrickDotX   = 420.0
	cbBrickLabelX = 430.0
)

// The left column: the risks of question 1, plus the engine line of question 5.
// Rows sit at the vertical centre of the bricks they answer, so the links stay
// shallow and nothing crosses a label.
var cbRisks = []cbNode{
	{key: "moteur", label: "Le moteur", note: "croissance de long terme", probe: "Le moteur (croissance de long terme", kind: cbEngine, y: 114},
	{key: "sequence", label: "Risque de séquence", probe: "Le risque de séquence", kind: cbDefended, y: 176},
	{key: "deflation", label: "Déflation", probe: "la déflation", kind: cbDefended, y: 204},
	{key: "inflation", label: "Inflation persistante", probe: "l'inflation persistante", kind: cbDefended, y: 274},
	{key: "monetaire", label: "Crise monétaire", probe: "la crise de confiance monétaire", kind: cbDefended, y: 330},
	{key: "baissier", label: "Marché baissier prolongé", note: "accepté : on est payé pour le porter", probe: "le marché baissier prolongé", kind: cbAccepted, y: 376},
}

// The risks the article puts outside the portfolio, in its own order.
var cbOffPortfolio = []cbNode{
	{key: "longevite", label: "longévité", probe: "La longévité", kind: cbOutside},
	{key: "dependance", label: "dépendance", probe: "la dépendance", kind: cbOutside},
	{key: "divorce", label: "divorce", probe: "le divorce", kind: cbOutside},
	{key: "accident", label: "accident personnel", probe: "l'accident personnel", kind: cbOutside},
}

// The right column: the bricks of question 5, in the order that keeps the links
// untangled.
var cbBricks = []cbNode{
	{key: "actions", label: "Actions mondiales", probe: "les actions mondiales", y: 114},
	{key: "matelas", label: "Matelas de liquidités", probe: "un matelas de liquidités", y: 148},
	{key: "flexibilite", label: "Flexibilité des dépenses", probe: "la flexibilité des dépenses", y: 176},
	{key: "duration", label: "Duration d'État de qualité", probe: "duration d'État de qualité", y: 204},
	{key: "linkers", label: "Obligations indexées", probe: "les obligations indexées", y: 232},
	{key: "trend", label: "Suivi de tendance (trend)", probe: "le trend", y: 260},
	{key: "value", label: "Tilt value", probe: "un tilt value", y: 288},
	{key: "or", label: "Or", probe: "l'or", y: 316},
	{key: "devise", label: "Devise refuge (le dollar)", probe: "une devise refuge", y: 344},
}

// The map itself. Each line is one answer the article writes down, with the
// fragment that justifies both the link and its rung on the intensity scale.
var cbLinks = []cbLink{
	{"sequence", "matelas", cbMain, "Séquence : un matelas de liquidités"},
	{"sequence", "flexibilite", cbSecondary, "un matelas de liquidités, la flexibilité des dépenses"},
	{"sequence", "duration", cbPartial, "et un peu de duration d'État"},
	{"inflation", "linkers", cbMain, "Inflation persistante : les obligations indexées"},
	{"inflation", "or", cbSecondary, "l'or ([[or-en-retrait]])"},
	{"inflation", "trend", cbSecondary, "le trend ([[managed-futures]])"},
	{"inflation", "value", cbPartial, "un tilt value"},
	{"deflation", "duration", cbMain, "Déflation ou krach désinflationniste : la duration d'État de qualité"},
	{"monetaire", "or", cbMain, "Crise monétaire : l'or"},
	{"monetaire", "devise", cbPartial, "et une devise refuge"},
	{"moteur", "actions", cbMain, "Le moteur (croissance de long terme, longévité) : les actions mondiales"},
}

// cbCurve draws one link as a flat cubic with horizontal ends, so links leave
// and land parallel to their labels and never run into the text.
func cbCurve(y0, y1, w float64, col string) string {
	dx := (cbLinkX1 - cbLinkX0) * 0.45
	return fmt.Sprintf(`<path d="M %.1f,%.1f C %.1f,%.1f %.1f,%.1f %.1f,%.1f" fill="none" stroke="%s" stroke-width="%.1f" stroke-linecap="round"/>`,
		cbLinkX0, y0, cbLinkX0+dx, y0, cbLinkX1-dx, y1, cbLinkX1, y1, col, w)
}

// figRisquesBriques draws the article's own risk list against its own brick
// list, one link per answer, and lets the two pathologies the text describes be
// seen instead of told: a risk defended four times while another carries no
// link at all, and the empty slot a brick without an incoming link would fill.
func figRisquesBriques() string {
	risk := func(key string) cbNode {
		for _, n := range cbRisks {
			if n.key == key {
				return n
			}
		}
		return cbNode{}
	}
	brick := func(key string) cbNode {
		for _, n := range cbBricks {
			if n.key == key {
				return n
			}
		}
		return cbNode{}
	}

	var b strings.Builder
	b.WriteString(plateHead("concevoir un portefeuille", "Un risque, une réponse : l'allocation tombe à la fin"))
	b.WriteString(sTxt(cbRiskLabelX, 78, 10.5, figMuted, "end", "400", "Vos risques (question 1)"))
	b.WriteString(sTxt(cbBrickLabelX, 78, 10.5, figMuted, "start", "400", "Vos briques (question 5)"))
	b.WriteString(line(20, 88, 620, 88, figGrid, 1))
	// the engine link is the one line of the map that is not a defence
	b.WriteString(sTxt(322, 105, 9.5, figBlue, "middle", "400", "le moteur, pas une défense"))

	// links first, thickest first, so the thin ones stay legible on top
	for _, want := range []cbWeight{cbMain, cbSecondary, cbPartial} {
		for _, l := range cbLinks {
			if l.w != want {
				continue
			}
			col := figAccent
			if l.risk == "moteur" {
				col = figBlue // the engine is not a defence
			}
			b.WriteString(cbCurve(risk(l.risk).y, brick(l.brick).y, l.w.stroke(), col))
		}
	}

	// the left column
	for _, n := range cbRisks {
		switch n.kind {
		case cbEngine:
			fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4" fill="%s"/>`, cbRiskDotX, n.y, figBlue)
		case cbAccepted:
			// no outgoing link: a stub that stops in the empty band says so
			b.WriteString(dashLine(cbLinkX0, n.y, cbLinkX0+62, n.y, figMuted, 1.2, "3 4"))
			fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4" fill="#fffdf9" stroke="%s" stroke-width="1.6"/>`, cbRiskDotX, n.y, figMuted)
		default:
			fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4" fill="%s"/>`, cbRiskDotX, n.y, figAccent)
		}
		col := figInk
		if n.kind == cbAccepted {
			col = figSoft
		}
		b.WriteString(sTxt(cbRiskLabelX, n.y+4, 11.5, col, "end", "600", n.label))
		if n.note != "" {
			b.WriteString(sTxt(cbRiskLabelX, n.y+18, 9.5, figMuted, "end", "400", n.note))
		}
	}

	// the right column
	for _, n := range cbBricks {
		col := figDeep
		if n.key == "actions" {
			col = figBlue
		}
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4" fill="%s"/>`, cbBrickDotX, n.y, col)
		b.WriteString(sTxt(cbBrickLabelX, n.y+4, 11.5, figInk, "start", "600", n.label))
	}

	// the foot: what the map leaves out, and the slot a brick must not sit in
	b.WriteString(line(20, 416, 620, 416, figGrid, 1))
	b.WriteString(sTxt(20, 436, 11, figSoft, "start", "600", "Hors portefeuille"))
	var off []string
	for _, n := range cbOffPortfolio {
		off = append(off, n.label)
	}
	b.WriteString(sTxt(20, 452, 10.5, figMuted, "start", "400", strings.Join(off, ", ")))
	b.WriteString(sTxt(20, 467, 9.5, figMuted, "start", "400", "le reste du plan y répond : rentes, provisions, assurance"))
	fmt.Fprintf(&b, `<rect x="350" y="424" width="272" height="50" rx="6" fill="none" stroke="%s" stroke-width="1.2" stroke-dasharray="4 4"/>`, figRule)
	b.WriteString(sTxt(364, 446, 10.5, figSoft, "start", "600", "Une brique sans lien entrant"))
	b.WriteString(sTxt(364, 462, 9.5, figMuted, "start", "400", "pas de diversification, seulement des frais"))

	// the intensity legend, the only place the editorial scale is stated
	lx := 20.0
	sample := func(w float64, col, dash, lbl string) {
		if dash != "" {
			b.WriteString(dashLine(lx, 496, lx+26, 496, col, w, dash))
		} else {
			b.WriteString(line(lx, 496, lx+26, 496, col, w))
		}
		b.WriteString(sTxt(lx+33, 500, 10, figMuted, "start", "400", lbl))
		lx += 33 + 5.6*float64(len([]rune(lbl))) + 26
	}
	sample(cbMain.stroke(), figAccent, "", "défense principale")
	sample(cbSecondary.stroke(), figAccent, "", "secondaire")
	sample(cbPartial.stroke(), figAccent, "", "partielle")
	sample(1.2, figMuted, "3 4", "risque accepté, aucun lien")
	return svg(640, 516, b.String())
}
