# Managed futures and trend following: the diversification that works in a crisis
<!-- source: managed-futures @ b597d0e63250 -->

In the table of defenses ([[defensive-assets]]), one slot was still open: **long regimes**. These are the bear markets and inflationary stretches that drag on for years and wear out the ordinary shock absorbers. The cash buffer runs dry, bonds can fall alongside stocks, and gold can sleep through the whole thing. The holder of that slot is the strangest defensive asset of them all: managed futures, meaning **trend following**. These are systematic programs that trade dozens of futures markets (equity indexes, rates, currencies, commodities), buying whatever is going up and shorting whatever is going down.

The result has no equivalent in the rest of the toolkit. It is the only defensive asset with a positive long-run **expected return**, and its best years are precisely the worst years of everything else (2008, +21% for the pure-trend index; 2022, the year that broke everything, +27%, its record since 2000). The price is just as unusual: complexity, real fees, wide dispersion between managers, and multi-year "winters" when the strategy disappoints while stocks fly. It is the hardest test of patience in the portfolio.

This article covers the mechanism and the evidence (a century of data), the structural reason behind its crisis profile, the industry and its indexes, the rise of replication (DBMF first among them), the retiree's case, how to buy it cleanly (the delicate part), the traps, and the size.

::: cle The idea in one sentence
Big market moves are not instant. They **last**: an initial underreaction, a gradual acceleration, then capitulation, over months and sometimes years ([[market-regimes]]). A rule that simply follows the direction of the last few months therefore captures part of every big move, up or down. When a crisis is long, that rule ends up mechanically positioned on the crisis side of it. Trend forecasts nothing. It is simply on the right side of whatever persists, and disasters persist.
:::

## How it actually works

A typical trend program runs a few simple steps. It tracks 50 to 100 liquid futures markets (S&P, Bund, oil, gold, euro-dollar, wheat and so on) and measures the trend in each one with simple signals: a moving average, or the return over the past 3 to 12 months. It goes long the markets trending up and short the ones trending down. It sizes every position inversely to its volatility, which is vol targeting: each market then contributes the same amount of risk, and the leverage built into futures makes that possible without borrowing. And it revises continuously.

**A closer look at the signals.** Three families do most of the work. Moving-average crossovers first: buy when the 50-day average crosses above the 200-day. Time-series momentum (TSMOM) next, the academic version, which just looks at the sign of the past 12 months' return. Breakouts last: buy the 100-day high, the school of Richard Dennis and his Turtles. All three measure the same thing, persistence, and their signals overlap heavily. The big shops combine them into an **ensemble**: several signals at several speeds, fast (1 to 3 months), medium (3 to 6), slow (9 to 12), to smooth the entries and the exits. Speed is the real design choice. A fast program turns early, which helps in short shocks but multiplies the false starts. A slow program eats the reversals but rides the big waves better. The ensemble diversifies that choice the way it diversifies everything else.

**The cash works too.** One detail decides how you read a fact sheet: a futures program ties up only 10% to 20% of its capital in margin. The rest sits in Treasury bills, and their yield adds to performance in full. A trend fund returns cash plus the trend premium minus fees. With short rates at zero (2015-2021), a program earning a 3% gross premium and charging 1% showed about 2%. With short rates at 3% to 4%, the same program shows 5% to 6% without changing a thing. Part of the disappointment of the winter, and part of the recent glow, is nothing more than that piece of arithmetic. Any comparison across eras, or between a fund and its index, has to be made in excess return, above cash.

::: figure mf-cash-prime-frais
The same recipe, in two short-rate regimes. The three bricks are drawn separately, positive above zero and negative below, with the net set beside them. From one column to the other the program has not moved an inch: the gross premium and the fees are the same, and only the yield on the collateral has changed. That is why the winter and the recent glow are read first in the short rate, and why the honest comparison is the one in excess return alone.
:::

