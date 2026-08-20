# Arithmetic mean, geometric mean, and volatility drag
<!-- source: rendements-arithmetiques-geometriques @ 2dc06d49e4c9 -->

Here is the most profitable trick question in all of personal finance: an investment gains 50% in its first year and loses 50% in its second. Average return: 0%. How much did you make?

Answer: you **lost** 25% (100 to 150 to 75). The "average" you were quoted is perfectly accurate and perfectly misleading. This page takes the mechanism apart. It explains the difference between the arithmetic mean and the geometric mean, and the volatility drag that separates them. That drag is the silent keystone of the entire FIRE subject. It is why the returns in the brochure are not returns you can live on, why volatility costs money even without a crash, why leveraged funds disappoint, and why the safe withdrawal rate sits so far below "stocks make 7%".

After this page, nobody gets to sell you an average again.

::: cle The two means
The **arithmetic** mean adds the returns and divides: (+50 − 50) / 2 = 0%. It answers "what does a typical year, taken on its own, pay?" The **geometric** mean compounds: √(1.50 × 0.50) − 1 = −13.4% a year. It answers "at what steady rate did my capital actually grow?" That is the CAGR, the only number that describes what happens to money left invested. The geometric mean is always at or below the arithmetic one, and the gap widens with volatility. Your wealth lives in geometric terms; brochures talk in arithmetic ones.
:::

::: figure vol-drag
Two investments with the **same arithmetic mean** (7% a year): one steady (+7% every year), one volatile (+27% and −13% in alternation, a volatility σ of 20 points). After 30 years the steady one is worth ×7.6 and the volatile one ×4.5, for the same advertised average. The gap is the volatility drag. Volatility on its own, with no crash anywhere, cuts about σ²/2 out of compounded growth, here 2 points a year.
:::

## Volatility drag: volatility is a cost

The gap between the two means follows a famous approximation that is remarkably accurate:

> geometric return ≈ arithmetic return − σ² / 2

The Greek letter σ (sigma) is the universal notation for **volatility**, the standard deviation of annual returns. It goes into the formula as a fraction, and that is where people trip: 15% volatility means σ = 0.15, so σ²/2 = 0.011, or 1.1 points of annual return gone. Not 15² / 2.

The σ²/2 term is exactly the **volatility drag**. It comes from an asymmetry everyone knows and nobody follows through on. After −20%, you need +25% to get back. After −50%, you need +100%. Losses weigh more than gains of the same size, and the wider the swings, the more compounding suffers.

You can redo the arithmetic on any holding you own, since both ingredients sit side by side in a portfolio visualizer. Over the past ten years, the MSCI World shows a CAGR of 13.0% and a volatility of 15.4% (in dollars, nominal). The drag is therefore 0.154² / 2 ≈ 1.2 points: the arithmetic mean of those ten years ran around 14.2%, while the capital itself compounded at only 13.0%. The missing 1.2 points did not go anywhere. They never existed outside the average.

A few ballpark numbers to calibrate the intuition (drag = σ²/2):

| Asset | Volatility | Volatility drag | Geometric result |
|---|---|---|---|
| Money market fund | ~1% | ~0.005% | 7.0% |
| 60/40 portfolio | ~10% | ~0.5% | 6.5% |
| Global stocks | ~15% | ~1.1% | 5.9% |
| Stacked 90/60 (efficient core) | ~15% | ~1.1% | 5.9% |
| Emerging market stocks | ~22% | ~2.4% | 4.6% |
| Stocks levered ×2 daily | ~30% | ~4.5% | ~2.5% (before the cost of the leverage) |

Every row starts from the same 7% arithmetic mean, to isolate the effect of volatility alone.

Compare the two leveraged rows, they say everything. The 90/60 stack carries 150% of exposure, and its volatility is still that of plain stocks: the leverage is spread across two lightly correlated engines and rolled in long-dated futures, never reset overnight ([[return-stacking]]). The daily ×2 piles its leverage onto a single asset and resets it every evening: it doubles the arithmetic mean but quadruples the variance.

