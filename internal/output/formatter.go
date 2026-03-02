package output

import (
	"encoding/json"
	"fmt"
	"os"
)

// Format represents an output format.
type Format string

const (
	FormatJSON  Format = "json"
	FormatTable Format = "table"
)

// Detect returns the appropriate format based on flags and TTY detection.
func Detect(flagValue string) Format {
	switch flagValue {
	case "json":
		return FormatJSON
	case "table":
		return FormatTable
	default:
		// Auto-detect: table for TTY, JSON for pipes
		if isTerminal() {
			return FormatTable
		}
		return FormatJSON
	}
}

func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// PrintJSON prints raw JSON bytes with optional pretty-printing.
func PrintJSON(raw []byte) error {
	var out json.RawMessage
	if err := json.Unmarshal(raw, &out); err != nil {
		// Not valid JSON, print as-is
		fmt.Println(string(raw))
		return nil
	}
	pretty, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(pretty))
	return nil
}

// PrintErrorJSON prints an error in JSON format to stderr.
func PrintErrorJSON(msg string, status int) {
	e := struct {
		Error  string `json:"error"`
		Status int    `json:"status"`
	}{Error: msg, Status: status}
	b, _ := json.Marshal(e)
	fmt.Fprintln(os.Stderr, string(b))
}
