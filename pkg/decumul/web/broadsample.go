package web

import (
	"github.com/bpineau/pofo/pkg/fire"
	"github.com/bpineau/pofo/pkg/scenario"
)

type countrySeries = fire.CountrySeries

func broadSampleCountries() []fire.CountrySeries  { return fire.BroadSampleCountries() }
func broadSampleEquity() [][]float64              { return fire.BroadSampleEquity() }
func broadSampleMixed() [][]float64               { return fire.BroadSampleMixed() }
func broadSampleSource(years int) scenario.Source { return fire.BroadSampleSource(years) }
