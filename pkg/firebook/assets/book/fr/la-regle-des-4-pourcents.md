# La règle des 4 % en dix minutes

C'est la règle la plus célèbre de toute la finance personnelle : retirez 4 % de votre capital la première année de retraite, réindexez ce montant sur l'inflation chaque année, et votre portefeuille tiendra 30 ans. Elle est si célèbre qu'on l'appelle « la règle », tout court.

Cette page explique exactement ce qu'elle dit (c'est plus subtil que la version de comptoir), d'où viennent ses chiffres, ce qu'elle suppose sans le dire, et pourquoi l'état de l'art la traite aujourd'hui comme un excellent point de départ et un mauvais point d'arrivée. À la fin, vous saurez l'utiliser pour ce qu'elle vaut : un ordre de grandeur, pas un plan. Et pour comprendre pourquoi, mathématiquement, ce chiffre-là tient (et ce qui le ferait bouger), la cascade complète est démontée dans [[les-maths-du-4-pourcent]].

::: cle Ce que la règle dit exactement
Année 1 : retirez 4 % du capital initial (40 000 € pour 1 000 000 €). Années suivantes : retirez le **même** montant ajusté de l'inflation (41 200 € si l'inflation a été de 3 %), quel que soit ce que fait le portefeuille. Historiquement, sur toutes les fenêtres de 30 ans américaines depuis 1926, un portefeuille 50 à 75 % actions n'a jamais été épuisé par ce régime. « Sûr » signifie ici : n'a jamais échoué dans le passé américain observé. Rien de plus.
:::

## La mécanique, pas à pas

Prenons Camille, qui part en retraite avec 1 000 000 € investis en 60 % actions mondiales, 40 % obligations.

1. **Année 1** : elle retire 40 000 € (4 % de 1 000 000).
2. **Année 2** : l'inflation a été de 2,5 %. Elle retire 40 000 × 1,025 = 41 000 €. Peu importe que son portefeuille ait gagné 15 % ou perdu 20 % : le retrait est le même.
3. **Année 3 et suivantes** : même logique, le retrait de l'an passé ajusté de l'inflation.

Trois propriétés découlent immédiatement de cette mécanique, et elles expliquent tout le reste du sujet.

**Le pouvoir d'achat est constant.** C'est la grande qualité de la règle : votre train de vie ne dépend jamais de l'humeur des marchés. C'est un contrat de rente que vous passez avec votre propre portefeuille.

**Le taux de retrait effectif, lui, flotte.** Si le portefeuille de Camille tombe à 700 000 € après un krach en année 2, ses 41 000 € représentent désormais 5,9 % du capital. La règle ne s'en soucie pas ; c'est précisément là que le risque se loge.

**Le « 4 % » ne compte qu'une fois.** Le taux ne s'applique qu'au capital initial, au jour du départ. D'où un paradoxe connu : deux voisins identiques, l'un parti en 2021 avec 1 M€ (retrait 40 k€), l'autre parti en 2022 après un krach avec 800 k€ (retrait 32 k€), retirent des montants différents alors qu'ils ont désormais le même portefeuille. Ce paradoxe n'est pas un détail. Il révèle que la règle est une simplification d'un objet plus profond (le taux de retrait dépend des valorisations de départ, [[valorisations-et-cape]]).

## D'où viennent les chiffres

La règle a deux actes fondateurs, détaillés dans [[etude-trinity]].

