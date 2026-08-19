# Where returns come from: risk premia
<!-- source: primes-de-risque @ e2f2477955b9 -->

This whole book rests on an assumption so basic that nobody stops to look at it: a portfolio of stocks and bonds earns something, durably, above inflation. Without it there is no 4% rule, no FIRE, nothing. This page drags that assumption into the light and asks the three uncomfortable questions. Why do stocks pay more than bonds, and bonds more than cash? Who pays those gaps, and why would they keep paying? And why do some assets, gold first among them, return nothing in real terms without that being a flaw?

The answer fits in one term, the **risk premium**, and in one discipline: every holding in your portfolio has to name the premium it harvests and the reason that premium will survive its own fame. By the end of this page you can audit a portfolio holding by holding with that grid. It is the best known defense against products that promise a return without saying who funds it.

::: cle The idea in one sentence
An asset does not pay because it is good. It pays because it hurts at the worst possible moment, and somebody has to be paid to take that pain. Stocks earn 4 to 6 points more than cash precisely because they can lose half their value the month you lose your job. Return is the wage of risk carried, not a magic property of the asset. That is also why premia do not vanish once everybody knows about them. Knowing does not remove the pain.
:::

## The equity premium: the biggest, the best paid, the best documented

The gap between what stocks return and what cash returns, the **equity risk premium**, is the central empirical fact of finance. Over more than a century, and in every developed country, it comes out between 4 and 6 points a year in geometric mean ([[anarkulova-cederburg]] for the world sample, less flattering than the US case on its own). Compounded over thirty years, that gap leaves you with three to six times the ending capital of a money market holding. It is the engine of every FIRE plan.

Why does it exist? Theory gives a precise answer: what matters is not volatility as such, but **covariance with the bad states of the world**. Stocks collapse in recessions, when incomes fall, when unemployment climbs, when your neighbor is selling his house, exactly when one more dollar would be worth the most. An asset that betrays you at those moments has to offer a fat compensation to find a buyer. On top of that sits **disaster risk**: a national stock market can lose 80 to 100% (Germany 1948, Japan 1945 to 1949, Russia 1917), and part of the premium pays for tails that are rarely observed and never abolished ([[fat-tails]], [[international-diversification]]).

Remarkably, the premium we observe is too big for standard theory to explain: this is the famous equity premium puzzle (Mehra and Prescott, 1985), still argued over forty years later. For someone living off a portfolio, the puzzle is good epistemic news. A premium that theory cannot fully justify as payment for risk is unlikely to be arbitraged down to zero, because nobody can buy the disappearance of recessions. The best reasons to expect it to last are structural. Human risk aversion does not update. Institutions run on short horizons that keep them from carrying long-horizon risk. And the world's savings hunt for safety on a massive scale. Honest prudence projects the premium below its history, 3 to 4 points above cash at today's valuations ([[expected-returns]], [[valuations-and-cape]]), rather than at zero.

## The term premium, the credit premium, and the loose change

**The term premium** pays you for holding long bonds instead of cash: you take interest rate risk and, above all, inflation risk on fixed nominal payments ([[bonds-in-retirement]]). Historically it is worth 1 to 2 points, and it is the least stable premium in the catalog. It was sumptuous through the forty years of disinflation from 1981 to 2021. It went negative at the bottom in 2020, when bonds yielded less than expected cash and you paid for the privilege. It has been positive again since. The practical rule: you read it live in the slope of the yield curve, and a bond layer earns its place only while the premium is there ([[return-stacking]] applies that test to the letter).

**The credit premium** pays you for the default risk of corporate bonds. Its weakness is how little survives once realized defaults are deducted: 0.5 to 1 point net for investment grade, barely more for high yield, with a correlation to stocks that shoots up in a crisis (defaults arrive in recessions). That is why this book files high yield under false defensives ([[defensive-assets]]). It stacks the equity premium and the credit premium into the same bad state of the world, at a bond's price.

**The illiquidity premium** (private equity, private credit, unlisted real estate) pays you for not being able to sell. It is real in theory and often illusory in practice for a retail buyer: the fees of the vehicles you can actually reach eat it, and accounting smoothing of the valuations dresses it up as low volatility ([[real-estate-in-retirement]] for the property case). For someone living off a portfolio, whose job is precisely to sell on a schedule, illiquidity is also a first-order cost, not a detail.

**The alternative premia** (trend, carry, value) get their own chapters ([[global-macro]], [[managed-futures]], [[factors-in-retirement]]). What sets them apart is that they are partly **behavioral**: they pay less for a macro risk than for a persistent mistake made by other participants (underreaction, lottery seeking, institutional constraints). That makes them both valuable (uncorrelated with the bad states of the world) and more fragile (a mistake can be learned, a constraint can be lifted).

::: figure primes-echelle
The major premia, in points of annual real return above cash (historical ranges, before fees and taxes). Gold and cash offer nothing here, by construction.
:::

