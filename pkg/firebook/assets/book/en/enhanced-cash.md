# Enhanced cash: money market funds, T-bills and AAA CLOs
<!-- source: cash-ameliore @ df55b396faae -->

The chapter on the buffer ([[cash-buffer]]) settles the question of size: how many years of spending to hold in cash, and what that costs in return. It does not say **what to fill that sleeve with**, and that is a decision in its own right. A retiree who keeps three years of spending on the side ties up something like $60,000 to $120,000, for thirty years. One extra point of yield on that money is a trip a year, forever.

The trouble is that every extra notch of yield is bought with risk, and that risk has a bad habit of waking up on exactly the day the short sleeve is needed. This chapter ranks the instruments from the most inert to the best paid, measures the real gap between them, and takes apart the rung fewest retirees know, the **AAA CLO**, which has been available in an ETF since 2020.

::: cle The three questions a short sleeve has to answer
A cash holding is only good if the answer is yes three times.

- Can I sell it on the worst day of the decade?
- Will I get roughly what it is worth that day?
- Is what it pays above a money market fund, after fees and after tax, worth the answers to the first two?

The third question kills more products than the other two combined.
:::

## The rungs of the ladder

**The savings account** is the only instrument here with no market risk at all. Principal is intact by construction, the money moves the same day, and deposits are federally insured to $250,000 per depositor, per insured bank, per ownership category. The price is a rate the bank sets rather than the market, and banks are quick to follow the Fed down and slow to follow it up. This is the ground floor, the first six months of spending, and nothing else does that job as well.

**The Treasury money market fund** is the yardstick everything else gets measured against. It holds T-bills and overnight repo, it tracks the short rate within days rather than months, it costs a tenth to a third of a point a year, and its worst day is measured in basis points. Buying T-bills outright, or a 0 to 1 year Treasury ETF, does the same job with the same near-zero volatility. One quiet bonus: interest on Treasuries escapes state and local income tax, which is worth real money to a retiree in a high-tax state and nothing at all to one in a state with no income tax.

**The stable value fund** exists in one place only, inside a workplace plan that offers it. It holds a short bond portfolio, wraps it in insurance contracts, and credits participants a smoothed rate at book value, so the balance never falls. Two rare qualities for a retiree: a value that only goes up, and a yield above what a money market fund pays in most years. Two matching defects: the credited rate follows market rates with a lag of a year or two, so it is generous on the way down and stingy on the way up; and the money is not really yours to move. Most plans carry an equity wash rule, which forces a transfer out of stable value to sit in a non-competing fund, usually for 90 days, before it can reach a money market or short bond option. Roll the plan over to an IRA and the fund does not come with you.

**The ultra-short bond fund** adds short corporate credit to the mix. The premium is real and small, on the order of a quarter of a point a year, for a worst drawdown of about a fifth of a point. An honest rung, with no surprises in it.

**The AAA CLO** sits at the top, and pays the most: roughly 130 basis points over the short rate, against a price risk in bad years. It is also the least understood of the lot, and the most disliked, mostly because of its acronym, which is reason enough to take it apart piece by piece. That is the next section.

**Defined-maturity bond funds** form a rung of their own, neither quite cash nor quite an ordinary bond fund. They have their own section further down.

## The CLO, taken apart

The acronym frightens people, since it rhymes with CDO, the mortgage securitizations at the heart of the 2008 crisis, and that fear is the main obstacle. It is cured by taking the thing apart, because the thing is simpler than it sounds.

**The fuel: corporate loans.** A CLO, a collateralized loan obligation, is a vehicle that buys a portfolio of 150 to 250 loans, the ones the market calls leveraged loans. They fund indebted companies rated below investment grade, often after a buyout. Three features define them. They float, at the short rate plus a margin of roughly 350 to 400 basis points. They are secured by the borrower's assets. And they are senior: in a bankruptcy they are paid ahead of every bond the company has issued, which is why they have historically recovered 60 to 70 cents on the dollar against about 40 for an ordinary bond.

**The waterfall.** To fund that portfolio, the vehicle issues tranches ranked in order, and the cash from the loans fills them the way a waterfall fills one basin before the next. Loan interest pays the AAA tranche first, then the AA, and on down to the residual, the **equity**, which takes whatever is left. Losses travel the other way: every default eats the equity first, then the lower tranches, one at a time. The AAA tranche is therefore protected by everything sitting underneath it, a cushion of roughly 38% of the portfolio. Here are the orders of magnitude of a typical structure, with the spreads quoted in 2026:

| Tranche | Share of the structure | Cushion underneath | Spread over SOFR |
|---|---|---|---|
| AAA | about 60% | about 38% | about 130 bp |
| AA | about 11% | about 27% | about 200 bp |
| A | about 6% | about 21% | about 250 bp |
| BBB | about 7% | about 15% | about 335 bp |
| BB | about 5% | about 10% | 550 to 850 bp |
| Equity | about 9% | none | the residual, nothing promised |

(bp, basis points: hundredths of a percentage point.)

**The arithmetic of the cushion.** The loss on a loan is the default times the share not recovered. At a 65% recovery, every default costs 35 cents of what was lent, so eating through a 38% cushion would take defaults on very nearly the whole portfolio over the life of the vehicle. The worst year in the history of leveraged loans, 2009, saw about 10% of loans default with recoveries below normal, for a portfolio loss on the order of 4%. It would take a decade of 2009s in a row to get near the cushion, and the internal guardrails would shut the taps long before that.

**The internal guardrails.** The structure polices itself through contractual tests checked every quarter. The overcollateralization test verifies that the value of the loans covers each tranche with the agreed margin; the interest coverage test does the same for the cash flows. The moment a test breaks, the waterfall closes: payments to the lower tranches and to the equity are suspended, and the money pays the AAA down early instead. In other words, the structure degrades in favor of the senior holder. On top of that come diversification rules, roughly 2% at most per borrower and caps by industry, which stop the manager from concentrating the portfolio.

**The life of the vehicle.** A CLO is born with a calendar: about two years during which it cannot be called, then four or five years of reinvestment, in which the manager recycles repayments into new loans, then amortization, where every dollar repaid runs down the waterfall and retires the AAA first. The ETF erases that calendar for you. It sells the tranches that are amortizing, subscribes to new issues, and delivers the same profile continuously. That is a convenience, and a reminder: you do not own a security that matures, you own a permanent exposure to a spread.

**And 2008?** The pre-crisis CLOs came through without a single AAA tranche losing a cent of principal, while their CDO cousins, backed by mortgage credit that all moved together, collapsed. The difference was the collateral and the waterfall. Today's structures, redrawn after 2010, carry a thicker cushion on top of that: subordination under the AAA went from about 25% to close to 40%. The tally is the argument worth keeping: across roughly 7,000 AAA tranches rated between 1993 and 2022, not one default, ever.

**So where does the spread come from, if defaults are nowhere to be found?** From three things that are not credit risk. Complexity: every CLO comes with several hundred pages of documentation that somebody has to be able to read. Liquidity: tranches trade over the counter, by bid lists, not continuously on an exchange. And a narrow buyer base: money market funds are not allowed to own them, and neither are many institutional mandates. It is a premium for a barrier to entry more than a premium for danger. The ETF is exactly what lowers the barrier, and if enough savings pour in, the premium will compress, the way every premium of this kind eventually does.

**What it pays, measured and not promised.** The US-listed funds have the longer record. Since October 2020, the category has returned about 4.4% a year in dollars against 3.0% for three-month T-bills, a spread of 133 basis points, with a worst drawdown of 2.60% in the summer of 2022, when spreads blew out. The best known of them is JAAA, and the category charges around 0.20% a year. A younger cohort of euro-listed funds, launched between September 2024 and mid-2025, lands within a few basis points of the same spread, each measured on its own short window. That uniformity is the real lesson. On a spread asset, as long as nothing breaks, the difference between managers does not show up, and the only durable difference between two vehicles is what they charge. Ranges, fees and rules all move: everything named here was checked in mid-2026, and it is the mechanism, not the level, that is meant to last.

::: figure echelle-du-cash
The four rungs, measured and not promised. On the right, the annualized yield above the money market fund. On the left, the worst drawdown taken over the same window. The two bars are not the same kind of thing, one is an annual flow and the other a one-off accident, and that is exactly the tradeoff to settle. The two middle rows were measured on euro-listed funds and read as category orders of magnitude; the bottom row is the US-listed fund, and it is the one with a full cycle behind it.
:::

**What can go wrong.** Not rates: the coupon floats and duration is close to zero, so 2022, the year that wrecked fixed-rate bonds, was a non-event for these tranches. Not defaults either, as we just saw. The risk is the market price, and history gives the measure of it, each time with the delay before things came back.

