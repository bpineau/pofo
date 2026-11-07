package marketdata

import (
	"bytes"
	"context"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPanelSeriesPreservesSparseCadence(t *testing.T) {
	var ds []time.Time
	var vs []float64
	for d := date(2000, 1, 3); d.Year() < 2020; d = d.AddDate(0, 0, 1) {
		if d.Weekday() == time.Saturday || d.Weekday() == time.Sunday {
			continue
		}
		if d.Year() >= 2014 || d.Weekday() == time.Friday {
			ds = append(ds, d)
			vs = append(vs, 100+float64(len(vs)%7))
		}
	}
	s, _ := NewSeries("MIXED", ds, vs)
	p, err := NewPanel(Daily, s)
	if err != nil {
		t.Fatal(err)
	}
	out, err := p.Series("MIXED")
	if err != nil {
		t.Fatal(err)
	}
	st, err := out.Stats()
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(st.PeriodsPerYear-p.PeriodsPerYear()) > 1e-10 {
		t.Fatalf("series cadence %v, panel %v", st.PeriodsPerYear, p.PeriodsPerYear())
	}
}

func TestLoadOwnExport(t *testing.T) {
	s, _ := NewSeries("A", []time.Time{date(2020, 1, 1), date(2020, 2, 1)}, []float64{100, 110})
	s.Currency = "USD"
	var buf bytes.Buffer
	if err := WriteCSV(&buf, s); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "export.csv")
	if err := os.WriteFile(path, buf.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	c := NewClient("")
	c.Offline = true
	got, err := c.Load(context.Background(), path, FetchOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if got.Symbol != "A" || got.Currency != "USD" || got.Last().Close != 110 {
		t.Fatalf("lost export metadata: %+v", got)
	}
}

func TestFXQualificationSurvivesLoadAndCSV(t *testing.T) {
	c := NewClient("")
	c.Offline = true
	s, err := c.Load(context.Background(), "SP500-USD", FetchOptions{Currency: "EUR"})
	if err != nil {
		t.Fatal(err)
	}
	if s.FXExtrapolatedBefore.IsZero() || len(s.Warnings) == 0 {
		t.Fatal("FX history limit was lost")
	}
	var b bytes.Buffer
	if err := WriteCSV(&b, s); err != nil {
		t.Fatal(err)
	}
	x, err := ReadLongCSV(&b)
	if err != nil {
		t.Fatal(err)
	}
	if x[0].FXExtrapolatedBefore != s.FXExtrapolatedBefore || len(x[0].Warnings) == 0 {
		t.Fatal("CSV lost FX qualification")
	}
}

func TestPanelMetadataAndCurrencyGuard(t *testing.T) {
	ds := []time.Time{date(2020, 1, 31), date(2020, 2, 29), date(2020, 3, 31)}
	a, _ := NewSeries("A", ds, []float64{1, 1.1, 1.2})
	b, _ := NewSeries("B", ds, []float64{1, 1.2, 1.1})
	a.Currency, b.Currency = "EUR", "USD"
	a.Warnings = []string{"synthetic qualification"}
	p, err := NewPanel(Monthly, a, b)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.Mix("mix", map[string]float64{"A": .5, "B": .5}); err == nil {
		t.Fatal("mixed currencies accepted")
	}
	b.Currency = "EUR"
	p, _ = NewPanel(Monthly, a, b)
	p, err = p.Mix("mix", map[string]float64{"A": .5, "B": .5})
	if err != nil {
		t.Fatal(err)
	}
	s, err := p.Series("mix")
	if err != nil {
		t.Fatal(err)
	}
	if s.Currency != "EUR" || len(s.Warnings) != 1 {
		t.Fatalf("lost metadata %+v", s)
	}
	s.Warnings[0] = "changed"
	again, _ := p.Series("mix")
	if again.Warnings[0] == "changed" {
		t.Fatal("mutable alias")
	}
}

func TestCSVNewMetadataValidation(t *testing.T) {
	s, _ := NewSeries("A", []time.Time{date(2020, 1, 1), date(2020, 2, 1)}, []float64{1, 2})
	s.Cadence = 114.8
	s.Warnings = []string{"a\nquoted \"warning\""}
	var buf bytes.Buffer
	if err := WriteCSV(&buf, s); err != nil {
		t.Fatal(err)
	}
	x, err := ReadSeriesCSV(&buf, "")
	if err != nil {
		t.Fatal(err)
	}
	if x[0].Cadence != s.Cadence || x[0].Warnings[0] != s.Warnings[0] {
		t.Fatal("metadata round trip")
	}
	s.Cadence = math.NaN()
	if err := WriteCSV(&buf, s); err == nil {
		t.Fatal("invalid cadence accepted")
	}
	for _, header := range []string{"# cadence: -1\n", "# warnings: not-json\n", "# fx-extrapolated-before: never\n"} {
		if _, err := ReadCSV(bytes.NewBufferString(header+"2020-01-01,1\n"), "A"); err == nil {
			t.Fatalf("accepted %q", header)
		}
	}
}
