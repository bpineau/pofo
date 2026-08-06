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
package decumul
