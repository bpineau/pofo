# The fixed inflation-adjusted withdrawal (Bengen): the benchmark rule
<!-- source: retrait-fixe-bengen @ 34578614d96c -->

The fixed inflation-adjusted withdrawal is the founding rule of the field ([[the-trinity-study]]). It is the zero point of every comparison. When a simulator prints a safe withdrawal, that is the rule it solves. When a study grades a dynamic rule, that is the rule it measures against. Being the yardstick makes one thing easy to forget: it is also a real strategy, and real people run it. It is the easiest rule to live with, the hardest to defend on paper, and still the right answer for a few specific profiles.

This article treats it as a working strategy rather than a historical object (the history is in [[the-trinity-study]], the plain-English version in [[the-4-percent-rule]]). We start with the exact mechanics and the execution details that move the results: where the money comes from, when it comes out, how you rebalance. Then its characteristic failure, and the variants that repair it without changing what it is (partial indexation, a freeze after bad years, an upward ratchet). Then the parameters the research recommends, the profiles it still suits, and how to run it in a simulator without picking the wrong settings.

::: cle The rule, operationally
Year 0: set the initial withdrawal R = rate × capital. The rate comes from your own analysis, 3 to 4% depending on horizon, valuations and margins ([[how-much-you-need]]). Every year after that: R ← R × (1 + realized inflation). The money comes out whatever the markets did, and you sell whatever sits above the target allocation, so the draw does the rebalancing for you. That is the whole rule. It listens to nothing: not the portfolio, not valuations, not your mood. That is its statistical weakness and its psychological strength.
:::

::: admin How to run it
- **The rate applies once, to the capital you hold on the day you retire.** This is the most widespread misunderstanding in the field, and it deserves to be written in black and white: the 4% is **never** recomputed on the current portfolio. Year 1 fixes an **amount in euros**, and that amount then lives its own life. This year's withdrawal-to-portfolio ratio is a warning light to watch, never an input to the calculation. Every rule in the articles that follow was born from exactly that observation: the current rate carries information, and the fixed rule refuses to read it.
- **Indexation.** Each year, amount ← amount × (1 + realized inflation), on the consumer price index. That is the rule's only update. The amended variant skips that line in the year after a negative portfolio return, and the adjustment is sometimes capped at 6% a year.
- **Frequency.** The amount is recomputed once a year, on a fixed date. The money itself comes out monthly. Two different cadences, and confusing them is expensive ([[the-annual-review]]).
- **Thresholds.** None, which is the definition of the rule. The amendments add some: an indexation freeze after a red year, an upward ratchet (+10% of the amount when the real portfolio passes 150% of its starting value, at most every three years, under a cap on the cumulative raise).
- **Floor.** Not applicable: the amount delivered **is** the floor. That is the whole difference with the rules that follow, where the floor becomes a parameter you have to write down yourself.
- **In your head.** R = rate × starting capital, once. Then each year R ← R × (1 + inflation).
:::

## The fine mechanics: three details that carry weight

The rule fits in two lines. But three execution choices, rarely spelled out, really do move the results.

**Where the money comes from: funding the withdrawal by selling the overweight.** The naive version sells a bit of everything, pro rata. Good practice sells the asset class that sits above its target weight first. After a good year for stocks, you sell stocks; after a crash, you sell bonds and leave stocks alone to recover. That one choice turns the withdrawal into free rebalancing, and it clearly improves survival in the worst vintages. It is one reason historical studies, which assume annual rebalancing, find higher SAFEMAX numbers than real portfolios that were never rebalanced. Almost every simulator makes the same assumption without saying so. They aggregate the portfolio at its target weights and draw from that aggregate, which amounts to rebalancing continuously and for free. So the number on the screen assumes you really do hold the target allocation.

**When: how often the money comes out.** Two conventions coexist: annual, at the start of the year, in the studies; monthly in real life. The gap in results is small, a few tenths of a point of failure probability, because monthly draws smooth out the entry point of the sales. But it has a governance virtue. It turns the withdrawal into a paycheck, and it removes the temptation to postpone the one big annual sale while watching the market, which is market timing in disguise.

**Which inflation: the index.** The canonical rule indexes on the national consumer price index. Your own inflation can drift away from it for years, because of health care, long-term care, or simply the weight of your own basket of spending ([[tracking-inflation]], [[spending-in-retirement]]). Indexing on the index while leaving the health-care drift unfunded is the most common blind spot in Bengen-style plans that run more than 30 years.

## How it fails: the silent cliff

What sets the fixed rule apart from all the others is the shape of its failure. You have to see that shape to understand both its danger and the monitoring it demands.

