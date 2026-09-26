// Package decumul evaluates decumulation (withdrawal / retirement / FIRE)
// portfolios. It runs a withdrawal kernel over the real-return paths of a
// scenario.Source to estimate the probability of ruin, FIRE outcome metrics
// and parameter sweeps, and to size a starting capital or a cash buffer
// against a target ruin probability.
//
// Everything is in real euros: the spending floor is constant in purchasing
// power, returns are real, pensions are entered as real Cashflows. The
// parametric model is i.i.d. with fat tails and is probably optimistic vs
// multi-country history; pair it with the bootstrap and historical-cohort
// scenario.Sources, and read ruin in relative orders of magnitude. This is a
// hypothesis-exploration tool, not investment advice.
//
// Spending rules. The kernel takes one policy per plan and applies them in a
// fixed precedence: Amortize (ABW/TPAW), Bounded (Vanguard dynamic spending),
// Percent (VPW), RiskGuard (risk-based guardrails), Guard (Guyton-Klinger),
// then the plain fixed rule with its optional Flex cut and Ratchet. The two
// guardrail flavours share one architecture, +-10 % moves when the withdrawal
// rate leaves a band, and differ only in the sensor: Guard bands the rate the
// plan STARTED at, RiskGuard bands the rate still safe for the horizon LEFT
// and reads it on total wealth (portfolio plus discounted future cashflows),
// which is the Kitces/Morningstar answer to the 2006 rule's blindness to age
// and to pensions still to come.
//
// Lifetime. By default a plan runs Plan.Years for certain and mortality enters
// afterwards, as a weighting (Ensemble.LifeCurve). Setting Plan.Lifetime draws
// the household's lifespan inside every path instead, which unlocks what the
// weighting cannot produce: ruin as broke-WHILE-ALIVE counted exactly, the
// estate at death as a first-class output (Ensemble.LifeOutcome,
// Ensemble.Estates), couple dynamics (a survivor spending less, a pension only
// partly reverting via Cashflow.Owner and Cashflow.Reversion), and an optional
// Annuity whose income stops with its annuitant, so the mortality credit is
// realised rather than narrated. It is opt-in: a nil Lifetime is the special
// case where the household is certain to reach the horizon, and every result is
// bit-for-bit what it was.
//
// Two conventions come with it. Plan.Years is the simulation length, so set it
// past any plausible age and use Plan.PlanHorizon for the horizon the spending
// rules plan over: the drawn death is an outcome of the world, never an input
// to a policy, and a rule that amortized over it would describe a clairvoyant
// retiree. And a path's series are FROZEN at the household's end (Wealth holds
// the estate, Spend holds 0) rather than zeroed, so a death is never read as a
// crash; the statistics that are genuinely per-lifetime stop at
// PathResult.LifeYears.
//
// Design and calibration, including what the bundled Gompertz law gets wrong
// and in which direction: docs/stochastic-lifetime-kernel-design.md.
//
// # Realism: the optimism that was measured out (2026-06)
//
// A ruin figure is only as honest as its defaults, and the first defaults
// were measured materially more optimistic than the broad-sample evidence
// (Anarkulova, Cederburg and O'Doherty, "The Safe Withdrawal Rate: Evidence
// from a Broad Sample of Developed Markets", 2023, SSRN 4227132: a fixed 4 %
// rule fails far more often than US-only backtests suggest; their 5 %-failure
// rate is near 2.3 % real). Toggling one assumption at a time on a 1 M, 40-year
// plan at a 4 % withdrawal, ruin moved from 21.8 % to 48.5 % with the flex cut
// off, to 45.6 % at a 3 % real mean instead of 4.5, to 37.8 % at 17 % volatility
// instead of 12, to 38.9 % over 50 years, and to 82.8 % with all of them at
// once. The kernel was sound; the optimism was defaults, a short favourable
// fit and i.i.d. draws that cannot produce a Japan-1990 decade. What guards
// against it since: the flex cut is an explicit opt-in (a headline is the
// FIXED rule everyone means), the return defaults are cautious and a
// broad-sample prior is one toggle away, scenario.MarkovRegime clusters bad
// years so sequence risk exists, and the horizon defaults past a FIRE life
// expectancy. Any change that makes a headline rosier must say which of these
// it undid.
package decumul
