# Glide paths: the bond tent, rising equity, and the fragile window
<!-- source: glidepaths @ c85b4b44a644 -->

The previous chapter settled where to sit on the stock-bond axis ([[stock-bond-allocation]]). This one adds the dimension a static allocation ignores: time. Risk in a withdrawal plan is not spread evenly. It piles up in the first five to ten years, the fragile window of sequence risk ([[sequence-of-returns]]). A constant allocation, though, pays for protection at the same price for life. That price is the equity return you give up. And the danger eventually passes.

Hence the glide path, a term borrowed from target-date funds: let the allocation move over time. Cautious while the danger is at its worst, rising in stocks as the window closes. The full shape has a name, Michael Kitces's bond tent. You raise the bond sleeve as the exit approaches, then bring it back down through the first decade of retirement. The result is a peak of caution planted exactly on the red zone. This is one of the few ideas in the field that every camp signs off on (Pfau-Kitces 2014, ERN Parts 19-20 and 43, institutional practice). On one condition: knowing its true proportions. The benefit is targeted, it aims at the worst cases, and it is not free. Running it means buying stocks through years when nobody wants any.

This chapter gives the numbers, how to run each leg, the honest criticisms, the comparison with the alternative (the cash buffer), and how to test it in simulation.

::: cle The idea in one sentence
The plan's fate is decided in the first decade, so caution belongs there too, as a temporary expense rather than a permanent tax. You start cautious, 50 to 60% stocks on day one, then climb steadily back to your long-run allocation (70 to 100%) over 10 to 15 years. That path captures most of the protection a permanently cautious allocation buys, for a fraction of what it costs in return.
:::

::: figure bond-tent
The equity share falls before the exit (you de-risk with new money), bottoms out on day one, then climbs back through the first decade (withdrawals drain the bond sleeve and the equity share rises on its own). The bond sleeve, the upper band, draws a tent. Its peak sits on the fragile window: maximum caution exactly where sequence risk peaks, and nowhere else.
:::

## The founding results, with numbers

**Pfau-Kitces (2014).** Their paper, "Reducing Retirement Risk with a Rising Equity Glide Path", tests 30-year retirements where the allocation climbs in a straight line, say from 30 to 70% stocks, against the static allocations that match them. It runs in Monte Carlo under three sets of return assumptions, from very conservative to long historical averages. The results read in two parts. In median scenarios the glide path barely beats, and sometimes ties, the static portfolio with the same average exposure. In the worst scenarios, the vintages that open with a crash, it clearly lifts both the sustainable rate and ending wealth, by as much as 0.2 to 0.4 points of SWR in the unfavorable setups. The mechanism is obvious once you see it. If the crash lands early, the dangerous case, the retiree rides it out at maximum caution, then buys stocks back at sale prices on the way up. Buying low, mechanically. If the crash lands late, the retiree has been rich for years and nothing much matters ([[horizon-and-life-expectancy]]). A rising glide path is a scheduled purchase of stocks, concentrated on the decade where buying low pays the most.

::: figure tente-transfert
The gap in sustainable withdrawal rate, vintage by vintage: a tent that starts at 58% stocks and climbs to 85% over ten years, minus the static allocation with the same average exposure (80/20). Over the 43 start years the reconstructed real US 60/40 allows (S&P 500 and 5-year Treasuries, deflated by CPI, thirty-year retirements beginning between 1954 and 1996), nine vintages gain and the median vintage pays a quarter of a point, while the ten hardest start years all land in the favorable half: the path moves results around, it does not manufacture them.
:::

**ERN (Parts 19-20), the FIRE version.** Jeske runs the same exercise over a 60-year horizon, on monthly data. Same conclusion in principle, different calibration: over 60 years, the Pfau-Kitces 30 to 70% path starts too low and finishes behind most static allocations. The winning shape starts around 60% stocks and climbs toward 100% over roughly ten years. The gain in SAFEMAX for the worst vintages runs about 0.1 to 0.3 points. Keep the scale in mind ([[choosing-your-strategy]]): that is a lot for protection that costs so little in expectation. One nuance matters. The gain peaks when starting valuations are high ([[valuations-and-cape]]), which is exactly when an early crash is most likely. So the glide path is insurance whose premium falls as the risk rises, which is rare.

