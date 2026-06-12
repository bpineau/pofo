// Command portfodor reads portfolio description files, downloads the price
// history of each asset, simulates the portfolios with periodic rebalancing
// and produces a self-contained HTML report comparing them.
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log"
	"math"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"portfodor/chart"
	"portfodor/marketdata"
	"portfodor/metrics"
	"portfodor/portfolio"
	"portfodor/report"
)

var palette = []string{
	"#1f77b4", "#ff7f0e", "#2ca02c", "#d62728",
	"#9467bd", "#8c564b", "#e377c2", "#17becf",
}

func main() {
	log.SetFlags(0)
	if err := run(os.Args[1:]); err != nil {
		log.Fatal("portfodor: ", err)
	}
}

type options struct {
	out        string
	dataDir    string
	simdataDir string
	rebalance  int
	start      time.Time
	end        time.Time // zéro = jusqu'à aujourd'hui
	benchmark  string
	noOpen     bool
	noSim      bool
	noFees     bool
	cli        bool
	width      int
	cacheAge   time.Duration
}

// result holds everything computed for one portfolio.
type result struct {
	p             *portfolio.Portfolio
	sim           *portfolio.SimResult
	color         string
	rebalanceDays int
	// Common-window view, renormalized to 100, used for stats and comparison.
	winDates  []time.Time
	winValues []float64
	stats     metrics.Stats
}

