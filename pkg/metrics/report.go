package metrics

import "time"

// Window is one named window of a period report; To is inclusive. The
// value at From is the comparison base of the window, so a "ytd" window
// starts on Dec 31 of the previous year.
type Window struct {
	Name     string
	From, To time.Time
}

// StandardWindows returns the usual trailing report windows ending at to:
// 1d, 7d, 1m, 3m, ytd, 1y and prev-yr (the last full calendar year). 7d -
// one calendar week - covers five trading sessions, what "a week" means
// to a human (and what finance UIs label 5D). Month and year arithmetic
// follows Go's AddDate normalization. Callers slice, filter or extend the
// result freely before passing it to Report.
func StandardWindows(to time.Time) []Window {
	dec31 := func(year int) time.Time {
		return time.Date(year, time.December, 31, 0, 0, 0, 0, time.UTC)
	}
	return []Window{
		{Name: "1d", From: to.AddDate(0, 0, -1), To: to},
		{Name: "7d", From: to.AddDate(0, 0, -7), To: to},
		{Name: "1m", From: to.AddDate(0, -1, 0), To: to},
		{Name: "3m", From: to.AddDate(0, -3, 0), To: to},
		{Name: "ytd", From: dec31(to.Year() - 1), To: to},
		{Name: "1y", From: to.AddDate(-1, 0, 0), To: to},
		{Name: "prev-yr", From: dec31(to.Year() - 2), To: dec31(to.Year() - 1)},
	}
}
