# Inflation-linked bonds: the only contract written in real terms
<!-- source: obligations-indexees @ b8d1d77f8012 -->

This whole book thinks in constant money. A retiree's spending is real, the floor is real, and ruin happens in real terms ([[inflation-and-withdrawal-rates]]). Yet nearly everything a retiree owns promises nominal. A conventional bond paying 3% while inflation runs at 5% makes its owner poorer on a schedule, by contract.

One asset class writes its contract in the same currency as your spending: inflation-linked bonds, TIPS in the United States, linkers in the trade, whose principal and coupons track the price index. For a retiree it is, conceptually, the risk-free asset: the only one that promises purchasing power. And almost nobody owns it. It even disappointed everyone in 2022, the very year inflation came back.

Both paradoxes have precise explanations, and this chapter walks through them. First the mechanics, and the breakeven, the tool you decide with. Then why 2022 was an instruction manual rather than a refutation, since the protection pays at maturity and not day by day. Then the most striking result in recent retirement finance, the ladder of linkers that guarantees in real terms what the 4% rule only hopes for. And finally the practice, the vehicles, the taxes, the trap of taxed indexation, and the dose.

::: cle The contract, and why nothing else looks like it
An inflation-linked bond promises this: your principal × (the price index at maturity / the index the day you bought), plus coupons computed on that inflated principal. Its quoted yield is therefore a real yield. "10-year TIPS at 1.2% real" means inflation plus 1.2% a year, backed by the US Treasury, whatever inflation does. It is the only instrument in the world whose expectation is quoted in the currency of the plan, which is real. For engineering a retiree's floor, that changes the nature of the problem. We stop projecting and start matching.
:::

## The mechanics, and the tool you decide with: the breakeven

**How it works.** The principal is re-indexed continuously to a reference price index, CPI-U for TIPS, the domestic consumer price index for their equivalents elsewhere. The coupon rate is fixed and applies to that re-indexed principal, so both legs follow prices. At maturity most issues, TIPS included, repay at least the original principal: a free deflation floor.

**The breakeven, the way to read them.** At any moment, put the nominal yield of a conventional bond of the same maturity next to the real yield of the linker. The difference is the inflation breakeven, the average inflation rate that would make the two investments equivalent. Nominal 10-year at 3.1%, real 10-year at 1.0%, and the breakeven is 2.1%. The decision rule follows. If you expect (or fear) inflation above the breakeven, the linker wins; below it, the nominal does. For a retiree the insurance reading is the truer one. At a breakeven near 2%, protection against any inflation above that costs no premium at all against the nominal, as long as inflation comes in at 2%. That is historically cheap insurance against the plan's first enemy ([[inflation-and-withdrawal-rates]]). The market sells it at the price of its central scenario, with no surcharge for the tail.

**So why does almost nobody hold them?** Three historical reasons. Real yields were negative from 2015 to 2021. The linker then guaranteed a real loss, and "inflation minus 1%" excited no one, even though it was already the best contract available in real terms. They are also nobody's default: no plan menu pushes them, and a quoted real yield reads like a typo to an investor raised on nominal ones. And then 2022, the great misunderstanding. Since 2023-24 real yields have been clearly positive again, around 2% on TIPS, and the asset class is back to being what it should be, the building block a retiree's floor is made of.

## 2022: the misunderstanding that teaches everything

Inflation at 8 to 10%, and long linker funds down 25%. The disappointment cooled savers for years. Take it apart, because the lesson is the instruction manual. Before maturity a linker runs on two engines. The first is realized inflation, credited to the principal, plus 8 to 10% in 2022: the protection paid, exactly as written. The second is the move in real rates, which hits the price, multiplied by duration. Real yields went from −2% to +1%, and on a duration of 12 that is −25 to −35% of price. The second engine crushed the first, and the long fund lost money with the indexation credited in full.

The lesson is not that linkers do not protect. It is that linkers protect at maturity, and that before maturity they carry real-rate risk exactly as nominals carry nominal-rate risk. Two correct uses follow, one per service ([[bonds-in-retirement]]). For the rebalanced portfolio sleeve, buy short linkers, duration 3 to 5, where indexation dominates and real-rate risk stays minor: that is what short-maturity TIPS funds are for. For matching the floor, hold to maturity, in a ladder, where the price along the way is beside the point because every rung is spent at its own maturity. 2022 punished only the third use, the incoherent one, long real duration held as though it were an indexed money market fund.

