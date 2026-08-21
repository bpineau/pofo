# The ten mistakes that wreck a FIRE plan
<!-- source: erreurs-classiques-fire @ a09fa8480c2c -->

FIRE plans rarely fail from plain bad luck. Hostile markets are already in the models, and a well-built plan absorbs them. They fail from construction errors, almost always the same ones, documented by twenty years of forums, post-mortem blog posts and research.

This page lists the ten deadliest, in the order you meet them while building a plan, each with its mechanism, its warning sign and its fix. Six of them can be measured, and the ranking by cost is not the order of the list. The chart below does that ranking. Read the page twice: once before you build your plan, and again while you are running it ([[the-annual-review]]).

::: figure cout-des-erreurs
The measured cost of the six mistakes that can be quantified, on one shared reference plan ($1M, a rigid 3.5%, 50 years, a pension from year 20, 60/40) and one shared empirical model (16 countries, 1870 to 2020, 200,000 draws). The top two alone add 36.8 points of failure probability, against 33.9 for the other four combined.
:::

::: cle The thread running through them
Nine of these ten mistakes share one root: taking a model's output for a promise, and optimism for a plan. The antidote never changes. Audit the inputs, run several models ([[simulator-traps]]), name the margins, and decide the adjustment rule in advance ([[withdrawal-strategies-overview]]).
:::

## 1. Underestimating what you spend

This is the number one mistake, by a wide margin, and a multiple of 25 to 33 amplifies it ([[how-much-you-need]]): miss $250 a month and the target comes out about $90,000 short. The usual omissions are health coverage at full price, the lumpy items (car, home repairs, dental work, the vet), the tax on withdrawals, and the spending of the life you are **aiming at** (travel, the hobbies that free time makes room for) rather than the life you have now.

**Warning sign**: your retirement budget comes out below what you spend today, "because we will be careful". **Fix**: 24 months of statements, lumpy items annualized, and a budget in three tiers (floor, comfort, dream).

## 2. Running the 4% rule over a 50-year plan

The 4% rule is calibrated on 30 years of US data, before fees and before taxes ([[the-4-percent-rule]]). Over a 45 to 55 year horizon, with a rigid rule, the historically safe rate sits closer to 3.25 to 3.5% ([[the-ern-series]]), and the world sample is harsher still ([[anarkulova-cederburg]]). Leaving at 42 on 25 times your spending, with no margins, means accepting a double-digit failure probability without knowing it.

**Fix**: size the plan at 3 to 3.5%, or offset the gap with margins you have actually put numbers on (a future pension, flexibility, part-time income). Feed them to the simulator instead of invoking them.

## 3. Ignoring the tax on withdrawals

A withdrawal is not net. Part of it goes to tax, and how much depends on which account you sell from, on how much of the sale is gain rather than your own capital, and on the rules where you live ([[us-taxes-in-the-withdrawal-phase]]). Health coverage can add a bill of its own once no employer is paying for it. Call the whole thing the friction: the share of a gross withdrawal that never reaches your bank account. It varies enormously, from close to nothing on a modest plan drawn from sheltered accounts to well into the double digits on a large one drawn from a taxable account.

**Fix**: gross the target spending up by an explicit friction assumption, the way the chart above uses 12%, then replace the assumption with your own figure once you have run it on your own accounts. Do not count the friction twice: a simulator that models taxes already grosses up each sale by the tax due. And set the accounts up years ahead, because the order you fill them in gets decided during accumulation, not on the day you quit.

## 4. Forgetting your own pensions

The "prudent" mistake that costs years: sizing the plan as if the state pension did not exist. Even a career cut short at 45 has already earned one, and it lands in your sixties, exactly in the scenarios where the portfolio is running out of breath ([[pensions-and-other-income]]). Counting it often divides the failure probability by two to four, and the lower the failure rate you are aiming at, the more it matters. On the reference plan of the opening chart, whose failure rate is already high to begin with, leaving the pension out adds 9 points.

**Fix**: pull your official statement (in the US, your Social Security estimate, [[us-healthcare-and-social-security]]), read it conservatively, then enter it in the plan as deferred income with the age it starts.

## 5. Confusing the average return with the return you actually get

A portfolio that "averages 7%" does not hand 7% to a retiree who is withdrawing: the order of the returns matters as much as their average ([[sequence-of-returns]]), and volatility drag eats into the compounded return ([[arithmetic-vs-geometric-returns]]). This is the central conceptual mistake of the whole subject. It is what makes people believe a good accumulation portfolio is automatically a good withdrawal portfolio.

**Fix**: think in sequences, not in averages. Test the plan against an early bear market, which is what a sequence stress is for, and look at the protections built for the fragile window ([[glidepaths]], [[cash-buffer]]).

## 6. The one-regime portfolio