The rule never fails all at once. It fails through a silent squeeze: spending climbs with inflation while the portfolio stalls. In a hostile regime ([[market-regimes]]), the portfolio falls while the withdrawal keeps rising with prices. The current withdrawal rate (withdrawal divided by portfolio) then climbs from 4% to 6, 8, 12%. Past a threshold, empirically around 8 to 10% with no pension close at hand, even a market that turns generous again is no longer enough. The draws outrun any plausible growth, and the path is doomed years before it reaches zero ([[failure-probability]]). The 1966 vintage takes fifteen years to become unrecoverable, and nearly thirty to run dry ([[the-trinity-study]]).

::: figure bengen-falaise
The 1966 vintage on the book's US 60/40: EUR 600,000, EUR 24,000 a year indexed to inflation, never adjusted, everything in constant euros. The warning light starts at 4.0%, crosses 8% in January 1975 and never comes back down. The household gets nine annual statements before that alarm even sounds, then nineteen years to act before the capital runs out. None of this needs a model: one division is enough.
:::

That shape of failure has two practical consequences.

First, the fixed rule is the only one whose failure is perfectly visible along the way. The current withdrawal rate is a reliable warning light, graduated and readable by anyone. Whoever runs Bengen has to track that ratio once a year, against written thresholds ([[when-to-worry]]).

Second, nobody in practice walks off the cliff: facing a current rate of 7%, humans cut. So the fixed rule as it is really lived is almost always a fixed rule with implicit flexibility attached. What guardrail rules add ([[guyton-klinger]], [[morningstar-guardrails]]) is making that flexibility explicit, decided in advance, while calm, instead of improvised in fear. That is the deeper point of the comparison on the frontier ([[withdrawal-strategies-overview]]): pure Bengen is a theoretical point, and the real choice is between written guardrails and improvised ones.

::: science What stability costs, in numbers
The fixed rule buys perfectly stable income at a steep price: it is the rule that ties up the most capital. Here are the orders of magnitude from the literature. For the same 5% failure probability over 45 years, the fixed rule typically needs 10 to 20% more capital than well-bounded guardrails, and 20 to 30% more than smoothed amortization ([[amortization-based-withdrawal]]). The mirror image is just as real: at equal capital, it delivers the lowest average total consumption. It hoards in the good scenarios, and its bequest turns enormous as soon as the vintage is kind, often several times the original stake ([[spending-in-retirement]]). Stability is a luxury good, and the fixed rule shows its price tag. That is not an argument against it, because some households want exactly that. It is the price to know before you buy.
:::

::: figure bengen-millesimes
The same household as in the rest of this part, moved start year by start year across the century. Six neighboring start dates lost everything and twelve left more than three times the original stake, without a single one of them doing anything differently. The right margin shows those same thirty-three numbers sorted, which makes the one thing that matters here visible: between the two pathologies, the distribution is almost empty. The fixed rule rarely delivers the outcome it was asked for.
:::

## The variants that repair it without changing what it is

Three historical amendments keep the spirit, a stable target income, while filing down the pathologies. They run from the gentlest to the most active.

**Partial or capped indexation.** Bengen proposed this himself: adjust only at inflation minus a small gap (inflation − 0.5 point), or cap the annual adjustment (at 6%, say). The reasoning is simple. The great disasters of the fixed rule are inflationary regimes, precisely because indexation turns into an accelerator there ([[inflation-and-withdrawal-rates]]). Slowing the adjustment during those episodes is a small and nearly painless flexibility, a few percent of purchasing power spread over several years, and it lifts the SAFEMAX by roughly 0.25 to 0.5 point. A behavioral version is simpler still: freeze the nominal amount in the year after the portfolio falls. Almost invisible in daily life, clearly effective in the simulations. This is the minimum viable flexibility.

**Ratcheting.** This is the symmetric amendment, proposed by Kitces. Since a cautious fixed rule ends up rich in most scenarios, you allow yourself irreversible raises once the plan is plainly won. For example: +10% of the withdrawal, in real terms, each time the portfolio passes 150% of its real starting value, at most every three years. The ratchet barely degrades safety, since you only step up from the top of a cushion, and it repairs the pathology of unspent luxury. A ratchet is four numbers, written down before you start. The size of the raise, the trigger, the minimum wait between two raises, and a cap on the cumulative raise. The first three are above, and the trigger has an equivalent form that is often handier, a current withdrawal rate that has fallen very low. The fourth is almost always forgotten. With no cap, a long bull market installs a standard of living that the years after it will not fund, and the ratchet, irreversible by definition, becomes the problem it was meant to solve.

**An initial rate conditioned on valuations.** The third amendment leaves the rule alone and moves its starting point: set the initial rate from today's CAPE, 3 to 3.25% in an expensive market, 4 to 4.5% in a market that has been cleaned out, instead of a universal 4% ([[valuations-and-cape]]). That is half the way to full CAPE rules ([[cape-based-rules]]), without their dynamics. It is one decision, taken on the day you are at your most clear-headed.

