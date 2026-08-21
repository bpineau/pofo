package firebook

import (
	"fmt"
	"math"
	"strings"
)

// The placement grid. The article ends on a seven-holding portfolio and says,
// in prose, which wrapper each holding goes in. Prose hides the two readings
// that matter: across a row, one block can sit in two wrappers at once; down a
// column, one wrapper can end up carrying almost the whole portfolio. The
// plate draws the same seven blocks as squares whose AREA is their weight,
// placed in the column of the wrapper that holds them, with the role each
// block serves in the right margin.
//
// No market data is involved: every weight is quoted by the article's own
// example, and the guard test re-reads them from the prose and checks that
// each edition's grid sums to 100.
//
// The two editions do not share this grid. The English article is an ADAPTED
// one: its example states US account types (taxable, Roth, traditional) and a
// different split of the same seven blocks, so translating the French drawing
// would put a PEA inside a US article. The renderer below is therefore single
// source, and only its DATA comes in two tables, both written in French so the
// figure dictionary translates their labels exactly as it does anywhere else.

// ucitsCell is one block in one wrapper: a weight in percent of the whole
// portfolio, or nothing. A barred cell is one the wrapper cannot legally hold,
// which is a different silence from a cell simply not used.
type ucitsCell struct {
	weight float64
	barred bool
}

// ucitsRow is one block: its name, where it sits, and the enemy it answers in
// the table of defenses.
type ucitsRow struct {
	brick string
	cells [3]ucitsCell
	role  string
}

// ucitsColumn is one wrapper: its name and the colour that carries it down the
// plate, so the column reads without a legend.
type ucitsColumn struct {
	name string
	fill string
}

// ucitsLayout is one edition's grid. The rows are the same seven blocks in
// both editions, because the book's blocks do not change with the tax code;
// the columns, the cells and the two notes do.
type ucitsLayout struct {
	columns  [3]ucitsColumn
	rows     []ucitsRow
	subtitle string
	note     string
	footnote string
}

// The seven roles, in the vocabulary of the table of defenses.
const (
	ucitsRoleGrowth    = "la croissance longue"
	ucitsRoleValue     = "la prime de valeur"
	ucitsRoleCrash     = "le krach et la déflation"
	ucitsRoleInflation = "l'inflation persistante"
	ucitsRoleTrust     = "la crise de confiance"
	ucitsRoleBear      = "le régime baissier long"
	ucitsRoleCash      = "la liquidité des retraits"
)

// The seven block names, shared by both grids.
const (
	ucitsBrickWorld  = "moteur monde"
	ucitsBrickValue  = "tilt SCV"
	ucitsBrickBond   = "cœur obligataire"
	ucitsBrickLinker = "linkers courts"
	ucitsBrickGold   = "or"
	ucitsBrickTrend  = "trend"
	ucitsBrickCash   = "buffer et tranche courte"
)

// ucitsHold and the two empty-cell values build a grid that reads as a table.
func ucitsHold(w float64) ucitsCell { return ucitsCell{weight: w} }

var (
	ucitsNone   = ucitsCell{}
	ucitsBarred = ucitsCell{barred: true}
)

// ucitsFR is the French grid: the article's example, 1,6 M€ across a PEA of
// 300 k€ (19 %), a life-insurance contract of 130 k€ (8 %) and a taxable
// account of 1 170 k€ (73 %). The PEA column is barred everywhere but the
// equities, which is the blind spot the plate exists to make obvious.
var ucitsFR = ucitsLayout{
	columns: [3]ucitsColumn{
		{"PEA", figGreen}, {"assurance-vie", figBlue}, {"CTO", figAccent},
	},
	rows: []ucitsRow{
		{ucitsBrickWorld, [3]ucitsCell{ucitsHold(19), ucitsNone, ucitsHold(38)}, ucitsRoleGrowth},
		{ucitsBrickValue, [3]ucitsCell{ucitsNone, ucitsNone, ucitsHold(8)}, ucitsRoleValue},
		{ucitsBrickBond, [3]ucitsCell{ucitsBarred, ucitsNone, ucitsHold(11)}, ucitsRoleCrash},
		{ucitsBrickLinker, [3]ucitsCell{ucitsBarred, ucitsNone, ucitsHold(6)}, ucitsRoleInflation},
		{ucitsBrickGold, [3]ucitsCell{ucitsBarred, ucitsNone, ucitsHold(5)}, ucitsRoleTrust},
		{ucitsBrickTrend, [3]ucitsCell{ucitsBarred, ucitsNone, ucitsHold(5)}, ucitsRoleBear},
		{ucitsBrickCash, [3]ucitsCell{ucitsBarred, ucitsHold(8), ucitsNone}, ucitsRoleCash},
	},
	subtitle: "La cible de Karim et Léa : 1,6 M€, sept lignes, trois enveloppes, environ 0,22 %/an tout compris",
	note:     "Le point discret marque une case où la brique n'y loge pas : le PEA ne peut détenir ni obligations, ni linkers, ni or, ni trend, ni fonds euros.",
	footnote: "Poids en % du portefeuille total. Le moteur monde se coupe en deux, le Monde synthétique du PEA et l'All-World physique du CTO.",
}

