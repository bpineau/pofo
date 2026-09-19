package firebook

import (
	"math"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/simgen"
)

// The guard test of the tente-matiere plate. It rebuilds the four defensive
// pockets from the bundled series, re-solves every vintage's sustainable rate
// under the tent and under its static twin, and fails when any frozen number
// drifts. The equity leg, the fold into December-to-December real years, the
// tent and the bisection are all the neighbouring plate's, shared verbatim
// (tenteLeg, tenteCPI, tenteYearly, tenteAlloc, tenteSWR): the whole point of
// the plate is that ONLY the pocket changes.

// matiereCash accrues the bundled 3-month bill rate into a total-return level:
// one twelfth of the PREVIOUS month's annualized rate, the book's standing cash
// convention (see the all-weather plate). The rate is a percent per year, so
// the monthly factor is 1 + r/1200.
func matiereCash(bills map[int]float64) map[int]float64 {
	keys := make([]int, 0, len(bills))
	for k := range bills {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	out := make(map[int]float64, len(keys))
	val := 100.0
	out[keys[0]] = val
	for i := 1; i < len(keys); i++ {
		if keys[i] != keys[i-1]+1 {
			continue // a hole ends the accrual rather than inventing a month
		}
		val *= 1 + bills[keys[i-1]]/1200
		out[keys[i]] = val
	}
	return out
}

// matiereLongTR builds the long pocket. The bundled TREASURY-LONG-USD cannot
// serve: the 20-year constant maturity was discontinued between 1987-01 and
// 1993-09 and the series carries that hole, which would cut every thirty-year
// vintage from 1958 to 1993 out of the sample. The long PAR YIELD is whole from
// 1953-04 on, so the pocket is priced off it with the same recipe the bundled
// series uses, a 20-year par bond at 0,10 %/yr, sampled month-end.
//
// The two agree where both exist: 0,978 correlation on yearly real returns over
// 1954-2025, and on the vintages the bundled series CAN carry the bundled leg
// makes the long pocket look slightly WORSE than the rebuilt one does, so
// nothing the plate concludes rests on the substitution.
func matiereLongTR(t *testing.T) map[int]float64 {
	t.Helper()
	frozenAgainstData(t)
	s, ok, err := marketdata.ReadSimdataFS(datasets.Refdata(), "TREASURY-LONG-YIELD")
	if err != nil || !ok {
		t.Fatalf("read TREASURY-LONG-YIELD: ok=%v err=%v", ok, err)
	}
	// month-end sample: the series is monthly before 1962 and daily after, and
	// the reconstruction wants one step a month either way.
	last := make(map[int]float64, len(s.Points))
	seen := make(map[int]time.Time, len(s.Points))
	for _, p := range s.Points {
		k := tenteMonth(p.Date)
		if d, ok := seen[k]; !ok || p.Date.After(d) {
			last[k], seen[k] = p.Close, p.Date
		}
	}
	keys := make([]int, 0, len(last))
	for k := range last {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	yields := &marketdata.Series{Name: "long par yield"}
	for _, k := range keys {
		yields.Points = append(yields.Points, marketdata.Point{
			Date:  time.Date(k/12, time.Month(k%12+1), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, -1),
			Close: last[k],
		})
	}
	out := map[int]float64{}
	for _, p := range simgen.TreasuryTR("long", yields, matiereLongMaturity, matiereLongFee).Points {
		out[tenteMonth(p.Date)] = p.Close
	}
	return out
}

// The long pocket's own parameters, the bundled 20-year series' own recipe.
const (
	matiereLongMaturity = 20.0
	matiereLongFee      = 0.001
)

// matiereSolved is one re-solved vintage on one pocket.
type matiereSolved struct {
	tent, static float64
}

// matiereSolve replays the plate's whole recipe from the bundled datasets and
// returns, per pocket name, the solved rates keyed by departure year.
func matiereSolve(t *testing.T) (map[string]map[int]matiereSolved, []int) {
	t.Helper()
	cpi := tenteCPI(t)
	eq := tenteYearly(tenteLeg(t, "SP500-USD"), cpi)
	cash := tenteYearly(matiereCash(tenteLeg(t, "TBILL-3M")), cpi)
	inter := tenteYearly(tenteLeg(t, "TREASURY-INT-USD"), cpi)
	long := tenteYearly(matiereLongTR(t), cpi)
	// the fifty-fifty is rebalanced every December, which on yearly growth
	// factors is exactly their average
	mix := map[int]float64{}
	for y, a := range cash {
		if b, ok := inter[y]; ok {
			mix[y] = (a + b) / 2
		}
	}
	pockets := map[string]map[int]float64{
		"cash": cash, "int": inter, "long": long, "mix": mix,
	}

	static := func(int) float64 { return tenteStatic }
	out := map[string]map[int]matiereSolved{}
	for name := range pockets {
		out[name] = map[int]matiereSolved{}
	}
	var years []int
	for y := matiereFirst; y <= matiereLast; y++ {
		solved := map[string]matiereSolved{}
		for name, d := range pockets {
			re := make([]float64, 0, tenteYears)
			rb := make([]float64, 0, tenteYears)
			for k := 0; k < tenteYears; k++ {
				a, oka := eq[y+k]
				b, okb := d[y+k]
				if !oka || !okb {
					break
				}
				re, rb = append(re, a), append(rb, b)
			}
			if len(re) < tenteYears {
				return out, years // the record stops here
			}
			solved[name] = matiereSolved{
				tent:   tenteSWR(re, rb, tenteAlloc) * 100,
				static: tenteSWR(re, rb, static) * 100,
			}
		}
		years = append(years, y)
		for name, s := range solved {
			out[name][y] = s
		}
	}
	return out, years
}

// Every one of the plate's forty-three triples is re-solved from the datasets.
func TestMatiereVintagesMatchTheSolver(t *testing.T) {
	solved, years := matiereSolve(t)
	if len(years) != len(matiereVintages) {
		t.Fatalf("the datasets give %d vintages, the plate draws %d", len(years), len(matiereVintages))
	}
	if years[0] != matiereFirst || years[len(years)-1] != matiereLast {
		t.Fatalf("vintages run %d-%d, the plate says %d-%d", years[0], years[len(years)-1], matiereFirst, matiereLast)
	}
	for i, v := range matiereVintages {
		y := years[i]
		if y != v.year {
			t.Errorf("row %d: solver says %d, plate says %d", i, y, v.year)
			continue
		}
		base := solved["int"][y].tent
		if math.Abs(base-v.base) > 0.005 {
			t.Errorf("%d: the intermediate pocket sustains %.3f %%, the plate draws %.3f %%", y, base, v.base)
		}
		if got := solved["long"][y].tent - base; math.Abs(got-v.long) > 0.005 {
			t.Errorf("%d: the long pocket is %+.3f pt away, the plate draws %+.3f pt", y, got, v.long)
		}
		if got := solved["cash"][y].tent - base; math.Abs(got-v.cash) > 0.005 {
			t.Errorf("%d: the cash pocket is %+.3f pt away, the plate draws %+.3f pt", y, got, v.cash)
		}
	}
}

// The readings printed on the plate: the single sign change of the long pocket,
// where it falls, and the vintage that is worst on every pocket.
func TestMatiereReadingsMatchTheCurves(t *testing.T) {
	solved, years := matiereSolve(t)
	neg, pos, lastNeg := 0, 0, 0
	for _, y := range years {
		switch gap := solved["long"][y].tent - solved["int"][y].tent; {
		case gap < 0:
			neg++
			lastNeg = y
		default:
			pos++
		}
	}
	if neg != matiereLongNeg || pos != matiereLongPos {
		t.Errorf("the long pocket costs on %d vintages and pays on %d, the plate says %d and %d",
			neg, pos, matiereLongNeg, matiereLongPos)
	}
	if lastNeg != matierePivot {
		t.Errorf("the last vintage the long pocket costs is %d, the plate says %d", lastNeg, matierePivot)
	}
	// One sign change and one only: every vintage up to the pivot loses, every
	// one after it gains. That is what makes the plate's two blocks honest.
	for _, y := range years {
		gap := solved["long"][y].tent - solved["int"][y].tent
		if y <= matierePivot && gap >= 0 {
			t.Errorf("%d is before the pivot but the long pocket gains %+.3f pt", y, gap)
		}
		if y > matierePivot && gap <= 0 {
			t.Errorf("%d is after the pivot but the long pocket loses %+.3f pt", y, gap)
		}
	}

	// The worst departure of the sample, on every pocket alike.
	for _, name := range []string{"cash", "int", "long", "mix"} {
		worstY, worst := 0, math.Inf(1)
		for _, y := range years {
			if v := solved[name][y].tent; v < worst {
				worst, worstY = v, y
			}
		}
		if worstY != matiereWorstYear {
			t.Errorf("the %s pocket's worst departure is %d, the plate says %d", name, worstY, matiereWorstYear)
		}
		if name == "mix" && math.Abs(worst-matiereMixWorst) > 0.005 {
			t.Errorf("the fifty-fifty sustains %.3f %% at the worst departure, the plate says %.2f %%",
				worst, matiereMixWorst)
		}
	}
}

// The plate's central claim: what the SLOPE is worth does not depend on the
// material. The band of the median gain and the band of the worst gain, taken
// across the four pockets, are the two numbers the notes print.
func TestMatiereSlopeDoesNotDependOnTheMaterial(t *testing.T) {
	solved, years := matiereSolve(t)
	medLo, medHi := math.Inf(1), math.Inf(-1)
	worstLo, worstHi := math.Inf(1), math.Inf(-1)
	for _, name := range []string{"cash", "int", "long", "mix"} {
		gains := make([]float64, 0, len(years))
		for _, y := range years {
			s := solved[name][y]
			gains = append(gains, s.tent-s.static)
		}
		sort.Float64s(gains)
		med, worst := gains[len(gains)/2], gains[0]
		medLo, medHi = math.Min(medLo, med), math.Max(medHi, med)
		worstLo, worstHi = math.Min(worstLo, worst), math.Max(worstHi, worst)
	}
	if math.Abs(medLo-matiereGainMedLo) > 0.005 || math.Abs(medHi-matiereGainMedHi) > 0.005 {
		t.Errorf("the median gain runs %+.3f to %+.3f pt, the plate says %+.3f to %+.3f",
			medLo, medHi, matiereGainMedLo, matiereGainMedHi)
	}
	if math.Abs(worstLo-matiereGainWorstLo) > 0.005 || math.Abs(worstHi-matiereGainWorstHi) > 0.005 {
		t.Errorf("the worst gain runs %+.3f to %+.3f pt, the plate says %+.3f to %+.3f",
			worstLo, worstHi, matiereGainWorstLo, matiereGainWorstHi)
	}
	// The claim itself, stated as a bound rather than as four numbers: whatever
	// the pocket, the tent's median vintage and its worst vintage land within a
	// twentieth of a point and a twentieth of a point of each other.
	if medHi-medLo > 0.05 {
		t.Errorf("the median gain spans %.3f pt across the pockets, more than the plate's claim", medHi-medLo)
	}
	if worstHi-worstLo > 0.06 {
		t.Errorf("the worst gain spans %.3f pt across the pockets, more than the plate's claim", worstHi-worstLo)
	}
}

// The pockets' inputs, checked against what is known about them before any
// conclusion is drawn on top: bills near zero real over the long run, the long
// bond deeply negative through the 1965-1981 bond bear and strongly positive
// through the disinflation, the intermediate leg between the two both times.
func TestMatierePocketsAreSane(t *testing.T) {
	cpi := tenteCPI(t)
	legs := map[string]map[int]float64{
		"cash": tenteYearly(matiereCash(tenteLeg(t, "TBILL-3M")), cpi),
		"int":  tenteYearly(tenteLeg(t, "TREASURY-INT-USD"), cpi),
		"long": tenteYearly(matiereLongTR(t), cpi),
	}
	cagr := func(name string, from, to int) float64 {
		acc, n := 1.0, 0
		for y := from; y <= to; y++ {
			f, ok := legs[name][y]
			if !ok {
				t.Fatalf("%s has no %d", name, y)
			}
			acc *= f
			n++
		}
		return (math.Pow(acc, 1/float64(n)) - 1) * 100
	}
	type band struct {
		name     string
		from, to int
		lo, hi   float64
	}
	for _, b := range []band{
		// The long run: bills barely keep up with inflation, bonds earn a
		// little, and the term premium orders the three.
		{"cash", 1954, 2025, 0.2, 1.3},
		{"int", 1954, 2025, 1.0, 2.2},
		{"long", 1954, 2025, 1.2, 2.5},
		// The bond bear. Ibbotson's long governments returned about 2,5 %/yr
		// nominal against 7 %/yr of inflation over those years.
		{"long", 1965, 1981, -5.0, -3.5},
		{"int", 1965, 1981, -2.2, -0.8},
		{"cash", 1965, 1981, -0.5, 0.6},
		// The disinflation, the mirror image.
		{"long", 1982, 2012, 6.5, 8.0},
		{"int", 1982, 2012, 4.2, 5.3},
		{"cash", 1982, 2012, 1.2, 2.1},
	} {
		if got := cagr(b.name, b.from, b.to); got < b.lo || got > b.hi {
			t.Errorf("%s over %d-%d: %+.2f %%/yr real, outside the expected %.1f to %.1f",
				b.name, b.from, b.to, got, b.lo, b.hi)
		}
	}
}

// House rules of the plate system, checked on the rendered surface.
func TestTenteMatiereRenders(t *testing.T) {
	if figures["tente-matiere"] == nil {
		t.Fatal("tente-matiere is not registered in the figures map")
	}
	s := figTenteMatiere()
	for _, want := range []string{
		"viewBox", "poche monétaire", "poche longue (État 20 ans)",
		"1966, le pire départ", "1954-1980", "3,96 %", "4,03 %", "3,69 %",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("the plate never mentions %q", want)
		}
	}
	// The banned list holds the em-dash as an escape, so this file carries none.
	for _, banned := range []string{"rgba(", "opacity", "—", "–", "rotate("} {
		if strings.Contains(s, banned) {
			t.Errorf("the plate uses %q, which the plate system forbids", banned)
		}
	}
}
