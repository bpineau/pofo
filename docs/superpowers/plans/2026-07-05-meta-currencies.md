# `#meta currencies` Multi-Currency Comparison Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a `#meta currencies:USD,EUR` directive that runs one portfolio once per listed currency and shows the results as adjacent columns (nominal + per-currency real statistics), exposing currency risk.

**Architecture:** `pkg/portfolio` parses the directive into `Spec.Currencies`. `cmd/pofo` expands each spec into one `result` per currency, fetching assets, benchmark and CPI deflator per currency; the existing report/CLI rendering handles the extra columns unchanged.

**Tech Stack:** Go stdlib only. Tests use the existing fakes (no network). Reference: `docs/superpowers/specs/2026-07-05-meta-currencies-design.md`.

## Global Constraints

- Stdlib only; no third-party dependency.
- English for all code, godoc and docs. Never write an em-dash.
- `make check` must pass; `make golden` must stay green (no change to single-currency calculations).
- Every package keeps `doc.go` and runnable `example_test.go` examples current.
- Units unchanged: `Holding.Fees`/`EnvelopeFees` in percent/year; weights fractions summing to 1.

---

### Task 1: Parse `#meta currencies` into `Spec.Currencies`

**Files:**
- Modify: `pkg/portfolio/parse.go` (add field to `Spec` struct ~line 68-72; add `case` in `applyMeta` ~line 339-346; add guard near line 165-167)
- Test: `pkg/portfolio/parse_test.go`
- Test: `pkg/portfolio/example_test.go`

**Interfaces:**
- Produces: `Spec.Currencies []string` — uppercased 3-letter ISO codes in written order, deduplicated; nil when the directive is absent.

- [ ] **Step 1: Write the failing tests**

Add to `pkg/portfolio/parse_test.go`:

```go
func TestParseCurrencies(t *testing.T) {
	spec, err := Parse("p", strings.NewReader("#meta currencies:usd,EUR,usd\n100 VOO\n"))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got, want := spec.Currencies, []string{"USD", "EUR"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Currencies = %v, want %v", got, want)
	}
	if spec.Meta["currencies"] != "usd,EUR,usd" {
		t.Errorf("Meta[currencies] = %q, want verbatim", spec.Meta["currencies"])
	}
}

func TestParseCurrenciesInvalid(t *testing.T) {
	for _, in := range []string{
		"#meta currencies:US\n100 VOO\n",     // too short
		"#meta currencies:USD,\n100 VOO\n",   // empty token
		"#meta currencies:US1\n100 VOO\n",    // not alpha
	} {
		if _, err := Parse("p", strings.NewReader(in)); err == nil {
			t.Errorf("Parse(%q): expected error", in)
		}
	}
}

func TestParseCurrenciesOptimizeConflict(t *testing.T) {
	in := "#meta currencies:USD,EUR\n#meta optimize:max-sharpe\n100 VOO\n"
	if _, err := Parse("p", strings.NewReader(in)); err == nil {
		t.Error("Parse: expected error combining currencies and optimize")
	}
}
```

Ensure `reflect` is imported in the test file (add to the import block if missing).

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./pkg/portfolio/ -run TestParseCurrencies -v`
Expected: FAIL (compile error: `spec.Currencies` undefined).

- [ ] **Step 3: Add the `Currencies` field to `Spec`**

In `pkg/portfolio/parse.go`, inside the `Spec` struct, right before the `Meta` field (~line 70):

```go
	// Currencies lists the base currencies in which the portfolio should
	// be evaluated ("#meta currencies:USD,EUR"): the CLI produces one
	// comparison column per currency, each with its own numeraire and CPI
	// deflator, so the difference between columns exposes currency risk.
	// Uppercased 3-letter codes in written order, deduplicated; nil when
	// the directive is absent (the caller then uses its single default
	// currency). Cannot be combined with Optimize.
	Currencies []string
```

- [ ] **Step 4: Parse the directive in `applyMeta`**

In `pkg/portfolio/parse.go`, add a `case` in the `switch key` of `applyMeta` (alongside the other cases, before `default:`):

```go
		case "currencies":
			cs, err := parseCurrencies(val)
			if err != nil {
				return fmt.Errorf("#meta currencies: %v", err)
			}
			s.Currencies = cs
