# The fixed percentage of the portfolio: indestructible but uncomfortable
<!-- source: pourcentage-fixe @ d406fa226094 -->

Every year, take 4% of the portfolio as it stands, not an indexed amount set years ago. That is the fixed percentage. It is the simplest strategy after Bengen's, and its exact mirror image on the impossible triangle ([[withdrawal-strategies-overview]]). It offers the most reassuring property in all of decumulation: ruin is mathematically impossible. But it charges the most visible price for it. Your standard of living tracks the market, point for point.

This article takes it apart piece by piece. Why it cannot run you out of money, and why that guarantee is worth less than it looks. The real shape of its risk, the lifestyle ruin that success tables never show. How to pick the percentage. The smoothing techniques that university endowments have refined for fifty years to make it livable, including the Yale rule, which transfers to a household unchanged. And finally, who it actually suits and what place to give it in a plan. It is worth the close look. Dominated in its raw form, it is still the base ingredient of half the modern rules: [[vpw]], [[amortization-based-withdrawal]] and [[floor-and-ceiling]] all descend from it.

::: cle The rule and its property
Each year the withdrawal is w × the current portfolio, with w set once and for all, 4% say. Ruin is impossible by construction: w% of something is never all of it. The portfolio heads toward zero without ever getting there. But read that again. It is the **portfolio** that never dies. Your income follows every lurch. The rule does not remove the risk. It moves the whole of it off the capital and onto daily life. This is the opposite pole from Bengen, the leftmost and highest point of the decumulation frontier.
:::

