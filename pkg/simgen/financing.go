package simgen

import (
	"fmt"
	"os"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// usdOvernight is the identifier under which financed() serves the USD
// overnight rate: the secured overnight financing rate (^SOFR) from its
// 2018-04 start, the effective federal funds rate (^FEDFUNDS, FRED daily from
// 1954-07) before it, and the 13-week T-bill rate (^IRX, itself extended by
// TBILL-3M) before that.
//
// It is what a FUTURES-BASED leg pays, and it is not what a leg holding bills
// earns. An index future prices off the implied repo, which is the overnight
// rate plus a roll richness; a T-bill is a scarce collateral asset that trades
// through it, and Yahoo's ^IRX quotes that bill on a DISCOUNT basis, ~0.1
// point under its own investment yield at a 5 % level. Financing an overlay at
// ^IRX therefore borrows too cheaply twice over. Measured on each fund's own
// live window, ^IRX runs 0.01 to 0.15 points a year under the overnight rate
// (0.34 over 2000-2018), and every capital-efficient reconstruction in this
// file was hot by roughly that much on top of its other errors; see
// docs/trend-reconstruction-design.md for the per-fund table.
//
// Two rules keep the change confined to financing:
//
//   - a leg that genuinely HOLDS bills keeps earning ^IRX. The 10 % collateral
//     sleeve of an efficient-core fund and the collateral of a managed-futures
//     programme are cash positions, not financing costs.
//   - a gearing that stands in for DURATION (zrozRecipe, iefRecipe, shyRecipe)
//     keeps ^IRX too. Its cash term is an arithmetic residue of writing
//     g × bond as cash + g × (bond − cash), not money anyone borrows.
const usdOvernight = "^USD-OVERNIGHT"

// financed wraps a Fetcher so that usdOvernight resolves to the spliced
// overnight rate. The series is built on ^IRX's own trading calendar and
// carries the overnight LEVEL on those dates: the federal funds series is
// published for every calendar day, and Align takes the UNION of its inputs'
// dates, so serving it raw would inject weekends into every frame it joins and
// hand a /252 accrual convention 365 days a year.
//
// When neither policy series answers, the financing falls back to ^IRX and the
// build says so on stderr rather than failing: a reconstruction that reaches
// 1871 must not depend on a rate feed that starts in 1954.
func financed(f Fetcher) Fetcher { return financingFetcher{inner: f} }

type financingFetcher struct{ inner Fetcher }

func (ff financingFetcher) Fetch(id string, from time.Time) (*marketdata.Series, error) {
	if id != usdOvernight {
		return ff.inner.Fetch(id, from)
	}
	bills, err := ff.inner.Fetch("^IRX", from)
	if err != nil {
		return nil, fmt.Errorf("%s: bill rate: %w", usdOvernight, err)
	}
	if bills == nil || len(bills.Points) < 2 {
		return nil, fmt.Errorf("%s: bill rate: empty history", usdOvernight)
	}
	overnight := ff.overnight(from)
	out := &marketdata.Series{
		Symbol: usdOvernight,
		Name:   "USD overnight financing rate (SOFR, effective fed funds, T-bill before)",
		Source: "simdata",
		Points: make([]marketdata.Point, len(bills.Points)),
	}
	for i, p := range bills.Points {
		out.Points[i] = p
		if v, _, ok := overnight.At(p.Date); ok {
			out.Points[i].Close = v
		}
	}
	// Say where the overnight era starts. It depends on which sources answered,
	// so a build that quietly financed a whole extra era at the bill rate would
	// otherwise ship a different file without saying so.
	if len(overnight.Points) > 0 {
		fmt.Fprintf(os.Stderr, "financing: %s: overnight from %s, bill rate before\n",
			usdOvernight, overnight.First().Date.Format("2006-01-02"))
	}
	return out, nil
}

// overnight splices ^SOFR over ^FEDFUNDS. Both are annualized percent levels,
// so the splice is a plain concatenation at the junction date: a rate is a
// rate, and nothing needs rescaling to it. A source that does not answer is
// reported and skipped, leaving the bill rate to cover its era.
func (ff financingFetcher) overnight(from time.Time) *marketdata.Series {
	out := &marketdata.Series{}
	sofr, serr := ff.inner.Fetch("^SOFR", from)
	eff, eerr := ff.inner.Fetch("^FEDFUNDS", from)
	if eerr != nil || eff == nil || len(eff.Points) < 2 {
		fmt.Fprintf(os.Stderr, "financing: %s: ^FEDFUNDS unavailable (%v), the bill rate stands before SOFR\n", usdOvernight, eerr)
	} else {
		out.Points = append(out.Points, eff.Points...)
	}
	if serr != nil || sofr == nil || len(sofr.Points) < 2 {
		fmt.Fprintf(os.Stderr, "financing: %s: ^SOFR unavailable (%v), the funds rate stands to the end\n", usdOvernight, serr)
		return out
	}
	cut := sofr.First().Date
	kept := out.Points[:0]
	for _, p := range out.Points {
		if p.Date.Before(cut) {
			kept = append(kept, p)
		}
	}
	out.Points = append(kept, sofr.Points...)
	return out
}