```

Add the helper near `parseNumber` at the bottom of the file:

```go
// parseCurrencies parses a comma-separated list of 3-letter currency codes
// ("USD,EUR"), uppercased and deduplicated while preserving written order.
func parseCurrencies(val string) ([]string, error) {
	seen := map[string]bool{}
	var out []string
	for _, tok := range strings.Split(val, ",") {
		code := strings.ToUpper(strings.TrimSpace(tok))
		if len(code) != 3 || !isAlpha(code) {
			return nil, fmt.Errorf("invalid code %q (expected a 3-letter currency, e.g. USD)", tok)
		}
		if !seen[code] {
			seen[code] = true
			out = append(out, code)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no currency listed")
	}
	return out, nil
}

// isAlpha reports whether s is all ASCII letters.
func isAlpha(s string) bool {
	for _, r := range s {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return s != ""
}
```

- [ ] **Step 5: Add the optimize conflict guard**

In `pkg/portfolio/parse.go`, right after the existing optimize/leverage guard (~line 167):

```go
	if spec.Optimize != nil && len(spec.Currencies) > 0 {
		return nil, fmt.Errorf("#meta optimize and #meta currencies cannot be combined")
	}
```

- [ ] **Step 6: Run tests to verify they pass**

Run: `go test ./pkg/portfolio/ -run TestParseCurrencies -v`
Expected: PASS (all three).

- [ ] **Step 7: Add a runnable example**

Append to `pkg/portfolio/example_test.go`:

```go
func ExampleParse_currencies() {
	spec, _ := portfolio.Parse("dragon", strings.NewReader(
		"#meta currencies:USD,EUR\n60 NTSGSIM\n40 XAUUSDSIM\n"))
	fmt.Println(spec.Currencies)
	// Output: [USD EUR]
}
```

Confirm `strings`, `fmt` and the `portfolio` import are already present in that file; add them if not.

- [ ] **Step 8: Run the package tests and examples**

Run: `go test ./pkg/portfolio/`
Expected: PASS (ok).

- [ ] **Step 9: Commit**

```bash
git add pkg/portfolio/parse.go pkg/portfolio/parse_test.go pkg/portfolio/example_test.go
git commit -m "portfolio: parse #meta currencies directive

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 2: Expand portfolios per currency in the CLI

**Files:**
- Modify: `cmd/pofo/main.go` — `result` struct (~line 82-98), fetch loop (~line 279-291), benchmark fetch (~line 294-303), simulate loop (~line 330-383), deflator block (~line 406-424), footnote (~line 1470-1473), `fetchAsset` (~line 1148-1157)
- Test: `cmd/pofo/currencies_test.go` (new)

**Interfaces:**
- Consumes: `Spec.Currencies` from Task 1.
- Produces: `effectiveCurrencies(spec *portfolio.Spec, def string) []string` — the currency list a spec expands into (`spec.Currencies` if set, else `[]string{def}`); `fetchAssetIn(ctx, c, id, opt, currency)` — `fetchAsset` with an explicit currency; `result.currency` field.

- [ ] **Step 1: Write the failing test for `effectiveCurrencies`**

Create `cmd/pofo/currencies_test.go`:

```go
package main

import (
	"reflect"
	"testing"

	"github.com/bpineau/pofo/pkg/portfolio"
)

func TestEffectiveCurrencies(t *testing.T) {
	cases := []struct {
		name string
		spec *portfolio.Spec
		def  string
		want []string
	}{
		{"default", &portfolio.Spec{}, "EUR", []string{"EUR"}},
		{"declared", &portfolio.Spec{Currencies: []string{"USD", "EUR"}}, "EUR", []string{"USD", "EUR"}},
	}
	for _, c := range cases {
		if got := effectiveCurrencies(c.spec, c.def); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: effectiveCurrencies = %v, want %v", c.name, got, c.want)
		}
	}
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `go test ./cmd/pofo/ -run TestEffectiveCurrencies -v`
Expected: FAIL (compile error: `effectiveCurrencies` undefined).

- [ ] **Step 3: Add `effectiveCurrencies` and `fetchAssetIn`**

In `cmd/pofo/main.go`, replace the `fetchAsset` function (~line 1146-1157) with:

```go
// fetchAsset runs the full library pipeline (SIM extension, currency
// conversion, window) for one asset, in the CLI's base currency.
func fetchAsset(ctx context.Context, c *marketdata.Client, id string, opt *options) (*marketdata.Series, error) {
	return fetchAssetIn(ctx, c, id, opt, opt.currency)
}

// fetchAssetIn is fetchAsset with an explicit target currency, used when a
// portfolio is evaluated in several currencies ("#meta currencies").
func fetchAssetIn(ctx context.Context, c *marketdata.Client, id string, opt *options, currency string) (*marketdata.Series, error) {
	return c.FetchExtended(ctx, id, marketdata.FetchOptions{
		From:     opt.start,
		To:       opt.end,
		NoSim:    opt.noSim,
		Simdata:  opt.simdata,
		Currency: currency,
	})
}

// effectiveCurrencies is the list of base currencies a spec expands into:
// its "#meta currencies" list when set, otherwise the single CLI default.
func effectiveCurrencies(spec *portfolio.Spec, def string) []string {
	if len(spec.Currencies) > 0 {
		return spec.Currencies
	}
	return []string{def}
}
```

- [ ] **Step 4: Run to verify it passes**

Run: `go test ./cmd/pofo/ -run TestEffectiveCurrencies -v`
Expected: PASS.

- [ ] **Step 5: Add the `currency` field to `result`**

In `cmd/pofo/main.go`, in the `result` struct (~line 82), after `rebalanceDays int`:

```go
	currency      string // base currency this column was evaluated in
```

- [ ] **Step 6: Fetch assets per currency**

Replace the fetch loop (~line 279-291, the `seriesByID := map[...]` block) with:

```go
	// Download every distinct (currency, asset) once. A "#meta currencies"
	// directive evaluates the same portfolio in several currencies.
	seriesByCur := map[string]map[string]*marketdata.Series{}
	for _, spec := range specs {
		for _, cur := range effectiveCurrencies(spec, opt.currency) {
			m := seriesByCur[cur]
			if m == nil {
				m = map[string]*marketdata.Series{}
				seriesByCur[cur] = m
			}
			for _, h := range spec.Holdings {
				if _, ok := m[h.ID]; ok {
					continue
				}
				s, err := fetchAssetIn(ctx, client, h.ID, &opt, cur)
				if err != nil {
					return fmt.Errorf("portfolio %s, asset %q (%s): %w", spec.Name, h.ID, cur, err)
				}
				m[h.ID] = s
			}
		}
	}
```

- [ ] **Step 7: Fetch the benchmark and deflator per currency (memoized)**

Replace the benchmark block (~line 294-303, `var bench *marketdata.Series ...`) with:

```go
	// Benchmark for Beta/CWARP, best effort, memoized per currency. The
	// chart's reference curve uses the default currency (benchIn(opt.currency)).
	benchCache := map[string]*marketdata.Series{}
	benchIn := func(cur string) *marketdata.Series {
		if opt.benchmark == "" {
			return nil
		}
		if b, ok := benchCache[cur]; ok {
			return b
		}
		b, err := client.FetchExtended(ctx, opt.benchmark, marketdata.FetchOptions{
			From: opt.start, NoSim: true, Currency: cur,
		})
		if err != nil {
			log.Printf("warning: benchmark %s unavailable in %s (no Beta): %v", opt.benchmark, cur, err)
			b = nil
		}
		benchCache[cur] = b
		return b
	}
	bench := benchIn(opt.currency)
```

- [ ] **Step 8: Set the currency on each result and wrap the simulate loop**

In `simulateInto` (~line 330), add a `currency` parameter and store it. Change the signature and the `results = append(...)` line:

```go
	simulateInto := func(p *portfolio.Portfolio, spec *portfolio.Spec, currency string) error {
```

and in its body set the field in the appended result literal:

```go
		results = append(results, &result{p: p, sim: sim, color: chart.PaletteColor(len(results)), rebalanceDays: days, currency: currency})
```

Then wrap the per-spec body (~line 351-383) in a currency loop and thread `cur`. Replace the whole `for _, spec := range specs { ... }` build loop with:

```go
	for _, spec := range specs {
		for _, cur := range effectiveCurrencies(spec, opt.currency) {
			p, err := portfolio.Build(spec, portfolio.BuildOptions{
				Fetch:        func(id string) (*marketdata.Series, error) { return seriesByCur[cur][id], nil },
				Fees:         feesFor,
				Cash:         cashRate,
				BorrowSpread: 1.0, // default: cash + 1 %/yr
				BaseCurrency: cur,
			})
			if err != nil {
				return err
			}
			// Multi-currency: tag each column with its currency.
			if len(spec.Currencies) > 0 {
				p.Name = fmt.Sprintf("%s (%s)", p.Name, cur)
			}
			// An optimized portfolio is shown next to its written weights, so
			// the optimizer's choice can be compared with the baseline.
			// (Optimize and currencies cannot be combined, so cur is unique here.)
			if spec.Optimize != nil {
				pOpt, note, err := optimizedPortfolio(p, spec, benchIn(cur))
				if err != nil {
					return fmt.Errorf("portfolio %s: %w", spec.Name, err)
				}
				p.Name = spec.Name + " (as written)"
				if err := simulateInto(p, spec, cur); err != nil {
					return err
				}
				if err := simulateInto(pOpt, spec, cur); err != nil {
					return err
				}
				results[len(results)-1].note = note
				continue
			}
			if err := simulateInto(p, spec, cur); err != nil {
				return err
			}
		}
	}
```

- [ ] **Step 9: Deflate each result with its own currency's CPI**

Replace the deflator block (~line 406-424). First replace the single-deflator declaration:

```go
	// Consumer-price index per currency, memoized, to report drawdowns/TTR
	// and real stats in purchasing-power terms. Best-effort: a currency
	// without a wired CPI simply has no real columns.
	deflatorCache := map[string]*marketdata.Series{}
	deflatorIn := func(cur string) (*marketdata.Series, bool) {
		if s, ok := deflatorCache[cur]; ok {
			return s, s != nil
		}
		s, ok := inflationSeries(ctx, client, cur, commonStart)
		if !ok {
			s = nil
		}
		deflatorCache[cur] = s
		return s, s != nil
	}
```

Then, in the stats loop, replace the `if hasDeflator { ... }` block with:

```go
		if d, ok := deflatorIn(r.currency); ok {
			if rs, err := metrics.Compute(r.winDates, deflate(r.winDates, r.winValues, d)); err == nil {
				r.realStats, r.hasReal = rs, true
			}
		}
```

And replace the `if bench != nil { ... }` block in the stats loop with a per-currency benchmark:

```go
		if b := benchIn(r.currency); b != nil {
			bd, bv := seriesSlices(b)
			if rel, ok := metrics.VsBenchmark(r.winDates, r.winValues, bd, bv); ok {
				st.Beta, st.HasBeta = rel.Beta, true
				r.rel, r.hasRel = rel, true
			}
			if c, ok := metrics.CWARPvs(r.winDates, r.winValues, bd, bv, metrics.CWARPParams{}); ok {
				st.CWARP, st.HasCWARP = c, true
			}
		}
```

Then delete the now-unused `benchDates`/`benchValues` block (~line 401-404):

```go
	var benchDates []time.Time
	var benchValues []float64
	if bench != nil {
		benchDates, benchValues = seriesSlices(bench)
	}
```

Those two were used only by the stats-loop block just rewritten. The chart's
benchmark curve is drawn inside `buildPage`, which recomputes its own slices
from `bench` (~line 1320-1323), so `bench` itself must stay. Removing the
declaration avoids a "declared and not used" compile error.

- [ ] **Step 10: Name the currencies in the FX footnote**

Replace the footnote block (~line 1470-1473) with one that lists the currencies actually used:

```go
	curSet := map[string]bool{}
	var curs []string
	for _, r := range results {
		if r.currency != "" && !curSet[r.currency] {
			curSet[r.currency] = true
			curs = append(curs, r.currency)
		}
	}
	if len(curs) > 0 {
		page.Footnotes = append(page.Footnotes, fmt.Sprintf(
			"Series converted to %s (daily Yahoo FX crosses; the earliest known rate is held constant before the FX history starts). Columns tagged with a currency show the same portfolio through that currency's numeraire and CPI.", strings.Join(curs, ", ")))
	}
```

- [ ] **Step 11: Build and run `make check`**

Run: `make check`
Expected: PASS (fmt-check + lint + test). Fix any `hasDeflator`/`deflator` leftover references the compiler reports (both were removed).

- [ ] **Step 12: Integration check against the real example**

Run: `make build && ./pofo -cli -currency EUR examples/claude-dragonlite.txt 2>/dev/null | head -40`
Expected: a single dragonlite column (directive not yet in the file).

Then temporarily verify expansion with an inline file:

```bash
printf '#meta currencies:USD,EUR\n60 NTSGSIM\n25 DBMFSIM\n15 XAUUSDSIM\n' > /tmp/dl-cur.txt
./pofo -cli /tmp/dl-cur.txt 2>/dev/null | head -40
```

Expected: two columns, `dl-cur (USD)` and `dl-cur (EUR)`, with visibly different CAGR/vol (FX effect) and a real-drawdown column populated for both.

- [ ] **Step 13: Confirm goldens unchanged**

Run: `make golden`
Expected: PASS (single-currency paths are byte-for-byte unchanged).

- [ ] **Step 14: Commit**

```bash
git add cmd/pofo/main.go cmd/pofo/currencies_test.go
git commit -m "pofo: expand portfolios per #meta currencies into comparison columns

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 3: Documentation and motivating example

**Files:**
- Modify: `README.md` (the `#meta` directive table)
- Modify: `examples/claude-dragonlite.txt` (add a commented directive)
- Modify: `CLAUDE.md` if it enumerates `#meta` directives (verify first)

**Interfaces:** none (docs only).

- [ ] **Step 1: Locate the `#meta` documentation**

Run: `grep -n "#meta" README.md | head`
Expected: a list/table of directives (rebalance, capital, contribute, withdraw, leverage, borrow-spread, optimize, extra-fees).

- [ ] **Step 2: Add the `currencies` row to the README**

Add an entry next to the other directives, matching the surrounding format, for example:

```markdown
- `#meta currencies:USD,EUR` — evaluate the portfolio in several base
  currencies at once; each becomes a comparison column with its own numeraire
  and CPI deflator (nominal and real stats), so the gap between columns shows
  the currency risk. Cannot be combined with `#meta optimize`.
```

- [ ] **Step 3: Add the directive to the dragonlite example**

Add near the top of `examples/claude-dragonlite.txt`, after the header comment block, a commented line the user can enable:

```
# Décommenter pour comparer le vécu d'un investisseur USD et EUR (risque de change) :
# #meta currencies:USD,EUR
```

- [ ] **Step 4: Update CLAUDE.md if needed**

Run: `grep -n "leverage:on\|optimize:\|extra-fees\|#meta" CLAUDE.md`
If a directive list exists, add `currencies:USD,EUR` alongside it; otherwise no change.

- [ ] **Step 5: Verify docs build/tests still pass**

Run: `make check`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add README.md examples/claude-dragonlite.txt CLAUDE.md
git commit -m "docs: document #meta currencies and add dragonlite example

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

## Self-Review

**Spec coverage:**
- Parsing (`Spec.Currencies`, validation, verbatim in Meta) → Task 1.
- Optimize+currencies guard → Task 1 Step 5.
- Per-currency fetch, simulate, labels → Task 2 Steps 6, 8.
- Per-currency deflator (real stats) and benchmark → Task 2 Steps 7, 9.
- Common window unchanged, extra columns flow through rendering → Task 2 (loop untouched at ~385-401; render untouched).
- FX footnote names currencies → Task 2 Step 10.
- Currency without CPI degrades gracefully → `deflatorIn` returns `false`, real columns omitted (Task 2 Step 9); no explicit test needed (existing `inflationSeries` gate covered by `deflate_test.go`).
- Docs + example → Task 3.

**Placeholder scan:** none — every code step shows full code.

**Type consistency:** `effectiveCurrencies(spec, def) []string`, `fetchAssetIn(ctx, c, id, opt, currency)`, `result.currency string`, `benchIn(cur) *marketdata.Series`, `deflatorIn(cur) (*marketdata.Series, bool)`, `simulateInto(p, spec, currency)` are used consistently across Tasks 1-2.

**Note on line numbers:** all `~line N` references are from the file state at planning time; the implementer should anchor on the quoted surrounding code, not the exact number.
