# docs/: design docs and plans

One document per feature; read the matching doc BEFORE reworking a feature
(`AGENTS.md` says which are load-bearing for which task). `*-design.md` is
the durable reference (findings, algorithms, data sources, trade-offs);
`*-plan.md` is the implementation plan that built it (kept for archaeology,
often stale afterwards); `*-spec.md`/`*-proposals.md` fed a specific rework.

| Doc | One line |
|---|---|
| `cwarp-design.md` | CWARP (Cole Wins Above Replacement Portfolio) statistic: definition, params, report integration |
| `daily-backcast-followups.md` | status ledger of the daily-granularity backcast campaign (treasuries, SP500, WTI, FX) and what remains |
| `darcet-permanent-portfolio-design.md` | tactical Permanent Portfolio 2.0: research notes, macro-breadth signals, the empirical-vs-a-priori epistemic ledger (pkg/permanent) |
| `dbmfe-simdata-validation-design.md` / `-results.md` | how the DBMFE managed-futures backcast was validated against SG CTA, and what came out |
| `decumulation-fire-design.md` | the FIRE/decumulation engine: scenario sources, ruin metrics, solvers (pkg/scenario + pkg/decumul) |
| `decumulation-fire-plan.md` | the (large, historical) implementation plan of the above |
| `decumulation-fire-program-2026-07.md` | the 2026-07 FIRE improvement backlog/program |
| `decumulation-fire-realism-spec.md` | realism & conservatism rules: valuation anchors, fat tails, why short-window fits are optimistic |
| `decumulation-fire-rewrite-spec.md` | the FIRE explorer UX rewrite spec (central case first, ruin figures usable) |
| `decumulation-fire-usability-proposals.md` | earlier proposals that fed the rewrite spec |
| `decumulation-fire-v3-enrichment.md` | the v3 enrichment drop: vintages, decades, income sections |
| `decumulation-monthly-sampling-design.md` / `-plan.md` | monthly (vs daily) sampling for the historical return models |
| `finador-integration-design.md` / `-plan.md` | how ../finador consumes pofo as a library (market data, metrics, charts) |
| `fire-book-design.md` | the embedded French FIRE book: plan, depth conventions, style rules, progress ledger (pkg/firebook) |
| `fire-book-review.md` | 2026-07-13 review pass of the book |
| `fire-redesign-terminal-design.md` | the FIRE explorer's dark "terminal" visual direction |
| `index-benchmarks-design.md` | fee-free `MSCIWORLD`/`SP500` index benchmarks: why bare ids are long-history and never SIM |
| `intraday-and-library-consumption-design.md` / `-plan.md` | intraday quotes and the library-consumer API surface |
| `latest-quote-design.md` / `-plan.md` | `Latest`: freshest-quote resolution for real-time valuation |
| `ntsz-eurozone-efficient-core-design.md` | NTSZ / euro-native Efficient Core backcasts, incl. the long euro govt sleeve (DBXG, MTH) and the euro reference series |
| `suggest-design.md` | `-suggest`/`-coverage`: metadata classification + out-of-sample validated gap-filling (pkg/suggest) |
| `webui-instrument-redesign.md` | the shared "instrument" visual identity (pkg/webui): tokens, fonts, chart chrome |

`superpowers/` is an unrelated vendored skill bundle; ignore it.
`SG-CTA-Index-Daily-Returns-since-1999-12-31.csv` is the raw SG CTA
reference series used by the DBMFE validation.
