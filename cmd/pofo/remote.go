package main

import (
	"context"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/remote"
)

type remoteClientKey struct{}

// buildVersion identifies a build made without VCS metadata, such as the
// container image, whose context excludes .git. It is set with
// -ldflags "-X main.buildVersion=<tag>@<commit>" and wins when non-empty.
var buildVersion string

// engineVersion is the build identity reported by the numerical API and MCP:
// buildVersion, else the module version and VCS revision Go recorded.
func engineVersion() string {
	if buildVersion != "" {
		return buildVersion
	}
	if b, ok := debug.ReadBuildInfo(); ok {
		rev := b.Main.Version
		for _, v := range b.Settings {
			if v.Key == "vcs.revision" {
				rev = v.Value
			}
			if v.Key == "vcs.modified" && v.Value == "true" {
				rev += "+modified"
			}
		}
		return rev
	}
	return "development"
}
func (s *server) mountRemote(mux *http.ServeMux) {
	if !s.opt.serveAPI && !s.opt.serveMCP {
		return
	}
	service := remote.New(s.client, engineVersion())
	service.BeforeFetch = func(ctx context.Context, id string) error {
		if id == "^HICP-FR" || marketdata.KnownLocal(id) || s.client.Cached(id) {
			return nil
		}
		if s.foreign == nil {
			return &remote.Error{Code: "unsupported_configuration", Message: "This server accepts locally known assets only."}
		}
		key, _ := ctx.Value(remoteClientKey{}).(string)
		if err := s.foreign.charge(key, 1); err != nil {
			return &remote.Error{Code: "rate_limited", Message: "Hourly allowance for uncached foreign identifiers exhausted.", Retryable: true, RetryAfterSeconds: 3600}
		}
		return nil
	}
	var origins []string
	for _, v := range strings.Split(s.opt.serveOrigins, ",") {
		if v = strings.TrimSpace(v); v != "" {
			origins = append(origins, v)
		}
	}
	h := remote.Handler(service, remote.HTTPOptions{API: s.opt.serveAPI, MCP: s.opt.serveMCP, Origins: origins, ClientID: clientIP, Acquire: s.acquireWork, Observe: func(e remote.Event) {
		log.Printf("numerical operation=%s duration=%s queue_wait=%s response_bytes=%d code=%s upstream_attempts_total=%d", e.Operation, e.Duration.Round(time.Millisecond), e.QueueWait.Round(time.Millisecond), e.ResponseBytes, e.Code, s.client.Controls.HTTPAttempts())
	}})
	wrapped := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), remoteClientKey{}, clientIP(r))))
	})
	if s.opt.serveAPI {
		mux.Handle("/api/v1/", wrapped)
	}
	if s.opt.serveMCP {
		mux.Handle("/mcp", wrapped)
	}
}

// acquireWork shares the /view capacity, with a bounded waiting room. No request
// gains another minute merely by sitting in the queue.
func (s *server) acquireWork(ctx context.Context) (func(), error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	select {
	case s.sem <- struct{}{}:
		return func() { <-s.sem }, nil
	default:
	}
	select {
	case s.workQueue <- struct{}{}:
		defer func() { <-s.workQueue }()
	default:
		return nil, &remote.Error{Code: "server_busy", Message: "Computation queue is full; retry later.", Retryable: true}
	}
	select {
	case s.sem <- struct{}{}:
		return func() { <-s.sem }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
