# The cash buffer: size, cost, and what it is really for
<!-- source: cash-buffer @ 22f8a9a94963 -->

No protection in the whole withdrawal phase feels more obvious than this one. Keep two or three years of spending in cash, live off it when markets dive, and never sell stocks at the worst possible moment. The instinct is so strong that a cash buffer shows up in every amateur plan and every advisor's pitch. Which is exactly why the numbers deserve a hard look, because the research is surprisingly lukewarm. Simulate the thing mechanically and the buffer barely moves the failure probability: the opportunity cost of holding cash gives back roughly what the sequence protection earns. Simulations say so plainly. Vary the size of the buffer and the failure curve usually comes out **flat**.

So, a gimmick? No. But what it is good for is not what most people think.

This chapter works through the whole file: why the instinct is right and the net effect still small, what separates a **useful** buffer from a decorative one, since the rules on drawing and refilling decide everything, then the right size and the right place to park it, the head-to-head against the glide path, and last the buffer's real value, which is behavioral and about governance. That value earns its place easily, as long as you do not overpay for it.

::: cle The buffer paradox in one sentence
Every euro in the buffer is a euro taken out of the engine. Protection against the bad sequences is paid for with forgone return in all the others. Good sequences being the majority, the net comes out near zero, plus or minus 0.5 points of failure depending on the rules. So the statistics do not justify a buffer. What justifies it is what it does to you: a retiree who sleeps, who does not improvise in a crash, and who executes the rule. In practice that is worth more than a lot of simulation points ([[the-psychology-of-spending]]).
:::

## The instinct, and why it is not enough

**The instinct rests on real math.** What kills a withdrawal plan is selling depressed shares in the holes. Every sale at the bottom turns a temporary loss into a permanent one ([[sequence-of-returns]]). A buffer that absorbs withdrawals through the crossing, the years spent inside a bear market, removes exactly those sales. On paper it is the perfect anti-sequence weapon.

**But funding the buffer costs just as much.** Two to three years of spending is 7 to 10% of a plan drawing 3.5%, and that slice earns about 0 to 1% real instead of the portfolio's 3 to 5%. Call it 0.25 to 0.4% a year on the whole plan, paid every year, including the thirty years when no crash ever bites. A simulation counts both sides. ERN (Part 12, "Six reasons to be suspicious about the Cash Cushion") finds the net slightly negative for naive static buffers. The friendlier studies, the ones with smart rules, find it slightly positive. The test that settles the argument is a sweep: replay the same plan with a buffer of 0 to 10 years, everything else held fixed, and plot the result. Depending on the plan, the failure curve comes out flat or with a soft optimum around 2 to 3 years. Read it on two outputs, though, not one. Failure barely moves while median ending wealth falls with every year of buffer added. That is the real trade, insurance paid for in ending wealth rather than in survival odds. The honest quantitative verdict fits on one line. A well-run buffer is roughly free, and almost never decisive.

::: figure buffer-flat
Failure probability against buffer size, in its typical shape. The curve is nearly flat: across the whole range, less than one point of failure separates the best size from the worst. There is a soft optimum around 2 to 3 years, and past it the curve turns back up, because too much buffer starves the engine more than it protects it. Quantitatively the buffer is close to neutral. Its real value is behavioral.
:::

**The gap between a strong instinct and a small effect has two deep causes.** The first is the arithmetic of duration. A crossing runs for years, counting the climb back to the old real high ([[market-regimes]]). On the real US 60/40, the nine episodes since 1953 that fell more than 10% lasted from 16 months to more than 10 years, with a median of 32 months. In a simulation the same duration shows up in the distribution of time spent below the last real high, path by path, and that distribution is what sizes a buffer. Two years of buffer covers half the months spent underwater, three years covers two thirds. It moves the sales at the bottom around; it does not remove them all.

