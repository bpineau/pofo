package firebook

// CategoriesEN is the English edition's table of contents. It mirrors the
// French Categories part for part and in the same order, with two deliberate
// differences: the French tax part ("Fiscalité et cadre français") has no
// English counterpart, and a shorter, simpler US-framework part sits in its
// slot.
//
// Articles arrive here as they are translated; an entry means the file exists
// under assets/book/en. Every translated article carries Source, the slug of
// the French original it was made from, and an in-file source stamp; Drift
// reads both to report what the translation still owes. The French edition
// stays the source of truth for everything but the US part.
//
// Which French articles have no counterpart is not recorded here: each of them
// says so in its own file, with the fr-only edition marker Drift reads (see
// drift.go).
var CategoriesEN = []Category{
	{
		Title: "Getting started",
		Blurb: "The wide shots: what FIRE is, the 4% rule, and how much you actually need.",
		Articles: []Article{
			{Slug: "what-is-fire", Title: "What FIRE actually is", Blurb: "History, the variants (Lean, Fat, Barista, Coast) and the ballpark numbers: the map you enter the subject with.", Source: "fire-cest-quoi"},
			{Slug: "the-4-percent-rule", Title: "The 4% rule in ten minutes", Blurb: "Where it came from, what it says exactly, and why it is only a starting point.", Source: "la-regle-des-4-pourcents"},
			{Slug: "how-much-you-need", Title: "How much you need", Blurb: "From the annual budget to the capital target: 25x, 28x, 33x, and everything that moves the multiple.", Source: "combien-il-vous-faut"},
			{Slug: "the-three-phases", Title: "Accumulation, transition, withdrawal: the three lives of a FIRE plan", Blurb: "What changes in each phase, what to optimize there, and the moves that carry you from one to the next.", Source: "les-trois-phases"},
			{Slug: "using-the-fire-simulator", Title: "Using the FIRE simulator", Blurb: "The full manual: the sections in reading order, every control in the parameters panel, and the classic ways to misuse it.", Source: "utiliser-la-page-fire"},
			{Slug: "ten-plan-wrecking-mistakes", Title: "The ten mistakes that wreck a FIRE plan", Blurb: "The most common traps, from an unrealistic rate to forgetting taxes, and how to steer clear of them.", Source: "erreurs-classiques-fire"},
		},
	},
	{
		Title: "The science of withdrawal",
		Blurb: "What research really knows about the safe withdrawal rate, from the classics to recent work.",
		Articles: []Article{
			{Slug: "the-trinity-study", Title: "Bengen, the Trinity study, and the birth of the safe withdrawal rate", Blurb: "The founding studies of 1994 to 1998: what they showed, and what they are wrongly made to say.", Source: "etude-trinity"},
			{Slug: "sequence-of-returns", Title: "Sequence of returns risk: the retiree's real enemy", Blurb: "Why the same returns in a different order ruin one retiree and enrich another, where the danger concentrates, and the full map of defenses.", Source: "sequence-des-rendements"},
			{Slug: "failure-probability", Title: "Failure probability: reading it, choosing it, and not letting it run you", Blurb: "What the number the simulators print really measures, how to pick your threshold, and why the decimals lie.", Source: "ruine-et-probabilites"},
			{Slug: "arithmetic-vs-geometric-returns", Title: "Arithmetic mean, geometric mean, and volatility drag", Blurb: "Why the returns in the brochures are not returns you can live on, and the cascade that leads to the withdrawal rate.", Source: "rendements-arithmetiques-geometriques"},
			{Slug: "anarkulova-cederburg", Title: "Beyond the United States: Anarkulova, Cederburg and the world sample", Blurb: "The withdrawal rate recomputed over the whole century of the developed world, its uncomfortable numbers, and its critics.", Source: "anarkulova-cederburg"},
			{Slug: "valuations-and-cape", Title: "Valuations, the CAPE, and what they say about your withdrawal rate", Blurb: "The best known predictor of a vintage's fate: what it is, the numbers, the criticisms, and the four legitimate uses in a plan.", Source: "valorisations-et-cape"},
			{Slug: "expected-returns", Title: "Forward-looking expected returns (Morningstar, Vanguard, the investment banks)", Blurb: "Building a μ you can defend: the building blocks, the ranges from Vanguard to GMO, how precise they really are, and how not to stack prudence on prudence.", Source: "rendements-attendus"},
			{Slug: "horizon-and-life-expectancy", Title: "Horizon, life expectancy, and 50-year retirements", Blurb: "The right survival quantile, the rate-horizon curve that flattens out, failure weighted by mortality, and the uncovered phase.", Source: "horizon-et-esperance-de-vie"},
			{Slug: "the-ern-series", Title: "ERN's Safe Withdrawal Rate series: a reader's guide", Blurb: "The modern reference on the subject: its major results part by part, the filters for reading it from outside the United States, and what an American reader can skip.", Source: "serie-ern"},
			{Slug: "the-math-of-4-percent", Title: "Why 4%? The mathematical anatomy of the rule", Blurb: "The three-tier cascade (real return, amortization bonus, sequence penalty), why it holds up so well, and what would break it.", Source: "les-maths-du-4-pourcent"},
			{Slug: "lifecycle-theory", Title: "Lifecycle theory: the academic backbone of the plan", Blurb: "Modigliani, Samuelson, Merton, Yaari: the optimal risky share, human capital, annuities as the default, and the labeled ledger of where this book departs from theory.", Source: "theorie-du-cycle-de-vie"},
			{Slug: "deciding-under-uncertainty", Title: "Deciding under uncertainty: utility, Kelly, regret, and robust choices", Blurb: "You only live one path: the certainty equivalent, tolerance against capacity, why to stay away from full Kelly, and the protocol in five rules.", Source: "decider-sous-incertitude"},
		},
	},
	{
		Title: "Modeling: Monte Carlo and other machines",
		Blurb: "Simulators from the inside: what they do well, what they make up, and how to read them.",
		Articles: []Article{
			{Slug: "monte-carlo-strengths-and-limits", Title: "Monte Carlo: strengths, weaknesses, and how to use it well", Blurb: "The machine behind every failure probability: how it works, its four structural weaknesses, and the eight rules for using it well.", Source: "monte-carlo-forces-faiblesses"},
			{Slug: "historical-vs-parametric", Title: "Historical windows, bootstrap, parametric: the three families of models", Blurb: "Where simulated futures come from, which question each family can really answer, and what to do when they disagree.", Source: "historique-vs-parametrique"},
			{Slug: "fat-tails", Title: "Fat tails, crises, and the Student-t", Blurb: "Why markets produce ten times too many disasters for the bell curve, and what the df dial really decides.", Source: "queues-epaisses"},
			{Slug: "reading-a-fan-chart", Title: "Reading a fan chart without fooling yourself", Blurb: "The anatomy of the wealth fan, the geometry that talks, the five classic reading mistakes, and the other fans on the page.", Source: "lire-un-fan-chart"},
			{Slug: "simulator-traps", Title: "Simulator traps: independence, US bias, survivorship", Blurb: "Why five tools return five verdicts on the same plan: the ten traps ranked by damage, and the ten-question audit.", Source: "pieges-des-simulateurs"},
			{Slug: "making-monte-carlo-relevant", Title: "Making a Monte Carlo relevant: blending, regimes, stress", Blurb: "The six fixes that turn a random number generator into a planning instrument, from blending toward the world prior to simulating the real plan.", Source: "rendre-monte-carlo-pertinent"},
			{Slug: "market-regimes", Title: "Market regimes (growth × inflation, sticky bears) and why they matter", Blurb: "The seasons of markets, the four-quadrant grid, the retiree's stagflation nightmare, and auditing a portfolio regime by regime.", Source: "regimes-de-marche"},
		},
	},
	{
		Title: "Withdrawal strategies",
		Blurb: "Which rule to draw by: the impossible triangle, every strategy in detail, and how to pick yours.",
		Articles: []Article{
			{Slug: "withdrawal-strategies-overview", Title: "Withdrawal strategies: the map before the territory", Blurb: "The impossible triangle, the two extremes that bound everything, the five families, and the six criteria for scoring them honestly.", Source: "panorama-strategies-retrait"},
			{Slug: "fixed-inflation-adjusted-withdrawal", Title: "The fixed inflation-adjusted withdrawal (Bengen): the benchmark rule", Blurb: "The founding rule as a working strategy: the fine mechanics, the silent cliff, and the three nearly free amendments.", Source: "retrait-fixe-bengen"},
			{Slug: "fixed-percentage", Title: "The fixed percentage of the portfolio: indestructible but uncomfortable", Blurb: "Impossible ruin and lifestyle ruin, endowment smoothing (the Yale rule), and how to pick the percentage.", Source: "pourcentage-fixe"},
			{Slug: "guyton-klinger", Title: "Guyton-Klinger: the original guardrails, their power and their limits", Blurb: "The four exact rules of 2006, the cascade of cuts in bad vintages, and the modern fixes, starting with the floor.", Source: "guyton-klinger"},
			{Slug: "vpw", Title: "VPW, the Bogleheads' variable percentage withdrawal", Blurb: "A loan annuity reversed: the exact mechanics, the pension bridge, the loss tolerance test everyone skips, and where it stands against ABW.", Source: "vpw"},
			{Slug: "cape-based-rules", Title: "CAPE-based rules: tying the withdrawal to valuations (ERN)", Blurb: "Rate = a + b/CAPE: the double countercyclicality that smooths the income by itself, ERN's parameters, and the finished form, ABW with a CAPE anchor.", Source: "regles-cape"},
			{Slug: "morningstar-guardrails", Title: "Modern guardrails (Morningstar): the state of the art", Blurb: "Morningstar's honest judge, the risk-based indicator of Kitces and Tharp, and the executable version, with a simulator as the instrument.", Source: "guardrails-morningstar"},
			{Slug: "amortization-based-withdrawal", Title: "Amortization-based withdrawal (ABW/TPAW): the actuarial approach", Blurb: "The loan run backwards and repriced every year: total wealth, four personal parameters, and the final head-to-head against guardrails.", Source: "amortissement-abw"},
			{Slug: "floor-and-ceiling", Title: "Floor and ceiling and the Vanguard rules: bounded flexibility", Blurb: "The corridor on the annual change (+5%/−2.5%): glides instead of falls, an honest failure probability again, second everywhere.", Source: "plancher-plafond"},
			{Slug: "annuities-and-safety-first", Title: "Annuities and safety first: buying a floor", Blurb: "Mortality credits, what the US market really sells (and the best annuity of all, delaying Social Security), the inflation objection, and when to annuitize.", Source: "rentes-et-annuites"},
			{Slug: "seven-ways-to-live-on-one-portfolio", Title: "Seven ways to live on one portfolio", Blurb: "Three real retirements replayed year by year: what each rule paid, when it cut, and what it left on the table.", Source: "sept-facons-de-vivre"},
			{Slug: "choosing-your-strategy", Title: "Choosing your strategy: criteria, comparison, worked case", Blurb: "The five-step procedure: the admission tests, the profile-to-rule matrix, hybrids by phase, and the written page that ends it.", Source: "choisir-sa-strategie"},
		},
	},
	{
		Title: "The retirement portfolio",
		Blurb: "What you live on for forty years: the allocation, the time dimension, and the blocks that hold in every regime.",
		Articles: []Article{
			{Slug: "risk-premia", Title: "Where returns come from: risk premia", Blurb: "Why stocks pay, who pays the premium, why it survives its own fame, and why gold returns nothing without that being a flaw.", Source: "primes-de-risque"},
			{Slug: "why-diversification-works", Title: "Why diversification works: the mechanics of the free lunch", Blurb: "The free lunch taken apart: correlations, the rebalancing premium, volatility harvesting, the doubled effect in decumulation, and the fake diversification to hunt down.", Source: "pourquoi-la-diversification-marche"},
			{Slug: "designing-a-portfolio", Title: "Designing a portfolio: the method, not the model", Blurb: "Designing by risks instead of by assets: the seven questions in order, and the allocation that falls out at the end as a result.", Source: "concevoir-un-portefeuille"},
			{Slug: "stock-bond-allocation", Title: "The stock and bond allocation in retirement", Blurb: "The 50 to 80% plateau that falls away on both sides, the three dials that place you inside it, and the 100% stocks debate put in its place.", Source: "allocation-actions-obligations"},
			{Slug: "glidepaths", Title: "Glide paths: the bond tent, rising equity, and the fragile window", Blurb: "Caution as a temporary expense: the Pfau-Kitces and ERN results, running the climb automatically through the withdrawals, and the head-to-head against the cash buffer.", Source: "glidepaths"},
			{Slug: "all-weather-portfolios", Title: "All-weather portfolios: Browne, All Weather, Golden Butterfly, Dragon", Blurb: "One winner per season: the exact compositions, the numbers that matter to a retiree, the honest criticisms, and a dose rather than a dogma.", Source: "portefeuilles-tous-temps"},
			{Slug: "defensive-assets", Title: "Defensive assets: the map and the roles", Blurb: "Defense against what: the spec, the candidates one by one, the gallery of false defensives, and how the pieces fit together.", Source: "actifs-defensifs"},
			{Slug: "false-defensive-assets", Title: "False defensives: what looks defensive and is not", Blurb: "Low vol, dividends, covered calls, REITs, high yield, private equity, crypto: why they look defensive, and why they let go in the 2008 and 2022 tests.", Source: "faux-actifs-defensifs"},
			{Slug: "gold-in-retirement", Title: "Gold in a retirement portfolio", Blurb: "A secular real return of zero, a decorrelation that survives crises: the three roles, the size of each, how to hold it in practice, and the mistakes.", Source: "or-en-retrait"},
			{Slug: "bonds-in-retirement", Title: "Bonds in retirement: types, duration, the exact job", Blurb: "Price and duration, YTM as the expectation on the label, the three services and the regimes they need, the four decisions, and where a stable value fund belongs.", Source: "obligations-en-retrait"},
			{Slug: "inflation-linked-bonds", Title: "Inflation-linked bonds: the only contract written in real terms", Blurb: "The breakeven as the decision tool, the lesson of 2022, the ladder of linkers that guarantees what the 4% rule only hopes for, and TIPS and I bonds in practice.", Source: "obligations-indexees"},
			{Slug: "factors-in-retirement", Title: "Factors (Fama-French, value, momentum) in the withdrawal phase", Blurb: "The core that survives replication, the retiree's case file (SCV, the value-inflation affinity), a decade of lagging the index, and the optional tilt, properly sized.", Source: "facteurs-fama-french"},
			{Slug: "international-diversification", Title: "International diversification (and home bias)", Blurb: "The dominant risk is your own country falling behind: the national fates in numbers, why the currency argument runs backwards for a dollar investor, and the target in one fund.", Source: "diversification-internationale"},
			{Slug: "building-it-with-us-etfs", Title: "Building it with US-listed ETFs: the retiree's shopping list", Blurb: "The 1940 Act fund and its four technical choices, the full cost chain, the table of blocks with US-listed examples, and selling cleanly across account types.", Source: "etf-ucits-europeens"},
		},
	},
	{
		Title: "Alternative assets",
		Blurb: "Past stocks and bonds: the blocks that truly diversify, what they cost, and how to buy them cleanly.",
		Articles: []Article{
			{Slug: "managed-futures", Title: "Managed futures and trend following: the diversification that works in a crisis", Blurb: "The only defensive asset with a positive expected return: a century of evidence, the crisis alpha of long regimes, how to buy it, and the winter you have to cross.", Source: "managed-futures"},
			{Slug: "long-volatility", Title: "Long volatility and tail hedging: paying for crashes", Blurb: "The convexity that explodes in fast crashes, the variance risk premium that overcharges for it, the toxic vehicles, and the rare legitimate uses.", Source: "long-volatility"},
			{Slug: "global-macro", Title: "Global macro and alternative risk premia: how the big institutions diversify", Blurb: "The macro bets that win in hostile regimes, the catalog of premia (commodity carry included), the purge of the ARP funds, and the five-question checklist.", Source: "global-macro"},
			{Slug: "insurance-premia", Title: "Insurance premia: cat bonds and merger arbitrage", Blurb: "Getting paid to carry the hurricane or the deal that breaks: causal decorrelation, the collateral currency trap, and what ten points really change (not much).", Source: "primes-d-assurance"},
			{Slug: "return-stacking", Title: "Return stacking, overlays and portable alpha: stacking the premia", Blurb: "Stacking the diversifiers on top of the core instead of funding them by selling it: the mechanics, the cost of cash, the lesson of 2022, and the sizing rules.", Source: "return-stacking"},
		},
	},
	{
		Title: "Buffers and protections",
		Blurb: "The shock absorbers of a plan: the cash buffer and its rules, buckets demystified, ladders, and the protections around your wealth.",
		Articles: []Article{
			{Slug: "cash-buffer", Title: "The cash buffer: size, cost, and what it is really for", Blurb: "The instinct is right, the arithmetic is stubborn (plus or minus 0.5 points), and the real value lies elsewhere: no panic, permission to spend, a household that can run the plan.", Source: "cash-buffer"},
			{Slug: "enhanced-cash", Title: "Enhanced cash: money market funds, T-bills and AAA CLOs", Blurb: "What to fill the short sleeve with: the rungs measured, AAA CLOs taken apart, and the three-layer assembly rule.", Source: "cash-ameliore"},
			{Slug: "the-bucket-strategy", Title: "The bucket strategy: the promise, and the case against it", Blurb: "An allocation in disguise plus flows that are either rebalancing or timing: the fair trial, and the clean version.", Source: "strategie-buckets"},
			{Slug: "bond-ladders", Title: "Bond ladders (and the ladder of linkers)", Blurb: "The matching that cancels interest-rate risk: the bridge to your pension, a floor backed by contract, and how to build one at retail with Treasuries, TIPS and defined-maturity ETFs.", Source: "echelle-obligataire"},
			{Slug: "refilling-the-buffer", Title: "Drawing on a buffer and refilling it: the rules that work", Blurb: "The drawdown trigger, refilling in calm weather, the one absolute prohibition, and the melting buffer that beats the permanent one.", Source: "recharger-ou-pas"},
			{Slug: "real-estate-in-retirement", Title: "Real estate in a FIRE plan (your home, and rentals)", Blurb: "Phantom rent and double counting, a rental counted as a discounted flow, the sell-or-keep call, and REITs and reverse mortgages in their place.", Source: "immobilier-en-retrait"},
			{Slug: "leverage-and-margin", Title: "Leverage, margin and borrowing against your portfolio (advanced)", Blurb: "Why leverage changes sign in retirement, the only three defensible uses, and the five non-negotiable rules.", Source: "levier-et-marges"},
		},
	},
	{
		Title: "Inflation",
		Blurb: "What kills retirees: its history, how you measure it, its exact effect on the withdrawal rate, and the defenses that work.",
		Articles: []Article{
			{Slug: "tracking-inflation", Title: "Tracking inflation: the indexes, and yours", Blurb: "CPI-U, CPI-W, PCE and their blind spots, the retiree's personal inflation (0.2 to 0.5 points above the index), the breakevens, and how to set the drift.", Source: "suivre-inflation"},
			{Slug: "inflation-and-withdrawal-rates", Title: "Inflation and the withdrawal rate: the exact link", Blurb: "The squeeze and simultaneous real compression: why 1966 beats 1929, the conditional numbers, and the plan's indexation inventory.", Source: "inflation-et-taux-de-retrait"},
			{Slug: "inflation-protection", Title: "Protecting against inflation: what actually works", Blurb: "The arsenal sorted by the kind of evidence: contractual, structural, episodic, behavioral, plus the blacklist and how to put it together phase by phase.", Source: "se-proteger-de-inflation"},
			{Slug: "hyperinflation-and-extremes", Title: "Hyperinflation and extreme scenarios", Blurb: "From Weimar to Cyprus: what survives, the near-free structural cover a good plan already has, and the mirror trap of the prepper.", Source: "hyperinflation-et-extremes"},
		},
	},
	{
		Title: "Taxes and the US framework",
		Blurb: "The retiree inside the US system: accounts and account order, tax on the withdrawal phase, health cover before Medicare, and Social Security.",
		Articles: []Article{
			{Slug: "us-accounts-and-account-order", Title: "US accounts, and the order you drain them", Blurb: "The three tax homes of a dollar, what each account is for, the standard filling ladder, the toolkit that opens the locked accounts before 59 1/2, and the draining order that comes out of it."},
			{Slug: "us-taxes-in-the-withdrawal-phase", Title: "US taxes in the withdrawal phase", Blurb: "Why a modest draw is often taxed at nothing: the gain inside a sale, the two rate ladders and how they stack, the friction computed in dollars, and the tools of the phase (gain and loss harvesting, specific identification, conversions, RMDs)."},
			{Slug: "us-healthcare-and-social-security", Title: "US health coverage and Social Security: the two ends of the gap", Blurb: "Covering yourself before Medicare: guaranteed issue, the premium tax credit as a function of your income and the 400% cliff that is back for 2026, the alternatives, then Medicare at 65 and its two-year lookback. And the Social Security a short career actually earns: 40 credits, 35 years with zeros in them, the bend points that hand much of it back, and how to read an SSA estimate that assumes you kept working."},
		},
	},
	{
		Title: "The human factor",
		Blurb: "The hard part, spending and living and lasting: what the models never see, and what the veterans report.",
		Articles: []Article{
			{Slug: "the-psychology-of-spending", Title: "The psychology of spending: why spending it is the hard part", Blurb: "Chronic underspending and panic in a crash: four mechanisms, the biases named, and the toolbox that equips all of it.", Source: "psychologie-du-retrait"},
			{Slug: "voices-from-real-retirees", Title: "What real FIRE retirees say: accounts and advice", Blurb: "Six findings from the corpus (the money works, the eighteen-month wall, the number one regret: not leaving sooner) and the canonical list of veterans' advice.", Source: "temoignages-fire"},
			{Slug: "meaning-and-identity", Title: "Meaning, identity, structure: life after work", Blurb: "The four workstreams a paycheck used to hide, the optimizer's traps, and the prototyping that makes all the difference.", Source: "sens-et-identite"},
			{Slug: "couples-and-family", Title: "FIRE as a couple, and as a family", Blurb: "A team sport: the mismatch in appetite, the staggered exit, governance for two (the reverse quiz), clear eyes on divorce, children and parents.", Source: "couple-et-famille"},
			{Slug: "flexibility-in-practice", Title: "Flexibility: myth and reality (what it can really absorb)", Blurb: "What it is really worth (0.3 to 0.5 points), why duration beats depth, the six forms ranked, and the floor tested rather than declared.", Source: "flexibilite-realite"},
			{Slug: "one-more-year", Title: "The one-more-year syndrome", Blurb: "The accounting asymmetry that manufactures it, both columns priced, rational OMY in three cases, and the commitment made while calm.", Source: "une-annee-de-plus"},
			{Slug: "going-back-to-work", Title: "Barista, Coast, side income: work you choose", Blurb: "The best defensive asset in the book: the forms it takes, everything that rides along with a little income, the option that melts unless you maintain it, and the toolbox.", Source: "retour-au-travail"},
		},
	},
	{
		Title: "In practice",
		Blurb: "Building the plan and flying it: assemble it, keep it alive, cross the storms, and three complete cases.",
		Articles: []Article{
			{Slug: "building-your-plan", Title: "Building your plan step by step", Blurb: "The seven steps in order, the eight-block template for the one-page plan, the validation recipe, and day one.", Source: "construire-son-plan"},
			{Slug: "the-annual-review", Title: "The annual review: the retiree's checklist", Blurb: "One session a year, seven blocks, the reverse quiz and the non-financial line: the review records and executes, it does not redesign.", Source: "revue-annuelle"},
			{Slug: "when-to-worry", Title: "When to worry, and when to let it ride", Blurb: "The current rate as the main warning light, sorting noise from signal, ruin that gives years of notice, and the five-step playbook.", Source: "quand-s-inquieter"},
			{Slug: "bear-markets-in-retirement", Title: "Riding out a bear market in retirement: the playbook", Blurb: "Three bear markets, week 1 of doing nothing on purpose, the mechanical moves at the bottom (tax-loss harvesting among them), the never-dos, and the exit.", Source: "marche-baissier-en-retraite"},
			{Slug: "pensions-and-other-income", Title: "Pensions and other income in the plan", Blurb: "Three categories, three treatments: the four things a flow does to a plan, the exact entries in the simulator, and the test for the over-backed plan.", Source: "revenus-complementaires"},
			{Slug: "spending-in-retirement", Title: "Real spending in retirement (the retirement smile, Die With Zero)", Blurb: "Blanchett's smile, dated time buckets, front-loading an early retirement on purpose, and engineering a budget in dated tiers.", Source: "depenses-en-retraite"},
		},
	},
	{
		Title: "References",
		Blurb: "The glossary, the annotated library, and the machine that computes this book.",
		Articles: []Article{
			{Slug: "glossary", Title: "The FIRE and withdrawal glossary", Blurb: "Every term in the book and in the forum jargon, defined and linked to its chapter: read straight through, an alphabetical summary of the whole book.", Source: "lexique"},
			{Slug: "the-library", Title: "The library: sites, papers, books, tools", Blurb: "Every reference annotated (why to read it, where to get it), the official sources, and three reading paths.", Source: "bibliotheque"},
			{Slug: "under-the-hood", Title: "Under the hood: how the simulator computes this book", Blurb: "The data, the central model's pipeline, the six columns, the month-by-month core, the limits it owns up to, and how to check it yourself.", Source: "la-machine-pofo"},
		},
	},
}

