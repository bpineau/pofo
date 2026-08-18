# Drawing on a buffer and refilling it: the rules that work
<!-- source: recharger-ou-pas @ b83c25419e3c -->

Arguments about the cash buffer always land on size. Two years, or three? Size is the least important variable in the whole thing. What a buffer is worth is decided by its flows: when you draw on it, when and how you refill it, and when you stop keeping it up. Two households can hold the same 30 months of cash. One gets real protection against a bad sequence out of it, the other an expensive parking account. The rules make the entire difference.

This is the plumbing chapter, and it follows [[cash-buffer]], which covers the why and the sizing. It runs through the candidate triggers for drawing on the buffer and what goes wrong with each, the three good sources for refilling it and the one absolute prohibition (never refill while under water), the case, often the strongest one, for not refilling at all (the melting buffer of the first decade), and how all of this fits with the withdrawal rule and with rebalancing. It ends on what a model has to represent before it can settle an argument between two sets of rules. The defaults you see everywhere (draw past ten points of drawdown, refill gradually with a monthly cap, stop refilling on a date) did not come from nowhere: they are what the literature recommends. Knowing them is what lets you depart from them on purpose.

::: cle The three golden rules of the flows
One. You draw on the buffer on a written trigger, a drawdown threshold, never on a mood. Draw too early and you burn it before the real hole arrives. Draw too late and it protects nothing. Two. You refill in calm weather, gradually, and never during a decline. Refilling at the bottom means selling depressed stocks to build up cash, the exact sin the buffer exists to prevent. Three. A buffer can legitimately never be refilled. It melts away over the first decade and then it is gone, delivering protection while protection matters and shutting off its opportunity cost once it does not ([[sequence-of-returns]]).
:::

## Drawing on it: choosing the trigger

The trigger decides everything. Here are the candidates.

**Portfolio drawdown (recommended).** "Withdrawals switch to the buffer for as long as the portfolio sits more than X% below its real high." This is the clean trigger. Reading it leaves no room for argument, it points straight at the mechanism you are fighting (selling depressed shares), and it disarms itself, since climbing back above the threshold puts withdrawals back on the portfolio. Setting X takes some judgment. Too low (5%) and the buffer fires on every breath the market takes, emptying itself before the real holes. Too high (30%) and it watches the ordinary declines go by, and those are where most of the cumulative damage is done. The sensible zone runs from 10 to 20%, and 10% is the usual starting point. Lean toward 15 to 20% with a small buffer, to save ammunition for the real crossings, the years spent inside a bear market, and toward 10% with a large one.

**The down year (acceptable, blunter).** "After a year of negative returns, the next year is funded from the buffer." Simple, and you can run it in your head, but it is all or nothing. A year at −2% fires it, while an 18% crash in June waits until January. Continuous drawdown does better for the same effort.

**The current withdrawal rate (keep it for steering the plan).** "Go to the buffer when the current rate passes 4.5%." That indicator belongs to the withdrawal rule and to the guardrails ([[when-to-worry]], [[morningstar-guardrails]]). Using it for the buffer as well mixes two levels of decision together. Keep the instruments separate.

**Discretion (never).** "I'll know it when I see it." This is the version that turns the buffer into a totem ([[the-bucket-strategy]]). Under stress, people draw too early, because an 8% correction feels frightening. Or they refuse to draw at all, out of the well-documented reflex that says protect the buffer ("that's my safety net, don't touch it"), and sell stocks at the bottom instead. The instrument then produces precisely what it was bought to prevent.

Two refinements are worth adding. The first is hysteresis: switch back to the portfolio only once the drawdown has fallen well below the threshold. Arm at 15% and disarm at 10%, say, and you stop bouncing back and forth around the line. The second is partial drawing: move half the withdrawal onto the buffer first, then all of it past a second threshold. It stretches the ammunition through a long crossing ([[market-regimes]]).

## Refilling: three good sources, one prohibition

**Source 1: calm weather, a little at a time.** This is the standard refill. Outside a drawdown, a slice of each month's sales goes to the buffer until it is back at target. It models easily: the refill runs only while the trigger is disarmed, and it is capped every month. Never one big rebuilding sale. Going slowly matters. Rebuilding 30 months of buffer in a single quarter concentrates exactly the timing risk that spreading it out dilutes.

**Source 2: new highs (the strict version).** "Refill only when the portfolio prints a new real high." Stricter than "outside a drawdown", and it guarantees the refill never sells anywhere but at the top. It has a cost. After a long crossing, the buffer can sit half empty for years waiting for that high. "Outside a drawdown" is the workable compromise, "at the highs" the ideal for a large buffer. Either way, inside the portfolio you sell out of the sleeves sitting on gains, not the ones on the floor ([[gold-in-retirement]]).

**Source 3: the surplus, with nothing sold.** The most elegant of the three, because it requires no sale at all. Budgeted withdrawals you never spent, side income ([[going-back-to-work]]), dividends thrown off by an old taxable account, one-off windfalls: it all goes straight to the buffer as long as the buffer is below target. The timing cost is zero, and governance improves. The buffer becomes the default home for good surprises.

