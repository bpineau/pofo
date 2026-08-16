# Forward-looking expected returns (Morningstar, Vanguard, the investment banks)
<!-- source: rendements-attendus @ a3dedabf33ec -->

Every moving part of a withdrawal plan rests on one number nobody knows: the real return your assets will deliver over the coming decades. Two attitudes compete to fill it.

One is the rear-view mirror. "Stocks have returned 7% real since 1900, so I will use 7%." The other is the forward-looking approach. It estimates what **today's** prices, yields and fundamentals can reasonably promise. That is the trade of capital market assumptions (CMA), the long-run return estimates published every year by Vanguard, Morningstar, BlackRock, Research Affiliates, GMO, AQR and the big pension funds.

This page argues five things. The rear-view mirror is the worse of the two, especially for a retiree. Forward-looking forecasts are simple to build, and you will be able to redo the arithmetic yourself on the back of an envelope. Today's ranges carry a clear message. Their precision is mediocre, and still far better than the alternative's. And all of it turns into simulator settings without stacking prudence twice. By the end you will be able to produce and defend **your own** μ.

::: cle The reversal to make
An asset's past return is not its expected return. Often it is partly the **opposite**. An exceptional decade is largely multiple expansion, which is future return already spent. Bonds, meanwhile, print their expectation on the label (the yield to maturity), and stocks print half of it (the earnings yield, earnings divided by price, [[valuations-and-cape]]). Forecasting this way is not predicting. It is reading what the prices already say.
:::

## Why the rear-view mirror misleads, and doubly so for a retiree

The case against it rests on three facts.

**Fact one: the long past carries its own bias.** The "7% real from stocks" is a US number ([[anarkulova-cederburg]]) and it includes a tailwind that cannot repeat. Over the century the US CAPE went from under 20 to over 40, and real rates ended in a secular decline. Part of that historical performance was a one-off repricing, not an operating return. The rest of the world returned about 4.5% real over the same period, and that already counts only the survivors.

**Fact two: the recent past is worse still.** Anchoring on the last 10 or 20 years of your own funds is the natural reflex, and it is also what a simulator fitted naively to your history would do. But that window is peculiar. It is the very window that carried you to your target, so it is probably a friendly one. This is the reasoning that sent generations of savers into retirement on bubble assumptions. A good model defends against it by **blending**: parameters fitted to your portfolio are pulled toward a prudent world prior, and pulled harder as the horizon runs past the history available ([[making-monte-carlo-relevant]]).

**Fact three, and it belongs to the retiree: the error is asymmetric in time.** For a saver, being one point off on expected return moves the arrival date. For a retiree, the return of the **first ten** years decides the fate of the plan ([[sequence-of-returns]]), and that is exactly the horizon where valuation measures have their maximum predictive power ([[valuations-and-cape]]). Ignoring forward-looking information throws away the only light available on the decade that decides.

## How a forecast is built: the building blocks

Every serious CMA, whatever its method, comes back to the same decomposition, the building blocks (Bogle popularized it back in 1991): an asset's return = what it pays out + its growth + the effect of its valuation. Class by class, because you can run this calculation yourself.

**Bonds: read the label.** The expected nominal return of a bond, or of a bond fund of constant duration, is roughly its current **yield to maturity** (YTM), measured at its duration horizon. It is the only asset class whose expectation is nearly observable. Historically, the correlation between the starting YTM and the return realized over ten years runs above 0.9. For the **real** return, take out expected inflation (the breakeven), or read the real yield of inflation-linked bonds directly ([[inflation-linked-bonds]]). A 10-year government bond at about 3%, with a breakeven near 2%, leaves about 1% real; TIPS print their real yield as it stands, currently around 2%. One consequence matters: the bond expectation **moves** a lot with rates. The 2021 version (negative real rates, so a guaranteed negative real expectation) has nothing in common with the 2024 to 2026 version (about 1.5 to 2% real). A plan calibrated on "bonds return 5%", their historical nominal average inflated by the disinflation of 1982 to 2021, is calibrated on a world that no longer exists.

