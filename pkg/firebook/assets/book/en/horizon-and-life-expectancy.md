# Horizon, life expectancy, and 50-year retirements
<!-- source: horizon-et-esperance-de-vie @ 84680a193740 -->

How long does your plan have to last? The question sounds trivial. It has three traps in it.

The first trap: almost everyone works from the wrong life expectancy. They take the number at birth, or an average, when what they need is a prudent quantile, conditional on their age, and for the **last** survivor of the couple. The second trap: the intuition that twice the horizon means twice the caution is simply wrong. The link between horizon and withdrawal rate flattens out remarkably past 40 years, and understanding why changes how an early retirement looks.

The third trap: a FIRE horizon is not one uniform stretch. It has an uncovered phase, the bridge years before your pensions start, and a covered phase after them. The first one governs the risk. This page works through all three. How to pick **your** horizon from real life tables. What horizon actually does to the withdrawal rate, with numbers. And how to cross failure with the mortality of a couple, the one calculation that answers the question that matters: what is the probability of being alive and broke at the same time?

::: cle Three rules to keep
1) Plan on a prudent survival quantile (the 85th to 90th percentile for the last survivor), never on the average: half of all people outlive the average. 2) The safe withdrawal rate drops fast between 30 and 40 years of horizon, then almost stops dropping: a plan that holds for 45 years holds roughly forever. 3) Real longevity risk is not handled by stretching the simulator's horizon toward infinity. It is handled by owning assets that pay for as long as you live: pensions and annuities ([[annuities-and-safety-first]]).
:::

## Life expectancy, done right

**Mistake 1: using life expectancy at birth.** "Life expectancy is 80 for men, 86 for women": those figures include everyone who died young. You have already survived to today, so life expectancy **conditional** on your age is meaningfully higher, and it keeps rising as you age. Ballpark numbers from recent French tables (INSEE, rounded, average age reached; US Social Security period tables run a little lower, and none of the reasoning changes):

| At age... | Men | Women |
|---|---|---|
| Birth | ~80 years | ~86 years |
| 45 | ~82 years | ~87 years |
| 65 | ~85 years | ~89 years |
| 85 | ~92 years | ~93 years |

**Mistake 2: using the average.** A life expectancy is the center of a wide distribution. A man of 45 has roughly one chance in five of reaching 92; a woman the same age, one chance in five of reaching 95. Planning on the average means accepting a 50% chance of outliving your plan, which is a strange definition of prudence. Actuaries and good planners work from the **85th to 90th percentile of survival**.

**Mistake 3: one person instead of the couple.** For a couple, what counts is the time until the **last** survivor dies. The portfolio has to last as long as either of them is breathing, and the odds combine. For a couple both aged 45, the probability that **at least** one of them reaches 95 approaches 30%, one chance in five for her and one in ten for him. The 90th percentile for the last survivor sits somewhere around 98 to 100. That is the mechanism a couple's plan has to adopt, combining the two mortality tables year by year instead of picking one of them.

