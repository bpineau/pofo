# The stock and bond allocation in retirement
<!-- source: allocation-actions-obligations @ 5ee9313fdf18 -->

"How much in stocks?" is the first question anyone asks about a retirement portfolio. Since Bengen, research has answered it with remarkable consistency. Plot the sustainable withdrawal rate against the equity share and you get a **plateau**. It is surprisingly wide and surprisingly flat, roughly 50 to 80% stocks, and it falls away on **both** sides. Too few stocks is more dangerous than too many. It collides head on with the instinct bank advice hands down: retirement means caution, and caution means bonds.

This chapter sets out the result and the mechanism behind it, which is why a plan **needs** growth to fund forty years of real withdrawals. It describes what happens at each edge of the plateau. It shows how your horizon, the coverage of your floor and your withdrawal rule move your best spot inside it. It takes stock of what recent research has added and what it has put in perspective, with Anarkulova-Cederburg and the 100% stocks debate. Then it unpacks the word "bonds" itself: the species, from the Bund to TIPS, what each one pays, what inflation and deflation do to each, what the ETF changes, and what zero-coupons are for. It ends with how to test all of it against your own plan, by re-simulating the allocation one notch at a time and reading the failure probability each time. The method that frames all this, designing by risks rather than by assets, is in [[designing-a-portfolio]]. This chapter opens the portfolio part. The ones that follow refine each block ([[bonds-in-retirement]], [[gold-in-retirement]], [[all-weather-portfolios]]) and the time dimension ([[glidepaths]]).

::: cle The central result
Across every dataset and every model, the sustainable withdrawal rate at a long horizon traces the same curve against the equity share. It stays low from 0 to 30% stocks, because bonds alone do not beat inflation plus withdrawals. It peaks and goes **flat** from about 50 to about 80%. It sags a little past 90%, where raw volatility starts to cost more than the premium pays. Two things follow. The serious mistake is sitting below the plateau, not above it. And inside the plateau, the fine choice between 60/40 and 75/25 barely moves the failure probability. It gets settled on other grounds: how deep a hole you can live through, and which withdrawal rule you run.
:::

::: figure allocation-plateau
The sustainable withdrawal rate as a function of the equity share (ballpark numbers, long horizon). Too few stocks and erosion wins; too many and volatility costs more than its premium pays. Between the two, a wide plateau that forgives imprecision.
:::

## Why the plan needs growth: the arithmetic of the plateau

Start with the mechanism. A 3.5% withdrawal plan over 45 years has to deliver **real** withdrawals, indexed, for decades. It needs an engine that durably produces more than 3.5% real a year, volatility drag and sequence included ([[arithmetic-vs-geometric-returns]]). Over long periods, high-quality nominal bonds return 1 to 2% real ([[expected-returns]]). A 20/80 portfolio therefore has a real geometric expectation near 2%. Arithmetically, it funds less than a 2.5% perpetual withdrawal. At 3.5% it runs dry **not** through bad luck but by construction, slowly, surely, with no crash anywhere. That is the erosion failure mode ([[reading-a-fan-chart]]), and it is the left edge of the plateau. It is why the Trinity grid already showed 25/75 failing one year in three where 75/25 held ([[the-trinity-study]]).

The right edge is subtler. From 80 to 100% stocks the expectation still rises, but two costs rise faster. The first is the drag: σ²/2, so going from 12 to 16% volatility costs about 0.6 points of compounding. The second, and the heavier one, is **sequence**. With no shock absorber at all, an early 50% crash hits the plan square in its fragile window ([[sequence-of-returns]]). Net result: the SAFEMAX of an all-stock portfolio sits slightly **below** that of a 70/30 in the US historical record, with far deeper worst cases. The plateau exists because those two forces, the need for growth and the cost of volatility, balance over a wide range. That is a piece of luck, because the problem itself forgives imprecision on this parameter.

Inside that frame, bonds have a precise and bounded job. They are not there to "pay" but to deliver three services. First, they **cushion** equity crashes: in a disinflationary regime their negative correlation lifts their price while stocks fall, and that is the shock absorber that worked so well from 2000 to 2021. Second, they **supply** the cash for withdrawals during the holes: you sell the bond that went up rather than the stock that went down, funding the withdrawal by selling the overweight ([[fixed-inflation-adjusted-withdrawal]]). Third, they **cap** the depth of drawdowns at a level your rule and your nerves can hold. Their Achilles heel is inflation, which breaks all three at once. That is the subject of [[bonds-in-retirement]] and the reason the rest of this part exists ([[market-regimes]]).

