# Beyond the United States: Anarkulova, Cederburg and the world sample
<!-- source: anarkulova-cederburg @ 15d151dec43a -->

The whole safe withdrawal rate tradition, from Bengen to Trinity and on to nearly every online simulator, rests on one data set: US markets since 1926 ([[the-trinity-study]]). The twentieth-century United States is not a neutral sample. It is the country that won the last century.

Since the 2010s, a line of research has redone those calculations on the **full** experience of the developed world: Dimson-Marsh-Staunton, Wade Pfau, and above all Aizhan Anarkulova, Scott Cederburg and Michael O'Doherty. The results are uncomfortable. The world safe withdrawal rate sits well below 4%, and bond-heavy portfolios protect less than we thought. This page walks through that work, its numbers, and the fair criticisms of it. It also shows how a simulator turns it into something usable, the model known as the broad sample, and what you need to know to read it without overreading it.

::: cle The idea in one sentence
Grant that France, Japan, Germany or Italy between 1900 and 2020 are **possible** futures for a developed-market investor today. Then long-horizon risk is higher than US history alone suggests. On the world sample, a rigid 4% rule fails far more often, and the rate carrying a 5% failure probability for a 65-year-old couple holding a 60/40 in its own home market lands nearer 2.3 to 2.7%, depending on the assumptions. That is not the only way to read the world, but it is the best documented prudent bound we have.
:::

## Geographic survivorship bias

Start with the problem. The Ibbotson data begins in 1926 in New York. It holds the Great Depression and the stagflation of the 1970s, but it also describes a country never invaded, never in default on its domestic debt, whose currency became the world's reserve and whose stock market was the great winner of the century. Pick that country as your only sample and you calibrate your plan on the winning lottery ticket.

The developed twentieth century produced plenty of what the US sample does not contain: stock markets shut down or expropriated (Germany 1948, Japan 1946), hyperinflations that pulverize bonds ([[hyperinflation-and-extremes]]), lost decades (Japan after 1990, stocks below their real peak for more than thirty years), defaults and financial repression. A German, Japanese, French or Italian retiree who left between 1900 and 1960 with an American-style plan was ruined in a fair share of cases. Not through bad luck in the sequence, but because their **country** lived through history.

Dimson, Marsh and Staunton, the London Business School trio behind the *Global Investment Returns Yearbook*, put a number on the gap back in 2002. Real stock returns from 1900 to 2020 ran about 6.5% a year in the United States against about 4.5% outside it. And in half the countries, government bonds did **worse** than 0% real over long stretches. Wade Pfau applied Bengen's method to 17 countries as early as 2010. The 30-year SAFEMAX, about 4% in the United States, falls below 3% in a third of them and below 1.5% in the worst, Japan, Italy and France among others, sunk by their postwar inflations.

::: figure safemax-pays
Jorda-Schularick-Taylor panel, sixteen developed countries, 1870 to 2020: the initial withdrawal rate a domestic 60/40 would have carried for 30 years in each country's worst vintage, computed here over 59 to 121 windows depending on how far back the country's data goes. The Pfau and Anarkulova-Cederburg SAFEMAX numbers quoted above come from other databases: the levels do not match to the tenth of a point, only the ranking is comparable.
:::

## Anarkulova, Cederburg and O'Doherty: the modern method

The work of Anarkulova, Cederburg and O'Doherty ("The Safe Withdrawal Rate: Evidence from a Broad Sample of Developed Markets", 2023, cosigned with Sias, plus the sister papers, among them "Beyond the Status Quo", 2023) modernizes the question on three fronts.

**The data.** The cleanest developed-market database available, built on the GFD reference set and on long-run academic work. It covers 38 developed countries and roughly 2,500 country-years of real returns in stocks, bonds and cash. The care taken over survivorship and look-ahead bias is unusual: a country enters the sample when it was developed **at the time**, not in hindsight. Argentina, rich in 1900, is therefore in. That is the sore point, and we come back to it.

**The method.** Instead of replaying windows from a single country, a **block bootstrap**. Ten-year blocks are drawn from the whole grid of countries and eras, which preserves clusters, trends and regimes ([[sequence-of-returns]]). Those blocks are then assembled into synthetic retirements of whatever length you want. Each simulated retirement lives the history of one coherent developed country, piece by piece, disasters included. That machinery is what the broad-sample models copy, on a more modest database, since the authors' own is not open ([[historical-vs-parametric]]).

