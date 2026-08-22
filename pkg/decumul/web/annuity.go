package web

import (
	"fmt"
	"strings"

	"github.com/bpineau/pofo/pkg/decumul"
	"github.com/bpineau/pofo/pkg/metrics"
)

// The page offers exactly one annuity, and prices it once: a joint-life,
// inflation-linked immediate annuity, bought in one go out of the growth
// sleeve at a chosen year of the plan.
//
// WHERE IT LIVES. An annuity has no price without an age and nothing to insure
// without a drawn lifespan, so decumul ignores Plan.Annuity on a plan with no
// Lifetime. That is not a limitation to work around: under a fixed horizon
// every household is certain to reach the end, so a lifelong income is simply
// paid for longer than it was priced for, and the page's old control (a
// Cashflow paying for ever, capital reduced by the premium, no tax on the
// sale) manufactured exactly that free lunch. The annuity is therefore bought
// and paid inside the mortality kernel only, which is the lifecycle view
// (section 05): the premium really pays its capital-gains tax on the way out,
// the income really stops with the covered lives, and the trade-off the
// literature describes (less risk of outliving the money, a smaller estate)
// is measured rather than assumed. The fixed-horizon sections are left
// untouched by the control, and the view says so.
const (
	// annuityRealRate is the real rate the insurer is assumed to price at. A
	// point of real yield is roughly what an index-linked government curve has
	// offered over the long run, and it is the rate the page has quoted since
	// the control existed.
	annuityRealRate = 0.01
	// defaultAnnuityLoad is the insurer's margin as a share of the fair
	// income, applied when none is posted. Retail annuity loads are
	// double-digit; 10 % is the conventional planning figure and the honest
	// default here. It matters that this is not 0: a link written before the
	// load control existed carries no margin at all, and must not quietly buy
	// an actuarially fair annuity nobody sells.
	defaultAnnuityLoad = 0.10
	// maxAnnuityLoad bounds a hand-written margin at something still meaning
	// "an expensive product" rather than a confiscation.
	maxAnnuityLoad = 0.50
)

// annuitantLaw is the mortality the INSURER prices with, not the household's.
// Annuitants are self-selected: people who buy lifelong income expect to
// collect it, and French insurers price on the generational TGH/TGF-05 tables,
// which run materially longer than an INSEE period table. A modal age of 92
// with a 9-year dispersion sits about three to four years of remaining life
// expectancy above the bundled FrenchMortality (88/10), which is the order of
// magnitude the annuitant-versus-population gap has in the published tables.
// It is a fixed choice, deliberately: folding it into the load would conflate
// two costs that move differently with age, and offering it as a control would
// ask a reader to price a table.
var annuitantLaw = decumul.Gompertz{Mode: 92, Dispersion: 9}

// annuity translates the rail's three annuity controls into the kernel's
// product, or nil when the control is off (share 0, the default) or the plan
// has no room for the purchase. The purchase year is clamped into the plan:
// the page models the post-retirement phase only, so year 0 is the retirement
// date and there is no buying before it, nor in a year the plan never reaches.
func (pr Params) annuity() *decumul.Annuity {
	if pr.AnnuityShare <= 0 || pr.Years <= 0 {
		return nil
	}
	share := min(pr.AnnuityShare, 1)
	year := max(pr.AnnuityYear, 0)
	if year >= pr.Years {
		year = pr.Years - 1
	}
	return &decumul.Annuity{
		Year: year, Share: share, Rate: annuityRealRate,
		Load: pr.annuityLoad(), Joint: true, Law: annuitantLaw,
	}
}

// annuityLoad resolves the insurer's margin, treating the unset zero as the
// default margin rather than as a fair annuity (see defaultAnnuityLoad).
func (pr Params) annuityLoad() float64 {
	if pr.AnnuityLoad <= 0 {
		return defaultAnnuityLoad
	}
	return min(pr.AnnuityLoad, maxAnnuityLoad)
}

// annuityAge is the age the annuity is bought at: the retirement age plus the
// purchase year.
func (pr Params) annuityAge(a *decumul.Annuity) float64 {
	return pr.age() + float64(a.Year)
}

