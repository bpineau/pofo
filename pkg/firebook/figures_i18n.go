package firebook

import "regexp"

// The English edition's figures.
//
// The ~100 plates stay SINGLE-SOURCE: every generator keeps its French
// literals and is never duplicated, never parameterized by language. The
// English edition post-processes the RENDERED SVG instead, translating its
// text nodes: the dictionary first, then a mechanical number reformat for the
// payloads that carry no words. A payload that is neither is left alone AND
// fails the coverage guard, so a plate cannot quietly ship half-translated.
//
// The consequence is deliberate: renaming a label in a plate an English
// article uses breaks the tests until figureDict is updated in the same
// commit. Figures change rarely and the dictionary lives next to the code, so
// they get the hard guard; prose gets the soft drift report instead.
//
// Captions are not figure text: they live in the article markdown and are
// translated with the prose.

// figureDict translates one rendered text payload. A key is either the bare
// French payload, which applies to every plate, or "<figure-id>|<french>",
// which wins over the bare key for that plate alone (the same French string
// sometimes needs different English in different plates).
//
// Payloads are already SVG-escaped, so a key carrying "&" spells it "&amp;".
// The dictionary grows with the translation campaign: an entry is needed only
// once an English article uses the plate.
var figureDict = map[string]string{
	// vol-drag
	"Le volatility drag : même moyenne, richesses opposées": "Volatility drag: same average, opposite outcomes",
	"régulier : +7 % chaque année":                          "steady: +7% every year",
	"volatil : +27 % / −13 % (même moyenne)":                "volatile: +27% / −13% (same average)",
	"années  →":                      "years  →",
	"richesse (× capital de départ)": "wealth (× starting capital)",

	// sequence-risk
	"Deux séquences, même rendement moyen, retraits identiques": "Two sequences, same average return, identical withdrawals",
	"krach précoce : ruine": "early crash: ruin",
	"krach tardif : survit": "late crash: survives",
	"années de retraite  →": "years of retirement  →",
	"capital de départ":     "starting capital",

	// millesimes-1966-1982
	"MILLÉSIME 1966 CONTRE MILLÉSIME 1982":           "1966 VINTAGE AGAINST 1982 VINTAGE",
	"Le même plan, deux dates de départ":             "The same plan, two starting dates",
	"capital réel restant, en millions d'euros":      "real capital left, in millions of euros",
	"années écoulées depuis le départ à la retraite": "years since retirement began",
	"même plan des deux côtés : 1 M€, 40 k€ retirés chaque année, indexés sur l'inflation, jamais ajustés": "same plan on both sides: EUR 1M, EUR 40k a year, indexed to inflation, never adjusted",
	"sous la mise de départ":     "below the starting stake",
	"départ en 1966":             "starting in 1966",
	"départ en 1982":             "starting in 1982",
	"millésime 1966":             "1966 vintage",
	"millésime 1982":             "1982 vintage",
	"réel, dix premières années": "real, first ten years",
	"réel, trente ans":           "real, thirty years",
	"à l'arrivée":                "at the finish",
	"zéro en 1994, l'année 29":   "zero in 1994, year 29",
	"zéro":                       "zero",
	"4,6 M€":                     "EUR 4.6M",
	"−1,2 %/an":                  "−1.2%/yr",
	"+4,2 %/an":                  "+4.2%/yr",
	"+7,0 %/an":                  "+7.0%/yr",
	"+11,4 %/an":                 "+11.4%/yr",
	"Le portefeuille de 1966 a pourtant rapporté plus que le retrait. C'est l'ordre des années qui a tué le plan.":                  "The 1966 portfolio still earned more than it paid out. The order of the years is what killed the plan.",
	"60/40 américain réel (S&amp;P 500, Treasuries 5 ans, déflatés CPI-U), reconstruction du livre ; retrait fixe, sans fiscalité.": "Real US 60/40 (S&amp;P 500, 5-year Treasuries, CPI-U deflated), reconstructed for this book; fixed withdrawal, no taxes.",

	// Shared by several plates. An age tick reads "40 ans" in French and
	// "age 40" in English, which is the same width.
	"VPW":          "VPW",
	"normal":       "normal",
	"revenu":       "income",
	"40 ans":       "age 40",
	"47 ans":       "age 47",
	"55 ans":       "age 55",
	"57 ans":       "age 57",
	"67 ans":       "age 67",
	"70 ans":       "age 70",
	"77 ans":       "age 77",
	"85 ans":       "age 85",
	"87 ans":       "age 87",
	"95 ans":       "age 95",
	"99 ans":       "age 99",
	"pont":         "bridge",
	"capital (k€)": "capital (EUR k)",

	// vpw-table
	"Le pourcentage monte avec l'âge, le capital fond exprès":    "The percentage rises with age, the capital melts on purpose",
	"le taux de retrait de la table (% du portefeuille)":         "the table's withdrawal rate (% of portfolio)",
	"le rendement supposé, 3,3 % réel : le plancher de la table": "the assumed return, 3.3% real: the table's floor",
	"à 91 ans : 13 %, puis jusqu'à 100 %":                        "at 91: 13%, then up to 100%",
	"vingt-cinq ans de quasi-pourcentage fixe":                   "twenty-five years of near-fixed percentage",
	"le capital, consommé jusqu'à zéro":                          "the capital, spent down to zero",
	"38,5 k€/an, plat":                                           "EUR 38.5k/yr, flat",
	"krach de −30 % à 70 ans : 27,0 k€/an, sans retour":          "a −30% crash at 70: EUR 27.0k/yr, no way back",
	"la vie que cette table produit, à partir de 40 ans avec 1 M€, si le marché sert les rendements supposés": "the life this table produces, from age 40 with EUR 1M, if the market delivers the assumed returns",
	"Annuité inversée à g = 3,3 % réel constant, horizon jusqu'à 100 ans.":                                    "Reversed annuity at a constant g = 3.3% real, horizon out to age 100.",

	// vpw-pont
	"Le pont de pension : vingt ans de revenu qui n'existe pas encore":                  "The pension bridge: twenty years of income that does not exist yet",
	"revenu réel servi (k€/an), ménage de 47 ans, 1,6 M€, pensions de 21,6 k€ à 67 ans": "real income served (EUR k/yr), household aged 47, EUR 1.6M, pensions of EUR 21.6k at 67",
	"le pont : 21,6 k€/an,":                                            "the bridge: EUR 21.6k/yr,",
	"prélevés sur 356 k€ d'obligations":                                "drawn from EUR 356k of bonds",
	"la pension, enfin liquidée":                                       "the pension, finally running",
	"le VPW sur les 1 243 k€ restants : 50,0 k€/an":                    "VPW on the remaining EUR 1,243k: EUR 50.0k/yr",
	"71,6 k€/an":                                                       "EUR 71.6k/yr",
	"67 ans : la pension prend le relais":                              "age 67: the pension takes over",
	"test de tolérance, actions −50 % : 56,6 k€, le pont ne bouge pas": "tolerance test, stocks −50%: EUR 56.6k, the bridge holds",
	"le plancher du ménage, 38 k€":                                     "the household floor, EUR 38k",
	"Aux rendements supposés de la table (3,3 % réel pour la part VPW, 1,9 % pour actualiser le pont), sans fiscalité.": "At the table's assumed returns (3.3% real for the VPW sleeve, 1.9% to discount the bridge), no taxes.",

	// vpw-test-de-perte
	"LE TEST DE TOLÉRANCE À LA PERTE":               "THE LOSS TOLERANCE TEST",
	"Le même choc, avec et sans le pont de pension": "The same shock, with and without the pension bridge",
	"actions −50 %":                                 "stocks −50%",
	"Avec le pont de pension":                       "With the pension bridge",
	"le choc laisse le ménage au-dessus du confort": "the shock leaves the household above comfort",
	"Sans le pont":                                  "Without the bridge",
	"le même choc passe sous le confort":            "the same shock drops below comfort",
	"un couple de 47 ans, 1,6 M€ en 60/40, pensions de 21,6 k€ à 67 ans ; revenu servi, k€ par an": "a couple aged 47, EUR 1.6M in 60/40, pensions of EUR 21.6k at 67; income served, EUR k a year",
	"pointillé bleu : le confort visé, 52 ; pointillé brun : le plancher, 38":                      "blue dashes: target comfort, 52; brown dashes: the floor, 38",
}