**Mortality.** Instead of a fixed 30-year horizon, a 65-year-old couple tracked through real mortality tables. Retirements therefore run for a random length, sometimes 35 years and more.

Here are the headline results for a 65-year-old couple holding a domestic 60/40. The rigid 4% rule fails in about 17% of cases, against roughly 2% on US data alone. The rate carrying a 5% failure probability comes out near **2.26%**. For an early retiree with a longer horizon it is worse still. The same team draws an even more uncomfortable conclusion in "Beyond the Status Quo": in their sample, all-equity portfolios diversified internationally beat stock-bond mixes over long horizons. The reason is that bonds get destroyed by inflation every so often, in exactly the eras that flatten local stocks. International diversification in equities protects better than domestic bonds do ([[international-diversification]], [[bonds-in-retirement]]).

Three samples travel under the same flag, and mixing them up wrecks any reading. Here is what separates them.

| Sample | Data and draw | Portfolio | Horizon | Published figure |
|---|---|---|---|---|
| The paper (2023) | 38 developed countries (GFD), blocks from every country | domestic 60/40 | real mortality, 65-year-old couple | ~17% failure at 4%; 2.26% at a 5% failure probability |
| The broad-sample model | 16 countries (JST panel, 1870 to 2020), blocks from one country | domestic 60/40 | your plan's | world basket at 3.3% over 30 years |
| US history | United States since 1926, rolling windows | 50/50 to 75/25 | fixed 30 years | SAFEMAX ~4.15%; ~2% failure at 4% |

Read that last column carefully, because the three numbers do not measure the same thing. The first is a probability over retirements of random length; the other two are worst-vintage rates, one across a basket of countries, the other for a single one. The databases differ too. On the panel behind the figure above, the United States carries only 3.75%, where the Bengen tradition prints 4.15%.

::: figure echantillon-croise
The same plan judged twice: a domestic 60/40 and a rigid indexed withdrawal over thirty years, measured first on the panel's American record alone, then on the broad-sample model that draws its blocks inside all sixteen countries. The two verdicts start almost together and separate exactly in the band where everyone sizes a plan. The numbers the 2023 paper publishes (17% and 2.26%) belong to a third setup, thirty-eight countries and real mortality, and the legend puts them back in their place.
:::


::: science Where these numbers sit in the literature
Keep the spread of bounds for a long horizon in mind. US history alone gives about 3.25 to 3.5% rigid ([[the-ern-series]]). Morningstar's forward-looking returns over 30 years give about 3.9% ([[morningstar-guardrails]], [[expected-returns]]). The Anarkulova-Cederburg world sample gives about 2.3 to 2.7%. The gap between those bounds is **not** technical disagreement: they answer different questions ("what if the future looks like America, like today's expectations, or like the whole developed century?"). A serious plan knows all three numbers and picks its spot deliberately, rather than ignoring the two that hurt.
:::

## The honest criticisms

This work has serious opponents, ERN first among them, who has devoted several analyses to it, and the objections are worth knowing, because they bound the pessimistic reading the way the American bias bounds the optimistic one.

**The weight of military catastrophe.** A large share of the sample's failures comes from world wars **fought on home soil** (Germany, Japan, Austria, France) and the hyperinflations that followed. If your capital-destruction scenario is a military occupation, an ETF portfolio is not your real problem anyway. The counter: strip out the wars and you also strip out the only observations of political tail risk anyone has, and events of that class, confiscation, financial repression, a currency destroyed, do not require a world war.

**Argentina, and where the sample stops.** Including countries that were developed **at the time** and later fell behind, Argentina among them, is defensible. It is exactly the survivorship bias you want to avoid: in 1900, nobody knew who would fall behind. But it does drag the numbers down for an investor in today's core countries.

**The simulated investor is domestic.** The simulated retirements live one country's history, in local stocks and local bonds. A globally diversified investor today, holding an unhedged world equity fund, would not have taken Japan 1990 or Italy 1970 full in the face: international diversification cushions precisely the worst blocks of the sample. This is probably the most important criticism in practice, and it argues for diversification ([[international-diversification]]) more than it argues **against** the study.

