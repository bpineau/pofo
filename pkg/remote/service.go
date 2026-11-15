package remote

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"reflect"
	"slices"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/bpineau/pofo/pkg/analyze"
	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
)

// MaxRequestBytes bounds JSON inputs before parsing.
const MaxRequestBytes = 64 << 10

// MaxResponseBytes bounds the entire encoded transport response.
const MaxResponseBytes = 1 << 20

// Service executes the five read-only operations. Configure before concurrent use.
// Source must honor context cancellation and marketdata.WithCacheOnly. A Client
// does so; a caller-owned in-memory analyze.SeriesSource is also suitable.
type Service struct {
	Source  analyze.Source // Source provides exact histories and fee metadata.
	Version string         // Version identifies the running build or module.
	// BeforeFetch applies shared server admission at the actual data boundary.
	// Called once per canonical identifier per operation, including benchmarks.
	// It must return public *Error values rather than credential-bearing errors.
	BeforeFetch func(context.Context, string) error
}

// New constructs a service. A nil source permits catalog-only operations.
func New(src analyze.Source, version string) *Service { return &Service{Source: src, Version: version} }

var revision = sync.OnceValue(func() string {
	h := sha256.New()
	h.Write(datasets.AssetMeta())
	h.Write(datasets.CAPE())
	h.Write(datasets.BroadSample())
	for _, f := range []fs.FS{datasets.Simdata(), datasets.Refdata()} {
		_ = fs.WalkDir(f, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() {
				h.Write([]byte(path))
				b, e := fs.ReadFile(f, path)
				if e != nil {
					return e
				}
				h.Write(b)
			}
			return nil
		})
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
})

// Execute validates an entire request before any data access. Input must be one
// JSON object without duplicate or unknown keys. Errors are always public *Error.
func (s *Service) Execute(ctx context.Context, name string, input json.RawMessage) (result *Result, err error) {
	start := time.Now()
	r := &Result{Meta: Meta{SchemaVersion: "1", EngineVersion: s.Version, DatasetRevision: revision(), ComputedAt: start.UTC().Format(time.RFC3339), Sections: []string{}, Defaulted: []Default{}}, Warnings: []Warning{}}
	defer func() {
		if err != nil {
			err = publicError(err)
		}
		if result != nil {
			result.Meta.DurationMS = time.Since(start).Milliseconds()
		}
	}()
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	switch name {
	case "search_assets":
		var in SearchInput
		if err = Decode(input, &in); err == nil {
			r.Data, err = s.search(in, r)
		}
	case "analyze_assets":
		var in AssetsInput
		if err = Decode(input, &in); err == nil {
			r.Data, err = s.assets(ctx, in, r)
		}
	case "analyze_portfolio":
		var in PortfoliosInput
		if err = Decode(input, &in); err == nil {
			r.Data, err = s.portfolios(ctx, in, r)
		}
	case "simulate_fire":
		var in SimulationInput
		if err = Decode(input, &in); err == nil {
			r.Data, err = s.simulate(ctx, in, r)
		}
	case "solve_fire":
		var in SolveInput
		if err = Decode(input, &in); err == nil {
			r.Data, err = s.solve(ctx, in, r)
		}
	default:
		return nil, invalid("", "Unknown operation.")
	}
	if err != nil {
		return nil, err
	}
	return r, nil
}

// Decode rejects duplicate keys at any depth, unknown struct fields, null root,
// trailing values and excessive nesting. It accepts at most MaxRequestBytes.
func Decode(data []byte, dst any) error {
	if !utf8.Valid(data) {
		return &Error{Code: "malformed_json", Message: "JSON must be UTF-8 encoded."}
	}
	if len(data) > MaxRequestBytes {
		return &Error{Code: "request_too_large", Message: "Request exceeds 64 KiB."}
	}
	if len(bytes.TrimSpace(data)) == 0 || bytes.TrimSpace(data)[0] != '{' {
		return &Error{Code: "malformed_json", Message: "Expected one JSON object."}
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	if err := walkJSON(d, 0); err != nil {
		return &Error{Code: "malformed_json", Message: err.Error()}
	}
	if _, err := d.Token(); err != io.EOF {
		return &Error{Code: "malformed_json", Message: "Trailing JSON is not allowed."}
	}
	var value any
	d = json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	if err := d.Decode(&value); err != nil {
		return invalid("", "Invalid JSON input.")
	}
	if err := exactFields(value, reflect.TypeOf(dst).Elem(), ""); err != nil {
		return err
	}
	d = json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err := d.Decode(dst); err != nil {
		return invalid("", "Invalid input shape: "+err.Error())
	}
	return nil
}
func walkJSON(d *json.Decoder, depth int) error {
	if depth > 64 {
		return errors.New("JSON nesting exceeds 64 levels")
	}
	t, err := d.Token()
	if err != nil {
		return errors.New("malformed JSON")
	}
	delim, ok := t.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		seen := map[string]bool{}
		for d.More() {
			k, e := d.Token()
			if e != nil {
				return errors.New("malformed JSON object")
			}
			key, ok := k.(string)
			if !ok || seen[key] {
				return errors.New("duplicate JSON key")
			}
			seen[key] = true
			if e = walkJSON(d, depth+1); e != nil {
				return e
			}
		}
	case '[':
		for d.More() {
			if e := walkJSON(d, depth+1); e != nil {
				return e
			}
		}
	default:
		return errors.New("malformed JSON")
	}
	_, err = d.Token()
	return err
}
func invalid(path, msg string) *Error {
	return &Error{Code: "invalid_argument", Message: msg, Path: path}
}
func publicError(err error) *Error {
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return &Error{Code: "deadline_exceeded", Message: "Computation deadline exceeded; reduce paths, models or optional sections.", Retryable: true}
	case errors.Is(err, context.Canceled):
		return &Error{Code: "cancelled", Message: "Request cancelled."}
	case errors.Is(err, marketdata.ErrOffline):
		return &Error{Code: "data_unavailable_offline", Message: "Required fresh data is unavailable locally; allow fetching or choose bundled history."}
	case errors.Is(err, marketdata.ErrUnknownIdentifier):
		return &Error{Code: "unknown_asset", Message: "Exact asset identifier not found. Search the catalog or supply a precise ticker or ISIN."}
	default:
		return &Error{Code: "upstream_unavailable", Message: "Required data could not be obtained or analyzed.", Retryable: true}
	}
}
func ptr[T any](v T) *T { return &v }
func def[T any](p *T, v T, path string, r *Result) T {
	if p != nil {
		return *p
	}
	r.Meta.Defaulted = append(r.Meta.Defaulted, Default{path, v})
	return v
}
func warn(r *Result, code, msg, path string) {
	w := Warning{code, msg, path}
	if !slices.Contains(r.Warnings, w) {
		r.Warnings = append(r.Warnings, w)
	}
}
func warnings(r *Result, ws []string, path string) {
	for _, w := range ws {
		warn(r, "data_warning", w, path)
	}
}
func date(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.DateOnly)
}
func validID(id string) bool {
	if id == "" || len(id) > 100 || id != strings.TrimSpace(id) || strings.ContainsAny(id, "/\\:\x00") || strings.Contains(id, "..") {
		return false
	}
	return id == "^HICP-FR" || marketdata.KnownLocal(id) || marketdata.PlausibleID(id)
}
func canonical(id string) string {
	base, sim := marketdata.SplitSim(id)
	base = marketdata.CanonicalID(base)
	if sim {
		return base + "SIM"
	}
	return base
}
func baseID(id string) string { b, _ := marketdata.SplitSim(id); return marketdata.CanonicalID(b) }
func checkIDs(ids []string, history *string, path string) error {
	for i, id := range ids {
		p := fmt.Sprintf("%s[%d]", path, i)
		if !validID(id) {
			return invalid(p, "Use a catalog ID, ticker or checksummed ISIN; paths and URLs are not allowed.")
		}
		_, sim := marketdata.SplitSim(id)
		if sim && history != nil && *history == "observed" {
			return invalid(p, "SIM conflicts with explicit observed history.")
		}
	}
	return nil
}