func run(argv []string) error {
	fs := flag.NewFlagSet("portfodor", flag.ContinueOnError)
	var opt options
	var startStr string
	fs.StringVar(&opt.out, "out", "", "fichier HTML de sortie (défaut: /tmp/portfodor-<horodatage>.html)")
	fs.StringVar(&opt.dataDir, "data", "data", "répertoire de cache des cotations")
	fs.StringVar(&opt.simdataDir, "simdata", "simdata", "répertoire des historiques simulés permanents")
	fs.IntVar(&opt.rebalance, "rebalance", 90, "rebalancement tous les N jours calendaires (0 = jamais)")
	fs.StringVar(&startStr, "start", "2006-01-01", "date de début souhaitée (AAAA-MM-JJ)")
	var endStr string
	fs.StringVar(&endStr, "end", "", "date de fin (AAAA-MM-JJ, défaut: dernière cotation disponible)")
	fs.StringVar(&opt.benchmark, "benchmark", "^GSPC", "symbole de référence pour le Beta (vide = pas de Beta)")
	fs.BoolVar(&opt.noOpen, "no-open", false, "ne pas ouvrir le rapport dans le navigateur")
	fs.BoolVar(&opt.noSim, "no-simulate", false, "ignorer les suffixes SIM: données réelles uniquement")
	fs.BoolVar(&opt.noFees, "no-fees", false, "ne pas récupérer les frais courants (TER) des actifs")
	fs.BoolVar(&opt.cli, "cli", false, "affichage dans le terminal (courbes + tableau récapitulatif), sans HTML")
	fs.IntVar(&opt.width, "width", 0, "largeur du graphe en mode -cli, en colonnes (défaut: $COLUMNS, sinon 100)")
	fs.DurationVar(&opt.cacheAge, "cache-age", 30*24*time.Hour, "retélécharger les cotations plus vieilles que cette durée")
	warmup := fs.Bool("warmup", false, "précharger le cache pour le catalogue d'actifs intégré, puis s'arrêter")
	assetsList := fs.String("assets", "", "liste de tickers/ISIN séparés par des virgules, chacun comparé comme un portefeuille investi à 100 % dessus")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), `Usage: portfodor [options] portefeuille.txt [portefeuille2.txt …]
       portfodor [options] -assets VOO,IWDA,NTSG

Sans fichier, -assets A,B,C compare chaque actif comme un portefeuille
investi à 100 %% dessus (cumulable avec des fichiers).

Format des fichiers — une ligne par actif :

    <poids en %%> <identifiant> [frais en %%/an] [texte libre]

  - Tout ce qui suit un # est un commentaire ; lignes vides ignorées.
  - Identifiant : ticker US (VOO), ticker européen de la liste embarquée
    (IWDA, CW8, CSPX…), ISIN, ou alias du catalogue (GOLD, NTSG, BHMG…).
  - Suffixe SIM (VOOSIM, DBMFSIM…) : étend l'historique avant la première
    cotation réelle via simdata/ ou un proxy ; identifiant nu = réel seul.
  - 3e colonne numérique optionnelle : TER de l'actif en %%/an (prime sur
    la récupération automatique) ; non numérique = texte libre.
  - Directives par portefeuille :
        #meta rebalance:N    rebalancement tous les N jours (0 = jamais)
        #meta extra-fees:X   frais d'enveloppe en %%/an, déduits de la
                             performance (synonyme: envelope-fees)

Exemple :
    #meta rebalance:30
    #meta extra-fees:0.5
    60   VTI           actions US
    25,5 IE00B4L5Y983  # ISIN ; virgule décimale acceptée
    14.5 GOLDSIM       or, historique étendu avant la 1re cotation

Options :
`)
		fs.PrintDefaults()
	}
	if err := fs.Parse(argv); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	files := fs.Args()
	if len(files) == 0 && *assetsList == "" && !*warmup {
		fs.Usage()
		return errors.New("aucun fichier de portefeuille ni option -assets")
	}
	start, err := time.ParseInLocation("2006-01-02", startStr, time.UTC)
	if err != nil {
		return fmt.Errorf("option -start invalide: %w", err)
	}
	opt.start = start
	if endStr != "" {
		end, err := time.ParseInLocation("2006-01-02", endStr, time.UTC)
		if err != nil {
			return fmt.Errorf("option -end invalide: %w", err)
		}
		if !end.After(opt.start) {
			return errors.New("-end doit être postérieure à -start")
		}
		opt.end = end
	}

	// Parse every portfolio file, disambiguating duplicate names, then add
	// one synthetic 100 % portfolio per -assets entry.
	specs := make([]*portfolio.Spec, 0, len(files))
	nameCount := map[string]int{}
	addSpec := func(spec *portfolio.Spec) {
		nameCount[spec.Name]++
		if n := nameCount[spec.Name]; n > 1 {
			spec.Name = fmt.Sprintf("%s (%d)", spec.Name, n)
		}
		specs = append(specs, spec)
	}
	for _, f := range files {
		spec, err := portfolio.ParseFile(f)
		if err != nil {
			return err
		}
		addSpec(spec)
	}
	for _, id := range strings.Split(*assetsList, ",") {
		if id = strings.TrimSpace(id); id != "" {
			addSpec(portfolio.Single(id))
		}
	}
	if len(specs) == 0 && !*warmup {
		return errors.New("l'option -assets ne contient aucun identifiant")
	}

	client := marketdata.NewClient(opt.dataDir)
	client.MaxAge = opt.cacheAge
	client.Logf = log.Printf

	if *warmup {
		return runWarmup(client, &opt)
	}

	// Download every distinct asset once.
	seriesByID := map[string]*marketdata.Series{}
	for _, spec := range specs {
		for _, h := range spec.Holdings {
			if _, ok := seriesByID[h.ID]; ok {
				continue
			}
			s, err := fetchAsset(client, h.ID, &opt)
			if err != nil {
				return fmt.Errorf("portefeuille %s, actif %q: %w", spec.Name, h.ID, err)
			}
			seriesByID[h.ID] = s
		}
	}

	// Benchmark for Beta, best effort.
	var bench *marketdata.Series
	if opt.benchmark != "" {
		b, err := client.Fetch(opt.benchmark, opt.start)
		if err != nil {
			log.Printf("avertissement: benchmark %s indisponible (pas de Beta): %v", opt.benchmark, err)
		} else {
			bench = b
		}
	}

	// Simulate each portfolio; a "#meta rebalance:N" directive overrides
	// the CLI default for that portfolio only.
	var feesFor func(string) (float64, bool)
	if !opt.noFees {
		feesFor = func(id string) (float64, bool) {
			base, _ := marketdata.SplitSim(id)
			return client.Fees(base)
		}
	}
	results := make([]*result, 0, len(specs))
	for i, spec := range specs {
		p := buildPortfolio(spec, seriesByID, feesFor)
		days := opt.rebalance
		if spec.RebalanceDays >= 0 {
			days = spec.RebalanceDays
		}
		sim, err := portfolio.Simulate(p, days)
		if err != nil {
			return fmt.Errorf("portefeuille %s: %w", spec.Name, err)
		}
		results = append(results, &result{p: p, sim: sim, color: palette[i%len(palette)], rebalanceDays: days})
	}

	// Common window across portfolios: statistics and the comparison chart
	// must cover the same period to be meaningful.
	commonStart := results[0].sim.Dates[0]
	commonEnd := results[0].sim.Dates[len(results[0].sim.Dates)-1]
	for _, r := range results[1:] {
		if f := r.sim.Dates[0]; f.After(commonStart) {
			commonStart = f
		}
		if l := r.sim.Dates[len(r.sim.Dates)-1]; l.Before(commonEnd) {
			commonEnd = l
		}
	}
	if !commonStart.Before(commonEnd) {
		return errors.New("pas de période commune entre les portefeuilles")
	}
	var benchDates []time.Time
	var benchValues []float64
	if bench != nil {
		benchDates, benchValues = seriesSlices(bench)
	}
	for _, r := range results {
		i, j := window(r.sim.Dates, commonStart, commonEnd)
		if j-i < 2 {
			return fmt.Errorf("portefeuille %s: trop peu de points sur la période commune", r.p.Name)
		}
		r.winDates = r.sim.Dates[i:j]
		r.winValues = rebase(r.sim.Values[i:j])
		st, err := metrics.Compute(r.winDates, r.winValues)
		if err != nil {
			return fmt.Errorf("portefeuille %s: %w", r.p.Name, err)
		}
		if bench != nil {
			if beta, ok := metrics.Beta(r.winDates, r.winValues, benchDates, benchValues); ok {
				st.Beta, st.HasBeta = beta, true
			}
		}
		r.stats = st
	}

	if opt.cli {
		return renderCLI(results, &opt, commonStart, commonEnd)
	}

	page := buildPage(results, &opt, bench, commonStart, commonEnd)
	var buf bytes.Buffer
	if err := report.Render(&buf, page); err != nil {
		return err
	}
	outPath := opt.out
	if outPath == "" {
		outPath = fmt.Sprintf("/tmp/portfodor-%s.html", time.Now().Format("20060102-150405"))
	}
	if err := os.WriteFile(outPath, buf.Bytes(), 0o644); err != nil {
		return err
	}
	log.Printf("rapport écrit dans %s", outPath)
	if !opt.noOpen {
		openBrowser(outPath)
	}
	return nil
}