::: figure drag-volatilite
The table, straightened into a curve. The cost of volatility is its square, so a table read row by row suggests a gentle slope where the truth bends. The six assets sit on the parabola at the volatilities the table quotes, and the two moves of the text are written under the axis. The two leveraged rows land at the ends of the second one: at plain-stock volatility the 90/60 stack keeps 5.9%, while the daily ×2 falls to 2.5%.
:::

Hence a result that surprises every beginner: a daily ×2 leveraged ETF on a volatile index can do **worse** than the index over a long stretch, even though it doubles it faithfully day after day. As soon as the extra drag exceeds the extra return, leverage destroys ([[leverage-and-margin]]). The same mechanics explain why diversification is the only free lunch in finance. Combining uncorrelated assets lowers σ without lowering the arithmetic mean, so it raises the geometric one. Diversification does not promise better average years. It promises better compounding. The full mechanism, rebalancing premium included, is in [[why-diversification-works]]; the assembly in [[all-weather-portfolios]] and [[defensive-assets]].

::: exemple Check it on two holdings
Holding A: +7% every year, never varying. Holding B: +27% and −13%, alternating. Its arithmetic mean is (27 − 13) / 2 = 7%, exactly A's, and every one of its years sits 20 points away from that mean: σ = 20 points.

Over 30 years, the capital does this:

- **A**: 1.07 to the 30th = **×7.61**.
- **B**: its years come in pairs, an up then a down, so 1.27 × 0.87 = 1.1049 every two years. Thirty years make fifteen pairs, hence 1.1049 to the 15th = **×4.47**.

A ends 70% richer than B, for the same advertised average. The 7% never shows up in B's arithmetic because it never went in: it is an average of years, not a growth rate.

Check it with the formula: the drag is σ²/2 = 0.20² / 2 = 0.02, or 2 points. B's predicted geometric mean is 7 − 2 = 5%, and 1.05 to the 30th = ×4.32, right next to the exact ×4.47. The σ²/2 approximation is an excellent back-of-the-envelope tool.
:::

## The cascade: from the brochure to the withdrawal rate

Volatility drag is the first step of a cascade that runs from the marketing number down to the number you can live on. Follow "stocks make 10%" all the way to the withdrawal rate, one step at a time:

1. **10% nominal, arithmetic**: the average of the years, the one in the brochures and the textbooks.
2. **minus the drag (~1.1% at 15% vol)** gives ~8.9% nominal geometric: what invested capital actually compounds at. This is the mean a portfolio visualizer reports as CAGR, never the arithmetic one.
3. **minus inflation (~2.5%)** gives **~6.4% real geometric**: the only currency that counts over 40 years ([[inflation-and-withdrawal-rates]]). Historically, global stocks have returned about 5% real geometric; diversified 60/40 portfolios more like 3.5 to 4.5%.
4. **minus fees and taxes** (0.3 to 1.5%, depending on the accounts and the funds you use, [[building-it-with-us-etfs]], [[us-taxes-in-the-withdrawal-phase]]).
5. **minus the sequence premium**: even a net real geometric return is only withdrawable if the returns show up in an orderly fashion; their lumpiness against fixed withdrawals costs another 1 to 1.5 points ([[sequence-of-returns]]).

Destination: 3 to 3.5% of sustainable rigid withdrawal over a long horizon, exactly the range modern research reaches by independent routes ([[the-4-percent-rule]], [[the-ern-series]]). That is no coincidence. The 4% rule is not mysterious, it is where this cascade lands once the accounting is done. Memorize the hierarchy: **arithmetic > geometric > real geometric > sustainable withdrawal rate**, with 0.5 to 1.5 points lost at every step.

