# Simulator traps: independence, US bias, survivorship
<!-- source: pieges-des-simulateurs @ 4de81676d4ea -->

Feed the same plan into five well-regarded retirement simulators: same capital, same spending, same horizon. You will get five success rates, and they can run anywhere from 99% down to 70%. None of these tools is lying. Each one simply made a dozen modeling choices, often without documenting them. Each of those choices moves the verdict by a few points.

This page takes inventory of those choices. Ten traps separate a flattering simulator from an honest one. For each one: the mechanism, the ballpark effect on failure probability, the audit question to put to any tool, and the answer an honest simulator gives (sometimes that answer is "the trap cannot be fixed, here is the margin to take"). Where [[monte-carlo-strengths-and-limits]] describes the machine, this page goes down into the plumbing. It doubles as a buyer's guide: ten questions, and you can audit the next simulator somebody recommends to you.

::: cle The traps, ranked
The traps do not weigh the same. Ranked by their typical effect on the failure probability of a FIRE plan: sample bias in the data (a factor of 2 to 5), error in the input parameters (a factor of 2), no fat tails (30 to 80% in relative terms), independent draws (20 to 50% in relative terms), omitted frictions (fees and taxes, worth 0.5 to 1.5 points of withdrawal rate), then the traps in how you read the output. A tool can be flawless on the last five and useless because of the first.
:::

## The data traps: what the machine learned on

**Trap 1: the US bias and survivorship bias.** Most historical simulators replay the United States since 1926, the winning country of the century, one of a kind ([[anarkulova-cederburg]]). The effect is not subtle. "Safe" rates come out 0.5 to 1.5 points too high against the full developed-market experience. The failure probability of a rigid 4% over a long horizon goes from about 2% on US data to 15 to 17% on a world sample. The audit question fits in one line: which data, from which country, going back how far? The honest answer keeps a model fitted on the century of 16 developed countries (1870-2020) on screen next to the US data, not buried at the bottom of a menu ([[historical-vs-parametric]]).

**Trap 2: survivorship squared, in your own data.** Sneakier still is the data from your own funds. A portfolio you hold today is made of funds that **still** exist, because the dead ones dropped out of the record. You often picked them for their good decade. And the backfilled histories of recent ETFs borrow indices chosen after the fact. Calibrate on that material and you inherit a built-in optimism. The defense is to calibrate on long, documented **index** series, rebuilding a young fund's missing past from the index it tracks if you have to, rather than on the track record of the survivors. And a well-built central model gets pulled toward a prudent prior anyway ([[making-monte-carlo-relevant]]).

**Trap 3: the short window.** Twenty years of data contain **no** independent long retirement at all. Depending on the window, they contain no persistent inflation and no lost decade either. A simulator that bootstraps your last twenty years explores nothing but recombinations of a single regime. That is the optimistic bound by construction. The audit question is simple: what does the tool do when the history is shorter than the horizon? A good answer says so out loud. It refuses to run cohorts when the windows do not exist. And it slides the parametric model toward the world prior in proportion to the history it lacks.

## The engine traps: how the machine draws

**Trap 4: independent draws (i.i.d.).** Standard Monte Carlo draws each year with no memory. Real markets come in clusters: multi-year bear markets, valuation trends. So the long mediocre stretches, the ones that ruin a plan, come out too rare. Failure is understated by 20 to 50% in relative terms for plans that are sensitive to sequence ([[sequence-of-returns]]). The audit question: can bad years follow one another in your engine? The defense is to run at least one engine with memory against the i.i.d. central case: block bootstrap, historical cohorts or Markov regimes. The gap between the central case and the sequence stress then **measures** this trap on your own plan.

**Trap 5: the normal distribution.** Thin tails make disasters all but impossible ([[fat-tails]]). That cuts failure by 30 to 80% in relative terms. And most commercial simulators are Gaussian without saying so. The audit question: what probability does your model give a year at −30% real? If the answer works out to once every four centuries, you know where you stand. The defense is a Student-t, with its df fitted to the monthly kurtosis observed in your own funds.

**Trap 6: coarse time steps.** Simulating in **annual** steps, with the withdrawal taken at the start of the year, smooths away everything that happens inside the year. The real retiree sells every month, including through the six months when the market fell 30%. The effect on failure is modest but real, a few tenths of a point, and above all it produces paths that could not happen. The defense is to draw and compound monthly, on a panel of monthly returns, with monthly withdrawals that look like a paycheck instead of one lump taken on January 1.

## The scope traps: what the machine leaves out

**Trap 7: the frictions left out.** Management fees, tax on sales, trading costs: most historical grids ignore all of them, and Trinity is gross of everything ([[the-trinity-study]]). Together they are worth 0.5 to 1.5 points of withdrawal rate, the difference between 4% and 2.8% net inside an expensive account. The audit question: is your success rate before or after fees and taxes? An honest simulator grosses every sale up for tax, at the blended rate you actually pay, with an embedded gain fraction that rises over the plan ([[us-taxes-in-the-withdrawal-phase]]). Fund fees are a different matter: they already sit in the prices, so in any total-return series. What is left is the account-level fees, and those belong in your spending ([[how-much-you-need]]).

