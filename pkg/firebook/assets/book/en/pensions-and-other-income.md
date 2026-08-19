# Pensions and other income in the plan
<!-- source: revenus-complementaires @ 5f57fdd075b4 -->

Almost no real plan is a portfolio alone against the world. There is Social Security, and a pension if you have one, arriving one day. Sometimes rent ([[real-estate-in-retirement]]). Work in small doses ([[going-back-to-work]]). An annuity, and maybe a distant inheritance. How you **count** those flows matters more to the plan than most portfolio choices do. Sensitivity work ranks the pension second among the variables of a plan, right behind spending.

This chapter is the accounting manual for that income. It opens with the taxonomy: guaranteed, near-certain, hoped-for, three treatments that people confuse at their own cost. Then the mechanisms, the ways a flow transforms a plan. A flow does not merely cut withdrawals. It shortens the horizon at risk, crushes the longevity tail, and counts as an implicit bond allocation. Next comes the modeling itself, inside a plan with real numbers in it: net real amount, start year, end year, pension assumptions, and the traps waiting in the entry. Then the awkward cases: rent, annuities, delicate inheritances, a spouse's income. And last the counter-check, because a plan can also lean on those promises **too** hard. When the promised flows dominate everything, their own risk (political, tenant, the health of the work) becomes the risk of the plan.

::: cle The taxonomy that decides everything
**Three** categories, three treatments.

The **guaranteed**: a public pension you can claim, annuities already paying. They go into the plan, marked down for their own risk, which means a 10% to 20% political haircut on the pension.

The **near-certain**: rent on a property you own, a spouse's pension. They go in too, with a 15% to 25% haircut.

The **hoped-for**: future work, side income, inheritances. Those do **not** go in. The plan has to hold without them. They are **margins**, named in the plan and cultivated, but never counted ([[going-back-to-work]]).

The discipline pays twice. The plan stays honest, and every good surprise strengthens it instead of rescuing it.
:::

## What a flow does to a plan: the four mechanisms

$15k a year of income in a plan that spends $50k does not "cut withdrawals by 30%". It does four things, and three of them run against intuition.

**1. It cuts net withdrawals, but only once it starts.** That much is obvious, and the timing is the subtle part. A pension that starts in year 17 cuts **nothing** during the bridge. That is why an early retirement runs in two regimes: an uncovered phase, the bridge years, then a covered phase ([[horizon-and-life-expectancy]]). And that is what sets up the next mechanism.

**2. It shortens the horizon at risk.** This is what a pension really buys you. The "50-year plan" becomes a 17-year bridge with the rest already covered. Sequence risk and the cliff of the fixed rule then bear on the bridge alone ([[sequence-of-returns]]). That is why adding the pension to the model often divides failure by two to four, and why leaving it out is the most expensive omission there is ([[ten-plan-wrecking-mistakes]]). The pension shows up **precisely** in the long scenarios, the ones where the portfolio is running out of breath.

**3. It crushes the longevity tail.** A **lifelong** flow (a pension, an annuity) covers the one risk capital covers badly: living a very long time ([[annuities-and-safety-first]], [[horizon-and-life-expectancy]]). Look at mortality-weighted failure. The scenarios that read "alive, broke, 95" become "alive, down to the pension floor, 95". That is a completely different event.

**4. It is an implicit bond allocation.** This is the TPAW reading ([[amortization-based-withdrawal]]). A pension of $20k a year, discounted, is worth $300k to $450k of inflation-linked government bonds inside your total wealth. At the same overall risk, the **visible** portfolio can therefore carry more stocks ([[stock-bond-allocation]], on covering the floor). The fifty-year-old with a pension coming who holds 40% bonds out of caution is usually under-invested. The caution was already in the pension.

## Modeling a flow: three numbers and their traps

