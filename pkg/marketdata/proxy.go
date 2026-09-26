package marketdata

import "sort"

// proxyFor maps common ETFs to a longer-lived series tracking the same
// market, used to reconstruct ("simulate") history before the ETF's
// inception. Every proxy is a TOTAL-RETURN series: a bundled reconstruction
// (SP500, VTI), a total-return index (^RUTTR), an older fund's adjusted NAV
// (the Vanguard mutual funds, QQQ) or a gold futures price, gold paying no
// income. A price index would drop the dividends over the simulated span, and
// the gap is not small: measured on the ETFs' own lives to 2026-09, ^RUT runs
// 1.37 points a year under IWM and ^NDX 0.57 under QQQ.
//
// The single exception is QQQ itself: no public total-return Nasdaq-100
// reaches before QQQ's own 1999-03 launch (^XNDX starts there), so its
// 1985-1999 span stays the price index and understates by roughly the
// index's dividend yield of those years.
//
// A proxy that has a longer history of its own (another entry, a bundled
// simdata id) is fetched extended, so that history comes along (see
// proxyHistory); TestProxyChainsEnd keeps the chains acyclic.
var proxyFor = map[string]string{
	// S&P 500 → the bundled S&P 500 total return (monthly 1871→, daily 1927→).
	"SPY": "SP500", "VOO": "SP500", "IVV": "SP500", "SPLG": "SP500",
	"CSPX.L": "SP500", "VUSA.L": "SP500", "VUAA.L": "SP500", "SXR8.DE": "SP500",
	// US total market → the bundled VTI reconstruction (VTSMX, CRSP 1926→).
	"ITOT": "VTI", "SCHB": "VTI",
	// Nasdaq 100 → QQQ (1999), itself on the ^NDX price index (see above).
	"QQQ": "^NDX", "QQQM": "QQQ", "EQQQ.L": "QQQ",
	// US small cap: Russell 2000 total return (1995) for the Russell fund,
	// VB's own mutual share class (1960; tracked the Russell 2000 to 2003,
	// then MSCI and CRSP small cap like VB) for the Vanguard one.
	"IWM": "^RUTTR", "VB": "NAESX",
	// Developed ex-US → Vanguard Developed Markets fund (1999).
	"EFA": "VTMGX", "VEA": "VTMGX", "IEFA": "VTMGX",
	// Emerging markets → Vanguard Emerging Markets fund (1994).
	"EEM": "VEIEX", "VWO": "VEIEX", "IEMG": "VEIEX",
	// US aggregate bonds → Vanguard Total Bond fund (1986).
	"AGG": "VBMFX", "BND": "VBMFX", "SCHZ": "VBMFX",
	// US treasuries by maturity → Vanguard treasury funds (1991).
	"TLT": "VUSTX", "VGLT": "VUSTX", "EDV": "VUSTX",
	"IEF": "VFITX", "VGIT": "VFITX",
	"SHY": "VFISX", "VGSH": "VFISX", "BIL": "VFISX",
	// TIPS → Vanguard Inflation-Protected fund (2000).
	"TIP": "VIPSX", "SCHP": "VIPSX",
	// Gold → COMEX gold futures (2000).
	"GLD": "GC=F", "IAU": "GC=F", "SGOL": "GC=F", "GLDM": "GC=F",
	// US REITs → Vanguard REIT fund (1996).
	"VNQ": "VGSIX", "IYR": "VGSIX", "SCHH": "VGSIX",
	// Gold-miner equity funds → VanEck Gold Miners ETF (2006).
	"0P000163EJ.F": "GDX",
	// WisdomTree Efficient Core UCITS (2023/2024) → the original US-listed
	// NTSX (2018). Exact strategy for the US fund; approximate (US-only
	// instead of global 90/60) for the global one.
	"IE000KF370H3": "NTSX", // NTSX UCITS
	"IE00077IIPQ8": "NTSX", // NTSG UCITS, approximation
}

// ProxySymbol returns the symbol used to reconstruct early history for
// symbol, if one is known.
func ProxySymbol(symbol string) (string, bool) {
	p, ok := proxyFor[symbol]
	return p, ok
}

// ExtendBack prepends proxy history, rescaled to the asset's first quote, for
// dates before the asset's own history starts. It reports whether the series
// was extended. Only the proxy's returns are kept, so it must be quoted in
// the asset's currency (a sub-unit such as GBp and its currency agree):
// FetchExtended converts it first.
func ExtendBack(s, proxy *Series) bool {
	if len(s.Points) == 0 || len(proxy.Points) == 0 || !s.SimulatedBefore.IsZero() {
		return false
	}
	anchor := s.Points[0]
	// Proxy value at the last date at or before the anchor date.
	i := sort.Search(len(proxy.Points), func(i int) bool {
		return proxy.Points[i].Date.After(anchor.Date)
	}) - 1
	if i < 0 || proxy.Points[i].Close <= 0 {
		return false
	}
	scale := anchor.Close / proxy.Points[i].Close
	var pre []Point
	for _, p := range proxy.Points {
		if !p.Date.Before(anchor.Date) {
			break
		}
		pre = append(pre, Point{Date: p.Date, Close: p.Close * scale})
	}
	if len(pre) == 0 {
		return false
	}
	s.Points = append(pre, s.Points...)
	s.SimulatedBefore = anchor.Date
	s.ProxySymbol = proxy.Symbol
	return true
}
