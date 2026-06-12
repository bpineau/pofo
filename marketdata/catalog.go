package marketdata

import (
	"bufio"
	_ "embed"
	"sort"
	"strings"
	"sync"
)

// CatalogEntry pins the resolution of a well-known asset: how to fetch its
// quotes without any network search, plus display metadata.
type CatalogEntry struct {
	ID       string   // canonical identifier (ticker or ISIN)
	ISIN     string   // informational; may be empty for indices/commodities
	Aliases  []string // identifiants alternatifs acceptés dans les fichiers portefeuille
	UCITS    bool     // fonds/ETF au format UCITS (les ETC, fonds US, indices… ne le sont pas)
	Name     string
	Source   string // "yahoo", "ft", "morningstar" or "stooq"
	Symbol   string // yahoo/stooq symbol or Morningstar id; unused for ft
	Xid      string // FT internal id; unused otherwise
	Currency string
	Fees     float64 // published TER, percent per year; 0 = unknown
}

// catalog lists the assets bundled with portfodor: everything here resolves
// deterministically (no search APIs), has its TER pinned when published, and
// is fully cached by a single --warmup. Entries are addressable by ID, ISIN,
// Aliases and the tickers of the embedded fund list — but never by quote
// Symbol, which may collide (e.g. US-listed NTSX vs the NTSX UCITS).
var catalog = []CatalogEntry{
	// US-listed ETFs and mutual funds (Yahoo, adjusted closes).
	{ID: "VOO", ISIN: "US9229083632", Name: "Vanguard S&P 500 ETF", Source: "yahoo", Symbol: "VOO", Currency: "USD", Fees: 0.03},
	{ID: "VT", ISIN: "US9220427424", Name: "Vanguard Total World Stock ETF", Source: "yahoo", Symbol: "VT", Currency: "USD", Fees: 0.06},
	{ID: "IVV", ISIN: "US4642872000", Name: "iShares Core S&P 500 ETF", Source: "yahoo", Symbol: "IVV", Currency: "USD", Fees: 0.03},
	{ID: "BND", ISIN: "US9219378356", Name: "Vanguard Total Bond Market ETF", Source: "yahoo", Symbol: "BND", Currency: "USD", Fees: 0.03},
	{ID: "URTH", Name: "iShares MSCI World ETF", Source: "yahoo", Symbol: "URTH", Currency: "USD", Fees: 0.24},
	{ID: "EEMA", Name: "iShares MSCI Emerging Markets Asia ETF", Source: "yahoo", Symbol: "EEMA", Currency: "USD", Fees: 0.49},
	{ID: "DBMF", Name: "iMGP DBi Managed Futures Strategy ETF", Source: "yahoo", Symbol: "DBMF", Currency: "USD", Fees: 0.85},
	{ID: "KMLM", Name: "KraneShares Mount Lucas Managed Futures Index Strategy ETF", Source: "yahoo", Symbol: "KMLM", Currency: "USD", Fees: 0.90},
	{ID: "CTA", Name: "Simplify Managed Futures Strategy ETF", Source: "yahoo", Symbol: "CTA", Currency: "USD", Fees: 0.75},
	{ID: "ZROZ", Name: "PIMCO 25+ Year Zero Coupon U.S. Treasury ETF", Source: "yahoo", Symbol: "ZROZ", Currency: "USD", Fees: 0.15},
	{ID: "PFOCX", Name: "PIMCO Preferred and Capital Securities Fund", Source: "yahoo", Symbol: "PFOCX", Currency: "USD", Fees: 1.72},
	{ID: "VBMFX", Name: "Vanguard Total Bond Market Index Fund", Source: "yahoo", Symbol: "VBMFX", Currency: "USD", Fees: 0.15},

	// Commodities. Yahoo carries both since ~2000; the Stooq spot series go
	// further back but sit behind an anti-bot wall as of 2026.
	{ID: "XAUUSD", Aliases: []string{"GOLD"}, Name: "Or XAU/USD (via futures GC=F)", Source: "yahoo", Symbol: "GC=F", Currency: "USD"},
	{ID: "CL=F", Aliases: []string{"WTI"}, Name: "Pétrole brut WTI (futures continus)", Source: "yahoo", Symbol: "CL=F", Currency: "USD"},

	// European funds and ETFs, resolutions established by the multi-source
	// pipeline then pinned here (see --warmup).
	{ID: "IE00B4L5Y983", ISIN: "IE00B4L5Y983", Name: "iShares Core MSCI World UCITS ETF USD (Acc)", Source: "yahoo", Symbol: "IWDA.L", Currency: "USD", Fees: 0.20, UCITS: true},
	{ID: "IE000KF370H3", Aliases: []string{"NTSX"}, ISIN: "IE000KF370H3", Name: "WisdomTree US Efficient Core UCITS ETF USD Acc", Source: "ft", Xid: "839245042", Currency: "USD", Fees: 0.20, UCITS: true},     // cotation LSE en USD, cohérente avec la simdata
	{ID: "IE00077IIPQ8", Aliases: []string{"NTSG"}, ISIN: "IE00077IIPQ8", Name: "WisdomTree Global Efficient Core UCITS ETF USD Acc", Source: "ft", Xid: "944239356", Currency: "USD", Fees: 0.25, UCITS: true}, // cotation LSE en USD, cohérente avec la simdata
	{ID: "IE000O1VI174", Aliases: []string{"WINTON-TREND-EQUITY"}, ISIN: "IE000O1VI174", Name: "Winton Trend Enhanced Global Equity Fund (UCITS) I USD Acc", Source: "ft", Xid: "989556146", Currency: "USD", UCITS: true},
	{ID: "GG00BQBFY362", Aliases: []string{"BHMG"}, ISIN: "GG00BQBFY362", Name: "BH Macro Ltd GBP", Source: "yahoo", Symbol: "BHMG.L", Currency: "GBp"},
	{ID: "LU0319687124", Aliases: []string{"AMUNDI-VOLATILITY", "AMUNDI-VOLATILITY-WORLD"}, ISIN: "LU0319687124", Name: "Amundi Funds - Volatility World A USD (C)", Source: "ft", Xid: "10219357", Currency: "USD", Fees: 1.49, UCITS: true},
	{ID: "LU0171310443", ISIN: "LU0171310443", Name: "BlackRock Global Funds - World Technology Fund A2", Source: "ft", Xid: "28295854", Currency: "EUR", Fees: 1.81, UCITS: true},
	{ID: "LU0171307068", ISIN: "LU0171307068", Name: "BGF World Healthscience A2", Source: "morningstar", Symbol: "0P0000VHO6", Currency: "USD", UCITS: true},
	{ID: "LU0280435461", ISIN: "LU0280435461", Name: "Pictet-Clean Energy Transition R EUR", Source: "ft", Xid: "129516373", Currency: "EUR", Fees: 2.70, UCITS: true},
	{ID: "FR0011147594", ISIN: "FR0011147594", Name: "Omnibond R", Source: "ft", Xid: "121432102", Currency: "EUR", UCITS: true},
	{ID: "FR0012336683", ISIN: "FR0012336683", Name: "Amundi Actions Or PC", Source: "yahoo", Symbol: "0P000163EJ.F", Currency: "EUR", Fees: 1.70, UCITS: true},
	// Part distribuante: Yahoo (clôtures ajustées, dividendes réinvestis)
	// plutôt que la VL FT qui les ignorerait.
	{ID: "DE0002635307", Aliases: []string{"DJXXF"}, ISIN: "DE0002635307", Name: "iShares STOXX Europe 600 UCITS ETF (DE) EUR (Dist)", Source: "yahoo", Symbol: "EXSA.DE", Currency: "EUR", Fees: 0.20, UCITS: true},
	{ID: "XS3022291473", Aliases: []string{"CRRY"}, ISIN: "XS3022291473", Name: "WisdomTree Enhanced Commodity Carry ETC", Source: "ft", Xid: "984123045", Currency: "USD", Fees: 0.34},
	{ID: "IE00B5BMR087", ISIN: "IE00B5BMR087", Name: "iShares Core S&P 500 UCITS ETF USD (Acc)", Source: "yahoo", Symbol: "CSSPX.MI", Currency: "EUR", Fees: 0.07, UCITS: true},
	{ID: "IE00B3YCGJ38", ISIN: "IE00B3YCGJ38", Name: "Invesco S&P 500 UCITS ETF Acc", Source: "ft", Xid: "24497790", Currency: "USD", Fees: 0.05, UCITS: true},
	{ID: "IE00B579F325", ISIN: "IE00B579F325", Name: "Invesco Physical Gold ETC", Source: "yahoo", Symbol: "SGLD.L", Currency: "USD", Fees: 0.12},
	{ID: "LU1681043599", Aliases: []string{"CW8"}, ISIN: "LU1681043599", Name: "Amundi MSCI World UCITS ETF EUR Acc (CW8)", Source: "yahoo", Symbol: "CW8.PA", Currency: "EUR", Fees: 0.38, UCITS: true},
	{ID: "FR0011871128", ISIN: "FR0011871128", Name: "Amundi PEA S&P 500 UCITS ETF Acc", Source: "yahoo", Symbol: "PSP5.PA", Currency: "EUR", Fees: 0.12, UCITS: true},
	{ID: "LU0131510165", ISIN: "LU0131510165", Name: "Indépendance AM - France Small & Mid A (C)", Source: "ft", Xid: "8542", Currency: "EUR", Fees: 2.16, UCITS: true},
	{ID: "LU1832174962", ISIN: "LU1832174962", Name: "Indépendance AM - Europe Small A (C)", Source: "ft", Xid: "118135654", Currency: "EUR", Fees: 2.15, UCITS: true},
	{ID: "LU2798962978", ISIN: "LU2798962978", Name: "Indépendance AM - Europe Mid A (C)", Source: "ft", Xid: "936737322", Currency: "EUR", UCITS: true},
	{ID: "WPEA", ISIN: "IE0002XZSHO1", Name: "iShares MSCI World Swap PEA UCITS ETF EUR (Acc)", Source: "yahoo", Symbol: "WPEA.PA", Currency: "EUR", Fees: 0.20, UCITS: true},
	{ID: "SPEA", ISIN: "IE000DQLYVB9", Name: "iShares S&P 500 Swap PEA UCITS ETF EUR (Acc)", Source: "ft", Symbol: "SPEA", Xid: "990931048", Currency: "EUR", Fees: 0.10, UCITS: true},

	// Fonds de la liste embarquée (résolutions et TER moissonnés par
	// --warmup puis figés ici; voir data/fund_tickers.csv pour les tickers).
	{ID: "IE00BK5BQT80", ISIN: "IE00BK5BQT80", Name: "Vanguard FTSE All-World UCITS ETF", Source: "yahoo", Symbol: "VWRA.L", Currency: "USD", Fees: 0.19, UCITS: true},
	{ID: "IE00B3RBWM25", ISIN: "IE00B3RBWM25", Name: "Vanguard FTSE All-World UCITS ETF", Source: "yahoo", Symbol: "VWRD.L", Currency: "USD", Fees: 0.19, UCITS: true},
	{ID: "IE0003XJA0J9", ISIN: "IE0003XJA0J9", Name: "Amundi Prime All Country World UCITS ETF USD Dist", Source: "ft", Symbol: "WEBN", Xid: "894306203", Currency: "MUN", Fees: 0.07, UCITS: true},
	{ID: "IE00BJ0KDQ92", ISIN: "IE00BJ0KDQ92", Name: "Xtrackers MSCI World UCITS ETF 1C", Source: "yahoo", Symbol: "XDWD.L", Currency: "USD", Fees: 0.12, UCITS: true},
	{ID: "IE00BFY0GT14", ISIN: "IE00BFY0GT14", Name: "State Street SPDR MSCI World UCITS ETF", Source: "yahoo", Symbol: "SPPW.DE", Currency: "EUR", Fees: 0.12, UCITS: true},
	{ID: "IE00B4X9L533", ISIN: "IE00B4X9L533", Name: "HSBC MSCI World UCITS ETF", Source: "yahoo", Symbol: "HMWD.L", Currency: "USD", Fees: 0.15, UCITS: true},
	{ID: "IE00B6R52259", ISIN: "IE00B6R52259", Name: "iShares MSCI ACWI UCITS ETF USD Acc", Source: "yahoo", Symbol: "SSAC.L", Currency: "GBp", Fees: 0.20, UCITS: true},
	{ID: "IE00B44Z5B48", ISIN: "IE00B44Z5B48", Name: "State Street SPDR MSCI All Country World UCITS ETF", Source: "yahoo", Symbol: "SPYY.DE", Currency: "EUR", Fees: 0.12, UCITS: true},
	{ID: "IE00B3YLTY66", ISIN: "IE00B3YLTY66", Name: "State Street SPDR MSCI All Country World Investable Market UCITS ETF", Source: "yahoo", Symbol: "SPYI.DE", Currency: "EUR", Fees: 0.17, UCITS: true},
	{ID: "IE000716YHJ7", ISIN: "IE000716YHJ7", Name: "Invesco FTSE All-World UCITS ETF USD Accumalation", Source: "ft", Symbol: "FWRA", Xid: "808530603", Currency: "USD", Fees: 0.15, UCITS: true},
	{ID: "LU2655993207", ISIN: "LU2655993207", Name: "Amundi Index Solutions - Amundi MSCI World Swap UCITS ETF EUR Dist", Source: "yahoo", Symbol: "EWLD.PA", Currency: "EUR", Fees: 0.38, UCITS: true},
	{ID: "IE00BKX55T58", ISIN: "IE00BKX55T58", Name: "Vanguard FTSE Developed World UCITS ETF USD Distributing", Source: "yahoo", Symbol: "VEVE.L", Currency: "GBP", Fees: 0.12, UCITS: true},
	{ID: "IE00BK5BQV03", ISIN: "IE00BK5BQV03", Name: "Vanguard FTSE Developed World UCITS ETF USD Accumulation", Source: "yahoo", Symbol: "VHVE.L", Currency: "USD", Fees: 0.12, UCITS: true},
	{ID: "FR0010315770", ISIN: "FR0010315770", Name: "Amundi MSCI World Swap II UCITS ETF Dist", Source: "ft", Symbol: "WLD", Xid: "4707035", Currency: "EUR", Fees: 0.30, UCITS: true},
	{ID: "IE00BF4RFH31", ISIN: "IE00BF4RFH31", Name: "iShares MSCI World Small Cap UCITS ETF USD (Acc)", Source: "yahoo", Symbol: "WSML.L", Currency: "USD", Fees: 0.35, UCITS: true},
	{ID: "FR001400U5Q4", ISIN: "FR001400U5Q4", Name: "AMUNDI PEA MONDE MSCI World UCI", Source: "yahoo", Symbol: "DCAM.PA", Currency: "EUR", Fees: 0.20},
	{ID: "FR0011550185", ISIN: "FR0011550185", Name: "BNP Paribas Easy S&P 500 UCITS ETF EUR C", Source: "yahoo", Symbol: "ESE.PA", Currency: "EUR", Fees: 0.14, UCITS: true},
	{ID: "FR0013412285", ISIN: "FR0013412285", Name: "Amundi PEA S&P 500 Screened UCITS ETF - Acc", Source: "yahoo", Symbol: "PE500.PA", Currency: "EUR", Fees: 0.25, UCITS: true},
	{ID: "FR0011871110", ISIN: "FR0011871110", Name: "Amundi PEA Nasdaq-100 UCITS ETF Acc", Source: "yahoo", Symbol: "PUST.PA", Currency: "EUR", Fees: 0.30, UCITS: true},
	{ID: "FR0013412020", ISIN: "FR0013412020", Name: "Amundi PEA Emergent (MSCI Emerging) ESG Transition UCITS ETF Acc", Source: "yahoo", Symbol: "PAEEM.PA", Currency: "EUR", Fees: 0.30, UCITS: true},
	{ID: "FR0013412038", ISIN: "FR0013412038", Name: "Amundi PEA MSCI Europe UCITS ETF Acc", Source: "yahoo", Symbol: "PCEU.PA", Currency: "EUR", Fees: 0.15, UCITS: true},
	{ID: "FR0007052782", ISIN: "FR0007052782", Name: "Amundi CAC 40 UCITS ETF Dist", Source: "ft", Symbol: "CAC", Xid: "69366", Currency: "EUR", Fees: 0.25, UCITS: true},
	{ID: "IE0031442068", ISIN: "IE0031442068", Name: "iShares Core S&P 500 UCITS ETF USD Dist", Source: "ft", Symbol: "IDUS", Xid: "5495140", Currency: "USD", Fees: 0.07, UCITS: true},
	{ID: "IE00B3XXRP09", ISIN: "IE00B3XXRP09", Name: "Vanguard S&P 500 UCITS ETF", Source: "yahoo", Symbol: "VUSA.L", Currency: "GBP", Fees: 0.07, UCITS: true},
	{ID: "IE00BFMXXD54", ISIN: "IE00BFMXXD54", Name: "Vanguard S&P 500 UCITS ETF USD Accumulation", Source: "yahoo", Symbol: "VUAA.L", Currency: "USD", Fees: 0.07, UCITS: true},
	{ID: "IE000XZSV718", ISIN: "IE000XZSV718", Name: "State Street SPDR S&P 500 UCITS ETF USD Acc", Source: "yahoo", Symbol: "SPYL.L", Currency: "USD", Fees: 0.03, UCITS: true},
	{ID: "IE00B6YX5C33", ISIN: "IE00B6YX5C33", Name: "State Street SPDR S&P 500 UCITS ETF", Source: "yahoo", Symbol: "SPY5.DE", Currency: "EUR", Fees: 0.03, UCITS: true},
	{ID: "LU0496786574", ISIN: "LU0496786574", Name: "Amundi Core S&P 500 Swap UCITS ETF EUR Dist", Source: "yahoo", Symbol: "LYPS.DE", Currency: "EUR", Fees: 0.05, UCITS: true},
	{ID: "IE00B3WJKG14", ISIN: "IE00B3WJKG14", Name: "iShares S&P 500 Information Technology Sector UCITS ETF USD (Acc)", Source: "yahoo", Symbol: "IITU.L", Currency: "GBp", Fees: 0.15, UCITS: true},
	{ID: "IE0032077012", ISIN: "IE0032077012", Name: "Invesco EQQQ NASDAQ-100 UCITS ETF", Source: "yahoo", Symbol: "EQQQ.MI", Currency: "EUR", Fees: 0.30, UCITS: true},
	{ID: "IE00B53SZB19", ISIN: "IE00B53SZB19", Name: "iShares VII PLC - iShares NASDAQ 100 UCITS ETF", Source: "yahoo", Symbol: "CSNDX.SW", Currency: "USD", Fees: 0.30, UCITS: true},
	{ID: "LU1681038243", ISIN: "LU1681038243", Name: "Amundi Index Solutions - Amundi Nasdaq-100 Swap ETF EUR Acc", Source: "yahoo", Symbol: "ANX.PA", Currency: "EUR", Fees: 0.23},
	{ID: "IE00BMC38736", ISIN: "IE00BMC38736", Name: "VanEck Semiconductor UCITS ETF", Source: "yahoo", Symbol: "SMH.L", Currency: "USD", Fees: 0.35, UCITS: true},
	{ID: "IE00BGV5VN51", ISIN: "IE00BGV5VN51", Name: "Xtrackers Artificial Intelligence & Big Data UCITS ETF 1C", Source: "ft", Symbol: "XAIX", Xid: "515873934", Currency: "EUR", Fees: 0.35, UCITS: true},
	{ID: "LU0908500753", ISIN: "LU0908500753", Name: "Amundi Core Stoxx Europe 600 UCITS ETF Acc", Source: "ft", Symbol: "MEUD", Xid: "57210679", Currency: "EUR", Fees: 0.07, UCITS: true},
	{ID: "LU0328475792", ISIN: "LU0328475792", Name: "Xtrackers Stoxx Europe 600 UCITS ETF 1C", Source: "yahoo", Symbol: "XSX6.DE", Currency: "EUR", Fees: 0.20, UCITS: true},
	{ID: "IE00B4K48X80", ISIN: "IE00B4K48X80", Name: "iShares Core MSCI Europe UCITS ETF EUR (Acc)", Source: "yahoo", Symbol: "IMEA.SW", Currency: "CHF", Fees: 0.12, UCITS: true},
	{ID: "IE00B945VV12", ISIN: "IE00B945VV12", Name: "Vanguard FTSE Developed Europe UCITS ETF", Source: "yahoo", Symbol: "VEUR.AS", Currency: "EUR", Fees: 0.10, UCITS: true},
	{ID: "IE00B52VJ196", ISIN: "IE00B52VJ196", Name: "iShares MSCI Europe SRI UCITS ETF EUR (Acc)", Source: "yahoo", Symbol: "IESG.L", Currency: "GBp", Fees: 0.20, UCITS: true},
	{ID: "DE0005933931", ISIN: "DE0005933931", Name: "iShares Core DAX® UCITS ETF (DE) EUR (Acc)", Source: "ft", Symbol: "EXS1", Xid: "136514", Currency: "EUR", Fees: 0.16, UCITS: true},
	{ID: "LU0274211480", ISIN: "LU0274211480", Name: "Xtrackers DAX UCITS ETF 1C", Source: "ft", Symbol: "DBXD", Xid: "6107331", Currency: "EUR", Fees: 0.09, UCITS: true},
	{ID: "IE00BKM4GZ66", ISIN: "IE00BKM4GZ66", Name: "iShares Core MSCI EM IMI UCITS ETF USD (Acc)", Source: "yahoo", Symbol: "EIMI.L", Currency: "USD", Fees: 0.18, UCITS: true},
	{ID: "IE00B3VVMM84", ISIN: "IE00B3VVMM84", Name: "Vanguard FTSE Emerging Markets UCITS ETF USD Distributing", Source: "yahoo", Symbol: "VFEM.L", Currency: "GBP", Fees: 0.17, UCITS: true},
	{ID: "IE00BK5BR733", ISIN: "IE00BK5BR733", Name: "Vanguard FTSE Emerging Markets UCITS ETF USD Accumulation", Source: "yahoo", Symbol: "VFEA.DE", Currency: "EUR", Fees: 0.17, UCITS: true},
	{ID: "LU1681045370", ISIN: "LU1681045370", Name: "Amundi Index Solutions - Amundi MSCI Emerging Markets Swap UCITS ETF EUR Acc", Source: "yahoo", Symbol: "AEEM.PA", Currency: "EUR", Fees: 0.20, UCITS: true},
	{ID: "IE00BTJRMP35", ISIN: "IE00BTJRMP35", Name: "Xtrackers MSCI Emerging Markets UCITS ETF 1C", Source: "yahoo", Symbol: "XMME.L", Currency: "USD", Fees: 0.18, UCITS: true},
	{ID: "LU0290358497", ISIN: "LU0290358497", Name: "Xtrackers II EUR Overnight Rate Swap UCITS ETF 1C", Source: "ft", Symbol: "XEON", Xid: "7140105", Currency: "EUR", Fees: 0.10, UCITS: true},
	{ID: "IE00BDBRDM35", ISIN: "IE00BDBRDM35", Name: "ISHARES III PLC ISH GLOBAL AGG ", Source: "yahoo", Symbol: "0GGH.L", Currency: "EUR", Fees: 0.10},
	{ID: "IE00B3DKXQ41", ISIN: "IE00B3DKXQ41", Name: "iShares € Aggregate Bond ESG SRI UCITS ETF EUR (Dist)", Source: "yahoo", Symbol: "SEAG.L", Currency: "GBP", Fees: 0.16, UCITS: true},
	{ID: "IE00BGSF1X88", ISIN: "IE00BGSF1X88", Name: "iShares $ Treasury Bond 0-1yr UCITS ETF", Source: "yahoo", Symbol: "IB01.L", Currency: "USD", Fees: 0.07, UCITS: true},
	{ID: "IE00BSKRJZ44", ISIN: "IE00BSKRJZ44", Name: "iShares $ Treasury Bond 20+yr UCITS ETF", Source: "yahoo", Symbol: "IDTL.L", Currency: "USD", Fees: 0.07, UCITS: true},
	{ID: "JE00B1VS3770", ISIN: "JE00B1VS3770", Name: "WisdomTree Physical Gold", Source: "ft", Symbol: "PHAU", Xid: "7299997", Currency: "USD", Fees: 0.39},
	{ID: "DE000A0S9GB0", ISIN: "DE000A0S9GB0", Name: "Xetra-Gold", Source: "ft", Symbol: "4GLD", Xid: "9894153", Currency: "EUR"},
	{ID: "IE00B4ND3602", ISIN: "IE00B4ND3602", Name: "iShares Physical Gold ETC", Source: "ft", Symbol: "IGLN", Xid: "33139564", Currency: "USD", Fees: 0.12},
	{ID: "IE00B8GKDB10", ISIN: "IE00B8GKDB10", Name: "Vanguard FTSE All-World High Dividend Yield UCITS ETF USD Distributing", Source: "yahoo", Symbol: "VHYL.L", Currency: "GBP", Fees: 0.29, UCITS: true},
	{ID: "NL0011683594", ISIN: "NL0011683594", Name: "VanEck Morningstar Developed Markets Dividend Leaders UCITS ETF", Source: "ft", Symbol: "TDIV", Xid: "97639063", Currency: "EUR", Fees: 0.38, UCITS: true},
	{ID: "IE00BSPLC413", ISIN: "IE00BSPLC413", Name: "State Street SPDR MSCI USA Small Cap Value Weighted UCITS ETF USD Acc", Source: "yahoo", Symbol: "USSC.L", Currency: "USD", Fees: 0.30, UCITS: true},
	{ID: "IE00BSPLC298", ISIN: "IE00BSPLC298", Name: "State Street SPDR MSCI Europe Small Cap Value Weighted UCITS ETF EUR Acc", Source: "yahoo", Symbol: "ZPRX.DE", Currency: "EUR", Fees: 0.30, UCITS: true},
	{ID: "IE00BP3QZ825", ISIN: "IE00BP3QZ825", Name: "iShares Edge MSCI World Momentum Factor UCITS ETF USD (Acc)", Source: "ft", Symbol: "IWMO", Xid: "79108371", Currency: "USD", Fees: 0.25, UCITS: true},
	{ID: "IE00BP3QZ601", ISIN: "IE00BP3QZ601", Name: "iShares Edge MSCI World Quality Factor UCITS ETF USD (Acc)", Source: "ft", Symbol: "IWQU", Xid: "79108372", Currency: "USD", Fees: 0.25, UCITS: true},
	{ID: "IE00BP3QZB59", ISIN: "IE00BP3QZB59", Name: "iShares Edge MSCI World Value Factor UCITS ETF USD (Acc)", Source: "ft", Symbol: "IWVL", Xid: "79108374", Currency: "USD", Fees: 0.25, UCITS: true},
	{ID: "IE0003R87OG3", ISIN: "IE0003R87OG3", Name: "Avantis Global Small Cap Value UCITS ETF USD Acc", Source: "ft", Symbol: "AVWS", Xid: "935302649", Currency: "EUR", Fees: 0.39, UCITS: true},
	{ID: "IE00BJRCLL96", ISIN: "IE00BJRCLL96", Name: "JPM Global Equity Multi-Factor UCITS ETF USD Acc", Source: "ft", Symbol: "JPGL", Xid: "542974689", Currency: "EUR", Fees: 0.19, UCITS: true},
	{ID: "IE00BMVB5R75", ISIN: "IE00BMVB5R75", Name: "Vanguard LifeStrategy 80% Equity UCITS ETF (EUR) Accumulating", Source: "ft", Symbol: "V80A", Xid: "633103242", Currency: "EUR", Fees: 0.25, UCITS: true},
	{ID: "IE00BMVB5P51", ISIN: "IE00BMVB5P51", Name: "Vanguard LifeStrategy 60% Equity UCITS ETF (EUR) Accumulating", Source: "ft", Symbol: "V60A", Xid: "633308512", Currency: "EUR", Fees: 0.25, UCITS: true},
	{ID: "IE00B1XNHC34", ISIN: "IE00B1XNHC34", Name: "iShares Global Clean Energy Transition UCITS ETF USD (Dist)", Source: "ft", Symbol: "IQQH", Xid: "9155309", Currency: "EUR", Fees: 0.65, UCITS: true},
	{ID: "FR0010755611", ISIN: "FR0010755611", Name: "Amundi MSCI USA Daily (2x) Leveraged UCITS ETF Acc", Source: "yahoo", Symbol: "CL2.PA", Currency: "EUR", Fees: 0.50, UCITS: true},
	{ID: "FR0010342592", ISIN: "FR0010342592", Name: "Amundi Nasdaq-100 Daily (2x) Leveraged UCITS ETF Acc", Source: "ft", Symbol: "LQQ", Xid: "5082783", Currency: "EUR", Fees: 0.60, UCITS: true},

	// US-listed twins, used by simgen for validation; the bare tickers map
	// to the UCITS share classes via the embedded fund list.
	{ID: "NTSX-US", Name: "WisdomTree US Efficient Core ETF (version US)", Source: "yahoo", Symbol: "NTSX", Currency: "USD", Fees: 0.20},
}

