# Les pièges des simulateurs (indépendance, biais américain, survivant...)

Entrez le même plan (même capital, mêmes dépenses, même horizon) dans cinq simulateurs de retraite réputés, et vous obtiendrez cinq probabilités de succès qui peuvent s'étaler de 99 % à 70 %. Aucun de ces outils ne ment ; chacun a simplement fait, souvent sans le documenter, une dizaine de choix de modélisation dont chacun déplace le verdict de quelques points.

Cette page est l'inventaire raisonné de ces choix : les dix pièges qui séparent un simulateur flatteur d'un simulateur honnête, avec pour chacun le mécanisme, l'ordre de grandeur de l'effet sur la ruine, la question d'audit à poser à n'importe quel outil, et la réponse qu'un simulateur honnête y apporte (parfois, « le piège est irréductible, voici la marge à prendre »). Elle complète [[monte-carlo-forces-faiblesses]] (la machine en général) en descendant dans la plomberie ; c'est aussi, en creux, un guide d'achat : vous saurez auditer en dix questions le prochain simulateur qu'on vous vantera.

::: cle La hiérarchie des pièges
Tous les pièges ne pèsent pas pareil. Par ordre d'impact typique sur la ruine d'un plan FIRE : le biais d'échantillon des données (facteur 2 à 5 sur la ruine), l'erreur sur les paramètres d'entrée (facteur 2), l'absence de queues épaisses (+30 à 80 % relatif), l'indépendance des tirages (+20 à 50 % relatif), les frictions omises (frais, impôts, équivalent à 0,5-1,5 point de retrait), puis les pièges de lecture. Un outil peut être irréprochable sur les cinq derniers et inutilisable à cause du premier.
:::

## Les pièges de données : sur quoi la machine a appris

**Piège 1 : le biais américain et le biais du survivant.** La plupart des simulateurs historiques rejouent les États-Unis depuis 1926 : le pays gagnant du siècle, seul de son espèce ([[anarkulova-cederburg]]). Effet : les taux « sûrs » ressortent 0,5 à 1,5 point trop hauts par rapport à l'expérience développée complète ; la ruine d'un 4 % rigide à horizon long passe de ~2 % (données US) à ~15-17 % (échantillon mondial). Question d'audit : « sur quelles données, de quel pays, depuis quand ? ». Réponse : la colonne broad-sample (16 pays, 1870-2020) en permanence à côté du reste ([[historique-vs-parametrique]]).

**Piège 2 : le survivant au carré, dans VOS données.** Plus sournois : les données de vos propres fonds. Un portefeuille actuel est fait de fonds qui existent **encore** (les fonds morts ont disparu des historiques), souvent choisis pour leur belle décennie, et les historiques reconstruits (backfill) des ETF récents empruntent des indices choisis rétrospectivement. Calibrer sur cette matière hérite d'un optimisme structurel. Réponse : les reconstructions `SIM` s'appuient sur des séries d'**indices** longs documentés plutôt que sur les seuls historiques de fonds, et le modèle central est de toute façon tiré vers un prior prudent ([[rendre-monte-carlo-pertinent]]).

**Piège 3 : la fenêtre courte.** Vingt ans de données ne contiennent **aucune** retraite longue indépendante et, selon la fenêtre, aucune inflation persistante ou aucune décennie perdue. Un simulateur qui bootstrappe vos vingt dernières années explore les recombinaisons... d'un seul régime. Effet : c'est la borne optimiste par construction. Question d'audit : « que fait l'outil quand l'historique est plus court que l'horizon ? ». Réponse : avertissement explicite dans ce cas, refus du modèle cohortes quand les fenêtres n'existent pas, et blending du paramétrique vers le prior mondial à proportion du déficit d'historique.

## Les pièges de moteur : comment la machine tire

**Piège 4 : les tirages indépendants (i.i.d.).** Le Monte-Carlo standard tire chaque année sans mémoire ; les vrais marchés font des grappes (marchés baissiers pluriannuels) et des tendances de valorisation. Effet : les longues séquences médiocres, celles qui ruinent, sont sous-produites ; la ruine est sous-estimée de 20 à 50 % en relatif pour les plans sensibles à la séquence ([[sequence-des-rendements]]). Question d'audit : « les mauvaises années peuvent-elles s'enchaîner dans votre moteur ? ». Réponse : trois moteurs à mémoire (bootstrap par blocs, cohortes, régimes de Markov du stress) affichés contre le central i.i.d. : l'écart central/stress **mesure** ce piège sur votre plan.

