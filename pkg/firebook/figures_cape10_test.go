package firebook

import (
	"bufio"
	"math"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
)

// capePair is one start month of the scatter, rebuilt from the bundled data.
type capePair struct {
	cape, ret float64
}

// capeRebuild redoes, from the three bundled sources, exactly what the plate
// froze: for every month with a full decade ahead of it, the Shiller CAPE of
// that month and the annualized real total return of the S&P 500 over the 120
// months that follow. Everything is offline (the CPI is served from the
// marketdata snapshot), so this never touches the network.
func capeRebuild(t *testing.T) []capePair {
	t.Helper()
	frozenAgainstData(t)
	key := func(d time.Time) int { return d.Year()*12 + int(d.Month()) - 1 }

	cape := map[int]float64{}
	sc := bufio.NewScanner(strings.NewReader(string(datasets.CAPE())))
	for sc.Scan() {
		l := strings.TrimSpace(sc.Text())
		if l == "" || strings.HasPrefix(l, "#") || strings.HasPrefix(l, "date") {
			continue
		}
		f := strings.Split(l, ",")
		if len(f) < 2 {
			continue
		}
		d, err := time.Parse("2006-01-02", f[0])
		if err != nil {
			continue
		}
		if v, err := strconv.ParseFloat(f[1], 64); err == nil && v > 0 {
			cape[key(d)] = v
		}
	}
	if len(cape) < 1500 {
		t.Fatalf("bundled CAPE holds only %d months", len(cape))
	}

	s, ok, err := marketdata.ReadSimdataFS(datasets.Refdata(), "SP500-USD")
	if err != nil || !ok {
		t.Fatalf("SP500-USD reference: %v (found=%v)", err, ok)
	}
	price := map[int]float64{}
	for _, p := range s.Points {
		if p.Close > 0 {
			price[key(p.Date)] = p.Close
		}
	}
	// The file ends on the last trading day available, which may open an
	// incomplete month; a start is only kept when its whole decade is quoted.
	last := s.Points[len(s.Points)-1].Date
	lastFull := key(last)
	if last.AddDate(0, 0, 8).Month() == last.Month() {
		lastFull--
	}

	cs, err := marketdata.NewClient("").Fetch(t.Context(), "^CPI-US", time.Time{})
	if err != nil {
		t.Fatalf("^CPI-US: %v", err)
	}
	cpi := map[int]float64{}
	for _, p := range cs.Points {
		if p.Date.Day() == 1 && p.Close > 0 {
			cpi[key(p.Date)] = p.Close // the monthly anchor of the interpolated index
		}
	}

	months := make([]int, 0, len(cape))
	for k := range cape {
		months = append(months, k)
	}
	sort.Ints(months)

	var out []capePair
	for _, k := range months {
		e := k + 120
		if e > lastFull {
			continue
		}
		p0, ok0 := price[k]
		p1, ok1 := price[e]
		c0, ok2 := cpi[k]
		c1, ok3 := cpi[e]
		if !ok0 || !ok1 || !ok2 || !ok3 {
			continue
		}
		real := (p1 / p0) / (c1 / c0)
		out = append(out, capePair{cape[k], (math.Pow(real, 0.1) - 1) * 100})
	}
	return out
}

// The scatter is frozen in the plate; the bundled series are the truth. Any
// drift in either (a CAPE refresh, a rebuilt SP500-USD, a CPI snapshot update)
// must be seen and re-frozen deliberately, never absorbed silently.
func TestCapeScatterMatchesTheBundledData(t *testing.T) {
	pairs := capeRebuild(t)
	if len(capeStarts) != len(capeForward10) {
		t.Fatalf("%d CAPE against %d returns", len(capeStarts), len(capeForward10))
	}
	if len(pairs) != len(capeStarts) {
		t.Fatalf("the data yields %d starts, the plate froze %d", len(pairs), len(capeStarts))
	}
	for i, p := range pairs {
		if math.Abs(p.cape-capeStarts[i]) > 0.006 {
			t.Errorf("start %d: CAPE %.4f in the data, %.2f on the plate", i, p.cape, capeStarts[i])
		}
		if math.Abs(p.ret-capeForward10[i]) > 0.006 {
			t.Errorf("start %d: ten-year real %.4f %% in the data, %.2f %% on the plate", i, p.ret, capeForward10[i])
		}
	}

	// Sanity anchors nobody should be able to break unnoticed: the record runs
	// over a century, and the S&P 500 pays around 7 % real per year over it.
	if len(pairs) < 1200 {
		t.Errorf("only %d starts: the bundled window shrank", len(pairs))
	}
	var mean float64
	for _, p := range pairs {
		mean += p.ret
	}
	if mean /= float64(len(pairs)); mean < 5 || mean > 9 {
		t.Errorf("mean ten-year real return %.2f %% per year, outside the 5-9 %% sanity band", mean)
	}
}