**The upstream leg: ERN Part 43 and the full tent.** The same logic applies before the exit. The last five years of accumulation already belong to the fragile window: a crash one year from target destroys the departure date ([[the-three-phases]]). Hence Kitces's full tent. You take stocks from roughly 80% down to 50 or 60% over the 5 to 10 years before you leave, and that is the declining leg, funded by pointing new contributions at bonds, selling nothing. Jeske doses it shorter: staying 100% in stocks right to the end is perfectly defensible as long as the departure date is still negotiable, and the de-risking, when it happens, fits in the last two to five years. Caution peaks on day one, then the rising leg takes over. The profile draws a tent whose peak sits on the maximum of risk. It is the anti target-date fund: that one descends forever and leaves an 80-year-old at 25% stocks, below the plateau, exposed to erosion ([[stock-bond-allocation]]).

## Running it, leg by leg

**The rising leg (retirement) comes down to three decisions.** How steep: a straight line over 10 to 15 years is the standard, 60 to 90% stocks, 2 to 3 points a year. Shorter concentrates the benefit but makes the execution abrupt. How to do it: the climb costs nothing to run, because the withdrawals do it for you. Take every withdrawal from the bond sleeve and the portfolio drifts toward stocks on its own. No stock sales for ten years, which is precisely the anti-sequence service you were after; rebalancing only resumes once the long-run allocation is reached. And where it stops: write down the target allocation and stop there. A glide path is not a permanent drift.

**The declining leg (the transition) asks for a different discipline.** Run it with flows, new savings into bonds and money market, building the cash buffer along the way ([[cash-buffer]]), rather than with big sales. Two reasons: taxes, and the risk of doing the very market timing you claim to avoid. Over five years at a high savings rate, redirecting contributions is usually enough to go from 85/15 to 60/40 without selling a single share, and without triggering a tax bill on the way.

::: attention The real difficulty is behavioral
Read again what the rising leg asks of you: raise your equity share every year, through the first decade of your retirement. Including, and above all, if that decade contains a crash. Buying stocks at 62, with no paycheck, in the middle of 2008, because the plan says so. It is the most counterintuitive thing this book asks anyone to do. And the typical failure of a glide path is not mathematical, it is human: the climb gets "suspended" at the first deep hole, hardens into permanent caution, and then pays full price in erosion. The defenses are known. Automation through the withdrawals, described above: the climb happens by itself if you draw on bonds, with no buy order to place. And writing the path into the plan, with the same standing as the withdrawal rule ([[building-your-plan]], [[the-psychology-of-spending]]).
:::

## Glide path or cash buffer: two tools for the same window

The glide path and the cash buffer ([[cash-buffer]]) aim at the same risk, selling stocks at the bottom during the fragile window, through two different mechanisms. Mixing them up, or stacking them without noticing, is common. Side by side:

| Criterion | Rising glide path | Cash buffer |
|---|---|---|
| Mechanism | The **allocation** moves over time | A dedicated **sleeve** absorbs withdrawals at the bottom |
| Cost in expectation | Low (caution is temporary) | Moderate (2 to 3 years of spending at ~0% real, permanently, if kept up) |
| Benefit | Concentrated on the worst early vintages | Concentrated on holes of 1 to 3 years; not enough on its own for a 7-year regime ([[market-regimes]]) |
| Governance | Automatic through the withdrawals; asks you to **stay on** the path | Very intuitive ("I live off the cash"); asks for rules on drawing and refilling ([[refilling-the-buffer]]) |
| Psychological value | Low (it is invisible) | Very high (you sleep) |

Simulate both protections on the same plan and the honest reading comes out like this. On the numbers they are worth about the same, and their benefits overlap heavily. A big buffer plus a deep tent means paying twice for the same insurance. On everything the numbers miss, the glide path wins on automatic execution and the buffer wins on psychology. The sensible combination fits on one line: a moderate tent (60 to 85% over ten years) plus a modest buffer (18 to 24 months), rather than the maximum of both ([[choosing-your-strategy]], where prudence gets a budget).

## The honest qualifications

**The average gain is modest. This is insurance, not alpha.** The scale again: 0.1 to 0.3 points of SWR in the bad cases, about zero at the median. Those bounds come from the Pfau-Kitces and ERN simulations. On the reconstructed US record used here, the median is even slightly negative, as the vintage-by-vintage gap above shows, which is the same verdict, only harsher. A glide path replaces neither a sane starting rate ([[valuations-and-cape]]) nor a flexible rule ([[choosing-your-strategy]]). It adds to them, nothing more. Anyone selling it as a revolution has not read the papers.