::: admin How to run it
- **The rate applies to the current portfolio, in full, with no memory.** No reference to the starting capital, none to last year's withdrawal. Two households holding the same portfolio on the same day withdraw the same amount, whatever their history. That is the exact opposite of Bengen, and it is what makes the paradox of two neighbors who retired a year apart disappear.
- **Indexation.** Nothing to index, and this is a classic implementation trap: if you compute on the portfolio in current euros, inflation is already in the answer, and indexing again would count it twice. The Yale rule is the exception, since its memory term does have to be re-indexed.
- **Frequency.** Annual. This is the one family where drawing more often makes life worse rather than better: taking a twelfth of the portfolio every month passes monthly volatility straight through to your income. If you want to smooth, smooth the portfolio and not the date, by drawing w × the average of the last twelve quarters.
- **Thresholds.** None, which is the whole point of the rule. The only parameter is w, and for once its bound is theoretical rather than calibrated: the portfolio's expected **real geometric** return. Above it, income erodes structurally. Below it, income grows slowly.
- **Floor.** The rule has none and cannot have one, since its income follows the portfolio all the way down. The floor has to come from outside, from a pension or an annuity. That is its admission requirement, not a refinement.
- **In your head.** R = w × portfolio. Yale rule: R = 0.7 × (last year's R, re-indexed to inflation) + 0.3 × (w × portfolio).
:::

## Why it cannot run you out of money, and what that is worth

The proof is one line. After the withdrawal and a return r, the portfolio is worth P × (1 − w) × (1 + r), a product of strictly positive factors, so never zero. The economic intuition is more vivid: the rule always sells a **fraction**, never an amount. When the portfolio halves, the draw halves too, and the pressure on the capital stays constant. The squeeze that kills the fixed rule, spending that climbs with inflation while the portfolio stalls ([[fixed-inflation-adjusted-withdrawal]]), is defused at the root.

That guarantee is worth less than it sounds. Look at what it actually promises: that the portfolio stays positive. It says nothing about the level. The portfolio can sit below 30% of its real starting value for a decade. Run the 1966 vintage under a fixed percentage and real income halves, then stays there for years. Above all, it does not promise that you can live on the draw. An income of w × (almost nothing) is still an income of almost nothing. Formal ruin gives way to a steady, perfectly legal impoverishment. The literature calls this **lifestyle ruin**: the portfolio survives, the life plan dies. Any honest verdict on the fixed percentage therefore rests on the distribution of the income delivered, never on the success rate, which is 100% by definition and completely hollow. That is criterion 3 of the scorecard ([[withdrawal-strategies-overview]]). A tool that reports nothing but a success rate has nothing to say about this family of rules.

Behind the guarantee sit two real virtues, and the rule deserves credit for both. It is **countercyclical** on the portfolio side. It draws little at the trough, which protects the recovery, and a lot at the top, which skims off the euphoria. That is the exact reverse of the fixed indexed rule, and the deep reason it leaves capital so robust. It is a genuine defense against sequence risk ([[sequence-of-returns]]). It also self-corrects when your assumptions turn out wrong. If returns disappoint for years ([[expected-returns]]), income adjusts continuously to the real world instead of piling up a silent debt until the cliff. Those two virtues are the inheritance every descendant of the rule tries to keep while taming its volatility.

## The real risk, in numbers: your income has your portfolio's volatility

Income is w × portfolio, so its volatility is the portfolio's. A volatility of 11% gives a standard of living that swings by plus or minus 11% in ordinary years and follows every drawdown to the bottom. Here are the ballpark numbers to internalize, for a global 60/40:

| Episode | Real drawdown | Your income |
|---|---|---|
| Ordinary correction (every 2 to 3 years) | −10 to −15% | −10 to −15% for 1 to 2 years |
| A 2008-style crash | −30 to −35% | −30% for 2 to 4 years |
| A hostile regime like 1966-1981 | −40 to −50% at worst, a decade underwater | real income cut by a third to a half for about 10 years |

The last row asks the real admission question: **can you live for ten years on 55 to 65% of your comfort level?** For a household whose floor sits at 90% of comfort ([[how-much-you-need]]), the answer is no, and the pure fixed percentage is inadmissible whatever its success rate. For a household whose pension already covers the floor, with the portfolio funding only the surplus ([[pensions-and-other-income]]), the answer can be yes, and the rule suddenly looks very attractive. That is the typical profile of the covered phase of a FIRE plan ([[horizon-and-life-expectancy]]).

The psychological reversal matters too. Under Bengen, the anxiety attaches to a distant, binary event: the cliff. Under a fixed percentage, it attaches to the next statement and to next year's income. Those two kinds of stress suit different people ([[the-psychology-of-spending]]), and neither is objectively smaller.

## Smoothing: fifty years of endowment engineering

The fixed percentage has a long-standing institutional user: foundations and university endowments. They are built to last forever, so they can never run out, and they fund stable budgets, so they have to smooth. Half a century of practice has produced techniques a household can lift straight off the shelf. They are the missing link between the raw fixed percentage and the modern rules.

**The moving average.** The simplest version. You draw w × the average portfolio over the last 12 quarters instead of the last data point. A 30% crash then arrives as three years of roughly 10% steps. Income volatility falls, very roughly, by the square root of the smoothing window. The cost: a lagging draw bites a little harder into capital at the trough, since you are drawing on an average that is still high. That leaves a very small chance of badly degraded paths. Smoothing sells back a crumb of the guarantee for a lot of comfort.

**The Yale rule (Tobin).** The standard among the large endowments, and remarkably elegant. The withdrawal is 70% × (last year's withdrawal, indexed to inflation) + 30% × (w × the current portfolio). This is exponential smoothing: each year, income travels only 30% of the way to its proportional target. The properties are exactly the hybrid you want. In the short run, income has the inertia of the fixed indexed rule, through its 70% of memory. In the long run, it comes back to the truth of the fixed percentage, converging on w × portfolio in four to five years. A 30% crash cuts income by about 9% in the first year and about 16% cumulatively in the second. The slope is livable, and the direction is honest.

**The corridor.** A third school: draw w × portfolio, but bound how far income may move from one year to the next. No more than +5%, no less than −2.5% in real terms. That is Vanguard's dynamic spending rule exactly, and this book gives it its own article ([[floor-and-ceiling]]). The corridor puts ruin back on the table, because a capped descent may fail to keep up with a collapse. It is a deliberate choice, a point in the middle of the frontier.

::: figure pourcentage-lissages
The same portfolio, the same crash, three ways of delivering it to the household. The raw percentage passes the shock through whole, in two years. The Yale rule spreads it over five. The corridor barely moves at all, and the last row shows what it is really doing: it funds that smoothness out of capital, so it borrows from the household of the next decade. None of the three removed the shock. They only chose who pays it, and when.
:::

All three teach the same lesson: the raw fixed percentage is not a finished rule but a **raw material**. Smoothed by an average, by memory or by a corridor, it gives you the rules in the middle of the frontier. Crossed with the remaining horizon, it gives you the actuarial family ([[vpw]], [[amortization-based-withdrawal]]).

::: science Choosing w: the geometric bound
Theory gives a clean bound for the percentage. Over the long run, a portfolio under a fixed percentage grows in real terms if and only if w stays below the expected **real geometric** return ([[arithmetic-vs-geometric-returns]]). Set w at 3 to 3.5% against a real geometric return of about 3.5 to 4.5% for a diversified portfolio ([[expected-returns]]) and median real income is flat or rising. At a w of 5 to 6%, it erodes as a trend. Each year then takes out more than growth puts in, and income follows the capital down without ever zeroing it. There is no cliff to fall off, so w can legitimately be more generous than a Bengen rate. A w of 4 to 4.5% is defensible where a fixed indexed rule would demand 3.25 to 3.5%. That is the dividend of self-correction. Endowment practice, around 4.5 to 5% smoothed for more aggressive portfolios, confirms the ballpark.
:::

::: figure borne-geometrique
With no randomness at all, assuming only that the geometric return holds, four rates and the income they deliver over thirty years. The tipping point comes down to two numbers: as long as w stays below g / (1 + g), the portfolio grows back faster than the draw takes out and income rises slowly. Above that, it erodes forever. The 6% rate starts at twice the 3% rate and drops below it at year twenty-two, which is exactly how long an ordinary retirement lasts.
:::

## Who it fits, and how to run it

**The profiles it suits.** Once smoothed, the fixed percentage works in three situations. First, a floor covered elsewhere, by a pension or an annuity, with the portfolio funding only the compressible part. That is the profile of a retiree whose pensions are already in payment. Second, naturally elastic budgets, dominated by discretionary spending, with a real ability to travel less in bad years ([[flexibility-in-practice]]). Third, a perpetual goal: passing on a real capital intact, endowment-style. The rule is contraindicated in the uncovered phase of a tight FIRE plan, the bridge years, when a high floor is funded entirely by the portfolio. That is where its lifestyle ruin bites exactly like the real thing.

::: astuce Testing the fixed percentage in a simulator
This is the easiest rule in the whole survey to program, and a spreadsheet is enough. In a decumulation simulator, look for a spending policy set as a percentage of the current portfolio, with w exposed as a parameter: "Spend % of portfolio (VPW)".

- **One spending policy at a time.** The percentage replaces the fixed indexed need and the flexible rules; it does not stack on top of them. Test the rules one after another, on the same plan and the same market assumptions.
- **Judge it on the income delivered.** Failure will come back at zero or close to it, and you now know that zero cannot be read on its own. The distribution of spending year by year is the only output that tells you whether the rule is admissible for you.
- **For the smoothed version, use the bounded cousin.** Few tools implement a moving average or Yale-style memory. A Vanguard-style corridor, "Bounded % of portfolio (Vanguard-style)", is often offered as it stands and does a similar job ([[floor-and-ceiling]]).
:::

::: exemple The same hostile regime, raw against smoothed
A portfolio of EUR 1.4M, w = 4%, so EUR 56,000. The real returns of 1973-1974: −35% over two years, then a limp recovery. Under the raw fixed percentage, income goes from EUR 56,000 to EUR 45,600 and then EUR 33,800 in two years, a fall of 40%. It climbs back at the market's pace and spends ten of the next twelve years below EUR 45,000. Under the Yale rule (70/30), it goes from EUR 56,000 to EUR 52,900 and then EUR 47,100, a fall of 16% over two years. It bottoms out near EUR 31,700 much later, and the descent leaves time to get organized. Same portfolio, same anti-ruin guarantee, but the smoothed version turns a free fall into a gentle slope. In a budget with 25% compressible, the first path is a crisis for the plan and the second is routine management. Smoothing is not a refinement. It is the admission requirement for the whole proportional family.
:::

## The essentials

- Withdrawal = w × the current portfolio. Ruin of the **capital** is mathematically impossible; ruin of your **standard of living** is not. Judge this rule on the distribution of the income delivered, never on its 100% success rate, which is completely hollow.
- Its income has the portfolio's volatility. The admission question is simple: can I live for ten years on 55 to 65% of my comfort level in a hostile regime? Yes if the floor is covered elsewhere, no in a tight uncovered phase.
- Its two deep virtues, inherited by every modern rule: it is countercyclical, drawing little at the trough, and it self-corrects when your assumptions turn out wrong.
- Smoothing makes it livable: a moving average, the Yale rule (70% memory plus 30% target, the endowment standard), or a bounded corridor (the Vanguard version, [[floor-and-ceiling]]). And w can be more generous than a Bengen rate, since the bound is the expected real geometric return.
- In its raw form it is a raw material more than a strategy. Its tamed descendants ([[vpw]], [[amortization-based-withdrawal]], [[floor-and-ceiling]]) occupy the middle of the frontier.

---

## Going further

- James Tobin and the endowment spending rule (Yale's own "spending rule"): the Yale Endowment's annual reports describe the version in force.
- Early Retirement Now, Part 10 (the fixed percentage against Guyton-Klinger) and Part 11 (the criteria) ([[the-ern-series]]).
- Bogleheads wiki, "Variable percentage withdrawal": the lineage from proportional to actuarial ([[vpw]]).
- In this book: [[floor-and-ceiling]] (the corridor industrialized), [[amortization-based-withdrawal]] (the percentage made aware of the horizon), and [[flexibility-in-practice]] (what living with variability really means).
