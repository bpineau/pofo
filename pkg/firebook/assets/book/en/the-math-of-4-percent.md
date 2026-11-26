# Why 4%? The mathematical anatomy of the rule
<!-- source: les-maths-du-4-pourcent @ c33b6e0e6db9 -->

[[the-4-percent-rule]] traces the rule's origins and [[the-trinity-study]] explains how it was tested. But why 4% if US stocks earned more? Start with a portfolio that earns exactly 4% after inflation every year. How much can you withdraw if you want to preserve the capital? What if you have thirty years to spend it?

Arithmetic answers the first two questions. For the third step, where returns vary from year to year, we turn to the historical record.

::: cle The idea in one sentence
With a **constant** 4% real return, you can withdraw 4% of the starting portfolio at the end of every year indefinitely, or about 5.8% a year for thirty years. In Bengen's US study, a withdrawal starting around 4% and then adjusted for inflation survived the difficult thirty-year periods. The gap from 5.8% reflects both the returns those retirees earned and the order in which they earned them.
:::

## Tier 1: the perpetuity, or real return as the natural ceiling

Start with $100,000 earning exactly 4% a year after inflation. At the end of year one it holds $104,000 in today's purchasing power. Withdraw $4,000 and the original $100,000 remains. You can do the same thing next year ([[inflation-and-withdrawal-rates]]).

Why use 4% here? It is a rough historical real geometric return for a portfolio of 60% stocks and 40% bonds. The **geometric** mean is the constant rate that would have produced the same ending balance with dividends reinvested and no withdrawals. The **arithmetic** mean of the annual returns is generally higher ([[arithmetic-vs-geometric-returns]]). Once you start withdrawing money, two portfolios with the same geometric return can finish with quite different balances.

## Tier 2: spending down the capital over thirty years

With thirty years to fund, you can spend some of the capital. At a constant 4% real return, $100,000 funds thirty end-of-year withdrawals of about $5,783 in today's money. After the last one the balance is zero. Spending down the capital adds about 1.8 percentage points to the initial withdrawal rate. This is the **amortization bonus** that ABW and VPW recalculate as the remaining horizon shortens ([[amortization-based-withdrawal]], [[vpw]]).

The bonus also explains why the horizon matters so much ([[horizon-and-life-expectancy]]). Stretch it to 50 years and the same formula gives 4.7%: the bonus melts to +0.7. Stretch it to infinity and it disappears. Very early retirement does not lose rate to some mystery. It simply loses the right to amortize fast.

::: figure amortissement-horizon
The withdrawal that pure amortization arithmetic funds at 4% real, by the number of years to cover. The curve falls fast, then flattens onto its asymptote, the return alone. The first twenty years carry most of the bonus, while stretching from a 50-year horizon to eternity costs only seven tenths of a point. Very early retirement does pay for its horizon, but far less than people think.
:::

## Tier 3: when the bad years come first

Start with $100 and withdraw $10 at the end of each year. If the market falls 20% then rises 25%, you have $70 left after the first withdrawal and $77.50 after the second. Reverse those returns and you have $115, then $82. With no withdrawals both orders would have taken you back to $100. With withdrawals, an early loss leaves less money to benefit from the recovery ([[sequence-of-returns]]).

Bengen compared people who retired in different years in the United States. They experienced different thirty-year returns and faced high inflation at different stages of retirement. The roughly 1.8-point gap between our theoretical payout and the historical withdrawal rate reflects those differences. To isolate the effect of order alone, keep the same real returns and rearrange them, as in the $100 example.

::: figure cascade-4pct
The first two steps assume a constant return. The last connects that calculation to the US historical result. The bar labelled “sequence” also reflects the level of returns during difficult periods.
:::

::: exemple The whole cascade, back of the envelope
In this example we go from a 4% real return to a 5.8% payout over thirty years at a constant return, then to roughly 4% in the historical test. The next figure illustrates the scale of changes when we vary one assumption at a time. A fifty-year horizon gives about 3.4% here. A global sample ([[anarkulova-cederburg]]) lowers the rate by 0.5 to 1 point. A high starting valuation ([[valuations-and-cape]]) trims a point from the assumed return in this scenario. Fees of 0.5% lower the illustrated payout rate by close to half a point. A flexible rule with a spending floor adds 0.3 to 0.5 point, in exchange for possible spending cuts. If your plan differs in several respects, rerun it with all those assumptions together.
:::

::: figure clavier-leviers
Each assumption moved on its own, starting from the cascade's reference plan: historical 60/40, 30 years, a fixed indexed withdrawal, so 4.0%. The ranges are the ones used in this chapter, and the sign of each bar says which way the lever pushes.
:::

## Why it holds up (and what would break it)

A rate chosen to survive the most difficult US starting date often leaves considerable wealth to households that retired in better years ([[reading-a-fan-chart]]). Why keep that margin? When you retire, you do not know which group you will join. Another household's ending wealth cannot cover your expenses if your own retirement gets off to a poor start.

Three checks bring the historical result closer to your situation. Use returns from the markets you actually invest in ([[international-diversification]]). Deduct annual portfolio fees and apply taxes to the cash flows they affect ([[building-it-with-us-etfs]], [[us-taxes-in-the-withdrawal-phase]]). Then check how long the withdrawals must last and which expenses you could actually cut if markets disappoint ([[flexibility-in-practice]]).

::: science What modern reconstructions confirm
The studies listed below answer different questions. Early Retirement Now compares withdrawals by starting valuation and asset allocation ([[the-ern-series]]). Blanchett and his coauthors examine how spending can adjust over retirement. Anarkulova and Cederburg broaden the sample of markets. Before comparing two published rates, check the horizon, the markets included and the definition of failure.
:::

## The essentials

- At a constant 4% real return, thirty annual end-of-year withdrawals can each amount to about 5.8% of initial capital in constant purchasing power.
- The historical rate of roughly 4% comes from replaying retirements. Comparing that outcome with a constant-return payout cannot isolate sequence risk.
- To see the effect of order by itself, rearrange the same returns while keeping withdrawal amounts and dates fixed.
- Spending indexed for fifty years, recurring fees or investments in different markets each call for a fresh calculation.
- When reading another published rate, look for the horizon, the markets studied and whether the retiree could cut spending along the way.

---

## Going further

- William Bengen, [“Determining Withdrawal Rates Using Historical Data”](https://www.financialplanningassociation.org/article/journal/MAR04-determining-withdrawal-rates-using-historical-data) (1994). The charts show how long portfolios lasted by starting year and initial withdrawal rate.
- Early Retirement Now, the SWR series (the parts on the CAPE and on the horizon in particular): the cascade recomputed on modern data ([[the-ern-series]]).
- Blanchett, Kowara & Chen, "Optimal Withdrawal Strategy for Retirement Income Portfolios" (2012) and Morningstar's State of Retirement Income reports: the amortization bonus in rates by age.
- In this book: [[arithmetic-vs-geometric-returns]] (tier 1 in detail), [[amortization-based-withdrawal]] (tier 2 turned into a strategy), [[sequence-of-returns]] (tier 3), [[the-4-percent-rule]] (the rule told as a story) and [[the-trinity-study]] (the rule measured).
