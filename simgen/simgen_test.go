package simgen

import (
	"math"
	"testing"
	"time"

	"portfodor/marketdata"
	"portfodor/metrics"
)

// fakeFetcher serves canned series, no network.
type fakeFetcher map[string]*marketdata.Series

func (f fakeFetcher) Fetch(id string, from time.Time) (*marketdata.Series, error) {
	return f[id], nil
}

func day(i int) time.Time {
	return time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, i)
}

// mkSeries builds a daily series from a constant daily growth rate.
func mkSeries(symbol string, n int, dailyGrowth float64) *marketdata.Series {
	s := &marketdata.Series{Symbol: symbol}
	v := 100.0
	for i := range n {
		s.Points = append(s.Points, marketdata.Point{Date: day(i), Close: v})
		v *= 1 + dailyGrowth
	}
	return s
}

// mkWobbly builds a series whose daily returns oscillate around drift —
// non-degenerate variance, deterministic.
func mkWobbly(symbol string, n int, drift, amp float64) *marketdata.Series {
	s := &marketdata.Series{Symbol: symbol}
	v := 100.0
	for i := range n {
		s.Points = append(s.Points, marketdata.Point{Date: day(i), Close: v})
		v *= 1 + drift + amp*math.Sin(float64(i))
	}
	return s
}

// mkLevels builds a series holding a constant level (for rates like ^IRX).
func mkLevels(symbol string, n int, level float64) *marketdata.Series {
	s := &marketdata.Series{Symbol: symbol}
	for i := range n {
		s.Points = append(s.Points, marketdata.Point{Date: day(i), Close: level})
	}
	return s
}

func near(t *testing.T, name string, got, want, tol float64) {
	t.Helper()
	if math.Abs(got-want) > tol {
		t.Errorf("%s = %v, attendu %v (±%v)", name, got, want, tol)
	}
}

func TestBuildFrameConvertsRates(t *testing.T) {
	f := fakeFetcher{
		"EQ":   mkSeries("EQ", 10, 0.01),
		"^IRX": mkLevels("^IRX", 10, 5.04), // 5.04 %/an → 0.02 %/jour
	}
	fr, err := BuildFrame(f, []string{"EQ", "^IRX"}, day(0))
	if err != nil {
		t.Fatal(err)
	}
	near(t, "rendement EQ", fr.Returns["EQ"][1], 0.01, 1e-12)
	near(t, "accrual ^IRX", fr.Returns["^IRX"][1], 5.04/100/252, 1e-12)
}

func TestCompositeNinetySixty(t *testing.T) {
	// 90 % actions (+1 %/j) + 60 % excess oblig (+0.1 %/j − cash) + 10 % cash.
	f := fakeFetcher{
		"EQ":   mkSeries("EQ", 5, 0.01),
		"BD":   mkSeries("BD", 5, 0.001),
		"^IRX": mkLevels("^IRX", 5, 2.52), // 0.0001/jour
	}
	fr, err := BuildFrame(f, []string{"EQ", "BD", "^IRX"}, day(0))
	if err != nil {
		t.Fatal(err)
	}
	values, err := Composite(fr, []Leg{
		{ID: "EQ", Weight: 0.9},
		{ID: "BD", Weight: 0.6, Excess: true},
		{ID: "^IRX", Weight: 0.1},
	}, "^IRX", 0.0)
	if err != nil {
		t.Fatal(err)
	}
	cash := 2.52 / 100 / 252
	want := 0.9*0.01 + 0.6*(0.001-cash) + 0.1*cash
	near(t, "rendement composite", values[1]/values[0]-1, want, 1e-12)
}

func TestFitBackcastRecoversLinearModel(t *testing.T) {
	n := 300
	x := mkWobbly("X", n, 0.001, 0.01)
	// real = 0.5×rendement(x) + 0.0002 par jour, exactement.
	xr := metrics.Returns(func() []float64 {
		out := make([]float64, len(x.Points))
		for i, p := range x.Points {
			out[i] = p.Close
		}
		return out
	}())
	real := &marketdata.Series{Symbol: "REAL"}
	v := 100.0
	for i := range n {
		real.Points = append(real.Points, marketdata.Point{Date: day(i), Close: v})
		if i < len(xr) {
			v *= 1 + 0.5*xr[i] + 0.0002
		}
	}
	f := fakeFetcher{"X": x}
	fr, err := BuildFrame(f, []string{"X"}, day(0))
	if err != nil {
		t.Fatal(err)
	}
	values, r2, coef, err := FitBackcast(fr, real, []string{"X"})
	if err != nil {
		t.Fatal(err)
	}
	near(t, "R²", r2, 1.0, 1e-6)
	near(t, "intercept", coef[0], 0.0002, 1e-9)
	near(t, "pente", coef[1], 0.5, 1e-9)
	if len(values) != len(fr.Dates) {
		t.Errorf("longueur reconstruite: %d", len(values))
	}
}