## Finding your spot on the plateau: the three dials that matter

Since failure probability barely tells 55/45 from 80/20, what should decide? Three things, in order.

**1. The drawdown you can live through.** This is the most concrete test. The real drawdown to expect once or twice a decade runs to about −0.55 × the equity share for a diversified portfolio: a 60/40 takes about −33%, an 80/20 about −44%, an all-stock portfolio −55% and worse. Two admission tests remain. Does your **rule** hold? That is the "stocks −50%" test of the proportional rules ([[vpw]]) and the floor under the guardrails ([[guyton-klinger]]). And do you hold? Your years as a saver prove nothing: taking −40% while you **live** off the portfolio is a different ordeal ([[the-psychology-of-spending]]). The maximum admissible equity share comes out of those two tests, not out of an optimizer.

**2. Floor coverage.** This is the big modulator, and it turns up everywhere. A floor covered by a pension, by Social Security or by an annuity ([[annuities-and-safety-first]]) frees the portfolio to move up the plateau. That is the TPAW logic: a discounted pension is a large implicit bond position ([[amortization-based-withdrawal]]), so the visible portfolio can carry more stocks. A floor funded entirely by the portfolio for twenty years pulls the other way, toward the middle and lower part of the plateau and toward dedicated defenses ([[cash-buffer]], [[glidepaths]]).

**3. The withdrawal rule.** Proportional and actuarial rules ([[fixed-percentage]], [[amortization-based-withdrawal]]) absorb volatility better by construction, because the withdrawal adjusts. They tolerate the top of the plateau. The fixed inflation-adjusted rule, which pushes the whole shock into capital, prefers the middle, 60 to 70%. Rule and allocation get chosen **together**. That is the whole point of the frontier that crosses the two ([[choosing-your-strategy]]).

::: science The 100% stocks debate (Anarkulova-Cederburg), put in its place
Cederburg's team made a lot of noise with "Beyond the Status Quo" (2023). In their world sample, an all-stock portfolio **diversified internationally** (a third domestic, two thirds international) beats the stock-bond mixes for a retiree. The reason: over a world century, nominal bonds get periodically destroyed by inflation, exactly when local stocks are suffering ([[anarkulova-cederburg]], [[market-regimes]]).

The honest reading is more nuanced. The result attacks, rightly, **domestic nominal bonds** as a long-term safe asset, not the idea of a shock absorber; in their portfolio, international equity diversification does the defensive work. ERN and Kitces pushed back on the violence of all-stock paths: real drawdowns of −60 to −80% in the tails, where no human admission test passes. They also point out how much rides on the treatment of the war years.

The takeaway you can use fits in two lines. The right edge of the plateau costs less than we used to say, **if** the stocks are diversified worldwide ([[international-diversification]]). And the real lesson is not "hold 100% stocks" but "the defensive sleeve must not be only nominal bonds". Hence linkers, gold and short duration, which the rest of this part covers ([[inflation-linked-bonds]], [[gold-in-retirement]], [[defensive-assets]]).
:::

## Past the ratio: what "stocks" and "bonds" have to contain

The ratio is not enough. What goes into the two sleeves moves the results as much as the ratio itself.

**The equity sleeve.** Global, broad cap, as simple as you can make it ([[building-it-with-us-etfs]]). Home bias is the first flaw to fix ([[international-diversification]]). Factor tilts (value, small caps, [[factors-in-retirement]]) are a second-order refinement, legitimate but optional. What is ruled out is concentration, whether in single stocks, in sectors or in one country: a retiree has no time left to average out an idiosyncratic risk.

**The bond sleeve.** This one takes the most care, and it has its own chapter ([[bonds-in-retirement]]). Three decisions in advance. **Quality** first: government bonds and investment grade only. High yield is a stock in disguise, defaulting exactly in the crashes and delivering none of the three services. **Duration** next: intermediate, 5 to 8 years, the standard compromise between cushioning power and rate risk. 2022 was a reminder of what long duration costs in an inflationary regime. And the **linked** share: linkers cover the service the nominals do not ([[inflation-linked-bonds]]).

**Rebalancing**, finally, is the third silent actor. It is what turns the allocation into a benefit: sell what went up to fund the withdrawals, buy back what fell when the bounds are hit. Practice converges on band rebalancing, plus or minus 5 absolute points around the target, rather than a strict calendar, and it uses the withdrawals themselves as the first tool for getting back to target. The exact frequency is third-order (ERN Part 39). The enemy is not the fine setting, it is failing to execute in a crash. So the rebalancing rule goes into the written plan, like the withdrawal rule ([[building-your-plan]]).

