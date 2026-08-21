# Fat tails, crises, and the Student-t
<!-- source: queues-epaisses @ 2b7e9a46365d -->

On October 19, 1987, the US market lost 20% in a single session. Under a normal law calibrated on the volatility of the day, that move sat more than twenty standard deviations out: a probability so small that the entire history of the universe should not have been enough to produce it once.

It happened anyway. And 2008, 2020, 1929, 1973 tell the same story at the annual scale: **markets throw extremes far more often than the bell curve allows**. These are fat tails. For a retiree, that is no statistical curiosity. A retirement plan fails in the tails, exactly where the normal law is not looking.

This page takes the subject in order. The phenomenon first: where the tails come from, and how they are measured. Then the standard tool for modeling them, the Student-t distribution and its df parameter. Then what tails really change in a withdrawal plan. And last, the honest limits of that modeling. By the end, df will have stopped being an esoteric knob. It is the number that decides how often your simulator is allowed to produce 2008.

::: cle The idea in one sentence
At **identical** volatility, two worlds can differ completely in how often they throw an extreme year. In the Gaussian world, a −30% real year is the stuff of legend. In the real world (df around 4 to 6), it shows up once or twice in a retirement. Volatility σ measures the **ordinary** agitation; the tail parameter df measures the appetite for **catastrophe**. A simulator that knows only σ is blind to half the risk that matters.
:::

::: figure fat-tails
Two laws with the **same mean and the same volatility**, but opposite tails. In the middle, ordinary years look alike. In the tail, everything changes: under the normal law, a −30% real year is the stuff of legend; under a Student-t with df 5 (fitted on real data), it is about **ten times** more frequent than under that normal, which amounts to an infinite df. The df parameter does not move the center. It sets how often catastrophes happen.
:::

## Why the normal law is tempting, and where it breaks

The normal law is not naive. It follows from the central limit theorem: a sum of many small independent effects tends toward a Gaussian. If returns were the sum of thousands of small independent pieces of news, they would be normal. And in the middle of the distribution they nearly are: ordinary years (−10% to +20%) roughly follow the bell.

The trouble is at the extremes, and it is huge. The standard measure is **kurtosis**. The normal law has a kurtosis of 3. **Monthly** stock returns typically come in at 6 to 12, daily returns far higher. Extreme months and extreme years are several times too frequent for the bell.

The deep reason comes down to two mechanisms that break the theorem's assumptions. The first: the effects are not independent. Volatility comes in **clusters**, and one big move announces others. This is the volatility clustering that GARCH models capture. Extremes then feed on themselves, through panic, margin calls and forced selling, everyone selling because everyone else is selling. The second: the market changes **regime**. The return distribution of a calm market and the one of a crisis are not the same law. And mixing two Gaussians with different volatilities mechanically produces a fat-tailed distribution.

So the tails are not a quirk. They are the statistical signature of a fact: markets are a social system with feedback loops, not an urn ([[market-regimes]]).

One reassuring fact: tails get **thinner** as you aggregate over longer horizons. Daily returns are wildly non-Gaussian. Monthly returns much less so, annual returns less again, because stacking twelve months smooths part of the extremes. Ten-year returns are almost respectable. A retirement plan lives at the monthly and annual scale, the rhythm of withdrawals. At those scales the tails are very real. That is where the model does its work.

## The Student-t and the df parameter, in plain terms

One family of distributions was built for exactly this problem: the **Student-t**. It is a bell with three parameters, the center μ, the scale (tied to σ), and the **degrees of freedom df**, which set how fat the tails are. At infinite df, the Student-t is the normal law. As df comes down, the center barely moves but the tails get heavier. Below df 4 they turn genuinely wild. This is the foundation of the central model in serious simulators: monthly Student-t draws, compounded into years ([[historical-vs-parametric]]).

