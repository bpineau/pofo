# Les pièges des simulateurs (indépendance, biais américain, survivant...)

Entrez le même plan dans cinq simulateurs de retraite réputés : même capital, mêmes dépenses, même horizon. Vous obtiendrez cinq probabilités de succès, qui peuvent s'étaler de 99 % à 70 %. Aucun de ces outils ne ment. Chacun a simplement fait, souvent sans le documenter, une dizaine de choix de modélisation. Et chacun de ces choix déplace le verdict de quelques points.

Cette page est l'inventaire raisonné de ces choix. Voici les dix pièges qui séparent un simulateur flatteur d'un simulateur honnête. Pour chacun, vous trouverez le mécanisme, l'ordre de grandeur de son effet sur la ruine, la question d'audit à poser à n'importe quel outil, et la réponse qu'un simulateur honnête y apporte (parfois, « le piège est irréductible, voici la marge à prendre »). Elle complète [[monte-carlo-forces-faiblesses]], la machine en général, en descendant dans la plomberie. C'est aussi, en creux, un guide d'achat. Vous saurez auditer en dix questions le prochain simulateur qu'on vous vantera.

::: cle La hiérarchie des pièges
Tous les pièges ne pèsent pas pareil. Par ordre d'impact typique sur la ruine d'un plan FIRE : le biais d'échantillon des données (facteur 2 à 5 sur la ruine), l'erreur sur les paramètres d'entrée (facteur 2), l'absence de queues épaisses (+30 à 80 % relatif), l'indépendance des tirages (+20 à 50 % relatif), les frictions omises (frais, impôts, équivalent à 0,5-1,5 point de retrait), puis les pièges de lecture. Un outil peut être irréprochable sur les cinq derniers et inutilisable à cause du premier.
:::

## Les pièges de données : sur quoi la machine a appris

**Piège 1 : le biais américain et le biais du survivant.** La plupart des simulateurs historiques rejouent les États-Unis depuis 1926, le pays gagnant du siècle, seul de son espèce ([[anarkulova-cederburg]]). L'effet est net. Les taux « sûrs » ressortent 0,5 à 1,5 point trop hauts par rapport à l'expérience développée complète. La ruine d'un 4 % rigide à horizon long passe d'environ 2 % sur données US à 15-17 % sur échantillon mondial. La question d'audit tient en une phrase : sur quelles données, de quel pays, depuis quand ? La réponse honnête garde la colonne broad-sample (16 pays, 1870-2020) en permanence à côté du reste ([[historique-vs-parametrique]]).

**Piège 2 : le survivant au carré, dans vos propres données.** Plus sournois encore, il y a les données de vos propres fonds. Un portefeuille actuel est fait de fonds qui existent **encore**, car les fonds morts ont disparu des historiques. On les a souvent choisis pour leur belle décennie. Et les historiques reconstruits (backfill) des ETF récents empruntent des indices choisis rétrospectivement. Calibrer sur cette matière hérite d'un optimisme structurel. La parade s'appuie sur des séries d'**indices** longs et documentés plutôt que sur les seuls historiques de fonds, ce que font les reconstructions `SIM`. Et le modèle central reste de toute façon tiré vers un prior prudent ([[rendre-monte-carlo-pertinent]]).

**Piège 3 : la fenêtre courte.** Vingt ans de données ne contiennent **aucune** retraite longue indépendante. Selon la fenêtre, ils ne contiennent pas non plus d'inflation persistante, ni de décennie perdue. Un simulateur qui rééchantillonne (bootstrap) vos vingt dernières années n'explore que les recombinaisons d'un seul régime. C'est la borne optimiste par construction. La question d'audit est simple : que fait l'outil quand l'historique est plus court que l'horizon ? La bonne réponse avertit explicitement dans ce cas. Elle refuse le modèle des cohortes quand les fenêtres n'existent pas. Et elle fait glisser le paramétrique vers le prior mondial à proportion du déficit d'historique.

## Les pièges de moteur : comment la machine tire