::: figure traversees-matelas
The nine episodes of the real US 60/40 (S&P 500 and 5-year Treasuries, rebalanced every January, deflated by CPI, 1953 to 2026) whose real decline passed 10%, from the high to the day that high is recovered. The 10% threshold is the drawing trigger written further down, since below it the buffer does not move. Five crossings out of nine fit inside 36 months; the other four leave 136 months underwater with nothing covering them.
:::

The second cause is that rebalancing already does half the job. A rebalanced 70/30 portfolio naturally draws from bonds during an equity crash ([[fixed-inflation-adjusted-withdrawal]], [[bonds-in-retirement]]). An explicit buffer therefore layers on top of a mechanism that is largely there already, and the returns diminish fast.

## What separates a useful buffer from a decorative one

The same pile of cash can be an instrument or a totem. Four choices decide which.

**1. A written rule for drawing on it.** The decorative buffer has no rule. You "feel" when to use it, which means improvising under stress. Too early, at the first 10% correction, burning the buffer before the real hole arrives. Or too late. The useful buffer has a numeric trigger: withdrawals switch to the buffer when the portfolio is more than 10 to 20% below its real high, and they stay there until it climbs back above that threshold. Simple, mechanical, and something a spouse can run ([[when-to-worry]], [[couples-and-family]]).

**2. The refill rule.** The topic is rich enough to deserve its own chapter ([[refilling-the-buffer]]). The principle is enough here. A drawn-down buffer is rebuilt at the highs, never in the hole, because refilling by selling depressed stocks cancels the whole benefit. And a buffer that never refills becomes a slice of prepaid spending. That is legitimate too, but it is a different object, the bridge of the early years, a cousin of the ladder ([[bond-ladders]]).

**3. The size: 18 to 36 months, no more.** Below 12 months the effect is cosmetic. Past 3 or 4 years the opportunity cost grows in a straight line while the marginal protection collapses. Crossings longer than 5 years, two of the nine counted above, are too long to prefund in cash, and that job belongs to flexibility and to the regime assets ([[flexibility-in-practice]], [[defensive-assets]]). The sweep above settles the number on your plan rather than on an average one. The soft optima almost always land in the 2 to 3 year range.

**4. Where to park it: paid, liquid, insensitive to rates.** A checking account is not a buffer. A Treasury money market fund or a short T-bill ladder is, with a high-yield savings account for the first few months of spending and a stable value fund if a workplace plan offers one ([[bonds-in-retirement]]). Choosing the vehicle well cuts the buffer's opportunity cost roughly in half, the cheapest improvement in this chapter; which account holds it is a separate question ([[us-accounts-and-account-order]]). The buffer never goes into long bonds, because staying insensitive to rates is its definition. And never into stocks, obviously.

::: attention How you fund it decides the verdict
Before comparing two failure numbers, settle a question of method. Is the buffer carved out of the starting capital, or stacked on top of it? Carving it out is the only honest convention. Three years of buffer on a EUR 1.5M plan then means about EUR 150k in the buffer and EUR 1.35M in the engine, with total wealth still the sum of the two. That convention is what makes the buffer pay its opportunity cost, and it is why the failure curve can turn back up as the buffer grows. Stacking it on top compares two plans of different sizes instead, which makes the buffer free by construction and flatters every version of it. Two assumptions also belong in the open. First the buffer's real return, which is not zero for a decent money market fund net of everything (about 0.5 to 1%) and moves the verdict. Then the year the refill stops, because a buffer limited to the first decade tracks where sequence risk is concentrated ([[sequence-of-returns]]).
:::

## The real value: what simulations never count

The statistical net is close to zero, and this book recommends a buffer anyway, because three real services escape the simulations, and they happen to be the scarcest ones.

**It stops the panic.** A simulation applies the rule without emotion; a human does not. The classic behavioral disaster, selling everything in March 2009 or March 2020 "to save what is left", costs 20 to 40% of ending wealth. Nothing debated in this chapter is on that scale. And the trigger is well identified: the fear of having to sell in order to eat. A retiree who knows the next 30 months of groceries are already sitting in a money market fund watches the same crash with a different nervous system. The buffer is a structural sedative, and its power against capitulation is the best documented of its virtues ([[the-psychology-of-spending]], [[bear-markets-in-retirement]]).

