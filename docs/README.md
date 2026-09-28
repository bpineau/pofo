# docs/

Two halves, for two readers.

- [`usage/`](usage/README.md): guides for PEOPLE using pofo, the command,
  the web app and the library. Start there.
- [`specs/`](specs/): records for whoever CHANGES pofo, human or agent: the
  rationale behind decisions the code cannot explain, the validation record
  of bundled data, the traps, the nomenclatures and the recurring procedures.
  `AGENTS.md` says which one to read before touching which feature.

Neither half holds plans, backlogs or pre-implementation specs: a shipped
package's design is its godoc (`go doc ./pkg/<name>`).

## specs/

| Spec | What it records |
|---|---|
| [`aqr-mf.txt`](specs/aqr-mf.txt) | the AQR Managed Futures share classes off the prospectus: fees and access per class, the RAEF fee waiver, the performance-fee mechanics the IAET backcast leans on |
| [`black-litterman-design.md`](specs/black-litterman-design.md) | `optimize:black-litterman`: the file's weights as the prior, lambda from `prior-return`, the He-Litterman golden |
| [`catbond-sleeve-design.md`](specs/catbond-sleeve-design.md) | the insurance-linked reference and fund backcasts, the euro hedge, what a cat bond sleeve does to a decumulation book |
| [`decumul-performance.md`](specs/decumul-performance.md) | the FIRE Monte Carlo's performance record: the bit-for-bit invariant, how a change is proven harmless, what the profiler found, what was refused |
| [`eres-fcpe-design.md`](specs/eres-fcpe-design.md) | the employee-savings funds: the airfund NAV feed, the donor-chain reconstruction, the NAV-timing finding behind the nowcast |
| [`fire-book-design.md`](specs/fire-book-design.md) | the French FIRE book: architecture, writing and style conventions, table of contents, how to add an article |
| [`fire-book-en-edition-design.md`](specs/fire-book-en-edition-design.md) | the English edition: Edition value, slugs, source stamps, drift report, how France-specific passages were adapted |
| [`fire-book-en-translation-brief.md`](specs/fire-book-en-translation-brief.md) | the procedure to translate one article, step by step |
| [`fire-book-en-glossary.md`](specs/fire-book-en-glossary.md) | the book's FR to EN vocabulary, which keeps translation sessions consistent |
| [`fire-envelopes-tax-model-design.md`](specs/fire-envelopes-tax-model-design.md) | the per-envelope tax model, measured and refused; the blended-rate calibration |
| [`inflation-deflators.md`](specs/inflation-deflators.md) | `^HICP-<geo>` and `^CPI-US`: sources, the 2026 Eurostat rebase that froze the old HICP dataset, its validation record, the freshness guards |
| [`index-benchmarks-design.md`](specs/index-benchmarks-design.md) | why `MSCIWORLD`/`SP500` are fee-free index benchmarks with bare ids; the MSCI tail policy |
| [`long-treasury-zero-coupon-design.md`](specs/long-treasury-zero-coupon-design.md) | the US Treasury references, the STRIPS reconstruction, the Vanguard donor defects |
| [`ntsg-global-efficient-core-design.md`](specs/ntsg-global-efficient-core-design.md) | the four-currency bond basket of the Global Efficient Core backcast and its references |
| [`ntsz-eurozone-efficient-core-design.md`](specs/ntsz-eurozone-efficient-core-design.md) | the euro-native backcasts, the deep euro reference series and their conventions, the month-average sweep |
| [`stochastic-lifetime-kernel-design.md`](specs/stochastic-lifetime-kernel-design.md) | the per-path lifetime draw in `pkg/decumul`, its rules and its Gompertz calibration |
| [`trend-reconstruction-design.md`](specs/trend-reconstruction-design.md) | the managed-futures field guide: reliability bounds length, donor chains, references, dead ends, error budgets |
| [`us-total-market-reference-design.md`](specs/us-total-market-reference-design.md) | `USMKT-USD` and the VTI backcast: why the S&P 500 is no stand-in, the measured grossness |
| [`webui-instrument-redesign.md`](specs/webui-instrument-redesign.md) | the web identity: the rules behind it, tokens, fonts, the validated chart palette |
| [`weight-search-design.md`](specs/weight-search-design.md) | the bounded optimizer, the held-out `train:` window, `-sweep`, and their traps |
| [`world-equity-capweight-design.md`](specs/world-equity-capweight-design.md) | why world-equity backcasts hold drifting cap weights, not today's split |
| [`wti-rolled-reference-design.md`](specs/wti-rolled-reference-design.md) | the rolled WTI crude reference: why spot is not investable, the method, its validation, its 2024 end |
