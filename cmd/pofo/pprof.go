package main

import (
	"log"
	"net/http"
	_ "net/http/pprof" // registers /debug/pprof/* on http.DefaultServeMux
)

// startPprof serves the net/http/pprof endpoints on addr in the background,
// when -pprof is set. It is off by default and meant for temporary, local
// profiling of the long-running -serve / -fire servers: bind it to a loopback
// address (e.g. localhost:6060) and reach it over an SSH tunnel, never expose
// it publicly. The app's own routes live on their own mux, so these debug
// handlers stay isolated on a separate listener and add nothing to the request
// path when the flag is unset.
//
// Collect a 30s CPU profile while exercising the slow page:
//
//	go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
//
// or an allocation profile to chase GC pressure:
//
//	go tool pprof http://localhost:6060/debug/pprof/allocs
func startPprof(addr string) {
	log.Printf("pprof: serving profiles on http://%s/debug/pprof/ (do not expose publicly)", addr)
	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Printf("pprof: %v", err)
		}
	}()
}