// renderCLI prints the comparison curves and the summary table straight to
// the terminal — quick checks without opening a browser. Per-portfolio
// details are intentionally omitted.
func renderCLI(results []*result, opt *options, commonStart, commonEnd time.Time) error {
	color := os.Getenv("NO_COLOR") == "" && isTerminal(os.Stdout)
	names := make([]string, len(results))
	cmp := make([]chart.Series, len(results))
	for i, r := range results {
		names[i] = r.p.Name
		if len(results) == 1 {
			cmp[i] = chart.Series{Name: r.p.Name, Dates: r.sim.Dates, Values: r.sim.Values}
		} else {
			cmp[i] = chart.Series{Name: r.p.Name, Dates: r.winDates, Values: r.winValues}
		}
	}
	title := "Comparaison (base 100"
	if len(results) == 1 {
		title = results[0].p.Name + " (base 100"
	}
	title += " au " + cmp[0].Dates[0].Format("2006-01-02") + ")"
	fmt.Print(chart.Term(chart.TermOptions{Title: title, Width: termWidth(opt.width), Color: color}, cmp))
	fmt.Println()

	page := &report.Page{
		Title:          "Portefeuilles : " + strings.Join(names, ", "),
		CommonStart:    commonStart.Format("2006-01-02"),
		CommonEnd:      commonEnd.Format("2006-01-02"),
		PortfolioNames: names,
		StatRows:       buildStatRows(results, opt.benchmark),
	}
	return report.RenderText(os.Stdout, page, color)
}

