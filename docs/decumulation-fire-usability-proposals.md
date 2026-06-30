# FIRE explorer — making the ruin figures usable (Claude's proposals)

Status: proposals for discussion (2026-06-30). Independent take, written to be
**synthesised later with a ChatGPT proposal** into a single plan.

Companion to `decumulation-fire-design.md`, `decumulation-fire-realism-spec.md`,
`decumulation-fire-followups.md`. Code read: `pkg/scenario/{regime,bootstrap,
cohorts,source}.go`, `pkg/decumul/{plan,outcome,simulate,run}.go`,
`pkg/decumul/web/{model.go,assets/app.js,assets/index.html}`.

---

## 0. The complaint, restated

The headline ruin probability moves from **0.1% to 27.9% to 92.3%** as toggles
are flipped, with no guidance on which configuration is realistic. The
accompanying text disclaims ("read the ruin figures with care") instead of
resolving the ambiguity. There is no picture of the market the simulation
assumes. Net effect: the tool offloads the single hardest judgement (which
assumptions are correct) onto the user, who came to the tool precisely because
he lacks that judgement. An estimate that ranges over three orders of magnitude
and refuses to commit carries no actionable information.

This is **two distinct failures** that must be fixed together:

1. **A modelling failure**: the toggles are not orthogonal, they are not
   anchored to anything, and one of them is mis-calibrated to the point of being
   wrong. So the spread is artificially huge and partly spurious.
2. **A framing failure**: a single hypersensitive probability is the wrong
   headline, presented as if the user should pick the assumptions, with no
   central case, no validated anchor, and no visualization to sanity-check.

---

## 1. The modelling smoking gun: "Stress regimes" secretly tanks the mean

`pofo` builds the stress source from the calm sliders like this
(`model.go:109`):

```go
BearMu: pr.Mu - 2*pr.Sigma, BearSigma: pr.Sigma * 1.5,
StayCalm: 0.92, StayBear: 0.65
```

Worked through (verified numerically):

| Sliders | Bear mean | Bear σ | π(bear) | **Blended long-run real mean** |
|---|---|---|---|---|
| default (μ 4%, σ 16%) | **−28%/yr** | 24% | 18.6% | **−1.95%/yr** |
| conservative (μ 3%, σ 18%) | **−33%/yr** | 27% | 18.6% | **−3.70%/yr** |

So enabling "Stress regimes" does **not** isolate sequence-of-returns risk. It
silently converts the portfolio into one whose **expected real return is
negative** (−2% default, −3.7% with the conservative prior), held there forever.
No globally diversified equity portfolio in the historical record has a negative
expected real return over a retirement; this is well below even Japan post-1990
or the 1929–32 / 1970s episodes. That single mis-calibration is what produces
the 92% ruin, and it is *entangled* with the conservative prior (which already
lowered μ and raised σ), so the two toggles **double-count** the pessimism.

Sequence risk is about the **ordering and clustering** of returns around an
unchanged long-run mean, not about secretly lowering the mean. The current
"stress" toggle conflates the two, which is exactly why the user cannot reason
about it: he turns on what is labelled "cluster bad years" and unknowingly also
subtracts ~6 points off the average annual return.

**This is the root cause of "0.1% vs 92%".** Two of the levers are not
independent dials of one underlying reality; they are an unanchored,
multiplicative pessimism stack with one badly-calibrated member.

---

## 2. What is actually realistic (so we have a target to calibrate to)

The point of the tool is to land on a *defensible central number*. The
literature has converged enough to give one. Anchors, for a globally
diversified ~100% equity sleeve, real, EUR-ish investor:

- **Bengen 1994 / Trinity**: 4% real, 30y, US 1926–. ~95% success. Known to be
  **optimistic**: US survivorship bias, only 30 years, only one country.
- **Anarkulova, Cederburg & O'Doherty 2023** (broad sample, 38 developed
  markets, 1890–2019, real, realistic longevity): a fixed 4% real rule fails far
  more often than US backtests imply; their 5%-failure SWR is **~2.26% real**
  for the full-longevity case. This is the pessimistic, sequence-risk-aware end.
- **Cederburg et al. "Beyond the Status Quo / Stocks for the Long Run?"**: in
  the broad sample, a globally diversified **all-equity** portfolio has *lower*
  failure than the classic 60/40, because bonds carry their own long-run real
  risk. Relevant: do not assume "more bonds = safer" for a 50y horizon.
- **Morningstar State of Retirement Income** (Blanchett/Benz, forward-looking,
  valuation-aware): ~**3.7%** starting safe rate for 90% success over 30y with a
  fixed real withdrawal; flexibility raises it.
- **Kitces / Big-ERN (Karsten)**: valuation regime (CAPE) matters a lot, and a
  **50–60 year FIRE horizon** (vs 30y) pulls the safe fixed rate down to roughly
  **3.0–3.5% real**. ERN's CAPE-based dynamic SWR work is the most FIRE-specific.