**Stocks: three blocks.** The real return splits into three terms. First the distribution yield, dividends plus net buybacks, about 2 to 2.5% for the S&P 500 today and about 3% in Europe. Then real growth in earnings per share, about 1.5 to 2% on a long trend, more while margins expand, though margins do not rise to the sky. Last the change in valuation, the disputed term: zero if you assume multiples hold, negative if you bet on a partial return toward more normal CAPEs. With 2024 to 2026 numbers, US stocks give about 2.25 + 1.75 + (0 to −1.5) = **2.5 to 4% real**. Non-US stocks, cheaper and paying out more, give **4 to 6% real**. You have just rebuilt the Vanguard and Research Affiliates tables to within a point.

**Cash and gold.** Cash follows the real policy rate, somewhere around 0 to 1% real. It moves with the cycle, and its long-run expectation sits near zero. Gold has neither coupon nor earnings, so its very-long-run real expectation is close to 0 to 1%. That is **not** what you own it for: it pays in expectation what it delivers in regime cover ([[gold-in-retirement]], [[defensive-assets]]).

::: figure briques-esperance
The five asset classes of this section, broken down with the article's numbers (2024 to 2026 conditions). The solid blocks can be read off today's prices; below the bar sits the valuation term, the only disputed one, and it carries the whole width of the range. The block count runs from three for a stock to zero for gold.
:::

::: exemple Building the μ of a 70/30 portfolio on the back of an envelope
Portfolio: 70% world equities (about 65% US), 30% investment-grade bonds. Stocks: weight 0.65 × 3% (US, the middle assumption) and 0.35 × 5% (the rest of the world), which comes to about 3.7% real. Bonds: YTM about 3.2%, expected inflation about 2%, so 1.2% real. Portfolio: 0.7 × 3.7 + 0.3 × 1.2, about **2.95% real geometric expected**. A parametric model takes μ as an **arithmetic** mean ([[arithmetic-vs-geometric-returns]]): add σ²/2, about 0.6 point at σ = 11%, so **μ ≈ 3.5%**. Compare that with your tool's default, and with what your own fund history suggests. The gap is the debate, and it is better settled on purpose than suffered.
:::

## Who publishes what, and what the ranges say

The forecasting landscape, with each house's method and its known bias:

| House | Main method | Style | Where to find it |
|---|---|---|---|
| Vanguard (VCMM) | Stochastic model, 10-year ranges | Centrist, publishes honest intervals | The annual "economic and market outlook" (free) |
| Morningstar (formerly Ibbotson) | Building blocks plus valuation | Centrist; feeds its annual SWR study | "State of Retirement Income" (free) |
| Research Affiliates | Mean reversion in valuations | Structurally cautious on expensive assets | Asset Allocation Interactive (free, interactive, country by country) |
| GMO | Aggressive mean reversion over 7 years | The most pessimistic on expensive assets; a record of underestimating bubbles, and their bursts too | Quarterly letters (free) |
| AQR | Theoretical risk premia plus valuations | Academic rigor; annual "Capital Market Assumptions" | Papers (free) |
| BlackRock, JP Morgan, the banks | Institutional CMAs over 10 to 15 years | Smoothed, built for asset allocation | "Long-Term Capital Market Assumptions" (free) |
| Pension funds (Dutch, Canadian and others) | Prudential regulatory assumptions | The binding bound: they **pay** if they are wrong | Public actuarial reports |