var catalogByID = sync.OnceValue(func() map[string]CatalogEntry {
	m := make(map[string]CatalogEntry, 2*len(catalog))
	for _, e := range catalog {
		m[e.ID] = e
		if e.ISIN != "" {
			m[e.ISIN] = e
		}
	}
	return m
})

// catalogResolution returns the pinned resolution for a canonical id.
func catalogResolution(id string) (resolution, bool) {
	e, ok := catalogByID()[id]
	if !ok {
		return resolution{}, false
	}
	return resolution{Source: e.Source, Symbol: e.Symbol, Xid: e.Xid, Name: e.Name, Currency: e.Currency}, true
}

// UCITSFlag reports whether a catalogued asset is a UCITS fund; known is
// false when the identifier is not in the catalog.
func UCITSFlag(id string) (ucits, known bool) {
	e, found := catalogByID()[CanonicalID(id)]
	if !found {
		return false, false
	}
	return e.UCITS, true
}

// GuessUCITS extends UCITSFlag to uncatalogued assets with a heuristic on
// the resolved name (UCITS funds advertise it in their share-class names).
func GuessUCITS(id, name string) (ucits, known bool) {
	if u, k := UCITSFlag(id); k {
		return u, true
	}
	if strings.Contains(strings.ToUpper(name), "UCITS") {
		return true, true
	}
	return false, false
}

