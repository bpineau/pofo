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
		},
	},
	{
		Title: "Withdrawal strategies",
		Blurb: "Which rule to draw by: the impossible triangle, every strategy in detail, and how to pick yours.",
		Articles: []Article{
			{Slug: "vpw", Title: "VPW, the Bogleheads' variable percentage withdrawal", Blurb: "A loan annuity reversed: the exact mechanics, the pension bridge, the loss tolerance test everyone skips, and where it stands against ABW.", Source: "vpw"},
		},
	},
	{
		Title: "The retirement portfolio",
		Blurb: "What you live on for forty years: the allocation, the time dimension, and the blocks that hold in every regime.",
	},
	{
		Title: "Alternative assets",
		Blurb: "Past stocks and bonds: the blocks that truly diversify, what they cost, and how to buy them cleanly.",
	},
	{
		Title: "Buffers and protections",
		Blurb: "The shock absorbers of a plan: the cash buffer and its rules, buckets demystified, ladders, and the protections around your wealth.",
	},
	{
		Title: "Inflation",
		Blurb: "What kills retirees: its history, how you measure it, its exact effect on the withdrawal rate, and the defenses that work.",
	},
	{
		Title: "Taxes and the US framework",
		Blurb: "The retiree inside the US system: accounts and account order, tax on the withdrawal phase, health cover before Medicare, and Social Security.",
	},
	{
		Title: "The human factor",
		Blurb: "The hard part, spending and living and lasting: what the models never see, and what the veterans report.",
	},
	{
		Title: "In practice",
		Blurb: "Building the plan and flying it: assemble it, keep it alive, cross the storms, and three complete cases.",
	},
	{
		Title: "References",
		Blurb: "The glossary, the annotated library, and the machine that computes this book.",
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
