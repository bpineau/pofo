package simgen

import (
	"fmt"
	"time"

	"portfodor/pkg/marketdata"
)

// ComponentsFrom is how far back component histories are requested; actual
// frames start at the youngest component's first quote.
var ComponentsFrom = time.Date(1962, 1, 1, 0, 0, 0, 0, time.UTC)

// minBackcastR2 is the floor under which a regression-based reconstruction
// is considered too unfaithful to be written at all.
const minBackcastR2 = 0.35

// All returns every bundled reconstruction recipe.
func All() []Recipe {
	return []Recipe{
		ntsxRecipe(),
		iefRecipe(),
		tltRecipe(),
		xauusdRecipe(),
		ntsgRecipe(),
		urthRecipe(),
		iwdaRecipe(),
		wintonRecipe(),
		zrozRecipe(),
		dbmfRecipe(),
		kmlmRecipe(),
		ctaRecipe(),
		amundiVolRecipe(),
		bhmgRecipe(),
	}
}

// iwdaRecipe gives the iShares Core MSCI World (2009) the same 60/40
// US/international reconstruction as URTH, so MSCI-World portfolios reach
// back to 1999.
func iwdaRecipe() Recipe {
	return Recipe{
		ID:     "IE00B4L5Y983",
		Name:   "iShares Core MSCI World — réplication 60/40 US/international",
		Method: "0.60×VFINX + 0.40×VTMGX, frais 0.20 %/an",
		Build: composite("IWDA (réplication MSCI World)", []Leg{
			{ID: "VFINX", Weight: 0.60},
			{ID: "VTMGX", Weight: 0.40},
		}, "", 0.0020),
		ValidateAgainst: "IE00B4L5Y983",
	}
}

// wintonRecipe rebuilds the Winton Trend-Enhanced Global Equity fund as
// global equities plus a half-sized overlay of the actual Winton Trend Fund
// (refdata/WINTON-TREND-REF: réel depuis 2019, simulation Winton avant).
func wintonRecipe() Recipe {
	return Recipe{
		ID:     "IE000O1VI174",
		Name:   "Winton Trend-Enhanced Global Equity — actions + fonds Trend réel",
		Method: "0.60×VFINX + 0.40×VTMGX + 0.50×(WINTON-TREND-REF − cash ^IRX), frais 0.80 %/an",
		Build: composite("Winton Trend-Enhanced Global Equity (réplication)", []Leg{
			{ID: "VFINX", Weight: 0.60},
			{ID: "VTMGX", Weight: 0.40},
			{ID: "WINTON-TREND-REF", Weight: 0.50, Excess: true},
		}, "^IRX", 0.0080),
		ValidateAgainst: "IE000O1VI174",
	}
}

// Find returns the recipe whose ID or validation target matches id.
func Find(id string) (Recipe, bool) {
	canonical := marketdata.CanonicalID(id)
	for _, r := range All() {
		if r.ID == canonical || r.ID == id {
			return r, true
		}
	}
	return Recipe{}, false
}

// refImport is the shared Build for reconstructions that ARE an imported
// reference series (official index, third-party simulation): the series is
// served as-is from refdata/, with an optional annual fee drag to bridge an
// index to its investable wrapper.
func refImport(refID, name string, annualFee float64) func(Fetcher, time.Time) (*marketdata.Series, error) {
	return composite(name, []Leg{{ID: refID, Weight: 1}}, "", annualFee)
}

// composite is the shared Build for constant-weight linear recipes.
func composite(name string, legs []Leg, cashID string, fee float64) func(Fetcher, time.Time) (*marketdata.Series, error) {
	return func(f Fetcher, from time.Time) (*marketdata.Series, error) {
		var ids []string
		if cashID != "" {
			ids = append(ids, cashID)
		}
		for _, l := range legs {
			ids = append(ids, l.ID)
		}
		fr, err := BuildFrame(f, ids, from)
		if err != nil {
			return nil, err
		}
		values, err := Composite(fr, legs, cashID, fee)
		if err != nil {
			return nil, err
		}
		return SeriesFromFrame(name, fr, values), nil
	}
}