// usFrameworkEN is the English edition's own part, written from scratch
// against US primary sources rather than translated, so those articles carry
// no Source and no source stamp.
var usFrameworkEN = []string{
	"us-accounts-and-account-order",
	"us-taxes-in-the-withdrawal-phase",
	"us-healthcare-and-social-security",
}

// enPlan is one line of the English edition's plan: the English slug, and the
// French article it will be translated from. FR is empty for the three
// articles the English edition writes for itself.
type enPlan struct{ EN, FR string }

// plannedEN lists every article of the English edition, written or not, in
// reading order, and is therefore the FR -> EN slug map the translation
// campaign works from. The pairing is data, not a trailing comment: Drift
// reads it to tell a translation agent which English slug an untranslated
// French article is owed under. Wiki-links in an English article may point at
// any EN slug listed here, written or not (they degrade to plain text until
// the target exists), exactly as planned works for the French edition.
//
// Proper nouns keep their slug (vpw, guyton-klinger, anarkulova-cederburg).
// The French articles marked fr-only in their own file (the tax part) are
// deliberately absent: the three us-* articles below are the English edition's
// own, written from scratch against US primary sources, and carry no Source.
var plannedEN = []enPlan{
	// I. Getting started
	{EN: "what-is-fire", FR: "fire-cest-quoi"},
	{EN: "the-4-percent-rule", FR: "la-regle-des-4-pourcents"},
	{EN: "how-much-you-need", FR: "combien-il-vous-faut"},
	{EN: "the-three-phases", FR: "les-trois-phases"},
	{EN: "using-the-fire-simulator", FR: "utiliser-la-page-fire"},
	{EN: "ten-plan-wrecking-mistakes", FR: "erreurs-classiques-fire"},
	// II. The science of withdrawal
	{EN: "the-trinity-study", FR: "etude-trinity"},
	{EN: "sequence-of-returns", FR: "sequence-des-rendements"},
	{EN: "failure-probability", FR: "ruine-et-probabilites"},
	{EN: "arithmetic-vs-geometric-returns", FR: "rendements-arithmetiques-geometriques"},
	{EN: "anarkulova-cederburg", FR: "anarkulova-cederburg"},
	{EN: "valuations-and-cape", FR: "valorisations-et-cape"},
	{EN: "expected-returns", FR: "rendements-attendus"},
	{EN: "horizon-and-life-expectancy", FR: "horizon-et-esperance-de-vie"},
	{EN: "the-ern-series", FR: "serie-ern"},
	{EN: "the-math-of-4-percent", FR: "les-maths-du-4-pourcent"},
	{EN: "lifecycle-theory", FR: "theorie-du-cycle-de-vie"},
	{EN: "deciding-under-uncertainty", FR: "decider-sous-incertitude"},
	// III. Modeling
	{EN: "monte-carlo-strengths-and-limits", FR: "monte-carlo-forces-faiblesses"},
	{EN: "historical-vs-parametric", FR: "historique-vs-parametrique"},
	{EN: "fat-tails", FR: "queues-epaisses"},
	{EN: "reading-a-fan-chart", FR: "lire-un-fan-chart"},
	{EN: "simulator-traps", FR: "pieges-des-simulateurs"},
	{EN: "making-monte-carlo-relevant", FR: "rendre-monte-carlo-pertinent"},
	{EN: "market-regimes", FR: "regimes-de-marche"},
	// IV. Withdrawal strategies
	{EN: "withdrawal-strategies-overview", FR: "panorama-strategies-retrait"},
	{EN: "fixed-inflation-adjusted-withdrawal", FR: "retrait-fixe-bengen"},
	{EN: "fixed-percentage", FR: "pourcentage-fixe"},
	{EN: "guyton-klinger", FR: "guyton-klinger"},
	{EN: "vpw", FR: "vpw"},
	{EN: "cape-based-rules", FR: "regles-cape"},
	{EN: "morningstar-guardrails", FR: "guardrails-morningstar"},
	{EN: "amortization-based-withdrawal", FR: "amortissement-abw"},
	{EN: "floor-and-ceiling", FR: "plancher-plafond"},
	{EN: "annuities-and-safety-first", FR: "rentes-et-annuites"}, // adapted: US SPIA/DIA market
	{EN: "seven-ways-to-live-on-one-portfolio", FR: "sept-facons-de-vivre"},
	{EN: "choosing-your-strategy", FR: "choisir-sa-strategie"},
	// V. The retirement portfolio
	{EN: "risk-premia", FR: "primes-de-risque"},
	{EN: "why-diversification-works", FR: "pourquoi-la-diversification-marche"},
	{EN: "designing-a-portfolio", FR: "concevoir-un-portefeuille"},
	{EN: "stock-bond-allocation", FR: "allocation-actions-obligations"},
	{EN: "glidepaths", FR: "glidepaths"},
	{EN: "all-weather-portfolios", FR: "portefeuilles-tous-temps"},
	{EN: "defensive-assets", FR: "actifs-defensifs"},
	{EN: "false-defensive-assets", FR: "faux-actifs-defensifs"},
	{EN: "gold-in-retirement", FR: "or-en-retrait"}, // adapted: US buying practice
	{EN: "bonds-in-retirement", FR: "obligations-en-retrait"},
	{EN: "inflation-linked-bonds", FR: "obligations-indexees"},
	{EN: "factors-in-retirement", FR: "facteurs-fama-french"},
	{EN: "international-diversification", FR: "diversification-internationale"},
	{EN: "building-it-with-us-etfs", FR: "etf-ucits-europeens"}, // adapted: US-listed ETFs
	// V bis. Alternative assets
	{EN: "managed-futures", FR: "managed-futures"},
	{EN: "long-volatility", FR: "long-volatility"},
	{EN: "global-macro", FR: "global-macro"},
	{EN: "insurance-premia", FR: "primes-d-assurance"},
	{EN: "return-stacking", FR: "return-stacking"},
	// VI. Buffers and protections
	{EN: "cash-buffer", FR: "cash-buffer"},
	{EN: "enhanced-cash", FR: "cash-ameliore"},
	{EN: "the-bucket-strategy", FR: "strategie-buckets"},
	{EN: "bond-ladders", FR: "echelle-obligataire"},
	{EN: "refilling-the-buffer", FR: "recharger-ou-pas"},
	{EN: "real-estate-in-retirement", FR: "immobilier-en-retrait"},
	{EN: "leverage-and-margin", FR: "levier-et-marges"},
	// VII. Inflation
	{EN: "tracking-inflation", FR: "suivre-inflation"},
	{EN: "inflation-and-withdrawal-rates", FR: "inflation-et-taux-de-retrait"},
	{EN: "inflation-protection", FR: "se-proteger-de-inflation"},
	{EN: "hyperinflation-and-extremes", FR: "hyperinflation-et-extremes"},
	// VIII. Taxes and the US framework (English-only, no French source)
	{EN: "us-accounts-and-account-order"},
	{EN: "us-taxes-in-the-withdrawal-phase"},
	{EN: "us-healthcare-and-social-security"},
	// IX. The human factor
	{EN: "the-psychology-of-spending", FR: "psychologie-du-retrait"},
	{EN: "voices-from-real-retirees", FR: "temoignages-fire"},
	{EN: "meaning-and-identity", FR: "sens-et-identite"},
	{EN: "couples-and-family", FR: "couple-et-famille"},
	{EN: "flexibility-in-practice", FR: "flexibilite-realite"},
	{EN: "one-more-year", FR: "une-annee-de-plus"},
	{EN: "going-back-to-work", FR: "retour-au-travail"},
	// X. In practice
	{EN: "building-your-plan", FR: "construire-son-plan"},
	{EN: "the-annual-review", FR: "revue-annuelle"},
	{EN: "when-to-worry", FR: "quand-s-inquieter"},
	{EN: "bear-markets-in-retirement", FR: "marche-baissier-en-retraite"},
	{EN: "pensions-and-other-income", FR: "revenus-complementaires"},
	{EN: "spending-in-retirement", FR: "depenses-en-retraite"},
	// XI. References
	{EN: "glossary", FR: "lexique"},
	{EN: "the-library", FR: "bibliotheque"},
	{EN: "under-the-hood", FR: "la-machine-pofo"},
}

// plannedENSource maps a French slug to the English slug planned for it,
// derived from plannedEN so the two can never disagree.
var plannedENSource = func() map[string]string {
	m := make(map[string]string, len(plannedEN))
	for _, p := range plannedEN {
		if p.FR != "" {
			m[p.FR] = p.EN
		}
	}
	return m
}()
