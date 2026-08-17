package firebook

import "regexp"

// The English edition's figures.
//
// The ~100 plates stay SINGLE-SOURCE: every generator keeps its French
// literals and is never duplicated, never parameterized by language. The
// English edition post-processes the RENDERED SVG instead, translating its
// text nodes: the dictionary first, then a mechanical number reformat for the
// payloads that carry no words. A payload that is neither is left alone AND
// fails the coverage guard, so a plate cannot quietly ship half-translated.
//
// The consequence is deliberate: renaming a label in a plate an English
// article uses breaks the tests until figureDict is updated in the same
// commit. Figures change rarely and the dictionary lives next to the code, so
// they get the hard guard; prose gets the soft drift report instead.
//
// Captions are not figure text: they live in the article markdown and are
// translated with the prose.

// figureDict translates one rendered text payload. A key is either the bare
// French payload, which applies to every plate, or "<figure-id>|<french>",
// which wins over the bare key for that plate alone (the same French string
// sometimes needs different English in different plates).
//
// Payloads are already SVG-escaped, so a key carrying "&" spells it "&amp;".
// The dictionary grows with the translation campaign: an entry is needed only
// once an English article uses the plate.
var figureDict = map[string]string{
	// vol-drag
	"Le volatility drag : même moyenne, richesses opposées": "Volatility drag: same average, opposite outcomes",
	"régulier : +7 % chaque année":                          "steady: +7% every year",
	"volatil : +27 % / −13 % (même moyenne)":                "volatile: +27% / −13% (same average)",
	"années  →":                      "years  →",
	"richesse (× capital de départ)": "wealth (× starting capital)",

	// sequence-risk
	"Deux séquences, même rendement moyen, retraits identiques": "Two sequences, same average return, identical withdrawals",
	"krach précoce : ruine": "early crash: ruin",
	"krach tardif : survit": "late crash: survives",
	"années de retraite  →": "years of retirement  →",
	"capital de départ":     "starting capital",

	// millesimes-1966-1982
	"MILLÉSIME 1966 CONTRE MILLÉSIME 1982":           "1966 VINTAGE AGAINST 1982 VINTAGE",
	"Le même plan, deux dates de départ":             "The same plan, two starting dates",
	"capital réel restant, en millions d'euros":      "real capital left, in millions of euros",
	"années écoulées depuis le départ à la retraite": "years since retirement began",
	"même plan des deux côtés : 1 M€, 40 k€ retirés chaque année, indexés sur l'inflation, jamais ajustés": "same plan on both sides: EUR 1M, EUR 40k a year, indexed to inflation, never adjusted",
	"sous la mise de départ":     "below the starting stake",
	"départ en 1966":             "starting in 1966",
	"départ en 1982":             "starting in 1982",
	"millésime 1966":             "1966 vintage",
	"millésime 1982":             "1982 vintage",
	"réel, dix premières années": "real, first ten years",
	"réel, trente ans":           "real, thirty years",
	"à l'arrivée":                "at the finish",
	"zéro en 1994, l'année 29":   "zero in 1994, year 29",
	"zéro":                       "zero",
	"4,6 M€":                     "EUR 4.6M",
	"−1,2 %/an":                  "−1.2%/yr",
	"+4,2 %/an":                  "+4.2%/yr",
	"+7,0 %/an":                  "+7.0%/yr",
	"+11,4 %/an":                 "+11.4%/yr",
	"Le portefeuille de 1966 a pourtant rapporté plus que le retrait. C'est l'ordre des années qui a tué le plan.":                  "The 1966 portfolio still earned more than it paid out. The order of the years is what killed the plan.",
	"60/40 américain réel (S&amp;P 500, Treasuries 5 ans, déflatés CPI-U), reconstruction du livre ; retrait fixe, sans fiscalité.": "Real US 60/40 (S&amp;P 500, 5-year Treasuries, CPI-U deflated), reconstructed for this book; fixed withdrawal, no taxes.",

	// millesimes-soutenables
	"SOIXANTE-SIX DÉPARTS EN RETRAITE":                                                                     "SIXTY-SIX RETIREMENT DATES",
	"Le taux qui a tenu trente ans, millésime par millésime":                                               "The rate that held for thirty years, vintage by vintage",
	"Retrait initial maximal qu'un 50/50 américain aurait soutenu trente ans, en pouvoir d'achat constant": "Highest initial withdrawal a US 50/50 would have carried for thirty years, in constant purchasing power",
	"66 millésimes, de 1926 à 1991, le dernier dont les trente années sont complètes":                      "66 vintages, 1926 to 1991, the last one whose thirty years are complete",
	"la règle":      "the rule",
	"médiane":       "median",
	"1982 · 10,2 %": "1982 · 10.2%",
	"2,2 points":    "2.2 points",
	"sous la règle : six départs, de 1964 à 1969": "below the rule: six retirement dates, 1964 to 1969",
	"1966 · 3,67 %": "1966 · 3.67%",
	"Le millésime médian aurait supporté 60 % de dépenses en plus que celui de 1966.":                                                               "The median vintage would have carried 60% more spending than the 1966 one.",
	"6 millésimes sur 66 passent sous 4 %, et ce sont six départs consécutifs, de 1964 à 1969. Les 60 autres tiennent la règle.":                    "6 vintages out of 66 fall below 4%, and they are six retirement dates in a row, 1964 to 1969. The other 60 hold the rule.",
	"Panel Jorda-Schularick-Taylor, États-Unis : actions et obligations d'État domestiques, rendements réels annuels déflatés du CPI américain.":    "Jorda-Schularick-Taylor panel, United States: domestic stocks and government bonds, annual real returns deflated by US CPI.",
	"50/50 rééquilibré chaque année, prélèvement en début d'année, ni frais ni impôt ; le capital a le droit de finir exactement à zéro.":           "50/50 rebalanced every year, withdrawal at the start of the year, no fees and no taxes; the capital is allowed to end at exactly zero.",
	"Bengen publie 4,15 % pour 1966. Sa reconstruction n'est pas celle-ci : obligations à moyen terme, autre indice d'actions, données mensuelles.": "Bengen publishes 4.15% for 1966. His reconstruction is not this one: intermediate bonds, another stock index, monthly data.",
	"L'ordre de grandeur tient, la décimale non. Le pire millésime, lui, est le même des deux côtés.":                                               "The ballpark holds, the decimal does not. The worst vintage is the same on both sides.",

	// Shared by several plates. An age tick reads "40 ans" in French and
	// "age 40" in English, which is the same width.
	"VPW":          "VPW",
	"normal":       "normal",
	"revenu":       "income",
	"40 ans":       "age 40",
	"47 ans":       "age 47",
	"55 ans":       "age 55",
	"57 ans":       "age 57",
	"67 ans":       "age 67",
	"70 ans":       "age 70",
	"77 ans":       "age 77",
	"85 ans":       "age 85",
	"87 ans":       "age 87",
	"95 ans":       "age 95",
	"99 ans":       "age 99",
	"pont":         "bridge",
	"capital (k€)": "capital (EUR k)",

	// vpw-table
	"Le pourcentage monte avec l'âge, le capital fond exprès":    "The percentage rises with age, the capital melts on purpose",
	"le taux de retrait de la table (% du portefeuille)":         "the table's withdrawal rate (% of portfolio)",
	"le rendement supposé, 3,3 % réel : le plancher de la table": "the assumed return, 3.3% real: the table's floor",
	"à 91 ans : 13 %, puis jusqu'à 100 %":                        "at 91: 13%, then up to 100%",
	"vingt-cinq ans de quasi-pourcentage fixe":                   "twenty-five years of a near-flat rate",
	"le capital, consommé jusqu'à zéro":                          "the capital, spent down to zero",
	"38,5 k€/an, plat":                                           "EUR 38.5k/yr, flat",
	"krach de −30 % à 70 ans : 27,0 k€/an, sans retour":          "a −30% crash at 70: EUR 27.0k/yr, no way back",
	"la vie que cette table produit, à partir de 40 ans avec 1 M€, si le marché sert les rendements supposés": "the life this table produces, from age 40 with EUR 1M, if the market delivers the assumed returns",
	"Annuité inversée à g = 3,3 % réel constant, horizon jusqu'à 100 ans.":                                    "Reversed annuity at a constant g = 3.3% real, horizon out to age 100.",

	// vpw-pont
	"Le pont de pension : vingt ans de revenu qui n'existe pas encore":                  "The pension bridge: twenty years of income that does not exist yet",
	"revenu réel servi (k€/an), ménage de 47 ans, 1,6 M€, pensions de 21,6 k€ à 67 ans": "real income served (EUR k/yr), household aged 47, EUR 1.6M, pensions of EUR 21.6k at 67",
	"le pont : 21,6 k€/an,":                                            "the bridge: EUR 21.6k/yr,",
	"prélevés sur 356 k€ d'obligations":                                "drawn from EUR 356k of bonds",
	"la pension, enfin liquidée":                                       "the pension, finally running",
	"le VPW sur les 1 243 k€ restants : 50,0 k€/an":                    "VPW on the remaining EUR 1,243k: EUR 50.0k/yr",
	"71,6 k€/an":                                                       "EUR 71.6k/yr",
	"67 ans : la pension prend le relais":                              "age 67: the pension takes over",
	"test de tolérance, actions −50 % : 56,6 k€, le pont ne bouge pas": "tolerance test, stocks −50%: EUR 56.6k, the bridge holds",
	"le plancher du ménage, 38 k€":                                     "the household floor, EUR 38k",
	"Aux rendements supposés de la table (3,3 % réel pour la part VPW, 1,9 % pour actualiser le pont), sans fiscalité.": "At the table's assumed returns (3.3% real for the VPW sleeve, 1.9% to discount the bridge), no taxes.",

	// vpw-test-de-perte
	"LE TEST DE TOLÉRANCE À LA PERTE":               "THE LOSS TOLERANCE TEST",
	"Le même choc, avec et sans le pont de pension": "The same shock, with and without the pension bridge",
	"actions −50 %":                                 "stocks −50%",
	"Avec le pont de pension":                       "With the pension bridge",
	"le choc laisse le ménage au-dessus du confort": "the shock leaves the household above comfort",
	"Sans le pont":                                  "Without the bridge",
	"le même choc passe sous le confort":            "the same shock drops below comfort",
	"un couple de 47 ans, 1,6 M€ en 60/40, pensions de 21,6 k€ à 67 ans ; revenu servi, k€ par an": "a couple aged 47, EUR 1.6M in 60/40, pensions of EUR 21.6k at 67; income served, EUR k a year",
	"pointillé bleu : le confort visé, 52 ; pointillé brun : le plancher, 38":                      "blue dashes: target comfort, 52; brown dashes: the floor, 38",

	// Shared by the two plates of how-much-you-need.
	"COMBIEN IL VOUS FAUT": "HOW MUCH YOU NEED",

	// cible-convexite
	"Chaque demi-point de prudence coûte plus cher que le précédent":                                          "Every extra half point of caution costs more than the one before",
	"le multiple est l'inverse exact du taux : les crochets donnent le prix de chaque demi-point de prudence": "the multiple is the exact inverse of the rate: the brackets price each extra half point of caution",
	"capital cible, en multiples de dépenses annuelles":                                                       "target capital, in multiples of annual spending",
	"+ 3,6x": "+ 3.6x",
	"+ 4,8x": "+ 4.8x",
	"+ 6,7x": "+ 6.7x",
	"25x":    "25x",
	"28,6x":  "28.6x",
	"33x":    "33x",
	"40x":    "40x",
	"au-delà de 33x, le risque dominant n'est plus la ruine :": "past 33x, the dominant risk is no longer ruin:",
	"c'est d'avoir travaillé des années de trop":               "it is having worked years too many",
	"taux de retrait initial":                                  "initial withdrawal rate",
	"plus prudent, plus cher →":                                "more cautious, more expensive →",
	"De 25x à 33x, l'écart représente typiquement 3 à 6 ans de travail de plus pour un taux d'épargne de 40 à 50 %.":   "From 25x to 33x, the gap is typically 3 to 6 more years of work at a savings rate of 40 to 50%.",
	"Courbe exacte, sans hypothèse de marché : c'est la définition du taux de retrait, pas un résultat de simulation.": "Exact curve, no market assumption: this is the definition of the withdrawal rate, not a simulation result.",

	// cible-cascade. The friction step names the French levies in the plate;
	// the English edition states what it is instead, as the article does.
	"Du relevé bancaire au capital cible, marche par marche":                              "From bank statement to capital target, step by step",
	"Nadia et Marc, étape 5 : à 3,5 %, un euro de dépense par mois pèse 343 € de capital": "Nadia and Marc, step 5: at 3.5%, one euro of monthly spending weighs EUR 343 of capital",
	"capital cible (k€)": "target capital (EUR k)",
	"étape 1":            "step 1",
	"étape 2":            "step 2",
	"étape 4":            "step 4",
	"dépenses observées": "observed spending",
	"24 mois de relevés": "24 months of statements",
	"3 400 €/mois":       "EUR 3,400/mo",
	"voyages et loisirs": "travel and leisure",
	"la vie visée":       "the life you aim at",
	"+ 350 €/mois":       "+ EUR 350/mo",
	"mutuelle santé":     "health coverage",
	"à votre charge":     "on your own tab",
	"+ 220 €/mois":       "+ EUR 220/mo",
	"friction fiscale":   "tax friction",
	"impôts et PUMa":     "tax on withdrawals",
	"+ 12 % du brut":     "+ 12% of gross",
	"cible":              "target",
	"1 547 000 €":        "EUR 1,547,000",
	"− 200 000 € : la retraite légale, comptée en revenu différé":                                                     "− EUR 200,000: the state pension, counted as deferred income",
	"sans elle, le même plan aurait exigé 1 747 000 €":                                                                "without it, the same plan would have demanded EUR 1,747,000",
	"Les deux marches que le calcul de comptoir oublie, la fiscalité et la pension, pèsent chacune plus lourd que le": "The two steps the barstool calculation leaves out, tax and the pension, each weigh more than the",
	"budget voyages, et en sens opposés. Les hypothèses sont celles de l'étape 5 : remplacez chacune par la vôtre.":   "travel budget, and they pull in opposite directions. The assumptions are step 5's: replace each with your own.",

	// flux-relatif-phases
	"LES TROIS PHASES": "THE THREE PHASES",
	"ACCUMULATION":     "ACCUMULATION",
	"TRANSITION":       "TRANSITION",
	"RETRAIT":          "WITHDRAWAL",
	"Le flux de l'année, rapporté au portefeuille où il tombe":                                                            "The year's flow, measured against the portfolio it lands in",
	"Une seule variable gouverne les trois phases : ce que vous versez ou prélevez dans l'année, en % du portefeuille.":   "One variable governs the three phases: what you pay in or draw out over the year, as a % of the portfolio.",
	"Projection déterministe sur les hypothèses de l'article, pas une simulation : 34 % d'épargne, 5 % de rendement réel": "A deterministic projection on the article's assumptions, not a simulation: 34% saved, 5% real return",
	"jusqu'au départ à 45 ans, puis un portefeuille dé-risqué à 4 % réel. Ni krach, ni bonne surprise.":                   "until the exit at 45, then a de-risked portfolio at 4% real. No crash, no happy surprise.",
	"axe coupé à +15 % : le ratio vaut 49 % à 22 ans, et diverge au premier euro versé":                                   "axis cut at +15%: the ratio is 49% at age 22, and diverges on the first euro paid in",
	"vous versez":   "you pay in",
	"vous prélevez": "you draw",
	"à 30 ans, le versement de l'année vaut encore 8 % du capital :": "at 30, the year's contribution is still 8% of the capital:",
	"un krach est effacé par les versements suivants.":               "the next few contributions erase a crash.",
	"À la transition, le portefeuille n'a jamais été":                "At the transition the portfolio has never been",
	"aussi gros par rapport au flux qui pourrait le réparer.":        "this large against the flow that could repair it.",
	"45 ans : le flux change de signe,":                              "45: the flow changes sign,",
	"de +2,1 % à −4,1 % du portefeuille, du jour au lendemain.":      "from +2.1% to −4.1% of the portfolio, overnight.",
	"à 85 ans :": "at 85:",
	"âge, de la première épargne sérieuse au grand âge": "age, from the first serious saving year to old age",

	// cout-des-erreurs
	"ERREURS CLASSIQUES":                                                "CLASSIC MISTAKES",
	"Six erreurs, un même plan, un coût mesuré":                         "Six mistakes, one plan, a measured cost",
	"probabilité de ruine du plan (%)":                                  "failure probability of the plan (%)",
	"le plan propre : 19,3 % de ruine":                                  "the clean plan: 19.3% failure probability",
	"Portefeuille mono-régime, tout obligataire":                        "One-regime portfolio, all bonds",
	"0 % d'actions au lieu de 60 %":                                     "0% equities instead of 60%",
	"+20,2 pts":                                                         "+20.2 pts",
	"Flexibilité surestimée":                                            "Flexibility overestimated",
	"4,5 % au lieu de 3,5 %, le taux qu'une coupe de 40 % justifierait": "4.5% instead of 3.5%, the rate a 40% cut would justify",
	"+16,6 pts":                           "+16.6 pts",
	"Dépenses sous-estimées de 20 %":      "Spending underestimated by 20%",
	"42 k€/an au lieu de 35 k€":           "EUR 42k/yr instead of EUR 35k",
	"+10,9 pts":                           "+10.9 pts",
	"Pension oubliée":                     "Pension forgotten",
	"aucune pension comptée dans le plan": "no pension counted in the plan",
	"+9,5 pts":                            "+9.5 pts",
	"payée en années de travail":          "paid in working years",
	"Taux de 4 % au lieu de 3,5 %":        "A 4% rate instead of 3.5%",
	"40 k€/an au lieu de 35 k€":           "EUR 40k/yr instead of EUR 35k",
	"+7,4 pts":                            "+7.4 pts",
	"Fiscalité et PUMa ignorées":          "Tax on withdrawals ignored",
	"12 % de friction non budgétée, soit 39,2 k€ à sortir": "12% friction never budgeted: EUR 39.2k to withdraw",
	"+6,1 pts": "+6.1 pts",
	"Plan de référence : 1 M€, 35 k€/an réels (3,5 %), 50 ans, 60/40, pension de 14 k€/an à l'année 20, coupe tenue de 10 %.": "Reference plan: EUR 1M, EUR 35k/yr real (3.5%), 50 years, 60/40, a EUR 14k/yr pension from year 20, a 10% cut held.",
	"Modèle empirique : 16 pays, 1870-2020, blocs de 10 ans, 200 000 tirages. Une seule entrée change par barre.":             "Empirical model: 16 countries, 1870 to 2020, 10-year blocks, 200,000 draws. One input changes per bar.",
	"Le tout-actions, lui, ne se voit pas sur cet axe (18,3 %) : il coûte 20 années de budget coupé sur 50, contre 16,8.":     "All-equity does not show on this axis (18.3%): it costs 20 years of cut budget out of 50, against 16.8.",

	// anarkulova-cederburg
	"SEIZE PAYS, 1870-2020":                                                          "SIXTEEN COUNTRIES, 1870 TO 2020",
	"Le taux de retrait sûr est d'abord une affaire de pays":                         "The safe withdrawal rate is above all a matter of country",
	"Taux initial d'un 60/40 domestique, tenu 30 ans dans le pire millésime du pays": "Initial rate of a domestic 60/40, held for 30 years in the country's worst vintage",
	"pire millésime":                     "worst vintage",
	"fenêtres":                           "windows",
	"la règle des 4 %":                   "the 4% rule",
	"panier mondial équipondéré · 3,3 %": "equal-weighted world basket · 3.3%",
	"Danemark":                           "Denmark",
	"États-Unis":                         "United States",
	"Australie":                          "Australia",
	"Japon*":                             "Japan*",
	"Pays-Bas":                           "Netherlands",
	"Suisse":                             "Switzerland",
	"Royaume-Uni":                        "United Kingdom",
	"Norvège":                            "Norway",
	"Belgique":                           "Belgium",
	"Suède":                              "Sweden",
	"Espagne":                            "Spain",
	"Finlande":                           "Finland",
	"Italie":                             "Italy",
	"Allemagne*":                         "Germany*",
	"Portugal":                           "Portugal",
	"France":                             "France",
	"Panel Jorda-Schularick-Taylor : 60 % actions / 40 % obligations du pays, réels, rééquilibrés, prélèvement en début d'année.":       "Jorda-Schularick-Taylor panel: 60% domestic stocks / 40% domestic bonds, real, rebalanced, withdrawal at the start of the year.",
	"Seules comptent les fenêtres de 30 ans complètes, d'où des comptes très inégaux d'un pays à l'autre.":                              "Only complete 30-year windows count, hence very uneven counts from one country to the next.",
	"* Japon et Allemagne : 1946-1947 et 1944-1948 manquent au panel, et leurs pires fenêtres avec. Ces deux barres sont trop longues.": "* Japan and Germany: 1946-1947 and 1944-1948 are missing from the panel, and their worst windows with them. These two bars are too long.",
	"Pfau et Anarkulova-Cederburg calculent sur d'autres bases : leur SAFEMAX américain ressort vers 4 %, ici 3,75 %.":                  "Pfau and Anarkulova-Cederburg compute on other databases: their US SAFEMAX comes out near 4%, here 3.75%.",
	"Le panel est très dur avec la France d'avant 1950 : −0,2 %/an réel en actions sur 1900-2020, mais +3,4 %/an sur 1950-2020.":        "The panel is very hard on pre-1950 France: −0.2%/yr real in stocks over 1900 to 2020, but +3.4%/yr over 1950 to 2020.",
	// cape-swr
	"Plus le marché est cher au départ, plus le taux soutenable baisse": "The more expensive the market at the start, the lower the sustainable rate",
	"marché bon marché": "cheap market",
	"marché cher":       "expensive market",
	"règle des 4 %":     "the 4% rule",
	"taux sûr":          "safe rate",
	"CAPE au départ  →": "CAPE at the start  →",

	// cape-dix-ans
	"Le CAPE d'un mois, et les dix années réelles qui l'ont suivi":                                              "One month's CAPE, and the ten real years that followed",
	"chaque point : un mois de départ, et le rendement réel annualisé du S&amp;P 500 sur les 120 mois suivants": "each dot: a start month, and the annualized real return of the S&amp;P 500 over the next 120 months",
	"1 242 départs mensuels de janvier 1913 à juin 2016, total return déflaté du CPI américain":                 "1,242 monthly start dates from January 1913 to June 2016, total return deflated by US CPI",
	"LE CENTRE BOUGE, LA LARGEUR RESTE":     "THE CENTER MOVES, THE WIDTH STAYS",
	"rendement réel annualisé, en % par an": "annualized real return, % a year",
	"ajustement des moindres carrés":        "least-squares fit",
	"rendement = 1,0 + 0,86 × 100 / CAPE":   "return = 1.0 + 0.86 × 100 / CAPE",
	"R² = 0,29":                             "R² = 0.29",
	"9 à 11":                                "9 to 11",
	"18 à 22":                               "18 to 22",
	"27 à 33":                               "27 to 33",
	"164 départs":                           "164 starts",
	"219 départs":                           "219 starts",
	"48 départs":                            "48 starts",
	"Le trait épais couvre huit départs sur dix, la colonne claire les couvre tous. À CAPE 9 à 11, la décennie":       "The thick bar covers eight start dates in ten, the pale column covers them all. At a CAPE of 9 to 11, the decade",
	"a payé de +2,7 à +18,9 % réels par an ; à CAPE 18 à 22, de −3,6 à +14,4 %. Le centre descend, la largeur reste.": "paid +2.7 to +18.9% real a year; at a CAPE of 18 to 22, −3.6 to +14.4%. The center drops, the width stays.",
	"La colonne 27 à 33 ne compte que 48 départs (1929, la fin des années 1990, 2013-2015) : largeur sous-mesurée.":   "The 27 to 33 column holds only 48 start dates (1929, the late 1990s, 2013 to 2015): its width is under-measured.",

	// rendements-attendus -> expected-returns
	"RENDEMENTS ATTENDUS": "EXPECTED RETURNS",
	"L'espérance d'une obligation est affichée, celle d'une action se construit": "A bond's expectation is printed on the label, a stock's has to be built",
	"classe d'actifs":                         "asset class",
	"espérance réelle":                        "real expectation",
	"Actions US":                              "US stocks",
	"S&amp;P 500, conditions 2024-2026":       "S&amp;P 500, 2024 to 2026 conditions",
	"0 à −1,5":                                "0 to −1.5",
	"terme de valorisation":                   "valuation term",
	"2,5 à 4,0 %":                             "2.5 to 4.0%",
	"Actions hors US":                         "Non-US stocks",
	"Europe, Japon, émergents":                "Europe, Japan, emerging markets",
	"4,0 à 6,0 %":                             "4.0 to 6.0%",
	"Obligations euro":                        "Government bonds",
	"OAT ou Bund 10 ans, nominal":             "10-year bond, nominal",
	"inflation anticipée":                     "expected inflation",
	"Monétaire euro":                          "Cash",
	"taux directeur réel":                     "real policy rate",
	"une seule brique, et elle suit le cycle": "one block only, and it follows the cycle",
	"0 à 1 %":                                 "0 to 1%",
	"Or":                                      "Gold",
	"ni coupon ni bénéfices":                  "no coupon, no earnings",
	"rien à empiler : le chiffre est une hypothèse, pas une lecture": "nothing to stack: the number is an assumption, not a reading",
	"briques-esperance|distribution":                                 "payout",
	"croissance des bénéfices":                                       "earnings growth",
	"Briques pleines : ce que les prix affichent déjà. Sous la barre, le terme de valorisation, le seul disputé, qui porte toute la fourchette.": "Solid blocks: what today's prices already show. Below the bar, the valuation term, the only disputed one, carries the whole range.",
	// escalier-morningstar, same article
	"Le taux « sûr » de Morningstar bouge avec les conditions d'entrée": "Morningstar's \"safe\" rate moves with the entry conditions",
	"taux de retrait initial recommandé, en % (30 ans, 90 % de succès)": "recommended initial withdrawal rate, in % (30 years, 90% success)",
	"la fourchette de travail du livre : 3 à 3,5 %":                     "the book's working range: 3 to 3.5%",
	"taux nuls,":                   "zero rates,",
	"actions chères":               "expensive stocks",
	"valorisations":                "valuations",
	"dégonflées":                   "deflated",
	"taux obligataires":            "bond yields",
	"restaurés":                    "restored",
	"escalier-morningstar|actions": "stocks",
	"redevenues chères":            "expensive again",
	"méthode revue,":               "method revised,",
	"pas les marchés":              "not the markets",
	"conditions d'entrée : le CAPE de Shiller au 30 septembre, date d'arrêté de chaque édition":                     "entry conditions: Shiller's CAPE on September 30, the data cutoff of each edition",
	"Le CAPE n'est qu'une des deux conditions : 2023 monte grâce aux taux obligataires restaurés, pas aux actions.": "The CAPE is only one of the two conditions: 2023 rises on restored bond yields, not on stocks.",
	"2025 monte pour une autre raison encore : à méthode 2024 inchangée, le 50/50 sortait à 3,6 % (tireté).":        "2025 rises for yet another reason: on the unchanged 2024 method, the 50/50 came out at 3.6% (dashed).",
	// horizon-flatten
	"Le taux soutenable par horizon : la courbe qui s'aplatit": "The sustainable rate by horizon: the curve that flattens",
	"chute rapide":           "steep fall",
	"(30 → 40 ans)":          "(30 → 40 years)",
	"au-delà, quasi plat":    "beyond that, almost flat",
	"≈ perpétuité (~3,25 %)": "≈ perpetuity (~3.25%)",
	"un plan qui tient 40 ans tient (presque) toujours": "a plan that holds 40 years holds (almost) forever",
	"horizon du plan (années)  →":                       "plan horizon (years)  →",

	// vivant-ruine-parti
	"RUINE ET MORTALITÉ": "FAILURE AND MORTALITY",
	"Vivant, ruiné ou parti : les trois états, année par année": "Alive, broke or gone: the three states, year by year",
	"de ruine brute": "raw failure",
	"de ruine vécue, un jour vivant et ruiné": "lived failure, alive and broke one day",
	"→":                                    "→",
	"part des scénarios (%)":               "share of scenarios (%)",
	"décédé":                               "gone",
	"le couple n'est plus là":              "the couple is no longer there",
	"vivant et solvable":                   "alive and solvent",
	"le plan tient":                        "the plan holds",
	"vivant et ruiné":                      "alive and broke",
	"le seul état qui coûte quelque chose": "the only state that costs anything",
	"au pic, à l'année 35":                 "at the peak, in year 35",
	"année du plan (départ à 47 ans, donc 87 ans à l'année 40)":                                                   "plan year (leaving at 47, so age 87 in year 40)",
	"Plan : 1 M€, 33 k€/an réels (3,3 %), 53 ans, pension de 14 k€/an à l'année 20, coupe tenue de 10 %.":         "Plan: EUR 1M, EUR 33k a year real (3.3%), 53 years, pension of EUR 14k a year from year 20, a 10% cut held.",
	"Modèle : 16 pays, 1871-2020, 60/40, blocs de 10 ans, 200 000 tirages. Gompertz unisexe, couple de même âge.": "Model: 16 countries, 1871 to 2020, 60/40, 10-year blocks, 200,000 draws. Unisex Gompertz, couple of the same age.",
	"À l'année 40, deux couples sur trois ont encore un survivant, d'où une remise d'un cinquième seulement.":     "In year 40, two couples in three still have a survivor, so the discount is only a fifth.",

	// the-math-of-4-percent
	"LES MATHS DU 4 %": "THE MATH OF 4%",
	"Du rendement au taux de retrait : la cascade": "From return to withdrawal rate: the cascade",
	"rendement réel":                 "real return",
	"géométrique du 60/40":           "geometric, on a 60/40",
	"bonus d'amortissement":          "amortization bonus",
	"(consommer le capital, 30 ans)": "(spending capital, 30 years)",
	"pénalité de séquence":           "sequence penalty",
	"(survivre au pire ordre)":       "(surviving the worst order)",
	"taux de retrait sûr":            "safe withdrawal rate",
	"(la règle de Bengen)":           "(Bengen's rule)",
	"Le clavier des leviers : ce que chaque hypothèse déplace":                                                       "The bank of levers: what each assumption moves",
	"chaque hypothèse déplacée seule, à partir du plan de référence : 60/40 historique, 30 ans, retrait fixe indexé": "each assumption moved on its own, from the reference plan: historical 60/40, 30 years, fixed indexed withdrawal",
	"fourchettes : le plein s'arrête à la borne basse, la teinte va jusqu'à la haute":                                "ranges: the solid bar stops at the low bound, the tint runs to the high one",
	"le plan de référence":                              "the reference plan",
	"taux de retrait obtenu, en % du capital de départ": "resulting withdrawal rate, as a % of starting capital",
	"CAPE élevé au départ":                              "High CAPE at the start",
	"étage 1, le rendement":                             "tier 1, the return",
	"Échantillon mondial":                               "World sample",
	"−0,5 à −1,0":                                       "−0.5 to −1.0",
	"Horizon de 50 ans":                                 "50-year horizon",
	"étage 2, le bonus":                                 "tier 2, the bonus",
	"Frais de 0,5 % par an":                             "Fees of 0.5% a year",
	"Règle flexible à plancher":                         "Flexible rule with a floor",
	"étage 3, la pénalité":                              "tier 3, the penalty",
	"+0,3 à +0,5":                                       "+0.3 to +0.5",
	// utilite-ce
	"DÉCIDER SOUS INCERTITUDE":                                   "DECIDING UNDER UNCERTAINTY",
	"L'équivalent certain : ce que vaut vraiment un plan risqué": "The certainty equivalent: what a risky plan is really worth",
	"utilité (bien-être)":                                        "utility (well-being)",
	"revenu annuel (k€)":                                         "annual income (EUR k)",
	"l'utilité croît de":                                         "utility grows more",
	"moins en moins vite":                                        "and more slowly",
	"mauvais monde (20 k€)":                                      "bad world (EUR 20k)",
	"bon monde (65 k€)":                                          "good world (EUR 65k)",
	"la loterie 50/50":                                           "the 50/50 lottery",
	"E = 42,5":                                                   "E = 42.5",
	"ÉC ≈ 36":                                                    "CE ≈ 36",
	"le prix du risque":                                          "the price of the risk",
	// monte-carlo-strengths-and-limits
	"MONTE-CARLO": "MONTE CARLO",
	"Ce qui déplace la ruine : les hypothèses, pas le nombre de tirages": "What moves failure: the assumptions, not the number of draws",
	"probabilité de ruine (%)": "failure probability (%)",
	"On bouge μ de ±0,5 point (un écart indétectable dans les données), à N = 10 000": "μ moves plus or minus 0.5 point (undetectable in the data), at N = 10,000",
	"μ 4,5 %":           "μ 4.5%",
	"μ 5,0 %":           "μ 5.0%",
	"μ 5,5 %":           "μ 5.5%",
	"×2,1 sur la ruine": "×2.1 on failure",
	"On multiplie N par dix, à hypothèses figées (μ = 5,0 %)": "N goes up tenfold, assumptions frozen (μ = 5.0%)",
	"N 1 000":  "N 1,000",
	"± 1,5 pt": "± 1.5 pts",
	"N 4 000":  "N 4,000",
	"± 0,7 pt": "± 0.7 pts",
	"N 10 000": "N 10,000",
	"± 0,5 pt": "± 0.5 pts",
	"1 M€, 32 k€/an réels (3,2 %), 35 ans, Student-t σ 11 %, df 5. Barres : erreur d'échantillonnage à 95 %.": "EUR 1M, EUR 32k a year real (3.2%), 35 years, Student-t σ 11%, df 5. Bars: 95% sampling error.",

	// fat-tails
	"À volatilité égale, deux mondes : la cloche et ses queues": "Same volatility, two worlds: the bell and its tails",
	"loi normale":                          "normal law",
	"Student-t (df 5)":                     "Student-t (df 5)",
	"les années ordinaires se ressemblent": "ordinary years look alike",
	"la queue épaisse":                     "the fat tail",
	"~10× plus d'années à −30 %":           "~10× more years at −30%",
	"−30 % réel":                           "−30% real",
	"rendement réel annuel (%)  →":         "annual real return (%)  →",
	"densité":                              "density",
	// reading-a-fan-chart
	"Une pile de distributions par date, pas un faisceau de chemins": "A stack of distributions by date, not a bundle of paths",
	"ruine (le zéro)":     "ruin (the zero line)",
	"prospère":            "prospers",
	"finit ruiné":         "ends ruined",
	"× capital de départ": "× starting capital",
	"Le bas du cône, dans les dix premières années, décide": "The bottom of the fan, over the first ten years, decides",
	"Plan défendu":                  "Defended plan",
	"le bas s'enfonce lentement":    "the bottom sinks slowly",
	"Plan tendu":                    "Stretched plan",
	"le 5e percentile pique à zéro": "the 5th percentile dives to zero",
	"années  \u2192  (la pente du bas, sur la première décennie, est l'exposition à la séquence)": "years  \u2192  (the slope of the bottom, over the first decade, is your sequence exposure)",

	// regime-grid (market-regimes)
	"Prospérité":           "Prosperity",
	"actions, obligations": "stocks, bonds",
	"Surchauffe":           "Overheating",
	"matières 1res, or":    "commodities, gold",
	"Déflation":            "Deflation",
	"obligations longues":  "long bonds",
	"Stagflation":          "Stagflation",
	"or, linkers, trend":   "gold, linkers, trend",
	"croissance +":         "growth +",
	"croissance −":         "growth −",
	"inflation basse":      "low inflation",
	"inflation haute":      "high inflation",
	"Les quatre régimes : un gagnant par saison": "The four regimes: one winner per season",
	// withdrawal-strategies-overview
	"l'idéal (inatteignable)":         "the ideal (unreachable)",
	"Bengen (fixe)":                   "Bengen (fixed)",
	"guardrails":                      "guardrails",
	"ABW / TPAW":                      "ABW / TPAW",
	"% fixe":                          "fixed %",
	"probabilité de ruine ↑":          "failure probability ↑",
	"variabilité du niveau de vie  →": "swings in the standard of living  →",
	"On ne supprime pas le risque : on choisit sa forme": "You do not remove the risk, you choose its shape",
	// familles-information, same article
	"LA CARTE DES FAMILLES": "THE MAP OF THE FAMILIES",
	"Plus une règle écoute, mieux elle se place, plus elle exige de vous":                                    "The more a rule listens, the better it sits, and the more it asks of you",
	"cercle plein quand toute la famille l'écoute, demi-cercle quand certaines règles seulement, vide sinon": "full circle when the whole family listens, half circle when only some of its rules do, empty otherwise",
	"familles-information|le":        "the",
	"portefeuille":                   "portfolio",
	"familles-information|l'horizon": "the remaining",
	"familles-information|restant":   "horizon",
	"familles-information|les":       "the",
	"familles-information|les flux":  "future",
	"familles-information|futurs":    "flows",
	"gouvernance":                    "governance",
	"exigée":                         "required",
	"Fixe":                           "Fixed",
	"Bengen et ses amendements":      "Bengen and its amendments",
	"Proportionnelle":                "Proportional",
	"pourcentage fixe, VPW":          "fixed percentage, VPW",
	"À garde-fous":                   "Guardrails",
	"Guyton-Klinger, corridor Vanguard, guardrails par risque": "Guyton-Klinger, Vanguard corridor, risk-based guardrails",
	"Actuarielle":            "Actuarial",
	"ABW, TPAW, règles CAPE": "ABW, TPAW, CAPE rules",
	"Par plancher garanti":   "Guaranteed floor",
	"rentes viagères, échelle de linkers, rachats de trimestres":                                                          "life annuities, linker ladder, delayed claiming",
	"n'écoute rien non plus, mais sort le plancher du portefeuille : elle change de terrain, pas de place sur la carte":   "listens to nothing either, but takes the floor out of the portfolio: it changes the game, not its place on the map",
	"Les colonnes se remplissent de haut en bas, et la jauge de droite avec elles. C'est tout l'arbitrage de la partie :": "The columns fill up from top to bottom, and the gauge on the right rises with them. That is the tradeoff of this whole part:",
	"le gain de position se mesure, le coût se paie en discipline, et une règle abandonnée ne vaut plus rien.":            "the gain in position can be measured, the cost is paid in discipline, and an abandoned rule is worth nothing.",
	// retrait-fixe-bengen: bengen-falaise, bengen-millesimes.
	"LE VOYANT DU RETRAIT FIXE":                                                               "THE FIXED RULE'S WARNING LIGHT",
	"La falaise silencieuse : dix-neuf ans de préavis":                                        "The silent cliff: nineteen years of warning",
	"taux de retrait courant lu chaque 1er janvier (retrait de l'année / portefeuille), en %": "current withdrawal rate read every January 1 (the year's withdrawal / portfolio), in %",
	"la zone d'alerte : 8 à 10 %":                                                             "the alert zone: 8 to 10%",
	"au-delà, aucun marché ne rattrape":                                                       "past it, no market catches up",
	"janvier 1975 : 8,2 %":                                                                    "January 1975: 8.2%",
	"1991 à 1993 : 31, 38 puis 58 %":                                                          "1991 to 1993: 31, 38 then 58%",
	"1994 : plus rien":                                                                        "1994: nothing left",
	"capital encore là ce jour-là, en k€ réels":                                               "capital still there that day, in real EUR k",
	"dix-neuf ans entre le premier signal et le zéro":                                         "nineteen years between the first signal and zero",
	"LE LEGS QU'ON NE CHOISIT PAS":                                                            "THE BEQUEST YOU DO NOT CHOOSE",
	"Trente-trois départs, deux pathologies opposées, une seule règle":                        "Thirty-three start dates, two opposite pathologies, one rule",
	"600 000 € et 24 000 € par an indexés, tenus quarante ans : ce qui restait à l'arrivée, en k€ réels": "EUR 600,000 and EUR 24,000 a year indexed, held for forty years: what was left at the end, in real EUR k",
	"un point par année de départ, de 1954 à 1986, sur le 60/40 américain du livre":                      "one dot per start year, 1954 to 1986, on the book's US 60/40",
	"M€":                            "EUR M",
	"la mise de départ, 600 k€":     "the original stake, EUR 600k",
	"année du départ à la retraite": "year of retirement",
	"la même chose, triée":          "the same thing, sorted",
	"le médian":                     "the median",
	"6 départs sur 33 finissent à zéro, et ils sont tous voisins : c'est la même mauvaise décennie.":            "6 start dates out of 33 end at zero, and they are all neighbors: it is the same bad decade.",
	"12 autres laissent plus de trois fois la mise, jusqu'à neuf fois. Le médian, lui, rend le capital intact.": "12 others leave more than three times the original stake, up to nine times. The median hands back the capital intact.",
	"La règle ne choisit pas entre ses deux pathologies, et vous non plus : c'est le tirage qui décide.":        "The rule does not choose between its two pathologies, and neither do you: the draw decides.",
	"CHOISIR LE POURCENTAGE": "CHOOSING THE PERCENTAGE",
	"La borne géométrique : au-delà, le revenu s'érode pour toujours":           "The geometric bound: above it, income erodes forever",
	"1 M€, rendement réel géométrique de 4 %/an, revenu servi en k€, sans aléa": "EUR 1M, a real geometric return of 4% a year, income delivered in EUR k, no randomness",
	"années de retraite": "years of retirement",
	"w = 3 %":            "w = 3%",
	"w = 4 %":            "w = 4%",
	"w = 5 %":            "w = 5%",
	"w = 6 %":            "w = 6%",
	"22 ans":             "22 years",
	"La bascule est à w = g / (1 + g), soit 3,8 % ici.":                                                          "The tipping point is w = g / (1 + g), which is 3.8% here.",
	"En dessous, le portefeuille croît plus vite qu'on ne le ponctionne et le revenu monte doucement.":           "Below it, the portfolio grows back faster than the draw takes out and income rises slowly.",
	"Au-dessus, chaque année prélève plus que la croissance : le 6 % part au double du 3 % et finit en dessous.": "Above it, each year takes out more than growth puts in: the 6% starts at twice the 3% and ends below it.",
	"LE LISSAGE DU POURCENTAGE":                                                                  "SMOOTHING THE PERCENTAGE",
	"Le même krach, trois façons de le transmettre au ménage":                                    "The same crash, three ways of passing it to the household",
	"1,4 M€, w = 4 %, à travers les rendements réels du 60/40 de 1973 à 1984 ; revenu servi, k€": "EUR 1.4M, w = 4%, through the real returns of the 60/40 from 1973 to 1984; income delivered, EUR k",
	"pointillé : les 45 k€ sous lesquels le ménage entame son incompressible":                    "dotted line: the EUR 45k below which the household eats into its floor",
	"corridor borné":   "bounded corridor",
	"règle de Yale":    "Yale rule",
	"pourcentage brut": "raw percentage",
	"56,0 au départ":   "56.0 at the start",
	"années passées sous 45 k€, une case par an": "years spent below EUR 45k, one cell per year",
	"corridor": "corridor",
	"Yale":     "Yale",
	"brut":     "raw",
	"Le confort du corridor se paie sur le capital : fin 1984 il reste 715 k€ contre 908 au brut. Il a emprunté à sa propre suite.": "The corridor's comfort is paid out of capital: EUR 715k left in 1984 against 908 raw. It borrowed from its own future.",
	"GUYTON-KLINGER": "GUYTON-KLINGER",
	"La cascade de coupes, et ce qu'un plancher change":      "The cascade of cuts, and what a floor changes",
	"revenu réel servi (k€/an), taux initial 4,3 % sur 1 M€": "real income delivered (EUR k a year), initial rate 4.3% on EUR 1M",
	"avec plancher à 78 % : la descente s'arrête à −22 %":    "with a 78% floor: the descent stops at −22%",
	"sans plancher : cinq coupes, −57 %":                     "no floor: five cuts, −57%",
	"43 k€ pendant quatre ans":                               "EUR 43k for four years",
	"Le portefeuille, lui, s'en sort : 886 k€ sans plancher, 313 k€ avec. La diète a payé la survie du capital.":       "The portfolio survives either way: EUR 886k with no floor, EUR 313k with one. The diet bought that survival.",
	"Millésime 1966, 60/40 américain réel ; règles 3 et 4 seules (corridor ±20 %, ajustements ±10 %), sans fiscalité.": "1966 vintage, real US 60/40; rules 3 and 4 only (corridor ±20%, adjustments ±10%), no taxes.",
	// cape-contracyclique
	"LA DOUBLE CONTRACYCLICITÉ":                                                       "DOUBLE COUNTERCYCLICALITY",
	"Le revenu, produit de deux facteurs en sens inverse":                             "Income, the product of two factors that move in opposite directions",
	"un portefeuille de 1,5 M€ à CAPE 30, frappé par un krach puis par une euphorie":  "a EUR 1.5M portfolio at a CAPE of 30, hit by a crash and then by a euphoria",
	"les courbes fines joignent les couples qui servent le même revenu, en k€ par an": "the thin curves join the pairs that deliver the same income, in EUR k a year",
	"portefeuille (M€ réels)":                                                         "portfolio (EUR M, real)",
	"taux servi (%)":                                                                  "rate paid (%)",
	"35,9 k€":                                                                         "EUR 35.9k",
	"pourcentage fixe":                                                                "fixed percentage",
	"71,8 k€":                                                                         "EUR 71.8k",
	"43,4 k€":                                                                         "EUR 43.4k",
	"krach, CAPE 21":                                                                  "crash, CAPE 21",
	"64,4 k€":                                                                         "EUR 64.4k",
	"euphorie, CAPE 38":                                                               "euphoria, CAPE 38",
	"51,3 k€":                                                                         "EUR 51.3k",
	"aujourd'hui, CAPE 30":                                                            "today, CAPE 30",
	"Trait plein : la règle CAPE, qui longe presque une courbe de revenu constant.":     "Solid line: the CAPE rule, which almost rides a constant-income curve.",
	"Pointillé : le pourcentage fixe, qui les traverse toutes et prend le choc entier.": "Dotted line: the fixed percentage, which cuts across them all and takes the whole shock.",

	// cape-depuis-1881
	"LE TAUX N'EST PAS UNE CONSTANTE":                                                               "THE RATE IS NOT A CONSTANT",
	"Ce que la règle CAPE aurait servi, chaque janvier depuis 1881":                                 "What the CAPE rule would have paid, every January since 1881",
	"taux de retrait a + b / CAPE (a = 1,75, b = 0,5) appliqué au CAPE du 1er janvier, en % par an": "withdrawal rate a + b / CAPE (a = 1.75, b = 0.5) applied to the CAPE on January 1, in % a year",
	"les 4 % de Bengen":        "Bengen's 4%",
	"l'après-guerre : CAPE 5":  "the postwar low: CAPE 5",
	"la bulle de 1929":         "the 1929 bubble",
	"le pire millésime":        "the worst vintage",
	"la fin de la stagflation": "the end of stagflation",
	"la bulle internet":        "the internet bubble",
	"La règle n'a presque jamais dit 4 %. Elle a dit 11,5 % en 1921 et 2,9 % en 2000, parce que le taux soutenable": "The rule almost never said 4%. It said 11.5% in 1921 and 2.9% in 2000, because the sustainable rate",
	"est une fonction du prix payé, pas une constante de la nature. Aux extrêmes elle déraille, faute de borne :":   "is a function of the price paid, not a constant of nature. At the extremes it derails, for lack of a cap:",
	"personne n'aurait dû retirer 11 % en 1921, et c'est le défaut d'une formule linéaire en 1 / CAPE.":             "nobody should have withdrawn 11% in 1921, and that is the flaw of a formula linear in 1 / CAPE.",
}