// ucitsEN is the English grid: the US article's own example, the same seven
// blocks across a taxable brokerage (50 %), a Roth (13 %) and the traditional
// 401(k) and IRA (37 %). No cell is barred, because a US account may hold
// anything: what places a block there is the tax its income pays, not
// eligibility, so the empty cells are choices and the plate says so.
var ucitsEN = ucitsLayout{
	columns: [3]ucitsColumn{
		{"compte imposable", figAccent}, {"Roth", figGreen}, {"401(k) et IRA", figBlue},
	},
	rows: []ucitsRow{
		{ucitsBrickWorld, [3]ucitsCell{ucitsHold(34), ucitsHold(13), ucitsHold(10)}, ucitsRoleGrowth},
		{ucitsBrickValue, [3]ucitsCell{ucitsHold(8), ucitsNone, ucitsNone}, ucitsRoleValue},
		{ucitsBrickBond, [3]ucitsCell{ucitsNone, ucitsNone, ucitsHold(11)}, ucitsRoleCrash},
		{ucitsBrickLinker, [3]ucitsCell{ucitsNone, ucitsNone, ucitsHold(6)}, ucitsRoleInflation},
		{ucitsBrickGold, [3]ucitsCell{ucitsNone, ucitsNone, ucitsHold(5)}, ucitsRoleTrust},
		{ucitsBrickTrend, [3]ucitsCell{ucitsNone, ucitsNone, ucitsHold(5)}, ucitsRoleBear},
		{ucitsBrickCash, [3]ucitsCell{ucitsHold(8), ucitsNone, ucitsNone}, ucitsRoleCash},
	},
	subtitle: "La cible de Karim et Léa : 1,6 M€, sept lignes, trois comptes, environ 0,12 %/an tout compris",
	note:     "Aucune case n'est interdite ici, et les cases vides sont un choix d'imposition : chaque brique à revenu ordinaire se loge là où son revenu est invisible.",
	footnote: "Poids en % du portefeuille total. Le moteur monde se coupe en trois, une part par compte, selon l'impôt que cette part paie.",
}

// ucitsTotal is what one wrapper ends up carrying, in percent of the whole
// portfolio: the column total the plate prints under the grid.
func (l ucitsLayout) ucitsTotal(col int) float64 {
	sum := 0.0
	for _, r := range l.rows {
		sum += r.cells[col].weight
	}
	return sum
}

// ucitsSum is the whole grid, which must be exactly 100.
func (l ucitsLayout) ucitsSum() float64 {
	return l.ucitsTotal(0) + l.ucitsTotal(1) + l.ucitsTotal(2)
}

// ucitsSide is the side of the square that carries a weight: proportional to
// its square root, so the AREA is the weight. The scale is fixed for both
// grids (6 px per root of a percent), so the two editions draw one block at
// one size.
func ucitsSide(w float64) float64 { return 6 * math.Sqrt(w) }

// ucitsPlate draws one edition's grid.
func ucitsPlate(l ucitsLayout) string {
	const (
		colX0, colPitch = 216.0, 90.0
		valueDX         = 70.0 // the values right-align here, one clean column
		roleX           = 616.0
		rowY0, rowPitch = 134.0, 48.0
		headY           = 100.0
	)
	colX := func(i int) float64 { return colX0 + colPitch*float64(i) }
	rowY := func(i int) float64 { return rowY0 + rowPitch*float64(i) }

	var b strings.Builder
	b.WriteString(plateHead("la liste de courses",
		"La grille d'implantation : quelle brique, dans quelle enveloppe"))
	b.WriteString(sTxt(24, 62, 10.5, figMuted, "start", "400", l.subtitle))

	// The header row: the block column, the three wrappers in their own
	// colours, and the margin that names what each block defends.
	b.WriteString(sTxt(24, headY, 10, figMuted, "start", "600", "la brique"))
	for i, c := range l.columns {
		b.WriteString(sTxt(colX(i), headY, 11, c.fill, "start", "600", c.name))
	}
	b.WriteString(sTxt(roleX, headY, 10, figMuted, "end", "600", "le rôle servi"))
	b.WriteString(line(24, headY+10, roleX, headY+10, figRule, 1))

	// One row per block: the name, the squares, the role.
	for i, r := range l.rows {
		y := rowY(i)
		b.WriteString(sTxt(24, y+4, 10.5, figInk, "start", "600", r.brick))
		for j, cell := range r.cells {
			switch {
			case cell.weight > 0:
				s := ucitsSide(cell.weight)
				fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="2" fill="%s"/>`,
					colX(j), y-s/2, s, s, l.columns[j].fill)
				b.WriteString(mTxt(colX(j)+valueDX, y+4, 10.5, figSoft, "end", "600",
					fmt.Sprintf("%.0f %%", cell.weight)))
			case cell.barred:
				fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="2.2" fill="%s"/>`,
					colX(j)+7, y, figRule)
			}
		}
		b.WriteString(sTxt(roleX, y+4, 10, figMuted, "end", "400", r.role))
		b.WriteString(line(24, y+24, roleX, y+24, figGrid, 1))
	}

	// What each wrapper ends up carrying, which is the column reading.
	yTot := rowY(len(l.rows)-1) + 48
	b.WriteString(sTxt(24, yTot, 10, figMuted, "start", "600", "total logé"))
	for i := range l.columns {
		b.WriteString(mTxt(colX(i)+valueDX, yTot, 13, l.columns[i].fill, "end", "600",
			fmt.Sprintf("%.0f %%", l.ucitsTotal(i))))
	}

	b.WriteString(sTxt(24, yTot+28, 9.5, figMuted, "start", "400", l.note))
	b.WriteString(sTxt(24, yTot+44, 9.5, figMuted, "start", "400", l.footnote))
	return svg(640, int(yTot)+60, b.String())
}

// figUcitsImplantation is the French grid, and figUcitsImplantationEN the
// English one; both are the same drawing over a different table.
func figUcitsImplantation() string   { return ucitsPlate(ucitsFR) }
func figUcitsImplantationEN() string { return ucitsPlate(ucitsEN) }
