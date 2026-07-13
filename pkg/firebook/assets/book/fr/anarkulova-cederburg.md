# Au-delà des États-Unis : Anarkulova, Cederburg et l'échantillon mondial

Toute la tradition du taux de retrait sûr, de Bengen à Trinity et à la quasi-totalité des simulateurs en ligne, repose sur un seul jeu de données : les marchés américains depuis 1926 ([[etude-trinity]]). Or les États-Unis du XXe siècle ne sont pas un échantillon neutre : c'est le pays qui a gagné le siècle. Depuis les années 2010, une lignée de travaux, Dimson-Marsh-Staunton, Wade Pfau, puis surtout Aizhan Anarkulova, Scott Cederburg et Michael O'Doherty, a refait les calculs sur l'expérience **complète** des pays développés, et les résultats dérangent : le taux de retrait « sûr » mondial est nettement sous 4 %, et les portefeuilles obligataires protègent moins qu'on ne le croyait. Cette page explique ces travaux, leurs chiffres, les critiques qu'on peut leur faire, et comment pofo les met littéralement à votre disposition : le modèle « broad sample » de la page FIRE est construit sur ces données.

::: cle L'idée en une phrase
Si vous admettez que la France, le Japon, l'Allemagne ou l'Italie de 1900-2020 sont des futurs **possibles** pour un investisseur développé d'aujourd'hui, alors le risque de long horizon est plus élevé que ce que l'histoire américaine seule laisse croire : sur l'échantillon mondial, la règle des 4 % rigide échoue nettement plus souvent, et le « taux à 5 % d'échec » d'un couple de 65 ans est plutôt vers 2,3-2,7 % selon les hypothèses. Ce n'est pas la seule lecture possible du monde, mais c'est la borne prudente la mieux documentée dont on dispose.
:::

## Le biais du survivant géographique

Commençons par le problème. Les données Ibbotson démarrent en 1926 à New York : elles contiennent la Grande Dépression et la stagflation, mais **aussi** un pays jamais envahi, jamais en défaut sur sa dette intérieure, dont la monnaie est devenue la réserve mondiale, et dont le marché actions a été **le** grand gagnant du siècle. Choisir ce pays comme unique échantillon, c'est calibrer son plan sur le billet gagnant de la loterie.

Ce que l'échantillon américain ne contient pas, et que le XXe siècle développé a pourtant produit en abondance : des marchés actions fermés ou spoliés (Allemagne 1948, Japon 1946), des hyperinflations qui pulvérisent les obligations ([[hyperinflation-et-extremes]]), des décennies perdues profondes (Japon post-1990 : les actions sous leur sommet réel pendant plus de trente ans), des défauts et des répressions financières. Un rentier allemand, japonais, français ou italien parti en 1900-1960 avec un plan « à l'américaine » a, dans une fraction non négligeable des cas, été ruiné non par malchance de séquence mais parce que son **pays** a traversé l'histoire.

