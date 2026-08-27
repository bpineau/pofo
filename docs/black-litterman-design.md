# Black-Litterman: the file's weights as the prior, views as the input

Status: specified 2026-08-27, SHIPPED 2026-08-27 (`pkg/optimize/bl.go`,
`pkg/datasets/golden/blacklitterman_test.go`). Companion to
`docs/weight-search-design.md`, which owns the box-simplex solver, the
feasibility limits and the `train:` window this objective reuses.

## 1. What it answers

Every mean-variance objective in `pkg/optimize` that touches expected
returns reads them off the sample, and the sample estimates them worst
(Merton 1980; Michaud 1989; DeMiguel, Garlappi and Uppal 2009, all cited in
`doc.go`). The result is a corner solution driven by whichever line was
lucky over the window. The toolkit's standing answer is to keep the means
out of the problem (RiskParity, MinVolatility) or to box the weights
(`bounds`, `min-weight`, `-sweep`).

Black-Litterman (Black and Litterman 1992; He and Litterman 1999) keeps the
means in but anchors them: it starts from a portfolio of reference, derives
the expected returns that portfolio implicitly assumes (reverse
optimization), then updates them with the owner's VIEWS, each carrying a
confidence, by a Bayesian blend. The weights that follow tilt away from the
reference only where a view says so, and only as far as its confidence
allows. It is the model a portfolio owner reaches for when they can state
"I believe X earns Y a year, with this much conviction" and want to see what
that belief is worth in weights.

Measured on a defensive eleven-line decumulation book (2026-08-27, EUR,
2005-11 to 2026-08): the sample-mean utility optimum put 51 % in small
value, 30 % in gold and nothing in linkers or duration; the Black-Litterman
posterior under the owner's own written views, at moderate confidence,
landed every sleeve inside the file's defended ranges but gold. Same
covariance, same solver: only the means changed.

## 2. The model, as implemented

Notation: n assets, `w` the prior weights (fractions, sum 1), `Σ` the
annualized sample covariance (`meanCov`), `λ` the risk aversion.

1. Implied (equilibrium) returns: `π = λ Σ w`.
2. Views: k rows of `P` (n columns), `Q` (k), confidences `c` in (0,1).
   An absolute view on asset i is `P_i = 1, Q = q`; a relative view "A
   beats B by q" is `P_A = 1, P_B = -1, Q = q`.
3. View uncertainty, closed form:
   `Ω = diag((1/c_k − 1) · (P τ Σ Pᵀ)_kk)`. At c = 0.5 this is
   He and Litterman's own choice (`Ω ∝ diag(P τ Σ Pᵀ)`); c → 1 makes the
   view certain, c → 0 switches it off. It is the parametrization Walters
   ("The Black-Litterman Model in Detail", 2014) describes as the
   proportional one; Idzorek's iterative confidence mapping (2005) is NOT
   used: the closed form is monotone in c, needs no inner solve, and lands
   on the reference paper's numbers at c = 0.5.
4. Posterior mean:
   `μ = π + τ Σ Pᵀ (P τ Σ Pᵀ + Ω)⁻¹ (Q − P π)`.
   With Ω proportional to τ, τ cancels from μ; it is fixed at 0.05 (He and
   Litterman) and exposed nowhere. The posterior covariance is not used:
   Σ stays the risk model for the weights, the common practitioner choice.
5. Weights: `argmax μᵀw − (λ/2) wᵀΣw` over the box simplex
   `{Σwᵢ = 1, loᵢ ≤ wᵢ ≤ hiᵢ}`, with the feasibility limits when set.

Property that makes the objective legible and testable: with no view,
`μ = π` and the unconstrained long-only optimum is EXACTLY `w` (the prior
is feasible and the utility's gradient vanishes there after projection).
`optimize:black-litterman` on a file with no view therefore returns the
file's own weights and reports what returns they imply. Max-Sharpe on the
posterior would not have this property (the tangency point is not the prior),
which is why the utility form was chosen.

## 3. The prior is the file

There is no market-capitalization prior in pofo: a managed-futures fund, a
stacked 90/60, an inflation-linked sleeve or gold have no capitalization
weight, so the canonical anchor is undefined for exactly the books this
toolkit is built for. The prior is the weights written in the portfolio
file, i.e. the allocation the owner already defends. Reverse optimization
then answers a question the owner rarely asks: "what return does my
allocation implicitly expect from each line?" (a 15 % gold line at the
sample covariance of the book above implies 5.6 % a year; the owner's
stated view was 2 %, and that gap is the whole disagreement).

The caller fills `Spec.Prior` (fractions, in asset order) exactly as it
fills `Lower`/`Upper` through `Resolve`: identifier resolution stays in the
caller. `Solve` refuses `BlackLitterman` without a prior of the right length
summing to 1 (tolerance 1e-6).

## 4. Risk aversion: `prior-return`, not He-Litterman's 2.5

`λ` sets the scale of `π`, and absolute views are read against that scale,
so it matters. Two ways to fix it, one token:

