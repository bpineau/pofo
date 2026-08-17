# Market regimes (growth × inflation, sticky bears) and why they matter
<!-- source: regimes-de-marche @ a963d2f484bb -->

Market returns do not come out of one uniform urn. They are produced by an economy that moves through **seasons** lasting years, and inside each one almost everything behaves differently: the growth of the 1990s, the stagflation of the 1970s, the deflation of the 1930s, Japan's lost decade. Those seasons, the market regimes, are the deep reason behind almost everything the earlier articles saw at the surface: the clusters of bad years ([[sequence-of-returns]]), the fat tails ([[fat-tails]]), and why independent draws are not enough ([[simulator-traps]]).

They are also the key to the question that runs through this book's portfolio chapters: why does a classic 60/40, so sturdy in some decades, get taken apart in others, and what do you have to own to survive every season? This page lays out the frame. The evidence that regimes exist and persist. The growth × inflation grid that sorts them, the common language of Browne, Dalio and macro research. What each regime does to each asset class. And the two practical consequences: how you model regimes (sticky bears), and how you prepare for them without pretending to forecast them.

::: cle The idea in one sentence
A portfolio is never neutral. It is an implicit bet on a regime, and the 60/40 bets on disinflationary growth. A withdrawal plan is a 40-year promise that will statistically pass through two or three hostile regimes. The job is not to dodge the bad seasons, since nobody forecasts them reliably, but to own something that survives each one, and to have run them through the simulator before life runs them at you.
:::

::: figure regime-grid
The growth × inflation grid: every season has its winners. The 60/40 covers the left column only; stagflation (bottom right) is the hole in the classic defense and the retiree's nightmare.
:::

## Regimes exist: the evidence in three facts

**Fact 1: bear markets are episodes, not accidents.** Over a century of US data, the big equity bear markets (real declines of 20% or worse) run one to two years at the median. But the tail is long: 1929-1932 (34 months, −80% real), 1973-1974 (21 months, −55% real once inflation is counted), 2000-2002 (30 months), 2007-2009 (17 months). Getting back to the previous real peak takes far longer still: 7 years after 1974, 13 years after 2000, more than 30 years in Japan after 1990. An i.i.d. world would produce falls that deep, but never stretches that long. A persistent run of mediocre years is the signature of a regime.

**Fact 2: volatility and correlations switch states.** Volatility comes in clusters, whole years at 25% following whole years at 10%. The correlation between stocks and bonds even changes sign with the regime. It was negative from 2000 to 2021, when bonds cushioned equity crashes: the golden age of the 60/40. It turns positive again in inflationary decades, as it did through the 1970s and abruptly in 2022: stocks −20%, long bonds −30%, the worst year for the 60/40 in a century. That correlation flip is the single most important market fact for a modern retiree, because every bit of the "classic" protection in the portfolio hangs on it ([[bonds-in-retirement]]).

**Fact 3: the econometrics back it up.** Since Hamilton (1989), regime-switching models (hidden Markov chains) beat homogeneous models at describing returns. The data prefer a description in persistent states (expansion or recession, calm or crisis) with low transition probabilities. You do not leave a regime easily, you stick in it. That is exactly the structure the sequence-stress models borrow ([[making-monte-carlo-relevant]]).

Where do regimes come from? From the fact that the economy itself has persistent states. Credit and debt cycles build over years. Monetary policy corrects late, then overcorrects. Recessions feed themselves: layoffs cut demand, which causes more layoffs. And collective psychology amplifies all of it. Euphoria manufactures expensive valuations ([[valuations-and-cape]]), fear manufactures the floors. Markets inherit the persistence of the real world they discount.

## The growth × inflation grid: the four seasons

The most productive frame for sorting regimes crosses two macroeconomic surprises: growth (above or below expectations) and inflation (same). Four quadrants, four seasons, each with its winners and its victims. This is the common language of Harry Browne (the Permanent Portfolio, 1981), Ray Dalio (All Weather, in the 1990s) and modern macro research ([[all-weather-portfolios]]).