## A small atlas of bonds

So far "bonds" has named a single block. That block does not exist. Under the word live very different contracts, which do not pay the same thing, do not fear the same shocks and do not deliver the same services. Once the ratio is set, what the sleeve is made of is still open. Hence this atlas: the anatomy, the species, how each behaves under inflation and under deflation, and what the vehicle changes. The fine mechanics of prices and the full spec for the sleeve stay in [[bonds-in-retirement]].

**The anatomy.** A bond is a listed loan. An issuer, a government or a company, borrows an amount (the par value), pays a fixed annual rent (the coupon) and repays on a known date, the **maturity**. In between, the security trades, and its price moves opposite to market rates. The yield to maturity (YTM), printed on every fact sheet, says what the bond returns if you buy it at today's price and hold it to the end. It is the only asset class whose expected return is written on the label ([[expected-returns]]).

**Maturity and duration, two words for time.** Maturity is the calendar, the date of the last cash flow. **Duration** is the weighted average life of all the flows, coupons included, and duration is what measures the risk: one point more of yield costs roughly the duration in points of price, one point less pays it. Coupons hand money back along the way, so they shorten duration below maturity. A 10-year bond with a 4% coupon lives around a duration of 8, and so does a "7-10 year" fund. At the extreme, a zero-coupon pays nothing before maturity and its duration equals its maturity, which we come back to below.

::: figure duration-choc
One point of rates, five different pains. The price moves by about the duration for every point of rate change, in both directions. Money market barely feels it, the aggregate swings plus or minus 7%, the 30-year zero plus or minus 29%. Choosing your duration means setting that dial, not forecasting rates.
:::

**The yield curve.** Line up one issuer's yields from 3 months to 30 years and you get the yield curve. Its normal shape slopes up, because the long end pays a **term premium**, compensation for tying money up and taking rate risk, historically 1 to 2 points at the very long end ([[risk-premia]]). It flattens or inverts when the market expects rate cuts. For a retiree the curve is a price list. Maturity by maturity, it shows what the market pays for each extra year of rate risk, so it shows the exact price at which you are selling the calm of short duration.

## The species, issuer by issuer

Every bond yield is assembled from four blocks. The real risk-free rate, the rent on money. Expected inflation, the compensation for the erosion to come. The term premium, which grows with maturity. And the **credit premium** (the spread), which pays for the issuer's default risk. Each type of bond doses those four differently, and that is how to read the species.

**The core sovereigns.** The US Treasury is the risk-free yardstick in dollars, as the German Bund is in euros and the gilt is in sterling. These are the havens: when recession or panic hits, money goes there, and almost only there. In the 2024-2026 window, a core 10-year pays roughly 1 to 1.5% real.

**Sovereigns with a spread.** Inside a currency union, some governments pay more than the core. The Italian BTP, and to a lesser degree Spanish and Portuguese debt, slide from sovereign toward credit. The extra, 0.5 to 2 points depending on the year, pays for a real risk, the one that showed up in 2011-2012, when peripheral debt sold off in the middle of a crisis instead of cushioning. A euro aggregate holds some by construction, and that is acceptable. But a defensive core is judged on how it behaves in a crisis, not on its yield.

**The inflation-linked.** TIPS in the US, and their equivalents elsewhere: the contract is written in real terms, principal and coupons follow the price index, and the quoted yield is a guaranteed real rate. This is the only type that covers inflation by surprise, and conceptually the retiree's risk-free asset. The full case, with the breakeven, the lesson of 2022 and the guaranteed ladder, is in [[inflation-linked-bonds]].

**Investment grade credit.** Well-rated corporate bonds (BBB− and above, which is what the investment grade label means) pay the sovereign plus 0.8 to 1.5 points of spread. The extra is real but conditional: spreads widen in recessions, exactly when stocks fall. So IG cushions less well than government paper. It is respectable in moderate size, never as the core.

**High yield and the exotics.** Below BBB−, the gross spread climbs to 3 to 5 points, but defaults take back a good part of it and the correlation to stocks rises toward 0.7 in a crisis. This is an equity position in disguise. File it on the risk side of the portfolio if you insist on holding it, never on the defense side ([[false-defensive-assets]]). Same verdict for emerging market debt (wide-spread credit in hard currency, plus currency risk in local currency) and for bank subordinated debt and CoCos, which stack clause upon clause. Supranationals and agencies, on the other hand, are AAA quasi-sovereigns and perfectly respectable. The sorting test never changes: in the defensive sleeve, every holding has to be able to rise while stocks fall.

