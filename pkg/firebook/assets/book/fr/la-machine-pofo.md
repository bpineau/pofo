# Sous le capot : comment pofo calcule ce livre

Ce livre cite la page FIRE de pofo à chaque chapitre. Le dernier article lui rend la politesse : voici, sans boîte noire, ce que la machine calcule exactement. Nous verrons d'où viennent ses données (les historiques reconstruits, le panel du siècle, le CAPE vivant). Puis comment le modèle central se fabrique, du portefeuille de vos lignes aux paramètres ajustés puis mélangés. Ce que chaque colonne du tableau tire des données, dans le récapitulatif technique des six lentilles. Comment le noyau de simulation déroule un plan mois par mois : règles de dépense, fiscalité qui majore les ventes, buffer et ses seuils, flux. Et ce que chaque section de la page calcule. Dans l'esprit de tout le livre, nous assumerons enfin les **limites** de l'ensemble. Un outil dont on ne connaît pas les simplifications est un oracle, et ce livre a assez écrit contre les oracles ([[pieges-des-simulateurs]], [[monte-carlo-forces-faiblesses]]). Ce chapitre est aussi le plus « versionné » du livre, car la machine évolue. Le volet « How this machine works » de la page elle-même fait foi pour l'état exact du moment ([[utiliser-la-page-fire]]).

::: cle La philosophie de conception, en trois choix
**Un** : tout en **réel**. Les séries sont déflatées par l'IPCH et les retraits fixés en pouvoir d'achat constant. L'inflation moyenne entre ainsi dans la machine par construction, et le risque d'un épisode inflationniste dans les colonnes de régimes ([[inflation-et-taux-de-retrait]]). **Deux** : le **faisceau** plutôt que le verdict. Quatre familles de modèles tournent côte à côte, dont deux que vos curseurs ne peuvent pas influencer ; les données restent juges ([[historique-vs-parametrique]]). **Trois** : le **plan réel** plutôt que la caricature. On simule les vraies règles de dépense, la vraie fiscalité de vente, le vrai buffer. La complexité du plan est gratuite en simulation, alors elle est servie ([[rendre-monte-carlo-pertinent]]).
:::

## Les données : ce que la machine sait du monde

**Les historiques longs de vos fonds.** En mode portefeuille, pofo étend chaque ligne au-delà de son historique coté. Les extensions SIM raccordent des backcasts documentés : indices longs, composites reconstruits, chaque recette étant calibrée et versionnée. Les managed futures, par exemple, sont un backcast TSMOM à volatilité contrôlée ([[managed-futures]]). Puis tout est converti en euros et **déflaté** par l'IPCH (`^HICP-FR`). Le panel final est une matrice de rendements **réels mensuels** de vos lignes, sur plusieurs décennies.

**Le siècle des seize pays.** Le modèle broad-sample embarque le panel académique Jorda-Schularick-Taylor (rendements réels annuels actions/obligations de 16 pays développés, 1870-2020) : la matière du rejeu « destins nationaux » ([[anarkulova-cederburg]]).