// guardedSource intentionally does not forward arbitrary caller loaders without
// checking IDs; local-file ingestion is never part of the remote contract.
type localOnlyKey struct{}

type guardedSource struct {
	service *Service
	mu      sync.Mutex
	seen    map[string]bool
}

func (s *Service) source() *guardedSource { return &guardedSource{service: s, seen: map[string]bool{}} }
func (g *guardedSource) FetchExtended(ctx context.Context, id string, opt marketdata.FetchOptions) (*marketdata.Series, error) {
	if !validID(id) {
		return nil, invalid("id", "Invalid asset identifier.")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	key := baseID(id)
	g.mu.Lock()
	defer g.mu.Unlock()
	if !g.seen[key] {
		if len(g.seen) >= 40 {
			return nil, invalid("ids", "At most 40 unique data dependencies are allowed.")
		}
		localOnly, _ := ctx.Value(localOnlyKey{}).(bool)
		if f := g.service.BeforeFetch; f != nil && !localOnly {
			if err := f(ctx, key); err != nil {
				return nil, err
			}
		}
		g.seen[key] = true
	}
	opt.ExactOnly = true
	src := g.service.Source
	if src == nil {
		return nil, errors.New("source unavailable")
	}
	if loader, ok := src.(interface {
		Load(context.Context, string, marketdata.FetchOptions) (*marketdata.Series, error)
	}); ok {
		return loader.Load(ctx, id, opt)
	}
	return src.FetchExtended(ctx, id, opt)
}
func (g *guardedSource) Fees(ctx context.Context, id string) (float64, bool) {
	if !validID(id) || g.service.Source == nil {
		return 0, false
	}
	return g.service.Source.Fees(ctx, id)
}

// encoding/json otherwise accepts case-insensitive struct field matches, which
// would let a differently cased duplicate overwrite an already validated key.
func exactFields(value any, t reflect.Type, path string) error {
	if t == reflect.TypeFor[json.RawMessage]() || t.Kind() == reflect.Interface {
		return nil
	}
	if value == nil {
		return invalid(path, "Null is not a supplied input value; omit optional fields instead.")
	}
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.Struct:
		obj, ok := value.(map[string]any)
		if !ok {
			return invalid(path, "Expected an object.")
		}
		fields := map[string]reflect.Type{}
		var gather func(reflect.Type)
		gather = func(t reflect.Type) {
			for i := 0; i < t.NumField(); i++ {
				f := t.Field(i)
				if !f.IsExported() {
					continue
				}
				name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
				if f.Anonymous && name == "" {
					gather(f.Type)
					continue
				}
				if name == "-" {
					continue
				}
				if name == "" {
					name = f.Name
				}
				fields[name] = f.Type
			}
		}
		gather(t)
		for key, v := range obj {
			ft, ok := fields[key]
			p := key
			if path != "" {
				p = path + "." + key
			}
			if !ok {
				return invalid(p, "Unknown field; use exact documented field names.")
			}
			if err := exactFields(v, ft, p); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		if xs, ok := value.([]any); ok {
			for i, v := range xs {
				if err := exactFields(v, t.Elem(), fmt.Sprintf("%s[%d]", path, i)); err != nil {
					return err
				}
			}
		}
	case reflect.Map:
		if obj, ok := value.(map[string]any); ok {
			for key, v := range obj {
				if err := exactFields(v, t.Elem(), path+"."+key); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
