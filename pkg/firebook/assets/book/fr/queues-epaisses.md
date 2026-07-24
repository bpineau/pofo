# Queues épaisses, crises et Student-t

Le 19 octobre 1987, le marché américain a perdu 20 % en une séance. Sous la loi normale calibrée sur la volatilité de l'époque, cet événement était à plus de vingt écarts-types : une probabilité si faible que l'univers entier n'aurait pas dû suffire à le produire une fois.

Il s'est pourtant produit. Et 2008, 2020, 1929, 1973 racontent la même histoire à l'échelle annuelle : **les marchés font des extrêmes bien plus souvent que la courbe en cloche ne le permet**. C'est ce qu'on appelle les queues épaisses (fat tails). Pour un rentier, ce n'est pas une curiosité statistique. La ruine d'un plan de retraite vit précisément dans les queues, là où la loi normale ne regarde pas.

Cette page déroule le sujet. D'abord le phénomène : d'où viennent les queues et comment on les mesure. Ensuite l'outil standard pour les modéliser, la distribution de Student-t et son paramètre df. Puis ce que les queues changent concrètement à un plan de retrait. Enfin les limites honnêtes de cette modélisation. À la fin, le curseur « Tail df » ne sera plus un réglage ésotérique. Il sera ce qu'il est : le bouton qui décide à quelle fréquence votre simulateur a le droit de produire 2008.

::: cle L'idée en une phrase
À volatilité **identique**, deux mondes peuvent différer du tout au tout par la fréquence de leurs années extrêmes. Dans le monde gaussien, l'année à −30 % réel est un événement de légende. Dans le monde réel (df ≈ 4-6), elle arrive une à deux fois par retraite. La volatilité σ mesure l'agitation **ordinaire**, le paramètre de queue df mesure la propension aux **catastrophes**. Un simulateur qui n'a que σ est aveugle à la moitié du risque qui compte.
:::

::: figure fat-tails
Deux lois de **même moyenne et même volatilité**, mais des queues opposées. Au centre, les années ordinaires se ressemblent. Dans la queue, tout change : sous la loi normale, l'année à −30 % réel est un événement de légende ; sous une Student-t à df 5 (calibrée sur les données réelles), elle est environ **dix fois** plus fréquente. Le paramètre df ne déplace pas le centre, il règle la fréquence des catastrophes.
:::

## Pourquoi la loi normale séduit, et où elle casse

La loi normale n'est pas une naïveté. C'est la conséquence du théorème central limite, qui veut que la somme de nombreux petits effets indépendants tende vers une gaussienne. Si les rendements étaient la somme de milliers de petites nouvelles indépendantes, ils seraient normaux. Et de fait, au centre de la distribution, ils le sont presque : les années ordinaires (−10 % à +20 %) suivent à peu près la cloche.

Le problème est aux extrêmes, et il est massif. La mesure standard est le **kurtosis**, ou coefficient d'aplatissement. La loi normale a un kurtosis de 3. Les rendements **mensuels** des actions affichent typiquement 6 à 12, les quotidiens bien davantage. Autrement dit, les mois et les années extrêmes sont plusieurs fois trop fréquents pour la cloche.

La raison profonde tient en deux mécanismes qui violent les hypothèses du théorème central limite. Le premier : les effets ne sont pas indépendants. La volatilité fait des **grappes**, une grosse variation en annonce d'autres. C'est le « volatility clustering » que modélisent les GARCH. Les extrêmes s'auto-alimentent alors, entre paniques, appels de marge et ventes forcées, chacun vendant parce que l'autre vend. Le second : le marché change de **régime**. La distribution des rendements d'un marché calme et celle d'une crise ne sont pas la même loi. Or mélanger deux gaussiennes de volatilités différentes produit mécaniquement une distribution à queues épaisses.

Les queues ne sont donc pas une bizarrerie. Elles sont la signature statistique d'un fait : les marchés forment un système social à rétroactions, pas une urne ([[regimes-de-marche]]).

Un fait rassurant : l'épaisseur des queues **diminue** avec l'horizon d'agrégation. Les rendements quotidiens sont sauvagement non gaussiens. Les mensuels le sont nettement moins, les annuels encore moins, car agréger douze mois lisse une partie des extrêmes. Et les rendements à dix ans sont presque fréquentables. Pour un plan de retraite, ce sont les échelles mensuelle et annuelle qui comptent, au rythme des retraits. À ces échelles, les queues restent bien réelles. C'est là que le modèle travaille.

## La Student-t et le paramètre df, en clair

Il existe une famille de distributions faite exactement pour ce problème, la **Student-t**. C'est une cloche à trois paramètres : le centre μ, l'échelle (liée à σ) et les **degrés de liberté df**, qui règlent l'épaisseur des queues. À df infini, la Student-t se confond avec la loi normale. À mesure que df descend, le centre change à peine mais les queues s'alourdissent. Sous df ≈ 4, elles deviennent véritablement sauvages. C'est le modèle central : des tirages mensuels Student-t, composés en années ([[historique-vs-parametrique]], [[la-machine-pofo]]).