**It gives permission to spend.** The mirror service, and an underrated one. Retirees underspend badly out of fear of what comes next ([[spending-in-retirement]]). A visible buffer unlocks the spending that was already planned. The trip gets booked, because the money is already there.

**It makes the household governable.** Nothing else in a plan hands over as cleanly as a written instruction. "If the weather turns, we live off the account at X" is something a non-managing spouse can carry out with no help ([[couples-and-family]]). No sophisticated withdrawal rule has that property.

The assembly rule then writes itself. Size the buffer at the effective minimum, 18 to 36 months, parked in a money market fund, rules on paper, and you buy all three behavioral services at the lowest statistical price. Refuse the monumental buffers of 5 to 10 years, which buy the same services at triple the cost. That leaves the glide path, which covers the same risk through the allocation ([[glidepaths]]). The two overlap, and a moderate combination, a gentle slope plus two years of buffer, beats the extreme version of either.

::: exemple A buffer with written rules, and the decade it lived through
A EUR 1.5M plan, EUR 52,000 a year, Vanguard-style dynamic spending. The buffer is EUR 130k (30 months) in a money market fund, with the rules on paper. Draw on it as soon as the real drawdown passes 18%. Refill at new real highs out of the surplus withdrawals, up to 30 months, never while under water. In simulation, failure goes from 4.1% without the buffer to 3.8% with it, the honest almost-nothing you would expect. The decade as it was lived tells another story. Year 3, a 26% crash: withdrawals sit on the money market fund for 19 months, and not one share is sold below −18%. That is the whole point, no sleepless nights and no urge to "move everything to safety" at the bottom. Years 5 and 6, back at the highs, refilled out of the surplus. Year 8, a 12% correction: nothing happens, the threshold is not crossed, the buffer does not move. The 0.3 points of failure saved were the footnote. A decade executed without panic was the product.
:::

## The essentials

- The instinct, never sell at the bottom, is right, and the arithmetic is stubborn. The opportunity cost of cash gives back roughly what the protection earns, for a net of plus or minus 0.5 points. A sweep over buffer size shows it plainly, as long as the buffer is carved out of the starting capital.
- A useful buffer has four attributes: a written trigger for drawing on it (a drawdown past 10 to 20%), a refill rule tied to the highs ([[refilling-the-buffer]]), a size of 18 to 36 months, and a money market or T-bill home with no duration in it.
- Its real value sits outside the simulation. It stops the panic, and the behavioral disaster weighs far more than any argument about size; it gives permission to spend; it makes the household governable. Buy those three at the effective minimum, not at the most reassuring maximum.
- Crossings run for years (a median of 32 months on the real US 60/40, more than 10 years in the worst case), and two years of buffer covers half the months spent underwater. The rest belongs to flexibility, to rebalancing and to the regime assets. A 5 to 10 year buffer pays three times over for the same thing.
- The buffer and the glide path cover the same risk, so a moderate combination beats the maximum of either. Model the buffer properly, with the real return of the vehicle and the year the refill stops, so the test lands on your version rather than on a caricature.

---

## Going further

- Early Retirement Now, Part 12 ("Six reasons to be suspicious about the Cash Cushion"): the founding case for the prosecution ([[the-ern-series]]).
- Michael Kitces, "Managing Sequence Of Return Risk With Bucket Strategies Vs A Total Return Rebalancing Approach": the reconciliation of buffer and rebalancing.
- In the simulator: the buffer sweep ("Buffer (years of spending)", with "Buffer real return" and "Buffer refill stops in year" as the two assumptions above) and the distribution of years spent underwater size the buffer on your own crossings ([[using-the-fire-simulator]]).
- In this book: [[refilling-the-buffer]] (the buffer's flow rules), [[the-bucket-strategy]] (the tiered version, and the case against it), [[glidepaths]] (the alternative through the allocation).