::: science The spectacular result: a ladder of linkers as a guaranteed pension
Here is the result that went around the community (Allan Roth in 2022, then Morningstar and the Bogleheads). With TIPS real yields near 2%, a 30-year TIPS ladder, one rung per year of spending, each rung held to maturity, funded an initial withdrawal of about 4.5% a year, indexed to inflation, backed by the US Treasury, for 30 years. That is more than the 4% rule, with no market risk and no sequence risk at all. It is the withdrawal rate people hoped to get out of a stock portfolio, reached by plain matching, for as long as real yields allow it. The number moves with those yields: at 1% real the same 30-year ladder funds about 3.9%.

But the conceptual point is enormous. The real yield in the linker market is the retiree's true risk-free rate, the yardstick every risky portfolio has to justify itself against ([[expected-returns]]).

When the ladder guarantees 3.9% and your risky plan targets 3.6% "with 8% failure", the question "why take the risk?" deserves a real answer. Good answers exist: a FIRE horizon runs well past 30 years, and the ladder has no growth, no bequest and no room to raise spending. But the answer has to be given, not dodged.

Then the limits. There is no reinvestment, since the ladder eats itself. Longevity past the last rung is hedged with an annuity ([[annuities-and-safety-first]]). And there is the real-yield window: at 0% real, the 30-year ladder funds only about 3.3%.
:::

::: figure linkers-echelle
The real annual withdrawal an indexed ladder funds, by the market real yield and the number of years covered, the ladder being fully consumed (nothing is left at the end). This is discounting arithmetic and not a market forecast; the only unknown is the real yield you buy at.
:::

## TIPS and I bonds in practice

**The vehicles.** TIPS are sold at auction in 5, 10 and 30-year terms, in $100 increments with a $100 minimum, either through a TreasuryDirect account or through a broker that bids for you. The secondary market at any broker quotes every issue still outstanding, which is where you go when you want one precise maturity rather than whichever one is being auctioned. Funds do the same job in one line: broad TIPS ETFs cover all maturities at a duration around 7, and short-maturity TIPS ETFs cover the front end, the right vehicle for the rebalanced sleeve. TIP and SCHP are the best-known broad ones, VTIP the best-known short one; pick on cost and on duration, not on last year's numbers ([[building-it-with-us-etfs]]).

**The ladder you can actually build.** This is where a US investor is unusually well served. Because the Treasury issues out to 30 years and the secondary market quotes every maturity still alive, an individual can buy the ladder of the section above rung by rung, at retail, for a few dollars of commission and no manager in the middle. What the previous section priced is executable, not theoretical. Two practical points. The maturity calendar has holes in it, a well-known stretch in the 2030s with no TIPS maturing at all, and the standard fix is to buy extra of the maturities on either side of the hole and spread the proceeds across the missing years. And a ladder is built once and left alone: the quoted price of a rung between today and its maturity is information you do not need.

**Series I savings bonds, the small line.** The I bond pays a composite rate made of a fixed rate, set when you buy and never changing for the life of the bond, plus an inflation rate reset every six months. It is bought only at TreasuryDirect, and only up to $10,000 per calendar year per Social Security number, which caps how much of a floor it can ever carry. You cannot cash it for twelve months; cash it before five years and you give up the last three months of interest; it earns for up to 30 years. The interest is exempt from state and local income tax, and federal tax on it can be deferred until you redeem, which makes the I bond the one linker that behaves well in a taxable account. Terms checked on treasurydirect.gov in August 2026: the limit and the penalties are set by the Treasury and do change, so read the site before you buy.

**The tax trap, and where it sends you.** In a taxable account the annual increase in TIPS principal is federally taxable in the year it happens, even though no cash reaches you until maturity. This is phantom income, and it is the tax on the protection eating into the protection. No fund wrapper makes it disappear: a TIPS fund distributes the accretion as taxable income every year. The fix is placement, not product. The TIPS sleeve belongs in a tax-deferred or Roth account, and taxable space is better spent on assets whose tax bill you control ([[us-accounts-and-account-order]]). One consolation for a resident of a high-tax state: TIPS interest, like all Treasury interest, is exempt from state and local income tax.

