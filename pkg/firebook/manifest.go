package firebook

import "embed"

//go:embed assets/book
var assets embed.FS

// Article is one page of the book.
type Article struct {
	Slug  string // file name (assets/book/fr/<slug>.md) and URL path
	Title string // display title (the page h1; the in-file # line is dropped)
	Blurb string // one-line teaser shown on the index page
}

// Category groups articles on the index page.
type Category struct {
	Title    string
	Blurb    string
	Articles []Article
}

// Categories is the book's table of contents, in reading order. It lists
// only WRITTEN articles; the full plan (written and future) is
// docs/fire-book-design.md, mirrored by planned below. The index page and
// all navigation are generated from this manifest.
var Categories = []Category{
	{
		Title: "Démarrer",
		Blurb: "Les vues d'ensemble : ce qu'est le FIRE, la règle des 4 %, et combien il vous faut vraiment.",
		Articles: []Article{
			{"fire-cest-quoi", "Le FIRE, c'est quoi ?", "Histoire, variantes (Lean, Fat, Barista, Coast), ordres de grandeur : la carte d'entrée du sujet."},
			{"la-regle-des-4-pourcents", "La règle des 4 % en dix minutes", "D'où elle vient, ce qu'elle dit exactement, et pourquoi ce n'est qu'un point de départ."},
			{"combien-il-vous-faut", "Combien il vous faut", "Du budget annuel au capital cible : 25x, 28x, 33x, et tout ce qui fait bouger le multiple."},
			{"les-trois-phases", "Accumulation, transition, retrait : les trois vies d'un plan FIRE", "Ce qui change à chaque phase, ce qu'il faut y optimiser, et les gestes de passage."},
			{"utiliser-la-page-fire", "Utiliser la page FIRE de pofo", "Le mode d'emploi complet : les sections dans l'ordre de lecture, chaque contrôle du tiroir, et les mésusages classiques."},
			{"erreurs-classiques-fire", "Les dix erreurs qui ruinent un plan FIRE", "Les pièges les plus fréquents, du taux irréaliste à l'oubli de la fiscalité, et comment les éviter."},
		},
	},
	{
		Title: "La science du retrait",
		Blurb: "Ce que la recherche sait vraiment du taux de retrait sûr, des classiques aux travaux récents.",
		Articles: []Article{
			{"etude-trinity", "Bengen, l'étude Trinity et la naissance du taux de retrait sûr", "Les études fondatrices de 1994-1998 : ce qu'elles ont montré, et ce qu'on leur fait dire à tort."},
			{"sequence-des-rendements", "Le risque de séquence : le vrai ennemi du retraité", "Pourquoi deux retraités avec le même rendement moyen finissent l'un riche, l'autre ruiné."},
			{"ruine-et-probabilites", "La probabilité de ruine : la lire, la choisir, ne pas la subir", "Ce que mesure vraiment le chiffre des simulateurs, comment choisir son seuil, et pourquoi les décimales mentent."},
			{"rendements-arithmetiques-geometriques", "Moyenne arithmétique, moyenne géométrique et volatility drag", "Pourquoi les rendements des plaquettes ne sont pas vivables, et la cascade qui mène au taux de retrait."},
			{"anarkulova-cederburg", "Au-delà des États-Unis : Anarkulova, Cederburg et l'échantillon mondial", "Le taux de retrait recalculé sur le siècle entier des pays développés, ses chiffres qui dérangent, et ses critiques."},
			{"valorisations-et-cape", "Les valorisations (CAPE) et ce qu'elles disent du taux de retrait", "Le meilleur prédicteur connu du sort d'un millésime : définition, chiffres, critiques, et les quatre usages légitimes dans un plan."},
			{"rendements-attendus", "Les rendements attendus prospectifs", "Construire un μ défendable : building blocks, les fourchettes de Vanguard à GMO, la précision réelle, et comment ne pas empiler les prudences."},
			{"horizon-et-esperance-de-vie", "Horizon, espérance de vie et retraites de 50 ans", "Le bon quantile de survie, la courbe taux-horizon qui s'aplatit, la ruine pondérée par la mortalité, et la phase à découvert."},
			{"serie-ern", "La série Safe Withdrawal Rate d'ERN : guide de lecture", "La référence moderne du sujet : ses résultats majeurs partie par partie, et les filtres pour la lire depuis la France."},
		},
	},
	{
		Title: "Modéliser : Monte-Carlo et autres machines",
		Blurb: "Comprendre les simulateurs de l'intérieur : ce qu'ils savent faire, ce qu'ils inventent, et comment les lire.",
		Articles: []Article{
			{"monte-carlo-forces-faiblesses", "Monte-Carlo : forces, faiblesses, bon usage", "La machine derrière toutes les probabilités de ruine : comment elle marche, ses quatre faiblesses structurelles, et les huit règles du bon usage."},
			{"historique-vs-parametrique", "Fenêtres historiques, bootstrap, paramétrique : trois familles de modèles", "D'où viennent les futurs simulés, quelle question chaque famille sait vraiment traiter, et que faire de leurs désaccords."},
			{"queues-epaisses", "Queues épaisses, crises et Student-t", "Pourquoi les marchés produisent dix fois trop de catastrophes pour la courbe en cloche, et ce que le curseur df décide vraiment."},
			{"lire-un-fan-chart", "Lire un fan chart et des percentiles sans se tromper", "L'anatomie du cône de richesse, sa géométrie qui parle, les cinq erreurs de lecture classiques et les autres éventails de la page."},
			{"pieges-des-simulateurs", "Les pièges des simulateurs", "Pourquoi cinq outils rendent cinq verdicts pour le même plan : les dix pièges hiérarchisés, et la grille d'audit en dix questions."},
			{"rendre-monte-carlo-pertinent", "Rendre un Monte-Carlo pertinent (blending, régimes, stress)", "Les six corrections qui transforment le générateur de nombres en instrument : du blending vers le prior mondial au plan réel simulé."},
			{"regimes-de-marche", "Les régimes de marché : croissance × inflation, ours collants", "Les saisons des marchés, la grille à quatre quadrants, le cauchemar stagflationniste du rentier, et l'audit de portefeuille par régime."},
		},
	},
	{
		Title: "Les stratégies de retrait",
		Blurb: "Selon quelle règle prélever : le triangle impossible, chaque stratégie en détail, et comment choisir la vôtre.",
		Articles: []Article{
			{"panorama-strategies-retrait", "Panorama des stratégies de retrait : la carte avant le territoire", "Le triangle impossible, les deux extrêmes qui bornent tout, les cinq familles, et les six critères pour les noter honnêtement."},
			{"retrait-fixe-bengen", "Le retrait fixe indexé (Bengen) : le classique de référence", "La règle fondatrice en stratégie opérationnelle : la mécanique fine, la falaise silencieuse, et les trois amendements quasi gratuits."},
			{"pourcentage-fixe", "Le pourcentage fixe du portefeuille : increvable mais inconfortable", "La ruine impossible et la ruine de train de vie, le lissage des dotations (règle de Yale), et le choix du pourcentage."},
			{"guyton-klinger", "Guyton-Klinger : les guardrails historiques, grandeur et limites", "Les quatre règles exactes de 2006, la cascade de coupes des mauvais millésimes, et les correctifs modernes, plancher en tête."},
			{"vpw", "VPW, le retrait à pourcentage variable des Bogleheads", "L'annuité inversée en table gravée : la mécanique exacte, le pont de pension, le test de tolérance à la perte, et la frontière avec l'ABW."},
			{"regles-cape", "Les règles CAPE : ajuster le retrait aux valorisations (ERN)", "Taux = a + b/CAPE : la double contra-cyclicité qui auto-lisse le revenu, les paramètres d'ERN, et la forme aboutie ABW + ancre CAPE."},
			{"guardrails-morningstar", "Les guardrails modernes (Morningstar) : l'état de l'art", "Le juge honnête de Morningstar, le capteur par risque de Kitces-Tharp, et la version exécutable avec pofo en instrument."},
			{"amortissement-abw", "Le retrait par amortissement (ABW/TPAW) : l'approche actuarielle", "Le crédit inversé re-coté chaque année : richesse totale, quatre paramètres personnels, et le match final contre les guardrails."},
		},
	},
}

