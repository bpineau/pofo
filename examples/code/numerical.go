//go:build ignore

// Run a compact numerical operation directly, without an HTTP or MCP transport.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/remote"
)

func main() {
	operation := flag.String("operation", "search_assets", "search_assets, analyze_assets, analyze_portfolio, simulate_fire or solve_fire")
	input := flag.String("input", `{"query":"IWDA","limit":1}`, "JSON operation input; use data_policy=cache_only to prohibit network work")
	flag.Parse()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	client := marketdata.NewClient("")
	client.Controls = marketdata.NewControls(60, time.Minute, 10)
	result, err := remote.New(client, "example").Execute(ctx, *operation, []byte(*input))
	if err != nil {
		log.Fatal(err)
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(data))
}
