package chart

import (
	"math"
	"testing"
)

// minSetDistance is the readability of a set of series colors: how far apart
// its LEAST distinguishable pair is, for the reader who sees them most alike.
func minSetDistance(colors []string) float64 {
	d := math.Inf(1)
	for i := range colors {
		for j := i + 1; j < len(colors); j++ {
			d = math.Min(d, labSetDistance(perceivedLabs(colors[i]), perceivedLabs(colors[j])))
		}
	}
	return d
}

func naivePrefix(n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = PaletteColor(i)
	}
	return out
}

// The reason this file exists: from four series on, taking the first n palette
// slots pairs rust with ochre, the palette's closest pair. Choosing the set
// must beat the prefix everywhere it has a choice, and by a wide margin where
// the prefix is worst.
func TestPaletteForBeatsThePrefix(t *testing.T) {
	for n := 3; n < len(defaultPalette); n++ {
		prefix := minSetDistance(naivePrefix(n))
		chosen := minSetDistance(PaletteFor(n))
		if chosen <= prefix {
			t.Errorf("n=%d: chosen set scores %.3f, the plain prefix %.3f", n, chosen, prefix)
		}
	}
	// The four-series case is the one users report; it must more than double.
	if got, want := minSetDistance(PaletteFor(4)), 2*minSetDistance(naivePrefix(4)); got < want {
		t.Errorf("four series: %.3f, expected at least %.3f", got, want)
	}
}

// Whatever the count, the palette must stay itself: the petrol accent leads,
// colors come from the palette, in slot order, without repetition.
func TestPaletteForKeepsTheIdentity(t *testing.T) {
	slot := make(map[string]int, len(defaultPalette))
	for i, c := range defaultPalette {
		slot[c] = i
	}
	for n := 1; n <= len(defaultPalette); n++ {
		got := PaletteFor(n)
		if len(got) != n {
			t.Fatalf("n=%d: %d colors", n, len(got))
		}
		if got[0] != defaultPalette[0] {
			t.Errorf("n=%d: leads with %s, expected the accent %s", n, got[0], defaultPalette[0])
		}
		seen, prev := map[string]bool{}, -1
		for _, c := range got {
			s, ok := slot[c]
			if !ok {
				t.Errorf("n=%d: %s is not a palette color", n, c)
				continue
			}
			if seen[c] {
				t.Errorf("n=%d: %s used twice", n, c)
			}
			if s <= prev {
				t.Errorf("n=%d: slot %d comes after slot %d, order is not the palette's", n, s, prev)
			}
			seen[c], prev = true, s
		}
	}
}

// Small charts keep the pairing the rest of the product uses (the book plates,
// the two-portfolio report), so a comparison does not change colors just
// because a third series was added later.
func TestPaletteForKeepsTheClassicPairing(t *testing.T) {
	for n := 1; n <= 2; n++ {
		for i, c := range PaletteFor(n) {
			if c != PaletteColor(i) {
				t.Errorf("n=%d: color %d is %s, expected %s", n, i, c, PaletteColor(i))
			}
		}
	}
}

// Past the palette, colors cycle and the caller must label the twins; the
// contract is only that the call stays defined and stable.
func TestPaletteForCyclesBeyondThePalette(t *testing.T) {
	got := PaletteFor(len(defaultPalette) + 2)
	if len(got) != len(defaultPalette)+2 {
		t.Fatalf("%d colors", len(got))
	}
	for i, c := range got {
		if c != PaletteColor(i) {
			t.Errorf("color %d is %s, expected %s", i, c, PaletteColor(i))
		}
	}
	if len(PaletteFor(0)) != 0 {
		t.Error("PaletteFor(0) should be empty")
	}
}

// The distance must be measured for the reader who is worst served, so a pair
// that collapses under one deficiency cannot be rescued by looking fine to
// everyone else. Rust and ochre are that pair.
func TestDistanceIsTheWorstObserver(t *testing.T) {
	rust, ochre := perceivedLabs(defaultPalette[1]), perceivedLabs(defaultPalette[3])
	dl, da, db := rust[0][0]-ochre[0][0], 2*(rust[0][1]-ochre[0][1]), 2*(rust[0][2]-ochre[0][2])
	normal := math.Sqrt(dl*dl + da*da + db*db)
	worst := labSetDistance(rust, ochre)
	if worst >= normal {
		t.Errorf("rust/ochre: worst observer %.3f, normal vision %.3f; the simulation is not biting", worst, normal)
	}
}

// A sanity check on the color pipeline itself: black and white are the two
// ends of the lightness axis, and a color is at distance zero from itself.
func TestOklabPipeline(t *testing.T) {
	if d := labSetDistance(perceivedLabs("#0880A8"), perceivedLabs("#0880A8")); d != 0 {
		t.Errorf("a color is at %.3f from itself", d)
	}
	br, bg, bb := hexToLinear("#000000")
	wr, wg, wb := hexToLinear("#FFFFFF")
	black, white := linearToOklab(br, bg, bb), linearToOklab(wr, wg, wb)
	if math.Abs(black[0]) > 1e-6 || math.Abs(white[0]-1) > 1e-3 {
		t.Errorf("lightness axis: black %.4f, white %.4f, expected 0 and 1", black[0], white[0])
	}
}
