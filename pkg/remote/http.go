package remote

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"mime"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bpineau/pofo/pkg/fire"
)

// Event records costs without request contents, query strings, identifiers or amounts.
type Event struct {
	Operation     string        // Operation is the stable operation name.
	Duration      time.Duration // Duration includes request admission and execution.
	QueueWait     time.Duration // QueueWait measures shared computation admission wait.
	ResponseBytes int           // ResponseBytes is the full encoded response size.
	Code          string        // Code is the public rejection code, empty on success.
}

// HTTPOptions configures independent API/MCP mounts and shared server admission.
// Origins is an exact origin allowlist, empty by default; no credentials are sent.
type HTTPOptions struct {
	API     bool     // API enables /api/v1 routes.
	MCP     bool     // MCP enables /mcp.
	Origins []string // Origins lists exact permitted browser origins.
	// ClientID must implement the deployment's trusted proxy policy. Nil uses socket IP.
	ClientID func(*http.Request) string
	// Acquire shares a bounded asset/portfolio queue with other server surfaces.
	// FIRE operations instead use fire.Acquire, shared with the web simulator.
	// It returns a release function, called exactly once after successful acquisition.
	// Nil uses a private gate with two active requests and sixteen waiting.
	Acquire func(context.Context) (func(), error)
	Observe func(Event) // Observe receives privacy-preserving operation cost events.
}
type httpServer struct {
	service *Service
	opt     HTTPOptions
	mu      sync.Mutex
	clients map[string]*bucket
}
type bucket struct {
	tokens float64
	at     time.Time
}

