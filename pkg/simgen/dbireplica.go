package simgen

import (
	"fmt"
	"math"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// DBiDonorID names the donor series the DBi family's pre-inception era stands
// on, bundled as refdata by cmd/gen-dbi-refdata: the published net all-styles
// managed-futures composite, half of it re-expressed through the ten futures
// contracts the replicating fund actually holds.
//
// # Why a composite alone is not the best donor for a replication fund
//
// The composite is an average of twenty programmes trading fifty-odd markets.
// The fund is a portfolio of TEN US-listed futures: its own quarterly schedule
// of investments reports exactly the S&P 500 E-mini, MSCI EAFE, MSCI Emerging
// Markets, the 2-year, 10-year and long US Treasury contracts, gold, WTI crude,
// the euro and the yen, and nothing else. Whatever the composite earns on the
// forty markets the fund cannot hold is therefore tracking error by
// construction, and no rescaling of the composite removes it.
//
// DBiProjection removes it in the only way the fund itself does: by regressing
// the composite's own recent daily returns on those ten contracts and holding
// the resulting portfolio. That is the published outline of the fund's process
// (a trailing 60-day window on the target's performance, US-traded contracts
// only, weekly refits, no discretion over the model's output) applied to the
// public index with public price series, not a fit to the fund.
//
// # Why the shipped donor is half of each
//
// The projection is estimated, so it carries estimation noise the composite
// does not; the composite carries exposures the fund does not. Averaging the
// two at equal weight cancels part of both errors, and measurement says the
// combination is worth more than either side alone (see
// docs/trend-reconstruction-design.md for the table). The weight is the
// unsearched midpoint, not an optimum: the measured curve is flat enough that
// a quarter or three quarters still beat the composite everywhere.
const DBiDonorID = "TREND-ALLSTYLES-DBI-USD"

// The engine's published parameters. None of them is fitted here: the window
// and the instrument list are what the fund's own prospectus and schedule of
// investments state, and the weekly refit is what its manager describes.
const (
	dbiLookback  = 60  // trailing days of the regression
	dbiRebalance = 5   // trading days between refits, i.e. weekly
	dbiBlend     = 0.5 // weight of the projection in the shipped donor
)

// dbiLegKind says how a leg's series turns into the excess return of a futures
// position, which is the only quantity a futures portfolio earns.
type dbiLegKind int

const (
	dbiFunded dbiLegKind = iota // a total return: excess = return - cash
	dbiPrice                    // a spot or futures price: the return IS the excess
	dbiFX                       // a currency: the spot move plus the rate differential
)

// dbiLeg is one of the ten contracts, with the series that price it.
//
// Near is a US-listed proxy that closes when the futures contract settles, and
// Deep is what stands behind it before it existed. The order matters at daily
// frequency: a mutual fund's NAV struck on a Tokyo or Frankfurt close is a
// different day from a US exchange settlement, and a regression run on the two
// mixed reads the difference as beta.
type dbiLeg struct {
	Name     string
	Near     string     // listed US-hours proxy, spliced in front
	NearFee  float64    // its published ongoing charge, fraction per year
	Deep     string     // what stands behind it (a fund, a price, or a yield)
	DeepFee  float64    // likewise; a yield reconstruction or a price charges none
	Maturity float64    // when > 0, Deep is a yield series priced as a par bond
	Kind     dbiLegKind //
	Carry    string     // FX only: the foreign money-market accrual index
}

// dbiLegs are the ten contracts the fund's consolidated schedule of investments
// reports, in the order that schedule lists them, each with the deepest series
// this repository can price it with.
//
// Two substitutions are worth stating, both confined to the era before the
// listed proxies exist:
//
//   - the 2-year note is priced off the FIVE-year constant-maturity yield until
//     the 1-3 year Treasury ETF opens in 2002-07, because the 2-year yield is
//     not fetchable through the client and the 5-year one is. What is held is
//     still a two-year par bond, so the leg carries the right duration, and over
//     1999-2002 the two yields' daily changes correlate at 0.92.
//   - the developed-ex-US and emerging legs stand on Vanguard index funds until
//     their ETFs open (2001-08 and 2003-04). Those NAVs are struck on foreign
//     closes, which costs daily precision in the first years and nothing at all
//     from 2003 on.
//
// Every proxy is also given back its own ongoing charge, because a futures
// contract pays none: a fund NAV is net of its manager's fee, and a book that
// is short 2.3 units of Treasury proxies would otherwise EARN those fees.
// Left uncorrected the projection reads about 0.35 %/yr hot, which is the same
// doctrine the rest of this package applies to every stand-in. The charges are
// published ongoing charges, none of them estimated:
//
//	SPY   0.09 %  SPDR S&P 500 ETF Trust           VFINX 0.14 %  Vanguard 500 Investor
//	EFA   0.32 %  iShares MSCI EAFE ETF            VTMGX 0.05 %  Vanguard Developed Markets
//	EEM   0.72 %  iShares MSCI Emerging Markets    VEIEX 0.29 %  Vanguard Emerging Markets
//	SHY   0.15 %  iShares 1-3 Year Treasury Bond   the deep Treasury legs are par-bond
//	IEF   0.15 %  iShares 7-10 Year Treasury Bond  reconstructions from published yields
//	TLT   0.15 %  iShares 20+ Year Treasury Bond   and charge nothing
var dbiLegs = []dbiLeg{
	{Name: "S&P 500", Near: "SPY", NearFee: 0.0009, Deep: "VFINX", DeepFee: vfinxTER, Kind: dbiFunded},
	{Name: "MSCI EAFE", Near: "EFA", NearFee: 0.0032, Deep: "VTMGX", DeepFee: vtmgxTER, Kind: dbiFunded},
	{Name: "MSCI EM", Near: "EEM", NearFee: 0.0072, Deep: "VEIEX", DeepFee: 0.0029, Kind: dbiFunded},
	{Name: "UST 2y", Near: "SHY", NearFee: 0.0015, Deep: "^FVX", Maturity: 2, Kind: dbiFunded},
	{Name: "UST 10y", Near: "IEF", NearFee: 0.0015, Deep: "^TNX", Maturity: 10, Kind: dbiFunded},
	{Name: "UST 30y", Near: "TLT", NearFee: 0.0015, Deep: "^TYX", Maturity: 30, Kind: dbiFunded},
	{Name: "Gold", Deep: "GC=F", Kind: dbiPrice},
	{Name: "WTI crude", Deep: "CL=F", Kind: dbiPrice},
	{Name: "EUR", Deep: "EURUSD=X", Kind: dbiFX, Carry: "EUR"},
	{Name: "JPY", Deep: "JPYUSD=X", Kind: dbiFX, Carry: "JPCASH-JPY"},
}

// DBiReplication builds the donor described at DBiDonorID: the composite
// blended with its own projection onto the fund's ten contracts.
//
// It starts sixty trading days after the composite does, which is the
// regression's warm-up and the only history it costs; the chain's next donor
// covers what is behind it, as it always did.
func DBiReplication(f Fetcher, from time.Time) (*marketdata.Series, error) {
	proj, ref, err := dbiProjection(f, from)
	if err != nil {
		return nil, err
	}
	out := blendReturns("all-styles composite, half projected on the fund's ten futures", ref, proj, dbiBlend)
	if len(out.Points) < 250 {
		return nil, fmt.Errorf("%s: only %d days blended", DBiDonorID, len(out.Points))
	}
	return out, nil
}

// DBiProjection is the projection alone, without the composite it is blended
// with: what a portfolio of the fund's ten contracts, sized by the composite's
// own trailing betas, would have earned. It is exported for the generator's
// validation report, which grades the two sides separately.
func DBiProjection(f Fetcher, from time.Time) (*marketdata.Series, error) {
	proj, _, err := dbiProjection(f, from)
	return proj, err
}

func dbiProjection(f Fetcher, from time.Time) (proj, ref *marketdata.Series, err error) {
	ef := extend(f)
	ref, err = f.Fetch(allStylesIndex, from)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: reference %s: %w", DBiDonorID, allStylesIndex, err)
	}
	cash, err := ef.Fetch("^IRX", from)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: cash: %w", DBiDonorID, err)
	}
	eur, err := eurOvernightDeep(ef, from)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: euro cash: %w", DBiDonorID, err)
	}

	// The calendar is the composite's own publication days: it is the series
	// being regressed, and a day it does not publish carries no information.
	dates := make([]time.Time, 0, len(ref.Points))
	for _, p := range ref.Points {
		if !p.Date.Before(from) {
			dates = append(dates, p.Date)
		}
	}
	if len(dates) < dbiLookback+250 {
		return nil, nil, fmt.Errorf("%s: reference too short (%d days)", DBiDonorID, len(dates))
	}

	cashRet := make([]float64, len(dates))
	for i := 1; i < len(dates); i++ {
		cashRet[i] = cashAccrual(cash, dates[i-1], dates[i])
	}

	x := make([][]float64, len(dbiLegs))
	for i, leg := range dbiLegs {
		s, cut, lerr := dbiLegSeries(ef, leg, from)
		if lerr != nil {
			return nil, nil, lerr
		}
		var carry *marketdata.Series
		if leg.Kind == dbiFX {
			carry = eur
			if leg.Carry != "EUR" {
				if carry, lerr = cashRefdata(ef, leg.Carry, leg.Name+" cash", from); lerr != nil {
					return nil, nil, fmt.Errorf("%s: %s carry: %w", DBiDonorID, leg.Name, lerr)
				}
			}
		}
		x[i] = dbiExcess(leg, s, carry, cut, dates, cashRet)
	}

	y := make([]float64, len(dates))
	for i := 1; i < len(dates); i++ {
		if ref.Points[0].Close > 0 {
			a, _, okA := ref.At(dates[i-1])
			b, _, okB := ref.At(dates[i])
			if okA && okB && a > 0 {
				y[i] = b/a - 1 - cashRet[i]
			}
		}
	}

	proj = &marketdata.Series{Name: "all-styles composite projected on the fund's ten futures", Source: "simdata"}
	val := 100.0
	first := dbiLookback + 1
	proj.Points = append(proj.Points, marketdata.Point{Date: dates[first], Close: val})
	betas := make([]float64, len(dbiLegs))
	since := dbiRebalance
	for k := first + 1; k < len(dates); k++ {
		if since >= dbiRebalance {
			since = 0
			if b, ok := dbiFit(y, x, k); ok {
				betas = b
			}
		}
		since++
		r := cashRet[k]
		for i := range betas {
			r += betas[i] * x[i][k]
		}
		val *= 1 + r
		proj.Points = append(proj.Points, marketdata.Point{Date: dates[k], Close: val})
	}
	return proj, ref, nil
}

