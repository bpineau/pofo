// Package portfolio reads portfolio descriptions and simulates them over
// time.
//
// # File format
//
// One line per asset:
//
//	<weight in %> <identifier> [fees in %/year]
//
// Everything after a # is a comment, and nothing else may follow the
// optional fee column; blank lines are ignored; the weight and fees accept
// a decimal comma and a % suffix. Weights that do not sum
// to 100 are normalized (warning in Spec.Warnings). "#meta key:value"
// lines carry directives:
//
//	#meta rebalance:N     rebalance every N days (0 = never)
//	#meta extra-fees:X    yearly fees applied to the whole portfolio
//	                      (envelope, managed account), deducted by Simulate
//	#meta sim:on          fetch every holding through its SIM (backcast-
//	                      extended) variant, as if each id carried the "SIM"
//	                      suffix (Spec.Sim); a holding with no backcast falls
//	                      back to its real quotes (see Build, SimFetchID)
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
//	                      ones; OBJ is max-sharpe, min-volatility,
//	                      risk-parity, black-litterman or any of the others
//	                      pkg/optimize supports, with optional constraints
//	                      (",max-weight:40"). Parse only records the request
//	                      in Spec.Optimize; the caller runs it. Under
//	                      black-litterman the written weights are not merely
//	                      the baseline: they are the PRIOR the views tilt.
//	#meta currencies:C,D  evaluate the portfolio in several base currencies
//	                      (Spec.Currencies); the caller builds one column per
//	                      currency. Cannot be combined with optimize.
//
// Interpreting identifiers (tickers, ISIN, aliases, SIM suffix) is the
// caller's job: Build turns a parsed Spec into a simulatable Portfolio
// through a fetch callback, typically marketdata.Client.FetchExtended.
// Parse → Build → Simulate is the whole pipeline in three calls:
//
//	spec, _ := portfolio.ParseFile("p.txt")
//	p, _ := portfolio.Build(spec, portfolio.BuildOptions{
//		Fetch: func(id string) (*marketdata.Series, error) {
//			return client.FetchExtended(ctx, id, marketdata.FetchOptions{Currency: "EUR"})
//		},
//	})
//	sim, _ := portfolio.Simulate(p, 90)
//
// A portfolio assembled in code rather than read from a file starts from
// NewSpec instead of Parse. Its Line takes the weight as a FRACTION (the
// in-memory convention) and runs through the very validation and
// normalization Parse applies, so NewSpec on some lines and Parse on the
// same lines written as a file give the same Spec:
//
//	spec, _ := portfolio.NewSpec("60/40",
//		portfolio.Line{ID: "IWDA", Weight: 0.6, Fees: -1},
//		portfolio.Line{ID: "AGGH", Weight: 0.4, Fees: -1})
//
// Directives are Spec fields there (RebalanceDays, Sim, ...), left unset.
// For the numbers of such a portfolio in one call (statistics, per-holding
// studies, correlation, risk budget, look-through), see pkg/analyze.
//
// # Simulation
//
// Simulate replays the portfolio at base 100 over the union of the quoting
// calendars (prices forward-filled via marketdata.Align), from the first
// day every asset trades to the last day they all still trade, rebalancing
// back to the target weights every N calendar days and deducting envelope
// fees as it goes. Asset TERs are never deducted: they are already reflected
// in prices.
//
// The ASSETS alone decide that calendar. The financing rate of a levered
// portfolio is read onto it (marketdata.SampleAt) rather than merged into it,
// and held flat before its own history starts instead of reading as 0 %/yr;
// SimResult.Warnings says when it had to be.
//
// Everything quoted PER YEAR (envelope fees, the financing rate, the borrow
// spread) accrues over the calendar time each step spans, 365.25 days to the
// year. Charging a fixed 1/252 per quote instead only agrees with the quoted
// rate on a daily calendar: a weekly-quoting holding would pay a fifth of its
// envelope fee and a monthly one a twentieth.
//
// Along the way it attributes each day's time-weighted return to its
// holdings (SimResult.Contributions: held shares x price move / value;
// envelope fees and the leverage cash leg stay unattributed).
// SimResult.MonthlyContributions folds the attribution into calendar
// months, the input for contribution timelines and per-regime views.
//
// # Units
//
// Holding.Weight, Asset.Weight and Line.Weight are FRACTIONS (0.60 = 60 %);
// RawWeight is the PERCENT a file writes; Fees (Line's included),
// EnvelopeFees and BorrowSpread are PERCENT per year as written in
// portfolio files. The simgen package uses fractions for its own fee
// parameters; do not mix them up.
package portfolio
