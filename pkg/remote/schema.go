package remote

import (
	"encoding/json"
	"reflect"
	"strings"
)

// ProtocolVersion is the single implemented stateless Streamable HTTP revision.
const ProtocolVersion = "2025-11-25"

// Instructions is the compact workflow guidance returned by MCP initialize.
const Instructions = `Search canonical asset IDs, then analyze, simulate_fire or solve_fire. Percentages use 5 for 5%, dates use YYYY-MM-DD. SIM IDs or history=extended explicitly request reconstructed history; default analysis uses observed quotes. Asset/portfolio returns are nominal; FIRE amounts are real EUR with annual withdrawal steps. Portfolio FIRE uses fixed monthly weights, not the analysis rebalancing strategy. Inspect model basis: broad_sample is an external domestic 60/40 reference, not your allocation. Historical results sample overlapping cohorts with replacement, not independent historical evidence. The default tax rate is a French taxable-account planning assumption: verify or override it and inspect meta.defaulted. Zero simulated failures does not prove zero risk. safe_spending is explicitly a fixed_real reference even for adaptive simulations. solve_fire supports fixed_real only: adaptive capital risk is not monotone. A boundary result is not a global optimum; an unreachable target has no solution. Use sections for details and cache_only to forbid network work. No portfolios are saved.`

// Tool is one MCP input contract. Outputs are documented in OpenAPI and returned once as text.
type Tool struct {
	Name        string          `json:"name"`        // Name is the stable tool identifier.
	Description string          `json:"description"` // Description is compact workflow guidance.
	InputSchema map[string]any  `json:"inputSchema"` // InputSchema is shared with the HTTP contract.
	Annotations map[string]bool `json:"annotations"` // Annotations describe read-only and external-access behavior.
}
type operation struct {
	name, path, description string
	input                   reflect.Type
	output                  reflect.Type
}

var operations = []operation{
	{"search_assets", "/api/v1/assets/search", "Search the bundled catalog for canonical IDs; no network. Empty query browses; limit 1-50, default 10.", reflect.TypeFor[SearchInput](), reflect.TypeFor[SearchData]()},
	{"analyze_assets", "/api/v1/assets/analyze", "Analyze 1-10 exact IDs. Default identity/statistics; optional metadata, annual_returns, drawdowns. Percentages, nominal returns; no raw quotes.", reflect.TypeFor[AssetsInput](), reflect.TypeFor[AssetsData]()},
	{"analyze_portfolio", "/api/v1/portfolios/analyze", "Compare 1-4 allocations (1-20 holdings each) on one common window. Weights sum to 100%; rebalance_days=0 is buy-and-hold, default 90.", reflect.TypeFor[PortfoliosInput](), reflect.TypeFor[PortfoliosData]()},
	{"simulate_fire", "/api/v1/fire/simulate", "Simulate real-EUR annual withdrawals with six models. Default central, paths 2000, seed 7. Portfolio or explicit parametric market. Optional annual_bands, outcome_metrics, fixed-rule safe_spending.", reflect.TypeFor[SimulationInput](), reflect.TypeFor[SimulationData]()},
	{"solve_fire", "/api/v1/fire/solve", "Solve fixed_real annual_spending or capital at a target ruin percent (default 5). Omit solved amount. Explicit bounds/statuses; default brackets 0-15% capital or EUR 200000-6000000.", reflect.TypeFor[SolveInput](), reflect.TypeFor[SolveData]()},
}

