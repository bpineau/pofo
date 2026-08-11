# Weight search: bounded optimization, constrained objectives, held-out fitting, per-sleeve sweep

`#meta optimize:` answers "what weights maximize X over this window". That is
one question, asked in the least useful way: unconstrained, single-objective,
and fitted on the same data it is then judged on. Three things were missing,
all of which had to be done by hand every time a portfolio file's weights were
revised:

1. **Bounds per line.** An unconstrained optimum is a corner solution. Every
   hand-revision of a portfolio file starts by deciding, per sleeve, a range
   that is defensible for reasons the backtest cannot see, then optimizing
   inside it. The optimizer could not express that, so the ranges lived only in
   the files' prose.
2. **Constraints instead of one objective.** The real question is rarely "the
   best Sharpe": it is "the most return I can get without going above X
   volatility", or "the least volatility that still returns Y". Those are
   constrained problems, and scalar objectives answer them only by accident.
   Sharpe in particular is a poor proxy here because this toolkit computes it
   at a ZERO risk-free rate (`pkg/metrics/doc.go`), so any cash-like sleeve
   buys ratio for free: a euro-linker sleeve raised the measured Sharpe of a
   research build while cutting 2.5 points of CAGR.
3. **A held-out window.** An optimizer that reports its own in-sample result is
   a machine for manufacturing overfitted weights with authority. Fitting on a
   sub-window and letting the report measure the result on the full one makes
   the number on screen honest by construction.

And one adjacent tool, `-sweep`, which is what actually produced the RANGE
CHECK sections of the hydra example files: vary ONE sleeve, renormalize the
rest, and read what each weight costs and buys.

Non-goals of this pass: multi-objective/Pareto search (`improve`), the
efficient-frontier chart in the HTML report. Both are designed for, not built:
see "What this leaves for later".

## 1. Bounds per line

`optimize.Spec` gains a resolved, per-asset box:

```go
Lower, Upper []float64 // per asset, fractions; nil = fall back to the scalars
MinWeight, MaxWeight float64 // scalar defaults, fractions; 0 = none / 1
Bounds map[string][2]float64 // parsed but UNRESOLVED, keyed by written id
```

`Bounds` is what `ParseSpec` can produce from text; it is keyed by the
identifier as written in the file, and `pkg/optimize` never looks at it. The
caller (`pkg/compare`) resolves it against the portfolio's assets into
`Lower`/`Upper`, which is what the solver reads. That keeps the solver free of
identifier resolution, which is `pkg/marketdata`'s job and nobody else's.

File syntax, inside the existing comma-separated grammar:

```
#meta optimize:max-sharpe,max-weight:25,min-weight:5
#meta optimize:max-return,max-vol:9.5,bounds:NTSG:15-30,bounds:GDE:10-25
```

`bounds:ID:LO-HI` repeats, once per constrained line; bounds win over the
scalars for the lines they name. Both ends are percentages, `LO-` and `-HI`
allow an open end.

Feasibility is checked up front: Σ Lower ≤ 1 ≤ Σ Upper, otherwise the solve
fails with a message naming the two sums, because that error is otherwise
diagnosed as "the optimizer returned something strange".

The projection generalizes with no new mathematics. `projectCappedSimplex`
solves Σ clamp(vᵢ − λ, 0, cap) = 1 for λ by bisection; the box version solves
Σ clamp(vᵢ − λ, loᵢ, hiᵢ) = 1 the same way. The rest of the projected-gradient
machinery is untouched.

## 2. Constrained objectives

```go
type Limits struct {
    MaxVolatility float64 // annualized, fraction; 0 = no limit
    MinReturn     float64 // annualized GEOMETRIC return (CAGR), fraction; 0 = no limit
    MaxDrawdown   float64 // tolerated depth, positive fraction; 0 = no limit
}
```

Zero means "unset" for all three because a 0 % volatility cap, a 0 % return
floor and a 0 % drawdown cap are all meaningless or infeasible, so no
information is lost. A `worst-5y` floor is the one constraint where 0 is a
real value ("never negative over five years"); it is deliberately left out of
this pass rather than given a second, inconsistent sentinel.

New objective `max-return`: maximize the blend's CAGR. On its own it is
degenerate (100 % of the best line) and it says so in the note; under a
volatility cap it is the frontier point, which is the point of it.

File syntax:

```
#meta optimize:max-return,max-vol:9.5
#meta optimize:min-volatility,min-return:10.5
#meta optimize:max-sharpe,max-drawdown:20
```

### How constraints are enforced

Any constrained solve routes through ONE code path: multi-start projected
descent over the box simplex, on a penalized score

```
score(w) = objective(w) - penalty * Σ relative violations
```

with the violations measured relative to each limit, so the three constraints
are commensurable and the gradient points back toward the feasible set from
anywhere. The unconstrained closed forms (`maxSharpe`'s QP, the min-variance
solve, risk-parity) stay exactly as they are and keep serving the
unconstrained calls: a constraint-free `max-sharpe` must return the same
weights as before this change, and its test asserts that.

