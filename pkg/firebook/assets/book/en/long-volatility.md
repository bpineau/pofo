# Long volatility and tail hedging: paying for crashes
<!-- source: long-volatility @ 11abc2b06627 -->

The table of defenses ([[defensive-assets]]) hands out roles by type of crisis. Long bonds defend against deflationary recessions, gold against monetary crises, trend ([[managed-futures]]) against bear regimes that drag on. One slot is still open, and it is the most spectacular of them: the **fast crash**, those two to eight weeks when everything falls at once (October 1987, the acute phase of autumn 2008, February and March 2020). That is the specialty of a family of strategies called **long volatility**, or **tail hedging**: holding, permanently, instruments that explode upward when markets collapse, mostly put options and their cousins.

Retirees warm to the idea right away, and for a good reason: sequence risk lives in the tails ([[sequence-of-returns]], [[fat-tails]]). It still deserves the most cautious article in this part. Permanent crash protection is **insurance**, and insurance costs a premium. What follows is how much that premium costs, why most of the vehicles you can actually buy are toxic, what the famous successes (Universa in 2020) owe to conditions nobody can reproduce, and the narrow cases where a small slice still makes sense. Above all, you will come out able to explain why your plan probably holds none of it, which is a skill in itself.

::: cle The idea in one sentence
Being long volatility means paying a premium every month for an asset that almost always loses a little and very rarely wins enormously, at exactly the worst moment for everything else. It is the mirror image of stocks, which almost always gain a little and rarely lose enormously. The concept is not the problem; the accounting is. Over a long stretch, the premiums paid usually exceed the claims collected, because you are buying the insurance everyone wants ([[risk-premia]]).
:::

## The mechanism: convexity and the variance risk premium

The basic instrument is the **put**: the right to sell the index at a set price. If the market rises or goes nowhere, the put expires worthless and you lose the premium. If the market collapses, the put is worth the fall, multiplied by a leverage that grows with the violence of the move. That asymmetry has a name, **convexity**: the position gains more in big moves than it loses in small ones. A sleeve of out-of-the-money puts can multiply five to forty times over in a −30% crash while costing only a few percent a year.

The machinery comes with a structural cost that research has been measuring for thirty years: the **variance risk premium**. Implied volatility (the price of insurance, quoted by the options market) runs 2 to 4 percentage points above realized volatility on average. Insurance is permanently sold too dear, precisely because demand for protection is huge and price-insensitive. Whoever buys puts systematically pays that markup year after year; whoever sells them collects it. This is one of the few empirical results finance agrees on, documented on every major exchange. A retiree thinking about long vol has to see the position clearly: you are standing on the wrong side of a risk premium, betting that the service rendered, convexity at the worst possible moment, is worth the markup.

**The cost, in numbers.** Public indexes settle this without argument. The CBOE PPUT (the S&P 500 permanently protected by a put 5% out of the money) returned 6.6% a year from June 1986 to December 2018, against 9.8% for the unprotected S&P 500: 3.2 points less every year, and the worst drawdown improved only from −51% to −38.9%. A pure put sleeve, without the equities, costs on the order of 2% to 5% of its notional a year in calm markets. The standard academic critique, "Pathetic Protection" (Roni Israelov, 2017), finds that systematic put protection is beaten over nearly the whole history by something trivial: hold fewer stocks and more cash or bonds ([[stock-bond-allocation]]). Its most telling number, measured from March 1996 to June 2014, fits in one sentence. Putting 36.5% in the S&P 500 and leaving the rest in cash returned exactly as much as PPUT.

::: figure longvol-profil
The typical profile of a systematic put sleeve (stylized): a premium paid out continuously, repaid by rare spikes of convexity. The spike only counts if it gets monetized.
:::

## The textbook case: Universa and March 2020

The other side of the argument has its own pedigree. Universa, the fund advised by Nassim Taleb ([[fat-tails]]), reported a return of roughly 42 times its money on its options sleeve for the first quarter of 2020, and 37 times in March alone. Bolted onto an equity portfolio, an allocation of a few percent was enough to erase the fall. Mark Spitznagel builds the thesis of his book Safe Haven on it. Insurance built well can cost so little and pay so much that it raises the compound growth rate of the whole portfolio, because it removes the volatility drag of crashes ([[arithmetic-vs-geometric-returns]]).