**Piège 5 : la loi normale.** Queues fines = catastrophes quasi impossibles ([[queues-epaisses]]). Effet : −30 à −80 % relatif sur la ruine. La majorité des simulateurs commerciaux sont gaussiens sans le dire. Question d'audit : « quelle est la probabilité d'une année à −30 % réel dans votre modèle ? » (si la réponse est « une fois tous les quatre siècles », vous savez). Réponse : Student-t, df ajusté sur le kurtosis mensuel de vos fonds.

**Piège 6 : l'agrégation grossière.** Simuler en pas **annuel** avec un retrait en début d'année lisse les à-coups intra-annuels : le retraité réel vend chaque mois, y compris pendant les six mois où le marché a perdu 30 %. Effet modeste mais réel (quelques dixièmes de point de ruine), et surtout des trajectoires irréalistes. Réponse : panel et tirages mensuels en mode portefeuille, retraits mensuels optionnels (« salary-like »), composition exacte ([[la-machine-pofo]]).

## Les pièges de périmètre : ce que la machine ne compte pas

**Piège 7 : les frictions omises.** Frais de gestion, fiscalité des ventes, frais de transaction : absents de la plupart des grilles historiques (Trinity est brut de tout, [[etude-trinity]]). Effet cumulé : 0,5 à 1,5 point de retrait, soit la différence entre 4 % et 2,8 % net pour un contrat chargé. Question d'audit : « votre taux de succès est-il avant ou après frais et impôts ? ». Réponse : la fiscalité majore chaque vente (au taux mixte que vous réglez, part de gains croissante au fil du plan, [[flat-tax-et-imposition]]) ; les frais des fonds sont déjà dans les prix des ETF simulés ; les frais d'enveloppe restent à intégrer dans vos dépenses ([[combien-il-vous-faut]]).

**Piège 8 : les dépenses de laboratoire.** Le retrait « constant en termes réels » est une fiction commode : les vraies dépenses ont une dérive propre (santé), un profil en sourire ([[depenses-en-retraite]]), des à-coups (toiture, dépendance), et surtout une **inflation personnelle** qui peut diverger de l'indice ([[suivre-inflation]]). Effet : dans les deux sens, mais l'omission de la dérive santé est la plus coûteuse à long horizon. Réponse : dérive réelle réglable, sourire de Blanchett optionnel, et les règles de dépense réalistes (flex, guardrails, VPW, ABW) simulées nativement plutôt que la seule caricature rigide.

## Les pièges de lecture : ce qu'on fait dire à la machine

**Piège 9 : le succès binaire.** « 95 % de succès » traite pareil l'échec à 71 ans et l'échec à 94 ans avec pension acquise, le succès qui finit à 40 fois la mise et celui qui finit à 3 000 € ([[ruine-et-probabilites]]). Effet : des décisions absurdes aux marges (travailler trois ans de plus pour passer de 94 à 97 %, alors que les échecs étaient tous tardifs et bénins). Réponse : la date et la cause des échecs (§05), la richesse finale distribuée (bequest), la ruine croisée avec la mortalité, le niveau de vie servi (§04) : le binaire est déplié.

**Piège 10 : l'utilisateur qui optimise le verdict.** Le p-hacking de scénarios : pousser μ « parce que mon fonds a fait mieux », relancer jusqu'au chiffre confortable, choisir la colonne verte, arrondir 7 % à « environ 5 % ». C'est le piège terminal, et aucun logiciel ne peut l'empêcher. Il peut seulement le rendre **visible**. Réponse, par conception : les quatre modèles toujours côte à côte, les ancres de rappel (CAPE, broad-sample prior), les aides au survol qui expliquent chaque curseur, et le solveur §09 qui reformule tout écart en prix concret (euros, années, flexibilité) plutôt qu'en débat de curseurs. Réponse comportementale : les huit règles du bon usage ([[monte-carlo-forces-faiblesses]]) et la revue annuelle à entrées auditées ([[revue-annuelle]]).

