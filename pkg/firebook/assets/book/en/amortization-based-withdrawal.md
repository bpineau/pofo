# Amortization-based withdrawal (ABW/TPAW): the actuarial approach
<!-- source: amortissement-abw @ 0abe7d717d46 -->

Ask an academic economist how to spend down a portfolio over a lifetime. You will not hear about Bengen, and you will not hear about Guyton-Klinger. You will get the Samuelson and Merton life-cycle model, and a prescription read straight off it. Every year, spend the actuarial annuity of your **total** wealth (the portfolio plus the present value of your future income) over your remaining horizon, at today's expected returns. The FIRE community has a name for that prescription: ABW, amortization-based withdrawal. It also has a finished implementation: Ben Mathew's TPAW (Total Portfolio Allocation and Withdrawal), a free planning tool that has become a Bogleheads favorite.

Recent research prefers this rule to any other. It is the only one that can neither run dry early nor leave its owner dying on a pile of gold. It adjusts in small continuous touches, where guardrails move in jolts of −10%. And it takes pensions, bequests and valuations into account from the start. It also asks a lot. Its income follows the market, and it does not run without a tool.

This article takes it apart in full: the mechanics, a mortgage run backwards and repriced every year; the theory it rests on; the four parameters that make the rule yours (return, horizon, bequest, spending tilt) and its pathologies; the head-to-head against guardrails; and how to actually run it.

::: cle The rule in one image
A mortgage computes the payment that pays off a principal in exactly so many years, at a given rate. ABW turns the formula around. Your portfolio is the loan life extended to you, and this year's withdrawal is the payment that would run it down to **exactly** zero over the years you have left, at today's expected return. That payment is recomputed every January 1, on the capital you actually have, with the horizon one year shorter and today's expectations. Whatever happens (a crash, a boom, inflation, one more year gone) is absorbed smoothly by the next calculation. It is a plan that replans itself, forever.
:::

::: admin How to run it
- **The rate applies to total wealth, not to the portfolio.** The portfolio, plus the present value of future pensions and income, minus the present value of the bequest you target and of any large planned expense. That is the deep difference with every rule before it, and it explains why ABW pays out more before your pensions start: it spends, in advance, wealth you already own and have simply not been paid yet.
- **Indexation.** Nothing to index, since repricing the annuity every year absorbs inflation. One strict implementation condition: work in real terms from end to end, both the return g and the rate you discount pensions at.
- **Frequency.** Annual, and no more often. The horizon shortens in years, so a monthly version of the rule would mean nothing.
- **Parameters.** g (the expected real return, marked down by 0.5 to 1 point), n (the horizon at a prudent survival quantile), the bequest you target and the spending tilt. No trigger thresholds: the rule is continuous, nothing gets breached, so there is nothing to calibrate on that side.
- **Floor.** The rule cannot exhaust the portfolio before the horizon, but it can very well deliver an income that is too low. The floor still has to be checked, against the distribution of the income delivered and not against failure probability, which is zero by construction.
- **In your head.** No, and the rule owns that: you need a spreadsheet or a tool. The formula is VPW's, applied to total wealth, R = W × g / (1 − (1+g) to the power −n), but each of the three letters takes a computation of its own first.
:::

## The mechanics, step by step

The formula is VPW's ([[vpw]]): withdrawal = W × g / (1 − (1+g)^(−n)). What changes is what the three letters contain, and every change buys a piece of realism.

**W is not the portfolio. It is total wealth.** ABW adds to the portfolio the **present value of every future flow**: the pensions ahead of you, Social Security included ([[us-healthcare-and-social-security]]), plus any scheduled side income, minus the one-off expenses you already know about. Take a couple of 47 with $1.4M in the portfolio and $20,000 a year of pensions from 67. Their total wealth is $1.4M plus the roughly $350,000 that pension is worth once discounted, and the annuity is computed on that total. The payoff is real. VPW's hand-built pension bridge becomes one term in the formula, and the withdrawal runs higher before the pension starts, because you are spending wealth that is already certain. This is the consumption smoothing economists have prescribed since Modigliani.

::: figure richesse-totale
A pension is an asset, even though no statement shows it. Adding it moves this year's withdrawal by 14%, and the bequest stops being whatever happens to be left and becomes a sum set aside on purpose. Both bars describe the same household at the same moment: only the definition of its wealth changes.
:::

**g is not carved in. It is today's expected return.** Where VPW freezes 5% real forever, ABW plugs in the current estimate, forward-looking expected returns ([[expected-returns]]). In its best form it anchors that estimate to valuations (equity g about 1/CAPE, [[cape-based-rules]], [[valuations-and-cape]]). In an expensive market the annuity is computed on 3% and the rule turns cautious by itself. In a market that has been cleaned out it is computed on 5.5% and dares more. That is the dividing line with VPW, already discussed: conditional accuracy against the robustness of a carved table.

**n is not "to age 100 by convention". It is the horizon you choose**, at a prudent survival quantile for the last survivor ([[horizon-and-life-expectancy]]). The bequest parameter, described below, can then stretch it or shorten it.