Ce que df change **concrètement**, à volatilité annuelle identique (σ = 12 %, μ = 4 % réel ; ordres de grandeur d'une année « catastrophique » à −30 % réel, soit un événement à presque 3 écarts-types) :

| df | Le monde qu'il décrit | Fréquence du −30 % réel | Sur 45 ans de retraite |
|---|---|---|---|
| 30+ (≈ normale) | Le monde des manuels | ~1 année sur 400 | Probablement jamais |
| 10 | Queues modérées | ~1 sur 150 | Peut-être une fois |
| 5 (défaut de la page FIRE) | Le monde des données mensuelles réelles | ~1 sur 40 | Une à deux fois |
| 3 | Monde à catastrophes | ~1 sur 20 | Deux à trois fois |

Comparez la ligne df 5 à la ligne df 30 : **le même σ, et un facteur dix sur la fréquence des catastrophes**. Voilà pourquoi deux simulateurs affichant « 12 % de volatilité » peuvent rendre des verdicts sans rapport. Tout est dans la loi des queues, que la plupart des outils commerciaux ne documentent même pas ; ils sont gaussiens sans le dire ([[pieges-des-simulateurs]]).

**D'où vient la valeur de df ?** En mode portefeuille, la page FIRE l'**ajuste** sur vos données. Le kurtosis des rendements mensuels de votre portefeuille se convertit en df. La relation théorique kurtosis ≈ 3 + 6/(df − 4) s'inverse, et des mois à kurtosis 7-9 donnent df ≈ 5-6. Des fonds actions classiques ressortent vers df 4-6. Un portefeuille très diversifié, avec des poches défensives, peut remonter vers 8-12. Le curseur reste ajustable pour les expériences. Le défaut de 5 n'est donc pas une opinion prudente. C'est la valeur qui ressort des données mensuelles réelles de la plupart des portefeuilles d'actions mondiales.

::: attention Ce que df ne mesure pas
Le df de Student est **symétrique** : il épaissit autant la queue des années miraculeuses que celle des désastres. Les vraies distributions de rendements sont en outre **asymétriques** (skew négatif, les extrêmes baissiers sont plus fréquents et plus violents que les haussiers, « l'escalier à la montée, l'ascenseur à la descente »). La Student-t symétrique sous-estime donc légèrement la méchanceté spécifique de la queue gauche. Les correctifs pour cette asymétrie ne passent pas par la distribution mais par la **séquence** : les colonnes stress (sticky bears, volatilité amplifiée dans les phases baissières) et décennie perdue mettent la violence là où elle vit réellement, dans les enchaînements ([[rendre-monte-carlo-pertinent]]).
:::

## Ce que les queues changent à un plan de retrait

Passons du modèle au plan. L'effet des queues épaisses sur un plan de retrait a une structure précise, et la comprendre évite deux contresens symétriques.

**Premier effet, direct : la ruine monte, la médiane ne bouge presque pas.** Passer de df 30 à df 5, à σ constant, laisse le scénario médian quasi identique, car les années ordinaires n'ont pas changé. Mais la probabilité de ruine gonfle, typiquement de 30 à 80 % en relatif selon le plan. Une ruine gaussienne de 4 % devient ainsi 6-7 % en df 5. Les queues ne changent pas votre vie **probable**. Elles changent la fréquence des vies improbables, justement celles contre lesquelles on planifie ([[ruine-et-probabilites]]).

**Deuxième effet, plus subtil : l'interaction avec la séquence.** Une année à −30 % n'a pas le même prix selon sa date. En année 2, elle ampute définitivement le plan ([[sequence-des-rendements]]). En année 30, elle est cosmétique. Les queues épaisses augmentent la probabilité que la fenêtre fragile **contienne** une catastrophe. C'est le produit des deux risques qui fait le danger. Corollaire pratique : les protections de la fenêtre fragile (glidepath, buffer, revenus précoces) sont **aussi** les meilleures protections anti-queues, sans rien ajouter ([[glidepaths]], [[cash-buffer]]).

**Troisième effet : la diversification promet moins qu'annoncé.** L'argument gaussien de la diversification (les corrélations moyennes réduisent σ) sous-estime un fait des crises : dans la queue, les corrélations **montent**. En 2008, tout a baissé ensemble, sauf les obligations d'État longues et le yen. La diversification fonctionne, mais elle fonctionne **moins** bien précisément dans les scénarios pour lesquels on l'achète. Sauf à inclure des actifs dont la décorrélation **survit** aux crises : la duration d'État, l'or, les managed futures systématiques ([[actifs-defensifs]], [[managed-futures]], [[portefeuilles-tous-temps]]). Le kurtosis d'un portefeuille se réduit donc par le **choix** des briques plus que par leur nombre.

Restent deux contresens à éviter. Le contresens optimiste dit : « la médiane ne bouge pas, donc les queues sont un détail. » Non : la planification de retraite est une gestion de queue. Le taux de retrait sûr est défini par les pires cas, pas par la médiane ([[etude-trinity]]). Le contresens catastrophiste dit : « alors mettons df 3 partout et n'en parlons plus. » Non plus. Empiler df 3 sur une moyenne déjà prudente et une ancre CAPE, c'est le triple comptage de la prudence déjà dénoncé ([[rendements-attendus]]). Et un monde df 3 permanent n'est pas le monde réel non plus. La calibration sur données (le fit), plus un test de sensibilité, vaut mieux que ces deux postures.

::: exemple Le test de sensibilité en deux minutes
Plan : 1,5 M€, 52 000 €/an, 45 ans, pension en année 16. Notez la ruine centrale à df ajusté (disons 5) → 5,2 %. Poussez df à 30 → 3,1 %. Descendez-le à 3 → 7,8 %. Lecture : l'intervalle 3-8 % est votre exposition au désaccord sur les queues, comparable ici à l'effet d'un demi-point de μ. Si votre décision (partir ou pas, [[une-annee-de-plus]]) survit aux deux bornes, les queues ne sont pas votre problème dominant. Si elle bascule, votre plan repose sur une hypothèse de docilité du monde. Les protections structurelles (flexibilité écrite, buffer, briques défensives) sont alors un meilleur remède qu'un débat de df ([[choisir-sa-strategie]]).
:::

## Un peu d'histoire des idées, pour finir de s'en convaincre

L'histoire mérite d'être connue, car elle vaccine contre les modèles trop propres. Dès 1900, Bachelier fonde la finance mathématique sur le mouvement brownien gaussien. En 1963, Benoît Mandelbrot étudie les prix du coton. Il montre que les queues sont si épaisses que la variance semble à peine définie, et propose des lois « stables » sauvages. Eugene Fama (1965) le confirme sur les actions. La profession, elle, a besoin de modèles calculables. Elle choisit donc un moyen terme pragmatique : conserver la variance, mais épaissir les queues (Student-t, mélanges, GARCH). Praetz (1972) puis Blattberg et Gonedes (1974) établissent que la Student-t colle remarquablement aux rendements. Les crises de 1987, 1998, 2008 et 2020 ont tranché le débat culturel. LTCM, en 1998, est d'ailleurs tombé précisément pour avoir fait confiance à des queues fines. L'expression « fat tails » est passée du statut d'hérésie mandelbrotienne à celui de fait de base, popularisé par Taleb (*The Black Swan*). La modélisation retenue (Student-t ajustée, plus des régimes pour la séquence) est l'héritière directe de ce compromis : plus honnête que la gaussienne, plus sobre que les lois stables, et calibrable sur **vos** données.

## L'essentiel à retenir

- Les marchés produisent des extrêmes bien plus souvent que la loi normale : kurtosis mensuel 6-12 contre 3 ; c'est la signature des grappes de volatilité et des changements de régime, pas un accident.
- La Student-t ajoute le paramètre manquant : df, l'épaisseur des queues. À σ identique, df 5 rend l'année à −30 % réel ~10 fois plus probable que df 30. La page FIRE ajuste df sur le kurtosis mensuel de **vos** fonds, typiquement 4-6.
- Effet sur le plan : la médiane ne bouge pas, la ruine monte, souvent de +30 à +80 % en relatif. Les queues aggravent surtout la fenêtre fragile, et les protections anti-séquence sont aussi les protections anti-queues.
- Limites du modèle : symétrique (le vrai risque a un skew négatif, couvert par les colonnes stress et décennie perdue) et sans mémoire (couvert par les régimes). La diversification protège moins dans la queue, sauf les briques dont la décorrélation survit aux crises.
- Réflexe pratique : gardez le df ajusté, faites une fois le test 3/30 pour connaître votre exposition, et si la décision bascule, réparez par la structure du plan, pas par le curseur.

---

## Pour aller plus loin

- Mandelbrot, « The Variation of Certain Speculative Prices » (1963) et *The (Mis)Behavior of Markets* (2004) : l'acte fondateur et sa version grand public.
- Fama, « The Behavior of Stock-Market Prices » (1965) ; Blattberg & Gonedes (1974) sur la Student-t : la confirmation académique.
- Taleb, *The Black Swan* (2007) : la culture générale des queues, à lire d'un œil critique.
- Le volet « How this machine works » de la page FIRE : la définition exacte du curseur df et de son fit ([[utiliser-la-page-fire]], [[la-machine-pofo]]).
- La suite dans ce livre : [[regimes-de-marche]] (d'où viennent les grappes) et [[rendre-monte-carlo-pertinent]] (comment queues et régimes se combinent dans le modèle central).