::: attention The salesman test
When someone quotes you a return, ask the cascade's three questions every time. Arithmetic or geometric (compounded)? Nominal or real (net of inflation)? Gross or net of fees? "8%" can mean 6.9% compounded, 4.4% real, 3.4% net: less than half the headline number, with no formal lie anywhere. Backtests of structured products and fund averages almost always show the most flattering corner of the table. A professional who cannot answer those three questions has not understood their own product.
:::

## Three direct applications to FIRE

**1. Calibrating a simulator.** A serious simulator works in real terms. The μ it asks for is an expected real return, and the engine applies the volatility σ to generate the paths. The drag then shows up on its own in the results ([[under-the-hood]]). The classic trap is to type in a nominal arithmetic μ ("8%"): you have just built yourself a dream world. Rule of thumb: for a globally diversified portfolio, a real μ of 4 to 5% with a σ of 12 to 15% is the reasonable zone. The simulator prefills it from your own funds' history, then pulls it toward a prudent prior (a prior is the assumption you hold before looking at your own data, here the one a century of developed markets suggests), precisely to keep you out of that trap ([[making-monte-carlo-relevant]], [[expected-returns]]).

**2. Judging a withdrawal portfolio.** Two portfolios with the same arithmetic expectation are not worth the same: the less volatile one compounds better and stands up better to sequence, a double advantage. That is why serious withdrawal portfolios trade average return for steadiness (bonds, gold, diversification across regimes, [[stock-bond-allocation]], [[all-weather-portfolios]]), and why "100% stocks is optimal in the long run", fair enough in accumulation, does not survive the first withdrawal ([[ten-plan-wrecking-mistakes]]).

**3. Reading your own performance.** The "+9% a year over 5 years" on your statement is probably an arithmetic mean. The only honest number for your own use: (ending value / starting value)^(1/n) − 1, corrected for contributions (a good simulator computes IRR and CAGR properly on your real flows). Plenty of investors discover that their real compounded performance is 2 to 3 points below their impression. The difference goes to drag, to fees, and to badly timed contributions.

## For the curious: why σ²/2 exactly

No formalism needed. The logarithm of a return, ln(1+r), is what really adds up from one year to the next: compounding is adding logs. And ln(1+r) is concave. It punishes moves down more than it rewards moves up. Expand ln(1+r) ≈ r − r²/2, take the expectation, and the −r²/2 term produces −σ²/2 on average. This is Jensen's inequality turned into accounting. The drag is neither a market friction nor a hidden fee. It is a mathematical property of compounding, as unavoidable as compound interest itself, and exactly its dark side. Models work natively in compounding, period by period, so the drag is captured there by construction rather than by approximation ([[under-the-hood]]).

## The essentials

- Two means: the arithmetic one (the typical year) and the geometric one (the growth you actually live); your capital lives in geometric terms, marketing speaks in arithmetic ones.
- The gap is the volatility drag ≈ σ²/2: volatility is a cost of compounding, with no crash required, even when the advertised average is zero.
- The cascade from brochure to livable: minus drag, minus inflation, minus fees, minus the sequence premium; 3 to 3.5% of rigid withdrawal survives it, which is the 4% rule demystified.
- Diversifying raises the geometric mean at an equal arithmetic mean. That is the mathematical case for a diversified withdrawal portfolio; leverage does the opposite.
- Three reflex questions in front of any number: compounded? real? net? And for your simulators, a **real** and modest μ, never the brochure's average.

---

## Going further

- Early Retirement Now, SWR Series part 8 (the technical appendix) and part 33 (computing a withdrawal rate without simulation): the cascade, formalized ([[the-ern-series]]).
- William Bernstein, *The Intelligent Asset Allocator*, chapter 1: the best written introduction to the two means.
- Markowitz (1952) and the modern "geometric mean maximization" reading (Kelly, Latané): why you maximize the geometric mean and not the arithmetic one.
- Next in this book: [[sequence-of-returns]] (the next step of the cascade) and [[expected-returns]] (which values of μ are defensible today).