**Which index, for whose life?** TIPS follow CPI-U, the national urban consumer basket, and your own basket is not that one. Housing, health care and where you live pull your personal inflation a few tenths a year away from the official number ([[tracking-inflation]]). The cover is excellent and never perfect. One more reason not to aim for 100% linkers but to share the job with gold and the other real assets ([[defensive-assets]]).

## The dose, and the test

**The dosing doctrine**, sketched already ([[bonds-in-retirement]]), comes down to two numbers. Put 25 to 50% of the bond sleeve in linkers, the share rising with the length of the uncovered phase and with the rigidity of the floor. Across the whole portfolio, that is 5 to 15%. And where the floor dominates the plan, with little flexibility and no pension close, add the partial ladder: the first 5 to 10 years of floor matched in linkers held to maturity, the portfolio funding the rest. It is the modern, contractual version of the bucket ([[the-bucket-strategy]], [[bond-ladders]]), without the magic and without the market timing.

**Testing the block in simulation.** The useful A/B swaps a third of the nominals for short linkers, same withdrawal rule, same horizon. Read the answer where the block does its work, in the replay of the inflationary vintages, 1973 first of all, and in the models built out of blocks of history, which drag the 1970s back into the paths ([[historical-vs-parametric]]). The central case will barely move. That is normal: insurance does not pay in the median world. Then comes the yardstick, which teaches more. Put the annual spending your risky plan calls sustainable next to the spending today's ladder funds, the real yield plus the amortization of the capital. The gap between them is what the market pays you for carrying risk. When it is small, matching deserves a bigger share.

::: exemple A matched floor, in numbers
A couple, 52 years old, a floor of EUR 40,000 a year, comfort at EUR 55,000, pensions covering the floor at 66. The uncovered phase of the floor therefore runs 14 years. Matching it takes about 14 rungs of about EUR 40,000 real. At a real yield of 1%, that costs about EUR 520,000 in linkers held to maturity. The rest of the money, EUR 1.08M out of EUR 1.6M, funds the comfort above the floor (EUR 15,000 a year, a withdrawal rate of 1.4%) and everything that comes after, on a generous VPW with no anxiety ([[vpw]], [[choosing-your-strategy]]). The verdict is clean. Failure of the floor goes from "5% in simulation" to "zero by contract until the pensions start", and failure of the comfort layer is negligible by construction. The price paid is the expected return the EUR 520,000 gives up, since the linkers will not compound. This is the safety-first trade in its cleanest form ([[annuities-and-safety-first]]).
:::

## The essentials

- The linker is the only asset whose contract is written in real terms, "inflation plus the real yield on the label, guaranteed". Conceptually it is the retiree's risk-free asset, the yardstick every risky portfolio has to justify its premium against.
- You read it through the breakeven, real against nominal: holding it is a bet on inflation running above that number (~2%). Inflation insurance sells with no surcharge for the tail, which is what makes it historically cheap.
- 2022 is an instruction manual, not a refutation. Before maturity a linker carries real-rate risk times duration. Hence short durations in the portfolio sleeve, and hold-to-maturity for matching the floor.
- The result worth knowing: a ladder of linkers guarantees in real terms what the 4% rule only hopes for when real yields sit at 1 to 2% (about 4.5% at 2% real in 2023-24, about 3.9% at 1%). A US investor can build that ladder rung by rung, at auction or on the secondary market.
- Dose: 25 to 50% of the bond sleeve (5 to 15% of the total), plus the partial ladder of the floor for rigid plans. On tax, put the sleeve in a tax-advantaged account: the accretion is taxed as it accrues, and phantom income eats the protection in a taxable one.

---

## Going further

- Allan Roth, "The 4% Rule Just Became a Whole Lot Easier" (2022) and the Morningstar TIPS ladder analyses: the ladder result, in the original.
- Zvi Bodie, *Worry-Free Investing*: the theorist who first argued that the linker is the individual's risk-free asset.
- [treasurydirect.gov](https://www.treasurydirect.gov) for TIPS and Series I bonds, terms, auction calendar and daily index ratios: the primary source.
- [tipsladder.com](https://www.tipsladder.com) to see the ladder tool at work on live prices.
- In this book: [[bonds-in-retirement]] (the whole sleeve), [[bond-ladders]] (building to maturity), [[inflation-and-withdrawal-rates]] (the enemy being hedged), [[tracking-inflation]] (your index against the index).
