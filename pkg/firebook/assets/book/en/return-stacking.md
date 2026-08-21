# Return stacking, overlays and portable alpha: stacking the premia
<!-- source: return-stacking @ 9e7bc7d9eaff -->

Every diversifier hands the retiree a funding problem. Adding 10% of trend ([[managed-futures]]) or gold ([[gold-in-retirement]]) means selling 10% of stocks or bonds. You give up part of the premium that pays for the plan ([[risk-premia]]). That is the classic dilemma of the table of defenses: protection is paid for in expected return. **Return stacking** offers a third way. It puts moderate leverage inside funds built for it. You then hold the diversifiers on top of the stock-and-bond core instead of in its place.

The idea is not a forum fad. It is the retail version of **portable alpha**, an institutional technique forty years old. Corey Hoffstein (Newfound) and ReSolve Asset Management brought it back into the light, and a generation of "efficient core" and "stacked" funds carries it today. This article gives you the exact mechanics and the full arithmetic, financing cost included, the piece the marketing forgets. It also works through the lesson of 2022, how to buy the thing, and the sizing rules. By the end you will know whether your plan deserves one more layer, and above all which one.

::: cle The idea in one sentence
An index future posts only a fraction of its exposure as collateral. A fund can therefore hold 90% in stocks outright and layer 60% of bonds on top through futures: 150% of exposure for $100 invested, with the 50% above capital financed implicitly at the short rate. Holding $67 of that fund is the same as a plain 60/40. That leaves you $33 for trend or gold with nothing sold. Leverage here is not there to amplify a bet. It is there to fit more diversification into the same portfolio.
:::

## The mechanics, and the full arithmetic

The prototype is the **90/60** fund: 90% in stocks outright and 60% of government bonds through futures, a 60/40 at 1.5 times leverage. The "stacked" versions layer something other than bonds, say 100% stocks plus 100% trend, or 100% bonds plus 100% trend. Three pieces of arithmetic drive everything else.

::: figure stacking-expo
Three ways to invest $100: the leverage held inside the fund carries exposure beyond the capital invested, and the stacked version holds its 60/40 and its diversifiers at the same time.
:::

**Financing cost first.** Exposure above 100% is financed at the money-market rate, the implicit price of a future ([[leverage-and-margin]] for the mechanics). A stacked layer therefore does not earn its gross return but its return **minus cash**. Stacking bonds that yield 3% when the short rate is 3.5% destroys value. The same stack pays again once the curve has a positive slope. A stacked layer captures a premium above cash, never an absolute return. Hence the admission test for any layer: does it have a positive expected return, net of the short rate and of fees?

**Volatility drag next.** Leverage amplifies volatility, and volatility costs compound return ([[arithmetic-vs-geometric-returns]]). A 60/40 at 1.5 times leverage suffers more drag than a plain 60/40. As long as the two legs offset each other, with a stock-bond correlation that is negative or zero, the effect stays small next to the premium added. Which leads straight to the third point.

**Correlation last, the lesson of 2022.** That year stocks and bonds fell together. The 90/60 took about one and a half times the loss of the 60/40, on the order of −25% against −17%. Leverage creates no new defense: it enlarges whatever you give it. Stacking bonds on stocks therefore helps only in the regimes where bonds defend ([[market-regimes]]). Hence the best argument, in the withdrawal phase, for stacking an asset with a genuinely low correlation rather than duration alone. Trend is the natural candidate here.

::: science Where the idea comes from: institutional portable alpha
In the 1980s PIMCO launched StocksPLUS. The equity exposure runs through futures, which tie up little capital. The rest is invested in actively managed bonds. If those beat cash, the cost of the future, the fund beats the equity index. Pension funds generalized the principle under the name portable alpha, the idea of carrying a source of return on top of a beta.

The 2008 crisis exposed its limit: illiquid stacks financed short, and margin calls in the middle of a panic. The modern "return stacked" version took the lesson. It stacks only liquid instruments (index futures, government bonds, trend programs), with capped leverage and daily rebalancing inside the fund, and no margin call for the holder.

The published backtests point where you would expect (Newfound and ReSolve, plus the academic work on the levered 60/40). Over the long run a 60/40 at 1.5 times leverage beats the plain 60/40 on return, for a drawdown comparable to an all-stock portfolio. And adding a trend layer clearly improves the worst paths, exactly the geometry a retiree is after ([[sequence-of-returns]]).
:::

## What it changes for a withdrawal plan

Return stacking answers two precise situations, and you have to resist the urge to make it say more.

**Situation one: the diversification budget hurts.** You are convinced by the case for trend (0.2 to 0.4 of a point of safe withdrawal rate on the worst vintages), but cutting the equity sleeve costs you in the central case. An efficient core frees the room. The plan then holds the equivalent of its target allocation and its diversifier at once, and the improvement in the tails is no longer paid for in the median. This is the cleanest use, and the best documented.

**Situation two: capital is a little short.** With the plan unchanged, overall leverage of 1.1 to 1.3 raises expected return, and with it the sustainable withdrawal rate, at the price of deeper tails. The use is defensible for a young, flexible plan and dangerous for a tight one. Leverage also magnifies sequence risk: reread [[leverage-and-margin]] before you decide, because its five rules apply here in full.