**Le CAPE vivant.** La série de Shiller (1871-aujourd'hui) est complétée en continu. Le §00 de la page et l'ancre CAPE s'en nourrissent ([[valorisations-et-cape]]).

## Le pipeline du modèle central

De vos lignes au modèle Student-t central, le chemin passe par cinq étapes ([[rendre-monte-carlo-pertinent]] justifie chacune d'elles).

1. **Le panel mensuel réel**, aux poids **courants** de vos curseurs de lignes.
2. **L'ajustement.** On calcule μ (la moyenne réalisée annualisée), σ (la dispersion mensuelle mise à l'échelle annuelle) et df (l'inverse du kurtosis mensuel, soit les queues de **vos** fonds, [[queues-epaisses]]). Le σ suit les ratios de variance de long horizon, ce qui le place plus bas que la volatilité « une année » des brochures.
3. **Le blending.** Les paramètres ajustés sont mélangés vers le prior mondial prudent (μ 4,5 %, σ 13, df 4), à proportion de ce que l'horizon excède l'historique, sans dépasser 50/50. Votre portefeuille atteint le central par ses statistiques, jamais par sa séquence.
4. **Les ancres optionnelles.** Le CAPE remplace la moyenne par l'estimation qu'impliquent les valorisations ; le prior broad-sample, lui, écrase les trois curseurs.
5. **Le tirage.** Des mois Student-t indépendants sont composés en années, et le frein de la volatilité (volatility drag) émerge par construction ([[rendements-arithmetiques-geometriques]]).

**Les six colonnes, mécanique exacte** ([[historique-vs-parametrique]] pour la grille de lecture) :

- **Historical windows** : chaque mois de départ possible de votre panel ouvre une fenêtre rejouée telle quelle (refus honnête si l'historique est plus court que l'horizon).
- **Block bootstrap** : bootstrap stationnaire de votre panel (blocs de longueur aléatoire, 24 mois en moyenne).
- **Student-t** : le central du pipeline ci-dessus.
- **Sequence stress** : chaîne de Markov à deux états calibrée (sticky bears, ~19 % d'années de marché baissier en épisodes de ~3 ans, volatilité ×1,5 dedans, **moyenne préservée** ; l'écart avec le central isole le prix de l'ordre, [[sequence-des-rendements]]).
- **Broad sample** : bootstrap par blocs **par pays** du panel JST, en 60/40 domestique (le portefeuille et les curseurs sont ignorés, c'est voulu).
- **Lost decade** : le régime de marché baissier long à la japonaise, moyenne délibérément dégradée. C'est le crash-test du modèle.

## Le noyau : un plan déroulé mois par mois

Chaque trajectoire simulée est une comptabilité exacte ([[monte-carlo-forces-faiblesses]]). Le noyau mensuel déroule, dans l'ordre, quatre gestes pour chaque mois.

1. **Le rendement** du mois est appliqué aux poches.
2. **Le besoin** du mois est fixé selon la règle active, puis diminué des **flux** entrants du moment (la pension à partir de son année, le side income jusqu'à la sienne, la rente si une part a été annuitisée). Chaque règle a sa formule : fixe indexé prend le douzième du besoin réel ; flex applique la coupe si le drawdown dépasse le seuil ; guardrails joue les ±10 % aux bornes du corridor, avec leur plancher ; VPW et pourcentage prennent une part du portefeuille courant ; ABW re-cote l'annuité sur la richesse plus les pensions actualisées ; bounded suit le corridor Vanguard ; le cliquet et le sourire modulent l'ensemble ([[panorama-strategies-retrait]]).
3. **La source** du retrait est choisie. Le buffer paie d'abord si le drawdown dépasse son seuil (~10 % par défaut), les poches sinon. Le buffer se recharge ensuite progressivement, sans dépasser son plafond, quand le calme revient, et il s'éteint à l'année d'arrêt ([[recharger-ou-pas]], la mécanique décrite là-bas est littéralement celle du code).
4. **La fiscalité** est appliquée. Chaque vente est **majorée** pour livrer le net : le taux du curseur porte sur la part de gain. Cette part démarre basse, puis dérive vers le haut à mesure que les plus-values latentes composent, ce qui reproduit le comportement du PMP français ([[flat-tax-et-imposition]]). Les enveloppes optionnelles (PEA, AV) affinent la répartition.

La ruine est constatée à l'épuisement, avec sa date. Tout le reste est enregistré : richesse, dépenses servies, temps sous l'eau. La machine tire 4 000 trajectoires par défaut, et le curseur monte à 10 000.

## Les sections : qui calcule quoi

Voici le tour rapide, chaque section renvoyant à sa lecture ([[utiliser-la-page-fire]]).

- **§00** : le CAPE du jour sur son siècle.
- **§01** : les cônes des quatre lentilles (percentiles par date, huit chemins à rangs réguliers dont les rouges comptent la ruine, axe écrêté à 10×, [[lire-un-fan-chart]]).
- **§02** : votre plan rejoué aux dates célèbres sur les données longues (1929, 1966, 2000, Japon 1990).
- **§03** : la ruine conditionnée au rendement de la première décennie de chaque trajectoire (la séquence rendue visible).
- **§04** : la distribution du niveau de vie **servi** et son financement (le juge des règles flexibles).
- **§05** : la mortalité d'un couple français croisée à la ruine (« alive, broke or gone »), la distribution des legs, et les **causes** d'échec (krach précoce, érosion ou longévité, le diagnostic qui choisit la parade).
- **§06** : la frontière des règles (ruine × variabilité du vécu) et les sensibilités des leviers.
- **§07** : l'arbitrage du buffer (balayage 0-10 ans) et l'histogramme des années sous l'eau.
- **§08-09** : le détail chiffré au modèle sélectionné, puis le **solveur**, ces mouvements équivalents qui ramènent la ruine sous votre seuil (le menu anti-OMY, [[une-annee-de-plus]]).

::: attention Les limites assumées, en face
Voici la liste que le volet « caveats » de la page affiche, commentée point par point.

- Le **portefeuille est agrégé** avant simulation. Il n'y a pas de rotation fine entre classes pendant les trajectoires ; la grille des régimes sert à la conception, et la machine teste ensuite la composition retenue ([[regimes-de-marche]]).
- Le **central est i.i.d.** Les colonnes stress et broad-sample existent précisément pour cette limite.
- La **fiscalité** est un taux mixte global, à calibrer soi-même, sans moteur fiscal par enveloppe ([[flat-tax-et-imposition]]).
- Les **historiques** de vos fonds sont une fenêtre favorable. Le blending et les avertissements le disent, mais aucun correctif ne rend une fenêtre courte exhaustive.
- La **mortalité** est ignorée par défaut, et le §05 la réintroduit.
- Le **hors-champ** habituel demeure : divorce, dépendance, queues politiques, renvoyés aux marges et aux chapitres dédiés ([[hyperinflation-et-extremes]]).

De tout le livre découle une règle de lecture : privilégier les colonnes pessimistes pour décider, pousser l'horizon au-delà de l'espérance de vie, et traiter les décimales en ordinal ([[ruine-et-probabilites]]).
:::

## Vérifier par soi-même

La machine s'audite de trois façons. D'abord, le moteur est du code Go inspectable : dans le dépôt pofo, chaque modèle, chaque règle et chaque recette de backcast est lisible et testée, et les goldens comparent les calculs à des références externes gelées. Ensuite, les **grands résultats** du livre s'y reproduisent en quelques minutes. Le SAFEMAX à ~4 % sur 30 ans sort des fenêtres historiques ; la falaise de Trinity apparaît en balayant le taux ; le prix de la séquence se lit en comparant le central et le stress ; l'effet pension se voit en basculant son curseur. Les chapitres correspondants donnent chaque recette. Enfin, les **désaccords** avec d'autres outils s'expliquent par la grille d'audit : queues, mémoire, frictions, échantillon ([[pieges-des-simulateurs]]). Il coche les cases que la plupart des simulateurs grand public laissent vides, et affiche celles qu'il laisse vides lui-même. C'est le critère final proposé au lecteur pour **tout** outil, celui-ci compris. La question n'est pas « a-t-il raison ? », car personne n'a raison sur 45 ans, mais « sait-on ce qu'il fait, et le dit-il ? ».

## L'essentiel à retenir

- Trois choix de conception : tout en réel (IPCH), le faisceau de modèles (deux colonnes insensibles aux curseurs, les données restent juges), le plan réel simulé (règles, fiscalité, buffer, flux, mois par mois).
- Le central se fabrique en cinq étapes : panel mensuel réel de **vos** lignes → ajustement (μ, σ long-horizon, df des queues) → blending vers le prior mondial (plafonné 50/50) → ancres optionnelles (CAPE, broad-sample) → tirages Student-t mensuels composés.
- Les six colonnes : vos fenêtres rejouées, votre panel bootstrappé, le central, le stress à moyenne préservée (le prix de l'ordre), le siècle JST en 60/40 domestique, la décennie perdue. Chaque désaccord entre elles est un diagnostic ([[historique-vs-parametrique]]).
- Le noyau majore chaque vente de sa fiscalité (part de gain croissante), sert le buffer à ses seuils, encaisse pensions et side income à leurs dates : la mécanique de ce livre est littéralement celle du code.
- Les limites sont affichées (agrégation, i.i.d. central, taux fiscal global, fenêtres courtes, hors-champ). Et c'est le critère proposé pour tout outil : non pas avoir raison, mais dire ce qu'on fait.

---

## Pour aller plus loin

- Le volet « How this machine works » et le volet « Method & honest caveats » de la page FIRE : l'état exact et à jour de la machine.
- Le mode d'emploi : [[utiliser-la-page-fire]]. Les fondements de chaque choix : [[rendre-monte-carlo-pertinent]], [[historique-vs-parametrique]], [[queues-epaisses]].
- Jorda-Schularick-Taylor (le panel du broad-sample) et le site de Shiller (le CAPE) : les données publiques que la machine embarque ([[bibliotheque]]).
