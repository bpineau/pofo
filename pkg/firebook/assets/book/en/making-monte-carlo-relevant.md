# Making a Monte Carlo relevant: blending, regimes, stress
<!-- source: rendre-monte-carlo-pertinent @ c372d5d62724 -->

A naive Monte Carlo (independent Gaussian draws, parameters copied from recent history, a cartoon of a plan) is worse than useless. It is convincing. It turns out handsome fan charts and probabilities to three decimals, all resting on assumptions that understate a retiree's risk every single time ([[simulator-traps]]).

The good news: every one of those flaws has a known fix, documented in the research and perfectly implementable. This page walks through the six fixes that turn a random number generator into a planning instrument, in the order they apply. Calibrate the inputs without inheriting the bias of your own window (blending, the most important idea here and the least known). Fatten the tails. Put market memory back in (regimes). Anchor to valuations. Keep the raw data as a judge. And simulate the real plan rather than a cartoon of it.

The six also make a scorecard. Faced with a simulator, the question worth asking is not how many paths it draws, but which of these fixes it applies, and which it quietly leaves out.

::: cle The guiding principle
There is no "best model". There is an honest **central** model (fitted, corrected, pulled toward prudence wherever the information runs out) and the **bounds** around it: the optimistic one, your own data replayed; the pessimistic ones, the sequence stress, the world century, the lost decade. Making a Monte Carlo relevant is not about finding the truth. It is about assembling the weight of the evidence and deciding inside it. Everything below is how that evidence gets built.
:::

## Fix 1: blending, or how not to believe your own window

Start with the basic problem of fitting. Your portfolio's parameters (μ, σ, df) can only be estimated on its own history, say twenty years. Twenty years is almost nothing for a mean. The standard error on μ runs at about σ/√n, which is plus or minus 2.5 points at σ = 11%. So an estimate of "6.8% real" really says "somewhere between 4% and 9%". And it is not merely imprecise, it is **biased**: your window is the one that carried you to your target, so it was probably a kind one, and it holds one or two market regimes out of the four that exist ([[market-regimes]]).

The classic statistical answer is shrinkage, the James-Stein intuition. When a single estimate is noisy, pulling it toward a broader reference mean always improves the forecast. Here that means blending the parameters fitted to your own funds with a **prior** drawn from a vastly deeper sample, the world's long-run experience (μ around 4.5% arithmetic real, σ around 13%, df around 4, the prudent values the developed century suggests, [[anarkulova-cederburg]]).

That leaves the weight of the blend. One simple rule doses it honestly: **the prior's weight grows with the amount by which the horizon outruns the history you have**, capped at 50/50. The logic is plain. With 20 years of data for a 20-year plan, your data speaks from experience. For a 45-year plan, it has seen nothing of the 25 years past the end of its window, and the prior has to speak for the unknown. With 20 years of history and a 45-year horizon, the blend sits at the cap: half your funds, half the century. Your portfolio then reaches the central model through its statistics, never through its particular sequence. Its measurable virtues (diversification, contained volatility) get credit; its luck of the window does not.

Two practical notes. Blending applies to μ, σ and df alike, so the prudence reaches the tails too. And it is no tyrant: a parameter you typed in yourself always wins, and if you built your μ from the building blocks ([[expected-returns]]), that is the one to simulate. Blending is only the sensible default for anyone who would rather not decide.

## Fixes 2 and 3: the tails, then the memory

**The tails.** Replace the normal distribution with a Student-t whose df is fitted to your funds' monthly kurtosis ([[fat-tails]]). The fix is local, it costs one parameter, and its effect is direct: disaster years get their true frequency back inside the central model, not just in some side stress scenario. That is a real design choice. Plenty of tools keep a Gaussian central case "for readability" and bury the tails in an expert mode nobody ever opens.

