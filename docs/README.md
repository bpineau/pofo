# docs/: the load-bearing design docs

Curated on purpose: only documents that a future change actually needs
survive here (rationale the code cannot carry, validation records for
bundled data, active ledgers). Implementation plans and session artifacts
are deleted once shipped; the code, its godoc and AGENTS.md are the source
of truth for everything current.

| Doc | Why it stays |
|---|---|
| `darcet-permanent-portfolio-design.md` | tactical Permanent Portfolio 2.0 research: macro-breadth signals, reconstruction choices, the empirical-vs-a-priori epistemic ledger; `pkg/permanent` godoc points here |
| `decumulation-fire-design.md` | the FIRE/decumulation engine's design (scenario sources, ruin metrics, solvers); `pkg/decumul` work starts here |
| `decumulation-fire-program-2026-07.md` | the ACTIVE FIRE improvement backlog |
| `decumulation-fire-realism-spec.md` | realism and conservatism principles (valuation anchors, fat tails, why short-window fits flatter); guards against the recurring too-doomy/too-rosy failure modes |
| `dbmfe-simdata-validation-design.md` / `-results.md` | how the bundled DBMFE managed-futures backcast was validated against the SG CTA index, and the evidence; the raw reference series is `SG-CTA-Index-Daily-Returns-since-1999-12-31.csv` |
| `epub-export-design.md` | EPUB 3 export of the embedded books (firebook, then locador): `pkg/bookmd` extraction, `pkg/epub` writer, delivery routes, the on-device validation gate |
| `fire-book-design.md` | the embedded French FIRE book: plan, depth conventions, style rules, progress ledger (`pkg/firebook` godoc points here) |
| `index-benchmarks-design.md` | why `MSCIWORLD`/`SP500` are fee-free long-history index benchmarks with bare ids and no SIM variant |
| `ntsz-eurozone-efficient-core-design.md` | euro-native Efficient Core backcasts and the deep euro reference series (DBXG/MTH long sleeve, equity-leg caveats), with their epistemic ledger |
| `suggest-design.md` | `-suggest`/`-coverage` classification and out-of-sample validation design (`pkg/suggest` godoc points here) |
| `webapp-design.md` | the `-serve` web constellation: route map, the `/view` URL grammar and its guardrails, catalog-only identifiers, style layering, the M2-M4 ladder (`cmd/pofo/serve.go`/`hub.go`/`view.go`) |
| `webui-instrument-redesign.md` | the shared "instrument" visual identity: tokens, fonts, chart chrome (`pkg/webui` godoc points here) |
