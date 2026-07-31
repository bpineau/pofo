# La taxe PUMa : le piège du rentier français

Il existe un prélèvement que la plupart des candidats FIRE français découvrent trop tard, souvent par un courrier de l'URSSAF deux ans après leur départ : la cotisation subsidiaire maladie (CSM), dite « taxe PUMa ». Son principe tient en une phrase qui vise **exactement** le profil de ce livre : celui qui vit de son capital sans activité professionnelle ni pension doit cotiser à l'assurance maladie sur ses revenus du capital : environ 6,5 % par an sur l'essentiel des revenus du patrimoine, **par-dessus** toute la fiscalité du chapitre précédent ([[flat-tax-et-imposition]]).

Pour un rentier précoce en phase de pont, c'est fréquemment plusieurs milliers d'euros par an : l'équivalent de 0,2 à 0,4 point de taux de retrait : et, bonne nouvelle au milieu du piège, c'est aussi l'un des prélèvements les plus **pilotables** du système : ses seuils, son assiette annuelle et son interrupteur (une petite activité l'éteint intégralement) en font un cas d'école d'optimisation légale de plan. Cet article démonte tout : l'histoire et la logique du dispositif, la formule exacte et ses seuils, qui est touché et qui y échappe, l'assiette (avec ses zones grises, nommées comme telles), les quatre stratégies de mitigation classées par efficacité, et l'intégration au plan et à pofo.

Avertissement redoublé ici : **le dispositif a déjà changé deux fois (2016, 2019) et reste politiquement instable : chiffres à jour de 2026, vérification annuelle obligatoire.**

::: cle Le principe, et pourquoi il vous vise
Depuis 2016, la Protection Universelle Maladie (PUMa) garantit la prise en charge des frais de santé à tout résident stable : y compris celui qui ne cotise plus nulle part. La contrepartie est la CSM : si vos revenus d'**activité** sont inférieurs à un seuil (~20 % du plafond de la Sécurité sociale, soit ~9 500 €/an) et que vos revenus du **capital** dépassent un autre seuil (~50 % du même plafond, soit ~23 500 €/an), vous cotisez ~6,5 % sur ces revenus du capital au-delà de la franchise. Traduction : le salarié, l'indépendant, le **retraité pensionné** et le chômeur indemnisé n'y sont pas soumis : le dispositif ne frappe, en pratique, qu'une population : le rentier d'avant la pension : vous, pendant la phase de pont ([[horizon-et-esperance-de-vie]]).
:::

## La formule, les seuils, la facture

**La formule (état 2019-2026).** La cotisation annuelle vaut approximativement :

> CSM ≈ 6,5 % × (revenus du capital − 0,5 × PASS) × (1 − revenus d'activité / (0,2 × PASS))

avec le PASS (plafond annuel de la Sécurité sociale) autour de 47-48 k€ ces années-ci, l'assiette des revenus du capital plafonnée à 8 PASS (~380 k€), et la cotisation due seulement si les **deux** conditions sont réunies (activité < 20 % du PASS, capital > 50 % du PASS). Le dernier facteur est la décote linéaire : plus vos revenus d'activité approchent du seuil, plus la cotisation fond : jusqu'à **zéro** au seuil : l'interrupteur dont tout découle.

**Les ordres de grandeur.** Un rentier sans activité avec 60 000 € de revenus du capital imposables : CSM ≈ 6,5 % × (60 000 − 23 700) ≈ **2 360 €/an**. Avec 120 000 € (une grosse année de plus-values) : ≈ 6 260 €. Le maximum théorique (assiette plafonnée) : ~23 000 €/an. Et l'appel arrive avec un décalage : la CSM de l'année N est calculée sur la déclaration de N et appelée fin N+1 : le fameux courrier-surprise : budgétez-la l'année où vous générez l'assiette, pas l'année où elle tombe.

**L'assiette : ce qui compte.** Les revenus du capital au sens du dispositif recouvrent pour l'essentiel les revenus fonciers, les revenus de capitaux mobiliers (dividendes, intérêts, dont les gains des rachats d'assurance-vie) et les plus-values imposables : en gros, ce que votre déclaration fait apparaître comme revenus du patrimoine. D'où les leviers structurels : ce qui n'apparaît **pas** (les gains latents non réalisés : l'enveloppe capitalisante par excellence, [[etf-ucits-europeens]]) n'entre pas dans l'assiette : et d'où aussi des **zones grises** documentées (le traitement fin des retraits de PEA, de certaines PV, l'articulation exacte avec le RFR) sur lesquelles la doctrine URSSAF a évolué : cet article s'interdit d'y trancher : le simulateur et le rescrit URSSAF font foi, chaque année ([[revue-annuelle]]).

