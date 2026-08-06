package marketdata

import (
	"regexp"
	"strings"
)

var isinPattern = regexp.MustCompile(`^[A-Z]{2}[A-Z0-9]{9}[0-9]$`)

// IsISIN reports whether id is a syntactically valid ISIN, including its
// check digit, so that 12-character tickers are not misclassified.
func IsISIN(id string) bool {
	id = strings.ToUpper(id)
	if !isinPattern.MatchString(id) {
		return false
	}
	// Expand letters to two digits each (A=10 … Z=35), then run the Luhn
	// algorithm over the resulting digit string, check digit included.
	digits := make([]int, 0, 2*len(id))
	for _, r := range id {
		if r >= '0' && r <= '9' {
			digits = append(digits, int(r-'0'))
			continue
		}
		v := int(r-'A') + 10
		digits = append(digits, v/10, v%10)
	}
	sum := 0
	double := false // rightmost digit (the check digit) is not doubled
	for i := len(digits) - 1; i >= 0; i-- {
		d := digits[i]
		if double {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		double = !double
	}
	return sum%10 == 0
}