::: figure obligations-rendements
What each species pays, in real terms and in ballpark numbers (the 2024-2026 window, [[expected-returns]]). Going down the list you first add duration (the term premium), then credit (the spread). Every notch of extra yield is paid for, in rate risk first, in correlation to stocks second. The high end of the range is never free.
:::

## Inflation, deflation: what each species takes

Two textbook shocks organize the whole of bond behavior ([[market-regimes]]). The **deflationary shock** first, a recession or a panic, 2008 or 2020. Central banks cut rates, and high-quality nominals gain in proportion to their duration: about +8% for the aggregate in 2008, +25% and more for the long end. That is the regime where the bond sleeve earns its keep. Credit goes the other way, as spreads widen and defaults arrive: high yield lost around 25% in 2008, at the same time as stocks. Linkers come out roughly neutral, since the indexation works against them but falling real rates make up for it.

The **inflation shock** next, 1973 or 2022. Rates rise, and every nominal loses its duration times the rise. The short end has an escape hatch: its bonds mature quickly and get reinvested at the new rates, so two years in purgatory and the yield is rebuilt. The long end has no such door. It takes the whole blow, −30 to −40% for durations of 15 and up in 2022, then spends years recovering in real terms. The only species protected by contract is the linker, provided its duration is short, as 2022 taught the holders of long linkers ([[inflation-linked-bonds]]). Credit takes both punishments when inflation ends in a recession.

The figure below is the summary worth memorizing, and it fits in one sentence: no species wins in both shocks. Long duration is deflation insurance, the linker is inflation insurance, and the short end is a compromise that insures neither but survives both. That is the foundation of the "diversify your defenses" doctrine ([[defensive-assets]]). And the real answer to "how many bonds?" usually starts with "which ones?".

::: figure obligations-regimes
Five species, two shocks, no double winner (stylized nominal returns over the year of the shock, 2008 for deflation, 2022 for inflation). Long nominals shine on one side and sink on the other. Short linkers hold up more or less everywhere, the only ones that follow prices. Credit loses on both sides, through the spread in a recession and through duration in inflation.
:::

## ETFs, target-maturity funds, zero-coupons: the vehicle changes things

**What the ETF changes.** A bond ETF holds hundreds of positions and rolls them continuously to stay inside its maturity band. Its duration is therefore roughly **constant**: the "7-10 year" fund will still be a 7-10 year fund in twenty years, because it never matures. Three practical consequences. One, its expected return is its current YTM, held over its duration, no more and no less ([[bonds-in-retirement]] takes apart the "I hold to maturity, so I never lose" illusion). Two, duration becomes a dial you set notch by notch: 1-3 years, 7-10, 25 and up. "Buying bonds" no longer exists; you buy a specific duration. Three, on an upward-sloping curve a small **roll-down** bonus is added, the gain on a bond that slides down the curve as it ages and that the fund sells before maturity. The usual ETF choices remain: an accumulating share class, and above all a share class hedged into your own currency for foreign bonds, since stability is exactly the service you are buying ([[building-it-with-us-etfs]]).

**The target-maturity fund.** Dated-maturity ETFs (iBonds and their equivalents) hold a basket of bonds that all expire in the same year, then liquidate. Falling duration, an end value you roughly know: it is the individual bond in fund form, the tool for dated spending and the natural rung of a ladder ([[bond-ladders]]).

**The zero-coupon, the bond in its pure state.** No coupon, one cash flow at the end. Three properties follow. Its duration equals its maturity, the maximum possible for a given date. Its purchase yield is locked in with no **reinvestment risk**, the risk of having to put coupons back to work at rates that have turned mediocre. And its volatility is extreme, plus or minus 29% per point of rates for a 30-year zero. Hence two legitimate uses and one misuse. Use one, matching a dated expense: EUR 50,000 twelve years out is funded by a zero maturing then, nothing to reinvest, nothing to decide. Use two, the densest deflation insurance on the menu, the most duration per euro invested, the long sleeve of the all-weather portfolios pushed to its limit ([[all-weather-portfolios]]). The misuse is everything else, starting with reaching for yield: long STRIPS, Treasuries stripped into zeros by the market, returned −40 to −50% in 2022. Pure zeros are also thin on the ground in fund form, so very long duration is usually bought through a government bond ETF of 25 years and up, duration 18 to 20: most of the effect, without the exact tool.