**The memory.** The corrected central model is still i.i.d. Its bad years fall at random, never in clusters, while the real hard stretches come as episodes: 2000-2002, 2007-2009, 1973-1974, and their longer cousins. Hence the sequence stress, a two-state Markov chain, normal and "bear", built with three precise properties. One, bear markets persist: entering one is rare, staying in it is likely. That yields episodes of about three years and roughly 19% of years spent in a bear market, in line with history. Two, volatility is amplified in the bear state (× 1.5): crises are turbulent, not merely down, and that is where the negative skew we see in the data lives ([[fat-tails]]). Three, and this is the decisive methodological point: the model is **mean-preserving**, its long-run average being exactly that of the central model. The stress hides no pessimism about the level. It isolates the risk of **order**, surgically. The gap in failure probability between the central model and the stress is, by construction, the price of sequence in your plan, and nothing else ([[sequence-of-returns]]). It is a controlled experiment, impossible to run on real data.

The extreme variant rounds out the kit: the lost decade, a long deep bear market in the style of Japan in the 1990s, deliberately left uncompensated (the mean is pulled down). This is no longer an estimate, it is a crash test. The question you put to it is not "what is my failure probability?" but "can this be survived?", and the answer you want is not a comfortable percentage but a plan for getting through ([[bear-markets-in-retirement]]).

## Fix 4: the anchors, or the information in the present

Blending corrects the past, which is to say your window. It stays blind to the present: where are valuations today? The same portfolio does not carry the same expectation at a CAPE of 20 and at a CAPE of 38 ([[valuations-and-cape]]), and no historical average, however well blended, holds that information. Hence anchoring, which rewrites the model's parameters with information from outside your portfolio. Two anchors are worth the trip.

**The CAPE anchor** replaces the central model's mean, and only its mean, with what today's CAPE implies (about 1/CAPE for the equity block), leaving σ and df at their fitted values. This is the forward-looking fix: the central model stops assuming the decisive decade will look like the average decade, and assumes it will look like what current prices allow. In an expensive market it is the harshest fix of the six. That is as it should be, since it is the one carrying the bad news. Few consumer tools offer it. TPAW Planner is the notable exception, with equity expectations derived from the current CAPE, either raw 1/CAPE or a regression on that same 1/CAPE.

**The broad-sample prior** rewrites all three parameters at once with the prudent values of the world century, which amounts to blending pushed to 100%. Useful as a bound, or as a default stance for anyone who distrusts every fit to recent data.

One safeguard has been named elsewhere but shows its full reach here: these mechanisms **do not stack** ([[expected-returns]]). Automatic blending, a manual μ, the CAPE anchor and the broad-sample prior are four ways to calibrate the same central model. Pick one, and let the others serve as cross-checks. Prudence gets a budget; it is not a virtue you can pile up.

## Fixes 5 and 6: the raw data, and the real plan

**Fix 5: keep the data as a judge.** Every fix so far refines a parametric model, and the last risk is to end up looking at nothing else. Hence the rule: keep the data models in view at all times (historical windows, a bootstrap of your funds, the broad sample of the century), because none of them depends on a number you typed ([[historical-vs-parametric]]). They play the part of reality in the room. When the corrected central model sits far from the historical windows, the gap makes your window bias visible. When it sits far from the broad sample, your bet that "the future will be gentler than the century" becomes explicit. Disagreement between models is not a defect to resolve, it is the finished product ([[failure-probability]]). If your tool shows one model at a time, rebuild that argument by hand: run the same plan through a historical replay (cFIREsim, FICalc) and through a bootstrap (Portfolio Visualizer).

**Fix 6: simulate the real plan.** The last fix is not about the market, it is about you. A perfect engine pointed at a cartoon plan (a rigid withdrawal forever, no pension, no taxes, no buffer) produces numbers about nothing. The plan's realism counts as much as the market's: your pension on the date it starts, the side income of the early years, the taxes that gross up every sale under the account rules where you live ([[us-taxes-in-the-withdrawal-phase]]), the buffer with its rules for spending and refilling ([[refilling-the-buffer]]), and above all your spending rule (a flexible cut, guardrails with a floor, VPW, ABW, [[withdrawal-strategies-overview]]). Each of these moves the failure probability by several points, which is more than most modeling debates are worth. That is the distinctive strength of simulation ([[monte-carlo-strengths-and-limits]]): this kind of complexity is free, and leaving it unused is a waste.

