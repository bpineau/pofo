# CAPE-based rules: tying the withdrawal to valuations (ERN)
<!-- source: regles-cape @ fbd29d3d0b14 -->

If the valuation on the day you leave predicts how a vintage turns out ([[valuations-and-cape]]), why read it only once? CAPE-based rules take the logic all the way. The withdrawal rate becomes a **function** of the current CAPE, recomputed every year. You draw more when markets are cheap, because future returns are high, and less when they are expensive, because future returns are compressed.

This is the strategy Karsten Jeske (ERN) worked out for himself in Part 54 of his series, and has been running for real since 2018. In many ways it is where the proportional family ends up: a percentage of the portfolio whose w is no longer an arbitrary constant but an estimate of what the market can sustainably pay right now. This page gives the exact formula and its parameters, the values ERN recommends and why. It shows the double countercyclicality that makes its income steadier than the fixed percentage's. It lays out the real difficulties: the level of the CAPE is arguable, the income still moves, and running the rule takes faith in the bad years. It ends on implementation, up to the modern form where CAPE-anchored amortization becomes the clean version of the same idea.

::: cle The rule in one formula
Every year, withdrawal rate = a + b × (1/CAPE), applied to the portfolio as it stands. 1/CAPE is the earnings yield, the real equity return that today's prices allow ([[valuations-and-cape]]). b is the share of that return you consume (about 0.5). a is the base that does not depend on valuations (about 1.5 to 2%). At a CAPE of 20: 1.75 + 0.5 × 5 = 4.25%. At 33: 1.75 + 0.5 × 3 = 3.27%. At 12: 5.9%. The rate breathes with prices. It is the fixed percentage, finally aware of the world around it.
:::

::: admin How to run it
- **The rate applies to the portfolio as it stands**, like any percentage rule, with no memory of your starting capital and none of last year's withdrawal. What changes from one year to the next is the rate itself, recomputed on the day's CAPE.
- **Indexation.** Nothing to index.
- **Frequency.** Once a year: read the CAPE, add, multiply.
- **Parameters.** a ≈ 1.75 and b ≈ 0.5 for a 60-year horizon, a ≈ 2 and b ≈ 0.5 for 30 years. Unlike guardrail thresholds, these are **estimated**, not chosen: they come out of regressions on monthly data from 1871 to 2016. So they carry a confidence interval rather than a tradition. b is the fraction of the earnings yield you agree to consume, a the base that does not come from stocks.
- **A cap is mandatory.** The formula is linear in 1/CAPE and has no ceiling of its own, so it offers 11.5% to anyone leaving at the bottom of a crash. Write a rate cap, somewhere around 5 to 5.5%, exactly as you would write a floor. It is the one parameter the published formula does not give you.
- **Floor.** External, as for the whole proportional family.
- **In your head.** w in percent = a + b × 100 / CAPE. At a CAPE of 30, w = 1.75 + 0.5 × 3.33 = 3.4%. At 20, w = 4.25%. At 12, w = 5.9%, capped.
:::

## The logic: spend the estimated return, not a number carved in stone

Read the formula as an economic argument. At any moment a diversified portfolio offers a rough sustainable real return: the earnings yield on the equity sleeve, the current real yield on the bonds, plus the growth of earnings ([[expected-returns]]). The CAPE rule says to spend a prudent fraction of that estimate, and to redo the estimate every year. The a term collects everything the CAPE does not drive: the real growth of earnings (1.5 to 2 points) and the contribution of the other sleeves. The b × (1/CAPE) term lets the equity part breathe with prices. ERN's parameters (a = 1.75, b = 0.5 for a 60-year horizon with partial capital preservation; a = 2, b = 0.5 for 30 years) come out of regressions on monthly data from 1871 to 2016. They are the values that would have preserved the purchasing power of the capital across every vintage, tails included.

That one shift, spending an estimate instead of a constant, buys three remarkable properties.

**The full proportional inheritance.** The rule is a percentage of the portfolio as it stands: ruin of the capital is impossible, mistakes self-correct, and the draws are countercyclical ([[fixed-percentage]]).

**Double countercyclicality, the property that changes everything.** Here is the least obvious and most important point. Under a fixed percentage, a 30% crash cuts your income by 30%. Under a CAPE rule, the crash also pulls the CAPE down, which pushes 1/CAPE up, which pushes the rate w up. Income is w × portfolio, and its two terms now move in opposite directions. Put numbers on it. The portfolio falls 30% while the CAPE goes from 30 to 21. w rises from 3.42% to 4.13%, up 21%. Income ends up down 15%, not 30. The CAPE rule is a percentage that smooths itself, and it cushions exactly the drops that come from multiple compression, the most common kind. The smoothing has an economic reason behind it, unlike a moving average ([[fixed-percentage]]): you spend a larger share of the capital when the capital promises more. It works the other way too. In a euphoria (portfolio up 40%, CAPE from 30 to 38), w falls to 3.07% and income rises only 26%: the rule refuses to spend bubble gains. This is exactly the behavior you would want, and exactly the behavior nobody has on their own.

