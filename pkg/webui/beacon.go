package webui

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
)

// BeaconScriptURL is Cloudflare's official Web Analytics loader. It is also the
// marker Beacon looks for to stay idempotent when two wrapped handlers nest
// (the -serve constellation mounts pkg/decumul/web, which can wrap its own mux).
const BeaconScriptURL = "https://static.cloudflareinsights.com/beacon.min.js"

// BeaconSnippet renders the Cloudflare Web Analytics tag for token, or "" when
// the token is empty. The tag is cookieless and stores nothing on the visitor's
// device: it loads the script above, which POSTs one page-view record to
// cloudflareinsights.com.
//
// The token is operator-supplied, and escaped anyway. encoding/json already
// turns "<", ">", "&", the double quote and every control character into
// escapes, which is exactly what a JSON payload sitting inside an HTML
// attribute needs; the single quote closing the attribute is the one character
// it leaves alone, so it is replaced here. The result is the tag Cloudflare
// documents, readable in the page source.
func BeaconSnippet(token string) string {
	if token == "" {
		return ""
	}
	raw, _ := json.Marshal(struct { // a struct of one string never fails
		Token string `json:"token"`
	}{token})
	attr := strings.ReplaceAll(string(raw), "'", "&#39;")
	return `<script defer src="` + BeaconScriptURL + `" data-cf-beacon='` + attr + `'></script>`
}

// Beacon wraps h so that every HTML page it serves carries the Cloudflare Web
// Analytics tag just before the closing </body>. An empty token returns h
// itself, so with the feature off not one byte of any page changes.
//
// It is a response filter rather than a template edit because pofo's HTML comes
// out of four independent renderers (the landing and hub templates, the
// comparison report in pkg/report, the book in pkg/firebook, the simulator page
// in pkg/decumul/web) that share design tokens but no head-or-foot helper.
// Wrapping the mux covers all of them, and every page mounted under them,
// from one place.
//
// Only a text/html response is buffered and rewritten, and only at a status
// that carries a page a visitor reads: 200, and the 4xx/5xx error pages, which
// are exactly where a broken link deserves to be counted. Everything else (the
// Markdown mirrors, the Atom feeds, the sitemap, JSON, SVG, the EPUB, the social
// cards, the courtesy stub in a redirect body, a 304) streams through untouched,
// which leaves the strong ETags those routes compute matching the bytes they
// send. Buffering is why Content-Length is dropped from a rewritten response:
// the body is longer than the inner handler announced. Handlers that need
// http.Flusher or http.Hijacker must not sit behind this wrapper; none of pofo's
// HTML handlers do (they all write a fully materialized page).
func Beacon(token string, h http.Handler) http.Handler {
	snippet := []byte(BeaconSnippet(token))
	if len(snippet) == 0 {
		return h
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bw := &beaconWriter{ResponseWriter: w}
		h.ServeHTTP(bw, r)
		bw.finish(snippet)
	})
}

// beaconWriter buffers an HTML response so the tag can be spliced into it, and
// passes anything else straight to the wrapped ResponseWriter.
type beaconWriter struct {
	http.ResponseWriter
	buf     bytes.Buffer
	status  int
	decided bool // WriteHeader has run: capture is settled
	capture bool // this response is HTML and is being buffered
}

func (b *beaconWriter) WriteHeader(code int) {
	if b.decided {
		return
	}
	b.decided, b.status = true, code
	if beaconable(code, b.Header().Get("Content-Type")) {
		// Hold the header back: the injected body has a different length,
		// so nothing may be committed before finish rewrites it.
		b.capture = true
		return
	}
	b.ResponseWriter.WriteHeader(code)
}

// beaconable reports whether a response is a page a visitor reads, hence one
// worth counting: HTML, at a status that carries a body someone lands on. A 1xx,
// a 204 and a 304 have no body at all; a 3xx body is a stub nobody reads.
func beaconable(code int, contentType string) bool {
	if !strings.HasPrefix(contentType, "text/html") {
		return false
	}
	return code == http.StatusOK || code >= 400
}

func (b *beaconWriter) Write(p []byte) (int, error) {
	if !b.decided {
		b.WriteHeader(http.StatusOK)
	}
	if b.capture {
		return b.buf.Write(p)
	}
	return b.ResponseWriter.Write(p)
}

// finish sends a buffered HTML response with the tag spliced in. A response
// that was never captured has already gone out as the handler wrote it.
func (b *beaconWriter) finish(snippet []byte) {
	if !b.capture {
		return
	}
	body := injectBeacon(b.buf.Bytes(), snippet)
	b.Header().Del("Content-Length")
	b.ResponseWriter.WriteHeader(b.status)
	_, _ = b.ResponseWriter.Write(body)
}

// injectBeacon splices snippet in front of the body's last </body>, or appends
// it when the page has none (a fragment, or a handler that closed nothing). A
// page that already carries the tag is returned as it is.
func injectBeacon(body, snippet []byte) []byte {
	if bytes.Contains(body, []byte(BeaconScriptURL)) {
		return body
	}
	at := bytes.LastIndex(body, []byte("</body>"))
	if at < 0 {
		at = len(body)
	}
	out := make([]byte, 0, len(body)+len(snippet))
	out = append(out, body[:at]...)
	out = append(out, snippet...)
	return append(out, body[at:]...)
}