// dbiFit regresses the reference's trailing dbiLookback excess returns on the
// legs', at date index k, and returns the slopes. The intercept is fitted and
// then DISCARDED, which is the whole point of replicating a replicator: what
// the regression cannot explain is the constituent managers' own skill and
// their fees, and a portfolio of ten futures earns neither.
func dbiFit(y []float64, x [][]float64, k int) ([]float64, bool) {
	n := len(x)
	p := n + 1 // the intercept
	xtx := make([][]float64, p)
	for i := range xtx {
		xtx[i] = make([]float64, p)
	}
	xty := make([]float64, p)
	row := make([]float64, p)
	row[n] = 1
	for t := k - dbiLookback; t < k; t++ {
		for i := 0; i < n; i++ {
			row[i] = x[i][t]
		}
		for i := 0; i < p; i++ {
			xty[i] += row[i] * y[t]
			for j := 0; j < p; j++ {
				xtx[i][j] += row[i] * row[j]
			}
		}
	}
	beta, ok := solveSymmetric(xtx, xty)
	if !ok {
		return nil, false
	}
	return beta[:n], true
}

// solveSymmetric solves A x = b by Gaussian elimination with partial pivoting.
// A singular system means a leg stopped moving, and the caller keeps the
// positions it already had rather than trading on a broken fit.
func solveSymmetric(a [][]float64, b []float64) ([]float64, bool) {
	n := len(b)
	m := make([][]float64, n)
	for i := range m {
		m[i] = append(append(make([]float64, 0, n+1), a[i]...), b[i])
	}
	for c := 0; c < n; c++ {
		p := c
		for r := c + 1; r < n; r++ {
			if math.Abs(m[r][c]) > math.Abs(m[p][c]) {
				p = r
			}
		}
		if math.Abs(m[p][c]) < 1e-18 {
			return nil, false
		}
		m[c], m[p] = m[p], m[c]
		for r := 0; r < n; r++ {
			if r == c {
				continue
			}
			k := m[r][c] / m[c][c]
			for j := c; j <= n; j++ {
				m[r][j] -= k * m[c][j]
			}
		}
	}
	x := make([]float64, n)
	for i := range x {
		x[i] = m[i][n] / m[i][i]
	}
	return x, true
}