::: figure cape-contracyclique
Each thin curve joins the portfolio-and-rate pairs that deliver the same income. The CAPE rule almost rides one of them, while the fixed percentage, whose rate never moves, cuts across them all: that is the whole difference in how the two feel to live on. A 30% crash costs the first $8k of income and the second $15k; a 40% euphoria hands the first $13k and the second $21k, much of it nothing but borrowed valuation.
:::

**Time consistency.** The paradox of the two neighbors ([[the-4-percent-rule]]), who left a year apart with the same portfolio and draw different amounts forever, goes away. The CAPE rule pays the same withdrawal to anyone holding the same portfolio on the same day, whatever their history. That is the mark of rules with no dead memory: only the present state counts, as with ABW ([[amortization-based-withdrawal]]).

::: figure cape-depuis-1881
A hundred and forty-five vintages, and a rate that is never the same twice. Anyone looking for "the" safe rate gets the honest answer here: there is none. There is a price paid, and a return that follows from it. Notice the flaw in the formula along the way, plain to the eye in 1921: linear in 1/CAPE, it has no upper bound and offers 11.5% to someone leaving at the bottom of a crash. Nobody should follow it there. That is the case for the capped versions, and for ERN's own recalibrations.
:::

## The honest difficulties

**The level of the CAPE is arguable.** The whole debate of [[valuations-and-cape]] (accounting drift, buybacks, interest rates, a modern CAPE worth 3 to 8 points less than a naive comparison suggests) lands hard on a rule that consumes the level and not just the rank. With ERN's historical parameters, a CAPE structurally higher than it used to be gives structurally lower rates: prudent, and perhaps too prudent. Two defenses. You can use an adjusted CAPE, on a total-return basis, or work from the gap between the earnings yield and the real risk-free rate, the Excess CAPE Yield, which takes the level of rates into account. You can also recalibrate a, and ERN has published variants himself. And you have to accept that a rule built on an estimate inherits the uncertainty of that estimate.

**The CAPE of what?** The canonical rule reads the US CAPE, while your equity sleeve is global ([[building-it-with-us-etfs]]). A weighted world CAPE is published (Barclays, Research Affiliates) but harder to get at. Standing in the US CAPE stays conservative, since it is the most expensive of them, and it fits its 60 to 70% weight in the world indices. The approximation is acceptable as long as you know it is one.

**The income still moves, and running the rule takes nerve.** Self-smoothing cushions; it does not cancel. In a genuinely hostile regime, where prices and earnings fall together, the CAPE may not fall as far as prices do, and the income drops. You also have to run the rule in both directions. Drawing 5.5% from a badly beaten-down portfolio in the middle of 2009 takes real confidence in the formula. Yet that is where the formula is right, because that is where future returns are highest. Plenty of users cheat downward at the bottom, and destroy the very property that justified the rule. As always, the admission requirement is the floor ([[how-much-you-need]]): the CAPE rule fits if the income in the worst case still covers the floor, prudence margin included. To settle it, look at the distribution of the income actually delivered along the bad paths, not at the failure probability alone, which any proportional rule flatters ([[flexibility-in-practice]]).

::: attention The confusion to avoid
"CAPE rule" means tying the withdrawal to valuations, not timing the market with the portfolio. The rule never sells stocks at a high CAPE, never steps out of the market, never shifts the allocation. It sets the spending tap and nothing else. The confusion is common and fatal: timing the allocation on the CAPE destroys value ([[valuations-and-cape]]), while tuning the withdrawal on the CAPE creates it. Same indicator, two uses, opposite verdicts.
:::

## Running it, from a spreadsheet cell to CAPE-anchored amortization

