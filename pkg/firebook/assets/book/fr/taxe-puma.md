# La taxe PUMa : le piège du rentier français
<!-- edition: fr-only: French law end to end, replaced by the US-framework part in the English edition -->

Un prélèvement échappe à la plupart des candidats FIRE français. Ils le découvrent trop tard, souvent par un courrier de l'URSSAF deux ans après leur départ. C'est la cotisation subsidiaire maladie (CSM), dite « taxe PUMa ». Son principe vise exactement le profil de ce livre. Celui qui vit de son capital, sans activité professionnelle ni pension, doit cotiser à l'assurance maladie sur ses revenus du capital. Comptez environ 6,5 % par an sur les revenus du patrimoine au-delà d'une franchise, par-dessus toute la fiscalité du chapitre précédent ([[flat-tax-et-imposition]]).

Pour un rentier précoce en phase de pont, la facture atteint souvent plusieurs milliers d'euros par an, soit l'équivalent de 0,2 à 0,4 point de taux de retrait. Bonne nouvelle au milieu du piège, c'est aussi l'un des prélèvements les plus pilotables du système. Ses seuils, son assiette annuelle et son interrupteur en font un cas d'école d'optimisation légale du plan (une petite activité l'éteint intégralement). Cet article déroule le dossier. L'histoire et la logique du dispositif, la formule exacte et ses seuils, qui est touché et qui y échappe, l'assiette et ses zones grises, nommées comme telles, les quatre parades classées par efficacité, enfin l'intégration au plan et au simulateur.

Avertissement redoublé ici. Le dispositif a été créé en 2016, refondu en 2019, et il reste politiquement instable. Les chiffres sont à jour de 2026, mais la vérification annuelle reste obligatoire.

::: cle Le principe, et pourquoi il vous vise
Depuis 2016, la protection universelle maladie (PUMa) garantit à tout résident stable la prise en charge de ses frais de santé, y compris s'il ne cotise plus nulle part. La contrepartie est la CSM. Elle se déclenche quand deux conditions sont réunies. Vos revenus d'activité sont inférieurs à un seuil, 20 % du plafond de la Sécurité sociale, soit ~9 600 €/an. Et vos revenus du capital dépassent un autre seuil, 50 % du même plafond, soit ~24 000 €/an. Vous cotisez alors ~6,5 % sur ces revenus du capital au-delà de la franchise. Le salarié, l'indépendant, le retraité pensionné et le chômeur indemnisé n'y sont pas soumis. En pratique, le dispositif ne frappe qu'une population, le rentier d'avant la pension. C'est-à-dire vous, pendant la phase de pont ([[horizon-et-esperance-de-vie]]).
:::

## La formule, les seuils, la facture

**La formule (état 2019-2026).** La cotisation annuelle suit la formule du code de la Sécurité sociale :

> CSM = 6,5 % × (revenus du capital − 0,5 × PASS) × (1 − revenus d'activité / (0,2 × PASS))

Le PASS (plafond annuel de la Sécurité sociale) vaut 48 060 € en 2026 et se revalorise chaque année. L'assiette des revenus du capital est plafonnée à 8 PASS, soit ~384 000 €. La cotisation n'est due que si les deux conditions sont réunies, activité < 20 % du PASS et capital > 50 % du PASS. Le dernier facteur est la décote linéaire. Plus vos revenus d'activité approchent du seuil, plus la cotisation fond, jusqu'à zéro au seuil. C'est l'interrupteur dont tout découle.

::: figure puma-interrupteur
Le montant de la cotisation pour chaque combinaison de revenus d'activité et de revenus du capital réalisés, au PASS 2026 de 48 060 €, selon la formule de l'article D. 380-1 du code de la Sécurité sociale. On y lit deux zones blanches, sous la franchise de 0,5 PASS de capital et au-delà de 0,2 PASS d'activité. Entre les deux, la cotisation monte régulièrement avec le capital réalisé.
:::

**Les ordres de grandeur.** Un rentier sans activité avec 60 000 € de revenus du capital imposables paie une CSM ≈ 6,5 % × (60 000 − 24 030) ≈ 2 340 €/an. Avec 120 000 €, lors d'une grosse année de plus-values, elle grimpe à ≈ 6 240 €. Le maximum théorique, assiette plafonnée, atteint ~23 400 €/an. L'appel arrive avec un décalage. La CSM de l'année N est calculée sur la déclaration de N, puis appelée fin N+1. C'est le fameux courrier-surprise. Budgétez-la l'année où vous générez l'assiette, pas l'année où elle tombe.

**L'assiette : ce qui compte.** Les revenus du capital au sens du dispositif recouvrent pour l'essentiel trois postes. Les revenus fonciers, les revenus de capitaux mobiliers (dividendes, intérêts, dont les gains des rachats d'assurance-vie) et les plus-values imposables. En gros, ce que votre déclaration fait apparaître comme revenus du patrimoine. De là viennent les leviers structurels. Ce qui n'apparaît pas n'entre pas dans l'assiette, à commencer par les gains latents, que l'enveloppe capitalisante garde hors de la déclaration ([[etf-ucits-europeens]]). De là viennent aussi des zones grises documentées, comme le traitement fin des retraits de PEA, de certaines PV et l'articulation exacte avec le revenu fiscal de référence (RFR). La doctrine URSSAF a évolué sur ces points. Cet article s'interdit d'y trancher, car c'est au simulateur URSSAF, et au besoin à un rescrit, de trancher votre cas, chaque année ([[revue-annuelle]]).

## Qui est touché, qui échappe : la cartographie

**Touchés** : le rentier précoce sans aucune activité, pendant toute la phase de pont. C'est la population cible, presque nommément. **Exemptés de droit**, plusieurs profils. Les titulaires de revenus d'activité au-dessus du seuil. Les titulaires d'une pension de retraite, de base ou complémentaire, même modeste, car la liquidation éteint la CSM à vie (un argument de plus au dossier des trimestres, [[retraite-legale]]). Les chômeurs indemnisés, ce qui fait des allocations post-rupture conventionnelle un sursis de CSM en début de FIRE. Enfin les situations de couple où les conditions ne sont pas réunies, l'appréciation tenant compte de la situation du foyer. L'activité suffisante d'un conjoint change la donne. Les modalités exactes de calcul en couple ont fait l'objet d'ajustements, donc vérifiez votre cas au simulateur, sans rien supposer.

Notez le zigzag. Le dispositif initial de 2016 taxait 8 % au-delà d'une franchise de 25 % du PASS seulement, avec un seuil d'activité à 10 % du PASS et sans décote. Il a été contesté, contentieux à l'appui, puis recalibré en 2019 vers la formule actuelle, plus douce. L'histoire enseigne que les paramètres bougent, dans les deux sens. La CSM est un poste de veille, pas une constante du plan. C'est la « saisie réglementaire douce » de [[hyperinflation-et-extremes]], en version ordinaire.

::: science Les quatre parades, classées
**1. L'interrupteur d'activité (radical).** Environ 9 600 €/an de revenus d'activité annulent la CSM intégralement par la décote (micro-entreprise, salariat partiel, quelques missions). Pour un rentier à 3 000 €+ de CSM, le « taux de rémunération » implicite de cette petite activité est spectaculaire. Elle rapporte son revenu plus l'économie de cotisation. C'est l'argument fiscal du Barista FIRE ([[retour-au-travail]]), et souvent la réponse la plus simple pour qui garde de toute façon un pied professionnel. Attention à la substance. L'activité doit être réelle, avec revenus déclarés et cotisations d'activité payées, car le montage fictif est un redressement en attente.
**2. Le pilotage d'assiette (structurel).** La CSM taxe les revenus réalisés de l'année. Les gestes du chapitre précédent la pilotent donc aussi. Capitaliser, car les ETF capitalisants ne créent aucun revenu tant qu'on ne vend pas. Doser les rachats AV, car seule la part de gain entre en compte. Étaler les grosses plus-values sur plusieurs années, puisque l'assiette a une franchise annuelle de ~24 k€ (trois années à 40 k€ de revenus du capital paient moins qu'une année à 120 k€). Arbitrer enfin l'ordre des robinets en intégrant la CSM au taux mixte ([[enveloppes-francaises]], [[flat-tax-et-imposition]]), où le lissage de taux devient un lissage taux + CSM. La tension est à arbitrer honnêtement. Le lissage fiscal pousse à réaliser plus dans les années creuses, la CSM pousse à réaliser moins. L'optimum conjoint réalise jusqu'à la franchise CSM et le haut des tranches basses, pas au-delà.
**3. La liquidation d'une pension (terminale).** La CSM meurt avec la première pension. Pour les fins de pont, on peut comparer l'avance d'un an de la liquidation, décote de pension contre années de CSM. C'est un calcul de cas, à faire chiffré.
**4. La provision (le minimum vital).** Si vous ne faites rien d'autre, budgétez-la. Comptez ~6,5 % des revenus du capital réalisés au-delà de la franchise, dans les dépenses du plan. C'est l'option « je paie pour la simplicité », légitime aussi ([[combien-il-vous-faut]], dont la majoration de friction de 10-15 % du budget la contenait déjà).
:::

## L'intégration au plan, et au simulateur

**Dans le dimensionnement.** La CSM est une friction sur les revenus réalisés de la phase de pont. Pour un plan organisé, avec enveloppes capitalisantes et rachats dosés, elle ressort typiquement à 0,5-2 k€/an. Pour un plan naïf, avec gros CTO distribuant et grosses ventes annuelles, elle grimpe à 3-8 k€/an. L'écart est un argument d'organisation de plus. **Dans un simulateur.** Le curseur fiscal unique l'approxime ([[utiliser-la-page-fire]]). Son défaut de 32,8 % correspond peu ou prou à « PFU + une CSM moyenne ». Si vous avez calibré votre taux mixte ([[flat-tax-et-imposition]]), ajoutez-y l'équivalent CSM de votre structure, soit la part des revenus réalisés dans vos flux × 6,5 %, au-delà de la franchise. Souvenez-vous enfin que la CSM s'éteint à la pension. Le taux mixte de la phase adossée est plus bas. C'est une raison de plus pour laquelle le plan français type s'améliore tout seul avec l'âge ([[horizon-et-esperance-de-vie]]).

**Dans la gouvernance.** Gardez trois réflexes. Le simulateur URSSAF chaque année de pont, car l'assiette et la doctrine bougent (dix minutes suffisent). La provision comptable l'année de réalisation, car l'appel arrive en N+1 et le plan de trésorerie doit le savoir ([[revue-annuelle]]). La veille législative enfin, car le dispositif est un marqueur politique. Chaque loi de financement de la Sécurité sociale peut le retoucher. C'est le poste fiscal du plan FIRE à surveiller de près.

::: exemple Trois versions du même pont, CSM comprise
Léa et Sam vivent une phase de pont de 12 ans, avec un besoin de 50 000 €/an et un patrimoine organisé ([[enveloppes-francaises]]). La version A, naïve, place tout en CTO distribuant, avec de grosses ventes. Les revenus du capital réalisés atteignent ~70 k€/an, et la CSM y ressort à ~3 000 €/an × 12 ans ≈ 36 000 € de pont. La version B, organisée, capitalise partout, avec rachats AV sous abattement, PEA et ventes CTO dosées. Les revenus réalisés tombent à ~32 k€/an, et la CSM n'est plus que ~520 €/an ≈ 6 200 € de pont. L'organisation vaut donc ~30 000 €. La version C, l'interrupteur, voit Sam garder une micro-activité de conseil à ~10 k€/an qui lui plaît. La CSM tombe à zéro. Les 10 k€ réduisent d'autant les retraits, et la meilleure assurance anti-séquence du livre se paie elle-même ([[revenus-complementaires]], [[sequence-des-rendements]]). Le choix entre B et C n'est pas fiscal. Il est de vie. Mais il se fait en connaissant les trois chiffres.
:::

## L'essentiel à retenir

- La CSM (« taxe PUMa ») vise exactement le rentier d'avant la pension. Elle prélève ~6,5 % des revenus du capital au-delà d'une franchise (0,5 PASS ≈ 24 k€), si les revenus d'activité restent sous 0,2 PASS (≈ 9,6 k€). Assiette plafonnée, appel en N+1, dispositif créé en 2016 puis refondu en 2019, d'où une veille annuelle obligatoire.
- Salariés, indépendants, chômeurs indemnisés et pensionnés y échappent. Elle ne concerne que la phase de pont. Et elle s'éteint à vie à la première pension liquidée.
- Quatre parades, par efficacité. L'interrupteur d'activité (~9,6 k€ réels d'activité = zéro CSM, l'argument fiscal du Barista), le pilotage d'assiette (capitaliser, doser les rachats, étaler les PV, l'optimum conjoint avec le lissage fiscal, qui réalise jusqu'à la franchise et pas au-delà), la liquidation calculée, la provision budgétée.
- Ordre de grandeur, 0,5-2 k€/an pour un plan organisé, 3-8 k€ pour un plan naïf, soit 0,2 à 0,4 point de taux de retrait dans ce dernier cas. À intégrer au taux mixte (le défaut 32,8 % l'approxime), en le baissant pour la phase adossée.
- Réflexes de gouvernance, simulateur URSSAF chaque année, provision l'année de réalisation, veille sur chaque loi de financement. C'est le prélèvement instable du plan français.

---

## Pour aller plus loin

- [urssaf.fr](https://www.urssaf.fr), la page cotisation subsidiaire maladie et son simulateur. C'est la source opérationnelle, à consulter chaque année de pont.
- L'article L. 380-2 du code de la Sécurité sociale et le décret de 2019, les textes pour les amateurs de sources primaires.
- Les analyses des contentieux 2016-2019 (doctrine et jurisprudence), l'histoire du zigzag, instructive sur l'instabilité du dispositif.
- Dans ce livre : [[flat-tax-et-imposition]] (le taux mixte où l'intégrer), [[enveloppes-francaises]] (le pilotage d'assiette), [[retour-au-travail]] (l'interrupteur d'activité comme choix de vie), [[retraite-legale]] (l'extinction par la pension).
