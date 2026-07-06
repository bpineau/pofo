package permanent

import (
	"testing"
	"time"
)

func TestLoadPanel(t *testing.T) {
	p, err := LoadPanel()
	if err != nil {
		t.Fatalf("LoadPanel: %v", err)
	}
	if len(p.Countries()) < 20 {
		t.Fatalf("panel covers only %d countries", len(p.Countries()))
	}
	// A well-known country-month should carry an inflation reading.
	if _, ok := p.yoy("cpi", "USA", time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC)); !ok {
		t.Fatal("missing USA CPI year-on-year at 2010-01")
	}
}

func TestParsePanelRejectsBadRows(t *testing.T) {
	if _, err := ParsePanel([]byte("iso,date,ip,cpi,shortrate,longrate,shareprice\nUSA,2020-01,1,2,3\n")); err == nil {
		t.Fatal("expected error on short row")
	}
	if _, err := ParsePanel([]byte("# only comments\n")); err == nil {
		t.Fatal("expected error on empty panel")
	}
}
