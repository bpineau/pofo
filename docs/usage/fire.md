# FIRE and decumulation

The decumulation engine answers one question: can this capital fund this
spending, for how long, and with what chance of running out? It works in
three places: a web simulator, two example scripts and a Go library. The
[book](#the-book) explains the ideas behind every choice.

Everything is in **real** money: spending keeps its purchasing power, returns
are net of inflation, and pensions are entered in today's money.

## The simulator

```sh
./pofo -fire                                            # the simulator alone
./pofo -fire examples/portfolios/fire-decumulation-core.txt   # seeded from a portfolio
```

It opens in your browser, and every slider recomputes the plan: capital,
spending floor, cash-buffer years, expected real return and volatility, tail
thickness, horizon, pension, spending rules and French taxes. The engine runs
in Go, in parallel; the page only draws.

The page reads top to bottom as one argument:

1. the same plan under every **return model**: fat-tailed Student-t, a
   sequence stress, a broad sample of developed-market history, a lost
   decade;
2. today's **valuation** (the Shiller CAPE) and what it implies;
3. the simulated **wealth fans**;
4. the plan replayed through the **worst retirements on record** (USA 1929,
   1966 and 2000, Japan 1990);
5. ruin split by the **first decade's return**, which is where sequence risk
   lives;
6. the **spending** actually delivered, and how it was funded;
7. ruin crossed with **mortality**: alive and broke, or gone first;
8. the **levers** that move ruin most, and the cash-buffer trade-off.

Given a portfolio file, the simulator derives its return assumptions from the
holdings' own history, reconstructed back with `SIM` and deflated by French
inflation. Three models then use it: a parametric fit, a historical bootstrap
(blocks of months resampled) and historical cohorts (every actual start
month). You can drag each holding's weight and watch ruin move.

The same page runs inside the web app at `/firesimulator/`
([web guide](web.md)), where each portfolio of a report links to it.

### Reading it honestly

Read ruin probabilities as orders of magnitude, and compare models rather
than trusting one. The defaults were chosen to be cautious on purpose:

- a headline uses the **fixed** rule everyone means; flexible spending is an
  explicit opt-in;
- expected returns default low, and a broad-sample prior (developed markets,
  not only the United States) is one toggle away;
- a regime model clusters bad years, so a Japan-like decade can happen;
- the horizon defaults past a typical FIRE life expectancy.

Each of these was measured: switching them all to the optimistic side took
the ruin of a 4 % plan over 40 years from about 22 % to 83 %. The simulator
is a tool to explore hypotheses, not investment advice.

## From the command line, as scripts

Two [example scripts](../../examples/code/README.md) run the engine without a
browser, offline, on the bundled US 60/40 from 1954:

```sh
go run examples/code/fire.go -capital 1200000 -spend 42000 -years 40
```

```text
history: 72 real years, 1954 to 2025, 5.53 %/yr real, worst year -22.8 %

bootstrap: 5.8 % of 20000 paths ran out; median real wealth at the end 4161084, 5th percentile 0
spending at 5 % ruin: 41093 a year, 3.42 % of the capital

history: 33 start years, 0 ran out; the worst start, 1966, ended with 294154
```

`replay.go` answers a different question: what life would each withdrawal
rule have given, year by year, from a given start?

```sh
go run examples/code/replay.go -start 1966 -years 35
```

## From Go

Three packages share the work.

| Package | Role |
|---|---|
| `pkg/scenario` | draws real-return paths: parametric (`ParametricSource`), regime-switching (`MarkovRegime`), resampled history (`BlockBootstrap`, `StationaryBootstrap`, `PooledBootstrap`), every start date (`HistoricalCohorts`); `Deflate` turns prices and a price index into real returns |
| `pkg/decumul` | the withdrawal engine: a `Plan` over a `scenario.Source`, `Simulate` for ruin and outcomes, `Solve` for the spending or capital that meets a target, `Sweep1D`/`Sweep2D` for maps; `Draw` once and ask several questions `On` the same paths (`SimulateOn`, `RuinProbOn`, `SolveOn`), and `RuinProb` when the ruin is all you read |
| `pkg/replay` | the seven canonical withdrawal rules run over history as it happened, without randomness |

A plan, its ruin, and the spending that keeps ruin at 5 %:

```go
// from decumul.Example_fire
p := decumul.Plan{
	Capital: 1_000_000, NeedAnnual: 32_000, Years: 35,
	Tax:    decumul.CTOFlatTax{Rate: 0.314},
	Source: scenario.ParametricSource{Mu: 0.035, Sigma: 0.12, Df: 6, Periods: 35},
}
o := p.Simulate(20_000, 4, 7).Outcome() // paths, workers, seed
fmt.Printf("ruin %.0f%%, median terminal wealth %.1f M\n", o.RuinProb*100, o.TerminalP50/1e6)

spend := p.Solve(0.05, decumul.WithdrawalAxis(10_000, 100_000), 20_000, 4, 7)
fmt.Printf("5%% ruin at %.0f a year\n", math.Round(spend/500)*500)
```

```text
ruin 25%, median terminal wealth 0.5 M
5% ruin at 22000 a year
```

Rates are fractions (`0.035` is 3.5 %/yr real). `Tax` is optional; spending
rules (guardrails, a percentage of wealth, amortization) and pensions are
fields of the same `Plan`.

### Lifetimes, couples and estates

By default a plan runs its `Years` for certain. Set `Plan.Lifetime` and the
household's lifespan is drawn inside every path instead. Ruin then counts
only while someone is alive, the estate at death becomes an output, and a
survivor can spend less:

```go
// from decumul.ExamplePlan_Simulate_lifetime
partner := decumul.Life{Age: 60}
p := decumul.Plan{
	Capital: 1_000_000, NeedAnnual: 32_000, Years: 50, PlanHorizon: 35,
	Tax:    decumul.CTOFlatTax{Rate: 0.314},
	Source: scenario.ParametricSource{Mu: 0.035, Sigma: 0.12, Df: 6, Periods: 50},
	Lifetime: &decumul.Lifetime{
		Self:          decumul.Life{Age: 60}, // FrenchMortality by default
		Partner:       &partner,
		SurvivorSpend: 0.7,
	},
}
o := p.Simulate(20_000, 4, 7).LifeOutcome()
fmt.Printf("broke while alive %.0f%%, for %.1f years on average\n", o.RuinAlive*100, o.BrokeYearsMean)
fmt.Printf("median life %.0f years, still alive at the horizon %.0f%%\n", o.MedianLifeYears, o.OutlivedPlan*100)
fmt.Printf("estate p50 %.1f M, nothing left %.0f%%\n", o.EstateP50/1e6, o.EstateZero*100)
```

```text
broke while alive 11%, for 1.0 years on average
median life 31 years, still alive at the horizon 0%
estate p50 0.8 M, nothing left 11%
```

`Years` is the simulation length and `PlanHorizon` the horizon the spending
rules plan over: the household never sees its own drawn death.

### History as it happened

`replay.Run` has no randomness: one start year, the real sequence of returns,
and the income each rule actually paid.

```go
// from replay.ExampleRun
res, err := replay.Run(replay.Setup{
	Start: 1973, Capital: 600000, Spend: 24000, Years: 40,
	Mu: 0.045, Sigma: 0.10, Df: 5, TargetRuin: 0.05, RaiseCap: 1.5,
})
if err != nil {
	panic(err)
}
fmt.Printf("%d-%d, real CAGR %.1f%%, worst year %d at %.0f%%\n",
	res.Start, res.End, res.CAGR*100, res.WorstAt, res.WorstYear*100)
for _, r := range res.Rules[:2] {
	fmt.Printf("%-10s mean %.0fk, leanest %.0fk, %d lean years, %.0fk left\n",
		r.Name, r.Mean/1000, r.Low/1000, r.LeanYears, r.Final/1000)
}
```

```text
1973-2012, real CAGR 4.7%, worst year 1974 at -23%
Fixed      mean 24k, leanest 24k, 0 lean years, 20k left
Flex -10%  mean 22k, leanest 22k, 33 lean years, 344k left
```

`go doc github.com/bpineau/pofo/pkg/decumul` holds the full design, including
the spending-rule precedence and the realism measurements above.

## The book

"Le FIRE tranquille" (French) and "The Quiet FIRE" (English) cover
decumulation from first principles: withdrawal rules, sequence risk,
resilient portfolios, cash buffers, taxes. Read it at
[pofo.zouh.org/firebook/en/](https://pofo.zouh.org/firebook/en/), in the
local web app, or as an EPUB (`pofo -export-epub the-quiet-fire.epub
-book-lang en`). Its article
[Using the FIRE simulator](https://pofo.zouh.org/firebook/en/using-the-fire-simulator)
is the page-by-page manual of the simulator.