// The plate names a specification and an R squared. Both are recomputed here
// from the same data: the plate may not quote a fit it does not have.
func TestCapeFitMatchesTheRegression(t *testing.T) {
	pairs := capeRebuild(t)
	n := float64(len(pairs))
	var sx, sy, sxx, sxy, syy float64
	for _, p := range pairs {
		x, y := 100/p.cape, p.ret // earnings yield in percent, return in percent
		sx += x
		sy += y
		sxx += x * x
		sxy += x * y
		syy += y * y
	}
	b := (n*sxy - sx*sy) / (n*sxx - sx*sx)
	a := (sy - b*sx) / n
	num := n*sxy - sx*sy
	r2 := num * num / ((n*sxx - sx*sx) * (n*syy - sy*sy))

	if math.Abs(a-capeFitA) > 0.01 {
		t.Errorf("intercept %.4f, the plate says %.4f", a, capeFitA)
	}
	if math.Abs(b-capeFitB) > 0.01 {
		t.Errorf("slope %.4f, the plate says %.4f", b, capeFitB)
	}
	if math.Abs(r2-capeFitR2) > 0.005 {
		t.Errorf("R2 %.4f, the plate says %.4f", r2, capeFitR2)
	}
	// The article claims the fit descends and explains around a third of the
	// variance at ten years, not the half it used to claim.
	if b <= 0 {
		t.Error("the fit must rise with the earnings yield, that is the whole point")
	}
	if r2 < 0.2 || r2 > 0.4 {
		t.Errorf("R2 %.3f is no longer the order of magnitude the article quotes", r2)
	}
}

// The three slices are the message: the centre moves, the width does not follow.
func TestCapeBandsMatchTheData(t *testing.T) {
	pairs := capeRebuild(t)
	for _, s := range capeBands {
		var v []float64
		for _, p := range pairs {
			if p.cape >= s.lo && p.cape <= s.hi {
				v = append(v, p.ret)
			}
		}
		sort.Float64s(v)
		if len(v) != s.n {
			t.Errorf("CAPE %.0f to %.0f: %d starts in the data, %d on the plate", s.lo, s.hi, len(v), s.n)
			continue
		}
		q := func(f float64) float64 {
			i := f * float64(len(v)-1)
			lo, hi := int(math.Floor(i)), int(math.Ceil(i))
			return v[lo] + (i-float64(lo))*(v[hi]-v[lo])
		}
		for _, c := range []struct {
			name       string
			got, plate float64
		}{
			{"minimum", v[0], s.min},
			{"first decile", q(0.1), s.d1},
			{"median", q(0.5), s.med},
			{"ninth decile", q(0.9), s.d9},
			{"maximum", v[len(v)-1], s.max_},
		} {
			if math.Abs(c.got-c.plate) > 0.006 {
				t.Errorf("CAPE %.0f to %.0f, %s: %.4f %% in the data, %.2f %% on the plate",
					s.lo, s.hi, c.name, c.got, c.plate)
			}
		}
	}

	// What the plate asserts in words, checked as arithmetic: the centre falls
	// from one slice to the next, while the eight-out-of-ten span does not.
	for i := 1; i < len(capeBands); i++ {
		if capeBands[i].med >= capeBands[i-1].med {
			t.Errorf("median at CAPE %.0f (%.2f) is not below the one at CAPE %.0f (%.2f)",
				capeBands[i].centre, capeBands[i].med, capeBands[i-1].centre, capeBands[i-1].med)
		}
	}
	if w0, w1 := capeBands[0].d9-capeBands[0].d1, capeBands[1].d9-capeBands[1].d1; w1 < w0 {
		t.Errorf("the spread narrows from %.2f to %.2f points between the first two slices, "+
			"the plate says it does not", w0, w1)
	}
}

// The plate must survive crengine and the house rules, and carry the readings
// it is built around.
func TestCapeDixAnsPlate(t *testing.T) {
	svg := figCapeDixAns()
	for _, banned := range []string{"rgba", "opacity", "—", "linearGradient"} {
		if strings.Contains(svg, banned) {
			t.Errorf("the plate contains %q, which crengine or the house rules forbid", banned)
		}
	}
	if got := strings.Count(svg, "<circle"); got != len(capeStarts)+len(capeBands) {
		t.Errorf("%d circles drawn for %d starts and %d medians", got, len(capeStarts), len(capeBands))
	}
	for _, want := range []string{
		"R² = 0,29",                 // the measured fit, named on the plate
		"rendement = 1,0 + 0,86",    // the specification, named on the plate
		"1 242 départs mensuels",    // the sample size
		"janvier 1913 à juin 2016",  // the window the bundled series allow
		"+10,2 %", "+5,7 %", "+2,8", // the three medians
		"164 départs", "219 départs", "48 départs",
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("the plate never says %q", want)
		}
	}
	if n := len(capeStarts); n != 1242 {
		t.Errorf("the plate says 1 242 starts but froze %d", n)
	}
}
