package portfolio

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"github.com/bpineau/pofo/pkg/optimize"
)

// Holding is one allocation line of a portfolio file.
type Holding struct {
	Weight    float64 // normalized fraction of the portfolio (sums to 1)
	RawWeight float64 // weight as written in the file, in percent
	ID        string  // ticker or ISIN, verbatim
	Fees      float64 // declared TER in percent per year; negative when absent
}

// Spec is a parsed portfolio description.
type Spec struct {
	Name     string
	Holdings []Holding
	Warnings []string

	// RebalanceDays is the per-portfolio rebalancing period set by a
	// "#meta rebalance:N" directive; negative when the file does not
	// specify one (callers then apply their default).
	RebalanceDays int

	// Capital is the starting amount in base currency ("#meta
	// capital:10000"); negative when absent (simulations then run on a
	// relative base-100 index). Required when Contribute or Withdraw is
	// set, so absolute flows have a meaningful scale.
	Capital float64

	// Contribute and Withdraw are periodic external flows
	// ("#meta contribute:500/month", "#meta withdraw:4%/year").
	Contribute Flow
	Withdraw   Flow

	// Leverage, set by "#meta leverage:on", keeps the weights as written
	// instead of normalizing them: a sum above 100 % is financed by a
	// negative cash position, below 100 % the residual sits in cash.
	Leverage bool

	// Sim, set by "#meta sim:on", asks Build to fetch every holding through
	// its SIM (backcast-extended) variant, exactly as if each identifier
	// carried the "SIM" suffix, so the file need not spell it out (and
	// de-suffixing to share the file stays a one-liner). Identifiers already
	// ending in "SIM" are left untouched (never double-suffixed), and a
	// holding without any backcast falls back to its real quotes: the
	// directive means "use the simulated history IF one exists". See Build.
	Sim bool

	// BorrowSpread is the yearly spread over the cash rate paid on
	// borrowed money ("#meta borrow-spread:X", percent per year);
	// negative when absent (callers apply their default).
	BorrowSpread float64

	// EnvelopeFees is the additional yearly fee of the hosting envelope
	// (life insurance, PER, managed account…) set by "#meta extra-fees:X", fees
	// applied on top of the WHOLE portfolio, in addition to the assets'
	// own TERs, in percent per year;
	// negative when absent. Unlike asset TERs (already reflected in
	// prices), it must be deducted from the simulated performance.
	EnvelopeFees float64

	// Optimize, when non-nil, asks an optimizer to compute the weights
	// instead of using those written in the file
	// ("#meta optimize:max-sharpe[,max-weight:40]"). The written weights
	// then serve only as a baseline for comparison, except under
	// black-litterman, where they are also the prior the views tilt away
	// from (the caller fills optimize.Spec.Prior with them).
	Optimize *optimize.Spec

	// Currencies lists the base currencies in which the portfolio should
	// be evaluated ("#meta currencies:USD,EUR"): the CLI produces one
	// comparison column per currency, each with its own numeraire and CPI
	// deflator, so the difference between columns exposes currency risk.
	// Uppercased 3-letter codes in written order, deduplicated; nil when
	// the directive is absent (the caller then uses its single default
	// currency). Cannot be combined with Optimize.
	Currencies []string

	// Meta holds every "#meta key:value" directive verbatim, for callers
	// with custom needs.
	Meta map[string]string
}

// ParseFile reads a portfolio description file. The portfolio name is the
// file name without its extension.
func ParseFile(path string) (*Spec, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	base := filepath.Base(path)
	spec, err := Parse(strings.TrimSuffix(base, filepath.Ext(base)), f)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return spec, nil
}