**Piège 4 : les tirages indépendants (i.i.d.).** Le Monte-Carlo standard tire chaque année sans mémoire. Or les vrais marchés font des grappes, comme les marchés baissiers pluriannuels, et des tendances de valorisation. Du coup, les longues séquences médiocres, celles qui ruinent, sont sous-produites. La ruine est sous-estimée de 20 à 50 % en relatif pour les plans sensibles à la séquence ([[sequence-des-rendements]]). La question d'audit : les mauvaises années peuvent-elles s'enchaîner dans votre moteur ? La parade fait tourner trois moteurs à mémoire, le bootstrap par blocs, les cohortes et les régimes de Markov du stress, face au central i.i.d. L'écart entre le central et le stress **mesure** alors ce piège sur votre plan.

**Piège 5 : la loi normale.** Des queues fines rendent les catastrophes quasi impossibles ([[queues-epaisses]]). L'effet sur la ruine va de −30 à −80 % en relatif. Or la majorité des simulateurs commerciaux sont gaussiens sans le dire. La question d'audit : quelle est la probabilité d'une année à −30 % réel dans votre modèle ? Si la réponse revient à « une fois tous les quatre siècles », vous êtes fixé. La parade emploie une loi de Student-t, dont le paramètre df est ajusté sur le kurtosis mensuel de vos fonds.

**Piège 6 : l'agrégation grossière.** Simuler en pas **annuel**, avec un retrait en début d'année, lisse les à-coups intra-annuels. Le retraité réel, lui, vend chaque mois, y compris pendant les six mois où le marché a perdu 30 %. L'effet reste modeste mais réel, de l'ordre de quelques dixièmes de point de ruine, et surtout il produit des trajectoires irréalistes. La parade travaille en panel et tirages mensuels en mode portefeuille, avec des retraits mensuels optionnels façon salaire et une composition exacte ([[la-machine-pofo]]).

## Les pièges de périmètre : ce que la machine ne compte pas

**Piège 7 : les frictions omises.** Frais de gestion, fiscalité des ventes, frais de transaction : la plupart des grilles historiques les ignorent, et Trinity est brut de tout ([[etude-trinity]]). L'effet cumulé vaut 0,5 à 1,5 point de retrait, soit la différence entre 4 % et 2,8 % net pour un contrat chargé. La question d'audit : votre taux de succès est-il avant ou après frais et impôts ? Un simulateur honnête majore chaque vente de sa fiscalité, au taux mixte que vous réglez et avec une part de gains croissante au fil du plan ([[flat-tax-et-imposition]]). Les frais des fonds, eux, sont déjà dans les prix des ETF simulés. Restent les frais d'enveloppe, à intégrer dans vos dépenses ([[combien-il-vous-faut]]).

**Piège 8 : les dépenses de laboratoire.** Le retrait « constant en termes réels » est une fiction commode. Les vraies dépenses ont une dérive propre, tirée par la santé, et un profil en sourire ([[depenses-en-retraite]]). Elles connaissent des à-coups, comme une toiture ou la dépendance. Elles suivent surtout une **inflation personnelle** qui peut diverger de l'indice ([[suivre-inflation]]). L'effet joue dans les deux sens, mais l'omission de la dérive santé est la plus coûteuse à long horizon. Un bon outil rend la dérive réelle réglable, propose le sourire de Blanchett en option, et simule nativement les règles de dépense réalistes (flex, guardrails, VPW, ABW) plutôt que la seule caricature rigide.

## Les pièges de lecture : ce qu'on fait dire à la machine

**Piège 9 : le succès binaire.** « 95 % de succès » traite de la même façon l'échec à 71 ans et l'échec à 94 ans une fois la pension acquise. Ce chiffre met sur le même plan le succès qui finit à 40 fois la mise et celui qui finit à 3 000 € ([[ruine-et-probabilites]]). D'où des décisions absurdes aux marges, comme travailler trois ans de plus pour passer de 94 à 97 % alors que les échecs étaient tous tardifs et bénins. La parade déplie ce binaire. Elle donne la date et la cause des échecs (§05), la richesse finale distribuée (bequest), la ruine croisée avec la mortalité et le niveau de vie servi (§04).

