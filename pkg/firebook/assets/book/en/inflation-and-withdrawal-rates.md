# Inflation and the withdrawal rate: the exact link
<!-- source: inflation-et-taux-de-retrait @ fa7b9fcccff9 -->

Everyone knows inflation is bad for retirees. Almost nobody can say **where** it attacks a withdrawal plan. Five years at 8% do incomparably more damage than thirty years running a point above plan, even though the cumulative loss of purchasing power is the same.

::: figure inflation-episode-derive
A deterministic run on the book's reference household ($600,000, $24,000 a year indexed to prices, 30 years), with the same nominal asset returns in both worlds: only the shape of the inflation changes, and the two price paths end at the same level. Yet the plan's warning light enters the alert zone as early as year 19 under the episode, against year 29 under the drift.
:::

This chapter traces the link, mechanism by mechanism. The first is the **squeeze**: spending climbs with inflation while the portfolio stalls, so the plan is attacked from both ends at once. The second is the compression of real returns during an episode, when almost everything loses ground in real terms **at the same time**; that correlation is what makes inflation episodes the worst vintages on record, 1966 ahead of 1929. From there come the conditional numbers, what withdrawal rates are actually worth depending on the inflation regime you retire into, a question ERN's series has measured. The payoff is an audit: the **indexation inventory** of your plan, which sorts what follows prices, what does not, and what follows them in reverse, so you can read your true net exposure. That leaves what a simulation in real terms already covers, and what you have to ask of it on top.

::: cle The technical heart of it
A plan simulated **in real terms** already carries average inflation. Working in constant money is the convention of every serious simulator. Returns are deflated, withdrawals are stated in constant purchasing power, and routine inflation is neutralized by construction. What real terms do **not** neutralize is the **risk of an episode**. An inflation episode is not "prices going up": it is a regime in which the **real** returns of nearly everything turn negative together for years ([[market-regimes]]). Seen from a simulation, it looks like a long run of correlated, very bad real years. Only models with memory can produce that run. Draw years independently and you underrepresent it ([[historical-vs-parametric]]).
:::

## Mechanism 1: the squeeze

Take the fixed inflation-adjusted rule ([[fixed-inflation-adjusted-withdrawal]]) through an episode. Inflation at 8% raises the withdrawal by 8% a year: that is the contract, and purchasing power holds. The nominal portfolio does not keep up, because bonds lose price and equity multiples compress. The current withdrawal rate, the withdrawal divided by the portfolio, then climbs at both ends: the numerator rises mechanically, the denominator falls in real terms. That is the squeeze, and its violence is pure arithmetic. Three years at 9% inflation with a flat nominal portfolio take a current rate of 4% to about 5.2%, the equivalent of a 23% crash, without a single day of crash.

This is why 1966 is a worse vintage than 1929 in every study ([[the-trinity-study]]). The 1929 crash is brutal, then it gives back: deflation pulls nominal withdrawals down with prices, and indexation starts working in your favor. The 1966 to 1981 episode never gives anything back. Fifteen years of squeeze.

::: figure voyants-1929-1966
The plan's warning light, year after year, for the same indexed 4% withdrawal starting in 1929 and then in 1966. The 1929 one rings for three years, tops out at 6.3% in 1932, then comes back down: real capital rebuilds while deflation pulls the nominal withdrawal down with it. The 1966 one crosses 8% in 1975 and never comes back, leaves the scale in 1988 and empties the account in 1991. A detail that confirms the thesis: the 1929 vintage's true maximum is not its crash, it is the post-war inflation of 1949.
:::


The corollary for withdrawal rules follows at once. The anti-inflation amendments to the fixed rule go straight at this mechanism: cap the indexation, or freeze it after a red year ([[fixed-inflation-adjusted-withdrawal]]). Giving up 2 to 3 points of indexation during an episode disarms half the squeeze at a modest cost in purchasing power, spread over years. It is the best pain-to-protection trade in the whole flexibility toolkit ([[flexibility-in-practice]]).

## Mechanism 2: everything loses in real terms at once

