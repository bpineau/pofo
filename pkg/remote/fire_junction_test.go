package remote

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

func TestPortfolioFIREExcludesDefinitionJump(t *testing.T) {
	src := fixtureSource()
	s := src["AAA"]
	cpi := &marketdata.Series{Symbol: "^HICP-FR"}
	for i := range s.Points {
		s.Points[i].Date = time.Date(2018, time.Month(i+2), 0, 0, 0, 0, 0, time.UTC)
		cpi.Points = append(cpi.Points, marketdata.Point{Date: s.Points[i].Date, Close: 100})
		if i >= 30 {
			s.Points[i].Close *= 10
		}
	}
	s.Junctions = []time.Time{s.Points[30].Date}
	src[cpi.Symbol] = cpi
	r, err := New(src, "test").Execute(context.Background(), "simulate_fire", []byte(`{"plan":{"capital":1000,"annual_spending":40,"horizon_years":2},"market":{"mode":"portfolio","holdings":[{"id":"AAA","weight_pct":100}]},"paths":100}`))
	if err != nil {
		t.Fatal(err)
	}
	if r.Meta.EffectiveWindow.From != "2020-08-01" {
		t.Fatalf("definition jump entered the FIRE history: effective window %+v", r.Meta.EffectiveWindow)
	}
	found := false
	for _, w := range r.Warnings {
		found = found || strings.Contains(w.Message, "definition")
	}
	if !found {
		t.Fatal("excluded definition change was not reported")
	}
}
