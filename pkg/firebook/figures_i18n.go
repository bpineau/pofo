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

	// millesimes-soutenables
	"SOIXANTE-SIX DÉPARTS EN RETRAITE":                                                                     "SIXTY-SIX RETIREMENT DATES",
	"Le taux qui a tenu trente ans, millésime par millésime":                                               "The rate that held for thirty years, vintage by vintage",
	"Retrait initial maximal qu'un 50/50 américain aurait soutenu trente ans, en pouvoir d'achat constant": "Highest initial withdrawal a US 50/50 would have carried for thirty years, in constant purchasing power",
	"66 millésimes, de 1926 à 1991, le dernier dont les trente années sont complètes":                      "66 vintages, 1926 to 1991, the last one whose thirty years are complete",
	"la règle":      "the rule",
	"médiane":       "median",
	"1982 · 10,2 %": "1982 · 10.2%",
	"2,2 points":    "2.2 points",
	"sous la règle : six départs, de 1964 à 1969": "below the rule: six retirement dates, 1964 to 1969",
	"1966 · 3,67 %": "1966 · 3.67%",
	"Le millésime médian aurait supporté 60 % de dépenses en plus que celui de 1966.":                                                               "The median vintage would have carried 60% more spending than the 1966 one.",
	"6 millésimes sur 66 passent sous 4 %, et ce sont six départs consécutifs, de 1964 à 1969. Les 60 autres tiennent la règle.":                    "6 vintages out of 66 fall below 4%, and they are six retirement dates in a row, 1964 to 1969. The other 60 hold the rule.",
	"Panel Jorda-Schularick-Taylor, États-Unis : actions et obligations d'État domestiques, rendements réels annuels déflatés du CPI américain.":    "Jorda-Schularick-Taylor panel, United States: domestic stocks and government bonds, annual real returns deflated by US CPI.",
	"50/50 rééquilibré chaque année, prélèvement en début d'année, ni frais ni impôt ; le capital a le droit de finir exactement à zéro.":           "50/50 rebalanced every year, withdrawal at the start of the year, no fees and no taxes; the capital is allowed to end at exactly zero.",
	"Bengen publie 4,15 % pour 1966. Sa reconstruction n'est pas celle-ci : obligations à moyen terme, autre indice d'actions, données mensuelles.": "Bengen publishes 4.15% for 1966. His reconstruction is not this one: intermediate bonds, another stock index, monthly data.",
	"L'ordre de grandeur tient, la décimale non. Le pire millésime, lui, est le même des deux côtés.":                                               "The ballpark holds, the decimal does not. The worst vintage is the same on both sides.",

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

	// Shared by the two plates of how-much-you-need.
	"COMBIEN IL VOUS FAUT": "HOW MUCH YOU NEED",

	// cible-convexite
	"Chaque demi-point de prudence coûte plus cher que le précédent":                                          "Every extra half point of caution costs more than the one before",
	"le multiple est l'inverse exact du taux : les crochets donnent le prix de chaque demi-point de prudence": "the multiple is the exact inverse of the rate: the brackets price each extra half point of caution",
	"capital cible, en multiples de dépenses annuelles":                                                       "target capital, in multiples of annual spending",
	"+ 3,6x": "+ 3.6x",
	"+ 4,8x": "+ 4.8x",
	"+ 6,7x": "+ 6.7x",
	"25x":    "25x",
	"28,6x":  "28.6x",
	"33x":    "33x",
	"40x":    "40x",
	"au-delà de 33x, le risque dominant n'est plus la ruine :": "past 33x, the dominant risk is no longer ruin:",
	"c'est d'avoir travaillé des années de trop":               "it is having worked years too many",
	"taux de retrait initial":                                  "initial withdrawal rate",
	"plus prudent, plus cher →":                                "more cautious, more expensive →",
	"De 25x à 33x, l'écart représente typiquement 3 à 6 ans de travail de plus pour un taux d'épargne de 40 à 50 %.":   "From 25x to 33x, the gap is typically 3 to 6 more years of work at a savings rate of 40 to 50%.",
	"Courbe exacte, sans hypothèse de marché : c'est la définition du taux de retrait, pas un résultat de simulation.": "Exact curve, no market assumption: this is the definition of the withdrawal rate, not a simulation result.",

	// cible-cascade. The friction step names the French levies in the plate;
	// the English edition states what it is instead, as the article does.
	"Du relevé bancaire au capital cible, marche par marche":                              "From bank statement to capital target, step by step",
	"Nadia et Marc, étape 5 : à 3,5 %, un euro de dépense par mois pèse 343 € de capital": "Nadia and Marc, step 5: at 3.5%, one euro of monthly spending weighs EUR 343 of capital",
	"capital cible (k€)": "target capital (EUR k)",
	"étape 1":            "step 1",
	"étape 2":            "step 2",
	"étape 4":            "step 4",
	"dépenses observées": "observed spending",
	"24 mois de relevés": "24 months of statements",
	"3 400 €/mois":       "EUR 3,400/mo",
	"voyages et loisirs": "travel and leisure",
	"la vie visée":       "the life you aim at",
	"+ 350 €/mois":       "+ EUR 350/mo",
	"mutuelle santé":     "health coverage",
	"à votre charge":     "on your own tab",
	"+ 220 €/mois":       "+ EUR 220/mo",
	"friction fiscale":   "tax friction",
	"impôts et PUMa":     "tax on withdrawals",
	"+ 12 % du brut":     "+ 12% of gross",
	"cible":              "target",
	"1 547 000 €":        "EUR 1,547,000",
	"− 200 000 € : la retraite légale, comptée en revenu différé":                                                     "− EUR 200,000: the state pension, counted as deferred income",
	"sans elle, le même plan aurait exigé 1 747 000 €":                                                                "without it, the same plan would have demanded EUR 1,747,000",
	"Les deux marches que le calcul de comptoir oublie, la fiscalité et la pension, pèsent chacune plus lourd que le": "The two steps the barstool calculation leaves out, tax and the pension, each weigh more than the",
	"budget voyages, et en sens opposés. Les hypothèses sont celles de l'étape 5 : remplacez chacune par la vôtre.":   "travel budget, and they pull in opposite directions. The assumptions are step 5's: replace each with your own.",

	// flux-relatif-phases
	"LES TROIS PHASES": "THE THREE PHASES",
	"ACCUMULATION":     "ACCUMULATION",
	"TRANSITION":       "TRANSITION",
	"RETRAIT":          "WITHDRAWAL",
	"Le flux de l'année, rapporté au portefeuille où il tombe":                                                            "The year's flow, measured against the portfolio it lands in",
	"Une seule variable gouverne les trois phases : ce que vous versez ou prélevez dans l'année, en % du portefeuille.":   "One variable governs the three phases: what you pay in or draw out over the year, as a % of the portfolio.",
	"Projection déterministe sur les hypothèses de l'article, pas une simulation : 34 % d'épargne, 5 % de rendement réel": "A deterministic projection on the article's assumptions, not a simulation: 34% saved, 5% real return",
	"jusqu'au départ à 45 ans, puis un portefeuille dé-risqué à 4 % réel. Ni krach, ni bonne surprise.":                   "until the exit at 45, then a de-risked portfolio at 4% real. No crash, no happy surprise.",
	"axe coupé à +15 % : le ratio vaut 49 % à 22 ans, et diverge au premier euro versé":                                   "axis cut at +15%: the ratio is 49% at age 22, and diverges on the first euro paid in",
	"vous versez":   "you pay in",
	"vous prélevez": "you draw",
	"à 30 ans, le versement de l'année vaut encore 8 % du capital :": "at 30, the year's contribution is still 8% of the capital:",
	"un krach est effacé par les versements suivants.":               "the next few contributions erase a crash.",
	"À la transition, le portefeuille n'a jamais été":                "At the transition the portfolio has never been",
	"aussi gros par rapport au flux qui pourrait le réparer.":        "this large against the flow that could repair it.",
	"45 ans : le flux change de signe,":                              "45: the flow changes sign,",
	"de +2,1 % à −4,1 % du portefeuille, du jour au lendemain.":      "from +2.1% to −4.1% of the portfolio, overnight.",
	"à 85 ans :": "at 85:",
	"âge, de la première épargne sérieuse au grand âge": "age, from the first serious saving year to old age",
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