| Regime | Growth | Inflation | What wins | What suffers | Typical episodes |
|---|---|---|---|---|---|
| **Prosperity** (disinflationary boom) | + | − | Stocks, bonds, anything long | Gold, cash (opportunity cost) | 1982-1999, 2009-2021 |
| **Overheating / inflation** | + | + | Commodities, gold, real assets, linkers | Nominal bonds, expensive stocks | 1965-1969, 2021-2022 |
| **Stagflation** | − | + | Gold, linkers, interest-bearing cash, TSMOM | **Everything** else: stocks and bonds | 1973-1981, the retiree's nightmare |
| **Deflation / bust** | − | − | Long government bonds, cash | Stocks, credit, real estate, commodities | 1929-1938, Japan in the 1990s, 2008 |

Three readings of this grid are worth spelling out.

**The 60/40 is a bet on the left column.** Stocks and nominal bonds win together in prosperity, and they hedge each other in deflation (bonds rise when stocks crash, as in 2008). But the inflation row hits them together: inflation destroys the real value of a bond's coupons and compresses equity multiples. A classic portfolio is not "diversified" in the sense regimes care about. It is diversified inside two quadrants out of four. As long as inflation sleeps, from 1982 to 2021, nobody notices. 2022 was the generational reminder.

**The retiree's worst quadrant is stagflation, and it is no accident that the worst vintage is 1966.** Reread [[the-trinity-study]] with the grid in hand. The 1966 retiree did not live through a crash. He lived through fifteen years in a hostile quadrant: zero real returns on stocks, bonds ground down by inflation, and indexed withdrawals swelling while everything else shrank ([[inflation-and-withdrawal-rates]]). Stagflation stacks the retiree's three plagues: negative real returns on both classic assets, withdrawals that keep climbing, and duration. This is the regime a withdrawal portfolio has to be armed against on purpose, with gold, inflation-linked bonds, real assets and trend following ([[gold-in-retirement]], [[inflation-linked-bonds]], [[managed-futures]]).

**No asset wins everywhere, but every regime has its winners.** That is the logic behind the all-weather portfolios. Instead of maximizing the return of the likely quadrant, which is the accumulator's reflex, you hold at least one winner per quadrant at all times and let rebalancing harvest the rotations ([[all-weather-portfolios]], [[defensive-assets]]). The cost is a lower expected return in the middle. The gain is a tighter distribution, exactly the trade the math of withdrawal rewards ([[arithmetic-vs-geometric-returns]], [[sequence-of-returns]]).

::: attention Can you forecast the regime? Honesty first
The natural temptation is to identify the current regime and overweight its winners. The research is mixed. Detecting the regime you are in is partly doable: macro indicators like industrial production, inflation and the slope of the yield curve classify the present reasonably well. Forecasting the switches has stayed largely out of reach, and retail tactical strategies destroy value on average, through their lags and their false signals. Disciplined quantitative approaches do exist. Tactical versions of the Permanent Portfolio shift their four sleeves with the measured macro regime, and managed futures amount to regime detection by prices ([[managed-futures]]). But they demand mechanical execution, with no second-guessing. For the vast majority of plans, structural preparation (all-weather) beats prediction: the grid is there to audit the portfolio, not to trade it.
:::

## Modeling regimes: how it is done, and why

What regimes mean for simulation has already been laid out ([[making-monte-carlo-relevant]]). Here it is again from the seasons' point of view, which is where it earns its keep.

**The sequence stress is a minimal regime machine.** Two states, normal and bear. Sticky transitions, calibrated so the bear state takes about one year in five, in episodes of roughly three years. Amplified volatility inside the bear state, and a long-run mean left untouched. It is Hamilton stripped to the bone: just enough structure to reproduce the property of regimes that kills retirees, the persistence of bad stretches, with no pretense of modeling the macroeconomy. The gap between the central case and the stress measures your exposure to that persistence.

**The broad sample carries real regimes.** The whole-country block bootstrap of the broad-sample model ([[historical-vs-parametric]]) runs your paths through real stagflations (the United Kingdom in the 1970s), real deflations (the 1930s) and real national collapses (Japan), each with the stock-bond correlation of its own era. It is the only common model where the grid's inflation row exists natively. Hence its role as the tiebreaker for portfolios that lean too far into the left column ([[anarkulova-cederburg]]).

**The lost decade is a quadrant turned into a scenario.** The long bust regime, isolated and stretched to its Japanese length: the crash test for the bust quadrant, the one that has to be made survivable ([[bear-markets-in-retirement]]).