::: science What the literature says about each fix
Every fix has an academic pedigree. Shrinking return estimates goes back to James-Stein (1961) and to Jorion's finance version (1986, "Bayes-Stein estimation"). Student-t tails go back to Praetz (1972) and Blattberg-Gonedes (1974). Markov regimes go back to Hamilton (1989), now the standard tool for persistent bear markets. Anchoring to valuations goes back to Campbell-Shiller (1988), with the withdrawal-rate version from Kitces (2008) and ERN (Part 54). The block bootstrap goes back to Politis-Romano (1994), brought to the retiree's problem by Anarkulova-Cederburg (2023). Nothing on that list is exotic. It is the ordinary toolbox of financial econometrics, applied with some consistency to a problem most consumer tools still handle with the means of 1998.
:::

## All of it together: relevance, step by step, with numbers

Take one plan through the fixes and watch what each one costs or reveals. The plan: EUR 1.4M, EUR 50,000 a year, 45 years, a global portfolio with 18 years of kind history (raw fitted μ of 6.5% real, σ of 11%, monthly kurtosis of 8).

| Step | Model | Failure | What it adds |
|---|---|---|---|
| 0 | Gaussian i.i.d., raw historical μ, rigid plan with no pension | ~1% | The brochure number |
| 1 | + blending toward the world prior (μ 5.5%, df 4, at the 50/50 cap) | ~4% | The kind window stops making the law |
| 2 | + Student-t (fitted df ≈ 5) | ~5.5% | Disasters get their frequency back |
| 3 | + the CAPE anchor (expensive market: equity μ pulled toward ~3%) | ~9% | The present enters the model |
| 4 | Sequence stress reading (mean-preserving, sticky bears) | ~12% | What the sequence costs: +3 points |
| 5 | Broad-sample reading | ~13% | The century confirms the zone |
| 6 | + the real plan: a EUR 14k pension in year 16, written flexibility of −10% | central ~3.5%, stress ~6% | The plan's realism gives back what the model's rigor took away |

The path from 1% to 9% to 3.5% tells the whole philosophy. The naive run flattered (1%), rigor sobered things up (9% to 13%), and the realism of a complete plan handed back honest margin (3.5% to 6%). The final number looks like the naive one and has nothing in common with it. It was won against the traps rather than handed over by them, and you know exactly which assumptions carry it and which margins defend it.

## The essentials

- A Monte Carlo becomes relevant through six fixes in order: blending the parameters toward a world prior (the antidote to the bias of your own window, dosed by how far the horizon outruns your history), fitted Student-t tails, mean-preserving regimes for memory (the stress measures what the sequence costs and nothing else), anchors to the present (CAPE), the raw data kept as a judge, and simulating the real plan.
- The finished product is not a number but a body of evidence: an honest central model bracketed by bounds. You decide inside it, on the harsh models.
- Calibrations do not stack: blending or a manual μ or the CAPE anchor as the central case, the others as cross-checks. Prudence piled three layers deep is paid in working years.
- The realism of the plan (pension, income, taxes, spending rules) often weighs more than any market refinement. That complexity is free in a simulation, so use it.
- Every fix has thirty years of literature behind it. The hard part is not finding a good one, it is applying them **all**, by default, instead of reserving half of them for whoever knows to ask.

---

## Going further

- Jorion, "Bayes-Stein Estimation for Portfolio Analysis" (1986): shrinkage applied to returns.
- Hamilton, "A New Approach to the Economic Analysis of Nonstationary Time Series" (1989): Markov regimes.
- Early Retirement Now, Part 54 (the CAPE rules) and Part 15 (measuring sequence risk) ([[the-ern-series]]).
- In this book: [[under-the-hood]] and [[using-the-fire-simulator]], for these six fixes as the simulator implements them.
- Next in this book: [[market-regimes]] (the economics behind sticky bears and the four macro seasons).