The objective is evaluated on the blended daily return path, not on
mean/covariance, because `max-return` targets the GEOMETRIC return and the
drawdown limit is path-dependent. Volatility is the usual √252 annualization
of the blend's daily standard deviation, so the cap reads exactly like the
report's headline "Volatility (annualized)" line.

Reported back in `Result`: the achieved values, and which constraints ended up
binding. A cap that binds is the interesting case (the search wanted more) and
a cap that does not bind means the constraint was not the thing in the way;
saying which is the difference between a number and an answer.

## 3. `train:` (fitting on a held-out window)

```
#meta optimize:max-return,max-vol:9.5,train:..2015
#meta optimize:max-sharpe,train:1996..2015
#meta optimize:max-sharpe,train:2000-01-01..2015-12-31
```

Either end may be a year or a full date, and either may be empty. `Spec.Train`
holds the parsed window, and `pkg/optimize` IGNORES IT: `Solve` is date-free
and stays that way. The caller slices the aligned returns to the training
window, solves there, and builds the report column over the FULL window, so
what the reader compares is an out-of-sample path.

The note then carries both halves:

```
weights computed by the optimizer (max-return, max-vol:9.5) over 1996-03-26→
2015-12-31: NTSG 22.0 %, …; in-sample CAGR 11.4 %/yr, volatility 9.3 %;
on the 2016-01-01→2026-08-07 stretch it did not see, CAGR 8.9 %/yr,
volatility 10.6 %
```

Two guardrails: the training slice must hold at least two years of
observations, and the holdout sentence is emitted only when the holdout does
too, since a six-month holdout says nothing and would read as if it did.

## 4. `pofo -sweep`

For each line of the portfolio, hold the other lines' RELATIVE proportions
fixed, move that line's weight across a grid, and report what happens:

```
$ pofo -sweep examples/hydra-five-engines-capital-efficient.txt

NTSG (written 18 %)
  weight    CAGR     vol  Sharpe    maxDD     TTR  worst5y
     0 %  10.96 %  9.83 %   1.07   -17.4 %   2.0 y   0.68 %
     5 %  10.91 %  9.78 %   1.07   -18.0 %   2.0 y   0.76 %
    …
    18 %  10.77 %  9.89 %   1.05   -19.5 %   2.0 y   0.97 %   <- written
    …
```

Columns: CAGR, volatility, Sharpe, max drawdown, longest recovery (TTR) and
worst rolling 5y CAGR. The first four come straight from `metrics.Stats`; the
last from `metrics.RollingCAGR`, the same call the report's row uses, so the
sweep and the report cannot disagree.

The reusable half lives in `pkg/compare` (`Sweep`), which already owns the
fetch/build/simulate pipeline and is presentation-neutral by charter; the CLI
file `cmd/pofo/sweep.go` only formats. That split is what lets the frontier
chart reuse it later without moving code out of `cmd/`.

Two decisions worth writing down:

- The sweep re-runs the REAL simulation (`portfolio.Simulate`, the file's own
  `#meta rebalance:` period, fees and flows included), not a blend of daily
  returns. It answers "what if I had written a different number in this file",
  and a continuously-rebalanced approximation would answer a different
  question. It costs one simulation per grid point, which is milliseconds.
- Renormalization is PROPORTIONAL over the other lines. Any other rule (fund
  from one named sleeve, fund from cash) is a different experiment and would
  have to say which; proportional is the only choice that needs no argument.

## What this leaves for later

- **`improve`**, the Pareto mode: start from the written weights and look for a
  point that beats them on BOTH axes, answering "am I already on the frontier"
  with a yes. It is two constrained solves on top of what this pass builds
  (`max-return` under the current volatility, `min-volatility` under the
  current return) plus a scalarized walk between them, and it surfaces through
  the same extra column as any other objective.
- **The frontier chart** in the HTML report: a curve of best-CAGR-per-
  volatility-cap with the written portfolio plotted on it. It needs a design
  pass of its own (how many solves, cached where, at what resolution, and
  whether the report can afford them on every render).

## Traps for whoever touches this next

- Sharpe here runs at a ZERO risk-free rate. Constrained objectives exist
  partly to route around that; do not "improve" them by scalarizing back into
  a ratio.
- `Spec.Train` is inert inside `pkg/optimize` by design. If a future caller
  forgets to slice, the optimizer will silently fit the whole window: the
  note's "over <window>" clause is what makes that visible, so keep it
  printing the range it actually received.
- Bounds are resolved by the caller against asset identifiers. An id that
  matches no line is an ERROR, not a silent no-op; a typo in
  `bounds:NTGS:15-30` must not read as "no bound".
- The unconstrained paths are covered by goldens: a constraint-free
  `max-sharpe` returning different weights than before is a regression, not an
  improvement.