- **Guyton-Klinger / dynamic rules**: with genuine spending flexibility, a
  meaningfully higher *initial* rate (≈5%) can hit the same success, at the cost
  of variable spending. The gap between "fixed" and "flexible" is itself the
  most useful number a FIRE planner can see.

**Defensible central case for this tool**: a fixed real rule, ~100% global
equity, ~45–50y horizon → a **~3.0–3.5% SWR for ~90–95% success**. The honest
headline ruin at 4% fixed / 45y should land somewhere around **10–20%**, not
0.1% and not 92%. Today the tool can produce both extremes and offers no reason
to believe either. The fix is to make the central case land here *by
construction and by validation*, and present the rest as a labelled band around
it.

---

## 3. Proposals

Five workstreams. A and C give the biggest usability jump for the least code; B
removes the spurious spread; D buys trust; E makes the number actionable.

### A. Replace the single number with a calibrated central estimate + an explicit band

The deepest fix is conceptual: **stop asking the user to choose the assumptions,
and stop reporting one point.** Report, always, three pre-defined scenarios side
by side, with the central one emphasised:

```
            Optimistic     CENTRAL        Pessimistic
Ruin           1%            12%             28%
SWR (95%)     4.4%          3.3%            2.4%
```

- **Optimistic** = the fund's own rosy fit, flex on, 40y (today's "0.1%" world),
  explicitly labelled "US-style backtest, likely too kind".
- **Central** = honest forward-looking world-equity assumptions, fixed rule,
  45–50y, *properly-calibrated* sequence risk (see B). This is the number to
  plan on, and it is highlighted.
- **Pessimistic** = broad-sample prior + sequence risk + 55–60y longevity, the
  "stress test you should survive".

This directly answers "what do I do with 0.1%–28%?": the answer is **the spread
is the point**, shown deliberately, with the planning number called out, instead
of being discovered accidentally by toggling. The à-la-carte checkboxes become
an "Advanced" disclosure for power users, not the primary interface.

Each scenario must state, in one plain line, *what it assumes* (μ, σ, horizon,
fixed vs flexible), so the band is legible rather than magical.

### B. Make the levers orthogonal and anchored (fix the model, shrink the spurious spread)

1. **Decouple sequence risk from the mean.** Re-calibrate `MarkovRegime` so the
   **blended long-run real mean is preserved** at the calm/prior mean (target
   the geometric mean explicitly), and the bear state injects *clustering,
   persistence and negative skew* only. Concretely: pick `BearMu`/`StayBear`/
   `piBear` so that (a) blended mean ≈ prior geometric mean, and (b) the
   resulting **worst rolling 10y / 30y real CAGR matches historical broad-sample
   statistics** (e.g. a lost decade near 0% real, a worst-30y near the
   Anarkulova tail), not by an ad-hoc `μ − 2σ`. The bear mean should be roughly
   −10% to −15%/yr real over ~2–4y clusters, not −28% to −33%.
2. **Calibrate, don't guess.** Add a test that drives the regime source to a
   *target worst-decade and target failure rate* and asserts the parameters
   reproduce a published benchmark within tolerance (see D). The regime
   parameters become *derived from a calibration target*, not magic constants.
3. **Stop double-counting.** "Conservative prior" and "Stress regime" currently
   stack multiplicatively. In the scenario-pack model (A) the packs are
   pre-composed and validated, so the user never accidentally stacks two
   independent pessimism multipliers. If the advanced toggles remain, make the
   regime *re-use the prior's mean* rather than re-deriving an even-deeper bear
   from the already-lowered μ.
4. **Optional, higher fidelity (the real W3 fix)**: bundle a long broad-sample
   real-return panel (DMS-style or a curated multi-country monthly series) under
   `pkg/datasets/` and bootstrap from it. Then "sequence risk" is *empirical*,
   not synthetic, and the calibration anchor is automatic. This is the
   gold-standard option flagged in the realism spec; it is more work and has
   data-licensing questions, so it can follow the synthetic re-calibration.

### C. Show the market the simulation assumes (the user explicitly asked for this)

Right now the only charts are buffer-arbitrage and recovery-time. The user
cannot see whether "stress" means a single −80% mega-crash or a March-2020
repeated and durable. Add:

1. **Wealth fan chart**: portfolio real value over the horizon, median line with
   p5–p95 (and p25–p75) bands, plus **a few highlighted sample paths including
   representative ruin paths**. This is the single most reassuring/illuminating
   chart and turns the abstract ruin% into something visceral.
