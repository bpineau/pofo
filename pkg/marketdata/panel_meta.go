package marketdata

import "fmt"

// mixMetadata retains the qualifications of the legs actually held. The map
// and records are immutable after construction, so selections can share them.
func (p *Panel) mixMetadata(id string, weights map[string]float64) (map[string]*Series, error) {
	meta := make(map[string]*Series, len(p.meta)+1)
	for key, s := range p.meta {
		meta[key] = s
	}
	mix := &Series{Symbol: id, Source: "blend"}
	unknown := false
	for _, key := range p.IDs {
		if weights[key] == 0 {
			continue
		}
		s := p.meta[key]
		if s == nil {
			unknown = true
			continue
		}
		if s.Currency == "" {
			unknown = true
		} else if mix.Currency == "" {
			mix.Currency = s.Currency
		} else if mix.Currency != s.Currency {
			return nil, fmt.Errorf("marketdata: Panel.Mix %s: convert %s and %s to one currency first", id, mix.Currency, s.Currency)
		}
		for _, w := range s.Warnings {
			mix.Warnings = append(mix.Warnings, key+": "+w)
		}
		if s.SimulatedBefore.After(mix.SimulatedBefore) {
			mix.SimulatedBefore = s.SimulatedBefore
		}
		if !s.EstimatedFrom.IsZero() && (mix.EstimatedFrom.IsZero() || s.EstimatedFrom.Before(mix.EstimatedFrom)) {
			mix.EstimatedFrom = s.EstimatedFrom
		}
		if s.FXExtrapolatedBefore.After(mix.FXExtrapolatedBefore) {
			mix.FXExtrapolatedBefore = s.FXExtrapolatedBefore
		}
	}
	if unknown {
		mix.Currency = ""
		mix.Warnings = append(mix.Warnings, "blend includes a leg with unknown currency")
	}
	meta[id] = mix
	return meta, nil
}
