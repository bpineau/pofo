package firebook

import (
	"fmt"
	"strings"
)

// The three timelines use the same paired monthly observations. Only the
// selection of months changes between replay and block resampling.
func figBootstrapBlocs() string {
	stock := [...]int{2, -4, 3, 1, -2, 5, 0, 2}
	bond := [...]int{1, 2, 0, -1, 1, -2, 1, 0}
	var b strings.Builder
	b.WriteString(plateHead("Simulation des rendements", "Comment assembler six mois de rendements"))
	b.WriteString(plateDeck("Exemple fictif. Chaque colonne réunit les deux rendements observés le même mois."))

	draw := func(y float64, title string, blocks [][]int) {
		b.WriteString(sTxt(24, y, 14, figInk, "start", "600", title))
		for i, label := range []string{"Mois d'origine", "Actions (%)", "Obligations (%)"} {
			b.WriteString(sTxt(24, y+27+float64(i)*20, 12, figSoft, "start", "400", label))
		}
		x := 135.0
		const width = 54.0
		for _, block := range blocks {
			for _, month := range block {
				ink, fill := figSoft, figWash
				switch month {
				case 2, 3:
					ink, fill = figBlue, "#EBF1FA"
				case 5, 6:
					ink, fill = figDeep, "#FAEEDD"
				}
				fmt.Fprintf(&b, `<rect x="%.0f" y="%.0f" width="53" height="60" fill="%s"/>`, x, y+12, fill)
				cx := x + width/2
				b.WriteString(mTxt(cx, y+27, 13, ink, "middle", "700", fmt.Sprint(month)))
				b.WriteString(mTxt(cx, y+47, 13, figInk, "middle", "400", frMinus(float64(stock[month-1]), 0)))
				b.WriteString(mTxt(cx, y+67, 13, figInk, "middle", "400", frMinus(float64(bond[month-1]), 0)))
				x += width
			}
			b.WriteString(line(x-width*float64(len(block)), y+75, x-1, y+75, figMuted, 1.5))
			x += 12
		}
	}
	draw(89, "Les huit mois disponibles", [][]int{{1, 2, 3, 4, 5, 6, 7, 8}})
	draw(196, "Rejeu historique · une fenêtre de six mois", [][]int{{2, 3, 4, 5, 6, 7}})
	draw(303, "Bootstrap · trois blocs de deux mois", [][]int{{5, 6}, {2, 3}, {5, 6}})
	b.WriteString(plateFoot(400, []string{
		"Les mois 5 et 6 reviennent ensemble. Au raccord, le mois 2 peut suivre le mois 6.",
	}))
	return svg(600, 422, b.String())
}