Three numbers describe any outside income. An annual amount, a start year, an end year. Enter the amount **net** and in constant dollars, so after tax and assumed to follow inflation, like everything else in a withdrawal plan ([[inflation-and-withdrawal-rates]]). The arithmetic that follows is disarmingly simple. Each year the flow comes off the need, and the portfolio funds only the difference. Two shapes cover everything. The **lifelong** flow has no end year: that is a pension or an annuity. The **dated** flow has one: rent on a property you will sell, a stretch of work, unemployment benefits. Any serious simulator offers both, and running without them means running a plan that is not yours.

**Enter the pension three ways, never one.** In go the public pension with its haircut, any annuity already paying, and a weighted survivor benefit, each on the date it starts. The amount stays uncertain twenty years out, so good practice is to run the plan under three sets of assumptions. One stressed: the pension shaved by 20% to 30% and the age pushed back. One made only of what you have already earned, the floor you would get if your career stopped today. One taken from the official estimate on your own statement. A public pension is indexed and most private ones are not, so read each for what it actually promises ([[us-healthcare-and-social-security]]). Three entries, three failure probabilities, one range. What informs you is the spread between the three, not the middle number.

Three entry traps wait along the way. First, the **gross** figure copied straight off the statement, when what you need is the net one with its haircut. Second, a start **year** that flatters you: assume the later claiming age rather than the earliest one allowed, unless you have done the calculation. Third, the forgotten **second** pension in a couple. Add up the household's flows, each on its own date, weighting a gap of a few years into a single average start if that is simpler.

**Side income is modeled as a dated flow.** What goes in there: net rent with its haircut on a property you will sell, the **structural** work of a semi-FIRE, and unemployment benefits during the transition. Work counts only in that case, when it is an assumed parameter of the plan ([[going-back-to-work]]). There is one trap here and it is serious: slipping the **hoped-for** income of category 3 into the box. The simulator will then tell you what you want to hear. That is p-hacking with scenarios, exactly ([[simulator-traps]]).

**The start year matters as much as the amount.** A dollar of side income is not worth the same thing depending on when it lands. In the first five years it saves you from selling shares into a decline, so it attacks sequence risk right where sequence risk is decided ([[sequence-of-returns]]). In year 25 it is a comfort supplement, because the fate of the plan is largely settled by then. A small, early income therefore buys more safety than a much larger pile of extra capital. It is the best sequence insurance available, and this one mechanism explains why semi-FIRE and a gentle half-time work as well as they do ([[going-back-to-work]]).

And then read the output properly. Sensitivity gets **tested**: move the pension by plus or minus 20% and by plus or minus two years. If failure swings, the plan is a bet on the pension (see the counter-check below). A solver that prices equivalent moves puts these trades side by side ([[one-more-year]]). You get to see "$500 a month of side income for five years" sitting next to "$80k of capital", a menu that puts the magnitudes back in proportion.

::: science The awkward cases: rent, inheritances, a spouse
**Rent** is near-certain, but it is neither lifelong nor free of work. The rule comes from [[real-estate-in-retirement]]: a realistic net-net, marked down 15% to 25%. Model it as a dated flow to the year you plan to sell, or as a lifelong flow if you will hold for life, and list the property's value as a reserve. Name its own risks while you are there: vacancy, regulation, managing a building at 75.

**Inheritances** are the uncomfortable subject. Statistically likely for plenty of FIRE retirees in their forties, they are **never** counted. The amount, the date and where the money ends up are all three uncertain, if only because parents may need long-term care. And a plan that only balances thanks to a death is unsound by construction. Category 3, a named margin, end of story. A plan that **needs** the inheritance is not a plan.

**A spouse who still works** is a different case ([[couples-and-family]]). Their salary is a dated flow, running to their own exit date, with their willingness as its own risk, to be marked down if the gap is imposed rather than chosen. Their future pension, on the other hand, joins the household's lifelong flow.

**Transition benefits** (unemployment after a negotiated exit) are a dated flow, short and certain. Those 18 to 24 months cushion exactly the start of the fragile window. Count them without a second thought.
:::