// Parse reads a portfolio description: one line per asset, formatted as
//
//	<weight in %> <ticker, ISIN or alias> [TER in %/year]
//
// Everything after a "#" is a comment; nothing else may follow the optional
// fee column. Blank lines and lines starting with // are ignored. Lines
// starting with "#meta" carry per-portfolio directives as key:value pairs;
// currently "rebalance:N" (days between rebalancings, 0 to never rebalance).
// Weights accept a decimal comma and an optional % suffix. If the weights do
// not sum to 100 they are normalized and a warning is recorded.
func Parse(name string, r io.Reader) (*Spec, error) {
	spec := newSpec(name)
	sc := bufio.NewScanner(r)
	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := strings.TrimSpace(sc.Text())
		if isMeta, rest := metaDirective(line); isMeta {
			if err := spec.applyMeta(rest); err != nil {
				return nil, fmt.Errorf("line %d: %w", lineNo, err)
			}
			continue
		}
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return nil, fmt.Errorf("line %d: expected \"<weight> <ticker/ISIN> [TER]\", got %q", lineNo, line)
		}
		w, err := parseWeight(fields[0])
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid weight %q: %v", lineNo, fields[0], err)
		}
		h := Holding{RawWeight: w, ID: fields[1], Fees: -1}
		rest := fields[2:]
		// Optional third column: the asset's TER in percent/year. Anything
		// else after the ticker must be a "#" comment, already stripped above.
		if len(rest) > 0 {
			fees, ferr := parseNumber(rest[0])
			if ferr != nil {
				return nil, fmt.Errorf("line %d: unexpected %q after the ticker; write a TER (number) or move free text behind a \"#\" comment", lineNo, strings.Join(rest, " "))
			}
			if !validFees(fees) {
				return nil, fmt.Errorf("line %d: fees %q out of range (0-20 %%/year)", lineNo, rest[0])
			}
			h.Fees = fees
			rest = rest[1:]
		}
		if len(rest) > 0 {
			return nil, fmt.Errorf("line %d: unexpected %q after the TER; move free text behind a \"#\" comment", lineNo, strings.Join(rest, " "))
		}
		spec.Holdings = append(spec.Holdings, h)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if err := spec.normalize(); err != nil {
		return nil, err
	}
	return spec, nil
}

// newSpec is an empty Spec with every numeric directive unset (negative), the
// state Parse starts from and NewSpec leaves in place.
func newSpec(name string) *Spec {
	return &Spec{Name: name, RebalanceDays: -1, EnvelopeFees: -1, BorrowSpread: -1, Capital: -1}
}

// normalize checks a Spec whose Holdings carry their RawWeight and whose
// directives are applied, then sets every Holding.Weight: RawWeight/100 under
// Leverage, RawWeight over the sum of them otherwise (with a warning when that
// sum is off 100). It is the one validation Parse and NewSpec share, so a spec
// built in code and the same lines in a file are the same spec.
func (spec *Spec) normalize() error {
	if len(spec.Holdings) == 0 {
		return fmt.Errorf("no allocation line found")
	}
	sum := 0.0
	for _, h := range spec.Holdings {
		sum += h.RawWeight
	}
	if sum <= 0 {
		return fmt.Errorf("weights sum to zero")
	}
	if (spec.Contribute.Active() || spec.Withdraw.Active()) && spec.Capital <= 0 {
		return fmt.Errorf("#meta contribute/withdraw need a starting amount: add \"#meta capital:<amount>\"")
	}
	if spec.Optimize != nil && spec.Leverage {
		return fmt.Errorf("#meta optimize and #meta leverage cannot be combined")
	}
	if spec.Optimize != nil && len(spec.Currencies) > 0 {
		return fmt.Errorf("#meta optimize and #meta currencies cannot be combined")
	}
	if spec.Leverage {
		// Explicit leverage: weights are fractions of the capital, as
		// written; the residual (100−sum) becomes a cash position.
		if sum > maxWeightPct {
			return fmt.Errorf("total exposure %.4g %% exceeds the 500 %% cap", sum)
		}
		for i := range spec.Holdings {
			spec.Holdings[i].Weight = spec.Holdings[i].RawWeight / 100
		}
		if math.Abs(sum-100) > 0.5 {
			spec.Warnings = append(spec.Warnings,
				fmt.Sprintf("explicit leverage: total exposure %.4g %%, cash residual %.4g %%", sum, 100-sum))
		}
		return nil
	}
	for _, h := range spec.Holdings {
		if h.RawWeight > 100 {
			return fmt.Errorf("weight %.4g %% > 100 %%; add \"#meta leverage:on\" if the exposure is intentional", h.RawWeight)
		}
	}
	if math.Abs(sum-100) > 0.5 {
		spec.Warnings = append(spec.Warnings,
			fmt.Sprintf("weights sum to %.4g %% instead of 100 %%, they were normalized", sum))
	}
	for i := range spec.Holdings {
		spec.Holdings[i].Weight = spec.Holdings[i].RawWeight / sum
	}
	return nil
}

// The bounds every allocation line is held to, in percent: a weight above
// zero and at most 500 % (the leverage cap), a TER from 0 to 20 %/year.
const (
	maxWeightPct = 500
	maxFeesPct   = 20
)

// validWeight reports whether a line's weight, in percent, is in range. NaN
// is not.
func validWeight(pct float64) bool { return pct > 0 && pct <= maxWeightPct }

// validFees reports whether a declared TER, in percent per year, is in range.
// NaN is not.
func validFees(pct float64) bool { return pct >= 0 && pct <= maxFeesPct }

