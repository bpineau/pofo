# docs/: the load-bearing design docs

Curated on purpose: only documents that a future change actually needs
survive here (rationale the code cannot carry, validation records for
bundled data, active ledgers). Implementation plans and session artifacts
are deleted once shipped; the code, its godoc and AGENTS.md are the source
of truth for everything current.

| Doc | Why it stays |
|---|---|
| `catbond-sleeve-design.md` | the insurance-linked family: which reference is bundled and why the market's own is not, how the fund backcasts and the euro hedge are built, and what a 5 to 10 % cat bond sleeve was measured to do to a decumulation book |
| `darcet-permanent-portfolio-design.md` | tactical Permanent Portfolio 2.0 research: macro-breadth signals, reconstruction choices, the empirical-vs-a-priori epistemic ledger; `pkg/permanent` godoc points here |
| `decumulation-fire-design.md` | the FIRE/decumulation engine's design (scenario sources, ruin metrics, solvers); `pkg/decumul` work starts here |
| `decumulation-fire-program-2026-07.md` | the ACTIVE FIRE improvement backlog |
| `decumulation-fire-realism-spec.md` | realism and conservatism principles (valuation anchors, fat tails, why short-window fits flatter); guards against the recurring too-doomy/too-rosy failure modes |
| `dbmfe-simdata-validation-design.md` / `-results.md` | how the bundled DBMFE managed-futures backcast was validated against the SG CTA index, and the evidence; the raw reference series is `SG-CTA-Index-Daily-Returns-since-1999-12-31.csv` |
| `epub-export-design.md` | EPUB 3 export of the embedded books (firebook, then locador): `pkg/bookmd` extraction, `pkg/epub` writer, delivery routes, the on-device validation gate |
| `fire-book-design.md` | the embedded French FIRE book: plan, depth conventions, style rules, progress ledger (`pkg/firebook` godoc points here) |
| `fire-book-en-edition-design.md` | The English edition: Edition value, translated slugs + source stamps + drift report, figure-translation pass, US-framework part, rollout plan (M1 shipped, M2 translation campaign next) |
| `fire-book-en-translation-brief.md` | The English translation campaign, procedurally: pick from `make book-drift`, the `fr-only` marker, generalize vs adapt, stamp, manifest entry, figure dictionary, ledger, one commit per article |
| `fire-book-en-glossary.md` | FR -> EN vocabulary of the book (coinages, glossed finance terms, fixed section titles, untranslatable names, number conventions) that seeds consistency across translation sessions |
| `fire-book-illustrations-2026-07.md` | ACTIVE figure backlog for the withdrawal-strategies part: twenty candidate plates, what each shows and what it costs; fourteen shipped and six deliberately dropped on 2026-07-29, each with its reason; the list is closed and the file can go |
| `fire-book-illustrations-portefeuille-2026-07.md` | ACTIVE figure backlog for the portfolio part: the fourteen reviewers' candidate figures with cost tags, awaiting selection; delete once shipped or dropped |
| `fire-book-illustrations-reste-2026-07.md` | ACTIVE figure backlog for every other part (alternatives, buffers, inflation, tax, human factor, references, science, starter): the 48 reviewers' candidate figures with cost tags, awaiting selection; delete once shipped or dropped |
| `aqr-mf.txt` | the AQR Managed Futures share-class dossier read off the prospectus: fee schedule and access type per class, the RAEF management-fee WAIVER that makes its NAV look like skill, the performance-fee mechanics the IAET backcast leans on, and what each pin depends on |
| `index-benchmarks-design.md` | why `MSCIWORLD`/`SP500` are fee-free long-history index benchmarks with bare ids and no SIM variant |
| `trend-reconstruction-design.md` | the managed-futures field guide: reliability bounds length (where each file stops and why), donor chains over reconstruction, the three references and which anchors what, the measured dead ends, the data survey (fetchable, walled, nonexistent), the per-fund error budget on live overlaps, and the rebuild-from-scratch pipeline spec |
| `ntsg-global-efficient-core-design.md` | the GLOBAL Efficient Core backcast: the four-currency government bond futures basket (local excess returns, measured duration blend, renormalization), the new German/Japanese/British reference series and their validation, the NTSI cross-check |
| `ntsz-eurozone-efficient-core-design.md` | euro-native Efficient Core backcasts and the deep euro reference series (DBXG/MTH long sleeve, equity-leg caveats), with their epistemic ledger |
| `suggest-design.md` | `-suggest`/`-coverage` classification and out-of-sample validation design (`pkg/suggest` godoc points here) |
| `weight-search-design.md` | the bounded/constrained optimizer, the held-out `train:` window and the `-sweep` per-sleeve grid: what each piece answers, where it surfaces, and what was deliberately left for later (Pareto `improve`, the frontier chart) |
| `webapp-design.md` | the `-serve` web constellation: route map, the `/view` URL grammar and its guardrails, catalog-only identifiers, style layering, the M2-M4 ladder (`cmd/pofo/serve.go`/`hub.go`/`view.go`) |
| `webui-instrument-redesign.md` | the shared "instrument" visual identity: tokens, fonts, chart chrome (`pkg/webui` godoc points here) |
