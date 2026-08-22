package firebook

import "embed"

//go:embed assets/book
var assets embed.FS

// Article is one page of the book.
type Article struct {
	Slug  string // file name (<Edition.AssetDir>/<slug>.md) and URL path
	Title string // display title (the page h1; the in-file # line is dropped)
	Blurb string // one-line teaser shown on the index page

	// Source pairs a translated article with its French original: the slug of
	// the French article it was translated from. It is empty on every French
	// article, and on the articles an edition writes for itself (the US
	// framework part of the English edition). Drift reads it, together with
	// the in-file source stamp, to report what a translation owes.
	Source string
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
			{Slug: "fire-cest-quoi", Title: "Le FIRE, c'est quoi ?", Blurb: "Histoire, variantes (Lean, Fat, Barista, Coast), ordres de grandeur : la carte d'entrée du sujet."},
			{Slug: "la-regle-des-4-pourcents", Title: "La règle des 4 % en dix minutes", Blurb: "D'où elle vient, ce qu'elle dit exactement, et pourquoi ce n'est qu'un point de départ."},
			{Slug: "combien-il-vous-faut", Title: "Combien il vous faut", Blurb: "Du budget annuel au capital cible : 25x, 28x, 33x, et tout ce qui fait bouger le multiple."},
			{Slug: "les-trois-phases", Title: "Accumulation, transition, retrait : les trois vies d'un plan FIRE", Blurb: "Ce qui change à chaque phase, ce qu'il faut y optimiser, et les gestes de passage."},
			{Slug: "utiliser-la-page-fire", Title: "Utiliser le simulateur FIRE de pofo", Blurb: "Le mode d'emploi complet : les sections dans l'ordre de lecture, chaque contrôle du panneau de paramètres, et les mésusages classiques."},
			{Slug: "erreurs-classiques-fire", Title: "Les dix erreurs qui ruinent un plan FIRE", Blurb: "Les pièges les plus fréquents, du taux irréaliste à l'oubli de la fiscalité, et comment les éviter."},
		},
	},
	{
		Title: "La science du retrait",
		Blurb: "Ce que la recherche sait vraiment du taux de retrait sûr, des classiques aux travaux récents.",
		Articles: []Article{
			{Slug: "etude-trinity", Title: "Bengen, l'étude Trinity et la naissance du taux de retrait sûr", Blurb: "Les études fondatrices de 1994-1998 : ce qu'elles ont montré, et ce qu'on leur fait dire à tort."},
			{Slug: "sequence-des-rendements", Title: "Le risque de séquence : le vrai ennemi du retraité", Blurb: "Pourquoi deux retraités avec le même rendement moyen finissent l'un riche, l'autre ruiné."},
			{Slug: "ruine-et-probabilites", Title: "La probabilité de ruine : la lire, la choisir, ne pas la subir", Blurb: "Ce que mesure vraiment le chiffre des simulateurs, comment choisir son seuil, et pourquoi les décimales mentent."},
			{Slug: "rendements-arithmetiques-geometriques", Title: "Moyenne arithmétique, moyenne géométrique et volatility drag", Blurb: "Pourquoi les rendements des brochures ne sont pas vivables, et la cascade qui mène au taux de retrait."},
			{Slug: "anarkulova-cederburg", Title: "Au-delà des États-Unis : Anarkulova, Cederburg et l'échantillon mondial", Blurb: "Le taux de retrait recalculé sur le siècle entier des pays développés, ses chiffres qui dérangent, et ses critiques."},
			{Slug: "valorisations-et-cape", Title: "Les valorisations (CAPE) et ce qu'elles disent du taux de retrait", Blurb: "Le meilleur prédicteur connu du sort d'un millésime : définition, chiffres, critiques, et les quatre usages légitimes dans un plan."},
			{Slug: "rendements-attendus", Title: "Les rendements attendus prospectifs", Blurb: "Construire un μ défendable : building blocks, les fourchettes de Vanguard à GMO, la précision réelle, et comment ne pas empiler les prudences."},
			{Slug: "horizon-et-esperance-de-vie", Title: "Horizon, espérance de vie et retraites de 50 ans", Blurb: "Le bon quantile de survie, la courbe taux-horizon qui s'aplatit, la ruine pondérée par la mortalité, et la phase à découvert."},
			{Slug: "serie-ern", Title: "La série Safe Withdrawal Rate d'ERN : guide de lecture", Blurb: "La référence moderne du sujet : ses résultats majeurs partie par partie, et les filtres pour la lire depuis la France."},
			{Slug: "les-maths-du-4-pourcent", Title: "Pourquoi 4 % ? L'anatomie mathématique de la règle", Blurb: "La cascade en trois étages (rendement réel, bonus d'amortissement, pénalité de séquence), pourquoi elle est si robuste, et ce qui la casserait."},
			{Slug: "theorie-du-cycle-de-vie", Title: "La théorie du cycle de vie : le socle académique du plan", Blurb: "Modigliani, Samuelson, Merton, Yaari : la part de risque optimale, le capital humain, la rente par défaut, et le registre étiqueté des écarts entre la théorie et ce livre."},
			{Slug: "decider-sous-incertitude", Title: "Décider sous incertitude : utilité, Kelly, regret", Blurb: "Vous ne vivez qu'une trajectoire : l'équivalent certain, tolérance contre capacité, pourquoi fuir le Kelly complet, et le protocole en cinq règles."},
		},
	},
	{
		Title: "Modéliser : Monte-Carlo et autres machines",
		Blurb: "Comprendre les simulateurs de l'intérieur : ce qu'ils savent faire, ce qu'ils inventent, et comment les lire.",
		Articles: []Article{
			{Slug: "monte-carlo-forces-faiblesses", Title: "Monte-Carlo : forces, faiblesses, bon usage", Blurb: "La machine derrière toutes les probabilités de ruine : comment elle marche, ses quatre faiblesses structurelles, et les huit règles du bon usage."},
			{Slug: "historique-vs-parametrique", Title: "Fenêtres historiques, bootstrap, paramétrique : trois familles de modèles", Blurb: "D'où viennent les futurs simulés, quelle question chaque famille sait vraiment traiter, et que faire de leurs désaccords."},
			{Slug: "queues-epaisses", Title: "Queues épaisses, crises et Student-t", Blurb: "Pourquoi les marchés produisent dix fois trop de catastrophes pour la courbe en cloche, et ce que le curseur df décide vraiment."},
			{Slug: "lire-un-fan-chart", Title: "Lire un fan chart et des percentiles sans se tromper", Blurb: "L'anatomie du cône de richesse, sa géométrie qui parle, les cinq erreurs de lecture classiques et les autres éventails de la page."},
			{Slug: "pieges-des-simulateurs", Title: "Les pièges des simulateurs", Blurb: "Pourquoi cinq outils rendent cinq verdicts pour le même plan : les dix pièges hiérarchisés, et la grille d'audit en dix questions."},
			{Slug: "rendre-monte-carlo-pertinent", Title: "Rendre un Monte-Carlo pertinent (blending, régimes, stress)", Blurb: "Les six corrections qui transforment le générateur de nombres en instrument : du blending vers le prior mondial au plan réel simulé."},
			{Slug: "regimes-de-marche", Title: "Les régimes de marché : croissance × inflation, sticky bears", Blurb: "Les saisons des marchés, la grille à quatre quadrants, le cauchemar stagflationniste du rentier, et l'audit de portefeuille par régime."},
		},
	},
	{
		Title: "Les stratégies de retrait",
		Blurb: "Selon quelle règle prélever : le triangle impossible, chaque stratégie en détail, et comment choisir la vôtre.",
		Articles: []Article{
			{Slug: "panorama-strategies-retrait", Title: "Panorama des stratégies de retrait : la carte avant le territoire", Blurb: "Le triangle impossible, les deux extrêmes qui bornent tout, les cinq familles, et les six critères pour les noter honnêtement."},
			{Slug: "retrait-fixe-bengen", Title: "Le retrait fixe indexé (Bengen) : le classique de référence", Blurb: "La règle fondatrice en stratégie opérationnelle : la mécanique fine, la falaise silencieuse, et les trois amendements quasi gratuits."},
			{Slug: "pourcentage-fixe", Title: "Le pourcentage fixe du portefeuille : increvable mais inconfortable", Blurb: "La ruine impossible et la ruine de train de vie, le lissage des dotations (règle de Yale), et le choix du pourcentage."},
			{Slug: "guyton-klinger", Title: "Guyton-Klinger : les guardrails historiques, grandeur et limites", Blurb: "Les quatre règles exactes de 2006, la cascade de coupes des mauvais millésimes, et les correctifs modernes, plancher en tête."},
			{Slug: "vpw", Title: "VPW, le retrait à pourcentage variable des Bogleheads", Blurb: "L'annuité inversée en table gravée : la mécanique exacte, le pont de pension, le test de tolérance à la perte, et la frontière avec l'ABW."},
			{Slug: "regles-cape", Title: "Les règles CAPE : ajuster le retrait aux valorisations (ERN)", Blurb: "Taux = a + b/CAPE : la double contra-cyclicité qui auto-lisse le revenu, les paramètres d'ERN, et la forme aboutie ABW + ancre CAPE."},
			{Slug: "guardrails-morningstar", Title: "Les guardrails modernes (Morningstar) : l'état de l'art", Blurb: "Le juge honnête de Morningstar, l'indicateur par risque de Kitces-Tharp, et la version exécutable avec pofo en instrument."},
			{Slug: "amortissement-abw", Title: "Le retrait par amortissement (ABW/TPAW) : l'approche actuarielle", Blurb: "Le crédit inversé re-coté chaque année : richesse totale, quatre paramètres personnels, et le match final contre les guardrails."},
			{Slug: "plancher-plafond", Title: "Plancher-plafond et règles Vanguard : la flexibilité bornée", Blurb: "Le corridor sur variation (+5 %/−2,5 %) : des glissements au lieu de chutes, une ruine redevenue honnête, deuxième partout."},
			{Slug: "rentes-et-annuites", Title: "Rentes et safety first : acheter un plancher", Blurb: "Les crédits de mortalité, le cadre français (dont le rachat de trimestres, meilleure rente du marché), l'objection inflation, et quand passer en rente."},
			{Slug: "sept-facons-de-vivre", Title: "Sept façons de vivre du même portefeuille", Blurb: "Trois retraites réelles rejouées année par année : ce que chaque règle a versé, quand elle a coupé, et ce qu'elle a laissé sur la table."},
			{Slug: "choisir-sa-strategie", Title: "Choisir sa stratégie : critères, comparatif, cas d'usage", Blurb: "La procédure en cinq étapes : tests d'admissibilité, matrice profils-règles, hybrides par phases, et la page écrite qui conclut."},
		},
	},
	{
		Title: "Le portefeuille de retrait",
		Blurb: "De quoi vivre 40 ans : l'allocation, la dimension temporelle, et les briques qui résistent à tous les régimes.",
		Articles: []Article{
			{Slug: "primes-de-risque", Title: "D'où viennent les rendements : les primes de risque", Blurb: "Pourquoi les actions paient, qui paie la prime, pourquoi elle survit à sa célébrité, et pourquoi l'or ne rapporte rien sans que ce soit un défaut."},
			{Slug: "pourquoi-la-diversification-marche", Title: "Pourquoi la diversification fonctionne", Blurb: "Le free lunch démonté : corrélations, rebalancing premium, volatility harvesting, l'effet doublé en décumulation, et la fausse diversification à débusquer."},
			{Slug: "concevoir-un-portefeuille", Title: "Concevoir un portefeuille : la méthode, pas le modèle", Blurb: "La conception par les risques plutôt que par les actifs : les sept questions dans l'ordre, et l'allocation qui tombe à la fin comme un résultat."},
			{Slug: "allocation-actions-obligations", Title: "L'allocation actions/obligations en retrait", Blurb: "Le plateau 50-80 % qui plonge des deux côtés, les trois curseurs pour s'y placer, et le débat 100 % actions remis à sa place."},
			{Slug: "glidepaths", Title: "Les glidepaths : bond tent, rising equity et la fenêtre fragile", Blurb: "La prudence comme dépense temporaire : les résultats Kitces-Pfau et ERN, l'exécution automatique par les retraits, et le match contre le buffer."},
			{Slug: "portefeuilles-tous-temps", Title: "Les portefeuilles tous-temps : Browne, All-Weather, Golden Butterfly, Dragon", Blurb: "Un gagnant par saison : compositions exactes, chiffres qui comptent pour un rentier, critiques honnêtes, et la dose plutôt que le dogme."},
			{Slug: "actifs-defensifs", Title: "Les actifs défensifs : panorama et rôles", Blurb: "Défendre contre quoi ? Le cahier des charges, la revue candidat par candidat, la galerie des faux défensifs, et l'assemblage."},
			{Slug: "faux-actifs-defensifs", Title: "Les faux actifs défensifs : ce qui en a l'air sans en être", Blurb: "Low vol, dividendes, covered calls, REIT, high yield, private equity, crypto : pourquoi ils semblent défensifs, et pourquoi ils lâchent au test de 2008 et 2022."},
			{Slug: "or-en-retrait", Title: "L'or dans un portefeuille de retrait", Blurb: "Zéro réel séculaire, décorrélation qui survit aux crises : les trois rôles, la dose de chacun, la pratique française et les erreurs."},
			{Slug: "obligations-en-retrait", Title: "Les obligations en retrait : types, durée, rôle exact", Blurb: "Prix et duration, YTM = espérance affichée, les trois services conditionnels au régime, les quatre décisions, et le fonds euros à sa place."},
			{Slug: "obligations-indexees", Title: "Les obligations indexées sur l'inflation", Blurb: "Le seul contrat écrit en réel : le point mort, la leçon de 2022, l'échelle de linkers qui garantit ce que le 4 % espère, et la pratique française."},
			{Slug: "facteurs-fama-french", Title: "Les facteurs (Fama-French, value, momentum) en phase de retrait", Blurb: "Le noyau répliqué, le dossier du rentier (SCV, affinité value-inflation), la tracking error décennale, et le tilt optionnel bien dosé."},
			{Slug: "diversification-internationale", Title: "La diversification internationale (et le biais domestique)", Blurb: "Le risque dominant est le décrochage de VOTRE pays : les destins nationaux chiffrés, le change qui amortit côté actions, et la cible en un ETF."},
			{Slug: "etf-ucits-europeens", Title: "Construire en UCITS : le portefeuille de retrait de l'investisseur européen", Blurb: "Capitalisant, synthétique-PEA, domicile irlandais, tracking difference : les quatre choix, la table des briques, et vendre proprement."},
		},
	},
	{
		Title: "Les actifs alternatifs",
		Blurb: "Au-delà des actions et des obligations : les briques qui diversifient vraiment, ce qu'elles coûtent, et comment les acheter proprement.",
		Articles: []Article{
			{Slug: "managed-futures", Title: "Managed futures et suivi de tendance", Blurb: "Le seul défensif à espérance positive : un siècle de preuves, le crisis alpha des régimes longs, la mise en œuvre UCITS, et l'hiver à traverser."},
			{Slug: "long-volatility", Title: "Long volatility et tail hedging : payer pour les krachs", Blurb: "La convexité qui explose dans les krachs rapides, la prime de variance qui la facture trop cher, les véhicules toxiques, et les rares usages légitimes."},
			{Slug: "global-macro", Title: "Global macro et primes alternatives", Blurb: "Les paris macro qui gagnent dans les régimes hostiles, le catalogue des primes (dont le commodity carry), la purge des fonds ARP, et la grille en cinq questions."},
			{Slug: "primes-d-assurance", Title: "Les primes d'assurance : cat bonds et arbitrage de fusions", Blurb: "Se faire payer pour porter l'ouragan ou la fusion ratée : décorrélation causale, le piège du collatéral en euros, et ce que dix points changent vraiment (peu)."},
			{Slug: "return-stacking", Title: "Return stacking, overlays et portable alpha", Blurb: "Empiler les diversifiants sur le cœur au lieu de les financer en le vendant : la mécanique, le coût du cash, la leçon 2022, et les règles de dose."},
		},
	},
	{
		Title: "Buffers et protections",
		Blurb: "Les amortisseurs du plan : le matelas et ses règles, les buckets démystifiés, l'échelle, et les protections patrimoniales.",
		Articles: []Article{
			{Slug: "cash-buffer", Title: "Le matelas de liquidités : taille, coût, vrai rôle", Blurb: "L'intuition juste, l'arithmétique têtue (±0,5 point), et la vraie valeur : anti-panique, permission de dépenser, gouvernance."},
			{Slug: "cash-ameliore", Title: "Le cash amélioré : monétaire, CLO AAA, fonds euros", Blurb: "Avec quoi remplir la poche courte : les étages mesurés, les CLO AAA expliqués, et la règle d'assemblage en trois couches."},
			{Slug: "strategie-buckets", Title: "Les buckets : la stratégie des seaux, promesse et critique", Blurb: "Une allocation déguisée plus des flux qui sont du rééquilibrage ou du timing : le procès équitable, et la version propre."},
			{Slug: "echelle-obligataire", Title: "Les échelles d'obligations (et l'échelle de linkers)", Blurb: "L'appariement qui annule le risque de taux : le pont vers la pension, le plancher adossé, et la pratique française des fonds à échéance."},
			{Slug: "recharger-ou-pas", Title: "Consommer et recharger un buffer : les règles qui marchent", Blurb: "Le déclencheur de drawdown, la recharge en terrain calme, l'interdiction absolue, et le buffer fondant qui domine le perpétuel."},
			{Slug: "immobilier-en-retrait", Title: "L'immobilier dans un plan FIRE (résidence, locatif)", Blurb: "Le loyer fantôme et le double comptage, le locatif compté en flux décoté, l'arbitrage vendre-ou-garder, SCPI et viager à leur place."},
			{Slug: "levier-et-marges", Title: "Levier, marge et lombard en retrait (avancé)", Blurb: "Pourquoi le levier change de signe à la retraite, les trois seuls usages défendables, et les cinq règles non négociables."},
		},
	},
	{
		Title: "L'inflation",
		Blurb: "La grande tueuse de rentiers : son histoire, sa mesure, son effet exact sur le taux de retrait, et les défenses qui marchent.",
		Articles: []Article{
			{Slug: "inflation-histoire", Title: "L'inflation sur les dernières décennies : ce que 1914-2025 enseigne", Blurb: "Le siècle français en cinq régimes, la mort du mot rentier, la répression financière, et les leçons structurelles pour le plan."},
			{Slug: "suivre-inflation", Title: "Suivre l'inflation : les indices, et la vôtre", Blurb: "IPC, IPCH et leurs angles morts, l'inflation personnelle du retraité (+0,2 à +0,5), les points morts, et le réglage de la dérive."},
			{Slug: "inflation-et-taux-de-retrait", Title: "Inflation et taux de retrait : le lien exact", Blurb: "Les ciseaux et la compression réelle simultanée : pourquoi 1966 bat 1929, les chiffres conditionnels, et l'inventaire d'indexation du plan."},
			{Slug: "se-proteger-de-inflation", Title: "Se protéger de l'inflation : ce qui marche vraiment", Blurb: "L'arsenal classé par nature de preuve : contractuel, structurel, épisodique, comportemental : la liste noire, et l'assemblage par phase."},
			{Slug: "hyperinflation-et-extremes", Title: "Hyperinflations et scénarios extrêmes", Blurb: "Weimar à Chypre : ce qui survit, la couverture structurelle quasi gratuite déjà en place, et le piège symétrique du prepper."},
		},
	},
	{
		Title: "Fiscalité et cadre français",
		Blurb: "Le rentier dans le droit français : enveloppes, imposition des retraits, PUMa, retraite légale, santé, transmission.",
		Articles: []Article{
			{Slug: "enveloppes-francaises", Title: "PEA, assurance-vie, CTO : les enveloppes du rentier français", Blurb: "Chaque enveloppe vue de la décumulation, les horloges à dater tôt, et l'ordre de consommation qui est un panachage annuel."},
			{Slug: "flat-tax-et-imposition", Title: "PFU, barème, abattements : l'imposition des retraits", Blurb: "La part de gain au PMP, l'arbitrage annuel PFU/barème, le lissage des tranches basses, et les purges de fin de partie."},
			{Slug: "taxe-puma", Title: "La taxe PUMa : le piège du rentier français", Blurb: "La cotisation qui vise le rentier d'avant la pension : formule, seuils, zones grises, et les quatre mitigations dont l'interrupteur d'activité."},
			{Slug: "retraite-legale", Title: "FIRE et retraite légale : trimestres, AGIRC-ARRCO, décote", Blurb: "L'actif négligé du plan : ce qu'une carrière écourtée produit, le taux plein automatique à 67 ans, les trimestres à bas coût, et la méthode M@REL."},
			{Slug: "sante-et-protection-sociale", Title: "Santé et protection sociale du rentier", Blurb: "PUMa et ALD comme socle, la mutuelle à charge pleine et sa dérive, le trou de prévoyance du pont, et la dépendance provisionnée."},
			{Slug: "succession-et-transmission", Title: "Succession et transmission", Blurb: "Conjoint exonéré, abattements rechargeables, AV avant 70 ans, démembrement : les six outils dans l'ordre, et la règle d'or qui prime tout."},
			{Slug: "expatriation-fiscale", Title: "L'expatriation : fiscalité et protection sociale", Blurb: "Le joker mental du rentier passé au calcul froid : résidence réelle, régimes datés, les attaches qui suivent, et le solde souvent proche de zéro."},
		},
	},
	{
		Title: "Le facteur humain",
		Blurb: "La partie difficile : dépenser, vivre, durer : ce que les modèles ne voient pas et que les vétérans racontent.",
		Articles: []Article{
			{Slug: "psychologie-du-retrait", Title: "La psychologie du retrait : pourquoi dépenser est si dur", Blurb: "La sous-consommation chronique et la panique du krach : quatre mécanismes, les biais nommés, et la boîte à outils qui équipe tout."},
			{Slug: "temoignages-fire", Title: "Ce que disent les vrais FIRE : témoignages et conseils", Blurb: "Six constats du corpus (l'argent marche, le mur des 18 mois, le regret n° 1 : pas parti assez tôt) et la liste canonique des vétérans."},
			{Slug: "sens-et-identite", Title: "Sens, identité, structure : la vie après le travail", Blurb: "Les quatre chantiers que le salaire masquait, les pièges du profil optimiseur, et le prototypage qui fait tout l'écart."},
			{Slug: "couple-et-famille", Title: "FIRE en couple et en famille", Blurb: "Le sport d'équipe : l'asymétrie d'appétence, le départ décalé, la gouvernance à deux (le quiz inversé), le divorce lucide, enfants et parents."},
			{Slug: "flexibilite-realite", Title: "La flexibilité : mythe et réalité", Blurb: "Ce qu'elle vaut vraiment (0,3-0,5 point), pourquoi la durée bat la profondeur, les six formes classées, et le plancher testé, pas déclaré."},
			{Slug: "une-annee-de-plus", Title: "Le syndrome de l'année de plus", Blurb: "L'asymétrie comptable qui le fabrique, le chiffrage des deux colonnes, l'OMY rationnel en trois cas, et l'engagement pris à froid."},
			{Slug: "retour-au-travail", Title: "Barista, coast, side income : le travail choisi", Blurb: "Le meilleur actif défensif du livre : les formes, le quadruplé français, l'option qui fond et s'entretient, et la boîte à outils."},
		},
	},
	{
		Title: "En pratique",
		Blurb: "Le chantier et le pilotage : construire le plan, le faire vivre, traverser les tempêtes, et les cas complets.",
		Articles: []Article{
			{Slug: "construire-son-plan", Title: "Construire son plan pas à pas", Blurb: "Les sept étapes dans l'ordre, le gabarit du plan d'une page à huit blocs, la recette de validation, et l'exécution du jour 1."},
			{Slug: "revue-annuelle", Title: "La revue annuelle : la check-list du rentier", Blurb: "Une séance par an, sept blocs, le quiz inversé et la ligne non financière : la revue constate et exécute, elle ne re-conçoit pas."},
			{Slug: "quand-s-inquieter", Title: "Quand s'inquiéter, quand laisser courir", Blurb: "Le taux courant en voyant central, le tri bruit/signal, la ruine qui prévient des années à l'avance, et le playbook à cinq paliers."},
			{Slug: "marche-baissier-en-retraite", Title: "Traverser un marché baissier en retraite : le playbook", Blurb: "Trois marchés baissiers, la semaine 1 du rien actif, les gestes mécaniques du creux (dont la récolte de moins-values), les interdits, et la sortie."},
			{Slug: "revenus-complementaires", Title: "Pensions et revenus complémentaires dans le plan", Blurb: "Trois catégories, trois traitements : les quatre mécanismes d'un flux, la saisie exacte dans pofo, et le test du plan trop adossé."},
			{Slug: "depenses-en-retraite", Title: "Les dépenses réelles en retraite (retirement smile, Die With Zero)", Blurb: "Le sourire de Blanchett, les time buckets datés, le front-loading assumé du FIRE précoce, et l'ingénierie du budget à étages temporels."},
			{Slug: "cas-types", Title: "Trois plans complets, chiffrés de bout en bout", Blurb: "Le solo précoce, le couple classique, le tardif au pont court : trois conceptions déroulées, validées, et leurs leçons croisées."},
		},
	},
	{
		Title: "Références",
		Blurb: "Le lexique, la bibliothèque annotée, et la machine qui calcule le livre.",
		Articles: []Article{
			{Slug: "lexique", Title: "Lexique du FIRE et du retrait", Blurb: "Tous les termes du livre et du jargon des forums, définis et reliés à leur chapitre : parcouru d'un trait, un résumé alphabétique."},
			{Slug: "bibliotheque", Title: "La bibliothèque : sites, papiers, livres, outils", Blurb: "Chaque référence annotée (pourquoi la lire, par où la prendre), les sources officielles françaises, et trois parcours de lecture."},
			{Slug: "la-machine-pofo", Title: "Sous le capot : comment pofo calcule ce livre", Blurb: "Les données, le pipeline du modèle central, les six colonnes, le noyau mensuel, les limites assumées, et comment vérifier soi-même."},
		},
	},
}