TPAW adds the final Mertonian touch. Total wealth contains a large implicit bond, because a discounted pension carries no market risk. So the **allocation** is reasoned on the total. Back to the couple: $1.4M at 70% equities, plus $350,000 of discounted pension, makes a whole that is about 56% equities. Someone in their forties with a pension ahead of them can therefore hold more equities in the visible portfolio than they think, at the same total risk. Allocation and withdrawal come out of the same calculation. Hence the name, Total Portfolio.

::: figure abw-1966
The 1966 vintage, the worst in the literature, lived under ABW and under a fixed withdrawal. On top, the two capitals: the fixed rule runs dry a year before the horizon ends, while ABW lands on zero exactly on the planned day, because the formula has been aiming at that day since January 1. Below, the two incomes: ABW starts 27% higher, breathes hard through the hostile decade (1974, the real market at −22.7%, the annuity repriced at −25%) and ends at $60k. **No cliff, no pile of gold, and $100k more of living delivered over the whole run.** The price is visible to the naked eye: that income moves, which is why the doctrine adds discounted pensions and display smoothing.
:::

## The four parameters that make it YOUR version

**1. The return g, and its margin of prudence.** The debate is an old one. Plug in the **central** expected return and the annuity is right in expectation, with every disappointment paid for by lower income later. **Mark it down** by 0.5 to 1 point instead and the annuity runs a little low in expectation, but the good surprises come back as raises. That is the direction to prefer. TPAW practice and this book's logic both lean toward a light markdown. It turns the distribution of income from "symmetric around the plan" into "a likely floor, plus good surprises", which is far better to live with ([[the-psychology-of-spending]]).

**2. The horizon n and the bequest.** Zero at the horizon is the default setting. To leave something behind, subtract from W the present value of the bequest you target. Leaving $300,000 real thirty years from now means amortizing W minus their present value, about $115,000 at a g of 3.2% real. The bequest becomes a **number you chose**, not an accidental leftover ([[spending-in-retirement]]). For longevity past n, the answer is VPW's: an annuity late in the run ([[annuities-and-safety-first]]), or an n already taken at the 90th percentile.

**3. The spending tilt.** TPAW lets you **tilt** the annuity. Spend more early, in the active years, while your health still allows for projects, and less at the end. Or the reverse. A tilt of −0.5% a year reproduces the retirement spending smile ([[spending-in-retirement]]) and lifts the first withdrawal by 10 to 15%. It is the Die With Zero argument, made adjustable and reversible.

**4. Display smoothing.** Raw ABW is recomputed every year, so its income moves every year. Serious implementations add a shock absorber: average wealth over twelve months, or move only halfway toward the new annuity. These are the proportional family's techniques ([[fixed-percentage]]), used here for the same reasons.

::: science Why the research prefers it
Three arguments come back.

1. **Coherence.** It is the only family derived from a decision model, Merton's life cycle (1969, the backbone laid out in [[lifecycle-theory]]), which applies to spending the utility framework laid out in [[deciding-under-uncertainty]]. The other rules are heuristics tested after the fact. Here every property has an explanation and every parameter means something.
2. **Dominance** on the modern criteria. In the comparisons (Morningstar on RMDs and their regulatory cousins, the Bogleheads and TPAW work, ERN on actuarial rules), amortization delivers the highest total consumption for a near-zero failure probability, with no cliff and no cascade of cuts. Its adjustments stay continuous and small, typically plus or minus 3 to 6% a year outside a crisis, against the plus or minus 10% steps of guardrails.
3. **No dead memory.** Like the CAPE rules, ABW depends only on the present state: no historical "reference withdrawal" fossilizes a decision made in year 1 ([[cape-based-rules]]).

The price is known. Income follows the market, and the rule demands a tool and a set of assumptions. That dependence on a model is its acknowledged Achilles heel ([[simulator-traps]]).
:::

## ABW against guardrails: the two finalists, head to head

The two winning families of the survey ([[withdrawal-strategies-overview]]) deserve a direct comparison, criterion by criterion:

| Criterion | [[morningstar-guardrails|Guardrails]] | ABW/TPAW |
|---|---|---|
| Day-to-day income | **Stable** between breaches (the comfort of a fixed rule) | Moves every year (damped, but moving) |
| Shape of the adjustments | Steps of ±10%, rare, after a breach | Small continuous touches (±3 to 6%) |
| Cliff, cascade of cuts | Possible if poorly bounded (a floor is mandatory) | Impossible by construction |
| Total consumption | Good | The highest in the field |
| Bequest | Residual, scattered | Chosen, set as a parameter |
| Pensions, valuations | Through the risk-based indicator (good) | Native to the formula (better) |
| Governance | Three written thresholds, one annual review | A tool to run, assumptions to own |
| Psychological profile served | "I want my income, touch it as little as possible" | "I want to spend right, and I accept that it breathes" |

