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
//
// The dictionary itself is one long data file, figures_i18n_data.go: entries
// are appended there, one plate at a time, and none of it needs reading to add
// one. This file is the mechanism, and it is meant to be read whole.

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
// comma. English thousands separators ("$1,243k") are commas between digits
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
	src := FigureSVG(id)
	if adapted, ok := figuresAdapted[id]; ok {
		src = adapted()
	}
	svg := translateNodes(id, src, reFigTspan, "tspan")
	return translateNodes(id, svg, reFigText, "text")
}

// figuresAdapted holds the rare plate whose DATA the English edition adapts
// rather than translates. A handful of English articles are adaptations, not
// translations: their worked example states US accounts and its own split of
// the same blocks, so rendering the French table there would put a French
// wrapper inside a US article. Such a plate keeps ONE renderer and one layout
// algorithm, registered here under the same id with the other table; both
// tables are written in French, so the dictionary pass below translates their
// labels exactly as it does for every other plate. This is not a second
// drawing of a plate, which the book never ships.
var figuresAdapted = map[string]func() string{
	"ucits-implantation": figUcitsImplantationEN,
}

// translateNodes rewrites the payload of every <tag> the regexp matches,
// keeping the element's attributes (capture 1) verbatim.
func translateNodes(id, svg string, re *regexp.Regexp, tag string) string {
	return re.ReplaceAllStringFunc(svg, func(node string) string {
		m := re.FindStringSubmatch(node)
		return "<" + tag + m[1] + ">" + translatePayload(id, m[2]) + "</" + tag + ">"
	})
}