// ntsxRecipe rebuilds the WisdomTree US Efficient Core (90 % US equities +
// 60 % treasury futures ladder) from Vanguard index funds and the T-bill
// rate. The simdata file extends the UCITS share class (IE000KF370H3); the
// validation runs against the US-listed twin NTSX, which has traded since
// 2018 with the exact same strategy.
func ntsxRecipe() Recipe {
	return Recipe{
		ID:     "IE000KF370H3",
		Name:   "WisdomTree US Efficient Core — réplication 90/60",
		Method: "0.90×VFINX + 0.60×(VFITX − cash ^IRX) + 0.10×cash, rebalancement quotidien, frais 0.20 %/an",
		Build: composite("NTSX (réplication 90/60)", []Leg{
			{ID: "VFINX", Weight: 0.90},
			{ID: "VFITX", Weight: 0.60, Excess: true},
			{ID: "^IRX", Weight: 0.10},
		}, "^IRX", 0.0020),
		ValidateAgainst: "NTSX-US",
		SpliceReal:      "NTSX-US",
	}
}

// ntsgRecipe is the global variant (NTSG UCITS): 90 % global developed
// equities approximated as 60/40 US/international, plus the same 60 %
// treasury overlay.
func ntsgRecipe() Recipe {
	return Recipe{
		ID:     "IE00077IIPQ8",
		Name:   "WisdomTree Global Efficient Core — réplication 90/60 monde",
		Method: "0.54×VFINX + 0.36×VTMGX + 0.60×(VFITX − cash ^IRX) + 0.10×cash, frais 0.25 %/an",
		Build: composite("NTSG (réplication 90/60 monde)", []Leg{
			{ID: "VFINX", Weight: 0.54},
			{ID: "VTMGX", Weight: 0.36},
			{ID: "VFITX", Weight: 0.60, Excess: true},
			{ID: "^IRX", Weight: 0.10},
		}, "^IRX", 0.0025),
		ValidateAgainst: "IE00077IIPQ8",
	}
}

// urthRecipe approximates the MSCI World as a fixed 60/40 US/developed-
// ex-US blend of Vanguard index funds.
func urthRecipe() Recipe {
	return Recipe{
		ID:     "URTH",
		Name:   "iShares MSCI World — réplication 60/40 US/international",
		Method: "0.60×VFINX + 0.40×VTMGX, frais 0.24 %/an",
		Build: composite("URTH (réplication MSCI World)", []Leg{
			{ID: "VFINX", Weight: 0.60},
			{ID: "VTMGX", Weight: 0.40},
		}, "", 0.0024),
		ValidateAgainst: "URTH",
	}
}

// iefRecipe and tltRecipe extend the treasury ETFs (2002) with third-party
// yield-curve reconstructions going back to 1962.
func iefRecipe() Recipe {
	return Recipe{
		ID:              "IEF",
		Name:            "iShares 7-10Y Treasury — courbes de taux (réf. importée)",
		Method:          "réf. IEF-REF (1962→), frais 0.15 %/an, réel IEF greffé depuis 2002",
		Build:           refImport("IEF-REF", "IEF (réf. courbes de taux)", 0.0015),
		ValidateAgainst: "IEF",
		SpliceReal:      "IEF",
	}
}

func tltRecipe() Recipe {
	return Recipe{
		ID:              "TLT",
		Name:            "iShares 20+Y Treasury — courbes de taux (réf. importée)",
		Method:          "réf. TLT-REF (1962→), frais 0.15 %/an, réel TLT greffé depuis 2002",
		Build:           refImport("TLT-REF", "TLT (réf. courbes de taux)", 0.0015),
		ValidateAgainst: "TLT",
		SpliceReal:      "TLT",
	}
}

// xauusdRecipe extends gold (real series: GC=F futures from 2000) with the
// imported spot fixings back to 1968.
func xauusdRecipe() Recipe {
	return Recipe{
		ID:              "XAUUSD",
		Name:            "Or XAU/USD — spot 1968→",
		Method:          "réf. XAUUSD-SPOT (fixings importés, 1968→), réel GC=F greffé depuis 2000",
		Build:           refImport("XAUUSD-SPOT", "Or (spot importé)", 0),
		ValidateAgainst: "XAUUSD",
		SpliceReal:      "XAUUSD",
	}
}

// zrozRecipe serves the imported yield-curve reconstruction of 25+ year
// STRIPS (refdata/ZROZ-REF, 1962→), real ZROZ quotes grafted on top. It
// replaces the earlier fixed-beta stretch of VUSTX, whose pre-2009 duration
// was too short (la comparaison croisée l'a montré).
func zrozRecipe() Recipe {
	return Recipe{
		ID:              "ZROZ",
		Name:            "PIMCO 25+Y zero-coupon — courbes de taux (réf. importée)",
		Method:          "réf. ZROZ-REF (STRIPS 25+ dérivés des courbes de taux US, 1962→), réel ZROZ greffé depuis 2009",
		Build:           refImport("ZROZ-REF", "ZROZ (réf. courbes de taux)", 0.0015),
		ValidateAgainst: "ZROZ",
		SpliceReal:      "ZROZ",
	}
}