Three things to read in that table. First, the **ballpark numbers converge** in the 2024 to 2026 zone: US stocks 2 to 4.5% real depending on how much weight goes to valuations reverting, international stocks 4 to 6%, investment-grade bonds 1.5 to 2.5% real, cash about 0.5%. That puts a world 60/40 around **2.5 to 3.5% real geometric**. Second, the **dispersion that remains**. Between GMO and the most optimistic bank, the gap on US stocks routinely runs 3 points. Nobody "knows", and a plan that needs you to pick a winner to the decimal is a plan stretched too thin. Third, the **signal from the institutions that are on the hook**. The real discount rates of the large pension funds (2.5 to 4% real for diversified portfolios that include illiquid assets) mark a reasonable upper bound on what professionals will **promise**. An individual who types 6% real into a simulator is promising more than CalPERS (California's public employee pension fund, one of the largest institutional investors in the world).

Morningstar earns its own paragraph, because it closes the loop straight onto the withdrawal rate. Every year since 2021, "The State of Retirement Income" recomputes the recommended initial withdrawal rate (30 years, 90% success, a balanced portfolio) from its **forward-looking** returns. The series speaks for itself: **3.3% in 2021** (expensive markets, zero rates), **3.8% in 2022** (valuations had deflated), **4.0% in 2023** (bond yields restored), **3.7% in 2024** (stocks expensive again), then **3.9% in the 2025 edition**, published in December 2025. That last step up does not come from the markets, and Morningstar says so plainly: the house changed the method behind its return assumptions, which now combine its top-down read of the markets with the views of its equity analysts, company by company. On the old method its 50/50 would have come out at 3.6%. The number **moves**, and that is the deepest message of the exercise. The sustainable withdrawal rate is not a universal constant. It is a function of the conditions you enter on ([[valuations-and-cape]]) and of the method that turns them into a number, and a serious house takes responsibility for republishing it every year. Use their latest edition as a free second opinion on your own calibration ([[morningstar-guardrails]] for their full framework).

::: figure escalier-morningstar
The initial withdrawal rate recommended by "The State of Retirement Income", edition by edition (30 years, 90% success, a balanced portfolio), above a band giving Shiller's CAPE on September 30, the data cutoff of each edition. Four steps out of five follow the entry conditions; the fifth comes from a change of method.
:::

::: science How accurate are these forecasts?
The retrospective studies of CMAs (Morningstar's audits of its own archives, and the academic comparisons) return a mixed verdict. At ten years, forward-looking forecasts carry a substantial average error, plus or minus 2 to 3 points, but they clearly beat the naive rear-view mirror. Above all, they err **less often** in the dangerous direction, the one that overshoots after a bubble. The reliability ranking runs from best to worst: bonds (excellent, the YTM is close to a contract), cash (good), stocks (useful but noisy, the valuation term dominates at ten years and stays uncertain), alternatives (poor). For your plan the translation is simple: take the ranges seriously, not the point estimates. And remember that failure probability, too, is read as a range ([[failure-probability]]).
:::

## Feeding all this into a simulator without double-counting prudence

A full simulator offers four mechanisms that touch expected return, and the most common trap for the conscientious planner is to **stack** them:

1. **μ fitted to your history, then blended** toward a prudent world prior (μ 4.5% arithmetic, σ 13%, df 4), the prior weighing more as your history falls short of the horizon. That is already an anti-rear-view correction, and it applies without you typing anything.
2. **The same prior taken as it stands**, rewriting all three parameters with the world century's values alone (about 3.5% real geometric): the same prudence pushed all the way, giving nothing to your own data.
3. **The CAPE anchor**, which replaces the mean, and only the mean, with the estimate today's CAPE implies (about 1/CAPE for the equity block): this page's forward-looking correction, automated.
4. **The data models** (historical windows, block bootstrap, the century of 16 developed countries), which ignore your parameters and replay observed returns: the empirical bound, independent of any calibration.

The first three are competing ways to calibrate the same central model ([[making-monte-carlo-relevant]]). The fourth is not a calibration, it is a judge. Few tools offer all four. TPAW Planner derives its equity expectations from today's CAPE, the simulator behind this book blends parameters toward a world prior and shows the data models beside the central case, and plenty of others simply take a μ you type in, with no net.

The discipline is simple: pick one central calibration, own it, and read the others as bounds. Three roads are open. Trust the automatic blending, a reasonable default. Type in your own building-blocks μ, as in the worked example above. Or adopt the CAPE anchor, the automated equivalent of the building blocks for the equity part. But adopting the CAPE anchor, **then** cutting μ again by hand, **then** judging the plan only on the world-century model, counts the same prudence three times. The plan will demand years of surplus work for it, and that is a planning error of its own ([[one-more-year]]). Prudence gets a budget like everything else.

One last thing the forward-looking exercise does **not** replace. Volatility and tails (σ, df) cannot be read in valuations. They come from history and from the structure of the portfolio, and a serious model fits them on your funds' history ([[fat-tails]]). The expectation says where the road leads on average. σ and df say how much the ride shakes, and for a retiree the two count almost equally ([[arithmetic-vs-geometric-returns]]).

## Common mistakes in using them

**Mixing up conventions.** A "5%" can be arithmetic or geometric, real or nominal, gross or net, and published CMAs mix conventions from one house to the next (Vanguard publishes 10-year nominal geometric, AQR real, Research Affiliates real). Convert **everything** to real geometric before comparing, then to arithmetic (add σ²/2) for the μ parameter. The three reflex questions of [[arithmetic-vs-geometric-returns]] apply to the forecasters first.

**Overreacting to the latest edition.** CMAs move every year, but your plan should not roll with them. The right rhythm: reread them at each annual review ([[the-annual-review]]), act only if the landscape has changed regime, as bonds did from 2021 to 2023, and never on a swing of plus or minus 0.3.

**Forgetting that the future can be worse than every forecast.** CMAs are central expectations, and none of them "contains" 1929, Japan in 1990 or a war. That is the job of the other models (sequence stress, lost decade, the world century) and of structural protections ([[all-weather-portfolios]], [[flexibility-in-practice]]). The forward-looking expectation calibrates the center. It replaces neither the tails nor the margins.

**And the mirror image: permanent pessimism.** GMO has been calling US returns close to zero since 2013. Anyone who followed that to the letter missed the best decade in recent history. The lesson is not that "the pessimists are wrong", since their hour does come, as 2000 and 2008 showed. It lies elsewhere: never hand the **portfolio** to a forecast, hand it the **plan** (the rate, the margins, the expectations). The difference between those two uses is exactly the one drawn for the CAPE ([[valuations-and-cape]]).

## The essentials

- Past return is not expected return. The bond expectation is printed on the label (the real YTM). The equity one has to be built: payout + growth + the valuation term.
- The 2024 to 2026 zone, in real geometric terms: US stocks 2 to 4.5%, international 4 to 6%, bonds 1.5 to 2.5%, a world 60/40 at 2.5 to 3.5%. A simulator μ above those ranges either has a justification or needs a correction.
- Morningstar recomputes the recommended withdrawal rate on these foundations every year (3.3, 3.8, 4.0, 3.7 then 3.9% from 2021 to 2025): living proof that the sustainable rate depends on the conditions you enter on, and on the method that reads them.
- The precision comes in ranges (plus or minus 2 to 3 points at ten years on stocks), but it beats the rear-view mirror, above all in the direction that protects you: after a bubble.
- In a simulator, run one central calibration (blending by default, a manual building-blocks μ, or the CAPE anchor) and read the others as bounds. Do not count the same prudence three times, because it is paid in working years.

---

## Going further

- Vanguard, "Economic and Market Outlook" (annual); Morningstar, "The State of Retirement Income" (annual); JP Morgan and BlackRock, "Long-Term Capital Market Assumptions": the four free reads that cover the landscape.
- Research Affiliates, "Asset Allocation Interactive": expectations by country and asset class, recomputed continuously; the most instructive interactive tool of the lot.
- John Bogle, "Investing in the 1990s" (1991) and *Common Sense on Mutual Funds*: the original building-blocks decomposition ("Occam's razor").
- Antti Ilmanen (AQR), *Expected Returns* (2011): the reference book on the subject, exhaustive and readable.
- In this book: [[valuations-and-cape]] (the equity block in detail), [[arithmetic-vs-geometric-returns]] (the conversions), [[making-monte-carlo-relevant]] (how the model blends your history with the priors), [[risk-premia]] (why these returns exist and should persist).
