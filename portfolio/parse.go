// Package portfolio parses portfolio description files and simulates their
// value over time.
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
)

// Holding is one allocation line of a portfolio file.
type Holding struct {
	Weight    float64 // normalized fraction of the portfolio (sums to 1)
	RawWeight float64 // weight as written in the file, in percent
	ID        string  // ticker or ISIN, verbatim
	Fees      float64 // declared TER in percent per year; negative when absent
	Note      string  // free text after the identifier, informational only
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

	// EnvelopeFees is the additional yearly fee of the hosting envelope
	// (assurance-vie, PER, mandat…) set by "#meta extra-fees:X" — fees
	// applied on top of the WHOLE portfolio, in addition to the assets'
	// own TERs — in percent per year;
	// negative when absent. Unlike asset TERs (already reflected in
	// prices), it must be deducted from the simulated performance.
	EnvelopeFees float64

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
//	<poids en %> <ticker, ISIN ou alias> [texte libre…]
//
// Everything after a # is a comment; blank lines and lines starting with //
// are ignored. Lines starting with "#meta" carry per-portfolio directives
// as key:value pairs — currently "rebalance:N" (days between rebalancings,
// 0 to never rebalance). Weights accept a decimal comma and an optional %
// suffix. If the weights do not sum to 100 they are normalized and a
// warning is recorded.
func Parse(name string, r io.Reader) (*Spec, error) {
	spec := &Spec{Name: name, RebalanceDays: -1, EnvelopeFees: -1}
	sc := bufio.NewScanner(r)
	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := strings.TrimSpace(sc.Text())
		if isMeta, rest := metaDirective(line); isMeta {
			if err := spec.applyMeta(rest); err != nil {
				return nil, fmt.Errorf("ligne %d: %w", lineNo, err)
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
			return nil, fmt.Errorf("ligne %d: attendu « <poids> <ticker/ISIN> [texte] », trouvé %q", lineNo, line)
		}
		w, err := parseWeight(fields[0])
		if err != nil {
			return nil, fmt.Errorf("ligne %d: poids %q invalide: %v", lineNo, fields[0], err)
		}
		h := Holding{RawWeight: w, ID: fields[1], Fees: -1}
		rest := fields[2:]
		// Optional third numeric column: the asset's TER in percent/year.
		if len(rest) > 0 {
			if fees, ferr := parseNumber(rest[0]); ferr == nil {
				if fees < 0 || fees > 20 {
					return nil, fmt.Errorf("ligne %d: frais %q hors limites (0–20 %%/an)", lineNo, rest[0])
				}
				h.Fees = fees
				rest = rest[1:]
			}
		}
		h.Note = strings.Join(rest, " ")
		spec.Holdings = append(spec.Holdings, h)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if len(spec.Holdings) == 0 {
		return nil, fmt.Errorf("aucune ligne d'allocation trouvée")
	}
	sum := 0.0
	for _, h := range spec.Holdings {
		sum += h.RawWeight
	}
	if math.Abs(sum-100) > 0.5 {
		spec.Warnings = append(spec.Warnings,
			fmt.Sprintf("les poids totalisent %.4g %% au lieu de 100 %% — ils ont été normalisés", sum))
	}
	for i := range spec.Holdings {
		spec.Holdings[i].Weight = spec.Holdings[i].RawWeight / sum
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
			return fmt.Errorf("directive #meta invalide %q (attendu clé:valeur, ex. rebalance:90)", tok)
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
				return fmt.Errorf("#meta rebalance: %q n'est pas un nombre de jours valide", val)
			}
			s.RebalanceDays = n
		case "extra-fees", "envelope-fees":
			// Frais additionnels appliqués à l'ensemble du portefeuille
			// (enveloppe, mandat, courtier), en plus des TER des actifs.
			f, err := parseNumber(val)
			if err != nil || f < 0 || f > 20 {
				return fmt.Errorf("#meta extra-fees: %q n'est pas un pourcentage annuel valide", val)
			}
			s.EnvelopeFees = f
		default:
			s.Warnings = append(s.Warnings, fmt.Sprintf("directive #meta inconnue ignorée: %s", key))
		}
	}
	return nil
}

// parseWeight parses a percentage that may use a decimal comma and may carry
// a % suffix.
func parseWeight(s string) (float64, error) {
	w, err := parseNumber(s)
	if err != nil {
		return 0, fmt.Errorf("nombre attendu")
	}
	if w <= 0 || w > 100 {
		return 0, fmt.Errorf("doit être compris entre 0 exclu et 100")
	}
	return w, nil
}

// parseNumber accepts a decimal comma and an optional % suffix.
func parseNumber(s string) (float64, error) {
	s = strings.TrimSuffix(s, "%")
	s = strings.ReplaceAll(s, ",", ".")
	return strconv.ParseFloat(s, 64)
}
