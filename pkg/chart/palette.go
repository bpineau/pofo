package chart

import "math"

// Series colors are picked as a SET, not as a prefix.
//
// defaultPalette lists eight hues in a deliberate slot order, and taking the
// first n of them was fine for two or three series. It stops being fine at
// four: slots 1 and 3 (rust and ochre) are the closest pair of the whole
// palette, and a reader with a common color vision deficiency sees one color
// where the chart drew two. paletteSlots fixes that by choosing WHICH n slots
// to use, so that the least distinguishable pair of the chosen set is as far
// apart as possible.
//
// The choice is anchored on slot 0 (the petrol accent always leads, so a
// chart keeps the product's identity and the first series keeps its color
// whatever the count) and returned in slot order, so colors stay stable for a
// given n and only shift when the number of series changes.

// cvdMatrix is a linear-RGB transform that simulates how a color vision
// deficiency compresses the spectrum (Machado, Oliveira & Fernandes 2009, at
// full severity). Two series that map to the same simulated color are
// indistinguishable for that reader, however far apart they look to us.
type cvdMatrix [3][3]float64

var (
	deuteranopia = cvdMatrix{
		{0.367322, 0.860646, -0.227968},
		{0.280085, 0.672501, 0.047413},
		{-0.011820, 0.042940, 0.968881},
	}
	protanopia = cvdMatrix{
		{0.152286, 1.052583, -0.204868},
		{0.114503, 0.786281, 0.099216},
		{-0.003882, -0.048116, 1.051998},
	}
	tritanopia = cvdMatrix{
		{1.255528, -0.076749, -0.178779},
		{-0.078411, 0.930809, 0.147602},
		{0.004733, 0.691367, 0.303900},
	}
)

// paletteSets holds the chosen slot indices per series count, computed once
// at init (127 candidate subsets in total, so the cost is negligible and the
// result is a compile-time-stable table in practice).
var paletteSets = buildPaletteSets()

// paletteSlots returns the palette slot indices to use for n series, in slot
// order. Up to two series it is the plain prefix (already well separated, and
// the pairing every other chart in the product uses); beyond the palette size
// it cycles, and the caller must disambiguate the twins with labels.
func paletteSlots(n int) []int {
	if n <= 0 {
		return nil
	}
	if n <= 2 || n > len(defaultPalette) {
		out := make([]int, n)
		for i := range out {
			out[i] = i % len(defaultPalette)
		}
		return out
	}
	return paletteSets[n]
}

// PaletteFor returns n series colors chosen so that the SET is as readable as
// possible: among the palette's hues it keeps the combination whose least
// distinguishable pair is the furthest apart, measured with normal vision and
// with the three common color vision deficiencies. The petrol accent always
// comes first, and the colors are returned in palette order.
//
// Use it whenever a chart draws n series at once; PaletteColor(i) remains the
// right call when an entity's color must not depend on how many others are on
// screen (asset classes, regimes, and other fixed vocabularies).
func PaletteFor(n int) []string {
	slots := paletteSlots(n)
	out := make([]string, len(slots))
	for i, s := range slots {
		out[i] = defaultPalette[s]
	}
	return out
}

// buildPaletteSets solves, for each count, "which slots do I keep", by
// exhaustive search over the subsets that contain slot 0. The palette is
// small enough that no heuristic is warranted.
func buildPaletteSets() [][]int {
	n := len(defaultPalette)
	lab := make([][4][3]float64, n)
	for i, hex := range defaultPalette {
		lab[i] = perceivedLabs(hex)
	}
	dist := make([][]float64, n)
	for i := range dist {
		dist[i] = make([]float64, n)
		for j := range dist[i] {
			dist[i][j] = labSetDistance(lab[i], lab[j])
		}
	}
	sets := make([][]int, n+1)
	for k := 1; k <= n; k++ {
		best, bestScore := []int(nil), -1.0
		// Every subset of {1..n-1} of size k-1, joined with the anchor.
		for mask := 0; mask < 1<<(n-1); mask++ {
			if popcount(mask) != k-1 {
				continue
			}
			set := []int{0}
			for b := 0; b < n-1; b++ {
				if mask&(1<<b) != 0 {
					set = append(set, b+1)
				}
			}
			score := math.Inf(1)
			for i := 0; i < len(set); i++ {
				for j := i + 1; j < len(set); j++ {
					score = math.Min(score, dist[set[i]][set[j]])
				}
			}
			if len(set) == 1 {
				score = math.Inf(1)
			}
			if score > bestScore {
				best, bestScore = set, score
			}
		}
		sets[k] = best
	}
	return sets
}

