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
package decumul