var (
	// reFigTspan and reFigText extract the LEAF text payloads of a rendered
	// plate: a <text> wrapping <tspan>s contributes its tspans, not itself
	// (the childless-<text> pattern cannot match it).
	reFigTspan = regexp.MustCompile(`<tspan([^>]*)>([^<]*)</tspan>`)
	reFigText  = regexp.MustCompile(`<text([^>]*)>([^<]*)</text>`)

	// reFrenchDecimal and reFrenchPercent flag French number formatting
	// surviving into English output; the guard tests share them.
	reFrenchDecimal = regexp.MustCompile(`\d,\d`)
	reFrenchPercent = regexp.MustCompile(`\d[\x{00a0}\x{202f} ]%`)

	// reNeutral matches a payload made only of numbers and symbols, with no
	// word to translate: it needs no dictionary entry, only reformatting.
	reNeutral = regexp.MustCompile(`^[\d\s\p{Zs}%×+.,:;()'’/\-−–]+$`)
	reDigit   = regexp.MustCompile(`\d`)

	reThousandsGroup = regexp.MustCompile(`(\d),(\d\d\d)`)

	reDecimalComma = regexp.MustCompile(`(\d),(\d)`)
	reThousands    = regexp.MustCompile(`(\d)[\x{00a0}\x{202f} ](\d\d\d)`)
	rePercentSpace = regexp.MustCompile(`(\d)[\x{00a0}\x{202f} ]?%`)
)

