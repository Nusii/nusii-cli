package output

import (
	"testing"
)

func TestDetectFormat(t *testing.T) {
	if Detect("json") != FormatJSON {
		t.Error("expected FormatJSON for 'json' flag")
	}
	if Detect("table") != FormatTable {
		t.Error("expected FormatTable for 'table' flag")
	}
}

func TestPrintTable(t *testing.T) {
	// Just verify it doesn't panic with various inputs
	PrintTable([]string{}, nil)
	PrintTable([]string{"A", "B"}, nil)
	PrintTable([]string{"A", "B"}, [][]string{{"1", "2"}})
	PrintTable([]string{"A", "B", "C"}, [][]string{{"1"}, {"1", "2", "3"}})
}