None of these models tries to capture the fine rotation of asset classes from one quadrant to the next, since the portfolio is aggregated before any draw is taken. The regime grid does its work upstream, when the portfolio is put together. The simulation then tests how robust that composition is. The two stages complete each other: build with the grid, test with the models.

::: exemple Auditing a portfolio with the grid
The portfolio: 70% global equities, 30% nominal government bonds of 7 to 10 years. Quadrant by quadrant: prosperity, excellent (both legs win); deflation, good (duration cushions); overheating, poor (both suffer, moderately); stagflation, catastrophic (both lose in real terms, for years, with no winning sleeve anywhere). Verdict: three quadrants out of four look covered, but the retiree's worst quadrant is wide open. A minimal fix: 10% gold and 10% inflation-linked bonds, taken out of both sleeves ([[gold-in-retirement]], [[inflation-linked-bonds]]). The central case's expected return then drops by about 0.2 points, and the broad-sample model, the one that holds the stagflations, typically hands back several points of failure probability. That is the regime-for-expectation trade, made visible and countable: the grid ran the audit, the simulator settled it.
:::

## What regimes change in running the plan

Beyond the portfolio and the model, three consequences for steering.

**The crossings are counted in years, so budget them that way.** A cash buffer sized for "a crash" (18 months) is sized for the wrong thing. On the real US 60/40, the median crossing, the years spent inside a bear market, lasts 32 months, and the four worst since 1953 all ran past three years, out to more than ten for the stagflation of the 1970s. That is the order of magnitude that should set the buffer and the flexibility rules ([[cash-buffer]], [[refilling-the-buffer]], [[flexibility-in-practice]]). You have to hold out through a regime, not absorb a jolt.

**The regime you retire into deserves a look, not an obsession.** Leaving at the euphoric end of a boom (expensive valuations, [[valuations-and-cape]]) and leaving at the bottom of a bust that has already cleaned itself out do not carry the same sequence risk. Which is why you locate the starting point first, today's valuations set back against their own century, before running any projection. But you do not choose the regime you start in, you choose your margins ([[the-three-phases]]).

**Do not mistake a regime for noise.** A year at −15% is not a regime change, and a written operating plan ([[when-to-worry]]) has to resist the urge to "see" a 1973 in every correction. Regimes are recognized by their macro persistence, inflation that has settled in, a recession that has been declared, not by one quarter's headlines. A plan that reacts to every pseudo-regime does worse than a rigid one. That is the whole point of quantitative thresholds decided in advance, while calm.

## The essentials

- Markets have persistent seasons that last for years. Their statistical signatures (long bear markets, volatility clustering, correlations that flip) are exactly what i.i.d. models miss.
- The growth × inflation grid sorts the seasons into four quadrants, and the 60/40 covers two of them. The whole inflation row (overheating, stagflation) hits stocks and bonds together: 1966-1981 back then, 2022 as the reminder.
- The retiree's worst quadrant is stagflation (negative real returns on both classic assets, plus swelling withdrawals, plus duration). A withdrawal portfolio arms itself against it on purpose, with gold, linkers, real assets and trend.
- Structural preparation beats prediction: the grid audits the composition, one winner per quadrant, and the simulator tests how robust it is (stress = persistence, broad sample = real regimes, lost decade = crash test).
- Steering: budget the crossings in years (32 months at the median, more than 10 in the worst case), look at the regime you retire into without obsessing over it, and do not take a correction for the end of an era. Macro persistence decides, not the headlines.

---

## Going further

- Hamilton, "A New Approach to the Economic Analysis of Nonstationary Time Series and the Business Cycle" (1989): the founding paper on Markov regimes.
- Harry Browne, *Fail-Safe Investing* (1999) and Ray Dalio (Bridgewater), "The All Weather Story": the two classic ways the grid became a portfolio ([[all-weather-portfolios]]).
- Ilmanen, *Investing Amid Low Expected Returns* (2022), the chapters on regimes and inflation: the practitioner state of the art.
- In this book: [[under-the-hood]], for how the sequence stress, the broad sample and the lost decade are implemented.
- Next in this book: [[all-weather-portfolios]] and [[defensive-assets]], where the grid becomes an allocation.
