package portfolio

import (
	"math"
	"strings"
	"testing"

	"github.com/bpineau/pofo/pkg/marketdata"
)

func TestParseBasic(t *testing.T) {
	in := `
# comment
60   VOO    # Vanguard S&P 500
40	BND  # US bonds
`
	spec, err := Parse("test", strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	if len(spec.Holdings) != 2 {
		t.Fatalf("want 2 lines, got %d", len(spec.Holdings))
	}
	h := spec.Holdings[0]
	if h.ID != "VOO" || math.Abs(h.Weight-0.60) > 1e-12 {
		t.Errorf("line 1 misread: %+v", h)
	}
	h = spec.Holdings[1]
	if h.ID != "BND" || math.Abs(h.Weight-0.40) > 1e-12 {
		t.Errorf("line 2 misread: %+v", h)
	}
	if len(spec.Warnings) != 0 {
		t.Errorf("no warning expected, got %v", spec.Warnings)
	}
}

func TestParseInlineComments(t *testing.T) {
	in := `
# Test portfolio
# https://example.invalid/doc

60 VOO # the S&P 500
40 BND# glued to the ticker
`
	spec, err := Parse("t", strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	if len(spec.Holdings) != 2 {
		t.Fatalf("want 2 lines, got %d", len(spec.Holdings))
	}
	if h := spec.Holdings[0]; h.ID != "VOO" {
		t.Errorf("comment not stripped: %+v", h)
	}
	if h := spec.Holdings[1]; h.ID != "BND" {
		t.Errorf("glued comment not stripped: %+v", h)
	}
	// Free text after the ticker, without a "#", is now rejected.
	if _, err := Parse("t", strings.NewReader("60 VOO useful note")); err == nil {
		t.Error("expected error for un-commented free text after the ticker")
	}
}

func TestParseDecimalCommaAndPercent(t *testing.T) {
	spec, err := Parse("t", strings.NewReader("33,5% IWDA.AS\n66.5 IE00B4L5Y983 # world"))
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(spec.Holdings[0].RawWeight-33.5) > 1e-12 {
		t.Errorf("decimal comma: %v", spec.Holdings[0].RawWeight)
	}
	if math.Abs(spec.Holdings[1].RawWeight-66.5) > 1e-12 {
		t.Errorf("decimal point: %v", spec.Holdings[1].RawWeight)
	}
}

func TestParseNormalizesWeights(t *testing.T) {
	spec, err := Parse("t", strings.NewReader("50 A\n100 B"))
	if err != nil {
		t.Fatal(err)
	}
	if len(spec.Warnings) != 1 {
		t.Fatalf("normalization warning expected, got %v", spec.Warnings)
	}
	if math.Abs(spec.Holdings[0].Weight-1.0/3) > 1e-12 || math.Abs(spec.Holdings[1].Weight-2.0/3) > 1e-12 {
		t.Errorf("weights not normalized: %+v", spec.Holdings)
	}
}

func TestParseMetaRebalance(t *testing.T) {
	in := `
#meta rebalance:30   # comment tolerated
60 VOO
40 BND
`
	spec, err := Parse("t", strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	if spec.RebalanceDays != 30 {
		t.Errorf("RebalanceDays = %d, want 30", spec.RebalanceDays)
	}
	if spec.Meta["rebalance"] != "30" {
		t.Errorf("raw Meta: %+v", spec.Meta)
	}

	// Without a directive: -1 (the caller's default applies).
	spec, err = Parse("t", strings.NewReader("100 VOO"))
	if err != nil {
		t.Fatal(err)
	}
	if spec.RebalanceDays != -1 {
		t.Errorf("RebalanceDays without directive = %d, want -1", spec.RebalanceDays)
	}

	// rebalance:0 = never rebalance (distinct from unspecified).
	spec, err = Parse("t", strings.NewReader("#meta rebalance:0"+"\n"+"100 VOO"))
	if err != nil {
		t.Fatal(err)
	}
	if spec.RebalanceDays != 0 {
		t.Errorf("RebalanceDays = %d, want 0", spec.RebalanceDays)
	}

	// Unknown key: warning, not an error.
	spec, err = Parse("t", strings.NewReader("#meta fancy:yes"+"\n"+"100 VOO"))
	if err != nil {
		t.Fatal(err)
	}
	if len(spec.Warnings) != 1 {
		t.Errorf("warning expected for unknown key: %v", spec.Warnings)
	}

	// "#metadata" is not a directive, just a comment.
	if _, err := Parse("t", strings.NewReader("#metadata blabla"+"\n"+"100 VOO")); err != nil {
		t.Errorf("#metadata must stay a comment: %v", err)
	}

	// Invalid value: explicit error.
	if _, err := Parse("t", strings.NewReader("#meta rebalance:often"+"\n"+"100 VOO")); err == nil {
		t.Error("expected error for a non-numeric value")
	}
}

func TestParseFeesColumnAndEnvelope(t *testing.T) {
	in := `
#meta extra-fees:0,60  # life-insurance envelope
60 VOO 0.03  # 3rd numeric column = TER
40 BND       # no declared fees
`
	spec, err := Parse("t", strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	if spec.EnvelopeFees != 0.60 {
		t.Errorf("EnvelopeFees = %v, want 0.60", spec.EnvelopeFees)
	}
	if h := spec.Holdings[0]; h.Fees != 0.03 {
		t.Errorf("fees column: %+v", h)
	}
	if h := spec.Holdings[1]; h.Fees != -1 {
		t.Errorf("absent fees: %+v", h)
	}
	// Fees out of range: error.
	if _, err := Parse("t", strings.NewReader("60 VOO 25 # note")); err == nil {
		t.Error("expected error for 25 %/year fees")
	}
	// Decimal point (default convention) with a % suffix, plus a comment.
	sp2bis, err := Parse("t", strings.NewReader("100 VOO 0.25% # note"))
	if err != nil || sp2bis.Holdings[0].Fees != 0.25 {
		t.Errorf("fees 0.25%%: %+v, %v", sp2bis.Holdings, err)
	}
	// Decimal comma and %% suffix accepted in the 3rd column.
	sp2, err := Parse("t", strings.NewReader("100 VOO 0,25%"))
	if err != nil || sp2.Holdings[0].Fees != 0.25 {
		t.Errorf("fees 0,25%%: %+v, %v", sp2.Holdings, err)
	}
	// A 3rd column that is neither a number nor a "#" comment = error.
	if _, err := Parse("t", strings.NewReader("100 VOO 3a-long-term goal")); err == nil {
		t.Error("expected error for a non-numeric, un-commented 3rd column")
	}
	// Accepted synonym.
	sp, err := Parse("t", strings.NewReader("#meta envelope-fees:1"+"\n"+"100 VOO"))
	if err != nil || sp.EnvelopeFees != 1 {
		t.Errorf("envelope-fees: %+v, %v", sp, err)
	}
	// The old French key no longer exists: unknown key = plain warning.
	sp, err = Parse("t", strings.NewReader("#meta frais:1"+"\n"+"100 VOO"))
	if err != nil || sp.EnvelopeFees != -1 || len(sp.Warnings) != 1 {
		t.Errorf("frais must be an unknown key: %+v, %v", sp, err)
	}
}

func TestSimulateEnvelopeFees(t *testing.T) {
	n := 253 // ~1 trading year
	p := &Portfolio{
		Name:         "t",
		EnvelopeFees: 2.52, // 0.01 %/trading day
		Assets: []Asset{
			{Symbol: "A", Weight: 1, Series: constSeries("A", 0, n, 100)},
		},
	}
	sim, err := Simulate(p, 0)
	if err != nil {
		t.Fatal(err)
	}
	want := 100 * math.Pow(1-0.0001, float64(n-1))
	got := sim.Values[len(sim.Values)-1]
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("final value with envelope fees: %v, want %v", got, want)
	}
}

func TestParseLeverage(t *testing.T) {
	// Without leverage:on, sum > 100 with a weight > 100: error with a hint.
	if _, err := Parse("t", strings.NewReader("150 SPY")); err == nil || !strings.Contains(err.Error(), "leverage:on") {
		t.Errorf("expected error with leverage hint, got: %v", err)
	}
	// With leverage:on: weights kept as written (fractions of capital).
	spec, err := Parse("t", strings.NewReader("#meta leverage:on\n#meta borrow-spread:0.5\n90 SPY\n60 IEF"))
	if err != nil {
		t.Fatal(err)
	}
	if !spec.Leverage || spec.BorrowSpread != 0.5 {
		t.Errorf("leverage meta: %+v", spec)
	}
	if spec.Holdings[0].Weight != 0.90 || spec.Holdings[1].Weight != 0.60 {
		t.Errorf("non-normalized weights expected: %+v", spec.Holdings)
	}
	if len(spec.Warnings) != 1 || !strings.Contains(spec.Warnings[0], "total exposure 150") {
		t.Errorf("exposure warning expected: %v", spec.Warnings)
	}
	// Cap.
	if _, err := Parse("t", strings.NewReader("#meta leverage:on\n400 SPY\n200 IEF")); err == nil {
		t.Error("expected 500 % cap error")
	}
	// Invalid value.
	if _, err := Parse("t", strings.NewReader("#meta leverage:maybe\n100 SPY")); err == nil {
		t.Error("expected invalid leverage error")
	}
}

func TestSimulateLeverage(t *testing.T) {
	n := 253
	flat := constSeries("A", 0, n, 100)
	rate := &marketdata.Series{Symbol: "^IRX"}
	for i := range n {
		rate.Points = append(rate.Points, marketdata.Point{Date: day(i), Close: 2.52}) // 0.01 %/day
	}
	// 150 % of a flat asset, financed at 2.52 % + 2.52 % spread: the debt of
	// 50 compounds at 0.02 %/day, eroding the NAV by as much.
	p := &Portfolio{
		Name: "t", Leverage: true, BorrowSpread: 2.52, Cash: rate,
		Assets: []Asset{{Symbol: "A", Weight: 1.5, Series: flat}},
	}
	sim, err := Simulate(p, 0)
	if err != nil {
		t.Fatal(err)
	}
	wantDebt := 50 * math.Pow(1.0002, float64(n-1))
	want := 150 - wantDebt
	got := sim.Values[len(sim.Values)-1]
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("NAV with financing: %v, want %v", got, want)
	}
	// Amplification without fees: asset +1 %/day at 1.5x leverage, zero rate.
	up := &marketdata.Series{Symbol: "B"}
	v := 100.0
	for i := range 50 {
		up.Points = append(up.Points, marketdata.Point{Date: day(i), Close: v})
		v *= 1.01
	}
	p2 := &Portfolio{Name: "t", Leverage: true, Assets: []Asset{{Symbol: "B", Weight: 1.5, Series: up}}}
	sim2, err := Simulate(p2, 0)
	if err != nil {
		t.Fatal(err)
	}
	want2 := 150*math.Pow(1.01, 49) - 50
	if got2 := sim2.Values[len(sim2.Values)-1]; math.Abs(got2-want2) > 1e-9 {
		t.Errorf("amplification: %v, want %v", got2, want2)
	}
	// Ruin: 3x leverage on a collapsing asset.
	down := &marketdata.Series{Symbol: "C"}
	v = 100.0
	for i := range 60 {
		down.Points = append(down.Points, marketdata.Point{Date: day(i), Close: v})
		v *= 0.95
	}
	p3 := &Portfolio{Name: "t", Leverage: true, Assets: []Asset{{Symbol: "C", Weight: 3, Series: down}}}
	sim3, err := Simulate(p3, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !sim3.Ruined {
		t.Error("ruin should have been detected")
	}
	if len(sim3.Values) >= 60 {
		t.Error("the series should have been truncated")
	}
}

func TestSingle(t *testing.T) {
	spec := Single(" NTSG ")
	if spec.Name != "NTSG" || len(spec.Holdings) != 1 {
		t.Fatalf("spec: %+v", spec)
	}
	h := spec.Holdings[0]
	if h.ID != "NTSG" || h.Weight != 1 || h.RawWeight != 100 {
		t.Errorf("holding: %+v", h)
	}
}

func TestParseErrors(t *testing.T) {
	for _, in := range []string{
		"", // empty
		"# only comments",
		"VOO",     // no weight
		"abc VOO", // non-numeric weight
		"0 VOO",   // zero weight
		"150 VOO", // weight > 100
		"60",      // no identifier
	} {
		if _, err := Parse("t", strings.NewReader(in)); err == nil {
			t.Errorf("expected error for %q", in)
		}
	}
}
