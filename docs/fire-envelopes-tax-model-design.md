# FIRE: per-envelope tax model, design spike and decision

Date: 2026-09-10. Question from the program TODO: should the FIRE simulator
gain complexity to model a French PEE (employee-savings plan), and more
generally is a per-envelope tax model (CTO, PEA, assurance-vie, PEE, each with
its own rate, cost basis and unlock date) worth its complexity against today's
single blended "Tax on gains" slider?

Short answer: no new tax structure. The kernel already carries the per-envelope
model, it is worth about 0.015 point of withdrawal rate over a correctly
calibrated blended rate, and the two things that really move the number are the
CALIBRATION of that rate (0.09 to 0.16 point) and the EMBEDDED GAIN FRACTION
(0.30 point), which the page does not even expose. The calibration recipe is
below; the UI item goes to the backlog.

## Rates used here, and where they come from

Never quoted from memory. Read from `~/github/paperasse/fiscaliste/data/`:

- `pfu-prelevements-sociaux.json`: LFSS 2026 (loi n° 2025-1403 du 30/12/2025,
  art. 12) raises CSG on capital from 9.2 to 10.6 %, so social levies go from
  17.2 to 18.6 % and the PFU from 30 to 31.4 % (12.8 % IR + 18.6 % PS). Under
  `revenus_du_patrimoine` (art. L. 136-6 CSS): securities gains on a CTO, at
  **31.4 %**, already for 2025 income. Under `produits_de_placement_2026`
  (art. L. 136-7 CSS): dividends, interest and **PEA exit gains**, at 31.4 %
  full rate or **18.6 % of social levies alone** where income tax is waived,
  for cash received from 01/01/2026. Assurance-vie and capitalisation contracts
  are listed in `taux_inchanges_17_2`: they stay at **17.2 %** permanently.
- `pea-assurance-vie.json`: PEA past 5 years, free withdrawals, total income
  tax exemption, social levies on gains (the file states 17.2 %, on the 2025
  basis; for a withdrawal taken from 2026 the rate to use is the 18.6 % of
  `produits_de_placement_2026` above). Assurance-vie past 8 years: annual
  allowance of 4 600 EUR alone or **9 200 EUR** for a couple on the gain share
  of the withdrawal, then 7.5 % + 17.2 % = **24.7 %** below 150 kEUR of net
  premiums, and the pro-rata rule (a partial withdrawal never returns capital
  first).
- `equity-salarial.json`, `pee_perco`: PEE employer match exempt from income
  tax and social levies within the caps (3 709 EUR in 2025, 8 % of the PASS),
  **5-year lock** with the legal early-release cases, and on exit past 5 years
  "exonération IR, PS sur les gains uniquement", so **18.6 %** on gains alone
  under L. 136-7 as amended. Same file: the **10 % employee contribution**
  (`contribution_salariale`) on the acquisition gain of qualifying free shares
  and options.

## 1. What the engine does today

The tax is an interface, `Tax.GrossUp(net, growth, cost)` in
`pkg/decumul/plan.go`, which returns the gross to sell, the new cost basis and
the tax paid. `CTOFlatTax` (same file) taxes the GAIN FRACTION of every sale
only: `gainFrac = 1 - cost/growth`, effective rate `Rate * gainFrac`, and the
sale is grossed up (`gross = net / (1 - eff)`) because the tax is itself paid
out of the sale. The basis is a weighted average carried pro rata:
`newCost = cost * (1 - gross/growth)`. So the effective burden starts low and
drifts toward the nominal rate as unrealised gains compound. The buffer sleeve
is untaxed; withdrawals hit it first under its drawdown rule.

The per-envelope model already exists (`pkg/decumul/envelope.go`, shipped with
the lifetime kernel work): `Plan.Envelopes []Envelope`, each with `Amount`
(pro rata of the growth sleeve), `GainFrac` (embedded unrealised gain at year
0) and its own `Tax`; `newPockets` carves the sleeve per path, `pocketOps.sell`
drains the pockets IN SLICE ORDER, and `liquidationNet` prices the whole book
after tax for the amortization rule. `AVTax` is a stateful `YearlyTax`
implementing the assurance-vie annual allowance, cloned per path and reset each
year. A nil `Plan.Envelopes` keeps the historical single sleeve taxed by
`Plan.Tax`, whose basis starts EQUAL to its value (zero embedded gain).

