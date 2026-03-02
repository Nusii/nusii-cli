package output

import (
	"fmt"
	"strings"
)

// PrintTable renders headers and rows as a simple aligned table to stdout.
func PrintTable(headers []string, rows [][]string) {
	if len(headers) == 0 {
		return
	}

	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print header
	printRow(headers, widths)
	// Print separator
	sep := make([]string, len(headers))
	for i, w := range widths {
		sep[i] = strings.Repeat("-", w)
	}
	printRow(sep, widths)

	// Print rows
	for _, row := range rows {
		// Pad row if shorter than headers
		for len(row) < len(headers) {
			row = append(row, "")
		}
		printRow(row, widths)
	}
}

func printRow(cells []string, widths []int) {
	parts := make([]string, len(cells))
	for i, cell := range cells {
		w := 0
		if i < len(widths) {
			w = widths[i]
		}
		parts[i] = fmt.Sprintf("%-*s", w, cell)
	}
	fmt.Println(strings.Join(parts, "  "))
}