::: figure duration-vehicules
Remaining duration is the weighted average life of the flows still to come, which is the price's sensitivity to rates. The rolling ETF renews its band and keeps its duration forever, while the dated fund and the zero-coupon watch theirs melt to zero on the day they hand the capital back (pure arithmetic, coupon and yield set at 3%, duration measured each year after the coupon).
:::

## Testing it on your plan: the allocation sweep

Everything above can be checked against your own plan, and the exercise asks for one discipline: vary the equity share and nothing else, re-simulate, and read the failure probability at each notch. Here is a typical session. Start from your real allocation. Raise the equity share from 40 to 90% in steps of 10, noting at each step the failure probability under the central model, then under the broad sample. You will **see** your plateau, usually flatter than you expected, and both of its edges. Then set the admissible share with your two tests, the drawdown you can live through and your withdrawal rule. Finally, look at what the surviving allocations do in the first decade, the one that decides everything ([[sequence-of-returns]]), and then in the real vintages. Look hard at what 80/20 did in 1966 and in 2000 before you sign. One caveat when you read it: the historical models replay **your** window, which has flattered stocks lately, while the broad sample judges the allocation over a world century. The tradeoff between the two is the same as everywhere else ([[historical-vs-parametric]]).

You still need a tool that takes your portfolio as it is. The US historical replayers (cFIREsim, FICalc) work in broad asset classes over long US data, which is enough to find the plateau. A simulator that reads the portfolio holding by holding shows you more, namely what the contents of the sleeves do to the verdict. Either way the drill is the same, and a plan entered once can be replayed as often as you like.

::: exemple A plateau made visible
The plan: EUR 1.5M, EUR 52,000 a year with guardrails, 45 years, a pension starting in year 16. The sweep gives, central failure then broad sample: 30/70 gives 11% and 19% (left edge, erosion), 50/50 gives 5% and 11%, 65/35 gives 4% and 9%, 80/20 gives 4% and 9%, 95/5 gives 5% and 10% (right edge, gentle because the rule is flexible). The 50 to 90 plateau is plain to see. Then come the admission tests. The guardrail floor holds up to about 75% stocks; past that, the worst quartile of delivered spending spends years below the floor. The couple themselves have been through −35% in 2020 without selling, and that is their limit. The decision: 70/30, with a bond sleeve of 20 in intermediate nominals and 10 in linkers, band rebalancing at plus or minus 5, written into the plan. The sweep took ten minutes. The conversation about the drawdown they could live through took an evening. The right proportions, once again.
:::

## The essentials

- The curve of sustainable rate against equity share is a wide plateau (about 50 to 80%) that falls away on both sides: the serious mistake is holding too few stocks (certain erosion), not too many (expensive volatility).
- Bonds are not there to pay, but to deliver three services: cushion the crashes, fund the withdrawals in the holes, and cap the drawdowns. Those services depend on the regime (inflation breaks all three, hence linkers, gold and diversification alongside).
- "Bonds" is not one block. Every bond yield is assembled from four blocks (real rate, expected inflation, term premium, spread), and each species doses them differently: long nominal is deflation insurance, the linker is inflation insurance, the short end is the compromise that survives both shocks, credit is a step toward stocks. In an ETF, duration is constant and set notch by notch: you do not buy "bonds", you buy a duration.
- Inside the plateau, three dials decide: the drawdown you can live through (about −0.55 × the equity share, tested against your rule and your nerves), floor coverage (a pension is an implicit bond and frees you upward), and the withdrawal rule (proportional tolerates more stocks than fixed).
- The 100% stocks debate mostly teaches that the defensive sleeve must not be all domestic nominal bonds, and that international equity diversification is itself defensive.
- Sweep **your** plateau in ten minutes (one equity share per notch, re-simulated, read under the central model and then under the broad sample), set the share with the admission tests, and write down the ratio, the contents and the rebalancing rule. Then move on to the time dimension: [[glidepaths]].

---

## Going further

- Bengen (1996) on the optimal allocation; Trinity (1998) for the grid by allocation ([[the-trinity-study]]).
- Cederburg et al., "Beyond the Status Quo" (2023) and the replies from ERN and Kitces: the 100% stocks debate in the original ([[anarkulova-cederburg]]).
- Early Retirement Now, Parts 19-20 (allocation and glidepaths) and Part 39 (rebalancing) ([[the-ern-series]]).
- In this book: the whole rest of part V, which fills the two sleeves block by block.
