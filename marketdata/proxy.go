package marketdata

import "sort"

// proxyFor maps common ETFs to indices or older mutual funds tracking the
// same market, used to reconstruct ("simulate") history before the ETF's
// inception. Index proxies (^GSPC, ^NDX, ^RUT) are price-only and therefore
// understate total return over the simulated span; mutual fund proxies
// (Vanguard funds) include dividends.
var proxyFor = map[string]string{
	// US large cap / total market → S&P 500 index.
	"SPY": "^GSPC", "VOO": "^GSPC", "IVV": "^GSPC", "SPLG": "^GSPC",
	"VTI": "^GSPC", "ITOT": "^GSPC", "SCHB": "^GSPC",
	"CSPX.L": "^GSPC", "VUSA.L": "^GSPC", "VUAA.L": "^GSPC", "SXR8.DE": "^GSPC",
	// Nasdaq 100.
	"QQQ": "^NDX", "QQQM": "^NDX", "EQQQ.L": "^NDX",
	// US small cap → Russell 2000.
	"IWM": "^RUT", "VB": "^RUT",
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
// was extended.
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
