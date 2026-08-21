package seo

import (
	"encoding/xml"
	"strings"
	"testing"
	"time"
)

func TestSitemapWellFormed(t *testing.T) {
	body := Sitemap([]URL{
		{Loc: "https://example.org/a?x=1&y=2"},
		{Loc: ""}, // skipped
		{Loc: "https://example.org/b", LastMod: time.Date(2026, 1, 2, 15, 4, 5, 0, time.UTC)},
	})
	var got struct {
		XMLName xml.Name
		URLs    []struct {
			Loc     string `xml:"loc"`
			LastMod string `xml:"lastmod"`
		} `xml:"url"`
	}
	if err := xml.Unmarshal(body, &got); err != nil {
		t.Fatalf("sitemap is not well-formed XML: %v", err)
	}
	if got.XMLName.Local != "urlset" || got.XMLName.Space != "http://www.sitemaps.org/schemas/sitemap/0.9" {
		t.Errorf("root element is %v, want urlset in the sitemaps.org namespace", got.XMLName)
	}
	if len(got.URLs) != 2 {
		t.Fatalf("got %d entries, want 2 (the empty Loc must be skipped)", len(got.URLs))
	}
	if got.URLs[0].Loc != "https://example.org/a?x=1&y=2" {
		t.Errorf("loc round-trip: %q", got.URLs[0].Loc)
	}
	if got.URLs[0].LastMod != "" {
		t.Errorf("a zero LastMod must emit no <lastmod>, got %q", got.URLs[0].LastMod)
	}
	if got.URLs[1].LastMod != "2026-01-02" {
		t.Errorf("lastmod = %q, want 2026-01-02", got.URLs[1].LastMod)
	}
	if !strings.HasPrefix(string(body), xml.Header) {
		t.Errorf("sitemap misses the XML declaration")
	}
}

func TestRobotsDefaults(t *testing.T) {
	got := string(Robots{Groups: []Group{{}, {Agents: []string{"CCBot"}, Disallow: []string{"/private/"}}}}.Text())
	if !strings.Contains(got, "User-agent: *\nAllow: /\n") {
		t.Errorf("an empty group must render the catch-all allow-everything record:\n%s", got)
	}
	if !strings.Contains(got, "User-agent: CCBot\nDisallow: /private/\n") {
		t.Errorf("explicit rules must be rendered verbatim:\n%s", got)
	}
	if strings.Contains(got, "User-agent: CCBot\nAllow: /") {
		t.Errorf("a group with explicit rules must not gain a default Allow:\n%s", got)
	}
	if strings.Contains(got, "Sitemap:") {
		t.Errorf("no sitemap was declared, none must be printed:\n%s", got)
	}
}

func TestRobotsMultilineComment(t *testing.T) {
	got := string(Robots{Preamble: []string{"one\ntwo"}}.Text())
	if !strings.HasPrefix(got, "# one\n# two\n") {
		t.Errorf("every comment line needs its own '#':\n%s", got)
	}
}

func TestLLMsShape(t *testing.T) {
	got := string(LLMs{
		Title:    "Site",
		Summary:  "A summary.",
		Sections: []Section{{Title: "Part", Links: []Link{{Title: "A", URL: "/a"}, {Title: "B", URL: "/b", Note: "n"}}}},
	}.Text())
	for _, want := range []string{"# Site\n", "\n> A summary.\n", "\n## Part\n", "- [A](/a)\n", "- [B](/b): n\n"} {
		if !strings.Contains(got, want) {
			t.Errorf("llms.txt misses %q:\n%s", want, got)
		}
	}
}