// hasFrenchDecimal reports whether a payload still carries a French decimal
// comma. English thousands separators ("EUR 1,243k") are commas between digits
// too, so they are folded away first; a comma left between digits after that
// is a decimal separator that escaped the translation.
func hasFrenchDecimal(s string) bool {
	for {
		next := reThousandsGroup.ReplaceAllString(s, "$1$2")
		if next == s {
			break
		}
		s = next
	}
	return reFrenchDecimal.MatchString(s)
}

// isNeutralPayload reports whether a text payload carries numbers and symbols
// only, so anglicizeNumbers alone can translate it.
func isNeutralPayload(s string) bool {
	return reNeutral.MatchString(s) && reDigit.MatchString(s)
}

// anglicizeNumbers rewrites French number formatting in one text payload:
// decimal commas become points, (narrow) no-break-space thousands become
// commas, and the space before a percent sign disappears ("6,6 %" -> "6.6%",
// "1 000 000" -> "1,000,000"). Words are left alone; translating them is the
// dictionary's job.
func anglicizeNumbers(s string) string {
	// In French figure text a comma between two digits is always a decimal
	// separator: thousands are grouped with spaces.
	s = reDecimalComma.ReplaceAllString(s, "$1.$2")
	// One pass consumes one separator, so "1 000 000" needs two.
	for {
		next := reThousands.ReplaceAllString(s, "$1,$2")
		if next == s {
			break
		}
		s = next
	}
	return rePercentSpace.ReplaceAllString(s, "$1%")
}