A fixed rule fitted with all three (indexation frozen after red years, a capped upward ratchet, an initial rate conditioned on the CAPE) sits honorably on the frontier, at a fraction of the complexity of the dynamic rules. That is the version this book recommends to anyone who wants Bengen.

## Who it fits

The amended fixed rule is still the right answer in four recognizable situations:

- **The floor is nearly the whole budget.** If 90% of your spending cannot be compressed ([[how-much-you-need]]), flexible rules have almost nothing to cut: their edge evaporates, so take the stability and size the capital accordingly.
- **Governance comes first.** One person runs the money, the spouse is uncomfortable with it, the estate will be settled by third parties: this rule fits on a postcard and outlives its author ([[couples-and-family]]).
- **The withdrawal is already very low.** Below an initial rate of about 3%, every rule converges, since failure is close to zero everywhere ([[horizon-and-life-expectancy]]): sophistication stops paying and simplicity wins by default.
- **A short uncovered phase.** If your pension covers the floor within ten years, the bridge years are short ([[pensions-and-other-income]]): the fixed rule has only a brief window of vulnerability to cross, and its simplicity is worth the small extra capital.

The pure fixed rule, on the other hand, is the wrong choice for the typical FIRE profile: a 45-year horizon, an expensive market at the start, a floor well below comfort. That is exactly the profile where its cliff is most likely and where flexibility pays the most ([[flexibility-in-practice]], [[choosing-your-strategy]]).

::: exemple The amended Bengen, run over ten years
The plan: EUR 1.2M, an initial rate of 3.4% (a high CAPE) for EUR 40,800, a freeze after a red year, a +10% ratchet when the portfolio passes 150% in real terms. Years 1 to 3: ordinary markets, normal indexation, so EUR 42,900 in year 3 with cumulative inflation of 5%. Year 4 brings −22%. The withdrawal is frozen at EUR 42,900 nominal; the 3% inflation is not passed through, which costs −3% of purchasing power, imperceptible. Years 5 to 9: recovery, indexation restarts, and the current rate, up at 4.9% at the trough, falls back to 3.1%. Year 10: the portfolio reaches 158% of its real starting value, so the ratchet fires, +10%, or EUR 51,300 nominal. The tally: ten years, two decisions that took some thought (one freeze, one raise), zero cliff anxiety. The current rate was tracked every year and never came near the alert thresholds. That is what the founding rule looks like when it is run well.
:::

## Running it in a simulator

The fixed inflation-adjusted rule is the default behavior of almost every decumulation tool. You get it by setting a real annual spending number and putting every flexibility rule at zero. That is also why it is the right methodological starting point. Each dynamic rule is then measured by what it adds to this base, or takes away from it.

Three readings belong to it. The first is the safe withdrawal, the rule solved backwards. Instead of fixing the amount and reading the failure probability, you fix the failure probability you accept and read the amount the plan supports. The second is the first decade, because the fixed rule is the one most exposed to the order of returns. The gap between a central model and the sequence stress is at its widest here, wider than with any flexible rule ([[sequence-of-returns]]). The third is the ratchet, when the tool offers it as an option.

Indexation frozen after a red year, on the other hand, is rarely offered as such. The practical approximation is to turn on a small cut in downturns, around 5%, whose effect on failure probability is comparable: the control is "Cut in downturns (0 = fixed rule)". The matching settings are covered in [[using-the-fire-simulator]].

## The essentials

- Operationally: initial R = rate × capital, then inflation indexation, a draw that rebalances (sell the overweight), monthly if you can, with an eye on the gap between the index and your own inflation.
- It fails as a silent cliff driven by the squeeze, the withdrawal rising while the portfolio falls: visible years ahead through the current withdrawal rate, which has to be tracked against written thresholds.
- Perfect stability is expensive: 10 to 30% more capital than the rules that listen, for the same safety, plus a huge median bequest of luxury you never spent.
- Three nearly free amendments repair it: freeze the indexation after a red year, cap the upward ratchet, condition the initial rate on the CAPE. The amended Bengen is a legitimate strategy on the frontier.
- A good choice when the floor is nearly the whole budget, when governance comes first, when the rate is already below 3%, or when the uncovered phase is short. A bad choice for typical FIRE with a 45-year horizon in an expensive market: read on.

---

## Going further

- Bengen, "Determining Withdrawal Rates Using Historical Data" (1994) and *Conserving Client Portfolios During Retirement* (2006): the rule, and its own author's amendments to it.
- Kitces, "The Ratcheting Safe Withdrawal Rate" (2015): the upward ratchet.
- Early Retirement Now, Part 5 (the indexation adjustments) and Part 24 (minimum viable flexibility) ([[the-ern-series]]).
- In this book: the two upstream articles ([[the-4-percent-rule]], [[the-trinity-study]]), and the direct heirs [[guyton-klinger]] and [[floor-and-ceiling]].