- **2009.** First-generation AAA tranches traded at 70 to 80% of **par**, the amount they repay at maturity. Every one of them was repaid at 100, but the quotes took two to three years to climb back, the time it took the credit market to heal.
- **March 2020.** Spreads went from about 130 to more than 500 basis points in a few weeks, and prices fell from 100 to 85 or 90. Par came back in the fourth quarter of 2020, nearly nine months later, once realized defaults (3.2%) turned out to be far below the 8 to 12% the market had feared.
- **Fall 2022.** British pension funds, caught out by their own rate hedges, dumped CLOs in a hurry to raise cash, and euro tranches gave way. On the dollar fund the trough was −2.6%; it was erased in five months, and the previous high recovered a little under a year after it was left behind.

Three episodes, one lesson: this holding sells at the price of the day, not at what it repays, and the day you need to sell is not necessarily a good day. Hence the rule of use further down, which keeps the next twelve months of spending out of it.

## Below the AAA: mezzanine and equity

The table above is the price list for the lower floors: a few dozen basis points more at each step down to BBB, then a jump to somewhere between 550 and 850 at the BB. Every step down removes cushion and adds yield.

The thing not to miss is that **the AAA record does not travel downward**. The perfect default history belongs to the top of the structure, precisely because the tranches below exist to absorb the losses in its place. The CLOs issued before 2008, the generation the market calls 1.0, did put their lower tranches through real pain, and a CLO BB behaves, in a bout of credit stress, like high yield with leverage built in. This is not enhanced cash. It is a risk asset, to be compared with stocks and not with a money market fund ([[false-defensive-assets]]).

The US market will not stop you here, which is worth knowing. Down the stack, mezzanine and even equity exposure are available in listed funds, so the guardrail has to be your own discipline rather than the shelf. There are also listed closed-end vehicles specializing in CLO equity, paying large quarterly distributions and often trading away from their net asset value. That is an equity investment in structured credit, with the volatility that goes with it, and it is no longer the subject of this chapter.

One regulatory point deserves a line, because it shapes what you are buying. In Europe, a risk retention rule forces the issuer to keep a slice of what it sells. In the US that rule was struck down for open-market CLO managers by the D.C. Circuit in 2018, on the grounds that a manager who buys loans in the market and never holds them is not a securitizer. So a US CLO manager has no skin in the game by law. What aligns it with you is the structure's own quarterly tests, its reputation with repeat buyers, and nothing else. Read that as one more reason to stay at the top of the waterfall, where the tests work in your favor.

## Defined-maturity funds: a basket of bonds with an end date

A defined-maturity fund, also called a target-maturity fund, holds a basket of bonds that all come due in the same year. On the appointed date it pays holders out and dissolves. That construction detail changes everything: it gives the fund a property ordinary bond funds do not have, an end date. An ordinary bond fund replaces its holdings as they mature, its duration stays put, and its value rises and falls with rates forever. A dated fund watches its duration melt year after year down to zero, exactly like the bond it imitates.

**What you buy on the day you buy it.** The yield to maturity of the basket is known and published at purchase. Hold to the end and that is your expected return, subject to two reservations that are not details, fees and defaults. Between the two dates, the price moves with rates and spreads like any bond fund, and that intermediate volatility does not concern you if, and only if, you hold to the end.

**What exists, and how far out.** Two ranges share the US market, the **iBonds** from iShares and the **BulletShares** from Invesco, and between them they cover more ground than most people realize: Treasuries, TIPS, investment grade corporates, high yield and municipals, one fund per maturity year. The TIPS vintages are the underrated ones, since they let a household assemble a real ladder of inflation-linked bonds in a handful of orders ([[inflation-linked-bonds]]). Fees run from about 0.07% on the Treasury vintages and 0.10% on corporates and TIPS to 0.18 to 0.40% on munis and 0.35 to 0.43% on high yield, so it is the credit segment that sets the price, not the brand.

The comparison that matters, then, is not one range against the other. It is the fund against the ladder you would build by hand.

| | Defined-maturity ETF | The ladder you build yourself |
|---|---|---|
| Contents | several hundred bonds of one segment and one year | the five or six issues you picked |
| Annual fee | 0.07 to 0.43% depending on the segment | none, but you pay a spread on anything that is not a Treasury |
| Buying | one order per vintage, any trading day | one order per issue, at whatever the desk quotes |
| Selling early | at the market price, any day | at the market price, wide on a small corporate lot |
| Diversification | several hundred issuers | five or six, so one default leaves a real hole |