// isTerminal reports whether f is attached to a character device.
func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

// termWidth picks the chart width: the explicit -width flag wins, then
// $COLUMNS (capped), then 100.
func termWidth(flag int) int {
	if flag >= 40 {
		return min(flag, 500)
	}
	if c := os.Getenv("COLUMNS"); c != "" {
		if n, err := strconv.Atoi(c); err == nil && n >= 40 {
			return min(n, 160)
		}
	}
	return 100
}

// runWarmup pre-fetches the whole bundled asset catalog into the cache so
// that later runs work fast and (mostly) offline.
func runWarmup(c *marketdata.Client, opt *options) error {
	ids := marketdata.WarmupIDs()
	var failed []string
	for i, id := range ids {
		if i > 0 {
			time.Sleep(300 * time.Millisecond) // ménage les sources
		}
		s, err := fetchAsset(c, id, opt)
		if err != nil {
			log.Printf("ÉCHEC %s: %v", id, err)
			failed = append(failed, id)
			continue
		}
		feesText := ""
		if !opt.noFees {
			if ter, ok := c.Fees(id); ok {
				feesText = fmt.Sprintf("  TER %.2f %%", ter)
			}
		}
		log.Printf("OK    %-24s %s → %s  (%d cotations)%s %s", id,
			s.First().Date.Format("2006-01-02"), s.Last().Date.Format("2006-01-02"),
			len(s.Points), feesText, s.Name)
	}
	log.Printf("warmup terminé: %d/%d actifs en cache", len(ids)-len(failed), len(ids))
	if len(failed) > 0 {
		log.Printf("en échec: %s", strings.Join(failed, ", "))
	}
	return nil
}

// fetchAsset downloads the history of an identifier (ticker or ISIN). A
// bare identifier sticks to the asset's real quotes, from its actual
// inception. A "SIM"-suffixed identifier (DBMFSIM, VOOSIM…) additionally
// extends the uncovered period backwards: first with the permanent simulated
// series stored under simdata/, then with a known index/fund proxy — real
// quotes always win wherever they exist.
func fetchAsset(c *marketdata.Client, id string, opt *options) (*marketdata.Series, error) {
	from, allowSim, simdataDir := opt.start, !opt.noSim, opt.simdataDir
	base, wantSim := marketdata.SplitSim(id)
	if !wantSim || !allowSim {
		s, err := c.Fetch(base, from)
		if err != nil {
			return nil, err
		}
		return trim(s, time.Time{}, opt.end), nil
	}
	canonical := marketdata.CanonicalID(base)
	sim, simOK, simErr := marketdata.ReadSimdata(simdataDir, canonical)
	if simErr != nil {
		log.Printf("avertissement: simdata %s illisible: %v", canonical, simErr)
	}
	if simOK {
		sim = trim(sim, from, time.Time{})
		simOK = len(sim.Points) >= 2
	}
	s, err := c.Fetch(base, from)
	if err != nil {
		if simOK {
			log.Printf("avertissement: %s indisponible (%v) — utilisation des seules données simulées", base, err)
			sim.SimulatedBefore = sim.Last().Date
			sim.ProxySymbol = "simdata"
			return trim(sim, time.Time{}, opt.end), nil
		}
		return nil, err
	}
	// The client memoizes series by symbol: work on a copy so that the
	// extension never leaks into the bare (real-data-only) variant of the
	// same asset. ExtendBack allocates a fresh Points slice.
	cp := *s
	s = &cp
	if simOK && marketdata.ExtendBack(s, sim) {
		s.ProxySymbol = "simdata"
		log.Printf("%s: historique étendu via simdata à partir du %s",
			canonical, s.First().Date.Format("2006-01-02"))
	}
	// Only bother with a proxy when at least six months are missing.
	if s.SimulatedBefore.IsZero() && s.First().Date.After(from.AddDate(0, 6, 0)) {
		proxySym, ok := marketdata.ProxySymbol(canonical)
		if !ok {
			proxySym, ok = marketdata.ProxySymbol(s.Symbol)
		}
		if ok {
			ps, perr := c.History(proxySym, from)
			if perr != nil {
				log.Printf("avertissement: proxy %s pour %s indisponible: %v", proxySym, s.Symbol, perr)
			} else if marketdata.ExtendBack(s, ps) {
				log.Printf("%s: historique étendu via %s à partir du %s",
					s.Symbol, proxySym, s.First().Date.Format("2006-01-02"))
			}
		}
	}
	return trim(s, time.Time{}, opt.end), nil
}