The thesis is intellectually serious, and three caveats bring it back to earth. First, **monetization**. The gain from a crash exists only if someone sells the puts at the peak of the panic and buys stocks back at the bottom. That is real-time execution discipline, which the fund exercises on your behalf and which an individual almost always misses. Second, **selection**. Universa is the famous survivor of a category, the tail risk funds. The category's aggregate index, the CBOE Eurekahedge Tail Risk Index, lost on the order of 4% to 6% a year over the 2011-2019 decade. The group average looks nothing like its showroom. Third, **access**. These funds are simply not open to individual investors, and no fund range you can buy from carries anything equivalent.

::: science Fast crash against slow bear market: the division of labor
What decides where long vol belongs is the **speed** of a crisis, not its depth. In a fast crash (1987, 2020), the puts explode while trend has no time to turn around (−5% to 0% in March 2020). In a slow bear market (2000-2002, 2022), it goes the other way. Implied volatility stays moderate, the puts expire worthless month after month while the portfolio erodes, and trend posts its historic year (2022). The blending studies (Man Institute, AQR, the work around the Dragon portfolio, [[all-weather-portfolios]]) agree: long vol and trend complement each other, each covering the crisis the other misses. But trend's long-run expectation is positive and long vol's is negative. On a limited defensive budget, research and common sense both put trend first and leave long vol as a top-up, for whoever can buy it cleanly.
:::

## What a retiree can actually buy (and what to run from)

This is where the article turns into a warning, because the gap between the concept and the vehicles within reach is the widest on the whole defensive menu.

**Run from these: VIX futures ETPs.** The "long VIX" trackers (and their leveraged versions) hold futures on the volatility index, which they have to roll every month by buying dearer than they sell, since contango is the normal shape of that curve. The structural bleed reaches 40% to 70% a year: these products lose more than 99% over any long stretch, whatever crashes happen along the way. They are built for bets lasting a few days, not for holding; inside a retirement plan they are a donation to the market maker. The inverse product (short VIX) has a rap sheet of its own: the XIV note lost 96% in a single evening in February 2018.

**Disappointing: retail "volatility" funds and structured products.** Most funds labeled volatility run **relative** strategies (volatility arbitrage, covered call writing) with none of the convexity you came for, and the category's history is littered with closures. "Protected" structured products sold through insurance and bank channels charge for the protection with opaque margins that destroy the point of it ([[building-it-with-us-etfs]] for what a clean vehicle looks like).

**Workable but demanding: buying puts directly.** A brokerage account with options approval gives you the do-it-yourself version. You buy puts on an ETF or an index, 10% to 20% out of the money, three to twelve months to expiry, rolled on a schedule, with a premium budget fixed in advance (0.5% to 1.5% of the portfolio a year). Taxes add their friction, and the operational burden is real, because someone has to keep rolling options past age 70, for ten or thirty years. None of it escapes the arithmetic of the variance risk premium either. The one variant that stands up economically is the **collar**: fund the put by selling a call, giving up the gains above a ceiling to be protected below a floor, which pens the portfolio into a tunnel at close to zero cost. It is a disguised way of cutting equity exposure, and that is exactly why it works.

| Vehicle | Real convexity | Carry cost | Verdict |
|---|---|---|---|
| VIX futures ETPs | yes, for a few days | 40% to 70% a year (contango) | banned from a retirement plan |
| Retail "volatility" funds | rarely | 1% to 2% a year | disappointing, check the strategy |
| "Protected" structured products | capped | opaque margins | beaten by bonds plus stocks |
| Direct puts (10% to 20% OTM) | yes | 0.5% to 1.5% a year in premium | workable, demands discipline |
| Collar (put funded by a call) | bounded on both sides | ~0 | the only economically clean one |
| Institutional tail risk funds | yes, managed | varies | closed to individual investors |

**The honest alternative, for almost everyone.** Israelov's conclusion is still the best default answer: if you want to suffer less in crashes, hold fewer stocks. Moving from 70/30 to 60/40 protects you with certainty, with no variance risk premium, no options to roll and no vehicle risk. The bond tent of the fragile years ([[glidepaths]]) and the cash buffer ([[cash-buffer]]) give the same defense against the same sequence risk, using instruments you already own.