- `prior-return:R` (percent per year, arithmetic, at this toolkit's zero
  risk-free rate): the return the owner expects from the prior allocation
  as a whole. Then `λ = R / (wᵀ Σ w)`.
- Default, when the token is absent: an assumed Sharpe of 0.4 for the prior,
  `λ = 0.4 / sqrt(wᵀ Σ w)`.

He and Litterman's `δ = 2.5` is NOT the default. It was calibrated on an
equity market at 15-20 % volatility; on an 8 %-volatility defensive book it
implies a 1.6 %/yr expected return, every ordinary view then reads as a
large surprise, and the posterior swings to corners (measured: linkers 0 %,
small value 35 % under the same views that land inside the owner's ranges
under `prior-return:4.6`). The note printed by the report names the λ used
and where it came from.

## 5. Grammar (`#meta optimize:`)

```
#meta optimize:black-litterman
#meta optimize:black-litterman,view:IGLN:2,view:DBMFE:5@60,prior-return:4.6
#meta optimize:black-litterman,view:DBMFE>DTLA:3@70,bounds:IGLN:10-20,max-vol:9
```

- `view:ID:Q[@C]`: absolute view, ID earns Q percent a year (arithmetic,
  nominal, in the run's currency), with confidence C percent in (0,100),
  default 50.
- `view:ID>ID2:Q[@C]`: relative view, ID beats ID2 by Q points a year.
- `prior-return:R`: see section 4.
- Everything the other objectives accept composes: `max-weight`,
  `min-weight`, `bounds`, `max-vol`, `min-return`, `max-drawdown`, `train`.
- `view:` and `prior-return:` on any other objective are refused at parse
  time, like limits on `risk-parity`, so no token is dropped in silence.
- A view naming an identifier the file does not hold fails in `Resolve`,
  where the holdings are known. Identifiers match as `bounds` do (the id
  written in the file, or the resolved symbol).
- Q may be negative or zero (a zero-real-return view on gold in a 2 %
  inflation world is a positive nominal number, but a negative view is a
  legitimate statement).

## 6. API (`pkg/optimize`, new file `bl.go`)

```go
const BlackLitterman Objective = "black-litterman"

// View is one belief about expected returns: absolute when Versus is
// empty, relative ("Asset beats Versus by Return") otherwise.
type View struct {
    Asset, Versus string  // identifiers as written in the file
    Return        float64 // fraction per year
    Confidence    float64 // in (0,1); 0.5 = He and Litterman's default
    // resolved positions, filled by Spec.Resolve; -1 when unset
    asset, versus int
}

// Spec gains:
Views       []View
Prior       []float64 // fractions in asset order, filled by the caller
PriorReturn float64   // fraction per year; 0 = the Sharpe-0.4 default

// ImpliedReturns is reverse optimization: the annualized expected returns
// under which prior is the optimal portfolio at risk aversion lambda.
func ImpliedReturns(prior []float64, cov [][]float64, lambda float64) []float64

// Posterior blends the implied returns with the views (rows already
// resolved to positions) and returns the Black-Litterman expected returns.
func Posterior(implied []float64, cov [][]float64, views []View) ([]float64, error)

// Result gains:
Implied   []float64 // π, annualized, per asset (BlackLitterman)
Posterior []float64 // μ, annualized, per asset (BlackLitterman)
Lambda    float64   // the risk aversion used (BlackLitterman)
```

`Spec.Resolve(ids)` resolves views alongside bounds. `Spec.Bounded()` is
unchanged. `ParseSpec` grows the two tokens and the refusals of section 5.

Solve path: `Solve` computes `mu, cov`, then for `BlackLitterman` derives
`λ`, `π`, `μ_BL`. Without bounds or limits the utility is a convex QP over
the simplex, solved by `minimizeSimplex` from the prior as the starting
point (global optimum, and the no-view identity holds to solver tolerance).
With bounds or limits it goes through `solveConstrained`, whose
`objectiveFn` gets a `BlackLitterman` case returning the utility at `μ_BL`;
the prior joins the starting points. `Result.Return`, `Volatility` and
`Sharpe` are computed at `μ_BL` (the EXPECTED figures, which is the point),
and `stats`' doc says so; `CAGR`/`Feasible` keep their path meaning.

## 7. Where it surfaces

- `pkg/compare/optimize.go`: fills `sp.Prior` from `base.Assets[i].Weight`,
  resolves views with the bounds, and extends the note: the λ and its
  origin, then one entry per line "ID: implied X % -> posterior Y %" (only
  lines a view touched change, the others are listed once as "unchanged"),
  then the views as parsed. `objectiveLabel` names the objective.
- Every enumeration of the objectives is extended: `README.md`, `AGENTS.md`,
  root `doc.go`, `cmd/pofo/main.go` (flag help), `pkg/metrics/doc.go`,
  `pkg/metrics/example_test.go` if it lists them, `pkg/optimize/series.go`,
  `pkg/optimize/doc.go` ("Ten objectives"), and the `CLAUDE.md` map line for
  `pkg/optimize`. `docs/README.md` indexes this document;
  `docs/weight-search-design.md` gets a pointer paragraph.
- `example_test.go`: one runnable example on a synthetic three-asset
  problem showing the no-view identity and one view moving one weight.

## 8. Validation

1. External reference (the golden): He and Litterman (1999), "The Intuition
   Behind Black-Litterman Model Portfolios", Goldman Sachs Investment
   Management Research (also Idzorek 2005, "A Step-by-Step Guide to the
   Black-Litterman Model", which reworks the same seven-country example with
   `Ω = diag(P τ Σ Pᵀ)`, i.e. c = 0.5 here). The published inputs are the
   seven-country correlation matrix, volatilities, market weights, δ = 2.5
   and τ = 0.05; the published outputs are π and the posterior means and
   weights under the stated views. The test transcribes the inputs and
   asserts the outputs to the paper's printed precision. The numbers must
   come from a FETCHED copy of the paper, cited by URL and table number in
   the test; if no copy is reachable, the test is written against the
   model's own closed-form identities only and says so in a comment, and no
   figure is typed from memory.
   The golden runs where `make golden` runs (`pkg/datasets/golden/`, see the
   Makefile), with the same fixture discipline as the other goldens.

   As shipped, both papers were fetched and both are transcribed. He and
   Litterman's Tables 1 and 2 (and its footnote 3, `δ = 2.5`) pin
   `ImpliedReturns` on the seven countries; Idzorek's Table 5, Table 2 and
   Table 1 pin it on the eight asset classes, his Tables 3a and 3b on the two
   mini-portfolios of his third view, his Table 6 pins `Posterior` at
   confidence 0.5, and his Table 6 weight vector is checked through the
   reverse-optimization identity it was produced by (those weights sum to
   103.63 %, so the long-only `Solve` neither does nor should reproduce them).
   ONE THING THE GRAMMAR CANNOT SAY, found while writing the golden: both
   papers state their headline view over SEVERAL lines per side, weighted by
   market capitalization (Idzorek's `P` row 3 is `.9 -.9 .1 -.1`, He and
   Litterman's is `1 -0.295 -0.705`), while `view:ID>ID2` compares one line
   with one line. Splitting such a view into two independent ones is a
   DIFFERENT model, and measurably so (6.18 % against the published 6.50 %
   for US Large Growth), so the golden restates it the way the papers
   themselves describe it, as a view between the two mini-portfolios: two
   extra lines are appended to the covariance, each a fixed blend of the
   originals, and the view is stated between them. A blend's covariance with
   anything is that blend of its legs' covariances, so the view row, and the
   posterior over the original lines, are exactly the papers'. A multi-leg
   view is therefore expressible in the MODEL and not in the file grammar;
   whether the grammar should grow one is left open.
2. Identities, in `bl_test.go`: no view returns the prior (unbounded and
   inside satisfied bounds); an absolute view above π raises that asset's
   posterior and weight, below lowers them; a relative view moves the pair
   in opposite directions and leaves the rest's SUM of implied returns
   consistent; posterior is monotone in confidence, equals π at c → 0 and
   satisfies `P μ = Q` at c → 1; `prior-return` scales π linearly; `λ`
   default equals 0.4/σ_p.
3. `ParseSpec`: both tokens, defaults, refusals on other objectives,
   malformed forms (`view:X`, `view:X:abc`, `@0`, `@100`, `>` with a missing
   side).
4. `Resolve`: unknown identifier in a view fails with the identifier named.
5. `pkg/compare`: the note carries λ, the implied/posterior pairs and the
   views; `make check` green; `make golden` green.

## 9. Traps

- The zero risk-free rate of this toolkit reaches BL too: a 3 %/yr view on
  a 3 %-volatility sleeve is a Sharpe of 1.0 in the utility's eyes, and the
  posterior weight follows. State views as excess returns over cash when
  that is what is meant, and say so in the note when a cash-like line
  carries a view.
- What BL prices is the MEAN. A line held for its crisis covariance (gold,
  an unhedged dollar duration leg) or for a liability (linkers) has a reason
  the model cannot read; BL will size it by its view alone. That is a
  feature (it shows what the hedge costs in expected return) as long as the
  reader knows it, so the note says "views price expected returns only".
- `Spec.Train` still slices the returns in the caller; the prior is the
  file's weights regardless of the window.
- CWARP and RiskParity refusals stay exactly as they are; BlackLitterman
  accepts every constraint they refuse.

## References

- Black, F. and Litterman, R. (1992), "Global Portfolio Optimization",
  Financial Analysts Journal 48(5).
- He, G. and Litterman, R. (1999), "The Intuition Behind Black-Litterman
  Model Portfolios", Goldman Sachs Investment Management Research.
- Idzorek, T. (2005), "A Step-by-Step Guide to the Black-Litterman Model".
- Walters, J. (2014), "The Black-Litterman Model in Detail", SSRN 1314585.
- Meucci, A. (2010), "The Black-Litterman Approach: Original Model and
  Extensions", SSRN 1117574.