var (
	// reFigTspan and reFigText extract the LEAF text payloads of a rendered
	// plate: a <text> wrapping <tspan>s contributes its tspans, not itself
	// (the childless-<text> pattern cannot match it).
	reFigTspan = regexp.MustCompile(`<tspan([^>]*)>([^<]*)</tspan>`)
	reFigText  = regexp.MustCompile(`<text([^>]*)>([^<]*)</text>`)

	// reFrenchDecimal and reFrenchPercent flag French number formatting
	// surviving into English output; the guard tests share them.
	reFrenchDecimal = regexp.MustCompile(`\d,\d`)
	reFrenchPercent = regexp.MustCompile(`\d[\x{00a0}\x{202f} ]%`)

	// reNeutral matches a payload made only of numbers and symbols, with no
	// word to translate: it needs no dictionary entry, only reformatting.
	reNeutral = regexp.MustCompile(`^[\d\s\p{Zs}%×+.,:;()'’/\-−–]+$`)
	reDigit   = regexp.MustCompile(`\d`)

	reThousandsGroup = regexp.MustCompile(`(\d),(\d\d\d)`)

	reDecimalComma = regexp.MustCompile(`(\d),(\d)`)
	reThousands    = regexp.MustCompile(`(\d)[\x{00a0}\x{202f} ](\d\d\d)`)
	rePercentSpace = regexp.MustCompile(`(\d)[\x{00a0}\x{202f} ]?%`)
)

