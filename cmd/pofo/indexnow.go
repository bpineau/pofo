// The -indexnow mode: after a deploy, tell the participating search engines
// what this host publishes instead of waiting to be crawled again.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/bpineau/pofo/pkg/firebook"
	"github.com/bpineau/pofo/pkg/seo"
)

// indexNowTimeout bounds the whole submission. It is one POST of a few hundred
// URLs to one endpoint; a minute is already generous.
const indexNowTimeout = time.Minute

// runIndexNow submits every URL this server publishes at origin to the shared
// IndexNow endpoint.
//
// This is the ONLY mode in the program that talks to a search engine, and it
// does so only when asked by name: nothing is submitted by a running server,
// on a schedule or as a side effect of anything else. Run it by hand after a
// deploy, from a machine that can reach the endpoint:
//
//	pofo -indexnow https://example.org -indexnow-key <key>
//
// The key must be the one the deployed server serves at /<key>.txt (the
// -indexnow-key flag of -serve): the engine fetches that file to check that
// whoever submits actually controls the host.
func runIndexNow(ctx context.Context, origin, key string) error {
	origin, host, err := parseOrigin(origin)
	if err != nil {
		return err
	}
	if key == "" {
		return errors.New("-indexnow needs -indexnow-key: the key the deployed server serves at /<key>.txt")
	}
	sub := indexNowSubmission(origin, host, key)
	if err := sub.Validate(); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "submitting %d URLs of %s to %s\n", len(sub.URLs), host, seo.IndexNowEndpoint)
	ctx, cancel := context.WithTimeout(ctx, indexNowTimeout)
	defer cancel()
	if err := sub.Submit(ctx, http.DefaultClient, seo.IndexNowEndpoint); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "accepted; the endpoint fetches %s to verify the host\n", sub.KeyLocation)
	return nil
}

// indexNowSubmission assembles what to submit for one origin: the URL list is
// the sitemap's, built by the same firebook.Site the running server serves, so
// what is pushed and what is crawlable cannot drift apart. The key file lives
// at the root of the host, which is where -serve mounts it.
func indexNowSubmission(origin, host, key string) seo.IndexNow {
	return seo.IndexNow{
		Host:        host,
		Key:         key,
		KeyLocation: origin + "/" + key + ".txt",
		URLs:        serveSite().URLs(origin),
	}
}

// parseOrigin validates the public origin a submission is about and returns it
// normalized (no trailing slash) together with its host. A path, a query or a
// missing scheme is a mistake worth catching here: the whole URL list is built
// by concatenation onto this string.
func parseOrigin(origin string) (normalized, host string, err error) {
	if origin == "" {
		return "", "", errors.New("-indexnow needs the public origin, e.g. https://example.org")
	}
	u, err := url.Parse(strings.TrimSuffix(origin, "/"))
	if err != nil {
		return "", "", fmt.Errorf("invalid origin %q: %w", origin, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", "", fmt.Errorf("invalid origin %q: it must start with http:// or https://", origin)
	}
	if u.Host == "" {
		return "", "", fmt.Errorf("invalid origin %q: it names no host", origin)
	}
	if u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return "", "", fmt.Errorf("invalid origin %q: give the scheme and host only", origin)
	}
	return u.Scheme + "://" + u.Host, u.Host, nil
}

// serveSite is the machine-readable face of the -serve constellation: the two
// book editions plus this server's own pages. The mux builds it to serve
// /sitemap.xml, /robots.txt and /llms.txt; -indexnow builds the same value to
// know what to submit, which is why it lives in one place.
func serveSite() firebook.Site {
	return firebook.BookSite(
		firebook.Page{Path: "/", Title: "pofo", Note: "the front door"},
		firebook.Page{Path: "/visualizer", Title: "Portfolio visualizer",
			Note: "compose portfolios and backtest them side by side"},
		firebook.Page{Path: fireBase + "/", Title: "FIRE simulator",
			Note: "stress-test a withdrawal plan against thousands of simulated futures"},
	)
}