**Why you might still build it yourself.** On Treasuries, the do-it-yourself ladder wins on the merits, not on principle. A Treasury bought to mature in June 2031 pays par in June 2031, to the day and to the dollar, at no fee at all; the fund's final distribution is close to that but is not contractually par, and its last months drift away from the yield it advertised (more on this below). For a rung that has to fund an exact amount on an exact date, that precision is the whole point. Everywhere else the fund wins and it is not close, since no household diversifies corporate or high yield credit with five holdings.

Then comes the question the brochures skip, which in the US is a placement question. The income from any of this is interest, taxed as ordinary income in the year it is paid, whether you spend it or not: there is no accumulating share class to defer it into. Treasury and TIPS interest is exempt from state and local tax, which makes the Treasury rungs a natural fit for a taxable account. Corporate and high yield vintages, whose whole return is fully taxable coupon, belong in a tax-advantaged account when you have the room. Municipal vintages are the answer at high federal brackets in a taxable account, and only there ([[us-taxes-in-the-withdrawal-phase]]).

**And how far out?** Both ranges publish one vintage per calendar year, from the current year to about ten years out: in mid-2026, the furthest ones mature in 2035. That is the limit to know before building anything on top of it. **There is no fifteen or twenty year vintage.** A dated fund can finance an expense in the coming decade, not the floor of a thirty-year retirement, which sends you back to individual bonds and a ladder built by hand ([[bond-ladders]]).

**What duration does to the yield.** Nothing that belongs to the product: it all comes from the yield curve. A vintage yields what bonds of its maturity yield, and the slope of the curve alone decides what a distant vintage pays against a near one. On July 31, 2026, the Treasury curve gave 4.08% at one year, 4.45% at five and 4.75% at ten, so stretching four years further out bought a little under four tenths of a point. And the sign flips, which is the real trap. On July 31, 2023, the same curve paid 5.37% at one year against 3.97% at ten, so the shortest vintage in the range was the best paid of the lot, by a mile.

The rule that follows is short. Pick the vintage on the **date of the expense**, never on the yield on the label. Look at the curve second, to know what the matching costs or pays that particular year. A higher yield at the far end of the range is not a free premium, it is the price of four more years of uncertainty.

**How to use it.** The clean use is matching, not reaching for yield. You know a dated expense: the bridge to a pension, a project, the share of the floor to be funded in a given year ([[bond-ladders]]). You buy the matching vintage, and you stop wondering where rates will be that day. This is a **floor** block, not a buffer block.

For a bridge of several years, you stack one vintage per year to be funded, sized on that year's spending less the income already coming in. It is a bond ladder bought in four trades instead of built one security at a time, with diversification thrown in, several hundred issuers against the five or six holdings a household would own directly.

Three execution details the brochures leave out. **The final year does not pay what the vintage advertised**: as the bonds mature, the fund banks the proceeds and parks them in cash instruments, so its yield drifts toward the short rate over the last months. On a vintage bought four years out, that shaves a fraction off the advertised yield, and the fraction is bigger the lower the short rate. **The maturity is a liquidation**: on the appointed date the fund is delisted and pays you out in cash, like a bond being redeemed. Nothing rolls over, and in a taxable account that makes a single tax event, worth anticipating if the gain is large. **The income arrives monthly**, whether the household needs it or not, which suits a regular spending need and works against a lump-sum one, where the coupons have to be reinvested at whatever rates prevail.

That leaves checking what you are actually buying. The size of the vintage, because the far-dated ones are sometimes thin and you pay for that in the order book. And the index behind it, because some ranges track filtered indexes that narrow the universe of issuers without the fund's name saying so.

**The four traps, in order.** Selling before maturity, first, which resurrects the very rate risk the product was meant to extinguish. Credit, second: a high yield vintage advertising 6% will lose issuers along the way, and the yield on the label is a ceiling, not a promise. Fees, which weigh heavily in proportion over three or four years. And reinvestment, last: at maturity you get cash back and the question opens again at whatever rates then prevail, which the brochure rarely calls a risk.

::: attention The real comparison is not the headline yield
Two classic traps. Tax first, because the premium has to be compared after tax. All of this pays interest, taxed at ordinary income rates every year, so the yield that matters is the one left over, and it depends on your bracket, your state and which account holds the money. Then currency: a CLO ETF denominated in another currency and left unhedged turns a 130 basis point argument into a currency bet with 10% volatility. For a short sleeve, only the version in the currency you spend makes any sense.
:::