// Handler serves only enabled /api/v1 routes and /mcp. Requests have a 60-second
// deadline, 64 KiB body limit, 1 MiB response limit and 60/minute/client burst 10.
// Context cancellation does not issue a second response or cancel unrelated calls.
func Handler(s *Service, opt HTTPOptions) http.Handler {
	if opt.Acquire == nil {
		opt.Acquire = boundedAdmission(2, 16)
	}
	return &httpServer{s, opt, sync.Mutex{}, map[string]*bucket{}}
}
func (h *httpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	mcp := r.URL.Path == "/mcp" && h.opt.MCP
	api := strings.HasPrefix(r.URL.Path, "/api/v1/") && h.opt.API
	if !mcp && !api {
		http.NotFound(w, r)
		return
	}
	if origin := r.Header.Get("Origin"); origin != "" {
		allowed := false
		for _, o := range h.opt.Origins {
			if origin == o {
				allowed = true
			}
		}
		if !allowed {
			http.Error(w, "Origin not allowed", http.StatusForbidden)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Add("Vary", "Origin")
	}
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, MCP-Protocol-Version")
		w.Header().Set("Access-Control-Max-Age", "600")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()
	r = r.WithContext(context.WithValue(ctx, costKey{}, &requestCost{}))
	if mcp {
		h.mcp(w, r)
		return
	}
	h.api(w, r)
}
func (h *httpServer) admit(r *http.Request, operation string) (func(), error) {
	key := ""
	if h.opt.ClientID != nil {
		key = h.opt.ClientID(r)
	} else {
		key, _, _ = net.SplitHostPort(r.RemoteAddr)
	}
	now := time.Now()
	h.mu.Lock()
	b := h.clients[key]
	if b == nil {
		if len(h.clients) >= 10000 {
			oldest := ""
			at := now
			for k, v := range h.clients {
				if v.at.Before(at) {
					oldest, at = k, v.at
				}
			}
			delete(h.clients, oldest)
		}
		b = &bucket{10, now}
		h.clients[key] = b
	}
	b.tokens = min(10, b.tokens+now.Sub(b.at).Seconds())
	b.at = now
	if b.tokens < 1 {
		h.mu.Unlock()
		return nil, &Error{Code: "rate_limited", Message: "Request allowance exhausted; retry in one second.", Retryable: true, RetryAfterSeconds: 1}
	}
	b.tokens--
	h.mu.Unlock()
	acquire := h.opt.Acquire
	if operation == "simulate_fire" || operation == "solve_fire" {
		acquire = fire.Acquire
	}
	if acquire != nil {
		start := time.Now()
		release, err := acquire(r.Context())
		if cost, ok := r.Context().Value(costKey{}).(*requestCost); ok {
			cost.wait = time.Since(start)
		}
		return release, err
	}
	return func() {}, nil
}
func (h *httpServer) execute(r *http.Request, name string, input []byte) (*Result, error) {
	release, err := h.admit(r, name)
	if err != nil {
		return nil, err
	}
	defer release()
	return h.service.Execute(r.Context(), name, input)
}
func method(w http.ResponseWriter, r *http.Request, want string) bool {
	if r.Method == want {
		return true
	}
	w.Header().Set("Allow", want)
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	return false
}
func readBody(w http.ResponseWriter, r *http.Request) ([]byte, *Error) {
	ct, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || ct != "application/json" {
		return nil, &Error{Code: "unsupported_media_type", Message: "Use Content-Type: application/json."}
	}
	b, err := io.ReadAll(http.MaxBytesReader(w, r.Body, MaxRequestBytes))
	if err != nil {
		return nil, &Error{Code: "request_too_large", Message: "Request exceeds 64 KiB."}
	}
	return b, nil
}
func (h *httpServer) api(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/v1/openapi.json" {
		if !method(w, r, "GET") {
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "public, max-age=300")
		_, _ = w.Write(OpenAPI(h.service.Version))
		return
	}
	if r.URL.Path == "/api/v1/" {
		if !method(w, r, "GET") {
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.WriteString(w, `<!doctype html><html lang="en"><meta charset="utf-8"><title>Pofo numerical API</title><h1>Pofo numerical API</h1><p>Five read-only numerical operations. Percentages use 5 for 5%; analysis is nominal, FIRE is real EUR.</p><p><a href="openapi.json">OpenAPI 3.1 contract</a></p><pre>curl '/api/v1/assets/search?query=IWDA'
curl -H 'Content-Type: application/json' -d '{"ids":["IWDA"]}' /api/v1/assets/analyze</pre></html>`)
		return
	}
	var op *operation
	for i := range operations {
		if operations[i].path == r.URL.Path {
			op = &operations[i]
			break
		}
	}
	if op == nil {
		http.NotFound(w, r)
		return
	}
	var input []byte
	if op.name == "search_assets" {
		if !method(w, r, "GET") {
			return
		}
		q, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			h.failure(w, r, &Error{Code: "malformed_json", Message: "Invalid query encoding."})
			return
		}
		in := map[string]any{}
		for k, v := range q {
			if len(v) != 1 {
				h.failure(w, r, invalid(k, "Duplicate query parameter."))
				return
			}
			if k == "limit" {
				n, e := strconv.Atoi(v[0])
				if e != nil {
					h.failure(w, r, invalid(k, "Limit must be an integer."))
					return
				}
				in[k] = n
			} else {
				in[k] = v[0]
			}
		}
		input, _ = json.Marshal(in)
	} else {
		if !method(w, r, "POST") {
			return
		}
		var e *Error
		input, e = readBody(w, r)
		if e != nil {
			h.failure(w, r, e)
			return
		}
	}
	start := time.Now()
	result, err := h.execute(r, op.name, input)
	if err != nil {
		h.failure(w, r, publicError(err))
		h.observe(r, op.name, start, 0, publicError(err).Code)
		return
	}
	n, e := writeJSON(w, r, result, http.StatusOK)
	h.observe(r, op.name, start, n, e)
}
func (h *httpServer) observe(r *http.Request, name string, start time.Time, n int, code string) {
	if h.opt.Observe != nil {
		wait := time.Duration(0)
		if cost, ok := r.Context().Value(costKey{}).(*requestCost); ok {
			wait = cost.wait
		}
		h.opt.Observe(Event{Operation: name, Duration: time.Since(start), QueueWait: wait, ResponseBytes: n, Code: code})
	}
}
func status(e *Error) int {
	switch e.Code {
	case "malformed_json":
		return 400
	case "request_too_large":
		return 413
	case "unsupported_media_type":
		return 415
	case "rate_limited":
		return 429
	case "upstream_unavailable":
		return 502
	case "server_busy":
		return 503
	case "deadline_exceeded", "cancelled":
		return 504
	case "internal_error":
		return 500
	default:
		return 422
	}
}
func (h *httpServer) failure(w http.ResponseWriter, r *http.Request, e *Error) {
	if e.RetryAfterSeconds > 0 {
		w.Header().Set("Retry-After", strconv.Itoa(e.RetryAfterSeconds))
	}
	writeJSON(w, r, ErrorResult{e}, status(e))
}
func writeJSON(w http.ResponseWriter, r *http.Request, v any, code int) (int, string) {
	b, err := json.Marshal(v)
	problem := ""
	if err != nil {
		problem = "internal_error"
		code = 500
		b, _ = json.Marshal(ErrorResult{&Error{Code: problem, Message: "Result could not be encoded."}})
	} else if len(b) > MaxResponseBytes {
		problem = "response_too_large"
		code = 422
		b, _ = json.Marshal(ErrorResult{&Error{Code: problem, Message: "Response exceeds 1 MiB; remove sections or split the request."}})
	}
	if r.Context().Err() == context.Canceled {
		return 0, "cancelled"
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	n, _ := w.Write(b)
	return n, problem
}

type rpcEnvelope struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   json.RawMessage `json:"error,omitempty"`
}
type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}
type content struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
type toolResult struct {
	Content []content `json:"content"`
	IsError bool      `json:"isError"`
}