**Piège 10 : l'utilisateur qui optimise le verdict.** C'est le p-hacking de scénarios : pousser le rendement moyen μ « parce que mon fonds a fait mieux », relancer jusqu'au chiffre confortable, choisir la colonne verte, arrondir 7 % à « environ 5 % ». C'est le piège terminal, et aucun logiciel ne peut l'empêcher. Il peut seulement le rendre **visible**. Par conception, un outil honnête affiche les quatre modèles toujours côte à côte et des ancres de rappel comme le CAPE et le prior broad-sample. Il explique chaque curseur par une aide au survol. Et son solveur (§09) reformule tout écart en prix concret, en euros, en années ou en flexibilité, plutôt qu'en débat de curseurs. Côté comportement, la parade tient aux huit règles du bon usage ([[monte-carlo-forces-faiblesses]]) et à la revue annuelle à entrées auditées ([[revue-annuelle]]).

::: attention Le piège commercial, en une remarque
Les simulateurs ne naissent pas neutres. Un outil de gestionnaire d'actifs a intérêt à des projections engageantes, car on collecte sur l'espoir. Un vendeur de rentes a intérêt à des ruines effrayantes, car on vend de la peur. Un planificateur payé à l'heure a intérêt à la complexité. Ce n'est pas du complot, c'est de la microéconomie. Demandez toujours qui a construit l'outil et ce qu'il vend. Les outils sans produit à vendre (cFIREsim, FICalc, la toolbox d'ERN, pofo) ne sont pas forcément meilleurs techniquement, mais leur erreur n'a pas de direction préférée.
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
Plan : 1,2 M€, 45 000 €/an, 50 ans. Simulateur A (gaussien, données US, brut de frais) → 97 % de succès. Simulateur B (rejeu historique US) → 95 %. Simulateur C (i.i.d. Student-t calibré prudent) → 91 %. Simulateur D (à mémoire, échantillon mondial) → 88 % en colonne stress, 84 % en colonne broad-sample. L'étalement 84-97 n'est **pas** du désaccord technique à arbitrer par la moyenne. C'est un dégradé de pièges levés un à un : frais et impôts, −2 points ; queues, −2 ; mémoire des marchés baissiers, −3 ; échantillon mondial, −4. La lecture professionnelle est simple. Le plan vit quelque part dans la moitié basse du dégradé, et les chiffres hauts mesurent surtout ce que leurs outils ignorent. Toute comparaison de simulateurs devrait se lire comme cette colonne de soustractions, jamais comme un concours.
:::

## L'essentiel à retenir

- Cinq simulateurs, cinq verdicts, et chaque écart est un piège levé ou non. La hiérarchie des dégâts : données biaisées > paramètres incertains > queues fines > tirages sans mémoire > frictions omises > lectures naïves.
- Les pièges de données (américain, survivant, fenêtre courte) dominent tout : demandez **toujours** sur quoi la machine a appris avant de regarder son verdict.
- Les pièges de moteur (i.i.d., gaussien, pas annuel) se testent en une question : « quelle est la probabilité d'une année à −30 % réel, et peut-elle être suivie d'une autre ? ».
- Les pièges de périmètre (frais, impôts, dépenses idéalisées) valent 0,5 à 1,5 point de retrait : un taux de succès brut de tout est un chiffre de brochure.
- Le dernier piège est l'utilisateur : aucune machine n'empêche le p-hacking de scénarios ; la parade est procédurale (entrées auditées, multi-modèles, décision sur les colonnes dures, revue annuelle). C'est exactement la conception de la page FIRE.

---

## Pour aller plus loin

- Early Retirement Now, volet 26 (les non-dits du 4 %) et la critique des simulateurs de la série ([[serie-ern]]).
- Anarkulova, Cederburg & O'Doherty (2023) : l'ampleur chiffrée du piège n° 1 ([[anarkulova-cederburg]]).
- Kitces.com, « Does Monte Carlo Analysis Actually Overstate Tail Risk In Retirement Projections? » : le contrepoint intelligent, où le retour de valorisation resserre les cônes longs, utile pour ne pas basculer dans le catastrophisme systématique.
- Dans ce livre : [[monte-carlo-forces-faiblesses]] (la machine), [[historique-vs-parametrique]] (les familles), [[rendre-monte-carlo-pertinent]] (les corrections implémentées), [[la-machine-pofo]] (la plomberie exacte, piège par piège).