## Qui est touché, qui échappe : la cartographie

**Touchés** : le rentier précoce sans aucune activité, pendant **toute** la phase de pont : c'est la population cible, presque nommément. **Exemptés de droit** : les titulaires de revenus d'activité au-dessus du seuil ; les titulaires d'une **pension** de retraite (de base ou complémentaire, même modeste : la liquidation éteint la CSM à vie : un argument de plus au dossier des trimestres, [[retraite-legale]]) ; les indemnisés chômage : ce qui fait des allocations post-rupture conventionnelle un sursis de CSM en début de FIRE ; et les situations de couple où les conditions ne sont pas réunies : l'appréciation tenant compte de la situation du foyer (l'activité suffisante d'un conjoint change la donne : les modalités exactes de calcul en couple ont fait l'objet d'ajustements : vérifier votre cas au simulateur, sans supposer).

Notez la philosophie du zigzag : le dispositif initial de 2016 (8 % sur une assiette plus large, seuils différents) a été contesté, contentieux à l'appui, puis recalibré en 2019 vers la formule actuelle, plus douce : l'histoire enseigne que les paramètres **bougent**, dans les deux sens : la CSM est un poste de veille, pas une constante du plan ([[hyperinflation-et-extremes]] : la « saisie réglementaire douce », en voici l'exemple ordinaire).

::: science Les quatre stratégies de mitigation, classées
**1. L'interrupteur d'activité (radical).** ~9 500 €/an de revenus d'activité (micro-entreprise, salariat partiel, quelques missions) annulent la CSM intégralement par la décote : pour un rentier à 3 000 €+ de CSM, le « taux de rémunération » implicite de cette petite activité est spectaculaire : elle rapporte son revenu **plus** l'économie de cotisation : c'est l'argument fiscal du Barista FIRE ([[retour-au-travail]]), et souvent la réponse la plus simple pour qui garde de toute façon un pied professionnel. Attention à la substance : l'activité doit être réelle (revenus déclarés, cotisations d'activité payées) : le montage fictif est un redressement en attente.
**2. Le pilotage d'assiette (structurel).** La CSM taxe les revenus **réalisés** de l'année : les mêmes gestes que le chapitre précédent la pilotent : capitaliser (les ETF capitalisants ne créent aucun revenu tant qu'on ne vend pas), doser les rachats AV (seule la part de **gain** entre en compte), étaler les grosses plus-values sur plusieurs années (l'assiette a une franchise annuelle de ~23,7 k€ : trois années à 40 k€ de revenus du capital paient moins qu'une année à 120 k€), et arbitrer l'ordre des robinets en intégrant la CSM au taux mixte ([[enveloppes-francaises]], [[flat-tax-et-imposition]] : le lissage de taux devient un lissage taux + CSM). Tension à arbitrer honnêtement : le lissage **fiscal** pousse à réaliser plus dans les années creuses, la CSM pousse à réaliser moins : l'optimum conjoint réalise jusqu'à la franchise CSM et les tranches basses, pas au-delà.
**3. La liquidation d'une pension (terminale).** La CSM meurt avec la première pension : pour les fins de pont, avancer la liquidation d'un an peut se comparer (décote de pension contre années de CSM) : un calcul de cas, à faire chiffré.
**4. La provision (le minimum vital).** Si vous ne faites rien d'autre : budgétez-la : ~6,5 % des revenus du capital réalisés au-delà de la franchise, dans les dépenses du plan : c'est l'option « je paie pour la simplicité », légitime aussi ([[combien-il-vous-faut]] : la majoration de friction de 10-15 % du budget la contenait déjà).
:::

## L'intégration au plan, et à pofo

**Dans le dimensionnement** : la CSM est une friction sur les revenus réalisés de la phase de pont : pour un plan organisé (enveloppes capitalisantes, rachats dosés), elle ressort typiquement à 0,5-2 k€/an ; pour un plan naïf (gros CTO distribuant, grosses ventes annuelles), 3-8 k€/an : l'écart est un argument d'organisation de plus. **Dans pofo** : le curseur fiscal unique ([[utiliser-la-page-fire]]) l'approxime : son défaut de 31,4 % correspond peu ou prou à « PFU + une CSM moyenne » : si vous avez calibré votre taux mixte ([[flat-tax-et-imposition]]), ajoutez-y l'équivalent CSM de **votre** structure (la part des revenus réalisés dans vos flux × 6,5 %, au-delà de la franchise) : et souvenez-vous que la CSM s'éteint à la pension : le taux mixte de la phase adossée est plus bas : une raison de plus pour laquelle le plan français type s'améliore tout seul avec l'âge ([[horizon-et-esperance-de-vie]]).

**Dans la gouvernance** : trois réflexes : le **simulateur** URSSAF chaque année de pont (l'assiette et la doctrine bougent : dix minutes) ; la **provision** comptable l'année de réalisation (l'appel arrive en N+1 : le plan de trésorerie doit le savoir, [[revue-annuelle]]) ; et la **veille** législative (le dispositif est un marqueur politique : chaque loi de financement de la Sécurité sociale peut le retoucher : c'est le poste fiscal du plan FIRE à surveiller nommément).

::: exemple Trois versions du même pont, CSM comprise
Léa et Sam, phase de pont de 12 ans, besoin 50 000 €/an, patrimoine organisé ([[enveloppes-francaises]]). **Version** A (naïve) : tout en CTO distribuant + grosses ventes : revenus du capital réalisés ~70 k€/an : CSM ~3 000 €/an × 12 ans ≈ 36 000 € de pont. **Version** B (organisée) : capitalisant partout, rachats AV sous abattement, PEA, ventes CTO dosées : revenus réalisés ~32 k€/an : CSM ~540 €/an ≈ 6 500 € de pont : l'organisation vaut ~30 000 €. **Version** C (interrupteur) : Sam garde une micro-activité de conseil à ~10 k€/an qui lui plaît : CSM : zéro : et les 10 k€ réduisent d'autant les retraits : la meilleure assurance anti-séquence du livre se paie elle-même ([[revenus-complementaires]], [[sequence-des-rendements]]). Le choix entre B et C n'est pas fiscal : il est de vie : mais il se fait en connaissant les trois chiffres.
:::

## L'essentiel à retenir

- La CSM (« taxe PUMa ») vise exactement le rentier d'avant la pension : ~6,5 % des revenus du capital au-delà d'une franchise (~0,5 PASS ≈ 23,7 k€), si les revenus d'activité sont sous ~0,2 PASS (≈ 9,5 k€) : assiette plafonnée, appel en N+1, formule retouchée deux fois depuis 2016 : veille annuelle obligatoire.
- Salariés, indépendants, chômeurs indemnisés et **pensionnés** y échappent : elle ne concerne que la phase de pont : et s'éteint à vie à la première pension liquidée.
- Quatre mitigations, par efficacité : l'interrupteur d'activité (~9,5 k€ réels d'activité = zéro CSM : l'argument fiscal du Barista), le pilotage d'assiette (capitaliser, doser les rachats, étaler les PV : l'optimum conjoint avec le lissage fiscal réalise jusqu'à la franchise, pas au-delà), la liquidation calculée, la provision budgétée.
- Ordre de grandeur : 0,5-2 k€/an pour un plan organisé, 3-8 k€ pour un plan naïf : soit 0,2-0,4 point de taux de retrait : à intégrer au taux mixte de pofo (le défaut 31,4 % l'approxime), en le baissant pour la phase adossée.
- Réflexes de gouvernance : simulateur URSSAF chaque année, provision l'année de réalisation, veille sur chaque loi de financement : c'est le prélèvement instable du plan français.

---

## Pour aller plus loin

- urssaf.fr : la page cotisation subsidiaire maladie et son simulateur : la source opérationnelle, à consulter chaque année de pont.
- L'article L380-2 du code de la Sécurité sociale et le décret de 2019 : les textes, pour les amateurs de sources primaires.
- Les analyses des contentieux 2016-2019 (doctrine et jurisprudence) : l'histoire du zigzag, instructive sur l'instabilité du dispositif.
- Dans ce livre : [[flat-tax-et-imposition]] (le taux mixte où l'intégrer), [[enveloppes-francaises]] (le pilotage d'assiette), [[retour-au-travail]] (l'interrupteur d'activité comme choix de vie), [[retraite-legale]] (l'extinction par la pension).