// dbiLegSeries prices one leg over the whole period: the listed proxy from its
// own first quote, the deep series (a fund, a price, or a par bond built from a
// yield) before it, joined at the junction level.
// It also returns the junction date, from which the leg carries the listed
// proxy's own ongoing charge rather than the deep one's.
func dbiLegSeries(f Fetcher, leg dbiLeg, from time.Time) (*marketdata.Series, time.Time, error) {
	deep, err := f.Fetch(leg.Deep, from)
	if err != nil || deep == nil || len(deep.Points) < 250 {
		return nil, time.Time{}, fmt.Errorf("%s: leg %s: %s: %w", DBiDonorID, leg.Name, leg.Deep, err)
	}
	if leg.Maturity > 0 {
		deep = TreasuryTR(leg.Name+" (par bond on "+leg.Deep+")", deep, leg.Maturity, 0)
	}
	if leg.Near == "" {
		return deep, time.Time{}, nil
	}
	near, err := f.Fetch(leg.Near, from)
	if err != nil || near == nil || len(near.Points) < 250 {
		return nil, time.Time{}, fmt.Errorf("%s: leg %s: %s: %w", DBiDonorID, leg.Name, leg.Near, err)
	}
	out := &marketdata.Series{Name: leg.Near + " over " + leg.Deep, Source: "simdata"}
	cut := near.First().Date
	val := 100.0
	for i, p := range deep.Points {
		if !p.Date.Before(cut) {
			break
		}
		if i > 0 && deep.Points[i-1].Close > 0 {
			val *= p.Close / deep.Points[i-1].Close
		}
		out.Points = append(out.Points, marketdata.Point{Date: p.Date, Close: val})
	}
	for i, p := range near.Points {
		if i > 0 && near.Points[i-1].Close > 0 {
			val *= p.Close / near.Points[i-1].Close
		}
		out.Points = append(out.Points, marketdata.Point{Date: p.Date, Close: val})
	}
	return out, cut, nil
}

