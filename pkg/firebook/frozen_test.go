package firebook

// TEMPORARY characterization harness for the Edition refactor (M1 of
// docs/fire-book-en-edition-design.md): records a digest of everything the
// French edition renders, then fails if any later change moves it.
// Deleted at the end of M1.

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"os"
	"testing"
	"time"
)

func TestFrozenFrenchRendering(t *testing.T) {
	path := os.Getenv("FIREBOOK_FROZEN")
	if path == "" {
		t.Skip("set FIREBOOK_FROZEN=<file> to record or compare the French rendering digest")
	}
	h := sha256.New()
	_, _ = io.WriteString(h, French.indexHTML(0))
	for _, cat := range Categories {
		for _, a := range cat.Articles {
			_, _ = io.WriteString(h, French.articleHTML(a, cat))
		}
	}
	epubBytes, err := EPUB(time.Unix(0, 0).UTC())
	if err != nil {
		t.Fatal(err)
	}
	h.Write(epubBytes)
	sum := hex.EncodeToString(h.Sum(nil))
	prev, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		if err := os.WriteFile(path, []byte(sum), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("recorded baseline %s", sum)
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	if string(prev) != sum {
		t.Errorf("French rendering changed: digest %s, baseline %s", sum, string(prev))
	}
}
