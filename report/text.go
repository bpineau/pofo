package report

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

// RenderText writes the comparison summary — title, common period and the
// statistics table — as aligned plain text for the CLI mode. The best cell
// of each row is shown in green (or marked with a star without colors).
// Per-portfolio details are deliberately omitted.
func RenderText(w io.Writer, page *Page, color bool) error {
	if _, err := fmt.Fprintf(w, "%s\nPériode commune : %s → %s\n\n", page.Title, page.CommonStart, page.CommonEnd); err != nil {
		return err
	}

	// Column widths (visual width ≈ rune count for our character set).
	widths := make([]int, len(page.PortfolioNames)+1)
	widths[0] = utf8.RuneCountInString("Métrique")
	for _, r := range page.StatRows {
		widths[0] = max(widths[0], utf8.RuneCountInString(r.Label))
	}
	for i, n := range page.PortfolioNames {
		widths[i+1] = utf8.RuneCountInString(n)
		for _, r := range page.StatRows {
			if i < len(r.Cells) {
				widths[i+1] = max(widths[i+1], utf8.RuneCountInString(r.Cells[i].Text)+2)
			}
		}
	}

	pad := func(s string, w int, right bool) string {
		fill := strings.Repeat(" ", max(w-utf8.RuneCountInString(s), 0))
		if right {
			return fill + s
		}
		return s + fill
	}
	line := func(parts []string) string {
		return strings.Join(parts, "  ")
	}

	header := []string{pad("Métrique", widths[0], false)}
	for i, n := range page.PortfolioNames {
		header = append(header, pad(n, widths[i+1], true))
	}
	fmt.Fprintln(w, line(header))
	total := widths[0] + 2*len(widths) - 2
	for _, x := range widths[1:] {
		total += x
	}
	fmt.Fprintln(w, strings.Repeat("─", total))

	for _, r := range page.StatRows {
		parts := []string{pad(r.Label, widths[0], false)}
		for i, c := range r.Cells {
			text := c.Text
			if c.Best {
				if color {
					text = pad(text, widths[i+1], true)
					text = strings.Replace(text, c.Text, "\x1b[32;1m"+c.Text+"\x1b[0m", 1)
					parts = append(parts, text)
					continue
				}
				text = "*" + text
			}
			parts = append(parts, pad(text, widths[i+1], true))
		}
		fmt.Fprintln(w, line(parts))
	}
	return nil
}