// dbiExcess turns a leg's levels into the excess return a fully collateralized
// futures position earns on each of the reference's days. A funded total return
// gives up the cash it earned; a price return is already an excess; a currency
// adds the interest differential its forward carries, which is what a currency
// future pays beyond the spot move. Whatever ongoing charge the proxy levies
// inside its own NAV is added back, a futures contract levying none.
func dbiExcess(leg dbiLeg, s, carry *marketdata.Series, cut time.Time, dates []time.Time, cashRet []float64) []float64 {
	out := make([]float64, len(dates))
	for i := 1; i < len(dates); i++ {
		a, _, okA := s.At(dates[i-1])
		b, _, okB := s.At(dates[i])
		if !okA || !okB || a <= 0 {
			continue
		}
		r := b/a - 1
		fee := leg.DeepFee
		if !cut.IsZero() && !dates[i].Before(cut) {
			fee = leg.NearFee
		}
		years := dates[i].Sub(dates[i-1]).Hours() / 24 / 365.25
		switch leg.Kind {
		case dbiFunded:
			r += fee*years - cashRet[i]
		case dbiFX:
			r += eurCashReturn(carry, dates[i-1], dates[i]) - cashRet[i]
		}
		out[i] = r
	}
	return out
}

// blendReturns compounds a weighted average of two series' daily returns over
// the dates they share, starting where the second one does. It is the plain
// combination of two estimates of the same quantity, not a rebalanced
// portfolio of two assets: what is averaged is the return, day by day.
func blendReturns(name string, a, b *marketdata.Series, w float64) *marketdata.Series {
	byDate := make(map[time.Time]float64, len(a.Points))
	for i := 1; i < len(a.Points); i++ {
		if a.Points[i-1].Close > 0 {
			byDate[a.Points[i].Date] = a.Points[i].Close/a.Points[i-1].Close - 1
		}
	}
	out := &marketdata.Series{Name: name, Source: "simdata"}
	if len(b.Points) == 0 {
		return out
	}
	val := 100.0
	out.Points = append(out.Points, marketdata.Point{Date: b.First().Date, Close: val})
	for i := 1; i < len(b.Points); i++ {
		ra, ok := byDate[b.Points[i].Date]
		if !ok || b.Points[i-1].Close <= 0 {
			continue
		}
		rb := b.Points[i].Close/b.Points[i-1].Close - 1
		val *= 1 + (1-w)*ra + w*rb
		out.Points = append(out.Points, marketdata.Point{Date: b.Points[i].Date, Close: val})
	}
	return out
}
