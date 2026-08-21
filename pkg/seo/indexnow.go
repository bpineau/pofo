package seo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// IndexNow (indexnow.org) is the push half of being indexed: instead of waiting
// for a crawler to come back, a site tells the participating engines (Bing,
// Yandex, Seznam, Naver, and anyone else sharing the endpoint) which URLs
// changed. One submission reaches all of them.
//
// Ownership is proved by a key file: a text file at the root of the host whose
// name is the key and whose only content is that same key. The engine fetches
// it before trusting a submission, which is why KeyLocation is part of the
// payload rather than implied.
const (
	// IndexNowEndpoint is the shared endpoint that forwards a submission to
	// every participating engine.
	IndexNowEndpoint = "https://api.indexnow.org/indexnow"

	// IndexNowBatch is the protocol's cap on URLs per request.
	IndexNowBatch = 10000

	// IndexNowKeyType is the Content-Type of the key file. The spec asks for
	// plain text and nothing else; the file's whole body is the key.
	IndexNowKeyType = "text/plain; charset=utf-8"
)

// IndexNow is one submission: the host it covers, the key proving ownership of
// that host, where the key file is served, and the URLs to submit.
//
// The URLs must all live on Host: the protocol rejects a batch that mixes
// hosts, and Validate says so before anything is sent.
type IndexNow struct {
	Host        string   // "example.org", no scheme
	Key         string   // the key, also the key file's name and content
	KeyLocation string   // absolute URL of the key file
	URLs        []string // absolute URLs to submit
}

// indexNowBody is the wire shape of one batch, per the IndexNow specification.
type indexNowBody struct {
	Host        string   `json:"host"`
	Key         string   `json:"key"`
	KeyLocation string   `json:"keyLocation,omitempty"`
	URLList     []string `json:"urlList"`
}

// ValidIndexNowKey reports whether key has the shape the protocol requires: 8
// to 128 characters, letters, digits and dashes only.
//
// The check matters twice over. A key that does not validate is rejected by
// the engines, and it is also used verbatim as a route ("/<key>.txt"), so a
// key carrying a slash or a query would mount something other than a file.
func ValidIndexNowKey(key string) bool {
	if len(key) < 8 || len(key) > 128 {
		return false
	}
	for _, r := range key {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-':
		default:
			return false
		}
	}
	return true
}

// Validate reports what would make the submission fail before it is sent: a
// malformed key, no host, no URLs, or a URL that does not live on Host.
func (n IndexNow) Validate() error {
	if !ValidIndexNowKey(n.Key) {
		return fmt.Errorf("seo: %q is not a valid IndexNow key (8 to 128 letters, digits and dashes)", n.Key)
	}
	if n.Host == "" {
		return fmt.Errorf("seo: IndexNow submission names no host")
	}
	if len(n.URLs) == 0 {
		return fmt.Errorf("seo: IndexNow submission carries no URL")
	}
	for _, u := range n.URLs {
		if !strings.Contains(u, "://"+n.Host+"/") && !strings.HasSuffix(u, "://"+n.Host) {
			return fmt.Errorf("seo: %q does not live on %s", u, n.Host)
		}
	}
	return nil
}

// Bodies renders the submission as the JSON request bodies to POST, one per
// batch of at most IndexNowBatch URLs.
func (n IndexNow) Bodies() [][]byte {
	var out [][]byte
	for rest := n.URLs; len(rest) > 0; {
		batch := rest
		if len(batch) > IndexNowBatch {
			batch = batch[:IndexNowBatch]
		}
		rest = rest[len(batch):]
		body, err := json.Marshal(indexNowBody{
			Host: n.Host, Key: n.Key, KeyLocation: n.KeyLocation, URLList: batch,
		})
		if err != nil { // a struct of strings never fails to marshal
			continue
		}
		out = append(out, body)
	}
	return out
}

// Submit validates the submission and POSTs it to endpoint (IndexNowEndpoint in
// production), one request per batch. The caller supplies the client, so a test
// points it at an httptest server and nothing here ever reaches the network on
// its own.
//
// 200 and 202 are both success: 202 means the key file has not been read yet
// and the submission is queued behind that check. Anything else is reported
// with the endpoint's own words, which say more than the status alone (403 is
// a key it could not verify, 422 URLs that do not match the host).
func (n IndexNow) Submit(ctx context.Context, client *http.Client, endpoint string) error {
	if err := n.Validate(); err != nil {
		return err
	}
	if client == nil {
		client = http.DefaultClient
	}
	for i, body := range n.Bodies() {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		res, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("seo: IndexNow batch %d: %w", i+1, err)
		}
		answer, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
		res.Body.Close()
		if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusAccepted {
			return fmt.Errorf("seo: IndexNow batch %d: %s: %s",
				i+1, res.Status, strings.TrimSpace(string(answer)))
		}
	}
	return nil
}