// annuityCards is the compact before-and-after readout of the purchase, on the
// two figures an annuity actually trades: the risk of outliving the money and
// what is left behind. Both ensembles are the same mortality kernel on the
// same draws, one with the annuity and one without, so each pair is a paired
// difference and the arrows carry no Monte-Carlo confusion.
//
// The direction is not assumed. A fairly priced annuity bought late buys the
// risk down and pays for it out of the estate; the same product bought at 45,
// where a lifetime of payments makes it pay barely 2.5 % against a growth
// sleeve earning far more, makes both readings worse. The readout states
// whichever happened.
func annuityCards(with, without decumul.Ensemble, pr Params, a *decumul.Annuity) []Card {
	wl, pl := with.LifeOutcome(), without.LifeOutcome()
	income, premium := annuityBought(with, pr, a)
	payout, wr := 0.0, 0.0
	if premium > 0 {
		payout = income / premium
	}
	if pr.Capital > 0 {
		wr = pr.NeedAnnual / pr.Capital
	}
	needShare := ""
	if pr.NeedAnnual > 0 {
		needShare = fmt.Sprintf(", %.0f%% of the need", 100*income/pr.NeedAnnual)
	}
	return []Card{
		{Label: "Annuity · broke while alive",
			Value: fmt.Sprintf("%.1f%% → %.1f%%", pl.RuinAlive*100, wl.RuinAlive*100),
			Help: fmt.Sprintf("Share of households that run out with somebody still alive, without the annuity and with it, on the same futures and the same drawn deaths. The annuity converts %.0f%% of the growth sleeve at age %.0f into an income that cannot run out; whether that lowers the risk is decided by the card on the right. Note the plan stops at your horizon, so payments the annuity would still make after it are neither collected nor counted: a longer horizon shows more of what it buys.",
				a.Share*100, pr.annuityAge(a))},
		{Label: "Annuity · median estate",
			Value: wealthPair(pl.EstateP50, wl.EstateP50),
			Help:  "Real wealth left at the household's own end, without the annuity and with it. Insurance is paid for out of the bequest: the premium leaves the portfolio for good, and what it buys stops the day the last covered life does."},
		{Label: "Annuity · pays vs withdraws",
			Value: fmt.Sprintf("%.1f%% vs %.1f%%", payout*100, wr*100),
			Help: fmt.Sprintf("The whole trade in one ratio: the annuity pays %.1f k€/yr for life%s, %.1f%% of its %s premium, where the plan withdraws %.1f%% of capital. Pay more than the plan withdraws and the insurance is bought cheaply; pay less, as here the quote often does, and the guaranteed income costs more than the risk it removes. The quote assumes a %.0f%% real rate, a %.0f%% insurer margin and an annuitant table (buyers of lifelong income outlive the general population, so the same premium buys less than your own mortality would suggest); on top of that the sale that raises the premium pays its capital-gains tax.",
				income/1000, needShare, payout*100, fmtWealth(premium), wr*100, annuityRealRate*100, pr.annuityLoad()*100)},
	}
}

// wealthPair renders a before-and-after wealth reading in one line, carrying
// the unit once when both figures share it ("601 → 290 k€") and twice when
// they do not. A card is a small box and a repeated unit is the first thing
// that wraps it onto a second line.
func wealthPair(before, after float64) string {
	b, a := fmtWealth(before), fmtWealth(after)
	if i, j := strings.LastIndex(b, " "), strings.LastIndex(a, " "); i > 0 && j > 0 && b[i:] == a[j:] {
		return b[:i] + " → " + a
	}
	return b + " → " + a
}

// annuityBought is the median income the purchase actually bought and the
// median premium it cost, read off the paths that bought it. At a purchase in
// year 0 both are deterministic; later they vary with the sleeve's value on
// the day, which is why the median is taken rather than a formula evaluated on
// the starting capital.
func annuityBought(e decumul.Ensemble, pr Params, a *decumul.Annuity) (income, premium float64) {
	premiums := make([]float64, 0, len(e.Paths))
	for _, p := range e.Paths {
		if p.Premium > 0 {
			premiums = append(premiums, p.Premium)
		}
	}
	if q := metrics.Quantiles(premiums, 0.50); len(q) > 0 {
		premium = q[0]
	}
	// Joint on one law over a same-age couple is exactly what the kernel
	// prices (Annuity.survivalFrom), so AnnuityIncome quotes the same product.
	income = decumul.AnnuityIncome(annuitantLaw, pr.annuityAge(a), premium, a.Rate, 1-pr.annuityLoad())
	return income, premium
}