func buildPortfolio(spec *portfolio.Spec, seriesByID map[string]*marketdata.Series, feesFor func(string) (float64, bool)) *portfolio.Portfolio {
	p := &portfolio.Portfolio{Name: spec.Name, Warnings: spec.Warnings}
	if spec.EnvelopeFees > 0 {
		p.EnvelopeFees = spec.EnvelopeFees
	}
	currencies := map[string]bool{}
	for _, h := range spec.Holdings {
		s := seriesByID[h.ID]
		fees := h.Fees // la colonne du fichier prime
		if fees < 0 && feesFor != nil {
			if ter, ok := feesFor(h.ID); ok {
				fees = ter
			}
		}
		p.Assets = append(p.Assets, portfolio.Asset{
			ID:     h.ID,
			Symbol: s.Symbol,
			Name:   s.Name,
			Weight: h.Weight,
			Fees:   fees,
			Series: s,
		})
		if s.Currency != "" {
			currencies[s.Currency] = true
		}
	}
	if len(currencies) > 1 {
		list := make([]string, 0, len(currencies))
		for c := range currencies {
			list = append(list, c)
		}
		sort.Strings(list)
		p.Warnings = append(p.Warnings, fmt.Sprintf(
			"devises mélangées (%s) — aucune conversion de change n'est appliquée", strings.Join(list, ", ")))
	}
	return p
}