::: science Decay after publication: McLean and Pontiff
What becomes of a premium once it is published in an academic journal? McLean and Pontiff ("Does Academic Research Destroy Stock Return Predictability?", 2016) measured the answer across 97 anomalies. Returns fall 26% as soon as the study period ends, and 58% after publication, but they never reach zero. The right reading is a hierarchy of robustness. Premia backed by an uninsurable macro risk (equities, term) are the sturdiest, since nobody can arbitrage recessions away. Behavioral premia with documented limits to arbitrage (trend, value) compress but survive, decades after publication. Fine statistical anomalies with no risk story and no behavioral story die, and they were usually data mining ([[simulator-traps]]). The lesson for a plan: build the bulk of the return on the premia of the first tier, size the second tier in measured doses, ignore the third.
:::

## Why gold returns nothing (and why you want some anyway)

Gold is the perfect teaching counterexample. No cash flow, no dividend, no coupon, nobody handing off a risk in exchange for a wage: there is no gold risk premium, and its secular real return is close to zero ([[gold-in-retirement]]). That is not a hidden flaw, it is the definition. Gold is not a return asset. It is an **alternative currency** whose price rises when confidence in official money falls. You do not hold it for its premium, it has none, but for its **correlation**: it pays in the states of the world (currency crises, financial repression, stagflation) where almost everything else betrays you. The same reasoning, minus the thousand years of history, applies to cryptocurrencies: no cash flow, no theoretical premium, a pure monetary and behavioral bet, which rules out sizing them like return assets in a withdrawal plan.

Cash brings up the rear. It is the yardstick, since premia are measured above it, and it returns roughly zero in real terms over long stretches, with long negative episodes whenever financial repression joins in. Holding it has a role, room to maneuver ([[cash-buffer]]), but there is no return to expect from it.

## Auditing your portfolio, premium by premium

The grid is three questions per holding: which premium does this position harvest, who pays it, and why will they still be paying in twenty years? Here are the answers for the eight holdings you meet most often in a private portfolio. The levels came earlier; what matters here is the payer and the reason it lasts, the two things a prospectus never tells you.

| The holding | Which premium | Who pays it | Why it lasts |
|---|---|---|---|
| World equity ETF | Equity premium | The real economy, out of profits | Nobody can arbitrage recessions away |
| Long government bonds | Term premium | The borrower who wants long-dated debt | Holds while the curve has slope, the least stable of the lot |
| Inflation-linked bonds (linkers) | Term premium, without the inflation bet | The government, which keeps the inflation risk | An indexed contract does not expire |
| Gold | None | Nobody, there is no cash flow | Nothing to arbitrage, you are buying a correlation |
| Trend | Behavioral premium | The crowd that underreacts and the managers under constraint | Compressed by publication, not extinguished |
| Thematic fund | None beyond the stocks you already own | You, in fees | Nothing to last, the holding doubles a risk you already carry |
| Principal-protected structured note | Convexity sold, bank credit bought | You, through the issuer's margins | Two negative premia never add up to a positive one |
| Non-traded real estate fund or retail private equity | Illiquidity premium | In theory the hurried seller, in practice you | Real in theory, eaten by fees |

The last three rows are the real payoff of the exercise. None of them harvests a net premium for you, all three sell beautifully, and spotting them once and for all saves years of performance.

::: exemple The book's portfolio, audited
The model portfolio of this book (60% world stocks, 25% bonds including linkers, 7.5% gold, 7.5% trend) goes through the grid with no holes. Four holdings, four different answers, and one genuine return premium, the equity one, sized at an expected 3 to 4 real points above cash. The bonds buy the shock absorber of disinflationary recessions and the real contract the linkers carry ([[inflation-linked-bonds]]), the gold buys a correlation, the trend buys insurance with a positive expected value in long regimes. Every holding names its premium or its role, and no two of them hurt at the same moment. That is exactly what "diversified" means, and the next step follows in [[why-diversification-works]].
:::

## The essentials

- Return is the wage of a risk carried: the assets that hurt in the bad states of the world (stocks, credit) have to pay a premium to find a buyer. The ones that make nobody carry anything (cash) or that serve as money (gold) return nothing in real terms, by construction.
- The equity premium (4 to 6 historical points above cash, 3 to 4 prudent points looking forward) is the best documented and the most robust, because nobody can arbitrage recessions away. The term premium is read off the slope of the curve; the net credit premium is small and badly placed.
- Published premia compress (McLean and Pontiff, close to −60% after publication), but the ones backed by a macro risk or by a durable limit to arbitrage survive. Anomalies with no story die.
- Gold and cryptocurrencies carry no premium. You can hold the first for its correlation with hostile monetary regimes, but neither one gets sized like a return asset.
- The audit grid: for each holding, which premium, who pays it, why they will still be paying in twenty years. A holding with no answer is a bet or a commission in disguise, and a withdrawal portfolio has room for neither.

---

## Going further

- Antti Ilmanen, *Expected Returns* (2011) and *Investing Amid Low Expected Returns* (2022): the reference treatise on the premia, asset class by asset class.
- Mehra and Prescott, "The Equity Premium: A Puzzle" (1985); Dimson, Marsh and Staunton, the *Global Investment Returns Yearbook* (the premia over 120 years and 20 countries).
- McLean and Pontiff, "Does Academic Research Destroy Stock Return Predictability?" (2016): what becomes of a premium once it is published.
- In this book: [[expected-returns]] (putting numbers on forward-looking premia), [[why-diversification-works]] (assembling them), [[defensive-assets]] (the roles with no premium), [[anarkulova-cederburg]] (the equity premium beyond the US case).