The last row is the real dividing line, and it is personal, not technical. Guardrails sell income stability and charge for it in jolts that are rare but brutal. ABW sells optimal consumption and charges for it in variability that is constant but gentle. A household with a high spending floor and an anxious temperament will live better under guardrails. An elastic household that is comfortable with numbers will live better under ABW. The hybrid is legitimate too: ABW for the annual reference calculation, with a corridor of plus or minus 10% around it for the budget you announce to yourself. Plenty of TPAW users do exactly that.

## Running ABW, with an example

The year's calculation is four moves, the same in a spreadsheet as in a simulator. Value the portfolio **after** the tax a withdrawal triggers, under whatever account rules apply where you live. **Add** the present value of future pensions and income, and subtract the present value of the bequest you target. Set g to the expected geometric return, marked down by the margin you chose, anchored to valuations if you follow that school ([[cape-based-rules]], [[arithmetic-vs-geometric-returns]]). Then compute the annuity over the exact number of years left, without rounding the horizon to the nearest decade. A spreadsheet is enough to fund the current year. A simulator is for something else: running those four moves thousands of times over drawn futures, to see what the rule would have you live on before you commit to it.

Three demands on the tool, whichever one you use. ABW has to be a spending policy in its own right, with its parameters exposed: horizon, bequest, discounted future flows, assumed return. It replaces the fixed withdrawal, the flexible rule or the guardrails, and it does not stack on top of them, so comparing two rules takes two runs, on the same plan and the same market assumptions. Failure probability means nothing here, and that is the second demand. An actuarial rule cannot run dry by construction, and reading its 0% as a victory is the standing mistake of this whole part of the book. What counts is the distribution of the standard of living delivered year after year, checked against your spending floor ([[flexibility-in-practice]]). And one number on its own is worth less than a relative position. Plotting the candidate rules on a chart that plots failure probability against income variability shows at a glance what each one buys, and ABW usually sits in the corner where failure is near zero and variability is middling. Few tools go that far: TPAW Planner is the reference implementation, and a general-purpose simulator carries the equivalent policy, "Amortize over the horizon (ABW / TPAW)" ([[under-the-hood]]).

::: exemple One year of ABW, numbers in hand
Solène and Marc are 49. Their horizon runs to 99, so 50 years. The portfolio is worth $1.55M, the pensions $19,000 a year from 67 (present value about $310,000), and the bequest they target $200,000 real (present value about $90,000). The marked-down return g is 3.2% real.

The wealth to amortize is therefore 1,550 + 310 − 90 = $1,770,000. The 50-year annuity at 3.2% works out to about 4.05% of W, that is **$71,700**. Subtract the tax on the withdrawal and they live on roughly $66,000 net.

The next year opens two worlds. If the market does −18% (the portfolio at $1.21M), then W = $1,440,000 and n = 49, which gives an annuity of about $58,700 gross, down 18%. Amortization does not protect you from the shock, it only spreads out the consequences. What really cushions it is the matched share of total wealth, and here that share is only 17%. The same crash in a household whose discounted pension makes up half its wealth cuts the annuity by half as much.

If the market does +12%, the annuity rises to about $79,300, up 11%. Ten years of that give an income that breathes by 5 to 8% a year around a tilt you chose, with no cliff and no ratchet. And the $200,000 bequest builds itself quietly inside the formula.
:::

## The essentials

- ABW is the actuarial annuity, repriced every year. The withdrawal is the annuity on (portfolio + discounted pensions − target bequest) over the remaining horizon, at today's expected return. It is the economists' life-cycle model made executable, and TPAW is its tooled-up form, allocation included.
- Its properties are unique: it can neither run dry early nor leave you dying on a pile of gold, its adjustments are continuous and gentle, pensions, bequests and valuations are built in from the start, and it carries no dead memory. Hence the preference of recent research.
- Its four personal parameters: g (mark it down by 0.5 to 1 point, for a likely floor and good surprises), n (the 90th percentile for the last survivor), the bequest (a number you chose) and the tilt (the retirement spending smile, made adjustable).
- Its price: an income that breathes by 3 to 8% a year, and an acknowledged dependence on a model. The tool is mandatory, the assumptions need auditing, and display smoothing is recommended.
- Against guardrails, it is stability against optimality, a choice of temperament more than of technique, and the hybrid stays legitimate. The final call belongs to [[choosing-your-strategy]], after running both rules on the same plan and comparing the incomes delivered, never the failure rates.

---

## Going further

- Ben Mathew, the TPAW planner ([tpawplanner.com](https://tpawplanner.com), free) and the "Total portfolio allocation and withdrawal" thread on Bogleheads: the full doctrine and the tool.
- Merton, "Lifetime Portfolio Selection under Uncertainty" (1969) and Samuelson (1969): the foundations. Irlam (aacalc) for the modern numerical versions.
- Bogleheads wiki, "Amortization based withdrawal formulas": the formulas and the variants.
- Early Retirement Now on actuarial rules and on the Die With Zero critique (Part 60) ([[the-ern-series]]).
- In this book: [[vpw]] (the cousin with a carved table), [[cape-based-rules]] (the same spirit on the valuation side), [[morningstar-guardrails]] (the rival finalist), [[choosing-your-strategy]] (the verdict).