func buildPage(results []*result, opt *options, bench *marketdata.Series, commonStart, commonEnd time.Time) *report.Page {
	names := make([]string, len(results))
	for i, r := range results {
		names[i] = r.p.Name
	}
	page := &report.Page{
		Title:          "Portefeuilles : " + strings.Join(names, ", "),
		GeneratedAt:    time.Now().Format("02/01/2006 à 15:04"),
		RebalanceDays:  opt.rebalance,
		CommonStart:    commonStart.Format("2006-01-02"),
		CommonEnd:      commonEnd.Format("2006-01-02"),
		PortfolioNames: names,
	}

	anySimulated := false
	for _, r := range results {
		first := r.sim.Dates[0].Format("2006-01-02")
		last := r.sim.Dates[len(r.sim.Dates)-1].Format("2006-01-02")
		svg := chart.Line(chart.Options{
			Title: fmt.Sprintf("%s — base 100 du %s au %s", r.p.Name, first, last),
		}, []chart.Series{{Name: r.p.Name, Dates: r.sim.Dates, Values: r.sim.Values, Color: r.color}})

		subtitle := fmt.Sprintf("%s → %s", first, last)
		if r.rebalanceDays != opt.rebalance {
			if r.rebalanceDays == 0 {
				subtitle += " — jamais rebalancé (#meta)"
			} else {
				subtitle += fmt.Sprintf(" — rebalancement %d j (#meta)", r.rebalanceDays)
			}
		}
		if r.p.EnvelopeFees > 0 {
			subtitle += fmt.Sprintf(" — frais d'enveloppe %.2f %%/an déduits", r.p.EnvelopeFees)
		}
		section := report.PortfolioSection{
			Name:     r.p.Name,
			Subtitle: subtitle,
			ChartSVG: template.HTML(svg),
			Warnings: r.p.Warnings,
		}
		for _, a := range r.p.Assets {
			var notes []string
			if !a.Series.SimulatedBefore.IsZero() {
				anySimulated = true
				notes = append(notes, fmt.Sprintf("simulé avant le %s via %s",
					a.Series.SimulatedBefore.Format("2006-01-02"), a.Series.ProxySymbol))
			}
			switch a.Series.Source {
			case "ft", "morningstar":
				src := "Financial Times"
				if a.Series.Source == "morningstar" {
					src = "Morningstar"
				}
				note := "source : " + src + " (VL)"
				if marketdata.LooksDistributing(a.Series.Name) {
					note += " — part distribuante : dividendes non réinvestis dans cette série"
				}
				notes = append(notes, note)
			case "stooq":
				notes = append(notes, "source : Stooq (non ajusté des dividendes)")
			}
			feesText := "—"
			if a.Fees >= 0 {
				feesText = fmt.Sprintf("%.2f %%", a.Fees)
			}
			base, _ := marketdata.SplitSim(a.ID)
			ucitsText := "?"
			if u, known := marketdata.GuessUCITS(base, a.Name); known {
				ucitsText = map[bool]string{true: "oui", false: "non"}[u]
			}
			section.Assets = append(section.Assets, report.AssetRow{
				Weight:   fmt.Sprintf("%.4g %%", a.Weight*100),
				ID:       a.ID,
				Symbol:   a.Symbol,
				Name:     a.Name,
				UCITS:    ucitsText,
				Fees:     feesText,
				Currency: a.Series.Currency,
				History: fmt.Sprintf("%s → %s",
					a.Series.First().Date.Format("2006-01-02"),
					a.Series.Last().Date.Format("2006-01-02")),
				Note: strings.Join(notes, " ; "),
			})
		}
		page.Portfolios = append(page.Portfolios, section)
	}

	if len(results) > 1 {
		cmp := make([]chart.Series, len(results))
		for i, r := range results {
			cmp[i] = chart.Series{Name: r.p.Name, Dates: r.winDates, Values: r.winValues, Color: r.color}
		}
		svg := chart.Line(chart.Options{
			Title:  "Comparaison des portefeuilles — base 100 au " + page.CommonStart,
			Height: 480,
		}, cmp)
		page.CompareSVG = template.HTML(svg)
	}

	page.StatRows = buildStatRows(results, opt.benchmark)

	page.Footnotes = []string{
		"Sources : Yahoo Finance (clôtures ajustées, dividendes et splits réinvestis), Financial Times et Morningstar (valeurs liquidatives des fonds) — cache local dans « " + opt.dataDir + " ».",
		fmt.Sprintf("Simulation : base 100, rebalancement vers les poids cibles tous les %d jours calendaires par défaut (surchargeable par portefeuille via « #meta rebalance:N »), sans frais ni fiscalité.", opt.rebalance),
		"Statistiques calculées sur la période commune à tous les portefeuilles ; volatilité et ratios annualisés sur 252 jours de bourse, taux sans risque nul pour Sharpe et Sortino (convention Curvo ; PortfolioVisualizer/LazyPortfolio utilisent les T-bills et des données mensuelles — leurs volatilités et drawdowns sortent donc plus faibles).",
		"Frais : TER publiés (sources FT/justETF), déjà inclus dans les cours et VL — colonne informative ; seuls les frais additionnels de portefeuille « #meta extra-fees:X » (enveloppe, mandat…) sont déduits de la performance simulée.",
		"Max Drawdown, Ulcer et TTR sur clôtures quotidiennes — plus sévères que les références en pas mensuel (ex. COVID 2020 : −33.7 % en quotidien, −20 % en clôtures mensuelles).",
		"TTR : durée de la plus longue période passée sous un précédent sommet (du pic au retour au pic).",
	}
	if anySimulated {
		page.Footnotes = append(page.Footnotes,
			"Historiques étendus avant la création de certains fonds : par proxy (indices ou fonds plus anciens — les indices de prix n'incluent pas les dividendes) ou par données simulées permanentes (fichiers simdata/<id>.csv générés par cmd/simgen, méthodologie et qualité de réplication en tête de chaque fichier).")
	}
	if bench != nil {
		page.Footnotes = append(page.Footnotes,
			"Beta : régression des rendements quotidiens contre "+bench.Symbol+" sur la période commune.")
	}
	return page
}