::: attention Le piège commercial, en une remarque
Les simulateurs ne naissent pas neutres. Un outil de gestionnaire d'actifs a intérêt à des projections engageantes (on collecte sur l'espoir) ; un outil de vendeur de rentes a intérêt à des ruines effrayantes (on vend de la peur) ; un planificateur payé à l'heure a intérêt à la complexité. Ce n'est pas du complot, c'est de la microéconomie : demandez toujours qui a construit l'outil et ce qu'il vend. Les outils sans produit à vendre (cFIREsim, FICalc, la toolbox d'ERN, pofo) ne sont pas forcément meilleurs techniquement, mais leur erreur n'a pas de direction préférée.
:::

## La grille d'audit complète

Les dix questions à poser à tout simulateur, avec le piège associé :

| # | Question d'audit | Piège testé |
|---|---|---|
| 1 | Quelles données, quel pays, quelle profondeur ? | Biais américain/survivant |
| 2 | D'où viennent les hypothèses de mes actifs ? | Survivant au carré, backfill |
| 3 | Que se passe-t-il si l'historique < horizon ? | Fenêtre courte |
| 4 | Les mauvaises années peuvent-elles s'enchaîner ? | i.i.d. |
| 5 | Quelle probabilité pour une année à −30 % réel ? | Gaussianité |
| 6 | Quel pas de temps, quand ont lieu les retraits ? | Agrégation |
| 7 | Avant ou après frais et impôts ? | Frictions |
| 8 | Puis-je modéliser mes vraies règles de dépense ? | Dépenses de laboratoire |
| 9 | Que sait-on des échecs (date, cause, gravité) ? | Succès binaire |
| 10 | Qu'est-ce qui m'empêche de me raconter des histoires ? | P-hacking, incitations |

Un outil qui répond clairement à huit de ces dix questions est un instrument ; un outil qui n'en documente que deux est une brochure.

::: exemple Cinq verdicts pour un même plan, réconciliés
Plan : 1,2 M€, 45 000 €/an, 50 ans. Simulateur A (gaussien, données US, brut de frais) : 97 % de succès. Simulateur B (rejeu historique US) : 95 %. Simulateur C (i.i.d. Student-t calibré prudent) : 91 %. Simulateur D (à mémoire, échantillon mondial), colonne stress : 88 % ; colonne broad-sample : 84 %. L'étalement 84-97 n'est **pas** du désaccord technique à arbitrer par la moyenne. C'est un dégradé de pièges levés un à un (frais et impôts, −2 points ; queues, −2 ; mémoire des marchés baissiers, −3 ; échantillon mondial, −4). La lecture professionnelle : le plan vit quelque part dans la moitié basse du dégradé ; les chiffres hauts mesurent surtout ce que leurs outils ignorent. Toute comparaison de simulateurs devrait se lire comme cette colonne de soustractions, jamais comme un concours.
:::

## L'essentiel à retenir

- Cinq simulateurs, cinq verdicts : chaque écart est un piège levé ou non ; la hiérarchie des dégâts : données biaisées > paramètres incertains > queues fines > tirages sans mémoire > frictions omises > lectures naïves.
- Les pièges de données (américain, survivant, fenêtre courte) dominent tout : demandez **toujours** sur quoi la machine a appris avant de regarder son verdict.
- Les pièges de moteur (i.i.d., gaussien, pas annuel) se testent en une question : « quelle est la probabilité d'une année à −30 % réel, et peut-elle être suivie d'une autre ? ».
- Les pièges de périmètre (frais, impôts, dépenses idéalisées) valent 0,5 à 1,5 point de retrait : un taux de succès brut de tout est un chiffre de brochure.
- Le dernier piège est l'utilisateur : aucune machine n'empêche le p-hacking de scénarios ; la parade est procédurale (entrées auditées, multi-modèles, décision sur les colonnes dures, revue annuelle). C'est exactement la conception de la page FIRE.

---

## Pour aller plus loin

- Early Retirement Now, volet 26 (les non-dits du 4 %) et la critique des simulateurs de la série ([[serie-ern]]).
- Anarkulova, Cederburg & O'Doherty (2023) : l'ampleur chiffrée du piège n° 1 ([[anarkulova-cederburg]]).
- Kitces.com : « Does Monte Carlo Analysis Actually Overstate Tail Risk In Retirement Projections? » : le contrepoint intelligent (le retour de valorisation resserre les cônes longs), utile pour ne pas basculer dans le catastrophisme systématique.
- Dans ce livre : [[monte-carlo-forces-faiblesses]] (la machine), [[historique-vs-parametrique]] (les familles), [[rendre-monte-carlo-pertinent]] (les corrections implémentées), [[la-machine-pofo]] (la plomberie exacte, piège par piège).