// translatePayload turns one French payload into English: the plate-scoped
// dictionary key first, then the global one, then the mechanical number
// reformat for neutral payloads. Anything else comes back unchanged, which is
// what the coverage guard reports.
func translatePayload(id, payload string) string {
	if en, ok := figureDict[id+"|"+payload]; ok {
		return en
	}
	if en, ok := figureDict[payload]; ok {
		return en
	}
	if isNeutralPayload(payload) {
		return anglicizeNumbers(payload)
	}
	return payload
}

// figureTextNodes lists every text payload of a rendered plate, in document
// order: the tspans first, then the childless <text> elements. The guard tests
// use it to check that a plate is fully covered.
func figureTextNodes(svg string) []string {
	var out []string
	for _, m := range reFigTspan.FindAllStringSubmatch(svg, -1) {
		out = append(out, m[2])
	}
	for _, m := range reFigText.FindAllStringSubmatch(svg, -1) {
		out = append(out, m[2])
	}
	return out
}

// FigureSVGEnglish renders a plate and translates its text nodes to English.
// It is English.Figure; the plate generators themselves stay French and are
// never duplicated. An untranslatable payload is left in French, which the
// coverage guard test turns into a failure.
func FigureSVGEnglish(id string) string {
	svg := translateNodes(id, FigureSVG(id), reFigTspan, "tspan")
	return translateNodes(id, svg, reFigText, "text")
}

// translateNodes rewrites the payload of every <tag> the regexp matches,
// keeping the element's attributes (capture 1) verbatim.
func translateNodes(id, svg string, re *regexp.Regexp, tag string) string {
	return re.ReplaceAllStringFunc(svg, func(node string) string {
		m := re.FindStringSubmatch(node)
		return "<" + tag + m[1] + ">" + translatePayload(id, m[2]) + "</" + tag + ">"
	})
}
