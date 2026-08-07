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
var CategoriesEN = []Category{
	{
		Title: "Getting started",
		Blurb: "The wide shots: what FIRE is, the 4% rule, and how much you actually need.",
	},
	{
		Title: "The science of withdrawal",
		Blurb: "What research really knows about the safe withdrawal rate, from the classics to recent work.",
	},
	{
		Title: "Modeling: Monte Carlo and other machines",
		Blurb: "Simulators from the inside: what they do well, what they make up, and how to read them.",
	},
	{
		Title: "Withdrawal strategies",
		Blurb: "Which rule to draw by: the impossible triangle, every strategy in detail, and how to pick yours.",
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

// taxOnlyFR is the French tax part ("Fiscalité et cadre français"). Those
// seven articles are French law end to end and are deliberately never
// translated: the English edition writes usFrameworkEN in their slot instead.
// Drift skips them when it lists what the translation owes.
var taxOnlyFR = map[string]bool{
	"enveloppes-francaises":       true,
	"flat-tax-et-imposition":      true,
	"taxe-puma":                   true,
	"retraite-legale":             true,
	"sante-et-protection-sociale": true,
	"succession-et-transmission":  true,
	"expatriation-fiscale":        true,
}

// usFrameworkEN is the English edition's own part, written from scratch
// against US primary sources rather than translated, so those articles carry
// no Source and no source stamp.
var usFrameworkEN = []string{
	"us-accounts-and-account-order",
	"us-taxes-in-the-withdrawal-phase",
	"us-healthcare-and-social-security",
}

// plannedEN lists every article of the English edition, written or not, in
// reading order, and is therefore the FR -> EN slug map the translation
// campaign works from: each line names the French original it comes from.
// Wiki-links in an English article may point at any slug listed here, written
// or not (they degrade to plain text until the target exists), exactly as
// planned works for the French edition.
//
// Proper nouns keep their slug (vpw, guyton-klinger, anarkulova-cederburg).
// The seven French tax articles are deliberately absent: the three us-*
// articles below are the English edition's own, written from scratch against
// US primary sources, and carry no Source.
var plannedEN = []string{
	// I. Getting started
	"what-is-fire",               // fire-cest-quoi
	"the-4-percent-rule",         // la-regle-des-4-pourcents
	"how-much-you-need",          // combien-il-vous-faut
	"the-three-phases",           // les-trois-phases
	"using-the-fire-simulator",   // utiliser-la-page-fire
	"ten-plan-wrecking-mistakes", // erreurs-classiques-fire
	// II. The science of withdrawal
	"the-trinity-study",               // etude-trinity
	"sequence-of-returns",             // sequence-des-rendements
	"failure-probability",             // ruine-et-probabilites
	"arithmetic-vs-geometric-returns", // rendements-arithmetiques-geometriques
	"anarkulova-cederburg",            // anarkulova-cederburg
	"valuations-and-cape",             // valorisations-et-cape
	"expected-returns",                // rendements-attendus
	"horizon-and-life-expectancy",     // horizon-et-esperance-de-vie
	"the-ern-series",                  // serie-ern
	"the-math-of-4-percent",           // les-maths-du-4-pourcent
	"deciding-under-uncertainty",      // decider-sous-incertitude
	// III. Modeling
	"monte-carlo-strengths-and-limits", // monte-carlo-forces-faiblesses
	"historical-vs-parametric",         // historique-vs-parametrique
	"fat-tails",                        // queues-epaisses
	"reading-a-fan-chart",              // lire-un-fan-chart
	"simulator-traps",                  // pieges-des-simulateurs
	"making-monte-carlo-relevant",      // rendre-monte-carlo-pertinent
	"market-regimes",                   // regimes-de-marche
	// IV. Withdrawal strategies
	"withdrawal-strategies-overview",      // panorama-strategies-retrait
	"fixed-inflation-adjusted-withdrawal", // retrait-fixe-bengen
	"fixed-percentage",                    // pourcentage-fixe
	"guyton-klinger",                      // guyton-klinger
	"vpw",                                 // vpw
	"cape-based-rules",                    // regles-cape
	"morningstar-guardrails",              // guardrails-morningstar
	"amortization-based-withdrawal",       // amortissement-abw
	"floor-and-ceiling",                   // plancher-plafond
	"annuities-and-safety-first",          // rentes-et-annuites (adapted: US SPIA/DIA market)
	"seven-ways-to-live-on-one-portfolio", // sept-facons-de-vivre
	"choosing-your-strategy",              // choisir-sa-strategie
	// V. The retirement portfolio
	"risk-premia",                   // primes-de-risque
	"why-diversification-works",     // pourquoi-la-diversification-marche
	"designing-a-portfolio",         // concevoir-un-portefeuille
	"stock-bond-allocation",         // allocation-actions-obligations
	"glidepaths",                    // glidepaths
	"all-weather-portfolios",        // portefeuilles-tous-temps
	"defensive-assets",              // actifs-defensifs
	"false-defensive-assets",        // faux-actifs-defensifs
	"gold-in-retirement",            // or-en-retrait (adapted: US buying practice)
	"bonds-in-retirement",           // obligations-en-retrait
	"inflation-linked-bonds",        // obligations-indexees
	"factors-in-retirement",         // facteurs-fama-french
	"international-diversification", // diversification-internationale
	"building-it-with-us-etfs",      // etf-ucits-europeens (adapted: US-listed ETFs)
	// V bis. Alternative assets
	"managed-futures",  // managed-futures
	"long-volatility",  // long-volatility
	"global-macro",     // global-macro
	"insurance-premia", // primes-d-assurance
	"return-stacking",  // return-stacking
	// VI. Buffers and protections
	"cash-buffer",               // cash-buffer
	"enhanced-cash",             // cash-ameliore
	"the-bucket-strategy",       // strategie-buckets
	"bond-ladders",              // echelle-obligataire
	"refilling-the-buffer",      // recharger-ou-pas
	"real-estate-in-retirement", // immobilier-en-retrait
	"leverage-and-margin",       // levier-et-marges
	// VII. Inflation
	"inflation-history",              // inflation-histoire
	"tracking-inflation",             // suivre-inflation
	"inflation-and-withdrawal-rates", // inflation-et-taux-de-retrait
	"inflation-protection",           // se-proteger-de-inflation
	"hyperinflation-and-extremes",    // hyperinflation-et-extremes
	// VIII. Taxes and the US framework (English-only, no French source)
	"us-accounts-and-account-order",
	"us-taxes-in-the-withdrawal-phase",
	"us-healthcare-and-social-security",
	// IX. The human factor
	"the-psychology-of-spending", // psychologie-du-retrait
	"voices-from-real-retirees",  // temoignages-fire
	"meaning-and-identity",       // sens-et-identite
	"couples-and-family",         // couple-et-famille
	"flexibility-in-practice",    // flexibilite-realite
	"one-more-year",              // une-annee-de-plus
	"going-back-to-work",         // retour-au-travail
	// X. In practice
	"building-your-plan",         // construire-son-plan
	"the-annual-review",          // revue-annuelle
	"when-to-worry",              // quand-s-inquieter
	"bear-markets-in-retirement", // marche-baissier-en-retraite
	"pensions-and-other-income",  // revenus-complementaires
	"spending-in-retirement",     // depenses-en-retraite
	"three-worked-plans",         // cas-types
	// XI. References
	"glossary",       // lexique
	"the-library",    // bibliotheque
	"under-the-hood", // la-machine-pofo
}
