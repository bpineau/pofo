# FIRE book: revue (2026-07-13)

## Passe 2 (retours de Ben) : appliquée

- **Bug de rendu corrigé** : le gras pouvait contenir de l'italique
  (`**Auteur, *Titre* (an)**`) — les `**` restaient bruts dans toute la
  section Livres de la bibliothèque. Résolu (italique avant gras) + test.
- **Références externes en liens** : impots.gouv.fr, info-retraite.fr,
  urssaf.fr, earlyretirementnow.com, portfoliocharts.com, bogleheads.org,
  kitces.com, tpawplanner.com, morningstar.com, r/financialindependence, etc.
  (68 liens ; aucune URL inventée).
- **« au Part 54 » → « au volet 54 »** (135 renvois ERN, anglicisme corrigé).
- **Dé-gras des mots-outils** : 284 emphases injustifiées retirées (articles,
  prépositions, conjonctions faibles, copules, pronoms faibles), les
  emphases porteuses gardées (« jamais », « sans », « votre »...).
- **Chaînes de « : » allégées** : 1069 « : » internes aux parenthèses passés
  en virgules (phrases à 4+ deux-points : 390 → 213) + réécriture manuelle du
  récit le plus lourd. Les énumérations inline (listes structurées) restent.
- **Illustrations** : support `::: figure` ajouté au moteur ; 4 diagrammes SVG
  thématisés (risque de séquence, CAPE→taux sûr, sourire des dépenses, grille
  croissance × inflation). Extensible (voir `figures.go`).
- Reste possible : étendre la passe « phrases courtes » aux énumérations
  inline (les convertir en vraies listes markdown), et ajouter d'autres
  diagrammes (plateau d'allocation, courbe taux-horizon, frontière des règles,
  fan chart annoté).

---

## Passe 1 (revue initiale)

Note de revue du livre FIRE (`pkg/firebook`, servi à `/book/fr/`). Elle liste
ce qui a été corrigé dans cette passe, le style qui reste à travailler, et
surtout les MANQUES de contenu et les clarifications souhaitables.

## Ce qui a été fait dans cette passe

- **Emphase en gras plutôt qu'en capitales.** 2 512 mots français mis en
  CAPITALES pour l'emphase (« le SENS et la TAILLE des flux ») passés en gras
  minuscule, comme locador. Acronymes (CAPE, ETF, CTO, ERN...), mot FIRE,
  sigles sans voyelle, PACS, M@REL et noms propres préservés. Conversion
  scriptée + recapitalisation en début de phrase.
- **Présentation façon locador.** Palette chaude scopée à `.book` (papier
  crème + dégradé ambre, accent brun-ambre au lieu du petrol), titres en serif
  éditorial (Georgia) sur corps sans, gras encre sombre, liens ambre, encarts
  à emoji + libellés (🔑 astuce 💡, ⚠, 🧮, 🔬, 🗣).
- **Intros aérées.** Le gros bloc d'ouverture (souvent 4-6 phrases d'un tenant)
  découpé en 2-3 paragraphes dans 71 articles.
- **Coquilles.** Résidus d'élision de la conversion corrigés (« plan D'une
  page » -> « d'une », « à L'inflation » -> « l' », 34 « À » -> « à » en milieu
  de phrase) ; allègement de « : car » -> «, car ».
- **Vérifs.** Liens `[[...]]` tous valides (garde par test), aucun lien externe
  mal formé, aucun mot répété fautif, gras équilibré, aucun tiret cadratin,
  tests verts.

## Style qui reste à travailler (optionnel, sur décision)

- **La chaîne de « : » est systémique.** Moyenne de ~3,9 « : » par ligne de
  prose : c'est la signature du style, mais ça alourdit. Beaucoup de « : »
  jouent le rôle d'un point ou d'une virgule + conjonction. Un vrai
  allègement (remplacer une partie des « : d'où / : et / : mais / : donc » par
  «. » ou «, ») demanderait une passe manuelle article par article (risque de
  casser un rythme voulu si fait en masse). À arbitrer : je peux le faire sur
  les articles d'entrée (Démarrer) si tu veux un rendu plus « grand public ».
