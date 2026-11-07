package marketdata

import (
	"fmt"
	"slices"
)

// DataWarnings returns a detached, deduplicated copy of Warnings plus known
// data qualifications implied by Source, Name and FullySimulated. In
// particular, Stooq closes and distributing FT/Morningstar NAVs omit income.
// It does not infer whether an unknown provider adjusts distributions.
//
// Use this when constructing a Series directly. Transformations, returns
// panels and CSV exports materialize these qualifications into Warnings so
// replacing or omitting Source cannot erase them. Analyze studies add their
// window-specific reconstruction, nowcast and definition-change warnings.
// A nil receiver has no warnings; calling this method never mutates it.
func (s *Series) DataWarnings() []string {
	if s == nil {
		return nil
	}
	var out []string
	for _, w := range s.Warnings {
		out = appendWarning(out, w)
	}
	switch s.Source {
	case "stooq":
		out = appendWarning(out, "Stooq closes are not dividend-adjusted: the income is missing from every statistic")
	case "ft", "morningstar":
		if LooksDistributing(s.Name) {
			out = appendWarning(out, fmt.Sprintf("distributing share class quoted as a NAV (%s): the income it pays out is missing from every statistic", s.Source))
		}
	}
	if s.FullySimulated {
		out = appendWarning(out, "entire history is reconstructed, not real quotes")
	}
	return out
}

func appendWarning(dst []string, warning string) []string {
	if !slices.Contains(dst, warning) {
		return append(dst, warning)
	}
	return dst
}