All equities because "over the long run it always goes up", or the mirror image, 60% parked in cash and stable-value holdings because "that is safe". Both ignore that market regimes exist ([[market-regimes]]). All-equity swallows lost decades in real terms (Japan 1990, the world 2000 to 2009) at the worst possible moment. Over fifty years it does not fail more often than a 60/40, but it forces three more years of living on a cut budget, and those are years you actually live through. The all-bond portfolio, meanwhile, gets eaten alive by a decade of inflation (the 1970s, 2022). Measured on the reference plan, that is the most expensive mistake on this page. The portfolio that "lives off its dividends" is the same trap in disguise, and no safer: a dividend gets cut in a crisis, and it is a forced withdrawal from your own capital anyway ([[false-defensive-assets]]). A withdrawal portfolio has to survive all **four** growth and inflation regimes, not the most likely one ([[all-weather-portfolios]], [[defensive-assets]]).

**Fix**: diversify by regime (global equities, duration, gold and/or linkers, possibly some [[managed-futures|managed futures]]), and run an explicit test against persistent inflation and against a lost decade.

## 7. Betting everything on flexibility

"If things go badly, we will spend less." True, useful, and bounded. The research ([[flexibility-in-practice]], ERN parts 23 to 25) puts realistic flexibility, a 10 to 20% cut you can hold for a few years, at roughly 0.3 to 0.5 points of withdrawal rate, and no more, because bad sequences sometimes run past ten years. A 40% cut "on paper" does not exist if the real floor sits at 90% of the budget. Measured, this is one of the two most expensive mistakes on the page, because it gets used to justify a higher withdrawal rate, and that rate is entirely real.

**Fix**: put a number on the real floor, the one you can hold for **five** years, morale included, and write it into a rule decided in advance ([[floor-and-ceiling]], [[morningstar-guardrails]]) rather than into an intention.

## 8. Talking yourself into the answer you wanted

Every variant of the same bias: picking the model that gives the answer you want, usually the most optimistic one (a historical backtest of your own portfolio), raising the return assumption "because the S&P did 10%", rounding an 8% failure probability down to "about zero", rerunning until a good number turns up. The simulator stops being an instrument and becomes a machine for proving yourself right ([[monte-carlo-strengths-and-limits]], [[reading-a-fan-chart]]).

**Fix**: read several models side by side, starting with the ones you do not like. One model on its own is an opinion; the spread between several is the information ([[historical-vs-parametric]]). Plan somewhere between the central case and the broad sample, and treat the lost decade as a scenario to hold up under, not an unlikely one.

## 9. Neglecting the human factor

A perfect financial plan for a life nobody prepared: identity, the shape of the days, a partner who wants something else ([[couples-and-family]]), social isolation. In practice this is the most common failure of all, and the best documented by first-hand accounts ([[voices-from-real-retirees]], [[meaning-and-identity]]). It sends people back to the office within two years, not because the money ran out but because the days were empty. The mirror image is just as real: being psychologically unable to **spend** the capital, which keeps multimillionaires living like students ([[the-psychology-of-spending]]).

**Fix**: prototype the life you are aiming at before you leave (a sabbatical, part-time work), have a "toward what" and not only an "away from", and decide the spending rules in advance the way you decide the withdrawal rules.

## 10. Mistiming an irreversible exit

Two mirror-image mistakes. Leaving too early on a tight plan, at the top of an expensive market ([[valuations-and-cape]]), without checking that it survives an immediate 40% crash. And the opposite, which is sneakier: "one more year" repeated five times ([[one-more-year]]), because no number ever feels safe enough, while every year of healthy working life is the one asset in the plan that cannot be replaced.

**Fix**: write the exit criteria **in advance** (target hit at X%, failure probability under Y% in the pessimistic models, a floor you can hold, a life project prototyped), then stick to them in both directions.

## The anti-mistake checklist

Run through it before you leave, then once a year:

- [ ] Spending built on 24 months of statements, lumpy items annualized, health coverage included
- [ ] Tax friction accounted for (a stated gross-up assumption, or a calculation on your own accounts)
- [ ] Pensions estimated and given to the simulator as deferred income
- [ ] Withdrawal rate consistent with the horizon (3 to 3.5% rigid beyond a 40-year horizon, or explicit margins)
- [ ] Portfolio tested against the four regimes, not only the central scenario
- [ ] Spending floor quantified and realistic; adjustment rule written down
- [ ] Failure probability acceptable even in the broad-sample model; lost decade survivable
- [ ] Cash buffer sized, with the rules for drawing on it and refilling it decided ([[cash-buffer]], [[refilling-the-buffer]])
- [ ] Life project prototyped; partner agreed on the budget and on the life
- [ ] Exit criteria written down, with a review date

::: terrain What the post-mortems say
The failure stories posted on FIRE forums share one striking constant: almost none of them says "the markets ruined me". They say "I had underestimated my spending by 30%", "the divorce cut the capital in half" ([[couples-and-family]]), "I got bored and burned cash on projects", "I never saw the tax coming". The markets were in the models. It was the inputs and the life that were not. That is good news: mistakes like those can all be fixed before you leave.
:::

---

## Going further

- Early Retirement Now, part 26 ("Ten things the 'Makers' of the 4% Rule don't want you to know") and part 50 ("Ten things the 'Makers' of the FIRE movement don't want you to know"): the movement's blind spots, from its best analyst ([[the-ern-series]]).
- Building the target, line by line: [[how-much-you-need]].
- The annual review that keeps the plan honest ([[the-annual-review]]), and when to actually worry ([[when-to-worry]]).