## The counter-check: the over-backed plan

Everything in this chapter pushes you to lean on these flows. Clear thinking sets the bound on the other side. A plan whose floor rests 70% to 80% on promised flows has **concentrated** its risk on those promises. A public pension carries political risk: that is what the 10% to 20% haircut is for, and a plan that breaks when the haircut goes to 30% deserves a hard look. Rent carries regulatory and management risk. Structural work carries the risk of health and of appetite: the 48-year-old Barista may not still be the Barista at 61.

The robustness test applies at design time and at every review ([[the-annual-review]]). Run the plan **without** each flow, one at a time, that flow set to zero, two minutes per test. The failure probability that comes out does not have to stay under your threshold, since covering the plan is the whole point of the flows. But it has to stay in **recoverable** territory, the zone the playbook has steps for ([[when-to-worry]]). If losing a flow makes the plan unrecoverable, that flow is not a margin of the plan. It is the plan. It then deserves the treatment of a core asset: diversification, upkeep, a written plan B.

::: exemple The flow accounting of a real plan
Take the book's running household, at design time.

The couple's public pension: the official estimates give $31k a year at 67, a 15% haircut, so **$26k of lifelong flow from year 19** (guaranteed, marked down). The rent on the one-bedroom they kept: $8.4k net-net, a 20% haircut, so **$6.7k of dated flow to year 12** (sale planned at 60, near-certain). Transition unemployment adds **$18k in years 1 and 2** (certain and short). Her likely contract work (about $10k a year), the plausible inheritance and the survivor benefit are **named** in the margins block, and counted nowhere.

The results speak. Central-case failure drops to 3.9%, against 10.8% with all these flows ignored, and half the gap comes from the pension alone. The counter-check holds. Without the rent, failure rises to 4.6% (recoverable). With the pension marked down 30% instead of 15%, it rises to 5.4% (recoverable). With **everything** at zero it returns to 10.8%, hard but not unrecoverable given the playbook steps. The plan is backed, not suspended. Twenty minutes of careful entry: the factor of 2.8 on failure was sitting there, not in a choice of ETF.
:::

## The essentials

- Three categories, three treatments: guaranteed (counted, haircut 10% to 20%), near-certain (counted, haircut 15% to 25%), hoped-for (**never** counted, named margins that get cultivated). Confusing the three is the structural error in sizing a plan.
- A flow does four things: it cuts withdrawals from its start date, shortens the horizon at risk (the plan becomes a bridge), crushes the longevity tail (if it is lifelong), and builds an implicit bond allocation (the visible portfolio can afford more).
- Three numbers are enough to enter a flow: a net real amount, a start year, an end year. The pension is a lifelong flow with a haircut, summed across the couple and tested under three sets of assumptions; side income is a dated flow, worth all the more the earlier it lands. And never the hoped-for, because a simulator that flatters you is a dead simulator.
- Inheritances are never counted; a spouse's salary is a dated flow with its own risk; rent keeps its haircut and its sale date.
- The counter-check closes the chapter: stripped of each flow in turn, the plan has to stay **recoverable**. A flow whose loss is unrecoverable is not a margin. It is the plan, and it gets managed as one.

---

## Going further

- Early Retirement Now, Part 4 (Social Security inside the SWR) and Part 32 ("You are a Pension Fund of One"): flow accounting, worked in detail ([[the-ern-series]]).
- In the simulator: the two flow entries, `Pension /yr (net real)` with `Pension starts in year` and `Side income /yr` with `Side income until year`, the pension assumptions behind them, and the equivalent-moves solver ([[using-the-fire-simulator]]).
- In this book: [[going-back-to-work]] (when work counts), [[real-estate-in-retirement]] (rent), [[amortization-based-withdrawal]] (discounted total wealth), [[horizon-and-life-expectancy]] (the bridge).
