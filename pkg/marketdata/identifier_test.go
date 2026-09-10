package marketdata

import "testing"

func TestPlausibleID(t *testing.T) {
	cases := []struct {
		id   string
		want bool
		why  string
	}{
		{"VOO", true, "a plain ticker"},
		{"voo", true, "case does not matter"},
		{" AVUV ", true, "surrounding space is trimmed"},
		{"IWDA.AS", true, "one exchange suffix"},
		{"7203.T", true, "a numeric base (Tokyo)"},
		{"BRK-B", true, "a share-class part"},
		{"BRK-B.MX", true, "a share class and an exchange"},
		{"US0378331005", true, "a valid ISIN"},
		{"US0378331004", false, "an ISIN with a wrong check digit is a typo, not a ticker"},
		{"", false, "nothing at all"},
		{"^GSPC", false, "a quote symbol is not an instrument a visitor may mint"},
		{"GC=F", false, "a futures symbol carries an illegal character"},
		{"BRK.B.MX", false, "two exchange suffixes"},
		{"VOO VTI", false, "a space"},
		{"MSCI WORLD", false, "a name, not an identifier"},
		{"VOO/../ETC", false, "path characters"},
		{"AVERYLONGIDENTIFIER", false, "longer than any real ticker"},
		{"VOO;DROP", false, "punctuation"},
	}
	for _, c := range cases {
		if got := PlausibleID(c.id); got != c.want {
			t.Errorf("PlausibleID(%q) = %v, want %v (%s)", c.id, got, c.want, c.why)
		}
	}
}