::: figure puts-domines
The whole range of stock and bond mixes, recomputed here (S&P 500 and intermediate Treasuries, month-end values, annual rebalancing), against permanent put protection, over the published window of the PPUT index, June 1986 to December 2018. At equal worst drawdown, a plain 80/20 returns 2.9 points more a year. The PPUT point sits at its published coordinates; it is not recomputed here.
:::

::: exemple The tail hedge against three crises
Portfolio of EUR 1,000,000, withdrawal of EUR 38,000 a year. A: 65% stocks / 35% bonds. B: 64% stocks / 35% bonds / 1% a year of put budget (rolled, 15% out of the money). March 2020: B wins. The puts return about ten times their premium, the drawdown goes from −22% to −12%, and monetizing at the bottom buys stocks on sale, if and only if the selling rule was written in advance. 2022: B loses twice over. The market falls 20% over twelve months with no volatility panic, the puts expire worthless quarter after quarter, and B gives up 1.2 points on top of the fall while the neighbor's trend fund makes +25%. 2013-2019: B pays seven years of premium for nothing, roughly 6 points of cumulative lag. Over the full cycle, B is worth having only if the fast crash arrives early in retirement and the monetization is executed coldly. It is a bet on one precise scenario, not an all-risks policy.
:::

## So: a small slice, or none at all?

For the vast majority of readers of this book the answer is none, and that is not a defeat. The anti-crash role is already filled, better and cheaper, by duration plus cash plus trend plus flexible withdrawal rules ([[choosing-your-strategy]]). Legitimate cases exist, but they are rare. A large portfolio, with options access and a taste for execution, can spend 0.5% to 1% a year of premium budget during the **fragile window**, the first five years of withdrawals, where sequence risk concentrates. The collar is the better form, with a monetization rule written before the purchase. Anyone drawn to the Dragon portfolio will take on the recipe's long vol sleeve (around 21%) knowing it is the hardest part of it to replicate honestly, and that the version within reach often betrays the spirit of the original.

If you keep one discipline from this article, keep this one: never hold an insurance asset without having written down when you will sell it. A put that went up tenfold in a crash and then gets kept "just in case" turns back into evaporated premium on the rebound. Insurance that never gets monetized never existed.

## The essentials

- Long volatility holds convexity, puts for the most part, that explodes in fast crashes. It is the exact complement of trend, which covers the slow crises, and it replaces neither trend nor anything else.
- The buyer of protection pays the variance risk premium (implied volatility above realized, 2 to 4 points on average): the long-run expectation is negative, on the order of −2% to −5% a year for a systematic put sleeve.
- The famous successes (Universa, up more than 40-fold in the first quarter of 2020) rest on professional monetization and on survivor selection. The tail risk category index lost 4% to 6% a year over 2011-2019, and none of this is available in an ordinary fund range.
- VIX ETPs bleed 40% to 70% a year in contango and have no place in a retirement plan; retail "volatility" funds usually lack the convexity they promise.
- The defense that does the same job with certainty is simple: hold fewer stocks (Israelov). The legitimate uses that survive are rare: a collar on a written budget during the fragile window, or a Dragon sleeve owned deliberately, with a monetization rule.

---

## Going further

- Roni Israelov, "Pathetic Protection: The Elusive Benefits of Protective Puts" (2017): the standard critique, with the numbers behind it.
- Mark Spitznagel, Safe Haven: Investing for Financial Storms (2021): the opposite thesis, from its best advocate; read it with this article next to you.
- CBOE: the PPUT (put protection) indexes and the variance premium data; Eurekahedge for the tail risk and long volatility fund indexes.
- Artemis Capital, "The Allegory of the Hawk and Serpent": where the Dragon portfolio and the long vol case come from ([[all-weather-portfolios]]).
- In this book: [[defensive-assets]] (the table of roles), [[managed-futures]] (the complement with a positive expectation), [[fat-tails]] (why crashes come round more often than the bell curve says), [[glidepaths]] and [[cash-buffer]] (the default defenses of the fragile window).
