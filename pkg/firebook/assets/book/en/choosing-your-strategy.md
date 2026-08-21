# Choosing your strategy: criteria, comparison, worked case
<!-- source: choisir-sa-strategie @ 0b833710619f -->

Nine articles took the rules apart one at a time, and a tenth replayed them side by side on three retirements that actually happened ([[seven-ways-to-live-on-one-portfolio]]). What is left is the decision. It is less technical than it looks from here. Good modern rules, properly set, land closer together than you fear. The gap between two good rules is almost always smaller than the one an input error opens: spending underestimated, a pension forgotten ([[ten-plan-wrecking-mistakes]]).

But "close" is not "identical". Rules differ in the **shape** of the risk they make you live with, and in what they demand of whoever runs them. That is where the choice is settled: temperament, the floor, governance. This article walks the whole decision procedure: the five steps in order, the profile-to-rule matrix, the hybrids worth having, the summary table for the part, and one case run end to end in a simulator. The most useful of those hybrids is also the most overlooked: a different rule for each phase of the plan. The judging criteria themselves are set elsewhere, from the 5th percentile you read instead of the average to what regret and robustness are worth ([[deciding-under-uncertainty]]).

By the end you will have a rule that is written down, calibrated and tested: yours.

::: cle The principle behind the decision
You are not picking "the best rule". You are picking which shape of risk you would rather live with: a rare bankruptcy, bounded adjustments, or steady variability. You pick it under the constraint of your real floor, and for one specific person, the one who will be holding the pen twenty years from now. Three questions, in this order. What can I **take** (the floor)? What do I **want** to live on (temperament)? Who will **run** it (governance)? The technical part comes after, and one sitting settles it.
:::

::: figure arbre-decision
The whole procedure fits on one page, and its order is not decorative. Each step narrows the field, so the expensive questions only come up when two candidates are left. Skipping step 1 to go straight to comparing rules is the most common mistake in the field: you polish the second-order decision and leave the first-order one wide open.
:::

## Step 1: the base, before any rule

No rule can be chosen without three settled numbers ([[how-much-you-need]]): the **floor** (the spending you can hold for five years, morale included), **comfort** (the life you are aiming at), and how much of the floor is **covered from outside** the portfolio, meaning a pension discounted to its start date, annuities and other near-certain income ([[annuities-and-safety-first]]; for what the US system actually promises, [[us-healthcare-and-social-security]]). The ratio that decides everything fits in one question. What share of the floor does the portfolio have to fund, and for how many years (the uncovered phase, [[horizon-and-life-expectancy]])? Everything else follows from it. A floor 80% covered by guaranteed income opens **every** rule, even the most variable ones. A floor funded 100% by the portfolio for twenty years closes half of them.

## Step 2: the admission test, family by family

Every family has one eliminating test. You have met them all already. Here they are in one place:

- **Fixed (Bengen, amended)**: is the initial rate at or below 3.5% (long horizon, expensive market), or offset by margins you have actually sized? If not, it is out: the cliff is too likely ([[fixed-inflation-adjusted-withdrawal]]).
- **Percentage, VPW, CAPE rules**: does the income under "stocks −50%" still cover the floor (the loss tolerance test, [[vpw]])? If not, it is out, unless a pension bridge or an annuity brings it back.
- **Bounded corridor (Vanguard)**: does the lower bound compounded over six years (about −14% in real terms) stay above the floor ([[floor-and-ceiling]])?
- **Guardrails**: is the cut floor (75 to 80% of the initial withdrawal) lined up with the real floor, and will the person running it **actually** take a 10% cut after three red years ([[guyton-klinger]], [[morningstar-guardrails]])?
- **ABW/TPAW**: is there a tool, someone to run it for life, and an acceptance that income will breathe by plus or minus 3 to 8% every year ([[amortization-based-withdrawal]])?

These tests take twenty minutes in a simulator, and they need only two outputs. The distribution of the spending actually delivered, year by year, settles the income and floor tests. A solver, which hunts for the limit value of a parameter at a given failure probability, settles the threshold tests. Those five questions usually knock out one or two families, and the decision gets simpler right away.

## Step 3: the profile-to-rule matrix

Among the families still standing, temperament and governance decide. The profiles as practice actually meets them:

| Dominant profile | Rule to reach for | Why |
|---|---|---|
| "I want my income; touch nothing unless you absolutely have to" | Modern guardrails with a floor ([[morningstar-guardrails]]) | Stable between breaches, safety you steer, thresholds written down |
| "I want to spend exactly right, and I accept that it breathes" | ABW/TPAW, g marked down, CAPE anchor ([[amortization-based-withdrawal]]) | Optimal spending, no cliff and no pile of gold, valuations built in |
| "As simple as possible, so my spouse can run it" | Vanguard's bounded corridor ([[floor-and-ceiling]]) | Second everywhere, no pathology, governance that fits on a postcard |
| "Floor already covered by a pension or an annuity, the portfolio is comfort" | VPW or a CAPE rule ([[vpw]], [[cape-based-rules]]) | Variability threatens nothing; generous spending, timed intelligently |
| "Any variation at all keeps me up at night" | Bengen amended to 3 or 3.25%, plus a ratchet ([[fixed-inflation-adjusted-withdrawal]]) | Perfect stability is bought with capital; the amendments limit the waste |
| "Leaving something behind matters most" | Cautious fixed plus a ratchet, or ABW with the bequest set as a parameter | A bequest you chose, instead of one left over |
| "More afraid of old age than of markets" | Safety first: an annuity, or delaying Social Security, up to the floor, then any rule on the rest ([[annuities-and-safety-first]]) | The risk you fear is covered by the right instrument, not by the rule |

Two things to keep in mind while you read it. First, the **temperament** that counts is the one you have in bad years, not the one the questionnaire records. People find themselves less flexible than advertised once the cut is real ([[flexibility-in-practice]], [[the-psychology-of-spending]]). When in doubt, take the steadier of the two rules on your shortlist. Second, **governance** is a first-order criterion, not a footnote. The best rule is the one still being applied twenty years from now, by the surviving spouse, in the middle of a bad decade ([[couples-and-family]]). A "suboptimal" rule that gets run always beats an optimal rule that gets abandoned.

## Step 4: the hybrids worth having, starting with the forgotten one

Three combinations keep showing up among serious practitioners, and the third is often the right answer for a FIRE plan.

**A guaranteed floor plus a free hand above it** (safety first): an annuity, a pension or linkers up to the floor, then whatever rule you like on the rest. Already covered ([[annuities-and-safety-first]]), and it works with every line of the matrix.

**The actuarial calculation with a smoothed display**: ABW as the annual reference number, with a corridor of plus or minus 5 to 10% around it for the budget you announce to yourself. You keep the optimality of one and the comfort of the other. Plenty of TPAW users run it exactly this way ([[amortization-based-withdrawal]]).

**A different rule per phase.** A FIRE plan has two lives ([[horizon-and-life-expectancy]], [[the-three-phases]]). The uncovered phase calls for steered caution: an amended Bengen at a prudent rate, or guardrails with a floor. The portfolio funds everything there, and the fragile window is open. The covered phase, once your pension covers the floor, frees your hand: VPW, a CAPE rule or a generous ABW on what is left. Writing "guardrails until the pension starts, VPW after" on day one is smarter than hunting for the single rule that serves both phases. The two problems are different, so the two answers should be. Plan the switch itself at the annual review of the year the pension starts ([[the-annual-review]]).

::: science What the differences between good rules are really worth
Put the ballpark numbers back in place, because they tell you where to spend your attention. Between two good, well-calibrated rules (guardrails with a floor against a marked-down ABW, say), the typical gap is 1 to 3 points of failure probability and plus or minus 5 to 8% of total spending: real, but second order. Against that: a 10% error on spending costs about 3 to 6 points of failure; a forgotten pension, 3 to 8 points; an initial rate of 4.5% instead of 3.7% in an expensive market, 5 to 10 points; abandoning the rule in a panic, more than anyone can compute. The pecking order for your attention follows. The **inputs** first ([[how-much-you-need]]), the initial rate next ([[valuations-and-cape]]), then whether the rule can really be run. The fine choice between the finalists comes last, calmly, because by then you can no longer go very wrong.
:::

::: figure hierarchie-attention
The first three bars get settled on the back of an envelope, in one evening of conversation. The fourth is the one that forums, comparison tables and months of hesitation are spent on. The fifth has no scale because it is not the same kind of thing: abandoning your rule at the bottom turns a bad decade into permanent failure, and it is the one line no simulation can do anything about.
:::

## The summary table for this part

