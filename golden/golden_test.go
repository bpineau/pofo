package golden

import (
	"bufio"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"portfodor/pkg/marketdata"
	"portfodor/pkg/metrics"
	"portfodor/pkg/portfolio"
)

// Références externes (établies et validées le 2026-06-12) :
//   - CAGR : rendements annuels officiels S&P 500 total return (S&P DJ via
//     Wikipedia) composés sur chaque fenêtre ; l'écart attendu vs SPY est
//     frais (0,09 %/an) + journée de départ, soit ≲ 0,35 point.
//   - MSCI World gross (USD) via Wikipedia ; URTH ≈ gross − 0,4–0,9 point
//     (retenues à la source + TER 0,24 %).
//   - Max Drawdown quotidiens canoniques : GFC −55,25 % (S&P TR), COVID
//     −33,7 %.
//   - Sharpe/Sortino : conventions maison (quotidien, rf = 0) — bornes
//     larges destinées à détecter les régressions de calcul, pas les
//     conventions.
type window struct {
	name     string
	from, to string
	check    func(t *testing.T, s metrics.Stats)
}

func TestGoldenSPY(t *testing.T) {
	dates, values := loadFixture(t, "testdata/spy.csv")
	sim := simulate(t, "SPY", dates, values)

	for _, w := range []window{
		{
			name: "2006-2025 (20 ans, GFC incluse)", from: "2006-01-01", to: "2025-12-31",
			check: func(t *testing.T, s metrics.Stats) {
				within(t, "CAGR", s.CAGR*100, 11.00, 0.40)         // réf indice TR 11.00 %
				within(t, "MaxDD", s.MaxDrawdown*100, -55.25, 1.0) // GFC quotidien canonique
				within(t, "Volatilité", s.Volatility*100, 19.4, 1.5)
				within(t, "Sharpe", s.Sharpe, 0.63, 0.08)
				within(t, "Sortino", s.Sortino, 0.89, 0.10)
				within(t, "Ulcer", s.Ulcer, 12.8, 1.5)
				// Récupération TR de la GFC: ~4,5–5 ans avec les frais SPY.
				if s.TTRDays < 1600 || s.TTRDays > 1900 {
					t.Errorf("TTR = %d j, attendu ~1773 j", s.TTRDays)
				}
			},
		},
		{
			name: "2010-2019 (décennie haussière)", from: "2010-01-01", to: "2019-12-31",
			check: func(t *testing.T, s metrics.Stats) {
				within(t, "CAGR", s.CAGR*100, 13.56, 0.45) // réf 13.56 %, −0.30 attendu (1er jour +1.7 %)
				within(t, "MaxDD", s.MaxDrawdown*100, -19.4, 1.0)
			},
		},
		{
			name: "2015-2024 (COVID incluse)", from: "2015-01-01", to: "2024-12-31",
			check: func(t *testing.T, s metrics.Stats) {
				within(t, "CAGR", s.CAGR*100, 13.10, 0.25)         // réf 13.10 %
				within(t, "MaxDD", s.MaxDrawdown*100, -33.72, 0.8) // COVID quotidien canonique
				within(t, "Volatilité", s.Volatility*100, 17.6, 1.2)
				within(t, "Sharpe", s.Sharpe, 0.78, 0.08)
			},
		},
		{
			name: "2020-2024 (5 ans)", from: "2020-01-01", to: "2024-12-31",
			check: func(t *testing.T, s metrics.Stats) {
				within(t, "CAGR", s.CAGR*100, 14.53, 0.45) // réf 14.53 %
			},
		},
	} {
		t.Run(w.name, func(t *testing.T) {
			w.check(t, computeWindow(t, sim, w.from, w.to))
		})
	}
}

func TestGoldenURTH(t *testing.T) {
	dates, values := loadFixture(t, "testdata/urth.csv")
	sim := simulate(t, "URTH", dates, values)

	for _, w := range []window{
		{
			name: "2013-2024", from: "2013-01-01", to: "2024-12-31",
			check: func(t *testing.T, s metrics.Stats) {
				// MSCI World gross 11.40 %; URTH attendu dessous de 0,4–0,9.
				within(t, "CAGR", s.CAGR*100, 10.90, 0.35)
			},
		},
		{
			name: "2015-2024", from: "2015-01-01", to: "2024-12-31",
			check: func(t *testing.T, s metrics.Stats) {
				within(t, "CAGR", s.CAGR*100, 10.13, 0.35) // gross 10.52 %
				within(t, "Volatilité", s.Volatility*100, 17.3, 1.2)
			},
		},
	} {
		t.Run(w.name, func(t *testing.T) {
			w.check(t, computeWindow(t, sim, w.from, w.to))
		})
	}
}

// simulate rejoue la chaîne réelle: un portefeuille mono-actif passé par
// portfolio.Simulate (rebalancement sans effet ici), pour couvrir la
// simulation et pas seulement le calcul de métriques.
func simulate(t *testing.T, symbol string, dates []time.Time, values []float64) *portfolio.SimResult {
	t.Helper()
	series := &marketdata.Series{Symbol: symbol, Name: symbol}
	for i := range dates {
		series.Points = append(series.Points, marketdata.Point{Date: dates[i], Close: values[i]})
	}
	p := &portfolio.Portfolio{
		Name:   symbol,
		Assets: []portfolio.Asset{{ID: symbol, Symbol: symbol, Weight: 1, Fees: -1, Series: series}},
	}
	sim, err := portfolio.Simulate(p, 90)
	if err != nil {
		t.Fatal(err)
	}
	return sim
}

func computeWindow(t *testing.T, sim *portfolio.SimResult, from, to string) metrics.Stats {
	t.Helper()
	f := mustDate(t, from)
	o := mustDate(t, to)
	i := sort.Search(len(sim.Dates), func(k int) bool { return !sim.Dates[k].Before(f) })
	j := sort.Search(len(sim.Dates), func(k int) bool { return sim.Dates[k].After(o) })
	if j-i < 2 {
		t.Fatalf("fenêtre vide %s → %s", from, to)
	}
	stats, err := metrics.Compute(sim.Dates[i:j], sim.Values[i:j])
	if err != nil {
		t.Fatal(err)
	}
	return stats
}

func within(t *testing.T, name string, got, want, tol float64) {
	t.Helper()
	if math.IsNaN(got) || math.Abs(got-want) > tol {
		t.Errorf("%s = %.3f, référence %.3f (tolérance ±%.2f)", name, got, want, tol)
	}
}

func mustDate(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := time.ParseInLocation("2006-01-02", s, time.UTC)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func loadFixture(t *testing.T, path string) ([]time.Time, []float64) {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var dates []time.Time
	var values []float64
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") || line == "date,close" {
			continue
		}
		ds, cs, _ := strings.Cut(line, ",")
		c, err := strconv.ParseFloat(cs, 64)
		if err != nil {
			t.Fatalf("fixture %s: %v", path, err)
		}
		dates = append(dates, mustDate(t, ds))
		values = append(values, c)
	}
	if len(dates) < 1000 {
		t.Fatalf("fixture %s tronquée: %d points", path, len(dates))
	}
	return dates, values
}