And a non-situation: stacking does not turn a bad layer into a good one. Stacking a strategy whose expected return is zero after fees, like most of the "alternatives" in [[global-macro]], adds risk and friction for nothing. Stacking amplifies convictions that were already justified. It is not a source of return in itself.

## Buying it without kidding yourself

The fund range is young and thin, and that is the real limit of this chapter. Efficient core funds have been around the longest, the original 90/60 and its variants; the stacked versions that layer trend or gold are newer and still few ([[building-it-with-us-etfs]]). Hence three buying disciplines. **The vehicle's liquidity**: a small ETF with little in it can close, which for a retiree means a forced sale and a tax bill nobody chose. **The transparency of the stack**: demand to know exactly what is stacked, at what leverage, and financed how. A document that does not let you rebuild "X% of this plus Y% of that, minus cash" is a refusal to buy. **The currency**: check which currency the stack is built in and whether the exposure is hedged, because that is what decides the cash leg.

The do-it-yourself alternative exists: replicate the stack yourself with futures in a margin account. The verdict is the one that home-made trend gets. It is a job, with margin calls, quarterly rolls and a tax treatment that will not thank you, and it belongs to people who already do it for a living. The daily-leveraged fund, the 2x ETF type, is no substitute either. Its daily rebalancing erodes value in choppy markets, the beta slippage, which disqualifies it as a long-term building block.

::: astuce Test the stack before you buy it
This block can be tested before it is bought. That takes a simulator whose catalog carries reconstructed histories of efficient core funds and of stacked stock-plus-trend strategies, calibrated on the reference indexes; this book's data includes them. Run your current plan against its "levered core plus freed-up trend" version, then read the same three judges as always. The central case should improve a little. The 1966 and 1973 stress should improve clearly. The maximum real drawdown tells you the psychological price ([[reading-a-fan-chart]]).
:::

::: exemple Freeing up 33% of diversification without selling a share
Plan: $1M, $38,000 a year, 60% stocks / 40% bonds, 45 years. Stacked version: 67% in a 90/60 fund (60 of stocks and 40 of bonds in exposure), 18% trend, 10% gold, 5% cash. Total exposure about 133%, including a defensive layer of 33% that did not exist before. Typical reading in a simulator. The central case goes from 3.9 to 3.7% (the stacked premium, net of cash). The sequence stress goes from 5.4 to 4.6% (trend and gold work the long regimes without having disarmed the equities). The worst real drawdown goes from −34% to −29%. But replaying 2022 shows the bill for the leverage. The stacked core loses about 21% there, where the naked 60/40 loses 17%, and the whole plan gets away with about −10% only because trend worked very well that year. The clause written into the plan: "overall leverage capped at 1.35; the stacked layer is judged net of the short rate; if the core fund closes or changes its policy, back to the plain 60/40 within a month." Unless you accept that dependence on the diversification layer, do not sign the other lines.
:::

::: figure stacking-2022
The 2022 vintage of the example's plan, set out as one addition. The naked 60/40 loses 17%, the stacked core 21%: the leverage bill is four points, and it is quite real. The whole plan still gets away with −10%, because that core weighs only 67% of the book and the freed third went to work. The case for stacking is judged right there, in its most hostile year.
:::

## The sizing rules

They fit in four lines. The plan's overall leverage (total exposure over capital) stays under 1.5, and under 1.3 for a plan that is already tight or a horizon that is short. The stacked layer is reserved for instruments with a solid case and a low correlation (government bonds when their premium over cash is positive, trend, gold), never for bets. The leverage lives inside the funds, never in a personal margin account backed by the portfolio you draw on. And the whole thing is judged like any other building block, on the worst vintages in a simulator and not on the backtest in the brochure ([[simulator-traps]]).

## The essentials

- Return stacking puts moderate leverage inside funds built for it. It holds the diversifiers on top of the stock-and-bond core instead of funding them by selling that core. It is institutional portable alpha in retail form.
- The arithmetic fits in three lines: a stacked layer earns its return minus cash, leverage amplifies volatility drag, and it enlarges whatever you give it. In 2022 a 90/60 lost 1.5 times what a 60/40 lost, when stocks and bonds fell together.
- For a retiree, the clean use is to free up the diversification budget: an efficient core, trend and gold set free, better tails without sacrificing the median. The leverage-for-return use exists, but it makes sequence risk worse.
- The fund range is young. Demand liquidity, full transparency of the stack, and know which currency it is built in. Steer clear of daily-leveraged ETFs as substitutes, and of margin-account replication unless you already have the skill.
- Sizing: overall leverage at or below 1.3 to 1.5, layers reserved for documented premia, a written dismantling clause. Stacking amplifies good decisions and bad ones alike. It takes none of them for you.

---

## Going further

- Corey Hoffstein (Newfound Research) and ReSolve AM: the founding papers, "Return Stacking: Strategies for Overcoming a Low Return Environment" (2021), and the site [returnstacked.com](https://www.returnstacked.com).
- WisdomTree: the documentation of the Efficient Core funds (the original 90/60 and its variants).
- PIMCO: the story of StocksPLUS and the portable alpha literature; AQR, "Why Not 100% Equities" (Asness, 1996), the academic case for the levered 60/40.
- In this book: [[leverage-and-margin]] (the non-negotiable rules, all of which apply here), [[managed-futures]] (the layer stacked most often), [[arithmetic-vs-geometric-returns]] (volatility drag), [[all-weather-portfolios]] (the cousin without leverage).