::: exemple Three years of spending, three ways to hold it
A household spends $30,000 a year and keeps three years on the side, $90,000. All of it in a Treasury money market fund earns the short rate, call it 4%, or $3,600 a year. Split in thirds, a savings account, a money market fund and an AAA CLO ETF, it earns about $400 a year more before tax, against a paper loss of $700 to $800 on the exposed third in a 2022-style episode. The tradeoff is defensible, and it stays modest. That is the most useful lesson in this chapter: the short sleeve is managed for safety, and its yield is a bonus, never an objective.
:::

## The assembly rule

The plan below is for a household in the withdrawal phase, living off its capital, whose short sleeve has one job: pay the bills without ever forcing the sale of a risk asset at the wrong moment ([[the-three-phases]]). A saver still working, whose paycheck covers the bills, has a much lighter liquidity constraint and can compress the first two layers. For a retiree, three layers are enough, and the order matters.

- **The first six months** of spending are untouchable and available on the spot. A savings account, or a Treasury money market fund. No market risk, no exceptions.
- **The next twelve to eighteen months** go into a Treasury money market fund or short T-bills, or a stable value fund if that is where the money already sits. This is the heart of the buffer, the layer that funds a crossing, the years spent inside a bear market, without selling risk assets ([[bear-markets-in-retirement]]).
- **The remainder, and only the remainder**, may climb a rung, into an ultra-short bond fund or an AAA CLO ETF. This slice is never the next twelve months of spending.

And a list of exclusions, which is worth as much as the rules. High yield bonds are not cash: they lose 20% when stocks lose 30%. A "short-term bond" fund with two or three years of duration is not cash either, as 2022 reminded everyone. "Principal protected" structured products charge more for the protection than it is worth ([[false-defensive-assets]]). And yield-bearing stablecoins, whatever the platform, have neither a guarantee nor a lender of last resort.

## The essentials

- The short sleeve is filled in rungs, from the savings account to the AAA CLO, and the test is not the yield on the label but whether you can sell at a decent price on the worst day of the decade.
- An AAA CLO is the senior tranche of a portfolio of secured corporate loans, protected by a cushion of about 38%: with recoveries of 60 to 70%, nearly every borrower would have to default before it is touched. Not one AAA tranche has defaulted out of roughly 7,000 rated since 1993, 2008 included.
- Measured and not promised, the US-listed AAA CLO funds have paid 133 basis points a year more than three-month T-bills since October 2020, net of a fee around 0.20%, with a worst drawdown of 2.60% in 2022. The younger euro-listed cohort sits within a few basis points of the same spread.
- The spread pays for complexity, illiquidity and a narrow buyer base, not for default risk. It is charged back in market price on the bad days: 70 to 80% of par in 2009, sharp breaks in March 2020 and in the fall of 2022.
- Below the AAA, the mezzanine adds yield by removing cushion, from about 335 basis points at BBB to 550 or 850 at BB, and the spotless record does not travel one floor down. Nothing under the AAA is cash, and the US shelf will sell it to you anyway.
- The defined-maturity fund is the neighboring block, with an end date and a yield to maturity known at purchase. It funds a dated expense, not the buffer, and it only keeps its promise if you hold it to the end.
- The vintages stop about ten years out: this block covers a decade, not a retirement. What duration adds to the yield is only the slope of the curve, a little under four tenths of a point in mid-2026 and deeply negative in 2023. Pick a vintage on the date of the expense, not on its yield.
- Use the fund for credit, where five holdings cannot diversify anything, and consider building the Treasury rungs yourself, where par on the day beats a final distribution. Three execution details to know: the last year's yield drifting toward the short rate, the cash liquidation that makes one tax event, and the monthly income that lands in the account whether you want it or not.
- Typical assembly: six months in a savings account, twelve to eighteen months in a money market fund or T-bills, and only the remainder a rung higher. Permanent exclusions: high yield, "short-term" bond funds with two years of duration, principal-protected structured products and yield-bearing stablecoins.

---

## Going further

- LSTA and S&P Global Ratings: the tranche-level default studies of the CLO market, the source of the record quoted here.
- J.P. Morgan, CLOIE indices: the reference series for the dollar and euro CLO markets, by rating tranche.
- Stable Value Investment Association: how a crediting rate, a wrap contract and an equity wash rule actually work, from the industry that runs them.
- In this book: [[cash-buffer]] (how much to hold), [[bond-ladders]] (the other use of the short end, matching), [[insurance-premia]] (the rung after this one, when the short sleeve turns into a premium sleeve), [[bonds-in-retirement]] (why duration is a separate decision).