**Bengen, 1994.** William Bengen rejoue toutes les retraites américaines possibles depuis 1926 : départ en 1926, en 1927, en 1928... Pour chaque « millésime », il calcule combien d'années un portefeuille 50/50 actions/obligations aurait tenu sous un retrait indexé donné. Résultat : le pire millésime de l'histoire (départ en 1966, dans les dents du marché plat et de l'inflation des années 1970) supportait un retrait initial d'environ 4,15 % sur 30 ans. Bengen appelle ce plancher SAFEMAX. Le chiffre rond « 4 % » vient de là. C'est le taux du **pire** cas historique américain, pas une moyenne (la moyenne des millésimes supporte plus de 6 %).

**Trinity, 1998.** Trois professeurs de la Trinity University (Cooley, Hubbard, Walz) transforment l'approche en grille de probabilités : pour chaque taux de retrait, allocation et horizon, quel pourcentage des fenêtres historiques a survécu ? La cellule restée célèbre : 4 %, portefeuille 50/50, 30 ans, 95 à 100 % de succès selon les mises à jour. C'est de cette étude que vient l'idée de « probabilité de succès » qui structure encore tous les simulateurs ([[ruine-et-probabilites]]).

::: encart La règle inversée : le multiple de 25
« Retirer 4 % » se retourne en « accumuler 25 fois ses dépenses » (1/0,04 = 25). C'est la même règle vue du côté accumulation, et c'est son usage le plus utile : transformer un budget en cible de capital. 3 % correspond à 33 fois, 3,5 % à environ 29 fois. Le détail du passage budget → cible, fiscalité comprise, est dans [[combien-il-vous-faut]].
:::

## Ce que la règle suppose sans le dire

La force marketing du « 4 % » a fait oublier la liste, pourtant longue, de ses hypothèses. Chacune mérite d'être confrontée à votre situation.

| Hypothèse implicite | La réalité de votre plan |
|---|---|
| Horizon de 30 ans | Un départ à 40 ans, c'est 45 à 55 ans ([[horizon-et-esperance-de-vie]]) |
| Actions et obligations **américaines**, 1926-1995 | Le meilleur marché du siècle, sur son meilleur siècle ([[anarkulova-cederburg]]) |
| Aucun frais | Un ETF coûte 0,1 à 0,4 %/an, un contrat d'assurance-vie parfois 1 % de plus ([[etf-ucits-europeens]]) |
| Aucun impôt sur les retraits | PFU, prélèvements sociaux, PUMa : comptez-les dans les dépenses ([[flat-tax-et-imposition]], [[taxe-puma]]) |
| Dépenses parfaitement rigides, indexées inflation | Les vraies dépenses varient, et c'est une marge exploitable ([[depenses-en-retraite]]) |
| Aucun autre revenu, jamais | Retraite légale, activités, héritages existent ([[revenus-complementaires]], [[retraite-legale]]) |
| Le retraité applique la règle mécaniquement 30 ans | Personne ne regarde son portefeuille fondre sans réagir ([[psychologie-du-retrait]]) |
| Succès = il reste 1 € au dernier jour | Finir à 82 ans avec 3 000 € n'est un « succès » que pour le simulateur |

Aucune de ces hypothèses n'invalide la règle comme ordre de grandeur. Mais leur somme explique pourquoi le chiffre qui sort d'une analyse sérieuse de **votre** situation peut s'écarter de 4 % dans les deux sens, et de beaucoup.

## Ce qu'en dit l'état de l'art

Trente ans de recherche ont précisé le tableau. Les résultats convergent sur quatre points.

**1. Pour 30 ans aux États-Unis, la règle a bien tenu.** Y compris à travers 2000-2009, la pire décennie boursière moderne : le retraité de janvier 2000 (deux krachs de 50 % dans les dix premières années) était encore solvable en 2024 avec la règle des 4 %. Le cadre historique n'était pas absurde.

**2. Pour un horizon long, 4 % est trop haut en régime rigide.** C'est le résultat central de la série d'Early Retirement Now ([[serie-ern]]) : sur 50 à 60 ans, le taux qui aurait survécu à tous les millésimes américains tombe vers 3,25 à 3,5 %. La probabilité d'échec du 4 % rigide sur 50 ans, mesurée sur l'histoire américaine, est de l'ordre de 10 à 20 % selon l'allocation, trop pour un plan de vie.

**3. Hors des États-Unis, c'est pire.** Sur l'échantillon mondial (le « broad sample » de 16 pays développés depuis 1870, [[anarkulova-cederburg]]), le taux « sûr » d'un portefeuille 60/40 domestique est nettement sous 4 % : la France, le Japon, l'Allemagne ou l'Italie du XXe siècle ont infligé aux rentiers des séquences que l'histoire américaine ne contient tout simplement pas. Un investisseur mondialisé d'aujourd'hui se situe quelque part entre ces deux mondes.

**4. Les valorisations de départ déplacent le taux.** Partir quand les marchés sont chers (CAPE élevé) a historiquement toujours donné les pires millésimes. Les règles modernes conditionnent donc le taux initial aux valorisations ([[valorisations-et-cape]], [[regles-cape]]). C'est l'une des améliorations les mieux établies.

::: science Où la recherche situe le « vrai » chiffre aujourd'hui
Pour une retraite précoce (45 ans et plus d'horizon), portefeuille mondial diversifié, sans flexibilité ni revenu complémentaire, la littérature converge vers 3,0 à 3,5 % en régime rigide (ERN, 3,25 % ; Morningstar 2024, sur 30 ans avec rendements prospectifs, 3,7 % ; Anarkulova-Cederburg, échantillon mondial, plutôt 2,3 à 2,7 % pour les plus pessimistes, une borne basse discutée). Chaque protection ajoutée (flexibilité des dépenses [[flexibilite-realite]], revenus partiels [[retour-au-travail]], retraite légale future [[retraite-legale]], guardrails [[guardrails-morningstar]]) remonte ce chiffre, parfois au-delà de 4 %. Le taux sûr n'est pas une constante de la nature. C'est la sortie d'un modèle, fonction de vos hypothèses et de vos marges.
:::

## Alors, la garder ou la jeter ?

La garder, mais à sa place. La règle des 4 % excelle dans trois rôles et échoue dans un quatrième.

**Excellente comme unité de mesure.** « Ce niveau de dépenses exige 25 fois plus de capital » est le réflexe mental le plus utile du sujet. Une dépense récurrente de 100 €/mois « coûte » 30 000 à 36 000 € de capital : ce simple calcul change la façon dont on arbitre un abonnement, une voiture, un déménagement.

**Excellente comme point de départ du dimensionnement.** Viser 25 fois ses dépenses, puis affiner avec un vrai modèle ([[utiliser-la-page-fire]]) et sa vraie situation. C'est la bonne séquence. L'erreur n'est pas de commencer par 4 %, c'est de s'y arrêter.

**Excellente comme borne de récit.** Quand un vendeur promet « 8 % de rente sans risque », la règle vous dit instantanément que c'est deux fois le taux que le meilleur marché de l'histoire a soutenu en rigide sur 30 ans. Réflexe salvateur.

**Mauvaise comme stratégie de retrait effective.** Personne ne devrait exécuter mécaniquement un retrait indexé aveugle pendant 40 ans. Non parce que la règle est fausse, mais parce qu'elle ignore l'information qui arrive : marchés, santé, dépenses réelles, valorisations. Toute la partie « stratégies de retrait » de ce livre ([[panorama-strategies-retrait]]) traite de ce qu'on met à la place : des règles qui écoutent le portefeuille ([[guyton-klinger]], [[guardrails-morningstar]], [[amortissement-abw]]) et rendent le même capital capable de financer davantage, ou le même train de vie beaucoup plus sûr.

::: attention Le contresens le plus fréquent
« 4 %, donc je peux retirer 4 % du portefeuille chaque année. » Non. Ça, c'est la stratégie du pourcentage fixe ([[pourcentage-fixe]]), qui ne peut jamais ruiner mais fait violemment fluctuer le train de vie. La règle de Bengen indexe le **montant** initial sur l'inflation ; le pourcentage du capital courant, lui, dérive. Les deux stratégies n'ont ni les mêmes risques ni le même confort ; les confondre fausse toute conversation sur le sujet.
:::

::: exemple La règle des 4 % à l'épreuve d'un vrai cas
Reprenons Camille : 1 M€, 60/40 mondial, départ à 45 ans, 40 000 €/an indexés. Ce que disent les modèles (page FIRE, [[utiliser-la-page-fire]]) sur 45 ans : le modèle central calibré donne typiquement une ruine de l'ordre de 10 à 15 %, le rejeu de l'échantillon mondial davantage, les fenêtres historiques du portefeuille moins. À 3,4 % (34 000 €/an), la ruine centrale passe sous 5 % ; à 4 % avec 800 € /mois de retraite légale à partir de 67 ans, elle repasse aussi sous 5 %. Voilà la règle bien utilisée : un point de départ, trois leviers testés, une décision informée.
:::

## L'essentiel à retenir

- La règle : 4 % du capital initial, puis le même montant indexé sur l'inflation ; historiquement imbattu sur 30 ans aux États-Unis.
- Le « 4 » vient du pire millésime américain (1966), pas d'une moyenne ; c'était déjà un chiffre prudent... dans son cadre.
- Hors de son cadre (horizon long, monde entier, frais, impôts, valorisations élevées), le taux rigide équivalent est plutôt 3 à 3,5 %.
- Inversée en « multiple de 25 », c'est le meilleur réflexe mental du sujet.
- Comme stratégie effective, elle est dominée par les règles adaptatives : lisez [[panorama-strategies-retrait]] avant de graver 4 % dans votre plan.

---

## Pour aller plus loin

- William Bengen, « Determining Withdrawal Rates Using Historical Data », *Journal of Financial Planning*, 1994.
- Cooley, Hubbard & Walz, « Retirement Savings: Choosing a Withdrawal Rate That Is Sustainable », *AAII Journal*, 1998 (l'étude Trinity, [[etude-trinity]]).
- Early Retirement Now, SWR Series volet 1 et volet 26 (« Ten Things the Makers of the 4% Rule Don't Want You to Know ») : la critique moderne la plus complète ([[serie-ern]]).
- Morningstar, *The State of Retirement Income* (annuel) : le taux recommandé recalculé chaque année avec des rendements prospectifs ([[guardrails-morningstar]]).
- Bengen lui-même, interviews récentes. Il considère aujourd'hui 4 % comme trop conservateur pour 30 ans avec un portefeuille plus diversifié... et ne s'applique qu'au cadre américain sur 30 ans. Les deux mouvements (le sien vers le haut, celui de la recherche « monde + horizon long » vers le bas) illustrent bien que le chiffre dépend du cadre.