// Line is one allocation of a portfolio built in code, in the in-memory
// convention: Weight is a FRACTION (0.6 for 60 %), Fees the asset's TER in
// PERCENT per year (0.2 for 0.20 %/yr), as Holding.Fees. A negative Fees
// means unknown, as an absent fee column does in a file; beware that the
// zero value DECLARES a zero TER, which Build then keeps instead of looking
// the fund's own up (BuildOptions.Fees), so write Fees: -1 when you do not
// know it.
type Line struct {
	ID     string  // ticker, ISIN or alias, SIM suffix allowed, as in a file
	Weight float64 // fraction of the portfolio, above 0 and at most 1
	Fees   float64 // TER in percent per year, 0 to 20; negative when unknown
}

// NewSpec builds the Spec of a portfolio assembled in code rather than read
// from a file. It validates and normalizes exactly as Parse does, through the
// same code: each Holding gets RawWeight = Weight*100 (the percent a file
// would carry) and a Weight renormalized so the weights sum to 1, with a
// Spec.Warnings line when the lines did not already sum to 100 %. Every
// directive is left unset: RebalanceDays, Capital, BorrowSpread and
// EnvelopeFees negative, Meta nil, no leverage, no flows, no SIM. So
// NewSpec(name, lines...) and Parse on the same lines written as a file are
// the same Spec, Name aside.
//
// It fails on no line at all, an empty identifier or one with a blank inside
// (a file could not carry it), a weight outside (0, 1] (a file line above
// 100 % needs "#meta leverage:on", which NewSpec does not set), a fee above
// 20 %/yr, and NaN anywhere. Duplicate identifiers are accepted, as Parse
// accepts them: they are two lines of the same asset.
func NewSpec(name string, lines ...Line) (*Spec, error) {
	spec := newSpec(name)
	for i, l := range lines {
		id := strings.TrimSpace(l.ID)
		if id == "" || strings.ContainsFunc(id, unicode.IsSpace) {
			return nil, fmt.Errorf("line %d: invalid identifier %q", i+1, l.ID)
		}
		h := Holding{RawWeight: l.Weight * 100, ID: id, Fees: -1}
		if !validWeight(h.RawWeight) || h.RawWeight > 100 {
			return nil, fmt.Errorf("line %d (%s): weight %g out of range (a fraction above 0 and at most 1)", i+1, id, l.Weight)
		}
		if l.Fees >= 0 || math.IsNaN(l.Fees) {
			if !validFees(l.Fees) {
				return nil, fmt.Errorf("line %d (%s): fees %g out of range (0-20 %%/year)", i+1, id, l.Fees)
			}
			h.Fees = l.Fees
		}
		spec.Holdings = append(spec.Holdings, h)
	}
	if err := spec.normalize(); err != nil {
		return nil, err
	}
	return spec, nil
}

// Single returns the specification of a portfolio entirely invested in one
// asset, named after the identifier. It backs comparison modes where each
// asset is treated as a standalone 100 % portfolio.
func Single(id string) *Spec {
	id = strings.TrimSpace(id)
	return &Spec{
		Name:          id,
		Holdings:      []Holding{{Weight: 1, RawWeight: 100, ID: id, Fees: -1}},
		RebalanceDays: -1,
		EnvelopeFees:  -1,
	}
}

// Flow is a periodic external cash flow: a fixed amount, or a percentage
// of the current value when Percent is true, applied once per Period.
type Flow struct {
	Amount  float64 // absolute amount, or percent of value when Percent
	Percent bool
	Period  Period
}

// Active reports whether the flow is set.
func (f Flow) Active() bool { return f.Amount > 0 }

// Period is a calendar flow frequency.
type Period string

// Supported flow periods.
const (
	Weekly    Period = "week"
	Monthly   Period = "month"
	Quarterly Period = "quarter"
	Yearly    Period = "year"
)

// parseFlow reads "500/month" or, when percentAllowed, "4%/year".
func parseFlow(val string, percentAllowed bool) (Flow, error) {
	amountStr, periodStr, found := strings.Cut(val, "/")
	if !found {
		return Flow{}, fmt.Errorf("%q: expected <amount>[%%]/<week|month|quarter|year>", val)
	}
	var f Flow
	if strings.HasSuffix(amountStr, "%") {
		if !percentAllowed {
			return Flow{}, fmt.Errorf("%q: percentages are not supported here", val)
		}
		f.Percent = true
		amountStr = strings.TrimSuffix(amountStr, "%")
	}
	amount, err := parseNumber(amountStr)
	if err != nil || amount <= 0 {
		return Flow{}, fmt.Errorf("%q: invalid amount", val)
	}
	if f.Percent && amount >= 100 {
		return Flow{}, fmt.Errorf("%q: percentage out of range", val)
	}
	f.Amount = amount
	switch Period(strings.ToLower(periodStr)) {
	case Weekly, Monthly, Quarterly, Yearly:
		f.Period = Period(strings.ToLower(periodStr))
	default:
		return Flow{}, fmt.Errorf("%q: unknown period (week, month, quarter or year)", val)
	}
	return f, nil
}

