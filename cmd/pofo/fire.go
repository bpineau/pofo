// The -fire mode: the interactive FIRE / decumulation explorer (pkg/decumul).
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/bpineau/pofo/pkg/decumul/web"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/portfolio"
	"github.com/bpineau/pofo/pkg/scenario"
)

// runFire starts the embedded decumulation explorer on a local port and
// opens it in the browser. With a portfolio file it builds a historical
// real-return panel from the holdings (deflated by ^HICP-FR) so the UI can
// switch to the bootstrap/cohort models and re-weight allocations live. It
// blocks, serving until interrupted.
func runFire(ctx context.Context, opt *options, c *marketdata.Client, specs []*portfolio.Spec) error {
	// The charts render dark through the web package's own wrappers
	// (pkg/decumul/web/theme.go), not the chart process-global, so the FIRE
	// UI stays dark whether it runs alone (-fire) or beside the light /view
	// report (-serve).
	panel, labels := firePanel(ctx, opt, c, specs)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	url := "http://" + ln.Addr().String() + "/"
	fmt.Fprintf(os.Stderr, "FIRE explorer on %s (Ctrl-C to stop)\n", url)
	if !opt.noOpen {
		openBrowser(url)
	}
	// main() routes SIGINT/SIGTERM into ctx (signal.NotifyContext), which
	// replaces the default die-on-Ctrl-C behavior, so the server must watch
	// the context and shut down when it fires.
	srv := &http.Server{Handler: web.Handler(panel, labels)}
	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutCtx)
	}()
	if err := srv.Serve(ln); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// firePanel builds the historical real-return panel for the FIRE UI from
// the first spec's holdings (SIM-extended, deflated by ^HICP-FR), or
// returns (nil, nil) when no portfolio was given: the UI then runs its
// parametric models only.
func firePanel(ctx context.Context, opt *options, c *marketdata.Client, specs []*portfolio.Spec) (*scenario.Panel, []string) {
	var panel *scenario.Panel
	var labels []string
	if len(specs) > 0 {
		var assets []web.AssetSeries
		for _, h := range specs[0].Holdings {
			// Honour "#meta sim:on" exactly like portfolio.Build: fetch the
			// SIM (backcast-extended) variant, falling back to the real
			// quotes when no backcast exists. The FIRE panel needs the deep
			// history; real-only overlaps of recent funds are often too
			// short to fit or resample a retirement-length horizon.
			fetchID := portfolio.SimFetchID(h.ID, specs[0].Sim)
			s, err := fetchAsset(ctx, c, fetchID, opt)
			if err != nil && fetchID != h.ID {
				log.Printf("fire: %s: no simulated history, using real quotes", h.ID)
				s, err = fetchAsset(ctx, c, h.ID, opt)
			}
			if err != nil {
				log.Printf("fire: skipping %s: %v", h.ID, err)
				continue
			}
			labels = append(labels, h.ID)
			assets = append(assets, web.AssetSeries{Weight: h.Weight, Points: s.Points})
		}
		if hicp, err := fetchAsset(ctx, c, "^HICP-FR", opt); err == nil {
			if pnl, err := web.BuildMonthlyPanel(assets, hicp.Points); err == nil {
				panel = &pnl
			} else {
				log.Printf("fire: no historical panel: %v", err)
			}
		}
	}
	return panel, labels
}