- **Gras triviaux.** La conversion a produit quelques « **la** », « **de** »,
  « **et** » (mots-outils qui étaient en capitales). Rares et peu gênants ; une
  passe pourrait les dé-emphaser (retirer le gras des mots-outils isolés).
- **Tableaux larges sur mobile.** Corrigé dans cette passe : chaque table est
  désormais enveloppée d'un `div.table-wrap` à `overflow-x:auto`.

## Manques de contenu / à compléter

1. **Illustrations : le manque n°1.** Le livre est 100 % texte alors que le
   sujet s'y prête et que la spec initiale voulait des illustrations. pofo a
   déjà `pkg/chart` (SVG stdlib). Schémas à forte valeur, un par article clé :
   le risque de séquence (deux trajectoires, même moyenne, issues opposées),
   la courbe CAPE -> SAFEMAX, le « sourire » des dépenses, le plateau
   taux/allocation, la courbe taux-vs-horizon qui s'aplatit, la frontière des
   règles, un fan chart annoté, la grille croissance x inflation à 4 quadrants.
   Nécessite d'ajouter le support image/SVG au mini-moteur markdown (absent
   aujourd'hui) + génération des SVG. C'est le plus gros levier de qualité.
2. **Fact-check des chiffres avant publication large.** Beaucoup de nombres
   sont écrits de mémoire/par raisonnement et sont cohérents, mais à confirmer
   sur sources vives : les tables SAFEMAX par CAPE et par horizon (ERN), la
   série Morningstar 3,3 -> 3,8 -> 4,0 -> 3,7 % et ses millésimes exacts, les
   chiffres Anarkulova-Cederburg (38 pays / ~2 500 années-pays, 17 % d'échec,
   ~2,26 %), l'espérance de vie en bonne santé (~64-65 ans, INSEE/DREES), les
   quantiles de survie. Écrire un petit spec de validation façon
   `make golden` serait dans l'esprit du dépôt.
3. **Chapitres fiscaux datés (2026) à maintenir.** PUMa (la formule donnée est
   approximative : le décret 2019 a des subtilités d'assiette et de plafond),
   seuils AV/PEA/donation/PFU/RVTO, retraite (réforme 2023 mouvante). Déjà
   signalés en tête de chaque chapitre ; prévoir une relecture à chaque loi de
   finances / LFSS.
4. **Sujets peu ou pas couverts** (à ajouter si tu veux élargir) :
   - le **rééquilibrage** n'a pas d'article dédié (fondu dans l'allocation) :
     bandes vs calendrier, l'étude ERN Part 39, le rééquilibrage par les flux ;
   - les **cryptomonnaies** : seulement écartées en une phrase ; vu la
     fréquence de la question, un encadré « pourquoi pas (encore) dans un plan
     de retrait » serait utile ;
   - le **non-coté / private equity / crowdfunding** : absent ; un mot sur leur
     inadéquation à la décumulation (illiquidité, J-curve) manque ;
   - l'**assurance-vie luxembourgeoise** (triangle de sécurité, super-privilège)
     pour gros patrimoines et non-résidents : non mentionnée ;
   - le **PER en sortie** (capital vs rente, fiscalité de sortie) mériterait un
     encadré comparatif plus étoffé.
5. **Clarifications ponctuelles souhaitables :**
   - **Échelle de linkers en euro** : je répète qu'elle est « peu accessible en
     direct » sans donner de how-to concret. Un lecteur voudra les véhicules
     précis (ETF à échéance indexés s'ils existent, ou construction via OAT€i
     chez quel type de courtier). Actuellement flou (à jour de l'offre).
   - **Renvois aux sections §00-§09 de la page FIRE** : dépendants de l'UI ; si
     la page renumérote/réorganise ses sections, ces renvois se périment. La
     dépendance est notée dans `la-machine-pofo` mais concerne ~15 articles.
   - **Relation kurtosis -> df** (queues-epaisses) : la formule
     `kurtosis ≈ 3 + 6/(df-4)` gagnerait une note sur son domaine de validité.

## Traduction

Le livre est FR uniquement ; la traduction EN à `/book/en/` reste à venir (le
segment de langue de l'URL et l'arbo `assets/book/fr/` la préparent déjà). À
faire APRÈS stabilisation du contenu FR (pour ne pas traduire deux fois).