// metaDirective reports whether a trimmed line is a "#meta" directive and
// returns its content, comment stripped.
func metaDirective(line string) (bool, string) {
	const prefix = "#meta"
	if len(line) < len(prefix) || !strings.EqualFold(line[:len(prefix)], prefix) {
		return false, ""
	}
	rest := line[len(prefix):]
	if rest != "" && rest[0] != ' ' && rest[0] != '\t' {
		return false, "" // e.g. "#metadata…" is a plain comment
	}
	if i := strings.IndexByte(rest, '#'); i >= 0 {
		rest = rest[:i]
	}
	return true, strings.TrimSpace(rest)
}

// applyMeta interprets the key:value pairs of a #meta directive.
func (s *Spec) applyMeta(directives string) error {
	for _, tok := range strings.Fields(directives) {
		key, val, ok := strings.Cut(tok, ":")
		if !ok || val == "" {
			return fmt.Errorf("invalid #meta directive %q (expected key:value, e.g. rebalance:90)", tok)
		}
		key = strings.ToLower(key)
		if s.Meta == nil {
			s.Meta = map[string]string{}
		}
		s.Meta[key] = val
		switch key {
		case "rebalance":
			n, err := strconv.Atoi(val)
			if err != nil || n < 0 {
				return fmt.Errorf("#meta rebalance: %q is not a valid number of days", val)
			}
			s.RebalanceDays = n
		case "capital":
			f, err := parseNumber(val)
			if err != nil || f <= 0 {
				return fmt.Errorf("#meta capital: %q is not a valid amount", val)
			}
			s.Capital = f
		case "contribute":
			fl, err := parseFlow(val, false)
			if err != nil {
				return fmt.Errorf("#meta contribute: %v", err)
			}
			s.Contribute = fl
		case "withdraw":
			fl, err := parseFlow(val, true)
			if err != nil {
				return fmt.Errorf("#meta withdraw: %v", err)
			}
			s.Withdraw = fl
		case "leverage":
			switch strings.ToLower(val) {
			case "on":
				s.Leverage = true
			case "off":
				s.Leverage = false
			default:
				return fmt.Errorf("#meta leverage: invalid %q (expected on or off)", val)
			}
		case "sim":
			// "on" backcasts every holding (SIM variant); anything else,
			// including "off", leaves the files' identifiers as written.
			s.Sim = strings.EqualFold(val, "on")
		case "borrow-spread":
			f, err := parseNumber(val)
			if err != nil || f < 0 || f > 10 {
				return fmt.Errorf("#meta borrow-spread: %q is not a valid yearly percentage", val)
			}
			s.BorrowSpread = f
		case "optimize":
			os, err := optimize.ParseSpec(val)
			if err != nil {
				return fmt.Errorf("#meta optimize: %v", err)
			}
			s.Optimize = &os
		case "extra-fees", "envelope-fees":
			// Additional fees applied to the whole portfolio (envelope,
			// managed account, broker), on top of the assets' TERs.
			f, err := parseNumber(val)
			if err != nil || f < 0 || f > 20 {
				return fmt.Errorf("#meta extra-fees: %q is not a valid yearly percentage", val)
			}
			s.EnvelopeFees = f
		case "currencies":
			cs, err := parseCurrencies(val)
			if err != nil {
				return fmt.Errorf("#meta currencies: %v", err)
			}
			s.Currencies = cs
		default:
			s.Warnings = append(s.Warnings, fmt.Sprintf("unknown #meta directive ignored: %s", key))
		}
	}
	return nil
}

// parseWeight parses a percentage that may use a decimal comma and may carry
// a % suffix.
func parseWeight(s string) (float64, error) {
	w, err := parseNumber(s)
	if err != nil {
		return 0, fmt.Errorf("expected a number")
	}
	if !validWeight(w) {
		return 0, fmt.Errorf("must be greater than 0 and at most 500")
	}
	return w, nil
}

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

// isAlpha reports whether s is non-empty and all uppercase ASCII letters.
func isAlpha(s string) bool {
	for _, r := range s {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return s != ""
}

// parseNumber accepts a decimal comma and an optional % suffix.
func parseNumber(s string) (float64, error) {
	s = strings.TrimSuffix(s, "%")
	s = strings.ReplaceAll(s, ",", ".")
	return strconv.ParseFloat(s, 64)
}
