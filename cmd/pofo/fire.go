// The -fire mode: the interactive FIRE / decumulation explorer (pkg/decumul).
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"strings"
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
	// Long-lived like -serve: print each informational fetch line once, keep
	// every warning (see logdedup.go).
	dedupServerLog()
	// The charts render dark through the web package's own wrappers
	// (pkg/decumul/web/theme.go), not the chart process-global, so the FIRE
	// UI stays dark whether it runs alone (-fire) or beside the light /view
	// report (-serve).
	build := firePanel(ctx, opt, c, specs)
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
	// No picker here: the loader's mounts (/firesimulator/e/, /p/) and the
	// catalog endpoint it searches only exist under -serve. The page then
	// states how to bind a portfolio from the command line instead. The
	// source label still rides, so a file argument names the market.
	srv := &http.Server{Handler: web.Handler(build.panel, build.labels, append(build.options(),
		web.WithSourceLabel(specsLabel(specs)),
		web.WithBeaconToken(opt.cfBeaconToken))...)}
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

// fireBuild is what a FIRE mount is built from: the historical panel (nil
// when there is none), its holdings' labels, and the notes stating what the
// panel could not carry of the written portfolio (web.WithNotes).
type fireBuild struct {
	panel  *scenario.Panel
	labels []string
	notes  []string
}

// options are the web.Handler options that carry the build's notes.
func (b fireBuild) options() []web.Option {
	if len(b.notes) == 0 {
		return nil
	}
	return []web.Option{web.WithNotes(b.notes)}
}

// specNotes names what of spec a FIRE panel cannot represent: its live
// allocation bar is a fully invested, fixed-weight mix of the holdings'
// monthly real returns, so a leverage is read as fully invested (no
// financing leg, no borrow cost), and portfolio-level fees, an optimizer and
// external flows are left to the report.
func specNotes(spec *portfolio.Spec) []string {
	var out []string
	sum := 0.0
	for _, h := range spec.Holdings {
		sum += h.Weight
	}
	if spec.Leverage && math.Abs(sum-1) > 1e-9 {
		out = append(out, fmt.Sprintf("The portfolio is %.0f %% invested, read here as 100 %%: the history carries no financing leg nor borrow cost, so it is the unlevered mix of the same holdings.", sum*100))
	}
	if spec.EnvelopeFees > 0 {
		out = append(out, fmt.Sprintf("The portfolio's extra fees (%.2f %%/yr) are not deducted from the history; the holdings' own fees are, being in their prices.", spec.EnvelopeFees))
	}
	if spec.Optimize != nil {
		out = append(out, "The history reads the WRITTEN weights, not the optimizer's.")
	}
	return out
}

// firePanel builds the historical real-return panel for the FIRE UI from
// the first spec's holdings (SIM-extended, deflated by ^HICP-FR), or
// returns no panel when no portfolio was given: the UI then runs its
// parametric models only. A holding that cannot be read leaves NO panel
// (the surviving holdings would be another portfolio), and the notes say
// why, as they state what the panel left out (Coverage.Notes) and what of
// the spec it cannot represent (specNotes).
func firePanel(ctx context.Context, opt *options, c *marketdata.Client, specs []*portfolio.Spec) fireBuild {
	var b fireBuild
	if len(specs) > 0 {
		b.notes = specNotes(specs[0])
		var assets []web.AssetSeries
		var missing []string
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
				log.Printf("warning: fire: %s unreadable, no historical panel: %v", h.ID, err)
				missing = append(missing, h.ID)
				continue
			}
			b.labels = append(b.labels, h.ID)
			assets = append(assets, web.AssetSeries{Weight: h.Weight, Points: s.Points})
		}
		if len(missing) > 0 {
			b.notes = append(b.notes, fmt.Sprintf("No historical panel: %s could not be read, and the other holdings alone are another portfolio.",
				strings.Join(missing, ", ")))
			return b
		}
		hicp, err := fetchAsset(ctx, c, "^HICP-FR", opt)
		if err != nil {
			log.Printf("warning: fire: no historical panel: inflation unreadable: %v", err)
			b.notes = append(b.notes, "No historical panel: the inflation series could not be read.")
			return b
		}
		pnl, cov, err := web.BuildMonthlyPanel(assets, hicp.Points)
		if err != nil {
			log.Printf("warning: fire: no historical panel: %v", err)
			b.notes = append(b.notes, "No historical panel: "+err.Error()+".")
			return b
		}
		b.panel = &pnl
		b.notes = append(b.notes, cov.Notes()...)
	}
	return b
}
