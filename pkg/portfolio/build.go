package portfolio

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// BuildOptions supplies the environment Build needs beyond the parsed spec.
// Fetch is required; everything else is optional.
type BuildOptions struct {
	// Fetch returns the price series of one holding identifier, typically
	// a closure over marketdata.Client.FetchExtended so SIM extension and
	// currency conversion are applied per asset.
	Fetch func(id string) (*marketdata.Series, error)

	// Fees returns an asset's TER in percent per year when the file does
	// not declare one, typically marketdata.Client.Fees on the bare
	// identifier. Nil leaves undeclared TERs unknown (negative), which
	// Simulate accepts: asset TERs are informational, already net in
	// prices.
	Fees func(id string) (float64, bool)

	// Cash is the financing/deposit rate series used when the spec
	// declares leverage, in annualized percent levels (e.g. ^IRX). Nil
	// means a flat 0 % rate.
	Cash *marketdata.Series

	// BorrowSpread is the default spread over the cash rate paid on
	// borrowed money, in percent per year, applied when the spec does not
	// set "#meta borrow-spread". The pofo CLI uses 1.
	BorrowSpread float64

	// BaseCurrency is the currency the fetched series are expected to be
	// quoted in. When set, holdings whose currency is unknown are flagged
	// in Portfolio.Warnings (they cannot have been converted).
	BaseCurrency string
}

// Build resolves a parsed Spec into a Portfolio ready for Simulate: each
// holding's series comes through opt.Fetch, a TER missing from the file is
// looked up via opt.Fees, and the spec's directives (capital, periodic
// flows, leverage, envelope fees) are carried over. Data quality remarks
// (unknown or mixed currencies) accumulate in Portfolio.Warnings. Build
// fails on the first holding Fetch cannot serve.
//
// Two spec fields intentionally stay with the caller: Spec.RebalanceDays
// (pass it, or a default when negative, to Simulate) and Spec.Optimize
// (run pkg/optimize and re-weight a copy if desired).
func Build(spec *Spec, opt BuildOptions) (*Portfolio, error) {
	if opt.Fetch == nil {
		return nil, fmt.Errorf("portfolio %s: BuildOptions.Fetch is required", spec.Name)
	}
	p := &Portfolio{Name: spec.Name, Warnings: slices.Clone(spec.Warnings)}
	if spec.EnvelopeFees > 0 {
		p.EnvelopeFees = spec.EnvelopeFees
	}
	if spec.Capital > 0 {
		p.Capital = spec.Capital
	}
	p.Contribute, p.Withdraw = spec.Contribute, spec.Withdraw
	if spec.Leverage {
		p.Leverage = true
		p.BorrowSpread = spec.BorrowSpread
		if p.BorrowSpread < 0 {
			p.BorrowSpread = opt.BorrowSpread
		}
		p.Cash = opt.Cash
	}
	currencies := map[string]bool{}
	for _, h := range spec.Holdings {
		s, err := opt.Fetch(h.ID)
		if err != nil {
			return nil, fmt.Errorf("portfolio %s, asset %q: %w", spec.Name, h.ID, err)
		}
		fees := h.Fees // the file column takes precedence
		if fees < 0 && opt.Fees != nil {
			if ter, ok := opt.Fees(h.ID); ok {
				fees = ter
			}
		}
		p.Assets = append(p.Assets, Asset{
			ID:     h.ID,
			Symbol: s.Symbol,
			Name:   s.Name,
			Weight: h.Weight,
			Fees:   fees,
			Series: s,
		})
		if s.Currency != "" {
			currencies[s.Currency] = true
		} else if opt.BaseCurrency != "" {
			p.Warnings = append(p.Warnings, fmt.Sprintf(
				"%s: unknown currency, left unconverted", s.Symbol))
		}
	}
	if len(currencies) > 1 {
		list := make([]string, 0, len(currencies))
		for c := range currencies {
			list = append(list, c)
		}
		sort.Strings(list)
		p.Warnings = append(p.Warnings, fmt.Sprintf(
			"mixed currencies (%s), no FX conversion applied", strings.Join(list, ", ")))
	}
	return p, nil
}
