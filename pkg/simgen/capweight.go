package simgen

import (
	"fmt"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// CapWeighted builds an index (base 100) from legs held the way a
// CAP-WEIGHTED index holds them: the published split at anchor, and nothing
// else. No rebalancing happens, because a cap-weighted index does not
// rebalance: between reconstitutions it is a buy-and-hold portfolio, and its
// country weights drift with the countries' own returns. So the weights of
// period t-1 are w_i,t / (1 + r_i,t) renormalized, going back, and w_i,t
// (1 + r_i,t+1) renormalized, going forward, which is exactly what holding the
// anchor's shares produces in both directions.
//
// That matters because the alternative, carrying TODAY'S split back through
// history at constant weights, is a look-ahead bias, and a large one: the US
// share of the world market was near 30 % at the 1989 Japan peak and above
// 50 % in 2000-2002. A world-equity reconstruction pinned at today's ~60 % US
// earns the US market's whole outperformance over decades when it held nothing
// like that weight, worth about +1.8 pts/yr over 1988-2008 on this blend.
//
// What the drift IGNORES is net issuance: an index's weights also move when
// companies issue or buy back shares, when a market's free float opens (China A
// shares entering the global indices, for one) and when the index's own
// coverage changes, none of which is a return. That is a real limit, measured
// below rather than waved away, and it is the reason the anchor is taken as
// close as possible to the era the reconstruction serves.
//
// annualFee is deducted pro rata temporis, exactly as in Composite. A leg
// missing from the frame is an error, and the weights are renormalized (the
// published splits are rounded percentages).
func CapWeighted(fr *Frame, legs []Leg, anchor time.Time, annualFee float64) ([]float64, error) {
	if len(legs) == 0 {
		return nil, fmt.Errorf("no leg")
	}
	var total float64
	for _, l := range legs {
		if _, ok := fr.Returns[l.ID]; !ok {
			return nil, fmt.Errorf("component %s missing from frame", l.ID)
		}
		if l.Excess {
			return nil, fmt.Errorf("leg %s: an excess leg has no market weight to drift", l.ID)
		}
		if l.Weight <= 0 {
			return nil, fmt.Errorf("leg %s: weight %v", l.ID, l.Weight)
		}
		total += l.Weight
	}
	// Each leg's own level path, base 1 at the anchor: holding the anchor's
	// shares is worth l.Weight × that path, before and after the anchor alike.
	a := anchorIndex(fr.Dates, anchor)
	basket := make([]float64, len(fr.Dates))
	for _, l := range legs {
		lv := levelsFrom(fr.Returns[l.ID], a)
		for k := range basket {
			basket[k] += l.Weight / total * lv[k]
		}
	}
	values := make([]float64, len(fr.Dates))
	values[0] = 100
	feeDaily := annualFee / 252
	for k := 1; k < len(fr.Dates); k++ {
		if basket[k-1] <= 0 {
			return nil, fmt.Errorf("%s: non-positive basket", fr.Dates[k-1].Format("2006-01-02"))
		}
		values[k] = values[k-1] * (1 + basket[k]/basket[k-1] - 1 - feeDaily)
	}
	return values, nil
}

// CapWeights reports the weights the drift implies on the frame's date k, the
// series a reader checks against a published split. It repeats CapWeighted's
// arithmetic rather than sharing a buffer with it, because a diagnostic that
// can fall out of step with what it diagnoses is worse than none.
func CapWeights(fr *Frame, legs []Leg, anchor time.Time, k int) map[string]float64 {
	a := anchorIndex(fr.Dates, anchor)
	out := make(map[string]float64, len(legs))
	var sum float64
	for _, l := range legs {
		ret, ok := fr.Returns[l.ID]
		if !ok {
			continue
		}
		v := l.Weight * levelsFrom(ret, a)[k]
		out[l.ID] = v
		sum += v
	}
	for id := range out {
		out[id] /= sum
	}
	return out
}

// levelsFrom turns a frame's daily returns into a level path worth 1 on step a,
// compounding forward from there and dividing backward.
func levelsFrom(ret []float64, a int) []float64 {
	out := make([]float64, len(ret))
	out[a] = 1
	for k := a + 1; k < len(ret); k++ {
		out[k] = out[k-1] * (1 + ret[k])
	}
	for k := a - 1; k >= 0; k-- {
		out[k] = out[k+1] / (1 + ret[k+1])
	}
	return out
}

// anchorIndex is the frame step the anchor date names: the last date on or
// before it, or the first step when the anchor predates the frame entirely (a
// frame that stops short of the anchor, as an offline test's synthetic one
// does, still gets a well-defined blend, pinned at the end it does reach).
func anchorIndex(dates []time.Time, anchor time.Time) int {
	a := 0
	for k, d := range dates {
		if d.After(anchor) {
			break
		}
		a = k
	}
	return a
}

// capWeighted is the Build wrapper of CapWeighted, the cap-weighted sibling of
// composite: same frame, same extend() splices, drifting weights instead of
// constant ones.
func capWeighted(name string, legs []Leg, anchor time.Time, fee float64) func(Fetcher, time.Time) (*marketdata.Series, error) {
	return func(f Fetcher, from time.Time) (*marketdata.Series, error) {
		ids := make([]string, 0, len(legs))
		for _, l := range legs {
			ids = append(ids, l.ID)
		}
		fr, err := BuildFrame(extend(f), ids, from)
		if err != nil {
			return nil, err
		}
		values, err := CapWeighted(fr, legs, anchor, fee)
		if err != nil {
			return nil, err
		}
		return SeriesFromFrame(name, fr, values), nil
	}
}

// The world equity market, as the three legs every world reconstruction here is
// built from, pinned to a PUBLISHED split on a dated observation.
//
// The anchor is the FTSE All-World index column of the market-diversification
// table in Vanguard Total World Stock Index Fund's annual report for the fiscal
// year ended 2009-10-31 (SEC Form N-CSR, accession 0000932471-09-002118, filed
// 2009-12-29): United States 40.4 %, Europe 28.5 %, Pacific 14.3 %, Canada
// 3.3 %, emerging markets 13.5 %. Canada joins the developed-ex-US leg, the
// only leg that can hold it, for 46.1 % against the US 40.4 % and emerging
// 13.5 %. That date is the earliest published split of the world fund's own
// index, which is what makes it the right anchor: every reconstruction here
// serves the years BEFORE its fund existed, so the drift runs backward from the
// anchor and the error grows with distance INTO the past, never into the window
// a consumer's real quotes cover.
//
// Two independent checks of the drift, neither fitted:
//
//   - the same report series, fiscal year ended 2023-10-31 (fund allocation as
//     of that date): United States 60.9 %. Drifting the anchor forward 14 years
//     gives 66.1 %, 5.2 points too much, which is the size of the net-issuance
//     effect the mechanism ignores over fourteen years (China A shares entering
//     the global indices in 2018-2019 alone moved emerging weights).
//   - the 1989 Japan peak: the drift puts the US at 28.5 % of the world at
//     1989-12 and the developed-ex-US leg at 66.2 %. Japan alone was 40 % of
//     the global market at its 1989 zenith (Dimson, Marsh and Staunton data, as
//     reported by CFA Institute's 2025 capital markets report on the 60/40
//     portfolio), which the 66.2 % leg can hold and a constant 30 % leg cannot.
//     The often-quoted "about 30 % US in 1989" is consistent with the 28.5 %
//     here, but is not verified against a primary source in this repository.
//
// The path over the dot-com peak is the third reading, and the one that shows
// the size of what a constant blend would have assumed: 52.1 % US at 2000-12,
// 54.5 % at 2001-12, 52.2 % at 2002-12, then down to 40.1 % by 2007-12.
var (
	worldAnchor = time.Date(2009, 10, 30, 0, 0, 0, 0, time.UTC) // the report's month end, a Friday
	worldLegs   = []Leg{
		{ID: "VFINX", Weight: 0.404}, // US large cap
		{ID: "VTMGX", Weight: 0.461}, // developed ex-US, Canada included
		{ID: "VEIEX", Weight: 0.135}, // emerging markets
	}
)

// worldMethod is the one line every world-equity recipe's Method starts with,
// so the three files cannot describe the same blend three ways.
const worldMethod = "cap-weighted VFINX + VTMGX + VEIEX (US / developed-ex-US / emerging), " +
	"weights pinned to the FTSE All-World split published for 2009-10-31 (40.4/46.1/13.5) " +
	"and drifted by the legs' own returns, net issuance ignored"

// worldEquityFloat is the shared Build of the world-equity reconstructions: the
// three legs at the published anchor split, drifting as a cap-weighted index
// does. fee is the target's own load where its donors' NAVs do not already
// carry it (see the fee constants next to composite).
func worldEquityFloat(name string, fee float64) func(Fetcher, time.Time) (*marketdata.Series, error) {
	return capWeighted(name, worldLegs, worldAnchor, fee)
}

// worldEquityID is the synthetic frame component the stacked recipes hold their
// world-equity sleeve as (see injected).
const worldEquityID = "WORLD-EQ"

// rssbBuild is the Return Stacked Global Stocks & Bonds replication: the whole
// world equity market funded, plus an intermediate-Treasury excess leg stacked
// on top of it. The equity sleeve is built first, as its own cap-weighted
// index, because its weights drift and a constant-weight composite cannot
// express that; the stack itself is then an ordinary two-leg composite.
func rssbBuild(f Fetcher, from time.Time) (*marketdata.Series, error) {
	eq, err := worldEquityFloat("RSSB (world equity sleeve)", 0)(f, from)
	if err != nil {
		return nil, err
	}
	inj := injected{inner: extend(f), have: map[string]*marketdata.Series{worldEquityID: eq}}
	return composite("RSSB (100/100 stocks+bonds replication)", []Leg{
		{ID: worldEquityID, Weight: 1.00},
		{ID: "VFITX", Weight: 1.00, Excess: true},
	}, usdOvernight, 0)(inj, from)
}