// Titles maps every written article's slug to its display title; it is the
// titles argument ToHTML expects for resolving [[slug]] links.
func Titles() map[string]string {
	m := make(map[string]string)
	for _, cat := range Categories {
		for _, a := range cat.Articles {
			m[a.Slug] = a.Title
		}
	}
	return m
}

// find returns the article and its category, or ok=false.
func find(slug string) (Article, Category, bool) {
	for _, cat := range Categories {
		for _, a := range cat.Articles {
			if a.Slug == slug {
				return a, cat, true
			}
		}
	}
	return Article{}, Category{}, false
}

// planned lists every article of the book's full plan, written or not
// (docs/fire-book-design.md is the human-readable version). Wiki-links are
// validated against this set, so an article may link forward to a page that
// does not exist yet (it renders as plain text until then), while a typo in
// a slug still fails the guard test.
var planned = []string{
	// I. Démarrer
	"fire-cest-quoi", "la-regle-des-4-pourcents", "combien-il-vous-faut",
	"les-trois-phases", "utiliser-la-page-fire", "erreurs-classiques-fire",
	// II. La science du retrait
	"etude-trinity", "sequence-des-rendements", "ruine-et-probabilites",
	"rendements-arithmetiques-geometriques", "anarkulova-cederburg",
	"valorisations-et-cape", "rendements-attendus",
	"horizon-et-esperance-de-vie", "serie-ern",
	// III. Modéliser
	"monte-carlo-forces-faiblesses", "historique-vs-parametrique",
	"queues-epaisses", "lire-un-fan-chart", "pieges-des-simulateurs",
	"rendre-monte-carlo-pertinent", "regimes-de-marche",
	// IV. Les stratégies de retrait
	"panorama-strategies-retrait", "retrait-fixe-bengen", "pourcentage-fixe",
	"guyton-klinger", "vpw", "regles-cape", "guardrails-morningstar",
	"amortissement-abw", "plancher-plafond", "rentes-et-annuites",
	"choisir-sa-strategie",
	// V. Le portefeuille de retrait
	"allocation-actions-obligations", "glidepaths", "portefeuilles-tous-temps",
	"actifs-defensifs", "or-en-retrait", "obligations-en-retrait",
	"obligations-indexees", "managed-futures", "facteurs-fama-french",
	"diversification-internationale", "etf-ucits-europeens",
	// VI. Buffers et protections
	"cash-buffer", "strategie-buckets", "echelle-obligataire",
	"recharger-ou-pas", "immobilier-en-retrait", "levier-et-marges",
	// VII. L'inflation
	"inflation-histoire", "suivre-inflation", "inflation-et-taux-de-retrait",
	"se-proteger-de-inflation", "hyperinflation-et-extremes",
	// VIII. Fiscalité et cadre français
	"enveloppes-francaises", "flat-tax-et-imposition", "taxe-puma",
	"retraite-legale", "sante-et-protection-sociale",
	"succession-et-transmission", "expatriation-fiscale",
	// IX. Le facteur humain
	"psychologie-du-retrait", "temoignages-fire", "sens-et-identite",
	"couple-et-famille", "flexibilite-realite", "une-annee-de-plus",
	"retour-au-travail",
	// X. En pratique
	"construire-son-plan", "revue-annuelle", "quand-s-inquieter",
	"marche-baissier-en-retraite", "revenus-complementaires",
	"depenses-en-retraite", "cas-types",
	// XI. Références
	"lexique", "bibliotheque", "la-machine-pofo",
}
