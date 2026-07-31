// Package replay runs the canonical withdrawal rules over the years as they
// actually happened, on one bundled reference portfolio, so that a decumulation
// strategy can be described by the life it delivered rather than by a failure
// probability.
//
// It answers a different question from pkg/decumul, which simulates thousands
// of imagined futures and reports how often a plan fails. Here there is no
// randomness at all: one start year, one sequence of real returns, seven rules,
// and the income each one actually paid out year by year. That is the reading a
// reader asks for first, and no Monte-Carlo summary can give it: how steady the
// cheque was, how deep and how long the lean spells got in a crisis, who ended
// up sitting on a fortune they never spent.
//
// # The reference portfolio
//
// Reference builds a US 60/40 (S&P 500 total return, 5-year Treasury total
// return, rebalanced every January) deflated by the US CPI-U, from 1954. It is
// the yardstick Bengen, the Trinity study and the guardrails literature are all
// written on. Everything comes from the bundled datasets and the embedded CPI
// snapshot, so a replay needs no network and gives the same numbers on every
// machine.
//
// EVERYTHING IS REAL. The index, the returns, the incomes and the capitals are
// all in constant purchasing power (inflation removed), and gross of tax. A flat
// income line therefore means a constant standard of living, not a constant
// cheque.
//
// # The rules
//
// Policies returns the seven canonical rules, from the most rigid to the most
// wealth-proportional: the fixed real withdrawal (Bengen), the same with a flex
// cut, Guyton-Klinger guardrails, risk-based guardrails (Kitces/Morningstar),
// Vanguard's bounded percentage, amortisation-based withdrawal (ABW/TPAW) and
// the pure percentage of portfolio (VPW). Each carries its English and French
// names, a narrow-column tag, a plain-language explanation and a colour, so a
// rule looks and reads the same in every surface that draws it.
//
// # The horizon
//
// Setup.Years is the household's PLAN, not the length of the record. The rules
// that reason about the horizon (the amortisation rule spreading wealth over the
// years remaining, the risk-based guardrail tracking the rate still safe) must
// face the horizon the household actually has, so the plan runs whole and the
// replay reports only the years the record covers. A withdrawal path is causal,
// so this is identical to running the shorter path, except the horizon-aware
// rules are no longer told the retirement conveniently ends when the data does.
//
// Their forward return assumptions (Setup.Mu, Sigma, Df) should be generic
// rather than fitted to the era being replayed: no retiree of 1973 or 2000 knew
// what was coming, and a rule fed hindsight is not the rule the literature
// describes.
package replay