| Rule | Failure | Income | Spending | Governance | Article |
|---|---|---|---|---|---|
| Amended Bengen | Possible (a predictable cliff) | Perfectly stable | The lowest | Trivial | [[fixed-inflation-adjusted-withdrawal]] |
| Smoothed fixed percentage | Impossible (capital) | Variable, smoothed | Good | Simple | [[fixed-percentage]] |
| Vanguard corridor | Low, honest | Glides of ±2.5 to 5% a year | Good | Trivial | [[floor-and-ceiling]] |
| Guyton-Klinger plus a floor | Low, honest | Stable, with −10% steps | Good | Medium | [[guyton-klinger]] |
| Risk-based guardrails | Steered (a written target) | Stable, with rare steps | Very good | Demanding (needs a tool) | [[morningstar-guardrails]] |
| VPW | Impossible (capital) | Variable | Very good, timed by age | Simple (a table) | [[vpw]] |
| CAPE rule | Impossible (capital) | Variable, self-smoothing | Very good, timed by prices | Simple (a formula) | [[cape-based-rules]] |
| ABW/TPAW | Structurally near zero | Breathes ±3 to 8% a year | The highest | Demanding (needs a tool) | [[amortization-based-withdrawal]] |
| Safety first (annuity to the floor) | Moved off the market | Guaranteed floor, plus the rest | Depends on the rule above it | One decision, then simple | [[annuities-and-safety-first]] |

## Step 5: calibrate, test, write it down

The final sitting takes an hour in front of a simulator. Set the candidate rule up with **your** parameters, not the ones from somebody else's blog post. Check the failure probability on the weight of the evidence: acceptable in the central case, tolerable on the broad sample, survivable in a lost decade ([[historical-vs-parametric]]). Then read the spending delivered in the worst quartile of paths, and make sure the life actually lived stays above the floor for stretches you could hold. Finally, run the candidate against the runner-up: same inputs, same market model, the rule the only thing that changes from one run to the next. That puts both finalists side by side on the decumulation frontier ([[withdrawal-strategies-overview]]). Then **write down** what matters: the rule, its parameters, its thresholds, the review date, the rule for the next phase, and the conditions for changing. You do not change rules in the middle of a crash. You decide it while calm, at the annual review, for reasons you wrote in advance. That one-page document is the final product of the whole of part IV, and [[building-your-plan]] folds it into the complete plan.

::: exemple The whole process, on a real case
Claire (51) and Idris (53): $1.7M, a floor of $45,000, comfort at $58,000, pensions of $24,000 a year starting in 13 years.

1. **Step 1**: an uncovered phase of 13 years, with the floor funded 100% by the portfolio, then 53% by the pensions.
2. **Step 2, admission**: the fixed rule at 3.4% passes (58,000 on 1.7M); VPW fails the loss test **before** the pensions (income under stocks −50% is $41,000, below the floor) but would pass after; ABW passes (g marked down, worst-quartile income $47,500); the corridor scrapes through.
3. **Step 3, temperament**: Claire wants stability, Idris wants precision; on governance, Claire is the one who will run it.
4. **Step 4**: the phase hybrid picks itself. Risk-based guardrails through the uncovered years (thresholds sized with the solver, a cut when central-case failure passes 13%, a raise below 1%, adjustments of plus or minus 10%, cut floor at 78%), then VPW on what is left once the pensions cover the floor.
5. **Step 5**: central-case failure 4%, broad sample 9% (late failures, in the covered years), two cuts at most in the worst quartile, never below $46,400. Signed, written, dated.

Total time: two evenings, one of them conversation rather than arithmetic. That is the right proportion.
:::

## The essentials

- You choose a **shape** of risk under the constraint of your floor, and it has to be runnable by the person who will really run it: temperament in bad years, then governance, then technique, in that order.
- Every family has one eliminating test (a rate at or below 3.5% for the fixed rule, "stocks −50% at or above the floor" for the proportional ones, the compounded lower bound for the corridor, an aligned floor for guardrails, a tool for life for ABW): twenty minutes.
- Hybrids are legitimate and often optimal, above all the forgotten one: two rules for two phases (steered caution while uncovered, proportional generosity once covered).
- The gaps between good rules are second order next to input errors and the initial rate: spend your attention in that order, and then choose calmly.
- The final product is **one written page**: rule, parameters, thresholds, review, next phase, conditions for change. Without it you do not have a strategy, you have an intention.

---

## Going further

- Early Retirement Now, Part 11 (scoring the rules) and the strategy Parts as a whole ([[the-ern-series]]); Morningstar, *The State of Retirement Income* (the annual comparison).
- Wade Pfau, *Retirement Planning Guidebook*: the most complete practitioner manual, and it covers both schools.
- The nine articles of this part, to go back to the detail of each candidate; then [[building-your-plan]] for the assembly, and [[the-annual-review]] for the life of the rule.
- Step 5's sitting, walked through one control at a time: [[using-the-fire-simulator]].
