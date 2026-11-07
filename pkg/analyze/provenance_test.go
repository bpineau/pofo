package analyze

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

func TestIncomeQualificationAfterCSVAndBlend(t *testing.T) {
	s, err := marketdata.NewSeries("PRICE", []time.Time{time.Date(2020, 1, 6, 0, 0, 0, 0, time.UTC), time.Date(2020, 1, 7, 0, 0, 0, 0, time.UTC), time.Date(2020, 1, 8, 0, 0, 0, 0, time.UTC)}, []float64{100, 101, 102})
	if err != nil {
		t.Fatal(err)
	}
	s.Source = "stooq"
	s.Currency = "USD"
	panel, err := marketdata.NewPanel(marketdata.Daily, s)
	if err != nil {
		t.Fatal(err)
	}
	panel, err = panel.Mix("MIX", map[string]float64{"PRICE": 1})
	if err != nil {
		t.Fatal(err)
	}
	mix, err := panel.Series("MIX")
	if err != nil {
		t.Fatal(err)
	}
	var b bytes.Buffer
	if err := marketdata.WriteCSV(&b, s); err != nil {
		t.Fatal(err)
	}
	read, err := marketdata.ReadLongCSV(&b)
	if err != nil {
		t.Fatal(err)
	}
	for _, input := range []*marketdata.Series{s, read[0], mix} {
		study, err := Series(input, nil, Options{})
		if err != nil {
			t.Fatal(err)
		}
		count := 0
		for _, w := range study.Warnings {
			if strings.Contains(w, "income") {
				count++
			}
		}
		if count != 1 {
			t.Errorf("source %s: expected exactly one income qualification: %v", input.Source, study.Warnings)
		}
	}
}