**Two calibration adjustments** finish the job cleanly. The first is the socioeconomic gradient. National tables average very different populations. In France, professionals outlive manual workers by 5 to 6 years (INSEE), and every rich country shows the same spread. The typical FIRE profile (education, income, wealth, access to care, no physically punishing work) sits at the top of that gradient, so add 2 to 4 years to the tables. The second is generational drift. Projected tables (TGH-05 and TGF-05 are the French insurers' set) build in the steady improvement in mortality. Insurers use them precisely to price annuities, and they are built on the observed mortality of hundreds of thousands of annuitants, extended by a projection of the improvement still to come. Someone in their forties in 2026 should be reasoning about how long people will live in 2070, not about who died in 2020.

::: exemple Working out a couple's horizon
Léa (44) and Sam (46), both professionals, plan to leave at 47 and 49. For the last survivor, the 90th percentile lands around Léa's 98th year. Add the socioeconomic correction and a little forward-looking caution, and the plan runs to **Léa at 100**, a **53-year** horizon. So the plan gets simulated over 53 years starting at 47. The rest is actuarial bookkeeping: each year of possible ruin is weighted by the chance that at least one of them is still there to live through it. The next section shows why that 53 is not twice as frightening as the 30 of a conventional retiree.
:::

## What horizon does to the withdrawal rate: the curve flattens

On to the second trap, the least intuitive and the most freeing. Here, in round numbers, is the rigid withdrawal rate that would have survived every US start year (Bengen's SAFEMAX, [[the-trinity-study]]) as a function of horizon, as the ERN series worked it out on monthly data ([[the-ern-series]]):

| Horizon | SAFEMAX (60 to 75% stocks) | Drop |
|---|---|---|
| 30 years | ~4.0 to 4.15% | (reference) |
| 40 years | ~3.5 to 3.6% | −0.5 point |
| 50 years | ~3.35 to 3.45% | −0.15 point |
| 60 years | ~3.25 to 3.4% | −0.05 point |
| Perpetuity (capital preserved) | ~3.25% | about 0 |

::: figure horizon-flatten
The withdrawal rate that would have survived every start year, by horizon (ballpark numbers, US data). Almost all of the tightening happens between 30 and 40 years; past that, the curve flattens toward an asymptote (about 3.25%, the portfolio's cruising real geometric return). A plan that gets through its first 30 to 40 years has all but become a perpetuity, so a very prudent horizon costs almost nothing.
:::

The conclusion is hard to miss. Almost all of the tightening happens between 30 and 40 years of horizon. Past that, the curve is nearly flat. Going from a 50-year retirement to a 60-year one costs next to nothing. Going from 30 to 40 is what costs. The mechanism behind the asymptote is clean once you see it. A draw of about 3.25% sits **below** the real geometric return of a diversified portfolio (about 3.5 to 4.5% historically, [[expected-returns]]). At that level, in the large majority of scenarios, the portfolio grows **faster** than you draw on it. After 30 to 35 years it has doubled or tripled in real terms, and the current withdrawal rate has fallen to 1 or 2%. A plan that has survived its first 30 to 35 years is no longer at risk. It has turned into a perpetuity. The extra horizon adds risk only in scenarios that were already bad at the start, and those fail inside the first 30 years anyway. The danger in a long plan is not lasting. It is **getting through the beginning**. Which lands us, once again, on the fragile window of sequence risk ([[sequence-of-returns]]).

Three practical corollaries follow from that asymptote:

**Leaving at 40 is not twice as hard as leaving at 60.** It is about 0.6 to 0.8 point of withdrawal rate harder (4% down to about 3.3%), which means a 30x multiple instead of 25x ([[how-much-you-need]]). The gap is real, but it is finite, and the young retiree's own margins go a long way toward covering it: employability, a pension still to come, flexibility ([[flexibility-in-practice]]).

**Uncertainty about your own longevity is a second-order problem for the portfolio.** Between planning 50 years and planning 60, the rate moves by less than 0.1 point. You can take the very prudent quantile at essentially no cost. That is excellent news: the most frightening unknown in the plan (how long will I live?) is the cheapest one to cover on the portfolio side. On the **spending** side of late life, it is another story ([[spending-in-retirement]]).

**Preserving capital is close to free past 45 years of horizon.** At around 3.25%, the portfolio holds its real value in most worlds. Aiming at preservation rather than depletion then costs a long plan only a sliver, whether you do it to leave something behind or simply for peace of mind. ERN calls this the convergence of capital depletion and capital preservation at FIRE horizons.

::: attention The opposite trap: too short a horizon
The mirror-image mistake exists, and it does more damage. Picture the 65-year-old who plans for "25 years, to 90", while his wife is 62 and has better than one chance in five of passing 95 on today's tables, and more than that on generational ones. For him the horizon sits in the steep part of the curve, where 25 to 35 years costs real money. Cutting it by ten years overstates the sustainable rate by half a point or more. The rule is simple: it is at **short** horizons that the choice of horizon is critical, and at FIRE horizons that it is nearly painless. Plan long, it costs nothing. Truncate, and it can cost everything.
:::

## Failure weighted by mortality: the calculation that puts the numbers in their place

The standard failure probability acts as if you never die. It assumes you are alive at 99 to watch the money run out ([[failure-probability]]). The real question is a joint one: **what is the probability of being alive and broke at the same time?** You answer it by crossing failure with the couple's mortality tables, a calculation sometimes summed up as alive, broke or gone. The method is actuarial and simple. In each year of each scenario you sort the household into three states: alive and solvent, alive and broke, or gone. The death probabilities come from national tables, INSEE and INED here, applied to both ages of the couple; a US reader can substitute the Social Security tables, which run a little lower, and nothing else in the reasoning moves.

The weighting is a relief, but a much smaller one than it is usually sold as. Measure it on Léa and Sam's plan. Ruin "40 years into the plan" arrives at 87 and 89, and it is only lived through if someone is still there. **Two couples in three** are. Raw failure for this plan comes out at 17.7%, failure actually lived at 14.1%. The weighting takes off **a fifth**, not the half or two thirds you often read. The reason fits in a sentence: a couple who left at forty-seven is almost always still alive when the plan gives way.

The discount grows with the age you leave at. The same failure profile, reread for a couple of sixty-five, would give 7.7% lived failure, a discount of more than half. So the weighting is above all a reading of the **timing** of your risk. If it takes almost nothing off, your failures are arriving early, and that is a serious signal ([[sequence-of-returns]]).

::: figure vivant-ruine-parti
The three states of Léa and Sam's plan, year by year, over 53 years ($1M, $33k a year in real terms, a pension of $14k a year from year 20). Empirical model from the JST panel (16 countries, 1871 to 2020, a 60/40 portfolio, 200,000 draws), with unisex Gompertz mortality fitted to the French tables and applied to a couple of the same age.
:::

None of which means you should size the plan on the weighted number. Size on raw failure over a prudent horizon: it covers the surviving spouse and the longevity surprises at once. Weighted failure then earns its keep on the **borderline** cases, and only at the margin. A discount of a fifth can tip a decision that was already close. It does not rescue a plan the raw reading condemns, and leaning on it is a good way to misjudge an exit date ([[one-more-year]]).

For the extreme longevity tail, the hundred-year-old case, the right tool is not the portfolio. It is the asset that pays **as long as** you live, calibrated to exactly that risk. Your state pension is already an inflation-linked life annuity (US claiming ages, and what they buy, are in [[us-healthcare-and-social-security]]). An annuity bought late, at 75 to 80, when pooling works hardest, is its technical complement ([[annuities-and-safety-first]]). Test the trade in a simulation: convert a slice of capital into an annuity at a chosen date, then compare the two risk profiles. Annuitizing often pushes the **average** failure probability up, because growth assets turn into floor income. What it crushes is the alive, broke and 95 tail, which is precisely the scenario you set out to remove.

## A FIRE horizon is not one thing: the uncovered phase

The last trap belongs to early retirements. A 50-year FIRE horizon really breaks into two very different regimes:

**The uncovered phase**, the bridge years from your exit to the day your pensions start, typically 15 to 25 years. The portfolio funds **everything**. This is where the fragile window sits, where the bulk of the net withdrawals happen, and where the risk concentrates. The current withdrawal rate is at its highest.

**The covered phase**, from the pensions to death. Pension income covers a substantial share of the spending floor ([[pensions-and-other-income]]). Net withdrawals collapse, and the portfolio drops back to a supplement and a reserve. In plenty of realistic plans the net withdrawal rate in this phase falls below 2%, which puts it out of structural danger.

That split explains something you see in simulation every single time. **Adding the pension to the model transforms the plan**, often cutting failure by a factor of two to four ([[ten-plan-wrecking-mistakes]]). The reason is simple: the pension shortens the **effective** horizon of the risk, from 50 years down to the uncovered phase alone. A well-built 50-year FIRE plan is really a 20-year bridge to a covered cruising regime, plus a reserve. Two design rules follow. Size the robustness on the uncovered phase, because that is the stretch the hostile vintages have to cross. And check the covered phase mainly for inflation risk, the one thing that threatens a pension and a spending floor 30 years out ([[inflation-and-withdrawal-rates]], [[inflation-protection]]).

::: terrain What it changes in your head
People a few years into FIRE say this constantly. Restate "my plan has to last 50 years" as "my plan has to get through 18 uncovered years, then the pensions take over the floor". It changes the mental load completely. The first version is an unimaginable void. The second is a bounded project with milestones. Every year crossed makes the crossing shorter. The right psychological unit for steering a plan is not a whole life, it is the crossing: the years you spend on your own capital ([[the-psychology-of-spending]], [[when-to-worry]]).
:::

## The essentials

- The right horizon target is the 85th to 90th percentile of survival for the **last** survivor, conditional on your ages and corrected for the socioeconomic gradient (plus 2 to 4 years for the typical FIRE profile). For a couple in their forties, that means planning to about 100, a horizon of 50 to 55 years.
- The rate-horizon curve flattens: about 4% at 30 years, about 3.5% at 40, about 3.3% at 50 to 60 and in perpetuity. A plan that survives its first 30 to 35 years has become a perpetuity. The risk in a long plan is its beginning, not its length.
- The corollaries: taking the very prudent quantile costs almost nothing at FIRE horizons; truncating the horizon is dangerous mostly for **short** retirements; and preserving capital is nearly free past 45 years.
- Crossing failure with the couple's mortality tables answers the real question, the alive-and-broke one, but the discount stays modest for an early exit (a fifth on our example plan, 17.7% raw failure against 14.1% lived). Size on raw failure, settle the borderline cases on the weighted number, and treat the longevity tail with annuities rather than with more capital.
- A 50-year FIRE plan equals a 15 to 25 year uncovered crossing plus a pension-covered phase: size the robustness on the first, the inflation defense on the second.

---

## Going further

- INSEE: the French mortality tables ([insee.fr](https://www.insee.fr), "tables de mortalité"), and its studies of the life-expectancy gap by social category; the projected TGH-05 and TGF-05 tables for the insurer's view. In the US, the Social Security Administration period life tables and the Society of Actuaries annuity tables play the same two roles.
- Early Retirement Now, SWR Series parts 1 and 2 (SAFEMAX by horizon, depletion against preservation) and part 56 (annuities and Social Security inside the plan) ([[the-ern-series]]).
- Moshe Milevsky, *The 7 Most Important Equations for Your Retirement*: the accessible actuarial formalization, including Gompertz's equation for longevity, Fibonacci's for how long capital lasts, and the logic of annuitization.
- A simulator that weights failure by mortality owes you two answers: which tables it uses, and whether they are projected. How this book's simulator does it is in [[under-the-hood]].