var integerID = regexp.MustCompile(`^-?(0|[1-9][0-9]*)$`)

func validRPCID(id json.RawMessage) bool {
	b := bytes.TrimSpace(id)
	if len(b) == 0 {
		return false
	}
	if b[0] == '"' {
		var s string
		return json.Unmarshal(b, &s) == nil
	}
	if integerID.Match(b) {
		return true
	}
	var number json.Number
	if json.Unmarshal(b, &number) != nil || number.String() == "" {
		return false
	}
	mantissa := number.String()
	exponent := int64(0)
	if i := strings.IndexAny(mantissa, "eE"); i >= 0 {
		raw := mantissa[i+1:]
		mantissa = mantissa[:i]
		v, err := strconv.ParseInt(raw, 10, 32)
		if err != nil {
			if strings.HasPrefix(raw, "-") {
				exponent = -MaxRequestBytes
			} else {
				exponent = MaxRequestBytes
			}
		} else {
			exponent = v
		}
	}
	fractional := 0
	if i := strings.IndexByte(mantissa, '.'); i >= 0 {
		fractional = len(mantissa) - i - 1
		mantissa = mantissa[:i] + mantissa[i+1:]
	}
	mantissa = strings.TrimPrefix(mantissa, "-")
	if strings.Trim(mantissa, "0") == "" {
		return true
	}
	trailing := len(mantissa) - len(strings.TrimRight(mantissa, "0"))
	return exponent+int64(trailing) >= int64(fractional)
}
func rpcFail(w http.ResponseWriter, r *http.Request, id json.RawMessage, code int, msg string) {
	if !validRPCID(id) {
		id = json.RawMessage("null")
	}
	writeJSON(w, r, rpcResponse{JSONRPC: "2.0", ID: id, Error: &rpcError{code, msg}}, 200)
}
func acceptsMCP(s string) bool {
	jsonOK, sseOK := false, false
	for _, part := range strings.Split(s, ",") {
		mt, params, e := mime.ParseMediaType(strings.TrimSpace(part))
		q := 1.
		if raw, ok := params["q"]; ok {
			var err error
			q, err = strconv.ParseFloat(raw, 64)
			if err != nil || math.IsNaN(q) || math.IsInf(q, 0) || q < 0 || q > 1 {
				continue
			}
		}
		if e != nil || q == 0 {
			continue
		}
		jsonOK = jsonOK || mt == "application/json" || mt == "*/*"
		sseOK = sseOK || mt == "text/event-stream" || mt == "*/*"
	}
	return jsonOK && sseOK
}
func (h *httpServer) mcp(w http.ResponseWriter, r *http.Request) {
	if v := r.Header.Get("MCP-Protocol-Version"); v != "" && v != ProtocolVersion {
		http.Error(w, "Unsupported MCP protocol version; initialize with 2025-11-25", 400)
		return
	}
	if !method(w, r, "POST") {
		return
	}
	if !acceptsMCP(r.Header.Get("Accept")) {
		http.Error(w, "Accept must include application/json and text/event-stream", http.StatusNotAcceptable)
		return
	}
	b, e := readBody(w, r)
	if e != nil {
		h.failure(w, r, e)
		return
	}
	var env rpcEnvelope
	if err := Decode(b, &env); err != nil {
		code := -32600
		if !json.Valid(b) {
			code = -32700
		}
		rpcFail(w, r, nil, code, "Invalid JSON-RPC message")
		return
	}
	if env.JSONRPC != "2.0" || len(env.ID) > 0 && !validRPCID(env.ID) || env.Method == "" {
		rpcFail(w, r, nil, -32600, "Invalid JSON-RPC envelope")
		return
	}
	if len(env.Result) > 0 || len(env.Error) > 0 {
		rpcFail(w, r, env.ID, -32600, "Unexpected response message")
		return
	}
	if len(env.ID) == 0 {
		if !strings.HasPrefix(env.Method, "notifications/") {
			w.WriteHeader(400)
			return
		}
		w.WriteHeader(202)
		return
	}
	respond := func(value any) { writeJSON(w, r, rpcResponse{JSONRPC: "2.0", ID: env.ID, Result: value}, 200) }
	params := env.Params
	if len(params) == 0 {
		params = []byte("{}")
	}
	switch env.Method {
	case "initialize":
		var p struct {
			ProtocolVersion string                     `json:"protocolVersion"`
			Capabilities    map[string]json.RawMessage `json:"capabilities"`
			ClientInfo      map[string]json.RawMessage `json:"clientInfo"`
			Meta            map[string]json.RawMessage `json:"_meta,omitempty"`
		}
		var clientName, clientVersion string
		decodeErr := Decode(params, &p)
		nameErr := json.Unmarshal(p.ClientInfo["name"], &clientName)
		versionErr := json.Unmarshal(p.ClientInfo["version"], &clientVersion)
		if decodeErr != nil || p.ProtocolVersion == "" || p.Capabilities == nil || nameErr != nil || versionErr != nil || clientName == "" || clientVersion == "" {
			rpcFail(w, r, env.ID, -32602, "Invalid initialize parameters")
			return
		}
		respond(map[string]any{"protocolVersion": ProtocolVersion, "capabilities": map[string]any{"tools": map[string]any{}}, "serverInfo": map[string]string{"name": "pofo", "version": h.service.Version}, "instructions": Instructions})
	case "ping":
		var p struct {
			Meta map[string]json.RawMessage `json:"_meta,omitempty"`
		}
		if Decode(params, &p) != nil {
			rpcFail(w, r, env.ID, -32602, "Invalid ping parameters")
			return
		}
		respond(map[string]any{})
	case "tools/list":
		var p struct {
			Cursor string                     `json:"cursor,omitempty"`
			Meta   map[string]json.RawMessage `json:"_meta,omitempty"`
		}
		if Decode(params, &p) != nil || p.Cursor != "" {
			rpcFail(w, r, env.ID, -32602, "This tool list has no cursor")
			return
		}
		respond(map[string]any{"tools": Tools()})
	case "tools/call":
		var p struct {
			Name      string                     `json:"name"`
			Arguments json.RawMessage            `json:"arguments,omitempty"`
			Meta      map[string]json.RawMessage `json:"_meta,omitempty"`
		}
		if Decode(params, &p) != nil || p.Name == "" {
			rpcFail(w, r, env.ID, -32602, "Invalid tool call")
			return
		}
		known := false
		for _, op := range operations {
			if op.name == p.Name {
				known = true
			}
		}
		if !known {
			rpcFail(w, r, env.ID, -32602, "Unknown tool")
			return
		}
		if len(p.Arguments) == 0 {
			p.Arguments = []byte("{}")
		}
		start := time.Now()
		result, err := h.execute(r, p.Name, p.Arguments)
		var domain any = result
		code := ""
		if err != nil {
			pe := publicError(err)
			domain = ErrorResult{pe}
			code = pe.Code
		}
		text, marshalErr := json.Marshal(domain)
		if marshalErr != nil {
			text, _ = json.Marshal(ErrorResult{&Error{Code: "internal_error", Message: "Result could not be encoded."}})
			err = marshalErr
			code = "internal_error"
		}
		response := rpcResponse{JSONRPC: "2.0", ID: env.ID, Result: toolResult{[]content{{"text", string(text)}}, err != nil}}
		encoded, _ := json.Marshal(response)
		if len(encoded) > MaxResponseBytes {
			code = "response_too_large"
			text, _ = json.Marshal(ErrorResult{&Error{Code: code, Message: "Remove sections or split the request; response exceeds 1 MiB."}})
			response.Result = toolResult{[]content{{"text", string(text)}}, true}
		}
		n, _ := writeJSON(w, r, response, 200)
		h.observe(r, p.Name, start, n, code)
	default:
		rpcFail(w, r, env.ID, -32601, fmt.Sprintf("Unknown method %q", env.Method))
	}
}

type costKey struct{}
type requestCost struct{ wait time.Duration }

func boundedAdmission(parallel, waiting int) func(context.Context) (func(), error) {
	slots, queue := make(chan struct{}, parallel), make(chan struct{}, waiting)
	return func(ctx context.Context) (func(), error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		release := func() { <-slots }
		select {
		case slots <- struct{}{}:
			return release, nil
		default:
		}
		select {
		case queue <- struct{}{}:
			defer func() { <-queue }()
		default:
			return nil, &Error{Code: "server_busy", Message: "Computation queue is full; retry later.", Retryable: true}
		}
		select {
		case slots <- struct{}{}:
			return release, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}