Three properties follow from the recipe. First, **symmetry**: making money in a lasting decline is as natural as making it in a rise, which is unique among your holdings. Second, it works across asset classes. In 2022 most of the gains came from short bond positions and long commodity and dollar positions, not from equities; the strategy hunts the trend wherever it happens to be ([[bonds-in-retirement]]). Third, **controlled risk**: vol targeting aims at a constant 10% to 15% volatility, a risk profile close to equities but steered.

**A century of evidence.** Few anomalies are this well documented. Moskowitz, Ooi and Pedersen ("Time Series Momentum", 2012) establish it across 58 markets back to 1965. Hurst, Ooi and Pedersen ("A Century of Evidence on Trend-Following Investing", AQR) rebuild it back to 1880: a positive return in **every** decade, including 1929-1939, 1970-1979 and 2000-2009, with a correlation to stocks and bonds close to zero. The explanations hold up in theory: behavioral underreaction to news (anchoring), hedging and risk-management flows that amplify moves already under way, institutional constraints. None of that should vanish quickly. The premium is well known and heavily arbitraged, and it has probably compressed; careful modern estimates put it at 2% to 4% real and gross for a diversified program, before fees ([[expected-returns]]).

**Crisis alpha, and its condition.** Why does the strategy shine in crises? Because big crises are trends. Through the eighteen-month decline of 2008, trend was short equities and long bonds from the autumn on. Same pattern in 1973-1974 and in 2000-2002. In 2022, the broadest trend of recent history, it was short bonds and long energy and the dollar. Pure trend finished the year at +27%, the wider CTA indexes between +14% and +20%, while the 60/40 lived through its worst year in a century ([[market-regimes]]). The condition is symmetric: the crisis has to **last**. The flash crash of 2020 is the textbook counterexample. One month down, a V-shaped recovery, and trend had no time to turn: −5% to 0%, neither defense nor disaster. Hence its exact place in the table. It is the starter for **long** regimes, and a useless backup against short shocks, which are the job of duration and cash ([[defensive-assets]]).

::: figure trend-smile
The trend "smile": its return against the return on equities, over rolling 12-month windows (stylized profile). Both tails pay, because big crashes and big bull markets are both trends. The dip in the middle is where the false starts live, those directionless markets that cost money. And 2022 is the reminder that the smile is multi-asset: that year the gain came from rates and energy, not from stocks.
:::