**Conditioning on the entry point is legitimate.** Since the gain grows with starting valuations, the refined version sizes the depth of the tent on the CAPE on day one. Below a CAPE of 20, a light tent or none at all: an early crash is unlikely and the opportunity cost is real. Above 30, the full tent. That fits the whole frame of this book, and it keeps you from pitching a tent out of ritual in a market that has already been cleaned out.

**How a pension changes it.** The climb in stocks is even easier to defend when a pension starts partway through ([[pensions-and-other-income]]). A discounted pension is a bond position that grows as it gets closer ([[amortization-based-withdrawal]]). Raising the equity share of the visible portfolio then does no more than hold total risk constant. Anyone whose pension or Social Security starts in fifteen years has two reasons to follow the path.

## What the tent is made of

The results above say a great deal about stocks and almost nothing about the rest. In Pfau-Kitces the non-equity sleeve comes down to a few average parameters, a return, a volatility and a correlation with stocks. Those averages cover a century of US history and its successive regimes without ever telling them apart. In Jeske, as in the vintage chart above, the sleeve is an actual intermediate Treasury, its historical returns replayed vintage by vintage. Both samples do contain both worlds, the disinflationary golden age and the inflationary stretch from 1966 to 1981. So the path survives the hostile regime, which is reassuring. But it survives with that sleeve. Nothing in this work says any bond fund would do the same job.

Replay the same tent on the same US history and change nothing but the material of the sleeve, and the question splits in two. What the path adds does not depend on it: the median vintage pays between 0.25 and 0.29 points, the worst vintage between 1.09 and 1.14, whether the sleeve holds bills, 5-year Treasuries, long government bonds or half and half. The level of the sustainable rate does depend on it, and its sign has flipped exactly once in forty-three start years. The long sleeve cost on all twenty-seven departures from 1954 to 1980 and paid on the sixteen that followed, which puts it on the wrong side. At the worst vintage in the sample, 1966, it takes away a quarter of a point of sustainable rate where bills add a little.

::: figure tente-matiere
Same tent, same history, one thing changes: the material of the defensive sleeve. On top, what each departure could sustain on 5-year Treasuries. Underneath, what bills and the 20-year government bond add to that rate or take from it. The long sleeve costs on all twenty-seven departures from 1954 to 1980 and pays on the sixteen that follow. One sign change in forty-three vintages, and it lands on the side where departures were tightest. The sample holds one bond bear market and one bull, which is too little to call it a law; it is enough to show that the material is not neutral.
:::

And the stock-bond correlation is not a constant. It is a regime variable, tied to inflation ([[why-diversification-works]], [[bonds-in-retirement]]). Campbell, Pflueger and Viceira (2020) trace its 2001 switch to a change in the relationship between inflation and economic activity. Brixton, Ilmanen and their AQR coauthors (2023) turn that into a working rule. The sign depends on the relative weight of growth shocks and inflation shocks, and on how the two move together. Negative from 2000 to 2021, the correlation has turned positive in episodes since 2022.

Which leaves the tent with a problem of its own. Its bond peak is not a long-term investment. It is the reserve that will fund the first ten years of withdrawals, and its useful horizon gets a year shorter every year. The rule matching duration to horizon therefore applies here in its strictest form ([[bonds-in-retirement]]). Short to intermediate duration, a linker sleeve against surprise inflation ([[inflation-linked-bonds]]), or a ladder held to maturity across the first years of spending ([[bond-ladders]]). That material works in both regimes. Long duration is something else, a deliberate bet on deflation, spectacular in 2008 and punished in 2022 ([[defensive-assets]]). It has a place in the portfolio, just not at the peak of the tent.

That leaves the backups. Trend following ([[managed-futures]]) and gold win in the regimes where nominal bonds lose, and 2022 showed it live. Their case is shorter and leans harder on the vehicle than the case for government bonds. They strengthen the tent's defensive sleeve, they do not replace the path. The founding result is about equity exposure through the fragile window, and it holds. What 2022 made visible is that the composition of everything else is a separate decision, and it gets made with the table of defenses ([[defensive-assets]]).

