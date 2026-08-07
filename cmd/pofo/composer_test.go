package main

import (
	"reflect"
	"strings"
	"testing"

	"github.com/bpineau/pofo/pkg/portfolio"
)

func parseSpec(t *testing.T, name, text string) *portfolio.Spec {
	t.Helper()
	spec, err := portfolio.Parse(name, strings.NewReader(text))
	if err != nil {
		t.Fatal(err)
	}
	return spec
}

// A catalog-only spec round-trips through the p= grammar: same holdings,
// weights, name and carried metas.
func TestSpecToPRoundTrip(t *testing.T) {
	spec := parseSpec(t, "my dragon", "#meta sim:on\n#meta rebalance:30\n60 IWDA\n40 IGLN\n")
	p, dropped := specToP(spec)
	if len(dropped) != 0 {
		t.Fatalf("dropped = %v, want none", dropped)
	}
	back, err := adhocSpec(p, 1)
	if err != nil {
		t.Fatalf("adhocSpec(%q): %v", p, err)
	}
	if back.Name != "my dragon" || !back.Sim || back.RebalanceDays != 30 {
		t.Errorf("round trip lost metas: name=%q sim=%v rebalance=%d", back.Name, back.Sim, back.RebalanceDays)
	}
	if len(back.Holdings) != 2 || back.Holdings[0].ID != "IWDA" || back.Holdings[0].RawWeight != 60 ||
		back.Holdings[1].ID != "IGLN" || back.Holdings[1].RawWeight != 40 {
		t.Errorf("round trip holdings = %+v", back.Holdings)
	}
}

// Non-local ids and page-shape metas are dropped and reported.
//
// optimize and currencies never co-occur in a real file (portfolio.Parse
// rejects the pair), so both keys are seeded into Meta directly: specToP must
// drop whichever page-shape directive it finds, and this asserts both paths.
func TestSpecToPDrops(t *testing.T) {
	spec := parseSpec(t, "mixed", "#meta sim:on\n50 IWDA\n50 ZZZNOTLOCAL\n")
	spec.Meta["optimize"] = "max-sharpe"
	spec.Meta["currencies"] = "USD,EUR"
	p, dropped := specToP(spec)
	want := []string{"ZZZNOTLOCAL", "currencies", "optimize"}
	if !reflect.DeepEqual(dropped, want) {
		t.Errorf("dropped = %v, want %v", dropped, want)
	}
	if strings.Contains(p, "ZZZNOTLOCAL") || strings.Contains(p, "optimize") || strings.Contains(p, "currencies") {
		t.Errorf("p carries dropped content: %q", p)
	}
	if !strings.Contains(p, "!sim:on") || !strings.Contains(p, "IWDA:50") {
		t.Errorf("p lost kept content: %q", p)
	}
}

// A meta value containing "!" would break the segment grammar: dropped.
func TestSpecToPBangValue(t *testing.T) {
	spec := parseSpec(t, "odd", "#meta note:a!b\n100 IWDA\n")
	p, dropped := specToP(spec)
	if !reflect.DeepEqual(dropped, []string{"note"}) {
		t.Errorf("dropped = %v, want [note]", dropped)
	}
	if strings.Contains(p, "note") {
		t.Errorf("p carries the bang meta: %q", p)
	}
}

// A name that carries a "!" cannot ride the "!"-delimited grammar: no name
// segment is emitted and "name" is reported as dropped.
func TestSpecToPBangName(t *testing.T) {
	spec := parseSpec(t, "boom!bang", "100 IWDA\n")
	p, dropped := specToP(spec)
	if !reflect.DeepEqual(dropped, []string{"name"}) {
		t.Errorf("dropped = %v, want [name]", dropped)
	}
	if strings.Contains(p, "name:") {
		t.Errorf("p carries the bang name: %q", p)
	}
}

// Fractional weights survive without trailing-zero noise.
func TestSpecToPFractionalWeight(t *testing.T) {
	spec := parseSpec(t, "frac", "12.5 IWDA\n87.5 IGLN\n")
	p, _ := specToP(spec)
	if !strings.Contains(p, "IWDA:12.5") || !strings.Contains(p, "IGLN:87.5") {
		t.Errorf("fractional weights mangled: %q", p)
	}
}
