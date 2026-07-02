package portfolio

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

func buildTestSeries(symbol, currency string) *marketdata.Series {
	s := &marketdata.Series{Symbol: symbol, Name: symbol + " fund", Currency: currency}
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := range 10 {
		s.Points = append(s.Points, marketdata.Point{Date: start.AddDate(0, 0, i), Close: 100 + float64(i)})
	}
	return s
}

func fetchFrom(m map[string]*marketdata.Series) func(string) (*marketdata.Series, error) {
	return func(id string) (*marketdata.Series, error) {
		if s, ok := m[id]; ok {
			return s, nil
		}
		return nil, errors.New("unknown asset")
	}
}

func TestBuildMapsSpec(t *testing.T) {
	spec, err := Parse("mix", strings.NewReader(`
#meta capital:10000 contribute:100/month extra-fees:0.5
60 EQ 0.07
40 BD
`))
	if err != nil {
		t.Fatal(err)
	}
	series := map[string]*marketdata.Series{
		"EQ": buildTestSeries("EQ", "EUR"),
		"BD": buildTestSeries("BD", "EUR"),
	}
	p, err := Build(spec, BuildOptions{
		Fetch:        fetchFrom(series),
		Fees:         func(id string) (float64, bool) { return 0.22, id == "BD" },
		BaseCurrency: "EUR",
	})
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "mix" || p.Capital != 10000 || p.EnvelopeFees != 0.5 || !p.Contribute.Active() {
		t.Errorf("directives not carried over: %+v", p)
	}
	if len(p.Assets) != 2 || p.Assets[0].Weight != 0.60 || p.Assets[1].Weight != 0.40 {
		t.Fatalf("assets misbuilt: %+v", p.Assets)
	}
	if p.Assets[0].Fees != 0.07 {
		t.Errorf("the file TER must take precedence, got %v", p.Assets[0].Fees)
	}
	if p.Assets[1].Fees != 0.22 {
		t.Errorf("the missing TER must come from the Fees lookup, got %v", p.Assets[1].Fees)
	}
	if len(p.Warnings) != 0 {
		t.Errorf("unexpected warnings: %v", p.Warnings)
	}
	if _, err := Simulate(p, 30); err != nil {
		t.Errorf("the built portfolio must simulate: %v", err)
	}
}

func TestBuildLeverageDefaults(t *testing.T) {
	spec, err := Parse("lev", strings.NewReader("#meta leverage:on\n90 EQ\n60 BD\n"))
	if err != nil {
		t.Fatal(err)
	}
	cash := buildTestSeries("^IRX", "")
	series := map[string]*marketdata.Series{
		"EQ": buildTestSeries("EQ", "USD"),
		"BD": buildTestSeries("BD", "USD"),
	}
	p, err := Build(spec, BuildOptions{Fetch: fetchFrom(series), Cash: cash, BorrowSpread: 1})
	if err != nil {
		t.Fatal(err)
	}
	if !p.Leverage || p.Cash != cash {
		t.Error("leverage wiring lost")
	}
	if p.BorrowSpread != 1 {
		t.Errorf("default borrow spread not applied: %v", p.BorrowSpread)
	}

	spec2, err := Parse("lev2", strings.NewReader("#meta leverage:on borrow-spread:2.5\n90 EQ\n60 BD\n"))
	if err != nil {
		t.Fatal(err)
	}
	p2, err := Build(spec2, BuildOptions{Fetch: fetchFrom(series), BorrowSpread: 1})
	if err != nil {
		t.Fatal(err)
	}
	if p2.BorrowSpread != 2.5 {
		t.Errorf("the spec borrow spread must take precedence, got %v", p2.BorrowSpread)
	}
}

func TestBuildCurrencyWarnings(t *testing.T) {
	spec, err := Parse("fx", strings.NewReader("50 EQ\n50 BD\n"))
	if err != nil {
		t.Fatal(err)
	}
	series := map[string]*marketdata.Series{
		"EQ": buildTestSeries("EQ", "USD"),
		"BD": buildTestSeries("BD", ""),
	}
	p, err := Build(spec, BuildOptions{Fetch: fetchFrom(series), BaseCurrency: "EUR"})
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(p.Warnings, "; ")
	if !strings.Contains(joined, "unknown currency") {
		t.Errorf("expected an unknown-currency warning, got %v", p.Warnings)
	}

	series["BD"] = buildTestSeries("BD", "EUR")
	p, err = Build(spec, BuildOptions{Fetch: fetchFrom(series), BaseCurrency: "EUR"})
	if err != nil {
		t.Fatal(err)
	}
	joined = strings.Join(p.Warnings, "; ")
	if !strings.Contains(joined, "mixed currencies (EUR, USD)") {
		t.Errorf("expected a mixed-currency warning, got %v", p.Warnings)
	}
}

func TestBuildErrors(t *testing.T) {
	spec, err := Parse("p", strings.NewReader("100 EQ\n"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Build(spec, BuildOptions{}); err == nil {
		t.Error("a missing Fetch must be rejected")
	}
	sentinel := errors.New("network down")
	_, err = Build(spec, BuildOptions{Fetch: func(string) (*marketdata.Series, error) { return nil, sentinel }})
	if !errors.Is(err, sentinel) {
		t.Errorf("the fetch error must be wrapped, got %v", err)
	}
	if err == nil || !strings.Contains(err.Error(), `asset "EQ"`) {
		t.Errorf("the failing asset must be named, got %v", err)
	}
}

func TestBuildDoesNotMutateSpecWarnings(t *testing.T) {
	spec, err := Parse("odd", strings.NewReader("30 EQ\n30 BD\n")) // sums to 60: one warning
	if err != nil {
		t.Fatal(err)
	}
	if len(spec.Warnings) != 1 {
		t.Fatalf("expected the normalization warning, got %v", spec.Warnings)
	}
	series := map[string]*marketdata.Series{
		"EQ": buildTestSeries("EQ", "USD"),
		"BD": buildTestSeries("BD", "EUR"),
	}
	p, err := Build(spec, BuildOptions{Fetch: fetchFrom(series)})
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Warnings) != 2 { // normalization + mixed currencies
		t.Errorf("warnings: %v", p.Warnings)
	}
	if len(spec.Warnings) != 1 {
		t.Errorf("Build must not grow the spec's warnings: %v", spec.Warnings)
	}
}