**Overlapping blocks and the real sample size.** Two thousand five hundred country-years sounds like a lot. But the draw pulls ten-year blocks out of overlapping windows, so the same German or Japanese decade shows up in a great many simulated retirements, and a single disaster weighs on thousands of paths. Crises are global and correlated on top of that: 1929, 1973 and 2008 hit everyone. The sample of **independent** disasters stays small, and the uncertainty around these numbers is itself wide.

The reasonable synthesis: the true risk facing a globally diversified investor today sits somewhere between US history and the domestic world sample, and nobody knows where. So here is the rule to work by, whatever tool you use. Keep both bounds on screen, side by side, all the time, and distrust any verdict that shows only one.

## What it changes for your plan

**The rigid rate you size with.** If you size with no margins ([[how-much-you-need]]), the world bound argues for 3 to 3.5% rather than 4%. What it really rules out is calling 4% "scientifically safe" over 45 years. But aiming at 2.3% (43 times your spending!) out of pure literalism would be an overreaction. That is the domestic worst-case bound, and your real margins, a pension, flexibility, diversification, dominate it comfortably.

**How to read a broad-sample model.** That model is **not** your portfolio, and your return assumptions change nothing in it, since it draws only from real data. If your plan holds there, it holds in the harshest developed world on record. That is the best robustness label available. If it does not hold, look at **where** the scenarios fail, usually in inflationary blocks ([[inflation-and-withdrawal-rates]]), and at what your portfolio lacks for those regimes ([[all-weather-portfolios]], [[defensive-assets]]).

**The portfolio.** Two direct lessons. International equity diversification is not a refinement: it is the defense against the dominant risk of the sample, one country falling behind, yours included. And domestic nominal bonds are not the safe asset of a long horizon, because their worst enemy, sustained inflation, is the retiree's worst enemy too ([[inflation-linked-bonds]], [[gold-in-retirement]]).

::: exemple The same plan under both bounds
The plan: $1.4M, $45,000 a year rigid (3.2%), 45 years, a pension of $15,000 a year from age 67, 70% world equities and 30% bonds. Historical windows of the portfolio give about 1% ruin. The calibrated central case gives about 4%. The broad sample gives about 9%, with the failures concentrated in blocks of persistent inflation, three quarters of them past age 80, pension already in hand. Reading: the plan is solid, because 3.2% with a pension is cautious to begin with. The broad-sample tail is handled with 10 to 15% of written flexibility ([[flexibility-in-practice]]) rather than with more capital. Without the broad sample, nobody would have known the residual failure mode was inflation and not a crash.
:::

## The essentials

- US history is the century's winning lottery ticket: calibrate a plan on it and you inherit its optimism; the full developed world tells a harder story.
- Anarkulova-Cederburg-O'Doherty (38 countries, block bootstrap, real mortality): the rigid 4% rule fails about 17% of the time, and the rate carrying a 5% failure probability comes out near 2.3% for a domestic 60/40 couple; domestic bonds protect less than international equity diversification does.
- The criticisms (wars, Argentina, the domestic investor, correlated crises) are serious: the truth for a globally diversified investor lies between the two bounds, and nobody knows where.
- In practice: 3 to 3.5% rigid for sizing, the broad sample as a robustness label, international diversification and inflation-resistant assets as the answers to the failure modes it reveals.
- You can run this evidence on your own plan: a broad-sample model replays the JST panel (16 countries, 1870 to 2020). How it is built is spelled out in [[under-the-hood]].

---

## Going further

- Anarkulova, Cederburg, O'Doherty & Sias, "The Safe Withdrawal Rate: Evidence from a Broad Sample of Developed Markets" (2023) and Cederburg et al., "Beyond the Status Quo: A Critical Assessment of Lifecycle Investment Advice" (2023): the source papers (SSRN, free access).
- Dimson, Marsh & Staunton, *Triumph of the Optimists* (2002) and the *Global Investment Returns Yearbook* (annual, UBS): the world century in numbers.
- Wade Pfau, "An International Perspective on Safe Withdrawal Rates" (2010): the forerunner.
- Early Retirement Now on these studies, in the parts he devotes to them: the counter-case, argued ([[the-ern-series]]).
- Jorda, Knoll, Kuvshinov, Schularick & Taylor, "The Rate of Return on Everything, 1870-2015": the academic panel behind the broad-sample model.
