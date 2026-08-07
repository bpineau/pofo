package main

import (
	"reflect"
	"testing"

	"github.com/bpineau/pofo/pkg/portfolio"
)

func TestEffectiveCurrencies(t *testing.T) {
	cases := []struct {
		name string
		spec *portfolio.Spec
		def  string
		want []string
	}{
		{"default", &portfolio.Spec{}, "EUR", []string{"EUR"}},
		{"declared", &portfolio.Spec{Currencies: []string{"USD", "EUR"}}, "EUR", []string{"USD", "EUR"}},
	}
	for _, c := range cases {
		if got := effectiveCurrencies(c.spec, c.def); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: effectiveCurrencies = %v, want %v", c.name, got, c.want)
		}
	}
}