What df changes **in practice**, at identical annual volatility (σ = 12%, μ = 4% real; the orders of magnitude are those of a "catastrophic" year at −30% real, close to a 3 standard deviation event):

| df | The world it describes | How often a −30% real year | Over a 45-year retirement |
|---|---|---|---|
| 30+ (about normal) | The world of math textbooks | ~1 year in 400 | Probably never |
| 10 | Moderate tails | ~1 in 150 | Maybe once |
| 5 (the sensible default) | The world of real monthly data | ~1 in 40 | Once or twice |
| 3 | A catastrophe-prone world | ~1 in 20 | Two or three times |

Compare the df 5 row with the df 30 row: **the same σ, and a factor of ten on how often catastrophes happen**. That is why two simulators both showing "12% volatility" can hand you unrelated verdicts. It all sits in the law of the tails, which most commercial tools do not even document; they are Gaussian without saying so ([[simulator-traps]]).

**Where does the df value come from?** You do not guess it, you **estimate** it from the data. The method is two steps. Measure the kurtosis of the portfolio's monthly returns. Then invert the Student-t's theoretical relation, kurtosis ≈ 3 + 6/(df − 4), to get the matching df. Months with a kurtosis of 7 to 9 give df around 5 to 6. Plain equity funds land near df 4 to 6, while a heavily diversified portfolio with defensive sleeves can climb to 8 to 12. A tool that can read the history of your holdings runs that computation for you and leaves the value editable for sensitivity tests. So a df of 5 is not a cautious opinion. It is the number that comes out of the real monthly data of most global equity portfolios.

::: attention What df does not measure
Student's df is **symmetric**: it fattens the tail of miraculous years exactly as much as the tail of disasters. Real return distributions are also **skewed** (negative skew: downside extremes are more frequent and more violent than upside ones, the market takes the stairs up and the elevator down). A symmetric Student-t therefore understates, a little, how nasty the left tail really is. The fix for that asymmetry does not come from the distribution but from the **sequence**: stress models (sticky bear markets, volatility amplified on the way down) and the lost decade put the violence where it actually lives, in the chains of events ([[making-monte-carlo-relevant]]).
:::

## What tails change in a withdrawal plan

From the model to the plan. Fat tails hit a withdrawal plan in a precise pattern, and understanding it heads off two opposite mistakes.

**First effect, direct: failure probability rises, the median barely moves.** Going from df 30 to df 5 at constant σ leaves the median scenario almost untouched, because ordinary years have not changed. But failure probability swells, typically by 30 to 80% in relative terms depending on the plan. A Gaussian 4% failure becomes 6 to 7% at df 5. Tails do not change your **likely** life. They change how often the unlikely lives happen, and those are exactly the ones you plan against ([[failure-probability]]).

**Second effect, subtler: the interaction with sequence.** A −30% year does not cost the same depending on when it lands. In year 2 it cuts the plan down for good ([[sequence-of-returns]]). In year 30 it is cosmetic. Fat tails raise the odds that the fragile window **contains** a catastrophe. The danger is the product of the two risks. The practical corollary: the defenses of the fragile window (glide path, buffer, early income) are **also** the best anti-tail defenses, at no extra cost ([[glidepaths]], [[cash-buffer]]).

**Third effect: diversification delivers less than advertised.** The Gaussian case for diversification (average correlations reduce σ) misses one fact about crises: in the tail, correlations **rise**. In 2008 everything fell together except long government bonds and the yen. Diversification works, but it works **less** well in exactly the scenarios you bought it for. Unless you own assets whose decorrelation **survives** a crisis: government duration, gold, systematic managed futures ([[defensive-assets]], [[managed-futures]], [[all-weather-portfolios]]). A portfolio's kurtosis comes down through the **choice** of building blocks, more than through their number.