func TestValidatePerfectTrack(t *testing.T) {
	a := mkWobbly("A", 200, 0.002, 0.005)
	b := mkWobbly("B", 200, 0.002, 0.005)
	v, err := Validate(a, b)
	if err != nil {
		t.Fatal(err)
	}
	near(t, "corr", v.Corr, 1.0, 1e-9)
	near(t, "TE", v.TrackingErr, 0.0, 1e-9)
	near(t, "beta", v.Beta, 1.0, 1e-9)
}

func TestSpliceRealOverComposite(t *testing.T) {
	sim := mkSeries("SIM", 100, 0.001)
	real := &marketdata.Series{Symbol: "REAL", Name: "Réel"}
	for i := 50; i < 100; i++ {
		real.Points = append(real.Points, marketdata.Point{Date: day(i), Close: 200 + float64(i)})
	}
	out := Splice(real, sim)
	if len(out.Points) != 100 {
		t.Fatalf("points: %d", len(out.Points))
	}
	if !out.Points[49].Date.Before(out.Points[50].Date) {
		t.Error("ordre des dates")
	}
	if out.Points[99].Close != 299 {
		t.Errorf("la partie réelle doit rester intacte: %v", out.Points[99].Close)
	}
	if out.SimulatedBefore.IsZero() {
		t.Error("SimulatedBefore doit être posé")
	}
}

func TestSimdataRoundTrip(t *testing.T) {
	dir := t.TempDir()
	sf := &marketdata.SimdataFile{
		ID:     "TESTID",
		Name:   "Actif de test",
		Method: "méthode: x+y",
		Points: []marketdata.Point{
			{Date: day(0), Close: 100},
			{Date: day(1), Close: 101.5},
		},
	}
	if err := marketdata.WriteSimdata(dir, sf); err != nil {
		t.Fatal(err)
	}
	s, ok, err := marketdata.ReadSimdata(dir, "TESTID")
	if err != nil || !ok {
		t.Fatalf("relecture: %v, %v", ok, err)
	}
	if s.Name != "Actif de test" || len(s.Points) != 2 || s.Points[1].Close != 101.5 {
		t.Fatalf("contenu: %+v", s)
	}
	if _, ok, _ := marketdata.ReadSimdata(dir, "ABSENT"); ok {
		t.Error("un id absent ne doit rien renvoyer")
	}
}

func TestWithRefDataServesLocalFiles(t *testing.T) {
	dir := t.TempDir()
	err := marketdata.WriteSimdata(dir, &marketdata.SimdataFile{
		ID: "REF-X", Name: "Référence X",
		Points: []marketdata.Point{{Date: day(0), Close: 10}, {Date: day(1), Close: 11}},
	})
	if err != nil {
		t.Fatal(err)
	}
	f := WithRefData(dir, fakeFetcher{"AUTRE": mkSeries("AUTRE", 5, 0.01)})
	s, err := f.Fetch("REF-X", day(0))
	if err != nil || len(s.Points) != 2 || s.Points[1].Close != 11 {
		t.Fatalf("référence locale: %+v, %v", s, err)
	}
	// Absent du répertoire: fallback.
	if s, err := f.Fetch("AUTRE", day(0)); err != nil || s.Symbol != "AUTRE" {
		t.Fatalf("fallback: %+v, %v", s, err)
	}
}

func TestRefImportAppliesFeeDrag(t *testing.T) {
	dir := t.TempDir()
	pts := []marketdata.Point{}
	v := 100.0
	for i := range 253 {
		pts = append(pts, marketdata.Point{Date: day(i), Close: v})
		v *= 1.001
	}
	if err := marketdata.WriteSimdata(dir, &marketdata.SimdataFile{ID: "REF-Y", Name: "Y", Points: pts}); err != nil {
		t.Fatal(err)
	}
	build := refImport("REF-Y", "Y avec frais", 0.0252) // 0.01 %/jour
	s, err := build(WithRefData(dir, nil), day(0))
	if err != nil {
		t.Fatal(err)
	}
	// rendement quotidien net = 0.001 − 0.0001
	want := 100 * math.Pow(1.0009, 252)
	got := s.Points[len(s.Points)-1].Close
	if math.Abs(got-want)/want > 1e-6 { // tolérance: les CSV simdata arrondissent à 6 décimales
		t.Errorf("frais mal appliqués: %v, attendu %v", got, want)
	}
}