// dbmfRecipe anchors DBMF on the official SG CTA Index (the very index the
// fund replicates), real DBMF quotes grafted on top.
func dbmfRecipe() Recipe {
	return Recipe{
		ID:              "DBMF",
		Name:            "iMGP DBi Managed Futures — SG CTA Index",
		Method:          "réf. SG-CTA (indice officiel, 2000→), réel DBMF greffé depuis 2019",
		Build:           refImport("SG-CTA", "DBMF (SG CTA Index)", 0),
		ValidateAgainst: "DBMF",
		SpliceReal:      "DBMF",
	}
}

// kmlmRecipe anchors KMLM on the official MLM Index history (1987→), real
// KMLM quotes grafted on top.
func kmlmRecipe() Recipe {
	return Recipe{
		ID:              "KMLM",
		Name:            "KraneShares MLM — indice MLM officiel",
		Method:          "réf. MLM-INDEX (historique officiel, 1987→), frais ETF 0.90 %/an, réel KMLM greffé depuis 2020",
		Build:           refImport("MLM-INDEX", "KMLM (indice MLM)", 0.0090),
		ValidateAgainst: "KMLM",
		SpliceReal:      "KMLM",
	}
}

// ctaRecipe anchors Simplify CTA on the official SG Trend Index, real CTA
// quotes grafted on top.
func ctaRecipe() Recipe {
	return Recipe{
		ID:              "CTA",
		Name:            "Simplify CTA — SG Trend Index",
		Method:          "réf. SG-TREND (indice officiel, 2000→), réel CTA greffé depuis 2022",
		Build:           refImport("SG-TREND", "CTA (SG Trend Index)", 0),
		ValidateAgainst: "CTA",
		SpliceReal:      "CTA",
	}
}

// backcastBuild wraps FitBackcast: the model is fitted on the asset's real
// history, then projected over the whole frame. Honest but limited — only
// the systematic exposures survive.
func backcastBuild(name, realID string, ids []string) func(Fetcher, time.Time) (*marketdata.Series, error) {
	return func(f Fetcher, from time.Time) (*marketdata.Series, error) {
		fr, err := BuildFrame(f, ids, from)
		if err != nil {
			return nil, err
		}
		real, err := f.Fetch(realID, from)
		if err != nil {
			return nil, err
		}
		values, r2, _, err := FitBackcast(fr, real, ids)
		if err != nil {
			return nil, err
		}
		if r2 < minBackcastR2 {
			return nil, fmt.Errorf("%w: R² in-sample %.2f < %.2f", ErrUnfaithful, r2, minBackcastR2)
		}
		return SeriesFromFrame(fmt.Sprintf("%s (backcast R²=%.2f)", name, r2), fr, values), nil
	}
}

// amundiVolRecipe attempts a regression backcast of the Amundi Volatility
// World fund on VIX variations; volatility-trading funds are idiosyncratic,
// so this only ships when the in-sample fit clears minBackcastR2.
func amundiVolRecipe() Recipe {
	return Recipe{
		ID:              "LU0319687124",
		Name:            "Amundi Volatility World — backcast sur ^VIX",
		Method:          "régression des rendements quotidiens du fonds sur Δ^VIX et VFISX (2007→), rejouée avant 2007; résidus ignorés",
		Build:           backcastBuild("Amundi Volatility World", "LU0319687124", []string{"^VIX", "VFISX"}),
		ValidateAgainst: "LU0319687124",
	}
}

// bhmgRecipe attempts the same exercise for BH Macro; discretionary macro
// rarely regresses well on asset-class factors, in which case nothing is
// written.
func bhmgRecipe() Recipe {
	return Recipe{
		ID:              "GG00BQBFY362",
		Name:            "BH Macro — backcast factoriel",
		Method:          "régression des rendements quotidiens sur VUSTX, VFINX, GC=F (2007→), rejouée avant; résidus ignorés",
		Build:           backcastBuild("BH Macro", "GG00BQBFY362", []string{"VUSTX", "VFINX", "GC=F"}),
		ValidateAgainst: "GG00BQBFY362",
	}
}
