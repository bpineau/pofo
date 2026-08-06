// Package portfolio reads portfolio descriptions and simulates them over
// time.
//
// # File format
//
// One line per asset:
//
//	<weight in %> <identifier> [fees in %/year] [free text]
//
// Everything after a # is a comment; blank lines are ignored; the weight
// and fees accept a decimal comma and a % suffix. Weights that do not sum
// to 100 are normalized (warning in Spec.Warnings). "#meta key:value"
// lines carry directives:
//
//	#meta rebalance:N     rebalance every N days (0 = never)
//	#meta extra-fees:X    yearly fees applied to the whole portfolio
//	                      (envelope, managed account), deducted by Simulate
//	#meta leverage:on     weights kept as written; the residual
//	                      (100−sum) is a cash position, negative and
//	                      financed at the cash rate plus spread above 100 %
//	#meta borrow-spread:X borrowing spread in %/year (caller's default)
//	#meta capital:X       starting amount (required for flows)
//	#meta contribute:A/P  add amount A every period P (week, month,
//	                      quarter or year), e.g. contribute:500/month
//	#meta withdraw:A/P    take A (or A% of the value: withdraw:4%/year)
//	                      out every period P
//	#meta optimize:OBJ    compute the weights instead of using the written
//	                      ones; OBJ is max-sharpe, min-volatility or
//	                      risk-parity, with an optional ",max-weight:40"
//	                      cap (see pkg/optimize). Parse only records the
//	                      request in Spec.Optimize; the caller runs it.
//
// Interpreting identifiers (tickers, ISIN, aliases, SIM suffix) is the
// caller's job — see marketdata.Client.Fetch and marketdata.SplitSim.
//
// # Simulation
//
// Simulate replays the portfolio at base 100 over the union of the quoting
// calendars (prices forward-filled via marketdata.Align), from the first
// day every asset trades to the last day they all still trade, rebalancing
// back to the target weights every N calendar days and deducting envelope
// fees daily. Asset TERs are never deducted: they are already reflected in
// prices.
// # Units
//
// Holding.Weight and Asset.Weight are FRACTIONS (0.60 = 60 %); RawWeight,
// Fees, EnvelopeFees and BorrowSpread are PERCENT per year as written in
// portfolio files. The simgen package uses fractions for its own fee
// parameters — do not mix them up.
package portfolio