2. **Worst-case inspector**: for the worst 5% of paths, state in plain words
   what happened: peak-to-trough real drawdown, its depth and duration, the
   worst single year, the worst decade CAGR. ("In the bad 5%: a −52% real
   drawdown bottoming at year 7, taking 14 years to recover.") We already
   compute CDaR, worst-10y, years-underwater; surface them as a *narrative*, not
   just cards.
3. **Return distribution / regime ribbon**: a small histogram of the simulated
   annual real returns (so the user sees the left tail and the mean), and, for
   the regime model, a ribbon showing the calm/bear state along one sample path
   so "clustered bad years" is visible. This also makes the model.B bug
   self-evident: a histogram centred below zero immediately looks wrong.
4. **Anchor line on the success curve**: plot success vs withdrawal rate (E.2)
   and mark the classic "4% rule" and the broad-sample SWR so the user sees
   where his plan sits relative to the literature.

### D. Earn trust: validate against published benchmarks, on screen

The realism spec lists "match Anarkulova's exact numbers" as a non-goal. Fair
for exactness, but the tool currently has **no anchor at all**, which is why no
number feels trustworthy. Add at least:

1. **A calibration test** (`pkg/decumul`): under a documented "US-historical-ish"
   parameterisation, a fixed 4% / 30y reproduces the Trinity ~95% success within
   tolerance; under the broad-sample parameterisation, it reproduces the
   Anarkulova-class failure within tolerance. If these two don't bracket
   correctly, the model is mis-calibrated, and the test will say so.
2. **An on-screen "sanity anchor"**: one line near the headline, e.g. "For
   reference: the classic US 4%/30y backtest gives ~95% success; broad
   century-long samples give ~75–80% for the same fixed rule. Your central case
   sits at N%." This *positions* the user's number inside known results instead
   of leaving it free-floating.

### E. Make the headline actionable (invert the question)

Ruin% is hard to act on. The questions the user actually has ("do I have
enough?", "what rate is safe?", "is my portfolio too aggressive?") are better
answered as:

1. **Safe withdrawal rate at a confidence level, in euros.** Promote the
   existing `CapitalForRuin` logic into a first-class headline: "At 95% success
   over 50y, your safe spending is **€X/yr (3.3%)**; you currently plan €48k
   (4.0%)." This is the single most useful sentence the tool can produce, and
   the engine already has the machinery. Today it is buried in a "Solve" panel.
2. **Success-vs-withdrawal-rate curve** (not one point): show the slope, so the
   user sees how fast risk rises as he spends more, and how much margin a small
   spending cut buys. Mark his current rate.
3. **"Am I too aggressive?" answered directly**: in the A/B allocation
   comparison, report not just ruin but the **worst-case real drawdown and worst
   decade** per allocation, so "too volatile" becomes concrete. The volatility
   term-structure / variance-ratio work already in the backlog feeds this.
4. **Quantify the value of flexibility.** Show fixed-rule ruin *and*
   guardrails/flex ruin together, labelled "the cost of rigidity": "Fixed: 18%.
   With a 10% spending cut in downturns: 6%." That reframes flex from "cheating
   that hides ruin" (today's framing) into "the most powerful lever you control",
   which is both true and actionable. It also makes Anarkulova's pessimism
   legible: it is the price of refusing to adapt.

---

## 4. Recommended default configuration (the central case)

- Return model: forward-looking world equity, **μ ≈ 3.0–3.5% real arithmetic,
  σ ≈ 16–18%, df ≈ 4–5**, with sequence risk on via the *re-calibrated* regime
  (mean-preserving).
- Horizon: **50y** default for FIRE (allow 40–60).
- Withdrawal: **fixed real rule** for the headline (honest), with the flexible
  variant shown alongside as the "value of adapting" number.
- Headline: **SWR at 95% success in €/yr**, plus the three-scenario band
  (optimistic / central / pessimistic), plus the wealth fan chart.
- Anchor line citing Trinity and broad-sample results.

Target sanity check: fixed 4% / 45–50y under this central case lands ~10–20%
ruin; the 95%-success SWR lands ~3.0–3.5%. If it doesn't, the model is wrong,
and the calibration test (D.1) catches it.

---

## 5. Effort / sequencing

1. **B.1–B.3 (re-calibrate the regime, mean-preserving)** — small, high impact,
   removes the spurious 92%. *Do first.*
2. **D.1 calibration test** — small, locks B in and makes every later number
   trustworthy.
3. **A scenario-pack headline (optimistic/central/pessimistic band)** — medium,
   the core UX reframe.
4. **C.1 wealth fan chart + C.2 worst-case narrative** — medium, the "show me
   the market" the user asked for.
5. **E.1 SWR-in-euros headline + E.2 success curve** — medium, mostly wiring
   existing engine output.
6. **E.4 value-of-flexibility, C.3/C.4, D.2 anchor line** — small polish.
7. **B.4 bundled broad-sample panel** — large, optional, the gold-standard
   fidelity upgrade for later.

---

## 6. Open questions for Ben

- **Scenario packs vs keep the à-la-carte toggles?** I strongly favour packs as
  the primary UI (toggles demoted to "Advanced"). Confirm.
- **Bundle a broad-sample historical panel (B.4)?** Worth the data/licensing
  effort, or stay synthetic-but-calibrated? Which dataset (DMS is not freely
  redistributable; a curated MSCI-region or country-index monthly panel?).
- **Exact central-case μ/σ/df** and the published anchors to cite (Morningstar
  3.7% / ERN 3.25–3.5% / Anarkulova 2.26%): which do you want as the on-screen
  references?
- **Headline metric**: SWR-in-euros-at-95% as the hero number (my
  recommendation), or keep ruin% as hero with SWR secondary?
