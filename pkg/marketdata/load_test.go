package marketdata

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// offlineClient is a client that can only read what is local: an empty
// cache and the bundle, so every Load below runs with no network.
func offlineClient(t *testing.T) *Client {
	t.Helper()
	c := NewClient(t.TempDir())
	c.Offline = true
	return c
}

func TestLoadDoors(t *testing.T) {
	ctx := context.Background()
	c := offlineClient(t)

	// The bundle: a reference series, then a catalog index.
	ref, err := c.Load(ctx, "tbill-3m", FetchOptions{})
	if err != nil || ref.Source != "refdata" || ref.Symbol != "TBILL-3M" {
		t.Fatalf("reference: %v, %+v", err, ref)
	}
	idx, err := c.Load(ctx, "SP500", FetchOptions{From: d(2000, 1, 1), To: d(2010, 12, 31)})
	if err != nil || idx.Source != "simdata" || idx.Symbol != "SP500" {
		t.Fatalf("catalog index: %v, %+v", err, idx)
	}
	if idx.First().Date.Before(d(2000, 1, 1)) || idx.Last().Date.After(d(2010, 12, 31)) {
		t.Errorf("window not applied: %s to %s", idx.First().Date, idx.Last().Date)
	}

	// The SIM convention: a quoted fund's bare identifier means its real
	// quotes, which an offline client with an empty cache does not hold. Its
	// reconstruction never answers in their place.
	if s, err := c.Load(ctx, "DBMF", FetchOptions{}); !errors.Is(err, ErrOffline) {
		t.Errorf("bare quoted fund offline: %v, source %q; want ErrOffline, never the backcast", err, sourceOf(s))
	}
	// The SIM suffix goes to the client, which offline and with an empty
	// cache still serves the SIM history it would splice, and says so.
	ext, err := c.Load(ctx, "DBMFSIM", FetchOptions{})
	if err != nil || ext.Len() < 1000 || ext.SimulatedBefore.IsZero() {
		t.Fatalf("SIM through the client: %v", err)
	}

	// A file, named by its path.
	path := filepath.Join(t.TempDir(), "mine.csv")
	if err := os.WriteFile(path, []byte("# currency: USD\ndate,value\n2020-01-31,100\n2020-02-29,101\n2020-03-31,99\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	file, err := c.Load(ctx, path, FetchOptions{})
	if err != nil || file.Symbol != path || file.Len() != 3 || file.Currency != "USD" {
		t.Fatalf("file: %v, %+v", err, file)
	}

	// Nothing local answers an unknown identifier offline.
	if _, err := c.Load(ctx, "NO-SUCH-FUND-XYZ", FetchOptions{}); !errors.Is(err, ErrOffline) {
		t.Errorf("unknown offline: %v, want ErrOffline", err)
	}
	// An empty window names the identifier.
	if _, err := c.Load(ctx, "TBILL-3M", FetchOptions{From: d(2200, 1, 1)}); err == nil || !strings.Contains(err.Error(), "TBILL-3M") {
		t.Errorf("empty window: %v", err)
	}
}

func TestLoadCurrency(t *testing.T) {
	ctx := context.Background()
	c := offlineClient(t)

	usd, err := c.Load(ctx, "SP500", FetchOptions{From: d(2015, 1, 1)})
	if err != nil {
		t.Fatal(err)
	}
	// The euro crosses are bundled, so the conversion runs offline.
	eur, err := c.Load(ctx, "SP500", FetchOptions{From: d(2015, 1, 1), Currency: "EUR"})
	if err != nil {
		t.Fatal(err)
	}
	if eur.Currency != "EUR" || eur.Len() != usd.Len() {
		t.Fatalf("converted: %q, %d points against %d", eur.Currency, eur.Len(), usd.Len())
	}
	if ru, re := usd.Last().Close/usd.First().Close, eur.Last().Close/eur.First().Close; ru == re {
		t.Errorf("conversion moved nothing: %.4f in both currencies", ru)
	}

	// A reference series states no currency: it passes through unchanged,
	// as FetchExtended passes one, and the log says so.
	var logged []string
	c.Logf = func(format string, args ...any) { logged = append(logged, fmt.Sprintf(format, args...)) }
	cash, err := c.Load(ctx, "EURCASH-EUR", FetchOptions{Currency: "EUR"})
	if err != nil || cash.Currency != "" {
		t.Errorf("reference series: %v, currency %q", err, cash.Currency)
	}
	if len(logged) != 1 || !strings.Contains(logged[0], "EURCASH-EUR") {
		t.Errorf("log = %q, want one line naming the series", logged)
	}
	// NoConvert demands the native line.
	if _, err := c.Load(ctx, "SP500", FetchOptions{Currency: "EUR", NoConvert: true}); !errors.Is(err, ErrWrongCurrency) {
		t.Errorf("NoConvert: %v, want ErrWrongCurrency", err)
	}
}

// sourceOf is s.Source, or "" for no series, for a failure message.
func sourceOf(s *Series) string {
	if s == nil {
		return ""
	}
	return s.Source
}