func buildStatRows(results []*result, benchmark string) []report.StatRow {
	// cell computes a row's value (for the best-of-row comparison) and its
	// display text for one portfolio.
	type def struct {
		label  string
		hint   string
		cell   func(r *result) (float64, string)
		better int // +1 higher wins, -1 lower wins, 0 no highlight
	}
	pct := func(get func(metrics.Stats) float64) func(*result) (float64, string) {
		return func(r *result) (float64, string) { v := get(r.stats); return v, fmtPct(v) }
	}
	num := func(get func(metrics.Stats) float64) func(*result) (float64, string) {
		return func(r *result) (float64, string) { v := get(r.stats); return v, fmtNum(v) }
	}
	defs := []def{
		{"CAGR (rendement annualisé)", "taux de croissance annuel moyen",
			pct(func(s metrics.Stats) float64 { return s.CAGR }), +1},
		{"Volatilité (annualisée)", "écart-type des rendements quotidiens, annualisé",
			pct(func(s metrics.Stats) float64 { return s.Volatility }), -1},
		{"Sharpe", "rendement annualisé / volatilité (taux sans risque 0)",
			num(func(s metrics.Stats) float64 { return s.Sharpe }), +1},
		{"Sortino", "rendement annualisé / volatilité des seuls jours de baisse",
			num(func(s metrics.Stats) float64 { return s.Sortino }), +1},
		{"Ulcer Index", "profondeur et durée moyennes des drawdowns (plus bas = mieux)",
			num(func(s metrics.Stats) float64 { return s.Ulcer }), -1},
		{"Max Drawdown", "pire baisse depuis un sommet",
			pct(func(s metrics.Stats) float64 { return s.MaxDrawdown }), +1},
		{"TTR (récupération la plus longue)", "durée de la plus longue période sous un sommet",
			func(r *result) (float64, string) { return float64(r.stats.TTRDays), fmtTTR(r.stats) }, -1},
		{"Frais courants pondérés", "Σ poids × TER publié, plus les extra-fees appliqués à tout le portefeuille (seuls ces derniers sont déduits de la simulation)",
			func(r *result) (float64, string) {
				w, incomplete := weightedFees(r.p)
				text := fmtPct(w / 100)
				if incomplete && !math.IsNaN(w) {
					text += " (incomplet)"
				}
				return w, text
			}, -1},
		{"Beta (vs " + benchmark + ")", "sensibilité aux variations du benchmark",
			func(r *result) (float64, string) {
				if !r.stats.HasBeta {
					return math.NaN(), "—"
				}
				return r.stats.Beta, fmtNum(r.stats.Beta)
			}, 0},
	}

	rows := make([]report.StatRow, 0, len(defs))
	for _, d := range defs {
		row := report.StatRow{Label: d.label, Hint: d.hint}
		vals := make([]float64, len(results))
		for i, r := range results {
			v, text := d.cell(r)
			vals[i] = v
			row.Cells = append(row.Cells, report.StatCell{Text: text})
		}
		markBest(row.Cells, vals, d.better)
		rows = append(rows, row)
	}
	return rows
}