Dimson, Marsh et Staunton (le « triumvirat de Cambridge », auteurs du *Global Investment Returns Yearbook*) ont chiffré l'écart dès 2002 : rendement réel des actions 1900-2020, ~6,5 %/an aux États-Unis contre ~4,5 % pour le monde hors États-Unis ; et des obligations d'État qui, dans la moitié des pays, ont fait **pire** que 0 % réel sur de longues périodes. Wade Pfau applique dès 2010 la méthode de Bengen à 17 pays : le SAFEMAX à 30 ans, ~4 % aux États-Unis, tombe sous 3 % dans la majorité des pays et sous 1,5 % dans les pires (France comprise, plombée par les inflations d'après-guerres).

## Anarkulova, Cederburg, O'Doherty : la méthode moderne

Les travaux d'Anarkulova, Cederburg et O'Doherty (« The Safe Withdrawal Rate: Evidence from a Broad Sample of Developed Markets », 2023, et la série d'articles sœurs dont « Beyond the Status Quo », 2023) modernisent la question sur trois plans.

**Les données** : la base la plus propre disponible sur les pays développés (construite sur le référentiel GFD et les travaux académiques de long terme), 38 pays développés, environ 2 500 années-pays de rendements réels actions, obligations et monétaire, avec un soin particulier contre les biais de survie et d'anticipation (un pays entre dans l'échantillon quand il est développé À L'**époque**, pas rétrospectivement : l'Argentine de 1900, alors riche, y figure ; c'est le point qui fâche, on y revient).

**La méthode** : plutôt que rejouer des fenêtres d'un seul pays, un **bootstrap par blocs** : on tire des blocs de dix ans (pour préserver les grappes, tendances et régimes, [[sequence-des-rendements]]) dans l'ensemble pays × époques, et on assemble des retraites synthétiques de la durée voulue. Chaque retraite simulée vit donc l'histoire d'**un** pays développé cohérent, morceau par morceau, catastrophes comprises. C'est très exactement ce que fait le modèle « broad sample » de pofo, qui embarque le panel académique Jorda-Schularick-Taylor (16 pays, 1870-2020) et le rejoue en blocs par pays sur un portefeuille 60/40 domestique ([[la-machine-pofo]]).

**La mortalité** : au lieu d'un horizon fixe de 30 ans, un couple de 65 ans avec les vraies tables de mortalité, ce qui donne des retraites de durée aléatoire, parfois 35 ans et plus.

Les résultats principaux, pour un couple de 65 ans en 60/40 domestique : la règle des 4 % rigide échoue dans environ 17 % des cas (contre ~2 % sur les données américaines seules) ; le taux à 5 % d'échec ressort vers **2,26 %**. Pour un retraité précoce à horizon plus long, c'est pire encore. Et un résultat dérangeant de la même équipe (« Beyond the Status Quo ») : dans leur échantillon, les portefeuilles **tout** actions (diversifiés internationalement) dominent les mélanges actions-obligations à long horizon, parce que les obligations se font périodiquement détruire par l'inflation précisément dans les mêmes époques que les actions locales : la diversification internationale des actions protège mieux que les obligations domestiques ([[diversification-internationale]], [[obligations-en-retrait]]).

::: science Où situer ces chiffres dans la littérature
Retenez le spectre des bornes pour un horizon long : histoire américaine seule, ~3,25-3,5 % rigide ([[serie-ern]]) ; rendements prospectifs Morningstar sur 30 ans, ~3,7 % ([[guardrails-morningstar]], [[rendements-attendus]]) ; échantillon mondial Anarkulova-Cederburg, ~2,3-2,7 %. L'écart entre ces bornes n'est **pas** du désaccord technique : ce sont des réponses à des questions différentes (« et si le futur ressemble à l'Amérique / aux anticipations actuelles / au siècle développé entier ? »). Un plan sérieux connaît les trois chiffres et choisit consciemment où il se place, plutôt que d'ignorer les deux qui dérangent.
:::

## Les critiques honnêtes

Ces travaux ont leurs contradicteurs sérieux (ERN en tête, qui y a consacré plusieurs analyses), et les objections méritent d'être connues, car elles bornent la lecture pessimiste comme le biais américain borne l'optimiste.

**Le poids des catastrophes militaires.** Une part importante des échecs de l'échantillon vient des guerres mondiales **vécues sur place** (Allemagne, Japon, Autriche, France...) et de leurs hyperinflations. Si votre scénario de destruction du capital est une occupation militaire, un portefeuille en ETF n'était de toute façon pas votre vrai problème. Contre-argument : retirer les guerres retire aussi les seules observations de « queue politique » dont on dispose, et des événements de cette classe (confiscation, répression financière, monnaie détruite) n'exigent pas une guerre mondiale.

**Le cas argentin et les frontières de l'échantillon.** Inclure des pays « développés à l'époque » qui ont ensuite décroché (l'Argentine) est méthodologiquement défendable (c'est exactement le biais de survie qu'on veut éviter : en 1900, personne ne savait qui décrocherait) mais tire les chiffres vers le bas pour un investisseur des pays cœur d'aujourd'hui.

**L'investisseur simulé est domestique.** Les retraites simulées vivent l'histoire d'**un** pays (actions **et** obligations locales). Un investisseur mondialisé d'aujourd'hui, en ETF monde non couvert en change, n'aurait pas vécu le Japon 1990 ou l'Italie 1970 en plein : la diversification internationale amortit précisément les pires blocs de l'échantillon. C'est probablement la critique la plus importante en pratique, et c'est un argument central **pour** la diversification ([[diversification-internationale]]) plus que **contre** l'étude.

**Le chevauchement des blocs** et la taille effective de l'échantillon : 2 500 années-pays semblent beaucoup, mais les crises sont mondiales et corrélées (1929, 1973, 2008 frappent tout le monde) ; l'échantillon de désastres **indépendants** reste petit. L'incertitude sur ces chiffres est donc elle-même large.

La synthèse raisonnable : le « vrai » risque d'un investisseur mondialisé d'aujourd'hui se situe quelque part **entre** l'histoire américaine et l'échantillon mondial domestique, sans qu'on sache où précisément. D'où la conception de pofo : les deux bornes affichées côte à côte, en permanence.

## Ce que ça change pour votre plan

**Le taux rigide de dimensionnement.** Si vous dimensionnez sans marges ([[combien-il-vous-faut]]), l'existence de la borne mondiale justifie 3-3,5 % plutôt que 4 %, et interdit de considérer 4 % comme « scientifiquement sûr » sur 45 ans. En revanche, viser 2,3 % (43 fois les dépenses !) par littéralisme serait sur-réagir : c'est la borne domestique-catastrophiste, que vos marges réelles (pension, flexibilité, diversification) dominent largement.

**La lecture de la page FIRE.** La colonne broad-sample de pofo N'**est pas** votre portefeuille : c'est un 60/40 domestique rejoué à travers le siècle des 16 pays, curseurs ignorés. Si votre plan tient dans cette colonne, il tient dans le pire monde développé documenté : c'est le meilleur label de robustesse disponible. S'il n'y tient pas, regardez **où** échouent les scénarios (souvent : blocs inflationnistes, [[inflation-et-taux-de-retrait]]) et ce qui manque à votre portefeuille pour ces régimes ([[portefeuilles-tous-temps]], [[actifs-defensifs]]).

**Le portefeuille.** Deux leçons directes : la diversification internationale des actions n'est pas un raffinement, c'est **la** protection contre le risque dominant de l'échantillon (le décrochage d'un pays, fût-il le vôtre) ; et les obligations nominales domestiques ne sont pas l'actif sûr du long horizon : leur pire ennemi (l'inflation soutenue) est aussi celui du rentier ([[obligations-indexees]], [[or-en-retrait]]).

::: exemple Le même plan sous les deux bornes
Plan : 1,4 M€, 45 000 €/an rigides (3,2 %), 45 ans, pension 15 000 €/an à 67 ans, actions mondiales 70 % / obligations 30 %. Fenêtres historiques du portefeuille : ruine ~1 %. Central calibré : ~4 %. Broad-sample : ~9 %, échecs concentrés dans les blocs à inflation persistante, aux trois quarts après 80 ans, pension acquise. Lecture : le plan est solide (3,2 % avec pension est déjà prudent) ; la queue broad-sample s'adresse par 10-15 % de flexibilité écrite ([[flexibilite-realite]]) plutôt que par du capital en plus. Sans la colonne broad-sample, on n'aurait jamais su que le mode de défaillance résiduel était l'inflation, pas le krach.
:::

## L'essentiel à retenir

- L'histoire américaine est le billet gagnant du siècle : calibrer un plan dessus, c'est hériter de son biais optimiste ; le monde développé complet raconte une histoire plus dure.
- Anarkulova-Cederburg-O'Doherty (38 pays, bootstrap par blocs, mortalité réelle) : le 4 % rigide échoue ~17 % du temps, le taux à 5 % d'échec est vers 2,3 % pour un couple domestique 60/40 ; et les obligations domestiques protègent moins que la diversification internationale des actions.
- Les critiques (guerres, Argentine, investisseur domestique, crises corrélées) sont sérieuses : la vérité d'un investisseur mondialisé est **entre** les deux bornes ; personne ne sait où.
- En pratique : 3-3,5 % rigide pour dimensionner, la colonne broad-sample de pofo comme label de robustesse, la diversification internationale et les actifs anti-inflation comme réponses aux modes de défaillance qu'elle révèle.
- Ces données sont dans votre outil : le modèle broad-sample de la page FIRE rejoue le panel JST 16 pays, 1870-2020 ([[la-machine-pofo]], [[utiliser-la-page-fire]]).

---

## Pour aller plus loin

- Anarkulova, Cederburg & O'Doherty, « The Safe Withdrawal Rate: Evidence from a Broad Sample of Developed Markets » (2023) et Cederburg et al., « Beyond the Status Quo: A Critical Assessment of Lifecycle Investment Advice » (2023) : les papiers sources (SSRN, accès libre).
- Dimson, Marsh & Staunton, *Triumph of the Optimists* (2002) et le *Global Investment Returns Yearbook* (annuel, UBS) : le siècle mondial en chiffres.
- Wade Pfau, « An International Perspective on Safe Withdrawal Rates » (2010) : le précurseur.
- Early Retirement Now sur ces études (critiques Parts consacrées) : la contradiction argumentée ([[serie-ern]]).
- Jorda, Schularick & Taylor, « The Rate of Return on Everything, 1870-2015 » : le panel académique que pofo embarque.