**The prohibition: never while under water.** It earns its own paragraph, because the temptation is real and it dresses up well. "The buffer is low, markets could fall further, let's get safe while we still can." Do the arithmetic. Selling at −20% to build up cash locks in the loss on the shares you sold and forfeits their rebound, the exact double penalty the whole setup exists to fight. A buffer that ends a crossing empty has not failed. It has done its job, and it will rebuild in calm weather. Writing the ban into the plan in plain words ("no refill while the drawdown exceeds the threshold") costs one line and saves the strategy.

::: science Never refilling at all: the melting buffer
The most underused option is also the most logical one. Sequence risk concentrates in the first decade ([[sequence-of-returns]]), so a buffer that melts lines the cost of the protection up with the years that protection is worth paying for. You draw on such a buffer when you need it, you never refill it past a set date, and the portfolio then absorbs what is left. It is the cash version of the glide path ([[glidepaths]]), and the two combine naturally, since a melting buffer is the first step of the climb back into equities. It takes one model parameter: the year the refill stops. Refills run normally through year N, typically 8 to 12, then stop. The overcoat becomes a jacket, then nothing at all. In simulation, a melting buffer usually beats a permanent buffer of the same size. Same protection while it counts, no opportunity cost once it stops counting. It is what this book recommends by default for the uncovered phase, the bridge years before your pensions start.
:::

## Fitting the pieces together: buffer, withdrawal rule, rebalancing

All three mechanisms live in the same plan. Write down which one goes first, so they stop doing each other's work twice.

**Where the money comes from in a decline.** With the trigger armed, withdrawals come first from the buffer, then, once it is empty, from the overweight bond sleeve, where funding the withdrawal by selling the overweight keeps working ([[bonds-in-retirement]]), and from stocks last. The buffer is the first line, not the only one. Rebalancing does the same job right behind it. That is why a modest buffer is enough ([[cash-buffer]]).

**How it meshes with flexibility.** If your rule cuts spending in a drawdown ([[guyton-klinger]], [[floor-and-ceiling]]), the buffer and the cut usually fire together. That is coherent enough, since the buffer funds the reduced floor and therefore lasts that much longer. But add the two effects up while you are still designing. A plan with guardrails and a melting buffer needs a smaller buffer than a rigid plan does, 18 to 24 months against 30 to 36. Flexibility is already a piece of the buffer ([[flexibility-in-practice]]).

**What a model has to represent.** Four numbers describe a buffer's flows, and a tool that exposes only one of them, the size, is not testing your rules but a caricature of them. You need the target size ("Buffer (years of spending)"), the real return of the vehicle ("Buffer real return", set from your money market fund and never at zero), the drawdown threshold that arms the drawing, and the year the refill stops ("Buffer refill stops in year"). The rest comes out of comparisons. Replay the same plan with and without a stop date to settle melting against permanent, then sweep the size to choose inside the 18 to 36 month range ([[cash-buffer]]). Half an hour of simulations beats a year of arguing about two years versus three.

::: exemple The whole rulebook, on a postcard
Salomé, 49, with a 15 year uncovered phase ahead of her. "Buffer: 24 months of the floor, in a money market fund. Drawing: withdrawals switch to the buffer when the real drawdown passes 15%, and switch back below 10%. Refilling: outside a drawdown only, out of a fifth of each month's sales, plus every surplus; target 24 months; never a refill while under water, even at a buffer of zero. Stop: no refill at all after year 10, the buffer melts and the portfolio absorbs it." Four sentences, and everything is in them: the trigger, the hysteresis, the source, the prohibition, the end date. The simulation backs it up, with failure equal to a permanent 36 month buffer at a third less opportunity cost. The postcard goes into the written plan, next to the withdrawal rule ([[building-your-plan]]).
:::

## The essentials

- A buffer's value sits in its flows, not in its size: a written trigger, a disciplined refill, a scheduled end. With no rules it is an expensive parking account, or a totem you protect by selling stocks at the bottom.
- **Drawing**: on a real drawdown of 10 to 20%, with hysteresis, and in two steps if you like. Never on a mood, and never on an indicator that belongs somewhere else, since the current withdrawal rate steers the withdrawal rule, not the buffer.
- **Refilling**: outside a drawdown, a little at a time out of each month's sales, at the highs if you want the strict version, and out of the surplus always. Never while under water. A buffer that comes out of a crossing empty has done its job, and it will rebuild in calm weather.
- The melting buffer, with refills stopped after year 8 to 12, lines the cost up with the fragile window and usually beats the permanent version. One date models it, and it is the setting this book recommends for the uncovered phase.
- Order of play: the buffer first, selling the overweight second, stocks last. With a flexible rule, cut the size to 18 to 24 months, because flexibility is already buffer.

---

## Going further

- Early Retirement Now, Part 12: the case against cash, aimed squarely at buffers with no rules ([[the-ern-series]]).
- Michael Kitces on bucket maintenance strategies: the plumbing of the refill, compared.
- In the simulator: the buffer settings and the sweep over its size ([[using-the-fire-simulator]]); the machinery underneath is described in [[under-the-hood]].
- In this book: [[cash-buffer]] (the why and the size), [[glidepaths]] (the melting buffer's twin in the allocation), [[the-bucket-strategy]] (what buckets become with no flow rules).
