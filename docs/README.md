# docs/: what the code cannot carry

Only three kinds of document live here, and nothing else survives a shipped
feature: the RATIONALE behind a decision the code cannot explain (why this
donor, why not that model), the VALIDATION RECORD of bundled data (what was
measured against which reference, what was walled or dead), and the
PROCEDURE an agent needs months later for a recurring job (translating an
article). Specs, plans, backlogs and campaign ledgers are deleted once the
code, its godoc and AGENTS.md carry the current state; the design of a shipped
package is `go doc ./pkg/<name>`.

| Doc | Why it stays |
|---|---|
| `aqr-mf.txt` | the AQR Managed Futures share-class dossier read off the prospectus: fee schedule and access type per class, the RAEF management-fee WAIVER that makes its NAV look like skill, the performance-fee mechanics the IAET backcast leans on, and what each pin depends on |
| `black-litterman-design.md` | `optimize:black-litterman`: why the prior is the file's weights and not market caps, why lambda comes from `prior-return`, the He-Litterman golden |
| `catbond-sleeve-design.md` | the insurance-linked family: which reference is bundled and why the market's own is not, how the fund backcasts and the euro hedge are built, and what a 5 to 10 % cat bond sleeve was measured to do to a decumulation book |
| `eres-fcpe-design.md` | the Eres employee-savings fund (ERESMONDEM): the airfund NAV feed and its offline snapshot, the two-leg donor-chain reconstruction and its measured validation, the NAV-timing finding that picks URTH as the nowcast proxy, and the nowcast contract |
| `fire-book-design.md` | the embedded French FIRE book: plan, depth conventions, style rules, progress ledger (`pkg/firebook` godoc points here) |
| `fire-book-en-edition-design.md` | the English edition: Edition value, translated slugs, source stamps, drift report, the France-specific passages and how each was generalized or adapted |
| `fire-book-en-translation-brief.md` | the translation procedure: pick from `make book-drift`, the `fr-only` marker, generalize vs adapt, stamp, manifest entry, figure dictionary, one commit per article |
| `fire-book-en-glossary.md` | FR -> EN vocabulary of the book (coinages, glossed finance terms, fixed section titles, untranslatable names, number conventions) that keeps independent translation sessions consistent |
| `fire-envelopes-tax-model-design.md` | the per-envelope tax question, measured and refused (2026-09-10): the calibration recipe for the single blended rate, why the explicit CTO/PEA/AV/PEE model is worth 0.015 point of withdrawal rate against 0.30 for the embedded gain fraction, the PEE break-even |
| `index-benchmarks-design.md` | why `MSCIWORLD`/`SP500` are fee-free long-history index benchmarks with bare ids and no SIM variant; the tail policy and validation bands behind `make msci-refdata` |
| `long-treasury-zero-coupon-design.md` | the bundled US Treasury references: the STRIPS reconstruction and why a geared coupon fund cannot stand in for a zero, the long-yield reference and its assembly, the 2026-09 rebuild of the two total-return series, the provider defects found in the Vanguard donors (VFITX repaired, VFINX repaired against the fund's own published returns, VUSTX measured and left) |
| `ntsg-global-efficient-core-design.md` | the GLOBAL Efficient Core backcast: the four-currency bond futures basket, the German/Japanese/British reference series and their validation, the NTSI cross-check |
| `ntsz-eurozone-efficient-core-design.md` | euro-native Efficient Core backcasts and the deep euro reference series with their epistemic ledger; the bundle-wide month-average stamping sweep of 2026-09-19 (what was re-sourced onto a real curve and what was deliberately left) |
| `stochastic-lifetime-kernel-design.md` | the per-path lifetime draw in `pkg/decumul`: why posterior survival weighting cannot price an annuity or an estate, the "the kernel draws the death, the household never sees it" rule, the Gompertz calibration against INSEE and which way its errors point |
| `suggest-design.md` | `-suggest`/`-coverage` classification and its out-of-sample validation design |
| `trend-reconstruction-design.md` | the managed-futures field guide: reliability bounds length, donor chains over reconstruction, the three references and which anchors what, the measured dead ends, the data survey (fetchable, walled, nonexistent), the per-fund error budget |
| `us-total-market-reference-design.md` | the US total-market reference `USMKT-USD` and the VTI backcast: why a 0.99 correlation to the S&P 500 does not make it a stand-in, the measured grossness constant |
| `webui-instrument-redesign.md` | the shared "instrument" visual identity: tokens, fonts, the CVD-validated chart palette and why (`pkg/webui` godoc points here) |
| `weight-search-design.md` | the bounded/constrained optimizer, the held-out `train:` window and the `-sweep` grid: the traps (zero risk-free Sharpe, `Spec.Train` inert inside `pkg/optimize`) and what was deliberately left for later |
| `world-equity-capweight-design.md` | why the world-equity reconstructions hold cap-weighted drifting weights instead of today's split carried back (a +1.8 pts/yr look-ahead bias over 1988-2008), the published 2009-10-31 anchor |
| `wti-rolled-reference-design.md` | the rolled WTI crude reference: why spot is not investable, the roll method, the per-year validation against the published S&P GSCI Crude Oil total return, the 1985-2024 bound, which sources answered or were walled |