func popcount(v int) int {
	c := 0
	for ; v != 0; v &= v - 1 {
		c++
	}
	return c
}

// perceivedLabs returns the color as four OKLab points: as seen with normal
// vision, then through each simulated deficiency.
func perceivedLabs(hex string) [4][3]float64 {
	r, g, b := hexToLinear(hex)
	var out [4][3]float64
	out[0] = linearToOklab(r, g, b)
	for i, m := range []cvdMatrix{deuteranopia, protanopia, tritanopia} {
		out[i+1] = linearToOklab(
			m[0][0]*r+m[0][1]*g+m[0][2]*b,
			m[1][0]*r+m[1][1]*g+m[1][2]*b,
			m[2][0]*r+m[2][1]*g+m[2][2]*b,
		)
	}
	return out
}

// labSetDistance scores how far apart two colors are for the READER who sees
// them most alike: the minimum over normal vision and the three deficiencies.
// Chroma is weighted twice the lightness, because two series distinguished by
// hue read faster than two distinguished by brightness alone.
func labSetDistance(a, b [4][3]float64) float64 {
	d := math.Inf(1)
	for i := range a {
		dl := a[i][0] - b[i][0]
		da := 2 * (a[i][1] - b[i][1])
		db := 2 * (a[i][2] - b[i][2])
		d = math.Min(d, math.Sqrt(dl*dl+da*da+db*db))
	}
	return d
}

// hexToLinear parses "#rrggbb" into linear-light RGB (sRGB transfer removed).
func hexToLinear(hex string) (r, g, b float64) {
	v := 0
	for i := 1; i < len(hex) && i < 7; i++ {
		c := hex[i]
		switch {
		case c >= '0' && c <= '9':
			v = v<<4 | int(c-'0')
		case c >= 'a' && c <= 'f':
			v = v<<4 | int(c-'a'+10)
		case c >= 'A' && c <= 'F':
			v = v<<4 | int(c-'A'+10)
		}
	}
	toLinear := func(c float64) float64 {
		if c <= 0.04045 {
			return c / 12.92
		}
		return math.Pow((c+0.055)/1.055, 2.4)
	}
	return toLinear(float64(v>>16&0xff) / 255),
		toLinear(float64(v>>8&0xff) / 255),
		toLinear(float64(v&0xff) / 255)
}

// linearToOklab converts linear-light RGB to OKLab, the perceptual space this
// file measures distances in (Ottosson 2020).
func linearToOklab(r, g, b float64) [3]float64 {
	l := 0.4122214708*r + 0.5363325363*g + 0.0514459929*b
	m := 0.2119034982*r + 0.6806995451*g + 0.1073969566*b
	s := 0.0883024619*r + 0.2817188376*g + 0.6299787005*b
	cbrt := func(x float64) float64 {
		if x < 0 {
			return -math.Cbrt(-x)
		}
		return math.Cbrt(x)
	}
	l, m, s = cbrt(l), cbrt(m), cbrt(s)
	return [3]float64{
		0.2104542553*l + 0.7936177850*m - 0.0040720468*s,
		1.9779984951*l - 2.4285922050*m + 0.4505937099*s,
		0.0259040371*l + 0.7827717662*m - 0.8086757660*s,
	}
}