The second mechanism belongs to regimes ([[market-regimes]]). During an episode the **real** returns of almost every asset class go negative together. Nominal bonds do it by definition: a fixed coupon runs behind prices. Cash and money market funds do it through financial repression, because the rate they pay follows with a lag. Stocks do it **during** the episode, through multiple compression: the US CAPE goes from 24 in 1966 to 7 in 1982, nominal earnings keep running but prices do not ([[valuations-and-cape]]). Classic diversification stops diversifying anything. This is the setup where a 60/40 has no sleeve that wins. 2022 was the flash reminder, 1973 to 1981 the long version.

For the plan, the two mechanisms together draw the profile of the enemy: negative real years, correlated across assets, persistent, with a liability growing on the other side. Read that sentence again. It is the exact definition of the worst case of sequence risk ([[sequence-of-returns]]). Inflation is not one plan risk among many. It is the historical first cause of the worst sequence case, and the two chapters describe the same animal from two angles.

::: science The conditional numbers: which regime do you retire into?
ERN gave the conditional question two parts, Part 41 on low-inflation environments and Part 51 on retiring into high inflation. The result runs against intuition. The **level** of inflation on the day you retire says almost nothing about the sustainable rate; what shaves it is the inflation of the years that **follow**, on the order of 0.2 points of withdrawal per point of inflation over ten years. The worst US vintages are therefore mid-1960s starts, 3.8 to 4.2% even over 30 years: retirements begun in low inflation and caught by the Great Inflation. The best are the bargain-market starts of the early 1980s, around 6 to 8%, when disinflation was already under way. The big conditioning variable you can read on day one is still the CAPE ([[valuations-and-cape]]), and the two subjects overlap, since episodes compress the CAPE.

The subtler point sits elsewhere. **Low** inflation is not a safe regime in itself. It comes with low rates and low expected returns ([[expected-returns]]), so the sustainable withdrawal rate gets compressed from the other end. In practice you **read** the inflation regime (the CPI, the breakevens, [[tracking-inflation]]) and let it tune your opening caution, without giving it the weight of the CAPE and without ever pretending it forecasts 30 years out.
:::

## The audit: your plan's indexation inventory

Here is the article's central practical tool. Since the enemy attacks through the gap between indexed liabilities and nominal assets, take inventory of **both** columns of your plan:

| Line of the plan | Indexation | Note |
|---|---|---|
| Spending and withdrawals | **Indexed** (the plan's contract) | Plus your personal drift ([[tracking-inflation]]) |
| Social Security | **Indexed** (the annual COLA, set by law on the price index) | The plan's number one inflation asset, though the raise is always argued over politically ([[us-healthcare-and-social-security]]) |
| Private pensions and annuities | **Mostly not** (a fixed payout, at best a fixed step written into the contract) | Corporate plans almost never index; a genuinely inflation-linked annuity exists and you buy it with a deep cut in starting income ([[annuities-and-safety-first]]) |
| TIPS, an indexed ladder | **Indexed** (by contract) | The clean hedge ([[inflation-linked-bonds]]) |
| Rents you collect | **Nearly** (they follow prices with a lag, and rent control can cap them) | The living linker, with regulatory risk ([[real-estate-in-retirement]]) |
| Global equities | **No** in the short run, **yes** in the long run | They suffer during and re-rate after: the slow protection |
| Nominal bonds, stable value, cash | **No** | The victims of the squeeze and of financial repression |
| Gold, real assets | Episodically | The crisis hedge, not the one for normal times ([[gold-in-retirement]]) |
| A fixed-rate mortgage still outstanding | **Negative indexation** | Inflation pays it down for you: the one line that **likes** an episode ([[real-estate-in-retirement]]) |

Your net exposure is right there in the table. The typical plan (a deferred, indexed Social Security check, a nominal 60/40, some cash) is **long** inflation in its later years, thanks to Social Security. It is badly **short** during the uncovered phase, the bridge years, when everything else pulls the other way, and that is exactly the stretch where the sequence decides the outcome ([[horizon-and-life-expectancy]]). The allocation conclusion writes itself. The uncovered phase is the one to index, with TIPS, held both as a sleeve and as a floor ladder, a slice of real assets, and a budgeted drift. The covered phase is already indexed, by the pension.

## What a simulation tests, and the checklist

Every idea in this chapter can go on a test bench, so here is the tooling. Running in real terms from end to end neutralizes average inflation, and there is nothing more to do about it. Your personal drift you declare by hand. It takes two settings any good simulator exposes: a real slope added to spending (the **"Real spending drift /yr"** control, where 0.3 to 0.5 points a year covers the health drift) and a spending profile by age, the retirement smile ([[tracking-inflation]], [[spending-in-retirement]]). Episode risk shows up only in the models that have memory.

Start with the resampled multi-country century, Anarkulova and Cederburg's data, what simulators call the **broad sample** ([[anarkulova-cederburg]]). Its blocks contain the real stagflations: Britain and Italy in the 1970s, France after the war. It is the only model where this chapter's enemy exists natively ([[historical-vs-parametric]]). Next comes the **sequence stress**, with its clusters of negative real years, agnostic about what caused them. Vintage replay closes the set: 1966 and 1973 are your two named inflation tests, and every US historical replay tool runs them cohort by cohort (cFIREsim and FICalc, for instance).

The checklist is four questions. Does the plan hold under the multi-country century? If the failures are inflation blocks, the answer is in the indexation inventory, not in more capital. Does it get through the 1966 vintage? Is your personal drift budgeted? And is the floor of the uncovered phase actually indexed, with TIPS and a ladder, or only hoped to be?

::: exemple The same plan, before and after the indexation audit
The starting plan: $1.6M, $52,000 a year, 45 years, with an indexed pension of $21,000 arriving in year 16. The portfolio starts at 65% stocks and 35% nominal (stable value and the aggregate). The read is 4.3% failure in the central case and 11.2% in the broad sample. Two thirds of the broad sample's failed paths are inflation blocks, and the 1973 vintage replay breaks in year 24. The audit is not close: the uncovered phase is 100% short inflation.

Apply the inventory. Take 8% in short TIPS, an approximately indexed ladder covering six years of floor, and 5% in gold, all of it out of the nominal sleeve. Budget a spending drift of 0.4 points a year above the index. Write the indexation freeze after a red year into the rule. New read: 4.1% in the central case (no better, insurance does not pay in normal times), 7.8% in the broad sample, and 1973 survived. No capital was added. The plan simply stopped being short its main enemy.
:::

## The essentials

- Two mechanisms combine. The **squeeze** lifts indexed withdrawals while nominal assets stall: three years at 9% inflation equals a 23% crash with no crash. **Simultaneous real compression** makes almost everything lose real value at the same time. Together they are the worst case of sequence risk, 1966 ahead of 1929.
- Deflation runs the other way for an indexed retiree: withdrawals fall with prices and bonds rise. That asymmetry is why the hedging budget goes to inflation, with deflation insurance kept to a small dose ([[bonds-in-retirement]]).
- Simulations in real terms carry **average** inflation, not the risk of an **episode**. You test that in models with memory, above all the broad sample, the only place the real stagflations live, and in the 1966 and 1973 vintage replays.
- The practical tool is the **indexation inventory**: Social Security and TIPS indexed, private annuities and nominal bonds not, a fixed-rate mortgage indexed in reverse. The typical plan ends up long inflation, thanks to Social Security, and dangerously short during the uncovered phase. That is the phase you index, with TIPS, a ladder, real assets and a budgeted drift.
- The indexation amendments to the withdrawal rule (a freeze after a red year, a cap) disarm half the squeeze at a cost spread thin over years. It is the best pain-to-protection trade flexibility offers.

---

## Going further

- Early Retirement Now, Part 41 (low inflation) and Part 51 (high inflation): SAFEMAX conditioned on the regime ([[the-ern-series]]).
- The 1966 to 1981 vintage as it appears in every historical study ([[the-trinity-study]]): the reference episode, worth knowing in detail.
- In this book: [[tracking-inflation]] (measuring it, and setting the drift), [[inflation-protection]] (the defenses one by one), [[inflation-linked-bonds]] (the contractual hedge).
- How a simulator implements these test benches, broad sample and vintages included, is described in [[under-the-hood]].