::: astuce Setting a glide path in a simulation
Not every simulator can move the allocation over the life of a plan. Where the option exists, it takes three numbers, and they are exactly the ones in the written plan: the starting equity share, the ending one, and how many years the climb takes. The control is usually labeled something like "Rising-equity glidepath (bond tent)". The rest is a matter of reading it right. Run the same plan with and without the path, and look first at the hard models, sequence stress and broad sample, because the benefit lives there and almost nowhere else ([[historical-vs-parametric]]). Then check how sensitive failure is to the returns of the first decade: it should visibly soften, and that is the signature of a glide path doing its job. If the tool plots failure against the number of buffer years, size both protections on the same plan rather than one at a time, and their overlap shows up in the numbers. None of it helps unless the path makes it into the written plan, and then into the annual review that keeps it alive ([[the-annual-review]]).
:::

::: exemple A tent sized end to end
Iris, 44, targeting 49, with a CAPE above 30: the full tent is justified. Declining leg (44 to 49). She sits at 85/15 today; every dollar of new savings ($2,800 a month) goes into intermediate bonds and linkers, plus 24 months of spending in money market in the final year. On day one: 58/34/8 (stocks/bonds/cash), no sales, no tax friction. Rising leg (49 to 61), written into the plan: "every withdrawal comes out of the bond sleeve until stocks reach 85%, then band rebalancing; no market event suspends the climb". Simulated at a 3.6% rate with guardrails, the tent cuts failure from 8% to 6% under sequence stress and from 12% to 10% under the broad sample, and the worst decile of first decades becomes far less lethal. The cost: about 0.15 points of expectation for twelve years. Iris signs. That is exactly what insurance costs, and she is the policyholder most likely to claim on it.
:::

## The essentials

- Risk sits in the first decade, so caution should sit there too: start at 50 to 60% stocks and climb back to your long-run allocation over 10 to 15 years (the rising leg), after cutting equities with flows over the five years before (the declining leg). That is Kitces's tent, peak on day one.
- The benefit is targeted: 0.1 to 0.3 points of SWR in the worst vintages, about zero at the median, largest when the starting CAPE is high. In expensive markets it is insurance with a negative premium, not alpha.
- The rising leg runs itself if every withdrawal comes out of bonds during the climb, so there is no counterintuitive order to place. That is also the defense against its real weakness, which is behavioral: the path suspended at the first crash.
- A glide path and a cash buffer cover the same risk, so a moderate combination of the two beats the maximum of each.
- The path is about stocks. What the defensive sleeve is made of is a separate decision, and 2022 made it visible, and the reconstructed record confirms it: the material does not change what the path buys, it changes the level of the sustainable rate. That sleeve funds the first ten years of withdrawals, which argues for short to intermediate duration, a linker sleeve, or a ladder ([[bonds-in-retirement]], [[inflation-linked-bonds]], [[bond-ladders]]). Long duration stays deflation insurance, and gold and trend following are regime backups ([[defensive-assets]], [[managed-futures]]).
- The anti-model: the target-date fund that descends forever and leaves an eighty-year-old below the allocation plateau, exposed to erosion ([[stock-bond-allocation]]).

---

## Going further

- Pfau and Kitces, "Reducing Retirement Risk with a Rising Equity Glide Path", *Journal of Financial Planning* (2014), and Kitces, "The Bond Tent" ([kitces.com](https://www.kitces.com)): the founding pieces.
- Early Retirement Now, Parts 19-20 (glide paths in retirement, the 60-year horizon version) and Part 43 (the run-up to retirement) ([[the-ern-series]]).
- Pfau, *Retirement Planning Guidebook*, the chapter on dynamic allocation.
- Campbell, Pflueger and Viceira, "Macroeconomic Drivers of Bond and Equity Risks", *Journal of Political Economy* (2020): why the stock-bond correlation changes sign with the macro regime.
- Brixton, Brooks, Hecht, Ilmanen, Maloney and McQuinn, "A Changing Stock-Bond Correlation", *The Journal of Portfolio Management* (2023): the practitioner reading of the same thing, and what it changes for allocation.
- In this book: [[sequence-of-returns]] (the risk being targeted), [[cash-buffer]] (the alternative), [[the-three-phases]] (the calendar), [[stock-bond-allocation]] (where the path ends up), [[bonds-in-retirement]] (what the defensive sleeve is made of), [[defensive-assets]] (the regime backups), [[lifecycle-theory]] (human capital, which derives glidepaths instead of decreeing them).
