package simgen

import (
	"fmt"
	"os"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// The global government bond futures basket of the WisdomTree Global Efficient
// Core fund (NTSG), as a set of LOCAL-CURRENCY excess-return sleeves.
//
// # Why four currencies and not one
//
// The fund's bond sleeve is not a US Treasury ladder. Its audited holdings put
// the notional at roughly 80 % US, 11 % German, 6 % Japanese and 3 % British
// paper (the statement of investments at 31/12/2025 shows 79.7/11.2/6.1/3.0,
// and the period averages of the SFDR annex corroborate it at 77.9/11.0/5.9/3.0).
// Rebuilding all of it from US Treasuries alone would hand the reconstruction a
// single yield curve, and the four have spent whole decades disagreeing: the
// Bund and the JGB rallied through the years the Treasury sold off, and the
// 1970s gilt market has no American equivalent at all.
//
// # Why the sleeves are LOCAL and carry no currency
//
// A bond FUTURE is entered at no cost, so what its holder earns is the bond's
// return less the financing of the notional, in the bond's OWN currency; and
// the fund rolls the currency exposure of the foreign notionals away with
// spot/next forwards, whose points are exactly the interest differential. What
// survives is therefore the local excess return, and nothing else: no spot
// conversion, and no carry on top. Adding either would count the same interest
// differential twice, once inside the forward points and once outside them,
// which is the standard way to make a foreign bond sleeve look richer than it
// was.
//
// # What each sleeve is made of
//
// The US sleeve rides the two real Vanguard Treasury funds the whole efficient
// core family is rebuilt from, in the duration blend measured below; the other
// three ride the bundled reference series cmd/gen-gbond-refdata writes, each
// financed at its own country's money-market rate:
//
//	USD  0.80  VFITX / VUSTX blend (CMT reconstructions behind them, ~1953)
//	           financed at usdOvernight
//	DEU  0.11  BUND-EUR (~1956, daily BUND-DAILY shape from 1997-08)
//	           financed at the euro overnight rate, DECASH-EUR before the euro
//	JPY  0.06  JGB-JPY (daily, ~1986-07) financed at JPCASH-JPY (~1985-07)
//	GBP  0.03  GILT-GBP (monthly, ~1960) financed at GBCASH-GBP (~1978-01)
//
// # What happens before a sleeve exists
//
// The weights are RENORMALIZED over the sleeves that quote, day by day, so the
// overlay always carries a full notional (see blendExcess). The sterling sleeve
// therefore starts in 1978 and the yen one in 1986-07, and before those dates
// their share is carried by the sleeves that do exist rather than by an invented
// proxy: a Japanese bond series that does not exist cannot be modelled, and
// substituting a Treasury for it would say the opposite of what the basket is
// for. The composite's own floor (~1969, set by the equity leg) sits inside the
// era where the US and German sleeves both quote, so the deepest years of the
// backcast run on 88 % of the basket, then 91 %, then all of it.
const (
	usdBondShare = 0.80
	deuBondShare = 0.11
	jpnBondShare = 0.06
	gbrBondShare = 0.03
)

// usdShortShare is the share of the US sleeve carried by the INTERMEDIATE
// Treasury donor (VFITX), the rest going to the long one (VUSTX). The fund
// holds a laddered basket of futures rather than a fund, so the pair has to be
// mixed to the ladder's duration, and the number is MEASURED rather than
// assumed, three ways that agree:
//
//   - Regressing the US sibling NTSX's own returns, less its 0.90 equity leg
//     and its 0.10 collateral leg, on the two donors' excess returns over its
//     whole record (2018-08 to 2026-07) gives 0.772 of the pair on VFITX at
//     both weekly (399 points) and monthly (95 points) horizons. The same
//     regression on the international sibling NTSI (2021-05 on, which finances
//     the same US ladder) gives 0.812 weekly and 0.833 monthly.
//   - The two donors' own effective durations, regressed on the daily
//     constant-maturity yields they track over 2015-2026, are 4.90 years for
//     VFITX (against the 5-year point) and 16.16 for VUSTX (against the
//     30-year). A 0.78/0.22 blend of them has a duration of 7.4 years.
//   - 7.4 years is what an equally weighted 2/10/30-year Treasury futures
//     ladder carries, and it is the effective duration WisdomTree publishes for
//     the sleeve (~7 years).
//
// So 0.78. The previous reconstruction put the whole sleeve on VFITX, i.e. on a
// 4.9-year duration where the fund runs 7.4: it under-owned the bond leg's
// interest-rate risk by a third, which is most of why the engine ran hot in
// equity-led years and cold in bond-led ones.
//
// One honest caveat: the donors' durations are not constant back to 1953. Their
// deep tails are constant-maturity reconstructions of a 5-year and a 20-year par
// bond (TREASURY-INT-USD, TREASURY-LONG-USD), whose durations move with the
// yield level, so the blend runs a little shorter than 7.4 years in the
// high-yield 1970s and a little longer in the 2010s. A single number is the
// right resolution for a sleeve rebuilt from two funds.
const usdShortShare = 0.78

// bondSleeve is one currency's contribution to the overlay: a weight in the
// basket and a builder returning that country's LOCAL excess-return index (a
// level series whose returns are already net of local financing).
type bondSleeve struct {
	name   string
	weight float64
	build  func(Fetcher, time.Time) (*marketdata.Series, error)
}

// globalBondOverlay builds the fund's whole bond sleeve as ONE excess-return
// index: the weighted, renormalized blend of the four local sleeves, expressed
// as a level series so the ordinary Composite machinery can carry it as a plain
// leg at the fund's 0.60 notional.
//
// The US sleeve sets the calendar: it is the deepest (1953) and the only one
// that quotes on every date the composite ever needs, so the foreign sleeves are
// read onto it by forward fill. That is exact for the daily ones and turns the
// monthly gilt series into a step at each month boundary, which is the accepted
// texture for 1.8 % of net assets (see gbrBondShare).
func globalBondOverlay(f Fetcher, from time.Time) (*marketdata.Series, error) {
	sleeves := []bondSleeve{
		{"USD Treasuries", usdBondShare, usdBondSleeve},
		{"German govt", deuBondShare, deuBondSleeve},
		{"Japanese govt", jpnBondShare, jpnBondSleeve},
		{"British govt", gbrBondShare, gbrBondSleeve},
	}
	built := make([]bondSleeve, 0, len(sleeves))
	series := make([]*marketdata.Series, 0, len(sleeves))
	for _, sl := range sleeves {
		s, err := sl.build(f, from)
		if err != nil {
			// A missing foreign reference must not sink a reconstruction whose
			// US sleeve is four fifths of the basket: the blend renormalizes,
			// and the build says which sleeve it lost.
			if sl.weight >= usdBondShare {
				return nil, fmt.Errorf("bond overlay: %s: %w", sl.name, err)
			}
			fmt.Fprintf(os.Stderr, "bond overlay: %s unavailable (%v), its weight is carried by the others\n", sl.name, err)
			continue
		}
		fmt.Fprintf(os.Stderr, "bond overlay: %s (%.0f%% of the basket) from %s\n",
			sl.name, sl.weight*100, s.First().Date.Format("2006-01-02"))
		built = append(built, sl)
		series = append(series, s)
	}
	if len(built) == 0 {
		return nil, fmt.Errorf("bond overlay: no sleeve available")
	}
	return blendExcess("NTSG global government bond overlay (local excess returns)", series[0].Points, built, series), nil
}

// blendExcess compounds the weighted average of the sleeves' returns on the
// calendar of the first one, renormalizing the weights over the sleeves that
// cover each step. A sleeve joins on the first date its own history reaches and
// never before: until then its weight is spread across the others, so the index
// always represents a fully invested basket rather than a partially empty one.
func blendExcess(name string, calendar []marketdata.Point, sleeves []bondSleeve, series []*marketdata.Series) *marketdata.Series {
	out := &marketdata.Series{Name: name, Source: "simdata"}
	level := 100.0
	out.Points = append(out.Points, marketdata.Point{Date: calendar[0].Date, Close: level})
	for k := 1; k < len(calendar); k++ {
		var sum, weight float64
		for i, sl := range sleeves {
			v0, _, ok0 := series[i].At(calendar[k-1].Date)
			v1, _, ok1 := series[i].At(calendar[k].Date)
			if !ok0 || !ok1 || v0 <= 0 {
				continue
			}
			sum += sl.weight * (v1/v0 - 1)
			weight += sl.weight
		}
		if weight > 0 {
			level *= 1 + sum/weight
		}
		out.Points = append(out.Points, marketdata.Point{Date: calendar[k].Date, Close: level})
	}
	return out
}

// usdBondSleeve is the US leg: the duration-matched blend of the two Vanguard
// Treasury funds (usdShortShare), each extended by its constant-maturity
// reconstruction to 1953, financed at the USD overnight rate a future's implied
// repo costs rather than at the bill rate its collateral earns (see
// usdOvernight).
func usdBondSleeve(f Fetcher, from time.Time) (*marketdata.Series, error) {
	return excessIndex("US Treasury futures sleeve", financed(extend(f)), usdOvernight, from, []Leg{
		{ID: "VFITX", Weight: usdShortShare, Excess: true},
		{ID: "VUSTX", Weight: 1 - usdShortShare, Excess: true},
	})
}

// deuBondSleeve is the German leg: the bundled 10-year Bund reconstruction
// (BUND-EUR from 1956, carried at daily granularity by the Bundesbank curve
// BUND-DAILY from 1997-08), financed at the euro overnight rate, itself carried
// before the euro by the German money-market accrual.
func deuBondSleeve(f Fetcher, from time.Time) (*marketdata.Series, error) {
	bund, err := shapedRefdata(f, "BUND-EUR", "BUND-DAILY", from)
	if err != nil {
		return nil, err
	}
	cash, err := eurOvernightDeep(f, from)
	if err != nil {
		return nil, err
	}
	return localExcess("German government bond futures sleeve", bund, cash, from)
}

// jpnBondSleeve is the Japanese leg: the bundled 10-year JGB reconstruction
// (daily from 1986-07) financed at the yen call-money accrual.
func jpnBondSleeve(f Fetcher, from time.Time) (*marketdata.Series, error) {
	jgb, err := refdata(f, "JGB-JPY", from)
	if err != nil {
		return nil, err
	}
	cash, err := cashRefdata(f, "JPCASH-JPY", "JPY call money (daily)", from)
	if err != nil {
		return nil, err
	}
	return localExcess("Japanese government bond futures sleeve", jgb, cash, from)
}

// gbrBondSleeve is the British leg: the bundled 10-year gilt reconstruction
// (monthly from 1960) financed at the sterling interbank accrual (from 1978).
// Both are monthly, so this sleeve steps once a month; it is 1.8 % of the
// fund's net assets and no daily gilt curve is worth a workbook parser for it.
func gbrBondSleeve(f Fetcher, from time.Time) (*marketdata.Series, error) {
	gilt, err := refdata(f, "GILT-GBP", from)
	if err != nil {
		return nil, err
	}
	cash, err := refdata(f, "GBCASH-GBP", from)
	if err != nil {
		return nil, err
	}
	return localExcess("British government bond futures sleeve", gilt, cash, from)
}

// localExcess turns a bond total-return index and its local money-market index
// into the excess-return index of a futures position in that market: the frame
// aligns the pair, and the composite earns the bond less the financing.
func localExcess(name string, bond, cash *marketdata.Series, from time.Time) (*marketdata.Series, error) {
	const bondID, cashID = "BOND", "CASH"
	inj := injected{have: map[string]*marketdata.Series{bondID: bond, cashID: cash}}
	return excessIndex(name, inj, cashID, from, []Leg{{ID: bondID, Weight: 1, Excess: true}})
}

// excessIndex builds a level index (base 100) whose returns are the legs'
// financed returns, so an overlay assembled elsewhere can carry it as an
// ordinary leg.
func excessIndex(name string, f Fetcher, cashID string, from time.Time, legs []Leg) (*marketdata.Series, error) {
	ids := make([]string, 0, len(legs)+1)
	ids = append(ids, cashID)
	for _, l := range legs {
		ids = append(ids, l.ID)
	}
	fr, err := BuildFrame(f, ids, from)
	if err != nil {
		return nil, err
	}
	values, err := Composite(fr, legs, cashID, 0)
	if err != nil {
		return nil, err
	}
	return SeriesFromFrame(name, fr, values), nil
}

// refdata fetches a bundled reference series and refuses an empty one, since a
// sleeve built on nothing would silently become a weight given to its
// neighbours.
func refdata(f Fetcher, id string, from time.Time) (*marketdata.Series, error) {
	s, err := f.Fetch(id, from)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", id, err)
	}
	if s == nil || len(s.Points) < 2 {
		return nil, fmt.Errorf("%s: empty history", id)
	}
	return s, nil
}