// Tools returns deterministic, independent compact input definitions without output schemas.
func Tools() []Tool {
	out := make([]Tool, 0, len(operations))
	for _, o := range operations {
		out = append(out, Tool{o.name, o.description, inputSchema(o), map[string]bool{"readOnlyHint": true, "destructiveHint": false, "idempotentHint": true, "openWorldHint": o.name != "search_assets"}})
	}
	return out
}
func schema(t reflect.Type, output bool) map[string]any {
	if t == nil {
		return map[string]any{"type": "object"}
	}
	if t == reflect.TypeFor[Number]() {
		return map[string]any{"type": []string{"number", "null"}}
	}
	if t.Kind() == reflect.Pointer {
		v := schema(t.Elem(), output)
		if output {
			return map[string]any{"anyOf": []any{v, map[string]any{"type": "null"}}}
		}
		return v
	}
	switch t.Kind() {
	case reflect.Struct:
		props := map[string]any{}
		var req []string
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if !f.IsExported() {
				continue
			}
			tag := f.Tag.Get("json")
			name, opts, _ := strings.Cut(tag, ",")
			if name == "-" {
				continue
			}
			if f.Anonymous && name == "" {
				sub := schema(f.Type, output)
				for k, v := range sub["properties"].(map[string]any) {
					props[k] = v
				}
				if rr, ok := sub["required"].([]string); ok {
					req = append(req, rr...)
				}
				continue
			}
			if name == "" {
				name = f.Name
			}
			props[name] = schema(f.Type, output)
			if !strings.Contains(opts, "omitempty") {
				req = append(req, name)
			}
		}
		r := map[string]any{"type": "object", "properties": props, "additionalProperties": false}
		if len(req) > 0 {
			r["required"] = req
		}
		return r
	case reflect.Slice, reflect.Array:
		return map[string]any{"type": "array", "items": schema(t.Elem(), output)}
	case reflect.Map:
		return map[string]any{"type": "object", "additionalProperties": schema(t.Elem(), output)}
	case reflect.String:
		return map[string]any{"type": "string"}
	case reflect.Bool:
		return map[string]any{"type": "boolean"}
	case reflect.Int, reflect.Int64, reflect.Uint64:
		return map[string]any{"type": "integer"}
	case reflect.Float64:
		return map[string]any{"type": "number"}
	default:
		return map[string]any{}
	}
}
func inputSchema(o operation) map[string]any {
	s := schema(o.input, false)
	decorateSchema(s)
	p := s["properties"].(map[string]any)
	switch o.name {
	case "analyze_assets":
		p["currency"].(map[string]any)["enum"] = evaluationCurrencies
		p["sections"].(map[string]any)["items"] = enum("identity", "statistics", "metadata", "annual_returns", "drawdowns")
		p["ids"].(map[string]any)["maxItems"] = 10
		p["ids"].(map[string]any)["minItems"] = 1
	case "analyze_portfolio":
		p["currency"].(map[string]any)["enum"] = evaluationCurrencies
		p["sections"].(map[string]any)["items"] = enum("statistics", "composition", "holdings", "correlations", "attribution", "annual_returns", "drawdowns")
		p["portfolios"].(map[string]any)["maxItems"] = 4
		p["portfolios"].(map[string]any)["minItems"] = 1
	case "simulate_fire":
		p["sections"].(map[string]any)["items"] = enum("annual_bands", "safe_spending", "outcome_metrics")
		plan := p["plan"].(map[string]any)
		plan["required"] = []string{"capital", "annual_spending", "horizon_years"}
		s["dependentSchemas"] = map[string]any{"safe_spending": map[string]any{"required": []string{"sections"}, "properties": map[string]any{"sections": map[string]any{"contains": map[string]any{"const": "safe_spending"}}}}}
	case "solve_fire":
		plan := p["plan"].(map[string]any)
		plan["properties"].(map[string]any)["policy"] = enum("fixed_real")
		constraint := func(required, omitted string) map[string]any {
			return map[string]any{"properties": map[string]any{"plan": map[string]any{"required": []string{required}, "not": map[string]any{"required": []string{omitted}}}}}
		}
		// Keep all fields in the object shape: consumers that render oneOf
		// as a union can lose the sibling properties of partial branches.
		s["if"] = discriminator("solve_for", "capital")
		s["then"] = constraint("annual_spending", "capital")
		s["else"] = constraint("capital", "annual_spending")
	}
	return s
}
func enum(xs ...string) map[string]any { return map[string]any{"type": "string", "enum": xs} }

