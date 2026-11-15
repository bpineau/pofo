package remote

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/bpineau/pofo/pkg/marketdata"
)

type consumerCase struct {
	Name, Prompt, Tool string
	Input              json.RawMessage
	Error              string
	Checks             map[string]any
	Absent             []string
	Warning            string
	WarningText        string `json:"warning_text"`
}

// These prompts state the consumer's intended interpretation. The assertions
// exercise the numerical contract, not an LLM's ability to choose calls.
func TestConsumerCorpus(t *testing.T) {
	raw, err := os.ReadFile("testdata/consumer_cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []consumerCase
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			t.Log(c.Prompt)
			src := fixtureSource()
			flat := *src["AAA"]
			flat.Symbol = "FLAT"
			flat.Points = append([]marketdata.Point(nil), flat.Points...)
			for i := range flat.Points {
				flat.Points[i].Close = 100
			}
			src["FLAT"] = &flat
			cpi := flat
			cpi.Symbol = "^HICP-FR"
			src[cpi.Symbol] = &cpi
			sim := *src["AAA"]
			sim.Symbol = "AAASIM"
			sim.SimulatedBefore = sim.Points[36].Date
			src[sim.Symbol] = &sim
			service := New(src, "corpus")
			result, e := service.Execute(context.Background(), c.Tool, c.Input)
			var direct []byte
			if e != nil {
				direct, err = json.Marshal(ErrorResult{publicError(e)})
			} else {
				direct, err = json.Marshal(result)
			}
			if err != nil {
				t.Fatal(err)
			}
			var want map[string]any
			if err := json.Unmarshal(direct, &want); err != nil {
				t.Fatal(err)
			}
			if c.Error != "" {
				if e == nil || publicError(e).Code != c.Error {
					t.Fatalf("error %v, want %s", e, c.Error)
				}
			} else if e != nil {
				t.Fatal(e)
			}
			for path, value := range c.Checks {
				got, ok := consumerValue(want, path)
				if !ok || !reflect.DeepEqual(got, value) {
					t.Errorf("%s = %#v (present %v), want %#v", path, got, ok, value)
				}
			}
			for _, path := range c.Absent {
				if _, ok := consumerValue(want, path); ok {
					t.Errorf("unexpected %s", path)
				}
			}
			if c.Warning != "" || c.WarningText != "" {
				found := false
				if result != nil {
					for _, w := range result.Warnings {
						if (c.Warning == "" || w.Code == c.Warning) && (c.WarningText == "" || strings.Contains(w.Message, c.WarningText)) {
							found = true
						}
					}
				}
				if !found {
					t.Errorf("missing warning %s / %s", c.Warning, c.WarningText)
				}
			}
			// One fresh handler per case keeps rate limits real without unrelated
			// corpus cases exhausting a shared artificial client quota.
			h := Handler(service, HTTPOptions{API: true, MCP: true})
			path := ""
			for _, op := range operations {
				if op.name == c.Tool {
					path = op.path
				}
			}
			if path == "" {
				t.Fatal("unknown corpus operation")
			}
			api := httptest.NewRecorder()
			if c.Tool == "search_assets" {
				var args map[string]any
				if err := json.Unmarshal(c.Input, &args); err != nil {
					t.Fatal(err)
				}
				q := url.Values{}
				for k, v := range args {
					q.Set(k, fmt.Sprint(v))
				}
				h.ServeHTTP(api, httptest.NewRequest("GET", path+"?"+q.Encode(), nil))
			} else {
				api = callHTTP(h, path, string(c.Input))
			}
			if (c.Error == "" && api.Code != 200) || (c.Error != "" && api.Code < 400) {
				t.Fatalf("HTTP status %d does not match expected error %q", api.Code, c.Error)
			}
			rpcBody, err := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": c.Name, "method": "tools/call", "params": map[string]any{"name": c.Tool, "arguments": c.Input}})
			if err != nil {
				t.Fatal(err)
			}
			mcp := callHTTP(h, "/mcp", string(rpcBody))
			var rpc struct {
				ID     string
				Result struct {
					IsError bool
					Content []struct{ Type, Text string }
				}
				Error any
			}
			if err := json.Unmarshal(mcp.Body.Bytes(), &rpc); err != nil {
				t.Fatal(err)
			}
			if mcp.Code != 200 || rpc.Error != nil || rpc.ID != c.Name || rpc.Result.IsError != (c.Error != "") || len(rpc.Result.Content) != 1 {
				t.Fatalf("invalid MCP envelope: %s", mcp.Body)
			}
			if rpc.Result.Content[0].Type != "text" {
				t.Fatalf("MCP content type = %q, want text", rpc.Result.Content[0].Type)
			}
			for transport, body := range map[string][]byte{"http": api.Body.Bytes(), "mcp": []byte(rpc.Result.Content[0].Text)} {
				var got map[string]any
				if err := json.Unmarshal(body, &got); err != nil {
					t.Fatalf("%s: %v", transport, err)
				}
				// Timestamps and execution durations intentionally differ.
				for _, obj := range []map[string]any{want, got} {
					if meta, ok := obj["meta"].(map[string]any); ok {
						delete(meta, "computed_at")
						delete(meta, "duration_ms")
					}
				}
				if !reflect.DeepEqual(want, got) {
					t.Errorf("%s differs from service: %s", transport, body)
				}
			}
		})
	}
}

func consumerValue(value any, path string) (any, bool) {
	for _, part := range strings.Split(path, ".") {
		switch v := value.(type) {
		case map[string]any:
			var ok bool
			value, ok = v[part]
			if !ok {
				return nil, false
			}
		case []any:
			i, err := strconv.Atoi(part)
			if err != nil || i < 0 || i >= len(v) {
				return nil, false
			}
			value = v[i]
		default:
			return nil, false
		}
	}
	return value, true
}