// shapedRefdata fetches a monthly reference series and blends in the daily
// shape of the same market where it reaches, exactly as extend() does for the
// proxies that stand behind a fetchable fund. It is spelled out here because
// these series stand behind no fund: they ARE the sleeve.
func shapedRefdata(f Fetcher, anchorID, shapeID string, from time.Time) (*marketdata.Series, error) {
	anchor, err := refdata(f, anchorID, from)
	if err != nil {
		return nil, err
	}
	if shape, serr := f.Fetch(shapeID, from); serr == nil && shape != nil && len(shape.Points) > 300 {
		return shapedSeries(anchor, shape), nil
	}
	fmt.Fprintf(os.Stderr, "bond overlay: %s: no daily shape %s, monthly cadence kept\n", anchorID, shapeID)
	return anchor, nil
}

// cashRefdata fetches a monthly money-market accrual and expands it to business
// days (cashDaily), so the excess return it finances is a smooth daily accrual
// rather than a jump at each month boundary.
func cashRefdata(f Fetcher, id, name string, from time.Time) (*marketdata.Series, error) {
	m, err := refdata(f, id, from)
	if err != nil {
		return nil, err
	}
	return cashDaily(name, m), nil
}

// eurOvernightDeep is the euro financing rate of the German sleeve, carried as
// far back as the sleeve reaches: the euro overnight accrual (ESTR from 2019-10,
// EONIA from 1999-01, the 3-month money-market index over 1994-1999) extended
// before the euro by the German money-market accrual DECASH-EUR (~1960).
// Germany was the anchor economy and the mark the reference currency, so its own
// short rate is what a Bund future would have financed at.
func eurOvernightDeep(f Fetcher, from time.Time) (*marketdata.Series, error) {
	idx, err := eurOvernightDaily(f, from)
	if err != nil {
		return nil, err
	}
	deep, derr := f.Fetch("DECASH-EUR", from)
	if derr != nil || deep == nil {
		return idx, nil
	}
	// eurOvernightDaily has already spliced the 3-month index under the
	// overnight one and stamped SimulatedBefore, which ExtendBack reads as "this
	// series has been extended once, do not do it again". Clearing it is what
	// lets a second, deeper era go in front, exactly as shapedSeries does.
	idx.SimulatedBefore = time.Time{}
	marketdata.ExtendBack(idx, deep)
	return idx, nil
}