func discriminator(field, value string) map[string]any {
	return map[string]any{"properties": map[string]any{field: map[string]any{"const": value}}, "required": []string{field}}
}
func decorateSchema(s map[string]any) {
	props, _ := s["properties"].(map[string]any)
	for name, v := range props {
		p := v.(map[string]any)
		decorateSchema(p)
		switch name {
		case "asset_class":
			p["enum"] = append([]string{""}, assetClasses()...)
		case "market":
			forbidden := func(names ...string) map[string]any {
				var choices []any
				for _, name := range names {
					choices = append(choices, map[string]any{"required": []string{name}})
				}
				return map[string]any{"anyOf": choices}
			}
			p["description"] = "Parametric mode requires real_mean_return_pct, annual_volatility_pct and degrees_of_freedom. Portfolio mode requires holdings and fits those parameters. Do not mix the two sets."
			p["if"] = discriminator("mode", "parametric")
			p["then"] = map[string]any{"required": []string{"real_mean_return_pct", "annual_volatility_pct", "degrees_of_freedom"}, "not": forbidden("holdings", "history", "data_policy", "from", "to")}
			p["else"] = map[string]any{"required": []string{"holdings"}, "not": forbidden("real_mean_return_pct", "annual_volatility_pct", "degrees_of_freedom")}
		case "mode":
			p["enum"] = []string{"parametric", "portfolio"}
		case "history":
			p["enum"] = []string{"observed", "extended"}
		case "data_policy":
			p["enum"] = []string{"cached_or_fetch", "cache_only"}
		case "policy":
			p["enum"] = []string{"fixed_real", "guyton_klinger"}
		case "solve_for":
			p["enum"] = []string{"annual_spending", "capital"}
			p["description"] = "Omit this amount from plan; supply the other amount and horizon_years."
		case "model":
			p["enum"] = modelNames
		case "models":
			p["items"] = enum(modelNames...)
			p["minItems"] = 1
			p["maxItems"] = 6
			p["uniqueItems"] = true
		case "sections":
			p["uniqueItems"] = true
		case "holdings":
			p["minItems"] = 1
			p["maxItems"] = 20
		case "paths":
			p["minimum"] = 100
			p["maximum"] = 10000
			p["default"] = 2000
		case "seed":
			p["minimum"] = 0
			p["maximum"] = int64(1<<53 - 1)
			p["default"] = 7
		case "horizon_years":
			p["minimum"] = 1
			p["maximum"] = 100
		case "limit":
			p["minimum"] = 1
			p["maximum"] = 50
			p["default"] = 10
		case "query":
			p["maxLength"] = 200
		case "label":
			p["maxLength"] = 80
		case "from", "to":
			p["format"] = "date"
		case "capital":
			p["exclusiveMinimum"] = 0
			p["maximum"] = 1e10
		case "annual_spending", "annual_amount", "min", "max":
			p["minimum"] = 0
			p["maximum"] = 1e10
		case "buffer_years", "start_year":
			p["minimum"] = 0
			p["maximum"] = 100
		case "buffer_real_return_pct":
			p["minimum"] = -99
			p["maximum"] = 100
		case "real_mean_return_pct":
			p["minimum"] = -100
			p["maximum"] = 100
			p["description"] = "Arithmetic annual real return, percent."
		case "annual_volatility_pct":
			p["minimum"] = 0
			p["maximum"] = 200
		case "degrees_of_freedom":
			p["minimum"] = 0
			p["maximum"] = 1e6
		case "inflation_pct", "annual_fee_pct":
			p["minimum"] = 0
			p["maximum"] = 20
		case "real_drift_pct":
			p["minimum"] = -5
			p["maximum"] = 5
		case "rebalance_days":
			p["minimum"] = 0
			p["maximum"] = 36500
		case "tax_rate_pct":
			p["minimum"] = 0
			p["exclusiveMaximum"] = 100
		case "weight_pct", "initial_gain_pct", "target_ruin_pct":
			p["minimum"] = 0
			p["maximum"] = 100
		}
	}
	if items, ok := s["items"].(map[string]any); ok {
		decorateSchema(items)
	}
}

var modelNames = []string{"central", "bootstrap", "historical", "stress", "broad_sample", "lost_decade"}

// OpenAPI returns the OpenAPI 3.1 document using the same input schemas as MCP.
func OpenAPI(version string) []byte {
	paths := map[string]any{}
	for _, o := range operations {
		input := inputSchema(o)
		method := "post"
		op := map[string]any{"operationId": o.name, "description": o.description}
		if o.name == "search_assets" {
			method = "get"
			var ps []any
			for _, k := range []string{"query", "asset_class", "currency", "limit", "cursor"} {
				ps = append(ps, map[string]any{"name": k, "in": "query", "schema": input["properties"].(map[string]any)[k]})
			}
			op["parameters"] = ps
		} else {
			op["requestBody"] = map[string]any{"required": true, "content": map[string]any{"application/json": map[string]any{"schema": input}}}
		}
		result := schema(reflect.TypeFor[Result](), true)
		result["properties"].(map[string]any)["data"] = schema(o.output, true)
		meta := result["properties"].(map[string]any)["meta"].(map[string]any)["properties"].(map[string]any)
		if o.name == "simulate_fire" || o.name == "solve_fire" {
			meta["assumptions"] = schema(reflect.TypeFor[FireAssumptions](), true)
		} else if o.name != "search_assets" {
			meta["assumptions"] = schema(reflect.TypeFor[AnalysisAssumptions](), true)
		} else {
			delete(meta, "assumptions")
		}
		responses := map[string]any{"200": map[string]any{"description": "Complete numerical result; solve bounds may be unreachable.", "content": map[string]any{"application/json": map[string]any{"schema": result}}}}
		for _, status := range []string{"400", "413", "415", "422", "429", "500", "502", "503", "504"} {
			responses[status] = map[string]any{"description": "Typed error; no partial results.", "content": map[string]any{"application/json": map[string]any{"schema": schema(reflect.TypeFor[ErrorResult](), true)}}}
		}
		op["responses"] = responses
		paths[o.path] = map[string]any{method: op}
	}
	b, _ := json.Marshal(map[string]any{"openapi": "3.1.0", "info": map[string]string{"title": "Pofo numerical API", "version": "1", "description": "Engine " + version + ". Percentages use 5 for 5%; analysis is nominal, FIRE is real EUR. Defaults evolve with the engine and are reported. Undefined requested statistics are null."}, "paths": paths})
	return b
}
