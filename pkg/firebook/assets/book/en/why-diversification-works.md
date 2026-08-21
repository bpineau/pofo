# Why diversification works: the mechanics of the free lunch
<!-- source: pourquoi-la-diversification-marche @ 8f5cbd053361 -->

"Don't put all your eggs in one basket" is the oldest piece of financial advice in the world, and the most misunderstood. Most savers diversify the way people recite a prayer. They never see that the proverb hides a precise mathematical mechanism, the only one in all of finance that gives you something without asking for anything back: the free lunch Markowitz called "the only free lunch in investing". The mechanism has conditions and limits. Above all it has a consequence almost nobody talks about. A diversified portfolio that gets rebalanced can return more than the average of its parts, a bonus with the pretty name of rebalancing premium.

This page takes the mechanism apart piece by piece, with the minimum arithmetic (three formulas, nothing past high school). It puts a number on the bonus. It shows why the effect is even stronger in the withdrawal phase than in accumulation. And it ends with the limits, because diversification has its folklore too. When you are done, you will be able to say what each block of your portfolio brings to the group. And you will spot the fake diversification that clutters so many real portfolios.

::: cle The idea in one sentence
When two assets do not fall at the same time, the mix has a lower volatility than the average of their volatilities. Its average return, though, is exactly the average of their returns. Risk gets diluted, return does not. That is the famous free lunch. And since volatility costs you compound return ([[arithmetic-vs-geometric-returns]]), less volatility at the same average return means more money at the end, not just a calmer ride.
:::

## The arithmetic of the basket

Take two assets with the same expected return (say 5%) and the same volatility (say 20%). The 50/50 portfolio has an expected return of 5%, always, whatever the relationship between them. Its volatility depends entirely on the correlation ρ. At ρ = 1 (they move together) you get 20% and nothing has changed. At ρ = 0.5, 17.3%. At ρ = 0 (independent), 14.1%. At ρ = −0.5, 10%. The general formula says it all: the volatility of the basket is the square root of w²σ² + w'²σ'² + 2ww'σσ'ρ. The only lever is ρ, and every notch of correlation you take out removes risk for free. The σσ'ρ term even has a name, covariance, the real raw material of diversification.

::: figure correlation-vol
Two assets with the same expected return and the same volatility (20%), mixed 50/50: the volatility of the basket depends only on their correlation. The average return does not budge an inch.
:::

That gives a first practical conclusion, and it doubles as an audit. Diversification is not counted in holdings, it is counted in low correlations. Thirty equity funds are a single asset: their cross correlations run from 0.85 to 1.00, and the thirty names dilute almost nothing. Two US equity funds are, to the rounding, the same asset. Four well-chosen blocks, on the other hand (world stocks, long bonds, gold, trend following) show cross correlations between −0.3 and +0.3, several of them frankly negative. They do more work than the thirty funds ([[defensive-assets]], [[all-weather-portfolios]]). The marginal benefit collapses fast. Going from one block to four uncorrelated ones transforms a portfolio. Going from eight to twenty changes almost nothing and multiplies the fees and the chances of a mistake.

::: figure triangle-correlations
Correlations of monthly returns for seven dollar series, over a single common window from January 2001 to May 2026: four equity funds (world, S&P 500, the whole US market, developed markets ex-US), then the three other blocks (Treasuries 20 years and longer, spot gold, trend following). The trend leg is built from the real net asset values of trend-following programs, and from the fund itself since 2022.
:::

## The hidden bonus: the rebalancing premium

Diluting risk is only half the story. The other half joins two facts you have met separately elsewhere. First: the return that compounds your capital is the geometric return, and it runs at roughly the average return minus half the variance, what is called volatility drag. Second: diversification cuts the variance while leaving the average return alone. So it raises the geometric return. Booth and Fama (1992) named the extra "diversification return". A rebalanced portfolio returns more than the weighted average of its components' geometric returns, and the gap is roughly half the variance you saved.

An honest ballpark is 0.2 to 0.5 points a year for a classic portfolio. The bonus grows when you mix blocks that are both volatile and uncorrelated. The canonical example is gold. It has no real return of its own ([[risk-premia]]), 15 to 20% volatility, and a correlation near zero with stocks. Inside a rebalanced basket, that "sterile" holding still manufactures basket return. Which settles the apparent paradox of all-weather portfolios, where a block with no expected return improves the total ([[gold-in-retirement]]).

Rebalancing is the pump that harvests the bonus, what is called volatility harvesting. You sell what has gone up and buy back what has gone down, mechanically, by bands or on the calendar ([[the-annual-review]]). Two warnings belong here. First, the pump does not create the bonus, it collects it. Most of the gain comes from the variance you avoided, and a portfolio that is never rebalanced gives that saving back as one holding takes over. Second, rebalancing has an enemy: trend. In a market that climbs or falls in a straight line for years, selling the winner costs you. The bonus is harvested on round trips, not on straight lines. On real data, where the two are mixed, the disciplined version wins modestly but reliably. Above all it holds the risk profile the plan promised, which in the withdrawal phase is its real job.