**Trap 8: laboratory spending.** The withdrawal that stays "constant in real terms" is a convenient fiction. Real spending has a drift of its own, driven by health care, and a smile-shaped profile ([[spending-in-retirement]]). It comes in lumps: a roof, a stretch of long-term care. Above all it follows a **personal inflation** rate that can drift away from the index ([[tracking-inflation]]). The effect cuts both ways, but leaving out the health drift is the costliest omission over a long horizon. A good tool makes the real drift adjustable, offers Blanchett's smile as an option, and simulates realistic spending rules natively (flexible cuts, guardrails, VPW, ABW) instead of the rigid caricature alone.

## The reading traps: what we make the machine say

**Trap 9: binary success.** "95% success" treats failure at 71 exactly like failure at 94 with a pension already running. It puts the path that ends at 40 times the original stake on the same footing as the one that ends at $3,000 ([[failure-probability]]). Absurd decisions follow at the margin: working three more years to go from 94 to 97%, when every failure was late and mild. The defense is to unfold the binary. A complete tool gives the date and the cause of each failure, the distribution of ending wealth (the bequest), failure crossed with mortality, and the standard of living actually delivered year by year.

**Trap 10: the user optimizing the verdict.** This is p-hacking with scenarios: nudging the mean return μ up "because my fund did better", rerunning until the number feels comfortable, keeping the model with the friendliest verdict, rounding a 7% failure down to "about 5%". It is the last trap, and no software can stop it. Software can only make it **visible**. An honest tool shows all its models side by side by design, without crowning one of them, and keeps its outside anchors in view: the CAPE, the world-sample prior. It documents every assumption instead of leaving it as a silent setting. And an equivalent-moves solver restates any gap as a concrete price, in dollars, in years or in flexibility, rather than as a haggle over assumptions. On the behavioral side, the defense is the eight rules for using the machine well ([[monte-carlo-strengths-and-limits]]) and an annual review with audited inputs ([[the-annual-review]]).

::: attention The commercial trap, in one remark
Simulators are not born neutral. An asset manager's tool has every reason to project encouraging numbers, because money is gathered on hope. An annuity seller has every reason to show frightening failure rates, because annuities are sold on fear. A planner billing by the hour has every reason to favor complexity. This is not conspiracy, it is microeconomics. Always ask who built the tool and what they sell. The tools with no product to sell (cFIREsim, FICalc, ERN's toolbox, the simulator behind this book) are not necessarily better engineered, but their errors have no preferred direction.
:::

## The full audit checklist

The ten questions to put to any simulator, and the trap each one tests:

| # | Audit question | Trap tested |
|---|---|---|
| 1 | Which data, from which country, how far back? | US and survivorship bias |
| 2 | Where do my asset assumptions come from? | Survivorship squared, backfill |
| 3 | What happens if the history is shorter than the horizon? | Short window |
| 4 | Can bad years follow one another? | i.i.d. |
| 5 | What probability for a year at −30% real? | Gaussian tails |
| 6 | What time step, and when are withdrawals taken? | Aggregation |
| 7 | Before or after fees and taxes? | Frictions |
| 8 | Can I model my real spending rules? | Laboratory spending |
| 9 | What do I learn about the failures (date, cause, severity)? | Binary success |
| 10 | What stops me from telling myself stories? | P-hacking, incentives |

A tool that answers eight of these ten clearly is an instrument. A tool that documents two of them is a brochure.

::: exemple Five verdicts on one plan, reconciled
The plan: $1.2M, $45,000 a year, 50 years. Simulator A (Gaussian, US data, gross of fees) says 97% success. Simulator B (US historical replay) says 95%. Simulator C (i.i.d. Student-t, prudently calibrated) says 91%. Simulator D (with memory, world sample) says 88% under its sequence stress and 84% under its broad sample. The spread from 84 to 97 is **not** a technical disagreement to settle by averaging. It is a gradient of traps removed one at a time: fees and taxes, −2 points; tails, −2; the memory of bear markets, −3; the world sample, −4. Reading it professionally is simple. The plan lives somewhere in the lower half of the gradient, and the high numbers mostly measure what their tools ignore. Every comparison of simulators should read like that column of subtractions, never like a contest.
:::

## The essentials

- Five simulators, five verdicts, and every gap is a trap either removed or left in place. Ranked by damage: biased data > uncertain parameters > thin tails > memoryless draws > omitted frictions > naive reading.
- The data traps (US bias, survivorship, short window) dominate everything else: **always** ask what the machine learned on before you look at its verdict.
- The engine traps (i.i.d., Gaussian, annual steps) are tested with one question: what is the probability of a year at −30% real, and can another one follow it?
- The scope traps (fees, taxes, idealized spending) are worth 0.5 to 1.5 points of withdrawal rate: a success rate gross of everything is a brochure number.
- The last trap is the user. No machine prevents p-hacking with scenarios, and the defense is procedural: audited inputs, several models, decide on the harsh ones, review once a year.

---

## Going further

- Early Retirement Now, part 26 (what the 4% rule leaves unsaid) and the series' running critique of simulators ([[the-ern-series]]).
- Anarkulova, Cederburg, O'Doherty and Sias (2023): trap 1, measured ([[anarkulova-cederburg]]).
- Kitces.com, "Does Monte Carlo Analysis Actually Overstate Tail Risk In Retirement Projections?": the intelligent counterpoint, where mean reversion in valuations pulls the long fans back in. Useful for not tipping into blanket doom.
- In this book: [[monte-carlo-strengths-and-limits]] (the machine), [[historical-vs-parametric]] (the families), [[making-monte-carlo-relevant]] (the corrections to demand from an engine), [[under-the-hood]] (how the simulator answers, trap by trap).