// hasFrenchDecimal reports whether a payload still carries a French decimal
// comma. English thousands separators ("EUR 1,243k") are commas between digits
// too, so they are folded away first; a comma left between digits after that
// is a decimal separator that escaped the translation.
func hasFrenchDecimal(s string) bool {
	for {
		next := reThousandsGroup.ReplaceAllString(s, "$1$2")
		if next == s {
			break
		}
		s = next
	}
	return reFrenchDecimal.MatchString(s)
}

// isNeutralPayload reports whether a text payload carries numbers and symbols
// only, so anglicizeNumbers alone can translate it.
func isNeutralPayload(s string) bool {
	return reNeutral.MatchString(s) && reDigit.MatchString(s)
}

// anglicizeNumbers rewrites French number formatting in one text payload:
// decimal commas become points, (narrow) no-break-space thousands become
// commas, and the space before a percent sign disappears ("6,6 %" -> "6.6%",
// "1 000 000" -> "1,000,000"). Words are left alone; translating them is the
// dictionary's job.
func anglicizeNumbers(s string) string {
	// In French figure text a comma between two digits is always a decimal
	// separator: thousands are grouped with spaces.
	s = reDecimalComma.ReplaceAllString(s, "$1.$2")
	// One pass consumes one separator, so "1 000 000" needs two.
	for {
		next := reThousands.ReplaceAllString(s, "$1,$2")
		if next == s {
			break
		}
		s = next
	}
	return rePercentSpace.ReplaceAllString(s, "$1%")
}

// translatePayload turns one French payload into English: the plate-scoped
// dictionary key first, then the global one, then the mechanical number
// reformat for neutral payloads. Anything else comes back unchanged, which is
// what the coverage guard reports.
func translatePayload(id, payload string) string {
	if en, ok := figureDict[id+"|"+payload]; ok {
		return en
	}
	if en, ok := figureDict[payload]; ok {
		return en
	}
	if isNeutralPayload(payload) {
		return anglicizeNumbers(payload)
	}
	return payload
}

// figureTextNodes lists every text payload of a rendered plate, in document
// order: the tspans first, then the childless <text> elements. The guard tests
// use it to check that a plate is fully covered.
func figureTextNodes(svg string) []string {
	var out []string
	for _, m := range reFigTspan.FindAllStringSubmatch(svg, -1) {
		out = append(out, m[2])
	}
	for _, m := range reFigText.FindAllStringSubmatch(svg, -1) {
		out = append(out, m[2])
	}
	return out
}

// FigureSVGEnglish renders a plate and translates its text nodes to English.
// It is English.Figure; the plate generators themselves stay French and are
// never duplicated. An untranslatable payload is left in French, which the
// coverage guard test turns into a failure.
func FigureSVGEnglish(id string) string {
	svg := translateNodes(id, FigureSVG(id), reFigTspan, "tspan")
	return translateNodes(id, svg, reFigText, "text")
}

// translateNodes rewrites the payload of every <tag> the regexp matches,
// keeping the element's attributes (capture 1) verbatim.
func translateNodes(id, svg string, re *regexp.Regexp, tag string) string {
	return re.ReplaceAllStringFunc(svg, func(node string) string {
		m := re.FindStringSubmatch(node)
		return "<" + tag + m[1] + ">" + translatePayload(id, m[2]) + "</" + tag + ">"
	})
}