::: exemple Shannon's demon, retiree version
The limiting illustration comes from Claude Shannon. An asset flips a coin every year: +100% or −50%. Its geometric return is zero, since one round trip leaves the capital where it started. Held alone, it builds nothing. Mixed 50/50 with cash at 0% and rebalanced every year, the basket compounds at roughly 6% a year. That return is manufactured out of two ingredients, neither of which earns anything. No real asset is that cartoonish, but the lesson is exact. Uncorrelated volatility, captured by disciplined rebalancing, is a raw material for return. This is volatility harvesting in its purest form, the formal version of "buy low, sell high", run by a rule instead of by talent.
:::

## Why the effect doubles in decumulation

In accumulation, diversification buys comfort and a little extra final return. In decumulation it pulls on a far more powerful lever, sequence of returns risk ([[sequence-of-returns]]). The sustainable withdrawal rate is not set by the average path, it is set by the worst paths. And that is exactly where diversification acts. It shortens the left tail and cuts both the depth and the length of real drawdowns. So it lifts the floor that sizes the whole plan. The effect shows up in any simulator. Move from 100% stocks to a basket of four blocks and median wealth at thirty years often slips a little, while the withdrawal rate that clears a 95% success rate goes up ([[reading-a-fan-chart]] to read both at once). A retiree does not diversify for the average. He diversifies for the 5th percentile, where his sleepless nights and his failure probability live.

::: figure panier-mediane-plancher
The same thirty-year plan, held by two portfolios, in the simulator's central model fitted on the four blocks' common history (1987-2026): on the left what capital is left in the middle of the cases after thirty years of a fixed 4% withdrawal, on the right the withdrawal rate that leaves 95% of the paths solvent. The basket sells a quarter of the median and buys 0.78 points of floor. An order of magnitude, not a universal result: other blocks, other weights or another window move both numbers.
:::

The same reasoning clears up a phrase that fools a lot of people, diversification across time. Spreading your withdrawals over thirty years exposes every dollar to a different market. Flexible withdrawal rules ([[choosing-your-strategy]]) are, at bottom, a way of diversifying your spending between the fat years and the lean ones. The portfolio and the withdrawal rule work the same risk from opposite ends.

## The limits, no folklore

**Correlations are fair-weather friends.** In a liquidity panic, nearly everything falls together for a few weeks (2008, March 2020). The correlation between stocks, real estate, credit and hedge funds heads toward 1 at the exact moment you were counting on them to diverge. Statisticians call this tail dependence. They model it with copulas, precisely because the ordinary correlation matrix, computed over the calm stretches, cannot see it. The list of what survives that test is short: high-quality government duration (in disinflationary crises only, as 2022 reminded everyone, [[market-regimes]]), cash, sometimes gold, and trend if the crisis lasts ([[managed-futures]]). Diversification does not remove short shocks, it tells regimes apart. That is already a great deal, but it is not immunity.

**Fake diversification, or diworsification.** Adding holdings correlated with what you already own gives you the feeling of a basket with the risk of a single block. Picture a world fund plus a US fund plus a tech fund plus ten American stocks. The test is one question per holding: "in which regime does this position win while the rest loses?" No answer, no diversification, just fees ([[risk-premia]] for the full version of the audit).

**And the psychological cost.** A genuinely diversified portfolio always contains one holding that disappoints. That is its signature: if everything rises together, everything will fall together. The gap to the neighbor who is 100% stocks peaks in the big bull markets. And that is where baskets get taken apart by their owners, right before they become useful. The defense is the same as everywhere else. The thesis for each block is written into the plan, and the judgment falls on the basket, never on one holding ([[the-psychology-of-spending]], [[building-your-plan]]).

## The essentials

- Mixing uncorrelated assets cuts volatility without cutting the average return. That is a theorem, not an opinion. The only lever is correlation, not the number of holdings (thirty equity funds = one asset).
- Less variance at the same average return means more geometric return. The rebalancing premium (0.2 to 0.5 points a year, more with volatile uncorrelated blocks like gold) is the forgotten half of the free lunch, harvested by disciplined rebalancing.
- In decumulation the effect is multiplied. Diversification works on the left tail and on sequence risk, so on the sustainable withdrawal rate, even when it pulls the median down a little. You diversify for the 5th percentile.
- The honest limits. Correlations head toward 1 in short panics, where only duration, cash, gold and trend survive, depending on the regime. And diworsification, correlated holdings stacked on each other, imitates the basket without owning the mechanism.
- A diversified portfolio always contains one disappointing holding, by construction. Anyone who does not accept that in writing will end up selling the block the day before it earns its keep.

---

## Going further

- Harry Markowitz, "Portfolio Selection" (1952): the founding paper, surprisingly readable.
- Booth and Fama, "Diversification Returns and Asset Contributions" (1992): the formalization of the diversification return.
- William Bernstein, "The Rebalancing Bonus" (Efficient Frontier): accessible numbers on the premium and on the conditions it needs.
- Portfolio Charts (portfoliocharts.com): the correlations and the regime behavior of dozens of model portfolios, visualized.
- In this book: [[risk-premia]] (what each block earns), [[arithmetic-vs-geometric-returns]] (volatility drag, the engine of the premium), [[all-weather-portfolios]] (diversification pushed all the way), [[sequence-of-returns]] (the risk all of this works on).