Two mistakes are left to avoid. The optimistic one says: "the median does not move, so tails are a detail." No: retirement planning is tail management. The safe withdrawal rate is defined by the worst cases, not by the median ([[the-trinity-study]]). The catastrophist one says: "then set df 3 everywhere and stop talking about it." No again. Stacking df 3 on top of an already prudent mean and a CAPE anchor is the triple-counted prudence this book keeps warning about ([[expected-returns]]). And a permanent df 3 world is not the real world either. Fitting to the data, plus one sensitivity test, beats both postures.

::: exemple The two-minute sensitivity test
Plan: $1.5M, $52,000 a year, 45 years, a pension starting in year 16. Note the central-case failure probability at the df fitted to your data (say 5): 5.2%. Push df to 30 and it drops to 3.1%. Take it down to 3 and it climbs to 7.8%. The reading: that 3 to 8% range is your exposure to disagreement about tails, comparable here to what half a point of μ does. If your decision (leave or not, [[one-more-year]]) survives both bounds, tails are not your dominant problem. If it flips, your plan rests on the world behaving itself. Structural protections (flexibility in writing, a buffer, defensive building blocks) are then a better remedy than a debate about df ([[choosing-your-strategy]]).
:::

## A little history of the idea, to settle the matter

The history is worth knowing, because it inoculates you against models that are too clean. As early as 1900, Bachelier founded mathematical finance on Gaussian Brownian motion. In 1963, Benoit Mandelbrot studied cotton prices. He showed that the tails were so fat that the variance itself looked barely defined, and proposed wild "stable" laws. Eugene Fama (1965) confirmed it on stocks. The profession, though, needs models it can compute with. So it settled on a pragmatic middle: keep the variance, but fatten the tails (Student-t, mixtures, GARCH). Praetz (1972), then Blattberg and Gonedes (1974), established that the Student-t fits returns remarkably well. The crises of 1987, 1998, 2008 and 2020 settled the cultural debate. LTCM, in 1998, went down precisely for trusting thin tails. "Fat tails" moved from Mandelbrotian heresy to basic fact, popularized by Taleb (*The Black Swan*). The modeling that won out in decumulation simulators (a fitted Student-t, plus regimes for the sequence) is the direct heir of that compromise: more honest than the Gaussian, more sober than the stable laws, and fitted to **your** data.

## The essentials

- Markets throw extremes far more often than the normal law allows: monthly kurtosis of 6 to 12 against 3. That is the signature of volatility clustering and regime changes, not an accident.
- The Student-t adds the missing parameter: df, the thickness of the tails. At identical σ, df 5 makes a −30% real year about 10 times more likely than df 30. You estimate df by inverting kurtosis ≈ 3 + 6/(df − 4) on **your** funds' monthly returns, which typically gives 4 to 6.
- Effect on the plan: the median does not move, failure probability rises, often by 30 to 80% in relative terms. Tails mostly worsen the fragile window, and the anti-sequence defenses are the anti-tail defenses too.
- Limits of the model: it is symmetric (real risk has a negative skew, covered by the sequence stresses and the lost decade) and memoryless (covered by the regimes). Diversification protects less in the tail, except for the blocks whose decorrelation survives a crisis.
- Practical habit: keep the df fitted to your data, run the 3/30 test once to learn your exposure, and if the decision flips, fix it with the structure of the plan, not with the parameter.

---

## Going further

- Mandelbrot, "The Variation of Certain Speculative Prices" (1963) and *The (Mis)Behavior of Markets* (2004): the founding act and its popular version.
- Fama, "The Behavior of Stock-Market Prices" (1965); Blattberg & Gonedes (1974) on the Student-t: the academic confirmation.
- Taleb, *The Black Swan* (2007): the general culture of tails, to be read with a critical eye.
- How df is estimated, and exactly how it is used: [[under-the-hood]], with [[using-the-fire-simulator]] for the manual.
- Next in this book: [[market-regimes]] (where the clusters come from) and [[making-monte-carlo-relevant]] (how tails and regimes combine in the central model).