// weightedFees sums weight×TER over the holdings whose TER is known, plus
// the envelope fee; incomplete reports whether some TER was unknown.
func weightedFees(p *portfolio.Portfolio) (fees float64, incomplete bool) {
	known := false
	for _, a := range p.Assets {
		if a.Fees >= 0 {
			fees += a.Weight * a.Fees
			known = true
		} else {
			incomplete = true
		}
	}
	if p.EnvelopeFees > 0 {
		fees += p.EnvelopeFees
		known = true
	}
	if !known {
		return math.NaN(), incomplete
	}
	return fees, incomplete
}

// markBest highlights the cell(s) holding the best value of a row.
func markBest(cells []report.StatCell, vals []float64, better int) {
	if better == 0 || len(vals) < 2 {
		return
	}
	best := math.NaN()
	for _, v := range vals {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			continue
		}
		if math.IsNaN(best) || (better > 0 && v > best) || (better < 0 && v < best) {
			best = v
		}
	}
	if math.IsNaN(best) {
		return
	}
	for i, v := range vals {
		if !math.IsNaN(v) && !math.IsInf(v, 0) && math.Abs(v-best) <= 1e-12*math.Max(1, math.Abs(best)) {
			cells[i].Best = true
		}
	}
}

func fmtPct(x float64) string {
	if math.IsNaN(x) || math.IsInf(x, 0) {
		return "—"
	}
	return fmt.Sprintf("%.2f %%", x*100)
}

func fmtNum(x float64) string {
	if math.IsNaN(x) || math.IsInf(x, 0) {
		return "—"
	}
	return fmt.Sprintf("%.2f", x)
}

func fmtTTR(s metrics.Stats) string {
	if s.TTRDays <= 0 {
		return "—"
	}
	out := fmt.Sprintf("%d j", s.TTRDays)
	if s.TTRDays >= 365 {
		out = fmt.Sprintf("%.1f ans (%d j)", float64(s.TTRDays)/365.25, s.TTRDays)
	}
	if s.TTROngoing {
		out += " (en cours)"
	}
	return out
}

// window returns the bounds [i, j) of dates within [from, to].
func window(dates []time.Time, from, to time.Time) (int, int) {
	i := sort.Search(len(dates), func(k int) bool { return !dates[k].Before(from) })
	j := sort.Search(len(dates), func(k int) bool { return dates[k].After(to) })
	return i, j
}

// trim returns s restricted to [from, to] — zero bounds are open. The input
// may be shared through memoization, so trimming always works on a copy; it
// is returned as-is when nothing falls outside the bounds.
func trim(s *marketdata.Series, from, to time.Time) *marketdata.Series {
	if len(s.Points) == 0 ||
		((from.IsZero() || !s.First().Date.Before(from)) &&
			(to.IsZero() || !s.Last().Date.After(to))) {
		return s
	}
	out := *s
	out.Points = nil
	for _, p := range s.Points {
		if (!from.IsZero() && p.Date.Before(from)) || (!to.IsZero() && p.Date.After(to)) {
			continue
		}
		out.Points = append(out.Points, p)
	}
	return &out
}

// rebase rescales a value slice so that it starts at 100.
func rebase(values []float64) []float64 {
	out := make([]float64, len(values))
	for i, v := range values {
		out[i] = v / values[0] * 100
	}
	return out
}

func seriesSlices(s *marketdata.Series) ([]time.Time, []float64) {
	dates := make([]time.Time, len(s.Points))
	values := make([]float64, len(s.Points))
	for i, p := range s.Points {
		dates[i] = p.Date
		values[i] = p.Close
	}
	return dates, values
}

func openBrowser(path string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "linux":
		cmd = exec.Command("xdg-open", path)
	default:
		log.Printf("ouvrez %s manuellement", path)
		return
	}
	if err := cmd.Start(); err != nil {
		log.Printf("impossible d'ouvrir le navigateur: %v", err)
	}
}