The web layer maps it in `pkg/decumul/web/model.go`: `Params.PEACapital`,
`Params.AVCapital` and `Params.GainFrac` feed `Params.envelopes()`, which
builds CTO (at the slider rate) then PEA (18.6 %) then AV (`AVTax{0.247,
9200}`) in that drain order. But NONE of those three fields is in the page's
control list (`GROUPS` in `pkg/decumul/web/assets/app.js`), so they are always
zero in practice: the shipped page runs the single sleeve at the default
32.8 % rate with NO embedded gain. The slider's help text says to use a blended
effective rate but gives no recipe.

## 2. Candidate models, cheapest first

(a) **Status quo plus a calibration recipe.** One rate, one gain fraction, both
documented. Zero code.

(b) **Ordered envelopes.** Already implemented in the kernel; what is missing
is the UI and, if anyone wants them, unlock dates (PEE and PEA 5 years, AV 8
years for the allowance).

(c) **Full per-envelope simulation with distinct allocations per envelope.**
Needs one correlated return path per envelope inside every path, per-envelope
rebalancing and an asset-location policy. `scenario.Source` yields one sequence
per path, so this is a new kernel. YAGNI: asset location is a second-order
effect that a ruin probability cannot see.

## 3. The measurement

Throwaway program (scratchpad, not committed) over `pkg/decumul` +
`pkg/scenario`: 1.5 MEUR, 40 years, 3 years of buffer at +0.5 % real,
parametric source mu 5 %, sigma 11 %, df 5 (the page's defaults), 8 000 paths,
`Plan.Solve` on `WithdrawalAxis` at a 10 % ruin target, and ruin at a 3.5 %
withdrawal rate. Split 25 / 15 / 60 PEA / PEE / CTO at 18.6 / 18.6 / 31.4 %,
capital-weighted blend 26.3 %. Solve shares one drawn path set across variants,
so the gaps below are not sampling noise; they were reproduced on four seeds
(+0.015 and -0.014 point, spread under 0.005).

| Model, embedded gain 50 % | ruin at 3.5 % | sustainable WR at 10 % ruin |
|---|---|---|
| blended 26.3 %, one sleeve | 26.75 % | 2.854 % |
| explicit CTO then PEA then PEE | 26.10 % | 2.867 % (+0.014 pt) |
| explicit PEA then PEE then CTO | 27.73 % | 2.837 % (-0.017 pt) |
| default slider 32.8 %, one sleeve | 33.39 % | 2.731 % (-0.123 pt) |

Same picture at a 30 % embedded gain (+0.019 / -0.024 / -0.090 pt) and at 70 %
(+0.015 / -0.009 / -0.155 pt). The 60 / 25 / 15 CTO / PEA / AV book with
unequal gain fractions (45 / 60 / 35 %) gives +0.021 point for the explicit
model over the recipe blend, of which 0.016 point is the assurance-vie
allowance. Scale: the whole tax is worth 0.48 point of withdrawal rate here
(3.331 % untaxed against 2.854 %), and starting from a zero embedded gain
rather than 50 % is worth 0.30 point (3.152 % against 2.854 %).

So: structure 0.015 point, drain order 0.03 point, rate calibration 0.12 point,
gain fraction 0.30 point. The structure is one twentieth of the gain fraction
the page cannot even set. Under the 0.2-point bar of the question, the status
quo wins.

**The calibration recipe.** With capital shares `w_i`, embedded gain fractions
`g_i` and marginal rates on gains `r_i`, set

    Gain fraction = sum(w_i * g_i)
    Tax on gains  = sum(w_i * g_i * r_i) / sum(w_i * g_i)

that is, a GAIN-weighted rate (the tax is levied on gains, not on capital) and
a capital-weighted gain fraction. Worked example, 60 / 25 / 15 CTO / PEA / AV
with gain fractions 45 / 60 / 35 % and rates 31.4 / 18.6 / 24.7 %: gain
fraction 47.3 %, blended rate 26.6 %. Measured against the explicit three-
envelope run: 0.021 point of withdrawal rate. Two riders: where a wrapper is
covered by an allowance (assurance-vie past 8 years, 9 200 EUR of gains a year
for a couple, worth roughly 2 400 EUR of tax) set its `r_i` to 0 for as long as
the yearly realised gain stays under it; and any envelope taxed only on gains
still needs its OWN `g_i`, since a PEA opened in 2014 and a CTO funded last
year do not carry the same embedded gain.

**Unlock dates are not worth modelling.** A lock binds only when the liquid
pockets cannot carry spending to the unlock date. At a 3.5 % withdrawal rate
five years cost 17.5 % of initial capital, and a market drop shrinks the locked
and the liquid pockets alike, so the PEE's 5-year lock binds only when the
locked share is over roughly four fifths of the book. Besides, a FIRE plan's
year 0 normally sits after those clocks have run: the 5-year PEA and PEE and
the 8-year assurance-vie are accumulation-phase deadlines, and their place is
the book's tax chapter, not the kernel.

## 4. The PEE specifics: match against tax, not "more money at a higher rate"

Three facts settle the framing. The 10 % employee contribution on the
acquisition gain of qualifying free shares (`equity-salarial.json`) is due
whatever the destination, so it cancels between the alternatives and never
enters the arbitrage. The employer match is exempt from income tax and social
levies within the caps, so it is capital that compounds from day one. And the
exit past 5 years is social levies on gains ONLY, 18.6 %, which is LOWER than
the 31.4 % PFU of a taxable account, not higher.

One unit contributed to a wrapper taxed at `rA` with a match `m`, against the
same unit in a wrapper taxed at `rB`, both compounding to the same multiple and
exited with a gain fraction `g` at the end, breaks even at

    m* = g * (rA - rB) / (1 - rA * g)

With the PEE as A (`rA` = 18.6 %) and a gain fraction of 50 %: against a CTO
(31.4 %) `m*` = -7.1 %, that is the PEE wins by a 7 % head start with NO match
at all; against a PEA (18.6 %) `m*` = 0, the wrapper is a wash and the match is
pure gain; against a pocket whose gains are effectively untaxed (assurance-vie
under its annual allowance) `m*` = +10.3 %, so any match above about ten
percent of the contribution beats it. At a 70 % gain fraction those become
-10.3 %, 0 and +15.0 %. Real French match schedules start at 25 % and often
reach 100 to 300 % of the contribution (`equity-salarial.json` calls the match
an immediate +30 % to +300 % lever), which is an order of magnitude above every
break-even above. The PEE question is therefore not a tax question at all, and
does not need a tax model: what it needs is the cap (3 709 EUR of match in
2025) and the reminder that the choice inside the plan is usually between a
world-equity fund and cash, which the simulator already covers through the
portfolio it is given.

## 5. Recommendation

Keep the single blended rate as the shipped control and document the recipe of
section 3 in the slider's help text; do not extend the tax structure, and do
not add unlock dates or per-envelope allocations. The per-envelope kernel stays
where it is: it costs nothing to keep, it is exercised by
`pkg/decumul/envelope_test.go`, and it is the honest answer for anyone driving
the API directly (a household whose assurance-vie allowance really is the
binding constraint). What IS worth doing is the cheap UI item the measurement
exposed: put the embedded gain fraction on the page (it is worth 0.30 point of
withdrawal rate, twenty times the structure question) and, next to it, the two
envelope amounts that already exist in `Params`, so the recipe becomes optional
rather than mandatory. That item goes to
`docs/decumulation-fire-program-2026-07.md`; the tax-model question is closed
with the decision recorded there as well.