// Titles maps every written article's slug to its display title; it is the
// titles argument ToHTML expects for resolving [[slug]] links.
func (e *Edition) Titles() map[string]string {
	m := make(map[string]string)
	for _, cat := range e.Categories {
		for _, a := range cat.Articles {
			m[a.Slug] = a.Title
		}
	}
	return m
}

// find returns the article and its category, or ok=false.
func (e *Edition) find(slug string) (Article, Category, bool) {
	for _, cat := range e.Categories {
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
	"les-maths-du-4-pourcent", "theorie-du-cycle-de-vie",
	"decider-sous-incertitude",
	// III. Modéliser
	"monte-carlo-forces-faiblesses", "historique-vs-parametrique",
	"queues-epaisses", "lire-un-fan-chart", "pieges-des-simulateurs",
	"rendre-monte-carlo-pertinent", "regimes-de-marche",
	// IV. Les stratégies de retrait
	"panorama-strategies-retrait", "retrait-fixe-bengen", "pourcentage-fixe",
	"guyton-klinger", "vpw", "regles-cape", "guardrails-morningstar",
	"amortissement-abw", "plancher-plafond", "rentes-et-annuites",
	"sept-facons-de-vivre", "choisir-sa-strategie",
	// V. Le portefeuille de retrait
	"primes-de-risque", "pourquoi-la-diversification-marche",
	"concevoir-un-portefeuille",
	"allocation-actions-obligations", "glidepaths", "portefeuilles-tous-temps",
	"actifs-defensifs", "faux-actifs-defensifs", "or-en-retrait",
	"obligations-en-retrait", "obligations-indexees", "facteurs-fama-french",
	"diversification-internationale", "etf-ucits-europeens",
	// V bis. Les actifs alternatifs
	"managed-futures", "long-volatility", "global-macro", "primes-d-assurance",
	"return-stacking",
	// VI. Buffers et protections
	"cash-buffer", "cash-ameliore", "strategie-buckets", "echelle-obligataire",
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