**By hand.** The rule fits in one spreadsheet cell. Every January 1, read the CAPE (Shiller's site, or [multpl.com](https://www.multpl.com)), compute a + b/CAPE, multiply by the portfolio and divide by twelve. That is the year's paycheck. Governance is a percentage rule's: simple, auditable, and a spouse can run it ([[the-annual-review]]).

**The clean route: CAPE-anchored amortization.** The a + b/CAPE formula has an actuarial big sister. Instead of spending a fraction of an estimated return, you compute each year the payment that would exhaust the portfolio over the horizon that is left, at the return implied by the day's CAPE ([[amortization-based-withdrawal]]). The spirit is exactly the CAPE rule's, spend what valuations promise, with two refinements the linear formula lacks: the exact remaining horizon (a + b/CAPE quietly assumes a long, constant one) and the discounting of future pensions. ERN's rule is the back-of-the-envelope version, and it is brilliant; CAPE-anchored amortization is the full actuarial form. Same family, same information, finished machinery. Few tools offer it ready-made. TPAW Planner is the notable exception, since it amortizes by construction and lets the equity expectation rest on the day's valuation; the simulator behind this book combines the two the same way ([[under-the-hood]]). Everywhere else you lay the anchor by hand, replacing the historical average in the amortization calculation with 1/CAPE weighted by the equity share.

**The gentle route: a hybrid.** Keep an amended fixed withdrawal ([[fixed-inflation-adjusted-withdrawal]]), but let the CAPE set the initial rate on the day you leave, and let it trigger the big revisions after that. In practice you recompute the reference withdrawal when the CAPE changes zone: below 18, between 18 and 28, above 28. Half the benefit for a tenth of the variability, and an excellent transition for anyone coming from the Bengen world.

::: exemple Ten years of the CAPE rule, hard years included
Sofia, $1.5M, a = 1.75, b = 0.5. Year 1, CAPE 32: w = 3.31%, income $49,700. Years 2 and 3, a crash: the portfolio falls to $1.15M and the CAPE to 22, so w = 4.02% and income $46,200. That is down 7%, where the fixed percentage would have paid down 23%. Years 4 to 7, a soft recovery: CAPE 24 to 26, income $47,000 to $51,000. Year 8, euphoria: the portfolio climbs to $1.9M and the CAPE to 36, so w = 3.14% and income $59,700. The rule lets income rise, but half as fast as the portfolio, and the rest is set aside for later. The decade in one line: income between $46,200 and $59,700, plus or minus 13% around the average, on a portfolio that swung plus or minus 27%. The self-smoothing did its job, and not one parameter was touched. CAPE-anchored amortization would have traced a very similar path, with a rate that creeps up over the years as the remaining horizon shortens.
:::

## Who it fits

The CAPE-rule profile overlaps the VPW one ([[vpw]]): a covered floor or an elastic budget, and a taste for something simple enough to actually run. Two traits are its own. First, you have to believe the valuation logic, because anyone who doubts the CAPE will cheat at the first trough. Second, you have to accept an openly variable income, and in exchange you get the best-timed spending in the whole zoo: the CAPE rule spends the most when spending costs the least in future returns given up. For an early retirement in the uncovered phase, it needs the same pension bridge as VPW. And when you leave into an expensive market it has the elegance of starting cautious on its own, where every other rule asks you to remember to mark it down.

## The essentials

- The rule: rate = a + b × (1/CAPE) on the portfolio as it stands (ERN, a ≈ 1.75 to 2, b ≈ 0.5). It is the fixed percentage with w turned into an estimate of what the market can sustainably pay right now.
- Its distinctive property is double countercyclicality. In a crash driven by multiple compression, w rises while the portfolio falls, and the income takes only a fraction of the shock. The smoothing has an economic reason behind it; it is not cosmetic.
- The difficulties: the level of the CAPE is arguable (use an adjusted version, or recalibrate a), the US CAPE stands in for a global portfolio as a conservative approximation, and running the rule at the bottom takes nerve in both directions.
- Never confuse it with timing the allocation: the rule tunes the spending tap, not the portfolio.
- Running it: one spreadsheet cell is enough. The finished form is CAPE-anchored amortization, which pays each year the amount that exhausts the portfolio over the remaining horizon at the return valuations imply. The gentle version lets the CAPE set the initial rate and the big revisions of an amended fixed withdrawal.

---

## Going further

- Early Retirement Now, Part 18 (the flexible CAPE rules) and above all Part 54 ("Dynamic Withdrawal Rates Based on the Shiller CAPE"): the formalization, the regressions and the parameters ([[the-ern-series]]).
- The CAPE data: Robert Shiller's site; [multpl.com](https://www.multpl.com) for today's reading; Barclays and Research Affiliates for CAPEs by country.
- Kitces, "Should Equity Valuation Impact Safe Withdrawal Rates?": the practitioner's version of the debate.
- In this book: [[valuations-and-cape]] (the indicator and its critics), [[amortization-based-withdrawal]] (the full actuarial form), [[choosing-your-strategy]] (the final tradeoff).