::: science The retiree's case
For the withdrawal phase, three facts matter. One. Allocation studies converge (ERN's Part 63 on momentum, the AQR and Man work on blends, Portfolio Charts on the versions a retail investor can buy): adding 10% to 15% of trend to a stock-and-bond portfolio lifts the safe withdrawal rate of the worst vintages by roughly 0.2 to 0.4 of a point, and clearly shortens the real drawdowns. The effect lands exactly where retirees die, in the regimes of 1966-1981 ([[the-trinity-study]]). Two. The improvement has a clean origin, with no hidden leverage and no disguised beta, just decorrelation with a positive expected return, the only partial free lunch on the defensive menu. Three. One warning outweighs all the rest. These results rest on indexes, or on programs charging reasonable fees. Run the same strategy at 2 and 20 with a mediocre manager and the case turns negative. Here more than anywhere, the vehicle is the thesis.
:::

## The industry, its indexes and its engines

Trend following is an industry born in the 1970s: the CTAs, commodity trading advisors, a US legal status that stuck to the whole category even though these programs trade far more than commodities. It has concentrated around twenty or so big quant shops. AHL (Man Group), Winton, Aspect, Lynx, Transtrend, Campbell and Dunn are among the oldest, each running billions in systematic programs refined over decades. To first order they all do the same thing: follow the trend across a broad universe. To second order they all differ, in the exact universe, the speed, the way risk is built. Which is why the indexes matter, since they measure the collective instead of one manager.

**The three benchmarks worth knowing.** The **SG Trend Index** (Société Générale) tracks the ten largest pure-trend programs, equally weighted and reconstituted every year. It is net of real fees and has been published daily since 2000. It is the yardstick of the category, and the judge of any vehicle that calls itself trend. The **SG CTA Index** widens to the twenty largest CTAs, trend or not, systematic macro and multi-model included. The **BTOP50** (BarclayHedge) captures the industry wider still, with history back to 1987. All three tell the same story on the same dates: 2008 and 2022 at the top, 2011-2019 flat. Using them well is a matter of matching. Check a pure-trend fund against the SG Trend, and a product that replicates the whole industry against the SG CTA or the BTOP50.

::: figure trend-annees
A quarter century of the SG Trend Index, year by year (ballpark figures, net of fees). The peaks land in the years when everything else breaks: 2008, 2022, and 2014, the year of the oil counter-shock. Between 2011 and 2019, the winter: nine years for a cumulative return close to zero while stocks tripled. Both halves are part of the same contract.
:::

## Replication, and the DBMF case

The category long had a distribution problem: the best programs were sold as hedge funds at 2 and 20 (2% flat, 20% of performance), out of reach for an individual investor or ruinous for one. **Replication** picks that lock. Instead of paying researchers, you watch the industry and copy it cheaply. Two schools exist. **Position-based** replication (return-based) regresses the recent returns of the big CTAs on a handful of very liquid futures, infers the aggregate positioning of the moment, then holds it directly. **Rule-based** replication runs a simple, transparent trend model, presented openly as an approximation of the industry's common core.

**DBMF, the textbook case.** Launched in 2019 by DBi (Dynamic Beta investments), the iMGP DBi Managed Futures Strategy ETF, ticker DBMF, became within a few years the largest US managed-futures ETF. Its recipe illustrates the first school. Every week, a regression over a window of about 60 days estimates the aggregate portfolio of the big CTAs in the SG CTA Index, then replicates it with a dozen very liquid futures (S&P, Treasuries, gold, oil, euro, yen and so on). The thesis has a name, **fee alpha**: copy the industry's average position for 0.85% flat and no performance fee, and you hand the holder back the 3 to 4 points that the 2-and-20 structure used to take. The bet is not to beat the managers but to beat their **net** result, which a cheap copy pulls off most years. The limits mirror the recipe. The regression window lags: in a violent turn, the fund is still carrying the day-before-yesterday's positioning. It copies the industry average, never the best manager. And its reference is the SG CTA, not the SG Trend, so its profile is a little less pure trend. The approach has spread. KMLM takes the other school, a rule-based model, single-speed, over 22 markets with no equity indexes, which gives a rougher and more diversifying profile. Others add carry on top of trend, and both schools keep showing up in new fund ranges. For the buyer the check never changes: compare to the right benchmark, insist on a flat fee with no performance fee, and know which school you are buying, positions or rules.

Trend can also be bought **stacked** on top of a stock-and-bond core, through return-stacking funds, when there is no room left for it in the allocation ([[return-stacking]]).

## What trend is not

Three confusions come up over and over, and they are expensive.

**It is not long volatility.** Both "win in crises", but not in the same ones. Long vol ([[long-volatility]]) buys options, bleeds a premium every year and explodes upward in flash crashes; 2020 was its finest hour, and the year trend did nothing. Trend wins in **long** crises, costs nothing in expectation, and misses one-month shocks. Different regimes, different premiums: the two holdings in the table of defenses are not interchangeable ([[defensive-assets]]).

**Its zero correlation is an average, not a state.** At any given moment the program is positioned, and its correlation right then follows those positions. Short equities in late 2008, it was negatively correlated. Long equities in 2021, positively. That is **conditional** correlation, and the famous zero shows up only once you average across regimes. The practical consequence: never judge the holding on six months of observed correlation, because that is positioning noise.

::: figure trend-correlation
Rolling 12-month correlation between a diversified trend-following program and the S&P 500, read at each month end from 2001 to 2026. The average over the period is +0.04, but one month in two falls outside the plus or minus 0.2 band: in late 2008 the ribbon drops to −0.42, in late 2021 it climbs back to +0.50. The trend leg uses real NAVs of trend-following programs over the whole period, and the fund's own NAV from 2022 on.
:::

**Its loss profile is the opposite of its reputation.** People picture a crisis asset, and assume it must be dangerous day to day. The truth is the reverse. Day to day, trend is a stream of small frequent losses, the false starts (whipsaws, signals that reverse the moment they fire), and its gains are rare and large. Year by year, its distribution leans the right way (positive skew), where equities lean the wrong way. The psychological consequence is real: the holding annoys you often and saves you rarely, the exact opposite of an equity holding. Knowing that in advance is half the discipline.

## Buying it cleanly: the delicate part

This is the hardest block in the book to buy well, and it is worth saying so plainly.

**What exists.** First, systematic trend funds run by the big quant shops: the historical programs packaged as daily-dealing funds, diversified pure trend. Their flat fees run around 0.7% to 1.5% a year, sometimes with a performance fee on top, which you skip whenever a flat-fee-only share class exists. Then, since 2019, the replication funds described above, by positions or by rules, at lower fees and with no performance fee. DBMF and KMLM are the best-known US-listed examples, and the same two schools now appear in fund ranges on both sides of the Atlantic ([[building-it-with-us-etfs]]). The control point is universal: compare the vehicle to its benchmark, the SG Trend for pure trend, the SG CTA or the BTOP50 for industry replication. A good vehicle tracks it within a few points. A bad one is doing something else under the same name.

**What you avoid.** Multi-strategy alternative funds sold as equivalents (usually disguised beta and fees), certificates and structured products written on CTA indexes (issuer risk and opaque margins), and 2-and-20 CTAs reachable only through exotic wrappers. Avoid do-it-yourself replication too, meaning running your own trend model on futures. It is possible in theory and a full-time job in practice: the daily execution discipline across 50 markets is not a retirement project ([[the-psychology-of-spending]]).

**Where it sits.** A systematic futures fund is rarely tax-efficient, so it belongs wherever your own rules tax it most lightly, and it competes with gold for the same shelf space. That is one more argument for splitting the hostile-regime slot between the two ([[gold-in-retirement]]).

::: astuce Test the size before you buy it
This block can be tested in a simulator, but the data gets in the way first. Trend earns its keep in long regimes, and the long regimes are old. Yet the deepest public index in the category only goes back to 1987, and the funds you can actually buy are far younger. An honest A/B test therefore needs history covering 1973 and 2000, ideally 1929. With no real series, you have to rebuild one: run a trend model over long price histories, then rescale the volatility you get onto the category's indexes. That is a backcast, a simulated past. Ask any tool where its series comes from before you believe its verdict; the reconstruction behind this book's data carries trend backcasts built exactly that way.

The test itself runs like the one for gold. Put 10% of trend against the same slice in bonds, then read the sequence stress, the lost decade, the broad sample and the inflationary vintages. What you should see: a central case barely moved, shorter tails, and above all a maximum real drawdown that pulls back on the mediocre paths ([[reading-a-fan-chart]]).
:::

## The traps, in order of lethality

**Quitting during the winter.** This is trap number one, by a wide margin. Trend goes through **dry seasons** that last years. The 2011-2019 stretch, the CTA winter (directionless markets, endless reversals, roughly 0% cumulative while stocks tripled), made most holders capitulate, right before 2022. The pain of the gap peaks there: nine years paying for insurance while the neighbors get rich. The defense is the same as everywhere, only stronger. A written size, a written thesis ("this sleeve loses or goes nowhere most years; it exists for 1973, 2008 and 2022"), and a ban on judging it over a stretch with no real regime in it ([[building-your-plan]]).

**The wrong vehicle.** Dispersion between trend managers is enormous, tens of points apart in a crisis year, because the signal, the universe and the speed differ from one shop to the next. Hence index replication or the very large diversified programs, and the check against the SG Trend. Do not hold a small CTA. Hold the industry, or its core.

**Getting the role wrong.** Two mirror-image mistakes. Buying trend for the returns guarantees disappointment, because the net expectation is modest. Selling it after a winning crisis year to take profits is just as wrong: that is **exactly** the moment to rebalance into the equities that got destroyed. The insurance has just paid out. You cash it and rearm, on the same rebalance-to-fund-your-paycheck logic as gold ([[gold-in-retirement]]).

**Sizing it on enthusiasm.** After 2022 the temptation to go to 25% or 30% is real. But the strategy is still a systematic program with clean tails and long winters, and it carries genuine model risk, since the premium can compress further. The sensible size is 5% to 15%, like gold, and for the same epistemic reasons ([[all-weather-portfolios]]).

::: exemple Ten percent of trend, put to the test
Plan: $1.5M, $51,000 a year, Vanguard corridor, 45 years. Variant A: 65% stocks / 25% bonds / 10% gold. Variant B: 65 / 20 / 7.5% gold / 7.5% trend (a flat-fee trend fund, checked against the SG Trend). Typical verdicts: failure goes from 3.8% to 3.7% in the central case (nothing, as it should be), and from 5.9% to 5.1% under the sequence stress. On the lost decade, the median real drawdown of the bad paths goes from −31% to −26%, and the replayed vintages of 1973 and 2000 both soften, trend being the only asset in the plan that won either one. Median wealth: −2%. The clause written into the plan fits in one sentence: "trend sleeve of 7.5%, judged **only** on regime years (the winter is normal operation), rebalanced at the bands like everything else." Without that clause, do not buy. The asset without the patience is a deferred donation to the market.
:::

## The essentials

- Trend follows the direction of 50 to 100 futures markets, long and short, at a targeted risk level, with collateral in Treasury bills that adds the cash yield on top. It ends up mechanically on the right side of anything that **lasts**, and big crises last: 2008, 1973, 2022 (its best year since 2000, the year everything else broke).
- You read the industry through its indexes: SG Trend for pure trend, SG CTA and BTOP50 for CTAs as a whole. More and more of it is bought through flat-fee **replication**, by positions (DBMF, a regression on the aggregate positioning of the big CTAs) or by rules (a simple model, openly so). The check never changes: track the benchmark, no performance fee.
- It is the only defensive asset with a positive expected return (2% to 4% real and gross, on current estimates), with a correlation close to zero, documented over a century (positive in every decade since 1880). It holds the long-regimes line in the table of defenses, and stays useless against short shocks (2020).
- For a retiree, 10% to 15% lifts the worst vintages by 0.2 to 0.4 of a point of safe withdrawal rate and shortens the drawdowns. The effect works exactly where plans die, provided the vehicle is clean.
- Buying it: a flat-fee trend fund (0.7% to 1.5%) or an index replicator, checked against the SG Trend Index. US-listed ETFs have carried this exposure since 2019. Skip the disguised multi-strategy funds, 2 and 20, and do-it-yourself replication.
- The lethal trap is behavioral: winters of five to nine years (2011-2019) make people capitulate right before the harvest. Written size, written thesis, judgment on regime years only. Without that discipline, this block is not for you.

---

## Going further

- Moskowitz, Ooi & Pedersen, "Time Series Momentum" (2012); Hurst, Ooi & Pedersen, "A Century of Evidence on Trend-Following Investing" (AQR): the evidence.
- Man Institute and AQR: the work on trend and crises (crisis alpha), and on blending trend into a classic portfolio.
- The SG Trend Index and the SG CTA Index (Société Générale), the BTOP50 (BarclayHedge): the industry's public benchmarks, for checking any vehicle.
- Andrew Beer (DBi) on replication and fee alpha: the low-cost camp's thesis, argued bluntly in his interviews and papers.
- Early Retirement Now, Part 63 (momentum and withdrawal) ([[the-ern-series]]).
- In this book: [[defensive-assets]] (the slot this fills), [[market-regimes]] (why crises are trends), [[all-weather-portfolios]] (the Dragon and where trend sits), [[gold-in-retirement]] (sharing the inflation slot).