// LooksDistributing reports whether a fund share-class name suggests a
// distributing class, whose NAV series (FT, Morningstar) excludes the
// dividends it pays out.
func LooksDistributing(name string) bool {
	n := strings.ToLower(name)
	for _, marker := range []string{"dist", "(dis)", " dis ", "(d)", " inc", "dy)"} {
		if strings.Contains(n, marker) {
			return true
		}
	}
	return false
}

// WarmupIDs lists the canonical ID of every catalog entry: the catalog IS
// the bundle of assets whose data is precomputed or one --warmup away.
func WarmupIDs() []string {
	ids := make([]string, 0, len(catalog))
	for _, e := range catalog {
		ids = append(ids, e.ID)
	}
	sort.Strings(ids)
	return ids
}

//go:embed data/fund_tickers.csv
var fundTickersCSV string

// fundISINByTicker maps exchange tickers of common European funds and ETFs
// to their ISIN, from the embedded list. US tickers are deliberately absent:
// they resolve directly on Yahoo.
var fundISINByTicker = sync.OnceValue(func() map[string]string {
	m := make(map[string]string)
	sc := bufio.NewScanner(strings.NewReader(fundTickersCSV))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, ";")
		if len(parts) != 3 || !IsISIN(parts[0]) {
			continue
		}
		for _, t := range strings.Fields(strings.ToUpper(parts[2])) {
			m[t] = parts[0]
		}
	}
	return m
})

// FundISIN maps a European fund/ETF exchange ticker to its ISIN using the
// embedded correspondence list.
func FundISIN(ticker string) (string, bool) {
	isin, ok := fundISINByTicker()[strings.ToUpper(strings.TrimSpace(ticker))]
	return isin, ok
}
